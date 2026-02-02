package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultCheckpointDir(t *testing.T) {
	dir := DefaultCheckpointDir()
	if dir == "" {
		t.Error("DefaultCheckpointDir() returned empty string")
	}
	if !filepath.IsAbs(dir) && dir != ".flowcli/checkpoints" {
		t.Errorf("DefaultCheckpointDir() should return absolute path or fallback, got %s", dir)
	}
}

func TestCheckpointManager_SaveAndLoad(t *testing.T) {
	// Create temp directory for checkpoints
	tmpDir, err := os.MkdirTemp("", "checkpoint_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test workflow file
	workflowPath := filepath.Join(tmpDir, "test_workflow.yaml")
	workflowContent := `name: test-workflow
description: A test workflow
steps:
  - id: step1
    name: Test Step
    type: http
    config:
      method: GET
      url: http://example.com
`
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}

	checkpointDir := filepath.Join(tmpDir, "checkpoints")
	manager := NewCheckpointManager(checkpointDir)

	// Create a workflow state
	workflow := &Workflow{
		Name:        "test-workflow",
		Description: "A test workflow",
	}

	state := &WorkflowState{
		Workflow:    workflow,
		CurrentStep: 2,
		Context:     map[string]interface{}{"key": "value"},
		Inputs:      map[string]interface{}{"input1": "val1"},
		Results: []StepResult{
			{
				StepID:   "step1",
				Success:  true,
				Data:     map[string]interface{}{"response": "data"},
				Duration: 1.5,
				Status:   "200 OK",
			},
			{
				StepID:   "step2",
				Success:  false,
				Error:    errors.New("something went wrong"),
				Duration: 0.5,
				Status:   "500 Error",
			},
		},
		Status: "paused",
	}

	// Save checkpoint
	id, err := manager.Save(state, workflowPath)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	if id == "" {
		t.Error("Save() returned empty checkpoint ID")
	}

	// Verify ID format: <workflow_name>_<timestamp_nanos>
	if len(id) < len("test-workflow_") {
		t.Errorf("Checkpoint ID too short: %s", id)
	}

	// Load checkpoint
	loaded, err := manager.Load(id)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded data
	if loaded.ID != id {
		t.Errorf("ID mismatch: got %s, want %s", loaded.ID, id)
	}
	if loaded.WorkflowPath != workflowPath {
		t.Errorf("WorkflowPath mismatch: got %s, want %s", loaded.WorkflowPath, workflowPath)
	}
	if loaded.WorkflowName != "test-workflow" {
		t.Errorf("WorkflowName mismatch: got %s, want %s", loaded.WorkflowName, "test-workflow")
	}
	if loaded.CurrentStep != 2 {
		t.Errorf("CurrentStep mismatch: got %d, want %d", loaded.CurrentStep, 2)
	}
	if loaded.Status != "paused" {
		t.Errorf("Status mismatch: got %s, want %s", loaded.Status, "paused")
	}
	if loaded.Inputs["input1"] != "val1" {
		t.Errorf("Inputs mismatch: got %v", loaded.Inputs)
	}
	if loaded.Context["key"] != "value" {
		t.Errorf("Context mismatch: got %v", loaded.Context)
	}
	if len(loaded.Results) != 2 {
		t.Errorf("Results length mismatch: got %d, want %d", len(loaded.Results), 2)
	}
	if loaded.Results[0].StepID != "step1" || !loaded.Results[0].Success {
		t.Errorf("Results[0] mismatch: got %+v", loaded.Results[0])
	}
	if loaded.Results[1].StepID != "step2" || loaded.Results[1].Success {
		t.Errorf("Results[1] mismatch: got %+v", loaded.Results[1])
	}
	if loaded.Results[1].Error == nil || loaded.Results[1].Error.Error() != "something went wrong" {
		t.Errorf("Results[1].Error mismatch: got %v", loaded.Results[1].Error)
	}
	if loaded.WorkflowHash == "" {
		t.Error("WorkflowHash should not be empty")
	}
}

