package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalNewExpression evaluates object instantiation (TClassName.Create(...)).
// It looks up the class or record, creates an instance, initializes fields, and calls the constructor.
func (i *Interpreter) evalNewExpression(ne *ast.NewExpression) Value {
	// Look up class in registry (case-insensitive)
	// Task 9.82: Case-insensitive class lookup (DWScript is case-insensitive)
	className := ne.ClassName.Value
	classInfo := i.resolveClassInfoByName(className)

	// Task 9.38: Also check if this is a record type with static methods
	// The parser creates NewExpression for TType.Create(...) syntax
	if classInfo == nil {
		// Check if this is a record type
		recordTypeKey := "__record_type_" + ident.Normalize(className)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
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
	if classInfo.IsAbstract {
		return i.newErrorWithLocation(ne, "Trying to create an instance of an abstract class")
	}

	// Check if trying to instantiate an external class
	// External classes are implemented outside DWScript and cannot be instantiated directly
	// Future: Provide hooks for Go FFI implementation
	if classInfo.IsExternal {
		return i.newErrorWithLocation(ne, "cannot instantiate external class '%s' - external classes are not supported", className)
	}

	// Create new object instance
	obj := NewObjectInstance(classInfo)

	// Task 9.5: Initialize all fields with field initializers or default values
	// Task 9.6: Create temporary environment with class constants for field initializer evaluation
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	// Add class constants to the temporary environment
	for constName, constValue := range classInfo.ConstantValues {
		tempEnv.Define(constName, constValue)
	}
	i.env = tempEnv

	for fieldName, fieldType := range classInfo.Fields {
		var fieldValue Value

		// Check if field has an initializer expression
		if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			// Evaluate the field initializer
			fieldValue = i.Eval(fieldDecl.InitValue)
			if isError(fieldValue) {
				i.env = savedEnv
				return fieldValue
			}
		} else {
			// Use getZeroValueForType to get appropriate default value
			// Pass nil for methodsLookup since class fields should not auto-instantiate nested types
			fieldValue = getZeroValueForType(fieldType, nil)
		}
		obj.SetField(fieldName, fieldValue)
	}

	// Restore environment
	i.env = savedEnv

	// Special handling for Exception.Create
	// Exception constructors are built-in and take predefined arguments.
	// NewExpression always implies Create constructor in DWScript.
	// Task 9.XX: Only apply hardcoded logic if no custom constructor exists
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

	// Task 9.68: Resolve constructor overload based on arguments
	// Check for constructor overloads first (supports both TClass.Create and new TClass)
	var constructor *ast.FunctionDecl
	// Task 9.3: Use default constructor if specified, otherwise fall back to "Create"
	constructorName := i.getDefaultConstructorName(classInfo)

	// Task 9.4: Get all constructor overloads from class hierarchy (case-insensitive lookup)
	// This ensures inherited virtual constructors are properly found
	constructorOverloads := i.getMethodOverloadsInHierarchy(classInfo, constructorName, false)

	// Task 9.68: Special handling for implicit parameterless constructor
	// If calling with 0 arguments and no parameterless constructor exists,
	// allow it (just initialize fields with default values)
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
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind constructor parameters to arguments
		for idx, param := range constructor.Parameters {
			if idx < len(args) {
				i.env.Define(param.Name.Value, args[idx])
			}
		}

		// For constructors with return types, initialize the Result variable
		// This allows constructors to use "Result := Self" to return the object
		if constructor.ReturnType != nil {
			i.env.Define("Result", obj)
			i.env.Define(constructor.Name.Value, obj)
		}

		// Task 9.73: Bind __CurrentClass__ so ClassName can be accessed in constructor
		i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})

		// Execute constructor body
		result := i.Eval(constructor.Body)
		if isError(result) {
			i.env = savedEnv
			return result
		}

		// For auto-detected constructors with explicit return types, check if Result was modified
		if constructor.IsConstructor && constructor.ReturnType != nil {
			resultVal, resultOk := i.env.Get("Result")
			if resultOk && resultVal.Type() != "NIL" {
				// Result was set, use it as the return value (should be the same object)
				if objInstance, ok := resultVal.(*ObjectInstance); ok {
					obj = objInstance
				}
			}
		}

		// Restore environment
		i.env = savedEnv
	}

	return obj
}

// getDefaultConstructorName returns the name of the default constructor for a class.
// It checks the class hierarchy for a constructor marked as 'default'.
// Falls back to "Create" if no default constructor is found.
// Task 9.3: Support for default constructors
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
