package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

func TestParseOperatorDeclaration_Binary(t *testing.T) {
	input := `
operator + (String, Integer) : String uses StrPlusInt;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.OperatorDecl)
	if !ok {
		t.Fatalf("statement is not *ast.OperatorDecl, got %T", program.Statements[0])
	}

	if stmt.Kind != ast.OperatorKindGlobal {
		t.Fatalf("expected OperatorKindGlobal, got %s", stmt.Kind)
	}

	if stmt.OperatorSymbol != "+" {
		t.Fatalf("expected operator '+', got %q", stmt.OperatorSymbol)
	}

	if stmt.Arity != 2 {
		t.Fatalf("expected arity 2, got %d", stmt.Arity)
	}

	if len(stmt.OperandTypes) != 2 {
		t.Fatalf("expected 2 operand types, got %d", len(stmt.OperandTypes))
	}

	if stmt.OperandTypes[0].String() != "String" || stmt.OperandTypes[1].String() != "Integer" {
		t.Fatalf("unexpected operand types: %q, %q", stmt.OperandTypes[0].String(), stmt.OperandTypes[1].String())
	}

	if stmt.ReturnType == nil || stmt.ReturnType.String() != "String" {
		t.Fatalf("expected return type String, got %v", stmt.ReturnType)
	}

	if stmt.Binding == nil || stmt.Binding.Value != "StrPlusInt" {
		t.Fatalf("expected binding StrPlusInt, got %v", stmt.Binding)
	}
}

func TestParseOperatorDeclaration_ImplicitConversion(t *testing.T) {
	input := `
operator implicit (Integer) : String uses IntToStr;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.OperatorDecl)
	if !ok {
		t.Fatalf("statement is not *ast.OperatorDecl, got %T", program.Statements[0])
	}

	if stmt.Kind != ast.OperatorKindConversion {
		t.Fatalf("expected OperatorKindConversion, got %s", stmt.Kind)
	}

	if stmt.OperatorSymbol != "implicit" {
		t.Fatalf("expected operator 'implicit', got %q", stmt.OperatorSymbol)
	}

	if stmt.Arity != 1 {
		t.Fatalf("expected arity 1, got %d", stmt.Arity)
	}

	if len(stmt.OperandTypes) != 1 || stmt.OperandTypes[0].String() != "Integer" {
		t.Fatalf("expected operand type Integer, got %v", stmt.OperandTypes)
	}

	if stmt.ReturnType == nil || stmt.ReturnType.String() != "String" {
		t.Fatalf("expected return type String, got %v", stmt.ReturnType)
	}

	if stmt.Binding == nil || stmt.Binding.Value != "IntToStr" {
		t.Fatalf("expected binding IntToStr, got %v", stmt.Binding)
	}
}

