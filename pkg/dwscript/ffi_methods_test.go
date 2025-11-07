package dwscript

import (
	"bytes"
	"strings"
	"testing"
)

// Counter is a test struct with methods for testing RegisterMethod
type Counter struct {
	value int64
}

func (c *Counter) Increment() {
	c.value++
}

func (c *Counter) Add(x int64) {
	c.value += x
}

func (c *Counter) GetValue() int64 {
	return c.value
}

func (c *Counter) MultiplyBy(factor int64) int64 {
	c.value *= factor
	return c.value
}

// Calculator is a test struct demonstrating more complex method behavior
type Calculator struct {
	result float64
	ops    int64
}

func (calc *Calculator) Add(x float64) {
	calc.result += x
	calc.ops++
}

func (calc *Calculator) Multiply(x float64) {
	calc.result *= x
	calc.ops++
}

func (calc *Calculator) GetResult() float64 {
	return calc.result
}

func (calc *Calculator) GetOpsCount() int64 {
	return calc.ops
}

func (calc *Calculator) Reset() {
	calc.result = 0.0
	calc.ops = 0
}

// Accumulator is a test struct with var parameter method
type Accumulator struct {
	total int64
}

// AddAndGetTotal is a method for TestFFIMethodWithVarParam.
func (a *Accumulator) AddAndGetTotal(value int64, total *int64) {
	a.total += value
	*total = a.total
}

