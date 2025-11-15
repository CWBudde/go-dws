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
			name:     "simple parameter",
			param:    NewTestParameter("x", "Integer", false),
			expected: "x: Integer",
		},
		{
			name:     "var parameter (by reference)",
			param:    NewTestParameter("s", "String", true),
			expected: "var s: String",
		},
		{
			name:     "float parameter",
			param:    NewTestParameter("x", "Float", false),
			expected: "x: Float",
		},
		{
			name: "lazy parameter",
			param: &Parameter{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "expr"},
				Name:   NewTestIdentifier("expr"),
				Type:   NewTestTypeAnnotation("Integer"),
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
			name:     "simple function with no parameters",
			fn:       NewTestFunctionDecl("GetValue", []*Parameter{}, NewTestTypeAnnotation("Integer")),
			expected: "function GetValue(): Integer begin\nend",
		},
		{
			name: "function with single parameter",
			fn: NewTestFunctionDecl(
				"Double",
				[]*Parameter{NewTestParameter("x", "Integer", false)},
				NewTestTypeAnnotation("Integer"),
			),
			expected: "function Double(x: Integer): Integer begin\nend",
		},
		{
			name: "function with multiple parameters",
			fn: NewTestFunctionDecl(
				"Add",
				[]*Parameter{
					NewTestParameter("a", "Integer", false),
					NewTestParameter("b", "Integer", false),
				},
				NewTestTypeAnnotation("Integer"),
			),
			expected: "function Add(a: Integer; b: Integer): Integer begin\nend",
		},
		{
			name: "procedure with no return type",
			fn: &FunctionDecl{
									BaseNode: BaseNode{Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
				Name:       NewTestIdentifier("Hello"),
				Parameters: []*Parameter{},
				ReturnType: nil,
				Body:       NewTestBlockStatement([]Statement{}),
			},
			expected: "procedure Hello begin\nend",
		},
		{
			name: "function with var parameter",
			fn: NewTestFunctionDecl(
				"Process",
				[]*Parameter{NewTestParameter("data", "String", true)},
				NewTestTypeAnnotation("Boolean"),
			),
			expected: "function Process(var data: String): Boolean begin\nend",
		},
		{
			name: "function with lazy parameter",
			fn: &FunctionDecl{
									BaseNode: BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
				Name: NewTestIdentifier("Compute"),
				Parameters: []*Parameter{
					{
						Name:   NewTestIdentifier("expr"),
						Type:   NewTestTypeAnnotation("Integer"),
						IsLazy: true,
					},
				},
				ReturnType: NewTestTypeAnnotation("Integer"),
				Body:       NewTestBlockStatement([]Statement{}),
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
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"},
				ReturnValue: NewTestIntegerLiteral(42),
			},
			expected: "Result := 42",
		},
		{
			name: "return with expression",
			stmt: &ReturnStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"},
				ReturnValue: NewTestBinaryExpression(NewTestIdentifier("a"), "+", NewTestIdentifier("b")),
			},
			expected: "Result := (a + b)",
		},
		{
			name: "return with string value",
			stmt: &ReturnStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"},
				ReturnValue: NewTestStringLiteral("hello"),
			},
			expected: "Result := \"hello\"",
		},
		{
			name: "exit without value",
			stmt: &ReturnStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"},
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