func TestParseOperatorDeclaration_InOperator(t *testing.T) {
	input := `
operator in (Integer, Float) : Boolean uses DigitInFloat;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.OperatorDecl)
	if !ok {
		t.Fatalf("statement is not *ast.OperatorDecl, got %T", program.Statements[0])
	}

	if stmt.OperatorSymbol != "in" {
		t.Fatalf("expected operator 'in', got %q", stmt.OperatorSymbol)
	}

	if stmt.Arity != 2 {
		t.Fatalf("expected arity 2, got %d", stmt.Arity)
	}

	if stmt.ReturnType == nil || stmt.ReturnType.String() != "Boolean" {
		t.Fatalf("expected return type Boolean, got %v", stmt.ReturnType)
	}

	if stmt.Binding == nil || stmt.Binding.Value != "DigitInFloat" {
		t.Fatalf("expected binding DigitInFloat, got %v", stmt.Binding)
	}
}

// Task 8.21b: Test parsing class operator
func TestParseOperatorDeclaration_ClassOperator(t *testing.T) {
	input := `
type TTest = class
  Field: String;
  constructor Create;
  function AddString(str: String): TTest;
  class operator + (TTest, String) : TTest uses AddString;
end;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("statement is not *ast.ClassDecl, got %T", program.Statements[0])
	}

	// Check that we have an operator declaration
	if len(classDecl.Operators) != 1 {
		t.Fatalf("expected 1 operator declaration, got %d", len(classDecl.Operators))
	}

	operatorDecl := classDecl.Operators[0]

	if operatorDecl.Kind != ast.OperatorKindClass {
		t.Fatalf("expected OperatorKindClass, got %s", operatorDecl.Kind)
	}

	if operatorDecl.OperatorSymbol != "+" {
		t.Fatalf("expected operator '+', got %q", operatorDecl.OperatorSymbol)
	}

	if operatorDecl.Arity != 2 {
		t.Fatalf("expected arity 2, got %d", operatorDecl.Arity)
	}

	if len(operatorDecl.OperandTypes) != 2 {
		t.Fatalf("expected 2 operand types, got %d", len(operatorDecl.OperandTypes))
	}

	if operatorDecl.OperandTypes[0].String() != "TTest" || operatorDecl.OperandTypes[1].String() != "String" {
		t.Fatalf("unexpected operand types: %q, %q", operatorDecl.OperandTypes[0].String(), operatorDecl.OperandTypes[1].String())
	}

	if operatorDecl.ReturnType == nil || operatorDecl.ReturnType.String() != "TTest" {
		t.Fatalf("expected return type TTest, got %v", operatorDecl.ReturnType)
	}

	if operatorDecl.Binding == nil || operatorDecl.Binding.Value != "AddString" {
		t.Fatalf("expected binding AddString, got %v", operatorDecl.Binding)
	}
}

