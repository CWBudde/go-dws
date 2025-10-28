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
					Name:  &Identifier{Value: "a", Token: lexer.Token{Type: lexer.IDENT, Literal: "a"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "a"},
				},
				{
					Name:  &Identifier{Value: "b", Token: lexer.Token{Type: lexer.IDENT, Literal: "b"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "b"},
				},
			},
			ReturnType: &TypeAnnotation{Name: "Integer"},
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
					Name:  &Identifier{Value: "msg", Token: lexer.Token{Type: lexer.IDENT, Literal: "msg"}},
					Type:  &TypeAnnotation{Name: "String"},
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
					Name:  &Identifier{Value: "Sender", Token: lexer.Token{Type: lexer.IDENT, Literal: "Sender"}},
					Type:  &TypeAnnotation{Name: "TObject"},
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
					Name:  &Identifier{Value: "x", Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				},
			},
			ReturnType: &TypeAnnotation{Name: "Boolean"},
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
					Name:  &Identifier{Value: "x", Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
					Type:  &TypeAnnotation{Name: "Integer"},
					Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
					ByRef: true,
				},
			},
			ReturnType: &TypeAnnotation{Name: "Boolean"},
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
			Operator: &Identifier{
				Value: "MyFunction",
				Token: lexer.Token{Type: lexer.IDENT, Literal: "MyFunction"},
			},
			Token: lexer.Token{Type: lexer.AT, Literal: "@"},
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
		funcPtrType := &TypeAnnotation{Name: "TComparator"}
		node := &AddressOfExpression{
			Operator: &Identifier{
				Value: "Ascending",
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Ascending"},
			},
			Type:  funcPtrType,
			Token: lexer.Token{Type: lexer.AT, Literal: "@"},
		}

		expected := "@Ascending"
		if node.String() != expected {
			t.Errorf("expected %q, got %q", expected, node.String())
		}

		if node.GetType() != funcPtrType {
			t.Error("expected type to be set")
		}
	})

	t.Run("address-of implements TypedExpression", func(t *testing.T) {
		node := &AddressOfExpression{
			Operator: &Identifier{
				Value: "Test",
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Test"},
			},
			Token: lexer.Token{Type: lexer.AT, Literal: "@"},
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

	t.Run("address-of position", func(t *testing.T) {
		pos := lexer.Position{Line: 5, Column: 10, Offset: 50}
		node := &AddressOfExpression{
			Operator: &Identifier{
				Value: "Test",
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Test"},
			},
			Token: lexer.Token{Type: lexer.AT, Literal: "@", Pos: pos},
		}

		if node.Pos() != pos {
			t.Errorf("expected position %v, got %v", pos, node.Pos())
		}
	})
}
