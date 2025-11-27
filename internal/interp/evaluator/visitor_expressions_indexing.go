package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains visitor methods for indexing and record literal expression AST nodes.
// These handle array/string indexing, indexed property access, and record construction.

// VisitIndexExpression evaluates an index expression array[index].
// Handles array, string, property, and JSON indexing with bounds checking.
func (e *Evaluator) VisitIndexExpression(node *ast.IndexExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil index expression")
	}

	if node.Left == nil {
		return e.newError(node, "index expression missing base")
	}

	// Task 3.5.99g: Check if this is indexed property access via MemberAccessExpression
	// obj.Property[index1, index2, ...] should be flattened and passed to the property getter
	// We only flatten indices if the base is a MemberAccessExpression (property access)
	// For regular array access like arr[i][j], we process each level separately
	base, indices := CollectIndices(node)

	// Check if this is indexed property access: obj.Property[index1, index2, ...]
	// Only flatten indices for property access, not for regular arrays
	if memberAccess, ok := base.(*ast.MemberAccessExpression); ok {
		// Evaluate the object being accessed
		objVal := e.Eval(memberAccess.Object, ctx)
		if isError(objVal) {
			return objVal
		}

		// Task 3.5.99g: Handle interface indexed property access
		if intfInst, ok := objVal.(InterfaceInstanceValue); ok {
			underlying := intfInst.GetUnderlyingObjectValue()
			if underlying == nil {
				return e.newError(node, "interface is nil")
			}

			// Check if interface has the indexed property
			if accessor, ok := objVal.(PropertyAccessor); ok {
				if propDesc := accessor.LookupProperty(memberAccess.Member.Value); propDesc != nil && propDesc.IsIndexed {
					// Evaluate all indices
					indexVals := make([]Value, len(indices))
					for idx, indexExpr := range indices {
						indexVals[idx] = e.Eval(indexExpr, ctx)
						if isError(indexVals[idx]) {
							return indexVals[idx]
						}
					}

					// Call indexed property getter on the underlying object
					if underlying.Type() == "OBJECT" {
						return e.adapter.CallIndexedPropertyGetter(underlying, propDesc.Impl, indexVals, node)
					}
					return e.newError(node, "interface underlying object is not a class instance")
				}
			}

			// Unwrap for further checks
			objVal = underlying
		}

		// Task 3.5.99g: Handle object indexed property access
		if objVal.Type() == "OBJECT" {
			if accessor, ok := objVal.(PropertyAccessor); ok {
				if propDesc := accessor.LookupProperty(memberAccess.Member.Value); propDesc != nil && propDesc.IsIndexed {
					// Evaluate all indices
					indexVals := make([]Value, len(indices))
					for idx, indexExpr := range indices {
						indexVals[idx] = e.Eval(indexExpr, ctx)
						if isError(indexVals[idx]) {
							return indexVals[idx]
						}
					}

					// Call indexed property getter
					return e.adapter.CallIndexedPropertyGetter(objVal, propDesc.Impl, indexVals, node)
				}
			}
		}

		// Task 3.5.99g: Handle record indexed property access
		if objVal.Type() == "RECORD" {
			if accessor, ok := objVal.(PropertyAccessor); ok {
				if propDesc := accessor.LookupProperty(memberAccess.Member.Value); propDesc != nil {
					// Evaluate all indices
					indexVals := make([]Value, len(indices))
					for idx, indexExpr := range indices {
						indexVals[idx] = e.Eval(indexExpr, ctx)
						if isError(indexVals[idx]) {
							return indexVals[idx]
						}
					}

					// Call record property getter
					return e.adapter.CallRecordPropertyGetter(objVal, propDesc.Impl, indexVals, node)
				}
			}
		}

		// Not an indexed property - fall through to normal member access handling
		// This will likely error, but let it be handled by the regular logic below
	}

	// Not a property access - this is regular array/string indexing
	// Process ONLY the outermost index, not all nested indices
	// This allows FData[x][y] to work as: (FData[x])[y]
	leftVal := e.Eval(node.Left, ctx)
	if isError(leftVal) {
		return leftVal
	}

	if node.Index == nil {
		return e.newError(node, "index expression missing index")
	}

	// Evaluate the index for this level only
	indexVal := e.Eval(node.Index, ctx)
	if isError(indexVal) {
		return indexVal
	}

	// Unwrap variants for indexing
	leftVal = unwrapVariant(leftVal)

	// Handle JSON indexing directly - Task 3.5.99b
	if leftVal.Type() == "JSON" {
		return e.indexJSON(leftVal, indexVal, node)
	}

	// Task 3.5.99c: Handle object default property access
	if leftVal.Type() == "OBJECT" {
		if accessor, ok := leftVal.(PropertyAccessor); ok {
			if defaultProp := accessor.GetDefaultProperty(); defaultProp != nil {
				// Delegate indexed property method call to adapter
				return e.adapter.CallIndexedPropertyGetter(leftVal, defaultProp.Impl, []Value{indexVal}, node)
			}
		}
	}

	// Task 3.5.99d: Handle interface default property access
	if leftVal.Type() == "INTERFACE" {
		// Unwrap interface to get underlying object
		if ifaceVal, ok := leftVal.(InterfaceInstanceValue); ok {
			underlying := ifaceVal.GetUnderlyingObjectValue()
			if underlying == nil {
				return e.newError(node, "interface is nil")
			}

			// Check if interface has a default property
			if accessor, ok := leftVal.(PropertyAccessor); ok {
				if defaultProp := accessor.GetDefaultProperty(); defaultProp != nil && defaultProp.IsIndexed {
					// The property is defined on the interface, but we need the underlying object for execution
					if objVal := underlying; objVal.Type() == "OBJECT" {
						return e.adapter.CallIndexedPropertyGetter(objVal, defaultProp.Impl, []Value{indexVal}, node)
					}
					return e.newError(node, "interface underlying object is not a class instance")
				}
			}

			// No default property on interface, continue with unwrapped object
			// Check if the underlying object has a default property
			leftVal = underlying
			if leftVal.Type() == "OBJECT" {
				if accessor, ok := leftVal.(PropertyAccessor); ok {
					if defaultProp := accessor.GetDefaultProperty(); defaultProp != nil {
						return e.adapter.CallIndexedPropertyGetter(leftVal, defaultProp.Impl, []Value{indexVal}, node)
					}
				}
			}
		}
	}

	// Task 3.5.99e: Handle record default property access
	if leftVal.Type() == "RECORD" {
		// Check if record has a default property
		if accessor, ok := leftVal.(PropertyAccessor); ok {
			if defaultProp := accessor.GetDefaultProperty(); defaultProp != nil {
				// Delegate record property getter method call to adapter
				return e.adapter.CallRecordPropertyGetter(leftVal, defaultProp.Impl, []Value{indexVal}, node)
			}
		}
		// No default property, fall through to normal indexing (which will error)
	}

	// Index must be an integer or enum for arrays and strings
	index, ok := ExtractIntegerIndex(indexVal)
	if !ok {
		return e.newError(node, "index must be an integer or enum, got %s", indexVal.Type())
	}

	// Check if left side is an array
	if arrayVal, ok := leftVal.(*runtime.ArrayValue); ok {
		return e.IndexArray(arrayVal, index, node)
	}

	// Check if left side is a string
	if strVal, ok := leftVal.(*runtime.StringValue); ok {
		return e.IndexString(strVal, index, node)
	}

	return e.newError(node, "cannot index type %s", leftVal.Type())
}

