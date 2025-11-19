package parser

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestDualMode_ParserCreation tests that both parser modes can be created
func TestDualMode_ParserCreation(t *testing.T) {
	source := "var x: Integer := 42;"

	// Test traditional parser creation
	traditionalParser := New(lexer.New(source))
	if traditionalParser == nil {
		t.Fatal("New() returned nil")
	}
	if traditionalParser.useCursor {
		t.Error("Traditional parser should have useCursor = false")
	}
	if traditionalParser.cursor != nil {
		t.Error("Traditional parser should have cursor = nil")
	}

	// Test cursor parser creation
	cursorParser := NewCursorParser(lexer.New(source))
	if cursorParser == nil {
		t.Fatal("NewCursorParser() returned nil")
	}
	if !cursorParser.useCursor {
		t.Error("Cursor parser should have useCursor = true")
	}
	if cursorParser.cursor == nil {
		t.Error("Cursor parser should have non-nil cursor")
	}
}

// TestDualMode_SimpleExpression tests that both modes parse simple expressions identically
func TestDualMode_SimpleExpression(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "integer literal",
			source: "42",
		},
		{
			name:   "string literal",
			source: `"hello"`,
		},
		{
			name:   "boolean true",
			source: "true",
		},
		{
			name:   "boolean false",
			source: "false",
		},
		{
			name:   "identifier",
			source: "myVar",
		},
		{
			name:   "binary expression",
			source: "3 + 5",
		},
		{
			name:   "complex expression",
			source: "(2 + 3) * 4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			traditionalParser := New(lexer.New(tt.source))
			traditionalExpr := traditionalParser.parseExpression(LOWEST)

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpression(LOWEST)

			// Both should succeed
			if traditionalExpr == nil {
				t.Error("Traditional parser returned nil expression")
			}
			if cursorExpr == nil {
				t.Error("Cursor parser returned nil expression")
			}

			// Both should have same structure (deep comparison)
			if !astEqual(traditionalExpr, cursorExpr) {
				t.Errorf("AST mismatch:\nTraditional: %#v\nCursor: %#v",
					traditionalExpr, cursorExpr)
			}

			// Both should have same string representation
			if traditionalExpr != nil && cursorExpr != nil {
				if traditionalExpr.String() != cursorExpr.String() {
					t.Errorf("String representation mismatch:\nTraditional: %s\nCursor: %s",
						traditionalExpr.String(), cursorExpr.String())
				}
			}
		})
	}
}

// TestDualMode_VarDeclaration tests that both modes parse variable declarations identically
func TestDualMode_VarDeclaration(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "typed var with initializer",
			source: "var x: Integer := 42;",
		},
		{
			name:   "typed var without initializer",
			source: "var x: Integer;",
		},
		{
			name:   "inferred var",
			source: "var x := 42;",
		},
		{
			name:   "string var",
			source: `var msg: String := "hello";`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			traditionalParser := New(lexer.New(tt.source))
			traditionalStmt := traditionalParser.parseStatement()

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorStmt := cursorParser.parseStatement()

			// Both should succeed
			if traditionalStmt == nil {
				t.Error("Traditional parser returned nil statement")
			}
			if cursorStmt == nil {
				t.Error("Cursor parser returned nil statement")
				if len(cursorParser.Errors()) > 0 {
					t.Logf("Cursor parser errors:")
					for _, err := range cursorParser.Errors() {
						t.Logf("  %v", err)
					}
				}
			}

			// Both should have same structure
			if !astEqual(traditionalStmt, cursorStmt) {
				t.Errorf("AST mismatch:\nTraditional: %#v\nCursor: %#v",
					traditionalStmt, cursorStmt)
			}
		})
	}
}

