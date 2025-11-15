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
					BaseNode: BaseNode{
						Token: tok,
					},
					Name:  &Identifier{Token: tok, Value: "X"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPublic,
				},
				{
					BaseNode: BaseNode{
						Token: tok,
					},
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
					BaseNode: BaseNode{
						Token: tok,
					},
					Name:  &Identifier{Token: tok, Value: "X"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPublic,
				},
				{
					BaseNode: BaseNode{
						Token: tok,
					},
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
					BaseNode: BaseNode{
						Token: tok,
					},
					Name:  &Identifier{Token: tok, Value: "FX"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPrivate,
				},
				{
					BaseNode: BaseNode{
						Token: tok,
					},
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
					BaseNode: BaseNode{
						Token: tok,
					},
					Name:  &Identifier{Token: tok, Value: "X"},
					Type: &TypeAnnotation{
						Token: tok,
						Name:  "Integer",
					},
					Visibility: VisibilityPublic,
				},
				{
					BaseNode: BaseNode{
						Token: tok,
					},
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
