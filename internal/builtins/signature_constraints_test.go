package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

func TestBuiltinSignatures_CallConstraints(t *testing.T) {
	integerAlias := &types.TypeAlias{Name: "IntegerAlias", AliasedType: types.INTEGER}
	variantAlias := &types.TypeAlias{Name: "VariantAlias", AliasedType: types.VARIANT}
	integerRange := &types.SubrangeType{BaseType: types.INTEGER, LowBound: 1, HighBound: 3}
	enum := types.NewEnumType("TColor", map[string]int{"red": 0}, []string{"red"})
	tests := []struct {
		actual types.Type
		name   string
		index  int
		valid  bool
	}{
		{types.INTEGER, "Sqrt", 0, true},
		{types.INTEGER, "YearOf", 0, false},
		{types.FLOAT, "YearOf", 0, true},
		{&types.TypeAlias{Name: "DateAlias", AliasedType: types.FLOAT}, "YearOf", 0, false},
		{types.VARIANT, "VarIsEmpty", 0, true},
		{types.JSON_VARIANT, "VarIsEmpty", 0, true},
		{types.STRING, "VarIsEmpty", 0, false},
		{variantAlias, "VarIsEmpty", 0, false},
		{types.STRING, "ToJSON", 0, true},
		{variantAlias, "LeftStr", 0, true},
		{variantAlias, "LeftStr", 1, false},
		{integerAlias, "LeftStr", 1, false},
		{integerAlias, "FloatToStr", 0, true},
		{integerAlias, "FloatToStr", 1, true},
		{variantAlias, "FloatToStr", 1, false},
		{enum, "IntToStr", 0, true},
		{integerRange, "IntToStr", 0, true},
		{enum, "IntToStr", 1, false},
		{integerRange, "IntToHex", 0, true},
		{integerRange, "IntToHex", 1, true},
		{variantAlias, "IntToHex", 1, false},
		{types.STRING, "VarAsType", 1, true},
		{types.INTEGER, "VarAsType", 1, true},
		{types.FLOAT, "VarAsType", 1, false},
		{types.NewStaticArrayType(types.STRING, 1, 2), "StrJoin", 0, true},
		{types.NewDynamicArrayType(types.INTEGER), "StrJoin", 0, false},
		{integerAlias, "CompareNum", 0, true},
		{types.VARIANT, "CompareNum", 0, false},
	}
	for _, test := range tests {
		t.Run(test.name+"/"+test.actual.String(), func(t *testing.T) {
			sig, ok := DefaultRegistry.GetSignature(test.name)
			if !ok {
				t.Fatal("missing signature")
			}
			if got := sig.AcceptsArgument(test.index, test.actual); got != test.valid {
				t.Errorf("parameter %d accepts %v: got %t; want %t", test.index, test.actual, got, test.valid)
			}
		})
	}
}

func TestBuiltinSignatures_CorrectedShapes(t *testing.T) {
	tests := []struct {
		result        types.Type
		name          string
		validCounts   []int
		invalidCounts []int
	}{
		{types.STRING, "Trim", []int{1, 3}, []int{0, 2, 4}},
		{types.STRING, "TrimLeft", []int{1, 2}, []int{0, 3}},
		{types.FLOAT, "RandG", []int{0}, []int{1, 2}},
		{types.STRING, "StringReplace", []int{3, 4}, []int{0, 2, 5}},
		{types.STRING, "StrReplace", []int{3, 4}, []int{0, 2, 5}},
		{types.NewDynamicArrayType(types.STRING), "StrSplit", []int{2}, []int{0, 1, 3}},
		{types.NewDynamicArrayType(types.STRING), "JSONKeys", []int{1}, []int{0, 2}},
		{types.NewDynamicArrayType(types.VARIANT), "JSONValues", []int{1}, []int{0, 2}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sig, ok := DefaultRegistry.GetSignature(test.name)
			if !ok {
				t.Fatal("missing signature")
			}
			if !test.result.Equals(sig.ReturnType) {
				t.Errorf("result = %v; want %v", sig.ReturnType, test.result)
			}
			for _, count := range test.validCounts {
				if !sig.AcceptsArgCount(count) {
					t.Errorf("rejected %d arguments", count)
				}
			}
			for _, count := range test.invalidCounts {
				if sig.AcceptsArgCount(count) {
					t.Errorf("accepted %d arguments", count)
				}
			}
		})
	}
}
