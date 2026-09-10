package evaluator

import (
	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// GlobalVars built-ins with var parameters
// ============================================================================
//
// TryReadGlobalVar and the four GlobalQueue read functions write their result
// into a caller-supplied variable. The registry hands built-ins already
// evaluated values, so these five are dispatched here, ahead of argument
// evaluation, where the lvalue is still available.
//
// All of them leave the target variable untouched when the global or queue
// entry is absent, matching DWScript.

// globalVarNameArgument evaluates the leading name argument shared by the
// var-parameter GlobalVars built-ins.
func (e *Evaluator) globalVarNameArgument(
	fnName string,
	arg ast.Expression,
	node ast.Node,
	ctx *ExecutionContext,
) (string, Value) {
	value := e.Eval(arg, ctx)
	if isError(value) {
		return "", value
	}
	if variant, ok := value.(*runtime.VariantValue); ok {
		value = variant.UnwrapVariant()
	}
	str, ok := value.(*runtime.StringValue)
	if !ok {
		return "", e.newError(node, "%s() expects a String name, got %s", fnName, value.Type())
	}
	return str.Value, nil
}

// storeGlobalVarResult assigns a retrieved global value to the caller's
// variable and returns True. It is only called when a value was found.
func (e *Evaluator) storeGlobalVarResult(
	fnName string,
	target ast.Expression,
	stored builtins.GlobalVarValue,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	_, assign, err := e.EvaluateLValue(target, ctx)
	if err != nil {
		return e.newError(node, "%s() second argument must be a variable: %s", fnName, err.Error())
	}
	if err := assign(stored.ToRuntimeValue()); err != nil {
		return e.newError(node, "%s() failed to update the target variable: %s", fnName, err.Error())
	}
	return &runtime.BooleanValue{Value: true}
}

// builtinTryReadGlobalVar implements TryReadGlobalVar(name, var value): Boolean.
func (e *Evaluator) builtinTryReadGlobalVar(args []ast.Expression, node ast.Node, ctx *ExecutionContext) Value {
	const fnName = "TryReadGlobalVar"
	if len(args) != 2 {
		return e.newError(node, "%s() expects exactly 2 arguments, got %d", fnName, len(args))
	}
	name, errVal := e.globalVarNameArgument(fnName, args[0], node, ctx)
	if errVal != nil {
		return errVal
	}
	stored, found := builtins.DefaultGlobalVars.Read(name)
	if !found {
		return &runtime.BooleanValue{Value: false}
	}
	return e.storeGlobalVarResult(fnName, args[1], stored, node, ctx)
}

// globalQueueReaders maps the lowercase built-in name onto the store operation
// it performs. Pull and Pop remove an entry; First and Peek do not.
var globalQueueReaders = map[string]struct {
	read func(string) (builtins.GlobalVarValue, bool)
	name string
}{
	"globalqueuepull": {read: func(n string) (builtins.GlobalVarValue, bool) {
		return builtins.DefaultGlobalVars.QueuePull(n)
	}, name: "GlobalQueuePull"},
	"globalqueuepop": {read: func(n string) (builtins.GlobalVarValue, bool) {
		return builtins.DefaultGlobalVars.QueuePop(n)
	}, name: "GlobalQueuePop"},
	"globalqueuefirst": {read: func(n string) (builtins.GlobalVarValue, bool) {
		return builtins.DefaultGlobalVars.QueueFirst(n)
	}, name: "GlobalQueueFirst"},
	"globalqueuepeek": {read: func(n string) (builtins.GlobalVarValue, bool) {
		return builtins.DefaultGlobalVars.QueuePeek(n)
	}, name: "GlobalQueuePeek"},
}

// builtinGlobalQueueRead implements the four queue readers that write their
// result into a var parameter.
func (e *Evaluator) builtinGlobalQueueRead(
	op string,
	args []ast.Expression,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	reader, ok := globalQueueReaders[op]
	if !ok {
		return e.newError(node, "unknown global queue operation %q", op)
	}
	if len(args) != 2 {
		return e.newError(node, "%s() expects exactly 2 arguments, got %d", reader.name, len(args))
	}
	name, errVal := e.globalVarNameArgument(reader.name, args[0], node, ctx)
	if errVal != nil {
		return errVal
	}
	stored, found := reader.read(name)
	if !found {
		return &runtime.BooleanValue{Value: false}
	}
	return e.storeGlobalVarResult(reader.name, args[1], stored, node, ctx)
}
