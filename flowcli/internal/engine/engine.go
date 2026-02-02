package engine

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/nameer-kp/flowcli/internal/config"
	"github.com/nameer-kp/flowcli/pkg/node"
)

// Engine executes workflows
type Engine struct {
	registry *node.Registry
	config   *config.ResolvedConfig
	logger   *slog.Logger
}

// NewEngine creates a workflow execution engine
func NewEngine(registry *node.Registry, cfg *config.ResolvedConfig, logger *slog.Logger) *Engine {
	return &Engine{
		registry: registry,
		config:   cfg,
		logger:   logger,
	}
}

// ExecuteStep runs a single workflow step
func (e *Engine) ExecuteStep(step *Step, ctx *ExecutionContext, tmpl *TemplateEngine) (*StepResult, error) {
	start := time.Now()

	// Get the node implementation
	n, err := e.registry.Get(step.Type)
	if err != nil {
		return &StepResult{
			StepID:  step.ID,
			Success: false,
			Error:   err,
			Status:  "Unknown node type",
		}, err
	}

	// Resolve templates in config
	resolvedConfig, err := tmpl.ResolveMap(step.Config)
	if err != nil {
		return &StepResult{
			StepID:  step.ID,
			Success: false,
			Error:   err,
			Status:  "Template error",
		}, err
	}

	e.logger.Info("executing step", "id", step.ID, "type", step.Type, "name", step.Name)

	// Execute the node
	result, err := n.Execute(ctx, resolvedConfig)

	duration := time.Since(start).Seconds()

	stepResult := &StepResult{
		StepID:   step.ID,
		Success:  result.Success,
		Data:     result.Data,
		Error:    err,
		Duration: duration,
		Status:   result.Status,
	}

	// Store output in context
	if step.Output != "" && result.Data != nil {
		ctx.Set(step.Output, result.Data)
		tmpl.SetContext(step.Output, result.Data)
	}

	return stepResult, err
}

// Run executes a complete workflow.
// The events channel is optional - pass nil if you don't need event notifications.
func (e *Engine) Run(workflow *Workflow, inputs map[string]interface{}, events chan<- ExecutionEvent) (*WorkflowState, error) {
	ctx := NewExecutionContext(e.config, e.logger)
	ctx.SetInputs(inputs)

	tmpl := NewTemplateEngine(e.config)
	tmpl.SetInputs(inputs)

	// Create event emitter only if events channel is provided
	var emitter *EventEmitter
	if events != nil {
		emitter = NewEventEmitter(events)
	}

	state := &WorkflowState{
		Workflow:    workflow,
		CurrentStep: 0,
		Context:     ctx.Data(),
		Inputs:      inputs,
		Results:     make([]StepResult, 0, len(workflow.Steps)),
		Status:      "running",
	}

	for i, step := range workflow.Steps {
		state.CurrentStep = i

		// Evaluate condition if present
		if step.Condition != "" {
			evaluator := NewEvaluator(ctx)
			shouldRun, err := evaluator.Evaluate(step.Condition)
			if err != nil {
				return state, fmt.Errorf("condition evaluation failed for step %s: %w", step.ID, err)
			}
			if !shouldRun {
				// Emit skip event, continue to next step
				if emitter != nil {
					emitter.Emit(ExecutionEvent{
						Type:      EventStepSkipped,
						Step:      &step,
						Message:   fmt.Sprintf("Condition not met: %s", step.Condition),
						Timestamp: time.Now(),
					})
				}
				// Add a skipped result to track the step
				state.Results = append(state.Results, StepResult{
					StepID:  step.ID,
					Success: true,
					Status:  "skipped",
				})
				continue
			}
		}

		// Emit step start event
		if emitter != nil {
			emitter.Emit(ExecutionEvent{
				Type:      EventStepStart,
				Step:      &step,
				Timestamp: time.Now(),
			})
		}

		result, err := e.ExecuteStep(&step, ctx, tmpl)
		state.Results = append(state.Results, *result)

		if err != nil || !result.Success {
			// Emit step error event
			if emitter != nil {
				emitter.Emit(ExecutionEvent{
					Type:      EventStepError,
					Step:      &step,
					Result:    result,
					Error:     err,
					Timestamp: time.Now(),
				})
			}
			state.Status = "failed"
			return state, err
		}

		// Emit step complete event
		if emitter != nil {
			emitter.Emit(ExecutionEvent{
				Type:      EventStepComplete,
				Step:      &step,
				Result:    result,
				Timestamp: time.Now(),
			})
		}
	}

	state.Status = "completed"

	// Emit workflow complete event
	if emitter != nil {
		emitter.Emit(ExecutionEvent{
			Type:      EventWorkflowComplete,
			Message:   fmt.Sprintf("Workflow '%s' completed successfully", workflow.Name),
			Timestamp: time.Now(),
		})
	}

	return state, nil
}
