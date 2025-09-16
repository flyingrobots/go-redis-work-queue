# TASKS and SLAPS

Level 1: You write code (no AI)
Level 2: You ask AI questions about code (AI as know-how)
Level 3: You use AI to help write code (AI pair programming)
Level 4: You make the plans, AI writes the code (Human-in-the-loop supervisor)
Level 5: You have ideas, AI makes them real (no human needed)

Leveling up on this scale, each level you attain feels like a productivity force multiplier. I would say "vibe coding" or "prompt engineering" is what goes on at level 3. "Spec engineering" is what goes on at level 4. But Level 5? That's when you say "Computer, add A/B tests to this app" and it just happens. That's the dream, right?

Once you hit level 3, it's not a big jump to reach level 4. But it's tricky. To optimize your performance on level 4, you need orchestration. I bet we've all seen or tried or even made up our own frameworks to make multiple AI agents work together. When it works well, the results are incredible. But it still takes quite a lot of effort to set up right.

T.A.S.K.S. and S.L.A.P.S. are my attempt to take things to the next level above human-in-the-loop orchestration. Yeah, that's what I said. Imagine: No more meticulous planning,  no more tedious supervision. I'm talking about full automation, baby. Idea in, product out.

## What are TASKS and SLAPS?

T.A.S.K.S. = **T**asks **A**re **S**equenced **K**ey **S**teps
S.L.A.P.S. = **S**ounds **L**ike **A** **P**lan **S**ystem

Use TASKS to plan how, use SLAPS to make it happen.

T.A.S.K.S. is an algorithm that takes your ideas and plans how to do them. It generates user stories, acceptance criteria, features, and tasks. It discovers dependencies between tasks and builds a graph where the edges represent hard and soft dependencies between them. It backs everything up with evidence and uses math.

## The Setup

- 10 Claude instances open
- 9 workers
- 1 coordinator
- 1 git branch (yes, really)

That's it.

Ahead of time, use T.A.S.K.S. to break down your ideas into features, features into tasks, identify dependencies (hard and soft) between tasks, build a DAG, and discover the root nodes (starting tasks). T.A.S.K.S. produces several artifacts, but the most important ones are:

```bash
# lists all tasks

- tasks.json



// arranges the tasks in a DAG where edges represent hard and soft dependencies between tasks

- dag.json



// declares how to manage access to shared runtime resources

- coordinator.json



## Tasks



Tasks are JSON objects that describe jobs that workers execute. There are two parts: task spec, and task state.



### Task Spec



The task is described pretty comprehensively because this is all the context that a Worker will have when they consume it.



### Task Info



What to do and how to do it.



- step-by-step implementation instructions

- boundaries

- expected complexity

- ~LoC

- definition of done

- scope

- what the task includes

- what is excludes

- restrictions

- instructions about when to log

- when to create checkpoints

- if there are any special resources/metrics/events to monitor

- resource requirements (est. and peak CPU/memory/disk io/etc)

- worker capability requirements

- expected execution duration (optimistic, most likely, pessimistic)

- shared resources, including exclusive locks and shared limited access

- acceptance criteria

- evidence (where did this task come from ex: technical spec doc w/excerpt, a confidence rating, and rationale)

- a detailed test plan (what sort of tests are required, what cases to write).



### Task State



Runtime job state, used by S.L.A.P.S.



- current state

- number of attempts

- last error





---



The coordinator is told:



You are the coordinator. Your job is to manage the Open Tasks directory and monitor the finished tasks directory and the help directory.



The workers are told "You are Worker 1", "You are Worker 2", etc. with instructions to watch a special directory for files to appear.



---



Appendix A: Task JSON object:



```json

