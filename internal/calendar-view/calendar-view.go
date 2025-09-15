package calendarview

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// CalendarManager manages calendar views and operations
type CalendarManager struct {
	config     *CalendarConfig
	dataSource DataSource
	cache      *CacheManager
	validator  *Validator
	auditor    *AuditLogger
}

// DataSource interface for retrieving calendar data
type DataSource interface {
	GetEvents(window ScheduleWindow) (*ScheduleResponse, error)
	GetRecurringRules(filter RuleFilter) ([]RecurringRule, error)
	RescheduleEvent(req *RescheduleRequest) (*RescheduleResult, error)
	CreateRule(req *RuleCreateRequest) (*RecurringRule, error)
	UpdateRule(req *RuleUpdateRequest) (*RecurringRule, error)
	DeleteRule(ruleID string) error
	PauseRule(ruleID string) error
	ResumeRule(ruleID string) error
}

// NewCalendarManager creates a new calendar manager
func NewCalendarManager(config *CalendarConfig, dataSource DataSource) (*CalendarManager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	validator := NewValidator()
	auditor := NewAuditLogger()
	cache := NewCacheManager(config.RefreshInterval)

	return &CalendarManager{
		config:     config,
		dataSource: dataSource,
		cache:      cache,
		validator:  validator,
		auditor:    auditor,
	}, nil
}

// GetCalendarData retrieves calendar data for the specified view and time range
func (cm *CalendarManager) GetCalendarData(view *CalendarView) (*CalendarData, error) {
	if err := cm.validator.ValidateCalendarView(view); err != nil {
		return nil, err
	}

	// Generate time range based on view type
	timeRange, err := cm.generateTimeRange(view)
	if err != nil {
		return nil, err
	}

	// Create schedule window
	window := ScheduleWindow{
		From:  timeRange.Start,
		Till:  timeRange.End,
		Limit: cm.config.MaxEventsPerCell * 50, // Generous limit
	}

	// Apply filters if present
	if len(view.Filter.QueueNames) > 0 {
		// For now, handle single queue filter
		// TODO: Support multiple queue filtering
		if len(view.Filter.QueueNames) == 1 {
			window.QueueName = view.Filter.QueueNames[0]
		}
	}

	// Get events from data source
	response, err := cm.dataSource.GetEvents(window)
	if err != nil {
		return nil, WrapCalendarError(ErrorCodeDatabaseError, "failed to retrieve events", err)
	}

	// Get recurring rules
	ruleFilter := RuleFilter{
		QueueNames: view.Filter.QueueNames,
		IsActive:   true,
	}
	rules, err := cm.dataSource.GetRecurringRules(ruleFilter)
	if err != nil {
		return nil, WrapCalendarError(ErrorCodeDatabaseError, "failed to retrieve recurring rules", err)
	}

	// Generate calendar cells
	cells, err := cm.generateCalendarCells(view, response.Events, timeRange)
	if err != nil {
		return nil, err
	}

	// Calculate peak density for color scaling
	peakDensity := cm.calculatePeakDensity(cells)

	return &CalendarData{
		Cells:          cells,
		TotalEvents:    response.TotalCount,
		PeakDensity:    peakDensity,
		TimeRange:      *timeRange,
		RecurringRules: rules,
	}, nil
}

// generateTimeRange creates a time range based on the view type and current date
func (cm *CalendarManager) generateTimeRange(view *CalendarView) (*TimeRange, error) {
	current := view.CurrentDate.In(view.Timezone)
	var start, end time.Time

	switch view.ViewType {
	case ViewTypeMonth:
		// Start of month
		start = time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, view.Timezone)
		// Start of next month
		end = start.AddDate(0, 1, 0)

	case ViewTypeWeek:
		// Start of week (Sunday)
		weekday := int(current.Weekday())
		start = current.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
		// End of week
		end = start.AddDate(0, 0, 7)

	case ViewTypeDay:
		// Start of day
		start = current.Truncate(24 * time.Hour)
		// Start of next day
		end = start.AddDate(0, 0, 1)

	default:
		return nil, NewCalendarError(ErrorCodeInvalidFilter, "unsupported view type", fmt.Sprintf("%d", view.ViewType))
	}

	return &TimeRange{Start: start, End: end}, nil
}

