package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Helper function to create a parser from input string

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

// TestBlockStatement tests parsing of block statements.
func TestBlockStatement(t *testing.T) {
	input := `
begin
  5;
  10;
end;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	block, ok := program.Statements[0].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("statement is not ast.BlockStatement. got=%T", program.Statements[0])
	}

	if len(block.Statements) != 2 {
		t.Fatalf("block has wrong number of statements. got=%d", len(block.Statements))
	}

	for i, stmt := range block.Statements {
		exprStmt, ok := stmt.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("block statement %d is not ExpressionStatement. got=%T", i, stmt)
		}
		if !testIntegerLiteral(t, exprStmt.Expression, int64((i*5)+5)) {
			return
		}
	}
}

// TestBlockStatementAssignments ensures assignments inside blocks are parsed correctly.
func TestBlockStatementAssignments(t *testing.T) {
	input := `
begin
  x := 1;
  y := x + 2;
end;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	block, ok := program.Statements[0].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("statement is not ast.BlockStatement. got=%T", program.Statements[0])
	}

	if len(block.Statements) != 2 {
		t.Fatalf("block has wrong number of statements. got=%d", len(block.Statements))
	}

	first, ok := block.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("first statement not AssignmentStatement. got=%T", block.Statements[0])
	}
	firstTarget, ok := first.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("first.Target not *ast.Identifier. got=%T", first.Target)
	}
	if firstTarget.Value != "x" {
		t.Errorf("first assignment name = %q, want %q", firstTarget.Value, "x")
	}
	if !testIntegerLiteral(t, first.Value, 1) {
		return
	}

	second, ok := block.Statements[1].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("second statement not AssignmentStatement. got=%T", block.Statements[1])
	}
	secondTarget, ok := second.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("second.Target not *ast.Identifier. got=%T", second.Target)
	}
	if secondTarget.Value != "y" {
		t.Errorf("second assignment name = %q, want %q", secondTarget.Value, "y")
	}
	if !testInfixExpression(t, second.Value, "x", "+", 2) {
		return
	}
}

// TestVarDeclarations tests parsing of variable declarations.
func TestVarDeclarations(t *testing.T) {
	input := `
var x: Integer;
var y := 5;
var s: String := 'hello';
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		assertValue func(*testing.T, ast.Expression)
		name        string
		expectedVar string
		expectedTyp string
		expectValue bool
	}{
		{
			name:        "typed declaration without initializer",
			expectedVar: "x",
			expectedTyp: "Integer",
			expectValue: false,
		},
		{
			name:        "inferred integer declaration",
			expectedVar: "y",
			expectedTyp: "",
			expectValue: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				if !testIntegerLiteral(t, expr, 5) {
					t.Fatalf("value is not expected integer literal")
				}
			},
		},
		{
			name:        "string declaration with type",
			expectedVar: "s",
			expectedTyp: "String",
			expectValue: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				if !testStringLiteralExpression(t, expr, "hello") {
					t.Fatalf("value is not expected string literal")
				}
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, ok := program.Statements[i].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("statement is not ast.VarDeclStatement. got=%T", program.Statements[i])
			}

			if len(stmt.Names) == 0 || stmt.Names[0].Value != tt.expectedVar {
				if len(stmt.Names) == 0 {
					t.Errorf("stmt.Names is empty, want %q", tt.expectedVar)
				} else {
					t.Errorf("stmt.Names[0].Value = %q, want %q", stmt.Names[0].Value, tt.expectedVar)
				}
			}

			if (stmt.Type == nil && tt.expectedTyp != "") || (stmt.Type != nil && stmt.Type.Name != tt.expectedTyp) {
				t.Errorf("stmt.Type = %q, want %q", stmt.Type, tt.expectedTyp)
			}

			if tt.expectValue {
				if stmt.Value == nil {
					t.Fatalf("expected initialization expression")
				}
				tt.assertValue(t, stmt.Value)
			} else if stmt.Value != nil {
				t.Fatalf("expected no initialization, got %T", stmt.Value)
			}
		})
	}
}

// TestExternalVarParsing tests parsing of external variable declarations.
// Task 7.143: External variables are declared with the 'external' keyword:
//
//	var x: Integer; external;
//	var y: String; external 'externalName';
func TestExternalVarParsing(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedVar  string
		expectedType string
		externalName string
		isExternal   bool
		expectError  bool
	}{
		{
			name:         "external variable without custom name",
			input:        "var x: Integer external;",
			expectedVar:  "x",
			expectedType: "Integer",
			isExternal:   true,
			externalName: "",
		},
		{
			name:         "external variable with custom name",
			input:        "var y: String external 'customName';",
			expectedVar:  "y",
			expectedType: "String",
			isExternal:   true,
			externalName: "customName",
		},
		{
			name:         "regular variable (not external)",
			input:        "var z: Float;",
			expectedVar:  "z",
			expectedType: "Float",
			isExternal:   false,
			externalName: "",
		},
		{
			name:         "regular variable with initialization",
			input:        "var w: Integer := 42;",
			expectedVar:  "w",
			expectedType: "Integer",
			isExternal:   false,
			externalName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()

			if tt.expectError {
				if len(p.Errors()) == 0 {
					t.Fatalf("expected parser error, got none")
				}
				return
			}

			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("statement is not ast.VarDeclStatement. got=%T", program.Statements[0])
			}

			if len(stmt.Names) == 0 || stmt.Names[0].Value != tt.expectedVar {
				if len(stmt.Names) == 0 {
					t.Errorf("stmt.Names is empty, want %q", tt.expectedVar)
				} else {
					t.Errorf("stmt.Names[0].Value = %q, want %q", stmt.Names[0].Value, tt.expectedVar)
				}
			}

			if stmt.Type == nil || stmt.Type.Name != tt.expectedType {
				var gotType string
				if stmt.Type != nil {
					gotType = stmt.Type.Name
				}
				t.Errorf("stmt.Type.Name = %q, want %q", gotType, tt.expectedType)
			}

			if stmt.IsExternal != tt.isExternal {
				t.Errorf("stmt.IsExternal = %v, want %v", stmt.IsExternal, tt.isExternal)
			}

			if stmt.ExternalName != tt.externalName {
				t.Errorf("stmt.ExternalName = %q, want %q", stmt.ExternalName, tt.externalName)
			}
		})
	}
}

// TestMultiIdentifierVarDeclarations tests parsing of multi-identifier variable declarations.
// Task 9.63: DWScript allows comma-separated variable names like `var a, b, c: Integer;`.
func TestMultiIdentifierVarDeclarations(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedNames []string
		expectedType  string
		expectError   bool
		errorContains string
	}{
		{
			name:          "two variables",
			input:         "var x, y: Integer;",
			expectedNames: []string{"x", "y"},
			expectedType:  "Integer",
		},
		{
			name:          "three variables",
			input:         "var a, b, c: String;",
			expectedNames: []string{"a", "b", "c"},
			expectedType:  "String",
		},
		{
			name:          "many variables",
			input:         "var i, j, k, l, m: Integer;",
			expectedNames: []string{"i", "j", "k", "l", "m"},
			expectedType:  "Integer",
		},
		{
			name:          "reject initializer with multiple names",
			input:         "var x, y: Integer := 42;",
			expectError:   true,
			errorContains: "cannot use initializer with multiple variable names",
		},
		{
			name:          "reject initializer without type",
			input:         "var a, b := 5;",
			expectError:   true,
			errorContains: "cannot use initializer with multiple variable names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()

			if tt.expectError {
				if len(p.Errors()) == 0 {
					t.Fatalf("expected error containing %q, got no errors", tt.errorContains)
				}
				found := false
				for _, err := range p.Errors() {
					if strings.Contains(err, tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got %v", tt.errorContains, p.Errors())
				}
				return
			}

			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("statement is not ast.VarDeclStatement. got=%T", program.Statements[0])
			}

			if len(stmt.Names) != len(tt.expectedNames) {
				t.Fatalf("wrong number of names. got=%d, want=%d", len(stmt.Names), len(tt.expectedNames))
			}

			for i, expectedName := range tt.expectedNames {
				if stmt.Names[i].Value != expectedName {
					t.Errorf("name[%d] = %q, want %q", i, stmt.Names[i].Value, expectedName)
				}
			}

			if stmt.Type == nil || stmt.Type.Name != tt.expectedType {
				var typeName string
				if stmt.Type != nil {
					typeName = stmt.Type.Name
				}
				t.Errorf("stmt.Type.Name = %q, want %q", typeName, tt.expectedType)
			}

			// Test String() method for multi-names
			expectedStr := "var " + strings.Join(tt.expectedNames, ", ") + ": " + tt.expectedType
			if stmt.String() != expectedStr {
				t.Errorf("stmt.String() = %q, want %q", stmt.String(), expectedStr)
			}
		})
	}
}

// TestAssignmentStatements tests parsing of assignment statements.
func TestAssignmentStatements(t *testing.T) {
	input := `
x := 10;
y := x + 1;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt1, ok := program.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement 0 is not ast.AssignmentStatement. got=%T", program.Statements[0])
	}
	target1, ok := stmt1.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("stmt1.Target is not *ast.Identifier. got=%T", stmt1.Target)
	}
	if target1.Value != "x" {
		t.Errorf("stmt1.Target.Value = %q, want %q", target1.Value, "x")
	}
	if !testIntegerLiteral(t, stmt1.Value, 10) {
		return
	}

	stmt2, ok := program.Statements[1].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement 1 is not ast.AssignmentStatement. got=%T", program.Statements[1])
	}
	target2, ok := stmt2.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("stmt2.Target is not *ast.Identifier. got=%T", stmt2.Target)
	}
	if target2.Value != "y" {
		t.Errorf("stmt2.Target.Value = %q, want %q", target2.Value, "y")
	}
	if !testInfixExpression(t, stmt2.Value, "x", "+", 1) {
		return
	}
}

