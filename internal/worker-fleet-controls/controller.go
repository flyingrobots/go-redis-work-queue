package workerfleetcontrols

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type WorkerControllerImpl struct {
	registry      WorkerRegistry
	signalHandler WorkerSignalHandler
	auditLogger   AuditLogger
	safetyChecker SafetyChecker
	config        Config
	logger        *slog.Logger
	activeActions sync.Map // requestID -> *WorkerActionResponse
}

func NewWorkerController(
	registry WorkerRegistry,
	signalHandler WorkerSignalHandler,
	auditLogger AuditLogger,
	safetyChecker SafetyChecker,
	config Config,
	logger *slog.Logger,
) *WorkerControllerImpl {
	return &WorkerControllerImpl{
		registry:      registry,
		signalHandler: signalHandler,
		auditLogger:   auditLogger,
		safetyChecker: safetyChecker,
		config:        config,
		logger:        logger,
	}
}

func (c *WorkerControllerImpl) PauseWorkers(workerIDs []string, reason string) (*WorkerActionResponse, error) {
	request := WorkerActionRequest{
		WorkerIDs: workerIDs,
		Action:    WorkerActionPause,
		Reason:    reason,
	}

	return c.executeAction(request)
}

func (c *WorkerControllerImpl) ResumeWorkers(workerIDs []string, reason string) (*WorkerActionResponse, error) {
	request := WorkerActionRequest{
		WorkerIDs: workerIDs,
		Action:    WorkerActionResume,
		Reason:    reason,
	}

	return c.executeAction(request)
}

func (c *WorkerControllerImpl) DrainWorkers(workerIDs []string, timeout time.Duration, reason string) (*WorkerActionResponse, error) {
	request := WorkerActionRequest{
		WorkerIDs:    workerIDs,
		Action:       WorkerActionDrain,
		Reason:       reason,
		DrainTimeout: &timeout,
	}

	return c.executeAction(request)
}

func (c *WorkerControllerImpl) StopWorkers(workerIDs []string, force bool, reason string) (*WorkerActionResponse, error) {
	request := WorkerActionRequest{
		WorkerIDs: workerIDs,
		Action:    WorkerActionStop,
		Reason:    reason,
		Force:     force,
	}

	return c.executeAction(request)
}

func (c *WorkerControllerImpl) RestartWorkers(workerIDs []string, reason string) (*WorkerActionResponse, error) {
	request := WorkerActionRequest{
		WorkerIDs: workerIDs,
		Action:    WorkerActionRestart,
		Reason:    reason,
	}

	return c.executeAction(request)
}

func (c *WorkerControllerImpl) executeAction(request WorkerActionRequest) (*WorkerActionResponse, error) {
	if err := c.safetyChecker.ValidateAction(request); err != nil {
		return nil, fmt.Errorf("safety check failed: %w", err)
	}

	requestID := uuid.New().String()
	response := &WorkerActionResponse{
		RequestID:      requestID,
		Action:         request.Action,
		TotalRequested: len(request.WorkerIDs),
		Successful:     make([]string, 0),
		Failed:         make([]WorkerActionError, 0),
		InProgress:     make([]string, 0),
		StartedAt:      time.Now(),
		Status:         ActionStatusInProgress,
	}

	c.activeActions.Store(requestID, response)

	go c.processAction(request, response)

	return response, nil
}

