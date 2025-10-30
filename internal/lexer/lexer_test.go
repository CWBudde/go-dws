package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `var x := 5;
	x := x + 10;
	`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"var", VAR},
		{"x", IDENT},
		{":=", ASSIGN},
		{"5", INT},
		{";", SEMICOLON},
		{"x", IDENT},
		{":=", ASSIGN},
		{"x", IDENT},
		{"+", PLUS},
		{"10", INT},
		{";", SEMICOLON},
		{"", EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := `begin end if then else while for do
		function procedure class var const
		true false nil and or not xor
		try except finally raise
		div mod shl shr`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"begin", BEGIN},
		{"end", END},
		{"if", IF},
		{"then", THEN},
		{"else", ELSE},
		{"while", WHILE},
		{"for", FOR},
		{"do", DO},
		{"function", FUNCTION},
		{"procedure", PROCEDURE},
		{"class", CLASS},
		{"var", VAR},
		{"const", CONST},
		{"true", TRUE},
		{"false", FALSE},
		{"nil", NIL},
		{"and", AND},
		{"or", OR},
		{"not", NOT},
		{"xor", XOR},
		{"try", TRY},
		{"except", EXCEPT},
		{"finally", FINALLY},
		{"raise", RAISE},
		{"div", DIV},
		{"mod", MOD},
		{"shl", SHL},
		{"shr", SHR},
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

func TestCaseInsensitiveKeywords(t *testing.T) {
	input := `BEGIN End IF Then WHILE WhILe`

	tests := []struct {
		expectedType TokenType
	}{
		{BEGIN},
		{END},
		{IF},
		{THEN},
		{WHILE},
		{WHILE},
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

func TestDelimiters(t *testing.T) {
	input := `( ) [ ] ; , . : ..`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"(", LPAREN},
		{")", RPAREN},
		{"[", LBRACK},
		{"]", RBRACK},
		{";", SEMICOLON},
		{",", COMMA},
		{".", DOT},
		{":", COLON},
		{"..", DOTDOT},
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

func TestIntegerLiterals(t *testing.T) {
	input := `123 0 $FF $ff $10 0xFF 0x10 %1010 %0`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"123", INT},
		{"0", INT},
		{"$FF", INT},
		{"$ff", INT},
		{"$10", INT},
		{"0xFF", INT},
		{"0x10", INT},
		{"%1010", INT},
		{"%0", INT},
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

func TestFloatLiterals(t *testing.T) {
	input := `123.45 0.5 3.14 1.5e10 1.5E10 1.5e-5 2.0E+3`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"123.45", FLOAT},
		{"0.5", FLOAT},
		{"3.14", FLOAT},
		{"1.5e10", FLOAT},
		{"1.5E10", FLOAT},
		{"1.5e-5", FLOAT},
		{"2.0E+3", FLOAT},
		{"", EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
		expectedType    TokenType
	}{
		{
			name:            "simple single quoted",
			input:           `'hello'`,
			expectedType:    STRING,
			expectedLiteral: "hello",
		},
		{
			name:            "simple double quoted",
			input:           `"world"`,
			expectedType:    STRING,
			expectedLiteral: "world",
		},
		{
			name:            "escaped single quote",
			input:           `'it''s'`,
			expectedType:    STRING,
			expectedLiteral: "it's",
		},
		{
			name:            "empty string",
			input:           `''`,
			expectedType:    STRING,
			expectedLiteral: "",
		},
		{
			name:            "string with spaces",
			input:           `'hello world'`,
			expectedType:    STRING,
			expectedLiteral: "hello world",
		},
		{
			name:            "multiline string",
			input:           "'hello\nworld'",
			expectedType:    STRING,
			expectedLiteral: "hello\nworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.expectedType {
				t.Fatalf("tokentype wrong. expected=%q, got=%q",
					tt.expectedType, tok.Type)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Fatalf("literal wrong. expected=%q, got=%q",
					tt.expectedLiteral, tok.Literal)
			}
		})
	}
}

func TestCharLiterals(t *testing.T) {
	input := `#65 #$41 #13 #10`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"#65", CHAR},
		{"#$41", CHAR},
		{"#13", CHAR},
		{"#10", CHAR},
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

func TestComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wants []TokenType
	}{
		{
			name: "line comment",
			input: `x // this is a comment
y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name:  "block comment with braces",
			input: `x { comment } y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name:  "block comment with parens",
			input: `x (* comment *) y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name:  "C-style comment",
			input: `x /* comment */ y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name: "multiline C-style comment",
			input: `x /* this is a
				multiline comment */ y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name: "multiline block comment",
			input: `x {
				this is a
				multiline comment
			} y`,
			wants: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name: "multiple comments",
			input: `x // comment 1
y { comment 2 } z`,
			wants: []TokenType{IDENT, IDENT, IDENT, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			for i, expectedType := range tt.wants {
				tok := l.NextToken()
				if tok.Type != expectedType {
					t.Fatalf("token[%d] wrong. expected=%q, got=%q (literal=%q)",
						i, expectedType, tok.Type, tok.Literal)
				}
			}
		})
	}
}

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

func TestSimpleProgram(t *testing.T) {
	input := `
	var x: Integer;
	begin
		x := 10;
		if x > 5 then
			PrintLn('x is greater than 5');
	end;
	`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"var", VAR},
		{"x", IDENT},
		{":", COLON},
		{"Integer", IDENT},
		{";", SEMICOLON},
		{"begin", BEGIN},
		{"x", IDENT},
		{":=", ASSIGN},
		{"10", INT},
		{";", SEMICOLON},
		{"if", IF},
		{"x", IDENT},
		{">", GREATER},
		{"5", INT},
		{"then", THEN},
		{"PrintLn", IDENT},
		{"(", LPAREN},
		{"x is greater than 5", STRING},
		{")", RPAREN},
		{";", SEMICOLON},
		{"end", END},
		{";", SEMICOLON},
		{"", EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestPositionTracking(t *testing.T) {
	input := `var x
y`

	tests := []struct {
		expectedType TokenType
		expectedLine int
		expectedCol  int
	}{
		{VAR, 1, 1},
		{IDENT, 1, 5},
		{IDENT, 2, 1},
		{EOF, 2, 2},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Pos.Line != tt.expectedLine {
			t.Fatalf("tests[%d] - line wrong. expected=%d, got=%d",
				i, tt.expectedLine, tok.Pos.Line)
		}

		if tok.Pos.Column != tt.expectedCol {
			t.Fatalf("tests[%d] - column wrong. expected=%d, got=%d",
				i, tt.expectedCol, tok.Pos.Column)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType TokenType
	}{
		{
			name:         "empty input",
			input:        "",
			expectedType: EOF,
		},
		{
			name:         "only whitespace",
			input:        "   \t\n  ",
			expectedType: EOF,
		},
		{
			name:         "illegal character",
			input:        "`",
			expectedType: ILLEGAL,
		},
		{
			name:         "unterminated string",
			input:        "'hello",
			expectedType: ILLEGAL,
		},
		{
			name:         "unterminated block comment",
			input:        "{ comment",
			expectedType: ILLEGAL,
		},
		{
			name:         "unterminated C-style comment",
			input:        "/* comment",
			expectedType: ILLEGAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.expectedType {
				t.Fatalf("tokentype wrong. expected=%q, got=%q (literal=%q)",
					tt.expectedType, tok.Type, tok.Literal)
			}
		})
	}
}

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

func TestAllKeywords(t *testing.T) {
	// Test all keywords to ensure none are missed
	keywords := []string{
		"begin", "end", "if", "then", "else", "case", "of",
		"while", "repeat", "until", "for", "to", "downto", "do",
		"break", "continue", "exit", "with", "asm",
		"var", "const", "type", "record", "array", "set", "enum", "flags",
		"resourcestring", "namespace", "unit", "uses", "program", "library",
		"implementation", "initialization", "finalization",
		"class", "object", "interface", "implements",
		"function", "procedure", "constructor", "destructor", "method",
		"property", "virtual", "override", "abstract", "sealed", "static", "final",
		"new", "inherited", "reintroduce", "operator", "helper", "partial", "lazy", "index",
		"try", "except", "raise", "finally", "on",
		"not", "and", "or", "xor",
		"true", "false", "nil",
		"is", "as", "in", "div", "mod", "shl", "shr", "sar", "impl",
		"inline", "external", "forward", "overload", "deprecated",
		"readonly", "export", "register", "cdecl",
		"safecall", "stdcall", "fastcall", "reference",
		"private", "protected", "public", "published", "strict",
		"read", "write", "default", "description",
		"old", "require", "ensure", "invariants",
		"async", "await", "lambda", "implies", "empty", "implicit",
	}

	for _, keyword := range keywords {
		t.Run(keyword, func(t *testing.T) {
			l := New(keyword)
			tok := l.NextToken()

			if tok.Type == IDENT {
				t.Fatalf("keyword %q was tokenized as IDENT", keyword)
			}

			if !tok.Type.IsKeyword() {
				t.Fatalf("keyword %q not recognized as keyword, got type %q", keyword, tok.Type)
			}
		})
	}
}

