package calendarview

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDataSource is a mock implementation of DataSource for testing
type MockDataSource struct {
	mock.Mock
}

func (m *MockDataSource) GetEvents(window ScheduleWindow) (*ScheduleResponse, error) {
	args := m.Called(window)
	return args.Get(0).(*ScheduleResponse), args.Error(1)
}

func (m *MockDataSource) GetRecurringRules(filter RuleFilter) ([]RecurringRule, error) {
	args := m.Called(filter)
	return args.Get(0).([]RecurringRule), args.Error(1)
}

func (m *MockDataSource) RescheduleEvent(req *RescheduleRequest) (*RescheduleResult, error) {
	args := m.Called(req)
	return args.Get(0).(*RescheduleResult), args.Error(1)
}

func (m *MockDataSource) CreateRule(req *RuleCreateRequest) (*RecurringRule, error) {
	args := m.Called(req)
	return args.Get(0).(*RecurringRule), args.Error(1)
}

func (m *MockDataSource) UpdateRule(req *RuleUpdateRequest) (*RecurringRule, error) {
	args := m.Called(req)
	return args.Get(0).(*RecurringRule), args.Error(1)
}

func (m *MockDataSource) DeleteRule(ruleID string) error {
	args := m.Called(ruleID)
	return args.Error(0)
}

func (m *MockDataSource) PauseRule(ruleID string) error {
	args := m.Called(ruleID)
	return args.Error(0)
}

func (m *MockDataSource) ResumeRule(ruleID string) error {
	args := m.Called(ruleID)
	return args.Error(0)
}

func TestNewCalendarManager(t *testing.T) {
	mockDS := &MockDataSource{}
	config := DefaultConfig()

	cm, err := NewCalendarManager(config, mockDS)

	assert.NoError(t, err)
	assert.NotNil(t, cm)
	assert.Equal(t, config, cm.config)
	assert.Equal(t, mockDS, cm.dataSource)
	assert.NotNil(t, cm.cache)
	assert.NotNil(t, cm.validator)
	assert.NotNil(t, cm.auditor)
}

func TestCalendarManager_GetCalendarData_MonthView(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	// Create test view
	now := time.Now()
	view := &CalendarView{
		ViewType:    ViewTypeMonth,
		CurrentDate: now,
		Timezone:    time.UTC,
		Filter:      EventFilter{},
	}

	// Create test events
	events := []CalendarEvent{
		{
			ID:          "event1",
			QueueName:   "test-queue",
			JobType:     "test-job",
			ScheduledAt: now,
			Status:      StatusScheduled,
			Priority:    1,
		},
		{
			ID:          "event2",
			QueueName:   "test-queue",
			JobType:     "test-job",
			ScheduledAt: now.AddDate(0, 0, 1),
			Status:      StatusScheduled,
			Priority:    2,
		},
	}

	response := &ScheduleResponse{
		Events:     events,
		TotalCount: len(events),
		HasMore:    false,
	}

	rules := []RecurringRule{
		{
			ID:       "rule1",
			Name:     "Test Rule",
			CronSpec: "0 0 * * *",
			IsActive: true,
			IsPaused: false,
		},
	}

	// Set up mock expectations
	mockDS.On("GetEvents", mock.AnythingOfType("ScheduleWindow")).Return(response, nil)
	mockDS.On("GetRecurringRules", mock.AnythingOfType("RuleFilter")).Return(rules, nil)

	// Execute test
	data, err := cm.GetCalendarData(view)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, len(events), data.TotalEvents)
	assert.Equal(t, len(rules), len(data.RecurringRules))
	assert.NotEmpty(t, data.Cells)
	assert.Equal(t, 6, len(data.Cells)) // 6 weeks for month view

	// Verify mock calls
	mockDS.AssertExpectations(t)
}

func TestCalendarManager_GetCalendarData_WeekView(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	view := &CalendarView{
		ViewType:    ViewTypeWeek,
		CurrentDate: time.Now(),
		Timezone:    time.UTC,
		Filter:      EventFilter{},
	}

	response := &ScheduleResponse{
		Events:     []CalendarEvent{},
		TotalCount: 0,
		HasMore:    false,
	}

	mockDS.On("GetEvents", mock.AnythingOfType("ScheduleWindow")).Return(response, nil)
	mockDS.On("GetRecurringRules", mock.AnythingOfType("RuleFilter")).Return([]RecurringRule{}, nil)

	data, err := cm.GetCalendarData(view)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Cells)) // 1 week row
	assert.Equal(t, 7, len(data.Cells[0])) // 7 days

	mockDS.AssertExpectations(t)
}

func TestCalendarManager_GetCalendarData_DayView(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	view := &CalendarView{
		ViewType:    ViewTypeDay,
		CurrentDate: time.Now(),
		Timezone:    time.UTC,
		Filter:      EventFilter{},
	}

	response := &ScheduleResponse{
		Events:     []CalendarEvent{},
		TotalCount: 0,
		HasMore:    false,
	}

	mockDS.On("GetEvents", mock.AnythingOfType("ScheduleWindow")).Return(response, nil)
	mockDS.On("GetRecurringRules", mock.AnythingOfType("RuleFilter")).Return([]RecurringRule{}, nil)

	data, err := cm.GetCalendarData(view)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 24, len(data.Cells)) // 24 hours
	for _, hourCell := range data.Cells {
		assert.Equal(t, 1, len(hourCell)) // 1 cell per hour
	}

	mockDS.AssertExpectations(t)
}

