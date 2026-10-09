package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_PropertyIndexDeclarationFixtures(t *testing.T) {
	for _, name := range []string{"array_params1", "array_params2", "array_params3"} {
		t.Run(name, func(t *testing.T) {
			result := CompileWithOptions(fixtureSource(t, name+".pas"), Options{HintsLevel: semantic.HintsLevelPedantic})
			want := fixtureExpectation(t, name+".txt")
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
				t.Fatalf("got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
			}
		})
	}
}

// CheckParams' mode matrix, after count/result/type checks; property spelling
// owns the detail, even when accessor parameter spelling is different.
func TestCompile_PropertyIndexModes_AccessorMatrix(t *testing.T) {
	modes := []struct{ name, prefix string }{{"value", ""}, {"var", "var "}, {"const", "const "}}
	details := [][]string{
		{"", "Value-parameter expected", "Value-parameter expected"},
		{"Var-parameter expected", "", "Var-parameter expected"},
		{"Const-parameter expected", "Value-parameter expected", ""},
	}
	for pi, prop := range modes {
		for ai, accessor := range modes {
			for _, write := range []bool{false, true} {
				kind, header, clause := "reader", "function GeT("+accessor.prefix+"AccessorName: Integer): Integer;", "read get"
				declared := "GeT"
				if write {
					kind, header, clause, declared = "writer", "procedure PuT("+accessor.prefix+"AccessorName: Integer; V: Integer);", "write put", "PuT"
				}
				t.Run(kind+"/"+prop.name+"/"+accessor.name, func(t *testing.T) {
					source := "type T = class\n " + header + "\n property P[" + prop.prefix + "PropertyName: Integer]: Integer " + clause[:strings.LastIndex(clause, " ")] + "\n {padding}\n " + clause[strings.LastIndex(clause, " ")+1:] + ";\n property Stop: Integer read ;\nend; Missing; {$ERROR 'late'}"
					var want []string
					if detail := details[pi][ai]; detail != "" {
						want = append(want, "Syntax Error: Parameter 0 (PropertyName) - "+detail+" [line: 5, column: 2]", `Syntax Error: Method "`+declared+`" has incompatible parameters [line: 5, column: 2]`)
					}
					want = append(want, `Syntax Error: Name expected [line: 6, column: 30]`)
					assertDiagnostics(t, source, "<test>", want)
				})
			}
		}
	}
}

func TestCompile_PropertyIndexModes_Priority(t *testing.T) {
	for _, tt := range []struct {
		name, member, property string
		messages               []string
	}{
		{"types suppress each mode", "function GeT(A: String; B: Integer): Integer;", "P[var First: Integer; const Second: Integer]: Integer read get", []string{`Parameter 0 - Type "Integer" expected (instead of "String")`, `Parameter 1 (Second) - Const-parameter expected`, `Method "GeT" has incompatible parameters`}},
		{"count before modes", "function GeT: Integer;", "P[var I: Integer]: Integer read get", []string{`Method "GeT" has incompatible parameters`}},
		{"result before modes", "function GeT(A: Integer): String;", "P[var I: Integer]: Integer read get", []string{`Field/method "GeT" has an incompatible type`}},
		{"writer count before modes", "procedure PuT(A: Integer);", "P[var I: Integer]: Integer write put", []string{`Method "PuT" has incompatible parameters`}},
		{"writer value before modes", "procedure PuT(A: Integer; V: String);", "P[var I: Integer]: Integer write put", []string{`Method "PuT" has incompatible parameters`}},
		{"writer kind before modes", "function PuT(A: Integer; V: Integer): Integer;", "P[var I: Integer]: Integer write put", []string{`Procedure expected`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			split := strings.LastIndex(tt.property, " ")
			source := "type T = class\n " + tt.member + "\n property " + tt.property[:split] + "\n {padding}\n " + tt.property[split+1:] + ";\n property Stop: Integer read ;\nend;"
			want := make([]string, 0, len(tt.messages)+1)
			for _, message := range tt.messages {
				want = append(want, "Syntax Error: "+message+" [line: 5, column: 2]")
			}
			want = append(want, `Syntax Error: Name expected [line: 6, column: 30]`)
			assertDiagnostics(t, source, "<test>", want)
		})
	}
}

func TestCompile_PropertyIndexModes_EmptyErrorDoesNotBlockAnalysis(t *testing.T) {
	source := "type T = class\n Field: Integer;\n function Get(A: Integer): Integer; begin Result := A; end;\n property Dummy[]: Integer read Field;\n property Bad[const Spelling: Integer]: Integer read\n Get;\nend;"
	assertDiagnostics(t, source, "<test>", []string{
		`Syntax Error: Parameters expected [line: 4, column: 17]`,
		`Syntax Error: Parameter 0 (Spelling) - Const-parameter expected [line: 6, column: 2]`,
		`Syntax Error: Method "Get" has incompatible parameters [line: 6, column: 2]`,
	})
}

