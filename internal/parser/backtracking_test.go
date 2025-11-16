package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestParserStatePreservation tests that saveState/restoreState correctly
// preserves all parser state including tokens, errors, and lexer position.
func TestParserStatePreservation(t *testing.T) {
	input := "var x: Integer := 42;"
	l := lexer.New(input)
	p := New(l)
	p.nextToken() // Initialize parser
	p.nextToken()

	// Record initial state
	initialCur := p.curToken
	initialPeek := p.peekToken
	initialErrorCount := len(p.errors)

	// Save state
	state := p.saveState()

	// Modify parser state
	p.nextToken()
	p.nextToken()
	p.addError("test error", ErrInvalidExpression)

	// Verify state changed
	if p.curToken.Type == initialCur.Type {
		t.Error("curToken should have changed")
	}
	if len(p.errors) == initialErrorCount {
		t.Error("errors should have been added")
	}

	// Restore state
	p.restoreState(state)

	// Verify restoration
	if p.curToken.Type != initialCur.Type {
		t.Errorf("curToken not restored: got %v, want %v", p.curToken.Type, initialCur.Type)
	}
	if p.peekToken.Type != initialPeek.Type {
		t.Errorf("peekToken not restored: got %v, want %v", p.peekToken.Type, initialPeek.Type)
	}
	if len(p.errors) != initialErrorCount {
		t.Errorf("errors not restored: got %d errors, want %d", len(p.errors), initialErrorCount)
	}
}

// TestNestedSaveRestore tests that nested save/restore operations work correctly.
func TestNestedSaveRestore(t *testing.T) {
	input := "var x, y, z: Integer;"
	l := lexer.New(input)
	p := New(l)
	p.nextToken()
	p.nextToken()

	// Save first state
	state1 := p.saveState()
	tok1 := p.curToken

	// Advance
	p.nextToken()
	p.nextToken()

	// Save second state
	state2 := p.saveState()
	tok2 := p.curToken

	// Advance more
	p.nextToken()
	p.nextToken()

	// Verify we're at third position
	if p.curToken.Type == tok2.Type {
		t.Error("should have advanced past second state")
	}

	// Restore to second state
	p.restoreState(state2)
	if p.curToken.Type != tok2.Type {
		t.Errorf("failed to restore to state2: got %v, want %v", p.curToken.Type, tok2.Type)
	}

	// Restore to first state
	p.restoreState(state1)
	if p.curToken.Type != tok1.Type {
		t.Errorf("failed to restore to state1: got %v, want %v", p.curToken.Type, tok1.Type)
	}

	// Verify we can still advance normally
	p.nextToken()
	if p.curToken.Type == tok1.Type {
		t.Error("nextToken should advance after restore")
	}
}

// TestPeekAfterRestore verifies that Peek() works correctly after state restoration.
func TestPeekAfterRestore(t *testing.T) {
	input := "var x: Integer;"
	l := lexer.New(input)
	p := New(l)
	p.nextToken()
	p.nextToken()

	// Save state
	state := p.saveState()

	// Advance and modify
	p.nextToken()
	p.nextToken()
	p.addError("test error", ErrInvalidExpression)

	// Restore
	p.restoreState(state)

	// Verify peek still works
	peekType := p.peekToken.Type
	p.nextToken()
	if p.curToken.Type != peekType {
		t.Errorf("peek token mismatch after restore: peeked %v, got %v", peekType, p.curToken.Type)
	}
}

// TestIsExpressionBacktracking verifies that parseIsExpression correctly
// backtracks when type parsing fails and successfully parses as boolean expression.
func TestIsExpressionBacktracking(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string // "type" or "boolean"
	}{
		{
			name:     "is with boolean True",
			input:    "if x is True then PrintLn('yes');",
			wantType: "boolean",
		},
		{
			name:     "is with boolean False",
			input:    "if x is False then PrintLn('no');",
			wantType: "boolean",
		},
		{
			name:     "is with type name",
			input:    "if x is Integer then PrintLn('int');",
			wantType: "type",
		},
		{
			name:     "is with class type",
			input:    "if obj is TMyClass then PrintLn('class');",
			wantType: "type",
		},
		{
			name:     "is with comparison expression",
			input:    "if x is (y > 0) then PrintLn('pos');",
			wantType: "boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)

			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			ifStmt, ok := program.Statements[0].(*ast.IfStatement)
			if !ok {
				t.Fatalf("expected IfStatement, got %T", program.Statements[0])
			}

			isExpr, ok := ifStmt.Condition.(*ast.IsExpression)
			if !ok {
				t.Fatalf("expected IsExpression in condition, got %T", ifStmt.Condition)
			}

			switch tt.wantType {
			case "type":
				if isExpr.TargetType == nil {
					t.Error("expected TargetType to be set for type check")
				}
				if isExpr.Right != nil {
					t.Error("expected Right to be nil for type check")
				}
			case "boolean":
				if isExpr.Right == nil {
					t.Error("expected Right to be set for boolean comparison")
				}
				if isExpr.TargetType != nil {
					t.Error("expected TargetType to be nil for boolean comparison")
				}
			}
		})
	}
}

// TestBacktrackingPreservesNoErrors verifies that backtracking doesn't leave
// spurious errors when the second parse attempt succeeds.
func TestBacktrackingPreservesNoErrors(t *testing.T) {
	// This input will fail to parse 'True' as a type but succeed as boolean expression
	input := "if x is True then PrintLn('yes');"
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	// Should have NO errors - backtracking should have removed them
	if len(p.errors) != 0 {
		t.Errorf("expected 0 errors after successful backtracking, got %d:", len(p.errors))
		for _, err := range p.errors {
			t.Logf("  - %s", err)
		}
	}

	if program == nil {
		t.Fatal("program should not be nil")
	}
}

// TestStateIndependence verifies that saved states are independent copies
// and modifying one doesn't affect the other.
func TestStateIndependence(t *testing.T) {
	input := "var x: Integer;"
	l := lexer.New(input)
	p := New(l)
	p.nextToken()
	p.nextToken()

	// Save two states
	state1 := p.saveState()
	state2 := p.saveState()

	// Modify parser
	p.addError("error1", ErrInvalidExpression)
	p.addError("error2", ErrInvalidExpression)

	// Restore first state
	p.restoreState(state1)
	if len(p.errors) != 0 {
		t.Errorf("state1 should have 0 errors, got %d", len(p.errors))
	}

	// Restore second state (should be same as first)
	p.restoreState(state2)
	if len(p.errors) != 0 {
		t.Errorf("state2 should have 0 errors, got %d", len(p.errors))
	}
}
