package nodes

import (
	"fmt"
	"sync"
	"time"

	"github.com/nameer-kp/atem/pkg/node"
)

// ParallelNode executes multiple operations concurrently
// Note: In current Atem architecture, this prepares parallel execution metadata
// Actual parallel execution would be handled by the engine
type ParallelNode struct{}

func NewParallelNode() *ParallelNode {
	return &ParallelNode{}
}

func (n *ParallelNode) Name() string {
	return "parallel"
}

func (n *ParallelNode) Description() string {
	return "Execute operations in parallel"
}

func (n *ParallelNode) ConfigSchema() map[string]any {
	return map[string]any{
		"tasks":       "array",  // array of task configs
		"max_workers": "number", // maximum concurrent workers
		"timeout":     "string", // overall timeout
		"fail_fast":   "bool",   // stop on first failure
		"collect":     "string", // how to collect results: all, first, any
	}
}

func (n *ParallelNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	tasks, ok := config["tasks"].([]any)
	if !ok || len(tasks) == 0 {
		return node.Result{}, fmt.Errorf("tasks array is required")
	}

	maxWorkers := len(tasks)
	if mw, ok := config["max_workers"].(float64); ok && mw > 0 {
		maxWorkers = int(mw)
	}

	timeout := 5 * time.Minute
	if t, ok := config["timeout"].(string); ok {
		if parsed, err := time.ParseDuration(t); err == nil {
			timeout = parsed
		}
	}

	failFast, _ := config["fail_fast"].(bool)
	collectMode, _ := config["collect"].(string)
	if collectMode == "" {
		collectMode = "all"
	}

	// Results collection
	type taskResult struct {
		index   int
		id      string
		success bool
		data    any
		err     error
	}

	results := make([]taskResult, len(tasks))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore for max workers
	sem := make(chan struct{}, maxWorkers)

	// Cancellation
	done := make(chan struct{})
	var cancelOnce sync.Once

	// Timeout context
	timeoutCh := time.After(timeout)

	for i, task := range tasks {
		taskConfig, ok := task.(map[string]any)
		if !ok {
			results[i] = taskResult{
				index:   i,
				success: false,
				err:     fmt.Errorf("task %d is not a valid config", i),
			}
			continue
		}

		wg.Add(1)
		go func(idx int, cfg map[string]any) {
			defer wg.Done()

			// Check for cancellation
			select {
			case <-done:
				return
			default:
			}

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-done:
				return
			case <-timeoutCh:
				mu.Lock()
				results[idx] = taskResult{
					index:   idx,
					success: false,
					err:     fmt.Errorf("task %d timed out waiting for worker", idx),
				}
				mu.Unlock()
				return
			}

			// Get task ID
			taskID, _ := cfg["id"].(string)
			if taskID == "" {
				taskID = fmt.Sprintf("task_%d", idx)
			}

			// Simulate task execution
			// In a real implementation, this would delegate to the engine
			simulateResult := map[string]any{
				"task_id": taskID,
				"index":   idx,
				"config":  cfg,
				"status":  "prepared",
			}

			mu.Lock()
			results[idx] = taskResult{
				index:   idx,
				id:      taskID,
				success: true,
				data:    simulateResult,
			}
			mu.Unlock()

			// Check fail_fast
			if failFast && results[idx].err != nil {
				cancelOnce.Do(func() { close(done) })
			}
		}(i, taskConfig)
	}

	// Wait with timeout
	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		// All done
	case <-timeoutCh:
		cancelOnce.Do(func() { close(done) })
		return node.Result{}, fmt.Errorf("parallel execution timed out after %v", timeout)
	}

	// Collect results
	var allResults []any
	var errors []string
	successCount := 0
	failCount := 0

	for _, r := range results {
		if r.success {
			successCount++
			allResults = append(allResults, r.data)
		} else if r.err != nil {
			failCount++
			errors = append(errors, r.err.Error())
		}
	}

	success := failCount == 0
	if collectMode == "any" {
		success = successCount > 0
	}

	return node.Result{
		Success: success,
		Data: map[string]any{
			"results":       allResults,
			"success_count": successCount,
			"fail_count":    failCount,
			"errors":        errors,
			"total":         len(tasks),
		},
		Status: fmt.Sprintf("%d/%d succeeded", successCount, len(tasks)),
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Parallel execution: %d succeeded, %d failed", successCount, failCount)},
		},
	}, nil
}
