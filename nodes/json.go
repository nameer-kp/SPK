package nodes

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/nameer-kp/atem/pkg/node"
)

// JSONNode provides JSON manipulation capabilities
type JSONNode struct{}

func NewJSONNode() *JSONNode {
	return &JSONNode{}
}

func (n *JSONNode) Name() string {
	return "json"
}

func (n *JSONNode) Description() string {
	return "Parse, query, and manipulate JSON data"
}

func (n *JSONNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // parse, stringify, get, set, delete, merge, keys, values, query
		"input":     "any",    // input data
		"path":      "string", // JSON path for get/set/delete
		"value":     "any",    // value for set
		"indent":    "number", // indentation for stringify
		"merge":     "any",    // object to merge
	}
}

func (n *JSONNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "parse"
	}

	switch operation {
	case "parse":
		input, ok := config["input"].(string)
		if !ok {
			return node.Result{}, fmt.Errorf("parse requires string input")
		}

		var result any
		if err := json.Unmarshal([]byte(input), &result); err != nil {
			return node.Result{}, fmt.Errorf("failed to parse JSON: %w", err)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
			},
			Status: "parsed",
		}, nil

	case "stringify":
		input := config["input"]
		indent := 0
		if i, ok := config["indent"].(float64); ok {
			indent = int(i)
		}

		var result []byte
		var err error
		if indent > 0 {
			result, err = json.MarshalIndent(input, "", strings.Repeat(" ", indent))
		} else {
			result, err = json.Marshal(input)
		}
		if err != nil {
			return node.Result{}, fmt.Errorf("failed to stringify: %w", err)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": string(result),
				"length": len(result),
			},
			Status: fmt.Sprintf("%d bytes", len(result)),
		}, nil

	case "get":
		input := config["input"]
		path, _ := config["path"].(string)
		if path == "" {
			return node.Result{}, fmt.Errorf("path is required for get")
		}

		value, found := getJSONPath(input, path)
		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": value,
				"found":  found,
			},
			Status: fmt.Sprintf("get %s", path),
		}, nil

	case "set":
		input := config["input"]
		path, _ := config["path"].(string)
		value := config["value"]
		if path == "" {
			return node.Result{}, fmt.Errorf("path is required for set")
		}

		result, err := setJSONPath(input, path, value)
		if err != nil {
			return node.Result{}, err
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
			},
			Status: fmt.Sprintf("set %s", path),
		}, nil

	case "delete":
		input := config["input"]
		path, _ := config["path"].(string)
		if path == "" {
			return node.Result{}, fmt.Errorf("path is required for delete")
		}

		result, err := deleteJSONPath(input, path)
		if err != nil {
			return node.Result{}, err
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
			},
			Status: fmt.Sprintf("delete %s", path),
		}, nil

	case "merge":
		input := config["input"]
		mergeWith := config["merge"]

		inputMap, ok := input.(map[string]any)
		if !ok {
			return node.Result{}, fmt.Errorf("input must be an object for merge")
		}
		mergeMap, ok := mergeWith.(map[string]any)
		if !ok {
			return node.Result{}, fmt.Errorf("merge value must be an object")
		}

		result := make(map[string]any)
		for k, v := range inputMap {
			result[k] = v
		}
		for k, v := range mergeMap {
			result[k] = v
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": result,
			},
			Status: "merged",
		}, nil

	case "keys":
		input := config["input"]
		inputMap, ok := input.(map[string]any)
		if !ok {
			return node.Result{}, fmt.Errorf("input must be an object for keys")
		}

		keys := make([]string, 0, len(inputMap))
		for k := range inputMap {
			keys = append(keys, k)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": keys,
				"count":  len(keys),
			},
			Status: fmt.Sprintf("%d keys", len(keys)),
		}, nil

	case "values":
		input := config["input"]
		inputMap, ok := input.(map[string]any)
		if !ok {
			return node.Result{}, fmt.Errorf("input must be an object for values")
		}

		values := make([]any, 0, len(inputMap))
		for _, v := range inputMap {
			values = append(values, v)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": values,
				"count":  len(values),
			},
			Status: fmt.Sprintf("%d values", len(values)),
		}, nil

	case "length":
		input := config["input"]
		var length int
		switch v := input.(type) {
		case []any:
			length = len(v)
		case map[string]any:
			length = len(v)
		case string:
			length = len(v)
		default:
			return node.Result{}, fmt.Errorf("length requires array, object, or string")
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": length,
			},
			Status: fmt.Sprintf("length: %d", length),
		}, nil

	case "type":
		input := config["input"]
		var typeName string
		switch input.(type) {
		case nil:
			typeName = "null"
		case bool:
			typeName = "boolean"
		case float64:
			typeName = "number"
		case string:
			typeName = "string"
		case []any:
			typeName = "array"
		case map[string]any:
			typeName = "object"
		default:
			typeName = fmt.Sprintf("%T", input)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result": typeName,
			},
			Status: typeName,
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

// getJSONPath retrieves a value at a dot-separated path
func getJSONPath(data any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, false
			}
			current = v[idx]
		default:
			return nil, false
		}
	}

	return current, true
}

// setJSONPath sets a value at a dot-separated path
func setJSONPath(data any, path string, value any) (any, error) {
	parts := strings.Split(path, ".")

	if len(parts) == 1 {
		if m, ok := data.(map[string]any); ok {
			m[parts[0]] = value
			return m, nil
		}
		return nil, fmt.Errorf("cannot set property on non-object")
	}

	switch v := data.(type) {
	case map[string]any:
		child, exists := v[parts[0]]
		if !exists {
			child = make(map[string]any)
		}
		newChild, err := setJSONPath(child, strings.Join(parts[1:], "."), value)
		if err != nil {
			return nil, err
		}
		v[parts[0]] = newChild
		return v, nil
	case []any:
		idx, err := strconv.Atoi(parts[0])
		if err != nil || idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("invalid array index: %s", parts[0])
		}
		newChild, err := setJSONPath(v[idx], strings.Join(parts[1:], "."), value)
		if err != nil {
			return nil, err
		}
		v[idx] = newChild
		return v, nil
	default:
		return nil, fmt.Errorf("cannot traverse %T", data)
	}
}

// deleteJSONPath deletes a value at a dot-separated path
func deleteJSONPath(data any, path string) (any, error) {
	parts := strings.Split(path, ".")

	if len(parts) == 1 {
		if m, ok := data.(map[string]any); ok {
			delete(m, parts[0])
			return m, nil
		}
		if arr, ok := data.([]any); ok {
			idx, err := strconv.Atoi(parts[0])
			if err != nil || idx < 0 || idx >= len(arr) {
				return nil, fmt.Errorf("invalid array index: %s", parts[0])
			}
			return append(arr[:idx], arr[idx+1:]...), nil
		}
		return nil, fmt.Errorf("cannot delete from non-collection")
	}

	switch v := data.(type) {
	case map[string]any:
		child, exists := v[parts[0]]
		if !exists {
			return v, nil
		}
		newChild, err := deleteJSONPath(child, strings.Join(parts[1:], "."))
		if err != nil {
			return nil, err
		}
		v[parts[0]] = newChild
		return v, nil
	case []any:
		idx, err := strconv.Atoi(parts[0])
		if err != nil || idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("invalid array index: %s", parts[0])
		}
		newChild, err := deleteJSONPath(v[idx], strings.Join(parts[1:], "."))
		if err != nil {
			return nil, err
		}
		v[idx] = newChild
		return v, nil
	default:
		return nil, fmt.Errorf("cannot traverse %T", data)
	}
}
