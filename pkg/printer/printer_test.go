package printer_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestReturnStatement tests that ReturnStatement uses Token.Literal correctly
func TestReturnStatement(t *testing.T) {
	tests := []struct {
		name        string
		stmt        *ast.ReturnStatement
		expected    string
		description string
	}{
		{
			name: "result assignment",
			stmt: &ast.ReturnStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "result"},
				},
				ReturnValue: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "42"},
						},
					},
					Value: 42,
				},
			},
			expected:    "result := 42",
			description: "Should use 'result' from token literal",
		},
		{
			name: "function name assignment",
			stmt: &ast.ReturnStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "Add"},
				},
				ReturnValue: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "10"},
						},
					},
					Value: 10,
				},
			},
			expected:    "Add := 10",
			description: "Should use function name from token literal",
		},
		{
			name: "exit without value",
			stmt: &ast.ReturnStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.EXIT, Literal: "exit"},
				},
				ReturnValue: nil,
			},
			expected:    "exit",
			description: "Should handle exit without assignment",
		},
		{
			name: "Result with uppercase",
			stmt: &ast.ReturnStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "Result"},
				},
				ReturnValue: &ast.StringLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.STRING, Literal: "\"hello\""},
						},
					},
					Value: "hello",
				},
			},
			expected:    "Result := \"hello\"",
			description: "Should preserve case from token literal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.stmt)
			if result != tt.expected {
				t.Errorf("%s\nExpected: %q\nGot:      %q", tt.description, tt.expected, result)
			}
		})
	}
}

// TestExitStatement tests that ExitStatement handles optional return values
func TestExitStatement(t *testing.T) {
	tests := []struct {
		name     string
		stmt     *ast.ExitStatement
		expected string
	}{
		{
			name: "exit without value",
			stmt: &ast.ExitStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.EXIT, Literal: "exit"},
				},
				ReturnValue: nil,
			},
			expected: "exit",
		},
		{
			name: "exit with integer value",
			stmt: &ast.ExitStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.EXIT, Literal: "exit"},
				},
				ReturnValue: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "42"},
						},
					},
					Value: 42,
				},
			},
			expected: "exit 42",
		},
		{
			name: "exit with expression",
			stmt: &ast.ExitStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.EXIT, Literal: "exit"},
				},
				ReturnValue: &ast.BinaryExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.PLUS, Literal: "+"},
						},
					},
					Left: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.INT, Literal: "1"},
							},
						},
						Value: 1,
					},
					Operator: "+",
					Right: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.INT, Literal: "2"},
							},
						},
						Value: 2,
					},
				},
			},
			expected: "exit 1 + 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.stmt)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestFunctionModifiers tests that all function modifiers are printed correctly
func TestFunctionModifiers(t *testing.T) {
	tests := []struct {
		name     string
		fn       *ast.FunctionDecl
		expected string
	}{
		{
			name: "virtual method",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "DoWork"},
						},
					},
					Value: "DoWork",
				},
				ReturnType: &ast.TypeAnnotation{
					Name: "Integer",
				},
				IsVirtual: true,
			},
			expected: "function DoWork: Integer; virtual",
		},
		{
			name: "override method",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "ToString"},
						},
					},
					Value: "ToString",
				},
				ReturnType: &ast.TypeAnnotation{
					Name: "String",
				},
				IsOverride: true,
			},
			expected: "function ToString: String; override",
		},
		{
			name: "abstract method",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "GetValue"},
						},
					},
					Value: "GetValue",
				},
				ReturnType: &ast.TypeAnnotation{
					Name: "Float",
				},
				IsAbstract: true,
			},
			expected: "function GetValue: Float; abstract",
		},
		{
			name: "class method",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "GetCount"},
						},
					},
					Value: "GetCount",
				},
				ReturnType: &ast.TypeAnnotation{
					Name: "Integer",
				},
				IsClassMethod: true,
			},
			expected: "class function GetCount: Integer",
		},
		{
			name: "overload method",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.PROCEDURE, Literal: "procedure"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Write"},
						},
					},
					Value: "Write",
				},
				IsOverload: true,
			},
			expected: "procedure Write; overload",
		},
		{
			name: "deprecated method with message",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "OldFunc"},
						},
					},
					Value: "OldFunc",
				},
				ReturnType: &ast.TypeAnnotation{
					Name: "Boolean",
				},
				IsDeprecated:      true,
				DeprecatedMessage: "Use NewFunc instead",
			},
			expected: "function OldFunc: Boolean; deprecated 'Use NewFunc instead'",
		},
		{
			name: "calling convention",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "WinAPI"},
						},
					},
					Value: "WinAPI",
				},
				ReturnType: &ast.TypeAnnotation{
					Name: "Integer",
				},
				CallingConvention: "stdcall",
			},
			expected: "function WinAPI: Integer; stdcall",
		},
		{
			name: "constructor",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CONSTRUCTOR, Literal: "constructor"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Create"},
						},
					},
					Value: "Create",
				},
				IsConstructor: true,
			},
			expected: "constructor Create",
		},
		{
			name: "destructor",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.DESTRUCTOR, Literal: "destructor"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Destroy"},
						},
					},
					Value: "Destroy",
				},
				IsDestructor: true,
			},
			expected: "destructor Destroy",
		},
		{
			name: "private visibility",
			fn: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Helper"},
						},
					},
					Value: "Helper",
				},
				ClassName: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TMyClass"},
						},
					},
					Value: "TMyClass",
				},
				ReturnType: &ast.TypeAnnotation{
					Name: "String",
				},
				Visibility: ast.VisibilityPrivate,
			},
			expected: "private function Helper: String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.fn)
			if result != tt.expected {
				t.Errorf("Expected: %q\nGot:      %q", tt.expected, result)
			}
		})
	}
}

