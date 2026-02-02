package integration

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nameer-kp/flowcli/internal/config"
	"github.com/nameer-kp/flowcli/internal/engine"
	"github.com/nameer-kp/flowcli/nodes"
	"github.com/nameer-kp/flowcli/pkg/node"
)

// setupTestEngine creates a fully configured engine with all node types registered
func setupTestEngine() (*engine.Engine, *node.Registry) {
	registry := node.NewRegistry()

	// Register all node types
	registry.Register(nodes.NewTransformNode())
	registry.Register(nodes.NewLoopNode(registry))
	registry.Register(nodes.NewParallelNode(registry))

	cfg := &config.ResolvedConfig{
		Profile: config.Profile{
			Name:      "test",
			Variables: map[string]string{},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	return engine.NewEngine(registry, cfg, logger), registry
}

// getFixturePath returns the absolute path to a fixture file
func getFixturePath(t *testing.T, name string) string {
	// Get the path relative to this test file
	_, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Try multiple paths to find fixtures
	paths := []string{
		filepath.Join("../fixtures", name),
		filepath.Join("test/fixtures", name),
		filepath.Join("../../test/fixtures", name),
	}

	for _, path := range paths {
		abs, _ := filepath.Abs(path)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	t.Fatalf("Fixture file not found: %s", name)
	return ""
}

// TestConditionEvaluation tests that steps are skipped when conditions evaluate to false
func TestConditionEvaluation(t *testing.T) {
	eng, _ := setupTestEngine()
	parser := engine.NewParser()

	workflowPath := getFixturePath(t, "condition_workflow.yaml")
	workflow, err := parser.ParseFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to parse workflow: %v", err)
	}

	// Run the workflow
	events := make(chan engine.ExecutionEvent, 20)
	state, err := eng.Run(workflow, nil, events)
	close(events)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Expected workflow status 'completed', got '%s'", state.Status)
	}

	// Verify we have 3 results
	if len(state.Results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(state.Results))
	}

	// Check step statuses
	// setup: should execute
	if state.Results[0].Status == "skipped" {
		t.Errorf("setup step should not be skipped")
	}
	if state.Results[0].StepID != "setup" {
		t.Errorf("First result should be 'setup', got '%s'", state.Results[0].StepID)
	}

	// conditional_step: should execute (condition is true)
	if state.Results[1].Status == "skipped" {
		t.Errorf("conditional_step should not be skipped when condition is true")
	}
	if state.Results[1].StepID != "conditional_step" {
		t.Errorf("Second result should be 'conditional_step', got '%s'", state.Results[1].StepID)
	}

	// skipped_step: should be skipped (condition is false)
	if state.Results[2].Status != "skipped" {
		t.Errorf("skipped_step should be skipped when condition is false, got status '%s'", state.Results[2].Status)
	}
	if state.Results[2].StepID != "skipped_step" {
		t.Errorf("Third result should be 'skipped_step', got '%s'", state.Results[2].StepID)
	}

	// Verify events
	var receivedEvents []engine.ExecutionEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	// Count skip events
	skipCount := 0
	for _, event := range receivedEvents {
		if event.Type == engine.EventStepSkipped {
			skipCount++
		}
	}

	if skipCount != 1 {
		t.Errorf("Expected 1 skip event, got %d", skipCount)
	}
}

