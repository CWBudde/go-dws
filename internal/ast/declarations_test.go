package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestConstDecl tests the ConstDecl AST node with an integer constant
func TestConstDecl(t *testing.T) {
	// const MAX = 100;
	constDecl := &ConstDecl{
		Token: lexer.Token{Type: lexer.CONST, Literal: "const", Pos: lexer.Position{Line: 1, Column: 1}},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "MAX", Pos: lexer.Position{Line: 1, Column: 7}},
			Value: "MAX",
		},
		Type: nil, // No type annotation
		Value: &IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT, Literal: "100", Pos: lexer.Position{Line: 1, Column: 13}},
			Value: 100,
		},
	}

	if constDecl.TokenLiteral() != "const" {
		t.Errorf("TokenLiteral() wrong. expected=%q, got=%q", "const", constDecl.TokenLiteral())
	}

	expectedString := "const MAX = 100"
	if constDecl.String() != expectedString {
		t.Errorf("String() wrong. expected=%q, got=%q", expectedString, constDecl.String())
	}

	if constDecl.Pos().Line != 1 || constDecl.Pos().Column != 1 {
		t.Errorf("Pos() wrong. expected Line=1, Column=1, got Line=%d, Column=%d",
			constDecl.Pos().Line, constDecl.Pos().Column)
	}
}

// TestConstDeclWithFloat tests the ConstDecl AST node with a float constant
func TestConstDeclWithFloat(t *testing.T) {
	// const PI = 3.14;
	constDecl := &ConstDecl{
		Token: lexer.Token{Type: lexer.CONST, Literal: "const", Pos: lexer.Position{Line: 1, Column: 1}},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "PI", Pos: lexer.Position{Line: 1, Column: 7}},
			Value: "PI",
		},
		Type: nil,
		Value: &FloatLiteral{
			Token: lexer.Token{Type: lexer.FLOAT, Literal: "3.14", Pos: lexer.Position{Line: 1, Column: 12}},
			Value: 3.14,
		},
	}

	expectedString := "const PI = 3.14"
	if constDecl.String() != expectedString {
		t.Errorf("String() wrong. expected=%q, got=%q", expectedString, constDecl.String())
	}
}

// TestConstDeclWithString tests the ConstDecl AST node with a string constant
func TestConstDeclWithString(t *testing.T) {
	// const APP_NAME = 'MyApp';
	constDecl := &ConstDecl{
		Token: lexer.Token{Type: lexer.CONST, Literal: "const", Pos: lexer.Position{Line: 1, Column: 1}},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "APP_NAME", Pos: lexer.Position{Line: 1, Column: 7}},
			Value: "APP_NAME",
		},
		Type: nil,
		Value: &StringLiteral{
			Token: lexer.Token{Type: lexer.STRING, Literal: "MyApp", Pos: lexer.Position{Line: 1, Column: 18}},
			Value: "MyApp",
		},
	}

	expectedString := "const APP_NAME = \"MyApp\""
	if constDecl.String() != expectedString {
		t.Errorf("String() wrong. expected=%q, got=%q", expectedString, constDecl.String())
	}
}

// TestConstDeclTyped tests the ConstDecl AST node with explicit type annotation
func TestConstDeclTyped(t *testing.T) {
	// const MAX_USERS: Integer = 1000;
	constDecl := &ConstDecl{
		Token: lexer.Token{Type: lexer.CONST, Literal: "const", Pos: lexer.Position{Line: 1, Column: 1}},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "MAX_USERS", Pos: lexer.Position{Line: 1, Column: 7}},
			Value: "MAX_USERS",
		},
		Type: &TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer", Pos: lexer.Position{Line: 1, Column: 18}},
			Name:  "Integer",
		},
		Value: &IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT, Literal: "1000", Pos: lexer.Position{Line: 1, Column: 28}},
			Value: 1000,
		},
	}

	expectedString := "const MAX_USERS: Integer = 1000"
	if constDecl.String() != expectedString {
		t.Errorf("String() wrong. expected=%q, got=%q", expectedString, constDecl.String())
	}
}

// ============================================================================
// TypeDeclaration Tests
// ============================================================================

