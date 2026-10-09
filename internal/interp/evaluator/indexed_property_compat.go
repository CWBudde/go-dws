package evaluator

import "github.com/cwbudde/go-dws/pkg/ast"

func (e *Evaluator) evalIndexedCompatibilityRead(binding *ast.IndexedPropertyReadBinding, ctx *ExecutionContext) Value {
	receiver := e.Eval(binding.Read.Read.Object, ctx)
	if isError(receiver) || ctx.Exception() != nil {
		return receiver
	}
	indices, err := e.preparePropertyIndices(binding.Read.Property, binding.Indices, binding.Read.Read, ctx)
	if err != nil || ctx.Exception() != nil {
		return err
	}
	return e.evalResolvedPropertyRead(binding.Read, receiver, indices, ctx)
}
