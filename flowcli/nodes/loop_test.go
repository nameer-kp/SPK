package nodes

import (
	"log/slog"
	"os"
	"testing"

	"github.com/nameer-kp/flowcli/pkg/node"
)

// mockContext implements node.Context for testing
type mockContext struct {
	data   map[string]interface{}
	inputs map[string]interface{}
	logger *slog.Logger
}

func newMockContext() *mockContext {
	return &mockContext{
		data:   make(map[string]interface{}),
		inputs: make(map[string]interface{}),
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}
}

func (m *mockContext) Get(key string) (interface{}, bool) {
	v, ok := m.data[key]
	return v, ok
}

func (m *mockContext) Set(key string, value interface{}) {
	m.data[key] = value
}

func (m *mockContext) GetInput(name string) (interface{}, bool) {
	v, ok := m.inputs[name]
	return v, ok
}

func (m *mockContext) GetProfileVar(name string) (string, bool) {
	return "", false
}

func (m *mockContext) GetEnv(name string) string {
	return os.Getenv(name)
}

func (m *mockContext) Logger() *slog.Logger {
	return m.logger
}

func TestLoopNode_Name(t *testing.T) {
	n := NewLoopNode(nil)
	if n.Name() != "loop" {
		t.Errorf("Name() = %q, want %q", n.Name(), "loop")
	}
}

func TestLoopNode_Description(t *testing.T) {
	n := NewLoopNode(nil)
	if n.Description() == "" {
		t.Error("Description() should not be empty")
	}
	expected := "Iterate over an array and execute nested steps"
	if n.Description() != expected {
		t.Errorf("Description() = %q, want %q", n.Description(), expected)
	}
}

func TestLoopNode_ConfigSchema(t *testing.T) {
	n := NewLoopNode(nil)
	schema := n.ConfigSchema()

	if schema == nil {
		t.Fatal("ConfigSchema() should not be nil")
	}

	requiredFields := []string{"items", "as", "index", "steps"}
	for _, field := range requiredFields {
		if _, ok := schema[field]; !ok {
			t.Errorf("ConfigSchema() missing field %q", field)
		}
	}
}

func TestLoopNode_Execute_WithValidArray(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	// Set up array in context
	ctx.data["users"] = []interface{}{"alice", "bob", "charlie"}

	config := map[string]interface{}{
		"items": "users",
		"as":    "user",
		"index": "i",
	}

	result, err := n.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	// Check result data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Execute() Data should be map[string]interface{}")
	}

	count, ok := data["count"].(int)
	if !ok || count != 3 {
		t.Errorf("Execute() count = %v, want 3", data["count"])
	}

	iterations, ok := data["iterations"].([]map[string]interface{})
	if !ok {
		t.Fatal("Execute() iterations should be []map[string]interface{}")
	}

	if len(iterations) != 3 {
		t.Fatalf("Execute() iterations length = %d, want 3", len(iterations))
	}

	// Verify iterations
	expectedItems := []string{"alice", "bob", "charlie"}
	for i, iter := range iterations {
		if iter["index"] != i {
			t.Errorf("iterations[%d].index = %v, want %d", i, iter["index"], i)
		}
		if iter["item"] != expectedItems[i] {
			t.Errorf("iterations[%d].item = %v, want %q", i, iter["item"], expectedItems[i])
		}
	}

	// Check status
	if result.Status != "3 iterations" {
		t.Errorf("Execute() Status = %q, want %q", result.Status, "3 iterations")
	}
}

func TestLoopNode_Execute_WithTemplateExpression(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	// Set up array in context
	ctx.data["items"] = []interface{}{1, 2, 3, 4, 5}

	config := map[string]interface{}{
		"items": "{{items}}",
		"as":    "num",
	}

	result, err := n.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	data := result.Data.(map[string]interface{})
	if data["count"] != 5 {
		t.Errorf("Execute() count = %v, want 5", data["count"])
	}
}

