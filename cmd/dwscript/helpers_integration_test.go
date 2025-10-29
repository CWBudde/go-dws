package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelpersScriptsExist verifies all helpers test scripts exist
// Task 9.90: Create integration tests for helpers
func TestHelpersScriptsExist(t *testing.T) {
	scripts := []string{
		"../../testdata/helpers/string_helper.dws",
		"../../testdata/helpers/integer_helper.dws",
		"../../testdata/helpers/record_helper_demo.dws",
		"../../testdata/helpers/class_helper_demo.dws",
		"../../testdata/helpers/multiple_helpers_demo.dws",
		"../../testdata/helpers/class_constants_demo.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Errorf("Required script %s does not exist", script)
			}
		})
	}
}

// TestHelpersParsing tests that all helpers scripts parse correctly
func TestHelpersParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/helpers/string_helper.dws",
		"../../testdata/helpers/integer_helper.dws",
		"../../testdata/helpers/record_helper_demo.dws",
		"../../testdata/helpers/class_helper_demo.dws",
		"../../testdata/helpers/multiple_helpers_demo.dws",
		"../../testdata/helpers/class_constants_demo.dws",
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
		})
	}
}

// TestHelpersExecution tests that helpers scripts execute correctly
func TestHelpersExecution(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	testCases := []struct {
		script         string
		expectedOutput []string // Strings that should appear in output
	}{
		{
			script: "../../testdata/helpers/string_helper.dws",
			expectedOutput: []string{
				"String Helper Demo",
				"IsEmpty: False",
				"ToUpper: UPPER:Hello World",
				"IsEmpty: True",
			},
		},
		{
			script: "../../testdata/helpers/integer_helper.dws",
			expectedOutput: []string{
				"Integer Helper Demo",
				"Number: 42",
				"IsEven: True",
				"Square: 1764",
				"Number: -7",
				"IsNegative: True",
				"Abs: 7",
			},
		},
		{
			script: "../../testdata/helpers/record_helper_demo.dws",
			expectedOutput: []string{
				"Record Helper Demo",
				"Point: (3, 4)",
				"Sum: 7",
				"Distance from origin: 5",
				"IsOrigin: False",
				"IsOrigin: True",
			},
		},
		{
			script: "../../testdata/helpers/class_helper_demo.dws",
			expectedOutput: []string{
				"Class Helper Demo",
				"Alice (age 25)",
				"IsAdult: True",
				"IsTeenager: False",
				"Bob (age 16)",
				"IsAdult: False",
				"IsTeenager: True",
			},
		},
		{
			script: "../../testdata/helpers/multiple_helpers_demo.dws",
			expectedOutput: []string{
				"Multiple Helpers Demo",
				"Number: 5",
				"Double: 10",
				"Triple: 15",
				"IsEven: False",
				"IsOdd: True",
				"Square: 25",
				"Cube: 125",
			},
		},
		{
			script: "../../testdata/helpers/class_constants_demo.dws",
			expectedOutput: []string{
				"Class Constants Demo",
				"Radius: 2",
				"Circumference: 12.56636",
				"Area: 12.56636",
				"Times PI: 31.4159",
				"Times E: 27.1828",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(filepath.Base(tc.script), func(t *testing.T) {
			// Check if script exists
			if _, err := os.Stat(tc.script); os.IsNotExist(err) {
				t.Skipf("Script %s does not exist, skipping", tc.script)
			}

			// Run the script
			cmd := exec.Command(binary, "run", tc.script)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Failed to run %s: %v\nOutput: %s", tc.script, err, output)
				return
			}

			outputStr := string(output)

			// Check for expected strings in output
			for _, expected := range tc.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Output missing expected string %q\nFull output:\n%s", expected, outputStr)
				}
			}

			// Check that there are no runtime errors
			if strings.Contains(outputStr, "Runtime error") || strings.Contains(outputStr, "ERROR:") {
				t.Errorf("Script %s produced runtime errors:\n%s", tc.script, outputStr)
			}
		})
	}
}

// TestHelperMethodDispatch tests that helper methods are properly dispatched
func TestHelperMethodDispatch(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	testCases := []struct {
		name           string
		code           string
		expectedOutput string
	}{
		{
			name: "string_helper_method",
			code: `
type TStringHelper = helper for String
  function Test: String;
  begin
    Result := 'WORKS';
  end;
end;
var s: String;
begin
  s := 'test';
  PrintLn(s.Test());
end.
`,
			expectedOutput: "WORKS",
		},
		{
			name: "integer_helper_method",
			code: `
type TIntHelper = helper for Integer
  function Double: Integer;
  begin
    Result := Self * 2;
  end;
end;
var n: Integer;
begin
  n := 21;
  PrintLn(IntToStr(n.Double()));
end.
`,
			expectedOutput: "42",
		},
		{
			name: "helper_with_self",
			code: `
type TIntHelper = helper for Integer
  function IsPositive: Boolean;
  begin
    Result := Self > 0;
  end;
end;
var n: Integer;
begin
  n := 42;
  PrintLn(BoolToStr(n.IsPositive()));
  n := -5;
  PrintLn(BoolToStr(n.IsPositive()));
end.
`,
			expectedOutput: "True\nFalse",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "run", "-e", tc.code)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Failed to run code: %v\nOutput: %s", err, output)
				return
			}

			outputStr := strings.TrimSpace(string(output))
			expectedStr := strings.TrimSpace(tc.expectedOutput)

			if outputStr != expectedStr {
				t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expectedStr, outputStr)
			}
		})
	}
}

// TestHelperSyntaxVariations tests both "helper for" and "record helper for" syntax
func TestHelperSyntaxVariations(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	testCases := []struct {
		name string
		code string
	}{
		{
			name: "helper_for_syntax",
			code: `
type TStringHelper = helper for String
  function Test: String;
  begin
    Result := 'OK';
  end;
end;
var s: String;
begin
  s := 'test';
  PrintLn(s.Test());
end.
`,
		},
		{
			name: "record_helper_for_syntax",
			code: `
type TPoint = record
  X: Integer;
end;
type TPointHelper = record helper for TPoint
  function GetX: Integer;
  begin
    Result := Self.X;
  end;
end;
var p: TPoint;
begin
  p.X := 42;
  PrintLn(IntToStr(p.GetX()));
end.
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "run", "-e", tc.code)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Failed to run code: %v\nOutput: %s", err, output)
				return
			}

			// Both should produce valid output without errors
			outputStr := string(output)
			if strings.Contains(outputStr, "error") || strings.Contains(outputStr, "ERROR") {
				t.Errorf("Code produced errors:\n%s", outputStr)
			}
		})
	}
}
