package printer_test

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestUnaryExpression tests printing of unary expressions
func TestUnaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ast.UnaryExpression
		expected string
	}{
		{
			name: "unary minus",
			expr: &ast.UnaryExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.MINUS, Literal: "-"},
					},
				},
				Operator: "-",
				Right: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "42"},
						},
					},
					Value: 42,
				},
			},
			expected: "-42",
		},
		{
			name: "not operator",
			expr: &ast.UnaryExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.NOT, Literal: "not"},
					},
				},
				Operator: "not",
				Right: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.TRUE, Literal: "true"},
						},
					},
					Value: true,
				},
			},
			expected: "nottrue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.CompactOptions())
			result := p.Print(tt.expr)
			if result != tt.expected {
				t.Errorf("Print() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestRangeExpression tests printing of range expressions
func TestRangeExpression(t *testing.T) {
	expr := &ast.RangeExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.DOTDOT, Literal: ".."},
			},
		},
		Start: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "1"},
				},
			},
			Value: 1,
		},
		RangeEnd: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "10"},
				},
			},
			Value: 10,
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	expected := "1..10"
	if result != expected {
		t.Errorf("Print() = %q, want %q", result, expected)
	}
}

// TestCallExpression tests printing of function call expressions
func TestCallExpression(t *testing.T) {
	expr := &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.LPAREN, Literal: "("},
			},
		},
		Function: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "PrintLn"},
				},
			},
			Value: "PrintLn",
		},
		Arguments: []ast.Expression{
			&ast.StringLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.STRING, Literal: "'Hello'"},
					},
				},
				Value: "Hello",
			},
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	expected := "PrintLn(\"Hello\")"
	if result != expected {
		t.Errorf("Print() = %q, want %q", result, expected)
	}
}

// TestArrayLiteral tests printing of array literal expressions
func TestArrayLiteral(t *testing.T) {
	expr := &ast.ArrayLiteralExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.LBRACK, Literal: "["},
			},
		},
		Elements: []ast.Expression{
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.INT, Literal: "1"},
					},
				},
				Value: 1,
			},
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.INT, Literal: "2"},
					},
				},
				Value: 2,
			},
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.INT, Literal: "3"},
					},
				},
				Value: 3,
			},
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	expected := "[1,2,3]"
	if result != expected {
		t.Errorf("Print() = %q, want %q", result, expected)
	}
}

// TestIndexExpression tests printing of array index expressions
func TestIndexExpression(t *testing.T) {
	expr := &ast.IndexExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.LBRACK, Literal: "["},
			},
		},
		Left: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "arr"},
				},
			},
			Value: "arr",
		},
		Index: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "0"},
				},
			},
			Value: 0,
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	expected := "arr[0]"
	if result != expected {
		t.Errorf("Print() = %q, want %q", result, expected)
	}
}

// TestNewArrayExpression tests printing of new array expressions
func TestNewArrayExpression(t *testing.T) {
	expr := &ast.NewArrayExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.NEW, Literal: "new"},
			},
		},
		ElementTypeName: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "Integer"},
				},
			},
			Value: "Integer",
		},
		Dimensions: []ast.Expression{
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.INT, Literal: "10"},
					},
				},
				Value: 10,
			},
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	// Should contain "new" and "Integer"
	if !strings.Contains(result, "new") {
		t.Errorf("Print() = %q, expected to contain 'new'", result)
	}
}

// TestSetLiteral tests printing of set literal expressions
func TestSetLiteral(t *testing.T) {
	expr := &ast.SetLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.LBRACK, Literal: "["},
			},
		},
		Elements: []ast.Expression{
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.INT, Literal: "1"},
					},
				},
				Value: 1,
			},
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.INT, Literal: "2"},
					},
				},
				Value: 2,
			},
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	expected := "[1,2]"
	if result != expected {
		t.Errorf("Print() = %q, want %q", result, expected)
	}
}

// TestNewExpression tests printing of new object expressions
func TestNewExpression(t *testing.T) {
	expr := &ast.NewExpression{
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
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	if !strings.Contains(result, "TObject") {
		t.Errorf("Print() = %q, expected to contain 'TObject'", result)
	}
}

// TestMemberAccessExpression tests printing of member access
func TestMemberAccessExpression(t *testing.T) {
	expr := &ast.MemberAccessExpression{
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
		Member: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "field"},
				},
			},
			Value: "field",
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	expected := "obj.field"
	if result != expected {
		t.Errorf("Print() = %q, want %q", result, expected)
	}
}

// TestMethodCallExpression tests printing of method calls
func TestMethodCallExpression(t *testing.T) {
	expr := &ast.MethodCallExpression{
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
					Token: token.Token{Type: token.IDENT, Literal: "DoSomething"},
				},
			},
			Value: "DoSomething",
		},
		Arguments: []ast.Expression{},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	expected := "obj.DoSomething()"
	if result != expected {
		t.Errorf("Print() = %q, want %q", result, expected)
	}
}

// TestIsExpression tests printing of 'is' expressions
func TestIsExpression(t *testing.T) {
	expr := &ast.IsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.IS, Literal: "is"},
			},
		},
		Left: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "obj"},
				},
			},
			Value: "obj",
		},
		TargetType: &ast.TypeAnnotation{
			Token: token.Token{Type: token.IDENT, Literal: "TObject"},
			Name:  "TObject",
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	if !strings.Contains(result, "is") || !strings.Contains(result, "TObject") {
		t.Errorf("Print() = %q, expected to contain 'is' and 'TObject'", result)
	}
}

// TestAsExpression tests printing of 'as' expressions
func TestAsExpression(t *testing.T) {
	expr := &ast.AsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.AS, Literal: "as"},
			},
		},
		Left: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "obj"},
				},
			},
			Value: "obj",
		},
		TargetType: &ast.TypeAnnotation{
			Token: token.Token{Type: token.IDENT, Literal: "TMyClass"},
			Name:  "TMyClass",
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	if !strings.Contains(result, "as") || !strings.Contains(result, "TMyClass") {
		t.Errorf("Print() = %q, expected to contain 'as' and 'TMyClass'", result)
	}
}

// TestImplementsExpression tests printing of 'implements' expressions
func TestImplementsExpression(t *testing.T) {
	expr := &ast.ImplementsExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.IMPLEMENTS, Literal: "implements"},
			},
		},
		Left: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "obj"},
				},
			},
			Value: "obj",
		},
		TargetType: &ast.TypeAnnotation{
			Token: token.Token{Type: token.IDENT, Literal: "IInterface"},
			Name:  "IInterface",
		},
	}

	p := printer.New(printer.CompactOptions())
	result := p.Print(expr)
	if !strings.Contains(result, "implements") || !strings.Contains(result, "IInterface") {
		t.Errorf("Print() = %q, expected to contain 'implements' and 'IInterface'", result)
	}
}
