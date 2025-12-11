package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Index Assignment Operations
// ============================================================================
//
// Handles array and string index assignment: arr[i] := value, str[i] := char
//
// Simple cases (arrays and strings) are handled directly in the evaluator.
// Complex cases (indexed properties, interfaces) are delegated to the adapter.
// ============================================================================

// evalIndexAssignmentDirect handles array/string index assignment directly.
// Complex cases (indexed properties, interfaces, default properties) are delegated to adapter.
//
// Handles:
// - Array element assignment: arr[i] := value (with bounds checking)
// - String character mutation: str[i] := 'c' (1-based, rune-aware)
//
// Delegates to adapter:
// - Indexed property writes: obj.Property[i] := value
// - Interface indexed properties
// - Object default properties: obj[i] := value (where obj has default indexed property)
// - Multi-index property access: obj.Prop[x, y] := value
func (e *Evaluator) evalIndexAssignmentDirect(
	target *ast.IndexExpression,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	// Check if this might be a multi-index property write
	// We only flatten indices if the base is a MemberAccessExpression (property access)
	base, indices := CollectIndices(target)

	// If base is a MemberAccessExpression, it's an indexed property: obj.Prop[i] := value
	// Handle directly using general-purpose method dispatch (no adapter fallback)
	if memberAccess, ok := base.(*ast.MemberAccessExpression); ok {
		return e.evalIndexedPropertyAssignment(memberAccess, indices, value, stmt, ctx)
	}

	// Evaluate the array/string being indexed
	// Process ONLY the outermost index, not all nested indices
	// This allows arr[i][j] := value to work as: (arr[i])[j] := value
	arrayVal := e.Eval(target.Left, ctx)
	if isError(arrayVal) {
		return arrayVal
	}

	// Check for exception during evaluation
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Evaluate the index
	indexVal := e.Eval(target.Index, ctx)
	if isError(indexVal) {
		return indexVal
	}

	// Check for exception during index evaluation
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Check for interface-based indexed properties or object with default indexed property
	// Both INTERFACE and OBJECT types may have default indexed properties
	if strings.HasPrefix(arrayVal.Type(), "INTERFACE") || strings.HasPrefix(arrayVal.Type(), "OBJECT[") {
		// Handle default property assignment using PropertyAccessor interface
		// Pattern: Same as 3.2.11g but lookup default property instead of named property
		return e.evalDefaultPropertyAssignment(arrayVal, indexVal, value, stmt, ctx)
	}

	// Extract integer index
	index, ok := ExtractIntegerIndex(indexVal)
	if !ok {
		return e.newError(stmt, "array index must be an ordinal, got %s", indexVal.Type())
	}

	// Handle array assignment
	if arrayValue, ok := arrayVal.(*runtime.ArrayValue); ok {
		return e.evalArrayElementAssignment(arrayValue, index, value, stmt)
	}

	// Handle string character assignment
	if strVal, ok := arrayVal.(*runtime.StringValue); ok {
		return e.evalStringCharAssignment(strVal, index, value, stmt)
	}

	return e.newError(stmt, "cannot index type %s", arrayVal.Type())
}

