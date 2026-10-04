package frontend

import (
	"slices"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_RecordArityDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"instance short", "type R = record procedure Take(v: Integer); begin end; end;\nvar item: R;\nitem.Take();", []string{"Syntax Error: More arguments expected [line: 3, column: 6]"}},
		{"instance parameterless", "type R = record procedure Take; begin end; end;\nvar item: R;\nitem.Take(1);", []string{"Syntax Error: Too many arguments [line: 3, column: 6]"}},
		{"instance type suppresses count", "type R = record procedure Take(v: Integer; w: String); begin end; end;\nvar item: R;\nitem.Take(true);", []string{"Syntax Error: Argument 1 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 6]"}},
		{"instance shifted type anchor", "type R = record procedure Take(v: Integer; w: String); begin end; end;\nvar item: R;\nitem.Take(true,\n  'ok');", []string{"Syntax Error: Argument 1 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 3]"}},
		{"instance last type anchor", "type R = record procedure Take(v: Integer; w: String); begin end; end;\nvar item: R;\nitem.Take(1, true);", []string{"Syntax Error: Argument 2 expects type \"String\" instead of \"Boolean\" [line: 3, column: 6]"}},
		{"instance excess type", "type R = record procedure Take(v: Integer); begin end; end;\nvar item: R;\nitem.Take(true,\n  1);", []string{"Syntax Error: Argument 1 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 3]"}},
		{"instance child before short", "type R = record procedure Take(v: Integer; w: String); begin end; end;\nvar item: R;\nitem.Take(\n  Missing);", []string{"Syntax Error: Unknown name \"Missing\" [line: 4, column: 3]", "Syntax Error: More arguments expected [line: 3, column: 6]"}},
		{"instance excess children", "type R = record procedure Take; begin end; end;\nvar item: R;\nitem.Take(\n  Missing, Other);", []string{"Syntax Error: Unknown name \"Missing\" [line: 4, column: 3]", "Syntax Error: Unknown name \"Other\" [line: 4, column: 12]", "Syntax Error: Too many arguments [line: 3, column: 6]"}},
		{"instance marked", "type R = record procedure Take(v: Integer); overload; begin end; end;\nvar item: R;\nitem.Take();", []string{"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 6]"}},
		{"instance overloaded", "type R = record procedure Take(v: Integer); overload; begin end; procedure Take(v: String); overload; begin end; end;\nvar item: R;\nitem.\n  Take(true);", []string{"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 4, column: 3]"}},
		{"class short", "type R = record class procedure Take(v: Integer); begin end; end;\nR.Take();", []string{"Syntax Error: More arguments expected [line: 2, column: 3]"}},
		{"class parameterless", "type R = record class procedure Take; begin end; end;\nR.Take(1);", []string{"Syntax Error: Too many arguments [line: 2, column: 3]"}},
		{"class type before count", "type R = record class procedure Take(v: Integer; w: String); begin end; end;\nR.Take(\n  true);", []string{"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 3]"}},
		{"class through instance", "type R = record class procedure Take(v: Integer); begin end; end;\nvar item: R;\nitem.Take();", []string{"Syntax Error: More arguments expected [line: 3, column: 6]"}},
		{"class marked anchor", "type R = record class procedure Take(v: Integer); overload; begin end; end;\nvar meta := R;\nmeta.\n  Take(true);", []string{"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 4, column: 3]"}},
		{"instance defaults", "type R = record procedure Take(v: Integer = 7); begin end; end;\nvar item: R; item.Take();", nil},
		{"class defaults", "type R = record class procedure Take(v: Integer = 7); begin end; end;\nR.Take();", nil},
		{"implicit instance short", "type R = record procedure Take(v: Integer); begin end;\nprocedure Run; begin\n  Take();\nend; end;", []string{"Syntax Error: More arguments expected [line: 3, column: 3]"}},
		{"implicit instance wrong type", "type R = record procedure Take(v: Integer); begin end;\nprocedure Run; begin\n  Take(true);\nend; end;", []string{"Syntax Error: Argument 1 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 3]"}},
		{"implicit class short", "type R = record class procedure Take(v: Integer); begin end;\nprocedure Run; begin\n  Take();\nend; end;", []string{"Syntax Error: More arguments expected [line: 3, column: 3]"}},
		{"implicit class wrong type", "type R = record class procedure Take(v: Integer); begin end;\nprocedure Run; begin\n  Take(true);\nend; end;", []string{"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 8]"}},
		{"return recovery", "type R = record function Take(v: Integer): String; begin Result := ''; end; end;\nvar item: R;\nvar n: Integer;\nn := item.Take();", []string{"Syntax Error: More arguments expected [line: 4, column: 11]", "Syntax Error: Incompatible types: Cannot assign \"String\" to \"Integer\" [line: 4, column: 6]"}},
		{"recursive function short", "type R = record function Take(v: Integer): String;\nbegin\n  Take();\nend; end;", []string{"Syntax Error: More arguments expected [line: 3, column: 3]"}},
		{"recursive function type", "type R = record function Take(v: Integer): String;\nbegin\n  Take(true);\nend; end;", []string{"Syntax Error: Argument 1 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 3]"}},
		{"recursive marked function", "type R = record function Take(v: Integer): String; overload;\nbegin\n  Take();\nend; end;", []string{"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 3]"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestCompile_RecordOverloadsReadAllChildren(t *testing.T) {
	for _, tt := range []struct{ name, source string }{
		{"instance", "type R = record procedure Take(v: Integer); overload; begin end; end;\nvar item: R;\nitem.Take(\n  Missing1,\n  Missing2);"},
		{"class", "type R = record class procedure Take(v: Integer); overload; begin end; end;\nvar meta := R;\nmeta.Take(\n  Missing1,\n  Missing2);"},
		{"implicit", "type R = record procedure Take(v: Integer); overload; begin end;\nprocedure Run; begin\n  Take(\n  Missing1,\n  Missing2);\nend; end;"},
		{"implicit class", "type R = record class procedure Take(v: Integer); overload; begin end;\nclass procedure Run; begin\n  Take(\n  Missing1,\n  Missing2);\nend; end;"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			want := []string{
				"Syntax Error: Unknown name \"Missing1\" [line: 4, column: 3]",
				"Syntax Error: Unknown name \"Missing2\" [line: 5, column: 3]",
			}
			if !slices.Equal(got, want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, want)
			}
		})
	}
}

func TestCompile_RecordOverloadChildBeforeFailure(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"instance", "type R = record procedure Take(v: Integer); overload; begin end; end;\nvar item: R;\nitem.Take(\n  1 as String, 2);", []string{
			"Syntax Error: Cannot cast \"Integer\" as \"String\" [line: 4, column: 5]",
			"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 6]",
		}},
		{"class", "type R = record class procedure Take(v: Integer); overload; begin end; end;\nvar meta := R;\nmeta.Take(\n  1 as String, 2);", []string{
			"Syntax Error: Cannot cast \"Integer\" as \"String\" [line: 4, column: 5]",
			"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 6]",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}
