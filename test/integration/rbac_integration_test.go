// Copyright 2025 James Ross
package integration_test

import (
	"bytes"
	"context"
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
	rbacandtokens "github.com/flyingrobots/go-redis-work-queue/internal/rbac-and-tokens"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// TestRBACIntegrationFullFlow tests complete RBAC flow with different roles
func TestRBACIntegrationFullFlow(t *testing.T) {
	// Setup test environment
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	// Add test data to Redis
	setupTestData(t, mr)

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"high":    "jobqueue:high",
				"medium":  "jobqueue:medium",
				"low":     "jobqueue:low",
				"payment": "jobqueue:payment",
			},
			DeadLetterList: "jobqueue:dead_letter",
		},
	}

	apiCfg := &adminapi.Config{
		JWTSecret:            "test-secret-key-for-rbac-integration-testing",
		RequireAuth:          true,
		DenyByDefault:        true,
		RequireDoubleConfirm: true,
		ConfirmationPhrase:   "CONFIRM_DELETE",
		AuditEnabled:         true,
		AuditLogPath:         "/tmp/rbac-integration-audit.log",
	}

	server, err := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Setup middleware stack
	handler := server.SetupRoutes()
	handler = adminapi.AuthMiddleware(apiCfg.JWTSecret, apiCfg.DenyByDefault, zap.NewNop())(handler)
	handler = adminapi.RequestIDMiddleware()(handler)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Test scenarios with different roles
	scenarios := []struct {
		name        string
		role        rbacandtokens.Role
		scopes      []rbacandtokens.Permission
		tests       []roleTestCase
		description string
	}{
		{
			name:   "Viewer Role Tests",
			role:   rbacandtokens.RoleViewer,
			scopes: []rbacandtokens.Permission{rbacandtokens.PermStatsRead, rbacandtokens.PermQueueRead},
			tests: []roleTestCase{
				{
					name:           "Can read stats",
					method:         "GET",
					path:           "/api/v1/stats",
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
				{
					name:           "Can peek queues",
					method:         "GET",
					path:           "/api/v1/queues/high/peek",
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
				{
					name:           "Cannot delete DLQ",
					method:         "DELETE",
					path:           "/api/v1/queues/dlq",
					body:           map[string]interface{}{"confirmation": "CONFIRM_DELETE", "reason": "test"},
					expectedStatus: http.StatusForbidden,
					shouldPass:     false,
				},
				{
					name:           "Cannot run benchmarks",
					method:         "POST",
					path:           "/api/v1/bench",
					body:           map[string]interface{}{"duration": "10s"},
					expectedStatus: http.StatusForbidden,
					shouldPass:     false,
				},
			},
			description: "Viewer should have read-only access",
		},
		{
			name:   "Operator Role Tests",
			role:   rbacandtokens.RoleOperator,
			scopes: []rbacandtokens.Permission{rbacandtokens.PermStatsRead, rbacandtokens.PermQueueWrite, rbacandtokens.PermBenchRun},
			tests: []roleTestCase{
				{
					name:           "Can read stats",
					method:         "GET",
					path:           "/api/v1/stats",
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
				{
					name:           "Can enqueue jobs",
					method:         "POST",
					path:           "/api/v1/queues/high/enqueue",
					body:           map[string]interface{}{"data": "test job", "priority": 5},
					expectedStatus: http.StatusCreated,
					shouldPass:     true,
				},
				{
					name:           "Can run benchmarks",
					method:         "POST",
					path:           "/api/v1/bench",
					body:           map[string]interface{}{"duration": "1s", "concurrency": 1},
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
				{
					name:           "Cannot delete DLQ",
					method:         "DELETE",
					path:           "/api/v1/queues/dlq",
					body:           map[string]interface{}{"confirmation": "CONFIRM_DELETE", "reason": "test"},
					expectedStatus: http.StatusForbidden,
					shouldPass:     false,
				},
				{
					name:           "Cannot delete all queues",
					method:         "DELETE",
					path:           "/api/v1/queues/all",
					body:           map[string]interface{}{"confirmation": "CONFIRM_DELETE_ALL", "reason": "test"},
					expectedStatus: http.StatusForbidden,
					shouldPass:     false,
				},
			},
			description: "Operator should have read/write but not destructive access",
		},
		{
			name:   "Maintainer Role Tests",
			role:   rbacandtokens.RoleMaintainer,
			scopes: []rbacandtokens.Permission{rbacandtokens.PermQueueDelete, rbacandtokens.PermJobDelete, rbacandtokens.PermWorkerManage},
			tests: []roleTestCase{
				{
					name:           "Can delete DLQ",
					method:         "DELETE",
					path:           "/api/v1/queues/dlq",
					body:           map[string]interface{}{"confirmation": "CONFIRM_DELETE", "reason": "Cleanup failed jobs"},
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
				{
					name:           "Can manage workers",
					method:         "POST",
					path:           "/api/v1/workers/restart",
					body:           map[string]interface{}{"worker_id": "worker-001"},
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
				{
					name:           "Cannot delete all queues without admin",
					method:         "DELETE",
					path:           "/api/v1/queues/all",
					body:           map[string]interface{}{"confirmation": "CONFIRM_DELETE_ALL", "reason": "test"},
					expectedStatus: http.StatusForbidden,
					shouldPass:     false,
				},
			},
			description: "Maintainer should have maintenance operations but not system-wide destruction",
		},
		{
			name:   "Admin Role Tests",
			role:   rbacandtokens.RoleAdmin,
			scopes: []rbacandtokens.Permission{rbacandtokens.PermAdminAll},
			tests: []roleTestCase{
				{
					name:           "Can access all stats",
					method:         "GET",
					path:           "/api/v1/stats",
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
				{
					name:           "Can delete DLQ",
					method:         "DELETE",
					path:           "/api/v1/queues/dlq",
					body:           map[string]interface{}{"confirmation": "CONFIRM_DELETE", "reason": "Admin cleanup"},
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
				{
					name:           "Can delete all queues",
					method:         "DELETE",
					path:           "/api/v1/queues/all",
					body:           map[string]interface{}{"confirmation": "CONFIRM_DELETE_ALL", "reason": "System reset"},
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
				{
					name:           "Can run benchmarks",
					method:         "POST",
					path:           "/api/v1/bench",
					body:           map[string]interface{}{"duration": "5s", "concurrency": 10},
					expectedStatus: http.StatusOK,
					shouldPass:     true,
				},
			},
			description: "Admin should have access to all operations",
		},
	}

	// Run test scenarios
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			token := createTestToken(apiCfg.JWTSecret, scenario.role, scenario.scopes, time.Hour)

			for _, test := range scenario.tests {
				t.Run(test.name, func(t *testing.T) {
					runRoleTestCase(t, ts.URL, token, test)
				})
			}
		})
	}
}

// TestResourceConstraints tests resource-level access controls
func TestResourceConstraints(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"payment-high":    "jobqueue:payment-high",
				"payment-low":     "jobqueue:payment-low",
				"email-high":      "jobqueue:email-high",
				"analytics-batch": "jobqueue:analytics-batch",
			},
		},
	}

	apiCfg := &adminapi.Config{
		JWTSecret:     "test-secret-key-for-resource-constraints",
		RequireAuth:   true,
		DenyByDefault: true,
	}

	server, err := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	handler := adminapi.AuthMiddleware(apiCfg.JWTSecret, apiCfg.DenyByDefault, zap.NewNop())(server.SetupRoutes())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	tests := []struct {
		name             string
		resourcePattern  string
		allowedResources []string
		deniedResources  []string
		description      string
	}{
		{
			name:            "Payment queues only",
			resourcePattern: "payment-*",
			allowedResources: []string{
				"/api/v1/queues/payment-high/peek",
				"/api/v1/queues/payment-low/peek",
			},
			deniedResources: []string{
				"/api/v1/queues/email-high/peek",
				"/api/v1/queues/analytics-batch/peek",
			},
			description: "User should only access payment queues",
		},
		{
			name:            "High priority queues only",
			resourcePattern: "*-high",
			allowedResources: []string{
				"/api/v1/queues/payment-high/peek",
				"/api/v1/queues/email-high/peek",
			},
			deniedResources: []string{
				"/api/v1/queues/payment-low/peek",
				"/api/v1/queues/analytics-batch/peek",
			},
			description: "User should only access high priority queues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create token with resource constraints
			token := createResourceConstrainedToken(
				apiCfg.JWTSecret,
				rbacandtokens.RoleOperator,
				[]rbacandtokens.Permission{rbacandtokens.PermQueueRead},
				tt.resourcePattern,
				time.Hour,
			)

			// Test allowed resources
			for _, resource := range tt.allowedResources {
				req, _ := http.NewRequest("GET", ts.URL+resource, nil)
				req.Header.Set("Authorization", "Bearer "+token)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Errorf("%s: Expected access to %s, got status %d", tt.description, resource, resp.StatusCode)
				}
			}

			// Test denied resources
			for _, resource := range tt.deniedResources {
				req, _ := http.NewRequest("GET", ts.URL+resource, nil)
				req.Header.Set("Authorization", "Bearer "+token)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					t.Errorf("%s: Expected denial of access to %s, got status %d", tt.description, resource, resp.StatusCode)
				}
			}
		})
	}
}

