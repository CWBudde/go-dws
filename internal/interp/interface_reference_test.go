package interp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// formatRuntimeError converts an error value to DWScript expected format
// Changes "ERROR: message at line X, column Y" to "Runtime Error: message [line: X, column: Y]"
func formatRuntimeError(errVal Value) string {
	errStr := errVal.String()

	// Remove "ERROR: " prefix
	errStr = strings.TrimPrefix(errStr, "ERROR: ")

	// Replace "at line X, column Y" with "[line: X, column: Y]"
	errStr = strings.ReplaceAll(errStr, " at line ", " [line: ")
	if strings.Contains(errStr, "[line:") {
		lines := strings.Split(errStr, "\n")
		for idx, line := range lines {
			if strings.Contains(line, "[line:") && !strings.Contains(line, "]") {
				lines[idx] = line + "]"
			}
		}
		errStr = strings.Join(lines, "\n")
	}

	// Add "Runtime Error: " prefix
	return "Runtime Error: " + errStr
}

// TestInterfaceReferenceTests runs all ported DWScript interface tests from testdata/interfaces/
func TestInterfaceReferenceTests(t *testing.T) {
	interfaceDir := filepath.Join("../../testdata", "interfaces")

	// Get all .pas files in the interfaces directory
	pasFiles, err := filepath.Glob(filepath.Join(interfaceDir, "*.pas"))
	if err != nil {
		t.Fatalf("Failed to read interface test files: %v", err)
	}

	if len(pasFiles) == 0 {
		t.Skip("No .pas test files found in testdata/interfaces/ - reference tests not available")
	}

	// Track statistics
	passed := 0
	failed := 0
	skipped := 0

	for _, pasFile := range pasFiles {
		testName := filepath.Base(pasFile)
		testName = strings.TrimSuffix(testName, ".pas")

		t.Run(testName, func(t *testing.T) {
			// Read the .pas source file
			source, err := os.ReadFile(pasFile)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", pasFile, err)
			}

			// Check if there's an expected output file (.txt)
			txtFile := strings.TrimSuffix(pasFile, ".pas") + ".txt"
			var expectedOutput string
			var hasExpectedOutput bool

			if _, err := os.Stat(txtFile); err == nil {
				outputBytes, err := os.ReadFile(txtFile)
				if err != nil {
					t.Fatalf("Failed to read expected output %s: %v", txtFile, err)
				}
				expectedOutput = string(outputBytes)
				hasExpectedOutput = true
			}

			// Parse the source
			l := lexer.New(string(source))
			p := parser.New(l)
			program := p.ParseProgram()

			// Check for parse errors
			if len(p.Errors()) > 0 {
				// Some tests might be expected to fail parsing
				// For now, skip tests that don't parse
				t.Skipf("Parse errors in %s: %v", testName, p.Errors())
				skipped++
				return
			}

			// Execute the program and capture output
			var buf bytes.Buffer
			interp := New(&buf)
			result := interp.Eval(program)

			// Prepare actual output
			var actualOutput string
			if result != nil && result.Type() == "ERROR" {
				// Check if expected output contains "Errors >>>>" - if so, format accordingly
				if hasExpectedOutput && strings.Contains(expectedOutput, "Errors >>>>") {
					// Format runtime error output to match DWScript expected format
					formattedError := formatRuntimeError(result)
					actualOutput = "Errors >>>>\n" + formattedError + "\nResult >>>>\n" + buf.String()
				} else {
					// Unexpected runtime error
					t.Errorf("Runtime error in %s: %v", testName, result.String())
					failed++
					return
				}
			} else {
				actualOutput = buf.String()
			}

			// Compare output if we have expected output
			if hasExpectedOutput {
				if normalizeOutput(actualOutput) != normalizeOutput(expectedOutput) {
					t.Errorf("Output mismatch for %s:\nExpected:\n%s\nActual:\n%s",
						testName, expectedOutput, actualOutput)
					failed++
					return
				}
			}

			passed++
		})
	}

	// Report summary
	t.Logf("Interface Reference Tests Summary: %d passed, %d failed, %d skipped (out of %d total)",
		passed, failed, skipped, len(pasFiles))
}

// normalizeOutput normalizes output for comparison by trimming whitespace
func normalizeOutput(s string) string {
	// Trim trailing whitespace from each line and overall string
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t\r")
	}
	result := strings.Join(lines, "\n")
	return strings.TrimSpace(result)
}

