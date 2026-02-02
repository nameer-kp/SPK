package engine

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/nameer-kp/flowcli/internal/config"
	"github.com/nameer-kp/flowcli/pkg/node"
)

// mockNode is a test node that can be configured to succeed or fail
type mockNode struct {
	name       string
	shouldFail bool
	data       interface{}
}

func (m *mockNode) Name() string {
	return m.name
}

func (m *mockNode) Description() string {
	return "Mock node for testing"
}

func (m *mockNode) ConfigSchema() map[string]interface{} {
	return nil
}

func (m *mockNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	if m.shouldFail {
		return node.Result{
			Success: false,
			Status:  "failed",
		}, nil
	}
	return node.Result{
		Success: true,
		Data:    m.data,
		Status:  "success",
	}, nil
}

// createTestEngine creates an engine with mock nodes for testing
func createTestEngine(nodes ...node.Node) *Engine {
	registry := node.NewRegistry()
	for _, n := range nodes {
		registry.Register(n)
	}

	cfg := &config.ResolvedConfig{
		Profile: config.Profile{
			Name:      "test",
			Variables: map[string]string{},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	return NewEngine(registry, cfg, logger)
}

func TestEngine_Run_WithConditionFalse(t *testing.T) {
	// Create a mock node
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{"status": 200},
	}

	engine := createTestEngine(mockHTTP)

	// Create a workflow with a step that has a false condition
	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:        "step1",
				Name:      "Conditional Step",
				Type:      "http",
				Condition: "false",
				Config:    map[string]interface{}{},
			},
		},
	}

	// Run without events channel
	state, err := engine.Run(workflow, nil, nil)

	if err != nil {
		t.Fatalf("Run() unexpected error = %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Run() status = %q, want %q", state.Status, "completed")
	}

	// Check that step was skipped
	if len(state.Results) != 1 {
		t.Fatalf("Run() results count = %d, want 1", len(state.Results))
	}

	if state.Results[0].Status != "skipped" {
		t.Errorf("Run() step status = %q, want %q", state.Results[0].Status, "skipped")
	}
}

func TestEngine_Run_WithConditionTrue(t *testing.T) {
	// Create a mock node
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{"status": 200},
	}

	engine := createTestEngine(mockHTTP)

	// Create a workflow with a step that has a true condition
	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:        "step1",
				Name:      "Conditional Step",
				Type:      "http",
				Condition: "true",
				Config:    map[string]interface{}{},
				Output:    "result",
			},
		},
	}

	state, err := engine.Run(workflow, nil, nil)

	if err != nil {
		t.Fatalf("Run() unexpected error = %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Run() status = %q, want %q", state.Status, "completed")
	}

	// Check that step was executed (not skipped)
	if len(state.Results) != 1 {
		t.Fatalf("Run() results count = %d, want 1", len(state.Results))
	}

	if state.Results[0].Status == "skipped" {
		t.Errorf("Run() step should have been executed, but was skipped")
	}

	if !state.Results[0].Success {
		t.Errorf("Run() step success = %v, want true", state.Results[0].Success)
	}
}

func TestEngine_Run_WithConditionBasedOnInput(t *testing.T) {
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{"response": "ok"},
	}

	engine := createTestEngine(mockHTTP)

	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:        "step1",
				Name:      "Input Conditional Step",
				Type:      "http",
				Condition: "inputs.enabled == true",
				Config:    map[string]interface{}{},
			},
		},
	}

	// Test with enabled = true
	inputs := map[string]interface{}{"enabled": true}
	state, err := engine.Run(workflow, inputs, nil)

	if err != nil {
		t.Fatalf("Run() unexpected error = %v", err)
	}

	if state.Results[0].Status == "skipped" {
		t.Errorf("Run() step should have been executed when inputs.enabled = true")
	}

	// Test with enabled = false
	inputs = map[string]interface{}{"enabled": false}
	state, err = engine.Run(workflow, inputs, nil)

	if err != nil {
		t.Fatalf("Run() unexpected error = %v", err)
	}

	if state.Results[0].Status != "skipped" {
		t.Errorf("Run() step should have been skipped when inputs.enabled = false")
	}
}

func TestEngine_Run_WithConditionBasedOnPreviousStep(t *testing.T) {
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{
			"success": true,
			"count":   100,
		},
	}

	engine := createTestEngine(mockHTTP)

	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:     "step1",
				Name:   "First Step",
				Type:   "http",
				Config: map[string]interface{}{},
				Output: "first_result",
			},
			{
				ID:        "step2",
				Name:      "Dependent Step",
				Type:      "http",
				Condition: "first_result.success == true && first_result.count > 50",
				Config:    map[string]interface{}{},
			},
		},
	}

	state, err := engine.Run(workflow, nil, nil)

	if err != nil {
		t.Fatalf("Run() unexpected error = %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Run() status = %q, want %q", state.Status, "completed")
	}

	// Both steps should be executed
	if len(state.Results) != 2 {
		t.Fatalf("Run() results count = %d, want 2", len(state.Results))
	}

	// Second step should NOT be skipped because condition is true
	if state.Results[1].Status == "skipped" {
		t.Errorf("Run() step2 should have been executed")
	}
}

