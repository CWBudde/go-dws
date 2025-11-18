package lexer

import (
	"testing"
)

func TestUnicodeInStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Greek in string",
			input:    "'Δημοκρατία'",
			expected: "Δημοκρατία",
		},
		{
			name:     "Chinese in string",
			input:    "'你好世界'",
			expected: "你好世界",
		},
		{
			name:     "Mixed Unicode in string",
			input:    "'Hello Δ 世界'",
			expected: "Hello Δ 世界",
		},
		{
			name:     "Emoji in string",
			input:    "'Test 🚀 String'",
			expected: "Test 🚀 String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != STRING {
				t.Errorf("wrong token type. expected=STRING, got=%q", tok.Type)
			}

			if tok.Literal != tt.expected {
				t.Errorf("wrong string literal. expected=%q, got=%q", tt.expected, tok.Literal)
			}
		})
	}
}

func TestRosettaUnicodeExample(t *testing.T) {
	// This is the exact code from examples/rosetta/Unicode_variable_names.dws
	input := `var Δ : Integer;

Δ := 1;
Inc(Δ);
PrintLn(Δ);`

	expectedTokens := []struct {
		literal string
		typ     TokenType
	}{
		{literal: "var", typ: VAR},
		{literal: "Δ", typ: IDENT},
		{literal: ":", typ: COLON},
		{literal: "Integer", typ: IDENT},
		{literal: ";", typ: SEMICOLON},
		{literal: "Δ", typ: IDENT},
		{literal: ":=", typ: ASSIGN},
		{literal: "1", typ: INT},
		{literal: ";", typ: SEMICOLON},
		{literal: "Inc", typ: IDENT},
		{literal: "(", typ: LPAREN},
		{literal: "Δ", typ: IDENT},
		{literal: ")", typ: RPAREN},
		{literal: ";", typ: SEMICOLON},
		{literal: "PrintLn", typ: IDENT},
		{literal: "(", typ: LPAREN},
		{literal: "Δ", typ: IDENT},
		{literal: ")", typ: RPAREN},
		{literal: ";", typ: SEMICOLON},
		{literal: "", typ: EOF},
	}

	l := New(input)

	for i, expected := range expectedTokens {
		tok := l.NextToken()

		if tok.Type != expected.typ {
			t.Errorf("token[%d] - wrong type. expected=%q, got=%q (literal=%q)",
				i, expected.typ, tok.Type, tok.Literal)
		}

		if tok.Literal != expected.literal {
			t.Errorf("token[%d] - wrong literal. expected=%q, got=%q",
				i, expected.literal, tok.Literal)
		}
	}
}