func TestCalendarManager_RescheduleEvent(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	eventID := "test-event"
	newTime := time.Now().Add(24 * time.Hour)
	userID := "test-user"
	reason := "Test reschedule"

	expectedResult := &RescheduleResult{
		Success:   true,
		EventID:   eventID,
		OldTime:   time.Now(),
		NewTime:   newTime,
		Message:   "Event rescheduled successfully",
		Timestamp: time.Now(),
	}

	mockDS.On("RescheduleEvent", mock.AnythingOfType("*calendarview.RescheduleRequest")).Return(expectedResult, nil)

	result, err := cm.RescheduleEvent(eventID, newTime, userID, reason)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResult.Success, result.Success)
	assert.Equal(t, expectedResult.EventID, result.EventID)

	mockDS.AssertExpectations(t)
}

func TestCalendarManager_Navigate(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	baseDate := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		viewType       ViewType
		action         NavigationAction
		expectedChange time.Duration
	}{
		{"Month Next", ViewTypeMonth, NavNextPeriod, 30 * 24 * time.Hour},
		{"Month Prev", ViewTypeMonth, NavPrevPeriod, -30 * 24 * time.Hour},
		{"Week Next", ViewTypeWeek, NavNextPeriod, 7 * 24 * time.Hour},
		{"Week Prev", ViewTypeWeek, NavPrevPeriod, -7 * 24 * time.Hour},
		{"Day Next", ViewTypeDay, NavNextPeriod, 24 * time.Hour},
		{"Day Prev", ViewTypeDay, NavPrevPeriod, -24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			view := &CalendarView{
				ViewType:    tc.viewType,
				CurrentDate: baseDate,
				Timezone:    time.UTC,
			}

			err := cm.Navigate(view, tc.action)
			assert.NoError(t, err)

			// For month navigation, we check if the month changed
			if tc.viewType == ViewTypeMonth {
				if tc.action == NavNextPeriod {
					assert.True(t, view.CurrentDate.After(baseDate))
				} else if tc.action == NavPrevPeriod {
					assert.True(t, view.CurrentDate.Before(baseDate))
				}
			} else {
				// For week and day, check exact duration
				actualChange := view.CurrentDate.Sub(baseDate)
				assert.Equal(t, tc.expectedChange, actualChange)
			}
		})
	}
}

func TestCalendarManager_NavigateToday(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	view := &CalendarView{
		ViewType:    ViewTypeMonth,
		CurrentDate: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
		Timezone:    time.UTC,
	}

	err := cm.Navigate(view, NavToday)
	assert.NoError(t, err)

	// Check that current date is now close to today
	now := time.Now().In(time.UTC)
	diff := view.CurrentDate.Sub(now)
	assert.True(t, diff < time.Hour && diff > -time.Hour)
}

func TestCalendarManager_GetViewTitle(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	baseDate := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		viewType ViewType
		expected string
	}{
		{"Month view", ViewTypeMonth, "June 2023"},
		{"Day view", ViewTypeDay, "Thursday, June 15, 2023"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			view := &CalendarView{
				ViewType:    tc.viewType,
				CurrentDate: baseDate,
				Timezone:    time.UTC,
			}

			title := cm.GetViewTitle(view)
			assert.Equal(t, tc.expected, title)
		})
	}
}

func TestCalendarManager_GroupEventsByDate(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	baseDate := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
	events := []CalendarEvent{
		{
			ID:          "event1",
			ScheduledAt: baseDate,
		},
		{
			ID:          "event2",
			ScheduledAt: baseDate.Add(2 * time.Hour),
		},
		{
			ID:          "event3",
			ScheduledAt: baseDate.AddDate(0, 0, 1),
		},
	}

	grouped := cm.groupEventsByDate(events, time.UTC)

	assert.Equal(t, 2, len(grouped))
	assert.Equal(t, 2, len(grouped["2023-06-15"]))
	assert.Equal(t, 1, len(grouped["2023-06-16"]))

	// Check sorting within each day
	day1Events := grouped["2023-06-15"]
	assert.Equal(t, "event1", day1Events[0].ID)
	assert.Equal(t, "event2", day1Events[1].ID)
}

