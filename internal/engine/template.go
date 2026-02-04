package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/nameer-kp/atem/internal/config"
)

// TemplateEngine handles {{variable}} substitution
type TemplateEngine struct {
	config  *config.ResolvedConfig
	context map[string]interface{}
	inputs  map[string]interface{}
}

// NewTemplateEngine creates a template engine with config context
func NewTemplateEngine(cfg *config.ResolvedConfig) *TemplateEngine {
	return &TemplateEngine{
		config:  cfg,
		context: make(map[string]interface{}),
		inputs:  make(map[string]interface{}),
	}
}

// SetContext updates the step output context
func (t *TemplateEngine) SetContext(key string, value interface{}) {
	t.context[key] = value
}

// SetInputs sets user-provided input values
func (t *TemplateEngine) SetInputs(inputs map[string]interface{}) {
	t.inputs = inputs
}

// pattern matches {{path.to.value}}
var templatePattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// Resolve replaces all {{}} templates in a string
func (t *TemplateEngine) Resolve(text string) (string, error) {
	result := templatePattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract path from {{path}}
		path := strings.TrimSpace(match[2 : len(match)-2])
		value, err := t.resolvePath(path)
		if err != nil {
			return match // Leave unresolved on error
		}
		return fmt.Sprintf("%v", value)
	})
	return result, nil
}

// ResolveMap resolves templates in a map recursively
func (t *TemplateEngine) ResolveMap(m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range m {
		resolved, err := t.resolveValue(v)
		if err != nil {
			return nil, err
		}
		result[k] = resolved
	}
	return result, nil
}

// resolveValue resolves templates in any value type
func (t *TemplateEngine) resolveValue(v interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		return t.Resolve(val)
	case map[string]interface{}:
		return t.ResolveMap(val)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			resolved, err := t.resolveValue(item)
			if err != nil {
				return nil, err
			}
			result[i] = resolved
		}
		return result, nil
	default:
		return v, nil
	}
}

// resolvePath looks up a dotted path like "profile.base_url" or "step_id.field"
func (t *TemplateEngine) resolvePath(path string) (interface{}, error) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	root := parts[0]
	var rest string
	if len(parts) > 1 {
		rest = parts[1]
	}

	switch root {
	case "inputs":
		if rest == "" {
			return t.inputs, nil
		}
		return t.getNestedValue(t.inputs, rest)

	case "profile":
		if rest == "" {
			return t.config.Profile.Variables, nil
		}
		if val, ok := t.config.Profile.Variables[rest]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("profile variable '%s' not found", rest)

	case "env":
		if rest == "" {
			return nil, fmt.Errorf("env requires variable name")
		}
		return os.Getenv(rest), nil

	default:
		// Assume it's a step ID reference
		if stepData, ok := t.context[root]; ok {
			if rest == "" {
				return stepData, nil
			}
			if m, ok := stepData.(map[string]interface{}); ok {
				return t.getNestedValue(m, rest)
			}
			return stepData, nil
		}

		// Check if it's a direct variable reference
		if val, ok := t.config.Variables[root]; ok {
			return val, nil
		}

		return nil, fmt.Errorf("unknown reference '%s'", root)
	}
}

// getNestedValue traverses a map by dotted path
func (t *TemplateEngine) getNestedValue(m map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := interface{}(m)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("key '%s' not found", part)
			}
		default:
			return nil, fmt.Errorf("cannot traverse into non-map at '%s'", part)
		}
	}

	return current, nil
}

// ResolveJSON resolves templates and returns JSON bytes
func (t *TemplateEngine) ResolveJSON(v interface{}) ([]byte, error) {
	resolved, err := t.resolveValue(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(resolved)
}

// EvaluateBool evaluates a template expression and returns a boolean
// Returns true if: result is bool true, non-empty string, non-zero number, non-nil
func (t *TemplateEngine) EvaluateBool(expr string) (bool, error) {
	// Strip {{ }} if present and resolve the path
	path := strings.TrimSpace(expr)
	if strings.HasPrefix(path, "{{") && strings.HasSuffix(path, "}}") {
		path = strings.TrimSpace(path[2 : len(path)-2])
	}

	resolved, err := t.resolvePath(path)
	if err != nil {
		return false, err
	}

	// Handle different types
	switch v := resolved.(type) {
	case bool:
		return v, nil
	case string:
		return v != "" && v != "false" && v != "0", nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case nil:
		return false, nil
	default:
		// Non-nil values are truthy
		return true, nil
	}
}
