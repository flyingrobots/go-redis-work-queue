# Calendar View API Documentation

## Overview

The Calendar View API provides comprehensive calendar functionality for visualizing and managing scheduled jobs in the go-redis-work-queue system. It supports multiple view types (month, week, day), interactive navigation, event filtering, rescheduling operations, and recurring rule management.

## Features

- **Multiple View Types**: Month, week, and day calendar views
- **Visual Density Mapping**: Heatmap-style visualization showing job density
- **Interactive Navigation**: Keyboard shortcuts and navigation controls
- **Event Filtering**: Filter by queue, job type, status, tags, and search text
- **Rescheduling**: Move events to different times with validation and audit trails
- **Recurring Rules**: Create, update, pause/resume, and delete cron-based recurring jobs
- **Real-time Updates**: Live data with configurable refresh intervals
- **Audit Logging**: Comprehensive logging of all calendar operations
- **Performance Optimized**: Caching, pagination, and efficient data structures

## Core Components

### CalendarManager

The main entry point for calendar operations, providing high-level methods for data retrieval, navigation, and event management.

```go
type CalendarManager struct {
    config     *CalendarConfig
    dataSource DataSource
    cache      *CacheManager
    validator  *Validator
    auditor    *AuditLogger
}
```

### Data Structures

#### CalendarView
```go
type CalendarView struct {
    ViewType    ViewType      `json:"view_type"`    // month, week, day
    CurrentDate time.Time     `json:"current_date"` // Current viewing date
    Timezone    *time.Location `json:"timezone"`    // Display timezone
    Filter      EventFilter   `json:"filter"`       // Active filters
}
```

#### CalendarEvent
```go
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
```

#### RecurringRule
```go
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
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
    NextRun     *time.Time        `json:"next_run"`
}
```

## HTTP API Endpoints

### Calendar Data

#### GET `/calendar/data`

Retrieves calendar data for the specified view configuration.

**Query Parameters:**
- `view` (string): View type - "month", "week", "day" (default: "month")
- `date` (string): Current date in YYYY-MM-DD format (default: today)
- `timezone` (string): Timezone identifier (default: "UTC")
- `queues` (string): Comma-separated list of queue names to filter
- `job_types` (string): Comma-separated list of job types to filter
- `tags` (string): Comma-separated list of tags to filter
- `statuses` (string): Comma-separated list of statuses to filter
- `search` (string): Search text for filtering events

**Example Request:**
```http
GET /calendar/data?view=month&date=2023-06-15&timezone=UTC&queues=high-priority,normal
```

**Response:**
```json
{
  "cells": [
    [
      {
        "date": "2023-06-01T00:00:00Z",
        "events": [
          {
            "id": "event123",
            "queue_name": "high-priority",
            "job_type": "process-order",
            "scheduled_at": "2023-06-01T09:00:00Z",
            "status": 0,
            "priority": 1,
            "metadata": {"order_id": "12345"},
            "tags": ["urgent", "payment"]
          }
        ],
        "event_count": 1,
        "density": 1.0,
        "is_today": false,
        "is_selected": false
      }
    ]
  ],
  "total_events": 45,
  "peak_density": 8.0,
  "time_range": {
    "start": "2023-06-01T00:00:00Z",
    "end": "2023-07-01T00:00:00Z"
  },
  "recurring_rules": [
    {
      "id": "rule456",
      "name": "Daily Report Generation",
      "cron_spec": "0 6 * * *",
      "queue_name": "reports",
      "job_type": "generate-report",
      "timezone": "UTC",
      "is_active": true,
      "is_paused": false,
      "next_run": "2023-06-16T06:00:00Z"
    }
  ]
}
```

#### GET `/calendar/events`

Retrieves events for a specific time window.

**Query Parameters:**
- `from` (string, required): Start time in RFC3339 format
- `till` (string, required): End time in RFC3339 format
- `queue` (string): Filter by specific queue name
- `limit` (int): Maximum number of events to return

**Example Request:**
```http
GET /calendar/events?from=2023-06-01T00:00:00Z&till=2023-06-30T23:59:59Z&queue=high-priority&limit=100
```

