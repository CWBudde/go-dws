package frontend

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// The pinned upstream Default result unit registers only Print and PrintLn.
// Expectations here follow unit-local lookup, not special-keyword recognition.
func TestCompile_DefaultNamespaceUnavailableMembers(t *testing.T) {
	names := []string{"Assert", "Assigned", "High", "Length", "Low", "Ord", "SizeOf", "Defined", "Declared", "Inc", "Dec", "Succ", "Pred", "Include", "Exclude", "Swap", "ConditionalDefined", "Default", "DebugBreak", "Abs", "Integer", "Sin", "Unknown"}
	contexts := []struct {
		name, prefix, suffix string
		column               int
	}{
		{"call", "Default.", "(Unknown); Later;", 9},
		{"bare", "var X := Default.", "; Later;", 18},
		{"address", "var P := @Default.", "; Later;", 19},
	}
	for _, name := range names {
		for _, context := range contexts {
			t.Run(name+"/"+context.name, func(t *testing.T) {
				result := Compile(context.prefix+name+context.suffix, "<test>", semantic.HintsLevelPedantic)
				want := fmt.Sprintf(`Syntax Error: Unknown name "Default.%s" [line: 1, column: %d]`, name, context.column)
				if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
					t.Fatalf("got %q; want %q", got, want)
				}
			})
		}
	}
}

func TestCompile_DefaultNamespaceValuelessInitializer(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"inferred bare procedure", "var P := Default.PrintLn;", `Syntax Error: Assignment's right-side-argument has no return type [line: 1, column: 7]`},
		{"inferred explicit procedure", "var P := Default.PrintLn();", `Syntax Error: Assignment's right-side-argument has no return type [line: 1, column: 7]`},
		{"typed explicit procedure", "var P: Variant := Default.PrintLn();", `Syntax Error: Assignment's right-side-argument has no return type [line: 1, column: 5]`},
		{"initializer scanner anchor", "var P {note}\n:= Default.PrintLn;", `Syntax Error: Assignment's right-side-argument has no return type [line: 2, column: 1]`},
		{"ordinary explicit procedure", "procedure Work; begin end;\nvar P := Work();", `Syntax Error: Assignment's right-side-argument has no return type [line: 2, column: 7]`},
		{"inferred recovery binds name", "var P := Default.PrintLn; PrintLn(P);", `Syntax Error: Assignment's right-side-argument has no return type [line: 1, column: 7]`},
		{"typed recovery binds name", "var P: Variant := Default.PrintLn(); PrintLn(P);", `Syntax Error: Assignment's right-side-argument has no return type [line: 1, column: 5]`},
		{"typed recovery preserves type", "var X: Integer := Default.PrintLn(); X := 'bad';", "Syntax Error: Assignment's right-side-argument has no return type [line: 1, column: 5]\nSyntax Error: Incompatible types: Cannot assign \"String\" to \"Integer\" [line: 1, column: 43]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelDisabled)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_DefaultNamespaceRecovery(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"qualifier hint precedes member stop", "default.low(Unknown); Later;", "Hint: \"default\" does not match case of declaration (\"Default\") [line: 1, column: 1]\nSyntax Error: Unknown name \"Default.low\" [line: 1, column: 9]"},
		{"no member casing hint", "default.println('ok');", "Hint: \"default\" does not match case of declaration (\"Default\") [line: 1, column: 1]"},
		{"comment and newline anchor", "Default. {note}\nlow(Unknown);", `Syntax Error: Unknown name "Default.low" [line: 2, column: 1]`},
		{"grouped callee", "(Default.Low)(Unknown);", `Syntax Error: Unknown name "Default.Low" [line: 1, column: 10]`},
		{"bare at EOF", "Default.Length", `Syntax Error: Unknown name "Default.Length" [line: 1, column: 9]`},
		{"earlier diagnostic remains", "var I: Integer; I := 'x';\nDefault.Low(Unknown);", "Syntax Error: Incompatible types: Cannot assign \"String\" to \"Integer\" [line: 1, column: 22]\nSyntax Error: Unknown name \"Default.Low\" [line: 2, column: 9]"},
		{"bare procedure cannot supply value", "Default.PrintLn(Default.PrintLn);", `Syntax Error: Argument 0 expects type "Variant" [line: 1, column: 17]`},
		{"grouped bare procedure cannot supply value", "PrintLn((Default.PrintLn));", `Syntax Error: Argument 0 expects type "Variant" [line: 1, column: 9]`},
		{"bare procedure user argument", "procedure P(V: Variant); begin end;\nP(Default.PrintLn);", `Syntax Error: Argument 0 expects type "Variant" [line: 2, column: 3]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}
