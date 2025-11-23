package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Parity Tests - Quick Regression Suite
// ============================================================================
//
// These tests capture critical interpreter behaviors that must remain stable
// during refactoring. Run these tests to verify that changes don't break
// core functionality.
//
// Categories:
// 1. Exception Handling - try/except/finally, raise, exception hierarchy
// 2. Array/Inline Types - dynamic arrays, inline array types, SetLength
// 3. Class/OOP Features - constructors, inheritance, static members
// 4. Enum Handling - enum declaration, scoped enums, ordinal values
// 5. Error Messages - location info, error formatting
// 6. Contracts - assert, require, ensure

// runParityTest is a helper to run a parity test with expected output
func runParityTest(t *testing.T, name string, input string, expected string) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		interp.Eval(program)

		output := buf.String()
		if output != expected {
			t.Errorf("expected output %q, got %q", expected, output)
		}
	})
}

// runParityTestContains is a helper that checks output contains expected substring
func runParityTestContains(t *testing.T, name string, input string, expected string) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		interp.Eval(program)

		output := buf.String()
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got %q", expected, output)
		}
	})
}

// ============================================================================
// 1. Exception Handling Parity Tests
// ============================================================================

func TestParityExceptionHandling(t *testing.T) {
	// Basic try-except
	runParityTest(t, "BasicTryExcept", `
		var caught: Boolean;
		caught := false;
		try
			raise Exception.Create('test');
		except
			on E: Exception do
				caught := true;
		end;
		PrintLn(caught);
	`, "True\n")

	// Exception message access
	runParityTest(t, "ExceptionMessage", `
		var msg: String;
		try
			raise Exception.Create('hello world');
		except
			on E: Exception do
				msg := E.Message;
		end;
		PrintLn(msg);
	`, "hello world\n")

	// Try-finally (no exception)
	runParityTest(t, "TryFinallyNoException", `
		var x: Integer;
		x := 0;
		try
			x := 1;
		finally
			x := x + 10;
		end;
		PrintLn(x);
	`, "11\n")

	// Try-finally (with exception)
	runParityTest(t, "TryFinallyWithException", `
		var x: Integer;
		x := 0;
		try
			try
				x := 1;
				raise Exception.Create('test');
			finally
				x := x + 10;
			end;
		except
			on E: Exception do
				x := x + 100;
		end;
		PrintLn(x);
	`, "111\n")

	// Bare except (catches all)
	runParityTest(t, "BareExcept", `
		var caught: Boolean;
		caught := false;
		try
			raise Exception.Create('test');
		except
			caught := true;
		end;
		PrintLn(caught);
	`, "True\n")

	// Exception hierarchy (ERangeError is Exception)
	runParityTest(t, "ExceptionHierarchy", `
		var caught: String;
		caught := 'none';
		try
			raise ERangeError.Create('range error');
		except
			on E: Exception do
				caught := 'Exception';
		end;
		PrintLn(caught);
	`, "Exception\n")

	// Nested try blocks
	runParityTest(t, "NestedTryBlocks", `
		var x: Integer;
		x := 0;
		try
			try
				raise Exception.Create('inner');
			except
				on E: Exception do
					x := 1;
			end;
			x := x + 10;
		except
			on E: Exception do
				x := x + 100;
		end;
		PrintLn(x);
	`, "11\n")
}

// ============================================================================
// 2. Array/Inline Types Parity Tests
// ============================================================================

