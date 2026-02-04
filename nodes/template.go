package nodes

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/nameer-kp/atem/pkg/node"
)

// TemplateNode provides string templating capabilities
type TemplateNode struct{}

func NewTemplateNode() *TemplateNode {
	return &TemplateNode{}
}

func (n *TemplateNode) Name() string {
	return "template"
}

func (n *TemplateNode) Description() string {
	return "Render string templates with variable substitution"
}

func (n *TemplateNode) ConfigSchema() map[string]any {
	return map[string]any{
		"template": "string", // template string (Go template syntax)
		"vars":     "object", // variables for substitution
		"format":   "string", // format type: go, simple, mustache
		"strict":   "bool",   // fail on missing variables
	}
}

func (n *TemplateNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	tmplStr, ok := config["template"].(string)
	if !ok || tmplStr == "" {
		return node.Result{}, fmt.Errorf("template string is required")
	}

	vars := make(map[string]any)
	if v, ok := config["vars"].(map[string]any); ok {
		vars = v
	}

	format, _ := config["format"].(string)
	if format == "" {
		format = "simple"
	}

	strict, _ := config["strict"].(bool)

	var result string
	var err error

	switch format {
	case "go":
		result, err = renderGoTemplate(tmplStr, vars, strict)
	case "simple":
		result, err = renderSimpleTemplate(tmplStr, vars, strict)
	case "mustache":
		result, err = renderMustacheTemplate(tmplStr, vars, strict)
	default:
		return node.Result{}, fmt.Errorf("unknown format: %s (valid: go, simple, mustache)", format)
	}

	if err != nil {
		return node.Result{}, err
	}

	return node.Result{
		Success: true,
		Data: map[string]any{
			"result": result,
			"length": len(result),
		},
		Status: fmt.Sprintf("rendered %d chars", len(result)),
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Template rendered: %d chars", len(result))},
		},
	}, nil
}

// renderGoTemplate uses Go's text/template
func renderGoTemplate(tmplStr string, vars map[string]any, strict bool) (string, error) {
	tmpl, err := template.New("template").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	if strict {
		tmpl = tmpl.Option("missingkey=error")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// renderSimpleTemplate uses {{var}} syntax (like Atem's native syntax)
func renderSimpleTemplate(tmplStr string, vars map[string]any, strict bool) (string, error) {
	result := tmplStr
	missingVars := []string{}

	// Pattern: {{varname}} or {{varname.field}}
	pattern := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result = pattern.ReplaceAllStringFunc(result, func(match string) string {
		// Extract variable path
		path := strings.TrimSpace(match[2 : len(match)-2])
		value, found := resolveVarPath(vars, path)
		if !found {
			if strict {
				missingVars = append(missingVars, path)
			}
			return match // Keep original if not found
		}
		return fmt.Sprintf("%v", value)
	})

	if strict && len(missingVars) > 0 {
		return "", fmt.Errorf("missing variables: %v", missingVars)
	}

	return result, nil
}

// renderMustacheTemplate uses {{var}} and {{#section}}...{{/section}} syntax
func renderMustacheTemplate(tmplStr string, vars map[string]any, strict bool) (string, error) {
	result := tmplStr
	missingVars := []string{}

	// Handle sections {{#name}}...{{/name}}
	sectionPattern := regexp.MustCompile(`\{\{#(\w+)\}\}([\s\S]*?)\{\{/\1\}\}`)
	result = sectionPattern.ReplaceAllStringFunc(result, func(match string) string {
		// Extract section name
		matches := sectionPattern.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}
		name := matches[1]
		content := matches[2]

		value, found := vars[name]
		if !found {
			if strict {
				missingVars = append(missingVars, name)
			}
			return ""
		}

		// Handle different value types
		switch v := value.(type) {
		case bool:
			if v {
				return content
			}
			return ""
		case []any:
			var buf strings.Builder
			for _, item := range v {
				itemVars := make(map[string]any)
				for k, val := range vars {
					itemVars[k] = val
				}
				if itemMap, ok := item.(map[string]any); ok {
					for k, val := range itemMap {
						itemVars[k] = val
					}
				} else {
					itemVars["."] = item
				}
				rendered, _ := renderMustacheTemplate(content, itemVars, false)
				buf.WriteString(rendered)
			}
			return buf.String()
		default:
			if value != nil {
				return content
			}
			return ""
		}
	})

	// Handle inverted sections {{^name}}...{{/name}}
	invertedPattern := regexp.MustCompile(`\{\{\^(\w+)\}\}([\s\S]*?)\{\{/\1\}\}`)
	result = invertedPattern.ReplaceAllStringFunc(result, func(match string) string {
		matches := invertedPattern.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}
		name := matches[1]
		content := matches[2]

		value, found := vars[name]
		if !found || isFalsy(value) {
			return content
		}
		return ""
	})

	// Handle simple variables
	varPattern := regexp.MustCompile(`\{\{([^#^/}][^}]*)\}\}`)
	result = varPattern.ReplaceAllStringFunc(result, func(match string) string {
		name := strings.TrimSpace(match[2 : len(match)-2])
		value, found := resolveVarPath(vars, name)
		if !found {
			if strict {
				missingVars = append(missingVars, name)
			}
			return ""
		}
		return fmt.Sprintf("%v", value)
	})

	if strict && len(missingVars) > 0 {
		return "", fmt.Errorf("missing variables: %v", missingVars)
	}

	return result, nil
}

func resolveVarPath(vars map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	current := any(vars)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}

func isFalsy(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case bool:
		return !val
	case string:
		return val == ""
	case float64:
		return val == 0
	case int:
		return val == 0
	case []any:
		return len(val) == 0
	}
	return false
}
