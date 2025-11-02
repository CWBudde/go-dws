package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestForStepIntegration tests the CLI with for loop step scripts
func TestForStepIntegration(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	tests := []struct {
		name       string
		scriptFile string
		outputFile string
	}{
		{
			name:       "Basic Step",
			scriptFile: "../../testdata/for_step/basic_step.dws",
			outputFile: "../../testdata/for_step/basic_step.out",
		},
		{
			name:       "Step Expressions",
			scriptFile: "../../testdata/for_step/step_expressions.dws",
			outputFile: "../../testdata/for_step/step_expressions.out",
		},
		{
			name:       "Step Validation",
			scriptFile: "../../testdata/for_step/step_validation.dws",
			outputFile: "../../testdata/for_step/step_validation.out",
		},
		{
			name:       "Lucas-Lehmer Test",
			scriptFile: "../../testdata/for_step/lucas_lehmer.dws",
			outputFile: "../../testdata/for_step/lucas_lehmer.out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if script file exists
			if _, err := os.Stat(tt.scriptFile); os.IsNotExist(err) {
				t.Skipf("Script file %s does not exist, skipping", tt.scriptFile)
			}

			// Check if output file exists
			if _, err := os.Stat(tt.outputFile); os.IsNotExist(err) {
				t.Skipf("Output file %s does not exist, skipping", tt.outputFile)
			}

			// Run the script
			cmd := exec.Command(binary, "run", tt.scriptFile)
			var out bytes.Buffer
			var errOut bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &errOut

			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to run %s: %v\nStderr: %s", tt.scriptFile, err, errOut.String())
			}

			// Read expected output
			expectedBytes, err := os.ReadFile(tt.outputFile)
			if err != nil {
				t.Fatalf("Failed to read expected output file %s: %v", tt.outputFile, err)
			}
			expected := string(expectedBytes)

			// Compare output
			actual := out.String()
			if actual != expected {
				t.Errorf("Output mismatch for %s\nExpected:\n%s\nActual:\n%s", tt.scriptFile, expected, actual)
			}
		})
	}
}

// TestForStepScriptsExist verifies all for step test scripts exist
func TestForStepScriptsExist(t *testing.T) {
	scripts := []string{
		"../../testdata/for_step/basic_step.dws",
		"../../testdata/for_step/step_expressions.dws",
		"../../testdata/for_step/step_validation.dws",
		"../../testdata/for_step/lucas_lehmer.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Errorf("Required script %s does not exist", script)
			}
		})
	}
}

// TestForStepParsing tests that all for step scripts parse correctly
func TestForStepParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/for_step/basic_step.dws",
		"../../testdata/for_step/step_expressions.dws",
		"../../testdata/for_step/step_validation.dws",
		"../../testdata/for_step/lucas_lehmer.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			// Check if script exists
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Skipf("Script %s does not exist, skipping", script)
			}

			// Parse the script
			cmd := exec.Command(binary, "parse", script)
			var errOut bytes.Buffer
			cmd.Stderr = &errOut

			if err := cmd.Run(); err != nil {
				t.Errorf("Failed to parse %s: %v\nStderr: %s", script, err, errOut.String())
			}

			// Check stderr for parser errors
			if errOut.Len() > 0 && strings.Contains(errOut.String(), "error") {
				t.Errorf("Parser reported errors for %s:\n%s", script, errOut.String())
			}
		})
	}
}

// TestForStepErrorCases tests error cases for invalid step values
func TestForStepErrorCases(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	tests := []struct {
		name        string
		script      string
		expectError string
	}{
		{
			name:        "Zero Step",
			script:      "for i := 1 to 10 step 0 do PrintLn(i);",
			expectError: "for loop step must be strictly positive",
		},
		{
			name:        "Negative Step",
			script:      "for i := 1 to 10 step -1 do PrintLn(i);",
			expectError: "for loop step must be strictly positive",
		},
		{
			name:        "Negative Step Variable",
			script:      "var s := -5; for i := 1 to 10 step s do PrintLn(i);",
			expectError: "FOR loop STEP should be strictly positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the script
			cmd := exec.Command(binary, "run", "-e", tt.script)
			var out bytes.Buffer
			var errOut bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &errOut

			// We expect this to fail
			err := cmd.Run()
			if err == nil {
				t.Errorf("Expected error for script %q, but it succeeded", tt.script)
				return
			}

			// Check that the error message contains expected text
			combinedOutput := out.String() + errOut.String()
			if !strings.Contains(combinedOutput, tt.expectError) {
				t.Errorf("Expected error message to contain %q, but got:\n%s", tt.expectError, combinedOutput)
			}
		})
	}
}
