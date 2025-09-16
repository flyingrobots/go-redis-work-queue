package calendarview

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateCalendarView(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name        string
		view        *CalendarView
		expectError bool
		errorCode   ErrorCode
	}{
		{
			name: "Valid view",
			view: &CalendarView{
				ViewType:    ViewTypeMonth,
				CurrentDate: time.Now(),
				Timezone:    time.UTC,
				Filter:      EventFilter{},
			},
			expectError: false,
		},
		{
			name:        "Nil view",
			view:        nil,
			expectError: true,
			errorCode:   ErrorCodeRuleValidation,
		},
		{
			name: "Invalid view type",
			view: &CalendarView{
				ViewType:    ViewType(999),
				CurrentDate: time.Now(),
				Timezone:    time.UTC,
			},
			expectError: true,
			errorCode:   ErrorCodeInvalidFilter,
		},
		{
			name: "Nil timezone",
			view: &CalendarView{
				ViewType:    ViewTypeMonth,
				CurrentDate: time.Now(),
				Timezone:    nil,
			},
			expectError: true,
			errorCode:   ErrorCodeTimezoneNotFound,
		},
		{
			name: "Zero current date",
			view: &CalendarView{
				ViewType:    ViewTypeMonth,
				CurrentDate: time.Time{},
				Timezone:    time.UTC,
			},
			expectError: true,
			errorCode:   ErrorCodeInvalidTimeRange,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateCalendarView(tc.view)

			if tc.expectError {
				assert.Error(t, err)
				if calErr, ok := err.(*CalendarError); ok {
					assert.Equal(t, tc.errorCode, calErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateEventFilter(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name        string
		filter      *EventFilter
		expectError bool
		errorCode   ErrorCode
	}{
		{
			name:        "Nil filter",
			filter:      nil,
			expectError: false,
		},
		{
			name: "Valid filter",
			filter: &EventFilter{
				QueueNames: []string{"queue1", "queue2"},
				JobTypes:   []string{"job1", "job2"},
				Tags:       []string{"tag1", "tag2"},
				Statuses:   []EventStatus{StatusScheduled, StatusRunning},
				SearchText: "test search",
			},
			expectError: false,
		},
		{
			name: "Too many queue names",
			filter: &EventFilter{
				QueueNames: make([]string, 11), // Max is 10
			},
			expectError: true,
			errorCode:   ErrorCodeInvalidFilter,
		},
		{
			name: "Too many job types",
			filter: &EventFilter{
				JobTypes: make([]string, 21), // Max is 20
			},
			expectError: true,
			errorCode:   ErrorCodeInvalidFilter,
		},
		{
			name: "Too many tags",
			filter: &EventFilter{
				Tags: make([]string, 16), // Max is 15
			},
			expectError: true,
			errorCode:   ErrorCodeInvalidFilter,
		},
		{
			name: "Search text too long",
			filter: &EventFilter{
				SearchText: string(make([]rune, 101)), // Max is 100
			},
			expectError: true,
			errorCode:   ErrorCodeInvalidFilter,
		},
		{
			name: "Invalid queue name",
			filter: &EventFilter{
				QueueNames: []string{"invalid queue name!"}, // Contains invalid characters
			},
			expectError: true,
			errorCode:   ErrorCodeRuleValidation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateEventFilter(tc.filter)

			if tc.expectError {
				assert.Error(t, err)
				if calErr, ok := err.(*CalendarError); ok {
					assert.Equal(t, tc.errorCode, calErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateRescheduleRequest(t *testing.T) {
	validator := NewValidator()

	validTime := time.Now().Add(24 * time.Hour)

	testCases := []struct {
		name        string
		request     *RescheduleRequest
		expectError bool
	}{
		{
			name: "Valid request",
			request: &RescheduleRequest{
				EventID: "event123",
				NewTime: validTime,
				UserID:  "user123",
				Reason:  "Testing reschedule",
			},
			expectError: false,
		},
		{
			name: "Empty event ID",
			request: &RescheduleRequest{
				EventID: "",
				NewTime: validTime,
				UserID:  "user123",
			},
			expectError: true,
		},
		{
			name: "Event ID too long",
			request: &RescheduleRequest{
				EventID: string(make([]rune, 51)), // Max is 50
				NewTime: validTime,
				UserID:  "user123",
			},
			expectError: true,
		},
		{
			name: "Zero new time",
			request: &RescheduleRequest{
				EventID: "event123",
				NewTime: time.Time{},
				UserID:  "user123",
			},
			expectError: true,
		},
		{
			name: "New time too far in past",
			request: &RescheduleRequest{
				EventID: "event123",
				NewTime: time.Now().Add(-48 * time.Hour), // More than 24 hours ago
				UserID:  "user123",
			},
			expectError: true,
		},
		{
			name: "New time too far in future",
			request: &RescheduleRequest{
				EventID: "event123",
				NewTime: time.Now().Add(400 * 24 * time.Hour), // More than 1 year
				UserID:  "user123",
			},
			expectError: true,
		},
		{
			name: "Empty user ID",
			request: &RescheduleRequest{
				EventID: "event123",
				NewTime: validTime,
				UserID:  "",
			},
			expectError: true,
		},
		{
			name: "Reason too long",
			request: &RescheduleRequest{
				EventID: "event123",
				NewTime: validTime,
				UserID:  "user123",
				Reason:  string(make([]rune, 501)), // Max is 500
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateRescheduleRequest(tc.request)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateRuleCreateRequest(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name        string
		request     *RuleCreateRequest
		expectError bool
	}{
		{
			name: "Valid request",
			request: &RuleCreateRequest{
				Name:        "Test Rule",
				CronSpec:    "0 0 * * *",
				QueueName:   "test-queue",
				JobType:     "test-job",
				Timezone:    "UTC",
				MaxInFlight: 5,
				Jitter:      30 * time.Minute,
				Metadata:    map[string]string{"key": "value"},
			},
			expectError: false,
		},
		{
			name: "Empty name",
			request: &RuleCreateRequest{
				Name:      "",
				CronSpec:  "0 0 * * *",
				QueueName: "test-queue",
				JobType:   "test-job",
			},
			expectError: true,
		},
		{
			name: "Name too long",
			request: &RuleCreateRequest{
				Name:      string(make([]rune, 101)), // Max is 100
				CronSpec:  "0 0 * * *",
				QueueName: "test-queue",
				JobType:   "test-job",
			},
			expectError: true,
		},
		{
			name: "Invalid name characters",
			request: &RuleCreateRequest{
				Name:      "Test Rule @#$",
				CronSpec:  "0 0 * * *",
				QueueName: "test-queue",
				JobType:   "test-job",
			},
			expectError: true,
		},
		{
			name: "Empty cron spec",
			request: &RuleCreateRequest{
				Name:      "Test Rule",
				CronSpec:  "",
				QueueName: "test-queue",
				JobType:   "test-job",
			},
			expectError: true,
		},
		{
			name: "Invalid cron spec",
			request: &RuleCreateRequest{
				Name:      "Test Rule",
				CronSpec:  "invalid cron",
				QueueName: "test-queue",
				JobType:   "test-job",
			},
			expectError: true,
		},
		{
			name: "Empty queue name",
			request: &RuleCreateRequest{
				Name:      "Test Rule",
				CronSpec:  "0 0 * * *",
				QueueName: "",
				JobType:   "test-job",
			},
			expectError: true,
		},
		{
			name: "Empty job type",
			request: &RuleCreateRequest{
				Name:      "Test Rule",
				CronSpec:  "0 0 * * *",
				QueueName: "test-queue",
				JobType:   "",
			},
			expectError: true,
		},
		{
			name: "Invalid timezone",
			request: &RuleCreateRequest{
				Name:      "Test Rule",
				CronSpec:  "0 0 * * *",
				QueueName: "test-queue",
				JobType:   "test-job",
				Timezone:  "Invalid/Timezone",
			},
			expectError: true,
		},
		{
			name: "Negative max in-flight",
			request: &RuleCreateRequest{
				Name:        "Test Rule",
				CronSpec:    "0 0 * * *",
				QueueName:   "test-queue",
				JobType:     "test-job",
				MaxInFlight: -1,
			},
			expectError: true,
		},
		{
			name: "Max in-flight too high",
			request: &RuleCreateRequest{
				Name:        "Test Rule",
				CronSpec:    "0 0 * * *",
				QueueName:   "test-queue",
				JobType:     "test-job",
				MaxInFlight: 1001, // Max is 1000
			},
			expectError: true,
		},
		{
			name: "Negative jitter",
			request: &RuleCreateRequest{
				Name:      "Test Rule",
				CronSpec:  "0 0 * * *",
				QueueName: "test-queue",
				JobType:   "test-job",
				Jitter:    -1 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "Jitter too large",
			request: &RuleCreateRequest{
				Name:      "Test Rule",
				CronSpec:  "0 0 * * *",
				QueueName: "test-queue",
				JobType:   "test-job",
				Jitter:    25 * time.Hour, // Max is 24 hours
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateRuleCreateRequest(tc.request)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateRuleUpdateRequest(t *testing.T) {
	validator := NewValidator()

	validName := "Updated Rule"
	validCronSpec := "0 0 * * *"
	validMaxInFlight := 10
	validJitter := 15 * time.Minute

	testCases := []struct {
		name        string
		request     *RuleUpdateRequest
		expectError bool
	}{
		{
			name: "Valid request",
			request: &RuleUpdateRequest{
				ID:          "rule123",
				Name:        &validName,
				CronSpec:    &validCronSpec,
				MaxInFlight: &validMaxInFlight,
				Jitter:      &validJitter,
			},
			expectError: false,
		},
		{
			name: "Empty ID",
			request: &RuleUpdateRequest{
				ID:   "",
				Name: &validName,
			},
			expectError: true,
		},
		{
			name: "Empty name update",
			request: &RuleUpdateRequest{
				ID:   "rule123",
				Name: func() *string { empty := ""; return &empty }(),
			},
			expectError: true,
		},
		{
			name: "Invalid cron spec update",
			request: &RuleUpdateRequest{
				ID:       "rule123",
				CronSpec: func() *string { invalid := "invalid cron"; return &invalid }(),
			},
			expectError: true,
		},
		{
			name: "Negative max in-flight update",
			request: &RuleUpdateRequest{
				ID:          "rule123",
				MaxInFlight: func() *int { negative := -1; return &negative }(),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateRuleUpdateRequest(tc.request)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateTimeRange(t *testing.T) {
	validator := NewValidator()

	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	farFuture := now.Add(400 * 24 * time.Hour) // More than 365 days

	testCases := []struct {
		name        string
		start       time.Time
		end         time.Time
		expectError bool
		errorCode   ErrorCode
	}{
		{
			name:        "Valid range",
			start:       past,
			end:         future,
			expectError: false,
		},
		{
			name:        "Zero start time",
			start:       time.Time{},
			end:         future,
			expectError: true,
			errorCode:   ErrorCodeInvalidTimeRange,
		},
		{
			name:        "Zero end time",
			start:       past,
			end:         time.Time{},
			expectError: true,
			errorCode:   ErrorCodeInvalidTimeRange,
		},
		{
			name:        "Start after end",
			start:       future,
			end:         past,
			expectError: true,
			errorCode:   ErrorCodeInvalidTimeRange,
		},
		{
			name:        "Range too large",
			start:       now,
			end:         farFuture,
			expectError: true,
			errorCode:   ErrorCodeInvalidTimeRange,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateTimeRange(tc.start, tc.end)

			if tc.expectError {
				assert.Error(t, err)
				if calErr, ok := err.(*CalendarError); ok {
					assert.Equal(t, tc.errorCode, calErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateScheduleWindow(t *testing.T) {
	validator := NewValidator()

	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	testCases := []struct {
		name        string
		window      *ScheduleWindow
		expectError bool
		errorCode   ErrorCode
	}{
		{
			name: "Valid window",
			window: &ScheduleWindow{
				From:      past,
				Till:      future,
				QueueName: "test-queue",
				Limit:     100,
			},
			expectError: false,
		},
		{
			name:        "Nil window",
			window:      nil,
			expectError: true,
			errorCode:   ErrorCodeRuleValidation,
		},
		{
			name: "Invalid time range",
			window: &ScheduleWindow{
				From: future,
				Till: past,
			},
			expectError: true,
			errorCode:   ErrorCodeInvalidTimeRange,
		},
		{
			name: "Invalid queue name",
			window: &ScheduleWindow{
				From:      past,
				Till:      future,
				QueueName: "invalid queue!",
			},
			expectError: true,
			errorCode:   ErrorCodeRuleValidation,
		},
		{
			name: "Negative limit",
			window: &ScheduleWindow{
				From:  past,
				Till:  future,
				Limit: -1,
			},
			expectError: true,
			errorCode:   ErrorCodeInvalidFilter,
		},
		{
			name: "Limit too large",
			window: &ScheduleWindow{
				From:  past,
				Till:  future,
				Limit: 10001, // Max is 10000
			},
			expectError: true,
			errorCode:   ErrorCodeInvalidFilter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateScheduleWindow(tc.window)

			if tc.expectError {
				assert.Error(t, err)
				if calErr, ok := err.(*CalendarError); ok {
					assert.Equal(t, tc.errorCode, calErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_IsValidIdentifier(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		input    string
		expected bool
	}{
		{"valid_name", true},
		{"valid-name", true},
		{"valid123", true},
		{"Valid_Name_123", true},
		{"", false},
		{"invalid name", false}, // Contains space
		{"invalid@name", false}, // Contains @
		{"invalid.name", false}, // Contains .
		{"invalid!name", false}, // Contains !
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := validator.isValidIdentifier(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidator_ValidateCronSpec(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name        string
		cronSpec    string
		expectError bool
	}{
		{"Valid daily", "0 0 * * *", false},
		{"Valid hourly", "0 * * * *", false},
		{"Valid weekly", "0 0 * * 0", false},
		{"Valid with seconds", "0 0 0 * * *", false},
		{"Empty spec", "", true},
		{"Invalid format", "invalid", true},
		{"Too many fields", "0 0 0 0 0 0 0", true},
		{"Every second (too frequent)", "* * * * * *", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.validateCronSpec(tc.cronSpec)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateMetadata(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name        string
		metadata    map[string]string
		expectError bool
	}{
		{
			name: "Valid metadata",
			metadata: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectError: false,
		},
		{
			name:        "Nil metadata",
			metadata:    nil,
			expectError: false,
		},
		{
			name:        "Empty metadata",
			metadata:    map[string]string{},
			expectError: false,
		},
		{
			name:        "Too many entries",
			metadata:    generateLargeMetadata(21), // Max is 20
			expectError: true,
		},
		{
			name: "Key too long",
			metadata: map[string]string{
				string(make([]rune, 51)): "value", // Max key length is 50
			},
			expectError: true,
		},
		{
			name: "Value too long",
			metadata: map[string]string{
				"key": string(make([]rune, 501)), // Max value length is 500
			},
			expectError: true,
		},
		{
			name: "Invalid key characters",
			metadata: map[string]string{
				"invalid key!": "value",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.validateMetadata(tc.metadata)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to generate large metadata for testing
func generateLargeMetadata(count int) map[string]string {
	metadata := make(map[string]string)
	for i := 0; i < count; i++ {
		metadata[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}
	return metadata
}

// Test ValidationErrors type

func TestValidationErrors(t *testing.T) {
	errors := NewValidationErrors()

	// Test empty errors
	assert.False(t, errors.HasErrors())
	assert.Equal(t, "validation failed", errors.Error())

	// Add first error
	errors.Add("field1", "error message 1", "value1")
	assert.True(t, errors.HasErrors())
	assert.Equal(t, "validation failed: error message 1", errors.Error())

	// Add second error
	errors.Add("field2", "error message 2")
	assert.Equal(t, "validation failed: 2 errors", errors.Error())
	assert.Equal(t, 2, len(errors.Errors))

	// Check error details
	assert.Equal(t, "field1", errors.Errors[0].Field)
	assert.Equal(t, "error message 1", errors.Errors[0].Message)
	assert.Equal(t, "value1", errors.Errors[0].Value)

	assert.Equal(t, "field2", errors.Errors[1].Field)
	assert.Equal(t, "error message 2", errors.Errors[1].Message)
	assert.Equal(t, "", errors.Errors[1].Value)
}