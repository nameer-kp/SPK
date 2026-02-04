package node

import "log/slog"

// Node is the interface all node types must implement
type Node interface {
	// Name returns the node type name (e.g., "http", "db")
	Name() string

	// Description returns a human-readable description
	Description() string

	// ConfigSchema returns JSON Schema for config validation (optional)
	ConfigSchema() map[string]interface{}

	// Execute runs the node with the given config
	Execute(ctx Context, config map[string]interface{}) (Result, error)
}

// Context provides access to workflow state during execution
type Context interface {
	// Get retrieves a value from the shared context
	Get(key string) (interface{}, bool)

	// Set stores a value in the shared context
	Set(key string, value interface{})

	// GetInput retrieves a user input value
	GetInput(name string) (interface{}, bool)

	// GetProfileVar retrieves a profile variable
	GetProfileVar(name string) (string, bool)

	// GetEnv retrieves an environment variable
	GetEnv(name string) string

	// Logger returns a structured logger
	Logger() *slog.Logger
}

// Result holds the outcome of a node execution
type Result struct {
	Success bool
	Data    interface{}
	Logs    []LogEntry
	Status  string // Display status (e.g., "200 OK")
}

// LogEntry represents a log message from node execution
type LogEntry struct {
	Level   string // debug, info, warn, error
	Message string
	Fields  map[string]interface{}
}
