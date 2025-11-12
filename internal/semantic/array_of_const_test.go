package semantic

import (
	"testing"
)

// TestArrayOfConstParameter tests that functions can have array of const parameters
func TestArrayOfConstParameter(t *testing.T) {
	input := `
		procedure Test(const items: array of const);
		begin
		end;
	`
	expectNoErrors(t, input)
}

// TestArrayOfConstInClassMethod tests array of const in class methods
func TestArrayOfConstInClassMethod(t *testing.T) {
	input := `
		type TTest = class
			procedure Log(const items: array of const);
		end;

		procedure TTest.Log(const items: array of const);
		begin
		end;
	`
	expectNoErrors(t, input)
}

// TestArrayOfConstLiteralHomogeneous tests homogeneous array literals with array of const
func TestArrayOfConstLiteralHomogeneous(t *testing.T) {
	input := `
		procedure Test(const items: array of const);
		begin
		end;

		begin
			Test([1, 2, 3]);
			Test(['a', 'b', 'c']);
		end.
	`
	expectNoErrors(t, input)
}

// TestArrayOfConstLiteralHeterogeneous tests heterogeneous array literals with array of const
func TestArrayOfConstLiteralHeterogeneous(t *testing.T) {
	input := `
		procedure Test(const items: array of const);
		begin
		end;

		begin
			Test([1, 'hello', 3.14, True]);
		end.
	`
	expectNoErrors(t, input)
}

// TestArrayOfConstEmptyLiteral tests empty array literals with array of const
func TestArrayOfConstEmptyLiteral(t *testing.T) {
	input := `
		procedure Test(const items: array of const);
		begin
		end;

		begin
			Test([]);
		end.
	`
	expectNoErrors(t, input)
}

// TestArrayOfConstTypeAlias tests array of const with type alias
func TestArrayOfConstTypeAlias(t *testing.T) {
	input := `
		type TArgs = array of const;

		procedure Test(const items: TArgs);
		begin
		end;

		begin
			Test([1, 2, 3]);
		end.
	`
	expectNoErrors(t, input)
}

// TestArrayOfConstVariantElementAccess tests that accessing array of const elements returns Variant
func TestArrayOfConstVariantElementAccess(t *testing.T) {
	input := `
		procedure Test(const items: array of const);
		var
			v: Variant;
		begin
			v := items[0];
		end;
	`
	expectNoErrors(t, input)
}

// TestArrayOfConstVariantToString tests that Variant from array of const can be assigned to String
func TestArrayOfConstVariantToString(t *testing.T) {
	input := `
		procedure Test(const items: array of const);
		var
			s: String;
		begin
			s := items[0];
		end;
	`
	expectNoErrors(t, input)
}

// TestArrayOfConstArrayCompatibility tests that different array types are compatible with array of const
func TestArrayOfConstArrayCompatibility(t *testing.T) {
	input := `
		procedure Test(const items: array of const);
		begin
		end;

		var
			intArray: array of Integer;
			strArray: array of String;
		begin
			intArray := [1, 2, 3];
			strArray := ['a', 'b'];

			// These should work due to array type compatibility
			// (array of Integer/String -> array of Variant)
			// Note: Direct array passing not yet implemented, using literals
			Test([1, 2, 3]);
			Test(['a', 'b']);
		end.
	`
	expectNoErrors(t, input)
}

// TestClassOperatorWithArrayOfConst tests class operators that use array of const
func TestClassOperatorWithArrayOfConst(t *testing.T) {
	input := `
		type TTest = class
			Field: String;
			procedure AppendStrings(const str: array of const);
			class operator += (const items: array of const) uses AppendStrings;
		end;

		procedure TTest.AppendStrings(const str: array of const);
		begin
		end;
	`
	expectNoErrors(t, input)
}

// TestClassOperatorCompoundAssignmentWithEmptyArray tests compound assignment with empty array
func TestClassOperatorCompoundAssignmentWithEmptyArray(t *testing.T) {
	input := `
		type TTest = class
			Field: String;
			procedure AppendStrings(const str: array of const);
			class operator += (const items: array of const) uses AppendStrings;
		end;

		procedure TTest.AppendStrings(const str: array of const);
		begin
		end;

		var t: TTest;
		begin
			t := TTest.Create;
			t += [];
			t += [1, 2];
			t += ['a', 'b', 'c'];
		end.
	`
	expectNoErrors(t, input)
}