// TestTypeDeclaration tests the TypeDeclaration AST node for type aliases
func TestTypeDeclaration(t *testing.T) {
	t.Run("Basic type alias to Integer", func(t *testing.T) {
		// type TUserID = Integer;
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 1, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID", Pos: lexer.Position{Line: 1, Column: 6}},
				Value: "TUserID",
			},
			IsAlias: true,
			AliasedType: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer", Pos: lexer.Position{Line: 1, Column: 16}},
				Name:  "Integer",
			},
		}

		// Test TokenLiteral()
		if typeDecl.TokenLiteral() != "type" {
			t.Errorf("TokenLiteral() wrong. expected=%q, got=%q", "type", typeDecl.TokenLiteral())
		}

		// Test String()
		expectedString := "type TUserID = Integer"
		if typeDecl.String() != expectedString {
			t.Errorf("String() wrong. expected=%q, got=%q", expectedString, typeDecl.String())
		}

		// Test Pos()
		if typeDecl.Pos().Line != 1 || typeDecl.Pos().Column != 1 {
			t.Errorf("Pos() wrong. expected Line=1, Column=1, got Line=%d, Column=%d",
				typeDecl.Pos().Line, typeDecl.Pos().Column)
		}

		// Test IsAlias field
		if !typeDecl.IsAlias {
			t.Error("IsAlias should be true for type alias")
		}

		// Test AliasedType field
		if typeDecl.AliasedType == nil {
			t.Error("AliasedType should not be nil for type alias")
		}

		if typeDecl.AliasedType.Name != "Integer" {
			t.Errorf("AliasedType.Name wrong. expected=%q, got=%q", "Integer", typeDecl.AliasedType.Name)
		}
	})

	t.Run("Type alias to String", func(t *testing.T) {
		// type TFileName = String;
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 2, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TFileName", Pos: lexer.Position{Line: 2, Column: 6}},
				Value: "TFileName",
			},
			IsAlias: true,
			AliasedType: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "String", Pos: lexer.Position{Line: 2, Column: 18}},
				Name:  "String",
			},
		}

		expectedString := "type TFileName = String"
		if typeDecl.String() != expectedString {
			t.Errorf("String() wrong. expected=%q, got=%q", expectedString, typeDecl.String())
		}
	})

	t.Run("Type alias to Float", func(t *testing.T) {
		// type TPrice = Float;
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 3, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TPrice", Pos: lexer.Position{Line: 3, Column: 6}},
				Value: "TPrice",
			},
			IsAlias: true,
			AliasedType: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Float", Pos: lexer.Position{Line: 3, Column: 15}},
				Name:  "Float",
			},
		}

		expectedString := "type TPrice = Float"
		if typeDecl.String() != expectedString {
			t.Errorf("String() wrong. expected=%q, got=%q", expectedString, typeDecl.String())
		}
	})

	t.Run("Type alias to Boolean", func(t *testing.T) {
		// type TFlag = Boolean;
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 4, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TFlag", Pos: lexer.Position{Line: 4, Column: 6}},
				Value: "TFlag",
			},
			IsAlias: true,
			AliasedType: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Boolean", Pos: lexer.Position{Line: 4, Column: 14}},
				Name:  "Boolean",
			},
		}

		expectedString := "type TFlag = Boolean"
		if typeDecl.String() != expectedString {
			t.Errorf("String() wrong. expected=%q, got=%q", expectedString, typeDecl.String())
		}
	})

	t.Run("Type alias to complex type", func(t *testing.T) {
		// type TIntArray = array of Integer;
		// Note: This tests that TypeAnnotation can hold complex type names
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 5, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TIntArray", Pos: lexer.Position{Line: 5, Column: 6}},
				Value: "TIntArray",
			},
			IsAlias: true,
			AliasedType: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.ARRAY, Literal: "array", Pos: lexer.Position{Line: 5, Column: 18}},
				Name:  "array of Integer",
			},
		}

		expectedString := "type TIntArray = array of Integer"
		if typeDecl.String() != expectedString {
			t.Errorf("String() wrong. expected=%q, got=%q", expectedString, typeDecl.String())
		}
	})

	t.Run("Type alias to another alias type", func(t *testing.T) {
		// type TMyInt = TUserID;
		// (where TUserID is itself an alias to Integer)
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 6, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TMyInt", Pos: lexer.Position{Line: 6, Column: 6}},
				Value: "TMyInt",
			},
			IsAlias: true,
			AliasedType: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID", Pos: lexer.Position{Line: 6, Column: 15}},
				Name:  "TUserID",
			},
		}

		expectedString := "type TMyInt = TUserID"
		if typeDecl.String() != expectedString {
			t.Errorf("String() wrong. expected=%q, got=%q", expectedString, typeDecl.String())
		}
	})

	t.Run("Non-alias type declaration (future)", func(t *testing.T) {
		// This tests the future case where TypeDeclaration might be used
		// for full type definitions (not just aliases)
		// For now, IsAlias=false just returns "type Name"
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 7, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TMyRecord", Pos: lexer.Position{Line: 7, Column: 6}},
				Value: "TMyRecord",
			},
			IsAlias:     false,
			AliasedType: nil,
		}

		expectedString := "type TMyRecord"
		if typeDecl.String() != expectedString {
			t.Errorf("String() wrong. expected=%q, got=%q", expectedString, typeDecl.String())
		}
	})
}

