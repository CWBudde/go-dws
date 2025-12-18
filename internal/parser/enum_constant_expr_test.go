package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Enum Constant Expression Tests - Helper Functions
// ============================================================================

// parseEnumDecl is a helper that parses input and returns the EnumDecl
func parseEnumDecl(t *testing.T, input string) *ast.EnumDecl {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
	}

	enumDecl, ok := program.Statements[0].(*ast.EnumDecl)
	if !ok {
		t.Fatalf("statement is not *ast.EnumDecl, got %T", program.Statements[0])
	}

	return enumDecl
}

// assertEnumName checks the enum declaration name
func assertEnumName(t *testing.T, enumDecl *ast.EnumDecl, expectedName string) {
	t.Helper()
	if enumDecl.Name.Value != expectedName {
		t.Errorf("enumDecl.Name.Value = %s, want %q", enumDecl.Name.Value, expectedName)
	}
}

// assertEnumValueCount checks the number of enum values
func assertEnumValueCount(t *testing.T, enumDecl *ast.EnumDecl, expectedCount int) {
	t.Helper()
	if len(enumDecl.Values) != expectedCount {
		t.Fatalf("enumDecl.Values should have %d elements, got %d", expectedCount, len(enumDecl.Values))
	}
}

// assertEnumValueName checks an enum value's name
func assertEnumValueName(t *testing.T, value ast.EnumValue, expectedName string) {
	t.Helper()
	if value.Name != expectedName {
		t.Errorf("enum value name = %s, want %q", value.Name, expectedName)
	}
}

// assertCallExpression checks if the expression is a CallExpression with the expected function name
func assertCallExpression(t *testing.T, expr ast.Expression, funcName string) *ast.CallExpression {
	t.Helper()
	callExpr, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression should be *ast.CallExpression, got %T", expr)
	}

	funcIdent, ok := callExpr.Function.(*ast.Identifier)
	if !ok {
		t.Fatalf("callExpr.Function should be *ast.Identifier, got %T", callExpr.Function)
	}
	if funcIdent.Value != funcName {
		t.Errorf("function name = %s, want %q", funcIdent.Value, funcName)
	}

	return callExpr
}

// assertStringLiteralArgument checks if the argument is a StringLiteral with expected value
func assertStringLiteralArgument(t *testing.T, callExpr *ast.CallExpression, argIndex int, expectedValue string) {
	t.Helper()
	if len(callExpr.Arguments) <= argIndex {
		t.Fatalf("expected at least %d arguments, got %d", argIndex+1, len(callExpr.Arguments))
	}

	strLit, ok := callExpr.Arguments[argIndex].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("argument[%d] should be *ast.StringLiteral, got %T", argIndex, callExpr.Arguments[argIndex])
	}
	if strLit.Value != expectedValue {
		t.Errorf("string literal value = %s, want %q", strLit.Value, expectedValue)
	}
}

// assertBinaryExpression checks if the expression is a BinaryExpression with the expected operator
func assertBinaryExpression(t *testing.T, expr ast.Expression, operator string) *ast.BinaryExpression {
	t.Helper()
	binExpr, ok := expr.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expression should be *ast.BinaryExpression, got %T", expr)
	}

	if binExpr.Operator != operator {
		t.Errorf("binary operator = %s, want %q", binExpr.Operator, operator)
	}

	return binExpr
}

// assertUnaryExpression checks if the expression is a UnaryExpression with the expected operator
func assertUnaryExpression(t *testing.T, expr ast.Expression, operator string) *ast.UnaryExpression {
	t.Helper()
	unaryExpr, ok := expr.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("expression should be *ast.UnaryExpression, got %T", expr)
	}

	if unaryExpr.Operator != operator {
		t.Errorf("unary operator = %s, want %q", unaryExpr.Operator, operator)
	}

	return unaryExpr
}

// assertValueExprNotNil checks that a value's expression is not nil
func assertValueExprNotNil(t *testing.T, value ast.EnumValue, context string) {
	t.Helper()
	if value.ValueExpr == nil {
		t.Fatalf("enum value %q ValueExpr should not be nil for %s", value.Name, context)
	}
}

// ============================================================================
// Enum Constant Expression Tests
// ============================================================================

// TestParseEnumWithConstantExpressions tests parsing enum values with
// constant expressions like Ord('A'), arithmetic, etc.
func TestParseEnumWithConstantExpressions(t *testing.T) {
	t.Run("Enum with Ord() function call", func(t *testing.T) {
		input := `type TEnumAlpha = (eAlpha = Ord('A'), eBeta, eGamma);`
		enumDecl := parseEnumDecl(t, input)

		assertEnumName(t, enumDecl, "TEnumAlpha")
		assertEnumValueCount(t, enumDecl, 3)

		// First value: eAlpha = Ord('A')
		firstVal := enumDecl.Values[0]
		assertEnumValueName(t, firstVal, "eAlpha")
		assertValueExprNotNil(t, firstVal, "constant expression")

		callExpr := assertCallExpression(t, firstVal.ValueExpr, "Ord")
		assertStringLiteralArgument(t, callExpr, 0, "A")

		// Second and third values should be implicit
		assertEnumValueName(t, enumDecl.Values[1], "eBeta")
		assertEnumValueName(t, enumDecl.Values[2], "eGamma")
	})

	t.Run("Enum with arithmetic expression", func(t *testing.T) {
		input := `type TEnum = (a = 1+2, b = 5*3);`
		enumDecl := parseEnumDecl(t, input)

		// First value: a = 1+2
		firstVal := enumDecl.Values[0]
		assertEnumValueName(t, firstVal, "a")
		assertValueExprNotNil(t, firstVal, "arithmetic expression")
		assertBinaryExpression(t, firstVal.ValueExpr, "+")

		// Second value: b = 5*3
		secondVal := enumDecl.Values[1]
		assertEnumValueName(t, secondVal, "b")
		assertBinaryExpression(t, secondVal.ValueExpr, "*")
	})

	t.Run("Enum with negative value expression", func(t *testing.T) {
		input := `type TEnum = (a = -5, b = -10);`
		enumDecl := parseEnumDecl(t, input)

		// First value: a = -5
		firstVal := enumDecl.Values[0]
		assertValueExprNotNil(t, firstVal, "unary expression")
		assertUnaryExpression(t, firstVal.ValueExpr, "-")
	})
}
