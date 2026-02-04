package main

import (
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nameer-kp/atem/internal/config"
	"github.com/nameer-kp/atem/internal/tui"
	"github.com/nameer-kp/atem/nodes"
	"github.com/nameer-kp/atem/pkg/node"
)

// Build-time variables (injected via ldflags)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Handle --version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("atem %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration and discover projects
	loader, err := config.NewLoader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	// Discover projects from FLOW_PROJECT_DIRS
	discovered, err := loader.DiscoverProjects()
	if err != nil {
		logger.Warn("Failed to discover projects", "error", err)
	}

	// Initialize node registry
	registry := node.NewRegistry()

	// Core I/O nodes
	registry.Register(nodes.NewHTTPNode())
	registry.Register(nodes.NewFileNode())
	registry.Register(nodes.NewShellNode())
	registry.Register(nodes.NewCommandNode())
	registry.Register(nodes.NewDBNode())
	registry.Register(nodes.NewInputNode())

	// Control flow nodes
	registry.Register(nodes.NewDelayNode())
	registry.Register(nodes.NewLoopNode())
	registry.Register(nodes.NewConditionNode())
	registry.Register(nodes.NewSwitchNode())
	registry.Register(nodes.NewParallelNode())

	// Data manipulation nodes
	registry.Register(nodes.NewTransformNode())
	registry.Register(nodes.NewVariableNode())
	registry.Register(nodes.NewJSONNode())
	registry.Register(nodes.NewTemplateNode())
	registry.Register(nodes.NewArrayNode())
	registry.Register(nodes.NewStringNode())

	// Computing primitives
	registry.Register(nodes.NewMathNode())
	registry.Register(nodes.NewBitwiseNode())
	registry.Register(nodes.NewMemoryNode())

	// Utility nodes
	registry.Register(nodes.NewLogNode())
	registry.Register(nodes.NewEnvNode())
	registry.Register(nodes.NewCacheNode())
	registry.Register(nodes.NewCryptoNode())
	registry.Register(nodes.NewAssertNode())

	_ = registry // Will be passed to TUI/engine later

	// Launch TUI with discovered projects
	var projects []config.Project
	if discovered != nil {
		projects = discovered.Projects
	}

	app := tui.NewApp(projects, version)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
