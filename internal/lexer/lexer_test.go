package lexer

import (
	"fmt"
	"strings"
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
		errorMessages []string
		expectedCount int
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
	state1 := l.SaveState()

	// Advance lexer by reading some tokens
	l.NextToken() // var
	l.NextToken() // x
	l.NextToken() // :=

	// Save state after tokens
	state2 := l.SaveState()

	// Advance more
	l.NextToken() // 5
	l.NextToken() // ;

	// Check we're at a different position
	if l.position == state2.position {
		t.Fatal("lexer should have advanced")
	}

	// Restore to state2
	l.RestoreState(state2)

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
	l.RestoreState(state1)

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
	state := l.SaveState()

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
	l.RestoreState(state)

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
		name         string
		input        string
		isStandalone bool
	}{
		{
			name:         "standalone character literal",
			input:        "#65",
			isStandalone: true,
		},
		{
			name:         "character literal followed by space and string",
			input:        "#65 'hello'",
			isStandalone: true, // space separates them
		},
		{
			name:         "character literal immediately followed by string",
			input:        "#65'hello'",
			isStandalone: false, // no space, part of concatenation
		},
		{
			name:         "character literal followed by another char literal",
			input:        "#65#66",
			isStandalone: false, // concatenation
		},
		{
			name:         "hex character literal standalone",
			input:        "#$41",
			isStandalone: true,
		},
		{
			name:         "hex character literal in concatenation",
			input:        "#$41#$42",
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

// TestMatchAndConsume tests the matchAndConsume helper method (Task 12.5.1)
func TestMatchAndConsume(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    rune
		shouldMatch bool
		expectedPos int  // position after matchAndConsume
		expectedCh  rune // current char after matchAndConsume
	}{
		{
			name:        "match succeeds",
			input:       "++",
			expected:    '+',
			shouldMatch: true,
			expectedPos: 1,
			expectedCh:  '+',
		},
		{
			name:        "match fails",
			input:       "+=",
			expected:    '+',
			shouldMatch: false,
			expectedPos: 0,
			expectedCh:  '+',
		},
		{
			name:        "match at end of input",
			input:       "+",
			expected:    0, // EOF
			shouldMatch: true,
			expectedPos: 1,
			expectedCh:  0,
		},
		{
			name:        "no match at end of input",
			input:       "+",
			expected:    'x',
			shouldMatch: false,
			expectedPos: 0,
			expectedCh:  '+',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			result := l.matchAndConsume(tt.expected)

			if result != tt.shouldMatch {
				t.Errorf("matchAndConsume(%q) = %v, expected %v", tt.expected, result, tt.shouldMatch)
			}

			if l.position != tt.expectedPos {
				t.Errorf("after matchAndConsume, position = %d, expected %d", l.position, tt.expectedPos)
			}

			if l.ch != tt.expectedCh {
				t.Errorf("after matchAndConsume, ch = %q, expected %q", l.ch, tt.expectedCh)
			}
		})
	}
}

// TestPeekToken tests basic token peeking functionality (Task 12.3.2)
func TestPeekToken(t *testing.T) {
	input := "var x := 5;"
	l := New(input)

	// Peek at first token without consuming
	tok := l.Peek(0)
	if tok.Type != VAR || tok.Literal != "var" {
		t.Fatalf("Peek(0) expected VAR(var), got %s(%s)", tok.Type, tok.Literal)
	}

	// Peek again - should get same token
	tok2 := l.Peek(0)
	if tok2.Type != VAR || tok2.Literal != "var" {
		t.Fatalf("Peek(0) second call expected VAR(var), got %s(%s)", tok2.Type, tok2.Literal)
	}

	// Peek ahead 1 token
	tok3 := l.Peek(1)
	if tok3.Type != IDENT || tok3.Literal != "x" {
		t.Fatalf("Peek(1) expected IDENT(x), got %s(%s)", tok3.Type, tok3.Literal)
	}

	// First peek should still return var
	tok4 := l.Peek(0)
	if tok4.Type != VAR {
		t.Fatalf("Peek(0) after Peek(1) expected VAR, got %s", tok4.Type)
	}

	// Now consume the first token
	consumed := l.NextToken()
	if consumed.Type != VAR || consumed.Literal != "var" {
		t.Fatalf("NextToken() expected VAR(var), got %s(%s)", consumed.Type, consumed.Literal)
	}

	// Peek should now return the next token
	tok5 := l.Peek(0)
	if tok5.Type != IDENT || tok5.Literal != "x" {
		t.Fatalf("Peek(0) after NextToken() expected IDENT(x), got %s(%s)", tok5.Type, tok5.Literal)
	}
}

