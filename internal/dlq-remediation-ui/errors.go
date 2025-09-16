package dlqremediationui

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrEntryNotFound          = errors.New("DLQ entry not found")
	ErrInvalidPagination      = errors.New("invalid pagination parameters")
	ErrInvalidFilter          = errors.New("invalid filter parameters")
	ErrBulkOperationTooLarge  = errors.New("bulk operation exceeds maximum allowed size")
	ErrOperationInProgress    = errors.New("another operation is already in progress")
	ErrInvalidEntryID         = errors.New("invalid entry ID format")
	ErrRequeueFailed          = errors.New("failed to requeue entry")
	ErrPurgeFailed            = errors.New("failed to purge entry")
	ErrStorageUnavailable     = errors.New("storage backend unavailable")
	ErrAnalysisFailed         = errors.New("pattern analysis failed")
	ErrConfigurationInvalid   = errors.New("invalid configuration")
)

type DLQError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
	Context   string                 `json:"context,omitempty"`
}

func (e *DLQError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Context, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewDLQError(code, message string) *DLQError {
	return &DLQError{
		Code:      code,
		Message:   message,
		Timestamp: fmt.Sprintf("%d", getCurrentTimestamp()),
	}
}

func NewDLQErrorWithDetails(code, message string, details map[string]interface{}) *DLQError {
	return &DLQError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: fmt.Sprintf("%d", getCurrentTimestamp()),
	}
}

func NewDLQErrorWithContext(code, message, context string) *DLQError {
	return &DLQError{
		Code:      code,
		Message:   message,
		Context:   context,
		Timestamp: fmt.Sprintf("%d", getCurrentTimestamp()),
	}
}

func WrapError(err error, code, context string) *DLQError {
	return &DLQError{
		Code:      code,
		Message:   err.Error(),
		Context:   context,
		Timestamp: fmt.Sprintf("%d", getCurrentTimestamp()),
	}
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrEntryNotFound) ||
		   (isDLQError(err) && err.(*DLQError).Code == "ENTRY_NOT_FOUND")
}

func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrInvalidPagination) ||
		   errors.Is(err, ErrInvalidFilter) ||
		   errors.Is(err, ErrInvalidEntryID) ||
		   (isDLQError(err) && isValidationCode(err.(*DLQError).Code))
}

func IsOperationError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrRequeueFailed) ||
		   errors.Is(err, ErrPurgeFailed) ||
		   (isDLQError(err) && isOperationCode(err.(*DLQError).Code))
}

func IsStorageError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrStorageUnavailable) ||
		   (isDLQError(err) && err.(*DLQError).Code == "STORAGE_ERROR")
}

func isDLQError(err error) bool {
	_, ok := err.(*DLQError)
	return ok
}

func isValidationCode(code string) bool {
	validationCodes := []string{
		"INVALID_PAGINATION",
		"INVALID_FILTER",
		"INVALID_ENTRY_ID",
		"INVALID_CONFIGURATION",
	}
	for _, validCode := range validationCodes {
		if code == validCode {
			return true
		}
	}
	return false
}

func isOperationCode(code string) bool {
	operationCodes := []string{
		"REQUEUE_FAILED",
		"PURGE_FAILED",
		"BULK_OPERATION_FAILED",
		"OPERATION_IN_PROGRESS",
	}
	for _, opCode := range operationCodes {
		if code == opCode {
			return true
		}
	}
	return false
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}