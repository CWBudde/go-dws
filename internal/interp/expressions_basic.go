package interp

import (
	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalAddressOfExpression evaluates an address-of expression (@Function).
// Creates a FunctionPointerValue, capturing Self for method pointers.
func (i *Interpreter) evalAddressOfExpression(expr *ast.AddressOfExpression) Value {
	switch operand := expr.Operator.(type) {
	case *ast.Identifier:
		// Regular function/procedure pointer: @FunctionName
		return i.evalFunctionPointer(operand.Value, nil, expr)

	case *ast.MemberAccessExpression:
		// Method pointer: @object.MethodName
		// First evaluate the object
		objectVal := i.Eval(operand.Object)
		if isError(objectVal) {
			return objectVal
		}

		// Get the method name
		methodName := operand.Member.Value

		// Create method pointer with the object as Self
		return i.evalFunctionPointer(methodName, objectVal, expr)

	default:
		return i.newRuntimeError(expr, "address-of operator requires function or method name, got %T", operand)
	}
}

// evalFunctionPointer creates a function pointer value for the named function.
// If selfObject is non-nil, creates a method pointer.
func (i *Interpreter) evalFunctionPointer(name string, selfObject Value, _ ast.Node) Value {
	var function *ast.FunctionDecl

	// If selfObject is provided, this is a method pointer - look up in the class
	if selfObject != nil {
		// Extract the object instance
		obj, ok := AsObject(selfObject)
		if !ok {
			return i.newRuntimeError(nil, "method pointer requires an object instance, got %s", selfObject.Type())
		}

		// Look up the method in the class hierarchy
		function = obj.Class.LookupMethod(name)
		if function == nil {
			return i.newUndefinedError(nil, "undefined method: %s.%s", obj.Class.GetName(), name)
		}
	} else {
		// Look up the function in the function registry
		// DWScript is case-insensitive, so normalize the function name
		overloads, exists := i.functions[ident.Normalize(name)]
		if !exists || len(overloads) == 0 {
			// Built-in function pointer
			if _, ok := builtins.DefaultRegistry.Lookup(name); ok {
				var pointerType *types.FunctionPointerType
				if sig, found := builtins.DefaultRegistry.GetSignature(name); found {
					// Create function pointer type from builtin signature
					var returnType types.Type
					if sig.ReturnType != nil && sig.ReturnType != types.VOID {
						returnType = sig.ReturnType
					}
					pointerType = types.NewFunctionPointerType(sig.ParamTypes, returnType)
				}
				return NewBuiltinFunctionPointerValue(name, pointerType)
			}
			return i.newUndefinedError(nil, "undefined function or procedure: %s", name)
		}

		// For overloaded functions, use the first overload
		// Note: Function pointers cannot represent overload sets, only single functions
		function = overloads[0]
	}

	// Get the function pointer type from the semantic analyzer's type information
	// For now, create a basic function pointer type from the function signature
	var pointerType *types.FunctionPointerType

	// Build parameter types
	paramTypes := make([]types.Type, len(function.Parameters))
	for idx, param := range function.Parameters {
		if param.Type != nil {
			paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
		} else {
			paramTypes[idx] = &types.IntegerType{} // Default fallback
		}
	}

	// Get return type
	var returnType types.Type
	if function.ReturnType != nil {
		returnType = i.getTypeFromAnnotation(function.ReturnType)
	}

	// Create the function pointer type
	// If this is a method pointer, create a MethodPointerType
	if selfObject != nil {
		methodPtr := types.NewMethodPointerType(paramTypes, returnType)
		// Cast to FunctionPointerType for storage
		pointerType = &methodPtr.FunctionPointerType
	} else if returnType != nil {
		pointerType = types.NewFunctionPointerType(paramTypes, returnType)
	} else {
		pointerType = types.NewProcedurePointerType(paramTypes)
	}

	// Create and return the function pointer value
	return NewFunctionPointerValue(function, i.Env(), selfObject, pointerType)
}

// getTypeFromAnnotation converts a type annotation to a types.Type
// This is a helper to extract type information from AST
func (i *Interpreter) getTypeFromAnnotation(typeExpr ast.TypeExpression) types.Type {
	if typeExpr == nil {
		return nil
	}

	// Get the type name from the type expression
	typeName := typeExpr.String()
	return i.getTypeByName(typeName)
}

// getTypeByName looks up a type by name
func (i *Interpreter) getTypeByName(name string) types.Type {
	switch name {
	case "Integer":
		return &types.IntegerType{}
	case "Float":
		return &types.FloatType{}
	case "String":
		return &types.StringType{}
	case "Boolean":
		return &types.BooleanType{}
	default:
		// Try to find in type registry (for custom types)
		// For now, return integer as placeholder
		return &types.IntegerType{}
	}
}
