package anomalyradarslobudget

import internal "github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget"

// Package anomalyradarslobudget re-exports the anomaly radar API for external consumers.
type (
	Config             = internal.Config
	SLOConfig          = internal.SLOConfig
	BurnRateThresholds = internal.BurnRateThresholds
	AnomalyThresholds  = internal.AnomalyThresholds
	MetricSnapshot     = internal.MetricSnapshot
	RollingWindow      = internal.RollingWindow
	SLOBudget          = internal.SLOBudget
	AlertLevel         = internal.AlertLevel
	AnomalyStatus      = internal.AnomalyStatus
	MetricStatus       = internal.MetricStatus
	Alert              = internal.Alert
	AlertType          = internal.AlertType
	AnomalyRadar       = internal.AnomalyRadar
	MetricsCollector   = internal.MetricsCollector
	AlertCallback      = internal.AlertCallback
	HandlerOption      = internal.HandlerOption
	HTTPHandler        = internal.HTTPHandler
	StatusRequest      = internal.StatusRequest
	StatusResponse     = internal.StatusResponse
	ConfigResponse     = internal.ConfigResponse
	MetricsRequest     = internal.MetricsRequest
	MetricsResponse    = internal.MetricsResponse
	AlertsResponse     = internal.AlertsResponse
	HealthResponse     = internal.HealthResponse
	StartStopResponse  = internal.StartStopResponse
	ErrorResponse      = internal.ErrorResponse
)

var (
	DefaultConfig        = internal.DefaultConfig
	GetRecommendedConfig = internal.GetRecommendedConfig
	ValidateConfig       = internal.ValidateConfig
	New                  = internal.New
	NewRollingWindow     = internal.NewRollingWindow
	ContextWithScopes    = internal.ContextWithScopes
	ScopesFromContext    = internal.ScopesFromContext
	WithScopeChecker     = internal.WithScopeChecker
	WithNow              = internal.WithNow
	NewHTTPHandler       = internal.NewHTTPHandler
)

const (
	ScopeReader = internal.ScopeReader
	ScopeAdmin  = internal.ScopeAdmin

	AlertLevelNone     = internal.AlertLevelNone
	AlertLevelInfo     = internal.AlertLevelInfo
	AlertLevelWarning  = internal.AlertLevelWarning
	AlertLevelCritical = internal.AlertLevelCritical

	MetricStatusHealthy  = internal.MetricStatusHealthy
	MetricStatusWarning  = internal.MetricStatusWarning
	MetricStatusCritical = internal.MetricStatusCritical

	AlertTypeBacklogGrowth = internal.AlertTypeBacklogGrowth
	AlertTypeErrorRate     = internal.AlertTypeErrorRate
	AlertTypeLatency       = internal.AlertTypeLatency
	AlertTypeBurnRate      = internal.AlertTypeBurnRate
)

var (
	ErrAlreadyRunning = internal.ErrAlreadyRunning
	ErrNotRunning     = internal.ErrNotRunning
)
