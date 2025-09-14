// Copyright 2025 James Ross
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	exactlyonce "github.com/james-ross/go-redis-work-queue/internal/exactly-once-patterns"
)

type Redis struct {
	Addr               string        `mapstructure:"addr"`
	Username           string        `mapstructure:"username"`
	Password           string        `mapstructure:"password"`
	DB                 int           `mapstructure:"db"`
	PoolSizeMultiplier int           `mapstructure:"pool_size_multiplier"`
	MinIdleConns       int           `mapstructure:"min_idle_conns"`
	DialTimeout        time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout"`
	MaxRetries         int           `mapstructure:"max_retries"`
}

type Backoff struct {
	Base time.Duration `mapstructure:"base"`
	Max  time.Duration `mapstructure:"max"`
}

type Worker struct {
	Count                 int               `mapstructure:"count"`
	HeartbeatTTL          time.Duration     `mapstructure:"heartbeat_ttl"`
	MaxRetries            int               `mapstructure:"max_retries"`
	Backoff               Backoff           `mapstructure:"backoff"`
	Priorities            []string          `mapstructure:"priorities"`
	Queues                map[string]string `mapstructure:"queues"`
	ProcessingListPattern string            `mapstructure:"processing_list_pattern"`
	HeartbeatKeyPattern   string            `mapstructure:"heartbeat_key_pattern"`
	CompletedList         string            `mapstructure:"completed_list"`
	DeadLetterList        string            `mapstructure:"dead_letter_list"`
	BRPopLPushTimeout     time.Duration     `mapstructure:"brpoplpush_timeout"`
	BreakerPause          time.Duration     `mapstructure:"breaker_pause"`
}

type Producer struct {
	ScanDir          string   `mapstructure:"scan_dir"`
	IncludeGlobs     []string `mapstructure:"include_globs"`
	ExcludeGlobs     []string `mapstructure:"exclude_globs"`
	DefaultPriority  string   `mapstructure:"default_priority"`
	HighPriorityExts []string `mapstructure:"high_priority_exts"`
	RateLimitPerSec  int      `mapstructure:"rate_limit_per_sec"`
	RateLimitKey     string   `mapstructure:"rate_limit_key"`
}

type CircuitBreaker struct {
	FailureThreshold float64       `mapstructure:"failure_threshold"`
	Window           time.Duration `mapstructure:"window"`
	CooldownPeriod   time.Duration `mapstructure:"cooldown_period"`
	MinSamples       int           `mapstructure:"min_samples"`
}

type TracingConfig struct {
	Enabled             bool              `mapstructure:"enabled"`
	Endpoint            string            `mapstructure:"endpoint"`
	Environment         string            `mapstructure:"environment"`
	SamplingStrategy    string            `mapstructure:"sampling_strategy"`
	SamplingRate        float64           `mapstructure:"sampling_rate"`
	BatchTimeout        time.Duration     `mapstructure:"batch_timeout"`
	MaxExportBatchSize  int               `mapstructure:"max_export_batch_size"`
	Headers             map[string]string `mapstructure:"headers"`
	Insecure            bool              `mapstructure:"insecure"`
	PropagationFormat   string            `mapstructure:"propagation_format"`
	AttributeAllowlist  []string          `mapstructure:"attribute_allowlist"`
	RedactSensitive     bool              `mapstructure:"redact_sensitive"`
	EnableMetricExemplars bool            `mapstructure:"enable_metric_exemplars"`
}

// Tracing is a backwards-compatible alias
type Tracing = TracingConfig

type ObservabilityConfig struct {
	MetricsPort         int           `mapstructure:"metrics_port"`
	LogLevel            string        `mapstructure:"log_level"`
	Tracing             TracingConfig `mapstructure:"tracing"`
	QueueSampleInterval time.Duration `mapstructure:"queue_sample_interval"`
}

// Observability is a backwards-compatible alias
type Observability = ObservabilityConfig

type Config struct {
	Redis          Redis               `mapstructure:"redis"`
	Worker         Worker              `mapstructure:"worker"`
	Producer       Producer            `mapstructure:"producer"`
	CircuitBreaker CircuitBreaker      `mapstructure:"circuit_breaker"`
	Observability  Observability       `mapstructure:"observability"`
	ExactlyOnce    exactlyonce.Config  `mapstructure:"exactly_once"`
}

func defaultConfig() *Config {
	return &Config{
		Redis: Redis{
			Addr:               "localhost:6379",
			PoolSizeMultiplier: 10,
			MinIdleConns:       5,
			DialTimeout:        5 * time.Second,
			ReadTimeout:        3 * time.Second,
			WriteTimeout:       3 * time.Second,
			MaxRetries:         3,
		},
		Worker: Worker{
			Count:                 16,
			HeartbeatTTL:          30 * time.Second,
			MaxRetries:            3,
			Backoff:               Backoff{Base: 500 * time.Millisecond, Max: 10 * time.Second},
			Priorities:            []string{"high", "low"},
			Queues:                map[string]string{"high": "jobqueue:high_priority", "low": "jobqueue:low_priority"},
			ProcessingListPattern: "jobqueue:worker:%s:processing",
			HeartbeatKeyPattern:   "jobqueue:processing:worker:%s",
			CompletedList:         "jobqueue:completed",
			DeadLetterList:        "jobqueue:dead_letter",
			BRPopLPushTimeout:     1 * time.Second,
			BreakerPause:          100 * time.Millisecond,
		},
		Producer: Producer{
			ScanDir:          "./data",
			IncludeGlobs:     []string{"**/*"},
			ExcludeGlobs:     []string{"**/*.tmp", "**/.DS_Store"},
			DefaultPriority:  "low",
			HighPriorityExts: []string{".pdf", ".docx", ".xlsx", ".zip"},
			RateLimitPerSec:  100,
			RateLimitKey:     "jobqueue:rate_limit:producer",
		},
		CircuitBreaker: CircuitBreaker{
			FailureThreshold: 0.5,
			Window:           1 * time.Minute,
			CooldownPeriod:   30 * time.Second,
			MinSamples:       20,
		},
		Observability: Observability{
			MetricsPort:         9090,
			LogLevel:            "info",
			Tracing:             Tracing{Enabled: false},
			QueueSampleInterval: 2 * time.Second,
		},
	}
}

