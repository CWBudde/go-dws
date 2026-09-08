package evaluator

// ExecuteFunctionPointerDirect invokes a callable from the host using an explicit
// execution context, including its environment, call stack and exception state.
func (e *Evaluator) ExecuteFunctionPointerDirect(funcPtr Value, args []Value, ctx *ExecutionContext) Value {
	return e.executeFunctionPointerDirect(funcPtr, args, ctx.CurrentNode(), ctx)
}
