package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nameer-kp/atem/internal/config"
	"github.com/nameer-kp/atem/internal/tui/screens"
)

type Screen int

const (
	ScreenProjectSelect Screen = iota
	ScreenEnvSelect
	ScreenWelcome
	ScreenWorkflowSelect
	ScreenInputWizard
	ScreenExecution
	ScreenListSelect
	ScreenErrorRecovery
	ScreenResult
)

type App struct {
	screen              Screen
	projectSelectScreen screens.ProjectSelectScreen
	envSelectScreen     screens.EnvSelectScreen
	welcomeScreen       screens.WelcomeScreen

	// Current selections
	projects        []config.Project
	selectedProject *config.Project
	selectedEnv     string
	selectedEnvVars map[string]string

	version string
	width   int
	height  int
}

func NewApp(projects []config.Project, version string) App {
	app := App{
		projects: projects,
		version:  version,
	}

	if len(projects) == 0 {
		// No projects found, go to welcome with default
		app.screen = ScreenWelcome
		app.welcomeScreen = screens.NewWelcomeScreen("default", version)
	} else if len(projects) == 1 {
		// Single project, skip to env selection
		app.selectedProject = &projects[0]
		app.screen = ScreenEnvSelect
		app.envSelectScreen = screens.NewEnvSelectScreen(projects[0].Name, projects[0].Environments)
	} else {
		// Multiple projects, start with project selection
		app.screen = ScreenProjectSelect
		app.projectSelectScreen = screens.NewProjectSelectScreen(projects)
	}

	return app
}

func (a App) Init() tea.Cmd {
	switch a.screen {
	case ScreenProjectSelect:
		return a.projectSelectScreen.Init()
	case ScreenEnvSelect:
		return a.envSelectScreen.Init()
	case ScreenWelcome:
		return a.welcomeScreen.Init()
	}
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case screens.ProjectSelectedMsg:
		a.selectedProject = &msg.Project
		a.screen = ScreenEnvSelect
		a.envSelectScreen = screens.NewEnvSelectScreen(msg.Project.Name, msg.Project.Environments)
		return a, nil

	case screens.EnvSelectedMsg:
		a.selectedEnv = msg.Name
		a.selectedEnvVars = msg.Environment.Variables
		a.screen = ScreenWelcome
		a.welcomeScreen = screens.NewWelcomeScreen(
			a.selectedProject.Name+" / "+msg.Name,
			a.version,
		)
		return a, nil

	case screens.BackToProjectsMsg:
		a.screen = ScreenProjectSelect
		a.projectSelectScreen = screens.NewProjectSelectScreen(a.projects)
		return a, nil
	}

	// Delegate to current screen
	switch a.screen {
	case ScreenProjectSelect:
		updated, cmd := a.projectSelectScreen.Update(msg)
		a.projectSelectScreen = updated.(screens.ProjectSelectScreen)
		return a, cmd
	case ScreenEnvSelect:
		updated, cmd := a.envSelectScreen.Update(msg)
		a.envSelectScreen = updated.(screens.EnvSelectScreen)
		return a, cmd
	case ScreenWelcome:
		updated, cmd := a.welcomeScreen.Update(msg)
		a.welcomeScreen = updated.(screens.WelcomeScreen)
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	switch a.screen {
	case ScreenProjectSelect:
		return a.projectSelectScreen.View()
	case ScreenEnvSelect:
		return a.envSelectScreen.View()
	case ScreenWelcome:
		return a.welcomeScreen.View()
	default:
		return "Unknown screen"
	}
}

// SelectedProject returns the currently selected project
func (a App) SelectedProject() *config.Project {
	return a.selectedProject
}

// SelectedEnvVars returns the environment variables for the selected environment
func (a App) SelectedEnvVars() map[string]string {
	return a.selectedEnvVars
}
