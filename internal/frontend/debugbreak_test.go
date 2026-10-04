package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// DebugBreak reads no argument expression: even a valid expression where ')'
// belongs is a punctuation stop, before any unknown-name or operand diagnostics.
func TestCompile_DebugBreak(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{"bare", "DebugBreak;", ""},
		{"empty parentheses", "DebugBreak();", ""},
		{"bare EOF", "DebugBreak", ""},
		{"empty parentheses EOF", "DebugBreak()", ""},
		{"missing close", "DebugBreak(;", `Syntax Error: ")" expected [line: 1, column: 12]`},
		{"argument is not read", "DebugBreak(Unknown);", `Syntax Error: ")" expected [line: 1, column: 12]`},
		{"literal is not read", "DebugBreak(1);", `Syntax Error: ")" expected [line: 1, column: 12]`},
		{"comma", "DebugBreak(,);", `Syntax Error: ")" expected [line: 1, column: 12]`},
		{"EOF anchor", "DebugBreak(", `Syntax Error: ")" expected [line: 1, column: 11]`},
		{"comment EOF", "DebugBreak( {note}", `Syntax Error: ")" expected [line: 1, column: 11]`},
		{"comment newline", "DebugBreak( {note}\n;", `Syntax Error: ")" expected [line: 2, column: 1]`},
		{"stop hides later errors", "DebugBreak(Unknown);\nMissing;", `Syntax Error: ")" expected [line: 1, column: 12]`},
		{"earlier error survives", "var I: Integer; I := 'x';\nDebugBreak(;", "Syntax Error: Incompatible types: Cannot assign \"String\" to \"Integer\" [line: 1, column: 22]\nSyntax Error: \")\" expected [line: 2, column: 12]"},
		{"ordinary call unchanged", "procedure F(X: Integer); begin end;\nF(;", `Syntax Error: Expression expected [line: 2, column: 3]`},
		{"no value", "PrintLn(DebugBreak);", `Syntax Error: Expression expected [line: 1, column: 19]`},
		{"no explicit value", "PrintLn(DebugBreak());", `Syntax Error: Expression expected [line: 1, column: 21]`},
		{"initializer has no value", "var X := DebugBreak;", `Syntax Error: Expression expected [line: 1, column: 20]`},
		{"constant recovery", "const X = DebugBreak;", `Syntax Error: Expression expected [line: 1, column: 21]`},
		{"typed constant recovery", "const X: Variant = DebugBreak;", `Syntax Error: Expression expected [line: 1, column: 30]`},
		{"constant stays defined during recovery", "const X = DebugBreak; PrintLn(X);", `Syntax Error: Expression expected [line: 1, column: 21]`},
		{"binary constant recovery", "const X = DebugBreak + 1; PrintLn(X);", `Syntax Error: Expression expected [line: 1, column: 22]`},
		{"unary constant recovery", "const X = -DebugBreak; PrintLn(X);", `Syntax Error: Expression expected [line: 1, column: 22]`},
		{"grouped statement has no value", "(DebugBreak);", `Syntax Error: Expression expected [line: 1, column: 12]`},
		{"grouped explicit statement has no value", "(DebugBreak());", `Syntax Error: Expression expected [line: 1, column: 14]`},
		{"nested grouped statement has no value", "((DebugBreak));", `Syntax Error: Expression expected [line: 1, column: 13]`},
		{"address punctuation stop", "var P := @DebugBreak(1);", `Syntax Error: ")" expected [line: 1, column: 22]`},
		{"no address", "var P := @DebugBreak;", "Syntax Error: Expression expected [line: 1, column: 21]\nSyntax Error: unexpected \"@\" [line: 1, column: 10]"},
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

func TestCompile_DebugBreakCaseHintBeforeStop(t *testing.T) {
	result := Compile("debugbreak(;", "<test>", semantic.HintsLevelPedantic)
	want := "Hint: \"debugbreak\" does not match case of declaration (\"DebugBreak\") [line: 1, column: 1]\nSyntax Error: \")\" expected [line: 1, column: 12]"
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}

func TestCompile_DebugBreakFixture(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "fixtures", "FailureScripts", "debugbreak")
	source, err := os.ReadFile(path + ".pas")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(path + ".txt")
	if err != nil {
		t.Fatal(err)
	}
	result := Compile(string(source), path+".pas", semantic.HintsLevelPedantic)
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.TrimSpace(string(want)) {
		t.Fatalf("got %q; want %q", got, want)
	}
}