func TestCheckpointManager_List(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test workflow file
	workflowPath := filepath.Join(tmpDir, "test_workflow.yaml")
	workflowContent := `name: test-workflow
steps:
  - id: step1
    type: http
    config: {}
`
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}

	checkpointDir := filepath.Join(tmpDir, "checkpoints")
	manager := NewCheckpointManager(checkpointDir)

	workflow := &Workflow{Name: "test-workflow"}

	// Save multiple checkpoints with slight delay
	var ids []string
	for i := 0; i < 3; i++ {
		state := &WorkflowState{
			Workflow:    workflow,
			CurrentStep: i,
			Context:     map[string]interface{}{},
			Inputs:      map[string]interface{}{},
			Results:     []StepResult{},
			Status:      "paused",
		}
		id, err := manager.Save(state, workflowPath)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}
		ids = append(ids, id)
		time.Sleep(10 * time.Millisecond) // Small delay to ensure different timestamps
	}

	// List checkpoints
	checkpoints, err := manager.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(checkpoints) != 3 {
		t.Errorf("List() returned %d checkpoints, want 3", len(checkpoints))
	}

	// Verify sorted by UpdatedAt descending (most recent first)
	for i := 0; i < len(checkpoints)-1; i++ {
		if checkpoints[i].UpdatedAt.Before(checkpoints[i+1].UpdatedAt) {
			t.Errorf("Checkpoints not sorted correctly: %v should be after %v",
				checkpoints[i].UpdatedAt, checkpoints[i+1].UpdatedAt)
		}
	}

	// Most recent should be the last one saved
	if checkpoints[0].ID != ids[2] {
		t.Errorf("First checkpoint should be most recent: got %s, want %s", checkpoints[0].ID, ids[2])
	}
}

func TestCheckpointManager_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test workflow file
	workflowPath := filepath.Join(tmpDir, "test_workflow.yaml")
	workflowContent := `name: test-workflow
steps:
  - id: step1
    type: http
    config: {}
`
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}

	checkpointDir := filepath.Join(tmpDir, "checkpoints")
	manager := NewCheckpointManager(checkpointDir)

	workflow := &Workflow{Name: "test-workflow"}
	state := &WorkflowState{
		Workflow:    workflow,
		CurrentStep: 0,
		Context:     map[string]interface{}{},
		Inputs:      map[string]interface{}{},
		Results:     []StepResult{},
		Status:      "paused",
	}

	// Save checkpoint
	id, err := manager.Save(state, workflowPath)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify it exists
	_, err = manager.Load(id)
	if err != nil {
		t.Fatalf("Load() failed after save: %v", err)
	}

	// Delete checkpoint
	err = manager.Delete(id)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify it's gone
	_, err = manager.Load(id)
	if err == nil {
		t.Error("Load() should fail after delete")
	}

	// Delete non-existent checkpoint
	err = manager.Delete("non-existent")
	if err == nil {
		t.Error("Delete() should fail for non-existent checkpoint")
	}
}

func TestCheckpointManager_FindByWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two test workflow files
	workflowPath1 := filepath.Join(tmpDir, "workflow1.yaml")
	workflowPath2 := filepath.Join(tmpDir, "workflow2.yaml")
	workflowContent := `name: test-workflow
steps:
  - id: step1
    type: http
    config: {}
`
	if err := os.WriteFile(workflowPath1, []byte(workflowContent), 0644); err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}
	if err := os.WriteFile(workflowPath2, []byte(workflowContent), 0644); err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}

	checkpointDir := filepath.Join(tmpDir, "checkpoints")
	manager := NewCheckpointManager(checkpointDir)

	// Save checkpoints for workflow1
	workflow1 := &Workflow{Name: "workflow1"}
	for i := 0; i < 2; i++ {
		state := &WorkflowState{
			Workflow:    workflow1,
			CurrentStep: i,
			Context:     map[string]interface{}{},
			Inputs:      map[string]interface{}{},
			Results:     []StepResult{},
			Status:      "paused",
		}
		_, err := manager.Save(state, workflowPath1)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Save checkpoint for workflow2
	workflow2 := &Workflow{Name: "workflow2"}
	state := &WorkflowState{
		Workflow:    workflow2,
		CurrentStep: 5,
		Context:     map[string]interface{}{},
		Inputs:      map[string]interface{}{},
		Results:     []StepResult{},
		Status:      "paused",
	}
	_, err = manager.Save(state, workflowPath2)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Find by workflow1 - should return most recent (CurrentStep=1)
	found, err := manager.FindByWorkflow(workflowPath1)
	if err != nil {
		t.Fatalf("FindByWorkflow() failed: %v", err)
	}
	if found.CurrentStep != 1 {
		t.Errorf("FindByWorkflow() returned wrong checkpoint: CurrentStep=%d, want 1", found.CurrentStep)
	}

	// Find by workflow2
	found, err = manager.FindByWorkflow(workflowPath2)
	if err != nil {
		t.Fatalf("FindByWorkflow() failed: %v", err)
	}
	if found.CurrentStep != 5 {
		t.Errorf("FindByWorkflow() returned wrong checkpoint: CurrentStep=%d, want 5", found.CurrentStep)
	}

	// Find non-existent workflow
	_, err = manager.FindByWorkflow(filepath.Join(tmpDir, "nonexistent.yaml"))
	if err == nil {
		t.Error("FindByWorkflow() should fail for non-existent workflow")
	}
}

