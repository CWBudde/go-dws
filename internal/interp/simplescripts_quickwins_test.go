package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// runQuickwinScript compiles and runs a DWScript source string and returns its
// combined program output. It fails the test on any compile or runtime error.
func runQuickwinScript(t *testing.T, source string) string {
	t.Helper()

	compiled := frontend.Compile(source, "quickwins.pas", semantic.HintsLevelNormal)
	if compiled.HasFatalDiagnostics() || !compiled.SemanticSuccessful {
		t.Fatalf("compile diagnostics:\n%s", strings.Join(compiled.DiagnosticStrings(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if compiled.SemanticInfo != nil {
		interp.SetSemanticInfo(compiled.SemanticInfo)
	}
	result := interp.Eval(compiled.Program)
	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("runtime error: %s", result.String())
	}
	return buf.String()
}

func assertOutput(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%q\ngot:\n%q", want, got)
	}
}

// assertCompileError asserts that the source fails semantic analysis with a
// diagnostic containing want.
func assertCompileError(t *testing.T, source, want string) {
	t.Helper()
	compiled := frontend.Compile(source, "quickwins.pas", semantic.HintsLevelNormal)
	if !compiled.HasFatalDiagnostics() && compiled.SemanticSuccessful {
		t.Fatalf("expected a compile error containing %q, but compilation succeeded", want)
	}
	diags := strings.Join(compiled.DiagnosticStrings(), "\n")
	if !strings.Contains(diags, want) {
		t.Fatalf("expected a diagnostic containing %q, got:\n%s", want, diags)
	}
}

// TestImpliesRejectsNonBooleanOperand covers the tightened semantic rule: a
// non-Boolean, non-Variant operand is rejected even when the other side is Variant.
func TestImpliesRejectsNonBooleanOperand(t *testing.T) {
	assertCompileError(t, `
var s := 'x';
var v : Variant;
PrintLn(s implies v);
`, "operator 'implies'")
}

// TestEmptyMethodOutOfLineBodyIsDuplicate covers that an "empty;" method is a
// complete definition, so a later out-of-line body is a duplicate, not a forward
// implementation.
func TestEmptyMethodOutOfLineBodyIsDuplicate(t *testing.T) {
	assertCompileError(t, `
type
   TTest = class
      procedure Proc; empty;
   end;
procedure TTest.Proc;
begin
end;
`, "duplicate method")
}

// TestResourcestringRejectsNonStringValue covers that a resourcestring must have
// a String value.
func TestResourcestringRejectsNonStringValue(t *testing.T) {
	assertCompileError(t, `
resourcestring d = 1234;
`, "String expected")
}

// TestResourcestringSectionMarksAllDeclarations covers that a `resourcestring`
// section marks every declaration, not only the first: the string-only rule must
// still reject a non-string value in a later entry of the same section.
func TestResourcestringSectionMarksAllDeclarations(t *testing.T) {
	assertCompileError(t, `
resourcestring
   a = 'first';
   b = 1234;
`, "String expected")
}

// TestImpliesOperator covers the logical implication operator, including its
// short-circuit behaviour (a False antecedent must not evaluate the consequent).
func TestImpliesOperator(t *testing.T) {
	got := runQuickwinScript(t, `
var t = True;
var f = not t;
PrintLn(t implies t);
PrintLn(t implies f);
PrintLn(f implies t);
PrintLn(f implies f);
`)
	assertOutput(t, got, "True\nFalse\nTrue\nTrue\n")
}

func TestImpliesShortCircuits(t *testing.T) {
	// A False antecedent must NOT evaluate the consequent: the div-by-zero on the
	// right would raise if it were evaluated, so reaching 'ok' proves short-circuit.
	got := runQuickwinScript(t, `
var f := False;
var z := 0;
if (f implies ((1 div z) = 0)) then
   PrintLn('ok');
`)
	assertOutput(t, got, "ok\n")
}

// TestEmptyMethodDirective covers the "empty;" directive: a body-less routine
// that is a no-op returning its result type's zero value.
func TestEmptyMethodDirective(t *testing.T) {
	got := runQuickwinScript(t, `
type
   TTest = class
      function Test : Integer; empty;
      function Name : String; empty;
      procedure Proc; empty;
   end;
var o := new TTest;
PrintLn(o.Test);
PrintLn('[' + o.Name + ']');
o.Proc;
PrintLn('done');
`)
	assertOutput(t, got, "0\n[]\ndone\n")
}

// TestInlineDirective covers the advisory "inline;" directive on a function.
func TestInlineDirective(t *testing.T) {
	got := runQuickwinScript(t, `
function Next(i : Integer) : Integer; inline;
begin
   Result := i + 1;
end;
PrintLn(Next(1));
`)
	assertOutput(t, got, "2\n")
}

// TestEmptyAndInlineAsIdentifiers verifies "empty" and "inline" remain usable as
// ordinary identifiers outside routine-directive position.
func TestEmptyAndInlineAsIdentifiers(t *testing.T) {
	got := runQuickwinScript(t, `
const empty = '';
var inline := 5;
PrintLn(Length(empty));
PrintLn(inline);
`)
	assertOutput(t, got, "0\n5\n")
}

// TestResourcestringDeclaration covers resourcestring declarations behaving as
// string constants.
func TestResourcestringDeclaration(t *testing.T) {
	got := runQuickwinScript(t, `
resourcestring hello = 'hi';
resourcestring bye = 'bye';
PrintLn(hello);
var v := bye;
PrintLn(v);
`)
	assertOutput(t, got, "hi\nbye\n")
}

// TestEscapedReservedWordIdentifiers covers &keyword escaped identifiers.
func TestEscapedReservedWordIdentifiers(t *testing.T) {
	got := runQuickwinScript(t, `
var &begin : Integer;
&begin := 2;
procedure &shl(&then : String);
begin
   PrintLn(&then + IntToStr(&begin));
end;
&shl('x');
`)
	assertOutput(t, got, "x2\n")
}

// TestNaNAndInfinityConstants covers the predefined NaN/Infinity float constants
// and their DWScript-style rendering (NAN, INF).
func TestNaNAndInfinityConstants(t *testing.T) {
	got := runQuickwinScript(t, `
var x : Float := NaN;
var y : Float := Infinity;
PrintLn(x);
PrintLn(y);
PrintLn(x = NaN);
`)
	assertOutput(t, got, "NAN\nINF\nFalse\n")
}

// --- SimpleScripts quick-wins batch #2 ---

// TestModOnFloats covers `mod` on Float operands (Delphi fmod semantics),
// including negative-zero normalization to "0".
func TestModOnFloats(t *testing.T) {
	got := runQuickwinScript(t, `
PrintLn((-2.5) mod 0.5);
PrintLn(5.5 mod 2.0);
var f := 1e14;
PrintLn(f mod 1);
`)
	assertOutput(t, got, "0\n1.5\n0\n")
}

// TestFloatDivByZeroIsInfNan covers that float division by zero yields ±Inf/NaN
// (rendered INF/NAN) and does not abort execution.
func TestFloatDivByZeroIsInfNan(t *testing.T) {
	got := runQuickwinScript(t, `
var a := 0.0;
PrintLn(1 / a);
PrintLn(IsNaN(a / a));
`)
	assertOutput(t, got, "INF\nTrue\n")
}

// TestFloatToStrAcceptsAliasedInteger covers FloatToStr on a type-aliased Integer
// argument and casting through the alias name.
func TestFloatToStrAcceptsAliasedInteger(t *testing.T) {
	got := runQuickwinScript(t, `
type TMyInt = Integer;
var i : TMyInt = 2;
PrintLn(FloatToStr(i, 1));
PrintLn(FloatToStr(TMyInt(3), 2));
`)
	assertOutput(t, got, "2.0\n3.00\n")
}

// TestReservedWordPropertyAccessor covers a reserved word (Set) used as a
// property write accessor name.
func TestReservedWordPropertyAccessor(t *testing.T) {
	got := runQuickwinScript(t, `
type TTest = class
	function Get(i : Integer) : String; begin Result := IntToStr(i); end;
	procedure Set(i : Integer; v : String); begin PrintLn('set ' + v); end;
	property Prop[i : Integer] : String read Get write Set;
end;
var t := TTest.Create;
t.Prop[1] := 'x';
PrintLn(t.Prop[7]);
`)
	assertOutput(t, got, "set x\n7\n")
}

// TestArrayOfPropertyType covers a property whose type is a composite `array of T`.
func TestArrayOfPropertyType(t *testing.T) {
	got := runQuickwinScript(t, `
type TTest = class
	private FNames : array of String;
	public property Names : array of String read FNames write FNames;
end;
var a : array of String;
var t := TTest.Create;
t.Names := a;
a.Add('hi');
PrintLn(t.Names[0]);
`)
	assertOutput(t, got, "hi\n")
}

// TestContractExpressionMessage covers a contract condition whose failure message
// is an arbitrary expression, not just a string literal.
func TestContractExpressionMessage(t *testing.T) {
	got := runQuickwinScript(t, `
function Foo(i : Integer) : Integer;
require
	i > 0 : 'must be positive, was ' + IntToStr(i);
begin
	Result := i;
end;
PrintLn(Foo(5));
`)
	assertOutput(t, got, "5\n")
}

// TestPropertyPromotion covers `property Prop;` redeclaring an inherited property
// under a wider visibility.
func TestPropertyPromotion(t *testing.T) {
	got := runQuickwinScript(t, `
type TBase = class
	protected
		Field : Integer;
		property Prop : Integer read Field;
end;
type TSub = class (TBase)
	public
		property Prop;
end;
PrintLn(TSub.Create.Prop);
`)
	assertOutput(t, got, "0\n")
}

// TestDefaultNamespaceCall covers `Default.<name>(args)` resolving against the
// global scope (bypassing a same-named class method).
func TestDefaultNamespaceCall(t *testing.T) {
	got := runQuickwinScript(t, `
type TStatic = class
	class procedure Print; static;
	begin
		Default.PrintLn('hi');
	end;
end;
TStatic.Print;
`)
	assertOutput(t, got, "hi\n")
}
