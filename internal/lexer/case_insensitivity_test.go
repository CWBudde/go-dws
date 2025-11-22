package lexer

import (
	"strings"
	"testing"
	"unicode"
)

// TestLexerKeywordCaseInsensitivity comprehensively tests that the lexer
// correctly identifies keywords regardless of their casing in the source code.
func TestLexerKeywordCaseInsensitivity(t *testing.T) {
	// Sample of keywords to test through the full lexer pipeline
	keywords := []struct {
		canonical    string
		expectedType TokenType
	}{
		{"begin", BEGIN},
		{"end", END},
		{"if", IF},
		{"then", THEN},
		{"else", ELSE},
		{"while", WHILE},
		{"for", FOR},
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
		{"div", DIV},
		{"mod", MOD},
	}

	for _, kw := range keywords {
		t.Run(kw.canonical, func(t *testing.T) {
			// Test lowercase
			testLexerKeywordCase(t, kw.canonical, kw.expectedType)

			// Test UPPERCASE
			testLexerKeywordCase(t, strings.ToUpper(kw.canonical), kw.expectedType)

			// Test MixedCase
			testLexerKeywordCase(t, capitalizeFirst(kw.canonical), kw.expectedType)

			// Test aLtErNaTiNg
			testLexerKeywordCase(t, alternatingCase(kw.canonical), kw.expectedType)
		})
	}
}

// testLexerKeywordCase is a helper that tests a keyword through the full lexer pipeline
func testLexerKeywordCase(t *testing.T, input string, expectedType TokenType) {
	t.Helper()

	l := New(input)
	tok := l.NextToken()

	if tok.Type != expectedType {
		t.Errorf("Lexer: input %q got type %v, want %v", input, tok.Type, expectedType)
	}

	// Critical: Verify original casing is preserved in literal
	if tok.Literal != input {
		t.Errorf("Lexer: input %q literal = %q (original casing lost!)", input, tok.Literal)
	}
}

// TestLexerPreservesOriginalCasing verifies that the lexer preserves
// the original casing in token literals for both keywords and identifiers.
// This is critical for error messages to show the user's original code.
func TestLexerPreservesOriginalCasing(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		expectedType    TokenType
		expectedLiteral string
	}{
		// Keywords in various cases - literal should match input exactly
		{"BEGIN_upper", "BEGIN", BEGIN, "BEGIN"},
		{"Begin_mixed", "Begin", BEGIN, "Begin"},
		{"begin_lower", "begin", BEGIN, "begin"},
		{"bEgIn_alternating", "bEgIn", BEGIN, "bEgIn"},

		{"FUNCTION_upper", "FUNCTION", FUNCTION, "FUNCTION"},
		{"Function_mixed", "Function", FUNCTION, "Function"},
		{"FuNcTiOn_alternating", "FuNcTiOn", FUNCTION, "FuNcTiOn"},

		{"TRUE_upper", "TRUE", TRUE, "TRUE"},
		{"True_mixed", "True", TRUE, "True"},
		{"true_lower", "true", TRUE, "true"},

		// Identifiers - literal should always match input
		{"MyVariable_mixed", "MyVariable", IDENT, "MyVariable"},
		{"MYVARIABLE_upper", "MYVARIABLE", IDENT, "MYVARIABLE"},
		{"myvariable_lower", "myvariable", IDENT, "myvariable"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := New(tc.input)
			tok := l.NextToken()

			if tok.Type != tc.expectedType {
				t.Errorf("Token type = %v, want %v", tok.Type, tc.expectedType)
			}

			if tok.Literal != tc.expectedLiteral {
				t.Errorf("Token literal = %q, want %q", tok.Literal, tc.expectedLiteral)
			}
		})
	}
}

