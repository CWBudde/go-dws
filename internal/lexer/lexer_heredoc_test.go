package lexer

import (
	"testing"
)

// TestTripleQuoteStrings tests heredoc-style (triple-quoted) string literals.
func TestTripleQuoteStrings(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
	}{
		{
			name:            "basic heredoc with matching indent",
			input:           "'''\n   hello world\n   '''",
			expectedLiteral: "hello world",
		},
		{
			name:            "no indent",
			input:           "'''\nno indent\n'''",
			expectedLiteral: "no indent",
		},
		{
			name:            "deeper content keeps extra indent",
			input:           "'''\n      three indents\n   '''",
			expectedLiteral: "   three indents",
		},
		{
			name:            "empty lines preserved, tab indent",
			input:           "'''\n\n\t\tempty lines\n\n\t\t'''",
			expectedLiteral: "\nempty lines\n",
		},
		{
			name:            "quotes inside heredoc are verbatim",
			input:           "'''\n   l'appel du \"lyon\" ou ''du loup''\n   '''",
			expectedLiteral: "l'appel du \"lyon\" ou ''du loup''",
		},
		{
			name:            "multiple content lines",
			input:           "'''\n  line one\n  line two\n  '''",
			expectedLiteral: "line one\nline two",
		},
		{
			name:            "CRLF line endings",
			input:           "'''\r\n   hello\r\n   '''",
			expectedLiteral: "hello",
		},
		{
			name:            "triple double quotes",
			input:           "\"\"\"\n   hello\n   \"\"\"",
			expectedLiteral: "hello",
		},
		{
			name:            "empty heredoc",
			input:           "'''\n\n'''",
			expectedLiteral: "",
		},
		{
			// UTokenizerTests.TripleApos: the indentation is matched column by column.
			name:            "tab indent keeps extra tabs",
			input:           "'''\n\t\tworld\t\n\t'''",
			expectedLiteral: "\tworld\t",
		},
		{
			name:            "leading and trailing empty content lines",
			input:           "'''\n\n      space\n\n   '''",
			expectedLiteral: "\n   space\n",
		},
		{
			name:            "whitespace-only line shorter than the indent",
			input:           "'''\n    a\n  \n    b\n    '''",
			expectedLiteral: "a\n\nb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != STRING {
				t.Fatalf("tokentype wrong. expected=%q, got=%q", STRING, tok.Type)
			}
			if tok.Literal != tt.expectedLiteral {
				t.Fatalf("literal wrong. expected=%q, got=%q", tt.expectedLiteral, tok.Literal)
			}
			if errs := l.Errors(); len(errs) != 0 {
				t.Fatalf("unexpected lexer errors: %v", errs)
			}
		})
	}
}

// TestTripleQuoteNotHeredoc verifies that quote sequences that are not
// heredoc openers keep their regular string semantics.
func TestTripleQuoteNotHeredoc(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
	}{
		{
			name:            "four quotes is an escaped quote",
			input:           "''''",
			expectedLiteral: "'",
		},
		{
			name:            "six quotes is two escaped quotes",
			input:           "''''''",
			expectedLiteral: "''",
		},
		{
			name:            "empty string",
			input:           "''",
			expectedLiteral: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != STRING {
				t.Fatalf("tokentype wrong. expected=%q, got=%q", STRING, tok.Type)
			}
			if tok.Literal != tt.expectedLiteral {
				t.Fatalf("literal wrong. expected=%q, got=%q", tt.expectedLiteral, tok.Literal)
			}
		})
	}
}

