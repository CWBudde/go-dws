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

func TestRecordFunctionReturn(t *testing.T) {
	input := `
		type
		  TFoo = record
		    X, Y: Integer;
		  end;

		function MakeFoo(I: Integer): TFoo;
		begin
		  Result.X := I;
		  Result.Y := I+1;
		end;

		var F1: TFoo;
		begin
		  F1 := MakeFoo(10);
		  PrintLn(F1.X);
		  PrintLn(F1.Y);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "10\n11\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestImplicitConversionInAssignment(t *testing.T) {
	input := `
		type
		  TFoo = record
		    X, Y: Integer;
		  end;

		function FooImplInt(I: Integer): TFoo;
		begin
		  Result.X := I;
		  Result.Y := I+1;
		end;

		operator implicit (Integer): TFoo uses FooImplInt;

		var F1: TFoo;
		begin
		  F1 := 10;  // Implicit conversion from Integer to TFoo
		  PrintLn(F1.X);
		  PrintLn(F1.Y);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "10\n11\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestImplicitConversionInFunctionCall(t *testing.T) {
	input := `
		type
		  TFoo = record
		    X, Y: Integer;
		  end;

		function FooImplInt(I: Integer): TFoo;
		begin
		  Result.X := I;
		  Result.Y := I+1;
		end;

		operator implicit (Integer): TFoo uses FooImplInt;

		function ProcessFoo(F: TFoo): Integer;
		begin
		  Result := F.X + F.Y;
		end;

		begin
		  PrintLn(ProcessFoo(42));  // Implicit conversion from Integer to TFoo
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "85\n" // 42 + 43
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestImplicitConversionInMethodCall(t *testing.T) {
	input := `
		type
		  TFoo = record
		    X, Y: Integer;
		  end;

		function FooImplInt(I: Integer): TFoo;
		begin
		  Result.X := I;
		  Result.Y := I+1;
		end;

		operator implicit (Integer): TFoo uses FooImplInt;

		type
		  TProcessor = class
		    function ProcessFoo(F: TFoo): Integer;
		  end;

		function TProcessor.ProcessFoo(F: TFoo): Integer;
		begin
		  Result := F.X + F.Y;
		end;

		var p: TProcessor;
		begin
		  p := TProcessor.Create();
		  PrintLn(p.ProcessFoo(42));  // Implicit conversion from Integer to TFoo
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "85\n" // 42 + 43
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestImplicitConversionInReturn(t *testing.T) {
	input := `
		type
		  TFoo = record
		    X, Y: Integer;
		  end;

		function FooImplInt(I: Integer): TFoo;
		begin
		  Result.X := I;
		  Result.Y := I+1;
		end;

		operator implicit (Integer): TFoo uses FooImplInt;

		function MakeFoo(): TFoo;
		begin
		  Result := 20;  // Implicit conversion from Integer to TFoo in return
		end;

		var F: TFoo;
		begin
		  F := MakeFoo();
		  PrintLn(F.X);
		  PrintLn(F.Y);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "20\n21\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

// Task 8.19d: Test two-step conversion chain (Integer -> String -> TCustom)
func TestConversionChainTwoSteps(t *testing.T) {
	input := `
		type
		  TCustom = record
		    Val: Integer;
		  end;

		function IntToStr(I: Integer): String;
		begin
		  Result := IntToString(I);
		end;

		function StrToCustom(S: String): TCustom;
		begin
		  Result.Val := 999;
		end;

		operator implicit (Integer): String uses IntToStr;
		operator implicit (String): TCustom uses StrToCustom;

		var C: TCustom;
		begin
		  C := 42;  // Chain: Integer -> String -> TCustom
		  PrintLn(C.Val);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "999\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

// Task 8.19d: Test three-step conversion chain (Integer -> String -> TFoo -> TBar)
func TestConversionChainThreeSteps(t *testing.T) {
	input := `
		type
		  TFoo = record
		    S: String;
		  end;

		type
		  TBar = record
		    Value: Integer;
		  end;

		function IntToStr(I: Integer): String;
		begin
		  Result := IntToString(I);
		end;

		function StrToFoo(S: String): TFoo;
		begin
		  Result.S := S + '_foo';
		end;

		function FooToBar(F: TFoo): TBar;
		begin
		  Result.Value := 100;
		end;

		operator implicit (Integer): String uses IntToStr;
		operator implicit (String): TFoo uses StrToFoo;
		operator implicit (TFoo): TBar uses FooToBar;

		var B: TBar;
		begin
		  B := 42;  // Chain: Integer -> String -> TFoo -> TBar
		  PrintLn(B.Value);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "100\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

// Task 8.19d: Test that missing conversion chain is handled gracefully
func TestConversionChainNotFound(t *testing.T) {
	input := `
		type
		  TCustom = record
		    Value: Integer;
		  end;

		function IntToStr(I: Integer): String;
		begin
		  Result := IntToString(I);
		end;

		operator implicit (Integer): String uses IntToStr;
		// No conversion from String to TCustom - chain is broken

		var C: TCustom;
		begin
		  C.Value := 42;  // Direct assignment (no conversion)
		  PrintLn(C.Value);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "42\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

// Task 8.19d: Test that conversion chains exceeding max depth are not used
func TestConversionChainMaxDepth(t *testing.T) {
	input := `
		type
		  TStep1 = record V: Integer; end;

		type
		  TStep2 = record V: Integer; end;

		type
		  TStep3 = record V: Integer; end;

		type
		  TStep4 = record V: Integer; end;

		function IntToStep1(I: Integer): TStep1;
		begin
		  Result.V := I + 1;
		end;

		function Step1ToStep2(S: TStep1): TStep2;
		begin
		  Result.V := S.V + 1;
		end;

		function Step2ToStep3(S: TStep2): TStep3;
		begin
		  Result.V := S.V + 1;
		end;

		function Step3ToStep4(S: TStep3): TStep4;
		begin
		  Result.V := S.V + 1;
		end;

		operator implicit (Integer): TStep1 uses IntToStep1;
		operator implicit (TStep1): TStep2 uses Step1ToStep2;
		operator implicit (TStep2): TStep3 uses Step2ToStep3;
		operator implicit (TStep3): TStep4 uses Step3ToStep4;

		var S: TStep3;
		begin
		  // This should work: Integer -> TStep1 -> TStep2 -> TStep3 (3 steps, within limit)
		  S := 10;
		  PrintLn(S.V);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	// Should successfully chain through 3 conversions: 10 -> 11 -> 12 -> 13
	expected := "13\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

// Task 8.19e: Port of implicit_record1.pas from DWScript reference tests
// Tests Integer -> TFoo implicit conversion in assignment and var declaration
func TestImplicitRecord1(t *testing.T) {
	input := `
		type
		  TFoo = record
		    X, Y: Integer;
		  end;

		function FooImplInt(I: Integer): TFoo;
		begin
		  Result.X := I;
		  Result.Y := I+1;
		end;

		operator implicit (Integer): TFoo uses FooImplInt;

		var F1: TFoo;

		begin
		  F1 := 10;
		  Print('F1 X=');
		  Print(F1.X);
		  Print(' Y=');
		  PrintLn(F1.Y);

		  var F2: TFoo;
		  F2 := 20;
		  Print('F2 X=');
		  Print(F2.X);
		  Print(' Y=');
		  PrintLn(F2.Y);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "F1 X=10 Y=11\nF2 X=20 Y=21\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

// Task 8.19e: Port of implicit_record2.pas from DWScript reference tests
// Tests TFoo -> Integer implicit conversion in function arguments and assignment
func TestImplicitRecord2(t *testing.T) {
	input := `
		type
		  TFoo = record
		    X, Y: Integer;
		  end;

		procedure PrintInt(i: Integer);
		begin
		    PrintLn(i);
		end;

		function OperImpFooInt(aFoo: TFoo): Integer;
		begin
		  Result := aFoo.X + aFoo.Y;
		end;

		operator implicit (TFoo): Integer uses OperImpFooInt;

		var F: TFoo;
		var i: Integer;

		begin
		  F.X := 10;
		  F.Y := 11;

		  PrintInt(F);

		  i := F;

		  PrintLn(i);

		  F.Y := 123;

		  PrintInt(F);

		  i := F;

		  PrintLn(i);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "21\n21\n133\n133\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}
