package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
)

// builtinInc implements the Inc() built-in function.
// It increments a variable in place: Inc(x) or Inc(x, delta)
// Supports any lvalue: Inc(x), Inc(arr[i]), Inc(obj.field)
func (i *Interpreter) builtinInc(args []ast.Expression) Value {
	// Validate argument count (1-2 arguments)
	if len(args) < 1 || len(args) > 2 {
		return i.newErrorWithLocation(i.currentNode, "Inc() expects 1-2 arguments, got %d", len(args))
	}

	// Get delta (default 1)
	delta := int64(1)
	if len(args) == 2 {
		deltaVal := i.Eval(args[1])
		if isError(deltaVal) {
			return deltaVal
		}
		deltaInt, ok := deltaVal.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "Inc() delta must be Integer, got %s", deltaVal.Type())
		}
		delta = deltaInt.Value
	}

	// Get current value of the lvalue
	currentVal := i.Eval(args[0])
	if isError(currentVal) {
		return currentVal
	}

	// Unwrap ReferenceValue if needed
	if ref, isRef := currentVal.(*ReferenceValue); isRef {
		actualVal, err := ref.Dereference()
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "%s", err.Error())
		}
		currentVal = actualVal
	}

	// Compute new value based on type
	var newValue Value

	switch val := currentVal.(type) {
	case *IntegerValue:
		// Increment integer by delta
		newValue = &IntegerValue{Value: val.Value + delta}

	case *EnumValue:
		// For enums, delta must be 1 (get successor)
		if delta != 1 {
			return i.newErrorWithLocation(i.currentNode, "Inc() with delta not supported for enum types")
		}

		// Get the enum type metadata
		// Normalize to lowercase for case-insensitive lookups
		enumTypeKey := "__enum_type_" + strings.ToLower(val.TypeName)
		enumTypeVal, ok := i.env.Get(enumTypeKey)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "enum type metadata not found for %s", val.TypeName)
		}

		enumTypeWrapper, ok := enumTypeVal.(*EnumTypeValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "invalid enum type metadata for %s", val.TypeName)
		}

		enumType := enumTypeWrapper.EnumType

		// Find current value's position in OrderedNames
		currentPos := -1
		for idx, name := range enumType.OrderedNames {
			if name == val.ValueName {
				currentPos = idx
				break
			}
		}

		if currentPos == -1 {
			return i.newErrorWithLocation(i.currentNode, "enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
		}

		// Check if we can increment (not at the end)
		if currentPos >= len(enumType.OrderedNames)-1 {
			return i.newErrorWithLocation(i.currentNode, "Inc() cannot increment enum beyond its maximum value")
		}

		// Get next value
		nextValueName := enumType.OrderedNames[currentPos+1]
		nextOrdinal := enumType.Values[nextValueName]

		// Create new enum value
		newValue = &EnumValue{
			TypeName:     val.TypeName,
			ValueName:    nextValueName,
			OrdinalValue: nextOrdinal,
		}

	default:
		return i.newErrorWithLocation(i.currentNode, "Inc() expects Integer or Enum, got %s", val.Type())
	}

	// Assign the new value back to the lvalue
	if err := i.assignToLValue(args[0], newValue); err != nil {
		return i.newErrorWithLocation(i.currentNode, "Inc() failed to assign: %s", err.Error())
	}

	// Return the new value (allows Inc to be used in expressions)
	return newValue
}

