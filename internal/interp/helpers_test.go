package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Helper Method Tests
// ============================================================================

func TestHelperMethod(t *testing.T) {
	input := `
		type TStringHelper = helper for String
			function ToUpperCase: String;
			begin
				// Simplified version - just return a test value
				Result := 'HELLO';
			end;
		end;

		var s: String;
		begin
			s := 'hello';
			PrintLn(s.ToUpperCase());
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "HELLO\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestHelperMethodWithParameters(t *testing.T) {
	input := `
		type TStringHelper = helper for String
			function Contains(substring: String): Boolean;
			begin
				// Simplified version - just check equality
				Result := (Self = substring);
			end;
		end;

		var s: String;
		var found: Boolean;
		begin
			s := 'hello';
			found := s.Contains('hello');
			PrintLn(found);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "True\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestHelperMethodOnInteger(t *testing.T) {
	input := `
		type TIntegerHelper = helper for Integer
			function IsEven: Boolean;
			begin
				// Use mod operator to check if even
				Result := (Self mod 2 = 0);
			end;
		end;

		var n: Integer;
		begin
			n := 42;
			PrintLn(n.IsEven());
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "True\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestHelperMethodOnRecord(t *testing.T) {
	input := `
		type TPoint = record
			X: Integer;
			Y: Integer;
		end;

		type TPointHelper = record helper for TPoint
			function Sum: Integer;
			begin
				Result := Self.X + Self.Y;
			end;
		end;

		var p: TPoint;
		begin
			p := (X: 10, Y: 20);
			PrintLn(p.Sum());
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "30\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestMultipleHelpersForSameType(t *testing.T) {
	input := `
		type TStringHelper1 = helper for String
			function ToUpper: String;
			begin
				Result := 'UPPER';
			end;
		end;

		type TStringHelper2 = helper for String
			function ToLower: String;
			begin
				Result := 'lower';
			end;
		end;

		var s: String;
		begin
			s := 'Hello';
			PrintLn(s.ToUpper());
			PrintLn(s.ToLower());
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "UPPER\nlower\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestIntrinsicIntegerBooleanHelpers(t *testing.T) {
	input := `
	var i := 42;
	PrintLn(i.ToString);
	PrintLn(i.ToString());
	var b := False;
	PrintLn(b.ToString);
	PrintLn(b.ToString());
	var arr := ['a', 'b'];
	PrintLn(arr.Join('-'));
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "42\n42\nFalse\nFalse\na-b\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestHelperPrecedence(t *testing.T) {
	input := `
	Type
		THelper1 = Helper For String
			function Label: String;
		Begin
			Result := 'H1-' + Self;
		End;
	End;

	var s := 'x';
	PrintLn(s.Label());

	Type
		THelper2 = Helper For String
			function Label: String;
		Begin
			Result := 'H2-' + Self;
		End;
	End;

	PrintLn(s.Label());
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "H1-x\nH2-x\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestIntrinsicFloatHelpers(t *testing.T) {
	input := `
	var f := 3.14159;
	PrintLn(f.ToString);
	PrintLn(f.ToString(2));
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "3.14159\n3.14\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Helper Property Tests
// ============================================================================

func TestHelperProperty(t *testing.T) {
	input := `
		type TPoint = record
			X: Integer;
			Y: Integer;
		end;

		type TPointHelper = record helper for TPoint
			property Sum: Integer read GetSum;

			function GetSum: Integer;
			begin
				Result := Self.X + Self.Y;
			end;
		end;

		var p: TPoint;
		var sum: Integer;
		begin
			p := (X: 5, Y: 7);
			sum := p.Sum;
			PrintLn(sum);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "12\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestHelperPropertyOnBasicType(t *testing.T) {
	input := `
		type TStringHelper = helper for String
			property Length: Integer read GetLength;

			function GetLength: Integer;
			begin
				// Simplified - just return a fixed value
				Result := 5;
			end;
		end;

		var s: String;
		var len: Integer;
		begin
			s := 'hello';
			len := s.Length;
			PrintLn(len);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "5\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Helper Class Vars/Consts Tests
// ============================================================================

func TestHelperClassConst(t *testing.T) {
	input := `
		type TMathHelper = helper for Float
			class const PI = 3.14159;

			function TimesPi: Float;
			begin
				Result := Self * PI;
			end;
		end;

		var radius: Float;
		var circumference: Float;
		begin
			radius := 2.0;
			circumference := radius.TimesPi();
			PrintLn(circumference);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	// Should output approximately 6.28318
	output := out.String()
	if output != "6.28318\n" {
		t.Errorf("wrong output. expected=~6.28318, got=%q", output)
	}
}

func TestHelperClassVar(t *testing.T) {
	input := `
		type TIntegerHelper = helper for Integer
			class var DefaultValue: Integer;

			function GetDefault: Integer;
			begin
				Result := DefaultValue;
			end;
		end;

		var n: Integer;
		begin
			// Note: Class vars would need proper initialization syntax
			// For now, test that it doesn't crash
			n := 42;
			PrintLn(n.GetDefault());
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	// Class var should be initialized to 0
	expected := "0\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Helper Syntax Variations Tests
// ============================================================================

func TestHelperWithoutRecordKeyword(t *testing.T) {
	input := `
		type TStringHelper = helper for String
			function Test: String;
			begin
				Result := 'TEST';
			end;
		end;

		var s: String;
		begin
			s := 'hello';
			PrintLn(s.Test());
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "TEST\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestHelperWithRecordKeyword(t *testing.T) {
	input := `
		type TPoint = record
			X: Integer;
			Y: Integer;
		end;

		type TPointHelper = record helper for TPoint
			function IsOrigin: Boolean;
			begin
				Result := (Self.X = 0) and (Self.Y = 0);
			end;
		end;

		var p: TPoint;
		begin
			p := (X: 0, Y: 0);
			PrintLn(p.IsOrigin());
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "True\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Helper Error Cases
// ============================================================================

func TestHelperMethodNotFound(t *testing.T) {
	input := `
		type TStringHelper = helper for String
			function ToUpper: String;
			begin
				Result := 'UPPER';
			end;
		end;

		var s: String;
		begin
			s := 'hello';
			s.ToMegaCase();  // Method doesn't exist
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if !isError(result) {
		t.Fatal("expected error for non-existent helper method")
	}

	errMsg := result.String()
	if !containsSubstring(errMsg, "no helper found") && !containsSubstring(errMsg, "not found") {
		t.Errorf("wrong error message: %s", errMsg)
	}
}

func TestHelperOnClassInstancePrefersMethods(t *testing.T) {
	input := `
		type TMyClass = class
			function Test: String;
			begin
				Result := 'CLASS METHOD';
			end;
		end;

		type TMyClassHelper = helper for TMyClass
			function Test: String;
			begin
				Result := 'HELPER METHOD';
			end;
		end;

		var obj: TMyClass;
		begin
			obj := TMyClass.Create();
			PrintLn(obj.Test());  // Should call class method, not helper
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	// Instance method should take precedence over helper
	expected := "CLASS METHOD\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// interpret is a helper function to parse and evaluate DWScript code
func interpret(interp *Interpreter, input string) Value {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return newError("parser errors: %v", p.Errors())
	}

	return interp.Eval(program)
}

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// ============================================================================
// Array Helper Properties Tests
// ============================================================================

func TestArrayHelperLength(t *testing.T) {
	input := `
		var dynArr: array of Integer;
		var len: Integer;
		begin
			SetLength(dynArr, 5);
			len := dynArr.Length;
			PrintLn(len);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "5\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestArrayHelperHigh(t *testing.T) {
	input := `
		var dynArr: array of Integer;
		var high: Integer;
		begin
			SetLength(dynArr, 10);
			high := dynArr.High;
			PrintLn(high);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "9\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestArrayHelperLow(t *testing.T) {
	input := `
		var dynArr: array of Integer;
		var low: Integer;
		begin
			SetLength(dynArr, 10);
			low := dynArr.Low;
			PrintLn(low);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "0\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestArrayHelperEmptyArray(t *testing.T) {
	input := `
		var arr: array of Integer;
		begin
			PrintLn(arr.Length);
			PrintLn(arr.Low);
			PrintLn(arr.High);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "0\n0\n-1\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestArrayHelperStaticArray(t *testing.T) {
	input := `
		var staticArr: array[1..5] of String;
		begin
			PrintLn(staticArr.Length);
			PrintLn(staticArr.Low);
			PrintLn(staticArr.High);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "5\n1\n5\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestArrayHelperInForLoop(t *testing.T) {
	input := `
		var arr: array of Integer;
		var i: Integer;
		begin
			SetLength(arr, 3);
			arr[0] := 10;
			arr[1] := 20;
			arr[2] := 30;

			for i := arr.Low to arr.High do
				PrintLn(arr[i]);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "10\n20\n30\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// Test case-insensitive array helper property access
func TestArrayHelperCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "lowercase properties",
			input: `
				var arr: array of Integer;
				begin
					SetLength(arr, 5);
					PrintLn(arr.length);
					PrintLn(arr.low);
					PrintLn(arr.high);
				end.
			`,
			expected: "5\n0\n4\n",
		},
		{
			name: "uppercase properties",
			input: `
				var arr: array of Integer;
				begin
					SetLength(arr, 3);
					PrintLn(arr.LENGTH);
					PrintLn(arr.LOW);
					PrintLn(arr.HIGH);
				end.
			`,
			expected: "3\n0\n2\n",
		},
		{
			name: "mixed case properties",
			input: `
				var arr: array of Integer;
				begin
					SetLength(arr, 4);
					PrintLn(arr.LeNgTh);
					PrintLn(arr.LoW);
					PrintLn(arr.HiGh);
				end.
			`,
			expected: "4\n0\n3\n",
		},
		{
			name: "lowercase in for loop (like one_dim_automata.pas)",
			input: `
				var arr: array of Integer;
				var i: Integer;
				begin
					SetLength(arr, 5);
					arr[0] := 1;
					arr[1] := 2;
					arr[2] := 3;
					arr[3] := 4;
					arr[4] := 5;

					for i := arr.low+1 to arr.high-1 do
						PrintLn(arr[i]);
				end.
			`,
			expected: "2\n3\n4\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			result := interpret(interp, tt.input)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("wrong output. expected=%q, got=%q", tt.expected, out.String())
			}
		})
	}
}
