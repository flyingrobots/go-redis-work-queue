// Copyright 2025 James Ross
package jsonpayloadstudio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

// JSONPayloadStudio provides JSON editing and validation capabilities
type JSONPayloadStudio struct {
	config       *StudioConfig
	redis        *redis.Client
	logger       *zap.Logger
	templates    map[string]*Template
	schemas      map[string]*JSONSchema
	snippets     map[string]*Snippet
	sessions     map[string]*SessionInfo
	lastEnqueued *EnqueueResult
	mu           sync.RWMutex
}

// NewJSONPayloadStudio creates a new JSON Payload Studio
func NewJSONPayloadStudio(config *StudioConfig, redis *redis.Client, logger *zap.Logger) (*JSONPayloadStudio, error) {
	if config == nil {
		config = &StudioConfig{
			EditorTheme:      "dark",
			SyntaxHighlight:  true,
			LineNumbers:      true,
			AutoFormat:       true,
			BracketMatching:  true,
			AutoComplete:     true,
			ValidateOnType:   true,
			TemplatesPath:    "./templates",
			SchemasPath:      "./schemas",
			MaxPayloadSize:   1024 * 1024, // 1MB
			MaxFieldCount:    1000,
			MaxNestingDepth:  20,
			StripSecrets:     true,
			ShowPreview:      true,
			PreviewLines:     50,
			HistorySize:      100,
			AutoSave:         true,
			AutoSaveInterval: 30 * time.Second,
		}
	}

	studio := &JSONPayloadStudio{
		config:    config,
		redis:     redis,
		logger:    logger,
		templates: make(map[string]*Template),
		schemas:   make(map[string]*JSONSchema),
		snippets:  make(map[string]*Snippet),
		sessions:  make(map[string]*SessionInfo),
	}

	// Load templates and schemas
	if err := studio.loadTemplates(); err != nil {
		logger.Warn("Failed to load templates", zap.Error(err))
	}

	if err := studio.loadSchemas(); err != nil {
		logger.Warn("Failed to load schemas", zap.Error(err))
	}

	// Initialize default snippets
	studio.initializeSnippets()

	return studio, nil
}

// CreateSession creates a new editor session
func (jps *JSONPayloadStudio) CreateSession() string {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	session := &SessionInfo{
		ID:           uuid.New().String(),
		StartedAt:    time.Now(),
		LastActivity: time.Now(),
		EditorState: &EditorState{
			Content:      "{}",
			CursorLine:   1,
			CursorColumn: 1,
			Modified:     false,
			History:      make([]string, 0, jps.config.HistorySize),
			HistoryIndex: 0,
		},
		Templates: make([]string, 0),
	}

	jps.sessions[session.ID] = session

	if jps.config.AutoSave {
		go jps.autoSaveSession(session.ID)
	}

	return session.ID
}

// UpdateEditorState updates the editor state for a session using the provided snapshot.
func (jps *JSONPayloadStudio) UpdateEditorState(sessionID string, newState *EditorState) error {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	session, exists := jps.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	if newState == nil {
		return fmt.Errorf("editor state required")
	}

	if session.EditorState == nil {
		session.EditorState = &EditorState{
			History: make([]string, 0, jps.config.HistorySize),
		}
	}

	state := session.EditorState

	if state.History == nil {
		state.History = make([]string, 0, jps.config.HistorySize)
	}

	if newState.Content != "" && newState.Content != state.Content {
		if state.HistoryIndex < len(state.History)-1 {
			state.History = state.History[:state.HistoryIndex+1]
		}
		state.History = append(state.History, state.Content)
		if jps.config.HistorySize > 0 && len(state.History) > jps.config.HistorySize {
			state.History = state.History[1:]
		} else if len(state.History) > 0 {
			state.HistoryIndex++
		}
		state.Content = newState.Content
		state.Modified = true
	}

	state.CursorLine = newState.CursorLine
	state.CursorColumn = newState.CursorColumn
	state.SelectionStart = newState.SelectionStart
	state.SelectionEnd = newState.SelectionEnd
	state.Schema = newState.Schema
	state.Template = newState.Template

	if jps.config.ValidateOnType {
		result := jps.ValidateJSON(state.Content, state.Schema)
		state.Errors = result.Errors
		state.Warnings = result.Warnings
	} else {
		state.Errors = newState.Errors
		state.Warnings = newState.Warnings
	}

	session.LastActivity = time.Now()
	return nil
}

// GetSession returns a snapshot of the session by ID.
func (jps *JSONPayloadStudio) GetSession(sessionID string) (*SessionInfo, error) {
	jps.mu.RLock()
	session, exists := jps.sessions[sessionID]
	jps.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	copy := *session
	if session.EditorState != nil {
		stateCopy := *session.EditorState
		copy.EditorState = &stateCopy
	}
	return &copy, nil
}

// DeleteSession removes a session from the studio.
func (jps *JSONPayloadStudio) DeleteSession(sessionID string) error {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	if _, exists := jps.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	delete(jps.sessions, sessionID)
	return nil
}

// GetSchema returns a schema by ID.
func (jps *JSONPayloadStudio) GetSchema(id string) (*JSONSchema, error) {
	jps.mu.RLock()
	schema, exists := jps.schemas[id]
	jps.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", id)
	}
	return cloneJSONSchema(schema), nil
}

