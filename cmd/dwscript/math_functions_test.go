package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMathFunctions tests the CLI with math function scripts
// Task 9.66: CLI integration tests for math functions (Min, Max, Sqr, Power, Ceil, Floor, RandomInt)
func TestMathFunctions(t *testing.T) {
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
			name:         "Min and Max Functions",
			scriptFile:   "../../testdata/math_functions/min_max.dws",
			expectedFile: "../../testdata/math_functions/min_max.expected",
			wantExitCode: 0,
		},
		{
			name:         "Sqr and Power Functions",
			scriptFile:   "../../testdata/math_functions/sqr_power.dws",
			expectedFile: "../../testdata/math_functions/sqr_power.expected",
			wantExitCode: 0,
		},
		{
			name:         "Ceil and Floor Functions",
			scriptFile:   "../../testdata/math_functions/ceil_floor.dws",
			expectedFile: "../../testdata/math_functions/ceil_floor.expected",
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

// TestRandomIntFunction tests RandomInt separately since it produces random output
// Task 9.66: CLI integration test for RandomInt (special case - random values)
func TestRandomIntFunction(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"
	scriptFile := "../../testdata/math_functions/random_int.dws"

	// Check if script file exists
	if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
		t.Skipf("Script file not found: %s", scriptFile)
	}

	// Run the CLI with the script
	cmd := exec.Command(binary, "run", scriptFile)
	output, err := cmd.CombinedOutput()

	// Check exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Output:\n%s",
			exitCode, string(output))
	}

	outputStr := string(output)

	// Verify expected patterns in output (values may vary but format should match)
	expectedPatterns := []string{
		"RandomInt(10) =",
		"RandomInt(100) =",
		"RandomInt(5) =",
		"RandomInt(1) = 0", // This one is deterministic
		"All values in range: true",
		"Valid random values generated: 10/10",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(outputStr, pattern) {
			t.Errorf("Expected output to contain %q, got:\n%s", pattern, outputStr)
		}
	}
}

// TestMathFunctionsParsing tests that math function scripts parse correctly
func TestMathFunctionsParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/math_functions/min_max.dws",
		"../../testdata/math_functions/sqr_power.dws",
		"../../testdata/math_functions/ceil_floor.dws",
		"../../testdata/math_functions/random_int.dws",
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

// TestMathFunctionsInlineCode tests math functions with inline code
func TestMathFunctionsInlineCode(t *testing.T) {
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
			name: "Min with integers",
			code: `
begin
	PrintLn(Min(5, 10));
end
			`,
			wantContains: []string{"5"},
		},
		{
			name: "Max with integers",
			code: `
begin
	PrintLn(Max(5, 10));
end
			`,
			wantContains: []string{"10"},
		},
		{
			name: "Sqr with integer",
			code: `
begin
	PrintLn(Sqr(5));
end
			`,
			wantContains: []string{"25"},
		},
		{
			name: "Power",
			code: `
begin
	PrintLn(Power(2, 8));
end
			`,
			wantContains: []string{"256"},
		},
		{
			name: "Ceil",
			code: `
begin
	PrintLn(Ceil(3.2));
end
			`,
			wantContains: []string{"4"},
		},
		{
			name: "Floor",
			code: `
begin
	PrintLn(Floor(3.8));
end
			`,
			wantContains: []string{"3"},
		},
		{
			name: "RandomInt in range",
			code: `
var r := RandomInt(10);
begin
	if (r >= 0) and (r < 10) then
		PrintLn('PASS')
	else
		PrintLn('FAIL');
end
			`,
			wantContains: []string{"PASS"},
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
