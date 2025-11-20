package token

import (
	"testing"
)

// TestPositionString tests Position.String()
func TestPositionString(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		pos      Position
	}{
		{"simple position", Position{Line: 1, Column: 5}, "1:5"},
		{"larger numbers", Position{Line: 123, Column: 456}, "123:456"},
		{"zero position", Position{Line: 0, Column: 0}, "0:0"},
		{"with offset", Position{Line: 10, Column: 20, Offset: 100}, "10:20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pos.String()
			if got != tt.expected {
				t.Errorf("Position.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestPositionIsValid tests Position.IsValid()
func TestPositionIsValid(t *testing.T) {
	tests := []struct {
		name     string
		pos      Position
		expected bool
	}{
		{"valid position", Position{Line: 1, Column: 1}, true},
		{"valid with offset", Position{Line: 10, Column: 5, Offset: 50}, true},
		{"zero line invalid", Position{Line: 0, Column: 1}, false},
		{"negative line invalid", Position{Line: -1, Column: 1}, false},
		{"zero column but valid line", Position{Line: 1, Column: 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pos.IsValid()
			if got != tt.expected {
				t.Errorf("Position.IsValid() = %v, want %v (pos: %+v)", got, tt.expected, tt.pos)
			}
		})
	}
}

// TestTokenString tests Token.String()
func TestTokenString(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		token    Token
	}{
		{
			"simple identifier",
			Token{Type: IDENT, Literal: "foo", Pos: Position{Line: 1, Column: 5}},
			`IDENT("foo") at 1:5`,
		},
		{
			"keyword",
			Token{Type: BEGIN, Literal: "begin", Pos: Position{Line: 2, Column: 1}},
			`BEGIN("begin") at 2:1`,
		},
		{
			"EOF token",
			Token{Type: EOF, Literal: "", Pos: Position{Line: 10, Column: 20}},
			`EOF at 10:20`,
		},
		{
			"long literal truncated",
			Token{Type: STRING, Literal: "this is a very long string literal that will be truncated", Pos: Position{Line: 5, Column: 10}},
			`STRING("this is a very long "...) at 5:10`,
		},
		{
			"operator",
			Token{Type: PLUS, Literal: "+", Pos: Position{Line: 3, Column: 7}},
			`PLUS("+") at 3:7`,
		},
		{
			"integer literal",
			Token{Type: INT, Literal: "42", Pos: Position{Line: 1, Column: 1}},
			`INT("42") at 1:1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.String()
			if got != tt.expected {
				t.Errorf("Token.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestNewToken tests NewToken()
func TestNewToken(t *testing.T) {
	pos := Position{Line: 5, Column: 10, Offset: 100}
	tok := NewToken(IDENT, "myVar", pos)

	if tok.Type != IDENT {
		t.Errorf("NewToken() Type = %v, want %v", tok.Type, IDENT)
	}
	if tok.Literal != "myVar" {
		t.Errorf("NewToken() Literal = %q, want %q", tok.Literal, "myVar")
	}
	if tok.Pos != pos {
		t.Errorf("NewToken() Pos = %+v, want %+v", tok.Pos, pos)
	}
}

// TestTokenTypeString tests TokenType.String()
func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		tt       TokenType
	}{
		{"ILLEGAL", ILLEGAL, "ILLEGAL"},
		{"EOF", EOF, "EOF"},
		{"IDENT", IDENT, "IDENT"},
		{"INT", INT, "INT"},
		{"FLOAT", FLOAT, "FLOAT"},
		{"STRING", STRING, "STRING"},
		{"BEGIN", BEGIN, "BEGIN"},
		{"END", END, "END"},
		{"IF", IF, "IF"},
		{"WHILE", WHILE, "WHILE"},
		{"CLASS", CLASS, "CLASS"},
		{"FUNCTION", FUNCTION, "FUNCTION"},
		{"PLUS", PLUS, "PLUS"},
		{"MINUS", MINUS, "MINUS"},
		{"LPAREN", LPAREN, "LPAREN"},
		{"RPAREN", RPAREN, "RPAREN"},
		{"ASSIGN", ASSIGN, "ASSIGN"},
		{"EQ", EQ, "EQ"},
		{"NOT_EQ", NOT_EQ, "NOT_EQ"},
		{"SWITCH", SWITCH, "SWITCH"},
		{"unknown type", TokenType(9999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tt.String()
			if got != tt.expected {
				t.Errorf("TokenType.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestTokenTypeIsLiteral tests TokenType.IsLiteral()
func TestTokenTypeIsLiteral(t *testing.T) {
	tests := []struct {
		name     string
		tt       TokenType
		expected bool
	}{
		{"IDENT is literal", IDENT, true},
		{"INT is literal", INT, true},
		{"FLOAT is literal", FLOAT, true},
		{"STRING is literal", STRING, true},
		{"CHAR is literal", CHAR, true},
		{"COMMENT is literal", COMMENT, true},
		// Note: TRUE, FALSE, NIL, NULL, UNASSIGNED are keywords, not literals
		{"TRUE is not literal", TRUE, false},
		{"FALSE is not literal", FALSE, false},
		{"NIL is not literal", NIL, false},
		{"NULL is not literal", NULL, false},
		{"UNASSIGNED is not literal", UNASSIGNED, false},
		{"BEGIN is not literal", BEGIN, false},
		{"IF is not literal", IF, false},
		{"PLUS is not literal", PLUS, false},
		{"LPAREN is not literal", LPAREN, false},
		{"EOF is not literal", EOF, false},
		{"ILLEGAL is not literal", ILLEGAL, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tt.IsLiteral()
			if got != tt.expected {
				t.Errorf("TokenType(%v).IsLiteral() = %v, want %v", tt.tt, got, tt.expected)
			}
		})
	}
}

// TestTokenTypeIsKeyword tests TokenType.IsKeyword()
func TestTokenTypeIsKeyword(t *testing.T) {
	tests := []struct {
		name     string
		tt       TokenType
		expected bool
	}{
		{"BEGIN is keyword", BEGIN, true},
		{"END is keyword", END, true},
		{"IF is keyword", IF, true},
		{"WHILE is keyword", WHILE, true},
		{"CLASS is keyword", CLASS, true},
		{"FUNCTION is keyword", FUNCTION, true},
		{"VAR is keyword", VAR, true},
		{"CONST is keyword", CONST, true},
		{"NOT is keyword", NOT, true},
		{"AND is keyword", AND, true},
		// Boolean/nil literal keywords
		{"TRUE is keyword", TRUE, true},
		{"FALSE is keyword", FALSE, true},
		{"NIL is keyword", NIL, true},
		{"NULL is keyword", NULL, true},
		{"UNASSIGNED is keyword", UNASSIGNED, true},
		{"IDENT is not keyword", IDENT, false},
		{"INT is not keyword", INT, false},
		{"COMMENT is not keyword", COMMENT, false},
		{"PLUS is not keyword", PLUS, false},
		{"LPAREN is not keyword", LPAREN, false},
		{"EOF is not keyword", EOF, false},
		{"ILLEGAL is not keyword", ILLEGAL, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tt.IsKeyword()
			if got != tt.expected {
				t.Errorf("TokenType(%v).IsKeyword() = %v, want %v", tt.tt, got, tt.expected)
			}
		})
	}
}

// TestTokenTypeIsOperator tests TokenType.IsOperator()
func TestTokenTypeIsOperator(t *testing.T) {
	tests := []struct {
		name     string
		tt       TokenType
		expected bool
	}{
		{"PLUS is operator", PLUS, true},
		{"MINUS is operator", MINUS, true},
		{"ASTERISK is operator", ASTERISK, true},
		{"SLASH is operator", SLASH, true},
		{"EQ is operator", EQ, true},
		{"NOT_EQ is operator", NOT_EQ, true},
		{"LESS is operator", LESS, true},
		{"GREATER is operator", GREATER, true},
		{"ASSIGN is operator", ASSIGN, true},
		{"PLUS_ASSIGN is operator", PLUS_ASSIGN, true},
		{"FAT_ARROW is operator", FAT_ARROW, true},
		{"IDENT is not operator", IDENT, false},
		{"BEGIN is not operator", BEGIN, false},
		{"LPAREN is not operator", LPAREN, false},
		{"EOF is not operator", EOF, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tt.IsOperator()
			if got != tt.expected {
				t.Errorf("TokenType(%v).IsOperator() = %v, want %v", tt.tt, got, tt.expected)
			}
		})
	}
}

// TestTokenTypeIsDelimiter tests TokenType.IsDelimiter()
func TestTokenTypeIsDelimiter(t *testing.T) {
	tests := []struct {
		name     string
		tt       TokenType
		expected bool
	}{
		{"LPAREN is delimiter", LPAREN, true},
		{"RPAREN is delimiter", RPAREN, true},
		{"LBRACK is delimiter", LBRACK, true},
		{"RBRACK is delimiter", RBRACK, true},
		{"LBRACE is delimiter", LBRACE, true},
		{"RBRACE is delimiter", RBRACE, true},
		{"SEMICOLON is delimiter", SEMICOLON, true},
		{"COMMA is delimiter", COMMA, true},
		{"DOT is delimiter", DOT, true},
		{"COLON is delimiter", COLON, true},
		{"DOTDOT is delimiter", DOTDOT, true},
		{"IDENT is not delimiter", IDENT, false},
		{"BEGIN is not delimiter", BEGIN, false},
		{"PLUS is not delimiter", PLUS, false},
		{"EOF is not delimiter", EOF, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tt.IsDelimiter()
			if got != tt.expected {
				t.Errorf("TokenType(%v).IsDelimiter() = %v, want %v", tt.tt, got, tt.expected)
			}
		})
	}
}

// TestLookupIdent tests LookupIdent()
func TestLookupIdent(t *testing.T) {
	tests := []struct {
		name     string
		ident    string
		expected TokenType
	}{
		// Keywords (case-insensitive)
		{"begin lowercase", "begin", BEGIN},
		{"BEGIN uppercase", "BEGIN", BEGIN},
		{"Begin mixed case", "Begin", BEGIN},
		{"end lowercase", "end", END},
		{"if lowercase", "if", IF},
		{"IF uppercase", "IF", IF},
		{"while lowercase", "while", WHILE},
		{"WHILE uppercase", "WHILE", WHILE},
		{"class lowercase", "class", CLASS},
		{"CLASS uppercase", "CLASS", CLASS},
		{"function lowercase", "function", FUNCTION},
		{"var lowercase", "var", VAR},
		{"const lowercase", "const", CONST},
		{"not lowercase", "not", NOT},
		{"and lowercase", "and", AND},
		{"or lowercase", "or", OR},
		{"xor lowercase", "xor", XOR},
		{"div lowercase", "div", DIV},
		{"mod lowercase", "mod", MOD},
		{"true lowercase", "true", TRUE},
		{"false lowercase", "false", FALSE},
		{"nil lowercase", "nil", NIL},

		// Identifiers (not keywords)
		{"regular identifier", "myVariable", IDENT},
		{"identifier with numbers", "var123", IDENT},
		{"identifier starting with underscore", "_test", IDENT},
		{"CamelCase identifier", "MyClass", IDENT},
		{"all caps non-keyword", "VARIABLE", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupIdent(tt.ident)
			if got != tt.expected {
				t.Errorf("LookupIdent(%q) = %v, want %v", tt.ident, got, tt.expected)
			}
		})
	}
}

// TestIsKeyword tests IsKeyword()
func TestIsKeyword(t *testing.T) {
	tests := []struct {
		name     string
		ident    string
		expected bool
	}{
		// Keywords (case-insensitive)
		{"begin lowercase", "begin", true},
		{"BEGIN uppercase", "BEGIN", true},
		{"Begin mixed case", "Begin", true},
		{"if lowercase", "if", true},
		{"IF uppercase", "IF", true},
		{"while lowercase", "while", true},
		{"class lowercase", "class", true},
		{"function lowercase", "function", true},
		{"var lowercase", "var", true},
		{"not lowercase", "not", true},
		{"and lowercase", "and", true},
		{"true lowercase", "true", true},
		{"false lowercase", "false", true},
		{"nil lowercase", "nil", true},

		// Non-keywords
		{"regular identifier", "myVariable", false},
		{"identifier with numbers", "var123", false},
		{"identifier starting with underscore", "_test", false},
		{"empty string", "", false},
		{"all caps non-keyword", "VARIABLE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsKeyword(tt.ident)
			if got != tt.expected {
				t.Errorf("IsKeyword(%q) = %v, want %v", tt.ident, got, tt.expected)
			}
		})
	}
}

// TestGetKeywordLiteral tests GetKeywordLiteral()
func TestGetKeywordLiteral(t *testing.T) {
	tests := []struct {
		name     string
		ident    string
		expected string
	}{
		// Keywords return lowercase canonical form
		{"begin lowercase", "begin", "begin"},
		{"BEGIN uppercase", "BEGIN", "begin"},
		{"Begin mixed case", "Begin", "begin"},
		{"IF uppercase", "IF", "if"},
		{"WhIlE mixed case", "WhIlE", "while"},
		{"CLASS uppercase", "CLASS", "class"},
		{"FuNcTiOn mixed case", "FuNcTiOn", "function"},
		{"TRUE uppercase", "TRUE", "true"},
		{"FaLsE mixed case", "FaLsE", "false"},
		{"NIL uppercase", "NIL", "nil"},

		// Non-keywords return original string
		{"regular identifier", "myVariable", "myVariable"},
		{"identifier with numbers", "var123", "var123"},
		{"CAPS identifier", "VARIABLE", "VARIABLE"},
		{"empty string", "", ""},
		{"mixed case non-keyword", "MyClass", "MyClass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetKeywordLiteral(tt.ident)
			if got != tt.expected {
				t.Errorf("GetKeywordLiteral(%q) = %q, want %q", tt.ident, got, tt.expected)
			}
		})
	}
}

// TestTokenTypeCategories tests that token types are properly categorized
func TestTokenTypeCategories(t *testing.T) {
	// Test that categories are mutually exclusive
	t.Run("categories are mutually exclusive", func(t *testing.T) {
		// Sample a few tokens from each category
		samples := []TokenType{
			IDENT, INT, STRING, // literals
			BEGIN, IF, CLASS, // keywords
			PLUS, EQ, ASSIGN, // operators
			LPAREN, COMMA, DOT, // delimiters
		}

		for _, tt := range samples {
			categories := 0
			if tt.IsLiteral() {
				categories++
			}
			if tt.IsKeyword() {
				categories++
			}
			if tt.IsOperator() {
				categories++
			}
			if tt.IsDelimiter() {
				categories++
			}

			// Each token should belong to at most one category
			if categories > 1 {
				t.Errorf("TokenType %v belongs to %d categories (should be 0 or 1)", tt, categories)
			}
		}
	})

	// Test special tokens (ILLEGAL and EOF only, COMMENT is considered a literal)
	t.Run("special tokens not in categories", func(t *testing.T) {
		special := []TokenType{ILLEGAL, EOF}
		for _, tt := range special {
			if tt.IsLiteral() || tt.IsKeyword() || tt.IsOperator() || tt.IsDelimiter() {
				t.Errorf("Special token %v should not be in any category", tt)
			}
		}
	})

	// Test that COMMENT is considered a literal
	t.Run("COMMENT is a literal", func(t *testing.T) {
		if !COMMENT.IsLiteral() {
			t.Error("COMMENT should be considered a literal")
		}
	})
}

// TestAllKeywordsInMap tests that all keyword token types are in the keywords map
func TestAllKeywordsInMap(t *testing.T) {
	// This is a sanity check to ensure consistency
	// Sample keywords that should definitely be in the map
	mustHaveKeywords := map[string]TokenType{
		"begin":    BEGIN,
		"end":      END,
		"if":       IF,
		"then":     THEN,
		"else":     ELSE,
		"while":    WHILE,
		"for":      FOR,
		"class":    CLASS,
		"function": FUNCTION,
		"var":      VAR,
		"const":    CONST,
		"not":      NOT,
		"and":      AND,
		"or":       OR,
		"true":     TRUE,
		"false":    FALSE,
		"nil":      NIL,
	}

	for kw, expectedType := range mustHaveKeywords {
		if got, ok := keywords[kw]; !ok {
			t.Errorf("Keyword %q not found in keywords map", kw)
		} else if got != expectedType {
			t.Errorf("Keyword %q maps to %v, want %v", kw, got, expectedType)
		}
	}
}
