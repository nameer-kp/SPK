# flowcli Phase 1 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the core foundation of flowcli - a working TUI that can load profiles, parse workflows, and execute HTTP nodes.

**Architecture:** Bubble Tea TUI with layered architecture: screens (UI) → engine (execution) → nodes (actions). Config loading uses priority chain (CLI > project > profile > env). Workflow YAML parsed into typed structs, executed sequentially with templated variable substitution.

**Tech Stack:** Go 1.21+, Bubble Tea (TUI), Lip Gloss (styling), Bubbles (components), gopkg.in/yaml.v3, expr-lang/expr (templating), godotenv.

---

## Task 1: Initialize Go Module

**Files:**
- Create: `flowcli/go.mod`
- Create: `flowcli/go.sum`
- Create: `flowcli/cmd/flowcli/main.go`

**Step 1: Initialize Go module**

```bash
cd /Users/nameer/Projects/SPK/flowcli
go mod init github.com/nameer-kp/flowcli
```

**Step 2: Create minimal main.go**

Create `flowcli/cmd/flowcli/main.go`:

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("flowcli v0.1.0")
	os.Exit(0)
}
```

**Step 3: Verify it compiles and runs**

Run: `go run ./cmd/flowcli`
Expected: `flowcli v0.1.0`

**Step 4: Commit**

```bash
git add flowcli/
git commit -m "feat(flowcli): initialize Go module with minimal main"
```

---

## Task 2: Add Bubble Tea Dependencies

**Files:**
- Modify: `flowcli/go.mod`

**Step 1: Add core dependencies**

```bash
cd /Users/nameer/Projects/SPK/flowcli
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles/list
go get github.com/charmbracelet/bubbles/textinput
go get github.com/charmbracelet/bubbles/spinner
go get gopkg.in/yaml.v3
go get github.com/expr-lang/expr
go get github.com/joho/godotenv
go get github.com/jmoiron/sqlx
```

**Step 2: Tidy dependencies**

```bash
go mod tidy
```

**Step 3: Verify dependencies installed**

Run: `go list -m all | grep bubbletea`
Expected: `github.com/charmbracelet/bubbletea v0.x.x`

**Step 4: Commit**

```bash
git add flowcli/go.mod flowcli/go.sum
git commit -m "feat(flowcli): add Bubble Tea and core dependencies"
```

---

## Task 3: Create Theme/Styles Package

**Files:**
- Create: `flowcli/internal/tui/styles/theme.go`

**Step 1: Create styles package**

Create `flowcli/internal/tui/styles/theme.go`:

```go
package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Primary   = lipgloss.Color("#7C3AED")
	Secondary = lipgloss.Color("#10B981")
	Muted     = lipgloss.Color("#6B7280")
	Error     = lipgloss.Color("#EF4444")
	Success   = lipgloss.Color("#10B981")
	Warning   = lipgloss.Color("#F59E0B")

	// Text styles
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary)

	Subtitle = lipgloss.NewStyle().
			Foreground(Muted)

	StatusSuccess = lipgloss.NewStyle().
			Foreground(Success)

	StatusError = lipgloss.NewStyle().
			Foreground(Error)

	StatusPending = lipgloss.NewStyle().
			Foreground(Muted)

	StatusRunning = lipgloss.NewStyle().
			Foreground(Warning)

	// Layout styles
	Container = lipgloss.NewStyle().
			Padding(1, 2)

	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Muted).
		Padding(1, 2)

	// Menu styles
	MenuItem = lipgloss.NewStyle().
			PaddingLeft(2)

	MenuItemSelected = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(Primary).
				Bold(true)

	// Help text
	Help = lipgloss.NewStyle().
		Foreground(Muted).
		MarginTop(1)
)
```

**Step 2: Verify it compiles**

Run: `go build ./internal/tui/styles`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/internal/
git commit -m "feat(flowcli): add TUI theme and styles"
```

---

## Task 4: Create Welcome Screen

