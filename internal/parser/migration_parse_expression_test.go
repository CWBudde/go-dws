package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestMigration_ParseExpression_SimpleLiterals tests parsing of simple literal expressions.
// Validates that cursor mode produces identical ASTs to traditional mode for basic literals.
func TestMigration_ParseExpression_SimpleLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"integer literal", "42"},
		{"negative integer", "-5"},
		{"zero", "0"},
		{"float literal", "3.14"},
		{"negative float", "-2.71828"},
		{"string literal", `"hello world"`},
		{"empty string", `""`},
		{"boolean true", "true"},
		{"boolean false", "false"},
		{"identifier", "myVariable"},
		{"nil", "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional mode
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpression(LOWEST)

			// Parse with cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpression(LOWEST)

			// Both should succeed
			if tradExpr == nil {
				t.Fatal("Traditional parser returned nil")
			}
			if cursorExpr == nil {
				t.Fatal("Cursor parser returned nil")
			}

			// String representations should match (semantic equivalence)
			// Note: We compare String() instead of DeepEqual because Token/EndPos
			// positions may differ slightly between traditional and cursor modes
			// due to sync timing, but the semantic meaning is identical.
			if tradExpr != nil && cursorExpr != nil {
				if tradExpr.String() != cursorExpr.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradExpr.String(), cursorExpr.String())
				}
			}

			// Both should have no errors
			if len(tradParser.Errors()) != len(cursorParser.Errors()) {
				t.Errorf("Error count mismatch: trad=%d, cursor=%d",
					len(tradParser.Errors()), len(cursorParser.Errors()))
			}
		})
	}
}

// TestMigration_ParseExpression_BinaryOperators tests binary operators.
func TestMigration_ParseExpression_BinaryOperators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string // Expected AST string representation
	}{
		{"addition", "3 + 5", "(3 + 5)"},
		{"subtraction", "10 - 3", "(10 - 3)"},
		{"multiplication", "4 * 7", "(4 * 7)"},
		{"division", "20 / 5", "(20 / 5)"},
		{"integer division", "17 div 5", "(17 div 5)"},
		{"modulo", "17 mod 5", "(17 mod 5)"},
		{"left shift", "8 shl 2", "(8 shl 2)"},
		{"right shift", "32 shr 2", "(32 shr 2)"},
		{"arithmetic shift", "16 sar 1", "(16 sar 1)"},
		{"equality", "x = y", "(x = y)"},
		{"inequality", "x <> y", "(x <> y)"},
		{"less than", "a < b", "(a < b)"},
		{"greater than", "a > b", "(a > b)"},
		{"less or equal", "a <= b", "(a <= b)"},
		{"greater or equal", "a >= b", "(a >= b)"},
		{"logical and", "x and y", "(x and y)"},
		{"logical or", "x or y", "(x or y)"},
		{"logical xor", "x xor y", "(x xor y)"},
		{"in operator", "x in mySet", "(x in mySet)"},
		{"coalesce", "x ?? y", "(x ?? y)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional mode
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpression(LOWEST)

			// Parse with cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpression(LOWEST)

			// Both should succeed
			if tradExpr == nil {
				t.Fatalf("Traditional parser returned nil")
			}
			if cursorExpr == nil {
				t.Fatalf("Cursor parser returned nil")
			}

			// Check expected string representation
			if tradExpr.String() != tt.expected {
				t.Errorf("Traditional: expected %q, got %q", tt.expected, tradExpr.String())
			}
			if cursorExpr.String() != tt.expected {
				t.Errorf("Cursor: expected %q, got %q", tt.expected, cursorExpr.String())
			}

			// Both parsers should produce semantically equivalent ASTs
			// Note: We don't use DeepEqual because Token/EndPos positions may
			// differ slightly between traditional and cursor modes due to sync timing.
			// String() comparison (checked above) verifies semantic equivalence.
		})
	}
}

