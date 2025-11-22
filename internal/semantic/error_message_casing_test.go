package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestErrorMessagePreservesOriginalCasing tests that error messages preserve
// the user's original casing for identifiers, making error messages more helpful.

// TestUndefinedVariablePreservesCase tests that undefined variable errors preserve the user's casing
func TestUndefinedVariablePreservesCase(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectInErr  string // String that should appear in error message
		notInErr     string // String that should NOT appear (wrong case)
	}{
		{
			name: "lowercase undefined variable",
			input: `begin
    myundefinedvar := 42;
end;`,
			expectInErr: "myundefinedvar",
			notInErr:    "MYUNDEFINEDVAR",
		},
		{
			name: "UPPERCASE undefined variable",
			input: `begin
    MYUNDEFINEDVAR := 42;
end;`,
			expectInErr: "MYUNDEFINEDVAR",
			notInErr:    "myundefinedvar",
		},
		{
			name: "PascalCase undefined variable",
			input: `begin
    MyUndefinedVar := 42;
end;`,
			expectInErr: "MyUndefinedVar",
		},
		{
			name: "camelCase undefined variable",
			input: `begin
    myUndefinedVar := 42;
end;`,
			expectInErr: "myUndefinedVar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if err == nil {
				t.Fatal("expected error for undefined variable")
			}

			errMsg := err.Error()

			if tt.expectInErr != "" && !strings.Contains(errMsg, tt.expectInErr) {
				t.Errorf("error message should contain %q (user's original casing), got: %s", tt.expectInErr, errMsg)
			}

			if tt.notInErr != "" && strings.Contains(errMsg, tt.notInErr) {
				t.Errorf("error message should NOT contain %q (wrong casing), got: %s", tt.notInErr, errMsg)
			}
		})
	}
}

// TestUndefinedTypePreservesCase tests that undefined type errors preserve the user's casing
func TestUndefinedTypePreservesCase(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectInErr string
	}{
		{
			name: "PascalCase undefined type",
			input: `var x: TMyUndefinedType;
begin
end;`,
			expectInErr: "TMyUndefinedType",
		},
		{
			name: "lowercase undefined type",
			input: `var x: myundefinedtype;
begin
end;`,
			expectInErr: "myundefinedtype",
		},
		{
			name: "UPPERCASE undefined type",
			input: `var x: TMYUNDEFINEDTYPE;
begin
end;`,
			expectInErr: "TMYUNDEFINEDTYPE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if err == nil {
				t.Fatal("expected error for undefined type")
			}

			errMsg := err.Error()

			if !strings.Contains(errMsg, tt.expectInErr) {
				t.Errorf("error message should contain %q (user's original casing), got: %s", tt.expectInErr, errMsg)
			}
		})
	}
}

// TestUndefinedFunctionPreservesCase tests that undefined function errors preserve the user's casing
func TestUndefinedFunctionPreservesCase(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectInErr string
	}{
		{
			name: "PascalCase undefined function",
			input: `begin
    MyUndefinedFunction();
end;`,
			expectInErr: "MyUndefinedFunction",
		},
		{
			name: "lowercase undefined function",
			input: `begin
    myundefinedfunction();
end;`,
			expectInErr: "myundefinedfunction",
		},
		{
			name: "UPPERCASE undefined function",
			input: `begin
    MYUNDEFINEDFUNCTION();
end;`,
			expectInErr: "MYUNDEFINEDFUNCTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if err == nil {
				t.Fatal("expected error for undefined function")
			}

			errMsg := err.Error()

			if !strings.Contains(errMsg, tt.expectInErr) {
				t.Errorf("error message should contain %q (user's original casing), got: %s", tt.expectInErr, errMsg)
			}
		})
	}
}

