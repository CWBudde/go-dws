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
// Task 3.5.105c: Migrated from Interpreter.evalIndexAssignment()
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
	base, _ := CollectIndices(target)

	// If base is a MemberAccessExpression, it might be an indexed property
	// Delegate to adapter for property-based access
	if _, ok := base.(*ast.MemberAccessExpression); ok {
		return e.adapter.EvalNode(stmt)
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

	// Check for interface-based indexed properties
	// InterfaceInstance is in interp package, check by type name
	if strings.HasPrefix(arrayVal.Type(), "INTERFACE") {
		// Delegate to adapter for interface handling
		return e.adapter.EvalNode(stmt)
	}

	// Check for object with default indexed property
	if strings.HasPrefix(arrayVal.Type(), "OBJECT[") {
		// Delegate to adapter for object default property handling
		return e.adapter.EvalNode(stmt)
	}

	// Extract integer index
	index, ok := ExtractIntegerIndex(indexVal)
	if !ok {
		return e.newError(stmt, "array index must be an integer, got %s", indexVal.Type())
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
