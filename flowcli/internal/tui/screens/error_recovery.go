package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/flowcli/internal/engine"
	"github.com/nameer-kp/flowcli/internal/tui/styles"
)

// RecoveryDecisionMsg is emitted when the user selects a recovery action
type RecoveryDecisionMsg struct {
	Action engine.RecoveryAction
}

// recoveryOption represents a selectable recovery option
type recoveryOption struct {
	action      engine.RecoveryAction
	key         string
	label       string
	description string
}

// ErrorRecoveryModel is the model for the error recovery screen
type ErrorRecoveryModel struct {
	step        *engine.Step
	err         error
	options     []recoveryOption
	selected    int
	showDetails bool
	width       int
	height      int
}

// NewErrorRecoveryModel creates a new error recovery screen
func NewErrorRecoveryModel(step *engine.Step, err error) ErrorRecoveryModel {
	options := []recoveryOption{
		{
			action:      engine.ActionRetry,
			key:         "R",
			label:       "Retry",
			description: "Attempt to re-execute this step",
		},
		{
			action:      engine.ActionSkip,
			key:         "S",
			label:       "Skip",
			description: "Skip this step and continue to the next",
		},
		{
			action:      engine.ActionAbort,
			key:         "A",
			label:       "Abort",
			description: "Stop the workflow execution",
		},
	}

	return ErrorRecoveryModel{
		step:        step,
		err:         err,
		options:     options,
		selected:    0,
		showDetails: false,
	}
}

// Init returns the initial command
func (m ErrorRecoveryModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ErrorRecoveryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case "down", "j":
			if m.selected < len(m.options)-1 {
				m.selected++
			}
			return m, nil

		case "enter":
			return m, m.selectAction(m.options[m.selected].action)

		case "r", "R":
			return m, m.selectAction(engine.ActionRetry)

		case "s", "S":
			return m, m.selectAction(engine.ActionSkip)

		case "a", "A":
			return m, m.selectAction(engine.ActionAbort)

		case "d", "D":
			m.showDetails = !m.showDetails
			return m, nil

		case "q", "ctrl+c":
			return m, m.selectAction(engine.ActionAbort)
		}
	}

	return m, nil
}

// selectAction returns a command that emits the recovery decision
func (m ErrorRecoveryModel) selectAction(action engine.RecoveryAction) tea.Cmd {
	return func() tea.Msg {
		return RecoveryDecisionMsg{Action: action}
	}
}

// View renders the screen
func (m ErrorRecoveryModel) View() string {
	// Header with error icon
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.StatusError.Bold(true).Render("Step Failed"),
		"",
	)

	// Step information
	stepName := "Unknown Step"
	if m.step != nil {
		if m.step.Name != "" {
			stepName = m.step.Name
		} else if m.step.ID != "" {
			stepName = m.step.ID
		}
	}

	stepInfo := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Subtitle.Render(fmt.Sprintf("Step: %s", stepName)),
		"",
	)

	// Error message (short version)
	errMsg := "Unknown error"
	if m.err != nil {
		errMsg = m.err.Error()
	}

	// Truncate error for short display if too long
	shortErr := errMsg
	if len(shortErr) > 80 && !m.showDetails {
		shortErr = shortErr[:77] + "..."
	}

	errorBox := styles.Box.
		BorderForeground(styles.Error).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			styles.StatusError.Render("Error:"),
			"",
			shortErr,
		))

	// Full details section (only shown when toggled)
	var detailsSection string
	if m.showDetails {
		detailsSection = lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			styles.Subtitle.Render("Full Error Details:"),
			"",
			m.wrapText(errMsg, m.getContentWidth()),
			"",
		)
	}

	// Options
	optionsHeader := styles.Subtitle.Bold(true).Render("Choose an action:")

	var optionLines []string
	for i, opt := range m.options {
		prefix := "  "
		style := styles.MenuItem
		if i == m.selected {
			prefix = "> "
			style = styles.MenuItemSelected
		}

		line := fmt.Sprintf("%s[%s] %s - %s", prefix, opt.key, opt.label, opt.description)
		optionLines = append(optionLines, style.Render(line))
	}

	optionsSection := lipgloss.JoinVertical(
		lipgloss.Left,
		optionsHeader,
		"",
		strings.Join(optionLines, "\n"),
	)

	// Help text
	detailToggle := "d show details"
	if m.showDetails {
		detailToggle = "d hide details"
	}
	help := styles.Help.Render(fmt.Sprintf("up/down or j/k navigate | enter select | r/s/a quick select | %s | q quit", detailToggle))

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		stepInfo,
		errorBox,
		detailsSection,
		"",
		optionsSection,
		"",
		help,
	)

	return styles.Container.Render(content)
}

// getContentWidth returns the available width for content
func (m ErrorRecoveryModel) getContentWidth() int {
	if m.width > 0 {
		return m.width - 8 // Account for padding and borders
	}
	return 72 // Default width
}

// wrapText wraps text to fit within the specified width
func (m ErrorRecoveryModel) wrapText(text string, width int) string {
	if width <= 0 {
		width = 72
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// SelectedAction returns the currently selected action
func (m ErrorRecoveryModel) SelectedAction() engine.RecoveryAction {
	return m.options[m.selected].action
}

// Step returns the step that caused the error
func (m ErrorRecoveryModel) Step() *engine.Step {
	return m.step
}

// Error returns the error that occurred
func (m ErrorRecoveryModel) Error() error {
	return m.err
}