func (c *WorkerControllerImpl) processAction(request WorkerActionRequest, response *WorkerActionResponse) {
	c.logger.Info("Processing worker action",
		"request_id", response.RequestID,
		"action", request.Action,
		"worker_count", len(request.WorkerIDs))

	for _, workerID := range request.WorkerIDs {
		response.InProgress = append(response.InProgress, workerID)

		err := c.executeWorkerAction(workerID, request)
		if err != nil {
			response.Failed = append(response.Failed, WorkerActionError{
				WorkerID: workerID,
				Error:    err.Error(),
				Code:     "EXECUTION_FAILED",
			})
			c.logger.Error("Worker action failed", "worker_id", workerID, "action", request.Action, "error", err)
		} else {
			response.Successful = append(response.Successful, workerID)
			c.logger.Info("Worker action successful", "worker_id", workerID, "action", request.Action)
		}

		response.InProgress = removeFromSlice(response.InProgress, workerID)
	}

	response.CompletedAt = &response.StartedAt
	*response.CompletedAt = time.Now()
	response.Status = ActionStatusCompleted

	if len(response.Failed) > 0 && len(response.Successful) == 0 {
		response.Status = ActionStatusFailed
	}

	auditLog := AuditLog{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Action:    request.Action,
		WorkerIDs: request.WorkerIDs,
		Reason:    request.Reason,
		Success:   len(response.Failed) == 0,
		Duration:  response.CompletedAt.Sub(response.StartedAt),
		Metadata: map[string]interface{}{
			"request_id":      response.RequestID,
			"successful":      len(response.Successful),
			"failed":          len(response.Failed),
			"total_requested": response.TotalRequested,
		},
	}

	if len(response.Failed) > 0 {
		auditLog.Error = fmt.Sprintf("%d workers failed", len(response.Failed))
	}

	err := c.auditLogger.LogAction(auditLog)
	if err != nil {
		c.logger.Error("Failed to log audit entry", "error", err)
	}

	c.logger.Info("Worker action completed",
		"request_id", response.RequestID,
		"action", request.Action,
		"successful", len(response.Successful),
		"failed", len(response.Failed),
		"duration", response.CompletedAt.Sub(response.StartedAt))
}

func (c *WorkerControllerImpl) executeWorkerAction(workerID string, request WorkerActionRequest) error {
	worker, err := c.registry.GetWorker(workerID)
	if err != nil {
		return fmt.Errorf("failed to get worker: %w", err)
	}

	switch request.Action {
	case WorkerActionPause:
		return c.pauseWorker(worker)
	case WorkerActionResume:
		return c.resumeWorker(worker)
	case WorkerActionDrain:
		return c.drainWorker(worker, request.DrainTimeout)
	case WorkerActionStop:
		return c.stopWorker(worker, request.Force)
	case WorkerActionRestart:
		return c.restartWorker(worker)
	default:
		return fmt.Errorf("unsupported action: %s", request.Action)
	}
}

func (c *WorkerControllerImpl) pauseWorker(worker *Worker) error {
	if worker.State == WorkerStatePaused {
		return nil // Already paused
	}

	signal := WorkerSignal{
		Type:      SignalTypePause,
		Timestamp: time.Now(),
		Source:    "controller",
	}

	err := c.signalHandler.SendSignal(worker.ID, signal)
	if err != nil {
		return fmt.Errorf("failed to send pause signal: %w", err)
	}

	err = c.registry.SetWorkerState(worker.ID, WorkerStatePaused)
	if err != nil {
		return fmt.Errorf("failed to update worker state: %w", err)
	}

	return nil
}

func (c *WorkerControllerImpl) resumeWorker(worker *Worker) error {
	if worker.State == WorkerStateRunning {
		return nil // Already running
	}

	signal := WorkerSignal{
		Type:      SignalTypeResume,
		Timestamp: time.Now(),
		Source:    "controller",
	}

	err := c.signalHandler.SendSignal(worker.ID, signal)
	if err != nil {
		return fmt.Errorf("failed to send resume signal: %w", err)
	}

	err = c.registry.SetWorkerState(worker.ID, WorkerStateRunning)
	if err != nil {
		return fmt.Errorf("failed to update worker state: %w", err)
	}

	return nil
}

func (c *WorkerControllerImpl) drainWorker(worker *Worker, timeout *time.Duration) error {
	if worker.State == WorkerStateDraining || worker.State == WorkerStateStopped {
		return nil // Already draining or stopped
	}

	drainTimeout := c.config.DefaultDrainTimeout
	if timeout != nil {
		drainTimeout = *timeout
	}

	signalPayload := map[string]interface{}{
		"timeout": drainTimeout.String(),
	}
	payloadBytes, _ := json.Marshal(signalPayload)

	signal := WorkerSignal{
		Type:      SignalTypeDrain,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
		Source:    "controller",
	}

	err := c.signalHandler.SendSignal(worker.ID, signal)
	if err != nil {
		return fmt.Errorf("failed to send drain signal: %w", err)
	}

	err = c.registry.SetWorkerState(worker.ID, WorkerStateDraining)
	if err != nil {
		return fmt.Errorf("failed to update worker state: %w", err)
	}

	return nil
}

