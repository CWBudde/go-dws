package parser

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestPropertyIndexModes_Groups(t *testing.T) {
	for _, tt := range []struct {
		name, indices       string
		names, types, modes []string
	}{
		{"mixed grouped and reset", "vAr Alpha, Beta: Integer; cOnSt Gamma, Delta: String; Epsilon, Zeta: Float", []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta"}, []string{"Integer", "Integer", "String", "String", "Float", "Float"}, []string{"var", "var", "const", "const", "value", "value"}},
		{"value control", "A, B: Integer; C: String", []string{"A", "B", "C"}, []string{"Integer", "Integer", "String"}, []string{"value", "value", "value"}},
		{"reset before another modifier", "const A: Integer; B: String; var C: Float; D: Boolean", []string{"A", "B", "C", "D"}, []string{"Integer", "String", "Float", "Boolean"}, []string{"const", "value", "var", "value"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser("property P[" + tt.indices + "]: Integer read Get;")
			prop := p.parsePropertyDeclaration()
			if errs := p.Errors(); len(errs) > 0 {
				t.Fatalf("parse: %v", errs)
			}
			if prop == nil || len(prop.IndexParams) != len(tt.names) {
				t.Fatalf("property = %#v", prop)
			}
			for i, param := range prop.IndexParams {
				if param.Name.Value != tt.names[i] || param.Type.String() != tt.types[i] || param.ByRef != (tt.modes[i] == "var") || param.IsConst != (tt.modes[i] == "const") {
					t.Errorf("index %d = %s, var=%v const=%v; want %s: %s %s", i, param, param.ByRef, param.IsConst, tt.names[i], tt.types[i], tt.modes[i])
				}
			}
			if !p.curTokenIs(lexer.SEMICOLON) {
				t.Fatalf("cursor = %v", p.cursor.Current())
			}
		})
	}
}

func TestPropertyIndexModes_EmptyRecovery(t *testing.T) {
	source := "type T = class\n Field: Integer;\n property P[]: Integer read Field;\n property Q: Integer read Field;\nend;"
	p := testParser(source)
	program := p.ParseProgram()
	errs := p.Errors()
	if len(errs) != 1 || errs[0].Message != "Parameters expected" || errs[0].Stop || errs[0].Pos.Line != 3 || errs[0].Pos.Column != 13 {
		t.Fatalf("errors = %#v", errs)
	}
	decl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok || len(decl.Properties) != 2 {
		t.Fatalf("class = %#v", program.Statements)
	}
	if decl.Properties[0].IndexParams == nil || len(decl.Properties[0].IndexParams) != 0 || decl.Properties[1].Name.Value != "Q" {
		t.Fatalf("properties = %#v", decl.Properties)
	}
}

func TestPropertyIndexModes_Stops(t *testing.T) {
	for _, tt := range []struct{ name, tail, message, anchor string }{
		{"missing first name", ": Integer]", "Name expected", ":"},
		{"var missing name", "var ]", "Name expected", "]"},
		{"const missing name EOF", "const {trailing}\n", "Name expected", "const"},
		{"lazy", "lazy I: Integer]", "Name expected", "lazy"},
		{"repeated", "var var I: Integer]", "Name expected", "var"},
		{"conflicting", "var const I: Integer]", "Name expected", "const"},
		{"reverse conflicting", "const var I: Integer]", "Name expected", "var"},
		{"const ref", "const(ref) I: Integer]", "Name expected", "("},
		{"comma missing name", "I, ]", "Name expected", "]"},
		{"missing colon", "I {comment}\n ]", `Colon ":" expected`, "]"},
		{"missing colon EOF", "I {comment}\n", `Colon ":" expected`, "I"},
		{"missing bracket", "I: Integer = 1]", `"]" expected`, "="},
		{"missing bracket EOF", "I: Integer {comment}\n", `"]" expected`, "Integer"},
		{"trailing separator", "I: Integer;]", "Name expected", "]"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "property P[" + tt.tail
			p := testParser(source)
			prop := p.parsePropertyDeclaration()
			errs := p.Errors()
			if prop != nil || len(errs) != 1 || errs[0].Message != tt.message || !errs[0].Stop {
				t.Fatalf("property=%v errors=%#v", prop, errs)
			}
			// Independently read the actual source token whose position owns this stop.
			l := lexer.New(source)
			var anchor lexer.Token
			for tok := l.NextToken(); tok.Type != lexer.EOF; tok = l.NextToken() {
				if tok.Literal == tt.anchor {
					anchor = tok
				}
			}
			if errs[0].Pos != anchor.Pos {
				t.Fatalf("position = %v, want source token %q at %v", errs[0].Pos, tt.anchor, anchor.Pos)
			}
		})
	}
}

func TestPropertyIndexModes_PartialClass(t *testing.T) {
	source := "type T = class\n function Get(a: Integer): Integer;\n property P[a: Integer]: Integer read Get;\n property Q[var a: Integer]: Integer read Get;\n property Stop[a: Integer {EOF}\n"
	p := testParser(source)
	program := p.ParseProgram()
	if got := parserErrorMessages(p); !reflect.DeepEqual(got, []string{`"]" expected`}) || !p.stopped() {
		t.Fatalf("errors = %v", p.Errors())
	}
	if len(program.Statements) != 1 {
		t.Fatalf("statements = %v", program.Statements)
	}
	decl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok || len(decl.Methods) != 1 || len(decl.Properties) != 2 || !decl.Properties[1].IndexParams[0].ByRef {
		t.Fatalf("partial class = %#v", decl)
	}
	if decl.Properties[0].Name.Value != "P" || decl.Properties[0].IndexParams[0].ByRef || decl.Properties[0].IndexParams[0].IsConst {
		t.Fatal("earlier value declaration was changed or lost")
	}
}
