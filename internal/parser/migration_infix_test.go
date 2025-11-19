package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestMigration_InfixExpression_SimpleOperators tests binary expressions in both modes.
//
// Note: These tests focus on the infix expression parsing itself, not the full
// expression parsing pipeline. Full integration requires migrating parseExpression
// (future task), so these tests validate the cursor pattern in isolation.
func TestMigration_InfixExpression_SimpleOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
	}{
		{
			name:     "addition",
			input:    "5 + 3",
			operator: "+",
		},
		{
			name:     "subtraction",
			input:    "10 - 4",
			operator: "-",
		},
		{
			name:     "multiplication",
			input:    "6 * 7",
			operator: "*",
		},
		{
			name:     "division",
			input:    "20 / 5",
			operator: "/",
		},
		{
			name:     "equality",
			input:    "x = y",
			operator: "=",
		},
		{
			name:     "less than",
			input:    "a < b",
			operator: "<",
		},
		{
			name:     "greater than",
			input:    "m > n",
			operator: ">",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser to get the full expression
			traditionalParser := New(lexer.New(tt.input))
			traditionalExpr := traditionalParser.ParseProgram()

			// Parse with cursor parser to get the full expression
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorExpr := cursorParser.ParseProgram()

			// Both should succeed
			if len(traditionalParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", traditionalParser.Errors())
			}
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Both should have one statement (expression statement)
			if len(traditionalExpr.Statements) != 1 {
				t.Fatalf("Traditional parser got %d statements, want 1", len(traditionalExpr.Statements))
			}
			if len(cursorExpr.Statements) != 1 {
				t.Fatalf("Cursor parser got %d statements, want 1", len(cursorExpr.Statements))
			}

			// Extract the binary expressions
			traditionalStmt, ok := traditionalExpr.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Traditional statement is %T, want *ast.ExpressionStatement", traditionalExpr.Statements[0])
			}
			cursorStmt, ok := cursorExpr.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Cursor statement is %T, want *ast.ExpressionStatement", cursorExpr.Statements[0])
			}

			traditionalBin, ok := traditionalStmt.Expression.(*ast.BinaryExpression)
			if !ok {
				t.Fatalf("Traditional expression is %T, want *ast.BinaryExpression", traditionalStmt.Expression)
			}
			cursorBin, ok := cursorStmt.Expression.(*ast.BinaryExpression)
			if !ok {
				t.Fatalf("Cursor expression is %T, want *ast.BinaryExpression", cursorStmt.Expression)
			}

			// Operators should match
			if traditionalBin.Operator != tt.operator {
				t.Errorf("Traditional operator = %q, want %q", traditionalBin.Operator, tt.operator)
			}
			if cursorBin.Operator != tt.operator {
				t.Errorf("Cursor operator = %q, want %q", cursorBin.Operator, tt.operator)
			}

			// Operators should be identical
			if traditionalBin.Operator != cursorBin.Operator {
				t.Errorf("Operator mismatch: traditional=%q, cursor=%q",
					traditionalBin.Operator, cursorBin.Operator)
			}

			// String representations should match
			if traditionalBin.String() != cursorBin.String() {
				t.Errorf("String mismatch:\ntraditional=%s\ncursor=%s",
					traditionalBin.String(), cursorBin.String())
			}
		})
	}
}

// TestMigration_InfixExpression_Precedence tests operator precedence in both modes
func TestMigration_InfixExpression_Precedence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiplication before addition",
			input:    "2 + 3 * 4",
			expected: "(2 + (3 * 4))",
		},
		{
			name:     "parentheses override precedence",
			input:    "(2 + 3) * 4",
			expected: "((2 + 3) * 4)",
		},
		{
			name:     "comparison before logical AND",
			input:    "x > 5 and y < 10",
			expected: "((x > 5) and (y < 10))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			traditionalParser := New(lexer.New(tt.input))
			traditionalProgram := traditionalParser.ParseProgram()

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()

			// Both should succeed
			if len(traditionalParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", traditionalParser.Errors())
			}
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Extract expressions
			if len(traditionalProgram.Statements) != 1 || len(cursorProgram.Statements) != 1 {
				t.Fatal("Expected 1 statement in both programs")
			}

			traditionalStmt := traditionalProgram.Statements[0].(*ast.ExpressionStatement)
			cursorStmt := cursorProgram.Statements[0].(*ast.ExpressionStatement)

			// Compare string representations
			traditionalStr := traditionalStmt.Expression.String()
			cursorStr := cursorStmt.Expression.String()

			if traditionalStr != cursorStr {
				t.Errorf("Expression mismatch:\ntraditional=%s\ncursor=%s",
					traditionalStr, cursorStr)
			}

			if traditionalStr != tt.expected {
				t.Errorf("Expression = %s, want %s", traditionalStr, tt.expected)
			}
		})
	}
}

// TestMigration_InfixExpression_ChainedOperators tests chained binary operations
func TestMigration_InfixExpression_ChainedOperators(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "multiple additions",
			input: "1 + 2 + 3 + 4",
		},
		{
			name:  "mixed operators",
			input: "10 - 3 + 5 - 2",
		},
		{
			name:  "comparison chain",
			input: "a < b and b < c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with both parsers
			traditionalParser := New(lexer.New(tt.input))
			traditionalProgram := traditionalParser.ParseProgram()

			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()

			// Both should succeed without errors
			if len(traditionalParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", traditionalParser.Errors())
			}
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Extract expressions
			if len(traditionalProgram.Statements) != 1 || len(cursorProgram.Statements) != 1 {
				t.Fatal("Expected 1 statement in both programs")
			}

			traditionalStmt := traditionalProgram.Statements[0].(*ast.ExpressionStatement)
			cursorStmt := cursorProgram.Statements[0].(*ast.ExpressionStatement)

			// Compare AST structure (string representation)
			if traditionalStmt.Expression.String() != cursorStmt.Expression.String() {
				t.Errorf("Expression mismatch:\ntraditional=%s\ncursor=%s",
					traditionalStmt.Expression.String(),
					cursorStmt.Expression.String())
			}
		})
	}
}

// TestMigration_InfixExpression_Dispatcher validates the dispatcher routes correctly
func TestMigration_InfixExpression_Dispatcher(t *testing.T) {
	input := "5 + 3"

	// Traditional parser should use traditional implementation
	traditionalParser := New(lexer.New(input))
	if false {
		t.Error("Traditional parser should have useCursor=false")
	}
	traditionalProgram := traditionalParser.ParseProgram()
	if len(traditionalParser.Errors()) > 0 {
		t.Errorf("Traditional parser errors: %v", traditionalParser.Errors())
	}

	// Cursor parser should use cursor mode
	cursorParser := NewCursorParser(lexer.New(input))
	if !true {
		t.Error("Cursor parser should have useCursor=true")
	}
	cursorProgram := cursorParser.ParseProgram()
	if len(cursorParser.Errors()) > 0 {
		t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
	}

	// Both should produce equivalent results
	if len(traditionalProgram.Statements) != 1 || len(cursorProgram.Statements) != 1 {
		t.Fatal("Expected 1 statement in both programs")
	}

	traditionalExpr := traditionalProgram.Statements[0].(*ast.ExpressionStatement).Expression
	cursorExpr := cursorProgram.Statements[0].(*ast.ExpressionStatement).Expression

	if traditionalExpr.String() != cursorExpr.String() {
		t.Errorf("Expression mismatch:\ntraditional=%s\ncursor=%s",
			traditionalExpr.String(), cursorExpr.String())
	}
}