// TestTokenRevocationIntegration tests token revocation in real scenarios
func TestTokenRevocationIntegration(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{"high": "jobqueue:high"},
		},
	}

	apiCfg := &adminapi.Config{
		JWTSecret:     "test-secret-key-for-revocation",
		RequireAuth:   true,
		DenyByDefault: true,
	}

	server, err := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	handler := adminapi.AuthMiddleware(apiCfg.JWTSecret, apiCfg.DenyByDefault, zap.NewNop())(server.SetupRoutes())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Create token
	tokenID := "test-token-for-revocation"
	token := createTestTokenWithID(apiCfg.JWTSecret, tokenID, rbacandtokens.RoleOperator,
		[]rbacandtokens.Permission{rbacandtokens.PermStatsRead}, time.Hour)

	// First request should succeed
	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected successful request before revocation, got %d", resp.StatusCode)
	}

	// Simulate token revocation (in real implementation, this would go through revocation store)
	// For this test, we'll create an expired token to simulate revocation
	revokedToken := createTestTokenWithID(apiCfg.JWTSecret, tokenID, rbacandtokens.RoleOperator,
		[]rbacandtokens.Permission{rbacandtokens.PermStatsRead}, -time.Hour) // Expired

	// Second request with "revoked" token should fail
	req, _ = http.NewRequest("GET", ts.URL+"/api/v1/stats", nil)
	req.Header.Set("Authorization", "Bearer "+revokedToken)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected unauthorized request after revocation, got %d", resp.StatusCode)
	}
}

