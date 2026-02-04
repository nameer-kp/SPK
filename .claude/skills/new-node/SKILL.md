---
name: new-node
description: Scaffold a new Atem node type with proper interface implementation
arguments:
  - name: node_name
    description: Name of the new node (e.g., "kafka", "redis")
    required: true
disable-model-invocation: true
---

# New Node Scaffolding

Create a new Atem node type called `{{node_name}}`.

## Steps

1. **Create the node file** at `nodes/{{node_name}}.go`

2. **Implement the `pkg/node/Node` interface**:
   - `Type() string` - returns the node type name used in workflow YAML
   - `Execute(ctx context.Context, config map[string]interface{}) (node.Result, error)` - executes the node logic
   - `ValidateConfig(config map[string]interface{}) error` - validates configuration before execution

3. **Follow existing patterns** from `nodes/http.go` or `nodes/shell.go`:
   - Extract config values with type assertions
   - Return `node.Result{Data: ...}` on success
   - Return descriptive errors for invalid configs

4. **Register the node** in `cmd/atem/main.go`:
   ```go
   registry.Register(&nodes.{{NodeName}}Node{})
   ```

## Reference

See CLAUDE.md for architecture details and data flow patterns.

## Template Structure

```go
package nodes

import (
	"context"
	"fmt"

	"github.com/nameer-kp/atem/pkg/node"
)

type {{NodeName}}Node struct{}

func (n *{{NodeName}}Node) Type() string {
	return "{{node_name}}"
}

func (n *{{NodeName}}Node) ValidateConfig(config map[string]interface{}) error {
	// Validate required config fields
	return nil
}

func (n *{{NodeName}}Node) Execute(ctx context.Context, config map[string]interface{}) (node.Result, error) {
	// Implementation here
	return node.Result{
		Data: map[string]interface{}{},
	}, nil
}
```