**Files:**
- Create: `flowcli/internal/tui/screens/welcome.go`

**Step 1: Create welcome screen**

Create `flowcli/internal/tui/screens/welcome.go`:

```go
package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/flowcli/internal/tui/styles"
)

type menuItem struct {
	title       string
	description string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.description }
func (i menuItem) FilterValue() string { return i.title }

type WelcomeScreen struct {
	list    list.Model
	profile string
	version string
}

func NewWelcomeScreen(profile, version string) WelcomeScreen {
	items := []list.Item{
		menuItem{title: "Run Workflow", description: "Execute a workflow file"},
		menuItem{title: "Manage Profiles", description: "View and edit profiles"},
		menuItem{title: "View Recent Runs", description: "See workflow execution history"},
		menuItem{title: "Settings", description: "Configure flowcli"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(styles.Primary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(styles.Muted)

	l := list.New(items, delegate, 50, 12)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return WelcomeScreen{
		list:    l,
		profile: profile,
		version: version,
	}
}

func (s WelcomeScreen) Init() tea.Cmd {
	return nil
}

func (s WelcomeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "enter":
			// Will handle navigation in later tasks
			return s, nil
		}
	case tea.WindowSizeMsg:
		s.list.SetWidth(msg.Width - 4)
		s.list.SetHeight(msg.Height - 10)
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s WelcomeScreen) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render(fmt.Sprintf("flowcli %s", s.version)),
		"",
		styles.Subtitle.Render(fmt.Sprintf("Profile: %s", s.profile)),
		"",
	)

	help := styles.Help.Render("↑/↓ navigate • enter select • q quit")

	return styles.Container.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			s.list.View(),
			help,
		),
	)
}

// SelectedIndex returns the currently selected menu item index
func (s WelcomeScreen) SelectedIndex() int {
	return s.list.Index()
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/tui/screens`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/internal/tui/screens/
git commit -m "feat(flowcli): add welcome screen component"
```

---

## Task 5: Create Main App Model

**Files:**
- Create: `flowcli/internal/tui/app.go`

**Step 1: Create main app model**

Create `flowcli/internal/tui/app.go`:

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nameer-kp/flowcli/internal/tui/screens"
)

type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenWorkflowSelect
	ScreenInputWizard
	ScreenExecution
	ScreenErrorRecovery
	ScreenResult
)

type App struct {
	screen        Screen
	welcomeScreen screens.WelcomeScreen
	profile       string
	version       string
	width         int
	height        int
}

func NewApp(profile, version string) App {
	return App{
		screen:        ScreenWelcome,
		welcomeScreen: screens.NewWelcomeScreen(profile, version),
		profile:       profile,
		version:       version,
	}
}

func (a App) Init() tea.Cmd {
	return a.welcomeScreen.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	}

	switch a.screen {
	case ScreenWelcome:
		updated, cmd := a.welcomeScreen.Update(msg)
		a.welcomeScreen = updated.(screens.WelcomeScreen)
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	switch a.screen {
	case ScreenWelcome:
		return a.welcomeScreen.View()
	default:
		return "Unknown screen"
	}
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/tui`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/internal/tui/app.go
git commit -m "feat(flowcli): add main TUI app model"
```

---

## Task 6: Wire Up TUI to Main

**Files:**
- Modify: `flowcli/cmd/flowcli/main.go`

**Step 1: Update main.go to launch TUI**

Replace `flowcli/cmd/flowcli/main.go`:

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nameer-kp/flowcli/internal/tui"
)

const version = "0.1.0"

func main() {
	profile := os.Getenv("FLOWCLI_PROFILE")
	if profile == "" {
		profile = "default"
	}

	app := tui.NewApp(profile, version)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Run and verify TUI displays**

Run: `go run ./cmd/flowcli`
Expected: TUI appears with welcome screen, menu items, profile shows "default"
Press: `q` to quit

**Step 3: Commit**

```bash
git add flowcli/cmd/flowcli/main.go
git commit -m "feat(flowcli): wire TUI to main entry point"
```

---

## Task 7: Create Config Types

**Files:**
- Create: `flowcli/internal/config/types.go`

**Step 1: Create config types**

Create `flowcli/internal/config/types.go`:

```go
package config

