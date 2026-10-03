package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestParser_HeaderCompilerStops(t *testing.T) {
	for _, source := range []string{
		"class bug; Missing;",
		"type T = class class bug; end; Missing;",
		"type T = record class bug; end; Missing;",
		"type T = helper for Integer class bug; end; Missing;",
		"type T = record F: Integer;",
		"type T = helper for Integer class const C = 1;",
	} {
		t.Run(source, func(t *testing.T) {
			p := New(lexer.New(source))
			p.ParseProgram()
			if len(p.Errors()) != 1 || !p.stopped() {
				t.Fatalf("want one compiler stop, got stopped=%v, errors=%v", p.stopped(), p.Errors())
			}
		})
	}
}
