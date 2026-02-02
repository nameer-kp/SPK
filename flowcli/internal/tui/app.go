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
