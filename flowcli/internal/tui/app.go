package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nameer-kp/flowcli/internal/config"
	"github.com/nameer-kp/flowcli/internal/engine"
	"github.com/nameer-kp/flowcli/internal/tui/screens"
)

type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenWorkflowSelect
	ScreenInputWizard
	ScreenExecution
	ScreenErrorRecovery
	ScreenResult
)

// WelcomeSelectedMsg is emitted when a menu item is selected on the welcome screen
type WelcomeSelectedMsg struct {
	Index int
}

type App struct {
	// Current screen
	screen Screen

	// Dependencies
	engine *engine.Engine
	parser *engine.Parser
	config *config.ResolvedConfig

	// App metadata
	profile string
	version string
	width   int
	height  int

	// Execution state
	selectedWorkflow *engine.Workflow
	workflowPath     string
	inputs           map[string]interface{}
	executionStart   time.Time
	failedStep       *engine.Step
	failedError      error

	// Screen models
	welcomeScreen       screens.WelcomeScreen
	workflowSelectModel screens.WorkflowSelectModel
	inputWizardModel    screens.InputWizardModel
	executionModel      screens.ExecutionModel
	errorRecoveryModel  screens.ErrorRecoveryModel
	resultModel         screens.ResultModel
}

func NewApp(eng *engine.Engine, parser *engine.Parser, cfg *config.ResolvedConfig, version string) App {
	return App{
		screen:        ScreenWelcome,
		engine:        eng,
		parser:        parser,
		config:        cfg,
		profile:       cfg.Profile.Name,
		version:       version,
		welcomeScreen: screens.NewWelcomeScreen(cfg.Profile.Name, version),
	}
}

func (a App) Init() tea.Cmd {
	return a.welcomeScreen.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle global messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Pass to current screen
	}

	// Handle screen-specific messages
	switch a.screen {
	case ScreenWelcome:
		return a.updateWelcome(msg)
	case ScreenWorkflowSelect:
		return a.updateWorkflowSelect(msg)
	case ScreenInputWizard:
		return a.updateInputWizard(msg)
	case ScreenExecution:
		return a.updateExecution(msg)
	case ScreenErrorRecovery:
		return a.updateErrorRecovery(msg)
	case ScreenResult:
		return a.updateResult(msg)
	}

	return a, nil
}

func (a App) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle enter key to navigate based on selection
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" {
			selectedIndex := a.welcomeScreen.SelectedIndex()
			switch selectedIndex {
			case 0: // Run Workflow
				// Initialize workflow select screen
				a.workflowSelectModel = screens.NewWorkflowSelectModel(a.parser, a.config.WorkflowsDir)
				a.screen = ScreenWorkflowSelect
				return a, a.workflowSelectModel.Init()
			// Other menu items can be handled later
			default:
				// For now, do nothing for other menu items
			}
		}
	}

	// Update welcome screen
	updated, cmd := a.welcomeScreen.Update(msg)
	a.welcomeScreen = updated.(screens.WelcomeScreen)
	return a, cmd
}

func (a App) updateWorkflowSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle workflow selection
	if selected, ok := msg.(screens.WorkflowSelectedMsg); ok {
		a.selectedWorkflow = selected.Workflow
		a.workflowPath = selected.Path

		// Check if workflow has inputs
		if len(selected.Workflow.Inputs) > 0 {
			// Go to input wizard
			a.inputWizardModel = screens.NewInputWizardModel(selected.Workflow)
			a.screen = ScreenInputWizard
			return a, a.inputWizardModel.Init()
		}

		// No inputs needed, go directly to execution
		a.inputs = make(map[string]interface{})
		return a.startExecution()
	}

	// Update workflow select screen
	updated, cmd := a.workflowSelectModel.Update(msg)
	a.workflowSelectModel = updated.(screens.WorkflowSelectModel)
	return a, cmd
}

func (a App) updateInputWizard(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle inputs complete
	if complete, ok := msg.(screens.InputsCompleteMsg); ok {
		a.inputs = complete.Inputs
		return a.startExecution()
	}

	// Update input wizard
	updated, cmd := a.inputWizardModel.Update(msg)
	a.inputWizardModel = updated.(screens.InputWizardModel)
	return a, cmd
}