// TestCompoundAssignmentStatements tests parsing of compound assignment operators.
func TestCompoundAssignmentStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		target   string
		operator lexer.TokenType
		value    int64
	}{
		{
			name:     "Plus assign",
			input:    "x += 5;",
			target:   "x",
			operator: lexer.PLUS_ASSIGN,
			value:    5,
		},
		{
			name:     "Minus assign",
			input:    "count -= 1;",
			target:   "count",
			operator: lexer.MINUS_ASSIGN,
			value:    1,
		},
		{
			name:     "Times assign",
			input:    "total *= 10;",
			target:   "total",
			operator: lexer.TIMES_ASSIGN,
			value:    10,
		},
		{
			name:     "Divide assign",
			input:    "result /= 2;",
			target:   "result",
			operator: lexer.DIVIDE_ASSIGN,
			value:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
			if !ok {
				t.Fatalf("statement is not ast.AssignmentStatement. got=%T", program.Statements[0])
			}

			target, ok := stmt.Target.(*ast.Identifier)
			if !ok {
				t.Fatalf("stmt.Target is not *ast.Identifier. got=%T", stmt.Target)
			}
			if target.Value != tt.target {
				t.Errorf("stmt.Target.Value = %q, want %q", target.Value, tt.target)
			}

			if stmt.Operator != tt.operator {
				t.Errorf("stmt.Operator = %v, want %v", stmt.Operator, tt.operator)
			}

			if !testIntegerLiteral(t, stmt.Value, tt.value) {
				return
			}
		})
	}
}

// TestCompoundAssignmentWithMemberAccess tests compound assignment with member access.
func TestCompoundAssignmentWithMemberAccess(t *testing.T) {
	input := "obj.field += 10;"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement is not ast.AssignmentStatement. got=%T", program.Statements[0])
	}

	memberAccess, ok := stmt.Target.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("stmt.Target is not *ast.MemberAccessExpression. got=%T", stmt.Target)
	}

	objIdent, ok := memberAccess.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("memberAccess.Object is not *ast.Identifier. got=%T", memberAccess.Object)
	}
	if objIdent.Value != "obj" {
		t.Errorf("object name = %q, want %q", objIdent.Value, "obj")
	}

	if memberAccess.Member.Value != "field" {
		t.Errorf("field name = %q, want %q", memberAccess.Member.Value, "field")
	}

	if stmt.Operator != lexer.PLUS_ASSIGN {
		t.Errorf("stmt.Operator = %v, want %v", stmt.Operator, lexer.PLUS_ASSIGN)
	}

	if !testIntegerLiteral(t, stmt.Value, 10) {
		return
	}
}

