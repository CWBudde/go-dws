package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestTypedDynArrayVarDecl_InitializesArrayValue(t *testing.T) {
	input := `
		var arrEmpty: array of Integer;
		begin
			PrintLn(Length(arrEmpty));
		end.
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		panic("parser errors: " + joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	val := interp.Eval(program)
	if isError(val) {
		t.Fatalf("execution returned error value: %v", val)
	}

	got, ok := interp.Env().Get("arrEmpty")
	if !ok {
		t.Fatalf("arrEmpty not found in environment")
	}

	arr, ok := got.(*ArrayValue)
	if !ok {
		t.Fatalf("arrEmpty expected *ArrayValue, got %T", got)
	}
	if arr.ArrayType == nil {
		t.Fatalf("arrEmpty.ArrayType is nil")
	}
}

func TestEmptyBracketLiteral_ParsesAsArrayLiteralInTypedArrayAssignment(t *testing.T) {
	input := `
		var arrEmpty: array of Integer;
		begin
			arrEmpty := [];
		end.
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		panic("parser errors: " + joinParserErrorsNewline(p.Errors()))
	}

	if len(program.Statements) < 2 {
		t.Fatalf("expected at least 2 top-level statements, got %d", len(program.Statements))
	}
	block, ok := program.Statements[1].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected second statement to be *ast.BlockStatement, got %T", program.Statements[1])
	}
	if len(block.Statements) < 1 {
		t.Fatalf("expected at least 1 statement in block")
	}
	assign, ok := block.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("expected first block statement to be *ast.AssignmentStatement, got %T", block.Statements[0])
	}
	if _, ok := assign.Target.(*ast.Identifier); !ok {
		t.Fatalf("expected assignment target to be *ast.Identifier, got %T", assign.Target)
	}
	if ident, ok := assign.Target.(*ast.Identifier); ok {
		if ident.Value != "arrEmpty" {
			t.Fatalf("expected assignment target identifier to be arrEmpty, got %q", ident.Value)
		}
	}
	if _, ok := assign.Value.(*ast.ArrayLiteralExpression); !ok {
		t.Fatalf("expected [] to parse as *ast.ArrayLiteralExpression in this context, got %T", assign.Value)
	}
}

func TestEmptyArrayLiteralAssignment_InheritsTargetType(t *testing.T) {
	input := `
		var arrEmpty: array of Integer;
		begin
			arrEmpty := [];
			PrintLn(Length(arrEmpty));
		end.
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		panic("parser errors: " + joinParserErrorsNewline(p.Errors()))
	}

	if len(program.Statements) < 2 {
		t.Fatalf("expected at least 2 top-level statements, got %d", len(program.Statements))
	}
	block, ok := program.Statements[1].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected second statement to be *ast.BlockStatement, got %T", program.Statements[1])
	}
	if len(block.Statements) < 2 {
		t.Fatalf("expected at least 2 statements in block, got %d", len(block.Statements))
	}
	assign, ok := block.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("expected first block statement to be *ast.AssignmentStatement, got %T", block.Statements[0])
	}

	var buf bytes.Buffer
	interp := New(&buf)

	// 1) Evaluate only the var declaration first.
	declResult := interp.Eval(program.Statements[0])
	if isError(declResult) {
		t.Fatalf("var decl returned error value: %v", declResult)
	}

	// Sanity: target exists and is typed before assignment.
	existing, ok := interp.Env().Get("arrEmpty")
	if !ok {
		t.Fatalf("arrEmpty not found in environment after decl")
	}
	arr, ok := existing.(*ArrayValue)
	if !ok {
		t.Fatalf("arrEmpty expected *ArrayValue, got %T", existing)
	}
	if arr.ArrayType == nil {
		t.Fatalf("arrEmpty.ArrayType is nil before assignment")
	}

	// 2) Evaluate the assignment and ensure it succeeds.
	assignResult := interp.Eval(assign)
	if isError(assignResult) {
		t.Fatalf("assignment returned error value: %v", assignResult)
	}

	// 3) After assignment, array should still be typed and empty.
	existing2, ok := interp.Env().Get("arrEmpty")
	if !ok {
		t.Fatalf("arrEmpty not found in environment after assignment")
	}
	arr2, ok := existing2.(*ArrayValue)
	if !ok {
		t.Fatalf("arrEmpty expected *ArrayValue, got %T", existing2)
	}
	if arr2.ArrayType == nil {
		t.Fatalf("arrEmpty.ArrayType is nil after assignment")
	}
	if len(arr2.Elements) != 0 {
		t.Fatalf("expected arrEmpty to be empty after assignment, got %d elements", len(arr2.Elements))
	}
}
