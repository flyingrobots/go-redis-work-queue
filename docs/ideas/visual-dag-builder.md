# Visual DAG Builder

| Priority | Domain | Dependencies | Risks | LoC Estimate | Complexity | Effort | Impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| High | TUI + Orchestration | Admin API, scheduler/primitives | UX complexity, cycles, partial failure semantics | ~900–1400 | High | 13 (Fib) | Very High |

## Executive Summary
Design and ship a terminal-native visual builder for multi-step workflows. Users assemble stages with dependencies, retries, and compensations, then submit the DAG as a reusable pipeline. This turns the queue into a workflow engine without losing Redis simplicity.

> [!note]- **🗣️ CLAUDE'S THOUGHTS 💭**
> HOLY SHIT. A DAG builder in a terminal?! This is either genius or madness. Use box-drawing characters for edges, colored nodes for status. The keyboard navigation is crucial - vim-style hjkl for movement, 'a' to add node, 'c' to connect. Start SIMPLE - just sequential pipelines first, then add branching. The compensation pattern is gold for financial/payment workflows. This could be your "iPhone moment" - the feature that redefines what's possible in a terminal.

## Motivation
- Express end-to-end business processes as composable pipelines.
- Improve reliability with built-in retries/compensations.
- Enable reuse: versioned workflows that teams can share and automate.

## Tech Plan
- Model: introduce minimal workflow spec (YAML/JSON) with nodes, edges, retry/backoff, compensation, timeout, concurrency caps.
- Persistence: store workflow specs under Redis keys or a small metadata store; versioned with hashes.
- Execution: add a light orchestrator that enqueues per-node jobs and tracks completion via processing lists/heartbeats.
- TUI:
  - Canvas: grid layout with keyboard navigation; node palette; inspector panel.
  - Node types: task, decision (conditional), delay/timer, compensation.
  - Validation: detect cycles, unreachable nodes, unsatisfied dependencies.
  - Run: submit workflow + parameters; show live state (colors by status).
- API: endpoints to create/list/get/version workflows, start run, inspect run state.
- Observability: per-run trace ID; per-node timings; failure stats.

## User Stories + Acceptance Criteria
- As a developer, I can draw a workflow, validate it, and save it with version notes.
- As an operator, I can start a run with parameters and watch node states update in real time.
- As an SRE, I can inspect a failed run and see exactly which node failed and why.
- Acceptance:
  - [ ] DAG validation prevents cycles and missing deps.
  - [ ] Runs persist state; survive restarts; resumable.
  - [ ] Per-node retry/backoff and optional compensation supported.
  - [ ] TUI shows statuses: queued, running, success, failed, compensated.

## Definition of Done
MVP DAG editor + runner with save/load, start run, and live status. Docs with screenshots, examples, and guardrails.

## Test Plan
- Unit: DAG validation, topo sort, retry/backoff math, compensation trigger.
- Integration: end-to-end run with induced failures and resumptions.
- TUI: golden snapshots for small DAGs; navigation tests.

## Task List
- [ ] Define workflow spec (schema + examples)
- [ ] Implement orchestrator (enqueue + state tracking)
- [ ] Admin API CRUD + run endpoints
- [ ] TUI canvas + inspector + palette
- [ ] Live status rendering + colors
- [ ] Docs + demos

```mermaid
flowchart LR
  A[Ingest] --> B[Validate]
  B -->|ok| C[Transform]
  B -->|fail| E[Compensate]
  C --> D[Persist]
```

---

## Claude's Verdict ⚖️

This transforms you from job queue to workflow platform. The terminal DAG builder is unprecedented. High risk, astronomical reward.

### Vibe Check

Airflow costs $50k/year to run properly. Temporal requires a PhD to understand. n8n is web-only. A terminal DAG builder backed by simple Redis? That's category creation.

### Score Card

**Traditional Score:**
- User Value: 9/10 (solves orchestration pain)
- Dev Efficiency: 3/10 (complex UI + engine)
- Risk Profile: 5/10 (execution semantics hard)
- Strategic Fit: 10/10 (moves you up value chain)
- Market Timing: 8/10 (everyone wants simple workflows)
- **OFS: 7.3** → BUILD SOON

