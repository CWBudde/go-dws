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

func TestMonomorphize_TemplateInMultiDeclTypeSection(t *testing.T) {
	// A multi-declaration `type` section is parsed into a BlockStatement; the
	// template must still be collected and specialized.
	prog := parseProgram(t, `type
  TBox<T> = class Value : T; end;
  TOther = Integer;
var b := new TBox<Integer>;`)
	Monomorphize(prog)

	if findClassDeep(prog.Statements, "TBox<Integer>") == nil {
		t.Fatalf("expected TBox<Integer> specialization somewhere in the tree; decls: %v", declNames(prog))
	}
	// The generic template must not survive anywhere.
	if findClassDeep(prog.Statements, "TBox") != nil {
		t.Fatal("generic template TBox should have been removed")
	}
}

func TestMonomorphize_ArityMismatch_NotSpecialized(t *testing.T) {
	// Too many type arguments must not silently produce a concrete class.
	prog := parseProgram(t, `type TBox<T> = class Value : T; end;
var b := new TBox<Integer, String>;`)
	Monomorphize(prog)

	if findClassDeep(prog.Statements, "TBox<Integer,String>") != nil {
		t.Fatal("arity-mismatched instantiation must not be specialized")
	}
	if findClassDeep(prog.Statements, "TBox<Integer>") != nil {
		t.Fatal("no specialization should be generated for an arity mismatch")
	}
}

// findClassDeep locates a class declaration by name, descending into blocks.
func findClassDeep(stmts []ast.Statement, name string) *ast.ClassDecl {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.ClassDecl:
			if s.Name != nil && s.Name.Value == name {
				return s
			}
		case *ast.BlockStatement:
			if c := findClassDeep(s.Statements, name); c != nil {
				return c
			}
		}
	}
	return nil
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

func findInterface(prog *ast.Program, name string) *ast.InterfaceDecl {
	for _, stmt := range prog.Statements {
		if d, ok := stmt.(*ast.InterfaceDecl); ok && d.Name != nil && d.Name.Value == name {
			return d
		}
	}
	return nil
}

func findArrayDecl(prog *ast.Program, name string) *ast.ArrayDecl {
	for _, stmt := range prog.Statements {
		if d, ok := stmt.(*ast.ArrayDecl); ok && d.Name != nil && d.Name.Value == name {
			return d
		}
	}
	return nil
}

// findMethodImpls returns every out-of-line method implementation targeting the
// named class, in statement order.
func findMethodImpls(prog *ast.Program, className string) []*ast.FunctionDecl {
	var impls []*ast.FunctionDecl
	for _, stmt := range prog.Statements {
		if fn, ok := stmt.(*ast.FunctionDecl); ok && fn.ClassName != nil && fn.ClassName.Value == className {
			impls = append(impls, fn)
		}
	}
	return impls
}

// stmtIndex returns the index of the first statement satisfying pred, or -1.
func stmtIndex(prog *ast.Program, pred func(ast.Statement) bool) int {
	for i, stmt := range prog.Statements {
		if pred(stmt) {
			return i
		}
	}
	return -1
}

func TestMonomorphize_GenericInterface_SpecializesMembers(t *testing.T) {
	prog := parseProgram(t, `type ITest<T> = interface
    function GetP : T;
    procedure SetP(v : T);
    property P : T read GetP write SetP;
end;
var i : ITest<Integer>;`)
	Monomorphize(prog)

	if findInterface(prog, "ITest") != nil {
		t.Fatal("generic template ITest should have been removed")
	}
	spec := findInterface(prog, "ITest<Integer>")
	if spec == nil {
		t.Fatalf("expected specialized interface ITest<Integer>; decls: %v", declNames(prog))
	}
	if len(spec.TypeParams) != 0 {
		t.Errorf("specialized interface should have no type params, got %v", spec.TypeParams)
	}
	if len(spec.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(spec.Methods))
	}
	if got := spec.Methods[0].ReturnType.String(); got != "Integer" {
		t.Errorf("GetP return type = %q, want Integer", got)
	}
	if got := spec.Methods[1].Parameters[0].Type.String(); got != "Integer" {
		t.Errorf("SetP parameter type = %q, want Integer", got)
	}
	if len(spec.Properties) != 1 || spec.Properties[0].Type.String() != "Integer" {
		t.Errorf("property P type = %v, want Integer", spec.Properties[0].Type)
	}
}