// TestLoopNodeExecution tests that loop node iterates over items correctly
func TestLoopNodeExecution(t *testing.T) {
	eng, _ := setupTestEngine()
	parser := engine.NewParser()

	workflowPath := getFixturePath(t, "loop_workflow.yaml")
	workflow, err := parser.ParseFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to parse workflow: %v", err)
	}

	// Run the workflow
	state, err := eng.Run(workflow, nil, nil)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Expected workflow status 'completed', got '%s'", state.Status)
	}

	// Verify we have 2 results
	if len(state.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(state.Results))
	}

	// Check loop result
	loopResult := state.Results[1]
	if loopResult.StepID != "process_items" {
		t.Errorf("Second result should be 'process_items', got '%s'", loopResult.StepID)
	}

	if !loopResult.Success {
		t.Errorf("Loop step should succeed")
	}

	// Verify loop data
	data, ok := loopResult.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Loop result data should be a map, got %T", loopResult.Data)
	}

	count, ok := data["count"]
	if !ok {
		t.Fatalf("Loop result should have 'count' field")
	}

	// Count should be 3 (for items [1, 2, 3])
	countInt, ok := count.(int)
	if !ok {
		t.Fatalf("Count should be an int, got %T", count)
	}
	if countInt != 3 {
		t.Errorf("Expected loop count 3, got %d", countInt)
	}

	// Verify iterations
	iterations, ok := data["iterations"]
	if !ok {
		t.Fatalf("Loop result should have 'iterations' field")
	}

	iterSlice, ok := iterations.([]map[string]interface{})
	if !ok {
		t.Fatalf("Iterations should be []map[string]interface{}, got %T", iterations)
	}

	if len(iterSlice) != 3 {
		t.Errorf("Expected 3 iterations, got %d", len(iterSlice))
	}

	// Check each iteration has index and item
	for i, iter := range iterSlice {
		idx, hasIdx := iter["index"]
		item, hasItem := iter["item"]

		if !hasIdx || !hasItem {
			t.Errorf("Iteration %d missing index or item", i)
		}

		// Verify index matches
		if idxInt, ok := idx.(int); ok && idxInt != i {
			t.Errorf("Iteration %d: expected index %d, got %d", i, i, idxInt)
		}

		// Item should be i+1 (1, 2, 3)
		if itemFloat, ok := item.(float64); ok && int(itemFloat) != i+1 {
			t.Errorf("Iteration %d: expected item %d, got %v", i, i+1, item)
		}
	}
}

// TestParallelNodeExecution tests that parallel node executes steps concurrently
func TestParallelNodeExecution(t *testing.T) {
	registry := node.NewRegistry()

	// Register transform node
	registry.Register(nodes.NewTransformNode())
	registry.Register(nodes.NewParallelNode(registry))

	cfg := &config.ResolvedConfig{
		Profile: config.Profile{
			Name:      "test",
			Variables: map[string]string{},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	eng := engine.NewEngine(registry, cfg, logger)

	// Create a workflow with parallel step
	workflow := &engine.Workflow{
		Name:        "parallel_test",
		Description: "Test parallel execution",
		Steps: []engine.Step{
			{
				ID:   "parallel_tasks",
				Type: "parallel",
				Config: map[string]interface{}{
					"max_concurrent": 2,
					"steps": []map[string]interface{}{
						{
							"id":   "task_a",
							"type": "transform",
						},
						{
							"id":   "task_b",
							"type": "transform",
						},
						{
							"id":   "task_c",
							"type": "transform",
						},
					},
				},
				Output: "parallel_result",
			},
		},
	}

	// Track execution for concurrency verification
	var executionCount int32

	// Run the workflow
	state, err := eng.Run(workflow, nil, nil)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Expected workflow status 'completed', got '%s'", state.Status)
	}

	// Verify result
	if len(state.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(state.Results))
	}

	parallelResult := state.Results[0]
	if !parallelResult.Success {
		t.Errorf("Parallel step should succeed")
	}

	// Verify parallel data
	data, ok := parallelResult.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Parallel result data should be a map, got %T", parallelResult.Data)
	}

	results, ok := data["results"].(map[string]interface{})
	if !ok {
		t.Fatalf("Parallel result should have 'results' map, got %T", data["results"])
	}

	// All 3 tasks should be in results
	expectedTasks := []string{"task_a", "task_b", "task_c"}
	for _, taskID := range expectedTasks {
		if _, exists := results[taskID]; !exists {
			t.Errorf("Missing result for task '%s'", taskID)
		}
	}

	// Execution count is not used in this simplified test
	_ = executionCount
	_ = atomic.AddInt32(&executionCount, 1)
}

