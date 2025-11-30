package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// VarType Constants
// ============================================================================
// These constants are re-exported from evaluator for compatibility.
// Task 3.5.143d: VarType constants now defined in evaluator/json_helpers.go

const (
	varEmpty   = evaluator.VarEmpty
	varNull    = evaluator.VarNull
	varInteger = evaluator.VarInteger
	varDouble  = evaluator.VarDouble
	varBoolean = evaluator.VarBoolean
	varInt64   = evaluator.VarInt64
	varString  = evaluator.VarString
	varArray   = evaluator.VarArray
	varJSON    = evaluator.VarJSON
)

// ============================================================================
// JSON Value Conversions
// ============================================================================

// jsonValueToValue converts a jsonvalue.Value to a DWScript runtime Value.
// Task 3.5.143d: Delegates to evaluator helper, with special handling for JSONValue.
func jsonValueToValue(jv *jsonvalue.Value) Value {
	if jv == nil {
		return &NilValue{}
	}

	// For arrays and objects, wrap in JSONValue to preserve reference semantics
	kind := jv.Kind()
	if kind == jsonvalue.KindArray || kind == jsonvalue.KindObject {
		return &JSONValue{Value: jv}
	}

	// For primitives, delegate to evaluator helper
	result := evaluator.JSONValueToValue(jv)

	// Convert runtime types back to interp types
	switch v := result.(type) {
	case *NilValue:
		return v
	case *BooleanValue:
		return v
	case *IntegerValue:
		return v
	case *FloatValue:
		return v
	case *StringValue:
		return v
	default:
		return &NilValue{}
	}
}

// valueToJSONValue converts a DWScript runtime Value to a jsonvalue.Value.
// Task 3.5.143d: Delegates to evaluator helper function.
func valueToJSONValue(val Value) *jsonvalue.Value {
	return evaluator.ValueToJSONValue(val)
}

// jsonValueToVariant wraps a jsonvalue.Value in a VariantValue.
//
// This creates a JSONValue wrapper and boxes it in a Variant.
// The Variant's ActualType is set to nil since JSON is a dynamic type.
func jsonValueToVariant(jv *jsonvalue.Value) *VariantValue {
	if jv == nil {
		return &VariantValue{Value: nil, ActualType: nil}
	}

	// Wrap in JSONValue
	jsonVal := &JSONValue{Value: jv}

	// Box in Variant
	// Note: We don't have a specific types.Type for JSON, so use nil
	return &VariantValue{
		Value:      jsonVal,
		ActualType: nil, // JSON is a dynamic type
	}
}

// variantToJSONValue extracts the underlying jsonvalue.Value from a Variant.
// If the Variant contains a JSONValue, returns its underlying jsonvalue.Value.
// Otherwise, converts the Variant's value to JSON.
func variantToJSONValue(variant *VariantValue) *jsonvalue.Value {
	if variant == nil || variant.Value == nil {
		return jsonvalue.NewNull()
	}

	// If it's already a JSONValue, extract it
	if jsonVal, ok := variant.Value.(*JSONValue); ok {
		return jsonVal.Value
	}

	// Otherwise, convert the value to JSON
	return valueToJSONValue(variant.Value)
}

// jsonKindToVarType maps a jsonvalue.Kind to a VarType code.
// Task 3.5.143d: Delegates to evaluator helper function.
func jsonKindToVarType(kind jsonvalue.Kind) int64 {
	return evaluator.JSONKindToVarType(kind)
}

// getJSONValueType returns the semantic type for a JSONValue.
// This is used for type checking in the semantic analyzer.
//
// Since JSON is a dynamic type, we return a generic JSONType.
// Individual elements (when accessed) have their actual types.
func getJSONValueType(jv *JSONValue) types.Type {
	// For now, return nil to indicate dynamic typing
	// TODO: Create a proper JSONType in types package if needed for semantic analysis
	return nil
}

// ============================================================================
// JSON Parsing Helpers
// ============================================================================
// Task 3.5.143d: All parsing helpers now delegate to evaluator functions.

// parseJSONString parses a JSON string and returns a jsonvalue.Value.
// This is the core JSON parsing function used by ParseJSON and related functions.
func parseJSONString(jsonStr string) (*jsonvalue.Value, error) {
	return evaluator.ParseJSONString(jsonStr)
}