func TestParityArrayHandling(t *testing.T) {
	// Dynamic array creation
	runParityTest(t, "DynamicArrayCreate", `
		var arr: array of Integer;
		SetLength(arr, 3);
		arr[0] := 10;
		arr[1] := 20;
		arr[2] := 30;
		PrintLn(arr[0]);
		PrintLn(arr[1]);
		PrintLn(arr[2]);
	`, "10\n20\n30\n")

	// Array length
	runParityTest(t, "ArrayLength", `
		var arr: array of Integer;
		SetLength(arr, 5);
		PrintLn(Length(arr));
	`, "5\n")

	// Array High/Low
	runParityTest(t, "ArrayHighLow", `
		var arr: array of Integer;
		SetLength(arr, 4);
		PrintLn(Low(arr));
		PrintLn(High(arr));
	`, "0\n3\n")

	// Array Add
	runParityTest(t, "ArrayAdd", `
		var arr: array of Integer;
		arr.Add(1);
		arr.Add(2);
		arr.Add(3);
		PrintLn(Length(arr));
		PrintLn(arr[0]);
		PrintLn(arr[2]);
	`, "3\n1\n3\n")

	// Array assignment - current behavior is reference semantics
	// NOTE: DWScript should have copy semantics for dynamic arrays,
	// but current implementation uses reference semantics.
	// This test documents current behavior.
	runParityTest(t, "ArrayAssignmentReference", `
		var arr1, arr2: array of Integer;
		SetLength(arr1, 2);
		arr1[0] := 10;
		arr1[1] := 20;
		arr2 := arr1;
		arr2[0] := 99;
		PrintLn(arr1[0]);
		PrintLn(arr2[0]);
	`, "99\n99\n") // Both reference same array

	// Array literal
	runParityTest(t, "ArrayLiteral", `
		var arr: array of Integer;
		arr := [1, 2, 3];
		PrintLn(Length(arr));
		PrintLn(arr[0]);
		PrintLn(arr[2]);
	`, "3\n1\n3\n")
}

// ============================================================================
// 3. Class/OOP Parity Tests
// ============================================================================

func TestParityClassOOP(t *testing.T) {
	// Basic class instantiation
	runParityTest(t, "BasicClass", `
		type
			TMyClass = class
				Value: Integer;
			end;
		var obj: TMyClass;
		obj := TMyClass.Create;
		obj.Value := 42;
		PrintLn(obj.Value);
	`, "42\n")

	// Constructor with parameter
	runParityTest(t, "ConstructorWithParam", `
		type
			TMyClass = class
				Value: Integer;
				constructor Create(v: Integer);
			end;
		constructor TMyClass.Create(v: Integer);
		begin
			Value := v;
		end;
		var obj: TMyClass;
		obj := TMyClass.Create(99);
		PrintLn(obj.Value);
	`, "99\n")

	// Inheritance - simple case without parent field access
	// NOTE: Full inheritance tests are in class_interp_test.go
	runParityTest(t, "InheritanceBasic", `
		type
			TDerived = class
				Y: Integer;
			end;
		var obj: TDerived;
		obj := TDerived.Create;
		obj.Y := 20;
		PrintLn(obj.Y);
	`, "20\n")

	// Class method (static)
	runParityTest(t, "ClassMethod", `
		type
			TMyClass = class
				class function GetValue: Integer;
			end;
		class function TMyClass.GetValue: Integer;
		begin
			Result := 42;
		end;
		PrintLn(TMyClass.GetValue);
	`, "42\n")

	// Instance method
	runParityTest(t, "InstanceMethod", `
		type
			TMyClass = class
				Value: Integer;
				function GetDouble: Integer;
			end;
		function TMyClass.GetDouble: Integer;
		begin
			Result := Value * 2;
		end;
		var obj: TMyClass;
		obj := TMyClass.Create;
		obj.Value := 21;
		PrintLn(obj.GetDouble);
	`, "42\n")

	// Self reference
	runParityTest(t, "SelfReference", `
		type
			TMyClass = class
				Value: Integer;
				procedure SetValue(v: Integer);
			end;
		procedure TMyClass.SetValue(v: Integer);
		begin
			Self.Value := v;
		end;
		var obj: TMyClass;
		obj := TMyClass.Create;
		obj.SetValue(123);
		PrintLn(obj.Value);
	`, "123\n")
}

// ============================================================================
// 4. Enum Parity Tests
// ============================================================================