func TestLoopNode_Execute_WithNestedObjectAccess(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	// Set up nested data in context
	ctx.data["api_response"] = map[string]interface{}{
		"data": map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"name": "alice", "age": 30},
				map[string]interface{}{"name": "bob", "age": 25},
			},
		},
	}

	config := map[string]interface{}{
		"items": "{{api_response.data.users}}",
		"as":    "user",
	}

	result, err := n.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	data := result.Data.(map[string]interface{})
	if data["count"] != 2 {
		t.Errorf("Execute() count = %v, want 2", data["count"])
	}

	iterations := data["iterations"].([]map[string]interface{})
	if len(iterations) != 2 {
		t.Fatalf("Execute() iterations length = %d, want 2", len(iterations))
	}

	// Verify first iteration item
	firstItem := iterations[0]["item"].(map[string]interface{})
	if firstItem["name"] != "alice" {
		t.Errorf("iterations[0].item.name = %v, want %q", firstItem["name"], "alice")
	}
}

func TestLoopNode_Execute_SetsContextVariables(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	ctx.data["numbers"] = []interface{}{10, 20, 30}

	config := map[string]interface{}{
		"items": "numbers",
		"as":    "num",
		"index": "idx",
	}

	_, err := n.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	// After loop, context should have the last item and index set
	if ctx.data["num"] != 30 {
		t.Errorf("context num = %v, want 30 (last item)", ctx.data["num"])
	}

	if ctx.data["idx"] != 2 {
		t.Errorf("context idx = %v, want 2 (last index)", ctx.data["idx"])
	}
}

func TestLoopNode_Execute_DefaultAsVariable(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	ctx.data["values"] = []interface{}{"a", "b"}

	config := map[string]interface{}{
		"items": "values",
		// No "as" specified, should default to "item"
	}

	_, err := n.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	// Should use default "item" variable name
	if ctx.data["item"] != "b" {
		t.Errorf("context item = %v, want %q (last item with default name)", ctx.data["item"], "b")
	}
}

func TestLoopNode_Execute_MissingItemsReturnsError(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	// Empty config - no items
	config := map[string]interface{}{}

	_, err := n.Execute(ctx, config)

	if err == nil {
		t.Fatal("Execute() expected error for missing items")
	}

	if err.Error() != "items is required" {
		t.Errorf("Execute() error = %q, want %q", err.Error(), "items is required")
	}
}

func TestLoopNode_Execute_ItemsNotFoundReturnsError(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	// Reference non-existent context variable
	config := map[string]interface{}{
		"items": "nonexistent",
	}

	_, err := n.Execute(ctx, config)

	if err == nil {
		t.Fatal("Execute() expected error for items not found")
	}

	if err.Error() != "items 'nonexistent' not found in context" {
		t.Errorf("Execute() error = %q, want items not found error", err.Error())
	}
}

func TestLoopNode_Execute_NonArrayReturnsError(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	// Set a non-array value
	ctx.data["notarray"] = "this is a string"

	config := map[string]interface{}{
		"items": "notarray",
	}

	_, err := n.Execute(ctx, config)

	if err == nil {
		t.Fatal("Execute() expected error for non-array items")
	}

	expectedPrefix := "items must be an array"
	if len(err.Error()) < len(expectedPrefix) || err.Error()[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Execute() error = %q, want error starting with %q", err.Error(), expectedPrefix)
	}
}

func TestLoopNode_Execute_EmptyArray(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	ctx.data["empty"] = []interface{}{}

	config := map[string]interface{}{
		"items": "empty",
	}

	result, err := n.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	data := result.Data.(map[string]interface{})
	if data["count"] != 0 {
		t.Errorf("Execute() count = %v, want 0", data["count"])
	}

	if result.Status != "0 iterations" {
		t.Errorf("Execute() Status = %q, want %q", result.Status, "0 iterations")
	}
}

func TestLoopNode_Execute_StringArray(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	// []string instead of []interface{}
	ctx.data["names"] = []string{"alice", "bob"}

	config := map[string]interface{}{
		"items": "names",
	}

	result, err := n.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	data := result.Data.(map[string]interface{})
	if data["count"] != 2 {
		t.Errorf("Execute() count = %v, want 2", data["count"])
	}
}

