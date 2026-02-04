package nodes

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/nameer-kp/atem/pkg/node"
)

// StringNode provides string manipulation operations
type StringNode struct{}

func NewStringNode() *StringNode {
	return &StringNode{}
}

func (n *StringNode) Name() string {
	return "string"
}

func (n *StringNode) Description() string {
	return "String manipulation: trim, split, replace, format, etc."
}

func (n *StringNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // trim, upper, lower, split, join, replace, substring, pad, repeat, contains, startsWith, endsWith, length, reverse, format
		"input":     "string", // input string
		"value":     "string", // value for various operations
		"find":      "string", // string to find
		"replace":   "string", // replacement string
		"start":     "number", // start index
		"end":       "number", // end index
		"length":    "number", // length for padding/substring
		"char":      "string", // padding character
		"separator": "string", // separator for split/join
		"count":     "number", // count for repeat/replace
		"args":      "array",  // arguments for format
	}
}

func (n *StringNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		return node.Result{}, fmt.Errorf("operation is required")
	}

	input, _ := config["input"].(string)

	switch operation {
	case "trim":
		result := strings.TrimSpace(input)
		return stringResult(result, "trimmed")

	case "trimLeft":
		result := strings.TrimLeftFunc(input, unicode.IsSpace)
		return stringResult(result, "trimmed left")

	case "trimRight":
		result := strings.TrimRightFunc(input, unicode.IsSpace)
		return stringResult(result, "trimmed right")

	case "upper":
		result := strings.ToUpper(input)
		return stringResult(result, "uppercase")

	case "lower":
		result := strings.ToLower(input)
		return stringResult(result, "lowercase")

	case "capitalize":
		if len(input) == 0 {
			return stringResult("", "capitalize")
		}
		result := strings.ToUpper(string(input[0])) + input[1:]
		return stringResult(result, "capitalized")

	case "title":
		result := strings.Title(input)
		return stringResult(result, "title case")

	case "split":
		separator, _ := config["separator"].(string)
		if separator == "" {
			separator = " "
		}
		parts := strings.Split(input, separator)
		var result []any
		for _, p := range parts {
			result = append(result, p)
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"count":  len(result),
			},
			Status: fmt.Sprintf("split [%d]", len(result)),
		}, nil

	case "join":
		arr, _ := config["input"].([]any)
		separator, _ := config["separator"].(string)
		var parts []string
		for _, v := range arr {
			parts = append(parts, fmt.Sprintf("%v", v))
		}
		result := strings.Join(parts, separator)
		return stringResult(result, "joined")

	case "replace":
		find, _ := config["find"].(string)
		replace, _ := config["replace"].(string)
		count := -1
		if c, ok := config["count"].(float64); ok {
			count = int(c)
		}
		result := strings.Replace(input, find, replace, count)
		return stringResult(result, "replaced")

	case "replaceAll":
		find, _ := config["find"].(string)
		replace, _ := config["replace"].(string)
		result := strings.ReplaceAll(input, find, replace)
		return stringResult(result, "replaced all")

	case "substring":
		start := 0
		if s, ok := config["start"].(float64); ok {
			start = int(s)
		}
		end := len(input)
		if e, ok := config["end"].(float64); ok {
			end = int(e)
		}

		// Handle negative indices
		if start < 0 {
			start = len(input) + start
		}
		if end < 0 {
			end = len(input) + end
		}

		// Bounds check
		if start < 0 {
			start = 0
		}
		if end > len(input) {
			end = len(input)
		}
		if start > end {
			start = end
		}

		result := input[start:end]
		return stringResult(result, "substring")

	case "padLeft":
		length := 0
		if l, ok := config["length"].(float64); ok {
			length = int(l)
		}
		char := " "
		if c, ok := config["char"].(string); ok && c != "" {
			char = string(c[0])
		}
		result := input
		for len(result) < length {
			result = char + result
		}
		return stringResult(result, "padded left")

	case "padRight":
		length := 0
		if l, ok := config["length"].(float64); ok {
			length = int(l)
		}
		char := " "
		if c, ok := config["char"].(string); ok && c != "" {
			char = string(c[0])
		}
		result := input
		for len(result) < length {
			result = result + char
		}
		return stringResult(result, "padded right")

	case "repeat":
		count := 1
		if c, ok := config["count"].(float64); ok {
			count = int(c)
		}
		if count > 10000 {
			return node.Result{}, fmt.Errorf("repeat count exceeds limit (10000)")
		}
		result := strings.Repeat(input, count)
		return stringResult(result, fmt.Sprintf("repeated %d times", count))

	case "contains":
		value, _ := config["value"].(string)
		found := strings.Contains(input, value)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": found,
			},
			Status: fmt.Sprintf("contains: %v", found),
		}, nil

	case "startsWith":
		value, _ := config["value"].(string)
		found := strings.HasPrefix(input, value)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": found,
			},
			Status: fmt.Sprintf("startsWith: %v", found),
		}, nil

	case "endsWith":
		value, _ := config["value"].(string)
		found := strings.HasSuffix(input, value)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": found,
			},
			Status: fmt.Sprintf("endsWith: %v", found),
		}, nil

	case "indexOf":
		value, _ := config["value"].(string)
		index := strings.Index(input, value)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": index,
				"found":  index >= 0,
			},
			Status: fmt.Sprintf("indexOf: %d", index),
		}, nil

	case "lastIndexOf":
		value, _ := config["value"].(string)
		index := strings.LastIndex(input, value)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": index,
				"found":  index >= 0,
			},
			Status: fmt.Sprintf("lastIndexOf: %d", index),
		}, nil

	case "length":
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": len(input),
			},
			Status: fmt.Sprintf("length: %d", len(input)),
		}, nil

	case "reverse":
		runes := []rune(input)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result := string(runes)
		return stringResult(result, "reversed")

	case "format":
		args, _ := config["args"].([]any)
		// Simple format using %v placeholders
		result := input
		for i, arg := range args {
			placeholder := fmt.Sprintf("{%d}", i)
			result = strings.Replace(result, placeholder, fmt.Sprintf("%v", arg), 1)
		}
		return stringResult(result, "formatted")

	case "count":
		value, _ := config["value"].(string)
		count := strings.Count(input, value)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": count,
			},
			Status: fmt.Sprintf("count: %d", count),
		}, nil

	case "isEmpty":
		isEmpty := len(strings.TrimSpace(input)) == 0
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": isEmpty,
			},
			Status: fmt.Sprintf("isEmpty: %v", isEmpty),
		}, nil

	case "lines":
		lines := strings.Split(input, "\n")
		var result []any
		for _, line := range lines {
			result = append(result, line)
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"count":  len(result),
			},
			Status: fmt.Sprintf("%d lines", len(result)),
		}, nil

	case "words":
		words := strings.Fields(input)
		var result []any
		for _, word := range words {
			result = append(result, word)
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"count":  len(result),
			},
			Status: fmt.Sprintf("%d words", len(result)),
		}, nil

	case "chars":
		var result []any
		for _, c := range input {
			result = append(result, string(c))
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
				"count":  len(result),
			},
			Status: fmt.Sprintf("%d chars", len(result)),
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

func stringResult(result, status string) (node.Result, error) {
	return node.Result{
		Success: true,
		Data: map[string]any{
			"result": result,
			"length": len(result),
		},
		Status: status,
	}, nil
}
