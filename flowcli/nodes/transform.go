package nodes

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/nameer-kp/flowcli/pkg/node"
)

// TransformNode transforms data using expressions
type TransformNode struct{}

func NewTransformNode() *TransformNode {
	return &TransformNode{}
}

func (n *TransformNode) Name() string {
	return "transform"
}

func (n *TransformNode) Description() string {
	return "Transform data using expressions"
}

func (n *TransformNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"input":      "string", // Variable reference
		"operations": "array",  // List of transformations
		"output":     "string", // Output variable name (optional)
	}
}

func (n *TransformNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	// Get input data
	inputRef, _ := config["input"].(string)
	var data interface{}

	if inputRef != "" {
		var ok bool
		data, ok = ctx.Get(inputRef)
		if !ok {
			return node.Result{}, fmt.Errorf("input '%s' not found in context", inputRef)
		}
	}

	// Apply operations
	operations, _ := config["operations"].([]interface{})
	var err error

	for i, opRaw := range operations {
		op, ok := opRaw.(map[string]interface{})
		if !ok {
			return node.Result{}, fmt.Errorf("operation %d: invalid format", i)
		}

		opType, _ := op["type"].(string)
		expression, _ := op["expression"].(string)

		ctx.Logger().Debug("applying transformation", "type", opType, "expression", expression)

		switch opType {
		case "expr":
			data, err = n.applyExpr(data, expression, ctx)
		case "jsonpath":
			data, err = n.applyJSONPath(data, expression)
		case "template":
			data, err = n.applyTemplate(data, expression, ctx)
		case "map":
			data, err = n.applyMap(data, expression, ctx)
		case "filter":
			data, err = n.applyFilter(data, expression, ctx)
		case "pick":
			data, err = n.applyPick(data, expression)
		default:
			err = fmt.Errorf("unknown operation type: %s", opType)
		}

		if err != nil {
			return node.Result{}, fmt.Errorf("operation %d (%s): %w", i, opType, err)
		}
	}

	// Store output if specified
	if output, ok := config["output"].(string); ok && output != "" {
		ctx.Set(output, data)
	}

	return node.Result{
		Success: true,
		Data:    data,
		Status:  "Transformed",
		Logs:    []node.LogEntry{{Level: "info", Message: "Data transformation complete"}},
	}, nil
}

// applyExpr evaluates an expr-lang expression
func (n *TransformNode) applyExpr(data interface{}, expression string, ctx node.Context) (interface{}, error) {
	env := map[string]interface{}{
		"data": data,
	}

	// Add context data to environment
	if ctxData, ok := ctx.Get("_all"); ok {
		if m, ok := ctxData.(map[string]interface{}); ok {
			for k, v := range m {
				env[k] = v
			}
		}
	}

	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return nil, fmt.Errorf("compile error: %w", err)
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("eval error: %w", err)
	}

	return result, nil
}

// applyJSONPath extracts data using simple JSON path (dot notation)
func (n *TransformNode) applyJSONPath(data interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

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
			return nil, fmt.Errorf("cannot traverse '%s' in non-object", part)
		}
	}

	return current, nil
}

// applyTemplate substitutes {{key}} patterns
func (n *TransformNode) applyTemplate(data interface{}, template string, ctx node.Context) (interface{}, error) {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result := re.ReplaceAllStringFunc(template, func(match string) string {
		key := strings.TrimSpace(match[2 : len(match)-2])

		// Try data first
		if m, ok := data.(map[string]interface{}); ok {
			if v, ok := m[key]; ok {
				return fmt.Sprintf("%v", v)
			}
		}

		// Try context
		if v, ok := ctx.Get(key); ok {
			return fmt.Sprintf("%v", v)
		}

		return match
	})

	return result, nil
}

// applyMap transforms each element of an array
func (n *TransformNode) applyMap(data interface{}, expression string, ctx node.Context) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		// Try to convert from []map[string]interface{}
		if mapArr, ok := data.([]map[string]interface{}); ok {
			arr = make([]interface{}, len(mapArr))
			for i, m := range mapArr {
				arr[i] = m
			}
		} else {
			return nil, fmt.Errorf("map requires array input")
		}
	}

	results := make([]interface{}, len(arr))
	for i, item := range arr {
		env := map[string]interface{}{
			"item":  item,
			"index": i,
			"data":  data,
		}

		program, err := expr.Compile(expression, expr.Env(env))
		if err != nil {
			return nil, fmt.Errorf("compile error: %w", err)
		}

		result, err := expr.Run(program, env)
		if err != nil {
			return nil, fmt.Errorf("eval error at index %d: %w", i, err)
		}

		results[i] = result
	}

	return results, nil
}

// applyFilter keeps elements matching expression
func (n *TransformNode) applyFilter(data interface{}, expression string, ctx node.Context) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("filter requires array input")
	}

	var results []interface{}
	for i, item := range arr {
		env := map[string]interface{}{
			"item":  item,
			"index": i,
		}

		program, err := expr.Compile(expression, expr.Env(env))
		if err != nil {
			return nil, fmt.Errorf("compile error: %w", err)
		}

		result, err := expr.Run(program, env)
		if err != nil {
			return nil, fmt.Errorf("eval error at index %d: %w", i, err)
		}

		if keep, ok := result.(bool); ok && keep {
			results = append(results, item)
		}
	}

	return results, nil
}

// applyPick selects specific fields from objects
func (n *TransformNode) applyPick(data interface{}, fields string) (interface{}, error) {
	fieldList := strings.Split(fields, ",")
	for i := range fieldList {
		fieldList[i] = strings.TrimSpace(fieldList[i])
	}

	pick := func(obj map[string]interface{}) map[string]interface{} {
		result := make(map[string]interface{})
		for _, f := range fieldList {
			if v, ok := obj[f]; ok {
				result[f] = v
			}
		}
		return result
	}

	switch v := data.(type) {
	case map[string]interface{}:
		return pick(v), nil
	case []interface{}:
		results := make([]interface{}, len(v))
		for i, item := range v {
			if obj, ok := item.(map[string]interface{}); ok {
				results[i] = pick(obj)
			} else {
				results[i] = item
			}
		}
		return results, nil
	default:
		return data, nil
	}
}

// Ensure json import is used
var _ = json.Marshal