// GetTemplate returns a template by ID.
func (jps *JSONPayloadStudio) GetTemplate(id string) (*Template, error) {
	jps.mu.RLock()
	tmpl, exists := jps.templates[id]
	jps.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return cloneTemplate(tmpl), nil
}

// ListTemplates returns all templates stored in the studio.
func (jps *JSONPayloadStudio) ListTemplates() []Template {
	jps.mu.RLock()
	defer jps.mu.RUnlock()

	result := make([]Template, 0, len(jps.templates))
	for _, tmpl := range jps.templates {
		result = append(result, *cloneTemplate(tmpl))
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	return result
}

// SaveTemplate stores or updates a template in memory.
func (jps *JSONPayloadStudio) SaveTemplate(template *Template) error {
	if template == nil {
		return fmt.Errorf("template is nil")
	}

	jps.mu.Lock()
	defer jps.mu.Unlock()

	jps.templates[template.ID] = cloneTemplate(template)
	return nil
}

// DeleteTemplate removes a template by ID.
func (jps *JSONPayloadStudio) DeleteTemplate(id string) error {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	if _, exists := jps.templates[id]; !exists {
		return fmt.Errorf("template not found: %s", id)
	}
	delete(jps.templates, id)
	return nil
}

// ApplyTemplate applies a template with optional variable overrides.
func (jps *JSONPayloadStudio) ApplyTemplate(templateID string, variables map[string]interface{}) (interface{}, error) {
	jps.mu.RLock()
	tmpl, exists := jps.templates[templateID]
	jps.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	resolved := make(map[string]string)
	for _, variable := range tmpl.Variables {
		if variable.DefaultValue != nil {
			resolved[variable.Name] = fmt.Sprint(variable.DefaultValue)
		}
	}
	for k, v := range variables {
		resolved[k] = fmt.Sprint(v)
	}

	contentCopy := cloneValue(tmpl.Content)
	return expandPlaceholders(contentCopy, resolved), nil
}

// ListSnippets returns all configured snippets.
func (jps *JSONPayloadStudio) ListSnippets() []Snippet {
	jps.mu.RLock()
	defer jps.mu.RUnlock()

	snippets := make([]Snippet, 0, len(jps.snippets))
	for _, snippet := range jps.snippets {
		copy := *snippet
		snippets = append(snippets, copy)
	}
	sort.Slice(snippets, func(i, j int) bool {
		return strings.ToLower(snippets[i].Trigger) < strings.ToLower(snippets[j].Trigger)
	})
	return snippets
}

// ExpandSnippet expands a snippet by trigger and applies optional variable overrides.
func (jps *JSONPayloadStudio) ExpandSnippet(trigger string, variables map[string]interface{}) (string, error) {
	jps.mu.RLock()
	defer jps.mu.RUnlock()

	snippet, ok := jps.snippets[trigger]
	if !ok {
		for _, candidate := range jps.snippets {
			if candidate.Trigger == trigger || candidate.ID == trigger {
				snippet = candidate
				ok = true
				break
			}
		}
	}
	if !ok {
		return "", fmt.Errorf("snippet not found: %s", trigger)
	}

	expanded := jps.expandSnippet(snippet)
	if variables != nil {
		for k, v := range variables {
			placeholder := fmt.Sprintf("${%s}", strings.ToUpper(k))
			expanded = strings.ReplaceAll(expanded, placeholder, fmt.Sprint(v))
		}
	}

	return expanded, nil
}

// DiffPayloads returns a diff between two JSON payloads.
func (jps *JSONPayloadStudio) DiffPayloads(oldPayload, newPayload interface{}) (*DiffResult, error) {
	return jps.compareJSON(oldPayload, newPayload), nil
}

// GeneratePreview renders a simple preview for arbitrary content.
func (jps *JSONPayloadStudio) GeneratePreview(content interface{}, maxLines int, truncate bool) *PreviewData {
	preview := &PreviewData{Content: content}

	formatted, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		preview.Valid = false
		preview.Error = err.Error()
		return preview
	}

	preview.Formatted = string(formatted)
	preview.Size = len(formatted)
	preview.LineCount = strings.Count(preview.Formatted, "\n") + 1
	preview.Valid = true

	limit := maxLines
	if limit <= 0 {
		limit = jps.config.PreviewLines
	}
	if truncate && limit > 0 && preview.LineCount > limit {
		lines := strings.Split(preview.Formatted, "\n")
		preview.Formatted = strings.Join(lines[:limit], "\n") + "\n..."
		preview.Truncated = true
	}

	return preview
}

func cloneTemplate(t *Template) *Template {
	if t == nil {
		return nil
	}

	clone := *t
	if t.Content != nil {
		if copied, ok := cloneValue(t.Content).(map[string]interface{}); ok {
			clone.Content = copied
		}
	}
	clone.Tags = append([]string(nil), t.Tags...)
	clone.Variables = append([]TemplateVariable(nil), t.Variables...)
	clone.Snippets = append([]Snippet(nil), t.Snippets...)
	if t.Schema != nil {
		clone.Schema = cloneJSONSchema(t.Schema)
	}
	return &clone
}

func cloneJSONSchema(schema *JSONSchema) *JSONSchema {
	if schema == nil {
		return nil
	}

	clone := *schema
	if schema.Properties != nil {
		if props, ok := cloneValue(schema.Properties).(map[string]interface{}); ok {
			clone.Properties = props
		}
	}
	clone.Required = append([]string(nil), schema.Required...)
	if schema.Definitions != nil {
		if defs, ok := cloneValue(schema.Definitions).(map[string]interface{}); ok {
			clone.Definitions = defs
		}
	}
	if schema.Additional != nil {
		if addl, ok := cloneValue(schema.Additional).(map[string]interface{}); ok {
			clone.Additional = addl
		}
	}
	return &clone
}

func cloneValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{}, len(v))
		for key, val := range v {
			m[key] = cloneValue(val)
		}
		return m
	case []interface{}:
		slice := make([]interface{}, len(v))
		for i, val := range v {
			slice[i] = cloneValue(val)
		}
		return slice
	default:
		return value
	}
}

func expandPlaceholders(value interface{}, vars map[string]string) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{}, len(v))
		for key, val := range v {
			m[key] = expandPlaceholders(val, vars)
		}
		return m
	case []interface{}:
		slice := make([]interface{}, len(v))
		for i, val := range v {
			slice[i] = expandPlaceholders(val, vars)
		}
		return slice
	case string:
		trimmed := strings.TrimSpace(v)
		if strings.HasPrefix(trimmed, "{{") && strings.HasSuffix(trimmed, "}}") && len(trimmed) >= 4 {
			token := strings.TrimSpace(trimmed[2 : len(trimmed)-2])
			return resolvePlaceholder(token, vars)
		}
		return v
	default:
		return v
	}
}

func resolvePlaceholder(token string, vars map[string]string) interface{} {
	if val, ok := vars[token]; ok {
		return val
	}
	if val, ok := vars[strings.ToUpper(token)]; ok {
		return val
	}
	if val, ok := vars[strings.ToLower(token)]; ok {
		return val
	}

	switch strings.ToLower(token) {
	case "now":
		return time.Now().UTC().Format(time.RFC3339)
	case "uuid":
		return uuid.NewString()
	}

	return "{{" + token + "}}"
}

// ValidateJSON validates JSON content
func (jps *JSONPayloadStudio) ValidateJSON(content string, schema *JSONSchema) *LintResult {
	result := &LintResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationError, 0),
		Info:     make([]ValidationError, 0),
	}

	// Parse JSON
	var parsed interface{}
	decoder := json.NewDecoder(strings.NewReader(content))
	decoder.UseNumber()

	if err := decoder.Decode(&parsed); err != nil {
		// Parse error to get line/column
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			line, col := getLineColumn(content, int(syntaxErr.Offset))
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Line:     line,
				Column:   col,
				Type:     "syntax",
				Message:  err.Error(),
				Severity: "error",
			})
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Line:     1,
				Column:   1,
				Type:     "parse",
				Message:  err.Error(),
				Severity: "error",
			})
		}
		return result
	}

	// Calculate statistics
	result.Stats = calculateStats(parsed, content)

	// Validate against schema if provided
	if schema != nil {
		schemaErrors := jps.validateAgainstSchema(parsed, schema)
		result.Errors = append(result.Errors, schemaErrors...)
		if len(schemaErrors) > 0 {
			result.Valid = false
		}
	}

	// Check size limits
	if len(content) > jps.config.MaxPayloadSize {
		result.Warnings = append(result.Warnings, ValidationError{
			Type:     "size",
			Message:  fmt.Sprintf("Payload size (%d bytes) exceeds maximum (%d bytes)", len(content), jps.config.MaxPayloadSize),
			Severity: "warning",
		})
	}

	// Check field count
	if result.Stats.Keys > jps.config.MaxFieldCount {
		result.Warnings = append(result.Warnings, ValidationError{
			Type:     "complexity",
			Message:  fmt.Sprintf("Field count (%d) exceeds maximum (%d)", result.Stats.Keys, jps.config.MaxFieldCount),
			Severity: "warning",
		})
	}

	// Check nesting depth
	if result.Stats.MaxDepth > jps.config.MaxNestingDepth {
		result.Warnings = append(result.Warnings, ValidationError{
			Type:     "complexity",
			Message:  fmt.Sprintf("Nesting depth (%d) exceeds maximum (%d)", result.Stats.MaxDepth, jps.config.MaxNestingDepth),
			Severity: "warning",
		})
	}

	// Check for potential secrets
	if jps.config.StripSecrets {
		secrets := jps.detectSecrets(content)
		for _, secret := range secrets {
			result.Warnings = append(result.Warnings, ValidationError{
				Line:     secret.Line,
				Column:   secret.Column,
				Type:     "security",
				Message:  "Potential secret detected",
				Path:     secret.Path,
				Severity: "warning",
			})
		}
	}

	return result
}

// FormatJSON formats JSON content
func (jps *JSONPayloadStudio) FormatJSON(content string) (string, error) {
	var parsed interface{}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	formatted, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}

// LoadTemplate loads a template by ID
func (jps *JSONPayloadStudio) LoadTemplate(templateID string) (*Template, error) {
	jps.mu.RLock()
	defer jps.mu.RUnlock()

	template, exists := jps.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template, nil
}

// SearchTemplates searches for templates
func (jps *JSONPayloadStudio) SearchTemplates(filter *TemplateFilter) *SearchResult {
	jps.mu.RLock()
	defer jps.mu.RUnlock()

	result := &SearchResult{
		Templates: make([]Template, 0),
		Query:     filter.Query,
		Filters:   make([]string, 0),
	}

	for _, template := range jps.templates {
		if jps.matchesFilter(template, filter) {
			result.Templates = append(result.Templates, *template)
		}

		if filter.MaxResults > 0 && len(result.Templates) >= filter.MaxResults {
			result.HasMore = true
			break
		}
	}

	// Sort by relevance
	if filter.Query != "" {
		sort.Slice(result.Templates, func(i, j int) bool {
			return fuzzyScore(result.Templates[i].Name, filter.Query) >
				fuzzyScore(result.Templates[j].Name, filter.Query)
		})
	}

	result.TotalCount = len(result.Templates)
	return result
}