// TestPeekMultipleTokens tests peeking multiple tokens ahead (Task 12.3.2)
func TestPeekMultipleTokens(t *testing.T) {
	input := "var x := 5;"
	l := New(input)

	tests := []struct {
		expectedLit  string
		peekN        int
		expectedType TokenType
	}{
		{expectedLit: "var", peekN: 0, expectedType: VAR},
		{expectedLit: "x", peekN: 1, expectedType: IDENT},
		{expectedLit: ":=", peekN: 2, expectedType: ASSIGN},
		{expectedLit: "5", peekN: 3, expectedType: INT},
		{expectedLit: ";", peekN: 4, expectedType: SEMICOLON},
		{expectedLit: "", peekN: 5, expectedType: EOF},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Peek(%d)", tt.peekN), func(t *testing.T) {
			tok := l.Peek(tt.peekN)
			if tok.Type != tt.expectedType {
				t.Errorf("Peek(%d) type: expected %s, got %s", tt.peekN, tt.expectedType, tok.Type)
			}
			if tok.Literal != tt.expectedLit {
				t.Errorf("Peek(%d) literal: expected %q, got %q", tt.peekN, tt.expectedLit, tok.Literal)
			}
		})
	}

	// Peek should not consume - first NextToken should still return VAR
	tok := l.NextToken()
	if tok.Type != VAR {
		t.Errorf("After all Peeks, NextToken() expected VAR, got %s", tok.Type)
	}
}

// TestPeekAndConsumeInterleaved tests interleaving Peek and NextToken (Task 12.3.2)
func TestPeekAndConsumeInterleaved(t *testing.T) {
	input := "a b c d"
	l := New(input)

	// Peek ahead to 'b'
	tok := l.Peek(1)
	if tok.Literal != "b" {
		t.Fatalf("Peek(1) expected 'b', got %q", tok.Literal)
	}

	// Consume 'a'
	tok = l.NextToken()
	if tok.Literal != "a" {
		t.Fatalf("NextToken() expected 'a', got %q", tok.Literal)
	}

	// Peek ahead to 'c'
	tok = l.Peek(1)
	if tok.Literal != "c" {
		t.Fatalf("Peek(1) expected 'c', got %q", tok.Literal)
	}

	// Peek 0 should be 'b'
	tok = l.Peek(0)
	if tok.Literal != "b" {
		t.Fatalf("Peek(0) expected 'b', got %q", tok.Literal)
	}

	// Consume 'b'
	tok = l.NextToken()
	if tok.Literal != "b" {
		t.Fatalf("NextToken() expected 'b', got %q", tok.Literal)
	}

	// Peek 0 should now be 'c'
	tok = l.Peek(0)
	if tok.Literal != "c" {
		t.Fatalf("Peek(0) expected 'c', got %q", tok.Literal)
	}

	// Peek 1 should now be 'd'
	tok = l.Peek(1)
	if tok.Literal != "d" {
		t.Fatalf("Peek(1) expected 'd', got %q", tok.Literal)
	}
}

