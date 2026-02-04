package nodes

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nameer-kp/atem/pkg/node"
)

// ShellNode executes shell commands
type ShellNode struct{}

func NewShellNode() *ShellNode {
	return &ShellNode{}
}

func (n *ShellNode) Name() string {
	return "shell"
}

func (n *ShellNode) Description() string {
	return "Execute shell commands"
}

func (n *ShellNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"command": "string",
		"args":    "array",
		"dir":     "string",
		"env":     "object",
		"stdin":   "string",
	}
}

func (n *ShellNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	command, ok := config["command"].(string)
	if !ok || command == "" {
		return node.Result{}, fmt.Errorf("command is required")
	}

	// Build args
	var args []string
	if argsRaw, ok := config["args"].([]interface{}); ok {
		for _, a := range argsRaw {
			args = append(args, fmt.Sprintf("%v", a))
		}
	}

	ctx.Logger().Info("executing shell command", "command", command, "args", args)

	// Use execFile pattern (not shell) to avoid injection
	cmd := exec.Command(command, args...)

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

	// Set stdin
	if stdin, ok := config["stdin"].(string); ok && stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	result := map[string]interface{}{
		"stdout":    stdout.String(),
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
		}, nil // Don't return error to allow workflow to handle it
	}

	return node.Result{
		Success: true,
		Data:    result,
		Status:  "Exit 0",
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("stdout: %s", strings.TrimSpace(stdout.String()))},
		},
	}, nil
}
