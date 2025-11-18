package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.7.1.1: Unit, uses clause, and type declaration migration tests
//
// This file tests the cursor migration of:
// - parseUsesClause
// - parseUnit
// - parseTypeDeclaration

// TestMigration_UsesClause_Basic tests basic uses clause parsing
func TestMigration_UsesClause_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // number of units
	}{
		{"single unit", "uses Unit1;", 1},
		{"two units", "uses Unit1, Unit2;", 2},
		{"three units", "uses Unit1, Unit2, Unit3;", 3},
		{"many units", "uses A, B, C, D, E;", 5},
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

			// Both should have one statement (uses clause)
			if len(tradProgram.Statements) != 1 || len(cursorProgram.Statements) != 1 {
				t.Errorf("Expected 1 statement, got traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			// Both should be uses clauses
			tradUses, ok1 := tradProgram.Statements[0].(*ast.UsesClause)
			cursorUses, ok2 := cursorProgram.Statements[0].(*ast.UsesClause)

			if !ok1 || !ok2 {
				t.Error("Statement is not a UsesClause")
			}

			// Should have expected number of units
			if len(tradUses.Units) != tt.expected {
				t.Errorf("Traditional: expected %d units, got %d", tt.expected, len(tradUses.Units))
			}
			if len(cursorUses.Units) != tt.expected {
				t.Errorf("Cursor: expected %d units, got %d", tt.expected, len(cursorUses.Units))
			}

			// Unit names should match
			for i := 0; i < tt.expected; i++ {
				if tradUses.Units[i].Value != cursorUses.Units[i].Value {
					t.Errorf("Unit %d name mismatch: traditional=%s, cursor=%s",
						i, tradUses.Units[i].Value, cursorUses.Units[i].Value)
				}
			}

			// AST strings should match
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestMigration_UsesClause_Errors tests error handling in uses clause parsing
func TestMigration_UsesClause_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing unit name", "uses ;"},
		{"missing semicolon", "uses Unit1"},
		{"trailing comma", "uses Unit1, Unit2,"},
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

			// Error counts should match (or be similar)
			if tradErrors != cursorErrors {
				t.Logf("Error count difference: traditional=%d, cursor=%d", tradErrors, cursorErrors)
				// This is logged but not failed as error details may differ
			}
		})
	}
}

// TestMigration_TypeDeclaration_Basic tests basic type declaration parsing
func TestMigration_TypeDeclaration_Basic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"type alias", "type TMyInt = Integer;"},
		{"enum", "type TColor = (Red, Green, Blue);"},
		{"subrange", "type TDigit = 0..9;"},
		{"class", "type TMyClass = class end;"},
		{"record", "type TPoint = record X, Y: Integer; end;"},
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

// TestMigration_TypeDeclaration_Multiple tests multiple type declarations
func TestMigration_TypeDeclaration_Multiple(t *testing.T) {
	input := `type
		TFirst = Integer;
		TSecond = String;
		TThird = Boolean;
	`

	// Traditional mode
	tradParser := New(lexer.New(input))
	tradProgram := tradParser.ParseProgram()
	if len(tradParser.Errors()) > 0 {
		t.Errorf("Traditional parser errors: %v", tradParser.Errors())
	}

	// Cursor mode
	cursorParser := NewCursorParser(lexer.New(input))
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
}

// TestMigration_Unit_Basic tests basic unit declaration parsing
func TestMigration_Unit_Basic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"minimal unit",
			`unit MyUnit;
			interface
			implementation
			end.`,
		},
		{
			"unit with uses",
			`unit MyUnit;
			interface
			  uses System;
			implementation
			end.`,
		},
		{
			"unit with initialization",
			`unit MyUnit;
			interface
			implementation
			initialization
			  // init code
			end.`,
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

			// Should have one statement (unit declaration)
			if len(tradProgram.Statements) != 1 || len(cursorProgram.Statements) != 1 {
				t.Errorf("Expected 1 statement, got traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			// Both should be unit declarations
			_, ok1 := tradProgram.Statements[0].(*ast.UnitDeclaration)
			_, ok2 := cursorProgram.Statements[0].(*ast.UnitDeclaration)

			if !ok1 || !ok2 {
				t.Error("Statement is not a UnitDeclaration")
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

// TestMigration_Unit_Errors tests error handling in unit declaration parsing
func TestMigration_Unit_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing unit name", "unit ;"},
		{"missing semicolon after name", "unit MyUnit interface"},
		{"missing end", "unit MyUnit; interface implementation"},
		{"missing dot", "unit MyUnit; interface implementation end"},
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
