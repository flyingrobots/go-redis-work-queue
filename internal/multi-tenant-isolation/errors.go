// Copyright 2025 James Ross
package multitenantiso

import (
	"errors"
	"fmt"
)

// Common error variables
var (
	ErrInvalidTenantIDLength                 = errors.New("tenant ID length must be between 3 and 32 characters")
	ErrInvalidTenantIDFormat                 = errors.New("tenant ID must contain only lowercase alphanumeric characters and hyphens")
	ErrTenantIDMustNotStartOrEndWithHyphen   = errors.New("tenant ID must not start or end with a hyphen")
	ErrTenantNameRequired                    = errors.New("tenant name is required")
	ErrTenantNotFound                        = errors.New("tenant not found")
	ErrTenantAlreadyExists                   = errors.New("tenant already exists")
	ErrAccessDenied                          = errors.New("access denied")
	ErrQuotaExceeded                         = errors.New("quota exceeded")
	ErrEncryptionNotEnabled                  = errors.New("encryption not enabled for tenant")
	ErrDecryptionFailed                      = errors.New("failed to decrypt payload")
	ErrInvalidEncryptionConfig               = errors.New("invalid encryption configuration")
	ErrKMSUnavailable                        = errors.New("KMS service unavailable")
	ErrInvalidPermission                     = errors.New("invalid permission configuration")
)

// TenantNotFoundError is returned when a tenant doesn't exist
type TenantNotFoundError struct {
	TenantID TenantID
}

func (e TenantNotFoundError) Error() string {
	return fmt.Sprintf("tenant not found: %s", e.TenantID)
}

// TenantAlreadyExistsError is returned when trying to create a tenant that already exists
type TenantAlreadyExistsError struct {
	TenantID TenantID
}

func (e TenantAlreadyExistsError) Error() string {
	return fmt.Sprintf("tenant already exists: %s", e.TenantID)
}

// AccessDeniedError provides detailed information about access denial
type AccessDeniedError struct {
	Reason   string
	Resource string
	Action   string
	TenantID TenantID
	UserID   string
}

func (e AccessDeniedError) Error() string {
	if e.Resource != "" && e.Action != "" {
		return fmt.Sprintf("access denied for user %s: cannot %s resource %s in tenant %s (reason: %s)",
			e.UserID, e.Action, e.Resource, e.TenantID, e.Reason)
	}
	return fmt.Sprintf("access denied for user %s in tenant %s (reason: %s)",
		e.UserID, e.TenantID, e.Reason)
}

// QuotaExceededError provides detailed information about quota violations
type QuotaExceededError struct {
	Type     string   // Type of quota exceeded
	Current  int64    // Current usage
	Limit    int64    // Quota limit
	TenantID TenantID // Tenant that exceeded quota
}

func (e QuotaExceededError) Error() string {
	return fmt.Sprintf("quota exceeded for tenant %s: %s usage %d exceeds limit %d",
		e.TenantID, e.Type, e.Current, e.Limit)
}

// EncryptionError wraps encryption-related errors
type EncryptionError struct {
	Operation string   // "encrypt" or "decrypt"
	TenantID  TenantID // Tenant involved
	Err       error    // Underlying error
}

func (e EncryptionError) Error() string {
	return fmt.Sprintf("encryption %s failed for tenant %s: %v", e.Operation, e.TenantID, e.Err)
}

func (e EncryptionError) Unwrap() error {
	return e.Err
}

// ValidationError represents data validation failures
type ValidationError struct {
	Field   string // Field that failed validation
	Value   string // Value that was invalid
	Message string // Validation error message
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%s': %s", e.Field, e.Value, e.Message)
}

// AuditError represents audit logging failures
type AuditError struct {
	Action   string   // Action that failed to be audited
	TenantID TenantID // Tenant involved
	Err      error    // Underlying error
}

func (e AuditError) Error() string {
	return fmt.Sprintf("audit logging failed for action %s in tenant %s: %v", e.Action, e.TenantID, e.Err)
}

func (e AuditError) Unwrap() error {
	return e.Err
}

// ConfigurationError represents configuration-related errors
type ConfigurationError struct {
	Component string // Component with invalid configuration
	Issue     string // Description of the issue
	TenantID  TenantID
}

func (e ConfigurationError) Error() string {
	if e.TenantID != "" {
		return fmt.Sprintf("configuration error in %s for tenant %s: %s", e.Component, e.TenantID, e.Issue)
	}
	return fmt.Sprintf("configuration error in %s: %s", e.Component, e.Issue)
}

// RateLimitExceededError indicates rate limiting has been triggered
type RateLimitExceededError struct {
	TenantID     TenantID
	Operation    string // "enqueue", "dequeue", etc.
	CurrentRate  int32  // Current operations per second
	Limit        int32  // Rate limit per second
	RetryAfter   int    // Seconds to wait before retrying
}

