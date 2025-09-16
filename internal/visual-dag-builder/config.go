// Copyright 2025 James Ross
package visual_dag_builder

import (
	"time"
)

// Config represents configuration for the Visual DAG Builder
type Config struct {
	Storage     StorageConfig     `json:"storage" yaml:"storage"`
	Execution   ExecutionConfig   `json:"execution" yaml:"execution"`
	UI          UIConfig          `json:"ui" yaml:"ui"`
	API         APIConfig         `json:"api" yaml:"api"`
	Redis       RedisConfig       `json:"redis" yaml:"redis"`
	Observability ObservabilityConfig `json:"observability" yaml:"observability"`
}

// StorageConfig defines where workflows and executions are stored
type StorageConfig struct {
	Type         string        `json:"type" yaml:"type"`                   // redis, memory, database
	Prefix       string        `json:"prefix" yaml:"prefix"`               // key prefix for Redis
	TTL          time.Duration `json:"ttl" yaml:"ttl"`                     // how long to keep executions
	Database     DatabaseConfig `json:"database,omitempty" yaml:"database,omitempty"`
	Compression  bool          `json:"compression" yaml:"compression"`      // compress stored data
}

// DatabaseConfig for SQL database storage
type DatabaseConfig struct {
	Driver     string `json:"driver" yaml:"driver"`         // postgres, mysql, sqlite
	DSN        string `json:"dsn" yaml:"dsn"`               // connection string
	MaxConns   int    `json:"max_conns" yaml:"max_conns"`
	MaxIdle    int    `json:"max_idle" yaml:"max_idle"`
	ConnMaxLife time.Duration `json:"conn_max_life" yaml:"conn_max_life"`
}

// ExecutionConfig defines execution behavior
type ExecutionConfig struct {
	DefaultTimeout       time.Duration `json:"default_timeout" yaml:"default_timeout"`
	MaxConcurrentNodes   int           `json:"max_concurrent_nodes" yaml:"max_concurrent_nodes"`
	HeartbeatInterval    time.Duration `json:"heartbeat_interval" yaml:"heartbeat_interval"`
	RetryPolicy          RetryPolicy   `json:"retry_policy" yaml:"retry_policy"`
	EnableCompensation   bool          `json:"enable_compensation" yaml:"enable_compensation"`
	CompensationTimeout  time.Duration `json:"compensation_timeout" yaml:"compensation_timeout"`
	CleanupInterval      time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
	MaxExecutionHistory  int           `json:"max_execution_history" yaml:"max_execution_history"`
}

// UIConfig defines terminal UI behavior
type UIConfig struct {
	Theme           string        `json:"theme" yaml:"theme"`             // dark, light, auto
	AnimationSpeed  time.Duration `json:"animation_speed" yaml:"animation_speed"`
	RefreshInterval time.Duration `json:"refresh_interval" yaml:"refresh_interval"`
	GridSize        int           `json:"grid_size" yaml:"grid_size"`
	MaxZoom         float64       `json:"max_zoom" yaml:"max_zoom"`
	MinZoom         float64       `json:"min_zoom" yaml:"min_zoom"`
	Colors          ColorConfig   `json:"colors" yaml:"colors"`
	Fonts           FontConfig    `json:"fonts" yaml:"fonts"`
	Shortcuts       map[string]string `json:"shortcuts" yaml:"shortcuts"`
}

// ColorConfig defines colors for different node states
type ColorConfig struct {
	NotStarted   string `json:"not_started" yaml:"not_started"`
	Queued       string `json:"queued" yaml:"queued"`
	Running      string `json:"running" yaml:"running"`
	Completed    string `json:"completed" yaml:"completed"`
	Failed       string `json:"failed" yaml:"failed"`
	Retrying     string `json:"retrying" yaml:"retrying"`
	Compensating string `json:"compensating" yaml:"compensating"`
	Compensated  string `json:"compensated" yaml:"compensated"`
	Selected     string `json:"selected" yaml:"selected"`
	Edge         string `json:"edge" yaml:"edge"`
	Grid         string `json:"grid" yaml:"grid"`
	Background   string `json:"background" yaml:"background"`
	Text         string `json:"text" yaml:"text"`
}