// TestCompoundAssignmentWithIndexExpression tests compound assignment with array indexing.
func TestCompoundAssignmentWithIndexExpression(t *testing.T) {
	input := "arr[i] *= 2;"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement is not ast.AssignmentStatement. got=%T", program.Statements[0])
	}

	indexExpr, ok := stmt.Target.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("stmt.Target is not *ast.IndexExpression. got=%T", stmt.Target)
	}

	arrIdent, ok := indexExpr.Left.(*ast.Identifier)
	if !ok {
		t.Fatalf("indexExpr.Left is not *ast.Identifier. got=%T", indexExpr.Left)
	}
	if arrIdent.Value != "arr" {
		t.Errorf("array name = %q, want %q", arrIdent.Value, "arr")
	}

	indexIdent, ok := indexExpr.Index.(*ast.Identifier)
	if !ok {
		t.Fatalf("indexExpr.Index is not *ast.Identifier. got=%T", indexExpr.Index)
	}
	if indexIdent.Value != "i" {
		t.Errorf("index = %q, want %q", indexIdent.Value, "i")
	}

	if stmt.Operator != lexer.TIMES_ASSIGN {
		t.Errorf("stmt.Operator = %v, want %v", stmt.Operator, lexer.TIMES_ASSIGN)
	}

	if !testIntegerLiteral(t, stmt.Value, 2) {
		return
	}
}

// TestMemberAssignmentStatements tests parsing of member assignment statements.
// This tests the pattern: obj.field := value
func TestMemberAssignmentStatements(t *testing.T) {
	tests := []struct {
		value      interface{}
		name       string
		input      string
		objectName string
		fieldName  string
	}{
		{
			name:       "Simple member assignment",
			input:      "p.X := 10;",
			objectName: "p",
			fieldName:  "X",
			value:      int64(10),
		},
		{
			name:       "Self member assignment",
			input:      "Self.X := 42;",
			objectName: "Self",
			fieldName:  "X",
			value:      int64(42),
		},
		{
			name:       "Member assignment with expression",
			input:      "obj.Value := x + 5;",
			objectName: "obj",
			fieldName:  "Value",
			value:      "x + 5", // Will check it's a binary expression
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
			if !ok {
				t.Fatalf("statement is not ast.AssignmentStatement. got=%T", program.Statements[0])
			}

			// Check that Name is actually a MemberAccessExpression, not just an Identifier
			// For now, we'll validate the basic structure once parser is updated
			// This test will fail until parser supports member assignments
			t.Logf("Assignment statement Name: %T, Value: %T", stmt.Target, stmt.Value)
		})
	}
}

// TestCallExpressionParsing verifies parsing of function call expressions.
func TestCallExpressionParsing(t *testing.T) {
	input := "Add(1, 2 * 3, foo);"

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

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is not ast.CallExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, call.Function, "Add") {
		return
	}

	if len(call.Arguments) != 3 {
		t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
	}

	if !testLiteralExpression(t, call.Arguments[0], 1) {
		return
	}

	if !testInfixExpression(t, call.Arguments[1], 2, "*", 3) {
		return
	}

	if !testIdentifier(t, call.Arguments[2], "foo") {
		return
	}
}

// TestCallExpressionWithStringArgument ensures string literals work as call arguments.
func TestCallExpressionWithStringArgument(t *testing.T) {
	input := "PrintLn('hello', name);"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is not ast.CallExpression. got=%T", stmt.Expression)
	}

	if len(call.Arguments) != 2 {
		t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
	}

	if !testStringLiteralExpression(t, call.Arguments[0], "hello") {
		return
	}

	if !testIdentifier(t, call.Arguments[1], "name") {
		return
	}
}

// TestCallExpressionTrailingComma validates parsing of trailing commas.
func TestCallExpressionTrailingComma(t *testing.T) {
	input := "Foo(1, 2,);"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is not ast.CallExpression. got=%T", stmt.Expression)
	}

	if len(call.Arguments) != 2 {
		t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
	}

	if !testLiteralExpression(t, call.Arguments[0], 1) {
		return
	}

	if !testLiteralExpression(t, call.Arguments[1], 2) {
		return
	}
}

// TestCallExpressionPrecedence checks call precedence relative to infix operators.
func TestCallExpressionPrecedence(t *testing.T) {
	input := "foo(1 + 2) * 3;"

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

	// The call expression wraps the argument, then multiplication happens
	expected := "(foo((1 + 2)) * 3)"
	if stmt.Expression.String() != expected {
		t.Errorf("expression String() = %q, want %q", stmt.Expression.String(), expected)
	}
}