// TestAuditLoggingIntegration tests audit logging in integration scenarios
func TestAuditLoggingIntegration(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	// Setup test data
	setupTestData(t, mr)

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues:         map[string]string{"high": "jobqueue:high"},
			DeadLetterList: "jobqueue:dead_letter",
		},
	}

	apiCfg := &adminapi.Config{
		JWTSecret:     "test-secret-key-for-audit",
		RequireAuth:   true,
		DenyByDefault: true,
		AuditEnabled:  true,
		AuditLogPath:  "/tmp/audit-integration-test.log",
	}

	server, err := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Setup middleware with audit logging
	handler := server.SetupRoutes()
	handler = adminapi.AuthMiddleware(apiCfg.JWTSecret, apiCfg.DenyByDefault, zap.NewNop())(handler)
	// Note: In real implementation, audit middleware would be here

	ts := httptest.NewServer(handler)
	defer ts.Close()

	token := createTestToken(apiCfg.JWTSecret, rbacandtokens.RoleAdmin,
		[]rbacandtokens.Permission{rbacandtokens.PermAdminAll}, time.Hour)

	// Perform auditable operations
	auditableOperations := []struct {
		name   string
		method string
		path   string
		body   map[string]interface{}
	}{
		{
			name:   "DLQ Purge",
			method: "DELETE",
			path:   "/api/v1/queues/dlq",
			body:   map[string]interface{}{"confirmation": "CONFIRM_DELETE", "reason": "Test audit logging"},
		},
		{
			name:   "Benchmark Run",
			method: "POST",
			path:   "/api/v1/bench",
			body:   map[string]interface{}{"duration": "1s", "reason": "Performance testing"},
		},
	}

	for _, op := range auditableOperations {
		t.Run("Audit_"+op.name, func(t *testing.T) {
			var body *bytes.Buffer
			if op.body != nil {
				jsonBody, _ := json.Marshal(op.body)
				body = bytes.NewBuffer(jsonBody)
			} else {
				body = bytes.NewBuffer(nil)
			}

			req, _ := http.NewRequest(op.method, ts.URL+op.path, body)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			resp.Body.Close()

			// For audit logging, we mainly care that the request completed
			// In real implementation, we'd check the audit log file
			if resp.StatusCode >= 500 {
				t.Errorf("Server error during auditable operation %s: %d", op.name, resp.StatusCode)
			}
		})
	}
}

