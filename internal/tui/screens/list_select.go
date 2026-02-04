package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/atem/internal/tui/styles"
)

type selectableItem struct {
	display string
	value   interface{}
	index   int
}

func (i selectableItem) Title() string       { return i.display }
func (i selectableItem) Description() string { return "" }
func (i selectableItem) FilterValue() string { return i.display }

// ListSelectionMsg is sent when an item is selected
type ListSelectionMsg struct {
	Selected      interface{}
	SelectedIndex int
}

// ListCancelledMsg is sent when selection is cancelled
type ListCancelledMsg struct{}

type ListSelectScreen struct {
	list   list.Model
	prompt string
	stepID string
}

func NewListSelectScreen(stepID, prompt string, items []interface{}, displayFn func(item interface{}) string) ListSelectScreen {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		display := displayFn(item)
		if display == "" {
			display = fmt.Sprintf("%v", item)
		}
		listItems[i] = selectableItem{
			display: display,
			value:   item,
			index:   i,
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(styles.Primary)
	delegate.ShowDescription = false

	l := list.New(listItems, delegate, 60, 15)
	l.Title = prompt
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	return ListSelectScreen{
		list:   l,
		prompt: prompt,
		stepID: stepID,
	}
}

func (s ListSelectScreen) Init() tea.Cmd {
	return nil
}

func (s ListSelectScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "esc":
			return s, func() tea.Msg {
				return ListCancelledMsg{}
			}
		case "enter":
			if item, ok := s.list.SelectedItem().(selectableItem); ok {
				return s, func() tea.Msg {
					return ListSelectionMsg{
						Selected:      item.value,
						SelectedIndex: item.index,
					}
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

func (s ListSelectScreen) View() string {
	help := styles.Help.Render("↑/↓ navigate • / filter • enter select • esc cancel • q quit")

	return styles.Container.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			s.list.View(),
			help,
		),
	)
}

// StepID returns the step that triggered this selection
func (s ListSelectScreen) StepID() string {
	return s.stepID
}
