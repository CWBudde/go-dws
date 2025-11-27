package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Enum Helper Method Implementations
// ============================================================================
// Task 3.5.102f: Migrate enum helper methods from Interpreter to Evaluator.
//
// These implementations avoid the adapter by directly manipulating runtime values.
// The goal is to remove EvalNode delegation for common enum operations.

// evalEnumHelper evaluates a built-in enum helper method directly in the evaluator.
// Returns the result value, or nil if this helper is not handled here (should fall through
// to the adapter).
//
// Task 3.5.102f: Handles enum operations that don't require the builtins Context.
func (e *Evaluator) evalEnumHelper(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	case "__enum_value":
		return e.evalEnumValue(selfValue, args, node)

	case "__enum_name":
		return e.evalEnumName(selfValue, args, node)

	case "__enum_qualifiedname":
		return e.evalEnumQualifiedName(selfValue, args, node)

	default:
		// Not an enum helper we handle - return nil to signal fallthrough to adapter
		return nil
	}
}

// evalEnumValue implements Enum.Value property.
// Returns the ordinal (integer) value of the enum.
func (e *Evaluator) evalEnumValue(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Enum.Value property does not take arguments")
	}

	enumVal, ok := selfValue.(*runtime.EnumValue)
	if !ok {
		return e.newError(node, "Enum.Value property requires enum receiver")
	}

	return &runtime.IntegerValue{Value: int64(enumVal.OrdinalValue)}
}

// evalEnumName implements Enum.Name property.
// Returns the value name of the enum (e.g., "Red" for TColor.Red).
// Returns "?" if the enum value has no name (invalid ordinal).
func (e *Evaluator) evalEnumName(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Enum.Name property does not take arguments")
	}

	enumVal, ok := selfValue.(*runtime.EnumValue)
	if !ok {
		return e.newError(node, "Enum.Name property requires enum receiver")
	}

	// If the enum value doesn't have a name (invalid ordinal), return "?"
	if enumVal.ValueName == "" {
		return &runtime.StringValue{Value: "?"}
	}
	return &runtime.StringValue{Value: enumVal.ValueName}
}

// evalEnumQualifiedName implements Enum.QualifiedName property.
// Returns the fully qualified name (e.g., "TColor.Red").
// Returns "TypeName.?" if the enum value has no name (invalid ordinal).
func (e *Evaluator) evalEnumQualifiedName(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Enum.QualifiedName property does not take arguments")
	}

	enumVal, ok := selfValue.(*runtime.EnumValue)
	if !ok {
		return e.newError(node, "Enum.QualifiedName property requires enum receiver")
	}

	// Return TypeName.ValueName (e.g., "TColor.Red")
	// If the enum value doesn't have a name (invalid ordinal), return "TypeName.?"
	valueName := enumVal.ValueName
	if valueName == "" {
		valueName = "?"
	}
	return &runtime.StringValue{Value: enumVal.TypeName + "." + valueName}
}
