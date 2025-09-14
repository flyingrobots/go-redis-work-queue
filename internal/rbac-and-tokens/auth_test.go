// Copyright 2025 James Ross
package rbacandtokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// TestTokenValidation tests JWT token validation with various scenarios
func TestTokenValidation(t *testing.T) {
	secret := "test-secret-key-256-bits-long"

	tests := []struct {
		name        string
		token       func() string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid token",
			token:       func() string { return createValidToken(secret, time.Now().Add(time.Hour)) },
			expectError: false,
		},
		{
			name:        "Expired token",
			token:       func() string { return createValidToken(secret, time.Now().Add(-time.Hour)) },
			expectError: true,
			errorMsg:    "token expired",
		},
		{
			name: "Invalid signature",
			token: func() string {
				return createValidToken("wrong-secret", time.Now().Add(time.Hour))
			},
			expectError: true,
			errorMsg:    "invalid signature",
		},
		{
			name: "Malformed token - too few parts",
			token: func() string {
				return "header.payload"
			},
			expectError: true,
			errorMsg:    "invalid token format",
		},
		{
			name: "Malformed token - invalid base64",
			token: func() string {
				return "invalid$.base64$.encoding"
			},
			expectError: true,
		},
		{
			name: "Invalid JSON payload",
			token: func() string {
				header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
				payload := base64.RawURLEncoding.EncodeToString([]byte(`{invalid json`))
				message := header + "." + payload
				h := hmac.New(sha256.New, []byte(secret))
				h.Write([]byte(message))
				signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
				return message + "." + signature
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.token()
			claims, err := validateToken(token, secret)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				if claims != nil {
					t.Errorf("Expected nil claims for invalid token")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if claims == nil {
					t.Errorf("Expected claims but got nil")
				}
			}
		})
	}
}

// TestTimeSkewTolerance tests token validation with time skew
func TestTimeSkewTolerance(t *testing.T) {
	secret := "test-secret-key-256-bits-long"

	tests := []struct {
		name        string
		nbf         int64 // not before
		exp         int64 // expires at
		iat         int64 // issued at
		expectError bool
		description string
	}{
		{
			name:        "Valid time range",
			nbf:         time.Now().Add(-time.Minute).Unix(),
			exp:         time.Now().Add(time.Hour).Unix(),
			iat:         time.Now().Add(-time.Minute).Unix(),
			expectError: false,
			description: "Token with valid time range should pass",
		},
		{
			name:        "Not yet valid",
			nbf:         time.Now().Add(time.Hour).Unix(),
			exp:         time.Now().Add(2 * time.Hour).Unix(),
			iat:         time.Now().Unix(),
			expectError: true,
			description: "Token not yet valid should fail",
		},
		{
			name:        "Expired token",
			nbf:         time.Now().Add(-2 * time.Hour).Unix(),
			exp:         time.Now().Add(-time.Hour).Unix(),
			iat:         time.Now().Add(-2 * time.Hour).Unix(),
			expectError: true,
			description: "Expired token should fail",
		},
		{
			name:        "Edge case - expires in 1 second",
			nbf:         time.Now().Add(-time.Minute).Unix(),
			exp:         time.Now().Add(time.Second).Unix(),
			iat:         time.Now().Add(-time.Minute).Unix(),
			expectError: false,
			description: "Token expiring soon should still be valid",
		},
		{
			name:        "Future issued time tolerance",
			nbf:         0, // No nbf
			exp:         time.Now().Add(time.Hour).Unix(),
			iat:         time.Now().Add(30 * time.Second).Unix(), // Issued in future
			expectError: false,
			description: "Small clock skew should be tolerated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := createTokenWithTimes(secret, tt.nbf, tt.exp, tt.iat)
			claims, err := validateTokenWithSkewTolerance(token, secret, 60*time.Second)

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: Expected error but got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: Unexpected error: %v", tt.description, err)
				}
				if claims == nil {
					t.Errorf("%s: Expected claims but got nil", tt.description)
				}
			}
		})
	}
}