func TestCompile_PropertyIndexModes_InheritedSelectedAccessor(t *testing.T) {
	source := "type TBase = class\n function GeT(AccessorName: Integer): Integer; begin Result := AccessorName; end;\nend;\ntype TChild = class(TBase)\n property P[var PropertyName: Integer]: Integer read\n get;\nend;"
	assertDiagnostics(t, source, "<test>", []string{
		`Syntax Error: Parameter 0 (PropertyName) - Var-parameter expected [line: 6, column: 2]`,
		`Syntax Error: Method "GeT" has incompatible parameters [line: 6, column: 2]`,
	})
}

func TestCompile_PropertyIndexModes_InterfaceAccessorMatrix(t *testing.T) {
	modes := []string{"", "var ", "const "}
	details := [][]string{{"", "Value-parameter expected", "Value-parameter expected"}, {"Var-parameter expected", "", "Var-parameter expected"}, {"Const-parameter expected", "Value-parameter expected", ""}}
	for pi, prop := range modes {
		for ai, accessor := range modes {
			t.Run(modes[pi]+"/"+modes[ai], func(t *testing.T) {
				source := "type ITest = interface\n function Get(" + accessor + "Different: Integer): Integer;\n procedure Put(" + accessor + "Other: Integer; V: Integer);\n property P[" + prop + "IndexName: Integer]: Integer read\n Get write\n Put;\nend;"
				var want []string
				if detail := details[pi][ai]; detail != "" {
					want = []string{
						"Syntax Error: Parameter 0 (IndexName) - " + detail + " [line: 5, column: 2]",
						`Syntax Error: Method "Get" has incompatible parameters [line: 5, column: 2]`,
						"Syntax Error: Parameter 0 (IndexName) - " + detail + " [line: 6, column: 2]",
						`Syntax Error: Method "Put" has incompatible parameters [line: 6, column: 2]`,
					}
				}
				assertDiagnostics(t, source, "<test>", want)
			})
		}
	}
}

func TestCompile_PropertyIndexModes_MultipleDetails(t *testing.T) {
	source := "type T = class\n function Get(B, A: Integer): Integer;\n procedure Put(B, A, V: Integer);\n property P[var First: Integer; const Second: Integer]: Integer read\n Get write\n Put;\n property Stop: Integer read ;\nend;"
	assertDiagnostics(t, source, "<test>", []string{
		`Syntax Error: Parameter 0 (First) - Var-parameter expected [line: 5, column: 2]`,
		`Syntax Error: Parameter 1 (Second) - Const-parameter expected [line: 5, column: 2]`,
		`Syntax Error: Method "Get" has incompatible parameters [line: 5, column: 2]`,
		`Syntax Error: Parameter 0 (First) - Var-parameter expected [line: 6, column: 2]`,
		`Syntax Error: Parameter 1 (Second) - Const-parameter expected [line: 6, column: 2]`,
		`Syntax Error: Method "Put" has incompatible parameters [line: 6, column: 2]`,
		`Syntax Error: Name expected [line: 7, column: 30]`,
	})
}

func TestCompile_PropertyIndexModes_InterfacePriority(t *testing.T) {
	for _, tt := range []struct {
		name, member, property string
		messages               []string
	}{
		{"type before mode", "function Get(A: String; B: Integer): Integer;", "P[var First: Integer; const Second: Integer]: Integer read Get", []string{`Parameter 0 - Type "Integer" expected (instead of "String")`, `Parameter 1 (Second) - Const-parameter expected`, `Method "Get" has incompatible parameters`}},
		{"count before modes", "function Get: Integer;", "P[var I: Integer]: Integer read Get", []string{`Method "Get" has incompatible parameters`}},
		{"result before modes", "function Get(A: Integer): String;", "P[var I: Integer]: Integer read Get", []string{`Field/method "Get" has an incompatible type`}},
		{"writer value before modes", "procedure Put(A: Integer; V: String);", "P[var I: Integer]: Integer write Put", []string{`Method "Put" has incompatible parameters`}},
		{"writer kind before modes", "function Put(A: Integer; V: Integer): Integer;", "P[var I: Integer]: Integer write Put", []string{`Procedure expected`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			split := strings.LastIndex(tt.property, " ")
			source := "type ITest = interface\n " + tt.member + "\n property " + tt.property[:split] + "\n {padding}\n " + tt.property[split+1:] + ";\nend;"
			want := make([]string, 0, len(tt.messages))
			for _, message := range tt.messages {
				want = append(want, "Syntax Error: "+message+" [line: 5, column: 2]")
			}
			assertDiagnostics(t, source, "<test>", want)
		})
	}
}
