package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

// TestIdentifiers tests parsing of identifiers.
func TestIdentifiers(t *testing.T) {
	input := "foobar;"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression is not ast.Identifier. got=%T", stmt.Expression)
	}

	if ident.Value != "foobar" {
		t.Errorf("ident.Value = %q, want %q", ident.Value, "foobar")
	}

	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral() = %q, want %q", ident.TokenLiteral(), "foobar")
	}
}

// TestPrefixExpressions tests parsing of prefix expressions.
func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		value    any
		input    string
		operator string
	}{
		{value: 5, input: "-5;", operator: "-"},
		{value: 10, input: "+10;", operator: "+"},
		{value: true, input: "not true;", operator: "not"},
		{value: false, input: "not false;", operator: "not"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
			}

			exp, ok := stmt.Expression.(*ast.UnaryExpression)
			if !ok {
				t.Fatalf("expression is not ast.UnaryExpression. got=%T", stmt.Expression)
			}

			if exp.Operator != tt.operator {
				t.Errorf("exp.Operator = %q, want %q", exp.Operator, tt.operator)
			}

			if !testLiteralExpression(t, exp.Right, tt.value) {
				return
			}
		})
	}
}

// TestInfixExpressions tests parsing of infix expressions.
func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		leftValue  any
		rightValue any
		input      string
		operator   string
	}{
		{leftValue: 5, rightValue: 5, input: "5 + 5;", operator: "+"},
		{leftValue: 5, rightValue: 5, input: "5 - 5;", operator: "-"},
		{leftValue: 5, rightValue: 5, input: "5 * 5;", operator: "*"},
		{leftValue: 5, rightValue: 5, input: "5 / 5;", operator: "/"},
		{leftValue: 5, rightValue: 5, input: "5 > 5;", operator: ">"},
		{leftValue: 5, rightValue: 5, input: "5 < 5;", operator: "<"},
		{leftValue: 5, rightValue: 5, input: "5 = 5;", operator: "="},
		{leftValue: 5, rightValue: 5, input: "5 <> 5;", operator: "<>"},
		{leftValue: true, rightValue: false, input: "true and false;", operator: "and"},
		{leftValue: true, rightValue: false, input: "true or false;", operator: "or"},
		{leftValue: 2, rightValue: 3, input: "2 shl 3;", operator: "shl"},
		{leftValue: 16, rightValue: 2, input: "16 shr 2;", operator: "shr"},
		{leftValue: 5, rightValue: 3, input: "5 and 3;", operator: "and"},
		{leftValue: 5, rightValue: 3, input: "5 or 3;", operator: "or"},
		{leftValue: 5, rightValue: 3, input: "5 xor 3;", operator: "xor"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
			}

			if !testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue) {
				return
			}
		})
	}
}

// TestOperatorPrecedence tests that operators have correct precedence.
func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"not -a", "(not (-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b / c", "(a + (b / c))"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4; -5 * 5", "(3 + 4)((-5) * 5)"},
		{"5 > 4 = 3 < 4", "((5 > 4) = (3 < 4))"},
		{"5 < 4 <> 3 > 4", "((5 < 4) <> (3 > 4))"},
		{"3 + 4 * 5 = 3 * 1 + 4 * 5", "((3 + (4 * 5)) = ((3 * 1) + (4 * 5)))"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 = false", "((3 > 5) = false)"},
		{"3 < 5 = true", "((3 < 5) = true)"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 / (5 + 5)", "(2 / (5 + 5))"},
		{"-(5 + 5)", "(-(5 + 5))"},
		{"not (true = true)", "(not (true = true))"},
		{"2 shl 3", "(2 shl 3)"},
		{"16 shr 2", "(16 shr 2)"},
		{"2 + 3 shl 4", "(2 + (3 shl 4))"},
		{"2 shl 3 * 5", "(2 shl (3 * 5))"},
		{"2 shl 3 + 4", "((2 shl 3) + 4)"},
		{"a and b shl 1", "(a and (b shl 1))"},
		{"(2 shl 1) or 1", "((2 shl 1) or 1)"},
		{"5 and 3 or 2", "((5 and 3) or 2)"},
		{"5 or 3 and 2", "(5 or (3 and 2))"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			actual := program.String()
			if actual != tt.expected {
				t.Errorf("expected=%q, got=%q", tt.expected, actual)
			}
		})
	}
}

