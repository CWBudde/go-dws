package semantic

import (
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
		methodType, found := interfaceType.GetMethod(methodName)

		// Check parent interfaces
		if !found && interfaceType.Parent != nil {
			allMethods := types.GetAllInterfaceMethods(interfaceType)
			methodType, found = allMethods[methodName]
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
			// Records can have methods - check if this method exists in the record itself
			method := recordType.GetMethod(methodName)
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

		// Validate helper method arguments
		if len(expr.Arguments) != len(helperMethod.Parameters) {
			a.addError("helper method '%s' expects %d arguments, got %d at %s",
				methodName, len(helperMethod.Parameters), len(expr.Arguments),
				expr.Token.Pos.String())
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

	// Check method visibility
	if methodOwner != nil && len(overloads) > 0 {
		visibility, hasVisibility := methodOwner.MethodVisibility[methodName]
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
		return classType
	}
	return methodType.ReturnType
}
