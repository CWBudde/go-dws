package interp

// CallExternalFunction invokes an external Go function with evaluator-prepared arguments.
func (i *Interpreter) CallExternalFunction(funcName string, args []Value) Value {
	if i.externalFunctions() == nil {
		return i.newErrorWithLocation(i.ctx.CurrentNode(), "external function registry not initialized")
	}

	registry := i.externalFunctions()
	extFunc, ok := registry.Get(funcName)
	if !ok {
		return i.newErrorWithLocation(i.ctx.CurrentNode(), "external function '%s' not found", funcName)
	}

	return i.callExternalFunction(extFunc, args)
}

// callExternalFunction calls an external Go function registered via FFI
// It uses the existing FFI error handling infrastructure to safely call the Go function
// and convert any errors or panics to DWScript exceptions.
func (i *Interpreter) callExternalFunction(extFunc *ExternalFunctionValue, args []Value) Value {
	// Set interpreter reference for callback support
	// This allows the FFI wrapper to create Go callbacks that call back into DWScript
	extFunc.Wrapper.SetInterpreter(i)

	// Use the existing callExternalFunctionSafe wrapper which handles panics
	// and converts them to EHost exceptions (from ffi_errors.go)
	return i.callExternalFunctionSafe(func() (Value, error) {
		// Call the wrapped Go function
		return extFunc.Wrapper.Call(args)
	})
}
