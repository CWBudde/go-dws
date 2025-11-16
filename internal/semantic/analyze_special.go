package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Expression Analysis
// ============================================================================

// analyzeInheritedExpression analyzes an inherited expression and returns its type.
func (a *Analyzer) analyzeInheritedExpression(ie *ast.InheritedExpression) types.Type {
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
		ctorType, ctorFound := parentClass.GetConstructor(memberName)
		if ctorFound {
			// This is an inherited constructor call
			if ie.IsCall || len(ie.Arguments) >= 0 {
				// Check argument count
				expectedParams := len(ctorType.Parameters)
				actualArgs := len(ie.Arguments)
				if actualArgs != expectedParams {
					a.addError("wrong number of arguments for inherited constructor '%s': expected %d, got %d at %s",
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

					paramType := ctorType.Parameters[idx]
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
	}

	// Try to find as a method first
	methodType, methodFound := parentClass.GetMethod(memberName)
	if methodFound {
		// Task 9.14.2: In DWScript, inherited MethodName without parens is still a call if method takes no params
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
	isTObjectParent := strings.EqualFold(parentClass.Name, "TObject")
	if isTObjectParent {
		a.addError("'inherited' cannot be used in class '%s' which has no parent class at %s",
			a.currentClass.Name, ie.Token.Pos.String())
	} else {
		a.addError("method, property, or field '%s' not found in parent class '%s' at %s",
			memberName, parentClass.Name, ie.Token.Pos.String())
	}
	return nil
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

	// Validate that we're in a class context
	if a.currentClass == nil {
		a.addError("'Self' can only be used inside a class method at %s", se.Token.Pos.String())
		return nil
	}

	// Class methods (static methods) cannot access Self
	if a.inClassMethod {
		a.addError("'Self' cannot be used in class methods (static methods) at %s", se.Token.Pos.String())
		return nil
	}

	// For instance methods, Self has the type of the class
	return a.currentClass
}
