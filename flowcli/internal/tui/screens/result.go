package screens

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nameer-kp/flowcli/internal/engine"
	"github.com/nameer-kp/flowcli/internal/tui/styles"
	"gopkg.in/yaml.v3"
)

// ResultActionMsg is emitted when a result action is selected
type ResultActionMsg struct {
	Action string // menu, rerun, export_json, export_yaml
}

// exportCompleteMsg is emitted when export is complete
type exportCompleteMsg struct {
	path string
	err  error
}

// ResultModel is the model for the result screen
type ResultModel struct {
	workflow *engine.Workflow
	state    *engine.WorkflowState
	duration time.Duration
	selected int
	width    int
	height   int
	message  string
}

// resultOption represents a menu option
type resultOption struct {
	key    string
	label  string
	action string
}

var resultOptions = []resultOption{
	{key: "enter", label: "Back to menu", action: "menu"},
	{key: "r", label: "Run again", action: "rerun"},
	{key: "e", label: "Export JSON", action: "export_json"},
	{key: "y", label: "Export YAML", action: "export_yaml"},
}

// NewResultModel creates a new result screen
func NewResultModel(workflow *engine.Workflow, state *engine.WorkflowState, duration time.Duration) ResultModel {
	return ResultModel{
		workflow: workflow,
		state:    state,
		duration: duration,
		selected: 0,
	}
}

// Init initializes the model
func (m ResultModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case exportCompleteMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Export failed: %v", msg.err)
		} else {
			m.message = fmt.Sprintf("Exported to: %s", msg.path)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(resultOptions)-1 {
				m.selected++
			}
		case "enter":
			return m, m.executeAction(resultOptions[m.selected].action)
		case "r", "R":
			return m, m.executeAction("rerun")
		case "e", "E":
			return m, m.executeAction("export_json")
		case "y", "Y":
			return m, m.executeAction("export_yaml")
		}
	}

	return m, nil
}

// executeAction handles the selected action
func (m ResultModel) executeAction(action string) tea.Cmd {
	switch action {
	case "export_json":
		return m.exportJSON()
	case "export_yaml":
		return m.exportYAML()
	default:
		return func() tea.Msg {
			return ResultActionMsg{Action: action}
		}
	}
}

// exportJSON exports results to JSON
func (m ResultModel) exportJSON() tea.Cmd {
	return func() tea.Msg {
		data := m.buildExportData()
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return exportCompleteMsg{err: err}
		}

		filename := m.generateFilename("json")
		err = os.WriteFile(filename, jsonData, 0644)
		if err != nil {
			return exportCompleteMsg{err: err}
		}

		return exportCompleteMsg{path: filename}
	}
}

// exportYAML exports results to YAML
func (m ResultModel) exportYAML() tea.Cmd {
	return func() tea.Msg {
		data := m.buildExportData()
		yamlData, err := yaml.Marshal(data)
		if err != nil {
			return exportCompleteMsg{err: err}
		}

		filename := m.generateFilename("yaml")
		err = os.WriteFile(filename, yamlData, 0644)
		if err != nil {
			return exportCompleteMsg{err: err}
		}

		return exportCompleteMsg{path: filename}
	}
}

// buildExportData creates the export data structure
func (m ResultModel) buildExportData() map[string]interface{} {
	successCount, failCount := m.countResults()

	results := make([]map[string]interface{}, 0, len(m.state.Results))
	for _, r := range m.state.Results {
		result := map[string]interface{}{
			"step_id":  r.StepID,
			"success":  r.Success,
			"duration": r.Duration,
			"status":   r.Status,
		}
		if r.Data != nil {
			result["data"] = r.Data
		}
		if r.Error != nil {
			result["error"] = r.Error.Error()
		}
		results = append(results, result)
	}

	return map[string]interface{}{
		"workflow": map[string]interface{}{
			"name":        m.workflow.Name,
			"description": m.workflow.Description,
			"version":     m.workflow.Version,
		},
		"execution": map[string]interface{}{
			"status":        m.state.Status,
			"duration":      m.duration.Seconds(),
			"duration_str":  m.formatDuration(),
			"success_count": successCount,
			"fail_count":    failCount,
			"timestamp":     time.Now().Format(time.RFC3339),
		},
		"results": results,
		"inputs":  m.state.Inputs,
	}
}

// generateFilename creates a unique filename for export
func (m ResultModel) generateFilename(ext string) string {
	timestamp := time.Now().Format("20060102-150405")
	name := m.workflow.Name
	if name == "" {
		name = "workflow"
	}
	return filepath.Join(".", fmt.Sprintf("%s-%s.%s", name, timestamp, ext))
}

