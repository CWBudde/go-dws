package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// ============================================================================
// VarType Constants
// ============================================================================
// These constants are re-exported from evaluator for compatibility.

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
func valueToJSONValue(val Value) *jsonvalue.Value {
	return evaluator.ValueToJSONValue(val)
}

// jsonValueToVariant wraps a jsonvalue.Value in a VariantValue.
func jsonValueToVariant(jv *jsonvalue.Value) *VariantValue {
	return runtime.BoxVariantWithJSON(jv)
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
func jsonKindToVarType(kind jsonvalue.Kind) int64 {
	return evaluator.JSONKindToVarType(kind)
}

// ============================================================================
// JSON Parsing Helpers
// ============================================================================

// parseJSONString parses a JSON string and returns a jsonvalue.Value.
// This is the core JSON parsing function used by ParseJSON and related functions.
func parseJSONString(jsonStr string) (*jsonvalue.Value, error) {
	return evaluator.ParseJSONString(jsonStr)
}
