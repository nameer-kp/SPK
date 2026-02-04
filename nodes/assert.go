package nodes

import (
	"fmt"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/nameer-kp/atem/pkg/node"
)

// AssertNode provides testing and validation capabilities
type AssertNode struct{}

func NewAssertNode() *AssertNode {
	return &AssertNode{}
}

func (n *AssertNode) Name() string {
	return "assert"
}

func (n *AssertNode) Description() string {
	return "Assert conditions for testing and validation"
}

func (n *AssertNode) ConfigSchema() map[string]any {
	return map[string]any{
		"that":     "string", // expression that should evaluate to true
		"equals":   "any",    // assert actual equals expected
		"actual":   "any",    // actual value for equals
		"expected": "any",    // expected value for equals
		"contains": "any",    // assert container contains value
		"in":       "any",    // container for contains check
		"type":     "string", // assert type (number, string, bool, array, object, nil)
		"value":    "any",    // value to type-check
		"message":  "string", // custom failure message
		"vars":     "object", // variables for expression
		"not":      "bool",   // negate the assertion
	}
}

func (n *AssertNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	message, _ := config["message"].(string)
	negate, _ := config["not"].(bool)

	var passed bool
	var assertDesc string

	// Expression-based assertion
	if thatExpr, ok := config["that"].(string); ok && thatExpr != "" {
		env := make(map[string]any)
		if vars, ok := config["vars"].(map[string]any); ok {
			for k, v := range vars {
				env[k] = v
			}
		}

		program, err := expr.Compile(thatExpr, expr.Env(env))
		if err != nil {
			return node.Result{}, fmt.Errorf("failed to compile assertion: %w", err)
		}

		result, err := expr.Run(program, env)
		if err != nil {
			return node.Result{}, fmt.Errorf("failed to evaluate assertion: %w", err)
		}

		boolResult, ok := result.(bool)
		if !ok {
			return node.Result{}, fmt.Errorf("assertion must evaluate to boolean, got %T", result)
		}
		passed = boolResult
		assertDesc = fmt.Sprintf("assert that: %s", thatExpr)

	} else if _, hasEquals := config["equals"]; hasEquals {
		// Simple equals assertion
		actual := config["actual"]
		expected := config["equals"]
		if exp, ok := config["expected"]; ok {
			expected = exp
		}
		passed = deepEqual(actual, expected)
		assertDesc = fmt.Sprintf("assert %v == %v", actual, expected)

	} else if containsVal, hasContains := config["contains"]; hasContains {
		// Contains assertion
		container := config["in"]
		if container == nil {
			container = config["value"]
		}
		passed = containsValue(container, containsVal)
		assertDesc = fmt.Sprintf("assert %v contains %v", container, containsVal)

	} else if typeCheck, hasType := config["type"].(string); hasType {
		// Type assertion
		value := config["value"]
		passed = checkType(value, typeCheck)
		assertDesc = fmt.Sprintf("assert type(%v) == %s", value, typeCheck)

	} else {
		return node.Result{}, fmt.Errorf("assertion requires 'that', 'equals', 'contains', or 'type'")
	}

	// Apply negation
	if negate {
		passed = !passed
		assertDesc = "NOT: " + assertDesc
	}

	status := "PASS"
	if !passed {
		status = "FAIL"
		if message == "" {
			message = assertDesc + " failed"
		}
	}

	result := node.Result{
		Success: passed,
		Data: map[string]any{
			"passed":  passed,
			"message": message,
		},
		Status: status,
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("%s: %s", status, assertDesc)},
		},
	}

	if !passed {
		result.Logs = append(result.Logs, node.LogEntry{
			Level:   "error",
			Message: message,
		})
		return result, fmt.Errorf("assertion failed: %s", message)
	}

	return result, nil
}

func deepEqual(a, b any) bool {
	// Handle numeric comparison
	if aNum, aOk := toFloat64(a); aOk {
		if bNum, bOk := toFloat64(b); bOk {
			return aNum == bNum
		}
	}

	// Use reflect.DeepEqual for complex types
	return reflect.DeepEqual(a, b)
}

func checkType(value any, expected string) bool {
	if value == nil {
		return expected == "nil" || expected == "null"
	}

	switch expected {
	case "number", "numeric":
		_, ok := toFloat64(value)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "bool", "boolean":
		_, ok := value.(bool)
		return ok
	case "array", "list":
		_, ok := value.([]any)
		return ok
	case "object", "map":
		_, ok := value.(map[string]any)
		return ok
	case "nil", "null":
		return value == nil
	default:
		return false
	}
}
