package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/flowcli/internal/engine"
	"github.com/nameer-kp/flowcli/internal/tui/styles"
)

// ExecutionCompleteMsg is emitted when workflow execution finishes
type ExecutionCompleteMsg struct {
	State *engine.WorkflowState
	Error error
}

// executionEventMsg wraps engine events for the TUI
type executionEventMsg struct {
	event engine.ExecutionEvent
}

// StepStatus represents the current status of a step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepCompleted
	StepFailed
	StepSkipped
)

// stepInfo tracks the state of each step during execution
type stepInfo struct {
	step      *engine.Step
	status    StepStatus
	startTime time.Time
	duration  time.Duration
	error     error
	logs      []string
}

// ExecutionModel is the model for the execution progress screen
type ExecutionModel struct {
	workflow *engine.Workflow
	engine   *engine.Engine
	inputs   map[string]interface{}
	events   chan engine.ExecutionEvent

	steps       []stepInfo
	currentStep int
	spinner     spinner.Model
	width       int
	height      int
	paused      bool
	aborted     bool
	completed   bool
	finalState  *engine.WorkflowState
	finalError  error
}

// NewExecutionModel creates a new execution progress screen
func NewExecutionModel(workflow *engine.Workflow, eng *engine.Engine, inputs map[string]interface{}) ExecutionModel {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Warning)

	// Initialize step tracking
	steps := make([]stepInfo, len(workflow.Steps))
	for i := range workflow.Steps {
		steps[i] = stepInfo{
			step:   &workflow.Steps[i],
			status: StepPending,
			logs:   make([]string, 0),
		}
	}

	// Create events channel with buffer
	events := make(chan engine.ExecutionEvent, 100)

	return ExecutionModel{
		workflow:    workflow,
		engine:      eng,
		inputs:      inputs,
		events:      events,
		steps:       steps,
		currentStep: -1,
		spinner:     s,
	}
}

// Init starts the workflow execution and event listening
func (m ExecutionModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.listenForEvents(),
		m.startExecution(),
	)
}

// startExecution runs the workflow in a goroutine
func (m ExecutionModel) startExecution() tea.Cmd {
	return func() tea.Msg {
		state, err := m.engine.Run(m.workflow, m.inputs, m.events)
		close(m.events)
		return ExecutionCompleteMsg{State: state, Error: err}
	}
}

// listenForEvents returns a command that listens for execution events
func (m ExecutionModel) listenForEvents() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-m.events
		if !ok {
			return nil
		}
		return executionEventMsg{event: event}
	}
}

// Update handles messages
func (m ExecutionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case executionEventMsg:
		m.handleEvent(msg.event)
		// Continue listening for more events
		return m, m.listenForEvents()

	case ExecutionCompleteMsg:
		m.completed = true
		m.finalState = msg.State
		m.finalError = msg.Error
		// Return the completion message to parent
		return m, func() tea.Msg { return msg }

	case tea.KeyMsg:
		switch msg.String() {
		case "P", "p":
			if !m.completed {
				m.paused = !m.paused
				m.addLog("Execution " + map[bool]string{true: "paused", false: "resumed"}[m.paused])
			}
			return m, nil
		case "esc":
			if !m.completed {
				m.aborted = true
				m.addLog("Execution aborted by user")
				return m, func() tea.Msg {
					return ExecutionCompleteMsg{
						State: nil,
						Error: fmt.Errorf("execution aborted by user"),
					}
				}
			}
			return m, nil
		case "q", "ctrl+c":
			if m.completed {
				return m, tea.Quit
			}
			return m, nil
		}
	}

	return m, nil
}

// handleEvent processes an execution event
func (m *ExecutionModel) handleEvent(event engine.ExecutionEvent) {
	switch event.Type {
	case engine.EventStepStart:
		if event.Step != nil {
			idx := m.findStepIndex(event.Step.ID)
			if idx >= 0 {
				m.steps[idx].status = StepRunning
				m.steps[idx].startTime = event.Timestamp
				m.currentStep = idx
				m.addLogForStep(idx, fmt.Sprintf("Started: %s", event.Step.Name))
			}
		}

	case engine.EventStepComplete:
		if event.Step != nil {
			idx := m.findStepIndex(event.Step.ID)
			if idx >= 0 {
				m.steps[idx].status = StepCompleted
				m.steps[idx].duration = time.Since(m.steps[idx].startTime)
				if event.Result != nil {
					m.addLogForStep(idx, fmt.Sprintf("Completed in %.2fs", event.Result.Duration))
				}
			}
		}

	case engine.EventStepError:
		if event.Step != nil {
			idx := m.findStepIndex(event.Step.ID)
			if idx >= 0 {
				m.steps[idx].status = StepFailed
				m.steps[idx].duration = time.Since(m.steps[idx].startTime)
				m.steps[idx].error = event.Error
				if event.Error != nil {
					m.addLogForStep(idx, fmt.Sprintf("Error: %v", event.Error))
				}
			}
		}

	case engine.EventStepSkipped:
		if event.Step != nil {
			idx := m.findStepIndex(event.Step.ID)
			if idx >= 0 {
				m.steps[idx].status = StepSkipped
				m.addLogForStep(idx, fmt.Sprintf("Skipped: %s", event.Message))
			}
		}

	case engine.EventPaused:
		m.paused = true

	case engine.EventResumed:
		m.paused = false
	}
}

// findStepIndex finds the index of a step by ID
func (m *ExecutionModel) findStepIndex(id string) int {
	for i, s := range m.steps {
		if s.step.ID == id {
			return i
		}
	}
	return -1
}

