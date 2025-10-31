package interp

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// evalNewExpression evaluates object instantiation (TClassName.Create(...)).
// It looks up the class, creates an object instance, initializes fields, and calls the constructor.
func (i *Interpreter) evalNewExpression(ne *ast.NewExpression) Value {
	// Look up class in registry
	className := ne.ClassName.Value
	classInfo, exists := i.classes[className]
	if !exists {
		return i.newErrorWithLocation(ne, "class '%s' not found", className)
	}

	// Check if trying to instantiate an abstract class
	if classInfo.IsAbstract {
		return i.newErrorWithLocation(ne, "cannot instantiate abstract class '%s'", className)
	}

	// Check if trying to instantiate an external class
	// External classes are implemented outside DWScript and cannot be instantiated directly
	// Future: Provide hooks for Go FFI implementation
	if classInfo.IsExternal {
		return i.newErrorWithLocation(ne, "cannot instantiate external class '%s' - external classes are not supported", className)
	}

	// Create new object instance
	obj := NewObjectInstance(classInfo)

	// Initialize all fields with default values based on their types
	for fieldName, fieldType := range classInfo.Fields {
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
		obj.SetField(fieldName, defaultValue)
	}

	// Special handling for Exception.Create
	// Exception constructors are built-in and take predefined arguments.
	// NewExpression always implies Create constructor in DWScript.
	if i.isExceptionClass(classInfo) {
		// EHost.Create(cls, msg) - first argument is exception class name, second is message.
		if classInfo.InheritsFrom("EHost") {
			if len(ne.Arguments) != 2 {
				return i.newErrorWithLocation(ne, "EHost.Create requires class name and message arguments")
			}

			classVal := i.Eval(ne.Arguments[0])
			if isError(classVal) {
				return classVal
			}
			messageVal := i.Eval(ne.Arguments[1])
			if isError(messageVal) {
				return messageVal
			}

			exceptionClass := classVal.String()
			if strVal, ok := classVal.(*StringValue); ok {
				exceptionClass = strVal.Value
			}

			message := messageVal.String()
			if strVal, ok := messageVal.(*StringValue); ok {
				message = strVal.Value
			}

			obj.SetField("ExceptionClass", &StringValue{Value: exceptionClass})
			obj.SetField("Message", &StringValue{Value: message})
			return obj
		}

		// Other exception classes accept a single message argument.
		if len(ne.Arguments) == 1 {
			msgVal := i.Eval(ne.Arguments[0])
			if isError(msgVal) {
				return msgVal
			}
			if strVal, ok := msgVal.(*StringValue); ok {
				obj.SetField("Message", &StringValue{Value: strVal.Value})
			} else {
				obj.SetField("Message", &StringValue{Value: msgVal.String()})
			}
			return obj
		}
	}

	// Call constructor if present
	if classInfo.Constructor != nil {
		// Evaluate constructor arguments
		args := make([]Value, len(ne.Arguments))
		for idx, arg := range ne.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Create method environment with Self bound to object
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind constructor parameters to arguments
		for idx, param := range classInfo.Constructor.Parameters {
			if idx < len(args) {
				i.env.Define(param.Name.Value, args[idx])
			}
		}

		// For constructors with return types, initialize the Result variable
		// This allows constructors to use "Result := Self" to return the object
		if classInfo.Constructor.ReturnType != nil {
			i.env.Define("Result", obj)
			i.env.Define(classInfo.Constructor.Name.Value, obj)
		}

		// Execute constructor body
		result := i.Eval(classInfo.Constructor.Body)
		if isError(result) {
			i.env = savedEnv
			return result
		}

		// Restore environment
		i.env = savedEnv
	}

	return obj
}

