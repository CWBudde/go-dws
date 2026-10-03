package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestCastRecovery_ResultTypes(t *testing.T) {
	for _, tt := range []struct {
		expr, want string
		errors     int
	}{
		{`o as "hello"`, "TObject", 1},
		{`"world" as TClass`, "TClass", 1},
		{`c as TObject`, "class of TObject", 0},
		{`o as TClass`, "TObject", 0},
		{`o as c`, "TObject", 0},
		{`o as i`, "TObject", 1},
		{`o is i`, "Boolean", 1},
		{`i is TObject`, "Boolean", 1},
		{`o = i`, "Boolean", 1},
		{`o < o`, "Boolean", 1},
		{`i is i`, "Boolean", 1},
		{`i = o`, "Boolean", 1},
		{`c = i`, "Boolean", 1},
		{`i <> c`, "Boolean", 1},
	} {
		t.Run(tt.expr, func(t *testing.T) {
			p := parser.New(lexer.New("var o: TObject; var c: TClass; var i: Integer; var result := " + tt.expr + ";"))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatal(p.Errors())
			}
			a := NewAnalyzer()
			_ = a.Analyze(program)
			if len(a.Errors()) != tt.errors {
				t.Fatalf("errors: %v", a.Errors())
			}
			decl := program.Statements[len(program.Statements)-1].(*ast.VarDeclStatement)
			resolved := a.GetSemanticInfo().GetResolvedType(decl.Value)
			if resolved == nil || resolved.String() != tt.want {
				t.Fatalf("resolved = %v, want %s", resolved, tt.want)
			}
			sym, ok := a.symbols.Resolve("result")
			if !ok || sym.Type.String() != tt.want {
				t.Fatalf("inferred symbol = %v", sym)
			}
		})
	}
}

func TestCastRecovery_CompilerStop(t *testing.T) {
	a := parseAndAnalyze(t, "var c: TClass; var o: TObject; c := c as o; Missing;")
	if !a.compileStopped {
		t.Fatal("analyzer did not stop")
	}
	errs := a.StructuredErrors()
	if len(errs) == 0 || !errs[0].Stop || errs[0].Message != "Class reference expected" {
		t.Fatalf("errors = %v", errs)
	}
}

func TestCastTarget_QualifiedValueUsageAndHints(t *testing.T) {
	source := "type THolder = record Target: TClass; end;\nvar System: THolder;\nvar Obj: TObject;\nPrintLn(Obj is system.target);\nPrintLn(Obj as system.target);"
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatal(p.Errors())
	}
	a := NewAnalyzer()
	a.SetHintsLevel(HintsLevelPedantic)
	_ = a.Analyze(program)
	for _, diagnostic := range a.StructuredErrors() {
		if diagnostic.Severity == SeverityError {
			t.Fatal(diagnostic)
		}
	}
	symbol, found := a.symbols.Resolve("System")
	if !found || len(symbol.Usages) == 0 {
		t.Fatal("qualified lexical receiver usage was lost")
	}
	want := []string{
		`Hint: "system" does not match case of declaration ("System") [line: 4, column: 16]`,
		`Hint: "target" does not match case of declaration ("Target") [line: 4, column: 23]`,
		`Hint: "system" does not match case of declaration ("System") [line: 5, column: 16]`,
		`Hint: "target" does not match case of declaration ("Target") [line: 5, column: 23]`,
	}
	if strings.Join(a.Errors(), "\n") != strings.Join(want, "\n") {
		t.Fatalf("hints = %v; want %v", a.Errors(), want)
	}
}

func TestCastTarget_QualifiedTypeDeprecation(t *testing.T) {
	source := "type TOuter = class\n type TInner = class end; deprecated;\nend;\nvar Obj: TObject;\nPrintLn(Obj is TOuter.TInner);\nPrintLn(Obj as TOuter.TInner);"
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatal(p.Errors())
	}
	a := NewAnalyzer()
	_ = a.Analyze(program)
	for _, diagnostic := range a.StructuredErrors() {
		if diagnostic.Severity == SeverityError {
			t.Fatal(diagnostic)
		}
	}
	want := []string{
		`Warning: "TOuter.TInner" has been deprecated [line: 5, column: 16]`,
		`Warning: "TOuter.TInner" has been deprecated [line: 6, column: 16]`,
	}
	if strings.Join(a.Errors(), "\n") != strings.Join(want, "\n") {
		t.Fatalf("warnings = %v; want %v", a.Errors(), want)
	}
}