// TestDualMode_Program tests that both modes parse complete programs identically
func TestDualMode_Program(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple program",
			source: `
				var x: Integer := 42;
				var y: String := "test";
			`,
		},
		{
			name: "program with function",
			source: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;
			`,
		},
		{
			name: "program with if statement",
			source: `
				begin
					if x > 0 then
						y := 1
					else
						y := 0;
				end;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			traditionalParser := New(lexer.New(tt.source))
			traditionalProgram := traditionalParser.ParseProgram()

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()

			// Both should succeed
			if traditionalProgram == nil {
				t.Error("Traditional parser returned nil program")
			}
			if cursorProgram == nil {
				t.Error("Cursor parser returned nil program")
			}

			// Both should have same number of statements
			if traditionalProgram != nil && cursorProgram != nil {
				if len(traditionalProgram.Statements) != len(cursorProgram.Statements) {
					t.Errorf("Statement count mismatch: traditional=%d, cursor=%d",
						len(traditionalProgram.Statements), len(cursorProgram.Statements))
				}
			}

			// Both should have same errors (or no errors)
			traditionalErrors := len(traditionalParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if traditionalErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					traditionalErrors, cursorErrors)
				if cursorErrors > 0 {
					t.Logf("Cursor errors:")
					for _, err := range cursorParser.Errors() {
						t.Logf("  %v", err)
					}
				}
				if traditionalErrors > 0 {
					t.Logf("Traditional errors:")
					for _, err := range traditionalParser.Errors() {
						t.Logf("  %v", err)
					}
				}
			}
		})
	}
}

// TestDualMode_Errors tests that both modes produce identical errors
func TestDualMode_Errors(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "missing semicolon",
			source: "var x: Integer",
		},
		{
			name:   "invalid expression",
			source: "var x := ;",
		},
		{
			name:   "missing type",
			source: "var x;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			traditionalParser := New(lexer.New(tt.source))
			_ = traditionalParser.ParseProgram()
			traditionalErrors := traditionalParser.Errors()

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			_ = cursorParser.ParseProgram()
			cursorErrors := cursorParser.Errors()

			// Both should have errors
			if len(traditionalErrors) == 0 {
				t.Error("Traditional parser should have errors but has none")
			}
			if len(cursorErrors) == 0 {
				t.Error("Cursor parser should have errors but has none")
			}

			// Error counts should match
			if len(traditionalErrors) != len(cursorErrors) {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					len(traditionalErrors), len(cursorErrors))
			}
		})
	}
}

// TestDualMode_StateManagement tests that saveState/restoreState work in both modes
func TestDualMode_StateManagement(t *testing.T) {
	source := "var x: Integer := 42; var y: String := 'test';"

	t.Run("traditional mode", func(t *testing.T) {
		p := New(lexer.New(source))

		// Save initial state
		state := p.saveState()

		// Parse something
		_ = p.parseStatement()

		// Restore state
		p.restoreState(state)

		// Should be back at start
		if p.curToken.Type != lexer.VAR {
			t.Errorf("After restore, expected VAR token, got %v", p.curToken.Type)
		}
	})

	t.Run("cursor mode", func(t *testing.T) {
		p := NewCursorParser(lexer.New(source))

		// Save initial state
		state := p.saveState()

		// Verify cursor was saved
		if state.cursor == nil {
			t.Error("saveState() should save cursor in cursor mode")
		}

		// Parse something
		_ = p.parseStatement()

		// Restore state
		p.restoreState(state)

		// Should be back at start
		if p.cursor == nil {
			t.Error("cursor should not be nil after restore")
		}
		if p.cursor.Current().Type != lexer.VAR {
			t.Errorf("After restore, expected VAR token, got %v", p.cursor.Current().Type)
		}
		// curToken should also be synced
		if p.curToken.Type != lexer.VAR {
			t.Errorf("After restore, curToken should be synced to VAR, got %v", p.curToken.Type)
		}
	})
}

// TestDualMode_CursorTokenSync tests that cursor mode syncs curToken/peekToken
func TestDualMode_CursorTokenSync(t *testing.T) {
	source := "var x: Integer := 42;"

	p := NewCursorParser(lexer.New(source))

	// In cursor mode, curToken and peekToken should be synced with cursor
	if p.cursor.Current().Type != p.curToken.Type {
		t.Errorf("curToken not synced with cursor: cursor=%v, curToken=%v",
			p.cursor.Current().Type, p.curToken.Type)
	}

	if p.cursor.Peek(1).Type != p.peekToken.Type {
		t.Errorf("peekToken not synced with cursor: cursor.Peek(1)=%v, peekToken=%v",
			p.cursor.Peek(1).Type, p.peekToken.Type)
	}

	// After calling syncCursorToTokens(), they should still be in sync
	p.syncCursorToTokens()

	if p.cursor.Current().Type != p.curToken.Type {
		t.Error("curToken not synced after explicit syncCursorToTokens()")
	}

	if p.cursor.Peek(1).Type != p.peekToken.Type {
		t.Error("peekToken not synced after explicit syncCursorToTokens()")
	}
}

