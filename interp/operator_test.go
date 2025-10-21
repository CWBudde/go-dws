package interp

import "testing"

func TestGlobalOperatorOverload(t *testing.T) {
	input := `
		function StrPlusInt(s: String; i: Integer): String;
		begin
			Result := s + '[value]';
		end;

		operator + (String, Integer) : String uses StrPlusInt;

		var s := 'abc' + 42;
		PrintLn(s);
	`

	_, output := testEvalWithOutput(input)

	if output != "abc[value]\n" {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestClassOperatorOverload(t *testing.T) {
	input := `
type TTest = class
  Field: String;
  constructor Create;
  function AddString(str: String): TTest;
  class operator + (TTest, String) : TTest uses AddString;
end;

var t: TTest;

constructor TTest.Create;
begin
	Field := '';
end;

function TTest.AddString(str: String): TTest;
begin
	Field := Field + str + ',';
	Result := Self;
end;

begin
	t := TTest.Create();
	t.Field := 'first,';
	t := t + 'second';
	PrintLn(t.Field);
end
`

	_, output := testEvalWithOutput(input)

	if output != "first,second,\n" {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestClassOperatorIn(t *testing.T) {
	input := `
		type TMyRange = class
		  FMin, FMax: Integer;
		  constructor Create(min, max: Integer);
		  function Contains(i: Integer): Boolean;
		  class operator IN (Integer) : Boolean uses Contains;
		end;

		var range: TMyRange;

		constructor TMyRange.Create(min, max: Integer);
		begin
			FMin := min;
			FMax := max;
		end;

		function TMyRange.Contains(i: Integer): Boolean;
		begin
			Result := (i >= FMin) and (i <= FMax);
		end;

		begin
			range := TMyRange.Create(1, 5);
			PrintLn(3 in range);
			PrintLn(7 in range);
		end
	`

	_, output := testEvalWithOutput(input)

	if output != "true\nfalse\n" {
		t.Fatalf("unexpected output: %s", output)
	}
}
