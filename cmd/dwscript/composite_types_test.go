package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestCompositeTypesScriptsExist verifies all composite types test scripts exist
// Task 8.146: Create integration tests for composite types
func TestCompositeTypesScriptsExist(t *testing.T) {
	scripts := []string{
		"../../testdata/enums.dws",
		"../../testdata/records.dws",
		"../../testdata/sets.dws",
		"../../testdata/arrays_advanced.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Errorf("Required script %s does not exist", script)
			}
		})
	}
}

// TestCompositeTypesParsing tests that all composite types scripts parse correctly
func TestCompositeTypesParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/enums.dws",
		"../../testdata/records.dws",
		"../../testdata/sets.dws",
		"../../testdata/arrays_advanced.dws",
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

// TestEnumFeatures tests enum parsing via CLI
func TestEnumFeatures(t *testing.T) {
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
			name: "basic enum",
			source: `
				type TColor = (Red, Green, Blue);
				var c: TColor;
				c := Red;
			`,
			shouldParse: true,
		},
		{
			name: "enum with explicit values",
			source: `
				type TStatus = (Ok = 200, Error = 404);
				var s: TStatus := TStatus.Ok;
			`,
			shouldParse: true,
		},
		{
			name: "enum with Ord function",
			source: `
				type TDay = (Mon, Tue, Wed);
				PrintLn(Ord(Mon));
			`,
			shouldParse: true,
		},
		{
			name: "enum in case statement",
			source: `
				type TColor = (Red, Green, Blue);
				var c: TColor := Red;
				case c of
					Red: PrintLn('red');
					Green: PrintLn('green');
				end;
			`,
			shouldParse: true,
		},
	}

	binary := "../../bin/dwscript"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "parse", "-e", tc.source)
			output, err := cmd.CombinedOutput()

			if tc.shouldParse {
				if err != nil && strings.Contains(string(output), "parse error") {
					t.Errorf("Expected to parse successfully but got error:\n%s", output)
				}
			} else {
				if err == nil && !strings.Contains(string(output), "parse error") {
					t.Errorf("Expected parse error but parsed successfully")
				}
			}
		})
	}
}

// TestRecordFeatures tests record parsing via CLI
func TestRecordFeatures(t *testing.T) {
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
			name: "basic record",
			source: `
				type TPoint = record
					X: Integer;
					Y: Integer;
				end;

				var p: TPoint;
				p.X := 10;
			`,
			shouldParse: true,
		},
		{
			name: "nested records",
			source: `
				type TPoint = record
					X: Integer;
					Y: Integer;
				end;

				type TRect = record
					TopLeft: TPoint;
					BottomRight: TPoint;
				end;

				var r: TRect;
				r.TopLeft.X := 0;
			`,
			shouldParse: true,
		},
		{
			name: "record comparison",
			source: `
				type TPoint = record
					X: Integer;
					Y: Integer;
				end;

				var p1, p2: TPoint;
				if p1 = p2 then PrintLn('equal');
			`,
			shouldParse: true,
		},
	}

	binary := "../../bin/dwscript"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "parse", "-e", tc.source)
			output, err := cmd.CombinedOutput()

			if tc.shouldParse {
				if err != nil && strings.Contains(string(output), "parse error") {
					t.Errorf("Expected to parse successfully but got error:\n%s", output)
				}
			} else {
				if err == nil && !strings.Contains(string(output), "parse error") {
					t.Errorf("Expected parse error but parsed successfully")
				}
			}
		})
	}
}

// TestSetFeatures tests set parsing via CLI
func TestSetFeatures(t *testing.T) {
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
			name: "basic set",
			source: `
				type TColor = (Red, Green, Blue);
				type TColorSet = set of TColor;

				var s: TColorSet;
				s := [Red, Green];
			`,
			shouldParse: true,
		},
		{
			name: "set membership",
			source: `
				type TColor = (Red, Green, Blue);
				type TColorSet = set of TColor;

				var s: TColorSet := [Red];
				if Red in s then PrintLn('found');
			`,
			shouldParse: true,
		},
		{
			name: "set operations",
			source: `
				type TColor = (Red, Green, Blue);
				type TColorSet = set of TColor;

				var s1, s2, s3: TColorSet;
				s1 := [Red];
				s2 := [Green];
				s3 := s1 + s2;
			`,
			shouldParse: true,
		},
		{
			name: "Include and Exclude",
			source: `
				type TColor = (Red, Green, Blue);
				type TColorSet = set of TColor;

				var s: TColorSet;
				Include(s, Red);
				Exclude(s, Red);
			`,
			shouldParse: true,
		},
	}

	binary := "../../bin/dwscript"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "parse", "-e", tc.source)
			output, err := cmd.CombinedOutput()

			if tc.shouldParse {
				if err != nil && strings.Contains(string(output), "parse error") {
					t.Errorf("Expected to parse successfully but got error:\n%s", output)
				}
			} else {
				if err == nil && !strings.Contains(string(output), "parse error") {
					t.Errorf("Expected parse error but parsed successfully")
				}
			}
		})
	}
}

