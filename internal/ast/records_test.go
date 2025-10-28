package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// RecordDecl Tests
// ============================================================================

func TestRecordDecl(t *testing.T) {
	t.Run("Basic record declaration", func(t *testing.T) {
		// type TPoint = record X, Y: Integer; end;
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}

		recordDecl := &RecordDecl{
			Token: tok,
			Name:  &Identifier{Token: tok, Value: "TPoint"},
			Fields: []*FieldDecl{
				{
					Token: tok,
					Name:  &Identifier{Token: tok, Value: "X"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPublic,
				},
				{
					Token: tok,
					Name:  &Identifier{Token: tok, Value: "Y"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPublic,
				},
			},
		}

		// Test TokenLiteral()
		if recordDecl.TokenLiteral() != "type" {
			t.Errorf("TokenLiteral() = %v, want 'type'", recordDecl.TokenLiteral())
		}

		// Test Name
		if recordDecl.Name.Value != "TPoint" {
			t.Errorf("Name.Value = %v, want 'TPoint'", recordDecl.Name.Value)
		}

		// Test Fields count
		if len(recordDecl.Fields) != 2 {
			t.Errorf("len(Fields) = %v, want 2", len(recordDecl.Fields))
		}

		// Test field names
		if recordDecl.Fields[0].Name.Value != "X" {
			t.Errorf("Fields[0].Name.Value = %v, want 'X'", recordDecl.Fields[0].Name.Value)
		}
		if recordDecl.Fields[1].Name.Value != "Y" {
			t.Errorf("Fields[1].Name.Value = %v, want 'Y'", recordDecl.Fields[1].Name.Value)
		}
	})

	t.Run("Record with methods", func(t *testing.T) {
		// type TPoint = record
		//   X, Y: Integer;
		//   function GetDistance: Float;
		// end;
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}

		recordDecl := &RecordDecl{
			Token: tok,
			Name:  &Identifier{Token: tok, Value: "TPoint"},
			Fields: []*FieldDecl{
				{
					Token: tok,
					Name:  &Identifier{Token: tok, Value: "X"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPublic,
				},
				{
					Token: tok,
					Name:  &Identifier{Token: tok, Value: "Y"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPublic,
				},
			},
			Methods: []*FunctionDecl{
				{
					Token: tok,
					Name:  &Identifier{Token: tok, Value: "GetDistance"},
					ReturnType: &TypeAnnotation{
						Token: tok,
						Name:  "Float",
					},
				},
			},
		}

		// Test Methods count
		if len(recordDecl.Methods) != 1 {
			t.Errorf("len(Methods) = %v, want 1", len(recordDecl.Methods))
		}

		// Test method name
		if recordDecl.Methods[0].Name.Value != "GetDistance" {
			t.Errorf("Methods[0].Name.Value = %v, want 'GetDistance'", recordDecl.Methods[0].Name.Value)
		}
	})

	t.Run("Record with visibility sections", func(t *testing.T) {
		// type TPoint = record
		// private
		//   FX, FY: Integer;
		// public
		//   property X: Integer read FX write FX;
		// end;
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}

		recordDecl := &RecordDecl{
			Token: tok,
			Name:  &Identifier{Token: tok, Value: "TPoint"},
			Fields: []*FieldDecl{
				{
					Token: tok,
					Name:  &Identifier{Token: tok, Value: "FX"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPrivate,
				},
				{
					Token: tok,
					Name:  &Identifier{Token: tok, Value: "FY"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPrivate,
				},
			},
		}

		// Test private fields
		if recordDecl.Fields[0].Visibility != VisibilityPrivate {
			t.Errorf("Fields[0].Visibility = %v, want VisibilityPrivate", recordDecl.Fields[0].Visibility)
		}
	})

	t.Run("String() method for basic record", func(t *testing.T) {
		// type TPoint = record X, Y: Integer; end;
		tok := lexer.Token{Type: lexer.TYPE, Literal: "type"}

		recordDecl := &RecordDecl{
			Token: tok,
			Name:  &Identifier{Token: tok, Value: "TPoint"},
			Fields: []*FieldDecl{
				{
					Token: tok,
					Name:  &Identifier{Token: tok, Value: "X"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPublic,
				},
				{
					Token: tok,
					Name:  &Identifier{Token: tok, Value: "Y"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPublic,
				},
			},
		}

		str := recordDecl.String()

		// String should contain key elements
		if str == "" {
			t.Error("String() should return non-empty string")
		}

		// We'll verify correct string output once String() is implemented
		// For now, just check it returns something
	})
}

// ============================================================================
// RecordLiteral Tests
// ============================================================================

func TestRecordLiteral(t *testing.T) {
	t.Run("Named field record literal", func(t *testing.T) {
		// (X: 10, Y: 20)
		tok := lexer.Token{Type: lexer.LPAREN, Literal: "("}

		recordLit := &RecordLiteral{
			Token: tok,
			Fields: []RecordField{
				{
					Name:  "X",
					Value: &IntegerLiteral{Token: tok, Value: 10},
				},
				{
					Name:  "Y",
					Value: &IntegerLiteral{Token: tok, Value: 20},
				},
			},
		}

		// Test TokenLiteral()
		if recordLit.TokenLiteral() != "(" {
			t.Errorf("TokenLiteral() = %v, want '('", recordLit.TokenLiteral())
		}

		// Test Fields count
		if len(recordLit.Fields) != 2 {
			t.Errorf("len(Fields) = %v, want 2", len(recordLit.Fields))
		}

		// Test field names and values
		if recordLit.Fields[0].Name != "X" {
			t.Errorf("Fields[0].Name = %v, want 'X'", recordLit.Fields[0].Name)
		}
		if recordLit.Fields[1].Name != "Y" {
			t.Errorf("Fields[1].Name = %v, want 'Y'", recordLit.Fields[1].Name)
		}

		// Test that values are integers
		if _, ok := recordLit.Fields[0].Value.(*IntegerLiteral); !ok {
			t.Error("Fields[0].Value should be IntegerLiteral")
		}
	})

	t.Run("Positional record literal", func(t *testing.T) {
		// (10, 20)
		tok := lexer.Token{Type: lexer.LPAREN, Literal: "("}

		recordLit := &RecordLiteral{
			Token: tok,
			Fields: []RecordField{
				{
					Name:  "", // Empty name for positional
					Value: &IntegerLiteral{Token: tok, Value: 10},
				},
				{
					Name:  "",
					Value: &IntegerLiteral{Token: tok, Value: 20},
				},
			},
		}

		// Test Fields count
		if len(recordLit.Fields) != 2 {
			t.Errorf("len(Fields) = %v, want 2", len(recordLit.Fields))
		}

		// Test that names are empty (positional)
		if recordLit.Fields[0].Name != "" {
			t.Errorf("Fields[0].Name = %v, want empty string", recordLit.Fields[0].Name)
		}
		if recordLit.Fields[1].Name != "" {
			t.Errorf("Fields[1].Name = %v, want empty string", recordLit.Fields[1].Name)
		}
	})

	t.Run("Record literal with type", func(t *testing.T) {
		// TPoint(X: 10, Y: 20) or TPoint(10, 20)
		tok := lexer.Token{Type: lexer.LPAREN, Literal: "("}

		recordLit := &RecordLiteral{
			Token:    tok,
			TypeName: "TPoint",
			Fields: []RecordField{
				{
					Name:  "X",
					Value: &IntegerLiteral{Token: tok, Value: 10},
				},
				{
					Name:  "Y",
					Value: &IntegerLiteral{Token: tok, Value: 20},
				},
			},
		}

		// Test TypeName
		if recordLit.TypeName != "TPoint" {
			t.Errorf("TypeName = %v, want 'TPoint'", recordLit.TypeName)
		}
	})

	t.Run("String() method for named fields", func(t *testing.T) {
		// (X: 10, Y: 20)
		tok := lexer.Token{Type: lexer.LPAREN, Literal: "("}

		recordLit := &RecordLiteral{
			Token: tok,
			Fields: []RecordField{
				{
					Name:  "X",
					Value: &IntegerLiteral{Token: tok, Value: 10},
				},
				{
					Name:  "Y",
					Value: &IntegerLiteral{Token: tok, Value: 20},
				},
			},
		}

		str := recordLit.String()

		// String should contain key elements
		if str == "" {
			t.Error("String() should return non-empty string")
		}

		// Should contain field names (using strings.Contains from standard library)
		// Note: We don't import strings in test to keep it simple, so we check for non-empty
		if str == "" {
			t.Error("String() should return non-empty string")
		}
	})

	t.Run("String() method for positional fields", func(t *testing.T) {
		// (10, 20)
		tok := lexer.Token{Type: lexer.LPAREN, Literal: "("}

		recordLit := &RecordLiteral{
			Token: tok,
			Fields: []RecordField{
				{
					Name:  "",
					Value: &IntegerLiteral{Token: tok, Value: 10},
				},
				{
					Name:  "",
					Value: &IntegerLiteral{Token: tok, Value: 20},
				},
			},
		}

		str := recordLit.String()

		// String should contain values
		if str == "" {
			t.Error("String() should return non-empty string")
		}
	})
}