// Task 8.21d: Test parsing explicit conversion operator
func TestParseOperatorDeclaration_ExplicitConversion(t *testing.T) {
	input := `
operator explicit (TFoo) : Integer uses FooToInt;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.OperatorDecl)
	if !ok {
		t.Fatalf("statement is not *ast.OperatorDecl, got %T", program.Statements[0])
	}

	if stmt.Kind != ast.OperatorKindConversion {
		t.Fatalf("expected OperatorKindConversion, got %s", stmt.Kind)
	}

	if stmt.OperatorSymbol != "explicit" {
		t.Fatalf("expected operator 'explicit', got %q", stmt.OperatorSymbol)
	}

	if stmt.Arity != 1 {
		t.Fatalf("expected arity 1, got %d", stmt.Arity)
	}

	if len(stmt.OperandTypes) != 1 || stmt.OperandTypes[0].String() != "TFoo" {
		t.Fatalf("expected operand type TFoo, got %v", stmt.OperandTypes)
	}

	if stmt.ReturnType == nil || stmt.ReturnType.String() != "Integer" {
		t.Fatalf("expected return type Integer, got %v", stmt.ReturnType)
	}

	if stmt.Binding == nil || stmt.Binding.Value != "FooToInt" {
		t.Fatalf("expected binding FooToInt, got %v", stmt.Binding)
	}
}

// Task 8.21e: Test parsing symbolic operators (==, !=, <<, >>, IN)
func TestParseOperatorDeclaration_SymbolicOperators(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedSymbol  string
		expectedBinding string
	}{
		{
			name:            "double equals",
			input:           `operator == (Integer, Integer) : Boolean uses EqualWithPrint;`,
			expectedSymbol:  "==",
			expectedBinding: "EqualWithPrint",
		},
		{
			name:            "not equals",
			input:           `operator != (Integer, Integer) : Boolean uses NotEqualWithPrint;`,
			expectedSymbol:  "!=",
			expectedBinding: "NotEqualWithPrint",
		},
		{
			name:            "left shift",
			input:           `operator << (TMyStream, String) : TMyStream uses StreamString;`,
			expectedSymbol:  "<<",
			expectedBinding: "StreamString",
		},
		{
			name:            "right shift",
			input:           `operator >> (TMyStream, String) : TMyStream uses StreamDummy;`,
			expectedSymbol:  ">>",
			expectedBinding: "StreamDummy",
		},
		{
			name:            "IN operator",
			input:           `operator in (Integer, Float) : Boolean uses DigitInFloat;`,
			expectedSymbol:  "in",
			expectedBinding: "DigitInFloat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.OperatorDecl)
			if !ok {
				t.Fatalf("statement is not *ast.OperatorDecl, got %T", program.Statements[0])
			}

			if stmt.Kind != ast.OperatorKindGlobal {
				t.Fatalf("expected OperatorKindGlobal, got %s", stmt.Kind)
			}

			if stmt.OperatorSymbol != tt.expectedSymbol {
				t.Fatalf("expected operator %q, got %q", tt.expectedSymbol, stmt.OperatorSymbol)
			}

			if stmt.Arity != 2 {
				t.Fatalf("expected arity 2, got %d", stmt.Arity)
			}

			if stmt.Binding == nil || stmt.Binding.Value != tt.expectedBinding {
				t.Fatalf("expected binding %s, got %v", tt.expectedBinding, stmt.Binding)
			}
		})
	}
}

// Task 8.21f: Test parsing unary operator
func TestParseOperatorDeclaration_Unary(t *testing.T) {
	input := `
operator - (TCustom) : TCustom uses Negate;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.OperatorDecl)
	if !ok {
		t.Fatalf("statement is not *ast.OperatorDecl, got %T", program.Statements[0])
	}

	if stmt.Kind != ast.OperatorKindGlobal {
		t.Fatalf("expected OperatorKindGlobal, got %s", stmt.Kind)
	}

	if stmt.OperatorSymbol != "-" {
		t.Fatalf("expected operator '-', got %q", stmt.OperatorSymbol)
	}

	if stmt.Arity != 1 {
		t.Fatalf("expected arity 1, got %d", stmt.Arity)
	}

	if len(stmt.OperandTypes) != 1 || stmt.OperandTypes[0].String() != "TCustom" {
		t.Fatalf("expected operand type TCustom, got %v", stmt.OperandTypes)
	}

	if stmt.ReturnType == nil || stmt.ReturnType.String() != "TCustom" {
		t.Fatalf("expected return type TCustom, got %v", stmt.ReturnType)
	}

	if stmt.Binding == nil || stmt.Binding.Value != "Negate" {
		t.Fatalf("expected binding Negate, got %v", stmt.Binding)
	}
}

// Task 8.21g: Test parsing operator declarations with errors
func TestParseOperatorDeclaration_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "missing uses clause",
			input:         `operator + (String, Integer) : String;`,
			expectedError: "expected 'uses' in operator declaration",
		},
		{
			name:          "invalid operator token",
			input:         `operator xyz (String) : String uses Foo;`,
			expectedError: "expected operator symbol after 'operator'",
		},
		{
			name:          "missing operand types",
			input:         `operator + : String uses Foo;`,
			expectedError: "expected next token to be LPAREN",
		},
		{
			name:          "empty operand list",
			input:         `operator + () : String uses Foo;`,
			expectedError: "operator declaration requires at least one operand type",
		},
		{
			name:          "missing binding identifier",
			input:         `operator + (String, Integer) : String uses;`,
			expectedError: "expected identifier after 'uses' in operator declaration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			_ = p.ParseProgram()

			if len(p.errors) == 0 {
				t.Fatalf("expected parser error, got none")
			}

			// Check that at least one error contains the expected message (substring match)
			foundExpected := false
			for _, err := range p.errors {
				if len(err) >= len(tt.expectedError) && err[:len(tt.expectedError)] == tt.expectedError {
					foundExpected = true
					break
				}
			}

			if !foundExpected {
				t.Fatalf("expected error starting with %q, got errors: %v", tt.expectedError, p.errors)
			}
		})
	}
}
