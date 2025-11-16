package lexer

import (
	"fmt"
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
	input := `begin end if then else while for step do
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
		{"step", STEP},
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
		{"#65", CHAR},  // #65 = ASCII 'A'
		{"#$41", CHAR}, // #$41 = ASCII 'A' (hex)
		{"#13", CHAR},  // #13 = CR
		{"#10", CHAR},  // #10 = LF
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
			expectedType: STRING, // Now returns STRING token with error accumulated
		},
		{
			name:         "unterminated block comment",
			input:        "{ comment",
			expectedType: EOF, // Comments skip to EOF, error is accumulated
		},
		{
			name:         "unterminated C-style comment",
			input:        "/* comment",
			expectedType: EOF, // Comments skip to EOF, error is accumulated
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

			// For error cases, verify error was accumulated (Task 12.1)
			if tt.name == "unterminated string" || tt.name == "unterminated block comment" ||
				tt.name == "unterminated C-style comment" || tt.name == "illegal character" {
				errors := l.Errors()
				if len(errors) == 0 {
					t.Errorf("expected error to be accumulated, but got none")
				}
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
		"while", "repeat", "until", "for", "to", "downto", "step", "do",
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
		"readonly", "export",
		// Note: Calling conventions (register, pascal, cdecl, etc.) are NOT keywords
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

func TestBOMHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectLit   string
		expectFirst TokenType
	}{
		{
			name:        "UTF-8 BOM followed by var",
			input:       "\xEF\xBB\xBFvar x := 5;",
			expectFirst: VAR,
			expectLit:   "var",
		},
		{
			name:        "UTF-8 BOM followed by comment then var",
			input:       "\xEF\xBB\xBF// comment\nvar x := 5;",
			expectFirst: VAR,
			expectLit:   "var",
		},
		{
			name:        "UTF-8 BOM followed by function",
			input:       "\xEF\xBB\xBFfunction Test(): Integer;\nbegin\nend;",
			expectFirst: FUNCTION,
			expectLit:   "function",
		},
		{
			name:        "No BOM - ensure no regression",
			input:       "var x := 5;",
			expectFirst: VAR,
			expectLit:   "var",
		},
		{
			name:        "Partial BOM (2 bytes) should treat as ILLEGAL",
			input:       "\xEF\xBBvar x := 5;",
			expectFirst: ILLEGAL,
			expectLit:   "\uFFFD", // Unicode replacement character for invalid UTF-8
		},
		{
			name:        "Empty file with just BOM",
			input:       "\xEF\xBB\xBF",
			expectFirst: EOF,
			expectLit:   "",
		},
		{
			name:        "UTF-8 BOM followed by integer",
			input:       "\xEF\xBB\xBF42",
			expectFirst: INT,
			expectLit:   "42",
		},
		{
			name:        "UTF-8 BOM followed by string",
			input:       "\xEF\xBB\xBF\"hello\"",
			expectFirst: STRING,
			expectLit:   "hello",
		},
		{
			name:        "UTF-8 BOM with whitespace",
			input:       "\xEF\xBB\xBF   var x := 5;",
			expectFirst: VAR,
			expectLit:   "var",
		},
		{
			name:        "UTF-8 BOM with block comment",
			input:       "\xEF\xBB\xBF{ comment }\nvar x;",
			expectFirst: VAR,
			expectLit:   "var",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			// Skip whitespace and comments to get to first real token
			for tok.Type == EOF {
				break
			}

			if tok.Type != tt.expectFirst {
				t.Errorf("expected first token type %v, got %v", tt.expectFirst, tok.Type)
			}

			if tok.Literal != tt.expectLit {
				t.Errorf("expected first token literal %q, got %q", tt.expectLit, tok.Literal)
			}

			// For ILLEGAL tokens with partial BOM, verify position is correct
			if tt.expectFirst == ILLEGAL && tok.Pos.Line != 1 {
				t.Errorf("expected ILLEGAL token at line 1, got line %d", tok.Pos.Line)
			}
		})
	}
}

func TestBOMWithRealFixtureContent(t *testing.T) {
	// Test with actual content similar to the failing fixture files
	tests := []struct {
		name      string
		input     string
		numTokens int // Approximate number of tokens to verify parsing continues
	}{
		{
			name: "BOM with comment and var declaration",
			input: "\xEF\xBB\xBF// rc algos\n" +
				"var i: Integer;\n" +
				"begin\n" +
				"  i := 1;\n" +
				"end;",
			numTokens: 10, // var, i, :, Integer, ;, begin, i, :=, 1, ;, end, ;
		},
		{
			name: "BOM with type declaration",
			input: "\xEF\xBB\xBFtype\n" +
				"  TMyType = Integer;\n" +
				"var x: TMyType;",
			numTokens: 8, // type, TMyType, =, Integer, ;, var, x, :, TMyType, ;
		},
		{
			name: "BOM with procedure",
			input: "\xEF\xBB\xBFprocedure Test;\n" +
				"begin\n" +
				"end;",
			numTokens: 4, // procedure, Test, ;, begin, end, ;
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			tokenCount := 0
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
				tokenCount++

				// Verify no ILLEGAL tokens are produced
				if tok.Type == ILLEGAL {
					t.Errorf("unexpected ILLEGAL token at position %d: %q", tok.Pos.Offset, tok.Literal)
				}
			}

			if tokenCount < tt.numTokens {
				t.Errorf("expected at least %d tokens, got %d", tt.numTokens, tokenCount)
			}
		})
	}
}

// TestNotInOperator tests tokenization of "not in" expression
// This verifies that "not" and "in" are correctly tokenized as separate tokens
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

// TestErrorAccumulation tests that lexer errors are properly accumulated
// instead of stopping at the first error (Task 12.1)
func TestErrorAccumulation(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCount int
		errorMessages []string
	}{
		{
			name:          "Unterminated string - single quote",
			input:         `'hello`,
			expectedCount: 1,
			errorMessages: []string{"unterminated string literal"},
		},
		{
			name:          "Unterminated string - double quote",
			input:         `"hello`,
			expectedCount: 1,
			errorMessages: []string{"unterminated string literal"},
		},
		{
			name:          "Unterminated block comment - brace style",
			input:         `{ this is a comment`,
			expectedCount: 1,
			errorMessages: []string{"unterminated block comment"},
		},
		{
			name:          "Unterminated block comment - paren style",
			input:         `(* this is a comment`,
			expectedCount: 1,
			errorMessages: []string{"unterminated block comment"},
		},
		{
			name:          "Unterminated C-style comment",
			input:         `/* this is a comment`,
			expectedCount: 1,
			errorMessages: []string{"unterminated C-style comment"},
		},
		{
			name:          "Invalid character literal",
			input:         `'hello'#XYZ'world'`,
			expectedCount: 1,
			errorMessages: []string{"invalid character literal"},
		},
		{
			name:          "Illegal character",
			input:         `¿`,
			expectedCount: 1,
			errorMessages: []string{"illegal character"},
		},
		{
			name:          "Multiple errors - illegal characters",
			input:         "x := 5; ¿ y := 10; ¡",
			expectedCount: 2,
			errorMessages: []string{"illegal character"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Consume all tokens
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
			}

			errors := l.Errors()
			if len(errors) != tt.expectedCount {
				t.Errorf("expected %d errors, got %d", tt.expectedCount, len(errors))
				for i, err := range errors {
					t.Logf("  error[%d]: %s", i, err.Message)
				}
				return
			}

			// Check that expected error messages are present
			for _, expectedMsg := range tt.errorMessages {
				found := false
				for _, err := range errors {
					if len(err.Message) >= len(expectedMsg) && err.Message[:len(expectedMsg)] == expectedMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error message containing %q not found", expectedMsg)
					for i, err := range errors {
						t.Logf("  error[%d]: %s", i, err.Message)
					}
				}
			}
		})
	}
}

