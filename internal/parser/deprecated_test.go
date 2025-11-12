package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestConstDeprecated tests parsing of deprecated constants
func TestConstDeprecated(t *testing.T) {
	input := `const c1 = 'world' deprecated;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	constDecl, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("Expected ConstDecl, got %T", program.Statements[0])
	}

	if constDecl.Name.Value != "c1" {
		t.Errorf("Expected const name 'c1', got '%s'", constDecl.Name.Value)
	}

	if !constDecl.IsDeprecated {
		t.Error("Expected IsDeprecated to be true")
	}

	if constDecl.DeprecatedMessage != "" {
		t.Errorf("Expected empty deprecation message, got '%s'", constDecl.DeprecatedMessage)
	}
}

// TestConstDeprecatedWithMessage tests parsing of deprecated constants with message
func TestConstDeprecatedWithMessage(t *testing.T) {
	input := `const c2 = 'hello' deprecated 'use bye';`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	constDecl, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("Expected ConstDecl, got %T", program.Statements[0])
	}

	if constDecl.Name.Value != "c2" {
		t.Errorf("Expected const name 'c2', got '%s'", constDecl.Name.Value)
	}

	if !constDecl.IsDeprecated {
		t.Error("Expected IsDeprecated to be true")
	}

	if constDecl.DeprecatedMessage != "use bye" {
		t.Errorf("Expected deprecation message 'use bye', got '%s'", constDecl.DeprecatedMessage)
	}
}

// TestFunctionDeprecated tests parsing of deprecated functions
func TestFunctionDeprecated(t *testing.T) {
	input := `procedure TestProc; deprecated;
begin
end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Statements[0])
	}

	if funcDecl.Name.Value != "TestProc" {
		t.Errorf("Expected function name 'TestProc', got '%s'", funcDecl.Name.Value)
	}

	if !funcDecl.IsDeprecated {
		t.Error("Expected IsDeprecated to be true")
	}

	if funcDecl.DeprecatedMessage != "" {
		t.Errorf("Expected empty deprecation message, got '%s'", funcDecl.DeprecatedMessage)
	}
}

// TestFunctionDeprecatedWithMessage tests parsing of deprecated functions with message
func TestFunctionDeprecatedWithMessage(t *testing.T) {
	input := `function TestFunc: Integer; deprecated 'returns 1';
begin
   Result := 1;
end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Statements[0])
	}

	if funcDecl.Name.Value != "TestFunc" {
		t.Errorf("Expected function name 'TestFunc', got '%s'", funcDecl.Name.Value)
	}

	if !funcDecl.IsDeprecated {
		t.Error("Expected IsDeprecated to be true")
	}

	if funcDecl.DeprecatedMessage != "returns 1" {
		t.Errorf("Expected deprecation message 'returns 1', got '%s'", funcDecl.DeprecatedMessage)
	}
}

// TestMethodDeprecated tests parsing of deprecated class methods
func TestMethodDeprecated(t *testing.T) {
	input := `type
   TTest = class
      method Meth; deprecated 'Meth is gone';
   end;

method TTest.Meth;
begin
end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("Expected ClassDecl, got %T", program.Statements[0])
	}

	if len(classDecl.Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(classDecl.Methods))
	}

	method := classDecl.Methods[0]
	if method.Name.Value != "Meth" {
		t.Errorf("Expected method name 'Meth', got '%s'", method.Name.Value)
	}

	if !method.IsDeprecated {
		t.Error("Expected IsDeprecated to be true")
	}

	if method.DeprecatedMessage != "Meth is gone" {
		t.Errorf("Expected deprecation message 'Meth is gone', got '%s'", method.DeprecatedMessage)
	}
}

