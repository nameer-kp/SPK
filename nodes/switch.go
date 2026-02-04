package nodes

import (
	"fmt"

	"github.com/nameer-kp/atem/pkg/node"
)

// SwitchNode provides multi-way branching (switch/case pattern)
type SwitchNode struct{}

func NewSwitchNode() *SwitchNode {
	return &SwitchNode{}
}

func (n *SwitchNode) Name() string {
	return "switch"
}

func (n *SwitchNode) Description() string {
	return "Multi-way branching based on value matching"
}

func (n *SwitchNode) ConfigSchema() map[string]any {
	return map[string]any{
		"value":   "any",   // value to switch on
		"cases":   "array", // array of {match, result} objects
		"default": "any",   // default result if no match
	}
}

func (n *SwitchNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	value := config["value"]

	cases, ok := config["cases"].([]any)
	if !ok {
		return node.Result{}, fmt.Errorf("cases array is required")
	}

	// Find matching case
	for i, c := range cases {
		caseObj, ok := c.(map[string]any)
		if !ok {
			return node.Result{}, fmt.Errorf("case %d is not a valid object", i)
		}

		match := caseObj["match"]
		result := caseObj["result"]

		// Support multiple matches (array)
		matches := []any{match}
		if matchArr, ok := match.([]any); ok {
			matches = matchArr
		}

		for _, m := range matches {
			if valuesEqual(value, m) {
				return node.Result{
					Success: true,
					Data: map[string]any{
						"matched": true,
						"case":    i,
						"match":   m,
						"result":  result,
					},
					Status: fmt.Sprintf("case %d", i),
					Logs: []node.LogEntry{
						{Level: "info", Message: fmt.Sprintf("Matched case %d: %v", i, m)},
					},
				}, nil
			}
		}
	}

	// No match - use default
	defaultVal := config["default"]

	return node.Result{
		Success: true,
		Data: map[string]any{
			"matched": false,
			"case":    -1,
			"result":  defaultVal,
		},
		Status: "default",
		Logs: []node.LogEntry{
			{Level: "info", Message: "No match, using default"},
		},
	}, nil
}

func valuesEqual(a, b any) bool {
	// Handle numeric comparison
	if aNum, aOk := toFloat64(a); aOk {
		if bNum, bOk := toFloat64(b); bOk {
			return aNum == bNum
		}
	}

	// String comparison
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
