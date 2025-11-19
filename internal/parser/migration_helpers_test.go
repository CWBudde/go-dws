package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.2.10 Phase 6: Comprehensive tests for expression helper migration
//
// This file tests the cursor-mode migration of helper functions used by parseCallOrRecordLiteral:
// - parseExpressionListCursor
// - parseEmptyCallCursor
// - parseCallWithExpressionListCursor
// - parseNamedFieldInitializerCursor
// - parseArgumentAsFieldInitializerCursor
// - parseSingleArgumentOrFieldCursor
// - advanceToNextItemCursor
// - parseArgumentsOrFieldsCursor
// - parseCallOrRecordLiteralCursor
//
// All tests use differential testing: run in both traditional and cursor mode,
// compare results for equality.

// TestMigration_ParseExpressionList tests parseExpressionListCursor against traditional mode
func TestMigration_ParseExpressionList(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectedLen int
		description string
	}{
		{
			name:        "empty list",
			source:      "Foo()",
			expectedLen: 0,
			description: "Empty argument list",
		},
		{
			name:        "single element",
			source:      "Foo(42)",
			expectedLen: 1,
			description: "Single integer argument",
		},
		{
			name:        "multiple elements",
			source:      "Foo(1, 2, 3)",
			expectedLen: 3,
			description: "Multiple integer arguments",
		},
		{
			name:        "trailing comma",
			source:      "Foo(1, 2, 3,)",
			expectedLen: 3,
			description: "Trailing comma should be allowed",
		},
		{
			name:        "complex expressions",
			source:      "Foo(2 + 3, x * y, true)",
			expectedLen: 3,
			description: "Complex expressions as arguments",
		},
		{
			name:        "nested calls",
			source:      "Foo(Bar(1), Baz(2, 3))",
			expectedLen: 2,
			description: "Nested function calls",
		},
		{
			name:        "string arguments",
			source:      `Foo("hello", "world")`,
			expectedLen: 2,
			description: "String literal arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpressionCursor(LOWEST)

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpressionCursor(LOWEST)

			// Both should succeed
			if tradExpr == nil {
				t.Error("Traditional parser returned nil expression")
			}
			if cursorExpr == nil {
				t.Error("Cursor parser returned nil expression")
				return
			}

			// Both should be CallExpression
			tradCall, ok := tradExpr.(*ast.CallExpression)
			if !ok {
				t.Errorf("Traditional: expected CallExpression, got %T", tradExpr)
				return
			}
			cursorCall, ok := cursorExpr.(*ast.CallExpression)
			if !ok {
				t.Errorf("Cursor: expected CallExpression, got %T", cursorExpr)
				return
			}

			// Check argument count
			if len(tradCall.Arguments) != tt.expectedLen {
				t.Errorf("Traditional: expected %d arguments, got %d", tt.expectedLen, len(tradCall.Arguments))
			}
			if len(cursorCall.Arguments) != tt.expectedLen {
				t.Errorf("Cursor: expected %d arguments, got %d", tt.expectedLen, len(cursorCall.Arguments))
			}

			// Compare semantic equivalence via String() representation
			// Note: DeepEqual may fail due to Token/EndPos position differences
			if tradCall != nil && cursorCall != nil {
				if tradCall.String() != cursorCall.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradCall.String(), cursorCall.String())
				}
			}
		})
	}
}

// TestMigration_ParseEmptyCall tests parseEmptyCallCursor
func TestMigration_ParseEmptyCall(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "simple empty call",
			source: "Foo()",
		},
		{
			name:   "typed empty call",
			source: "MyType()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpressionCursor(LOWEST)

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpressionCursor(LOWEST)

			// Both should be CallExpression with empty Arguments
			tradCall, ok := tradExpr.(*ast.CallExpression)
			if !ok {
				t.Errorf("Traditional: expected CallExpression, got %T", tradExpr)
				return
			}
			cursorCall, ok := cursorExpr.(*ast.CallExpression)
			if !ok {
				t.Errorf("Cursor: expected CallExpression, got %T", cursorExpr)
				return
			}

			// Both should have zero arguments
			if len(tradCall.Arguments) != 0 {
				t.Errorf("Traditional: expected 0 arguments, got %d", len(tradCall.Arguments))
			}
			if len(cursorCall.Arguments) != 0 {
				t.Errorf("Cursor: expected 0 arguments, got %d", len(cursorCall.Arguments))
			}

			// Compare semantic equivalence via String() representation
			// Note: DeepEqual may fail due to Token/EndPos position differences
			if tradCall != nil && cursorCall != nil {
				if tradCall.String() != cursorCall.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradCall.String(), cursorCall.String())
				}
			}
		})
	}
}

