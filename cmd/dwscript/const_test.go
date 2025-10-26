package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestConstDeclarations tests the CLI with const declaration scripts
// Task 8.260: Create CLI integration tests for const declarations
func TestConstDeclarations(t *testing.T) {
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
			name:       "Basic Const Declarations",
			scriptFile: "../../testdata/const/basic_const.dws",
			wantOutputs: []string{
				"=== Basic Const Declaration Tests ===",
				"Test 1: Integer const",
				"MAX = 100",
				"PASS: MAX equals 100",
				"Test 2: Float const",
				"PI = 3.14",
				"PASS: PI equals 3.14",
				"Test 3: String const",
				"APP_NAME = DWScript",
				"PASS: APP_NAME equals DWScript",
				"Test 4: Boolean const",
				"FLAG = true",
				"PASS: FLAG equals true",
				"=== All Basic Const Tests Complete ===",
			},
			wantPasses: 4,
		},
		{
			name:       "Const Types",
			scriptFile: "../../testdata/const/const_types.dws",
			wantOutputs: []string{
				"=== Const Types Tests ===",
				"Test 1: Typed integer const",
				"MAX_USERS = 1000",
				"PASS: Typed integer const works",
				"Test 2: Typed float const",
				"RADIUS = 5",
				"PASS: Typed float const works",
				"Test 3: Typed string const",
				"VERSION = v1.0.0",
				"PASS: Typed string const works",
				"Test 4: Typed boolean const",
				"DEBUG = false",
				"PASS: Typed boolean const works",
				"Test 5: Multiple consts",
				"Range = 100",
				"PASS: Multiple consts work together",
				"=== All Const Types Tests Complete ===",
			},
			wantPasses: 5,
		},
		{
			name:       "Const Expressions",
			scriptFile: "../../testdata/const/const_expressions.dws",
			wantOutputs: []string{
				"=== Const Expression Tests ===",
				"Test 1: Arithmetic with consts",
				"PASS: Const arithmetic correct",
				"Test 2: Const in variable initializer",
				"result = FACTOR * 2 = 20",
				"PASS: Const in initializer works",
				"Test 3: Multiple consts in expression",
				"sum = X + Y + Z = 18",
				"PASS: Multiple consts in expression works",
				"Test 4: Const with expression value",
				"CALCULATED = 10 + 20 * 2 = 50",
				"PASS: Const with expression value works",
				"Test 5: Const in comparison",
				"PASS: Const in comparison works",
				"Test 6: String const concatenation",
				"message = Hello World",
				"PASS: String const concatenation works",
				"=== All Const Expression Tests Complete ===",
			},
			wantPasses: 6,
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
				t.Fatalf("Failed to run script %s: %v\nOutput: %s", tt.scriptFile, err, string(output))
			}

			outputStr := string(output)

			// Check for required output strings
			for _, want := range tt.wantOutputs {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Output missing expected string:\nWant: %q\nGot output:\n%s", want, outputStr)
				}
			}

			// Count PASS occurrences
			passCount := strings.Count(outputStr, "PASS")
			if passCount < tt.wantPasses {
				t.Errorf("Expected at least %d PASS occurrences, got %d\nOutput:\n%s",
					tt.wantPasses, passCount, outputStr)
			}
		})
	}
}
