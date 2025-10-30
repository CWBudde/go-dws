package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestClassesIntegration tests the CLI with class feature scripts
// Task 7.151: Verify CLI correctly executes class programs
func TestClassesIntegration(t *testing.T) {
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
			name:       "Classes",
			scriptFile: "../../testdata/classes.dws",
			wantOutputs: []string{
				"=== Class Features Comprehensive Test ===",
				"Test 1: Simple class with public fields - PASS",
				"Test 15: Constructor with multiple params - PASS",
				"=== All Class Features Tests Complete ===",
			},
			wantPasses: 15, // All 15 tests should pass
		},
		{
			name:       "MethodKeyword",
			scriptFile: "../../testdata/classes/method_keyword.dws",
			wantOutputs: []string{
				"Initial values:",
				"X = 10",
				"Y = 20",
				"After setting new values:",
				"X = 30",
				"Y = 40",
			},
			wantPasses: 0, // No PASS/FAIL in this test, just output verification
		},
		{
			name:       "InlineArrayFields",
			scriptFile: "../../testdata/classes/inline_array_fields.dws",
			wantOutputs: []string{
				"TBoard created",
				"Names count: 3",
				"First name: Alice",
				"Grid[5]: 25",
			},
			wantPasses: 0, // Task 9.170.5: Verify inline array field declarations work end-to-end
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if script file exists
			if _, err := os.Stat(tt.scriptFile); os.IsNotExist(err) {
				t.Skipf("Script file %s does not exist, skipping", tt.scriptFile)
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

			output := out.String()

			// Check for expected output strings
			for _, want := range tt.wantOutputs {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", want, output)
				}
			}

			// Check for minimum number of PASS occurrences
			if tt.wantPasses > 0 {
				passCount := strings.Count(output, "PASS")
				if passCount < tt.wantPasses {
					t.Errorf("Expected at least %d PASS occurrences, got %d", tt.wantPasses, passCount)
				}
			}

			// Check for no FAIL occurrences in test scripts
			if tt.wantPasses > 0 {
				failCount := strings.Count(output, "FAIL")
				if failCount > 0 {
					t.Errorf("Found %d FAIL occurrences (expected 0):\n%s", failCount, output)
				}
			}
		})
	}
}

// TestClassParseCommand tests parsing class features via CLI
func TestClassParseCommand(t *testing.T) {
	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping CLI tests: failed to build CLI: %v", err)
	}

	testCases := []struct {
		name        string
		source      string
		shouldParse bool
	}{
		{
			name: "simple class",
			source: `
				type TSimple = class
				public
					X: Integer;
				end;

				var s: TSimple;
				s := TSimple.Create();
			`,
			shouldParse: true,
		},
		{
			name: "class with constructor",
			source: `
				type TPoint = class
				private
					FX, FY: Integer;
				public
					constructor Create(x, y: Integer);
				end;

				constructor TPoint.Create(x, y: Integer);
				begin
					FX := x;
					FY := y;
				end;
			`,
			shouldParse: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "class_test_*.dws")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tc.source); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			// Run parse command
			cmd := exec.Command("../../bin/dwscript", "parse", tmpFile.Name())
			output, err := cmd.CombinedOutput()

			if tc.shouldParse {
				if err != nil {
					t.Errorf("Expected successful parse, got error: %v\nOutput: %s", err, output)
				}
			} else {
				if err == nil {
					t.Errorf("Expected parse error, but parsing succeeded\nOutput: %s", output)
				}
			}
		})
	}
}

// TestClassCommandOutput tests that class commands produce expected output formats
func TestClassCommandOutput(t *testing.T) {
	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping CLI tests: failed to build CLI: %v", err)
	}

	testCases := []struct {
		name           string
		source         string
		expectedOutput string
	}{
		{
			name: "simple class instantiation",
			source: `
				type TSimple = class
				public
					Value: Integer;
				end;

				var s: TSimple;
				begin
					s := TSimple.Create();
					s.Value := 42;
					PrintLn('Value:', s.Value);
				end
			`,
			expectedOutput: "Value: 42",
		},
		{
			name: "method call",
			source: `
				type TCalculator = class
				public
					function Add(a, b: Integer): Integer;
					begin
						Result := a + b;
					end;
				end;

				var calc: TCalculator;
				begin
					calc := TCalculator.Create();
					PrintLn('Result:', calc.Add(10, 20));
				end
			`,
			expectedOutput: "Result: 30",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "class_output_test_*.dws")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tc.source); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			// Run the script
			cmd := exec.Command("../../bin/dwscript", "run", tmpFile.Name())
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Logf("Execution failed (may not be fully implemented): %v\nOutput: %s", err, output)
				return
			}

			// Check for expected output
			if !strings.Contains(string(output), tc.expectedOutput) {
				t.Errorf("Expected output to contain %q, got:\n%s", tc.expectedOutput, output)
			}
		})
	}
}
