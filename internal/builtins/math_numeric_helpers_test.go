package builtins

import (
	"math"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

func TestCompareNum_PrecisionAndSpecialValues(t *testing.T) {
	integer := func(value int64) Value { return &runtime.IntegerValue{Value: value} }
	float := func(value float64) Value { return &runtime.FloatValue{Value: value} }
	for _, tc := range []struct {
		a, b Value
		name string
		want int64
	}{
		{name: "adjacent large integers", a: integer(1<<53 + 1), b: integer(1 << 53), want: 1},
		{name: "adjacent large integers reversed", a: integer(1 << 53), b: integer(1<<53 + 1), want: -1},
		{name: "adjacent maximum integers", a: integer(math.MaxInt64), b: integer(math.MaxInt64 - 1), want: 1},
		{name: "adjacent minimum integers", a: integer(math.MinInt64), b: integer(math.MinInt64 + 1), want: -1},
		{name: "integer extremes", a: integer(math.MinInt64), b: integer(math.MaxInt64), want: -1},
		{name: "equal maximum integers", a: integer(math.MaxInt64), b: integer(math.MaxInt64), want: 0},
		{name: "mixed promotes to float", a: integer(1<<53 + 1), b: float(1 << 53), want: 0},
		{name: "mixed promotes to float reversed", a: float(1 << 53), b: integer(1<<53 + 1), want: 0},
		{name: "nan first", a: float(math.NaN()), b: float(1), want: 1},
		{name: "nan second", a: float(1), b: float(math.NaN()), want: 1},
		{name: "both nan", a: float(math.NaN()), b: float(math.NaN()), want: 1},
		{name: "nan and integer", a: float(math.NaN()), b: integer(1), want: 1},
		{name: "equal infinities", a: float(math.Inf(1)), b: float(math.Inf(1)), want: 0},
		{name: "opposite infinities", a: float(math.Inf(-1)), b: float(math.Inf(1)), want: -1},
		{name: "signed zeros", a: float(math.Copysign(0, -1)), b: float(0), want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := CompareNum(newMockContext(), []Value{tc.a, tc.b})
			value, ok := got.(*runtime.IntegerValue)
			if !ok || value.Value != tc.want {
				t.Fatalf("CompareNum(%v, %v) = %v; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestTestBit_Int64Boundaries(t *testing.T) {
	for _, tc := range []struct {
		name       string
		value, bit int64
		want       bool
	}{
		{"sign bit set", math.MinInt64, 63, true},
		{"sign bit clear", math.MaxInt64, 63, false},
		{"negative value", -1, 63, true},
		{"negative bit", -1, -1, false},
		{"past sign bit", -1, 64, false},
		{"minimum bit index", -1, math.MinInt64, false},
		{"maximum bit index", -1, math.MaxInt64, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := TestBit(newMockContext(), []Value{
				&runtime.IntegerValue{Value: tc.value},
				&runtime.IntegerValue{Value: tc.bit},
			})
			value, ok := got.(*runtime.BooleanValue)
			if !ok || value.Value != tc.want {
				t.Fatalf("TestBit(%d, %d) = %v; want %v", tc.value, tc.bit, got, tc.want)
			}
		})
	}
}