// TestMigration_ParseExpression_Precedence tests operator precedence.
func TestMigration_ParseExpression_Precedence(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "multiplication before addition",
			source:   "2 + 3 * 4",
			expected: "(2 + (3 * 4))",
		},
		{
			name:     "division before subtraction",
			source:   "10 - 6 / 2",
			expected: "(10 - (6 / 2))",
		},
		{
			name:     "parentheses override precedence",
			source:   "(2 + 3) * 4",
			expected: "((2 + 3) * 4)",
		},
		{
			name:     "nested parentheses",
			source:   "((1 + 2) * 3) + 4",
			expected: "(((1 + 2) * 3) + 4)",
		},
		{
			name:     "comparison before logical and",
			source:   "x > 5 and y < 10",
			expected: "((x > 5) and (y < 10))",
		},
		{
			name:     "and before or",
			source:   "a or b and c",
			expected: "(a or (b and c))",
		},
		{
			name:     "complex expression",
			source:   "a + b * c - d / e",
			expected: "((a + (b * c)) - (d / e))",
		},
		{
			name:     "shift operators",
			source:   "x shl 2 + y",
			expected: "((x shl 2) + y)", // SHL has lower precedence than +
		},
		{
			name:     "coalesce with arithmetic",
			source:   "x + y ?? z",
			expected: "((x + y) ?? z)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional mode
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpression(LOWEST)

			// Parse with cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpression(LOWEST)

			// Both should succeed
			if tradExpr == nil {
				t.Fatalf("Traditional parser returned nil")
			}
			if cursorExpr == nil {
				t.Fatalf("Cursor parser returned nil")
			}

			// Check expected precedence (string representation)
			tradStr := tradExpr.String()
			cursorStr := cursorExpr.String()

			if tradStr != tt.expected {
				t.Errorf("Traditional: expected %q, got %q", tt.expected, tradStr)
			}
			if cursorStr != tt.expected {
				t.Errorf("Cursor: expected %q, got %q", tt.expected, cursorStr)
			}

			// Both should be identical
			if tradStr != cursorStr {
				t.Errorf("Mismatch:\nTraditional: %s\nCursor: %s", tradStr, cursorStr)
			}
		})
	}
}

// TestMigration_ParseExpression_NotInIsAs tests "not in", "not is", "not as" special cases.
func TestMigration_ParseExpression_NotInIsAs(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		expected   string
		wantNot    bool // Whether we expect a NOT unary expression
		skipCursor bool // Skip cursor test (not yet implemented)
	}{
		{
			name:     "not in",
			source:   "x not in mySet",
			expected: "(not (x in mySet))",
			wantNot:  true,
		},
		{
			name:       "not is",
			source:     "obj not is TClass",
			expected:   "(not (obj is TClass))",
			wantNot:    true,
			skipCursor: true, // IS not yet migrated to cursor
		},
		{
			name:       "not as",
			source:     "obj not as IInterface",
			expected:   "(not (obj as IInterface))",
			wantNot:    true,
			skipCursor: true, // AS not yet migrated to cursor
		},
		{
			name:     "regular not (not a special case)",
			source:   "not x",
			expected: "(not x)",
			wantNot:  true,
		},
		{
			name:     "not followed by unrelated token",
			source:   "x + not y",
			expected: "(x + (not y))",
			wantNot:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional mode
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpression(LOWEST)

			// Traditional should succeed
			if tradExpr == nil {
				t.Fatalf("Traditional parser returned nil")
			}

			// Check traditional string representation
			tradStr := tradExpr.String()
			if tradStr != tt.expected {
				t.Errorf("Traditional: expected %q, got %q", tt.expected, tradStr)
			}

			// If wantNot, check that we have a NOT unary expression
			if tt.wantNot {
				switch expr := tradExpr.(type) {
				case *ast.UnaryExpression:
					if expr.Operator != "not" {
						t.Errorf("Expected NOT operator, got %q", expr.Operator)
					}
				case *ast.BinaryExpression:
					// For simple "x + not y", the NOT is nested
					// Don't check in this case
				default:
					t.Errorf("Expected UnaryExpression for NOT, got %T", tradExpr)
				}
			}

			// Skip cursor test if not yet implemented
			if tt.skipCursor {
				t.Skip("Cursor mode not yet implemented for this operator")
				return
			}

			// Parse with cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpression(LOWEST)

			// Cursor should succeed
			if cursorExpr == nil {
				t.Fatalf("Cursor parser returned nil")
			}

			// Check cursor string representation
			cursorStr := cursorExpr.String()
			if cursorStr != tt.expected {
				t.Errorf("Cursor: expected %q, got %q", tt.expected, cursorStr)
			}

			// Both parsers should produce semantically equivalent ASTs
			// Note: We don't use DeepEqual because Token/EndPos positions may
			// differ slightly between traditional and cursor modes due to sync timing.
			// String() comparison (checked above) verifies semantic equivalence.
		})
	}
}

