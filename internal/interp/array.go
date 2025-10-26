package interp

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Array Declaration Evaluation (Task 8.128-8.129)
// ============================================================================

// evalArrayDeclaration evaluates an array type declaration.
// Example: type TMyArray = array[1..10] of Integer;
// Task 8.128: Register array types in the environment for later use.
func (i *Interpreter) evalArrayDeclaration(decl *ast.ArrayDecl) Value {
	if decl == nil {
		return &ErrorValue{Message: "nil array declaration"}
	}

	arrayName := decl.Name.Value

	// Build the array type from the declaration
	arrayTypeAnnotation := decl.ArrayType
	if arrayTypeAnnotation == nil {
		return i.newErrorWithLocation(decl, "invalid array type declaration")
	}

	// Resolve the element type
	elementTypeName := arrayTypeAnnotation.ElementType.Name
	elementType, err := i.resolveType(elementTypeName)
	if err != nil {
		return i.newErrorWithLocation(decl, "unknown element type '%s'", elementTypeName)
	}

	// Create the array type
	var arrayType *types.ArrayType
	if arrayTypeAnnotation.IsDynamic() {
		arrayType = types.NewDynamicArrayType(elementType)
	} else {
		arrayType = types.NewStaticArrayType(elementType, *arrayTypeAnnotation.LowBound, *arrayTypeAnnotation.HighBound)
	}

	// Store array type in environment with a special prefix
	// This allows var declarations to look up the type
	typeKey := "__array_type_" + arrayName
	arrayTypeValue := &ArrayTypeValue{
		Name:      arrayName,
		ArrayType: arrayType,
	}
	i.env.Define(typeKey, arrayTypeValue) // Use Define, not Set

	return &NilValue{} // Type declarations don't return a value
}

// evalIndexExpression evaluates array/string indexing: arr[i]
// Task 8.129: Implement array indexing (read).
func (i *Interpreter) evalIndexExpression(expr *ast.IndexExpression) Value {
	if expr == nil {
		return &ErrorValue{Message: "nil index expression"}
	}

	// Evaluate the left side (what's being indexed)
	leftVal := i.Eval(expr.Left)
	if isError(leftVal) {
		return leftVal
	}

	// Evaluate the index
	indexVal := i.Eval(expr.Index)
	if isError(indexVal) {
		return indexVal
	}

	// Index must be an integer
	indexInt, ok := indexVal.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(expr, "index must be an integer, got %s", indexVal.Type())
	}
	index := int(indexInt.Value)

	// Check if left side is an array
	if arrayVal, ok := leftVal.(*ArrayValue); ok {
		return i.indexArray(arrayVal, index, expr)
	}

	// Check if left side is a string
	if strVal, ok := leftVal.(*StringValue); ok {
		return i.indexString(strVal, index, expr)
	}

	return i.newErrorWithLocation(expr, "cannot index type %s", leftVal.Type())
}

// indexArray performs array indexing with bounds checking.
func (i *Interpreter) indexArray(arr *ArrayValue, index int, expr *ast.IndexExpression) Value {
	if arr.ArrayType == nil {
		return i.newErrorWithLocation(expr, "array has no type information")
	}

	// Convert logical index to physical index
	var physicalIndex int
	if arr.ArrayType.IsStatic() {
		// Static array: check bounds and adjust for low bound
		lowBound := *arr.ArrayType.LowBound
		highBound := *arr.ArrayType.HighBound

		if index < lowBound || index > highBound {
			return i.newErrorWithLocation(expr, "index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
		}

		physicalIndex = index - lowBound
	} else {
		// Dynamic array: zero-based indexing
		if index < 0 || index >= len(arr.Elements) {
			return i.newErrorWithLocation(expr, "index out of bounds: %d (array length is %d)", index, len(arr.Elements))
		}

		physicalIndex = index
	}

	// Check physical bounds
	if physicalIndex < 0 || physicalIndex >= len(arr.Elements) {
		return i.newErrorWithLocation(expr, "index out of bounds: physical index %d, length %d", physicalIndex, len(arr.Elements))
	}

	// Return the element
	elem := arr.Elements[physicalIndex]
	if elem == nil {
		// Return nil value for uninitialized elements
		return &NilValue{}
	}

	return elem
}

// indexString performs string indexing (returns a single-character string).
func (i *Interpreter) indexString(str *StringValue, index int, expr *ast.IndexExpression) Value {
	// DWScript strings are 1-indexed
	if index < 1 || index > len(str.Value) {
		return i.newErrorWithLocation(expr, "string index out of bounds: %d (string length is %d)", index, len(str.Value))
	}

	// Convert to 0-based index
	char := string(str.Value[index-1])
	return &StringValue{Value: char}
}

// ArrayTypeValue is an internal value that stores array type metadata in the environment.
type ArrayTypeValue struct {
	ArrayType *types.ArrayType
	Name      string
}

// Type returns "ARRAY_TYPE".
func (a *ArrayTypeValue) Type() string {
	return "ARRAY_TYPE"
}

// String returns the array type name.
func (a *ArrayTypeValue) String() string {
	return "array type " + a.Name
}
