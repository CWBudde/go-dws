package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// VarType constants (Delphi-compatible) used by Variant type system.
const (
	VarEmpty   = 0
	VarNull    = 1
	VarInteger = 3
	VarDouble  = 5
	VarBoolean = 11
	VarInt64   = 20
	VarString  = 256
	VarArray   = 0x2000
	VarJSON    = 0x4000
)

// JSONValueToValue converts a jsonvalue.Value to a DWScript runtime Value.
func JSONValueToValue(jv *jsonvalue.Value) Value {
	if jv == nil {
		return &NilValue{}
	}

	switch jv.Kind() {
	case jsonvalue.KindUndefined, jsonvalue.KindNull:
		return &NilValue{}
	case jsonvalue.KindBoolean:
		return &BooleanValue{Value: jv.BoolValue()}
	case jsonvalue.KindInt64:
		return &IntegerValue{Value: jv.Int64Value()}
	case jsonvalue.KindNumber:
		return &FloatValue{Value: jv.NumberValue()}
	case jsonvalue.KindString:
		return &StringValue{Value: jv.StringValue()}
	case jsonvalue.KindArray, jsonvalue.KindObject:
		return NewJSONValue(jv)
	default:
		return &NilValue{}
	}
}

// ValueToJSONValue converts a DWScript runtime Value to a jsonvalue.Value.
func ValueToJSONValue(val Value) *jsonvalue.Value {
	if val == nil {
		return jsonvalue.NewNull()
	}
	if wrapper, ok := val.(VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		if unwrapped == nil {
			return jsonvalue.NewNull()
		}
		val = unwrapped
	}
	return valueToJSONValueUnwrapped(val)
}

func valueToJSONValueUnwrapped(val Value) *jsonvalue.Value {
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
	case *ArrayValue:
		arr := jsonvalue.NewArray()
		for _, elem := range v.Elements {
			arr.ArrayAppend(ValueToJSONValue(elem))
		}
		return arr
	case *RecordValue:
		obj := jsonvalue.NewObject()
		for fieldName, fieldValue := range v.Fields {
			obj.ObjectSet(fieldName, ValueToJSONValue(fieldValue))
		}
		return obj
	case *JSONValue:
		if v.Value == nil {
			return jsonvalue.NewNull()
		}
		return v.Value
	default:
		return jsonvalue.NewNull()
	}
}

// JSONKindToVarType maps a jsonvalue.Kind to a VarType code.
func JSONKindToVarType(kind jsonvalue.Kind) int64 {
	switch kind {
	case jsonvalue.KindUndefined:
		return VarEmpty
	case jsonvalue.KindNull:
		return VarNull
	case jsonvalue.KindBoolean:
		return VarBoolean
	case jsonvalue.KindInt64:
		return VarInt64
	case jsonvalue.KindNumber:
		return VarDouble
	case jsonvalue.KindString:
		return VarString
	case jsonvalue.KindArray:
		return VarArray
	case jsonvalue.KindObject:
		return VarJSON
	default:
		return VarEmpty
	}
}

// ParseJSONString parses a JSON string and returns a jsonvalue.Value.
func ParseJSONString(jsonStr string) (*jsonvalue.Value, error) {
	var data any
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return nil, FormatJSONError(err, jsonStr)
	}
	return GoValueToJSONValue(data), nil
}

// GoValueToJSONValue converts encoding/json decoded values to jsonvalue.Value.
func GoValueToJSONValue(data any) *jsonvalue.Value {
	if data == nil {
		return jsonvalue.NewNull()
	}

	switch v := data.(type) {
	case bool:
		return jsonvalue.NewBoolean(v)
	case json.Number:
		if i64, err := v.Int64(); err == nil {
			return jsonvalue.NewInt64(i64)
		}
		if f64, err := v.Float64(); err == nil {
			return jsonvalue.NewNumber(f64)
		}
		return jsonvalue.NewString(v.String())
	case float64:
		return jsonvalue.NewNumber(v)
	case string:
		return jsonvalue.NewString(v)
	case []any:
		arr := jsonvalue.NewArray()
		for _, elem := range v {
			arr.ArrayAppend(GoValueToJSONValue(elem))
		}
		return arr
	case map[string]any:
		obj := jsonvalue.NewObject()
		for key, value := range v {
			obj.ObjectSet(key, GoValueToJSONValue(value))
		}
		return obj
	default:
		return jsonvalue.NewNull()
	}
}

// FormatJSONError formats JSON parsing errors with position hints.
func FormatJSONError(err error, jsonStr string) error {
	// Keep behavior simple/compatible: include original error.
	// Detailed position formatting remains in evaluator if needed.
	return fmt.Errorf("%w", err)
}
