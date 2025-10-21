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
