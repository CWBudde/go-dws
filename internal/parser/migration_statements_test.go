package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.2.14: Statement migration test infrastructure
//
// This file provides the test framework for migrating statement parsing to cursor mode.
// Tests will be added incrementally as each statement type is migrated.

// TestStatementInfrastructure_Basic tests that parseStatementCursor exists and can be called
func TestStatementInfrastructure_Basic(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple expression", "42"},
		{"simple assignment", "x := 5"},
		{"if statement", "if x > 0 then y := 1"},
		{"while statement", "while x > 0 do x := x - 1"},
		{"begin end block", "begin x := 1; y := 2 end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.source))
			tradProgram := tradParser.ParseProgram()

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()

			// Both should produce programs
			if tradProgram == nil {
				t.Error("Traditional parser returned nil program")
			}
			if cursorProgram == nil {
				t.Error("Cursor parser returned nil program")
			}

			// Error counts should match
			tradErrors := len(tradParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
				if tradErrors > 0 {
					t.Logf("Traditional errors: %v", tradParser.Errors())
				}
				if cursorErrors > 0 {
					t.Logf("Cursor errors: %v", cursorParser.Errors())
				}
			}

			// Statement counts should match
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Errorf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			// Program strings should match (semantic equivalence)
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("Program String mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestStatementInfrastructure_EmptyProgram tests parsing empty programs
func TestStatementInfrastructure_EmptyProgram(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"comments only", "// comment\n// another comment"},
		{"semicolons only", ";;;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.source))
			tradProgram := tradParser.ParseProgram()
			checkParserErrors(t, tradParser)

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()
			checkParserErrors(t, cursorParser)

			// Both should produce empty or nearly empty programs
			if tradProgram == nil {
				t.Error("Traditional parser returned nil program")
			}
			if cursorProgram == nil {
				t.Error("Cursor parser returned nil program")
			}

			// Statement counts should match
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Errorf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}
		})
	}
}