func TestCheckpointManager_ValidateWorkflowUnchanged(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test workflow file
	workflowPath := filepath.Join(tmpDir, "test_workflow.yaml")
	workflowContent := `name: test-workflow
steps:
  - id: step1
    type: http
    config: {}
`
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}

	checkpointDir := filepath.Join(tmpDir, "checkpoints")
	manager := NewCheckpointManager(checkpointDir)

	workflow := &Workflow{Name: "test-workflow"}
	state := &WorkflowState{
		Workflow:    workflow,
		CurrentStep: 0,
		Context:     map[string]interface{}{},
		Inputs:      map[string]interface{}{},
		Results:     []StepResult{},
		Status:      "paused",
	}

	// Save checkpoint
	id, err := manager.Save(state, workflowPath)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	checkpoint, err := manager.Load(id)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Validate unchanged
	unchanged, err := manager.ValidateWorkflowUnchanged(checkpoint)
	if err != nil {
		t.Fatalf("ValidateWorkflowUnchanged() failed: %v", err)
	}
	if !unchanged {
		t.Error("ValidateWorkflowUnchanged() should return true for unchanged workflow")
	}

	// Modify workflow file
	modifiedContent := `name: test-workflow-modified
steps:
  - id: step1
    type: http
    config: {}
  - id: step2
    type: shell
    config: {}
`
	if err := os.WriteFile(workflowPath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify workflow file: %v", err)
	}

	// Validate changed
	unchanged, err = manager.ValidateWorkflowUnchanged(checkpoint)
	if err != nil {
		t.Fatalf("ValidateWorkflowUnchanged() failed: %v", err)
	}
	if unchanged {
		t.Error("ValidateWorkflowUnchanged() should return false for changed workflow")
	}
}

func TestCheckpointManager_MissingDirectory(t *testing.T) {
	// Test with non-existent directory that will be created
	tmpDir, err := os.MkdirTemp("", "checkpoint_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test workflow file
	workflowPath := filepath.Join(tmpDir, "test_workflow.yaml")
	workflowContent := `name: test-workflow
steps:
  - id: step1
    type: http
    config: {}
`
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}

	// Use a nested path that doesn't exist
	checkpointDir := filepath.Join(tmpDir, "nested", "path", "checkpoints")
	manager := NewCheckpointManager(checkpointDir)

	// List should return empty for non-existent directory
	checkpoints, err := manager.List()
	if err != nil {
		t.Fatalf("List() failed for non-existent directory: %v", err)
	}
	if len(checkpoints) != 0 {
		t.Errorf("List() should return empty for non-existent directory, got %d", len(checkpoints))
	}

	// Save should create the directory
	workflow := &Workflow{Name: "test-workflow"}
	state := &WorkflowState{
		Workflow:    workflow,
		CurrentStep: 0,
		Context:     map[string]interface{}{},
		Inputs:      map[string]interface{}{},
		Results:     []StepResult{},
		Status:      "paused",
	}

	id, err := manager.Save(state, workflowPath)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(checkpointDir); os.IsNotExist(err) {
		t.Error("Save() should create checkpoint directory")
	}

	// Verify checkpoint was saved
	_, err = manager.Load(id)
	if err != nil {
		t.Fatalf("Load() failed after Save(): %v", err)
	}
}

func TestCheckpointManager_LoadNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager := NewCheckpointManager(tmpDir)

	// Try to load non-existent checkpoint
	_, err = manager.Load("non-existent-checkpoint")
	if err == nil {
		t.Error("Load() should fail for non-existent checkpoint")
	}
}

func TestCheckpointManager_WorkflowNameSanitization(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test workflow file
	workflowPath := filepath.Join(tmpDir, "test_workflow.yaml")
	workflowContent := `name: test workflow with spaces/slashes
steps:
  - id: step1
    type: http
    config: {}
`
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatalf("Failed to write workflow file: %v", err)
	}

	checkpointDir := filepath.Join(tmpDir, "checkpoints")
	manager := NewCheckpointManager(checkpointDir)

	workflow := &Workflow{Name: "test workflow with spaces/slashes"}
	state := &WorkflowState{
		Workflow:    workflow,
		CurrentStep: 0,
		Context:     map[string]interface{}{},
		Inputs:      map[string]interface{}{},
		Results:     []StepResult{},
		Status:      "paused",
	}

	// Save checkpoint - ID should have sanitized name
	id, err := manager.Save(state, workflowPath)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// ID should not contain spaces or slashes
	if containsAny(id, " /") {
		t.Errorf("Checkpoint ID should not contain spaces or slashes: %s", id)
	}

	// Should still be loadable
	_, err = manager.Load(id)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
}

func containsAny(s, chars string) bool {
	for _, c := range chars {
		for _, sc := range s {
			if c == sc {
				return true
			}
		}
	}
	return false
}
