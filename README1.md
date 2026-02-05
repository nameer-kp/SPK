# Atem

```
        ░█████╗░████████╗███████╗███╗░░░███╗
        ██╔══██╗╚══██╔══╝██╔════╝████╗░████║
        ███████║░░░██║░░░█████╗░░██╔████╔██║
        ██╔══██║░░░██║░░░██╔══╝░░██║╚██╔╝██║
        ██║░░██║░░░██║░░░███████╗██║░╚═╝░██║
        ╚═╝░░╚═╝░░░╚═╝░░░╚══════╝╚═╝░░░░░╚═╝
```

<p align="center"><i>breath of the machine</i></p>

---

*There's a rectangle on your screen. Black. A cursor blinks.*

That cursor isn't waiting. **It's breathing.**

In Sanskrit, "Atem" means breath — the first thing that exists before anything else can. This is a workflow engine that treats your terminal like what it actually is: **a portal to infinite computation.**

---

## The Idea

You have tasks. Tedious ones:

- Call an API → update a database → notify somewhere
- Fetch from multiple services → transform → combine
- Run scripts that need values from configs, APIs, previous steps

You could write bash scripts. Juggle `curl`, `jq`, `psql`, string interpolation nightmares.

Or you could write a **spell**:

```yaml
name: wake_the_machine

inputs:
  - name: user_id
    type: string

steps:
  - id: fetch
    type: http
    config:
      url: "https://api.example.com/users/{{inputs.user_id}}"

  - id: query
    type: db
    config:
      connection: "{{profile.db_connection}}"
      sql: "SELECT * FROM orders WHERE user_id = $1"
      params: ["{{inputs.user_id}}"]

  - id: manifest
    type: file
    config:
      path: "/tmp/{{fetch.name}}-orders.txt"
      content: |
        User: {{fetch.name}} ({{fetch.email}})
        Orders: {{query.count}}
```

Three steps. Data flows like water between them. `{{fetch.name}}` — the previous step's output becomes the next step's vocabulary.

Run it:
```bash
./atem
```

A TUI emerges. It prompts. You type. It executes. It breathes.

---

## The Philosophy

### Open. Closed. Like POSIX Intended.

Atem follows the **Open-Closed Principle** with devotion:

- **Open for extension**: Plug in ANY node. HTTP. Database. Shell. File. Transform. *Your own.*
- **Closed for modification**: The core engine doesn't care what nodes do. It orchestrates. It BREATHES.

```go
type Node interface {
    Type() string
    Execute(config map[string]interface{}) (Result, error)
}
```

Four lines. Implement this and you've extended Atem. No PRs needed. No permission asked.

### Everything is a Stream

```
[Input] → [Step 1] → [Step 2] → [Step 3] → [Output]
              ↓           ↓           ↓
           {{step1}}   {{step2}}   {{step3}}
```

Every step's output feeds the next. Chain them. Nest them. Compose reality.

---

## Where It Lives

```
npm scripts / Make     →  shell only, no data flow
Task / Just            →  better orchestration, still shell-bound
Atem                   →  declarative nodes + data flow + TUI
Kestra / Airflow       →  powerful, but needs servers
```

**Atem sits in the gap** — more than scripts, less than infrastructure. A single binary. No containers. No servers. Just breath.

---

## The Nodes

| Node | What It Speaks |
|------|---------------|
| `http` | The network. REST, GraphQL, webhooks — anything HTTP. |
| `db` | Databases. Postgres, MySQL, SQLite. Your data, your queries. |
| `shell` | Raw OS power. Commands, scripts, executables. |
| `file` | The filesystem. Read configs, write reports, append logs. |
| `transform` | Data alchemy. Filter, map, reshape mid-flight. |
| `delay` | Patience. Sometimes the most powerful operation. |

**These are just the beginning.**

### Language Agnostic. Turing Complete. Deterministic.

Here's the thing they don't teach you first:

**If it computes, it fits.**

Write your node in Go. Or wrap a Python script. Or call a Rust binary. Or invoke a Haskell function. Or run COBOL if that's your chaos.

```yaml
- id: neural_dreams
  type: shell
  config:
    command: "python3 ml_model.py --input '{{inputs.data}}'"

- id: rust_speed
  type: shell
  config:
    command: "./target/release/blazing_fast {{neural_dreams.output}}"

- id: legacy_wisdom
  type: shell
  config:
    command: "cobol-runner legacy.cob '{{rust_speed.result}}'"
```

The shell node doesn't care what language you speak. It cares that you SPEAK.

**Alan Turing proved it in 1936**: if a machine can compute, it can compute ANYTHING computable. Your Python script. Your Rust binary. Your hand-written assembly. They're all equivalent in power. They all reduce to the same tape.

