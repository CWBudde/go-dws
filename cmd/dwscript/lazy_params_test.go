package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestLazyParamsScriptsExist verifies all lazy parameter test scripts exist
func TestLazyParamsScriptsExist(t *testing.T) {
	scripts := []string{
		"../../testdata/lazy_params/jensens_device.dws",
		"../../testdata/lazy_params/conditional_eval.dws",
		"../../testdata/lazy_params/lazy_logging.dws",
		"../../testdata/lazy_params/multiple_access.dws",
		"../../testdata/lazy_params/lazy_with_loops.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Errorf("Required script %s does not exist", script)
			}

			// Also check for expected output file
			outFile := strings.TrimSuffix(script, ".dws") + ".out"
			if _, err := os.Stat(outFile); os.IsNotExist(err) {
				t.Errorf("Expected output file %s does not exist", outFile)
			}
		})
	}
}

// TestLazyParamsParsing tests that lazy parameter scripts parse correctly
// This validates AST, parser, and semantic analysis for lazy parameters
func TestLazyParamsParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/lazy_params/jensens_device.dws",
		"../../testdata/lazy_params/conditional_eval.dws",
		"../../testdata/lazy_params/lazy_logging.dws",
		"../../testdata/lazy_params/multiple_access.dws",
		"../../testdata/lazy_params/lazy_with_loops.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
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

			// Check for semantic errors in output
			if strings.Contains(string(output), "semantic error") {
				t.Errorf("Semantic analyzer reported errors for %s:\n%s", script, output)
			}
		})
	}
}

// TestLazyParamsExecution validates that lazy parameter scripts produce expected output
func TestLazyParamsExecution(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	testCases := []struct {
		script   string
		outFile  string
		testName string
	}{
		{
			script:   "../../testdata/lazy_params/jensens_device.dws",
			outFile:  "../../testdata/lazy_params/jensens_device.out",
			testName: "Jensen's Device",
		},
		{
			script:   "../../testdata/lazy_params/conditional_eval.dws",
			outFile:  "../../testdata/lazy_params/conditional_eval.out",
			testName: "Conditional Evaluation",
		},
		{
			script:   "../../testdata/lazy_params/lazy_logging.dws",
			outFile:  "../../testdata/lazy_params/lazy_logging.out",
			testName: "Lazy Logging",
		},
		{
			script:   "../../testdata/lazy_params/multiple_access.dws",
			outFile:  "../../testdata/lazy_params/multiple_access.out",
			testName: "Multiple Access",
		},
		{
			script:   "../../testdata/lazy_params/lazy_with_loops.dws",
			outFile:  "../../testdata/lazy_params/lazy_with_loops.out",
			testName: "Lazy with Loops",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// Check if script and expected output exist
			if _, err := os.Stat(tc.script); os.IsNotExist(err) {
				t.Skipf("Script %s does not exist, skipping", tc.script)
			}
			if _, err := os.Stat(tc.outFile); os.IsNotExist(err) {
				t.Skipf("Expected output %s does not exist, skipping", tc.outFile)
			}

			// Run the script
			cmd := exec.Command(binary, "run", tc.script)
			actualOutput, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("Failed to run %s: %v\nOutput: %s", tc.script, err, actualOutput)
			}

			// Read expected output
			expectedBytes, err := os.ReadFile(tc.outFile)
			if err != nil {
				t.Fatalf("Failed to read expected output from %s: %v", tc.outFile, err)
			}
			expected := string(expectedBytes)

			// Compare outputs
			if string(actualOutput) != expected {
				t.Errorf("Output mismatch for %s\nExpected:\n%s\n\nActual:\n%s",
					tc.script, expected, string(actualOutput))
			}
		})
	}
}

// TestJensensDevice tests the canonical Jensen's Device example
// This is the classic use case for lazy parameters
func TestJensensDevice(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"
	script := "../../testdata/lazy_params/jensens_device.dws"

	if _, err := os.Stat(script); os.IsNotExist(err) {
		t.Skipf("Jensen's Device script does not exist, skipping")
	}

	cmd := exec.Command(binary, "run", script)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Failed to run Jensen's Device: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Verify harmonic series computation
	if !strings.Contains(outputStr, "Harmonic series H(5)") {
		t.Error("Output should contain harmonic series H(5) computation")
	}

	// Verify the result is approximately correct (2.283333...)
	if !strings.Contains(outputStr, "2.28") {
		t.Error("Harmonic series H(5) should be approximately 2.28")
	}

	// Verify sum of squares
	if !strings.Contains(outputStr, "Sum of squares 1-5 = 55") {
		t.Error("Sum of squares should be 55")
	}

	// Verify H(10)
	if !strings.Contains(outputStr, "Harmonic series H(10)") {
		t.Error("Output should contain harmonic series H(10) computation")
	}
}

