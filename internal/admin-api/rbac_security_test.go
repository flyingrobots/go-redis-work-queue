// Copyright 2025 James Ross
//go:build security

package adminapi_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
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
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TestSecurityFuzzHeaders tests fuzzing of HTTP headers for security vulnerabilities
func TestSecurityFuzzHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security fuzz tests in short mode")
	}

	system := setupSecurityTestSystem(t)
	defer system.cleanup()

	validToken := system.createValidToken()

	// Fuzz test data for various headers
	fuzzPayloads := []struct {
		name   string
		header string
		values []string
		desc   string
	}{
		{
			name:   "Authorization Header Fuzzing",
			header: "Authorization",
			values: []string{
				"Bearer " + strings.Repeat("A", 10000),                              // Extremely long token
				"Bearer ../../../etc/passwd",                                        // Path traversal
				"Bearer <script>alert('xss')</script>",                              // XSS attempt
				"Bearer \x00\x01\x02\x03",                                           // Null bytes and control chars
				"Bearer ' OR 1=1 --",                                                // SQL injection attempt
				"Basic " + base64.StdEncoding.EncodeToString([]byte("admin:admin")), // Wrong auth type
				"Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.", // "none" algorithm
			},
			desc: "Authorization header should handle malformed values gracefully",
		},
		{
			name:   "Content-Type Header Fuzzing",
			header: "Content-Type",
			values: []string{
				"application/json; charset=../../../etc/passwd",
				"application/json\r\nX-Injected: malicious",
				"application/json\x00\x0a\x0dX-Evil: true",
				strings.Repeat("application/json", 1000),
			},
			desc: "Content-Type header should resist injection",
		},
		{
			name:   "User-Agent Header Fuzzing",
			header: "User-Agent",
			values: []string{
				strings.Repeat("A", 100000),     // Extremely long UA
				"<script>alert('xss')</script>", // XSS in UA
				"Mozilla/5.0\r\nX-Evil: injected\r\n\r\nGET /evil HTTP/1.1",
				"../../../etc/passwd",
				"User-Agent\x00\x0a\x0dHost: evil.com",
			},
			desc: "User-Agent should handle malicious content",
		},
		{
			name:   "X-Forwarded-For Header Fuzzing",
			header: "X-Forwarded-For",
			values: []string{
				strings.Repeat("127.0.0.1,", 10000), // IP flooding
				"<script>alert('xss')</script>",     // XSS in IP
				"192.168.1.1\r\nX-Evil: true",       // Header injection
				"' OR 1=1 --",                       // SQL injection
			},
			desc: "X-Forwarded-For should sanitize IP addresses",
		},
	}

	for _, fuzzTest := range fuzzPayloads {
		t.Run(fuzzTest.name, func(t *testing.T) {
			for i, maliciousValue := range fuzzTest.values {
				t.Run(fmt.Sprintf("Payload_%d", i), func(t *testing.T) {
					req := httptest.NewRequest("GET", "/api/v1/stats", nil)

					// Set the token for auth (except when fuzzing auth header)
					if fuzzTest.header != "Authorization" {
						req.Header.Set("Authorization", "Bearer "+validToken)
					}

					// Set the malicious header value
					req.Header.Set(fuzzTest.header, maliciousValue)

					w := httptest.NewRecorder()
					system.handler.ServeHTTP(w, req)

					// Verify system didn't crash or leak information
					if w.Code == 500 {
						// Read the response to check for info leaks
						body := w.Body.String()
						if strings.Contains(body, "panic") || strings.Contains(body, "stack trace") {
							t.Errorf("%s: System panic exposed in response: %s", fuzzTest.desc, body)
						}
					}

					// Check response headers for injections
					for headerName, headerValues := range w.Header() {
						for _, headerValue := range headerValues {
							if strings.Contains(headerValue, "<script>") ||
								strings.Contains(headerValue, "X-Evil") ||
								strings.Contains(headerValue, "X-Injected") {
								t.Errorf("%s: Header injection detected in response header %s: %s",
									fuzzTest.desc, headerName, headerValue)
							}
						}
					}

					// System should be stable (not 500 unless expected)
					if w.Code >= 500 && fuzzTest.header != "Authorization" {
						t.Errorf("%s: Unexpected server error %d for payload %q",
							fuzzTest.desc, w.Code, maliciousValue)
					}
				})
			}
		})
	}
}

