package interp

import (
	"testing"
)

// TestRecordReturnTypeInitialization verifies that record return types are properly
// initialized when using ExecuteUserFunction.
func TestRecordReturnTypeInitialization(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "Direct record return",
			script: `
				type TPoint = record
					X, Y: Integer;
				end;

				function MakePoint(a: Integer): TPoint;
				begin
					Result.X := a;
					Result.Y := a + 1;
				end;

				var p: TPoint;
				begin
					p := MakePoint(10);
					PrintLn(p.X);
					PrintLn(p.Y);
				end
			`,
			expected: "10\n11\n",
		},
		{
			name: "Record return via implicit conversion",
			script: `
				type TFoo = record
					Value: Integer;
				end;

				function IntToFoo(i: Integer): TFoo;
				begin
					Result.Value := i * 2;
				end;

				operator implicit (Integer): TFoo uses IntToFoo;

				var f: TFoo;
				begin
					f := 21;  // Implicit conversion
					PrintLn(f.Value);
				end
			`,
			expected: "42\n",
		},
		{
			name: "Nested record return",
			script: `
				type TInner = record
					Val: Integer;
				end;

				type TOuter = record
					Inner: TInner;
					Extra: Integer;
				end;

				function MakeOuter(x: Integer): TOuter;
				begin
					Result.Inner.Val := x;
					Result.Extra := x + 100;
				end;

				var o: TOuter;
				begin
					o := MakeOuter(5);
					PrintLn(o.Inner.Val);
					PrintLn(o.Extra);
				end
			`,
			expected: "5\n105\n",
		},
		{
			name: "Record with multiple fields",
			script: `
				type TComplex = record
					A, B, C: Integer;
					S: String;
				end;

				function MakeComplex(): TComplex;
				begin
					Result.A := 1;
					Result.B := 2;
					Result.C := 3;
					Result.S := 'test';
				end;

				var c: TComplex;
				begin
					c := MakeComplex();
					PrintLn(c.A);
					PrintLn(c.B);
					PrintLn(c.C);
					PrintLn(c.S);
				end
			`,
			expected: "1\n2\n3\ntest\n",
		},
		{
			name: "Record return assigned to function name",
			script: `
				type TValue = record
					N: Integer;
				end;

				function GetValue(x: Integer): TValue;
				begin
					GetValue.N := x * 10;
				end;

				var v: TValue;
				begin
					v := GetValue(7);
					PrintLn(v.N);
				end
			`,
			expected: "70\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.script)
			if isError(result) {
				t.Fatalf("evaluation error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestRecordReturnViaFunctionPointer verifies record return types work with function pointers.
func TestRecordReturnViaFunctionPointer(t *testing.T) {
	input := `
		type TPoint = record
			X, Y: Integer;
		end;

		function CreatePoint(a, b: Integer): TPoint;
		begin
			Result.X := a;
			Result.Y := b;
		end;

		type TPointFunc = function(a, b: Integer): TPoint;

		var fn: TPointFunc;
		var p: TPoint;
		begin
			fn := CreatePoint;
			p := fn(5, 7);
			PrintLn(p.X);
			PrintLn(p.Y);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "5\n7\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

// TestRecordReturnViaOverloadedFunction verifies record return types work with overloaded functions.
func TestRecordReturnViaOverloadedFunction(t *testing.T) {
	input := `
		type TPair = record
			First, Second: Integer;
		end;

		function MakePair(a: Integer): TPair; overload;
		begin
			Result.First := a;
			Result.Second := a;
		end;

		function MakePair(a, b: Integer): TPair; overload;
		begin
			Result.First := a;
			Result.Second := b;
		end;

		var p1, p2: TPair;
		begin
			p1 := MakePair(10);
			p2 := MakePair(20, 30);
			PrintLn(p1.First);
			PrintLn(p1.Second);
			PrintLn(p2.First);
			PrintLn(p2.Second);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "10\n10\n20\n30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
