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

// TestPrintAssignmentStatement tests assignment statement printing
func TestPrintAssignmentStatement(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.AssignmentStatement
		expected string
	}{
		{
			name: "simple assignment",
			node: &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.ASSIGN, Literal: ":="},
				},
				Target: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "x"},
						},
					},
					Value: "x",
				},
				Operator: token.ASSIGN,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "42"},
						},
					},
					Value: 42,
				},
			},
			expected: "x:=42",
		},
		{
			name: "plus assign",
			node: &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.PLUS_ASSIGN, Literal: "+="},
				},
				Target: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "count"},
						},
					},
					Value: "count",
				},
				Operator: token.PLUS_ASSIGN,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "1"},
						},
					},
					Value: 1,
				},
			},
			expected: "count+=1",
		},
		{
			name: "minus assign",
			node: &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.MINUS_ASSIGN, Literal: "-="},
				},
				Target: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "total"},
						},
					},
					Value: "total",
				},
				Operator: token.MINUS_ASSIGN,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "5"},
						},
					},
					Value: 5,
				},
			},
			expected: "total-=5",
		},
		{
			name: "times assign",
			node: &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TIMES_ASSIGN, Literal: "*="},
				},
				Target: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "n"},
						},
					},
					Value: "n",
				},
				Operator: token.TIMES_ASSIGN,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "2"},
						},
					},
					Value: 2,
				},
			},
			expected: "n*=2",
		},
		{
			name: "divide assign",
			node: &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.DIVIDE_ASSIGN, Literal: "/="},
				},
				Target: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "value"},
						},
					},
					Value: "value",
				},
				Operator: token.DIVIDE_ASSIGN,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "10"},
						},
					},
					Value: 10,
				},
			},
			expected: "value/=10",
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

// TestPrintWhileStatement tests while loop printing
func TestPrintWhileStatement(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.WhileStatement
		contains string
	}{
		{
			name: "simple while loop",
			node: &ast.WhileStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.WHILE, Literal: "while"},
				},
				Condition: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.TRUE, Literal: "true"},
						},
					},
					Value: true,
				},
				Body: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{},
				},
			},
			contains: "while true do",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleDetailed,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintRepeatStatement tests repeat-until loop printing
func TestPrintRepeatStatement(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.RepeatStatement
		contains string
	}{
		{
			name: "simple repeat loop",
			node: &ast.RepeatStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.REPEAT, Literal: "repeat"},
				},
				Body: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{},
				},
				Condition: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.FALSE, Literal: "false"},
						},
					},
					Value: false,
				},
			},
			contains: "repeat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleDetailed,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintForStatement tests for loop printing
func TestPrintForStatement(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ForStatement
		contains string
	}{
		{
			name: "for to loop",
			node: &ast.ForStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FOR, Literal: "for"},
				},
				Variable: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "i"},
						},
					},
					Value: "i",
				},
				Start: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "1"},
						},
					},
					Value: 1,
				},
				Direction: ast.ForTo,
				EndValue: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "10"},
						},
					},
					Value: 10,
				},
				Body: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{},
				},
			},
			contains: "for i := 1 to 10 do",
		},
		{
			name: "for downto loop",
			node: &ast.ForStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FOR, Literal: "for"},
				},
				Variable: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "i"},
						},
					},
					Value: "i",
				},
				Start: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "10"},
						},
					},
					Value: 10,
				},
				Direction: ast.ForDownto,
				EndValue: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "1"},
						},
					},
					Value: 1,
				},
				Body: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{},
				},
			},
			contains: "for i := 10 downto 1 do",
		},
		{
			name: "for loop with step",
			node: &ast.ForStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FOR, Literal: "for"},
				},
				Variable: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "i"},
						},
					},
					Value: "i",
				},
				Start: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "0"},
						},
					},
					Value: 0,
				},
				Direction: ast.ForTo,
				EndValue: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "100"},
						},
					},
					Value: 100,
				},
				Step: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "10"},
						},
					},
					Value: 10,
				},
				Body: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{},
				},
			},
			contains: "for i := 0 to 100 step 10 do",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleDetailed,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintForInStatement tests for-in loop printing
