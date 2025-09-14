// Copyright 2025 James Ross
package rbacandtokens

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestHandler_GenerateToken(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	handler := NewHandler(manager, zap.NewNop())

	tests := []struct {
		name           string
		request        TokenRequest
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "Valid token request",
			request: TokenRequest{
				Subject: "test@example.com",
				Roles:   []Role{RoleViewer},
				TTL:     "1h",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "Missing subject",
			request: TokenRequest{
				Roles: []Role{RoleViewer},
				TTL:   "1h",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "Missing roles",
			request: TokenRequest{
				Subject: "test@example.com",
				TTL:     "1h",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "Invalid TTL",
			request: TokenRequest{
				Subject: "test@example.com",
				Roles:   []Role{RoleViewer},
				TTL:     "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/auth/token", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.GenerateToken(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectToken {
				var response TokenResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.Token == "" {
					t.Error("Expected token in response")
				}

				if response.Subject != tt.request.Subject {
					t.Errorf("Expected subject %s, got %s", tt.request.Subject, response.Subject)
				}
			}
		})
	}
}

func TestHandler_ValidateToken(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	handler := NewHandler(manager, zap.NewNop())

	// Generate a valid token
	token, err := manager.GenerateToken("test@example.com", []Role{RoleViewer}, nil, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedValid  bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name:           "Missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedValid:  false,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/auth/validate", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ValidateToken(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedValid {
				var response TokenInfoResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if !response.Valid {
					t.Error("Expected valid token response")
				}

				if response.Subject != "test@example.com" {
					t.Errorf("Expected subject test@example.com, got %s", response.Subject)
				}
			}
		})
	}
}

func TestHandler_CheckAccess(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	handler := NewHandler(manager, zap.NewNop())

	// Generate tokens with different roles
	viewerToken, _ := manager.GenerateToken("viewer@example.com", []Role{RoleViewer}, nil, 1*time.Hour)
	adminToken, _ := manager.GenerateToken("admin@example.com", []Role{RoleAdmin}, nil, 1*time.Hour)

	tests := []struct {
		name           string
		authHeader     string
		action         string
		resource       string
		expectedStatus int
		expectAllowed  bool
	}{
		{
			name:           "Admin can delete queues",
			authHeader:     "Bearer " + adminToken,
			action:         "queue:delete",
			resource:       "/api/v1/queues/dlq",
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
		{
			name:           "Viewer can read stats",
			authHeader:     "Bearer " + viewerToken,
			action:         "stats:read",
			resource:       "/api/v1/stats",
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
		{
			name:           "Viewer cannot delete queues",
			authHeader:     "Bearer " + viewerToken,
			action:         "queue:delete",
			resource:       "/api/v1/queues/dlq",
			expectedStatus: http.StatusOK,
			expectAllowed:  false,
		},
		{
			name:           "No token provided",
			authHeader:     "",
			action:         "stats:read",
			resource:       "/api/v1/stats",
			expectedStatus: http.StatusOK,
			expectAllowed:  false,
		},
		{
			name:           "Missing action parameter",
			authHeader:     "Bearer " + viewerToken,
			action:         "",
			resource:       "/api/v1/stats",
			expectedStatus: http.StatusBadRequest,
			expectAllowed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/auth/check"
			if tt.action != "" {
				url += "?action=" + tt.action
			}
			if tt.resource != "" {
				if tt.action != "" {
					url += "&resource=" + tt.resource
				} else {
					url += "?resource=" + tt.resource
				}
			}

			req := httptest.NewRequest("POST", url, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.CheckAccess(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response AuthorizationResult
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.Allowed != tt.expectAllowed {
					t.Errorf("Expected allowed=%v, got allowed=%v (reason: %s)",
						tt.expectAllowed, response.Allowed, response.Reason)
				}
			}
		})
	}
}

func TestHandler_RevokeToken(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	handler := NewHandler(manager, zap.NewNop())

	// Generate a token to revoke
	token, err := manager.GenerateToken("test@example.com", []Role{RoleViewer}, nil, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Get token ID
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	tests := []struct {
		name           string
		request        RevokeTokenRequest
		expectedStatus int
	}{
		{
			name: "Valid revocation",
			request: RevokeTokenRequest{
				TokenID: claims.JWTID,
				Reason:  "Test revocation",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing token ID",
			request: RevokeTokenRequest{
				Reason: "Test revocation",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Revocation with default reason",
			request: RevokeTokenRequest{
				TokenID: "some-token-id",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/auth/token/revoke", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RevokeToken(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if success, ok := response["success"].(bool); !ok || !success {
					t.Error("Expected success response")
				}
			}
		})
	}

	// Verify original token is now revoked
	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating revoked token")
	}
}

func TestHandler_GetTokenInfo(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	handler := NewHandler(manager, zap.NewNop())

	// Generate a valid token
	token, err := manager.GenerateToken("test@example.com", []Role{RoleViewer}, nil, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedValid  bool
	}{
		{
			name:           "Valid token info request",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name:           "Missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedValid:  false,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusOK,
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/auth/token/info", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.GetTokenInfo(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response TokenInfoResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.Valid != tt.expectedValid {
					t.Errorf("Expected valid=%v, got valid=%v", tt.expectedValid, response.Valid)
				}

				if tt.expectedValid && response.Subject != "test@example.com" {
					t.Errorf("Expected subject test@example.com, got %s", response.Subject)
				}
			}
		})
	}
}