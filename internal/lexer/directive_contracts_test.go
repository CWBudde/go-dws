package lexer

import (
	"strconv"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/pkg/token"
)

// Synthesized include literals retain the directive's position and hint policy;
// surrounding source tokens must neither disappear nor be emitted twice.
func TestIncludeMacro_TokenContract(t *testing.T) {
	l := New("PrintLn({$i%LiNe%});\n100 + {$INCLUDE %LINENUM%}")
	want := []struct {
		literal      string
		kind         TokenType
		line, column int
	}{
		{"PrintLn", IDENT, 1, 1}, {"(", LPAREN, 1, 8},
		{"1", STRING, 1, 11}, {")", RPAREN, 1, 19}, {";", SEMICOLON, 1, 20},
		{"100", INT, 2, 1}, {"+", PLUS, 2, 5}, {"2", INT, 2, 9},
	}
	for _, expected := range want {
		assertDirectiveToken(t, l.NextToken(), expected.kind, expected.literal, expected.line, expected.column)
	}
	if got := l.NextToken(); got.Type != EOF || len(l.Errors()) != 0 {
		t.Fatalf("end = %v, diagnostics = %v", got, l.Errors())
	}

	t.Run("multiline argument uses argument line", func(t *testing.T) {
		l := New("\n  {$I\n %LINE%} + 7")
		got := l.NextToken()
		if got.Type != STRING || got.Literal != "3" || got.Pos.Line != 2 || got.Pos.Column != 5 {
			t.Fatalf("multiline substitution = %v, want STRING 3 at directive 2:5", got)
		}
		if got := l.NextToken(); got.Type != PLUS || got.Pos.Line != 3 || got.Pos.Column != 10 {
			t.Fatalf("following token = %v, want PLUS at 3:10", got)
		}
	})
}

func TestIncludeMacro_ClockTokenContract(t *testing.T) {
	for _, name := range []string{"TIME", "DATE", "TIMESTAMP"} {
		t.Run(name, func(t *testing.T) {
			before := time.Now()
			l := New("{$HINTS PEDANTIC}\n  {$I %" + name + "%} Tail")
			got := l.NextToken()
			after := time.Now()
			wantKind := assertIncludeClockValue(t, name, got.Literal, before, after)
			if got.Type != wantKind || got.Pos.Line != 2 || got.Pos.Column != 5 || got.Pos.Hints != token.HintLevelPedantic || got.End().Hints != token.HintLevelPedantic {
				t.Fatalf("substitution metadata = %v, hints = %v; want %v at 2:5 with pedantic hints", got, got.Pos.Hints, wantKind)
			}
			assertDirectiveToken(t, l.NextToken(), IDENT, "Tail", 2, len(name)+11)
			if next := l.NextToken(); next.Type != EOF || len(l.Errors()) != 0 {
				t.Fatalf("end = %v, diagnostics = %v", next, l.Errors())
			}
		})
	}
}

func assertIncludeClockValue(t *testing.T, name, literal string, before, after time.Time) TokenType {
	t.Helper()
	switch name {
	case "TIMESTAMP":
		stamp, err := strconv.ParseInt(literal, 10, 64)
		if err != nil || stamp < before.Unix() || stamp > after.Unix() {
			t.Fatalf("timestamp = %q, err = %v; want current Unix seconds", literal, err)
		}
		return INT
	case "DATE":
		if literal != before.Format("2006-01-02") && literal != after.Format("2006-01-02") {
			t.Fatalf("date = %q, want current local calendar date", literal)
		}
	case "TIME":
		if parsed, err := time.Parse("15:04:05", literal); err != nil || parsed.Format("15:04:05") != literal {
			t.Fatalf("time = %q, err = %v; want a valid hh:mm:ss string", literal, err)
		}
	default:
		t.Fatalf("no clock contract for %q", name)
	}
	return STRING
}

func assertDirectiveToken(t *testing.T, got Token, kind TokenType, literal string, line, column int) {
	t.Helper()
	if got.Type != kind || got.Literal != literal || got.Pos.Line != line || got.Pos.Column != column {
		t.Fatalf("token = %v, want %v %q at %d:%d", got, kind, literal, line, column)
	}
}

func TestIncludeMacro_RecoveryTokenContract(t *testing.T) {
	tests := []struct {
		name, source string
		messages     []string
		columns      []int
		stopped      bool
	}{
		{"unknown substitutes empty string", "{$I %DUMMY%} Tail", []string{`Include item "DUMMY" unknown`}, []int{5}, false},
		{"missing percent substitutes empty string", "{$I %LINE} Tail", []string{"Include item expected"}, []int{5}, false},
		{"missing name substitutes empty string", "{$I %} Tail", []string{"Include item expected"}, []int{5}, false},
		{"failed percent retains inspected token", "{$I %LINE extra} Tail", []string{"Include item expected", `"}" expected`}, []int{5, 11}, true},
		{"reserved name remains current", "{$I %END%} Tail", []string{"Include item expected", `"}" expected`}, []int{5, 6}, true},
		{"numeric name remains current", "{$I %42%} Tail", []string{"Include item expected", `"}" expected`}, []int{5, 6}, true},
		{"EOF after final percent", "{$I %LINE%", []string{`"}" expected`}, []int{10}, true},
		{"EOF after opening percent", "{$I %", []string{"Include item expected", `"}" expected`}, []int{5, 5}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)
			got := l.NextToken()
			if tt.stopped {
				if got.Type != EOF {
					t.Fatalf("stopped macro emitted %v instead of EOF", got)
				}
			} else {
				if got.Type != STRING || got.Literal != "" || got.Pos.Line != 1 || got.Pos.Column != 3 {
					t.Fatalf("recovery = %v, want empty STRING at 1:3", got)
				}
				if next := l.NextToken(); next.Type != IDENT || next.Literal != "Tail" {
					t.Fatalf("recovery lost following source token: %v", next)
				}
			}
			if l.StoppedByFatal() != tt.stopped {
				t.Fatalf("compiler-stop state = %v, want %v", l.StoppedByFatal(), tt.stopped)
			}
			assertDirectiveContract(t, l, tt.messages, 1, tt.columns)
			for i := 0; i < 2; i++ {
				if end := l.NextToken(); end.Type != EOF {
					t.Fatalf("end %d = %v", i, end)
				}
			}
			assertDirectiveContract(t, l, tt.messages, 1, tt.columns)
		})
	}
}

