package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
