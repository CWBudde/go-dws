package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.7.1.2: Declaration migration tests
//
// This file tests the cursor migration of:
// - parseConstDeclaration (dispatcher)
// - parseVarDeclaration (dispatcher)
// - parseProgramDeclaration

// TestMigration_ConstDeclaration_Dispatcher tests const declaration dispatcher
func TestMigration_ConstDeclaration_Dispatcher(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single const", "const X = 42;"},
		{"typed const", "const Y: Integer = 100;"},
		{"multiple consts", "const A = 1; B = 2; C = 3;"},
		{"string const", "const Greeting = 'Hello';"},
		{"float const", "const PI: Float = 3.14;"},
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

// TestMigration_VarDeclaration_Dispatcher tests var declaration dispatcher
func TestMigration_VarDeclaration_Dispatcher(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single var", "var x: Integer;"},
		{"var with init", "var y: Integer := 10;"},
		{"multiple vars", "var a, b, c: Integer;"},
		{"multiple declarations", "var x: Integer; y: String;"},
		{"inferred type", "var z := 42;"},
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

// TestMigration_ProgramDeclaration tests program declaration migration
func TestMigration_ProgramDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple program", "program MyProgram; var x: Integer;"},
		{"program with begin", "program Test; begin end."},
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

// TestMigration_ConstDeclaration_Errors tests error handling in const declarations
func TestMigration_ConstDeclaration_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing value", "const X;"},
		{"missing equals", "const Y: Integer;"},
		{"invalid syntax", "const =;"},
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

// TestMigration_VarDeclaration_Errors tests error handling in var declarations
func TestMigration_VarDeclaration_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing type and init", "var x;"},
		{"invalid syntax", "var =;"},
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

// TestMigration_ProgramDeclaration_Errors tests error handling in program declarations
func TestMigration_ProgramDeclaration_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing name", "program ;"},
		{"missing semicolon", "program Test begin end."},
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

// TestMigration_MultipleDeclarations tests mixed declarations
func TestMigration_MultipleDeclarations(t *testing.T) {
	input := `
		const PI = 3.14;
		var radius: Float;
		const TAU = 6.28;
		var area: Float;
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

	// Should have 4 statements
	expectedStatements := 4
	if len(tradProgram.Statements) != expectedStatements {
		t.Errorf("Traditional: expected %d statements, got %d", expectedStatements, len(tradProgram.Statements))
	}
	if len(cursorProgram.Statements) != expectedStatements {
		t.Errorf("Cursor: expected %d statements, got %d", expectedStatements, len(cursorProgram.Statements))
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

	// Verify types of statements
	for i, stmt := range tradProgram.Statements {
		cursorStmt := cursorProgram.Statements[i]

		// Both should be same type
		switch stmt.(type) {
		case *ast.ConstDecl:
			if _, ok := cursorStmt.(*ast.ConstDecl); !ok {
				t.Errorf("Statement %d type mismatch: traditional=ConstDecl, cursor=%T", i, cursorStmt)
			}
		case *ast.VarDeclStatement:
			if _, ok := cursorStmt.(*ast.VarDeclStatement); !ok {
				t.Errorf("Statement %d type mismatch: traditional=VarDeclStatement, cursor=%T", i, cursorStmt)
			}
		}
	}
}
