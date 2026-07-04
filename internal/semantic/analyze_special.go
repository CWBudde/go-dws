package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Expression Analysis
// ============================================================================

// analyzeInheritedExpression analyzes an inherited expression and returns its type.
func (a *Analyzer) analyzeInheritedExpression(ie *ast.InheritedExpression) types.Type {
	if a.currentHelperType != nil {
		return a.analyzeHelperInheritedExpression(ie, a.currentHelperType)
	}

	// Validate that we're in a method context (must have currentClass set)
	if a.currentClass == nil {
		a.addError("'inherited' can only be used inside a class method at %s", ie.Token.Pos.String())
		return nil
	}

	// Verify that the current class has a parent
	if a.currentClass.Parent == nil {
		a.addError("'inherited' cannot be used in class '%s' which has no parent class at %s",
			a.currentClass.Name, ie.Token.Pos.String())
		return nil
	}

	parentClass := a.currentClass.Parent

	// Determine which method/property to look up
	var memberName string
	if ie.Method != nil {
		// Explicit method/property name: inherited MethodName or inherited MethodName(args)
		memberName = ie.Method.Value
	} else {
		// Bare inherited: need to get current method name from currentFunction
		if a.currentFunction == nil {
			a.addError("bare 'inherited' requires method context at %s", ie.Token.Pos.String())
			return nil
		}
		memberName = a.currentFunction.Name.Value
	}

	// Check if we're calling a constructor from within a constructor
	// If we're in a constructor and the member is a constructor in the parent, handle it specially
	if a.currentFunction != nil && a.currentFunction.IsConstructor {
		if _, ctorFound := parentClass.GetConstructor(memberName); ctorFound {
			// Collect the parent's constructor overload set and resolve against
			// the provided arguments.
			var ctorOverloads []*types.MethodInfo
			for _, overload := range a.getMethodOverloadsInHierarchy(memberName, parentClass) {
				if overload.IsConstructor {
					ctorOverloads = append(ctorOverloads, overload)
				}
			}

			var ctorType *types.FunctionType
			if len(ctorOverloads) > 1 {
				argTypes := make([]types.Type, len(ie.Arguments))
				for idx, arg := range ie.Arguments {
					argType := a.analyzeOverloadArgument(arg)
					if argType == nil {
						return types.VOID
					}
					argTypes[idx] = argType
				}
				candidates := make([]*Symbol, len(ctorOverloads))
				for idx, overload := range ctorOverloads {
					candidates[idx] = &Symbol{Type: overload.Signature}
				}
				selected, err := ResolveOverload(candidates, argTypes)
				if err != nil {
					a.addStructuredError(NewNoOverloadMatchError(ie.Token.Pos, memberName))
					return types.VOID
				}
				ctorType = selected.Type.(*types.FunctionType)
			} else if len(ctorOverloads) == 1 {
				ctorType = ctorOverloads[0].Signature
			} else {
				ctorType, _ = parentClass.GetConstructor(memberName)
			}

			// Check argument count (defaulted parameters are optional)
			actualArgs := len(ie.Arguments)
			if actualArgs > len(ctorType.Parameters) || actualArgs < requiredParamCount(ctorType) {
				a.addError("wrong number of arguments for inherited constructor '%s': expected %d, got %d at %s",
					memberName, len(ctorType.Parameters), actualArgs, ie.Token.Pos.String())
				return nil
			}

			// Type check each argument (in the context of the selected
			// signature, so literals such as [] or nil adopt the parameter's type)
			for idx, arg := range ie.Arguments {
				paramType := ctorType.Parameters[idx]
				argType := a.analyzeExpressionWithExpectedType(arg, paramType)
				if argType == nil {
					// Error already reported
					continue
				}
				// Check type compatibility
				if !a.canAssign(argType, paramType) {
					a.addError("argument %d to inherited constructor '%s' has type %s, expected %s at %s",
						idx+1, memberName, argType.String(), paramType.String(), ie.Token.Pos.String())
				}
			}

			// Constructors don't have explicit return types in expressions
			return types.VOID
		}
	}

	// Try to find as a method first
	methodType, methodFound := parentClass.GetMethod(memberName)
	if methodFound {
		// In DWScript, inherited MethodName without parens is still a call if method takes no params
		// Determine if this should be treated as a call
		isMethodCall := ie.IsCall || len(ie.Arguments) > 0

		// Also treat as a call if method name is specified without parens but method exists
		// This matches DWScript semantics where parameterless methods can be called without parens
		if !isMethodCall && ie.Method != nil && len(methodType.Parameters) == 0 {
			isMethodCall = true
		}

		if isMethodCall {
			// Check argument count
			expectedParams := len(methodType.Parameters)
			actualArgs := len(ie.Arguments)
			if actualArgs != expectedParams {
				a.addError("wrong number of arguments for inherited method '%s': expected %d, got %d at %s",
					memberName, expectedParams, actualArgs, ie.Token.Pos.String())
				return nil
			}

			// Type check each argument
			for idx, arg := range ie.Arguments {
				argType := a.analyzeExpression(arg)
				if argType == nil {
					// Error already reported
					continue
				}

				paramType := methodType.Parameters[idx]
				// Check type compatibility (allow implicit conversions via canAssign)
				if !a.canAssign(argType, paramType) {
					a.addError("argument %d to inherited method '%s' has type %s, expected %s at %s",
						idx+1, memberName, argType.String(), paramType.String(), ie.Token.Pos.String())
				}
			}

			// Return the method's return type
			if methodType.ReturnType != nil {
				return methodType.ReturnType
			}
			return types.VOID
		}

		// Method reference (not a call) - return the method type
		return methodType
	}

	// Try to find as a property
	propInfo, propFound := parentClass.GetProperty(memberName)
	if propFound {
		// Property access returns the property's type
		if ie.IsCall || len(ie.Arguments) > 0 {
			a.addError("cannot call property '%s' as a method at %s",
				memberName, ie.Token.Pos.String())
			return nil
		}
		return propInfo.Type
	}

	// Try to find as a field
	fieldType, fieldFound := parentClass.GetField(memberName)
	if fieldFound {
		if ie.IsCall || len(ie.Arguments) > 0 {
			a.addError("cannot call field '%s' as a method at %s",
				memberName, ie.Token.Pos.String())
			return nil
		}
		return fieldType
	}

	// Member not found in parent class
	// If parent is TObject and member not found, treat as "no meaningful parent"
	isTObjectParent := ident.Equal(parentClass.Name, "TObject")
	if isTObjectParent {
		a.addError("'inherited' cannot be used in class '%s' which has no parent class at %s",
			a.currentClass.Name, ie.Token.Pos.String())
	} else {
		a.addError("method, property, or field '%s' not found in parent class '%s' at %s",
			memberName, parentClass.Name, ie.Token.Pos.String())
	}
	return nil
}