// generateCalendarCells creates calendar cells based on the view type and events
func (cm *CalendarManager) generateCalendarCells(view *CalendarView, events []CalendarEvent, timeRange *TimeRange) ([][]CalendarCell, error) {
	switch view.ViewType {
	case ViewTypeMonth:
		return cm.generateMonthCells(view, events, timeRange)
	case ViewTypeWeek:
		return cm.generateWeekCells(view, events, timeRange)
	case ViewTypeDay:
		return cm.generateDayCells(view, events, timeRange)
	default:
		return nil, NewCalendarError(ErrorCodeInvalidFilter, "unsupported view type", fmt.Sprintf("%d", view.ViewType))
	}
}

// generateMonthCells creates a grid of cells for month view
func (cm *CalendarManager) generateMonthCells(view *CalendarView, events []CalendarEvent, timeRange *TimeRange) ([][]CalendarCell, error) {
	// Group events by date
	eventsByDate := cm.groupEventsByDate(events, view.Timezone)

	// Start from the first Sunday of the month view
	start := timeRange.Start
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}

	today := time.Now().In(view.Timezone).Truncate(24 * time.Hour)
	var cells [][]CalendarCell

	// Generate 6 weeks (42 days)
	for week := 0; week < 6; week++ {
		var weekCells []CalendarCell
		for day := 0; day < 7; day++ {
			cellDate := start.AddDate(0, 0, week*7+day)
			dateKey := cellDate.Format("2006-01-02")

			dayEvents := eventsByDate[dateKey]

			// Apply filters
			filteredEvents := cm.applyEventFilter(dayEvents, view.Filter)

			cell := CalendarCell{
				Date:       cellDate,
				Events:     filteredEvents,
				EventCount: len(filteredEvents),
				Density:    float64(len(filteredEvents)),
				IsToday:    cellDate.Equal(today),
				IsSelected: false, // Will be set by UI
			}

			weekCells = append(weekCells, cell)
		}
		cells = append(cells, weekCells)
	}

	return cells, nil
}

// generateWeekCells creates a grid of cells for week view
func (cm *CalendarManager) generateWeekCells(view *CalendarView, events []CalendarEvent, timeRange *TimeRange) ([][]CalendarCell, error) {
	eventsByDate := cm.groupEventsByDate(events, view.Timezone)
	today := time.Now().In(view.Timezone).Truncate(24 * time.Hour)

	var cells [][]CalendarCell
	var weekCells []CalendarCell

	// Generate 7 days
	for day := 0; day < 7; day++ {
		cellDate := timeRange.Start.AddDate(0, 0, day)
		dateKey := cellDate.Format("2006-01-02")

		dayEvents := eventsByDate[dateKey]
		filteredEvents := cm.applyEventFilter(dayEvents, view.Filter)

		cell := CalendarCell{
			Date:       cellDate,
			Events:     filteredEvents,
			EventCount: len(filteredEvents),
			Density:    float64(len(filteredEvents)),
			IsToday:    cellDate.Equal(today),
			IsSelected: false,
		}

		weekCells = append(weekCells, cell)
	}

	cells = append(cells, weekCells)
	return cells, nil
}

// generateDayCells creates hourly cells for day view
func (cm *CalendarManager) generateDayCells(view *CalendarView, events []CalendarEvent, timeRange *TimeRange) ([][]CalendarCell, error) {
	eventsByHour := cm.groupEventsByHour(events, view.Timezone)
	currentHour := time.Now().In(view.Timezone).Truncate(time.Hour)

	var cells [][]CalendarCell

	// Generate 24 hours
	for hour := 0; hour < 24; hour++ {
		cellTime := timeRange.Start.Add(time.Duration(hour) * time.Hour)
		hourKey := cellTime.Format("2006-01-02-15")

		hourEvents := eventsByHour[hourKey]
		filteredEvents := cm.applyEventFilter(hourEvents, view.Filter)

		cell := CalendarCell{
			Date:       cellTime,
			Events:     filteredEvents,
			EventCount: len(filteredEvents),
			Density:    float64(len(filteredEvents)),
			IsToday:    cellTime.Equal(currentHour),
			IsSelected: false,
		}

		// Create single row with one cell per hour
		cells = append(cells, []CalendarCell{cell})
	}

	return cells, nil
}