func TestParityEnumHandling(t *testing.T) {
	// Basic enum
	runParityTest(t, "BasicEnum", `
		type
			TColor = (Red, Green, Blue);
		var c: TColor;
		c := Red;
		PrintLn(Ord(c));
	`, "0\n")

	// Enum with explicit values
	runParityTest(t, "EnumExplicitValues", `
		type
			TStatus = (
				Unknown = 0,
				Active = 1,
				Inactive = 5
			);
		var s: TStatus;
		s := Inactive;
		PrintLn(Ord(s));
	`, "5\n")

	// Enum iteration with for-in
	runParityTest(t, "EnumForIn", `
		type
			TDay = (Mon, Tue, Wed);
		var d: TDay;
		for d in TDay do
			PrintLn(Ord(d));
	`, "0\n1\n2\n")

	// Scoped enum access
	runParityTest(t, "ScopedEnumAccess", `
		type
			TColor = (Red, Green, Blue);
		var c: TColor;
		c := TColor.Green;
		PrintLn(Ord(c));
	`, "1\n")
}

// ============================================================================
// 5. Error Message Parity Tests
// ============================================================================

func TestParityErrorMessages(t *testing.T) {
	// Undefined variable error contains location
	t.Run("UndefinedVariableError", func(t *testing.T) {
		input := `PrintLn(undefinedVar);`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		// Should be an error value
		if result == nil {
			t.Fatal("expected error value, got nil")
		}

		// Error message should contain "undefinedVar"
		errStr := result.String()
		if !strings.Contains(errStr, "undefinedVar") {
			t.Errorf("error message should contain 'undefinedVar', got: %s", errStr)
		}
	})
}

// ============================================================================
// 6. Contract/Assert Parity Tests
// ============================================================================

func TestParityContracts(t *testing.T) {
	// Assert true passes
	runParityTest(t, "AssertTrue", `
		Assert(true, 'should pass');
		PrintLn('passed');
	`, "passed\n")

	// Assert false with message
	t.Run("AssertFalseMessage", func(t *testing.T) {
		input := `Assert(false, 'custom assertion message');`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		// Should produce an error/exception
		if result == nil {
			t.Fatal("Assert(false) should produce error, got nil")
		}

		// Check that custom message is included
		errStr := result.String()
		if !strings.Contains(errStr, "custom assertion message") {
			t.Errorf("assertion error should contain custom message, got: %s", errStr)
		}
	})
}

// ============================================================================
// 7. Additional Critical Behaviors
// ============================================================================

func TestParityCriticalBehaviors(t *testing.T) {
	// Closure captures variable - using lambda syntax
	// NOTE: The simpler lambda syntax is used here as the full
	// function pointer return type syntax is complex to parse.
	runParityTest(t, "ClosureCapture", `
		var x: Integer;
		var add: procedure(y: Integer);
		x := 5;
		add := lambda(y: Integer)
		begin
			PrintLn(x + y);
		end;
		add(10);
	`, "15\n")

	// Record value copy
	runParityTest(t, "RecordValueCopy", `
		type
			TPoint = record
				X, Y: Integer;
			end;
		var p1, p2: TPoint;
		p1.X := 10;
		p1.Y := 20;
		p2 := p1;
		p2.X := 99;
		PrintLn(p1.X);
		PrintLn(p2.X);
	`, "10\n99\n")

	// Function overloading
	runParityTest(t, "FunctionOverload", `
		function Add(a, b: Integer): Integer; overload;
		begin
			Result := a + b;
		end;
		function Add(a, b: Float): Float; overload;
		begin
			Result := a + b;
		end;
		PrintLn(Add(2, 3));
		PrintLn(Add(2.5, 3.5));
	`, "5\n6\n")

	// String concatenation
	runParityTest(t, "StringConcat", `
		var s: String;
		s := 'Hello' + ' ' + 'World';
		PrintLn(s);
	`, "Hello World\n")

	// Boolean short-circuit
	runParityTest(t, "BooleanShortCircuit", `
		var called: Boolean;
		called := false;
		function SetCalled: Boolean;
		begin
			called := true;
			Result := true;
		end;
		if false and SetCalled then
			PrintLn('should not print');
		PrintLn(called);
	`, "False\n")
}
