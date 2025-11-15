package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// SetDecl Tests
// ============================================================================

func TestSetDecl(t *testing.T) {
	t.Run("Named set declaration", func(t *testing.T) {
		// type TDays = set of TWeekday;
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}

		setDecl := &SetDecl{
			BaseNode:    BaseNode{Token: tok},
			Name:        NewTestIdentifier("TDays"),
			ElementType: NewTestTypeAnnotation("TWeekday"),
		}

		// Test TokenLiteral()
		if setDecl.TokenLiteral() != "type" {
			t.Errorf("TokenLiteral() = %v, want 'type'", setDecl.TokenLiteral())
		}

		// Test Name
		if setDecl.Name.Value != "TDays" {
			t.Errorf("Name.Value = %v, want 'TDays'", setDecl.Name.Value)
		}

		// Test ElementType
		if setDecl.ElementType.Name != "TWeekday" {
			t.Errorf("ElementType.Name = %v, want 'TWeekday'", setDecl.ElementType.Name)
		}
	})

	t.Run("String() method", func(t *testing.T) {
		// type TDays = set of TWeekday;
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}

		setDecl := &SetDecl{
			BaseNode:    BaseNode{Token: tok},
			Name:        NewTestIdentifier("TDays"),
			ElementType: NewTestTypeAnnotation("TWeekday"),
		}

		str := setDecl.String()
		// Should contain meaningful representation
		if str == "" {
			t.Error("String() should not be empty")
		}
		// The string should contain "set of" and the type names
		// Exact format can vary but should be meaningful
	})

	t.Run("Implements Statement interface", func(_ *testing.T) {
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}
		setDecl := &SetDecl{
			BaseNode:    BaseNode{Token: tok},
			Name:        NewTestIdentifier("TDays"),
			ElementType: NewTestTypeAnnotation("TWeekday"),
		}

		// Ensure it implements Statement interface
		var _ Statement = setDecl
	})
}

// ============================================================================
// SetLiteral Tests
// ============================================================================

func TestSetLiteral(t *testing.T) {
	t.Run("Set with elements", func(t *testing.T) {
		// [one, two, three]
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		setLit := &SetLiteral{
			Token: tok,
			Elements: []Expression{
				NewTestIdentifier("one"),
				NewTestIdentifier("two"),
				NewTestIdentifier("three"),
			},
		}

		// Test TokenLiteral()
		if setLit.TokenLiteral() != "[" {
			t.Errorf("TokenLiteral() = %v, want '['", setLit.TokenLiteral())
		}

		// Test Elements count
		if len(setLit.Elements) != 3 {
			t.Errorf("len(Elements) = %v, want 3", len(setLit.Elements))
		}
	})

	t.Run("Empty set", func(t *testing.T) {
		// []
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		setLit := &SetLiteral{
			Token:    tok,
			Elements: []Expression{},
		}

		// Test empty Elements
		if len(setLit.Elements) != 0 {
			t.Errorf("len(Elements) = %v, want 0", len(setLit.Elements))
		}
	})

	t.Run("String() method", func(t *testing.T) {
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		setLit := &SetLiteral{
			Token: tok,
			Elements: []Expression{
				NewTestIdentifier("one"),
				NewTestIdentifier("two"),
			},
		}

		str := setLit.String()
		// Should contain meaningful representation
		if str == "" {
			t.Error("String() should not be empty")
		}
	})

	t.Run("Implements Expression interface", func(_ *testing.T) {
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}
		setLit := &SetLiteral{
			Token:    tok,
			Elements: []Expression{},
		}

		// Ensure it implements Expression interface
		var _ Expression = setLit
	})
}

// ============================================================================
// Set Operators Tests
// ============================================================================

