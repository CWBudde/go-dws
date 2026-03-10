package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// executeUserFunctionViaEvaluator is a wrapper that calls the evaluator's ExecuteUserFunction.
//
// This allows gradual migration of call sites from callUserFunction to the new evaluator-based
// implementation without changing their signatures.
//
// This is necessary because:
// 1. Some call sites (adapter_methods.go) set up Self in i.Env() before calling
// 2. ExecuteUserFunction uses ctx.Env() to create the function's enclosed environment
// 3. Without sync, ctx.Env() won't see the Self binding set up in i.Env()
func (i *Interpreter) executeUserFunctionViaEvaluator(fn *ast.FunctionDecl, args []Value) Value {
	// Create callbacks for interpreter-dependent operations
	callbacks := i.createUserFunctionCallbacks()
	ctx := i.ctx
	if evalCtx := i.evaluatorInstance.CurrentContext(); evalCtx != nil {
		ctx = evalCtx
	}

	// Execute via evaluator
	result, err := i.evaluatorInstance.ExecuteUserFunction(fn, args, ctx, callbacks)
	if err != nil {
		// The evaluator returns a specific error message for this case
		if strings.Contains(err.Error(), "maximum recursion depth exceeded") {
			return i.raiseMaxRecursionExceededInContext(ctx)
		}
		return newError("%s", err.Error())
	}

	return result
}

// callFunctionPointer calls a function through a function pointer.
// Implement function pointer call execution.
//
// This handles both regular function pointers and method pointers.
// For method pointers, it binds the Self object before calling.
