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

func TestParseGenericInterfaceDeclaration(t *testing.T) {
	prog := parseGenericProgram(t, `type ITest<T> = interface function GetP : T; end;`)
	decl, ok := prog.Statements[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("expected *ast.InterfaceDecl, got %T", prog.Statements[0])
	}
	if len(decl.TypeParams) != 1 || decl.TypeParams[0] != "T" {
		t.Fatalf("TypeParams = %v, want [T]", decl.TypeParams)
	}
}

func TestParseGenericArrayDeclaration(t *testing.T) {
	prog := parseGenericProgram(t, `type TTest<T> = array of T;`)
	decl, ok := prog.Statements[0].(*ast.ArrayDecl)
	if !ok {
		t.Fatalf("expected *ast.ArrayDecl, got %T", prog.Statements[0])
	}
	if len(decl.TypeParams) != 1 || decl.TypeParams[0] != "T" {
		t.Fatalf("TypeParams = %v, want [T]", decl.TypeParams)
	}
}

// TestParseGenericInterfaceInInheritanceList covers `class (ITest<Integer>)`:
// the identifier keeps the base name and carries the type arguments, which the
// monomorphizer later replaces with the mangled specialization name.
func TestParseGenericInterfaceInInheritanceList(t *testing.T) {
	prog := parseGenericProgram(t, `type TTest = class (ITest<Integer>) Field : Integer; end;`)
	class, ok := prog.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected *ast.ClassDecl, got %T", prog.Statements[0])
	}
	if len(class.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(class.Interfaces))
	}
	iface := class.Interfaces[0]
	if iface.Value != "ITest" {
		t.Errorf("interface name = %q, want ITest (base name, unmangled)", iface.Value)
	}
	if len(iface.TypeArgs) != 1 || iface.TypeArgs[0].String() != "Integer" {
		t.Fatalf("TypeArgs = %v, want [Integer]", iface.TypeArgs)
	}
}

func TestParseOutOfLineGenericMethod(t *testing.T) {
	prog := parseGenericProgram(t, `type TTest<T> = class function Test(v : T) : T; end;
function TTest<T>.Test(v : T) : T; begin Result := v; end;`)
	fn := lastFunctionDecl(t, prog)
	if fn.ClassName == nil || fn.ClassName.Value != "TTest" {
		t.Errorf("ClassName = %v, want TTest (base name, unmangled)", fn.ClassName)
	}
	if len(fn.ClassTypeParams) != 1 || fn.ClassTypeParams[0] != "T" {
		t.Errorf("ClassTypeParams = %v, want [T]", fn.ClassTypeParams)
	}
	if fn.Name.Value != "Test" {
		t.Errorf("Name = %q, want Test", fn.Name.Value)
	}
	if fn.Body == nil {
		t.Error("expected a body on the out-of-line implementation")
	}
}

func TestParseOutOfLineGenericConstructor(t *testing.T) {
	prog := parseGenericProgram(t, `type TTest<T> = class constructor Create(n : Integer); end;
constructor TTest<T>.Create(n : Integer); begin end;`)
	fn := lastFunctionDecl(t, prog)
	if !fn.IsConstructor {
		t.Error("expected IsConstructor")
	}
	if len(fn.ClassTypeParams) != 1 || fn.ClassTypeParams[0] != "T" {
		t.Errorf("ClassTypeParams = %v, want [T]", fn.ClassTypeParams)
	}
}

func TestParseOutOfLineGenericMethodTwoParams(t *testing.T) {
	prog := parseGenericProgram(t, `type TPair<A, B> = class procedure Swap; end;
procedure TPair<A, B>.Swap; begin end;`)
	fn := lastFunctionDecl(t, prog)
	if got := fn.ClassTypeParams; len(got) != 2 || got[0] != "A" || got[1] != "B" {
		t.Errorf("ClassTypeParams = %v, want [A B]", got)
	}
}

func TestParseOutOfLineGenericClassMethod(t *testing.T) {
	prog := parseGenericProgram(t, `type TTest<T> = class class procedure P; end;
class procedure TTest<T>.P; begin end;`)
	fn := lastFunctionDecl(t, prog)
	if !fn.IsClassMethod {
		t.Error("expected IsClassMethod")
	}
	if len(fn.ClassTypeParams) != 1 || fn.ClassTypeParams[0] != "T" {
		t.Errorf("ClassTypeParams = %v, want [T]", fn.ClassTypeParams)
	}
}

