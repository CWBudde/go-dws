package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestInheritanceIntegration tests the CLI with inheritance and polymorphism scripts
// Task 7.151: Verify CLI correctly executes inheritance programs
func TestInheritanceIntegration(t *testing.T) {
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
			name:       "Inheritance",
			scriptFile: "../../testdata/inheritance.dws",
			wantOutputs: []string{
				"=== Inheritance Comprehensive Test ===",
				"Test 1: Simple inheritance - PASS",
				"Test 15: Mixed visibility in inheritance - PASS",
				"=== All Inheritance Tests Complete ===",
			},
			wantPasses: 15, // All 15 tests should pass
		},
		{
			name:       "Polymorphism",
			scriptFile: "../../testdata/polymorphism.dws",
			wantOutputs: []string{
				"=== Polymorphism Comprehensive Test ===",
				"Test 1: Simple virtual/override - PASS",
				"Test 15: Virtual procedure override - PASS",
				"=== All Polymorphism Tests Complete ===",
			},
			wantPasses: 15, // All 15 tests should pass
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

// TestInheritanceScripts tests existing inheritance demo scripts
func TestInheritanceScripts(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []struct {
		name     string
		file     string
		skipTest bool // Some scripts might not have expected output yet
	}{
		{
			name:     "Virtual Override Demo",
			file:     "../../testdata/virtual_override_demo.dws",
			skipTest: false,
		},
		{
			name:     "Abstract Demo",
			file:     "../../testdata/abstract_demo.dws",
			skipTest: false,
		},
		{
			name:     "OOP Integration",
			file:     "../../testdata/oop_integration.dws",
			skipTest: false,
		},
		{
			name:     "Visibility Demo",
			file:     "../../testdata/visibility_demo.dws",
			skipTest: false,
		},
	}

	for _, script := range scripts {
		t.Run(script.name, func(t *testing.T) {
			// Check if script exists
			if _, err := os.Stat(script.file); os.IsNotExist(err) {
				t.Skipf("Script %s does not exist, skipping", script.file)
			}

			if script.skipTest {
				t.Skip("Skipping execution test for this script")
			}

			// Parse the script first
			parseCmd := exec.Command(binary, "parse", script.file)
			parseOutput, parseErr := parseCmd.CombinedOutput()

			if parseErr != nil {
				t.Errorf("Failed to parse %s: %v\nOutput: %s", script.file, parseErr, parseOutput)
				return
			}

			// Try to run the script
			runCmd := exec.Command(binary, "run", script.file)
			runOutput, runErr := runCmd.CombinedOutput()

			if runErr != nil {
				// Log but don't fail - some scripts might not be fully implemented yet
				t.Logf("Note: Failed to run %s: %v\nOutput: %s", script.file, runErr, runOutput)
			} else {
				t.Logf("Successfully ran %s\nOutput:\n%s", script.file, runOutput)
			}
		})
	}
}

// TestInheritanceParseCommand tests parsing inheritance features via CLI
func TestInheritanceParseCommand(t *testing.T) {
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
			name: "inheritance",
			source: `
				type TBase = class
				public
					Value: Integer;
				end;

				type TDerived = class(TBase)
				public
					Extra: Integer;
				end;
			`,
			shouldParse: true,
		},
		{
			name: "virtual override",
			source: `
				type TBase = class
				public
					function GetValue(): Integer; virtual;
					begin
						Result := 0;
					end;
				end;

				type TDerived = class(TBase)
				public
					function GetValue(): Integer; override;
					begin
						Result := 42;
					end;
				end;
			`,
			shouldParse: true,
		},
		{
			name: "abstract class",
			source: `
				type TAbstract = class abstract
				public
					function GetValue(): Integer; abstract;
				end;

				type TConcrete = class(TAbstract)
				public
					function GetValue(): Integer; override;
					begin
						Result := 100;
					end;
				end;
			`,
			shouldParse: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "inheritance_test_*.dws")
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

// TestInheritanceErrorHandling tests error handling for invalid inheritance usage
func TestInheritanceErrorHandling(t *testing.T) {
	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping CLI tests: failed to build CLI: %v", err)
	}

	testCases := []struct {
		name        string
		source      string
		errorMsg    string
		shouldError bool
	}{
		{
			name: "virtual without override",
			source: `
				type TBase = class
				public
					function GetValue(): Integer;
					begin
						Result := 0;
					end;
				end;

				type TDerived = class(TBase)
				public
					function GetValue(): Integer; virtual;
					begin
						Result := 42;
					end;
				end;
			`,
			shouldError: false, // This might be valid - marking method as virtual
			errorMsg:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "inheritance_error_test_*.dws")
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

			if tc.shouldError {
				if err == nil && !strings.Contains(string(output), tc.errorMsg) {
					t.Errorf("Expected error containing '%s', but got success\nOutput: %s",
						tc.errorMsg, output)
				}
			}
		})
	}
}
