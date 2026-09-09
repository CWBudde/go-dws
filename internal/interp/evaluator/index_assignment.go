package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Index Assignment Operations
// ============================================================================
//
// Handles array and string index assignment: arr[i] := value, str[i] := char
//
// Array/string cases and supported indexed-property cases are handled directly
// in the evaluator.
// ============================================================================

// evalIndexAssignmentDirect handles array/string index assignment directly.
//
// Handles:
// - Array element assignment: arr[i] := value (with bounds checking)
// - String character mutation: str[i] := 'c' (1-based, rune-aware)
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
	// or an indexed array/string field: obj.Field[i] := value.
	if memberAccess, ok := base.(*ast.MemberAccessExpression); ok {
		baseObj := e.Eval(memberAccess.Object, ctx)
		if isError(baseObj) {
			return baseObj
		}
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}
		if accessor, ok := baseObj.(runtime.PropertyAccessor); ok {
			if propDesc := accessor.LookupProperty(memberAccess.Member.Value); propDesc != nil && propDesc.IsIndexed {
				return e.evalIndexedPropertyAssignmentOnObject(baseObj, memberAccess.Member.Value, indices, value, stmt, ctx)
			}
		}

		// Indexed property written through a class name, the write counterpart of
		// evalClassMetaIndexedProperty.
		if classMetaVal, ok := baseObj.(ClassMetaValue); ok {
			if result, handled := e.evalClassMetaIndexedPropertyWrite(baseObj, classMetaVal, memberAccess.Member.Value, indices, value, stmt, ctx); handled {
				return result
			}
		}

		memberVal := e.Eval(memberAccess, ctx)
		if isError(memberVal) {
			return memberVal
		}
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}
		if refVal, isRef := memberVal.(ReferenceAccessor); isRef {
			deref, err := refVal.Dereference()
			if err != nil {
				return e.newError(stmt, "failed to dereference indexed member: %s", err.Error())
			}
			memberVal = deref
		}

		if len(indices) == 1 {
			indexVal := e.Eval(indices[0], ctx)
			if isError(indexVal) {
				return indexVal
			}
			if ctx.Exception() != nil {
				return &runtime.NilValue{}
			}
			if assoc, ok := memberVal.(*runtime.AssociativeArrayValue); ok {
				assoc.Set(unwrapVariant(indexVal), cloneIfCopyable(value))
				return value
			}
			index, ok := e.ExtractIndexWithVariantCast(indexVal, ctx)
			if !ok {
				if ctx.Exception() != nil {
					return &runtime.NilValue{}
				}
				return e.newError(stmt, "array index must be an ordinal, got %s", indexVal.Type())
			}
			if arrayValue, ok := memberVal.(*runtime.ArrayValue); ok {
				return e.evalArrayElementAssignment(arrayValue, index, value, stmt, ctx)
			}
			if strVal, ok := memberVal.(*runtime.StringValue); ok {
				return e.evalStringCharAssignment(strVal, index, value, stmt)
			}
		}

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

	// JSON index write: obj['key'] := value / arr[i] := value.
	if isJSONBoxed(arrayVal) {
		return e.assignJSONIndex(jsonValueOf(arrayVal), indexVal, value, stmt, ctx)
	}

	// Check for interface-based indexed properties or object with default indexed property
	// Both INTERFACE and OBJECT types may have default indexed properties
	if runtime.KindOf(arrayVal) == runtime.KindInterface || runtime.KindOf(arrayVal) == runtime.KindObject {
		// Handle default property assignment using PropertyAccessor interface
		// Pattern: Same as 3.2.11g but lookup default property instead of named property
		return e.evalDefaultPropertyAssignment(arrayVal, indexVal, value, stmt, ctx)
	}

	// Associative array write: a[key] := value. Inserts a new key or updates an
	// existing one; there is no bounds check. Element value semantics are
	// preserved by snapshotting record/static-array values.
	if assoc, ok := arrayVal.(*runtime.AssociativeArrayValue); ok {
		assoc.Set(unwrapVariant(indexVal), cloneIfCopyable(value))
		return value
	}

	// Extract integer index (Variant indexes are cast per DWScript rules)
	index, ok := e.ExtractIndexWithVariantCast(indexVal, ctx)
	if !ok {
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}
		return e.newError(stmt, "array index must be an ordinal, got %s", indexVal.Type())
	}

	// Handle array assignment
	if arrayValue, ok := arrayVal.(*runtime.ArrayValue); ok {
		return e.evalArrayElementAssignment(arrayValue, index, value, stmt, ctx)
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
	ctx *ExecutionContext,
) Value {
	if arrayValue.ArrayType == nil {
		return e.newError(stmt, "array has no type information")
	}

	arrayType := arrayValue.ArrayType

	// Point bounds diagnostics at the index expression's closing bracket.
	diagNode := ast.Node(stmt)
	if stmt != nil {
		if target, ok := stmt.Target.(*ast.IndexExpression); ok {
			diagNode = target
		}
	}

	var physicalIndex int
	if arrayType.IsStatic() {
		// Static array: check bounds and adjust for low bound
		lowBound := *arrayType.LowBound
		highBound := *arrayType.HighBound

		if index < lowBound {
			return e.raiseIndexBoundExceeded(diagNode, index, false, ctx)
		}
		if index > highBound {
			return e.raiseIndexBoundExceeded(diagNode, index, true, ctx)
		}

		physicalIndex = index - lowBound
	} else {
		// Dynamic array: zero-based indexing
		if index < 0 {
			return e.raiseIndexBoundExceeded(diagNode, index, false, ctx)
		}
		if index >= len(arrayValue.Elements) {
			return e.raiseIndexBoundExceeded(diagNode, index, true, ctx)
		}

		physicalIndex = index
	}

	// Check physical bounds (safety check)
	if physicalIndex < 0 || physicalIndex >= len(arrayValue.Elements) {
		return e.newError(stmt, "array index out of bounds: physical index %d, length %d", physicalIndex, len(arrayValue.Elements))
	}

	// Update the array element. Record/static-array elements have value
	// semantics, so store a snapshot instead of aliasing the source value.
	arrayValue.Elements[physicalIndex] = cloneIfCopyable(value)

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
// This follows the pattern from executeRecordPropertyWrite:
// 1. Evaluate the base object (obj.Prop) to get the property metadata
// 2. Extract the property setter method reference
// 3. Build argument list: [indices..., value]
// 4. Execute setter through evaluator-owned object method dispatch
//
// Supports multi-index properties: obj.Prop[x, y] := value → args = [x, y, value]
//
// Uses general-purpose method dispatch instead of property-specific callbacks.
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

	return e.evalIndexedPropertyAssignmentOnObject(baseObj, propName, indices, value, stmt, ctx)
}