// GlobalConfig represents ~/.flowcli/config.yaml
type GlobalConfig struct {
	DefaultProfile  string `yaml:"default_profile"`
	Editor          string `yaml:"editor"`
	LogLevel        string `yaml:"log_level"`
	PluginsDir      string `yaml:"plugins_dir"`
	CheckpointsDir  string `yaml:"checkpoints_dir"`
	WorkflowsDir    string `yaml:"workflows_dir"`
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

// ProjectConfig represents .flowcli.yaml in project root
type ProjectConfig struct {
	DefaultProfile string            `yaml:"default_profile"`
	WorkflowsDir   string            `yaml:"workflows_dir"`
	PluginsDir     string            `yaml:"plugins_dir"`
	Variables      map[string]string `yaml:"variables"`
}

// ResolvedConfig is the final merged configuration
type ResolvedConfig struct {
	Profile    Profile
	Variables  map[string]string // Merged from all sources
	WorkflowsDir string
	PluginsDir   string
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/config`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/internal/config/
git commit -m "feat(flowcli): add config and profile types"
```

---

## Task 8: Create Config Loader

**Files:**
- Create: `flowcli/internal/config/loader.go`

**Step 1: Create config loader**

Create `flowcli/internal/config/loader.go`:

```go
package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Loader handles loading configuration from multiple sources
type Loader struct {
	homeDir    string
	projectDir string
}

// NewLoader creates a config loader
func NewLoader() (*Loader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &Loader{
		homeDir:    home,
		projectDir: cwd,
	}, nil
}

// Load resolves configuration from all sources
func (l *Loader) Load(profileName string) (*ResolvedConfig, error) {
	// Load .env file first (lowest priority)
	envPath := filepath.Join(l.projectDir, ".env")
	_ = godotenv.Load(envPath) // Ignore error if not exists

	// Load global config
	globalConfig := l.loadGlobalConfig()

	// Determine which profile to use
	if profileName == "" {
		profileName = os.Getenv("FLOWCLI_PROFILE")
	}
	if profileName == "" {
		profileName = globalConfig.DefaultProfile
	}
	if profileName == "" {
		profileName = "default"
	}

	// Load the profile
	profile := l.loadProfile(profileName)

	// Load project config
	projectConfig := l.loadProjectConfig()

	// Merge variables (priority: project > profile > env)
	variables := make(map[string]string)

	// Start with profile variables
	for k, v := range profile.Variables {
		variables[k] = v
	}

	// Override with project variables
	for k, v := range projectConfig.Variables {
		variables[k] = v
	}

	// Resolve workflows and plugins directories
	workflowsDir := projectConfig.WorkflowsDir
	if workflowsDir == "" {
		workflowsDir = globalConfig.WorkflowsDir
	}
	if workflowsDir == "" {
		workflowsDir = filepath.Join(l.homeDir, ".flowcli", "workflows")
	}

	pluginsDir := projectConfig.PluginsDir
	if pluginsDir == "" {
		pluginsDir = globalConfig.PluginsDir
	}
	if pluginsDir == "" {
		pluginsDir = filepath.Join(l.homeDir, ".flowcli", "plugins")
	}

	return &ResolvedConfig{
		Profile:      profile,
		Variables:    variables,
		WorkflowsDir: workflowsDir,
		PluginsDir:   pluginsDir,
	}, nil
}

func (l *Loader) loadGlobalConfig() GlobalConfig {
	path := filepath.Join(l.homeDir, ".flowcli", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return GlobalConfig{}
	}

	var config GlobalConfig
	_ = yaml.Unmarshal(data, &config)
	return config
}

func (l *Loader) loadProfile(name string) Profile {
	// Try loading from ~/.flowcli/profiles/
	path := filepath.Join(l.homeDir, ".flowcli", "profiles", name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Profile{Name: name, Variables: make(map[string]string)}
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return Profile{Name: name, Variables: make(map[string]string)}
	}

	if profile.Variables == nil {
		profile.Variables = make(map[string]string)
	}

	return profile
}

func (l *Loader) loadProjectConfig() ProjectConfig {
	path := filepath.Join(l.projectDir, ".flowcli.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}
	}

	var config ProjectConfig
	_ = yaml.Unmarshal(data, &config)
	return config
}

// ListProfiles returns available profile names
func (l *Loader) ListProfiles() []string {
	profilesDir := filepath.Join(l.homeDir, ".flowcli", "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return []string{"default"}
	}

	var profiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".yaml" {
			name := e.Name()[:len(e.Name())-5] // Remove .yaml
			profiles = append(profiles, name)
		}
	}

	if len(profiles) == 0 {
		return []string{"default"}
	}
	return profiles
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/config`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/internal/config/loader.go
git commit -m "feat(flowcli): add multi-source config loader"
```

---

## Task 9: Create Workflow Types

**Files:**
- Create: `flowcli/internal/engine/types.go`

**Step 1: Create workflow types**

Create `flowcli/internal/engine/types.go`:

```go
package engine

// Workflow represents a parsed workflow YAML file
type Workflow struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Version     string  `yaml:"version"`
	Inputs      []Input `yaml:"inputs"`
	Steps       []Step  `yaml:"steps"`
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
	Workflow     *Workflow
	CurrentStep  int
	Context      map[string]interface{} // Shared data between steps
	Inputs       map[string]interface{} // User-provided inputs
	Results      []StepResult
	Status       string // pending, running, paused, completed, failed
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/engine`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/internal/engine/
git commit -m "feat(flowcli): add workflow and step types"
```

---

## Task 10: Create Workflow Parser

**Files:**
- Create: `flowcli/internal/engine/parser.go`

**Step 1: Create workflow parser**

Create `flowcli/internal/engine/parser.go`:

```go
package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Parser handles loading and parsing workflow YAML files
type Parser struct{}

// NewParser creates a new workflow parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile loads and parses a workflow from a file path
func (p *Parser) ParseFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses workflow YAML content
func (p *Parser) Parse(data []byte) (*Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	if err := p.validate(&workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// validate checks workflow structure for required fields
func (p *Parser) validate(w *Workflow) error {
	if w.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(w.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	stepIDs := make(map[string]bool)
	for i, step := range w.Steps {
		if step.ID == "" {
			return fmt.Errorf("step %d: id is required", i+1)
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("step %d: duplicate id '%s'", i+1, step.ID)
		}
		stepIDs[step.ID] = true

		if step.Type == "" {
			return fmt.Errorf("step '%s': type is required", step.ID)
		}

		if step.Config == nil {
			return fmt.Errorf("step '%s': config is required", step.ID)
		}
	}

	return nil
}

// ListWorkflows returns workflow files in a directory
func (p *Parser) ListWorkflows(dir string) ([]WorkflowInfo, error) {
	var workflows []WorkflowInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return workflows, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		w, err := p.ParseFile(path)
		if err != nil {
			continue // Skip invalid workflows
		}

		workflows = append(workflows, WorkflowInfo{
			Path:        path,
			Name:        w.Name,
			Description: w.Description,
			Version:     w.Version,
			StepCount:   len(w.Steps),
		})
	}

	return workflows, nil
}

// WorkflowInfo holds summary info for listing workflows
type WorkflowInfo struct {
	Path        string
	Name        string
	Description string
	Version     string
	StepCount   int
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/engine`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/internal/engine/parser.go
git commit -m "feat(flowcli): add workflow YAML parser"
```

---

## Task 11: Create Template Engine

**Files:**
- Create: `flowcli/internal/engine/template.go`

**Step 1: Create template engine**

Create `flowcli/internal/engine/template.go`:

```go
package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/nameer-kp/flowcli/internal/config"
)

// TemplateEngine handles {{variable}} substitution
type TemplateEngine struct {
	config  *config.ResolvedConfig
	context map[string]interface{}
	inputs  map[string]interface{}
}

// NewTemplateEngine creates a template engine with config context
func NewTemplateEngine(cfg *config.ResolvedConfig) *TemplateEngine {
	return &TemplateEngine{
		config:  cfg,
		context: make(map[string]interface{}),
		inputs:  make(map[string]interface{}),
	}
}

// SetContext updates the step output context
func (t *TemplateEngine) SetContext(key string, value interface{}) {
	t.context[key] = value
}

// SetInputs sets user-provided input values
func (t *TemplateEngine) SetInputs(inputs map[string]interface{}) {
	t.inputs = inputs
}

// pattern matches {{path.to.value}}
var templatePattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// Resolve replaces all {{}} templates in a string
func (t *TemplateEngine) Resolve(text string) (string, error) {
	result := templatePattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract path from {{path}}
		path := strings.TrimSpace(match[2 : len(match)-2])
		value, err := t.resolvePath(path)
		if err != nil {
			return match // Leave unresolved on error
		}
		return fmt.Sprintf("%v", value)
	})
	return result, nil
}

