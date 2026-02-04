package nodes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nameer-kp/atem/internal/tools"
	"github.com/nameer-kp/atem/pkg/node"
)

// CommandNode executes shell commands with tool checking
type CommandNode struct {
	checker *tools.Checker
}

func NewCommandNode() *CommandNode {
	return &CommandNode{
		checker: tools.NewChecker(),
	}
}

func (n *CommandNode) Name() string {
	return "command"
}

func (n *CommandNode) Description() string {
	return "Execute shell commands with tool verification"
}

func (n *CommandNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"run":   "string",
		"check": "string",
		"dir":   "string",
		"env":   "object",
	}
}

func (n *CommandNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	// Check required tool if specified
	if checkTool, ok := config["check"].(string); ok && checkTool != "" {
		result := n.checker.Check(checkTool)
		if !result.Installed {
			return node.Result{
				Success: false,
				Data: map[string]interface{}{
					"error":           fmt.Sprintf("Required tool '%s' is not installed", checkTool),
					"install_command": result.InstallCommand,
				},
				Status: "Tool not found",
				Logs: []node.LogEntry{
					{Level: "error", Message: fmt.Sprintf("Tool '%s' not found. Install with: %s", checkTool, result.InstallCommand)},
				},
			}, nil
		}
	}

	run, ok := config["run"].(string)
	if !ok || run == "" {
		return node.Result{}, fmt.Errorf("run is required")
	}

	ctx.Logger().Info("executing command", "run", run)

	// Execute via shell to support pipes and redirects
	cmd := exec.Command("sh", "-c", run)

	// Set working directory
	if dir, ok := config["dir"].(string); ok && dir != "" {
		cmd.Dir = dir
	}

	// Set environment
	cmd.Env = os.Environ()
	if envMap, ok := config["env"].(map[string]interface{}); ok {
		for k, v := range envMap {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", k, v))
		}
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	// Try to parse stdout as JSON
	var data interface{}
	stdoutStr := strings.TrimSpace(stdout.String())
	if err := json.Unmarshal([]byte(stdoutStr), &data); err != nil {
		// Not JSON, use raw string
		data = map[string]interface{}{
			"stdout":    stdoutStr,
			"stderr":    stderr.String(),
			"exit_code": 0,
		}
	}

	result := map[string]interface{}{
		"data":      data,
		"stdout":    stdoutStr,
		"stderr":    stderr.String(),
		"exit_code": 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["exit_code"] = exitErr.ExitCode()
		}

		return node.Result{
			Success: false,
			Data:    result,
			Status:  fmt.Sprintf("Exit %v", result["exit_code"]),
			Logs: []node.LogEntry{
				{Level: "error", Message: fmt.Sprintf("Command failed: %v", err)},
				{Level: "info", Message: fmt.Sprintf("stderr: %s", stderr.String())},
			},
		}, nil
	}

	return node.Result{
		Success: true,
		Data:    result,
		Status:  "Exit 0",
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Output: %s", stdoutStr)},
		},
	}, nil
}
