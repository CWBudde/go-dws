package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Index Operations
// ============================================================================

// IndexArray performs array indexing with bounds checking.
// Returns the element at the given index or an error if out of bounds.
//
// For static arrays, the index is checked against low/high bounds and
// converted to a physical index. For dynamic arrays, zero-based indexing
// is used.
func (e *Evaluator) IndexArray(arr *runtime.ArrayValue, index int, node ast.Node, ctx *ExecutionContext) Value {
	if arr.ArrayType == nil {
		return e.newError(node, "array has no type information")
	}

	// Convert logical index to physical index.
	// Read diagnostics point one past the index's closing bracket (DWScript).
	// ⚠️ SimpleScripts/const_array_empty wants the bracket itself, one column
	// earlier; ArrayPass/array_element_byref wants what is here. They differ
	// because upstream reports a by-reference bind one column further on than a
	// plain read, and go-dws routes `a[a.High+1]` down the read path by mistake
	// (see PLAN.md §3.5 E8). Fix the routing before moving this anchor.
	var physicalIndex int
	if arr.ArrayType.IsStatic() {
		// Static array: check bounds and adjust for low bound
		lowBound := *arr.ArrayType.LowBound
		highBound := *arr.ArrayType.HighBound

		if index < lowBound {
			return e.raiseIndexBoundExceededAt(node.End(), index, false, ctx)
		}
		if index > highBound {
			return e.raiseIndexBoundExceededAt(node.End(), index, true, ctx)
		}

		physicalIndex = index - lowBound
	} else {
		// Dynamic array: zero-based indexing
		if index < 0 {
			return e.raiseIndexBoundExceededAt(node.End(), index, false, ctx)
		}
		if index >= len(arr.Elements) {
			return e.raiseIndexBoundExceededAt(node.End(), index, true, ctx)
		}

		physicalIndex = index
	}

	// Check physical bounds
	if physicalIndex < 0 || physicalIndex >= len(arr.Elements) {
		return e.newError(node, "index out of bounds: physical index %d, length %d", physicalIndex, len(arr.Elements))
	}

	// Return the element
	elem := arr.Elements[physicalIndex]
	if elem == nil {
		// Return properly typed zero value for uninitialized elements
		return e.getZeroValueForType(arr.ArrayType.ElementType, ctx)
	}

	return elem
}

// IndexString performs string indexing (returns a single-character string).
// DWScript strings are 1-indexed.
//
// An out-of-range index raises the same catchable "Lower/Upper bound exceeded!"
// exception an array does, not a fatal error — SimpleScripts/string_bounds writes
// through four such accesses inside try/except and expects none of them to reach
// the statement after the assignment. The anchor is the opening bracket
// (SimpleScripts/string_bounds2 wants column 10 of `PrintLn(s[0])`), one column
// short of where an array read is reported; the two are separate expression
// classes upstream and position themselves separately.
func (e *Evaluator) IndexString(str *runtime.StringValue, index int, node ast.Node, ctx *ExecutionContext) Value {
	// DWScript strings are 1-indexed
	// Use rune-based indexing to handle UTF-8 correctly
	strLen := RuneLength(str.Value)
	if index < 1 || index > strLen {
		return e.raiseIndexBoundExceededAt(stringIndexBracketPos(node), index, index > strLen, ctx)
	}

	// Get the character at the given position
	char, ok := RuneAt(str.Value, index)
	if !ok {
		return e.raiseIndexBoundExceededAt(stringIndexBracketPos(node), index, true, ctx)
	}
	return &runtime.StringValue{Value: string(char)}
}

// stringIndexBracketPos returns the position of the opening bracket of `s[i]`,
// derived from the end of the indexed expression. Falls back to the node's own
// position when the node is not an index expression.
func stringIndexBracketPos(node ast.Node) token.Position {
	if idx, ok := node.(*ast.IndexExpression); ok && idx.Left != nil {
		return idx.Left.End()
	}
	if node == nil {
		return token.Position{}
	}
	return node.Pos()
}

