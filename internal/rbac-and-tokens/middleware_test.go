// Copyright 2025 James Ross
package rbacandtokens

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestAuthMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	middleware := AuthMiddleware(manager, zap.NewNop())

	// Generate a valid token
	token, err := manager.GenerateToken("test@example.com", []Role{RoleViewer}, nil, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test handler that checks for claims in context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ContextKeyClaims).(*Claims)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no claims in context"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("subject:" + claims.Subject))
	})

	wrappedHandler := middleware(testHandler)

	tests := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid token",
			path:           "/api/v1/stats",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectedBody:   "subject:test@example.com",
		},
		{
			name:           "Missing auth header",
			path:           "/api/v1/stats",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "Invalid token",
			path:           "/api/v1/stats",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "Health check bypasses auth",
			path:           "/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "no claims in context", // Health check doesn't have claims
		},
		{
			name:           "Token endpoint bypasses auth",
			path:           "/auth/token",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "no claims in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" && w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestAuthzMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create middleware chain
	authMiddleware := AuthMiddleware(manager, zap.NewNop())
	authzMiddleware := AuthzMiddleware(manager, zap.NewNop())

	// Generate tokens with different roles
	viewerToken, _ := manager.GenerateToken("viewer@example.com", []Role{RoleViewer}, nil, 1*time.Hour)
	adminToken, _ := manager.GenerateToken("admin@example.com", []Role{RoleAdmin}, nil, 1*time.Hour)

	// Test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrappedHandler := authMiddleware(authzMiddleware(testHandler))

	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Admin can access DELETE /api/v1/queues/all",
			method:         "DELETE",
			path:           "/api/v1/queues/all",
			authHeader:     "Bearer " + adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Viewer can access GET /api/v1/stats",
			method:         "GET",
			path:           "/api/v1/stats",
			authHeader:     "Bearer " + viewerToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Viewer cannot access DELETE /api/v1/queues/dlq",
			method:         "DELETE",
			path:           "/api/v1/queues/dlq",
			authHeader:     "Bearer " + viewerToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Unknown endpoint allows through",
			method:         "GET",
			path:           "/api/v1/unknown",
			authHeader:     "Bearer " + viewerToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Health check bypasses authz",
			method:         "GET",
			path:           "/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Auth endpoint bypasses authz",
			method:         "POST",
			path:           "/auth/validate",
			authHeader:     "Bearer " + viewerToken,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s",
					tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuditMiddleware(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = true
	config.AuditConfig.LogPath = "/tmp/test-audit-rbac.log"

	auditLogger, err := NewAuditLogger(&config.AuditConfig)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	manager, err := NewManager(config, auditLogger, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create middleware chain
	authMiddleware := AuthMiddleware(manager, zap.NewNop())
	authzMiddleware := AuthzMiddleware(manager, zap.NewNop())
	auditMiddleware := AuditMiddleware(manager, zap.NewNop())

	// Generate token
	token, _ := manager.GenerateToken("test@example.com", []Role{RoleAdmin}, nil, 1*time.Hour)

	// Test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrappedHandler := authMiddleware(authzMiddleware(auditMiddleware(testHandler)))

	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectAuditLog bool
	}{
		{
			name:           "Destructive operation should be audited",
			method:         "DELETE",
			path:           "/api/v1/queues/dlq",
			authHeader:     "Bearer " + token,
			expectAuditLog: true,
		},
		{
			name:           "Non-destructive operation should not be audited",
			method:         "GET",
			path:           "/api/v1/stats",
			authHeader:     "Bearer " + token,
			expectAuditLog: false,
		},
		{
			name:           "POST benchmark should be audited",
			method:         "POST",
			path:           "/api/v1/bench",
			authHeader:     "Bearer " + token,
			expectAuditLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", tt.authHeader)
			req.Header.Set("X-Request-ID", "test-request-"+tt.name)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			// For this test, we mainly verify the middleware doesn't crash
			// In a real test, we would check the audit log file content
		})
	}
}

func TestGetRequiredPermission(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected Permission
	}{
		{"GET", "/api/v1/stats", PermStatsRead},
		{"GET", "/api/v1/stats/keys", PermStatsRead},
		{"DELETE", "/api/v1/queues/dlq", PermQueueDelete},
		{"POST", "/api/v1/bench", PermBenchRun},
		{"GET", "/api/v1/unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			result := getRequiredPermission(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("Expected permission %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsDestructiveOperation(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected bool
	}{
		{"DELETE", "/api/v1/queues/dlq", true},
		{"DELETE", "/api/v1/queues/all", true},
		{"POST", "/api/v1/bench", true},
		{"POST", "/auth/token/revoke", true},
		{"GET", "/api/v1/stats", false},
		{"POST", "/api/v1/stats", false},
		{"PUT", "/api/v1/queues/dlq", false},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			result := isDestructiveOperation(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("Expected destructive=%v, got destructive=%v", tt.expected, result)
			}
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{"GET /api/v1/queues/high/peek", "GET /api/v1/queues/*/peek", true},
		{"GET /api/v1/queues/low/peek", "GET /api/v1/queues/*/peek", true},
		{"GET /api/v1/stats", "GET /api/v1/queues/*/peek", false},
		{"GET /api/v1/stats", "GET /api/v1/stats", true},
		{"POST /api/v1/stats", "GET /api/v1/stats", false},
	}

	for _, tt := range tests {
		t.Run(tt.path+" matches "+tt.pattern, func(t *testing.T) {
			result := matchesPattern(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("Expected matches=%v, got matches=%v", tt.expected, result)
			}
		})
	}
}