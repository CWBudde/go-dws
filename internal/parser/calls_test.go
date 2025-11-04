package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

func TestCallExpressionParsing(t *testing.T) {
	input := "Add(1, 2 * 3, foo);"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is not ast.CallExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, call.Function, "Add") {
		return
	}

	if len(call.Arguments) != 3 {
		t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
	}

	if !testLiteralExpression(t, call.Arguments[0], 1) {
		return
	}

	if !testInfixExpression(t, call.Arguments[1], 2, "*", 3) {
		return
	}

	if !testIdentifier(t, call.Arguments[2], "foo") {
		return
	}
}

// TestCallExpressionWithStringArgument ensures string literals work as call arguments.
func TestCallExpressionWithStringArgument(t *testing.T) {
	input := "PrintLn('hello', name);"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is not ast.CallExpression. got=%T", stmt.Expression)
	}

	if len(call.Arguments) != 2 {
		t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
	}

	if !testStringLiteralExpression(t, call.Arguments[0], "hello") {
		return
	}

	if !testIdentifier(t, call.Arguments[1], "name") {
		return
	}
}

// TestCallExpressionTrailingComma validates parsing of trailing commas.
func TestCallExpressionTrailingComma(t *testing.T) {
	input := "Foo(1, 2,);"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is not ast.CallExpression. got=%T", stmt.Expression)
	}

	if len(call.Arguments) != 2 {
		t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
	}

	if !testLiteralExpression(t, call.Arguments[0], 1) {
		return
	}

	if !testLiteralExpression(t, call.Arguments[1], 2) {
		return
	}
}

// TestCallExpressionPrecedence checks call precedence relative to infix operators.
func TestCallExpressionPrecedence(t *testing.T) {
	input := "foo(1 + 2) * 3;"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	// The call expression wraps the argument, then multiplication happens
	expected := "(foo((1 + 2)) * 3)"
	if stmt.Expression.String() != expected {
		t.Errorf("expression String() = %q, want %q", stmt.Expression.String(), expected)
	}
}

// TestStatementDispatch ensures identifiers followed by parentheses parse as calls, not assignments.