// TestStatementDispatch ensures identifiers followed by parentheses parse as calls, not assignments.
func TestStatementDispatch(t *testing.T) {
	input := `
var foo := 1;
foo();
foo := foo + 1;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	if _, ok := program.Statements[0].(*ast.VarDeclStatement); !ok {
		t.Fatalf("statement 0 expected VarDeclStatement. got=%T", program.Statements[0])
	}

	callStmt, ok := program.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement 1 expected ExpressionStatement. got=%T", program.Statements[1])
	}
	if _, ok := callStmt.Expression.(*ast.CallExpression); !ok {
		t.Fatalf("statement 1 expression expected CallExpression. got=%T", callStmt.Expression)
	}

	if _, ok := program.Statements[2].(*ast.AssignmentStatement); !ok {
		t.Fatalf("statement 2 expected AssignmentStatement. got=%T", program.Statements[2])
	}

	assign := program.Statements[2].(*ast.AssignmentStatement)
	if !testInfixExpression(t, assign.Value, "foo", "+", 1) {
		return
	}
}

// Helper function to test literal expressions

// TestCompleteSimplePrograms tests parsing of complete simple programs with multiple statement types.
func TestCompleteSimplePrograms(t *testing.T) {
	tests := []struct {
		assertions func(*testing.T, *ast.Program)
		name       string
		input      string
		stmtCount  int
	}{
		{
			name: "variable declaration and usage",
			input: `
var x: Integer := 5;
var y: Integer := 10;
x := x + y;
`,
			stmtCount: 3,
			assertions: func(t *testing.T, program *ast.Program) {
				// First statement: var x: Integer := 5;
				varDecl1, ok := program.Statements[0].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 0 is not VarDeclStatement. got=%T", program.Statements[0])
				}
				if varDecl1.Names[0].Value != "x" {
					t.Errorf("varDecl1.Names[0].Value = %q, want 'x'", varDecl1.Names[0].Value)
				}
				if varDecl1.Type == nil || varDecl1.Type.Name != "Integer" {
					t.Errorf("varDecl1.Type = %q, want 'Integer'", varDecl1.Type)
				}
				if !testIntegerLiteral(t, varDecl1.Value, 5) {
					return
				}

				// Second statement: var y: Integer := 10;
				varDecl2, ok := program.Statements[1].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 1 is not VarDeclStatement. got=%T", program.Statements[1])
				}
				if varDecl2.Names[0].Value != "y" {
					t.Errorf("varDecl2.Names[0].Value = %q, want 'y'", varDecl2.Names[0].Value)
				}
				if !testIntegerLiteral(t, varDecl2.Value, 10) {
					return
				}

				// Third statement: x := x + y;
				assign, ok := program.Statements[2].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("statement 2 is not AssignmentStatement. got=%T", program.Statements[2])
				}
				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assign.Target.Value = %q, want 'x'", assignTarget.Value)
				}
				if !testInfixExpression(t, assign.Value, "x", "+", "y") {
					return
				}
			},
		},
		{
			name: "program with function call",
			input: `
var message: String := 'Hello, World!';
PrintLn(message);
`,
			stmtCount: 2,
			assertions: func(t *testing.T, program *ast.Program) {
				// First statement: var message: String := 'Hello, World!';
				varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 0 is not VarDeclStatement. got=%T", program.Statements[0])
				}
				if varDecl.Names[0].Value != "message" {
					t.Errorf("varDecl.Names[0].Value = %q, want 'message'", varDecl.Names[0].Value)
				}
				if !testStringLiteralExpression(t, varDecl.Value, "Hello, World!") {
					return
				}

				// Second statement: PrintLn(message);
				exprStmt, ok := program.Statements[1].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("statement 1 is not ExpressionStatement. got=%T", program.Statements[1])
				}
				call, ok := exprStmt.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("statement 1 expression is not CallExpression. got=%T", exprStmt.Expression)
				}
				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
				if len(call.Arguments) != 1 {
					t.Fatalf("call has %d arguments, want 1", len(call.Arguments))
				}
				if !testIdentifier(t, call.Arguments[0], "message") {
					return
				}
			},
		},
		{
			name: "program with block statement",
			input: `
var x := 0;
begin
  x := x + 1;
  x := x * 2;
end;
`,
			stmtCount: 2,
			assertions: func(t *testing.T, program *ast.Program) {
				// First statement: var x := 0;
				varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 0 is not VarDeclStatement. got=%T", program.Statements[0])
				}
				if varDecl.Names[0].Value != "x" {
					t.Errorf("varDecl.Names[0].Value = %q, want 'x'", varDecl.Names[0].Value)
				}

				// Second statement: begin...end block
				block, ok := program.Statements[1].(*ast.BlockStatement)
				if !ok {
					t.Fatalf("statement 1 is not BlockStatement. got=%T", program.Statements[1])
				}
				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First block statement: x := x + 1;
				assign1, ok := block.Statements[0].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("block statement 0 is not AssignmentStatement. got=%T", block.Statements[0])
				}
				if !testInfixExpression(t, assign1.Value, "x", "+", 1) {
					return
				}

				// Second block statement: x := x * 2;
				assign2, ok := block.Statements[1].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("block statement 1 is not AssignmentStatement. got=%T", block.Statements[1])
				}
				if !testInfixExpression(t, assign2.Value, "x", "*", 2) {
					return
				}
			},
		},
		{
			name: "mixed statements",
			input: `
var a := 1;
var b := 2;
var sum := a + b;
PrintLn(sum);
begin
  a := a + 1;
  b := b + 1;
end;
`,
			stmtCount: 5,
			assertions: func(t *testing.T, program *ast.Program) {
				// Verify first three are variable declarations
				for i := 0; i < 3; i++ {
					if _, ok := program.Statements[i].(*ast.VarDeclStatement); !ok {
						t.Fatalf("statement %d is not VarDeclStatement. got=%T", i, program.Statements[i])
					}
				}

				// Fourth statement: PrintLn(sum);
				exprStmt, ok := program.Statements[3].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("statement 3 is not ExpressionStatement. got=%T", program.Statements[3])
				}
				if _, ok := exprStmt.Expression.(*ast.CallExpression); !ok {
					t.Fatalf("statement 3 expression is not CallExpression. got=%T", exprStmt.Expression)
				}

				// Fifth statement: begin...end block
				if _, ok := program.Statements[4].(*ast.BlockStatement); !ok {
					t.Fatalf("statement 4 is not BlockStatement. got=%T", program.Statements[4])
				}
			},
		},
		{
			name: "arithmetic expressions",
			input: `
var result := (10 + 5) * 2 - 3;
result := result / 3;
`,
			stmtCount: 2,
			assertions: func(t *testing.T, program *ast.Program) {
				// First statement: var result := (10 + 5) * 2 - 3;
				varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 0 is not VarDeclStatement. got=%T", program.Statements[0])
				}
				if varDecl.Names[0].Value != "result" {
					t.Errorf("varDecl.Names[0].Value = %q, want 'result'", varDecl.Names[0].Value)
				}
				// The value should be a complex binary expression
				if varDecl.Value == nil {
					t.Fatal("varDecl.Value is nil")
				}

				// Second statement: result := result / 3;
				assign, ok := program.Statements[1].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("statement 1 is not AssignmentStatement. got=%T", program.Statements[1])
				}
				if !testInfixExpression(t, assign.Value, "result", "/", 3) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != tt.stmtCount {
				t.Fatalf("program has %d statements, want %d", len(program.Statements), tt.stmtCount)
			}

			tt.assertions(t, program)
		})
	}
}

// TestImplicitProgramBlock tests that programs without explicit begin/end work correctly.
func TestImplicitProgramBlock(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		stmtCount int
	}{
		{
			name:      "single variable declaration",
			input:     "var x := 5;",
			stmtCount: 1,
		},
		{
			name:      "single assignment",
			input:     "x := 10;",
			stmtCount: 1,
		},
		{
			name: "multiple statements without begin/end",
			input: `
var x := 1;
var y := 2;
x := x + y;
`,
			stmtCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != tt.stmtCount {
				t.Fatalf("program has %d statements, want %d", len(program.Statements), tt.stmtCount)
			}
		})
	}
}

// TestParserErrors tests various parser error conditions to improve coverage.
func TestParserErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "invalid integer literal",
			input:         "999999999999999999999999999999;",
			expectedError: "could not parse",
		},
		{
			name:          "invalid float literal",
			input:         "99999999999999999999999999999.9e999;",
			expectedError: "could not parse",
		},
		{
			name:          "missing semicolon after var declaration",
			input:         "var x: Integer",
			expectedError: "expected next token to be SEMICOLON",
		},
		{
			name:          "unclosed parentheses",
			input:         "(3 + 5",
			expectedError: "expected next token to be RPAREN",
		},
		{
			name:          "invalid prefix operator",
			input:         "};",
			expectedError: "no prefix parse function",
		},
		{
			name:          "missing identifier in var declaration",
			input:         "var ;",
			expectedError: "expected next token to be IDENT",
		},
		{
			name:          "missing expression in if condition",
			input:         "if then x := 1;",
			expectedError: "no prefix parse function",
		},
		{
			name:          "missing then keyword in if",
			input:         "if x > 0 x := 1;",
			expectedError: "expected next token to be THEN",
		},
		{
			name:          "missing do keyword in while",
			input:         "while x < 10 x := x + 1;",
			expectedError: "expected next token to be DO",
		},
		{
			name:          "missing until keyword in repeat",
			input:         "repeat x := x + 1 x >= 10;",
			expectedError: "expected 'until' after repeat body",
		},
		{
			name:          "missing identifier in for loop",
			input:         "for := 1 to 10 do PrintLn(i);",
			expectedError: "expected next token to be IDENT",
		},
		{
			name:          "missing assign in for loop",
			input:         "for i = 1 to 10 do PrintLn(i);",
			expectedError: "expected next token to be ASSIGN",
		},
		{
			name:          "missing direction in for loop",
			input:         "for i := 1 10 do PrintLn(i);",
			expectedError: "expected 'to' or 'downto'",
		},
		{
			name:          "missing expression after case",
			input:         "case of 1: x := 1; end;",
			expectedError: "no prefix parse function",
		},
		{
			name:          "missing of keyword in case",
			input:         "case x 1: x := 1; end;",
			expectedError: "expected next token to be OF",
		},
		{
			name:          "missing colon in case branch",
			input:         "case x of 1 x := 1; end;",
			expectedError: "expected next token to be COLON",
		},
		{
			name:          "missing end keyword in case",
			input:         "case x of 1: x := 1;",
			expectedError: "expected 'end' to close case statement",
		},
		{
			name:          "missing end keyword in block",
			input:         "begin x := 1; y := 2;",
			expectedError: "expected 'end' to close block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Fatalf("expected parser errors, got none")
			}

			found := false
			for _, err := range errors {
				if contains(err, tt.expectedError) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error containing %q, got %v", tt.expectedError, errors)
			}
		})
	}
}

// TestNilLiteral tests parsing of nil literals.
func TestNilLiteral(t *testing.T) {
	input := "nil;"

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

	_, ok = stmt.Expression.(*ast.NilLiteral)
	if !ok {
		t.Fatalf("expression is not ast.NilLiteral. got=%T", stmt.Expression)
	}
}

// TestFunctionDeclarations tests parsing of function declarations.
func TestFunctionDeclarations(t *testing.T) {
	tests := []struct {
		expected func(*testing.T, ast.Statement)
		name     string
		input    string
	}{
		{
			name:  "simple function with no parameters",
			input: "function GetValue: Integer; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if fn.Name.Value != "GetValue" {
					t.Errorf("function name = %q, want %q", fn.Name.Value, "GetValue")
				}
				if fn.ReturnType == nil || fn.ReturnType.Name != "Integer" {
					t.Errorf("return type = %q, want %q", fn.ReturnType, "Integer")
				}
				if len(fn.Parameters) != 0 {
					t.Errorf("parameters count = %d, want 0", len(fn.Parameters))
				}
			},
		},
		{
			name:  "procedure with no parameters",
			input: "procedure Hello; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if fn.Name.Value != "Hello" {
					t.Errorf("function name = %q, want %q", fn.Name.Value, "Hello")
				}
				if fn.ReturnType != nil {
					t.Errorf("return type = %q, want empty string for procedure", fn.ReturnType)
				}
			},
		},
		{
			name:  "function with single parameter",
			input: "function Double(x: Integer): Integer; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if fn.Name.Value != "Double" {
					t.Errorf("function name = %q, want %q", fn.Name.Value, "Double")
				}
				if len(fn.Parameters) != 1 {
					t.Fatalf("parameters count = %d, want 1", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if param.Name.Value != "x" {
					t.Errorf("parameter name = %q, want %q", param.Name.Value, "x")
				}
				if param.Type == nil || param.Type.Name != "Integer" {
					t.Errorf("parameter type = %q, want %q", param.Type, "Integer")
				}
				if param.ByRef {
					t.Errorf("parameter ByRef = true, want false")
				}
			},
		},
		{
			name:  "function with multiple parameters",
			input: "function Add(a: Integer; b: Integer): Integer; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if len(fn.Parameters) != 2 {
					t.Fatalf("parameters count = %d, want 2", len(fn.Parameters))
				}
				if fn.Parameters[0].Name.Value != "a" {
					t.Errorf("first param name = %q, want %q", fn.Parameters[0].Name.Value, "a")
				}
				if fn.Parameters[1].Name.Value != "b" {
					t.Errorf("second param name = %q, want %q", fn.Parameters[1].Name.Value, "b")
				}
			},
		},
		{
			name:  "function with var parameter",
			input: "function Process(var data: String): Boolean; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if len(fn.Parameters) != 1 {
					t.Fatalf("parameters count = %d, want 1", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if !param.ByRef {
					t.Errorf("parameter ByRef = false, want true")
				}
				if param.Name.Value != "data" {
					t.Errorf("parameter name = %q, want %q", param.Name.Value, "data")
				}
				if param.Type == nil || param.Type.Name != "String" {
					t.Errorf("parameter type = %q, want %q", param.Type, "String")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			tt.expected(t, program.Statements[0])
		})
	}
}

// TestParameters tests parameter parsing in function declarations - Task 5.14
func TestParameters(t *testing.T) {
	tests := []struct {
		expected func(*testing.T, *ast.FunctionDecl)
		name     string
		input    string
	}{
		{
			name:  "single parameter",
			input: "function Test(x: Integer): Boolean; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if param.Name.Value != "x" {
					t.Errorf("param name = %q, want 'x'", param.Name.Value)
				}
				if param.Type == nil || param.Type.Name != "Integer" {
					t.Errorf("param type = %q, want 'Integer'", param.Type)
				}
				if param.ByRef {
					t.Error("param should not be by reference")
				}
			},
		},
		{
			name:  "multiple parameters with different types",
			input: "function Calculate(x: Integer; y: Float; name: String): Float; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 3 {
					t.Fatalf("expected 3 parameters, got %d", len(fn.Parameters))
				}

				// Check first parameter
				if fn.Parameters[0].Name.Value != "x" {
					t.Errorf("param[0] name = %q, want 'x'", fn.Parameters[0].Name.Value)
				}
				if fn.Parameters[0].Type == nil || fn.Parameters[0].Type.Name != "Integer" {
					t.Errorf("param[0] type = %q, want 'Integer'", fn.Parameters[0].Type)
				}

				// Check second parameter
				if fn.Parameters[1].Name.Value != "y" {
					t.Errorf("param[1] name = %q, want 'y'", fn.Parameters[1].Name.Value)
				}
				if fn.Parameters[1].Type == nil || fn.Parameters[1].Type.Name != "Float" {
					t.Errorf("param[1] type = %q, want 'Float'", fn.Parameters[1].Type)
				}

				// Check third parameter
				if fn.Parameters[2].Name.Value != "name" {
					t.Errorf("param[2] name = %q, want 'name'", fn.Parameters[2].Name.Value)
				}
				if fn.Parameters[2].Type == nil || fn.Parameters[2].Type.Name != "String" {
					t.Errorf("param[2] type = %q, want 'String'", fn.Parameters[2].Type)
				}
			},
		},
		{
			name:  "var parameter by reference",
			input: "procedure Swap(var a: Integer; var b: Integer); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
				}

				// Both parameters should be by reference
				if !fn.Parameters[0].ByRef {
					t.Error("param[0] should be by reference")
				}
				if !fn.Parameters[1].ByRef {
					t.Error("param[1] should be by reference")
				}

				if fn.Parameters[0].Name.Value != "a" {
					t.Errorf("param[0] name = %q, want 'a'", fn.Parameters[0].Name.Value)
				}
				if fn.Parameters[1].Name.Value != "b" {
					t.Errorf("param[1] name = %q, want 'b'", fn.Parameters[1].Name.Value)
				}
			},
		},
		{
			name:  "mixed var and value parameters",
			input: "procedure Update(var x: Integer; y: Integer; var z: String); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 3 {
					t.Fatalf("expected 3 parameters, got %d", len(fn.Parameters))
				}

				// Check ByRef flags
				if !fn.Parameters[0].ByRef {
					t.Error("param[0] should be by reference")
				}
				if fn.Parameters[1].ByRef {
					t.Error("param[1] should not be by reference")
				}
				if !fn.Parameters[2].ByRef {
					t.Error("param[2] should be by reference")
				}
			},
		},
		{
			name:  "function with no parameters",
			input: "function GetRandom: Integer; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 0 {
					t.Fatalf("expected 0 parameters, got %d", len(fn.Parameters))
				}
			},
		},
		{
			name:  "procedure with no parameters",
			input: "procedure PrintHello; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 0 {
					t.Fatalf("expected 0 parameters, got %d", len(fn.Parameters))
				}
			},
		},
		{
			name:  "lazy parameter - basic",
			input: "function Test(lazy x: Integer): Integer; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if !param.IsLazy {
					t.Error("param should be lazy")
				}
				if param.ByRef {
					t.Error("param should not be by reference")
				}
				if param.Name.Value != "x" {
					t.Errorf("param name = %q, want 'x'", param.Name.Value)
				}
				if param.Type == nil || param.Type.Name != "Integer" {
					t.Errorf("param type = %q, want 'Integer'", param.Type)
				}
			},
		},
		{
			name:  "lazy parameter - mixed with regular parameters",
			input: "procedure Log(level: Integer; lazy msg: String); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
				}

				// First parameter should be regular (not lazy, not by reference)
				if fn.Parameters[0].IsLazy {
					t.Error("param[0] should not be lazy")
				}
				if fn.Parameters[0].ByRef {
					t.Error("param[0] should not be by reference")
				}
				if fn.Parameters[0].Name.Value != "level" {
					t.Errorf("param[0] name = %q, want 'level'", fn.Parameters[0].Name.Value)
				}

				// Second parameter should be lazy
				if !fn.Parameters[1].IsLazy {
					t.Error("param[1] should be lazy")
				}
				if fn.Parameters[1].ByRef {
					t.Error("param[1] should not be by reference")
				}
				if fn.Parameters[1].Name.Value != "msg" {
					t.Errorf("param[1] name = %q, want 'msg'", fn.Parameters[1].Name.Value)
				}
			},
		},
		{
			name:  "lazy parameter - multiple lazy parameters with shared type",
			input: "function If(cond: Boolean; lazy trueVal, falseVal: Integer): Integer; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 3 {
					t.Fatalf("expected 3 parameters, got %d", len(fn.Parameters))
				}

				// First parameter (cond) should be regular
				if fn.Parameters[0].IsLazy {
					t.Error("param[0] should not be lazy")
				}
				if fn.Parameters[0].Name.Value != "cond" {
					t.Errorf("param[0] name = %q, want 'cond'", fn.Parameters[0].Name.Value)
				}

				// Second parameter (trueVal) should be lazy
				if !fn.Parameters[1].IsLazy {
					t.Error("param[1] should be lazy")
				}
				if fn.Parameters[1].Name.Value != "trueVal" {
					t.Errorf("param[1] name = %q, want 'trueVal'", fn.Parameters[1].Name.Value)
				}

				// Third parameter (falseVal) should be lazy (shares type with trueVal)
				if !fn.Parameters[2].IsLazy {
					t.Error("param[2] should be lazy")
				}
				if fn.Parameters[2].Name.Value != "falseVal" {
					t.Errorf("param[2] name = %q, want 'falseVal'", fn.Parameters[2].Name.Value)
				}

				// Both lazy parameters should have Integer type
				if fn.Parameters[1].Type == nil || fn.Parameters[1].Type.Name != "Integer" {
					t.Errorf("param[1] type = %q, want 'Integer'", fn.Parameters[1].Type)
				}
				if fn.Parameters[2].Type == nil || fn.Parameters[2].Type.Name != "Integer" {
					t.Errorf("param[2] type = %q, want 'Integer'", fn.Parameters[2].Type)
				}
			},
		},
		{
			name:  "const parameter - basic",
			input: "procedure Process(const data: array of Integer); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if !param.IsConst {
					t.Error("param should be const")
				}
				if param.ByRef {
					t.Error("param should not be by reference")
				}
				if param.IsLazy {
					t.Error("param should not be lazy")
				}
				if param.Name.Value != "data" {
					t.Errorf("param name = %q, want 'data'", param.Name.Value)
				}
			},
		},
		{
			name:  "const parameter - with string type",
			input: "procedure Display(const message: String); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if !param.IsConst {
					t.Error("param should be const")
				}
				if param.Name.Value != "message" {
					t.Errorf("param name = %q, want 'message'", param.Name.Value)
				}
				if param.Type == nil || param.Type.Name != "String" {
					t.Errorf("param type = %q, want 'String'", param.Type)
				}
			},
		},
		{
			name:  "const parameter - mixed with var and regular parameters",
			input: "procedure Update(const src: String; var dest: String; count: Integer); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 3 {
					t.Fatalf("expected 3 parameters, got %d", len(fn.Parameters))
				}

				// First parameter should be const
				if !fn.Parameters[0].IsConst {
					t.Error("param[0] should be const")
				}
				if fn.Parameters[0].ByRef {
					t.Error("param[0] should not be by reference")
				}
				if fn.Parameters[0].Name.Value != "src" {
					t.Errorf("param[0] name = %q, want 'src'", fn.Parameters[0].Name.Value)
				}

				// Second parameter should be var (by reference)
				if fn.Parameters[1].IsConst {
					t.Error("param[1] should not be const")
				}
				if !fn.Parameters[1].ByRef {
					t.Error("param[1] should be by reference")
				}
				if fn.Parameters[1].Name.Value != "dest" {
					t.Errorf("param[1] name = %q, want 'dest'", fn.Parameters[1].Name.Value)
				}

				// Third parameter should be regular (not const, not by reference)
				if fn.Parameters[2].IsConst {
					t.Error("param[2] should not be const")
				}
				if fn.Parameters[2].ByRef {
					t.Error("param[2] should not be by reference")
				}
				if fn.Parameters[2].Name.Value != "count" {
					t.Errorf("param[2] name = %q, want 'count'", fn.Parameters[2].Name.Value)
				}
			},
		},
		{
			name:  "const parameter - multiple const parameters with shared type",
			input: "procedure Compare(const a, b: Integer); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
				}

				// Both parameters should be const
				if !fn.Parameters[0].IsConst {
					t.Error("param[0] should be const")
				}
				if !fn.Parameters[1].IsConst {
					t.Error("param[1] should be const")
				}

				if fn.Parameters[0].Name.Value != "a" {
					t.Errorf("param[0] name = %q, want 'a'", fn.Parameters[0].Name.Value)
				}
				if fn.Parameters[1].Name.Value != "b" {
					t.Errorf("param[1] name = %q, want 'b'", fn.Parameters[1].Name.Value)
				}

				// Both should have Integer type
				if fn.Parameters[0].Type == nil || fn.Parameters[0].Type.Name != "Integer" {
					t.Errorf("param[0] type = %q, want 'Integer'", fn.Parameters[0].Type)
				}
				if fn.Parameters[1].Type == nil || fn.Parameters[1].Type.Name != "Integer" {
					t.Errorf("param[1] type = %q, want 'Integer'", fn.Parameters[1].Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			fn, ok := program.Statements[0].(*ast.FunctionDecl)
			if !ok {
				t.Fatalf("statement is not *ast.FunctionDecl, got %T", program.Statements[0])
			}

			tt.expected(t, fn)
		})
	}
}

// TestParameterErrors tests error cases for parameter parsing
func TestParameterErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		errorContains string
	}{
		{
			name:          "lazy and var are mutually exclusive",
			input:         "function Test(lazy var x: Integer): Integer; begin end;",
			errorContains: "parameter modifiers are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			_ = p.ParseProgram()

			if len(p.Errors()) == 0 {
				t.Fatalf("expected parser error, got none")
			}

			// Check that error message contains expected text
			found := false
			for _, err := range p.Errors() {
				if strings.Contains(err, tt.errorContains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing %q, got %v", tt.errorContains, p.Errors())
			}
		})
	}
}

// TestUserDefinedFunctionCallsWithArguments tests calling user-defined functions with arguments - Task 5.15
func TestUserDefinedFunctionCallsWithArguments(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "call user function with integer arguments",
			input: `
				function Add(a: Integer; b: Integer): Integer;
				begin
					end;

				begin
					Add(1, 2);
				end;
			`,
		},
		{
			name: "call user function with mixed argument types",
			input: `
				function Format(name: String; age: Integer): String;
				begin
				end;

				begin
					Format('John', 25);
				end;
			`,
		},
		{
			name: "call user function with expression arguments",
			input: `
				function Calculate(x: Integer; y: Integer): Integer;
				begin
				end;

				begin
					Calculate(1 + 2, 3 * 4);
				end;
			`,
		},
		{
			name: "call user function with no arguments",
			input: `
				function GetValue: Integer;
				begin
				end;

				begin
					GetValue();
				end;
			`,
		},
		{
			name: "call procedure with arguments",
			input: `
				procedure PrintValue(x: Integer);
				begin
				end;

				begin
					PrintValue(42);
				end;
			`,
		},
		{
			name: "nested function calls as arguments",
			input: `
				function Double(x: Integer): Integer;
				begin
				end;

				function Triple(x: Integer): Integer;
				begin
				end;

				begin
					Double(Triple(5));
				end;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			// The program should parse successfully
			if len(program.Statements) < 2 {
				t.Fatalf("expected at least 2 statements (function + main block), got %d", len(program.Statements))
			}

			// First statement(s) should be function declarations
			for i := 0; i < len(program.Statements)-1; i++ {
				if _, ok := program.Statements[i].(*ast.FunctionDecl); !ok {
					t.Errorf("statement %d is not *ast.FunctionDecl, got %T", i, program.Statements[i])
				}
			}

			// Last statement should be the main block containing the call
			lastStmt := program.Statements[len(program.Statements)-1]
			if _, ok := lastStmt.(*ast.BlockStatement); !ok {
				t.Errorf("last statement is not *ast.BlockStatement, got %T", lastStmt)
			}
		})
	}
}

