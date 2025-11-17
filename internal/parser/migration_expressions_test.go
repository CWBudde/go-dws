package parser

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestMigration_Identifier_Basic tests basic identifier parsing in both modes
func TestMigration_Identifier_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple identifier",
			input:    "myVar",
			expected: "myVar",
		},
		{
			name:     "uppercase identifier",
			input:    "MYVAR",
			expected: "MYVAR",
		},
		{
			name:     "mixed case identifier",
			input:    "MyVar",
			expected: "MyVar",
		},
		{
			name:     "identifier with underscore",
			input:    "my_var",
			expected: "my_var",
		},
		{
			name:     "identifier with numbers",
			input:    "var123",
			expected: "var123",
		},
		{
			name:     "single letter",
			input:    "x",
			expected: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test traditional mode
			traditionalParser := New(lexer.New(tt.input))
			traditionalExpr := traditionalParser.parseIdentifierTraditional()

			// Test cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorExpr := cursorParser.parseIdentifierCursor()

			// Both should succeed
			if traditionalExpr == nil {
				t.Error("Traditional parser returned nil")
			}
			if cursorExpr == nil {
				t.Error("Cursor parser returned nil")
			}

			// Both should be Identifier
			traditionalIdent, ok := traditionalExpr.(*ast.Identifier)
			if !ok {
				t.Errorf("Traditional parser returned %T, want *ast.Identifier", traditionalExpr)
			}
			cursorIdent, ok := cursorExpr.(*ast.Identifier)
			if !ok {
				t.Errorf("Cursor parser returned %T, want *ast.Identifier", cursorExpr)
			}

			// Values should match
			if traditionalIdent != nil && traditionalIdent.Value != tt.expected {
				t.Errorf("Traditional parser value = %q, want %q", traditionalIdent.Value, tt.expected)
			}
			if cursorIdent != nil && cursorIdent.Value != tt.expected {
				t.Errorf("Cursor parser value = %q, want %q", cursorIdent.Value, tt.expected)
			}

			// Values should be identical
			if traditionalIdent != nil && cursorIdent != nil {
				if traditionalIdent.Value != cursorIdent.Value {
					t.Errorf("Value mismatch: traditional=%q, cursor=%q",
						traditionalIdent.Value, cursorIdent.Value)
				}
			}

			// Token should be identical
			if traditionalIdent != nil && cursorIdent != nil {
				if traditionalIdent.Token.Literal != cursorIdent.Token.Literal {
					t.Errorf("Token literal mismatch: traditional=%q, cursor=%q",
						traditionalIdent.Token.Literal, cursorIdent.Token.Literal)
				}
			}
		})
	}
}

// TestMigration_Identifier_Dispatcher tests the dispatcher function
func TestMigration_Identifier_Dispatcher(t *testing.T) {
	input := "myVar"

	// Traditional parser should use traditional implementation
	traditionalParser := New(lexer.New(input))
	if traditionalParser.useCursor {
		t.Error("Traditional parser should have useCursor=false")
	}
	traditionalExpr := traditionalParser.parseIdentifier()
	if traditionalExpr == nil {
		t.Error("Traditional parser returned nil")
	}

	// Cursor parser should use cursor implementation
	cursorParser := NewCursorParser(lexer.New(input))
	if !cursorParser.useCursor {
		t.Error("Cursor parser should have useCursor=true")
	}
	cursorExpr := cursorParser.parseIdentifier()
	if cursorExpr == nil {
		t.Error("Cursor parser returned nil")
	}

	// Both should produce same result
	if traditionalIdent, ok := traditionalExpr.(*ast.Identifier); ok {
		if cursorIdent, ok := cursorExpr.(*ast.Identifier); ok {
			if traditionalIdent.Value != cursorIdent.Value {
				t.Errorf("Value mismatch: traditional=%q, cursor=%q",
					traditionalIdent.Value, cursorIdent.Value)
			}
		}
	}
}

