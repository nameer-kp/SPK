package nodes

import (
	"fmt"
	"math"

	"github.com/expr-lang/expr"
	"github.com/nameer-kp/atem/pkg/node"
)

// MathNode provides arithmetic and mathematical operations
type MathNode struct{}

func NewMathNode() *MathNode {
	return &MathNode{}
}

func (n *MathNode) Name() string {
	return "math"
}

func (n *MathNode) Description() string {
	return "Perform arithmetic and mathematical operations"
}

func (n *MathNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation":  "string", // add, sub, mul, div, mod, pow, sqrt, abs, min, max, floor, ceil, round, expr
		"a":          "number", // first operand
		"b":          "number", // second operand (for binary ops)
		"values":     "array",  // array of numbers (for min, max, sum, avg)
		"expression": "string", // expr-lang expression for complex math
		"vars":       "object", // variables for expression evaluation
		"precision":  "number", // decimal precision for rounding
	}
}

func (n *MathNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "expr"
	}

	var result float64
	var err error

	switch operation {
	case "add":
		a, b, e := getTwoNumbers(config)
		if e != nil {
			return node.Result{}, e
		}
		result = a + b

	case "sub":
		a, b, e := getTwoNumbers(config)
		if e != nil {
			return node.Result{}, e
		}
		result = a - b

	case "mul":
		a, b, e := getTwoNumbers(config)
		if e != nil {
			return node.Result{}, e
		}
		result = a * b

	case "div":
		a, b, e := getTwoNumbers(config)
		if e != nil {
			return node.Result{}, e
		}
		if b == 0 {
			return node.Result{}, fmt.Errorf("division by zero")
		}
		result = a / b

	case "mod":
		a, b, e := getTwoNumbers(config)
		if e != nil {
			return node.Result{}, e
		}
		if b == 0 {
			return node.Result{}, fmt.Errorf("modulo by zero")
		}
		result = math.Mod(a, b)

	case "pow":
		a, b, e := getTwoNumbers(config)
		if e != nil {
			return node.Result{}, e
		}
		result = math.Pow(a, b)

	case "sqrt":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		if a < 0 {
			return node.Result{}, fmt.Errorf("sqrt of negative number")
		}
		result = math.Sqrt(a)

	case "abs":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		result = math.Abs(a)

	case "floor":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		result = math.Floor(a)

	case "ceil":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		result = math.Ceil(a)

	case "round":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		precision := 0
		if p, ok := config["precision"].(float64); ok {
			precision = int(p)
		}
		multiplier := math.Pow(10, float64(precision))
		result = math.Round(a*multiplier) / multiplier

	case "min":
		values, e := getNumberArray(config)
		if e != nil {
			return node.Result{}, e
		}
		if len(values) == 0 {
			return node.Result{}, fmt.Errorf("min requires at least one value")
		}
		result = values[0]
		for _, v := range values[1:] {
			if v < result {
				result = v
			}
		}

	case "max":
		values, e := getNumberArray(config)
		if e != nil {
			return node.Result{}, e
		}
		if len(values) == 0 {
			return node.Result{}, fmt.Errorf("max requires at least one value")
		}
		result = values[0]
		for _, v := range values[1:] {
			if v > result {
				result = v
			}
		}

	case "sum":
		values, e := getNumberArray(config)
		if e != nil {
			return node.Result{}, e
		}
		for _, v := range values {
			result += v
		}

	case "avg":
		values, e := getNumberArray(config)
		if e != nil {
			return node.Result{}, e
		}
		if len(values) == 0 {
			return node.Result{}, fmt.Errorf("avg requires at least one value")
		}
		for _, v := range values {
			result += v
		}
		result /= float64(len(values))

	case "sin":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		result = math.Sin(a)

	case "cos":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		result = math.Cos(a)

	case "tan":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		result = math.Tan(a)

	case "log":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		if a <= 0 {
			return node.Result{}, fmt.Errorf("log of non-positive number")
		}
		result = math.Log(a)

	case "log10":
		a, e := getOneNumber(config)
		if e != nil {
			return node.Result{}, e
		}
		if a <= 0 {
			return node.Result{}, fmt.Errorf("log10 of non-positive number")
		}
		result = math.Log10(a)

	case "expr":
		expression, ok := config["expression"].(string)
		if !ok || expression == "" {
			return node.Result{}, fmt.Errorf("expression is required for expr operation")
		}

		// Build environment
		env := map[string]any{
			"pi":    math.Pi,
			"e":     math.E,
			"sqrt":  math.Sqrt,
			"abs":   math.Abs,
			"floor": math.Floor,
			"ceil":  math.Ceil,
			"round": math.Round,
			"sin":   math.Sin,
			"cos":   math.Cos,
			"tan":   math.Tan,
			"log":   math.Log,
			"log10": math.Log10,
			"pow":   math.Pow,
			"min":   math.Min,
			"max":   math.Max,
		}

		if vars, ok := config["vars"].(map[string]any); ok {
			for k, v := range vars {
				env[k] = v
			}
		}

		program, err := expr.Compile(expression, expr.Env(env))
		if err != nil {
			return node.Result{}, fmt.Errorf("failed to compile expression: %w", err)
		}

		output, err := expr.Run(program, env)
		if err != nil {
			return node.Result{}, fmt.Errorf("failed to evaluate expression: %w", err)
		}

		if r, ok := toFloat64(output); ok {
			result = r
		} else if boolVal, ok := output.(bool); ok {
			if boolVal {
				result = 1
			} else {
				result = 0
			}
		} else {
			return node.Result{
				Success: true,
				Data: map[string]any{
					"result": output,
				},
				Status: "evaluated",
				Logs: []node.LogEntry{
					{Level: "info", Message: fmt.Sprintf("Result: %v", output)},
				},
			}, nil
		}

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}

	if err != nil {
		return node.Result{}, err
	}

	// Check for integer result
	var displayResult any = result
	if result == math.Floor(result) && result >= math.MinInt64 && result <= math.MaxInt64 {
		displayResult = int64(result)
	}

	return node.Result{
		Success: true,
		Data: map[string]any{
			"result": result,
			"int":    int64(result),
		},
		Status: fmt.Sprintf("%v", displayResult),
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Result: %v", displayResult)},
		},
	}, nil
}

func getOneNumber(config map[string]any) (float64, error) {
	if a, ok := config["a"].(float64); ok {
		return a, nil
	}
	if a, ok := config["value"].(float64); ok {
		return a, nil
	}
	return 0, fmt.Errorf("operand 'a' or 'value' is required")
}

func getTwoNumbers(config map[string]any) (float64, float64, error) {
	a, ok := config["a"].(float64)
	if !ok {
		return 0, 0, fmt.Errorf("operand 'a' is required")
	}
	b, ok := config["b"].(float64)
	if !ok {
		return 0, 0, fmt.Errorf("operand 'b' is required")
	}
	return a, b, nil
}

func getNumberArray(config map[string]any) ([]float64, error) {
	values, ok := config["values"].([]any)
	if !ok {
		// Try to use a and b
		a, aOk := config["a"].(float64)
		b, bOk := config["b"].(float64)
		if aOk && bOk {
			return []float64{a, b}, nil
		}
		if aOk {
			return []float64{a}, nil
		}
		return nil, fmt.Errorf("values array is required")
	}

	result := make([]float64, 0, len(values))
	for i, v := range values {
		if n, ok := toFloat64(v); ok {
			result = append(result, n)
		} else {
			return nil, fmt.Errorf("values[%d] is not a number", i)
		}
	}
	return result, nil
}
