package evaluator

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestAssignmentCallReading_UncheckedLocalName(t *testing.T) {
	evaluator := &Evaluator{}
	ctx := runtime.NewExecutionContext(runtime.NewEnvironment())
	declaration := &ast.FunctionDecl{
		Name:       &ast.Identifier{Value: "Local"},
		Parameters: []*ast.Parameter{{Name: &ast.Identifier{Value: "Required"}}},
	}
	evaluator.defineLocalFunction(declaration, ctx)
	result := evaluator.VisitIdentifier(&ast.Identifier{Value: "Local"}, ctx)
	if !isError(result) || !strings.Contains(result.String(), "wrong number of arguments") {
		t.Fatalf("unchecked local name must retain implicit-call arity error, got %v", result)
	}
}
