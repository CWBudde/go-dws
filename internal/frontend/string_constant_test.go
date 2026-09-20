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
			// The dot's "Name expected" is a compiler stop upstream
			// (ReadSymbolMemberExpr), so its lazily pulled tokenizer never reaches
			// the unterminated constant that follows.
			name:   "constant error after a compiler stop is not reached",
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
// declaration with a missing or malformed name reports "Name expected" and nothing
// else (FailureScripts/reserved_escape_empty, reserved_escape_number). Upstream
// reports it from ReadNameList via AddCompilerStop, which raises out of the compiler,
// so neither the declaration's own remainder nor any following statement is compiled.
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
			name:   "the stop abandons the code after the declaration",
			source: "var &1 : Integer := 2;\nPrintLn(Undeclared);",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 5]`},
		},
		{
			name:   "the stop abandons a statement on the malformed name's own line",
			source: "var &\nPrintLn(Undeclared);",
			want:   []string{`Syntax Error: Name expected [line: 1, column: 5]`},
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

// TestCompile_CompilerStopDropsLaterLexerDiagnostics checks that a compiler stop also
// suppresses the lexer's own diagnostics that follow it. Upstream's AddCompilerStop
// raises ECompileError out of the compiler, so its lazily pulled tokenizer never reads
// past the stop and a later message directive is never processed. Directives before
// the stop have already been reached and are kept.
func TestCompile_CompilerStopDropsLaterLexerDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "directive after the stop is never reached",
			source: "var x := ;\n{$ERROR 'late'}\n",
			want:   []string{`Syntax Error: Expression expected [line: 1, column: 10]`},
		},
		{
			name:   "directive before the stop is kept",
			source: "{$HINT 'early'}\nvar x := ;\n{$ERROR 'late'}\n",
			want: []string{
				`Hint: early [line: 1, column: 3]`,
				`Syntax Error: Expression expected [line: 2, column: 10]`,
			},
		},
		{
			name:   "without a stop a later directive is still reported",
			source: "{$ERROR 'late'}\n",
			want:   []string{`Compile Error: late [line: 1, column: 3]`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "stop.pas", semantic.HintsLevelPedantic).DiagnosticStrings()
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Fatalf("diagnostics = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCompile_RecoverableErrorKeepsLaterConstantError checks that only a compiler stop
// cuts the lexer's constant diagnostics off, not any parser error. DWScript's
// AddCompilerError leaves the compilation running, so its tokenizer does reach a
// malformed constant further down and reports it; only AddCompilerStop abandons the
// compile. `case` without `of` is such a recoverable error.
func TestCompile_RecoverableErrorKeepsLaterConstantError(t *testing.T) {
	const source = "var i := 1;\ncase i\n  1: PrintLn('one);\nend;\n"
	got := Compile(source, "recover.pas", semantic.HintsLevelPedantic).DiagnosticStrings()
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "End of string constant not found (end of line)") {
		t.Fatalf("constant diagnostic dropped after a recoverable parser error: %q", got)
	}
}
