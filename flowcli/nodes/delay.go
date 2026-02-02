package nodes

import (
	"fmt"
	"time"

	"github.com/nameer-kp/flowcli/pkg/node"
)

// DelayNode pauses execution for a specified duration
type DelayNode struct{}

func NewDelayNode() *DelayNode {
	return &DelayNode{}
}

func (n *DelayNode) Name() string {
	return "delay"
}

func (n *DelayNode) Description() string {
	return "Pause execution for a specified duration"
}

func (n *DelayNode) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"duration": "string", // e.g., "5s", "1m", "500ms"
	}
}

func (n *DelayNode) Execute(ctx node.Context, config map[string]interface{}) (node.Result, error) {
	durationStr, ok := config["duration"].(string)
	if !ok || durationStr == "" {
		return node.Result{}, fmt.Errorf("duration is required (e.g., '5s', '1m')")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return node.Result{}, fmt.Errorf("invalid duration '%s': %w", durationStr, err)
	}

	ctx.Logger().Info("delaying execution", "duration", duration)
	time.Sleep(duration)

	return node.Result{
		Success: true,
		Data:    map[string]interface{}{"delayed": durationStr},
		Status:  fmt.Sprintf("Delayed %s", durationStr),
		Logs: []node.LogEntry{
			{Level: "info", Message: fmt.Sprintf("Paused for %s", durationStr)},
		},
	}, nil
}
