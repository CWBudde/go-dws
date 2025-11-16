package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
)

func (i *Interpreter) evalMethodCall(mc *ast.MethodCallExpression) Value {
	// Check if the left side is an identifier (could be unit, class, or instance variable)
	if ident, ok := mc.Object.(*ast.Identifier); ok {
		// First, check if this identifier refers to a unit
		if i.unitRegistry != nil {
			if _, exists := i.unitRegistry.GetUnit(ident.Value); exists {
				// This is a unit-qualified function call: UnitName.FunctionName()
				fn, err := i.ResolveQualifiedFunction(ident.Value, mc.Method.Value)
				if err == nil {
					// Evaluate arguments
					args := make([]Value, len(mc.Arguments))
					for idx, arg := range mc.Arguments {
						val := i.Eval(arg)
						if isError(val) {
							return val
						}
						args[idx] = val
					}
					return i.callUserFunction(fn, args)
				}
				// Function not found in unit
				return i.newErrorWithLocation(mc, "function '%s' not found in unit '%s'", mc.Method.Value, ident.Value)
			}
		}

		// Check if this identifier refers to a class (case-insensitive)
		// Task 9.82: Case-insensitive class lookup (DWScript is case-insensitive)
		var classInfo *ClassInfo
		for className, class := range i.classes {
			if strings.EqualFold(className, ident.Value) {
				classInfo = class
				break
			}
		}
		if classInfo != nil {
			// Task 9.67: Check for method overloads and resolve based on argument types
			// Task 9.68: getMethodOverloadsInHierarchy now handles constructors automatically when isClassMethod = false
			classMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, mc.Method.Value, true)
			instanceMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, mc.Method.Value, false) // includes constructors

			// Task 9.68: Special handling for constructor calls with 0 arguments
			// If no parameterless constructor exists but we're calling with 0 args,
			// allow it as an implicit parameterless constructor (just initialize fields)
			hasConstructor := false
			hasParameterlessConstructor := false
			for _, method := range instanceMethodOverloads {
				if method.IsConstructor {
					hasConstructor = true
					if len(method.Parameters) == 0 {
						hasParameterlessConstructor = true
						break
					}
				}
			}
			if len(mc.Arguments) == 0 && hasConstructor && !hasParameterlessConstructor {
				// Create object with default field values (no constructor body execution)
				obj := NewObjectInstance(classInfo)

				// Task 9.6: Initialize all fields with field initializers or default values
				savedEnv := i.env
				tempEnv := NewEnclosedEnvironment(i.env)
				for constName, constValue := range classInfo.ConstantValues {
					tempEnv.Define(constName, constValue)
				}
				i.env = tempEnv

				for fieldName, fieldType := range classInfo.Fields {
					var fieldValue Value
					if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
						fieldValue = i.Eval(fieldDecl.InitValue)
						if isError(fieldValue) {
							i.env = savedEnv
							return fieldValue
						}
					} else {
						fieldValue = getZeroValueForType(fieldType, nil)
					}
					obj.SetField(fieldName, fieldValue)
				}

				i.env = savedEnv
				return obj
			}

			var classMethod *ast.FunctionDecl
			var instanceMethod *ast.FunctionDecl
			var err error

			// Resolve class method overload
			if len(classMethodOverloads) > 0 {
				classMethod, err = i.resolveMethodOverload(classInfo.Name, mc.Method.Value, classMethodOverloads, mc.Arguments)
				if err != nil && len(instanceMethodOverloads) == 0 {
					return i.newErrorWithLocation(mc, "%s", err.Error())
				}
			}

			// Resolve instance method overload (including constructors)
			if len(instanceMethodOverloads) > 0 {
				instanceMethod, err = i.resolveMethodOverload(classInfo.Name, mc.Method.Value, instanceMethodOverloads, mc.Arguments)
				if err != nil && classMethod == nil {
					return i.newErrorWithLocation(mc, "%s", err.Error())
				}
			}

			if classMethod != nil {
				// This is a class method - execute it with Self bound to the class
				return i.executeClassMethod(classInfo, classMethod, mc)
			} else if instanceMethod != nil {
				// This is an instance method being called from the class (e.g., TClass.Create())
				// Create a new instance and call the method on it
				obj := NewObjectInstance(classInfo)

				// Task 9.6: Initialize all fields with field initializers or default values
				// Create temporary environment with class constants for field initializer evaluation
				fieldInitEnv := i.env
				fieldTempEnv := NewEnclosedEnvironment(i.env)
				for constName, constValue := range classInfo.ConstantValues {
					fieldTempEnv.Define(constName, constValue)
				}
				i.env = fieldTempEnv

				for fieldName, fieldType := range classInfo.Fields {
					var fieldValue Value
					// Check if field has an initializer expression
					if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
						// Evaluate the field initializer
						fieldValue = i.Eval(fieldDecl.InitValue)
						if isError(fieldValue) {
							i.env = fieldInitEnv
							return fieldValue
						}
					} else {
						// Use getZeroValueForType to get appropriate default value
						fieldValue = getZeroValueForType(fieldType, nil)
					}
					obj.SetField(fieldName, fieldValue)
				}

				// Restore environment
				i.env = fieldInitEnv

				// Evaluate method arguments
				args := make([]Value, len(mc.Arguments))
				for idx, arg := range mc.Arguments {
					val := i.Eval(arg)
					if isError(val) {
						return val
					}
					args[idx] = val
				}

				// Check argument count matches parameter count
				if len(args) != len(instanceMethod.Parameters) {
					return i.newErrorWithLocation(mc, "wrong number of arguments for method '%s': expected %d, got %d",
						mc.Method.Value, len(instanceMethod.Parameters), len(args))
				}

				// Create method environment with Self bound to new object
				methodEnv := NewEnclosedEnvironment(i.env)
				savedEnv := i.env
				i.env = methodEnv

				// Bind Self to the object
				i.env.Define("Self", obj)

				// Add class constants to method scope so they can be accessed directly
				i.bindClassConstantsToEnv(classInfo)

				// NOTE: We do NOT add fields to the environment here.
				// The evalSimpleAssignment and Eval(Identifier) functions already handle
				// field access by checking if Self is bound and looking up fields on the object.
				// Adding fields to the environment would break this mechanism.

				// Bind method parameters to arguments with implicit conversion
				for idx, param := range instanceMethod.Parameters {
					arg := args[idx]

					// Apply implicit conversion if parameter has a type and types don't match
					if param.Type != nil {
						paramTypeName := param.Type.Name
						if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
							arg = converted
						}
					}

					i.env.Define(param.Name.Value, arg)
				}

				// For functions (not procedures), initialize the Result variable
				// For constructors, always initialize Result even if no explicit return type
				// Task 9.221: Use appropriate default value based on return type
				if instanceMethod.ReturnType != nil || instanceMethod.IsConstructor {
					var defaultVal Value
					if instanceMethod.IsConstructor {
						// Constructors default to NIL (or will be set to Self)
						defaultVal = &NilValue{}
					} else {
						returnType := i.resolveTypeFromAnnotation(instanceMethod.ReturnType)
						defaultVal = i.getDefaultValue(returnType)
					}
					i.env.Define("Result", defaultVal)
					// In DWScript, the method name can be used as an alias for Result
					i.env.Define(instanceMethod.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
				}

				// Execute method body
				result := i.Eval(instanceMethod.Body)
				if isError(result) {
					i.env = savedEnv
					return result
				}

				// NOTE: We do NOT need to copy field values back from the environment.
				// Field assignments go directly to the object via evalSimpleAssignment.

				// Extract return value (same logic as regular functions)
				var returnValue Value
				if instanceMethod.ReturnType != nil || instanceMethod.IsConstructor {
					// Task 9.32: For constructors, always return the object instance
					// (constructors don't use Result variables)
					if instanceMethod.IsConstructor {
						returnValue = obj
					} else {
						// Check both Result and method name variable
						resultVal, resultOk := i.env.Get("Result")
						methodNameVal, methodNameOk := i.env.Get(instanceMethod.Name.Value)

						// Use whichever variable is not nil, preferring Result if both are set
						if resultOk && resultVal.Type() != "NIL" {
							returnValue = resultVal
						} else if methodNameOk && methodNameVal.Type() != "NIL" {
							returnValue = methodNameVal
						} else if resultOk {
							returnValue = resultVal
						} else if methodNameOk {
							returnValue = methodNameVal
						} else {
							returnValue = &NilValue{}
						}
					}

					// Apply implicit conversion if return type doesn't match (but not for constructors)
					if instanceMethod.ReturnType != nil && returnValue.Type() != "NIL" {
						expectedReturnType := instanceMethod.ReturnType.Name
						if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
							returnValue = converted
						}
					}
				} else {
					// Procedure - no return value
					returnValue = &NilValue{}
				}

				// Restore environment
				i.env = savedEnv

				return returnValue
			} else {
				// Neither class method nor instance method found
				return i.newErrorWithLocation(mc, "method '%s' not found in class '%s'", mc.Method.Value, classInfo.Name)
			}
		}

		// Task 9.7f: Check if this identifier refers to a record type
		// Task 9.225: Normalize to lowercase for case-insensitive lookups
		recordTypeKey := "__record_type_" + strings.ToLower(ident.Value)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// This is TRecord.Method() - check for static method with overload support
				methodNameLower := strings.ToLower(mc.Method.Value)
				classMethodOverloads, hasOverloads := rtv.ClassMethodOverloads[methodNameLower]

				if !hasOverloads || len(classMethodOverloads) == 0 {
					// Static method not found
					return i.newErrorWithLocation(mc, "static method '%s' not found in record type '%s'", mc.Method.Value, ident.Value)
				}

				// Resolve overload if multiple methods exist
				var staticMethod *ast.FunctionDecl
				var err error

				if len(classMethodOverloads) > 1 {
					// Multiple overloads - need to resolve based on arguments
					staticMethod, err = i.resolveMethodOverload(rtv.RecordType.Name, mc.Method.Value, classMethodOverloads, mc.Arguments)
					if err != nil {
						return i.newErrorWithLocation(mc, "%s", err.Error())
					}
				} else {
					// Single method - use it directly
					staticMethod = classMethodOverloads[0]
				}

				// Execute static method WITHOUT Self binding
				return i.callRecordStaticMethod(rtv, staticMethod, mc.Arguments, mc)
			}
		}
	}

	// Not static method call - evaluate the object expression for instance method call
	objVal := i.Eval(mc.Object)
	if isError(objVal) {
		return objVal
	}

	// Task 9.7: Check if it's a ClassInfoValue (Self in a class method)
	// This handles: Self.Method() where Self is bound to ClassInfoValue in a class method
	if classInfoVal, ok := objVal.(*ClassInfoValue); ok {
		classInfo := classInfoVal.ClassInfo
		// Look for class methods only (class methods on ClassInfoValue)
		classMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, mc.Method.Value, true)

		if len(classMethodOverloads) == 0 {
			return i.newErrorWithLocation(mc, "class method '%s' not found in class '%s'", mc.Method.Value, classInfo.Name)
		}

		// Resolve overload
		classMethod, err := i.resolveMethodOverload(classInfo.Name, mc.Method.Value, classMethodOverloads, mc.Arguments)
		if err != nil {
			return i.newErrorWithLocation(mc, "%s", err.Error())
		}

		// Execute class method using helper
		return i.executeClassMethod(classInfo, classMethod, mc)
	}

	// Task 9.4.4: Check if it's a metaclass (ClassValue) calling a constructor
	// This handles: var cls: class of TParent; cls := TChild; obj := cls.Create;
	if classVal, ok := objVal.(*ClassValue); ok {
		// Only constructors can be called on metaclass values
		methodName := mc.Method.Value

		// Look up constructor in the runtime class (virtual dispatch)
		runtimeClass := classVal.ClassInfo
		if runtimeClass == nil {
			return i.newErrorWithLocation(mc, "invalid class reference")
		}

		// Get all constructor overloads with this name from class hierarchy
		// Use getMethodOverloadsInHierarchy with isClassMethod=false to get constructors
		constructorOverloads := i.getMethodOverloadsInHierarchy(runtimeClass, methodName, false)

		if len(constructorOverloads) == 0 {
			return i.newErrorWithLocation(mc, "constructor '%s' not found in class '%s'", methodName, runtimeClass.Name)
		}

		// Task 9.67: Resolve constructor overload based on argument types
		constructor, err := i.resolveMethodOverload(runtimeClass.Name, methodName, constructorOverloads, mc.Arguments)
		if err != nil {
			return i.newErrorWithLocation(mc, "%s", err.Error())
		}

		// Evaluate constructor arguments
		args := make([]Value, len(mc.Arguments))
		for idx, arg := range mc.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Check argument count
		if len(args) != len(constructor.Parameters) {
			return i.newErrorWithLocation(mc, "wrong number of arguments for constructor '%s': expected %d, got %d",
				methodName, len(constructor.Parameters), len(args))
		}

		// Create new instance of the runtime class (the class stored in ClassValue)
		newInstance := NewObjectInstance(runtimeClass)

		// Initialize all fields with default values
		for fieldName, fieldType := range runtimeClass.Fields {
			var defaultValue Value
			switch fieldType {
			case types.INTEGER:
				defaultValue = &IntegerValue{Value: 0}
			case types.FLOAT:
				defaultValue = &FloatValue{Value: 0.0}
			case types.STRING:
				defaultValue = &StringValue{Value: ""}
			case types.BOOLEAN:
				defaultValue = &BooleanValue{Value: false}
			default:
				defaultValue = &NilValue{}
			}
			newInstance.SetField(fieldName, defaultValue)
		}

		// Create method environment with Self bound to new instance
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv
		i.env.Define("Self", newInstance)

		// Add class constants to method scope so they can be accessed directly
		i.bindClassConstantsToEnv(runtimeClass)

		// Bind constructor parameters to arguments
		for idx, param := range constructor.Parameters {
			arg := args[idx]
			if param.Type != nil {
				paramTypeName := param.Type.Name
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
			i.env.Define(param.Name.Value, arg)
		}

		// Task 9.73: Bind __CurrentClass__ so ClassName can be accessed in constructor
		i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: runtimeClass})

		// Execute constructor body
		result := i.Eval(constructor.Body)
		if isError(result) {
			i.env = savedEnv
			return result
		}

		// Restore environment
		i.env = savedEnv

		// Return the newly created instance
		return newInstance
	}

	// Check if it's a set value with built-in methods (Include, Exclude)
	if setVal, ok := objVal.(*SetValue); ok {
		methodName := mc.Method.Value

		// Evaluate method arguments
		args := make([]Value, len(mc.Arguments))
		for idx, arg := range mc.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Dispatch to appropriate set method
		switch methodName {
		case "Include":
			if len(args) != 1 {
				return i.newErrorWithLocation(mc, "Include expects 1 argument, got %d", len(args))
			}
			return i.evalSetInclude(setVal, args[0])

		case "Exclude":
			if len(args) != 1 {
				return i.newErrorWithLocation(mc, "Exclude expects 1 argument, got %d", len(args))
			}
			return i.evalSetExclude(setVal, args[0])

		default:
			return i.newErrorWithLocation(mc, "method '%s' not found for set type", methodName)
		}
	}

	// Task 9.7: Check if it's a record value with methods
	if recVal, ok := objVal.(*RecordValue); ok {
		// Convert MethodCallExpression to member access for record method calls
		memberAccess := &ast.MemberAccessExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: mc.Token,
				},
			},
			Object: mc.Object,
			Member: mc.Method,
		}
		return i.evalRecordMethodCall(recVal, memberAccess, mc.Arguments, mc.Object)
	}

	// Task 9.1.1: Check if it's an interface instance
	// If so, extract the underlying object and delegate method call to it
	if intfInst, ok := objVal.(*InterfaceInstance); ok {
		if intfInst.Object == nil {
			return i.newErrorWithLocation(mc, "Interface is nil")
		}

		// Verify the method exists in the interface definition
		// This ensures we only call methods that are part of the interface contract
		if !intfInst.Interface.HasMethod(mc.Method.Value) {
			return i.newErrorWithLocation(mc, "method '%s' not found in interface '%s'",
				mc.Method.Value, intfInst.Interface.Name)
		}

		// Delegate to the underlying object for actual method dispatch
		objVal = intfInst.Object
	}

	// Task 9.7: Check if object is nil
	// When calling o.Method() where o is nil, always raise "Object not instantiated"
	// Note: Class methods can only be called without error when called directly on
	// the class name (TClass.Method), not via a nil instance variable (o.Method)
	if _, isNil := objVal.(*NilValue); isNil {
		return i.newErrorWithLocation(mc, "Object not instantiated")
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
		// Special handling for enum type methods: Low(), High(), and ByName()
		if tmv, isTypeMeta := objVal.(*TypeMetaValue); isTypeMeta {
			if enumType, isEnum := tmv.TypeInfo.(*types.EnumType); isEnum {
				methodName := strings.ToLower(mc.Method.Value)
				if methodName == "low" {
					return &IntegerValue{Value: int64(enumType.Low())}
				} else if methodName == "high" {
					return &IntegerValue{Value: int64(enumType.High())}
				} else if methodName == "byname" {
					// ByName(name: string): Integer
					if len(mc.Arguments) != 1 {
						return i.newErrorWithLocation(mc, "ByName expects 1 argument, got %d", len(mc.Arguments))
					}
					nameArg := i.Eval(mc.Arguments[0])
					if isError(nameArg) {
						return nameArg
					}
					nameStr, ok := nameArg.(*StringValue)
					if !ok {
						return i.newErrorWithLocation(mc, "ByName expects string argument, got %s", nameArg.Type())
					}

					// Look up the enum value by name (case-insensitive)
					// Support both simple names ('a') and qualified names ('MyEnum.a')
					searchName := nameStr.Value
					if searchName == "" {
						// Empty string returns 0 (DWScript behavior - returns first enum ordinal value)
						return &IntegerValue{Value: 0}
					}

					// Check for qualified name (TypeName.ValueName)
					parts := strings.Split(searchName, ".")
					if len(parts) == 2 {
						// Use the value name part
						searchName = parts[1]
					}

					// Look up the value (case-insensitive)
					for valueName, ordinalValue := range enumType.Values {
						if strings.EqualFold(valueName, searchName) {
							return &IntegerValue{Value: int64(ordinalValue)}
						}
					}

					// Value not found, return 0 (DWScript behavior - returns 0 instead of raising error)
					return &IntegerValue{Value: 0}
				}
			}
		}

		// Task 9.86: Not an object - check if helpers provide this method
		helper, helperMethod, builtinSpec := i.findHelperMethod(objVal, mc.Method.Value)
		if helperMethod == nil && builtinSpec == "" {
			return i.newErrorWithLocation(mc, "cannot call method '%s' on type '%s' (no helper found)",
				mc.Method.Value, objVal.Type())
		}

		// Evaluate method arguments
		args := make([]Value, len(mc.Arguments))
		for idx, arg := range mc.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Call the helper method
		return i.callHelperMethod(helper, helperMethod, builtinSpec, objVal, args, mc)
	}

	// Handle built-in methods that are available on all objects (inherited from TObject)
	if mc.Method.Value == "ClassName" {
		// ClassName returns the runtime type name of the object
		return &StringValue{Value: obj.Class.Name}
	}

	// Task 9.67: Resolve method overload based on argument types
	methodOverloads := i.getMethodOverloadsInHierarchy(obj.Class, mc.Method.Value, false)

	// Task 9.7: Also check class methods (which can be called on instances)
	classMethodOverloads := i.getMethodOverloadsInHierarchy(obj.Class, mc.Method.Value, true)

	var method *ast.FunctionDecl
	var err error
	var isClassMethod bool

	// Try instance methods first
	if len(methodOverloads) > 0 {
		method, err = i.resolveMethodOverload(obj.Class.Name, mc.Method.Value, methodOverloads, mc.Arguments)
		if err != nil {
			return i.newErrorWithLocation(mc, "%s", err.Error())
		}
	}

	// If no instance method found, try class methods
	if method == nil && len(classMethodOverloads) > 0 {
		method, err = i.resolveMethodOverload(obj.Class.Name, mc.Method.Value, classMethodOverloads, mc.Arguments)
		if err != nil {
			return i.newErrorWithLocation(mc, "%s", err.Error())
		}
		isClassMethod = true
	}

	if method == nil {
		// Task 9.86: Check if helpers provide this method
		helper, helperMethod, builtinSpec := i.findHelperMethod(obj, mc.Method.Value)
		if helperMethod == nil && builtinSpec == "" {
			return i.newErrorWithLocation(mc, "method '%s' not found in class '%s'", mc.Method.Value, obj.Class.Name)
		}

		// Evaluate method arguments
		args := make([]Value, len(mc.Arguments))
		for idx, arg := range mc.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Call the helper method
		return i.callHelperMethod(helper, helperMethod, builtinSpec, obj, args, mc)
	}

	// Evaluate method arguments
	args := make([]Value, len(mc.Arguments))
	for idx, arg := range mc.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Check argument count matches parameter count
	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(mc, "wrong number of arguments for method '%s': expected %d, got %d",
			mc.Method.Value, len(method.Parameters), len(args))
	}

	// Task 9.4.3 & 9.73.3: Special handling for virtual constructors called on instances
	// When calling a constructor on an instance (o.Create), create a NEW instance
	// of the object's runtime type using virtual dispatch
	if method.IsConstructor {
		// For virtual constructor dispatch, find the constructor in the object's runtime class
		// Start from the runtime class and work up the hierarchy to find the most derived constructor
		actualConstructor := method // fallback to the already-resolved method

		// Try to find a constructor with the same name in the runtime class hierarchy
		// This implements virtual dispatch - we start from the most derived class
		constructorName := mc.Method.Value
		found := false
		for class := obj.Class; class != nil; class = class.Parent {
			// Check if this class has the constructor (case-sensitive match first)
			if ctor, exists := class.Constructors[constructorName]; exists {
				actualConstructor = ctor
				found = true
				break
			}
			// Case-insensitive fallback
			for name, ctor := range class.Constructors {
				if strings.EqualFold(name, constructorName) {
					actualConstructor = ctor
					found = true
					break
				}
			}
			if found {
				break // Found a constructor
			}
		}

		// Create a NEW instance of the runtime class (not the existing object)
		// Always use obj.Class (the runtime type), not actualClass (where constructor was found)
		newObj := NewObjectInstance(obj.Class)

		// Initialize all fields with default values
		for fieldName, fieldType := range obj.Class.Fields {
			var defaultValue Value
			switch fieldType {
			case types.INTEGER:
				defaultValue = &IntegerValue{Value: 0}
			case types.FLOAT:
				defaultValue = &FloatValue{Value: 0.0}
			case types.STRING:
				defaultValue = &StringValue{Value: ""}
			case types.BOOLEAN:
				defaultValue = &BooleanValue{Value: false}
			default:
				defaultValue = &NilValue{}
			}
			newObj.SetField(fieldName, defaultValue)
		}

		// Create method environment with Self bound to NEW object
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv
		i.env.Define("Self", newObj)

		// Add class constants to method scope so they can be accessed directly
		i.bindClassConstantsToEnv(obj.Class)

		// Bind method parameters to arguments
		for idx, param := range actualConstructor.Parameters {
			arg := args[idx]
			if param.Type != nil {
				paramTypeName := param.Type.Name
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
			i.env.Define(param.Name.Value, arg)
		}

		// Execute constructor body
		result := i.Eval(actualConstructor.Body)
		if isError(result) {
			i.env = savedEnv
			return result
		}

		// Restore environment and return the NEW object
		i.env = savedEnv
		return newObj
	}

	// Normal method call (not a constructor)
	// Create method environment with Self bound to object
	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	// Task 9.x: Check recursion depth before pushing to call stack
	if len(i.callStack) >= i.maxRecursionDepth {
		i.env = savedEnv // Restore environment before raising exception
		return i.raiseMaxRecursionExceeded()
	}

	// Task 9.108: Push method name onto call stack for stack traces
	fullMethodName := obj.Class.Name + "." + mc.Method.Value
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	// Task 9.7: Bind Self appropriately based on method type
	if isClassMethod {
		// For class methods, bind Self to the class (not the instance)
		i.env.Define("Self", &ClassInfoValue{ClassInfo: obj.Class})
	} else {
		// For instance methods, bind Self to the object
		i.env.Define("Self", obj)
	}

	// Add class constants to method scope so they can be accessed directly
	i.bindClassConstantsToEnv(obj.Class)

	// Bind method parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		i.env.Define(param.Name.Value, arg)
	}

	// For functions (not procedures), initialize the Result variable
	// Task 9.221: Use appropriate default value based on return type
	if method.ReturnType != nil {
		returnType := i.resolveTypeFromAnnotation(method.ReturnType)
		defaultVal := i.getDefaultValue(returnType)
		i.env.Define("Result", defaultVal)
		// In DWScript, the method name can be used as an alias for Result
		i.env.Define(method.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
	}

	// Execute method body
	result := i.Eval(method.Body)
	if isError(result) {
		i.env = savedEnv
		return result
	}

	// Extract return value (same logic as regular functions)
	var returnValue Value
	if method.ReturnType != nil {
		// Check both Result and method name variable
		resultVal, resultOk := i.env.Get("Result")
		methodNameVal, methodNameOk := i.env.Get(method.Name.Value)

		// Dereference ReferenceValue if needed
		if resultOk {
			if refVal, isRef := resultVal.(*ReferenceValue); isRef {
				if derefVal, err := refVal.Dereference(); err == nil {
					resultVal = derefVal
				}
			}
		}
		if methodNameOk {
			if refVal, isRef := methodNameVal.(*ReferenceValue); isRef {
				if derefVal, err := refVal.Dereference(); err == nil {
					methodNameVal = derefVal
				}
			}
		}

		// Use whichever variable is not nil, preferring Result if both are set
		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if methodNameOk && methodNameVal.Type() != "NIL" {
			returnValue = methodNameVal
		} else if resultOk {
			returnValue = resultVal
		} else if methodNameOk {
			returnValue = methodNameVal
		} else {
			returnValue = &NilValue{}
		}

		// Apply implicit conversion if return type doesn't match
		if returnValue.Type() != "NIL" {
			expectedReturnType := method.ReturnType.Name
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		// Procedure - no return value
		returnValue = &NilValue{}
	}

	// Restore environment
	i.env = savedEnv

	return returnValue
}

// executeClassMethod executes a class method with Self bound to ClassInfo.
// This is used for both direct class method calls (TClass.Method) and
// class method calls on Self when Self is already a ClassInfoValue.
func (i *Interpreter) executeClassMethod(
	classInfo *ClassInfo,
	classMethod *ast.FunctionDecl,
	mc *ast.MethodCallExpression,
) Value {
	// Evaluate arguments
	args := make([]Value, len(mc.Arguments))
	for idx, arg := range mc.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Check argument count
	if len(args) != len(classMethod.Parameters) {
		return i.newErrorWithLocation(mc, "wrong number of arguments for class method '%s': expected %d, got %d",
			mc.Method.Value, len(classMethod.Parameters), len(args))
	}

	// Create method environment
	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	// Check recursion depth before pushing to call stack
	if len(i.callStack) >= i.maxRecursionDepth {
		i.env = savedEnv
		return i.raiseMaxRecursionExceeded()
	}

	// Push method name onto call stack for stack traces
	fullMethodName := classInfo.Name + "." + mc.Method.Value
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	// Bind Self to ClassInfo for class methods
	i.env.Define("Self", &ClassInfoValue{ClassInfo: classInfo})

	// Bind __CurrentClass__
	i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})

	// Add class constants to method scope
	i.bindClassConstantsToEnv(classInfo)

	// Bind method parameters with implicit conversion
	for idx, param := range classMethod.Parameters {
		arg := args[idx]
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}
		i.env.Define(param.Name.Value, arg)
	}

	// Initialize Result for functions
	if classMethod.ReturnType != nil {
		returnType := i.resolveTypeFromAnnotation(classMethod.ReturnType)
		defaultVal := i.getDefaultValue(returnType)
		i.env.Define("Result", defaultVal)
		i.env.Define(classMethod.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
	}

	// Execute method body
	result := i.Eval(classMethod.Body)
	if isError(result) {
		i.env = savedEnv
		return result
	}

	// Extract return value
	var returnValue Value
	if classMethod.ReturnType != nil {
		resultVal, resultOk := i.env.Get("Result")
		methodNameVal, methodNameOk := i.env.Get(classMethod.Name.Value)

		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if methodNameOk && methodNameVal.Type() != "NIL" {
			returnValue = methodNameVal
		} else if resultOk {
			returnValue = resultVal
		} else if methodNameOk {
			returnValue = methodNameVal
		} else {
			returnValue = &NilValue{}
		}

		if returnValue.Type() != "NIL" {
			expectedReturnType := classMethod.ReturnType.Name
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		returnValue = &NilValue{}
	}

	// Restore environment
	i.env = savedEnv
	return returnValue
}

// resolveMethodOverload resolves method overload based on argument types.
// Task 9.67: Overload resolution for method calls
func (i *Interpreter) resolveMethodOverload(className, methodName string, overloads []*ast.FunctionDecl, argExprs []ast.Expression) (*ast.FunctionDecl, error) {
	// If only one overload, use it (fast path)
	if len(overloads) == 1 {
		return overloads[0], nil
	}

	// Evaluate arguments to get their types
	argTypes := make([]types.Type, len(argExprs))
	for idx, argExpr := range argExprs {
		val := i.Eval(argExpr)
		if isError(val) {
			return nil, fmt.Errorf("error evaluating argument %d: %v", idx+1, val)
		}
		argTypes[idx] = i.getValueType(val)
	}

	// Convert method declarations to semantic symbols for resolution
	candidates := make([]*semantic.Symbol, len(overloads))
	for idx, method := range overloads {
		methodType := i.extractFunctionType(method)
		if methodType == nil {
			return nil, fmt.Errorf("unable to extract method type for overload %d of '%s.%s'", idx+1, className, methodName)
		}

		candidates[idx] = &semantic.Symbol{
			Name:                 method.Name.Value,
			Type:                 methodType,
			HasOverloadDirective: method.IsOverload,
		}
	}

	// Use semantic analyzer's overload resolution
	selected, err := semantic.ResolveOverload(candidates, argTypes)
	if err != nil {
		// Task 9.63: Use DWScript-compatible error message
		return nil, fmt.Errorf("There is no overloaded version of \"%s.%s\" that can be called with these arguments", className, methodName)
	}

	// Find which method declaration corresponds to the selected symbol
	selectedType := selected.Type.(*types.FunctionType)
	for _, method := range overloads {
		methodType := i.extractFunctionType(method)
		if methodType != nil && semantic.SignaturesEqual(methodType, selectedType) &&
			methodType.ReturnType.Equals(selectedType.ReturnType) {
			return method, nil
		}
	}

	return nil, fmt.Errorf("internal error: resolved overload not found in candidate list")
}

// getMethodOverloadsInHierarchy collects all overloads of a method from the class hierarchy.
// Task 9.67: Support inheritance for method overloads
// Task 9.68: Also includes constructor overloads when the method name is a constructor
func (i *Interpreter) getMethodOverloadsInHierarchy(classInfo *ClassInfo, methodName string, isClassMethod bool) []*ast.FunctionDecl {
	var result []*ast.FunctionDecl

	// Task 9.68: Check if this is a constructor call
	// Task 9.82: Case-insensitive lookup (DWScript is case-insensitive)
	// Constructors are stored separately in ConstructorOverloads
	// Task 9.21: Only return constructors when isClassMethod = false (constructors are instance methods)
	// Task 9.4: Constructors are copied from parent to child (unlike regular methods),
	// so we only need to check the current class's ConstructorOverloads
	if !isClassMethod {
		for ctorName, constructorOverloads := range classInfo.ConstructorOverloads {
			if strings.EqualFold(ctorName, methodName) && len(constructorOverloads) > 0 {
				// This is a constructor - return constructor overloads from this class
				// (which includes inherited constructors due to copying in evalClassDeclaration)
				result = append(result, constructorOverloads...)
				return result
			}
		}
	}

	// Walk up the class hierarchy for regular methods
	// Task 9.21.6: Fix overload resolution - child methods hide parent methods with same signature
	// Task 9.82: Case-insensitive method lookup
	for classInfo != nil {
		var overloads []*ast.FunctionDecl
		if isClassMethod {
			// Case-insensitive lookup in ClassMethodOverloads
			for name, methods := range classInfo.ClassMethodOverloads {
				if strings.EqualFold(name, methodName) {
					overloads = methods
					break
				}
			}
		} else {
			// Case-insensitive lookup in MethodOverloads
			for name, methods := range classInfo.MethodOverloads {
				if strings.EqualFold(name, methodName) {
					overloads = methods
					break
				}
			}
		}

		// Add overloads from this class level
		for _, candidate := range overloads {
			// Check if this method signature is already in result (hidden by child class)
			hidden := false
			candidateType := i.extractFunctionType(candidate)
			if candidateType != nil {
				for _, existing := range result {
					existingType := i.extractFunctionType(existing)
					if existingType != nil && semantic.SignaturesEqual(candidateType, existingType) {
						// Same signature already exists from child class - this one is hidden
						hidden = true
						break
					}
				}
			}
			// Only add if not hidden by a child class method
			if !hidden {
				result = append(result, candidate)
			}
		}

		// Move to parent class
		classInfo = classInfo.Parent
	}

	return result
}
