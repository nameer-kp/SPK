package screens

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/atem/internal/config"
	"github.com/nameer-kp/atem/internal/tui/styles"
)

type envItem struct {
	name string
	env  config.Environment
}

func (i envItem) Title() string       { return i.name }
func (i envItem) Description() string { return "" }
func (i envItem) FilterValue() string { return i.name }

// EnvSelectedMsg is sent when an environment is selected
type EnvSelectedMsg struct {
	Name        string
	Environment config.Environment
}

type EnvSelectScreen struct {
	list        list.Model
	projectName string
}

func NewEnvSelectScreen(projectName string, environments map[string]config.Environment) EnvSelectScreen {
	items := make([]list.Item, 0, len(environments))
	for name, env := range environments {
		items = append(items, envItem{name: name, env: env})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(styles.Primary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(styles.Muted)
	delegate.ShowDescription = false

	l := list.New(items, delegate, 50, 10)
	l.Title = "Select Environment"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return EnvSelectScreen{
		list:        l,
		projectName: projectName,
	}
}

func (s EnvSelectScreen) Init() tea.Cmd {
	return nil
}

func (s EnvSelectScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "esc":
			// Go back to project selection
			return s, func() tea.Msg {
				return BackToProjectsMsg{}
			}
		case "enter":
			if item, ok := s.list.SelectedItem().(envItem); ok {
				return s, func() tea.Msg {
					return EnvSelectedMsg{Name: item.name, Environment: item.env}
				}
			}
		}
	case tea.WindowSizeMsg:
		s.list.SetWidth(msg.Width - 4)
		s.list.SetHeight(msg.Height - 6)
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s EnvSelectScreen) View() string {
	header := styles.Subtitle.Render("Project: " + s.projectName)
	help := styles.Help.Render("↑/↓ navigate • enter select • esc back • q quit")

	return styles.Container.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			s.list.View(),
			help,
		),
	)
}

// BackToProjectsMsg signals navigation back to project selection
type BackToProjectsMsg struct{}
