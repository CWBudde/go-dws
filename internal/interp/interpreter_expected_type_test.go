package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestEvalWithExpectedType_UsesEvaluatorArrayContext(t *testing.T) {
	interp := newTestInterpreter()

	arrayLit := &ast.ArrayLiteralExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: token.Token{}},
		},
		Elements: []ast.Expression{
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{Token: token.Token{}},
				},
				Value: 1,
			},
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{Token: token.Token{}},
				},
				Value: 2,
			},
		},
	}

	expectedType := types.NewStaticArrayType(types.INTEGER, 1, 2)

	value := interp.EvalWithExpectedType(arrayLit, expectedType)
	arrayVal, ok := value.(*ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", value)
	}
	if arrayVal.ArrayType == nil {
		t.Fatalf("expected array type metadata")
	}
	if !arrayVal.ArrayType.IsStatic() {
		t.Fatalf("expected static array type, got %s", arrayVal.ArrayType.String())
	}
	if got := arrayVal.ArrayType.Size(); got != 2 {
		t.Fatalf("expected static array size 2, got %d", got)
	}
	if len(arrayVal.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arrayVal.Elements))
	}
	first, ok := arrayVal.Elements[0].(*IntegerValue)
	if !ok || first.Value != 1 {
		t.Fatalf("expected first element 1, got %#v", arrayVal.Elements[0])
	}
	second, ok := arrayVal.Elements[1].(*IntegerValue)
	if !ok || second.Value != 2 {
		t.Fatalf("expected second element 2, got %#v", arrayVal.Elements[1])
	}
	if interp.ctx.ArrayTypeContext() != nil {
		t.Fatalf("array type context leaked after evaluation")
	}
}