// TestNestedFunctions tests nested function declarations - Task 5.16
// Note: DWScript may or may not support nested functions. This test documents current behavior.
func TestNestedFunctions(t *testing.T) {
	input := `
		function Outer(x: Integer): Integer;
		begin
			function Inner(y: Integer): Integer;
			begin
			end;
		end;
	`

	p := testParser(input)
	program := p.ParseProgram()

	// Check if parser supports nested functions
	// If there are parser errors, nested functions are not yet supported
	errors := p.Errors()
	if len(errors) > 0 {
		t.Skip("Nested functions not yet supported - this is expected per PLAN.md task 5.11")
		return
	}

	// If we get here, nested functions ARE supported
	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	outerFn, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("statement is not *ast.FunctionDecl, got %T", program.Statements[0])
	}

	if outerFn.Name.Value != "Outer" {
		t.Errorf("outer function name = %q, want 'Outer'", outerFn.Name.Value)
	}

	// Check if the body contains the nested function
	// This would require the AST to support nested function declarations
	// For now, we just verify the outer function parses correctly
	if outerFn.Body == nil {
		t.Error("outer function body is nil")
	}
}

// TestNewKeywordExpression tests parsing of 'new' keyword expressions
// The 'new' keyword creates a NewExpression: new T(args) -> NewExpression{ClassName: T, Arguments: args}
func TestNewKeywordExpression(t *testing.T) {
	tests := []struct {
		input    string
		typeName string
		numArgs  int
	}{
		{"new Exception('test');", "Exception", 1},
		{"new TPoint(10, 20);", "TPoint", 2},
		{"new TMyClass();", "TMyClass", 0},
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

			// new T(args) should create a NewExpression
			newExpr, ok := stmt.Expression.(*ast.NewExpression)
			if !ok {
				t.Fatalf("expression is not ast.NewExpression. got=%T", stmt.Expression)
			}

			// Check class name
			if newExpr.ClassName.Value != tt.typeName {
				t.Fatalf("wrong class name. expected=%s, got=%s", tt.typeName, newExpr.ClassName.Value)
			}

			// Check number of arguments
			if len(newExpr.Arguments) != tt.numArgs {
				t.Fatalf("wrong number of arguments. expected=%d, got=%d", tt.numArgs, len(newExpr.Arguments))
			}
		})
	}
}

