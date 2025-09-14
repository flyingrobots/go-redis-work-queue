// Copyright 2025 James Ross
package e2e_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	adminapi "github.com/flyingrobots/go-redis-work-queue/internal/admin-api"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	rbacandtokens "github.com/flyingrobots/go-redis-work-queue/internal/rbac-and-tokens"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestE2ETokenLifecycle tests the complete token lifecycle
func TestE2ETokenLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup complete system
	system := setupE2ESystem(t)
	defer system.cleanup()

	// Test complete token workflow
	scenarios := []struct {
		name        string
		role        rbacandtokens.Role
		tokenTTL    time.Duration
		operations  []e2eOperation
		description string
	}{
		{
			name:     "DevOps Engineer Workflow",
			role:     rbacandtokens.RoleOperator,
			tokenTTL: 8 * time.Hour,
			operations: []e2eOperation{
				{
					name:           "Check system status",
					method:         "GET",
					path:           "/api/v1/stats",
					expectSuccess:  true,
					validateResult: validateStatsResponse,
				},
				{
					name:           "Enqueue deployment job",
					method:         "POST",
					path:           "/api/v1/queues/deployment/enqueue",
					body:           map[string]interface{}{"task": "deploy", "version": "v1.2.3"},
					expectSuccess:  true,
					validateResult: validateJobEnqueue,
				},
				{
					name:           "Monitor queue status",
					method:         "GET",
					path:           "/api/v1/queues/deployment/peek",
					expectSuccess:  true,
					validateResult: validateQueuePeek,
				},
				{
					name:          "Attempt DLQ purge (should fail)",
					method:        "DELETE",
					path:          "/api/v1/queues/dlq",
					body:          map[string]interface{}{"confirmation": "CONFIRM_DELETE", "reason": "cleanup"},
					expectSuccess: false,
				},
			},
			description: "DevOps engineer should be able to deploy but not purge",
		},
		{
			name:     "Site Reliability Engineer Workflow",
			role:     rbacandtokens.RoleMaintainer,
			tokenTTL: 12 * time.Hour,
			operations: []e2eOperation{
				{
					name:           "Check system health",
					method:         "GET",
					path:           "/api/v1/stats",
					expectSuccess:  true,
					validateResult: validateStatsResponse,
				},
				{
					name:           "Purge dead letter queue",
					method:         "DELETE",
					path:           "/api/v1/queues/dlq",
					body:           map[string]interface{}{"confirmation": "CONFIRM_DELETE", "reason": "SRE cleanup"},
					expectSuccess:  true,
					validateResult: validateDLQPurge,
				},
				{
					name:           "Restart worker",
					method:         "POST",
					path:           "/api/v1/workers/restart",
					body:           map[string]interface{}{"worker_id": "worker-001"},
					expectSuccess:  true,
					validateResult: validateWorkerRestart,
				},
				{
					name:          "Attempt system purge (should fail)",
					method:        "DELETE",
					path:          "/api/v1/queues/all",
					body:          map[string]interface{}{"confirmation": "CONFIRM_DELETE_ALL", "reason": "test"},
					expectSuccess: false,
				},
			},
			description: "SRE should have maintenance access but not system-wide destruction",
		},
		{
			name:     "Security Admin Workflow",
			role:     rbacandtokens.RoleAdmin,
			tokenTTL: 24 * time.Hour,
			operations: []e2eOperation{
				{
					name:           "Full system status",
					method:         "GET",
					path:           "/api/v1/stats",
					expectSuccess:  true,
					validateResult: validateStatsResponse,
				},
				{
					name:           "Emergency queue purge",
					method:         "DELETE",
					path:           "/api/v1/queues/all",
					body:           map[string]interface{}{"confirmation": "CONFIRM_DELETE_ALL", "reason": "Security incident response"},
					expectSuccess:  true,
					validateResult: validateQueuePurgeAll,
				},
				{
					name:           "Performance benchmark",
					method:         "POST",
					path:           "/api/v1/bench",
					body:           map[string]interface{}{"duration": "2s", "concurrency": 5, "reason": "Security testing"},
					expectSuccess:  true,
					validateResult: validateBenchmarkRun,
				},
			},
			description: "Admin should have full system access",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Generate token for this scenario
			token := system.createToken(scenario.role, nil, scenario.tokenTTL)

			// Execute operations in sequence
			for _, op := range scenario.operations {
				t.Run(op.name, func(t *testing.T) {
					result := system.executeOperation(op, token)

					if op.expectSuccess && !result.success {
						t.Errorf("%s: Expected success but got error: %s", scenario.description, result.error)
					}

					if !op.expectSuccess && result.success {
						t.Errorf("%s: Expected failure but operation succeeded", scenario.description)
					}

					// Validate result if provided
					if result.success && op.validateResult != nil {
						if err := op.validateResult(result.response); err != nil {
							t.Errorf("%s: Validation failed: %v", scenario.description, err)
						}
					}
				})
			}
		})
	}
}

