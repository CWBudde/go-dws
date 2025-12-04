package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Assignment Helpers
// ============================================================================
//
// This file contains helpers for evaluating assignment statements directly
// in the evaluator, reducing adapter dependency.
// ============================================================================

// ReferenceValueAccessor is an interface for values that can be dereferenced and assigned.
// This allows the evaluator to work with ReferenceValue without importing the interp package.
type ReferenceValueAccessor interface {
	// Dereference returns the current value of the referenced variable.
	Dereference() (Value, error)
	// Assign sets the value of the referenced variable.
	Assign(value Value) error
}

// SubrangeValueAccessor is an interface for subrange values.
type SubrangeValueAccessor interface {
	// ValidateAndSet validates that the value is in range and sets it.
	ValidateAndSet(intValue int) error
	// GetValue returns the current integer value.
	GetValue() int
	// GetTypeName returns the subrange type name.
	GetTypeName() string
}

// Note: ExternalVarAccessor is defined in evaluator.go

// cloneIfCopyable returns a defensive copy for values that implement CopyableValue.
// DWScript static arrays have value semantics, so assignments should duplicate
// their backing storage to avoid accidental aliasing between variables.
// Dynamic arrays keep reference semantics.
func cloneIfCopyable(val Value) Value {
	if val == nil {
		return nil
	}

	// Dynamic arrays should keep reference semantics (DWScript behavior).
	if arr, ok := val.(*runtime.ArrayValue); ok {
		if arr.ArrayType == nil || arr.ArrayType.IsDynamic() {
			return val
		}
	}

	if copyable, ok := val.(runtime.CopyableValue); ok {
		if copied := copyable.Copy(); copied != nil {
			// Copy() returns interface{}, cast to Value
			return copied.(Value)
		}
	}

	return val
}