// builtinDec implements the Dec() built-in function.
// It decrements a variable in place: Dec(x) or Dec(x, delta)
// Supports any lvalue: Dec(x), Dec(arr[i]), Dec(obj.field)
func (i *Interpreter) builtinDec(args []ast.Expression) Value {
	// Validate argument count (1-2 arguments)
	if len(args) < 1 || len(args) > 2 {
		return i.newErrorWithLocation(i.currentNode, "Dec() expects 1-2 arguments, got %d", len(args))
	}

	// Get delta (default 1)
	delta := int64(1)
	if len(args) == 2 {
		deltaVal := i.Eval(args[1])
		if isError(deltaVal) {
			return deltaVal
		}
		deltaInt, ok := deltaVal.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "Dec() delta must be Integer, got %s", deltaVal.Type())
		}
		delta = deltaInt.Value
	}

	// Get current value of the lvalue
	currentVal := i.Eval(args[0])
	if isError(currentVal) {
		return currentVal
	}

	// Unwrap ReferenceValue if needed
	if ref, isRef := currentVal.(*ReferenceValue); isRef {
		actualVal, err := ref.Dereference()
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "%s", err.Error())
		}
		currentVal = actualVal
	}

	// Compute new value based on type
	var newValue Value

	switch val := currentVal.(type) {
	case *IntegerValue:
		// Decrement integer by delta
		newValue = &IntegerValue{Value: val.Value - delta}

	case *EnumValue:
		// For enums, delta must be 1 (get predecessor)
		if delta != 1 {
			return i.newErrorWithLocation(i.currentNode, "Dec() with delta not supported for enum types")
		}

		// Get the enum type metadata
		// Normalize to lowercase for case-insensitive lookups
		enumTypeKey := "__enum_type_" + strings.ToLower(val.TypeName)
		enumTypeVal, ok := i.env.Get(enumTypeKey)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "enum type metadata not found for %s", val.TypeName)
		}

		enumTypeWrapper, ok := enumTypeVal.(*EnumTypeValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "invalid enum type metadata for %s", val.TypeName)
		}

		enumType := enumTypeWrapper.EnumType

		// Find current value's position in OrderedNames
		currentPos := -1
		for idx, name := range enumType.OrderedNames {
			if name == val.ValueName {
				currentPos = idx
				break
			}
		}

		if currentPos == -1 {
			return i.newErrorWithLocation(i.currentNode, "enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
		}

		// Check if we can decrement (not at the beginning)
		if currentPos <= 0 {
			return i.newErrorWithLocation(i.currentNode, "Dec() cannot decrement enum below its minimum value")
		}

		// Get previous value
		prevValueName := enumType.OrderedNames[currentPos-1]
		prevOrdinal := enumType.Values[prevValueName]

		// Create new enum value
		newValue = &EnumValue{
			TypeName:     val.TypeName,
			ValueName:    prevValueName,
			OrdinalValue: prevOrdinal,
		}

	default:
		return i.newErrorWithLocation(i.currentNode, "Dec() expects Integer or Enum, got %s", val.Type())
	}

	// Assign the new value back to the lvalue
	if err := i.assignToLValue(args[0], newValue); err != nil {
		return i.newErrorWithLocation(i.currentNode, "Dec() failed to assign: %s", err.Error())
	}

	// Return the new value (allows Dec to be used in expressions)
	return newValue
}

// builtinSucc implements the Succ() built-in function.
// It returns the successor of an ordinal value (Integer or Enum).
// Succ() function for ordinal types
func (i *Interpreter) builtinSucc(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Succ() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	switch val := arg.(type) {
	case *IntegerValue:
		return &IntegerValue{Value: val.Value + 1}
	case *EnumValue:
		// Normalize to lowercase for case-insensitive lookups
		enumTypeKey := "__enum_type_" + strings.ToLower(val.TypeName)
		enumTypeVal, ok := i.env.Get(enumTypeKey)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "enum type metadata not found for %s", val.TypeName)
		}
		enumTypeWrapper, ok := enumTypeVal.(*EnumTypeValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "invalid enum type metadata for %s", val.TypeName)
		}
		enumType := enumTypeWrapper.EnumType

		// Find current position
		currentPos := -1
		for idx, name := range enumType.OrderedNames {
			if name == val.ValueName {
				currentPos = idx
				break
			}
		}
		if currentPos == -1 {
			return i.newErrorWithLocation(i.currentNode, "enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
		}
		if currentPos >= len(enumType.OrderedNames)-1 {
			return i.newErrorWithLocation(i.currentNode, "Succ() cannot get successor of maximum enum value")
		}
		nextValueName := enumType.OrderedNames[currentPos+1]
		nextOrdinal := enumType.Values[nextValueName]
		return &EnumValue{
			TypeName:     val.TypeName,
			ValueName:    nextValueName,
			OrdinalValue: nextOrdinal,
		}
	default:
		return i.newErrorWithLocation(i.currentNode, "Succ() expects Integer or Enum, got %s", arg.Type())
	}
}

