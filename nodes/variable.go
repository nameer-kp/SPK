package nodes

import (
	"fmt"
	"sync"

	"github.com/nameer-kp/atem/pkg/node"
)

// Global variable store (workflow-scoped in practice via context)
var (
	varStore   = make(map[string]any)
	varStoreMu sync.RWMutex
)

// VariableNode manages workflow variables (state)
type VariableNode struct{}

func NewVariableNode() *VariableNode {
	return &VariableNode{}
}

func (n *VariableNode) Name() string {
	return "variable"
}

func (n *VariableNode) Description() string {
	return "Set, get, and manage workflow variables"
}

func (n *VariableNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // set, get, delete, increment, decrement, append, list
		"name":      "string", // variable name
		"value":     "any",    // value to set
		"default":   "any",    // default if not exists (for get)
		"by":        "number", // increment/decrement amount (default 1)
	}
}

func (n *VariableNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "set"
	}

	name, _ := config["name"].(string)

	switch operation {
	case "set":
		if name == "" {
			return node.Result{}, fmt.Errorf("variable name is required for set")
		}
		value := config["value"]
		varStoreMu.Lock()
		varStore[name] = value
		varStoreMu.Unlock()

		ctx.Set(name, value)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":  name,
				"value": value,
			},
			Status: fmt.Sprintf("set %s", name),
			Logs: []node.LogEntry{
				{Level: "info", Message: fmt.Sprintf("Set %s = %v", name, value)},
			},
		}, nil

	case "get":
		if name == "" {
			return node.Result{}, fmt.Errorf("variable name is required for get")
		}

		varStoreMu.RLock()
		value, exists := varStore[name]
		varStoreMu.RUnlock()

		if !exists {
			if ctxVal, ok := ctx.Get(name); ok {
				value = ctxVal
				exists = true
			}
		}

		if !exists {
			if defaultVal, hasDefault := config["default"]; hasDefault {
				value = defaultVal
				exists = true
			}
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":   name,
				"value":  value,
				"exists": exists,
			},
			Status: fmt.Sprintf("get %s", name),
			Logs: []node.LogEntry{
				{Level: "info", Message: fmt.Sprintf("Get %s = %v (exists: %v)", name, value, exists)},
			},
		}, nil

	case "delete":
		if name == "" {
			return node.Result{}, fmt.Errorf("variable name is required for delete")
		}

		varStoreMu.Lock()
		_, existed := varStore[name]
		delete(varStore, name)
		varStoreMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":    name,
				"deleted": existed,
			},
			Status: fmt.Sprintf("delete %s", name),
			Logs: []node.LogEntry{
				{Level: "info", Message: fmt.Sprintf("Deleted %s (existed: %v)", name, existed)},
			},
		}, nil

	case "increment", "decrement":
		if name == "" {
			return node.Result{}, fmt.Errorf("variable name is required for %s", operation)
		}

		by := 1.0
		if b, ok := config["by"].(float64); ok {
			by = b
		}
		if operation == "decrement" {
			by = -by
		}

		varStoreMu.Lock()
		current, exists := varStore[name]
		var newValue float64
		if exists {
			if v, ok := toFloat64(current); ok {
				newValue = v + by
			} else {
				varStoreMu.Unlock()
				return node.Result{}, fmt.Errorf("variable %s is not numeric", name)
			}
		} else {
			newValue = by
		}
		varStore[name] = newValue
		varStoreMu.Unlock()

		ctx.Set(name, newValue)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":     name,
				"value":    newValue,
				"previous": current,
			},
			Status: fmt.Sprintf("%s %s", operation, name),
			Logs: []node.LogEntry{
				{Level: "info", Message: fmt.Sprintf("%s %s: %v -> %v", operation, name, current, newValue)},
			},
		}, nil

	case "append":
		if name == "" {
			return node.Result{}, fmt.Errorf("variable name is required for append")
		}
		value := config["value"]

		varStoreMu.Lock()
		current, exists := varStore[name]
		var newValue []any
		if exists {
			if arr, ok := current.([]any); ok {
				newValue = append(arr, value)
			} else {
				newValue = []any{current, value}
			}
		} else {
			newValue = []any{value}
		}
		varStore[name] = newValue
		varStoreMu.Unlock()

		ctx.Set(name, newValue)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":   name,
				"value":  newValue,
				"length": len(newValue),
			},
			Status: fmt.Sprintf("append to %s", name),
			Logs: []node.LogEntry{
				{Level: "info", Message: fmt.Sprintf("Appended to %s, length: %d", name, len(newValue))},
			},
		}, nil

	case "list":
		varStoreMu.RLock()
		names := make([]string, 0, len(varStore))
		snapshot := make(map[string]any, len(varStore))
		for k, v := range varStore {
			names = append(names, k)
			snapshot[k] = v
		}
		varStoreMu.RUnlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"variables": snapshot,
				"names":     names,
				"count":     len(names),
			},
			Status: fmt.Sprintf("%d variables", len(names)),
			Logs: []node.LogEntry{
				{Level: "info", Message: fmt.Sprintf("Listed %d variables", len(names))},
			},
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s (valid: set, get, delete, increment, decrement, append, list)", operation)
	}
}

// ResetVariables clears the variable store (for testing)
func ResetVariables() {
	varStoreMu.Lock()
	varStore = make(map[string]any)
	varStoreMu.Unlock()
}
