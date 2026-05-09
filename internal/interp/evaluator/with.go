package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func (e *Evaluator) VisitWithStatement(node *ast.WithStatement, ctx *ExecutionContext) Value {
	if node == nil {
		return &runtime.NilValue{}
	}

	ctx.PushEnv()
	defer ctx.PopEnv()

	var result Value = &runtime.NilValue{}
	for _, decl := range node.Declarations {
		result = e.VisitVarDeclStatement(decl, ctx)
		if isError(result) {
			return result
		}
		if ctx.Exception() != nil || ctx.ControlFlow().IsActive() {
			return result
		}
	}

	result = e.Eval(node.Body, ctx)
	if result == nil {
		return &runtime.NilValue{}
	}
	return result
}
