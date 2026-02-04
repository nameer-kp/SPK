package screens

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/atem/internal/config"
	"github.com/nameer-kp/atem/internal/tui/styles"
)

type projectItem struct {
	project config.Project
}

func (i projectItem) Title() string       { return i.project.Name }
func (i projectItem) Description() string { return i.project.Description }
func (i projectItem) FilterValue() string { return i.project.Name }

// ProjectSelectedMsg is sent when a project is selected
type ProjectSelectedMsg struct {
	Project config.Project
}

type ProjectSelectScreen struct {
	list     list.Model
	projects []config.Project
}

func NewProjectSelectScreen(projects []config.Project) ProjectSelectScreen {
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = projectItem{project: p}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(styles.Primary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(styles.Muted)

	l := list.New(items, delegate, 50, 14)
	l.Title = "Select Project"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	return ProjectSelectScreen{
		list:     l,
		projects: projects,
	}
}

func (s ProjectSelectScreen) Init() tea.Cmd {
	return nil
}

func (s ProjectSelectScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "enter":
			if item, ok := s.list.SelectedItem().(projectItem); ok {
				return s, func() tea.Msg {
					return ProjectSelectedMsg{Project: item.project}
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

func (s ProjectSelectScreen) View() string {
	help := styles.Help.Render("↑/↓ navigate • / filter • enter select • q quit")

	return styles.Container.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			s.list.View(),
			help,
		),
	)
}
