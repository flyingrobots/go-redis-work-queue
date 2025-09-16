package budgeting

import (
	"fmt"
	"time"
)

// BudgetEnforcer handles budget enforcement and graduated responses
type BudgetEnforcer struct {
	budgetManager *BudgetManager
	notifier      *NotificationService
	alertTracker  map[string]time.Time // Track last alert times
}

// NewBudgetEnforcer creates a new budget enforcer
func NewBudgetEnforcer(budgetManager *BudgetManager, notifier *NotificationService) *BudgetEnforcer {
	return &BudgetEnforcer{
		budgetManager: budgetManager,
		notifier:      notifier,
		alertTracker:  make(map[string]time.Time),
	}
}

// CheckBudgetCompliance checks budget compliance and returns enforcement action
func (e *BudgetEnforcer) CheckBudgetCompliance(tenantID, queueName string) (EnforcementAction, error) {
	budget, err := e.budgetManager.GetBudgetForTenant(tenantID, queueName)
	if err != nil {
		if err == ErrBudgetNotFound {
			// No budget configured, allow all operations
			return EnforcementAction{Type: "allow"}, nil
		}
		return EnforcementAction{}, err
	}

	// Check if budget is active and within period
	now := time.Now()
	if !budget.Active || now.Before(budget.Period.StartDate) || now.After(budget.Period.EndDate) {
		return EnforcementAction{Type: "allow"}, nil
	}

	currentSpend, err := e.budgetManager.aggregator.GetCurrentSpend(tenantID, queueName, budget.Period)
	if err != nil {
		return EnforcementAction{}, err
	}

	utilization := currentSpend / budget.Amount

	// Check grace period for new budgets
	if budget.EnforcementPolicy.GracePeriodHours > 0 {
		graceEnd := budget.CreatedAt.Add(time.Duration(budget.EnforcementPolicy.GracePeriodHours) * time.Hour)
		if now.Before(graceEnd) {
			return EnforcementAction{
				Type:    "allow",
				Message: fmt.Sprintf("Grace period active until %s", graceEnd.Format("2006-01-02 15:04")),
			}, nil
		}
	}

	// Determine enforcement action based on utilization
	alertKey := fmt.Sprintf("%s:%s", tenantID, queueName)

	switch {
	case utilization >= budget.BlockThreshold && budget.EnforcementPolicy.BlockNewJobs && !budget.EnforcementPolicy.WarnOnly:
		e.sendAlert(tenantID, queueName, "budget_exceeded", currentSpend, budget.Amount, alertKey)
		return EnforcementAction{
			Type:    "block",
			Message: fmt.Sprintf("Budget exceeded: $%.2f/$%.2f (%.1f%%)", currentSpend, budget.Amount, utilization*100),
		}, nil

	case utilization >= budget.ThrottleThreshold && !budget.EnforcementPolicy.WarnOnly:
		throttleFactor := budget.EnforcementPolicy.ThrottleFactor
		if throttleFactor == 0 {
			throttleFactor = 0.5 // Default to 50% throttling
		}

		e.sendAlert(tenantID, queueName, "budget_throttle", currentSpend, budget.Amount, alertKey)
		return EnforcementAction{
			Type:    "throttle",
			Factor:  throttleFactor,
			Message: fmt.Sprintf("Budget throttling: $%.2f/$%.2f (%.1f%% capacity)", currentSpend, budget.Amount, throttleFactor*100),
		}, nil

	case utilization >= budget.WarningThreshold:
		e.sendAlert(tenantID, queueName, "budget_warning", currentSpend, budget.Amount, alertKey)
		return EnforcementAction{
			Type:    "warn",
			Message: fmt.Sprintf("Budget warning: $%.2f/$%.2f (%.1f%%)", currentSpend, budget.Amount, utilization*100),
		}, nil

	default:
		return EnforcementAction{Type: "allow"}, nil
	}
}

// CheckJobAllowed checks if a specific job should be allowed based on budget and priority
func (e *BudgetEnforcer) CheckJobAllowed(tenantID, queueName string, jobPriority int) (EnforcementAction, error) {
	action, err := e.CheckBudgetCompliance(tenantID, queueName)
	if err != nil {
		return action, err
	}

	// Handle emergency bypass for high-priority jobs
	if action.Type == "block" || action.Type == "throttle" {
		budget, err := e.budgetManager.GetBudgetForTenant(tenantID, queueName)
		if err != nil {
			return action, err
		}

		// Allow emergency bypass for high-priority jobs if configured
		if budget.EnforcementPolicy.AllowEmergency && jobPriority >= 9 { // Priority 9-10 are emergency
			action.Type = "allow"
			action.Message = "Emergency bypass for high-priority job"
			action.BypassAllowed = true
		}
	}

	return action, nil
}

