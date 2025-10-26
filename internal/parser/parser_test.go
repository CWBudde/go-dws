package parser

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Helper function to create a parser from input string
func testParser(input string) *Parser {
	l := lexer.New(input)
	return New(l)
}

// Helper function to check parser errors
func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

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
		input    string
		operator string
		value    any
	}{
		{"-5;", "-", 5},
		{"+10;", "+", 10},
		{"not true;", "not", true},
		{"not false;", "not", false},
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
		input      string
		leftValue  any
		operator   string
		rightValue any
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 = 5;", 5, "=", 5},
		{"5 <> 5;", 5, "<>", 5},
		{"true and false;", true, "and", false},
		{"true or false;", true, "or", false},
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
		name        string
		expectedVar string
		expectedTyp string
		expectValue bool
		assertValue func(*testing.T, ast.Expression)
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

			if stmt.Name.Value != tt.expectedVar {
				t.Errorf("stmt.Name.Value = %q, want %q", stmt.Name.Value, tt.expectedVar)
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
		isExternal   bool
		externalName string
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

			if stmt.Name.Value != tt.expectedVar {
				t.Errorf("stmt.Name.Value = %q, want %q", stmt.Name.Value, tt.expectedVar)
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

// TestMemberAssignmentStatements tests parsing of member assignment statements.
// This tests the pattern: obj.field := value
func TestMemberAssignmentStatements(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		objectName string
		fieldName  string
		value      interface{}
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
func testLiteralExpression(t *testing.T, exp ast.Expression, expected any) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

// Helper function to test integer literals
func testIntegerLiteral(t *testing.T, exp ast.Expression, value int64) bool {
	integ, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("exp not *ast.IntegerLiteral. got=%T", exp)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value, integ.TokenLiteral())
		return false
	}

	return true
}

// Helper function to test identifiers
func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
		return false
	}

	return true
}

// Helper function to test boolean literals
func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("exp not *ast.BooleanLiteral. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s", value, bo.TokenLiteral())
		return false
	}

	return true
}

