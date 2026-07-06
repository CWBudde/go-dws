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
