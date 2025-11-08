package interp

import (
	"fmt"
	"reflect"
)

// callDWScriptFunction invokes a DWScript function from Go context.
//
// The function:
//  1. Marshals Go arguments to DWScript values
//  2. Calls the DWScript function using existing infrastructure
//  3. Handles exceptions and converts them to Go errors
//  4. Marshals the return value back to Go
//
// This enables Go functions to accept DWScript function pointers as parameters
// and call them back, supporting patterns like:
//
//	engine.RegisterFunction("ForEach", func(items []int64, callback func(int64)) {
//	    for _, item := range items {
//	        callback(item)
//	    }
//	})
//
// DWScript code:
//
//	ForEach([1, 2, 3], lambda(x: Integer) begin PrintLn(IntToStr(x)); end);
func (i *Interpreter) callDWScriptFunction(
	funcPtr *FunctionPointerValue,
	goArgs []any,
) (any, error) {
	// 1. Marshal Go arguments to DWScript values
	dwsArgs := make([]Value, len(goArgs))
	for idx, arg := range goArgs {
		val, err := MarshalToDWS(arg)
		if err != nil {
			return nil, fmt.Errorf("callback argument %d: %w", idx, err)
		}
		dwsArgs[idx] = val
	}

	// 2. Save current interpreter state for re-entrancy
	// When Go calls back into DWScript, we need to preserve the current node
	// so that error messages show the correct location
	savedNode := i.currentNode
	defer func() { i.currentNode = savedNode }()

	// 3. Call the DWScript function
	// Use existing callLambda or callFunctionPointer infrastructure
	// These functions handle:
	// - Environment creation and scope management
	// - Recursion depth checking
	// - Call stack tracking
	// - Parameter binding
	// - Exception handling
	var result Value
	if funcPtr.Lambda != nil {
		// Lambda expression with captured closure
		result = i.callLambda(funcPtr.Lambda, funcPtr.Closure, dwsArgs, nil)
	} else if funcPtr.Function != nil {
		// Regular function pointer
		result = i.callFunctionPointer(funcPtr, dwsArgs, nil)
	} else {
		return nil, fmt.Errorf("invalid function pointer: no function or lambda")
	}

	// 4. Check for exceptions
	// If the DWScript callback raised an exception, convert it to a Go error
	if i.exception != nil {
		exceptionMsg := i.exception.Message
		exceptionClass := i.exception.ClassInfo.Name

		// Clear exception for Go context
		// The exception has been converted to an error and will be returned
		i.exception = nil

		// Convert to Go error
		return nil, fmt.Errorf("DWScript callback exception [%s]: %s", exceptionClass, exceptionMsg)
	}

	// 5. Marshal return value to Go
	// The result is a DWScript Value; we need to convert it to a Go value
	return marshalValueToGo(result)
}

// callDWScriptFunctionSafe is a wrapper around callDWScriptFunction with panic recovery.
//
// If the DWScript callback panics (e.g., nil pointer dereference in interpreter code),
// we catch it and convert to a Go error rather than crashing the Go program.
func (i *Interpreter) callDWScriptFunctionSafe(
	funcPtr *FunctionPointerValue,
	goArgs []any,
) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in DWScript callback: %v", r)
		}
	}()

	return i.callDWScriptFunction(funcPtr, goArgs)
}

