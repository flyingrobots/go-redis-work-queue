// Copyright 2025 James Ross
package backpressure

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Priority levels for job processing
type Priority int

const (
	LowPriority Priority = iota
	MediumPriority
	HighPriority
)

func (p Priority) String() string {
	switch p {
	case HighPriority:
		return "high"
	case MediumPriority:
		return "medium"
	case LowPriority:
		return "low"
	default:
		return "unknown"
	}
}

// InfiniteDelay represents a delay that effectively sheds the job
const InfiniteDelay = time.Duration(1<<63 - 1)

// BacklogWindow defines backpressure thresholds for different load levels
type BacklogWindow struct {
	Green  int `json:"green_max"`  // 0-Green: no throttling
	Yellow int `json:"yellow_max"` // Green-Yellow: light throttling
	Red    int `json:"red_max"`    // Yellow-Red: heavy throttling/shedding
}

// BacklogThresholds defines thresholds for each priority level
type BacklogThresholds struct {
	HighPriority   BacklogWindow `json:"high_priority"`
	MediumPriority BacklogWindow `json:"medium_priority"`
	LowPriority    BacklogWindow `json:"low_priority"`
}

// DefaultThresholds returns sensible default backlog thresholds
func DefaultThresholds() BacklogThresholds {
	return BacklogThresholds{
		HighPriority: BacklogWindow{
			Green:  1000, // High priority gets more headroom
			Yellow: 5000,
			Red:    10000,
		},
		MediumPriority: BacklogWindow{
			Green:  500,
			Yellow: 2000,
			Red:    5000,
		},
		LowPriority: BacklogWindow{
			Green:  100, // Low priority throttles early
			Yellow: 500,
			Red:    1000,
		},
	}
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

func (cs CircuitState) String() string {
	switch cs {
	case Closed:
		return "closed"
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitConfig configures circuit breaker behavior
type CircuitConfig struct {
	FailureThreshold  int           `json:"failure_threshold"`  // Trip after N failures
	RecoveryThreshold int           `json:"recovery_threshold"` // Close after N successes
	TripWindow        time.Duration `json:"trip_window"`        // Time window for failure counting
	RecoveryTimeout   time.Duration `json:"recovery_timeout"`   // Wait before half-open
	ProbeInterval     time.Duration `json:"probe_interval"`     // Half-open probe frequency
}

// DefaultCircuitConfig returns sensible circuit breaker defaults
func DefaultCircuitConfig() CircuitConfig {
	return CircuitConfig{
		FailureThreshold:  5,
		RecoveryThreshold: 3,
		TripWindow:        30 * time.Second,
		RecoveryTimeout:   60 * time.Second,
		ProbeInterval:     5 * time.Second,
	}
}

// CircuitBreaker provides circuit breaker functionality for backpressure
type CircuitBreaker struct {
	State           CircuitState  `json:"state"`
	FailureCount    int           `json:"failure_count"`
	LastFailureTime time.Time     `json:"last_failure_time"`
	SuccessCount    int           `json:"success_count"`
	LastProbe       time.Time     `json:"last_probe"`
	Config          CircuitConfig `json:"config"`
	mu              sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker with given config
func NewCircuitBreaker(config CircuitConfig) *CircuitBreaker {
	return &CircuitBreaker{
		State:  Closed,
		Config: config,
	}
}

// PollingConfig configures backpressure polling behavior
type PollingConfig struct {
	Interval     time.Duration `json:"interval"`      // Base polling interval
	Jitter       time.Duration `json:"jitter"`        // Jitter to prevent thundering herd
	Timeout      time.Duration `json:"timeout"`       // API call timeout
	MaxBackoff   time.Duration `json:"max_backoff"`   // Maximum backoff on failures
	CacheTTL     time.Duration `json:"cache_ttl"`     // How long to cache throttle decisions
	Enabled      bool          `json:"enabled"`       // Enable/disable polling
}

// DefaultPollingConfig returns sensible polling defaults
func DefaultPollingConfig() PollingConfig {
	return PollingConfig{
		Interval:   5 * time.Second,
		Jitter:     1 * time.Second,
		Timeout:    3 * time.Second,
		MaxBackoff: 60 * time.Second,
		CacheTTL:   30 * time.Second,
		Enabled:    true,
	}
}

// QueueStats represents current queue backlog statistics
type QueueStats struct {
	QueueName       string    `json:"queue_name"`
	BacklogCount    int       `json:"backlog_count"`
	ProcessingCount int       `json:"processing_count"`
	LastUpdated     time.Time `json:"last_updated"`
	RateLimit       struct {
		Budget    int `json:"budget"`
		Remaining int `json:"remaining"`
	} `json:"rate_limit"`
}

// BackpressureConfig configures the overall backpressure system
type BackpressureConfig struct {
	Thresholds BacklogThresholds `json:"thresholds"`
	Circuit    CircuitConfig     `json:"circuit"`
	Polling    PollingConfig     `json:"polling"`
	Recovery   RecoveryStrategy  `json:"recovery"`
}

// DefaultConfig returns a complete default configuration
func DefaultConfig() BackpressureConfig {
	return BackpressureConfig{
		Thresholds: DefaultThresholds(),
		Circuit:    DefaultCircuitConfig(),
		Polling:    DefaultPollingConfig(),
		Recovery:   DefaultRecoveryStrategy(),
	}
}

// RecoveryStrategy configures failure recovery behavior
type RecoveryStrategy struct {
	FallbackMode    bool          `json:"fallback_mode"`    // Use cached values when API unavailable
	GracefulDegrade time.Duration `json:"graceful_degrade"` // Gradually relax throttling during outages
	ManualOverride  bool          `json:"manual_override"`  // Allow ops to disable backpressure
	EmergencyMode   bool          `json:"emergency_mode"`   // Disable all throttling in emergencies
}

// DefaultRecoveryStrategy returns sensible recovery defaults
func DefaultRecoveryStrategy() RecoveryStrategy {
	return RecoveryStrategy{
		FallbackMode:    true,
		GracefulDegrade: 5 * time.Minute,
		ManualOverride:  false,
		EmergencyMode:   false,
	}
}

// ThrottleDecision represents a backpressure decision
type ThrottleDecision struct {
	Priority    Priority      `json:"priority"`
	QueueName   string        `json:"queue_name"`
	Delay       time.Duration `json:"delay"`
	ShouldShed  bool          `json:"should_shed"`
	Reason      string        `json:"reason"`
	Timestamp   time.Time     `json:"timestamp"`
	BacklogSize int           `json:"backlog_size"`
}

// BackpressureMetrics holds prometheus metrics for backpressure operations
type BackpressureMetrics struct {
	ThrottleEventsTotal    *prometheus.CounterVec
	ShedEventsTotal        *prometheus.CounterVec
	ThrottleDelayHistogram *prometheus.HistogramVec
	CircuitBreakerState    *prometheus.GaugeVec
	QueueBacklogGauge      *prometheus.GaugeVec
	ProducerCompliance     *prometheus.GaugeVec
	PollingErrors          *prometheus.CounterVec
	CacheHitRate           *prometheus.GaugeVec
}

// NewBackpressureMetrics creates and registers backpressure metrics
func NewBackpressureMetrics() *BackpressureMetrics {
	return &BackpressureMetrics{
		ThrottleEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "backpressure_throttle_events_total",
				Help: "Total number of throttle events by priority and queue",
			},
			[]string{"priority", "queue"},
		),
		ShedEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "backpressure_shed_events_total",
				Help: "Total number of job shed events by priority and queue",
			},
			[]string{"priority", "queue"},
		),
		ThrottleDelayHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "backpressure_throttle_delay_seconds",
				Help: "Distribution of throttle delay durations",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"priority", "queue"},
		),
		CircuitBreakerState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "backpressure_circuit_breaker_state",
				Help: "Current circuit breaker state (0=closed, 1=open, 2=half-open)",
			},
			[]string{"queue"},
		),
		QueueBacklogGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "backpressure_queue_backlog_size",
				Help: "Current queue backlog size",
			},
			[]string{"queue"},
		),
		ProducerCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "backpressure_producer_compliance_ratio",
				Help: "Percentage of producers complying with throttle recommendations",
			},
			[]string{"queue"},
		),
		PollingErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "backpressure_polling_errors_total",
				Help: "Total number of errors polling queue statistics",
			},
			[]string{"error_type"},
		),
		CacheHitRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "backpressure_cache_hit_rate",
				Help: "Cache hit rate for throttle decisions",
			},
			[]string{},
		),
	}
}

