package nodes

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nameer-kp/atem/pkg/node"
)

// InputNode provides interactive user input during workflow execution
type InputNode struct{}

func NewInputNode() *InputNode {
	return &InputNode{}
}

func (n *InputNode) Name() string {
	return "input"
}

func (n *InputNode) Description() string {
	return "Get interactive user input during workflow execution"
}

func (n *InputNode) ConfigSchema() map[string]any {
	return map[string]any{
		"prompt":   "string", // prompt to display
		"type":     "string", // string, number, bool, choice
		"default":  "any",    // default value if empty
		"required": "bool",   // whether input is required
		"choices":  "array",  // for type=choice: list of options
		"min":      "number", // for type=number: minimum value
		"max":      "number", // for type=number: maximum value
		"pattern":  "string", // regex pattern for validation
		"mask":     "bool",   // mask input (for passwords)
	}
}

func (n *InputNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	prompt, _ := config["prompt"].(string)
	if prompt == "" {
		prompt = "Enter value"
	}

	inputType, _ := config["type"].(string)
	if inputType == "" {
		inputType = "string"
	}

	required, _ := config["required"].(bool)
	defaultVal := config["default"]

	// Build prompt string
	promptStr := prompt
	if defaultVal != nil {
		promptStr = fmt.Sprintf("%s [%v]", prompt, defaultVal)
	}

	// For choice type, show options
	var choices []string
	if inputType == "choice" {
		if choiceList, ok := config["choices"].([]any); ok {
			for _, c := range choiceList {
				choices = append(choices, fmt.Sprintf("%v", c))
			}
		}
		if len(choices) > 0 {
			promptStr += "\n"
			for i, c := range choices {
				promptStr += fmt.Sprintf("  %d) %s\n", i+1, c)
			}
			promptStr += "Select (1-" + strconv.Itoa(len(choices)) + ")"
		}
	}

	// Read input
	fmt.Print(promptStr + ": ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return node.Result{}, fmt.Errorf("failed to read input: %w", err)
	}
	input = strings.TrimSpace(input)

	// Apply default if empty
	if input == "" && defaultVal != nil {
		input = fmt.Sprintf("%v", defaultVal)
	}

	// Check required
	if required && input == "" {
		return node.Result{}, fmt.Errorf("input is required")
	}

	// Process based on type
	var value any = input
	switch inputType {
	case "number":
		num, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return node.Result{}, fmt.Errorf("invalid number: %v", input)
		}
		if minVal, ok := config["min"].(float64); ok && num < minVal {
			return node.Result{}, fmt.Errorf("value must be >= %v", minVal)
		}
		if maxVal, ok := config["max"].(float64); ok && num > maxVal {
			return node.Result{}, fmt.Errorf("value must be <= %v", maxVal)
		}
		value = num

	case "bool":
		lower := strings.ToLower(input)
		switch lower {
		case "true", "yes", "y", "1":
			value = true
		case "false", "no", "n", "0":
			value = false
		default:
			return node.Result{}, fmt.Errorf("invalid boolean: %v", input)
		}

	case "choice":
		if len(choices) == 0 {
			return node.Result{}, fmt.Errorf("no choices provided")
		}

		// Try to parse as number
		if idx, err := strconv.Atoi(input); err == nil {
			if idx >= 1 && idx <= len(choices) {
				value = choices[idx-1]
			} else {
				return node.Result{}, fmt.Errorf("invalid choice: %v (must be 1-%d)", input, len(choices))
			}
		} else {
			// Check if input matches a choice
			found := false
			for _, c := range choices {
				if strings.EqualFold(input, c) {
					value = c
					found = true
					break
				}
			}
			if !found {
				return node.Result{}, fmt.Errorf("invalid choice: %v", input)
			}
		}

	case "string":
		// Already a string, just apply pattern if specified
		if pattern, ok := config["pattern"].(string); ok && pattern != "" {
			// Simple pattern matching (would need regexp for full support)
			ctx.Logger().Debug("pattern validation not implemented", "pattern", pattern)
		}
	}

	return node.Result{
		Success: true,
		Data: map[string]any{
			"value": value,
			"raw":   input,
			"type":  inputType,
		},
		Status: fmt.Sprintf("input: %v", value),
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("User input: %v", value)},
		},
	}, nil
}
