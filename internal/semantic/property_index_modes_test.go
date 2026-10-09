package semantic

import (
	"strings"
	"testing"
)

func TestPropertyIndexModes_InterfaceAccessorMatrix(t *testing.T) {
	modes := []string{"", "var ", "const "}
	details := [][]string{{"", "Value-parameter expected", "Value-parameter expected"}, {"Var-parameter expected", "", "Var-parameter expected"}, {"Const-parameter expected", "Value-parameter expected", ""}}
	for pi, prop := range modes {
		for ai, accessor := range modes {
			t.Run(modes[pi]+"/"+modes[ai], func(t *testing.T) {
				source := "type ITest = interface\n function Get(" + accessor + "Different: Integer): Integer;\n procedure Put(" + accessor + "Other: Integer; V: Integer);\n property P[" + prop + "IndexName: Integer]: Integer read\n Get write\n Put;\nend;"
				analyzer, _ := analyzeSource(t, source)
				var want []string
				if detail := details[pi][ai]; detail != "" {
					want = []string{
						"Parameter 0 (IndexName) - " + detail + " at 5:2",
						`Syntax Error: Method "Get" has incompatible parameters [line: 5, column: 2]`,
						"Parameter 0 (IndexName) - " + detail + " at 6:2",
						`Syntax Error: Method "Put" has incompatible parameters [line: 6, column: 2]`,
					}
				}
				if got := strings.Join(analyzer.Errors(), "\n"); got != strings.Join(want, "\n") {
					t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
				}
			})
		}
	}
}

func TestPropertyIndexModes_MatchingClassInterface(t *testing.T) {
	for _, mode := range []string{"", "var ", "const "} {
		t.Run(mode, func(t *testing.T) {
			source := "type ITest = interface function Get(" + mode + "A: Integer): Integer; property P[" + mode + "I: Integer]: Integer read Get; end;\ntype T = class(ITest) function Get(" + mode + "A: Integer): Integer; begin Result := A; end; property P[" + mode + "I: Integer]: Integer read Get; end;"
			expectNoErrors(t, source)
		})
	}
}