// marshalValueToGo converts a DWScript Value to a Go any value.
// This is used for callback return values.
//
// Unlike MarshalToGo which requires a target type, this function
// infers the Go type from the DWScript value type.
func marshalValueToGo(val Value) (any, error) {
	switch val.Type() {
	case "INTEGER":
		return GoInt(val)
	case "FLOAT":
		return GoFloat(val)
	case "STRING":
		return GoString(val)
	case "BOOLEAN":
		return GoBool(val)
	case "NIL":
		return nil, nil
	case "ARRAY":
		// Convert array to []any
		arrayVal, ok := val.(*ArrayValue)
		if !ok {
			return nil, fmt.Errorf("expected ArrayValue, got %T", val)
		}

		result := make([]any, len(arrayVal.Elements))
		for i, elem := range arrayVal.Elements {
			goElem, err := marshalValueToGo(elem)
			if err != nil {
				return nil, fmt.Errorf("array element %d: %w", i, err)
			}
			result[i] = goElem
		}
		return result, nil

	case "RECORD":
		// Convert record to map[string]any
		recordVal, ok := val.(*RecordValue)
		if !ok {
			return nil, fmt.Errorf("expected RecordValue, got %T", val)
		}

		result := make(map[string]any)
		for key, fieldVal := range recordVal.Fields {
			goField, err := marshalValueToGo(fieldVal)
			if err != nil {
				return nil, fmt.Errorf("record field %s: %w", key, err)
			}
			result[key] = goField
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported DWScript type for callback return: %s", val.Type())
	}
}

// createGoFunctionWrapper creates a Go function that calls back into DWScript.
//
// This function uses reflection to create a Go function value with the correct
// signature that wraps a DWScript function pointer. When the Go function is called,
// it marshals the arguments to DWScript, calls the DWScript function, and returns
// the result.
//
// Parameters:
//   - funcPtr: The DWScript function pointer to wrap
//   - targetType: The Go function type to create (e.g., func(int64) int64)
//   - interp: The interpreter instance for executing the callback
//
// Returns:
//   - A Go function value that can be passed to Go functions expecting callbacks
//
// Example:
//
//	// DWScript function pointer: lambda(x: Integer): Integer
//	// Target type: func(int64) int64
//	// Returns: Go function that calls the lambda
func createGoFunctionWrapper(
	funcPtr *FunctionPointerValue,
	targetType reflect.Type,
	interp *Interpreter,
) any {
	// Validate that targetType is a function
	if targetType.Kind() != reflect.Func {
		panic(fmt.Sprintf("createGoFunctionWrapper: targetType must be a function, got %s", targetType.Kind()))
	}

	// Use reflection to create a function with the correct signature
	fn := reflect.MakeFunc(targetType, func(args []reflect.Value) []reflect.Value {
		// Convert reflect.Value arguments to []any for marshaling
		goArgs := make([]any, len(args))
		for i, arg := range args {
			goArgs[i] = arg.Interface()
		}

		// Call DWScript function with panic recovery
		result, err := interp.callDWScriptFunctionSafe(funcPtr, goArgs)

		// Handle error according to Go function signature
		if err != nil {
			// Check if function returns error as last return value
			numOut := targetType.NumOut()
			if numOut >= 2 {
				lastOut := targetType.Out(numOut - 1)
				errorType := reflect.TypeOf((*error)(nil)).Elem()
				if lastOut.Implements(errorType) {
					// Function signature includes error return: (T, error) or just (error)
					if numOut == 2 {
						// (T, error) - return zero value and error
						return []reflect.Value{
							reflect.Zero(targetType.Out(0)),
							reflect.ValueOf(err),
						}
					} else {
						// Just (error)
						return []reflect.Value{reflect.ValueOf(err)}
					}
				}
			}

			// No error return in signature - panic (Go convention)
			panic(fmt.Sprintf("callback error: %v", err))
		}

		// Marshal result(s) based on function signature
		numOut := targetType.NumOut()

		if numOut == 0 {
			// Procedure (no return value)
			return []reflect.Value{}
		}

		// Check if last return is error type
		hasErrorReturn := false
		if numOut >= 2 {
			lastOut := targetType.Out(numOut - 1)
			errorType := reflect.TypeOf((*error)(nil)).Elem()
			hasErrorReturn = lastOut.Implements(errorType)
		}

		if hasErrorReturn {
			// Function returns (T, error)
			return []reflect.Value{
				reflect.ValueOf(result),
				reflect.Zero(targetType.Out(1)), // nil error
			}
		}

		// Function returns single value
		return []reflect.Value{reflect.ValueOf(result)}
	})

	return fn.Interface()
}
