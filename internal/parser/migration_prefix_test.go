package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.2.12: Test prefix expression handler migration to cursor mode
//
// This file tests the migration of prefix expression handlers to cursor mode:
// - parsePrefixExpressionCursor: -x, +x, not x
// - parseGroupedExpressionCursor: (expr), ()
// - parseArrayLiteralCursor: [1, 2, 3], [], [one..five]
// - parseNilLiteralCursor: nil
// - parseNullIdentifierCursor: Null
// - parseUnassignedIdentifierCursor: Unassigned
// - parseCharLiteralCursor: #65, #$41
//
// Each test parses the same source in both traditional and cursor modes,
// then compares the resulting ASTs using String() representation to verify
// semantic equivalence (ignoring minor Token position differences).

// TestParsePrefixExpression_Traditional_vs_Cursor tests unary prefix operators
func TestParsePrefixExpression_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"minus integer", "-5"},
		{"minus identifier", "-x"},
		{"minus complex", "-(x + y)"},
		{"plus integer", "+42"},
		{"plus identifier", "+count"},
		{"not boolean", "not true"},
		{"not identifier", "not flag"},
		{"not comparison", "not (x > y)"},
		{"nested prefix", "-(-x)"},        // Double negation with explicit grouping
		{"nested not", "not not enabled"}, // Double NOT
		{"complex nested", "-(x * -y)"},
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

			// Compare ASTs (semantic equivalence via String())
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Fatalf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			if len(tradProgram.Statements) == 0 {
				return // Empty program
			}

			// Get expressions from ExpressionStatements
			tradStmt, ok := tradProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Traditional: expected ExpressionStatement, got %T", tradProgram.Statements[0])
			}
			cursorStmt, ok := cursorProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Cursor: expected ExpressionStatement, got %T", cursorProgram.Statements[0])
			}

			tradExpr := tradStmt.Expression
			cursorExpr := cursorStmt.Expression

			// String representations should match (semantic equivalence)
			if tradExpr != nil && cursorExpr != nil {
				if tradExpr.String() != cursorExpr.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradExpr.String(), cursorExpr.String())
				}
			}
		})
	}
}

// TestParseGroupedExpression_Traditional_vs_Cursor tests grouped expressions
func TestParseGroupedExpression_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple grouped", "(42)"},
		{"grouped expression", "(x + y)"},
		{"nested grouped", "((x))"},
		{"grouped infix", "(a * b + c)"},
		{"complex nested", "((x + y) * (z - w))"},
		{"empty parens", "()"},                          // Empty array literal
		{"grouped with member", "(obj).Field"},          // Grouped then member access
		{"grouped call result", "(GetObj()).Method()"}, // Grouped call then method
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

			// Compare ASTs
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Fatalf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			if len(tradProgram.Statements) == 0 {
				return
			}

			// Get expressions from ExpressionStatements
			tradStmt, ok := tradProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Traditional: expected ExpressionStatement, got %T", tradProgram.Statements[0])
			}
			cursorStmt, ok := cursorProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Cursor: expected ExpressionStatement, got %T", cursorProgram.Statements[0])
			}

			tradExpr := tradStmt.Expression
			cursorExpr := cursorStmt.Expression

			if tradExpr != nil && cursorExpr != nil {
				if tradExpr.String() != cursorExpr.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradExpr.String(), cursorExpr.String())
				}
			}
		})
	}
}

// TestParseArrayLiteral_Traditional_vs_Cursor tests array literal expressions
func TestParseArrayLiteral_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty array", "[]"},
		{"single element", "[42]"},
		{"multiple elements", "[1, 2, 3]"},
		{"string array", "['a', 'b', 'c']"},
		{"identifier array", "[x, y, z]"},
		{"expression array", "[x + 1, y * 2, z - 3]"},
		{"nested array", "[[1, 2], [3, 4]]"},
		{"range literal", "[1..10]"},         // Range syntax
		{"range identifiers", "[one..five]"}, // Range with identifiers
		{"mixed range", "[1, 5..10, 20]"},    // Mix ranges and values
		{"trailing comma", "[1, 2, 3, ]"},    // Trailing comma allowed
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

			// Compare ASTs
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Fatalf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			if len(tradProgram.Statements) == 0 {
				return
			}

			// Get expressions from ExpressionStatements
			tradStmt, ok := tradProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Traditional: expected ExpressionStatement, got %T", tradProgram.Statements[0])
			}
			cursorStmt, ok := cursorProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Cursor: expected ExpressionStatement, got %T", cursorProgram.Statements[0])
			}

			tradExpr := tradStmt.Expression
			cursorExpr := cursorStmt.Expression

			if tradExpr != nil && cursorExpr != nil {
				if tradExpr.String() != cursorExpr.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradExpr.String(), cursorExpr.String())
				}
			}
		})
	}
}

