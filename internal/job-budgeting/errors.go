package budgeting

import (
	"errors"
	"fmt"
)

var (
	// ErrBudgetNotFound is returned when a budget cannot be found
	ErrBudgetNotFound = errors.New("budget not found")

	// ErrBudgetExists is returned when trying to create a budget that already exists
	ErrBudgetExists = errors.New("budget already exists")

	// ErrInvalidBudgetPeriod is returned for invalid budget period configurations
	ErrInvalidBudgetPeriod = errors.New("invalid budget period")

	// ErrInvalidThreshold is returned when budget thresholds are invalid
	ErrInvalidThreshold = errors.New("invalid threshold")

	// ErrBudgetExceeded is returned when a budget is exceeded and enforcement is active
	ErrBudgetExceeded = errors.New("budget exceeded")

	// ErrInvalidCostModel is returned when cost model configuration is invalid
	ErrInvalidCostModel = errors.New("invalid cost model")

	// ErrInsufficientData is returned when there's not enough data for forecasting
	ErrInsufficientData = errors.New("insufficient data for analysis")

	// ErrEnforcementBlocked is returned when a job is blocked by budget enforcement
	ErrEnforcementBlocked = errors.New("job blocked by budget enforcement")

	// ErrThrottlingActive is returned when throttling is applied due to budget
	ErrThrottlingActive = errors.New("throttling active due to budget limits")

	// ErrInvalidNotificationChannel is returned for invalid notification configurations
	ErrInvalidNotificationChannel = errors.New("invalid notification channel")

	// ErrForecastFailed is returned when budget forecasting fails
	ErrForecastFailed = errors.New("forecast calculation failed")

	// ErrAggregationFailed is returned when cost aggregation fails
	ErrAggregationFailed = errors.New("cost aggregation failed")

	// ErrCalibrationFailed is returned when cost model calibration fails
	ErrCalibrationFailed = errors.New("cost model calibration failed")
)

// BudgetError wraps budget-specific errors with additional context
type BudgetError struct {
	BudgetID  string
	TenantID  string
	QueueName string
	Operation string
	Err       error
}

func (e *BudgetError) Error() string {
	msg := fmt.Sprintf("budget operation %s failed", e.Operation)
	if e.TenantID != "" {
		msg += fmt.Sprintf(" (tenant: %s)", e.TenantID)
	}
	if e.QueueName != "" {
		msg += fmt.Sprintf(" (queue: %s)", e.QueueName)
	}
	if e.BudgetID != "" {
		msg += fmt.Sprintf(" (budget: %s)", e.BudgetID)
	}
	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}
	return msg
}

func (e *BudgetError) Unwrap() error {
	return e.Err
}

// NewBudgetError creates a new budget error with context
func NewBudgetError(budgetID, tenantID, queueName, operation string, err error) *BudgetError {
	return &BudgetError{
		BudgetID:  budgetID,
		TenantID:  tenantID,
		QueueName: queueName,
		Operation: operation,
		Err:       err,
	}
}

// CostCalculationError represents errors in cost calculation
type CostCalculationError struct {
	JobID     string
	TenantID  string
	Component string
	Reason    string
	Err       error
}

func (e *CostCalculationError) Error() string {
	msg := fmt.Sprintf("cost calculation failed for job %s", e.JobID)
	if e.TenantID != "" {
		msg += fmt.Sprintf(" (tenant: %s)", e.TenantID)
	}
	if e.Component != "" {
		msg += fmt.Sprintf(" (component: %s)", e.Component)
	}
	if e.Reason != "" {
		msg += fmt.Sprintf(": %s", e.Reason)
	}
	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}
	return msg
}

func (e *CostCalculationError) Unwrap() error {
	return e.Err
}

// NewCostCalculationError creates a new cost calculation error
func NewCostCalculationError(jobID, tenantID, component, reason string, err error) *CostCalculationError {
	return &CostCalculationError{
		JobID:     jobID,
		TenantID:  tenantID,
		Component: component,
		Reason:    reason,
		Err:       err,
	}
}

// EnforcementError represents budget enforcement errors
type EnforcementError struct {
	TenantID      string
	QueueName     string
	Action        string
	CurrentSpend  float64
	BudgetAmount  float64
	Utilization   float64
	Reason        string
	BypassAllowed bool
}

func (e *EnforcementError) Error() string {
	msg := fmt.Sprintf("budget enforcement %s for tenant %s", e.Action, e.TenantID)
	if e.QueueName != "" {
		msg += fmt.Sprintf(" queue %s", e.QueueName)
	}
	msg += fmt.Sprintf(": spend $%.2f/$%.2f (%.1f%%)",
		e.CurrentSpend, e.BudgetAmount, e.Utilization*100)
	if e.Reason != "" {
		msg += fmt.Sprintf(" - %s", e.Reason)
	}
	if e.BypassAllowed {
		msg += " (bypass available)"
	}
	return msg
}

// NewEnforcementError creates a new enforcement error
func NewEnforcementError(tenantID, queueName, action string, currentSpend, budgetAmount, utilization float64, reason string, bypassAllowed bool) *EnforcementError {
	return &EnforcementError{
		TenantID:      tenantID,
		QueueName:     queueName,
		Action:        action,
		CurrentSpend:  currentSpend,
		BudgetAmount:  budgetAmount,
		Utilization:   utilization,
		Reason:        reason,
		BypassAllowed: bypassAllowed,
	}
}