func (a *Analyzer) analyzeHelperInheritedExpression(ie *ast.InheritedExpression, helperType *types.HelperType) types.Type {
	if helperType == nil {
		a.addError("'inherited' can only be used inside a helper method at %s", ie.Token.Pos.String())
		return nil
	}

	memberName := ""
	if ie.Method != nil {
		memberName = ie.Method.Value
	} else {
		if a.currentFunction == nil {
			a.addError("bare 'inherited' requires method context at %s", ie.Token.Pos.String())
			return nil
		}
		memberName = a.currentFunction.Name.Value
	}

	candidates := a.helperInheritedCandidates(helperType)
	for _, candidate := range candidates {
		if methodType := helperMethodType(candidate, memberName); methodType != nil {
			if len(ie.Arguments) != len(methodType.Parameters) {
				a.addError("wrong number of arguments for inherited helper method '%s': expected %d, got %d at %s",
					memberName, len(methodType.Parameters), len(ie.Arguments), ie.Token.Pos.String())
				return nil
			}
			for idx, arg := range ie.Arguments {
				argType := a.analyzeExpression(arg)
				if argType != nil && !a.canAssign(argType, methodType.Parameters[idx]) {
					a.addError("argument %d to inherited helper method '%s' has type %s, expected %s at %s",
						idx+1, memberName, argType.String(), methodType.Parameters[idx].String(), ie.Token.Pos.String())
				}
			}
			if methodType.ReturnType != nil {
				return methodType.ReturnType
			}
			return types.VOID
		}
		if propType := helperPropertyType(candidate, memberName); propType != nil {
			if ie.IsCall || len(ie.Arguments) > 0 {
				a.addError("cannot call property '%s' as a method at %s", memberName, ie.Token.Pos.String())
				return nil
			}
			return propType
		}
	}

	if classType, ok := types.GetUnderlyingType(helperType.TargetType).(*types.ClassType); ok {
		if ident.Equal(memberName, "ClassName") {
			return types.STRING
		}
		if ident.Equal(memberName, "ClassType") {
			return types.NewClassOfType(classType)
		}
		if methodType, found := classType.GetMethod(memberName); found {
			return methodType.ReturnType
		}
		if propType, found := classType.GetProperty(memberName); found {
			return propType.Type
		}
		if fieldType, found := classType.GetField(memberName); found {
			return fieldType
		}
	}

	a.addError("method, property, or field '%s' not found for inherited helper lookup at %s",
		memberName, ie.Token.Pos.String())
	return nil
}

