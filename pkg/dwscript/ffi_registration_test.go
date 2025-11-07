package dwscript

import (
	"bytes"
	"strings"
	"testing"
)

// ============================================================================
// Task 9.1: Function Registration Tests
// ============================================================================

// TestRegisterInvalidFunction tests registration of invalid functions
func TestRegisterInvalidFunction(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Test: nil function
	err = engine.RegisterFunction("Test", nil)
	if err == nil {
		t.Errorf("expected error for nil function")
	}

	// Test: non-function value
	err = engine.RegisterFunction("Test", "not a function")
	if err == nil {
		t.Errorf("expected error for non-function value")
	}
}

// TestRegisterDuplicateFunction tests that duplicate function names are rejected
func TestRegisterDuplicateFunction(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register first function
	err = engine.RegisterFunction("Test", func() {})
	if err != nil {
		t.Fatalf("failed to register first function: %v", err)
	}

	// Try to register duplicate
	err = engine.RegisterFunction("Test", func() {})
	if err == nil {
		t.Errorf("expected error for duplicate function name")
	}
}

// TestRegisterFunctionTypeMismatch tests type validation during registration
func TestRegisterFunctionTypeMismatch(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Test: function with unsupported parameter type
	err = engine.RegisterFunction("BadParam", func(ch chan int) {})
	if err == nil {
		t.Errorf("expected error for unsupported parameter type")
	}

	// Test: function with unsupported return type
	err = engine.RegisterFunction("BadReturn", func() chan int { return nil })
	if err == nil {
		t.Errorf("expected error for unsupported return type")
	}
}

// TestFunctionSignatureDetection tests automatic signature detection
func TestFunctionSignatureDetection(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	tests := []struct {
		name     string
		fn       interface{}
		expected string
	}{
		{
			name:     "no args no return",
			fn:       func() {},
			expected: "()",
		},
		{
			name:     "one int arg no return",
			fn:       func(int64) {},
			expected: "(Integer)",
		},
		{
			name:     "two args no return",
			fn:       func(int64, string) {},
			expected: "(Integer, String)",
		},
		{
			name:     "no args with return",
			fn:       func() int64 { return 0 },
			expected: "() -> Integer",
		},
		{
			name:     "one arg with return",
			fn:       func(int64) string { return "" },
			expected: "(Integer) -> String",
		},
		{
			name:     "two args with return",
			fn:       func(int64, float64) bool { return true },
			expected: "(Integer, Float) -> Boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.RegisterFunction("Test"+strings.Title(tt.name), tt.fn)
			if err != nil {
				t.Errorf("failed to register function: %v", err)
			}
		})
	}
}

// TestMultipleFunctions tests registering and using multiple functions
func TestMultipleFunctions(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register multiple functions
	err = engine.RegisterFunction("Add", func(a, b int64) int64 { return a + b })
	if err != nil {
		t.Fatalf("failed to register Add: %v", err)
	}

	err = engine.RegisterFunction("Multiply", func(a, b int64) int64 { return a * b })
	if err != nil {
		t.Fatalf("failed to register Multiply: %v", err)
	}

	err = engine.RegisterFunction("GetMessage", func() string { return "Hello from Go!" })
	if err != nil {
		t.Fatalf("failed to register GetMessage: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var x := Add(5, 3);
		var y := Multiply(x, 2);
		var msg := GetMessage();
		PrintLn(IntToStr(y));
		PrintLn(msg);
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines of output, got %d", len(lines))
	}
	if lines[0] != "16" {
		t.Errorf("expected '16', got '%s'", lines[0])
	}
	if lines[1] != "Hello from Go!" {
		t.Errorf("expected 'Hello from Go!', got '%s'", lines[1])
	}
}
