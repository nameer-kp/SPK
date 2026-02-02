# FlowCLI Phase 3: Advanced Execution - Design Document

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enable complex branching workflows with conditions, loops, parallel execution, error recovery, and checkpoints.

**Architecture:** Event-driven TUI communicating with engine via channels. Engine evaluates conditions before steps, supports nested execution for loop/parallel nodes, and emits events for TUI updates.

**Tech Stack:** Bubble Tea (TUI), expr-lang/expr (conditions), goroutines (parallel), JSON (checkpoints)

---

## 1. Condition Evaluation

### File: `flowcli/internal/engine/condition.go`

**Evaluator struct:**
```go
type Evaluator struct {
    ctx *ExecutionContext
}

func NewEvaluator(ctx *ExecutionContext) *Evaluator
func (e *Evaluator) Evaluate(condition string) (bool, error)
```

**Supported expressions:**
- `step_id.field == value` - Compare step output
- `inputs.var != ""` - Check input exists
- `env.DEBUG == "true"` - Check environment
- Boolean operators: `&&`, `||`, `!`
- Comparisons: `==`, `!=`, `>`, `<`, `>=`, `<=`

**Integration point:** `engine.go` line 98, before `ExecuteStep()`

**Example workflow:**
```yaml
steps:
  - id: check_user
    type: http
    config:
      url: "{{api_url}}/users/{{user_id}}"
    output: user_result

  - id: update_user
    type: http
    condition: "user_result.data.status == 'active'"
    config:
      method: PUT
      url: "{{api_url}}/users/{{user_id}}"
```

---

## 2. Loop Node

### File: `flowcli/nodes/loop.go`

**Config schema:**
```go
type LoopConfig struct {
    Items string   `yaml:"items"` // Template expression for array
    As    string   `yaml:"as"`    // Variable name for current item
    Index string   `yaml:"index"` // Optional index variable name
    Steps []Step   `yaml:"steps"` // Nested steps to execute
}
```

**Execution:**
1. Resolve `items` template to get array
2. For each item, inject `as` and `index` into context
3. Execute nested steps sequentially
4. Collect results as array

**Output format:**
```json
{
  "iterations": [
    {"index": 0, "item": {...}, "results": {...}},
    {"index": 1, "item": {...}, "results": {...}}
  ],
  "count": 2
}
```

**Example:**
```yaml
- id: process_users
  type: loop
  config:
    items: "{{fetch_users.data.users}}"
    as: "user"
    index: "i"
  steps:
    - id: notify_user
      type: http
      config:
        url: "{{api_url}}/notify/{{user.id}}"
```

---

## 3. Parallel Node

### File: `flowcli/nodes/parallel.go`

**Config schema:**
```go
type ParallelConfig struct {
    MaxConcurrent int    `yaml:"max_concurrent"` // 0 = unlimited
    Steps         []Step `yaml:"steps"`
}
```

**Execution:**
1. Parse nested steps
2. Create goroutine per step (with semaphore if max_concurrent > 0)
3. Use `sync.WaitGroup` to wait for all
4. Aggregate results by step ID

**Output format:**
```json
{
  "results": {
    "fetch_orders": {"success": true, "data": {...}},
    "fetch_products": {"success": true, "data": {...}}
  },
  "all_success": true,
  "duration": "1.2s"
}
```

**Example:**
```yaml
- id: fetch_all
  type: parallel
  config:
    max_concurrent: 3
  steps:
    - id: fetch_orders
      type: http
      config: { url: "{{api_url}}/orders" }
    - id: fetch_products
      type: http
      config: { url: "{{api_url}}/products" }
```

---

## 4. Error Recovery

### Workflow-level config
```yaml
name: my_workflow
on_error: pause  # pause, abort (default), skip
```

### Step-level override
```yaml
- id: risky_step
  type: http
  on_error: retry  # retry, skip, abort, pause
  config:
    url: "{{api_url}}/unstable"
```

### Error handler behavior

| on_error | Behavior |
|----------|----------|
| `abort`  | Stop workflow immediately (default) |
| `skip`   | Log error, continue to next step |
| `retry`  | Retry 3x with exponential backoff (1s, 2s, 4s) |
| `pause`  | Show error recovery TUI, wait for user decision |

### File: `flowcli/internal/engine/recovery.go`

