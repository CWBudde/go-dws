package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/astutil"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// cloneIfCopyable returns a defensive copy for copyable values (e.g., static arrays).
// Dynamic arrays keep reference semantics per DWScript behavior.
func cloneIfCopyable(val Value) Value {
	if val == nil {
		return nil
	}
	if arr, ok := val.(*ArrayValue); ok {
		if arr.ArrayType == nil || arr.ArrayType.IsDynamic() {
			return val
		}
	}
	if copyable, ok := val.(CopyableValue); ok {
		if copied := copyable.Copy(); copied != nil {
			return copied
		}
	}
	return val
}

// evalAssignmentStatement handles simple and compound assignments.
// Supports: x := value, obj.field := value, arr[i] := value, x += value, etc.
func (i *Interpreter) evalAssignmentStatement(stmt *ast.AssignmentStatement) Value {
	isCompound := stmt.Operator != lexer.ASSIGN && stmt.Operator != lexer.TokenType(0)
	var value Value

	if isCompound {
		// Compound assignment: read current, evaluate RHS, apply operation
		currentValue := i.Eval(stmt.Target)
		if isError(currentValue) {
			return currentValue
		}
		rhsValue := i.Eval(stmt.Value)
		if isError(rhsValue) {
			return rhsValue
		}
		if i.exception != nil {
			return &NilValue{}
		}
		value = i.applyCompoundOperation(stmt.Operator, currentValue, rhsValue, stmt)
		if isError(value) {
			return value
		}
	} else {
		handledLiteral := false
		// Regular assignment - evaluate RHS with potential type context
		if arrayLit, ok := stmt.Value.(*ast.ArrayLiteralExpression); ok {
			handledLiteral = true
			var expected *types.ArrayType
			if targetIdent, ok := stmt.Target.(*ast.Identifier); ok {
				if existingVal, exists := i.Env().Get(targetIdent.Value); exists {
					if arrVal, ok := existingVal.(*ArrayValue); ok {
						expected = arrVal.ArrayType
					}
				}
			}
			value = i.evalArrayLiteralWithExpected(arrayLit, expected)
			if isError(value) {
				return value
			}
		} else if recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
			// Untyped record literal - infer type from target variable
			if targetIdent, ok := stmt.Target.(*ast.Identifier); ok {
				targetVar, exists := i.Env().Get(targetIdent.Value)
				if exists {
					if recVal, ok := targetVar.(*RecordValue); ok {
						recordLit.TypeName = &ast.Identifier{Value: recVal.RecordType.Name}
						value = i.Eval(recordLit)
						recordLit.TypeName = nil
					} else {
						value = i.Eval(stmt.Value)
					}
				} else {
					value = i.Eval(stmt.Value)
				}
			} else {
				value = i.Eval(stmt.Value)
			}
		} else {
			value = i.Eval(stmt.Value)
		}

		if isError(value) {
			return value
		}
		if i.exception != nil {
			return &NilValue{}
		}
		// Records have value semantics - copy when assigning
		if recordVal, ok := value.(*RecordValue); ok && !handledLiteral {
			value = recordVal.Copy()
		}
	}

	// Handle target types: identifier, member access, or index
	switch target := stmt.Target.(type) {
	case *ast.Identifier:
		return i.evalSimpleAssignment(target, value, stmt)
	case *ast.MemberAccessExpression:
		return i.evalMemberAssignment(target, value, stmt)
	case *ast.IndexExpression:
		return i.evalIndexAssignment(target, value, stmt)
	default:
		return i.newErrorWithLocation(stmt, "invalid assignment target type: %T", target)
	}
}

