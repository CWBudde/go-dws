package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.7.1.5: Array migration tests
//
// This file tests the cursor migration of:
// - parseArrayDeclaration

// TestMigration_ArrayDeclaration tests array type declaration parsing
func TestMigration_ArrayDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"static array", "type TInts = array[1..10] of Integer;"},
		{"dynamic array", "type TDynamic = array of String;"},
		{"zero-based array", "type TZeroBased = array[0..9] of Float;"},
		{"char array", "type TChars = array[1..100] of Char;"},
		{"2D array", "type TMatrix = array[0..1, 0..2] of Integer;"},
		{"3D array", "type TCube = array[1..3, 1..4, 1..5] of Float;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			tradProgram := tradParser.ParseProgram()
			if len(tradParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", tradParser.Errors())
			}

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Both should produce programs
			if tradProgram == nil || cursorProgram == nil {
				t.Fatal("Parser returned nil program")
			}

			// Should have same number of statements
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Errorf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			// Error counts should match
			tradErrors := len(tradParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
			}

			// AST strings should match
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestMigration_ArrayDeclaration_Errors tests error handling
func TestMigration_ArrayDeclaration_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing of", "type TArr = array[1..10];"},
		{"missing element type", "type TArr = array[1..10] of;"},
		{"missing bounds", "type TArr = array[] of Integer;"},
		{"missing semicolon", "type TArr = array[1..10] of Integer"},
		{"invalid bounds", "type TArr = array[..10] of Integer;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			_ = tradParser.ParseProgram()
			tradErrors := len(tradParser.Errors())

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			_ = cursorParser.ParseProgram()
			cursorErrors := len(cursorParser.Errors())

			// Both should have errors
			if tradErrors == 0 {
				t.Error("Traditional parser should have errors")
			}
			if cursorErrors == 0 {
				t.Error("Cursor parser should have errors")
			}

			// Error counts should be similar (logged but not enforced)
			if tradErrors != cursorErrors {
				t.Logf("Error count difference: traditional=%d, cursor=%d", tradErrors, cursorErrors)
			}
		})
	}
}

// TestMigration_ArrayIndexExpression tests array indexing
func TestMigration_ArrayIndexExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple index", "var x := arr[0];"},
		{"expression index", "var x := arr[i + 1];"},
		{"nested brackets", "var x := matrix[i][j];"},
		{"comma syntax 2D", "var x := matrix[i, j];"},
		{"comma syntax 3D", "var x := cube[i, j, k];"},
		{"mixed syntax", "var x := arr[i][j, k];"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			tradProgram := tradParser.ParseProgram()
			if len(tradParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", tradParser.Errors())
			}

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Both should produce programs
			if tradProgram == nil || cursorProgram == nil {
				t.Fatal("Parser returned nil program")
			}

			// Error counts should match
			tradErrors := len(tradParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
			}

			// AST strings should match
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestMigration_ArrayLiteral tests array literal parsing
func TestMigration_ArrayLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty array", "var x := [];"},
		{"single element", "var x := [1];"},
		{"multiple elements", "var x := [1, 2, 3];"},
		{"expression elements", "var x := [a + 1, b * 2, c - 3];"},
		{"nested array", "var x := [[1, 2], [3, 4]];"},
		{"trailing comma", "var x := [1, 2, 3,];"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			tradProgram := tradParser.ParseProgram()
			if len(tradParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", tradParser.Errors())
			}

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Both should produce programs
			if tradProgram == nil || cursorProgram == nil {
				t.Fatal("Parser returned nil program")
			}

			// Error counts should match
			tradErrors := len(tradParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
			}

			// AST strings should match
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestMigration_ArrayDeclaration_Complex tests complex array declarations
func TestMigration_ArrayDeclaration_Complex(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"array of arrays",
			"type TMatrix = array[1..10] of array[1..20] of Integer;",
		},
		{
			"dynamic array of static",
			"type TDynStatic = array of array[1..5] of String;",
		},
		{
			"4D array",
			"type T4D = array[0..1, 0..2, 0..3, 0..4] of Float;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			tradProgram := tradParser.ParseProgram()
			if len(tradParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", tradParser.Errors())
			}

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Both should produce programs
			if tradProgram == nil || cursorProgram == nil {
				t.Fatal("Parser returned nil program")
			}

			// Error counts should match
			tradErrors := len(tradParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
			}

			// AST strings should match
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}
