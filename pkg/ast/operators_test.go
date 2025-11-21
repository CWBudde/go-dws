package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestOperatorDeclString_GlobalBinary(t *testing.T) {
	decl := &OperatorDecl{
		BaseNode:       NewTestBaseNode(lexer.OPERATOR, "operator"),
		Kind:           OperatorKindGlobal,
		OperatorToken:  lexer.Token{Type: lexer.PLUS, Literal: "+"},
		OperatorSymbol: "+",
		Arity:          2,
		OperandTypes: []TypeExpression{
			NewTestTypeAnnotation("String"),
			NewTestTypeAnnotation("Integer"),
		},
		ReturnType: NewTestTypeAnnotation("String"),
		Binding:    NewTestIdentifier("StrPlusInt"),
	}

	got := decl.String()
	want := "operator + (String, Integer) : String uses StrPlusInt"

	if got != want {
		t.Fatalf("OperatorDecl.String() = %q, want %q", got, want)
	}
}

func TestOperatorDeclString_Conversion(t *testing.T) {
	decl := &OperatorDecl{
		BaseNode:       NewTestBaseNode(lexer.OPERATOR, "operator"),
		Kind:           OperatorKindConversion,
		OperatorToken:  lexer.Token{Type: lexer.IMPLICIT, Literal: "implicit"},
		OperatorSymbol: "implicit",
		Arity:          1,
		OperandTypes: []TypeExpression{
			NewTestTypeAnnotation("Integer"),
		},
		ReturnType: NewTestTypeAnnotation("String"),
		Binding:    NewTestIdentifier("IntToStr"),
	}

	got := decl.String()
	want := "operator implicit (Integer) : String uses IntToStr"

	if got != want {
		t.Fatalf("OperatorDecl.String() = %q, want %q", got, want)
	}
}

func TestOperatorDeclString_Class(t *testing.T) {
	decl := &OperatorDecl{
		BaseNode:       NewTestBaseNode(lexer.CLASS, "class"),
		Kind:           OperatorKindClass,
		OperatorToken:  lexer.Token{Type: lexer.PLUS_ASSIGN, Literal: "+="},
		OperatorSymbol: "+=",
		Arity:          1,
		OperandTypes: []TypeExpression{
			NewTestTypeAnnotation("String"),
		},
		Binding:    NewTestIdentifier("AppendString"),
		Visibility: VisibilityPublic,
	}

	got := decl.String()
	want := "class operator += String uses AppendString"

	if got != want {
		t.Fatalf("OperatorDecl.String() = %q, want %q", got, want)
	}
}
