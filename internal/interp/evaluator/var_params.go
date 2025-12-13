package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file provides helpers for evaluating lvalue expressions (expressions that
// can appear on the left side of an assignment) and implements built-in functions
// that modify their arguments in place.
//
// Lvalue expressions include:
// - Identifiers: x, myVar
// - Index expressions: arr[i], matrix[i][j]
// - Member access expressions: obj.field, point.x
//
// The key pattern is EvaluateLValue, which returns:
// 1. The current value at the lvalue
// 2. A closure to assign a new value to that lvalue
// This avoids double-evaluation of side-effecting expressions.

// AssignFunc is a function that assigns a value to an lvalue.
type AssignFunc func(Value) error

// EvaluateLValue evaluates an lvalue expression once and returns:
// 1. The current value at that lvalue
// 2. A closure function to assign a new value to that lvalue
// 3. An error if evaluation failed
//
// This avoids double-evaluation of side-effecting expressions in Inc/Dec etc.
// For example, Inc(arr[f()]) only calls f() once, not twice.
func (e *Evaluator) EvaluateLValue(lvalue ast.Expression, ctx *ExecutionContext) (Value, AssignFunc, error) {
	switch target := lvalue.(type) {
	case *ast.Identifier:
		return e.evaluateLValueIdentifier(target, ctx)

	case *ast.IndexExpression:
		return e.evaluateLValueIndex(target, ctx)

	case *ast.MemberAccessExpression:
		return e.evaluateLValueMember(target, ctx)

	default:
		return nil, nil, fmt.Errorf("invalid lvalue type: %T", lvalue)
	}
}

// evaluateLValueIdentifier handles simple variable lvalues: x
func (e *Evaluator) evaluateLValueIdentifier(target *ast.Identifier, ctx *ExecutionContext) (Value, AssignFunc, error) {
	varName := target.Value

	// Get current value
	currentVal, exists := e.GetVar(ctx, varName)
	if !exists {
		// Not in environment - check implicit Self context (fields) so member assignment
		// like `Inner.Value := ...` works inside instance methods.
		if selfRaw, selfOk := ctx.Env().Get("Self"); selfOk {
			if selfVal, ok := selfRaw.(Value); ok {
				// Object instance fields
				if selfVal.Type() == "OBJECT" {
					if objVal, ok := selfVal.(ObjectValue); ok {
						fieldVal := objVal.GetField(varName)
						fieldExists := fieldVal != nil
						if !fieldExists {
							// Field may exist but be unset; consult class metadata.
							if objInst, ok := selfVal.(*runtime.ObjectInstance); ok {
								if objInst.Class != nil && objInst.Class.FieldExists(ident.Normalize(varName)) {
									fieldExists = true
									fieldVal = &runtime.NilValue{}
								}
							}
						}

						if fieldExists {
							assignFunc := func(value Value) error {
								if setter, ok := selfVal.(ObjectFieldSetter); ok {
									setter.SetField(varName, value)
									return nil
								}
								return fmt.Errorf("object does not support field assignment")
							}
							return fieldVal, assignFunc, nil
						}
					}
				}
			}
		}

		return nil, nil, fmt.Errorf("undefined variable: %s", varName)
	}

	// Create assignment function
	assignFunc := func(value Value) error {
		// Check if this is a var parameter (ReferenceValue)
		// ReferenceValue has Type() == "REFERENCE" and we can detect it via interface
		if refVal, isRef := currentVal.(ReferenceAccessor); isRef {
			// Write through the reference
			return refVal.Assign(value)
		}
		// Normal variable assignment
		if !e.SetVar(ctx, varName, value) {
			return fmt.Errorf("failed to set variable: %s", varName)
		}
		return nil
	}

	return currentVal, assignFunc, nil
}

// evaluateLValueIndex handles array index lvalues: arr[i]
func (e *Evaluator) evaluateLValueIndex(target *ast.IndexExpression, ctx *ExecutionContext) (Value, AssignFunc, error) {
	// Evaluate array and index ONCE
	arrVal := e.Eval(target.Left, ctx)
	if isError(arrVal) {
		if errVal, ok := arrVal.(*runtime.ErrorValue); ok {
			return nil, nil, fmt.Errorf("failed to evaluate array: %s", errVal.Message)
		}
		return nil, nil, fmt.Errorf("failed to evaluate array: unknown error")
	}

	indexVal := e.Eval(target.Index, ctx)
	if isError(indexVal) {
		if errVal, ok := indexVal.(*runtime.ErrorValue); ok {
			return nil, nil, fmt.Errorf("failed to evaluate index: %s", errVal.Message)
		}
		return nil, nil, fmt.Errorf("failed to evaluate index: unknown error")
	}

	indexInt, ok := indexVal.(*runtime.IntegerValue)
	if !ok {
		return nil, nil, fmt.Errorf("index must be Integer, got %s", indexVal.Type())
	}
	index := int(indexInt.Value)

	// Unwrap ReferenceValue if the array itself is a var parameter
	if ref, isRef := arrVal.(ReferenceAccessor); isRef {
		deref, err := ref.Dereference()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to dereference array: %s", err.Error())
		}
		arrVal = deref
	}

	// Get the array
	arr, ok := arrVal.(*runtime.ArrayValue)
	if !ok {
		return nil, nil, fmt.Errorf("cannot index into %s", arrVal.Type())
	}

	// Perform bounds checking and get physical index
	if arr.ArrayType == nil {
		return nil, nil, fmt.Errorf("array has no type information")
	}

	var physicalIndex int
	if arr.ArrayType.IsStatic() {
		// Static array: check bounds and adjust for low bound
		lowBound := *arr.ArrayType.LowBound
		highBound := *arr.ArrayType.HighBound

		if index < lowBound || index > highBound {
			return nil, nil, fmt.Errorf("array index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
		}

		physicalIndex = index - lowBound
	} else {
		// Dynamic array: zero-based indexing
		if index < 0 || index >= len(arr.Elements) {
			return nil, nil, fmt.Errorf("array index out of bounds: %d (array length is %d)", index, len(arr.Elements))
		}

		physicalIndex = index
	}

	// Check physical bounds
	if physicalIndex < 0 || physicalIndex >= len(arr.Elements) {
		return nil, nil, fmt.Errorf("array index out of bounds: physical index %d, length %d", physicalIndex, len(arr.Elements))
	}

	// Get current value
	currentVal := arr.Elements[physicalIndex]

	// Create assignment function (captures arr and physicalIndex)
	assignFunc := func(value Value) error {
		arr.Elements[physicalIndex] = value
		return nil
	}

	return currentVal, assignFunc, nil
}

// evaluateLValueMember handles object/record field lvalues: obj.field
func (e *Evaluator) evaluateLValueMember(target *ast.MemberAccessExpression, ctx *ExecutionContext) (Value, AssignFunc, error) {
	// Evaluate object ONCE
	objVal := e.Eval(target.Object, ctx)
	if isError(objVal) {
		if errVal, ok := objVal.(*runtime.ErrorValue); ok {
			return nil, nil, fmt.Errorf("failed to evaluate object: %s", errVal.Message)
		}
		return nil, nil, fmt.Errorf("failed to evaluate object: unknown error")
	}

	fieldName := target.Member.Value

	// Handle ReferenceValue
	if ref, isRef := objVal.(ReferenceAccessor); isRef {
		deref, err := ref.Dereference()
		if err != nil {
			return nil, nil, err
		}
		objVal = deref
	}

	// Handle ObjectValue (class instance)
	if obj, ok := objVal.(ObjectValue); ok {
		currentVal := obj.GetField(fieldName)
		if currentVal == nil {
			return nil, nil, fmt.Errorf("field '%s' not found in class '%s'", fieldName, obj.ClassName())
		}

		assignFunc := func(value Value) error {
			// ObjectValue interface doesn't have SetField, need to use type assertion
			if setter, ok := objVal.(ObjectFieldSetter); ok {
				setter.SetField(fieldName, value)
				return nil
			}
			return fmt.Errorf("object does not support field assignment")
		}

		return currentVal, assignFunc, nil
	}

	// Handle RecordInstanceValue
	if rec, ok := objVal.(RecordInstanceValue); ok {
		currentVal, exists := rec.GetRecordField(fieldName)
		if !exists {
			return nil, nil, fmt.Errorf("field '%s' not found in record '%s'", fieldName, rec.GetRecordTypeName())
		}

		assignFunc := func(value Value) error {
			// RecordInstanceValue interface doesn't have SetRecordField, need type assertion
			if setter, ok := objVal.(RecordFieldSetter); ok {
				setter.SetRecordField(fieldName, value)
				return nil
			}
			return fmt.Errorf("record does not support field assignment")
		}

		return currentVal, assignFunc, nil
	}

	return nil, nil, fmt.Errorf("cannot access field of %s", objVal.Type())
}

// ReferenceAccessor is an optional interface for reference values.
// This allows the evaluator to dereference and assign through var parameters
// without importing the interp package directly.
type ReferenceAccessor interface {
	Value
	Dereference() (Value, error)
	Assign(value Value) error
}

// ObjectFieldSetter is an optional interface for objects that support field assignment.
type ObjectFieldSetter interface {
	Value
	SetField(name string, value Value)
}

// RecordFieldSetter is an optional interface for records that support field assignment.
type RecordFieldSetter interface {
	Value
	SetRecordField(name string, value Value) bool
}

// IsVarTarget checks if a node is a valid lvalue (can appear on left side of assignment).
func IsVarTarget(node ast.Node) bool {
	switch node.(type) {
	case *ast.Identifier, *ast.IndexExpression, *ast.MemberAccessExpression:
		return true
	default:
		return false
	}
}

// ============================================================================
// Inc/Dec Built-in Functions
// ============================================================================

// builtinInc implements the Inc() built-in function.
// It increments a variable in place: Inc(x) or Inc(x, delta)
// Supports any lvalue: Inc(x), Inc(arr[i]), Inc(obj.field)
func (e *Evaluator) builtinInc(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (1-2 arguments)
	if len(args) < 1 || len(args) > 2 {
		return e.newError(nil, "Inc() expects 1-2 arguments, got %d", len(args))
	}

	// Get delta (default 1)
	delta := int64(1)
	if len(args) == 2 {
		deltaVal := e.Eval(args[1], ctx)
		if isError(deltaVal) {
			return deltaVal
		}
		deltaInt, ok := deltaVal.(*runtime.IntegerValue)
		if !ok {
			return e.newError(nil, "Inc() delta must be Integer, got %s", deltaVal.Type())
		}
		delta = deltaInt.Value
	}

	// Evaluate lvalue once and get both current value and assignment target
	lvalue := args[0]
	currentVal, assignFunc, err := e.EvaluateLValue(lvalue, ctx)
	if err != nil {
		return e.newError(nil, "Inc() failed to evaluate lvalue: %s", err.Error())
	}

	// Unwrap ReferenceValue if needed
	if ref, isRef := currentVal.(ReferenceAccessor); isRef {
		actualVal, err := ref.Dereference()
		if err != nil {
			return e.newError(nil, "%s", err.Error())
		}
		currentVal = actualVal
	}

	// Handle nil values (uninitialized array/record elements default to 0)
	if currentVal == nil {
		currentVal = &runtime.IntegerValue{Value: 0}
	}

	// Compute new value based on type
	var newValue Value

	switch val := currentVal.(type) {
	case *runtime.IntegerValue:
		// Increment integer by delta
		newValue = &runtime.IntegerValue{Value: val.Value + delta}

	case *runtime.EnumValue:
		// For enums, delta must be 1 (get successor)
		if delta != 1 {
			return e.newError(nil, "Inc() with delta not supported for enum types")
		}

		// Get the enum type metadata
		enumType, err := e.lookupEnumType(val.TypeName, ctx)
		if err != nil {
			return e.newError(nil, "enum type metadata not found for %s", val.TypeName)
		}

		// Find current value's position in OrderedNames
		currentPos := -1
		for idx, name := range enumType.OrderedNames {
			if name == val.ValueName {
				currentPos = idx
				break
			}
		}

		if currentPos == -1 {
			return e.newError(nil, "enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
		}

		// Check if we can increment (not at the end)
		if currentPos >= len(enumType.OrderedNames)-1 {
			return e.newError(nil, "Inc() cannot increment enum beyond its maximum value")
		}

		// Get next value
		nextValueName := enumType.OrderedNames[currentPos+1]
		nextOrdinal := enumType.Values[nextValueName]

		// Create new enum value
		newValue = &runtime.EnumValue{
			TypeName:     val.TypeName,
			ValueName:    nextValueName,
			OrdinalValue: nextOrdinal,
		}

	default:
		return e.newError(nil, "Inc() expects Integer or Enum, got %s", currentVal.Type())
	}

	// Assign the new value back using the pre-evaluated assignment function
	if err := assignFunc(newValue); err != nil {
		return e.newError(nil, "Inc() failed to assign: %s", err.Error())
	}

	// Return the new value (allows Inc to be used in expressions)
	return newValue
}

// builtinDec implements the Dec() built-in function.
// It decrements a variable in place: Dec(x) or Dec(x, delta)
// Supports any lvalue: Dec(x), Dec(arr[i]), Dec(obj.field)
func (e *Evaluator) builtinDec(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (1-2 arguments)
	if len(args) < 1 || len(args) > 2 {
		return e.newError(nil, "Dec() expects 1-2 arguments, got %d", len(args))
	}

	// Get delta (default 1)
	delta := int64(1)
	if len(args) == 2 {
		deltaVal := e.Eval(args[1], ctx)
		if isError(deltaVal) {
			return deltaVal
		}
		deltaInt, ok := deltaVal.(*runtime.IntegerValue)
		if !ok {
			return e.newError(nil, "Dec() delta must be Integer, got %s", deltaVal.Type())
		}
		delta = deltaInt.Value
	}

	// Evaluate lvalue once and get both current value and assignment target
	lvalue := args[0]
	currentVal, assignFunc, err := e.EvaluateLValue(lvalue, ctx)
	if err != nil {
		return e.newError(nil, "Dec() failed to evaluate lvalue: %s", err.Error())
	}

	// Unwrap ReferenceValue if needed
	if ref, isRef := currentVal.(ReferenceAccessor); isRef {
		actualVal, err := ref.Dereference()
		if err != nil {
			return e.newError(nil, "%s", err.Error())
		}
		currentVal = actualVal
	}

	// Handle nil values (uninitialized array/record elements default to 0)
	if currentVal == nil {
		currentVal = &runtime.IntegerValue{Value: 0}
	}

	// Compute new value based on type
	var newValue Value

	switch val := currentVal.(type) {
	case *runtime.IntegerValue:
		// Decrement integer by delta
		newValue = &runtime.IntegerValue{Value: val.Value - delta}

	case *runtime.EnumValue:
		// For enums, delta must be 1 (get predecessor)
		if delta != 1 {
			return e.newError(nil, "Dec() with delta not supported for enum types")
		}

		// Get the enum type metadata
		enumType, err := e.lookupEnumType(val.TypeName, ctx)
		if err != nil {
			return e.newError(nil, "enum type metadata not found for %s", val.TypeName)
		}

		// Find current value's position in OrderedNames
		currentPos := -1
		for idx, name := range enumType.OrderedNames {
			if name == val.ValueName {
				currentPos = idx
				break
			}
		}

		if currentPos == -1 {
			return e.newError(nil, "enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
		}

		// Check if we can decrement (not at the beginning)
		if currentPos <= 0 {
			return e.newError(nil, "Dec() cannot decrement enum below its minimum value")
		}

		// Get previous value
		prevValueName := enumType.OrderedNames[currentPos-1]
		prevOrdinal := enumType.Values[prevValueName]

		// Create new enum value
		newValue = &runtime.EnumValue{
			TypeName:     val.TypeName,
			ValueName:    prevValueName,
			OrdinalValue: prevOrdinal,
		}

	default:
		return e.newError(nil, "Dec() expects Integer or Enum, got %s", currentVal.Type())
	}

	// Assign the new value back using the pre-evaluated assignment function
	if err := assignFunc(newValue); err != nil {
		return e.newError(nil, "Dec() failed to assign: %s", err.Error())
	}

	// Return the new value (allows Dec to be used in expressions)
	return newValue
}

// ============================================================================
// SetLength Built-in Function
// ============================================================================

// builtinSetLength implements the SetLength() built-in function.
// SetLength(arr, newSize) - resizes a dynamic array
// SetLength(str, newLength) - resizes a string
func (e *Evaluator) builtinSetLength(args []ast.Expression, ctx *ExecutionContext) Value {
	if len(args) != 2 {
		return e.newError(nil, "SetLength() expects exactly 2 arguments, got %d", len(args))
	}

	// Use EvaluateLValue to support identifiers, indexed arrays, member access, etc.
	currentVal, assignFunc, err := e.EvaluateLValue(args[0], ctx)
	if err != nil {
		return e.newError(nil, "SetLength() first argument must be a variable: %s", err.Error())
	}

	// Dereference if it's a var parameter (ReferenceValue)
	if ref, isRef := currentVal.(ReferenceAccessor); isRef {
		actualVal, err := ref.Dereference()
		if err != nil {
			return e.newError(nil, "%s", err.Error())
		}
		currentVal = actualVal
	}

	// Evaluate the second argument (new length)
	lengthVal := e.Eval(args[1], ctx)
	if isError(lengthVal) {
		return lengthVal
	}

	lengthInt, ok := lengthVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(nil, "SetLength() expects integer as second argument, got %s", lengthVal.Type())
	}

	newLength := int(lengthInt.Value)
	// DWScript/Delphi behavior: negative lengths are treated as 0
	if newLength < 0 {
		newLength = 0
	}

	// Handle arrays
	if arrayVal, ok := currentVal.(*runtime.ArrayValue); ok {
		// Check that it's a dynamic array
		if arrayVal.ArrayType == nil {
			return e.newError(nil, "array has no type information")
		}

		if arrayVal.ArrayType.IsStatic() {
			return e.newError(nil, "SetLength() can only be used with dynamic arrays, not static arrays")
		}

		currentLength := len(arrayVal.Elements)

		if newLength != currentLength {
			if newLength < currentLength {
				// Truncate the slice
				arrayVal.Elements = arrayVal.Elements[:newLength]
			} else {
				// Extend the slice with nil values
				additional := make([]runtime.Value, newLength-currentLength)
				arrayVal.Elements = append(arrayVal.Elements, additional...)
			}
		}

		return &runtime.NilValue{}
	}

	// Handle strings
	if strVal, ok := currentVal.(*runtime.StringValue); ok {
		// Use rune-based SetLength to handle UTF-8 correctly
		newStr := runeSetLength(strVal.Value, newLength)

		// Create new StringValue
		newValue := &runtime.StringValue{Value: newStr}

		// Use the assignment function to update the string
		if err := assignFunc(newValue); err != nil {
			return e.newError(nil, "failed to update string variable: %s", err)
		}

		return &runtime.NilValue{}
	}

	return e.newError(nil, "SetLength() expects array or string as first argument, got %s", currentVal.Type())
}

// ============================================================================
// Insert/Delete Built-in Functions
// ============================================================================

// builtinInsert implements the Insert() built-in function.
// Insert(source, target, pos) - inserts source string into target at position pos
// target is modified in place (must be a variable)
func (e *Evaluator) builtinInsert(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (3 arguments)
	if len(args) != 3 {
		return e.newError(nil, "Insert() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: source string to insert (evaluate it)
	sourceVal := e.Eval(args[0], ctx)
	if isError(sourceVal) {
		return sourceVal
	}
	sourceStr, ok := sourceVal.(*runtime.StringValue)
	if !ok {
		return e.newError(nil, "Insert() expects String as first argument (source), got %s", sourceVal.Type())
	}

	// Second argument: target string variable (must be an lvalue)
	currentVal, assignFunc, err := e.EvaluateLValue(args[1], ctx)
	if err != nil {
		return e.newError(nil, "Insert() second argument (target) must be a variable: %s", err.Error())
	}

	// Dereference if it's a var parameter
	if ref, isRef := currentVal.(ReferenceAccessor); isRef {
		actualVal, err := ref.Dereference()
		if err != nil {
			return e.newError(nil, "%s", err.Error())
		}
		currentVal = actualVal
	}

	targetStr, ok := currentVal.(*runtime.StringValue)
	if !ok {
		return e.newError(nil, "Insert() expects target to be String, got %s", currentVal.Type())
	}

	// Third argument: position (1-based index)
	posVal := e.Eval(args[2], ctx)
	if isError(posVal) {
		return posVal
	}
	posInt, ok := posVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(nil, "Insert() expects Integer as third argument (position), got %s", posVal.Type())
	}

	pos := int(posInt.Value)
	target := targetStr.Value
	source := sourceStr.Value

	// Use rune-based insertion to handle UTF-8 correctly
	newStr := runeInsert(source, target, pos)

	// Update the target variable with the new string
	newValue := &runtime.StringValue{Value: newStr}
	if err := assignFunc(newValue); err != nil {
		return e.newError(nil, "failed to update target variable: %s", err)
	}

	return &runtime.NilValue{}
}

// builtinDeleteString implements the Delete() built-in function for strings.
// Delete(s, pos, count) - deletes count characters from string s starting at position pos
// s is modified in place (must be a variable)
func (e *Evaluator) builtinDeleteString(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (3 arguments)
	if len(args) != 3 {
		return e.newError(nil, "Delete() for strings expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: string variable (must be an lvalue)
	currentVal, assignFunc, err := e.EvaluateLValue(args[0], ctx)
	if err != nil {
		return e.newError(nil, "Delete() first argument must be a variable: %s", err.Error())
	}

	// Dereference if it's a var parameter
	if ref, isRef := currentVal.(ReferenceAccessor); isRef {
		actualVal, err := ref.Dereference()
		if err != nil {
			return e.newError(nil, "%s", err.Error())
		}
		currentVal = actualVal
	}

	strVal, ok := currentVal.(*runtime.StringValue)
	if !ok {
		return e.newError(nil, "Delete() expects first argument to be String, got %s", currentVal.Type())
	}

	// Second argument: position (1-based index)
	posVal := e.Eval(args[1], ctx)
	if isError(posVal) {
		return posVal
	}
	posInt, ok := posVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(nil, "Delete() expects Integer as second argument (position), got %s", posVal.Type())
	}

	// Third argument: count (number of characters to delete)
	countVal := e.Eval(args[2], ctx)
	if isError(countVal) {
		return countVal
	}
	countInt, ok := countVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(nil, "Delete() expects Integer as third argument (count), got %s", countVal.Type())
	}

	pos := int(posInt.Value)
	count := int(countInt.Value)
	str := strVal.Value

	// Use rune-based deletion to handle UTF-8 correctly
	newStr := runeDelete(str, pos, count)

	// Update the string variable with the new value
	newValue := &runtime.StringValue{Value: newStr}
	if err := assignFunc(newValue); err != nil {
		return e.newError(nil, "failed to update variable: %s", err)
	}

	return &runtime.NilValue{}
}

// ============================================================================
// String Helper Functions (rune-based for UTF-8 correctness)
// ============================================================================

// runeDelete deletes count characters from string s starting at position pos (1-based).
func runeDelete(s string, pos, count int) string {
	if pos < 1 || count <= 0 {
		return s
	}

	runes := []rune(s)
	length := len(runes)

	startPos := pos - 1 // Convert to 0-based
	if startPos >= length {
		return s
	}

	endPos := startPos + count
	if endPos > length {
		endPos = length
	}

	// Concatenate the part before deletion and the part after
	return string(runes[:startPos]) + string(runes[endPos:])
}

// runeInsert inserts source string into target at position pos (1-based).
func runeInsert(source, target string, pos int) string {
	if pos < 1 {
		pos = 1
	}

	runes := []rune(target)
	length := len(runes)

	insertPos := pos - 1 // Convert to 0-based
	if insertPos > length {
		insertPos = length
	}

	return string(runes[:insertPos]) + source + string(runes[insertPos:])
}

// runeSetLength sets the length of string s to newLength characters.
// If truncating, removes characters from the end.
// If extending, pads with spaces to match DWScript semantics.
func runeSetLength(s string, newLength int) string {
	if newLength < 0 {
		newLength = 0
	}

	runes := []rune(s)
	currentLength := len(runes)

	if newLength == currentLength {
		return s
	}

	if newLength < currentLength {
		// Truncate
		return string(runes[:newLength])
	}

	// Extend with spaces to match DWScript semantics
	padding := newLength - currentLength
	spaces := make([]rune, padding)
	for i := range spaces {
		spaces[i] = ' '
	}
	return s + string(spaces)
}

// ============================================================================
// Swap/DivMod Built-in Functions
// ============================================================================

// builtinSwap implements the Swap() built-in function.
// Swap(a, b) - exchanges the values of two variables
// Both arguments must be lvalues (variables, array elements, object fields).
func (e *Evaluator) builtinSwap(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (exactly 2 arguments)
	if len(args) != 2 {
		return e.newError(nil, "Swap() expects exactly 2 arguments, got %d", len(args))
	}

	// Evaluate both lvalues
	val1, assign1, err1 := e.EvaluateLValue(args[0], ctx)
	if err1 != nil {
		return e.newError(nil, "Swap() first argument must be a variable: %s", err1.Error())
	}

	val2, assign2, err2 := e.EvaluateLValue(args[1], ctx)
	if err2 != nil {
		return e.newError(nil, "Swap() second argument must be a variable: %s", err2.Error())
	}

	// Dereference if either is a var parameter (ReferenceValue)
	actualVal1 := val1
	if ref, isRef := val1.(ReferenceAccessor); isRef {
		deref, err := ref.Dereference()
		if err != nil {
			return e.newError(nil, "%s", err.Error())
		}
		actualVal1 = deref
	}

	actualVal2 := val2
	if ref, isRef := val2.(ReferenceAccessor); isRef {
		deref, err := ref.Dereference()
		if err != nil {
			return e.newError(nil, "%s", err.Error())
		}
		actualVal2 = deref
	}

	// Swap the values using the assignment closures
	if err := assign1(actualVal2); err != nil {
		return e.newError(nil, "Swap() failed to update first variable: %s", err.Error())
	}

	if err := assign2(actualVal1); err != nil {
		return e.newError(nil, "Swap() failed to update second variable: %s", err.Error())
	}

	return &runtime.NilValue{}
}

// builtinDivMod implements the DivMod() built-in function.
// DivMod(dividend, divisor, quotient, remainder) - performs integer division
// First two arguments are values, last two are var parameters for output.
func (e *Evaluator) builtinDivMod(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (exactly 4 arguments)
	if len(args) != 4 {
		return e.newError(nil, "DivMod() expects exactly 4 arguments, got %d", len(args))
	}

	// Evaluate first two arguments (dividend and divisor)
	dividendVal := e.Eval(args[0], ctx)
	if isError(dividendVal) {
		return dividendVal
	}
	dividendInt, ok := dividendVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(nil, "DivMod() expects integer as first argument, got %s", dividendVal.Type())
	}

	divisorVal := e.Eval(args[1], ctx)
	if isError(divisorVal) {
		return divisorVal
	}
	divisorInt, ok := divisorVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(nil, "DivMod() expects integer as second argument, got %s", divisorVal.Type())
	}

	// Check for division by zero
	if divisorInt.Value == 0 {
		return e.newError(nil, "DivMod() division by zero")
	}

	// Calculate quotient and remainder
	quotient := dividendInt.Value / divisorInt.Value
	remainder := dividendInt.Value % divisorInt.Value

	// Evaluate the output lvalues (quotient and remainder variables)
	_, assignQuotient, err := e.EvaluateLValue(args[2], ctx)
	if err != nil {
		return e.newError(nil, "DivMod() third argument must be a variable: %s", err.Error())
	}

	_, assignRemainder, err := e.EvaluateLValue(args[3], ctx)
	if err != nil {
		return e.newError(nil, "DivMod() fourth argument must be a variable: %s", err.Error())
	}

	// Assign the results
	quotientResult := &runtime.IntegerValue{Value: quotient}
	remainderResult := &runtime.IntegerValue{Value: remainder}

	if err := assignQuotient(quotientResult); err != nil {
		return e.newError(nil, "DivMod() failed to update quotient variable: %s", err.Error())
	}

	if err := assignRemainder(remainderResult); err != nil {
		return e.newError(nil, "DivMod() failed to update remainder variable: %s", err.Error())
	}

	return &runtime.NilValue{}
}

// ============================================================================
// TryStrToInt/TryStrToFloat Built-in Functions
// ============================================================================

// builtinTryStrToInt implements the TryStrToInt() built-in function.
// TryStrToInt(str: String, var outValue: Integer): Boolean
// TryStrToInt(str: String, base: Integer, var outValue: Integer): Boolean
// Returns true and updates outValue on successful parsing, false otherwise.
func (e *Evaluator) builtinTryStrToInt(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (2 or 3 arguments)
	if len(args) < 2 || len(args) > 3 {
		return e.newError(nil, "TryStrToInt() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument: string to convert
	strArg := e.Eval(args[0], ctx)
	if isError(strArg) {
		return strArg
	}
	strVal, ok := strArg.(*runtime.StringValue)
	if !ok {
		return e.newError(nil, "TryStrToInt() expects String as first argument, got %s", strArg.Type())
	}

	// Determine if we have 2 or 3 arguments
	var base int
	var valueArg ast.Expression

	if len(args) == 2 {
		// TryStrToInt(str, var value) - base defaults to 10
		base = 10
		valueArg = args[1]
	} else {
		// TryStrToInt(str, base, var value)
		baseArg := e.Eval(args[1], ctx)
		if isError(baseArg) {
			return baseArg
		}
		baseInt, ok := baseArg.(*runtime.IntegerValue)
		if !ok {
			return e.newError(nil, "TryStrToInt() expects Integer as second argument (base), got %s", baseArg.Type())
		}
		base = int(baseInt.Value)

		// Validate base range (2-36)
		if base < 2 || base > 36 {
			// Invalid base - return false without modifying variable
			return &runtime.BooleanValue{Value: false}
		}

		valueArg = args[2]
	}

	// Use EvaluateLValue to get the var parameter (supports any lvalue)
	_, assignFunc, err := e.EvaluateLValue(valueArg, ctx)
	if err != nil {
		return e.newError(nil, "TryStrToInt() var parameter must be a variable: %s", err.Error())
	}

	// Try to parse the string
	s := strings.TrimSpace(strVal.Value)
	if s == "" {
		// Empty string - return false without modifying variable
		return &runtime.BooleanValue{Value: false}
	}

	intValue, parseErr := strconv.ParseInt(s, base, 64)
	if parseErr != nil {
		// Parsing failed - return false without modifying variable
		return &runtime.BooleanValue{Value: false}
	}

	// Parsing succeeded - update the variable and return true
	result := &runtime.IntegerValue{Value: intValue}
	if err := assignFunc(result); err != nil {
		return e.newError(nil, "TryStrToInt() failed to update variable: %s", err.Error())
	}

	return &runtime.BooleanValue{Value: true}
}

// builtinTryStrToFloat implements the TryStrToFloat() built-in function.
// TryStrToFloat(str: String, var outValue: Float): Boolean
// Returns true and updates outValue on successful parsing, false otherwise.
func (e *Evaluator) builtinTryStrToFloat(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (exactly 2 arguments)
	if len(args) != 2 {
		return e.newError(nil, "TryStrToFloat() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to convert
	strArg := e.Eval(args[0], ctx)
	if isError(strArg) {
		return strArg
	}
	strVal, ok := strArg.(*runtime.StringValue)
	if !ok {
		return e.newError(nil, "TryStrToFloat() expects String as first argument, got %s", strArg.Type())
	}

	// Use EvaluateLValue to get the var parameter (supports any lvalue)
	_, assignFunc, err := e.EvaluateLValue(args[1], ctx)
	if err != nil {
		return e.newError(nil, "TryStrToFloat() var parameter must be a variable: %s", err.Error())
	}

	// Try to parse the string as a float
	s := strings.TrimSpace(strVal.Value)
	if s == "" {
		// Empty string - return false without modifying variable
		return &runtime.BooleanValue{Value: false}
	}

	floatValue, parseErr := strconv.ParseFloat(s, 64)
	if parseErr != nil {
		// Parsing failed - return false without modifying variable
		return &runtime.BooleanValue{Value: false}
	}

	// Parsing succeeded - update the variable and return true
	result := &runtime.FloatValue{Value: floatValue}
	if err := assignFunc(result); err != nil {
		return e.newError(nil, "TryStrToFloat() failed to update variable: %s", err.Error())
	}

	return &runtime.BooleanValue{Value: true}
}

// ============================================================================
// DecodeDate/DecodeTime Built-in Functions
// ============================================================================

// builtinDecodeDate implements the DecodeDate() built-in function.
// DecodeDate(dt: TDateTime; var year, month, day: Integer)
// Extracts year, month, day components from a TDateTime and assigns to var parameters.
func (e *Evaluator) builtinDecodeDate(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (4 arguments: dt + 3 var params)
	if len(args) != 4 {
		return e.newError(nil, "DecodeDate() expects 4 arguments (dt, var year, var month, var day), got %d", len(args))
	}

	// Evaluate the first argument (the TDateTime value)
	dtVal := e.Eval(args[0], ctx)
	if isError(dtVal) {
		return dtVal
	}

	floatVal, ok := dtVal.(*runtime.FloatValue)
	if !ok {
		return e.newError(nil, "DecodeDate() expects Float/TDateTime as first argument, got %s", dtVal.Type())
	}

	// Extract date components using the datetime utility function
	year, month, day := extractDateComponents(floatVal.Value)

	// Set the var parameters (args 1, 2, 3) using EvaluateLValue
	components := []int{year, month, day}
	paramNames := []string{"year", "month", "day"}

	for idx, val := range components {
		_, assignFunc, err := e.EvaluateLValue(args[idx+1], ctx)
		if err != nil {
			return e.newError(nil, "DecodeDate() %s parameter must be a variable: %s", paramNames[idx], err.Error())
		}

		result := &runtime.IntegerValue{Value: int64(val)}
		if err := assignFunc(result); err != nil {
			return e.newError(nil, "DecodeDate() failed to update %s variable: %s", paramNames[idx], err.Error())
		}
	}

	return &runtime.NilValue{}
}

// builtinDecodeTime implements the DecodeTime() built-in function.
// DecodeTime(dt: TDateTime; var hour, minute, second, msec: Integer)
// Extracts hour, minute, second, millisecond components from a TDateTime and assigns to var parameters.
func (e *Evaluator) builtinDecodeTime(args []ast.Expression, ctx *ExecutionContext) Value {
	// Validate argument count (5 arguments: dt + 4 var params)
	if len(args) != 5 {
		return e.newError(nil, "DecodeTime() expects 5 arguments (dt, var hour, var minute, var second, var msec), got %d", len(args))
	}

	// Evaluate the first argument (the TDateTime value)
	dtVal := e.Eval(args[0], ctx)
	if isError(dtVal) {
		return dtVal
	}

	floatVal, ok := dtVal.(*runtime.FloatValue)
	if !ok {
		return e.newError(nil, "DecodeTime() expects Float/TDateTime as first argument, got %s", dtVal.Type())
	}

	// Extract time components using the datetime utility function
	hour, minute, second, msec := extractTimeComponents(floatVal.Value)

	// Set the var parameters (args 1, 2, 3, 4) using EvaluateLValue
	components := []int{hour, minute, second, msec}
	paramNames := []string{"hour", "minute", "second", "msec"}

	for idx, val := range components {
		_, assignFunc, err := e.EvaluateLValue(args[idx+1], ctx)
		if err != nil {
			return e.newError(nil, "DecodeTime() %s parameter must be a variable: %s", paramNames[idx], err.Error())
		}

		result := &runtime.IntegerValue{Value: int64(val)}
		if err := assignFunc(result); err != nil {
			return e.newError(nil, "DecodeTime() failed to update %s variable: %s", paramNames[idx], err.Error())
		}
	}

	return &runtime.NilValue{}
}

// Note: extractDateComponents and extractTimeComponents are defined in internal/interp/datetime_utils.go
// Note: lookupEnumType is defined in set_helpers.go and reused here.
// Note: isError is defined in visitor_expressions.go and reused here.
