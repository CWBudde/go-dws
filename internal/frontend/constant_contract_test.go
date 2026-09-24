package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_ConstantContractFixtures(t *testing.T) {
	for _, name := range []string{"contracts_warnings", "contracts_error1", "contracts_error3", "contracts_unfinished3", "contracts_types"} {
		t.Run(name, func(t *testing.T) {
			got := Compile(fixtureSource(t, name+".pas"), name+".pas", semantic.HintsLevelPedantic).DiagnosticStrings()
			if want := fixtureExpectation(t, name+".txt"); !reflect.DeepEqual(got, want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
			}
		})
	}
}

func TestCompile_ConstantContractOrder(t *testing.T) {
	source := "procedure Test;\nrequire\n  1 : 2;\nbegin\n  while True do ;\nensure\n  true implies true : 3;\nend;"
	want := []string{
		"Syntax Error: Boolean expected [line: 3, column: 3]",
		"Warning: Constant condition [line: 3, column: 3]",
		"Syntax Error: String expected [line: 3, column: 3]",
		"Warning: Infinite loop [line: 5, column: 3]",
		"Warning: Constant condition [line: 7, column: 3]",
		"Syntax Error: String expected [line: 7, column: 3]",
	}
	if got := Compile(source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings(); !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestCompile_ConstantContractsAcrossRoutines(t *testing.T) {
	for _, source := range []string{
		"procedure Test; require true; begin ensure false; end;",
		"const Flag = true; procedure Test; require Flag; begin ensure 1 = (1+0); end;",
		"type TTest = class procedure Test; require true; begin ensure false; end; end;",
		"type TTest = class procedure Test; end; procedure TTest.Test; require true; begin ensure false; end;",
		"type TTest = record procedure Test; end; procedure TTest.Test; require true; begin ensure false; end;",
		"type TTest = helper for Integer procedure Test; require true; begin ensure false; end; end;",
	} {
		t.Run(source, func(t *testing.T) {
			result := Compile(source, "<test>", semantic.HintsLevelDisabled)
			if len(result.Diagnostics) != 2 {
				t.Fatalf("expected two constant warnings, got %q", result.DiagnosticStrings())
			}
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Severity != SeverityWarning || diagnostic.Message != "Warning: Constant condition" {
					t.Fatalf("unexpected diagnostic: %#v", diagnostic)
				}
			}
		})
	}
}

func TestCompile_NonconstantContracts(t *testing.T) {
	source := "procedure Test(value: Boolean); require value; begin ensure value; end;"
	if got := Compile(source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings(); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %q", got)
	}
}