// TestTripleQuoteErrors tests the DWScript diagnostics for malformed triple
// apostrophe strings. They are surfaced as directive-channel diagnostics; only an
// unterminated literal is a compiler stop, the others let compilation continue.
func TestTripleQuoteErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		message string
		line    int
		column  int
		fatal   bool
	}{
		{
			name:    "unterminated heredoc",
			input:   "'''\n   hello",
			message: "End of string constant not found (end of line)",
			line:    1,
			column:  1,
			fatal:   true,
		},
		{
			name:    "single-quoted string spanning lines without triple terminator",
			input:   "x := 'hello\nworld';",
			message: "End of string constant not found (end of line)",
			line:    1,
			column:  6,
			fatal:   true,
		},
		{
			name:    "terminator preceded by non-whitespace",
			input:   "'''\n   hello x'''",
			message: "Incorrect triple apostrophe string indentation",
			line:    1,
			column:  1,
		},
		{
			name:    "no content line",
			input:   "'''\n'''",
			message: "Incorrect triple apostrophe string",
			line:    1,
			column:  1,
		},
		{
			name:    "content indented less than the terminator",
			input:   "var s := '''\n   qsd\n     ''';",
			message: "Incorrect triple apostrophe string indentation",
			line:    1,
			column:  10,
		},
		{
			name:    "triple apostrophe without line break reaching end of line",
			input:   "x := ''' qsdq \n''';",
			message: "Incorrect triple apostrophe string",
			line:    1,
			column:  6,
		},
		{
			// DWScript buffers only the decoded content (the opening quote is not
			// kept), so a plain literal crossing a line break is not a triple string.
			name:    "plain single-quoted literal crossing a line break",
			input:   "x := 'a''b\n  c\n  ''';",
			message: "Incorrect triple apostrophe string",
			line:    1,
			column:  6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			drainTokens(l)
			diags := l.DirectiveDiagnostics()
			if len(diags) != 1 {
				t.Fatalf("expected one diagnostic, got %v", diags)
			}
			d := diags[0]
			if d.Message != tt.message || d.Pos.Line != tt.line || d.Pos.Column != tt.column {
				t.Fatalf("got %q at %d:%d, want %q at %d:%d",
					d.Message, d.Pos.Line, d.Pos.Column, tt.message, tt.line, tt.column)
			}
			if l.StoppedByFatal() != tt.fatal {
				t.Fatalf("StoppedByFatal() = %v, want %v", l.StoppedByFatal(), tt.fatal)
			}
		})
	}
}

// TestIndentedStrings covers #'...' and #"..." strings, which may span lines and lose
// their common indentation as in DWScript's TTokenBuffer.AppendMultiToStr
// (SimpleScripts/heredoc_indent, heredoc_special).
func TestIndentedStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "single line keeps no leading blanks", input: `#"  !Second"`, want: "!Second"},
		{name: "blank first line is dropped", input: "#\"\n   First\n     Second\n   Third\"", want: "First\n  Second\nThird"},
		{name: "last line does not count toward the indent", input: "#'\n    a\n  b'", want: "a\nb"},
		{name: "doubled quotes", input: "#'\n  ''x''\n  \"\"y\"\"\n'", want: "'x'\n\"\"y\"\"\n"},
		{name: "first line with content", input: "#\" First\n   Second\n Third\"", want: "First\n  Second\nThird"},
		{name: "concatenated with char literals", input: `#"a"#13#10#"  b"`, want: "a\r\nb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			if tok.Type != STRING || tok.Literal != tt.want {
				t.Fatalf("got %s %q, want %q", tok.Type, tok.Literal, tt.want)
			}
			if errs := l.Errors(); len(errs) != 0 {
				t.Fatalf("unexpected lexer errors: %v", errs)
			}
		})
	}
}

// TestIndentedStringUnterminated checks that an open #'...' string reports DWScript's
// end-of-file error past the terminated last line.
func TestIndentedStringUnterminated(t *testing.T) {
	l := New("x := #'\n  abc")
	drainTokens(l)
	diags := l.DirectiveDiagnostics()
	if len(diags) != 1 || diags[0].Message != "End of string constant not found (end of file)" ||
		diags[0].Pos.Line != 4 || diags[0].Pos.Column != 1 || !l.StoppedByFatal() {
		t.Fatalf("got %v", diags)
	}
}
