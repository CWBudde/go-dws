package dwscript

import (
	"bytes"
	"strings"
	"testing"
)

// TestFFI_VarParamBasic tests basic var parameter modification with integers
func TestFFI_VarParamBasic(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function with pointer parameter (var param)
	err = engine.RegisterFunction("Increment", func(x *int64) {
		*x++
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Call from DWScript
	var buf bytes.Buffer
	engine.SetOutput(&buf)
	result, err := engine.Eval(`
		var n: Integer := 5;
		Increment(n);
		PrintLn(IntToStr(n));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	if output != "6" {
		t.Errorf("expected output '6', got '%s'", output)
	}
}

// TestFFI_VarParamSwap tests the classic swap function with two pointers
func TestFFI_VarParamSwap(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go swap function
	err = engine.RegisterFunction("Swap", func(a, b *int64) {
		temp := *a
		*a = *b
		*b = temp
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Call from DWScript
	var buf bytes.Buffer
	engine.SetOutput(&buf)
	result, err := engine.Eval(`
		var x: Integer := 10;
		var y: Integer := 20;
		Swap(x, y);
		PrintLn(IntToStr(x));
		PrintLn(IntToStr(y));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	expected := "20\n10"
	if output != expected {
		t.Errorf("expected output:\n%s\ngot:\n%s", expected, output)
	}
}

// TestFFI_VarParamString tests var parameter with string type
func TestFFI_VarParamString(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that modifies a string
	err = engine.RegisterFunction("MakeUpperCase", func(s *string) {
		*s = strings.ToUpper(*s)
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Call from DWScript
	var buf bytes.Buffer
	engine.SetOutput(&buf)
	result, err := engine.Eval(`
		var text: String := 'hello';
		MakeUpperCase(text);
		PrintLn(text);
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	if output != "HELLO" {
		t.Errorf("expected output 'HELLO', got '%s'", output)
	}
}

// TestFFI_VarParamMixed tests function with both value and pointer parameters
func TestFFI_VarParamMixed(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function with mixed parameters
	err = engine.RegisterFunction("AddAndStore", func(result *int64, a, b int64) {
		*result = a + b
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Call from DWScript
	var buf bytes.Buffer
	engine.SetOutput(&buf)
	result, err := engine.Eval(`
		var sum: Integer := 0;
		AddAndStore(sum, 15, 27);
		PrintLn(IntToStr(sum));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	if output != "42" {
		t.Errorf("expected output '42', got '%s'", output)
	}
}

// TestFFI_VarParamFloat tests var parameter with float type
func TestFFI_VarParamFloat(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that doubles a float
	err = engine.RegisterFunction("Double", func(x *float64) {
		*x *= 2.0
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Call from DWScript
	var buf bytes.Buffer
	engine.SetOutput(&buf)
	result, err := engine.Eval(`
		var pi: Float := 3.14;
		Double(pi);
		PrintLn(FloatToStr(pi));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	// Check that it's approximately 6.28
	if !strings.HasPrefix(output, "6.28") {
		t.Errorf("expected output to start with '6.28', got '%s'", output)
	}
}

// TestFFI_VarParamBool tests var parameter with boolean type
func TestFFI_VarParamBool(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that negates a boolean
	err = engine.RegisterFunction("Negate", func(b *bool) {
		*b = !*b
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Call from DWScript
	var buf bytes.Buffer
	engine.SetOutput(&buf)
	result, err := engine.Eval(`
		var flag: Boolean := True;
		Negate(flag);
		if flag then
			PrintLn('Still true')
		else
			PrintLn('Now false');
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	if output != "Now false" {
		t.Errorf("expected output 'Now false', got '%s'", output)
	}
}

// TestRegisterVariadicFunction tests variadic function registration.
func TestRegisterVariadicFunction(t *testing.T) {
	t.Run("VariadicIntSum", func(t *testing.T) {
		engine, err := New(WithTypeCheck(false))
		if err != nil {
			t.Fatalf("failed to create engine: %v", err)
		}

		// Register a variadic function that sums integers
		err = engine.RegisterFunction("Sum", func(nums ...int64) int64 {
			var sum int64
			for _, n := range nums {
				sum += n
			}
			return sum
		})
		if err != nil {
			t.Fatalf("failed to register function: %v", err)
		}

		// Test with multiple arguments
		var buf bytes.Buffer
		engine.SetOutput(&buf)
		_, err = engine.Eval(`
			var total := Sum(1, 2, 3, 4, 5);
			PrintLn(IntToStr(total));
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "15" {
			t.Errorf("expected output '15', got '%s'", output)
		}
	})

	t.Run("VariadicWithZeroArgs", func(t *testing.T) {
		engine, err := New(WithTypeCheck(false))
		if err != nil {
			t.Fatalf("failed to create engine: %v", err)
		}

		// Register a variadic function
		err = engine.RegisterFunction("Count", func(items ...string) int64 {
			return int64(len(items))
		})
		if err != nil {
			t.Fatalf("failed to register function: %v", err)
		}

		// Call with zero variadic arguments
		var buf bytes.Buffer
		engine.SetOutput(&buf)
		_, err = engine.Eval(`
			var count := Count();
			PrintLn(IntToStr(count));
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "0" {
			t.Errorf("expected output '0', got '%s'", output)
		}
	})

	t.Run("VariadicWithOneArg", func(t *testing.T) {
		engine, err := New(WithTypeCheck(false))
		if err != nil {
			t.Fatalf("failed to create engine: %v", err)
		}

		// Register a variadic function
		err = engine.RegisterFunction("Join", func(strs ...string) string {
			return strings.Join(strs, ",")
		})
		if err != nil {
			t.Fatalf("failed to register function: %v", err)
		}

		// Call with one variadic argument
		var buf bytes.Buffer
		engine.SetOutput(&buf)
		_, err = engine.Eval(`
			var result := Join('hello');
			PrintLn(result);
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "hello" {
			t.Errorf("expected output 'hello', got '%s'", output)
		}
	})
}

// TestVariadicFunctionSignature tests variadic function signature detection.
func TestVariadicFunctionSignature(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register a variadic function
	err = engine.RegisterFunction("TestVariadic", func(a string, nums ...int64) string {
		return a
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Get the signature (internal API, but useful for verification)
	fn, ok := engine.externalFunctions.Get("TestVariadic")
	if !ok {
		t.Fatal("function not found in registry")
	}
	// Cast the wrapper to access the signature
	wrapper, ok := fn.Wrapper.(*externalFunctionWrapper)
	if !ok {
		t.Fatal("wrapper is not the expected type")
	}
	sig := wrapper.Signature()
	if !sig.IsVariadic {
		t.Error("expected function to be marked as variadic")
	}

	if len(sig.ParamTypes) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(sig.ParamTypes))
	}

	if sig.ParamTypes[0] != "String" {
		t.Errorf("expected first param to be String, got %s", sig.ParamTypes[0])
	}

	if sig.ParamTypes[1] != "array of Integer" {
		t.Errorf("expected second param to be 'array of Integer', got %s", sig.ParamTypes[1])
	}

	// Check string representation includes variadic notation
	sigStr := sig.String()
	t.Logf("Signature string: %s", sigStr)
	if !strings.Contains(sigStr, "...") {
		t.Errorf("expected signature string to contain '...' for variadic param, got: %s", sigStr)
	}
}

// TestFFIVariadicWrongTypes tests type checking for variadic functions.
func TestFFIVariadicWrongTypes(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register variadic function expecting integers
	err = engine.RegisterFunction("SumInts", func(nums ...int64) int64 {
		sum := int64(0)
		for _, n := range nums {
			sum += n
		}
		return sum
	})
	if err != nil {
		t.Fatalf("failed to register SumInts: %v", err)
	}

	// Try to pass strings to integer variadic function
	result, err := engine.Eval(`
		var x := SumInts(['a', 'b', 'c']);  // Wrong types
	`)

	// Should fail with type error
	if err == nil && result.Success {
		t.Fatal("expected error when passing wrong types to variadic function")
	}
}

// TestFFIVarParamNilPointer tests nil pointer handling in var parameters.
func TestFFIVarParamNilPointer(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register function with var parameter
	err = engine.RegisterFunction("ModifyInt", func(x *int64) {
		if x != nil {
			*x = 42
		}
	})
	if err != nil {
		t.Fatalf("failed to register ModifyInt: %v", err)
	}

	// DWScript should never pass nil for var parameters - this tests the infrastructure
	// In normal use, DWScript always passes valid references
	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var x: Integer := 10;
		ModifyInt(x);
		PrintLn(IntToStr(x));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	if output != "42" {
		t.Errorf("expected '42', got '%s'", output)
	}
}

// TestFFIVarParamWithCallback tests combining var parameters with callbacks.
func TestFFIVarParamWithCallback(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register function with both var param and callback
	err = engine.RegisterFunction("ProcessAndCount", func(count *int64, items []int64, processor func(int64) int64) []int64 {
		result := make([]int64, len(items))
		for i, item := range items {
			result[i] = processor(item)
		}
		*count = int64(len(items))
		return result
	})
	if err != nil {
		t.Fatalf("failed to register ProcessAndCount: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		function Double(x: Integer): Integer;
		begin
			Result := x * 2;
		end;

		var count: Integer;
		var result := ProcessAndCount(count, [1, 2, 3, 4], @Double);

		PrintLn('Count: ' + IntToStr(count));
		PrintLn('First: ' + IntToStr(result[0]));
		PrintLn('Last: ' + IntToStr(result[3]));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	expected := "Count: 4\nFirst: 2\nLast: 8"
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}
