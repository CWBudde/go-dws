package generics

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func parseProgram(t *testing.T, src string) *ast.Program {
	t.Helper()
	p := parser.New(lexer.New(src))
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}
	return prog
}

// declNames returns the names of all top-level type declarations in order.
func declNames(prog *ast.Program) []string {
	var names []string
	for _, stmt := range prog.Statements {
		if n := declName(stmt); n != "" {
			names = append(names, n)
		}
	}
	return names
}

func findClass(prog *ast.Program, name string) *ast.ClassDecl {
	for _, stmt := range prog.Statements {
		if c, ok := stmt.(*ast.ClassDecl); ok && c.Name != nil && c.Name.Value == name {
			return c
		}
	}
	return nil
}

func findRecord(prog *ast.Program, name string) *ast.RecordDecl {
	for _, stmt := range prog.Statements {
		if r, ok := stmt.(*ast.RecordDecl); ok && r.Name != nil && r.Name.Value == name {
			return r
		}
	}
	return nil
}

func TestMonomorphize_NoGenerics_LeavesProgramUnchanged(t *testing.T) {
	prog := parseProgram(t, `type TFoo = class Field : Integer; end;
var f := new TFoo;`)
	before := len(prog.Statements)
	Monomorphize(prog)
	if len(prog.Statements) != before {
		t.Fatalf("expected %d statements, got %d", before, len(prog.Statements))
	}
	if findClass(prog, "TFoo") == nil {
		t.Fatal("TFoo class declaration was unexpectedly removed")
	}
}

func TestMonomorphize_GenericClass_SpecializesAndSubstitutes(t *testing.T) {
	prog := parseProgram(t, `type TBox<T> = class Value : T; end;
var b := new TBox<Integer>;`)
	Monomorphize(prog)

	// The generic template must be removed and replaced by the specialization.
	if findClass(prog, "TBox") != nil {
		t.Fatal("generic template TBox should have been removed")
	}
	spec := findClass(prog, "TBox<Integer>")
	if spec == nil {
		t.Fatalf("expected specialized class TBox<Integer>; decls: %v", declNames(prog))
	}
	if len(spec.TypeParams) != 0 {
		t.Errorf("specialized class should have no type params, got %v", spec.TypeParams)
	}
	if len(spec.Fields) != 1 || spec.Fields[0].Type == nil || spec.Fields[0].Type.String() != "Integer" {
		t.Errorf("expected field of type Integer after substitution, got %+v", spec.Fields)
	}

	// The `new` expression must reference the mangled name with no type args.
	var found bool
	ast.Inspect(prog, func(n ast.Node) bool {
		if ne, ok := n.(*ast.NewExpression); ok {
			found = true
			if ne.ClassName.Value != "TBox<Integer>" {
				t.Errorf("new expression class name = %q, want TBox<Integer>", ne.ClassName.Value)
			}
			if ne.TypeArgs != nil {
				t.Errorf("new expression type args should be cleared, got %v", ne.TypeArgs)
			}
		}
		return true
	})
	if !found {
		t.Fatal("no NewExpression found in program")
	}
}

func TestMonomorphize_SpecializationInsertedBeforeUse(t *testing.T) {
	prog := parseProgram(t, `type TBox<T> = class Value : T; end;
var a := new TBox<Integer>;
var b := new TBox<String>;`)
	Monomorphize(prog)

	names := declNames(prog)
	// Two distinct specializations, each declared before its use.
	want := []string{"TBox<Integer>", "TBox<String>"}
	for _, w := range want {
		if findClass(prog, w) == nil {
			t.Errorf("missing specialization %s; decls: %v", w, names)
		}
	}
}

func TestMonomorphize_SameInstantiationEmittedOnce(t *testing.T) {
	prog := parseProgram(t, `type TBox<T> = class Value : T; end;
var a := new TBox<Integer>;
var b := new TBox<Integer>;`)
	Monomorphize(prog)

	count := 0
	for _, name := range declNames(prog) {
		if name == "TBox<Integer>" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one TBox<Integer> specialization, got %d", count)
	}
}

func TestMonomorphize_GenericRecord_TwoParams(t *testing.T) {
	prog := parseProgram(t, `type TPair<A, B> = record First : A; Second : B; end;
var p : TPair<Integer, String>;`)
	Monomorphize(prog)

	spec := findRecord(prog, "TPair<Integer,String>")
	if spec == nil {
		t.Fatalf("expected specialized record TPair<Integer,String>; decls: %v", declNames(prog))
	}
	if len(spec.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(spec.Fields))
	}
	if got := spec.Fields[0].Type.String(); got != "Integer" {
		t.Errorf("First field type = %q, want Integer", got)
	}
	if got := spec.Fields[1].Type.String(); got != "String" {
		t.Errorf("Second field type = %q, want String", got)
	}
}
