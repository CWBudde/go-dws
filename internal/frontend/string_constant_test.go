package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_ReportsStringConstantErrors checks that malformed string and char
// constants stop compilation with DWScript's tokenizer diagnostic and nothing else
// (FailureScripts/triple_apos1, triple_apos2, heredoc, invalid_ucs2_char).
func TestCompile_ReportsStringConstantErrors(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "triple apostrophe indentation",
			source: "var s := '''\n   qsd\n     ''';",
			want:   []string{`Syntax Error: Incorrect triple apostrophe string indentation [line: 1, column: 10]`},
		},
		{
			name:   "triple apostrophe string without line break",
			source: "var pass := ''' qsdq ''';\nvar fail := ''' qsdq \n''';",
			want:   []string{`Syntax Error: Incorrect triple apostrophe string [line: 2, column: 13]`},
		},
		{
			name:   "single-quoted string reaching end of line",
			source: "var srv : Integer;\n\nPrintLn(1, ');",
			want:   []string{`Syntax Error: End of string constant not found (end of line) [line: 3, column: 12]`},
		},
		{
			name:   "double-quoted string reaching end of file",
			source: "PrintLn(\"unfinished\nheredoc",
			want:   []string{`Syntax Error: End of string constant not found (end of file) [line: 4, column: 1]`},
		},
		{
			name:   "double-quoted string reaching end of terminated file",
			source: "PrintLn(\"unfinished\nheredoc\n",
			want:   []string{`Syntax Error: End of string constant not found (end of file) [line: 4, column: 1]`},
		},
		{
			name:   "char constant beyond the Unicode range",
			source: "const s = #$10000;\nconst e = #$200000;",
			want:   []string{`Syntax Error: Invalid char constant "$200000" [line: 2, column: 19]`},
		},
		{
			// DWScript's parser stops at the first syntax error, before its lazily
			// pulled tokenizer reaches the unterminated constant.
			name:   "constant error after a syntax error is not reached",
			source: "const CText = 'foo'''./bar.';",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 23]`},
		},
		{
			name:   "decimal char constant beyond the Unicode range inside a string sequence",
			source: "var s := 'a'#99999999'b';",
			want:   []string{`Syntax Error: Invalid char constant "99999999" [line: 1, column: 22]`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "strings.pas", semantic.HintsLevelPedantic)
			got := result.DiagnosticStrings()
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Fatalf("diagnostics = %q, want %q", got, tt.want)
			}
			if !result.HasFatalDiagnostics() {
				t.Fatalf("expected a fatal diagnostic")
			}
		})
	}
}

// TestCompile_NamelessVarDeclarationReportsOnlyNameExpected checks that a var
// declaration with a missing or malformed name is skipped after "Name expected", so its
// type and initializer are not re-parsed as statements
// (FailureScripts/reserved_escape_empty, reserved_escape_number).
func TestCompile_NamelessVarDeclarationReportsOnlyNameExpected(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "bare escape",
			source: "var & : Integer := 1;\n",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 5]`},
		},
		{
			name:   "escaped number",
			source: "var &1 : Integer := 2;",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 5]`},
		},
		{
			name:   "number",
			source: "var 1 : Integer := 2;",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 5]`},
		},
		{
			name:   "code after the declaration is still compiled",
			source: "var &1 : Integer := 2;\nPrintLn(Undeclared);",
			want: []string{
				`Syntax Error: Name expected [line: 1, column: 5]`,
				`Syntax Error: Unknown name "Undeclared" [line: 2, column: 9]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "var.pas", semantic.HintsLevelPedantic).DiagnosticStrings()
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Fatalf("diagnostics = %q, want %q", got, tt.want)
			}
		})
	}
}
