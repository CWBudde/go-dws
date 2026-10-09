package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/generics"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestRecordIndexDeclarations_RoundTrip(t *testing.T) {
	source := `type TR = record
 class function Get(var x,y: Integer; const z: Integer; w: Integer): Integer; begin Result := 0; end;
 class property P[var a,b: Integer; const c: Integer; d: Integer]: Integer read Get; default;
end;`
	p := New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatal(p.Errors())
	}
	record := program.Statements[0].(*ast.RecordDecl)
	prop := record.Properties[0]
	printed := "type TR = record " + prop.String() + "; end;"
	p = New(lexer.New(printed))
	round := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("%s: %v", printed, p.Errors())
	}
	other := round.Statements[0].(*ast.RecordDecl).Properties[0]
	if !other.IsClassProperty || !other.IsDefault {
		t.Fatalf("lost flags: %s", printed)
	}
	if len(other.IndexParams) != 4 {
		t.Fatal("lost grouped parameters")
	}
	for i, param := range prop.IndexParams {
		actual := other.IndexParams[i]
		if param.Name.Value != actual.Name.Value || param.ByRef != actual.ByRef || param.IsConst != actual.IsConst || param.Type.String() != actual.Type.String() {
			t.Fatalf("lost parameter %d", i)
		}
	}

}

func TestRecordIndexDeclarations_Composite(t *testing.T) {
	p := New(lexer.New(`type TR = record property P[var a,b: array of Integer; const c: array[0..1] of String]: Integer read (1); end;`))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatal(p.Errors())
	}
	prop := program.Statements[0].(*ast.RecordDecl).Properties[0]
	if len(prop.IndexParams) != 3 || !prop.IndexParams[0].ByRef || !prop.IndexParams[1].ByRef || !prop.IndexParams[2].IsConst {
		t.Fatal("lost composite index parameters")
	}
	if _, ok := prop.IndexParams[0].Type.(*ast.ArrayTypeNode); !ok {
		t.Fatal("dynamic array type lost")
	}
	if _, ok := prop.IndexParams[2].Type.(*ast.ArrayTypeNode); !ok {
		t.Fatal("static array type lost")
	}
}

func TestRecordIndexDeclarations_MissingTypeRecovery(t *testing.T) {
	p := New(lexer.New(`property P[var a,b:; const c: Integer]: Integer read (1);`))
	prop := p.parseRecordPropertyDeclaration()
	errs := p.Errors()
	if prop == nil || len(errs) != 1 || errs[0].Message != "Type expected" || errs[0].Stop {
		t.Fatalf("property %v, errors %+v", prop, errs)
	}
	if p.cursor.Current().Type != lexer.SEMICOLON {
		t.Fatalf("lost declaration cursor: %v", p.cursor.Current())
	}
	assertRecoveredRecordIndexGroups(t, prop)
}

func assertRecoveredRecordIndexGroups(t *testing.T, prop *ast.RecordPropertyDecl) {
	t.Helper()
	if len(prop.IndexParams) != 3 || !prop.IndexParams[0].ByRef || !prop.IndexParams[1].ByRef || !prop.IndexParams[2].IsConst {
		t.Fatal("recovery lost names or modes")
	}
	if prop.IndexParams[0].Name.Value != "a" || prop.IndexParams[1].Name.Value != "b" || prop.IndexParams[2].Name.Value != "c" {
		t.Fatal("recovery lost grouped names")
	}
	if prop.IndexParams[0].Type.String() != "Variant" || prop.IndexParams[1].Type.String() != "Variant" || prop.IndexParams[2].Type.String() != "Integer" {
		t.Fatal("missing-type fallback affected a valid group")
	}
}

func TestRecordIndexDeclarations_MalformedComposite(t *testing.T) {
	p := New(lexer.New(`property P[var a: array of ]: Integer read (1);`))
	prop := p.parseRecordPropertyDeclaration()
	if prop == nil || prop.Type != nil || !p.stopped() {
		t.Fatalf("malformed array type recovered as a valid declaration: %v, %v", prop, p.Errors())
	}
	if len(p.Errors()) < 1 || p.Errors()[0].Message != "expected type expression after 'array of'" {
		t.Fatal(p.Errors())
	}
}

func TestRecordIndexDeclarations_PartialAST(t *testing.T) {
	for _, prefix := range []string{"type TR = record", "var r: record", "var r := record"} {
		t.Run(prefix, func(t *testing.T) {
			p := New(lexer.New(prefix + " property P[var a: Missing : Integer read (1); end;"))
			program := p.ParseProgram()
			if !p.stopped() || len(p.Errors()) != 1 {
				t.Fatal(p.Errors())
			}
			count := 0
			ast.Inspect(program, func(node ast.Node) bool {
				if prop, ok := node.(*ast.RecordPropertyDecl); ok {
					if prop.Type != nil || !strings.Contains(prop.String(), "P[var a: Missing]") {
						t.Fatal("lost partial property")
					}
				}
				if param, ok := node.(*ast.Parameter); ok && param.Name.Value == "a" {
					count++
					if !param.ByRef || param.Type.String() != "Missing" {
						t.Fatal("lost reached index")
					}
				}
				return true
			})
			if count != 1 {
				t.Fatalf("visited %d reached indices", count)
			}
			if printed := program.String(); printed == "" {
				t.Fatal("partial program printing produced no text")
			}
		})
	}
}

func TestRecordIndexDeclarations_GenericClone(t *testing.T) {
	p := New(lexer.New(`type TR<T> = record
 property P[var a: T]: Integer read Get;
 property Q[const b: T : Integer read Get;
end;`))
	program := p.ParseProgram()
	if !p.stopped() {
		t.Fatal("missing bracket did not stop")
	}
	original := program.Statements[0].(*ast.RecordDecl)
	program.Statements = append(program.Statements, &ast.VarDeclStatement{Names: []*ast.Identifier{{Value: "r"}}, Type: &ast.TypeAnnotation{Name: "TR", TypeArgs: []ast.TypeExpression{&ast.TypeAnnotation{Name: "Integer"}}}})
	if err := generics.Monomorphize(program); err != nil {
		t.Fatal(err)
	}
	clone := program.Statements[0].(*ast.RecordDecl)
	if len(clone.Properties) != 2 || clone.Properties[1].Type != nil {
		t.Fatal("lost partial property during cloning")
	}
	for i, prop := range clone.Properties {
		if prop.ReadAccessorPos != original.Properties[i].ReadAccessorPos || prop.IndexParams[0].Name.Value != original.Properties[i].IndexParams[0].Name.Value || prop.IndexParams[0].ByRef != original.Properties[i].IndexParams[0].ByRef || prop.IndexParams[0].IsConst != original.Properties[i].IndexParams[0].IsConst || prop.IndexParams[0].Type.String() != "Integer" {
			t.Fatalf("clone lost signature or anchor %d", i)
		}
	}
	if clone.Properties[0].IndexParams[0] == original.Properties[0].IndexParams[0] {
		t.Fatal("clone shared mutable AST parameters")
	}
}
