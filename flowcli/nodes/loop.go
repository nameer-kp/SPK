package nodes

import (
	"fmt"
	"strings"

	"github.com/nameer-kp/flowcli/pkg/node"
)

// LoopNode iterates over an array and executes nested steps for each item
type LoopNode struct {
	registry *node.Registry // For resolving nested step node types (future use)
}

// NewLoopNode creates a new loop node with the given registry
func NewLoopNode(registry *node.Registry) *LoopNode {
	return &LoopNode{
		registry: registry,
	}
}

func (n *LoopNode) Name() string {
	return "loop"
}

func (n *LoopNode) Description() string {
	return "Iterate over an array and execute nested steps"
}

func (n *LoopNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"items": "string", // Template expression for array
		"as":    "string", // Variable name for current item (default: "item")
		"index": "string", // Optional index variable name
		"steps": "array",  // Nested steps to execute (not implemented yet, for future)
	}
}

func (n *LoopNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	// Get items expression
	itemsExpr, ok := config["items"].(string)
	if !ok || itemsExpr == "" {
		return node.Result{}, fmt.Errorf("items is required")
	}

	// Get variable names
	asVar := "item"
	if as, ok := config["as"].(string); ok && as != "" {
		asVar = as
	}

	indexVar := ""
	if idx, ok := config["index"].(string); ok {
		indexVar = idx
	}

	// Resolve items from context
	key := resolveItemsKey(itemsExpr)

	var value interface{}
	var found bool

	// First try to get the full key directly
	value, found = ctx.Get(key)

	// If not found and key contains dots, try nested object access
	if !found && strings.Contains(key, ".") {
		parts := strings.SplitN(key, ".", 2)
		baseKey := parts[0]
		path := parts[1]

		baseValue, baseFound := ctx.Get(baseKey)
		if !baseFound {
			return node.Result{}, fmt.Errorf("items '%s' not found in context", baseKey)
		}

		resolved, err := resolveNestedPath(baseValue, path)
		if err != nil {
			return node.Result{}, fmt.Errorf("failed to resolve '%s': %w", key, err)
		}
		value = resolved
		found = true
	}

	if !found {
		return node.Result{}, fmt.Errorf("items '%s' not found in context", key)
	}

	// Convert to array
	items, err := toSlice(value)
	if err != nil {
		return node.Result{}, fmt.Errorf("items must be an array: %w", err)
	}

	ctx.Logger().Info("starting loop", "items_count", len(items), "as", asVar)

	// Track iterations
	iterations := make([]map[string]interface{}, 0, len(items))

	for i, item := range items {
		// Set loop variables in context
		ctx.Set(asVar, item)
		if indexVar != "" {
			ctx.Set(indexVar, i)
		}

		// Record the iteration
		iteration := map[string]interface{}{
			"index": i,
			"item":  item,
		}
		iterations = append(iterations, iteration)

		ctx.Logger().Debug("loop iteration", "index", i, "item", item)

		// Note: Full nested step execution will be added when engine integration is complete
		// For now, we just track the iterations
	}

	return node.Result{
		Success: true,
		Data: map[string]interface{}{
			"iterations": iterations,
			"count":      len(items),
		},
		Status: fmt.Sprintf("%d iterations", len(items)),
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Completed %d iterations", len(items))},
		},
	}, nil
}

// resolveItemsKey strips the {{ }} wrapper if present and returns the key
func resolveItemsKey(expr string) string {
	expr = strings.TrimSpace(expr)

	// Strip {{ }} wrapper if present
	if strings.HasPrefix(expr, "{{") && strings.HasSuffix(expr, "}}") {
		expr = strings.TrimPrefix(expr, "{{")
		expr = strings.TrimSuffix(expr, "}}")
		expr = strings.TrimSpace(expr)
	}

	return expr
}

// resolveNestedPath resolves a dot-separated path in an object
func resolveNestedPath(value interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := value

	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("key '%s' not found", part)
			}
		default:
			return nil, fmt.Errorf("cannot traverse '%s' in non-object type %T", part, current)
		}
	}

	return current, nil
}

// toSlice converts various array types to []interface{}
func toSlice(value interface{}) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case []string:
		result := make([]interface{}, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result, nil
	case []int:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil
	case []float64:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil
	case []map[string]interface{}:
		result := make([]interface{}, len(v))
		for i, m := range v {
			result[i] = m
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected array, got %T", value)
	}
}
