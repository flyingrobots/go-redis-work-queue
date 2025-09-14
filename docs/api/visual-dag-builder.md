# Visual DAG Builder API Documentation

## Overview

The Visual DAG Builder API provides endpoints for creating, managing, and executing workflow DAGs (Directed Acyclic Graphs). This RESTful API allows users to programmatically interact with the workflow system.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, authentication is optional and can be configured via the API configuration. When enabled, the API supports:
- JWT Bearer tokens
- Basic authentication
- API key authentication

## Content Type

All requests and responses use `application/json` content type unless otherwise specified.

## Error Response Format

All error responses follow this format:

```json
{
  "error": "Error description",
  "status": 400,
  "details": "Detailed error message"
}
```

## Endpoints

### Workflow Management

#### List Workflows

```http
GET /workflows
```

Returns a list of all workflows.

**Response:**
```json
{
  "workflows": [
    {
      "id": "workflow_1234567890",
      "name": "Sample Workflow",
      "description": "A sample workflow description",
      "version": "1.0.0",
      "created_at": "2025-09-14T12:00:00Z",
      "updated_at": "2025-09-14T12:00:00Z",
      "tags": ["sample", "demo"],
      "nodes": [...],
      "edges": [...]
    }
  ],
  "count": 1
}
```

#### Create Workflow

```http
POST /workflows
```

Creates a new workflow.

**Request Body:**
```json
{
  "name": "My New Workflow",
  "description": "Description of my workflow",
  "tags": ["tag1", "tag2"]
}
```

**Response:** Returns the created workflow object (201 Created).

#### Get Workflow

```http
GET /workflows/{id}
```

Retrieves a specific workflow by ID.

**Parameters:**
- `id` (path) - Workflow ID

**Response:** Returns the workflow object.

#### Update Workflow

```http
PUT /workflows/{id}
```

Updates an existing workflow.

**Parameters:**
- `id` (path) - Workflow ID

**Request Body:** Complete workflow object.

**Response:** Returns the updated workflow object.

#### Delete Workflow

```http
DELETE /workflows/{id}
```

Deletes a workflow.

**Parameters:**
- `id` (path) - Workflow ID

**Response:** 204 No Content on success.

#### Validate Workflow

```http
POST /workflows/{id}/validate
```

Validates a workflow definition.

**Parameters:**
- `id` (path) - Workflow ID

**Response:**
```json
{
  "valid": true,
  "errors": [],
  "warnings": [
    {
      "type": "unreachable_node",
      "message": "Node is not reachable from start nodes",
      "node_id": "isolated_node",
      "location": "node:isolated_node"
    }
  ]
}
```

### Node Management

#### Add Node

```http
POST /workflows/{id}/nodes
```

Adds a node to a workflow.

**Parameters:**
- `id` (path) - Workflow ID

**Request Body:**
```json
{
  "id": "task_node_1",
  "type": "task",
  "name": "Process Data",
  "description": "Processes incoming data",
  "position": {
    "x": 100,
    "y": 200
  },
  "job": {
    "queue": "data_processing",
    "type": "process_data",
    "payload": {
      "timeout": 300
    },
    "priority": "high",
    "timeout": "5m"
  },
  "retry": {
    "strategy": "exponential",
    "max_attempts": 3,
    "initial_delay": "1s",
    "max_delay": "5m",
    "multiplier": 2.0,
    "jitter": true
  }
}
```

**Response:** Returns the created node object (201 Created).

#### Update Node

```http
PUT /workflows/{id}/nodes/{nodeId}
```

Updates a node in a workflow.

**Parameters:**
- `id` (path) - Workflow ID
- `nodeId` (path) - Node ID

**Request Body:** Complete node object.

**Response:** Returns the updated node object.

#### Delete Node

```http
DELETE /workflows/{id}/nodes/{nodeId}
```

Deletes a node from a workflow.

**Parameters:**
- `id` (path) - Workflow ID
- `nodeId` (path) - Node ID

**Response:** 204 No Content on success.

### Edge Management

#### Add Edge

```http
POST /workflows/{id}/edges
```

Adds an edge to a workflow.

**Parameters:**
- `id` (path) - Workflow ID

