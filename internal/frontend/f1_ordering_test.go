package frontend

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_F1FlowAndResultFixtures(t *testing.T) {
	for _, name := range []string{"unreachable", "unreachable_case_of", "raise_error", "raise_syntax", "result_redefine"} {
		t.Run(name, func(t *testing.T) {
			got := Compile(fixtureSource(t, name+".pas"), name+".pas", semantic.HintsLevelPedantic).DiagnosticStrings()
			if want := fixtureExpectation(t, name+".txt"); !reflect.DeepEqual(got, want) {
				t.Fatalf("got %q\nwant %q", got, want)
			}
		})
	}
}

func TestCompile_EarlierRoutineErrorSuppressesMethodSelfAssignment(t *testing.T) {
	source := "procedure Before; begin Missing := 1; end;\n" +
		"type TTest = class procedure M; begin var x := 1; x := x; end; end;"
	result := Compile(source, "<test>", semantic.HintsLevelPedantic)
	if !result.HasFatalDiagnostics() {
		t.Fatal("expected undefined variable error")
	}
	for _, diagnostic := range result.DiagnosticStrings() {
		if strings.Contains(diagnostic, "Assigning x to itself") {
			t.Fatalf("hint follows earlier routine error: %q", result.DiagnosticStrings())
		}
	}
}

func TestCompile_InlineMethodsKeepDeclarationOrder(t *testing.T) {
	for _, middle := range []string{
		"while True do ;",
		"procedure Between; begin while True do ; end;",
	} {
		t.Run(middle, func(t *testing.T) {
			source := "type A = class\nprocedure M; begin\nwhile True do ;\nend;\nend;\n" + middle +
				"\ntype B = class\nprocedure M; begin\nwhile True do ;\nend;\nend;"
			want := []string{
				"Warning: Infinite loop [line: 3, column: 1]",
				"Warning: Infinite loop [line: 6, column: 1]",
				"Warning: Infinite loop [line: 9, column: 1]",
			}
			if middle[0] == 'p' {
				want[1] = "Warning: Infinite loop [line: 6, column: 26]"
			}
			got := Compile(source, "<test>", semantic.HintsLevelPedantic).DiagnosticStrings()
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %q\nwant %q", got, want)
			}
		})
	}
}

func TestCompile_InlineMethodsRespectEarlierErrors(t *testing.T) {
	source := "type A = class\n procedure M; begin\n  var x := 1; x := x;\n end;\nend;\nMissing := 1;\ntype B = class\n procedure M; begin\n  var y := 1; y := y;\n end;\nend;"
	want := []string{
		"Hint: Assigning x to itself [line: 3, column: 17]",
		"Syntax Error: Undefined variable 'Missing' [line: 6, column: 9]",
	}
	if got := Compile(source, "<test>", semantic.HintsLevelNormal).DiagnosticStrings(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestCompile_DeferredOverrideErrorPrecedesBody(t *testing.T) {
	for _, body := range []string{"while True do ;", "var x := 1; x := x;"} {
		t.Run(body, func(t *testing.T) {
			source := "type TChild = class(TBase)\n procedure M; override;\n begin\n  " + body +
				"\n end;\nend;\ntype TBase = class\n procedure M; begin end;\nend;"
			want := []string{"method 'M' marked as override, but parent method is not virtual"}
			if body == "while True do ;" {
				want = append(want, "Warning: Infinite loop [line: 4, column: 3]")
			}
			if got := Compile(source, "<test>", semantic.HintsLevelNormal).DiagnosticStrings(); !reflect.DeepEqual(got, want) {
				t.Fatalf("got %q\nwant %q", got, want)
			}
		})
	}
}
