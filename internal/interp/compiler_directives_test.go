package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

func TestCompilerDirectives_IfDef(t *testing.T) {
	source := `
{$define TEST}
var s: String := '';
{$ifdef TEST}
	s := 'defined';
{$else}
	s := 'not defined';
{$endif}
PrintLn(s);
`

	_, out := testEvalWithOutput(source)
	if out != "defined\n" {
		t.Fatalf("expected output to use defined branch, got %q", out)
	}
}

func TestCompilerDirectives_IfNDefSkipsInvalidCode(t *testing.T) {
	source := `
{$ifndef SOMETHING}
{$define SOMETHING}
{$endif}

{$ifndef SOMETHING}
	this should never be parsed
{$endif}
PrintLn('ok');
`

	l := lexer.New(source)
	p := parser.New(l)
	_ = p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	if errs := p.LexerErrors(); len(errs) != 0 {
		t.Fatalf("unexpected lexer errors: %v", errs)
	}
}

func TestCompilerDirectives_UnbalancedReportsError(t *testing.T) {
	source := `{$ifdef TEST}
var x := 1;
`

	l := lexer.New(source)
	p := parser.New(l)
	_ = p.ParseProgram()

	if len(p.LexerErrors()) == 0 {
		t.Fatalf("expected lexer errors for unfinished conditional directive, got none")
	}
}