**Request Body:**
```json
{
  "id": "edge_1_to_2",
  "from": "task_node_1",
  "to": "task_node_2",
  "type": "sequential",
  "label": "Success Path",
  "condition": "result.status == 'success'",
  "priority": 1,
  "delay": "0s"
}
```

**Response:** Returns the created edge object (201 Created).

#### Delete Edge

```http
DELETE /workflows/{id}/edges/{edgeId}
```

Deletes an edge from a workflow.

**Parameters:**
- `id` (path) - Workflow ID
- `edgeId` (path) - Edge ID

**Response:** 204 No Content on success.

### Execution Management

#### Execute Workflow

```http
POST /workflows/{id}/execute
```

Starts a new execution of a workflow.

**Parameters:**
- `id` (path) - Workflow ID

**Request Body:**
```json
{
  "parameters": {
    "input_file": "/data/input.csv",
    "batch_size": 1000,
    "notification_email": "user@example.com"
  }
}
```

**Response:**
```json
{
  "id": "exec_1234567890",
  "workflow_id": "workflow_1234567890",
  "workflow_version": "1.0.0",
  "status": "running",
  "parameters": {
    "input_file": "/data/input.csv",
    "batch_size": 1000,
    "notification_email": "user@example.com"
  },
  "started_at": "2025-09-14T12:30:00Z",
  "node_states": {
    "task_node_1": {
      "node_id": "task_node_1",
      "status": "running",
      "attempts": 1,
      "started_at": "2025-09-14T12:30:01Z"
    }
  }
}
```

#### List Executions

```http
GET /executions?workflow_id={workflowId}
```

Lists executions, optionally filtered by workflow ID.

**Query Parameters:**
- `workflow_id` (optional) - Filter by workflow ID

**Response:**
```json
{
  "executions": [
    {
      "id": "exec_1234567890",
      "workflow_id": "workflow_1234567890",
      "status": "completed",
      "started_at": "2025-09-14T12:00:00Z",
      "completed_at": "2025-09-14T12:05:00Z",
      "duration": "5m0s"
    }
  ],
  "count": 1
}
```

#### Get Execution

```http
GET /executions/{id}
```

Retrieves detailed information about an execution.

**Parameters:**
- `id` (path) - Execution ID

**Response:** Returns the complete execution object with node states.

#### Cancel Execution

```http
POST /executions/{id}/cancel
```

Cancels a running execution.

**Parameters:**
- `id` (path) - Execution ID

**Response:**
```json
{
  "status": "cancelled"
}
```

#### Get Execution Events

```http
GET /executions/{id}/events
```

Retrieves events for an execution.

**Parameters:**
- `id` (path) - Execution ID

**Response:**
```json
{
  "events": [
    {
      "id": "event_1234567890",
      "execution_id": "exec_1234567890",
      "node_id": "task_node_1",
      "event_type": "node_started",
      "status": "running",
      "message": "Node execution started",
      "timestamp": "2025-09-14T12:30:01Z"
    }
  ],
  "count": 1
}
```

### Utility Endpoints

#### Get Node Types

```http
GET /node-types
```

Returns available node types and their configuration requirements.

**Response:**
```json
{
  "node_types": [
    {
      "type": "task",
      "name": "Task",
      "description": "Executes a single job",
      "icon": "⚙️",
      "color": "#007bff",
      "required_fields": ["job.queue", "job.type"]
    },
    {
      "type": "decision",
      "name": "Decision",
      "description": "Evaluates conditions for branching",
      "icon": "🔀",
      "color": "#6f42c1",
      "required_fields": ["conditions"]
    }
  ]
}
```

#### Get Templates

```http
GET /templates
```

Returns available workflow templates.

**Response:**
```json
{
  "templates": [
    {
      "id": "linear-pipeline",
      "name": "Linear Pipeline",
      "description": "Simple sequential workflow",
      "category": "basic",
      "nodes": [...],
      "edges": [...]
    }
  ]
}
```

## Data Models

### Workflow Definition

```json
{
  "id": "string",
  "name": "string",
  "version": "string",
  "description": "string",
  "nodes": [Node],
  "edges": [Edge],
  "config": {
    "timeout": "duration",
    "concurrency_limit": "integer",
    "enable_compensation": "boolean",
    "enable_tracing": "boolean",
    "failure_strategy": "string"
  },
  "created_at": "datetime",
  "created_by": "string",
  "updated_at": "datetime",
  "updated_by": "string",
  "tags": ["string"],
  "metadata": {}
}
```

