package main

import (
	"fmt"
	"os"

	// TUI framework
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/bubbles/list"
	_ "github.com/charmbracelet/bubbles/spinner"
	_ "github.com/charmbracelet/bubbles/textinput"
	_ "github.com/charmbracelet/lipgloss"

	// Utilities
	_ "github.com/expr-lang/expr"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv"
	_ "gopkg.in/yaml.v3"
)

func main() {
	fmt.Println("flowcli v0.1.0")
	os.Exit(0)
}