// Load reads configuration from YAML file and env overrides.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	def := defaultConfig()
	v.SetDefault("redis.addr", def.Redis.Addr)
	v.SetDefault("redis.pool_size_multiplier", def.Redis.PoolSizeMultiplier)
	v.SetDefault("redis.min_idle_conns", def.Redis.MinIdleConns)
	v.SetDefault("redis.dial_timeout", def.Redis.DialTimeout)
	v.SetDefault("redis.read_timeout", def.Redis.ReadTimeout)
	v.SetDefault("redis.write_timeout", def.Redis.WriteTimeout)
	v.SetDefault("redis.max_retries", def.Redis.MaxRetries)

	v.SetDefault("worker.count", def.Worker.Count)
	v.SetDefault("worker.heartbeat_ttl", def.Worker.HeartbeatTTL)
	v.SetDefault("worker.max_retries", def.Worker.MaxRetries)
	v.SetDefault("worker.backoff.base", def.Worker.Backoff.Base)
	v.SetDefault("worker.backoff.max", def.Worker.Backoff.Max)
	v.SetDefault("worker.priorities", def.Worker.Priorities)
	v.SetDefault("worker.queues", def.Worker.Queues)
	v.SetDefault("worker.processing_list_pattern", def.Worker.ProcessingListPattern)
	v.SetDefault("worker.heartbeat_key_pattern", def.Worker.HeartbeatKeyPattern)
	v.SetDefault("worker.completed_list", def.Worker.CompletedList)
	v.SetDefault("worker.dead_letter_list", def.Worker.DeadLetterList)
	v.SetDefault("worker.brpoplpush_timeout", def.Worker.BRPopLPushTimeout)
	v.SetDefault("worker.breaker_pause", def.Worker.BreakerPause)

	v.SetDefault("producer.scan_dir", def.Producer.ScanDir)
	v.SetDefault("producer.include_globs", def.Producer.IncludeGlobs)
	v.SetDefault("producer.exclude_globs", def.Producer.ExcludeGlobs)
	v.SetDefault("producer.default_priority", def.Producer.DefaultPriority)
	v.SetDefault("producer.high_priority_exts", def.Producer.HighPriorityExts)
	v.SetDefault("producer.rate_limit_per_sec", def.Producer.RateLimitPerSec)
	v.SetDefault("producer.rate_limit_key", def.Producer.RateLimitKey)

	v.SetDefault("circuit_breaker.failure_threshold", def.CircuitBreaker.FailureThreshold)
	v.SetDefault("circuit_breaker.window", def.CircuitBreaker.Window)
	v.SetDefault("circuit_breaker.cooldown_period", def.CircuitBreaker.CooldownPeriod)
	v.SetDefault("circuit_breaker.min_samples", def.CircuitBreaker.MinSamples)

	v.SetDefault("observability.metrics_port", def.Observability.MetricsPort)
	v.SetDefault("observability.log_level", def.Observability.LogLevel)
	v.SetDefault("observability.tracing.enabled", def.Observability.Tracing.Enabled)
	v.SetDefault("observability.tracing.endpoint", def.Observability.Tracing.Endpoint)
	v.SetDefault("observability.queue_sample_interval", def.Observability.QueueSampleInterval)

	// Optional file read
	if _, err := os.Stat(path); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := Validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks config constraints and returns an error on invalid settings.
func Validate(cfg *Config) error {
	if cfg.Worker.Count < 1 {
		return fmt.Errorf("worker.count must be >= 1")
	}
	if len(cfg.Worker.Priorities) == 0 {
		return fmt.Errorf("worker.priorities must be non-empty")
	}
	for _, p := range cfg.Worker.Priorities {
		if _, ok := cfg.Worker.Queues[p]; !ok {
			return fmt.Errorf("worker.queues missing entry for priority %q", p)
		}
	}
	if cfg.Worker.HeartbeatTTL < 5*time.Second {
		return fmt.Errorf("worker.heartbeat_ttl must be >= 5s")
	}
	if cfg.Worker.BRPopLPushTimeout <= 0 || cfg.Worker.BRPopLPushTimeout > cfg.Worker.HeartbeatTTL/2 {
		return fmt.Errorf("worker.brpoplpush_timeout must be >0 and <= heartbeat_ttl/2")
	}
	if cfg.Producer.RateLimitPerSec < 0 {
		return fmt.Errorf("producer.rate_limit_per_sec must be >= 0")
	}
	if cfg.Observability.MetricsPort <= 0 || cfg.Observability.MetricsPort > 65535 {
		return fmt.Errorf("observability.metrics_port must be 1..65535")
	}
	return nil
}