// TestSecurityScopeEscalation tests various scope escalation attempts
func TestSecurityScopeEscalation(t *testing.T) {
	system := setupSecurityTestSystem(t)
	defer system.cleanup()

	escalationTests := []struct {
		name        string
		attack      func() (string, string)
		description string
	}{
		{
			name: "Token Scope Tampering",
			attack: func() (string, string) {
				// Create a viewer token, then manually tamper with scopes
				viewerToken := system.createTokenWithRole(rbacandtokens.RoleViewer)

				// Decode, modify scopes to admin, re-encode without proper signing
				parts := strings.Split(viewerToken, ".")
				if len(parts) != 3 {
					return "", "Invalid token format"
				}

				// Decode payload
				payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
				var claims map[string]interface{}
				json.Unmarshal(payload, &claims)

				// Escalate scopes
				claims["scopes"] = []string{string(rbacandtokens.PermAdminAll)}
				claims["roles"] = []string{string(rbacandtokens.RoleAdmin)}

				// Re-encode with original signature (should fail validation)
				modifiedPayload, _ := json.Marshal(claims)
				parts[1] = base64.RawURLEncoding.EncodeToString(modifiedPayload)
				tamperedToken := strings.Join(parts, ".")

				return tamperedToken, "DELETE /api/v1/queues/all"
			},
			description: "Tampered token scopes should be rejected",
		},
		{
			name: "Role Hierarchy Bypass",
			attack: func() (string, string) {
				// Try to create token with conflicting roles
				token := createMaliciousTokenWithRoles(
					system.secret,
					[]rbacandtokens.Role{rbacandtokens.RoleViewer, rbacandtokens.RoleAdmin},
					[]rbacandtokens.Permission{rbacandtokens.PermQueueRead, rbacandtokens.PermAdminAll},
				)
				return token, "DELETE /api/v1/queues/dlq"
			},
			description: "Conflicting roles should not grant highest permission",
		},
		{
			name: "Scope Injection",
			attack: func() (string, string) {
				// Try to inject additional scopes via malformed JSON
				token := createTokenWithMalformedClaims(system.secret)
				return token, "DELETE /api/v1/queues/all"
			},
			description: "Malformed token claims should be rejected",
		},
		{
			name: "Algorithm Confusion",
			attack: func() (string, string) {
				// Try "none" algorithm attack
				token := createUnsignedToken()
				return token, "GET /api/v1/stats"
			},
			description: "Unsigned tokens should be rejected",
		},
	}

	for _, test := range escalationTests {
		t.Run(test.name, func(t *testing.T) {
			tamperedToken, endpoint := test.attack()

			if tamperedToken == "" {
				t.Skip("Could not create attack token")
			}

			// Try to use the tampered token
			parts := strings.Split(endpoint, " ")
			method, path := parts[0], parts[1]

			var body *bytes.Buffer
			if method == "DELETE" {
				reqBody := map[string]interface{}{
					"confirmation": "CONFIRM_DELETE",
					"reason":       "security test",
				}
				if strings.Contains(path, "all") {
					reqBody["confirmation"] = "CONFIRM_DELETE_ALL"
				}
				jsonBody, _ := json.Marshal(reqBody)
				body = bytes.NewBuffer(jsonBody)
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(method, path, body)
			req.Header.Set("Authorization", "Bearer "+tamperedToken)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			system.handler.ServeHTTP(w, req)

			// Attack should be blocked
			if w.Code >= 200 && w.Code < 300 {
				t.Errorf("%s: Escalation attack succeeded when it should have been blocked", test.description)
			}

			// Specifically should be 401 (unauthorized) or 403 (forbidden)
			if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
				t.Errorf("%s: Expected 401/403, got %d", test.description, w.Code)
			}
		})
	}
}

