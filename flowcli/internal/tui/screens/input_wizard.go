package screens

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/flowcli/internal/engine"
	"github.com/nameer-kp/flowcli/internal/tui/styles"
)

// InputsCompleteMsg is emitted when all inputs have been collected
type InputsCompleteMsg struct {
	Inputs map[string]interface{}
}

// InputWizardModel handles collecting user inputs for a workflow
type InputWizardModel struct {
	workflow    *engine.Workflow
	inputs      []engine.Input
	textInputs  []textinput.Model
	focusIndex  int
	values      map[string]interface{}
	boolValues  map[int]bool   // Track bool toggles by input index
	choiceIndex map[int]int    // Track selected choice index by input index
	width       int
	height      int
	err         string
}

// NewInputWizardModel creates a new input wizard for the given workflow
func NewInputWizardModel(workflow *engine.Workflow) InputWizardModel {
	m := InputWizardModel{
		workflow:    workflow,
		inputs:      workflow.Inputs,
		textInputs:  make([]textinput.Model, len(workflow.Inputs)),
		focusIndex:  0,
		values:      make(map[string]interface{}),
		boolValues:  make(map[int]bool),
		choiceIndex: make(map[int]int),
	}

	// Initialize text inputs for each input field
	for i, input := range workflow.Inputs {
		ti := textinput.New()
		ti.Placeholder = input.Prompt
		ti.CharLimit = 256
		ti.Width = 40

		// Set default value if provided
		if input.Default != "" {
			ti.SetValue(input.Default)
		}

		// Configure based on input type
		switch input.Type {
		case "secret":
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '*'
		case "number":
			ti.Placeholder = fmt.Sprintf("%s (number)", input.Prompt)
		case "bool":
			// Bool uses toggle, not text input
			if input.Default == "true" {
				m.boolValues[i] = true
			} else {
				m.boolValues[i] = false
			}
		case "choice":
			// Choice uses selection, not text input
			m.choiceIndex[i] = 0
			// Find default in options
			for j, opt := range input.Options {
				if opt == input.Default {
					m.choiceIndex[i] = j
					break
				}
			}
		}

		if i == 0 {
			ti.Focus()
		}

		m.textInputs[i] = ti
	}

	return m
}

// Init returns the initial command
func (m InputWizardModel) Init() tea.Cmd {
	if len(m.textInputs) > 0 {
		return textinput.Blink
	}
	return nil
}

