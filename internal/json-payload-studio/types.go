// Copyright 2025 James Ross
package jsonpayloadstudio

import (
	"time"
)

// ValidationError represents a JSON validation error
type ValidationError struct {
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Type        string `json:"type"`
	Message     string `json:"message"`
	Path        string `json:"path,omitempty"`
	SchemaPath  string `json:"schema_path,omitempty"`
	Severity    string `json:"severity"` // error, warning, info
}

// Template represents a JSON payload template
type Template struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Content     map[string]interface{} `json:"content"`
	Schema      *JSONSchema            `json:"schema,omitempty"`
	Variables   []TemplateVariable     `json:"variables,omitempty"`
	Snippets    []Snippet              `json:"snippets,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Author      string                 `json:"author"`
	Version     string                 `json:"version"`
}

// TemplateVariable represents a variable in a template
type TemplateVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Required     bool        `json:"required"`
	Pattern      string      `json:"pattern,omitempty"`
	Options      []string    `json:"options,omitempty"`
}

// Snippet represents a code snippet
type Snippet struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Trigger     string                 `json:"trigger"`
	Description string                 `json:"description"`
	Content     interface{}            `json:"content"`
	Variables   []string               `json:"variables,omitempty"`
	Category    string                 `json:"category"`
	Expansion   string                 `json:"expansion,omitempty"`
}

// JSONSchema represents a JSON Schema for validation
type JSONSchema struct {
	ID          string                 `json:"id,omitempty"`
	Schema      string                 `json:"$schema,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        interface{}            `json:"type"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Definitions map[string]interface{} `json:"definitions,omitempty"`
	Additional  map[string]interface{} `json:",inline"` // For additional schema properties
}

// EditorState represents the current state of the JSON editor
type EditorState struct {
	Content       string              `json:"content"`
	CursorLine    int                 `json:"cursor_line"`
	CursorColumn  int                 `json:"cursor_column"`
	SelectionStart *Position          `json:"selection_start,omitempty"`
	SelectionEnd   *Position          `json:"selection_end,omitempty"`
	Errors        []ValidationError   `json:"errors"`
	Warnings      []ValidationError   `json:"warnings"`
	Modified      bool                `json:"modified"`
	Template      *Template           `json:"template,omitempty"`
	Schema        *JSONSchema         `json:"schema,omitempty"`
	History       []string            `json:"history,omitempty"`
	HistoryIndex  int                 `json:"history_index"`
}

// Position represents a position in the editor
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// EnqueueOptions represents options for enqueuing a job
type EnqueueOptions struct {
	Queue       string        `json:"queue"`
	Count       int           `json:"count"`
	Priority    int           `json:"priority"`
	Delay       time.Duration `json:"delay,omitempty"`
	RunAt       *time.Time    `json:"run_at,omitempty"`
	CronSpec    string        `json:"cron_spec,omitempty"`
	TTL         time.Duration `json:"ttl,omitempty"`
	MaxRetries  int           `json:"max_retries"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// EnqueueResult represents the result of enqueuing jobs
type EnqueueResult struct {
	JobIDs       []string          `json:"job_ids"`
	Queue        string            `json:"queue"`
	Count        int               `json:"count"`
	Priority     int               `json:"priority"`
	RunAt        *time.Time        `json:"run_at,omitempty"`
	CronJobID    string            `json:"cron_job_id,omitempty"`
	Payload      interface{}       `json:"payload"`
	PayloadSize  int               `json:"payload_size"`
	EnqueuedAt   time.Time         `json:"enqueued_at"`
}

// DiffResult represents the difference between two JSON payloads
type DiffResult struct {
	HasChanges   bool          `json:"has_changes"`
	Added        []DiffChange  `json:"added"`
	Removed      []DiffChange  `json:"removed"`
	Modified     []DiffChange  `json:"modified"`
	Summary      string        `json:"summary"`
}

// DiffChange represents a single change in a diff
type DiffChange struct {
	Path     string      `json:"path"`
	Type     string      `json:"type"` // added, removed, modified
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
	Line     int         `json:"line,omitempty"`
}

// SearchResult represents a template search result
type SearchResult struct {
	Templates    []Template `json:"templates"`
	TotalCount   int        `json:"total_count"`
	Query        string     `json:"query"`
	Filters      []string   `json:"filters,omitempty"`
	HasMore      bool       `json:"has_more"`
}

// StudioConfig represents the configuration for the JSON Payload Studio
type StudioConfig struct {
	// Editor settings
	EditorTheme      string `json:"editor_theme"`
	SyntaxHighlight  bool   `json:"syntax_highlight"`
	LineNumbers      bool   `json:"line_numbers"`
	AutoFormat       bool   `json:"auto_format"`
	BracketMatching  bool   `json:"bracket_matching"`
	AutoComplete     bool   `json:"auto_complete"`
	ValidateOnType   bool   `json:"validate_on_type"`

	// Template settings
	TemplatesPath    string   `json:"templates_path"`
	AutoLoadTemplates bool    `json:"auto_load_templates"`
	TemplateDirs     []string `json:"template_dirs"`

	// Schema settings
	SchemasPath      string   `json:"schemas_path"`
	DefaultSchema    string   `json:"default_schema,omitempty"`
	StrictValidation bool     `json:"strict_validation"`

	// Safety settings
	MaxPayloadSize   int      `json:"max_payload_size"`
	MaxFieldCount    int      `json:"max_field_count"`
	MaxNestingDepth  int      `json:"max_nesting_depth"`
	StripSecrets     bool     `json:"strip_secrets"`
	SecretPatterns   []string `json:"secret_patterns"`
	RequireConfirm   bool     `json:"require_confirm"`

	// UI settings
	ShowPreview      bool     `json:"show_preview"`
	PreviewLines     int      `json:"preview_lines"`
	HistorySize      int      `json:"history_size"`
	AutoSave         bool     `json:"auto_save"`
	AutoSaveInterval time.Duration `json:"auto_save_interval"`
}

// LintResult represents the result of JSON linting
type LintResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors"`
	Warnings []ValidationError `json:"warnings"`
	Info     []ValidationError `json:"info"`
	Stats    LintStats         `json:"stats"`
}