// ValidationError represents configuration validation errors
type ValidationError struct {
	Field   string
	Value   interface{}
	Rule    string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %s (value: %v, rule: %s): %s",
		e.Field, e.Value, e.Rule, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, rule, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Rule:    rule,
		Message: message,
	}
}

// ForecastError represents forecasting errors
type ForecastError struct {
	TenantID  string
	QueueName string
	Period    string
	Reason    string
	Err       error
}

func (e *ForecastError) Error() string {
	msg := fmt.Sprintf("forecast failed for tenant %s", e.TenantID)
	if e.QueueName != "" {
		msg += fmt.Sprintf(" queue %s", e.QueueName)
	}
	if e.Period != "" {
		msg += fmt.Sprintf(" period %s", e.Period)
	}
	if e.Reason != "" {
		msg += fmt.Sprintf(": %s", e.Reason)
	}
	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}
	return msg
}

func (e *ForecastError) Unwrap() error {
	return e.Err
}

// NewForecastError creates a new forecast error
func NewForecastError(tenantID, queueName, period, reason string, err error) *ForecastError {
	return &ForecastError{
		TenantID:  tenantID,
		QueueName: queueName,
		Period:    period,
		Reason:    reason,
		Err:       err,
	}
}

// NotificationError represents notification delivery errors
type NotificationError struct {
	Channel     string
	Target      string
	Event       string
	RetryCount  int
	LastAttempt string
	Err         error
}

func (e *NotificationError) Error() string {
	msg := fmt.Sprintf("notification failed for %s channel %s", e.Channel, e.Target)
	if e.Event != "" {
		msg += fmt.Sprintf(" event %s", e.Event)
	}
	if e.RetryCount > 0 {
		msg += fmt.Sprintf(" (retry %d)", e.RetryCount)
	}
	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}
	return msg
}

func (e *NotificationError) Unwrap() error {
	return e.Err
}

// NewNotificationError creates a new notification error
func NewNotificationError(channel, target, event string, retryCount int, err error) *NotificationError {
	return &NotificationError{
		Channel:    channel,
		Target:     target,
		Event:      event,
		RetryCount: retryCount,
		Err:        err,
	}
}

// IsRetryable returns true if the error indicates a retryable condition
func IsRetryable(err error) bool {
	switch {
	case errors.Is(err, ErrInsufficientData):
		return true
	case errors.Is(err, ErrForecastFailed):
		return true
	case errors.Is(err, ErrAggregationFailed):
		return true
	default:
		// Check for wrapped errors
		var budgetErr *BudgetError
		if errors.As(err, &budgetErr) {
			return IsRetryable(budgetErr.Err)
		}

		var costErr *CostCalculationError
		if errors.As(err, &costErr) {
			return IsRetryable(costErr.Err)
		}

		var notifErr *NotificationError
		if errors.As(err, &notifErr) {
			return true // Notifications are always retryable
		}

		return false
	}
}

// IsPermanent returns true if the error indicates a permanent failure
func IsPermanent(err error) bool {
	switch {
	case errors.Is(err, ErrBudgetNotFound):
		return true
	case errors.Is(err, ErrInvalidBudgetPeriod):
		return true
	case errors.Is(err, ErrInvalidThreshold):
		return true
	case errors.Is(err, ErrInvalidCostModel):
		return true
	case errors.Is(err, ErrInvalidNotificationChannel):
		return true
	default:
		// Check for validation errors
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			return true
		}

		return false
	}
}

// ErrorCode returns a stable error code for the error
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrBudgetNotFound):
		return "BUDGET_NOT_FOUND"
	case errors.Is(err, ErrBudgetExists):
		return "BUDGET_EXISTS"
	case errors.Is(err, ErrInvalidBudgetPeriod):
		return "INVALID_BUDGET_PERIOD"
	case errors.Is(err, ErrInvalidThreshold):
		return "INVALID_THRESHOLD"
	case errors.Is(err, ErrBudgetExceeded):
		return "BUDGET_EXCEEDED"
	case errors.Is(err, ErrInvalidCostModel):
		return "INVALID_COST_MODEL"
	case errors.Is(err, ErrInsufficientData):
		return "INSUFFICIENT_DATA"
	case errors.Is(err, ErrEnforcementBlocked):
		return "ENFORCEMENT_BLOCKED"
	case errors.Is(err, ErrThrottlingActive):
		return "THROTTLING_ACTIVE"
	case errors.Is(err, ErrInvalidNotificationChannel):
		return "INVALID_NOTIFICATION_CHANNEL"
	case errors.Is(err, ErrForecastFailed):
		return "FORECAST_FAILED"
	case errors.Is(err, ErrAggregationFailed):
		return "AGGREGATION_FAILED"
	case errors.Is(err, ErrCalibrationFailed):
		return "CALIBRATION_FAILED"
	default:
		// Check for typed errors
		var budgetErr *BudgetError
		if errors.As(err, &budgetErr) {
			return "BUDGET_ERROR"
		}

		var costErr *CostCalculationError
		if errors.As(err, &costErr) {
			return "COST_CALCULATION_ERROR"
		}

		var enforcementErr *EnforcementError
		if errors.As(err, &enforcementErr) {
			return "ENFORCEMENT_ERROR"
		}

		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			return "VALIDATION_ERROR"
		}

		var forecastErr *ForecastError
		if errors.As(err, &forecastErr) {
			return "FORECAST_ERROR"
		}

		var notifErr *NotificationError
		if errors.As(err, &notifErr) {
			return "NOTIFICATION_ERROR"
		}

		return "UNKNOWN_ERROR"
	}
}