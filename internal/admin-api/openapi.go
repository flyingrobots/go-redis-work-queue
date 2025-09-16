// Copyright 2025 James Ross
package adminapi

const openAPISpec = `openapi: 3.0.3
info:
  title: Redis Work Queue Admin API
  description: Secure admin API for managing Redis work queues
  version: 1.0.0
  contact:
    name: API Support
  license:
    name: MIT

servers:
  - url: http://localhost:8080/api/v1
    description: Local development server
  - url: https://api.example.com/api/v1
    description: Production server

security:
  - bearerAuth: []

tags:
  - name: stats
    description: Queue statistics and monitoring
  - name: queues
    description: Queue management operations
  - name: dlq
    description: Dead Letter Queue listing and remediation
  - name: workers
    description: Worker fleet information
  - name: benchmark
    description: Performance testing

paths:
  /stats:
    get:
      tags:
        - stats
      summary: Get queue statistics
      description: Returns current queue lengths, processing lists, and heartbeats
      operationId: getStats
      responses:
        '200':
          description: Statistics retrieved successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/StatsResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
        '500':
          $ref: '#/components/responses/InternalError'

  /stats/keys:
    get:
      tags:
        - stats
      summary: Get Redis keys statistics
      description: Returns detailed information about all managed Redis keys
      operationId: getStatsKeys
      responses:
        '200':
          description: Key statistics retrieved successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/StatsKeysResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
        '500':
          $ref: '#/components/responses/InternalError'

  /queues/{queue}/peek:
    get:
      tags:
        - queues
      summary: Peek at queue items
      description: View jobs in a queue without removing them
      operationId: peekQueue
      parameters:
        - name: queue
          in: path
          required: true
          description: Queue name (high, low, completed, dead_letter, or full key)
          schema:
            type: string
        - name: count
          in: query
          description: Number of items to peek (1-100)
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 10
      responses:
        '200':
          description: Queue items retrieved successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PeekResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'

  /queues/dlq:
    delete:
      tags:
        - queues
      summary: Purge dead letter queue
      description: Delete all items from the dead letter queue (requires confirmation)
      operationId: purgeDLQ
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PurgeRequest'
      responses:
        '200':
          description: Dead letter queue purged successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PurgeResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
        '500':
          $ref: '#/components/responses/InternalError'

  /queues/all:
    delete:
      tags:
        - queues
      summary: Purge all queues
      description: Delete all items from all queues (requires double confirmation)
      operationId: purgeAll
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PurgeRequest'
      responses:
        '200':
          description: All queues purged successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PurgeResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
        '500':
          $ref: '#/components/responses/InternalError'

  /bench:
    post:
      tags:
        - benchmark
      summary: Run performance benchmark
      description: Enqueue test jobs and measure throughput and latency
      operationId: runBenchmark
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/BenchRequest'
      responses:
        '200':
          description: Benchmark completed successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/BenchResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
        '500':
          $ref: '#/components/responses/InternalError'

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT token for authentication

  responses:
    BadRequest:
      description: Bad request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'

    Unauthorized:
      description: Authentication required
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'

    RateLimited:
      description: Rate limit exceeded
      headers:
        X-RateLimit-Limit:
          schema:
            type: integer
          description: Rate limit per minute
        X-RateLimit-Remaining:
          schema:
            type: integer
          description: Remaining requests
        X-RateLimit-Reset:
          schema:
            type: integer
          description: Unix timestamp when limit resets
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'

    InternalError:
      description: Internal server error
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'

  schemas:
    ErrorResponse:
      type: object
      required:
        - error
      properties:
        error:
          type: string
          description: Error message
        code:
          type: string
          description: Error code for programmatic handling
        details:
          type: object
          additionalProperties:
            type: string
          description: Additional error details

    StatsResponse:
      type: object
      required:
        - queues
        - processing_lists
        - heartbeats
        - timestamp
      properties:
        queues:
          type: object
          additionalProperties:
            type: integer
          description: Queue names and their lengths
        processing_lists:
          type: object
          additionalProperties:
            type: integer
          description: Worker processing lists and their lengths
        heartbeats:
          type: integer
          description: Number of active worker heartbeats
        timestamp:
          type: string
          format: date-time
          description: When the stats were collected

    StatsKeysResponse:
      type: object
      required:
        - queue_lengths
        - processing_lists
        - processing_items
        - heartbeats
        - timestamp
      properties:
        queue_lengths:
          type: object
          additionalProperties:
            type: integer
          description: Queue names and their lengths
        processing_lists:
          type: integer
          description: Number of processing lists
        processing_items:
          type: integer
          description: Total items in processing
        heartbeats:
          type: integer
          description: Number of active heartbeats
        rate_limit_key:
          type: string
          description: Rate limiter key name
        rate_limit_ttl:
          type: string
          description: Rate limiter TTL
        timestamp:
          type: string
          format: date-time

    PeekResponse:
      type: object
      required:
        - queue
        - items
        - count
        - timestamp
      properties:
        queue:
          type: string
          description: Full Redis key of the queue
        items:
          type: array
          items:
            type: string
          description: Job payloads (JSON strings)
        count:
          type: integer
          description: Number of items returned
        timestamp:
          type: string
          format: date-time

    PurgeRequest:
      type: object
      required:
        - confirmation
        - reason
      properties:
        confirmation:
          type: string
          description: Confirmation phrase (CONFIRM_DELETE for DLQ, CONFIRM_DELETE_ALL for all queues)
        reason:
          type: string
          minLength: 3
          maxLength: 500
          description: Reason for the destructive operation

    PurgeResponse:
      type: object
      required:
        - success
        - message
        - timestamp
      properties:
        success:
          type: boolean
        items_deleted:
          type: integer
          description: Number of items or keys deleted
        message:
          type: string
          description: Result message
        timestamp:
          type: string
          format: date-time

    BenchRequest:
      type: object
      required:
        - count
        - priority
      properties:
        count:
          type: integer
          minimum: 1
          maximum: 10000
          description: Number of test jobs to enqueue
        priority:
          type: string
          enum: [high, low]
          description: Queue priority for test jobs
        rate:
          type: integer
          minimum: 1
          maximum: 1000
          default: 100
          description: Jobs per second enqueue rate
        timeout_seconds:
          type: integer
          minimum: 1
          maximum: 300
          default: 30
          description: Maximum time to wait for completion

    BenchResponse:
      type: object
      required:
        - count
        - duration
        - throughput_jobs_per_sec
        - timestamp
      properties:
        count:
          type: integer
          description: Number of jobs processed
        duration:
          type: string
          description: Total benchmark duration
        throughput_jobs_per_sec:
          type: number
          format: float
          description: Jobs processed per second
        p50_latency:
          type: string
          description: 50th percentile latency
        p95_latency:
          type: string
          description: 95th percentile latency
        timestamp:
          type: string
          format: date-time

  /dlq:
    get:
      tags:
        - dlq
      summary: List DLQ items
      description: Returns a page of DLQ items with an opaque next cursor
      operationId: listDLQ
      parameters:
        - name: ns
          in: query
          required: false
          schema:
            type: string
          description: Namespace/prefix
        - name: cursor
          in: query
          required: false
          schema:
            type: string
          description: Opaque cursor for pagination
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            minimum: 1
            maximum: 500
            default: 100
          description: Page size
      responses:
        '200':
          description: DLQ items page
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DLQListResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
        '500':
          $ref: '#/components/responses/InternalError'

  /dlq/requeue:
    post:
      tags:
        - dlq
      summary: Requeue selected DLQ items
      operationId: requeueDLQ
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DLQRequeueRequest'
      responses:
        '200':
          description: Requeue summary
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DLQRequeueResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
        '500':
          $ref: '#/components/responses/InternalError'

  /dlq/purge:
    post:
      tags:
        - dlq
      summary: Purge selected DLQ items
      operationId: purgeDLQSelection
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DLQPurgeSelectionRequest'
      responses:
        '200':
          description: Purge summary
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DLQPurgeSelectionResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
        '500':
          $ref: '#/components/responses/InternalError'

  /workers:
    get:
      tags:
        - workers
      summary: List workers
      description: Returns summary of worker fleet
      operationId: listWorkers
      parameters:
        - name: ns
          in: query
          required: false
          schema:
            type: string
      responses:
        '200':
          description: Workers list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WorkersResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
        '500':
          $ref: '#/components/responses/InternalError'
    DLQItem:
      type: object
      required: [id, payload]
      properties:
        id:
          type: string
        queue:
          type: string
        payload:
          type: string
          description: Job payload as JSON string
        reason:
          type: string
        attempts:
          type: integer
        first_seen:
          type: string
          format: date-time
        last_seen:
          type: string
          format: date-time

    DLQListResponse:
      type: object
      required: [items, count, timestamp]
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/DLQItem'
        next_cursor:
          type: string
        count:
          type: integer
        timestamp:
          type: string
          format: date-time

    DLQRequeueRequest:
      type: object
      required: [ids]
      properties:
        ns:
          type: string
        ids:
          type: array
          items:
            type: string
        dest_queue:
          type: string

    DLQRequeueResponse:
      type: object
      required: [requeued, timestamp]
      properties:
        requeued:
          type: integer
        timestamp:
          type: string
          format: date-time

    DLQPurgeSelectionRequest:
      type: object
      required: [ids]
      properties:
        ns:
          type: string
        ids:
          type: array
          items:
            type: string

    DLQPurgeSelectionResponse:
      type: object
      required: [purged, timestamp]
      properties:
        purged:
          type: integer
        timestamp:
          type: string
          format: date-time

    WorkerInfo:
      type: object
      required: [id, last_heartbeat]
      properties:
        id:
          type: string
        last_heartbeat:
          type: string
          format: date-time
        queue:
          type: string
        job_id:
          type: string
        started_at:
          type: string
          format: date-time
        version:
          type: string
        host:
          type: string

    WorkersResponse:
      type: object
      required: [workers, timestamp]
      properties:
        workers:
          type: array
          items:
            $ref: '#/components/schemas/WorkerInfo'
        timestamp:
          type: string
          format: date-time
`
