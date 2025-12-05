package ast

import (
	"testing"
)

func TestExtractIntegerLiteral(t *testing.T) {
	tests := []struct {
		expr     Expression
		name     string
		expected int64
		ok       bool
	}{
		{
			name:     "plain positive integer",
			expr:     &IntegerLiteral{Value: 42},
			expected: 42,
			ok:       true,
		},
		{
			name:     "plain zero",
			expr:     &IntegerLiteral{Value: 0},
			expected: 0,
			ok:       true,
		},
		{
			name:     "plain large integer",
			expr:     &IntegerLiteral{Value: 999999},
			expected: 999999,
			ok:       true,
		},
		{
			name: "unary minus with integer literal",
			expr: &UnaryExpression{
				Operator: "-",
				Right:    &IntegerLiteral{Value: 5},
			},
			expected: -5,
			ok:       true,
		},
		{
			name: "unary minus with zero",
			expr: &UnaryExpression{
				Operator: "-",
				Right:    &IntegerLiteral{Value: 0},
			},
			expected: 0,
			ok:       true,
		},
		{
			name: "unary minus with large integer",
			expr: &UnaryExpression{
				Operator: "-",
				Right:    &IntegerLiteral{Value: 12345},
			},
			expected: -12345,
			ok:       true,
		},
		{
			name: "unary plus with integer literal (not supported)",
			expr: &UnaryExpression{
				Operator: "+",
				Right:    &IntegerLiteral{Value: 5},
			},
			expected: 0,
			ok:       false,
		},
		{
			name: "unary not with integer literal (not supported)",
			expr: &UnaryExpression{
				Operator: "not",
				Right:    &IntegerLiteral{Value: 5},
			},
			expected: 0,
			ok:       false,
		},
		{
			name: "unary minus with non-integer expression",
			expr: &UnaryExpression{
				Operator: "-",
				Right:    &Identifier{Value: "x"},
			},
			expected: 0,
			ok:       false,
		},
		{
			name:     "string literal (not supported)",
			expr:     &StringLiteral{Value: "42"},
			expected: 0,
			ok:       false,
		},
		{
			name:     "float literal (not supported)",
			expr:     &FloatLiteral{Value: 42.5},
			expected: 0,
			ok:       false,
		},
		{
			name:     "boolean literal (not supported)",
			expr:     &BooleanLiteral{Value: true},
			expected: 0,
			ok:       false,
		},
		{
			name:     "identifier (not supported)",
			expr:     &Identifier{Value: "someVar"},
			expected: 0,
			ok:       false,
		},
		{
			name: "binary expression (not supported)",
			expr: &BinaryExpression{
				Left:     &IntegerLiteral{Value: 2},
				Operator: "+",
				Right:    &IntegerLiteral{Value: 3},
			},
			expected: 0,
			ok:       false,
		},
		{
			name: "call expression (not supported)",
			expr: &CallExpression{
				Function:  &Identifier{Value: "getValue"},
				Arguments: []Expression{},
			},
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ExtractIntegerLiteral(tt.expr)
			if ok != tt.ok {
				t.Errorf("ExtractIntegerLiteral() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.expected {
				t.Errorf("ExtractIntegerLiteral() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test edge case: nil expression
func TestExtractIntegerLiteral_Nil(t *testing.T) {
	// This should not panic
	got, ok := ExtractIntegerLiteral(nil)
	if ok {
		t.Errorf("ExtractIntegerLiteral(nil) ok = true, want false")
	}
	if got != 0 {
		t.Errorf("ExtractIntegerLiteral(nil) = %v, want 0", got)
	}
}
