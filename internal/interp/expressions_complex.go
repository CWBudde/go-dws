package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalAsExpression evaluates the 'as' type casting operator
// Example: obj as IMyInterface
// Creates an InterfaceInstance wrapper around the object.
func (i *Interpreter) evalAsExpression(expr *ast.AsExpression) Value {
	// Evaluate the left expression (the object to cast)
	left := i.Eval(expr.Left)
	if isError(left) {
		return left
	}

	// Get the target type name from the type expression
	targetTypeName := ""
	if typeAnnotation, ok := expr.TargetType.(*ast.TypeAnnotation); ok {
		targetTypeName = typeAnnotation.Name
	} else {
		return i.newErrorWithLocation(expr, "cannot determine target type")
	}
	targetLower := ident.Normalize(targetTypeName)

	// Variant-specific casting: allow using 'as' to coerce to primitive types
	if variantVal, ok := left.(*VariantValue); ok {
		switch targetLower {
		case "integer":
			return i.castToInteger(variantVal)
		case "float":
			return i.castToFloat(variantVal)
		case "string":
			return i.castToString(variantVal)
		case "boolean":
			return i.castToBoolean(variantVal)
		case "variant":
			return variantVal
		}

		// For class/interface targets, unwrap and continue with normal logic
		left = variantVal.Value
		if left == nil {
			left = &UnassignedValue{}
		}
	}

	// Handle nil specially - nil can be cast to any class or interface
	if _, isNil := left.(*NilValue); isNil {
		return &NilValue{}
	}

	// Handle interface-to-object/interface casting
	// If left side is an InterfaceInstance, we need special handling
	if intfInst, ok := left.(*InterfaceInstance); ok {
		// Check if target is a class
		// PR #147: Use lowercase key for O(1) case-insensitive lookup
		if targetClass, isClass := i.classes[ident.Normalize(targetTypeName)]; isClass {
			// Interface-to-class casting
			// Extract the underlying object
			underlyingObj := intfInst.Object
			if underlyingObj == nil {
				return i.newErrorWithLocation(expr, "cannot cast nil interface to class '%s'", targetClass.Name)
			}

			// Check if the underlying object's class is compatible with the target class
			currentClass := underlyingObj.Class
			isCompatible := false
			for currentClass != nil {
				if ident.Equal(currentClass.GetName(), targetClass.Name) {
					isCompatible = true
					break
				}
				currentClass = currentClass.GetParent()
			}

			if !isCompatible {
				// Throw exception with proper format including position
				message := fmt.Sprintf("Cannot cast interface of '%s' to class '%s' [line: %d, column: %d]",
					underlyingObj.Class.GetName(), targetClass.Name, expr.Token.Pos.Line, expr.Token.Pos.Column)
				i.raiseException("Exception", message, &expr.Token.Pos)
				return nil
			}

			// Cast is valid - return the underlying object
			return underlyingObj
		}

		// Check if target is an interface
		if targetIface := i.lookupInterfaceInfo(targetTypeName); targetIface != nil {
			// Interface-to-interface casting
			underlyingObj := intfInst.Object
			if underlyingObj == nil {
				// DWScript: nil interface cast to interface yields nil (no exception)
				return &InterfaceInstance{Interface: targetIface, Object: nil}
			}

			// Check if the underlying object's class implements the target interface
			// Need concrete ClassInfo for classImplementsInterface
			concreteClass, ok := underlyingObj.Class.(*ClassInfo)
			if !ok {
				return i.newErrorWithLocation(expr, "object has invalid class type")
			}
			if !classImplementsInterface(concreteClass, targetIface) {
				message := fmt.Sprintf("Cannot cast interface of \"%s\" to interface \"%s\" [line: %d, column: %d]",
					underlyingObj.Class.GetName(), targetIface.Name, expr.Token.Pos.Line, expr.Token.Pos.Column)
				i.raiseException("Exception", message, &expr.Token.Pos)
				return nil
			}

			// Create and return new interface instance
			return NewInterfaceInstance(targetIface, underlyingObj)
		}

		return i.newErrorWithLocation(expr, "type '%s' not found (neither class nor interface)", targetTypeName)
	}

	// Ensure we have an object instance
	obj, ok := AsObject(left)
	if !ok {
		return i.newErrorWithLocation(expr, "'as' operator requires object instance, got %s", left.Type())
	}

	// Try class-to-class casting first
	// Look up the target as a class
	// PR #147: Use lowercase key for O(1) case-insensitive lookup
	targetClass, isClass := i.classes[ident.Normalize(targetTypeName)]
	if isClass {
		// This is a class-to-class cast
		// Validate that the object's actual runtime type is compatible with the target
		// Walk up the object's class hierarchy to check if it matches or derives from target
		currentClass := obj.Class
		isCompatible := false
		for currentClass != nil {
			if ident.Equal(currentClass.GetName(), targetClass.Name) {
				isCompatible = true
				break
			}
			currentClass = currentClass.GetParent()
		}

		// If the object's runtime type doesn't match or derive from target, the cast is invalid
		if !isCompatible {
			message := fmt.Sprintf("instance of type \"%s\" cannot be cast to class \"%s\" [line: %d, column: %d]",
				obj.Class.GetName(), targetClass.Name, expr.Token.Pos.Line, expr.Token.Pos.Column)
			i.raiseException("Exception", message, &expr.Token.Pos)
			return nil
		}

		// Cast is valid - return the same object
		// The object's runtime type is guaranteed to be compatible with the target
		return obj
	}

	// Try interface casting
	// Look up the interface in the registry
	iface := i.lookupInterfaceInfo(targetTypeName)
	if iface == nil {
		return i.newErrorWithLocation(expr, "type '%s' not found (neither class nor interface)", targetTypeName)
	}

	// Validate that the object's class implements the interface
	// Need concrete ClassInfo for classImplementsInterface
	concreteClass, ok := obj.Class.(*ClassInfo)
	if !ok {
		return i.newErrorWithLocation(expr, "object has invalid class type")
	}
	if !classImplementsInterface(concreteClass, iface) {
		message := fmt.Sprintf("Class \"%s\" does not implement interface \"%s\" [line: %d, column: %d]",
			obj.Class.GetName(), iface.Name, expr.Token.Pos.Line, expr.Token.Pos.Column)
		i.raiseException("Exception", message, &expr.Token.Pos)
		return nil
	}

	// Create and return the interface instance
	return NewInterfaceInstance(iface, obj)
}