// getZeroValueForType returns the zero/default value for a given type.
// This is used when accessing uninitialized array elements and record field initialization.
func (e *Evaluator) getZeroValueForType(t types.Type, ctx *ExecutionContext) runtime.Value {
	if t == nil {
		return &runtime.NilValue{}
	}
	t = types.GetUnderlyingType(t)

	switch t.TypeKind() {
	case "INTEGER":
		return &runtime.IntegerValue{Value: 0}
	case "FLOAT":
		return &runtime.FloatValue{Value: 0.0}
	case "STRING":
		return &runtime.StringValue{Value: ""}
	case "BOOLEAN":
		return &runtime.BooleanValue{Value: false}
	case "FUNCTION_POINTER":
		// A proc-typed variable/field holds a nil function pointer carrying its
		// declared signature, so a bare call raises "Function pointer is nil" and
		// Assigned() reports false until a routine is bound.
		if fpType, ok := t.(*types.FunctionPointerType); ok {
			return &runtime.FunctionPointerValue{PointerType: fpType}
		}
		return &runtime.NilValue{}
	case "METHOD_POINTER":
		if mpType, ok := t.(*types.MethodPointerType); ok {
			return &runtime.FunctionPointerValue{PointerType: &mpType.FunctionPointerType}
		}
		return &runtime.NilValue{}
	case "ARRAY":
		// For array types, create an empty array with proper type
		if arrayType, ok := t.(*types.ArrayType); ok {
			return runtime.NewArrayValue(arrayType, nil)
		}
		return &runtime.NilValue{}
	case "ASSOCIATIVE_ARRAY":
		if assocType, ok := t.(*types.AssociativeArrayType); ok {
			return runtime.NewAssociativeArrayValue(assocType)
		}
		return &runtime.NilValue{}
	case "SET":
		// An uninitialized set is the empty set, not nil, so `x in r.Field`
		// works on a field a record literal did not name.
		if setType, ok := t.(*types.SetType); ok {
			return runtime.NewSetValue(setType)
		}
		return &runtime.NilValue{}
	case "RECORD":
		// Recursively create nested records, applying each field's default
		// initializer expression (e.g. `Field : Integer = 1`) when present so
		// nested record fields match a top-level `var r : TRec` declaration.
		if recordType, ok := t.(*types.RecordType); ok {
			nestedMetadata := e.typeSystem.LookupRecordMetadata(recordType.Name)

			// Field default initializers live on the registered record type
			// value (RecordTypeValue.FieldDecls), not in the metadata.
			var fieldDecls map[string]*ast.FieldDecl
			if recordType.Name != "" {
				if recordTypeAny := e.typeSystem.LookupRecord(recordType.Name); recordTypeAny != nil {
					fieldDecls = recordTypeAny.GetFieldDecls()
				}
			}

			zeroInit := func(nestedFieldName string, nestedFieldType types.Type) runtime.Value {
				if fieldDecls != nil {
					if fieldDecl, ok := fieldDecls[ident.Normalize(nestedFieldName)]; ok && fieldDecl.InitValue != nil {
						if ctx == nil {
							return e.getZeroValueForType(nestedFieldType, ctx)
						}
						// Establish the field's record type context so record-literal
						// initializers (e.g. `Sub : TChild = (A: 1)`) resolve, mirroring
						// the top-level record-init path.
						prevRecordType := ctx.RecordTypeContext()
						if nestedRecordType, ok := types.GetUnderlyingType(nestedFieldType).(*types.RecordType); ok {
							ctx.SetRecordTypeContext(nestedRecordType)
						}
						fieldValue := e.Eval(fieldDecl.InitValue, ctx)
						ctx.SetRecordTypeContext(prevRecordType)
						// Propagate errors (including from initializers) instead of
						// masking them with a zero value.
						return fieldValue
					}
				}
				return e.getZeroValueForType(nestedFieldType, ctx)
			}
			return runtime.NewRecordValueWithInitializer(recordType, nestedMetadata, zeroInit)
		}
		return &runtime.NilValue{}
	case "INTERFACE":
		// Create a proper InterfaceInstance with nil object for uninitialized interface fields.
		// This preserves the interface type information needed for proper error messages.
		if ifaceType, ok := t.(*types.InterfaceType); ok {
			ifaceName := ifaceType.Name
			if ifaceName != "" {
				if ifaceInfoAny := e.typeSystem.LookupInterface(ifaceName); ifaceInfoAny != nil {
					return runtime.NewInterfaceInstance(ifaceInfoAny, nil)
				}
			}
		}
		// Fallback to NilValue if interface info not found
		return &runtime.NilValue{}
	case "CLASS":
		// Class fields initialize as nil
		if classType, ok := t.(*types.ClassType); ok {
			return &runtime.NilValue{ClassType: classType.Name}
		}
		return &runtime.NilValue{}
	case "VARIANT":
		// Variant fields initialize as nil (VariantValue is in interp package)
		// For now, return nil - the adapter will handle variant initialization if needed
		return &runtime.NilValue{}
	default:
		// For other types, return nil
		return &runtime.NilValue{}
	}
}