func TestLoopNode_Execute_IntArray(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	// []int instead of []interface{}
	ctx.data["numbers"] = []int{1, 2, 3}

	config := map[string]interface{}{
		"items": "numbers",
	}

	result, err := n.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	data := result.Data.(map[string]interface{})
	if data["count"] != 3 {
		t.Errorf("Execute() count = %v, want 3", data["count"])
	}
}

func TestLoopNode_Execute_MapArray(t *testing.T) {
	n := NewLoopNode(nil)
	ctx := newMockContext()

	// []map[string]interface{} instead of []interface{}
	ctx.data["objects"] = []map[string]interface{}{
		{"id": 1},
		{"id": 2},
	}

	config := map[string]interface{}{
		"items": "objects",
	}

	result, err := n.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	data := result.Data.(map[string]interface{})
	if data["count"] != 2 {
		t.Errorf("Execute() count = %v, want 2", data["count"])
	}
}

func TestLoopNode_Execute_WithRegistry(t *testing.T) {
	// Test that registry is stored correctly (for future nested step execution)
	registry := node.NewRegistry()
	n := NewLoopNode(registry)

	if n.registry != registry {
		t.Error("NewLoopNode() should store the registry")
	}
}

func Test_resolveItemsKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "users"},
		{"{{users}}", "users"},
		{"{{ users }}", "users"},
		{"{{step1.data}}", "step1.data"},
		{"  {{  api.response  }}  ", "api.response"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := resolveItemsKey(tt.input)
			if result != tt.expected {
				t.Errorf("resolveItemsKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func Test_resolveNestedPath(t *testing.T) {
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"value": "deep",
			},
			"array": []interface{}{1, 2, 3},
		},
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
		wantErr  bool
	}{
		{"single level", "level1", data["level1"], false},
		{"two levels", "level1.level2", data["level1"].(map[string]interface{})["level2"], false},
		{"three levels", "level1.level2.value", "deep", false},
		{"array access", "level1.array", []interface{}{1, 2, 3}, false},
		{"not found", "level1.nonexistent", nil, true},
		{"traverse non-object", "level1.level2.value.bad", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveNestedPath(data, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("resolveNestedPath(%q) expected error", tt.path)
				}
				return
			}

			if err != nil {
				t.Errorf("resolveNestedPath(%q) unexpected error: %v", tt.path, err)
				return
			}

			// For slices, compare length since direct comparison won't work
			if slice, ok := tt.expected.([]interface{}); ok {
				resultSlice, ok := result.([]interface{})
				if !ok {
					t.Errorf("resolveNestedPath(%q) result is not a slice", tt.path)
					return
				}
				if len(resultSlice) != len(slice) {
					t.Errorf("resolveNestedPath(%q) slice length = %d, want %d", tt.path, len(resultSlice), len(slice))
				}
				return
			}

			// For maps, just check it's a map (deep comparison is complex)
			if _, ok := tt.expected.(map[string]interface{}); ok {
				if _, ok := result.(map[string]interface{}); !ok {
					t.Errorf("resolveNestedPath(%q) result is not a map", tt.path)
				}
				return
			}

			if result != tt.expected {
				t.Errorf("resolveNestedPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func Test_toSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantLen int
		wantErr bool
	}{
		{"[]interface{}", []interface{}{1, 2, 3}, 3, false},
		{"[]string", []string{"a", "b"}, 2, false},
		{"[]int", []int{1, 2, 3, 4}, 4, false},
		{"[]float64", []float64{1.0, 2.0}, 2, false},
		{"[]map", []map[string]interface{}{{"a": 1}}, 1, false},
		{"string", "not an array", 0, true},
		{"int", 42, 0, true},
		{"map", map[string]interface{}{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toSlice(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("toSlice() expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("toSlice() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("toSlice() length = %d, want %d", len(result), tt.wantLen)
			}
		})
	}
}
