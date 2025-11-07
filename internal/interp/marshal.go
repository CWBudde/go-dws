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
//   - FUNCTION POINTER → func(...) (Task 9.4a - callback support)
//
// The interp parameter is optional and only required for function pointer marshaling.
// Pass nil if callbacks are not needed.
func MarshalToGo(dwsValue Value, targetType reflect.Type, interp *Interpreter) (any, error) {
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

	case reflect.Int32:
		goVal, err := GoInt(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected Integer, got %s", dwsValue.Type())
		}
		return int32(goVal), nil

	case reflect.Int16:
		goVal, err := GoInt(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected Integer, got %s", dwsValue.Type())
		}
		return int16(goVal), nil

	case reflect.Int8:
		goVal, err := GoInt(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected Integer, got %s", dwsValue.Type())
		}
		return int8(goVal), nil

	case reflect.Float64:
		goVal, err := GoFloat(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected Float, got %s", dwsValue.Type())
		}
		return goVal, nil

	case reflect.Float32:
		goVal, err := GoFloat(dwsValue)
		if err != nil {
			return nil, fmt.Errorf("expected Float, got %s", dwsValue.Type())
		}
		return float32(goVal), nil

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
			goElem, err := MarshalToGo(elem, elemType, interp)
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
			goElem, err := MarshalToGo(fieldVal, elemType, interp)
			if err != nil {
				return nil, fmt.Errorf("record field %s: %w", key, err)
			}
			goMap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(goElem))
		}

		return goMap.Interface(), nil

	case reflect.Ptr:
		// Handle pointer parameters (var parameters)
		// For pointer types, we need to:
		// 1. Get the underlying value from the DWScript variable
		// 2. Marshal it to the pointed-to type
		// 3. Create a Go pointer to that value
		// 4. The pointer will be passed to the Go function, which can modify it
		// 5. After the call, we'll unmarshal the modified value back (handled in Call())

		// Get the pointed-to type
		elemType := targetType.Elem()

		// Marshal the DWScript value to the pointed-to type
		elemValue, err := MarshalToGo(dwsValue, elemType, interp)
		if err != nil {
			return nil, fmt.Errorf("pointer element: %w", err)
		}

		// Create a pointer to the marshaled value
		ptrValue := reflect.New(elemType)
		ptrValue.Elem().Set(reflect.ValueOf(elemValue))

		return ptrValue.Interface(), nil

	case reflect.Func:
		// Marshal DWScript function pointers to Go callbacks
		// Check if DWScript value is a function pointer
		funcPtr, ok := dwsValue.(*FunctionPointerValue)
		if !ok {
			return nil, fmt.Errorf("expected function pointer, got %s", dwsValue.Type())
		}

		// Require interpreter for callbacks
		if interp == nil {
			return nil, fmt.Errorf("interpreter required for function pointer marshaling")
		}

		// Create Go wrapper function using reflection
		// The wrapper will capture the interpreter context and call back into DWScript
		goWrapper := createGoFunctionWrapper(funcPtr, targetType, interp)
		return goWrapper, nil

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

// UnmarshalFromGoPtr reads a value from a Go pointer and converts it back to DWScript.
//
// This function takes a Go pointer (reflect.Value) that was passed to a Go function
// as a var parameter, reads the (potentially modified) value it points to, and converts
// it back to a DWScript Value.
//
// Example:
//
//	func Increment(x *int64) { *x++ }
//	// After calling Increment, we use UnmarshalFromGoPtr to get the modified value
func UnmarshalFromGoPtr(ptrValue reflect.Value) (Value, error) {
	if ptrValue.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("expected pointer, got %s", ptrValue.Kind())
	}

	// Dereference the pointer to get the actual value
	elemValue := ptrValue.Elem()

	// Convert the dereferenced value to DWScript
	return MarshalToDWS(elemValue.Interface())
}