func TestPrintForInStatement(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ForInStatement
		contains string
	}{
		{
			name: "for in loop",
			node: &ast.ForInStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FOR, Literal: "for"},
				},
				Variable: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "item"},
						},
					},
					Value: "item",
				},
				Collection: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "items"},
						},
					},
					Value: "items",
				},
				Body: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{},
				},
			},
			contains: "for item in items do",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleDetailed,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintCaseStatement tests case statement printing
func TestPrintCaseStatement(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.CaseStatement
		contains string
	}{
		{
			name: "simple case statement",
			node: &ast.CaseStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CASE, Literal: "case"},
				},
				Expression: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "x"},
						},
					},
					Value: "x",
				},
				Cases: []*ast.CaseBranch{
					{
						Values: []ast.Expression{
							&ast.IntegerLiteral{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.INT, Literal: "1"},
									},
								},
								Value: 1,
							},
						},
						Statement: &ast.ExpressionStatement{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "a"},
							},
							Expression: &ast.Identifier{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.IDENT, Literal: "a"},
									},
								},
								Value: "a",
							},
						},
					},
					{
						Values: []ast.Expression{
							&ast.IntegerLiteral{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.INT, Literal: "2"},
									},
								},
								Value: 2,
							},
						},
						Statement: &ast.ExpressionStatement{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "b"},
							},
							Expression: &ast.Identifier{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.IDENT, Literal: "b"},
									},
								},
								Value: "b",
							},
						},
					},
				},
			},
			contains: "case x of",
		},
		{
			name: "case with else",
			node: &ast.CaseStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CASE, Literal: "case"},
				},
				Expression: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "x"},
						},
					},
					Value: "x",
				},
				Cases: []*ast.CaseBranch{
					{
						Values: []ast.Expression{
							&ast.IntegerLiteral{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.INT, Literal: "1"},
									},
								},
								Value: 1,
							},
						},
						Statement: &ast.ExpressionStatement{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "a"},
							},
							Expression: &ast.Identifier{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.IDENT, Literal: "a"},
									},
								},
								Value: "a",
							},
						},
					},
				},
				Else: &ast.ExpressionStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.IDENT, Literal: "default"},
					},
					Expression: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "default"},
							},
						},
						Value: "default",
					},
				},
			},
			contains: "else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleDetailed,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintTryStatement tests exception handling printing
func TestPrintTryStatement(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.TryStatement
		contains string
	}{
		{
			name: "try-except",
			node: &ast.TryStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TRY, Literal: "try"},
				},
				TryBlock: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{},
				},
				ExceptClause: &ast.ExceptClause{
					Handlers: []*ast.ExceptionHandler{
						{
							Variable: &ast.Identifier{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.IDENT, Literal: "e"},
									},
								},
								Value: "e",
							},
							ExceptionType: &ast.TypeAnnotation{Name: "Exception"},
							Statement: &ast.ExpressionStatement{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "handle"},
								},
								Expression: &ast.Identifier{
									TypedExpressionBase: ast.TypedExpressionBase{
										BaseNode: ast.BaseNode{
											Token: token.Token{Type: token.IDENT, Literal: "handle"},
										},
									},
									Value: "handle",
								},
							},
						},
					},
				},
			},
			contains: "try",
		},
		{
			name: "try-finally",
			node: &ast.TryStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TRY, Literal: "try"},
				},
				TryBlock: &ast.BlockStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.BEGIN, Literal: "begin"},
					},
					Statements: []ast.Statement{},
				},
				FinallyClause: &ast.FinallyClause{
					Block: &ast.BlockStatement{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.BEGIN, Literal: "begin"},
						},
						Statements: []ast.Statement{},
					},
				},
			},
			contains: "finally",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleDetailed,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintRaiseStatement tests raise statement printing
func TestPrintRaiseStatement(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.RaiseStatement
		expected string
	}{
		{
			name: "raise without exception",
			node: &ast.RaiseStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.RAISE, Literal: "raise"},
				},
			},
			expected: "raise",
		},
		{
			name: "raise with exception",
			node: &ast.RaiseStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.RAISE, Literal: "raise"},
				},
				Exception: &ast.NewExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.NEW, Literal: "new"},
						},
					},
					ClassName: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "Exception"},
							},
						},
						Value: "Exception",
					},
				},
			},
			expected: "raisenewException",
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

