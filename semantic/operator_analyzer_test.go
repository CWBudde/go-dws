package semantic

import "testing"

func TestGlobalOperatorOverload(t *testing.T) {
	input := `
		function StrPlusInt(s: String; i: Integer): String;
		begin
			Result := s;
		end;

		operator + (String, Integer) : String uses StrPlusInt;

		var result := 'abc' + 1;
	`
	expectNoErrors(t, input)
}

func TestImplicitConversionOperator(t *testing.T) {
	input := `
	function IntToStr(i: Integer): String;
	begin
		Result := 'value';
	end;

		operator implicit (Integer) : String uses IntToStr;

		var s: String;
		s := 123;
	`
	expectNoErrors(t, input)
}

func TestDuplicateGlobalOperator(t *testing.T) {
	input := `
		function StrPlusInt(s: String; i: Integer): String;
		begin
			Result := s;
		end;

		operator + (String, Integer) : String uses StrPlusInt;
		operator + (String, Integer) : String uses StrPlusInt;
	`
	expectError(t, input, "already defined")
}

// Task 8.22d: Test class operator overload validation
func TestClassOperatorOverload(t *testing.T) {
	input := `
		type TTest = class
			Field: String;
			constructor Create;
			function AddString(str: String): TTest;
			class operator + (TTest, String) : TTest uses AddString;
		end;

		constructor TTest.Create;
		begin
			Field := '';
		end;

		function TTest.AddString(str: String): TTest;
		begin
			Field := Field + str;
			Result := Self;
		end;

		var t: TTest;
		begin
			t := TTest.Create();
			t := t + 'test';
		end
	`
	expectNoErrors(t, input)
}

// Task 8.22e: Test operator signature mismatch errors
func TestOperatorSignatureMismatch(t *testing.T) {
	// Test 1: Wrong number of parameters
	t.Run("wrong parameter count", func(t *testing.T) {
		input := `
			function WrongParams(s: String): String;
			begin
				Result := s;
			end;

			operator + (String, Integer) : String uses WrongParams;
		`
		expectError(t, input, "expects 2 parameters")
	})

	// Test 2: Wrong parameter types
	t.Run("wrong parameter types", func(t *testing.T) {
		input := `
			function WrongTypes(i: Integer; s: String): String;
			begin
				Result := s;
			end;

			operator + (String, Integer) : String uses WrongTypes;
		`
		expectError(t, input, "does not match operator operand type")
	})

	// Test 3: Wrong return type (for class operators)
	t.Run("wrong return type class operator", func(t *testing.T) {
		input := `
			type TTest = class
				Field: String;
				function WrongReturn(str: String): Integer;
				class operator + (TTest, String) : TTest uses WrongReturn;
			end;

			function TTest.WrongReturn(str: String): Integer;
			begin
				Result := 0;
			end;
		`
		expectError(t, input, "return type")
	})
}

// Task 8.22f: Test invalid binding function (not found, wrong signature)
func TestInvalidBindingFunction(t *testing.T) {
	// Test 1: Binding function not found
	t.Run("binding not found", func(t *testing.T) {
		input := `
			operator + (String, Integer) : String uses NonExistentFunction;
		`
		expectError(t, input, "not found")
	})

	// Test 2: Binding is not a function (bound to a variable)
	t.Run("binding not a function", func(t *testing.T) {
		input := `
			var NotAFunction: String;

			operator + (String, Integer) : String uses NotAFunction;
		`
		expectError(t, input, "not a function")
	})

	// Test 3: Class operator binding method not found
	t.Run("class operator method not found", func(t *testing.T) {
		input := `
			type TTest = class
				Field: String;
				class operator + (TTest, String) : TTest uses NonExistentMethod;
			end;
		`
		expectError(t, input, "not found")
	})
}