```go
type ErrorHandler struct {
    defaultAction string // from workflow.OnError
    events        chan<- ExecutionEvent
    decisions     <-chan RecoveryAction
}

type RecoveryAction int
const (
    ActionRetry RecoveryAction = iota
    ActionSkip
    ActionAbort
    ActionEditRetry
)

func (h *ErrorHandler) Handle(step *Step, err error) (RecoveryAction, error)
```

### Error Recovery TUI Screen

File: `flowcli/internal/tui/screens/error_recovery.go`

```
┌─ Step Failed: risky_step ─────────────────────────┐
│                                                    │
│  Error: HTTP 503 Service Unavailable               │
│  After: 2.3s                                       │
│                                                    │
│  [R] Retry this step                               │
│  [S] Skip and continue                             │
│  [E] Edit config and retry                         │
│  [V] View full error details                       │
│  [A] Abort workflow                                │
│                                                    │
└────────────────────────────────────────────────────┘
```

---

## 5. Checkpoint System

### File: `flowcli/internal/engine/checkpoint.go`

**Location:** `~/.flowcli/checkpoints/<workflow_name>_<timestamp>.json`

**Checkpoint structure:**
```go
type Checkpoint struct {
    WorkflowPath string                 `json:"workflow_path"`
    WorkflowHash string                 `json:"workflow_hash"`
    StartedAt    time.Time              `json:"started_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
    CurrentStep  int                    `json:"current_step"`
    Status       string                 `json:"status"`
    Inputs       map[string]interface{} `json:"inputs"`
    Context      map[string]interface{} `json:"context"`
    Results      []StepResult           `json:"results"`
}

type CheckpointManager struct {
    dir string // ~/.flowcli/checkpoints/
}

