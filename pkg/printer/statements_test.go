package printer_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/cwbudde/go-dws/pkg/token"
)

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
