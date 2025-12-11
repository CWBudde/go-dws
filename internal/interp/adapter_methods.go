package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ===== Method and Qualified Call Methods =====

// CallQualifiedOrConstructor calls a unit-qualified function or class constructor.
func (i *Interpreter) CallQualifiedOrConstructor(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression) evaluator.Value {
	// This method encapsulates the complex logic from evalCallExpression lines 122-201

	// Check if the left side is a unit identifier (for qualified access: UnitName.FunctionName)
	if unitIdent, ok := memberAccess.Object.(*ast.Identifier); ok {
		// This could be a unit-qualified call: UnitName.FunctionName()
		if i.evaluatorInstance.UnitRegistry() != nil {
			if _, exists := i.evaluatorInstance.UnitRegistry().GetUnit(unitIdent.Value); exists {
				// Resolve the qualified function
				fn, err := i.ResolveQualifiedFunction(unitIdent.Value, memberAccess.Member.Value)
				if err == nil {
					// Prepare arguments - lazy parameters get LazyThunks, var parameters get References
					args := make([]Value, len(callExpr.Arguments))
					for idx, arg := range callExpr.Arguments {
						// Check parameter flags
						isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
						isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

						if isLazy {
							// For lazy parameters, reuse existing thunks to avoid self-recursive wrapping
							args[idx] = i.wrapLazyArgument(arg)
						} else if isByRef {
							// For var parameters, create a reference
							if argIdent, ok := arg.(*ast.Identifier); ok {
								if val, exists := i.Env().Get(argIdent.Value); exists {
									if refVal, isRef := val.(*ReferenceValue); isRef {
										args[idx] = refVal // Pass through existing reference
									} else {
										args[idx] = &ReferenceValue{Env: i.Env(), VarName: argIdent.Value}
									}
								} else {
									args[idx] = &ReferenceValue{Env: i.Env(), VarName: argIdent.Value}
								}
							} else {
								return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
							}
						} else {
							// For regular parameters, evaluate immediately
							val := i.Eval(arg)
							if isError(val) {
								return val
							}
							args[idx] = val
						}
					}
					return i.executeUserFunctionViaEvaluator(fn, args)
				}
				// Function not found in unit
				return i.newErrorWithLocation(callExpr, "function '%s' not found in unit '%s'", memberAccess.Member.Value, unitIdent.Value)
			}
		}

		// Check if this is a class constructor call (TClass.Create(...))
		var classInfo *ClassInfo
		for className, class := range i.classes {
			if ident.Equal(className, unitIdent.Value) {
				classInfo = class
				break
			}
		}
		if classInfo != nil {
			// This is a class constructor/method call - convert to MethodCallExpression
			mc := &ast.MethodCallExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: callExpr.Token,
					},
				},
				Object:    unitIdent,
				Method:    memberAccess.Member,
				Arguments: callExpr.Arguments,
			}
			return i.evalMethodCall(mc)
		}
	}

	return i.newErrorWithLocation(callExpr, "cannot call member expression that is not a method or unit-qualified function")
}

