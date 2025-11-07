package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestControlFlowStatements tests the CLI with control flow scripts
func TestControlFlowStatements(t *testing.T) {
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
		wantPasses  int      // Minimum number of "PASS" occurrences
	}{
		{
			name:       "Break Statement",
			scriptFile: "../../testdata/control_flow/break_statement.dws",
			wantOutputs: []string{
				"=== Break Statement Tests ===",
				"Test 1: Break in for loop",
				"PASS: Loop exited after 5 iterations (expected 5)",
				"Test 2: Break in while loop",
				"PASS: Loop exited after 7 iterations (expected 7)",
				"Test 3: Break in repeat loop",
				"PASS: Loop exited after 3 iterations (expected 3)",
				"Test 4: Break in nested loops",
				"PASS: Outer loop ran 3 times (expected 3)",
				"PASS: Inner loop total iterations: 6 (expected 6)",
				"Test 5: Break with condition",
				"PASS: Sum exceeded 50, final value: 55",
				"=== All Break Statement Tests Complete ===",
			},
			wantPasses: 6,
		},
		{
			name:       "Continue Statement",
			scriptFile: "../../testdata/control_flow/continue_statement.dws",
			wantOutputs: []string{
				"=== Continue Statement Tests ===",
				"Test 1: Continue in for loop",
				"PASS: Sum = 50 (expected 50, skipped 5)",
				"Test 2: Continue in while loop",
				"PASS: Sum = 50 (expected 50, skipped 5)",
				"Test 3: Continue in repeat loop",
				"PASS: Sum = 50 (expected 50, skipped 5)",
				"Test 4: Continue with multiple conditions",
				"PASS: Sum of odd numbers = 100 (expected 100)",
				"Test 5: Continue in nested loops",
				"PASS: Count = 12 (expected 12)",
				"=== All Continue Statement Tests Complete ===",
			},
			wantPasses: 5,
		},
		{
			name:       "Exit Statement",
			scriptFile: "../../testdata/control_flow/exit_statement.dws",
			wantOutputs: []string{
				"=== Exit Statement Tests ===",
				"Test 1: Exit terminates function early",
				"GetValueWithExit(-1) = 0 (expected 0)",
				"GetValueWithExit(1) = 20 (expected 20)",
				"PASS: Exit terminated function at correct points",
				"Test 2: Exit preserves Result variable",
				"Calculate(3) = 16 (expected 16)",
				"Calculate(10) = 20 (expected 20)",
				"PASS: Result variable preserved correctly",
				"Test 3: Exit in procedure",
				"Positive value: 10",
				"Positive value: 25",
				"PASS: Procedure exited correctly for non-positive values",
				"Test 4: Exit in nested function calls",
				"Level1() = 123 (expected 123)",
				"PASS: Nested exits worked correctly",
				"Test 5: Exit with complex logic",
				"FindFirst(20) = 7 (expected 7)",
				"FindFirst(5) = -1 (expected -1)",
				"PASS: Exit from loop within function worked correctly",
				"Test 6: Multiple exit points",
				"Classify(-5) = negative (expected negative)",
				"Classify(0) = zero (expected zero)",
				"Classify(10) = positive (expected positive)",
				"PASS: Multiple exit points handled correctly",
				"=== All Exit Statement Tests Complete ===",
			},
			wantPasses: 6,
		},
		{
			name:       "Nested Loops",
			scriptFile: "../../testdata/control_flow/nested_loops.dws",
			wantOutputs: []string{
				"=== Nested Loops Control Flow Tests ===",
				"Test 1: Break only exits innermost loop",
				"Outer loop iterations: 4 (expected 4)",
				"Inner loop total iterations: 12 (expected 12)",
				"PASS: Break only affected innermost loop",
				"Test 2: Continue only affects innermost loop",
				"Total increments: 12 (expected 12)",
				"PASS: Continue only affected innermost loop",
				"Test 3: Triple nested loops with break",
				"Level 1 iterations: 2 (expected 2)",
				"Level 2 iterations: 6 (expected 6)",
				"Level 3 iterations: 12 (expected 12)",
				"PASS: Break in triple nested loop affected only innermost",
				"Test 4: Break and continue combined in nested loops",
				"Total: 12 (expected 12)",
				"PASS: Break and continue worked together correctly",
				"Test 5: Mixed loop types with break/continue",
				"Mixed loops count: 9 (expected 9)",
				"PASS: Break/continue work with mixed loop types",
				"=== All Nested Loops Tests Complete ===",
			},
			wantPasses: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if script file exists
			if _, err := os.Stat(tt.scriptFile); os.IsNotExist(err) {
				t.Fatalf("Script file does not exist: %s", tt.scriptFile)
			}

			// Run the script
			cmd := exec.Command(binary, "run", tt.scriptFile)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to run script: %v\nOutput: %s", err, string(output))
			}

			outputStr := string(output)

			// Check for expected outputs
			for _, want := range tt.wantOutputs {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Output missing expected string:\n  want: %q\n  in output:\n%s", want, outputStr)
				}
			}

			// Count PASS occurrences
			passCount := strings.Count(outputStr, "PASS")
			if passCount < tt.wantPasses {
				t.Errorf("Expected at least %d PASS occurrences, got %d\nOutput:\n%s", tt.wantPasses, passCount, outputStr)
			}
		})
	}
}

// TestControlFlowWithExpectedOutput compares actual output with expected output files
func TestControlFlowWithExpectedOutput(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	tests := []struct {
		name           string
		scriptFile     string
		expectedOutput string
	}{
		{
			name:           "Break Statement",
			scriptFile:     "../../testdata/control_flow/break_statement.dws",
			expectedOutput: "../../testdata/control_flow/break_statement.txt",
		},
		{
			name:           "Continue Statement",
			scriptFile:     "../../testdata/control_flow/continue_statement.dws",
			expectedOutput: "../../testdata/control_flow/continue_statement.txt",
		},
		{
			name:           "Exit Statement",
			scriptFile:     "../../testdata/control_flow/exit_statement.dws",
			expectedOutput: "../../testdata/control_flow/exit_statement.txt",
		},
		{
			name:           "Nested Loops",
			scriptFile:     "../../testdata/control_flow/nested_loops.dws",
			expectedOutput: "../../testdata/control_flow/nested_loops.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if files exist
			if _, err := os.Stat(tt.scriptFile); os.IsNotExist(err) {
				t.Fatalf("Script file does not exist: %s", tt.scriptFile)
			}
			if _, err := os.Stat(tt.expectedOutput); os.IsNotExist(err) {
				t.Fatalf("Expected output file does not exist: %s", tt.expectedOutput)
			}

			// Run the script
			cmd := exec.Command(binary, "run", tt.scriptFile)
			actualOutput, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to run script: %v\nOutput: %s", err, string(actualOutput))
			}

			// Read expected output
			expectedBytes, err := os.ReadFile(tt.expectedOutput)
			if err != nil {
				t.Fatalf("Failed to read expected output file: %v", err)
			}

			expected := string(expectedBytes)
			actual := string(actualOutput)

			// Compare outputs
			if actual != expected {
				t.Errorf("Output mismatch:\n=== EXPECTED ===\n%s\n=== ACTUAL ===\n%s\n=== END ===", expected, actual)
			}
		})
	}
}
