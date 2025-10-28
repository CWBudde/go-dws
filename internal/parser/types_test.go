package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestParseTypeExpression_SimpleType tests parsing of simple type identifiers
func TestParseTypeExpression_SimpleType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "integer type",
			input:    "Integer",
			expected: "Integer",
			wantErr:  false,
		},
		{
			name:     "string type",
			input:    "String",
			expected: "String",
			wantErr:  false,
		},
		{
			name:     "custom type",
			input:    "TMyType",
			expected: "TMyType",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			// Note: New() already calls nextToken() twice, so curToken is ready

			typeExpr := p.parseTypeExpression()

			if tt.wantErr {
				if len(p.Errors()) == 0 {
					t.Errorf("expected error but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("unexpected errors: %v", p.Errors())
				return
			}

			if typeExpr == nil {
				t.Fatal("parseTypeExpression returned nil")
			}

			if typeExpr.String() != tt.expected {
				t.Errorf("expected type %q, got %q", tt.expected, typeExpr.String())
			}
		})
	}
}

// TestParseTypeExpression_FunctionPointer tests parsing inline function pointer types
func TestParseTypeExpression_FunctionPointer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "function with one parameter",
			input:    "function(x: Integer): String",
			expected: "function(x: Integer): String",
			wantErr:  false,
		},
		{
			name:     "function with multiple parameters",
			input:    "function(x: Integer; y: String): Boolean",
			expected: "function(x: Integer; y: String): Boolean",
			wantErr:  false,
		},
		{
			name:     "procedure with one parameter",
			input:    "procedure(msg: String)",
			expected: "procedure(msg: String)",
			wantErr:  false,
		},
		{
			name:     "procedure with no parameters",
			input:    "procedure()",
			expected: "procedure()",
			wantErr:  false,
		},
		{
			name:     "function with no parameters",
			input:    "function(): Integer",
			expected: "function(): Integer",
			wantErr:  false,
		},
		{
			name:     "function of object",
			input:    "function(x: Integer): String of object",
			expected: "function(x: Integer): String of object",
			wantErr:  false,
		},
		{
			name:     "procedure of object",
			input:    "procedure(msg: String) of object",
			expected: "procedure(msg: String) of object",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			// Note: New() already calls nextToken() twice, so curToken is ready

			typeExpr := p.parseTypeExpression()

			if tt.wantErr {
				if len(p.Errors()) == 0 {
					t.Errorf("expected error but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("unexpected errors: %v", p.Errors())
				return
			}

			if typeExpr == nil {
				t.Fatal("parseTypeExpression returned nil")
			}

			if typeExpr.String() != tt.expected {
				t.Errorf("expected type %q, got %q", tt.expected, typeExpr.String())
			}
		})
	}
}

// TestParseTypeExpression_ArrayType tests parsing array of Type syntax
func TestParseTypeExpression_ArrayType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "array of integer",
			input:    "array of Integer",
			expected: "array of Integer",
			wantErr:  false,
		},
		{
			name:     "array of string",
			input:    "array of String",
			expected: "array of String",
			wantErr:  false,
		},
		{
			name:     "array of custom type",
			input:    "array of TMyType",
			expected: "array of TMyType",
			wantErr:  false,
		},
		{
			name:     "nested array",
			input:    "array of array of Integer",
			expected: "array of array of Integer",
			wantErr:  false,
		},
		{
			name:     "array of function pointer",
			input:    "array of function(x: Integer): String",
			expected: "array of function(x: Integer): String",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			// Note: New() already calls nextToken() twice, so curToken is ready

			typeExpr := p.parseTypeExpression()

			if tt.wantErr {
				if len(p.Errors()) == 0 {
					t.Errorf("expected error but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("unexpected errors: %v", p.Errors())
				return
			}

			if typeExpr == nil {
				t.Fatal("parseTypeExpression returned nil")
			}

			if typeExpr.String() != tt.expected {
				t.Errorf("expected type %q, got %q", tt.expected, typeExpr.String())
			}
		})
	}
}

// TestParseTypeExpression_ErrorCases tests various error conditions
func TestParseTypeExpression_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "function without parameter list",
			input: "function",
		},
		{
			name:  "function without return type",
			input: "function(x: Integer)",
		},
		{
			name:  "array without element type",
			input: "array of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.nextToken() // Initialize parser

			typeExpr := p.parseTypeExpression()

			// Should have errors
			if len(p.Errors()) == 0 {
				t.Errorf("expected errors but got none")
			}

			// May return nil or partial result
			if typeExpr != nil {
				t.Logf("partial result: %s", typeExpr.String())
			}
		})
	}
}