// FontConfig defines font styles
type FontConfig struct {
	NodeName     string `json:"node_name" yaml:"node_name"`
	NodeType     string `json:"node_type" yaml:"node_type"`
	EdgeLabel    string `json:"edge_label" yaml:"edge_label"`
	Inspector    string `json:"inspector" yaml:"inspector"`
	StatusBar    string `json:"status_bar" yaml:"status_bar"`
}

// APIConfig defines API server settings
type APIConfig struct {
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	Host        string        `json:"host" yaml:"host"`
	Port        int           `json:"port" yaml:"port"`
	TLS         TLSConfig     `json:"tls" yaml:"tls"`
	Auth        AuthConfig    `json:"auth" yaml:"auth"`
	CORS        CORSConfig    `json:"cors" yaml:"cors"`
	RateLimit   RateLimitConfig `json:"rate_limit" yaml:"rate_limit"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
}

// TLSConfig for HTTPS
type TLSConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
}

// AuthConfig for API authentication
type AuthConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Type    string `json:"type" yaml:"type"`     // jwt, basic, apikey
	Secret  string `json:"secret" yaml:"secret"`
}

// CORSConfig for cross-origin requests
type CORSConfig struct {
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	AllowedOrigins  []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods  []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders  []string `json:"allowed_headers" yaml:"allowed_headers"`
	AllowCredentials bool    `json:"allow_credentials" yaml:"allow_credentials"`
}

// RateLimitConfig for API rate limiting
type RateLimitConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`
	Requests   int           `json:"requests" yaml:"requests"`
	Window     time.Duration `json:"window" yaml:"window"`
	BurstSize  int           `json:"burst_size" yaml:"burst_size"`
}

// RedisConfig for Redis connection
type RedisConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	Password     string        `json:"password" yaml:"password"`
	Database     int           `json:"database" yaml:"database"`
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
}

// ObservabilityConfig for monitoring and tracing
type ObservabilityConfig struct {
	Metrics MetricsConfig `json:"metrics" yaml:"metrics"`
	Tracing TracingConfig `json:"tracing" yaml:"tracing"`
	Logging LoggingConfig `json:"logging" yaml:"logging"`
}

// MetricsConfig for metrics collection
type MetricsConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Endpoint   string `json:"endpoint" yaml:"endpoint"`
	Namespace  string `json:"namespace" yaml:"namespace"`
	Subsystem  string `json:"subsystem" yaml:"subsystem"`
}

// TracingConfig for distributed tracing
type TracingConfig struct {
	Enabled      bool    `json:"enabled" yaml:"enabled"`
	Endpoint     string  `json:"endpoint" yaml:"endpoint"`
	ServiceName  string  `json:"service_name" yaml:"service_name"`
	SampleRate   float64 `json:"sample_rate" yaml:"sample_rate"`
}

