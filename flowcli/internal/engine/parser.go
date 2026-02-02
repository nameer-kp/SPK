package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Parser handles loading and parsing workflow YAML files
type Parser struct{}

// NewParser creates a new workflow parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile loads and parses a workflow from a file path
func (p *Parser) ParseFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses workflow YAML content
func (p *Parser) Parse(data []byte) (*Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	if err := p.validate(&workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// validate checks workflow structure for required fields
func (p *Parser) validate(w *Workflow) error {
	if w.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(w.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	stepIDs := make(map[string]bool)
	for i, step := range w.Steps {
		if step.ID == "" {
			return fmt.Errorf("step %d: id is required", i+1)
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("step %d: duplicate id '%s'", i+1, step.ID)
		}
		stepIDs[step.ID] = true

		if step.Type == "" {
			return fmt.Errorf("step '%s': type is required", step.ID)
		}

		if step.Config == nil {
			return fmt.Errorf("step '%s': config is required", step.ID)
		}
	}

	return nil
}

// ListWorkflows returns workflow files in a directory
func (p *Parser) ListWorkflows(dir string) ([]WorkflowInfo, error) {
	var workflows []WorkflowInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return workflows, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		w, err := p.ParseFile(path)
		if err != nil {
			continue // Skip invalid workflows
		}

		workflows = append(workflows, WorkflowInfo{
			Path:        path,
			Name:        w.Name,
			Description: w.Description,
			Version:     w.Version,
			StepCount:   len(w.Steps),
		})
	}

	return workflows, nil
}

// WorkflowInfo holds summary info for listing workflows
type WorkflowInfo struct {
	Path        string
	Name        string
	Description string
	Version     string
	StepCount   int
}
