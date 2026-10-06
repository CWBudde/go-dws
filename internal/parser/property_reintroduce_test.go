package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Deferred property punctuation must still reject malformed source through the
// parser API, before semantic resolution is available.
func TestParse_ReintroducedPropertyBoundaryDiagnostics(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"dynamic receiver", "var B := new ByteBuffer;\nB.SetLength(;", "Expression expected at 2:13"},
		{"property receiver", "type T = class F: Integer; property Prop: Integer read F reintroduce; end;\nObj.Prop(;", "Expression expected at 2:10"},
		{"nested call", "type T = class procedure M; begin end; end;\nvar O := new T;\nPrintLn(O.M(;\nMissing;", "Expression expected at 3:13"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(lexer.New(tt.source))
			_ = p.ParseProgram()
			errors := p.Errors()
			if len(errors) != 1 || errors[0].Error() != tt.want || !errors[0].Stop {
				t.Fatalf("got %v; want one compiler stop %q", errors, tt.want)
			}
		})
	}
}