// TestE2ESecurityBoundaries tests security boundaries in realistic scenarios
func TestE2ESecurityBoundaries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E security test in short mode")
	}

	system := setupE2ESystem(t)
	defer system.cleanup()

	securityTests := []struct {
		name        string
		attack      func(*e2eSystem) attackResult
		description string
	}{
		{
			name: "Token Forgery Attack",
			attack: func(sys *e2eSystem) attackResult {
				// Attempt to forge token with wrong signature
				fakeToken := createMaliciousToken("wrong-secret", rbacandtokens.RoleAdmin)
				result := sys.executeOperation(e2eOperation{
					name:   "Forged admin access",
					method: "GET",
					path:   "/api/v1/stats",
				}, fakeToken)

				return attackResult{
					blocked:     !result.success,
					description: "Forged token should be rejected",
				}
			},
			description: "System should detect and block forged tokens",
		},
		{
			name: "Privilege Escalation Attack",
			attack: func(sys *e2eSystem) attackResult {
				// Create viewer token, then try admin operations
				viewerToken := sys.createToken(rbacandtokens.RoleViewer, nil, time.Hour)
				result := sys.executeOperation(e2eOperation{
					name:   "Privilege escalation",
					method: "DELETE",
					path:   "/api/v1/queues/all",
					body:   map[string]interface{}{"confirmation": "CONFIRM_DELETE_ALL", "reason": "attack"},
				}, viewerToken)

				return attackResult{
					blocked:     !result.success,
					description: "Viewer should not be able to perform admin operations",
				}
			},
			description: "System should prevent privilege escalation",
		},
		{
			name: "Replay Attack",
			attack: func(sys *e2eSystem) attackResult {
				// Use expired token
				expiredToken := sys.createToken(rbacandtokens.RoleAdmin, nil, -time.Hour)
				result := sys.executeOperation(e2eOperation{
					name:   "Replay with expired token",
					method: "GET",
					path:   "/api/v1/stats",
				}, expiredToken)

				return attackResult{
					blocked:     !result.success,
					description: "Expired token should be rejected",
				}
			},
			description: "System should block replay attacks with expired tokens",
		},
		{
			name: "Resource Boundary Violation",
			attack: func(sys *e2eSystem) attackResult {
				// Create token with payment-* resource constraint
				constrainedToken := sys.createResourceConstrainedToken(
					rbacandtokens.RoleOperator,
					[]rbacandtokens.Permission{rbacandtokens.PermQueueRead},
					"payment-*",
					time.Hour,
				)

				// Try to access email queue
				result := sys.executeOperation(e2eOperation{
					name:   "Resource boundary violation",
					method: "GET",
					path:   "/api/v1/queues/email/peek",
				}, constrainedToken)

				return attackResult{
					blocked:     !result.success,
					description: "Should not access resources outside constraints",
				}
			},
			description: "System should enforce resource boundaries",
		},
	}

	for _, test := range securityTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.attack(system)

			if !result.blocked {
				t.Errorf("%s: %s - Attack was not blocked!", test.description, result.description)
			} else {
				t.Logf("%s: Successfully blocked - %s", test.description, result.description)
			}
		})
	}
}

// TestE2EMultiTenancy tests multi-tenant scenarios
func TestE2EMultiTenancy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E multi-tenancy test in short mode")
	}

	system := setupE2ESystem(t)
	defer system.cleanup()

	// Setup multi-tenant queues
	tenants := []struct {
		name     string
		queues   []string
		operator string
	}{
		{
			name:     "TenantA",
			queues:   []string{"tenant-a-high", "tenant-a-low"},
			operator: "operator-a@example.com",
		},
		{
			name:     "TenantB",
			queues:   []string{"tenant-b-priority", "tenant-b-batch"},
			operator: "operator-b@example.com",
		},
	}

	for _, tenant := range tenants {
		t.Run("Tenant_"+tenant.name, func(t *testing.T) {
			// Create tenant-specific token
			token := system.createResourceConstrainedToken(
				rbacandtokens.RoleOperator,
				[]rbacandtokens.Permission{rbacandtokens.PermQueueRead, rbacandtokens.PermQueueWrite},
				tenant.name+"-*",
				time.Hour,
			)

			// Test access to own queues
			for _, queue := range tenant.queues {
				result := system.executeOperation(e2eOperation{
					name:   fmt.Sprintf("Access own queue %s", queue),
					method: "GET",
					path:   fmt.Sprintf("/api/v1/queues/%s/peek", queue),
				}, token)

				if !result.success {
					t.Errorf("Tenant %s should access own queue %s", tenant.name, queue)
				}
			}

			// Test denial of access to other tenant queues
			for _, otherTenant := range tenants {
				if otherTenant.name == tenant.name {
					continue
				}

				for _, otherQueue := range otherTenant.queues {
					result := system.executeOperation(e2eOperation{
						name:   fmt.Sprintf("Deny access to %s queue %s", otherTenant.name, otherQueue),
						method: "GET",
						path:   fmt.Sprintf("/api/v1/queues/%s/peek", otherQueue),
					}, token)

					if result.success {
						t.Errorf("Tenant %s should not access other tenant queue %s", tenant.name, otherQueue)
					}
				}
			}
		})
	}
}