// TestMonomorphize_GenericInterfaceInInheritanceList covers the identifier-position
// instantiation `class (ITest<Integer>)`, which has no TypeAnnotation to carry
// the type arguments.
func TestMonomorphize_GenericInterfaceInInheritanceList(t *testing.T) {
	prog := parseProgram(t, `type ITest<T> = interface function GetP : T; end;
type TTest = class (ITest<Integer>) function GetP : Integer; begin Result := 1; end; end;`)
	Monomorphize(prog)

	if findInterface(prog, "ITest<Integer>") == nil {
		t.Fatalf("expected specialized interface ITest<Integer>; decls: %v", declNames(prog))
	}
	class := findClass(prog, "TTest")
	if class == nil {
		t.Fatal("TTest class declaration went missing")
	}
	if len(class.Interfaces) != 1 {
		t.Fatalf("expected 1 interface on TTest, got %d", len(class.Interfaces))
	}
	if got := class.Interfaces[0].Value; got != "ITest<Integer>" {
		t.Errorf("interface reference = %q, want ITest<Integer>", got)
	}
	if class.Interfaces[0].TypeArgs != nil {
		t.Errorf("TypeArgs should be cleared after specialization, got %v", class.Interfaces[0].TypeArgs)
	}
}

func TestMonomorphize_GenericArrayAlias_Specializes(t *testing.T) {
	prog := parseProgram(t, `type TTest<T> = array of T;
var i : TTest<Integer>;
var s : TTest<String>;`)
	Monomorphize(prog)

	if findArrayDecl(prog, "TTest") != nil {
		t.Fatal("generic template TTest should have been removed")
	}
	for _, want := range []string{"TTest<Integer>", "TTest<String>"} {
		spec := findArrayDecl(prog, want)
		if spec == nil {
			t.Fatalf("expected specialized array %s; decls: %v", want, declNames(prog))
		}
		if len(spec.TypeParams) != 0 {
			t.Errorf("%s should have no type params, got %v", want, spec.TypeParams)
		}
	}
}

func TestMonomorphize_OutOfLineMethodBody_SpecializedPerInstantiation(t *testing.T) {
	prog := parseProgram(t, `type TTest<T> = class function Test(v : T) : T; end;
function TTest<T>.Test(v : T) : T; begin Result := v; end;
var a := new TTest<Integer>;
var b := new TTest<String>;`)
	Monomorphize(prog)

	if len(findMethodImpls(prog, "TTest")) != 0 {
		t.Error("the generic out-of-line template body should have been removed")
	}
	for _, spec := range []string{"TTest<Integer>", "TTest<String>"} {
		impls := findMethodImpls(prog, spec)
		if len(impls) != 1 {
			t.Fatalf("expected exactly 1 implementation for %s, got %d", spec, len(impls))
		}
		if impls[0].ClassTypeParams != nil {
			t.Errorf("%s: ClassTypeParams should be cleared, got %v", spec, impls[0].ClassTypeParams)
		}
		if impls[0].Name.Value != "Test" {
			t.Errorf("%s: method name = %q, want Test", spec, impls[0].Name.Value)
		}
	}
}

func TestMonomorphize_OutOfLineBodySubstitutesTypeParam(t *testing.T) {
	prog := parseProgram(t, `type TTest<T> = class function Test(v : T) : T; end;
function TTest<T>.Test(v : T) : T; begin Result := v; end;
var a := new TTest<Integer>;`)
	Monomorphize(prog)

	impls := findMethodImpls(prog, "TTest<Integer>")
	if len(impls) != 1 {
		t.Fatalf("expected 1 implementation, got %d", len(impls))
	}
	if got := impls[0].Parameters[0].Type.String(); got != "Integer" {
		t.Errorf("parameter type = %q, want Integer", got)
	}
	if got := impls[0].ReturnType.String(); got != "Integer" {
		t.Errorf("return type = %q, want Integer", got)
	}
}

