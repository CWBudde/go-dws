package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestLambdaScriptsExist verifies all lambda test scripts exist
func TestLambdaScriptsExist(t *testing.T) {
	scripts := []string{
		"../../testdata/lambdas/basic_lambda.dws",
		"../../testdata/lambdas/closure.dws",
		"../../testdata/lambdas/higher_order.dws",
		"../../testdata/lambdas/nested_lambda.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Errorf("Required script %s does not exist", script)
			}
		})
	}
}

// TestLambdaParsing tests that lambda scripts parse correctly
// This validates AST, parser, and semantic analysis for lambdas
func TestLambdaParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/lambdas/basic_lambda.dws",
		"../../testdata/lambdas/closure.dws",
		"../../testdata/lambdas/higher_order.dws",
		"../../testdata/lambdas/nested_lambda.dws",
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

// TestLambdaFeatures tests lambda syntax via CLI parse command
// Validates shorthand/full syntax, closures, and various signatures
func TestLambdaFeatures(t *testing.T) {
	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping CLI tests: failed to build CLI: %v", err)
	}

	testCases := []struct {
		name        string
		source      string
		description string
		shouldParse bool
	}{
		{
			name: "shorthand lambda syntax",
			source: `
				var double := lambda(x: Integer): Integer => x * 2;
			`,
			shouldParse: true,
			description: "Lambda with shorthand syntax (=>)",
		},
		{
			name: "full lambda syntax",
			source: `
				var add := lambda(a: Integer; b: Integer): Integer begin
					Result := a + b;
				end;
			`,
			shouldParse: true,
			description: "Lambda with full begin/end syntax",
		},
		{
			name: "lambda with no parameters",
			source: `
				var getValue := lambda(): Integer => 42;
			`,
			shouldParse: true,
			description: "Lambda with no parameters",
		},
		{
			name: "lambda with multiple parameters",
			source: `
				var sum3 := lambda(x: Integer; y: Integer; z: Integer): Integer => x + y + z;
			`,
			shouldParse: true,
			description: "Lambda with three parameters",
		},
		{
			name: "lambda with String type",
			source: `
				var greet := lambda(name: String): String => 'Hello, ' + name;
			`,
			shouldParse: true,
			description: "Lambda with String parameter and return",
		},
		{
			name: "lambda with Boolean return",
			source: `
				var isEven := lambda(n: Integer): Boolean => (n mod 2) = 0;
			`,
			shouldParse: true,
			description: "Lambda with Boolean return type",
		},
		{
			name: "procedure lambda",
			source: `
				var printMsg := lambda(msg: String) begin
					PrintLn(msg);
				end;
			`,
			shouldParse: true,
			description: "Procedure lambda (no return value)",
		},
		{
			name: "lambda with control flow",
			source: `
				var abs := lambda(n: Integer): Integer begin
					if n < 0 then
						Result := -n
					else
						Result := n;
				end;
			`,
			shouldParse: true,
			description: "Lambda with if/then/else",
		},
		{
			name: "lambda capturing variable",
			source: `
				var x := 10;
				var captureX := lambda(): Integer => x;
			`,
			shouldParse: true,
			description: "Lambda capturing outer variable (closure)",
		},
		// Inline function pointer types not yet supported (need type aliases)
		// {
		// 	name: "nested lambda",
		// 	source: `
		// 		var makeAdder := lambda(x: Integer): function(y: Integer): Integer begin
		// 			Result := lambda(y: Integer): Integer => x + y;
		// 		end;
		// 	`,
		// 	shouldParse: true,
		// 	description: "Lambda returning lambda (nested)",
		// },
		// Array of TypeName syntax not yet supported
		// {
		// 	name: "lambda with higher-order function",
		// 	source: `
		// 		var numbers: array of Integer := [1, 2, 3];
		// 		var doubled := Map(numbers, lambda(x: Integer): Integer => x * 2);
		// 	`,
		// 	shouldParse: true,
		// 	description: "Lambda passed to Map function",
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write test source to temporary file
			tmpfile, err := os.CreateTemp("", "lambda_test_*.dws")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tc.source)); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			// Run parse command
			cmd := exec.Command("../../bin/dwscript", "parse", tmpfile.Name())
			output, err := cmd.CombinedOutput()

			if tc.shouldParse {
				if err != nil {
					t.Errorf("%s: Expected to parse successfully but got error: %v\nOutput: %s",
						tc.description, err, output)
				}
				if strings.Contains(string(output), "error") {
					t.Errorf("%s: Parse succeeded but output contains errors:\n%s",
						tc.description, output)
				}
			} else {
				if err == nil && !strings.Contains(string(output), "error") {
					t.Errorf("%s: Expected parse to fail but it succeeded\nOutput: %s",
						tc.description, output)
				}
			}
		})
	}
}

// TestLambdaExecution tests runtime execution of lambda expressions
// Validates interpreter support for lambdas, closures, and higher-order functions
func TestLambdaExecution(t *testing.T) {
	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping execution tests: failed to build CLI: %v", err)
	}

	scripts := map[string]string{
		"basic_lambda.dws": "../../testdata/lambdas/basic_lambda.txt",
		"closure.dws":      "../../testdata/lambdas/closure.txt",
		// "higher_order.dws" skipped - requires dynamic array literals which aren't implemented
		"nested_lambda.dws": "../../testdata/lambdas/nested_lambda.txt",
	}

	for script, expectedFile := range scripts {
		t.Run(script, func(t *testing.T) {
			scriptPath := filepath.Join("../../testdata/lambdas", script)

			// Check if script exists
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				t.Skipf("Script %s does not exist, skipping", scriptPath)
			}

			// Check if expected output exists
			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				t.Skipf("Expected output %s does not exist, skipping", expectedFile)
			}

			// Run the script
			cmd := exec.Command("../../bin/dwscript", "run", scriptPath)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Failed to run %s: %v\nOutput: %s", scriptPath, err, output)
				return
			}

			// Read expected output
			expectedBytes, err := os.ReadFile(expectedFile)
			if err != nil {
				t.Fatalf("Failed to read expected output: %v", err)
			}

			expected := strings.TrimSpace(string(expectedBytes))
			actual := strings.TrimSpace(string(output))

			if actual != expected {
				t.Errorf("Output mismatch for %s\nExpected:\n%s\n\nActual:\n%s", script, expected, actual)
			}
		})
	}
}
