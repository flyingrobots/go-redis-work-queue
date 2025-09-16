package calendarview

import (
	"regexp"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// Validator provides validation for calendar operations
type Validator struct {
	cronParser cron.Parser
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	return &Validator{
		cronParser: parser,
	}
}

// ValidateCalendarView validates a calendar view configuration
func (v *Validator) ValidateCalendarView(view *CalendarView) error {
	if view == nil {
		return NewCalendarError(ErrorCodeRuleValidation, "calendar view cannot be nil")
	}

	// Validate view type
	if view.ViewType < ViewTypeMonth || view.ViewType > ViewTypeDay {
		return NewCalendarError(ErrorCodeInvalidFilter, "invalid view type")
	}

	// Validate timezone
	if view.Timezone == nil {
		return NewCalendarError(ErrorCodeTimezoneNotFound, "timezone cannot be nil")
	}

	// Validate current date
	if view.CurrentDate.IsZero() {
		return NewCalendarError(ErrorCodeInvalidTimeRange, "current date cannot be zero")
	}

	// Validate filter
	if err := v.ValidateEventFilter(&view.Filter); err != nil {
		return err
	}

	return nil
}

// ValidateEventFilter validates an event filter
func (v *Validator) ValidateEventFilter(filter *EventFilter) error {
	if filter == nil {
		return nil // Filter is optional
	}

	// Validate queue names
	if len(filter.QueueNames) > 10 {
		return NewCalendarError(ErrorCodeInvalidFilter, "too many queue names", "maximum 10 allowed")
	}
	for _, queueName := range filter.QueueNames {
		if err := v.validateQueueName(queueName); err != nil {
			return err
		}
	}

	// Validate job types
	if len(filter.JobTypes) > 20 {
		return NewCalendarError(ErrorCodeInvalidFilter, "too many job types", "maximum 20 allowed")
	}
	for _, jobType := range filter.JobTypes {
		if err := v.validateJobType(jobType); err != nil {
			return err
		}
	}

	// Validate tags
	if len(filter.Tags) > 15 {
		return NewCalendarError(ErrorCodeInvalidFilter, "too many tags", "maximum 15 allowed")
	}
	for _, tag := range filter.Tags {
		if err := v.validateTag(tag); err != nil {
			return err
		}
	}

	// Validate statuses
	for _, status := range filter.Statuses {
		if err := v.validateEventStatus(status); err != nil {
			return err
		}
	}

	// Validate search text
	if len(filter.SearchText) > 100 {
		return NewCalendarError(ErrorCodeInvalidFilter, "search text too long", "maximum 100 characters")
	}

	return nil
}

// ValidateRescheduleRequest validates a reschedule request
func (v *Validator) ValidateRescheduleRequest(req *RescheduleRequest) error {
	errors := NewValidationErrors()

	// Validate event ID
	if req.EventID == "" {
		errors.Add("event_id", "event ID is required")
	} else if len(req.EventID) > 50 {
		errors.Add("event_id", "event ID too long", req.EventID)
	}

	// Validate new time
	if req.NewTime.IsZero() {
		errors.Add("new_time", "new time is required")
	} else {
		// Check if new time is not too far in the past
		if req.NewTime.Before(time.Now().Add(-24 * time.Hour)) {
			errors.Add("new_time", "cannot reschedule to more than 24 hours in the past")
		}
		// Check if new time is not too far in the future (1 year)
		if req.NewTime.After(time.Now().Add(365 * 24 * time.Hour)) {
			errors.Add("new_time", "cannot reschedule to more than 1 year in the future")
		}
	}

	// Validate new queue (optional)
	if req.NewQueue != "" {
		if err := v.validateQueueName(req.NewQueue); err != nil {
			errors.Add("new_queue", err.Error())
		}
	}

	// Validate reason (optional but recommended)
	if len(req.Reason) > 500 {
		errors.Add("reason", "reason too long", "maximum 500 characters")
	}

	// Validate user ID
	if req.UserID == "" {
		errors.Add("user_id", "user ID is required")
	} else if len(req.UserID) > 50 {
		errors.Add("user_id", "user ID too long", req.UserID)
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateRuleCreateRequest validates a rule creation request
func (v *Validator) ValidateRuleCreateRequest(req *RuleCreateRequest) error {
	errors := NewValidationErrors()

	// Validate name
	if req.Name == "" {
		errors.Add("name", "name is required")
	} else if len(req.Name) > 100 {
		errors.Add("name", "name too long", "maximum 100 characters")
	} else if !v.isValidName(req.Name) {
		errors.Add("name", "name contains invalid characters")
	}

	// Validate cron spec
	if req.CronSpec == "" {
		errors.Add("cron_spec", "cron specification is required")
	} else if err := v.validateCronSpec(req.CronSpec); err != nil {
		errors.Add("cron_spec", err.Error())
	}

	// Validate queue name
	if req.QueueName == "" {
		errors.Add("queue_name", "queue name is required")
	} else if err := v.validateQueueName(req.QueueName); err != nil {
		errors.Add("queue_name", err.Error())
	}

	// Validate job type
	if req.JobType == "" {
		errors.Add("job_type", "job type is required")
	} else if err := v.validateJobType(req.JobType); err != nil {
		errors.Add("job_type", err.Error())
	}

	// Validate timezone
	if req.Timezone == "" {
		req.Timezone = "UTC" // Default
	} else if _, err := time.LoadLocation(req.Timezone); err != nil {
		errors.Add("timezone", "invalid timezone", req.Timezone)
	}

	// Validate max in-flight
	if req.MaxInFlight < 0 {
		errors.Add("max_in_flight", "max in-flight cannot be negative")
	} else if req.MaxInFlight > 1000 {
		errors.Add("max_in_flight", "max in-flight too high", "maximum 1000")
	}

	// Validate jitter
	if req.Jitter < 0 {
		errors.Add("jitter", "jitter cannot be negative")
	} else if req.Jitter > 24*time.Hour {
		errors.Add("jitter", "jitter too large", "maximum 24 hours")
	}

	// Validate metadata
	if err := v.validateMetadata(req.Metadata); err != nil {
		errors.Add("metadata", err.Error())
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateRuleUpdateRequest validates a rule update request
func (v *Validator) ValidateRuleUpdateRequest(req *RuleUpdateRequest) error {
	errors := NewValidationErrors()

	// Validate ID
	if req.ID == "" {
		errors.Add("id", "rule ID is required")
	}

	// Validate name (if provided)
	if req.Name != nil {
		if *req.Name == "" {
			errors.Add("name", "name cannot be empty")
		} else if len(*req.Name) > 100 {
			errors.Add("name", "name too long", "maximum 100 characters")
		} else if !v.isValidName(*req.Name) {
			errors.Add("name", "name contains invalid characters")
		}
	}

	// Validate cron spec (if provided)
	if req.CronSpec != nil {
		if *req.CronSpec == "" {
			errors.Add("cron_spec", "cron specification cannot be empty")
		} else if err := v.validateCronSpec(*req.CronSpec); err != nil {
			errors.Add("cron_spec", err.Error())
		}
	}

	// Validate max in-flight (if provided)
	if req.MaxInFlight != nil {
		if *req.MaxInFlight < 0 {
			errors.Add("max_in_flight", "max in-flight cannot be negative")
		} else if *req.MaxInFlight > 1000 {
			errors.Add("max_in_flight", "max in-flight too high", "maximum 1000")
		}
	}

	// Validate jitter (if provided)
	if req.Jitter != nil {
		if *req.Jitter < 0 {
			errors.Add("jitter", "jitter cannot be negative")
		} else if *req.Jitter > 24*time.Hour {
			errors.Add("jitter", "jitter too large", "maximum 24 hours")
		}
	}

	// Validate metadata (if provided)
	if req.Metadata != nil {
		if err := v.validateMetadata(req.Metadata); err != nil {
			errors.Add("metadata", err.Error())
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// validateQueueName validates a queue name
func (v *Validator) validateQueueName(name string) error {
	if name == "" {
		return NewCalendarError(ErrorCodeRuleValidation, "queue name cannot be empty")
	}
	if len(name) > 50 {
		return NewCalendarError(ErrorCodeRuleValidation, "queue name too long", "maximum 50 characters")
	}
	if !v.isValidIdentifier(name) {
		return NewCalendarError(ErrorCodeRuleValidation, "queue name contains invalid characters")
	}
	return nil
}

// validateJobType validates a job type
func (v *Validator) validateJobType(jobType string) error {
	if jobType == "" {
		return NewCalendarError(ErrorCodeRuleValidation, "job type cannot be empty")
	}
	if len(jobType) > 100 {
		return NewCalendarError(ErrorCodeRuleValidation, "job type too long", "maximum 100 characters")
	}
	if !v.isValidIdentifier(jobType) {
		return NewCalendarError(ErrorCodeRuleValidation, "job type contains invalid characters")
	}
	return nil
}

// validateTag validates a tag
func (v *Validator) validateTag(tag string) error {
	if tag == "" {
		return NewCalendarError(ErrorCodeRuleValidation, "tag cannot be empty")
	}
	if len(tag) > 50 {
		return NewCalendarError(ErrorCodeRuleValidation, "tag too long", "maximum 50 characters")
	}
	if !v.isValidTag(tag) {
		return NewCalendarError(ErrorCodeRuleValidation, "tag contains invalid characters")
	}
	return nil
}

// validateEventStatus validates an event status
func (v *Validator) validateEventStatus(status EventStatus) error {
	switch status {
	case StatusScheduled, StatusRunning, StatusCompleted, StatusFailed, StatusCanceled:
		return nil
	default:
		return NewCalendarError(ErrorCodeRuleValidation, "invalid event status")
	}
}

// validateCronSpec validates a cron specification
func (v *Validator) validateCronSpec(spec string) error {
	if spec == "" {
		return NewCalendarError(ErrorCodeInvalidCronSpec, "cron specification cannot be empty")
	}

	// Parse the cron specification
	_, err := v.cronParser.Parse(spec)
	if err != nil {
		return WrapCalendarError(ErrorCodeInvalidCronSpec, "invalid cron specification", err)
	}

	// Additional validation for safety
	if err := v.validateCronSafety(spec); err != nil {
		return err
	}

	return nil
}

// validateCronSafety performs additional safety checks on cron specifications
func (v *Validator) validateCronSafety(spec string) error {
	// Check for overly frequent schedules (less than 1 minute)
	// This is a basic check - in practice, you might want more sophisticated validation
	fields := strings.Fields(spec)
	if len(fields) >= 2 {
		// Check if seconds field is present and is "*" (every second)
		if len(fields) >= 6 && fields[0] == "*" {
			return NewCalendarError(ErrorCodeInvalidCronSpec, "cron schedule too frequent", "minimum interval is 1 minute")
		}
	}

	return nil
}

// validateMetadata validates metadata map
func (v *Validator) validateMetadata(metadata map[string]string) error {
	if len(metadata) > 20 {
		return NewCalendarError(ErrorCodeRuleValidation, "too many metadata entries", "maximum 20 allowed")
	}

	for key, value := range metadata {
		if len(key) > 50 {
			return NewCalendarError(ErrorCodeRuleValidation, "metadata key too long", "maximum 50 characters")
		}
		if len(value) > 500 {
			return NewCalendarError(ErrorCodeRuleValidation, "metadata value too long", "maximum 500 characters")
		}
		if !v.isValidMetadataKey(key) {
			return NewCalendarError(ErrorCodeRuleValidation, "metadata key contains invalid characters", key)
		}
	}

	return nil
}

// isValidIdentifier checks if a string is a valid identifier (alphanumeric + underscore + hyphen)
func (v *Validator) isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, s)
	return matched
}

// isValidName checks if a string is a valid name (alphanumeric + spaces + common punctuation)
func (v *Validator) isValidName(s string) bool {
	if s == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s\-_.()]+$`, s)
	return matched
}

// isValidTag checks if a string is a valid tag
func (v *Validator) isValidTag(s string) bool {
	if s == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, s)
	return matched
}

// isValidMetadataKey checks if a string is a valid metadata key
func (v *Validator) isValidMetadataKey(s string) bool {
	if s == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.-]+$`, s)
	return matched
}

// ValidateTimeRange validates a time range
func (v *Validator) ValidateTimeRange(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return NewCalendarError(ErrorCodeInvalidTimeRange, "start and end times cannot be zero")
	}
	if !start.Before(end) {
		return ErrInvalidTimeRange(start, end)
	}
	if end.Sub(start) > 365*24*time.Hour {
		return NewCalendarError(ErrorCodeInvalidTimeRange, "time range too large", "maximum 365 days")
	}
	return nil
}

// ValidateScheduleWindow validates a schedule window
func (v *Validator) ValidateScheduleWindow(window *ScheduleWindow) error {
	if window == nil {
		return NewCalendarError(ErrorCodeRuleValidation, "schedule window cannot be nil")
	}

	if err := v.ValidateTimeRange(window.From, window.Till); err != nil {
		return err
	}

	if window.QueueName != "" {
		if err := v.validateQueueName(window.QueueName); err != nil {
			return err
		}
	}

	if window.Limit < 0 {
		return NewCalendarError(ErrorCodeInvalidFilter, "limit cannot be negative")
	} else if window.Limit > 10000 {
		return NewCalendarError(ErrorCodeInvalidFilter, "limit too large", "maximum 10000")
	}

	return nil
}