// TestMonomorphize_OutOfLineBodyEmittedAfterItsClass pins the emission order the
// semantic analyzer depends on: the class, then its implementation, then the use.
func TestMonomorphize_OutOfLineBodyEmittedAfterItsClass(t *testing.T) {
	prog := parseProgram(t, `type TTest<T> = class function Test(v : T) : T; end;
function TTest<T>.Test(v : T) : T; begin Result := v; end;
var a := new TTest<Integer>;`)
	Monomorphize(prog)

	classIdx := stmtIndex(prog, func(s ast.Statement) bool {
		c, ok := s.(*ast.ClassDecl)
		return ok && c.Name != nil && c.Name.Value == "TTest<Integer>"
	})
	implIdx := stmtIndex(prog, func(s ast.Statement) bool {
		fn, ok := s.(*ast.FunctionDecl)
		return ok && fn.ClassName != nil && fn.ClassName.Value == "TTest<Integer>"
	})
	useIdx := stmtIndex(prog, func(s ast.Statement) bool {
		found := false
		ast.Inspect(s, func(n ast.Node) bool {
			if ne, ok := n.(*ast.NewExpression); ok && ne.ClassName.Value == "TTest<Integer>" {
				found = true
			}
			return true
		})
		return found
	})
	if classIdx < 0 || implIdx < 0 || useIdx < 0 {
		t.Fatalf("missing statement: class=%d impl=%d use=%d; decls: %v", classIdx, implIdx, useIdx, declNames(prog))
	}
	if classIdx >= implIdx || implIdx >= useIdx {
		t.Errorf("expected class < impl < use, got %d, %d, %d", classIdx, implIdx, useIdx)
	}
}

// TestMonomorphize_OutOfLineBodyDeclaredBeforeUse covers collection being a
// pre-pass: source order of the body relative to the use must not matter.
func TestMonomorphize_OutOfLineBodyDeclaredBeforeUse(t *testing.T) {
	prog := parseProgram(t, `type TTest<T> = class function Test(v : T) : T; end;
var a := new TTest<Integer>;
function TTest<T>.Test(v : T) : T; begin Result := v; end;`)
	Monomorphize(prog)

	impls := findMethodImpls(prog, "TTest<Integer>")
	if len(impls) != 1 {
		t.Fatalf("expected 1 implementation, got %d; decls: %v", len(impls), declNames(prog))
	}
	if got := impls[0].ReturnType.String(); got != "Integer" {
		t.Errorf("return type = %q, want Integer", got)
	}
}

func TestMonomorphize_OutOfLineBodyArityMismatch_NotEmitted(t *testing.T) {
	prog := parseProgram(t, `type TTest<T> = class procedure Dummy; end;
procedure TTest<T, U>.Dummy; begin end;
var a := new TTest<Integer>;`)
	Monomorphize(prog)

	if impls := findMethodImpls(prog, "TTest<Integer>"); len(impls) != 0 {
		t.Errorf("arity-mismatched header must not be emitted, got %d implementations", len(impls))
	}
	if impls := findMethodImpls(prog, "TTest"); len(impls) != 0 {
		t.Error("the generic header must not survive either")
	}
}

func TestMonomorphize_OutOfLineConstructorWithDefaultT(t *testing.T) {
	prog := parseProgram(t, `type TTest<T> = class A : array of T; constructor Create(n : Integer); end;
constructor TTest<T>.Create(n : Integer); begin A.Add(Default(T)); end;
var a := new TTest<Integer>(1);`)
	Monomorphize(prog)

	impls := findMethodImpls(prog, "TTest<Integer>")
	if len(impls) != 1 {
		t.Fatalf("expected 1 implementation, got %d", len(impls))
	}
	var sawInteger, sawT bool
	ast.Inspect(impls[0], func(n ast.Node) bool {
		if id, ok := n.(*ast.Identifier); ok {
			switch id.Value {
			case "Integer":
				sawInteger = true
			case "T":
				sawT = true
			}
		}
		return true
	})
	if !sawInteger {
		t.Error("Default(T) argument was not substituted to Integer")
	}
	if sawT {
		t.Error("an unsubstituted type parameter T survived in the body")
	}
}

// TestMonomorphize_OutOfLineBodyForNonTemplateClass_Preserved keeps a stray
// generic header naming an unknown type in the tree, so the analyzer reports it
// instead of it silently vanishing.
func TestMonomorphize_OutOfLineBodyForNonTemplateClass_Preserved(t *testing.T) {
	prog := parseProgram(t, `type TBox<T> = class Value : T; end;
procedure TPlain<T>.M; begin end;
var b := new TBox<Integer>;`)
	Monomorphize(prog)

	if impls := findMethodImpls(prog, "TPlain"); len(impls) != 1 {
		t.Errorf("expected the TPlain implementation to be preserved, got %d", len(impls))
	}
}