// TestPeekWithComments tests that Peek handles comments correctly (Task 12.3.2)
func TestPeekWithComments(t *testing.T) {
	input := "var // comment\nx := 5"
	l := New(input)

	// Peek should skip the comment
	tok := l.Peek(0)
	if tok.Type != VAR || tok.Literal != "var" {
		t.Fatalf("Peek(0) expected VAR(var), got %s(%s)", tok.Type, tok.Literal)
	}

	// Peek 1 should be 'x', skipping the comment
	tok = l.Peek(1)
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Fatalf("Peek(1) expected IDENT(x), got %s(%s)", tok.Type, tok.Literal)
	}

	// Consume tokens should also skip comments
	tok = l.NextToken()
	if tok.Type != VAR {
		t.Fatalf("NextToken() expected VAR, got %s", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != IDENT {
		t.Fatalf("NextToken() expected IDENT, got %s", tok.Type)
	}
}

// TestNextTokenAfterPeekConsumesBuffer tests that NextToken consumes buffered tokens (Task 12.3.3)
func TestNextTokenAfterPeekConsumesBuffer(t *testing.T) {
	input := "a b c"
	l := New(input)

	// Peek ahead multiple times - this buffers tokens
	_ = l.Peek(0) // buffers 'a'
	_ = l.Peek(1) // buffers 'b'
	_ = l.Peek(2) // buffers 'c'

	// NextToken should consume from buffer
	tok1 := l.NextToken()
	if tok1.Literal != "a" {
		t.Fatalf("NextToken() expected 'a', got %q", tok1.Literal)
	}

	tok2 := l.NextToken()
	if tok2.Literal != "b" {
		t.Fatalf("NextToken() expected 'b', got %q", tok2.Literal)
	}

	tok3 := l.NextToken()
	if tok3.Literal != "c" {
		t.Fatalf("NextToken() expected 'c', got %q", tok3.Literal)
	}

	// Should be at EOF now
	tok4 := l.NextToken()
	if tok4.Type != EOF {
		t.Fatalf("NextToken() expected EOF, got %s", tok4.Type)
	}
}

// TestPeekPreservesExistingBehavior tests that existing code still works (Task 12.3.3)
func TestPeekPreservesExistingBehavior(t *testing.T) {
	input := "var x := 5;"
	l := New(input)

	// Consume tokens as usual without using Peek
	tokens := []struct {
		lit string
		typ TokenType
	}{
		{lit: "var", typ: VAR},
		{lit: "x", typ: IDENT},
		{lit: ":=", typ: ASSIGN},
		{lit: "5", typ: INT},
		{lit: ";", typ: SEMICOLON},
		{lit: "", typ: EOF},
	}

	for i, expected := range tokens {
		tok := l.NextToken()
		if tok.Type != expected.typ {
			t.Errorf("token[%d] type: expected %s, got %s", i, expected.typ, tok.Type)
		}
		if tok.Literal != expected.lit {
			t.Errorf("token[%d] literal: expected %q, got %q", i, expected.lit, tok.Literal)
		}
	}
}

// TestLexerOptions tests the options pattern for Lexer configuration.
func TestLexerOptions(t *testing.T) {
	input := `var x := 42; // comment
{ block comment }`

	t.Run("default configuration", func(t *testing.T) {
		l := New(input)
		if l.preserveComments {
			t.Error("preserveComments should be false by default")
		}
		if l.tracing {
			t.Error("tracing should be false by default")
		}
	})

	t.Run("WithPreserveComments(true)", func(t *testing.T) {
		l := New(input, WithPreserveComments(true))
		if !l.preserveComments {
			t.Error("preserveComments should be true")
		}

		// Verify comments are preserved
		tokens := []TokenType{}
		for {
			tok := l.NextToken()
			tokens = append(tokens, tok.Type)
			if tok.Type == EOF {
				break
			}
		}

		// Should have COMMENT tokens
		hasComment := false
		for _, tt := range tokens {
			if tt == COMMENT {
				hasComment = true
				break
			}
		}
		if !hasComment {
			t.Error("expected COMMENT token when preserveComments is true")
		}
	})

	t.Run("WithPreserveComments(false)", func(t *testing.T) {
		l := New(input, WithPreserveComments(false))
		if l.preserveComments {
			t.Error("preserveComments should be false")
		}

		// Verify comments are skipped
		tokens := []TokenType{}
		for {
			tok := l.NextToken()
			tokens = append(tokens, tok.Type)
			if tok.Type == EOF {
				break
			}
		}

		// Should NOT have COMMENT tokens
		for _, tt := range tokens {
			if tt == COMMENT {
				t.Error("unexpected COMMENT token when preserveComments is false")
			}
		}
	})

	t.Run("WithTracing(true)", func(t *testing.T) {
		l := New(input, WithTracing(true))
		if !l.tracing {
			t.Error("tracing should be true")
		}
	})

	t.Run("WithTracing(false)", func(t *testing.T) {
		l := New(input, WithTracing(false))
		if l.tracing {
			t.Error("tracing should be false")
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		l := New(input, WithPreserveComments(true), WithTracing(true))
		if !l.preserveComments {
			t.Error("preserveComments should be true")
		}
		if !l.tracing {
			t.Error("tracing should be true")
		}
	})

	t.Run("backwards compatibility - no options", func(t *testing.T) {
		// Should work without any options (backwards compatible)
		l := New(input)
		tok := l.NextToken()
		if tok.Type != VAR {
			t.Errorf("expected VAR token, got %s", tok.Type)
		}
	})
}

// TestOptionsBackwardsCompatibility ensures that existing code without options still works.
func TestOptionsBackwardsCompatibility(t *testing.T) {
	input := `var x: Integer := 42;`

	// Old usage pattern (no options)
	l := New(input)

	tokens := []struct {
		lit string
		typ TokenType
	}{
		{lit: "var", typ: VAR},
		{lit: "x", typ: IDENT},
		{lit: ":", typ: COLON},
		{lit: "Integer", typ: IDENT},
		{lit: ":=", typ: ASSIGN},
		{lit: "42", typ: INT},
		{lit: ";", typ: SEMICOLON},
		{lit: "", typ: EOF},
	}

	for i, expected := range tokens {
		tok := l.NextToken()
		if tok.Type != expected.typ {
			t.Errorf("token[%d] type: expected %s, got %s", i, expected.typ, tok.Type)
		}
		if tok.Literal != expected.lit {
			t.Errorf("token[%d] literal: expected %q, got %q", i, expected.lit, tok.Literal)
		}
	}
}

// TestErrorPositionsUnterminatedStrings tests error positions for unterminated strings at various locations.
func TestErrorPositionsUnterminatedStrings(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		errorLine int
		errorCol  int
	}{
		{
			name:      "unterminated string at start of file",
			input:     `'unterminated`,
			errorLine: 1,
			errorCol:  1,
		},
		{
			name:      "unterminated string on line 1 after valid token",
			input:     `var x := 'unterminated`,
			errorLine: 1,
			errorCol:  10,
		},
		{
			name: "unterminated string on line 2",
			input: `var x := 42;
'unterminated on line 2`,
			errorLine: 2,
			errorCol:  1,
		},
		{
			name: "unterminated string on line 3",
			input: `var x := 42;
var y := 'valid';
'unterminated on line 3`,
			errorLine: 3,
			errorCol:  1,
		},
		{
			name:      "unterminated string mid-line",
			input:     `var x := 42; var s := 'unterminated`,
			errorLine: 1,
			errorCol:  23,
		},
		{
			name:      "unterminated double-quoted string",
			input:     `var x := "unterminated`,
			errorLine: 1,
			errorCol:  10,
		},
		{
			name: "unterminated multiline string",
			input: `var x := 'line 1
line 2
unterminated`,
			errorLine: 1,
			errorCol:  10,
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

			// Check that we have errors
			errors := l.Errors()
			if len(errors) == 0 {
				t.Fatal("expected at least one error, got none")
			}

			// Find the unterminated string error
			found := false
			for _, err := range errors {
				if err.Pos.Line == tt.errorLine && err.Pos.Column == tt.errorCol {
					found = true
					if !strings.Contains(strings.ToLower(err.Message), "unterminated") {
						t.Errorf("error message should contain 'unterminated', got: %s", err.Message)
					}
				}
			}

			if !found {
				t.Errorf("expected error at line %d, column %d, but got errors at:", tt.errorLine, tt.errorCol)
				for _, err := range errors {
					t.Logf("  - line %d, column %d: %s", err.Pos.Line, err.Pos.Column, err.Message)
				}
			}
		})
	}
}

