package lexer

import (
	"testing"
)

func TestComplexExpression(t *testing.T) {
	input := `result := (x + y) * 2 - z / 3.5;`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"result", IDENT},
		{":=", ASSIGN},
		{"(", LPAREN},
		{"x", IDENT},
		{"+", PLUS},
		{"y", IDENT},
		{")", RPAREN},
		{"*", ASTERISK},
		{"2", INT},
		{"-", MINUS},
		{"z", IDENT},
		{"/", SLASH},
		{"3.5", FLOAT},
		{";", SEMICOLON},
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

func TestFunctionDeclaration(t *testing.T) {
	input := `function Add(a, b: Integer): Integer;
begin
	Result := a + b;
end;`

	tests := []struct {
		expectedType TokenType
	}{
		{FUNCTION},
		{IDENT},     // Add
		{LPAREN},    // (
		{IDENT},     // a
		{COMMA},     // ,
		{IDENT},     // b
		{COLON},     // :
		{IDENT},     // Integer
		{RPAREN},    // )
		{COLON},     // :
		{IDENT},     // Integer
		{SEMICOLON}, // ;
		{BEGIN},
		{IDENT},     // Result
		{ASSIGN},    // :=
		{IDENT},     // a
		{PLUS},      // +
		{IDENT},     // b
		{SEMICOLON}, // ;
		{END},
		{SEMICOLON}, // ;
		{EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
	}
}

func TestClassDeclaration(t *testing.T) {
	input := `type TPoint = class
	private
		FX: Integer;
	public
		property X: Integer read FX;
	end;`

	tests := []struct {
		expectedType TokenType
	}{
		{TYPE},
		{IDENT},     // TPoint
		{EQ},        // =
		{CLASS},     // class
		{PRIVATE},   // private
		{IDENT},     // FX
		{COLON},     // :
		{IDENT},     // Integer
		{SEMICOLON}, // ;
		{PUBLIC},    // public
		{PROPERTY},  // property
		{IDENT},     // X
		{COLON},     // :
		{IDENT},     // Integer
		{READ},      // read
		{IDENT},     // FX
		{SEMICOLON}, // ;
		{END},       // end
		{SEMICOLON}, // ;
		{EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
	}
}

// TestNewKeyword tests that "new" is recognized as a keyword
func TestNewKeyword(t *testing.T) {
	input := `new`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"new", NEW},
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

// TestLambdaExpressions tests tokenization of lambda expressions
// Add lambda keyword to lexer
func TestLambdaExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			literal string
			typ     TokenType
		}
	}{
		{
			name:  "simple lambda keyword",
			input: "lambda",
			expected: []struct {
				literal string
				typ     TokenType
			}{
				{"lambda", LAMBDA},
				{"", EOF},
			},
		},
		{
			name:  "lambda with arrow operator",
			input: "lambda(x) => x * 2",
			expected: []struct {
				literal string
				typ     TokenType
			}{
				{"lambda", LAMBDA},
				{"(", LPAREN},
				{"x", IDENT},
				{")", RPAREN},
				{"=>", FAT_ARROW},
				{"x", IDENT},
				{"*", ASTERISK},
				{"2", INT},
				{"", EOF},
			},
		},
		{
			name:  "lambda with typed parameters",
			input: "lambda(x: Integer): Integer",
			expected: []struct {
				literal string
				typ     TokenType
			}{
				{"lambda", LAMBDA},
				{"(", LPAREN},
				{"x", IDENT},
				{":", COLON},
				{"Integer", IDENT},
				{")", RPAREN},
				{":", COLON},
				{"Integer", IDENT},
				{"", EOF},
			},
		},
		{
			name:  "lambda with begin/end body",
			input: "lambda(x: Integer): Integer begin Result := x * 2; end",
			expected: []struct {
				literal string
				typ     TokenType
			}{
				{"lambda", LAMBDA},
				{"(", LPAREN},
				{"x", IDENT},
				{":", COLON},
				{"Integer", IDENT},
				{")", RPAREN},
				{":", COLON},
				{"Integer", IDENT},
				{"begin", BEGIN},
				{"Result", IDENT},
				{":=", ASSIGN},
				{"x", IDENT},
				{"*", ASTERISK},
				{"2", INT},
				{";", SEMICOLON},
				{"end", END},
				{"", EOF},
			},
		},
		{
			name:  "lambda with multiple parameters",
			input: "lambda(a, b: Integer, c: Float)",
			expected: []struct {
				literal string
				typ     TokenType
			}{
				{"lambda", LAMBDA},
				{"(", LPAREN},
				{"a", IDENT},
				{",", COMMA},
				{"b", IDENT},
				{":", COLON},
				{"Integer", IDENT},
				{",", COMMA},
				{"c", IDENT},
				{":", COLON},
				{"Float", IDENT},
				{")", RPAREN},
				{"", EOF},
			},
		},
		{
			name:  "lambda as part of function call",
			input: "Map(arr, lambda(x) => x * 2)",
			expected: []struct {
				literal string
				typ     TokenType
			}{
				{"Map", IDENT},
				{"(", LPAREN},
				{"arr", IDENT},
				{",", COMMA},
				{"lambda", LAMBDA},
				{"(", LPAREN},
				{"x", IDENT},
				{")", RPAREN},
				{"=>", FAT_ARROW},
				{"x", IDENT},
				{"*", ASTERISK},
				{"2", INT},
				{")", RPAREN},
				{"", EOF},
			},
		},
		{
			name:  "lambda with no parameters",
			input: "lambda() => 42",
			expected: []struct {
				literal string
				typ     TokenType
			}{
				{"lambda", LAMBDA},
				{"(", LPAREN},
				{")", RPAREN},
				{"=>", FAT_ARROW},
				{"42", INT},
				{"", EOF},
			},
		},
		{
			name:  "case insensitive lambda",
			input: "LAMBDA(x) => X",
			expected: []struct {
				literal string
				typ     TokenType
			}{
				{"LAMBDA", LAMBDA},
				{"(", LPAREN},
				{"x", IDENT},
				{")", RPAREN},
				{"=>", FAT_ARROW},
				{"X", IDENT},
				{"", EOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			for i, expectedToken := range tt.expected {
				tok := l.NextToken()

				if tok.Type != expectedToken.typ {
					t.Fatalf("test[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
						i, expectedToken.typ, tok.Type, tok.Literal)
				}

				if tok.Literal != expectedToken.literal {
					t.Fatalf("test[%d] - literal wrong. expected=%q, got=%q",
						i, expectedToken.literal, tok.Literal)
				}
			}
		})
	}
}
