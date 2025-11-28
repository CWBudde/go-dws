package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalInOperator evaluates the 'in' operator for checking membership in sets, arrays, or strings
// Syntax: value in container
// Returns: Boolean indicating whether value is found in the container
func (i *Interpreter) evalInOperator(value Value, container Value, node ast.Node) Value {
	// Handle set membership (now supports all ordinal types)
	if setVal, ok := container.(*SetValue); ok {
		// Value must be an ordinal type to be in a set
		ordinal, err := evaluator.GetOrdinalValue(value)
		if err != nil {
			return i.newErrorWithLocation(node, "type mismatch: %s", err.Error())
		}
		// Use existing evalSetMembership function from set.go
		return i.evalSetMembership(value, ordinal, setVal)
	}

	// Handle string character membership: 'x' in 'abc'
	// This checks if a character/string is contained in another string
	if strContainer, ok := container.(*StringValue); ok {
		// Value must be a string (character)
		strValue, ok := value.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "type mismatch: %s in STRING", value.Type())
		}
		// Check if the string contains the character/substring
		// In DWScript, this is typically used for single characters
		// e.g., 'a' in 'abc' returns true
		if len(strValue.Value) > 0 {
			// Check if the container string contains the value string
			contains := strings.Contains(strContainer.Value, strValue.Value)
			return &BooleanValue{Value: contains}
		}
		// Empty string is not in any string
		return &BooleanValue{Value: false}
	}

	// Handle array membership (existing code)
	arrVal, ok := container.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(node, "type mismatch: %s in %s", value.Type(), container.Type())
	}

	// Search for the value in the array
	for _, elem := range arrVal.Elements {
		// Compare values for equality
		if i.valuesEqual(value, elem) {
			return &BooleanValue{Value: true}
		}
	}

	// Value not found
	return &BooleanValue{Value: false}
}

// evalIsExpression evaluates the 'is' operator which can be used for:
// 1. Type checking: obj is TMyClass -> Boolean
// 2. Boolean value comparison: boolExpr is True -> Boolean
// Returns true if the condition matches, false otherwise.
func (i *Interpreter) evalIsExpression(expr *ast.IsExpression) Value {
	// Check if this is a boolean value comparison (expr.Right is set)
	// or a type check (expr.TargetType is set)
	if expr.Right != nil {
		// Boolean value comparison: left is right
		// This is essentially checking if left == right for boolean values
		left := i.Eval(expr.Left)
		if isError(left) {
			return left
		}

		right := i.Eval(expr.Right)
		if isError(right) {
			return right
		}

		// Convert both to boolean values using variantToBool
		leftBool := variantToBool(left)
		rightBool := variantToBool(right)

		return &BooleanValue{Value: leftBool == rightBool}
	}

	// Type checking mode
	// Evaluate the left expression (the object to check)
	left := i.Eval(expr.Left)
	if isError(left) {
		return left
	}

	// Handle nil - nil is not an instance of any type
	if _, isNil := left.(*NilValue); isNil {
		return &BooleanValue{Value: false}
	}

	// Ensure we have an object instance
	obj, ok := AsObject(left)
	if !ok {
		// Not an object - return false
		return &BooleanValue{Value: false}
	}

	// Get the target type name from the type expression
	targetTypeName := ""
	if typeAnnotation, ok := expr.TargetType.(*ast.TypeAnnotation); ok {
		targetTypeName = typeAnnotation.Name
	} else {
		return i.newErrorWithLocation(expr, "cannot determine target type")
	}

	// First, check if the object's class matches or is derived from the target class
	// Walk up the class hierarchy
	currentClass := obj.Class
	for currentClass != nil {
		if ident.Equal(currentClass.Name, targetTypeName) {
			return &BooleanValue{Value: true}
		}
		// Move to parent class
		currentClass = currentClass.Parent
	}

	// If not a class match, check if the target is an interface
	// and if the object's class implements it
	if iface, exists := i.interfaces[ident.Normalize(targetTypeName)]; exists {
		result := classImplementsInterface(obj.Class, iface)
		return &BooleanValue{Value: result}
	}

	return &BooleanValue{Value: false}
}

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
				if ident.Equal(currentClass.Name, targetClass.Name) {
					isCompatible = true
					break
				}
				currentClass = currentClass.Parent
			}

			if !isCompatible {
				// Throw exception with proper format including position
				message := fmt.Sprintf("Cannot cast interface of '%s' to class '%s' [line: %d, column: %d]",
					underlyingObj.Class.Name, targetClass.Name, expr.Token.Pos.Line, expr.Token.Pos.Column)
				i.raiseException("Exception", message, &expr.Token.Pos)
				return nil
			}

			// Cast is valid - return the underlying object
			return underlyingObj
		}

		// Check if target is an interface
		if targetIface, isInterface := i.interfaces[ident.Normalize(targetTypeName)]; isInterface {
			// Interface-to-interface casting
			underlyingObj := intfInst.Object
			if underlyingObj == nil {
				// DWScript: nil interface cast to interface yields nil (no exception)
				return &InterfaceInstance{Interface: targetIface, Object: nil}
			}

			// Check if the underlying object's class implements the target interface
			if !classImplementsInterface(underlyingObj.Class, targetIface) {
				message := fmt.Sprintf("Cannot cast interface of \"%s\" to interface \"%s\" [line: %d, column: %d]",
					underlyingObj.Class.Name, targetIface.Name, expr.Token.Pos.Line, expr.Token.Pos.Column)
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
			if ident.Equal(currentClass.Name, targetClass.Name) {
				isCompatible = true
				break
			}
			currentClass = currentClass.Parent
		}

		// If the object's runtime type doesn't match or derive from target, the cast is invalid
		if !isCompatible {
			message := fmt.Sprintf("instance of type \"%s\" cannot be cast to class \"%s\" [line: %d, column: %d]",
				obj.Class.Name, targetClass.Name, expr.Token.Pos.Line, expr.Token.Pos.Column)
			i.raiseException("Exception", message, &expr.Token.Pos)
			return nil
		}

		// Cast is valid - return the same object
		// The object's runtime type is guaranteed to be compatible with the target
		return obj
	}

	// Try interface casting
	// Look up the interface in the registry
	iface, exists := i.interfaces[ident.Normalize(targetTypeName)]
	if !exists {
		return i.newErrorWithLocation(expr, "type '%s' not found (neither class nor interface)", targetTypeName)
	}

	// Validate that the object's class implements the interface
	if !classImplementsInterface(obj.Class, iface) {
		message := fmt.Sprintf("Class \"%s\" does not implement interface \"%s\" [line: %d, column: %d]",
			obj.Class.Name, iface.Name, expr.Token.Pos.Line, expr.Token.Pos.Column)
		i.raiseException("Exception", message, &expr.Token.Pos)
		return nil
	}

	// Create and return the interface instance
	return NewInterfaceInstance(iface, obj)
}

