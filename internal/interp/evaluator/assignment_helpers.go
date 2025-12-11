package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Assignment Helpers
// ============================================================================

// ReferenceValueAccessor allows dereferencing and assignment to var parameters.
type ReferenceValueAccessor interface {
	Dereference() (Value, error)
	Assign(value Value) error
}

// SubrangeValueAccessor validates and assigns subrange values.
type SubrangeValueAccessor interface {
	ValidateAndSet(intValue int) error
	GetValue() int
	GetTypeName() string
}

// cloneIfCopyable returns a defensive copy for static arrays (value semantics).
// Dynamic arrays keep reference semantics.
func cloneIfCopyable(val Value) Value {
	if val == nil {
		return nil
	}

	// Dynamic arrays keep reference semantics
	if arr, ok := val.(*runtime.ArrayValue); ok {
		if arr.ArrayType == nil || arr.ArrayType.IsDynamic() {
			return val
		}
	}

	// Clone copyable values (static arrays)
	if copyable, ok := val.(runtime.CopyableValue); ok {
		if copied := copyable.Copy(); copied != nil {
			return copied.(Value)
		}
	}

	return val
}

// evalSimpleAssignmentDirect handles simple variable assignment: x := value
// Handles regular variables, var parameters, subranges, interface/object ref counting.
func (e *Evaluator) evalSimpleAssignmentDirect(
	target *ast.Identifier,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	targetName := target.Value

	existingValRaw, exists := ctx.Env().Get(targetName)
	if !exists {
		// Not in environment - check implicit Self context (fields/properties/class vars)
		if selfRaw, selfOk := ctx.Env().Get("Self"); selfOk {
			if selfVal, ok := selfRaw.(Value); ok && selfVal.Type() == "OBJECT" {
				if objVal, ok := selfVal.(ObjectValue); ok {
					// Try instance field
					if fieldValue := objVal.GetField(targetName); fieldValue != nil {
						if objInst, ok := selfVal.(*runtime.ObjectInstance); ok {
							objInst.SetField(targetName, value)
							return value
						}
						return e.newError(target, "cannot assign to field '%s': invalid object type", targetName)
					}

					// Try class variable
					if _, found := objVal.GetClassVar(targetName); found {
						if objInst, ok := selfVal.(*runtime.ObjectInstance); ok {
							classInfo := objInst.Class
							if classMetaVal, ok := classInfo.(ClassMetaValue); ok {
								if classMetaVal.SetClassVar(targetName, value) {
									return value
								}
								return e.newError(target, "failed to set class variable '%s'", targetName)
							}
						}
						return e.newError(target, "cannot assign to class variable '%s': class does not support SetClassVar", targetName)
					}

					// Try property
					if objVal.HasProperty(targetName) {
						return objVal.WriteProperty(targetName, value, func(propInfo any, val Value) Value {
							return e.executePropertyWrite(selfVal, propInfo, val, target, ctx)
						})
					}
				}
			}
		}

		// Check class method context (__CurrentClass__)
		if currentClassRaw, hasCurrentClass := ctx.Env().Get("__CurrentClass__"); hasCurrentClass {
			if classInfoVal, ok := currentClassRaw.(Value); ok && classInfoVal.Type() == "CLASSINFO" {
				if classMetaVal, ok := classInfoVal.(ClassMetaValue); ok {
					if _, found := classMetaVal.GetClassVar(targetName); found {
						if classMetaVal.SetClassVar(targetName, value) {
							return value
						}
						return e.newError(target, "failed to set class variable '%s'", targetName)
					}
				}
			}
		}

		return e.newError(target, "undefined variable '%s'", targetName)
	}

	existingVal, ok := existingValRaw.(Value)
	if !ok {
		return e.newError(target, "variable '%s' has invalid type (not a Value)", targetName)
	}

	// Var parameter (ReferenceValue)
	if refVal, isRef := existingVal.(ReferenceValueAccessor); isRef {
		return e.evalReferenceAssignment(refVal, value, target, stmt, ctx)
	}

	// External variable
	if existingVal.Type() == "EXTERNAL_VAR" {
		if extVar, ok := existingVal.(ExternalVarAccessor); ok {
			return e.newError(target, "unsupported external variable assignment: %s", extVar.ExternalVarName())
		}
		return e.newError(target, "unsupported external variable assignment")
	}

	// Subrange validation
	if subrangeVal, isSubrange := existingVal.(SubrangeValueAccessor); isSubrange {
		return e.evalSubrangeAssignment(subrangeVal, value, target)
	}

	// Interface variable - ref counting
	if ifaceInst, isIface := existingVal.(*runtime.InterfaceInstance); isIface {
		refMgr := ctx.RefCountManager()
		refMgr.ReleaseInterface(ifaceInst)

		// Wrap new value in interface
		if objInst, ok := value.(*runtime.ObjectInstance); ok {
			value = refMgr.WrapInInterface(ifaceInst.Interface, objInst)
		} else if srcIface, isSrcIface := value.(*runtime.InterfaceInstance); isSrcIface {
			value = refMgr.WrapInInterface(ifaceInst.Interface, srcIface.Object)
		} else if _, isNil := value.(*runtime.NilValue); isNil {
			value = &runtime.InterfaceInstance{
				Interface: ifaceInst.Interface,
				Object:    nil,
			}
		}

		e.SetVar(ctx, targetName, value)
		return value
	}

	// Object variable - ref counting
	if objInst, isObj := existingVal.(*runtime.ObjectInstance); isObj {
		refMgr := ctx.RefCountManager()

		if _, isNil := value.(*runtime.NilValue); isNil {
			refMgr.ReleaseObject(objInst)
		} else if newObj, isNewObj := value.(*runtime.ObjectInstance); isNewObj {
			if objInst != newObj {
				refMgr.ReleaseObject(objInst)
				refMgr.IncrementRef(newObj)
			}
		} else {
			refMgr.ReleaseObject(objInst)
		}

		e.SetVar(ctx, targetName, value)
		return value
	}

	// Implicit type conversion
	if value != nil {
		targetType := existingVal.Type()
		sourceType := value.Type()
		if targetType != sourceType {
			if converted, ok := e.TryImplicitConversion(value, targetType, ctx); ok {
				value = converted
			}
		}

		// Box value if target is Variant
		if targetType == "VARIANT" && sourceType != "VARIANT" {
			value = runtime.BoxVariant(value)
		}
	}

	// Clone copyable values (static arrays), except indexed expressions (keep reference for write-back)
	if stmt == nil {
		value = cloneIfCopyable(value)
	} else {
		if _, isIndexExpr := stmt.Value.(*ast.IndexExpression); !isIndexExpr {
			value = cloneIfCopyable(value)
		}
	}

	// Object value - increment ref count (interfaces handle wrapping separately)
	if newObj, isNewObj := value.(*runtime.ObjectInstance); isNewObj {
		if _, isIface := existingVal.(*runtime.InterfaceInstance); !isIface {
			refMgr := ctx.RefCountManager()
			refMgr.IncrementRef(newObj)
		}
	}

	// Method pointer - increment ref count for SelfObject
	if value != nil && value.Type() == "METHOD_POINTER" {
		refMgr := ctx.RefCountManager()
		if funcPtr, isFuncPtr := value.(*runtime.FunctionPointerValue); isFuncPtr {
			if funcPtr.SelfObject != nil {
				refMgr.IncrementRef(funcPtr.SelfObject)
			}
		}
	}

	// Update variable
	if e.SetVar(ctx, targetName, value) {
		return value
	}

	return e.newError(target, "undefined variable: %s", targetName)
}

