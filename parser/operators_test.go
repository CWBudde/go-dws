package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/ast"
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
