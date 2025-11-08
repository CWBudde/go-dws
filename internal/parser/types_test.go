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
// Covers both dynamic arrays (array of Type) and static arrays (array[low..high] of Type)
func TestParseTypeExpression_ArrayType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Dynamic arrays (no bounds)
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
			name:     "nested dynamic array",
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

		// Static arrays (with bounds)
		{
			name:     "static array basic",
			input:    "array[1..10] of Integer",
			expected: "array[1..10] of Integer",
			wantErr:  false,
		},
		{
			name:     "static array zero-based",
			input:    "array[0..99] of String",
			expected: "array[0..99] of String",
			wantErr:  false,
		},
		{
			name:     "static array single element",
			input:    "array[1..1] of Boolean",
			expected: "array[1..1] of Boolean",
			wantErr:  false,
		},
		{
			name:     "static array large range",
			input:    "array[1..1000] of Float",
			expected: "array[1..1000] of Float",
			wantErr:  false,
		},
		{
			name:     "static array negative bounds",
			input:    "array[(-10)..10] of Integer",
			expected: "array[(-10)..10] of Integer",
			wantErr:  false,
		},
		{
			name:     "nested static arrays",
			input:    "array[1..5] of array[1..10] of Integer",
			expected: "array[1..5] of array[1..10] of Integer",
			wantErr:  false,
		},
		{
			name:     "static array of function pointer",
			input:    "array[1..3] of function(): Integer",
			expected: "array[1..3] of function(): Integer",
			wantErr:  false,
		},
		{
			name:     "mixed static and dynamic arrays",
			input:    "array[1..5] of array of String",
			expected: "array[1..5] of array of String",
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

// TestFunctionPointerSyntaxDetection tests the parser's ability to distinguish
// between full syntax (with parameter names) and shorthand syntax (types only).
func TestFunctionPointerSyntaxDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Full syntax variations
		{
			name:     "full syntax single parameter",
			input:    "function(x: Integer): String",
			expected: "function(x: Integer): String",
			wantErr:  false,
		},
		{
			name:     "full syntax two parameters",
			input:    "function(a, b: Integer): Boolean",
			expected: "function(a: Integer; b: Integer): Boolean",
			wantErr:  false,
		},
		{
			name:     "full syntax three parameters same type",
			input:    "function(a, b, c: Integer): Float",
			expected: "function(a: Integer; b: Integer; c: Integer): Float",
			wantErr:  false,
		},
		{
			name:     "full syntax multiple parameters different types",
			input:    "function(x: Integer; y: String): Boolean",
			expected: "function(x: Integer; y: String): Boolean",
			wantErr:  false,
		},
		{
			name:     "full syntax complex multi-param",
			input:    "function(a, b, c: Integer; d: String): Boolean",
			expected: "function(a: Integer; b: Integer; c: Integer; d: String): Boolean",
			wantErr:  false,
		},
		{
			name:     "procedure full syntax",
			input:    "procedure(x, y: Integer)",
			expected: "procedure(x: Integer; y: Integer)",
			wantErr:  false,
		},

		// Shorthand syntax variations
		{
			name:     "shorthand single type",
			input:    "function(Integer): String",
			expected: "function(Integer): String",
			wantErr:  false,
		},
		{
			name:     "shorthand two types",
			input:    "function(Integer, String): Boolean",
			expected: "function(Integer; String): Boolean",
			wantErr:  false,
		},
		{
			name:     "shorthand three types",
			input:    "function(Integer, String, Boolean): Float",
			expected: "function(Integer; String; Boolean): Float",
			wantErr:  false,
		},
		{
			name:     "shorthand with semicolons",
			input:    "function(Integer; String): Float",
			expected: "function(Integer; String): Float",
			wantErr:  false,
		},
		{
			name:     "procedure shorthand",
			input:    "procedure(Integer, String)",
			expected: "procedure(Integer; String)",
			wantErr:  false,
		},

		// Nested function pointers
		{
			name:     "nested function pointer in parameter",
			input:    "function(function(Integer): String): Boolean",
			expected: "function(function(Integer): String): Boolean",
			wantErr:  false,
		},
		{
			name:     "nested function pointer with full syntax",
			input:    "function(f: function(x: Integer): String): Boolean",
			expected: "function(f: function(x: Integer): String): Boolean",
			wantErr:  false,
		},
		{
			name:     "multiple nested function pointers",
			input:    "function(Integer, function(String): Boolean): Float",
			expected: "function(Integer; function(String): Boolean): Float",
			wantErr:  false,
		},

		// With modifiers
		{
			name:     "const parameter shorthand",
			input:    "function(const Integer): String",
			expected: "function(const Integer): String",
			wantErr:  false,
		},
		{
			name:     "const parameter full syntax",
			input:    "function(const x: Integer): String",
			expected: "function(const x: Integer): String",
			wantErr:  false,
		},
		{
			name:     "var parameter full syntax",
			input:    "function(var x: Integer): Boolean",
			expected: "function(var x: Integer): Boolean",
			wantErr:  false,
		},
		{
			name:     "var parameter shorthand",
			input:    "function(var Integer): Boolean",
			expected: "function(var Integer): Boolean",
			wantErr:  false,
		},
		{
			name:     "mixed modifiers",
			input:    "function(const x: Integer; var y: String): Boolean",
			expected: "function(const x: Integer; var y: String): Boolean",
			wantErr:  false,
		},

		// Of object variants
		{
			name:     "full syntax of object",
			input:    "function(x: Integer): String of object",
			expected: "function(x: Integer): String of object",
			wantErr:  false,
		},
		{
			name:     "shorthand of object",
			input:    "function(Integer): String of object",
			expected: "function(Integer): String of object",
			wantErr:  false,
		},
		{
			name:     "procedure full syntax of object",
			input:    "procedure(x, y: Integer) of object",
			expected: "procedure(x: Integer; y: Integer) of object",
			wantErr:  false,
		},
		{
			name:     "procedure shorthand of object",
			input:    "procedure(Integer, String) of object",
			expected: "procedure(Integer; String) of object",
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
		// Static array error cases
		{
			name:  "array with missing low bound",
			input: "array[..10] of Integer",
		},
		{
			name:  "array with missing high bound",
			input: "array[1..] of Integer",
		},
		{
			name:  "array with missing dotdot",
			input: "array[1 10] of Integer",
		},
		{
			name:  "array with invalid bounds (low > high)",
			input: "array[10..1] of Integer",
		},
		{
			name:  "array with missing closing bracket",
			input: "array[1..10 of Integer",
		},
		{
			name:  "array with bounds but no 'of'",
			input: "array[1..10] Integer",
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
