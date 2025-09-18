package chaosharness

import internal "github.com/flyingrobots/go-redis-work-queue/internal/chaos-harness"

// Re-export the chaos harness types and constructors for external consumers.
type (
	ChaosHarness         = internal.ChaosHarness
	Config               = internal.Config
	InjectorType         = internal.InjectorType
	InjectorScope        = internal.InjectorScope
	FaultInjector        = internal.FaultInjector
	ChaosScenario        = internal.ChaosScenario
	ScenarioStage        = internal.ScenarioStage
	LoadConfig           = internal.LoadConfig
	LoadPattern          = internal.LoadPattern
	ScenarioStatus       = internal.ScenarioStatus
	ScenarioMetrics      = internal.ScenarioMetrics
	TimeSeriesPoint      = internal.TimeSeriesPoint
	ScenarioGuardrails   = internal.ScenarioGuardrails
	InjectorConfig       = internal.InjectorConfig
	ChaosReport          = internal.ChaosReport
	ScenarioResult       = internal.ScenarioResult
	Finding              = internal.Finding
	FindingSeverity      = internal.FindingSeverity
	WorkerInjectorState  = internal.WorkerInjectorState
	APIHandler           = internal.APIHandler
	FaultInjectorManager = internal.FaultInjectorManager
	ScenarioRunner       = internal.ScenarioRunner
	LoadGenerator        = internal.LoadGenerator
	MetricsCollector     = internal.MetricsCollector
)

var (
	DefaultConfig           = internal.DefaultConfig
	NewChaosHarness         = internal.NewChaosHarness
	NewFaultInjectorManager = internal.NewFaultInjectorManager
	NewScenarioRunner       = internal.NewScenarioRunner
	NewLoadGenerator        = internal.NewLoadGenerator
	NewAPIHandler           = internal.NewAPIHandler
	NewMetricsCollector     = internal.NewMetricsCollector
)

const (
	InjectorLatency      = internal.InjectorLatency
	InjectorError        = internal.InjectorError
	InjectorPanic        = internal.InjectorPanic
	InjectorPartialFail  = internal.InjectorPartialFail
	InjectorResourceHog  = internal.InjectorResourceHog
	InjectorRedisLatency = internal.InjectorRedisLatency
	InjectorRedisDrop    = internal.InjectorRedisDrop

	ScopeGlobal = internal.ScopeGlobal
	ScopeWorker = internal.ScopeWorker
	ScopeQueue  = internal.ScopeQueue
	ScopeTenant = internal.ScopeTenant

	LoadConstant = internal.LoadConstant
	LoadLinear   = internal.LoadLinear
	LoadSine     = internal.LoadSine
	LoadSpike    = internal.LoadSpike
	LoadRandom   = internal.LoadRandom

	StatusPending   = internal.StatusPending
	StatusRunning   = internal.StatusRunning
	StatusCompleted = internal.StatusCompleted
	StatusFailed    = internal.StatusFailed
	StatusAborted   = internal.StatusAborted

	ResultPassed  = internal.ResultPassed
	ResultFailed  = internal.ResultFailed
	ResultPartial = internal.ResultPartial

	SeverityCritical = internal.SeverityCritical
	SeverityHigh     = internal.SeverityHigh
	SeverityMedium   = internal.SeverityMedium
	SeverityLow      = internal.SeverityLow
	SeverityInfo     = internal.SeverityInfo
)
