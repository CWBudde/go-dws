package dwscript

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// TestRegisterSimpleFunction tests registering and calling a simple Go function.
func TestRegisterSimpleFunction(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register a simple addition function
	err = engine.RegisterFunction("AddNumbers", func(a, b int64) int64 {
		return a + b
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Call it from DWScript
	var buf bytes.Buffer
	engine.SetOutput(&buf)
	result, err := engine.Eval(`
		var sum := AddNumbers(40, 2);
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

// TestRegisterFunctionWithError tests error handling in external functions.
func TestRegisterFunctionWithError(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register a function that returns an error
	err = engine.RegisterFunction("Divide", func(a, b int64) (int64, error) {
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Test successful call
	var buf bytes.Buffer
	engine.SetOutput(&buf)
	_, err = engine.Eval(`
		var quotient := Divide(10, 2);
		PrintLn(IntToStr(quotient));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "5" {
		t.Errorf("expected output '5', got '%s'", output)
	}

	// Test error case - should raise an EHost exception
	buf.Reset()
	_, err = engine.Eval(`
		try
			var quotient := Divide(10, 0);
			PrintLn('Should not reach here');
		except
			on E: EHost do
				PrintLn('Caught error: ' + E.Message);
		end;
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	output = strings.TrimSpace(buf.String())
	if !strings.Contains(output, "division by zero") {
		t.Errorf("expected error message about division by zero, got '%s'", output)
	}
}

// TestRegisterFunctionWithStrings tests string parameters and return values.
func TestRegisterFunctionWithStrings(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register a greeting function
	err = engine.RegisterFunction("Greet", func(name string) string {
		return "Hello, " + name + "!"
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)
	_, err = engine.Eval(`
		var greeting := Greet('World');
		PrintLn(greeting);
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got '%s'", output)
	}
}

// TestRegisterFunctionWithBool tests boolean parameters and return values.
func TestRegisterFunctionWithBool(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register a logical AND function
	err = engine.RegisterFunction("LogicalAnd", func(a, b bool) bool {
		return a && b
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)
	_, err = engine.Eval(`
		var result1 := LogicalAnd(true, true);
		var result2 := LogicalAnd(true, false);
		PrintLn(BoolToStr(result1));
		PrintLn(BoolToStr(result2));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines of output, got %d", len(lines))
	}
	// DWScript's BoolToStr returns "True"/"False" with capital first letter
	if lines[0] != "True" {
		t.Errorf("expected 'True', got '%s'", lines[0])
	}
	if lines[1] != "False" {
		t.Errorf("expected 'False', got '%s'", lines[1])
	}
}

// TestRegisterFunctionWithFloat tests float parameters and return values.
func TestRegisterFunctionWithFloat(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register a function that calculates circle area
	err = engine.RegisterFunction("CircleArea", func(radius float64) float64 {
		return 3.14159 * radius * radius
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)
	_, err = engine.Eval(`
		var area := CircleArea(2.0);
		PrintLn(FloatToStr(area));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	// Should be approximately 12.56636
	if !strings.HasPrefix(output, "12.5") {
		t.Errorf("expected output starting with '12.5', got '%s'", output)
	}
}

// TestRegisterFunctionWithMixedTypes tests multiple parameter types.
func TestRegisterFunctionWithMixedTypes(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register a function with mixed parameter types
	err = engine.RegisterFunction("FormatResult", func(count int64, item string, available bool) string {
		if available {
			return "Found " + string(rune('0'+count)) + " " + item
		}
		return "No " + item + " available"
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)
	_, err = engine.Eval(`
		var msg1 := FormatResult(5, 'apples', true);
		var msg2 := FormatResult(0, 'oranges', false);
		PrintLn(msg1);
		PrintLn(msg2);
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines of output, got %d", len(lines))
	}
	if lines[0] != "Found 5 apples" {
		t.Errorf("expected 'Found 5 apples', got '%s'", lines[0])
	}
	if lines[1] != "No oranges available" {
		t.Errorf("expected 'No oranges available', got '%s'", lines[1])
	}
}

// TestRegisterFunctionWithNoReturn tests a procedure (no return value).
func TestRegisterFunctionWithNoReturn(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register a function that returns only error (procedure with error handling)
	callCount := 0
	err = engine.RegisterFunction("DoSomething", func() error {
		callCount++
		return nil
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	result, err := engine.Eval(`
		DoSomething();
		DoSomething();
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	if callCount != 2 {
		t.Errorf("expected function to be called 2 times, got %d", callCount)
	}
}

// TestAPIDesign tests the overall API design for FFI.
func TestAPIDesign(t *testing.T) {
	t.Run("SimpleRegistration", func(t *testing.T) {
		engine, err := New(WithTypeCheck(false))
		if err != nil {
			t.Fatalf("failed to create engine: %v", err)
		}

		// Registration should be simple and type-safe
		err = engine.RegisterFunction("Double", func(x int64) int64 { return x * 2 })
		if err != nil {
			t.Fatalf("registration failed: %v", err)
		}

		// Should work immediately
		var buf bytes.Buffer
		engine.SetOutput(&buf)
		_, err = engine.Eval(`var result := Double(21); PrintLn(IntToStr(result));`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		if strings.TrimSpace(buf.String()) != "42" {
			t.Errorf("expected '42', got '%s'", buf.String())
		}
	})

	t.Run("ChainedRegistration", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Multiple registrations
		engine.RegisterFunction("F1", func() int64 { return 1 })
		engine.RegisterFunction("F2", func() int64 { return 2 })
		engine.RegisterFunction("F3", func() int64 { return 3 })

		var buf bytes.Buffer
		engine.SetOutput(&buf)
		_, err := engine.Eval(`
			var sum := F1() + F2() + F3();
			PrintLn(IntToStr(sum));
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		if strings.TrimSpace(buf.String()) != "6" {
			t.Errorf("expected '6', got '%s'", buf.String())
		}
	})

	t.Run("ErrorReporting", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Registration errors should be clear
		err := engine.RegisterFunction("BadFunc", "not a function")
		if err == nil {
			t.Error("expected error when registering non-function")
		}
		if !strings.Contains(err.Error(), "function") {
			t.Errorf("expected error message about function, got: %v", err)
		}

		// Duplicate registration
		engine.RegisterFunction("MyFunc", func() {})
		err = engine.RegisterFunction("MyFunc", func() {})
		if err == nil {
			t.Error("expected error when registering duplicate function")
		}
	})

	t.Run("ComplexWorkflow", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Realistic workflow: data processing
		engine.RegisterFunction("LoadData", func() []int64 {
			return []int64{5, 2, 8, 1, 9, 3}
		})

		engine.RegisterFunction("FilterEven", func(numbers []int64) []int64 {
			result := []int64{}
			for _, n := range numbers {
				if n%2 == 0 {
					result = append(result, n)
				}
			}
			return result
		})

		engine.RegisterFunction("Sum", func(numbers []int64) int64 {
			sum := int64(0)
			for _, n := range numbers {
				sum += n
			}
			return sum
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)
		_, err := engine.Eval(`
			var data := LoadData();
			var evens := FilterEven(data);
			var total := Sum(evens);
			PrintLn(IntToStr(total));
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		// Even numbers: 2, 8 -> sum = 10
		if strings.TrimSpace(buf.String()) != "10" {
			t.Errorf("expected '10', got '%s'", buf.String())
		}
	})
}