// CallMethod calls a method on objects (record, interface, or object instance).
// This is the primary adapter method for method dispatch from the evaluator.
func (i *Interpreter) CallMethod(obj evaluator.Value, methodName string, args []evaluator.Value, node ast.Node) evaluator.Value {
	// Convert to internal types
	internalObj := obj.(Value)
	internalArgs := convertEvaluatorArgs(args)

	// Handle CLASS_INFO values (class method calls)
	// Pattern: ClassInfoValue.Method() where Self is already a ClassInfoValue
	if classInfoVal, ok := internalObj.(*ClassInfoValue); ok {
		classInfo := classInfoVal.ClassInfo
		if classInfo == nil {
			return newError("ClassInfoValue has no class information")
		}

		// Look up class method (case-insensitive)
		classMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, methodName, true)
		if len(classMethodOverloads) == 0 {
			return newError("class method '%s' not found in class '%s'", methodName, classInfo.Name)
		}

		// Create a synthetic MethodCallExpression for overload resolution
		// We need this to support overloaded methods with different parameter types
		argExprs := make([]ast.Expression, len(internalArgs))
		for idx := range internalArgs {
			// We don't have actual expressions, so we pass nil
			// The overload resolver will fall back to argument count matching
			argExprs[idx] = nil
		}

		// Simple case: single overload or select by argument count
		var classMethod *ast.FunctionDecl
		if len(classMethodOverloads) == 1 {
			classMethod = classMethodOverloads[0]
		} else {
			// Find matching overload by argument count
			for _, m := range classMethodOverloads {
				if len(m.Parameters) == len(internalArgs) {
					classMethod = m
					break
				}
			}
			if classMethod == nil {
				return newError("no matching overload for class method '%s' in class '%s' with %d arguments",
					methodName, classInfo.Name, len(internalArgs))
			}
		}

		// Execute class method with Self bound to ClassInfoValue
		// Phase 3.1.4: unified scope management
		defer i.PushScope()()

		// Check recursion depth
		if i.ctx.GetCallStack().WillOverflow() {
			return i.raiseMaxRecursionExceeded()
		}

		// Push to call stack
		fullMethodName := classInfo.Name + "." + methodName
		i.pushCallStack(fullMethodName)
		defer i.popCallStack()

		// Bind Self to ClassInfoValue for class methods
		i.Env().Define("Self", classInfoVal)
		i.Env().Define("__CurrentClass__", classInfoVal)

		// Add class constants
		i.bindClassConstantsToEnv(classInfo)

		// Bind parameters with implicit conversion
		for idx, param := range classMethod.Parameters {
			arg := internalArgs[idx]
			if param.Type != nil {
				paramTypeName := param.Type.String()
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
			i.Env().Define(param.Name.Value, arg)
		}

		// Initialize Result for functions
		if classMethod.ReturnType != nil {
			returnType := i.resolveTypeFromAnnotation(classMethod.ReturnType)
			defaultVal := i.getDefaultValue(returnType)
			i.Env().Define("Result", defaultVal)
			i.Env().Define(classMethod.Name.Value, &ReferenceValue{Env: i.Env(), VarName: "Result"})
		}

		// Execute method body
		result := i.Eval(classMethod.Body)
		if isError(result) {
			return result
		}

		// Extract return value
		var returnValue Value
		if classMethod.ReturnType != nil {
			resultVal, resultOk := i.Env().Get("Result")
			methodNameVal, methodNameOk := i.Env().Get(classMethod.Name.Value)

			if resultOk && resultVal.Type() != "NIL" {
				returnValue = resultVal
			} else if methodNameOk && methodNameVal.Type() != "NIL" {
				returnValue = methodNameVal
			} else if resultOk {
				returnValue = resultVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		return returnValue
	}

	// Handle ClassValue (metaclass) constructor calls
	// Pattern: classVar.Create(args) where classVar is a "class of TBase" metaclass variable
	if classVal, ok := internalObj.(*ClassValue); ok {
		runtimeClass := classVal.ClassInfo
		if runtimeClass == nil {
			return newError("invalid class reference")
		}

		// Look up constructor in the runtime class
		constructorOverloads := i.getMethodOverloadsInHierarchy(runtimeClass, methodName, false)
		if len(constructorOverloads) == 0 {
			return newError("constructor '%s' not found in class '%s'", methodName, runtimeClass.Name)
		}

		// Simple overload resolution by argument count
		var constructor *ast.FunctionDecl
		if len(constructorOverloads) == 1 {
			constructor = constructorOverloads[0]
		} else {
			for _, c := range constructorOverloads {
				if len(c.Parameters) == len(internalArgs) {
					constructor = c
					break
				}
			}
			if constructor == nil {
				return newError("no matching overload for constructor '%s' in class '%s' with %d arguments",
					methodName, runtimeClass.Name, len(internalArgs))
			}
		}

		// Check argument count
		if len(internalArgs) != len(constructor.Parameters) {
			return newError("wrong number of arguments for constructor '%s': expected %d, got %d",
				methodName, len(constructor.Parameters), len(internalArgs))
		}

		// Create new instance of the runtime class
		newInstance := NewObjectInstance(runtimeClass)

		// Initialize all fields with default values
		for fieldName, fieldType := range runtimeClass.Fields {
			var defaultValue Value
			if fieldDecl, hasDecl := runtimeClass.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
				// Use field initializer
				// Phase 3.1.4: unified scope management
				func() {
					defer i.PushScope()()
					for constName, constValue := range runtimeClass.ConstantValues {
						i.Env().Define(constName, constValue)
					}
					defaultValue = i.Eval(fieldDecl.InitValue)
				}()
				if isError(defaultValue) {
					return defaultValue
				}
			} else {
				defaultValue = getZeroValueForType(fieldType, nil)
			}
			newInstance.SetField(fieldName, defaultValue)
		}

		// Execute constructor
		// Phase 3.1.4: unified scope management
		defer i.PushScope()()

		// Check recursion depth
		if i.ctx.GetCallStack().WillOverflow() {
			return i.raiseMaxRecursionExceeded()
		}

		// Push to call stack
		fullMethodName := runtimeClass.Name + "." + methodName
		i.pushCallStack(fullMethodName)
		defer i.popCallStack()

		// Bind Self to the new instance
		i.Env().Define("Self", newInstance)
		i.Env().Define("__CurrentClass__", &ClassInfoValue{ClassInfo: runtimeClass})

		// Add class constants
		i.bindClassConstantsToEnv(runtimeClass)

		// Bind constructor parameters with implicit conversion
		for idx, param := range constructor.Parameters {
			arg := internalArgs[idx]
			if param.Type != nil {
				paramTypeName := param.Type.String()
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
			i.Env().Define(param.Name.Value, arg)
		}

		// Execute constructor body
		result := i.Eval(constructor.Body)
		if isError(result) {
			return result
		}

		// Constructors return the new instance (Result is Self)
		return newInstance
	}

	// Handle INTERFACE values (interface method calls)
	// Pattern: intf.Method(args) where intf is an interface instance
	if intfInst, ok := internalObj.(*InterfaceInstance); ok {
		// Check for nil interface (wrapped object is nil)
		if intfInst.Object == nil {
			return newError("Interface is nil")
		}

		// Verify method exists in interface contract (case-insensitive)
		if !intfInst.HasInterfaceMethod(methodName) {
			return newError("method '%s' not found in interface '%s'", methodName, intfInst.InterfaceName())
		}

		// Extract underlying object and dispatch to it
		// The interface validates the contract; actual method lives on the object
		objVal := intfInst.Object

		// Get class info
		classInfo := objVal.Class
		if classInfo == nil {
			return newError("object has no class information")
		}

		// Find method (case-insensitive)
		method := classInfo.LookupMethod(methodName)
		if method == nil {
			return newError("method '%s' not found in class '%s'", methodName, classInfo.GetName())
		}

		// Call the method with Self bound to the underlying object (not the interface)
		// Phase 3.1.4: unified scope management
		defer i.PushScope()()
		i.Env().Define("Self", objVal)

		result := i.executeUserFunctionViaEvaluator(method, internalArgs)

		return result
	}

	// Handle RECORD values (record method calls)
	// Pattern: record.Method(args) where record is a record instance
	if recVal, ok := internalObj.(*RecordValue); ok {
		// Method lookup - check instance methods first
		method := GetRecordMethod(recVal, methodName)

		// Check for class/static methods on the record type
		var rtv *RecordTypeValue
		recordTypeKey := "__record_type_" + ident.Normalize(recVal.RecordType.Name)
		if typeVal, found := i.Env().Get(recordTypeKey); found {
			rtv, _ = typeVal.(*RecordTypeValue)
		}

		if method == nil && rtv != nil {
			if classMethod, exists := rtv.ClassMethods[ident.Normalize(methodName)]; exists {
				// Static method - no Self, just constants and class vars
				// Phase 3.1.4: unified scope management
				defer i.PushScope()()

				// Bind __CurrentRecord__ for record context
				i.Env().Define("__CurrentRecord__", rtv)

				// Bind constants and class variables
				for constName, constValue := range rtv.Constants {
					i.Env().Define(constName, constValue)
				}
				for varName, varValue := range rtv.ClassVars {
					i.Env().Define(varName, varValue)
				}

				// Check recursion depth
				if i.ctx.GetCallStack().WillOverflow() {
					return i.raiseMaxRecursionExceeded()
				}

				// Push to call stack
				fullMethodName := recVal.RecordType.Name + "." + methodName
				i.pushCallStack(fullMethodName)
				defer i.popCallStack()

				result := i.executeUserFunctionViaEvaluator(classMethod, internalArgs)
				return result
			}
		}

		if method == nil {
			return newError("method '%s' not found in record type '%s'", methodName, recVal.RecordType.Name)
		}

		// Records have value semantics - copy before method execution
		recordCopy := recVal.Copy()

		// Create method environment
		// Phase 3.1.4: unified scope management
		defer i.PushScope()()

		// Bind Self to the record copy
		i.Env().Define("Self", recordCopy)

		// Bind all record fields to environment for direct access
		for fieldName, fieldValue := range recordCopy.Fields {
			i.Env().Define(fieldName, fieldValue)
		}

		// Bind properties for simple field-backed properties
		if recVal.RecordType.Properties != nil {
			for propName, propInfo := range recVal.RecordType.Properties {
				if propInfo.ReadField != "" {
					if fval, exists := recordCopy.Fields[ident.Normalize(propInfo.ReadField)]; exists {
						i.Env().Define(propName, fval)
					}
				}
			}
		}

		// Bind constants and class variables from RecordTypeValue
		if rtv != nil {
			for constName, constValue := range rtv.Constants {
				i.Env().Define(constName, constValue)
			}
			for varName, varValue := range rtv.ClassVars {
				i.Env().Define(varName, varValue)
			}
		}

		// Check recursion depth
		if i.ctx.GetCallStack().WillOverflow() {
			return i.raiseMaxRecursionExceeded()
		}

		// Push to call stack
		fullMethodName := recVal.RecordType.Name + "." + methodName
		i.pushCallStack(fullMethodName)
		defer i.popCallStack()

		// Call the method
		result := i.executeUserFunctionViaEvaluator(method, internalArgs)

		return result
	}

	// Original OBJECT handling
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		panic(fmt.Sprintf("not an object: %s", internalObj.Type()))
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		panic("object has no class information")
	}

	// Find method (case-insensitive) using the existing helper
	method := classInfo.LookupMethod(methodName)
	if method == nil {
		// Neither class method found - panic
		panic(fmt.Sprintf("method '%s' not found in class '%s'", methodName, classInfo.GetName()))
	}

	// Call the method using existing infrastructure
	// Phase 3.1.4: unified scope management
	defer i.PushScope()()
	i.Env().Define("Self", objVal)

	result := i.executeUserFunctionViaEvaluator(method, internalArgs)

	return result
}

// CallInheritedMethod executes an inherited (parent) method with the given arguments.
func (i *Interpreter) CallInheritedMethod(obj evaluator.Value, methodName string, args []evaluator.Value) evaluator.Value {
	// Convert to internal types
	internalObj := obj.(Value)
	internalArgs := convertEvaluatorArgs(args)

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return newError("inherited requires Self to be an object instance, got %s", internalObj.Type())
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		return newError("object has no class information")
	}

	// Check parent class
	if classInfo.GetParent() == nil {
		return newError("class '%s' has no parent class", classInfo.GetName())
	}

	parentInfo := classInfo.GetParent()

	// Find method in parent (case-insensitive)
	methodNameLower := ident.Normalize(methodName)
	methods := parentInfo.GetMethodsMap()
	method, exists := methods[methodNameLower]
	if !exists {
		return newError("method, property, or field '%s' not found in parent class '%s'", methodName, parentInfo.GetName())
	}

	// Call the method using existing infrastructure
	// Phase 3.1.4: unified scope management
	defer i.PushScope()()
	i.Env().Define("Self", objVal)

	result := i.executeUserFunctionViaEvaluator(method, internalArgs)

	return result
}

// ExecuteMethodWithSelf executes a method with Self bound to the given object.
func (i *Interpreter) ExecuteMethodWithSelf(self evaluator.Value, methodDecl any, args []evaluator.Value) evaluator.Value {
	// Type-assert method declaration
	method, ok := methodDecl.(*ast.FunctionDecl)
	if !ok {
		return newError("invalid method declaration type")
	}

	// Convert to internal types
	internalSelf := self.(Value)
	internalArgs := convertEvaluatorArgs(args)

	// Call the method using existing infrastructure
	// Phase 3.1.4: unified scope management
	defer i.PushScope()()
	i.Env().Define("Self", internalSelf)

	result := i.executeUserFunctionViaEvaluator(method, internalArgs)

	return result
}
