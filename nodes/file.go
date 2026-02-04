package nodes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nameer-kp/atem/pkg/node"
	"gopkg.in/yaml.v3"
)

// FileNode handles file I/O operations
type FileNode struct{}

func NewFileNode() *FileNode {
	return &FileNode{}
}

func (n *FileNode) Name() string {
	return "file"
}

func (n *FileNode) Description() string {
	return "Read, write, and manage files"
}

func (n *FileNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"operation": "string", // read, write, append, delete, exists
		"path":      "string",
		"content":   "string", // for write/append
		"format":    "string", // text, json, yaml (for read)
	}
}

func (n *FileNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "read"
	}

	path, ok := config["path"].(string)
	if !ok || path == "" {
		return node.Result{}, fmt.Errorf("path is required")
	}

	// Expand path
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}

	ctx.Logger().Info("file operation", "operation", operation, "path", path)

	switch operation {
	case "read":
		return n.read(ctx, path, config)
	case "write":
		return n.write(ctx, path, config, false)
	case "append":
		return n.write(ctx, path, config, true)
	case "delete":
		return n.delete(ctx, path)
	case "exists":
		return n.exists(ctx, path)
	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *FileNode) read(ctx node.Context, path string, config map[string]interface{}) (node.Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return node.Result{
			Success: false,
			Status:  "Error",
			Logs:    []node.LogEntry{{Level: "error", Message: err.Error()}},
		}, err
	}

	format, _ := config["format"].(string)
	var result interface{}

	switch format {
	case "json":
		if err := json.Unmarshal(data, &result); err != nil {
			return node.Result{}, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(data, &result); err != nil {
			return node.Result{}, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		result = string(data)
	}

	return node.Result{
		Success: true,
		Data:    result,
		Status:  fmt.Sprintf("Read %d bytes", len(data)),
		Logs:    []node.LogEntry{{Level: "info", Message: fmt.Sprintf("Read file: %s", path)}},
	}, nil
}

func (n *FileNode) write(ctx node.Context, path string, config map[string]interface{}, appendMode bool) (node.Result, error) {
	content, _ := config["content"].(string)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return node.Result{}, fmt.Errorf("failed to create directory: %w", err)
	}

	var flag int
	if appendMode {
		flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flag = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	}

	f, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return node.Result{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	written, err := f.WriteString(content)
	if err != nil {
		return node.Result{}, fmt.Errorf("failed to write: %w", err)
	}

	op := "Wrote"
	if appendMode {
		op = "Appended"
	}

	return node.Result{
		Success: true,
		Data:    map[string]interface{}{"bytes_written": written},
		Status:  fmt.Sprintf("%s %d bytes", op, written),
		Logs:    []node.LogEntry{{Level: "info", Message: fmt.Sprintf("%s to file: %s", op, path)}},
	}, nil
}

func (n *FileNode) delete(ctx node.Context, path string) (node.Result, error) {
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return node.Result{
				Success: true,
				Data:    map[string]interface{}{"existed": false},
				Status:  "Not found",
				Logs:    []node.LogEntry{{Level: "info", Message: "File did not exist"}},
			}, nil
		}
		return node.Result{}, fmt.Errorf("failed to delete: %w", err)
	}

	return node.Result{
		Success: true,
		Data:    map[string]interface{}{"deleted": true},
		Status:  "Deleted",
		Logs:    []node.LogEntry{{Level: "info", Message: fmt.Sprintf("Deleted: %s", path)}},
	}, nil
}

func (n *FileNode) exists(ctx node.Context, path string) (node.Result, error) {
	_, err := os.Stat(path)
	exists := err == nil

	return node.Result{
		Success: true,
		Data:    map[string]interface{}{"exists": exists},
		Status:  fmt.Sprintf("Exists: %v", exists),
		Logs:    []node.LogEntry{{Level: "info", Message: fmt.Sprintf("Checked: %s (exists=%v)", path, exists)}},
	}, nil
}
