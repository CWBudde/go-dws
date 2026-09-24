package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestArrayType_RetainsBoundsDuringRecovery(t *testing.T) {
	for _, source := range []string{
		"var a: array[1.. TProc; var b: Integer;",
		"var a: array[1..2 of Integer; var b: Integer;",
		"var a: array[1..2] Integer; var b: Integer;",
	} {
		t.Run(source, func(t *testing.T) {
			p := New(lexer.New(source))
			program := p.ParseProgram()
			if len(p.Errors()) != 1 {
				t.Fatalf("errors = %v, want one", p.Errors())
			}
			if len(program.Statements) != 2 {
				t.Fatalf("statements = %d, want two", len(program.Statements))
			}
			decl, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("first statement = %T", program.Statements[0])
			}
			array, ok := decl.Type.(*ast.ArrayTypeNode)
			if !ok || array.LowBound == nil || array.HighBound == nil {
				t.Fatalf("array bounds lost: %#v", decl.Type)
			}
		})
	}
}

func TestArrayDeclaration_RetainsBoundsDuringRecovery(t *testing.T) {
	for _, source := range []string{
		"type A = array[1.. TProc; var b: Integer;",
		"type A = array[1..2 of Integer; var b: Integer;",
		"type A = array[1..2] Integer; var b: Integer;",
	} {
		t.Run(source, func(t *testing.T) {
			p := New(lexer.New(source))
			program := p.ParseProgram()
			if len(p.Errors()) != 1 {
				t.Fatalf("errors = %v, want one", p.Errors())
			}
			if len(program.Statements) != 2 {
				t.Fatalf("statements = %d, want two", len(program.Statements))
			}
			decl, ok := program.Statements[0].(*ast.ArrayDecl)
			if !ok || decl.ArrayType == nil || decl.ArrayType.LowBound == nil || decl.ArrayType.HighBound == nil {
				t.Fatalf("array bounds lost: %#v", program.Statements[0])
			}
		})
	}
}
