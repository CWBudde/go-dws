package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

func (i *Interpreter) evalMethodCall(mc *ast.MethodCallExpression) Value {
	methodNameLower := pkgident.Normalize(mc.Method.Value)

	// Check if the left side is an identifier (unit, class, or instance variable)
	if ident, ok := mc.Object.(*ast.Identifier); ok {
		// Check for unit-qualified function call
		if i.unitRegistry != nil {
			if _, exists := i.unitRegistry.GetUnit(ident.Value); exists {
				fn, err := i.ResolveQualifiedFunction(ident.Value, mc.Method.Value)
				if err == nil {
					args := make([]Value, len(mc.Arguments))
					for idx, arg := range mc.Arguments {
						val := i.Eval(arg)
						if isError(val) {
							return val
						}
						args[idx] = val
					}
					return i.executeUserFunctionViaEvaluator(fn, args)
				}
				return i.newErrorWithLocation(mc, "function '%s' not found in unit '%s'", mc.Method.Value, ident.Value)
			}
		}

		// Check for class name
		classInfo := i.resolveClassInfoByName(ident.Value)
		if classInfo != nil {
			classMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, mc.Method.Value, true)
			instanceMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, mc.Method.Value, false)

			// Allow implicit parameterless constructor
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
				obj := NewObjectInstance(classInfo)

				savedEnv := i.env
				tempEnv := i.PushEnvironment(i.env)
				for constName, constValue := range classInfo.ConstantValues {
					tempEnv.Define(constName, constValue)
				}

				for fieldName, fieldType := range classInfo.Fields {
					var fieldValue Value
					if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
						fieldValue = i.Eval(fieldDecl.InitValue)
						if isError(fieldValue) {
							i.RestoreEnvironment(savedEnv)
							return fieldValue
						}
					} else {
						fieldValue = getZeroValueForType(fieldType, nil)
					}
					obj.SetField(fieldName, fieldValue)
				}

				i.RestoreEnvironment(savedEnv)
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
				// Instance method called on class name (e.g., TClass.Create)
				obj := NewObjectInstance(classInfo)

				fieldInitEnv := i.env
				fieldTempEnv := i.PushEnvironment(i.env)
				for constName, constValue := range classInfo.ConstantValues {
					fieldTempEnv.Define(constName, constValue)
				}

				for fieldName, fieldType := range classInfo.Fields {
					var fieldValue Value
					// Check if field has an initializer expression
					if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
						fieldValue = i.Eval(fieldDecl.InitValue)
						if isError(fieldValue) {
							i.RestoreEnvironment(fieldInitEnv)
							return fieldValue
						}
					} else {
						fieldValue = getZeroValueForType(fieldType, nil)
					}
					obj.SetField(fieldName, fieldValue)
				}

				// Restore environment
				i.RestoreEnvironment(fieldInitEnv)

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
				savedEnv := i.env
				methodEnv := i.PushEnvironment(i.env)

				// Bind Self to the object
				methodEnv.Define("Self", obj)

				// Add class constants to method scope so they can be accessed directly
				i.bindClassConstantsToEnv(classInfo)
				for idx, param := range instanceMethod.Parameters {
					arg := args[idx]

					// Apply implicit conversion if parameter has a type and types don't match
					if param.Type != nil {
						paramTypeName := param.Type.String()
						if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
							arg = converted
						}
					}

					methodEnv.Define(param.Name.Value, arg)
				}

				if instanceMethod.ReturnType != nil || instanceMethod.IsConstructor {
					var defaultVal Value
					if instanceMethod.IsConstructor {
						defaultVal = &NilValue{}
					} else {
						returnType := i.resolveTypeFromAnnotation(instanceMethod.ReturnType)
						defaultVal = i.getDefaultValue(returnType)
					}
					methodEnv.Define("Result", defaultVal)
					// In DWScript, the method name can be used as an alias for Result
					methodEnv.Define(instanceMethod.Name.Value, &ReferenceValue{Env: methodEnv, VarName: "Result"})
				}

				// Execute method body
				result := i.Eval(instanceMethod.Body)
				if isError(result) {
					i.RestoreEnvironment(savedEnv)
					return result
				}

				var returnValue Value
				if instanceMethod.ReturnType != nil || instanceMethod.IsConstructor {
					if instanceMethod.IsConstructor && instanceMethod.ReturnType == nil {
						returnValue = obj
					} else if instanceMethod.IsConstructor && instanceMethod.ReturnType != nil {
						resultVal, resultOk := i.env.Get("Result")
						if resultOk && resultVal.Type() != "NIL" {
							returnValue = resultVal
						} else {
							returnValue = obj
						}
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
						expectedReturnType := instanceMethod.ReturnType.String()
						if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
							returnValue = converted
						}
					}
				} else {
					returnValue = &NilValue{}
				}

				i.RestoreEnvironment(savedEnv)

				return returnValue
			} else {
				return i.newErrorWithLocation(mc, "method '%s' not found in class '%s'", mc.Method.Value, classInfo.Name)
			}
		}

		// Check for record type
		recordTypeKey := "__record_type_" + pkgident.Normalize(ident.Value)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				methodNameLower := pkgident.Normalize(mc.Method.Value)
				classMethodOverloads, hasOverloads := rtv.ClassMethodOverloads[methodNameLower]

				if !hasOverloads || len(classMethodOverloads) == 0 {
					return i.newErrorWithLocation(mc, "static method '%s' not found in record type '%s'", mc.Method.Value, ident.Value)
				}

				// Resolve overload if multiple methods exist
				var staticMethod *ast.FunctionDecl
				var err error

				if len(classMethodOverloads) > 1 {
					staticMethod, err = i.resolveMethodOverload(rtv.RecordType.Name, mc.Method.Value, classMethodOverloads, mc.Arguments)
					if err != nil {
						return i.newErrorWithLocation(mc, "%s", err.Error())
					}
				} else {
					staticMethod = classMethodOverloads[0]
				}

				return i.callRecordStaticMethod(rtv, staticMethod, mc.Arguments, mc)
			}
		}
	}

	// Evaluate object expression for instance method call
	objVal := i.Eval(mc.Object)
	if isError(objVal) {
		return objVal
	}

	// Unwrap type cast wrappers so method dispatch uses the underlying value.
	if castVal, ok := objVal.(*TypeCastValue); ok {
		objVal = castVal.Object
	}

	// Check if it's a ClassInfoValue (Self in a class method)
	if classInfoVal, ok := objVal.(*ClassInfoValue); ok {
		classInfo := classInfoVal.ClassInfo
		classMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, mc.Method.Value, true)

		if len(classMethodOverloads) == 0 {
			return i.newErrorWithLocation(mc, "class method '%s' not found in class '%s'", mc.Method.Value, classInfo.Name)
		}

		classMethod, err := i.resolveMethodOverload(classInfo.Name, mc.Method.Value, classMethodOverloads, mc.Arguments)
		if err != nil {
			return i.newErrorWithLocation(mc, "%s", err.Error())
		}

		return i.executeClassMethod(classInfo, classMethod, mc)
	}

	// Check for metaclass (ClassValue) calling a constructor
	if classVal, ok := objVal.(*ClassValue); ok {
		methodName := mc.Method.Value
		runtimeClass := classVal.ClassInfo
		if runtimeClass == nil {
			return i.newErrorWithLocation(mc, "invalid class reference")
		}

		constructorOverloads := i.getMethodOverloadsInHierarchy(runtimeClass, methodName, false)

		if len(constructorOverloads) == 0 {
			return i.newErrorWithLocation(mc, "constructor '%s' not found in class '%s'", methodName, runtimeClass.Name)
		}

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
		savedEnv := i.env
		methodEnv := i.PushEnvironment(i.env)
		methodEnv.Define("Self", newInstance)

		// Add class constants to method scope so they can be accessed directly
		i.bindClassConstantsToEnv(runtimeClass)

		// Bind constructor parameters to arguments
		for idx, param := range constructor.Parameters {
			arg := args[idx]
			if param.Type != nil {
				paramTypeName := param.Type.String()
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
			methodEnv.Define(param.Name.Value, arg)
		}

		methodEnv.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: runtimeClass})

		// Execute constructor body
		result := i.Eval(constructor.Body)
		if isError(result) {
			i.RestoreEnvironment(savedEnv)
			return result
		}

		i.RestoreEnvironment(savedEnv)
		return newInstance
	}

	// Check for set value with built-in methods
	if setVal, ok := objVal.(*SetValue); ok {
		methodName := pkgident.Normalize(mc.Method.Value)

		// Evaluate method arguments
		args := make([]Value, len(mc.Arguments))
		for idx, arg := range mc.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Dispatch to appropriate set method (case-insensitive)
		switch methodName {
		case "include":
			if len(args) != 1 {
				return i.newErrorWithLocation(mc, "Include expects 1 argument, got %d", len(args))
			}
			return i.evalSetInclude(setVal, args[0])

		case "exclude":
			if len(args) != 1 {
				return i.newErrorWithLocation(mc, "Exclude expects 1 argument, got %d", len(args))
			}
			return i.evalSetExclude(setVal, args[0])

		default:
			return i.newErrorWithLocation(mc, "method '%s' not found for set type", methodName)
		}
	}

	// Check for record value with methods
	if recVal, ok := objVal.(*RecordValue); ok {
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

	// Check for interface instance - delegate to underlying object
	if intfInst, ok := objVal.(*InterfaceInstance); ok {
		if intfInst.Object == nil {
			return i.newErrorWithLocation(mc, "Interface is nil")
		}

		if !intfInst.Interface.HasMethod(mc.Method.Value) {
			return i.newErrorWithLocation(mc, "method '%s' not found in interface '%s'",
				mc.Method.Value, intfInst.Interface.GetName())
		}

		objVal = intfInst.Object
	}

	// Initialize typed nil values when possible (e.g., dynamic arrays with default nil).
	if objVal != nil && objVal.Type() == "NIL" && i.semanticInfo != nil {
		if objType := i.semanticInfo.GetType(mc.Object); objType != nil {
			typeName := objType.String()
			if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
				if arrayType := i.parseInlineArrayType(typeName); arrayType != nil {
					objVal = NewArrayValue(arrayType)
				}
			}
		}
	}

	// Check if object is nil (TObject.Free is nil-safe)
	if objVal == nil || objVal.Type() == "NIL" {
		if strings.EqualFold(strings.TrimSpace(mc.Method.Value), "Free") {
			return &NilValue{}
		}
		message := fmt.Sprintf("Object not instantiated [line: %d, column: %d]", mc.Token.Pos.Line, mc.Token.Pos.Column+1)
		i.raiseException("Exception", message, &mc.Token.Pos)
		return &NilValue{}
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
		// Special handling for enum type methods: Low(), High(), and ByName()
		if tmv, isTypeMeta := objVal.(*TypeMetaValue); isTypeMeta {
			if enumType, isEnum := tmv.TypeInfo.(*types.EnumType); isEnum {
				methodName := pkgident.Normalize(mc.Method.Value)
				switch methodName {
				case "low":
					return &IntegerValue{Value: int64(enumType.Low())}
				case "high":
					return &IntegerValue{Value: int64(enumType.High())}
				case "byname":
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

					searchName := nameStr.Value
					if searchName == "" {
						return &IntegerValue{Value: 0}
					}

					parts := strings.Split(searchName, ".")
					if len(parts) == 2 {
						searchName = parts[1]
					}

					for valueName, ordinalValue := range enumType.Values {
						if pkgident.Equal(valueName, searchName) {
							return &IntegerValue{Value: int64(ordinalValue)}
						}
					}

					return &IntegerValue{Value: 0}
				}
			}
		}

		// Check if helpers provide this method
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

	// Prevent method calls on destroyed instances
	if obj.Destroyed {
		message := fmt.Sprintf("Object already destroyed [line: %d, column: %d]", mc.Token.Pos.Line, mc.Token.Pos.Column)
		i.raiseException("Exception", message, &mc.Token.Pos)
		return &NilValue{}
	}

	// TObject.Free delegates to Destroy and is available on all classes
	if methodNameLower == "free" {
		if len(mc.Arguments) != 0 {
			return i.newErrorWithLocation(mc, "wrong number of arguments for method '%s': expected %d, got %d",
				"Free", 0, len(mc.Arguments))
		}
		return i.runDestructor(obj, obj.Class.LookupMethod("Destroy"), mc)
	}

	// Handle built-in methods that are available on all objects (inherited from TObject)
	if methodNameLower == "classname" {
		return &StringValue{Value: obj.Class.GetName()}
	}
	concreteClass, ok := obj.Class.(*ClassInfo)
	if !ok {
		return i.newErrorWithLocation(mc, "object has invalid class type")
	}

	methodOverloads := i.getMethodOverloadsInHierarchy(concreteClass, mc.Method.Value, false)
	classMethodOverloads := i.getMethodOverloadsInHierarchy(concreteClass, mc.Method.Value, true)

	var method *ast.FunctionDecl
	var err error
	var isClassMethod bool

	// Try instance methods first
	if len(methodOverloads) > 0 {
		method, err = i.resolveMethodOverload(obj.Class.GetName(), mc.Method.Value, methodOverloads, mc.Arguments)
		if err != nil {
			return i.newErrorWithLocation(mc, "%s", err.Error())
		}

		// Use virtual method table for virtual/override methods
		if method != nil && (method.IsVirtual || method.IsOverride) && concreteClass.VirtualMethodTable != nil {
			sig := methodSignature(method)
			if entry, exists := concreteClass.VirtualMethodTable[sig]; exists && entry != nil && entry.IsVirtual {
				if entry.Implementation != nil {
					method = entry.Implementation
				}
			}
		}
	}

	// If no instance method found, try class methods
	if method == nil && len(classMethodOverloads) > 0 {
		method, err = i.resolveMethodOverload(obj.Class.GetName(), mc.Method.Value, classMethodOverloads, mc.Arguments)
		if err != nil {
			return i.newErrorWithLocation(mc, "%s", err.Error())
		}
		isClassMethod = true

		// For non-virtual class methods, use static binding
		if method != nil && !method.IsVirtual && !method.IsOverride {
			topMostMethod := method
			for currentClass := concreteClass.Parent; currentClass != nil; currentClass = currentClass.Parent {
				parentClassMethodOverloads := make([]*ast.FunctionDecl, 0)
				for name, methods := range currentClass.ClassMethodOverloads {
					if pkgident.Equal(name, mc.Method.Value) {
						parentClassMethodOverloads = methods
						break
					}
				}

				if len(parentClassMethodOverloads) > 0 {
					parentMethod, parentErr := i.resolveMethodOverload(currentClass.Name, mc.Method.Value, parentClassMethodOverloads, mc.Arguments)
					if parentErr == nil && parentMethod != nil {
						topMostMethod = parentMethod
					}
				}
			}
			method = topMostMethod
		}

		// Use virtual method table for virtual class methods
		if method != nil && (method.IsVirtual || method.IsOverride) && concreteClass.VirtualMethodTable != nil {
			sig := methodSignature(method)
			if entry, exists := concreteClass.VirtualMethodTable[sig]; exists && entry != nil && entry.IsVirtual {
				method = entry.Implementation
			}
		}
	}

	if method == nil {
		helper, helperMethod, builtinSpec := i.findHelperMethod(obj, mc.Method.Value)
		if helperMethod == nil && builtinSpec == "" {
			return i.newErrorWithLocation(mc, "method '%s' not found in class '%s'", mc.Method.Value, obj.Class.GetName())
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

	if method.IsDestructor {
		return i.runDestructor(obj, method, mc)
	}

	// Virtual constructor dispatch: create new instance of runtime class
	if method.IsConstructor {
		actualConstructor := method
		constructorName := mc.Method.Value
		found := false
		for class := concreteClass; class != nil; class = class.Parent {
			if ctor, exists := class.Constructors[constructorName]; exists {
				actualConstructor = ctor
				found = true
				break
			}
			for name, ctor := range class.Constructors {
				if pkgident.Equal(name, constructorName) {
					actualConstructor = ctor
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		newObj := NewObjectInstance(obj.Class)
		savedEnv := i.env
		methodEnv := i.PushEnvironment(i.env)
		methodEnv.Define("Self", newObj)
		i.bindClassConstantsToEnv(concreteClass)

		// Bind method parameters to arguments
		for idx, param := range actualConstructor.Parameters {
			arg := args[idx]
			if param.Type != nil {
				paramTypeName := param.Type.String()
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
			methodEnv.Define(param.Name.Value, arg)
		}

		// Execute constructor body
		result := i.Eval(actualConstructor.Body)
		if isError(result) {
			i.RestoreEnvironment(savedEnv)
			return result
		}

		// Restore environment and return the NEW object
		i.RestoreEnvironment(savedEnv)
		return newObj
	}

	savedEnv := i.env
	methodEnv := i.PushEnvironment(i.env)

	if i.ctx.GetCallStack().WillOverflow() {
		i.RestoreEnvironment(savedEnv)
		return i.raiseMaxRecursionExceeded()
	}

	fullMethodName := obj.Class.GetName() + "." + mc.Method.Value
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	if isClassMethod {
		methodEnv.Define("Self", &ClassInfoValue{ClassInfo: concreteClass})
	} else {
		methodEnv.Define("Self", obj)
	}

	// Determine the declaring class for inherited resolution
	methodOwner := i.findMethodOwner(concreteClass, method, isClassMethod)
	methodEnv.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: methodOwner})
	i.bindClassConstantsToEnv(concreteClass)

	// Bind method parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.String()
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		methodEnv.Define(param.Name.Value, arg)
	}

	if method.ReturnType != nil {
		returnType := i.resolveTypeFromAnnotation(method.ReturnType)
		defaultVal := i.getDefaultValue(returnType)
		methodEnv.Define("Result", defaultVal)
		methodEnv.Define(method.Name.Value, &ReferenceValue{Env: methodEnv, VarName: "Result"})
	}

	// Execute method body
	result := i.Eval(method.Body)
	if isError(result) {
		i.RestoreEnvironment(savedEnv)
		return result
	}

	// Extract return value (same logic as regular functions)
	var returnValue Value
	if method.ReturnType != nil {
		resultVal, resultOk := i.env.Get("Result")
		methodNameVal, methodNameOk := i.env.Get(method.Name.Value)
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

		switch {
		case resultOk && resultVal.Type() != "NIL":
			returnValue = resultVal
		case methodNameOk && methodNameVal.Type() != "NIL":
			returnValue = methodNameVal
		case resultOk:
			returnValue = resultVal
		case methodNameOk:
			returnValue = methodNameVal
		default:
			returnValue = &NilValue{}
		}

		if returnValue.Type() != "NIL" {
			expectedReturnType := method.ReturnType.String()
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		returnValue = &NilValue{}
	}

	// Restore environment
	i.RestoreEnvironment(savedEnv)

	return returnValue
}

// executeClassMethod executes a class method with Self bound to ClassInfo.
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
	savedEnv := i.env
	methodEnv := i.PushEnvironment(i.env)

	// Check recursion depth before pushing to call stack
	if i.ctx.GetCallStack().WillOverflow() {
		i.RestoreEnvironment(savedEnv)
		return i.raiseMaxRecursionExceeded()
	}

	// Push method name onto call stack for stack traces
	fullMethodName := classInfo.Name + "." + mc.Method.Value
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	// Bind Self to ClassInfo for class methods
	methodEnv.Define("Self", &ClassInfoValue{ClassInfo: classInfo})

	// Bind __CurrentClass__
	methodEnv.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})

	// Add class constants to method scope
	i.bindClassConstantsToEnv(classInfo)

	// Bind method parameters with implicit conversion
	for idx, param := range classMethod.Parameters {
		arg := args[idx]
		if param.Type != nil {
			paramTypeName := param.Type.String()
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}
		methodEnv.Define(param.Name.Value, arg)
	}

	// Initialize Result for functions
	if classMethod.ReturnType != nil {
		returnType := i.resolveTypeFromAnnotation(classMethod.ReturnType)
		defaultVal := i.getDefaultValue(returnType)
		methodEnv.Define("Result", defaultVal)
		methodEnv.Define(classMethod.Name.Value, &ReferenceValue{Env: methodEnv, VarName: "Result"})
	}

	// Execute method body
	result := i.Eval(classMethod.Body)
	if isError(result) {
		i.RestoreEnvironment(savedEnv)
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
			expectedReturnType := classMethod.ReturnType.String()
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		returnValue = &NilValue{}
	}

	// Restore environment
	i.RestoreEnvironment(savedEnv)
	return returnValue
}

// resolveMethodOverload resolves method overload based on argument types.
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
func (i *Interpreter) getMethodOverloadsInHierarchy(classInfo *ClassInfo, methodName string, isClassMethod bool) []*ast.FunctionDecl {
	var result []*ast.FunctionDecl

	// Check for constructors (only when isClassMethod = false)
	if !isClassMethod {
		for ctorName, constructorOverloads := range classInfo.ConstructorOverloads {
			if pkgident.Equal(ctorName, methodName) && len(constructorOverloads) > 0 {
				result = append(result, constructorOverloads...)
				return result
			}
		}
	}

	// Walk up the class hierarchy for regular methods
	for classInfo != nil {
		var overloads []*ast.FunctionDecl
		if isClassMethod {
			for name, methods := range classInfo.ClassMethodOverloads {
				if pkgident.Equal(name, methodName) {
					overloads = methods
					break
				}
			}
		} else {
			for name, methods := range classInfo.MethodOverloads {
				if pkgident.Equal(name, methodName) {
					overloads = methods
					break
				}
			}
		}

		// Add overloads from this class level
		for _, candidate := range overloads {
			hidden := false
			candidateType := i.extractFunctionType(candidate)
			if candidateType != nil {
				for _, existing := range result {
					existingType := i.extractFunctionType(existing)
					if existingType != nil && semantic.SignaturesEqual(candidateType, existingType) {
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

// findMethodOwner returns the class in the hierarchy that declares the given method.
// Falls back to the runtime class if not found.
func (i *Interpreter) findMethodOwner(classInfo *ClassInfo, method *ast.FunctionDecl, isClassMethod bool) *ClassInfo {
	if classInfo == nil || method == nil {
		return classInfo
	}

	for c := classInfo; c != nil; c = c.Parent {
		var overloads map[string][]*ast.FunctionDecl
		if isClassMethod {
			overloads = c.ClassMethodOverloads
		} else {
			overloads = c.MethodOverloads
		}

		for _, methods := range overloads {
			for _, m := range methods {
				if m == method {
					return c
				}
			}
		}
	}

	return classInfo
}
