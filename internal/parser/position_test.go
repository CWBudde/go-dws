package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestPositionTracking(t *testing.T) {
	input := `var x: Integer := 42;
var y := 'hello';
begin
  x := x + 1;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if program == nil {
		t.Fatal("Failed to parse program")
	}

	// Check that program has position information
	if program.Pos().Line == 0 {
		t.Error("Program Pos() not set")
	}

	if program.End().Line == 0 {
		t.Error("Program End() not set (EndPos not populated)")
	}

	t.Logf("Program: Pos=%v End=%v", program.Pos(), program.End())

	// Check each statement has position information
	for i, stmt := range program.Statements {
		if stmt.Pos().Line == 0 {
			t.Errorf("Statement %d Pos() not set", i)
		}

		if stmt.End().Line == 0 {
			t.Errorf("Statement %d End() not set (EndPos not populated)", i)
		}

		t.Logf("Statement %d: Pos=%v End=%v Type=%T", i, stmt.Pos(), stmt.End(), stmt)
	}
}

// TestLiteralPositions tests that literal nodes have proper position tracking
func TestLiteralPositions(t *testing.T) {
	tests := []struct {
		input    string
		wantLine int
		wantCol  int
	}{
		{"42", 1, 1},
		{"3.14", 1, 1},
		{"'hello'", 1, 1},
		{"true", 1, 1},
		{"false", 1, 1},
		{"nil", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			expr := p.parseExpressionCursor(LOWEST)

			if expr == nil {
				t.Fatal("Failed to parse expression")
			}

			pos := expr.Pos()
			if pos.Line != tt.wantLine || pos.Column != tt.wantCol {
				t.Errorf("Pos() = (%d,%d), want (%d,%d)",
					pos.Line, pos.Column, tt.wantLine, tt.wantCol)
			}

			end := expr.End()
			if end.Line == 0 {
				t.Error("End() not set (EndPos not populated)")
			}

			t.Logf("%s: Pos=%v End=%v", tt.input, pos, end)
		})
	}
}

// TestBinaryExpressionPositions tests that binary expressions have proper position tracking
func TestBinaryExpressionPositions(t *testing.T) {
	input := "1 + 2"

	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpressionCursor(LOWEST)

	if expr == nil {
		t.Fatal("Failed to parse expression")
	}

	// Binary expression's Pos() returns the position of its Token (the operator)
	// which is at column 3 in "1 + 2"
	pos := expr.Pos()
	if pos.Line != 1 {
		t.Errorf("Pos().Line = %d, want 1", pos.Line)
	}

	// The End() should be set and point after the last token
	end := expr.End()
	if end.Line == 0 {
		t.Error("End() not set (EndPos not populated)")
	}

	// For "1 + 2", the end should be after '2' (column 6)
	if end.Column < 5 {
		t.Errorf("End().Column = %d, expected at least 5", end.Column)
	}

	t.Logf("Binary expression '1 + 2': Pos=%v End=%v", pos, end)
}

// TestArrayLiteralEndPos verifies that EndPos is correctly set for array literals
func TestArrayLiteralEndPos(t *testing.T) {
	input := "[1, 2, 3]"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	exprStmt := program.Statements[0]
	if exprStmt.End().Column == 0 {
		t.Errorf("EndPos not set: %v", exprStmt.End())
	}

	// The closing ']' is at column 9, EndPos points after it at column 10
	expectedCol := 10
	if exprStmt.End().Column != expectedCol {
		t.Errorf("expected EndPos.Column = %d, got %d", expectedCol, exprStmt.End().Column)
	}

	t.Logf("Array literal '%s': End=%v", input, exprStmt.End())
}

// TestSetLiteralEndPos verifies that EndPos is correctly set for set literals
func TestSetLiteralEndPos(t *testing.T) {
	input := "[Red, Blue]"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	exprStmt := program.Statements[0]
	if exprStmt.End().Column == 0 {
		t.Errorf("EndPos not set: %v", exprStmt.End())
	}

	// The closing ']' is at column 11, EndPos points after it at column 12
	expectedCol := 12
	if exprStmt.End().Column != expectedCol {
		t.Errorf("expected EndPos.Column = %d, got %d", expectedCol, exprStmt.End().Column)
	}

	t.Logf("Set literal '%s': End=%v", input, exprStmt.End())
}

// TestRecordLiteralEndPos verifies that EndPos is correctly set for record literals
func TestRecordLiteralEndPos(t *testing.T) {
	input := "p := (x: 10; y: 20);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	// The closing ')' is at position 19 (0-indexed column 18, after ')' is 19)
	assignStmt := program.Statements[0]
	if assignStmt.End().Column == 0 {
		t.Errorf("EndPos not set: %v", assignStmt.End())
	}

	t.Logf("Record literal assignment '%s': End=%v", input, assignStmt.End())
}

// TestRangeExpressionEndPos verifies that EndPos is correctly set for range expressions
func TestRangeExpressionEndPos(t *testing.T) {
	input := "[Red..Blue]"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	exprStmt := program.Statements[0]
	if exprStmt.End().Column == 0 {
		t.Errorf("EndPos not set: %v", exprStmt.End())
	}

	// The closing ']' is at column 11, EndPos points after it at column 12
	expectedCol := 12
	if exprStmt.End().Column != expectedCol {
		t.Errorf("expected EndPos.Column = %d, got %d", expectedCol, exprStmt.End().Column)
	}

	t.Logf("Range expression '%s': End=%v", input, exprStmt.End())
}