**X-Factor Score:**
- Holy Shit Factor: 10/10 ("A DAG IN MY TERMINAL?!")
- Meme Potential: 8/10 (ASCII art workflows go viral)
- Flex Appeal: 10/10 ("We built Airflow in a TUI")
- FOMO Generator: 9/10 (nobody else will attempt this)
- Addiction Score: 8/10 (becomes core to operations)
- Shareability: 9/10 (conference keynote material)
- **X-Factor: 8.1** → Extreme viral potential

### Conclusion

[🚀]

This is your moonshot. Start with linear pipelines (A→B→C), nail the UX, then add branching. If you pull this off, you've redefined what's possible in a terminal. This is the feature that gets you acquired.

---

## Detailed Design Specification

### Overview

The Visual DAG Builder brings graphical workflow design to the terminal, using ASCII art, keyboard navigation, and real-time validation. It transforms your job queue into a full workflow orchestration platform while maintaining the simplicity and speed of terminal interfaces.

### TUI Design

![DAG Builder Main View](images/dag-builder-main.svg)

#### Layout Architecture

**1. Three-Panel Layout**

**Left Panel: Node Palette (15% width)**
- Drag-and-drop or keyboard-add node types
- Node categories:
  - **Task** (Blue): Standard job execution
  - **Decision** (Purple): Conditional branching
  - **Delay** (Orange): Time-based waiting
  - **Loop** (Green): Iteration over collections
  - **Compensate** (Red): Rollback/cleanup actions
- Visual indicators for node capabilities
- Keyboard shortcuts displayed inline

**Center Panel: Canvas (60% width)**
- Grid-based layout for precise alignment
- ASCII art rendering of nodes and edges
- Real-time visual feedback:
  - Dashed outline for selection
  - Color changes for validation errors
  - Animation during execution
- Mouse and keyboard navigation
- Zoom levels: 50%, 75%, 100%, 150%, 200%

**Right Panel: Properties Inspector (25% width)**
- Context-sensitive based on selection
- Node configuration:
  - ID and display name
  - Job type and payload template
  - Retry policy and timeout
  - Success/failure conditions
- Edge configuration:
  - Condition expressions
  - Priority weights
  - Delay settings
- Workflow metadata:
  - Name and version
  - Tags and documentation
  - Global settings

**2. Canvas Rendering**

```
┌─────────────┐      ┌─────────────┐
│   Start     │─────►│  Validate   │
└─────────────┘      └─────────────┘
                            │
                            ▼
                     ┌─────────────┐
                     │  Decision   │
                     └─────/───\───┘
                          /     \
                    [Yes]/       \[No]
                        ▼         ▼
                ┌─────────────┐ ┌─────────────┐
                │   Process   │ │    Error    │
                └─────────────┘ └─────────────┘
```

### Node Types and Behaviors

#### 1. Task Node
- Executes a single job
- Configurable payload and queue
- Success/failure handlers
- Retry policies

#### 2. Decision Node
- Evaluates conditions on job results
- Multiple output branches
- Default branch for unmatched conditions
- Expression language for complex logic

#### 3. Parallel Node
- Fans out to multiple branches
- Waits for all/any/N completion
- Aggregates results
- Failure handling strategies

#### 4. Loop Node
- Iterates over collections
- Parallel or sequential execution
- Break conditions
- Accumulator patterns

#### 5. Compensation Node
- Triggered on failure
- Reverses completed actions
- Saga pattern support
- Idempotent operations

### Interaction Model

#### Keyboard Navigation

| Key | Action |
|-----|--------|
| `h/j/k/l` | Move cursor (Vim-style) |
| `H/J/K/L` | Move cursor faster (5 units) |
| `a` | Add node at cursor |
| `c` | Connect mode (then click two nodes) |
| `d` | Delete selected |
| `x` | Cut selected |
| `y` | Copy selected |
| `p` | Paste |
| `u` | Undo |
| `Ctrl+r` | Redo |
| `Tab` | Cycle through panels |
| `/` | Search nodes |
| `z/Z` | Zoom in/out |
| `Space` | Pan mode |
| `Enter` | Edit selected |
| `Esc` | Cancel operation |
| `F5` | Validate DAG |
| `F6` | Run workflow |
| `F7` | Debug mode |