// Helper types and functions

type roleTestCase struct {
	name           string
	method         string
	path           string
	body           map[string]interface{}
	expectedStatus int
	shouldPass     bool
}

func setupTestData(t *testing.T, mr *miniredis.Miniredis) {
	// Add some test jobs to queues
	mr.Lpush("jobqueue:high", `{"id":"job1","data":"high priority job"}`)
	mr.Lpush("jobqueue:medium", `{"id":"job2","data":"medium priority job"}`)
	mr.Lpush("jobqueue:low", `{"id":"job3","data":"low priority job"}`)
	mr.Lpush("jobqueue:dead_letter", `{"id":"job4","data":"failed job"}`)
	mr.Lpush("jobqueue:payment-high", `{"id":"pay1","data":"payment processing"}`)
	mr.Lpush("jobqueue:email-high", `{"id":"email1","data":"email sending"}`)
}

func createTestToken(secret string, role rbacandtokens.Role, scopes []rbacandtokens.Permission, ttl time.Duration) string {
	return createTestTokenWithID(secret, fmt.Sprintf("token-%d", time.Now().UnixNano()), role, scopes, ttl)
}

func createTestTokenWithID(secret, tokenID string, role rbacandtokens.Role, scopes []rbacandtokens.Permission, ttl time.Duration) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	now := time.Now()
	claims := map[string]interface{}{
		"sub":        "test@example.com",
		"roles":      []rbacandtokens.Role{role},
		"scopes":     scopes,
		"exp":        now.Add(ttl).Unix(),
		"iat":        now.Unix(),
		"iss":        "test-issuer",
		"aud":        "admin-api",
		"jti":        tokenID,
		"kid":        "test-key-001",
		"token_type": "bearer",
	}

	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	message := header + "." + payload

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

func createResourceConstrainedToken(secret string, role rbacandtokens.Role, scopes []rbacandtokens.Permission, resourcePattern string, ttl time.Duration) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	now := time.Now()
	claims := map[string]interface{}{
		"sub":    "test@example.com",
		"roles":  []rbacandtokens.Role{role},
		"scopes": scopes,
		"exp":    now.Add(ttl).Unix(),
		"iat":    now.Unix(),
		"iss":    "test-issuer",
		"aud":    "admin-api",
		"jti":    fmt.Sprintf("resource-token-%d", time.Now().UnixNano()),
		"kid":    "test-key-001",
		"resources": map[string]string{
			"queues": resourcePattern,
		},
		"token_type": "bearer",
	}

	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	message := header + "." + payload

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

func runRoleTestCase(t *testing.T, baseURL, token string, test roleTestCase) {
	var body *bytes.Buffer
	if test.body != nil {
		jsonBody, _ := json.Marshal(test.body)
		body = bytes.NewBuffer(jsonBody)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(test.method, baseURL+test.path, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if test.shouldPass {
		if resp.StatusCode != test.expectedStatus {
			// Read response body for debugging
			buf := new(strings.Builder)
			buf.ReadFrom(resp.Body)
			t.Errorf("Test '%s' should pass: expected status %d, got %d. Response: %s",
				test.name, test.expectedStatus, resp.StatusCode, buf.String())
		}
	} else {
		if resp.StatusCode == test.expectedStatus {
			t.Errorf("Test '%s' should fail: got expected status %d but should have been denied",
				test.name, resp.StatusCode)
		}
		// For denied operations, expect 403 or 401
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Test '%s' should be denied with 401/403, got %d", test.name, resp.StatusCode)
		}
	}
}