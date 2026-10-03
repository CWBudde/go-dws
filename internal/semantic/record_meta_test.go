package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestRecordMeta_Inference(t *testing.T) {
	a := parseAndAnalyze(t, `type R = record x: Integer; end; type A = R;
 var direct := R; var aliasValue := A; var copied := direct; var instance: A;`)
	if len(a.Errors()) != 0 {
		t.Fatal(a.Errors())
	}
	var meta types.Type
	for _, name := range []string{"direct", "aliasValue", "copied"} {
		sym, ok := a.symbols.Resolve(name)
		if !ok || sym.Type.TypeKind() != "RECORD_META" || sym.Type.String() != "meta of R" {
			t.Fatalf("%s = %v, want meta of R", name, sym)
		}
		if meta != nil && !meta.Equals(sym.Type) {
			t.Errorf("%s lost canonical metatype", name)
		}
		meta = sym.Type
	}
	sym, _ := a.symbols.Resolve("instance")
	if sym.Type.TypeKind() != "RECORD" || meta.Equals(sym.Type) || sym.Type.Equals(meta) {
		t.Fatalf("annotation = %v, meta = %v", sym.Type, meta)
	}
	if types.GetUnderlyingType(meta).TypeKind() != "RECORD_META" {
		t.Fatal("metatype erased by unwrapping")
	}
}

func TestRecordMeta_AssignmentOwnership(t *testing.T) {
	for _, tt := range []struct {
		name, tail string
		wantError  bool
	}{
		{"same meta", `var m := R; var n := R; m := n;`, false},
		{"instance to meta", `var m := R; var item: R; m := item;`, true},
		{"meta to instance", `var item: R; item := R;`, true},
		{"different record", `var m := R; m := S;`, true},
		{"record parameter", `procedure Use(r: R); begin end; Use(R);`, true},
		{"instance field on meta", `var m := R; PrintLn(m.x);`, true},
		{"shadow remains scalar", `procedure Test; begin var R := 4; var n := R; n := 5; end; Test;`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := parseAndAnalyze(t, `type R = record x: Integer; end; type S = record x: Integer; end; `+tt.tail)
			if (len(a.Errors()) != 0) != tt.wantError {
				t.Fatalf("errors = %v", a.Errors())
			}
		})
	}
}

func TestRecordMeta_SelfOwnership(t *testing.T) {
	for _, tt := range []struct {
		name, declaration string
		wantError         bool
	}{
		{"ordinary class method has no Self", `type R = record class function F: Integer; begin Result := Self.F; end; end;`, true},
		{"helper class Self is meta", `type R = record x: Integer; end; type H = helper for R class function F: Integer; begin var m := R; m := Self; Result := 1; end; end;`, false},
		{"helper class Self has no instance fields", `type R = record x: Integer; end; type H = helper for R class function F: Integer; begin Result := Self.x; end; end;`, true},
		{"helper class bare field unavailable", `type R = record x: Integer; end; type H = helper for R class function F: Integer; begin Result := x; end; end;`, true},
		{"static helper has no Self", `type R = record x: Integer; end; type H = helper for R class function F: Integer; static; begin Result := Self.x; end; end;`, true},
		{"instance helper unavailable on meta", `type R = record x: Integer; end; type H = helper for R function F: Integer; begin Result := Self.x; end; end; PrintLn(R.F);`, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := parseAndAnalyze(t, tt.declaration)
			if (len(a.Errors()) != 0) != tt.wantError {
				t.Fatalf("errors = %v", a.Errors())
			}
		})
	}
}

func TestRecordMeta_OrdinaryClassSelfAndDeclarationAssignment(t *testing.T) {
	for _, source := range []string{
		`type R = record x: Integer; class function F: Integer; begin Result := Self.x; end; end;`,
		`type R = record x: Integer; class function F: Integer; begin Result := x; end; end;`,
		`type R = record x: Integer; end; R := R;`,
	} {
		a := parseAndAnalyze(t, source)
		if len(a.Errors()) == 0 {
			t.Errorf("expected ownership error for %s", source)
		}
	}
}

func TestRecordMeta_CastTargetValueCategory(t *testing.T) {
	for _, right := range []string{"R", "A", "m"} {
		for _, op := range []string{"as", "is"} {
			t.Run(op+" "+right, func(t *testing.T) {
				p := parser.New(lexer.New(`type R=record x:Integer; end; type A=R; var m:=R; var obj:TObject; var result:=obj ` + op + ` ` + right + `;`))
				program := p.ParseProgram()
				if len(p.Errors()) != 0 {
					t.Fatal(p.Errors())
				}
				a := NewAnalyzer()
				_ = a.Analyze(program)
				if len(a.Errors()) != 1 || !strings.Contains(a.Errors()[0], "Class reference expected") {
					t.Fatal(a.Errors())
				}
				value := program.Statements[len(program.Statements)-1].(*ast.VarDeclStatement).Value
				var target ast.Node
				switch expr := value.(type) {
				case *ast.AsExpression:
					target = expr.TargetType
				case *ast.IsExpression:
					target = expr.TargetType
				}
				typ := a.GetSemanticInfo().GetResolvedType(target)
				if typ == nil || typ.TypeKind() != "RECORD_META" {
					t.Fatalf("cast RHS type = %v", typ)
				}
			})
		}
	}
}

func TestRecordMeta_HelperOverloadOwnership(t *testing.T) {
	const declaration = `type R=record x: Integer; end;
type H=helper for R
 class function Pick(n: Integer): String; overload; begin Result:='class'; end;
 function Pick(s: String): String; overload; begin Result:='instance'; end;
end;`
	for _, tt := range []struct {
		call  string
		valid bool
	}{
		{`PrintLn(R.Pick(1));`, true},
		{`PrintLn(R.Pick('bad'));`, false},
		{`var item:R; PrintLn(item.Pick('ok'));`, true},
	} {
		t.Run(tt.call, func(t *testing.T) {
			a := parseAndAnalyze(t, declaration+tt.call)
			if (len(a.Errors()) == 0) != tt.valid {
				t.Fatal(a.Errors())
			}
		})
	}
}

func TestRecordMeta_DefaultPropertyIndexOwnership(t *testing.T) {
	const declaration = `type R=record
 class function Get(i:Integer; s:String):Integer; begin Result:=i; end;
 property Items[i:Integer; s:String]:Integer read Get; default;
end;
var values := [R];`
	for _, tt := range []struct {
		expression string
		valid      bool
	}{
		{`values[0][1,'ok']`, true},
		{`(values[0])[1,'ok']`, true},
		{`values[0].Items[1,'ok']`, true},
		{`values[0]['bad','ok']`, false},
		{`values[0][1,2]`, false},
		{`values[0][1]`, false},
		{`values[0][1,'ok',2]`, false},
		{`values[0][1]['ok']`, false},
	} {
		t.Run(tt.expression, func(t *testing.T) {
			a := parseAndAnalyze(t, declaration+`PrintLn(`+tt.expression+`);`)
			if (len(a.Errors()) == 0) != tt.valid {
				t.Fatal(a.Errors())
			}
		})
	}
}
