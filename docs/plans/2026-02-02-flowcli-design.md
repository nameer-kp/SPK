# flowcli Design Document

**Date:** 2026-02-02
**Status:** Approved
**Version:** 1.0

## Overview

**flowcli** is a Go-based developer CLI with Bubble Tea TUI for automating multi-step API workflows with conditional branching. It replaces the existing Python `booking_cli.py` with a more powerful, extensible workflow automation tool.

### Design Goals

- **Generic**: Not tied to any specific API or service (iTravel is just an example workflow)
- **Extensible**: Custom nodes via Go plugins
- **Interactive**: Wizard-style TUI with step-by-step prompts
- **Recoverable**: Pause on errors, retry/skip/abort options
- **Configurable**: Multi-tier config with profiles and environment support

### Core Concepts

| Concept | Description |
|---------|-------------|
| **Workflow** | A YAML file defining a sequence of nodes with conditions and data flow |
| **Node** | A unit of work (HTTP call, DB query, shell command, etc.) |
| **Context** | A shared data store that nodes read from and write to |
| **Profile** | Environment configuration (base URLs, tokens, DB connections) |

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      flowcli                            │
├─────────────────────────────────────────────────────────┤
│  TUI Layer (Bubble Tea)                                 │
│  - Wizard prompts, progress display, error recovery     │
├─────────────────────────────────────────────────────────┤
│  Workflow Engine                                        │
│  - YAML parser, node executor, condition evaluator      │
│  - Context management, checkpoint/resume                │
├─────────────────────────────────────────────────────────┤
│  Built-in Nodes              │  Plugin Nodes (.so)      │
│  - http, db, shell, file     │  - Custom Go plugins     │
│  - transform, delay, parallel│                          │
├─────────────────────────────────────────────────────────┤
│  Config Layer                                           │
│  - ~/.flowcli/profiles/, .env, .flowcli.yaml            │
└─────────────────────────────────────────────────────────┘
```

---

## Workflow File Format

### Schema

```yaml
name: string                    # Workflow identifier
description: string             # Human-readable description
version: string                 # Semantic version

inputs:                         # User prompts at start
  - name: string
    type: string|number|bool|choice|secret
    prompt: string
    required: bool
    default: any                # Optional
    options: []                 # For choice type

steps:                          # Execution sequence
  - id: string                  # Unique step ID
    name: string                # Display name in TUI
    type: string                # Node type (http, db, shell, etc.)
    condition: string           # Optional, expression to evaluate
    config: object              # Node-specific configuration
    output: string              # Variable name to store result
    on_error: retry|skip|abort  # Override default error behavior
```

### Templating Syntax

| Pattern | Description |
|---------|-------------|
| `{{inputs.var}}` | User-provided input |
| `{{profile.var}}` | From active profile/env |
| `{{step_id.field}}` | Output from previous step |
| `{{env.VAR}}` | System environment variable |

### Sample Workflow Location

```
~/.flowcli/workflows/examples/
  itravel-migration.yaml    # iTravel migration workflow
  rest-api-test.yaml        # Generic API testing
  db-seed.yaml              # Database seeding example
```

---

## Built-in Node Types

### HTTP Node

```yaml
type: http
config:
  method: GET|POST|PUT|DELETE|PATCH
  url: string
  headers: object
  body: object|string
  timeout: duration           # e.g., "30s"
  retry:
    attempts: number
    delay: duration
```

### Database Node

```yaml
type: db
config:
  connection: string          # Profile reference or inline DSN
  driver: postgres|mysql|sqlite
  operation: query|exec
  sql: string                 # Supports templating
  params: []                  # Positional parameters
```

### Shell Node

```yaml
type: shell
config:
  command: string
  args: []
  dir: string                 # Working directory
  env: object                 # Additional env vars
  stdin: string               # Pipe data to stdin
```

### File Node

```yaml
type: file
config:
  operation: read|write|append|delete|exists
  path: string
  content: string             # For write/append
  format: text|json|yaml      # Auto-parse on read