func (e RateLimitExceededError) Error() string {
	return fmt.Sprintf("rate limit exceeded for tenant %s: %s rate %d/s exceeds limit %d/s (retry after %ds)",
		e.TenantID, e.Operation, e.CurrentRate, e.Limit, e.RetryAfter)
}

// KeyRotationError represents errors during encryption key rotation
type KeyRotationError struct {
	TenantID   TenantID
	KEKKeyID   string
	Phase      string // "generate", "encrypt", "store", "cleanup"
	Err        error
}

func (e KeyRotationError) Error() string {
	return fmt.Sprintf("key rotation failed for tenant %s (KEK: %s) during %s phase: %v",
		e.TenantID, e.KEKKeyID, e.Phase, e.Err)
}

func (e KeyRotationError) Unwrap() error {
	return e.Err
}

// StorageError wraps Redis or other storage-related errors
type StorageError struct {
	Operation string   // "get", "set", "delete", etc.
	Key       string   // Redis key involved
	TenantID  TenantID
	Err       error
}

func (e StorageError) Error() string {
	return fmt.Sprintf("storage %s failed for key %s in tenant %s: %v", e.Operation, e.Key, e.TenantID, e.Err)
}

func (e StorageError) Unwrap() error {
	return e.Err
}

// NewTenantNotFoundError creates a TenantNotFoundError
func NewTenantNotFoundError(tenantID TenantID) error {
	return TenantNotFoundError{TenantID: tenantID}
}

// NewTenantAlreadyExistsError creates a TenantAlreadyExistsError
func NewTenantAlreadyExistsError(tenantID TenantID) error {
	return TenantAlreadyExistsError{TenantID: tenantID}
}

// NewAccessDeniedError creates an AccessDeniedError
func NewAccessDeniedError(userID string, tenantID TenantID, resource, action, reason string) error {
	return AccessDeniedError{
		UserID:   userID,
		TenantID: tenantID,
		Resource: resource,
		Action:   action,
		Reason:   reason,
	}
}

// NewQuotaExceededError creates a QuotaExceededError
func NewQuotaExceededError(tenantID TenantID, quotaType string, current, limit int64) error {
	return QuotaExceededError{
		TenantID: tenantID,
		Type:     quotaType,
		Current:  current,
		Limit:    limit,
	}
}

// NewEncryptionError creates an EncryptionError
func NewEncryptionError(operation string, tenantID TenantID, err error) error {
	return EncryptionError{
		Operation: operation,
		TenantID:  tenantID,
		Err:       err,
	}
}

// NewValidationError creates a ValidationError
func NewValidationError(field, value, message string) error {
	return ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NewAuditError creates an AuditError
func NewAuditError(action string, tenantID TenantID, err error) error {
	return AuditError{
		Action:   action,
		TenantID: tenantID,
		Err:      err,
	}
}

// NewConfigurationError creates a ConfigurationError
func NewConfigurationError(component, issue string, tenantID TenantID) error {
	return ConfigurationError{
		Component: component,
		Issue:     issue,
		TenantID:  tenantID,
	}
}

// NewRateLimitExceededError creates a RateLimitExceededError
func NewRateLimitExceededError(tenantID TenantID, operation string, currentRate, limit int32, retryAfter int) error {
	return RateLimitExceededError{
		TenantID:    tenantID,
		Operation:   operation,
		CurrentRate: currentRate,
		Limit:       limit,
		RetryAfter:  retryAfter,
	}
}

// NewKeyRotationError creates a KeyRotationError
func NewKeyRotationError(tenantID TenantID, kekKeyID, phase string, err error) error {
	return KeyRotationError{
		TenantID: tenantID,
		KEKKeyID: kekKeyID,
		Phase:    phase,
		Err:      err,
	}
}

// NewStorageError creates a StorageError
func NewStorageError(operation, key string, tenantID TenantID, err error) error {
	return StorageError{
		Operation: operation,
		Key:       key,
		TenantID:  tenantID,
		Err:       err,
	}
}

// IsQuotaExceeded checks if an error is a quota exceeded error
func IsQuotaExceeded(err error) bool {
	var qe QuotaExceededError
	return errors.As(err, &qe)
}

// IsAccessDenied checks if an error is an access denied error
func IsAccessDenied(err error) bool {
	var ad AccessDeniedError
	return errors.As(err, &ad)
}

// IsTenantNotFound checks if an error is a tenant not found error
func IsTenantNotFound(err error) bool {
	var tnf TenantNotFoundError
	return errors.As(err, &tnf)
}

// IsRateLimited checks if an error is a rate limit exceeded error
func IsRateLimited(err error) bool {
	var rle RateLimitExceededError
	return errors.As(err, &rle)
}

// IsEncryptionError checks if an error is an encryption error
func IsEncryptionError(err error) bool {
	var ee EncryptionError
	return errors.As(err, &ee)
}