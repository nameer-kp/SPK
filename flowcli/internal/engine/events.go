package engine

import "time"

// EventType represents the type of execution event
type EventType int

const (
	// EventStepStart is emitted when a step begins execution
	EventStepStart EventType = iota
	// EventStepComplete is emitted when a step finishes successfully
	EventStepComplete
	// EventStepError is emitted when a step fails with an error
	EventStepError
	// EventStepSkipped is emitted when a step is skipped (e.g., condition not met)
	EventStepSkipped
	// EventWorkflowComplete is emitted when the entire workflow finishes
	EventWorkflowComplete
	// EventPaused is emitted when execution is paused
	EventPaused
	// EventResumed is emitted when execution resumes after being paused
	EventResumed
)

// String returns a human-readable name for the event type
func (e EventType) String() string {
	switch e {
	case EventStepStart:
		return "StepStart"
	case EventStepComplete:
		return "StepComplete"
	case EventStepError:
		return "StepError"
	case EventStepSkipped:
		return "StepSkipped"
	case EventWorkflowComplete:
		return "WorkflowComplete"
	case EventPaused:
		return "Paused"
	case EventResumed:
		return "Resumed"
	default:
		return "Unknown"
	}
}

// ExecutionEvent represents an event that occurs during workflow execution.
// Events are sent from the engine to the TUI for real-time updates.
type ExecutionEvent struct {
	Type      EventType   // The type of event
	Step      *Step       // The step involved (nil for workflow events)
	Result    *StepResult // For complete events
	Error     error       // For error events
	Message   string      // Optional message
	Timestamp time.Time   // When the event occurred
}

// EventEmitter sends execution events to a channel for consumption by the TUI
type EventEmitter struct {
	ch chan<- ExecutionEvent
}

// NewEventEmitter creates a new EventEmitter that sends events to the provided channel
func NewEventEmitter(ch chan<- ExecutionEvent) *EventEmitter {
	return &EventEmitter{ch: ch}
}

// Emit sends an event to the channel in a non-blocking manner.
// If the channel is full or nil, the event is dropped silently.
func (e *EventEmitter) Emit(event ExecutionEvent) {
	if e.ch == nil {
		return
	}
	select {
	case e.ch <- event:
		// Event sent successfully
	default:
		// Channel is full, drop the event to avoid blocking
	}
}