// PreviewTemplate generates a preview of a template
func (jps *JSONPayloadStudio) PreviewTemplate(templateID string, variables map[string]interface{}) (*PreviewData, error) {
	template, err := jps.LoadTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Apply variables
	content := jps.applyVariables(template.Content, variables)

	// Format
	formatted, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return nil, err
	}

	preview := &PreviewData{
		Content:   content,
		Formatted: string(formatted),
		Size:      len(formatted),
		LineCount: strings.Count(string(formatted), "\n") + 1,
		Valid:     true,
	}

	// Truncate if needed
	if jps.config.PreviewLines > 0 && preview.LineCount > jps.config.PreviewLines {
		lines := strings.Split(preview.Formatted, "\n")
		preview.Formatted = strings.Join(lines[:jps.config.PreviewLines], "\n") + "\n..."
		preview.Truncated = true
	}

	return preview, nil
}

// GetCompletions returns auto-completion suggestions for the provided context.
func (jps *JSONPayloadStudio) GetCompletions(document string, position *Position, schema *JSONSchema) []CompletionItem {
	jps.mu.RLock()
	defer jps.mu.RUnlock()

	pos := Position{Line: 1, Column: 1}
	if position != nil {
		pos = *position
	}

	ctx := jps.getContextAtPosition(document, pos)
	completions := make([]CompletionItem, 0)

	if schema != nil {
		completions = append(completions, jps.getSchemaCompletions(schema, ctx)...)
	}

	for trigger, snippet := range jps.snippets {
		if strings.HasPrefix(trigger, ctx.Prefix) {
			completions = append(completions, CompletionItem{
				Label:      snippet.Name,
				Kind:       "snippet",
				Detail:     snippet.Description,
				InsertText: jps.expandSnippet(snippet),
				FilterText: trigger,
			})
		}
	}

	if jps.lastEnqueued != nil {
		completions = append(completions, jps.getRecentCompletions(jps.lastEnqueued.Payload, ctx)...)
	}

	return completions
}

// InsertSnippet inserts a snippet at the current position
func (jps *JSONPayloadStudio) InsertSnippet(sessionID string, snippetID string) error {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	session, exists := jps.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	snippet, exists := jps.snippets[snippetID]
	if !exists {
		return fmt.Errorf("snippet not found")
	}

	// Expand snippet
	expanded := jps.expandSnippet(snippet)

	// Insert at cursor position
	state := session.EditorState
	before := state.Content[:getCursorOffset(state.Content, state.CursorLine, state.CursorColumn)]
	after := state.Content[getCursorOffset(state.Content, state.CursorLine, state.CursorColumn):]
	state.Content = before + expanded + after
	state.Modified = true

	// Update cursor position
	lines := strings.Split(expanded, "\n")
	if len(lines) > 1 {
		state.CursorLine += len(lines) - 1
		state.CursorColumn = len(lines[len(lines)-1]) + 1
	} else {
		state.CursorColumn += len(expanded)
	}

	return nil
}