// evalReferenceAssignment handles assignment through a var parameter.
func (e *Evaluator) evalReferenceAssignment(
	refVal ReferenceValueAccessor,
	value Value,
	target *ast.Identifier,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	currentVal, err := refVal.Dereference()
	if err != nil {
		return e.newError(target, "%s", err.Error())
	}

	// Interface/object var parameter - ref counting
	if currentVal.Type() == "INTERFACE" || currentVal.Type() == "OBJECT" {
		refMgr := ctx.RefCountManager()

		// Release old reference
		if oldIntf, isOldIntf := currentVal.(*runtime.InterfaceInstance); isOldIntf {
			refMgr.ReleaseInterface(oldIntf)
		} else if oldObj, isOldObj := currentVal.(*runtime.ObjectInstance); isOldObj {
			refMgr.ReleaseObject(oldObj)
		}

		// Increment new reference
		if value != nil {
			refMgr.IncrementRef(value)
		}

		if err := refVal.Assign(value); err != nil {
			return e.newError(target, "%s", err.Error())
		}

		return value
	}

	// Implicit type conversion
	targetType := currentVal.Type()
	sourceType := value.Type()
	if targetType != sourceType {
		if converted, ok := e.TryImplicitConversion(value, targetType, ctx); ok {
			value = converted
		}
	}

	// Box value if target is Variant
	if targetType == "VARIANT" && sourceType != "VARIANT" {
		value = runtime.BoxVariant(value)
	}

	// Clone copyable values
	value = cloneIfCopyable(value)

	if err := refVal.Assign(value); err != nil {
		return e.newError(target, "%s", err.Error())
	}

	return value
}