// applyCompoundOperation applies compound assignment operators (+=, -=, *=, /=).
func (i *Interpreter) applyCompoundOperation(op lexer.TokenType, left, right Value, stmt *ast.AssignmentStatement) Value {
	switch op {
	case lexer.PLUS_ASSIGN:
		// Check for class operator overrides first
		if objInst, ok := left.(*ObjectInstance); ok {
			result := i.tryCallClassOperator(objInst, "+=", []Value{right}, stmt)
			if result != nil {
				return result
			}
		}
		// += works with Integer, Float, String, Variant
		switch l := left.(type) {
		case *VariantValue:
			result := i.evalVariantBinaryOp("+", l, right, stmt)
			if isError(result) {
				return result
			}
			return result
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				return &IntegerValue{Value: l.Value + r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot add %s to Integer", right.Type())
		case *FloatValue:
			switch r := right.(type) {
			case *FloatValue:
				return &FloatValue{Value: l.Value + r.Value}
			case *IntegerValue:
				return &FloatValue{Value: l.Value + float64(r.Value)}
			default:
				return i.newErrorWithLocation(stmt, "type mismatch: cannot add %s to Float", right.Type())
			}
		case *StringValue:
			if r, ok := right.(*StringValue); ok {
				return &StringValue{Value: l.Value + r.Value}
			}
			// Handle Variant-to-String conversion
			if variantVal, ok := right.(*VariantValue); ok {
				innerVal, ok := unboxVariant(variantVal)
				if !ok {
					return i.newErrorWithLocation(stmt, "failed to unbox variant")
				}
				strVal := i.convertToString(innerVal)
				return &StringValue{Value: l.Value + strVal}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot add %s to String", right.Type())
		default:
			return i.newErrorWithLocation(stmt, "operator += not supported for type %s", left.Type())
		}

	case lexer.MINUS_ASSIGN:
		switch l := left.(type) {
		case *VariantValue:
			result := i.evalVariantBinaryOp("-", l, right, stmt)
			if isError(result) {
				return result
			}
			return result
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				return &IntegerValue{Value: l.Value - r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot subtract %s from Integer", right.Type())
		case *FloatValue:
			switch r := right.(type) {
			case *FloatValue:
				return &FloatValue{Value: l.Value - r.Value}
			case *IntegerValue:
				return &FloatValue{Value: l.Value - float64(r.Value)}
			default:
				return i.newErrorWithLocation(stmt, "type mismatch: cannot subtract %s from Float", right.Type())
			}
		default:
			return i.newErrorWithLocation(stmt, "operator -= not supported for type %s", left.Type())
		}

	case lexer.TIMES_ASSIGN:
		switch l := left.(type) {
		case *VariantValue:
			result := i.evalVariantBinaryOp("*", l, right, stmt)
			if isError(result) {
				return result
			}
			return result
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				return &IntegerValue{Value: l.Value * r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot multiply Integer by %s", right.Type())
		case *FloatValue:
			switch r := right.(type) {
			case *FloatValue:
				return &FloatValue{Value: l.Value * r.Value}
			case *IntegerValue:
				return &FloatValue{Value: l.Value * float64(r.Value)}
			default:
				return i.newErrorWithLocation(stmt, "type mismatch: cannot multiply Float by %s", right.Type())
			}
		default:
			return i.newErrorWithLocation(stmt, "operator *= not supported for type %s", left.Type())
		}

	case lexer.DIVIDE_ASSIGN:
		switch l := left.(type) {
		case *VariantValue:
			result := i.evalVariantBinaryOp("/", l, right, stmt)
			if isError(result) {
				return result
			}
			return result
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				if r.Value == 0 {
					return i.NewRuntimeError(stmt, "division_by_zero", "Division by zero",
						map[string]string{"left": fmt.Sprintf("%d", l.Value), "right": fmt.Sprintf("%d", r.Value)})
				}
				return &IntegerValue{Value: l.Value / r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot divide Integer by %s", right.Type())
		case *FloatValue:
			switch r := right.(type) {
			case *FloatValue:
				if r.Value == 0.0 {
					return i.NewRuntimeError(stmt, "division_by_zero", "Division by zero",
						map[string]string{"left": fmt.Sprintf("%v", l.Value), "right": fmt.Sprintf("%v", r.Value)})
				}
				return &FloatValue{Value: l.Value / r.Value}
			case *IntegerValue:
				if r.Value == 0 {
					return i.NewRuntimeError(stmt, "division_by_zero", "Division by zero",
						map[string]string{"left": fmt.Sprintf("%v", l.Value), "right": fmt.Sprintf("%d", r.Value)})
				}
				return &FloatValue{Value: l.Value / float64(r.Value)}
			default:
				return i.newErrorWithLocation(stmt, "type mismatch: cannot divide Float by %s", right.Type())
			}
		default:
			return i.newErrorWithLocation(stmt, "operator /= not supported for type %s", left.Type())
		}

	default:
		return i.newErrorWithLocation(stmt, "unknown compound operator: %v", op)
	}
}

// evalSimpleAssignment handles simple variable assignment: x := value
func (i *Interpreter) evalSimpleAssignment(target *ast.Identifier, value Value, stmt *ast.AssignmentStatement) Value {
	// Check if target is a var parameter (ReferenceValue)
	if existingVal, ok := i.Env().Get(target.Value); ok {
		if refVal, isRef := existingVal.(*ReferenceValue); isRef {
			// Write through var parameter reference
			currentVal, err := refVal.Dereference()
			if err != nil {
				return &ErrorValue{Message: err.Error()}
			}
			// Try implicit conversion if types don't match
			targetType := currentVal.Type()
			sourceType := value.Type()
			if targetType != sourceType {
				if converted, ok := i.tryImplicitConversion(value, targetType); ok {
					value = converted
				}
			}
			if targetType == "VARIANT" && sourceType != "VARIANT" {
				value = BoxVariant(value)
			}
			value = cloneIfCopyable(value)

			// Handle interface ref counting through var parameters
			if oldIntf, isOldIntf := currentVal.(*InterfaceInstance); isOldIntf {
				i.ReleaseInterfaceReference(oldIntf)
			}
			if intfInst, isIntf := value.(*InterfaceInstance); isIntf {
				if intfInst.Object != nil {
					i.evaluatorInstance.RefCountManager().IncrementRef(intfInst.Object)
				}
			}

			if err := refVal.Assign(value); err != nil {
				return &ErrorValue{Message: err.Error()}
			}
			return value
		}

		if extVar, isExternal := existingVal.(*ExternalVarValue); isExternal {
			return newError("Unsupported external variable assignment: %s", extVar.Name)
		}

		// Handle subrange variable assignment
		if subrangeVal, isSubrange := existingVal.(*SubrangeValue); isSubrange {
			var intValue int
			if intVal, ok := value.(*IntegerValue); ok {
				intValue = int(intVal.Value)
			} else if srcSubrange, ok := value.(*SubrangeValue); ok {
				intValue = srcSubrange.Value
			} else {
				return newError("cannot assign %s to subrange type %s", value.Type(), subrangeVal.SubrangeType.Name)
			}
			if err := subrangeVal.ValidateAndSet(intValue); err != nil {
				return &ErrorValue{Message: err.Error()}
			}
			return subrangeVal
		}

		// Try implicit conversion if types don't match
		if value != nil {
			targetType := existingVal.Type()
			sourceType := value.Type()
			if targetType != sourceType {
				if converted, ok := i.tryImplicitConversion(value, targetType); ok {
					value = converted
				}
			}
			if targetType == "VARIANT" && sourceType != "VARIANT" {
				value = BoxVariant(value)
			}
		}

		// Clone copyable values unless assigning from indexed expression (keeps reference)
		if stmt == nil {
			value = cloneIfCopyable(value)
		} else {
			if _, isIndexExpr := stmt.Value.(*ast.IndexExpression); !isIndexExpr {
				value = cloneIfCopyable(value)
			}
		}

		// Handle object ref counting when variable holds an object
		if objInst, isObj := existingVal.(*ObjectInstance); isObj {
			if _, isNil := value.(*NilValue); isNil {
				i.callDestructorIfNeeded(objInst)
			} else if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
				if objInst != newObj {
					i.callDestructorIfNeeded(objInst)
					i.evaluatorInstance.RefCountManager().IncrementRef(newObj)
				}
			}
		} else {
			// Increment ref count for new object (unless target is interface)
			if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
				if _, isIface := existingVal.(*InterfaceInstance); !isIface {
					i.evaluatorInstance.RefCountManager().IncrementRef(newObj)
				}
			}
		}

		// Wrap objects in InterfaceInstance when assigning to interface variables
		if ifaceInst, isIface := existingVal.(*InterfaceInstance); isIface {
			i.ReleaseInterfaceReference(ifaceInst)

			if objInst, ok := value.(*ObjectInstance); ok {
				value = NewInterfaceInstance(ifaceInst.Interface, objInst)
			} else if _, isNil := value.(*NilValue); isNil {
				value = &InterfaceInstance{Interface: ifaceInst.Interface, Object: nil}
			} else if srcIface, isSrcIface := value.(*InterfaceInstance); isSrcIface {
				// Copy semantics: both variables hold references
				if srcIface.Object != nil {
					i.evaluatorInstance.RefCountManager().IncrementRef(srcIface.Object)
				}
				value = &InterfaceInstance{Interface: ifaceInst.Interface, Object: srcIface.Object}
				if shouldReleaseInterfaceSource(stmt, i.Env()) {
					defer i.ReleaseInterfaceReference(srcIface)
				}
			}
		}
	}

	// Increment ref count for function pointers holding object references
	if funcPtr, isFuncPtr := value.(*FunctionPointerValue); isFuncPtr {
		if objInst, isObj := funcPtr.SelfObject.(*ObjectInstance); isObj {
			i.evaluatorInstance.RefCountManager().IncrementRef(objInst)
		}
	}

	// Try to set in current environment
	err := i.Env().Set(target.Value, value)
	if err == nil {
		return value
	}

	// Check method context for implicit Self field/class variable access
	selfVal, selfOk := i.Env().Get("Self")
	if selfOk {
		if obj, ok := AsObject(selfVal); ok {
			normalizedName := ident.Normalize(target.Value)

			// Check field in inheritance hierarchy
			if runtime.LookupFieldInHierarchy(obj.Class.GetMetadata(), normalizedName) != nil {
				obj.SetField(target.Value, value)
				return value
			}
			// Fallback: AST-based FieldsMap
			fields := obj.Class.GetFieldsMap()
			if fields != nil && fields[target.Value] != nil {
				obj.SetField(target.Value, value)
				return value
			}
			// Check class variables
			if _, ownerClass := obj.Class.LookupClassVar(target.Value); ownerClass != nil {
				concreteOwner, ok := ownerClass.(*ClassInfo)
				if ok {
					concreteOwner.ClassVars[target.Value] = value
					return value
				}
			}
			// Check properties (can be assigned without Self.)
			if propDesc := obj.Class.LookupProperty(target.Value); propDesc != nil {
				propInfo, ok := propDesc.Impl.(*types.PropertyInfo)
				if !ok {
					return i.newErrorWithLocation(target, "invalid property descriptor")
				}
				// Field-backed properties: write field directly
				if propInfo.WriteKind == types.PropAccessField {
					if fields != nil && fields[propInfo.WriteSpec] != nil {
						obj.SetField(propInfo.WriteSpec, value)
						return value
					}
				}
				return i.evalPropertyWrite(obj, propInfo, value, target)
			}
		}
	}

	// Check class method context
	currentClassVal, hasCurrentClass := i.Env().Get("__CurrentClass__")
	if hasCurrentClass {
		if classInfo, ok := currentClassVal.(*ClassInfoValue); ok {
			if _, exists := classInfo.ClassInfo.ClassVars[target.Value]; exists {
				classInfo.ClassInfo.ClassVars[target.Value] = value
				return value
			}
		}
	}

	return newError("undefined variable: %s", target.Value)
}

// evalRecordPropertyWrite handles property assignment for record values.
func (i *Interpreter) evalRecordPropertyWrite(recordVal *RecordValue, fieldName string, value Value, stmt *ast.AssignmentStatement, target *ast.MemberAccessExpression) Value {
	fieldNameNorm := ident.Normalize(fieldName)

	// Properties take precedence over fields
	if propInfo, exists := recordVal.RecordType.Properties[fieldNameNorm]; exists {
		if propInfo.WriteField != "" {
			// Try as setter method first
			if setterMethod := GetRecordMethod(recordVal, propInfo.WriteField); setterMethod != nil {
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: stmt.Token,
						},
					},
					Object: target.Object,
					Method: &ast.Identifier{
						Value: propInfo.WriteField,
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{Token: stmt.Token},
						},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{
							Value: "__temp_write_value__",
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{Token: stmt.Token},
							},
						},
					},
				}
				// Temporarily bind the value for the method call
				i.Env().Define("__temp_write_value__", value)
				result := i.evalMethodCall(methodCall)
				if isError(result) {
					return result
				}
				return value
			}
			// Not a method - direct field assignment
			recordVal.Fields[ident.Normalize(propInfo.WriteField)] = value
			return value
		}
		return i.newErrorWithLocation(stmt, "property '%s' is read-only", fieldName)
	}

	// Direct field assignment
	if _, exists := recordVal.RecordType.Fields[fieldNameNorm]; !exists {
		return i.newErrorWithLocation(stmt, "field '%s' not found in record '%s'", fieldName, recordVal.RecordType.Name)
	}
	recordVal.Fields[fieldNameNorm] = value
	return value
}