func TestEngine_Run_ConditionEvaluationError(t *testing.T) {
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{},
	}

	engine := createTestEngine(mockHTTP)

	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:        "step1",
				Name:      "Invalid Condition Step",
				Type:      "http",
				Condition: "invalid && syntax &&",
				Config:    map[string]interface{}{},
			},
		},
	}

	_, err := engine.Run(workflow, nil, nil)

	if err == nil {
		t.Fatalf("Run() expected error for invalid condition, got nil")
	}

	// Check error message contains step ID
	if !containsStr(err.Error(), "step1") {
		t.Errorf("Run() error should mention step ID, got: %v", err)
	}

	if !containsStr(err.Error(), "condition evaluation failed") {
		t.Errorf("Run() error should mention condition evaluation, got: %v", err)
	}
}

func TestEngine_Run_EmitsEvents(t *testing.T) {
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{"status": 200},
	}

	engine := createTestEngine(mockHTTP)

	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:     "step1",
				Name:   "First Step",
				Type:   "http",
				Config: map[string]interface{}{},
			},
			{
				ID:     "step2",
				Name:   "Second Step",
				Type:   "http",
				Config: map[string]interface{}{},
			},
		},
	}

	// Create buffered channel to receive events
	events := make(chan ExecutionEvent, 10)

	state, err := engine.Run(workflow, nil, events)
	close(events)

	if err != nil {
		t.Fatalf("Run() unexpected error = %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Run() status = %q, want %q", state.Status, "completed")
	}

	// Collect all events
	var receivedEvents []ExecutionEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	// Expect: step1 start, step1 complete, step2 start, step2 complete, workflow complete
	expectedTypes := []EventType{
		EventStepStart,
		EventStepComplete,
		EventStepStart,
		EventStepComplete,
		EventWorkflowComplete,
	}

	if len(receivedEvents) != len(expectedTypes) {
		t.Fatalf("Run() emitted %d events, want %d", len(receivedEvents), len(expectedTypes))
	}

	for i, expected := range expectedTypes {
		if receivedEvents[i].Type != expected {
			t.Errorf("Event %d: type = %v, want %v", i, receivedEvents[i].Type, expected)
		}
	}

	// Check that step events have step reference
	if receivedEvents[0].Step == nil || receivedEvents[0].Step.ID != "step1" {
		t.Errorf("First event should reference step1")
	}

	// Check that complete events have result
	if receivedEvents[1].Result == nil {
		t.Errorf("Step complete event should have result")
	}

	// Check workflow complete event
	if receivedEvents[4].Type != EventWorkflowComplete {
		t.Errorf("Last event should be WorkflowComplete")
	}
}

func TestEngine_Run_EmitsSkipEvent(t *testing.T) {
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{},
	}

	engine := createTestEngine(mockHTTP)

	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:        "step1",
				Name:      "Skipped Step",
				Type:      "http",
				Condition: "false",
				Config:    map[string]interface{}{},
			},
		},
	}

	events := make(chan ExecutionEvent, 10)

	_, err := engine.Run(workflow, nil, events)
	close(events)

	if err != nil {
		t.Fatalf("Run() unexpected error = %v", err)
	}

	// Collect all events
	var receivedEvents []ExecutionEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	// Expect: step skipped, workflow complete
	if len(receivedEvents) != 2 {
		t.Fatalf("Run() emitted %d events, want 2", len(receivedEvents))
	}

	if receivedEvents[0].Type != EventStepSkipped {
		t.Errorf("First event type = %v, want %v", receivedEvents[0].Type, EventStepSkipped)
	}

	if receivedEvents[0].Step == nil || receivedEvents[0].Step.ID != "step1" {
		t.Errorf("Skip event should reference step1")
	}

	// Check that skip message contains condition info
	if receivedEvents[0].Message == "" {
		t.Errorf("Skip event should have message about condition")
	}

	if receivedEvents[1].Type != EventWorkflowComplete {
		t.Errorf("Second event type = %v, want %v", receivedEvents[1].Type, EventWorkflowComplete)
	}
}

