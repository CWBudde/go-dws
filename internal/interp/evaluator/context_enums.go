package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// ============================================================================
// Enum Operation Methods
// ============================================================================
//
// This file implements the enum operation methods of the builtins.Context
// interface for the Evaluator:
// - GetEnumOrdinal(): Get ordinal value of an enum
// - GetEnumSuccessor(): Get next enum value in sequence
// - GetEnumPredecessor(): Get previous enum value in sequence
// - GetJSONVarType(): Get VarType code for JSON values
// - GetEnumMetadata(): Get enum type metadata by type name
//
// These implementations are self-contained and do not require callbacks
// to the interpreter.
// ============================================================================

// VarType constants matching DWScript's VarType values.
const (
	varEmpty = 0
)

// GetEnumOrdinal returns the ordinal value of an enum Value.
// This implements the builtins.Context interface.
func (e *Evaluator) GetEnumOrdinal(value Value) (int64, bool) {
	if enumVal, ok := value.(*runtime.EnumValue); ok {
		return int64(enumVal.OrdinalValue), true
	}
	return 0, false
}

// GetEnumSuccessor returns the successor of an enum value.
// This implements the builtins.Context interface.
func (e *Evaluator) GetEnumSuccessor(enumVal Value) (Value, error) {
	val, ok := enumVal.(*runtime.EnumValue)
	if !ok {
		return nil, fmt.Errorf("expected EnumValue, got %T", enumVal)
	}

	// Get enum type metadata via TypeSystem
	enumMetadata := e.typeSystem.LookupEnumMetadata(val.TypeName)
	if enumMetadata == nil {
		return nil, fmt.Errorf("enum type metadata not found for %s", val.TypeName)
	}

	etv, ok := enumMetadata.(EnumTypeValueAccessor)
	if !ok {
		return nil, fmt.Errorf("invalid enum type metadata for %s", val.TypeName)
	}
	enumType := etv.GetEnumType()

	currentPos, err := runtime.EnumValueIndex(val, enumType)
	if err != nil {
		return nil, err
	}

	// Check if we can increment (not at the end)
	if currentPos >= len(enumType.OrderedNames)-1 {
		return nil, fmt.Errorf("cannot get successor of maximum enum value")
	}

	// Get next value
	return runtime.EnumValueAtIndex(val.TypeName, enumType, currentPos+1)
}

// GetEnumPredecessor returns the predecessor of an enum value.
// This implements the builtins.Context interface.
func (e *Evaluator) GetEnumPredecessor(enumVal Value) (Value, error) {
	val, ok := enumVal.(*runtime.EnumValue)
	if !ok {
		return nil, fmt.Errorf("expected EnumValue, got %T", enumVal)
	}

	// Get enum type metadata via TypeSystem
	enumMetadata := e.typeSystem.LookupEnumMetadata(val.TypeName)
	if enumMetadata == nil {
		return nil, fmt.Errorf("enum type metadata not found for %s", val.TypeName)
	}

	etv, ok := enumMetadata.(EnumTypeValueAccessor)
	if !ok {
		return nil, fmt.Errorf("invalid enum type metadata for %s", val.TypeName)
	}
	enumType := etv.GetEnumType()

	currentPos, err := runtime.EnumValueIndex(val, enumType)
	if err != nil {
		return nil, err
	}

	// Check if we can decrement (not at the beginning)
	if currentPos <= 0 {
		return nil, fmt.Errorf("cannot get predecessor of minimum enum value")
	}

	// Get previous value
	return runtime.EnumValueAtIndex(val.TypeName, enumType, currentPos-1)
}

// GetJSONVarType returns the VarType code for a JSON value based on its kind.
// This implements the builtins.Context interface.
func (e *Evaluator) GetJSONVarType(value Value) (int64, bool) {
	jsonVal, ok := value.(*runtime.JSONValue)
	if !ok {
		return 0, false
	}

	// Return VarType code based on JSON kind
	if jsonVal.Value == nil {
		return varEmpty, true
	}
	return jsonKindToVarType(jsonVal.Value.Kind()), true
}

// jsonKindToVarType converts a JSON kind to its VarType code.
func jsonKindToVarType(kind jsonvalue.Kind) int64 {
	// VarType values matching DWScript conventions
	const (
		varNull    = 1
		varBoolean = 11
		varInteger = 3
		varDouble  = 5
		varString  = 8
		varArray   = 8204 // vArray + vVariant
		varObject  = 9    // Similar to vDispatch
	)

	switch kind {
	case jsonvalue.KindNull:
		return varNull
	case jsonvalue.KindBoolean:
		return varBoolean
	case jsonvalue.KindNumber, jsonvalue.KindInt64:
		// Check if it's an integer or float
		return varDouble // Default to double for JSON numbers
	case jsonvalue.KindString:
		return varString
	case jsonvalue.KindArray:
		return varArray
	case jsonvalue.KindObject:
		return varObject
	default:
		return varEmpty
	}
}

// GetEnumMetadata retrieves enum type metadata by type name.
// This implements the builtins.Context interface.
// Used by Succ/Pred builtins to navigate enum ordinals.
func (e *Evaluator) GetEnumMetadata(typeName string) builtins.Value {
	metadata := e.typeSystem.LookupEnumMetadata(typeName)
	if metadata == nil {
		return nil
	}
	if val, ok := metadata.(builtins.Value); ok {
		return val
	}
	return nil
}
