package nodes

import (
	"fmt"

	"github.com/nameer-kp/atem/pkg/node"
)

// LoopNode provides iteration capabilities
// Supports: for (count-based), while (condition-based), foreach (over arrays)
type LoopNode struct{}

func NewLoopNode() *LoopNode {
	return &LoopNode{}
}

func (n *LoopNode) Name() string {
	return "loop"
}

func (n *LoopNode) Description() string {
	return "Iterate over ranges, arrays, or conditions"
}

func (n *LoopNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":      "string", // for, foreach, while, times
		"times":     "number", // for type=times: repeat N times
		"from":      "number", // for type=for: start value
		"to":        "number", // for type=for: end value (exclusive)
		"step":      "number", // for type=for: step increment (default 1)
		"items":     "array",  // for type=foreach: array to iterate
		"as":        "string", // variable name for current item (default: "item")
		"index_as":  "string", // variable name for index (default: "index")
		"max_iters": "number", // safety limit (default 10000)
	}
}

func (n *LoopNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	loopType, _ := config["type"].(string)
	if loopType == "" {
		loopType = "times"
	}

	maxIters := 10000
	if m, ok := config["max_iters"].(float64); ok {
		maxIters = int(m)
	}

	asVar := "item"
	if as, ok := config["as"].(string); ok && as != "" {
		asVar = as
	}

	indexVar := "index"
	if idx, ok := config["index_as"].(string); ok && idx != "" {
		indexVar = idx
	}

	var iterations []map[string]interface{}

	switch loopType {
	case "times":
		times := 1
		if t, ok := config["times"].(float64); ok {
			times = int(t)
		}
		if times > maxIters {
			return node.Result{}, fmt.Errorf("times %d exceeds max_iters %d", times, maxIters)
		}
		for i := 0; i < times; i++ {
			iterations = append(iterations, map[string]interface{}{
				indexVar: i,
				asVar:    i,
			})
		}

	case "for":
		from := 0
		if f, ok := config["from"].(float64); ok {
			from = int(f)
		}
		to := 10
		if t, ok := config["to"].(float64); ok {
			to = int(t)
		}
		step := 1
		if s, ok := config["step"].(float64); ok {
			step = int(s)
		}
		if step == 0 {
			return node.Result{}, fmt.Errorf("step cannot be 0")
		}

		idx := 0
		if step > 0 {
			for i := from; i < to && idx < maxIters; i += step {
				iterations = append(iterations, map[string]interface{}{
					indexVar: idx,
					asVar:    i,
				})
				idx++
			}
		} else {
			for i := from; i > to && idx < maxIters; i += step {
				iterations = append(iterations, map[string]interface{}{
					indexVar: idx,
					asVar:    i,
				})
				idx++
			}
		}

	case "foreach":
		items, ok := config["items"].([]interface{})
		if !ok {
			return node.Result{}, fmt.Errorf("foreach requires items array")
		}
		if len(items) > maxIters {
			return node.Result{}, fmt.Errorf("items length %d exceeds max_iters %d", len(items), maxIters)
		}
		for i, item := range items {
			iterations = append(iterations, map[string]interface{}{
				indexVar: i,
				asVar:    item,
			})
		}

	default:
		return node.Result{}, fmt.Errorf("unknown loop type: %s (valid: times, for, foreach)", loopType)
	}

	return node.Result{
		Success: true,
		Data: map[string]interface{}{
			"iterations": iterations,
			"count":      len(iterations),
		},
		Status: fmt.Sprintf("%d iterations", len(iterations)),
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Generated %d iterations", len(iterations))},
		},
	}, nil
}
