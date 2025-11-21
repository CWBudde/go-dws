package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestUserDefinedFunctionCallsWithArguments tests calling user-defined functions with arguments - Task 5.15
func TestUserDefinedFunctionCallsWithArguments(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "call user function with integer arguments",
			input: `
				function Add(a: Integer; b: Integer): Integer;
				begin
					end;

				begin
					Add(1, 2);
				end;
			`,
		},
		{
			name: "call user function with mixed argument types",
			input: `
				function Format(name: String; age: Integer): String;
				begin
				end;

				begin
					Format('John', 25);
				end;
			`,
		},
		{
			name: "call user function with expression arguments",
			input: `
				function Calculate(x: Integer; y: Integer): Integer;
				begin
				end;

				begin
					Calculate(1 + 2, 3 * 4);
				end;
			`,
		},
		{
			name: "call user function with no arguments",
			input: `
				function GetValue: Integer;
				begin
				end;

				begin
					GetValue();
				end;
			`,
		},
		{
			name: "call procedure with arguments",
			input: `
				procedure PrintValue(x: Integer);
				begin
				end;

				begin
					PrintValue(42);
				end;
			`,
		},
		{
			name: "nested function calls as arguments",
			input: `
				function Double(x: Integer): Integer;
				begin
				end;

				function Triple(x: Integer): Integer;
				begin
				end;

				begin
					Double(Triple(5));
				end;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			// The program should parse successfully
			if len(program.Statements) < 2 {
				t.Fatalf("expected at least 2 statements (function + main block), got %d", len(program.Statements))
			}

			// First statement(s) should be function declarations
			for i := 0; i < len(program.Statements)-1; i++ {
				if _, ok := program.Statements[i].(*ast.FunctionDecl); !ok {
					t.Errorf("statement %d is not *ast.FunctionDecl, got %T", i, program.Statements[i])
				}
			}

			// Last statement should be the main block containing the call
			lastStmt := program.Statements[len(program.Statements)-1]
			if _, ok := lastStmt.(*ast.BlockStatement); !ok {
				t.Errorf("last statement is not *ast.BlockStatement, got %T", lastStmt)
			}
		})
	}
}

// TestNewKeywordExpression tests parsing of 'new' keyword expressions
// The 'new' keyword creates a NewExpression: new T(args) -> NewExpression{ClassName: T, Arguments: args}
func TestNewKeywordExpression(t *testing.T) {
	tests := []struct {
		input    string
		typeName string
		numArgs  int
	}{
		{"new Exception('test');", "Exception", 1},
		{"new TPoint(10, 20);", "TPoint", 2},
		{"new TMyClass();", "TMyClass", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
			}

			// new T(args) should create a NewExpression
			newExpr, ok := stmt.Expression.(*ast.NewExpression)
			if !ok {
				t.Fatalf("expression is not ast.NewExpression. got=%T", stmt.Expression)
			}

			// Check class name
			if newExpr.ClassName.Value != tt.typeName {
				t.Fatalf("wrong class name. expected=%s, got=%s", tt.typeName, newExpr.ClassName.Value)
			}

			// Check number of arguments
			if len(newExpr.Arguments) != tt.numArgs {
				t.Fatalf("wrong number of arguments. expected=%d, got=%d", tt.numArgs, len(newExpr.Arguments))
			}
		})
	}
}