**Response:**
```json
{
  "events": [
    {
      "id": "event123",
      "queue_name": "high-priority",
      "job_type": "process-order",
      "scheduled_at": "2023-06-01T09:00:00Z",
      "status": 0,
      "priority": 1,
      "metadata": {"order_id": "12345"},
      "tags": ["urgent", "payment"]
    }
  ],
  "total_count": 45,
  "window": {
    "from": "2023-06-01T00:00:00Z",
    "till": "2023-06-30T23:59:59Z",
    "queue_name": "high-priority",
    "limit": 100
  },
  "has_more": false
}
```

### Event Rescheduling

#### POST `/calendar/reschedule`

Reschedules a single event to a new time.

**Request Body:**
```json
{
  "event_id": "event123",
  "new_time": "2023-06-02T10:00:00Z",
  "new_queue": "normal-priority",
  "reason": "Customer requested delay",
  "user_id": "user456"
}
```

**Response:**
```json
{
  "success": true,
  "old_time": "2023-06-01T09:00:00Z",
  "new_time": "2023-06-02T10:00:00Z",
  "event_id": "event123",
  "message": "Event rescheduled successfully",
  "timestamp": "2023-06-01T14:30:00Z"
}
```

#### POST `/calendar/reschedule/bulk`

Reschedules multiple events in a single operation.

**Request Body:**
```json
[
  {
    "event_id": "event123",
    "new_time": "2023-06-02T10:00:00Z",
    "reason": "Batch reschedule",
    "user_id": "user456"
  },
  {
    "event_id": "event124",
    "new_time": "2023-06-02T11:00:00Z",
    "reason": "Batch reschedule",
    "user_id": "user456"
  }
]
```

**Response:**
```json
{
  "results": [
    {
      "success": true,
      "old_time": "2023-06-01T09:00:00Z",
      "new_time": "2023-06-02T10:00:00Z",
      "event_id": "event123",
      "message": "Event rescheduled successfully",
      "timestamp": "2023-06-01T14:30:00Z"
    },
    {
      "success": false,
      "event_id": "event124",
      "message": "Event not found",
      "timestamp": "2023-06-01T14:30:00Z"
    }
  ],
  "errors": [null, "Event not found"]
}
```

### Recurring Rules Management

#### POST `/calendar/rules`

Creates a new recurring rule.

**Request Body:**
```json
{
  "name": "Daily Report Generation",
  "cron_spec": "0 6 * * *",
  "queue_name": "reports",
  "job_type": "generate-report",
  "timezone": "UTC",
  "max_in_flight": 1,
  "jitter": "300s",
  "metadata": {
    "report_type": "daily_summary",
    "notification_email": "admin@company.com"
  }
}
```

**Response:**
```json
{
  "id": "rule789",
  "name": "Daily Report Generation",
  "cron_spec": "0 6 * * *",
  "queue_name": "reports",
  "job_type": "generate-report",
  "timezone": "UTC",
  "is_active": true,
  "is_paused": false,
  "max_in_flight": 1,
  "jitter": "300s",
  "metadata": {
    "report_type": "daily_summary",
    "notification_email": "admin@company.com"
  },
  "created_at": "2023-06-01T14:30:00Z",
  "updated_at": "2023-06-01T14:30:00Z",
  "next_run": "2023-06-02T06:00:00Z"
}
```

#### GET `/calendar/rules`

Lists recurring rules with optional filtering.

**Query Parameters:**
- `queues` (string): Comma-separated queue names
- `job_types` (string): Comma-separated job types
- `active` (bool): Filter by active status
- `paused` (bool): Filter by paused status

**Example Request:**
```http
GET /calendar/rules?queues=reports&active=true&paused=false
```

**Response:**
```json
{
  "rules": [
    {
      "id": "rule789",
      "name": "Daily Report Generation",
      "cron_spec": "0 6 * * *",
      "queue_name": "reports",
      "job_type": "generate-report",
      "timezone": "UTC",
      "is_active": true,
      "is_paused": false,
      "max_in_flight": 1,
      "jitter": "300s",
      "created_at": "2023-06-01T14:30:00Z",
      "updated_at": "2023-06-01T14:30:00Z",
      "next_run": "2023-06-02T06:00:00Z"
    }
  ],
  "count": 1
}
```

