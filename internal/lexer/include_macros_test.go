package lexer

import (
	"slices"
	"testing"
)

// Include substitutions use the active source's line counter, then resume the
// parent's counter when the include ends.
func TestIncludeMacro_SourceLines(t *testing.T) {
	files := map[string]string{"lines.inc": "{$I %LINENUM%}\n{$I %LINE%}"}
	l := New("\n{$I 'lines.inc'}\n{$I %LINENUM%}", WithIncludeResolver(mapIncludeResolver(files)))
	for _, want := range []struct {
		text string
		kind TokenType
	}{{"1", INT}, {"2", STRING}, {"3", INT}} {
		tok := l.NextToken()
		if tok.Type != want.kind || tok.Literal != want.text {
			t.Fatalf("token = %v %q, want %v %q", tok.Type, tok.Literal, want.kind, want.text)
		}
	}
	if tok := l.NextToken(); tok.Type != EOF || len(l.Errors()) != 0 {
		t.Fatalf("remaining token = %v, errors = %v", tok, l.Errors())
	}
}

func TestIncludeMacro_LookaheadAndRewind(t *testing.T) {
	l := New("{$I %LINENUM%} +\n{$I %LINE%}")
	state := l.SaveState()
	if tok := l.Peek(2); tok.Type != STRING || tok.Literal != "2" {
		t.Fatalf("lookahead = %v, want substituted line string", tok)
	}
	for i := 0; i < 2; i++ {
		l.RestoreState(state)
		if got := collectLiterals(l); !slices.Equal(got, []string{"1", "+", "2"}) {
			t.Fatalf("rewind %d = %v, want one copy of each substituted token", i, got)
		}
	}
}

func TestConditional_BooleanLiteralsSelectBranches(t *testing.T) {
	l := New("{$IF True} 1 {$ELSE} 2 {$ENDIF}\n{$IF False} 3 {$ELSE} 4 {$ENDIF}")
	if got := collectLiterals(l); !slices.Equal(got, []string{"1", "4"}) {
		t.Fatalf("branch tokens = %v, want true branch followed by false ELSE branch", got)
	}
}
