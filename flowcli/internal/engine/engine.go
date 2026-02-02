package engine

import (
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

// Run executes a complete workflow
func (e *Engine) Run(workflow *Workflow, inputs map[string]interface{}) (*WorkflowState, error) {
	ctx := NewExecutionContext(e.config, e.logger)
	ctx.SetInputs(inputs)

	tmpl := NewTemplateEngine(e.config)
	tmpl.SetInputs(inputs)

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

		// TODO: Evaluate condition

		result, err := e.ExecuteStep(&step, ctx, tmpl)
		state.Results = append(state.Results, *result)

		if err != nil || !result.Success {
			state.Status = "failed"
			return state, err
		}
	}

	state.Status = "completed"
	return state, nil
}
