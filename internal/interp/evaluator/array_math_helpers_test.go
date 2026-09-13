package evaluator

import (
	"math"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

func floatHelperArray(values ...float64) *runtime.ArrayValue {
	arr := &runtime.ArrayValue{ArrayType: types.NewDynamicArrayType(types.FLOAT)}
	for _, value := range values {
		arr.Elements = append(arr.Elements, &runtime.FloatValue{Value: value})
	}
	return arr
}

func TestArrayMathHelper_MutationAndIdentity(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []Value
		want []float64
	}{
		{"__array_offset", []Value{&runtime.IntegerValue{Value: 3}}, []float64{5, 7}},
		{"__array_multiply", []Value{&runtime.FloatValue{Value: -2}}, []float64{-4, -8}},
		{"__array_multiplyadd", []Value{&runtime.IntegerValue{Value: 2}, &runtime.FloatValue{Value: 1}}, []float64{5, 9}},
		{"__array_reciprocal", nil, []float64{0.5, 0.25}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			arr := floatHelperArray(2, 4)
			firstScalar := arr.Elements[0].(*runtime.FloatValue)
			result := (&Evaluator{}).evalArrayHelper(tt.name, arr, tt.args, nil, nil)
			if result != arr {
				t.Fatalf("result=%v (%T), want original array", result, result)
			}
			for i, want := range tt.want {
				if got := arr.Elements[i].(*runtime.FloatValue).Value; got != want {
					t.Errorf("element %d=%v, want %v", i, got, want)
				}
			}
			if firstScalar.Value != 2 {
				t.Fatal("array mutation changed a separately held scalar value")
			}
		})
	}
}

func TestArrayMathHelper_ReciprocalIEEE(t *testing.T) {
	arr := floatHelperArray(0, math.Copysign(0, -1), math.Inf(1), math.Inf(-1), math.NaN())
	result := (&Evaluator{}).evalArrayHelper("__array_reciprocal", arr, nil, nil, nil)
	if result != arr {
		t.Fatalf("result=%v, want original array", result)
	}
	values := make([]float64, len(arr.Elements))
	for i, element := range arr.Elements {
		values[i] = element.(*runtime.FloatValue).Value
	}
	if !math.IsInf(values[0], 1) || !math.IsInf(values[1], -1) ||
		values[2] != 0 || math.Signbit(values[2]) || values[3] != 0 || !math.Signbit(values[3]) || !math.IsNaN(values[4]) {
		t.Fatalf("reciprocal results=%v", values)
	}
}

func TestArrayMathHelper_MultiplyAddRoundsProduct(t *testing.T) {
	arr := floatHelperArray(1 + math.Ldexp(1, -27))
	args := []Value{&runtime.FloatValue{Value: 1 - math.Ldexp(1, -27)}, &runtime.FloatValue{Value: -1}}
	result := (&Evaluator{}).evalArrayHelper("__array_multiplyadd", arr, args, nil, nil)
	if result != arr {
		t.Fatalf("result=%v, want original array", result)
	}
	if got := arr.Elements[0].(*runtime.FloatValue).Value; got != 0 {
		t.Fatalf("result=%g, want separately rounded multiply then add (0)", got)
	}
}

func TestArrayMathHelper_RejectsInvalidRuntimeInputs(t *testing.T) {
	for _, tt := range []struct {
		name string
		self Value
		args []Value
	}{
		{"non-array", &runtime.IntegerValue{Value: 2}, []Value{&runtime.IntegerValue{Value: 1}}},
		{"wrong empty array", &runtime.ArrayValue{ArrayType: types.NewDynamicArrayType(types.STRING)}, []Value{&runtime.IntegerValue{Value: 1}}},
		{"missing metadata", &runtime.ArrayValue{}, []Value{&runtime.IntegerValue{Value: 1}}},
		{"invalid element", &runtime.ArrayValue{ArrayType: types.NewDynamicArrayType(types.FLOAT), Elements: []Value{&runtime.StringValue{Value: "bad"}}}, []Value{&runtime.IntegerValue{Value: 1}}},
		{"missing operand", floatHelperArray(2), nil},
		{"extra operand", floatHelperArray(2), []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 2}}},
		{"non-numeric operand", floatHelperArray(2), []Value{&runtime.StringValue{Value: "bad"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := (&Evaluator{}).evalArrayHelper("__array_offset", tt.self, tt.args, nil, nil)
			if result == nil || !isError(result) {
				t.Fatalf("result=%v, want runtime error", result)
			}
		})
	}
}