// TestParseNonGenericQualifiedMethodUnaffected guards that adding type-parameter
// support to qualified names left ordinary method implementations alone.
func TestParseNonGenericQualifiedMethodUnaffected(t *testing.T) {
	prog := parseGenericProgram(t, `procedure TFoo.Bar; begin end;`)
	fn := lastFunctionDecl(t, prog)
	if fn.ClassTypeParams != nil {
		t.Errorf("ClassTypeParams = %v, want nil", fn.ClassTypeParams)
	}
	if fn.ClassName == nil || fn.ClassName.Value != "TFoo" {
		t.Errorf("ClassName = %v, want TFoo", fn.ClassName)
	}
}

func TestParseNestedQualifiedMethodUnaffected(t *testing.T) {
	prog := parseGenericProgram(t, `procedure TOuter.TInner.Bar; begin end;`)
	fn := lastFunctionDecl(t, prog)
	if fn.ClassName == nil || fn.ClassName.Value != "TOuter.TInner" {
		t.Errorf("ClassName = %v, want TOuter.TInner", fn.ClassName)
	}
	if fn.ClassTypeParams != nil {
		t.Errorf("ClassTypeParams = %v, want nil", fn.ClassTypeParams)
	}
}

// TestParseComparisonInBodyNotMistakenForMethodTypeParams guards
// looksLikeMethodTypeParams against '<' used as a comparison operator.
func TestParseComparisonInBodyNotMistakenForMethodTypeParams(t *testing.T) {
	prog := parseGenericProgram(t, `procedure Foo; begin if a < b then c := d > e; end;`)
	fn := lastFunctionDecl(t, prog)
	if fn.ClassTypeParams != nil {
		t.Errorf("ClassTypeParams = %v, want nil", fn.ClassTypeParams)
	}
	if fn.ClassName != nil {
		t.Errorf("ClassName = %v, want nil for an unqualified procedure", fn.ClassName)
	}
}

// lastFunctionDecl returns the final top-level function declaration, which is
// the out-of-line implementation in these tests.
func lastFunctionDecl(t *testing.T, prog *ast.Program) *ast.FunctionDecl {
	t.Helper()
	for i := len(prog.Statements) - 1; i >= 0; i-- {
		if fn, ok := prog.Statements[i].(*ast.FunctionDecl); ok {
			return fn
		}
	}
	t.Fatalf("no *ast.FunctionDecl in program; statements: %v", prog.Statements)
	return nil
}

// TestParseGenericInterfaceParent covers `interface (IBase<T>)`: like a class
// inheritance entry, the parent identifier keeps the base name and carries the
// type arguments for the monomorphizer to specialize.
func TestParseGenericInterfaceParent(t *testing.T) {
	prog := parseGenericProgram(t, `type IBase<T> = interface function GetP : T; end;
type IChild<T> = interface(IBase<T>) procedure SetP(v : T); end;`)

	var child *ast.InterfaceDecl
	for _, stmt := range prog.Statements {
		if d, ok := stmt.(*ast.InterfaceDecl); ok && d.Name != nil && d.Name.Value == "IChild" {
			child = d
		}
	}
	if child == nil {
		t.Fatal("no InterfaceDecl named IChild found")
	}
	if child.Parent == nil || child.Parent.Value != "IBase" {
		t.Fatalf("Parent = %v, want IBase (base name, unmangled)", child.Parent)
	}
	if len(child.Parent.TypeArgs) != 1 || child.Parent.TypeArgs[0].String() != "T" {
		t.Fatalf("Parent.TypeArgs = %v, want [T]", child.Parent.TypeArgs)
	}
}

// TestParseMalformedTypeDeclarationDoesNotPanic guards the typed-nil collapse: a
// failed body parse must not reach later phases as a non-nil ast.Statement
// wrapping a nil pointer.
func TestParseMalformedTypeDeclarationDoesNotPanic(t *testing.T) {
	cases := []string{
		`type IChild<T> = interface(123) procedure SetP(v : T); end;`,
		`type IBase<T> = interface function GetP : T; end;
type IChild<T> = interface(IBase<) procedure SetP(v : T); end;`,
		`type TTest<T> = class(999) Field : T; end;`,
	}
	for _, src := range cases {
		p := New(lexer.New(src))
		prog := p.ParseProgram() // must not panic
		if len(p.Errors()) == 0 {
			t.Errorf("expected parser errors for %q", src)
		}
		for i, stmt := range prog.Statements {
			if isNilDecl(stmt) {
				t.Errorf("statement %d of %q is a typed-nil declaration", i, src)
			}
		}
	}
}