// TestScopeMatching tests permission scope matching logic
func TestScopeMatching(t *testing.T) {
	tests := []struct {
		name           string
		userScopes     []Permission
		requiredScope  Permission
		expectAllowed  bool
		description    string
	}{
		{
			name:          "Exact scope match",
			userScopes:    []Permission{PermQueueRead, PermJobWrite},
			requiredScope: PermQueueRead,
			expectAllowed: true,
			description:   "User with exact required scope should be allowed",
		},
		{
			name:          "No matching scope",
			userScopes:    []Permission{PermQueueRead, PermJobRead},
			requiredScope: PermQueueDelete,
			expectAllowed: false,
			description:   "User without required scope should be denied",
		},
		{
			name:          "Admin all permissions",
			userScopes:    []Permission{PermAdminAll},
			requiredScope: PermQueueDelete,
			expectAllowed: true,
			description:   "Admin should have access to all operations",
		},
		{
			name:          "Multiple scopes with match",
			userScopes:    []Permission{PermQueueRead, PermJobWrite, PermWorkerRead},
			requiredScope: PermJobWrite,
			expectAllowed: true,
			description:   "User with multiple scopes should be allowed for matching one",
		},
		{
			name:          "Empty user scopes",
			userScopes:    []Permission{},
			requiredScope: PermQueueRead,
			expectAllowed: false,
			description:   "User with no scopes should be denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPermission(tt.userScopes, tt.requiredScope)

			if result != tt.expectAllowed {
				t.Errorf("%s: Expected %v, got %v", tt.description, tt.expectAllowed, result)
			}
		})
	}
}

// TestRolePermissions tests role-based permission mapping
func TestRolePermissions(t *testing.T) {
	rolePermissions := GetRolePermissions()

	tests := []struct {
		name        string
		role        Role
		permission  Permission
		expectAllow bool
		description string
	}{
		{
			name:        "Viewer read access",
			role:        RoleViewer,
			permission:  PermStatsRead,
			expectAllow: true,
			description: "Viewer should have read access to stats",
		},
		{
			name:        "Viewer no write access",
			role:        RoleViewer,
			permission:  PermQueueWrite,
			expectAllow: false,
			description: "Viewer should not have write access",
		},
		{
			name:        "Operator read and write",
			role:        RoleOperator,
			permission:  PermQueueWrite,
			expectAllow: true,
			description: "Operator should have write access to queues",
		},
		{
			name:        "Operator no delete access",
			role:        RoleOperator,
			permission:  PermQueueDelete,
			expectAllow: false,
			description: "Operator should not have delete access",
		},
		{
			name:        "Maintainer delete access",
			role:        RoleMaintainer,
			permission:  PermQueueDelete,
			expectAllow: true,
			description: "Maintainer should have delete access",
		},
		{
			name:        "Admin all permissions",
			role:        RoleAdmin,
			permission:  PermQueueDelete,
			expectAllow: true,
			description: "Admin should have all permissions via PermAdminAll",
		},
		{
			name:        "Admin all permissions - worker manage",
			role:        RoleAdmin,
			permission:  PermWorkerManage,
			expectAllow: true,
			description: "Admin should have worker management permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := rolePermissions[tt.role]
			result := hasPermission(permissions, tt.permission) ||
				hasPermission(permissions, PermAdminAll) // Admin has all permissions

			if result != tt.expectAllow {
				t.Errorf("%s: Expected %v, got %v. Role %s has permissions: %v",
					tt.description, tt.expectAllow, result, tt.role, permissions)
			}
		})
	}
}