func (e *Evaluator) evalIndexedPropertyAssignmentOnObject(
	baseObj Value,
	propName string,
	indices []ast.Expression,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
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

	// An expression-based setter, e.g. `property Arr[i: Integer]: Integer
	// write (F[i])`, has no setter method and leaves WriteSpec empty, so it must
	// be handled before the read-only check below.
	if result, handled := e.tryIndexedPropertyExpressionWrite(baseObj, propDesc, indexValues, value, stmt, ctx); handled {
		return result
	}

	// Check if property has write access
	if propDesc.WriteSpec == "" {
		return e.newError(stmt, readOnlyPropertyWriteMessage)
	}

	// Look up the setter method declaration
	// Handle interfaces - get the underlying object
	var objVal ObjectValue
	if ifaceInst, isIface := baseObj.(*runtime.InterfaceInstance); isIface {
		if ifaceInst.Object == nil {
			return e.newError(stmt, "property '%s' setter cannot be executed on nil interface", propName)
		}
		objVal = ifaceInst.Object
	} else {
		var ok bool
		objVal, ok = baseObj.(ObjectValue)
		if !ok {
			return e.newError(stmt, "property '%s' setter cannot be executed on non-object", propName)
		}
	}

	// The setter may be a class method, the write-side counterpart of
	// executeIndexedPropertyGetterMethod's fallback.
	isClassMethod := false
	methodDecl := objVal.GetMethodDecl(propDesc.WriteSpec)
	if methodDecl == nil {
		methodDecl = objVal.GetClassMethodDecl(propDesc.WriteSpec)
		isClassMethod = methodDecl != nil
	}
	if methodDecl == nil {
		return e.newError(stmt, "indexed property '%s' setter method '%s' not found", propName, propDesc.WriteSpec)
	}

	// Build argument list for setter: [indices..., value]
	args := make([]Value, 0, len(indexValues)+1)
	args = append(args, indexValues...)
	args = append(args, value)

	var result Value
	if isClassMethod {
		result = e.executeIndexedPropertySetterClassMethod(e.classSelfForInstance(objVal, baseObj), methodDecl, args, propName, stmt, ctx)
	} else {
		result = e.executeObjectMethodDirect(objVal, methodDecl, args, stmt, ctx)
	}

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
// 5. Execute setter through evaluator-owned object method dispatch
//
// INTERFACE handling: InterfaceInstance.GetDefaultProperty() delegates to underlying interface
// OBJECT handling: ObjectInstance.GetDefaultProperty() uses IClassInfo.GetDefaultProperty()
//
// Uses general-purpose method dispatch instead of property-specific callbacks.
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

	// Expression-based setters leave WriteSpec empty (see the named-property
	// path); handle them before the read-only check.
	if result, handled := e.tryIndexedPropertyExpressionWrite(obj, propDesc, []Value{indexVal}, value, stmt, ctx); handled {
		return result
	}

	// Check if property has write access
	if propDesc.WriteSpec == "" {
		return e.newError(stmt, readOnlyPropertyWriteMessage)
	}

	// Look up the setter method declaration
	// Handle interfaces - get the underlying object
	var objVal ObjectValue
	if ifaceInst, isIface := obj.(*runtime.InterfaceInstance); isIface {
		if ifaceInst.Object == nil {
			return e.newError(stmt, "default property setter cannot be executed on nil interface")
		}
		objVal = ifaceInst.Object
	} else {
		var ok bool
		objVal, ok = obj.(ObjectValue)
		if !ok {
			return e.newError(stmt, "default property setter cannot be executed on non-object")
		}
	}

	isClassMethod := false
	methodDecl := objVal.GetMethodDecl(propDesc.WriteSpec)
	if methodDecl == nil {
		methodDecl = objVal.GetClassMethodDecl(propDesc.WriteSpec)
		isClassMethod = methodDecl != nil
	}
	if methodDecl == nil {
		return e.newError(stmt, "default property setter method '%s' not found", propDesc.WriteSpec)
	}

	// Build argument list for setter: [index, value]
	// Note: For default properties, we have a single index (not multi-index like named properties)
	args := []Value{indexVal, value}

	var result Value
	if isClassMethod {
		result = e.executeIndexedPropertySetterClassMethod(e.classSelfForInstance(objVal, obj), methodDecl, args, propDesc.Name, stmt, ctx)
	} else {
		result = e.executeObjectMethodDirect(objVal, methodDecl, args, stmt, ctx)
	}

	// Check for errors from method execution
	if isError(result) {
		return result
	}

	return value
}