// Register registers all metrics with the default prometheus registry
func (m *BackpressureMetrics) Register() {
	prometheus.MustRegister(
		m.ThrottleEventsTotal,
		m.ShedEventsTotal,
		m.ThrottleDelayHistogram,
		m.CircuitBreakerState,
		m.QueueBacklogGauge,
		m.ProducerCompliance,
		m.PollingErrors,
		m.CacheHitRate,
	)
}

// StatsProvider defines the interface for getting queue statistics
type StatsProvider interface {
	GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error)
	GetAllQueueStats(ctx context.Context) (map[string]*QueueStats, error)
}

// BackpressureController is the main interface for backpressure operations
type BackpressureController interface {
	// SuggestThrottle returns recommended delay for given priority and queue
	SuggestThrottle(ctx context.Context, priority Priority, queueName string) (*ThrottleDecision, error)

	// Run executes work function with automatic throttling
	Run(ctx context.Context, priority Priority, queueName string, work func() error) error

	// ProcessBatch processes multiple jobs with backpressure awareness
	ProcessBatch(ctx context.Context, jobs []BatchJob) error

	// GetCircuitState returns current circuit breaker state for queue
	GetCircuitState(queueName string) CircuitState

	// SetManualOverride enables/disables manual override mode
	SetManualOverride(enabled bool)

	// Start begins background polling and maintenance
	Start(ctx context.Context) error

	// Stop shuts down the controller gracefully
	Stop() error

	// Health returns controller health status
	Health() map[string]interface{}
}

// BatchJob represents a job in batch processing
type BatchJob struct {
	Priority  Priority    `json:"priority"`
	QueueName string      `json:"queue_name"`
	Payload   interface{} `json:"payload"`
	Work      func() error
}