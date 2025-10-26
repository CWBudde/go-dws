package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/lexer"
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