// TestPrintParameter tests parameter printing
func TestPrintParameter(t *testing.T) {
	// Parameters are tested indirectly through function declarations
	// This test ensures different parameter types work correctly
	tests := []struct {
		name     string
		node     *ast.FunctionDecl
		contains string
	}{
		{
			name: "function with const parameter",
			node: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Test"},
						},
					},
					Value: "Test",
				},
				Parameters: []*ast.Parameter{
					{
						IsConst: true,
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
				ReturnType: &ast.TypeAnnotation{Name: "Integer"},
			},
			contains: "constx:Integer",
		},
		{
			name: "function with var parameter",
			node: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Test"},
						},
					},
					Value: "Test",
				},
				Parameters: []*ast.Parameter{
					{
						ByRef: true,
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "y"},
								},
							},
							Value: "y",
						},
						Type: &ast.TypeAnnotation{Name: "String"},
					},
				},
				ReturnType: &ast.TypeAnnotation{Name: "Boolean"},
			},
			contains: "vary:String",
		},
		{
			name: "function with default value",
			node: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Test"},
						},
					},
					Value: "Test",
				},
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "z"},
								},
							},
							Value: "z",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
						DefaultValue: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "42"},
								},
							},
							Value: 42,
						},
					},
				},
				ReturnType: &ast.TypeAnnotation{Name: "Integer"},
			},
			contains: "z:Integer=42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintFieldDecl tests field declaration printing
func TestPrintFieldDecl(t *testing.T) {
	// Field declarations are tested through class declarations
	tests := []struct {
		name     string
		node     *ast.ClassDecl
		contains string
	}{
		{
			name: "class with private field",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Fields: []*ast.FieldDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "fValue"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "fValue"},
								},
							},
							Value: "fValue",
						},
						Type:       &ast.TypeAnnotation{Name: "Integer"},
						Visibility: ast.VisibilityPrivate,
					},
				},
			},
			contains: "private fValue: Integer",
		},
		{
			name: "class with field with init value",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Fields: []*ast.FieldDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Count"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Count"},
								},
							},
							Value: "Count",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
						InitValue: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "0"},
								},
							},
							Value: 0,
						},
					},
				},
			},
			contains: "Count: Integer = 0",
		},
		{
			name: "class with field type inference",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Fields: []*ast.FieldDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Name"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Name"},
								},
							},
							Value: "Name",
						},
						InitValue: &ast.StringLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.STRING, Literal: "\"test\""},
								},
							},
							Value: "test",
						},
					},
				},
			},
			contains: "Name := \"test\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintPropertyDecl tests property declaration printing
func TestPrintPropertyDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ClassDecl
		contains string
	}{
		{
			name: "class with property",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Properties: []*ast.PropertyDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.PROPERTY, Literal: "property"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Value"},
								},
							},
							Value: "Value",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
						ReadSpec: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "fValue"},
								},
							},
							Value: "fValue",
						},
						WriteSpec: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "fValue"},
								},
							},
							Value: "fValue",
						},
					},
				},
			},
			contains: "property Value: Integer read fValue write fValue",
		},
		{
			name: "class with default property",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Properties: []*ast.PropertyDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.PROPERTY, Literal: "property"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Items"},
								},
							},
							Value: "Items",
						},
						Type: &ast.TypeAnnotation{Name: "String"},
						ReadSpec: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "GetItem"},
								},
							},
							Value: "GetItem",
						},
						IsDefault: true,
					},
				},
			},
			contains: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintRecordDecl tests record declaration printing
func TestPrintRecordDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.RecordDecl
		contains string
	}{
		{
			name: "simple record",
			node: &ast.RecordDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.RECORD, Literal: "record"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TPoint"},
						},
					},
					Value: "TPoint",
				},
				Fields: []*ast.FieldDecl{
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
						Type: &ast.TypeAnnotation{Name: "Integer"},
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
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
			},
			contains: "type TPoint = record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintEnumDecl tests enum declaration printing
func TestPrintEnumDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.EnumDecl
		expected string
	}{
		{
			name: "simple enum",
			node: &ast.EnumDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TColor"},
						},
					},
					Value: "TColor",
				},
				Values: []ast.EnumValue{
					{Name: "Red"},
					{Name: "Green"},
					{Name: "Blue"},
				},
			},
			expected: "typeTColor=(Red,Green,Blue)",
		},
		{
			name: "enum with explicit values",
			node: &ast.EnumDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TStatus"},
						},
					},
					Value: "TStatus",
				},
				Values: []ast.EnumValue{
					{Name: "Idle", Value: intPtr(0)},
					{Name: "Running", Value: intPtr(1)},
					{Name: "Stopped", Value: intPtr(2)},
				},
			},
			expected: "typeTStatus=(Idle=0,Running=1,Stopped=2)",
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

// TestPrintArrayDecl tests array declaration printing
func TestPrintArrayDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ArrayDecl
		contains string
	}{
		{
			name: "simple array declaration",
			node: &ast.ArrayDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TIntArray"},
						},
					},
					Value: "TIntArray",
				},
				ArrayType: &ast.ArrayTypeAnnotation{
					ElementType: &ast.TypeAnnotation{Name: "Integer"},
				},
			},
			contains: "typeTIntArray=array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintSetDecl tests set declaration printing
func TestPrintSetDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.SetDecl
		expected string
	}{
		{
			name: "simple set declaration",
			node: &ast.SetDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TCharSet"},
						},
					},
					Value: "TCharSet",
				},
				ElementType: &ast.TypeAnnotation{Name: "Char"},
			},
			expected: "typeTCharSet=setofChar",
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

// TestPrintInterfaceDecl tests interface declaration printing
func TestPrintInterfaceDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.InterfaceDecl
		contains string
	}{
		{
			name: "simple interface",
			node: &ast.InterfaceDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INTERFACE, Literal: "interface"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "IComparable"},
						},
					},
					Value: "IComparable",
				},
				Methods: []*ast.InterfaceMethodDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.FUNCTION, Literal: "function"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "CompareTo"},
								},
							},
							Value: "CompareTo",
						},
						Parameters: []*ast.Parameter{
							{
								Name: &ast.Identifier{
									TypedExpressionBase: ast.TypedExpressionBase{
										BaseNode: ast.BaseNode{
											Token: token.Token{Type: token.IDENT, Literal: "other"},
										},
									},
									Value: "other",
								},
								Type: &ast.TypeAnnotation{Name: "IComparable"},
							},
						},
						ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
			},
			contains: "type IComparable = interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintInterfaceMethodDecl tests interface method declaration printing
func TestPrintInterfaceMethodDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.InterfaceDecl
		contains string
	}{
		{
			name: "interface with procedure",
			node: &ast.InterfaceDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INTERFACE, Literal: "interface"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "ITest"},
						},
					},
					Value: "ITest",
				},
				Methods: []*ast.InterfaceMethodDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.PROCEDURE, Literal: "procedure"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "DoSomething"},
								},
							},
							Value: "DoSomething",
						},
					},
				},
			},
			contains: "procedure DoSomething",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintTypeDeclaration tests type alias declaration printing
func TestPrintTypeDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.TypeDeclaration
		expected string
	}{
		{
			name: "type alias",
			node: &ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TMyInt"},
						},
					},
					Value: "TMyInt",
				},
				AliasedType: &ast.TypeAnnotation{Name: "Integer"},
			},
			expected: "typeTMyInt=Integer",
		},
		{
			name: "subrange type",
			node: &ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TDigit"},
						},
					},
					Value: "TDigit",
				},
				IsSubrange: true,
				LowBound: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "0"},
						},
					},
					Value: 0,
				},
				HighBound: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "9"},
						},
					},
					Value: 9,
				},
			},
			expected: "typeTDigit=0..9",
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