// evalArrayElementAssignment handles array element assignment with bounds checking.
// Supports both static arrays (with low/high bounds) and dynamic arrays (0-based).
func (e *Evaluator) evalArrayElementAssignment(
	arrayValue *runtime.ArrayValue,
	index int,
	value Value,
	stmt *ast.AssignmentStatement,
) Value {
	if arrayValue.ArrayType == nil {
		return e.newError(stmt, "array has no type information")
	}

	arrayType := arrayValue.ArrayType

	var physicalIndex int
	if arrayType.IsStatic() {
		// Static array: check bounds and adjust for low bound
		lowBound := *arrayType.LowBound
		highBound := *arrayType.HighBound

		if index < lowBound || index > highBound {
			return e.newError(stmt, "array index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
		}

		physicalIndex = index - lowBound
	} else {
		// Dynamic array: zero-based indexing
		if index < 0 || index >= len(arrayValue.Elements) {
			return e.newError(stmt, "array index out of bounds: %d (array length is %d)", index, len(arrayValue.Elements))
		}

		physicalIndex = index
	}

	// Check physical bounds (safety check)
	if physicalIndex < 0 || physicalIndex >= len(arrayValue.Elements) {
		return e.newError(stmt, "array index out of bounds: physical index %d, length %d", physicalIndex, len(arrayValue.Elements))
	}

	// Update the array element
	arrayValue.Elements[physicalIndex] = value

	return value
}

// evalStringCharAssignment handles string character mutation.
// DWScript strings are 1-indexed and support Unicode (rune-aware).
func (e *Evaluator) evalStringCharAssignment(
	strVal *runtime.StringValue,
	index int,
	value Value,
	stmt *ast.AssignmentStatement,
) Value {
	// Bounds check using rune length (DWScript strings are 1-based)
	strLen := RuneLength(strVal.Value)
	if index < 1 || index > strLen {
		return e.newError(stmt, "string index out of bounds: %d (string length is %d)", index, strLen)
	}

	// Value to assign must be a string (character); use first rune
	charVal, ok := value.(*runtime.StringValue)
	if !ok {
		return e.newError(stmt, "cannot assign %s to string index (expected STRING)", value.Type())
	}

	if RuneLength(charVal.Value) == 0 {
		return e.newError(stmt, "cannot assign empty string to string index")
	}

	// Get the first rune from the assigned string
	r, _ := RuneAt(charVal.Value, 1)

	// Replace rune at position
	if newStr, ok := RuneReplace(strVal.Value, index, r); ok {
		strVal.Value = newStr
		return value
	}

	return e.newError(stmt, "string index out of bounds: %d (string length is %d)", index, strLen)
}

// evalIndexedPropertyAssignment handles indexed property assignment: obj.Prop[i] := value
//
// This follows the pattern from executeRecordPropertyWrite (task 3.2.11d):
// 1. Evaluate the base object (obj.Prop) to get the property metadata
// 2. Extract the property setter method reference
// 3. Build argument list: [indices..., value]
// 4. Execute setter via adapter.ExecuteMethodWithSelf() (general OOP facility)
//
// Supports multi-index properties: obj.Prop[x, y] := value → args = [x, y, value]
//
// Uses general-purpose method dispatch instead of property-specific adapter interface.
func (e *Evaluator) evalIndexedPropertyAssignment(
	memberAccess *ast.MemberAccessExpression,
	indices []ast.Expression,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	// Evaluate the base object (e.g., obj in obj.Prop[i])
	baseObj := e.Eval(memberAccess.Object, ctx)
	if isError(baseObj) {
		return baseObj
	}

	// Check for exception during evaluation
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Get the property name
	propName := memberAccess.Member.Value

	// Evaluate all indices
	indexValues := make([]Value, 0, len(indices))
	for _, indexExpr := range indices {
		indexVal := e.Eval(indexExpr, ctx)
		if isError(indexVal) {
			return indexVal
		}
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}
		indexValues = append(indexValues, indexVal)
	}

	// Try to get property descriptor from the object
	// Different types have different property lookup mechanisms
	var propDesc *runtime.PropertyDescriptor

	// Check if object implements PropertyAccessor interface
	if accessor, ok := baseObj.(runtime.PropertyAccessor); ok {
		propDesc = accessor.LookupProperty(propName)
	}

	if propDesc == nil {
		return e.newError(stmt, "property '%s' not found on %s", propName, baseObj.Type())
	}

	// Check if property is indexed
	if !propDesc.IsIndexed {
		return e.newError(stmt, "property '%s' is not an indexed property", propName)
	}

	// Get the underlying PropertyInfo for setter access
	propInfo, ok := propDesc.Impl.(*runtime.PropertyInfo)
	if !ok {
		return e.newError(stmt, "invalid property metadata for '%s'", propName)
	}

	// Check if property has write access
	if propInfo.WriteSpec == "" {
		return e.newError(stmt, "property '%s' is read-only", propName)
	}

	// Build argument list for setter: [indices..., value]
	args := make([]Value, 0, len(indexValues)+1)
	args = append(args, indexValues...)
	args = append(args, value)

	// Execute setter method via general OOP facility (adapter.ExecuteMethodWithSelf)
	// This delegates to the interpreter for method execution, but it's a general
	// method dispatch mechanism, not a property-specific adapter interface
	result := e.oopEngine.ExecuteMethodWithSelf(baseObj, propInfo.WriteSpec, args)

	// Check for errors from method execution
	if isError(result) {
		return result
	}

	return value
}

// evalDefaultPropertyAssignment handles default indexed property assignment: obj[i] := value
//
// This follows the same pattern as evalIndexedPropertyAssignment (3.2.11g),
// but looks up the default property instead of a named property.
//
// Uses PropertyAccessor.GetDefaultProperty() which already exists in runtime:
// 1. Get PropertyAccessor from value (obj implements PropertyAccessor interface)
// 2. Call accessor.GetDefaultProperty() - returns PropertyDescriptor
// 3. Extract PropertyInfo with setter method reference
// 4. Build argument list: [index, value]
// 5. Execute setter via adapter.ExecuteMethodWithSelf() (general OOP facility)
//
// INTERFACE handling: InterfaceInstance.GetDefaultProperty() delegates to underlying interface
// OBJECT handling: ObjectInstance.GetDefaultProperty() uses IClassInfo.GetDefaultProperty()
//
// Uses general-purpose method dispatch instead of property-specific adapter interface.
func (e *Evaluator) evalDefaultPropertyAssignment(
	obj Value,
	indexVal Value,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	// Check if object implements PropertyAccessor interface
	accessor, ok := obj.(runtime.PropertyAccessor)
	if !ok {
		return e.newError(stmt, "type %s does not support indexed access", obj.Type())
	}

	// Get the default property - already exists in runtime!
	propDesc := accessor.GetDefaultProperty()
	if propDesc == nil {
		return e.newError(stmt, "type %s has no default indexed property", obj.Type())
	}

	// Check if property is indexed (default properties should be indexed)
	if !propDesc.IsIndexed {
		return e.newError(stmt, "default property on %s is not an indexed property", obj.Type())
	}

	// Get the underlying PropertyInfo for setter access
	propInfo, ok := propDesc.Impl.(*runtime.PropertyInfo)
	if !ok {
		return e.newError(stmt, "invalid default property metadata for %s", obj.Type())
	}

	// Check if property has write access
	if propInfo.WriteSpec == "" {
		return e.newError(stmt, "default property on %s is read-only", obj.Type())
	}

	// Build argument list for setter: [index, value]
	// Note: For default properties, we have a single index (not multi-index like named properties)
	args := []Value{indexVal, value}

	// Execute setter method via general OOP facility (adapter.ExecuteMethodWithSelf)
	// This delegates to the interpreter for method execution, but it's a general
	// method dispatch mechanism, not a property-specific adapter interface
	result := e.oopEngine.ExecuteMethodWithSelf(obj, propInfo.WriteSpec, args)

	// Check for errors from method execution
	if isError(result) {
		return result
	}

	return value
}
