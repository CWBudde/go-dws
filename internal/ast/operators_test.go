package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestOperatorDeclString_GlobalBinary(t *testing.T) {
	decl := &OperatorDecl{
		Token:          lexer.Token{Type: lexer.OPERATOR, Literal: "operator"},
		Kind:           OperatorKindGlobal,
		OperatorToken:  lexer.Token{Type: lexer.PLUS, Literal: "+"},
		OperatorSymbol: "+",
		Arity:          2,
		OperandTypes: []*TypeAnnotation{
			{Token: lexer.Token{Type: lexer.IDENT, Literal: "String"}, Name: "String"},
			{Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"}, Name: "Integer"},
		},
		ReturnType: &TypeAnnotation{Token: lexer.Token{Type: lexer.IDENT, Literal: "String"}, Name: "String"},
		Binding:    &Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "StrPlusInt"}, Value: "StrPlusInt"},
	}

	got := decl.String()
	want := "operator + (String, Integer) : String uses StrPlusInt"

	if got != want {
		t.Fatalf("OperatorDecl.String() = %q, want %q", got, want)
	}
}

func TestOperatorDeclString_Conversion(t *testing.T) {
	decl := &OperatorDecl{
		Token:          lexer.Token{Type: lexer.OPERATOR, Literal: "operator"},
		Kind:           OperatorKindConversion,
		OperatorToken:  lexer.Token{Type: lexer.IMPLICIT, Literal: "implicit"},
		OperatorSymbol: "implicit",
		Arity:          1,
		OperandTypes: []*TypeAnnotation{
			{Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"}, Name: "Integer"},
		},
		ReturnType: &TypeAnnotation{Token: lexer.Token{Type: lexer.IDENT, Literal: "String"}, Name: "String"},
		Binding:    &Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "IntToStr"}, Value: "IntToStr"},
	}

	got := decl.String()
	want := "operator implicit (Integer) : String uses IntToStr"

	if got != want {
		t.Fatalf("OperatorDecl.String() = %q, want %q", got, want)
	}
}

func TestOperatorDeclString_Class(t *testing.T) {
	decl := &OperatorDecl{
		Token:          lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Kind:           OperatorKindClass,
		OperatorToken:  lexer.Token{Type: lexer.PLUS_ASSIGN, Literal: "+="},
		OperatorSymbol: "+=",
		Arity:          1,
		OperandTypes: []*TypeAnnotation{
			{Token: lexer.Token{Type: lexer.IDENT, Literal: "String"}, Name: "String"},
		},
		Binding:    &Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "AppendString"}, Value: "AppendString"},
		Visibility: VisibilityPublic,
	}

	got := decl.String()
	want := "class operator += String uses AppendString"

	if got != want {
		t.Fatalf("OperatorDecl.String() = %q, want %q", got, want)
	}
}