// EnqueuePayload enqueues a job with the current payload
func (jps *JSONPayloadStudio) EnqueuePayload(sessionID string, options *EnqueueOptions) (*EnqueueResult, error) {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	session, exists := jps.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Parse payload
	var payload interface{}
	if err := json.Unmarshal([]byte(session.EditorState.Content), &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Strip secrets if configured
	if jps.config.StripSecrets {
		payload = jps.stripSecrets(payload)
	}

	// Validate size
	payloadBytes, _ := json.Marshal(payload)
	if len(payloadBytes) > jps.config.MaxPayloadSize {
		return nil, fmt.Errorf("payload too large: %d bytes (max: %d)", len(payloadBytes), jps.config.MaxPayloadSize)
	}

	// Generate job IDs
	jobIDs := make([]string, options.Count)
	for i := 0; i < options.Count; i++ {
		jobIDs[i] = uuid.New().String()
	}

	// Enqueue to Redis
	ctx := context.Background()
	pipe := jps.redis.Pipeline()

	for i := 0; i < options.Count; i++ {
		job := map[string]interface{}{
			"id":         jobIDs[i],
			"payload":    payload,
			"priority":   options.Priority,
			"created_at": time.Now().Unix(),
			"metadata":   options.Metadata,
		}

		if options.MaxRetries > 0 {
			job["max_retries"] = options.MaxRetries
		}

		if options.TTL > 0 {
			job["ttl"] = options.TTL.Seconds()
		}

		jobData, _ := json.Marshal(job)

		// Handle scheduling
		if options.RunAt != nil {
			// Scheduled job
			score := float64(options.RunAt.Unix())
			pipe.ZAdd(ctx, fmt.Sprintf("scheduled:%s", options.Queue), redis.Z{
				Score:  score,
				Member: string(jobData),
			})
		} else if options.Delay > 0 {
			// Delayed job
			runAt := time.Now().Add(options.Delay)
			score := float64(runAt.Unix())
			pipe.ZAdd(ctx, fmt.Sprintf("delayed:%s", options.Queue), redis.Z{
				Score:  score,
				Member: string(jobData),
			})
		} else {
			// Immediate job
			if options.Priority > 0 {
				pipe.ZAdd(ctx, fmt.Sprintf("priority:%s", options.Queue), redis.Z{
					Score:  float64(options.Priority),
					Member: string(jobData),
				})
			} else {
				pipe.RPush(ctx, fmt.Sprintf("queue:%s", options.Queue), string(jobData))
			}
		}
	}

	// Handle cron scheduling
	var cronJobID string
	if options.CronSpec != "" {
		cronJobID = uuid.New().String()
		cronJob := map[string]interface{}{
			"id":       cronJobID,
			"spec":     options.CronSpec,
			"queue":    options.Queue,
			"payload":  payload,
			"priority": options.Priority,
			"metadata": options.Metadata,
		}
		cronData, _ := json.Marshal(cronJob)
		pipe.HSet(ctx, "cron:jobs", cronJobID, string(cronData))
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to enqueue: %w", err)
	}

	result := &EnqueueResult{
		JobIDs:      jobIDs,
		Queue:       options.Queue,
		Count:       options.Count,
		Priority:    options.Priority,
		RunAt:       options.RunAt,
		CronJobID:   cronJobID,
		Payload:     payload,
		PayloadSize: len(payloadBytes),
		EnqueuedAt:  time.Now(),
	}

	// Update session
	session.JobsEnqueued += options.Count
	jps.lastEnqueued = result

	jps.logger.Info("Payload enqueued",
		zap.String("session", sessionID),
		zap.String("queue", options.Queue),
		zap.Int("count", options.Count))

	return result, nil
}

// GetDiff compares current editor content with last enqueued payload
func (jps *JSONPayloadStudio) GetDiff(sessionID string) (*DiffResult, error) {
	jps.mu.RLock()
	defer jps.mu.RUnlock()

	session, exists := jps.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if jps.lastEnqueued == nil {
		return &DiffResult{
			HasChanges: false,
			Summary:    "No previous payload to compare",
		}, nil
	}

	// Parse current content
	var current interface{}
	if err := json.Unmarshal([]byte(session.EditorState.Content), &current); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Compare
	diff := jps.compareJSON(jps.lastEnqueued.Payload, current)
	return diff, nil
}

// Undo undoes the last edit
func (jps *JSONPayloadStudio) Undo(sessionID string) error {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	session, exists := jps.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	state := session.EditorState
	if state.HistoryIndex > 0 {
		state.HistoryIndex--
		state.Content = state.History[state.HistoryIndex]
		state.Modified = true
	}

	return nil
}

// Redo redoes the last undone edit
func (jps *JSONPayloadStudio) Redo(sessionID string) error {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	session, exists := jps.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	state := session.EditorState
	if state.HistoryIndex < len(state.History)-1 {
		state.HistoryIndex++
		state.Content = state.History[state.HistoryIndex]
		state.Modified = true
	}

	return nil
}

// SaveTemplateFromSession saves the current content as a template
func (jps *JSONPayloadStudio) SaveTemplateFromSession(sessionID string, name, description, category string, tags []string) (*Template, error) {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	session, exists := jps.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Parse content
	var content map[string]interface{}
	if err := json.Unmarshal([]byte(session.EditorState.Content), &content); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	template := &Template{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Category:    category,
		Tags:        tags,
		Content:     content,
		Schema:      session.EditorState.Schema,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Author:      os.Getenv("USER"),
		Version:     "1.0.0",
	}

	// Extract variables
	template.Variables = jps.extractVariables(content)

	// Save to disk
	if err := jps.saveTemplateToDisk(template); err != nil {
		return nil, err
	}

	jps.templates[template.ID] = template

	// Track template usage
	session.Templates = append(session.Templates, template.ID)

	return template, nil
}

// CloseSession closes an editor session
func (jps *JSONPayloadStudio) CloseSession(sessionID string) error {
	jps.mu.Lock()
	defer jps.mu.Unlock()

	delete(jps.sessions, sessionID)
	return nil
}

// Helper methods

func (jps *JSONPayloadStudio) loadTemplates() error {
	// Load from templates directory
	if jps.config.TemplatesPath == "" {
		return nil
	}

	return filepath.Walk(jps.config.TemplatesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var template Template
		if err := json.Unmarshal(data, &template); err != nil {
			jps.logger.Warn("Failed to parse template", zap.String("path", path), zap.Error(err))
			return nil
		}

		if template.ID == "" {
			template.ID = uuid.New().String()
		}

		jps.templates[template.ID] = &template
		return nil
	})
}

func (jps *JSONPayloadStudio) loadSchemas() error {
	if jps.config.SchemasPath == "" {
		return nil
	}

	return filepath.Walk(jps.config.SchemasPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var schema JSONSchema
		if err := json.Unmarshal(data, &schema); err != nil {
			jps.logger.Warn("Failed to parse schema", zap.String("path", path), zap.Error(err))
			return nil
		}

		if schema.ID == "" {
			schema.ID = strings.TrimSuffix(filepath.Base(path), ".json")
		}

		jps.schemas[schema.ID] = &schema
		return nil
	})
}