```

### Transform Node

```yaml
type: transform
config:
  input: string               # Source variable
  operations:
    - type: jq|jsonpath|template|map
      expression: string
  output: string
```

### Control Nodes

```yaml
# Delay execution
type: delay
config:
  duration: "5s"

# Run steps concurrently
type: parallel
config:
  steps: []                   # Nested steps run concurrently

# Iterate over array
type: loop
config:
  over: string                # Array variable to iterate
  as: string                  # Loop variable name
  steps: []                   # Steps to repeat
```

---

## Configuration System

### Priority Order (highest wins)

1. CLI flags (`--var key=value`)
2. `.flowcli.yaml` (project)
3. Active profile
4. `.env` file
5. System environment

### Global Config (`~/.flowcli/config.yaml`)

```yaml
default_profile: dev
editor: vim
log_level: info
plugins_dir: ~/.flowcli/plugins
checkpoints_dir: ~/.flowcli/checkpoints
```

### Profile (`~/.flowcli/profiles/*.yaml`)

```yaml
name: itravel-sandbox
description: iTravel Sandbox Environment

variables:
  base_url: https://sandbox.itravel.ibsplc.org/iTravel
  dm_service_url: http://localhost:8082
  itravel_local_url: http://localhost:8080

secrets:                      # Prompted if not in env/keychain
  session_token:
    env: X_SESSION_TOKEN      # Read from env var
  db_password:
    prompt: "Enter DB password"

database:
  host: localhost
  port: 5432
  name: itrvl_inf
  user: postgres
  password: "{{secrets.db_password}}"
  driver: postgres
```

### Project Config (`.flowcli.yaml`)

```yaml
default_profile: itravel-sandbox
workflows_dir: ./workflows
plugins_dir: ./plugins

variables:
  dm_service_url: http://localhost:8081
```

---

## TUI Design

### Screen Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Welcome   │───▶│   Select    │───▶│   Input     │───▶│  Execution  │
│   Screen    │    │  Workflow   │    │   Wizard    │    │   View      │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                                │
                                            ┌───────────────────┤
                                            ▼                   ▼
                                     ┌─────────────┐    ┌─────────────┐
                                     │   Error     │    │   Result    │
                                     │  Recovery   │    │   Summary   │
                                     └─────────────┘    └─────────────┘
```

### Welcome Screen

```
┌────────────────────────────────────────────────────────┐
│  flowcli v1.0.0                                        │
│                                                        │
│  Profile: itravel-sandbox                              │
│                                                        │
│  > Run Workflow                                        │
│    Manage Profiles                                     │
│    View Recent Runs                                    │
│    Settings                                            │
│                                                        │
│  ↑/↓ navigate  enter select  q quit                   │
└────────────────────────────────────────────────────────┘
```

### Execution View

```
┌────────────────────────────────────────────────────────┐
│  itravel-migration                        [2/4 steps]  │
├────────────────────────────────────────────────────────┤
│  ✓ retrieve_booking              200 OK      1.2s     │
│  ● insert_data                   running...           │
│  ○ migrate_booking               pending              │
│  ○ verify_result                 pending              │
├────────────────────────────────────────────────────────┤
│  POST http://localhost:8082/itravel/insertData         │
│  ├─ Request:  {"SuperPnrNumber": "10012529", ...}     │
│  └─ Status:   Waiting for response...                  │
├────────────────────────────────────────────────────────┤
│  space pause  r retry  s skip  esc abort              │
└────────────────────────────────────────────────────────┘
```

### Error Recovery Screen

```
┌────────────────────────────────────────────────────────┐
│  ✗ Step Failed: insert_data                           │
├────────────────────────────────────────────────────────┤
│  Error: connection refused                             │
│                                                        │
│  POST http://localhost:8082/itravel/insertData         │
│  Status: 0 (no response)                               │
├────────────────────────────────────────────────────────┤
│  > Retry this step                                     │
│    Skip and continue                                   │
│    Edit request and retry                              │
│    View full error details                             │
│    Abort workflow                                      │
│                                                        │
│  ↑/↓ navigate  enter select                           │
└────────────────────────────────────────────────────────┘
```

---

## Plugin System

### Node Interface

```go
// pkg/node/interface.go
package node

type Node interface {
    // Metadata
    Name() string
    Description() string
    ConfigSchema() map[string]any  // JSON Schema for validation

    // Execution
    Execute(ctx Context, config map[string]any) (Result, error)
}

type Context interface {
    Get(key string) (any, bool)       // Read from context
    Set(key string, value any)        // Write to context
    Profile() Profile                  // Access active profile
    Logger() Logger                    // Structured logging
}

type Result struct {
    Success bool
    Data    any               // Stored in context under output key
    Logs    []LogEntry        // Displayed in TUI
}
```

### Example Plugin

```go
// plugins/slack-notify/main.go
package main

import "github.com/yourorg/flowcli/pkg/node"

type SlackNotifyNode struct{}

func (n *SlackNotifyNode) Name() string { return "slack" }
func (n *SlackNotifyNode) Description() string {
    return "Send Slack notifications"
}

func (n *SlackNotifyNode) ConfigSchema() map[string]any {
    return map[string]any{
        "webhook_url": "string",
        "message":     "string",
        "channel":     "string",
    }
}

func (n *SlackNotifyNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
    // Implementation here
    return node.Result{Success: true}, nil
}

// Plugin export - required
var Node SlackNotifyNode
```

### Building & Installing Plugins

```bash
# Build plugin
go build -buildmode=plugin -o slack-notify.so ./plugins/slack-notify

# Install globally
cp slack-notify.so ~/.flowcli/plugins/

# Or project-local
cp slack-notify.so ./plugins/
```

### Using in Workflow

```yaml
steps:
  - id: notify_team
    type: slack              # Matches plugin Name()
    config:
      webhook_url: "{{profile.slack_webhook}}"
      message: "Migration complete: {{migrate_booking.pnr}}"
      channel: "#dev-alerts"
```

---

## Project Structure

```
flowcli/
├── cmd/
│   └── flowcli/
│       └── main.go              # Entry point
│
├── internal/
│   ├── tui/                     # Bubble Tea UI
│   │   ├── app.go               # Main app model
│   │   ├── screens/
│   │   │   ├── welcome.go
│   │   │   ├── workflow_select.go
│   │   │   ├── input_wizard.go
│   │   │   ├── execution.go
│   │   │   ├── error_recovery.go
│   │   │   └── result.go
│   │   ├── components/          # Reusable UI components
│   │   │   ├── list.go
│   │   │   ├── input.go
│   │   │   ├── progress.go
│   │   │   └── log_viewer.go
│   │   └── styles/
│   │       └── theme.go         # Lipgloss styles
│   │
│   ├── engine/                  # Workflow execution
│   │   ├── engine.go            # Main executor
│   │   ├── parser.go            # YAML parsing
│   │   ├── context.go           # Execution context
│   │   ├── template.go          # {{}} templating
│   │   ├── condition.go         # Expression evaluation
│   │   └── checkpoint.go        # State persistence
│   │
│   ├── config/                  # Configuration
│   │   ├── config.go            # Global config
│   │   ├── profile.go           # Profile management
│   │   └── loader.go            # Multi-source loading
│   │
│   └── plugin/                  # Plugin loader
│       └── loader.go
│
├── pkg/
│   └── node/                    # Public node interface
│       ├── interface.go         # Node, Context, Result
│       └── registry.go          # Node type registry
│
├── nodes/                       # Built-in nodes
│   ├── http.go
│   ├── db.go
│   ├── shell.go
│   ├── file.go
│   ├── transform.go
│   ├── delay.go
│   ├── parallel.go
│   └── loop.go
│
├── examples/
│   └── workflows/
│       └── itravel-migration.yaml
│
├── go.mod
├── go.sum
└── README.md
```

### Key Dependencies

```go
// go.mod
module github.com/yourorg/flowcli

require (
    github.com/charmbracelet/bubbletea   // TUI framework
    github.com/charmbracelet/lipgloss    // Styling
    github.com/charmbracelet/bubbles     // UI components
    gopkg.in/yaml.v3                     // YAML parsing
    github.com/expr-lang/expr            // Condition evaluation
    github.com/jmoiron/sqlx              // Database operations
    github.com/joho/godotenv             // .env loading
)
```

---

## Implementation Roadmap

### Phase 1: Core Foundation

- Project scaffolding with Go modules
- Basic TUI shell (welcome screen, navigation)
- Config/profile loading (`.env`, `.flowcli.yaml`, profiles)
- Workflow YAML parser with validation
- Context and templating engine (`{{variable}}` substitution)
- HTTP node only (covers main use case)

**Milestone:** Run iTravel retrieve endpoint via workflow

### Phase 2: Full Node Suite

- Database node (postgres, mysql, sqlite)
- Shell node with stdin/stdout capture
- File node (read/write/json/yaml parsing)
- Transform node (jq expressions, jsonpath)
- Delay node

**Milestone:** Run complete iTravel migration workflow

### Phase 3: Advanced Execution

- Condition evaluation (skip steps based on expressions)
- Interactive error recovery TUI
- Checkpoint save/resume
- Parallel and loop nodes

**Milestone:** Complex branching workflows with recovery

### Phase 4: Plugin System

- Go plugin loader
- Plugin discovery (global + project directories)
- Plugin template/generator command
- Documentation for plugin authors

**Milestone:** Custom Slack notification plugin working

### Phase 5: Polish

- Profile manager TUI screen
- Recent runs history
- Workflow validation command (`flowcli validate`)
- Export run logs/results
- Man pages / shell completions

---

## Sample iTravel Migration Workflow

```yaml
# examples/workflows/itravel-migration.yaml
name: itravel-migration
description: Retrieve booking, insert data, and migrate
version: "1.0"

inputs:
  - name: pnr_number
    type: string
    prompt: "Enter PNR number"
    required: true

steps:
  - id: retrieve_booking
    name: Retrieve Booking
    type: http
    config:
      method: POST
      url: "{{profile.base_url}}/selling/api/public-booking/v1/rest/bkg/pnr/retrieve"
      headers:
        Content-Type: application/json
        Accept: application/json
        x-session-token: "{{profile.session_token}}"
        x-auth-channel: "ItravelNativeUI@RCG"
      body:
        SuperPnrNumber: "{{inputs.pnr_number}}"
    output: booking_data

  - id: insert_data
    name: Insert to Migration DB
    type: http
    config:
      method: POST
      url: "{{profile.dm_service_url}}/itravel/insertData"
      headers:
        Content-Type: application/json
      body: "{{booking_data}}"
    output: insert_result

  - id: migrate_booking
    name: Call Migrate Booking API
    type: http
    condition: "{{insert_result.success == true}}"
    config:
      method: POST
      url: "{{profile.itravel_local_url}}/iTravel/api/public-booking/v1/rest/bkg/pnr/migrateBooking"
      headers:
        Content-Type: application/json
        Accept: application/json
        x-session-token: "{{profile.session_token}}"
        x-auth-channel: "ItravelNativeUI@RCG"
      body: "{{insert_result.request_payload}}"
    output: migration_result

  - id: show_result
    name: Display Result
    type: transform
    config:
      input: migration_result
      operations:
        - type: template
          expression: |
            Migration Status: {{.status}}
            New PNR: {{.newPnrNumber}}
    output: summary
```

---

## Decisions Log

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Primary use case | Both migration + extensible | Build for today, design for tomorrow |
| Workflow definition | YAML files | Version control friendly, team shareable |
| TUI interaction | Wizard mode | Developer-friendly exploration |
| Node types | Full toolkit | Maximum flexibility for dev workflows |
| Environment management | Hybrid (project + global) | Flexible for different scenarios |
| Custom nodes | Go plugins | Full power, type-safe |
| Error handling | Interactive recovery | Good for debugging complex workflows |
| CLI name | flowcli | Generic, not tied to iTravel |
