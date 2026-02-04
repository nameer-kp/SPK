package tools

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Checker verifies CLI tool availability
type Checker struct{}

// NewChecker creates a new tool checker
func NewChecker() *Checker {
	return &Checker{}
}

// CheckResult holds the result of a tool check
type CheckResult struct {
	Installed      bool
	Path           string
	InstallCommand string
}

// Check verifies if a tool is installed and returns install instructions if not
func (c *Checker) Check(tool string) CheckResult {
	path, err := exec.LookPath(tool)
	if err == nil {
		return CheckResult{Installed: true, Path: path}
	}

	return CheckResult{
		Installed:      false,
		InstallCommand: c.getInstallCommand(tool),
	}
}

// getInstallCommand returns OS-specific install command for a tool
func (c *Checker) getInstallCommand(tool string) string {
	commands := map[string]map[string]string{
		"gcloud": {
			"darwin": "brew install google-cloud-sdk",
			"linux":  "curl https://sdk.cloud.google.com | bash",
		},
		"chafa": {
			"darwin": "brew install chafa",
			"linux":  "apt install chafa",
		},
		"viu": {
			"darwin": "brew install viu",
			"linux":  "cargo install viu",
		},
		"imgcat": {
			"darwin": "brew install imgcat",
			"linux":  "(iTerm2 only)",
		},
		"timg": {
			"darwin": "brew install timg",
			"linux":  "apt install timg",
		},
	}

	if toolCmds, ok := commands[tool]; ok {
		if cmd, ok := toolCmds[runtime.GOOS]; ok {
			return cmd
		}
	}

	// Generic fallback
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("brew install %s", tool)
	case "linux":
		return fmt.Sprintf("apt install %s", tool)
	default:
		return fmt.Sprintf("Install %s for your system", tool)
	}
}