// TestMigration_Identifier_Position tests position tracking in both modes
func TestMigration_Identifier_Position(t *testing.T) {
	input := "  myVar  " // With leading/trailing whitespace

	traditionalParser := New(lexer.New(input))
	traditionalExpr := traditionalParser.parseIdentifierTraditional()

	cursorParser := NewCursorParser(lexer.New(input))
	cursorExpr := cursorParser.parseIdentifierCursor()

	// Both should track position correctly
	if traditionalIdent, ok := traditionalExpr.(*ast.Identifier); ok {
		if cursorIdent, ok := cursorExpr.(*ast.Identifier); ok {
			// Positions should match
			if traditionalIdent.Token.Pos.Line != cursorIdent.Token.Pos.Line {
				t.Errorf("Line mismatch: traditional=%d, cursor=%d",
					traditionalIdent.Token.Pos.Line, cursorIdent.Token.Pos.Line)
			}
			if traditionalIdent.Token.Pos.Column != cursorIdent.Token.Pos.Column {
				t.Errorf("Column mismatch: traditional=%d, cursor=%d",
					traditionalIdent.Token.Pos.Column, cursorIdent.Token.Pos.Column)
			}

			// EndPos should match
			if traditionalIdent.EndPos.Line != cursorIdent.EndPos.Line {
				t.Errorf("EndPos Line mismatch: traditional=%d, cursor=%d",
					traditionalIdent.EndPos.Line, cursorIdent.EndPos.Line)
			}
			if traditionalIdent.EndPos.Column != cursorIdent.EndPos.Column {
				t.Errorf("EndPos Column mismatch: traditional=%d, cursor=%d",
					traditionalIdent.EndPos.Column, cursorIdent.EndPos.Column)
			}
		}
	}
}

// TestMigration_FloatLiteral_Basic tests basic float parsing in both modes
func TestMigration_FloatLiteral_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "zero",
			input:    "0.0",
			expected: 0.0,
		},
		{
			name:     "simple positive",
			input:    "3.14",
			expected: 3.14,
		},
		{
			name:     "large number",
			input:    "123456.789",
			expected: 123456.789,
		},
		{
			name:     "scientific notation",
			input:    "1.5e10",
			expected: 1.5e10,
		},
		{
			name:     "scientific negative exponent",
			input:    "2.5e-3",
			expected: 2.5e-3,
		},
		{
			name:     "very small number",
			input:    "0.0001",
			expected: 0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test traditional mode
			traditionalParser := New(lexer.New(tt.input))
			traditionalExpr := traditionalParser.parseFloatLiteralTraditional()

			// Test cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorExpr := cursorParser.parseFloatLiteralCursor()

			// Both should succeed
			if traditionalExpr == nil {
				t.Error("Traditional parser returned nil")
			}
			if cursorExpr == nil {
				t.Error("Cursor parser returned nil")
			}

			// Both should be FloatLiteral
			traditionalLit, ok := traditionalExpr.(*ast.FloatLiteral)
			if !ok {
				t.Errorf("Traditional parser returned %T, want *ast.FloatLiteral", traditionalExpr)
			}
			cursorLit, ok := cursorExpr.(*ast.FloatLiteral)
			if !ok {
				t.Errorf("Cursor parser returned %T, want *ast.FloatLiteral", cursorExpr)
			}

			// Values should match
			if traditionalLit != nil && traditionalLit.Value != tt.expected {
				t.Errorf("Traditional parser value = %f, want %f", traditionalLit.Value, tt.expected)
			}
			if cursorLit != nil && cursorLit.Value != tt.expected {
				t.Errorf("Cursor parser value = %f, want %f", cursorLit.Value, tt.expected)
			}

			// Values should be identical
			if traditionalLit != nil && cursorLit != nil {
				if traditionalLit.Value != cursorLit.Value {
					t.Errorf("Value mismatch: traditional=%f, cursor=%f",
						traditionalLit.Value, cursorLit.Value)
				}
			}

			// Token should be identical
			if traditionalLit != nil && cursorLit != nil {
				if traditionalLit.Token.Literal != cursorLit.Token.Literal {
					t.Errorf("Token literal mismatch: traditional=%q, cursor=%q",
						traditionalLit.Token.Literal, cursorLit.Token.Literal)
				}
			}
		})
	}
}

// TestMigration_FloatLiteral_Errors tests error handling in both modes
func TestMigration_FloatLiteral_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid format",
			input: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: The lexer must produce a FLOAT token for these tests to be meaningful.
			// If the lexer rejects the input, we skip the test.

			// Test traditional mode
			traditionalParser := New(lexer.New(tt.input))
			if traditionalParser.curToken.Type != lexer.FLOAT {
				t.Skip("Lexer did not produce FLOAT token")
			}
			traditionalExpr := traditionalParser.parseFloatLiteralTraditional()

			// Test cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			if cursorParser.cursor.Current().Type != lexer.FLOAT {
				t.Skip("Lexer did not produce FLOAT token")
			}
			cursorExpr := cursorParser.parseFloatLiteralCursor()

			// Both should return nil on error
			if traditionalExpr != nil {
				t.Errorf("Traditional parser should return nil on error, got %v", traditionalExpr)
			}
			if cursorExpr != nil {
				t.Errorf("Cursor parser should return nil on error, got %v", cursorExpr)
			}

			// Both should have errors
			if len(traditionalParser.Errors()) == 0 {
				t.Error("Traditional parser should have errors")
			}
			if len(cursorParser.Errors()) == 0 {
				t.Error("Cursor parser should have errors")
			}
		})
	}
}

