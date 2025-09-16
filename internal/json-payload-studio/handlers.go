package jsonpayloadstudio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Handler provides HTTP handlers for the JSON Payload Studio
type Handler struct {
	studio *JSONPayloadStudio
}

// NewHandler creates a new HTTP handler for the JSON Payload Studio
func NewHandler(studio *JSONPayloadStudio) *Handler {
	return &Handler{
		studio: studio,
	}
}

// HandleValidate handles JSON validation requests
func (h *Handler) HandleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content  string      `json:"content"`
		SchemaID string      `json:"schema_id,omitempty"`
		Schema   *JSONSchema `json:"schema,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var schema *JSONSchema
	if req.SchemaID != "" {
		s, err := h.studio.GetSchema(req.SchemaID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Schema not found: %v", err), http.StatusNotFound)
			return
		}
		schema = s
	} else if req.Schema != nil {
		schema = req.Schema
	}

	result := h.studio.ValidateJSON(req.Content, schema)
	h.sendJSON(w, result)
}

// HandleFormat handles JSON formatting requests
func (h *Handler) HandleFormat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
		Indent  string `json:"indent,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	formatted, err := h.studio.FormatJSON(req.Content)
	if err != nil {
		http.Error(w, fmt.Sprintf("Format error: %v", err), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, map[string]string{
		"formatted": formatted,
	})
}

// HandleTemplates handles template operations
func (h *Handler) HandleTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleListTemplates(w, r)
	case http.MethodPost:
		h.handleSaveTemplate(w, r)
	case http.MethodDelete:
		h.handleDeleteTemplate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if search := query.Get("search"); search != "" {
		filter := &TemplateFilter{
			Query:      search,
			MaxResults: 50,
		}

		if categories := query.Get("categories"); categories != "" {
			filter.Categories = strings.Split(categories, ",")
		}

		if tags := query.Get("tags"); tags != "" {
			filter.Tags = strings.Split(tags, ",")
		}

		result := h.studio.SearchTemplates(filter)
		h.sendJSON(w, result)
	} else {
		templates := h.studio.ListTemplates()
		h.sendJSON(w, templates)
	}
}

func (h *Handler) handleSaveTemplate(w http.ResponseWriter, r *http.Request) {
	var template Template
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid template data", http.StatusBadRequest)
		return
	}

	if template.ID == "" {
		template.ID = generateID("tmpl")
	}

	template.UpdatedAt = time.Now()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}

	if err := h.studio.SaveTemplate(&template); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save template: %v", err), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, template)
}

func (h *Handler) handleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := r.URL.Query().Get("id")
	if templateID == "" {
		http.Error(w, "Template ID required", http.StatusBadRequest)
		return
	}

	if err := h.studio.DeleteTemplate(templateID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete template: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleApplyTemplate handles template application requests
func (h *Handler) HandleApplyTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TemplateID string                 `json:"template_id"`
		Variables  map[string]interface{} `json:"variables,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.studio.ApplyTemplate(req.TemplateID, req.Variables)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply template: %v", err), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, map[string]interface{}{
		"result": result,
	})
}

// HandleEnqueue handles job enqueue requests
func (h *Handler) HandleEnqueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string          `json:"session_id"`
		Options   *EnqueueOptions `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.studio.EnqueuePayload(req.SessionID, req.Options)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to enqueue: %v", err), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, result)
}

// HandleSessions handles session operations
func (h *Handler) HandleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetSession(w, r)
	case http.MethodPost:
		h.handleCreateSession(w, r)
	case http.MethodPut:
		h.handleUpdateSession(w, r)
	case http.MethodDelete:
		h.handleDeleteSession(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleGetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	session, err := h.studio.GetSession(sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Session not found: %v", err), http.StatusNotFound)
		return
	}

	h.sendJSON(w, session)
}

func (h *Handler) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	sessionID := h.studio.CreateSession()

	session, _ := h.studio.GetSession(sessionID)
	h.sendJSON(w, session)
}

