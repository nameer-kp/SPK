package engine

import (
	"log/slog"
	"os"
	"testing"

	"github.com/nameer-kp/flowcli/internal/config"
)

func TestEvaluator(t *testing.T) {
	// Create a test context
	cfg := &config.ResolvedConfig{
		Profile: config.Profile{
			Name: "test",
			Variables: map[string]string{
				"api_url":     "https://api.example.com",
				"environment": "staging",
			},
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	ctx := NewExecutionContext(cfg, logger)

	// Set up some step outputs
	ctx.Set("step1", map[string]interface{}{
		"success": true,
		"count":   42,
		"message": "hello",
		"data": map[string]interface{}{
			"nested": "value",
			"number": 100,
		},
	})
	ctx.Set("step2", map[string]interface{}{
		"enabled": false,
		"value":   10,
	})

	// Set up inputs
	ctx.SetInputs(map[string]interface{}{
		"name":     "test-user",
		"count":    5,
		"enabled":  true,
		"disabled": false,
	})

	// Set environment variable for testing
	os.Setenv("TEST_ENV_VAR", "test-value")
	defer os.Unsetenv("TEST_ENV_VAR")

	evaluator := NewEvaluator(ctx)

	tests := []struct {
		name       string
		condition  string
		want       bool
		wantErr    bool
		errContain string
	}{
		// Empty condition returns true
		{
			name:      "empty condition returns true",
			condition: "",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "whitespace only condition returns true",
			condition: "   ",
			want:      true,
			wantErr:   false,
		},

		// Simple true/false literals
		{
			name:      "literal true",
			condition: "true",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "literal false",
			condition: "false",
			want:      false,
			wantErr:   false,
		},

		// Step output access - boolean
		{
			name:      "step output boolean true",
			condition: "step1.success == true",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "step output boolean false",
			condition: "step2.enabled == false",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "step output boolean direct access",
			condition: "step1.success",
			want:      true,
			wantErr:   false,
		},

		// Step output access - number comparison
		{
			name:      "step output number equals",
			condition: "step1.count == 42",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "step output number not equals",
			condition: "step1.count != 0",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "step output number greater than",
			condition: "step1.count > 10",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "step output number less than",
			condition: "step2.value < 20",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "step output number greater or equal",
			condition: "step1.count >= 42",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "step output number less or equal",
			condition: "step2.value <= 10",
			want:      true,
			wantErr:   false,
		},

		// Step output access - string comparison
		{
			name:      "step output string equals",
			condition: `step1.message == "hello"`,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "step output string not equals",
			condition: `step1.message != "goodbye"`,
			want:      true,
			wantErr:   false,
		},

		// Nested step output access
		{
			name:      "nested step output access",
			condition: `step1.data.nested == "value"`,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "nested step output number comparison",
			condition: "step1.data.number > 50",
			want:      true,
			wantErr:   false,
		},

		// Input access
		{
			name:      "input string equals",
			condition: `inputs.name == "test-user"`,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "input number comparison",
			condition: "inputs.count > 3",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "input boolean true",
			condition: "inputs.enabled == true",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "input boolean direct",
			condition: "inputs.enabled",
			want:      true,
			wantErr:   false,
		},

		// Profile variable access
		{
			name:      "profile variable equals",
			condition: `profile.environment == "staging"`,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "profile variable not equals",
			condition: `profile.environment != "production"`,
			want:      true,
			wantErr:   false,
		},

		// Environment variable access
		{
			name:      "env variable equals",
			condition: `env.TEST_ENV_VAR == "test-value"`,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "env variable not equals",
			condition: `env.TEST_ENV_VAR != "other"`,
			want:      true,
			wantErr:   false,
		},

		// Boolean operators - AND
		{
			name:      "and operator both true",
			condition: "step1.success && inputs.enabled",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "and operator one false",
			condition: "step1.success && step2.enabled",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "and operator both false",
			condition: "step2.enabled && inputs.disabled",
			want:      false,
			wantErr:   false,
		},

		// Boolean operators - OR
		{
			name:      "or operator both true",
			condition: "step1.success || inputs.enabled",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "or operator one true",
			condition: "step1.success || step2.enabled",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "or operator both false",
			condition: "step2.enabled || inputs.disabled",
			want:      false,
			wantErr:   false,
		},

		// Boolean operators - NOT
		{
			name:      "not operator on true",
			condition: "!step2.enabled",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "not operator on false",
			condition: "!step1.success",
			want:      false,
			wantErr:   false,
		},

		// Complex expressions
		{
			name:      "complex expression with multiple operators",
			condition: "(step1.success && step1.count > 40) || step2.enabled",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "complex expression with nested access",
			condition: `step1.data.number > 50 && profile.environment == "staging"`,
			want:      true,
			wantErr:   false,
		},

		// Error cases
		{
			name:       "invalid expression syntax",
			condition:  "step1.success &&",
			want:       false,
			wantErr:    true,
			errContain: "invalid expression",
		},
		{
			name:       "non-boolean result - number",
			condition:  "step1.count",
			want:       false,
			wantErr:    true,
			errContain: "non-boolean",
		},
		{
			name:       "non-boolean result - string",
			condition:  "step1.message",
			want:       false,
			wantErr:    true,
			errContain: "non-boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluator.Evaluate(tt.condition)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Evaluate() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContain != "" && !contains(err.Error(), tt.errContain) {
					t.Errorf("Evaluate() error = %v, want error containing %q", err, tt.errContain)
				}
				return
			}
			if err != nil {
				t.Errorf("Evaluate() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_MissingData(t *testing.T) {
	cfg := &config.ResolvedConfig{
		Profile: config.Profile{
			Name:      "test",
			Variables: map[string]string{},
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	ctx := NewExecutionContext(cfg, logger)
	evaluator := NewEvaluator(ctx)

	// Test 1: Undefined variable by itself is nil
	result, err := evaluator.Evaluate("nonexistent == nil")
	if err != nil {
		t.Errorf("Evaluate('nonexistent == nil') unexpected error = %v", err)
	}
	if !result {
		t.Errorf("Evaluate('nonexistent == nil') = %v, want true", result)
	}

	// Test 2: Accessing field on nil/undefined causes an error
	// This is expected behavior - user should ensure step exists before accessing fields
	_, err = evaluator.Evaluate("nonexistent.field == nil")
	if err == nil {
		t.Errorf("Evaluate('nonexistent.field') expected error for accessing field on nil")
	}

	// Test 3: Missing input returns nil
	result, err = evaluator.Evaluate("inputs.missing_input == nil")
	if err != nil {
		t.Errorf("Evaluate('inputs.missing_input == nil') unexpected error = %v", err)
	}
	if !result {
		t.Errorf("Evaluate('inputs.missing_input == nil') = %v, want true", result)
	}

	// Test 4: Missing profile var returns nil
	result, err = evaluator.Evaluate("profile.missing_var == nil")
	if err != nil {
		t.Errorf("Evaluate('profile.missing_var == nil') unexpected error = %v", err)
	}
	if !result {
		t.Errorf("Evaluate('profile.missing_var == nil') = %v, want true", result)
	}

	// Test 5: Missing env var returns empty string
	result, err = evaluator.Evaluate(`env.NONEXISTENT_VAR_12345 == ""`)
	if err != nil {
		t.Errorf("Evaluate('env.NONEXISTENT_VAR == \"\"') unexpected error = %v", err)
	}
	if !result {
		t.Errorf("Evaluate('env.NONEXISTENT_VAR == \"\"') = %v, want true (empty string)", result)
	}
}

// contains checks if substr is in s
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
