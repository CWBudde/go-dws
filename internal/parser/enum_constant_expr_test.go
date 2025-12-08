package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Enum Constant Expression Tests
// ============================================================================

// TestParseEnumWithConstantExpressions tests parsing enum values with
// constant expressions like Ord('A'), arithmetic, etc.
func TestParseEnumWithConstantExpressions(t *testing.T) {
	t.Run("Enum with Ord() function call", func(t *testing.T) {
		input := `type TEnumAlpha = (eAlpha = Ord('A'), eBeta, eGamma);`

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

		if enumDecl.Name.Value != "TEnumAlpha" {
			t.Errorf("enumDecl.Name.Value = %s, want 'TEnumAlpha'", enumDecl.Name.Value)
		}

		if len(enumDecl.Values) != 3 {
			t.Fatalf("enumDecl.Values should have 3 elements, got %d", len(enumDecl.Values))
		}

		// First value: eAlpha = Ord('A')
		firstVal := enumDecl.Values[0]
		if firstVal.Name != "eAlpha" {
			t.Errorf("enumDecl.Values[0].Name = %s, want 'eAlpha'", firstVal.Name)
		}

		// Check that ValueExpr is set (not Value)
		if firstVal.ValueExpr == nil {
			t.Fatal("enumDecl.Values[0].ValueExpr should not be nil for constant expression")
		}

		// Verify it's a CallExpression for Ord('A')
		callExpr, ok := firstVal.ValueExpr.(*ast.CallExpression)
		if !ok {
			t.Fatalf("ValueExpr should be *ast.CallExpression, got %T", firstVal.ValueExpr)
		}

		// Check function name is "Ord"
		funcIdent, ok := callExpr.Function.(*ast.Identifier)
		if !ok {
			t.Fatalf("callExpr.Function should be *ast.Identifier, got %T", callExpr.Function)
		}
		if funcIdent.Value != "Ord" {
			t.Errorf("function name = %s, want 'Ord'", funcIdent.Value)
		}

		// Check argument is string literal 'A'
		if len(callExpr.Arguments) != 1 {
			t.Fatalf("Ord() should have 1 argument, got %d", len(callExpr.Arguments))
		}
		strLit, ok := callExpr.Arguments[0].(*ast.StringLiteral)
		if !ok {
			t.Fatalf("Ord() argument should be *ast.StringLiteral, got %T", callExpr.Arguments[0])
		}
		if strLit.Value != "A" {
			t.Errorf("string literal value = %s, want 'A'", strLit.Value)
		}

		// Second and third values should be implicit (nil ValueExpr and Value)
		if enumDecl.Values[1].Name != "eBeta" {
			t.Errorf("enumDecl.Values[1].Name = %s, want 'eBeta'", enumDecl.Values[1].Name)
		}
		if enumDecl.Values[2].Name != "eGamma" {
			t.Errorf("enumDecl.Values[2].Name = %s, want 'eGamma'", enumDecl.Values[2].Name)
		}
	})

	t.Run("Enum with arithmetic expression", func(t *testing.T) {
		input := `type TEnum = (a = 1+2, b = 5*3);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		enumDecl := program.Statements[0].(*ast.EnumDecl)

		// First value: a = 1+2
		firstVal := enumDecl.Values[0]
		if firstVal.Name != "a" {
			t.Errorf("enumDecl.Values[0].Name = %s, want 'a'", firstVal.Name)
		}

		if firstVal.ValueExpr == nil {
			t.Fatal("enumDecl.Values[0].ValueExpr should not be nil for arithmetic expression")
		}

		binExpr, ok := firstVal.ValueExpr.(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("ValueExpr should be *ast.BinaryExpression, got %T", firstVal.ValueExpr)
		}

		if binExpr.Operator != "+" {
			t.Errorf("binary operator = %s, want '+'", binExpr.Operator)
		}

		// Second value: b = 5*3
		secondVal := enumDecl.Values[1]
		if secondVal.Name != "b" {
			t.Errorf("enumDecl.Values[1].Name = %s, want 'b'", secondVal.Name)
		}

		binExpr2, ok := secondVal.ValueExpr.(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("ValueExpr should be *ast.BinaryExpression, got %T", secondVal.ValueExpr)
		}

		if binExpr2.Operator != "*" {
			t.Errorf("binary operator = %s, want '*'", binExpr2.Operator)
		}
	})

	t.Run("Enum with negative value expression", func(t *testing.T) {
		input := `type TEnum = (a = -5, b = -10);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		enumDecl := program.Statements[0].(*ast.EnumDecl)

		// First value: a = -5
		firstVal := enumDecl.Values[0]
		if firstVal.ValueExpr == nil {
			t.Fatal("enumDecl.Values[0].ValueExpr should not be nil for unary expression")
		}

		unaryExpr, ok := firstVal.ValueExpr.(*ast.UnaryExpression)
		if !ok {
			t.Fatalf("ValueExpr should be *ast.UnaryExpression, got %T", firstVal.ValueExpr)
		}

		if unaryExpr.Operator != "-" {
			t.Errorf("unary operator = %s, want '-'", unaryExpr.Operator)
		}
	})
}
