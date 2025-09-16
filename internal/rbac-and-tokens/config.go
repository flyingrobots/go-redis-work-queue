// Copyright 2025 James Ross
package rbacandtokens

import (
	"time"
)

// Config represents the RBAC and token configuration
type Config struct {
	// Token configuration
	TokenConfig TokenConfig `yaml:"token" json:"token"`

	// Key management configuration
	KeyConfig KeyConfig `yaml:"keys" json:"keys"`

	// Audit configuration
	AuditConfig AuditConfig `yaml:"audit" json:"audit"`

	// Authorization configuration
	AuthzConfig AuthzConfig `yaml:"authz" json:"authz"`
}

// TokenConfig contains token-specific configuration
type TokenConfig struct {
	// Token format: "jwt" or "paseto"
	Format string `yaml:"format" json:"format"`

	// Default token lifetime
	DefaultTTL time.Duration `yaml:"default_ttl" json:"default_ttl"`

	// Maximum token lifetime
	MaxTTL time.Duration `yaml:"max_ttl" json:"max_ttl"`

	// Issuer identifier
	Issuer string `yaml:"issuer" json:"issuer"`

	// Audience identifier
	Audience string `yaml:"audience" json:"audience"`

	// Allow refresh tokens
	AllowRefresh bool `yaml:"allow_refresh" json:"allow_refresh"`

	// Refresh token TTL
	RefreshTTL time.Duration `yaml:"refresh_ttl" json:"refresh_ttl"`
}

// KeyConfig contains key management configuration
type KeyConfig struct {
	// Key rotation interval
	RotationInterval time.Duration `yaml:"rotation_interval" json:"rotation_interval"`

	// Key grace period after rotation
	GracePeriod time.Duration `yaml:"grace_period" json:"grace_period"`

	// Signing algorithm (HS256, RS256, etc.)
	Algorithm string `yaml:"algorithm" json:"algorithm"`

	// Key size for RSA keys
	KeySize int `yaml:"key_size" json:"key_size"`

	// Storage backend for keys
	Storage KeyStorageConfig `yaml:"storage" json:"storage"`
}

// KeyStorageConfig contains key storage configuration
type KeyStorageConfig struct {
	// Storage type: "memory", "file", "redis", "vault"
	Type string `yaml:"type" json:"type"`

	// Connection string or file path
	Connection string `yaml:"connection" json:"connection"`

	// Encryption key for stored keys
	EncryptionKey string `yaml:"encryption_key" json:"encryption_key"`
}

// AuditConfig contains audit logging configuration
type AuditConfig struct {
	// Enable audit logging
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Log file path
	LogPath string `yaml:"log_path" json:"log_path"`

	// Log rotation size (bytes)
	RotateSize int64 `yaml:"rotate_size" json:"rotate_size"`

	// Maximum backup files
	MaxBackups int `yaml:"max_backups" json:"max_backups"`

	// Log compression
	Compress bool `yaml:"compress" json:"compress"`

	// Retention period
	RetentionDays int `yaml:"retention_days" json:"retention_days"`

	// Filter sensitive fields
	FilterSensitive bool `yaml:"filter_sensitive" json:"filter_sensitive"`

	// Include request/response bodies
	IncludeBodies bool `yaml:"include_bodies" json:"include_bodies"`
}

// AuthzConfig contains authorization configuration
type AuthzConfig struct {
	// Default deny policy
	DefaultDeny bool `yaml:"default_deny" json:"default_deny"`

	// Cache authorization decisions
	CacheEnabled bool `yaml:"cache_enabled" json:"cache_enabled"`

	// Cache TTL for authorization decisions
	CacheTTL time.Duration `yaml:"cache_ttl" json:"cache_ttl"`

	// Role definitions file
	RolesFile string `yaml:"roles_file" json:"roles_file"`

	// Resource patterns file
	ResourcesFile string `yaml:"resources_file" json:"resources_file"`

	// Allow dynamic role assignment
	DynamicRoles bool `yaml:"dynamic_roles" json:"dynamic_roles"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		TokenConfig: TokenConfig{
			Format:       "jwt",
			DefaultTTL:   24 * time.Hour,
			MaxTTL:       7 * 24 * time.Hour,
			Issuer:       "redis-work-queue",
			Audience:     "admin-api",
			AllowRefresh: true,
			RefreshTTL:   7 * 24 * time.Hour,
		},
		KeyConfig: KeyConfig{
			RotationInterval: 30 * 24 * time.Hour, // 30 days
			GracePeriod:      24 * time.Hour,      // 1 day
			Algorithm:        "HS256",
			KeySize:          256,
			Storage: KeyStorageConfig{
				Type:       "file",
				Connection: "./keys",
			},
		},
		AuditConfig: AuditConfig{
			Enabled:         true,
			LogPath:         "./audit.log",
			RotateSize:      100 * 1024 * 1024, // 100MB
			MaxBackups:      10,
			Compress:        true,
			RetentionDays:   90,
			FilterSensitive: true,
			IncludeBodies:   false,
		},
		AuthzConfig: AuthzConfig{
			DefaultDeny:   true,
			CacheEnabled:  true,
			CacheTTL:      5 * time.Minute,
			RolesFile:     "./roles.yaml",
			ResourcesFile: "./resources.yaml",
			DynamicRoles:  false,
		},
	}
}

// GetRolePermissions returns the default permissions for each role
func GetRolePermissions() map[Role][]Permission {
	return map[Role][]Permission{
		RoleViewer: {
			PermStatsRead,
			PermQueueRead,
			PermJobRead,
			PermWorkerRead,
		},
		RoleOperator: {
			PermStatsRead,
			PermQueueRead,
			PermQueueWrite,
			PermJobRead,
			PermJobWrite,
			PermWorkerRead,
			PermBenchRun,
		},
		RoleMaintainer: {
			PermStatsRead,
			PermQueueRead,
			PermQueueWrite,
			PermQueueDelete,
			PermJobRead,
			PermJobWrite,
			PermJobDelete,
			PermWorkerRead,
			PermWorkerManage,
			PermBenchRun,
		},
		RoleAdmin: {
			PermAdminAll, // Admin has all permissions
		},
	}
}

// GetEndpointPermissions returns the required permissions for each endpoint
func GetEndpointPermissions() map[string][]Permission {
    return map[string][]Permission{
        "GET /api/v1/stats":        {PermStatsRead},
        "GET /api/v1/stats/keys":   {PermStatsRead},
        "GET /api/v1/queues/*/peek": {PermQueueRead},
        "DELETE /api/v1/queues/dlq": {PermQueueDelete},
        "DELETE /api/v1/queues/all": {PermQueueDelete, PermAdminAll}, // Requires admin
        "POST /api/v1/bench":        {PermBenchRun},
        // DLQ list/requeue/purge (selection)
        "GET /api/v1/dlq":          {PermQueueRead},
        "POST /api/v1/dlq/requeue":  {PermQueueWrite},
        "POST /api/v1/dlq/purge":    {PermQueueDelete},
        // Workers
        "GET /api/v1/workers":      {PermWorkerRead},
    }
}
