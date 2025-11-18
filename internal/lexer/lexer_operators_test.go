package lexer

import (
	"testing"
)

func TestOperators(t *testing.T) {
	input := `+ - * / % ^ **
		= <> < > <= >= == === !=
		:= += -= *= /= %= ^= @= ~=
		++ --
		<< >> | || & &&
		? ?? ?. =>
		@ ~ \ $ !`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"+", PLUS},
		{"-", MINUS},
		{"*", ASTERISK},
		{"/", SLASH},
		{"%", PERCENT},
		{"^", CARET},
		{"**", POWER},
		{"=", EQ},
		{"<>", NOT_EQ},
		{"<", LESS},
		{">", GREATER},
		{"<=", LESS_EQ},
		{">=", GREATER_EQ},
		{"==", EQ_EQ},
		{"===", EQ_EQ_EQ},
		{"!=", EXCL_EQ},
		{":=", ASSIGN},
		{"+=", PLUS_ASSIGN},
		{"-=", MINUS_ASSIGN},
		{"*=", TIMES_ASSIGN},
		{"/=", DIVIDE_ASSIGN},
		{"%=", PERCENT_ASSIGN},
		{"^=", CARET_ASSIGN},
		{"@=", AT_ASSIGN},
		{"~=", TILDE_ASSIGN},
		{"++", INC},
		{"--", DEC},
		{"<<", LESS_LESS},
		{">>", GREATER_GREATER},
		{"|", PIPE},
		{"||", PIPE_PIPE},
		{"&", AMP},
		{"&&", AMP_AMP},
		{"?", QUESTION},
		{"??", QUESTION_QUESTION},
		{"?.", QUESTION_DOT},
		{"=>", FAT_ARROW},
		{"@", AT},
		{"~", TILDE},
		{"\\", BACKSLASH},
		{"$", DOLLAR},
		{"!", EXCLAMATION},
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

// TestNotInOperator tests that "not in" is correctly tokenized as two separate keywords
// for use in expressions like "char not in [#0..#255]"
func TestNotInOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []struct {
			literal string
			typ     TokenType
		}
	}{
		{
			name:  "simple not in",
			input: "not in",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{"not", NOT},
				{"in", IN},
				{"", EOF},
			},
		},
		{
			name:  "variable not in set",
			input: "x not in mySet",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{"x", IDENT},
				{"not", NOT},
				{"in", IN},
				{"mySet", IDENT},
				{"", EOF},
			},
		},
		{
			name:  "char not in char range",
			input: "char not in [#0..#255]",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{"char", IDENT},
				{"not", NOT},
				{"in", IN},
				{"[", LBRACK},
				{"#0", CHAR},
				{"..", DOTDOT},
				{"#255", CHAR},
				{"]", RBRACK},
				{"", EOF},
			},
		},
		{
			name:  "if statement with not in",
			input: "if char not in [#0..#255] then exit;",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{"if", IF},
				{"char", IDENT},
				{"not", NOT},
				{"in", IN},
				{"[", LBRACK},
				{"#0", CHAR},
				{"..", DOTDOT},
				{"#255", CHAR},
				{"]", RBRACK},
				{"then", THEN},
				{"exit", EXIT},
				{";", SEMICOLON},
				{"", EOF},
			},
		},
		{
			name:  "enum not in set",
			input: "meOne not in s",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{"meOne", IDENT},
				{"not", NOT},
				{"in", IN},
				{"s", IDENT},
				{"", EOF},
			},
		},
		{
			name:  "NOT capitalized",
			input: "x NOT in mySet",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{"x", IDENT},
				{"NOT", NOT},
				{"in", IN},
				{"mySet", IDENT},
				{"", EOF},
			},
		},
		{
			name:  "IN capitalized",
			input: "x not IN mySet",
			want: []struct {
				literal string
				typ     TokenType
			}{
				{"x", IDENT},
				{"not", NOT},
				{"IN", IN},
				{"mySet", IDENT},
				{"", EOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			for i, expected := range tt.want {
				tok := l.NextToken()

				if tok.Type != expected.typ {
					t.Fatalf("token[%d] type wrong. expected=%q, got=%q (literal=%q)",
						i, expected.typ, tok.Type, tok.Literal)
				}

				if tok.Literal != expected.literal {
					t.Fatalf("token[%d] literal wrong. expected=%q, got=%q",
						i, expected.literal, tok.Literal)
				}
			}
		})
	}
}