// ResolveMap resolves templates in a map recursively
func (t *TemplateEngine) ResolveMap(m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range m {
		resolved, err := t.resolveValue(v)
		if err != nil {
			return nil, err
		}
		result[k] = resolved
	}
	return result, nil
}

// ResolveValue resolves templates in any value type
func (t *TemplateEngine) resolveValue(v interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		return t.Resolve(val)
	case map[string]interface{}:
		return t.ResolveMap(val)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			resolved, err := t.resolveValue(item)
			if err != nil {
				return nil, err
			}
			result[i] = resolved
		}
		return result, nil
	default:
		return v, nil
	}
}

// resolvePath looks up a dotted path like "profile.base_url" or "step_id.field"
func (t *TemplateEngine) resolvePath(path string) (interface{}, error) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	root := parts[0]
	var rest string
	if len(parts) > 1 {
		rest = parts[1]
	}

	switch root {
	case "inputs":
		if rest == "" {
			return t.inputs, nil
		}
		return t.getNestedValue(t.inputs, rest)

	case "profile":
		if rest == "" {
			return t.config.Profile.Variables, nil
		}
		if val, ok := t.config.Profile.Variables[rest]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("profile variable '%s' not found", rest)

	case "env":
		if rest == "" {
			return nil, fmt.Errorf("env requires variable name")
		}
		return os.Getenv(rest), nil

	default:
		// Assume it's a step ID reference
		if stepData, ok := t.context[root]; ok {
			if rest == "" {
				return stepData, nil
			}
			if m, ok := stepData.(map[string]interface{}); ok {
				return t.getNestedValue(m, rest)
			}
			return stepData, nil
		}

		// Check if it's a direct variable reference
		if val, ok := t.config.Variables[root]; ok {
			return val, nil
		}

		return nil, fmt.Errorf("unknown reference '%s'", root)
	}
}

