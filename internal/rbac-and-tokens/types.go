// Copyright 2025 James Ross
package rbacandtokens

import (
	"time"
)

// Role represents the available roles in the system
type Role string

const (
	RoleAdmin      Role = "admin"
	RoleOperator   Role = "operator"
	RoleMaintainer Role = "maintainer"
	RoleViewer     Role = "viewer"
)

// Permission represents individual permissions that can be granted
type Permission string

const (
	// Queue permissions
	PermQueueRead   Permission = "queue:read"
	PermQueueWrite  Permission = "queue:write"
	PermQueueDelete Permission = "queue:delete"

	// Job permissions
	PermJobRead   Permission = "job:read"
	PermJobWrite  Permission = "job:write"
	PermJobDelete Permission = "job:delete"

	// Worker permissions
	PermWorkerRead   Permission = "worker:read"
	PermWorkerManage Permission = "worker:manage"

	// Admin permissions
	PermAdminAll Permission = "admin:all"

	// Stats permissions
	PermStatsRead Permission = "stats:read"

	// Benchmark permissions
	PermBenchRun Permission = "bench:run"
)

// TokenType represents the type of token
type TokenType string

const (
	TokenTypeBearer  TokenType = "bearer"
	TokenTypeAPIKey  TokenType = "api_key"
	TokenTypeSession TokenType = "session"
)

// Claims represents the JWT/PASETO token claims
type Claims struct {
	Subject   string       `json:"sub"`           // User identifier
	Issuer    string       `json:"iss"`           // Token issuer
	Audience  string       `json:"aud"`           // Intended audience
	ExpiresAt int64        `json:"exp"`           // Expiration time
	IssuedAt  int64        `json:"iat"`           // Issued at time
	NotBefore int64        `json:"nbf,omitempty"` // Not valid before
	JWTID     string       `json:"jti,omitempty"` // JWT ID
	KeyID     string       `json:"kid"`           // Key ID for rotation
	Roles     []Role       `json:"roles"`         // User roles
	Scopes    []Permission `json:"scopes"`        // Permitted actions
	TokenType TokenType    `json:"token_type"`    // Type of token
}

// TokenInfo represents information about a token for display
type TokenInfo struct {
	Subject   string       `json:"subject"`
	Roles     []Role       `json:"roles"`
	Scopes    []Permission `json:"scopes"`
	ExpiresAt time.Time    `json:"expires_at"`
	IssuedAt  time.Time    `json:"issued_at"`
	TokenType TokenType    `json:"token_type"`
	KeyID     string       `json:"key_id"`
}

// KeyPair represents a signing key pair
type KeyPair struct {
	ID         string    `json:"id"`          // Key ID
	Algorithm  string    `json:"algorithm"`   // Signing algorithm (HS256, RS256, etc.)
	PublicKey  string    `json:"public_key"`  // Base64 encoded public key
	PrivateKey string    `json:"private_key"` // Base64 encoded private key
	CreatedAt  time.Time `json:"created_at"`  // Creation time
	ExpiresAt  time.Time `json:"expires_at"`  // Key expiration
	Active     bool      `json:"active"`      // Whether key is active
}

// RevokedToken represents a revoked token entry
type RevokedToken struct {
	JWTID     string    `json:"jti"`
	Subject   string    `json:"subject"`
	RevokedAt time.Time `json:"revoked_at"`
	Reason    string    `json:"reason"`
}

// AuthorizationResult represents the result of an authorization check
type AuthorizationResult struct {
	Allowed   bool        `json:"allowed"`
	Subject   string      `json:"subject"`
	Roles     []Role      `json:"roles"`
	Scopes    []Permission `json:"scopes"`
	Reason    string      `json:"reason,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Subject   string                 `json:"subject"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	Reason    string                 `json:"reason,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IP        string                 `json:"ip"`
	UserAgent string                 `json:"user_agent"`
	RequestID string                 `json:"request_id"`
}

// ResourcePattern represents a resource access pattern
type ResourcePattern struct {
	Pattern string   `json:"pattern"` // e.g., "queue:high:*", "worker:*"
	Actions []string `json:"actions"` // e.g., ["read", "write"]
}

// RoleDefinition represents a role with its permissions
type RoleDefinition struct {
	Name        Role              `json:"name"`
	Description string            `json:"description"`
	Permissions []Permission      `json:"permissions"`
	Resources   []ResourcePattern `json:"resources"`
}