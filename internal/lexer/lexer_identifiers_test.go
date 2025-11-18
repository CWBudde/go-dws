package lexer

import (
	"testing"
)

func TestIdentifiers(t *testing.T) {
	input := `myVar MyClass _private x1 test123 _`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"myVar", IDENT},
		{"MyClass", IDENT},
		{"_private", IDENT},
		{"x1", IDENT},
		{"test123", IDENT},
		{"_", IDENT},
		{"", EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestUnicodeIdentifiers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []struct {
			literal string
			typ     TokenType
		}
	}{
		{
			name:  "Greek letter Delta",
			input: "var Δ : Integer;",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{"var", VAR},
				{"Δ", IDENT},
				{":", COLON},
				{"Integer", IDENT},
				{";", SEMICOLON},
				{"", EOF},
			},
		},
		{
			name:  "Greek letters alpha and beta",
			input: "α := β + 1;",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{"α", IDENT},
				{":=", ASSIGN},
				{"β", IDENT},
				{"+", PLUS},
				{"1", INT},
				{";", SEMICOLON},
				{"", EOF},
			},
		},
		{
			name:  "Cyrillic variable names",
			input: "var переменная : Integer;",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{literal: "var", typ: VAR},
				{literal: "переменная", typ: IDENT},
				{literal: ":", typ: COLON},
				{literal: "Integer", typ: IDENT},
				{literal: ";", typ: SEMICOLON},
				{literal: "", typ: EOF},
			},
		},
		{
			name:  "Chinese characters",
			input: "var 变量 := 42;",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{literal: "var", typ: VAR},
				{literal: "变量", typ: IDENT},
				{literal: ":=", typ: ASSIGN},
				{literal: "42", typ: INT},
				{literal: ";", typ: SEMICOLON},
				{literal: "", typ: EOF},
			},
		},
		{
			name:  "Japanese hiragana and katakana",
			input: "var へんすう := カタカナ;",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{literal: "var", typ: VAR},
				{literal: "へんすう", typ: IDENT},
				{literal: ":=", typ: ASSIGN},
				{literal: "カタカナ", typ: IDENT},
				{literal: ";", typ: SEMICOLON},
				{literal: "", typ: EOF},
			},
		},
		{
			name:  "Mixed ASCII and Unicode",
			input: "var myΔValue := 100;",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{literal: "var", typ: VAR},
				{literal: "myΔValue", typ: IDENT},
				{literal: ":=", typ: ASSIGN},
				{literal: "100", typ: INT},
				{literal: ";", typ: SEMICOLON},
				{literal: "", typ: EOF},
			},
		},
		{
			name:  "Underscore with Unicode",
			input: "var test_Δ := 42;",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{literal: "var", typ: VAR},
				{literal: "test_Δ", typ: IDENT},
				{literal: ":=", typ: ASSIGN},
				{literal: "42", typ: INT},
				{literal: ";", typ: SEMICOLON},
				{literal: "", typ: EOF},
			},
		},
		{
			name:  "Unicode in function call",
			input: "PrintLn(Δ);",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{literal: "PrintLn", typ: IDENT},
				{literal: "(", typ: LPAREN},
				{literal: "Δ", typ: IDENT},
				{literal: ")", typ: RPAREN},
				{literal: ";", typ: SEMICOLON},
				{literal: "", typ: EOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			for i, expected := range tt.want {
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
		})
	}
}
