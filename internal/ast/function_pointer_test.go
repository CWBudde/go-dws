package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestFunctionPointerTypeNode tests the FunctionPointerTypeNode AST node.
func TestFunctionPointerTypeNode(t *testing.T) {
	t.Run("function pointer with parameters and return type", func(t *testing.T) {
		node := &FunctionPointerTypeNode{
			Parameters: []*Parameter{
				{
					Name:  NewTestIdentifier("a"),
					Type:  NewTestTypeAnnotation("Integer"),
					Token: lexer.Token{Type: lexer.IDENT, Literal: "a"},
				},
				{
					Name:  NewTestIdentifier("b"),
					Type:  NewTestTypeAnnotation("Integer"),
					Token: lexer.Token{Type: lexer.IDENT, Literal: "b"},
				},
			},
			ReturnType: NewTestTypeAnnotation("Integer"),
			Token:      lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
			OfObject:   false,
		}

		expected := "function(a: Integer; b: Integer): Integer"
		if node.String() != expected {
			t.Errorf("expected %q, got %q", expected, node.String())
		}

		if node.TokenLiteral() != "function" {
			t.Errorf("expected token literal 'function', got %q", node.TokenLiteral())
		}
	})

	t.Run("procedure pointer with one parameter", func(t *testing.T) {
		node := &FunctionPointerTypeNode{
			Parameters: []*Parameter{
				{
					Name:  NewTestIdentifier("msg"),
					Type:  NewTestTypeAnnotation("String"),
					Token: lexer.Token{Type: lexer.IDENT, Literal: "msg"},
				},
			},
			ReturnType: nil,
			Token:      lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
			OfObject:   false,
		}

		expected := "procedure(msg: String)"
		if node.String() != expected {
			t.Errorf("expected %q, got %q", expected, node.String())
		}
	})

	t.Run("procedure pointer with no parameters", func(t *testing.T) {
		node := &FunctionPointerTypeNode{
			Parameters: []*Parameter{},
			ReturnType: nil,
			Token:      lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
			OfObject:   false,
		}

		expected := "procedure()"
		if node.String() != expected {
			t.Errorf("expected %q, got %q", expected, node.String())
		}
	})

	t.Run("method pointer (of object)", func(t *testing.T) {
		node := &FunctionPointerTypeNode{
			Parameters: []*Parameter{
				{
					Name:  NewTestIdentifier("Sender"),
					Type:  NewTestTypeAnnotation("TObject"),
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Sender"},
				},
			},
			ReturnType: nil,
			Token:      lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
			OfObject:   true,
		}

		expected := "procedure(Sender: TObject) of object"
		if node.String() != expected {
			t.Errorf("expected %q, got %q", expected, node.String())
		}
	})

	t.Run("function pointer of object", func(t *testing.T) {
		node := &FunctionPointerTypeNode{
			Parameters: []*Parameter{
				{
					Name:  NewTestIdentifier("x"),
					Type:  NewTestTypeAnnotation("Integer"),
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				},
			},
			ReturnType: NewTestTypeAnnotation("Boolean"),
			Token:      lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
			OfObject:   true,
		}

		expected := "function(x: Integer): Boolean of object"
		if node.String() != expected {
			t.Errorf("expected %q, got %q", expected, node.String())
		}
	})

	t.Run("function pointer with by-ref parameter", func(t *testing.T) {
		node := &FunctionPointerTypeNode{
			Parameters: []*Parameter{
				{
					Name:  NewTestIdentifier("x"),
					Type:  NewTestTypeAnnotation("Integer"),
					Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
					ByRef: true,
				},
			},
			ReturnType: NewTestTypeAnnotation("Boolean"),
			Token:      lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
			OfObject:   false,
		}

		expected := "function(var x: Integer): Boolean"
		if node.String() != expected {
			t.Errorf("expected %q, got %q", expected, node.String())
		}
	})
}

// TestAddressOfExpression tests the AddressOfExpression AST node.
func TestAddressOfExpression(t *testing.T) {
	t.Run("address-of simple identifier", func(t *testing.T) {
		node := &AddressOfExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.AT, Literal: "@"}},
			},
			Operator: NewTestIdentifier("MyFunction"),
		}

		expected := "@MyFunction"
		if node.String() != expected {
			t.Errorf("expected %q, got %q", expected, node.String())
		}

		if node.TokenLiteral() != "@" {
			t.Errorf("expected token literal '@', got %q", node.TokenLiteral())
		}

		// Verify it implements Expression interface
		var _ Expression = node
	})

	t.Run("address-of with type information", func(t *testing.T) {
		node := &AddressOfExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.AT, Literal: "@"}},
			},
			Operator: NewTestIdentifier("Ascending"),
		}

		expected := "@Ascending"
		if node.String() != expected {
			t.Errorf("expected %q, got %q", expected, node.String())
		}

		// Task 9.18: Type information is now stored in SemanticInfo, not on nodes
		// Type checking is tested via SemanticInfo tests
	})

	t.Run("address-of implements TypedExpression", func(t *testing.T) {
		node := &AddressOfExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.AT, Literal: "@"}},
			},
			Operator: NewTestIdentifier("Test"),
		}

		// Verify it implements TypedExpression interface
		var _ TypedExpression = node

		// Task 9.18: GetType/SetType methods removed - type info now in SemanticInfo
		// See pkg/ast/metadata_test.go for type annotation tests
	})

	t.Run("address-of position", func(t *testing.T) {
		pos := lexer.Position{Line: 5, Column: 10, Offset: 50}
		node := &AddressOfExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.AT, Literal: "@", Pos: pos}},
			},
			Operator: NewTestIdentifier("Test"),
		}

		if node.Pos() != pos {
			t.Errorf("expected position %v, got %v", pos, node.Pos())
		}
	})
}
