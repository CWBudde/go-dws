package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestNewArrayExpressions tests the CLI with new array instantiation
// Tasks 9.167-9.168: CLI integration tests for array instantiation with new keyword
func TestNewArrayExpressions(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	tests := []struct {
		name         string
		scriptFile   string
		expectedFile string
		wantExitCode int
	}{
		{
			name:         "Basic 1D Array Creation",
			scriptFile:   "../../testdata/new_array/new_array_basic.dws",
			expectedFile: "../../testdata/new_array/new_array_basic.expected",
			wantExitCode: 0,
		},
		{
			name:         "Multi-dimensional Arrays",
			scriptFile:   "../../testdata/new_array/new_array_multidim.dws",
			expectedFile: "../../testdata/new_array/new_array_multidim.expected",
			wantExitCode: 0,
		},
		{
			name:         "Expression-based Sizes",
			scriptFile:   "../../testdata/new_array/new_array_expressions.dws",
			expectedFile: "../../testdata/new_array/new_array_expressions.expected",
			wantExitCode: 0,
		},
		{
			name:         "Various Element Types",
			scriptFile:   "../../testdata/new_array/new_array_types.dws",
			expectedFile: "../../testdata/new_array/new_array_types.expected",
			wantExitCode: 0,
		},
		{
			name:         "Levenshtein Distance (Working)",
			scriptFile:   "../../testdata/new_array/levenshtein_working.dws",
			expectedFile: "../../testdata/new_array/levenshtein_working.expected",
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the script
			cmd := exec.Command(binary, "run", tt.scriptFile)
			output, err := cmd.CombinedOutput()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("Failed to execute script: %v", err)
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.wantExitCode, exitCode)
				t.Logf("Output:\n%s", output)
				return
			}

			// Read expected output
			expectedBytes, err := os.ReadFile(tt.expectedFile)
			if err != nil {
				t.Fatalf("Failed to read expected output file: %v", err)
			}
			expected := string(expectedBytes)

			// Compare outputs (normalize line endings)
			actualStr := strings.ReplaceAll(string(output), "\r\n", "\n")
			expectedStr := strings.ReplaceAll(expected, "\r\n", "\n")

			if actualStr != expectedStr {
				t.Errorf("Output mismatch\nExpected:\n%s\nGot:\n%s", expectedStr, actualStr)
			}
		})
	}
}