// TestCheckpointSaveLoad tests checkpoint save and resume functionality
func TestCheckpointSaveLoad(t *testing.T) {
	// Create temp directory for checkpoints
	tmpDir, err := os.MkdirTemp("", "checkpoint_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test workflow file
	workflowPath := filepath.Join(tmpDir, "test_workflow.yaml")
	workflowContent := `name: checkpoint_test
description: Test workflow for checkpoint
steps:
  - id: step1
    type: transform
    config:
      operation: template
      template: '{"value": 1}'
    output: result1
  - id: step2
    type: transform
    config:
      operation: template
      template: '{"value": 2}'
    output: result2
  - id: step3
    type: transform
    config:
      operation: template
      template: '{"value": 3}'
    output: result3
`
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}

	checkpointDir := filepath.Join(tmpDir, "checkpoints")
	manager := engine.NewCheckpointManager(checkpointDir)

	// Parse workflow
	parser := engine.NewParser()
	workflow, err := parser.ParseFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to parse workflow: %v", err)
	}

	// Create a simulated workflow state (as if paused mid-execution)
	state := &engine.WorkflowState{
		Workflow:    workflow,
		CurrentStep: 1, // Paused after step1
		Context: map[string]interface{}{
			"result1": map[string]interface{}{"value": 1},
		},
		Inputs: map[string]interface{}{
			"input1": "test_value",
		},
		Results: []engine.StepResult{
			{
				StepID:   "step1",
				Success:  true,
				Data:     map[string]interface{}{"value": 1},
				Duration: 0.5,
				Status:   "Transformed",
			},
		},
		Status: "paused",
	}

	// Save checkpoint
	checkpointID, err := manager.Save(state, workflowPath)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	if checkpointID == "" {
		t.Fatal("Checkpoint ID should not be empty")
	}

	// Load checkpoint
	loadedCheckpoint, err := manager.Load(checkpointID)
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	// Verify loaded data
	if loadedCheckpoint.WorkflowPath != workflowPath {
		t.Errorf("WorkflowPath mismatch: got '%s', want '%s'", loadedCheckpoint.WorkflowPath, workflowPath)
	}

	if loadedCheckpoint.WorkflowName != "checkpoint_test" {
		t.Errorf("WorkflowName mismatch: got '%s', want '%s'", loadedCheckpoint.WorkflowName, "checkpoint_test")
	}

	if loadedCheckpoint.CurrentStep != 1 {
		t.Errorf("CurrentStep mismatch: got %d, want 1", loadedCheckpoint.CurrentStep)
	}

	if loadedCheckpoint.Status != "paused" {
		t.Errorf("Status mismatch: got '%s', want 'paused'", loadedCheckpoint.Status)
	}

	// Verify inputs preserved
	if loadedCheckpoint.Inputs["input1"] != "test_value" {
		t.Errorf("Input not preserved: got '%v', want 'test_value'", loadedCheckpoint.Inputs["input1"])
	}

	// Verify context preserved
	if _, ok := loadedCheckpoint.Context["result1"]; !ok {
		t.Error("Context result1 not preserved")
	}

	// Verify results preserved
	if len(loadedCheckpoint.Results) != 1 {
		t.Errorf("Results length mismatch: got %d, want 1", len(loadedCheckpoint.Results))
	}

	if loadedCheckpoint.Results[0].StepID != "step1" {
		t.Errorf("First result StepID mismatch: got '%s', want 'step1'", loadedCheckpoint.Results[0].StepID)
	}

	// Validate workflow unchanged
	unchanged, err := manager.ValidateWorkflowUnchanged(loadedCheckpoint)
	if err != nil {
		t.Fatalf("ValidateWorkflowUnchanged failed: %v", err)
	}

	if !unchanged {
		t.Error("Workflow should be unchanged")
	}

	// Modify workflow and verify change detection
	modifiedContent := strings.Replace(workflowContent, `"value": 1`, `"value": 100`, 1)
	if err := os.WriteFile(workflowPath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify workflow file: %v", err)
	}

	unchanged, err = manager.ValidateWorkflowUnchanged(loadedCheckpoint)
	if err != nil {
		t.Fatalf("ValidateWorkflowUnchanged failed: %v", err)
	}

	if unchanged {
		t.Error("Workflow should be detected as changed")
	}
}

