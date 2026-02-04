package nodes

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nameer-kp/atem/pkg/node"
)

// LogNode provides structured logging and output capabilities
type LogNode struct{}

func NewLogNode() *LogNode {
	return &LogNode{}
}

func (n *LogNode) Name() string {
	return "log"
}

func (n *LogNode) Description() string {
	return "Log messages and data for debugging and output"
}

func (n *LogNode) ConfigSchema() map[string]any {
	return map[string]any{
		"message": "string", // message to log
		"level":   "string", // debug, info, warn, error
		"data":    "any",    // structured data to include
		"format":  "string", // text, json
		"output":  "string", // stdout, stderr, file
		"file":    "string", // file path if output=file
		"prefix":  "string", // prefix for the message
		"color":   "bool",   // colorize output
	}
}

func (n *LogNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	message, _ := config["message"].(string)

	level, _ := config["level"].(string)
	if level == "" {
		level = "info"
	}

	format, _ := config["format"].(string)
	if format == "" {
		format = "text"
	}

	output, _ := config["output"].(string)
	if output == "" {
		output = "stdout"
	}

	prefix, _ := config["prefix"].(string)
	useColor, _ := config["color"].(bool)

	// Build log entry
	timestamp := time.Now().Format(time.RFC3339)
	data := config["data"]

	var formatted string
	if format == "json" {
		entry := map[string]any{
			"timestamp": timestamp,
			"level":     level,
			"message":   message,
		}
		if data != nil {
			entry["data"] = data
		}
		jsonBytes, _ := json.Marshal(entry)
		formatted = string(jsonBytes)
	} else {
		levelStr := strings.ToUpper(level)
		if useColor {
			levelStr = colorizeLevel(level)
		}

		if prefix != "" {
			formatted = fmt.Sprintf("[%s] %s [%s] %s", timestamp, levelStr, prefix, message)
		} else {
			formatted = fmt.Sprintf("[%s] %s %s", timestamp, levelStr, message)
		}

		if data != nil {
			dataJSON, _ := json.MarshalIndent(data, "", "  ")
			formatted += "\n" + string(dataJSON)
		}
	}

	// Output
	switch output {
	case "stdout":
		fmt.Fprintln(os.Stdout, formatted)
	case "stderr":
		fmt.Fprintln(os.Stderr, formatted)
	case "file":
		filePath, _ := config["file"].(string)
		if filePath == "" {
			return node.Result{}, fmt.Errorf("file path is required for file output")
		}
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return node.Result{}, fmt.Errorf("failed to open log file: %w", err)
		}
		defer f.Close()
		if _, err := f.WriteString(formatted + "\n"); err != nil {
			return node.Result{}, fmt.Errorf("failed to write to log file: %w", err)
		}
	}

	return node.Result{
		Success: true,
		Data: map[string]any{
			"message":   message,
			"level":     level,
			"timestamp": timestamp,
			"formatted": formatted,
		},
		Status: level,
		Logs: []node.LogEntry{
			{Level: level, Message: message},
		},
	}, nil
}

func colorizeLevel(level string) string {
	// ANSI color codes
	const (
		reset  = "\033[0m"
		red    = "\033[31m"
		yellow = "\033[33m"
		blue   = "\033[34m"
		gray   = "\033[90m"
	)

	switch strings.ToLower(level) {
	case "debug":
		return gray + "DEBUG" + reset
	case "info":
		return blue + "INFO" + reset
	case "warn", "warning":
		return yellow + "WARN" + reset
	case "error":
		return red + "ERROR" + reset
	default:
		return strings.ToUpper(level)
	}
}