// TestErrorPositionsIllegalCharacters tests error positions for illegal characters.
func TestErrorPositionsIllegalCharacters(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		errorLine int
		errorCol  int
		char      rune
	}{
		{
			name:      "illegal character at start",
			input:     "§ var x := 42;",
			errorLine: 1,
			errorCol:  1,
			char:      '§',
		},
		{
			name:      "illegal character mid-line",
			input:     "var x := 42 § 10;",
			errorLine: 1,
			errorCol:  13,
			char:      '§',
		},
		{
			name: "illegal character on line 2",
			input: `var x := 42;
var y := § 10;`,
			errorLine: 2,
			errorCol:  10,
			char:      '§',
		},
		{
			name: "illegal character on line 3",
			input: `var x := 42;
var y := 10;
var z := ¶ 20;`,
			errorLine: 3,
			errorCol:  10,
			char:      '¶',
		},
		{
			name:      "multiple illegal characters",
			input:     "§ var x := ¶ 42;",
			errorLine: 1, // Should report first error
			errorCol:  1,
			char:      '§',
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

			// Check that we have errors
			errors := l.Errors()
			if len(errors) == 0 {
				t.Fatal("expected at least one error, got none")
			}

			// Find the illegal character error at the expected position
			found := false
			for _, err := range errors {
				if err.Pos.Line == tt.errorLine && err.Pos.Column == tt.errorCol {
					found = true
					if !strings.Contains(strings.ToLower(err.Message), "illegal") {
						t.Errorf("error message should contain 'illegal', got: %s", err.Message)
					}
				}
			}

			if !found {
				t.Errorf("expected error at line %d, column %d, but got errors at:", tt.errorLine, tt.errorCol)
				for _, err := range errors {
					t.Logf("  - line %d, column %d: %s", err.Pos.Line, err.Pos.Column, err.Message)
				}
			}
		})
	}
}