// TestLazyEvaluationCount verifies re-evaluation on each access
// Lazy parameters should be evaluated each time they're accessed, not once
func TestLazyEvaluationCount(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"
	script := "../../testdata/lazy_params/multiple_access.dws"

	if _, err := os.Stat(script); os.IsNotExist(err) {
		t.Skipf("Multiple access script does not exist, skipping")
	}

	cmd := exec.Command(binary, "run", script)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Failed to run multiple access test: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Verify that GetNext is called 3 times (once per access)
	count := strings.Count(outputStr, "GetNext called")
	if count != 3 {
		t.Errorf("GetNext should be called 3 times, was called %d times", count)
	}

	// Verify counter increments: 1, 2, 3
	if !strings.Contains(outputStr, "counter = 1") {
		t.Error("First call should have counter = 1")
	}
	if !strings.Contains(outputStr, "counter = 2") {
		t.Error("Second call should have counter = 2")
	}
	if !strings.Contains(outputStr, "counter = 3") {
		t.Error("Third call should have counter = 3")
	}

	// Verify final result is 1 + 2 + 3 = 6
	if !strings.Contains(outputStr, "Result: 6") {
		t.Error("Final result should be 6 (1 + 2 + 3)")
	}
}

// TestLazyConditional verifies lazy expressions not evaluated when skipped
// This tests that lazy parameters provide true conditional evaluation
func TestLazyConditional(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"
	script := "../../testdata/lazy_params/conditional_eval.dws"

	if _, err := os.Stat(script); os.IsNotExist(err) {
		t.Skipf("Conditional evaluation script does not exist, skipping")
	}

	cmd := exec.Command(binary, "run", script)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Failed to run conditional evaluation test: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// In test 1 (condition is true), only ExpensiveTrue should be called
	if !strings.Contains(outputStr, "ExpensiveTrue called (count: 1)") {
		t.Error("When condition is true, ExpensiveTrue should be called exactly once")
	}

	// Total evaluations in test 1 should be 1 (only true branch)
	lines := strings.Split(outputStr, "\n")
	foundTest1 := false
	for i, line := range lines {
		if strings.Contains(line, "=== Test 1: Condition is true ===") {
			foundTest1 = true
		}
		if foundTest1 && strings.Contains(line, "Total evaluations: 1") {
			// Found the correct result
			break
		}
		if foundTest1 && strings.Contains(line, "Total evaluations:") && !strings.Contains(line, "Total evaluations: 1") {
			t.Errorf("Test 1 should have 1 evaluation, found: %s", lines[i])
		}
	}

	// In test 2 (condition is false), only ExpensiveFalse should be called
	if !strings.Contains(outputStr, "ExpensiveFalse called (count: 1)") {
		t.Error("When condition is false, ExpensiveFalse should be called exactly once")
	}
}

// TestLazyParameterSyntax tests various lazy parameter syntax via CLI
// Validates parser support for lazy keyword in different contexts
func TestLazyParameterSyntax(t *testing.T) {
	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping CLI tests: failed to build CLI: %v", err)
	}

	binary := "../../bin/dwscript"

	testCases := []struct {
		name        string
		source      string
		shouldParse bool
		description string
	}{
		{
			name: "basic lazy parameter",
			source: `
				function Test(lazy x: Integer): Integer;
				begin
					Result := x + 1;
				end;
			`,
			shouldParse: true,
			description: "Function with single lazy parameter",
		},
		{
			name: "mixed lazy and regular parameters",
			source: `
				procedure Log(level: Integer; lazy msg: String);
				begin
					PrintLn(msg);
				end;
			`,
			shouldParse: true,
			description: "Mixed lazy and regular parameters",
		},
		{
			name: "multiple lazy parameters",
			source: `
				function IfThen(cond: Boolean; lazy trueVal, falseVal: Integer): Integer;
				begin
					if cond then Result := trueVal else Result := falseVal;
				end;
			`,
			shouldParse: true,
			description: "Multiple lazy parameters with shared type",
		},
		{
			name: "lazy with var should fail",
			source: `
				function Bad(lazy var x: Integer): Integer;
				begin
					Result := x;
				end;
			`,
			shouldParse: false,
			description: "Lazy and var are mutually exclusive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use eval mode to test inline code
			cmd := exec.Command(binary, "parse", "-e", tc.source)
			output, err := cmd.CombinedOutput()

			if tc.shouldParse {
				if err != nil {
					t.Errorf("%s: Expected to parse successfully, but got error: %v\nOutput: %s",
						tc.description, err, output)
				}
				if strings.Contains(string(output), "error") {
					t.Errorf("%s: Parse succeeded but output contains errors:\n%s",
						tc.description, output)
				}
			} else {
				// Should fail to parse
				if err == nil && !strings.Contains(string(output), "error") {
					t.Errorf("%s: Expected parse to fail, but it succeeded",
						tc.description)
				}
			}
		})
	}
}
