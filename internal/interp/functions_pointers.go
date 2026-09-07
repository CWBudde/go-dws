package interp

// EvalFunctionPointer invokes a callable through the evaluator using the engine's
// execution context. Host callbacks share the same dispatch as script calls.
func (i *Interpreter) EvalFunctionPointer(funcPtr Value, args []Value) Value {
	return i.evaluatorInstance.ExecuteFunctionPointerDirect(funcPtr, args, i.ctx)
}
