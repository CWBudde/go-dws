package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Expression Analysis
// ============================================================================
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

	// Try to find as a method first
	methodType, methodFound := parentClass.GetMethod(memberName)
	if methodFound {
		// Check if this is a method call (has arguments or IsCall flag)
		if ie.IsCall || len(ie.Arguments) > 0 {
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
	a.addError("method, property, or field '%s' not found in parent class '%s' at %s",
		memberName, parentClass.Name, ie.Token.Pos.String())
	return nil
}