// LoggingConfig for structured logging
type LoggingConfig struct {
	Level      string `json:"level" yaml:"level"`       // debug, info, warn, error
	Format     string `json:"format" yaml:"format"`     // json, text
	Output     string `json:"output" yaml:"output"`     // stdout, stderr, file
	Filename   string `json:"filename" yaml:"filename"`
	MaxSize    int    `json:"max_size" yaml:"max_size"`    // MB
	MaxAge     int    `json:"max_age" yaml:"max_age"`      // days
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	Compress   bool   `json:"compress" yaml:"compress"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Storage: StorageConfig{
			Type:        "redis",
			Prefix:      "dag:",
			TTL:         24 * time.Hour,
			Compression: true,
		},
		Execution: ExecutionConfig{
			DefaultTimeout:       30 * time.Minute,
			MaxConcurrentNodes:   10,
			HeartbeatInterval:    30 * time.Second,
			EnableCompensation:   true,
			CompensationTimeout:  5 * time.Minute,
			CleanupInterval:      1 * time.Hour,
			MaxExecutionHistory:  100,
			RetryPolicy: RetryPolicy{
				Strategy:     "exponential",
				MaxAttempts:  3,
				InitialDelay: 1 * time.Second,
				MaxDelay:     5 * time.Minute,
				Multiplier:   2.0,
				Jitter:       true,
			},
		},
		UI: UIConfig{
			Theme:           "dark",
			AnimationSpeed:  300 * time.Millisecond,
			RefreshInterval: 1 * time.Second,
			GridSize:        20,
			MaxZoom:         3.0,
			MinZoom:         0.25,
			Colors: ColorConfig{
				NotStarted:   "#6c757d",
				Queued:       "#ffc107",
				Running:      "#007bff",
				Completed:    "#28a745",
				Failed:       "#dc3545",
				Retrying:     "#fd7e14",
				Compensating: "#e83e8c",
				Compensated:  "#6f42c1",
				Selected:     "#20c997",
				Edge:         "#495057",
				Grid:         "#343a40",
				Background:   "#212529",
				Text:         "#ffffff",
			},
			Fonts: FontConfig{
				NodeName:     "bold",
				NodeType:     "italic",
				EdgeLabel:    "normal",
				Inspector:    "normal",
				StatusBar:    "normal",
			},
			Shortcuts: map[string]string{
				"move_up":         "k",
				"move_down":       "j",
				"move_left":       "h",
				"move_right":      "l",
				"fast_up":         "K",
				"fast_down":       "J",
				"fast_left":       "H",
				"fast_right":      "L",
				"add_node":        "a",
				"connect":         "c",
				"delete":          "d",
				"cut":             "x",
				"copy":            "y",
				"paste":           "p",
				"undo":            "u",
				"redo":            "ctrl+r",
				"zoom_in":         "z",
				"zoom_out":        "Z",
				"pan_mode":        "space",
				"validate":        "F5",
				"run":             "F6",
				"debug":           "F7",
				"save":            "ctrl+s",
				"load":            "ctrl+o",
				"search":          "/",
				"help":            "?",
				"quit":            "q",
			},
		},
		API: APIConfig{
			Enabled: true,
			Host:    "localhost",
			Port:    8080,
			TLS: TLSConfig{
				Enabled: false,
			},
			Auth: AuthConfig{
				Enabled: false,
				Type:    "jwt",
			},
			CORS: CORSConfig{
				Enabled:         true,
				AllowedOrigins:  []string{"*"},
				AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:  []string{"*"},
				AllowCredentials: false,
			},
			RateLimit: RateLimitConfig{
				Enabled:   true,
				Requests:  100,
				Window:    1 * time.Minute,
				BurstSize: 50,
			},
			Timeout: 30 * time.Second,
		},
		Redis: RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Database:     0,
			PoolSize:     10,
			MinIdleConns: 2,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			IdleTimeout:  5 * time.Minute,
		},
		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled:   true,
				Namespace: "dag_builder",
				Subsystem: "workflows",
			},
			Tracing: TracingConfig{
				Enabled:     true,
				ServiceName: "visual-dag-builder",
				SampleRate:  0.1,
			},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxAge:     30,
				MaxBackups: 3,
				Compress:   true,
			},
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Storage.Type != "redis" && c.Storage.Type != "memory" && c.Storage.Type != "database" {
		return ErrInvalidConfiguration
	}

	if c.Storage.Type == "database" && c.Storage.Database.Driver == "" {
		return ErrInvalidConfiguration
	}

	if c.Execution.MaxConcurrentNodes <= 0 {
		return ErrInvalidConfiguration
	}

	if c.UI.GridSize <= 0 {
		c.UI.GridSize = 20
	}

	if c.UI.MaxZoom <= c.UI.MinZoom {
		return ErrInvalidConfiguration
	}

	if c.API.Port <= 0 || c.API.Port > 65535 {
		return ErrInvalidConfiguration
	}

	return nil
}