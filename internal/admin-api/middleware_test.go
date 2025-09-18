// Copyright 2025 James Ross
package adminapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	anomalyradarslobudget "github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget"
	"github.com/flyingrobots/go-redis-work-queue/internal/rbac-and-tokens"
	"go.uber.org/zap"
)

func TestRequestIDMiddleware(t *testing.T) {
	handler := RequestIDMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that request ID is in context
		reqID := r.Context().Value(contextKeyRequestID)
		if reqID == nil {
			t.Error("Request ID not found in context")
		}

		// Check that header is set
		if w.Header().Get("X-Request-ID") == "" {
			t.Error("X-Request-ID header not set")
		}

		w.WriteHeader(http.StatusOK)
	}))

	// Test with no existing request ID
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID not set in response")
	}

	// Test with existing request ID
	existingID := "test-request-id"
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != existingID {
		t.Errorf("Expected X-Request-ID %s, got %s", existingID, w.Header().Get("X-Request-ID"))
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	logger := zap.NewNop()

	handler := RecoveryMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler.ServeHTTP(w, req)

	// Should return 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestAuthMiddlewareWithoutDenyByDefault(t *testing.T) {
	logger := zap.NewNop()
	secret := "test-secret"

	handler := AuthMiddleware(secret, false, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Without deny-by-default, requests should pass through
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareInjectsScopesForDownstreamHandlers(t *testing.T) {
	logger := zap.NewNop()
	secret := "test-secret"
	token := mustMakeScopedToken(t, secret, []string{string(rbacandtokens.PermAdminAll)})

	handler := AuthMiddleware(secret, true, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopesVal := r.Context().Value(contextKeyScopes)
		if scopesVal == nil {
			t.Fatal("expected scopes to be present in context")
		}

		scopes, ok := scopesVal.([]string)
		if !ok {
			t.Fatalf("expected []string scopes, got %T", scopesVal)
		}

		if len(scopes) != 1 || scopes[0] != string(rbacandtokens.PermAdminAll) {
			t.Fatalf("unexpected scopes in context: %v", scopes)
		}

		radarScopes := anomalyradarslobudget.ScopesFromContext(r.Context())
		if len(radarScopes) != 1 || radarScopes[0] != string(rbacandtokens.PermAdminAll) {
			t.Fatalf("anomaly radar scopes missing or incorrect: %v", radarScopes)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}

func mustMakeScopedToken(t *testing.T, secret string, scopes []string) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := map[string]interface{}{
		"sub":    "test@example.com",
		"roles":  []string{"admin"},
		"scopes": scopes,
		"exp":    time.Now().Add(time.Hour).Unix(),
		"iat":    time.Now().Unix(),
		"iss":    "unit-test",
		"aud":    "admin-api",
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal claims: %v", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	message := header + "." + payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return message + "." + signature
}

func TestRateLimitBucketRefill(t *testing.T) {
	bucket := &rateBucket{
		tokens:    0,
		lastFill:  time.Now().Add(-2 * time.Second),
		maxTokens: 10,
		fillRate:  5.0, // 5 tokens per second
	}

	// After 2 seconds, should have refilled
	if !bucket.consume() {
		t.Error("Should have tokens after refill")
	}

	// Verify tokens were refilled correctly
	if bucket.tokens < 9 { // Should be close to 10
		t.Errorf("Expected ~10 tokens after refill, got %f", bucket.tokens)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "1.2.3.4"},
			remoteAddr: "5.6.7.8:1234",
			expectedIP: "1.2.3.4",
		},
		{
			name:       "X-Forwarded-For single",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4"},
			remoteAddr: "5.6.7.8:1234",
			expectedIP: "1.2.3.4",
		},
		{
			name:       "X-Forwarded-For multiple",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4, 5.6.7.8, 9.10.11.12"},
			remoteAddr: "127.0.0.1:1234",
			expectedIP: "1.2.3.4",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "5.6.7.8:1234",
			expectedIP: "5.6.7.8:1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestIsDestructiveOperation(t *testing.T) {
	tests := []struct {
		method      string
		path        string
		destructive bool
	}{
		{"GET", "/api/v1/stats", false},
		{"DELETE", "/api/v1/queues/dlq", true},
		{"DELETE", "/api/v1/queues/all", true},
		{"POST", "/api/v1/bench", true},
		{"POST", "/api/v1/stats", false},
		{"DELETE", "/api/v1/other", false},
	}

	for _, tt := range tests {
		result := isDestructiveOperation(tt.method, tt.path)
		if result != tt.destructive {
			t.Errorf("For %s %s, expected destructive=%v, got %v",
				tt.method, tt.path, tt.destructive, result)
		}
	}
}

func TestAuditLoggerRotation(t *testing.T) {
	// Create temp directory for test
	tmpDir := "/tmp/test-audit-" + generateID()
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	logPath := tmpDir + "/audit.log"

	// Create logger with small max size
	logger, err := NewAuditLogger(logPath, 100, 2) // 100 bytes max, 2 backups
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer logger.Close()

	// Write entries to trigger rotation
	for i := 0; i < 5; i++ {
		entry := AuditEntry{
			ID:        generateID(),
			Timestamp: time.Now(),
			User:      "test",
			Action:    "TEST",
			Result:    "SUCCESS",
		}
		if err := logger.Log(entry); err != nil {
			t.Errorf("Failed to log entry %d: %v", i, err)
		}
	}

	// Check that rotation occurred
	if _, err := os.Stat(logPath); err != nil {
		t.Error("Current log file should exist")
	}

	// Should have at least one backup
	files, _ := os.ReadDir(tmpDir)
	if len(files) < 2 {
		t.Error("Expected at least 2 files after rotation")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("X-Request-ID", "req-123")

	writeError(w, http.StatusBadRequest, "TEST_ERROR", "Test error message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	if got := w.Header().Get("X-Request-ID"); got != "req-123" {
		t.Errorf("Expected X-Request-ID header to be preserved, got %s", got)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Code != "TEST_ERROR" {
		t.Errorf("Expected code TEST_ERROR, got %s", resp.Code)
	}

	if resp.Status != http.StatusBadRequest {
		t.Errorf("Expected status 400 in body, got %d", resp.Status)
	}

	if resp.RequestID != "req-123" {
		t.Errorf("Expected request ID req-123, got %s", resp.RequestID)
	}

	if resp.Timestamp.IsZero() {
		t.Error("Expected timestamp to be populated")
	}
}

func TestWriteErrorGeneratesRequestIDWhenMissing(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusInternalServerError, "INTERNAL", "boom")

	if got := w.Header().Get("X-Request-ID"); got == "" {
		t.Error("Expected X-Request-ID header to be generated")
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.RequestID == "" {
		t.Error("Expected response request_id to be generated")
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"test": "value"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestResponseWriterWrapper(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	// Test WriteHeader
	rw.WriteHeader(http.StatusCreated)
	if rw.statusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rw.statusCode)
	}

	// Test that underlying writer is called
	if w.Code != http.StatusCreated {
		t.Errorf("Expected underlying writer status 201, got %d", w.Code)
	}
}

func TestCORSMiddlewareOptions(t *testing.T) {
	handler := CORSMiddleware([]string{"*"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("Missing Access-Control-Allow-Origin header")
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Missing Access-Control-Allow-Methods header")
	}
}

func TestAuditMiddlewareNonDestructive(t *testing.T) {
	logger := zap.NewNop()

	// Create temp audit log
	tmpFile := "/tmp/test-audit-" + generateID() + ".log"
	defer os.Remove(tmpFile)

	auditLog, _ := NewAuditLogger(tmpFile, 1024*1024, 5)
	defer auditLog.Close()

	callCount := 0
	handler := AuditMiddleware(auditLog, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))

	// Non-destructive operation should not be logged
	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if callCount != 1 {
		t.Error("Handler not called")
	}

	// Check audit log is empty
	stat, _ := os.Stat(tmpFile)
	if stat.Size() > 0 {
		t.Error("Non-destructive operation should not be logged")
	}
}

func TestMethodHandler(t *testing.T) {
	handler := methodHandler("GET", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test correct method
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test wrong method
	req = httptest.NewRequest("POST", "/test", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"/api/v1/queues/dlq", "/dlq", true},
		{"/api/v1/queues/all", "/all", true},
		{"/api/v1/queues/high/peek", "/peek", true},
		{"/api/v1/stats", "/dlq", false},
		{"", "/test", false},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %v, expected %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}