// TestErrorPositionsMultiLineErrors tests error reporting across multiple lines.
func TestErrorPositionsMultiLineErrors(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors []struct {
			line int
			col  int
		}
	}{
		{
			name: "multiple illegal characters on different lines",
			input: `var x := 42;
var y := § 42;
var z := ¶ 20;`,
			expectedErrors: []struct{ line, col int }{
				{2, 10}, // illegal character on line 2
				{3, 10}, // illegal character on line 3
			},
		},
		{
			name: "unterminated comment on line 1",
			input: `{ unterminated comment
var x := 42;`,
			expectedErrors: []struct{ line, col int }{
				{1, 1}, // unterminated comment
			},
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

			// Check that we have errors
			errors := l.Errors()
			if len(errors) < len(tt.expectedErrors) {
				t.Errorf("expected at least %d errors, got %d", len(tt.expectedErrors), len(errors))
				for i, err := range errors {
					t.Logf("error %d: line %d, column %d: %s", i, err.Pos.Line, err.Pos.Column, err.Message)
				}
			}

			// Check each expected error position
			for _, expected := range tt.expectedErrors {
				found := false
				for _, err := range errors {
					if err.Pos.Line == expected.line && err.Pos.Column == expected.col {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error at line %d, column %d, but didn't find it", expected.line, expected.col)
				}
			}
		})
	}
}

// TestErrorPositionsWithUnicode tests error positions with Unicode characters.
// This verifies that column positions are reported correctly for multi-byte characters.
func TestErrorPositionsWithUnicode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		errorLine int
		errorCol  int
	}{
		{
			name:      "illegal char after ASCII",
			input:     "var § x := 42;",
			errorLine: 1,
			errorCol:  5,
		},
		{
			name:      "illegal char after Unicode identifier",
			input:     "var Δ § x := 42;",
			errorLine: 1,
			errorCol:  7,
		},
		{
			name:      "unterminated string with Unicode",
			input:     "var s := 'Hello Δ unterminated",
			errorLine: 1,
			errorCol:  10,
		},
		{
			name:      "illegal char after emoji in comment",
			input:     "// 🚀 comment\nvar x § 42;",
			errorLine: 2,
			errorCol:  7,
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

			// Check that we have errors
			errors := l.Errors()
			if len(errors) == 0 {
				t.Fatal("expected at least one error, got none")
			}

			// Find error at expected position
			found := false
			for _, err := range errors {
				if err.Pos.Line == tt.errorLine && err.Pos.Column == tt.errorCol {
					found = true
				}
			}

			if !found {
				t.Errorf("expected error at line %d, column %d (rune count), but got errors at:", tt.errorLine, tt.errorCol)
				for _, err := range errors {
					t.Logf("  - line %d, column %d: %s", err.Pos.Line, err.Pos.Column, err.Message)
				}
			}
		})
	}
}

