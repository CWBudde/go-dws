package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestParseExpressionList_PreservesCursorOnRecoveredMissingArgument(t *testing.T) {
	p := testParser("(, 2)")

	args := p.parseExpressionList()

	if len(args) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(args))
	}
	if !isInvalidExpression(args[0]) {
		t.Fatalf("expected first argument to be invalid, got %T", args[0])
	}
	if !testIntegerLiteral(t, args[1], 2) {
		t.Fatalf("expected second argument to be integer 2")
	}
	if p.cursor.Current().Type != lexer.RPAREN {
		t.Fatalf("cursor should finish at RPAREN, got %s", p.cursor.Current().Type)
	}
	if len(p.Errors()) == 0 {
		t.Fatalf("expected parser errors for recovered missing argument")
	}
}

func TestParseGroupedExpression_PreservesCursorOnRecoveredArrayElement(t *testing.T) {
	p := testParser("(, 2)")

	expr := p.parseGroupedExpression()
	arrayLit, ok := expr.(*ast.ArrayLiteralExpression)
	if !ok {
		t.Fatalf("expected array literal expression, got %T", expr)
	}
	if len(arrayLit.Elements) != 2 {
		t.Fatalf("expected 2 array elements, got %d", len(arrayLit.Elements))
	}
	if !isInvalidExpression(arrayLit.Elements[0]) {
		t.Fatalf("expected first array element to be invalid, got %T", arrayLit.Elements[0])
	}
	if !testIntegerLiteral(t, arrayLit.Elements[1], 2) {
		t.Fatalf("expected second array element to be integer 2")
	}
	if p.cursor.Current().Type != lexer.RPAREN {
		t.Fatalf("cursor should finish at RPAREN, got %s", p.cursor.Current().Type)
	}
}

func TestParseIndexExpression_PreservesCursorOnRecoveredMissingIndex(t *testing.T) {
	p := testParser("[, 2]")

	expr := p.parseIndexExpression(&ast.Identifier{Value: "arr"})
	indexExpr, ok := expr.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expected index expression, got %T", expr)
	}
	if !testIntegerLiteral(t, indexExpr.Index, 2) {
		t.Fatalf("expected outer index to be integer 2")
	}
	inner, ok := indexExpr.Left.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expected nested inner index expression, got %T", indexExpr.Left)
	}
	if !isInvalidExpression(inner.Index) {
		t.Fatalf("expected inner index to be invalid, got %T", inner.Index)
	}
	if p.cursor.Current().Type != lexer.RBRACK {
		t.Fatalf("cursor should finish at RBRACK, got %s", p.cursor.Current().Type)
	}
}

func TestParseTypeOnlyParameterListAtToken_PreservesCursorOnRecoveredNestedType(t *testing.T) {
	p := testParser("array of , Integer)")

	params := p.parseTypeOnlyParameterListAtToken()

	if len(params) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(params))
	}
	if !isInvalidTypeExpression(params[0].Type) {
		t.Fatalf("expected first parameter type to be invalid, got %T", params[0].Type)
	}
	secondType, ok := params[1].Type.(*ast.TypeAnnotation)
	if !ok {
		t.Fatalf("expected second parameter type annotation, got %T", params[1].Type)
	}
	if secondType.Name != "Integer" {
		t.Fatalf("expected second parameter type Integer, got %q", secondType.Name)
	}
	if p.cursor.Current().Type != lexer.RPAREN {
		t.Fatalf("cursor should finish at RPAREN, got %s", p.cursor.Current().Type)
	}
	if len(p.Errors()) == 0 {
		t.Fatalf("expected parser errors for recovered nested type")
	}
}