func (h *Handler) handleUpdateSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	var state EditorState
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		http.Error(w, "Invalid editor state", http.StatusBadRequest)
		return
	}

	if err := h.studio.UpdateEditorState(sessionID, &state); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update session: %v", err), http.StatusInternalServerError)
		return
	}

	session, _ := h.studio.GetSession(sessionID)
	h.sendJSON(w, session)
}

func (h *Handler) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	if err := h.studio.DeleteSession(sessionID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete session: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleCompletions handles auto-completion requests
func (h *Handler) HandleCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Context  string      `json:"context"`
		Position *Position   `json:"position"`
		SchemaID string      `json:"schema_id,omitempty"`
		Schema   *JSONSchema `json:"schema,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var schema *JSONSchema
	if req.SchemaID != "" {
		s, err := h.studio.GetSchema(req.SchemaID)
		if err != nil {
			schema = nil // Schema is optional for completions
		} else {
			schema = s
		}
	} else if req.Schema != nil {
		schema = req.Schema
	}

	completions := h.studio.GetCompletions(req.Context, req.Position, schema)
	h.sendJSON(w, completions)
}

// HandleDiff handles payload diff requests
func (h *Handler) HandleDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Old interface{} `json:"old"`
		New interface{} `json:"new"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.studio.DiffPayloads(req.Old, req.New)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to diff payloads: %v", err), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, result)
}

// HandleSnippets handles snippet operations
func (h *Handler) HandleSnippets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleListSnippets(w, r)
	case http.MethodPost:
		h.handleExpandSnippet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleListSnippets(w http.ResponseWriter, r *http.Request) {
	snippets := h.studio.ListSnippets()
	h.sendJSON(w, snippets)
}

func (h *Handler) handleExpandSnippet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Trigger   string                 `json:"trigger"`
		Variables map[string]interface{} `json:"variables,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.studio.ExpandSnippet(req.Trigger, req.Variables)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to expand snippet: %v", err), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, map[string]interface{}{
		"expanded": result,
	})
}

// HandleHistory handles history operations (undo/redo)
func (h *Handler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
		Action    string `json:"action"` // "undo" or "redo"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var err error
	switch req.Action {
	case "undo":
		err = h.studio.Undo(req.SessionID)
	case "redo":
		err = h.studio.Redo(req.SessionID)
	default:
		http.Error(w, "Invalid action. Use 'undo' or 'redo'", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("History operation failed: %v", err), http.StatusBadRequest)
		return
	}

	session, _ := h.studio.GetSession(req.SessionID)
	h.sendJSON(w, session)
}

// HandlePreview handles payload preview requests
func (h *Handler) HandlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content   interface{} `json:"content"`
		MaxLines  int         `json:"max_lines,omitempty"`
		Truncate  bool        `json:"truncate,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	preview := h.studio.GeneratePreview(req.Content, req.MaxLines, req.Truncate)
	h.sendJSON(w, preview)
}

// RegisterRoutes registers all HTTP routes for the JSON Payload Studio
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/json-studio/validate", h.HandleValidate)
	mux.HandleFunc("/api/json-studio/format", h.HandleFormat)
	mux.HandleFunc("/api/json-studio/templates", h.HandleTemplates)
	mux.HandleFunc("/api/json-studio/templates/apply", h.HandleApplyTemplate)
	mux.HandleFunc("/api/json-studio/enqueue", h.HandleEnqueue)
	mux.HandleFunc("/api/json-studio/sessions", h.HandleSessions)
	mux.HandleFunc("/api/json-studio/completions", h.HandleCompletions)
	mux.HandleFunc("/api/json-studio/diff", h.HandleDiff)
	mux.HandleFunc("/api/json-studio/snippets", h.HandleSnippets)
	mux.HandleFunc("/api/json-studio/history", h.HandleHistory)
	mux.HandleFunc("/api/json-studio/preview", h.HandlePreview)
}

// Helper function to send JSON responses
func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Helper function to generate unique IDs
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}