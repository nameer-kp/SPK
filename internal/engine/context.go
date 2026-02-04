package engine

import (
	"log/slog"
	"os"

	"github.com/nameer-kp/atem/internal/config"
	"github.com/nameer-kp/atem/pkg/node"
)

// ExecutionContext implements node.Context
type ExecutionContext struct {
	config *config.ResolvedConfig
	data   map[string]interface{}
	inputs map[string]interface{}
	logger *slog.Logger
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(cfg *config.ResolvedConfig, logger *slog.Logger) *ExecutionContext {
	return &ExecutionContext{
		config: cfg,
		data:   make(map[string]interface{}),
		inputs: make(map[string]interface{}),
		logger: logger,
	}
}

func (c *ExecutionContext) Get(key string) (interface{}, bool) {
	v, ok := c.data[key]
	return v, ok
}

func (c *ExecutionContext) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *ExecutionContext) GetInput(name string) (interface{}, bool) {
	v, ok := c.inputs[name]
	return v, ok
}

func (c *ExecutionContext) SetInputs(inputs map[string]interface{}) {
	c.inputs = inputs
}

func (c *ExecutionContext) GetProfileVar(name string) (string, bool) {
	v, ok := c.config.Profile.Variables[name]
	return v, ok
}

func (c *ExecutionContext) GetEnv(name string) string {
	return os.Getenv(name)
}

func (c *ExecutionContext) Logger() *slog.Logger {
	return c.logger
}

func (c *ExecutionContext) Data() map[string]interface{} {
	return c.data
}

// Ensure ExecutionContext implements node.Context
var _ node.Context = (*ExecutionContext)(nil)
