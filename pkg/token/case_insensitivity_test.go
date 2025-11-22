package token

import (
	"strings"
	"testing"
	"unicode"
)

// TestAllKeywordsCaseInsensitivity comprehensively tests that all DWScript keywords
// are case-insensitive. This is a critical requirement for DWScript language compatibility.
func TestAllKeywordsCaseInsensitivity(t *testing.T) {
	// Get all keywords from the keywords map
	for keyword, expectedType := range keywords {
		t.Run(keyword, func(t *testing.T) {
			// Test lowercase (original)
			t.Run("lowercase", func(t *testing.T) {
				got := LookupIdent(keyword)
				if got != expectedType {
					t.Errorf("LookupIdent(%q) = %v, want %v", keyword, got, expectedType)
				}
			})

			// Test UPPERCASE
			t.Run("UPPERCASE", func(t *testing.T) {
				upper := strings.ToUpper(keyword)
				got := LookupIdent(upper)
				if got != expectedType {
					t.Errorf("LookupIdent(%q) = %v, want %v", upper, got, expectedType)
				}
			})

			// Test MixedCase (capitalize first letter)
			t.Run("MixedCase", func(t *testing.T) {
				mixed := capitalizeFirst(keyword)
				got := LookupIdent(mixed)
				if got != expectedType {
					t.Errorf("LookupIdent(%q) = %v, want %v", mixed, got, expectedType)
				}
			})

			// Test aLtErNaTiNg case
			t.Run("aLtErNaTiNg", func(t *testing.T) {
				alternating := alternatingCase(keyword)
				got := LookupIdent(alternating)
				if got != expectedType {
					t.Errorf("LookupIdent(%q) = %v, want %v", alternating, got, expectedType)
				}
			})
		})
	}
}

// TestOriginalCasingPreservedInTokenLiteral verifies that the lexer preserves
// the original casing in the token literal even though it correctly identifies
// the keyword type. This is critical for error messages to show the user's
// original code.
func TestOriginalCasingPreservedInTokenLiteral(t *testing.T) {
	testCases := []struct {
		input        string
		expectedType TokenType
	}{
		// Control flow keywords
		{"BEGIN", BEGIN},
		{"Begin", BEGIN},
		{"begin", BEGIN},
		{"bEgIn", BEGIN},
		{"END", END},
		{"End", END},
		{"IF", IF},
		{"If", IF},
		{"THEN", THEN},
		{"Then", THEN},
		{"ELSE", ELSE},
		{"Else", ELSE},
		{"WHILE", WHILE},
		{"While", WHILE},
		{"FOR", FOR},
		{"For", FOR},

		// Declaration keywords
		{"VAR", VAR},
		{"Var", VAR},
		{"CONST", CONST},
		{"Const", CONST},
		{"FUNCTION", FUNCTION},
		{"Function", FUNCTION},
		{"PROCEDURE", PROCEDURE},
		{"Procedure", PROCEDURE},
		{"CLASS", CLASS},
		{"Class", CLASS},
		{"RECORD", RECORD},
		{"Record", RECORD},

		// Boolean/Logical keywords
		{"NOT", NOT},
		{"Not", NOT},
		{"AND", AND},
		{"And", AND},
		{"OR", OR},
		{"Or", OR},
		{"XOR", XOR},
		{"Xor", XOR},

		// Boolean literals
		{"TRUE", TRUE},
		{"True", TRUE},
		{"FALSE", FALSE},
		{"False", FALSE},
		{"NIL", NIL},
		{"Nil", NIL},

		// Operators
		{"DIV", DIV},
		{"Div", DIV},
		{"MOD", MOD},
		{"Mod", MOD},
		{"SHL", SHL},
		{"Shl", SHL},
		{"SHR", SHR},
		{"Shr", SHR},

		// Type checking
		{"IS", IS},
		{"Is", IS},
		{"AS", AS},
		{"As", AS},
		{"IN", IN},
		{"In", IN},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// Create a token as the lexer would
			tok := NewToken(LookupIdent(tc.input), tc.input, Position{Line: 1, Column: 1})

			// Verify token type is correct
			if tok.Type != tc.expectedType {
				t.Errorf("Token type = %v, want %v", tok.Type, tc.expectedType)
			}

			// Verify original casing is preserved in literal
			if tok.Literal != tc.input {
				t.Errorf("Token literal = %q, want %q (original casing lost!)", tok.Literal, tc.input)
			}
		})
	}
}