// evalSubrangeAssignment validates and assigns to a subrange variable.
func (e *Evaluator) evalSubrangeAssignment(
	subrangeVal SubrangeValueAccessor,
	value Value,
	target *ast.Identifier,
) Value {
	// Extract integer value from source
	var intValue int
	switch v := value.(type) {
	case *runtime.IntegerValue:
		intValue = int(v.Value)
	case SubrangeValueAccessor:
		intValue = v.GetValue()
	default:
		return e.newError(target, "cannot assign %s to subrange type %s", value.Type(), subrangeVal.GetTypeName())
	}

	// Validate the value is in range
	if err := subrangeVal.ValidateAndSet(intValue); err != nil {
		return e.newError(target, "%s", err.Error())
	}

	// Return the subrange value (modified in place)
	return value
}

// evalCompoundIdentifierAssignment handles compound assignment (+=, -=, *=, /=).
// Read current value, evaluate RHS, apply operation, write result back.
func (e *Evaluator) evalCompoundIdentifierAssignment(
	target *ast.Identifier,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	targetName := target.Value

	currentValRaw, exists := ctx.Env().Get(targetName)
	if !exists {
		// Not in environment - check implicit Self context
		if selfRaw, selfOk := ctx.Env().Get("Self"); selfOk {
			if selfVal, ok := selfRaw.(Value); ok && selfVal.Type() == "OBJECT" {
				if objVal, ok := selfVal.(ObjectValue); ok {
					// Try instance field
					if fieldValue := objVal.GetField(targetName); fieldValue != nil {
						rightVal := e.Eval(stmt.Value, ctx)
						if isError(rightVal) {
							return rightVal
						}

						// Apply compound operation
						result := e.applyCompoundOperation(stmt.Operator, fieldValue, rightVal, target)
						if isError(result) {
							return result
						}

						// Write back to field
						if objInst, ok := selfVal.(*runtime.ObjectInstance); ok {
							objInst.SetField(targetName, result)
							return result
						}
						return e.newError(target, "cannot assign to field '%s': invalid object type", targetName)
					}

					// Try class variable
					if classVarValue, found := objVal.GetClassVar(targetName); found {
						rightVal := e.Eval(stmt.Value, ctx)
						if isError(rightVal) {
							return rightVal
						}

						result := e.applyCompoundOperation(stmt.Operator, classVarValue, rightVal, target)
						if isError(result) {
							return result
						}

						if objInst, ok := selfVal.(*runtime.ObjectInstance); ok {
							classInfo := objInst.Class
							if classMetaVal, ok := classInfo.(ClassMetaValue); ok {
								if classMetaVal.SetClassVar(targetName, result) {
									return result
								}
								return e.newError(target, "failed to set class variable '%s'", targetName)
							}
						}
						return e.newError(target, "cannot assign to class variable '%s': class does not support SetClassVar", targetName)
					}

					// Try property (read-modify-write)
					if objVal.HasProperty(targetName) {
						currentPropValue := objVal.ReadProperty(targetName, func(propInfo any) Value {
							return e.executePropertyRead(selfVal, propInfo, target, ctx)
						})
						if isError(currentPropValue) {
							return currentPropValue
						}

						// Evaluate RHS
						rightVal := e.Eval(stmt.Value, ctx)
						if isError(rightVal) {
							return rightVal
						}

						// Apply compound operation
						result := e.applyCompoundOperation(stmt.Operator, currentPropValue, rightVal, target)
						if isError(result) {
							return result
						}

						// Write back to property
						return objVal.WriteProperty(targetName, result, func(propInfo any, val Value) Value {
							return e.executePropertyWrite(selfVal, propInfo, val, target, ctx)
						})
					}
				}
			}
		}

		// Check class method context (__CurrentClass__)
		if currentClassRaw, hasCurrentClass := ctx.Env().Get("__CurrentClass__"); hasCurrentClass {
			if classInfoVal, ok := currentClassRaw.(Value); ok && classInfoVal.Type() == "CLASSINFO" {
				if classMetaVal, ok := classInfoVal.(ClassMetaValue); ok {
					// Check for class variable
					if classVarValue, found := classMetaVal.GetClassVar(targetName); found {
						rightVal := e.Eval(stmt.Value, ctx)
						if isError(rightVal) {
							return rightVal
						}

						// Apply compound operation
						result := e.applyCompoundOperation(stmt.Operator, classVarValue, rightVal, target)
						if isError(result) {
							return result
						}

						// Write back to class variable
						if classMetaVal.SetClassVar(targetName, result) {
							return result
						}
						return e.newError(target, "failed to set class variable '%s'", targetName)
					}
				}
			}
		}

		return e.newError(target, "undefined variable '%s'", targetName)
	}

	currentVal, ok := currentValRaw.(Value)
	if !ok {
		return e.newError(target, "variable '%s' has invalid type (not a Value)", targetName)
	}

	// Var parameter - read-modify-write
	if refVal, isRef := currentVal.(ReferenceValueAccessor); isRef {
		// Dereference to get the actual value
		derefVal, err := refVal.Dereference()
		if err != nil {
			return e.newError(target, "%s", err.Error())
		}

		// Evaluate the RHS
		rightVal := e.Eval(stmt.Value, ctx)
		if isError(rightVal) {
			return rightVal
		}
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}

		// Apply the compound operation
		result := e.applyCompoundOperation(stmt.Operator, derefVal, rightVal, stmt)
		if isError(result) {
			return result
		}

		// Write back through the reference
		if err := refVal.Assign(result); err != nil {
			return e.newError(target, "%s", err.Error())
		}

		return result
	}

	// Regular variable
	rightVal := e.Eval(stmt.Value, ctx)
	if isError(rightVal) {
		return rightVal
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Apply the compound operation
	result := e.applyCompoundOperation(stmt.Operator, currentVal, rightVal, stmt)
	if isError(result) {
		return result
	}

	// Write the result back to the environment
	if e.SetVar(ctx, targetName, result) {
		return result
	}

	// Set failed - return error
	return e.newError(target, "undefined variable: %s", targetName)
}