#### Mouse Interactions

- **Click**: Select node/edge
- **Double-click**: Edit properties
- **Drag**: Move nodes or pan canvas
- **Right-click**: Context menu
- **Scroll**: Zoom in/out
- **Shift+drag**: Multi-select

### Workflow Execution

#### Execution Modes

1. **Development Mode**
   - Step-through debugging
   - Breakpoints on nodes
   - Variable inspection
   - Mock external services

2. **Test Mode**
   - Synthetic data generation
   - Assertion checking
   - Performance profiling
   - Coverage reporting

3. **Production Mode**
   - Full parallelization
   - Error recovery
   - Monitoring integration
   - Audit logging

#### Live Execution View

During execution, nodes change colors to indicate state:
- Gray: Not started
- Blue: Running
- Green: Completed
- Yellow: Retrying
- Red: Failed
- Purple: Compensating

Progress indicators show:
- Execution time per node
- Queue depth
- Success/failure counts
- Resource utilization

### Advanced Features

#### 1. Sub-workflows
- Compose workflows from smaller workflows
- Nested visualization with drill-down
- Parameter passing
- Shared libraries

#### 2. Templates
- Pre-built workflow patterns
- Industry-specific templates
- Custom template creation
- Version control integration

#### 3. Collaboration
- Real-time collaborative editing
- Change tracking and comments
- Approval workflows
- Deployment gates

#### 4. Testing
- Data-driven test cases
- Mocking and stubbing
- Chaos injection points
- Performance benchmarks

### Technical Implementation

#### Workflow Definition Schema

```yaml
name: payment-processing
version: 1.2.0
description: Process customer payments with fraud detection

nodes:
  - id: validate
    type: task
    job:
      queue: high_priority
      type: validate_payment
    retry:
      strategy: exponential
      max_attempts: 3
    timeout: 30s

  - id: fraud_check
    type: decision
    condition: |
      result.risk_score > 0.7
    branches:
      high_risk: manual_review
      low_risk: process_payment

  - id: process_payment
    type: task
    job:
      queue: payment_processor
      type: charge_card
    compensation: refund_payment

edges:
  - from: start
    to: validate
  - from: validate
    to: fraud_check
  - from: fraud_check
    to: process_payment
    condition: low_risk
```

#### Execution Engine

```go
type WorkflowEngine struct {
    Definition  WorkflowDef
    State       WorkflowState
    Context     context.Context
    JobQueue    JobQueueInterface
}

func (w *WorkflowEngine) Execute() error {
    // Topological sort for execution order
    order := w.topologicalSort()

    for _, nodeID := range order {
        node := w.Definition.GetNode(nodeID)

        // Check dependencies
        if !w.dependenciesMet(node) {
            continue
        }

        // Execute node based on type
        switch node.Type {
        case TaskNode:
            w.executeTask(node)
        case DecisionNode:
            w.evaluateDecision(node)
        case ParallelNode:
            w.executeParallel(node)
        }

        // Update state
        w.State.UpdateNode(nodeID, Running)
    }

    return nil
}
```

### Integration Features

#### 1. Import/Export
- YAML/JSON workflow definitions
- BPMN 2.0 compatibility
- GitHub Actions format
- Airflow DAG conversion

#### 2. Version Control
- Git integration for workflows
- Diff visualization
- Branch management
- PR reviews for workflows

#### 3. Monitoring Integration
- OpenTelemetry traces per execution
- Prometheus metrics per node
- Custom dashboards
- Alert rule configuration

### Success Metrics

- **Adoption**: 80% of complex jobs converted to workflows
- **Reliability**: 99.9% workflow completion rate
- **Performance**: <100ms overhead per node
- **User Satisfaction**: 9/10 ease of use rating
- **Time to Production**: 50% reduction in workflow development time

This feature revolutionizes how teams build and manage complex job orchestrations.