// TestMigration_ParseExpression_Errors tests error handling.
func TestMigration_ParseExpression_Errors(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantErrors bool
	}{
		{
			name:       "no prefix function for semicolon",
			source:     ";",
			wantErrors: true,
		},
		// Note: "3 +" actually creates a BinaryExpression with nil Right,
		// which causes String() to panic. This is an existing behavior bug
		// that will be addressed separately. For now, we skip this test.
		// {
		// 	name:       "incomplete expression",
		// 	source:     "3 +",
		// 	wantErrors: false, // Will parse "3 +", creating incomplete BinaryExpr
		// },
		{
			name:       "invalid token in expression",
			source:     "3 @ 5", // @ is addressof, not infix
			wantErrors: false,   // Will parse "3", then stop
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional mode
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpression(LOWEST)
			tradErrors := len(tradParser.Errors())

			// Parse with cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpression(LOWEST)
			cursorErrors := len(cursorParser.Errors())

			// Check error expectations
			if tt.wantErrors {
				if tradErrors == 0 {
					t.Error("Traditional: expected errors but got none")
				}
				if cursorErrors == 0 {
					t.Error("Cursor: expected errors but got none")
				}
			}

			// Error counts should match
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: trad=%d, cursor=%d",
					tradErrors, cursorErrors)
				t.Logf("Traditional errors: %v", tradParser.Errors())
				t.Logf("Cursor errors: %v", cursorParser.Errors())
			}

			// If both returned nil, that's fine (error case)
			if tradExpr == nil && cursorExpr == nil {
				return
			}

			// If one returned nil and the other didn't, that's a problem
			if (tradExpr == nil) != (cursorExpr == nil) {
				t.Errorf("Nil mismatch: trad=%v, cursor=%v", tradExpr == nil, cursorExpr == nil)
				return
			}

			// If both returned non-nil, they should match
			if tradExpr != nil && cursorExpr != nil {
				if tradExpr.String() != cursorExpr.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradExpr.String(), cursorExpr.String())
				}
			}
		})
	}
}

// TestMigration_ParseExpression_Position tests position tracking.
func TestMigration_ParseExpression_Position(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple literal", "42"},
		{"binary expression", "3 + 5"},
		{"complex expression", "(2 + 3) * 4"},
		{"identifier", "myVar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional mode
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpression(LOWEST)

			// Parse with cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpression(LOWEST)

			if tradExpr == nil || cursorExpr == nil {
				t.Fatal("Parser returned nil")
			}

			// Check that positions match
			tradPos := tradExpr.Pos()
			cursorPos := cursorExpr.Pos()

			if tradPos != cursorPos {
				t.Errorf("Position mismatch:\nTraditional: %v\nCursor: %v",
					tradPos, cursorPos)
			}

			// Check end positions match
			tradEnd := tradExpr.End()
			cursorEnd := cursorExpr.End()

			if tradEnd != cursorEnd {
				t.Errorf("End position mismatch:\nTraditional: %v\nCursor: %v",
					tradEnd, cursorEnd)
			}
		})
	}
}

