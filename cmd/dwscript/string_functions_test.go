package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestStringFunctions tests the CLI with string function scripts
// Task 9.53: CLI integration tests for string functions (Trim, Insert, Delete, StringReplace)
func TestStringFunctions(t *testing.T) {
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
			name:         "Trim Functions",
			scriptFile:   "../../testdata/string_functions/trim.dws",
			expectedFile: "../../testdata/string_functions/trim.out",
			wantExitCode: 0,
		},
		{
			name:         "Insert and Delete Functions",
			scriptFile:   "../../testdata/string_functions/insert_delete.dws",
			expectedFile: "../../testdata/string_functions/insert_delete.out",
			wantExitCode: 0,
		},
		{
			name:         "StringReplace Function",
			scriptFile:   "../../testdata/string_functions/replace.dws",
			expectedFile: "../../testdata/string_functions/replace.out",
			wantExitCode: 0,
		},
		{
			name:         "Format Function",
			scriptFile:   "../../testdata/string_functions/format.dws",
			expectedFile: "../../testdata/string_functions/format.out",
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

// TestStringFunctionsParsing tests that string function scripts parse correctly
func TestStringFunctionsParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/string_functions/trim.dws",
		"../../testdata/string_functions/insert_delete.dws",
		"../../testdata/string_functions/replace.dws",
		"../../testdata/string_functions/format.dws",
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

// TestStringFunctionsInlineCode tests string functions with inline code
func TestStringFunctionsInlineCode(t *testing.T) {
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
			name: "Trim with spaces",
			code: `
var s: String := '  hello  ';
PrintLn('[' + Trim(s) + ']');
			`,
			wantContains: []string{"[hello]"},
		},
		{
			name: "TrimLeft",
			code: `
var s: String := '  hello';
PrintLn('[' + TrimLeft(s) + ']');
			`,
			wantContains: []string{"[hello]"},
		},
		{
			name: "TrimRight",
			code: `
var s: String := 'hello  ';
PrintLn('[' + TrimRight(s) + ']');
			`,
			wantContains: []string{"[hello]"},
		},
		{
			name: "Insert",
			code: `
var s: String := 'Helo';
Insert('l', s, 3);
PrintLn(s);
			`,
			wantContains: []string{"Hello"},
		},
		{
			name: "Delete",
			code: `
var s: String := 'Hello';
Delete(s, 3, 2);
PrintLn(s);
			`,
			wantContains: []string{"Heo"},
		},
		{
			name: "StringReplace",
			code: `
var s: String := 'hello world';
PrintLn(StringReplace(s, 'l', 'L'));
			`,
			wantContains: []string{"heLLo worLd"},
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
