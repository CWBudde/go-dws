package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestExitStatementVariants(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectValue bool
		assert      func(t *testing.T, expr ast.Expression)
	}{
		{
			name:        "ExitWithoutValue",
			input:       "Exit;",
			expectValue: false,
		},
		{
			name:  "ExitFalse",
			input: "Exit False;",
			assert: func(t *testing.T, expr ast.Expression) {
				boolLit, ok := expr.(*ast.BooleanLiteral)
				if !ok {
					t.Fatalf("expected *ast.BooleanLiteral, got %T", expr)
				}
				if boolLit.Value {
					t.Fatalf("expected false literal, got true")
				}
			},
			expectValue: true,
		},
		{
			name:  "ExitInteger",
			input: "Exit 42;",
			assert: func(t *testing.T, expr ast.Expression) {
				intLit, ok := expr.(*ast.IntegerLiteral)
				if !ok {
					t.Fatalf("expected *ast.IntegerLiteral, got %T", expr)
				}
				if intLit.Value != 42 {
					t.Fatalf("expected integer literal 42, got %d", intLit.Value)
				}
			},
			expectValue: true,
		},
		{
			name:  "ExitBinaryExpression",
			input: "Exit x + y;",
			assert: func(t *testing.T, expr ast.Expression) {
				binExpr, ok := expr.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("expected *ast.BinaryExpression, got %T", expr)
				}
				if binExpr.Operator != "+" {
					t.Fatalf("expected '+', got %q", binExpr.Operator)
				}

				leftIdent, ok := binExpr.Left.(*ast.Identifier)
				if !ok || leftIdent.Value != "x" {
					t.Fatalf("expected left identifier 'x', got %T (%v)", binExpr.Left, binExpr.Left)
				}

				rightIdent, ok := binExpr.Right.(*ast.Identifier)
				if !ok || rightIdent.Value != "y" {
					t.Fatalf("expected right identifier 'y', got %T (%v)", binExpr.Right, binExpr.Right)
				}
			},
			expectValue: true,
		},
		{
			name:  "ExitParenthesizedIdentifier",
			input: "Exit(value);",
			assert: func(t *testing.T, expr ast.Expression) {
				ident, ok := expr.(*ast.Identifier)
				if !ok {
					t.Fatalf("expected *ast.Identifier, got %T", expr)
				}
				if ident.Value != "value" {
					t.Fatalf("expected identifier 'value', got %q", ident.Value)
				}
			},
			expectValue: true,
		},
		{
			name:  "ExitFunctionCall",
			input: "Exit Foo();",
			assert: func(t *testing.T, expr ast.Expression) {
				callExpr, ok := expr.(*ast.CallExpression)
				if !ok {
					t.Fatalf("expected *ast.CallExpression, got %T", expr)
				}

				fnIdent, ok := callExpr.Function.(*ast.Identifier)
				if !ok || fnIdent.Value != "Foo" {
					t.Fatalf("expected function identifier 'Foo', got %T (%v)", callExpr.Function, callExpr.Function)
				}

				if len(callExpr.Arguments) != 0 {
					t.Fatalf("expected no arguments, got %d", len(callExpr.Arguments))
				}
			},
			expectValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			exitStmt, ok := program.Statements[0].(*ast.ExitStatement)
			if !ok {
				t.Fatalf("statement is not *ast.ExitStatement, got %T", program.Statements[0])
			}

			if tt.expectValue {
				if exitStmt.ReturnValue == nil {
					t.Fatalf("expected ReturnValue, got nil")
				}
				tt.assert(t, exitStmt.ReturnValue)
			} else {
				if exitStmt.ReturnValue != nil {
					t.Fatalf("expected nil ReturnValue, got %T", exitStmt.ReturnValue)
				}
			}
		})
	}
}
