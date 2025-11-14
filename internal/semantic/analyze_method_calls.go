package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// analyzeMethodCallExpression analyzes a method call on an object
func (a *Analyzer) analyzeMethodCallExpression(expr *ast.MethodCallExpression) types.Type {
	// Analyze the object expression
	objectType := a.analyzeExpression(expr.Object)
	if objectType == nil {
		// Error already reported
		return nil
	}

	methodName := expr.Method.Value
	// Task 9.16.2.9: Normalize method name to lowercase for case-insensitive lookup
	methodNameLower := strings.ToLower(methodName)

	// Task 9.7: Handle metaclass type (class of T) for constructor calls
	// When we have TExample.CreateWith(...), TExample has type ClassOfType(TExample)
	// We need to unwrap to ClassType(TExample) to look up constructors
	if metaclassType, ok := objectType.(*types.ClassOfType); ok {
		if metaclassType.ClassType != nil {
			objectType = metaclassType.ClassType
		}
	}

	// Check if object is an interface type
	if interfaceType, ok := objectType.(*types.InterfaceType); ok {
		// Look up method in interface (including inherited methods from parent interfaces)
		// Use lowercase for case-insensitive lookup
		methodType, found := interfaceType.GetMethod(methodNameLower)

		// Check parent interfaces
		if !found && interfaceType.Parent != nil {
			allMethods := types.GetAllInterfaceMethods(interfaceType)
			methodType, found = allMethods[methodNameLower]
		}

		if !found {
			a.addError("interface '%s' has no method '%s' at %s",
				interfaceType.Name, methodName, expr.Token.Pos.String())
			return nil
		}

		// Validate arguments
		if len(expr.Arguments) != len(methodType.Parameters) {
			a.addError("method '%s' expects %d arguments, got %d at %s",
				methodName, len(methodType.Parameters), len(expr.Arguments),
				expr.Token.Pos.String())
			return methodType.ReturnType
		}

		// Check argument types
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			expectedType := methodType.Parameters[i]
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("argument %d to method '%s' has type %s, expected %s at %s",
					i+1, methodName, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		}

		return methodType.ReturnType
	}

	// Check if object is a class type
	classType, ok := objectType.(*types.ClassType)
	if !ok {
		// Task 9.7: Check if object is a record type with methods
		if recordType, isRecord := objectType.(*types.RecordType); isRecord {
			// First check for class methods (static methods) with overload support
			classOverloads := recordType.GetClassMethodOverloads(methodNameLower)
			if len(classOverloads) > 0 {
				// This is a static method call - resolve overload based on arguments
				argTypes := make([]types.Type, len(expr.Arguments))
				for i, arg := range expr.Arguments {
					argType := a.analyzeExpression(arg)
					if argType == nil {
						return nil
					}
					argTypes[i] = argType
				}

				// Convert MethodInfo to Symbol for ResolveOverload
				candidates := make([]*Symbol, len(classOverloads))
				for i, overload := range classOverloads {
					candidates[i] = &Symbol{
						Type: overload.Signature,
					}
				}

				// Resolve overload
				selected, err := ResolveOverload(candidates, argTypes)
				if err != nil {
					a.addError("no matching overload for class method '%s.%s' with these arguments at %s",
						recordType.Name, methodName, expr.Token.Pos.String())
					return nil
				}

				methodType := selected.Type.(*types.FunctionType)

				// Validate argument types (for better error messages)
				for i, arg := range expr.Arguments {
					if i >= len(methodType.Parameters) {
						break
					}
					paramType := methodType.Parameters[i]
					argType := a.analyzeExpressionWithExpectedType(arg, paramType)
					if argType != nil && !a.canAssign(argType, paramType) {
						a.addError("argument %d to class method '%s.%s' has type %s, expected %s at %s",
							i+1, recordType.Name, methodName, argType.String(), paramType.String(),
							expr.Token.Pos.String())
					}
				}

				return methodType.ReturnType
			}

			// Check for instance methods
			method := recordType.GetMethod(methodNameLower)
			if method == nil {
				// Method not found in record, check if a helper provides it
				_, helperMethod := a.hasHelperMethod(objectType, methodName)
				if helperMethod == nil {
					a.addError("method '%s' not found in record type '%s' at %s",
						methodName, recordType.Name, expr.Token.Pos.String())
					return nil
				}
				// Use the helper method
				method = helperMethod
			}

			// Validate method arguments
			if len(expr.Arguments) != len(method.Parameters) {
				a.addError("record method '%s' expects %d arguments, got %d at %s",
					methodName, len(method.Parameters), len(expr.Arguments),
					expr.Token.Pos.String())
				return method.ReturnType
			}

			// Check argument types
			for i, arg := range expr.Arguments {
				argType := a.analyzeExpression(arg)
				expectedType := method.Parameters[i]
				if argType != nil && !a.canAssign(argType, expectedType) {
					a.addError("argument %d to record method '%s' has type %s, expected %s at %s",
						i+1, methodName, argType.String(), expectedType.String(),
						expr.Token.Pos.String())
				}
			}

			return method.ReturnType
		}

		// Task 9.83: For non-class, non-record types, check if helpers provide this method
		_, helperMethod := a.hasHelperMethod(objectType, methodName)
		if helperMethod == nil {
			a.addError("method call on type %s requires a helper, got no helper with method '%s' at %s",
				objectType.String(), methodName, expr.Token.Pos.String())
			return nil
		}

		// Validate helper method arguments (support optional parameters)
		// Count required parameters (those without defaults)
		requiredParams := len(helperMethod.Parameters)
		if helperMethod.DefaultValues != nil {
			requiredParams = 0
			for _, defaultVal := range helperMethod.DefaultValues {
				if defaultVal == nil {
					requiredParams++
				}
			}
		}

		// Check argument count is within valid range
		if len(expr.Arguments) < requiredParams || len(expr.Arguments) > len(helperMethod.Parameters) {
			if requiredParams == len(helperMethod.Parameters) {
				// All parameters are required
				a.addError("helper method '%s' expects %d arguments, got %d at %s",
					methodName, len(helperMethod.Parameters), len(expr.Arguments),
					expr.Token.Pos.String())
			} else {
				// Method has optional parameters
				a.addError("helper method '%s' expects %d-%d arguments, got %d at %s",
					methodName, requiredParams, len(helperMethod.Parameters), len(expr.Arguments),
					expr.Token.Pos.String())
			}
			return helperMethod.ReturnType
		}

		// Check argument types
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			// Task 9.218: Guard against nil Parameters (properties have no parameters)
			if helperMethod.Parameters != nil && i < len(helperMethod.Parameters) {
				expectedType := helperMethod.Parameters[i]
				if argType != nil && expectedType != nil && !a.canAssign(argType, expectedType) {
					a.addError("argument %d to helper method '%s' has type %s, expected %s at %s",
						i+1, methodName, argType.String(), expectedType.String(),
						expr.Token.Pos.String())
				}
			}
		}

		return helperMethod.ReturnType
	}

	// Handle built-in methods available on all objects (inherited from TObject)
	if methodName == "ClassName" {
		// ClassName() returns String
		return types.STRING
	}

	// Task 9.61: Check if method is overloaded
	var methodType *types.FunctionType
	methodOwner := a.getMethodOwner(classType, methodName)
	overloads := a.getMethodOverloadsInHierarchy(methodName, classType)

	if len(overloads) > 1 {
		// Method is overloaded - resolve based on argument types
		// Analyze argument types first
		argTypes := make([]types.Type, len(expr.Arguments))
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			if argType == nil {
				return nil // Error already reported
			}
			argTypes[i] = argType
		}

		// Convert MethodInfo to Symbol for ResolveOverload
		candidates := make([]*Symbol, len(overloads))
		for i, overload := range overloads {
			candidates[i] = &Symbol{
				Type: overload.Signature,
			}
		}

		// Resolve overload based on argument types
		selected, err := ResolveOverload(candidates, argTypes)
		if err != nil {
			// Task 9.63: Provide DWScript-compatible error message for failed overload resolution
			a.addError("Syntax Error: There is no overloaded version of \"%s\" that can be called with these arguments [line: %d, column: %d]",
				methodName, expr.Token.Pos.Line, expr.Token.Pos.Column)
			return nil
		}

		methodType = selected.Type.(*types.FunctionType)
	} else if len(overloads) == 1 {
		// Single method (not overloaded)
		methodType = overloads[0].Signature
	} else {
		// Method not found - check helpers
		_, helperMethod := a.hasHelperMethod(objectType, methodName)
		if helperMethod != nil {
			methodType = helperMethod
		} else {
			a.addError("class '%s' has no method '%s' at %s",
				classType.Name, methodName, expr.Token.Pos.String())
			return nil
		}
	}

	// Check method visibility - Task 9.16.1
	if methodOwner != nil && len(overloads) > 0 {
		// Use lowercase key for case-insensitive lookup
		visibility, hasVisibility := methodOwner.MethodVisibility[strings.ToLower(methodName)]
		if hasVisibility && !a.checkVisibility(methodOwner, visibility, methodName, "method") {
			visibilityStr := ast.Visibility(visibility).String()
			if methodOwner.HasConstructor(methodName) {
				a.addError("cannot access %s constructor '%s' of class '%s' at %s",
					visibilityStr, methodName, methodOwner.Name, expr.Token.Pos.String())
				return classType
			}
			a.addError("cannot call %s method '%s' of class '%s' at %s",
				visibilityStr, methodName, methodOwner.Name, expr.Token.Pos.String())
			return methodType.ReturnType
		}
	}

	// For non-overloaded methods, check argument types (overloaded methods already validated by ResolveOverload)
	if len(overloads) <= 1 {
		// Check argument count
		if len(expr.Arguments) != len(methodType.Parameters) {
			a.addError("method '%s' of class '%s' expects %d arguments, got %d at %s",
				methodName, classType.Name, len(methodType.Parameters), len(expr.Arguments),
				expr.Token.Pos.String())
			return methodType.ReturnType
		}

		// Check argument types
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			expectedType := methodType.Parameters[i]
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("argument %d to method '%s' of class '%s' has type %s, expected %s at %s",
					i+1, methodName, classType.Name, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		}
	}

	if classType.HasConstructor(methodName) {
		// Task 9.2: Check if trying to instantiate an abstract class via constructor call
		if classType.IsAbstract {
			a.addError("Trying to create an instance of an abstract class at [line: %d, column: %d]",
				expr.Token.Pos.Line, expr.Token.Pos.Column)
			return classType
		}

		// Task 9.2: Check if class has unimplemented abstract methods
		unimplementedMethods := a.getUnimplementedAbstractMethods(classType)
		if len(unimplementedMethods) > 0 {
			a.addError("Trying to create an instance of an abstract class at [line: %d, column: %d]",
				expr.Token.Pos.Line, expr.Token.Pos.Column)
			return classType
		}

		return classType
	}
	return methodType.ReturnType
}