// evalSimpleAssignmentDirect handles simple variable assignment: x := value
//
// This handles the simplest cases directly:
// - Regular variable assignment with matching types
// - Var parameter (ReferenceValue) write-through
// - Subrange value validation
// - Basic type compatibility (same type or simple conversion)
//
// For complex cases (interface wrapping, object ref counting, property assignment,
// Self/class context, etc.), it delegates to the adapter.
//
// Returns the assigned value on success, or an error Value.
func (e *Evaluator) evalSimpleAssignmentDirect(
	target *ast.Identifier,
	value Value,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	targetName := target.Value

	// Get existing value to check for special types
	existingValRaw, exists := ctx.Env().Get(targetName)
	if !exists {
		// The variable might be:
		// 1. An instance field (Self.Field) in a method
		// 2. A class variable (TClass.ClassVar)
		// The interpreter owns environment management and has access to Self/class context.
		// The evaluator cannot check Self/class scope without circular import.
		// KEEP: Architectural constraint - environment ownership
		return e.adapter.EvalNode(stmt)
	}

	// Cast to Value interface
	existingVal, ok := existingValRaw.(Value)
	if !ok {
		// Not a Value - this is an internal error
		return e.newError(target, "variable '%s' has invalid type (not a Value)", targetName)
	}

	// Check if target is a var parameter (ReferenceValue)
	if refVal, isRef := existingVal.(ReferenceValueAccessor); isRef {
		return e.evalReferenceAssignment(refVal, value, target, stmt, ctx)
	}

	// Check for external variable
	if existingVal.Type() == "EXTERNAL_VAR" {
		if extVar, ok := existingVal.(ExternalVarAccessor); ok {
			return e.newError(target, "unsupported external variable assignment: %s", extVar.ExternalVarName())
		}
		return e.newError(target, "unsupported external variable assignment")
	}

	// Check if assigning to a subrange variable
	if subrangeVal, isSubrange := existingVal.(SubrangeValueAccessor); isSubrange {
		return e.evalSubrangeAssignment(subrangeVal, value, target)
	}

	// Task 3.5.41a: Assigning to interface variable - native ref counting
	// Handle interface variable assignment with proper ref counting
	if ifaceInst, isIface := existingVal.(*runtime.InterfaceInstance); isIface {
		refMgr := ctx.RefCountManager()

		// Release old interface reference (decrements ref count, may invoke destructor)
		refMgr.ReleaseInterface(ifaceInst)

		// Wrap new value in interface (increments ref count automatically)
		if objInst, ok := value.(*runtime.ObjectInstance); ok {
			// Assigning object to interface - wrap it
			value = refMgr.WrapInInterface(ifaceInst.Interface, objInst)
		} else if srcIface, isSrcIface := value.(*runtime.InterfaceInstance); isSrcIface {
			// Interface-to-interface assignment - wrap underlying object
			value = refMgr.WrapInInterface(ifaceInst.Interface, srcIface.Object)
		} else if _, isNil := value.(*runtime.NilValue); isNil {
			// Assigning nil - create interface with nil object (no ref count needed)
			value = &runtime.InterfaceInstance{
				Interface: ifaceInst.Interface,
				Object:    nil,
			}
		}

		// Update variable with wrapped value
		e.SetVar(ctx, targetName, value)
		return value
	}

	// Task 3.5.41b: Assigning to object variable - native ref counting
	// Handle object variable assignment with proper ref counting
	if objInst, isObj := existingVal.(*runtime.ObjectInstance); isObj {
		refMgr := ctx.RefCountManager()

		if _, isNil := value.(*runtime.NilValue); isNil {
			// Setting to nil - release old object
			refMgr.ReleaseObject(objInst)
		} else if newObj, isNewObj := value.(*runtime.ObjectInstance); isNewObj {
			// Replacing with new object
			if objInst != newObj {
				// Different objects - release old, increment new
				refMgr.ReleaseObject(objInst)
				refMgr.IncrementRef(newObj)
			}
			// Same instance: no ref count change
		} else {
			// Replacing object with non-object - release old
			refMgr.ReleaseObject(objInst)
		}

		// Update variable
		e.SetVar(ctx, targetName, value)
		return value
	}

	// Try implicit conversion if types don't match
	if value != nil {
		targetType := existingVal.Type()
		sourceType := value.Type()
		if targetType != sourceType {
			if converted, ok := e.TryImplicitConversion(value, targetType, ctx); ok {
				value = converted
			}
		}

		// Box value if target is a Variant
		if targetType == "VARIANT" && sourceType != "VARIANT" {
			value = runtime.BoxVariant(value)
		}
	}

	// Ensure value semantics for types that support copying (e.g., static arrays)
	// Exception: when assigning directly from an indexed expression (e.g., row := matrix[i])
	// we keep the reference so mutations write back into the parent container.
	if stmt == nil {
		value = cloneIfCopyable(value)
	} else {
		if _, isIndexExpr := stmt.Value.(*ast.IndexExpression); !isIndexExpr {
			value = cloneIfCopyable(value)
		}
	}

	// Task 3.5.41c: Assigning object VALUE - native ref counting
	// When assigning an object VALUE, increment ref count for the new reference
	// Exception: interface variables handle wrapping separately (don't double-increment)
	if newObj, isNewObj := value.(*runtime.ObjectInstance); isNewObj {
		// Check if target is NOT an interface (interface wrapping increments separately)
		if _, isIface := existingVal.(*runtime.InterfaceInstance); !isIface {
			refMgr := ctx.RefCountManager()
			refMgr.IncrementRef(newObj)
		}
	}

	// Task 3.5.41d: Interface VALUE assignment is handled by line 127 migration
	// The WrapInInterface method automatically increments ref count
	// This delegation is redundant - removed in Task 3.5.41

	// Task 3.5.41e: Handle function pointer ref counting
	// Only method pointers (with SelfObject) need ref counting
	// Simple function pointers can be handled natively
	if value != nil {
		valueType := value.Type()
		if valueType == "METHOD_POINTER" {
			// Method pointer with SelfObject - increment ref count
			refMgr := ctx.RefCountManager()
			if funcPtr, isFuncPtr := value.(*runtime.FunctionPointerValue); isFuncPtr {
				if funcPtr.SelfObject != nil {
					refMgr.IncrementRef(funcPtr.SelfObject)
				}
			}
		}
		// FUNCTION_POINTER and LAMBDA can be handled natively (no ref counting needed)
	}

	// Simple case: update the variable in the environment
	if e.SetVar(ctx, targetName, value) {
		return value
	}

	// Set failed - return error
	return e.newError(target, "undefined variable: %s", targetName)
}

