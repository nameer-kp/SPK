package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Checkpoint represents a saved workflow execution state
type Checkpoint struct {
	ID           string                 `json:"id"`
	WorkflowPath string                 `json:"workflow_path"`
	WorkflowHash string                 `json:"workflow_hash"` // SHA256 of workflow file
	WorkflowName string                 `json:"workflow_name"`
	StartedAt    time.Time              `json:"started_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CurrentStep  int                    `json:"current_step"`
	Status       string                 `json:"status"`
	Inputs       map[string]interface{} `json:"inputs"`
	Context      map[string]interface{} `json:"context"`
	Results      []StepResult           `json:"results"`
}

// checkpointStepResult is used for JSON serialization since StepResult.Error is an error type
type checkpointStepResult struct {
	StepID   string      `json:"step_id"`
	Success  bool        `json:"success"`
	Data     interface{} `json:"data"`
	Error    string      `json:"error,omitempty"`
	Duration float64     `json:"duration"`
	Status   string      `json:"status"`
}

// CheckpointManager handles saving and loading workflow checkpoints
type CheckpointManager struct {
	dir string
}

// NewCheckpointManager creates a new checkpoint manager with the given directory
func NewCheckpointManager(dir string) *CheckpointManager {
	return &CheckpointManager{
		dir: dir,
	}
}

// DefaultCheckpointDir returns the default checkpoint directory (~/.flowcli/checkpoints/)
func DefaultCheckpointDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".flowcli/checkpoints"
	}
	return filepath.Join(home, ".flowcli", "checkpoints")
}

// Save saves the current workflow state as a checkpoint and returns the checkpoint ID
func (m *CheckpointManager) Save(state *WorkflowState, workflowPath string) (string, error) {
	// Ensure directory exists
	if err := os.MkdirAll(m.dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	// Calculate workflow hash
	hash, err := m.calculateWorkflowHash(workflowPath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate workflow hash: %w", err)
	}

	now := time.Now()

	// Generate checkpoint ID
	workflowName := state.Workflow.Name
	// Sanitize workflow name for use in filename
	safeName := strings.ReplaceAll(workflowName, " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	id := fmt.Sprintf("%s_%d", safeName, now.UnixNano())

	checkpoint := &Checkpoint{
		ID:           id,
		WorkflowPath: workflowPath,
		WorkflowHash: hash,
		WorkflowName: workflowName,
		StartedAt:    now,
		UpdatedAt:    now,
		CurrentStep:  state.CurrentStep,
		Status:       state.Status,
		Inputs:       state.Inputs,
		Context:      state.Context,
		Results:      state.Results,
	}

	// Write checkpoint to file
	if err := m.writeCheckpoint(checkpoint); err != nil {
		return "", err
	}

	return id, nil
}

// Load loads a checkpoint by ID
func (m *CheckpointManager) Load(id string) (*Checkpoint, error) {
	path := filepath.Join(m.dir, id+".json")
	return m.readCheckpoint(path)
}

// List returns all checkpoints sorted by UpdatedAt descending (most recent first)
func (m *CheckpointManager) List() ([]Checkpoint, error) {
	if _, err := os.Stat(m.dir); os.IsNotExist(err) {
		return []Checkpoint{}, nil
	}

	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint directory: %w", err)
	}

	var checkpoints []Checkpoint
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(m.dir, entry.Name())
		checkpoint, err := m.readCheckpoint(path)
		if err != nil {
			continue // Skip invalid checkpoint files
		}

		checkpoints = append(checkpoints, *checkpoint)
	}

	// Sort by UpdatedAt descending
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].UpdatedAt.After(checkpoints[j].UpdatedAt)
	})

	return checkpoints, nil
}

// Delete removes a checkpoint by ID
func (m *CheckpointManager) Delete(id string) error {
	path := filepath.Join(m.dir, id+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("checkpoint not found: %s", id)
	}
	return os.Remove(path)
}

// FindByWorkflow finds the most recent checkpoint for a given workflow path
func (m *CheckpointManager) FindByWorkflow(workflowPath string) (*Checkpoint, error) {
	checkpoints, err := m.List()
	if err != nil {
		return nil, err
	}

	// Convert workflow path to absolute for comparison
	absPath, err := filepath.Abs(workflowPath)
	if err != nil {
		absPath = workflowPath
	}

	for _, cp := range checkpoints {
		cpAbsPath, err := filepath.Abs(cp.WorkflowPath)
		if err != nil {
			cpAbsPath = cp.WorkflowPath
		}

		if cpAbsPath == absPath {
			// Return a copy to avoid mutation issues
			result := cp
			return &result, nil
		}
	}

	return nil, fmt.Errorf("no checkpoint found for workflow: %s", workflowPath)
}

// ValidateWorkflowUnchanged checks if the workflow file has changed since the checkpoint was created
func (m *CheckpointManager) ValidateWorkflowUnchanged(checkpoint *Checkpoint) (bool, error) {
	currentHash, err := m.calculateWorkflowHash(checkpoint.WorkflowPath)
	if err != nil {
		return false, fmt.Errorf("failed to calculate current workflow hash: %w", err)
	}

	return currentHash == checkpoint.WorkflowHash, nil
}

// calculateWorkflowHash computes SHA256 hash of a workflow file
func (m *CheckpointManager) calculateWorkflowHash(workflowPath string) (string, error) {
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// writeCheckpoint writes a checkpoint to a JSON file
func (m *CheckpointManager) writeCheckpoint(checkpoint *Checkpoint) error {
	path := filepath.Join(m.dir, checkpoint.ID+".json")

	// Convert StepResults to serializable format
	serializable := struct {
		ID           string                 `json:"id"`
		WorkflowPath string                 `json:"workflow_path"`
		WorkflowHash string                 `json:"workflow_hash"`
		WorkflowName string                 `json:"workflow_name"`
		StartedAt    time.Time              `json:"started_at"`
		UpdatedAt    time.Time              `json:"updated_at"`
		CurrentStep  int                    `json:"current_step"`
		Status       string                 `json:"status"`
		Inputs       map[string]interface{} `json:"inputs"`
		Context      map[string]interface{} `json:"context"`
		Results      []checkpointStepResult `json:"results"`
	}{
		ID:           checkpoint.ID,
		WorkflowPath: checkpoint.WorkflowPath,
		WorkflowHash: checkpoint.WorkflowHash,
		WorkflowName: checkpoint.WorkflowName,
		StartedAt:    checkpoint.StartedAt,
		UpdatedAt:    checkpoint.UpdatedAt,
		CurrentStep:  checkpoint.CurrentStep,
		Status:       checkpoint.Status,
		Inputs:       checkpoint.Inputs,
		Context:      checkpoint.Context,
		Results:      make([]checkpointStepResult, len(checkpoint.Results)),
	}

	for i, r := range checkpoint.Results {
		errStr := ""
		if r.Error != nil {
			errStr = r.Error.Error()
		}
		serializable.Results[i] = checkpointStepResult{
			StepID:   r.StepID,
			Success:  r.Success,
			Data:     r.Data,
			Error:    errStr,
			Duration: r.Duration,
			Status:   r.Status,
		}
	}

	data, err := json.MarshalIndent(serializable, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %w", err)
	}

	return nil
}

// readCheckpoint reads a checkpoint from a JSON file
func (m *CheckpointManager) readCheckpoint(path string) (*Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	// Parse into serializable format first
	var serializable struct {
		ID           string                 `json:"id"`
		WorkflowPath string                 `json:"workflow_path"`
		WorkflowHash string                 `json:"workflow_hash"`
		WorkflowName string                 `json:"workflow_name"`
		StartedAt    time.Time              `json:"started_at"`
		UpdatedAt    time.Time              `json:"updated_at"`
		CurrentStep  int                    `json:"current_step"`
		Status       string                 `json:"status"`
		Inputs       map[string]interface{} `json:"inputs"`
		Context      map[string]interface{} `json:"context"`
		Results      []checkpointStepResult `json:"results"`
	}

	if err := json.Unmarshal(data, &serializable); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	checkpoint := &Checkpoint{
		ID:           serializable.ID,
		WorkflowPath: serializable.WorkflowPath,
		WorkflowHash: serializable.WorkflowHash,
		WorkflowName: serializable.WorkflowName,
		StartedAt:    serializable.StartedAt,
		UpdatedAt:    serializable.UpdatedAt,
		CurrentStep:  serializable.CurrentStep,
		Status:       serializable.Status,
		Inputs:       serializable.Inputs,
		Context:      serializable.Context,
		Results:      make([]StepResult, len(serializable.Results)),
	}

	for i, r := range serializable.Results {
		var stepErr error
		if r.Error != "" {
			stepErr = fmt.Errorf("%s", r.Error)
		}
		checkpoint.Results[i] = StepResult{
			StepID:   r.StepID,
			Success:  r.Success,
			Data:     r.Data,
			Error:    stepErr,
			Duration: r.Duration,
			Status:   r.Status,
		}
	}

	return checkpoint, nil
}
