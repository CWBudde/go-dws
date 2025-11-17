package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestOrdinalFunctions tests the CLI with ordinal function scripts
func TestOrdinalFunctions(t *testing.T) {
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
			name:         "Inc and Dec Functions",
			scriptFile:   "../../testdata/ordinal_functions/inc_dec.dws",
			expectedFile: "../../testdata/ordinal_functions/inc_dec.out",
			wantExitCode: 0,
		},
		{
			name:         "Succ and Pred Functions",
			scriptFile:   "../../testdata/ordinal_functions/succ_pred.dws",
			expectedFile: "../../testdata/ordinal_functions/succ_pred.out",
			wantExitCode: 0,
		},
		{
			name:         "Low and High for Enums",
			scriptFile:   "../../testdata/ordinal_functions/low_high_enum.dws",
			expectedFile: "../../testdata/ordinal_functions/low_high_enum.out",
			wantExitCode: 0,
		},
		{
			name:         "For Loop with Enums",
			scriptFile:   "../../testdata/ordinal_functions/for_loop_enum.dws",
			expectedFile: "../../testdata/ordinal_functions/for_loop_enum.out",
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

			// Run the CLI with the script
			cmd := exec.Command(binary, "run", tt.scriptFile)
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

// TestOrdinalFunctionsParsing tests that ordinal function scripts parse correctly
func TestOrdinalFunctionsParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/ordinal_functions/inc_dec.dws",
		"../../testdata/ordinal_functions/succ_pred.dws",
		"../../testdata/ordinal_functions/low_high_enum.dws",
		"../../testdata/ordinal_functions/for_loop_enum.dws",
	}

	for _, script := range scripts {
		t.Run(script, func(t *testing.T) {
			// Check if script exists
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Skipf("Script %s does not exist, skipping", script)
			}

			// Parse the script
			cmd := exec.Command(binary, "parse", script)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Failed to parse %s: %v\nOutput: %s", script, err, output)
			}

			// Check for parser errors in output
			if strings.Contains(string(output), "parse error") {
				t.Errorf("Parser reported errors for %s:\n%s", script, output)
			}
		})
	}
}

// TestOrdinalFunctionsInlineCode tests ordinal functions with inline code
func TestOrdinalFunctionsInlineCode(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	tests := []struct {
		name         string
		code         string
		wantContains []string
	}{
		{
			name: "Inc with integer",
			code: `
var x: Integer := 10;
Inc(x);
PrintLn(x);
			`,
			wantContains: []string{"11"},
		},
		{
			name: "Dec with integer",
			code: `
var x: Integer := 10;
Dec(x);
PrintLn(x);
			`,
			wantContains: []string{"9"},
		},
		{
			name: "Succ with integer",
			code: `
PrintLn(Succ(5));
			`,
			wantContains: []string{"6"},
		},
		{
			name: "Pred with integer",
			code: `
PrintLn(Pred(5));
			`,
			wantContains: []string{"4"},
		},
		{
			name: "Low and High with enum",
			code: `
type TColor = (Red, Green, Blue);
var c: TColor := Green;
PrintLn(Low(c));
PrintLn(High(c));
			`,
			// Enums print as ordinal values by default (use .Name for enum names)
			wantContains: []string{"0", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, "run", "-e", tt.code)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Failed to run code: %v\nOutput: %s", err, output)
			}

			outputStr := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Expected output to contain %q, got:\n%s", want, outputStr)
				}
			}
		})
	}
}