### Node Types

#### Task Node
```json
{
  "id": "string",
  "type": "task",
  "name": "string",
  "description": "string",
  "position": {"x": 0, "y": 0},
  "job": {
    "queue": "string",
    "type": "string",
    "payload": {},
    "priority": "string",
    "timeout": "duration"
  },
  "retry": {
    "strategy": "exponential|fixed|linear",
    "max_attempts": "integer",
    "initial_delay": "duration",
    "max_delay": "duration",
    "multiplier": "float",
    "jitter": "boolean"
  }
}
```

#### Decision Node
```json
{
  "id": "string",
  "type": "decision",
  "name": "string",
  "conditions": [
    {
      "expression": "string",
      "target": "string",
      "label": "string"
    }
  ],
  "default_target": "string"
}
```

#### Parallel Node
```json
{
  "id": "string",
  "type": "parallel",
  "name": "string",
  "parallel": {
    "wait_for": "all|any|count",
    "count": "integer",
    "branches": ["string"],
    "concurrency_limit": "integer"
  }
}
```

### Edge Types

```json
{
  "id": "string",
  "from": "string",
  "to": "string",
  "type": "sequential|conditional|compensation|loopback",
  "condition": "string",
  "label": "string",
  "priority": "integer",
  "delay": "duration"
}
```

### Execution Status Values

- `pending` - Execution is queued but not started
- `running` - Execution is currently running
- `completed` - Execution completed successfully
- `failed` - Execution failed
- `cancelled` - Execution was cancelled
- `paused` - Execution is paused
- `compensating` - Execution is running compensation logic

### Node Status Values

- `not_started` - Node has not been executed
- `queued` - Node is queued for execution
- `running` - Node is currently executing
- `completed` - Node completed successfully
- `failed` - Node execution failed
- `retrying` - Node is scheduled for retry
- `compensating` - Node compensation is running
- `compensated` - Node was compensated
- `skipped` - Node was skipped

## Rate Limiting

The API implements rate limiting with the following defaults:
- 100 requests per minute
- Burst size of 50 requests

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Request limit per window
- `X-RateLimit-Remaining`: Requests remaining in window
- `X-RateLimit-Reset`: Time when the rate limit resets

## Pagination

List endpoints support pagination using query parameters:
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10, max: 100)

Paginated responses include metadata:
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_pages": 5,
    "total_items": 47,
    "has_next": true,
    "has_prev": false
  }
}
```

## WebSocket Events

The API supports WebSocket connections for real-time updates:

```
ws://localhost:8080/api/v1/ws/executions/{execution_id}
```

Event types:
- `execution_started`
- `execution_completed`
- `execution_failed`
- `node_started`
- `node_completed`
- `node_failed`
- `node_retrying`

## Examples

### Create and Execute a Simple Workflow

1. Create a workflow:
```bash
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Data Processing Pipeline",
    "description": "Processes customer data"
  }'
```

2. Add nodes:
```bash
curl -X POST http://localhost:8080/api/v1/workflows/{workflow_id}/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "id": "validate",
    "type": "task",
    "name": "Validate Data",
    "job": {
      "queue": "validation",
      "type": "validate_customer_data"
    }
  }'
```

3. Add edge:
```bash
curl -X POST http://localhost:8080/api/v1/workflows/{workflow_id}/edges \
  -H "Content-Type: application/json" \
  -d '{
    "id": "validate_to_process",
    "from": "validate",
    "to": "process",
    "type": "sequential"
  }'
```

4. Execute workflow:
```bash
curl -X POST http://localhost:8080/api/v1/workflows/{workflow_id}/execute \
  -H "Content-Type: application/json" \
  -d '{
    "parameters": {
      "input_file": "/data/customers.csv"
    }
  }'
```

## Error Codes

| Status Code | Description |
|-------------|-------------|
| 200 | OK |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 422 | Unprocessable Entity |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

## OpenAPI Specification

The complete OpenAPI 3.0 specification is available at:
```
http://localhost:8080/api/v1/openapi.json
```

Interactive API documentation is available at:
```
http://localhost:8080/api/v1/docs
```