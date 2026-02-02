package engine

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
)

// Evaluator evaluates condition expressions against an ExecutionContext
type Evaluator struct {
	ctx *ExecutionContext
}

// NewEvaluator creates a new condition evaluator with the given execution context
func NewEvaluator(ctx *ExecutionContext) *Evaluator {
	return &Evaluator{ctx: ctx}
}

// Evaluate evaluates a condition expression and returns a boolean result.
// Empty conditions return true.
// Returns an error if the expression is invalid or doesn't evaluate to a boolean.
func (e *Evaluator) Evaluate(condition string) (bool, error) {
	// Empty condition always returns true
	if strings.TrimSpace(condition) == "" {
		return true, nil
	}

	// Build the environment map for the expression
	env := e.buildEnvironment(condition)

	// Compile the expression without AsBool() first to check syntax
	// Then check if the result is boolean
	program, err := expr.Compile(condition, expr.Env(env), expr.AllowUndefinedVariables())
	if err != nil {
		return false, fmt.Errorf("invalid expression: %w", err)
	}

	// Run the expression
	output, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("expression evaluation failed: %w", err)
	}

	// Check if result is boolean
	result, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("non-boolean result: expression returned %T, expected bool", output)
	}

	return result, nil
}

// envVarPattern matches env.VAR_NAME patterns in expressions
var envVarPattern = regexp.MustCompile(`env\.([A-Za-z_][A-Za-z0-9_]*)`)

// extractEnvVarNames extracts all environment variable names referenced in a condition
func extractEnvVarNames(condition string) []string {
	matches := envVarPattern.FindAllStringSubmatch(condition, -1)
	names := make([]string, 0, len(matches))
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			names = append(names, match[1])
			seen[match[1]] = true
		}
	}
	return names
}

// buildEnvironment creates the environment map for expr evaluation
// It provides access to:
// - Step outputs: step_id.field or step_id.data.nested
// - Inputs namespace: inputs.var_name
// - Profile namespace: profile.var_name
// - Env namespace: env.VAR_NAME
func (e *Evaluator) buildEnvironment(condition string) map[string]interface{} {
	env := make(map[string]interface{})

	// Add all step outputs from the context data
	for key, value := range e.ctx.Data() {
		env[key] = value
	}

	// Add inputs namespace
	inputs := make(map[string]interface{})
	// Use a type assertion to access internal inputs
	// The ExecutionContext stores inputs in the inputs field
	if e.ctx != nil {
		// Copy all inputs
		for k, v := range e.ctx.inputs {
			inputs[k] = v
		}
	}
	env["inputs"] = inputs

	// Add profile namespace
	profile := make(map[string]interface{})
	if e.ctx != nil && e.ctx.config != nil {
		for k, v := range e.ctx.config.Profile.Variables {
			profile[k] = v
		}
	}
	env["profile"] = profile

	// Add env namespace - extract referenced env vars from condition and populate map
	envVars := make(map[string]interface{})
	for _, name := range extractEnvVarNames(condition) {
		envVars[name] = os.Getenv(name)
	}
	env["env"] = envVars

	return env
}