// LintStats provides statistics about the JSON document
type LintStats struct {
	Lines         int `json:"lines"`
	Characters    int `json:"characters"`
	Keys          int `json:"keys"`
	MaxDepth      int `json:"max_depth"`
	ArrayCount    int `json:"array_count"`
	ObjectCount   int `json:"object_count"`
	StringCount   int `json:"string_count"`
	NumberCount   int `json:"number_count"`
	BooleanCount  int `json:"boolean_count"`
	NullCount     int `json:"null_count"`
}

// CompletionItem represents an auto-completion suggestion
type CompletionItem struct {
	Label       string `json:"label"`
	Kind        string `json:"kind"` // property, value, snippet
	Detail      string `json:"detail,omitempty"`
	InsertText  string `json:"insert_text"`
	SortText    string `json:"sort_text,omitempty"`
	FilterText  string `json:"filter_text,omitempty"`
	Preselect   bool   `json:"preselect,omitempty"`
}

// ScheduleInfo represents scheduling information for a job
type ScheduleInfo struct {
	Type        string        `json:"type"` // immediate, delayed, scheduled, cron
	Delay       time.Duration `json:"delay,omitempty"`
	RunAt       *time.Time    `json:"run_at,omitempty"`
	CronSpec    string        `json:"cron_spec,omitempty"`
	NextRun     *time.Time    `json:"next_run,omitempty"`
	Occurrences int           `json:"occurrences,omitempty"`
}

// PreviewData represents preview data for a template or payload
type PreviewData struct {
	Content     interface{} `json:"content"`
	Formatted   string      `json:"formatted"`
	Size        int         `json:"size"`
	LineCount   int         `json:"line_count"`
	Truncated   bool        `json:"truncated"`
	Valid       bool        `json:"valid"`
	Error       string      `json:"error,omitempty"`
}

// SessionInfo represents an editor session
type SessionInfo struct {
	ID           string          `json:"id"`
	StartedAt    time.Time       `json:"started_at"`
	LastActivity time.Time       `json:"last_activity"`
	EditorState  *EditorState    `json:"editor_state"`
	JobsEnqueued int             `json:"jobs_enqueued"`
	Templates    []string        `json:"templates_used"`
	AutoSaved    bool            `json:"auto_saved"`
}

// TemplateFilter represents filters for template search
type TemplateFilter struct {
	Query       string   `json:"query,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Author      string   `json:"author,omitempty"`
	MaxResults  int      `json:"max_results,omitempty"`
}

// ValidationRule represents a custom validation rule
type ValidationRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Type        string `json:"type"` // regex, range, enum, custom
	Rule        string `json:"rule"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
}

// EditorAction represents an action in the editor
type EditorAction string

const (
	ActionFormat      EditorAction = "format"
	ActionValidate    EditorAction = "validate"
	ActionEnqueue     EditorAction = "enqueue"
	ActionSave        EditorAction = "save"
	ActionLoad        EditorAction = "load"
	ActionUndo        EditorAction = "undo"
	ActionRedo        EditorAction = "redo"
	ActionComplete    EditorAction = "complete"
	ActionInsertSnippet EditorAction = "insert_snippet"
)

// EditorEvent represents an event in the editor
type EditorEvent struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Action    EditorAction `json:"action,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}