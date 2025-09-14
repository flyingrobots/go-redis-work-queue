// Copyright 2025 James Ross
// +build security

package adminapi_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	adminapi "github.com/flyingrobots/go-redis-work-queue/internal/admin-api"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// TestSecurityAuthRequired verifies that auth is enforced when enabled
func TestSecurityAuthRequired(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"high": "jobqueue:high",
				"low":  "jobqueue:low",
			},
		},
	}

	// Enable auth with deny-by-default
	apiCfg := &adminapi.Config{
		JWTSecret:     "test-secret",
		RequireAuth:   true,
		DenyByDefault: true,
	}

	server, _ := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	handler := server.SetupRoutes()

	// Apply auth middleware
	handler = adminapi.AuthMiddleware(apiCfg.JWTSecret, apiCfg.DenyByDefault, zap.NewNop())(handler)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "No auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Wrong scheme",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid JWT",
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Expired JWT",
			authHeader:     "Bearer " + createExpiredJWT("test-secret"),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid JWT",
			authHeader:     "Bearer " + createValidJWT("test-secret"),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL+"/api/v1/stats", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestSecurityDestructiveOperations verifies extra security for dangerous operations
func TestSecurityDestructiveOperations(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"high": "jobqueue:high",
				"low":  "jobqueue:low",
			},
			DeadLetterList: "jobqueue:dead_letter",
		},
	}

	apiCfg := &adminapi.Config{
		RequireDoubleConfirm: true,
		ConfirmationPhrase:   "CONFIRM_DELETE",
		AuditEnabled:         true,
		AuditLogPath:         "/tmp/test-audit-security.log",
	}

	server, _ := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	ts := httptest.NewServer(server.SetupRoutes())
	defer ts.Close()

	// Add test data
	mr.Lpush("jobqueue:dead_letter", "job1")
	mr.Lpush("jobqueue:dead_letter", "job2")

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		{
			name:   "PurgeDLQ without confirmation",
			method: "DELETE",
			path:   "/api/v1/queues/dlq",
			body: adminapi.PurgeRequest{
				Reason: "Test",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject without confirmation",
		},
		{
			name:   "PurgeDLQ with wrong confirmation",
			method: "DELETE",
			path:   "/api/v1/queues/dlq",
			body: adminapi.PurgeRequest{
				Confirmation: "YES",
				Reason:       "Test",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject with wrong confirmation",
		},
		{
			name:   "PurgeDLQ without reason",
			method: "DELETE",
			path:   "/api/v1/queues/dlq",
			body: adminapi.PurgeRequest{
				Confirmation: "CONFIRM_DELETE",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should require reason",
		},
		{
			name:   "PurgeAll with single confirmation",
			method: "DELETE",
			path:   "/api/v1/queues/all",
			body: adminapi.PurgeRequest{
				Confirmation: "CONFIRM_DELETE",
				Reason:       "Test purge all",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should require double confirmation for purge all",
		},
		{
			name:   "PurgeAll with correct double confirmation",
			method: "DELETE",
			path:   "/api/v1/queues/all",
			body: adminapi.PurgeRequest{
				Confirmation: "CONFIRM_DELETE_ALL",
				Reason:       "Valid reason for purging everything",
			},
			expectedStatus: http.StatusOK,
			description:    "Should accept with correct double confirmation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(tt.method, ts.URL+tt.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				var errResp adminapi.ErrorResponse
				json.NewDecoder(resp.Body).Decode(&errResp)
				t.Errorf("%s: Expected status %d, got %d: %s",
					tt.description, tt.expectedStatus, resp.StatusCode, errResp.Error)
			}
		})
	}
}

// TestSecurityTokenLeakage verifies tokens aren't leaked in responses
func TestSecurityTokenLeakage(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{"high": "jobqueue:high"},
		},
	}

	apiCfg := &adminapi.Config{
		JWTSecret:     "secret-key",
		RequireAuth:   true,
		DenyByDefault: true,
	}

	server, _ := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	handler := adminapi.AuthMiddleware(apiCfg.JWTSecret, apiCfg.DenyByDefault, zap.NewNop())(server.SetupRoutes())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	token := createValidJWT("secret-key")

	// Make various requests with auth
	endpoints := []string{
		"/api/v1/stats",
		"/api/v1/stats/keys",
		"/api/v1/queues/high/peek",
	}

	for _, endpoint := range endpoints {
		req, _ := http.NewRequest("GET", ts.URL+endpoint, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request to %s failed: %v", endpoint, err)
		}
		defer resp.Body.Close()

		// Read response body
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		body := buf.String()

		// Check for token leakage
		if strings.Contains(body, token) {
			t.Errorf("Token leaked in response from %s", endpoint)
		}

		if strings.Contains(body, "secret-key") {
			t.Errorf("Secret key leaked in response from %s", endpoint)
		}

		// Check response headers for leakage
		for key, values := range resp.Header {
			for _, value := range values {
				if strings.Contains(value, token) {
					t.Errorf("Token leaked in header %s from %s", key, endpoint)
				}
				if strings.Contains(value, "secret-key") {
					t.Errorf("Secret key leaked in header %s from %s", key, endpoint)
				}
			}
		}
	}
}