// TestSecurityReplayAttacks tests various replay attack scenarios
func TestSecurityReplayAttacks(t *testing.T) {
	system := setupSecurityTestSystem(t)
	defer system.cleanup()

	replayTests := []struct {
		name        string
		setupAttack func() string
		description string
	}{
		{
			name: "Expired Token Replay",
			setupAttack: func() string {
				// Create expired token
				return system.createExpiredToken()
			},
			description: "Expired tokens should be rejected",
		},
		{
			name: "Future Token Replay",
			setupAttack: func() string {
				// Create token that's valid in the future (nbf in future)
				return system.createFutureToken()
			},
			description: "Tokens not yet valid should be rejected",
		},
		{
			name: "Clock Skew Exploitation",
			setupAttack: func() string {
				// Create token with suspicious time claims
				return system.createClockSkewToken()
			},
			description: "Tokens with excessive clock skew should be rejected",
		},
		{
			name: "Replay with Modified Claims",
			setupAttack: func() string {
				// Create valid token, then modify non-signature parts
				validToken := system.createValidToken()
				return modifyTokenTimestamps(validToken, system.secret)
			},
			description: "Tokens with modified timestamps should be rejected",
		},
	}

	for _, test := range replayTests {
		t.Run(test.name, func(t *testing.T) {
			attackToken := test.setupAttack()

			// Try multiple requests with the attack token
			endpoints := []string{
				"/api/v1/stats",
				"/api/v1/queues/high/peek",
			}

			for _, endpoint := range endpoints {
				req := httptest.NewRequest("GET", endpoint, nil)
				req.Header.Set("Authorization", "Bearer "+attackToken)

				w := httptest.NewRecorder()
				system.handler.ServeHTTP(w, req)

				// Should be rejected
				if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
					t.Errorf("%s: Replay attack not blocked for endpoint %s, got status %d",
						test.description, endpoint, w.Code)
				}
			}
		})
	}
}

// TestSecurityTimingAttacks tests for timing-based vulnerabilities
func TestSecurityTimingAttacks(t *testing.T) {
	system := setupSecurityTestSystem(t)
	defer system.cleanup()

	// Test token validation timing
	timingTests := []struct {
		name   string
		tokens []string
		desc   string
	}{
		{
			name: "Valid vs Invalid Token Timing",
			tokens: []string{
				system.createValidToken(),
				"invalid.token.here",
				system.createExpiredToken(),
				system.createFutureToken(),
			},
			desc: "Token validation should have consistent timing",
		},
	}

	for _, test := range timingTests {
		t.Run(test.name, func(t *testing.T) {
			timings := make([]time.Duration, len(test.tokens))

			for i, token := range test.tokens {
				start := time.Now()

				req := httptest.NewRequest("GET", "/api/v1/stats", nil)
				req.Header.Set("Authorization", "Bearer "+token)

				w := httptest.NewRecorder()
				system.handler.ServeHTTP(w, req)

				timings[i] = time.Since(start)
			}

			// Check for significant timing differences
			// In a real security test, you'd use statistical analysis
			// For this test, we just check that no timing is extremely different
			for i := 1; i < len(timings); i++ {
				ratio := float64(timings[i]) / float64(timings[0])
				if ratio > 10.0 || ratio < 0.1 {
					t.Logf("Warning: Potential timing difference detected: %v vs %v (ratio: %.2f)",
						timings[0], timings[i], ratio)
					// Don't fail the test as timing can be variable, just log
				}
			}
		})
	}
}

