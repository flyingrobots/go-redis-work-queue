// Copyright 2025 James Ross
package adminapi

import (
	"strings"
	"time"
)

type Config struct {
	// Server settings
	ListenAddr      string        `mapstructure:"listen_addr"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`

	// Auth settings
	JWTSecret     string `mapstructure:"jwt_secret"`
	JWTIssuer     string `mapstructure:"jwt_issuer"`
	RequireAuth   bool   `mapstructure:"require_auth"`
	DenyByDefault bool   `mapstructure:"deny_by_default"`

	// Rate limiting
	RateLimitEnabled   bool          `mapstructure:"rate_limit_enabled"`
	RateLimitPerMinute int           `mapstructure:"rate_limit_per_minute"`
	RateLimitBurst     int           `mapstructure:"rate_limit_burst"`
	RateLimitWindow    time.Duration `mapstructure:"rate_limit_window"`

	// Audit logging
	AuditEnabled    bool   `mapstructure:"audit_enabled"`
	AuditLogPath    string `mapstructure:"audit_log_path"`
	AuditRotateSize int64  `mapstructure:"audit_rotate_size"`
	AuditMaxBackups int    `mapstructure:"audit_max_backups"`

	// Security
	CORSEnabled      bool     `mapstructure:"cors_enabled"`
	CORSAllowOrigins []string `mapstructure:"cors_allow_origins"`
	TLSEnabled       bool     `mapstructure:"tls_enabled"`
	TLSCertFile      string   `mapstructure:"tls_cert_file"`
	TLSKeyFile       string   `mapstructure:"tls_key_file"`

	// Destructive operation confirmations
	RequireDoubleConfirm       bool   `mapstructure:"require_double_confirm"`
	ConfirmationPhrase         string `mapstructure:"confirmation_phrase"`
	DLQConfirmationPhrase      string `mapstructure:"dlq_confirmation_phrase"`
	PurgeAllConfirmationPhrase string `mapstructure:"purge_all_confirmation_phrase"`
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr:      ":8080",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,

		RequireAuth:   true,
		DenyByDefault: true,

		RateLimitEnabled:   true,
		RateLimitPerMinute: 100,
		RateLimitBurst:     10,
		RateLimitWindow:    time.Minute,

		AuditEnabled:    true,
		AuditLogPath:    "/var/log/admin-api/audit.log",
		AuditRotateSize: 100 * 1024 * 1024, // 100MB
		AuditMaxBackups: 10,

		CORSEnabled:      false,
		CORSAllowOrigins: []string{"*"},

		RequireDoubleConfirm:       true,
		ConfirmationPhrase:         "CONFIRM_DELETE",
		DLQConfirmationPhrase:      "CONFIRM_DELETE",
		PurgeAllConfirmationPhrase: "CONFIRM_DELETE_ALL",
	}
}

// DLQPhrase returns the configured DLQ confirmation phrase, falling back to the legacy confirmation phrase if unset.
func (c *Config) DLQPhrase() string {
	if c == nil {
		return ""
	}
	if strings.TrimSpace(c.DLQConfirmationPhrase) != "" {
		return c.DLQConfirmationPhrase
	}
	return c.ConfirmationPhrase
}

// PurgeAllPhrase returns the configured purge-all confirmation phrase, falling back to legacy conventions.
func (c *Config) PurgeAllPhrase() string {
	if c == nil {
		return ""
	}
	if strings.TrimSpace(c.PurgeAllConfirmationPhrase) != "" {
		return c.PurgeAllConfirmationPhrase
	}
	if strings.TrimSpace(c.ConfirmationPhrase) != "" {
		// Backwards compatibility: legacy configs expect `_ALL` suffix.
		return c.ConfirmationPhrase + "_ALL"
	}
	return ""
}
