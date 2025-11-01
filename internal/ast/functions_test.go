package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestParameterString(t *testing.T) {
	tests := []struct {
		name     string
		param    *Parameter
		expected string
	}{
		{
			name: "simple parameter",
			param: &Parameter{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				Name:  &Identifier{Value: "x"},
				Type:  &TypeAnnotation{Name: "Integer"},
				ByRef: false,
			},
			expected: "x: Integer",
		},
		{
			name: "var parameter (by reference)",
			param: &Parameter{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "s"},
				Name:  &Identifier{Value: "s"},
				Type:  &TypeAnnotation{Name: "String"},
				ByRef: true,
			},
			expected: "var s: String",
		},
		{
			name: "float parameter",
			param: &Parameter{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				Name:  &Identifier{Value: "x"},
				Type:  &TypeAnnotation{Name: "Float"},
				ByRef: false,
			},
			expected: "x: Float",
		},
		{
			name: "lazy parameter",
			param: &Parameter{
				Token:  lexer.Token{Type: lexer.IDENT, Literal: "expr"},
				Name:   &Identifier{Value: "expr"},
				Type:   &TypeAnnotation{Name: "Integer"},
				IsLazy: true,
				ByRef:  false,
			},
			expected: "lazy expr: Integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.param.String()
			if result != tt.expected {
				t.Errorf("Parameter.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFunctionDeclString(t *testing.T) {
	tests := []struct {
		name     string
		fn       *FunctionDecl
		expected string
	}{
		{
			name: "simple function with no parameters",
			fn: &FunctionDecl{
				Token:      lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
				Name:       &Identifier{Value: "GetValue"},
				Parameters: []*Parameter{},
				ReturnType: &TypeAnnotation{Name: "Integer"},
				Body: &BlockStatement{
					Statements: []Statement{},
				},
			},
			expected: "function GetValue(): Integer begin\nend",
		},
		{
			name: "function with single parameter",
			fn: &FunctionDecl{
				Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
				Name:  &Identifier{Value: "Double"},
				Parameters: []*Parameter{
					{
						Name:  &Identifier{Value: "x"},
						Type:  &TypeAnnotation{Name: "Integer"},
						ByRef: false,
					},
				},
				ReturnType: &TypeAnnotation{Name: "Integer"},
				Body: &BlockStatement{
					Statements: []Statement{},
				},
			},
			expected: "function Double(x: Integer): Integer begin\nend",
		},
		{
			name: "function with multiple parameters",
			fn: &FunctionDecl{
				Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
				Name:  &Identifier{Value: "Add"},
				Parameters: []*Parameter{
					{
						Name:  &Identifier{Value: "a"},
						Type:  &TypeAnnotation{Name: "Integer"},
						ByRef: false,
					},
					{
						Name:  &Identifier{Value: "b"},
						Type:  &TypeAnnotation{Name: "Integer"},
						ByRef: false,
					},
				},
				ReturnType: &TypeAnnotation{Name: "Integer"},
				Body: &BlockStatement{
					Statements: []Statement{},
				},
			},
			expected: "function Add(a: Integer; b: Integer): Integer begin\nend",
		},
		{
			name: "procedure with no return type",
			fn: &FunctionDecl{
				Token:      lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
				Name:       &Identifier{Value: "Hello"},
				Parameters: []*Parameter{},
				ReturnType: nil,
				Body: &BlockStatement{
					Statements: []Statement{},
				},
			},
			expected: "procedure Hello begin\nend",
		},
		{
			name: "function with var parameter",
			fn: &FunctionDecl{
				Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
				Name:  &Identifier{Value: "Process"},
				Parameters: []*Parameter{
					{
						Name:  &Identifier{Value: "data"},
						Type:  &TypeAnnotation{Name: "String"},
						ByRef: true,
					},
				},
				ReturnType: &TypeAnnotation{Name: "Boolean"},
				Body: &BlockStatement{
					Statements: []Statement{},
				},
			},
			expected: "function Process(var data: String): Boolean begin\nend",
		},
		{
			name: "function with lazy parameter",
			fn: &FunctionDecl{
				Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
				Name:  &Identifier{Value: "Compute"},
				Parameters: []*Parameter{
					{
						Name:   &Identifier{Value: "expr"},
						Type:   &TypeAnnotation{Name: "Integer"},
						IsLazy: true,
					},
				},
				ReturnType: &TypeAnnotation{Name: "Integer"},
				Body: &BlockStatement{
					Statements: []Statement{},
				},
			},
			expected: "function Compute(lazy expr: Integer): Integer begin\nend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn.String()
			if result != tt.expected {
				t.Errorf("FunctionDecl.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestReturnStatementString(t *testing.T) {
	tests := []struct {
		name     string
		stmt     *ReturnStatement
		expected string
	}{
		{
			name: "return with integer value",
			stmt: &ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"},
				ReturnValue: &IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "42"},
					Value: 42,
				},
			},
			expected: "Result := 42",
		},
		{
			name: "return with expression",
			stmt: &ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"},
				ReturnValue: &BinaryExpression{
					Left:     &Identifier{Value: "a"},
					Operator: "+",
					Right:    &Identifier{Value: "b"},
				},
			},
			expected: "Result := (a + b)",
		},
		{
			name: "return with string value",
			stmt: &ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"},
				ReturnValue: &StringLiteral{
					Token: lexer.Token{Type: lexer.STRING, Literal: "hello"},
					Value: "hello",
				},
			},
			expected: "Result := \"hello\"",
		},
		{
			name: "exit without value",
			stmt: &ReturnStatement{
				Token:       lexer.Token{Type: lexer.EXIT, Literal: "exit"},
				ReturnValue: nil,
			},
			expected: "exit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stmt.String()
			if result != tt.expected {
				t.Errorf("ReturnStatement.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
