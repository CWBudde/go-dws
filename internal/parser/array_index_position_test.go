package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestParseArrayIndexing_CommaPositions(t *testing.T) {
	p := New(lexer.New("a[0, 1, 2][3]"))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	index := asIndexExpr(t, getExprFromStmt(t, program.Statements[0]))
	for _, want := range []struct{ bracket, comma int }{{11, 0}, {2, 7}, {2, 4}, {2, 0}} {
		if index.Token.Pos.Column != want.bracket || index.CommaPos.Column != want.comma {
			t.Fatalf("positions = bracket %d, comma %d; want %d, %d", index.Token.Pos.Column, index.CommaPos.Column, want.bracket, want.comma)
		}
		if want.comma != 0 || want.bracket == 11 {
			index = asIndexExpr(t, index.Left)
		}
	}
}
