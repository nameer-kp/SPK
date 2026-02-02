package engine

import (
	"sync"
	"time"
)

// RecoveryAction represents the action to take when a step fails
type RecoveryAction int

const (
	// ActionAbort stops the workflow execution immediately
	ActionAbort RecoveryAction = iota
	// ActionSkip skips the failed step and continues to the next one
	ActionSkip
	// ActionRetry attempts to re-execute the failed step
	ActionRetry
	// ActionPause pauses execution and waits for user decision
	ActionPause
)

// String returns a human-readable name for the recovery action
func (a RecoveryAction) String() string {
	switch a {
	case ActionAbort:
		return "abort"
	case ActionSkip:
		return "skip"
	case ActionRetry:
		return "retry"
	case ActionPause:
		return "pause"
	default:
		return "unknown"
	}
}

// MaxRetries is the maximum number of retry attempts before giving up
const MaxRetries = 3

// ErrorHandler manages error recovery decisions during workflow execution
type ErrorHandler struct {
	defaultAction string                 // from workflow.OnError
	events        chan<- ExecutionEvent  // for pause to notify TUI
	decisions     <-chan RecoveryAction  // for pause to receive user decision
	retryCounts   map[string]int         // track retries per step
	mu            sync.Mutex             // protects retryCounts
}

// NewErrorHandler creates a new ErrorHandler with the specified default action.
// The events channel is used to emit events when pausing for user input.
// The decisions channel is used to receive user decisions during pause.
func NewErrorHandler(defaultAction string, events chan<- ExecutionEvent, decisions <-chan RecoveryAction) *ErrorHandler {
	return &ErrorHandler{
		defaultAction: defaultAction,
		events:        events,
		decisions:     decisions,
		retryCounts:   make(map[string]int),
	}
}

// Handle determines the recovery action for a failed step.
// It uses the step's OnError setting if specified, otherwise falls back to the workflow default.
func (h *ErrorHandler) Handle(step *Step, err error) RecoveryAction {
	// Determine which action to take
	action := h.defaultAction
	if step.OnError != "" {
		action = step.OnError
	}

	switch action {
	case "abort":
		return ActionAbort

	case "skip":
		return ActionSkip

	case "retry":
		h.mu.Lock()
		count := h.retryCounts[step.ID]
		if count >= MaxRetries {
			h.mu.Unlock()
			return ActionAbort
		}
		h.retryCounts[step.ID] = count + 1
		h.mu.Unlock()
		return ActionRetry

	case "pause":
		// Emit error event to notify TUI
		if h.events != nil {
			h.events <- ExecutionEvent{
				Type:      EventStepError,
				Step:      step,
				Error:     err,
				Message:   "Step failed, waiting for user decision",
				Timestamp: time.Now(),
			}
		}

		// Wait for user decision from decisions channel
		if h.decisions != nil {
			decision := <-h.decisions
			return decision
		}
		// If no decisions channel, fall back to abort
		return ActionAbort

	default:
		// Unknown action, default to abort for safety
		return ActionAbort
	}
}

// ResetRetryCount resets the retry counter for a specific step.
// This should be called when a step succeeds after retries.
func (h *ErrorHandler) ResetRetryCount(stepID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.retryCounts, stepID)
}

// GetRetryCount returns the current retry count for a specific step.
func (h *ErrorHandler) GetRetryCount(stepID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.retryCounts[stepID]
}

// retryDelay calculates the delay before a retry attempt using exponential backoff.
// Returns 1s, 2s, 4s for attempts 0, 1, 2 respectively.
func retryDelay(attempt int) time.Duration {
	return time.Duration(1<<attempt) * time.Second
}
