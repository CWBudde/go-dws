package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
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

	// Collect indices - flatten for property access, not for regular arrays
	base, indices := CollectIndices(node)

	// Check if this is indexed property access: obj.Property[index1, index2, ...]
	if memberAccess, ok := base.(*ast.MemberAccessExpression); ok {
		// Evaluate the object being accessed
		objVal := e.Eval(memberAccess.Object, ctx)
		if isError(objVal) {
			return objVal
		}

		// Handle interface indexed property access
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

					// Call indexed property getter on underlying object
					if underlying.Type() == "OBJECT" {
						if objVal, ok := underlying.(ObjectValue); ok {
							return objVal.ReadIndexedProperty(propDesc.Impl, indexVals, func(pi any, idx []Value) Value {
								return e.executeIndexedPropertyRead(underlying, pi, idx, node, ctx)
							})
						}
					}
					return e.newError(node, "interface underlying object is not a class instance")
				}
			}

			// Unwrap for further checks
			objVal = underlying
		}

		// Handle object indexed property access
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

					// Call indexed property getter via ObjectValue interface
						if ov, ok := objVal.(ObjectValue); ok {
						return ov.ReadIndexedProperty(propDesc.Impl, indexVals, func(pi any, idx []Value) Value {
							return e.executeIndexedPropertyRead(objVal, pi, idx, node, ctx)
						})
					}
				}
			}
		}

		// Handle record indexed property access
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

						recVal, ok := objVal.(RecordInstanceValue)
					if !ok {
						return e.newError(node, "internal error: RECORD value does not implement RecordInstanceValue interface")
					}
					return recVal.ReadIndexedProperty(propDesc.Impl, indexVals, func(pi any, idx []Value) Value {
						return e.oopEngine.ExecuteRecordPropertyRead(objVal, pi, idx, node)
					})
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

	// Handle JSON indexing
	if leftVal.Type() == "JSON" {
		return e.indexJSON(leftVal, indexVal, node)
	}

	// Handle object default property access
	if leftVal.Type() == "OBJECT" {
		if accessor, ok := leftVal.(PropertyAccessor); ok {
			if defaultProp := accessor.GetDefaultProperty(); defaultProp != nil {
				if objVal, ok := leftVal.(ObjectValue); ok {
					return objVal.ReadIndexedProperty(defaultProp.Impl, []Value{indexVal}, func(pi any, idx []Value) Value {
						return e.executeIndexedPropertyRead(leftVal, pi, idx, node, ctx)
					})
				}
			}
		}
	}

	// Handle interface default property access
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
					if underlying.Type() == "OBJECT" {
										if objVal, ok := underlying.(ObjectValue); ok {
							return objVal.ReadIndexedProperty(defaultProp.Impl, []Value{indexVal}, func(pi any, idx []Value) Value {
								return e.executeIndexedPropertyRead(underlying, pi, idx, node, ctx)
							})
						}
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
										if objVal, ok := leftVal.(ObjectValue); ok {
							return objVal.ReadIndexedProperty(defaultProp.Impl, []Value{indexVal}, func(pi any, idx []Value) Value {
								return e.executeIndexedPropertyRead(leftVal, pi, idx, node, ctx)
							})
						}
					}
				}
			}
		}
	}

	// Handle record default property access
	if leftVal.Type() == "RECORD" {
		// Check if record has a default property
		if accessor, ok := leftVal.(PropertyAccessor); ok {
			if defaultProp := accessor.GetDefaultProperty(); defaultProp != nil {
				recVal, ok := leftVal.(RecordInstanceValue)
				if !ok {
					return e.newError(node, "internal error: RECORD value does not implement RecordInstanceValue interface")
				}
				return recVal.ReadIndexedProperty(defaultProp.Impl, []Value{indexVal}, func(pi any, idx []Value) Value {
					return e.oopEngine.ExecuteRecordPropertyRead(leftVal, pi, idx, node)
				})
			}
		}
		// No default property, fall through to normal indexing (which will error)
	}

	// Index must be an integer or enum for arrays and strings
	index, ok := ExtractIntegerIndex(indexVal)
	if !ok {
		return e.newError(node, "index must be an ordinal value, got %s", indexVal.Type())
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

// VisitRecordLiteralExpression evaluates record literal expressions like TMyRecord(Field1: 1, Field2: 'hello').
// Handles typed and anonymous literals with field initialization and default values.
func (e *Evaluator) VisitRecordLiteralExpression(node *ast.RecordLiteralExpression, ctx *ExecutionContext) Value {
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

	// Look up record type via TypeSystem
	recordTypeAny := e.typeSystem.LookupRecord(recordTypeName)
	if recordTypeAny == nil {
		return e.newError(node, "unknown record type '%s'", recordTypeName)
	}

	// Type-assert to access RecordType, Metadata, and FieldDecls
	// This is safe because TypeSystem stores *RecordTypeValue
	type recordTypeAccess interface {
		GetRecordType() *types.RecordType
		GetMetadata() any
	}

	recordTypeAccessor, ok := recordTypeAny.(recordTypeAccess)
	if !ok {
		return e.newError(node, "failed to access record type '%s'", recordTypeName)
	}

	recordType := recordTypeAccessor.GetRecordType()
	if recordType == nil {
		return e.newError(node, "failed to extract record type for '%s'", recordTypeName)
	}

	// Extract Metadata (may be nil)
	var metadata *runtime.RecordMetadata
	if mdAny := recordTypeAccessor.GetMetadata(); mdAny != nil {
		if md, ok := mdAny.(*runtime.RecordMetadata); ok {
			metadata = md
		}
	}

	// Extract FieldDecls using struct field access
	// Since we know the concrete type is *RecordTypeValue from interp package
	var fieldDecls map[string]*ast.FieldDecl
	type hasFieldDecls interface {
		GetFieldDecls() map[string]*ast.FieldDecl
	}
	if rtVal, ok := recordTypeAny.(hasFieldDecls); ok {
		fieldDecls = rtVal.GetFieldDecls()
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

	// Create field initializer callback for runtime constructor
	initializer := func(fieldName string, fieldType types.Type) runtime.Value {
		// Check if field was provided in literal (case-insensitive lookup)
		for providedName, val := range fieldValues {
			if ident.Equal(providedName, fieldName) {
				return val
			}
		}

		// Field not in literal - need to initialize it
		fieldNameNorm := ident.Normalize(fieldName)

		// Check for field initializer expression in FieldDecls
		if fieldDecls != nil {
			if fieldDecl, hasDecl := fieldDecls[fieldNameNorm]; hasDecl && fieldDecl.InitValue != nil {
				// Evaluate the field initializer AST expression directly
				fieldValue := e.Eval(fieldDecl.InitValue, ctx)
				if isError(fieldValue) {
					// Return error value (constructor will propagate it)
					return fieldValue
				}
				return fieldValue
			}
		}

		// No initializer - generate zero value
		return e.getZeroValueForType(fieldType)
	}

	// Create record using runtime constructor with initializer callback
	recordValue := runtime.NewRecordValueWithInitializer(recordType, metadata, initializer)

	return recordValue
}