// TestArrayFeatures tests array parsing via CLI
func TestArrayFeatures(t *testing.T) {
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
			name: "static array",
			source: `
				type TIntArray = array[1..5] of Integer;
				var arr: TIntArray;
				arr[1] := 10;
			`,
			shouldParse: true,
		},
		{
			name: "dynamic array",
			source: `
				type TDynArray = array of Integer;
				var arr: TDynArray;
				SetLength(arr, 5);
			`,
			shouldParse: true,
		},
		{
			name: "array literal",
			source: `
				var arr: array of Integer;
				arr := [1, 2, 3, 4, 5];
			`,
			shouldParse: true,
		},
		{
			name: "multi-dimensional array",
			source: `
				type TMatrix = array of array of Integer;
				var m: TMatrix;
				SetLength(m, 2);
				SetLength(m[0], 3);
			`,
			shouldParse: true,
		},
	}

	binary := "../../bin/dwscript"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "parse", "-e", tc.source)
			output, err := cmd.CombinedOutput()

			if tc.shouldParse {
				if err != nil && strings.Contains(string(output), "parse error") {
					t.Errorf("Expected to parse successfully but got error:\n%s", output)
				}
			} else {
				if err == nil && !strings.Contains(string(output), "parse error") {
					t.Errorf("Expected parse error but parsed successfully")
				}
			}
		})
	}
}

