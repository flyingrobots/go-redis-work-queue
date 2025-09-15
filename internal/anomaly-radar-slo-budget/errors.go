package anomalyradarslobudget

import (
	"errors"
	"fmt"
)

// Error constants for anomaly radar operations
var (
	// Configuration errors
	ErrInvalidConfig = errors.New("invalid configuration")
	ErrInvalidSLOTarget = errors.New("invalid SLO target")
	ErrInvalidThreshold = errors.New("invalid threshold")
	ErrInvalidWindow = errors.New("invalid time window")

	// Runtime errors
	ErrRadarNotRunning = errors.New("anomaly radar is not running")
	ErrRadarAlreadyRunning = errors.New("anomaly radar is already running")
	ErrMetricsCollectionFailed = errors.New("metrics collection failed")
	ErrInsufficientData = errors.New("insufficient data for analysis")

	// Alert errors
	ErrAlertNotFound = errors.New("alert not found")
	ErrInvalidAlertType = errors.New("invalid alert type")
	ErrAlertCallbackFailed = errors.New("alert callback failed")

	// SLO budget errors
	ErrBudgetCalculationFailed = errors.New("SLO budget calculation failed")
	ErrBurnRateCalculationFailed = errors.New("burn rate calculation failed")
)

// ConfigurationError represents an error in configuration validation
type ConfigurationError struct {
	Field   string
	Value   interface{}
	Message string
	Err     error
}

func (e *ConfigurationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("configuration error in field '%s': %s (value: %v): %v",
			e.Field, e.Message, e.Value, e.Err)
	}
	return fmt.Sprintf("configuration error in field '%s': %s (value: %v)",
		e.Field, e.Message, e.Value)
}

func (e *ConfigurationError) Unwrap() error {
	return e.Err
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(field string, value interface{}, message string, err error) *ConfigurationError {
	return &ConfigurationError{
		Field:   field,
		Value:   value,
		Message: message,
		Err:     err,
	}
}

// ThresholdError represents an error when a metric exceeds thresholds
type ThresholdError struct {
	MetricName string
	Value      float64
	Threshold  float64
	Severity   AlertLevel
	Timestamp  string
}

func (e *ThresholdError) Error() string {
	return fmt.Sprintf("threshold exceeded for %s: %.2f > %.2f (%s) at %s",
		e.MetricName, e.Value, e.Threshold, e.Severity.String(), e.Timestamp)
}

// NewThresholdError creates a new threshold error
func NewThresholdError(metricName string, value, threshold float64, severity AlertLevel, timestamp string) *ThresholdError {
	return &ThresholdError{
		MetricName: metricName,
		Value:      value,
		Threshold:  threshold,
		Severity:   severity,
		Timestamp:  timestamp,
	}
}

// SLOBudgetError represents an error in SLO budget calculations
type SLOBudgetError struct {
	BudgetType string
	Message    string
	Err        error
}

func (e *SLOBudgetError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("SLO budget error (%s): %s: %v", e.BudgetType, e.Message, e.Err)
	}
	return fmt.Sprintf("SLO budget error (%s): %s", e.BudgetType, e.Message)
}

func (e *SLOBudgetError) Unwrap() error {
	return e.Err
}

// NewSLOBudgetError creates a new SLO budget error
func NewSLOBudgetError(budgetType, message string, err error) *SLOBudgetError {
	return &SLOBudgetError{
		BudgetType: budgetType,
		Message:    message,
		Err:        err,
	}
}

// MetricsCollectionError represents an error during metrics collection
type MetricsCollectionError struct {
	Source    string
	Operation string
	Message   string
	Err       error
}

func (e *MetricsCollectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("metrics collection error from %s during %s: %s: %v",
			e.Source, e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("metrics collection error from %s during %s: %s",
		e.Source, e.Operation, e.Message)
}

func (e *MetricsCollectionError) Unwrap() error {
	return e.Err
}

// NewMetricsCollectionError creates a new metrics collection error
func NewMetricsCollectionError(source, operation, message string, err error) *MetricsCollectionError {
	return &MetricsCollectionError{
		Source:    source,
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// AlertError represents an error in alert processing
type AlertError struct {
	AlertID   string
	AlertType AlertType
	Operation string
	Message   string
	Err       error
}

func (e *AlertError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("alert error for %s (%s) during %s: %s: %v",
			e.AlertID, e.AlertType.String(), e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("alert error for %s (%s) during %s: %s",
		e.AlertID, e.AlertType.String(), e.Operation, e.Message)
}

func (e *AlertError) Unwrap() error {
	return e.Err
}

// NewAlertError creates a new alert error
func NewAlertError(alertID string, alertType AlertType, operation, message string, err error) *AlertError {
	return &AlertError{
		AlertID:   alertID,
		AlertType: alertType,
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// ValidationError represents a validation failure
type ValidationError struct {
	Component string
	Field     string
	Value     interface{}
	Rule      string
	Message   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s.%s: %s (value: %v, rule: %s)",
		e.Component, e.Field, e.Message, e.Value, e.Rule)
}

// NewValidationError creates a new validation error
func NewValidationError(component, field string, value interface{}, rule, message string) *ValidationError {
	return &ValidationError{
		Component: component,
		Field:     field,
		Value:     value,
		Rule:      rule,
		Message:   message,
	}
}

// IsConfigurationError checks if an error is a configuration error
func IsConfigurationError(err error) bool {
	var configErr *ConfigurationError
	return errors.As(err, &configErr)
}

// IsThresholdError checks if an error is a threshold error
func IsThresholdError(err error) bool {
	var thresholdErr *ThresholdError
	return errors.As(err, &thresholdErr)
}

// IsSLOBudgetError checks if an error is an SLO budget error
func IsSLOBudgetError(err error) bool {
	var sloErr *SLOBudgetError
	return errors.As(err, &sloErr)
}

// IsMetricsCollectionError checks if an error is a metrics collection error
func IsMetricsCollectionError(err error) bool {
	var metricsErr *MetricsCollectionError
	return errors.As(err, &metricsErr)
}

// IsAlertError checks if an error is an alert error
func IsAlertError(err error) bool {
	var alertErr *AlertError
	return errors.As(err, &alertErr)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}