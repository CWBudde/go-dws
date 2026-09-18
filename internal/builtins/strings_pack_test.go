package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

func TestStrArrayPack_MutatesAndReturnsReceiver(t *testing.T) {
	for _, tt := range []struct {
		name        string
		input, want []string
	}{
		{"empty", nil, nil},
		{"all empty", []string{"", "", ""}, nil},
		{"stable compact", []string{"", "a", "", " ", "a", "b", ""}, []string{"a", " ", "a", "b"}},
		{"already packed", []string{"a", "b"}, []string{"a", "b"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			arrayType := types.NewDynamicArrayType(types.STRING)
			arr := &runtime.ArrayValue{ArrayType: arrayType}
			for _, value := range tt.input {
				arr.Elements = append(arr.Elements, &runtime.StringValue{Value: value})
			}
			result := StrArrayPack(newMockContext(), []Value{arr})
			if result != arr {
				t.Fatalf("returned %p (%T), want original array %p", result, result, arr)
			}
			if arr.ArrayType != arrayType || len(arr.Elements) != len(tt.want) {
				t.Fatalf("metadata changed or length=%d, want %d", len(arr.Elements), len(tt.want))
			}
			for i, want := range tt.want {
				if got := arr.Elements[i].(*runtime.StringValue).Value; got != want {
					t.Errorf("element %d=%q, want %q", i, got, want)
				}
			}
		})
	}
}

func TestStrArrayPack_RejectsInvalidInputs(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []Value
	}{
		{"missing argument", nil},
		{"non-array", []Value{&runtime.StringValue{Value: "bad"}}},
		{"wrong empty array", []Value{&runtime.ArrayValue{ArrayType: types.NewDynamicArrayType(types.FLOAT)}}},
		{"invalid element", []Value{&runtime.ArrayValue{ArrayType: types.NewDynamicArrayType(types.STRING), Elements: []Value{&runtime.IntegerValue{Value: 1}}}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := StrArrayPack(newMockContext(), tt.args)
			if result == nil || result.Type() != "ERROR" {
				t.Fatalf("result=%v (%T), want error", result, result)
			}
		})
	}
}
