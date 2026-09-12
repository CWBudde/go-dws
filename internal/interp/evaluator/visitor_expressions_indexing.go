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
					if runtime.KindOf(underlying) == runtime.KindObject {
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
		if runtime.KindOf(objVal) == runtime.KindObject {
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
		if recVal, ok := objVal.(RecordInstanceValue); ok {
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

					return recVal.ReadIndexedProperty(propDesc.Impl, indexVals, func(pi any, idx []Value) Value {
						return e.executeRecordIndexedPropertyRead(objVal, pi, idx, node, ctx)
					})
				}
			}
		}

		// Handle indexed property access through a class name. The property need not
		// be a `class property`: DWScript allows an ordinary indexed property whose
		// accessor is a class method to be read through the class, since the accessor
		// needs no instance.
		if classMetaVal, ok := objVal.(ClassMetaValue); ok {
			if result, handled := e.evalClassMetaIndexedProperty(objVal, classMetaVal, memberAccess.Member.Value, indices, node, ctx); handled {
				return result
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

	return e.indexResolvedValue(leftVal, node, ctx)
}

// indexResolvedValue applies one level of indexing to an already-evaluated
// container. It is the tail of VisitIndexExpression, split out so that lvalue
// resolution (which must resolve the container itself, vivifying missing
// associative-array slots along the way) can reuse the read semantics without
// re-evaluating the base expression.
func (e *Evaluator) indexResolvedValue(leftVal Value, node *ast.IndexExpression, ctx *ExecutionContext) Value {
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
	if runtime.KindOf(leftVal) == runtime.KindJSON {
		return e.indexJSON(leftVal, indexVal, node)
	}

	// Handle default property access on objects, interfaces and records. An
	// interface with no default property of its own is replaced by its
	// underlying object, so indexing continues against that.
	if result, handled := e.indexViaDefaultProperty(&leftVal, indexVal, node, ctx); handled {
		return result
	}

	// Associative array read: a[key]. The key is an arbitrary value (not an
	// ordinal index). A missing key returns the element's zero value without
	// inserting it.
	if assoc, ok := leftVal.(*runtime.AssociativeArrayValue); ok {
		key, errVal := e.coerceAssociativeKey(assoc, indexVal, ctx)
		if errVal != nil {
			return errVal
		}
		if stored, present := assoc.Get(key); present {
			// Return the live stored value (like regular array indexing); value
			// semantics for record/static-array elements are enforced at
			// assignment time (`var b := a[k]` clones).
			return stored
		}
		return e.getZeroValueForType(assoc.ElementType(), ctx)
	}

	// Index must be an integer or enum for arrays and strings.
	// Variant indexes are cast per DWScript rules (may raise).
	index, ok := e.ExtractIndexWithVariantCast(indexVal, ctx)
	if !ok {
		if ctx.Exception() != nil {
			return e.nilValue()
		}
		return e.newError(node, "index must be an ordinal value, got %s", indexVal.Type())
	}

	// Check if left side is an array
	if arrayVal, ok := leftVal.(*runtime.ArrayValue); ok {
		return e.IndexArray(arrayVal, index, node, ctx)
	}

	// Check if left side is a string
	if strVal, ok := leftVal.(*runtime.StringValue); ok {
		return e.IndexString(strVal, index, node, ctx)
	}

	return e.newError(node, "cannot index type %s", leftVal.Type())
}

// indexViaDefaultProperty resolves `x[i]` where x is an object, interface or
// record carrying a default (indexed) property. It reports whether it handled
// the access. When the receiver is an interface without a default property of
// its own, *leftVal is replaced by the underlying object so the caller keeps
// indexing against that.
func (e *Evaluator) indexViaDefaultProperty(
	leftVal *Value,
	indexVal Value,
	node *ast.IndexExpression,
	ctx *ExecutionContext,
) (Value, bool) {
	switch runtime.KindOf(*leftVal) {
	case runtime.KindObject:
		return e.readObjectDefaultProperty(*leftVal, indexVal, node, ctx)

	case runtime.KindInterface:
		ifaceVal, ok := (*leftVal).(InterfaceInstanceValue)
		if !ok {
			return nil, false
		}
		underlying := ifaceVal.GetUnderlyingObjectValue()
		if underlying == nil {
			return e.newError(node, "interface is nil"), true
		}

		// A default property declared on the interface itself is executed
		// against the underlying instance.
		if accessor, ok := (*leftVal).(PropertyAccessor); ok {
			if defaultProp := accessor.GetDefaultProperty(); defaultProp != nil && defaultProp.IsIndexed {
				if runtime.KindOf(underlying) == runtime.KindObject {
					if objVal, ok := underlying.(ObjectValue); ok {
						return objVal.ReadIndexedProperty(defaultProp.Impl, []Value{indexVal}, func(pi any, idx []Value) Value {
							return e.executeIndexedPropertyRead(underlying, pi, idx, node, ctx)
						}), true
					}
				}
				return e.newError(node, "interface underlying object is not a class instance"), true
			}
		}

		// Otherwise fall through to the underlying object.
		*leftVal = underlying
		if runtime.KindOf(underlying) == runtime.KindObject {
			return e.readObjectDefaultProperty(underlying, indexVal, node, ctx)
		}
		return nil, false
	}

	// Records: no default property means ordinary indexing (which will error).
	if recVal, ok := (*leftVal).(RecordInstanceValue); ok {
		if accessor, ok := (*leftVal).(PropertyAccessor); ok {
			if defaultProp := accessor.GetDefaultProperty(); defaultProp != nil {
				obj := *leftVal
				return recVal.ReadIndexedProperty(defaultProp.Impl, []Value{indexVal}, func(pi any, idx []Value) Value {
					return e.executeRecordIndexedPropertyRead(obj, pi, idx, node, ctx)
				}), true
			}
		}
	}

	return nil, false
}

// readObjectDefaultProperty reads a class instance's default indexed property,
// reporting whether the instance has one.
func (e *Evaluator) readObjectDefaultProperty(
	obj Value,
	indexVal Value,
	node *ast.IndexExpression,
	ctx *ExecutionContext,
) (Value, bool) {
	accessor, ok := obj.(PropertyAccessor)
	if !ok {
		return nil, false
	}
	defaultProp := accessor.GetDefaultProperty()
	if defaultProp == nil {
		return nil, false
	}
	objVal, ok := obj.(ObjectValue)
	if !ok {
		return nil, false
	}
	return objVal.ReadIndexedProperty(defaultProp.Impl, []Value{indexVal}, func(pi any, idx []Value) Value {
		return e.executeIndexedPropertyRead(obj, pi, idx, node, ctx)
	}), true
}

// VisitRecordLiteralExpression evaluates record literal expressions like TMyRecord(Field1: 1, Field2: 'hello').
// Handles typed and anonymous literals with field initialization and default values.
func (e *Evaluator) VisitRecordLiteralExpression(node *ast.RecordLiteralExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil record literal")
	}

	// Determine record type
	var recordTypeName string
	var recordType *types.RecordType
	var metadata *runtime.RecordMetadata
	var fieldDecls map[string]*ast.FieldDecl
	switch {
	case node.TypeName != nil:
		recordTypeName = node.TypeName.Value
		if resolved, err := e.resolveTypeReference(node, recordTypeName, ctx); err == nil {
			if resolvedRecord, ok := types.GetUnderlyingType(resolved).(*types.RecordType); ok {
				recordType = resolvedRecord
			}
		}
	case ctx.RecordTypeContext() != nil:
		recordType = ctx.RecordTypeContext()
		recordTypeName = recordType.Name
	default:
		// Anonymous literal requires type context (should have been set by caller)
		return e.newError(node, "record literal requires explicit type name or type context")
	}

	if recordType == nil {
		// Look up record type via TypeSystem
		recordTypeAny := e.typeSystem.LookupRecord(recordTypeName)
		if recordTypeAny == nil {
			return e.newError(node, "unknown record type '%s'", recordTypeName)
		}

		recordTypeAccessor := recordTypeAny

		recordType = recordTypeAccessor.GetRecordType()
		if recordType == nil {
			return e.newError(node, "failed to extract record type for '%s'", recordTypeName)
		}

		metadata = recordTypeAccessor.GetMetadata()

		fieldDecls = recordTypeAny.GetFieldDecls()
	}

	if registered := e.typeSystem.LookupRecord(recordType.Name); registered != nil {
		metadata = registered.GetMetadata()
		fieldDecls = registered.GetFieldDecls()
	}

	// Evaluate field values
	fieldValues := make(map[string]Value)
	for _, field := range node.Fields {
		// Skip positional fields (not yet implemented)
		if field.Name == nil {
			return e.newError(node, "positional record field initialization not yet supported")
		}

		fieldName := field.Name.Value

		// Validate that the field exists in the record type
		fieldNameNorm := ident.Normalize(fieldName)
		if !recordType.HasField(fieldNameNorm) {
			return e.newError(node, "field '%s' does not exist in record type '%s'", fieldName, recordTypeName)
		}

		expectedFieldType := recordType.Fields[fieldNameNorm]
		prevRecordType := ctx.RecordTypeContext()
		if nestedRecordType, ok := types.GetUnderlyingType(expectedFieldType).(*types.RecordType); ok {
			ctx.SetRecordTypeContext(nestedRecordType)
		}

		// Evaluate the field value expression
		fieldValue := e.Eval(field.Value, ctx)
		ctx.SetRecordTypeContext(prevRecordType)
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
		return e.getZeroValueForType(fieldType, ctx)
	}

	// Create record using runtime constructor with initializer callback
	recordValue := runtime.NewRecordValueWithInitializer(recordType, metadata, initializer)

	return recordValue
}