// sendAlert sends budget alerts with rate limiting
func (e *BudgetEnforcer) sendAlert(tenantID, queueName, alertType string, currentSpend, budgetAmount float64, alertKey string) {
	// Rate limit alerts (max once per hour per tenant/queue)
	lastAlert, exists := e.alertTracker[alertKey]
	if exists && time.Since(lastAlert) < time.Hour {
		return // Skip alert due to rate limiting
	}

	e.alertTracker[alertKey] = time.Now()

	// Send notification
	if e.notifier != nil {
		alert := BudgetAlert{
			TenantID:     tenantID,
			QueueName:    queueName,
			AlertType:    alertType,
			CurrentSpend: currentSpend,
			BudgetAmount: budgetAmount,
			Utilization:  currentSpend / budgetAmount,
			CreatedAt:    time.Now(),
		}

		e.notifier.SendBudgetAlert(alert)
	}
}

// GetEnforcementSummary returns a summary of current enforcement actions
func (e *BudgetEnforcer) GetEnforcementSummary(tenantID string) (*EnforcementSummary, error) {
	budgets, err := e.budgetManager.ListBudgets(tenantID, true)
	if err != nil {
		return nil, err
	}

	summary := &EnforcementSummary{
		TenantID:       tenantID,
		TotalBudgets:   len(budgets),
		ActiveWarnings: 0,
		ActiveThrottles: 0,
		ActiveBlocks:   0,
		GeneratedAt:    time.Now(),
	}

	for _, budget := range budgets {
		action, err := e.CheckBudgetCompliance(tenantID, budget.QueueName)
		if err != nil {
			continue // Skip budgets with errors
		}

		switch action.Type {
		case "warn":
			summary.ActiveWarnings++
		case "throttle":
			summary.ActiveThrottles++
		case "block":
			summary.ActiveBlocks++
		}
	}

	summary.TotalActive = summary.ActiveWarnings + summary.ActiveThrottles + summary.ActiveBlocks

	return summary, nil
}

// EnforcementSummary provides an overview of budget enforcement status
type EnforcementSummary struct {
	TenantID        string    `json:"tenant_id"`
	TotalBudgets    int       `json:"total_budgets"`
	TotalActive     int       `json:"total_active"`
	ActiveWarnings  int       `json:"active_warnings"`
	ActiveThrottles int       `json:"active_throttles"`
	ActiveBlocks    int       `json:"active_blocks"`
	GeneratedAt     time.Time `json:"generated_at"`
}

// ResetAlertTracker clears the alert tracking cache
func (e *BudgetEnforcer) ResetAlertTracker() {
	e.alertTracker = make(map[string]time.Time)
}

// GetThrottleFactor returns the appropriate throttle factor for a tenant/queue
func (e *BudgetEnforcer) GetThrottleFactor(tenantID, queueName string) float64 {
	action, err := e.CheckBudgetCompliance(tenantID, queueName)
	if err != nil || action.Type != "throttle" {
		return 1.0 // No throttling
	}

	return action.Factor
}

// ShouldBlockJob determines if a job should be blocked
func (e *BudgetEnforcer) ShouldBlockJob(tenantID, queueName string, jobPriority int) bool {
	action, err := e.CheckJobAllowed(tenantID, queueName, jobPriority)
	if err != nil {
		// In case of error, allow the job to proceed (fail open)
		return false
	}

	return action.Type == "block"
}

// GetBudgetViolations returns current budget violations
func (e *BudgetEnforcer) GetBudgetViolations(tenantID string) ([]BudgetViolation, error) {
	budgets, err := e.budgetManager.ListBudgets(tenantID, true)
	if err != nil {
		return nil, err
	}

	var violations []BudgetViolation

	for _, budget := range budgets {
		status, err := e.budgetManager.GetBudgetStatus(budget.ID)
		if err != nil {
			continue
		}

		if status.CurrentThreshold != "none" {
			violation := BudgetViolation{
				BudgetID:         budget.ID,
				TenantID:         budget.TenantID,
				QueueName:        budget.QueueName,
				ViolationType:    status.CurrentThreshold,
				CurrentSpend:     status.CurrentSpend,
				BudgetAmount:     status.BudgetAmount,
				Utilization:      status.Utilization,
				DaysRemaining:    status.DaysRemaining,
				ProjectedOverrun: status.ProjectedSpend - status.BudgetAmount,
				DetectedAt:       time.Now(),
			}

			if violation.ProjectedOverrun < 0 {
				violation.ProjectedOverrun = 0
			}

			violations = append(violations, violation)
		}
	}

	return violations, nil
}

// BudgetViolation represents a budget threshold violation
type BudgetViolation struct {
	BudgetID         string    `json:"budget_id"`
	TenantID         string    `json:"tenant_id"`
	QueueName        string    `json:"queue_name"`
	ViolationType    string    `json:"violation_type"`    // "warning", "throttle", "block"
	CurrentSpend     float64   `json:"current_spend"`
	BudgetAmount     float64   `json:"budget_amount"`
	Utilization      float64   `json:"utilization"`
	DaysRemaining    int       `json:"days_remaining"`
	ProjectedOverrun float64   `json:"projected_overrun"` // How much over budget at period end
	DetectedAt       time.Time `json:"detected_at"`
}