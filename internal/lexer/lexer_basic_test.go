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