func (c *WorkerControllerImpl) stopWorker(worker *Worker, force bool) error {
	if worker.State == WorkerStateStopped {
		return nil // Already stopped
	}

	signalPayload := map[string]interface{}{
		"force": force,
	}
	payloadBytes, _ := json.Marshal(signalPayload)

	signal := WorkerSignal{
		Type:      SignalTypeStop,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
		Source:    "controller",
	}

	err := c.signalHandler.SendSignal(worker.ID, signal)
	if err != nil {
		return fmt.Errorf("failed to send stop signal: %w", err)
	}

	err = c.registry.SetWorkerState(worker.ID, WorkerStateStopped)
	if err != nil {
		return fmt.Errorf("failed to update worker state: %w", err)
	}

	return nil
}

func (c *WorkerControllerImpl) restartWorker(worker *Worker) error {
	signal := WorkerSignal{
		Type:      SignalTypeRestart,
		Timestamp: time.Now(),
		Source:    "controller",
	}

	err := c.signalHandler.SendSignal(worker.ID, signal)
	if err != nil {
		return fmt.Errorf("failed to send restart signal: %w", err)
	}

	return nil
}

func (c *WorkerControllerImpl) RollingRestart(request RollingRestartRequest) (*RollingRestartResponse, error) {
	if err := c.safetyChecker.ValidateConfirmation(WorkerActionRestart, []string{}, request.Confirmation); err != nil {
		return nil, fmt.Errorf("confirmation validation failed: %w", err)
	}

	listRequest := WorkerListRequest{
		Filter: request.Filter,
		Pagination: Pagination{
			Page:     1,
			PageSize: 10000,
		},
	}

	listResponse, err := c.registry.ListWorkers(listRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get workers for rolling restart: %w", err)
	}

	workers := listResponse.Workers

	if len(workers) == 0 {
		return nil, fmt.Errorf("no workers match the filter criteria")
	}

	if request.Concurrency <= 0 {
		request.Concurrency = 1
	}

	requestID := uuid.New().String()
	phases := c.createRestartPhases(workers, request.Concurrency)

	response := &RollingRestartResponse{
		RequestID:    requestID,
		TotalWorkers: len(workers),
		Phases:       phases,
		CurrentPhase: 0,
		Status:       ActionStatusInProgress,
		StartedAt:    time.Now(),
		SuccessCount: 0,
		FailureCount: 0,
	}

	go c.processRollingRestart(request, response)

	return response, nil
}

func (c *WorkerControllerImpl) createRestartPhases(workers []Worker, concurrency int) []RestartPhase {
	phases := make([]RestartPhase, 0)

	for i := 0; i < len(workers); i += concurrency {
		end := i + concurrency
		if end > len(workers) {
			end = len(workers)
		}

		phaseWorkers := workers[i:end]
		workerIDs := make([]string, len(phaseWorkers))
		for j, worker := range phaseWorkers {
			workerIDs[j] = worker.ID
		}

		phase := RestartPhase{
			PhaseNumber: len(phases) + 1,
			WorkerIDs:   workerIDs,
			Status:      ActionStatusPending,
		}

		phases = append(phases, phase)
	}

	return phases
}

func (c *WorkerControllerImpl) processRollingRestart(request RollingRestartRequest, response *RollingRestartResponse) {
	c.logger.Info("Starting rolling restart",
		"request_id", response.RequestID,
		"total_workers", response.TotalWorkers,
		"phases", len(response.Phases),
		"concurrency", request.Concurrency)

	for i := range response.Phases {
		response.CurrentPhase = i
		phase := &response.Phases[i]

		c.logger.Info("Starting restart phase",
			"request_id", response.RequestID,
			"phase", phase.PhaseNumber,
			"workers", len(phase.WorkerIDs))

		now := time.Now()
		phase.StartedAt = &now
		phase.Status = ActionStatusInProgress

		for _, workerID := range phase.WorkerIDs {
			drainRequest := WorkerActionRequest{
				WorkerIDs:    []string{workerID},
				Action:       WorkerActionDrain,
				DrainTimeout: &request.DrainTimeout,
				Reason:       "Rolling restart",
			}

			_, err := c.executeAction(drainRequest)
			if err != nil {
				phase.Errors = append(phase.Errors, WorkerActionError{
					WorkerID: workerID,
					Error:    err.Error(),
					Code:     "DRAIN_FAILED",
				})
				response.FailureCount++
				c.logger.Error("Failed to drain worker in rolling restart",
					"worker_id", workerID, "error", err)
			} else {
				response.SuccessCount++
			}
		}

		now = time.Now()
		phase.CompletedAt = &now
		phase.Status = ActionStatusCompleted

		if len(phase.Errors) > 0 {
			phase.Status = ActionStatusFailed
		}

		if request.HealthChecks && i < len(response.Phases)-1 {
			c.logger.Info("Waiting for health checks before next phase",
				"request_id", response.RequestID,
				"phase", phase.PhaseNumber)
			time.Sleep(30 * time.Second)
		}
	}

	now := time.Now()
	response.CompletedAt = &now
	response.Status = ActionStatusCompleted

	if response.FailureCount > 0 && response.SuccessCount == 0 {
		response.Status = ActionStatusFailed
	}

	c.logger.Info("Rolling restart completed",
		"request_id", response.RequestID,
		"successful", response.SuccessCount,
		"failed", response.FailureCount,
		"duration", response.CompletedAt.Sub(response.StartedAt))
}

