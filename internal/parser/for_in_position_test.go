package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestForInStatement_DoPosition(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         token.Position
		wantError    bool
	}{
		{name: "ordinary", source: "for c in a do ;", want: token.Position{Line: 1, Column: 12, Offset: 11}},
		{name: "multiline", source: "for c in a\n    do ;", want: token.Position{Line: 2, Column: 5, Offset: 15}},
		{name: "step", source: "for c in a step 2 do ;", want: token.Position{Line: 1, Column: 19, Offset: 18}},
		{name: "inline variable", source: "for var c in a do ;", want: token.Position{Line: 1, Column: 16, Offset: 15}},
		{name: "missing do", source: "for c in a PrintLn(c);", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.source)
			program := p.ParseProgram()
			if tt.wantError {
				if len(p.Errors()) == 0 {
					t.Fatal("expected the missing-do diagnostic")
				}
			} else {
				checkParserErrors(t, p)
			}
			if len(program.Statements) == 0 {
				t.Fatal("expected the for-in statement to survive parsing")
			}
			stmt, ok := program.Statements[0].(*ast.ForInStatement)
			if !ok {
				t.Fatalf("expected ForInStatement, got %T", program.Statements[0])
			}
			if stmt.DoPos != tt.want {
				t.Fatalf("do position = %+v, want %+v", stmt.DoPos, tt.want)
			}
		})
	}
}