// TestSecurityResourceExhaustion tests for DoS vulnerabilities
func TestSecurityResourceExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource exhaustion tests in short mode")
	}

	system := setupSecurityTestSystem(t)
	defer system.cleanup()

	exhaustionTests := []struct {
		name   string
		attack func() error
		desc   string
	}{
		{
			name: "Large Token DoS",
			attack: func() error {
				// Create extremely large token
				largeToken := system.createLargeToken()

				req := httptest.NewRequest("GET", "/api/v1/stats", nil)
				req.Header.Set("Authorization", "Bearer "+largeToken)

				w := httptest.NewRecorder()
				system.handler.ServeHTTP(w, req)

				// Should handle gracefully without crashing
				if w.Code == 500 {
					return fmt.Errorf("server crashed processing large token")
				}
				return nil
			},
			desc: "Large tokens should be handled gracefully",
		},
		{
			name: "Rapid Request DoS",
			attack: func() error {
				token := system.createValidToken()

				// Send many rapid requests
				for i := 0; i < 100; i++ {
					req := httptest.NewRequest("GET", "/api/v1/stats", nil)
					req.Header.Set("Authorization", "Bearer "+token)

					w := httptest.NewRecorder()
					system.handler.ServeHTTP(w, req)

					// Should eventually rate limit
					if w.Code == http.StatusTooManyRequests {
						return nil // Rate limiting working
					}
				}
				return fmt.Errorf("rate limiting not triggered")
			},
			desc: "Rate limiting should prevent DoS",
		},
	}

	for _, test := range exhaustionTests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.attack(); err != nil {
				t.Errorf("%s: %v", test.desc, err)
			}
		})
	}
}

// Support types and functions

type securityTestSystem struct {
	t       *testing.T
	handler http.Handler
	server  *adminapi.Server
	config  *adminapi.Config
	secret  string
	redis   *miniredis.Miniredis
	rdb     *redis.Client
}

func setupSecurityTestSystem(t *testing.T) *securityTestSystem {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	appCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"high":   "jobqueue:high",
				"medium": "jobqueue:medium",
				"low":    "jobqueue:low",
			},
			DeadLetterList: "jobqueue:dead_letter",
		},
	}

	secret := "security-test-secret-key-for-comprehensive-security-testing"
	apiCfg := &adminapi.Config{
		JWTSecret:            secret,
		RequireAuth:          true,
		DenyByDefault:        true,
		RequireDoubleConfirm: true,
		ConfirmationPhrase:   "CONFIRM_DELETE",
	}

	server, err := adminapi.NewServer(apiCfg, appCfg, rdb, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Setup full middleware stack
	handler := server.SetupRoutes()
	handler = adminapi.RateLimitMiddleware(60, 10, zap.NewNop())(handler) // Add rate limiting
	handler = adminapi.AuthMiddleware(apiCfg.JWTSecret, apiCfg.DenyByDefault, zap.NewNop())(handler)
	handler = adminapi.RequestIDMiddleware()(handler)

	return &securityTestSystem{
		t:       t,
		handler: handler,
		server:  server,
		config:  apiCfg,
		secret:  secret,
		redis:   mr,
		rdb:     rdb,
	}
}

func (sys *securityTestSystem) cleanup() {
	sys.rdb.Close()
	sys.redis.Close()
}

func (sys *securityTestSystem) createValidToken() string {
	return createSecurityTestToken(sys.secret, rbacandtokens.RoleAdmin,
		[]rbacandtokens.Permission{rbacandtokens.PermAdminAll}, time.Hour)
}

func (sys *securityTestSystem) createTokenWithRole(role rbacandtokens.Role) string {
	rolePerms := rbacandtokens.GetRolePermissions()
	return createSecurityTestToken(sys.secret, role, rolePerms[role], time.Hour)
}

func (sys *securityTestSystem) createExpiredToken() string {
	return createSecurityTestToken(sys.secret, rbacandtokens.RoleAdmin,
		[]rbacandtokens.Permission{rbacandtokens.PermAdminAll}, -time.Hour)
}

func (sys *securityTestSystem) createFutureToken() string {
	return createSecurityTestTokenWithTimes(sys.secret, rbacandtokens.RoleAdmin,
		[]rbacandtokens.Permission{rbacandtokens.PermAdminAll},
		time.Now().Add(time.Hour),   // expires
		time.Now().Add(time.Hour),   // issued at (future)
		time.Now().Add(time.Hour/2)) // not before (future)
}

