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
	registry.Register(nodes.NewDelayNode())
	registry.Register(nodes.NewFileNode())
	registry.Register(nodes.NewShellNode())
	registry.Register(nodes.NewDBNode())
	registry.Register(nodes.NewTransformNode())
	registry.Register(nodes.NewLoopNode(registry))
	registry.Register(nodes.NewParallelNode(registry))

	// Initialize workflow engine
	eng := engine.NewEngine(registry, cfg, logger)

	// Initialize parser
	parser := engine.NewParser()

	// Launch TUI
	app := tui.NewApp(eng, parser, cfg, version)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