// TestLexerMixedCaseProgram tests a realistic program with mixed case keywords
func TestLexerMixedCaseProgram(t *testing.T) {
	// This is valid DWScript with mixed case keywords
	input := `
		Var x: Integer;
		BEGIN
			If x > 0 Then
				PrintLn('positive')
			Else
				PrintLn('non-positive');
		END;
	`

	expectedTokens := []struct {
		tokenType TokenType
		literal   string
	}{
		{VAR, "Var"},
		{IDENT, "x"},
		{COLON, ":"},
		{IDENT, "Integer"},
		{SEMICOLON, ";"},
		{BEGIN, "BEGIN"},
		{IF, "If"},
		{IDENT, "x"},
		{GREATER, ">"},
		{INT, "0"},
		{THEN, "Then"},
		{IDENT, "PrintLn"},
		{LPAREN, "("},
		{STRING, "positive"},
		{RPAREN, ")"},
		{ELSE, "Else"},
		{IDENT, "PrintLn"},
		{LPAREN, "("},
		{STRING, "non-positive"},
		{RPAREN, ")"},
		{SEMICOLON, ";"},
		{END, "END"},
		{SEMICOLON, ";"},
		{EOF, ""},
	}

	l := New(input)

	for i, expected := range expectedTokens {
		tok := l.NextToken()

		if tok.Type != expected.tokenType {
			t.Errorf("token[%d]: type = %v, want %v (literal=%q)", i, tok.Type, expected.tokenType, tok.Literal)
		}

		if tok.Literal != expected.literal {
			t.Errorf("token[%d]: literal = %q, want %q", i, tok.Literal, expected.literal)
		}
	}
}

// TestLexerKeywordIdentifierBoundary tests that keywords with suffixes become identifiers
func TestLexerKeywordIdentifierBoundary(t *testing.T) {
	testCases := []struct {
		input        string
		expectedType TokenType
	}{
		// These should be keywords
		{"begin", BEGIN},
		{"BEGIN", BEGIN},
		{"if", IF},
		{"IF", IF},

		// These should be identifiers (not keywords)
		{"begin123", IDENT},
		{"BEGIN123", IDENT},
		{"ifx", IDENT},
		{"IFX", IDENT},
		{"_begin", IDENT},
		{"beginEnd", IDENT},
		{"endif", IDENT}, // Not a DWScript keyword

		// Identifiers that contain keyword substrings
		{"mybegin", IDENT},
		{"MyBegin", IDENT},
		{"ifelse", IDENT},
		{"whileloop", IDENT},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			l := New(tc.input)
			tok := l.NextToken()

			if tok.Type != tc.expectedType {
				t.Errorf("Lexer input %q: got type %v, want %v", tc.input, tok.Type, tc.expectedType)
			}

			// Literal should always match input
			if tok.Literal != tc.input {
				t.Errorf("Lexer input %q: literal = %q (original casing lost!)", tc.input, tok.Literal)
			}
		})
	}
}

// TestLexerMultipleKeywordsSameProgram tests that multiple keywords in various
// cases are all correctly identified in a single program
func TestLexerMultipleKeywordsSameProgram(t *testing.T) {
	// Mix of uppercase, lowercase, and mixed case keywords
	input := "BEGIN begin Begin bEgIn END end End eNd"

	expectedTypes := []TokenType{
		BEGIN, BEGIN, BEGIN, BEGIN, // All variations of begin
		END, END, END, END, // All variations of end
		EOF,
	}

	expectedLiterals := []string{
		"BEGIN", "begin", "Begin", "bEgIn",
		"END", "end", "End", "eNd",
		"",
	}

	l := New(input)

	for i := 0; i < len(expectedTypes); i++ {
		tok := l.NextToken()

		if tok.Type != expectedTypes[i] {
			t.Errorf("token[%d]: type = %v, want %v", i, tok.Type, expectedTypes[i])
		}

		if tok.Literal != expectedLiterals[i] {
			t.Errorf("token[%d]: literal = %q, want %q", i, tok.Literal, expectedLiterals[i])
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