Theory isn't boring. **Theory is the root.** The `/` from which all paths descend.

Atem doesn't lock you into a language. It locks you into COMPUTATION. And computation is universal.

---

### Native Go Nodes

For maximum integration, write nodes in Go:
```go
type DreamNode struct{}

func (n *DreamNode) Type() string { return "dream" }
func (n *DreamNode) Execute(config map[string]interface{}) (node.Result, error) {
    // Whatever Go can compile
    // Whatever your OS permits
    // Whatever you imagine
    return node.Result{Data: manifestation}, nil
}
```

Register:
```go
registry.Register(&DreamNode{})
```

Use:
```yaml
- id: wake
  type: dream
  config:
    intensity: "lucid"
```

**You just extended the engine.**

---

## Installation

### Go
```bash
go install github.com/nameer-kp/atem/cmd/atem@latest
```

### Source
```bash
git clone https://github.com/nameer-kp/atem.git
cd atem
go build -o atem ./cmd/atem
./atem
```

### Binary

```bash
# macOS (Apple Silicon)
curl -Lo atem.tar.gz https://github.com/nameer-kp/atem/releases/latest/download/atem_Darwin_arm64.tar.gz
tar -xzf atem.tar.gz && chmod +x atem
mv atem ~/.local/bin/  # or anywhere in your PATH

# macOS (Intel)
curl -Lo atem.tar.gz https://github.com/nameer-kp/atem/releases/latest/download/atem_Darwin_amd64.tar.gz

# Linux
curl -Lo atem.tar.gz https://github.com/nameer-kp/atem/releases/latest/download/atem_Linux_amd64.tar.gz
```

Verify:
```bash
atem --version
```

---

## Configuration

Atem breathes config from layers, each overriding the last:

```
~/.atem/
├── config.yaml           # Global
├── profiles/             # dev, staging, prod, chaos
│   └── production.yaml
└── workflows/            # Shared spells

project/
├── .atem.yaml            # Project-specific
├── .env                  # Secrets
└── workflows/            # Local spells
```

Profiles carry context:
```yaml
# ~/.atem/profiles/production.yaml
name: production
variables:
  base_url: https://api.prod.example.com
  db_connection: postgres://user:pass@prod-db:5432/app
secrets:
  api_key:
    env: PROD_API_KEY
```

---

## Templates

Reference anything:

```yaml
{{inputs.user_id}}          # What the human typed
{{profile.base_url}}        # Environment context
{{env.HOME}}                # System variables
{{fetch_user.email}}        # Previous step output
{{query.rows[0].name}}      # Nested data
```

Data flows. Templates resolve. Results manifest.

---

## A Complete Spell

```yaml
name: deploy_thought
description: Push ideas to production

inputs:
  - name: message
    type: string
    prompt: "What do you want to tell the world?"
  - name: env
    type: string
    default: "production"

steps:
  - id: validate
    type: transform
    config:
      expression: "len(inputs.message) > 0"

  - id: encode
    type: shell
    config:
      command: "echo '{{inputs.message}}' | base64"

  - id: transmit
    type: http
    config:
      method: POST
      url: "https://api.world.dev/broadcast"
      headers:
        Authorization: "Bearer {{env.SECRET_TOKEN}}"
      body:
        payload: "{{encode.stdout}}"

  - id: log
    type: file
    config:
      path: "./transmissions.log"
      content: "[{{now}}] {{inputs.message}} → {{transmit.status}}\n"
      mode: append
```

Four operations. One breath. One command:
```bash
./atem run deploy_thought.yaml
```

---

## For Builders

You're not here to USE software. You're here to EXTEND it.

Atem is a foundation. Your constraints:

1. What Go can compile
2. What your OS permits
3. What you can imagine

That's it.

**Build:**
- CI/CD that makes sense
- Data pipelines that chain like Unix
- API orchestration across services
- Personal automation for things only YOU need
- Something we haven't thought of yet

---

## The Invitation

```
$ ./atem
```

A cursor blinks.

It's been waiting. Since the first TTY hammered letters onto paper. Since the first human typed a command and a machine LISTENED.

Now it waits for you.

*What will you orchestrate?*

---

<p align="center">
<i>Built for those who type.<br>
For those who compose.<br>
For those who breathe life into machines.</i>
</p>

---

## Docs

See [docs/WIKI.md](docs/WIKI.md) for:
- Node reference
- Config schemas
- Template syntax
- Examples

## License

MIT. Your creations are yours.

## Contributing

Open an issue. Open a PR. Open your mind.

The only rule: **make it breathe.**