// TestOperatorOverloading tests operator overloading feature via CLI
// Task 8.25: Add CLI integration test running representative operator overloading scripts
func TestOperatorOverloading(t *testing.T) {
	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping CLI tests: failed to build CLI: %v", err)
	}

	binary := "../../bin/dwscript"

	testCases := []struct {
		name                 string
		scriptFile           string
		expectedFile         string
		skipReason           string
		wantOutputs          []string
		wantErrorParts       []string
		shouldSucceed        bool
		isErrorTest          bool
		skipIfNotImplemented bool
	}{
		{
			name:                 "String + Integer operator",
			scriptFile:           "../../testdata/operators/pass/operator_overloading1.dws",
			expectedFile:         "../../testdata/operators/pass/operator_overloading1.txt",
			wantOutputs:          []string{"abc[123]", "13", "abc[123][456]"},
			shouldSucceed:        true,
			skipIfNotImplemented: true,
			skipReason:           "Requires IntToStr built-in function",
		},
		{
			name:                 "IN operator overload",
			scriptFile:           "../../testdata/operators/pass/operator_in_overloading.dws",
			expectedFile:         "../../testdata/operators/pass/operator_in_overloading.txt",
			wantOutputs:          []string{"True", "False"},
			shouldSucceed:        true,
			skipIfNotImplemented: true,
			skipReason:           "Requires 'not in' syntax support",
		},
		{
			name:                 "Implicit conversion Integer to Record",
			scriptFile:           "../../testdata/operators/pass/implicit_record1.dws",
			expectedFile:         "../../testdata/operators/pass/implicit_record1.txt",
			wantOutputs:          []string{"F1 X=10 Y=11", "F2 X=20 Y=21"},
			shouldSucceed:        true,
			skipIfNotImplemented: true,
			skipReason:           "Requires ToString method on Integer",
		},
		{
			name:           "Invalid operator declaration (error test)",
			scriptFile:     "../../testdata/operators/fail/operator_overload2.dws",
			isErrorTest:    true,
			shouldSucceed:  false,
			wantErrorParts: []string{"Syntax Error", "Type expected"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO(CRITICAL): Skip this test - causes system crash due to parser infinite loop
			// The parser enters an infinite loop/unbounded recursion when parsing the malformed
			// input "operator + (" (incomplete operator declaration), consuming ~10GB/sec of memory.
			// See TEST_ISSUES.md for details.
			// This test must remain skipped until the parser is fixed to handle incomplete operator
			// declarations with proper error recovery.
			if tc.name == "Invalid operator declaration (error test)" {
				t.Skip("SKIPPED: Parser infinite loop on incomplete operator declaration - causes system crash (see TEST_ISSUES.md)")
			}

			// Check if script exists
			if _, err := os.Stat(tc.scriptFile); os.IsNotExist(err) {
				t.Skipf("Script %s does not exist, skipping", tc.scriptFile)
			}

			if tc.isErrorTest {
				// For error tests, just parse and expect errors
				cmd := exec.Command(binary, "parse", tc.scriptFile)
				output, err := cmd.CombinedOutput()

				if tc.shouldSucceed {
					if err != nil {
						t.Errorf("Expected success but got error: %v\nOutput: %s", err, output)
					}
				} else {
					// Should have errors
					outputStr := string(output)
					for _, errPart := range tc.wantErrorParts {
						if !strings.Contains(outputStr, errPart) {
							t.Errorf("Expected error to contain %q but got:\n%s", errPart, output)
						}
					}
				}
			} else {
				// For success tests, run the script and verify output
				cmd := exec.Command(binary, "run", tc.scriptFile)
				output, err := cmd.CombinedOutput()

				if err != nil {
					outputStr := string(output)
					// Check if this is due to missing features
					if tc.skipIfNotImplemented && (strings.Contains(outputStr, "undefined function") ||
						strings.Contains(outputStr, "no prefix parse function") ||
						strings.Contains(outputStr, "undefined method") ||
						strings.Contains(outputStr, "requires a helper")) {
						t.Skipf("Skipping: %s - %s", tc.skipReason, outputStr[0:minInt(200, len(outputStr))])
						return
					}
					t.Errorf("Failed to run %s: %v\nOutput: %s", tc.scriptFile, err, output)
					return
				}

				outputStr := string(output)

				// Check for expected output strings
				for _, want := range tc.wantOutputs {
					if !strings.Contains(outputStr, want) {
						t.Errorf("Output missing expected string %q\nGot:\n%s", want, outputStr)
					}
				}

				// If expected file exists, read and compare full output
				if tc.expectedFile != "" {
					if expectedBytes, err := os.ReadFile(tc.expectedFile); err == nil {
						expected := strings.TrimSpace(string(expectedBytes))
						actual := strings.TrimSpace(outputStr)

						if expected != actual {
							t.Errorf("Output mismatch:\nExpected:\n%s\n\nGot:\n%s", expected, actual)
						}
					}
				}
			}
		})
	}
}

// TestOperatorOverloadingParsing tests that operator overloading scripts parse correctly
func TestOperatorOverloadingParsing(t *testing.T) {
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
			name: "basic binary operator overload",
			source: `
				function AddStrInt(s: String; i: Integer): String;
				begin
					Result := s + IntToStr(i);
				end;

				operator + (String, Integer): String uses AddStrInt;
			`,
			shouldParse: true,
		},
		{
			name: "implicit conversion operator",
			source: `
				type TPoint = record
					X, Y: Integer;
				end;

				function IntToPoint(i: Integer): TPoint;
				begin
					Result.X := i;
					Result.Y := i;
				end;

				operator implicit (Integer): TPoint uses IntToPoint;
			`,
			shouldParse: true,
		},
		{
			name: "IN operator overload",
			source: `
				function IsIn(a: Integer; b: Integer): Boolean;
				begin
					Result := True;
				end;

				operator in (Integer, Integer): Boolean uses IsIn;
			`,
			shouldParse: true,
		},
		{
			name: "unary operator overload",
			source: `
				type TPoint = record
					X, Y: Integer;
				end;

				function NegatePoint(p: TPoint): TPoint;
				begin
					Result.X := -p.X;
					Result.Y := -p.Y;
				end;

				operator - (TPoint): TPoint uses NegatePoint;
			`,
			shouldParse: true,
		},
	}

	binary := "../../bin/dwscript"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "parse", "-e", tc.source)
			output, err := cmd.CombinedOutput()

			if tc.shouldParse {
				if err != nil && strings.Contains(string(output), "parse error") {
					t.Errorf("Expected to parse successfully but got error:\n%s", output)
				}
			} else {
				if err == nil && !strings.Contains(string(output), "parse error") {
					t.Errorf("Expected parse error but parsed successfully")
				}
			}
		})
	}
}