// getNestedValue traverses a map by dotted path
func (t *TemplateEngine) getNestedValue(m map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := interface{}(m)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("key '%s' not found", part)
			}
		default:
			return nil, fmt.Errorf("cannot traverse into non-map at '%s'", part)
		}
	}

	return current, nil
}

// ResolveJSON resolves templates and returns JSON bytes
func (t *TemplateEngine) ResolveJSON(v interface{}) ([]byte, error) {
	resolved, err := t.resolveValue(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(resolved)
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/engine`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/internal/engine/template.go
git commit -m "feat(flowcli): add template engine for variable substitution"
```

---

## Task 12: Create Node Interface

**Files:**
- Create: `flowcli/pkg/node/interface.go`

**Step 1: Create public node interface**

Create `flowcli/pkg/node/interface.go`:

```go
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
```

**Step 2: Verify it compiles**

Run: `go build ./pkg/node`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/pkg/
git commit -m "feat(flowcli): add public node interface for plugins"
```

---

## Task 13: Create HTTP Node

**Files:**
- Create: `flowcli/nodes/http.go`

**Step 1: Create HTTP node implementation**

Create `flowcli/nodes/http.go`:

```go
package nodes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nameer-kp/flowcli/pkg/node"
)

// HTTPNode executes HTTP requests
type HTTPNode struct{}

func NewHTTPNode() *HTTPNode {
	return &HTTPNode{}
}

func (n *HTTPNode) Name() string {
	return "http"
}

func (n *HTTPNode) Description() string {
	return "Execute HTTP requests"
}

func (n *HTTPNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"method":  "string",
		"url":     "string",
		"headers": "object",
		"body":    "any",
		"timeout": "string",
	}
}

