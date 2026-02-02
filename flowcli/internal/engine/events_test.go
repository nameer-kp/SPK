package engine

import (
	"testing"
	"time"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventStepStart, "StepStart"},
		{EventStepComplete, "StepComplete"},
		{EventStepError, "StepError"},
		{EventStepSkipped, "StepSkipped"},
		{EventWorkflowComplete, "WorkflowComplete"},
		{EventPaused, "Paused"},
		{EventResumed, "Resumed"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.eventType.String()
			if got != tt.expected {
				t.Errorf("EventType.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEventType_String_Unknown(t *testing.T) {
	unknown := EventType(99)
	got := unknown.String()
	if got != "Unknown" {
		t.Errorf("Unknown EventType.String() = %q, want %q", got, "Unknown")
	}
}

func TestNewEventEmitter(t *testing.T) {
	ch := make(chan ExecutionEvent, 1)
	emitter := NewEventEmitter(ch)

	if emitter == nil {
		t.Fatal("NewEventEmitter returned nil")
	}
	if emitter.ch != ch {
		t.Error("NewEventEmitter did not store the channel correctly")
	}
}

func TestEventEmitter_Emit(t *testing.T) {
	ch := make(chan ExecutionEvent, 10)
	emitter := NewEventEmitter(ch)

	step := &Step{
		ID:   "test-step",
		Name: "Test Step",
		Type: "http",
	}

	result := &StepResult{
		StepID:  "test-step",
		Success: true,
		Data:    map[string]string{"key": "value"},
	}

	event := ExecutionEvent{
		Type:      EventStepComplete,
		Step:      step,
		Result:    result,
		Message:   "Step completed successfully",
		Timestamp: time.Now(),
	}

	emitter.Emit(event)

	select {
	case received := <-ch:
		if received.Type != EventStepComplete {
			t.Errorf("Received event type = %v, want %v", received.Type, EventStepComplete)
		}
		if received.Step.ID != "test-step" {
			t.Errorf("Received step ID = %q, want %q", received.Step.ID, "test-step")
		}
		if received.Result.Success != true {
			t.Errorf("Received result success = %v, want %v", received.Result.Success, true)
		}
		if received.Message != "Step completed successfully" {
			t.Errorf("Received message = %q, want %q", received.Message, "Step completed successfully")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected event was not received within timeout")
	}
}

func TestEventEmitter_Emit_AllEventTypes(t *testing.T) {
	ch := make(chan ExecutionEvent, 10)
	emitter := NewEventEmitter(ch)

	eventTypes := []EventType{
		EventStepStart,
		EventStepComplete,
		EventStepError,
		EventStepSkipped,
		EventWorkflowComplete,
		EventPaused,
		EventResumed,
	}

	for _, et := range eventTypes {
		emitter.Emit(ExecutionEvent{
			Type:      et,
			Timestamp: time.Now(),
		})
	}

	for _, expected := range eventTypes {
		select {
		case received := <-ch:
			if received.Type != expected {
				t.Errorf("Received event type = %v, want %v", received.Type, expected)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Expected event %v was not received within timeout", expected)
		}
	}
}

func TestEventEmitter_Emit_NonBlocking(t *testing.T) {
	// Create a channel with buffer size 1 (will be full after first event)
	ch := make(chan ExecutionEvent, 1)
	emitter := NewEventEmitter(ch)

	// Fill the channel
	emitter.Emit(ExecutionEvent{
		Type:      EventStepStart,
		Timestamp: time.Now(),
	})

	// This should not block even though the channel is full
	done := make(chan bool, 1)
	go func() {
		emitter.Emit(ExecutionEvent{
			Type:      EventStepComplete,
			Timestamp: time.Now(),
		})
		done <- true
	}()

	select {
	case <-done:
		// Success - Emit returned without blocking
	case <-time.After(100 * time.Millisecond):
		t.Error("Emit blocked when channel was full - should be non-blocking")
	}
}

func TestEventEmitter_Emit_WithError(t *testing.T) {
	ch := make(chan ExecutionEvent, 1)
	emitter := NewEventEmitter(ch)

	testErr := &testError{message: "test error"}

	event := ExecutionEvent{
		Type:      EventStepError,
		Step:      &Step{ID: "failing-step"},
		Error:     testErr,
		Message:   "Step failed",
		Timestamp: time.Now(),
	}

	emitter.Emit(event)

	select {
	case received := <-ch:
		if received.Type != EventStepError {
			t.Errorf("Received event type = %v, want %v", received.Type, EventStepError)
		}
		if received.Error == nil {
			t.Error("Expected error in event, got nil")
		}
		if received.Error.Error() != "test error" {
			t.Errorf("Received error = %q, want %q", received.Error.Error(), "test error")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected event was not received within timeout")
	}
}

func TestEventEmitter_Emit_NilChannel(t *testing.T) {
	// Emitter with nil channel should not panic
	emitter := &EventEmitter{ch: nil}

	// This should not panic
	done := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Emit panicked with nil channel: %v", r)
			}
			done <- true
		}()
		emitter.Emit(ExecutionEvent{
			Type:      EventStepStart,
			Timestamp: time.Now(),
		})
	}()

	select {
	case <-done:
		// Success - no panic
	case <-time.After(100 * time.Millisecond):
		t.Error("Emit with nil channel blocked or panicked")
	}
}

func TestExecutionEvent_Fields(t *testing.T) {
	now := time.Now()
	step := &Step{ID: "my-step", Name: "My Step", Type: "shell"}
	result := &StepResult{StepID: "my-step", Success: true, Duration: 1.5}
	err := &testError{message: "something went wrong"}

	event := ExecutionEvent{
		Type:      EventStepComplete,
		Step:      step,
		Result:    result,
		Error:     err,
		Message:   "Custom message",
		Timestamp: now,
	}

	if event.Type != EventStepComplete {
		t.Errorf("Type = %v, want %v", event.Type, EventStepComplete)
	}
	if event.Step != step {
		t.Error("Step was not stored correctly")
	}
	if event.Result != result {
		t.Error("Result was not stored correctly")
	}
	if event.Error != err {
		t.Error("Error was not stored correctly")
	}
	if event.Message != "Custom message" {
		t.Errorf("Message = %q, want %q", event.Message, "Custom message")
	}
	if !event.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", event.Timestamp, now)
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