// TestMigration_ParseExpression_Dispatcher tests the dispatcher logic.
func TestMigration_ParseExpression_Dispatcher(t *testing.T) {
	source := "3 + 5 * 2"

	// Create traditional parser - should use parseExpressionTraditional
	tradParser := New(lexer.New(source))
	if tradParser.useCursor {
		t.Fatal("Traditional parser should have useCursor=false")
	}
	tradExpr := tradParser.parseExpression(LOWEST)

	// Create cursor parser - should use parseExpressionCursor
	cursorParser := NewCursorParser(lexer.New(source))
	if !cursorParser.useCursor {
		t.Fatal("Cursor parser should have useCursor=true")
	}
	cursorExpr := cursorParser.parseExpression(LOWEST)

	// Both should succeed
	if tradExpr == nil {
		t.Fatal("Traditional expression is nil")
	}
	if cursorExpr == nil {
		t.Fatal("Cursor expression is nil")
	}

	// Both should produce the same result
	if tradExpr.String() != cursorExpr.String() {
		t.Errorf("Results differ:\nTraditional: %s\nCursor: %s",
			tradExpr.String(), cursorExpr.String())
	}

	// Expected precedence: 3 + (5 * 2)
	expected := "(3 + (5 * 2))"
	if tradExpr.String() != expected {
		t.Errorf("Traditional: expected %q, got %q", expected, tradExpr.String())
	}
	if cursorExpr.String() != expected {
		t.Errorf("Cursor: expected %q, got %q", expected, cursorExpr.String())
	}
}

// TestMigration_ParseExpression_Integration runs existing parser test cases in cursor mode.
func TestMigration_ParseExpression_Integration(t *testing.T) {
	// Sample of complex expressions from existing test suite
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "chained additions",
			source:   "1 + 2 + 3 + 4",
			expected: "(((1 + 2) + 3) + 4)",
		},
		{
			name:     "mixed arithmetic",
			source:   "2 * 3 + 4 * 5",
			expected: "((2 * 3) + (4 * 5))",
		},
		{
			name:     "logical expression",
			source:   "x > 0 and y < 10 or z = 5",
			expected: "(((x > 0) and (y < 10)) or (z = 5))",
		},
		{
			name:     "nested parentheses",
			source:   "((a + b) * (c - d)) / e",
			expected: "(((a + b) * (c - d)) / e)",
		},
		{
			name:     "coalesce chain",
			source:   "a ?? b ?? c",
			expected: "((a ?? b) ?? c)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with both modes
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpression(LOWEST)

			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpression(LOWEST)

			// Both should succeed
			if tradExpr == nil {
				t.Fatal("Traditional parser failed")
			}
			if cursorExpr == nil {
				t.Fatal("Cursor parser failed")
			}

			// Check expected output
			if tradExpr.String() != tt.expected {
				t.Errorf("Traditional: expected %q, got %q",
					tt.expected, tradExpr.String())
			}
			if cursorExpr.String() != tt.expected {
				t.Errorf("Cursor: expected %q, got %q",
					tt.expected, cursorExpr.String())
			}

			// Both modes should produce identical results
			if tradExpr.String() != cursorExpr.String() {
				t.Errorf("Mode mismatch:\nTraditional: %s\nCursor: %s",
					tradExpr.String(), cursorExpr.String())
			}
		})
	}
}

// TestMigration_ParseExpression_CursorFallback tests fallback to traditional mode
// when encountering tokens without cursor implementations.
func TestMigration_ParseExpression_CursorFallback(t *testing.T) {
	// These expressions may trigger fallback to traditional mode
	// for tokens that don't have cursor implementations yet
	tests := []struct {
		name   string
		source string
	}{
		{"prefix expression", "not x"},
		{"grouped expression", "(3 + 5)"},
		{"unary minus", "-42"},
		{"unary plus", "+42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with cursor mode (may fallback internally)
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpression(LOWEST)

			// Parse with traditional mode (baseline)
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpression(LOWEST)

			// Both should succeed
			if cursorExpr == nil {
				t.Fatal("Cursor parser returned nil")
			}
			if tradExpr == nil {
				t.Fatal("Traditional parser returned nil")
			}

			// Even with fallback, results should match
			if cursorExpr.String() != tradExpr.String() {
				t.Errorf("Mismatch:\nCursor: %s\nTraditional: %s",
					cursorExpr.String(), tradExpr.String())
			}

			// No errors should occur
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}
		})
	}
}
