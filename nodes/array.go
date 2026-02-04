package nodes

import (
	"fmt"
	"sort"

	"github.com/nameer-kp/atem/pkg/node"
)

// ArrayNode provides array manipulation operations
type ArrayNode struct{}

func NewArrayNode() *ArrayNode {
	return &ArrayNode{}
}

func (n *ArrayNode) Name() string {
	return "array"
}

func (n *ArrayNode) Description() string {
	return "Array manipulation: push, pop, slice, sort, filter, map, reduce"
}

func (n *ArrayNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // push, pop, shift, unshift, slice, concat, reverse, sort, unique, flatten, find, filter, map, reduce, includes, indexOf, join, split
		"input":     "array",  // input array
		"value":     "any",    // value for push/unshift/includes
		"start":     "number", // start index for slice
		"end":       "number", // end index for slice
		"with":      "array",  // array for concat
		"separator": "string", // separator for join/split
		"initial":   "any",    // initial value for reduce
	}
}

func (n *ArrayNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		return node.Result{}, fmt.Errorf("operation is required")
	}

	switch operation {
	case "create":
		// Create array from values
		var arr []any
		if vals, ok := config["values"].([]any); ok {
			arr = vals
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": arr,
				"length": len(arr),
			},
			Status: fmt.Sprintf("created [%d]", len(arr)),
		}, nil

	case "push":
		arr := getArray(config)
		value := config["value"]
		result := append(arr, value)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("pushed [%d]", len(result)),
		}, nil

	case "pop":
		arr := getArray(config)
		if len(arr) == 0 {
			return node.Result{
				Success: true,
				Data: map[string]any{
					"result": arr,
					"popped": nil,
					"length": 0,
				},
				Status: "empty",
			}, nil
		}
		popped := arr[len(arr)-1]
		result := arr[:len(arr)-1]
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"popped": popped,
				"length": len(result),
			},
			Status: fmt.Sprintf("popped [%d]", len(result)),
		}, nil

	case "shift":
		arr := getArray(config)
		if len(arr) == 0 {
			return node.Result{
				Success: true,
				Data: map[string]any{
					"result":  arr,
					"shifted": nil,
					"length":  0,
				},
				Status: "empty",
			}, nil
		}
		shifted := arr[0]
		result := arr[1:]
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result":  result,
				"shifted": shifted,
				"length":  len(result),
			},
			Status: fmt.Sprintf("shifted [%d]", len(result)),
		}, nil

	case "unshift":
		arr := getArray(config)
		value := config["value"]
		result := append([]any{value}, arr...)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("unshifted [%d]", len(result)),
		}, nil

	case "slice":
		arr := getArray(config)
		start := 0
		if s, ok := config["start"].(float64); ok {
			start = int(s)
		}
		end := len(arr)
		if e, ok := config["end"].(float64); ok {
			end = int(e)
		}

		// Handle negative indices
		if start < 0 {
			start = len(arr) + start
		}
		if end < 0 {
			end = len(arr) + end
		}

		// Bounds check
		if start < 0 {
			start = 0
		}
		if end > len(arr) {
			end = len(arr)
		}
		if start > end {
			start = end
		}

		result := arr[start:end]
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("slice [%d]", len(result)),
		}, nil

	case "concat":
		arr := getArray(config)
		with, _ := config["with"].([]any)
		result := append(arr, with...)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("concat [%d]", len(result)),
		}, nil

	case "reverse":
		arr := getArray(config)
		result := make([]any, len(arr))
		for i, v := range arr {
			result[len(arr)-1-i] = v
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("reversed [%d]", len(result)),
		}, nil

	case "sort":
		arr := getArray(config)
		result := make([]any, len(arr))
		copy(result, arr)

		// Sort by string representation (simple sort)
		sort.Slice(result, func(i, j int) bool {
			// Try numeric comparison first
			iNum, iOk := toFloat64(result[i])
			jNum, jOk := toFloat64(result[j])
			if iOk && jOk {
				return iNum < jNum
			}
			return fmt.Sprintf("%v", result[i]) < fmt.Sprintf("%v", result[j])
		})

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("sorted [%d]", len(result)),
		}, nil

	case "unique":
		arr := getArray(config)
		seen := make(map[string]bool)
		var result []any
		for _, v := range arr {
			key := fmt.Sprintf("%v", v)
			if !seen[key] {
				seen[key] = true
				result = append(result, v)
			}
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("unique [%d]", len(result)),
		}, nil

	case "flatten":
		arr := getArray(config)
		result := flattenArray(arr, 1)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("flattened [%d]", len(result)),
		}, nil

	case "includes":
		arr := getArray(config)
		value := config["value"]
		found := false
		for _, v := range arr {
			if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", value) {
				found = true
				break
			}
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": found,
			},
			Status: fmt.Sprintf("includes: %v", found),
		}, nil

	case "indexOf":
		arr := getArray(config)
		value := config["value"]
		index := -1
		for i, v := range arr {
			if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", value) {
				index = i
				break
			}
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": index,
				"found":  index >= 0,
			},
			Status: fmt.Sprintf("indexOf: %d", index),
		}, nil

	case "join":
		arr := getArray(config)
		separator, _ := config["separator"].(string)
		if separator == "" {
			separator = ","
		}
		var result string
		for i, v := range arr {
			if i > 0 {
				result += separator
			}
			result += fmt.Sprintf("%v", v)
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
			},
			Status: "joined",
		}, nil

	case "split":
		input, _ := config["input"].(string)
		separator, _ := config["separator"].(string)
		if separator == "" {
			separator = ","
		}
		var result []any
		if input != "" {
			parts := splitString(input, separator)
			for _, p := range parts {
				result = append(result, p)
			}
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("split [%d]", len(result)),
		}, nil

	case "length":
		arr := getArray(config)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": len(arr),
			},
			Status: fmt.Sprintf("length: %d", len(arr)),
		}, nil

	case "first":
		arr := getArray(config)
		if len(arr) == 0 {
			return node.Result{
				Success: true,
				Data: map[string]any{
					"result": nil,
					"found":  false,
				},
				Status: "empty",
			}, nil
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": arr[0],
				"found":  true,
			},
			Status: "first",
		}, nil

	case "last":
		arr := getArray(config)
		if len(arr) == 0 {
			return node.Result{
				Success: true,
				Data: map[string]any{
					"result": nil,
					"found":  false,
				},
				Status: "empty",
			}, nil
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": arr[len(arr)-1],
				"found":  true,
			},
			Status: "last",
		}, nil

	case "at":
		arr := getArray(config)
		index := 0
		if i, ok := config["index"].(float64); ok {
			index = int(i)
		}

		// Handle negative indices
		if index < 0 {
			index = len(arr) + index
		}

		if index < 0 || index >= len(arr) {
			return node.Result{
				Success: true,
				Data: map[string]any{
					"result": nil,
					"found":  false,
				},
				Status: "out of bounds",
			}, nil
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": arr[index],
				"found":  true,
			},
			Status: fmt.Sprintf("at[%d]", index),
		}, nil

	case "fill":
		arr := getArray(config)
		value := config["value"]
		result := make([]any, len(arr))
		for i := range result {
			result[i] = value
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"length": len(result),
			},
			Status: fmt.Sprintf("filled [%d]", len(result)),
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

func getArray(config map[string]any) []any {
	if arr, ok := config["input"].([]any); ok {
		return arr
	}
	return []any{}
}

func flattenArray(arr []any, depth int) []any {
	var result []any
	for _, v := range arr {
		if nested, ok := v.([]any); ok && depth > 0 {
			result = append(result, flattenArray(nested, depth-1)...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	if sep == "" {
		// Split into characters
		result := make([]string, len(s))
		for i, c := range s {
			result[i] = string(c)
		}
		return result
	}

	var result []string
	for len(s) > 0 {
		idx := -1
		for i := 0; i <= len(s)-len(sep); i++ {
			if s[i:i+len(sep)] == sep {
				idx = i
				break
			}
		}
		if idx < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	return result
}
