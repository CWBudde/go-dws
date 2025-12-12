package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// ============================================================================
// VarType Constants
// ============================================================================
// These constants are re-exported from runtime for compatibility.

const (
	varEmpty   = runtime.VarEmpty
	varNull    = runtime.VarNull
	varInteger = runtime.VarInteger
	varDouble  = runtime.VarDouble
	varBoolean = runtime.VarBoolean
	varInt64   = runtime.VarInt64
	varString  = runtime.VarString
	varArray   = runtime.VarArray
	varJSON    = runtime.VarJSON
)

// ============================================================================
// JSON Value Conversions
// ============================================================================

// jsonValueToValue converts a jsonvalue.Value to a DWScript runtime Value.
func jsonValueToValue(jv *jsonvalue.Value) Value {
	return runtime.JSONValueToValue(jv)
}

// valueToJSONValue converts a DWScript runtime Value to a jsonvalue.Value.
func valueToJSONValue(val Value) *jsonvalue.Value {
	return runtime.ValueToJSONValue(val)
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
	return runtime.JSONKindToVarType(kind)
}

// ============================================================================
// JSON Parsing Helpers
// ============================================================================

// parseJSONString parses a JSON string and returns a jsonvalue.Value.
// This is the core JSON parsing function used by ParseJSON and related functions.
func parseJSONString(jsonStr string) (*jsonvalue.Value, error) {
	return runtime.ParseJSONString(jsonStr)
}