func TestSetOperators(t *testing.T) {
	t.Run("Set union operation", func(t *testing.T) {
		// s1 + s2
		tok := lexer.Token{Type: lexer.PLUS, Literal: "+"}

		s1 := NewTestIdentifier("s1")
		s2 := NewTestIdentifier("s2")

		unionExpr := &BinaryExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{
					Token: tok,
				},
			},
			Left:     s1,
			Operator: "+",
			Right:    s2,
		}

		// Test that BinaryExpression handles set union
		if unionExpr.Operator != "+" {
			t.Errorf("Operator = %v, want '+'", unionExpr.Operator)
		}
	})

	t.Run("Set difference operation", func(t *testing.T) {
		// s1 - s2
		tok := lexer.Token{Type: lexer.MINUS, Literal: "-"}

		s1 := NewTestIdentifier("s1")
		s2 := NewTestIdentifier("s2")

		diffExpr := &BinaryExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{
					Token: tok,
				},
			},
			Left:     s1,
			Operator: "-",
			Right:    s2,
		}

		// Test that BinaryExpression handles set difference
		if diffExpr.Operator != "-" {
			t.Errorf("Operator = %v, want '-'", diffExpr.Operator)
		}
	})

	t.Run("Set intersection operation", func(t *testing.T) {
		// s1 * s2
		tok := lexer.Token{Type: lexer.ASTERISK, Literal: "*"}

		s1 := NewTestIdentifier("s1")
		s2 := NewTestIdentifier("s2")

		intersectExpr := &BinaryExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{
					Token: tok,
				},
			},
			Left:     s1,
			Operator: "*",
			Right:    s2,
		}

		// Test that BinaryExpression handles set intersection
		if intersectExpr.Operator != "*" {
			t.Errorf("Operator = %v, want '*'", intersectExpr.Operator)
		}
	})

	t.Run("Set membership test", func(t *testing.T) {
		// one in mySet
		tok := lexer.Token{Type: lexer.IN, Literal: "in"}

		elem := NewTestIdentifier("one")
		set := NewTestIdentifier("mySet")

		inExpr := &BinaryExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{
					Token: tok,
				},
			},
			Left:     elem,
			Operator: "in",
			Right:    set,
		}

		// Test that BinaryExpression handles 'in' operator
		if inExpr.Operator != "in" {
			t.Errorf("Operator = %v, want 'in'", inExpr.Operator)
		}
	})

	t.Run("Set equality comparison", func(t *testing.T) {
		// s1 = s2
		tok := lexer.Token{Type: lexer.EQ, Literal: "="}

		s1 := NewTestIdentifier("s1")
		s2 := NewTestIdentifier("s2")

		eqExpr := &BinaryExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{
					Token: tok,
				},
			},
			Left:     s1,
			Operator: "=",
			Right:    s2,
		}

		// Test that BinaryExpression handles set equality
		if eqExpr.Operator != "=" {
			t.Errorf("Operator = %v, want '='", eqExpr.Operator)
		}
	})

	t.Run("Set inequality comparison", func(t *testing.T) {
		// s1 <> s2
		tok := lexer.Token{Type: lexer.NOT_EQ, Literal: "<>"}

		s1 := NewTestIdentifier("s1")
		s2 := NewTestIdentifier("s2")

		neqExpr := &BinaryExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{
					Token: tok,
				},
			},
			Left:     s1,
			Operator: "<>",
			Right:    s2,
		}

		// Test that BinaryExpression handles set inequality
		if neqExpr.Operator != "<>" {
			t.Errorf("Operator = %v, want '<>'", neqExpr.Operator)
		}
	})

	t.Run("Set subset comparison", func(t *testing.T) {
		// s1 <= s2 (s1 is subset of s2)
		tok := lexer.Token{Type: lexer.LESS_EQ, Literal: "<="}

		s1 := NewTestIdentifier("s1")
		s2 := NewTestIdentifier("s2")

		subsetExpr := &BinaryExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{
					Token: tok,
				},
			},
			Left:     s1,
			Operator: "<=",
			Right:    s2,
		}

		// Test that BinaryExpression handles subset comparison
		if subsetExpr.Operator != "<=" {
			t.Errorf("Operator = %v, want '<='", subsetExpr.Operator)
		}
	})

	t.Run("Set superset comparison", func(t *testing.T) {
		// s1 >= s2 (s1 is superset of s2)
		tok := lexer.Token{Type: lexer.GREATER_EQ, Literal: ">="}

		s1 := NewTestIdentifier("s1")
		s2 := NewTestIdentifier("s2")

		supersetExpr := &BinaryExpression{
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{
					Token: tok,
				},
			},
			Left:     s1,
			Operator: ">=",
			Right:    s2,
		}

		// Test that BinaryExpression handles superset comparison
		if supersetExpr.Operator != ">=" {
			t.Errorf("Operator = %v, want '>='", supersetExpr.Operator)
		}
	})
}
