// Copyright 2025 James Ross
package rbacandtokens

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestManager_GenerateToken(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false // Disable for testing
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	subject := "test@example.com"
	roles := []Role{RoleViewer, RoleOperator}
	scopes := []Permission{PermStatsRead, PermQueueRead}
	ttl := 1 * time.Hour

	token, err := manager.GenerateToken(subject, roles, scopes, ttl)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Validate the generated token
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate generated token: %v", err)
	}

	if claims.Subject != subject {
		t.Errorf("Expected subject %s, got %s", subject, claims.Subject)
	}

	if len(claims.Roles) != len(roles) {
		t.Errorf("Expected %d roles, got %d", len(roles), len(claims.Roles))
	}

	if len(claims.Scopes) != len(scopes) {
		t.Errorf("Expected %d scopes, got %d", len(scopes), len(claims.Scopes))
	}
}

func TestManager_ValidateToken(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test empty token
	_, err = manager.ValidateToken("")
	if err == nil {
		t.Error("Expected error for empty token")
	}

	// Test invalid token format
	_, err = manager.ValidateToken("invalid.token")
	if err == nil {
		t.Error("Expected error for invalid token format")
	}

	// Generate a valid token
	token, err := manager.GenerateToken("test@example.com", []Role{RoleViewer}, nil, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test valid token
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Errorf("Unexpected error validating valid token: %v", err)
	}

	if claims.Subject != "test@example.com" {
		t.Errorf("Expected subject test@example.com, got %s", claims.Subject)
	}

	// Test Bearer prefix
	claims, err = manager.ValidateToken("Bearer " + token)
	if err != nil {
		t.Errorf("Unexpected error validating token with Bearer prefix: %v", err)
	}

	if claims.Subject != "test@example.com" {
		t.Errorf("Expected subject test@example.com, got %s", claims.Subject)
	}
}

func TestManager_Authorize(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		name      string
		roles     []Role
		scopes    []Permission
		action    Permission
		resource  string
		expected  bool
	}{
		{
			name:     "Admin has all permissions",
			roles:    []Role{RoleAdmin},
			scopes:   []Permission{},
			action:   PermQueueDelete,
			resource: "/api/v1/queues/dlq",
			expected: true,
		},
		{
			name:     "Viewer can read stats",
			roles:    []Role{RoleViewer},
			scopes:   []Permission{},
			action:   PermStatsRead,
			resource: "/api/v1/stats",
			expected: true,
		},
		{
			name:     "Viewer cannot delete queues",
			roles:    []Role{RoleViewer},
			scopes:   []Permission{},
			action:   PermQueueDelete,
			resource: "/api/v1/queues/dlq",
			expected: false,
		},
		{
			name:     "Operator can run benchmarks",
			roles:    []Role{RoleOperator},
			scopes:   []Permission{},
			action:   PermBenchRun,
			resource: "/api/v1/bench",
			expected: true,
		},
		{
			name:     "Direct scope permission",
			roles:    []Role{},
			scopes:   []Permission{PermQueueWrite},
			action:   PermQueueWrite,
			resource: "/api/v1/queues/high",
			expected: true,
		},
		{
			name:     "No permission",
			roles:    []Role{},
			scopes:   []Permission{},
			action:   PermAdminAll,
			resource: "/api/v1/admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				Subject: "test@example.com",
				Roles:   tt.roles,
				Scopes:  tt.scopes,
			}

			result, err := manager.Authorize(claims, tt.action, tt.resource)
			if err != nil {
				t.Fatalf("Unexpected error in authorization: %v", err)
			}

			if result.Allowed != tt.expected {
				t.Errorf("Expected allowed=%v, got allowed=%v (reason: %s)",
					tt.expected, result.Allowed, result.Reason)
			}
		})
	}
}

func TestManager_RevokeToken(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Generate a token
	token, err := manager.GenerateToken("test@example.com", []Role{RoleViewer}, nil, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate it works
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	jwtID := claims.JWTID

	// Revoke the token
	err = manager.RevokeToken(jwtID, "Test revocation")
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Try to validate again - should fail
	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating revoked token")
	}

	if !isRevokedTokenError(err) {
		t.Errorf("Expected revoked token error, got: %v", err)
	}
}

