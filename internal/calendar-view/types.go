package calendarview

import (
	"time"
)

// ViewType represents the different calendar view modes
type ViewType int

const (
	ViewTypeMonth ViewType = iota
	ViewTypeWeek
	ViewTypeDay
)

// CalendarEvent represents a scheduled job or event in the calendar
type CalendarEvent struct {
	ID          string            `json:"id"`
	QueueName   string            `json:"queue_name"`
	JobType     string            `json:"job_type"`
	ScheduledAt time.Time         `json:"scheduled_at"`
	Status      EventStatus       `json:"status"`
	Priority    int               `json:"priority"`
	Metadata    map[string]string `json:"metadata"`
	Tags        []string          `json:"tags"`
}

// EventStatus represents the status of a calendar event
type EventStatus int

const (
	StatusScheduled EventStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
	StatusCanceled
)

// RecurringRule represents a cron-based recurring job rule
type RecurringRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	CronSpec    string            `json:"cron_spec"`
	QueueName   string            `json:"queue_name"`
	JobType     string            `json:"job_type"`
	Timezone    string            `json:"timezone"`
	IsActive    bool              `json:"is_active"`
	IsPaused    bool              `json:"is_paused"`
	MaxInFlight int               `json:"max_in_flight"`
	Jitter      time.Duration     `json:"jitter"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	NextRun     *time.Time        `json:"next_run"`
}

// CalendarView represents the main calendar view component
type CalendarView struct {
	ViewType    ViewType      `json:"view_type"`
	CurrentDate time.Time     `json:"current_date"`
	TimeRange   TimeRange     `json:"time_range"`
	Timezone    *time.Location `json:"timezone"`
	Filter      EventFilter   `json:"filter"`
}

// TimeRange represents a time window for calendar views
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// EventFilter represents filtering options for calendar events
type EventFilter struct {
	QueueNames []string      `json:"queue_names"`
	JobTypes   []string      `json:"job_types"`
	Tags       []string      `json:"tags"`
	Statuses   []EventStatus `json:"statuses"`
	SearchText string        `json:"search_text"`
}

// CalendarCell represents a single cell in the calendar view
type CalendarCell struct {
	Date       time.Time        `json:"date"`
	Events     []CalendarEvent  `json:"events"`
	EventCount int              `json:"event_count"`
	Density    float64          `json:"density"`
	IsToday    bool             `json:"is_today"`
	IsSelected bool             `json:"is_selected"`
}

// CalendarData represents the data for a calendar view
type CalendarData struct {
	Cells         [][]CalendarCell `json:"cells"`
	TotalEvents   int              `json:"total_events"`
	PeakDensity   float64          `json:"peak_density"`
	TimeRange     TimeRange        `json:"time_range"`
	RecurringRules []RecurringRule `json:"recurring_rules"`
}

// RescheduleRequest represents a request to reschedule an event
type RescheduleRequest struct {
	EventID     string    `json:"event_id"`
	NewTime     time.Time `json:"new_time"`
	NewQueue    string    `json:"new_queue,omitempty"`
	Reason      string    `json:"reason"`
	UserID      string    `json:"user_id"`
}

// RescheduleResult represents the result of a reschedule operation
type RescheduleResult struct {
	Success   bool      `json:"success"`
	OldTime   time.Time `json:"old_time"`
	NewTime   time.Time `json:"new_time"`
	EventID   string    `json:"event_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// NavigationAction represents keyboard navigation actions
type NavigationAction int

const (
	NavUp NavigationAction = iota
	NavDown
	NavLeft
	NavRight
	NavNextPeriod
	NavPrevPeriod
	NavToday
	NavEnd
	NavSelect
	NavReschedule
	NavPause
	NavResume
	NavFilter
	NavInspect
)

// KeyBinding represents a keyboard shortcut
type KeyBinding struct {
	Key         string           `json:"key"`
	Action      NavigationAction `json:"action"`
	Description string           `json:"description"`
}

// CalendarConfig represents configuration for the calendar view
type CalendarConfig struct {
	DefaultView     ViewType       `json:"default_view"`
	DefaultTimezone string         `json:"default_timezone"`
	MaxEventsPerCell int           `json:"max_events_per_cell"`
	KeyBindings     []KeyBinding   `json:"key_bindings"`
	Colors          ColorScheme    `json:"colors"`
	RefreshInterval time.Duration  `json:"refresh_interval"`
}

// ColorScheme represents the color configuration for the calendar
type ColorScheme struct {
	Background     string `json:"background"`
	Border         string `json:"border"`
	Text           string `json:"text"`
	Today          string `json:"today"`
	Selected       string `json:"selected"`
	LowDensity     string `json:"low_density"`
	MediumDensity  string `json:"medium_density"`
	HighDensity    string `json:"high_density"`
	StatusRunning  string `json:"status_running"`
	StatusFailed   string `json:"status_failed"`
	StatusCompleted string `json:"status_completed"`
}

// ScheduleWindow represents a time window for schedule queries
type ScheduleWindow struct {
	From      time.Time `json:"from"`
	Till      time.Time `json:"till"`
	QueueName string    `json:"queue_name,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

// ScheduleResponse represents the response from a schedule window query
type ScheduleResponse struct {
	Events     []CalendarEvent `json:"events"`
	TotalCount int             `json:"total_count"`
	Window     ScheduleWindow  `json:"window"`
	HasMore    bool            `json:"has_more"`
}

// RuleCreateRequest represents a request to create a recurring rule
type RuleCreateRequest struct {
	Name        string            `json:"name"`
	CronSpec    string            `json:"cron_spec"`
	QueueName   string            `json:"queue_name"`
	JobType     string            `json:"job_type"`
	Timezone    string            `json:"timezone"`
	MaxInFlight int               `json:"max_in_flight"`
	Jitter      time.Duration     `json:"jitter"`
	Metadata    map[string]string `json:"metadata"`
}

// RuleUpdateRequest represents a request to update a recurring rule
type RuleUpdateRequest struct {
	ID          string            `json:"id"`
	Name        *string           `json:"name,omitempty"`
	CronSpec    *string           `json:"cron_spec,omitempty"`
	IsActive    *bool             `json:"is_active,omitempty"`
	IsPaused    *bool             `json:"is_paused,omitempty"`
	MaxInFlight *int              `json:"max_in_flight,omitempty"`
	Jitter      *time.Duration    `json:"jitter,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// AuditEntry represents an audit log entry for calendar operations
type AuditEntry struct {
	ID        string            `json:"id"`
	Action    string            `json:"action"`
	UserID    string            `json:"user_id"`
	Timestamp time.Time         `json:"timestamp"`
	Details   map[string]string `json:"details"`
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
}