// TestClassDeclaration tests that class constants and operators are printed
func TestClassDeclaration(t *testing.T) {
	// Create a class with constants and operators
	classDecl := &ast.ClassDecl{
		BaseNode: ast.BaseNode{
			Token: token.Token{Type: token.CLASS, Literal: "class"},
		},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "TMyClass"},
				},
			},
			Value: "TMyClass",
		},
		Constants: []*ast.ConstDecl{
			{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CONST, Literal: "const"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "MaxSize"},
						},
					},
					Value: "MaxSize",
				},
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "100"},
						},
					},
					Value: 100,
				},
			},
		},
		Operators: []*ast.OperatorDecl{
			{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.OPERATOR, Literal: "operator"},
				},
				OperatorSymbol: "+",
				OperandTypes: []ast.TypeExpression{
					&ast.TypeAnnotation{Name: "TMyClass"},
					&ast.TypeAnnotation{Name: "TMyClass"},
				},
				ReturnType: &ast.TypeAnnotation{
					Name: "TMyClass",
				},
			},
		},
	}

	p := printer.New(printer.Options{
		Format:      printer.FormatDWScript,
		Style:       printer.StyleDetailed,
		IndentWidth: 2,
	})
	result := p.Print(classDecl)

	// Check that the output contains the constant
	if !contains(result, "const MaxSize = 100") {
		t.Errorf("Expected class to contain constant 'const MaxSize = 100', got:\n%s", result)
	}

	// Check that the output contains the operator
	if !contains(result, "operator +") {
		t.Errorf("Expected class to contain operator '+', got:\n%s", result)
	}
}

// TestOutputFormats tests different output formats
func TestOutputFormats(t *testing.T) {
	// Create a simple expression
	expr := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.PLUS, Literal: "+"},
			},
		},
		Left: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "3"},
				},
			},
			Value: 3,
		},
		Operator: "+",
		Right: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "5"},
				},
			},
			Value: 5,
		},
	}

	tests := []struct {
		name     string
		contains string
		format   printer.Format
	}{
		{
			name:     "DWScript format",
			format:   printer.FormatDWScript,
			contains: "3 + 5",
		},
		{
			name:     "Tree format",
			format:   printer.FormatTree,
			contains: "BinaryExpression",
		},
		{
			name:     "JSON format",
			format:   printer.FormatJSON,
			contains: `"operator"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: tt.format,
				Style:  printer.StyleDetailed,
			})
			result := p.Print(expr)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got:\n%s", tt.contains, result)
			}
		})
	}
}

// TestEdgeCases tests edge cases and nil handling
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		node     ast.Node
		expected string
	}{
		{
			name:     "nil node",
			node:     nil,
			expected: "",
		},
		{
			name: "empty program",
			node: &ast.Program{
				Statements: []ast.Statement{},
			},
			expected: "",
		},
		{
			name: "nil literal",
			node: &ast.NilLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.NIL, Literal: "nil"},
					},
				},
			},
			expected: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