// Supporting types and functions

type e2eSystem struct {
	t       *testing.T
	server  *httptest.Server
	config  *adminapi.Config
	redis   *miniredis.Miniredis
	rdb     *redis.Client
	logger  *zap.Logger
}

type e2eOperation struct {
	name           string
	method         string
	path           string
	body           map[string]interface{}
	expectSuccess  bool
	validateResult func([]byte) error
}

type operationResult struct {
	success  bool
	response []byte
	status   int
	error    string
}

type attackResult struct {
	blocked     bool
	description string
}

func setupE2ESystem(t *testing.T) *e2eSystem {
	// Setup miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// Setup test data
	setupE2ETestData(mr)

	// Setup config
	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"high":                 "jobqueue:high",
				"medium":               "jobqueue:medium",
				"low":                  "jobqueue:low",
				"deployment":           "jobqueue:deployment",
				"email":                "jobqueue:email",
				"payment":              "jobqueue:payment",
				"tenant-a-high":        "jobqueue:tenant-a-high",
				"tenant-a-low":         "jobqueue:tenant-a-low",
				"tenant-b-priority":    "jobqueue:tenant-b-priority",
				"tenant-b-batch":       "jobqueue:tenant-b-batch",
			},
			DeadLetterList: "jobqueue:dead_letter",
		},
	}

	apiCfg := &adminapi.Config{
		JWTSecret:            "e2e-test-secret-key-for-comprehensive-testing",
		RequireAuth:          true,
		DenyByDefault:        true,
		RequireDoubleConfirm: true,
		ConfirmationPhrase:   "CONFIRM_DELETE",
		AuditEnabled:         true,
		AuditLogPath:         "/tmp/e2e-audit.log",
	}

	// Setup logger
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel) // Reduce noise
	logger, _ := config.Build()

	// Create server
	server, err := adminapi.NewServer(apiCfg, appCfg, rdb, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Setup middleware stack
	handler := server.SetupRoutes()
	handler = adminapi.AuthMiddleware(apiCfg.JWTSecret, apiCfg.DenyByDefault, logger)(handler)
	handler = adminapi.RequestIDMiddleware()(handler)

	ts := httptest.NewServer(handler)

	return &e2eSystem{
		t:      t,
		server: ts,
		config: apiCfg,
		redis:  mr,
		rdb:    rdb,
		logger: logger,
	}
}

func (sys *e2eSystem) cleanup() {
	sys.server.Close()
	sys.rdb.Close()
	sys.redis.Close()
	sys.logger.Sync()

	// Clean up audit log
	os.Remove(sys.config.AuditLogPath)
}

func (sys *e2eSystem) createToken(role rbacandtokens.Role, scopes []rbacandtokens.Permission, ttl time.Duration) string {
	if scopes == nil {
		// Use default role permissions
		rolePerms := rbacandtokens.GetRolePermissions()
		scopes = rolePerms[role]
	}

	return createE2EToken(sys.config.JWTSecret, role, scopes, ttl, nil)
}

func (sys *e2eSystem) createResourceConstrainedToken(role rbacandtokens.Role, scopes []rbacandtokens.Permission, resourcePattern string, ttl time.Duration) string {
	resources := map[string]string{"queues": resourcePattern}
	return createE2EToken(sys.config.JWTSecret, role, scopes, ttl, resources)
}

