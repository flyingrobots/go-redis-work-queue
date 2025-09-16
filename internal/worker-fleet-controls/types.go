package workerfleetcontrols

import (
	"encoding/json"
	"time"
)

type WorkerState string

const (
	WorkerStateRunning   WorkerState = "running"
	WorkerStatePaused    WorkerState = "paused"
	WorkerStateDraining  WorkerState = "draining"
	WorkerStateStopped   WorkerState = "stopped"
	WorkerStateUnknown   WorkerState = "unknown"
	WorkerStateOffline   WorkerState = "offline"
)

type WorkerAction string

const (
	WorkerActionPause       WorkerAction = "pause"
	WorkerActionResume      WorkerAction = "resume"
	WorkerActionDrain       WorkerAction = "drain"
	WorkerActionStop        WorkerAction = "stop"
	WorkerActionRestart     WorkerAction = "restart"
	WorkerActionKill        WorkerAction = "kill"
	WorkerActionQuarantine  WorkerAction = "quarantine"
)

type Worker struct {
	ID             string                 `json:"id"`
	State          WorkerState            `json:"state"`
	LastHeartbeat  time.Time              `json:"last_heartbeat"`
	StartedAt      time.Time              `json:"started_at"`
	Version        string                 `json:"version"`
	Hostname       string                 `json:"hostname"`
	PID            int                    `json:"pid"`
	CurrentJob     *ActiveJob             `json:"current_job,omitempty"`
	Capabilities   []string               `json:"capabilities"`
	Metadata       map[string]interface{} `json:"metadata"`
	Stats          WorkerStats            `json:"stats"`
	Config         WorkerConfig           `json:"config"`
	Labels         map[string]string      `json:"labels"`
	Health         WorkerHealth           `json:"health"`
}

type ActiveJob struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Queue       string                 `json:"queue"`
	StartedAt   time.Time              `json:"started_at"`
	EstimatedETA *time.Time            `json:"estimated_eta,omitempty"`
	Progress    *JobProgress           `json:"progress,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type JobProgress struct {
	Percentage  float64 `json:"percentage"`
	Stage       string  `json:"stage"`
	Message     string  `json:"message"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkerStats struct {
	JobsProcessed     int64         `json:"jobs_processed"`
	JobsSuccessful    int64         `json:"jobs_successful"`
	JobsFailed        int64         `json:"jobs_failed"`
	TotalRuntime      time.Duration `json:"total_runtime"`
	AverageJobTime    time.Duration `json:"average_job_time"`
	LastJobCompleted  *time.Time    `json:"last_job_completed,omitempty"`
	MemoryUsage       int64         `json:"memory_usage"`
	CPUUsage          float64       `json:"cpu_usage"`
	GoroutineCount    int           `json:"goroutine_count"`
}

