package ast

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestConditionString(t *testing.T) {
	tests := []struct {
		name     string
		cond     *Condition
		expected string
	}{
		{
			name: "simple condition without message",
			cond: &Condition{
				Test:     NewTestIdentifier("x"),
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
			},
			expected: "x",
		},
		{
			name: "condition with message",
			cond: &Condition{
				Test: &BinaryExpression{
					TypedExpressionBase: TypedExpressionBase{
						BaseNode: BaseNode{Token: lexer.Token{Type: lexer.GREATER, Literal: ">"}},
					},
					Left:     &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}}}, Value: "x"},
					Operator: ">",
					Right:    &IntegerLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.INT, Literal: "0"}}}, Value: 0},
				},
				Message:  NewTestStringLiteral("x must be positive", "'x must be positive'"),
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
			},
			expected: "(x > 0) : \"x must be positive\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cond.String()
			if result != tt.expected {
				t.Errorf("Condition.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPreConditionsString(t *testing.T) {
	precond := &PreConditions{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.REQUIRE, Literal: "require"}},
		Conditions: []*Condition{
			{
				Test:     NewTestIdentifier("x"),
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
			},
			{
				Test: &BinaryExpression{
					TypedExpressionBase: TypedExpressionBase{
						BaseNode: BaseNode{Token: lexer.Token{Type: lexer.NOT_EQ, Literal: "<>"}},
					},
					Left:     &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "y"}}}, Value: "y"},
					Operator: "<>",
					Right:    &IntegerLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.INT, Literal: "0"}}}, Value: 0},
				},
				Message:  NewTestStringLiteral("y cannot be zero", "'y cannot be zero'"),
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "y"}},
			},
		},
	}

	result := precond.String()
	expected := "require\n   x;\n   (y <> 0) : \"y cannot be zero\"\n"

	if result != expected {
		t.Errorf("PreConditions.String() =\n%q\nwant\n%q", result, expected)
	}
}

func TestPostConditionsString(t *testing.T) {
	postcond := &PostConditions{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.ENSURE, Literal: "ensure"}},
		Conditions: []*Condition{
			{
				Test: &BinaryExpression{
					TypedExpressionBase: TypedExpressionBase{
						BaseNode: BaseNode{Token: lexer.Token{Type: lexer.GREATER, Literal: ">"}},
					},
					Left:     &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"}}}, Value: "Result"},
					Operator: ">",
					Right:    &IntegerLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.INT, Literal: "0"}}}, Value: 0},
				},
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"}},
			},
		},
	}

	result := postcond.String()
	expected := "ensure\n   (Result > 0)\n"

	if result != expected {
		t.Errorf("PostConditions.String() =\n%q\nwant\n%q", result, expected)
	}
}

func TestOldExpressionString(t *testing.T) {
	oldExpr := &OldExpression{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.OLD, Literal: "old"}},
		},
		Identifier: NewTestIdentifier("x"),
	}

	result := oldExpr.String()
	expected := "old x"

	if result != expected {
		t.Errorf("OldExpression.String() = %q, want %q", result, expected)
	}
}

func TestConditionTokenLiteral(t *testing.T) {
	cond := &Condition{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
		Test:     NewTestIdentifier("x"),
	}

	if cond.TokenLiteral() != "x" {
		t.Errorf("Condition.TokenLiteral() = %q, want %q", cond.TokenLiteral(), "x")
	}
}

func TestPreConditionsTokenLiteral(t *testing.T) {
	precond := &PreConditions{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.REQUIRE, Literal: "require"}},
	}

	if precond.TokenLiteral() != "require" {
		t.Errorf("PreConditions.TokenLiteral() = %q, want %q", precond.TokenLiteral(), "require")
	}
}

func TestPostConditionsTokenLiteral(t *testing.T) {
	postcond := &PostConditions{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.ENSURE, Literal: "ensure"}},
	}

	if postcond.TokenLiteral() != "ensure" {
		t.Errorf("PostConditions.TokenLiteral() = %q, want %q", postcond.TokenLiteral(), "ensure")
	}
}