// evalMemberAccess evaluates field access (obj.field) or class variable access (TClass.Variable).
// It evaluates the object expression and retrieves the field value.
// For class variable access, it checks if the left side is a class name.
func (i *Interpreter) evalMemberAccess(ma *ast.MemberAccessExpression) Value {
	// Check if the left side is a class identifier (for static access: TClass.Variable)
	if ident, ok := ma.Object.(*ast.Identifier); ok {
		// First, check if this identifier refers to a unit (for qualified access: UnitName.Symbol)
		// Task 9.134: Support unit-qualified access
		if i.unitRegistry != nil {
			if _, exists := i.unitRegistry.GetUnit(ident.Value); exists {
				// This is unit-qualified access: UnitName.Symbol
				// Try to resolve as a variable/constant first
				if val, err := i.ResolveQualifiedVariable(ident.Value, ma.Member.Value); err == nil {
					return val
				}
				// If not a variable, it might be a function being passed as a reference
				// For now, we'll return an error since function references aren't fully supported yet
				// The actual function call will be handled in evalCallExpression
				return i.newErrorWithLocation(ma, "qualified name '%s.%s' cannot be used as a value (functions must be called)", ident.Value, ma.Member.Value)
			}
		}

		// Check if this identifier refers to a class
		if classInfo, exists := i.classes[ident.Value]; exists {
			// This is static access: TClass.Variable
			// Look up the class variable
			if classVarValue, exists := classInfo.ClassVars[ma.Member.Value]; exists {
				return classVarValue
			}
			// Not a class variable - this is an error
			return i.newErrorWithLocation(ma, "class variable '%s' not found in class '%s'", ma.Member.Value, classInfo.Name)
		}

		// Check if this identifier refers to an enum type (for scoped access: TColor.Red)
		// Look for enum type metadata stored in environment
		enumTypeKey := "__enum_type_" + ident.Value
		if enumTypeVal, ok := i.env.Get(enumTypeKey); ok {
			if _, isEnumType := enumTypeVal.(*EnumTypeValue); isEnumType {
				// This is scoped enum access: TColor.Red
				// Look up the enum value
				valueName := ma.Member.Value
				if val, exists := i.env.Get(valueName); exists {
					if enumVal, isEnum := val.(*EnumValue); isEnum {
						// Verify the value belongs to this enum type
						if enumVal.TypeName == ident.Value {
							return enumVal
						}
					}
				}
				// Enum value not found
				return i.newErrorWithLocation(ma, "enum value '%s' not found in enum '%s'", ma.Member.Value, ident.Value)
			}
		}
	}

	// Not static access - evaluate the object expression for instance access
	objVal := i.Eval(ma.Object)
	if isError(objVal) {
		return objVal
	}

	// Task 8.75: Check if it's a record value
	if recordVal, ok := objVal.(*RecordValue); ok {
		// Access record field
		fieldValue, exists := recordVal.Fields[ma.Member.Value]
		if !exists {
			// Task 9.86: Check if helpers provide this member
			helper, helperProp := i.findHelperProperty(recordVal, ma.Member.Value)
			if helperProp != nil {
				return i.evalHelperPropertyRead(helper, helperProp, recordVal, ma)
			}
			return i.newErrorWithLocation(ma, "field '%s' not found in record '%s'", ma.Member.Value, recordVal.RecordType.Name)
		}
		return fieldValue
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
		// Task 9.86: Not an object - check if helpers provide this member
		helper, helperProp := i.findHelperProperty(objVal, ma.Member.Value)
		if helperProp != nil {
			return i.evalHelperPropertyRead(helper, helperProp, objVal, ma)
		}
		return i.newErrorWithLocation(ma, "cannot access member '%s' of type '%s' (no helper found)",
			ma.Member.Value, objVal.Type())
	}

	memberName := ma.Member.Value

	// Handle built-in properties/methods available on all objects (inherited from TObject)
	if memberName == "ClassName" {
		// ClassName returns the runtime type name of the object
		return &StringValue{Value: obj.Class.Name}
	}

	// Task 8.53: Check if this is a property access (properties take precedence over fields)
	if propInfo := obj.Class.lookupProperty(memberName); propInfo != nil {
		return i.evalPropertyRead(obj, propInfo, ma)
	}

	// Not a property - try direct field access
	fieldValue := obj.GetField(memberName)
	if fieldValue == nil {
		// Task 9.173: Check if it's a method
		if method, exists := obj.Class.Methods[memberName]; exists {
			// Task 9.173: If the method has no parameters, auto-invoke it
			// This allows DWScript syntax: obj.Method instead of obj.Method()
			if len(method.Parameters) == 0 {
				// Create a synthetic method call expression to use existing infrastructure
				methodCall := &ast.MethodCallExpression{
					Token:     ma.Token,
					Object:    ma.Object,
					Method:    ma.Member,
					Arguments: []ast.Expression{},
				}
				return i.evalMethodCall(methodCall)
			}

			// Method has parameters - return as method pointer for passing as callback
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

		// Task 9.86: Check if helpers provide this member
		helper, helperProp := i.findHelperProperty(obj, memberName)
		if helperProp != nil {
			return i.evalHelperPropertyRead(helper, helperProp, obj, ma)
		}
		return i.newErrorWithLocation(ma, "field '%s' not found in class '%s'", memberName, obj.Class.Name)
	}

	return fieldValue
}

// evalPropertyRead evaluates a property read access.
// Handles field-backed, method-backed, and expression-backed properties.
func (i *Interpreter) evalPropertyRead(obj *ObjectInstance, propInfo *types.PropertyInfo, node ast.Node) Value {
	switch propInfo.ReadKind {
	case types.PropAccessField:
		// Task 8.53a: Field or method access - check at runtime which it is
		// First try as a field
		if _, exists := obj.Class.Fields[propInfo.ReadSpec]; exists {
			fieldValue := obj.GetField(propInfo.ReadSpec)
			if fieldValue == nil {
				return i.newErrorWithLocation(node, "property '%s' read field '%s' not found", propInfo.Name, propInfo.ReadSpec)
			}
			return fieldValue
		}

		// Not a field - try as a method (getter)
		method := obj.Class.lookupMethod(propInfo.ReadSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' read specifier '%s' not found as field or method", propInfo.Name, propInfo.ReadSpec)
		}

		// Call the getter method
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// For functions, initialize the Result variable
		if method.ReturnType != nil {
			i.env.Define("Result", &NilValue{})
			i.env.Define(method.Name.Value, &NilValue{})
		}

		// Execute method body
		i.Eval(method.Body)

		// Get return value
		var returnValue Value
		if method.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				returnValue = resultVal
			} else if methodNameVal, ok := i.env.Get(method.Name.Value); ok {
				returnValue = methodNameVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue

	case types.PropAccessMethod:
		// Task 8.53b: Method access - call getter method
		// Check if method exists
		method := obj.Class.lookupMethod(propInfo.ReadSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' getter method '%s' not found", propInfo.Name, propInfo.ReadSpec)
		}

		// Call the getter method with no arguments (indexed properties handled separately)
		// Create method environment with Self bound to object
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// For functions, initialize the Result variable
		if method.ReturnType != nil {
			i.env.Define("Result", &NilValue{})
			i.env.Define(method.Name.Value, &NilValue{})
		}

		// Execute method body
		i.Eval(method.Body)

		// Get return value
		var returnValue Value
		if method.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				returnValue = resultVal
			} else if methodNameVal, ok := i.env.Get(method.Name.Value); ok {
				returnValue = methodNameVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue

	case types.PropAccessExpression:
		// Task 8.53c / 8.56: Expression access - evaluate expression in context of object
		// For now, return an error as expression evaluation is complex
		return i.newErrorWithLocation(node, "expression-based property getters not yet supported")

	default:
		return i.newErrorWithLocation(node, "property '%s' has no read access", propInfo.Name)
	}
}

// evalPropertyWrite evaluates a property write access.
// Handles field-backed and method-backed property setters.
func (i *Interpreter) evalPropertyWrite(obj *ObjectInstance, propInfo *types.PropertyInfo, value Value, node ast.Node) Value {
	switch propInfo.WriteKind {
	case types.PropAccessField:
		// Task 8.54a: Field or method access - check at runtime which it is
		// First try as a field
		if _, exists := obj.Class.Fields[propInfo.WriteSpec]; exists {
			obj.SetField(propInfo.WriteSpec, value)
			return value
		}

		// Not a field - try as a method (setter)
		method := obj.Class.lookupMethod(propInfo.WriteSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' write specifier '%s' not found as field or method", propInfo.Name, propInfo.WriteSpec)
		}

		// Call the setter method with the value as argument
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind the value parameter (setter should have exactly one parameter)
		if len(method.Parameters) >= 1 {
			paramName := method.Parameters[0].Name.Value
			i.env.Define(paramName, value)
		}

		// Execute method body
		i.Eval(method.Body)

		// Restore environment
		i.env = savedEnv

		return value

	case types.PropAccessMethod:
		// Task 8.54b: Method access - call setter method with value
		// Check if method exists
		method := obj.Class.lookupMethod(propInfo.WriteSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' setter method '%s' not found", propInfo.Name, propInfo.WriteSpec)
		}

		// Call the setter method with the value as argument
		// Create method environment with Self bound to object
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind the value parameter (setter should have exactly one parameter)
		if len(method.Parameters) >= 1 {
			i.env.Define(method.Parameters[0].Name.Value, value)
		}

		// Execute method body
		i.Eval(method.Body)

		// Restore environment
		i.env = savedEnv

		return value

	case types.PropAccessNone:
		// Read-only property
		return i.newErrorWithLocation(node, "property '%s' is read-only", propInfo.Name)

	default:
		return i.newErrorWithLocation(node, "property '%s' has no write access", propInfo.Name)
	}
}

// evalMethodCall evaluates a method call (obj.Method(...)) or class method call (TClass.Method(...)).
// It looks up the method in the object's class hierarchy and executes it with Self bound to the object.
// For class methods, Self is not bound as they are static methods.
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

		// Check if this identifier refers to a class
		if classInfo, exists := i.classes[ident.Value]; exists {
			// Check if this is a class method (static method) or instance method called as constructor
			classMethod, isClassMethod := classInfo.ClassMethods[mc.Method.Value]
			instanceMethod, isInstanceMethod := classInfo.Methods[mc.Method.Value]

			if isClassMethod {
				// This is a class method - execute it without Self binding
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
				if len(args) != len(classMethod.Parameters) {
					return i.newErrorWithLocation(mc, "wrong number of arguments for class method '%s': expected %d, got %d",
						mc.Method.Value, len(classMethod.Parameters), len(args))
				}

				// Create method environment (NO Self binding for class methods)
				methodEnv := NewEnclosedEnvironment(i.env)
				savedEnv := i.env
				i.env = methodEnv

				// Bind __CurrentClass__ so class variables can be accessed
				i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})

				// Bind method parameters to arguments with implicit conversion
				for idx, param := range classMethod.Parameters {
					arg := args[idx]

					// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
					if param.Type != nil {
						paramTypeName := param.Type.Name
						if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
							arg = converted
						}
					}

					i.env.Define(param.Name.Value, arg)
				}

				// For functions (not procedures), initialize the Result variable
				if classMethod.ReturnType != nil {
					i.env.Define("Result", &NilValue{})
					// Also define the method name as an alias for Result (DWScript style)
					i.env.Define(classMethod.Name.Value, &NilValue{})
				}

				// Execute method body
				result := i.Eval(classMethod.Body)
				if isError(result) {
					i.env = savedEnv
					return result
				}

				// Extract return value (same logic as regular functions)
				var returnValue Value
				if classMethod.ReturnType != nil {
					// Check both Result and method name variable
					resultVal, resultOk := i.env.Get("Result")
					methodNameVal, methodNameOk := i.env.Get(classMethod.Name.Value)

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

					// Task 8.19c: Apply implicit conversion if return type doesn't match
					if returnValue.Type() != "NIL" {
						expectedReturnType := classMethod.ReturnType.Name
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
			} else if isInstanceMethod {
				// This is an instance method being called from the class (e.g., TClass.Create())
				// Create a new instance and call the method on it
				obj := NewObjectInstance(classInfo)

				// Initialize all fields with default values
				for fieldName, fieldType := range classInfo.Fields {
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
					obj.SetField(fieldName, defaultValue)
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

				// Bind method parameters to arguments with implicit conversion
				for idx, param := range instanceMethod.Parameters {
					arg := args[idx]

					// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
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
				if instanceMethod.ReturnType != nil || instanceMethod.IsConstructor {
					i.env.Define("Result", &NilValue{})
					// Also define the method name as an alias for Result (DWScript style)
					i.env.Define(instanceMethod.Name.Value, &NilValue{})
				}

				// Execute method body
				result := i.Eval(instanceMethod.Body)
				if isError(result) {
					i.env = savedEnv
					return result
				}

				// Extract return value (same logic as regular functions)
				var returnValue Value
				if instanceMethod.ReturnType != nil || instanceMethod.IsConstructor {
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
					} else if instanceMethod.IsConstructor {
						// Constructors return the object instance by default
						returnValue = obj
					} else {
						returnValue = &NilValue{}
					}

					// Task 8.19c: Apply implicit conversion if return type doesn't match (but not for constructors)
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
	}

	// Not static method call - evaluate the object expression for instance method call
	objVal := i.Eval(mc.Object)
	if isError(objVal) {
		return objVal
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
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

	// Look up method in object's class
	method := obj.GetMethod(mc.Method.Value)
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

	// Create method environment with Self bound to object
	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	// Bind Self to the object
	i.env.Define("Self", obj)

	// Bind method parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		i.env.Define(param.Name.Value, arg)
	}

	// For functions (not procedures), initialize the Result variable
	if method.ReturnType != nil {
		i.env.Define("Result", &NilValue{})
		// Also define the method name as an alias for Result (DWScript style)
		i.env.Define(method.Name.Value, &NilValue{})
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

		// Task 8.19c: Apply implicit conversion if return type doesn't match
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
