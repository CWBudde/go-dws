package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for identifier expression AST nodes.
// Identifiers include variable references, enum literals, and name resolution.

// VisitIdentifier evaluates an identifier (variable reference).
func (e *Evaluator) VisitIdentifier(node *ast.Identifier, ctx *ExecutionContext) Value {
	// Self keyword refers to current object instance
	if node.Value == "Self" {
		val, ok := ctx.Env().Get("Self")
		if !ok {
			return e.newError(node, "Self used outside method context")
		}
		// Environment stores interface{}, cast to Value
		if selfVal, ok := val.(Value); ok {
			return selfVal
		}
		return e.newError(node, "Self has invalid type")
	}

	// Try to find identifier in current environment (variables, parameters, constants)
	// Task 3.5.70: Use direct environment access instead of adapter
	if valRaw, ok := ctx.Env().Get(node.Value); ok {
		val := valRaw.(Value)
		// Check if this is an external variable (not yet supported)
		// Task 3.5.73: Use type assertion instead of adapter
		if extVar, ok := val.(ExternalVarAccessor); ok {
			return e.newError(node, "Unsupported external variable access: %s", extVar.ExternalVarName())
		}

		// Check if this is a lazy parameter (LazyThunk)
		// If so, force evaluation - each access re-evaluates the expression
		// Task 3.5.73: Use type assertion instead of adapter
		if thunk, ok := val.(LazyEvaluator); ok {
			return thunk.Evaluate()
		}

		// Check if this is a var parameter (ReferenceValue)
		// If so, dereference it to get the actual value
		// Task 3.5.132: Use ReferenceAccessor interface directly
		if refVal, ok := val.(ReferenceAccessor); ok {
			actualVal, err := refVal.Dereference()
			if err != nil {
				return e.newError(node, "%s", err.Error())
			}
			return actualVal
		}

		// Variable found - return the value directly
		// All value types (primitives, arrays, objects, records) can be returned as-is
		return val
	}

	// Check if we're in an instance method context (Self is bound)
	// When Self is bound, identifiers can refer to instance fields, class variables,
	// properties, methods (auto-invoked if zero params), or ClassName/ClassType
	// Task 3.5.71: Use Type() check instead of adapter for IsObjectInstance
	if selfRaw, selfOk := ctx.Env().Get("Self"); selfOk {
		if selfVal, ok := selfRaw.(Value); ok && selfVal.Type() == "OBJECT" {
			// Check for instance field
			if fieldValue, found := e.adapter.GetObjectFieldValue(selfVal, node.Value); found {
				return fieldValue
			}

			// Check for class variable
			if classVarValue, found := e.adapter.GetClassVariableValue(selfVal, node.Value); found {
				return classVarValue
			}

			// Check for property - but skip if we're in a property getter/setter to prevent recursion
			propCtx := ctx.PropContext()
			if propCtx == nil || (!propCtx.InPropertyGetter && !propCtx.InPropertySetter) {
				// Task 3.5.72: Use ObjectValue interface for direct property check
				// Task 3.5.116: Use ObjectValue.ReadProperty with callback pattern
				if objVal, ok := selfVal.(ObjectValue); ok && objVal.HasProperty(node.Value) {
					propValue := objVal.ReadProperty(node.Value, func(propInfo any) Value {
						return e.adapter.ExecutePropertyRead(selfVal, propInfo, node)
					})
					return propValue
				}
			}

			// Check for method - auto-invoke if parameterless, or create method pointer
			// Task 3.5.72: Use ObjectValue interface for direct method check
			if objVal, ok := selfVal.(ObjectValue); ok && objVal.HasMethod(node.Value) {
				// Task 3.5.119: Use InvokeParameterlessMethod with callback pattern
				if result, invoked := objVal.InvokeParameterlessMethod(node.Value, func(methodDecl any) Value {
					return e.adapter.ExecuteMethodWithSelf(selfVal, methodDecl, []Value{})
				}); invoked {
					return result
				}
				// Task 3.5.120: Use CreateMethodPointer with callback pattern
				if methodPtr, created := objVal.CreateMethodPointer(node.Value, func(methodDecl any) Value {
					return e.adapter.CreateBoundMethodPointer(selfVal, methodDecl)
				}); created {
					return methodPtr
				}
				// Method not found (shouldn't reach here due to HasMethod check above)
				return e.newError(node, "method '%s' not found", node.Value)
			}

			// Check for ClassName special identifier (case-insensitive)
			if ident.Equal(node.Value, "ClassName") {
				className := e.adapter.GetClassName(selfVal)
				return &runtime.StringValue{Value: className}
			}

			// Check for ClassType special identifier (case-insensitive)
			if ident.Equal(node.Value, "ClassType") {
				return e.adapter.GetClassType(selfVal)
			}
		}
	}

	// Check if we're in a class method context (__CurrentClass__ is bound)
	// Identifiers can refer to ClassName, ClassType, or class variables
	// Task 3.5.71: Use Type() check instead of adapter for IsClassInfoValue
	if currentClassRaw, hasCurrentClass := ctx.Env().Get("__CurrentClass__"); hasCurrentClass {
		if classInfoVal, ok := currentClassRaw.(Value); ok && classInfoVal.Type() == "CLASSINFO" {
			// Check for ClassName identifier (case-insensitive)
			if ident.Equal(node.Value, "ClassName") {
				className := e.adapter.GetClassNameFromClassInfo(classInfoVal)
				return &runtime.StringValue{Value: className}
			}

			// Check for ClassType identifier (case-insensitive)
			if ident.Equal(node.Value, "ClassType") {
				return e.adapter.GetClassTypeFromClassInfo(classInfoVal)
			}

			// Check for class variable
			if classVarValue, found := e.adapter.GetClassVariableFromClassInfo(classInfoVal, node.Value); found {
				return classVarValue
			}
		}
	}

	// Check if this identifier is a user-defined function name
	// Functions are auto-invoked if they have zero parameters, or converted to function pointers if they have parameters
	// Task 3.5.67: Use direct FunctionRegistry access instead of adapter
	// Task 3.5.85: Direct evaluation without adapter EvalNode call
	funcNameLower := ident.Normalize(node.Value)
	if overloads := e.FunctionRegistry().Lookup(funcNameLower); len(overloads) > 0 {
		// Find the appropriate overload
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

		// Check if function has zero parameters - auto-invoke
		if len(fn.Parameters) == 0 {
			return e.adapter.CallUserFunction(fn, []Value{})
		}

		// Function has parameters - create function pointer
		// Task 3.5.122: Direct function pointer creation without adapter
		return createFunctionPointerFromDecl(fn, ctx.Env())
	}

	// Check if this identifier is a class name (metaclass reference)
	// Task 3.5.64: Use direct TypeRegistry access instead of adapter
	// Task 3.5.85: Direct ClassValue creation without adapter EvalNode call
	if e.typeSystem.HasClass(node.Value) {
		classVal, err := e.adapter.CreateClassValue(node.Value)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		return classVal
	}

	// Final check: check for built-in functions or return undefined error
	// Task 3.5.85: Direct built-in invocation without adapter EvalNode call
	if e.FunctionRegistry().IsBuiltin(node.Value) {
		// Parameterless built-in functions are auto-invoked
		return e.adapter.CallBuiltinFunction(node.Value, []Value{})
	}

	// Still not found - return error
	return e.newError(node, "undefined variable '%s'", node.Value)
}

// VisitEnumLiteral evaluates an enum literal (EnumType.Value).
func (e *Evaluator) VisitEnumLiteral(node *ast.EnumLiteral, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil enum literal")
	}

	valueName := node.ValueName
	val, ok := ctx.Env().Get(valueName)
	if !ok {
		return e.newError(node, "undefined enum value '%s'", valueName)
	}

	if value, ok := val.(Value); ok {
		return value
	}

	return e.newError(node, "enum value '%s' has invalid type", valueName)
}