type WorkerConfig struct {
	MaxConcurrentJobs int           `json:"max_concurrent_jobs"`
	Queues            []string      `json:"queues"`
	JobTypes          []string      `json:"job_types"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	GracefulTimeout   time.Duration `json:"graceful_timeout"`
	EnableProfiling   bool          `json:"enable_profiling"`
}

type WorkerHealth struct {
	Status         HealthStatus           `json:"status"`
	LastCheck      time.Time              `json:"last_check"`
	Checks         map[string]HealthCheck `json:"checks"`
	ErrorCount     int                    `json:"error_count"`
	RecoveryCount  int                    `json:"recovery_count"`
}

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusCritical  HealthStatus = "critical"
)

type HealthCheck struct {
	Name      string       `json:"name"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message"`
	Timestamp time.Time    `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

type WorkerFilter struct {
	States        []WorkerState         `json:"states,omitempty"`
	Labels        map[string]string     `json:"labels,omitempty"`
	Capabilities  []string              `json:"capabilities,omitempty"`
	HealthStatus  []HealthStatus        `json:"health_status,omitempty"`
	MinHeartbeat  *time.Time            `json:"min_heartbeat,omitempty"`
	MaxHeartbeat  *time.Time            `json:"max_heartbeat,omitempty"`
	HasCurrentJob *bool                 `json:"has_current_job,omitempty"`
	Version       string                `json:"version,omitempty"`
	Hostname      string                `json:"hostname,omitempty"`
}

type WorkerListRequest struct {
	Filter     WorkerFilter `json:"filter"`
	Pagination Pagination   `json:"pagination"`
	SortBy     string       `json:"sort_by"`
	SortOrder  SortOrder    `json:"sort_order"`
}

type WorkerListResponse struct {
	Workers      []Worker   `json:"workers"`
	TotalCount   int        `json:"total_count"`
	Page         int        `json:"page"`
	PageSize     int        `json:"page_size"`
	TotalPages   int        `json:"total_pages"`
	HasNext      bool       `json:"has_next"`
	HasPrevious  bool       `json:"has_previous"`
	Filter       WorkerFilter `json:"filter"`
	Summary      FleetSummary `json:"summary"`
}

type FleetSummary struct {
	TotalWorkers    int                      `json:"total_workers"`
	StateDistribution map[WorkerState]int    `json:"state_distribution"`
	HealthDistribution map[HealthStatus]int  `json:"health_distribution"`
	ActiveJobs      int                      `json:"active_jobs"`
	AverageLoad     float64                  `json:"average_load"`
	UpdatedAt       time.Time                `json:"updated_at"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type WorkerActionRequest struct {
	WorkerIDs     []string              `json:"worker_ids"`
	Action        WorkerAction          `json:"action"`
	Reason        string                `json:"reason,omitempty"`
	Force         bool                  `json:"force,omitempty"`
	DrainTimeout  *time.Duration        `json:"drain_timeout,omitempty"`
	Confirmation  string                `json:"confirmation,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type WorkerActionResponse struct {
	RequestID       string                    `json:"request_id"`
	Action          WorkerAction              `json:"action"`
	TotalRequested  int                       `json:"total_requested"`
	Successful      []string                  `json:"successful"`
	Failed          []WorkerActionError       `json:"failed"`
	InProgress      []string                  `json:"in_progress"`
	StartedAt       time.Time                 `json:"started_at"`
	CompletedAt     *time.Time                `json:"completed_at,omitempty"`
	EstimatedETA    *time.Time                `json:"estimated_eta,omitempty"`
	Status          ActionStatus              `json:"status"`
}

type WorkerActionError struct {
	WorkerID string `json:"worker_id"`
	Error    string `json:"error"`
	Code     string `json:"code"`
}

type ActionStatus string

const (
	ActionStatusPending    ActionStatus = "pending"
	ActionStatusInProgress ActionStatus = "in_progress"
	ActionStatusCompleted  ActionStatus = "completed"
	ActionStatusFailed     ActionStatus = "failed"
	ActionStatusCancelled  ActionStatus = "cancelled"
)

type RollingRestartRequest struct {
	Filter          WorkerFilter   `json:"filter"`
	Concurrency     int            `json:"concurrency"`
	DrainTimeout    time.Duration  `json:"drain_timeout"`
	RestartTimeout  time.Duration  `json:"restart_timeout"`
	MaxUnavailable  int            `json:"max_unavailable"`
	HealthChecks    bool           `json:"health_checks"`
	Confirmation    string         `json:"confirmation"`
}

type RollingRestartResponse struct {
	RequestID       string                    `json:"request_id"`
	TotalWorkers    int                       `json:"total_workers"`
	Phases          []RestartPhase            `json:"phases"`
	CurrentPhase    int                       `json:"current_phase"`
	Status          ActionStatus              `json:"status"`
	StartedAt       time.Time                 `json:"started_at"`
	CompletedAt     *time.Time                `json:"completed_at,omitempty"`
	EstimatedETA    *time.Time                `json:"estimated_eta,omitempty"`
	SuccessCount    int                       `json:"success_count"`
	FailureCount    int                       `json:"failure_count"`
}

type RestartPhase struct {
	PhaseNumber int      `json:"phase_number"`
	WorkerIDs   []string `json:"worker_ids"`
	Status      ActionStatus `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Errors      []WorkerActionError `json:"errors,omitempty"`
}

type AuditLog struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Action      WorkerAction           `json:"action"`
	WorkerIDs   []string               `json:"worker_ids"`
	UserID      string                 `json:"user_id,omitempty"`
	Reason      string                 `json:"reason,omitempty"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
}

type WorkerRegistry interface {
	RegisterWorker(worker *Worker) error
	UpdateWorker(workerID string, updates *Worker) error
	GetWorker(workerID string) (*Worker, error)
	ListWorkers(request WorkerListRequest) (*WorkerListResponse, error)
	RemoveWorker(workerID string) error
	UpdateHeartbeat(workerID string, heartbeat time.Time, currentJob *ActiveJob) error
	GetFleetSummary() (*FleetSummary, error)
	SetWorkerState(workerID string, state WorkerState) error
	GetWorkersByState(state WorkerState) ([]Worker, error)
}

type WorkerController interface {
	PauseWorkers(workerIDs []string, reason string) (*WorkerActionResponse, error)
	ResumeWorkers(workerIDs []string, reason string) (*WorkerActionResponse, error)
	DrainWorkers(workerIDs []string, timeout time.Duration, reason string) (*WorkerActionResponse, error)
	StopWorkers(workerIDs []string, force bool, reason string) (*WorkerActionResponse, error)
	RestartWorkers(workerIDs []string, reason string) (*WorkerActionResponse, error)
	RollingRestart(request RollingRestartRequest) (*RollingRestartResponse, error)
	GetActionStatus(requestID string) (*WorkerActionResponse, error)
	CancelAction(requestID string) error
}

type WorkerSignalHandler interface {
	SendSignal(workerID string, signal WorkerSignal) error
	ReceiveSignals(workerID string) (<-chan WorkerSignal, error)
	CloseSignalChannel(workerID string) error
}

type WorkerSignal struct {
	Type      SignalType             `json:"type"`
	Payload   json.RawMessage        `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type SignalType string

const (
	SignalTypePause       SignalType = "pause"
	SignalTypeResume      SignalType = "resume"
	SignalTypeDrain       SignalType = "drain"
	SignalTypeStop        SignalType = "stop"
	SignalTypeRestart     SignalType = "restart"
	SignalTypeHealthCheck SignalType = "health_check"
	SignalTypeUpdate      SignalType = "update"
)

type AuditLogger interface {
	LogAction(log AuditLog) error
	GetAuditLogs(filter AuditLogFilter) ([]AuditLog, error)
	GetAuditLogsByWorker(workerID string, limit int) ([]AuditLog, error)
	GetAuditLogsByUser(userID string, limit int) ([]AuditLog, error)
}

type AuditLogFilter struct {
	StartTime *time.Time     `json:"start_time,omitempty"`
	EndTime   *time.Time     `json:"end_time,omitempty"`
	Actions   []WorkerAction `json:"actions,omitempty"`
	WorkerIDs []string       `json:"worker_ids,omitempty"`
	UserIDs   []string       `json:"user_ids,omitempty"`
	Success   *bool          `json:"success,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Offset    int            `json:"offset,omitempty"`
}

type SafetyChecker interface {
	ValidateAction(request WorkerActionRequest) error
	CheckFleetHealth(action WorkerAction, workerIDs []string) error
	RequiresConfirmation(action WorkerAction, workerIDs []string) bool
	GenerateConfirmationPrompt(action WorkerAction, workerIDs []string) string
	ValidateConfirmation(action WorkerAction, workerIDs []string, confirmation string) error
}

type Config struct {
	RedisAddr             string        `json:"redis_addr"`
	RedisPassword         string        `json:"redis_password"`
	RedisDB               int           `json:"redis_db"`
	HeartbeatTimeout      time.Duration `json:"heartbeat_timeout"`
	DefaultDrainTimeout   time.Duration `json:"default_drain_timeout"`
	MaxConcurrentActions  int           `json:"max_concurrent_actions"`
	RequireConfirmation   bool          `json:"require_confirmation"`
	SafetyChecksEnabled   bool          `json:"safety_checks_enabled"`
	MinHealthyWorkers     int           `json:"min_healthy_workers"`
	MaxDrainPercentage    float64       `json:"max_drain_percentage"`
	AuditLogRetention     time.Duration `json:"audit_log_retention"`
	EnableMetrics         bool          `json:"enable_metrics"`
	MetricsPrefix         string        `json:"metrics_prefix"`
}

type WorkerFleetManager interface {
	Registry() WorkerRegistry
	Controller() WorkerController
	SignalHandler() WorkerSignalHandler
	AuditLogger() AuditLogger
	SafetyChecker() SafetyChecker
	Start() error
	Stop() error
	Health() error
}