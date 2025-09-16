package collaborativesession

import (
	"fmt"
	"time"
)

// Config represents configuration for the collaborative session system
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server" yaml:"server"`

	// Session configuration
	Session SessionConfig `json:"session" yaml:"session"`

	// Transport configuration
	Transport TransportConfig `json:"transport" yaml:"transport"`

	// Security configuration
	Security SecurityConfig `json:"security" yaml:"security"`

	// Redaction configuration
	Redaction RedactionConfig `json:"redaction" yaml:"redaction"`

	// Logging configuration
	Logging LoggingConfig `json:"logging" yaml:"logging"`
}

// ServerConfig configures the session server
type ServerConfig struct {
	// Address to bind the server to
	Address string `json:"address" yaml:"address"`

	// Port to listen on
	Port int `json:"port" yaml:"port"`

	// ReadTimeout for client connections
	ReadTimeout time.Duration `json:"read_timeout" yaml:"read_timeout"`

	// WriteTimeout for client connections
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`

	// IdleTimeout for idle connections
	IdleTimeout time.Duration `json:"idle_timeout" yaml:"idle_timeout"`

	// MaxConnections limits concurrent connections
	MaxConnections int `json:"max_connections" yaml:"max_connections"`

	// EnableMetrics enables metrics collection
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`

	// MetricsPath is the HTTP path for metrics endpoint
	MetricsPath string `json:"metrics_path" yaml:"metrics_path"`
}

// SessionConfig configures session behavior
type SessionConfig struct {
	// DefaultExpiryDuration is the default session expiry time
	DefaultExpiryDuration time.Duration `json:"default_expiry_duration" yaml:"default_expiry_duration"`

	// MaxSessionDuration is the maximum allowed session duration
	MaxSessionDuration time.Duration `json:"max_session_duration" yaml:"max_session_duration"`

	// MaxParticipants is the default maximum participants per session
	MaxParticipants int `json:"max_participants" yaml:"max_participants"`

	// CleanupInterval for expired sessions
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`

	// DefaultFrameRate is the default frame rate for sessions
	DefaultFrameRate int `json:"default_frame_rate" yaml:"default_frame_rate"`

	// MaxFrameRate is the maximum allowed frame rate
	MaxFrameRate int `json:"max_frame_rate" yaml:"max_frame_rate"`

	// ControlTimeout is the default control timeout
	ControlTimeout time.Duration `json:"control_timeout" yaml:"control_timeout"`

	// HandoffTimeout is timeout for handoff requests
	HandoffTimeout time.Duration `json:"handoff_timeout" yaml:"handoff_timeout"`

	// RequireApproval determines if control handoff requires approval
	RequireApproval bool `json:"require_approval" yaml:"require_approval"`

	// AllowControlHandoff enables control handoff feature
	AllowControlHandoff bool `json:"allow_control_handoff" yaml:"allow_control_handoff"`
}

// TransportConfig configures the transport layer
type TransportConfig struct {
	// Type specifies the transport type (websocket, tcp, etc.)
	Type string `json:"type" yaml:"type"`

	// EnableCompression enables frame compression
	EnableCompression bool `json:"enable_compression" yaml:"enable_compression"`

	// CompressionLevel sets the compression level (1-9)
	CompressionLevel int `json:"compression_level" yaml:"compression_level"`

	// MaxMessageSize limits message size
	MaxMessageSize int64 `json:"max_message_size" yaml:"max_message_size"`

	// PingInterval for keepalive pings
	PingInterval time.Duration `json:"ping_interval" yaml:"ping_interval"`

	// PongTimeout for ping responses
	PongTimeout time.Duration `json:"pong_timeout" yaml:"pong_timeout"`

	// BufferSize for message buffers
	BufferSize int `json:"buffer_size" yaml:"buffer_size"`

	// EnableBinary enables binary message format
	EnableBinary bool `json:"enable_binary" yaml:"enable_binary"`
}

// SecurityConfig configures security settings
type SecurityConfig struct {
	// EnableTLS enables TLS encryption
	EnableTLS bool `json:"enable_tls" yaml:"enable_tls"`

	// TLSCertFile path to TLS certificate
	TLSCertFile string `json:"tls_cert_file" yaml:"tls_cert_file"`

	// TLSKeyFile path to TLS private key
	TLSKeyFile string `json:"tls_key_file" yaml:"tls_key_file"`

	// TokenSigningKey for JWT token signing
	TokenSigningKey string `json:"token_signing_key" yaml:"token_signing_key"`

	// TokenExpiryDuration default token expiry
	TokenExpiryDuration time.Duration `json:"token_expiry_duration" yaml:"token_expiry_duration"`

	// AllowedOrigins for CORS
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins"`

	// RequireAuth enables authentication
	RequireAuth bool `json:"require_auth" yaml:"require_auth"`

	// RateLimitRequests per minute per IP
	RateLimitRequests int `json:"rate_limit_requests" yaml:"rate_limit_requests"`

	// RateLimitDuration for rate limiting window
	RateLimitDuration time.Duration `json:"rate_limit_duration" yaml:"rate_limit_duration"`
}

// RedactionConfig configures frame redaction
type RedactionConfig struct {
	// EnableRedaction enables the redaction feature
	EnableRedaction bool `json:"enable_redaction" yaml:"enable_redaction"`

	// DefaultPatterns are always applied
	DefaultPatterns []string `json:"default_patterns" yaml:"default_patterns"`

	// CustomPatterns can be configured per session
	CustomPatterns []string `json:"custom_patterns" yaml:"custom_patterns"`

	// RedactionChar is the character used for redaction
	RedactionChar rune `json:"redaction_char" yaml:"redaction_char"`

	// CaseSensitive determines if pattern matching is case sensitive
	CaseSensitive bool `json:"case_sensitive" yaml:"case_sensitive"`

	// RedactCompleteLine redacts entire lines containing matches
	RedactCompleteLine bool `json:"redact_complete_line" yaml:"redact_complete_line"`
}