func TestEngine_Run_EmitsErrorEvent(t *testing.T) {
	// Create a mock node that fails
	mockHTTP := &mockNode{
		name:       "http",
		shouldFail: true,
	}

	engine := createTestEngine(mockHTTP)

	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:     "step1",
				Name:   "Failing Step",
				Type:   "http",
				Config: map[string]interface{}{},
			},
		},
	}

	events := make(chan ExecutionEvent, 10)

	state, _ := engine.Run(workflow, nil, events)
	close(events)

	if state.Status != "failed" {
		t.Errorf("Run() status = %q, want %q", state.Status, "failed")
	}

	// Collect all events
	var receivedEvents []ExecutionEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	// Expect: step start, step error (no workflow complete because it failed)
	if len(receivedEvents) != 2 {
		t.Fatalf("Run() emitted %d events, want 2", len(receivedEvents))
	}

	if receivedEvents[0].Type != EventStepStart {
		t.Errorf("First event type = %v, want %v", receivedEvents[0].Type, EventStepStart)
	}

	if receivedEvents[1].Type != EventStepError {
		t.Errorf("Second event type = %v, want %v", receivedEvents[1].Type, EventStepError)
	}

	if receivedEvents[1].Result == nil {
		t.Errorf("Error event should have result")
	}
}

func TestEngine_Run_NilEventsChannel(t *testing.T) {
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{"status": 200},
	}

	engine := createTestEngine(mockHTTP)

	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:     "step1",
				Name:   "Step",
				Type:   "http",
				Config: map[string]interface{}{},
			},
		},
	}

	// Should not panic with nil events channel
	done := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Run() panicked with nil events channel: %v", r)
			}
			done <- true
		}()

		state, err := engine.Run(workflow, nil, nil)
		if err != nil {
			t.Errorf("Run() unexpected error = %v", err)
		}
		if state.Status != "completed" {
			t.Errorf("Run() status = %q, want %q", state.Status, "completed")
		}
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Error("Run() with nil events channel timed out")
	}
}

func TestEngine_Run_NoCondition(t *testing.T) {
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{"status": 200},
	}

	engine := createTestEngine(mockHTTP)

	// Step without any condition should always execute
	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:     "step1",
				Name:   "No Condition Step",
				Type:   "http",
				Config: map[string]interface{}{},
			},
		},
	}

	state, err := engine.Run(workflow, nil, nil)

	if err != nil {
		t.Fatalf("Run() unexpected error = %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Run() status = %q, want %q", state.Status, "completed")
	}

	if len(state.Results) != 1 {
		t.Fatalf("Run() results count = %d, want 1", len(state.Results))
	}

	if state.Results[0].Status == "skipped" {
		t.Errorf("Step without condition should not be skipped")
	}
}

func TestEngine_Run_MixedConditions(t *testing.T) {
	mockHTTP := &mockNode{
		name: "http",
		data: map[string]interface{}{"status": 200},
	}

	engine := createTestEngine(mockHTTP)

	workflow := &Workflow{
		Name: "Test Workflow",
		Steps: []Step{
			{
				ID:     "step1",
				Name:   "Always Run",
				Type:   "http",
				Config: map[string]interface{}{},
			},
			{
				ID:        "step2",
				Name:      "Skip This",
				Type:      "http",
				Condition: "false",
				Config:    map[string]interface{}{},
			},
			{
				ID:        "step3",
				Name:      "Run This",
				Type:      "http",
				Condition: "true",
				Config:    map[string]interface{}{},
			},
		},
	}

	events := make(chan ExecutionEvent, 20)
	state, err := engine.Run(workflow, nil, events)
	close(events)

	if err != nil {
		t.Fatalf("Run() unexpected error = %v", err)
	}

	if state.Status != "completed" {
		t.Errorf("Run() status = %q, want %q", state.Status, "completed")
	}

	// Check results
	if len(state.Results) != 3 {
		t.Fatalf("Run() results count = %d, want 3", len(state.Results))
	}

	// step1: executed
	if state.Results[0].Status == "skipped" {
		t.Errorf("step1 should not be skipped")
	}

	// step2: skipped
	if state.Results[1].Status != "skipped" {
		t.Errorf("step2 should be skipped, got status = %q", state.Results[1].Status)
	}

	// step3: executed
	if state.Results[2].Status == "skipped" {
		t.Errorf("step3 should not be skipped")
	}

	// Verify events
	var receivedEvents []ExecutionEvent
	for event := range events {
		receivedEvents = append(receivedEvents, event)
	}

	// Expected events:
	// step1: start, complete
	// step2: skipped
	// step3: start, complete
	// workflow: complete
	expectedTypes := []EventType{
		EventStepStart,
		EventStepComplete,
		EventStepSkipped,
		EventStepStart,
		EventStepComplete,
		EventWorkflowComplete,
	}

	if len(receivedEvents) != len(expectedTypes) {
		t.Fatalf("Run() emitted %d events, want %d", len(receivedEvents), len(expectedTypes))
	}

	for i, expected := range expectedTypes {
		if receivedEvents[i].Type != expected {
			t.Errorf("Event %d: type = %v, want %v", i, receivedEvents[i].Type, expected)
		}
	}
}

// containsStr checks if substr is in s
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
