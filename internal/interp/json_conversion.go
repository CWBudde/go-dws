package interp

import (
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// JSON Value Conversions
// Task 9.89: Bidirectional conversions between jsonvalue.Value and DWScript runtime values
// ============================================================================

// jsonValueToValue converts a jsonvalue.Value to a DWScript runtime Value.
// This is used when accessing JSON properties/elements to get native DWScript values.
//
// Conversion rules:
//   - JSON null → NilValue
//   - JSON boolean → BooleanValue
//   - JSON number → FloatValue
//   - JSON int64 → IntegerValue
//   - JSON string → StringValue
//   - JSON array → JSONValue (keep as reference for mutation)
//   - JSON object → JSONValue (keep as reference for mutation)
//
// Containers (arrays/objects) remain as JSONValue to preserve reference semantics
// and allow mutations to be visible.
func jsonValueToValue(jv *jsonvalue.Value) Value {
	if jv == nil {
		return &NilValue{}
	}

	switch jv.Kind() {
	case jsonvalue.KindUndefined:
		return &NilValue{}
	case jsonvalue.KindNull:
		return &NilValue{}
	case jsonvalue.KindBoolean:
		return &BooleanValue{Value: jv.BoolValue()}
	case jsonvalue.KindInt64:
		return &IntegerValue{Value: jv.Int64Value()}
	case jsonvalue.KindNumber:
		return &FloatValue{Value: jv.NumberValue()}
	case jsonvalue.KindString:
		return &StringValue{Value: jv.StringValue()}
	case jsonvalue.KindArray:
		// Keep arrays as JSONValue for reference semantics
		return &JSONValue{Value: jv}
	case jsonvalue.KindObject:
		// Keep objects as JSONValue for reference semantics
		return &JSONValue{Value: jv}
	default:
		return &NilValue{}
	}
}

// valueToJSONValue converts a DWScript runtime Value to a jsonvalue.Value.
// This is used when building JSON from DWScript values (for ToJSON/Stringify).
//
// Conversion rules:
//   - NilValue → JSON null
//   - BooleanValue → JSON boolean
//   - IntegerValue → JSON int64
//   - FloatValue → JSON number
//   - StringValue → JSON string
//   - ArrayValue → JSON array (recursive conversion)
//   - RecordValue → JSON object (recursive conversion)
//   - JSONValue → unwrap and return underlying jsonvalue.Value
//   - VariantValue → unwrap and convert the wrapped value
func valueToJSONValue(val Value) *jsonvalue.Value {
	if val == nil {
		return jsonvalue.NewNull()
	}

	// Unwrap Variant if present
	val = unwrapVariant(val)

	switch v := val.(type) {
	case *NilValue:
		return jsonvalue.NewNull()
	case *BooleanValue:
		return jsonvalue.NewBoolean(v.Value)
	case *IntegerValue:
		return jsonvalue.NewInt64(v.Value)
	case *FloatValue:
		return jsonvalue.NewNumber(v.Value)
	case *StringValue:
		return jsonvalue.NewString(v.Value)
	case *JSONValue:
		// Already a JSON value, return its underlying value
		return v.Value
	case *ArrayValue:
		// Convert DWScript array to JSON array
		arr := jsonvalue.NewArray()
		for _, elem := range v.Elements {
			// Recursively convert each element
			jsonElem := valueToJSONValue(elem)
			arr.ArrayAppend(jsonElem)
		}
		return arr
	case *RecordValue:
		// Convert DWScript record to JSON object
		obj := jsonvalue.NewObject()
		for fieldName, fieldValue := range v.Fields {
			// Recursively convert each field
			jsonField := valueToJSONValue(fieldValue)
			obj.ObjectSet(fieldName, jsonField)
		}
		return obj
	default:
		// For unknown types, return null
		return jsonvalue.NewNull()
	}
}

// jsonValueToVariant wraps a jsonvalue.Value in a VariantValue.
// Task 9.89: Enable JSON values to be stored in Variants.
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
// Task 9.90: Support VarType introspection for JSON values.
//
// This allows VarType(jsonVar) to return meaningful type codes.
// We use custom codes in the high range to avoid conflicts with standard VarType codes.
func jsonKindToVarType(kind jsonvalue.Kind) int64 {
	switch kind {
	case jsonvalue.KindUndefined:
		return varEmpty // 0
	case jsonvalue.KindNull:
		return varNull // 1
	case jsonvalue.KindBoolean:
		return varBoolean // 11
	case jsonvalue.KindInt64:
		return varInt64 // 20
	case jsonvalue.KindNumber:
		return varDouble // 5
	case jsonvalue.KindString:
		return varString // 256
	case jsonvalue.KindArray:
		return varArray // 0x2000
	case jsonvalue.KindObject:
		return varJSON // Custom code for JSON objects (to be defined)
	default:
		return varEmpty
	}
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
