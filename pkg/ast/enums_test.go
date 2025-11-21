package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// EnumDecl Tests
// ============================================================================

func TestEnumDecl(t *testing.T) {
	t.Run("Basic enum declaration", func(t *testing.T) {
		// type TColor = (Red, Green, Blue);
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}

		enumDecl := &EnumDecl{
			BaseNode: BaseNode{Token: tok},
			Name:     NewTestIdentifier("TColor"),
			Values: []EnumValue{
				{Name: "Red", Value: nil},
				{Name: "Green", Value: nil},
				{Name: "Blue", Value: nil},
			},
		}

		// Test TokenLiteral()
		if enumDecl.TokenLiteral() != "type" {
			t.Errorf("TokenLiteral() = %v, want 'type'", enumDecl.TokenLiteral())
		}

		// Test Name
		if enumDecl.Name.Value != "TColor" {
			t.Errorf("Name.Value = %v, want 'TColor'", enumDecl.Name.Value)
		}

		// Test Values count
		if len(enumDecl.Values) != 3 {
			t.Errorf("len(Values) = %v, want 3", len(enumDecl.Values))
		}

		// Test individual values
		if enumDecl.Values[0].Name != "Red" {
			t.Errorf("Values[0].Name = %v, want 'Red'", enumDecl.Values[0].Name)
		}
		if enumDecl.Values[1].Name != "Green" {
			t.Errorf("Values[1].Name = %v, want 'Green'", enumDecl.Values[1].Name)
		}
		if enumDecl.Values[2].Name != "Blue" {
			t.Errorf("Values[2].Name = %v, want 'Blue'", enumDecl.Values[2].Name)
		}

		// Test that values without explicit values have nil
		if enumDecl.Values[0].Value != nil {
			t.Error("Values[0].Value should be nil (implicit value)")
		}
	})

	t.Run("Enum with explicit values", func(t *testing.T) {
		// type TEnum = (One = 1, Two = 5, Three = 10);
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}

		one := 1
		five := 5
		ten := 10

		enumDecl := &EnumDecl{
			BaseNode: BaseNode{Token: tok},
			Name:     NewTestIdentifier("TEnum"),
			Values: []EnumValue{
				{Name: "One", Value: &one},
				{Name: "Two", Value: &five},
				{Name: "Three", Value: &ten},
			},
		}

		// Test explicit values
		if enumDecl.Values[0].Value == nil || *enumDecl.Values[0].Value != 1 {
			t.Error("Values[0].Value should be 1")
		}
		if enumDecl.Values[1].Value == nil || *enumDecl.Values[1].Value != 5 {
			t.Error("Values[1].Value should be 5")
		}
		if enumDecl.Values[2].Value == nil || *enumDecl.Values[2].Value != 10 {
			t.Error("Values[2].Value should be 10")
		}
	})

	t.Run("String() method", func(t *testing.T) {
		// type TColor = (Red, Green, Blue);
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}

		enumDecl := &EnumDecl{
			BaseNode: BaseNode{Token: tok},
			Name:     NewTestIdentifier("TColor"),
			Values: []EnumValue{
				{Name: "Red", Value: nil},
				{Name: "Green", Value: nil},
				{Name: "Blue", Value: nil},
			},
		}

		str := enumDecl.String()
		// Should contain the type name and value names
		if str == "" {
			t.Error("String() should not be empty")
		}
		// At minimum, should contain the type name
		// The exact format can vary, but it should be meaningful
	})

	t.Run("Implements Statement interface", func(_ *testing.T) {
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}
		enumDecl := &EnumDecl{
			BaseNode: BaseNode{Token: tok},
			Name:     NewTestIdentifier("TColor"),
			Values: []EnumValue{
				{Name: "Red", Value: nil},
			},
		}

		// Ensure it implements Statement interface
		var _ Statement = enumDecl
	})
}

// ============================================================================
// EnumLiteral Tests
// ============================================================================

func TestEnumLiteral(t *testing.T) {
	t.Run("Direct enum value reference", func(t *testing.T) {
		// Red (without scope)
		tok := lexer.Token{Type: lexer.IDENT, Literal: "Red"}

		enumLit := &EnumLiteral{
			Token:     tok,
			EnumName:  "",
			ValueName: "Red",
		}

		// Test TokenLiteral()
		if enumLit.TokenLiteral() != "Red" {
			t.Errorf("TokenLiteral() = %v, want 'Red'", enumLit.TokenLiteral())
		}

		// Test ValueName
		if enumLit.ValueName != "Red" {
			t.Errorf("ValueName = %v, want 'Red'", enumLit.ValueName)
		}

		// Test EnumName is empty for unscoped reference
		if enumLit.EnumName != "" {
			t.Errorf("EnumName = %v, want empty string", enumLit.EnumName)
		}
	})

	t.Run("Scoped enum value reference", func(t *testing.T) {
		// TColor.Red
		tok := lexer.Token{Type: lexer.IDENT, Literal: "TColor"}

		enumLit := &EnumLiteral{
			Token:     tok,
			EnumName:  "TColor",
			ValueName: "Red",
		}

		// Test both parts
		if enumLit.EnumName != "TColor" {
			t.Errorf("EnumName = %v, want 'TColor'", enumLit.EnumName)
		}
		if enumLit.ValueName != "Red" {
			t.Errorf("ValueName = %v, want 'Red'", enumLit.ValueName)
		}
	})

	t.Run("String() method", func(t *testing.T) {
		tok := lexer.Token{Type: lexer.IDENT, Literal: "Red"}

		// Test unscoped
		unscopedLit := &EnumLiteral{
			Token:     tok,
			EnumName:  "",
			ValueName: "Red",
		}
		if unscopedLit.String() == "" {
			t.Error("String() should not be empty for unscoped literal")
		}

		// Test scoped
		scopedLit := &EnumLiteral{
			Token:     tok,
			EnumName:  "TColor",
			ValueName: "Red",
		}
		scopedStr := scopedLit.String()
		if scopedStr == "" {
			t.Error("String() should not be empty for scoped literal")
		}
	})

	t.Run("Implements Expression interface", func(_ *testing.T) {
		tok := lexer.Token{Type: lexer.IDENT, Literal: "Red"}
		enumLit := &EnumLiteral{
			Token:     tok,
			EnumName:  "",
			ValueName: "Red",
		}

		// Ensure it implements Expression interface
		var _ Expression = enumLit
	})
}
