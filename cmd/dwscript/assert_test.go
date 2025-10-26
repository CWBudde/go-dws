package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestAssertIntegration tests the CLI with Assert function scripts
// Task 9.37-9.39: Create CLI integration tests for Assert function
func TestAssertIntegration(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	tests := []struct {
		name        string
		scriptFile  string
		wantOutputs []string // Strings that must appear in output
		shouldFail  bool     // True if script should exit with error
	}{
		{
			name:       "Reference Assert Test",
			scriptFile: "../../testdata/assert/assert.dws",
			wantOutputs: []string{
				"Assertion failed [line: 7, column: 15]",
				"Assertion failed [line: 15, column: 18] : boom",
				"Assertion failed [line: 23, column: 11]",
				"Assertion failed [line: 31, column: 18] : reboom",
			},
			shouldFail: false,
		},
		{
			name:       "Basic Assert Usage",
			scriptFile: "../../testdata/assert/assert_basic.dws",
			wantOutputs: []string{
				"Test 1 passed",
				"Test 2 passed",
				"Test 3: Caught assertion failure",
				"Math is broken!",
				"All basic tests completed",
			},
			shouldFail: false,
		},
		{
			name:       "Assert for Validation",
			scriptFile: "../../testdata/assert/assert_validation.dws",
			wantOutputs: []string{
				"Valid: 10",
				"Value must be positive",
				"Division: 5",
				"Cannot divide by zero",
				"Validation tests completed",
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if script file exists
			if _, err := os.Stat(tt.scriptFile); os.IsNotExist(err) {
				t.Skipf("Script file not found: %s", tt.scriptFile)
			}

			// Run the script
			cmd := exec.Command(binary, "run", tt.scriptFile)
			output, err := cmd.CombinedOutput()

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected script to fail, but it succeeded")
				}
			} else {
				if err != nil {
					t.Errorf("Script execution failed: %v\nOutput: %s", err, string(output))
				}
			}

			outputStr := string(output)

			// Check that all expected strings appear in output
			for _, want := range tt.wantOutputs {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s", want, outputStr)
				}
			}
		})
	}
}

// TestAssertFailureMessage tests that assertion failures include position information
func TestAssertFailureMessage(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	// Create a simple test script
	script := `
try
  Assert(false);
except
  on E: EAssertionFailed do
    PrintLn(E.Message);
end;
`

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "assert_test_*.dws")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Run the script
	cmd := exec.Command(binary, "run", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Script execution failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Verify the message contains "Assertion failed" and position info
	if !strings.Contains(outputStr, "Assertion failed") {
		t.Errorf("Expected output to contain 'Assertion failed', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "line:") {
		t.Errorf("Expected output to contain 'line:', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "column:") {
		t.Errorf("Expected output to contain 'column:', got: %s", outputStr)
	}
}

// TestAssertWithCustomMessage tests that custom messages are included
func TestAssertWithCustomMessage(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	// Create a test script with custom message
	script := `
try
  Assert(false, 'This is a custom error message');
except
  on E: EAssertionFailed do
    PrintLn(E.Message);
end;
`

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "assert_custom_*.dws")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Run the script
	cmd := exec.Command(binary, "run", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Script execution failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Verify the message contains both position and custom message
	if !strings.Contains(outputStr, "Assertion failed") {
		t.Errorf("Expected output to contain 'Assertion failed', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "This is a custom error message") {
		t.Errorf("Expected output to contain custom message, got: %s", outputStr)
	}
}