func (n *HTTPNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	// Extract config values
	method, _ := config["method"].(string)
	if method == "" {
		method = "GET"
	}

	url, ok := config["url"].(string)
	if !ok || url == "" {
		return node.Result{}, fmt.Errorf("url is required")
	}

	// Build request body
	var bodyReader io.Reader
	if body := config["body"]; body != nil {
		var bodyBytes []byte
		var err error

		switch b := body.(type) {
		case string:
			bodyBytes = []byte(b)
		case map[string]interface{}, []interface{}:
			bodyBytes, err = json.Marshal(b)
			if err != nil {
				return node.Result{}, fmt.Errorf("failed to marshal body: %w", err)
			}
		default:
			bodyBytes, err = json.Marshal(b)
			if err != nil {
				return node.Result{}, fmt.Errorf("failed to marshal body: %w", err)
			}
		}

		bodyReader = bytes.NewReader(bodyBytes)
		ctx.Logger().Debug("request body", "body", string(bodyBytes))
	}

	// Create request
	req, err := http.NewRequest(strings.ToUpper(method), url, bodyReader)
	if err != nil {
		return node.Result{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}

	// Default content-type for POST/PUT with body
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set timeout
	timeout := 30 * time.Second
	if t, ok := config["timeout"].(string); ok {
		if parsed, err := time.ParseDuration(t); err == nil {
			timeout = parsed
		}
	}

	client := &http.Client{Timeout: timeout}

	// Execute request
	ctx.Logger().Info("executing HTTP request", "method", method, "url", url)
	resp, err := client.Do(req)
	if err != nil {
		return node.Result{
			Success: false,
			Status:  "Error",
			Logs: []node.LogEntry{
				{Level: "error", Message: err.Error()},
			},
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return node.Result{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response as JSON if possible
	var data interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		// If not JSON, return as string
		data = string(respBody)
	}

	status := fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return node.Result{
		Success: success,
		Data:    data,
		Status:  status,
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Response: %s", status)},
		},
	}, nil
}
```

**Step 2: Verify it compiles**

Run: `go build ./nodes`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/nodes/
git commit -m "feat(flowcli): add HTTP node implementation"
```

---

## Task 14: Create Node Registry

**Files:**
- Create: `flowcli/pkg/node/registry.go`

**Step 1: Create node registry**

Create `flowcli/pkg/node/registry.go`:

```go
package node

import (
	"fmt"
	"sync"
)

// Registry holds registered node types
type Registry struct {
	mu    sync.RWMutex
	nodes map[string]Node
}

// NewRegistry creates a new node registry
func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[string]Node),
	}
}

// Register adds a node type to the registry
func (r *Registry) Register(n Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := n.Name()
	if _, exists := r.nodes[name]; exists {
		return fmt.Errorf("node type '%s' already registered", name)
	}

	r.nodes[name] = n
	return nil
}

// Get retrieves a node by type name
func (r *Registry) Get(name string) (Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n, ok := r.nodes[name]
	if !ok {
		return nil, fmt.Errorf("unknown node type '%s'", name)
	}

	return n, nil
}

// List returns all registered node type names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.nodes))
	for name := range r.nodes {
		names = append(names, name)
	}
	return names
}
```

**Step 2: Verify it compiles**

Run: `go build ./pkg/node`
Expected: No errors

**Step 3: Commit**

```bash
git add flowcli/pkg/node/registry.go
git commit -m "feat(flowcli): add node registry"
```

---

## Task 15: Create Workflow Execution Engine

**Files:**
- Create: `flowcli/internal/engine/engine.go`
- Create: `flowcli/internal/engine/context.go`

**Step 1: Create execution context**

Create `flowcli/internal/engine/context.go`:

```go
package engine

import (
	"log/slog"
	"os"

	"github.com/nameer-kp/flowcli/internal/config"
	"github.com/nameer-kp/flowcli/pkg/node"
)

// ExecutionContext implements node.Context
type ExecutionContext struct {
	config  *config.ResolvedConfig
	data    map[string]interface{}
	inputs  map[string]interface{}
	logger  *slog.Logger
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(cfg *config.ResolvedConfig, logger *slog.Logger) *ExecutionContext {
	return &ExecutionContext{
		config:  cfg,
		data:    make(map[string]interface{}),
		inputs:  make(map[string]interface{}),
		logger:  logger,
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
```

**Step 2: Create execution engine**

Create `flowcli/internal/engine/engine.go`:

```go
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
```

**Step 3: Verify it compiles**

Run: `go build ./internal/engine`
Expected: No errors

**Step 4: Commit**

```bash
git add flowcli/internal/engine/context.go flowcli/internal/engine/engine.go
git commit -m "feat(flowcli): add workflow execution engine"
```

---

## Task 16: Create Example Workflow

**Files:**
- Create: `flowcli/examples/workflows/itravel-migration.yaml`

**Step 1: Create example workflow directory**

```bash
mkdir -p /Users/nameer/Projects/SPK/flowcli/examples/workflows
```

**Step 2: Create iTravel migration workflow**

Create `flowcli/examples/workflows/itravel-migration.yaml`:

```yaml
name: itravel-migration
description: Retrieve booking, insert data, and migrate
version: "1.0"

inputs:
  - name: pnr_number
    type: string
    prompt: "Enter PNR number"
    required: true

steps:
  - id: retrieve_booking
    name: Retrieve Booking
    type: http
    config:
      method: POST
      url: "{{profile.base_url}}/selling/api/public-booking/v1/rest/bkg/pnr/retrieve"
      headers:
        Content-Type: application/json
        Accept: application/json
        x-session-token: "{{profile.session_token}}"
        x-auth-channel: "ItravelNativeUI@RCG"
      body:
        SuperPnrNumber: "{{inputs.pnr_number}}"
    output: booking_data

  - id: insert_data
    name: Insert to Migration DB
    type: http
    config:
      method: POST
      url: "{{profile.dm_service_url}}/itravel/insertData"
      headers:
        Content-Type: application/json
      body: "{{booking_data}}"
    output: insert_result

  - id: migrate_booking
    name: Call Migrate Booking API
    type: http
    condition: "insert_result.success == true"
    config:
      method: POST
      url: "{{profile.itravel_local_url}}/iTravel/api/public-booking/v1/rest/bkg/pnr/migrateBooking"
      headers:
        Content-Type: application/json
        Accept: application/json
        x-session-token: "{{profile.session_token}}"
        x-auth-channel: "ItravelNativeUI@RCG"
      body: "{{insert_result.request_payload}}"
    output: migration_result
```

**Step 3: Commit**

```bash
git add flowcli/examples/
git commit -m "feat(flowcli): add iTravel migration example workflow"
```

---

## Task 17: Create Example Profile

**Files:**
- Create: `flowcli/examples/profiles/itravel-sandbox.yaml`

**Step 1: Create example profiles directory**

```bash
mkdir -p /Users/nameer/Projects/SPK/flowcli/examples/profiles
```

**Step 2: Create example profile**

Create `flowcli/examples/profiles/itravel-sandbox.yaml`:

```yaml
name: itravel-sandbox
description: iTravel Sandbox Environment

variables:
  base_url: https://sandbox.itravel.ibsplc.org/iTravel
  dm_service_url: http://localhost:8082
  itravel_local_url: http://localhost:8080

secrets:
  session_token:
    env: X_SESSION_TOKEN
```

**Step 3: Commit**

```bash
git add flowcli/examples/profiles/
git commit -m "feat(flowcli): add example iTravel sandbox profile"
```

---

## Task 18: Integration - Wire Everything Together

**Files:**
- Modify: `flowcli/cmd/flowcli/main.go`

**Step 1: Update main to initialize engine**

Replace `flowcli/cmd/flowcli/main.go`:

```go
package main

import (
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nameer-kp/flowcli/internal/config"
	"github.com/nameer-kp/flowcli/internal/engine"
	"github.com/nameer-kp/flowcli/internal/tui"
	"github.com/nameer-kp/flowcli/nodes"
	"github.com/nameer-kp/flowcli/pkg/node"
)

const version = "0.1.0"

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	loader, err := config.NewLoader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	profileName := os.Getenv("FLOWCLI_PROFILE")
	cfg, err := loader.Load(profileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize node registry
	registry := node.NewRegistry()
	registry.Register(nodes.NewHTTPNode())

	// Initialize workflow engine
	eng := engine.NewEngine(registry, cfg, logger)
	_ = eng // Will be used by TUI later

	// Initialize parser
	parser := engine.NewParser()
	_ = parser // Will be used by TUI later

	// Launch TUI
	app := tui.NewApp(cfg.Profile.Name, version)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Run and verify everything compiles**

Run: `go run ./cmd/flowcli`
Expected: TUI appears with welcome screen, can quit with `q`

**Step 3: Commit**

```bash
git add flowcli/cmd/flowcli/main.go
git commit -m "feat(flowcli): wire config, registry, and engine to main"
```

---

## Task 19: Final Build and Test

**Step 1: Build the binary**

```bash
cd /Users/nameer/Projects/SPK/flowcli
go build -o flowcli ./cmd/flowcli
```

**Step 2: Verify binary runs**

Run: `./flowcli`
Expected: TUI launches with welcome screen

**Step 3: Create ~/.flowcli directories for testing**

```bash
mkdir -p ~/.flowcli/profiles
mkdir -p ~/.flowcli/workflows
cp examples/profiles/itravel-sandbox.yaml ~/.flowcli/profiles/
cp examples/workflows/itravel-migration.yaml ~/.flowcli/workflows/
```

**Step 4: Test with profile**

```bash
FLOWCLI_PROFILE=itravel-sandbox ./flowcli
```

Expected: TUI shows "Profile: itravel-sandbox"

**Step 5: Add binary to gitignore and commit**

```bash
echo "flowcli/flowcli" >> ../.gitignore
git add ../.gitignore
git commit -m "chore: add flowcli binary to gitignore"
```

---

## Phase 1 Complete Checklist

- [ ] Go module initialized
- [ ] Bubble Tea dependencies installed
- [ ] Theme/styles package created
- [ ] Welcome screen component
- [ ] Main app model
- [ ] TUI wired to main
- [ ] Config types defined
- [ ] Config loader (multi-source)
- [ ] Workflow types defined
- [ ] Workflow YAML parser
- [ ] Template engine for {{}} substitution
- [ ] Node interface (public API)
- [ ] HTTP node implementation
- [ ] Node registry
- [ ] Workflow execution engine
- [ ] Example workflow (itravel-migration.yaml)
- [ ] Example profile (itravel-sandbox.yaml)
- [ ] Everything wired together in main
- [ ] Binary builds and runs
