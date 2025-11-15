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
					Target: NewTestIdentifier("Result"),
					Value: &BinaryExpression{
						TypedExpressionBase: TypedExpressionBase{
							BaseNode: BaseNode{Token: lexer.Token{Type: lexer.ASTERISK, Literal: "*"}},
						},
						Left:     &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}}}, Value: "x"},
						Operator: "*",
						Right:    &IntegerLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.INT, Literal: "2"}}}, Value: 2},
					},
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.ASSIGN, Literal: ":="}},
				},
			},
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
			Parameters: []*Parameter{
				{
					Name:     NewTestIdentifier("x"),
					Type:     NewTestTypeAnnotation("Integer"),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
				},
			},
			ReturnType:  NewTestTypeAnnotation("Integer"),
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
				TypedExpressionBase: TypedExpressionBase{
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.PLUS, Literal: "+"}},
				},
				Left:     &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}}}, Value: "x"},
				Operator: "+",
				Right:    &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "y"}}}, Value: "y"},
			},
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"}},
		}

		body := &BlockStatement{
			Statements: []Statement{returnStmt},
			BaseNode:   BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
			Parameters: []*Parameter{
				{
					Name:     NewTestIdentifier("x"),
					Type:     NewTestTypeAnnotation("Integer"),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
				},
				{
					Name:     NewTestIdentifier("y"),
					Type:     NewTestTypeAnnotation("Integer"),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "y"}},
				},
			},
			ReturnType:  nil, // Type inference
			Body:        body,
			IsShorthand: true,
		}

		// Test String() output - should show shorthand syntax
		str := node.String()
		if str != "lambda(x: Integer; y: Integer) => (x + y)" {
			t.Errorf("unexpected String() output: %q", str)
		}
	})

	t.Run("lambda with no parameters", func(t *testing.T) {
		body := &BlockStatement{
			Statements: []Statement{
				&ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{NewTestStringLiteral("Hello")},
					),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "PrintLn"}},
				},
			},
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode:    BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
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
					Expression: NewTestCallExpression(
						NewTestIdentifier("DoSomething"),
						[]Expression{NewTestIntegerLiteral(42)},
					),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "DoSomething"}},
				},
			},
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
			Parameters: []*Parameter{
				{
					Name:     NewTestIdentifier("n"),
					Type:     NewTestTypeAnnotation("Integer"),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "n"}},
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
						TypedExpressionBase: TypedExpressionBase{
							BaseNode: BaseNode{Token: lexer.Token{Type: lexer.PLUS, Literal: "+"}},
						},
						Left: &BinaryExpression{
							TypedExpressionBase: TypedExpressionBase{
								BaseNode: BaseNode{Token: lexer.Token{Type: lexer.PLUS, Literal: "+"}},
							},
							Left:     &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "a"}}}, Value: "a"},
							Operator: "+",
							Right:    &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "b"}}}, Value: "b"},
						},
						Operator: "+",
						Right:    &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "c"}}}, Value: "c"},
					},
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"}},
				},
			},
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
			Parameters: []*Parameter{
				{
					Name:     NewTestIdentifier("a"),
					Type:     NewTestTypeAnnotation("Integer"),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "a"}},
				},
				{
					Name:     NewTestIdentifier("b"),
					Type:     NewTestTypeAnnotation("Integer"),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "b"}},
				},
				{
					Name:     NewTestIdentifier("c"),
					Type:     NewTestTypeAnnotation("Integer"),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "c"}},
				},
			},
			ReturnType:  NewTestTypeAnnotation("Integer"),
			Body:        body,
			IsShorthand: true,
		}

		str := node.String()
		if str != "lambda(a: Integer; b: Integer; c: Integer): Integer => ((a + b) + c)" {
			t.Errorf("unexpected String() output: %q", str)
		}
	})

	t.Run("lambda with by-ref parameter", func(t *testing.T) {
		body := &BlockStatement{
			Statements: []Statement{
				&AssignmentStatement{
					Target:   NewTestIdentifier("x"),
					Value:    NewTestIntegerLiteral(100),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.ASSIGN, Literal: ":="}},
				},
			},
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
			Parameters: []*Parameter{
				{
					Name:     NewTestIdentifier("x"),
					Type:     NewTestTypeAnnotation("Integer"),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.VAR, Literal: "var"}},
					ByRef:    true,
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
			BaseNode:    BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
			Parameters:  []*Parameter{},
			ReturnType:  nil,
			Body:        &BlockStatement{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}}, Statements: []Statement{}},
			IsShorthand: false,
		}

		// Verify it implements TypedExpression interface
		var _ TypedExpression = node

		// Test SetType and GetType
		newType := NewTestTypeAnnotation("TCallback")
		node.SetType(newType)
		if node.GetType() != newType {
			t.Error("SetType/GetType not working correctly")
		}
	})

	t.Run("lambda position tracking", func(t *testing.T) {
		pos := lexer.Position{Line: 10, Column: 5, Offset: 150}
		node := &LambdaExpression{
			BaseNode:    BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda", Pos: pos}},
			Parameters:  []*Parameter{},
			ReturnType:  nil,
			Body:        &BlockStatement{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}}, Statements: []Statement{}},
			IsShorthand: false,
		}

		if node.Pos() != pos {
			t.Errorf("expected position %v, got %v", pos, node.Pos())
		}
	})

	t.Run("lambda empty body", func(t *testing.T) {
		node := &LambdaExpression{
			BaseNode:    BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
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
					ReturnValue: &ArrayLiteralExpression{
						Elements: []Expression{
							NewTestIntegerLiteral(1),
							NewTestIntegerLiteral(2),
						},
						BaseNode: BaseNode{Token: lexer.Token{Type: lexer.LBRACK, Literal: "["}},
					},
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"}},
				},
			},
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode:    BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
			Parameters:  []*Parameter{},
			ReturnType:  NewTestTypeAnnotation("TIntArray"),
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
					Expression: NewTestCallExpression(
						NewTestIdentifier("DoIt"),
						[]Expression{},
					),
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "DoIt"}},
				},
			},
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode:    BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
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
			BaseNode:   BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode:    BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
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
			BaseNode:   BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
		}

		node := &LambdaExpression{
			BaseNode:    BaseNode{Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda"}},
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