func (sys *e2eSystem) executeOperation(op e2eOperation, token string) operationResult {
	var body io.Reader
	if op.body != nil {
		jsonBody, _ := json.Marshal(op.body)
		body = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(op.method, sys.server.URL+op.path, body)
	if err != nil {
		return operationResult{success: false, error: fmt.Sprintf("Request creation failed: %v", err)}
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return operationResult{success: false, error: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	return operationResult{
		success:  resp.StatusCode >= 200 && resp.StatusCode < 300,
		response: responseBody,
		status:   resp.StatusCode,
	}
}

func setupE2ETestData(mr *miniredis.Miniredis) {
	// Add test jobs to various queues
	testJobs := map[string][]string{
		"jobqueue:high":              {`{"id":"high1","task":"urgent_task"}`},
		"jobqueue:medium":            {`{"id":"med1","task":"normal_task"}`},
		"jobqueue:low":               {`{"id":"low1","task":"background_task"}`},
		"jobqueue:deployment":        {`{"id":"deploy1","version":"v1.2.3"}`},
		"jobqueue:email":             {`{"id":"email1","template":"welcome"}`},
		"jobqueue:payment":           {`{"id":"pay1","amount":1000}`},
		"jobqueue:dead_letter":       {`{"id":"failed1","error":"timeout"}`},
		"jobqueue:tenant-a-high":     {`{"id":"ta1","tenant":"a"}`},
		"jobqueue:tenant-a-low":      {`{"id":"ta2","tenant":"a"}`},
		"jobqueue:tenant-b-priority": {`{"id":"tb1","tenant":"b"}`},
		"jobqueue:tenant-b-batch":    {`{"id":"tb2","tenant":"b"}`},
	}

	for queue, jobs := range testJobs {
		for _, job := range jobs {
			mr.Lpush(queue, job)
		}
	}
}

func createE2EToken(secret string, role rbacandtokens.Role, scopes []rbacandtokens.Permission, ttl time.Duration, resources map[string]string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	now := time.Now()
	claims := map[string]interface{}{
		"sub":        fmt.Sprintf("%s@example.com", strings.ToLower(string(role))),
		"roles":      []rbacandtokens.Role{role},
		"scopes":     scopes,
		"exp":        now.Add(ttl).Unix(),
		"iat":        now.Unix(),
		"iss":        "e2e-test-issuer",
		"aud":        "admin-api",
		"jti":        fmt.Sprintf("e2e-token-%d", now.UnixNano()),
		"kid":        "e2e-test-key-001",
		"token_type": "bearer",
	}

	if resources != nil {
		claims["resources"] = resources
	}

	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	message := header + "." + payload

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

func createMaliciousToken(wrongSecret string, role rbacandtokens.Role) string {
	// This will create a token with wrong signature
	return createE2EToken(wrongSecret, role, []rbacandtokens.Permission{rbacandtokens.PermAdminAll}, time.Hour, nil)
}

// Validation functions

func validateStatsResponse(response []byte) error {
	var stats map[string]interface{}
	if err := json.Unmarshal(response, &stats); err != nil {
		return fmt.Errorf("invalid stats response: %v", err)
	}

	// Check for expected fields
	expectedFields := []string{"total_jobs", "queue_counts"}
	for _, field := range expectedFields {
		if _, ok := stats[field]; !ok {
			return fmt.Errorf("missing expected field %s in stats", field)
		}
	}
	return nil
}

func validateJobEnqueue(response []byte) error {
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return fmt.Errorf("invalid enqueue response: %v", err)
	}

	if result["status"] != "enqueued" {
		return fmt.Errorf("expected enqueued status, got %v", result["status"])
	}
	return nil
}

func validateQueuePeek(response []byte) error {
	var jobs []interface{}
	if err := json.Unmarshal(response, &jobs); err != nil {
		return fmt.Errorf("invalid queue peek response: %v", err)
	}
	// Just verify it's a valid array response
	return nil
}

func validateDLQPurge(response []byte) error {
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return fmt.Errorf("invalid purge response: %v", err)
	}

	if result["purged"] == nil {
		return fmt.Errorf("expected purged count in response")
	}
	return nil
}

func validateWorkerRestart(response []byte) error {
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return fmt.Errorf("invalid restart response: %v", err)
	}

	if result["status"] != "restarted" {
		return fmt.Errorf("expected restarted status, got %v", result["status"])
	}
	return nil
}

func validateQueuePurgeAll(response []byte) error {
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return fmt.Errorf("invalid purge all response: %v", err)
	}

	if result["queues_purged"] == nil {
		return fmt.Errorf("expected queues_purged count in response")
	}
	return nil
}

func validateBenchmarkRun(response []byte) error {
	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return fmt.Errorf("invalid benchmark response: %v", err)
	}

	expectedFields := []string{"duration", "operations", "ops_per_second"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			return fmt.Errorf("missing expected field %s in benchmark result", field)
		}
	}
	return nil
}