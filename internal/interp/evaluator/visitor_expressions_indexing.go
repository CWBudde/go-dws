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
func (e *Evaluator) VisitIndexExpression(node *ast.IndexExpression, ctx *ExecutionContext) (result Value) {
	if e.SemanticInfo() != nil && e.SemanticInfo().IsImplicitCall(node) {
		defer func() { result = e.finishImplicitCallableRead(result, node, ctx, true) }()
	}
	if node == nil {
		return e.newError(node, "nil index expression")
	}

	if node.Left == nil {
		return e.newError(node, "index expression missing base")
	}

	if record, prop, indices, handled, err := e.resolveRecordMetaIndexedProperty(node, ctx); handled {
		if err != nil {
			return err
		}
		return e.recordMetaPropertyRead(record, prop, indices, node, ctx)
	}

	if obj, prop, indices, handled, err := e.resolveInterfaceIndexedProperty(node, ctx); handled {
		if err != nil {
			return err
		}
		return e.readInterfaceIndexedProperty(obj, prop, indices, node, ctx)
	}
	// Collect indices - flatten for property access, not for regular arrays
	base, indices := CollectIndices(node)

	// Check named properties before normal array/string indexing.
	if memberAccess, ok := base.(*ast.MemberAccessExpression); ok && !e.interfacePropertyResultIndex(node, ctx) {
		if value, handled := e.readMemberIndexedProperty(memberAccess, indices, node, ctx); handled {
			return value
		}
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

// readMemberIndexedProperty preserves the receiver dispatch order of a named
// indexed property. An unhandled read falls back to ordinary member indexing.
func (e *Evaluator) readMemberIndexedProperty(member *ast.MemberAccessExpression, indices []ast.Expression, node *ast.IndexExpression, ctx *ExecutionContext) (Value, bool) {
	obj := e.Eval(member.Object, ctx)
	if isError(obj) {
		return obj, true
	}
	if record, ok := obj.(*runtime.RecordTypeValue); ok {
		if prop := recordMetaProperty(record, member.Member.Value); prop != nil && prop.IsIndexed {
			values, err := e.recordMetaPropertyIndices(indices, ctx)
			if err != nil {
				return err, true
			}
			return e.recordMetaPropertyRead(record, prop, values, node, ctx), true
		}
	}
	if instance, ok := obj.(InterfaceInstanceValue); ok {
		underlying, value, handled := e.readInterfaceMemberIndex(instance, member.Member.Value, indices, node, ctx)
		if handled {
			return value, true
		}
		obj = underlying
	}
	if runtime.KindOf(obj) == runtime.KindObject {
		if value, handled := e.readObjectMemberIndex(obj, member.Member.Value, indices, node, ctx); handled {
			return value, true
		}
	}
	if record, ok := obj.(RecordInstanceValue); ok {
		if value, handled := e.readRecordMemberIndex(record, member.Member.Value, indices, node, ctx); handled {
			return value, true
		}
	}
	// Ordinary properties with class accessors can also be read through a class.
	if classMeta, ok := obj.(ClassMetaValue); ok {
		return e.evalClassMetaIndexedProperty(obj, classMeta, member.Member.Value, indices, node, ctx)
	}
	return nil, false
}

// readInterfaceMemberIndex returns the unwrapped receiver when no interface
// indexed property matches, so subsequent object/record dispatch uses it.
func (e *Evaluator) readInterfaceMemberIndex(instance InterfaceInstanceValue, name string, indices []ast.Expression, node *ast.IndexExpression, ctx *ExecutionContext) (Value, Value, bool) {
	underlying := instance.GetUnderlyingObjectValue()
	if underlying == nil {
		return nil, e.newError(node, "interface is nil"), true
	}
	if accessor, ok := instance.(PropertyAccessor); ok {
		if prop := accessor.LookupProperty(name); prop != nil && prop.IsIndexed {
			indexVals := make([]Value, len(indices))
			for idx, expr := range indices {
				indexVals[idx] = e.Eval(expr, ctx)
				if isError(indexVals[idx]) {
					return nil, indexVals[idx], true
				}
			}
			if runtime.KindOf(underlying) == runtime.KindObject {
				if obj, ok := underlying.(ObjectValue); ok {
					value := obj.ReadIndexedProperty(prop.Impl, indexVals, func(pi any, idx []Value) Value {
						return e.executeIndexedPropertyRead(underlying, pi, idx, node, ctx)
					})
					return nil, value, true
				}
			}
			return nil, e.newError(node, "interface underlying object is not a class instance"), true
		}
	}
	return underlying, nil, false
}

func (e *Evaluator) readObjectMemberIndex(obj Value, name string, indices []ast.Expression, node *ast.IndexExpression, ctx *ExecutionContext) (Value, bool) {
	if accessor, ok := obj.(PropertyAccessor); ok {
		if prop := accessor.LookupProperty(name); prop != nil && prop.IsIndexed {
			indexVals := make([]Value, len(indices))
			for idx, expr := range indices {
				indexVals[idx] = e.Eval(expr, ctx)
				if isError(indexVals[idx]) {
					return indexVals[idx], true
				}
			}
			if object, ok := obj.(ObjectValue); ok {
				return object.ReadIndexedProperty(prop.Impl, indexVals, func(pi any, idx []Value) Value {
					return e.executeIndexedPropertyRead(obj, pi, idx, node, ctx)
				}), true
			}
		}
	}
	return nil, false
}

func (e *Evaluator) readRecordMemberIndex(record RecordInstanceValue, name string, indices []ast.Expression, node *ast.IndexExpression, ctx *ExecutionContext) (Value, bool) {
	if accessor, ok := record.(PropertyAccessor); ok {
		if prop := accessor.LookupProperty(name); prop != nil {
			indexVals := make([]Value, len(indices))
			for idx, expr := range indices {
				indexVals[idx] = e.Eval(expr, ctx)
				if isError(indexVals[idx]) {
					return indexVals[idx], true
				}
			}
			return record.ReadIndexedProperty(prop.Impl, indexVals, func(pi any, idx []Value) Value {
				return e.executeRecordIndexedPropertyRead(record, pi, idx, node, ctx)
			}), true
		}
	}
	return nil, false
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
	if ctx.Exception() != nil {
		return e.nilValue()
	}

	return e.readResolvedIndex(leftVal, indexVal, node, ctx)
}

// readResolvedIndex reads from an already captured container and index.
func (e *Evaluator) readResolvedIndex(leftVal, indexVal Value, node *ast.IndexExpression, ctx *ExecutionContext) Value {
	// Unwrap variants for indexing
	leftVal = unwrapVariant(leftVal)

	if record, ok := leftVal.(*runtime.RecordTypeValue); ok {
		if prop := recordMetaProperty(record, ""); prop != nil {
			return e.recordMetaPropertyRead(record, prop, []Value{indexVal}, node, ctx)
		}
	}

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
	if _, ok := unwrapPropertyInfo(propDesc.Impl); !ok {
		return e.newError(node, "invalid property info type"), true
	}
	indexVals := make([]Value, len(indices))
	for i, indexExpr := range indices {
		indexVals[i] = e.Eval(indexExpr, ctx)
		if isError(indexVals[i]) {
			return indexVals[i], true
		}
	}
	return e.evalClassMetaIndexedPropertyValues(obj, classMetaVal, memberName, indexVals, node, ctx)
}

// evalClassMetaIndexedPropertyValues reads a class indexed property after its
// index arguments have been captured by the caller.
func (e *Evaluator) evalClassMetaIndexedPropertyValues(
	obj Value,
	classMetaVal ClassMetaValue,
	memberName string,
	indexVals []Value,
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