// TestSecurityInjectionAttacks tests for various injection vulnerabilities
func TestSecurityInjectionAttacks(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"high": "jobqueue:high",
				"low":  "jobqueue:low",
			},
		},
	}

	apiCfg := &adminapi.Config{}

	server, _ := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	ts := httptest.NewServer(server.SetupRoutes())
	defer ts.Close()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Path traversal attempt",
			path:           "/api/v1/queues/../../etc/passwd/peek",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject path traversal",
		},
		{
			name:           "Redis command injection",
			path:           "/api/v1/queues/high%3B%20FLUSHALL/peek",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject Redis command injection",
		},
		{
			name:           "SQL injection attempt",
			path:           "/api/v1/queues/high'%20OR%20'1'%3D'1/peek",
			expectedStatus: http.StatusBadRequest,
			description:    "Should handle SQL injection safely",
		},
		{
			name:           "Script tag in queue name",
			path:           "/api/v1/queues/%3Cscript%3Ealert(1)%3C%2Fscript%3E/peek",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject script injection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tt.path)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("%s: Expected status %d, got %d",
					tt.description, tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestSecurityCORS verifies CORS headers are properly set
func TestSecurityCORS(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{"high": "jobqueue:high"},
		},
	}

	apiCfg := &adminapi.Config{
		CORSEnabled:      true,
		CORSAllowOrigins: []string{"https://example.com", "https://app.example.com"},
	}

	server, _ := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	handler := adminapi.CORSMiddleware(apiCfg.CORSAllowOrigins)(server.SetupRoutes())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	tests := []struct {
		name           string
		origin         string
		method         string
		expectCORS     bool
		expectedOrigin string
	}{
		{
			name:           "Allowed origin",
			origin:         "https://example.com",
			method:         "GET",
			expectCORS:     true,
			expectedOrigin: "https://example.com",
		},
		{
			name:           "Another allowed origin",
			origin:         "https://app.example.com",
			method:         "GET",
			expectCORS:     true,
			expectedOrigin: "https://app.example.com",
		},
		{
			name:       "Disallowed origin",
			origin:     "https://evil.com",
			method:     "GET",
			expectCORS: false,
		},
		{
			name:           "OPTIONS preflight",
			origin:         "https://example.com",
			method:         "OPTIONS",
			expectCORS:     true,
			expectedOrigin: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, ts.URL+"/api/v1/stats", nil)
			req.Header.Set("Origin", tt.origin)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			corsHeader := resp.Header.Get("Access-Control-Allow-Origin")

			if tt.expectCORS {
				if corsHeader != tt.expectedOrigin {
					t.Errorf("Expected CORS header %s, got %s", tt.expectedOrigin, corsHeader)
				}

				// Check other CORS headers
				if resp.Header.Get("Access-Control-Allow-Methods") == "" {
					t.Error("Missing Access-Control-Allow-Methods header")
				}
			} else {
				if corsHeader != "" {
					t.Errorf("Expected no CORS header for %s, got %s", tt.origin, corsHeader)
				}
			}

			// OPTIONS should return 204
			if tt.method == "OPTIONS" && resp.StatusCode != http.StatusNoContent {
				t.Errorf("Expected status 204 for OPTIONS, got %d", resp.StatusCode)
			}
		})
	}
}

// Helper functions

func createValidJWT(secret string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := fmt.Sprintf(`{"sub":"test@example.com","roles":["admin"],"exp":%d,"iat":%d}`,
		time.Now().Add(1*time.Hour).Unix(),
		time.Now().Unix())
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))

	message := header + "." + payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

func createExpiredJWT(secret string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := fmt.Sprintf(`{"sub":"test@example.com","roles":["admin"],"exp":%d,"iat":%d}`,
		time.Now().Add(-1*time.Hour).Unix(), // Expired
		time.Now().Add(-2*time.Hour).Unix())
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))

	message := header + "." + payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}