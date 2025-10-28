package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestLambdaExpression tests the LambdaExpression AST node.
func TestLambdaExpression(t *testing.T) {
	t.Run("lambda with parameters and return type (full syntax)", func(t *testing.T) {
		// Create a simple body: Result := x * 2
		body := &BlockStatement{
			Statements: []Statement{
				&AssignmentStatement{
					Target: &Identifier{Value: "Result", Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"}},
					Value: &BinaryExpression{
						Left:     &Identifier{Value: "x", Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
						Operator: "*",
						Right:    &IntegerLiteral{Value: 2, Token: lexer.Token{Type: lexer.INT, Literal: "2"}},
						Token:    lexer.Token{Type: lexer.ASTERISK, Literal: "*"},
					},
					Token: lexer.Token{Type: lexer.ASSIGN, Literal: ":="},
				},
			},
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters: []*Parameter{
				{
					Name:  &Identifier{Value: "x", Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				},
			},
			ReturnType:  &TypeAnnotation{Name: "Integer"},
			Body:        body,
			IsShorthand: false,
		}

		// Test String() output
		str := node.String()
		if str != "lambda(x: Integer): Integer begin Result := (x * 2); end" {
			t.Errorf("unexpected String() output: %q", str)
		}

		// Test TokenLiteral
		if node.TokenLiteral() != "lambda" {
			t.Errorf("expected token literal 'lambda', got %q", node.TokenLiteral())
		}

		// Verify it implements Expression interface
		var _ Expression = node
	})

	t.Run("lambda with shorthand syntax", func(t *testing.T) {
		// Shorthand lambda desugared to: lambda(x, y) => x + y
		// Internally stored as: begin Result := x + y; end
		returnStmt := &ReturnStatement{
			ReturnValue: &BinaryExpression{
				Left:     &Identifier{Value: "x", Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
				Operator: "+",
				Right:    &Identifier{Value: "y", Token: lexer.Token{Type: lexer.IDENT, Literal: "y"}},
				Token:    lexer.Token{Type: lexer.PLUS, Literal: "+"},
			},
			Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"},
		}

		body := &BlockStatement{
			Statements: []Statement{returnStmt},
			Token:      lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters: []*Parameter{
				{
					Name:  &Identifier{Value: "x", Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				},
				{
					Name:  &Identifier{Value: "y", Token: lexer.Token{Type: lexer.IDENT, Literal: "y"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "y"},
				},
			},
			ReturnType:  nil, // Type inference
			Body:        body,
			IsShorthand: true,
		}

		// Test String() output - should show shorthand syntax
		str := node.String()
		if str != "lambda(x: Integer, y: Integer) => (x + y)" {
			t.Errorf("unexpected String() output: %q", str)
		}
	})

	t.Run("lambda with no parameters", func(t *testing.T) {
		body := &BlockStatement{
			Statements: []Statement{
				&ExpressionStatement{
					Expression: &CallExpression{
						Function: &Identifier{Value: "PrintLn", Token: lexer.Token{Type: lexer.IDENT, Literal: "PrintLn"}},
						Arguments: []Expression{
							&StringLiteral{Value: "Hello", Token: lexer.Token{Type: lexer.STRING, Literal: "'Hello'"}},
						},
						Token: lexer.Token{Type: lexer.LPAREN, Literal: "("},
					},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "PrintLn"},
				},
			},
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token:       lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters:  []*Parameter{},
			ReturnType:  nil,
			Body:        body,
			IsShorthand: false,
		}

		str := node.String()
		// Note: StringLiteral.String() uses double quotes, not single quotes
		if str != "lambda() begin PrintLn(\"Hello\"); end" {
			t.Errorf("unexpected String() output: %q", str)
		}
	})

	t.Run("lambda with no return type (procedure)", func(t *testing.T) {
		body := &BlockStatement{
			Statements: []Statement{
				&ExpressionStatement{
					Expression: &CallExpression{
						Function: &Identifier{Value: "DoSomething", Token: lexer.Token{Type: lexer.IDENT, Literal: "DoSomething"}},
						Arguments: []Expression{
							&IntegerLiteral{Value: 42, Token: lexer.Token{Type: lexer.INT, Literal: "42"}},
						},
						Token: lexer.Token{Type: lexer.LPAREN, Literal: "("},
					},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "DoSomething"},
				},
			},
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters: []*Parameter{
				{
					Name:  &Identifier{Value: "n", Token: lexer.Token{Type: lexer.IDENT, Literal: "n"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "n"},
				},
			},
			ReturnType:  nil, // Procedure lambda
			Body:        body,
			IsShorthand: false,
		}

		str := node.String()
		if str != "lambda(n: Integer) begin DoSomething(42); end" {
			t.Errorf("unexpected String() output: %q", str)
		}
	})

	t.Run("lambda with multiple parameters", func(t *testing.T) {
		body := &BlockStatement{
			Statements: []Statement{
				&ReturnStatement{
					ReturnValue: &BinaryExpression{
						Left: &BinaryExpression{
							Left:     &Identifier{Value: "a", Token: lexer.Token{Type: lexer.IDENT, Literal: "a"}},
							Operator: "+",
							Right:    &Identifier{Value: "b", Token: lexer.Token{Type: lexer.IDENT, Literal: "b"}},
							Token:    lexer.Token{Type: lexer.PLUS, Literal: "+"},
						},
						Operator: "+",
						Right:    &Identifier{Value: "c", Token: lexer.Token{Type: lexer.IDENT, Literal: "c"}},
						Token:    lexer.Token{Type: lexer.PLUS, Literal: "+"},
					},
					Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"},
				},
			},
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters: []*Parameter{
				{
					Name:  &Identifier{Value: "a", Token: lexer.Token{Type: lexer.IDENT, Literal: "a"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "a"},
				},
				{
					Name:  &Identifier{Value: "b", Token: lexer.Token{Type: lexer.IDENT, Literal: "b"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "b"},
				},
				{
					Name:  &Identifier{Value: "c", Token: lexer.Token{Type: lexer.IDENT, Literal: "c"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "c"},
				},
			},
			ReturnType:  &TypeAnnotation{Name: "Integer"},
			Body:        body,
			IsShorthand: true,
		}

		str := node.String()
		if str != "lambda(a: Integer, b: Integer, c: Integer): Integer => ((a + b) + c)" {
			t.Errorf("unexpected String() output: %q", str)
		}
	})

	t.Run("lambda with by-ref parameter", func(t *testing.T) {
		body := &BlockStatement{
			Statements: []Statement{
				&AssignmentStatement{
					Target: &Identifier{Value: "x", Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
					Value:  &IntegerLiteral{Value: 100, Token: lexer.Token{Type: lexer.INT, Literal: "100"}},
					Token:  lexer.Token{Type: lexer.ASSIGN, Literal: ":="},
				},
			},
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters: []*Parameter{
				{
					Name:  &Identifier{Value: "x", Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
					ByRef: true,
				},
			},
			ReturnType:  nil,
			Body:        body,
			IsShorthand: false,
		}

		str := node.String()
		if str != "lambda(var x: Integer) begin x := 100; end" {
			t.Errorf("unexpected String() output: %q", str)
		}
	})

	t.Run("lambda implements TypedExpression", func(t *testing.T) {
		node := &LambdaExpression{
			Token:       lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters:  []*Parameter{},
			ReturnType:  nil,
			Body:        &BlockStatement{Statements: []Statement{}, Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
			IsShorthand: false,
		}

		// Verify it implements TypedExpression interface
		var _ TypedExpression = node

		// Test SetType and GetType
		newType := &TypeAnnotation{Name: "TCallback"}
		node.SetType(newType)
		if node.GetType() != newType {
			t.Error("SetType/GetType not working correctly")
		}
	})

	t.Run("lambda position tracking", func(t *testing.T) {
		pos := lexer.Position{Line: 10, Column: 5, Offset: 150}
		node := &LambdaExpression{
			Token:       lexer.Token{Type: lexer.LAMBDA, Literal: "lambda", Pos: pos},
			Parameters:  []*Parameter{},
			ReturnType:  nil,
			Body:        &BlockStatement{Statements: []Statement{}, Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
			IsShorthand: false,
		}

		if node.Pos() != pos {
			t.Errorf("expected position %v, got %v", pos, node.Pos())
		}
	})

	t.Run("lambda empty body", func(t *testing.T) {
		node := &LambdaExpression{
			Token:       lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters:  []*Parameter{},
			ReturnType:  nil,
			Body:        nil,
			IsShorthand: false,
		}

		str := node.String()
		if str != "lambda() begin end" {
			t.Errorf("unexpected String() output for empty body: %q", str)
		}
	})

	t.Run("lambda with array return type", func(t *testing.T) {
		body := &BlockStatement{
			Statements: []Statement{
				&ReturnStatement{
					ReturnValue: &ArrayLiteral{
						Elements: []Expression{
							&IntegerLiteral{Value: 1, Token: lexer.Token{Type: lexer.INT, Literal: "1"}},
							&IntegerLiteral{Value: 2, Token: lexer.Token{Type: lexer.INT, Literal: "2"}},
						},
						Token: lexer.Token{Type: lexer.LBRACK, Literal: "["},
					},
					Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"},
				},
			},
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token:      lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters: []*Parameter{},
			ReturnType: &TypeAnnotation{
				Name: "TIntArray",
			},
			Body:        body,
			IsShorthand: true,
		}

		// Just verify it doesn't panic and produces some output
		str := node.String()
		if len(str) == 0 {
			t.Error("expected non-empty string output")
		}
		// Should contain the type name
		if !stringContains(str, "TIntArray") {
			t.Errorf("expected string to contain 'TIntArray', got: %q", str)
		}
	})
}

// stringContains checks if s contains substr
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestLambdaExpressionEdgeCases tests edge cases and corner cases.
func TestLambdaExpressionEdgeCases(t *testing.T) {
	t.Run("shorthand lambda with statement instead of return", func(t *testing.T) {
		// Edge case: shorthand lambda but body contains a statement, not a return
		body := &BlockStatement{
			Statements: []Statement{
				&ExpressionStatement{
					Expression: &CallExpression{
						Function:  &Identifier{Value: "DoIt", Token: lexer.Token{Type: lexer.IDENT, Literal: "DoIt"}},
						Arguments: []Expression{},
						Token:     lexer.Token{Type: lexer.LPAREN, Literal: "("},
					},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "DoIt"},
				},
			},
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token:       lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters:  []*Parameter{},
			ReturnType:  nil,
			Body:        body,
			IsShorthand: true,
		}

		str := node.String()
		// Should handle this gracefully
		if str != "lambda() => DoIt()" {
			t.Errorf("unexpected String() output: %q", str)
		}
	})

	t.Run("full lambda with empty body", func(t *testing.T) {
		body := &BlockStatement{
			Statements: []Statement{},
			Token:      lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token:       lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters:  []*Parameter{},
			ReturnType:  nil,
			Body:        body,
			IsShorthand: false,
		}

		str := node.String()
		if str != "lambda() begin end" {
			t.Errorf("unexpected String() output: %q", str)
		}
	})

	t.Run("shorthand lambda with empty body", func(t *testing.T) {
		body := &BlockStatement{
			Statements: []Statement{},
			Token:      lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		}

		node := &LambdaExpression{
			Token:       lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"},
			Parameters:  []*Parameter{},
			ReturnType:  nil,
			Body:        body,
			IsShorthand: true,
		}

		str := node.String()
		// Should handle empty shorthand lambda
		if len(str) == 0 {
			t.Error("expected non-empty output")
		}
	})
}
