package interp

import "github.com/cwbudde/go-dws/pkg/ast"

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
	return i.evaluatorInstance.ExecuteUserFunctionDirect(fn, args, i.ctx)
}

// callFunctionPointer calls a function through a function pointer.
// Implement function pointer call execution.
//
// This handles both regular function pointers and method pointers.
// For method pointers, it binds the Self object before calling.
