package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/lexer"
)

func TestIfStatementString(t *testing.T) {
	tests := []struct {
		name     string
		stmt     *IfStatement
		expected string
	}{
		{
			name: "simple if without else",
			stmt: &IfStatement{
				Token: lexer.Token{Type: lexer.IF, Literal: "if"},
				Condition: &BinaryExpression{
					Left:     &Identifier{Value: "x"},
					Operator: ">",
					Right:    &IntegerLiteral{Token: lexer.Token{Literal: "0"}},
				},
				Consequence: &ExpressionStatement{
					Expression: &CallExpression{
						Function: &Identifier{Value: "PrintLn"},
						Arguments: []Expression{
							&StringLiteral{Value: "positive"},
						},
					},
				},
			},
			expected: "if (x > 0) then PrintLn(\"positive\")",
		},
		{
			name: "if with else",
			stmt: &IfStatement{
				Token: lexer.Token{Type: lexer.IF, Literal: "if"},
				Condition: &BinaryExpression{
					Left:     &Identifier{Value: "x"},
					Operator: ">",
					Right:    &IntegerLiteral{Token: lexer.Token{Literal: "0"}},
				},
				Consequence: &ExpressionStatement{
					Expression: &CallExpression{
						Function: &Identifier{Value: "PrintLn"},
						Arguments: []Expression{
							&StringLiteral{Value: "positive"},
						},
					},
				},
				Alternative: &ExpressionStatement{
					Expression: &CallExpression{
						Function: &Identifier{Value: "PrintLn"},
						Arguments: []Expression{
							&StringLiteral{Value: "non-positive"},
						},
					},
				},
			},
			expected: "if (x > 0) then PrintLn(\"positive\") else PrintLn(\"non-positive\")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stmt.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWhileStatementString(t *testing.T) {
	stmt := &WhileStatement{
		Token: lexer.Token{Type: lexer.WHILE, Literal: "while"},
		Condition: &BinaryExpression{
			Left:     &Identifier{Value: "x"},
			Operator: "<",
			Right:    &IntegerLiteral{Token: lexer.Token{Literal: "10"}},
		},
		Body: &AssignmentStatement{
			Target: &Identifier{Value: "x"},
			Value: &BinaryExpression{
				Left:     &Identifier{Value: "x"},
				Operator: "+",
				Right:    &IntegerLiteral{Token: lexer.Token{Literal: "1"}},
			},
		},
	}

	expected := "while (x < 10) do x := (x + 1)"
	if got := stmt.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}

func TestRepeatStatementString(t *testing.T) {
	stmt := &RepeatStatement{
		Token: lexer.Token{Type: lexer.REPEAT, Literal: "repeat"},
		Body: &AssignmentStatement{
			Target: &Identifier{Value: "x"},
			Value: &BinaryExpression{
				Left:     &Identifier{Value: "x"},
				Operator: "+",
				Right:    &IntegerLiteral{Token: lexer.Token{Literal: "1"}},
			},
		},
		Condition: &BinaryExpression{
			Left:     &Identifier{Value: "x"},
			Operator: ">=",
			Right:    &IntegerLiteral{Token: lexer.Token{Literal: "10"}},
		},
	}

	expected := "repeat x := (x + 1) until (x >= 10)"
	if got := stmt.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}

func TestForStatementString(t *testing.T) {
	tests := []struct {
		name     string
		stmt     *ForStatement
		expected string
	}{
		{
			name: "for loop ascending",
			stmt: &ForStatement{
				Token:     lexer.Token{Type: lexer.FOR, Literal: "for"},
				Variable:  &Identifier{Value: "i"},
				Start:     &IntegerLiteral{Token: lexer.Token{Literal: "1"}},
				End:       &IntegerLiteral{Token: lexer.Token{Literal: "10"}},
				Direction: ForTo,
				Body: &ExpressionStatement{
					Expression: &CallExpression{
						Function: &Identifier{Value: "PrintLn"},
						Arguments: []Expression{
							&Identifier{Value: "i"},
						},
					},
				},
			},
			expected: "for i := 1 to 10 do PrintLn(i)",
		},
		{
			name: "for loop descending",
			stmt: &ForStatement{
				Token:     lexer.Token{Type: lexer.FOR, Literal: "for"},
				Variable:  &Identifier{Value: "i"},
				Start:     &IntegerLiteral{Token: lexer.Token{Literal: "10"}},
				End:       &IntegerLiteral{Token: lexer.Token{Literal: "1"}},
				Direction: ForDownto,
				Body: &ExpressionStatement{
					Expression: &CallExpression{
						Function: &Identifier{Value: "PrintLn"},
						Arguments: []Expression{
							&Identifier{Value: "i"},
						},
					},
				},
			},
			expected: "for i := 10 downto 1 do PrintLn(i)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stmt.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCaseStatementString(t *testing.T) {
	stmt := &CaseStatement{
		Token:      lexer.Token{Type: lexer.CASE, Literal: "case"},
		Expression: &Identifier{Value: "x"},
		Cases: []*CaseBranch{
			{
				Values: []Expression{
					&IntegerLiteral{Token: lexer.Token{Literal: "1"}},
				},
				Statement: &ExpressionStatement{
					Expression: &CallExpression{
						Function: &Identifier{Value: "PrintLn"},
						Arguments: []Expression{
							&StringLiteral{Value: "one"},
						},
					},
				},
			},
			{
				Values: []Expression{
					&IntegerLiteral{Token: lexer.Token{Literal: "2"}},
					&IntegerLiteral{Token: lexer.Token{Literal: "3"}},
				},
				Statement: &ExpressionStatement{
					Expression: &CallExpression{
						Function: &Identifier{Value: "PrintLn"},
						Arguments: []Expression{
							&StringLiteral{Value: "two or three"},
						},
					},
				},
			},
		},
		Else: &ExpressionStatement{
			Expression: &CallExpression{
				Function: &Identifier{Value: "PrintLn"},
				Arguments: []Expression{
					&StringLiteral{Value: "other"},
				},
			},
		},
	}

	result := stmt.String()

	// Check that it contains the expected parts
	expectedParts := []string{
		"case x of",
		"1: PrintLn(\"one\")",
		"2, 3: PrintLn(\"two or three\")",
		"else",
		"PrintLn(\"other\")",
		"end",
	}

	for _, part := range expectedParts {
		if !contains(result, part) {
			t.Errorf("String() output does not contain expected part %q\nGot: %q", part, result)
		}
	}
}

func TestForDirectionString(t *testing.T) {
	tests := []struct {
		direction ForDirection
		expected  string
	}{
		{ForTo, "to"},
		{ForDownto, "downto"},
		{ForDirection(999), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.direction.String(); got != tt.expected {
			t.Errorf("ForDirection(%v).String() = %q, want %q", tt.direction, got, tt.expected)
		}
	}
}

func TestBreakStatementString(t *testing.T) {
	stmt := &BreakStatement{
		Token: lexer.Token{Type: lexer.BREAK, Literal: "break"},
	}

	expected := "break;"
	if got := stmt.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}

func TestContinueStatementString(t *testing.T) {
	stmt := &ContinueStatement{
		Token: lexer.Token{Type: lexer.CONTINUE, Literal: "continue"},
	}

	expected := "continue;"
	if got := stmt.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}

func TestExitStatementString(t *testing.T) {
	tests := []struct {
		name     string
		stmt     *ExitStatement
		expected string
	}{
		{
			name: "exit without value",
			stmt: &ExitStatement{
				Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"},
				Value: nil,
			},
			expected: "exit;",
		},
		{
			name: "exit with integer value",
			stmt: &ExitStatement{
				Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"},
				Value: &IntegerLiteral{Token: lexer.Token{Literal: "-1"}, Value: -1},
			},
			expected: "exit(-1);",
		},
		{
			name: "exit with identifier value",
			stmt: &ExitStatement{
				Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"},
				Value: &Identifier{Value: "result"},
			},
			expected: "exit(result);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stmt.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestControlFlowNodesImplementInterfaces(t *testing.T) {
	// Ensure all control flow nodes implement the Statement interface
	var _ Statement = (*IfStatement)(nil)
	var _ Statement = (*WhileStatement)(nil)
	var _ Statement = (*RepeatStatement)(nil)
	var _ Statement = (*ForStatement)(nil)
	var _ Statement = (*CaseStatement)(nil)
	var _ Statement = (*BreakStatement)(nil)
	var _ Statement = (*ContinueStatement)(nil)
	var _ Statement = (*ExitStatement)(nil)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