func (c *WorkerControllerImpl) GetActionStatus(requestID string) (*WorkerActionResponse, error) {
	if response, ok := c.activeActions.Load(requestID); ok {
		return response.(*WorkerActionResponse), nil
	}
	return nil, fmt.Errorf("action not found: %s", requestID)
}

func (c *WorkerControllerImpl) CancelAction(requestID string) error {
	if response, ok := c.activeActions.Load(requestID); ok {
		actionResponse := response.(*WorkerActionResponse)
		actionResponse.Status = ActionStatusCancelled
		now := time.Now()
		actionResponse.CompletedAt = &now
		return nil
	}
	return fmt.Errorf("action not found: %s", requestID)
}

func removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0)
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

type RedisSignalHandler struct {
	redis  *redis.Client
	config Config
	logger *slog.Logger
}

func NewRedisSignalHandler(redisClient *redis.Client, config Config, logger *slog.Logger) *RedisSignalHandler {
	return &RedisSignalHandler{
		redis:  redisClient,
		config: config,
		logger: logger,
	}
}

func (s *RedisSignalHandler) SendSignal(workerID string, signal WorkerSignal) error {
	ctx := context.Background()

	data, err := json.Marshal(signal)
	if err != nil {
		return fmt.Errorf("failed to marshal signal: %w", err)
	}

	key := fmt.Sprintf("worker:signals:%s", workerID)
	err = s.redis.LPush(ctx, key, data).Err()
	if err != nil {
		return fmt.Errorf("failed to send signal: %w", err)
	}

	err = s.redis.Expire(ctx, key, 24*time.Hour).Err()
	if err != nil {
		s.logger.Warn("Failed to set signal expiration", "key", key, "error", err)
	}

	s.logger.Debug("Signal sent", "worker_id", workerID, "signal_type", signal.Type)
	return nil
}

func (s *RedisSignalHandler) ReceiveSignals(workerID string) (<-chan WorkerSignal, error) {
	signals := make(chan WorkerSignal, 10)

	go func() {
		defer close(signals)

		ctx := context.Background()
		key := fmt.Sprintf("worker:signals:%s", workerID)

		for {
			result, err := s.redis.BRPop(ctx, 5*time.Second, key).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				if errors.Is(err, redis.ErrClosed) || errors.Is(err, context.Canceled) {
					s.logger.Debug("Signal receiver stopping", "worker_id", workerID, "reason", err)
					return
				}
				s.logger.Error("Failed to receive signal", "worker_id", workerID, "error", err)
				continue
			}

			if len(result) < 2 {
				continue
			}

			var signal WorkerSignal
			err = json.Unmarshal([]byte(result[1]), &signal)
			if err != nil {
				s.logger.Error("Failed to unmarshal signal", "worker_id", workerID, "error", err)
				continue
			}

			select {
			case signals <- signal:
				s.logger.Debug("Signal received", "worker_id", workerID, "signal_type", signal.Type)
			default:
				s.logger.Warn("Signal channel full, dropping signal", "worker_id", workerID, "signal_type", signal.Type)
			}
		}
	}()

	return signals, nil
}

func (s *RedisSignalHandler) CloseSignalChannel(workerID string) error {
	// Signal channels are closed automatically when the worker stops listening
	s.logger.Debug("Signal channel closed", "worker_id", workerID)
	return nil
}