// groupEventsByDate groups events by date string
func (cm *CalendarManager) groupEventsByDate(events []CalendarEvent, tz *time.Location) map[string][]CalendarEvent {
	grouped := make(map[string][]CalendarEvent)

	for _, event := range events {
		dateKey := event.ScheduledAt.In(tz).Format("2006-01-02")
		grouped[dateKey] = append(grouped[dateKey], event)
	}

	// Sort events within each day by scheduled time
	for dateKey := range grouped {
		sort.Slice(grouped[dateKey], func(i, j int) bool {
			return grouped[dateKey][i].ScheduledAt.Before(grouped[dateKey][j].ScheduledAt)
		})
	}

	return grouped
}

// groupEventsByHour groups events by hour string
func (cm *CalendarManager) groupEventsByHour(events []CalendarEvent, tz *time.Location) map[string][]CalendarEvent {
	grouped := make(map[string][]CalendarEvent)

	for _, event := range events {
		hourKey := event.ScheduledAt.In(tz).Format("2006-01-02-15")
		grouped[hourKey] = append(grouped[hourKey], event)
	}

	// Sort events within each hour by scheduled time
	for hourKey := range grouped {
		sort.Slice(grouped[hourKey], func(i, j int) bool {
			return grouped[hourKey][i].ScheduledAt.Before(grouped[hourKey][j].ScheduledAt)
		})
	}

	return grouped
}

