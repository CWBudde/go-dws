package printer_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestPrintRecordLiteral tests record literal printing
func TestPrintRecordLiteral(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.RecordLiteralExpression
		expected string
	}{
		{
			name: "record with named fields",
			node: &ast.RecordLiteralExpression{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "TPoint"},
				},
				TypeName: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TPoint"},
						},
					},
					Value: "TPoint",
				},
				Fields: []*ast.FieldInitializer{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "X"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "X"},
								},
							},
							Value: "X",
						},
						Value: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "10"},
								},
							},
							Value: 10,
						},
					},
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Y"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Y"},
								},
							},
							Value: "Y",
						},
						Value: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "20"},
								},
							},
							Value: 20,
						},
					},
				},
			},
			expected: "TPoint.Create(X:10,Y:20)",
		},
		{
			name: "record with positional fields",
			node: &ast.RecordLiteralExpression{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.LPAREN, Literal: "("},
				},
				Fields: []*ast.FieldInitializer{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "5"},
						},
						Value: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "5"},
								},
							},
							Value: 5,
						},
					},
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "15"},
						},
						Value: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "15"},
								},
							},
							Value: 15,
						},
					},
				},
			},
			expected: "(5,15)",
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

// TestPrintLambdaExpression tests lambda expression printing
func TestPrintLambdaExpression(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.LambdaExpression
		expected string
	}{
		{
			name: "lambda without parameters",
			node: &ast.LambdaExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.LAMBDA, Literal: "lambda"},
					},
				},
				Parameters: []*ast.Parameter{},
				Body: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.INT, Literal: "42"},
							},
							Expression: &ast.IntegerLiteral{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.INT, Literal: "42"},
									},
								},
								Value: 42,
							},
						},
					},
				},
			},
			expected: "lambdabegin42;end",
		},
		{
			name: "lambda with single parameter",
			node: &ast.LambdaExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.LAMBDA, Literal: "lambda"},
					},
				},
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "x"},
								},
							},
							Value: "x",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
				Body: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.PLUS, Literal: "+"},
							},
							Expression: &ast.BinaryExpression{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.PLUS, Literal: "+"},
									},
								},
								Left: &ast.Identifier{
									TypedExpressionBase: ast.TypedExpressionBase{
										BaseNode: ast.BaseNode{
											Token: token.Token{Type: token.IDENT, Literal: "x"},
										},
									},
									Value: "x",
								},
								Operator: "+",
								Right: &ast.IntegerLiteral{
									TypedExpressionBase: ast.TypedExpressionBase{
										BaseNode: ast.BaseNode{
											Token: token.Token{Type: token.INT, Literal: "1"},
										},
									},
									Value: 1,
								},
							},
						},
					},
				},
			},
			expected: "lambda(x:Integer)beginx + 1;end",
		},
		{
			name: "lambda with multiple parameters",
			node: &ast.LambdaExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.LAMBDA, Literal: "lambda"},
					},
				},
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "x"},
								},
							},
							Value: "x",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
					{
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "y"},
								},
							},
							Value: "y",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
				Body: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.PLUS, Literal: "+"},
							},
							Expression: &ast.BinaryExpression{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.PLUS, Literal: "+"},
									},
								},
								Left: &ast.Identifier{
									TypedExpressionBase: ast.TypedExpressionBase{
										BaseNode: ast.BaseNode{
											Token: token.Token{Type: token.IDENT, Literal: "x"},
										},
									},
									Value: "x",
								},
								Operator: "+",
								Right: &ast.Identifier{
									TypedExpressionBase: ast.TypedExpressionBase{
										BaseNode: ast.BaseNode{
											Token: token.Token{Type: token.IDENT, Literal: "y"},
										},
									},
									Value: "y",
								},
							},
						},
					},
				},
			},
			expected: "lambda(x:Integer;y:Integer)beginx + y;end",
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

// TestPrintIfExpression tests if expression (ternary) printing
func TestPrintIfExpression(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.IfExpression
		expected string
	}{
		{
			name: "simple if expression",
			node: &ast.IfExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.IF, Literal: "if"},
					},
				},
				Condition: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.TRUE, Literal: "true"},
						},
					},
					Value: true,
				},
				Consequence: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "1"},
						},
					},
					Value: 1,
				},
				Alternative: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "0"},
						},
					},
					Value: 0,
				},
			},
			expected: "iftruethen1else0",
		},
		{
			name: "if expression without alternative",
			node: &ast.IfExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.IF, Literal: "if"},
					},
				},
				Condition: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.TRUE, Literal: "true"},
						},
					},
					Value: true,
				},
				Consequence: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "42"},
						},
					},
					Value: 42,
				},
			},
			expected: "iftruethen42",
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

// TestPrintEnumLiteral tests enum literal printing
func TestPrintEnumLiteral(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.EnumLiteral
		expected string
	}{
		{
			name: "simple enum value",
			node: &ast.EnumLiteral{
				Token:     token.Token{Type: token.IDENT, Literal: "Red"},
				ValueName: "Red",
			},
			expected: "Red",
		},
		{
			name: "qualified enum value",
			node: &ast.EnumLiteral{
				Token:     token.Token{Type: token.IDENT, Literal: "TColor"},
				EnumName:  "TColor",
				ValueName: "Red",
			},
			expected: "TColor.Red",
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

// TestPrintOldExpression tests old expression printing
func TestPrintOldExpression(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.OldExpression
		expected string
	}{
		{
			name: "old expression",
			node: &ast.OldExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.OLD, Literal: "old"},
					},
				},
				Identifier: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "value"},
						},
					},
					Value: "value",
				},
			},
			expected: "oldvalue",
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

// TestPrintNewExpressionWithArguments tests new expression with constructor arguments
func TestPrintNewExpressionWithArguments(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.NewExpression
		expected string
	}{
		{
			name: "new with no arguments",
			node: &ast.NewExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.NEW, Literal: "new"},
					},
				},
				ClassName: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TObject"},
						},
					},
					Value: "TObject",
				},
				Arguments: []ast.Expression{},
			},
			expected: "newTObject",
		},
		{
			name: "new with arguments",
			node: &ast.NewExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.NEW, Literal: "new"},
					},
				},
				ClassName: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TPoint"},
						},
					},
					Value: "TPoint",
				},
				Arguments: []ast.Expression{
					&ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.INT, Literal: "10"},
							},
						},
						Value: 10,
					},
					&ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.INT, Literal: "20"},
							},
						},
						Value: 20,
					},
				},
			},
			expected: "newTPoint(10,20)",
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

// TestPrintMethodCallExpressionEmptyArgs tests method call with no arguments
func TestPrintMethodCallExpressionEmptyArgs(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.MethodCallExpression
		expected string
	}{
		{
			name: "method call with no arguments",
			node: &ast.MethodCallExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.DOT, Literal: "."},
					},
				},
				Object: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "obj"},
						},
					},
					Value: "obj",
				},
				Method: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "ToString"},
						},
					},
					Value: "ToString",
				},
				Arguments: []ast.Expression{},
			},
			expected: "obj.ToString()",
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
