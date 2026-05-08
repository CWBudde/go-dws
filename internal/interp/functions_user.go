package interp

import "github.com/cwbudde/go-dws/pkg/ast"

// executeUserFunctionViaEvaluator calls the evaluator's direct user-function
// path with the interpreter's current execution context.
func (i *Interpreter) executeUserFunctionViaEvaluator(fn *ast.FunctionDecl, args []Value) Value {
	return i.evaluatorInstance.ExecuteUserFunctionDirect(fn, args, i.ctx)
}

// callFunctionPointer calls a function through a function pointer.
// Implement function pointer call execution.
//
// This handles both regular function pointers and method pointers.
// For method pointers, it binds the Self object before calling.
