package engine

import (
	"errors"
	"testing"
	"time"
)

func TestErrorHandler_HandleAbort(t *testing.T) {
	handler := NewErrorHandler("abort", nil, nil)
	step := &Step{ID: "test-step", Name: "Test Step"}
	err := errors.New("test error")

	action := handler.Handle(step, err)

	if action != ActionAbort {
		t.Errorf("Expected ActionAbort, got %v", action)
	}
}

func TestErrorHandler_HandleSkip(t *testing.T) {
	handler := NewErrorHandler("skip", nil, nil)
	step := &Step{ID: "test-step", Name: "Test Step"}
	err := errors.New("test error")

	action := handler.Handle(step, err)

	if action != ActionSkip {
		t.Errorf("Expected ActionSkip, got %v", action)
	}
}

func TestErrorHandler_HandleRetry(t *testing.T) {
	handler := NewErrorHandler("retry", nil, nil)
	step := &Step{ID: "test-step", Name: "Test Step"}
	err := errors.New("test error")

	// First 3 attempts should return ActionRetry
	for i := 0; i < MaxRetries; i++ {
		action := handler.Handle(step, err)
		if action != ActionRetry {
			t.Errorf("Attempt %d: Expected ActionRetry, got %v", i, action)
		}
		if handler.GetRetryCount(step.ID) != i+1 {
			t.Errorf("Attempt %d: Expected retry count %d, got %d", i, i+1, handler.GetRetryCount(step.ID))
		}
	}

	// 4th attempt should return ActionAbort (max retries exceeded)
	action := handler.Handle(step, err)
	if action != ActionAbort {
		t.Errorf("After max retries: Expected ActionAbort, got %v", action)
	}
}

func TestErrorHandler_HandlePause(t *testing.T) {
	events := make(chan ExecutionEvent, 1)
	decisions := make(chan RecoveryAction, 1)
	handler := NewErrorHandler("pause", events, decisions)
	step := &Step{ID: "test-step", Name: "Test Step"}
	err := errors.New("test error")

	// Send a decision before calling Handle (to avoid blocking)
	decisions <- ActionSkip

	action := handler.Handle(step, err)

	// Check that an event was emitted
	select {
	case event := <-events:
		if event.Type != EventStepError {
			t.Errorf("Expected EventStepError, got %v", event.Type)
		}
		if event.Step != step {
			t.Errorf("Expected step to match")
		}
		if event.Error != err {
			t.Errorf("Expected error to match")
		}
	default:
		t.Error("Expected an event to be emitted")
	}

	// Check that the decision was used
	if action != ActionSkip {
		t.Errorf("Expected ActionSkip (from decision), got %v", action)
	}
}

func TestErrorHandler_HandlePauseWithDifferentDecisions(t *testing.T) {
	testCases := []struct {
		name     string
		decision RecoveryAction
	}{
		{"Abort", ActionAbort},
		{"Skip", ActionSkip},
		{"Retry", ActionRetry},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			events := make(chan ExecutionEvent, 1)
			decisions := make(chan RecoveryAction, 1)
			handler := NewErrorHandler("pause", events, decisions)
			step := &Step{ID: "test-step", Name: "Test Step"}

			decisions <- tc.decision

			action := handler.Handle(step, errors.New("test error"))

			// Drain the event
			<-events

			if action != tc.decision {
				t.Errorf("Expected %v, got %v", tc.decision, action)
			}
		})
	}
}

func TestErrorHandler_StepOnErrorOverridesDefault(t *testing.T) {
	// Workflow default is abort
	handler := NewErrorHandler("abort", nil, nil)

	// Step overrides to skip
	step := &Step{ID: "test-step", Name: "Test Step", OnError: "skip"}
	err := errors.New("test error")

	action := handler.Handle(step, err)

	if action != ActionSkip {
		t.Errorf("Expected ActionSkip (step override), got %v", action)
	}
}

func TestErrorHandler_StepOnErrorOverridesDefaultRetry(t *testing.T) {
	// Workflow default is skip
	handler := NewErrorHandler("skip", nil, nil)

	// Step overrides to retry
	step := &Step{ID: "test-step", Name: "Test Step", OnError: "retry"}
	err := errors.New("test error")

	action := handler.Handle(step, err)

	if action != ActionRetry {
		t.Errorf("Expected ActionRetry (step override), got %v", action)
	}
}

