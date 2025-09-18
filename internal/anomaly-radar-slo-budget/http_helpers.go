package anomalyradarslobudget

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type scopeKeyType string

const (
	ScopeReader = "slo_reader"
	ScopeAdmin  = "slo_admin"

	contextKeyScopes    scopeKeyType = "anomaly_radar_scopes"
	contextKeyRequestID              = "request_id"
)

// ScopeChecker evaluates whether the request context has a required scope.
type ScopeChecker func(ctx context.Context, required string) bool

// HandlerOption customises HTTPHandler behaviour.
type HandlerOption func(*HTTPHandler)

// WithScopeChecker provides a function that validates required scopes.
func WithScopeChecker(checker ScopeChecker) HandlerOption {
	return func(h *HTTPHandler) {
		h.scopeChecker = checker
	}
}

// WithNow allows overriding time source (tests).
type nowFunc func() time.Time

func WithNow(fn nowFunc) HandlerOption {
	return func(h *HTTPHandler) {
		h.now = fn
	}
}

// ContextWithScopes stores scopes in context for downstream checks.
func ContextWithScopes(ctx context.Context, scopes []string) context.Context {
	return context.WithValue(ctx, contextKeyScopes, scopes)
}

// ScopesFromContext retrieves the scopes previously stored via ContextWithScopes.
// A copy is returned to prevent callers from mutating the underlying slice.
func ScopesFromContext(ctx context.Context) []string {
	scopes := scopesFromContext(ctx)
	if len(scopes) == 0 {
		return nil
	}
	copyScopes := make([]string, len(scopes))
	copy(copyScopes, scopes)
	return copyScopes
}

func scopesFromContext(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}
	if scopes, ok := ctx.Value(contextKeyScopes).([]string); ok {
		return scopes
	}
	return nil
}

// writeJSON writes the payload as JSON with the provided status code.
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// ErrorResponse represents a standard error envelope.
type ErrorResponse struct {
	Error     string            `json:"error"`
	Code      string            `json:"code,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

func writeJSONError(w http.ResponseWriter, r *http.Request, status int, code, message string, details map[string]string) {
	resp := ErrorResponse{
		Error:     message,
		Code:      code,
		Details:   details,
		RequestID: requestIDFromContext(r.Context()),
		Timestamp: time.Now().UTC(),
	}
	writeJSON(w, status, resp)
}

func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(contextKeyRequestID).(string); ok {
		return v
	}
	// Fallback to common header keys if set in context via http.Request
	return ""
}

func hasScope(scopes []string, required string) bool {
	if len(scopes) == 0 {
		return false
	}
	for _, scope := range scopes {
		if strings.EqualFold(scope, required) {
			return true
		}
	}
	return false
}

func encodeCursor(idx int) string {
	return strconv.Itoa(idx)
}

func decodeCursor(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	idx, err := strconv.Atoi(raw)
	if err != nil || idx < 0 {
		return 0, err
	}
	return idx, nil
}