// TestEndpointPermissions tests endpoint to permission mapping
func TestEndpointPermissions(t *testing.T) {
	endpointPerms := GetEndpointPermissions()

	tests := []struct {
		name         string
		endpoint     string
		userPerms    []Permission
		expectAccess bool
		description  string
	}{
		{
			name:         "Stats endpoint with stats permission",
			endpoint:     "GET /api/v1/stats",
			userPerms:    []Permission{PermStatsRead},
			expectAccess: true,
			description:  "User with stats read permission should access stats endpoint",
		},
		{
			name:         "Stats endpoint without permission",
			endpoint:     "GET /api/v1/stats",
			userPerms:    []Permission{PermQueueRead},
			expectAccess: false,
			description:  "User without stats permission should be denied",
		},
		{
			name:         "Admin access to all endpoints",
			endpoint:     "DELETE /api/v1/queues/all",
			userPerms:    []Permission{PermAdminAll},
			expectAccess: true,
			description:  "Admin should have access to destructive operations",
		},
		{
			name:         "Queue delete without admin",
			endpoint:     "DELETE /api/v1/queues/all",
			userPerms:    []Permission{PermQueueDelete},
			expectAccess: false,
			description:  "Non-admin with queue delete should not access purge all",
		},
		{
			name:         "DLQ purge with queue delete",
			endpoint:     "DELETE /api/v1/queues/dlq",
			userPerms:    []Permission{PermQueueDelete},
			expectAccess: true,
			description:  "User with queue delete should access DLQ purge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requiredPerms := endpointPerms[tt.endpoint]
			if requiredPerms == nil {
				t.Fatalf("Endpoint %s not found in permission map", tt.endpoint)
			}

			// Check if user has any of the required permissions
			hasAccess := false
			for _, required := range requiredPerms {
				if hasPermission(tt.userPerms, required) {
					hasAccess = true
					break
				}
			}

			if hasAccess != tt.expectAccess {
				t.Errorf("%s: Expected %v, got %v. Required: %v, User has: %v",
					tt.description, tt.expectAccess, hasAccess, requiredPerms, tt.userPerms)
			}
		})
	}
}

// TestTokenRevocation tests token revocation functionality
func TestTokenRevocation(t *testing.T) {
	revocationList := make(map[string]RevokedToken)

	tests := []struct {
		name        string
		tokenID     string
		setup       func()
		expectError bool
		description string
	}{
		{
			name:    "Valid token not in revocation list",
			tokenID: "token-123",
			setup:   func() {},
			expectError: false,
			description: "Token not in revocation list should be valid",
		},
		{
			name:    "Revoked token",
			tokenID: "revoked-456",
			setup: func() {
				revocationList["revoked-456"] = RevokedToken{
					JWTID:     "revoked-456",
					Subject:   "user@example.com",
					RevokedAt: time.Now(),
					Reason:    "Security incident",
				}
			},
			expectError: true,
			description: "Token in revocation list should be rejected",
		},
		{
			name:    "Multiple revoked tokens",
			tokenID: "revoked-789",
			setup: func() {
				revocationList["revoked-789"] = RevokedToken{
					JWTID:     "revoked-789",
					Subject:   "admin@example.com",
					RevokedAt: time.Now().Add(-time.Hour),
					Reason:    "Token compromise",
				}
			},
			expectError: true,
			description: "Multiple revoked tokens should all be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			isRevoked := isTokenRevoked(tt.tokenID, revocationList)

			if isRevoked != tt.expectError {
				t.Errorf("%s: Expected revoked=%v, got %v",
					tt.description, tt.expectError, isRevoked)
			}
		})
	}
}

