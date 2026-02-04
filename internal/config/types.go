package config

// GlobalConfig represents ~/.atem/config.yaml
type GlobalConfig struct {
	DefaultProfile string `yaml:"default_profile"`
	Editor         string `yaml:"editor"`
	LogLevel       string `yaml:"log_level"`
	PluginsDir     string `yaml:"plugins_dir"`
	CheckpointsDir string `yaml:"checkpoints_dir"`
	WorkflowsDir   string `yaml:"workflows_dir"`
}

// Profile represents a named environment configuration
type Profile struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Variables   map[string]string `yaml:"variables"`
	Secrets     map[string]Secret `yaml:"secrets"`
	Database    *DatabaseConfig   `yaml:"database,omitempty"`
}

// Secret defines how to obtain a secret value
type Secret struct {
	Env    string `yaml:"env,omitempty"`    // Read from environment variable
	Prompt string `yaml:"prompt,omitempty"` // Prompt user for input
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// ProjectConfig represents .atem.yaml in project root
type ProjectConfig struct {
	DefaultProfile string            `yaml:"default_profile"`
	WorkflowsDir   string            `yaml:"workflows_dir"`
	PluginsDir     string            `yaml:"plugins_dir"`
	Variables      map[string]string `yaml:"variables"`
}

// ResolvedConfig is the final merged configuration
type ResolvedConfig struct {
	Profile      Profile
	Variables    map[string]string // Merged from all sources
	WorkflowsDir string
	PluginsDir   string
}

// Project represents a discovered project with environments
type Project struct {
	Name          string                 `yaml:"name"`
	Description   string                 `yaml:"description"`
	Path          string                 `yaml:"-"` // Set by loader, not from YAML
	Environments  map[string]Environment `yaml:"environments"`
	WorkflowsDir  string                 `yaml:"workflows_dir"`
	Tools         ToolsConfig            `yaml:"tools"`
	RequiredTools []string               `yaml:"required_tools"`
}

// Environment holds env-specific variables
type Environment struct {
	Variables map[string]string `yaml:"variables,inline"`
}

// ToolsConfig holds configurable tool paths
type ToolsConfig struct {
	ImageViewer string `yaml:"image_viewer"`
}

// DiscoveredProjects holds all found projects
type DiscoveredProjects struct {
	Projects []Project
}
