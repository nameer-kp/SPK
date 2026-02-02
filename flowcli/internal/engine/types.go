package engine

// Workflow represents a parsed workflow YAML file
type Workflow struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Version     string  `yaml:"version"`
	Inputs      []Input `yaml:"inputs"`
	Steps       []Step  `yaml:"steps"`
	OnError     string  `yaml:"on_error,omitempty"` // abort, skip, pause - default error handling for all steps
}

// Input defines a user prompt at workflow start
type Input struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"` // string, number, bool, choice, secret
	Prompt   string   `yaml:"prompt"`
	Required bool     `yaml:"required"`
	Default  string   `yaml:"default,omitempty"`
	Options  []string `yaml:"options,omitempty"` // For choice type
}

// Step represents a single workflow step
type Step struct {
	ID        string                 `yaml:"id"`
	Name      string                 `yaml:"name"`
	Type      string                 `yaml:"type"` // http, db, shell, file, transform, delay, parallel, loop
	Condition string                 `yaml:"condition,omitempty"`
	Config    map[string]interface{} `yaml:"config"`
	Output    string                 `yaml:"output,omitempty"`
	OnError   string                 `yaml:"on_error,omitempty"` // retry, skip, abort
}

// HTTPConfig for http node type
type HTTPConfig struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    interface{}       `yaml:"body,omitempty"`
	Timeout string            `yaml:"timeout,omitempty"`
	Retry   *RetryConfig      `yaml:"retry,omitempty"`
}

// RetryConfig for retry behavior
type RetryConfig struct {
	Attempts int    `yaml:"attempts"`
	Delay    string `yaml:"delay"`
}

// StepResult holds the outcome of executing a step
type StepResult struct {
	StepID   string
	Success  bool
	Data     interface{}
	Error    error
	Duration float64 // seconds
	Status   string  // e.g., "200 OK" for HTTP
}

// WorkflowState holds the current execution state
type WorkflowState struct {
	Workflow    *Workflow
	CurrentStep int
	Context     map[string]interface{} // Shared data between steps
	Inputs      map[string]interface{} // User-provided inputs
	Results     []StepResult
	Status      string // pending, running, paused, completed, failed
}