func TestCalendarManager_ApplyEventFilter(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	events := []CalendarEvent{
		{
			ID:        "event1",
			QueueName: "queue1",
			JobType:   "job1",
			Status:    StatusScheduled,
			Tags:      []string{"tag1", "tag2"},
		},
		{
			ID:        "event2",
			QueueName: "queue2",
			JobType:   "job2",
			Status:    StatusRunning,
			Tags:      []string{"tag2", "tag3"},
		},
		{
			ID:        "event3",
			QueueName: "queue1",
			JobType:   "job1",
			Status:    StatusCompleted,
			Tags:      []string{"tag1"},
		},
	}

	testCases := []struct {
		name           string
		filter         EventFilter
		expectedCount  int
		expectedIDs    []string
	}{
		{
			name:          "No filter",
			filter:        EventFilter{},
			expectedCount: 3,
			expectedIDs:   []string{"event1", "event2", "event3"},
		},
		{
			name:          "Queue filter",
			filter:        EventFilter{QueueNames: []string{"queue1"}},
			expectedCount: 2,
			expectedIDs:   []string{"event1", "event3"},
		},
		{
			name:          "Job type filter",
			filter:        EventFilter{JobTypes: []string{"job2"}},
			expectedCount: 1,
			expectedIDs:   []string{"event2"},
		},
		{
			name:          "Status filter",
			filter:        EventFilter{Statuses: []EventStatus{StatusScheduled}},
			expectedCount: 1,
			expectedIDs:   []string{"event1"},
		},
		{
			name:          "Tag filter",
			filter:        EventFilter{Tags: []string{"tag2"}},
			expectedCount: 2,
			expectedIDs:   []string{"event1", "event2"},
		},
		{
			name:          "Multiple filters",
			filter:        EventFilter{QueueNames: []string{"queue1"}, Tags: []string{"tag1"}},
			expectedCount: 2,
			expectedIDs:   []string{"event1", "event3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filtered := cm.applyEventFilter(events, tc.filter)
			assert.Equal(t, tc.expectedCount, len(filtered))

			actualIDs := make([]string, len(filtered))
			for i, event := range filtered {
				actualIDs[i] = event.ID
			}

			for _, expectedID := range tc.expectedIDs {
				assert.Contains(t, actualIDs, expectedID)
			}
		})
	}
}

func TestCalendarManager_CalculatePeakDensity(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	cells := [][]CalendarCell{
		{
			{Density: 5.0},
			{Density: 10.0},
			{Density: 3.0},
		},
		{
			{Density: 15.0},
			{Density: 8.0},
			{Density: 12.0},
		},
	}

	peak := cm.calculatePeakDensity(cells)
	assert.Equal(t, 15.0, peak)
}

func TestCalendarManager_GenerateTimeRange(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	baseDate := time.Date(2023, 6, 15, 12, 30, 45, 0, time.UTC)

	testCases := []struct {
		name         string
		viewType     ViewType
		expectedStart string
		expectedEnd   string
	}{
		{
			name:          "Month view",
			viewType:      ViewTypeMonth,
			expectedStart: "2023-06-01T00:00:00Z",
			expectedEnd:   "2023-07-01T00:00:00Z",
		},
		{
			name:          "Day view",
			viewType:      ViewTypeDay,
			expectedStart: "2023-06-15T00:00:00Z",
			expectedEnd:   "2023-06-16T00:00:00Z",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			view := &CalendarView{
				ViewType:    tc.viewType,
				CurrentDate: baseDate,
				Timezone:    time.UTC,
			}

			timeRange, err := cm.generateTimeRange(view)
			assert.NoError(t, err)
			assert.NotNil(t, timeRange)
			assert.Equal(t, tc.expectedStart, timeRange.Start.Format(time.RFC3339))
			assert.Equal(t, tc.expectedEnd, timeRange.End.Format(time.RFC3339))
		})
	}
}

func TestCalendarManager_InvalidViewType(t *testing.T) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	view := &CalendarView{
		ViewType:    ViewType(999), // Invalid view type
		CurrentDate: time.Now(),
		Timezone:    time.UTC,
	}

	_, err := cm.GetCalendarData(view)
	assert.Error(t, err)
	assert.IsType(t, &CalendarError{}, err)
}

// Benchmark tests

func BenchmarkCalendarManager_GetCalendarData(b *testing.B) {
	mockDS := &MockDataSource{}
	cm, _ := NewCalendarManager(DefaultConfig(), mockDS)

	view := &CalendarView{
		ViewType:    ViewTypeMonth,
		CurrentDate: time.Now(),
		Timezone:    time.UTC,
		Filter:      EventFilter{},
	}

	// Create test data
	events := make([]CalendarEvent, 100)
	for i := 0; i < 100; i++ {
		events[i] = CalendarEvent{
			ID:          fmt.Sprintf("event%d", i),
			QueueName:   "test-queue",
			JobType:     "test-job",
			ScheduledAt: time.Now().Add(time.Duration(i) * time.Hour),
			Status:      StatusScheduled,
		}
	}

	response := &ScheduleResponse{
		Events:     events,
		TotalCount: len(events),
		HasMore:    false,
	}

	mockDS.On("GetEvents", mock.AnythingOfType("ScheduleWindow")).Return(response, nil)
	mockDS.On("GetRecurringRules", mock.AnythingOfType("RuleFilter")).Return([]RecurringRule{}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cm.GetCalendarData(view)
		if err != nil {
			b.Fatal(err)
		}
	}
}