func TestOldExpressionTokenLiteral(t *testing.T) {
	oldExpr := &OldExpression{
		TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.OLD, Literal: "old"}}},
		Identifier:          NewTestIdentifier("x"),
	}

	if oldExpr.TokenLiteral() != "old" {
		t.Errorf("OldExpression.TokenLiteral() = %q, want %q", oldExpr.TokenLiteral(), "old")
	}
}

func TestFunctionDeclWithContracts(t *testing.T) {
	funcDecl := &FunctionDecl{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
		Name:     NewTestIdentifier("TestFunc"),
		Parameters: []*Parameter{
			{
				Name:  NewTestIdentifier("x"),
				Type:  NewTestTypeAnnotation("Integer"),
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
			},
		},
		ReturnType: NewTestTypeAnnotation("Integer"),
		PreConditions: &PreConditions{
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.REQUIRE, Literal: "require"}},
			Conditions: []*Condition{
				{
					Test: &BinaryExpression{
						TypedExpressionBase: TypedExpressionBase{
							BaseNode: BaseNode{Token: lexer.Token{Type: lexer.GREATER, Literal: ">"}},
						},
						Left:     &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}}}, Value: "x"},
						Operator: ">",
						Right:    &IntegerLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.INT, Literal: "0"}}}, Value: 0},
					},
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}},
				},
			},
		},
		Body: &BlockStatement{
			BaseNode:   BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"}},
			Statements: []Statement{},
		},
		PostConditions: &PostConditions{
			BaseNode: BaseNode{Token: lexer.Token{Type: lexer.ENSURE, Literal: "ensure"}},
			Conditions: []*Condition{
				{
					Test: &BinaryExpression{
						TypedExpressionBase: TypedExpressionBase{
							BaseNode: BaseNode{Token: lexer.Token{Type: lexer.GREATER, Literal: ">"}},
						},
						Left:     &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"}}}, Value: "Result"},
						Operator: ">",
						Right:    &IntegerLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.INT, Literal: "0"}}}, Value: 0},
					},
					BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"}},
				},
			},
		},
	}

	result := funcDecl.String()

	// Check that the string contains key elements
	if !strings.Contains(result, "function TestFunc") {
		t.Errorf("FunctionDecl.String() should contain function signature")
	}
	if !strings.Contains(result, "require") {
		t.Errorf("FunctionDecl.String() should contain preconditions")
	}
	if !strings.Contains(result, "ensure") {
		t.Errorf("FunctionDecl.String() should contain postconditions")
	}
	if !strings.Contains(result, "(x > 0)") {
		t.Errorf("FunctionDecl.String() should contain precondition test")
	}
	if !strings.Contains(result, "(Result > 0)") {
		t.Errorf("FunctionDecl.String() should contain postcondition test")
	}
}

func TestOldExpressionInPostCondition(t *testing.T) {
	postcond := &PostConditions{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.ENSURE, Literal: "ensure"}},
		Conditions: []*Condition{
			{
				Test: &BinaryExpression{
					TypedExpressionBase: TypedExpressionBase{
						BaseNode: BaseNode{Token: lexer.Token{Type: lexer.EQ, Literal: "="}},
					},
					Left:     &Identifier{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"}}}, Value: "Result"},
					Operator: "=",
					Right: &BinaryExpression{
						TypedExpressionBase: TypedExpressionBase{
							BaseNode: BaseNode{Token: lexer.Token{Type: lexer.PLUS, Literal: "+"}},
						},
						Left: &OldExpression{
							TypedExpressionBase: TypedExpressionBase{
								BaseNode: BaseNode{Token: lexer.Token{Type: lexer.OLD, Literal: "old"}},
							},
							Identifier: NewTestIdentifier("x"),
						},
						Operator: "+",
						Right:    &IntegerLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: BaseNode{Token: lexer.Token{Type: lexer.INT, Literal: "1"}}}, Value: 1},
					},
				},
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"}},
			},
		},
	}

	result := postcond.String()

	// Check that 'old x' appears in the output
	if !strings.Contains(result, "old x") {
		t.Errorf("PostConditions.String() should contain 'old x', got: %s", result)
	}
}