func (jps *JSONPayloadStudio) initializeSnippets() {
	// Default snippets
	jps.snippets["now"] = &Snippet{
		ID:          "now",
		Name:        "Current Timestamp",
		Trigger:     "now",
		Description: "Insert current timestamp",
		Category:    "time",
		Expansion:   `"${TIMESTAMP}"`,
	}

	jps.snippets["uuid"] = &Snippet{
		ID:          "uuid",
		Name:        "UUID",
		Trigger:     "uuid",
		Description: "Insert a new UUID",
		Category:    "id",
		Expansion:   `"${UUID}"`,
	}

	jps.snippets["date"] = &Snippet{
		ID:          "date",
		Name:        "Current Date",
		Trigger:     "date",
		Description: "Insert current date",
		Category:    "time",
		Expansion:   `"${DATE}"`,
	}

	jps.snippets["user"] = &Snippet{
		ID:          "user",
		Name:        "User Info",
		Trigger:     "user",
		Description: "Insert user information",
		Category:    "user",
		Content: map[string]interface{}{
			"user_id": "${USER_ID}",
			"email":   "${USER_EMAIL}",
			"name":    "${USER_NAME}",
		},
	}
}

func (jps *JSONPayloadStudio) expandSnippet(snippet *Snippet) string {
	if snippet.Expansion != "" {
		expanded := snippet.Expansion
		expanded = strings.ReplaceAll(expanded, "${TIMESTAMP}", fmt.Sprintf("%d", time.Now().Unix()))
		expanded = strings.ReplaceAll(expanded, "${UUID}", uuid.New().String())
		expanded = strings.ReplaceAll(expanded, "${DATE}", time.Now().Format("2006-01-02"))
		expanded = strings.ReplaceAll(expanded, "${USER_ID}", uuid.New().String())
		expanded = strings.ReplaceAll(expanded, "${USER_EMAIL}", "user@example.com")
		expanded = strings.ReplaceAll(expanded, "${USER_NAME}", os.Getenv("USER"))
		return expanded
	}

	if snippet.Content != nil {
		data, _ := json.MarshalIndent(snippet.Content, "", "  ")
		return string(data)
	}

	return ""
}

func (jps *JSONPayloadStudio) validateAgainstSchema(data interface{}, schema *JSONSchema) []ValidationError {
	errors := make([]ValidationError, 0)

	// Convert to JSON for schema validation
	dataJSON, _ := json.Marshal(data)
	schemaJSON, _ := json.Marshal(schema)

	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	documentLoader := gojsonschema.NewBytesLoader(dataJSON)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		errors = append(errors, ValidationError{
			Type:     "schema",
			Message:  fmt.Sprintf("Schema validation error: %v", err),
			Severity: "error",
		})
		return errors
	}

	if !result.Valid() {
		for _, err := range result.Errors() {
			// Try to get line number from path
			line := 1
			if err.Field() != "" {
				line = getFieldLine(string(dataJSON), err.Field())
			}

			errors = append(errors, ValidationError{
				Line:       line,
				Type:       "schema",
				Message:    err.Description(),
				Path:       err.Field(),
				SchemaPath: err.Context().String(),
				Severity:   "error",
			})
		}
	}

	return errors
}

func (jps *JSONPayloadStudio) detectSecrets(content string) []ValidationError {
	secrets := make([]ValidationError, 0)

	// Default secret patterns
	patterns := []string{
		`"[^"]*(?:password|secret|token|key|api_key|apikey|auth|credential)[^"]*"\s*:\s*"[^"]+?"`,
		`"[^"]*"\s*:\s*"(?:sk_|pk_|api_|key_|secret_|token_)[^"]+?"`,
		`"[^"]*"\s*:\s*"[A-Za-z0-9+/]{40,}"`, // Base64 encoded strings
	}

	// Add custom patterns
	patterns = append(patterns, jps.config.SecretPatterns...)

	for _, pattern := range patterns {
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			continue
		}

		matches := re.FindAllStringIndex(content, -1)
		for _, match := range matches {
			line, col := getLineColumn(content, match[0])
			secrets = append(secrets, ValidationError{
				Line:     line,
				Column:   col,
				Type:     "security",
				Message:  "Potential secret detected",
				Severity: "warning",
			})
		}
	}

	return secrets
}

func (jps *JSONPayloadStudio) stripSecrets(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			lowerKey := strings.ToLower(key)
			if strings.Contains(lowerKey, "password") ||
				strings.Contains(lowerKey, "secret") ||
				strings.Contains(lowerKey, "token") ||
				strings.Contains(lowerKey, "key") ||
				strings.Contains(lowerKey, "auth") {
				result[key] = "***REDACTED***"
			} else {
				result[key] = jps.stripSecrets(value)
			}
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = jps.stripSecrets(item)
		}
		return result

	default:
		return v
	}
}

