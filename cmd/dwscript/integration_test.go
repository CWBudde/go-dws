package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestControlFlowIntegration tests the CLI with control flow scripts
func TestControlFlowIntegration(t *testing.T) {
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
			name:       "If/Else Statements",
			scriptFile: "../../testdata/if_else.dws",
			wantOutputs: []string{
				"=== If/Else Comprehensive Test ===",
				"Test 1: Simple if (true condition)",
				"Test 18: Deeply nested if (3 levels)",
				"=== All If/Else Tests Complete ===",
			},
			wantPasses: 18, // All 18 tests should pass
		},
		{
			name:       "While Loops",
			scriptFile: "../../testdata/while_loop.dws",
			wantOutputs: []string{
				"=== While Loop Comprehensive Test ===",
				"Test 1: Count from 1 to 5",
				"Test 15: Single iteration loop",
				"=== All While Loop Tests Complete ===",
			},
			wantPasses: 15, // All 15 tests should pass
		},
		{
			name:       "For Loops",
			scriptFile: "../../testdata/for_loop.dws",
			wantOutputs: []string{
				"=== For Loop Comprehensive Test ===",
				"Test 1: Basic ascending loop",
				"Test 18: Downto with negatives",
				"=== All For Loop Tests Complete ===",
			},
			wantPasses: 18, // All 18 tests should pass
		},
		{
			name:       "Nested Loops",
			scriptFile: "../../testdata/nested_loops.dws",
			wantOutputs: []string{
				"=== Nested Loops Comprehensive Test ===",
				"Test 1: For inside For (2D grid)",
				"Test 18: Alternating loop directions",
				"=== All Nested Loop Tests Complete ===",
			},
			wantPasses: 18, // All 18 tests should pass
		},
		{
			name:       "If Demo",
			scriptFile: "../../testdata/if_demo.dws",
			wantOutputs: []string{
				"x is greater than 5",
				"x equals 10",
				"x is greater than 0 and greater than 5",
				"Line 1 in block",
				"Line 2 in block",
				"Done!",
			},
			wantPasses: 0, // Demo doesn't have PASS markers
		},
		{
			name:       "While Demo",
			scriptFile: "../../testdata/while_demo.dws",
			wantOutputs: []string{
				"Counting from 1 to 5:",
				"Sum of 1 to 10:",
				"55",
				"Countdown:",
				"Blast off!",
			},
			wantPasses: 0, // Demo doesn't have PASS markers
		},
		{
			name:       "For Demo",
			scriptFile: "../../testdata/for_demo.dws",
			wantOutputs: []string{
				"=== For Loop Demo ===",
				"Example 1: Count from 1 to 5",
				"Example 5: Multiplication table (3x3)",
				"=== Demo Complete ===",
			},
			wantPasses: 0, // Demo doesn't have PASS markers
		},
		{
			name:       "Repeat Demo",
			scriptFile: "../../testdata/repeat_demo.dws",
			wantOutputs: []string{
				"=== Repeat-Until Loop Demo ===",
				"Example 1: Count from 1 to 5",
				"Example 4: Boolean flag control",
				"=== Demo Complete ===",
			},
			wantPasses: 0, // Demo doesn't have PASS markers
		},
		{
			name:       "Case Demo",
			scriptFile: "../../testdata/case_demo.dws",
			wantOutputs: []string{
				"=== Case Statement Examples ===",
				"1. Day of week example:",
				"Wednesday",
				"2. Language greeting:",
				"Bonjour!",
				"=== All tests complete ===",
			},
			wantPasses: 0, // Demo doesn't have PASS markers
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

// TestControlFlowScriptsExist verifies all control flow test scripts exist
func TestControlFlowScriptsExist(t *testing.T) {
	scripts := []string{
		"../../testdata/if_else.dws",
		"../../testdata/while_loop.dws",
		"../../testdata/for_loop.dws",
		"../../testdata/nested_loops.dws",
		"../../testdata/if_demo.dws",
		"../../testdata/while_demo.dws",
		"../../testdata/for_demo.dws",
		"../../testdata/repeat_demo.dws",
		"../../testdata/case_demo.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Errorf("Required script %s does not exist", script)
			}
		})
	}
}

// TestControlFlowParsing tests that all control flow scripts parse correctly
func TestControlFlowParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/if_else.dws",
		"../../testdata/while_loop.dws",
		"../../testdata/for_loop.dws",
		"../../testdata/nested_loops.dws",
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
