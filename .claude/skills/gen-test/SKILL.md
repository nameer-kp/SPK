---
name: gen-test
description: Generate Go tests for a package or file
arguments:
  - name: target
    description: Package path or file to test (e.g., ./nodes, nodes/http.go)
    required: true
disable-model-invocation: true
---

# Generate Go Tests

Generate comprehensive tests for `{{target}}`.

## Guidelines

1. **Use table-driven tests** with `t.Run()` subtests:
   ```go
   tests := []struct {
       name    string
       input   Type
       want    Type
       wantErr bool
   }{
       // test cases
   }
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           // test logic
       })
   }
   ```

2. **Test coverage priorities**:
   - Happy path (valid inputs, expected outputs)
   - Error cases (invalid inputs, edge cases)
   - Boundary conditions

3. **For node tests**, mock external dependencies:
   - HTTP nodes: use `httptest.NewServer`
   - DB nodes: use interfaces for mockability
   - File nodes: use `t.TempDir()`

4. **File placement**: Create `*_test.go` adjacent to source file

5. **Naming conventions**:
   - Test functions: `TestFunctionName`
   - Test files: `filename_test.go`

## Example Test Structure

```go
package nodes

import (
	"context"
	"testing"
)

func TestNodeType_Execute(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"key": "value",
			},
			want:    expectedResult,
			wantErr: false,
		},
		{
			name:    "missing required field",
			config:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeType{}
			got, err := n.Execute(context.Background(), tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Data != tt.want {
				t.Errorf("Execute() = %v, want %v", got.Data, tt.want)
			}
		})
	}
}
```
