package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestParseLambdaExpressions tests parsing of lambda expressions.
// Tasks 9.212-9.215: Parser support for lambda expressions
func TestParseLambdaExpressions(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		paramCount      int
		hasReturnType   bool
		returnTypeName  string
		isShorthand     bool
		firstParamName  string
		firstParamType  string
		firstParamByRef bool
	}{
		{
			name:           "shorthand lambda with one parameter",
			input:          "lambda(x) => x * 2",
			paramCount:     1,
			hasReturnType:  false,
			isShorthand:    true,
			firstParamName: "x",
			firstParamType: "",
		},
		{
			name:           "shorthand lambda with typed parameter",
			input:          "lambda(x: Integer) => x * 2",
			paramCount:     1,
			hasReturnType:  false,
			isShorthand:    true,
			firstParamName: "x",
			firstParamType: "Integer",
		},
		{
			name:           "shorthand lambda with return type",
			input:          "lambda(x: Integer): Integer => x * 2",
			paramCount:     1,
			hasReturnType:  true,
			returnTypeName: "Integer",
			isShorthand:    true,
			firstParamName: "x",
			firstParamType: "Integer",
		},
		{
			name:           "shorthand lambda with multiple parameters (shared type)",
			input:          "lambda(a, b: Integer) => a + b",
			paramCount:     2,
			hasReturnType:  false,
			isShorthand:    true,
			firstParamName: "a",
			firstParamType: "Integer",
		},
		{
			name:          "shorthand lambda with no parameters",
			input:         "lambda() => 42",
			paramCount:    0,
			hasReturnType: false,
			isShorthand:   true,
		},
		{
			name:           "full lambda with begin/end",
			input:          "lambda(x: Integer): Integer begin Result := x * 2; end",
			paramCount:     1,
			hasReturnType:  true,
			returnTypeName: "Integer",
			isShorthand:    false,
			firstParamName: "x",
			firstParamType: "Integer",
		},
		{
			name:           "full lambda no return type",
			input:          "lambda(x: Integer) begin PrintLn(x); end",
			paramCount:     1,
			hasReturnType:  false,
			isShorthand:    false,
			firstParamName: "x",
			firstParamType: "Integer",
		},
		{
			name:          "full lambda with no parameters",
			input:         "lambda() begin PrintLn('Hello'); end",
			paramCount:    0,
			hasReturnType: false,
			isShorthand:   false,
		},
		{
			name:            "lambda with by-ref parameter",
			input:           "lambda(var x: Integer) begin x := x + 1; end",
			paramCount:      1,
			hasReturnType:   false,
			isShorthand:     false,
			firstParamName:  "x",
			firstParamType:  "Integer",
			firstParamByRef: true,
		},
		{
			name:           "lambda with semicolon-separated parameters",
			input:          "lambda(x: Integer; y: Integer): Integer => x + y",
			paramCount:     2,
			hasReturnType:  true,
			returnTypeName: "Integer",
			isShorthand:    true,
			firstParamName: "x",
			firstParamType: "Integer",
		},
		{
			name:           "lambda with mixed parameter groups",
			input:          "lambda(a, b: Integer; c: String): String => c",
			paramCount:     3,
			hasReturnType:  true,
			returnTypeName: "String",
			isShorthand:    true,
			firstParamName: "a",
			firstParamType: "Integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)

			// Parse as expression
			expr := p.parseExpression(LOWEST)
			checkParserErrors(t, p)

			lambdaExpr, ok := expr.(*ast.LambdaExpression)
			if !ok {
				t.Fatalf("expected *ast.LambdaExpression, got %T", expr)
			}

			// Check parameter count
			if len(lambdaExpr.Parameters) != tt.paramCount {
				t.Errorf("expected %d parameters, got %d", tt.paramCount, len(lambdaExpr.Parameters))
			}

			// Check first parameter if present
			if tt.paramCount > 0 && tt.firstParamName != "" {
				if lambdaExpr.Parameters[0].Name.Value != tt.firstParamName {
					t.Errorf("first param name: expected %q, got %q",
						tt.firstParamName, lambdaExpr.Parameters[0].Name.Value)
				}

				if tt.firstParamType != "" {
					if lambdaExpr.Parameters[0].Type == nil {
						t.Error("expected parameter type, got nil")
					} else if lambdaExpr.Parameters[0].Type.Name != tt.firstParamType {
						t.Errorf("first param type: expected %q, got %q",
							tt.firstParamType, lambdaExpr.Parameters[0].Type.Name)
					}
				} else {
					if lambdaExpr.Parameters[0].Type != nil {
						t.Errorf("expected no parameter type, got %q", lambdaExpr.Parameters[0].Type.Name)
					}
				}

				if lambdaExpr.Parameters[0].ByRef != tt.firstParamByRef {
					t.Errorf("first param byRef: expected %v, got %v",
						tt.firstParamByRef, lambdaExpr.Parameters[0].ByRef)
				}
			}

			// Check return type
			if tt.hasReturnType {
				if lambdaExpr.ReturnType == nil {
					t.Error("expected return type, got nil")
				} else if lambdaExpr.ReturnType.Name != tt.returnTypeName {
					t.Errorf("return type: expected %q, got %q",
						tt.returnTypeName, lambdaExpr.ReturnType.Name)
				}
			} else {
				if lambdaExpr.ReturnType != nil {
					t.Errorf("expected no return type, got %q", lambdaExpr.ReturnType.Name)
				}
			}

			// Check syntax type
			if lambdaExpr.IsShorthand != tt.isShorthand {
				t.Errorf("isShorthand: expected %v, got %v", tt.isShorthand, lambdaExpr.IsShorthand)
			}

			// Check body is present
			if lambdaExpr.Body == nil {
				t.Error("expected body, got nil")
			}
		})
	}
}

