package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestAssignmentEndPos verifies that assignment statements have correct EndPos tracking
func TestAssignmentEndPos(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Simple Assignment", "x := 42;"},
		{"Member Assignment", "obj.field := 100;"},
		{"Array Assignment", "arr[0] := 5;"},
		{"Compound Assignment", "x += 10;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) == 0 {
				t.Fatal("No statements parsed")
			}

			stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
			if !ok {
				t.Fatalf("Expected AssignmentStatement, got %T", program.Statements[0])
			}

			// Verify EndPos is set
			if stmt.EndPos.Line == 0 && stmt.EndPos.Column == 0 {
				t.Errorf("EndPos not set for assignment statement")
			}

			// Verify EndPos is after Pos
			if stmt.EndPos.Column <= stmt.Pos().Column {
				t.Errorf("EndPos (%d) should be after Pos (%d)", stmt.EndPos.Column, stmt.Pos().Column)
			}
		})
	}
}

// TestLambdaEndPos verifies that lambda expressions have correct EndPos tracking
func TestLambdaEndPos(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Lambda Shorthand", "lambda(x: Integer) => x * 2"},
		{"Lambda Full", "lambda(x: Integer): Integer begin Result := x * 2; end"},
		{"Lambda No Params", "lambda() => 42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) == 0 {
				t.Fatal("No statements parsed")
			}

			exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
			}

			lambda, ok := exprStmt.Expression.(*ast.LambdaExpression)
			if !ok {
				t.Fatalf("Expected LambdaExpression, got %T", exprStmt.Expression)
			}

			// Verify EndPos is set
			if lambda.EndPos.Line == 0 && lambda.EndPos.Column == 0 {
				t.Errorf("EndPos not set for lambda expression")
			}

			// Verify EndPos is after Pos
			if lambda.EndPos.Column <= lambda.Pos().Column {
				t.Errorf("EndPos (%d) should be after Pos (%d)", lambda.EndPos.Column, lambda.Pos().Column)
			}
		})
	}
}

// TestInheritedEndPos verifies that inherited expressions have correct EndPos tracking
func TestInheritedEndPos(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Inherited Bare", "inherited"},
		{"Inherited Method", "inherited DoSomething"},
		{"Inherited Call", "inherited DoSomething(1, 2)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) == 0 {
				t.Fatal("No statements parsed")
			}

			exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
			}

			inherited, ok := exprStmt.Expression.(*ast.InheritedExpression)
			if !ok {
				t.Fatalf("Expected InheritedExpression, got %T", exprStmt.Expression)
			}

			// Verify EndPos is set
			if inherited.EndPos.Line == 0 && inherited.EndPos.Column == 0 {
				t.Errorf("EndPos not set for inherited expression")
			}

			// Verify EndPos is at or after Pos
			if inherited.EndPos.Column < inherited.Pos().Column {
				t.Errorf("EndPos (%d) should be at or after Pos (%d)", inherited.EndPos.Column, inherited.Pos().Column)
			}
		})
	}
}
