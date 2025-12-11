package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
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
	// Convert []Value to []evaluator.Value (they implement the same interface)
	evalArgs := make([]evaluator.Value, len(args))
	copy(evalArgs, args)

	// Create callbacks for interpreter-dependent operations
	callbacks := i.createUserFunctionCallbacks()

	// Sync i.Env() with i.ctx: ExecuteUserFunction creates its function environment from
	// ctx.Env().NewEnclosedEnvironment(). We need ctx.Env() to reflect i.Env() so that
	// any bindings set up by the caller (like Self in adapter methods) are visible.
	//
	savedCtxEnv := i.ctx.Env()
	if selfVal, hasSelf := savedCtxEnv.Get("Self"); hasSelf {
		// Self is defined in the current context - preserve it in i.Env()
		if selfValue, ok := selfVal.(Value); ok {
			i.Env().Define("Self", selfValue)
		}
	}
	i.ctx.SetEnv(i.Env())
	defer func() { i.ctx.SetEnv(savedCtxEnv) }()

	// We only push to i.callStack, NOT i.ctx.GetCallStack(), because ExecuteUserFunction
	// already pushes to ctx.callStack (the context's CallStack abstraction).
	// The i.callStack field is used by Interpreter.GetCallStackArray().
	var pos *lexer.Position
	if i.evaluatorInstance.CurrentNode() != nil {
		nodePos := i.evaluatorInstance.CurrentNode().Pos()
		pos = &nodePos
	}
	frame := errors.NewStackFrame(fn.Name.Value, i.evaluatorInstance.SourceFile(), pos)
	i.callStack = append(i.callStack, frame)
	defer func() {
		if len(i.callStack) > 0 {
			i.callStack = i.callStack[:len(i.callStack)-1]
		}
	}()

	// Execute via evaluator
	result, err := i.evaluatorInstance.ExecuteUserFunction(fn, evalArgs, i.ctx, callbacks)
	if err != nil {
		// The evaluator returns a specific error message for this case
		if strings.Contains(err.Error(), "maximum recursion depth exceeded") {
			return i.raiseMaxRecursionExceeded()
		}
		return newError("%s", err.Error())
	}

	// The evaluator's ExecuteUserFunction sets exceptions via ctx.SetException(),
	// but the interpreter's exception handling relies on i.exception.
	if exc := i.ctx.Exception(); exc != nil {
		if excVal, ok := exc.(*runtime.ExceptionValue); ok {
			i.exception = excVal
		}
	}

	return result
}

// callFunctionPointer calls a function through a function pointer.
// Implement function pointer call execution.
//
// This handles both regular function pointers and method pointers.
// For method pointers, it binds the Self object before calling.
