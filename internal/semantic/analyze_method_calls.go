package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
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
	methodNameLower := ident.Normalize(methodName)
	isMetaclass := false

	// Handle metaclass type (class of T) for constructor calls.
	// When we have TExample.CreateWith(...), unwrap ClassOfType to ClassType for constructor lookup.
	if metaclassType, ok := objectType.(*types.ClassOfType); ok {
		isMetaclass = true
		if metaclassType.ClassType != nil {
			objectType = metaclassType.ClassType
		}
	}

	// Check if object is an interface type
	if interfaceType, ok := objectType.(*types.InterfaceType); ok {
		// Look up method in interface (including inherited methods from parent interfaces)
		methodType, found := interfaceType.GetMethod(methodNameLower)

		// Check parent interfaces
		if !found && interfaceType.Parent != nil {
			allMethods := types.GetAllInterfaceMethods(interfaceType)
			methodType, found = allMethods[methodNameLower]
		}

		if !found {
			helperMethod := a.hasHelperMethod(objectType, methodName)
			if helperMethod == nil {
				a.addStructuredError(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, objectType.String()))
				return nil
			}
			methodType = helperMethod
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
		if arrayType, isArray := types.GetUnderlyingType(objectType).(*types.ArrayType); isArray {
			if result := a.analyzeArrayMethodCall(expr, arrayType); result != nil {
				return result
			}
		}

		// Check if object is a record type with methods
		if recordType, isRecord := objectType.(*types.RecordType); isRecord {
			// First check for class methods (static methods) with overload support.
			// On an instance receiver, same-named instance methods join the set.
			classOverloads := recordType.GetClassMethodOverloads(methodNameLower)
			if len(classOverloads) > 0 {
				classOverloads = append(append([]*types.MethodInfo{}, classOverloads...),
					recordType.GetMethodOverloads(methodNameLower)...)
				// This is a static method call - resolve overload based on arguments
				argTypes := make([]types.Type, len(expr.Arguments))
				for i, arg := range expr.Arguments {
					argType := a.analyzeOverloadArgument(arg)
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
					a.addStructuredError(NewNoOverloadMatchError(expr.Token.Pos, methodName))
					return nil
				}

				methodType, ok := selected.Type.(*types.FunctionType)
				if !ok {
					a.addError("internal error: expected function type for selected record static method, but got %T", selected.Type)
					return nil
				}

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

			// Check for instance methods (overload-aware)
			var method *types.FunctionType
			if instanceOverloads := recordType.GetMethodOverloads(methodNameLower); len(instanceOverloads) > 1 {
				argTypes := make([]types.Type, len(expr.Arguments))
				for i, arg := range expr.Arguments {
					argType := a.analyzeOverloadArgument(arg)
					if argType == nil {
						return nil
					}
					argTypes[i] = argType
				}
				candidates := make([]*Symbol, len(instanceOverloads))
				for i, overload := range instanceOverloads {
					candidates[i] = &Symbol{Type: overload.Signature}
				}
				selected, err := ResolveOverload(candidates, argTypes)
				if err != nil {
					a.addStructuredError(NewNoOverloadMatchError(expr.Token.Pos, methodName))
					return nil
				}
				method = selected.Type.(*types.FunctionType)
			} else {
				method = recordType.GetMethod(methodNameLower)
			}
			if method == nil {
				// Method not found in record, check if a helper provides it
				helperMethod := a.hasHelperMethod(objectType, methodName)
				if helperMethod == nil {
					a.addStructuredError(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, objectType.String()))
					return nil
				}
				// Use the helper method
				method = helperMethod
			}

			// Validate method arguments (defaulted parameters are optional)
			if len(expr.Arguments) > len(method.Parameters) ||
				len(expr.Arguments) < requiredParamCount(method) {
				a.addError("record method '%s' expects %d arguments, got %d at %s",
					methodName, len(method.Parameters), len(expr.Arguments),
					expr.Token.Pos.String())
				return method.ReturnType
			}

			// Check argument types (in the context of the selected signature,
			// so literals such as [] or nil adopt the parameter's type)
			for i, arg := range expr.Arguments {
				expectedType := method.Parameters[i]
				argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
				if argType != nil && !a.canAssign(argType, expectedType) {
					a.addError("argument %d to record method '%s' has type %s, expected %s at %s",
						i+1, methodName, argType.String(), expectedType.String(),
						expr.Token.Pos.String())
				}
			}

			return method.ReturnType
		}

		// Handle set types with built-in methods (Include/Exclude) without helpers
		if setType, isSet := types.GetUnderlyingType(objectType).(*types.SetType); isSet {
			switch methodNameLower {
			case "include", "exclude":
				if len(expr.Arguments) != 1 {
					a.addError("set method '%s' expects 1 argument, got %d at %s",
						methodName, len(expr.Arguments), expr.Token.Pos.String())
					return types.VOID
				}

				expectedElemType := setType.ElementType
				argType := a.analyzeExpressionWithExpectedType(expr.Arguments[0], expectedElemType)
				if argType != nil && expectedElemType != nil && !a.canAssign(argType, expectedElemType) {
					a.addError("argument 1 to set method '%s' has type %s, expected %s at %s",
						methodName, argType.String(), expectedElemType.String(), expr.Token.Pos.String())
				}
				return types.VOID
			default:
				a.addStructuredError(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, objectType.String()))
				return nil
			}
		}

		// Check if helpers provide this method for non-class, non-record types
		helperMethod := a.resolveHelperMethodForCall(objectType, methodName, expr.Arguments)
		if helperMethod == nil {
			a.addStructuredError(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, objectType.String()))
			return nil
		}

		// Record the receiver's static type so runtime helper dispatch honors
		// alias-specific (strict) helpers over the underlying type's helpers.
		if a.semanticInfo != nil && expr.Method != nil {
			a.semanticInfo.SetType(expr.Method, &ast.TypeAnnotation{
				Token: expr.Method.Token,
				Name:  "__helper_receiver:" + objectType.String(),
			})
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
			var expectedType types.Type
			if helperMethod.Parameters != nil && i < len(helperMethod.Parameters) {
				expectedType = helperMethod.Parameters[i]
			}
			// Use analyzeExpressionWithExpectedType to enable lambda parameter type inference
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			if expectedType != nil && argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("argument %d to helper method '%s' has type %s, expected %s at %s",
					i+1, methodName, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		}

		return helperMethod.ReturnType
	}

	// Handle built-in methods available on all objects (inherited from TObject)
	if methodName == "ClassName" {
		// ClassName() returns String
		return types.STRING
	}

	// Constructors are stored separately from methods and can be inherited.
	// The hierarchy lookup merges constructors with same-named class methods,
	// so the resolved overload decides whether this call is a construction.
	if constructorOverloads := a.getMethodOverloadsInHierarchy(methodName, classType); len(constructorOverloads) > 0 && classType.HasConstructor(methodName) {
		selectedInfo := constructorOverloads[0]

		if len(constructorOverloads) > 1 {
			argTypes := make([]types.Type, len(expr.Arguments))
			for i, arg := range expr.Arguments {
				argType := a.analyzeOverloadArgument(arg)
				if argType == nil {
					return classType
				}
				argTypes[i] = argType
			}

			candidates := make([]*Symbol, len(constructorOverloads))
			for i, overload := range constructorOverloads {
				candidates[i] = &Symbol{Type: overload.Signature}
			}

			selected, err := ResolveOverload(candidates, argTypes)
			if err != nil {
				a.addStructuredError(NewNoOverloadMatchError(expr.Token.Pos, methodName))
				return classType
			}

			selectedInfo = nil
			for i := range candidates {
				if candidates[i] == selected {
					selectedInfo = constructorOverloads[i]
					break
				}
			}
			if selectedInfo == nil {
				a.addError("internal error: resolved constructor overload not found in candidate list")
				return classType
			}
		}

		methodType := selectedInfo.Signature

		if len(expr.Arguments) > len(methodType.Parameters) ||
			len(expr.Arguments) < requiredParamCount(methodType) {
			a.addError("constructor '%s' of class '%s' expects %d arguments, got %d at %s",
				methodName, classType.Name, len(methodType.Parameters), len(expr.Arguments),
				expr.Token.Pos.String())
			return classType
		}

		for i, arg := range expr.Arguments {
			expectedType := methodType.Parameters[i]
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("argument %d to constructor '%s' of class '%s' has type %s, expected %s at %s",
					i+1, methodName, classType.Name, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		}

		// Resolved to a same-named class method rather than a constructor.
		if !selectedInfo.IsConstructor {
			return methodType.ReturnType
		}

		if classType.IsAbstract {
			a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
			return classType
		}

		if unimplementedMethods := a.getUnimplementedAbstractMethods(classType); len(unimplementedMethods) > 0 {
			a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
			return classType
		}

		return classType
	}

	// Check if method is overloaded
	var methodType *types.FunctionType
	var selectedOverload *types.MethodInfo
	methodOwner := a.getMethodOwner(classType, methodName)
	overloads := a.getMethodOverloadsInHierarchy(methodName, classType)

	if len(overloads) > 1 {
		// Method is overloaded - resolve based on argument types
		// Analyze argument types first
		argTypes := make([]types.Type, len(expr.Arguments))
		for i, arg := range expr.Arguments {
			argType := a.analyzeOverloadArgument(arg)
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
			a.addStructuredError(NewNoOverloadMatchError(expr.Token.Pos, methodName))
			return nil
		}

		var ok bool
		methodType, ok = selected.Type.(*types.FunctionType)
		if !ok {
			a.addError("internal error: expected function type for selected overloaded method, but got %T", selected.Type)
			return nil
		}
		for i := range candidates {
			if candidates[i] == selected {
				selectedOverload = overloads[i]
				break
			}
		}

		// Re-analyze arguments against the selected signature so literals get
		// their contextual type annotations (e.g. [o] becomes an array literal
		// of the parameter's element type instead of a set literal).
		for i, arg := range expr.Arguments {
			if i >= len(methodType.Parameters) {
				break
			}
			a.analyzeExpressionWithExpectedType(arg, methodType.Parameters[i])
		}
	} else if len(overloads) == 1 {
		// Single method (not overloaded). A method reached through a metaclass value must
		// be a class method or constructor. Use the resolved overload's flag so inherited
		// class methods (not present in this class's own ClassMethodFlags map) are accepted.
		if isMetaclass && !overloads[0].IsClassMethod {
			a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Method.Token.Pos))
			return nil
		}
		methodType = overloads[0].Signature
	} else {
		// Method not found - check helpers
		if isMetaclass {
			a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Method.Token.Pos))
			return nil
		}
		helperMethod := a.hasHelperMethod(objectType, methodName)
		if helperMethod != nil {
			methodType = helperMethod
		} else {
			a.addStructuredError(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, objectType.String()))
			return nil
		}
	}

	// Check method visibility. Overloads carry their own visibility, so the
	// SELECTED overload's visibility governs (overloads of one name can mix
	// private and public sections).
	if methodOwner != nil && len(overloads) > 0 {
		// Use lowercase key for case-insensitive lookup
		visibility, hasVisibility := methodOwner.MethodVisibility[ident.Normalize(methodName)]
		if selectedOverload != nil {
			visibility, hasVisibility = selectedOverload.Visibility, true
		}
		if hasVisibility && !a.checkVisibility(methodOwner, visibility, methodName, "method") {
			a.addStructuredError(NewVisibilityScopeError(expr.Method.Token.Pos, expr.Method.Value))
			if methodOwner.HasConstructor(methodName) {
				return classType
			}
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
		// Check if trying to instantiate an abstract class via constructor call
		if classType.IsAbstract {
			a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
			return classType
		}

		// Check if class has unimplemented abstract methods
		unimplementedMethods := a.getUnimplementedAbstractMethods(classType)
		if len(unimplementedMethods) > 0 {
			a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
			return classType
		}

		return classType
	}
	return methodType.ReturnType
}