// TestMigration_ParseCallWithExpressionList tests parseCallWithExpressionListCursor
func TestMigration_ParseCallWithExpressionList(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectedLen int
		description string
	}{
		{
			name:        "single argument",
			source:      "Foo(42)",
			expectedLen: 1,
			description: "Single integer argument",
		},
		{
			name:        "multiple arguments",
			source:      "Add(1, 2)",
			expectedLen: 2,
			description: "Two integer arguments",
		},
		{
			name:        "complex expressions",
			source:      "Calculate(2 + 3, x * y, z / 2)",
			expectedLen: 3,
			description: "Complex binary expressions",
		},
		{
			name:        "mixed types",
			source:      `Mix(42, "hello", true, 3.14)`,
			expectedLen: 4,
			description: "Mixed argument types",
		},
		{
			name:        "nested function calls",
			source:      "Outer(Inner(1, 2), Inner(3, 4))",
			expectedLen: 2,
			description: "Nested function calls as arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpressionCursor(LOWEST)

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpressionCursor(LOWEST)

			// Both should be CallExpression
			tradCall, ok := tradExpr.(*ast.CallExpression)
			if !ok {
				t.Errorf("Traditional: expected CallExpression, got %T", tradExpr)
				return
			}
			cursorCall, ok := cursorExpr.(*ast.CallExpression)
			if !ok {
				t.Errorf("Cursor: expected CallExpression, got %T", cursorExpr)
				return
			}

			// Check argument count
			if len(tradCall.Arguments) != tt.expectedLen {
				t.Errorf("Traditional: expected %d arguments, got %d", tt.expectedLen, len(tradCall.Arguments))
			}
			if len(cursorCall.Arguments) != tt.expectedLen {
				t.Errorf("Cursor: expected %d arguments, got %d", tt.expectedLen, len(cursorCall.Arguments))
			}

			// Compare semantic equivalence via String() representation
			// Note: DeepEqual may fail due to Token/EndPos position differences
			if tradCall != nil && cursorCall != nil {
				if tradCall.String() != cursorCall.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradCall.String(), cursorCall.String())
				}
			}
		})
	}
}

// TestMigration_ParseArgumentsOrFields tests parseArgumentsOrFieldsCursor
func TestMigration_ParseArgumentsOrFields(t *testing.T) {
	tests := []struct {
		name               string
		source             string
		expectedType       string // "call" or "record"
		expectedFieldCount int
		description        string
	}{
		{
			name:               "all function arguments",
			source:             "Foo(1, 2, 3)",
			expectedType:       "call",
			expectedFieldCount: 3,
			description:        "All plain expressions -> function call",
		},
		{
			name:               "all field initializers",
			source:             "Point(x: 10, y: 20)",
			expectedType:       "record",
			expectedFieldCount: 2,
			description:        "All named fields -> record literal",
		},
		{
			name:               "single field",
			source:             "Wrapper(value: 42)",
			expectedType:       "record",
			expectedFieldCount: 1,
			description:        "Single named field -> record literal",
		},
		{
			name:               "complex field values",
			source:             "Complex(real: 2 + 3, imag: x * y)",
			expectedType:       "record",
			expectedFieldCount: 2,
			description:        "Complex expressions as field values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpressionCursor(LOWEST)

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpressionCursor(LOWEST)

			// Both should succeed
			if tradExpr == nil {
				t.Error("Traditional parser returned nil expression")
			}
			if cursorExpr == nil {
				t.Error("Cursor parser returned nil expression")
				return
			}

			// Check expected type
			if tt.expectedType == "record" {
				tradRec, ok := tradExpr.(*ast.RecordLiteralExpression)
				if !ok {
					t.Errorf("Traditional: expected RecordLiteralExpression, got %T", tradExpr)
					return
				}
				cursorRec, ok := cursorExpr.(*ast.RecordLiteralExpression)
				if !ok {
					t.Errorf("Cursor: expected RecordLiteralExpression, got %T", cursorExpr)
					return
				}

				// Check field count
				if len(tradRec.Fields) != tt.expectedFieldCount {
					t.Errorf("Traditional: expected %d fields, got %d", tt.expectedFieldCount, len(tradRec.Fields))
				}
				if len(cursorRec.Fields) != tt.expectedFieldCount {
					t.Errorf("Cursor: expected %d fields, got %d", tt.expectedFieldCount, len(cursorRec.Fields))
				}

				// Compare semantic equivalence via String() representation
				if tradRec != nil && cursorRec != nil {
					if tradRec.String() != cursorRec.String() {
						t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
							tradRec.String(), cursorRec.String())
					}
				}
			} else { // "call"
				tradCall, ok := tradExpr.(*ast.CallExpression)
				if !ok {
					t.Errorf("Traditional: expected CallExpression, got %T", tradExpr)
					return
				}
				cursorCall, ok := cursorExpr.(*ast.CallExpression)
				if !ok {
					t.Errorf("Cursor: expected CallExpression, got %T", cursorExpr)
					return
				}

				// Check argument count
				if len(tradCall.Arguments) != tt.expectedFieldCount {
					t.Errorf("Traditional: expected %d arguments, got %d", tt.expectedFieldCount, len(tradCall.Arguments))
				}
				if len(cursorCall.Arguments) != tt.expectedFieldCount {
					t.Errorf("Cursor: expected %d arguments, got %d", tt.expectedFieldCount, len(cursorCall.Arguments))
				}

				// Compare semantic equivalence via String() representation
				if tradCall != nil && cursorCall != nil {
					if tradCall.String() != cursorCall.String() {
						t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
							tradCall.String(), cursorCall.String())
					}
				}
			}
		})
	}
}

