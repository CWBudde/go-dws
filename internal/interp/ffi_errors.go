package interp

import (
	"fmt"
	"runtime"
)

// raiseGoErrorAsException converts a Go error returned from host code into a DWScript exception.
// It builds an EHost instance that captures the original error message and type information.
func (i *Interpreter) raiseGoErrorAsException(err error) {
	if err == nil {
		return
	}

	message := err.Error()
	goType := fmt.Sprintf("%T", err)

	// Capture current DWScript call stack for diagnostics.
	callStack := make([]string, len(i.callStack))
	copy(callStack, i.callStack)

	// Look up EHost class; fall back to basic Exception if it is unavailable.
	hostClass, ok := i.classes["EHost"]
	if !ok {
		if baseClass, exists := i.classes["Exception"]; exists {
			hostClass = baseClass
		} else {
			// As a last resort, leave exception unset.
			return
		}
	}

	instance := NewObjectInstance(hostClass)

	// Ensure Message field is populated.
	instance.SetField("Message", &StringValue{Value: message})

	// Populate ExceptionClass when supported (only defined for EHost).
	if hostClass.InheritsFrom("EHost") {
		instance.SetField("ExceptionClass", &StringValue{Value: goType})
	}

	i.exception = &ExceptionValue{
		ClassInfo: hostClass,
		Instance:  instance,
		Message:   message,
		CallStack: callStack,
	}
}

// handleExternalCallResult provides a shared path for external call wrappers to marshal Go errors.
// It returns the result when no error occurred; otherwise it raises an EHost exception and returns nil.
func (i *Interpreter) handleExternalCallResult(result Value, err error) Value {
	if err == nil {
		return result
	}

	i.raiseGoErrorAsException(err)
	return &NilValue{}
}

// raiseGoPanicAsException converts a recovered panic value into an EHost exception.
// It captures the panic message, original type, DWScript call stack, and optional Go stack trace.
func (i *Interpreter) raiseGoPanicAsException(panicValue interface{}) {
	typeName := "nil"
	var message string

	switch v := panicValue.(type) {
	case error:
		typeName = fmt.Sprintf("%T", v)
		message = v.Error()
	case string:
		typeName = "string"
		message = v
	case fmt.Stringer:
		typeName = fmt.Sprintf("%T", v)
		message = v.String()
	default:
		if panicValue != nil {
			typeName = fmt.Sprintf("%T", panicValue)
			message = fmt.Sprintf("%v", panicValue)
		}
	}

	if message == "" {
		message = fmt.Sprintf("%#v", panicValue)
	}
	message = "panic: " + message

	// Optionally append Go stack trace for debugging.
	stackBuf := make([]byte, 2048)
	if n := runtime.Stack(stackBuf, false); n > 0 {
		message = message + "\n" + string(stackBuf[:n])
	}

	// Capture current DWScript call stack.
	callStack := make([]string, len(i.callStack))
	copy(callStack, i.callStack)

	// Reuse exception creation logic.
	hostClass, ok := i.classes["EHost"]
	if !ok {
		if baseClass, exists := i.classes["Exception"]; exists {
			hostClass = baseClass
		} else {
			return
		}
	}

	instance := NewObjectInstance(hostClass)
	instance.SetField("Message", &StringValue{Value: message})

	if hostClass.InheritsFrom("EHost") {
		instance.SetField("ExceptionClass", &StringValue{Value: typeName})
	}

	i.exception = &ExceptionValue{
		ClassInfo: hostClass,
		Instance:  instance,
		Message:   message,
		CallStack: callStack,
	}
}

// callExternalFunctionSafe executes a host function capturing panics and converting them into exceptions.
// The supplied callback should perform marshaling, invoke the Go function, and return the DWScript value plus error.
func (i *Interpreter) callExternalFunctionSafe(call func() (Value, error)) (result Value) {
	defer func() {
		if r := recover(); r != nil {
			i.raiseGoPanicAsException(r)
			result = &NilValue{}
		}
	}()

	res, err := call()
	result = i.handleExternalCallResult(res, err)
	return result
}
