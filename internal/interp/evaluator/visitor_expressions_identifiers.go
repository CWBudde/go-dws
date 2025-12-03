package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/builtins"
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
	if valRaw, ok := ctx.Env().Get(node.Value); ok {
		val := valRaw.(Value)
		// Check if this is an external variable (not yet supported)
		if extVar, ok := val.(ExternalVarAccessor); ok {
			return e.newError(node, "Unsupported external variable access: %s", extVar.ExternalVarName())
		}

		// Check if this is a lazy parameter (LazyThunk)
		// If so, force evaluation - each access re-evaluates the expression
		if thunk, ok := val.(LazyEvaluator); ok {
			return thunk.Evaluate()
		}

		// Check if this is a var parameter (ReferenceValue)
		// If so, dereference it to get the actual value
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
	if selfRaw, selfOk := ctx.Env().Get("Self"); selfOk {
		if selfVal, ok := selfRaw.(Value); ok && selfVal.Type() == "OBJECT" {
			if objVal, ok := selfVal.(ObjectValue); ok {
				// Check for instance field
				if fieldValue := objVal.GetField(node.Value); fieldValue != nil {
					return fieldValue
				}

				// Check for class variable
				if classVarValue, found := objVal.GetClassVar(node.Value); found {
					return classVarValue
				}
			}

			// Check for property - but skip if we're in a property getter/setter to prevent recursion
			propCtx := ctx.PropContext()
			if propCtx == nil || (!propCtx.InPropertyGetter && !propCtx.InPropertySetter) {
				// Use ObjectValue interface for direct property check
				// Use ObjectValue.ReadProperty with callback pattern
				if objVal, ok := selfVal.(ObjectValue); ok && objVal.HasProperty(node.Value) {
					propValue := objVal.ReadProperty(node.Value, func(propInfo any) Value {
						return e.adapter.ExecutePropertyRead(selfVal, propInfo, node)
					})
					return propValue
				}
			}

			// Check for method - auto-invoke if parameterless, or create method pointer
			// Use ObjectValue interface for direct method check
			if objVal, ok := selfVal.(ObjectValue); ok && objVal.HasMethod(node.Value) {
				// Use InvokeParameterlessMethod with callback pattern
				if result, invoked := objVal.InvokeParameterlessMethod(node.Value, func(methodDecl any) Value {
					return e.adapter.ExecuteMethodWithSelf(selfVal, methodDecl, []Value{})
				}); invoked {
					return result
				}
				// Use CreateMethodPointer with callback pattern
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
				if objVal, ok := selfVal.(ObjectValue); ok {
					return &runtime.StringValue{Value: objVal.ClassName()}
				}
			}

			// Check for ClassType special identifier (case-insensitive)
			if ident.Equal(node.Value, "ClassType") {
				if objVal, ok := selfVal.(ObjectValue); ok {
					return objVal.GetClassType()
				}
			}
		}
	}

	// Check if we're in a class method context (__CurrentClass__ is bound)
	// Identifiers can refer to ClassName, ClassType, or class variables
	if currentClassRaw, hasCurrentClass := ctx.Env().Get("__CurrentClass__"); hasCurrentClass {
		if classInfoVal, ok := currentClassRaw.(Value); ok && classInfoVal.Type() == "CLASSINFO" {
			if classMetaVal, ok := classInfoVal.(ClassMetaValue); ok {
				// Check for ClassName identifier (case-insensitive)
				if ident.Equal(node.Value, "ClassName") {
					return &runtime.StringValue{Value: classMetaVal.GetClassName()}
				}

				// Check for ClassType identifier (case-insensitive)
				if ident.Equal(node.Value, "ClassType") {
					return classMetaVal.GetClassType()
				}

				// Check for class variable
				if classVarValue, found := classMetaVal.GetClassVar(node.Value); found {
					return classVarValue
				}
			}
		}
	}

	// Check if this identifier is a user-defined function name
	// Functions are auto-invoked if they have zero parameters, or converted to function pointers if they have parameters
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
			// Use evaluator-native parameterless function invocation
			return e.invokeParameterlessUserFunction(fn, node, ctx)
		}

		// Function has parameters - create function pointer
		// Direct function pointer creation without adapter
		return createFunctionPointerFromDecl(fn, ctx.Env())
	}

	// Check if this identifier is a class name (metaclass reference)
	if e.typeSystem.HasClass(node.Value) {
		classVal, err := e.typeSystem.CreateClassValue(node.Value)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		// Type assert to Value (classVal is any to avoid circular imports in types package)
		if val, ok := classVal.(Value); ok {
			return val
		}
		return e.newError(node, "internal error: ClassValueFactory returned non-Value type")
	}

	// Final check: check for built-in functions or return undefined error
	if e.FunctionRegistry().IsBuiltin(node.Value) {
		// Parameterless built-in functions are auto-invoked
		if fn, ok := builtins.DefaultRegistry.Lookup(node.Value); ok {
			return fn(e, []Value{}) // Call with empty args (parameterless auto-invoke)
		}
		// Builtin registered but not found in registry - should not happen
		return e.newError(node, "builtin function '%s' registered but not found in registry", node.Value)
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

// invokeParameterlessUserFunction invokes a parameterless user function.
//
// This is a SUBSET of callUserFunction that:
// - Uses evaluator's stack-based environment model (PushEnv/PopEnv)
// - Skips argument validation and parameter binding (no parameters)
// - Defers complex features with TODO markers
//
// Deferred to future tasks:
// - TODO(3.5.142a): Preconditions (requires contracts.go migration)
// - TODO(3.5.142b): Postconditions + old values (requires contracts.go migration)
// - TODO(3.5.142c): Interface cleanup (requires interface.go migration)
// - TODO(3.5.142d): Advanced Result init (records, interfaces, subranges)
// - TODO(3.5.142e): Function name alias to Result (requires ReferenceValue)
func (e *Evaluator) invokeParameterlessUserFunction(fn *ast.FunctionDecl, node ast.Node, ctx *ExecutionContext) Value {
	// 1. Create new enclosed environment (evaluator-native stack pattern)
	ctx.PushEnv()
	defer ctx.PopEnv()

	// 2. Check recursion depth
	if ctx.GetCallStack().WillOverflow() {
		return e.raiseMaxRecursionExceeded(node)
	}

	// 3. Push function name onto call stack for stack traces
	funcName := fn.Name.Value
	pos := node.Pos()
	if err := ctx.GetCallStack().Push(funcName, e.config.SourceFile, &pos); err != nil {
		return e.newError(node, "recursion depth exceeded calling '%s'", funcName)
	}
	defer ctx.GetCallStack().Pop()

	// 4. Initialize Result variable
	var resultValue Value
	if fn.ReturnType != nil {
		// Resolve the return type to a types.Type
		returnType, err := e.ResolveTypeFromAnnotation(fn.ReturnType)
		if err != nil {
			return e.newError(node, "cannot resolve return type '%s': %v", fn.ReturnType.String(), err)
		}

		// Get proper zero value for the type (handles all types correctly)
		resultValue = e.GetDefaultValue(returnType)

		e.DefineVar(ctx, "Result", resultValue)

		// Create function name alias to Result
		// DWScript allows `MyFunc := value` as synonym for `Result := value`
		// Implemented using ReferenceValue that points to Result variable
		env := ctx.Env()
		getter := func() (Value, error) {
			val, ok := env.Get("Result")
			if !ok {
				return &runtime.NilValue{}, fmt.Errorf("Result variable not found")
			}
			return val.(Value), nil
		}
		setter := func(val Value) error {
			env.Set("Result", val)
			return nil
		}
		funcNameAlias := runtime.NewReferenceValue("Result", getter, setter)
		e.DefineVar(ctx, funcName, funcNameAlias)
	}

	// 4. Check preconditions before function body
	if fn.PreConditions != nil {
		if err := e.checkPreconditions(funcName, fn.PreConditions, ctx); err != nil {
			return err
		}
		// If exception was raised during precondition checking, return early
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}
	}

	// 4b. Capture old values for postconditions
	// This must be called BEFORE the function body executes
	var oldValues map[string]Value
	if fn.PostConditions != nil {
		oldValues = e.captureOldValues(fn, ctx)
		// Convert map[string]Value to map[string]interface{} for ExecutionContext
		oldValuesInterface := make(map[string]interface{}, len(oldValues))
		for k, v := range oldValues {
			oldValuesInterface[k] = v
		}
		ctx.PushOldValues(oldValuesInterface)
		defer ctx.PopOldValues()
	}

	// 5. Execute function body
	if fn.Body == nil {
		return e.newError(node, "function '%s' has no body", funcName)
	}

	e.Eval(fn.Body, ctx)

	// 6. Handle exceptions during execution
	if ctx.Exception() != nil {
		return &runtime.NilValue{} // Exception active, return value doesn't matter
	}

	// 7. Handle exit statement (clear signal, don't propagate to caller)
	if ctx.ControlFlow().IsExit() {
		ctx.ControlFlow().Clear()
	}

	// 8. Extract return value
	if fn.ReturnType != nil {
		if val, ok := e.GetVar(ctx, "Result"); ok {
			resultValue = val
		} else {
			resultValue = &runtime.NilValue{}
		}

		// Implicit conversion for return type
		// Task 3.5.22h: Use evaluator's native TryImplicitConversion
		if resultValue.Type() != "NIL" {
			returnTypeName := fn.ReturnType.String()
			if converted, ok := e.TryImplicitConversion(resultValue, returnTypeName, ctx); ok {
				resultValue = converted
			}
		}
	} else {
		// Procedure - no return value
		resultValue = &runtime.NilValue{}
	}

	// 9. Check postconditions after function body
	// Old values are available via ctx.GetOldValue() during postcondition evaluation
	if fn.PostConditions != nil {
		if err := e.checkPostconditions(funcName, fn.PostConditions, ctx); err != nil {
			return err
		}
		// If exception was raised during postcondition checking, return early
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}
	}

	// 10. Clean up interface references
	// This releases interface-held and object-held references, decrementing ref counts
	// and calling destructors when reference count reaches zero.
	// Note: Cleanup happens via defer in PushEnv/PopEnv, but for now we call explicitly
	// before PopEnv to ensure cleanup happens even if function returns early.
	if e.adapter != nil {
		e.adapter.CleanupInterfaceReferences(ctx.Env())
	}

	return resultValue
}
