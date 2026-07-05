package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

func TestConcatArrayResultType(t *testing.T) {
	intArr := types.NewDynamicArrayType(types.INTEGER)
	strArr := types.NewDynamicArrayType(types.STRING)
	varArr := types.NewDynamicArrayType(types.VARIANT)
	staticInt := types.NewStaticArrayType(types.INTEGER, 0, 2)

	tests := []struct {
		left     *types.ArrayType
		right    *types.ArrayType
		wantElem types.Type
		name     string
	}{
		{nil, nil, types.VARIANT, "both nil"},
		{nil, intArr, types.INTEGER, "left nil"},
		{intArr, nil, types.INTEGER, "right nil"},
		{intArr, intArr, types.INTEGER, "same element"},
		{staticInt, intArr, types.INTEGER, "static+dynamic same element"},
		{varArr, strArr, types.VARIANT, "variant left wins"},
		{intArr, varArr, types.VARIANT, "variant right wins"},
		{intArr, strArr, types.VARIANT, "mixed elements widen to variant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := concatArrayResultType(tt.left, tt.right)
			if got == nil || !got.IsDynamic() {
				t.Fatalf("expected dynamic array result, got %v", got)
			}
			if !got.ElementType.Equals(tt.wantElem) {
				t.Errorf("element type = %s, want %s", got.ElementType.String(), tt.wantElem.String())
			}
		})
	}
}

func TestStringToBoolCast(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"true", true},
		{"True", true},
		{"T", true},
		{"yes", true},
		{"y", true},
		{"1", true},
		{"-3.5", true},
		{"0", false},
		{"0.0", false},
		{"false", false},
		{"b", false},
		{"", false},
		{"  true  ", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := stringToBoolCast(tt.in); got != tt.want {
				t.Errorf("stringToBoolCast(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestArrayElementPhysicalIndex(t *testing.T) {
	dyn := &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(types.INTEGER),
		Elements: []runtime.Value{
			&runtime.IntegerValue{Value: 10},
			&runtime.IntegerValue{Value: 20},
		},
	}
	static := &runtime.ArrayValue{
		ArrayType: types.NewStaticArrayType(types.INTEGER, 5, 6),
		Elements: []runtime.Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
		},
	}

	tests := []struct {
		arr      *runtime.ArrayValue
		wantErr  string
		name     string
		index    int
		wantPhys int
	}{
		{dyn, "", "dynamic in range", 1, 1},
		{dyn, "Lower bound exceeded! Index -1", "dynamic below", -1, 0},
		{dyn, "Upper bound exceeded! Index 2", "dynamic above", 2, 0},
		{static, "", "static low bound maps to zero", 5, 0},
		{static, "", "static high bound", 6, 1},
		{static, "Lower bound exceeded! Index 4", "static below", 4, 0},
		{static, "Upper bound exceeded! Index 7", "static above", 7, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phys, err := arrayElementPhysicalIndex(tt.arr, tt.index)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if phys != tt.wantPhys {
					t.Errorf("physical index = %d, want %d", phys, tt.wantPhys)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got none", tt.wantErr)
			}
			if _, ok := err.(boundExceededError); !ok {
				t.Errorf("expected boundExceededError, got %T", err)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestOrdinalValueLike(t *testing.T) {
	tests := []struct {
		bound Value
		want  Value
		name  string
		ord   int
	}{
		{&runtime.StringValue{Value: "a"}, &runtime.StringValue{Value: "b"}, "string bound yields char string", 'b'},
		{&runtime.BooleanValue{Value: false}, &runtime.BooleanValue{Value: true}, "boolean bound yields boolean", 1},
		{&runtime.BooleanValue{Value: false}, &runtime.BooleanValue{Value: false}, "boolean zero yields False", 0},
		{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 7}, "integer bound yields integer", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ordinalValueLike(tt.bound, tt.ord)
			if got.String() != tt.want.String() || got.Type() != tt.want.Type() {
				t.Errorf("ordinalValueLike(%v, %d) = %v (%s), want %v (%s)",
					tt.bound, tt.ord, got, got.Type(), tt.want, tt.want.Type())
			}
		})
	}
}
