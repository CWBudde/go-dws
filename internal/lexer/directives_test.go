// Package lexer provides lexical analysis for DWScript source code.
// This file contains tests for compiler directive support.
package lexer

import (
	"testing"
)

// TestCompilerDirectiveDefine tests {$DEFINE} directives.
func TestCompilerDirectiveDefine(t *testing.T) {
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

// TestCompilerDirectiveIfDef tests {$IFDEF} and {$IFNDEF} directives.
func TestCompilerDirectiveIfDef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name: "IFDEF defined symbol",
			input: `{$DEFINE DEBUG}
{$IFDEF DEBUG}
x := 1;
{$ENDIF}`,
			expected: []TokenType{IDENT, ASSIGN, INT, SEMICOLON, EOF},
		},
		{
			name: "IFDEF undefined symbol",
			input: `{$IFDEF NOTDEFINED}
x := 1;
{$ENDIF}
y := 2;`,
			expected: []TokenType{IDENT, ASSIGN, INT, SEMICOLON, EOF},
		},
		{
			name: "IFNDEF undefined symbol",
			input: `{$IFNDEF NOTDEFINED}
x := 1;
{$ENDIF}`,
			expected: []TokenType{IDENT, ASSIGN, INT, SEMICOLON, EOF},
		},
		{
			name: "IFNDEF defined symbol",
			input: `{$DEFINE DEBUG}
{$IFNDEF DEBUG}
x := 1;
{$ENDIF}
y := 2;`,
			expected: []TokenType{IDENT, ASSIGN, INT, SEMICOLON, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			for i, expectedType := range tt.expected {
				tok := l.NextToken()
				if tok.Type != expectedType {
					t.Errorf("token[%d] - wrong type. expected=%q, got=%q (literal=%q)",
						i, expectedType, tok.Type, tok.Literal)
				}
			}
		})
	}
}

// TestCompilerDirectiveElse tests {$ELSE} directives.
func TestCompilerDirectiveElse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "IFDEF with ELSE - defined",
			input: `{$DEFINE DEBUG}
{$IFDEF DEBUG}
x := 1;
{$ELSE}
y := 2;
{$ENDIF}`,
			expected: []string{"x"},
		},
		{
			name: "IFDEF with ELSE - undefined",
			input: `{$IFDEF NOTDEFINED}
x := 1;
{$ELSE}
y := 2;
{$ENDIF}`,
			expected: []string{"y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			var identifiers []string
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
				if tok.Type == IDENT {
					identifiers = append(identifiers, tok.Literal)
				}
			}

			if len(identifiers) != len(tt.expected) {
				t.Fatalf("wrong number of identifiers. expected=%d, got=%d",
					len(tt.expected), len(identifiers))
			}

			for i, expected := range tt.expected {
				if identifiers[i] != expected {
					t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
						i, expected, identifiers[i])
				}
			}
		})
	}
}

// TestCompilerDirectiveNested tests nested {$IFDEF} blocks.
func TestCompilerDirectiveNested(t *testing.T) {
	input := `{$DEFINE A}
{$DEFINE B}
{$IFDEF A}
x := 1;
{$IFDEF B}
y := 2;
{$ENDIF}
z := 3;
{$ENDIF}`

	expected := []string{"x", "y", "z"}

	l := New(input)

	var identifiers []string
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == IDENT {
			identifiers = append(identifiers, tok.Literal)
		}
	}

	if len(identifiers) != len(expected) {
		t.Fatalf("wrong number of identifiers. expected=%d, got=%d",
			len(expected), len(identifiers))
	}

	for i, exp := range expected {
		if identifiers[i] != exp {
			t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
				i, exp, identifiers[i])
		}
	}
}

// TestCompilerDirectiveUndef tests {$UNDEF} directives.
func TestCompilerDirectiveUndef(t *testing.T) {
	input := `{$DEFINE DEBUG}
{$IFDEF DEBUG}
x := 1;
{$ENDIF}
{$UNDEF DEBUG}
{$IFDEF DEBUG}
y := 2;
{$ENDIF}
z := 3;`

	expected := []string{"x", "z"}

	l := New(input)

	var identifiers []string
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == IDENT {
			identifiers = append(identifiers, tok.Literal)
		}
	}

	if len(identifiers) != len(expected) {
		t.Fatalf("wrong number of identifiers. expected=%d, got=%d",
			len(expected), len(identifiers))
	}

	for i, exp := range expected {
		if identifiers[i] != exp {
			t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
				i, exp, identifiers[i])
		}
	}
}

// TestCompilerDirectiveIf tests {$IF} expression directives.
func TestCompilerDirectiveIf(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "$IF with defined()",
			input: `{$DEFINE DEBUG}
{$IF defined(DEBUG)}
x := 1;
{$ENDIF}`,
			expected: []string{"x"},
		},
		{
			name: "$IF with not defined()",
			input: `{$IF not defined(NOTDEFINED)}
x := 1;
{$ENDIF}`,
			expected: []string{"x"},
		},
		{
			name: "$IF with boolean operators",
			input: `{$DEFINE A}
{$DEFINE B}
{$IF defined(A) and defined(B)}
x := 1;
{$ENDIF}`,
			expected: []string{"x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			var identifiers []string
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
				if tok.Type == IDENT {
					identifiers = append(identifiers, tok.Literal)
				}
			}

			if len(identifiers) != len(tt.expected) {
				t.Fatalf("wrong number of identifiers. expected=%v, got=%v",
					tt.expected, identifiers)
			}

			for i, exp := range tt.expected {
				if identifiers[i] != exp {
					t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
						i, exp, identifiers[i])
				}
			}
		})
	}
}

