package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func runForLoopRegressionScript(t *testing.T, source string) string {
	t.Helper()

	compiled := frontend.Compile(source, "for_loop_regression.pas", semantic.HintsLevelPedantic)
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

func TestForInStepSkipsArrayElements(t *testing.T) {
	got := runForLoopRegressionScript(t, `
const values : array [3..6] of Integer = [3, 5, 7, 9];
var i : Integer;
for i in values step 2 do
	PrintLn(i);
`)
	want := "3\n7\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestForStepRuntimeExceptionIsCatchable(t *testing.T) {
	got := runForLoopRegressionScript(t, `
var i : Integer;
try
	var s := -1;
	for i:=1 to 5 step s do
		PrintLn(i);
except
	on E: Exception do PrintLn(E.Message);
end;
`)
	want := "FOR loop STEP should be strictly positive: -1 [line: 5, column: 2]\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestForInStringUsesExistingLoopVariableType(t *testing.T) {
	got := runForLoopRegressionScript(t, `
var c : Integer;
for c in 'Az' do
	PrintLn(IntToStr(c));

var s := '';
for var ch in 'Az' do
	PrintLn(ch);
`)
	want := "65\n122\nA\nz\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestForVarLoopVariableVisibleToEndExpression(t *testing.T) {
	got := runForLoopRegressionScript(t, `
for var i:=10 downto i-2 do
	PrintLn(i);
`)
	want := "10\n9\n8\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestForFixturesArrayAndConstructorContexts(t *testing.T) {
	got := runForLoopRegressionScript(t, `
type
	TRec = record
		num: Integer;
	end;

var records: array of TRec = ((num: 1), (num: 3));
for var rec in records do
	PrintLn(rec.num);

type TBase = class end;
type TChild = class(TBase) end;

var children : array of TChild;
children.Add(TChild.Create);
for var child in children do
	PrintLn(child.ClassName);

var words : array of String;
words.Add('one', 'two');
for var word in words do
	PrintLn(word);
`)
	want := "1\n3\nTChild\none\ntwo\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}
