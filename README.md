# Atem

**Task automation for developers and ops teams.** Define workflows in YAML, run them with an interactive TUI.

Atem is like `npm scripts` or `Makefile` but with built-in support for HTTP requests, database queries, file operations, and data transformation—with results flowing between steps.

## Why Atem?

You have tasks that are tedious to do manually:

- Calling an API, then updating a database with the result
- Fetching data from multiple services and combining them
- Running a script that needs values from a config file or API
- Testing a workflow that spans HTTP → transform → DB → notify

You could write a bash script. But then you're juggling `curl`, `jq`, `psql`, and string interpolation. Or you could click through a UI. Every. Single. Time.

**Atem sits in the gap between shell scripts and heavyweight workflow platforms:**

```
npm scripts / Make     →  shell commands only, no data flow between steps
Task / Just            →  better orchestration, still shell-centric
Atem                →  declarative nodes (HTTP, DB, shell, file, transform) + data flow + TUI
Kestra / Airflow       →  powerful, but needs servers and infrastructure
```

## Quick Example

```yaml
name: check-user-orders
description: Fetch user from API, query their orders, write a summary

inputs:
  - name: user_id
    type: string
    prompt: "User ID"

steps:
  - id: user
    name: Fetch user
    type: http
    config:
      url: "https://api.example.com/users/{{inputs.user_id}}"

  - id: orders
    name: Get orders from DB
    type: db
    config:
      connection: "{{profile.db_connection}}"
      sql: "SELECT * FROM orders WHERE user_id = $1"
      params:
        - "{{inputs.user_id}}"

  - id: summary
    name: Write summary
    type: file
    config:
      operation: write
      path: "/tmp/{{user.name}}-orders.txt"
      content: |
        User: {{user.name}} ({{user.email}})
        Total orders: {{orders.count}}
```

Run it:
```bash
./atem
```

The TUI prompts for inputs, executes steps sequentially, and shows progress. Each step's output is available to subsequent steps via `{{step_id.field}}`.

## Features

- **YAML workflows** — Version-controlled, reviewable, shareable
- **Built-in nodes** — HTTP, database (Postgres/MySQL/SQLite), shell, file, delay, transform
- **Data flow** — Reference previous step outputs with `{{step_id.field}}`
- **Interactive TUI** — Prompts for inputs, shows execution progress
- **Profiles** — Switch between dev/staging/prod configs
- **Single binary** — No runtime dependencies, no containers, no servers

## Installation

### Using Go

```bash
go install github.com/nameer-kp/atem/cmd/atem@latest
```

### From Source

```bash
git clone https://github.com/nameer-kp/atem.git
cd atem
make build
./atem
```

### Download Binary

Download the latest release from the [Releases page](https://github.com/nameer-kp/atem/releases).

```bash
# macOS (Apple Silicon)
curl -Lo atem.tar.gz https://github.com/nameer-kp/atem/releases/latest/download/atem_Darwin_arm64.tar.gz
tar -xzf atem.tar.gz
chmod +x atem
sudo mv atem /usr/local/bin/

# macOS (Intel)
curl -Lo atem.tar.gz https://github.com/nameer-kp/atem/releases/latest/download/atem_Darwin_amd64.tar.gz

# Linux (amd64)
curl -Lo atem.tar.gz https://github.com/nameer-kp/atem/releases/latest/download/atem_Linux_amd64.tar.gz
```

### Verify Installation

```bash
atem --version
```

## Node Types

| Node | Purpose | Example Use |
|------|---------|-------------|
| `http` | HTTP requests | Call APIs, webhooks |
| `db` | Database queries | Query/update Postgres, MySQL, SQLite |
| `shell` | Run commands | Execute scripts, CLI tools |
| `file` | File operations | Read configs, write reports |
| `delay` | Pause execution | Wait for external processes |
| `transform` | Data manipulation | Filter, map, extract fields |

## Configuration

Atem uses a layered config system:

```
~/.atem/
├── config.yaml           # Global settings
├── profiles/             # Environment-specific configs (dev, prod)
│   └── production.yaml
└── workflows/            # Shared workflows

project/
├── .atem.yaml         # Project settings
├── .env                   # Environment variables
└── workflows/            # Project workflows
```

Profiles let you define environment-specific variables and secrets:

```yaml
# ~/.atem/profiles/production.yaml
name: production
variables:
  base_url: https://api.example.com
  db_connection: postgres://user:pass@prod-db:5432/app
secrets:
  api_key:
    env: PROD_API_KEY
```

## Template Variables

Reference data anywhere in your workflow:

```yaml
{{inputs.user_id}}          # User input
{{profile.base_url}}        # Profile variable
{{env.HOME}}                # Environment variable
{{fetch_user.email}}        # Previous step output
{{query.rows[0].name}}      # Nested field access
```

## Comparison with Alternatives

| Tool | Type | Strengths | Atem Advantage |
|------|------|-----------|-------------------|
| **Task** | Task runner | Simple, fast, Go-based | Built-in HTTP/DB nodes, TUI |
| **Just** | Command runner | Clean syntax | Data flow between steps |
| **Make** | Build tool | Universal | Not limited to shell commands |
| **Ansible** | Config mgmt | Powerful, agentless | Lighter weight, dev-focused |
| **Kestra** | Orchestration | Feature-rich | No server required |

## Use Cases

- **API automation** — Chain API calls with data transformation
- **Database tasks** — Query, update, migrate with templated SQL
- **Dev environment setup** — Fetch configs, run setup scripts
- **Testing workflows** — End-to-end flows across services
- **Ops chores** — Anything you'd otherwise click through a UI to do

## Documentation

See [docs/WIKI.md](docs/WIKI.md) for complete documentation including:
- All node types and their options
- Configuration file reference
- Template syntax
- Complete workflow examples

## License

MIT