// VisitRecordLiteralExpression evaluates a record literal expression.
// Handles typed and anonymous record literals with field initialization and default values.
// Task 3.5.40: Full migration from Interpreter.evalRecordLiteral()
func (e *Evaluator) VisitRecordLiteralExpression(node *ast.RecordLiteralExpression, ctx *ExecutionContext) Value {
	// Task 3.5.40: Record literal evaluation with field initialization
	//
	// Record literal syntax:
	// - Typed: TMyRecord(Field1: 1, Field2: 'hello')
	// - Anonymous: (Field1: 1, Field2: 'hello') - requires type context
	//
	// Type determination:
	// - Explicit type name in literal -> use that type
	// - Anonymous literal -> requires type context (set by variable/const declaration)
	//
	// Field initialization:
	// - Evaluate field value expressions
	// - Validate fields exist in record type
	// - Initialize missing fields with field initializers or zero values
	// - Handle nested records recursively
	// - Handle interface-typed fields specially
	//
	// Positional field initialization:
	// - Not yet supported (requires field.Name == nil)
	// - Returns error if attempted
	//
	// Field validation:
	// - All specified fields must exist in record type
	// - Field names are case-insensitive
	// - Duplicate field names not allowed (parser enforces)
	//
	// Default value initialization:
	// - Missing fields get field initializers if defined
	// - Otherwise get zero values for their types
	// - Nested records initialized recursively
	// - Interface fields get nil InterfaceInstance

	if node == nil {
		return e.newError(node, "nil record literal")
	}

	// Determine record type
	var recordTypeName string
	switch {
	case node.TypeName != nil:
		recordTypeName = node.TypeName.Value
	case ctx.RecordTypeContext() != "":
		// Anonymous literal with type context from caller (e.g., var/const declaration)
		recordTypeName = ctx.RecordTypeContext()
	default:
		// Anonymous literal requires type context (should have been set by caller)
		return e.newError(node, "record literal requires explicit type name or type context")
	}

	// Look up record type
	// Task 3.5.46: Use TypeSystem directly instead of adapter
	if !e.typeSystem.HasRecord(recordTypeName) {
		return e.newError(node, "unknown record type '%s'", recordTypeName)
	}

	// Evaluate field values
	fieldValues := make(map[string]Value)
	for _, field := range node.Fields {
		// Skip positional fields (not yet implemented)
		if field.Name == nil {
			return e.newError(node, "positional record field initialization not yet supported")
		}

		fieldName := field.Name.Value

		// Evaluate the field value expression
		fieldValue := e.Eval(field.Value, ctx)
		if isError(fieldValue) {
			return fieldValue
		}

		// Store the field value (case-insensitive)
		fieldValues[fieldName] = fieldValue
	}

	// Create record value with field initialization
	recordValue, err := e.adapter.CreateRecordValue(recordTypeName, fieldValues)
	if err != nil {
		return e.newError(node, "%v", err)
	}

	return recordValue
}
