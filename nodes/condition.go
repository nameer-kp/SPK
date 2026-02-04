package nodes

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/nameer-kp/atem/pkg/node"
)

// ConditionNode evaluates conditions and provides branching logic
type ConditionNode struct{}

func NewConditionNode() *ConditionNode {
	return &ConditionNode{}
}

func (n *ConditionNode) Name() string {
	return "condition"
}

func (n *ConditionNode) Description() string {
	return "Evaluate conditions for branching logic"
}

func (n *ConditionNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"if":      "string", // expression to evaluate (using expr-lang)
		"then":    "any",    // value if true
		"else":    "any",    // value if false
		"vars":    "object", // variables for expression evaluation
		"compare": "object", // simple comparison: {left, op, right}
	}
}

func (n *ConditionNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	var conditionResult bool

	// Method 1: expr-lang expression
	if ifExpr, ok := config["if"].(string); ok && ifExpr != "" {
		// Build environment from vars
		env := make(map[string]interface{})
		if vars, ok := config["vars"].(map[string]interface{}); ok {
			for k, v := range vars {
				env[k] = v
			}
		}

		// Compile and run expression
		program, err := expr.Compile(ifExpr, expr.Env(env))
		if err != nil {
			return node.Result{}, fmt.Errorf("failed to compile condition: %w", err)
		}

		result, err := expr.Run(program, env)
		if err != nil {
			return node.Result{}, fmt.Errorf("failed to evaluate condition: %w", err)
		}

		var ok bool
		conditionResult, ok = result.(bool)
		if !ok {
			return node.Result{}, fmt.Errorf("condition must evaluate to boolean, got %T", result)
		}

	} else if compare, ok := config["compare"].(map[string]interface{}); ok {
		// Method 2: Simple comparison
		left := compare["left"]
		right := compare["right"]
		op, _ := compare["op"].(string)

		conditionResult = evaluateComparison(left, op, right)
	} else {
		return node.Result{}, fmt.Errorf("condition requires 'if' expression or 'compare' object")
	}

	// Determine result value
	var resultValue interface{}
	if conditionResult {
		resultValue = config["then"]
	} else {
		resultValue = config["else"]
	}

	branch := "else"
	if conditionResult {
		branch = "then"
	}

	return node.Result{
		Success: true,
		Data: map[string]interface{}{
			"result": conditionResult,
			"value":  resultValue,
			"branch": branch,
		},
		Status: fmt.Sprintf("branch: %s", branch),
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Condition evaluated to %v, taking %s branch", conditionResult, branch)},
		},
	}, nil
}

func evaluateComparison(left interface{}, op string, right interface{}) bool {
	// Convert to float64 for numeric comparisons
	leftNum, leftIsNum := toFloat64(left)
	rightNum, rightIsNum := toFloat64(right)

	switch op {
	case "==", "eq":
		if leftIsNum && rightIsNum {
			return leftNum == rightNum
		}
		return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right)

	case "!=", "ne":
		if leftIsNum && rightIsNum {
			return leftNum != rightNum
		}
		return fmt.Sprintf("%v", left) != fmt.Sprintf("%v", right)

	case ">", "gt":
		if leftIsNum && rightIsNum {
			return leftNum > rightNum
		}
		return fmt.Sprintf("%v", left) > fmt.Sprintf("%v", right)

	case ">=", "gte":
		if leftIsNum && rightIsNum {
			return leftNum >= rightNum
		}
		return fmt.Sprintf("%v", left) >= fmt.Sprintf("%v", right)

	case "<", "lt":
		if leftIsNum && rightIsNum {
			return leftNum < rightNum
		}
		return fmt.Sprintf("%v", left) < fmt.Sprintf("%v", right)

	case "<=", "lte":
		if leftIsNum && rightIsNum {
			return leftNum <= rightNum
		}
		return fmt.Sprintf("%v", left) <= fmt.Sprintf("%v", right)

	case "contains":
		return containsValue(left, right)

	case "startswith":
		leftStr := fmt.Sprintf("%v", left)
		rightStr := fmt.Sprintf("%v", right)
		return len(leftStr) >= len(rightStr) && leftStr[:len(rightStr)] == rightStr

	case "endswith":
		leftStr := fmt.Sprintf("%v", left)
		rightStr := fmt.Sprintf("%v", right)
		return len(leftStr) >= len(rightStr) && leftStr[len(leftStr)-len(rightStr):] == rightStr

	default:
		return false
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	default:
		return 0, false
	}
}

func containsValue(container, value interface{}) bool {
	switch c := container.(type) {
	case string:
		valueStr := fmt.Sprintf("%v", value)
		for i := 0; i <= len(c)-len(valueStr); i++ {
			if c[i:i+len(valueStr)] == valueStr {
				return true
			}
		}
		return false
	case []interface{}:
		for _, item := range c {
			if fmt.Sprintf("%v", item) == fmt.Sprintf("%v", value) {
				return true
			}
		}
		return false
	case map[string]interface{}:
		key := fmt.Sprintf("%v", value)
		_, exists := c[key]
		return exists
	default:
		return false
	}
}
