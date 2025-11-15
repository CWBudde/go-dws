package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestCaseInsensitiveFunctionCall tests that functions can be called with any case combination.
// DWScript is case-insensitive, so PrintProc, printproc, and PRINTPROC should all work.
func TestCaseInsensitiveFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "lowercase call to TitleCase function",
			code: `
				procedure PrintProc();
				begin
					PrintLn('Called');
				end;

				printproc();  // lowercase call
			`,
			expected: "Called\n",
		},
		{
			name: "UPPERCASE call to TitleCase function",
			code: `
				procedure PrintProc();
				begin
					PrintLn('Called');
				end;

				PRINTPROC();  // uppercase call
			`,
			expected: "Called\n",
		},
		{
			name: "MixedCase call to TitleCase function",
			code: `
				procedure PrintProc();
				begin
					PrintLn('Called');
				end;

				PrInTpRoC();  // mixed case call
			`,
			expected: "Called\n",
		},
		{
			name: "TitleCase call to lowercase function",
			code: `
				procedure printproc();
				begin
					PrintLn('Called');
				end;

				PrintProc();  // TitleCase call
			`,
			expected: "Called\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			result := interpretCode(interp, tt.code)
			if isError(result) {
				t.Fatalf("execution failed: %s", result.String())
			}
			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestCaseInsensitiveRecursion tests that recursive function calls work with different case.
// This is the core bug - QuickSort calls quicksort (lowercase) which should work.
func TestCaseInsensitiveRecursion(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "factorial with lowercase recursive call",
			code: `
				function Factorial(n: Integer): Integer;
				begin
					if n <= 1 then
						Result := 1
					else
						Result := n * factorial(n - 1);  // lowercase recursive call!
				end;

				PrintLn(IntToStr(Factorial(5)));
			`,
			expected: "120\n",
		},
		{
			name: "factorial with UPPERCASE recursive call",
			code: `
				function Factorial(n: Integer): Integer;
				begin
					if n <= 1 then
						Result := 1
					else
						Result := n * FACTORIAL(n - 1);  // UPPERCASE recursive call!
				end;

				PrintLn(IntToStr(Factorial(5)));
			`,
			expected: "120\n",
		},
		{
			name: "countdown with mixed case recursive call",
			code: `
				procedure Countdown(n: Integer);
				begin
					PrintLn(IntToStr(n));
					if n > 0 then
						countdown(n - 1);  // lowercase recursive call!
				end;

				Countdown(3);
			`,
			expected: "3\n2\n1\n0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			result := interpretCode(interp, tt.code)
			if isError(result) {
				t.Fatalf("execution failed: %s", result.String())
			}
			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestCaseInsensitiveMutualRecursion tests mutual recursion with mixed case.
// IsEven calls isodd (lowercase), IsOdd calls iseven (lowercase).
func TestCaseInsensitiveMutualRecursion(t *testing.T) {
	code := `
		function IsEven(n: Integer): Boolean;
		begin
			if n = 0 then
				Result := true
			else
				Result := isodd(n - 1);  // lowercase call!
		end;

		function IsOdd(n: Integer): Boolean;
		begin
			if n = 0 then
				Result := false
			else
				Result := iseven(n - 1);  // lowercase call!
		end;

		PrintLn(BoolToStr(IsEven(4)));  // Should be True
		PrintLn(BoolToStr(IsOdd(4)));   // Should be False
		PrintLn(BoolToStr(IsEven(3)));  // Should be False
		PrintLn(BoolToStr(IsOdd(3)));   // Should be True
	`

	expected := "True\nFalse\nFalse\nTrue\n"

	var out bytes.Buffer
	interp := New(&out)
	result := interpretCode(interp, code)
	if isError(result) {
		t.Fatalf("execution failed: %s", result.String())
	}
	if out.String() != expected {
		t.Errorf("expected output %q, got %q", expected, out.String())
	}
}

// TestMixedCaseCallChain tests a chain of calls with different cases.
// a() -> B() -> C() where each call uses different case.
func TestMixedCaseCallChain(t *testing.T) {
	code := `
		procedure A();
		begin
			PrintLn('In A');
			b();  // lowercase call to B
		end;

		procedure B();
		begin
			PrintLn('In B');
			C();  // uppercase call to c
		end;

		procedure c();
		begin
			PrintLn('In C');
		end;

		a();  // lowercase call to A
	`

	expected := "In A\nIn B\nIn C\n"

	var out bytes.Buffer
	interp := New(&out)
	result := interpretCode(interp, code)
	if isError(result) {
		t.Fatalf("execution failed: %s", result.String())
	}
	if out.String() != expected {
		t.Errorf("expected output %q, got %q", expected, out.String())
	}
}

// TestCaseInsensitiveFunctionWithReturn tests functions with return values called with different case.
func TestCaseInsensitiveFunctionWithReturn(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "lowercase call to function with return value",
			code: `
				function GetValue(): Integer;
				begin
					Result := 42;
				end;

				PrintLn(IntToStr(getvalue()));  // lowercase call
			`,
			expected: "42\n",
		},
		{
			name: "chained function calls with mixed case",
			code: `
				function Double(x: Integer): Integer;
				begin
					Result := x * 2;
				end;

				function Triple(x: Integer): Integer;
				begin
					Result := x * 3;
				end;

				function AddThem(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				PrintLn(IntToStr(addthem(double(5), TRIPLE(3))));  // mixed case
			`,
			expected: "19\n",  // 5*2 + 3*3 = 10 + 9 = 19
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			result := interpretCode(interp, tt.code)
			if isError(result) {
				t.Fatalf("execution failed: %s", result.String())
			}
			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestCaseInsensitiveWithVarParam tests var parameters work with case-insensitive calls.
func TestCaseInsensitiveWithVarParam(t *testing.T) {
	code := `
		procedure ModifyValue(var x: Integer);
		begin
			x := x * 2;
		end;

		var value: Integer;
		value := 5;
		modifyvalue(value);  // lowercase call with var param
		PrintLn(IntToStr(value));
	`

	expected := "10\n"

	var out bytes.Buffer
	interp := New(&out)
	result := interpretCode(interp, code)
	if isError(result) {
		t.Fatalf("execution failed: %s", result.String())
	}
	if out.String() != expected {
		t.Errorf("expected output %q, got %q", expected, out.String())
	}
}

// TestBuiltinFunctionsStillWork verifies that built-in functions still work
// (they already handle case insensitivity correctly).
func TestBuiltinFunctionsStillWork(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "PrintLn with different cases",
			code: `
				PrintLn('A');
				println('B');
				PRINTLN('C');
			`,
			expected: "A\nB\nC\n",
		},
		{
			name: "IntToStr with different cases",
			code: `
				PrintLn(IntToStr(42));
				PrintLn(inttostr(43));
				PrintLn(INTTOSTR(44));
			`,
			expected: "42\n43\n44\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			result := interpretCode(interp, tt.code)
			if isError(result) {
				t.Fatalf("execution failed: %s", result.String())
			}
			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// interpretCode is a helper function to parse and evaluate DWScript code
func interpretCode(interp *Interpreter, input string) Value {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return newError("parser errors: %v", p.Errors())
	}

	return interp.Eval(program)
}
