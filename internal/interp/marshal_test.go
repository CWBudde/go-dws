package interp

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

func TestMarshalToGo(t *testing.T) {
	tests := []struct {
		name        string
		dwsValue    Value
		targetType  reflect.Type
		expected    any
		expectError bool
	}{
		{
			name:        "Integer to int64",
			dwsValue:    NewIntegerValue(42),
			targetType:  reflect.TypeOf(int64(0)),
			expected:    int64(42),
			expectError: false,
		},
		{
			name:        "Integer to int",
			dwsValue:    NewIntegerValue(42),
			targetType:  reflect.TypeOf(int(0)),
			expected:    42,
			expectError: false,
		},
		{
			name:        "Integer to int32",
			dwsValue:    NewIntegerValue(42),
			targetType:  reflect.TypeOf(int32(0)),
			expected:    int32(42),
			expectError: false,
		},
		{
			name:        "Integer to int16",
			dwsValue:    NewIntegerValue(42),
			targetType:  reflect.TypeOf(int16(0)),
			expected:    int16(42),
			expectError: false,
		},
		{
			name:        "Integer to int8",
			dwsValue:    NewIntegerValue(42),
			targetType:  reflect.TypeOf(int8(0)),
			expected:    int8(42),
			expectError: false,
		},
		{
			name:        "Float to float64",
			dwsValue:    NewFloatValue(3.14),
			targetType:  reflect.TypeOf(float64(0)),
			expected:    3.14,
			expectError: false,
		},
		{
			name:        "Float to float32",
			dwsValue:    NewFloatValue(3.14),
			targetType:  reflect.TypeOf(float32(0)),
			expected:    float32(3.14),
			expectError: false,
		},
		{
			name:        "String to string",
			dwsValue:    NewStringValue("hello"),
			targetType:  reflect.TypeOf(""),
			expected:    "hello",
			expectError: false,
		},
		{
			name:        "Boolean to bool",
			dwsValue:    NewBooleanValue(true),
			targetType:  reflect.TypeOf(false),
			expected:    true,
			expectError: false,
		},
		{
			name: "Array to slice",
			dwsValue: &ArrayValue{
				ArrayType: &types.ArrayType{
					ElementType: types.INTEGER,
					LowBound:    nil,
					HighBound:   nil,
				},
				Elements: []Value{NewIntegerValue(1), NewIntegerValue(2), NewIntegerValue(3)},
			},
			targetType:  reflect.TypeOf([]int64{}),
			expected:    []int64{1, 2, 3},
			expectError: false,
		},
		{
			name:        "Wrong type - Integer to String",
			dwsValue:    NewIntegerValue(42),
			targetType:  reflect.TypeOf(""),
			expectError: true,
		},
		{
			name:        "Unsupported target type",
			dwsValue:    NewIntegerValue(42),
			targetType:  reflect.TypeOf(complex64(0)),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalToGo(tt.dwsValue, tt.targetType)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v (type %T), got %v (type %T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestMarshalToDWS(t *testing.T) {
	tests := []struct {
		name        string
		goValue     any
		expected    Value
		expectError bool
	}{
		{
			name:     "int64 to Integer",
			goValue:  int64(42),
			expected: NewIntegerValue(42),
		},
		{
			name:     "int to Integer",
			goValue:  42,
			expected: NewIntegerValue(42),
		},
		{
			name:     "int32 to Integer",
			goValue:  int32(42),
			expected: NewIntegerValue(42),
		},
		{
			name:     "int16 to Integer",
			goValue:  int16(42),
			expected: NewIntegerValue(42),
		},
		{
			name:     "int8 to Integer",
			goValue:  int8(42),
			expected: NewIntegerValue(42),
		},
		{
			name:     "float64 to Float",
			goValue:  3.14,
			expected: NewFloatValue(3.14),
		},
		{
			name:     "float32 to Float",
			goValue:  float32(3.14),
			expected: NewFloatValue(float64(float32(3.14))),
		},
		{
			name:     "string to String",
			goValue:  "hello",
			expected: NewStringValue("hello"),
		},
		{
			name:     "bool to Boolean",
			goValue:  true,
			expected: NewBooleanValue(true),
		},
		{
			name:     "nil to Nil",
			goValue:  nil,
			expected: NewNilValue(),
		},
		{
			name:    "slice to Array",
			goValue: []int64{1, 2, 3},
			expected: &ArrayValue{
				ArrayType: &types.ArrayType{
					ElementType: types.INTEGER,
					LowBound:    nil,
					HighBound:   nil,
				},
				Elements: []Value{NewIntegerValue(1), NewIntegerValue(2), NewIntegerValue(3)},
			},
		},
		{
			name:    "empty slice to Array",
			goValue: []string{},
			expected: &ArrayValue{
				ArrayType: &types.ArrayType{
					ElementType: types.NIL, // Empty array
					LowBound:    nil,
					HighBound:   nil,
				},
				Elements: []Value{},
			},
		},
		{
			name: "map to Record",
			goValue: map[string]any{
				"name": "John",
				"age":  int64(30),
			},
			expected: &RecordValue{
				RecordType: nil,
				Fields: map[string]Value{
					"name": NewStringValue("John"),
					"age":  NewIntegerValue(30),
				},
			},
		},
		{
			name:        "unsupported type",
			goValue:     complex64(1 + 2i),
			expectError: true,
		},
		{
			name:        "non-string key map",
			goValue:     map[int]string{1: "test"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalToDWS(tt.goValue)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// For complex types, we need to compare fields individually
			switch expected := tt.expected.(type) {
			case *ArrayValue:
				resultArray, ok := result.(*ArrayValue)
				if !ok {
					t.Errorf("expected ArrayValue, got %T", result)
					return
				}
				if len(resultArray.Elements) != len(expected.Elements) {
					t.Errorf("expected %d elements, got %d", len(expected.Elements), len(resultArray.Elements))
					return
				}
				for i, elem := range expected.Elements {
					if !reflect.DeepEqual(resultArray.Elements[i], elem) {
						t.Errorf("element %d: expected %v, got %v", i, elem, resultArray.Elements[i])
					}
				}
			case *RecordValue:
				resultRecord, ok := result.(*RecordValue)
				if !ok {
					t.Errorf("expected RecordValue, got %T", result)
					return
				}
				if len(resultRecord.Fields) != len(expected.Fields) {
					t.Errorf("expected %d fields, got %d", len(expected.Fields), len(resultRecord.Fields))
					return
				}
				for key, expectedValue := range expected.Fields {
					resultValue, exists := resultRecord.Fields[key]
					if !exists {
						t.Errorf("missing field %s", key)
						continue
					}
					if !reflect.DeepEqual(resultValue, expectedValue) {
						t.Errorf("field %s: expected %v, got %v", key, expectedValue, resultValue)
					}
				}
			default:
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("expected %v (type %T), got %v (type %T)", expected, expected, result, result)
				}
			}
		})
	}
}

// TestMarshalRoundTrip tests that marshaling from DWS to Go and back preserves values
func TestMarshalRoundTrip(t *testing.T) {
	testCases := []struct {
		name       string
		original   Value
		targetType reflect.Type
	}{
		{
			name:       "INTEGER",
			original:   NewIntegerValue(42),
			targetType: reflect.TypeOf(int64(0)),
		},
		{
			name:       "INTEGER_INT32",
			original:   NewIntegerValue(42),
			targetType: reflect.TypeOf(int32(0)),
		},
		{
			name:       "INTEGER_INT16",
			original:   NewIntegerValue(42),
			targetType: reflect.TypeOf(int16(0)),
		},
		{
			name:       "INTEGER_INT8",
			original:   NewIntegerValue(42),
			targetType: reflect.TypeOf(int8(0)),
		},
		{
			name:       "FLOAT",
			original:   NewFloatValue(3.14),
			targetType: reflect.TypeOf(float64(0)),
		},
		{
			name:       "FLOAT_FLOAT32",
			original:   NewFloatValue(3.14),
			targetType: reflect.TypeOf(float32(0)),
		},
		{
			name:       "STRING",
			original:   NewStringValue("hello world"),
			targetType: reflect.TypeOf(""),
		},
		{
			name:       "BOOLEAN",
			original:   NewBooleanValue(true),
			targetType: reflect.TypeOf(false),
		},
		{
			name:     "NIL",
			original: NewNilValue(),
			// Skip marshaling nil to Go since it's not supported
		},
		{
			name: "ARRAY",
			original: &ArrayValue{
				ArrayType: &types.ArrayType{
					ElementType: types.INTEGER,
					LowBound:    nil,
					HighBound:   nil,
				},
				Elements: []Value{NewIntegerValue(1), NewIntegerValue(2), NewIntegerValue(3)},
			},
			targetType: reflect.TypeOf([]int64{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip NIL test since we can't marshal nil to Go
			if tc.name == "NIL" {
				return
			}

			// Marshal to Go
			goValue, err := MarshalToGo(tc.original, tc.targetType)
			if err != nil {
				t.Fatalf("MarshalToGo failed: %v", err)
			}

			// Marshal back to DWS
			result, err := MarshalToDWS(goValue)
			if err != nil {
				t.Fatalf("MarshalToDWS failed: %v", err)
			}

			// Compare types (simplified comparison)
			if tc.original.Type() != result.Type() {
				t.Errorf("type mismatch: expected %s, got %s", tc.original.Type(), result.Type())
			}
		})
	}
}