// TestMigration_StringLiteral_Basic tests basic string parsing in both modes
func TestMigration_StringLiteral_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string single quotes",
			input:    "''",
			expected: "",
		},
		{
			name:     "empty string double quotes",
			input:    `""`,
			expected: "",
		},
		{
			name:     "simple string single quotes",
			input:    "'hello'",
			expected: "hello",
		},
		{
			name:     "simple string double quotes",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "string with spaces",
			input:    "'hello world'",
			expected: "hello world",
		},
		{
			name:     "string with escaped quotes",
			input:    "'it''s'",
			expected: "it's",
		},
		{
			name:     "string with multiple escaped quotes",
			input:    "'can''t won''t'",
			expected: "can't won't",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test traditional mode
			traditionalParser := New(lexer.New(tt.input))
			traditionalExpr := traditionalParser.parseStringLiteralTraditional()

			// Test cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorExpr := cursorParser.parseStringLiteralCursor()

			// Both should succeed
			if traditionalExpr == nil {
				t.Error("Traditional parser returned nil")
			}
			if cursorExpr == nil {
				t.Error("Cursor parser returned nil")
			}

			// Both should be StringLiteral
			traditionalLit, ok := traditionalExpr.(*ast.StringLiteral)
			if !ok {
				t.Errorf("Traditional parser returned %T, want *ast.StringLiteral", traditionalExpr)
			}
			cursorLit, ok := cursorExpr.(*ast.StringLiteral)
			if !ok {
				t.Errorf("Cursor parser returned %T, want *ast.StringLiteral", cursorExpr)
			}

			// Values should match
			if traditionalLit != nil && traditionalLit.Value != tt.expected {
				t.Errorf("Traditional parser value = %q, want %q", traditionalLit.Value, tt.expected)
			}
			if cursorLit != nil && cursorLit.Value != tt.expected {
				t.Errorf("Cursor parser value = %q, want %q", cursorLit.Value, tt.expected)
			}

			// Values should be identical
			if traditionalLit != nil && cursorLit != nil {
				if traditionalLit.Value != cursorLit.Value {
					t.Errorf("Value mismatch: traditional=%q, cursor=%q",
						traditionalLit.Value, cursorLit.Value)
				}
			}
		})
	}
}

// TestMigration_BooleanLiteral_Basic tests basic boolean parsing in both modes
func TestMigration_BooleanLiteral_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "true lowercase",
			input:    "true",
			expected: true,
		},
		{
			name:     "TRUE uppercase",
			input:    "TRUE",
			expected: true,
		},
		{
			name:     "True mixed case",
			input:    "True",
			expected: true,
		},
		{
			name:     "false lowercase",
			input:    "false",
			expected: false,
		},
		{
			name:     "FALSE uppercase",
			input:    "FALSE",
			expected: false,
		},
		{
			name:     "False mixed case",
			input:    "False",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test traditional mode
			traditionalParser := New(lexer.New(tt.input))
			traditionalExpr := traditionalParser.parseBooleanLiteralTraditional()

			// Test cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorExpr := cursorParser.parseBooleanLiteralCursor()

			// Both should succeed
			if traditionalExpr == nil {
				t.Error("Traditional parser returned nil")
			}
			if cursorExpr == nil {
				t.Error("Cursor parser returned nil")
			}

			// Both should be BooleanLiteral
			traditionalLit, ok := traditionalExpr.(*ast.BooleanLiteral)
			if !ok {
				t.Errorf("Traditional parser returned %T, want *ast.BooleanLiteral", traditionalExpr)
			}
			cursorLit, ok := cursorExpr.(*ast.BooleanLiteral)
			if !ok {
				t.Errorf("Cursor parser returned %T, want *ast.BooleanLiteral", cursorExpr)
			}

			// Values should match
			if traditionalLit != nil && traditionalLit.Value != tt.expected {
				t.Errorf("Traditional parser value = %v, want %v", traditionalLit.Value, tt.expected)
			}
			if cursorLit != nil && cursorLit.Value != tt.expected {
				t.Errorf("Cursor parser value = %v, want %v", cursorLit.Value, tt.expected)
			}

			// Values should be identical
			if traditionalLit != nil && cursorLit != nil {
				if traditionalLit.Value != cursorLit.Value {
					t.Errorf("Value mismatch: traditional=%v, cursor=%v",
						traditionalLit.Value, cursorLit.Value)
				}
			}

			// Token type should be correct
			if traditionalLit != nil && cursorLit != nil {
				expectedType := lexer.TRUE
				if !tt.expected {
					expectedType = lexer.FALSE
				}
				if traditionalLit.Token.Type != expectedType {
					t.Errorf("Traditional token type = %v, want %v",
						traditionalLit.Token.Type, expectedType)
				}
				if cursorLit.Token.Type != expectedType {
					t.Errorf("Cursor token type = %v, want %v",
						cursorLit.Token.Type, expectedType)
				}
			}
		})
	}
}

