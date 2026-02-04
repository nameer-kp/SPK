# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Atem is a workflow orchestration and automation framework written in Go. It executes multi-step workflows defined in YAML, with data flowing between steps via template variable substitution.

## Build & Run Commands

```bash
go build -o atem ./cmd/atem    # Build binary
./atem                             # Run with TUI
go test ./...                         # Run all tests
go test ./internal/engine/...         # Run tests for a specific package
```

## Architecture

```
cmd/atem/main.go          # Entry point: initializes logger, config, registry, engine
internal/
  config/                    # Multi-level config loading (global → project → profile → env)
  engine/                    # Workflow parser, executor, template resolver, state management
  tui/                       # Bubble Tea terminal UI
nodes/                       # Node implementations (http, db, shell, file, delay, transform)
pkg/node/                    # Public Node interface and Registry
```

### Core Concepts

- **Node**: Unit of work (HTTP call, DB query, shell command, etc). Implements `pkg/node/Node` interface
- **Registry**: Central registry where nodes register by type name, retrieved during execution
- **Workflow**: YAML definition with inputs and sequential steps
- **Template Engine**: Resolves `{{reference}}` syntax in configs (e.g., `{{inputs.user_id}}`, `{{step_id.field}}`)

### Data Flow

1. Config loads from hierarchy: `.env` → `~/.atem/config.yaml` → profiles → `.atem.yaml` → env vars
2. Workflow YAML parsed and validated
3. User inputs collected via TUI
4. Steps execute sequentially; each step's output available to subsequent steps via `{{step_id}}` or `{{step_id.field}}`
5. Results collected in `WorkflowState`

### Adding a New Node Type

1. Create `nodes/yournode.go` implementing `pkg/node/Node` interface
2. Register in `cmd/atem/main.go`: `registry.Register(&nodes.YourNode{})`
3. Node receives config with templates already resolved; return `node.Result` with data

### Configuration Hierarchy (lowest to highest priority)

1. `.env` file
2. `~/.atem/config.yaml` (global)
3. `~/.atem/profiles/*.yaml` (named profiles)
4. `.atem.yaml` (project)
5. Environment variables (`FLOWCLI_PROFILE`)

## Key Dependencies

- `charmbracelet/bubbletea` - TUI framework
- `expr-lang/expr` - Expression evaluation for transform node
- `jmoiron/sqlx` - Database queries
- `gopkg.in/yaml.v3` - YAML parsing

## Current Limitations

- Sequential execution only (no parallel steps)
- Conditional execution (`condition` field) defined but not implemented
- Error handling strategies (`on_error`) defined but not implemented
- Plugin system referenced but not implemented