// TestDualMode_ModeFlag tests that the useCursor flag correctly identifies parser mode
func TestDualMode_ModeFlag(t *testing.T) {
	source := "var x := 42;"

	traditional := New(lexer.New(source))
	if traditional.useCursor {
		t.Error("Traditional parser should have useCursor=false")
	}

	cursor := NewCursorParser(lexer.New(source))
	if !cursor.useCursor {
		t.Error("Cursor parser should have useCursor=true")
	}
}

// TestDualMode_TypeDeclarations tests that both modes parse type declarations identically
// This test covers Bug #1 (contextual keywords as type names) and Bug #2 (non-alias type declarations)
func TestDualMode_TypeDeclarations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "type alias",
			source: "type TUserID = Integer;",
		},
		{
			name:   "class declaration",
			source: "type TMyClass = class end;",
		},
		{
			name:   "class with fields",
			source: "type TPoint = class X, Y: Integer; end;",
		},
		{
			name:   "record declaration",
			source: "type TPoint = record X, Y: Integer; end;",
		},
		{
			name:   "interface declaration",
			source: "type ITest = interface end;",
		},
		{
			name:   "set declaration",
			source: "type TDays = set of Integer;",
		},
		{
			name:   "array declaration",
			source: "type TArr = array[1..5] of Integer;",
		},
		{
			name:   "dynamic array",
			source: "type TDynArray = array of String;",
		},
		{
			name:   "contextual keyword as type name (STEP)",
			source: "type Step = Integer;",
		},
		{
			name:   "contextual keyword as type name (SELF)",
			source: "type Self = String;",
		},
		{
			name:   "multiple type declarations",
			source: "type TInt = Integer; type TStr = String;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			traditionalParser := New(lexer.New(tt.source))
			traditionalProgram := traditionalParser.ParseProgram()

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()

			// Check for errors
			traditionalErrors := traditionalParser.Errors()
			cursorErrors := cursorParser.Errors()

			// Both should succeed (no errors)
			if len(traditionalErrors) > 0 {
				t.Errorf("Traditional parser has errors:")
				for _, err := range traditionalErrors {
					t.Logf("  %v", err)
				}
			}
			if len(cursorErrors) > 0 {
				t.Errorf("Cursor parser has errors:")
				for _, err := range cursorErrors {
					t.Logf("  %v", err)
				}
			}

			// Both should produce non-nil programs
			if traditionalProgram == nil {
				t.Error("Traditional parser returned nil program")
			}
			if cursorProgram == nil {
				t.Error("Cursor parser returned nil program")
			}

			// Both should have same number of statements
			if traditionalProgram != nil && cursorProgram != nil {
				if len(traditionalProgram.Statements) != len(cursorProgram.Statements) {
					t.Errorf("Statement count mismatch: traditional=%d, cursor=%d",
						len(traditionalProgram.Statements), len(cursorProgram.Statements))
				}

				// Both should have same structure
				if !astEqual(traditionalProgram, cursorProgram) {
					t.Errorf("AST mismatch:\nTraditional: %#v\nCursor: %#v",
						traditionalProgram, cursorProgram)
				}
			}
		})
	}
}

// astEqual performs a deep comparison of two AST nodes.
// This is a simplified version - in practice you might use reflect.DeepEqual
// or a custom comparison that handles position info appropriately.
func astEqual(a, b ast.Node) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// For now, use reflect.DeepEqual
	// In the future, we might want custom comparison logic
	// that ignores position differences or other metadata
	return reflect.DeepEqual(a, b)
}
