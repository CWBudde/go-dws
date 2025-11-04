package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "invalid integer literal",
			input:         "999999999999999999999999999999;",
			expectedError: "could not parse",
		},
		{
			name:          "invalid float literal",
			input:         "99999999999999999999999999999.9e999;",
			expectedError: "could not parse",
		},
		{
			name:          "missing semicolon after var declaration",
			input:         "var x: Integer",
			expectedError: "expected next token to be SEMICOLON",
		},
		{
			name:          "unclosed parentheses",
			input:         "(3 + 5",
			expectedError: "expected next token to be RPAREN",
		},
		{
			name:          "invalid prefix operator",
			input:         "};",
			expectedError: "no prefix parse function",
		},
		{
			name:          "missing identifier in var declaration",
			input:         "var ;",
			expectedError: "expected next token to be IDENT",
		},
		{
			name:          "missing expression in if condition",
			input:         "if then x := 1;",
			expectedError: "no prefix parse function",
		},
		{
			name:          "missing then keyword in if",
			input:         "if x > 0 x := 1;",
			expectedError: "expected next token to be THEN",
		},
		{
			name:          "missing do keyword in while",
			input:         "while x < 10 x := x + 1;",
			expectedError: "expected next token to be DO",
		},
		{
			name:          "missing until keyword in repeat",
			input:         "repeat x := x + 1 x >= 10;",
			expectedError: "expected 'until' after repeat body",
		},
		{
			name:          "missing identifier in for loop",
			input:         "for := 1 to 10 do PrintLn(i);",
			expectedError: "expected next token to be IDENT",
		},
		{
			name:          "missing assign in for loop",
			input:         "for i = 1 to 10 do PrintLn(i);",
			expectedError: "expected next token to be ASSIGN",
		},
		{
			name:          "missing direction in for loop",
			input:         "for i := 1 10 do PrintLn(i);",
			expectedError: "expected 'to' or 'downto'",
		},
		{
			name:          "missing expression after case",
			input:         "case of 1: x := 1; end;",
			expectedError: "no prefix parse function",
		},
		{
			name:          "missing of keyword in case",
			input:         "case x 1: x := 1; end;",
			expectedError: "expected next token to be OF",
		},
		{
			name:          "missing colon in case branch",
			input:         "case x of 1 x := 1; end;",
			expectedError: "expected next token to be COLON",
		},
		{
			name:          "missing end keyword in case",
			input:         "case x of 1: x := 1;",
			expectedError: "expected 'end' to close case statement",
		},
		{
			name:          "missing end keyword in block",
			input:         "begin x := 1; y := 2;",
			expectedError: "expected 'end' to close block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Fatalf("expected parser errors, got none")
			}

			found := false
			for _, err := range errors {
				if contains(err, tt.expectedError) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error containing %q, got %v", tt.expectedError, errors)
			}
		})
	}
}

// TestNilLiteral tests parsing of nil literals.
func TestNilLiteral(t *testing.T) {
	input := "nil;"

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

	_, ok = stmt.Expression.(*ast.NilLiteral)
	if !ok {
		t.Fatalf("expression is not ast.NilLiteral. got=%T", stmt.Expression)
	}
}

// TestFunctionDeclarations tests parsing of function declarations.