func (jps *JSONPayloadStudio) matchesFilter(template *Template, filter *TemplateFilter) bool {
	// Check query
	if filter.Query != "" {
		query := strings.ToLower(filter.Query)
		if !strings.Contains(strings.ToLower(template.Name), query) &&
			!strings.Contains(strings.ToLower(template.Description), query) {
			return false
		}
	}

	// Check categories
	if len(filter.Categories) > 0 {
		found := false
		for _, cat := range filter.Categories {
			if template.Category == cat {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check tags
	if len(filter.Tags) > 0 {
		found := false
		for _, filterTag := range filter.Tags {
			for _, templateTag := range template.Tags {
				if filterTag == templateTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check author
	if filter.Author != "" && template.Author != filter.Author {
		return false
	}

	return true
}

func (jps *JSONPayloadStudio) applyVariables(content map[string]interface{}, variables map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range content {
		switch v := value.(type) {
		case string:
			// Replace variables
			for varName, varValue := range variables {
				placeholder := fmt.Sprintf("${%s}", varName)
				v = strings.ReplaceAll(v, placeholder, fmt.Sprintf("%v", varValue))
			}
			result[key] = v

		case map[string]interface{}:
			result[key] = jps.applyVariables(v, variables)

		case []interface{}:
			arr := make([]interface{}, len(v))
			for i, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					arr[i] = jps.applyVariables(m, variables)
				} else {
					arr[i] = item
				}
			}
			result[key] = arr

		default:
			result[key] = value
		}
	}

	return result
}

func (jps *JSONPayloadStudio) extractVariables(content map[string]interface{}) []TemplateVariable {
	variables := make([]TemplateVariable, 0)
	varMap := make(map[string]bool)

	jps.findVariables(content, &varMap)

	for varName := range varMap {
		variables = append(variables, TemplateVariable{
			Name:     varName,
			Type:     "string",
			Required: false,
		})
	}

	return variables
}

func (jps *JSONPayloadStudio) findVariables(data interface{}, varMap *map[string]bool) {
	switch v := data.(type) {
	case string:
		// Find ${VAR_NAME} patterns
		re := regexp.MustCompile(`\$\{([^}]+)\}`)
		matches := re.FindAllStringSubmatch(v, -1)
		for _, match := range matches {
			if len(match) > 1 {
				(*varMap)[match[1]] = true
			}
		}

	case map[string]interface{}:
		for _, value := range v {
			jps.findVariables(value, varMap)
		}

	case []interface{}:
		for _, item := range v {
			jps.findVariables(item, varMap)
		}
	}
}

func (jps *JSONPayloadStudio) compareJSON(old, new interface{}) *DiffResult {
	diff := &DiffResult{
		HasChanges: false,
		Added:      make([]DiffChange, 0),
		Removed:    make([]DiffChange, 0),
		Modified:   make([]DiffChange, 0),
	}

	jps.compareValues(old, new, "", diff)

	diff.HasChanges = len(diff.Added) > 0 || len(diff.Removed) > 0 || len(diff.Modified) > 0

	if diff.HasChanges {
		diff.Summary = fmt.Sprintf("%d added, %d removed, %d modified",
			len(diff.Added), len(diff.Removed), len(diff.Modified))
	} else {
		diff.Summary = "No changes"
	}

	return diff
}

func (jps *JSONPayloadStudio) compareValues(old, new interface{}, path string, diff *DiffResult) {
	// Handle nil values
	if old == nil && new == nil {
		return
	}
	if old == nil {
		diff.Added = append(diff.Added, DiffChange{
			Path:     path,
			Type:     "added",
			NewValue: new,
		})
		return
	}
	if new == nil {
		diff.Removed = append(diff.Removed, DiffChange{
			Path:     path,
			Type:     "removed",
			OldValue: old,
		})
		return
	}

	// Compare based on type
	switch oldVal := old.(type) {
	case map[string]interface{}:
		if newVal, ok := new.(map[string]interface{}); ok {
			// Compare objects
			for key, oldValue := range oldVal {
				newPath := path
				if newPath == "" {
					newPath = key
				} else {
					newPath = path + "." + key
				}

				if newValue, exists := newVal[key]; exists {
					jps.compareValues(oldValue, newValue, newPath, diff)
				} else {
					diff.Removed = append(diff.Removed, DiffChange{
						Path:     newPath,
						Type:     "removed",
						OldValue: oldValue,
					})
				}
			}

			for key, newValue := range newVal {
				if _, exists := oldVal[key]; !exists {
					newPath := path
					if newPath == "" {
						newPath = key
					} else {
						newPath = path + "." + key
					}
					diff.Added = append(diff.Added, DiffChange{
						Path:     newPath,
						Type:     "added",
						NewValue: newValue,
					})
				}
			}
		} else {
			diff.Modified = append(diff.Modified, DiffChange{
				Path:     path,
				Type:     "modified",
				OldValue: old,
				NewValue: new,
			})
		}

	case []interface{}:
		if newVal, ok := new.([]interface{}); ok {
			// Compare arrays
			for i := 0; i < len(oldVal) && i < len(newVal); i++ {
				newPath := fmt.Sprintf("%s[%d]", path, i)
				jps.compareValues(oldVal[i], newVal[i], newPath, diff)
			}

			// Handle different lengths
			if len(oldVal) > len(newVal) {
				for i := len(newVal); i < len(oldVal); i++ {
					diff.Removed = append(diff.Removed, DiffChange{
						Path:     fmt.Sprintf("%s[%d]", path, i),
						Type:     "removed",
						OldValue: oldVal[i],
					})
				}
			} else if len(newVal) > len(oldVal) {
				for i := len(oldVal); i < len(newVal); i++ {
					diff.Added = append(diff.Added, DiffChange{
						Path:     fmt.Sprintf("%s[%d]", path, i),
						Type:     "added",
						NewValue: newVal[i],
					})
				}
			}
		} else {
			diff.Modified = append(diff.Modified, DiffChange{
				Path:     path,
				Type:     "modified",
				OldValue: old,
				NewValue: new,
			})
		}

	default:
		// Compare primitives
		if !reflect.DeepEqual(old, new) {
			diff.Modified = append(diff.Modified, DiffChange{
				Path:     path,
				Type:     "modified",
				OldValue: old,
				NewValue: new,
			})
		}
	}
}

func (jps *JSONPayloadStudio) saveTemplateToDisk(template *Template) error {
	if jps.config.TemplatesPath == "" {
		return nil
	}

	// Create templates directory if needed
	if err := os.MkdirAll(jps.config.TemplatesPath, 0755); err != nil {
		return err
	}

	// Save template
	filename := filepath.Join(jps.config.TemplatesPath, fmt.Sprintf("%s.json", template.ID))
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (jps *JSONPayloadStudio) autoSaveSession(sessionID string) {
	ticker := time.NewTicker(jps.config.AutoSaveInterval)
	defer ticker.Stop()

	for range ticker.C {
		jps.mu.RLock()
		session, exists := jps.sessions[sessionID]
		jps.mu.RUnlock()

		if !exists {
			return
		}

		if session.EditorState.Modified {
			// Auto-save to Redis
			ctx := context.Background()
			data, _ := json.Marshal(session)
			jps.redis.Set(ctx, fmt.Sprintf("studio:session:%s", sessionID), string(data), 24*time.Hour)

			jps.mu.Lock()
			session.AutoSaved = true
			session.EditorState.Modified = false
			jps.mu.Unlock()
		}
	}
}

// Context for completions
type completionContext struct {
	Prefix string
	Path   string
	InKey  bool
}

func (jps *JSONPayloadStudio) getContextAtPosition(content string, position Position) *completionContext {
	// Simplified context detection
	offset := getCursorOffset(content, position.Line, position.Column)

	// Find start of current token
	start := offset - 1
	for start >= 0 && !strings.ContainsRune(`{}[],:"\n\t `, rune(content[start])) {
		start--
	}
	start++

	prefix := ""
	if start < offset {
		prefix = content[start:offset]
	}

	return &completionContext{
		Prefix: prefix,
		InKey:  strings.Count(content[:offset], `"`)%2 == 1,
	}
}

func (jps *JSONPayloadStudio) getSchemaCompletions(schema *JSONSchema, context *completionContext) []CompletionItem {
	completions := make([]CompletionItem, 0)

	if schema.Properties != nil && context.InKey {
		for key, propSchema := range schema.Properties {
			if strings.HasPrefix(key, context.Prefix) {
				detail := ""
				if propMap, ok := propSchema.(map[string]interface{}); ok {
					if desc, ok := propMap["description"].(string); ok {
						detail = desc
					}
				}

				completions = append(completions, CompletionItem{
					Label:      key,
					Kind:       "property",
					Detail:     detail,
					InsertText: fmt.Sprintf(`"%s": `, key),
					FilterText: key,
				})
			}
		}
	}

	return completions
}

func (jps *JSONPayloadStudio) getRecentCompletions(payload interface{}, context *completionContext) []CompletionItem {
	completions := make([]CompletionItem, 0)

	// Extract field names from recent payload
	if m, ok := payload.(map[string]interface{}); ok {
		for key := range m {
			if strings.HasPrefix(key, context.Prefix) {
				completions = append(completions, CompletionItem{
					Label:      key,
					Kind:       "value",
					Detail:     "Recent field",
					InsertText: fmt.Sprintf(`"%s"`, key),
					FilterText: key,
				})
			}
		}
	}

	return completions
}

// Utility functions

func getLineColumn(content string, offset int) (int, int) {
	line := 1
	column := 1

	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}

	return line, column
}

func getCursorOffset(content string, line, column int) int {
	offset := 0
	currentLine := 1

	for i := 0; i < len(content); i++ {
		if currentLine == line && column == 1 {
			return i
		}

		if content[i] == '\n' {
			currentLine++
			if currentLine == line {
				return i + column
			}
		}
	}

	return offset
}

func getFieldLine(content, field string) int {
	// Simple field line detection
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, fmt.Sprintf(`"%s"`, field)) {
			return i + 1
		}
	}
	return 1
}

func calculateStats(data interface{}, content string) LintStats {
	stats := LintStats{
		Lines:      strings.Count(content, "\n") + 1,
		Characters: len(content),
	}

	calculateStatsRecursive(data, 0, &stats)
	return stats
}

func calculateStatsRecursive(data interface{}, depth int, stats *LintStats) {
	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	switch v := data.(type) {
	case map[string]interface{}:
		stats.ObjectCount++
		stats.Keys += len(v)
		for _, value := range v {
			calculateStatsRecursive(value, depth+1, stats)
		}

	case []interface{}:
		stats.ArrayCount++
		for _, item := range v {
			calculateStatsRecursive(item, depth+1, stats)
		}

	case string:
		stats.StringCount++

	case float64, int, int64:
		stats.NumberCount++

	case bool:
		stats.BooleanCount++

	case nil:
		stats.NullCount++
	}
}

func fuzzyScore(s, query string) int {
	s = strings.ToLower(s)
	query = strings.ToLower(query)

	if s == query {
		return 100
	}

	if strings.Contains(s, query) {
		return 50 + (50 - len(s) + len(query))
	}

	// Simple fuzzy matching
	score := 0
	queryIndex := 0
	for i := 0; i < len(s) && queryIndex < len(query); i++ {
		if s[i] == query[queryIndex] {
			score += 10
			queryIndex++
		}
	}

	if queryIndex == len(query) {
		return score
	}

	return 0
}