// TestParseSimpleLiterals_Traditional_vs_Cursor tests simple literal expressions
func TestParseSimpleLiterals_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"nil literal", "nil"},
		{"Null identifier", "Null"},
		{"Unassigned identifier", "Unassigned"},
		{"char decimal", "#65"},      // 'A'
		{"char hex lowercase", "#$41"}, // 'A' in hex
		{"char hex uppercase", "#$5A"}, // 'Z' in hex
		{"char space", "#32"},        // Space character
		{"nil in expression", "x := nil"},
		{"Null in expression", "y := Null"},
		{"Unassigned in expression", "z := Unassigned"},
		{"char in array", "[#65, #66, #67]"},
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

			// Compare ASTs
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Fatalf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			if len(tradProgram.Statements) == 0 {
				return
			}

			// Compare full program strings
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("Program String mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestPrefixIntegration_Traditional_vs_Cursor tests combinations of prefix handlers
func TestPrefixIntegration_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"array of negations", "[-1, -2, -3]"},
		{"not with grouped", "not (x and y)"},
		{"prefix in array", "[not flag, -count, +value]"},
		{"grouped prefix", "(not enabled)"},
		{"prefix of array index", "-arr[0]"},
		{"prefix of member access", "-obj.Value"},
		{"prefix of call", "-GetCount()"},
		{"array range with prefix", "[-10..-1]"},
		{"complex integration", "[(x + y), not flag, -count, nil]"},
		{"nested arrays with prefix", "[[-1, -2], [+3, +4]]"},
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

			// Compare ASTs
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Fatalf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			if len(tradProgram.Statements) == 0 {
				return
			}

			// Get expressions from ExpressionStatements
			tradStmt, ok := tradProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Traditional: expected ExpressionStatement, got %T", tradProgram.Statements[0])
			}
			cursorStmt, ok := cursorProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Cursor: expected ExpressionStatement, got %T", cursorProgram.Statements[0])
			}

			tradExpr := tradStmt.Expression
			cursorExpr := cursorStmt.Expression

			if tradExpr != nil && cursorExpr != nil {
				if tradExpr.String() != cursorExpr.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradExpr.String(), cursorExpr.String())
				}
			}
		})
	}
}

// TestPrefixEdgeCases_Traditional_vs_Cursor tests edge cases and error handling
func TestPrefixEdgeCases_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectErrors  bool
	}{
		{"unclosed grouped", "(x + y", true},
		{"unclosed array", "[1, 2", true},
		{"invalid char format", "#", true},
		{"char hex no value", "#$", true},
		{"empty expression", "", false}, // Empty program, no errors
		{"whitespace only", "   ", false},
		{"comment only", "// comment", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.source))
			tradProgram := tradParser.ParseProgram()
			tradHasErrors := len(tradParser.Errors()) > 0

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()
			cursorHasErrors := len(cursorParser.Errors()) > 0

			// Both modes should agree on whether there are errors
			if tradHasErrors != cursorHasErrors {
				t.Errorf("Error state mismatch: traditional has errors=%v, cursor has errors=%v",
					tradHasErrors, cursorHasErrors)
				if tradHasErrors {
					t.Logf("Traditional errors: %v", tradParser.Errors())
				}
				if cursorHasErrors {
					t.Logf("Cursor errors: %v", cursorParser.Errors())
				}
			}

			// If we expect errors, make sure both modes have them
			if tt.expectErrors && (!tradHasErrors || !cursorHasErrors) {
				t.Errorf("Expected errors but got: traditional=%v, cursor=%v",
					tradHasErrors, cursorHasErrors)
			}

			// For valid programs, compare AST strings
			if !tt.expectErrors && !tradHasErrors && !cursorHasErrors {
				if tradProgram.String() != cursorProgram.String() {
					t.Errorf("Program String mismatch:\nTraditional: %s\nCursor: %s",
						tradProgram.String(), cursorProgram.String())
				}
			}
		})
	}
}
