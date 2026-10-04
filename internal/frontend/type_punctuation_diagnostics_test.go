package frontend

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// These fixtures distinguish symbol-directed punctuation from ordinary expression recovery.
func TestCompile_TypeDirectedPunctuationFixtures(t *testing.T) {
	for _, name := range []string{"const_record1", "special_funcs1", "at_integer", "const_record4"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", "testdata", "fixtures", "FailureScripts", name)
			source, err := os.ReadFile(path + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(path + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			result := Compile(string(source), path+".pas", semantic.HintsLevelPedantic)
			want := strings.TrimSpace(string(expected))
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
				t.Fatalf("diagnostics mismatch\n got: %s\nwant: %s", got, want)
			}
		})
	}
}

func TestCompile_TypeDirectedPunctuationRecovery(t *testing.T) {
	tests := []struct {
		name, source string
		want         []string
	}{
		{"record alias missing initializer", "type R = record X: Integer; end; type A = R;\nconst C: A = ;", []string{`Syntax Error: "(" expected [line: 2, column: 14]`}},
		{"record constant without opening delimiter", "type R = record X: Integer; end;\nconst C: R = 3;", []string{`Syntax Error: "(" expected [line: 2, column: 14]`}},
		{"record stop suppresses later errors", "type R = record X: Integer; end;\nconst C: R = 3;\nUnknown;", []string{`Syntax Error: "(" expected [line: 2, column: 14]`}},
		{"local record constant", "type R = record X: Integer; end;\nprocedure P; begin const C: R = ; end;", []string{`Syntax Error: "(" expected [line: 2, column: 33]`}},
		{"record expression stops before operand analysis", "type R = record X: Integer; end;\nconst C: R = 1 + Unknown;", []string{`Syntax Error: "(" expected [line: 2, column: 14]`}},
		{"bare Low skips comments", "var I := Low {note} ;", []string{`Syntax Error: "(" expected [line: 1, column: 21]`}},
		{"bare Low at newline", "var I := Low\n;", []string{`Syntax Error: "(" expected [line: 2, column: 1]`}},
		{"bare Low at EOF", "var I := Low", []string{`Syntax Error: "(" expected [line: 1, column: 10]`}},
		{"address of type skips comments", "var P := @Integer {note} ;", []string{`Syntax Error: "(" expected [line: 1, column: 26]`, `Syntax Error: unexpected "@" [line: 1, column: 10]`}},
		{"bare High requires opening delimiter", "var I := High;", []string{`Syntax Error: "(" expected [line: 1, column: 14]`}},
		{"interrupted scalar duplicate stays parser only", "const C = 1; const C: Integer = ;", []string{`Syntax Error: Expression expected [line: 1, column: 33]`}},
		{"scalar missing initializer", "const C: Integer = ;", []string{`Syntax Error: Expression expected [line: 1, column: 20]`}},
		{"ordinary call missing first argument", "procedure F(x: Integer); begin end;\nF(;", []string{`Syntax Error: Expression expected [line: 2, column: 3]`}},
		{"earlier semantic error survives record stop", "var I: Integer; I := 'x';\ntype R = record X: Integer; end;\nconst C: R = ;", []string{`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 1, column: 22]`, `Syntax Error: "(" expected [line: 3, column: 14]`}},
		{"bare special stops later analysis", "var I := Low;\nUnknown;", []string{`Syntax Error: "(" expected [line: 1, column: 13]`}},
		{"address of variable uses unexpected at", "var I: Integer; var P := @I;", []string{`Syntax Error: unexpected "@" [line: 1, column: 26]`}},
		{"address of type preserves child first", "var P := @Integer;", []string{`Syntax Error: "(" expected [line: 1, column: 18]`, `Syntax Error: unexpected "@" [line: 1, column: 10]`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelDisabled)
			if got := result.DiagnosticStrings(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_TypeDirectedPunctuationValidControls(t *testing.T) {
	for _, source := range []string{
		"type R = record X: Integer; end; const C: R = (X: 3); PrintLn(C.X);",
		"var I := Low(Integer); PrintLn(I);",
		"var Low: Integer := 3; var I := Low; PrintLn(I);",
		"function Low: Integer; begin Result := 3; end; var I := Low; PrintLn(I);",
		"function F: Integer; begin Result := 3; end; var P := @F; PrintLn(P());",
		"var P := @Abs; PrintLn(P(-3));",
		"type T = function: Integer; function F: Integer; begin Result := 3; end; var P: T := @F; var Q := @P; PrintLn(Q());",
	} {
		t.Run(source, func(t *testing.T) {
			result := Compile(source, "<test>", semantic.HintsLevelDisabled)
			if got := result.DiagnosticStrings(); len(got) != 0 {
				t.Fatalf("unexpected diagnostics: %q", got)
			}
		})
	}
}