// ============================================================================
// Context Inference Helpers
// ============================================================================
//
// These enable array and record literals to infer their types from the
// target variable during assignment.
// ============================================================================

// getArrayTypeFromTarget extracts ArrayType from target variable.
// This enables context inference for array literals during assignment:
// var arr: array of Integer; arr := [1, 2, 3]; // Literal adopts arr's type
//
// Returns nil if:
// - Target variable doesn't exist
// - Target variable is not an ArrayValue
// - Target variable has no type information
func (e *Evaluator) getArrayTypeFromTarget(target *ast.Identifier, ctx *ExecutionContext) *types.ArrayType {
	existingVal, exists := ctx.Env().Get(target.Value)
	if !exists {
		return nil
	}
	// runtime.ArrayValue has ArrayType directly accessible
	if arrVal, ok := existingVal.(*runtime.ArrayValue); ok {
		return arrVal.ArrayType
	}
	return nil
}

// getRecordTypeNameFromTarget extracts record type name from target variable.
// This enables context inference for anonymous record literals during assignment:
// var p: TPoint; p := (x: 1, y: 2); // Literal adopts TPoint type
//
// RecordValue.Type() returns the record type name (e.g., "TPoint") or "RECORD"
// for anonymous records. For context inference, we need the actual type name.
//
// Returns empty string if:
// - Target variable doesn't exist
// - Target variable is not a record
// - Target variable is an anonymous record (Type() returns "RECORD")
// - Type name is not a registered record type
func (e *Evaluator) getRecordTypeNameFromTarget(target *ast.Identifier, ctx *ExecutionContext) string {
	existingVal, exists := ctx.Env().Get(target.Value)
	if !exists {
		return ""
	}
	if existingVal == nil {
		return ""
	}
	// RecordValue.Type() returns record type name or "RECORD" for anonymous
	// We use the Value interface which is available
	if v, ok := existingVal.(Value); ok {
		typeName := v.Type()
		// Only return named record types, not generic "RECORD"
		if typeName != "" && typeName != "RECORD" && e.typeSystem.HasRecord(typeName) {
			return typeName
		}
	}
	return ""
}
