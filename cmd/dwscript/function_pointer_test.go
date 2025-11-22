package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFunctionPointerScriptsExist verifies all function pointer test scripts exist
func TestFunctionPointerScriptsExist(t *testing.T) {
	scripts := []string{
		"../../testdata/function_pointers/basic_function_pointer.dws",
		"../../testdata/function_pointers/callback.dws",
		"../../testdata/function_pointers/method_pointer.dws",
		"../../testdata/function_pointers/sort_with_comparator.dws",
		"../../testdata/function_pointers/procedure_pointer.dws",
		"../../testdata/function_pointers/invalid_cases.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Errorf("Required script %s does not exist", script)
			}
		})
	}
}

// TestFunctionPointerParsing tests that function pointer scripts parse correctly
func TestFunctionPointerParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/function_pointers/basic_function_pointer.dws",
		"../../testdata/function_pointers/callback.dws",
		"../../testdata/function_pointers/method_pointer.dws",
		"../../testdata/function_pointers/sort_with_comparator.dws",
		"../../testdata/function_pointers/procedure_pointer.dws",
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

// TestFunctionPointerFeatures tests function pointer syntax via CLI parse command
// Validates type declarations, address-of operator, and various signatures
func TestFunctionPointerFeatures(t *testing.T) {
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
			name: "basic function pointer type",
			source: `
				type TIntFunc = function(x: Integer): Integer;
			`,
			shouldParse: true,
			description: "Simple function pointer type declaration",
		},
		{
			name: "procedure pointer type",
			source: `
				type TProc = procedure(x: Integer);
			`,
			shouldParse: true,
			description: "Procedure pointer (no return value)",
		},
		{
			name: "method pointer type",
			source: `
				type TMethod = function(x: Integer): Integer of object;
			`,
			shouldParse: true,
			description: "Method pointer with 'of object'",
		},
		{
			name: "address-of operator",
			source: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				type TFunc = function(a, b: Integer): Integer;
				var f: TFunc;
				begin
					f := @Add;
				end.
			`,
			shouldParse: true,
			description: "Address-of operator @ to get function pointer",
		},
		{
			name: "function pointer with no parameters",
			source: `
				type TSimpleFunc = function(): Integer;
			`,
			shouldParse: true,
			description: "Function pointer with no parameters",
		},
		{
			name: "function pointer with multiple parameters",
			source: `
				type TMultiParam = function(a: Integer; b: String; c: Boolean): String;
			`,
			shouldParse: true,
			description: "Function pointer with multiple parameters",
		},
		{
			name: "procedure pointer with no parameters",
			source: `
				type TSimpleProc = procedure();
			`,
			shouldParse: true,
			description: "Procedure pointer with no parameters",
		},
		{
			name: "callback parameter",
			source: `
				type TTransform = function(x: Integer): Integer;

				function Apply(value: Integer; transform: TTransform): Integer;
				begin
					Result := transform(value);
				end;
			`,
			shouldParse: true,
			description: "Function pointer as parameter (callback)",
		},
		{
			name: "nested function pointer types",
			source: `
				type TInner = function(x: Integer): Integer;
				type TOuter = function(f: TInner): Integer;
			`,
			shouldParse: true,
			description: "Function pointer type using another function pointer type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write test source to temporary file
			tmpfile, err := os.CreateTemp("", "funcptr_test_*.dws")
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

// TestFunctionPointerInvalidCases tests error detection for invalid function pointer usage
func TestFunctionPointerInvalidCases(t *testing.T) {
	// Note: These tests verify that invalid function pointer code is properly parsed into AST.
	// Semantic errors (duplicate params, undefined types) are caught by the semantic analyzer,
	// which is tested in internal/semantic/function_pointer_test.go
	// The parser's job is just to build the AST structure, even for semantically invalid code.

	t.Skip("Invalid cases are tested at the semantic analyzer level (internal/semantic/function_pointer_test.go)")

	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skip invalid cases tests: failed to build CLI: %v", err)
	}

	// These test cases parse successfully but fail during semantic analysis
	testCases := []struct {
		name        string
		source      string
		errorType   string
		expectError bool
	}{
		{
			name: "duplicate parameter names",
			source: `
				type TDupParam = function(x: Integer; x: String): Integer;
			`,
			expectError: false, // Parses OK, semantic error
			errorType:   "duplicate parameter (caught by semantic analyzer)",
		},
		{
			name: "undefined parameter type",
			source: `
				type TInvalidType = function(x: NonExistentType): Integer;
			`,
			expectError: false, // Parses OK, semantic error
			errorType:   "undefined type (caught by semantic analyzer)",
		},
		{
			name: "undefined return type",
			source: `
				type TInvalidReturn = function(x: Integer): NonExistentType;
			`,
			expectError: false, // Parses OK, semantic error
			errorType:   "undefined type (caught by semantic analyzer)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write test source to temporary file
			tmpfile, err := os.CreateTemp("", "funcptr_invalid_*.dws")
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

			if tc.expectError {
				// Should either fail or output should contain error message
				hasError := err != nil || strings.Contains(string(output), "error")
				if !hasError {
					t.Errorf("Expected %s error but parse succeeded\nOutput: %s",
						tc.errorType, output)
				}
			}
		})
	}
}

// TestFunctionPointerExecution tests runtime execution of function pointers
// NOTE: This test will be skipped until interpreter support is implemented
func TestFunctionPointerExecution(t *testing.T) {
	// TODO: Implement function pointers
	t.Skip("Function pointer execution not yet implemented in interpreter")

	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping execution tests: failed to build CLI: %v", err)
	}

	// TODO: Once interpreter supports function pointers, add execution tests here
	// The tests should verify that scripts produce expected output
	// Similar to how control_flow_integration_test.go works

	scripts := map[string]string{
		"basic_function_pointer.dws": "../../testdata/function_pointers/basic_function_pointer.txt",
		"callback.dws":               "../../testdata/function_pointers/callback.txt",
		"method_pointer.dws":         "../../testdata/function_pointers/method_pointer.txt",
		"sort_with_comparator.dws":   "../../testdata/function_pointers/sort_with_comparator.txt",
		"procedure_pointer.dws":      "../../testdata/function_pointers/procedure_pointer.txt",
	}

	for script, expectedFile := range scripts {
		t.Run(script, func(t *testing.T) {
			scriptPath := filepath.Join("../../testdata/function_pointers", script)

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