// TestTypeDeclarationImplementsStatement verifies that TypeDeclaration implements the Statement interface
func TestTypeDeclarationImplementsStatement(t *testing.T) {
	typeDecl := &TypeDeclaration{
		Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 1, Column: 1}},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID", Pos: lexer.Position{Line: 1, Column: 6}},
			Value: "TUserID",
		},
		IsAlias: true,
		AliasedType: &TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer", Pos: lexer.Position{Line: 1, Column: 16}},
			Name:  "Integer",
		},
	}

	// This will fail to compile if TypeDeclaration doesn't implement Statement
	var _ Statement = typeDecl

	// Verify statementNode() is callable (even though it does nothing)
	typeDecl.statementNode()
}

// ============================================================================
// Subrange Type Declaration Tests (Tasks 9.94-9.95)
// ============================================================================

// TestSubrangeTypeDeclaration tests the TypeDeclaration AST node for subrange types
func TestSubrangeTypeDeclaration(t *testing.T) {
	t.Run("Basic digit subrange (0..9)", func(t *testing.T) {
		// type TDigit = 0..9;
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 1, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TDigit", Pos: lexer.Position{Line: 1, Column: 6}},
				Value: "TDigit",
			},
			IsSubrange: true,
			LowBound: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: lexer.Position{Line: 1, Column: 16}},
				Value: 0,
			},
			HighBound: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "9", Pos: lexer.Position{Line: 1, Column: 19}},
				Value: 9,
			},
		}

		// Test that IsSubrange flag is set
		if !typeDecl.IsSubrange {
			t.Error("IsSubrange should be true")
		}

		// Test that bounds are set correctly
		lowBound, ok := typeDecl.LowBound.(*IntegerLiteral)
		if !ok {
			t.Fatal("LowBound should be an IntegerLiteral")
		}
		if lowBound.Value != 0 {
			t.Errorf("LowBound value = %d, want 0", lowBound.Value)
		}

		highBound, ok := typeDecl.HighBound.(*IntegerLiteral)
		if !ok {
			t.Fatal("HighBound should be an IntegerLiteral")
		}
		if highBound.Value != 9 {
			t.Errorf("HighBound value = %d, want 9", highBound.Value)
		}

		// Test String() output
		expectedString := "type TDigit = 0..9"
		if typeDecl.String() != expectedString {
			t.Errorf("String() = %q, want %q", typeDecl.String(), expectedString)
		}
	})

	t.Run("Percentage subrange (0..100)", func(t *testing.T) {
		// type TPercent = 0..100;
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 2, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TPercent", Pos: lexer.Position{Line: 2, Column: 6}},
				Value: "TPercent",
			},
			IsSubrange: true,
			LowBound: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: lexer.Position{Line: 2, Column: 18}},
				Value: 0,
			},
			HighBound: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "100", Pos: lexer.Position{Line: 2, Column: 21}},
				Value: 100,
			},
		}

		expectedString := "type TPercent = 0..100"
		if typeDecl.String() != expectedString {
			t.Errorf("String() = %q, want %q", typeDecl.String(), expectedString)
		}
	})

	t.Run("Negative range subrange (-40..50)", func(t *testing.T) {
		// type TTemperature = -40..50;
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 3, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TTemperature", Pos: lexer.Position{Line: 3, Column: 6}},
				Value: "TTemperature",
			},
			IsSubrange: true,
			LowBound: &UnaryExpression{
				Token:    lexer.Token{Type: lexer.MINUS, Literal: "-", Pos: lexer.Position{Line: 3, Column: 22}},
				Operator: "-",
				Right: &IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "40", Pos: lexer.Position{Line: 3, Column: 23}},
					Value: 40,
				},
			},
			HighBound: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "50", Pos: lexer.Position{Line: 3, Column: 27}},
				Value: 50,
			},
		}

		expectedString := "type TTemperature = (-40)..50"
		if typeDecl.String() != expectedString {
			t.Errorf("String() = %q, want %q", typeDecl.String(), expectedString)
		}
	})

	t.Run("Single value range (42..42)", func(t *testing.T) {
		// type TAnswer = 42..42;
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 4, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TAnswer", Pos: lexer.Position{Line: 4, Column: 6}},
				Value: "TAnswer",
			},
			IsSubrange: true,
			LowBound: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: lexer.Position{Line: 4, Column: 17}},
				Value: 42,
			},
			HighBound: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: lexer.Position{Line: 4, Column: 21}},
				Value: 42,
			},
		}

		expectedString := "type TAnswer = 42..42"
		if typeDecl.String() != expectedString {
			t.Errorf("String() = %q, want %q", typeDecl.String(), expectedString)
		}
	})
}