func (a *Analyzer) helperInheritedCandidates(helperType *types.HelperType) []*types.HelperType {
	var candidates []*types.HelperType
	seen := map[*types.HelperType]bool{helperType: true}
	if helperType.ParentHelper != nil {
		candidates = append(candidates, helperType.ParentHelper)
		seen[helperType.ParentHelper] = true
	}
	for _, candidate := range a.getHelpersForType(helperType.TargetType) {
		if candidate == nil || seen[candidate] {
			continue
		}
		candidates = append(candidates, candidate)
		seen[candidate] = true
	}
	return candidates
}

func helperMethodType(helperType *types.HelperType, name string) *types.FunctionType {
	if helperType == nil {
		return nil
	}
	if overloads := helperType.MethodOverloads[ident.Normalize(name)]; len(overloads) > 0 {
		return overloads[len(overloads)-1]
	}
	for methodName, methodType := range helperType.Methods {
		if ident.Equal(methodName, name) {
			return methodType
		}
	}
	return helperMethodType(helperType.ParentHelper, name)
}

func helperPropertyType(helperType *types.HelperType, name string) types.Type {
	if helperType == nil {
		return nil
	}
	for propName, propInfo := range helperType.Properties {
		if ident.Equal(propName, name) && propInfo != nil {
			return propInfo.Type
		}
	}
	return helperPropertyType(helperType.ParentHelper, name)
}

// analyzeSelfExpression analyzes a Self expression and returns its type.
// Self refers to the current instance in instance methods.
// Self is NOT allowed in class methods (static methods).
func (a *Analyzer) analyzeSelfExpression(se *ast.SelfExpression) types.Type {
	// Validate that we're in a method context
	if a.currentFunction == nil {
		a.addError("'Self' can only be used inside a method at %s", se.Token.Pos.String())
		return nil
	}
	if a.currentSelfType != nil {
		return a.currentSelfType
	}

	// Validate that we're in a class or record context
	if a.currentClass == nil {
		if a.currentRecord == nil {
			a.addError("'Self' can only be used inside a class method at %s", se.Token.Pos.String())
			return nil
		}
		// Record methods allow Self as the current record type
		return a.currentRecord
	}

	// Class methods (static methods) cannot access Self
	if a.inClassMethod {
		if len(a.getHelpersForType(a.currentClass)) > 0 {
			return a.currentClass
		}
		a.addError("'Self' cannot be used in class methods (static methods) at %s", se.Token.Pos.String())
		return nil
	}

	// For instance methods, Self has the type of the class
	return a.currentClass
}