// A diagnostic's origin and severity are part of the front-end contract, not
// merely its printed text: these syntax errors must stay directive-owned.
func assertDirectiveContract(t *testing.T, l *Lexer, messages []string, line int, columns []int) {
	t.Helper()
	diags := l.DirectiveDiagnostics()
	if len(diags) != len(messages) || len(l.Errors()) != len(messages) || len(l.IncludeErrors()) != 0 {
		t.Fatalf("directive diagnostics = %v, all = %v, include = %v; want %v", diags, l.Errors(), l.IncludeErrors(), messages)
	}
	for i, diag := range diags {
		if diag.Message != messages[i] || diag.Pos.Line != line || diag.Pos.Column != columns[i] || diag.Severity != SeverityError || !diag.Directive || diag.Constant || diag.Rendered != "" {
			t.Fatalf("diagnostic %d = %+v, want directive syntax error %q at %d:%d", i, diag, messages[i], line, columns[i])
		}
	}
}

func TestConditional_InactiveTokenAndEOFContract(t *testing.T) {
	tests := []struct {
		name, source string
		message      string
		line, column int
	}{
		{"comments are not skipped tokens", "{$ifdef test}\n{ comment } (* another *) // line\n/* final */", "Unbalanced conditional directive", 1, 3},
		{"quoted directive text is a skipped token", "{$ifdef test}\n'{$endif}'", "Unbalanced conditional directive", 1, 9},
		{"double quoted directive text", "{$ifdef test}\n\"{$endif}\"", "Unbalanced conditional directive", 1, 9},
		{"unknown switch does not replace provenance", "{$ifdef test}\n{$NOP}\n42", "Unbalanced conditional directive", 1, 9},
		{"known nested switch owns later skipped token", "{$ifdef test}\n{$ifdef inner}\n{$endif}\n42", "Unbalanced conditional directive", 3, 3},
		{"inactive macro requires brace only", "{$ifdef test}\n{$I %LINE%", `"}" expected`, 2, 10},
		{"inactive unknown macro is not validated", "{$ifdef test}\n{$I %DUMMY%", `"}" expected`, 2, 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)
			if got := l.NextToken(); got.Type != EOF || !l.StoppedByFatal() {
				t.Fatalf("inactive EOF = %v, stop = %v", got, l.StoppedByFatal())
			}
			assertDirectiveContract(t, l, []string{tt.message}, tt.line, []int{tt.column})
			l.NextToken()
			assertDirectiveContract(t, l, []string{tt.message}, tt.line, []int{tt.column})
		})
	}
	t.Run("completed inactive macro emits nothing", func(t *testing.T) {
		l := New("{$ifdef absent}\n{$I %DUMMY%}\n{$endif}\nTail")
		if got := l.NextToken(); got.Type != IDENT || got.Literal != "Tail" || got.Pos.Line != 4 || got.Pos.Column != 1 {
			t.Fatalf("inactive macro leaked or lost tokens: %v", got)
		}
		if got := l.NextToken(); got.Type != EOF || l.StoppedByFatal() || len(l.Errors()) != 0 {
			t.Fatalf("end = %v, stop = %v, diagnostics = %v", got, l.StoppedByFatal(), l.Errors())
		}
	})
}

func TestIncludeMacro_BufferedTokenSnapshot(t *testing.T) {
	l := New("{$I %LINENUM%} + 7")
	// Public lookahead leaves a synthesized token pending in the buffer. A later
	// speculative read must not corrupt the saved prefix or duplicate delivery.
	if got := l.Peek(0); got.Type != INT || got.Literal != "1" {
		t.Fatalf("lookahead = %v, want substituted integer", got)
	}
	saved := l.SaveState()
	if got := l.Peek(1); got.Type != PLUS {
		t.Fatalf("speculative following token = %v, want PLUS", got)
	}
	for attempt := 0; attempt < 2; attempt++ {
		l.RestoreState(saved)
		assertDirectiveToken(t, l.NextToken(), INT, "1", 1, 3)
		assertDirectiveToken(t, l.NextToken(), PLUS, "+", 1, 16)
		assertDirectiveToken(t, l.NextToken(), INT, "7", 1, 18)
		if got := l.NextToken(); got.Type != EOF || len(l.Errors()) != 0 {
			t.Fatalf("attempt %d end = %v, diagnostics = %v", attempt, got, l.Errors())
		}
	}
}