// View renders the screen
func (m ResultModel) View() string {
	successCount, failCount := m.countResults()

	// Header
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Workflow Results"),
		"",
	)

	// Summary section
	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Padding(0, 2).
		MarginBottom(1)

	statusIcon := m.getStatusIcon()
	statusText := m.state.Status
	if m.state.Status == "completed" {
		statusText = styles.StatusSuccess.Render(statusText)
	} else if m.state.Status == "failed" {
		statusText = styles.StatusError.Render(statusText)
	}

	summary := lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("%s %s", statusIcon, styles.Title.Render(m.workflow.Name)),
		"",
		fmt.Sprintf("Status:   %s", statusText),
		fmt.Sprintf("Duration: %s", m.formatDuration()),
		fmt.Sprintf("Steps:    %s success, %s failed",
			styles.StatusSuccess.Render(fmt.Sprintf("%d", successCount)),
			styles.StatusError.Render(fmt.Sprintf("%d", failCount))),
	)

	// Step results
	stepsHeader := styles.Subtitle.Render("Step Results:")
	steps := m.renderStepResults()

	// Options
	options := m.renderOptions()

	// Message (export status)
	messageView := ""
	if m.message != "" {
		messageView = "\n" + styles.Subtitle.Render(m.message)
	}

	// Help
	help := styles.Help.Render("arrow keys navigate | enter select | q quit")

	return styles.Container.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			summaryStyle.Render(summary),
			stepsHeader,
			steps,
			"",
			options,
			messageView,
			help,
		),
	)
}

// renderStepResults renders the step results with icons
func (m ResultModel) renderStepResults() string {
	if len(m.state.Results) == 0 {
		return styles.Subtitle.Render("  No step results available")
	}

	var lines []string
	for i, result := range m.state.Results {
		icon := m.getStepIcon(result.Success)
		stepName := result.StepID
		if i < len(m.workflow.Steps) {
			stepName = m.workflow.Steps[i].Name
			if stepName == "" {
				stepName = m.workflow.Steps[i].ID
			}
		}

		duration := fmt.Sprintf("%.2fs", result.Duration)
		status := result.Status
		if status == "" {
			if result.Success {
				status = "OK"
			} else {
				status = "FAILED"
			}
		}

		line := fmt.Sprintf("  %s %s - %s (%s)", icon, stepName, status, duration)
		if result.Success {
			lines = append(lines, line)
		} else {
			lines = append(lines, styles.StatusError.Render(line))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderOptions renders the available options
func (m ResultModel) renderOptions() string {
	var lines []string
	for i, opt := range resultOptions {
		prefix := "  "
		style := styles.MenuItem
		if i == m.selected {
			prefix = "> "
			style = styles.MenuItemSelected
		}
		line := fmt.Sprintf("%s[%s] %s", prefix, opt.key, opt.label)
		lines = append(lines, style.Render(line))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// countResults counts success and failure results
func (m ResultModel) countResults() (success, fail int) {
	for _, r := range m.state.Results {
		if r.Success {
			success++
		} else {
			fail++
		}
	}
	return
}

// formatDuration formats the duration for display
func (m ResultModel) formatDuration() string {
	if m.duration < time.Second {
		return fmt.Sprintf("%dms", m.duration.Milliseconds())
	}
	if m.duration < time.Minute {
		return fmt.Sprintf("%.2fs", m.duration.Seconds())
	}
	return fmt.Sprintf("%dm %ds", int(m.duration.Minutes()), int(m.duration.Seconds())%60)
}

// getStatusIcon returns an icon based on workflow status
func (m ResultModel) getStatusIcon() string {
	switch m.state.Status {
	case "completed":
		return styles.StatusSuccess.Render("✓")
	case "failed":
		return styles.StatusError.Render("✗")
	case "running":
		return styles.StatusRunning.Render("~")
	default:
		return styles.StatusPending.Render("?")
	}
}

// getStepIcon returns an icon based on step success
func (m ResultModel) getStepIcon(success bool) string {
	if success {
		return styles.StatusSuccess.Render("✓")
	}
	return styles.StatusError.Render("✗")
}

// Workflow returns the workflow
func (m ResultModel) Workflow() *engine.Workflow {
	return m.workflow
}

// State returns the workflow state
func (m ResultModel) State() *engine.WorkflowState {
	return m.state
}
