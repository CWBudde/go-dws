package evaluator

import "github.com/cwbudde/go-dws/pkg/ast"

func (e *Evaluator) evalIndexedCompatibilityRead(binding *ast.IndexedPropertyReadBinding, ctx *ExecutionContext) Value {
	receiver := e.Eval(binding.Read.Read.Object, ctx)
	if isError(receiver) || ctx.Exception() != nil {
		return receiver
	}
	indices := make([]Value, len(binding.Indices))
	for i, index := range binding.Indices {
		indices[i] = e.Eval(index, ctx)
		if isError(indices[i]) || ctx.Exception() != nil {
			return indices[i]
		}
	}
	return e.evalResolvedPropertyRead(binding.Read, receiver, indices, ctx)
}