// tryIndexedPropertyExpressionWrite executes an expression-based setter of an
// indexed property, e.g. `property Arr[i: Integer]: Integer write (F[i])`, which
// the parser normalizes to the statement `F[i] := Value`. The index parameters
// and the implicit `Value` are bound by name, mirroring the read side in
// executeIndexedPropertyExpressionRead. Reports handled=false when the property
// does not use an expression setter, leaving method dispatch to the caller.
func (e *Evaluator) tryIndexedPropertyExpressionWrite(
	obj Value,
	propDesc *runtime.PropertyDescriptor,
	indexValues []Value,
	value Value,
	stmt ast.Node,
	ctx *ExecutionContext,
) (Value, bool) {
	pInfo, ok := unwrapPropertyInfo(propDesc.Impl)
	if !ok || pInfo.WriteKind != types.PropAccessExpression {
		return nil, false
	}
	return e.executeIndexedPropertyExpressionWrite(obj, pInfo, indexValues, value, stmt, ctx)
}

// executeIndexedPropertyExpressionWrite runs the normalized `F[i] := Value`
// statement of an expression-based indexed setter against the given receiver,
// which is an instance for an instance property and the class meta value for a
// class property.
func (e *Evaluator) executeIndexedPropertyExpressionWrite(
	obj Value,
	pInfo *types.PropertyInfo,
	indexValues []Value,
	value Value,
	stmt ast.Node,
	ctx *ExecutionContext,
) (Value, bool) {
	writeStmt, ok := pInfo.WriteExpr.(ast.Statement)
	if !ok {
		return e.newError(stmt, "property '%s' has invalid write statement type", pInfo.Name), true
	}

	target := obj
	if ifaceInst, isIface := obj.(*runtime.InterfaceInstance); isIface {
		if ifaceInst.Object == nil {
			return e.newError(stmt, "property '%s' setter cannot be executed on nil interface", pInfo.Name), true
		}
		target = ifaceInst.Object
	}

	if len(pInfo.IndexParamNames) != len(indexValues) {
		return e.newError(stmt, "indexed property '%s' expects %d index argument(s), got %d",
			pInfo.Name, len(pInfo.IndexParamNames), len(indexValues)), true
	}

	ctx.PushEnv()
	defer ctx.PopEnv()

	if errVal := e.bindIndexedPropertyWriteScope(target, pInfo.IndexParamNames, indexValues, ctx); errVal != nil {
		return errVal, true
	}
	e.DefineVar(ctx, "Value", value)

	if result := e.Eval(writeStmt, ctx); isError(result) {
		return result, true
	}
	return value, true
}

