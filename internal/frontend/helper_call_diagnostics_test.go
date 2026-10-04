package frontend

import (
	"slices"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_HelperReceiverCallDiagnostics(t *testing.T) {
	for _, target := range []struct{ name, declaration string }{
		{"primitive", "Integer"},
		{"record", "record value: Integer; end"},
		{"class", "class end"},
		{"interface", "interface end"},
	} {
		t.Run(target.name, func(t *testing.T) {
			for _, tt := range []struct {
				name, member, call string
				want               []string
			}{
				{"short", "procedure Take(v: Integer);", "item.Take();", []string{`Syntax Error: More arguments expected [line: 4, column: 6]`}},
				{"excess", "procedure Take(v: Integer);", "item.Take(1, 2);", []string{`Syntax Error: Too many arguments [line: 4, column: 6]`}},
				{"parameterless excess", "procedure Take;", "item.Take(1);", []string{`Syntax Error: Too many arguments [line: 4, column: 6]`}},
				{"type before short", "procedure Take(v: Integer; w: String);", "item.Take(true);", []string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 6]`}},
				{"type before excess", "procedure Take(v: Integer);", "item.Take(true, 1);", []string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 17]`}},
				{"last type anchor", "procedure Take(v: Integer; w: String);", "item.Take(1, true);", []string{`Syntax Error: Argument 2 expects type "String" instead of "Boolean" [line: 4, column: 6]`}},
				{"multiline shift", "procedure Take(v: Integer);", "item.Take(true,\n  1);", []string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 5, column: 3]`}},
				{"child before short", "procedure Take(v: Integer; w: String);", "item.Take(\n  Missing);", []string{`Syntax Error: Unknown name "Missing" [line: 5, column: 3]`, `Syntax Error: More arguments expected [line: 4, column: 6]`}},
				{"excess children", "procedure Take;", "item.Take(\n  Missing, Other);", []string{`Syntax Error: Unknown name "Missing" [line: 5, column: 3]`, `Syntax Error: Unknown name "Other" [line: 5, column: 12]`, `Syntax Error: Too many arguments [line: 4, column: 6]`}},
				{"case insensitive", "procedure Take(v: Integer);", "item.tAKE();", []string{`Syntax Error: More arguments expected [line: 4, column: 6]`}},
				{"valid", "procedure Take(v: Integer);", "item.Take(1);", nil},
				{"valid parameterless", "procedure Take;", "item.Take();", nil},
				{"contextual empty array", "procedure Take(v: array of Integer);", "item.Take([]);", nil},
				{"contextual nil", "procedure Take(v: TObject);", "item.Take(nil);", nil},
			} {
				t.Run(tt.name, func(t *testing.T) {
					source := "type T = " + target.declaration + ";\ntype H = helper for T " + tt.member + " begin end; end;\nvar item: T;\n" + tt.call
					assertHelperCallDiagnostics(t, source, tt.want)
				})
			}
		})
	}
}

func TestCompile_HelperClassAndStaticCallDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, target, modifier, directives, call string
		want                                     []string
	}{
		{"primitive class", "Integer", "class ", "", "T.Take(true, 1);", []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 3, column: 8]`}},
		{"interface class", "interface end", "class ", "", "T.Take(true, 1);", []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 3, column: 8]`}},
		{"record class", "record value: Integer; end", "class ", "", "T.Take(true, 1);", []string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 3, column: 14]`}},
		{"object class", "class end", "class ", "", "T.Take(true, 1);", []string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 3, column: 14]`}},
		{"record static", "record value: Integer; end", "class ", "static; ", "T.Take(true, 1);", []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 3, column: 8]`}},
		{"object static", "class end", "class ", "static; ", "T.Take(true, 1);", []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 3, column: 8]`}},
		{"primitive static short", "Integer", "class ", "static; ", "T.Take();", []string{`Syntax Error: More arguments expected [line: 3, column: 3]`}},
		{"record static excess", "record value: Integer; end", "class ", "static; ", "T.Take(1, 2);", []string{`Syntax Error: Too many arguments [line: 3, column: 3]`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type T = " + tt.target + ";\ntype H = helper for T " + tt.modifier + "procedure Take(v: Integer); " + tt.directives + "begin end; end;\n" + tt.call
			assertHelperCallDiagnostics(t, source, tt.want)
		})
	}
}

func TestCompile_HelperFunctionInheritanceAndRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"function helper", "function Take(s: String; v: Integer): String; helper; begin Result := s; end;\n('x').Take(true, 1);", []string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 2, column: 18]`}},
		{"function helper parameterless excess", "procedure Take(s: String); helper; begin end;\n('x').Take(1);", []string{`Syntax Error: Too many arguments [line: 2, column: 7]`}},
		{"inherited helper", "type H = helper for Integer procedure Take(v: Integer); begin end; end;\ntype J = helper(H) for Integer end;\nvar item: Integer;\nitem.Take(true, 1);", []string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 4, column: 17]`}},
		{"return recovery", "type H = helper for Integer function Take(v: Integer): String; begin Result := ''; end; end;\nvar item: Integer; var n: Integer;\nn := item.Take();", []string{`Syntax Error: More arguments expected [line: 3, column: 11]`, `Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 3, column: 6]`}},
		{"implicit Self helper", "type T = class procedure Run; end;\ntype H = helper for T procedure Take(v: Integer; w: String); begin end; end;\nprocedure T.Run; begin Take(true); end;", []string{`Syntax Error: Argument 1 expects type "Integer" instead of "Boolean" [line: 3, column: 24]`}},
		{"mixed overload ownership", "type T = record value: Integer; end;\ntype H = helper for T\n procedure Take(v: String); overload; begin end;\n class procedure Take(v: Integer); overload; static; begin end;\nend;\nT.Take(true);", []string{`Syntax Error: Argument 0 expects type "Integer" instead of "Boolean" [line: 6, column: 8]`}},
	} {
		t.Run(tt.name, func(t *testing.T) { assertHelperCallDiagnostics(t, tt.source, tt.want) })
	}
}

func assertHelperCallDiagnostics(t *testing.T, source string, want []string) {
	t.Helper()
	got := Compile(source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
	if !slices.Equal(got, want) {
		t.Fatalf("diagnostics:\n got %q\nwant %q", got, want)
	}
}
