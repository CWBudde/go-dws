package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func parseGenericProgram(t *testing.T, src string) *ast.Program {
	t.Helper()
	p := New(lexer.New(src))
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors for %q: %v", src, errs)
	}
	return prog
}

func TestParseGenericClassDeclaration(t *testing.T) {
	prog := parseGenericProgram(t, `type TTest<A, B> = class FieldA : A; FieldB : B; end;`)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	class, ok := prog.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected *ast.ClassDecl, got %T", prog.Statements[0])
	}
	if got := class.TypeParams; len(got) != 2 || got[0] != "A" || got[1] != "B" {
		t.Fatalf("TypeParams = %v, want [A B]", got)
	}
}

func TestParseGenericRecordDeclaration(t *testing.T) {
	prog := parseGenericProgram(t, `type TRec<T> = record Value : T; end;`)
	rec, ok := prog.Statements[0].(*ast.RecordDecl)
	if !ok {
		t.Fatalf("expected *ast.RecordDecl, got %T", prog.Statements[0])
	}
	if len(rec.TypeParams) != 1 || rec.TypeParams[0] != "T" {
		t.Fatalf("TypeParams = %v, want [T]", rec.TypeParams)
	}
}

func TestParseGenericTypeAnnotationArgs(t *testing.T) {
	prog := parseGenericProgram(t, `var x : TList<Integer>;`)
	var ta *ast.TypeAnnotation
	ast.Inspect(prog, func(n ast.Node) bool {
		if t, ok := n.(*ast.TypeAnnotation); ok && t.Name == "TList" {
			ta = t
		}
		return true
	})
	if ta == nil {
		t.Fatal("no TypeAnnotation named TList found")
	}
	if len(ta.TypeArgs) != 1 || ta.TypeArgs[0].String() != "Integer" {
		t.Fatalf("TypeArgs = %v, want [Integer]", ta.TypeArgs)
	}
}

func TestParseGenericNewExpression(t *testing.T) {
	prog := parseGenericProgram(t, `var x := new TList<String>;`)
	var ne *ast.NewExpression
	ast.Inspect(prog, func(n ast.Node) bool {
		if e, ok := n.(*ast.NewExpression); ok {
			ne = e
		}
		return true
	})
	if ne == nil {
		t.Fatal("no NewExpression found")
	}
	if ne.ClassName.Value != "TList" {
		t.Fatalf("ClassName = %q, want TList", ne.ClassName.Value)
	}
	if len(ne.TypeArgs) != 1 || ne.TypeArgs[0].String() != "String" {
		t.Fatalf("TypeArgs = %v, want [String]", ne.TypeArgs)
	}
}

func TestParseGenericTypeRefInExpression(t *testing.T) {
	prog := parseGenericProgram(t, `PrintLn(TTest<Integer>.Identity(10));`)
	var gtr *ast.GenericTypeRef
	ast.Inspect(prog, func(n ast.Node) bool {
		if g, ok := n.(*ast.GenericTypeRef); ok {
			gtr = g
		}
		return true
	})
	if gtr == nil {
		t.Fatal("no GenericTypeRef found for TTest<Integer>.Identity")
	}
	if gtr.Base.Value != "TTest" {
		t.Fatalf("Base = %q, want TTest", gtr.Base.Value)
	}
	if len(gtr.TypeArgs) != 1 || gtr.TypeArgs[0].String() != "Integer" {
		t.Fatalf("TypeArgs = %v, want [Integer]", gtr.TypeArgs)
	}
}

// TestParseComparisonNotMistakenForGenerics guards the disambiguation: '<' and
// '>' in ordinary expression position must still parse as comparison operators,
// not as generic type arguments.
func TestParseComparisonNotMistakenForGenerics(t *testing.T) {
	cases := []string{
		`var b := a < c;`,
		`var b := (x < y) and (y > z);`,
		`if a < b then PrintLn('yes');`,
		`while i < n do i := i + 1;`,
	}
	for _, src := range cases {
		prog := parseGenericProgram(t, src)
		saw := false
		ast.Inspect(prog, func(n ast.Node) bool {
			if _, ok := n.(*ast.GenericTypeRef); ok {
				saw = true
			}
			return true
		})
		if saw {
			t.Errorf("comparison %q was wrongly parsed as a generic type reference", src)
		}
	}
}
