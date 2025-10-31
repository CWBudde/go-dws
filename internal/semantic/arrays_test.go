package semantic

import "testing"

func TestArrayLiteralHomogeneousIntegers(t *testing.T) {
	input := `
		type TIntArray = array of Integer;
		var arr: TIntArray;
		var first: Integer;
		var arrEmpty: TIntArray;
		begin
			arr := [1, 2, 3];
			arrEmpty := [];
			first := arr[0];
		end.
	`
	expectNoErrors(t, input)
}

func TestArrayLiteralTypePromotion(t *testing.T) {
	input := `
		type TFloatArray = array of Float;
		var arr: TFloatArray;
		var value: Float;
		begin
			arr := [1, 2.5, 3];
			value := arr[1];
		end.
	`
	expectNoErrors(t, input)
}

func TestArrayLiteralMixedTypesError(t *testing.T) {
	input := `
		type TIntArray = array of Integer;
		var bad: TIntArray;
		begin
			bad := [1, 'hello'];
		end.
	`
	expectError(t, input, "expected Integer")
}

func TestArrayLiteralNestedArrays(t *testing.T) {
	input := `
		type TRow = array of Integer;
		type TMatrix = array of TRow;
		var matrix: TMatrix;
		var result: Integer;
		begin
			matrix := [[1, 2], [3, 4]];
			result := matrix[0][1];
		end.
	`
	expectNoErrors(t, input)
}

func TestArrayLiteralEmptyWithoutContext(t *testing.T) {
	input := `
		begin
			[];
		end.
	`
	expectError(t, input, "cannot infer type for empty array literal")
}