// executeIndexedPropertySetterClassMethod invokes an indexed property setter that is
// a class method, binding the metaclass as the receiver.
func (e *Evaluator) executeIndexedPropertySetterClassMethod(
	classSelf Value,
	methodDecl *runtime.MethodMetadata,
	args []Value,
	propName string,
	stmt ast.Node,
	ctx *ExecutionContext,
) Value {
	classMeta, ok := classSelf.(ClassMetaValue)
	if !ok {
		return e.newError(stmt, "indexed property '%s' setter '%s' requires a class receiver", propName, methodDecl.Name)
	}
	return e.executeClassMethodDirect(classMeta, methodDecl, args, stmt, ctx)
}

// evalClassMetaIndexedPropertyWrite writes an indexed property through a class name,
// e.g. `TConvert.Prop[i] := v`. It reports handled=false when the class declares no
// such indexed property, so the caller falls through to ordinary handling.
func (e *Evaluator) evalClassMetaIndexedPropertyWrite(
	obj Value,
	classMetaVal ClassMetaValue,
	memberName string,
	indices []ast.Expression,
	value Value,
	stmt ast.Node,
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
		return e.newError(stmt, "invalid property info type"), true
	}

	indexValues := make([]Value, len(indices))
	for i, indexExpr := range indices {
		indexValues[i] = e.Eval(indexExpr, ctx)
		if isError(indexValues[i]) {
			return indexValues[i], true
		}
	}
	if errVal := e.checkIndexedPropertyArity(pInfo, len(indexValues), stmt); errVal != nil {
		return errVal, true
	}

	switch pInfo.WriteKind {
	case types.PropAccessField, types.PropAccessMethod:
		method := classInfo.LookupClassMethod(pInfo.WriteSpec)
		if method == nil {
			return e.newError(stmt, "indexed property '%s' setter '%s' is not a class method, so it cannot be written through class '%s'",
				pInfo.Name, pInfo.WriteSpec, classMetaVal.GetClassName()), true
		}
		args := make([]Value, 0, len(indexValues)+1)
		args = append(args, indexValues...)
		args = append(args, value)
		if errVal := e.checkIndexedAccessorArity(pInfo, method, pInfo.WriteSpec, len(args), "setter", stmt); errVal != nil {
			return errVal, true
		}
		if result := e.executeIndexedPropertyClassMethod(obj, method, args, pInfo, stmt, ctx); isError(result) {
			return result, true
		}
		return value, true

	case types.PropAccessExpression:
		// Mirrors the read side, which evaluates an expression accessor in class
		// context: an expression setter needs no instance either.
		return e.executeIndexedPropertyExpressionWrite(obj, pInfo, indexValues, value, stmt, ctx)

	default:
		return e.newError(stmt, readOnlyPropertyWriteMessage), true
	}
}