// ExtractIntegerIndex extracts an integer index from a Value.
// Returns the index and true if successful, or 0 and false if the value
// is not an integer or enum type.
func ExtractIntegerIndex(indexVal Value) (int, bool) {
	switch iv := indexVal.(type) {
	case *runtime.IntegerValue:
		return int(iv.Value), true
	case *runtime.BooleanValue:
		if iv.Value {
			return 1, true
		}
		return 0, true
	case *runtime.EnumValue:
		return iv.OrdinalValue, true
	default:
		return 0, false
	}
}

// ============================================================================
// Multi-Dimensional Array Construction
// ============================================================================

// CreateMultiDimArray creates a multi-dimensional array with the given dimensions.
// For 1D arrays, creates a single array with the specified size.
// For multi-dimensional arrays, recursively creates nested arrays.
//
// Example:
//
//	new Integer[10] → single 1D array with 10 elements
//	new String[3, 4] → 3x4 nested arrays
func (e *Evaluator) CreateMultiDimArray(elementType types.Type, dimensions []int, ctx *ExecutionContext) *runtime.ArrayValue {
	if len(dimensions) == 0 {
		// This shouldn't happen, but handle gracefully
		return &runtime.ArrayValue{
			ArrayType: types.NewDynamicArrayType(elementType),
			Elements:  []runtime.Value{},
		}
	}

	size := dimensions[0]

	if len(dimensions) == 1 {
		// Base case: 1D array
		arrayType := types.NewDynamicArrayType(elementType)

		// Create elements filled with zero values
		elements := make([]runtime.Value, size)
		for idx := 0; idx < size; idx++ {
			elements[idx] = e.getZeroValueForType(elementType, ctx)
		}

		return &runtime.ArrayValue{
			ArrayType: arrayType,
			Elements:  elements,
		}
	}

	// Recursive case: multi-dimensional array
	// The element type for this level is an array of the remaining dimensions
	innerElementType := buildArrayTypeForDimensions(elementType, dimensions[1:])

	// Create the outer array type
	arrayType := types.NewDynamicArrayType(innerElementType)

	// Create elements, each being an array of the remaining dimensions
	elements := make([]runtime.Value, size)
	for idx := 0; idx < size; idx++ {
		elements[idx] = e.CreateMultiDimArray(elementType, dimensions[1:], ctx)
	}

	return &runtime.ArrayValue{
		ArrayType: arrayType,
		Elements:  elements,
	}
}

// buildArrayTypeForDimensions builds an array type for the given dimensions.
// For example, dimensions [3, 4] with elementType Integer produces:
// array of array of Integer
func buildArrayTypeForDimensions(elementType types.Type, dimensions []int) types.Type {
	if len(dimensions) == 0 {
		return elementType
	}

	// Build from innermost to outermost
	currentType := elementType
	for range dimensions {
		currentType = types.NewDynamicArrayType(currentType)
	}

	return currentType
}
