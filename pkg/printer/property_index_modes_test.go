package printer_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/printer"
)

func TestPropertyIndexModes_PrintRoundTrip(t *testing.T) {
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
		decl := &ast.ClassDecl{Name: &ast.Identifier{Value: "T"}, Properties: []*ast.PropertyDecl{prop}}
		rendered := printer.New(printer.DefaultOptions()).Print(decl) + ";"
		p := parser.New(lexer.New(rendered))
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
				t.Fatalf("index %d lost declaration: got %s var=%v const=%v", i, param, param.ByRef, param.IsConst)
			}
		}
	}
}

// Both independent serializers preserve legal const declarations and value-copy
// behavior. Var caller mutation is deliberately a later runtime slice.
func TestPropertyIndexModes_SerializedExecution(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"const", `type T = class
 Field: Integer;
 function Get(const I: Integer): Integer;
 begin Result := I; end;
 procedure SetIt(const I: Integer; V: Integer);
 begin Field := I + V; end;
 property P[const I: Integer]: Integer read Get write SetIt;
end;
var O := new T;
PrintLn(O.P[2 + 3]);
O.P[5] := 3;
PrintLn(O.Field);`, "5\n8\n"},
		{"value copies", `type T = class
 function Get(I: Integer): Integer;
 begin I += 1; Result := I; end;
 procedure SetIt(I: Integer; V: Integer);
 begin I += V; end;
 property P[I: Integer]: Integer read Get write SetIt;
end;
var O := new T;
var X := 5;
PrintLn(O.P[X]);
PrintLn(X);
O.P[X] := 3;
PrintLn(X);`, "6\n5\n5\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.source))
			program := p.ParseProgram()
			if errs := p.Errors(); len(errs) > 0 {
				t.Fatalf("parse: %v", errs)
			}
			prop := program.Statements[0].(*ast.ClassDecl).Properties[0]
			start := strings.Index(tt.source, "property P[")
			end := start + strings.Index(tt.source[start:], ";") + 1
			sources := map[string]string{"original": tt.source, "AST.String": tt.source[:start] + prop.String() + tt.source[end:], "printer": printer.New(printer.DefaultOptions()).Print(program)}
			for serializer, source := range sources {
				t.Run(serializer, func(t *testing.T) {
					var output bytes.Buffer
					engine, err := dwscript.New(dwscript.WithOutput(&output))
					if err != nil {
						t.Fatal(err)
					}
					if _, err = engine.Eval(source); err != nil {
						t.Fatalf("execution: %v\n%s", err, source)
					}
					if got := output.String(); got != tt.want {
						t.Fatalf("output = %q; want %q\n%s", got, tt.want, source)
					}
				})
			}
		})
	}
}
