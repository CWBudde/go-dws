package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// evalCallExpression evaluates a function call expression.
func (i *Interpreter) callUserFunction(fn *ast.FunctionDecl, args []Value) Value {
	// Calculate required parameter count (parameters without defaults)
	requiredParams := 0
	for _, param := range fn.Parameters {
		if param.DefaultValue == nil {
			requiredParams++
		}
	}

	// Check argument count is within valid range
	if len(args) < requiredParams {
		return newError("wrong number of arguments: expected at least %d, got %d",
			requiredParams, len(args))
	}
	if len(args) > len(fn.Parameters) {
		return newError("wrong number of arguments: expected at most %d, got %d",
			len(fn.Parameters), len(args))
	}

	// Fill in missing optional arguments with default values
	// Evaluate default expressions in the CALLER'S environment
	if len(args) < len(fn.Parameters) {
		savedEnv := i.env // Save caller's environment
		for idx := len(args); idx < len(fn.Parameters); idx++ {
			param := fn.Parameters[idx]
			if param.DefaultValue == nil {
				// This should never happen due to requiredParams check above
				return newError("internal error: missing required parameter at index %d", idx)
			}
			// Evaluate default value in caller's environment
			defaultVal := i.Eval(param.DefaultValue)
			if isError(defaultVal) {
				return defaultVal
			}
			args = append(args, defaultVal)
		}
		i.env = savedEnv // Restore environment (should be unchanged, but be safe)
	}

	// Create a new environment for the function scope
	funcEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = funcEnv

	// Phase 3.3.3: Check recursion depth using CallStack abstraction
	if i.ctx.GetCallStack().WillOverflow() {
		i.env = savedEnv // Restore environment before raising exception
		return i.raiseMaxRecursionExceeded()
	}

	// Push function name onto call stack for stack traces
	i.pushCallStack(fn.Name.Value)
	// Ensure it's popped when function exits (even if exception occurs)
	defer i.popCallStack()

	// Bind parameters to arguments
	for idx, param := range fn.Parameters {
		arg := args[idx]

		// For var parameters, arg should already be a ReferenceValue
		// Don't apply implicit conversion to references - they need to stay as references
		if !param.ByRef {
			// Apply implicit conversion if parameter has a type and types don't match
			if param.Type != nil {
				paramTypeName := param.Type.String()
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
		}

		// Store the argument in the function's environment
		// For var parameters, this will be a ReferenceValue
		// For regular parameters, this will be the actual value
		i.env.Define(param.Name.Value, arg)
	}

	// For functions (not procedures), initialize the Result variable
	if fn.ReturnType != nil {
		// Initialize Result based on return type with appropriate defaults
		returnType := i.resolveTypeFromAnnotation(fn.ReturnType)
		var resultValue = i.getDefaultValue(returnType)

		// Check if return type is a record (overrides default)
		returnTypeName := fn.ReturnType.String()
		recordTypeKey := "__record_type_" + strings.ToLower(returnTypeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Use createRecordValue for proper nested record initialization
				resultValue = i.createRecordValue(rtv.RecordType, rtv.Methods)
			}
		}

		// Check if return type is an array (overrides default)
		// Array return types should be initialized to empty arrays, not NIL
		// This allows methods like .Add() and .High to work on the Result variable
		arrayTypeKey := "__array_type_" + strings.ToLower(returnTypeName)
		if typeVal, ok := i.env.Get(arrayTypeKey); ok {
			if atv, ok := typeVal.(*ArrayTypeValue); ok {
				resultValue = NewArrayValue(atv.ArrayType)
			}
		} else if strings.HasPrefix(returnTypeName, "array of ") || strings.HasPrefix(returnTypeName, "array[") {
			// Handle inline array return types like "array of Integer"
			// For inline array types, create the array type directly from the type name
			elementTypeName := strings.TrimPrefix(returnTypeName, "array of ")
			if elementTypeName != returnTypeName {
				// Dynamic array: "array of Integer" -> elementTypeName = "Integer"
				elementType, err := i.resolveType(elementTypeName)
				if err == nil {
					arrayType := types.NewDynamicArrayType(elementType)
					resultValue = NewArrayValue(arrayType)
				}
			}
			// TODO: Handle static inline arrays like "array[1..10] of Integer"
			// For now, those should use named types
		}

		// Task 9.1.5: Check if return type is an interface (overrides default)
		// Interface return types should be initialized to InterfaceInstance with nil object
		// This ensures proper reference counting when assigning to Result
		if interfaceInfo, ok := i.interfaces[strings.ToLower(returnTypeName)]; ok {
			resultValue = &InterfaceInstance{
				Interface: interfaceInfo,
				Object:    nil,
			}
		}

		i.env.Define("Result", resultValue)
		// Also define the function name as an alias for Result
		// In DWScript, assigning to either Result or the function name sets the return value
		// We implement this by making the function name a reference to Result
		// This ensures that assigning to either one updates the same underlying value
		i.env.Define(fn.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
	}

	// Check preconditions before executing function body
	if fn.PreConditions != nil {
		if err := i.checkPreconditions(fn.Name.Value, fn.PreConditions, i.env); err != nil {
			i.env = savedEnv
			return err
		}
		// If exception was raised during precondition checking, propagate it
		if i.exception != nil {
			i.env = savedEnv
			return &NilValue{}
		}
	}

	// Capture old values for postcondition evaluation
	oldValues := i.captureOldValues(fn, i.env)
	i.pushOldValues(oldValues)
	// Ensure old values are popped even if function exits early
	defer i.popOldValues()

	// Execute the function body
	if fn.Body == nil {
		// Function has no body (forward declaration) - this is an error
		i.env = savedEnv
		return newError("function '%s' has no body", fn.Name.Value)
	}

	i.Eval(fn.Body)

	// If an exception was raised during function execution, propagate it immediately
	if i.exception != nil {
		i.env = savedEnv
		return &NilValue{} // Return NilValue - actual value doesn't matter when exception is active
	}

	// If exit was called, clear the signal (don't propagate to caller)
	if i.ctx.ControlFlow().IsExit() {
		i.ctx.ControlFlow().Clear()
		// Exit was called, function returns immediately with current Result value
	}

	// Extract return value
	var returnValue Value
	if fn.ReturnType != nil {
		// In DWScript, you can assign to either Result or the function name to set the return value
		// We implement the function name as a ReferenceValue pointing to Result
		// So we just need to get Result's value
		resultVal, resultOk := i.env.Get("Result")

		if resultOk {
			returnValue = resultVal
		} else {
			// Result not found (shouldn't happen)
			returnValue = &NilValue{}
		}

		// Task 9.1.5: If returning an interface, increment RefCount for the caller's reference
		// This will be balanced by cleanup releasing Result after we return
		if intfInst, isIntf := returnValue.(*InterfaceInstance); isIntf {
			if intfInst.Object != nil {
				intfInst.Object.RefCount++
			}
		}

		// Apply implicit conversion if return type doesn't match
		if returnValue.Type() != "NIL" {
			expectedReturnType := fn.ReturnType.String()
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		// Procedure - no return value
		returnValue = &NilValue{}
	}

	// Check postconditions after function body executes
	// Note: old values are available via oldValuesStack during postcondition evaluation
	if fn.PostConditions != nil {
		if err := i.checkPostconditions(fn.Name.Value, fn.PostConditions, i.env); err != nil {
			i.env = savedEnv
			return err
		}
		// If exception was raised during postcondition checking, propagate it
		if i.exception != nil {
			i.env = savedEnv
			return &NilValue{}
		}
	}

	// Task 9.1.5: Clean up interface references before restoring environment
	// This releases references to interface-held objects and calls destructors if ref count reaches 0
	// Now we don't skip Result anymore - it will be properly cleaned up
	i.cleanupInterfaceReferences(funcEnv)

	// Restore the original environment
	i.env = savedEnv

	return returnValue
}

// callFunctionPointer calls a function through a function pointer.
// Implement function pointer call execution.
//
// This handles both regular function pointers and method pointers.
// For method pointers, it binds the Self object before calling.
