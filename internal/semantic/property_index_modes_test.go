package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
)

func TestPropertyIndexModes_Metadata(t *testing.T) {
	input := `type T = class
 function Get(var A, B: Integer; const C: Integer; D: Integer): Integer;
 begin Result := A + B + C + D; end;
 property P[var A, B: Integer; const C: Integer; D: Integer]: Integer read Get;
 property Q[var A, B: Integer; const C: Integer; D: Integer]: Integer read P;
end;
type U = class(T) property P; end;`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatal(p.Errors())
	}
	a := NewAnalyzer()
	if err := a.Analyze(program); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"t", "u"} {
		for _, property := range []string{"P", "Q"} {
			prop, found := a.GetClasses()[name].GetProperty(property)
			if !found {
				t.Fatalf("%s.%s missing", name, property)
			}
			want := []types.PropertyIndexMode{types.PropertyIndexVar, types.PropertyIndexVar, types.PropertyIndexConst, types.PropertyIndexValue}
			if len(prop.IndexParamModes) != len(want) {
				t.Fatalf("mode count = %d, want %d", len(prop.IndexParamModes), len(want))
			}
			for i, mode := range want {
				if prop.IndexMode(i) != mode {
					t.Errorf("%s.%s mode %d = %d, want %d", name, property, i, prop.IndexMode(i), mode)
				}
			}
		}
	}
}

func TestPropertyIndexModes_ForwardingMismatch(t *testing.T) {
	for _, mode := range []string{"", "const "} {
		input := `type T = class
 function Get(var I: Integer): Integer; begin Result := I; end;
 procedure SetIt(var I: Integer; V: Integer); begin I := V; end;
 property P[var I: Integer]: Integer read Get write SetIt;
 property Q[` + mode + `I: Integer]: Integer read P write P;
end;`
		p := parser.New(lexer.New(input))
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			t.Fatal(p.Errors())
		}
		a := NewAnalyzer()
		if err := a.Analyze(program); err == nil {
			t.Fatal("forwarding mismatched property index modes must reject")
		}
	}
}

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