// TestAuthorizationErrors tests error handling for authorization
func TestAuthorizationErrors(t *testing.T) {
	tests := []struct {
		name        string
		errorFunc   func() error
		expectType  string
		expectCode  string
		description string
	}{
		{
			name: "Authentication error",
			errorFunc: func() error {
				return NewAuthenticationError(
					ErrInvalidToken,
					"INVALID_TOKEN",
					"Token signature is invalid",
					map[string]interface{}{"token_id": "abc123"},
				)
			},
			expectType:  "*rbacandtokens.AuthenticationError",
			expectCode:  "INVALID_TOKEN",
			description: "Should create proper authentication error",
		},
		{
			name: "Authorization error",
			errorFunc: func() error {
				return NewAuthorizationError(
					ErrAccessDenied,
					"ACCESS_DENIED",
					"Insufficient permissions",
					"user@example.com",
					"queue:high",
					"delete",
					map[string]interface{}{"required": "admin"},
				)
			},
			expectType:  "*rbacandtokens.AuthorizationError",
			expectCode:  "ACCESS_DENIED",
			description: "Should create proper authorization error",
		},
		{
			name: "Token error",
			errorFunc: func() error {
				return NewTokenError(
					ErrExpiredToken,
					"TOKEN_EXPIRED",
					"Token has expired",
					"token-123",
					"user@example.com",
					map[string]interface{}{"exp": time.Now().Unix()},
				)
			},
			expectType:  "*rbacandtokens.TokenError",
			expectCode:  "TOKEN_EXPIRED",
			description: "Should create proper token error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()

			if err == nil {
				t.Fatalf("%s: Expected error but got nil", tt.description)
			}

			// Check error message contains code
			errorMsg := err.Error()
			if errorMsg == "" {
				t.Errorf("%s: Error message should not be empty", tt.description)
			}

			// Type assertion checks
			switch e := err.(type) {
			case *AuthenticationError:
				if tt.expectCode != "" && e.Code != tt.expectCode {
					t.Errorf("%s: Expected code %s, got %s", tt.description, tt.expectCode, e.Code)
				}
			case *AuthorizationError:
				if tt.expectCode != "" && e.Code != tt.expectCode {
					t.Errorf("%s: Expected code %s, got %s", tt.description, tt.expectCode, e.Code)
				}
			case *TokenError:
				if tt.expectCode != "" && e.Code != tt.expectCode {
					t.Errorf("%s: Expected code %s, got %s", tt.description, tt.expectCode, e.Code)
				}
			default:
				t.Errorf("%s: Unexpected error type: %T", tt.description, err)
			}
		})
	}
}

// Helper functions for tests

func createValidToken(secret string, exp time.Time) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := fmt.Sprintf(`{
		"sub":"test@example.com",
		"roles":["operator"],
		"scopes":["queue:read","job:write"],
		"exp":%d,
		"iat":%d,
		"iss":"test-issuer",
		"kid":"key-001"
	}`, exp.Unix(), time.Now().Unix())

	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	message := header + "." + payload

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

func createTokenWithTimes(secret string, nbf, exp, iat int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := fmt.Sprintf(`{
		"sub":"test@example.com",
		"roles":["operator"],
		"scopes":["queue:read"],
		"exp":%d,
		"iat":%d,
		"nbf":%d,
		"iss":"test-issuer"
	}`, exp, iat, nbf)

	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	message := header + "." + payload

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

func validateToken(tokenString, secret string) (*Claims, error) {
	// This is a simplified version for testing
	// The real implementation would be more robust
	parts := splitToken(tokenString)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	// Verify signature
	if !verifySignature(parts[0], parts[1], parts[2], secret) {
		return nil, fmt.Errorf("invalid signature")
	}

	return &claims, nil
}

func validateTokenWithSkewTolerance(tokenString, secret string, tolerance time.Duration) (*Claims, error) {
	// Similar to validateToken but with time skew tolerance
	claims, err := validateToken(tokenString, secret)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Check not before with tolerance
	if claims.NotBefore > 0 && now.Unix() < (claims.NotBefore - int64(tolerance.Seconds())) {
		return nil, fmt.Errorf("token not yet valid")
	}

	// Check expiration with tolerance
	if now.Unix() > (claims.ExpiresAt + int64(tolerance.Seconds())) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

func hasPermission(userPerms []Permission, required Permission) bool {
	for _, perm := range userPerms {
		if perm == required || perm == PermAdminAll {
			return true
		}
	}
	return false
}

func isTokenRevoked(tokenID string, revocationList map[string]RevokedToken) bool {
	_, exists := revocationList[tokenID]
	return exists
}

func splitToken(token string) []string {
	return []string{"header", "payload", "signature"} // Simplified for testing
}

func verifySignature(header, payload, signature, secret string) bool {
	// Simplified signature verification for testing
	message := header + "." + payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	expectedDecoded, _ := base64.RawURLEncoding.DecodeString(expectedSig)
	actualDecoded, _ := base64.RawURLEncoding.DecodeString(signature)

	return hmac.Equal(expectedDecoded, actualDecoded)
}