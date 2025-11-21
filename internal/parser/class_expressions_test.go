package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// New Expression Parsing Tests (Object Creation)
// ============================================================================

func TestNewExpression(t *testing.T) {
	input := `TPoint.Create(10, 20)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	newExpr, ok := stmt.Expression.(*ast.NewExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.NewExpression. got=%T",
			stmt.Expression)
	}

	if newExpr.ClassName.Value != "TPoint" {
		t.Errorf("newExpr.ClassName.Value not 'TPoint'. got=%s", newExpr.ClassName.Value)
	}

	if len(newExpr.Arguments) != 2 {
		t.Fatalf("newExpr.Arguments should contain 2 arguments. got=%d", len(newExpr.Arguments))
	}
}

func TestNewExpressionNoArguments(t *testing.T) {
	input := `TObject.Create()`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	newExpr, ok := stmt.Expression.(*ast.NewExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.NewExpression. got=%T",
			stmt.Expression)
	}

	if newExpr.ClassName.Value != "TObject" {
		t.Errorf("newExpr.ClassName.Value not 'TObject'. got=%s", newExpr.ClassName.Value)
	}

	if len(newExpr.Arguments) != 0 {
		t.Fatalf("newExpr.Arguments should be empty. got=%d", len(newExpr.Arguments))
	}
}

func TestNewExpressionOptionalParentheses(t *testing.T) {
	// Test that parentheses are optional for parameterless constructors
	input := `new TTest`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	newExpr, ok := stmt.Expression.(*ast.NewExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.NewExpression. got=%T",
			stmt.Expression)
	}

	if newExpr.ClassName.Value != "TTest" {
		t.Errorf("newExpr.ClassName.Value not 'TTest'. got=%s", newExpr.ClassName.Value)
	}

	if len(newExpr.Arguments) != 0 {
		t.Fatalf("newExpr.Arguments should be empty. got=%d", len(newExpr.Arguments))
	}
}

// ============================================================================
// Member Access Parsing Tests
// ============================================================================

func TestMemberAccess(t *testing.T) {
	input := `point.X`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	memberExpr, ok := stmt.Expression.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.MemberAccessExpression. got=%T",
			stmt.Expression)
	}

	// Check object is an identifier
	ident, ok := memberExpr.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("memberExpr.Object is not *ast.Identifier. got=%T",
			memberExpr.Object)
	}

	if ident.Value != "point" {
		t.Errorf("ident.Value not 'point'. got=%s", ident.Value)
	}

	if memberExpr.Member.Value != "X" {
		t.Errorf("memberExpr.Member.Value not 'X'. got=%s", memberExpr.Member.Value)
	}
}

func TestChainedMemberAccess(t *testing.T) {
	input := `obj.field1.field2`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	// The outer member access (field2)
	outerMember, ok := stmt.Expression.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.MemberAccessExpression. got=%T",
			stmt.Expression)
	}

	if outerMember.Member.Value != "field2" {
		t.Errorf("outerMember.Member.Value not 'field2'. got=%s", outerMember.Member.Value)
	}

	// The inner member access (obj.field1)
	innerMember, ok := outerMember.Object.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("outerMember.Object is not *ast.MemberAccessExpression. got=%T",
			outerMember.Object)
	}

	if innerMember.Member.Value != "field1" {
		t.Errorf("innerMember.Member.Value not 'field1'. got=%s", innerMember.Member.Value)
	}

	// The base object
	ident, ok := innerMember.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("innerMember.Object is not *ast.Identifier. got=%T",
			innerMember.Object)
	}

	if ident.Value != "obj" {
		t.Errorf("ident.Value not 'obj'. got=%s", ident.Value)
	}
}

// ============================================================================
// Method Call Parsing Tests
// ============================================================================

func TestMethodCall(t *testing.T) {
	input := `obj.DoSomething(42, "hello")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	methodCall, ok := stmt.Expression.(*ast.MethodCallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.MethodCallExpression. got=%T",
			stmt.Expression)
	}

	// Check object is an identifier
	ident, ok := methodCall.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("methodCall.Object is not *ast.Identifier. got=%T",
			methodCall.Object)
	}

	if ident.Value != "obj" {
		t.Errorf("ident.Value not 'obj'. got=%s", ident.Value)
	}

	if methodCall.Method.Value != "DoSomething" {
		t.Errorf("methodCall.Method.Value not 'DoSomething'. got=%s", methodCall.Method.Value)
	}

	if len(methodCall.Arguments) != 2 {
		t.Fatalf("methodCall.Arguments should contain 2 arguments. got=%d", len(methodCall.Arguments))
	}
}