func TestManager_GetTokenInfo(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	subject := "test@example.com"
	roles := []Role{RoleViewer}
	scopes := []Permission{PermStatsRead}

	token, err := manager.GenerateToken(subject, roles, scopes, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	info, err := manager.GetTokenInfo(token)
	if err != nil {
		t.Fatalf("Failed to get token info: %v", err)
	}

	if info.Subject != subject {
		t.Errorf("Expected subject %s, got %s", subject, info.Subject)
	}

	if len(info.Roles) != 1 || info.Roles[0] != RoleViewer {
		t.Errorf("Expected roles [%s], got %v", RoleViewer, info.Roles)
	}

	if len(info.Scopes) != 1 || info.Scopes[0] != PermStatsRead {
		t.Errorf("Expected scopes [%s], got %v", PermStatsRead, info.Scopes)
	}

	if info.TokenType != TokenTypeBearer {
		t.Errorf("Expected token type %s, got %s", TokenTypeBearer, info.TokenType)
	}
}

func TestRolePermissions(t *testing.T) {
	rolePerms := GetRolePermissions()

	// Test that admin has all permissions
	adminPerms := rolePerms[RoleAdmin]
	if len(adminPerms) != 1 || adminPerms[0] != PermAdminAll {
		t.Errorf("Expected admin to have PermAdminAll, got %v", adminPerms)
	}

	// Test that viewer has read permissions
	viewerPerms := rolePerms[RoleViewer]
	expectedViewerPerms := []Permission{PermStatsRead, PermQueueRead, PermJobRead, PermWorkerRead}
	if !permissionsEqual(viewerPerms, expectedViewerPerms) {
		t.Errorf("Expected viewer permissions %v, got %v", expectedViewerPerms, viewerPerms)
	}

	// Test that operator includes viewer permissions plus write permissions
	operatorPerms := rolePerms[RoleOperator]
	hasStatsRead := false
	hasQueueWrite := false
	hasBenchRun := false

	for _, perm := range operatorPerms {
		switch perm {
		case PermStatsRead:
			hasStatsRead = true
		case PermQueueWrite:
			hasQueueWrite = true
		case PermBenchRun:
			hasBenchRun = true
		}
	}

	if !hasStatsRead || !hasQueueWrite || !hasBenchRun {
		t.Errorf("Operator missing expected permissions. Has stats:read=%v, queue:write=%v, bench:run=%v",
			hasStatsRead, hasQueueWrite, hasBenchRun)
	}
}

func TestEndpointPermissions(t *testing.T) {
	endpointPerms := GetEndpointPermissions()

	tests := []struct {
		endpoint string
		expected Permission
	}{
		{"GET /api/v1/stats", PermStatsRead},
		{"GET /api/v1/stats/keys", PermStatsRead},
		{"DELETE /api/v1/queues/dlq", PermQueueDelete},
		{"POST /api/v1/bench", PermBenchRun},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			perms, exists := endpointPerms[tt.endpoint]
			if !exists {
				t.Errorf("No permissions defined for endpoint %s", tt.endpoint)
				return
			}

			if len(perms) == 0 {
				t.Errorf("Empty permissions for endpoint %s", tt.endpoint)
				return
			}

			if perms[0] != tt.expected {
				t.Errorf("Expected permission %s for %s, got %s", tt.expected, tt.endpoint, perms[0])
			}
		})
	}
}

func TestExpiredToken(t *testing.T) {
	config := DefaultConfig()
	config.AuditConfig.Enabled = false
	manager, err := NewManager(config, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Generate token with very short TTL
	token, err := manager.GenerateToken("test@example.com", []Role{RoleViewer}, nil, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired token
	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating expired token")
	}

	if !isExpiredTokenError(err) {
		t.Errorf("Expected expired token error, got: %v", err)
	}
}

// Helper functions for tests

func isRevokedTokenError(err error) bool {
	if authErr, ok := err.(*AuthenticationError); ok {
		return authErr.Code == "TOKEN_REVOKED"
	}
	return false
}

func isExpiredTokenError(err error) bool {
	if authErr, ok := err.(*AuthenticationError); ok {
		return authErr.Code == "TOKEN_EXPIRED"
	}
	return false
}

func permissionsEqual(a, b []Permission) bool {
	if len(a) != len(b) {
		return false
	}

	// Convert to maps for easier comparison
	mapA := make(map[Permission]bool)
	mapB := make(map[Permission]bool)

	for _, perm := range a {
		mapA[perm] = true
	}

	for _, perm := range b {
		mapB[perm] = true
	}

	for perm := range mapA {
		if !mapB[perm] {
			return false
		}
	}

	return true
}