// TestPrintUnitDeclaration tests unit declaration printing
func TestPrintUnitDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.UnitDeclaration
		contains string
	}{
		{
			name: "simple unit",
			node: &ast.UnitDeclaration{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.UNIT, Literal: "unit"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "MyUnit"},
						},
					},
					Value: "MyUnit",
				},
				InterfaceSection: &ast.BlockStatement{
					Statements: []ast.Statement{},
				},
				ImplementationSection: &ast.BlockStatement{
					Statements: []ast.Statement{},
				},
			},
			contains: "unit MyUnit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
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
				Token: token.Token{Type: token.IDENT, Literal: "Red"},
				ValueName: "Red",
			},
			expected: "Red",
		},
		{
			name: "qualified enum value",
			node: &ast.EnumLiteral{
				Token: token.Token{Type: token.IDENT, Literal: "TColor"},
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

// TestPrintHelperDecl tests helper declaration printing
func TestPrintHelperDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.HelperDecl
		contains string
	}{
		{
			name: "simple helper",
			node: &ast.HelperDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.HELPER, Literal: "helper"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TStringHelper"},
						},
					},
					Value: "TStringHelper",
				},
				ForType: &ast.TypeAnnotation{Name: "String"},
				Methods: []*ast.FunctionDecl{},
			},
			contains: "type TStringHelper = helper for String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintRecordPropertyDecl tests record property declaration printing
func TestPrintRecordPropertyDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.RecordPropertyDecl
		expected string
	}{
		{
			name: "record property with read/write",
			node: &ast.RecordPropertyDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.PROPERTY, Literal: "property"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Value"},
						},
					},
					Value: "Value",
				},
				Type:       &ast.TypeAnnotation{Name: "Integer"},
				ReadField:  "fValue",
				WriteField: "fValue",
			},
			expected: "propertyValue:IntegerreadfValuewritefValue",
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

// TestPrintArrayTypeAnnotation tests array type annotation printing
func TestPrintArrayTypeAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ArrayTypeAnnotation
		expected string
	}{
		{
			name: "dynamic array",
			node: &ast.ArrayTypeAnnotation{
				ElementType: &ast.TypeAnnotation{Name: "Integer"},
			},
			expected: "arrayofInteger",
		},
		{
			name: "static array",
			node: &ast.ArrayTypeAnnotation{
				LowBound: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "0"},
						},
					},
					Value: 0,
				},
				HighBound: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "9"},
						},
					},
					Value: 9,
				},
				ElementType: &ast.TypeAnnotation{Name: "String"},
			},
			expected: "array[0..9]ofString",
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

// TestPrintArrayTypeNode tests array type node printing
func TestPrintArrayTypeNode(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ArrayTypeNode
		expected string
	}{
		{
			name: "dynamic array type",
			node: &ast.ArrayTypeNode{
				ElementType: &ast.TypeAnnotation{Name: "Float"},
			},
			expected: "arrayofFloat",
		},
		{
			name: "indexed array type",
			node: &ast.ArrayTypeNode{
				IndexType: &ast.TypeAnnotation{Name: "TColor"},
				ElementType: &ast.TypeAnnotation{Name: "String"},
			},
			expected: "array[TColor]ofString",
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

// TestPrintSetTypeNode tests set type node printing
func TestPrintSetTypeNode(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.SetTypeNode
		expected string
	}{
		{
			name: "set type",
			node: &ast.SetTypeNode{
				ElementType: &ast.TypeAnnotation{Name: "Byte"},
			},
			expected: "setofByte",
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

// TestPrintClassOfTypeNode tests class-of type node printing
func TestPrintClassOfTypeNode(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ClassOfTypeNode
		expected string
	}{
		{
			name: "class of type",
			node: &ast.ClassOfTypeNode{
				ClassType: &ast.TypeAnnotation{Name: "TComponent"},
			},
			expected: "classofTComponent",
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

// TestPrintFunctionPointerTypeNode tests function pointer type node printing
func TestPrintFunctionPointerTypeNode(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.FunctionPointerTypeNode
		expected string
	}{
		{
			name: "function pointer",
			node: &ast.FunctionPointerTypeNode{
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
				ReturnType: &ast.TypeAnnotation{Name: "String"},
			},
			expected: "function(x:Integer):String",
		},
		{
			name: "procedure pointer",
			node: &ast.FunctionPointerTypeNode{
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "msg"},
								},
							},
							Value: "msg",
						},
						Type: &ast.TypeAnnotation{Name: "String"},
					},
				},
			},
			expected: "procedure(msg:String)",
		},
		{
			name: "method pointer (of object)",
			node: &ast.FunctionPointerTypeNode{
				Parameters: []*ast.Parameter{},
				ReturnType: &ast.TypeAnnotation{Name: "Boolean"},
				OfObject:   true,
			},
			expected: "function:Booleanof object",
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

// Helper function
func intPtr(i int) *int {
	return &i
}