{

"id": "P1.T012",

"name": "Implement Distributed Tracing Integration",

"task": {

"id": "P1.T012",

"feature_id": "F004",

"title": "Implement Distributed Tracing Integration",

"description": "Implement task for Distributed Tracing Integration",

"boundaries": {

"expected_complexity": {

"value": "~ LoC",

"breakdown": "Core logic (60%), Tests (25%), Integration (15%)"

},

"definition_of_done": {

"criteria": [

"All functions implemented per specification",

"Unit tests passing with 80% coverage",

"Integration tests passing",

"Code reviewed and approved",

"Documentation updated",

"No linting errors",

"Performance benchmarks met",

"Spans emitted for enqueue/dequeue/process with consistent attributes.",

"Context propagates via metadata; upstream trace linkage verified.",

"TUI shows trace IDs and open/copy actions."

],

"stop_when": "Core functionality complete; do NOT add extra features"

},

"scope": {

"includes": [

"internal/distributed-tracing-integration/",

"internal/distributed-tracing-integration/*_test.go",

"docs/api/distributed-tracing-integration.md"

],

"excludes": [

"UI unless specified",

"deployment configs"

],

"restrictions": "Follow existing code patterns and style guide"

}

},

"execution_guidance": {

"logging": {

"format": "JSON Lines (JSONL)",

"required_fields": [

"timestamp",

"task_id",

"step",

"status",

"message"

],

"optional_fields": [

"percent",

"data",

"checkpoint"

],

"status_values": [

"start",

"progress",

"done",

"error",

"checkpoint"

]

},

"checkpoints": [

{

"id": "setup",

"at_percent": 10,

"description": "Module structure created"

},

{

"id": "core",

"at_percent": 40,

"description": "Core logic implemented"

},

{

"id": "integration",

"at_percent": 60,

"description": "Integration complete"

},

{

"id": "tests",

"at_percent": 80,

"description": "Tests passing"

},

{

"id": "docs",

"at_percent": 100,

"description": "Documentation complete"

}

],

"monitoring": {

"metrics_to_track": [],

"alerts": []

}

},

"resource_requirements": {

"estimated": {

"cpu_cores": 1,

"memory_mb": 1024,

"disk_io_mbps": 10

},

"peak": {

"cpu_cores": 2,

"memory_mb": 2048,

"disk_io_mbps": 50,

"during": "compilation or testing"

},

"worker_capabilities_required": [

"golang",

"backend",

"redis"

]

},

"scheduling_hints": {

"priority": "high",

"preemptible": false,

"retry_on_failure": true,

"max_retries": 3,

"checkpoint_capable": true

},

"duration": {

"optimistic": 8,

"mostLikely": 16,

"pessimistic": 24

},

"shared_resources": {

"exclusive_locks": [],

"shared_limited": [

{

"resource": "test_redis",

"quantity": 1

}

],

"creates": [],

"modifies": []

},

"acceptance_checks": [

{

"type": "automated",

"description": "Spans emitted for enqueue/dequeue/process with consistent attributes.",

"script": "test_p1.t012.sh"

},

{

"type": "automated",

"description": "Context propagates via metadata; upstream trace linkage verified.",

"script": "test_p1.t012.sh"

},

{

"type": "automated",

"description": "TUI shows trace IDs and open/copy actions.",

"script": "test_p1.t012.sh"

},

{

"type": "manual",

"description": "Add otel setup in `internal/obs/tracing.go` with config",

"script": null

},

{

"type": "manual",

"description": "Instrument producer/worker/admin critical paths",

"script": null

}

],

"evidence": [

{

"type": "plan",

"source": "docs/ideas/distributed-tracing-integration.md",

"excerpt": "Make tracing first\u2011class with OpenTelemetry: automatically create spans for enqueue, dequeue, and job processing, propagate context through job payloads/metadata, and link to external tracing backends",

"confidence": 1.0,

"rationale": "Primary feature specification"

}

],

"implementation_spec": {

"implementation_checklist": [

"Add otel setup in `internal/obs/tracing.go` with config",

"Instrument producer/worker/admin critical paths",

"Inject/extract trace headers in metadata",

"Add TUI trace actions in Peek/Info",

"Docs with backend examples and sampling guidance"

],

"technical_approach": [

"SDK & instrumentation:",

"Use `go.opentelemetry.io/otel` across producer, worker, and admin.",

"Enqueue: start span `queue.enqueue` with attributes (queue, size, priority, tenant, idempotency_id).",

"Dequeue: span `queue.dequeue` with wait time and queue depth at dequeue.",

"Process: span `job.process` around user handler; record retries, outcome, error class.",

"Link parent context if `traceparent`/`tracestate` present in payload metadata; otherwise start a new root and inject on enqueue.",

"Propagation:",

"Embed W3C trace headers in job metadata (not payload) to avoid accidental redaction.",

"Ensure workers extract before processing and reinject on any outbound calls.",

"Exporters & sampling:",

"Default OTLP exporter to local Collector; config for endpoints/auth.",

"Head sampling with per\u2011route/queue rates; tail sampling via Collector for high\u2011value spans (errors, long latency).",

"Metrics + exemplars:",

"Attach trace IDs to latency/error metrics as exemplars when sampled.",

"TUI integration:",

"Show trace ID in Peek/Info; provide an \u201cOpen Trace\u201d action and copyable link; enable quick filter by trace ID.",

"Security & privacy:",

"Redact sensitive attributes; configurable allowlist for span attributes.",

"Disable/limit tracing in prod via config and sampling controls."

],

"code_structure": {

"main_package": "internal/distributed-tracing-integration",

"files": [

"distributed-tracing-integration.go - Main implementation",

"types.go - Data structures",

"handlers.go - Request handlers",

"errors.go - Custom errors",

"config.go - Configuration"

]

},

"tracing_components": [

"OpenTelemetry SDK integration",

"Trace context propagation (W3C format)",

"Span management (create, end, annotate)",

"Baggage propagation for metadata",

"Sampling strategies (always, probabilistic, adaptive)",

"Exporters (Jaeger, Zipkin, OTLP)",

"Metrics correlation"

],

"context_propagation": [

"Extract trace context from incoming requests",

"Inject context into outgoing requests",

"Create child spans for operations",

"Add span attributes and events"

]

}

},

"state": "PENDING",

"attempts": 0,

"lastError": null

}

```