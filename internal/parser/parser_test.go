package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
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
			expectedError: "expected ';' after variable declaration",
		},
		{
			name:          "unclosed parentheses",
			input:         "(3 + 5",
			expectedError: "expected ')', got EOF",
		},
		{
			name:          "invalid prefix operator",
			input:         "};",
			expectedError: "no prefix parse function",
		},
		{
			name:          "missing identifier in var declaration",
			input:         "var ;",
			expectedError: "expected identifier in var declaration",
		},
		{
			name:          "missing expression in if condition",
			input:         "if then x := 1;",
			expectedError: "no prefix parse function",
		},
		{
			name:          "missing then keyword in if",
			input:         "if x > 0 x := 1;",
			expectedError: "expected 'then' after if condition",
		},
		{
			name:          "missing do keyword in while",
			input:         "while x < 10 x := x + 1;",
			expectedError: "expected 'do' after while condition",
		},
		{
			name:          "missing until keyword in repeat",
			input:         "repeat x := x + 1 x >= 10;",
			expectedError: "expected 'until' after repeat body",
		},
		{
			name:          "missing identifier in for loop",
			input:         "for := 1 to 10 do PrintLn(i);",
			expectedError: "expected identifier after 'for'",
		},
		{
			name:          "missing assign in for loop",
			input:         "for i = 1 to 10 do PrintLn(i);",
			expectedError: "expected ':=' after for loop variable",
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
			expectedError: "expected 'of' after case expression",
		},
		{
			name:          "missing colon in case branch",
			input:         "case x of 1 x := 1; end;",
			expectedError: "expected ':' after case value",
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

// TestParserLexerErrors tests that parser can access lexer errors (Task 12.1.4)
func TestParserLexerErrors(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedLexerErrs  int
		expectedParserErrs int
	}{
		{
			name:               "Unterminated string",
			input:              `var x := 'hello`,
			expectedLexerErrs:  1,
			expectedParserErrs: 0, // Parser may or may not add errors
		},
		{
			name:               "Unterminated comment",
			input:              `var x := 5; { comment`,
			expectedLexerErrs:  1,
			expectedParserErrs: 0,
		},
		{
			name:               "Valid input",
			input:              `var x := 5;`,
			expectedLexerErrs:  0,
			expectedParserErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			lexerErrs := p.LexerErrors()
			if len(lexerErrs) != tt.expectedLexerErrs {
				t.Errorf("expected %d lexer errors, got %d", tt.expectedLexerErrs, len(lexerErrs))
				for i, err := range lexerErrs {
					t.Logf("  lexer error[%d]: %s", i, err.Message)
				}
			}
		})
	}
}

// TestLookaheadFunctions tests the looksLikeVarDeclaration and looksLikeConstDeclaration
// functions which use Peek(0) for 2-token lookahead
func TestLookaheadFunctions(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		shouldParse    bool
		expectedStmts  int
		expectedErrors int
	}{
		{
			name: "var declaration continuation with colon",
			input: `var
				x : Integer;
				y : String;
				z : Boolean;`,
			shouldParse:   true,
			expectedStmts: 3, // 3 separate var declarations (unwrapped)
		},
		{
			name: "var declaration with inferred type needs var keyword",
			input: `var x := 42;
				var y := "hello";`,
			shouldParse:   true,
			expectedStmts: 2, // Separate statements since := is ambiguous
		},
		{
			name: "var declaration with comma-separated names",
			input: `var
				x, y : Integer;
				a, b, c : String;`,
			shouldParse:   true,
			expectedStmts: 2, // 2 separate var declarations (unwrapped)
		},
		{
			name: "const declaration continuation with colon",
			input: `const
				MAX : Integer = 100;
				NAME : String = "test";`,
			shouldParse:   true,
			expectedStmts: 1,
		},
		{
			name: "const declaration continuation with equals",
			input: `const
				MAX = 100;
				PI = 3.14;`,
			shouldParse:   true,
			expectedStmts: 1,
		},
		{
			name: "mixed var declarations should not continue incorrectly",
			input: `var x : Integer;
				DoSomething();`,
			shouldParse:   true,
			expectedStmts: 2, // var decl + expression statement
		},
		{
			name: "identifier followed by lparen should not be var decl",
			input: `var x : Integer;
				MyFunction(x);`,
			shouldParse:   true,
			expectedStmts: 2,
		},
		{
			name:          "single var declaration",
			input:         `var x : Integer;`,
			shouldParse:   true,
			expectedStmts: 1,
		},
		{
			name:          "single const declaration",
			input:         `const MAX = 100;`,
			shouldParse:   true,
			expectedStmts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()

			errors := p.Errors()
			if tt.expectedErrors > 0 {
				if len(errors) != tt.expectedErrors {
					t.Errorf("expected %d errors, got %d: %v", tt.expectedErrors, len(errors), errors)
				}
			} else if len(errors) > 0 {
				t.Errorf("expected no errors, got %d: %v", len(errors), errors)
			}

			if tt.shouldParse {
				if program == nil {
					t.Fatal("ParseProgram() returned nil")
				}
				if len(program.Statements) != tt.expectedStmts {
					t.Errorf("expected %d statements, got %d", tt.expectedStmts, len(program.Statements))
					for i, stmt := range program.Statements {
						t.Logf("  stmt[%d]: %T", i, stmt)
					}
				}
			}
		})
	}
}
