package interp

import (
	"testing"
)

// TestArrayReturnTypeInitialization verifies that array return types are properly
// initialized when using ExecuteUserFunction.
func TestArrayReturnTypeInitialization(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "Array of Integer with Result.Add",
			script: `
				type TIntArray = array of Integer;

				function BuildArray(): TIntArray;
				begin
					Result.Add(10);
					Result.Add(20);
					Result.Add(30);
				end;

				var arr := BuildArray();
				PrintLn(Length(arr));
				PrintLn(arr[0]);
				PrintLn(arr[1]);
				PrintLn(arr[2]);
			`,
			expected: "3\n10\n20\n30\n",
		},
		{
			name: "Array of String with Result.Add",
			script: `
				type TStrArray = array of String;

				function BuildStrings(): TStrArray;
				begin
					Result.Add('hello');
					Result.Add('world');
				end;

				var arr := BuildStrings();
				PrintLn(Length(arr));
				PrintLn(arr[0]);
				PrintLn(arr[1]);
			`,
			expected: "2\nhello\nworld\n",
		},
		{
			name: "Array of Float with multiple adds",
			script: `
				function GetFloats(): array of Float;
				var i: Integer;
				begin
					for i := 1 to 4 do
						Result.Add(i * 1.5);
				end;

				var f := GetFloats();
				PrintLn(f.High);
				PrintLn(f[0]);
				PrintLn(f[3]);
			`,
			expected: "3\n1.5\n6\n",
		},
		{
			name: "Inline array of Integer return type",
			script: `
				function GetNumbers(): array of Integer;
				begin
					Result.Add(100);
					Result.Add(200);
				end;

				var nums := GetNumbers();
				PrintLn(nums.High);
				PrintLn(nums[0] + nums[1]);
			`,
			expected: "1\n300\n",
		},
		{
			name: "Named array type with record elements",
			script: `
				type TPoint = record
					X, Y: Integer;
				end;

				type TPointArray = array of TPoint;

				function MakePoints(): TPointArray;
				var p1, p2: TPoint;
				begin
					p1.X := 1;
					p1.Y := 2;
					p2.X := 3;
					p2.Y := 4;
					Result.Add(p1);
					Result.Add(p2);
				end;

				var pts := MakePoints();
				PrintLn(Length(pts));
				PrintLn(pts[0].X);
				PrintLn(pts[1].Y);
			`,
			expected: "2\n1\n4\n",
		},
		{
			name: "Function returns local array assigned to Result",
			script: `
				function BuildLocalArray(): array of Integer;
				var localArr: array of Integer;
				begin
					localArr.Add(5);
					localArr.Add(6);
					localArr.Add(7);
					Result := localArr;
				end;

				var arr := BuildLocalArray();
				PrintLn(arr[0]);
				PrintLn(arr[1]);
				PrintLn(arr[2]);
			`,
			expected: "5\n6\n7\n",
		},
		{
			name: "Empty array returned",
			script: `
				function GetEmpty(): array of Integer;
				begin
					// Don't add anything
				end;

				var arr := GetEmpty();
				PrintLn(Length(arr));
				PrintLn(arr.High);
			`,
			expected: "0\n-1\n",
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

// TestArrayReturnViaFunctionPointer verifies array return types work with function pointers.
func TestArrayReturnViaFunctionPointer(t *testing.T) {
	input := `
		type TIntArray = array of Integer;

		function CreateArray(count: Integer): TIntArray;
		var i: Integer;
		begin
			for i := 1 to count do
				Result.Add(i * 10);
		end;

		type TArrayFunc = function(count: Integer): TIntArray;

		var fn: TArrayFunc;
		var arr: TIntArray;
		begin
			fn := CreateArray;
			arr := fn(3);
			PrintLn(Length(arr));
			PrintLn(arr[0]);
			PrintLn(arr[1]);
			PrintLn(arr[2]);
		end
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "3\n10\n20\n30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

// TestArrayReturnViaOverloadedFunction verifies array return types work with overloaded functions.
func TestArrayReturnViaOverloadedFunction(t *testing.T) {
	input := `
		function MakeArray(n: Integer): array of Integer; overload;
		begin
			Result.Add(n);
		end;

		function MakeArray(a, b: Integer): array of Integer; overload;
		begin
			Result.Add(a);
			Result.Add(b);
		end;

		function MakeArray(a, b, c: Integer): array of Integer; overload;
		begin
			Result.Add(a);
			Result.Add(b);
			Result.Add(c);
		end;

		var arr1 := MakeArray(100);
		var arr2 := MakeArray(10, 20);
		var arr3 := MakeArray(1, 2, 3);

		PrintLn(arr1.High);
		PrintLn(arr2.High);
		PrintLn(arr3.High);
		PrintLn(arr1[0]);
		PrintLn(arr2[0] + arr2[1]);
		PrintLn(arr3[0] + arr3[1] + arr3[2]);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "0\n1\n2\n100\n30\n6\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

// TestInlineArrayOfRecordReturnType tests array of record return types.
func TestInlineArrayOfRecordReturnType(t *testing.T) {
	// Note: Using separate record variables to avoid record reference aliasing issue
	// (records passed to .Add() share the same reference - known behavior)
	input := `
		type TData = record
			Name: String;
			Value: Integer;
		end;

		function GetData(): array of TData;
		var d1, d2: TData;
		begin
			d1.Name := 'first';
			d1.Value := 1;
			Result.Add(d1);

			d2.Name := 'second';
			d2.Value := 2;
			Result.Add(d2);
		end;

		var items := GetData();
		PrintLn(Length(items));
		PrintLn(items[0].Name);
		PrintLn(items[0].Value);
		PrintLn(items[1].Name);
		PrintLn(items[1].Value);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}
	expected := "2\nfirst\n1\nsecond\n2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