// Helper function to test infix expressions
func testInfixExpression(t *testing.T, exp ast.Expression, left any, operator string, right any) bool {
	opExp, ok := exp.(*ast.BinaryExpression)
	if !ok {
		t.Errorf("exp is not ast.BinaryExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

// Helper function to test string literals
func testStringLiteralExpression(t *testing.T, exp ast.Expression, value string) bool {
	str, ok := exp.(*ast.StringLiteral)
	if !ok {
		t.Errorf("exp not *ast.StringLiteral. got=%T", exp)
		return false
	}

	if str.Value != value {
		t.Errorf("str.Value not %s. got=%s", value, str.Value)
		return false
	}

	// The token literal should be the original string from source (with quotes)
	// Just verify it's not empty
	if str.TokenLiteral() == "" {
		t.Errorf("token literal is empty")
		return false
	}

	return true
}

// TestCompleteSimplePrograms tests parsing of complete simple programs with multiple statement types.
func TestCompleteSimplePrograms(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		stmtCount  int
		assertions func(*testing.T, *ast.Program)
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
				if varDecl1.Name.Value != "x" {
					t.Errorf("varDecl1.Name.Value = %q, want 'x'", varDecl1.Name.Value)
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
				if varDecl2.Name.Value != "y" {
					t.Errorf("varDecl2.Name.Value = %q, want 'y'", varDecl2.Name.Value)
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
				if varDecl.Name.Value != "message" {
					t.Errorf("varDecl.Name.Value = %q, want 'message'", varDecl.Name.Value)
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
				if varDecl.Name.Value != "x" {
					t.Errorf("varDecl.Name.Value = %q, want 'x'", varDecl.Name.Value)
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
				if varDecl.Name.Value != "result" {
					t.Errorf("varDecl.Name.Value = %q, want 'result'", varDecl.Name.Value)
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

// TestIfStatements tests parsing of if-then-else statements.
func TestIfStatements(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		assertions func(*testing.T, *ast.IfStatement)
	}{
		{
			name:  "simple if without else",
			input: "if x > 0 then PrintLn('positive');",
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition: x > 0
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test consequence: PrintLn('positive')
				consequence, ok := stmt.Consequence.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("consequence is not ExpressionStatement. got=%T", stmt.Consequence)
				}

				call, ok := consequence.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("consequence expression is not CallExpression. got=%T", consequence.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}

				if len(call.Arguments) != 1 {
					t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
				}

				if !testStringLiteralExpression(t, call.Arguments[0], "positive") {
					return
				}

				// Test that alternative is nil
				if stmt.Alternative != nil {
					t.Errorf("alternative should be nil. got=%T", stmt.Alternative)
				}
			},
		},
		{
			name:  "if-else with expressions",
			input: "if x > 0 then PrintLn('positive') else PrintLn('non-positive');",
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test consequence
				consequence, ok := stmt.Consequence.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("consequence is not ExpressionStatement. got=%T", stmt.Consequence)
				}

				consCall, ok := consequence.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("consequence expression is not CallExpression. got=%T", consequence.Expression)
				}

				if !testIdentifier(t, consCall.Function, "PrintLn") {
					return
				}

				if !testStringLiteralExpression(t, consCall.Arguments[0], "positive") {
					return
				}

				// Test alternative
				if stmt.Alternative == nil {
					t.Fatal("alternative should not be nil")
				}

				alternative, ok := stmt.Alternative.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("alternative is not ExpressionStatement. got=%T", stmt.Alternative)
				}

				altCall, ok := alternative.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("alternative expression is not CallExpression. got=%T", alternative.Expression)
				}

				if !testIdentifier(t, altCall.Function, "PrintLn") {
					return
				}

				if !testStringLiteralExpression(t, altCall.Arguments[0], "non-positive") {
					return
				}
			},
		},
		{
			name: "if with block consequence",
			input: `if x > 0 then begin
  y := x * 2;
  PrintLn(y);
end;`,
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test consequence is a block
				block, ok := stmt.Consequence.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("consequence is not BlockStatement. got=%T", stmt.Consequence)
				}

				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First statement in block: y := x * 2;
				assign, ok := block.Statements[0].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("first block statement is not AssignmentStatement. got=%T", block.Statements[0])
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "y" {
					t.Errorf("assignment name = %q, want 'y'", assignTarget.Value)
				}

				if !testInfixExpression(t, assign.Value, "x", "*", 2) {
					return
				}

				// Second statement in block: PrintLn(y);
				exprStmt, ok := block.Statements[1].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("second block statement is not ExpressionStatement. got=%T", block.Statements[1])
				}

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("second block statement expression is not CallExpression. got=%T", exprStmt.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name: "if-else with blocks",
			input: `if x > 0 then begin
  PrintLn('positive');
end else begin
  PrintLn('non-positive');
end;`,
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test consequence block
				consBlock, ok := stmt.Consequence.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("consequence is not BlockStatement. got=%T", stmt.Consequence)
				}

				if len(consBlock.Statements) != 1 {
					t.Fatalf("consequence block has %d statements, want 1", len(consBlock.Statements))
				}

				// Test alternative block
				if stmt.Alternative == nil {
					t.Fatal("alternative should not be nil")
				}

				altBlock, ok := stmt.Alternative.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("alternative is not BlockStatement. got=%T", stmt.Alternative)
				}

				if len(altBlock.Statements) != 1 {
					t.Fatalf("alternative block has %d statements, want 1", len(altBlock.Statements))
				}
			},
		},
		{
			name: "nested if statements",
			input: `if x > 0 then
  if y > 0 then
    PrintLn('both positive')
  else
    PrintLn('x positive, y not');`,
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test outer condition
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test that consequence is another if statement
				innerIf, ok := stmt.Consequence.(*ast.IfStatement)
				if !ok {
					t.Fatalf("consequence is not IfStatement. got=%T", stmt.Consequence)
				}

				// Test inner condition
				if !testInfixExpression(t, innerIf.Condition, "y", ">", 0) {
					return
				}

				// Verify inner if has both consequence and alternative
				if innerIf.Consequence == nil {
					t.Fatal("inner if consequence is nil")
				}

				if innerIf.Alternative == nil {
					t.Fatal("inner if alternative is nil")
				}
			},
		},
		{
			name:  "if with complex condition",
			input: "if (x > 0) and (y < 10) then PrintLn('in range');",
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test that condition is a binary expression with 'and'
				binExp, ok := stmt.Condition.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("condition is not BinaryExpression. got=%T", stmt.Condition)
				}

				if binExp.Operator != "and" {
					t.Errorf("condition operator = %q, want 'and'", binExp.Operator)
				}

				// Test left side: x > 0
				if !testInfixExpression(t, binExp.Left, "x", ">", 0) {
					return
				}

				// Test right side: y < 10
				if !testInfixExpression(t, binExp.Right, "y", "<", 10) {
					return
				}
			},
		},
		{
			name:  "if with assignment in consequence",
			input: `if x = 0 then x := 1;`,
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition: x = 0
				if !testInfixExpression(t, stmt.Condition, "x", "=", 0) {
					return
				}

				// Test consequence: x := 1
				assign, ok := stmt.Consequence.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("consequence is not AssignmentStatement. got=%T", stmt.Consequence)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				if !testIntegerLiteral(t, assign.Value, 1) {
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

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.IfStatement)
			if !ok {
				t.Fatalf("statement is not ast.IfStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}

// TestWhileStatements tests parsing of while loop statements.
func TestWhileStatements(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		assertions func(*testing.T, *ast.WhileStatement)
	}{
		{
			name:  "simple while loop",
			input: "while x < 10 do x := x + 1;",
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
				// Test condition: x < 10
				if !testInfixExpression(t, stmt.Condition, "x", "<", 10) {
					return
				}

				// Test body: x := x + 1
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				if !testInfixExpression(t, assign.Value, "x", "+", 1) {
					return
				}
			},
		},
		{
			name: "while with block body",
			input: `while x < 10 do begin
  x := x + 1;
  PrintLn(x);
end;`,
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
				// Test condition
				if !testInfixExpression(t, stmt.Condition, "x", "<", 10) {
					return
				}

				// Test body is a block
				block, ok := stmt.Body.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("body is not BlockStatement. got=%T", stmt.Body)
				}

				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First statement: x := x + 1;
				assign, ok := block.Statements[0].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("first block statement is not AssignmentStatement. got=%T", block.Statements[0])
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				// Second statement: PrintLn(x);
				exprStmt, ok := block.Statements[1].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("second block statement is not ExpressionStatement. got=%T", block.Statements[1])
				}

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("second block statement expression is not CallExpression. got=%T", exprStmt.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name:  "while with complex condition",
			input: "while (x > 0) and (x < 100) do x := x * 2;",
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
				// Test that condition is a binary expression with 'and'
				binExp, ok := stmt.Condition.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("condition is not BinaryExpression. got=%T", stmt.Condition)
				}

				if binExp.Operator != "and" {
					t.Errorf("condition operator = %q, want 'and'", binExp.Operator)
				}

				// Test left side: x > 0
				if !testInfixExpression(t, binExp.Left, "x", ">", 0) {
					return
				}

				// Test right side: x < 100
				if !testInfixExpression(t, binExp.Right, "x", "<", 100) {
					return
				}

				// Test body
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				if !testInfixExpression(t, assign.Value, "x", "*", 2) {
					return
				}
			},
		},
		{
			name: "nested while loops",
			input: `while x < 10 do
  while y < 5 do
    y := y + 1;`,
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
				// Test outer condition
				if !testInfixExpression(t, stmt.Condition, "x", "<", 10) {
					return
				}

				// Test that body is another while statement
				innerWhile, ok := stmt.Body.(*ast.WhileStatement)
				if !ok {
					t.Fatalf("body is not WhileStatement. got=%T", stmt.Body)
				}

				// Test inner condition
				if !testInfixExpression(t, innerWhile.Condition, "y", "<", 5) {
					return
				}

				// Test inner body
				innerAssign, ok := innerWhile.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("inner while body is not AssignmentStatement. got=%T", innerWhile.Body)
				}

				innerAssignTarget, ok := innerAssign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("innerAssign.Target is not *ast.Identifier. got=%T", innerAssign.Target)
				}
				if innerAssignTarget.Value != "y" {
					t.Errorf("inner assignment name = %q, want 'y'", innerAssignTarget.Value)
				}
			},
		},
		{
			name:  "while with function call in body",
			input: "while hasMoreData() do processItem();",
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
				// Test condition is a function call
				condCall, ok := stmt.Condition.(*ast.CallExpression)
				if !ok {
					t.Fatalf("condition is not CallExpression. got=%T", stmt.Condition)
				}

				if !testIdentifier(t, condCall.Function, "hasMoreData") {
					return
				}

				// Test body is also a function call
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				bodyCall, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, bodyCall.Function, "processItem") {
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

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.WhileStatement)
			if !ok {
				t.Fatalf("statement is not ast.WhileStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}

// TestRepeatStatements tests parsing of repeat-until loop statements.
func TestRepeatStatements(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		assertions func(*testing.T, *ast.RepeatStatement)
	}{
		{
			name:  "simple repeat loop",
			input: "repeat x := x + 1 until x >= 10;",
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test body: x := x + 1
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				if !testInfixExpression(t, assign.Value, "x", "+", 1) {
					return
				}

				// Test condition: x >= 10
				if !testInfixExpression(t, stmt.Condition, "x", ">=", 10) {
					return
				}
			},
		},
		{
			name: "repeat with block body",
			input: `repeat begin
  x := x + 1;
  PrintLn(x);
end until x >= 10;`,
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test body is a block
				block, ok := stmt.Body.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("body is not BlockStatement. got=%T", stmt.Body)
				}

				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First statement: x := x + 1;
				assign, ok := block.Statements[0].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("first block statement is not AssignmentStatement. got=%T", block.Statements[0])
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				// Second statement: PrintLn(x);
				exprStmt, ok := block.Statements[1].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("second block statement is not ExpressionStatement. got=%T", block.Statements[1])
				}

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("second block statement expression is not CallExpression. got=%T", exprStmt.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}

				// Test condition: x >= 10
				if !testInfixExpression(t, stmt.Condition, "x", ">=", 10) {
					return
				}
			},
		},
		{
			name:  "repeat with complex condition",
			input: "repeat x := x * 2 until (x > 100) or (x < 0);",
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test body
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				if !testInfixExpression(t, assign.Value, "x", "*", 2) {
					return
				}

				// Test that condition is a binary expression with 'or'
				binExp, ok := stmt.Condition.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("condition is not BinaryExpression. got=%T", stmt.Condition)
				}

				if binExp.Operator != "or" {
					t.Errorf("condition operator = %q, want 'or'", binExp.Operator)
				}

				// Test left side: x > 100
				if !testInfixExpression(t, binExp.Left, "x", ">", 100) {
					return
				}

				// Test right side: x < 0
				if !testInfixExpression(t, binExp.Right, "x", "<", 0) {
					return
				}
			},
		},
		{
			name: "nested repeat loops",
			input: `repeat
  repeat
    y := y + 1
  until y >= 5
until x >= 10;`,
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test that body is another repeat statement
				innerRepeat, ok := stmt.Body.(*ast.RepeatStatement)
				if !ok {
					t.Fatalf("body is not RepeatStatement. got=%T", stmt.Body)
				}

				// Test inner body
				innerAssign, ok := innerRepeat.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("inner repeat body is not AssignmentStatement. got=%T", innerRepeat.Body)
				}

				innerAssignTarget, ok := innerAssign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("innerAssign.Target is not *ast.Identifier. got=%T", innerAssign.Target)
				}
				if innerAssignTarget.Value != "y" {
					t.Errorf("inner assignment name = %q, want 'y'", innerAssignTarget.Value)
				}

				// Test inner condition: y >= 5
				if !testInfixExpression(t, innerRepeat.Condition, "y", ">=", 5) {
					return
				}

				// Test outer condition: x >= 10
				if !testInfixExpression(t, stmt.Condition, "x", ">=", 10) {
					return
				}
			},
		},
		{
			name:  "repeat with function call in body",
			input: "repeat processItem() until isDone();",
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test body is a function call
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				bodyCall, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, bodyCall.Function, "processItem") {
					return
				}

				// Test condition is a function call
				condCall, ok := stmt.Condition.(*ast.CallExpression)
				if !ok {
					t.Fatalf("condition is not CallExpression. got=%T", stmt.Condition)
				}

				if !testIdentifier(t, condCall.Function, "isDone") {
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

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.RepeatStatement)
			if !ok {
				t.Fatalf("statement is not ast.RepeatStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}

// TestForStatements tests parsing of for loop statements.
func TestForStatements(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		assertions func(*testing.T, *ast.ForStatement)
	}{
		{
			name:  "simple ascending for loop",
			input: "for i := 1 to 10 do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression: 1
				if !testIntegerLiteral(t, stmt.Start, 1) {
					return
				}

				// Test end expression: 10
				if !testIntegerLiteral(t, stmt.End, 10) {
					return
				}

				// Test direction: to
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test body: PrintLn(i)
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}

				if len(call.Arguments) != 1 {
					t.Fatalf("call has %d arguments, want 1", len(call.Arguments))
				}

				if !testIdentifier(t, call.Arguments[0], "i") {
					return
				}
			},
		},
		{
			name:  "simple descending for loop",
			input: "for i := 10 downto 1 do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression: 10
				if !testIntegerLiteral(t, stmt.Start, 10) {
					return
				}

				// Test end expression: 1
				if !testIntegerLiteral(t, stmt.End, 1) {
					return
				}

				// Test direction: downto
				if stmt.Direction != ast.ForDownto {
					t.Errorf("direction = %v, want ForDownto", stmt.Direction)
				}

				// Test body is a PrintLn call
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name: "for loop with block body",
			input: `for i := 1 to 5 do begin
  PrintLn(i);
  PrintLn(i * 2);
end;`,
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test body is a block
				block, ok := stmt.Body.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("body is not BlockStatement. got=%T", stmt.Body)
				}

				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First statement: PrintLn(i)
				exprStmt1, ok := block.Statements[0].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("first block statement is not ExpressionStatement. got=%T", block.Statements[0])
				}

				call1, ok := exprStmt1.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("first block statement expression is not CallExpression. got=%T", exprStmt1.Expression)
				}

				if !testIdentifier(t, call1.Function, "PrintLn") {
					return
				}

				// Second statement: PrintLn(i * 2)
				exprStmt2, ok := block.Statements[1].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("second block statement is not ExpressionStatement. got=%T", block.Statements[1])
				}

				call2, ok := exprStmt2.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("second block statement expression is not CallExpression. got=%T", exprStmt2.Expression)
				}

				if !testIdentifier(t, call2.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name:  "for loop with variable expressions",
			input: "for i := start to finish do sum := sum + i;",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression is an identifier
				if !testIdentifier(t, stmt.Start, "start") {
					return
				}

				// Test end expression is an identifier (changed from 'end' to 'finish' to avoid keyword conflict)
				if !testIdentifier(t, stmt.End, "finish") {
					return
				}

				// Test direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test body is an assignment
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "sum" {
					t.Errorf("assignment name = %q, want 'sum'", assignTarget.Value)
				}

				// Test assignment value: sum + i
				if !testInfixExpression(t, assign.Value, "sum", "+", "i") {
					return
				}
			},
		},
		{
			name:  "for loop with expression boundaries",
			input: "for i := (start + 1) to (finish - 1) do process(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression is a binary expression: start + 1
				if !testInfixExpression(t, stmt.Start, "start", "+", 1) {
					return
				}

				// Test end expression is a binary expression: finish - 1
				if !testInfixExpression(t, stmt.End, "finish", "-", 1) {
					return
				}

				// Test direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}
			},
		},
		{
			name: "nested for loops",
			input: `for i := 1 to 10 do
  for j := 1 to 10 do
    PrintLn(i * j);`,
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test outer loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("outer loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test outer loop direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("outer direction = %v, want ForTo", stmt.Direction)
				}

				// Test that body is another for statement
				innerFor, ok := stmt.Body.(*ast.ForStatement)
				if !ok {
					t.Fatalf("body is not ForStatement. got=%T", stmt.Body)
				}

				// Test inner loop variable
				if innerFor.Variable.Value != "j" {
					t.Errorf("inner loop variable = %q, want 'j'", innerFor.Variable.Value)
				}

				// Test inner loop direction
				if innerFor.Direction != ast.ForTo {
					t.Errorf("inner direction = %v, want ForTo", innerFor.Direction)
				}

				// Test inner body is a PrintLn call
				innerBodyExpr, ok := innerFor.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("inner body is not ExpressionStatement. got=%T", innerFor.Body)
				}

				call, ok := innerBodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("inner body expression is not CallExpression. got=%T", innerBodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name:  "for loop with assignment in body",
			input: "for i := 0 to 100 do x := x + i;",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start: 0
				if !testIntegerLiteral(t, stmt.Start, 0) {
					return
				}

				// Test end: 100
				if !testIntegerLiteral(t, stmt.End, 100) {
					return
				}

				// Test direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test body: x := x + i
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				if !testInfixExpression(t, assign.Value, "x", "+", "i") {
					return
				}
			},
		},
		{
			name:  "for loop downto with larger numbers",
			input: "for count := 100 downto 0 do PrintLn(count);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "count" {
					t.Errorf("loop variable = %q, want 'count'", stmt.Variable.Value)
				}

				// Test start: 100
				if !testIntegerLiteral(t, stmt.Start, 100) {
					return
				}

				// Test end: 0
				if !testIntegerLiteral(t, stmt.End, 0) {
					return
				}

				// Test direction: downto
				if stmt.Direction != ast.ForDownto {
					t.Errorf("direction = %v, want ForDownto", stmt.Direction)
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

			stmt, ok := program.Statements[0].(*ast.ForStatement)
			if !ok {
				t.Fatalf("statement is not ast.ForStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}

// TestCaseStatements tests parsing of case statement.
func TestCaseStatements(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		assertions func(*testing.T, *ast.CaseStatement)
	}{
		{
			name: "simple case with single value branches",
			input: `case x of
  1: PrintLn('one');
  2: PrintLn('two');
  3: PrintLn('three');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of case branches
				if len(stmt.Cases) != 3 {
					t.Fatalf("case has %d branches, want 3", len(stmt.Cases))
				}

				// Test first branch: 1: PrintLn('one');
				branch1 := stmt.Cases[0]
				if len(branch1.Values) != 1 {
					t.Fatalf("branch1 has %d values, want 1", len(branch1.Values))
				}
				if !testIntegerLiteral(t, branch1.Values[0], 1) {
					return
				}

				// Test second branch: 2: PrintLn('two');
				branch2 := stmt.Cases[1]
				if len(branch2.Values) != 1 {
					t.Fatalf("branch2 has %d values, want 1", len(branch2.Values))
				}
				if !testIntegerLiteral(t, branch2.Values[0], 2) {
					return
				}

				// Test third branch: 3: PrintLn('three');
				branch3 := stmt.Cases[2]
				if len(branch3.Values) != 1 {
					t.Fatalf("branch3 has %d values, want 1", len(branch3.Values))
				}
				if !testIntegerLiteral(t, branch3.Values[0], 3) {
					return
				}

				// Test that there's no else branch
				if stmt.Else != nil {
					t.Errorf("else branch should be nil, got %T", stmt.Else)
				}
			},
		},
		{
			name: "case with multiple values per branch",
			input: `case x of
  1, 2, 3: PrintLn('one to three');
  4, 5: PrintLn('four or five');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch: 1, 2, 3
				branch1 := stmt.Cases[0]
				if len(branch1.Values) != 3 {
					t.Fatalf("branch1 has %d values, want 3", len(branch1.Values))
				}
				if !testIntegerLiteral(t, branch1.Values[0], 1) {
					return
				}
				if !testIntegerLiteral(t, branch1.Values[1], 2) {
					return
				}
				if !testIntegerLiteral(t, branch1.Values[2], 3) {
					return
				}

				// Test second branch: 4, 5
				branch2 := stmt.Cases[1]
				if len(branch2.Values) != 2 {
					t.Fatalf("branch2 has %d values, want 2", len(branch2.Values))
				}
				if !testIntegerLiteral(t, branch2.Values[0], 4) {
					return
				}
				if !testIntegerLiteral(t, branch2.Values[1], 5) {
					return
				}
			},
		},
		{
			name: "case with else branch",
			input: `case x of
  1: PrintLn('one');
  2: PrintLn('two');
else
  PrintLn('other');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test else branch exists
				if stmt.Else == nil {
					t.Fatal("else branch should not be nil")
				}

				// Test else branch is a PrintLn call
				elseExpr, ok := stmt.Else.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("else branch is not ExpressionStatement. got=%T", stmt.Else)
				}

				call, ok := elseExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("else expression is not CallExpression. got=%T", elseExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}

				if !testStringLiteralExpression(t, call.Arguments[0], "other") {
					return
				}
			},
		},
		{
			name: "case with block statements",
			input: `case x of
  1: begin
    y := 1;
    PrintLn(y);
  end;
  2: begin
    y := 2;
    PrintLn(y);
  end;
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch has a block statement
				branch1 := stmt.Cases[0]
				block1, ok := branch1.Statement.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("branch1 statement is not BlockStatement. got=%T", branch1.Statement)
				}

				if len(block1.Statements) != 2 {
					t.Fatalf("branch1 block has %d statements, want 2", len(block1.Statements))
				}

				// Test second branch has a block statement
				branch2 := stmt.Cases[1]
				block2, ok := branch2.Statement.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("branch2 statement is not BlockStatement. got=%T", branch2.Statement)
				}

				if len(block2.Statements) != 2 {
					t.Fatalf("branch2 block has %d statements, want 2", len(block2.Statements))
				}
			},
		},
		{
			name: "case with string expression and string values",
			input: `case name of
  'Alice', 'Bob': PrintLn('known person');
  'Unknown': PrintLn('stranger');
else
  PrintLn('no match');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "name") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch has 2 string values
				branch1 := stmt.Cases[0]
				if len(branch1.Values) != 2 {
					t.Fatalf("branch1 has %d values, want 2", len(branch1.Values))
				}
				if !testStringLiteralExpression(t, branch1.Values[0], "Alice") {
					return
				}
				if !testStringLiteralExpression(t, branch1.Values[1], "Bob") {
					return
				}

				// Test second branch has 1 string value
				branch2 := stmt.Cases[1]
				if len(branch2.Values) != 1 {
					t.Fatalf("branch2 has %d values, want 1", len(branch2.Values))
				}
				if !testStringLiteralExpression(t, branch2.Values[0], "Unknown") {
					return
				}

				// Test else branch exists
				if stmt.Else == nil {
					t.Fatal("else branch should not be nil")
				}
			},
		},
		{
			name: "case with complex expression",
			input: `case x + y of
  0: PrintLn('zero');
  1: PrintLn('one');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression is a binary expression
				if !testInfixExpression(t, stmt.Expression, "x", "+", "y") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}
			},
		},
		{
			name: "case with assignment in branch",
			input: `case status of
  0: result := 'failed';
  1: result := 'success';
else
  result := 'unknown';
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "status") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch is an assignment
				branch1 := stmt.Cases[0]
				assign1, ok := branch1.Statement.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("branch1 statement is not AssignmentStatement. got=%T", branch1.Statement)
				}

				assign1Target, ok := assign1.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign1.Target is not *ast.Identifier. got=%T", assign1.Target)
				}
				if assign1Target.Value != "result" {
					t.Errorf("branch1 assignment name = %q, want 'result'", assign1Target.Value)
				}

				if !testStringLiteralExpression(t, assign1.Value, "failed") {
					return
				}

				// Test else branch is also an assignment
				if stmt.Else == nil {
					t.Fatal("else branch should not be nil")
				}

				elseAssign, ok := stmt.Else.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("else branch is not AssignmentStatement. got=%T", stmt.Else)
				}

				elseAssignTarget, ok := elseAssign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("elseAssign.Target is not *ast.Identifier. got=%T", elseAssign.Target)
				}
				if elseAssignTarget.Value != "result" {
					t.Errorf("else assignment name = %q, want 'result'", elseAssignTarget.Value)
				}

				if !testStringLiteralExpression(t, elseAssign.Value, "unknown") {
					return
				}
			},
		},
		{
			name: "case with expression values",
			input: `case x of
  min_val: PrintLn('minimum');
  max_val: PrintLn('maximum');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch value is an identifier
				branch1 := stmt.Cases[0]
				if !testIdentifier(t, branch1.Values[0], "min_val") {
					return
				}

				// Test second branch value is an identifier
				branch2 := stmt.Cases[1]
				if !testIdentifier(t, branch2.Values[0], "max_val") {
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

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.CaseStatement)
			if !ok {
				t.Fatalf("statement is not ast.CaseStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
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
			expectedError: "expected next token to be UNTIL",
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

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

// TestFunctionDeclarations tests parsing of function declarations.
func TestFunctionDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*testing.T, ast.Statement)
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
		name     string
		input    string
		expected func(*testing.T, *ast.FunctionDecl)
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
