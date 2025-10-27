package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestArrayFunctions tests the CLI with array function scripts
// Tasks 9.79-9.80: CLI integration tests for array functions (Copy, IndexOf, Contains, Reverse, Sort)
func TestArrayFunctions(t *testing.T) {
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
			name:         "Copy Function",
			scriptFile:   "../../testdata/array_functions/copy.dws",
			expectedFile: "../../testdata/array_functions/copy.out",
			wantExitCode: 0,
		},
		{
			name:         "Search Functions (IndexOf and Contains)",
			scriptFile:   "../../testdata/array_functions/search.dws",
			expectedFile: "../../testdata/array_functions/search.out",
			wantExitCode: 0,
		},
		{
			name:         "Reverse Function",
			scriptFile:   "../../testdata/array_functions/reverse.dws",
			expectedFile: "../../testdata/array_functions/reverse.out",
			wantExitCode: 0,
		},
		{
			name:         "Sort Function",
			scriptFile:   "../../testdata/array_functions/sort.dws",
			expectedFile: "../../testdata/array_functions/sort.out",
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if script file exists
			if _, err := os.Stat(tt.scriptFile); os.IsNotExist(err) {
				t.Skipf("Script file not found: %s", tt.scriptFile)
			}

			// Check if expected output file exists
			if _, err := os.Stat(tt.expectedFile); os.IsNotExist(err) {
				t.Skipf("Expected output file not found: %s", tt.expectedFile)
			}

			// Read expected output
			expectedBytes, err := os.ReadFile(tt.expectedFile)
			if err != nil {
				t.Fatalf("Failed to read expected output file: %v", err)
			}
			expectedOutput := string(expectedBytes)

			// Run the CLI with the script (disable type checking for now)
			cmd := exec.Command(binary, "run", "--type-check=false", tt.scriptFile)
			output, err := cmd.CombinedOutput()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Expected exit code %d, got %d. Output:\n%s",
					tt.wantExitCode, exitCode, string(output))
			}

			// Compare actual output with expected output
			actualOutput := strings.TrimSpace(string(output))
			expectedOutput = strings.TrimSpace(expectedOutput)

			if actualOutput != expectedOutput {
				t.Errorf("Output mismatch:\n=== Expected ===\n%s\n\n=== Got ===\n%s\n",
					expectedOutput, actualOutput)
			}
		})
	}
}

// TestArrayFunctionsParsing tests that array function scripts parse correctly
func TestArrayFunctionsParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/array_functions/copy.dws",
		"../../testdata/array_functions/search.dws",
		"../../testdata/array_functions/reverse.dws",
		"../../testdata/array_functions/sort.dws",
	}

	for _, scriptFile := range scripts {
		t.Run(scriptFile, func(t *testing.T) {
			// Check if script file exists
			if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
				t.Skipf("Script file not found: %s", scriptFile)
			}

			// Try to parse the script
			cmd := exec.Command(binary, "parse", scriptFile)
			output, err := cmd.CombinedOutput()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if exitCode != 0 {
				t.Errorf("Parse failed with exit code %d. Output:\n%s",
					exitCode, string(output))
			}
		})
	}
}
