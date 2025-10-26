package lexer

import (
	"strings"
	"testing"
)

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		expected  string
		tokenType TokenType
	}{
		{"ILLEGAL", ILLEGAL},
		{"EOF", EOF},
		{"IDENT", IDENT},
		{"INT", INT},
		{"FLOAT", FLOAT},
		{"STRING", STRING},
		{"BEGIN", BEGIN},
		{"END", END},
		{"IF", IF},
		{"THEN", THEN},
		{"FUNCTION", FUNCTION},
		{"CLASS", CLASS},
		{"PLUS", PLUS},
		{"MINUS", MINUS},
		{"ASSIGN", ASSIGN},
		{"EQ", EQ},
		{"NOT_EQ", NOT_EQ},
		{"LPAREN", LPAREN},
		{"RPAREN", RPAREN},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.tokenType.String(); got != tt.expected {
				t.Errorf("TokenType.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTokenTypePredicates(t *testing.T) {
	t.Run("IsLiteral", func(t *testing.T) {
		literals := []TokenType{IDENT, INT, FLOAT, STRING, CHAR}
		for _, tt := range literals {
			if !tt.IsLiteral() {
				t.Errorf("%s.IsLiteral() = false, want true", tt)
			}
		}

		// TRUE, FALSE, NIL are keywords that represent literal values,
		// but IsLiteral() is for non-keyword literals (numbers, strings, identifiers)
		nonLiterals := []TokenType{ILLEGAL, EOF, TRUE, FALSE, NIL, BEGIN, END, PLUS, LPAREN}
		for _, tt := range nonLiterals {
			if tt.IsLiteral() {
				t.Errorf("%s.IsLiteral() = true, want false", tt)
			}
		}
	})

	t.Run("IsKeyword", func(t *testing.T) {
		keywords := []TokenType{BEGIN, END, IF, THEN, ELSE, WHILE, FOR, FUNCTION, CLASS, VAR, CONST, TRUE, FALSE, NIL}
		for _, tt := range keywords {
			if !tt.IsKeyword() {
				t.Errorf("%s.IsKeyword() = false, want true", tt)
			}
		}

		nonKeywords := []TokenType{ILLEGAL, EOF, IDENT, INT, STRING, PLUS, LPAREN}
		for _, tt := range nonKeywords {
			if tt.IsKeyword() {
				t.Errorf("%s.IsKeyword() = true, want false", tt)
			}
		}
	})

	t.Run("IsOperator", func(t *testing.T) {
		operators := []TokenType{PLUS, MINUS, ASTERISK, SLASH, EQ, NOT_EQ, LESS, GREATER, ASSIGN}
		for _, tt := range operators {
			if !tt.IsOperator() {
				t.Errorf("%s.IsOperator() = false, want true", tt)
			}
		}

		nonOperators := []TokenType{ILLEGAL, EOF, IDENT, INT, BEGIN, END, LPAREN}
		for _, tt := range nonOperators {
			if tt.IsOperator() {
				t.Errorf("%s.IsOperator() = true, want false", tt)
			}
		}
	})

	t.Run("IsDelimiter", func(t *testing.T) {
		delimiters := []TokenType{LPAREN, RPAREN, LBRACK, RBRACK, LBRACE, RBRACE, SEMICOLON, COMMA, DOT, COLON}
		for _, tt := range delimiters {
			if !tt.IsDelimiter() {
				t.Errorf("%s.IsDelimiter() = false, want true", tt)
			}
		}

		nonDelimiters := []TokenType{ILLEGAL, EOF, IDENT, INT, BEGIN, PLUS}
		for _, tt := range nonDelimiters {
			if tt.IsDelimiter() {
				t.Errorf("%s.IsDelimiter() = true, want false", tt)
			}
		}
	})
}

func TestNewToken(t *testing.T) {
	pos := Position{Line: 1, Column: 5, Offset: 4}
	tok := NewToken(BEGIN, "begin", pos)

	if tok.Type != BEGIN {
		t.Errorf("NewToken Type = %v, want %v", tok.Type, BEGIN)
	}
	if tok.Literal != "begin" {
		t.Errorf("NewToken Literal = %q, want %q", tok.Literal, "begin")
	}
	if tok.Pos != pos {
		t.Errorf("NewToken Pos = %v, want %v", tok.Pos, pos)
	}
}

func TestTokenString(t *testing.T) {
	tests := []struct {
		name     string
		contains []string
		token    Token
	}{
		{
			name:     "simple token",
			token:    Token{Type: BEGIN, Literal: "begin", Pos: Position{Line: 1, Column: 1}},
			contains: []string{"BEGIN", "begin", "1:1"},
		},
		{
			name:     "EOF token",
			token:    Token{Type: EOF, Literal: "", Pos: Position{Line: 10, Column: 5}},
			contains: []string{"EOF", "10:5"},
		},
		{
			name:     "long literal truncated",
			token:    Token{Type: STRING, Literal: "this is a very long string that should be truncated", Pos: Position{Line: 2, Column: 3}},
			contains: []string{"STRING", "...", "2:3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.token.String()
			for _, substr := range tt.contains {
				if !strings.Contains(str, substr) {
					t.Errorf("Token.String() = %q, should contain %q", str, substr)
				}
			}
		})
	}
}

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		ident    string
		expected TokenType
	}{
		// Keywords - various cases
		{"begin", BEGIN},
		{"BEGIN", BEGIN},
		{"Begin", BEGIN},
		{"end", END},
		{"END", END},
		{"if", IF},
		{"IF", IF},
		{"then", THEN},
		{"else", ELSE},
		{"while", WHILE},
		{"for", FOR},
		{"to", TO},
		{"downto", DOWNTO},
		{"do", DO},
		{"repeat", REPEAT},
		{"until", UNTIL},

		// Boolean literals
		{"true", TRUE},
		{"TRUE", TRUE},
		{"True", TRUE},
		{"false", FALSE},
		{"FALSE", FALSE},
		{"False", FALSE},
		{"nil", NIL},
		{"NIL", NIL},
		{"Nil", NIL},

		// Object-oriented keywords
		{"function", FUNCTION},
		{"FUNCTION", FUNCTION},
		{"procedure", PROCEDURE},
		{"PROCEDURE", PROCEDURE},
		{"class", CLASS},
		{"CLASS", CLASS},
		{"interface", INTERFACE},
		{"constructor", CONSTRUCTOR},
		{"destructor", DESTRUCTOR},
		{"property", PROPERTY},
		{"virtual", VIRTUAL},
		{"override", OVERRIDE},
		{"abstract", ABSTRACT},
		{"inherited", INHERITED},

		// Declaration keywords
		{"var", VAR},
		{"VAR", VAR},
		{"const", CONST},
		{"type", TYPE},
		{"record", RECORD},
		{"array", ARRAY},
		{"set", SET},

		// Exception handling
		{"try", TRY},
		{"except", EXCEPT},
		{"finally", FINALLY},
		{"raise", RAISE},
		{"on", ON},

		// Logical operators
		{"and", AND},
		{"AND", AND},
		{"or", OR},
		{"OR", OR},
		{"not", NOT},
		{"NOT", NOT},
		{"xor", XOR},
		{"XOR", XOR},

		// Special keywords
		{"is", IS},
		{"as", AS},
		{"in", IN},
		{"div", DIV},
		{"DIV", DIV},
		{"mod", MOD},
		{"MOD", MOD},
		{"shl", SHL},
		{"shr", SHR},
		{"sar", SAR},

		// Modern features
		{"async", ASYNC},
		{"await", AWAIT},
		{"lambda", LAMBDA},

		// Access modifiers
		{"private", PRIVATE},
		{"protected", PROTECTED},
		{"public", PUBLIC},
		{"published", PUBLISHED},

		// Identifiers (not keywords)
		{"myVar", IDENT},
		{"MyClass", IDENT},
		{"x", IDENT},
		{"variable123", IDENT},
		{"_underscore", IDENT},
		{"camelCase", IDENT},
		{"PascalCase", IDENT},
		{"UPPERCASE", IDENT},
		{"lowercase", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.ident, func(t *testing.T) {
			if got := LookupIdent(tt.ident); got != tt.expected {
				t.Errorf("LookupIdent(%q) = %v (%s), want %v (%s)",
					tt.ident, got, got.String(), tt.expected, tt.expected.String())
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	// Test that all keywords are recognized
	keywords := []string{
		"begin", "end", "if", "then", "else", "case", "of", "while", "repeat", "until",
		"for", "to", "downto", "do", "break", "continue", "exit", "with",
		"var", "const", "type", "record", "array", "set", "enum",
		"function", "procedure", "class", "interface", "constructor", "destructor",
		"try", "except", "finally", "raise", "on",
		"true", "false", "nil",
		"and", "or", "not", "xor",
		"is", "as", "in", "div", "mod", "shl", "shr",
	}

	for _, kw := range keywords {
		t.Run(kw, func(t *testing.T) {
			// Test lowercase
			if !IsKeyword(kw) {
				t.Errorf("IsKeyword(%q) = false, want true", kw)
			}
			// Test uppercase
			upper := strings.ToUpper(kw)
			if !IsKeyword(upper) {
				t.Errorf("IsKeyword(%q) = false, want true", upper)
			}
			// Test mixed case
			mixed := strings.Title(kw)
			if !IsKeyword(mixed) {
				t.Errorf("IsKeyword(%q) = false, want true", mixed)
			}
		})
	}

	// Test non-keywords
	nonKeywords := []string{"myVar", "MyClass", "x123", "_test", "notAKeyword"}
	for _, nk := range nonKeywords {
		t.Run(nk, func(t *testing.T) {
			if IsKeyword(nk) {
				t.Errorf("IsKeyword(%q) = true, want false", nk)
			}
		})
	}
}

func TestGetKeywordLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Keywords should be lowercased
		{"BEGIN", "begin"},
		{"Begin", "begin"},
		{"begin", "begin"},
		{"END", "end"},
		{"FUNCTION", "function"},
		{"Function", "function"},
		{"TRUE", "true"},
		{"False", "false"},
		{"NIL", "nil"},

		// Non-keywords should be unchanged
		{"MyClass", "MyClass"},
		{"myVar", "myVar"},
		{"CONSTANT", "CONSTANT"},
		{"x123", "x123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := GetKeywordLiteral(tt.input); got != tt.expected {
				t.Errorf("GetKeywordLiteral(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAllKeywordsCovered(t *testing.T) {
	// Verify that all keyword TokenTypes have an entry in the keywords map
	keywordTypes := []TokenType{
		BEGIN, END, IF, THEN, ELSE, CASE, OF, WHILE, REPEAT, UNTIL,
		FOR, TO, DOWNTO, DO, BREAK, CONTINUE, EXIT, WITH, ASM,
		VAR, CONST, TYPE, RECORD, ARRAY, SET, ENUM, FLAGS,
		CLASS, OBJECT, INTERFACE, IMPLEMENTS, FUNCTION, PROCEDURE,
		CONSTRUCTOR, DESTRUCTOR, METHOD, PROPERTY, VIRTUAL, OVERRIDE,
		ABSTRACT, SEALED, STATIC, FINAL, NEW, INHERITED, REINTRODUCE,
		OPERATOR, HELPER, PARTIAL, LAZY, INDEX,
		TRY, EXCEPT, RAISE, FINALLY, ON,
		NOT, AND, OR, XOR,
		TRUE, FALSE, NIL, IS, AS, IN, DIV, MOD, SHL, SHR, SAR,
		INLINE, EXTERNAL, FORWARD, OVERLOAD, DEPRECATED, READONLY, EXPORT,
		REGISTER, PASCAL, CDECL, SAFECALL, STDCALL, FASTCALL, REFERENCE,
		PRIVATE, PROTECTED, PUBLIC, PUBLISHED, STRICT,
		READ, WRITE, DEFAULT, DESCRIPTION,
		OLD, REQUIRE, ENSURE, INVARIANTS,
		ASYNC, AWAIT, LAMBDA, IMPLIES, EMPTY, IMPLICIT,
	}

	// Check that each keyword type can be looked up
	for _, kt := range keywordTypes {
		found := false
		for kw, tokType := range keywords {
			if tokType == kt {
				found = true
				// Verify LookupIdent returns the correct type
				if got := LookupIdent(kw); got != kt {
					t.Errorf("LookupIdent(%q) = %v, want %v", kw, got, kt)
				}
				break
			}
		}
		if !found {
			t.Errorf("Keyword type %v (%s) has no entry in keywords map", kt, kt.String())
		}
	}
}

func TestKeywordCaseInsensitivity(t *testing.T) {
	// Test a sample of keywords in different cases
	testCases := []string{"begin", "BEGIN", "Begin", "BeGiN", "bEgIn"}

	for _, tc := range testCases {
		if got := LookupIdent(tc); got != BEGIN {
			t.Errorf("LookupIdent(%q) = %v, want BEGIN", tc, got)
		}
	}
}

func BenchmarkLookupIdent(b *testing.B) {
	identifiers := []string{
		"begin", "end", "if", "function", "class", "myVariable",
		"BEGIN", "END", "MyClass", "x", "variable123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ident := identifiers[i%len(identifiers)]
		LookupIdent(ident)
	}
}

func BenchmarkIsKeyword(b *testing.B) {
	words := []string{
		"begin", "end", "if", "myVariable", "BEGIN", "MyClass",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		word := words[i%len(words)]
		IsKeyword(word)
	}
}