// evalReferenceAssignment handles assignment through a var parameter (ReferenceValue).
func (e *Evaluator) evalReferenceAssignment(
	refVal ReferenceValueAccessor,
	value Value,
	target *ast.Identifier,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	// Get current value to check type compatibility
	currentVal, err := refVal.Dereference()
	if err != nil {
		return e.newError(target, "%s", err.Error())
	}

	// Task 3.5.41f: Var parameter to interface/object - native ref counting
	// When assigning through a var parameter that references an interface/object:
	// - Release old interface/object reference
	// - Increment ref count for new reference
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

		// Write through the reference
		if err := refVal.Assign(value); err != nil {
			return e.newError(target, "%s", err.Error())
		}

		return value
	}

	// Try implicit conversion if types don't match
	targetType := currentVal.Type()
	sourceType := value.Type()
	if targetType != sourceType {
		if converted, ok := e.TryImplicitConversion(value, targetType, ctx); ok {
			value = converted
		}
	}

	// Box value if target is a Variant
	if targetType == "VARIANT" && sourceType != "VARIANT" {
		value = runtime.BoxVariant(value)
	}

	// Ensure value semantics for copyable types
	value = cloneIfCopyable(value)

	// Task 3.5.41f: Assigning object/interface VALUE through var parameter
	// This case is already handled by the first check above (line 261)
	// When the value is an object/interface, the IncrementRef call handles it
	// No additional delegation needed

	// Write through the reference
	if err := refVal.Assign(value); err != nil {
		return e.newError(target, "%s", err.Error())
	}

	return value
}

// evalSubrangeAssignment handles assignment to a subrange variable.
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
		// Assigning from another subrange - extract the value
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

// evalCompoundIdentifierAssignment handles compound assignment operators (+=, -=, *=, /=)
// for simple identifier targets.
//
// This handles the compound assignment flow:
// 1. Read current value from environment
// 2. Evaluate the right-hand side expression
// 3. Apply the compound operation (handled by applyCompoundOperation in compound_ops.go)
// 4. Write the result back to the environment
//
// For complex targets (var parameters, objects needing ref counting), delegates to adapter.
func (e *Evaluator) evalCompoundIdentifierAssignment(
	target *ast.Identifier,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	targetName := target.Value

	// Get current value from environment
	currentValRaw, exists := ctx.Env().Get(targetName)
	if !exists {
		// Task 3.5.36: Compound assignment to non-env variable - ESSENTIAL for Self/class context
		// The variable might be:
		// 1. An instance field (Self.Field += 1) in a method
		// 2. A class variable (TClass.ClassVar += 1)
		// The interpreter owns environment management and has access to Self/class context.
		// The evaluator cannot check Self/class scope without circular import.
		// KEEP: Architectural constraint - environment ownership
		return e.adapter.EvalNode(stmt)
	}

	// Cast to Value interface
	currentVal, ok := currentValRaw.(Value)
	if !ok {
		// Not a Value - this is an internal error
		return e.newError(target, "variable '%s' has invalid type (not a Value)", targetName)
	}

	// Check if target is a var parameter (ReferenceValue)
	// Compound assignment to var params requires read-modify-write
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

	// Evaluate the RHS
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