// TestErrorAccumulationPositions tests that error positions are correct
func TestErrorAccumulationPositions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedLine int
		expectedCol  int
	}{
		{
			name:         "Unterminated string at start",
			input:        `'hello`,
			expectedLine: 1,
			expectedCol:  1,
		},
		{
			name:         "Unterminated string on line 2",
			input:        "x := 5;\n'hello",
			expectedLine: 2,
			expectedCol:  1,
		},
		{
			name:         "Unterminated comment on line 3",
			input:        "x := 5;\ny := 10;\n{ comment",
			expectedLine: 3,
			expectedCol:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Consume all tokens
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
			}

			errors := l.Errors()
			if len(errors) == 0 {
				t.Fatal("expected at least one error")
			}

			err := errors[0]
			if err.Pos.Line != tt.expectedLine {
				t.Errorf("error line wrong. expected=%d, got=%d", tt.expectedLine, err.Pos.Line)
			}
			if err.Pos.Column != tt.expectedCol {
				t.Errorf("error column wrong. expected=%d, got=%d", tt.expectedCol, err.Pos.Column)
			}
		})
	}
}

// TestNoErrorsOnValidInput tests that no errors are accumulated for valid input
func TestNoErrorsOnValidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Simple program",
			input: `var x := 5; x := x + 10;`,
		},
		{
			name:  "String literals",
			input: `'hello' "world" 'it''s'`,
		},
		{
			name:  "Block comments",
			input: `{ comment } (* another *) /* c-style */`,
		},
		{
			name:  "Character literals",
			input: `#13 #10 #$0D #$0A`,
		},
		{
			name:  "String concatenation",
			input: `'hello'#13#10'world'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Consume all tokens
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
			}

			errors := l.Errors()
			if len(errors) != 0 {
				t.Errorf("expected no errors, got %d", len(errors))
				for i, err := range errors {
					t.Logf("  error[%d]: %s at %d:%d", i, err.Message, err.Pos.Line, err.Pos.Column)
				}
			}
		})
	}
}

// TestPeekCharDoesNotModifyState tests that peekChar() doesn't modify lexer state (Task 12.2.3)
func TestPeekCharDoesNotModifyState(t *testing.T) {
	input := "abc"
	l := New(input)

	// Current character should be 'a'
	if l.ch != 'a' {
		t.Fatalf("expected current char 'a', got %c", l.ch)
	}

	// Peek should return 'b'
	peeked := l.peekChar()
	if peeked != 'b' {
		t.Fatalf("peekChar() expected 'b', got %c", peeked)
	}

	// Current character should still be 'a'
	if l.ch != 'a' {
		t.Fatalf("after peekChar(), expected current char 'a', got %c", l.ch)
	}

	// Position should not have changed
	if l.position != 0 {
		t.Fatalf("after peekChar(), expected position 0, got %d", l.position)
	}
}

// TestPeekCharN tests peekCharN() for various N values (Task 12.2.3)
func TestPeekCharN(t *testing.T) {
	input := "abcdef"
	l := New(input)

	tests := []struct {
		n        int
		expected rune
	}{
		{1, 'b'}, // peek 1 ahead
		{2, 'c'}, // peek 2 ahead
		{3, 'd'}, // peek 3 ahead
		{4, 'e'}, // peek 4 ahead
		{5, 'f'}, // peek 5 ahead
		{6, 0},   // peek beyond EOF
		{10, 0},  // peek way beyond EOF
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("peek_%d", tt.n), func(t *testing.T) {
			peeked := l.peekCharN(tt.n)
			if peeked != tt.expected {
				t.Errorf("peekCharN(%d) expected %q, got %q", tt.n, tt.expected, peeked)
			}

			// Current character should still be 'a'
			if l.ch != 'a' {
				t.Errorf("after peekCharN(%d), expected current char 'a', got %c", tt.n, l.ch)
			}

			// Position should not have changed
			if l.position != 0 {
				t.Errorf("after peekCharN(%d), expected position 0, got %d", tt.n, l.position)
			}
		})
	}
}

// TestPeekCharNWithUTF8 tests peekCharN() with multi-byte UTF-8 characters (Task 12.2.3)
func TestPeekCharNWithUTF8(t *testing.T) {
	input := "aβγδ" // 'a' (1 byte), 'β' (2 bytes), 'γ' (2 bytes), 'δ' (2 bytes)
	l := New(input)

	tests := []struct {
		n        int
		expected rune
	}{
		{1, 'β'},
		{2, 'γ'},
		{3, 'δ'},
		{4, 0}, // beyond EOF
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("peek_%d", tt.n), func(t *testing.T) {
			peeked := l.peekCharN(tt.n)
			if peeked != tt.expected {
				t.Errorf("peekCharN(%d) expected %q, got %q", tt.n, tt.expected, peeked)
			}

			// Current character should still be 'a'
			if l.ch != 'a' {
				t.Errorf("after peekCharN(%d), expected current char 'a', got %c", tt.n, l.ch)
			}
		})
	}
}

// TestSaveRestoreStateSymmetry tests that saveState/restoreState is symmetric (Task 12.2.3)
func TestSaveRestoreStateSymmetry(t *testing.T) {
	input := "var x := 5;\ny := 10;"
	l := New(input)

	// Save initial state
	state1 := l.saveState()

	// Advance lexer by reading some tokens
	l.NextToken() // var
	l.NextToken() // x
	l.NextToken() // :=

	// Save state after tokens
	state2 := l.saveState()

	// Advance more
	l.NextToken() // 5
	l.NextToken() // ;

	// Check we're at a different position
	if l.position == state2.position {
		t.Fatal("lexer should have advanced")
	}

	// Restore to state2
	l.restoreState(state2)

	// Verify state matches state2
	if l.position != state2.position {
		t.Errorf("position mismatch: expected %d, got %d", state2.position, l.position)
	}
	if l.readPosition != state2.readPosition {
		t.Errorf("readPosition mismatch: expected %d, got %d", state2.readPosition, l.readPosition)
	}
	if l.ch != state2.ch {
		t.Errorf("ch mismatch: expected %c, got %c", state2.ch, l.ch)
	}
	if l.line != state2.line {
		t.Errorf("line mismatch: expected %d, got %d", state2.line, l.line)
	}
	if l.column != state2.column {
		t.Errorf("column mismatch: expected %d, got %d", state2.column, l.column)
	}

	// Next token should be '5' again
	tok := l.NextToken()
	if tok.Type != INT || tok.Literal != "5" {
		t.Errorf("after restore, expected INT(5), got %s(%s)", tok.Type, tok.Literal)
	}

	// Restore to initial state
	l.restoreState(state1)

	// Next token should be 'var' again
	tok = l.NextToken()
	if tok.Type != VAR || tok.Literal != "var" {
		t.Errorf("after restore to initial, expected VAR(var), got %s(%s)", tok.Type, tok.Literal)
	}
}

// TestSaveRestoreStatePreservesLineColumn tests that line/column are correctly saved/restored (Task 12.2.3)
func TestSaveRestoreStatePreservesLineColumn(t *testing.T) {
	input := "var x := 5;\ny := 10;\nz := 15;"
	l := New(input)

	// Read some tokens to advance through the input
	l.NextToken() // var
	l.NextToken() // x
	l.NextToken() // :=
	l.NextToken() // 5
	l.NextToken() // ;

	// Should be on line 2 now (after the newline)
	line1 := l.line
	col1 := l.column

	// Save state
	state := l.saveState()

	// Read more tokens
	l.NextToken() // y
	l.NextToken() // :=
	l.NextToken() // 10
	l.NextToken() // ;

	// Should have advanced
	if l.line == line1 && l.column == col1 {
		t.Fatal("lexer should have advanced")
	}

	line2 := l.line
	col2 := l.column

	// Restore to saved state
	l.restoreState(state)

	// Should be back at the saved position
	if l.line != state.line {
		t.Errorf("after restore, line mismatch: expected %d, got %d", state.line, l.line)
	}
	if l.column != state.column {
		t.Errorf("after restore, column mismatch: expected %d, got %d", state.column, l.column)
	}
	if l.line == line2 || l.column == col2 {
		t.Error("after restore, should not be at the advanced position")
	}
}

// TestCharLiteralStandaloneStillWorks tests that isCharLiteralStandalone works after refactoring (Task 12.2.2)
func TestCharLiteralStandaloneStillWorks(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		isStandalone bool
	}{
		{
			name:        "standalone character literal",
			input:       "#65",
			isStandalone: true,
		},
		{
			name:        "character literal followed by space and string",
			input:       "#65 'hello'",
			isStandalone: true, // space separates them
		},
		{
			name:        "character literal immediately followed by string",
			input:       "#65'hello'",
			isStandalone: false, // no space, part of concatenation
		},
		{
			name:        "character literal followed by another char literal",
			input:       "#65#66",
			isStandalone: false, // concatenation
		},
		{
			name:        "hex character literal standalone",
			input:       "#$41",
			isStandalone: true,
		},
		{
			name:        "hex character literal in concatenation",
			input:       "#$41#$42",
			isStandalone: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			result := l.isCharLiteralStandalone()
			if result != tt.isStandalone {
				t.Errorf("isCharLiteralStandalone() = %v, expected %v", result, tt.isStandalone)
			}

			// Verify state wasn't changed
			if l.position != 0 {
				t.Errorf("isCharLiteralStandalone() changed position to %d, expected 0", l.position)
			}
			if l.ch != '#' {
				t.Errorf("isCharLiteralStandalone() changed ch to %c, expected '#'", l.ch)
			}
		})
	}
}
