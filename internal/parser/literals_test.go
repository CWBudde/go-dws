package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestIntegerLiterals tests parsing of integer literals.
func TestIntegerLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5;", 5},
		{"10;", 10},
		{"0;", 0},
		{"999;", 999},
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

			literal, ok := stmt.Expression.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("expression is not ast.IntegerLiteral. got=%T", stmt.Expression)
			}

			if literal.Value != tt.expected {
				t.Errorf("literal.Value = %d, want %d", literal.Value, tt.expected)
			}

			if literal.TokenLiteral() != fmt.Sprintf("%d", tt.expected) {
				t.Errorf("literal.TokenLiteral() = %q, want %q", literal.TokenLiteral(), fmt.Sprintf("%d", tt.expected))
			}
		})
	}
}

func TestIntegerLiterals_AlternateBases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{name: "hex_dollar_prefix", input: "$2A;", expected: 42},
		{name: "hex_zero_x_prefix_lower", input: "0x2a;", expected: 42},
		{name: "hex_zero_x_prefix_upper", input: "0X2A;", expected: 42},
		{name: "binary_percent_prefix", input: "%101010;", expected: 42},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
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

			literal, ok := stmt.Expression.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("expression is not ast.IntegerLiteral. got=%T", stmt.Expression)
			}

			if literal.Value != tt.expected {
				t.Errorf("literal.Value = %d, want %d", literal.Value, tt.expected)
			}

			if literal.TokenLiteral() != strings.TrimSuffix(tt.input, ";") {
				t.Errorf("literal.TokenLiteral() = %q, want %q", literal.TokenLiteral(), strings.TrimSuffix(tt.input, ";"))
			}
		})
	}
}

// TestFloatLiterals tests parsing of float literals.
func TestFloatLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5.5;", 5.5},
		{"0.0;", 0.0},
		{"3.14159;", 3.14159},
		{"1.5e10;", 1.5e10},
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

			literal, ok := stmt.Expression.(*ast.FloatLiteral)
			if !ok {
				t.Fatalf("expression is not ast.FloatLiteral. got=%T", stmt.Expression)
			}

			if literal.Value != tt.expected {
				t.Errorf("literal.Value = %f, want %f", literal.Value, tt.expected)
			}
		})
	}
}

// TestStringLiterals tests parsing of string literals.
func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`'hello';`, "hello"},
		{`'';`, ""},
		{`'hello world';`, "hello world"},
		{`"test";`, "test"},
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

			literal, ok := stmt.Expression.(*ast.StringLiteral)
			if !ok {
				t.Fatalf("expression is not ast.StringLiteral. got=%T", stmt.Expression)
			}

			if literal.Value != tt.expected {
				t.Errorf("literal.Value = %q, want %q", literal.Value, tt.expected)
			}
		})
	}
}

// TestBooleanLiterals tests parsing of boolean literals.
func TestBooleanLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
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

			literal, ok := stmt.Expression.(*ast.BooleanLiteral)
			if !ok {
				t.Fatalf("expression is not ast.BooleanLiteral. got=%T", stmt.Expression)
			}

			if literal.Value != tt.expected {
				t.Errorf("literal.Value = %v, want %v", literal.Value, tt.expected)
			}
		})
	}
}

// TestCharLiterals tests parsing of character literals.
func TestCharLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected rune
	}{
		{"#65;", 'A'},   // Decimal: A
		{"#$41;", 'A'},  // Hex: A
		{"#13;", '\r'},  // Carriage return
		{"#10;", '\n'},  // Line feed
		{"#$61;", 'a'},  // Hex: a
		{"#32;", ' '},   // Space
		{"#$0D;", '\r'}, // Hex CR
		{"#$0A;", '\n'}, // Hex LF
		{"#48;", '0'},   // Digit 0
		{"#$30;", '0'},  // Hex digit 0
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

			literal, ok := stmt.Expression.(*ast.CharLiteral)
			if !ok {
				t.Fatalf("expression is not ast.CharLiteral. got=%T", stmt.Expression)
			}

			if literal.Value != tt.expected {
				t.Errorf("literal.Value = %q (%d), want %q (%d)", literal.Value, literal.Value, tt.expected, tt.expected)
			}
		})
	}
}