// TestSaveRestoreStateWithPeekPreservesTokenBuffer tests that tokenBuffer is correctly saved/restored
// This is critical for parser backtracking when Peek() is used during speculative parsing.
// Bug: Before the fix, tokenBuffer was not included in LexerState, causing token duplication/skipping.
func TestSaveRestoreStateWithPeekPreservesTokenBuffer(t *testing.T) {
	input := "var x := 5 + 10;"
	l := New(input)

	// Save initial state (tokenBuffer is empty)
	state := l.SaveState()

	// Use Peek to fill the tokenBuffer
	tok0 := l.Peek(0) // var
	tok1 := l.Peek(1) // x
	tok2 := l.Peek(2) // :=

	// Verify peeked tokens
	if tok0.Type != VAR || tok0.Literal != "var" {
		t.Errorf("Peek(0) expected VAR(var), got %s(%s)", tok0.Type, tok0.Literal)
	}
	if tok1.Type != IDENT || tok1.Literal != "x" {
		t.Errorf("Peek(1) expected IDENT(x), got %s(%s)", tok1.Type, tok1.Literal)
	}
	if tok2.Type != ASSIGN || tok2.Literal != ":=" {
		t.Errorf("Peek(2) expected ASSIGN(:=), got %s(%s)", tok2.Type, tok2.Literal)
	}

	// At this point, tokenBuffer has 3 tokens: [var, x, :=]
	// Internal lexer position has advanced past these tokens

	// Restore to initial state
	// BUG: If tokenBuffer is not restored, it still contains [var, x, :=]
	// but position is back at start, causing tokens to be consumed from stale buffer
	l.RestoreState(state)

	// Now consume tokens via NextToken()
	// With the fix: tokenBuffer is empty, so tokens are generated fresh from position 0
	// Without the fix: tokenBuffer still has [var, x, :=], consuming them first

	tok := l.NextToken()
	if tok.Type != VAR || tok.Literal != "var" {
		t.Errorf("NextToken() after restore expected VAR(var), got %s(%s)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Errorf("NextToken() after restore expected IDENT(x), got %s(%s)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != ASSIGN || tok.Literal != ":=" {
		t.Errorf("NextToken() after restore expected ASSIGN(:=), got %s(%s)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != INT || tok.Literal != "5" {
		t.Errorf("NextToken() after restore expected INT(5), got %s(%s)", tok.Type, tok.Literal)
	}
}

// TestSaveRestoreStateWithPeekInMiddle tests save/restore with Peek() at a non-initial position
// This simulates the actual parser backtracking scenario in parseIsExpression.
func TestSaveRestoreStateWithPeekInMiddle(t *testing.T) {
	input := "if obj is function(x: Integer): String then end"
	l := New(input)

	// Advance to 'is' keyword
	l.NextToken() // if
	l.NextToken() // obj
	l.NextToken() // is

	// Save state at 'is' position
	state := l.SaveState()

	// Simulate parser's speculative parsing using Peek
	// This is what detectFunctionPointerFullSyntax does
	peeked := []Token{
		l.Peek(0), // function
		l.Peek(1), // (
		l.Peek(2), // x
		l.Peek(3), // :
		l.Peek(4), // Integer
	}

	// Verify peeked tokens
	expectedTypes := []TokenType{FUNCTION, LPAREN, IDENT, COLON, IDENT}
	for i, expected := range expectedTypes {
		if peeked[i].Type != expected {
			t.Errorf("Peek(%d) expected %s, got %s", i, expected, peeked[i].Type)
		}
	}

	// Now restore to the saved state (back to 'is' position)
	l.RestoreState(state)

	// Consume tokens - should start from 'function' again
	tok := l.NextToken()
	if tok.Type != FUNCTION || tok.Literal != "function" {
		t.Errorf("After restore, expected FUNCTION(function), got %s(%s)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != LPAREN || tok.Literal != "(" {
		t.Errorf("After restore, expected LPAREN((), got %s(%s)", tok.Type, tok.Literal)
	}

	// Continue verifying the sequence is correct
	tok = l.NextToken()
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Errorf("After restore, expected IDENT(x), got %s(%s)", tok.Type, tok.Literal)
	}
}
