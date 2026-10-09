package evaluator

import (
	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalDefaultNamespaceCall dispatches the result unit's own procedures without
// repeating lexical, user-routine or implicit Self lookup for their names.
func (e *Evaluator) evalDefaultNamespaceCall(member *ast.Identifier, expressions []ast.Expression, node ast.Node, ctx *ExecutionContext) Value {
	if !isDefaultOutputMember(member.Value) {
		return e.newError(node, "Unknown name \"Default.%s\"", member.Value)
	}
	args := make([]Value, len(expressions))
	for i, expression := range expressions {
		args[i] = e.evalValueContextExpression(expression, ctx)
		if isError(args[i]) {
			return args[i]
		}
		if ctx.Exception() != nil {
			return e.nilValue()
		}
	}
	if err := e.coerceBuiltinArgsToSignature(member, expressions, args, ctx); err != nil {
		return err
	}
	if ctx.Exception() != nil {
		return e.nilValue()
	}
	function, _ := builtins.DefaultRegistry.Lookup(member.Value)
	return function(e.builtinContext(ctx), args)
}

func isDefaultOutputMember(name string) bool {
	return ident.Equal(name, "Print") || ident.Equal(name, "PrintLn")
}

func (e *Evaluator) defaultNamespacePointer(member *ast.Identifier, node ast.Node) Value {
	if !isDefaultOutputMember(member.Value) {
		return e.newError(node, "Unknown name \"Default.%s\"", member.Value)
	}
	pointer := types.NewProcedurePointerType([]types.Type{types.VARIANT})
	if info := e.SemanticInfo(); info != nil {
		if selected, ok := info.GetResolvedType(member).(*types.FunctionPointerType); ok {
			pointer = selected
		}
	}
	return &runtime.FunctionPointerValue{BuiltinName: member.Value, PointerType: pointer}
}