// TestGroupedExpressions tests parsing of grouped expressions.
func TestGroupedExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"(5)", "5"},
		{"(5 + 5)", "(5 + 5)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 / (5 + 5)", "(2 / (5 + 5))"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			actual := program.String()
			if actual != tt.expected {
				t.Errorf("expected=%q, got=%q", tt.expected, actual)
			}
		})
	}
}

// TestNotInOperator tests parsing of "not in" operator combinations.
// DWScript supports both "not (x in set)" and "x not in set" syntax,
// and they should parse to the same AST structure.
func TestNotInOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "not with parentheses",
			input:    "not (x in mySet);",
			expected: "(not (x in mySet))",
		},
		{
			name:     "x not in mySet",
			input:    "x not in mySet;",
			expected: "(not (x in mySet))",
		},
		{
			name:     "x not in set literal",
			input:    "x not in [1, 2, 3];",
			expected: "(not (x in [1, 2, 3]))",
		},
		{
			name:     "not in with complex left expression",
			input:    "x + 1 not in mySet;",
			expected: "(not ((x + 1) in mySet))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			actual := program.String()
			if actual != tt.expected {
				t.Errorf("expected=%q, got=%q", tt.expected, actual)
			}
		})
	}
}

// TestNotIsOperator tests parsing of "not is" operator combinations.
// Similar to "not in", DWScript supports both "not (obj is TClass)" and "obj not is TClass".
func TestNotIsOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "not with parentheses",
			input:    "not (obj is TMyClass);",
			expected: "(not (obj is TMyClass))",
		},
		{
			name:     "obj not is TClass",
			input:    "obj not is TMyClass;",
			expected: "(not (obj is TMyClass))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			actual := program.String()
			if actual != tt.expected {
				t.Errorf("expected=%q, got=%q", tt.expected, actual)
			}
		})
	}
}

// TestNotAsOperator tests parsing of "not as" operator combinations.
// Similar to "not in/is", DWScript supports both "not (obj as IInterface)" and "obj not as IInterface".
func TestNotAsOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "not with parentheses",
			input:    "not (obj as IMyInterface);",
			expected: "(not (obj as IMyInterface))",
		},
		{
			name:     "obj not as IInterface",
			input:    "obj not as IMyInterface;",
			expected: "(not (obj as IMyInterface))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			actual := program.String()
			if actual != tt.expected {
				t.Errorf("expected=%q, got=%q", tt.expected, actual)
			}
		})
	}
}

// TestIsOperatorWithBooleans tests parsing of 'is' operator with boolean value expressions.
func TestIsOperatorWithBooleans(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "is True",
			input:    "(x = 0) is True;",
			expected: "((x = 0) is True)",
		},
		{
			name:     "is False",
			input:    "(x = 1) is False;",
			expected: "((x = 1) is False)",
		},
		{
			name:     "True is True",
			input:    "True is True;",
			expected: "(True is True)",
		},
		{
			name:     "True is False",
			input:    "True is False;",
			expected: "(True is False)",
		},
		{
			name:     "is with parenthesized expression",
			input:    "(x = 0) is (y = 1);",
			expected: "((x = 0) is (y = 1))",
		},
		{
			name:     "not with is True",
			input:    "(x = 0) is not True;",
			expected: "((x = 0) is (not True))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			actual := program.String()
			if actual != tt.expected {
				t.Errorf("expected=%q, got=%q", tt.expected, actual)
			}
		})
	}
}

func TestIsOperatorPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "is True and - precedence ensures 'and' is outside is",
			input:    "a is True and b;",
			expected: "((a is True) and b)",
		},
		{
			name:     "is False or - precedence ensures 'or' is outside is",
			input:    "a is False or b;",
			expected: "((a is False) or b)",
		},
		{
			name:     "complex: is True and is False",
			input:    "a is True and b is False;",
			expected: "((a is True) and (b is False))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			actual := program.String()
			if actual != tt.expected {
				t.Errorf("expected=%q, got=%q", tt.expected, actual)
			}
		})
	}
}