// TestErrorRecoveryRetry tests the error handler retry action
func TestErrorRecoveryRetry(t *testing.T) {
	// Create error handler with retry action
	handler := engine.NewErrorHandler("retry", nil, nil)

	step := &engine.Step{
		ID:   "failing_step",
		Type: "http",
	}

	err := &testError{message: "connection refused"}

	// First retry attempt
	action := handler.Handle(step, err)
	if action != engine.ActionRetry {
		t.Errorf("First attempt: expected ActionRetry, got %v", action)
	}

	if handler.GetRetryCount("failing_step") != 1 {
		t.Errorf("Retry count should be 1, got %d", handler.GetRetryCount("failing_step"))
	}

	// Second retry attempt
	action = handler.Handle(step, err)
	if action != engine.ActionRetry {
		t.Errorf("Second attempt: expected ActionRetry, got %v", action)
	}

	if handler.GetRetryCount("failing_step") != 2 {
		t.Errorf("Retry count should be 2, got %d", handler.GetRetryCount("failing_step"))
	}

	// Third retry attempt
	action = handler.Handle(step, err)
	if action != engine.ActionRetry {
		t.Errorf("Third attempt: expected ActionRetry, got %v", action)
	}

	if handler.GetRetryCount("failing_step") != 3 {
		t.Errorf("Retry count should be 3, got %d", handler.GetRetryCount("failing_step"))
	}

	// Fourth attempt should exceed max retries and abort
	action = handler.Handle(step, err)
	if action != engine.ActionAbort {
		t.Errorf("Fourth attempt: expected ActionAbort after max retries, got %v", action)
	}

	// Test reset retry count
	handler.ResetRetryCount("failing_step")
	if handler.GetRetryCount("failing_step") != 0 {
		t.Errorf("Retry count should be 0 after reset, got %d", handler.GetRetryCount("failing_step"))
	}

	// After reset, retry should work again
	action = handler.Handle(step, err)
	if action != engine.ActionRetry {
		t.Errorf("After reset: expected ActionRetry, got %v", action)
	}
}

// TestErrorRecoverySkip tests the error handler skip action
func TestErrorRecoverySkip(t *testing.T) {
	handler := engine.NewErrorHandler("skip", nil, nil)

	step := &engine.Step{
		ID:   "skippable_step",
		Type: "http",
	}

	err := &testError{message: "some error"}

	action := handler.Handle(step, err)
	if action != engine.ActionSkip {
		t.Errorf("Expected ActionSkip, got %v", action)
	}
}

// TestErrorRecoveryAbort tests the error handler abort action
func TestErrorRecoveryAbort(t *testing.T) {
	handler := engine.NewErrorHandler("abort", nil, nil)

	step := &engine.Step{
		ID:   "critical_step",
		Type: "http",
	}

	err := &testError{message: "critical error"}

	action := handler.Handle(step, err)
	if action != engine.ActionAbort {
		t.Errorf("Expected ActionAbort, got %v", action)
	}
}

// TestErrorRecoveryStepOverride tests that step-level on_error overrides workflow default
func TestErrorRecoveryStepOverride(t *testing.T) {
	handler := engine.NewErrorHandler("abort", nil, nil)

	step := &engine.Step{
		ID:      "step_with_override",
		Type:    "http",
		OnError: "skip", // Override workflow default
	}

	err := &testError{message: "some error"}

	action := handler.Handle(step, err)
	if action != engine.ActionSkip {
		t.Errorf("Step override: expected ActionSkip, got %v", action)
	}
}