// TestKeywordIdentifierBoundary tests edge cases at the keyword/identifier boundary
func TestKeywordIdentifierBoundary(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		expectedType TokenType
	}{
		// These should be keywords
		{"begin_keyword", "begin", BEGIN},
		{"BEGIN_keyword", "BEGIN", BEGIN},
		{"if_keyword", "if", IF},
		{"IF_keyword", "IF", IF},

		// These should be identifiers (not keywords)
		{"begin123", "begin123", IDENT},
		{"BEGIN123", "BEGIN123", IDENT},
		{"ifx", "ifx", IDENT},
		{"IFX", "IFX", IDENT},
		{"_begin", "_begin", IDENT},
		{"beginEnd", "beginEnd", IDENT},
		{"endif", "endif", IDENT}, // Not a DWScript keyword

		// Identifiers that contain keyword substrings
		{"mybegin", "mybegin", IDENT},
		{"MyBegin", "MyBegin", IDENT},
		{"ifelse", "ifelse", IDENT},
		{"whileloop", "whileloop", IDENT},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := LookupIdent(tc.input)
			if got != tc.expectedType {
				t.Errorf("LookupIdent(%q) = %v, want %v", tc.input, got, tc.expectedType)
			}
		})
	}
}

// TestIsKeywordCaseInsensitive verifies IsKeyword function is case-insensitive
func TestIsKeywordCaseInsensitive(t *testing.T) {
	// Sample keywords to test
	sampleKeywords := []string{
		"begin", "end", "if", "then", "else", "while", "for", "function",
		"procedure", "class", "var", "const", "true", "false", "nil",
		"and", "or", "not", "div", "mod",
	}

	for _, kw := range sampleKeywords {
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
		mixed := capitalizeFirst(kw)
		if !IsKeyword(mixed) {
			t.Errorf("IsKeyword(%q) = false, want true", mixed)
		}
	}

	// Test non-keywords
	nonKeywords := []string{
		"myVar", "MYVAR", "MyVar",
		"begin123", "ifx", "thenelse",
	}
	for _, nk := range nonKeywords {
		if IsKeyword(nk) {
			t.Errorf("IsKeyword(%q) = true, want false", nk)
		}
	}
}

// TestGetKeywordLiteralCaseInsensitive verifies GetKeywordLiteral returns
// canonical (lowercase) form regardless of input casing
func TestGetKeywordLiteralCaseInsensitive(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"begin", "begin"},
		{"BEGIN", "begin"},
		{"Begin", "begin"},
		{"bEgIn", "begin"},
		{"FUNCTION", "function"},
		{"Function", "function"},
		{"CLASS", "class"},
		{"Class", "class"},
		// Non-keywords should return original
		{"MyClass", "MyClass"},
		{"MYVAR", "MYVAR"},
		{"myVar", "myVar"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := GetKeywordLiteral(tc.input)
			if got != tc.expected {
				t.Errorf("GetKeywordLiteral(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// TestTokenTypeConsistency verifies that keyword token types are consistent
// regardless of input casing
func TestTokenTypeConsistency(t *testing.T) {
	// For each keyword, verify that all case variations return the same type
	for keyword := range keywords {
		lower := keyword
		upper := strings.ToUpper(keyword)
		mixed := capitalizeFirst(keyword)

		lowerType := LookupIdent(lower)
		upperType := LookupIdent(upper)
		mixedType := LookupIdent(mixed)

		if lowerType != upperType {
			t.Errorf("Inconsistent types for %q: lowercase=%v, uppercase=%v", keyword, lowerType, upperType)
		}
		if lowerType != mixedType {
			t.Errorf("Inconsistent types for %q: lowercase=%v, mixed=%v", keyword, lowerType, mixedType)
		}
	}
}

// Helper function to capitalize first letter
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// Helper function to create alternating case
func alternatingCase(s string) string {
	runes := []rune(s)
	for i := range runes {
		if i%2 == 0 {
			runes[i] = unicode.ToLower(runes[i])
		} else {
			runes[i] = unicode.ToUpper(runes[i])
		}
	}
	return string(runes)
}
