# Atem Wiki

Atem is a workflow orchestration and automation framework. Define multi-step workflows in YAML and execute them with data flowing between steps.

## Table of Contents

- [Getting Started](#getting-started)
- [Configuration Files](#configuration-files)
  - [Global Config](#global-config)
  - [Project Config](#project-config)
  - [Profiles](#profiles)
- [Workflow YAML](#workflow-yaml)
  - [Inputs](#inputs)
  - [Steps](#steps)
  - [Template Variables](#template-variables)
- [Node Types](#node-types)
  - [http](#http-node)
  - [db](#db-node)
  - [shell](#shell-node)
  - [file](#file-node)
  - [delay](#delay-node)
  - [transform](#transform-node)

---

## Getting Started

### Installation

```bash
# Build from source
go build -o atem ./cmd/atem

# Run
./atem
```

### Directory Structure

```
~/.atem/
├── config.yaml              # Global configuration
├── profiles/                # Named environment profiles
│   └── production.yaml
└── workflows/               # Default workflows directory

project/
├── .atem.yaml            # Project configuration (optional)
├── .env                     # Environment variables (optional)
└── workflows/               # Project workflows
```

---

## Configuration Files

Configuration is loaded in priority order (highest wins):
1. Environment variables (`FLOWCLI_PROFILE`)
2. Project config (`.atem.yaml`)
3. Profile (`~/.atem/profiles/{name}.yaml`)
4. Global config (`~/.atem/config.yaml`)
5. `.env` file

### Global Config

**Location:** `~/.atem/config.yaml`

```yaml
# Default profile to use when none specified
default_profile: development

# Editor for workflow editing (used by TUI)
editor: vim

# Logging level: debug, info, warn, error
log_level: info

# Directory for plugins
plugins_dir: ~/.atem/plugins

# Directory for execution checkpoints
checkpoints_dir: ~/.atem/checkpoints

# Default directory for workflows
workflows_dir: ~/.atem/workflows
```

| Field | Type | Description |
|-------|------|-------------|
| `default_profile` | string | Profile to use when none specified |
| `editor` | string | Text editor command |
| `log_level` | string | Logging verbosity (debug/info/warn/error) |
| `plugins_dir` | string | Directory containing plugins |
| `checkpoints_dir` | string | Directory for saving execution state |
| `workflows_dir` | string | Default workflows directory |

### Project Config

**Location:** `.atem.yaml` (project root)

```yaml
# Override default profile for this project
default_profile: staging

# Project-specific workflows directory
workflows_dir: ./workflows

# Project-specific plugins directory
plugins_dir: ./plugins

# Project-level variables (available in all workflows)
variables:
  api_version: v2
  timeout: "30s"
```

| Field | Type | Description |
|-------|------|-------------|
| `default_profile` | string | Override global default profile |
| `workflows_dir` | string | Override workflows directory |
| `plugins_dir` | string | Override plugins directory |
| `variables` | map[string]string | Project-level variables |

### Profiles

**Location:** `~/.atem/profiles/{name}.yaml`

Profiles define environment-specific configuration (dev, staging, production).

```yaml
name: production
description: Production Environment

# Variables available via {{profile.var_name}}
variables:
  base_url: https://api.example.com
  api_version: v2
  max_retries: "3"

# Secrets - obtained from env vars or user prompts
secrets:
  api_key:
    env: PROD_API_KEY          # Read from environment variable
  db_password:
    prompt: "Enter DB password" # Prompt user at runtime

# Optional database configuration
database:
  driver: postgres             # postgres, mysql, sqlite
  host: db.example.com
  port: 5432
  name: myapp
  user: admin
  password: "{{secrets.db_password}}"
```

#### Variables

```yaml
variables:
  key: value
```

Access in workflows via `{{profile.key}}`.

#### Secrets

```yaml
secrets:
  secret_name:
    env: ENV_VAR_NAME      # Read from environment variable
  another_secret:
    prompt: "Enter value"  # Prompt user for input
```

| Field | Type | Description |
|-------|------|-------------|
| `env` | string | Environment variable name to read from |
| `prompt` | string | Prompt message for user input |

#### Database Config

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  name: mydb
  user: admin
  password: secret
```

| Field | Type | Description |
|-------|------|-------------|
| `driver` | string | Database driver: `postgres`, `mysql`, `sqlite` |
| `host` | string | Database host |
| `port` | int | Database port |
| `name` | string | Database name |
| `user` | string | Database user |
| `password` | string | Database password |

---

## Workflow YAML

Workflows define the steps to execute.

```yaml
name: my-workflow
description: Description of what this workflow does
version: "1.0"

inputs:
  - name: user_id
    type: string
    prompt: "Enter user ID"
    required: true

steps:
  - id: fetch_user
    name: Fetch user data
    type: http
    config:
      method: GET
      url: "https://api.example.com/users/{{inputs.user_id}}"
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Workflow name |
| `description` | string | No | Human-readable description |
| `version` | string | No | Version string |
| `inputs` | array | No | User prompts at workflow start |
| `steps` | array | Yes | Steps to execute |

### Inputs

Inputs collect values from the user before workflow execution.

```yaml
inputs:
  - name: user_id
    type: string
    prompt: "Enter user ID"
    required: true
    default: "12345"

  - name: count
    type: number
    prompt: "How many items?"
    required: false
    default: "10"

  - name: confirm
    type: bool
    prompt: "Proceed with operation?"
    required: true

  - name: environment
    type: choice
    prompt: "Select environment"
    options:
      - development
      - staging
      - production
    required: true

  - name: api_key
    type: secret
    prompt: "Enter API key"
    required: true
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Variable name (access via `{{inputs.name}}`) |
| `type` | string | Yes | Input type: `string`, `number`, `bool`, `choice`, `secret` |
| `prompt` | string | Yes | Prompt message shown to user |
| `required` | bool | No | Whether input is required (default: false) |
| `default` | string | No | Default value if not provided |
| `options` | array | No | Available options (only for `choice` type) |

### Steps

Each step executes a node with a specific configuration.

```yaml
steps:
  - id: step_id           # Unique identifier (required)
    name: Step Name       # Human-readable name (required)
    type: http            # Node type (required)
    condition: "..."      # Condition to evaluate (optional, not yet implemented)
    config:               # Node-specific configuration (required)
      key: value
    output: result_name   # Store result as variable (optional)
    on_error: retry       # Error handling: retry, skip, abort (optional, not yet implemented)
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique step identifier |
| `name` | string | Yes | Human-readable step name |
| `type` | string | Yes | Node type to execute |
| `condition` | string | No | Condition expression (not yet implemented) |
| `config` | object | Yes | Node-specific configuration |
| `output` | string | No | Variable name to store result |
| `on_error` | string | No | Error handling strategy (not yet implemented) |

### Template Variables

Use `{{reference}}` syntax to reference values in configs.

```yaml
# User inputs
{{inputs.variable_name}}

# Profile variables
{{profile.variable_name}}

# Environment variables
{{env.ENV_VAR_NAME}}

# Previous step output (entire result)
{{step_id}}

# Nested field from step output
{{step_id.field.nested.value}}

# Step output data
{{step_id.data}}
{{step_id.stdout}}
{{step_id.rows}}
```

---

## Node Types

### HTTP Node

Execute HTTP requests.

```yaml
- id: api_call
  name: Call API
  type: http
  config:
    method: GET                    # GET, POST, PUT, DELETE, PATCH (default: GET)
    url: "https://api.example.com/endpoint"
    headers:                       # Optional headers
      Authorization: "Bearer {{inputs.token}}"
      Content-Type: application/json
    body:                          # Optional body (for POST/PUT/PATCH)
      key: value
    timeout: "30s"                 # Optional timeout (default: 30s)
    retry:                         # Optional retry config
      attempts: 3
      delay: "1s"
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `method` | string | No | GET | HTTP method |
| `url` | string | Yes | - | Request URL |
| `headers` | object | No | - | Request headers |
| `body` | any | No | - | Request body (auto-serialized to JSON) |
| `timeout` | string | No | 30s | Request timeout |
| `retry.attempts` | int | No | - | Number of retry attempts |
| `retry.delay` | string | No | - | Delay between retries |

**Output:**
```yaml
# Response parsed as JSON (or string if not JSON)
{{step_id}}           # Entire response body
{{step_id.field}}     # Access JSON field
```

---

### DB Node

Execute database queries.

```yaml
- id: query_users
  name: Query Users
  type: db
  config:
    driver: postgres              # postgres, mysql, sqlite (default: postgres)
    connection: "postgres://user:pass@host:5432/dbname?sslmode=disable"
    operation: query              # query, exec (default: query)
    sql: "SELECT * FROM users WHERE id = $1"
    params:                       # Query parameters
      - "{{inputs.user_id}}"
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `driver` | string | No | postgres | Database driver |
| `connection` | string | Yes* | - | DSN connection string |
| `operation` | string | No | query | `query` (returns rows) or `exec` (returns affected count) |
| `sql` | string | Yes | - | SQL statement |
| `params` | array | No | - | Query parameters |

*If `connection` is not provided, Atem looks for `db_connection` in profile variables.

**Connection String Formats:**

```yaml
# PostgreSQL
connection: "postgres://user:password@host:5432/dbname?sslmode=disable"

# MySQL
connection: "user:password@tcp(host:3306)/dbname"

# SQLite
connection: "/path/to/database.db"
```

**Output for `query`:**
```yaml
{{step_id.rows}}      # Array of row objects
{{step_id.count}}     # Number of rows returned
{{step_id.rows[0].column_name}}  # Access specific field
```

**Output for `exec`:**
```yaml
{{step_id.rows_affected}}   # Number of rows affected
{{step_id.last_insert_id}}  # Last inserted ID (if applicable)
```

---

### Shell Node

Execute shell commands.

```yaml
- id: run_script
  name: Run Script
  type: shell
  config:
    command: "python"             # Command to run
    args:                         # Command arguments
      - "script.py"
      - "--input"
      - "{{inputs.file}}"
    dir: "/path/to/directory"     # Working directory (optional)
    env:                          # Additional environment variables (optional)
      DEBUG: "true"
      API_KEY: "{{profile.api_key}}"
    stdin: "input data"           # Stdin input (optional)
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `command` | string | Yes | Command to execute |
| `args` | array | No | Command arguments |
| `dir` | string | No | Working directory |
| `env` | object | No | Additional environment variables |
| `stdin` | string | No | Data to pass to stdin |

**Output:**
```yaml
{{step_id.stdout}}     # Standard output
{{step_id.stderr}}     # Standard error
{{step_id.exit_code}}  # Exit code (0 = success)
```

---

### File Node

Read, write, and manage files.

```yaml
# Read file
- id: read_config
  name: Read Config
  type: file
  config:
    operation: read
    path: "{{inputs.config_path}}"
    format: json                  # text, json, yaml (default: text)

# Write file
- id: write_output
  name: Write Output
  type: file
  config:
    operation: write
    path: "/tmp/output.txt"
    content: "Result: {{previous_step.data}}"

# Append to file
- id: append_log
  name: Append Log
  type: file
  config:
    operation: append
    path: "/var/log/workflow.log"
    content: "\n{{step_id}}: completed"

# Delete file
- id: cleanup
  name: Delete Temp File
  type: file
  config:
    operation: delete
    path: "/tmp/temp_file.txt"

# Check if file exists
- id: check_exists
  name: Check File
  type: file
  config:
    operation: exists
    path: "{{inputs.file_path}}"
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `operation` | string | No | `read`, `write`, `append`, `delete`, `exists` (default: read) |
| `path` | string | Yes | File path (supports `~` expansion) |
| `content` | string | No | Content to write (for write/append) |
| `format` | string | No | Parse format for read: `text`, `json`, `yaml` (default: text) |

**Output for `read`:**
```yaml
{{step_id}}           # File content (parsed if json/yaml, string if text)
```

**Output for `write`/`append`:**
```yaml
{{step_id.bytes_written}}  # Number of bytes written
```

**Output for `delete`:**
```yaml
{{step_id.deleted}}   # true if deleted
{{step_id.existed}}   # false if file didn't exist
```

**Output for `exists`:**
```yaml
{{step_id.exists}}    # true/false
```

---

### Delay Node

Pause execution for a specified duration.

```yaml
- id: wait
  name: Wait for processing
  type: delay
  config:
    duration: "5s"                # Duration string (required)
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `duration` | string | Yes | Duration (e.g., `500ms`, `5s`, `1m`, `1h`) |

**Duration Format:**
- `500ms` - milliseconds
- `5s` - seconds
- `1m` - minutes
- `1h` - hours
- `1h30m` - combinations

**Output:**
```yaml
{{step_id.delayed}}   # The duration string
```

---

### Transform Node

Transform data using expressions.

```yaml
- id: transform_data
  name: Transform Data
  type: transform
  config:
    input: previous_step          # Reference to input data
    operations:
      - type: jsonpath
        expression: "data.users"
      - type: filter
        expression: "item.active == true"
      - type: map
        expression: "item.name"
    output: result_var            # Optional: store in context
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `input` | string | No | Variable reference for input data |
| `operations` | array | Yes | List of transformation operations |
| `output` | string | No | Variable name to store result |

#### Operation Types

**jsonpath** - Extract data using dot notation
```yaml
- type: jsonpath
  expression: "data.users[0].name"
```

**expr** - Evaluate expressions using [expr-lang](https://expr-lang.org/)
```yaml
- type: expr
  expression: "data.price * data.quantity"
```
Available variables: `data` (current data)

**template** - Substitute `{{key}}` patterns
```yaml
- type: template
  expression: "Hello, {{name}}! You have {{count}} items."
```

**map** - Transform each element of an array
```yaml
- type: map
  expression: "item.name + ' (' + item.email + ')'"
```
Available variables: `item` (current element), `index` (position), `data` (full array)

**filter** - Keep elements matching condition
```yaml
- type: filter
  expression: "item.age >= 18 && item.active"
```
Available variables: `item` (current element), `index` (position)

**pick** - Select specific fields from objects
```yaml
- type: pick
  expression: "id, name, email"     # Comma-separated field names
```

#### Transform Examples

**Extract and filter data:**
```yaml
- id: process_users
  type: transform
  config:
    input: api_response
    operations:
      - type: jsonpath
        expression: "data.users"
      - type: filter
        expression: "item.status == 'active'"
      - type: map
        expression: "{'id': item.id, 'name': item.name}"
```

**Calculate totals:**
```yaml
- id: calculate_total
  type: transform
  config:
    input: order_items
    operations:
      - type: expr
        expression: "sum(data, {.price * .quantity})"
```

**Format output:**
```yaml
- id: format_message
  type: transform
  config:
    input: user_data
    operations:
      - type: template
        expression: "User {{name}} ({{email}}) has {{order_count}} orders"
```

---

## Complete Workflow Example

```yaml
name: user-report
description: Generate a report for a user
version: "1.0"

inputs:
  - name: user_id
    type: string
    prompt: "Enter user ID"
    required: true
  - name: output_file
    type: string
    prompt: "Output file path"
    default: "/tmp/user-report.txt"

steps:
  # Fetch user from API
  - id: fetch_user
    name: Fetch User
    type: http
    config:
      method: GET
      url: "{{profile.base_url}}/users/{{inputs.user_id}}"
      headers:
        Authorization: "Bearer {{profile.api_key}}"

  # Get user's orders from database
  - id: get_orders
    name: Get Orders
    type: db
    config:
      driver: postgres
      connection: "{{profile.db_connection}}"
      operation: query
      sql: "SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT 10"
      params:
        - "{{inputs.user_id}}"

  # Transform order data
  - id: transform_orders
    name: Format Orders
    type: transform
    config:
      input: get_orders
      operations:
        - type: jsonpath
          expression: "rows"
        - type: map
          expression: "item.id + ': $' + string(item.total)"

  # Generate report
  - id: write_report
    name: Write Report
    type: file
    config:
      operation: write
      path: "{{inputs.output_file}}"
      content: |
        User Report
        ===========
        Name: {{fetch_user.name}}
        Email: {{fetch_user.email}}

        Recent Orders:
        {{transform_orders}}

  # Notify completion
  - id: notify
    name: Send Notification
    type: http
    config:
      method: POST
      url: "{{profile.webhook_url}}"
      body:
        message: "Report generated for {{fetch_user.name}}"
        file: "{{inputs.output_file}}"
```
