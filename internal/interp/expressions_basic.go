package interp

import (
	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalIdentifier looks up an identifier in the environment.
// Phase 3.9: Delegates to evaluator's canonical implementation.
func (i *Interpreter) evalIdentifier(node *ast.Identifier) Value {
	return i.evaluatorInstance.VisitIdentifier(node, i.ctx)
}

// evalAddressOfExpression evaluates an address-of expression (@Function).
// Implement address-of operator evaluation to create function pointers.
//
// This creates a FunctionPointerValue that wraps the target function/procedure.
// For methods, it also captures the Self object to create a method pointer.
func (i *Interpreter) evalAddressOfExpression(expr *ast.AddressOfExpression) Value {
	// The operator should be an identifier (function/procedure name) or member access (for methods)
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

// evalLambdaExpression evaluates a lambda expression and creates a closure.
//
// A lambda expression evaluates to a function pointer value that captures the
// environment where it was created (closure). The closure allows the lambda to
// access variables from outer scopes when it's eventually called.
//
// Examples:
//   - var double := lambda(x: Integer): Integer begin Result := x * 2; end;
//   - var add := lambda(a, b: Integer) => a + b;  // shorthand syntax
//   - Capturing outer variable: var factor := 10;
//     var multiply := lambda(x: Integer) => x * factor;
func (i *Interpreter) evalLambdaExpression(expr *ast.LambdaExpression) Value {
	// The current environment becomes the closure environment
	// This captures all variables accessible at the point where the lambda is defined
	closureEnv := i.Env()

	// Get the function pointer type from the semantic analyzer
	// The semantic analyzer already computed the type during type checking
	var pointerType *types.FunctionPointerType
	var typeAnnot *ast.TypeAnnotation
	if i.evaluatorInstance.SemanticInfo() != nil {
		typeAnnot = i.evaluatorInstance.SemanticInfo().GetType(expr)
	}
	if typeAnnot != nil {
		// Extract the type information from the annotation
		// The semantic analyzer stored a FunctionPointerType in typeAnnot
		pointerType = i.getFunctionPointerTypeFromAnnotation(typeAnnot)
	} else {
		// Fallback: construct type from lambda signature
		// Build parameter types
		paramTypes := make([]types.Type, len(expr.Parameters))
		for idx, param := range expr.Parameters {
			if param.Type != nil {
				paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
			} else {
				paramTypes[idx] = &types.IntegerType{} // Default fallback
			}
		}

		// Get return type
		var returnType types.Type
		if expr.ReturnType != nil {
			returnType = i.getTypeFromAnnotation(expr.ReturnType)
		}

		// Create the function pointer type
		if returnType != nil {
			pointerType = types.NewFunctionPointerType(paramTypes, returnType)
		} else {
			pointerType = types.NewProcedurePointerType(paramTypes)
		}
	}

	// Create and return a lambda value (closure)
	// The lambda captures the current environment (closureEnv) which includes
	// all variables from outer scopes listed in expr.CapturedVars
	return NewLambdaValue(expr, closureEnv, pointerType)
}

// getFunctionPointerTypeFromAnnotation extracts FunctionPointerType from a type annotation.
// Helper for lambda evaluation to get the type computed by semantic analysis.
func (i *Interpreter) getFunctionPointerTypeFromAnnotation(typeAnnotation *ast.TypeAnnotation) *types.FunctionPointerType {
	if typeAnnotation == nil {
		return nil
	}

	// For lambda expressions, the semantic analyzer stores a FunctionPointerType
	// in the Type field. We need to reconstruct it from the annotation.
	// For now, we'll use the type name to determine if it's a function pointer

	// TODO: This is a simplified implementation. In a full implementation,
	// the semantic analyzer should provide a way to get the computed type directly.
	// For now, return nil to trigger the fallback in evalLambdaExpression

	return nil
}
