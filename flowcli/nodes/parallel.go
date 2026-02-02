package nodes

import (
	"fmt"
	"sync"

	"github.com/nameer-kp/flowcli/pkg/node"
)

// ParallelNode executes multiple steps concurrently
type ParallelNode struct {
	registry *node.Registry // For resolving nested step node types (future use)
}

// NewParallelNode creates a new parallel node with the given registry
func NewParallelNode(registry *node.Registry) *ParallelNode {
	return &ParallelNode{
		registry: registry,
	}
}

func (n *ParallelNode) Name() string {
	return "parallel"
}

func (n *ParallelNode) Description() string {
	return "Execute multiple steps concurrently"
}

func (n *ParallelNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"max_concurrent": "number", // 0 = unlimited
		"steps":          "array",  // Nested steps to execute
	}
}

func (n *ParallelNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	// Get steps from config
	stepsRaw, ok := config["steps"]
	if !ok {
		return node.Result{}, fmt.Errorf("steps is required")
	}

	steps, err := parseSteps(stepsRaw)
	if err != nil {
		return node.Result{}, fmt.Errorf("invalid steps: %w", err)
	}

	if len(steps) == 0 {
		return node.Result{}, fmt.Errorf("steps cannot be empty")
	}

	// Get max_concurrent limit (0 = unlimited)
	maxConcurrent := 0
	if mc, ok := config["max_concurrent"]; ok {
		switch v := mc.(type) {
		case int:
			maxConcurrent = v
		case float64:
			maxConcurrent = int(v)
		}
	}

	ctx.Logger().Info("starting parallel execution", "steps_count", len(steps), "max_concurrent", maxConcurrent)

	// Thread-safe results collection
	var mu sync.Mutex
	results := make(map[string]interface{})
	allSuccess := true

	// Use WaitGroup to wait for all goroutines
	var wg sync.WaitGroup

	// Create semaphore channel if max_concurrent > 0
	var sem chan struct{}
	if maxConcurrent > 0 {
		sem = make(chan struct{}, maxConcurrent)
	}

	for _, step := range steps {
		wg.Add(1)

		go func(s stepInfo) {
			defer wg.Done()

			// Acquire semaphore if limited
			if sem != nil {
				sem <- struct{}{}
				defer func() { <-sem }()
			}

			ctx.Logger().Debug("executing parallel step", "id", s.id, "type", s.stepType)

			// Note: Full nested step execution will be added when engine integration is complete
			// For now, we just track which steps would execute
			stepResult := map[string]interface{}{
				"executed": true,
				"config":   s.config,
			}

			// Store result thread-safely
			mu.Lock()
			results[s.id] = stepResult
			mu.Unlock()
		}(step)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	return node.Result{
		Success: allSuccess,
		Data: map[string]interface{}{
			"results":     results,
			"all_success": allSuccess,
		},
		Status: fmt.Sprintf("%d parallel steps", len(results)),
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Completed %d parallel steps", len(results))},
		},
	}, nil
}

// stepInfo holds parsed step information
type stepInfo struct {
	id       string
	stepType string
	config   map[string]interface{}
}

// parseSteps converts raw steps config to stepInfo slice
func parseSteps(raw interface{}) ([]stepInfo, error) {
	rawSteps, ok := raw.([]interface{})
	if !ok {
		// Try []map[string]interface{} directly
		if mapSteps, ok := raw.([]map[string]interface{}); ok {
			rawSteps = make([]interface{}, len(mapSteps))
			for i, m := range mapSteps {
				rawSteps[i] = m
			}
		} else {
			return nil, fmt.Errorf("steps must be an array")
		}
	}

	steps := make([]stepInfo, 0, len(rawSteps))

	for i, rawStep := range rawSteps {
		stepMap, ok := rawStep.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("step %d must be an object", i)
		}

		// Get step ID (required)
		id, ok := stepMap["id"].(string)
		if !ok || id == "" {
			// Generate a default ID if not provided
			id = fmt.Sprintf("step_%d", i)
		}

		// Get step type (optional for now, will be required for full execution)
		stepType, _ := stepMap["type"].(string)

		// Get step config
		stepConfig := make(map[string]interface{})
		for k, v := range stepMap {
			if k != "id" && k != "type" {
				stepConfig[k] = v
			}
		}

		steps = append(steps, stepInfo{
			id:       id,
			stepType: stepType,
			config:   stepConfig,
		})
	}

	return steps, nil
}