#### GET `/calendar/rules/{id}`

Retrieves a specific recurring rule.

**Response:**
```json
{
  "id": "rule789",
  "name": "Daily Report Generation",
  "cron_spec": "0 6 * * *",
  "queue_name": "reports",
  "job_type": "generate-report",
  "timezone": "UTC",
  "is_active": true,
  "is_paused": false,
  "max_in_flight": 1,
  "jitter": "300s",
  "metadata": {
    "report_type": "daily_summary",
    "notification_email": "admin@company.com"
  },
  "created_at": "2023-06-01T14:30:00Z",
  "updated_at": "2023-06-01T14:30:00Z",
  "next_run": "2023-06-02T06:00:00Z"
}
```

#### PUT `/calendar/rules/{id}`

Updates an existing recurring rule.

**Request Body:**
```json
{
  "name": "Updated Daily Report",
  "cron_spec": "0 7 * * *",
  "is_active": true,
  "is_paused": false,
  "max_in_flight": 2,
  "metadata": {
    "report_type": "enhanced_summary",
    "notification_email": "reports@company.com"
  }
}
```

**Response:**
```json
{
  "id": "rule789",
  "name": "Updated Daily Report",
  "cron_spec": "0 7 * * *",
  "queue_name": "reports",
  "job_type": "generate-report",
  "timezone": "UTC",
  "is_active": true,
  "is_paused": false,
  "max_in_flight": 2,
  "jitter": "300s",
  "metadata": {
    "report_type": "enhanced_summary",
    "notification_email": "reports@company.com"
  },
  "created_at": "2023-06-01T14:30:00Z",
  "updated_at": "2023-06-01T15:45:00Z",
  "next_run": "2023-06-02T07:00:00Z"
}
```

#### DELETE `/calendar/rules/{id}`

Deletes a recurring rule.

**Response:**
```http
204 No Content
```

#### POST `/calendar/rules/{id}/pause`

Pauses a recurring rule.

**Response:**
```json
{
  "success": true,
  "message": "Rule paused successfully"
}
```

#### POST `/calendar/rules/{id}/resume`

Resumes a paused recurring rule.

**Response:**
```json
{
  "success": true,
  "message": "Rule resumed successfully"
}
```

### Configuration

#### GET `/calendar/config`

Returns the current calendar configuration.

**Response:**
```json
{
  "default_view": 0,
  "default_timezone": "UTC",
  "max_events_per_cell": 100,
  "refresh_interval": "30s",
  "key_bindings": [
    {
      "key": "h",
      "action": 2,
      "description": "Move left"
    },
    {
      "key": "j",
      "action": 1,
      "description": "Move down"
    }
  ],
  "colors": {
    "background": "#1e1e1e",
    "border": "#444444",
    "text": "#ffffff",
    "today": "#ffff00",
    "selected": "#0078d4"
  }
}
```

#### GET `/calendar/timezones`

Returns the list of supported timezones.

**Response:**
```json
{
  "timezones": [
    "UTC",
    "America/New_York",
    "America/Chicago",
    "America/Denver",
    "America/Los_Angeles",
    "Europe/London",
    "Europe/Paris",
    "Europe/Berlin",
    "Asia/Tokyo",
    "Asia/Shanghai",
    "Asia/Kolkata",
    "Australia/Sydney"
  ]
}
```

## Error Handling

The API uses structured error responses with specific error codes:

```json
{
  "error": "Event not found",
  "message": "Event with ID 'invalid-id' does not exist",
  "code": 3,
  "details": "Event ID: invalid-id"
}
```

### Error Codes