// TestMigration_ParseCallOrRecordLiteral tests the top-level orchestrator
func TestMigration_ParseCallOrRecordLiteral(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		expectedType string // "call" or "record"
		description  string
	}{
		{
			name:         "empty call",
			source:       "Foo()",
			expectedType: "call",
			description:  "Empty parentheses -> function call",
		},
		{
			name:         "simple function call",
			source:       "Add(1, 2)",
			expectedType: "call",
			description:  "Plain arguments -> function call",
		},
		{
			name:         "simple record literal",
			source:       "Point(x: 10, y: 20)",
			expectedType: "record",
			description:  "All named fields -> record literal",
		},
		{
			name:         "call with non-ident first arg",
			source:       "Foo(42, 'hello')",
			expectedType: "call",
			description:  "Non-identifier first element -> function call",
		},
		{
			name:         "call with expression first arg",
			source:       "Foo(2 + 3, x * y)",
			expectedType: "call",
			description:  "Expression first element -> function call",
		},
		{
			name:         "record with single field",
			source:       "Wrapper(value: 42)",
			expectedType: "record",
			description:  "Single named field -> record literal",
		},
		{
			name:         "record with complex values",
			source:       "Complex(a: 1 + 2, b: x * y, c: true)",
			expectedType: "record",
			description:  "Named fields with complex expressions",
		},
		{
			name:         "call with string literal",
			source:       `Message("hello")`,
			expectedType: "call",
			description:  "String literal argument -> function call",
		},
		{
			name:         "nested calls",
			source:       "Outer(Inner(1, 2))",
			expectedType: "call",
			description:  "Nested function calls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpressionCursor(LOWEST)

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpressionCursor(LOWEST)

			// Both should succeed
			if tradExpr == nil {
				t.Error("Traditional parser returned nil expression")
			}
			if cursorExpr == nil {
				t.Error("Cursor parser returned nil expression")
				return
			}

			// Verify expected type
			if tt.expectedType == "record" {
				if _, ok := tradExpr.(*ast.RecordLiteralExpression); !ok {
					t.Errorf("Traditional: expected RecordLiteralExpression, got %T", tradExpr)
				}
				if _, ok := cursorExpr.(*ast.RecordLiteralExpression); !ok {
					t.Errorf("Cursor: expected RecordLiteralExpression, got %T", cursorExpr)
				}
			} else {
				if _, ok := tradExpr.(*ast.CallExpression); !ok {
					t.Errorf("Traditional: expected CallExpression, got %T", tradExpr)
				}
				if _, ok := cursorExpr.(*ast.CallExpression); !ok {
					t.Errorf("Cursor: expected CallExpression, got %T", cursorExpr)
				}
			}

			// String representations should match (semantic equivalence)
			// Note: DeepEqual may fail due to Token/EndPos position differences
			if tradExpr != nil && cursorExpr != nil {
				if tradExpr.String() != cursorExpr.String() {
					t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
						tradExpr.String(), cursorExpr.String())
				}
			}
		})
	}
}

// TestMigration_Helpers_EdgeCases tests edge cases and error scenarios
func TestMigration_Helpers_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectError bool
		description string
	}{
		{
			name:        "trailing comma in call",
			source:      "Foo(1, 2, 3,)",
			expectError: false,
			description: "Trailing comma should be allowed",
		},
		{
			name:        "trailing comma in record",
			source:      "Point(x: 1, y: 2,)",
			expectError: false,
			description: "Trailing comma in record literal",
		},
		{
			name:        "single argument call",
			source:      "Single(42)",
			expectError: false,
			description: "Single argument should work",
		},
		{
			name:        "deeply nested",
			source:      "A(B(C(D(1))))",
			expectError: false,
			description: "Deeply nested calls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			tradParser := New(lexer.New(tt.source))
			tradExpr := tradParser.parseExpressionCursor(LOWEST)
			tradErrors := len(tradParser.Errors())

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorExpr := cursorParser.parseExpressionCursor(LOWEST)
			cursorErrors := len(cursorParser.Errors())

			// Error counts should match
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
				if cursorErrors > 0 {
					t.Logf("Cursor errors:")
					for _, err := range cursorParser.Errors() {
						t.Logf("  %v", err)
					}
				}
			}

			// If we expect no errors, both should succeed
			if !tt.expectError {
				if tradExpr == nil {
					t.Error("Traditional parser returned nil")
				}
				if cursorExpr == nil {
					t.Error("Cursor parser returned nil")
				}

				// Compare semantic equivalence via String() representation
				if tradExpr != nil && cursorExpr != nil {
					if tradExpr.String() != cursorExpr.String() {
						t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
							tradExpr.String(), cursorExpr.String())
					}
				}
			}
		})
	}
}
