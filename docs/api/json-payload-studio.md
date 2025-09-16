# JSON Payload Studio API Documentation

## Overview

The JSON Payload Studio is a comprehensive in-TUI JSON editor for authoring, validating, and enqueuing job payloads. It provides advanced features including templates, snippets, schema validation, scheduling, and diff/peek integration for rapid iteration.

## Features

- **JSON Editing**: Full-featured JSON editor with syntax highlighting, bracket matching, and auto-formatting
- **Validation**: Real-time JSON validation with error highlighting (line/column numbers)
- **Schema Support**: JSON Schema (draft 7) validation with auto-completion
- **Templates**: Load and apply templates with variable substitution
- **Snippets**: Quick insertion of common patterns with expansion
- **Job Enqueueing**: Direct job submission to Redis queues with scheduling options
- **Diff/Peek**: Compare payloads and preview enqueued jobs
- **Safety Features**: Payload size limits, secret stripping, and confirmation prompts

## API Endpoints

### Validation

#### POST /api/json-studio/validate
Validates JSON content with optional schema validation.

**Request:**
```json
{
  "content": "{ \"name\": \"test\" }",
  "schema_id": "user-schema",  // Optional
  "schema": { ... }            // Optional inline schema
}
```

**Response:**
```json
{
  "valid": true,
  "errors": [],
  "warnings": [],
  "info": [],
  "stats": {
    "lines": 1,
    "characters": 18,
    "keys": 1,
    "max_depth": 1,
    "object_count": 1,
    "string_count": 1
  }
}
```

### Formatting

#### POST /api/json-studio/format
Formats JSON content with proper indentation.

**Request:**
```json
{
  "content": "{\"name\":\"test\",\"value\":123}",
  "indent": "  "  // Optional, defaults to 2 spaces
}
```

**Response:**
```json
{
  "formatted": "{\n  \"name\": \"test\",\n  \"value\": 123\n}"
}
```

### Templates

#### GET /api/json-studio/templates
Lists all available templates or searches with filters.

**Query Parameters:**
- `search`: Search query string
- `categories`: Comma-separated category list
- `tags`: Comma-separated tag list

**Response:**
```json
[
  {
    "id": "job-template-1",
    "name": "Worker Job Template",
    "description": "Standard worker job template",
    "category": "jobs",
    "tags": ["worker", "queue"],
    "content": { ... },
    "variables": [ ... ],
    "created_at": "2025-01-14T12:00:00Z",
    "updated_at": "2025-01-14T12:00:00Z"
  }
]
```

#### POST /api/json-studio/templates
Saves a new template or updates an existing one.

**Request:**
```json
{
  "id": "custom-template",
  "name": "Custom Template",
  "description": "My custom template",
  "category": "custom",
  "tags": ["custom", "test"],
  "content": {
    "type": "{{job_type}}",
    "data": "{{payload}}"
  },
  "variables": [
    {
      "name": "job_type",
      "type": "string",
      "required": true,
      "default_value": "process"
    }
  ]
}
```

#### DELETE /api/json-studio/templates?id={template_id}
Deletes a template by ID.

#### POST /api/json-studio/templates/apply
Applies a template with variable substitution.

**Request:**
```json
{
  "template_id": "job-template-1",
  "variables": {
    "job_type": "process",
    "priority": 5
  }
}
```