// TestSubrangeTypeDeclarationFields verifies that subrange-specific fields exist
func TestSubrangeTypeDeclarationFields(t *testing.T) {
	typeDecl := &TypeDeclaration{
		Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 1, Column: 1}},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "TDigit", Pos: lexer.Position{Line: 1, Column: 6}},
			Value: "TDigit",
		},
		IsSubrange: true,
		LowBound: &IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: lexer.Position{Line: 1, Column: 16}},
			Value: 0,
		},
		HighBound: &IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT, Literal: "9", Pos: lexer.Position{Line: 1, Column: 19}},
			Value: 9,
		},
	}

	// Verify Name field exists and is accessible
	if typeDecl.Name == nil || typeDecl.Name.Value != "TDigit" {
		t.Error("Name field should be accessible and equal to 'TDigit'")
	}

	// Verify IsSubrange field exists and is accessible
	if !typeDecl.IsSubrange {
		t.Error("IsSubrange field should be accessible and true")
	}

	// Verify LowBound field exists and is accessible
	if typeDecl.LowBound == nil {
		t.Error("LowBound field should be accessible and non-nil")
	}

	// Verify HighBound field exists and is accessible
	if typeDecl.HighBound == nil {
		t.Error("HighBound field should be accessible and non-nil")
	}

	// Verify bounds implement Expression interface
	// This is a compile-time check that will fail if types don't implement Expression
	if _, ok := interface{}(typeDecl.LowBound).(Expression); !ok {
		t.Error("LowBound should implement Expression interface")
	}
	if _, ok := interface{}(typeDecl.HighBound).(Expression); !ok {
		t.Error("HighBound should implement Expression interface")
	}
}

// TestSubrangeVsAliasTypeDeclaration ensures subrange and alias types are mutually exclusive
func TestSubrangeVsAliasTypeDeclaration(t *testing.T) {
	t.Run("Subrange type should not be alias", func(t *testing.T) {
		// type TDigit = 0..9; (subrange, not alias)
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 1, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TDigit", Pos: lexer.Position{Line: 1, Column: 6}},
				Value: "TDigit",
			},
			IsSubrange: true,
			LowBound: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: lexer.Position{Line: 1, Column: 16}},
				Value: 0,
			},
			HighBound: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "9", Pos: lexer.Position{Line: 1, Column: 19}},
				Value: 9,
			},
			IsAlias: false, // Should be false for subranges
		}

		if typeDecl.IsAlias {
			t.Error("IsAlias should be false for subrange types")
		}
		if !typeDecl.IsSubrange {
			t.Error("IsSubrange should be true for subrange types")
		}

		// String should show subrange format, not alias format
		expectedString := "type TDigit = 0..9"
		if typeDecl.String() != expectedString {
			t.Errorf("String() = %q, want %q", typeDecl.String(), expectedString)
		}
	})

	t.Run("Alias type should not be subrange", func(t *testing.T) {
		// type TUserID = Integer; (alias, not subrange)
		typeDecl := &TypeDeclaration{
			Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 2, Column: 1}},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID", Pos: lexer.Position{Line: 2, Column: 6}},
				Value: "TUserID",
			},
			IsAlias: true,
			AliasedType: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer", Pos: lexer.Position{Line: 2, Column: 16}},
				Name:  "Integer",
			},
			IsSubrange: false, // Should be false for aliases
		}

		if typeDecl.IsSubrange {
			t.Error("IsSubrange should be false for alias types")
		}
		if !typeDecl.IsAlias {
			t.Error("IsAlias should be true for alias types")
		}

		// String should show alias format, not subrange format
		expectedString := "type TUserID = Integer"
		if typeDecl.String() != expectedString {
			t.Errorf("String() = %q, want %q", typeDecl.String(), expectedString)
		}
	})
}
