package frontend

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// Upstream ReadName identifies special keywords only for an unqualified name
// token. A System or Internal prefix resolves its member with FindLocal in that
// unit's own table, which holds no special keyword, so each qualified special
// name stops with CPE_UnknownNameDotName at the member, before any argument is
// read. Anchors are derived from the pinned compiler source, not an oracle run.
func TestCompile_UnitQualifiedSpecialNames(t *testing.T) {
	specials := []string{
		"Assert", "Assigned", "Default", "High", "Length", "Low", "Ord", "SizeOf",
		"Defined", "Declared", "Inc", "Dec", "Succ", "Pred", "Include", "Exclude",
		"Swap", "ConditionalDefined", "DebugBreak",
	}
	forms := []struct{ name, prefix, suffix string }{
		{"statement", "", "(Unknown); Later;"},
		{"argument", "PrintLn(", "(Unknown)); Later;"},
		{"value", "var X := ", "; Later;"},
	}
	for _, unit := range []string{"System", "Internal"} {
		for _, name := range specials {
			for _, form := range forms {
				t.Run(unit+"/"+name+"/"+form.name, func(t *testing.T) {
					source := form.prefix + unit + "." + name + form.suffix
					column := len(form.prefix) + len(unit) + 2
					want := fmt.Sprintf(`Syntax Error: Unknown name "%s.%s" [line: 1, column: %d]`, unit, name, column)
					result := Compile(source, "<test>", semantic.HintsLevelDisabled)
					if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
						t.Fatalf("%s\ngot  %q\nwant %q", source, got, want)
					}
				})
			}
		}
	}
}

func TestCompile_UnitQualifiedSpecialNameSpelling(t *testing.T) {
	tests := []struct{ source, want string }{
		{"system.low(Unknown);", `Syntax Error: Unknown name "System.low" [line: 1, column: 8]`},
		{"INTERNAL.HIGH(Unknown);", `Syntax Error: Unknown name "Internal.HIGH" [line: 1, column: 10]`},
		{"System. {note}\nLow(Unknown);", `Syntax Error: Unknown name "System.Low" [line: 2, column: 1]`},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelDisabled)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestCompile_UnitQualifiedSpecialNameControls(t *testing.T) {
	for _, source := range []string{
		"var X: System.Integer := 3; PrintLn(Low(X)); PrintLn(High(Integer));",
		"type T = class function Low: Integer; begin Result := 1; end; end; var O := T.Create; PrintLn(O.Low());",
		"Default.PrintLn(Length('abc'));",
	} {
		t.Run(source, func(t *testing.T) {
			result := Compile(source, "<test>", semantic.HintsLevelDisabled)
			if got := result.DiagnosticStrings(); len(got) != 0 {
				t.Fatalf("unexpected diagnostics: %q", got)
			}
		})
	}
}