// Update handles messages
func (m InputWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Go back to workflow selection
			return m, tea.Quit

		case "tab", "down":
			// Move to next field
			m.focusIndex++
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}
			return m, m.updateFocus()

		case "shift+tab", "up":
			// Move to previous field
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}
			return m, m.updateFocus()

		case "enter":
			// If on submit button, validate and submit
			if m.focusIndex == len(m.inputs) {
				return m.submit()
			}
			// Otherwise move to next field
			m.focusIndex++
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}
			return m, m.updateFocus()

		case " ":
			// Space toggles bool and cycles choice
			if m.focusIndex < len(m.inputs) {
				input := m.inputs[m.focusIndex]
				switch input.Type {
				case "bool":
					m.boolValues[m.focusIndex] = !m.boolValues[m.focusIndex]
					return m, nil
				case "choice":
					if len(input.Options) > 0 {
						m.choiceIndex[m.focusIndex] = (m.choiceIndex[m.focusIndex] + 1) % len(input.Options)
					}
					return m, nil
				}
			}

		case "left":
			// For choice fields, go to previous option
			if m.focusIndex < len(m.inputs) {
				input := m.inputs[m.focusIndex]
				if input.Type == "choice" && len(input.Options) > 0 {
					m.choiceIndex[m.focusIndex]--
					if m.choiceIndex[m.focusIndex] < 0 {
						m.choiceIndex[m.focusIndex] = len(input.Options) - 1
					}
					return m, nil
				}
			}

		case "right":
			// For choice fields, go to next option
			if m.focusIndex < len(m.inputs) {
				input := m.inputs[m.focusIndex]
				if input.Type == "choice" && len(input.Options) > 0 {
					m.choiceIndex[m.focusIndex] = (m.choiceIndex[m.focusIndex] + 1) % len(input.Options)
					return m, nil
				}
			}
		}
	}

	// Update the focused text input
	if m.focusIndex < len(m.inputs) {
		input := m.inputs[m.focusIndex]
		if input.Type != "bool" && input.Type != "choice" {
			var cmd tea.Cmd
			m.textInputs[m.focusIndex], cmd = m.textInputs[m.focusIndex].Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// updateFocus updates which input has focus
func (m *InputWizardModel) updateFocus() tea.Cmd {
	var cmds []tea.Cmd

	for i := range m.textInputs {
		if i == m.focusIndex {
			cmds = append(cmds, m.textInputs[i].Focus())
		} else {
			m.textInputs[i].Blur()
		}
	}

	return tea.Batch(cmds...)
}

// submit validates and submits the form
func (m InputWizardModel) submit() (tea.Model, tea.Cmd) {
	m.err = ""
	values := make(map[string]interface{})

	for i, input := range m.inputs {
		var value interface{}

		switch input.Type {
		case "bool":
			value = m.boolValues[i]
		case "choice":
			if len(input.Options) > 0 {
				value = input.Options[m.choiceIndex[i]]
			} else {
				value = ""
			}
		case "number":
			strVal := m.textInputs[i].Value()
			if strVal == "" {
				if input.Required {
					m.err = fmt.Sprintf("'%s' is required", input.Name)
					return m, nil
				}
				value = 0
			} else {
				// Try parsing as int first, then float
				if intVal, err := strconv.ParseInt(strVal, 10, 64); err == nil {
					value = intVal
				} else if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
					value = floatVal
				} else {
					m.err = fmt.Sprintf("'%s' must be a valid number", input.Name)
					return m, nil
				}
			}
		default: // string, secret
			strVal := m.textInputs[i].Value()
			if strVal == "" && input.Required {
				m.err = fmt.Sprintf("'%s' is required", input.Name)
				return m, nil
			}
			value = strVal
		}

		values[input.Name] = value
	}

	return m, func() tea.Msg {
		return InputsCompleteMsg{Inputs: values}
	}
}

// View renders the input wizard
func (m InputWizardModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(styles.Title.Render(fmt.Sprintf("Configure: %s", m.workflow.Name)))
	b.WriteString("\n")
	if m.workflow.Description != "" {
		b.WriteString(styles.Subtitle.Render(m.workflow.Description))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Input fields
	for i, input := range m.inputs {
		focused := i == m.focusIndex

		// Label
		label := input.Name
		if input.Required {
			label += " *"
		}

		labelStyle := styles.Subtitle
		if focused {
			labelStyle = styles.Title
		}
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")

		// Input field
		switch input.Type {
		case "bool":
			b.WriteString(m.renderBoolInput(i, focused))
		case "choice":
			b.WriteString(m.renderChoiceInput(i, input, focused))
		default:
			b.WriteString(m.renderTextInput(i, focused))
		}
		b.WriteString("\n\n")
	}

	// Submit button
	submitFocused := m.focusIndex == len(m.inputs)
	submitStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted)

	if submitFocused {
		submitStyle = submitStyle.
			BorderForeground(styles.Primary).
			Foreground(styles.Primary).
			Bold(true)
	}
	b.WriteString(submitStyle.Render("Submit"))
	b.WriteString("\n")

	// Error message
	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(styles.StatusError.Render(m.err))
	}

	// Help text
	b.WriteString("\n")
	help := "tab/enter next | shift+tab prev"
	if m.focusIndex < len(m.inputs) {
		input := m.inputs[m.focusIndex]
		if input.Type == "bool" {
			help += " | space toggle"
		} else if input.Type == "choice" {
			help += " | space/arrows cycle"
		}
	}
	help += " | esc cancel"
	b.WriteString(styles.Help.Render(help))

	return styles.Container.Render(b.String())
}

// renderTextInput renders a text input field
func (m InputWizardModel) renderTextInput(index int, focused bool) string {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Padding(0, 1).
		Width(44)

	if focused {
		inputStyle = inputStyle.BorderForeground(styles.Primary)
	}

	return inputStyle.Render(m.textInputs[index].View())
}

// renderBoolInput renders a boolean toggle
func (m InputWizardModel) renderBoolInput(index int, focused bool) string {
	val := m.boolValues[index]

	var toggle string
	if val {
		toggle = lipgloss.NewStyle().
			Foreground(styles.Success).
			Bold(true).
			Render("[X] Yes") + "  [ ] No"
	} else {
		toggle = "[ ] Yes  " + lipgloss.NewStyle().
			Foreground(styles.Error).
			Bold(true).
			Render("[X] No")
	}

	toggleStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Padding(0, 1)

	if focused {
		toggleStyle = toggleStyle.BorderForeground(styles.Primary)
	}

	return toggleStyle.Render(toggle)
}

// renderChoiceInput renders a choice selector
func (m InputWizardModel) renderChoiceInput(index int, input engine.Input, focused bool) string {
	selectedIdx := m.choiceIndex[index]

	var options []string
	for i, opt := range input.Options {
		if i == selectedIdx {
			options = append(options, lipgloss.NewStyle().
				Foreground(styles.Primary).
				Bold(true).
				Render(fmt.Sprintf("(%d) %s", i+1, opt)))
		} else {
			options = append(options, lipgloss.NewStyle().
				Foreground(styles.Muted).
				Render(fmt.Sprintf("(%d) %s", i+1, opt)))
		}
	}

	choiceStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Padding(0, 1)

	if focused {
		choiceStyle = choiceStyle.BorderForeground(styles.Primary)
	}

	return choiceStyle.Render(strings.Join(options, "  "))
}

// GetValues returns the current input values
func (m InputWizardModel) GetValues() map[string]interface{} {
	return m.values
}

// GetWorkflow returns the workflow being configured
func (m InputWizardModel) GetWorkflow() *engine.Workflow {
	return m.workflow
}