**Response:**
```json
{
  "result": {
    "type": "process",
    "priority": 5,
    "timestamp": "2025-01-14T12:00:00Z",
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### Sessions

#### POST /api/json-studio/sessions
Creates a new editing session.

**Response:**
```json
{
  "id": "session-123",
  "started_at": "2025-01-14T12:00:00Z",
  "last_activity": "2025-01-14T12:00:00Z",
  "editor_state": {
    "content": "",
    "cursor_line": 1,
    "cursor_column": 1,
    "errors": [],
    "warnings": [],
    "modified": false
  }
}
```

#### GET /api/json-studio/sessions?id={session_id}
Gets session information.

#### PUT /api/json-studio/sessions?id={session_id}
Updates session editor state.

**Request:**
```json
{
  "content": "{ \"updated\": true }",
  "cursor_line": 1,
  "cursor_column": 10,
  "modified": true
}
```

#### DELETE /api/json-studio/sessions?id={session_id}
Deletes a session.

### Job Enqueueing

#### POST /api/json-studio/enqueue
Enqueues a job with the current session's payload.

**Request:**
```json
{
  "session_id": "session-123",
  "options": {
    "queue": "worker-queue",
    "count": 1,
    "priority": 5,
    "delay": 60000000000,      // Nanoseconds (60 seconds)
    "run_at": "2025-01-14T15:00:00Z",  // Optional scheduled time
    "cron_spec": "0 */5 * * *",        // Optional cron schedule
    "max_retries": 3,
    "metadata": {
      "source": "json-studio",
      "user": "admin"
    }
  }
}
```

**Response:**
```json
{
  "job_ids": ["job-456"],
  "queue": "worker-queue",
  "count": 1,
  "priority": 5,
  "payload": { ... },
  "payload_size": 256,
  "enqueued_at": "2025-01-14T12:00:00Z"
}
```

### Auto-Completion

#### POST /api/json-studio/completions
Gets auto-completion suggestions for the current context.

**Request:**
```json
{
  "context": "{ \"na",
  "position": {
    "line": 1,
    "column": 5
  },
  "schema_id": "user-schema"  // Optional
}
```

**Response:**
```json
[
  {
    "label": "name",
    "kind": "property",
    "detail": "User's name",
    "insert_text": "\"name\": ",
    "preselect": false
  }
]
```

### Diff Comparison

#### POST /api/json-studio/diff
Compares two JSON payloads and returns differences.

**Request:**
```json
{
  "old": {
    "name": "John",
    "age": 30
  },
  "new": {
    "name": "John",
    "age": 31,
    "city": "New York"
  }
}
```

**Response:**
```json
{
  "has_changes": true,
  "added": [
    {
      "path": "$.city",
      "type": "added",
      "new_value": "New York"
    }
  ],
  "modified": [
    {
      "path": "$.age",
      "type": "modified",
      "old_value": 30,
      "new_value": 31
    }
  ],
  "removed": [],
  "summary": "1 field added, 1 field modified"
}
```

### Snippets

#### GET /api/json-studio/snippets
Lists all available snippets.

**Response:**
```json
[
  {
    "id": "uuid-snippet",
    "name": "UUID Generator",
    "trigger": "uuid",
    "description": "Generates a UUID",
    "content": "{{uuid}}",
    "category": "generators"
  }
]
```

#### POST /api/json-studio/snippets
Expands a snippet with variable substitution.

**Request:**
```json
{
  "trigger": "timestamp",
  "variables": {
    "format": "ISO8601"
  }
}
```

**Response:**
```json
{
  "expanded": "2025-01-14T12:00:00Z"
}
```

### History Management

#### POST /api/json-studio/history
Performs undo or redo operations.

**Request:**
```json
{
  "session_id": "session-123",
  "action": "undo"  // or "redo"
}
```

### Preview

#### POST /api/json-studio/preview
Generates a preview of JSON content.

**Request:**
```json
{
  "content": { ... },
  "max_lines": 20,
  "truncate": true
}
```

**Response:**
```json
{
  "content": { ... },
  "formatted": "{\n  \"preview\": \"...\"\n}",
  "size": 1024,
  "line_count": 20,
  "truncated": false,
  "valid": true
}
```

## Configuration

The JSON Payload Studio can be configured through environment variables or a configuration file.

### Environment Variables

- `JSON_STUDIO_THEME`: Editor theme (dark/light)
- `JSON_STUDIO_MAX_PAYLOAD_SIZE`: Maximum payload size in bytes
- `JSON_STUDIO_MAX_FIELDS`: Maximum number of fields
- `JSON_STUDIO_MAX_DEPTH`: Maximum nesting depth
- `JSON_STUDIO_STRIP_SECRETS`: Enable secret stripping (true/false)
- `JSON_STUDIO_REQUIRE_CONFIRM`: Require confirmation before enqueue (true/false)
- `JSON_STUDIO_AUTO_SAVE`: Enable auto-save (true/false)

### Configuration File

```json
{
  "editor_theme": "dark",
  "syntax_highlight": true,
  "line_numbers": true,
  "auto_format": false,
  "bracket_matching": true,
  "auto_complete": true,
  "validate_on_type": true,
  "templates_path": "config/templates",
  "schemas_path": "config/schemas",
  "max_payload_size": 10485760,
  "max_field_count": 10000,
  "max_nesting_depth": 50,
  "strip_secrets": true,
  "secret_patterns": [
    "password",
    "secret",
    "token",
    "api_key"
  ],
  "require_confirm": true,
  "history_size": 100,
  "auto_save": true,
  "auto_save_interval": 30000000000
}
```

## Error Handling

All API endpoints return structured error responses:

```json
{
  "error": {
    "type": "validation",
    "message": "Invalid JSON syntax",
    "position": {
      "line": 3,
      "column": 15
    },
    "details": {
      "expected": "closing brace",
      "found": "EOF"
    }
  }
}
```

### Error Types

- `validation`: JSON validation errors
- `syntax`: JSON syntax errors
- `schema`: Schema validation errors
- `size`: Payload size limit exceeded
- `depth`: Nesting depth limit exceeded
- `field_count`: Field count limit exceeded
- `not_found`: Resource not found
- `session`: Session-related errors
- `template`: Template-related errors
- `enqueue`: Job enqueue errors
- `internal`: Internal server errors

## Security Features

### Secret Stripping

The studio automatically detects and strips sensitive information from payloads before logging or displaying in the UI:

- Passwords, tokens, API keys
- Bearer tokens in authorization headers
- Custom patterns defined in configuration

### Size and Complexity Limits

- Maximum payload size (default: 10MB)
- Maximum field count (default: 10,000)
- Maximum nesting depth (default: 50)

### Confirmation Prompts

In non-test environments, the studio requires confirmation before:
- Enqueuing jobs to production queues
- Applying templates with sensitive variables
- Bulk operations

## Dynamic Variables

The studio supports dynamic variable expansion in templates and snippets:

- `{{uuid}}`: Generates a UUID v4
- `{{now}}`: Current timestamp (ISO8601)
- `{{timestamp}}`: Unix timestamp
- `{{date}}`: Current date (YYYY-MM-DD)
- `{{time}}`: Current time (HH:MM:SS)
- `{{random}}`: Random number
- `{{env:VAR_NAME}}`: Environment variable
- `{{$1:placeholder}}`: User input placeholder

## Integration

### Redis Client Interface

The JSON Payload Studio requires a Redis client that implements:

```go
type RedisClient interface {
    Enqueue(queue string, payload interface{}, options ...interface{}) (string, error)
    EnqueueWithSchedule(queue string, payload interface{}, schedule time.Time) (string, error)
    EnqueueWithCron(queue string, payload interface{}, cronSpec string) (string, error)
}
```

### Usage Example

```go
import (
    "github.com/yourusername/go-redis-work-queue/internal/json-payload-studio"
)

