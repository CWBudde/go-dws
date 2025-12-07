package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalIdentifier looks up an identifier in the environment.
// Special handling for "Self" keyword in method contexts.
// Also handles class variable access from within methods.
func (i *Interpreter) evalIdentifier(node *ast.Identifier) Value {
	// Special case for Self keyword
	if node.Value == "Self" {
		val, ok := i.env.Get("Self")
		if !ok {
			return i.newErrorWithLocation(node, "Self used outside method context")
		}
		return val
	}

	// First, try to find in current environment
	val, ok := i.env.Get(node.Value)
	if ok {
		// Check if this is an external variable
		if extVar, isExternal := val.(*ExternalVarValue); isExternal {
			return i.newErrorWithLocation(node, "Unsupported external variable access: %s", extVar.Name)
		}

		// Check if this is a lazy parameter (LazyThunk)
		// If so, force evaluation - each access re-evaluates the expression
		if thunk, isLazy := val.(*LazyThunk); isLazy {
			return thunk.Evaluate()
		}

		// Check if this is a var parameter (ReferenceValue)
		// If so, dereference it to get the actual value
		if refVal, isRef := val.(*ReferenceValue); isRef {
			actualVal, err := refVal.Dereference()
			if err != nil {
				return i.newErrorWithLocation(node, "%s", err.Error())
			}
			return actualVal
		}

		// Return the value as-is
		return val
	}

	// Not found in environment - check if we're in a method context (Self is bound)
	selfVal, selfOk := i.env.Get("Self")
	if selfOk {
		// We're in an instance method context - check for instance fields first
		if obj, ok := AsObject(selfVal); ok {
			// Check if it's an instance field
			if fieldValue := obj.GetField(node.Value); fieldValue != nil {
				return fieldValue
			}

			// Check if it's a class variable
			classVars := obj.Class.GetClassVarsMap()
			if classVarValue, exists := classVars[node.Value]; exists {
				return classVarValue
			}

			// If we're inside a property getter/setter, skip property checks
			// to prevent infinite recursion (e.g., property Line read GetLine, where GetLine
			// references other properties). But we still allow field access above.
			if i.propContext != nil && (i.propContext.InPropertyGetter || i.propContext.InPropertySetter) {
				// Don't check properties - this prevents recursion
				// Fall through to error below
			} else {
				// Check if it's a property (properties can be accessed without Self.)
				if propInfo := obj.Class.LookupProperty(node.Value); propInfo != nil {
					// Extract the actual *types.PropertyInfo from the Impl field
					typesPropInfo, ok := propInfo.Impl.(*types.PropertyInfo)
					if !ok {
						return i.newErrorWithLocation(node, "invalid property info implementation for '%s'", node.Value)
					}

					// For field-backed properties, read the field directly to avoid recursion
					if typesPropInfo.ReadKind == types.PropAccessField {
						// Check if ReadSpec is actually a field (not a method)
						fields := obj.Class.GetFieldsMap()
						if _, isField := fields[typesPropInfo.ReadSpec]; isField {
							if fieldValue := obj.GetField(typesPropInfo.ReadSpec); fieldValue != nil {
								return fieldValue
							}
							return i.newErrorWithLocation(node, "property '%s' field '%s' not found", node.Value, typesPropInfo.ReadSpec)
						}
					}
					// For method-backed or expression-backed properties, use evalPropertyRead
					return i.evalPropertyRead(obj, typesPropInfo, node)
				}
			} // End else block for property check

			// Check if it's a method of the current class
			// This allows methods to reference other methods as method pointers
			methods := obj.Class.GetMethodsMap()
			if method, exists := methods[ident.Normalize(node.Value)]; exists {
				// In DWScript/Pascal, parameterless methods can be called without parentheses
				// When referenced as an identifier, they should be treated as implicit calls
				if len(method.Parameters) == 0 {
					// Implicit call - execute the method
					// Create a synthetic method call and evaluate it
					selfIdent := &ast.Identifier{}
					selfIdent.Token = node.Token
					selfIdent.Value = "Self"

					syntheticCall := &ast.MethodCallExpression{
						Object:    selfIdent,
						Method:    node,
						Arguments: []ast.Expression{},
					}
					return i.evalMethodCall(syntheticCall)
				}

				// Methods with parameters cannot be called without parentheses
				// This would be an error, but since semantic analysis should have caught it,
				// we'll create a method pointer for compatibility
				// Create a method pointer bound to the current object (self)
				// Build the pointer type
				paramTypes := make([]types.Type, len(method.Parameters))
				for idx, param := range method.Parameters {
					if param.Type != nil {
						paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
					}
				}
				var returnType types.Type
				if method.ReturnType != nil {
					returnType = i.getTypeFromAnnotation(method.ReturnType)
				}
				pointerType := types.NewFunctionPointerType(paramTypes, returnType)

				return NewFunctionPointerValue(method, i.env, obj, pointerType)
			}

			// Check for ClassName in instance method/constructor context
			// This handles `PrintLn(ClassName)` inside constructors
			if ident.Equal(node.Value, "ClassName") {
				return &StringValue{Value: obj.Class.GetName()}
			}
			// Check for ClassType in instance method/constructor context
			if ident.Equal(node.Value, "ClassType") {
				// Need concrete ClassInfo for ClassValue
				concreteClass, ok := obj.Class.(*ClassInfo)
				if !ok {
					return i.newErrorWithLocation(node, "invalid class type for ClassType")
				}
				return &ClassValue{ClassInfo: concreteClass}
			}
		}
	}

	// Record method context: allow field access and implicit calls
	if recVal, ok := selfVal.(*RecordValue); ok {
		if fieldVal, exists := recVal.Fields[ident.Normalize(node.Value)]; exists {
			return fieldVal
		}

		if RecordHasMethod(recVal, node.Value) {
			selfIdent := &ast.Identifier{
				Value: "Self",
			}
			selfIdent.Token = node.Token

			syntheticCall := &ast.MethodCallExpression{
				Object:    selfIdent,
				Method:    node,
				Arguments: []ast.Expression{},
			}
			return i.evalMethodCall(syntheticCall)
		}
	}

	// Check if we're in a class method context (__CurrentClass__ marker)
	currentClassVal, hasCurrentClass := i.env.Get("__CurrentClass__")
	if hasCurrentClass {
		if classInfo, ok := currentClassVal.(*ClassInfoValue); ok {
			// Check for ClassName identifier in class method context (case-insensitive)
			if ident.Equal(node.Value, "ClassName") {
				return &StringValue{Value: classInfo.ClassInfo.Name}
			}
			// Check for ClassType identifier in class method context (case-insensitive)
			if ident.Equal(node.Value, "ClassType") {
				return &ClassValue{ClassInfo: classInfo.ClassInfo}
			}

			// Check if it's a class variable
			if classVarValue, exists := classInfo.ClassInfo.ClassVars[node.Value]; exists {
				return classVarValue
			}
		}
	}

	// Before returning error, check if this is a function name
	// In DWScript, functions can be referenced as values (function pointers)
	// or called without parentheses if they have zero parameters
	// DWScript is case-insensitive, so normalize the function name
	if overloads, exists := i.functions[ident.Normalize(node.Value)]; exists && len(overloads) > 0 {
		// For function pointer or parameterless call, resolve to the appropriate overload
		var fn *ast.FunctionDecl
		if len(overloads) == 1 {
			fn = overloads[0]
		} else {
			// Multiple overloads - try to find the one with zero parameters
			for _, candidate := range overloads {
				if len(candidate.Parameters) == 0 {
					fn = candidate
					break
				}
			}
			// If no zero-param overload, default to first one (for function pointer use)
			if fn == nil {
				fn = overloads[0]
			}
		}

		// Check if function has zero parameters
		if len(fn.Parameters) == 0 {
			// Auto-invoke the parameterless function/procedure
			// This allows DWScript code like:
			//   var n := GetTickCount;
			//   if TestFlag then ...
			// to work correctly (calling the function, not creating a pointer)
			// For explicit function pointer creation, use @ syntax: var fp := @GetTickCount;
			return i.executeUserFunctionViaEvaluator(fn, []Value{})
		}

		// If function has parameters, it's being used as a value (function pointer)
		// Create a function pointer value so it can be passed to higher-order functions
		paramTypes := make([]types.Type, len(fn.Parameters))
		for idx, param := range fn.Parameters {
			if param.Type != nil {
				paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
			}
		}
		var returnType types.Type
		if fn.ReturnType != nil {
			returnType = i.getTypeFromAnnotation(fn.ReturnType)
		}
		pointerType := types.NewFunctionPointerType(paramTypes, returnType)

		return NewFunctionPointerValue(fn, i.env, nil, pointerType)
	}

	// Check if this is a parameterless built-in function
	// Built-in functions like PrintLn can be called without parentheses
	if i.isBuiltinFunction(node.Value) {
		// Call the built-in function with no arguments
		return i.callBuiltinFunction(node.Value, []Value{})
	}

	// Check if this is a class name identifier
	// Class names can be used in expressions like TObj.Create or new TObj
	// DWScript is case-insensitive, so we need to search all classes
	for className, classInfo := range i.classes {
		if ident.Equal(className, node.Value) {
			// Return a ClassValue to represent a metaclass reference
			// This allows assignments like: var meta: class of TBase; meta := TBase;
			return &ClassValue{ClassInfo: classInfo}
		}
	}

	// Still not found - return error
	return i.newErrorWithLocation(node, "undefined variable '%s'", node.Value)
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
			// Task 9.24.6: Get signature from builtin registry
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
	return NewFunctionPointerValue(function, i.env, selfObject, pointerType)
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
// Task Lambda evaluation - creates a closure capturing the current environment.
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
	closureEnv := i.env

	// Get the function pointer type from the semantic analyzer
	// The semantic analyzer already computed the type during type checking
	var pointerType *types.FunctionPointerType
	var typeAnnot *ast.TypeAnnotation
	if i.semanticInfo != nil {
		typeAnnot = i.semanticInfo.GetType(expr)
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