// evalClassMetaIndexedProperty reads an indexed property through a class name, e.g.
// `TConvert.Prop[i]`. It reports handled=false when the class declares no such
// indexed property, so the caller can fall through to ordinary member access and
// produce the usual not-found diagnostic.
//
// Only accessors that need no instance are reachable this way: a class method, or an
// expression accessor evaluated in class context. An indexed property backed by an
// instance method is a compile-time error (see analyzeIndexedPropertyAccess); the
// runtime check here is the backstop for anything that slips past it.
func (e *Evaluator) evalClassMetaIndexedProperty(
	obj Value,
	classMetaVal ClassMetaValue,
	memberName string,
	indices []ast.Expression,
	node ast.Node,
	ctx *ExecutionContext,
) (Value, bool) {
	classInfo := classMetaVal.GetClassInfo()
	if classInfo == nil {
		return nil, false
	}
	propDesc := classInfo.LookupProperty(memberName)
	if propDesc == nil || !propDesc.IsIndexed {
		return nil, false
	}
	pInfo, ok := unwrapPropertyInfo(propDesc.Impl)
	if !ok {
		return e.newError(node, "invalid property info type"), true
	}

	indexVals := make([]Value, len(indices))
	for i, indexExpr := range indices {
		indexVals[i] = e.Eval(indexExpr, ctx)
		if isError(indexVals[i]) {
			return indexVals[i], true
		}
	}
	if errVal := e.checkIndexedPropertyArity(pInfo, len(indexVals), node); errVal != nil {
		return errVal, true
	}

	switch pInfo.ReadKind {
	case types.PropAccessField, types.PropAccessMethod:
		method := classInfo.LookupClassMethod(pInfo.ReadSpec)
		if method == nil {
			return e.newError(node, "indexed property '%s' getter '%s' is not a class method, so it cannot be read through class '%s'",
				pInfo.Name, pInfo.ReadSpec, classMetaVal.GetClassName()), true
		}
		if errVal := e.checkIndexedAccessorArity(pInfo, method, pInfo.ReadSpec, len(indexVals), "getter", node); errVal != nil {
			return errVal, true
		}
		return e.executeIndexedPropertyClassMethod(obj, method, indexVals, pInfo, node, ctx), true

	case types.PropAccessExpression:
		return e.executeIndexedPropertyExpressionRead(obj, pInfo, indexVals, node, ctx), true

	default:
		return e.newError(node, "indexed property '%s' has no read access", pInfo.Name), true
	}
}