// builtinPred implements the Pred() built-in function.
// It returns the predecessor of an ordinal value (Integer or Enum).
func (i *Interpreter) builtinPred(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Pred() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	switch val := arg.(type) {
	case *IntegerValue:
		return &IntegerValue{Value: val.Value - 1}
	case *EnumValue:
		// Normalize to lowercase for case-insensitive lookups
		enumTypeKey := "__enum_type_" + strings.ToLower(val.TypeName)
		enumTypeVal, ok := i.env.Get(enumTypeKey)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "enum type metadata not found for %s", val.TypeName)
		}
		enumTypeWrapper, ok := enumTypeVal.(*EnumTypeValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "invalid enum type metadata for %s", val.TypeName)
		}
		enumType := enumTypeWrapper.EnumType

		// Find current position
		currentPos := -1
		for idx, name := range enumType.OrderedNames {
			if name == val.ValueName {
				currentPos = idx
				break
			}
		}
		if currentPos == -1 {
			return i.newErrorWithLocation(i.currentNode, "enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
		}
		if currentPos <= 0 {
			return i.newErrorWithLocation(i.currentNode, "Pred() cannot get predecessor of minimum enum value")
		}
		prevValueName := enumType.OrderedNames[currentPos-1]
		prevOrdinal := enumType.Values[prevValueName]
		return &EnumValue{
			TypeName:     val.TypeName,
			ValueName:    prevValueName,
			OrdinalValue: prevOrdinal,
		}
	default:
		return i.newErrorWithLocation(i.currentNode, "Pred() expects Integer or Enum, got %s", arg.Type())
	}
}

// builtinAssert implements the Assert() built-in function.
// Usage: Assert(condition) or Assert(condition, message)
func (i *Interpreter) builtinAssert(args []Value) Value {
	// Validate argument count (1-2 arguments)
	if len(args) < 1 || len(args) > 2 {
		return i.newErrorWithLocation(i.currentNode, "Assert() expects 1-2 arguments, got %d", len(args))
	}

	// First argument must be Boolean
	condition, ok := args[0].(*BooleanValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Assert() first argument must be Boolean, got %s", args[0].Type())
	}

	// If condition is true, assertion passes - return nil
	if condition.Value {
		return &NilValue{}
	}

	// Condition is false - raise EAssertionFailed exception
	// Build the assertion message with position information
	var message string
	if i.currentNode != nil {
		pos := i.currentNode.Pos()
		message = fmt.Sprintf("Assertion failed [line: %d, column: %d]", pos.Line, pos.Column)
	} else {
		message = "Assertion failed"
	}

	// If custom message provided, append it
	if len(args) == 2 {
		customMsg, ok := args[1].(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "Assert() second argument must be String, got %s", args[1].Type())
		}
		message = message + " : " + customMsg.Value
	}

	// Create EAssertionFailed exception
	assertClass, ok := i.classes["EAssertionFailed"]
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EAssertionFailed exception class not found")
	}

	// Create exception instance
	instance := &ObjectInstance{
		Class:  assertClass,
		Fields: make(map[string]Value),
	}

	// Set the Message field
	instance.Fields["Message"] = &StringValue{Value: message}

	// Create exception value and set it
	// Position is nil for built-in function exceptions
	i.exception = &ExceptionValue{
		ClassInfo: assertClass,
		Message:   message,
		Instance:  instance,
		Position:  nil,
		CallStack: nil,
	}

	return nil
}