// TestEnumElementDeprecated tests parsing of deprecated enum elements
func TestEnumElementDeprecated(t *testing.T) {
	input := `type TEnum = (
	zzero deprecated,
	zero = 0,
	One = 1,
	deux deprecated 'use two' = 2,
	three = 3,
	two = 2
	);`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	enumDecl, ok := program.Statements[0].(*ast.EnumDecl)
	if !ok {
		t.Fatalf("Expected EnumDecl, got %T", program.Statements[0])
	}

	if len(enumDecl.Values) != 6 {
		t.Fatalf("Expected 6 enum values, got %d", len(enumDecl.Values))
	}

	// Check first element (zzero deprecated)
	if enumDecl.Values[0].Name != "zzero" {
		t.Errorf("Expected first enum value 'zzero', got '%s'", enumDecl.Values[0].Name)
	}
	if !enumDecl.Values[0].IsDeprecated {
		t.Error("Expected first enum value to be deprecated")
	}
	if enumDecl.Values[0].DeprecatedMessage != "" {
		t.Errorf("Expected empty deprecation message for first value, got '%s'", enumDecl.Values[0].DeprecatedMessage)
	}

	// Check fourth element (deux deprecated 'use two' = 2)
	if enumDecl.Values[3].Name != "deux" {
		t.Errorf("Expected fourth enum value 'deux', got '%s'", enumDecl.Values[3].Name)
	}
	if !enumDecl.Values[3].IsDeprecated {
		t.Error("Expected fourth enum value to be deprecated")
	}
	if enumDecl.Values[3].DeprecatedMessage != "use two" {
		t.Errorf("Expected deprecation message 'use two' for fourth value, got '%s'", enumDecl.Values[3].DeprecatedMessage)
	}
	if enumDecl.Values[3].Value == nil || *enumDecl.Values[3].Value != 2 {
		t.Error("Expected fourth enum value to have explicit value 2")
	}

	// Check that zero (second element) is NOT deprecated
	if enumDecl.Values[1].IsDeprecated {
		t.Error("Expected second enum value 'zero' to NOT be deprecated")
	}
}

// TestMultipleConstDeprecated tests parsing multiple deprecated constants
func TestMultipleConstDeprecated(t *testing.T) {
	input := `const c1 = 'world' deprecated;
const c2 = 'hello' deprecated 'use bye';
const c3 = 42;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("Expected 3 statements, got %d", len(program.Statements))
	}

	// Check first const (deprecated without message)
	constDecl1, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("Expected first statement to be ConstDecl, got %T", program.Statements[0])
	}
	if constDecl1.Name.Value != "c1" {
		t.Errorf("Expected first const name 'c1', got '%s'", constDecl1.Name.Value)
	}
	if !constDecl1.IsDeprecated {
		t.Error("First const should be deprecated")
	}

	// Check second const (deprecated with message)
	constDecl2, ok := program.Statements[1].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("Expected second statement to be ConstDecl, got %T", program.Statements[1])
	}
	if constDecl2.Name.Value != "c2" {
		t.Errorf("Expected second const name 'c2', got '%s'", constDecl2.Name.Value)
	}
	if !constDecl2.IsDeprecated {
		t.Error("Second const should be deprecated")
	}
	if constDecl2.DeprecatedMessage != "use bye" {
		t.Errorf("Expected deprecation message 'use bye', got '%s'", constDecl2.DeprecatedMessage)
	}

	// Check third const (NOT deprecated)
	constDecl3, ok := program.Statements[2].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("Expected third statement to be ConstDecl, got %T", program.Statements[2])
	}
	if constDecl3.Name.Value != "c3" {
		t.Errorf("Expected third const name 'c3', got '%s'", constDecl3.Name.Value)
	}
	if constDecl3.IsDeprecated {
		t.Error("Third const should NOT be deprecated")
	}
}

// TestForwardDeprecatedFunction tests deprecated directive on forward declarations
func TestForwardDeprecatedFunction(t *testing.T) {
	input := `procedure TestProc; forward; deprecated;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Statements[0])
	}

	if funcDecl.Name.Value != "TestProc" {
		t.Errorf("Expected function name 'TestProc', got '%s'", funcDecl.Name.Value)
	}

	if !funcDecl.IsForward {
		t.Error("Expected IsForward to be true")
	}

	if !funcDecl.IsDeprecated {
		t.Error("Expected IsDeprecated to be true")
	}
}
