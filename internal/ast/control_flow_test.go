package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
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
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IF, Literal: "if"}},
				Condition: NewTestBinaryExpression(NewTestIdentifier("x"), ">", NewTestIntegerLiteral(0)),
				Consequence: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestStringLiteral("positive"),
						},
					),
				},
			},
			expected: "if (x > 0) then PrintLn(\"positive\")",
		},
		{
			name: "if with else",
			stmt: &IfStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IF, Literal: "if"}},
				Condition: NewTestBinaryExpression(NewTestIdentifier("x"), ">", NewTestIntegerLiteral(0)),
				Consequence: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestStringLiteral("positive"),
						},
					),
				},
				Alternative: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestStringLiteral("non-positive"),
						},
					),
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
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.WHILE, Literal: "while"}},
		Condition: NewTestBinaryExpression(NewTestIdentifier("x"), "<", NewTestIntegerLiteral(10)),
		Body: &AssignmentStatement{
			Target: NewTestIdentifier("x"),
			Value:  NewTestBinaryExpression(NewTestIdentifier("x"), "+", NewTestIntegerLiteral(1)),
		},
	}

	expected := "while (x < 10) do x := (x + 1)"
	if got := stmt.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}

func TestRepeatStatementString(t *testing.T) {
	stmt := &RepeatStatement{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.REPEAT, Literal: "repeat"}},
		Body: &AssignmentStatement{
			Target: NewTestIdentifier("x"),
			Value:  NewTestBinaryExpression(NewTestIdentifier("x"), "+", NewTestIntegerLiteral(1)),
		},
		Condition: NewTestBinaryExpression(NewTestIdentifier("x"), ">=", NewTestIntegerLiteral(10)),
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
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.FOR, Literal: "for"}},
				Variable:  NewTestIdentifier("i"),
				Start:     NewTestIntegerLiteral(1),
				EndValue:  NewTestIntegerLiteral(10),
				Direction: ForTo,
				Body: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestIdentifier("i"),
						},
					),
				},
			},
			expected: "for i := 1 to 10 do PrintLn(i)",
		},
		{
			name: "for loop descending",
			stmt: &ForStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.FOR, Literal: "for"}},
				Variable:  NewTestIdentifier("i"),
				Start:     NewTestIntegerLiteral(10),
				EndValue:  NewTestIntegerLiteral(1),
				Direction: ForDownto,
				Body: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestIdentifier("i"),
						},
					),
				},
			},
			expected: "for i := 10 downto 1 do PrintLn(i)",
		},
		{
			name: "for loop ascending with step",
			stmt: &ForStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.FOR, Literal: "for"}},
				Variable:  NewTestIdentifier("i"),
				Start:     NewTestIntegerLiteral(1),
				EndValue:  NewTestIntegerLiteral(10),
				Direction: ForTo,
				Step:      NewTestIntegerLiteral(2),
				Body: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestIdentifier("i"),
						},
					),
				},
			},
			expected: "for i := 1 to 10 step 2 do PrintLn(i)",
		},
		{
			name: "for loop descending with step",
			stmt: &ForStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.FOR, Literal: "for"}},
				Variable:  NewTestIdentifier("i"),
				Start:     NewTestIntegerLiteral(10),
				EndValue:  NewTestIntegerLiteral(1),
				Direction: ForDownto,
				Step:      NewTestIntegerLiteral(3),
				Body: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestIdentifier("i"),
						},
					),
				},
			},
			expected: "for i := 10 downto 1 step 3 do PrintLn(i)",
		},
		{
			name: "for loop with step expression",
			stmt: &ForStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.FOR, Literal: "for"}},
				Variable:  NewTestIdentifier("i"),
				Start:     NewTestIntegerLiteral(0),
				EndValue:  NewTestIntegerLiteral(20),
				Direction: ForTo,
				Step:      NewTestBinaryExpression(NewTestIntegerLiteral(2), "+", NewTestIntegerLiteral(1)),
				Body: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestIdentifier("i"),
						},
					),
				},
			},
			expected: "for i := 0 to 20 step (2 + 1) do PrintLn(i)",
		},
		{
			name: "for loop with inline var and step",
			stmt: &ForStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.FOR, Literal: "for"}},
				Variable:  NewTestIdentifier("i"),
				Start:     NewTestIntegerLiteral(0),
				EndValue:  NewTestIntegerLiteral(10),
				Direction: ForTo,
				Step:      NewTestIntegerLiteral(2),
				InlineVar: true,
				Body: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestIdentifier("i"),
						},
					),
				},
			},
			expected: "for var i := 0 to 10 step 2 do PrintLn(i)",
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
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.CASE, Literal: "case"}},
		Expression: NewTestIdentifier("x"),
		Cases: []*CaseBranch{
			{
				Values: []Expression{
					NewTestIntegerLiteral(1),
				},
				Statement: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestStringLiteral("one"),
						},
					),
				},
			},
			{
				Values: []Expression{
					NewTestIntegerLiteral(2),
					NewTestIntegerLiteral(3),
				},
				Statement: &ExpressionStatement{
					Expression: NewTestCallExpression(
						NewTestIdentifier("PrintLn"),
						[]Expression{
							NewTestStringLiteral("two or three"),
						},
					),
				},
			},
		},
		Else: &ExpressionStatement{
			Expression: NewTestCallExpression(
				NewTestIdentifier("PrintLn"),
				[]Expression{
					NewTestStringLiteral("other"),
				},
			),
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
		expected  string
		direction ForDirection
	}{
		{expected: "to", direction: ForTo},
		{expected: "downto", direction: ForDownto},
		{expected: "unknown", direction: ForDirection(999)},
	}

	for _, tt := range tests {
		if got := tt.direction.String(); got != tt.expected {
			t.Errorf("ForDirection(%v).String() = %q, want %q", tt.direction, got, tt.expected)
		}
	}
}

func TestBreakStatementString(t *testing.T) {
	stmt := &BreakStatement{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BREAK, Literal: "break"}}},
	}

	expected := "break;"
	if got := stmt.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}

func TestContinueStatementString(t *testing.T) {
	stmt := &ContinueStatement{
		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.CONTINUE, Literal: "continue"}}},
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
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"}},
				ReturnValue: nil,
			},
			expected: "Exit",
		},
		{
			name: "exit with integer value",
			stmt: &ExitStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"}},
				ReturnValue: NewTestIntegerLiteral(-1),
			},
			expected: "Exit -1",
		},
		{
			name: "exit with identifier value",
			stmt: &ExitStatement{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.EXIT, Literal: "exit"}},
				ReturnValue: NewTestIdentifier("result"),
			},
			expected: "Exit result",
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

func TestControlFlowNodesImplementInterfaces(_ *testing.T) {
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