// assignToLValue assigns a value to an lvalue expression (Identifier, IndexExpression, MemberAccessExpression).
// This is used by Inc/Dec and other built-ins that modify lvalues in place.
func (i *Interpreter) assignToLValue(lvalue ast.Expression, value Value) error {
	switch target := lvalue.(type) {
	case *ast.Identifier:
		// Simple variable assignment
		varName := target.Value

		// Check if this is a var parameter (ReferenceValue)
		if currentVal, exists := i.env.Get(varName); exists {
			if refVal, isRef := currentVal.(*ReferenceValue); isRef {
				// Write through the reference
				return refVal.Assign(value)
			}
		}

		// Normal variable assignment
		return i.env.Set(varName, value)

	case *ast.IndexExpression:
		// Array/string index assignment: arr[i] := value
		// Evaluate the array/string
		arrVal := i.Eval(target.Left)
		if isError(arrVal) {
			return fmt.Errorf("failed to evaluate array: %s", arrVal.(*ErrorValue).Message)
		}

		// Evaluate the index
		indexVal := i.Eval(target.Index)
		if isError(indexVal) {
			return fmt.Errorf("failed to evaluate index: %s", indexVal.(*ErrorValue).Message)
		}

		indexInt, ok := indexVal.(*IntegerValue)
		if !ok {
			return fmt.Errorf("index must be Integer, got %s", indexVal.Type())
		}

		index := int(indexInt.Value)

		// Handle different array types
		arr, ok := arrVal.(*ArrayValue)
		if !ok {
			return fmt.Errorf("cannot index into %s", arrVal.Type())
		}

		// Perform bounds checking and get physical index
		if arr.ArrayType == nil {
			return fmt.Errorf("array has no type information")
		}

		var physicalIndex int
		if arr.ArrayType.IsStatic() {
			// Static array: check bounds and adjust for low bound
			lowBound := *arr.ArrayType.LowBound
			highBound := *arr.ArrayType.HighBound

			if index < lowBound || index > highBound {
				return fmt.Errorf("array index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
			}

			physicalIndex = index - lowBound
		} else {
			// Dynamic array: zero-based indexing
			if index < 0 || index >= len(arr.Elements) {
				return fmt.Errorf("array index out of bounds: %d (array length is %d)", index, len(arr.Elements))
			}

			physicalIndex = index
		}

		// Check physical bounds
		if physicalIndex < 0 || physicalIndex >= len(arr.Elements) {
			return fmt.Errorf("array index out of bounds: physical index %d, length %d", physicalIndex, len(arr.Elements))
		}

		// Update the array element
		arr.Elements[physicalIndex] = value
		return nil

	case *ast.MemberAccessExpression:
		// Object/record field assignment: obj.field := value
		// Evaluate the object/record
		objVal := i.Eval(target.Object)
		if isError(objVal) {
			return fmt.Errorf("failed to evaluate object: %s", objVal.(*ErrorValue).Message)
		}

		fieldName := target.Member.Value

		// Handle different object types
		switch obj := objVal.(type) {
		case *ObjectInstance:
			obj.Fields[fieldName] = value
			return nil

		case *RecordValue:
			obj.Fields[fieldName] = value
			return nil

		case *ReferenceValue:
			// Dereference and try again
			deref, err := obj.Dereference()
			if err != nil {
				return err
			}
			if objInst, ok := deref.(*ObjectInstance); ok {
				objInst.Fields[fieldName] = value
				return nil
			}
			if recVal, ok := deref.(*RecordValue); ok {
				recVal.Fields[fieldName] = value
				return nil
			}
			return fmt.Errorf("cannot access field of %s", deref.Type())

		default:
			return fmt.Errorf("cannot access field of %s", objVal.Type())
		}

	default:
		return fmt.Errorf("invalid lvalue type: %T", lvalue)
	}
}