// TestParseLambdaInExpressionContext tests lambda expressions in various contexts.
// Task 9.214: Handle lambda in expression context
func TestParseLambdaInExpressionContext(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "lambda in function call",
			input: "Map(arr, lambda(x) => x * 2);",
		},
		{
			name:  "lambda in variable assignment",
			input: "var f := lambda(x: Integer): Integer => x * 2;",
		},
		{
			name:  "lambda as function argument",
			input: "ProcessList(items, lambda(item) begin PrintLn(item); end);",
		},
		{
			name:  "multiple lambdas as arguments",
			input: "Transform(data, lambda(x) => x + 1, lambda(y) => y * 2);",
		},
		{
			name:  "nested lambda",
			input: "var f := lambda(x) => lambda(y) => x + y;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) == 0 {
				t.Fatal("expected at least one statement")
			}

			// Just verify it parsed without errors
			// Detailed structure verification is done in other tests
		})
	}
}

// TestParseLambdaEdgeCases tests edge cases and error conditions.
func TestParseLambdaEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "lambda with no body",
			input:       "lambda(x)",
			expectError: true,
		},
		{
			name:        "lambda with invalid syntax",
			input:       "lambda(x) :",
			expectError: true,
		},
		{
			name:        "lambda missing closing paren",
			input:       "lambda(x => x * 2",
			expectError: true,
		},
		{
			name:        "lambda with complex expression",
			input:       "lambda(x, y) => (x + y) * (x - y)",
			expectError: false,
		},
		{
			name:        "lambda with string literal",
			input:       "lambda() => 'Hello World'",
			expectError: false,
		},
		{
			name:        "lambda with function call in body",
			input:       "lambda(x) => PrintLn(x)",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.parseExpression(LOWEST)

			hasErrors := len(p.Errors()) > 0
			if hasErrors != tt.expectError {
				if tt.expectError {
					t.Errorf("expected parser errors, got none")
				} else {
					t.Errorf("unexpected parser errors: %v", p.Errors())
				}
			}
		})
	}
}

// TestParseLambdaComplexCases tests more complex lambda usage patterns.
func TestParseLambdaComplexCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "lambda with array literal",
			input: "lambda() => [1, 2, 3]",
		},
		{
			name:  "lambda with multiple statements",
			input: "lambda(x: Integer) begin var y := x * 2; Result := y + 1; end",
		},
		{
			name:  "lambda with binary expression",
			input: "lambda(x, y) => x * y + (x - y)",
		},
		{
			name:  "lambda returning lambda",
			input: "lambda(x) => lambda(y) => x + y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			expr := p.parseExpression(LOWEST)
			checkParserErrors(t, p)

			if _, ok := expr.(*ast.LambdaExpression); !ok {
				t.Fatalf("expected *ast.LambdaExpression, got %T", expr)
			}
		})
	}
}