// Create configuration
config := jsonpayloadstudio.DefaultConfig()
config.MaxPayloadSize = 5 * 1024 * 1024  // 5MB
config.StripSecrets = true

// Initialize studio
studio := jsonpayloadstudio.NewJSONPayloadStudio(config, redisClient)

// Create HTTP handler
handler := jsonpayloadstudio.NewHandler(studio)

// Register routes
mux := http.NewServeMux()
handler.RegisterRoutes(mux)

// Start server
http.ListenAndServe(":8080", mux)
```

## Performance Considerations

- Sessions are stored in memory - consider implementing persistent storage for production
- Template and schema caching reduces file I/O
- Validation is performed incrementally during typing for better responsiveness
- Large payloads are automatically truncated in preview mode
- Background auto-save prevents data loss

## Testing

The JSON Payload Studio includes comprehensive test coverage:

```bash
go test ./internal/json-payload-studio/... -v -cover
```

Benchmark tests are available for performance-critical operations:

```bash
go test ./internal/json-payload-studio/... -bench=. -benchmem
```

## Keyboard Shortcuts (TUI)

When integrated with a TUI interface:

- `Ctrl+S`: Save current payload
- `Ctrl+Z`: Undo
- `Ctrl+Y`: Redo
- `Ctrl+F`: Format JSON
- `Ctrl+Space`: Trigger auto-completion
- `e`: Enqueue with default options
- `E`: Open enqueue dialog with options
- `Ctrl+D`: Diff with previous version
- `Ctrl+P`: Preview formatted output
- `Ctrl+T`: Load template
- `Tab`: Expand snippet