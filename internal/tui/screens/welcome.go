package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/atem/internal/tui/styles"
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
		menuItem{title: "Settings", description: "Configure atem"},
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
		styles.Title.Render(fmt.Sprintf("atem %s", s.version)),
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