// evalImplementsExpression evaluates the 'implements' operator.
// Example: obj implements IMyInterface -> Boolean
// Returns true if the object's class implements the interface.
func (i *Interpreter) evalImplementsExpression(expr *ast.ImplementsExpression) Value {
	// Evaluate the left expression (the object or class to check)
	left := i.Eval(expr.Left)
	if isError(left) {
		return left
	}

	// Handle nil - nil implements no interfaces
	if _, isNil := left.(*NilValue); isNil {
		return &BooleanValue{Value: false}
	}

	// The 'implements' operator can work with:
	// 1. Object instances (extract class from instance)
	// 2. Class type references (ClassValue from metaclass variables)
	// 3. Class type identifiers (ClassInfoValue from class names)
	var classInfo *ClassInfo

	if obj, ok := AsObject(left); ok {
		// Object instance - extract class
		classInfo = obj.Class
	} else if classVal, ok := left.(*ClassValue); ok {
		// Class reference (e.g., from metaclass variable: var cls: class of TParent)
		classInfo = classVal.ClassInfo
	} else if classInfoVal, ok := left.(*ClassInfoValue); ok {
		// Class type identifier (e.g., TMyImplementation in: if TMyImplementation implements IMyInterface then)
		classInfo = classInfoVal.ClassInfo
	} else {
		return i.newErrorWithLocation(expr, "'implements' operator requires object instance or class type, got %s", left.Type())
	}

	// Get the target interface name from the type expression
	targetInterfaceName := ""
	if typeAnnotation, ok := expr.TargetType.(*ast.TypeAnnotation); ok {
		targetInterfaceName = typeAnnotation.Name
	} else {
		return i.newErrorWithLocation(expr, "cannot determine target interface type")
	}

	// Look up the interface in the registry
	iface, exists := i.interfaces[ident.Normalize(targetInterfaceName)]
	if !exists {
		return i.newErrorWithLocation(expr, "interface '%s' not found", targetInterfaceName)
	}

	// Guard against nil ClassInfo (e.g., uninitialized metaclass variables)
	// Return false instead of panicking when classInfo is nil
	if classInfo == nil {
		return &BooleanValue{Value: false}
	}

	// Check if the class implements the interface (explicit declarations only for DWScript 'implements')
	result := classExplicitlyImplementsInterface(classInfo, iface)
	return &BooleanValue{Value: result}
}

// evalIfExpression evaluates an inline if-then-else conditional expression.
// Syntax: if <condition> then <expression> [else <expression>]
// Returns the value of the consequence if condition is true, otherwise the alternative (or default value).
func (i *Interpreter) evalIfExpression(expr *ast.IfExpression) Value {
	// Evaluate the condition
	condition := i.Eval(expr.Condition)
	if isError(condition) {
		return condition
	}

	// Use isTruthy to support Variant→Boolean implicit conversion
	// If condition is true, evaluate and return consequence
	if isTruthy(condition) {
		result := i.Eval(expr.Consequence)
		if isError(result) {
			return result
		}
		return result
	}

	// Condition is false
	if expr.Alternative != nil {
		// Evaluate and return alternative
		result := i.Eval(expr.Alternative)
		if isError(result) {
			return result
		}
		return result
	}

	// No else clause - return default value for the consequence type
	// The type should have been set during semantic analysis
	var typeAnnot *ast.TypeAnnotation
	if i.semanticInfo != nil {
		typeAnnot = i.semanticInfo.GetType(expr)
	}
	if typeAnnot == nil {
		return i.newErrorWithLocation(expr, "if expression missing type annotation")
	}

	// Return default value based on type name
	typeName := ident.Normalize(typeAnnot.Name)
	switch typeName {
	case "integer", "int64":
		return &IntegerValue{Value: 0}
	case "float", "float64", "double", "real":
		return &FloatValue{Value: 0.0}
	case "string":
		return &StringValue{Value: ""}
	case "boolean", "bool":
		return &BooleanValue{Value: false}
	default:
		// For class types and other reference types, return nil
		return &NilValue{}
	}
}
