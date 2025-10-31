package interp

import (
	"fmt"
	"reflect"

	"github.com/cwbudde/go-dws/internal/types"
)

// MarshalToGo converts a DWScript Value to a Go value of the target type.
// This function handles the conversion from DWScript runtime values to Go native types
// for use in FFI (Foreign Function Interface) calls.
//
// Supported conversions:
//   - INTEGER → int64, int, int32, int16, int8
//   - FLOAT → float64, float32
//   - STRING → string
//   - BOOLEAN → bool
//   - ARRAY → []T (Go slices)
//   - RECORD → map[string]T (Go maps with string keys)
func MarshalToGo(dwsValue Value, targetType reflect.Type) (any, error) {
	switch targetType.Kind() {
	case reflect.Int64:
		goVal, err := GoInt(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected Integer, got %s", dwsValue.Type())
		}
		return goVal, nil

	case reflect.Int:
		goVal, err := GoInt(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected Integer, got %s", dwsValue.Type())
		}
		return int(goVal), nil

	case reflect.Float64:
		goVal, err := GoFloat(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected Float, got %s", dwsValue.Type())
		}
		return goVal, nil

	case reflect.String:
		goVal, err := GoString(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected String, got %s", dwsValue.Type())
		}
		return goVal, nil

	case reflect.Bool:
		goVal, err := GoBool(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected Boolean, got %s", dwsValue.Type())
		}
		return goVal, nil

	case reflect.Slice:
		// Convert DWScript array to Go slice
		if dwsValue.Type() != "ARRAY" {
			return nil, fmt.Errorf("expected ARRAY, got %s", dwsValue.Type())
		}
		arrayVal, ok := dwsValue.(*ArrayValue)
		if !ok {
			return nil, fmt.Errorf("expected ArrayValue, got %T", dwsValue)
		}

		// Create a slice of the target element type
		elemType := targetType.Elem()
		goSlice := reflect.MakeSlice(targetType, len(arrayVal.Elements), len(arrayVal.Elements))

		// Convert each element
		for i, elem := range arrayVal.Elements {
			goElem, err := MarshalToGo(elem, elemType)
			if err != nil {
				return nil, fmt.Errorf("array element %d: %w", i, err)
			}
			goSlice.Index(i).Set(reflect.ValueOf(goElem))
		}

		return goSlice.Interface(), nil

	case reflect.Map:
		// Convert DWScript record to Go map[string]T
		if targetType.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("only map[string]T is supported")
		}

		if dwsValue.Type() != "RECORD" {
			return nil, fmt.Errorf("expected RECORD, got %s", dwsValue.Type())
		}
		recordVal, ok := dwsValue.(*RecordValue)
		if !ok {
			return nil, fmt.Errorf("expected RecordValue, got %T", dwsValue)
		}

		// Create a map of the target type
		elemType := targetType.Elem()
		goMap := reflect.MakeMap(targetType)

		// Convert each field
		for key, fieldVal := range recordVal.Fields {
			goElem, err := MarshalToGo(fieldVal, elemType)
			if err != nil {
				return nil, fmt.Errorf("record field %s: %w", key, err)
			}
			goMap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(goElem))
		}

		return goMap.Interface(), nil

	default:
		return nil, fmt.Errorf("unsupported target type: %s", targetType)
	}
}

// MarshalToDWS converts a Go value to a DWScript Value.
// This function handles the conversion from Go native types to DWScript runtime values
// for use in FFI (Foreign Function Interface) return values.
//
// Supported conversions:
//   - int64, int, int32, int16, int8 → INTEGER
//   - float64, float32 → FLOAT
//   - string → STRING
//   - bool → BOOLEAN
//   - []T → ARRAY
//   - map[string]T → RECORD
//   - nil → NIL
func MarshalToDWS(goValue any) (Value, error) {
	// Handle nil
	if goValue == nil {
		return NewNilValue(), nil
	}

	// Use reflection to handle more complex types
	v := reflect.ValueOf(goValue)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewIntegerValue(v.Int()), nil

	case reflect.Float32, reflect.Float64:
		return NewFloatValue(v.Float()), nil

	case reflect.String:
		return NewStringValue(v.String()), nil

	case reflect.Bool:
		return NewBooleanValue(v.Bool()), nil

	case reflect.Slice:
		// Convert Go slice to DWScript array
		elements := make([]Value, v.Len())
		for i := 0; i < v.Len(); i++ {
			elem, err := MarshalToDWS(v.Index(i).Interface())
			if err != nil {
				return nil, fmt.Errorf("slice element %d: %w", i, err)
			}
			elements[i] = elem
		}

		// Determine element type from first element (if any)
		var elemType types.Type = types.NIL
		if len(elements) > 0 {
			switch elements[0].Type() {
			case "INTEGER":
				elemType = types.INTEGER
			case "FLOAT":
				elemType = types.FLOAT
			case "STRING":
				elemType = types.STRING
			case "BOOLEAN":
				elemType = types.BOOLEAN
			}
		} else {
			// Empty array - try to infer from Go type
			if v.Type().Elem().Kind() == reflect.Int64 {
				elemType = types.INTEGER
			}
		}

		return &ArrayValue{
			ArrayType: &types.ArrayType{
				ElementType: elemType,
				LowBound:    nil, // Dynamic array
				HighBound:   nil,
			},
			Elements: elements,
		}, nil

	case reflect.Map:
		// Convert Go map[string]T to DWScript record
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("only map[string]T is supported for marshaling")
		}

		fields := make(map[string]Value)
		for _, key := range v.MapKeys() {
			keyStr := key.String()
			val, err := MarshalToDWS(v.MapIndex(key).Interface())
			if err != nil {
				return nil, fmt.Errorf("map field %s: %w", keyStr, err)
			}
			fields[keyStr] = val
		}

		return &RecordValue{
			RecordType: nil, // Anonymous record
			Fields:     fields,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported Go return type: %T", goValue)
	}
}