func (m *CheckpointManager) Save(state *WorkflowState) error
func (m *CheckpointManager) Load(workflowPath string) (*Checkpoint, error)
func (m *CheckpointManager) List() ([]Checkpoint, error)
func (m *CheckpointManager) Delete(id string) error
```

**Save triggers:**
- After each successful step
- On pause (error recovery or manual)
- On Ctrl+P in TUI

**Resume validation:**
- Compare `workflow_hash` (SHA256 of workflow file)
- Warn if workflow modified since checkpoint

---

## 6. TUI Screens

### 6.1 WorkflowSelect Screen

File: `flowcli/internal/tui/screens/workflow_select.go`

```
┌─ Select Workflow ──────────────────────────────────┐
│                                                    │
│  > booking_workflow        Create hotel bookings   │
│    user_sync              Sync users to CRM        │
│    data_export            Export reports to S3     │
│                                                    │
│  ~/.flowcli/workflows/ (3 workflows)               │
│                                                    │
│  [Enter] Select  [/] Filter  [q] Quit              │
└────────────────────────────────────────────────────┘
```

**Features:**
- List workflows from `parser.ListWorkflows()`
- Show name, description, step count
- Filter by typing
- Return selected workflow path

### 6.2 InputWizard Screen

File: `flowcli/internal/tui/screens/input_wizard.go`

```
┌─ booking_workflow - Inputs ────────────────────────┐
│                                                    │
│  API URL:  [https://api.example.com_____]          │
│  User ID:  [12345_________________________]        │
│  Environment: ● Production  ○ Staging              │
│  Dry Run:  [x] Yes                                 │
│                                                    │
│  [Enter] Continue  [Tab] Next field  [Esc] Back    │
└────────────────────────────────────────────────────┘
```

**Input types:**
- `string`: Text input
- `number`: Numeric input with validation
- `bool`: Checkbox
- `choice`: Radio buttons from options
- `secret`: Password input (masked)

**Validation:**
- Required fields must be filled
- Type validation (number must be numeric)
- Show errors inline

### 6.3 Execution Screen

File: `flowcli/internal/tui/screens/execution.go`

```
┌─ Running: booking_workflow ────────────────────────┐
│                                                    │
│  ✓ fetch_user         200 OK           0.3s        │
│  ✓ check_availability 200 OK           0.8s        │
│  ⟳ create_booking     Running...                   │
│  ○ send_confirmation  Pending                      │
│                                                    │
│  ├─ POST /api/bookings                             │
│  │  {"user_id": 123, "room": "deluxe"}             │
│                                                    │
│  [P] Pause  [Esc] Abort                            │
└────────────────────────────────────────────────────┘
```

**Features:**
- Real-time step status updates
- Show current request details
- Display response preview
- Keyboard controls: Pause, Abort

### 6.4 Result Screen

File: `flowcli/internal/tui/screens/result.go`

```
┌─ Workflow Complete ────────────────────────────────┐
│                                                    │
│  ✓ booking_workflow completed in 4.2s              │
│                                                    │
│  Steps: 4/4 successful                             │
│  Output: booking_id = "BK-98765"                   │
│                                                    │
│  [E] Export JSON  [Y] Export YAML  [R] Run again   │
│  [Enter] Back to menu                              │
└────────────────────────────────────────────────────┘
```

**Features:**
- Summary with duration and status
- Key outputs displayed
- Export to JSON/YAML
- Run again option

---

## 7. Event System

### File: `flowcli/internal/engine/events.go`

```go
type EventType int
const (
    EventStepStart EventType = iota
    EventStepComplete
    EventStepError
    EventWorkflowComplete
    EventPaused
    EventResumed
)

type ExecutionEvent struct {
    Type      EventType
    Step      *Step
    Result    *StepResult
    Error     error
    Timestamp time.Time
}

type EventEmitter struct {
    ch chan<- ExecutionEvent
}

func (e *EventEmitter) Emit(event ExecutionEvent)
```

**TUI Integration:**
- Engine runs in goroutine
- Sends events via channel
- TUI listens and updates via `tea.Cmd`

---

## 8. Engine Modifications

### Updated Run() method

```go
func (e *Engine) Run(workflow *Workflow, inputs map[string]interface{}, events chan<- ExecutionEvent) (*WorkflowState, error) {
    ctx := NewExecutionContext(e.config, e.logger)
    ctx.SetInputs(inputs)

    evaluator := NewEvaluator(ctx)
    checkpoint := NewCheckpointManager()
    errorHandler := NewErrorHandler(workflow.OnError, events)

    state := &WorkflowState{
        Workflow: workflow,
        Status:   "running",
    }

    for i, step := range workflow.Steps {
        // Check condition
        if step.Condition != "" {
            shouldRun, err := evaluator.Evaluate(step.Condition)
            if err != nil {
                return state, err
            }
            if !shouldRun {
                events <- ExecutionEvent{Type: EventStepSkipped, Step: &step}
                continue
            }
        }

        events <- ExecutionEvent{Type: EventStepStart, Step: &step}

        result, err := e.ExecuteStep(ctx, &step)
        if err != nil {
            action, err := errorHandler.Handle(&step, err)
            switch action {
            case ActionRetry:
                i-- // Retry same step
                continue
            case ActionSkip:
                continue
            case ActionAbort:
                return state, err
            }
        }

        state.Results = append(state.Results, result)
        state.CurrentStep = i + 1

        events <- ExecutionEvent{Type: EventStepComplete, Step: &step, Result: &result}

        // Save checkpoint
        checkpoint.Save(state)
    }

    state.Status = "completed"
    events <- ExecutionEvent{Type: EventWorkflowComplete}
    return state, nil
}
```

---

## 9. File Summary

### New Files
| File | Purpose |
|------|---------|
| `internal/engine/condition.go` | Expression evaluation with expr-lang |
| `internal/engine/checkpoint.go` | Save/load workflow state |
| `internal/engine/recovery.go` | Error handling logic |
| `internal/engine/events.go` | Event types and emitter |
| `nodes/loop.go` | Loop iteration node |
| `nodes/parallel.go` | Parallel execution node |
| `internal/tui/screens/workflow_select.go` | Workflow picker |
| `internal/tui/screens/input_wizard.go` | Dynamic input form |
| `internal/tui/screens/execution.go` | Progress display |
| `internal/tui/screens/error_recovery.go` | Error handling UI |
| `internal/tui/screens/result.go` | Results and export |

### Modified Files
| File | Changes |
|------|---------|
| `internal/engine/engine.go` | Add condition check, event emission, error handling |
| `internal/engine/types.go` | Add OnError field to Workflow, add event types |
| `internal/tui/app.go` | Wire all screens, manage engine lifecycle |
| `cmd/flowcli/main.go` | Register loop and parallel nodes |

---

## 10. Testing Strategy

Each component should have unit tests:
- `condition_test.go` - Expression evaluation cases
- `checkpoint_test.go` - Save/load/validate
- `loop_test.go` - Iteration with variable binding
- `parallel_test.go` - Concurrent execution, result aggregation
- `recovery_test.go` - Error handling logic

Integration tests:
- Full workflow with conditions
- Checkpoint resume after simulated failure
- Parallel node timing verification