// isTemporaryInterfaceSource returns true for RHS expressions producing temporary interface values.
func isTemporaryInterfaceSource(stmt *ast.AssignmentStatement) bool {
	if stmt == nil {
		return false
	}
	switch stmt.Value.(type) {
	case *ast.Identifier, *ast.MemberAccessExpression, *ast.IndexExpression:
		return false
	default:
		return true
	}
}

// shouldReleaseInterfaceSource determines if RHS is a temporary interface value.
func shouldReleaseInterfaceSource(stmt *ast.AssignmentStatement, env *Environment) bool {
	if isTemporaryInterfaceSource(stmt) {
		return true
	}
	ident, ok := stmt.Value.(*ast.Identifier)
	if !ok {
		return false
	}
	if envVal, exists := env.Get(ident.Value); exists {
		_, isIface := envVal.(*InterfaceInstance)
		return !isIface
	}
	return true
}

// evalMemberAssignment handles member assignment: obj.field := value or TClass.Variable := value
func (i *Interpreter) evalMemberAssignment(target *ast.MemberAccessExpression, value Value, stmt *ast.AssignmentStatement) Value {
	// Check for static class member assignment (TClass.Variable or TClass.Property)
	if targetIdent, ok := target.Object.(*ast.Identifier); ok {
		var classInfo *ClassInfo
		for className, class := range i.classes {
			if ident.Equal(className, targetIdent.Value) {
				classInfo = class
				break
			}
		}
		if classInfo != nil {
			memberName := target.Member.Value
			// Class properties take precedence
			if propDesc := classInfo.LookupProperty(memberName); propDesc != nil {
				propInfo, ok := propDesc.Impl.(*types.PropertyInfo)
				if ok && propInfo.IsClassProperty {
					return i.evalClassPropertyWrite(classInfo, propInfo, value, stmt)
				}
			}
			// Class variable assignment
			if _, exists := classInfo.ClassVars[memberName]; !exists {
				return i.newErrorWithLocation(stmt, "class variable '%s' not found in class '%s'", memberName, targetIdent.Value)
			}
			classInfo.ClassVars[memberName] = value
			return value
		}

		// Static helper class variable assignment
		if typeMetaVal, exists := i.Env().Get(targetIdent.Value); exists {
			if tmv, ok := typeMetaVal.(*TypeMetaValue); ok {
				helpers := i.getHelpersForValue(tmv)
				varNameLower := ident.Normalize(target.Member.Value)
				for idx := len(helpers) - 1; idx >= 0; idx-- {
					if _, ok := helpers[idx].ClassVars[varNameLower]; ok {
						helpers[idx].ClassVars[varNameLower] = value
						return value
					}
				}
			}
		}
	}

	// Instance member access
	objVal := i.Eval(target.Object)
	if isError(objVal) {
		return objVal
	}
	if recordVal, ok := objVal.(*RecordValue); ok {
		return i.evalRecordPropertyWrite(recordVal, target.Member.Value, value, stmt, target)
	}

	// Auto-initialize uninitialized record array elements
	if _, isNil := objVal.(*NilValue); isNil {
		if indexExpr, ok := target.Object.(*ast.IndexExpression); ok {
			arrayVal := i.Eval(indexExpr.Left)
			if isError(arrayVal) {
				return arrayVal
			}
			if arrVal, ok := arrayVal.(*ArrayValue); ok {
				if arrVal.ArrayType != nil && arrVal.ArrayType.ElementType != nil {
					if recordType, ok := arrVal.ArrayType.ElementType.(*types.RecordType); ok {
						newRecord := &RecordValue{RecordType: recordType, Fields: make(map[string]Value)}
						assignStmt := &ast.AssignmentStatement{
							BaseNode: ast.BaseNode{Token: stmt.Token},
							Target:   indexExpr,
							Value:    &ast.Identifier{Value: "__temp__"},
						}
						tempResult := i.evalIndexAssignment(indexExpr, newRecord, assignStmt)
						if isError(tempResult) {
							return tempResult
						}
						objVal = newRecord
					}
				}
			}
		}
	}

	// Re-check after potential initialization
	if recordVal, ok := objVal.(*RecordValue); ok {
		return i.evalRecordPropertyWrite(recordVal, target.Member.Value, value, stmt, target)
	}

	// Unwrap interface instances
	if intfInst, ok := objVal.(*InterfaceInstance); ok {
		if intfInst.Object == nil {
			return i.newErrorWithLocation(stmt, "Interface is nil")
		}
		objVal = intfInst.Object
		if propInfo := intfInst.Interface.GetProperty(target.Member.Value); propInfo != nil {
			if obj, ok := AsObject(objVal); ok {
				typesPropertyInfo, ok := propInfo.Impl.(*types.PropertyInfo)
				if !ok {
					return i.newErrorWithLocation(stmt, "invalid property info type")
				}
				return i.evalPropertyWrite(obj, typesPropertyInfo, value, stmt)
			}
			return i.newErrorWithLocation(stmt, "interface underlying object is not a class instance")
		}
	}

	// Unwrap type casts
	if typeCast, ok := objVal.(*TypeCastValue); ok {
		objVal = typeCast.Object
	}

	// Handle object instance
	obj, ok := AsObject(objVal)
	if !ok {
		// Try helper properties first
		helper, helperProp := i.findHelperProperty(objVal, target.Member.Value)
		if helperProp != nil && helperProp.WriteKind != types.PropAccessNone {
			return i.evalHelperPropertyWrite(helper, helperProp, objVal, value, stmt, target)
		}

		// Try helper class variables
		helpers := i.getHelpersForValue(objVal)
		if helpers != nil {
			varNameLower := ident.Normalize(target.Member.Value)
			for idx := len(helpers) - 1; idx >= 0; idx-- {
				if _, ok := helpers[idx].ClassVars[varNameLower]; ok {
					helpers[idx].ClassVars[varNameLower] = value
					return value
				}
			}
		}
		return i.newErrorWithLocation(stmt, "cannot assign to field of non-object type '%s'", objVal.Type())
	}

	memberName := target.Member.Value
	// Properties take precedence over fields
	if propDesc := obj.Class.LookupProperty(memberName); propDesc != nil {
		propInfo, ok := propDesc.Impl.(*types.PropertyInfo)
		if !ok {
			return i.newErrorWithLocation(stmt, "invalid property descriptor")
		}
		return i.evalPropertyWrite(obj, propInfo, value, stmt)
	}

	// Direct field assignment (walks inheritance chain)
	obj.SetField(memberName, value)
	return value
}