// TestRegisterMethodValue tests registering methods using method values (Option 1)
// This approach uses the existing RegisterFunction API with method values.
func TestRegisterMethodValue(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	counter := &Counter{value: 10}

	// Register using method values - the simplest approach
	err = engine.RegisterFunction("Increment", counter.Increment)
	if err != nil {
		t.Fatalf("failed to register Increment: %v", err)
	}

	err = engine.RegisterFunction("Add", counter.Add)
	if err != nil {
		t.Fatalf("failed to register Add: %v", err)
	}

	err = engine.RegisterFunction("GetValue", counter.GetValue)
	if err != nil {
		t.Fatalf("failed to register GetValue: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		Increment();
		Add(5);
		var result := GetValue();
		PrintLn(IntToStr(result));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: 10 + 1 + 5 = 16
	output := strings.TrimSpace(buf.String())
	if output != "16" {
		t.Errorf("expected output '16', got '%s'", output)
	}
}

// TestRegisterMethod tests the new RegisterMethod API (Option 2)
// This is a more explicit API that validates the method exists.
func TestRegisterMethod(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	counter := &Counter{value: 5}

	// Register using RegisterMethod API - more explicit
	err = engine.RegisterMethod("DoIncrement", counter, "Increment")
	if err != nil {
		t.Fatalf("failed to register Increment: %v", err)
	}

	err = engine.RegisterMethod("Add", counter, "Add")
	if err != nil {
		t.Fatalf("failed to register Add: %v", err)
	}

	err = engine.RegisterMethod("Get", counter, "GetValue")
	if err != nil {
		t.Fatalf("failed to register GetValue: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		DoIncrement();
		DoIncrement();
		Add(3);
		var result := Get();
		PrintLn(IntToStr(result));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: 5 + 1 + 1 + 3 = 10
	output := strings.TrimSpace(buf.String())
	if output != "10" {
		t.Errorf("expected output '10', got '%s'", output)
	}
}

// TestRegisterMethodPointerReceiver tests methods with pointer receivers
// Pointer receivers can modify the receiver's state.
func TestRegisterMethodPointerReceiver(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	calc := &Calculator{result: 100.0}

	err = engine.RegisterMethod("Add", calc, "Add")
	if err != nil {
		t.Fatalf("failed to register Add: %v", err)
	}

	err = engine.RegisterMethod("Multiply", calc, "Multiply")
	if err != nil {
		t.Fatalf("failed to register Multiply: %v", err)
	}

	err = engine.RegisterMethod("GetResult", calc, "GetResult")
	if err != nil {
		t.Fatalf("failed to register GetResult: %v", err)
	}

	err = engine.RegisterMethod("GetOpsCount", calc, "GetOpsCount")
	if err != nil {
		t.Fatalf("failed to register GetOpsCount: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		Add(50.0);
		Multiply(2.0);
		var result := GetResult();
		var ops := GetOpsCount();
		PrintLn(FloatToStr(result));
		PrintLn(IntToStr(ops));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: (100 + 50) * 2 = 300.0, ops = 2
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines of output, got %d", len(lines))
	}
	if lines[0] != "300" && lines[0] != "300.0" && lines[0] != "300.00" {
		t.Errorf("expected result '300', got '%s'", lines[0])
	}
	if lines[1] != "2" {
		t.Errorf("expected ops count '2', got '%s'", lines[1])
	}
}

// TestRegisterMethodWithReturnValue tests methods that return values
func TestRegisterMethodWithReturnValue(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	counter := &Counter{value: 7}

	err = engine.RegisterMethod("MultiplyBy", counter, "MultiplyBy")
	if err != nil {
		t.Fatalf("failed to register MultiplyBy: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var result := MultiplyBy(3);
		PrintLn(IntToStr(result));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: 7 * 3 = 21
	output := strings.TrimSpace(buf.String())
	if output != "21" {
		t.Errorf("expected output '21', got '%s'", output)
	}
}

// TestRegisterMethodMultipleInstances tests that each instance maintains separate state
func TestRegisterMethodMultipleInstances(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	counter1 := &Counter{value: 0}
	counter2 := &Counter{value: 100}

	// Register methods from different instances with different names
	err = engine.RegisterMethod("Increment1", counter1, "Increment")
	if err != nil {
		t.Fatalf("failed to register Increment1: %v", err)
	}

	err = engine.RegisterMethod("Get1", counter1, "GetValue")
	if err != nil {
		t.Fatalf("failed to register Get1: %v", err)
	}

	err = engine.RegisterMethod("Increment2", counter2, "Increment")
	if err != nil {
		t.Fatalf("failed to register Increment2: %v", err)
	}

	err = engine.RegisterMethod("Get2", counter2, "GetValue")
	if err != nil {
		t.Fatalf("failed to register Get2: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		Increment1();
		Increment1();
		Increment2();
		var v1 := Get1();
		var v2 := Get2();
		PrintLn(IntToStr(v1));
		PrintLn(IntToStr(v2));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: counter1 = 0 + 2 = 2, counter2 = 100 + 1 = 101
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines of output, got %d", len(lines))
	}
	if lines[0] != "2" {
		t.Errorf("expected counter1 value '2', got '%s'", lines[0])
	}
	if lines[1] != "101" {
		t.Errorf("expected counter2 value '101', got '%s'", lines[1])
	}
}

// TestRegisterMethodErrors tests error cases for RegisterMethod
func TestRegisterMethodErrors(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	counter := &Counter{value: 0}

	// Test: nil receiver
	err = engine.RegisterMethod("Test", nil, "Increment")
	if err == nil {
		t.Errorf("expected error for nil receiver")
	}

	// Test: empty method name
	err = engine.RegisterMethod("Test", counter, "")
	if err == nil {
		t.Errorf("expected error for empty method name")
	}

	// Test: non-existent method
	err = engine.RegisterMethod("Test", counter, "NonExistent")
	if err == nil {
		t.Errorf("expected error for non-existent method")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}

	// Test: private method (lowercase, unexported)
	// Note: Go reflection can't access unexported methods at all,
	// so MethodByName will return false for unexported methods
	err = engine.RegisterMethod("Test", counter, "privateMethod")
	if err == nil {
		t.Errorf("expected error for private method")
	}
}

// TestRegisterMethodWithComplexOperations tests a more complex scenario
func TestRegisterMethodWithComplexOperations(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	calc := &Calculator{result: 10.0}

	// Register all methods
	err = engine.RegisterMethod("Add", calc, "Add")
	if err != nil {
		t.Fatalf("failed to register Add: %v", err)
	}

	err = engine.RegisterMethod("Multiply", calc, "Multiply")
	if err != nil {
		t.Fatalf("failed to register Multiply: %v", err)
	}

	err = engine.RegisterMethod("GetResult", calc, "GetResult")
	if err != nil {
		t.Fatalf("failed to register GetResult: %v", err)
	}

	err = engine.RegisterMethod("Reset", calc, "Reset")
	if err != nil {
		t.Fatalf("failed to register Reset: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		// Start with 10.0
		Add(5.0);      // 15.0
		Multiply(2.0); // 30.0
		Add(10.0);     // 40.0
		var result1 := GetResult();
		PrintLn(FloatToStr(result1));

		// Reset and start over
		Reset();
		Add(100.0);
		var result2 := GetResult();
		PrintLn(FloatToStr(result2));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines of output, got %d: %v", len(lines), lines)
	}
	// First result: (10 + 5) * 2 + 10 = 40
	if lines[0] != "40" && lines[0] != "40.0" && lines[0] != "40.00" {
		t.Errorf("expected first result '40', got '%s'", lines[0])
	}
	// After reset: 0 + 100 = 100
	if lines[1] != "100" && lines[1] != "100.0" && lines[1] != "100.00" {
		t.Errorf("expected second result '100', got '%s'", lines[1])
	}
}

// TestFFIMethodRegistrationErrors tests error handling when registering methods.
func TestFFIMethodRegistrationErrors(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	type Counter struct {
		value int64
	}

	counter := &Counter{value: 0}

	// Try to register non-existent method
	err = engine.RegisterMethod("NonExistent", counter, "NonExistent")
	if err == nil {
		t.Fatal("expected error when registering non-existent method")
	}

	// Try to register with nil receiver
	err = engine.RegisterMethod("Increment", nil, "Increment")
	if err == nil {
		t.Fatal("expected error when registering method with nil receiver")
	}

	// Try to register with empty name
	err = engine.RegisterMethod("", counter, "Increment")
	if err == nil {
		t.Fatal("expected error when registering method with empty name")
	}
}

// TestFFIMethodWithVarParam tests methods with var parameters.
func TestFFIMethodWithVarParam(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	acc := &Accumulator{total: 0}

	// Register method that uses var parameter
	err = engine.RegisterMethod("AddAndGetTotal", acc, "AddAndGetTotal")
	if err != nil {
		t.Fatalf("failed to register method: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var newTotal: Integer;
		AddAndGetTotal(10, newTotal);
		PrintLn('Total after 10: ' + IntToStr(newTotal));

		AddAndGetTotal(20, newTotal);
		PrintLn('Total after 20: ' + IntToStr(newTotal));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("execution was not successful")
	}

	output := strings.TrimSpace(buf.String())
	expected := "Total after 10: 10\nTotal after 20: 30"
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}
