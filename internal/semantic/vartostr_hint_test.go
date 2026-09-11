package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// analyzeWithHints analyzes input at the given hint level and returns every
// diagnostic the analyzer accumulated.
func analyzeWithHints(t *testing.T, input string, level HintsLevel) []string {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.SetHintsLevel(level)
	_ = analyzer.Analyze(program)
	return analyzer.Errors()
}

// VarToStr accepts any type, but DWScript points at the dedicated conversion
// when the static type already tells which one applies. The expected wording and
// column are recorded in testdata/fixtures/FunctionsVariant/vartostr*.txt.
func TestVarToStrHints(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "integer prefers IntToStr",
			input: "var i := 123;\nPrintLn(VarToStr(i));",
			want:  "Hint: Prefer .ToString or IntToStr() [line: 2, column: 9]",
		},
		{
			name:  "float prefers FloatToStr",
			input: "var f := 1.5;\nPrintLn(VarToStr(f));",
			want:  "Hint: Prefer .ToString or FloatToStr() [line: 2, column: 9]",
		},
		{
			name:  "string conversion is redundant",
			input: "var s := 'x';\nPrintLn(VarToStr(s));",
			want:  "Hint: Redundant function call [line: 2, column: 9]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !containsDiagnostic(analyzeWithHints(t, tt.input, HintsLevelPedantic), tt.want) {
				t.Errorf("missing %q in %v", tt.want, analyzeWithHints(t, tt.input, HintsLevelPedantic))
			}
			if got := analyzeWithHints(t, tt.input, HintsLevelDisabled); containsDiagnostic(got, tt.want) {
				t.Errorf("hint emitted with hints disabled: %v", got)
			}
		})
	}
}

// A Variant argument is exactly what VarToStr is for, so no hint applies.
func TestVarToStrHints_NotEmittedForVariant(t *testing.T) {
	input := "var v : Variant;\nv := 12.3;\nPrintLn(VarToStr(v));"
	for _, diagnostic := range analyzeWithHints(t, input, HintsLevelPedantic) {
		if strings.Contains(diagnostic, "Prefer .ToString") || strings.Contains(diagnostic, "Redundant function call") {
			t.Errorf("unexpected VarToStr hint for a Variant argument: %s", diagnostic)
		}
	}
}

func containsDiagnostic(diagnostics []string, want string) bool {
	for _, d := range diagnostics {
		if d == want {
			return true
		}
	}
	return false
}