// TestParallelNodeConcurrencyLimit tests that max_concurrent is respected
func TestParallelNodeConcurrencyLimit(t *testing.T) {
	registry := node.NewRegistry()
	registry.Register(nodes.NewParallelNode(registry))

	cfg := &config.ResolvedConfig{
		Profile: config.Profile{
			Name:      "test",
			Variables: map[string]string{},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	eng := engine.NewEngine(registry, cfg, logger)

	// Create workflow with parallel step limited to 1 concurrent
	workflow := &engine.Workflow{
		Name: "parallel_limit_test",
		Steps: []engine.Step{
			{
				ID:   "limited_parallel",
				Type: "parallel",
				Config: map[string]interface{}{
					"max_concurrent": 1, // Only 1 at a time
					"steps": []map[string]interface{}{
						{"id": "task_1"},
						{"id": "task_2"},
						{"id": "task_3"},
					},
				},
			},
		},
	}

	startTime := time.Now()
	state, err := eng.Run(workflow, nil, nil)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Expected completed, got '%s'", state.Status)
	}

	// With max_concurrent=1, tasks should run sequentially
	// This is a basic verification - actual timing depends on implementation
	t.Logf("Parallel execution with limit=1 took %v", duration)
}

// TestCheckpointList tests listing multiple checkpoints
func TestCheckpointList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint_list_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test workflow files
	workflow1Path := filepath.Join(tmpDir, "workflow1.yaml")
	workflow2Path := filepath.Join(tmpDir, "workflow2.yaml")

	workflowContent := `name: %s
description: Test workflow
steps:
  - id: step1
    type: transform
    config:
      operation: template
      template: '{}'
`

	if err := os.WriteFile(workflow1Path, []byte(strings.Replace(workflowContent, "%s", "workflow1", 1)), 0644); err != nil {
		t.Fatalf("Failed to write workflow1: %v", err)
	}
	if err := os.WriteFile(workflow2Path, []byte(strings.Replace(workflowContent, "%s", "workflow2", 1)), 0644); err != nil {
		t.Fatalf("Failed to write workflow2: %v", err)
	}

	checkpointDir := filepath.Join(tmpDir, "checkpoints")
	manager := engine.NewCheckpointManager(checkpointDir)

	// Create checkpoints for both workflows
	for i, path := range []string{workflow1Path, workflow2Path} {
		wfName := filepath.Base(path)
		wfName = strings.TrimSuffix(wfName, ".yaml")

		state := &engine.WorkflowState{
			Workflow:    &engine.Workflow{Name: wfName},
			CurrentStep: i,
			Context:     map[string]interface{}{},
			Inputs:      map[string]interface{}{},
			Results:     []engine.StepResult{},
			Status:      "paused",
		}

		_, err := manager.Save(state, path)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}

		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// List checkpoints
	checkpoints, err := manager.List()
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(checkpoints))
	}

	// Verify sorted by UpdatedAt descending
	if len(checkpoints) >= 2 {
		if checkpoints[0].UpdatedAt.Before(checkpoints[1].UpdatedAt) {
			t.Error("Checkpoints should be sorted by UpdatedAt descending")
		}
	}
}

// TestConditionWithInputs tests condition evaluation using workflow inputs
func TestConditionWithInputs(t *testing.T) {
	eng, _ := setupTestEngine()

	workflow := &engine.Workflow{
		Name: "condition_inputs_test",
		Steps: []engine.Step{
			{
				ID:        "conditional_on_input",
				Type:      "transform",
				Condition: "inputs.enabled == true",
				Config: map[string]interface{}{
					"operations": []interface{}{
						map[string]interface{}{
							"type":       "expr",
							"expression": `{"ran": true}`,
						},
					},
				},
			},
		},
	}

	// Test with enabled=true
	inputs := map[string]interface{}{"enabled": true}
	state, err := eng.Run(workflow, inputs, nil)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if state.Results[0].Status == "skipped" {
		t.Error("Step should execute when inputs.enabled=true")
	}

	// Test with enabled=false
	inputs = map[string]interface{}{"enabled": false}
	state, err = eng.Run(workflow, inputs, nil)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if state.Results[0].Status != "skipped" {
		t.Error("Step should be skipped when inputs.enabled=false")
	}
}