func (sys *securityTestSystem) createClockSkewToken() string {
	return createSecurityTestTokenWithTimes(sys.secret, rbacandtokens.RoleAdmin,
		[]rbacandtokens.Permission{rbacandtokens.PermAdminAll},
		time.Now().Add(time.Hour),      // expires
		time.Now().Add(10*time.Minute), // issued at (future - excessive skew)
		time.Now())                     // not before
}

func (sys *securityTestSystem) createLargeToken() string {
	// Create token with very large claims
	largeData := strings.Repeat("A", 100000)
	return createSecurityTestTokenWithCustomClaims(sys.secret, map[string]interface{}{
		"sub":        "test@example.com",
		"roles":      []string{string(rbacandtokens.RoleViewer)},
		"scopes":     []string{string(rbacandtokens.PermStatsRead)},
		"exp":        time.Now().Add(time.Hour).Unix(),
		"iat":        time.Now().Unix(),
		"large_data": largeData,
	})
}

// Token creation helpers

func createSecurityTestToken(secret string, role rbacandtokens.Role, scopes []rbacandtokens.Permission, ttl time.Duration) string {
	exp := time.Now().Add(ttl)
	iat := time.Now()
	nbf := time.Now()
	return createSecurityTestTokenWithTimes(secret, role, scopes, exp, iat, nbf)
}

func createSecurityTestTokenWithTimes(secret string, role rbacandtokens.Role, scopes []rbacandtokens.Permission,
	exp, iat, nbf time.Time) string {
	claims := map[string]interface{}{
		"sub":    "test@example.com",
		"roles":  []rbacandtokens.Role{role},
		"scopes": scopes,
		"exp":    exp.Unix(),
		"iat":    iat.Unix(),
		"nbf":    nbf.Unix(),
		"iss":    "security-test",
		"aud":    "admin-api",
		"jti":    fmt.Sprintf("security-test-%d", time.Now().UnixNano()),
	}
	return createSecurityTestTokenWithCustomClaims(secret, claims)
}

func createSecurityTestTokenWithCustomClaims(secret string, claims map[string]interface{}) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	message := header + "." + payload

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

func createMaliciousTokenWithRoles(secret string, roles []rbacandtokens.Role, scopes []rbacandtokens.Permission) string {
	claims := map[string]interface{}{
		"sub":    "malicious@example.com",
		"roles":  roles,
		"scopes": scopes,
		"exp":    time.Now().Add(time.Hour).Unix(),
		"iat":    time.Now().Unix(),
		"iss":    "malicious-issuer",
	}
	return createSecurityTestTokenWithCustomClaims(secret, claims)
}

func createTokenWithMalformedClaims(secret string) string {
	// Create token with malformed JSON injection attempt
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	// Malformed payload attempting injection
	malformedJSON := `{"sub":"test@example.com","roles":["admin"],"exp":` +
		fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()) +
		`,"scopes":["admin:all"],"injection":"value\",\"roles\":[\"admin\"],\"extra\":\"exploit"}`

	payload := base64.RawURLEncoding.EncodeToString([]byte(malformedJSON))
	message := header + "." + payload

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

func createUnsignedToken() string {
	// "none" algorithm attack
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))

	claims := map[string]interface{}{
		"sub":    "attacker@example.com",
		"roles":  []string{string(rbacandtokens.RoleAdmin)},
		"scopes": []string{string(rbacandtokens.PermAdminAll)},
		"exp":    time.Now().Add(time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	}

	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// No signature for "none" algorithm
	return header + "." + payload + "."
}

func modifyTokenTimestamps(token, secret string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return token
	}

	payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims map[string]interface{}
	json.Unmarshal(payload, &claims)

	// Modify timestamps to be suspicious
	claims["iat"] = time.Now().Add(-time.Hour).Unix()    // Issued in past
	claims["exp"] = time.Now().Add(2 * time.Hour).Unix() // Expires in future

	modifiedPayload, _ := json.Marshal(claims)
	parts[1] = base64.RawURLEncoding.EncodeToString(modifiedPayload)

	// Keep original signature (should fail validation)
	return strings.Join(parts, ".")
}
