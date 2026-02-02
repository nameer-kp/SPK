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
