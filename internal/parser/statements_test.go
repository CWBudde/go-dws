package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestExitStatementVariants(t *testing.T) {
	tests := []struct {
		assert      func(t *testing.T, expr ast.Expression)
		name        string
		input       string
		expectValue bool
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

func TestBlockStatement(t *testing.T) {
	input := `
begin
  5;
  10;
end;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	block, ok := program.Statements[0].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("statement is not ast.BlockStatement. got=%T", program.Statements[0])
	}

	if len(block.Statements) != 2 {
		t.Fatalf("block has wrong number of statements. got=%d", len(block.Statements))
	}

	for i, stmt := range block.Statements {
		exprStmt, ok := stmt.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("block statement %d is not ExpressionStatement. got=%T", i, stmt)
		}
		if !testIntegerLiteral(t, exprStmt.Expression, int64((i*5)+5)) {
			return
		}
	}
}

// TestBlockStatementAssignments ensures assignments inside blocks are parsed correctly.
func TestBlockStatementAssignments(t *testing.T) {
	input := `
begin
  x := 1;
  y := x + 2;
end;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	block, ok := program.Statements[0].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("statement is not ast.BlockStatement. got=%T", program.Statements[0])
	}

	if len(block.Statements) != 2 {
		t.Fatalf("block has wrong number of statements. got=%d", len(block.Statements))
	}

	first, ok := block.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("first statement not AssignmentStatement. got=%T", block.Statements[0])
	}
	firstTarget, ok := first.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("first.Target not *ast.Identifier. got=%T", first.Target)
	}
	if firstTarget.Value != "x" {
		t.Errorf("first assignment name = %q, want %q", firstTarget.Value, "x")
	}
	if !testIntegerLiteral(t, first.Value, 1) {
		return
	}

	second, ok := block.Statements[1].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("second statement not AssignmentStatement. got=%T", block.Statements[1])
	}
	secondTarget, ok := second.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("second.Target not *ast.Identifier. got=%T", second.Target)
	}
	if secondTarget.Value != "y" {
		t.Errorf("second assignment name = %q, want %q", secondTarget.Value, "y")
	}
	if !testInfixExpression(t, second.Value, "x", "+", 2) {
		return
	}
}

// TestVarDeclarations tests parsing of variable declarations.

func TestStatementDispatch(t *testing.T) {
	input := `
var foo := 1;
foo();
foo := foo + 1;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	if _, ok := program.Statements[0].(*ast.VarDeclStatement); !ok {
		t.Fatalf("statement 0 expected VarDeclStatement. got=%T", program.Statements[0])
	}

	callStmt, ok := program.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement 1 expected ExpressionStatement. got=%T", program.Statements[1])
	}
	if _, ok := callStmt.Expression.(*ast.CallExpression); !ok {
		t.Fatalf("statement 1 expression expected CallExpression. got=%T", callStmt.Expression)
	}

	if _, ok := program.Statements[2].(*ast.AssignmentStatement); !ok {
		t.Fatalf("statement 2 expected AssignmentStatement. got=%T", program.Statements[2])
	}

	assign := program.Statements[2].(*ast.AssignmentStatement)
	if !testInfixExpression(t, assign.Value, "foo", "+", 1) {
		return
	}
}

// Helper function to test literal expressions

// TestCompleteSimplePrograms tests parsing of complete simple programs with multiple statement types.