// TestMigration_BooleanLiteral_Dispatcher tests the dispatcher function
func TestMigration_BooleanLiteral_Dispatcher(t *testing.T) {
	input := "true"

	// Traditional parser should use traditional implementation
	traditionalParser := New(lexer.New(input))
	if traditionalParser.useCursor {
		t.Error("Traditional parser should have useCursor=false")
	}
	traditionalExpr := traditionalParser.parseBooleanLiteral()
	if traditionalExpr == nil {
		t.Error("Traditional parser returned nil")
	}

	// Cursor parser should use cursor implementation
	cursorParser := NewCursorParser(lexer.New(input))
	if !cursorParser.useCursor {
		t.Error("Cursor parser should have useCursor=true")
	}
	cursorExpr := cursorParser.parseBooleanLiteral()
	if cursorExpr == nil {
		t.Error("Cursor parser returned nil")
	}

	// Both should produce same result
	if traditionalLit, ok := traditionalExpr.(*ast.BooleanLiteral); ok {
		if cursorLit, ok := cursorExpr.(*ast.BooleanLiteral); ok {
			if traditionalLit.Value != cursorLit.Value {
				t.Errorf("Value mismatch: traditional=%v, cursor=%v",
					traditionalLit.Value, cursorLit.Value)
			}
		}
	}
}

// TestMigration_AllExpressions_Integration tests all migrated expressions together
func TestMigration_AllExpressions_Integration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		parseFunc   string
		expectedAST string
	}{
		{
			name:        "identifier",
			input:       "myVar",
			parseFunc:   "identifier",
			expectedAST: "*ast.Identifier",
		},
		{
			name:        "integer",
			input:       "42",
			parseFunc:   "integer",
			expectedAST: "*ast.IntegerLiteral",
		},
		{
			name:        "float",
			input:       "3.14",
			parseFunc:   "float",
			expectedAST: "*ast.FloatLiteral",
		},
		{
			name:        "string",
			input:       "'hello'",
			parseFunc:   "string",
			expectedAST: "*ast.StringLiteral",
		},
		{
			name:        "boolean",
			input:       "true",
			parseFunc:   "boolean",
			expectedAST: "*ast.BooleanLiteral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with traditional parser
			traditionalParser := New(lexer.New(tt.input))
			var traditionalExpr ast.Expression
			switch tt.parseFunc {
			case "identifier":
				traditionalExpr = traditionalParser.parseIdentifier()
			case "integer":
				traditionalExpr = traditionalParser.parseIntegerLiteral()
			case "float":
				traditionalExpr = traditionalParser.parseFloatLiteral()
			case "string":
				traditionalExpr = traditionalParser.parseStringLiteral()
			case "boolean":
				traditionalExpr = traditionalParser.parseBooleanLiteral()
			}

			// Parse with cursor parser
			cursorParser := NewCursorParser(lexer.New(tt.input))
			var cursorExpr ast.Expression
			switch tt.parseFunc {
			case "identifier":
				cursorExpr = cursorParser.parseIdentifier()
			case "integer":
				cursorExpr = cursorParser.parseIntegerLiteral()
			case "float":
				cursorExpr = cursorParser.parseFloatLiteral()
			case "string":
				cursorExpr = cursorParser.parseStringLiteral()
			case "boolean":
				cursorExpr = cursorParser.parseBooleanLiteral()
			}

			// Both should succeed
			if traditionalExpr == nil {
				t.Error("Traditional parser returned nil")
			}
			if cursorExpr == nil {
				t.Error("Cursor parser returned nil")
			}

			// Both should have same type
			if traditionalExpr != nil && cursorExpr != nil {
				traditionalType := fmt.Sprintf("%T", traditionalExpr)
				cursorType := fmt.Sprintf("%T", cursorExpr)
				if traditionalType != cursorType {
					t.Errorf("Type mismatch: traditional=%s, cursor=%s",
						traditionalType, cursorType)
				}
				if traditionalType != tt.expectedAST {
					t.Errorf("Type = %s, want %s", traditionalType, tt.expectedAST)
				}
			}

			// Both should have same string representation
			if traditionalExpr != nil && cursorExpr != nil {
				if traditionalExpr.String() != cursorExpr.String() {
					t.Errorf("String mismatch: traditional=%q, cursor=%q",
						traditionalExpr.String(), cursorExpr.String())
				}
			}
		})
	}
}