| Code | Type | Description |
|------|------|-------------|
| 0 | ErrorCodeUnknown | Unknown error |
| 1 | ErrorCodeInvalidTimeRange | Invalid time range |
| 2 | ErrorCodeInvalidCronSpec | Invalid cron specification |
| 3 | ErrorCodeTimezoneNotFound | Timezone not found |
| 4 | ErrorCodeEventNotFound | Event not found |
| 5 | ErrorCodeRuleNotFound | Recurring rule not found |
| 6 | ErrorCodeRescheduleConflict | Reschedule conflict |
| 7 | ErrorCodeRuleValidation | Rule validation failed |
| 8 | ErrorCodeDatabaseError | Database operation failed |
| 9 | ErrorCodePermissionDenied | Permission denied |
| 10 | ErrorCodeRateLimited | Rate limited |
| 11 | ErrorCodeInvalidFilter | Invalid filter parameters |
| 12 | ErrorCodeMaxEventsExceeded | Maximum events exceeded |

## Navigation and Keyboard Shortcuts

The calendar supports various navigation actions:

### Navigation Actions

| Action | Key | Description |
|--------|-----|-------------|
| NavUp | k, ↑ | Move cursor up |
| NavDown | j, ↓ | Move cursor down |
| NavLeft | h, ← | Move cursor left |
| NavRight | l, → | Move cursor right |
| NavNextPeriod | ] | Next period (month/week/day) |
| NavPrevPeriod | [ | Previous period |
| NavToday | g | Go to today |
| NavEnd | G | Go to end/future |
| NavSelect | Space | Select/deselect current cell |
| NavInspect | Enter | Inspect selected event |
| NavReschedule | r | Reschedule selected event |
| NavPause | p | Pause/resume recurring rule |
| NavFilter | / | Open filter dialog |

### View Types

| Type | Value | Description |
|------|-------|-------------|
| ViewTypeMonth | 0 | Monthly calendar view |
| ViewTypeWeek | 1 | Weekly calendar view |
| ViewTypeDay | 2 | Daily hourly view |

### Event Statuses

| Status | Value | Description |
|--------|-------|-------------|
| StatusScheduled | 0 | Event is scheduled |
| StatusRunning | 1 | Event is currently running |
| StatusCompleted | 2 | Event completed successfully |
| StatusFailed | 3 | Event failed |
| StatusCanceled | 4 | Event was canceled |

## Usage Examples

### Go Integration

```go
package main

import (
    "fmt"
    "time"

    "github.com/flyingrobots/go-redis-work-queue/internal/calendar-view"
)

func main() {
    // Create calendar manager
    config := calendarview.DefaultConfig()
    dataSource := &MyDataSource{} // Implement DataSource interface

    manager, err := calendarview.NewCalendarManager(config, dataSource)
    if err != nil {
        panic(err)
    }

    // Create calendar view
    view := &calendarview.CalendarView{
        ViewType:    calendarview.ViewTypeMonth,
        CurrentDate: time.Now(),
        Timezone:    time.UTC,
        Filter: calendarview.EventFilter{
            QueueNames: []string{"high-priority"},
            Statuses:   []calendarview.EventStatus{calendarview.StatusScheduled},
        },
    }

    // Get calendar data
    data, err := manager.GetCalendarData(view)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Total events: %d\n", data.TotalEvents)
    fmt.Printf("Peak density: %.2f\n", data.PeakDensity)

    // Navigate to next month
    err = manager.Navigate(view, calendarview.NavNextPeriod)
    if err != nil {
        panic(err)
    }

    fmt.Printf("New title: %s\n", manager.GetViewTitle(view))
}
```

### HTTP Client Example

```javascript
// Fetch monthly calendar data
async function getCalendarData() {
    const params = new URLSearchParams({
        view: 'month',
        date: '2023-06-15',
        timezone: 'UTC',
        queues: 'high-priority,normal'
    });

    const response = await fetch(`/calendar/data?${params}`);
    const data = await response.json();

    console.log(`Total events: ${data.total_events}`);
    console.log(`Peak density: ${data.peak_density}`);

    return data;
}

// Reschedule an event
async function rescheduleEvent(eventId, newTime, reason) {
    const response = await fetch('/calendar/reschedule', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-User-ID': 'user123'
        },
        body: JSON.stringify({
            event_id: eventId,
            new_time: newTime,
            reason: reason,
            user_id: 'user123'
        })
    });

    const result = await response.json();

    if (result.success) {
        console.log('Event rescheduled successfully');
    } else {
        console.error('Reschedule failed:', result.message);
    }

    return result;
}

// Create a recurring rule
async function createRecurringRule(rule) {
    const response = await fetch('/calendar/rules', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-User-ID': 'user123'
        },
        body: JSON.stringify(rule)
    });

    const result = await response.json();
    console.log('Rule created:', result.id);

    return result;
}
```

## Performance Considerations

### Caching Strategy

The calendar implementation includes intelligent caching:

- **Calendar Data**: Cached by view configuration (type, date, timezone, filters)
- **Events**: Cached by time window and queue filter
- **Recurring Rules**: Cached by filter criteria
- **TTL**: Configurable cache timeout (default: 5 minutes)
- **Invalidation**: Automatic cache invalidation on data changes

### Optimization Tips

1. **Use appropriate time ranges**: Avoid querying excessively large time windows
2. **Apply filters**: Use queue, job type, and status filters to reduce data transfer
3. **Leverage caching**: Identical requests will be served from cache
4. **Batch operations**: Use bulk reschedule for multiple events
5. **Pagination**: Use limit parameters for large result sets

### Performance Metrics

- **Calendar Data Query**: Typically < 100ms for monthly views
- **Event Rescheduling**: < 50ms for single events
- **Rule Operations**: < 30ms for CRUD operations
- **Cache Hit Rate**: 80-90% for typical usage patterns
- **Memory Usage**: ~1-2MB per 1000 cached events

## Security

### Authentication

All endpoints require proper authentication. Include user identification in requests:

```http
X-User-ID: user123
Authorization: Bearer <jwt-token>
```

### Authorization

Access control is enforced at multiple levels:

- **Queue Access**: Users can only view/modify events in authorized queues
- **Rule Management**: Requires elevated permissions
- **Audit Logging**: All operations are logged with user context
- **Rate Limiting**: Configurable rate limits per user/endpoint

### Input Validation

All inputs are validated:

- **Time Ranges**: Must be reasonable (max 1 year)
- **Cron Expressions**: Validated for safety (min 1-minute intervals)
- **User Input**: Sanitized to prevent injection attacks
- **File Uploads**: Not supported to avoid security risks

## Monitoring and Observability

### Metrics

The calendar system exposes various metrics:

```go
// Example metrics
calendarview_requests_total{endpoint="/calendar/data", status="success"}
calendarview_request_duration_seconds{endpoint="/calendar/data"}
calendarview_cache_hits_total{type="calendar_data"}
calendarview_cache_misses_total{type="calendar_data"}
calendarview_events_rescheduled_total{queue="high-priority"}
calendarview_rules_created_total
```

### Health Checks

Health check endpoint provides system status:

```http
GET /calendar/health
```

```json
{
  "status": "healthy",
  "cache_status": "operational",
  "database_status": "connected",
  "memory_usage": "45MB",
  "active_sessions": 23
}
```

### Audit Trails

All calendar operations are audited:

- **User Actions**: View access, reschedules, rule changes
- **System Events**: Cache invalidations, errors, performance issues
- **Data Changes**: Before/after values for all modifications
- **Access Patterns**: Usage statistics and performance metrics

## Troubleshooting

### Common Issues

#### Calendar Data Not Loading

1. Check timezone configuration
2. Verify time range parameters
3. Confirm queue access permissions
4. Review filter parameters

#### Reschedule Failures

1. Validate new time is not in the past
2. Check event exists and is modifiable
3. Verify no scheduling conflicts
4. Confirm user has reschedule permissions

#### Performance Issues

1. Reduce time range scope
2. Apply more specific filters
3. Check cache hit rates
4. Monitor database query performance

### Debug Endpoints

Debug information is available at:

```http
GET /calendar/debug/cache-stats
GET /calendar/debug/audit-summary
GET /calendar/debug/performance-metrics
```

This comprehensive API documentation provides everything needed to integrate and use the Calendar View functionality effectively.
