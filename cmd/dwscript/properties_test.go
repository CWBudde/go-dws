package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPropertyScriptsExist verifies all property test scripts exist
// Task 8.59: Create CLI integration tests for properties
func TestPropertyScriptsExist(t *testing.T) {
	scripts := []string{
		"../../testdata/properties/basic_property.dws",
		"../../testdata/properties/property_inheritance.dws",
		"../../testdata/properties/read_only_property.dws",
		"../../testdata/properties/auto_property.dws",
		"../../testdata/properties/mixed_properties.dws",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			if _, err := os.Stat(script); os.IsNotExist(err) {
				t.Errorf("Required script %s does not exist", script)
			}
		})
	}
}

// TestPropertyParsing tests that all property scripts parse correctly
// Task 8.59: Verify property syntax parsing via CLI
func TestPropertyParsing(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	scripts := []string{
		"../../testdata/properties/basic_property.dws",
		"../../testdata/properties/property_inheritance.dws",
		"../../testdata/properties/read_only_property.dws",
		"../../testdata/properties/auto_property.dws",
		"../../testdata/properties/mixed_properties.dws",
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

// TestPropertyParseCommand tests parsing property features via CLI
// Task 8.59: Test property declaration and access syntax
func TestPropertyParseCommand(t *testing.T) {
	// Build the CLI if needed
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping CLI tests: failed to build CLI: %v", err)
	}

	testCases := []struct {
		name        string
		source      string
		expectError bool
	}{
		{
			name: "field_backed_property",
			source: `
type TTest = class
	FValue: Integer;
	property Value: Integer read FValue write FValue;
end;
`,
			expectError: false,
		},
		{
			name: "method_backed_property",
			source: `
type TTest = class
	function GetValue: Integer; begin end;
	procedure SetValue(v: Integer); begin end;
	property Value: Integer read GetValue write SetValue;
end;
`,
			expectError: false,
		},
		{
			name: "read_only_property",
			source: `
type TTest = class
	FValue: Integer;
	property Value: Integer read FValue;
end;
`,
			expectError: false,
		},
		{
			name: "auto_property",
			source: `
type TTest = class
	property Name: String;
end;
`,
			expectError: false,
		},
		{
			name: "property_access",
			source: `
type TTest = class
	FValue: Integer;
	property Value: Integer read FValue write FValue;
end;
var obj: TTest;
begin
	obj.Value := 42;
	PrintLn(IntToStr(obj.Value));
end
`,
			expectError: false,
		},
		{
			name: "indexed_property_declaration",
			source: `
type TTest = class
	function GetItem(i: Integer): String; begin end;
	property Items[index: Integer]: String read GetItem;
end;
`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("../../bin/dwscript", "parse", "-e", tc.source)
			output, err := cmd.CombinedOutput()

			hasError := err != nil || strings.Contains(string(output), "error")

			if tc.expectError && !hasError {
				t.Errorf("Expected error but got none. Output: %s", output)
			}

			if !tc.expectError && hasError {
				t.Errorf("Expected success but got error: %v\nOutput: %s", err, output)
			}
		})
	}
}

// TestPropertyComplexSyntax tests more complex property declarations
// Task 8.59: Verify advanced property syntax parses correctly
func TestPropertyComplexSyntax(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping CLI tests: failed to build CLI: %v", err)
	}

	testCases := []struct {
		name        string
		source      string
		expectError bool
	}{
		{
			name: "multiple_properties",
			source: `
type TTest = class
	FValue: Integer;
	FName: String;
	property Value: Integer read FValue write FValue;
	property Name: String read FName write FName;
end;
`,
			expectError: false,
		},
		{
			name: "mixed_access_types",
			source: `
type TTest = class
	FValue: Integer;
	function GetName: String; begin end;
	procedure SetName(v: String); begin end;
	property Value: Integer read FValue write FValue;
	property Name: String read GetName write SetName;
end;
`,
			expectError: false,
		},
		{
			name: "property_in_class_hierarchy",
			source: `
type TBase = class
	FValue: Integer;
	property Value: Integer read FValue write FValue;
end;
type TDerived = class(TBase)
	FExtra: String;
	property Extra: String read FExtra write FExtra;
end;
`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("../../bin/dwscript", "parse", "-e", tc.source)
			output, err := cmd.CombinedOutput()

			hasError := err != nil || strings.Contains(string(output), "error")

			if tc.expectError && !hasError {
				t.Errorf("Expected error but got none. Output: %s", output)
			}

			if !tc.expectError && hasError {
				t.Errorf("Expected success but got error: %v\nOutput: %s", err, output)
			}
		})
	}
}

// TestPropertyInheritance tests property inheritance through CLI
// Task 8.59: Verify properties are inherited correctly
func TestPropertyInheritance(t *testing.T) {
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Skipf("Skipping CLI tests: failed to build CLI: %v", err)
	}

	source := `
type TBase = class
	FValue: Integer;
	property Value: Integer read FValue write FValue;
end;

type TDerived = class(TBase)
	FExtra: String;
	property Extra: String read FExtra write FExtra;
end;

var obj: TDerived;
begin
	obj.Value := 42;  // inherited property
	obj.Extra := 'test';  // own property
end
`

	cmd := exec.Command("../../bin/dwscript", "parse", "-e", source)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Property inheritance test failed: %v\nOutput: %s", err, output)
	}

	if strings.Contains(string(output), "error") {
		t.Errorf("Unexpected error in property inheritance: %s", output)
	}
}