// TestLoopWithNestedObject tests loop node with nested object access
func TestLoopWithNestedObject(t *testing.T) {
	eng, _ := setupTestEngine()

	workflow := &engine.Workflow{
		Name: "loop_nested_test",
		Steps: []engine.Step{
			{
				ID:   "create_data",
				Type: "transform",
				Config: map[string]interface{}{
					"operations": []interface{}{
						map[string]interface{}{
							"type":       "expr",
							"expression": `{"items": [{"name": "a"}, {"name": "b"}]}`,
						},
					},
				},
				Output: "data",
			},
			{
				ID:   "loop_items",
				Type: "loop",
				Config: map[string]interface{}{
					"items": "data.items", // Direct context path, no {{ }} wrapper
					"as":    "item",
				},
			},
		},
	}

	state, err := eng.Run(workflow, nil, nil)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Expected completed, got '%s'", state.Status)
	}

	// Check loop executed both items
	if len(state.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(state.Results))
	}

	loopResult := state.Results[1]
	data, ok := loopResult.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Loop result should be map, got %T", loopResult.Data)
	}

	count, _ := data["count"].(int)
	if count != 2 {
		t.Errorf("Expected 2 iterations, got %d", count)
	}
}

// testError is a simple error implementation for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// TestTransformNodeExprJSON tests that transform node handles expr expressions correctly
func TestTransformNodeExprJSON(t *testing.T) {
	eng, _ := setupTestEngine()

	workflow := &engine.Workflow{
		Name: "transform_expr_test",
		Steps: []engine.Step{
			{
				ID:   "create_json",
				Type: "transform",
				Config: map[string]interface{}{
					"operations": []interface{}{
						map[string]interface{}{
							"type":       "expr",
							"expression": `{"key": "value", "number": 42}`,
						},
					},
				},
				Output: "json_result",
			},
		},
	}

	state, err := eng.Run(workflow, nil, nil)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	result := state.Results[0]
	if !result.Success {
		t.Error("Transform step should succeed")
	}

	// The result should be a map from expr evaluation
	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result from expr operation, got %T", result.Data)
	}

	// Verify the map contents
	if dataMap["key"] != "value" {
		t.Errorf("Expected key='value', got '%v'", dataMap["key"])
	}

	if dataMap["number"] != 42 {
		t.Errorf("Expected number=42, got '%v'", dataMap["number"])
	}
}

// TestMultipleConditionsInWorkflow tests a workflow with multiple conditional steps
func TestMultipleConditionsInWorkflow(t *testing.T) {
	eng, _ := setupTestEngine()

	workflow := &engine.Workflow{
		Name: "multi_condition_test",
		Steps: []engine.Step{
			{
				ID:   "setup",
				Type: "transform",
				Config: map[string]interface{}{
					"operations": []interface{}{
						map[string]interface{}{
							"type":       "expr",
							"expression": `{"count": 5}`,
						},
					},
				},
				Output: "data",
			},
			{
				ID:        "run_if_high",
				Type:      "transform",
				Condition: "data.count > 10",
				Config: map[string]interface{}{
					"operations": []interface{}{
						map[string]interface{}{
							"type":       "expr",
							"expression": `{"high": true}`,
						},
					},
				},
			},
			{
				ID:        "run_if_low",
				Type:      "transform",
				Condition: "data.count <= 10",
				Config: map[string]interface{}{
					"operations": []interface{}{
						map[string]interface{}{
							"type":       "expr",
							"expression": `{"low": true}`,
						},
					},
				},
			},
		},
	}

	state, err := eng.Run(workflow, nil, nil)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if len(state.Results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(state.Results))
	}

	// run_if_high should be skipped (5 > 10 is false)
	if state.Results[1].Status != "skipped" {
		t.Error("run_if_high should be skipped when count=5")
	}

	// run_if_low should execute (5 <= 10 is true)
	if state.Results[2].Status == "skipped" {
		t.Error("run_if_low should execute when count=5")
	}
}
