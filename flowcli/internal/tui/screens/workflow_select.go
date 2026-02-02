package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/flowcli/internal/engine"
	"github.com/nameer-kp/flowcli/internal/tui/styles"
)

// WorkflowSelectedMsg is emitted when a workflow is selected
type WorkflowSelectedMsg struct {
	Path     string
	Workflow *engine.Workflow
}

// workflowsLoadedMsg is emitted when workflows are loaded
type workflowsLoadedMsg struct {
	items []workflowItem
	err   error
}

// workflowItem represents a workflow in the list
type workflowItem struct {
	path        string
	workflow    *engine.Workflow
	name        string
	description string
	stepCount   int
}

func (i workflowItem) Title() string       { return i.name }
func (i workflowItem) Description() string { return i.description }
func (i workflowItem) FilterValue() string { return i.name }

// WorkflowSelectModel is the model for the workflow selection screen
type WorkflowSelectModel struct {
	list         list.Model
	parser       *engine.Parser
	workflowsDir string
	workflows    []workflowItem
	width        int
	height       int
	err          error
}

// NewWorkflowSelectModel creates a new workflow selection screen
func NewWorkflowSelectModel(parser *engine.Parser, workflowsDir string) WorkflowSelectModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(styles.Primary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(styles.Muted)

	l := list.New([]list.Item{}, delegate, 50, 12)
	l.Title = "Select Workflow"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	return WorkflowSelectModel{
		list:         l,
		parser:       parser,
		workflowsDir: workflowsDir,
	}
}

// Init returns a command to load workflows
func (m WorkflowSelectModel) Init() tea.Cmd {
	return m.loadWorkflows()
}

// loadWorkflows returns a command that loads workflows from the directory
func (m WorkflowSelectModel) loadWorkflows() tea.Cmd {
	return func() tea.Msg {
		infos, err := m.parser.ListWorkflows(m.workflowsDir)
		if err != nil {
			return workflowsLoadedMsg{err: err}
		}

		items := make([]workflowItem, 0, len(infos))
		for _, info := range infos {
			// Parse the full workflow to include in the item
			workflow, err := m.parser.ParseFile(info.Path)
			if err != nil {
				continue // Skip invalid workflows
			}

			desc := info.Description
			if desc == "" {
				desc = fmt.Sprintf("%d steps", info.StepCount)
			} else {
				desc = fmt.Sprintf("%s (%d steps)", desc, info.StepCount)
			}

			items = append(items, workflowItem{
				path:        info.Path,
				workflow:    workflow,
				name:        info.Name,
				description: desc,
				stepCount:   info.StepCount,
			})
		}

		return workflowsLoadedMsg{items: items}
	}
}

// Update handles messages
func (m WorkflowSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 10)
		return m, nil

	case workflowsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.workflows = msg.items
		items := make([]list.Item, len(msg.items))
		for i, item := range msg.items {
			items[i] = item
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		// Don't handle quit keys when filtering
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(workflowItem); ok {
				return m, func() tea.Msg {
					return WorkflowSelectedMsg{
						Path:     item.path,
						Workflow: item.workflow,
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the screen
func (m WorkflowSelectModel) View() string {
	if m.err != nil {
		return styles.Container.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				styles.Title.Render("Workflow Selection"),
				"",
				styles.StatusError.Render(fmt.Sprintf("Error: %v", m.err)),
				"",
				styles.Help.Render("q quit"),
			),
		)
	}

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Workflow Selection"),
		"",
		styles.Subtitle.Render(fmt.Sprintf("Directory: %s", m.workflowsDir)),
		"",
	)

	help := styles.Help.Render("↑/↓ navigate • / filter • enter select • q/esc quit")

	return styles.Container.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			m.list.View(),
			help,
		),
	)
}

// SelectedWorkflow returns the currently selected workflow, if any
func (m WorkflowSelectModel) SelectedWorkflow() (*engine.Workflow, string) {
	if item, ok := m.list.SelectedItem().(workflowItem); ok {
		return item.workflow, item.path
	}
	return nil, ""
}
