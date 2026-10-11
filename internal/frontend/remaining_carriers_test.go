package frontend

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestSortDiagnostics_ReachedSemanticBeforeParserStop(t *testing.T) {
	mainFile, unitFile := "", ""
	semanticError := Diagnostic{Message: "reached check", Phase: PhaseSemantic, Line: 1, Column: 29, Severity: SeverityError}
	parserStop := Diagnostic{Message: "closer expected", Phase: PhaseParsing, Line: 1, Column: 29, Severity: SeverityError, Stop: true}
	for _, tt := range []struct {
		name       string
		sameSource bool
	}{
		{"same source", true},
		{"different sources with empty filenames", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			left, right := parserStop, semanticError
			left.sourceFile = &mainFile
			right.sourceFile = &mainFile
			want := []string{"reached check", "closer expected"}
			if !tt.sameSource {
				right.sourceFile = &unitFile
				want = []string{"closer expected", "reached check"}
			}
			diags := []Diagnostic{left, right}
			sortDiagnostics(diags)
			got := []string{diags[0].Message, diags[1].Message}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("order = %v, want %v", got, want)
			}
		})
	}
}

func TestSortDiagnostics_ReachedSemanticBeforeParserStopMixedSources(t *testing.T) {
	mainFile, unitFile := "", ""
	mainSemantic := Diagnostic{Message: "main reached check", Phase: PhaseSemantic, Line: 1, Column: 29, Severity: SeverityError, sourceFile: &mainFile}
	mainStop := Diagnostic{Message: "main closer", Phase: PhaseParsing, Line: 1, Column: 29, Severity: SeverityError, Stop: true, sourceFile: &mainFile}
	unitSemantic := Diagnostic{Message: "unit error", Phase: PhaseSemantic, Line: 1, Column: 29, Severity: SeverityError, sourceFile: &unitFile}
	for _, tt := range []struct {
		name  string
		diags []Diagnostic
	}{
		{"stop unit main", []Diagnostic{mainStop, unitSemantic, mainSemantic}},
		{"stop main unit", []Diagnostic{mainStop, mainSemantic, unitSemantic}},
		{"main stop unit", []Diagnostic{mainSemantic, mainStop, unitSemantic}},
		{"main unit stop", []Diagnostic{mainSemantic, unitSemantic, mainStop}},
		{"unit stop main", []Diagnostic{unitSemantic, mainStop, mainSemantic}},
		{"unit main stop", []Diagnostic{unitSemantic, mainSemantic, mainStop}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var wantSemanticOrder []string
			for _, diag := range tt.diags {
				if diag.Phase == PhaseSemantic {
					wantSemanticOrder = append(wantSemanticOrder, diag.Message)
				}
			}
			// Compile sorts both source-local and merged diagnostics. Repeating
			// the sort must preserve the restoration and semantic emission order.
			for pass := 0; pass < 2; pass++ {
				sortDiagnostics(tt.diags)
				var gotSemanticOrder []string
				semanticIndex, stopIndex := -1, -1
				for i, diag := range tt.diags {
					if diag.Phase == PhaseSemantic {
						gotSemanticOrder = append(gotSemanticOrder, diag.Message)
					}
					if diag.Message == mainSemantic.Message {
						semanticIndex = i
					}
					if diag.Stop {
						stopIndex = i
					}
				}
				if semanticIndex >= stopIndex {
					t.Errorf("pass %d: same-source semantic check remained after parser stop: %v", pass, tt.diags)
				}
				if !reflect.DeepEqual(gotSemanticOrder, wantSemanticOrder) {
					t.Errorf("pass %d: semantic order = %v, want %v", pass, gotSemanticOrder, wantSemanticOrder)
				}
			}
		})
	}
}

func TestCompile_RemainingCarrierFixtures(t *testing.T) {
	for _, name := range []string{"try_except1", "class_error4", "class_error2", "class_error3"} {
		t.Run(name, func(t *testing.T) {
			result := Compile(fixtureSource(t, name+".pas"), name+".pas", semantic.HintsLevelPedantic)
			want := fixtureExpectation(t, name+".txt")
			if got := result.DiagnosticStrings(); strings.Join(got, "\n") != strings.Join(want, "\n") {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
			}
			for i, diag := range result.Diagnostics {
				if diag.Stop != (i == len(result.Diagnostics)-1) {
					t.Errorf("diagnostic %d Stop = %v, want only the final diagnostic to stop", i, diag.Stop)
				}
			}
		})
	}
}

func TestCompile_RemainingCarrierControls(t *testing.T) {
	tests := []struct {
		name, source string
		want         []string
	}{
		{"alias handler", "type TAlias = Exception;\ntry\nexcept\non e: TAlias do ;\nend;", nil},
		{"class handler", "type TPlain = class end;\ntry\nexcept\non e: TPlain do ;\nend;", nil},
		{"ordinary handler error keeps reached statements", "try\nexcept\non e: Integer do\nPrintLn(1 as String);\non e: Exception do\nPrintLn(2 as String);\nend;", []string{
			`Syntax Error: Class reference expected [line: 3, column: 7]`,
			`Syntax Error: Cannot cast "Integer" as "String" [line: 4, column: 11]`,
			`Syntax Error: Cannot cast "Integer" as "String" [line: 6, column: 11]`,
		}},
		{"try body stop", "try\nPrintLn(1 + );\nexcept\non e: Unread do ;\nend;\nPrintLn(UnreadTail);", []string{
			`Syntax Error: Expression expected [line: 2, column: 13]`,
		}},
		{"handler body stop", "try\nexcept\non e: Integer do PrintLn(1 + );\non e: Unread do ;\nend;\nPrintLn(UnreadTail);", []string{
			`Syntax Error: Class reference expected [line: 3, column: 7]`,
			`Syntax Error: Expression expected [line: 3, column: 30]`,
		}},
		{"handler missing do keeps reached type", "try\nexcept\non e: Integer ;\nend;\nPrintLn(UnreadTail);", []string{
			`Syntax Error: Class reference expected [line: 3, column: 7]`,
			`Syntax Error: DO expected [line: 3, column: 15]`,
		}},
		{"class header stop skips unread body", "type TTest = class(TObject, Integer\nprivate\nF: Unread;\nprocedure P;\nend;\nPrintLn(UnreadTail);", []string{
			`Syntax Error: "Integer" is not an interface [line: 1, column: 29]`,
			`Syntax Error: ")" expected [line: 1, column: 29]`,
		}},
		{"class header stop skips interface implementation", "type ITest = interface procedure P; end;\ntype TTest = class(TObject, ITest", []string{
			`Syntax Error: ")" expected [line: 2, column: 29]`,
		}},
		{"complete invalid ancestry", "type TTest = class(TObject, Integer) end;", []string{
			`Syntax Error: "Integer" is not an interface [line: 1, column: 29]`,
		}},
		{"complete ancestry check precedes body stop", "type TTest = class(TObject, Integer)\nF: Integer := Missing;\nend;", []string{
			`Syntax Error: "Integer" is not an interface [line: 1, column: 29]`,
			`Syntax Error: Unknown name "Missing" [line: 2, column: 15]`,
		}},
		{"complete valid ancestry", "type ITest = interface end;\ntype TTest = class(TObject, ITest) end;", nil},
		{"truncated routine with try skips completion hints", "procedure P;\nbegin\ntry\nvar unused: Integer;\nexcept\non e: Exception do ;", []string{
			`Syntax Error: END expected [line: 6, column: 20]`,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertDiagnostics(t, tt.source, "<test>", tt.want) })
	}
}