// applyEventFilter applies the event filter to a list of events
func (cm *CalendarManager) applyEventFilter(events []CalendarEvent, filter EventFilter) []CalendarEvent {
	if cm.isEmptyFilter(filter) {
		return events
	}

	var filtered []CalendarEvent

	for _, event := range events {
		if cm.matchesFilter(event, filter) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// isEmptyFilter checks if the filter is empty
func (cm *CalendarManager) isEmptyFilter(filter EventFilter) bool {
	return len(filter.QueueNames) == 0 &&
		len(filter.JobTypes) == 0 &&
		len(filter.Tags) == 0 &&
		len(filter.Statuses) == 0 &&
		filter.SearchText == ""
}

// matchesFilter checks if an event matches the filter criteria
func (cm *CalendarManager) matchesFilter(event CalendarEvent, filter EventFilter) bool {
	// Queue name filter
	if len(filter.QueueNames) > 0 {
		found := false
		for _, queueName := range filter.QueueNames {
			if event.QueueName == queueName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Job type filter
	if len(filter.JobTypes) > 0 {
		found := false
		for _, jobType := range filter.JobTypes {
			if event.JobType == jobType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, eventTag := range event.Tags {
				if eventTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Status filter
	if len(filter.Statuses) > 0 {
		found := false
		for _, status := range filter.Statuses {
			if event.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Search text filter
	if filter.SearchText != "" {
		searchText := strings.ToLower(filter.SearchText)
		if !strings.Contains(strings.ToLower(event.JobType), searchText) &&
			!strings.Contains(strings.ToLower(event.QueueName), searchText) {
			// Check metadata and tags
			found := false
			for key, value := range event.Metadata {
				if strings.Contains(strings.ToLower(key), searchText) ||
					strings.Contains(strings.ToLower(value), searchText) {
					found = true
					break
				}
			}
			if !found {
				for _, tag := range event.Tags {
					if strings.Contains(strings.ToLower(tag), searchText) {
						found = true
						break
					}
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// calculatePeakDensity finds the maximum density across all cells
func (cm *CalendarManager) calculatePeakDensity(cells [][]CalendarCell) float64 {
	maxDensity := 0.0

	for _, row := range cells {
		for _, cell := range row {
			if cell.Density > maxDensity {
				maxDensity = cell.Density
			}
		}
	}

	return maxDensity
}

// RescheduleEvent reschedules an event to a new time
func (cm *CalendarManager) RescheduleEvent(eventID string, newTime time.Time, userID string, reason string) (*RescheduleResult, error) {
	req := &RescheduleRequest{
		EventID: eventID,
		NewTime: newTime,
		Reason:  reason,
		UserID:  userID,
	}

	if err := cm.validator.ValidateRescheduleRequest(req); err != nil {
		return nil, err
	}

	result, err := cm.dataSource.RescheduleEvent(req)
	if err != nil {
		cm.auditor.LogError("reschedule_event", userID, err)
		return nil, err
	}

	cm.auditor.LogSuccess("reschedule_event", userID, map[string]string{
		"event_id": eventID,
		"old_time": result.OldTime.Format(time.RFC3339),
		"new_time": result.NewTime.Format(time.RFC3339),
		"reason":   reason,
	})

	// Invalidate cache
	cm.cache.Invalidate()

	return result, nil
}

// Navigate handles navigation actions within the calendar
func (cm *CalendarManager) Navigate(view *CalendarView, action NavigationAction) error {
	switch action {
	case NavNextPeriod:
		return cm.navigateNext(view)
	case NavPrevPeriod:
		return cm.navigatePrev(view)
	case NavToday:
		return cm.navigateToday(view)
	case NavEnd:
		return cm.navigateEnd(view)
	default:
		return NewCalendarError(ErrorCodeInvalidFilter, "unsupported navigation action", fmt.Sprintf("%d", action))
	}
}

// navigateNext moves to the next period
func (cm *CalendarManager) navigateNext(view *CalendarView) error {
	switch view.ViewType {
	case ViewTypeMonth:
		view.CurrentDate = view.CurrentDate.AddDate(0, 1, 0)
	case ViewTypeWeek:
		view.CurrentDate = view.CurrentDate.AddDate(0, 0, 7)
	case ViewTypeDay:
		view.CurrentDate = view.CurrentDate.AddDate(0, 0, 1)
	}
	return nil
}

// navigatePrev moves to the previous period
func (cm *CalendarManager) navigatePrev(view *CalendarView) error {
	switch view.ViewType {
	case ViewTypeMonth:
		view.CurrentDate = view.CurrentDate.AddDate(0, -1, 0)
	case ViewTypeWeek:
		view.CurrentDate = view.CurrentDate.AddDate(0, 0, -7)
	case ViewTypeDay:
		view.CurrentDate = view.CurrentDate.AddDate(0, 0, -1)
	}
	return nil
}

// navigateToday moves to today
func (cm *CalendarManager) navigateToday(view *CalendarView) error {
	view.CurrentDate = time.Now().In(view.Timezone)
	return nil
}

// navigateEnd moves to a far future date
func (cm *CalendarManager) navigateEnd(view *CalendarView) error {
	view.CurrentDate = time.Now().In(view.Timezone).AddDate(1, 0, 0)
	return nil
}

// GetViewTitle returns a formatted title for the current view
func (cm *CalendarManager) GetViewTitle(view *CalendarView) string {
	current := view.CurrentDate.In(view.Timezone)

	switch view.ViewType {
	case ViewTypeMonth:
		return current.Format("January 2006")
	case ViewTypeWeek:
		weekStart := current.AddDate(0, 0, -int(current.Weekday()))
		weekEnd := weekStart.AddDate(0, 0, 6)
		if weekStart.Month() == weekEnd.Month() {
			return fmt.Sprintf("%s %d-%d, %d",
				weekStart.Month().String()[:3],
				weekStart.Day(),
				weekEnd.Day(),
				weekStart.Year())
		}
		return fmt.Sprintf("%s %d - %s %d, %d",
			weekStart.Month().String()[:3],
			weekStart.Day(),
			weekEnd.Month().String()[:3],
			weekEnd.Day(),
			weekStart.Year())
	case ViewTypeDay:
		return current.Format("Monday, January 2, 2006")
	default:
		return "Calendar"
	}
}