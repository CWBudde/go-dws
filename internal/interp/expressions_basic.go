package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
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
			if classVarValue, exists := obj.Class.ClassVars[node.Value]; exists {
				return classVarValue
			}

			// If we're inside a property getter/setter, skip property checks
			// to prevent infinite recursion (e.g., property Line read GetLine, where GetLine
			// references other properties). But we still allow field access above.
			if i.propContext != nil && (i.propContext.inPropertyGetter || i.propContext.inPropertySetter) {
				// Don't check properties - this prevents recursion
				// Fall through to error below
			} else {
				// Check if it's a property (properties can be accessed without Self.)
				if propInfo := obj.Class.lookupProperty(node.Value); propInfo != nil {
					// For field-backed properties, read the field directly to avoid recursion
					if propInfo.ReadKind == types.PropAccessField {
						// Check if ReadSpec is actually a field (not a method)
						if _, isField := obj.Class.Fields[propInfo.ReadSpec]; isField {
							if fieldValue := obj.GetField(propInfo.ReadSpec); fieldValue != nil {
								return fieldValue
							}
							return i.newErrorWithLocation(node, "property '%s' field '%s' not found", node.Value, propInfo.ReadSpec)
						}
					}
					// For method-backed or expression-backed properties, use evalPropertyRead
					return i.evalPropertyRead(obj, propInfo, node)
				}
			} // End else block for property check

			// Check if it's a method of the current class
			// This allows methods to reference other methods as method pointers
			if method, exists := obj.Class.Methods[strings.ToLower(node.Value)]; exists {
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
			if strings.EqualFold(node.Value, "ClassName") {
				return &StringValue{Value: obj.Class.Name}
			}
			// Check for ClassType in instance method/constructor context
			if strings.EqualFold(node.Value, "ClassType") {
				return &ClassValue{ClassInfo: obj.Class}
			}
		}
	}

	// Check if we're in a class method context (__CurrentClass__ marker)
	currentClassVal, hasCurrentClass := i.env.Get("__CurrentClass__")
	if hasCurrentClass {
		if classInfo, ok := currentClassVal.(*ClassInfoValue); ok {
			// Check for ClassName identifier in class method context (case-insensitive)
			if strings.EqualFold(node.Value, "ClassName") {
				return &StringValue{Value: classInfo.ClassInfo.Name}
			}
			// Check for ClassType identifier in class method context (case-insensitive)
			if strings.EqualFold(node.Value, "ClassType") {
				return &ClassValue{ClassInfo: classInfo.ClassInfo}
			}

			// Check if it's a class variable
			if classVarValue, exists := classInfo.ClassInfo.ClassVars[node.Value]; exists {
				return classVarValue
			}
		}
	}

	// Before returning error, check if this is a parameterless function/procedure call
	// In DWScript, you can call parameterless procedures without parentheses: "Test;" instead of "Test();"
	if overloads, exists := i.functions[node.Value]; exists && len(overloads) > 0 {
		// For parameterless call or function pointer, resolve to the no-arg overload
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
			return i.callUserFunction(fn, []Value{})
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
		if strings.EqualFold(className, node.Value) {
			// Return a ClassValue to represent a metaclass reference
			// This allows assignments like: var meta: class of TBase; meta := TBase;
			return &ClassValue{ClassInfo: classInfo}
		}
	}

	// Still not found - return error
	return i.newErrorWithLocation(node, "undefined variable '%s'", node.Value)
}

// evalUnaryExpression evaluates a unary expression.
func (i *Interpreter) evalUnaryExpression(expr *ast.UnaryExpression) Value {
	right := i.Eval(expr.Right)
	if isError(right) {
		return right
	}

	if result, ok := i.tryUnaryOperator(expr.Operator, right, expr); ok {
		return result
	}

	switch expr.Operator {
	case "-":
		return i.evalMinusUnaryOp(right)
	case "+":
		return i.evalPlusUnaryOp(right)
	case "not":
		return i.evalNotUnaryOp(right)
	default:
		return newError("unknown operator: %s%s", expr.Operator, right.Type())
	}
}

func (i *Interpreter) tryUnaryOperator(operator string, operand Value, node ast.Node) (Value, bool) {
	operands := []Value{operand}
	operandTypes := []string{valueTypeKey(operand)}

	if obj, ok := operand.(*ObjectInstance); ok {
		if entry, found := obj.Class.lookupOperator(operator, operandTypes); found {
			return i.invokeRuntimeOperator(entry, operands, node), true
		}
	}

	if entry, found := i.globalOperators.lookup(operator, operandTypes); found {
		return i.invokeRuntimeOperator(entry, operands, node), true
	}

	return nil, false
}

// evalMinusUnaryOp evaluates the unary minus operator.
func (i *Interpreter) evalMinusUnaryOp(right Value) Value {
	switch v := right.(type) {
	case *IntegerValue:
		return &IntegerValue{Value: -v.Value}
	case *FloatValue:
		return &FloatValue{Value: -v.Value}
	default:
		return i.newErrorWithLocation(i.currentNode, "expected integer or float for unary minus, got %s", right.Type())
	}
}

// evalPlusUnaryOp evaluates the unary plus operator.
func (i *Interpreter) evalPlusUnaryOp(right Value) Value {
	switch right.(type) {
	case *IntegerValue, *FloatValue:
		return right
	default:
		return i.newErrorWithLocation(i.currentNode, "expected integer or float for unary plus, got %s", right.Type())
	}
}

// evalNotUnaryOp evaluates the not operator.
func (i *Interpreter) evalNotUnaryOp(right Value) Value {
	// Unwrap Variant values to get the actual runtime value
	actualValue := unwrapVariant(right)

	// Handle boolean NOT
	if boolVal, ok := actualValue.(*BooleanValue); ok {
		return &BooleanValue{Value: !boolVal.Value}
	}

	// Handle bitwise NOT for integers
	if intVal, ok := actualValue.(*IntegerValue); ok {
		return &IntegerValue{Value: ^intVal.Value}
	}

	return i.newErrorWithLocation(i.currentNode, "NOT operator requires Boolean or Integer operand, got %s", actualValue.Type())
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
		return newError("address-of operator requires function or method name, got %T", operand)
	}
}

// evalFunctionPointer creates a function pointer value for the named function.
// If selfObject is non-nil, creates a method pointer.
func (i *Interpreter) evalFunctionPointer(name string, selfObject Value, _ ast.Node) Value {
	// Look up the function in the function registry
	overloads, exists := i.functions[name]
	if !exists || len(overloads) == 0 {
		return newError("undefined function or procedure: %s", name)
	}

	// For overloaded functions, use the first overload
	// Note: Function pointers cannot represent overload sets, only single functions
	function := overloads[0]

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
func (i *Interpreter) getTypeFromAnnotation(typeAnnotation *ast.TypeAnnotation) types.Type {
	if typeAnnotation == nil {
		return nil
	}

	// TypeAnnotation has a Name field that contains the type name
	typeName := typeAnnotation.Name
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
	if expr.Type != nil {
		// Extract the type information from the annotation
		// The semantic analyzer stored a FunctionPointerType in expr.Type
		pointerType = i.getFunctionPointerTypeFromAnnotation(expr.Type)
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
