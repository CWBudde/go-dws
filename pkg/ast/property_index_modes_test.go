package ast_test

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestPropertyIndexModes_StringRoundTrip(t *testing.T) {
	for _, classProperty := range []bool{false, true} {
		prop := &ast.PropertyDecl{Name: &ast.Identifier{Value: "P"}, Type: &ast.TypeAnnotation{Name: "Integer"}, ReadSpec: &ast.Identifier{Value: "Get"}, IsClassProperty: classProperty}
		for _, index := range []struct {
			name, typ      string
			byRef, isConst bool
		}{
			{"Alpha", "Integer", true, false}, {"Beta", "Integer", true, false}, {"Gamma", "String", false, true}, {"Delta", "String", false, true}, {"Epsilon", "Float", false, false},
		} {
			prop.IndexParams = append(prop.IndexParams, &ast.Parameter{Name: &ast.Identifier{Value: index.name}, Type: &ast.TypeAnnotation{Name: index.typ}, ByRef: index.byRef, IsConst: index.isConst})
		}
		prefix := ""
		if classProperty {
			prefix = "class "
		}
		want := prefix + "property P[var Alpha: Integer; var Beta: Integer; const Gamma: String; const Delta: String; Epsilon: Float]: Integer read Get;"
		rendered := prop.String()
		if rendered != want {
			t.Fatalf("String = %q; want %q", rendered, want)
		}
		p := parser.New(lexer.New("type T = class " + rendered + " end;"))
		program := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			t.Fatalf("reparse %q: %v", rendered, errs)
		}
		got := program.Statements[0].(*ast.ClassDecl).Properties[0]
		if got.IsClassProperty != classProperty || len(got.IndexParams) != len(prop.IndexParams) {
			t.Fatalf("reparsed property = %#v", got)
		}
		for i, param := range got.IndexParams {
			original := prop.IndexParams[i]
			if param.Name.Value != original.Name.Value || param.Type.String() != original.Type.String() || param.ByRef != original.ByRef || param.IsConst != original.IsConst {
				t.Fatalf("index %d lost declaration: %#v", i, param)
			}
		}
	}
}