// LoggingConfig configures logging
type LoggingConfig struct {
	// Level sets the log level
	Level string `json:"level" yaml:"level"`

	// Format sets the log format (json, text)
	Format string `json:"format" yaml:"format"`

	// EnableSessionLogs enables session event logging
	EnableSessionLogs bool `json:"enable_session_logs" yaml:"enable_session_logs"`

	// EnableAccessLogs enables access logging
	EnableAccessLogs bool `json:"enable_access_logs" yaml:"enable_access_logs"`

	// LogFile path for log output
	LogFile string `json:"log_file" yaml:"log_file"`

	// MaxLogSize maximum log file size in MB
	MaxLogSize int `json:"max_log_size" yaml:"max_log_size"`

	// MaxLogBackups number of backup log files
	MaxLogBackups int `json:"max_log_backups" yaml:"max_log_backups"`

	// LogRetentionDays days to retain logs
	LogRetentionDays int `json:"log_retention_days" yaml:"log_retention_days"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address:        "0.0.0.0",
			Port:           8080,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			IdleTimeout:    120 * time.Second,
			MaxConnections: 1000,
			EnableMetrics:  true,
			MetricsPath:    "/metrics",
		},
		Session: SessionConfig{
			DefaultExpiryDuration: 2 * time.Hour,
			MaxSessionDuration:    8 * time.Hour,
			MaxParticipants:       10,
			CleanupInterval:       5 * time.Minute,
			DefaultFrameRate:      30,
			MaxFrameRate:          60,
			ControlTimeout:        5 * time.Minute,
			HandoffTimeout:        30 * time.Second,
			RequireApproval:       true,
			AllowControlHandoff:   true,
		},
		Transport: TransportConfig{
			Type:              "websocket",
			EnableCompression: true,
			CompressionLevel:  6,
			MaxMessageSize:    1024 * 1024, // 1MB
			PingInterval:      30 * time.Second,
			PongTimeout:       10 * time.Second,
			BufferSize:        1024,
			EnableBinary:      true,
		},
		Security: SecurityConfig{
			EnableTLS:           false,
			TLSCertFile:         "",
			TLSKeyFile:          "",
			TokenSigningKey:     "", // Should be set in production
			TokenExpiryDuration: 1 * time.Hour,
			AllowedOrigins:      []string{"*"},
			RequireAuth:         true,
			RateLimitRequests:   100,
			RateLimitDuration:   1 * time.Minute,
		},
		Redaction: RedactionConfig{
			EnableRedaction: true,
			DefaultPatterns: []string{
				`(?i)password\s*[:=]\s*\S+`,
				`(?i)token\s*[:=]\s*\S+`,
				`(?i)key\s*[:=]\s*\S+`,
				`(?i)secret\s*[:=]\s*\S+`,
				`(?i)api[_-]?key\s*[:=]\s*\S+`,
				`(?i)auth[_-]?token\s*[:=]\s*\S+`,
				`ssh-[a-z0-9]+\s+[A-Za-z0-9+/=]+`,
				`-----BEGIN [A-Z ]+-----.*-----END [A-Z ]+-----`,
			},
			CustomPatterns:     []string{},
			RedactionChar:      '*',
			CaseSensitive:      false,
			RedactCompleteLine: false,
		},
		Logging: LoggingConfig{
			Level:             "info",
			Format:            "json",
			EnableSessionLogs: true,
			EnableAccessLogs:  true,
			LogFile:           "",
			MaxLogSize:        100, // 100MB
			MaxLogBackups:     3,
			LogRetentionDays:  30,
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return NewValidationError("server.port", c.Server.Port, "must be between 1 and 65535")
	}

	if c.Session.MaxParticipants <= 0 {
		return NewValidationError("session.max_participants", c.Session.MaxParticipants, "must be greater than 0")
	}

	if c.Session.DefaultFrameRate <= 0 || c.Session.DefaultFrameRate > c.Session.MaxFrameRate {
		return NewValidationError("session.default_frame_rate", c.Session.DefaultFrameRate, "must be between 1 and max_frame_rate")
	}

	if c.Transport.CompressionLevel < 1 || c.Transport.CompressionLevel > 9 {
		return NewValidationError("transport.compression_level", c.Transport.CompressionLevel, "must be between 1 and 9")
	}

	if c.Security.EnableTLS {
		if c.Security.TLSCertFile == "" {
			return NewValidationError("security.tls_cert_file", c.Security.TLSCertFile, "required when TLS is enabled")
		}
		if c.Security.TLSKeyFile == "" {
			return NewValidationError("security.tls_key_file", c.Security.TLSKeyFile, "required when TLS is enabled")
		}
	}

	if c.Security.RequireAuth && c.Security.TokenSigningKey == "" {
		return NewValidationError("security.token_signing_key", c.Security.TokenSigningKey, "required when auth is enabled")
	}

	return nil
}

// Address returns the full server address
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Server.Address, c.Server.Port)
}

// IsTLSEnabled returns whether TLS is enabled
func (c *Config) IsTLSEnabled() bool {
	return c.Security.EnableTLS && c.Security.TLSCertFile != "" && c.Security.TLSKeyFile != ""
}