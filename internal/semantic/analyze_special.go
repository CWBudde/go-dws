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
	if ie.Truncated {
		return a.analyzeTruncatedCall(ie.Arguments)
	}
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

	callPos := ie.Token.Pos
	if ie.Method != nil {
		callPos = ie.Method.Token.Pos
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

			selected := a.selectMemberCallOverload(ctorOverloads, ie.Arguments, memberName, callPos)
			if selected == nil {
				return types.VOID
			}
			a.analyzeClassCallArguments(selected.Signature, ie.Arguments, callPos)

			// Constructors don't have explicit return types in expressions
			return types.VOID
		}
	}

	// Named scalar property reads select the nearest parent member before
	// the method-only inherited path searches farther up the hierarchy.
	if result, handled := a.analyzeInheritedPropertyRead(ie, parentClass); handled {
		return result
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
			overloads := a.getMethodOverloadsInHierarchy(memberName, parentClass)
			selected := a.selectMemberCallOverload(overloads, ie.Arguments, memberName, callPos)
			if selected == nil {
				return nil
			}
			methodType = selected.Signature
			a.analyzeClassCallArguments(methodType, ie.Arguments, callPos)

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
			callPos := ie.Token.Pos
			if ie.Method != nil {
				callPos = ie.Method.Token.Pos
			}
			for idx := range ie.Arguments {
				a.analyzeSelfCallArgument(idx, ie.Arguments, methodType.Parameters[idx], callPos,
					idx < len(methodType.StrictParams) && methodType.StrictParams[idx])
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
	// Inside a property read/write expression, Self is available even though no
	// method frame is active. Class properties see the metaclass; instance
	// properties see the instance type.
	if a.inPropertyExpr && a.currentFunction == nil {
		if a.currentClass != nil {
			if a.inClassMethod {
				return types.NewClassOfType(a.currentClass)
			}
			return a.currentClass
		}
		if a.currentRecord != nil {
			return a.currentRecord
		}
	}

	// A static helper class method has no Self at all: upstream reports the
	// name as simply unknown (HelpersFail/static_class_method_self).
	if a.inStaticHelperMethod {
		a.addStructuredError(NewUnknownNameError(se.Token.Pos, "Self"))
		return nil
	}

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
		// Ordinary record class methods are static and have no Self.
		if a.inClassMethod {
			a.addStructuredError(NewUnknownNameError(se.Token.Pos, "Self"))
			return nil
		}
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