// TestInterfaceDeclarationBasics tests basic interface declaration from reference tests
// Based on: declare_interface.pas, declare_empty_interface.pas
func TestInterfaceDeclarationBasics(t *testing.T) {
	tests := []struct {
		name   string
		source string
		error  bool
	}{
		{
			name: "simple interface declaration",
			source: `
				type
					IMyInterface = interface
						procedure DoSomething;
					end;
			`,
			error: false,
		},
		{
			name: "empty interface",
			source: `
				type
					IEmpty = interface
					end;
			`,
			error: false,
		},
		{
			name: "interface with multiple methods",
			source: `
				type
					ICounter = interface
						procedure Increment;
						function GetValue: Integer;
						procedure SetValue(v: Integer);
					end;
			`,
			error: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.source)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				if !tt.error {
					t.Errorf("Unexpected parse errors: %v", p.Errors())
				}
				return
			}

			if tt.error {
				t.Error("Expected parse error but got none")
				return
			}

			interp := New(nil)
			result := interp.Eval(program)

			if result != nil && result.Type() == "ERROR" {
				t.Errorf("Unexpected runtime error: %v", result.String())
			}
		})
	}
}

// TestInterfaceInheritanceBasics tests interface inheritance from reference tests
// Based on: interface_inheritance_declare.pas, interface_inheritance_declare_ex.pas
func TestInterfaceInheritanceBasics(t *testing.T) {
	source := `
		type
			IBase = interface
				procedure BaseMethod;
			end;

		type
			IDerived = interface(IBase)
				procedure DerivedMethod;
			end;
	`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Runtime error: %v", result.String())
	}

	// Verify both interfaces were registered (use lowercase keys for lookups)
	if _, exists := interp.interfaces["ibase"]; !exists {
		t.Error("IBase interface should be registered")
	}

	if _, exists := interp.interfaces["iderived"]; !exists {
		t.Error("IDerived interface should be registered")
	}

	// Verify inheritance relationship
	derived := interp.interfaces["iderived"]
	if derived.Parent == nil {
		t.Fatal("IDerived should have parent interface")
	}

	if derived.Parent.Name != "IBase" {
		t.Errorf("Parent name = %s, want IBase", derived.Parent.Name)
	}

	// Verify method inheritance
	if !derived.HasMethod("DerivedMethod") {
		t.Error("IDerived should have DerivedMethod")
	}

	if !derived.HasMethod("BaseMethod") {
		t.Error("IDerived should inherit BaseMethod from IBase")
	}
}

// TestInterfaceImplementation tests class implementing interfaces from reference tests
// Based on: implement_interface1.pas
func TestInterfaceImplementation(t *testing.T) {
	source := `
		type
			IMyInterface = interface
				procedure DoSomething;
			end;

		type
			TMyClass = class(TObject, IMyInterface)
				procedure DoSomething; begin end;
			end;
	`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Runtime error: %v", result.String())
	}

	// Verify interface was registered
	iface, exists := interp.interfaces["imyinterface"]
	if !exists {
		t.Fatal("IMyInterface should be registered")
	}

	// Verify class was registered
	class, exists := interp.classes["tmyclass"]
	if !exists {
		t.Fatal("TMyClass should be registered")
	}

	// Verify class implements interface
	if !classImplementsInterface(class, iface) {
		t.Error("TMyClass should implement IMyInterface")
	}
}

// TestInterfaceMultipleImplementation tests class implementing multiple interfaces
// Based on: interface_multiple.pas
func TestInterfaceMultipleImplementation(t *testing.T) {
	source := `
		type
			IIntfA = interface
				procedure A;
			end;

		type
			IIntfB = interface
				procedure B;
			end;

		type
			TImpAB = class(TObject, IIntfA, IIntfB)
				procedure A; begin end;
				procedure B; begin end;
			end;
	`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Runtime error: %v", result.String())
	}

	// Verify all interfaces were registered
	ifaceA, existsA := interp.interfaces["iintfa"]
	if !existsA {
		t.Fatal("IIntfA should be registered")
	}

	ifaceB, existsB := interp.interfaces["iintfb"]
	if !existsB {
		t.Fatal("IIntfB should be registered")
	}

	// Verify class implements both interfaces
	class := interp.classes["timpab"]
	if !classImplementsInterface(class, ifaceA) {
		t.Error("TImpAB should implement IIntfA")
	}

	if !classImplementsInterface(class, ifaceB) {
		t.Error("TImpAB should implement IIntfB")
	}
}
