package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Assignment Helpers
// ============================================================================
//
// Task 3.5.105a: Simple variable assignment migration.
// This file contains helpers for evaluating assignment statements directly
// in the evaluator, reducing adapter dependency.
// ============================================================================

// ReferenceValueAccessor is an interface for values that can be dereferenced and assigned.
// This allows the evaluator to work with ReferenceValue without importing the interp package.
// Task 3.5.105a: Interface-based approach for var parameter handling.
type ReferenceValueAccessor interface {
	// Dereference returns the current value of the referenced variable.
	Dereference() (Value, error)
	// Assign sets the value of the referenced variable.
	Assign(value Value) error
}

// SubrangeValueAccessor is an interface for subrange values.
// Task 3.5.105a: Interface-based approach for subrange validation.
type SubrangeValueAccessor interface {
	// ValidateAndSet validates that the value is in range and sets it.
	ValidateAndSet(intValue int) error
	// GetValue returns the current integer value.
	GetValue() int
	// GetTypeName returns the subrange type name.
	GetTypeName() string
}

// Note: ExternalVarAccessor is defined in evaluator.go (Task 3.5.73).

// cloneIfCopyable returns a defensive copy for values that implement CopyableValue.
// DWScript static arrays have value semantics, so assignments should duplicate
// their backing storage to avoid accidental aliasing between variables.
// Dynamic arrays keep reference semantics.
//
// Task 3.5.105a: Migrated from statements_assignments.go.
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
			if copiedValue, ok := copied.(Value); ok {
				return copiedValue
			}
		}
	}

	return val
}

// evalSimpleAssignmentDirect handles simple variable assignment: x := value
// Task 3.5.105a: Migrated from Interpreter.evalSimpleAssignment()
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
		// Variable not in environment - check if we're in a method context
		// For now, delegate to adapter for Self/class context handling
		return e.adapter.EvalNode(stmt)
	}

	// Cast to Value interface
	existingVal, ok := existingValRaw.(Value)
	if !ok {
		// Not a Value - delegate to adapter
		return e.adapter.EvalNode(stmt)
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

	// Check for interface variable - delegate to adapter for complex handling
	if existingVal.Type() == "INTERFACE" {
		return e.adapter.EvalNode(stmt)
	}

	// Check for object variable - delegate to adapter for ref counting
	if existingVal.Type() == "OBJECT" {
		return e.adapter.EvalNode(stmt)
	}

	// Try implicit conversion if types don't match
	if value != nil {
		targetType := existingVal.Type()
		sourceType := value.Type()
		if targetType != sourceType {
			if converted, ok := e.adapter.TryImplicitConversion(value, targetType); ok {
				value = converted
			}
		}

		// Box value if target is a Variant
		if targetType == "VARIANT" && sourceType != "VARIANT" {
			value = e.adapter.BoxVariant(value)
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

	// Check if assigning an object (need ref counting) - delegate to adapter
	if value != nil && value.Type() == "OBJECT" {
		return e.adapter.EvalNode(stmt)
	}

	// Check if assigning an interface (need ref counting) - delegate to adapter
	if value != nil && value.Type() == "INTERFACE" {
		return e.adapter.EvalNode(stmt)
	}

	// Check if assigning a function pointer with object reference - delegate to adapter
	if value != nil && value.Type() == "FUNCTION_POINTER" {
		return e.adapter.EvalNode(stmt)
	}

	// Simple case: update the variable in the environment
	if e.SetVar(ctx, targetName, value) {
		return value
	}

	// Set failed - return error
	return e.newError(target, "undefined variable: %s", targetName)
}

// evalReferenceAssignment handles assignment through a var parameter (ReferenceValue).
// Task 3.5.105a: Extracted from evalSimpleAssignment for clarity.
func (e *Evaluator) evalReferenceAssignment(
	refVal ReferenceValueAccessor,
	value Value,
	target *ast.Identifier,
	stmt *ast.AssignmentStatement,
	_ *ExecutionContext, // ctx reserved for future use
) Value {
	// Get current value to check type compatibility
	currentVal, err := refVal.Dereference()
	if err != nil {
		return e.newError(target, "%s", err.Error())
	}

	// For complex types (interface, object), delegate to adapter
	if currentVal.Type() == "INTERFACE" || currentVal.Type() == "OBJECT" {
		return e.adapter.EvalNode(stmt)
	}

	// Try implicit conversion if types don't match
	targetType := currentVal.Type()
	sourceType := value.Type()
	if targetType != sourceType {
		if converted, ok := e.adapter.TryImplicitConversion(value, targetType); ok {
			value = converted
		}
	}

	// Box value if target is a Variant
	if targetType == "VARIANT" && sourceType != "VARIANT" {
		value = e.adapter.BoxVariant(value)
	}

	// Ensure value semantics for copyable types
	value = cloneIfCopyable(value)

	// Check if value is an object/interface - delegate for ref counting
	if value != nil && (value.Type() == "OBJECT" || value.Type() == "INTERFACE") {
		return e.adapter.EvalNode(stmt)
	}

	// Write through the reference
	if err := refVal.Assign(value); err != nil {
		return e.newError(target, "%s", err.Error())
	}

	return value
}

// evalSubrangeAssignment handles assignment to a subrange variable.
// Task 3.5.105a: Extracted from evalSimpleAssignment for clarity.
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
// Task 3.5.105b: Migrated from Interpreter.evalAssignmentStatement + applyCompoundOperation.
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
		// Variable not in environment - could be in Self/class context
		// Delegate to adapter for method context handling
		return e.adapter.EvalNode(stmt)
	}

	// Cast to Value interface
	currentVal, ok := currentValRaw.(Value)
	if !ok {
		// Not a Value - delegate to adapter
		return e.adapter.EvalNode(stmt)
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