// TestContextualKeywordStep tests that 'step' can be used both as a keyword in for loops
// and as a variable name in other contexts.
func TestContextualKeywordStep(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldParse bool
	}{
		{
			name: "step as variable name in declaration",
			input: `
				var step: Integer;
			`,
			shouldParse: true,
		},
		{
			name: "step as variable name with initialization",
			input: `
				var step := 0;
			`,
			shouldParse: true,
		},
		{
			name: "step in assignment statement",
			input: `
				var step := 1;
				step := 2;
			`,
			shouldParse: true,
		},
		{
			name: "step in expression",
			input: `
				var step := 1;
				var result := step + 5;
			`,
			shouldParse: true,
		},
		{
			name: "step as keyword in for loop",
			input: `
				for i := 1 to 10 step 2 do
					PrintLn(i);
			`,
			shouldParse: true,
		},
		{
			name: "step used in both contexts",
			input: `
				var step := 2;
				for i := 1 to 10 step step do
					PrintLn(i);
			`,
			shouldParse: true,
		},
		{
			name: "step in function parameter",
			input: `
				function Process(step: Integer): Integer;
				begin
					Result := step * 2;
				end;
			`,
			shouldParse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()

			if tt.shouldParse {
				checkParserErrors(t, p)
				if program == nil {
					t.Fatal("ParseProgram() returned nil")
				}
				if len(program.Statements) == 0 {
					t.Fatal("ParseProgram() returned empty statements")
				}
			} else {
				if len(p.Errors()) == 0 {
					t.Fatalf("expected parser errors, but got none")
				}
			}
		})
	}
}
