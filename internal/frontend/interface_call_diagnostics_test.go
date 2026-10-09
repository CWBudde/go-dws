package frontend

import (
	"slices"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_InterfaceCallDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, members, call string
		want                []string
	}{
		{"short", "procedure Take(v: Integer);", "item.Take();", []string{"Syntax Error: More arguments expected [line: 3, column: 6]"}},
		{"excess", "procedure Take(v: Integer);", "item.Take(1, 2);", []string{"Syntax Error: Too many arguments [line: 3, column: 6]"}},
		{"parameterless", "procedure Take;", "item.Take(1);", []string{"Syntax Error: Too many arguments [line: 3, column: 6]"}},
		{"type before short", "procedure Take(v: Integer; w: String);", "item.Take(true);", []string{"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 11]"}},
		{"type before excess", "procedure Take(v: Integer);", "item.Take(true, 1);", []string{"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 11]"}},
		{"second type", "procedure Take(v: Integer; w: String);", "item.Take(1, true);", []string{"Syntax Error: Argument 1 expects type \"String\" instead of \"Boolean\" [line: 3, column: 14]"}},
		{"multiline method anchor", "procedure Take(v: Integer);", "item.\n  Take();", []string{"Syntax Error: More arguments expected [line: 4, column: 3]"}},
		{"multiline type anchor", "procedure Take(v: Integer);", "item.Take(\n  true, 2);", []string{"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 3]"}},
		{"child before short", "procedure Take(v: Integer; w: String);", "item.Take(\n  Missing);", []string{"Syntax Error: Unknown name \"Missing\" [line: 4, column: 3]"}},
		{"excess children", "procedure Take;", "item.Take(\n  Missing, Other);", []string{"Syntax Error: Unknown name \"Missing\" [line: 4, column: 3]"}},
		{"cast child before short", "procedure Take(v: Integer; w: String);", "item.Take(\n  1 as String);", []string{"Syntax Error: Cannot cast \"Integer\" as \"String\" [line: 4, column: 5]", "Syntax Error: More arguments expected [line: 3, column: 6]"}},
		{"case insensitive", "procedure Take(v: Integer);", "item.tAKE();", []string{"Syntax Error: More arguments expected [line: 3, column: 6]"}},
		{"valid", "procedure Take(v: Integer);", "item.Take(1);", nil},
		{"valid parameterless", "procedure Take;", "item.Take();", nil},
		{"contextual empty array", "procedure Take(v: array of Integer);", "item.Take([]);", nil},
		{"contextual nil", "procedure Take(v: I);", "item.Take(nil);", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type I = interface " + tt.members + " end;\nvar item: I;\n" + tt.call
			got := Compile(source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestCompile_InterfaceCallInheritanceAndRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"inherited short", "type I = interface procedure Take(v: Integer); end;\ntype J = interface(I) end;\nvar item: J;\nitem.Take();", []string{"Syntax Error: More arguments expected [line: 4, column: 6]"}},
		{"inherited type", "type I = interface procedure Take(v: Integer); end;\ntype J = interface(I) end;\nvar item: J;\nitem.Take(true, 1);", []string{"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 11]"}},
		{"return recovery", "type I = interface function Take(v: Integer): String; end;\nvar item: I;\nvar n: Integer;\nn := item.Take();", []string{"Syntax Error: More arguments expected [line: 4, column: 11]", "Syntax Error: Incompatible types: Cannot assign \"String\" to \"Integer\" [line: 4, column: 6]"}},
		{"receiver expression", "type I = interface procedure Take(v: Integer); end;\nfunction Make: I; begin Result := nil; end;\nMake().Take(true, 1);", []string{"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 13]"}},
		{"grouped receiver", "type I = interface procedure Take(v: Integer); end;\nvar item: I;\n(item).Take(true, 1);", []string{"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 13]"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}