func TestErrorHandler_ResetRetryCount(t *testing.T) {
	handler := NewErrorHandler("retry", nil, nil)
	step := &Step{ID: "test-step", Name: "Test Step"}
	err := errors.New("test error")

	// Trigger some retries
	handler.Handle(step, err)
	handler.Handle(step, err)

	if handler.GetRetryCount(step.ID) != 2 {
		t.Errorf("Expected retry count 2, got %d", handler.GetRetryCount(step.ID))
	}

	// Reset the count
	handler.ResetRetryCount(step.ID)

	if handler.GetRetryCount(step.ID) != 0 {
		t.Errorf("Expected retry count 0 after reset, got %d", handler.GetRetryCount(step.ID))
	}

	// Should be able to retry again
	action := handler.Handle(step, err)
	if action != ActionRetry {
		t.Errorf("Expected ActionRetry after reset, got %v", action)
	}
}

func TestErrorHandler_GetRetryCount(t *testing.T) {
	handler := NewErrorHandler("retry", nil, nil)

	// Initially should be 0 for unknown step
	if handler.GetRetryCount("unknown-step") != 0 {
		t.Errorf("Expected retry count 0 for unknown step")
	}

	step := &Step{ID: "test-step", Name: "Test Step"}
	err := errors.New("test error")

	handler.Handle(step, err)
	if handler.GetRetryCount(step.ID) != 1 {
		t.Errorf("Expected retry count 1, got %d", handler.GetRetryCount(step.ID))
	}

	handler.Handle(step, err)
	if handler.GetRetryCount(step.ID) != 2 {
		t.Errorf("Expected retry count 2, got %d", handler.GetRetryCount(step.ID))
	}
}

func TestErrorHandler_UnknownActionDefaultsToAbort(t *testing.T) {
	handler := NewErrorHandler("invalid", nil, nil)
	step := &Step{ID: "test-step", Name: "Test Step"}
	err := errors.New("test error")

	action := handler.Handle(step, err)

	if action != ActionAbort {
		t.Errorf("Expected ActionAbort for unknown action, got %v", action)
	}
}

func TestErrorHandler_EmptyDefaultActionIsAbort(t *testing.T) {
	handler := NewErrorHandler("", nil, nil)
	step := &Step{ID: "test-step", Name: "Test Step"}
	err := errors.New("test error")

	action := handler.Handle(step, err)

	if action != ActionAbort {
		t.Errorf("Expected ActionAbort for empty default, got %v", action)
	}
}

func TestErrorHandler_PauseWithNoDecisionsChannel(t *testing.T) {
	events := make(chan ExecutionEvent, 1)
	// No decisions channel
	handler := NewErrorHandler("pause", events, nil)
	step := &Step{ID: "test-step", Name: "Test Step"}
	err := errors.New("test error")

	action := handler.Handle(step, err)

	// Should fall back to abort
	if action != ActionAbort {
		t.Errorf("Expected ActionAbort when no decisions channel, got %v", action)
	}
}

func TestErrorHandler_MultipleStepsIndependentRetryCounts(t *testing.T) {
	handler := NewErrorHandler("retry", nil, nil)
	step1 := &Step{ID: "step-1", Name: "Step 1"}
	step2 := &Step{ID: "step-2", Name: "Step 2"}
	err := errors.New("test error")

	// Trigger retries on step1
	handler.Handle(step1, err)
	handler.Handle(step1, err)

	// Trigger retry on step2
	handler.Handle(step2, err)

	if handler.GetRetryCount(step1.ID) != 2 {
		t.Errorf("Expected retry count 2 for step1, got %d", handler.GetRetryCount(step1.ID))
	}
	if handler.GetRetryCount(step2.ID) != 1 {
		t.Errorf("Expected retry count 1 for step2, got %d", handler.GetRetryCount(step2.ID))
	}
}

func TestRetryDelay(t *testing.T) {
	testCases := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			delay := retryDelay(tc.attempt)
			if delay != tc.expected {
				t.Errorf("retryDelay(%d) = %v, expected %v", tc.attempt, delay, tc.expected)
			}
		})
	}
}

func TestRecoveryAction_String(t *testing.T) {
	testCases := []struct {
		action   RecoveryAction
		expected string
	}{
		{ActionAbort, "abort"},
		{ActionSkip, "skip"},
		{ActionRetry, "retry"},
		{ActionPause, "pause"},
		{RecoveryAction(99), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.action.String() != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, tc.action.String())
			}
		})
	}
}
