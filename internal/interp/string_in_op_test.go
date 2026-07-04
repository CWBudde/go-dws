package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestStringInBracketList verifies DWScript's string comparison semantics for
// 'in' with a bracket list: ranges compare the whole string lexicographically
// and single elements compare for equality (see fixture string_in_op3).
func TestStringInBracketList(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		{"", "False"},
		{"a", "True"},
		{"alpha", "True"},
		{"B", "False"},
		{"Beta", "False"},
		{"5", "True"},
		{"12345", "True"},
		{"$", "False"},
		{"_", "False"},
		{"---", "False"},
		{"-", "True"},
	}

	for _, tt := range tests {
		t.Run("value_"+tt.value, func(t *testing.T) {
			src := "PrintLn('" + tt.value + "' in ['a'..'z', '0'..'9', '-']);"
			l := lexer.New(src)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
			}

			var buf bytes.Buffer
			interp := New(&buf)
			interp.Eval(program)

			if got := strings.TrimSpace(buf.String()); got != tt.expected {
				t.Errorf("%q in [...] printed %q, want %q", tt.value, got, tt.expected)
			}
		})
	}
}
