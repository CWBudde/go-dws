package interp

import (
	"strings"
	"testing"
)

// TestFunctionCalls tests user-defined function calls.
func TestFunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple function returning integer",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				PrintLn(Add(2, 3))
			`,
			expected: "5\n",
		},
		{
			name: "Function with single parameter",
			input: `
				function Double(x: Integer): Integer;
				begin
					Result := x * 2;
				end;

				PrintLn(Double(21))
			`,
			expected: "42\n",
		},
		{
			name: "Function using function name for return value",
			input: `
				function GetTen: Integer;
				begin
					GetTen := 10;
				end;

				PrintLn(GetTen())
			`,
			expected: "10\n",
		},
		{
			name: "Function called multiple times",
			input: `
				function Square(x: Integer): Integer;
				begin
					Result := x * x;
				end;

				PrintLn(Square(3));
				PrintLn(Square(4));
				PrintLn(Square(5))
			`,
			expected: "9\n16\n25\n",
		},
		{
			name: "Function with string parameter and return",
			input: `
				function Greet(name: String): String;
				begin
					Result := "Hello, " + name;
				end;

				PrintLn(Greet("World"))
			`,
			expected: "Hello, World\n",
		},
		{
			name: "Function with local variables",
			input: `
				function Calculate(x: Integer): Integer;
				begin
					var temp: Integer := x * 2;
					var result: Integer := temp + 10;
					Result := result;
				end;

				PrintLn(Calculate(5))
			`,
			expected: "20\n",
		},
		{
			name: "Multiple functions",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				function Multiply(a, b: Integer): Integer;
				begin
					Result := a * b;
				end;

				PrintLn(Add(2, 3));
				PrintLn(Multiply(4, 5))
			`,
			expected: "5\n20\n",
		},
		{
			name: "Nested function calls",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				function Multiply(a, b: Integer): Integer;
				begin
					Result := a * b;
				end;

				PrintLn(Add(Multiply(2, 3), 4))
			`,
			expected: "10\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expected {
				t.Errorf("wrong output.\nexpected=%q\ngot=%q", tt.expected, output)
			}
		})
	}
}

// TestProcedures tests procedures (functions without return values).
func TestProcedures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple procedure",
			input: `
				procedure SayHello;
				begin
					PrintLn("Hello!");
				end;

				SayHello()
			`,
			expected: "Hello!\n",
		},
		{
			name: "Procedure with parameters",
			input: `
				procedure Greet(name: String);
				begin
					PrintLn("Hello, " + name);
				end;

				Greet("Alice");
				Greet("Bob")
			`,
			expected: "Hello, Alice\nHello, Bob\n",
		},
		{
			name: "Procedure modifying outer variable",
			input: `
				var counter: Integer := 0;

				procedure Increment;
				begin
					counter := counter + 1;
				end;

				Increment();
				Increment();
				Increment();
				PrintLn(counter)
			`,
			expected: "3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expected {
				t.Errorf("wrong output.\nexpected=%q\ngot=%q", tt.expected, output)
			}
		})
	}
}

// TestRecursiveFunctions tests recursive function calls.
func TestRecursiveFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Factorial",
			input: `
				function Factorial(n: Integer): Integer;
				begin
					if n <= 1 then
						Result := 1
					else
						Result := n * Factorial(n - 1);
				end;

				PrintLn(Factorial(5))
			`,
			expected: "120\n",
		},
		{
			name: "Fibonacci",
			input: `
				function Fibonacci(n: Integer): Integer;
				begin
					if n <= 1 then
						Result := n
					else
						Result := Fibonacci(n - 1) + Fibonacci(n - 2);
				end;

				PrintLn(Fibonacci(0));
				PrintLn(Fibonacci(1));
				PrintLn(Fibonacci(6))
			`,
			expected: "0\n1\n8\n",
		},
		{
			name: "Countdown",
			input: `
				procedure Countdown(n: Integer);
				begin
					if n > 0 then
					begin
						PrintLn(n);
						Countdown(n - 1);
					end;
				end;

				Countdown(5)
			`,
			expected: "5\n4\n3\n2\n1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expected {
				t.Errorf("wrong output.\nexpected=%q\ngot=%q", tt.expected, output)
			}
		})
	}
}

// TestFunctionScopeIsolation tests that function scopes are properly isolated.
func TestFunctionScopeIsolation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Local variable doesn't leak to global scope",
			input: `
				function Test: Integer;
				begin
					var local: Integer := 42;
					Result := local;
				end;

				var x: Integer := Test();
				PrintLn(x)
			`,
			expected: "42\n",
		},
		{
			name: "Same variable name in different scopes",
			input: `
				var x: Integer := 10;

				function GetX: Integer;
				begin
					var x: Integer := 20;
					Result := x;
				end;

				PrintLn(GetX());
				PrintLn(x)
			`,
			expected: "20\n10\n",
		},
		{
			name: "Function can access global variables",
			input: `
				var global: Integer := 100;

				function AddToGlobal(x: Integer): Integer;
				begin
					Result := global + x;
				end;

				PrintLn(AddToGlobal(23))
			`,
			expected: "123\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expected {
				t.Errorf("wrong output.\nexpected=%q\ngot=%q", tt.expected, output)
			}
		})
	}
}

// TestFunctionErrors tests error handling in function calls.
func TestFunctionErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "Wrong number of arguments",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				PrintLn(Add(1))
			`,
			expectedErr: "wrong number of arguments",
		},
		{
			name: "Too many arguments",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				PrintLn(Add(1, 2, 3))
			`,
			expectedErr: "wrong number of arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)
			if !isError(val) {
				t.Errorf("expected error, got %T (%+v)", val, val)
				return
			}

			// Get error message (works for both interp.ErrorValue and runtime.ErrorValue)
			errorMsg := val.String()
			if !strings.Contains(errorMsg, tt.expectedErr) {
				t.Errorf("wrong error message. expected to contain %q, got=%q",
					tt.expectedErr, errorMsg)
			}
		})
	}
}

// TestMemberAssignment tests member assignment (obj.field := value) functionality.
// This is crucial for class functionality to work properly.
func TestMemberAssignment(t *testing.T) {
	tests := []struct {
		expected interface{}
		name     string
		input    string
	}{
		{
			name: "Simple member assignment in constructor",
			input: `
				type TPoint = class
					X: Integer;
					Y: Integer;

					function Create(x: Integer; y: Integer): TPoint;
					begin
						Self.X := x;
						Self.Y := y;
					end;

					function GetX(): Integer;
					begin
						Result := Self.X;
					end;
				end;

				var p: TPoint;
				p := TPoint.Create(10, 20);
				p.GetX()
			`,
			expected: int64(10),
		},
		{
			name: "Member assignment in method",
			input: `
				type TCounter = class
					Count: Integer;

					function Create(): TCounter;
					begin
						Self.Count := 0;
					end;

					procedure Increment();
					begin
						Self.Count := Self.Count + 1;
					end;

					function GetCount(): Integer;
					begin
						Result := Self.Count;
					end;
				end;

				var c: TCounter;
				c := TCounter.Create();
				c.Increment();
				c.Increment();
				c.GetCount()
			`,
			expected: int64(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)

			if isError(val) {
				t.Fatalf("evaluation error: %s", val.String())
			}

			switch expected := tt.expected.(type) {
			case int64:
				testIntegerValue(t, val, expected)
			case string:
				testStringValue(t, val, expected)
			}
		})
	}
}

// TestExternalVarRuntime tests runtime behavior of external variables.
// External variables should raise errors when accessed until
// getter/setter functions are provided.
func TestExternalVarRuntime(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "reading external variable raises error",
			input: `
				var x: Integer external;
				x
			`,
			expectedError: "Unsupported external variable access: x",
		},
		{
			name: "writing external variable raises error",
			input: `
				var y: String external 'externalY';
				y := 'test'
			`,
			expectedError: "Unsupported external variable assignment: y",
		},
		{
			name: "reading external variable in expression raises error",
			input: `
				var z: Integer external;
				var result: Integer;
				result := z + 10
			`,
			expectedError: "Unsupported external variable access: z",
		},
		{
			name: "external variable can be declared but not used",
			input: `
				var ext: Float external;
				var regular: Float;
				regular := 3.14;
				regular
			`,
			expectedError: "", // No error - external var is declared but not accessed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)

			if tt.expectedError == "" {
				// Should not be an error
				if isError(val) {
					t.Fatalf("unexpected error: %s", val.String())
				}
			} else {
				// Should be an error
				if !isError(val) {
					t.Fatalf("expected error containing %q, got: %s", tt.expectedError, val.String())
				}

				// Check error message (works for both interp.ErrorValue and runtime.ErrorValue)
				errorString := val.String()
				if !strings.Contains(errorString, tt.expectedError) {
					t.Errorf("error = %q, want to contain %q", errorString, tt.expectedError)
				}
			}
		})
	}
}

// TestFunctionArgumentSingleEvaluation verifies that function arguments are evaluated only once
// This test addresses the bug where functions called as arguments were evaluated twice:
// once during overload resolution and once during argument preparation.
func TestFunctionArgumentSingleEvaluation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name: "regular parameter - single evaluation",
			input: `
var callCount: Integer := 0;

function Test(x: Integer): Integer;
begin
    callCount := callCount + 1;
    PrintLn('Called');
    Result := x;
end;

begin
    PrintLn(Test(5));
    PrintLn('Call count: ' + IntToStr(callCount));
end;
`,
			expectedOutput: "Called\n5\nCall count: 1\n",
		},
		{
			name: "nested function calls - single evaluation each",
			input: `
var callCount: Integer := 0;

function TestA(x: Integer): Integer;
begin
    callCount := callCount + 1;
    PrintLn('TestA called');
    Result := x + 10;
end;

function TestB(x: Integer): Integer;
begin
    callCount := callCount + 100;
    PrintLn('TestB called');
    Result := x + 20;
end;

begin
    PrintLn(TestA(TestB(5)));
    PrintLn('Call count: ' + IntToStr(callCount));
end;
`,
			expectedOutput: "TestB called\nTestA called\n35\nCall count: 101\n",
		},
		{
			name: "lazy parameter - deferred evaluation",
			input: `
var callCount: Integer := 0;

function Test(x: Integer): Integer;
begin
    callCount := callCount + 1;
    PrintLn('Test called');
    Result := x;
end;

function LazyFunc(lazy arg: Integer): Integer;
begin
    PrintLn('LazyFunc entered');
    Result := arg;  // Evaluates the lazy argument here
end;

begin
    PrintLn(LazyFunc(Test(5)));
    PrintLn('Call count: ' + IntToStr(callCount));
end;
`,
			expectedOutput: "LazyFunc entered\nTest called\n5\nCall count: 1\n",
		},
		{
			name: "var parameter - reference creation",
			input: `
var callCount: Integer := 0;
var value: Integer := 10;

function Test(): Integer;
begin
    callCount := callCount + 1;
    PrintLn('Test called');
    Result := 42;
end;

procedure Increment(var x: Integer);
begin
    x := x + 1;
end;

begin
    // This should not call Test() - var parameters require identifiers
    // So we use a direct variable reference
    Increment(value);
    PrintLn(value);
    PrintLn('Call count: ' + IntToStr(callCount));
end;
`,
			expectedOutput: "11\nCall count: 0\n",
		},
		{
			name: "overloaded function with side effects",
			input: `
var callCount: Integer := 0;

function Test(x: Integer): Integer;
begin
    callCount := callCount + 1;
    PrintLn('Test(Integer) called');
    Result := x;
end;

function Test(x: String): String; overload;
begin
    callCount := callCount + 10;
    PrintLn('Test(String) called');
    Result := x;
end;

begin
    PrintLn(Test(42));
    PrintLn(Test('hello'));
    PrintLn('Call count: ' + IntToStr(callCount));
end;
`,
			expectedOutput: "Test(Integer) called\n42\nTest(String) called\nhello\nCall count: 11\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expectedOutput {
				t.Errorf("output mismatch\nexpected:\n%q\ngot:\n%q", tt.expectedOutput, output)
			}
		})
	}
}