// addLog adds a log message to the current step
func (m *ExecutionModel) addLog(msg string) {
	if m.currentStep >= 0 && m.currentStep < len(m.steps) {
		m.addLogForStep(m.currentStep, msg)
	}
}

// addLogForStep adds a log message to a specific step
func (m *ExecutionModel) addLogForStep(idx int, msg string) {
	if idx >= 0 && idx < len(m.steps) {
		m.steps[idx].logs = append(m.steps[idx].logs, msg)
		// Keep only the last 5 logs per step
		if len(m.steps[idx].logs) > 5 {
			m.steps[idx].logs = m.steps[idx].logs[len(m.steps[idx].logs)-5:]
		}
	}
}

// View renders the execution screen
func (m ExecutionModel) View() string {
	var b strings.Builder

	// Header
	header := styles.Title.Render(fmt.Sprintf("Executing: %s", m.workflow.Name))
	b.WriteString(header)
	b.WriteString("\n\n")

	// Status line
	statusLine := m.renderStatusLine()
	b.WriteString(statusLine)
	b.WriteString("\n\n")

	// Steps list
	stepsView := m.renderSteps()
	b.WriteString(stepsView)
	b.WriteString("\n")

	// Recent logs
	logsView := m.renderLogs()
	if logsView != "" {
		b.WriteString("\n")
		b.WriteString(styles.Subtitle.Render("Recent Activity:"))
		b.WriteString("\n")
		b.WriteString(logsView)
	}

	// Help
	b.WriteString("\n")
	help := m.renderHelp()
	b.WriteString(help)

	return styles.Container.Render(b.String())
}

// renderStatusLine renders the current execution status
func (m ExecutionModel) renderStatusLine() string {
	if m.aborted {
		return styles.StatusError.Render("Status: Aborted")
	}
	if m.completed {
		if m.finalError != nil {
			return styles.StatusError.Render(fmt.Sprintf("Status: Failed - %v", m.finalError))
		}
		return styles.StatusSuccess.Render("Status: Completed")
	}
	if m.paused {
		return styles.StatusRunning.Render("Status: Paused (press P to resume)")
	}

	// Calculate progress
	completed := 0
	for _, s := range m.steps {
		if s.status == StepCompleted || s.status == StepSkipped {
			completed++
		}
	}
	progress := fmt.Sprintf("Progress: %d/%d steps", completed, len(m.steps))
	return styles.StatusRunning.Render(progress)
}

// renderSteps renders the list of steps with status icons
func (m ExecutionModel) renderSteps() string {
	var lines []string

	for i, s := range m.steps {
		line := m.renderStepLine(i, s)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderStepLine renders a single step line
func (m ExecutionModel) renderStepLine(idx int, s stepInfo) string {
	// Status icon
	var icon string
	var style lipgloss.Style

	switch s.status {
	case StepCompleted:
		icon = "\u2713" // checkmark
		style = styles.StatusSuccess
	case StepRunning:
		icon = m.spinner.View()
		style = styles.StatusRunning
	case StepFailed:
		icon = "\u2717" // X
		style = styles.StatusError
	case StepSkipped:
		icon = "\u25cb" // circle (same as pending)
		style = styles.StatusPending
	default: // Pending
		icon = "\u25cb" // circle
		style = styles.StatusPending
	}

	// Step name
	name := s.step.Name
	if name == "" {
		name = s.step.ID
	}

	// Duration
	var duration string
	if s.status == StepCompleted || s.status == StepFailed {
		duration = fmt.Sprintf(" (%.2fs)", s.duration.Seconds())
	} else if s.status == StepRunning {
		duration = fmt.Sprintf(" (%.1fs...)", time.Since(s.startTime).Seconds())
	}

	// Build line
	line := fmt.Sprintf("  %s %s%s", icon, name, duration)

	// Add error message for failed steps
	if s.status == StepFailed && s.error != nil {
		line += fmt.Sprintf("\n      Error: %v", s.error)
	}

	return style.Render(line)
}

// renderLogs renders recent log messages
func (m ExecutionModel) renderLogs() string {
	var logs []string

	// Collect logs from all steps (most recent first)
	for i := len(m.steps) - 1; i >= 0; i-- {
		for j := len(m.steps[i].logs) - 1; j >= 0; j-- {
			stepName := m.steps[i].step.Name
			if stepName == "" {
				stepName = m.steps[i].step.ID
			}
			log := fmt.Sprintf("  [%s] %s", stepName, m.steps[i].logs[j])
			logs = append(logs, log)
			if len(logs) >= 5 {
				break
			}
		}
		if len(logs) >= 5 {
			break
		}
	}

	// Reverse to show oldest first
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return styles.Subtitle.Render(strings.Join(logs, "\n"))
}

// renderHelp renders the help text
func (m ExecutionModel) renderHelp() string {
	if m.completed {
		return styles.Help.Render("enter continue \u2022 q quit")
	}
	return styles.Help.Render("P pause/resume \u2022 esc abort")
}

// IsCompleted returns whether execution has finished
func (m ExecutionModel) IsCompleted() bool {
	return m.completed
}

// GetState returns the final workflow state
func (m ExecutionModel) GetState() *engine.WorkflowState {
	return m.finalState
}

// GetError returns the final error, if any
func (m ExecutionModel) GetError() error {
	return m.finalError
}

// EventsChannel returns the events channel for the engine
func (m ExecutionModel) EventsChannel() chan engine.ExecutionEvent {
	return m.events
}
