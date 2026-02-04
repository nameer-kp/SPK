package nodes

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/nameer-kp/atem/pkg/node"
)

// EnvNode provides environment variable management
type EnvNode struct{}

func NewEnvNode() *EnvNode {
	return &EnvNode{}
}

func (n *EnvNode) Name() string {
	return "env"
}

func (n *EnvNode) Description() string {
	return "Get, set, and manage environment variables"
}

func (n *EnvNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // get, set, delete, list, expand
		"name":      "string", // variable name
		"value":     "string", // value for set
		"default":   "string", // default value for get
		"prefix":    "string", // prefix filter for list
		"text":      "string", // text for expand operation
	}
}

func (n *EnvNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "get"
	}

	switch operation {
	case "get":
		name, _ := config["name"].(string)
		if name == "" {
			return node.Result{}, fmt.Errorf("name is required for get")
		}

		value := os.Getenv(name)
		exists := value != "" || os.Environ() != nil

		// Check if actually set (not just empty)
		for _, e := range os.Environ() {
			if strings.HasPrefix(e, name+"=") {
				exists = true
				break
			}
		}

		if value == "" {
			if defaultVal, ok := config["default"].(string); ok {
				value = defaultVal
			}
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":   name,
				"value":  value,
				"exists": exists,
			},
			Status: fmt.Sprintf("%s=%s", name, truncate(value, 30)),
		}, nil

	case "set":
		name, _ := config["name"].(string)
		if name == "" {
			return node.Result{}, fmt.Errorf("name is required for set")
		}

		value, _ := config["value"].(string)
		if err := os.Setenv(name, value); err != nil {
			return node.Result{}, fmt.Errorf("failed to set env: %w", err)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":  name,
				"value": value,
			},
			Status: fmt.Sprintf("set %s", name),
		}, nil

	case "delete", "unset":
		name, _ := config["name"].(string)
		if name == "" {
			return node.Result{}, fmt.Errorf("name is required for delete")
		}

		if err := os.Unsetenv(name); err != nil {
			return node.Result{}, fmt.Errorf("failed to unset env: %w", err)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":    name,
				"deleted": true,
			},
			Status: fmt.Sprintf("unset %s", name),
		}, nil

	case "list":
		prefix, _ := config["prefix"].(string)

		envVars := make(map[string]string)
		var names []string

		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			if len(pair) == 2 {
				name := pair[0]
				value := pair[1]

				if prefix == "" || strings.HasPrefix(name, prefix) {
					envVars[name] = value
					names = append(names, name)
				}
			}
		}

		sort.Strings(names)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"variables": envVars,
				"names":     names,
				"count":     len(names),
			},
			Status: fmt.Sprintf("%d variables", len(names)),
		}, nil

	case "expand":
		text, _ := config["text"].(string)
		if text == "" {
			return node.Result{}, fmt.Errorf("text is required for expand")
		}

		expanded := os.ExpandEnv(text)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result":   expanded,
				"original": text,
			},
			Status: "expanded",
		}, nil

	case "require":
		name, _ := config["name"].(string)
		if name == "" {
			return node.Result{}, fmt.Errorf("name is required for require")
		}

		value := os.Getenv(name)
		if value == "" {
			return node.Result{
				Success: false,
				Data: map[string]any{
					"name":   name,
					"exists": false,
				},
				Status: "missing",
			}, fmt.Errorf("required environment variable %s is not set", name)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":   name,
				"value":  value,
				"exists": true,
			},
			Status: "present",
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s (valid: get, set, delete, list, expand, require)", operation)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