// TestCompilerDirectiveErrors tests error handling in directives.
func TestCompilerDirectiveErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "unterminated directive",
			input:         "{$DEFINE DEBUG",
			expectedError: "unterminated compiler directive",
		},
		{
			name:          "empty directive",
			input:         "{$}",
			expectedError: "empty compiler directive",
		},
		{
			name:          "DEFINE without name",
			input:         "{$DEFINE}\nx := 1;",
			expectedError: "name expected after $define",
		},
		{
			name:          "UNDEF without name",
			input:         "{$UNDEF}\nx := 1;",
			expectedError: "name expected after $undef",
		},
		{
			name:          "IFDEF without name",
			input:         "{$IFDEF}\nx := 1;",
			expectedError: "name expected after $ifdef",
		},
		{
			name:          "unbalanced ELSE",
			input:         "{$ELSE}\nx := 1;",
			expectedError: "unbalanced conditional directive",
		},
		{
			name:          "unbalanced ENDIF",
			input:         "{$ENDIF}\nx := 1;",
			expectedError: "unbalanced conditional directive",
		},
		{
			name:          "double ELSE",
			input:         "{$IFDEF DEBUG}\nx := 1;\n{$ELSE}\ny := 2;\n{$ELSE}\nz := 3;\n{$ENDIF}",
			expectedError: "unfinished conditional directive",
		},
		{
			name:          "unknown directive",
			input:         "{$UNKNOWN}\nx := 1;",
			expectedError: "unknown compiler directive: unknown",
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
				t.Fatalf("expected error containing %q, but got no errors", tt.expectedError)
			}

			found := false
			for _, err := range errors {
				if contains(err.Message, tt.expectedError) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error containing %q, but got errors: %v",
					tt.expectedError, errors)
			}
		})
	}
}

// TestCompilerDirectiveCaseInsensitive tests case insensitivity of directives.
func TestCompilerDirectiveCaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "lowercase define",
			input: "{$define DEBUG}",
		},
		{
			name:  "uppercase DEFINE",
			input: "{$DEFINE DEBUG}",
		},
		{
			name:  "mixed Define",
			input: "{$DeFiNe DEBUG}",
		},
		{
			name:  "lowercase ifdef",
			input: "{$define DEBUG}\n{$ifdef DEBUG}\nx := 1;\n{$endif}",
		},
		{
			name:  "uppercase IFDEF",
			input: "{$DEFINE DEBUG}\n{$IFDEF DEBUG}\nx := 1;\n{$ENDIF}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Consume all tokens - should not produce errors
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
			}

			errors := l.Errors()
			if len(errors) > 0 {
				t.Errorf("expected no errors, but got: %v", errors)
			}
		})
	}
}

// TestCompilerDirectiveConstTracking tests constant value tracking for $IF expressions.
// Note: Const tracking works by monitoring the token stream, so consts must appear
// BEFORE directives, and only simple "const NAME = INTEGER" patterns are tracked.
func TestCompilerDirectiveConstTracking(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "const before directive - not yet implemented",
			input: `const VERSION = 5;
x := 1;`,
			expected: []string{"VERSION", "x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			var identifiers []string
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
				if tok.Type == IDENT {
					identifiers = append(identifiers, tok.Literal)
				}
			}

			if len(identifiers) != len(tt.expected) {
				t.Fatalf("wrong number of identifiers. expected=%v, got=%v",
					tt.expected, identifiers)
			}

			for i, exp := range tt.expected {
				if identifiers[i] != exp {
					t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
						i, exp, identifiers[i])
				}
			}
		})
	}
}

// TestCompilerDirectiveMultiline tests multiline directives.
func TestCompilerDirectiveMultiline(t *testing.T) {
	input := `{$DEFINE
	DEBUG}
x := 1;`

	l := New(input)

	tok := l.NextToken()
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Errorf("expected identifier 'x', got type=%q literal=%q", tok.Type, tok.Literal)
	}
}

// TestCompilerDirectiveComplex tests complex directive combinations.
func TestCompilerDirectiveComplex(t *testing.T) {
	input := `{$DEFINE PLATFORM_WINDOWS}
{$DEFINE VERSION_5}

{$IF defined(PLATFORM_WINDOWS) and defined(VERSION_5)}
x := 1;
{$ENDIF}

{$IFDEF PLATFORM_WINDOWS}
	y := 2;
{$ELSE}
	z := 3;
{$ENDIF}`

	expected := []string{"x", "y"}

	l := New(input)

	var identifiers []string
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == IDENT {
			identifiers = append(identifiers, tok.Literal)
		}
	}

	if len(identifiers) != len(expected) {
		t.Fatalf("wrong number of identifiers. expected=%v, got=%v",
			expected, identifiers)
	}

	for i, exp := range expected {
		if identifiers[i] != exp {
			t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
				i, exp, identifiers[i])
		}
	}
}

// Helper function to check if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (hasPrefix(s, substr) || hasSuffix(s, substr) || containsMiddle(s, substr)))
}

func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
