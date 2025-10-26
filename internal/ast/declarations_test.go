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