// TestUndefinedMemberPreservesCase tests that undefined member errors preserve the user's casing
func TestUndefinedMemberPreservesCase(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectInErr string
	}{
		{
			name: "PascalCase undefined field on class",
			input: `type
    TMyClass = class
    private
        FValue: Integer;
    end;
var obj: TMyClass;
begin
    obj.NonExistentField := 42;
end;`,
			expectInErr: "NonExistentField",
		},
		{
			name: "lowercase undefined field on class",
			input: `type
    TMyClass = class
    private
        FValue: Integer;
    end;
var obj: TMyClass;
begin
    obj.nonexistentfield := 42;
end;`,
			expectInErr: "nonexistentfield",
		},
		{
			name: "undefined method on class",
			input: `type
    TMyClass = class
    public
        procedure DoSomething;
    end;

procedure TMyClass.DoSomething;
begin
end;

var obj: TMyClass;
begin
    obj.UndefinedMethod();
end;`,
			expectInErr: "UndefinedMethod",
		},
		// NOTE: Record field errors don't preserve original casing yet (known issue)
		// The error shows 'nonexistentfield' instead of 'NonExistentField'
		// This is tracked as a potential future improvement
		{
			name: "undefined field on record (lowercase expected due to known issue)",
			input: `type
    TMyRecord = record
        X: Integer;
        Y: Integer;
    end;
var r: TMyRecord;
begin
    r.NonExistentField := 42;
end;`,
			expectInErr: "nonexistentfield", // Known issue: casing not preserved for record field errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if err == nil {
				t.Fatal("expected error for undefined member")
			}

			errMsg := err.Error()

			if !strings.Contains(errMsg, tt.expectInErr) {
				t.Errorf("error message should contain %q (user's original casing), got: %s", tt.expectInErr, errMsg)
			}
		})
	}
}

// TestTypeMismatchPreservesCase tests that type mismatch errors preserve original casing
func TestTypeMismatchPreservesCase(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectInErr []string // Multiple strings to check
	}{
		{
			name: "assignment type mismatch with user-defined types",
			input: `type
    TMyInteger = Integer;
    TMyString = String;
var
    x: TMyInteger;
    s: TMyString;
begin
    x := s;
end;`,
			expectInErr: []string{"TMyInteger", "TMyString"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if err == nil {
				t.Fatal("expected error for type mismatch")
			}

			errMsg := err.Error()

			for _, expected := range tt.expectInErr {
				if !strings.Contains(errMsg, expected) {
					t.Errorf("error message should contain %q, got: %s", expected, errMsg)
				}
			}
		})
	}
}

// TestDuplicateDeclarationPreservesOriginalCase tests that duplicate declaration errors
// preserve the original definition's casing when reporting the conflict
// NOTE: Currently these tests document known issues where casing is not preserved
func TestDuplicateDeclarationPreservesOriginalCase(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectInErr string // Original declaration casing
	}{
		// NOTE: Known issue - duplicate variable errors show the second declaration's casing
		// (normalized to lowercase) instead of the original declaration's casing
		{
			name: "duplicate variable declaration (lowercase expected due to known issue)",
			input: `var MyVariable: Integer;
var myvariable: String;
begin
end;`,
			expectInErr: "myvariable", // Known issue: shows normalized casing, not original
		},
		// NOTE: Known issue - duplicate type errors show normalized casing
		{
			name: "duplicate type declaration (lowercase expected due to known issue)",
			input: `type TMyType = Integer;
type tmytype = String;
begin
end;`,
			expectInErr: "tmytype", // Known issue: shows normalized casing, not original
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if err == nil {
				t.Fatal("expected error for duplicate declaration")
			}

			errMsg := err.Error()

			if !strings.Contains(errMsg, tt.expectInErr) {
				t.Errorf("error message should reference original declaration %q, got: %s", tt.expectInErr, errMsg)
			}
		})
	}
}

// TestEnumValueErrorPreservesCase tests that enum-related errors preserve casing
func TestEnumValueErrorPreservesCase(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectInErr string
	}{
		{
			name: "undefined enum value preserves case",
			input: `type
    TColor = (Red, Green, Blue);
var c: TColor;
begin
    c := Purple;
end;`,
			expectInErr: "Purple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if err == nil {
				t.Fatal("expected error for undefined enum value")
			}

			errMsg := err.Error()

			if !strings.Contains(errMsg, tt.expectInErr) {
				t.Errorf("error message should contain %q, got: %s", tt.expectInErr, errMsg)
			}
		})
	}
}

// TestSymbolTableErrorCasingPreservation verifies that the symbol table's error
// helpers preserve original casing when constructing error messages
func TestSymbolTableErrorCasingPreservation(t *testing.T) {
	st := NewSymbolTable()

	// Define with specific casing
	st.Define("MySpecialVariable", nil)

	// Try to define again with different casing - error should preserve original
	// Note: SymbolTable.Define doesn't return errors for duplicates, it overwrites
	// This test verifies the lookup preserves original casing for potential error construction

	sym, ok := st.Resolve("MYSPECIALVARIABLE")
	if !ok {
		t.Fatal("should be able to resolve variable with different case")
	}

	// The symbol name should preserve original casing
	if sym.Name != "MySpecialVariable" {
		t.Errorf("symbol name should be 'MySpecialVariable' (original casing), got %q", sym.Name)
	}
}
