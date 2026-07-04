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
		name     string
		left     *types.ArrayType
		right    *types.ArrayType
		wantElem types.Type
	}{
		{"both nil", nil, nil, types.VARIANT},
		{"left nil", nil, intArr, types.INTEGER},
		{"right nil", intArr, nil, types.INTEGER},
		{"same element", intArr, intArr, types.INTEGER},
		{"static+dynamic same element", staticInt, intArr, types.INTEGER},
		{"variant left wins", varArr, strArr, types.VARIANT},
		{"variant right wins", intArr, varArr, types.VARIANT},
		{"mixed elements widen to variant", intArr, strArr, types.VARIANT},
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
		name     string
		arr      *runtime.ArrayValue
		index    int
		wantPhys int
		wantErr  string
	}{
		{"dynamic in range", dyn, 1, 1, ""},
		{"dynamic below", dyn, -1, 0, "Lower bound exceeded! Index -1"},
		{"dynamic above", dyn, 2, 0, "Upper bound exceeded! Index 2"},
		{"static low bound maps to zero", static, 5, 0, ""},
		{"static high bound", static, 6, 1, ""},
		{"static below", static, 4, 0, "Lower bound exceeded! Index 4"},
		{"static above", static, 7, 0, "Upper bound exceeded! Index 7"},
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