// evalIndexAssignment handles array index assignment: arr[i] := value
func (i *Interpreter) evalIndexAssignment(target *ast.IndexExpression, value Value, stmt *ast.AssignmentStatement) Value {
	// Check for multi-index property write (only for property access, not regular arrays)
	base, indices := astutil.CollectIndices(target)

	// Handle indexed property write: obj.Property[index1, index2, ...] := value
	if memberAccess, ok := base.(*ast.MemberAccessExpression); ok {
		objVal := i.Eval(memberAccess.Object)
		// Auto-create nil class properties
		if objVal != nil && objVal.Type() == "NIL" {
			if maObj, ok := memberAccess.Object.(*ast.MemberAccessExpression); ok {
				if ensured := i.ensureClassPropertyInstance(maObj); ensured != nil {
					objVal = ensured
				}
			}
		}
		if isError(objVal) {
			return objVal
		}

		// Interface-based indexed properties
		if intfInst, ok := objVal.(*InterfaceInstance); ok {
			if intfInst.Object == nil {
				return i.newErrorWithLocation(stmt, "Interface is nil")
			}
			objVal = intfInst.Object
			if propInfo := intfInst.Interface.GetProperty(memberAccess.Member.Value); propInfo != nil && propInfo.IsIndexed {
				indexVals := make([]Value, len(indices))
				for idx, indexExpr := range indices {
					indexVals[idx] = i.Eval(indexExpr)
					if isError(indexVals[idx]) {
						return indexVals[idx]
					}
				}
				if obj, ok := AsObject(objVal); ok {
					typesPropertyInfo, ok := propInfo.Impl.(*types.PropertyInfo)
					if !ok {
						return i.newErrorWithLocation(stmt, "invalid property info type")
					}
					return i.evalIndexedPropertyWrite(obj, typesPropertyInfo, indexVals, value, stmt)
				}
				return i.newErrorWithLocation(stmt, "interface underlying object is not a class instance")
			}
		}

		// Class instance indexed property
		if obj, ok := AsObject(objVal); ok {
			propDesc := obj.Class.LookupProperty(memberAccess.Member.Value)
			if propDesc != nil && propDesc.IsIndexed {
				propInfo, ok := propDesc.Impl.(*types.PropertyInfo)
				if !ok {
					return i.newErrorWithLocation(stmt, "invalid property descriptor")
				}
				indexVals := make([]Value, len(indices))
				for idx, indexExpr := range indices {
					indexVals[idx] = i.Eval(indexExpr)
					if isError(indexVals[idx]) {
						return indexVals[idx]
					}
				}
				return i.evalIndexedPropertyWrite(obj, propInfo, indexVals, value, stmt)
			}
		}
	}

	// Regular array indexing - process outermost index only (allows FData[x][y] := value)
	arrayVal := i.Eval(target.Left)
	if isError(arrayVal) {
		return arrayVal
	}
	indexVal := i.Eval(target.Index)
	if isError(indexVal) {
		return indexVal
	}

	// Auto-instantiate nil class-typed properties for patterns like obj.Sub['x'] := value
	if arrayVal != nil && arrayVal.Type() == "NIL" {
		if memberAccess, ok := target.Left.(*ast.MemberAccessExpression); ok {
			ownerVal := i.Eval(memberAccess.Object)
			if !isError(ownerVal) {
				if ownerObj, ok := AsObject(ownerVal); ok {
					if propDesc := ownerObj.Class.LookupProperty(memberAccess.Member.Value); propDesc != nil {
						propInfo, ok := propDesc.Impl.(*types.PropertyInfo)
						if ok {
							if classType, ok := propInfo.Type.(*types.ClassType); ok {
								if classInfo := i.resolveClassInfoByName(classType.Name); classInfo != nil {
									if propInfo.ReadKind == types.PropAccessField && propInfo.ReadSpec != "" {
										newInst := NewObjectInstance(classInfo)
										ownerObj.SetField(propInfo.ReadSpec, newInst)
										arrayVal = newInst
									}
								}
							}
						}
					}
				} else if ownerVal != nil && ownerVal.Type() == "NIL" {
					// Handle nested nil properties
					if maObj, ok := memberAccess.Object.(*ast.MemberAccessExpression); ok {
						if ensured := i.ensureClassPropertyInstance(maObj); ensured != nil && ensured.Type() != "NIL" {
							ownerVal = ensured
							if ownerObj, ok := AsObject(ensured); ok {
								if propDesc := ownerObj.Class.LookupProperty(memberAccess.Member.Value); propDesc != nil {
									propInfo, ok := propDesc.Impl.(*types.PropertyInfo)
									if ok {
										if classType, ok := propInfo.Type.(*types.ClassType); ok {
											if classInfo := i.resolveClassInfoByName(classType.Name); classInfo != nil {
												if propInfo.ReadKind == types.PropAccessField && propInfo.ReadSpec != "" {
													newInst := NewObjectInstance(classInfo)
													ownerObj.SetField(propInfo.ReadSpec, newInst)
													arrayVal = newInst
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Handle interface default indexed properties (e.g., intf['x'] := y)
	if intfInst, ok := arrayVal.(*InterfaceInstance); ok {
		if intfInst.Object == nil {
			return i.newErrorWithLocation(stmt, "Interface is nil")
		}
		if propInfo := intfInst.Interface.GetDefaultProperty(); propInfo != nil && propInfo.IsIndexed {
			if obj, ok := AsObject(intfInst.Object); ok {
				typesPropertyInfo, ok := propInfo.Impl.(*types.PropertyInfo)
				if !ok {
					return i.newErrorWithLocation(stmt, "invalid property info type")
				}
				return i.evalIndexedPropertyWrite(obj, typesPropertyInfo, []Value{indexVal}, value, stmt)
			}
			return i.newErrorWithLocation(stmt, "interface underlying object is not a class instance")
		}
		arrayVal = intfInst.Object
	}

	// Handle object default properties (obj[index] := value -> obj.DefaultProperty[index] := value)
	if obj, ok := AsObject(arrayVal); ok {
		defaultPropDesc := obj.Class.GetDefaultProperty()
		if defaultPropDesc != nil {
			propInfo, ok := defaultPropDesc.Impl.(*types.PropertyInfo)
			if !ok {
				return i.newErrorWithLocation(stmt, "invalid default property descriptor")
			}
			return i.evalIndexedPropertyWrite(obj, propInfo, []Value{indexVal}, value, stmt)
		}
	}

	// Index must be ordinal type
	var index int
	switch idx := indexVal.(type) {
	case *IntegerValue:
		index = int(idx.Value)
	case *EnumValue:
		index = idx.OrdinalValue
	case *BooleanValue:
		if idx.Value {
			index = 1
		}
	case *SubrangeValue:
		index = idx.Value
	default:
		return i.newErrorWithLocation(stmt, "array index must be an ordinal, got %s", indexVal.Type())
	}

	// Handle array assignment
	arrayValue, ok := arrayVal.(*ArrayValue)
	if !ok {
		// String indexed assignment (1-based)
		if strVal, ok := arrayVal.(*StringValue); ok {
			strLen := runeLength(strVal.Value)
			if index < 1 || index > strLen {
				return i.newErrorWithLocation(stmt, "string index out of bounds: %d (string length is %d)", index, strLen)
			}
			charVal, ok := value.(*StringValue)
			if !ok {
				return i.newErrorWithLocation(stmt, "cannot assign %s to string index (expected STRING)", value.Type())
			}
			if runeLength(charVal.Value) == 0 {
				return i.newErrorWithLocation(stmt, "cannot assign empty string to string index")
			}
			r, _ := runeAt(charVal.Value, 1)
			if newStr, ok := runeReplace(strVal.Value, index, r); ok {
				strVal.Value = newStr
				return value
			}
			return i.newErrorWithLocation(stmt, "string index out of bounds: %d (string length is %d)", index, strLen)
		}
		return i.newErrorWithLocation(stmt, "cannot index type %s", arrayVal.Type())
	}

	// Bounds checking
	if arrayValue.ArrayType == nil {
		return i.newErrorWithLocation(stmt, "array has no type information")
	}
	arrayType := arrayValue.ArrayType

	var physicalIndex int
	if arrayType.IsStatic() {
		lowBound := *arrayType.LowBound
		highBound := *arrayType.HighBound
		if index < lowBound || index > highBound {
			return i.newErrorWithLocation(stmt, "array index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
		}
		physicalIndex = index - lowBound
	} else {
		if index < 0 || index >= len(arrayValue.Elements) {
			return i.newErrorWithLocation(stmt, "array index out of bounds: %d (array length is %d)", index, len(arrayValue.Elements))
		}
		physicalIndex = index
	}

	if physicalIndex < 0 || physicalIndex >= len(arrayValue.Elements) {
		return i.newErrorWithLocation(stmt, "array index out of bounds: physical index %d, length %d", physicalIndex, len(arrayValue.Elements))
	}

	arrayValue.Elements[physicalIndex] = value
	return value
}