func (a App) updateExecution(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle execution complete
	if complete, ok := msg.(screens.ExecutionCompleteMsg); ok {
		duration := time.Since(a.executionStart)

		if complete.Error != nil {
			// Check if this is a step failure that can be recovered
			// For now, we go to error recovery if there was an error
			// In a more sophisticated implementation, we'd extract the failed step
			a.failedError = complete.Error

			// Find the failed step from the state
			if complete.State != nil {
				for i, result := range complete.State.Results {
					if !result.Success && i < len(a.selectedWorkflow.Steps) {
						a.failedStep = &a.selectedWorkflow.Steps[i]
						break
					}
				}
			}

			// If we have a failed step, show error recovery
			if a.failedStep != nil {
				a.errorRecoveryModel = screens.NewErrorRecoveryModel(a.failedStep, complete.Error)
				a.screen = ScreenErrorRecovery
				return a, a.errorRecoveryModel.Init()
			}
		}

		// Go to result screen
		state := complete.State
		if state == nil {
			// Create a minimal state for aborted workflows
			state = &engine.WorkflowState{
				Workflow: a.selectedWorkflow,
				Status:   "aborted",
				Inputs:   a.inputs,
				Results:  []engine.StepResult{},
			}
		}

		a.resultModel = screens.NewResultModel(a.selectedWorkflow, state, duration)
		a.screen = ScreenResult
		return a, a.resultModel.Init()
	}

	// Update execution screen
	updated, cmd := a.executionModel.Update(msg)
	a.executionModel = updated.(screens.ExecutionModel)
	return a, cmd
}

func (a App) updateErrorRecovery(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle recovery decision
	if decision, ok := msg.(screens.RecoveryDecisionMsg); ok {
		switch decision.Action {
		case engine.ActionRetry:
			// Re-run from current step (for now, restart the whole workflow)
			return a.startExecution()
		case engine.ActionSkip:
			// Skip the failed step and continue
			// For simplicity, we just go to result with current state
			duration := time.Since(a.executionStart)
			state := a.executionModel.GetState()
			if state == nil {
				state = &engine.WorkflowState{
					Workflow: a.selectedWorkflow,
					Status:   "partial",
					Inputs:   a.inputs,
					Results:  []engine.StepResult{},
				}
			}
			a.resultModel = screens.NewResultModel(a.selectedWorkflow, state, duration)
			a.screen = ScreenResult
			return a, a.resultModel.Init()
		case engine.ActionAbort:
			// Go back to workflow selection
			a.screen = ScreenWelcome
			a.selectedWorkflow = nil
			a.workflowPath = ""
			a.inputs = nil
			a.failedStep = nil
			a.failedError = nil
			return a, nil
		}
	}

	// Update error recovery screen
	updated, cmd := a.errorRecoveryModel.Update(msg)
	a.errorRecoveryModel = updated.(screens.ErrorRecoveryModel)
	return a, cmd
}

func (a App) updateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle result action
	if action, ok := msg.(screens.ResultActionMsg); ok {
		switch action.Action {
		case "menu":
			// Go back to welcome screen
			a.screen = ScreenWelcome
			a.selectedWorkflow = nil
			a.workflowPath = ""
			a.inputs = nil
			a.failedStep = nil
			a.failedError = nil
			return a, nil
		case "rerun":
			// Re-run the same workflow with same inputs
			return a.startExecution()
		}
	}

	// Update result screen
	updated, cmd := a.resultModel.Update(msg)
	a.resultModel = updated.(screens.ResultModel)
	return a, cmd
}

// startExecution initializes and starts workflow execution
func (a App) startExecution() (tea.Model, tea.Cmd) {
	a.executionStart = time.Now()
	a.failedStep = nil
	a.failedError = nil
	a.executionModel = screens.NewExecutionModel(a.selectedWorkflow, a.engine, a.inputs)
	a.screen = ScreenExecution
	return a, a.executionModel.Init()
}

func (a App) View() string {
	switch a.screen {
	case ScreenWelcome:
		return a.welcomeScreen.View()
	case ScreenWorkflowSelect:
		return a.workflowSelectModel.View()
	case ScreenInputWizard:
		return a.inputWizardModel.View()
	case ScreenExecution:
		return a.executionModel.View()
	case ScreenErrorRecovery:
		return a.errorRecoveryModel.View()
	case ScreenResult:
		return a.resultModel.View()
	default:
		return "Unknown screen"
	}
}
