package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalNewExpression evaluates object instantiation (TClassName.Create(...)).
// It looks up the class or record, creates an instance, initializes fields, and calls the constructor.
func (i *Interpreter) evalNewExpression(ne *ast.NewExpression) Value {
	// Look up class in registry (case-insensitive)
	className := ne.ClassName.Value
	classInfo := i.resolveClassInfoByName(className)

	// Check if this is a record type with static methods
	if classInfo == nil {
		// Check if this is a record type
		recordTypeKey := "__record_type_" + ident.Normalize(className)
		if typeVal, ok := i.Env().Get(recordTypeKey); ok {
			if _, ok := typeVal.(*RecordTypeValue); ok {
				// This is a record static method call (TRecord.Create(...))
				// Convert to MethodCallExpression and delegate to evalMethodCall
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: ne.Token,
						},
					},
					Object: ne.ClassName,
					Method: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: ne.Token,
							},
						},
						Value: "Create",
					},
					Arguments: ne.Arguments,
				}
				return i.evalMethodCall(methodCall)
			}
		}

		return i.newErrorWithLocation(ne, "class '%s' not found", className)
	}

	// Check if trying to instantiate an abstract class
	if classInfo.IsAbstractFlag {
		return i.newErrorWithLocation(ne, "Trying to create an instance of an abstract class")
	}

	// Check if trying to instantiate an external class
	// External classes are implemented outside DWScript and cannot be instantiated directly
	// Future: Provide hooks for Go FFI implementation
	if classInfo.IsExternalFlag {
		return i.newErrorWithLocation(ne, "cannot instantiate external class '%s' - external classes are not supported", className)
	}

	// Create new object instance
	obj := NewObjectInstance(classInfo)

	// Initialize all fields with field initializers or default values
	defer i.PushScope()()
	// Add class constants to the temporary environment
	for constName, constValue := range classInfo.ConstantValues {
		i.Env().Define(constName, constValue)
	}

	for fieldName, fieldType := range classInfo.Fields {
		var fieldValue Value

		// Check if field has an initializer expression
		if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			// Evaluate the field initializer
			fieldValue = i.Eval(fieldDecl.InitValue)
			if isError(fieldValue) {
				return fieldValue
			}
		} else {
			// Use getZeroValueForType to get appropriate default value
			// Pass nil for methodsLookup since class fields should not auto-instantiate nested types
			fieldValue = getZeroValueForType(fieldType, nil)
		}
		obj.SetField(fieldName, fieldValue)
	}

	// Special handling for exception classes
	if i.isExceptionClass(classInfo) {
		// EHost.Create(cls, msg) - first argument is exception class name, second is message.
		// Keep this special case for EHost with 2 arguments
		if classInfo.InheritsFrom("EHost") && len(ne.Arguments) == 2 {
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

		// For other exception classes, let normal constructor resolution handle it
		// This allows custom constructors while preserving backward compatibility
	}

	// Resolve constructor overload based on arguments
	var constructor *ast.FunctionDecl
	constructorName := i.getDefaultConstructorName(classInfo)
	constructorOverloads := i.getMethodOverloadsInHierarchy(classInfo, constructorName, false)

	// Allow parameterless instantiation even if no explicit constructor exists
	if len(ne.Arguments) == 0 && len(constructorOverloads) > 0 {
		hasParameterlessConstructor := false
		for _, ctor := range constructorOverloads {
			if len(ctor.Parameters) == 0 {
				hasParameterlessConstructor = true
				break
			}
		}
		if !hasParameterlessConstructor {
			// No constructor body to execute - just return object with default fields
			// (fields already initialized above)
			return obj
		}
	}

	// Resolve overload if multiple constructors exist
	if len(constructorOverloads) > 0 {
		var err error
		constructor, err = i.resolveMethodOverload(className, constructorName, constructorOverloads, ne.Arguments)
		if err != nil {
			return i.newErrorWithLocation(ne, "%s", err.Error())
		}
	} else {
		// No constructor found - use default (empty) constructor
		constructor = nil

		// Fallback: For exception classes with single message argument, use built-in logic
		if i.isExceptionClass(classInfo) && len(ne.Arguments) == 1 {
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
	if constructor != nil {
		// Evaluate constructor arguments
		args := make([]Value, len(ne.Arguments))
		for idx, arg := range ne.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Check argument count matches parameter count
		if len(args) != len(constructor.Parameters) {
			return i.newErrorWithLocation(ne, "wrong number of arguments for constructor '%s': expected %d, got %d",
				constructorName, len(constructor.Parameters), len(args))
		}

		// Create method environment with Self bound to object
		defer i.PushScope()()

		// Bind Self to the object
		i.Env().Define("Self", obj)

		// Bind constructor parameters to arguments
		for idx, param := range constructor.Parameters {
			if idx < len(args) {
				i.Env().Define(param.Name.Value, args[idx])
			}
		}

		// For constructors with return types, initialize the Result variable
		// This allows constructors to use "Result := Self" to return the object
		if constructor.ReturnType != nil {
			i.Env().Define("Result", obj)
			i.Env().Define(constructor.Name.Value, obj)
		}

		// Bind __CurrentClass__ for ClassName access in constructor
		i.Env().Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})

		// Execute constructor body
		result := i.Eval(constructor.Body)
		if isError(result) {
			return result
		}

		// For auto-detected constructors with explicit return types, check if Result was modified
		if constructor.IsConstructor && constructor.ReturnType != nil {
			resultVal, resultOk := i.Env().Get("Result")
			if resultOk && resultVal.Type() != "NIL" {
				// Result was set, use it as the return value (should be the same object)
				if objInstance, ok := resultVal.(*ObjectInstance); ok {
					obj = objInstance
				}
			}
		}
	}

	return obj
}

// getDefaultConstructorName returns the name of the default constructor for a class.
// It checks the class hierarchy for a constructor marked as 'default', falling back to "Create".
func (i *Interpreter) getDefaultConstructorName(class *ClassInfo) string {
	// Check current class and parents for default constructor
	for current := class; current != nil; current = current.Parent {
		if current.DefaultConstructor != "" {
			return current.DefaultConstructor
		}
	}
	// No default constructor found, use "Create" as fallback
	return "Create"
}
