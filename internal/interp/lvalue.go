package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// evaluateLValue evaluates an lvalue expression once and returns:
// 1. The current value at that lvalue
// 2. A closure function to assign a new value to that lvalue
// 3. An error if evaluation failed
//
// This avoids double-evaluation of side-effecting expressions in Inc/Dec.
// For example, Inc(arr[f()]) only calls f() once, not twice.
func (i *Interpreter) evaluateLValue(lvalue ast.Expression) (Value, func(Value) error, error) {
	switch target := lvalue.(type) {
	case *ast.Identifier:
		return i.evaluateIdentifierLValue(target)
	case *ast.IndexExpression:
		return i.evaluateIndexLValue(target)
	case *ast.MemberAccessExpression:
		return i.evaluateMemberLValue(target)
	default:
		return nil, nil, fmt.Errorf("invalid lvalue type: %T", lvalue)
	}
}

// evaluateIdentifierLValue handles simple variable lvalues (e.g., x).
func (i *Interpreter) evaluateIdentifierLValue(target *ast.Identifier) (Value, func(Value) error, error) {
	varName := target.Value

	currentVal, exists := i.env.Get(varName)
	if !exists {
		return nil, nil, fmt.Errorf("undefined variable: %s", varName)
	}

	assignFunc := func(value Value) error {
		// Check if this is a var parameter (ReferenceValue)
		if currentVal, exists := i.env.Get(varName); exists {
			if refVal, isRef := currentVal.(*ReferenceValue); isRef {
				return refVal.Assign(value)
			}
		}
		return i.env.Set(varName, value)
	}

	return currentVal, assignFunc, nil
}

// evaluateIndexLValue handles array index lvalues (e.g., arr[i]).
func (i *Interpreter) evaluateIndexLValue(target *ast.IndexExpression) (Value, func(Value) error, error) {
	arrVal := i.Eval(target.Left)
	if errVal, ok := arrVal.(*ErrorValue); ok {
		return nil, nil, fmt.Errorf("failed to evaluate array: %s", errVal.Message)
	}

	indexVal := i.Eval(target.Index)
	if errVal, ok := indexVal.(*ErrorValue); ok {
		return nil, nil, fmt.Errorf("failed to evaluate index: %s", errVal.Message)
	}

	indexInt, ok := indexVal.(*IntegerValue)
	if !ok {
		return nil, nil, fmt.Errorf("index must be Integer, got %s", indexVal.Type())
	}
	index := int(indexInt.Value)

	arr, err := i.unwrapToArray(arrVal)
	if err != nil {
		return nil, nil, err
	}

	physicalIndex, err := i.computePhysicalIndex(arr, index)
	if err != nil {
		return nil, nil, err
	}

	currentVal := arr.Elements[physicalIndex]
	assignFunc := func(value Value) error {
		arr.Elements[physicalIndex] = value
		return nil
	}

	return currentVal, assignFunc, nil
}

// unwrapToArray dereferences a ReferenceValue if needed and returns the underlying ArrayValue.
func (i *Interpreter) unwrapToArray(val Value) (*ArrayValue, error) {
	if ref, isRef := val.(*ReferenceValue); isRef {
		deref, err := ref.Dereference()
		if err != nil {
			return nil, fmt.Errorf("failed to dereference array: %s", err.Error())
		}
		val = deref
	}

	arr, ok := val.(*ArrayValue)
	if !ok {
		return nil, fmt.Errorf("cannot index into %s", val.Type())
	}
	return arr, nil
}

// computePhysicalIndex validates and computes the physical array index.
func (i *Interpreter) computePhysicalIndex(arr *ArrayValue, index int) (int, error) {
	if arr.ArrayType == nil {
		return 0, fmt.Errorf("array has no type information")
	}

	var physicalIndex int
	if arr.ArrayType.IsStatic() {
		lowBound := *arr.ArrayType.LowBound
		highBound := *arr.ArrayType.HighBound

		if index < lowBound || index > highBound {
			return 0, fmt.Errorf("array index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
		}
		physicalIndex = index - lowBound
	} else {
		if index < 0 || index >= len(arr.Elements) {
			return 0, fmt.Errorf("array index out of bounds: %d (array length is %d)", index, len(arr.Elements))
		}
		physicalIndex = index
	}

	if physicalIndex < 0 || physicalIndex >= len(arr.Elements) {
		return 0, fmt.Errorf("array index out of bounds: physical index %d, length %d", physicalIndex, len(arr.Elements))
	}

	return physicalIndex, nil
}

// evaluateMemberLValue handles object/record field lvalues (e.g., obj.field).
func (i *Interpreter) evaluateMemberLValue(target *ast.MemberAccessExpression) (Value, func(Value) error, error) {
	objVal := i.Eval(target.Object)
	if errVal, ok := objVal.(*ErrorValue); ok {
		return nil, nil, fmt.Errorf("failed to evaluate object: %s", errVal.Message)
	}

	fieldName := target.Member.Value

	if ref, isRef := objVal.(*ReferenceValue); isRef {
		deref, err := ref.Dereference()
		if err != nil {
			return nil, nil, err
		}
		objVal = deref
	}

	switch obj := objVal.(type) {
	case *ObjectInstance:
		return i.evaluateObjectFieldLValue(obj, fieldName)
	case *RecordValue:
		return i.evaluateRecordFieldLValue(obj, fieldName)
	default:
		return nil, nil, fmt.Errorf("cannot access field of %s", objVal.Type())
	}
}

// evaluateObjectFieldLValue handles class instance field access.
func (i *Interpreter) evaluateObjectFieldLValue(obj *ObjectInstance, fieldName string) (Value, func(Value) error, error) {
	currentVal := obj.GetField(fieldName)
	if currentVal == nil {
		return nil, nil, fmt.Errorf("field '%s' not found in class '%s'", fieldName, obj.Class.GetName())
	}

	assignFunc := func(value Value) error {
		obj.SetField(fieldName, value)
		return nil
	}

	return currentVal, assignFunc, nil
}

// evaluateRecordFieldLValue handles record field access.
func (i *Interpreter) evaluateRecordFieldLValue(obj *RecordValue, fieldName string) (Value, func(Value) error, error) {
	currentVal, exists := obj.Fields[fieldName]
	if !exists {
		return nil, nil, fmt.Errorf("field '%s' not found in record '%s'", fieldName, obj.RecordType.Name)
	}

	assignFunc := func(value Value) error {
		obj.Fields[fieldName] = value
		return nil
	}

	return currentVal, assignFunc, nil
}
