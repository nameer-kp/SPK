package nodes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nameer-kp/flowcli/pkg/node"
)

// HTTPNode executes HTTP requests
type HTTPNode struct{}

func NewHTTPNode() *HTTPNode {
	return &HTTPNode{}
}

func (n *HTTPNode) Name() string {
	return "http"
}

func (n *HTTPNode) Description() string {
	return "Execute HTTP requests"
}

func (n *HTTPNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"method":  "string",
		"url":     "string",
		"headers": "object",
		"body":    "any",
		"timeout": "string",
	}
}

func (n *HTTPNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	// Extract config values
	method, _ := config["method"].(string)
	if method == "" {
		method = "GET"
	}

	url, ok := config["url"].(string)
	if !ok || url == "" {
		return node.Result{}, fmt.Errorf("url is required")
	}

	// Build request body
	var bodyReader io.Reader
	if body := config["body"]; body != nil {
		var bodyBytes []byte
		var err error

		switch b := body.(type) {
		case string:
			bodyBytes = []byte(b)
		case map[string]interface{}, []interface{}:
			bodyBytes, err = json.Marshal(b)
			if err != nil {
				return node.Result{}, fmt.Errorf("failed to marshal body: %w", err)
			}
		default:
			bodyBytes, err = json.Marshal(b)
			if err != nil {
				return node.Result{}, fmt.Errorf("failed to marshal body: %w", err)
			}
		}

		bodyReader = bytes.NewReader(bodyBytes)
		ctx.Logger().Debug("request body", "body", string(bodyBytes))
	}

	// Create request
	req, err := http.NewRequest(strings.ToUpper(method), url, bodyReader)
	if err != nil {
		return node.Result{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}

	// Default content-type for POST/PUT with body
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set timeout
	timeout := 30 * time.Second
	if t, ok := config["timeout"].(string); ok {
		if parsed, err := time.ParseDuration(t); err == nil {
			timeout = parsed
		}
	}

	client := &http.Client{Timeout: timeout}

	// Execute request
	ctx.Logger().Info("executing HTTP request", "method", method, "url", url)
	resp, err := client.Do(req)
	if err != nil {
		return node.Result{
			Success: false,
			Status:  "Error",
			Logs: []node.LogEntry{
				{Level: "error", Message: err.Error()},
			},
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return node.Result{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response as JSON if possible
	var data interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		// If not JSON, return as string
		data = string(respBody)
	}

	status := fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return node.Result{
		Success: success,
		Data:    data,
		Status:  status,
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Response: %s", status)},
		},
	}, nil
}
