package lexer

import (
	"strings"
	"testing"
)

// TestErrorAccumulation tests that lexer errors are properly accumulated
// instead of stopping at the first error
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