func TestCompilerDirective(t *testing.T) {
	input := `{$DEFINE DEBUG}
	x := 5;`

	tests := []struct {
		expectedType TokenType
	}{
		{IDENT}, // x
		{ASSIGN},
		{INT},
		{SEMICOLON},
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

// Benchmark tests
func BenchmarkLexer(b *testing.B) {
	input := `
	function Fibonacci(n: Integer): Integer;
	begin
		if n <= 1 then
			Result := n
		else
			Result := Fibonacci(n-1) + Fibonacci(n-2);
	end;

	var i: Integer;
	begin
		for i := 0 to 10 do
			PrintLn(Fibonacci(i));
	end.
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

func BenchmarkLexerKeywords(b *testing.B) {
	input := "begin end if then else while for do function procedure class var const"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

func BenchmarkLexerNumbers(b *testing.B) {
	input := "123 456 789 $FF $10 0xFF 123.45 1.5e10 %1010"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

func BenchmarkLexerStrings(b *testing.B) {
	input := `'hello' 'world' 'this is a test' 'it''s working'`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
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
// Task 9.211: Add lambda keyword to lexer
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

func TestUnicodeIdentifiers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []struct {
			typ     TokenType
			literal string
		}
	}{
		{
			name:  "Greek letter Delta",
			input: "var Δ : Integer;",
			want: []struct {
				typ     TokenType
				literal string
			}{
				{VAR, "var"},
				{IDENT, "Δ"},
				{COLON, ":"},
				{IDENT, "Integer"},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
		{
			name:  "Greek letters alpha and beta",
			input: "α := β + 1;",
			want: []struct {
				typ     TokenType
				literal string
			}{
				{IDENT, "α"},
				{ASSIGN, ":="},
				{IDENT, "β"},
				{PLUS, "+"},
				{INT, "1"},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
		{
			name:  "Cyrillic variable names",
			input: "var переменная : Integer;",
			want: []struct {
				typ     TokenType
				literal string
			}{
				{VAR, "var"},
				{IDENT, "переменная"},
				{COLON, ":"},
				{IDENT, "Integer"},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
		{
			name:  "Chinese characters",
			input: "var 变量 := 42;",
			want: []struct {
				typ     TokenType
				literal string
			}{
				{VAR, "var"},
				{IDENT, "变量"},
				{ASSIGN, ":="},
				{INT, "42"},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
		{
			name:  "Japanese hiragana and katakana",
			input: "var へんすう := カタカナ;",
			want: []struct {
				typ     TokenType
				literal string
			}{
				{VAR, "var"},
				{IDENT, "へんすう"},
				{ASSIGN, ":="},
				{IDENT, "カタカナ"},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
		{
			name:  "Mixed ASCII and Unicode",
			input: "var myΔValue := 100;",
			want: []struct {
				typ     TokenType
				literal string
			}{
				{VAR, "var"},
				{IDENT, "myΔValue"},
				{ASSIGN, ":="},
				{INT, "100"},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
		{
			name:  "Underscore with Unicode",
			input: "var test_Δ := 42;",
			want: []struct {
				typ     TokenType
				literal string
			}{
				{VAR, "var"},
				{IDENT, "test_Δ"},
				{ASSIGN, ":="},
				{INT, "42"},
				{SEMICOLON, ";"},
				{EOF, ""},
			},
		},
		{
			name:  "Unicode in function call",
			input: "PrintLn(Δ);",
			want: []struct {
				typ     TokenType
				literal string
			}{
				{IDENT, "PrintLn"},
				{LPAREN, "("},
				{IDENT, "Δ"},
				{RPAREN, ")"},
				{SEMICOLON, ";"},
				{EOF, ""},
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
		typ     TokenType
		literal string
	}{
		{VAR, "var"},
		{IDENT, "Δ"},
		{COLON, ":"},
		{IDENT, "Integer"},
		{SEMICOLON, ";"},
		{IDENT, "Δ"},
		{ASSIGN, ":="},
		{INT, "1"},
		{SEMICOLON, ";"},
		{IDENT, "Inc"},
		{LPAREN, "("},
		{IDENT, "Δ"},
		{RPAREN, ")"},
		{SEMICOLON, ";"},
		{IDENT, "PrintLn"},
		{LPAREN, "("},
		{IDENT, "Δ"},
		{RPAREN, ")"},
		{SEMICOLON, ";"},
		{EOF, ""},
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

func TestDebugSHR(t *testing.T) {
	input := "shl shr"
	l := New(input)

	tok1 := l.NextToken()
	t.Logf("Token 1: Type=%s, Literal=%q", tok1.Type, tok1.Literal)

	tok2 := l.NextToken()
	t.Logf("Token 2: Type=%s, Literal=%q", tok2.Type, tok2.Literal)

	if tok1.Type != SHL {
		t.Errorf("Expected SHL, got %s", tok1.Type)
	}
	if tok2.Type != SHR {
		t.Errorf("Expected SHR, got %s", tok2.Type)
	}
}

func TestDebugPositions(t *testing.T) {
	l := New("shr")

	t.Logf("Initial: pos=%d, readPos=%d, ch=%q", l.position, l.readPosition, l.ch)

	// Manually test readIdentifier
	startPos := l.position
	l.readChar()
	t.Logf("After 1st readChar: pos=%d, readPos=%d, ch=%q", l.position, l.readPosition, l.ch)
	l.readChar()
	t.Logf("After 2nd readChar: pos=%d, readPos=%d, ch=%q", l.position, l.readPosition, l.ch)
	l.readChar()
	t.Logf("After 3rd readChar: pos=%d, readPos=%d, ch=%q", l.position, l.readPosition, l.ch)

	result := l.input[startPos:l.position]
	t.Logf("Identifier slice [%d:%d] = %q", startPos, l.position, result)
}
