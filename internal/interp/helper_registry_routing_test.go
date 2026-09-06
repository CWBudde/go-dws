package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// runHelperScript compiles and runs source, returning its output. It fails the
// test on compile diagnostics or a runtime error.
func runHelperScript(t *testing.T, source string) string {
	t.Helper()

	compiled := frontend.Compile(source, "helper_routing.pas", semantic.HintsLevelNormal)
	if compiled.HasFatalDiagnostics() || !compiled.SemanticSuccessful {
		t.Fatalf("compile diagnostics:\n%s", strings.Join(compiled.DiagnosticStrings(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if compiled.SemanticInfo != nil {
		interp.SetSemanticInfo(compiled.SemanticInfo)
	}
	result := interp.Eval(compiled.Program)
	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("runtime error: %s", result.String())
	}
	return buf.String()
}

// TestStringHelperRoutesRegistryBackedSpecs covers helper specs that are plain
// builtin names (PadLeft, StripAccents, StrDeleteLeft, ...) rather than
// "__"-prefixed evaluator specs. They have no evaluator case and must be routed
// to builtins.DefaultRegistry with the receiver as the first argument.
func TestStringHelperRoutesRegistryBackedSpecs(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"PadLeft method", `PrintLn('ab'.PadLeft(5, '0'));`, "000ab\n"},
		{"PadRight method", `PrintLn('ab'.PadRight(4, '-'));`, "ab--\n"},
		{"StripAccents method", `PrintLn('héllo'.StripAccents);`, "hello\n"},
		{"StripAccents property", `var s := 'héllo'; PrintLn(s.StripAccents);`, "hello\n"},
		{"DeleteLeft method", `PrintLn('abcdef'.DeleteLeft(2));`, "cdef\n"},
		{"DeleteRight method", `PrintLn('abcdef'.DeleteRight(2));`, "abcd\n"},
		// 'e' + U+0301 (combining acute) normalizes under NFC to a single U+00E9.
		{"Normalize method", `var s := 'e' + #$0301; PrintLn(Length(s)); PrintLn(Length(s.Normalize('NFC')));`, "2\n1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runHelperScript(t, tt.source)
			if got != tt.want {
				t.Fatalf("output mismatch:\nwant: %q\ngot:  %q", tt.want, got)
			}
		})
	}
}
