package workerfleetcontrols

import (
	"fmt"
	"log/slog"
	"strings"
)

type SafetyCheckerImpl struct {
	registry WorkerRegistry
	config   Config
	logger   *slog.Logger
}

func NewSafetyChecker(registry WorkerRegistry, config Config, logger *slog.Logger) *SafetyCheckerImpl {
	return &SafetyCheckerImpl{
		registry: registry,
		config:   config,
		logger:   logger,
	}
}

func (s *SafetyCheckerImpl) ValidateAction(request WorkerActionRequest) error {
	if !s.config.SafetyChecksEnabled {
		return nil
	}

	if len(request.WorkerIDs) == 0 {
		return fmt.Errorf("no workers specified")
	}

	summary, err := s.registry.GetFleetSummary()
	if err != nil {
		return fmt.Errorf("failed to get fleet summary: %w", err)
	}

	switch request.Action {
	case WorkerActionDrain, WorkerActionStop:
		return s.validateDrainStopAction(request, summary)
	case WorkerActionPause:
		return s.validatePauseAction(request, summary)
	default:
		return nil // Other actions are generally safe
	}
}

func (s *SafetyCheckerImpl) validateDrainStopAction(request WorkerActionRequest, summary *FleetSummary) error {
	if summary.TotalWorkers == 0 {
		return fmt.Errorf("no workers in fleet")
	}

	runningWorkers := summary.StateDistribution[WorkerStateRunning]
	healthyWorkers := summary.HealthDistribution[HealthStatusHealthy]

	affectedWorkers := len(request.WorkerIDs)
	remainingHealthy := healthyWorkers - affectedWorkers

	if remainingHealthy < s.config.MinHealthyWorkers {
		return fmt.Errorf("action would leave only %d healthy workers, minimum required: %d",
			remainingHealthy, s.config.MinHealthyWorkers)
	}

	drainPercentage := float64(affectedWorkers) / float64(summary.TotalWorkers) * 100
	if drainPercentage > s.config.MaxDrainPercentage {
		return fmt.Errorf("action would affect %.1f%% of workers, maximum allowed: %.1f%%",
			drainPercentage, s.config.MaxDrainPercentage)
	}

	if affectedWorkers == runningWorkers && !request.Force {
		return fmt.Errorf("action would stop all running workers, use force=true to override")
	}

	return nil
}

func (s *SafetyCheckerImpl) validatePauseAction(request WorkerActionRequest, summary *FleetSummary) error {
	runningWorkers := summary.StateDistribution[WorkerStateRunning]
	affectedWorkers := len(request.WorkerIDs)

	if affectedWorkers >= runningWorkers && !request.Force {
		return fmt.Errorf("action would pause all running workers, use force=true to override")
	}

	return nil
}

func (s *SafetyCheckerImpl) CheckFleetHealth(action WorkerAction, workerIDs []string) error {
	summary, err := s.registry.GetFleetSummary()
	if err != nil {
		return fmt.Errorf("failed to get fleet summary: %w", err)
	}

	unhealthyCount := summary.HealthDistribution[HealthStatusUnhealthy] +
		summary.HealthDistribution[HealthStatusCritical]

	if unhealthyCount > len(workerIDs) {
		s.logger.Warn("Fleet has unhealthy workers",
			"unhealthy_count", unhealthyCount,
			"total_workers", summary.TotalWorkers,
			"action", action)
	}

	offlineCount := summary.StateDistribution[WorkerStateOffline]
	if offlineCount > 0 {
		s.logger.Warn("Fleet has offline workers",
			"offline_count", offlineCount,
			"total_workers", summary.TotalWorkers,
			"action", action)
	}

	return nil
}

func (s *SafetyCheckerImpl) RequiresConfirmation(action WorkerAction, workerIDs []string) bool {
	if !s.config.RequireConfirmation {
		return false
	}

	summary, err := s.registry.GetFleetSummary()
	if err != nil {
		s.logger.Error("Failed to get fleet summary for confirmation check", "error", err)
		return true // Err on the side of caution
	}

	affectedWorkers := len(workerIDs)
	totalWorkers := summary.TotalWorkers

	switch action {
	case WorkerActionDrain, WorkerActionStop:
		drainPercentage := float64(affectedWorkers) / float64(totalWorkers) * 100
		return drainPercentage >= 25.0 || affectedWorkers >= 5
	case WorkerActionPause:
		pausePercentage := float64(affectedWorkers) / float64(totalWorkers) * 100
		return pausePercentage >= 50.0 || affectedWorkers >= 10
	case WorkerActionRestart:
		return affectedWorkers >= 3
	default:
		return false
	}
}

func (s *SafetyCheckerImpl) GenerateConfirmationPrompt(action WorkerAction, workerIDs []string) string {
	summary, err := s.registry.GetFleetSummary()
	if err != nil {
		return fmt.Sprintf("Type 'CONFIRM' to proceed with %s action on %d workers", action, len(workerIDs))
	}

	affectedWorkers := len(workerIDs)
	totalWorkers := summary.TotalWorkers
	percentage := float64(affectedWorkers) / float64(totalWorkers) * 100

	var impact string
	switch action {
	case WorkerActionDrain, WorkerActionStop:
		activeJobs := summary.ActiveJobs
		impact = fmt.Sprintf("This will affect %d workers (%.1f%% of fleet) and may impact %d active jobs.",
			affectedWorkers, percentage, activeJobs)
	case WorkerActionPause:
		impact = fmt.Sprintf("This will pause %d workers (%.1f%% of fleet).",
			affectedWorkers, percentage)
	case WorkerActionRestart:
		impact = fmt.Sprintf("This will restart %d workers (%.1f%% of fleet).",
			affectedWorkers, percentage)
	default:
		impact = fmt.Sprintf("This will affect %d workers (%.1f%% of fleet).",
			affectedWorkers, percentage)
	}

	return fmt.Sprintf("%s Type 'CONFIRM' to proceed.", impact)
}

func (s *SafetyCheckerImpl) ValidateConfirmation(action WorkerAction, workerIDs []string, confirmation string) error {
	if !s.RequiresConfirmation(action, workerIDs) {
		return nil
	}

	confirmation = strings.TrimSpace(strings.ToUpper(confirmation))
	if confirmation != "CONFIRM" {
		return fmt.Errorf("invalid confirmation, expected 'CONFIRM', got '%s'", confirmation)
	}

	return nil
}