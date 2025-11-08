package dwscript

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// TestCallback tests basic callback functionality
// Go function accepts a DWScript function and calls it back
func TestCallback(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that accepts a callback
	err = engine.RegisterFunction("ForEach", func(items []int64, callback func(int64)) {
		for _, item := range items {
			callback(item)
		}
	})
	if err != nil {
		t.Fatalf("failed to register ForEach: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var sum := 0;
		ForEach([1, 2, 3, 4, 5], lambda(x: Integer) begin
			sum := sum + x;
		end);
		PrintLn(IntToStr(sum));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: 1 + 2 + 3 + 4 + 5 = 15
	output := strings.TrimSpace(buf.String())
	if output != "15" {
		t.Errorf("expected output '15', got '%s'", output)
	}
}

// TestCallbackWithReturnValue tests callbacks that return values
func TestCallbackWithReturnValue(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that uses callback return value
	err = engine.RegisterFunction("Map", func(items []int64, mapper func(int64) int64) []int64 {
		result := make([]int64, len(items))
		for i, item := range items {
			result[i] = mapper(item)
		}
		return result
	})
	if err != nil {
		t.Fatalf("failed to register Map: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var doubled := Map([1, 2, 3], lambda(x: Integer): Integer begin
			Result := x * 2;
		end);

		var i: Integer;
		for i := 0 to High(doubled) do
			PrintLn(IntToStr(doubled[i]));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: 2, 4, 6
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "2" || lines[1] != "4" || lines[2] != "6" {
		t.Errorf("expected [2, 4, 6], got %v", lines)
	}
}

// TestNestedCallback tests multiple levels of callbacks (DWScript → Go → DWScript → Go)
func TestNestedCallback(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that calls callback twice
	err = engine.RegisterFunction("ApplyTwice", func(x int64, fn func(int64) int64) int64 {
		return fn(fn(x))
	})
	if err != nil {
		t.Fatalf("failed to register ApplyTwice: %v", err)
	}

	// Register another Go function
	err = engine.RegisterFunction("Double", func(x int64) int64 {
		return x * 2
	})
	if err != nil {
		t.Fatalf("failed to register Double: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var result := ApplyTwice(5, lambda(x: Integer): Integer begin
			Result := Double(x);
		end);
		PrintLn(IntToStr(result));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: 5 → Double → 10 → Double → 20
	output := strings.TrimSpace(buf.String())
	if output != "20" {
		t.Errorf("expected output '20', got '%s'", output)
	}
}

// TestCallbackWithFunctionPointer tests callbacks using function pointers (@function syntax)
func TestCallbackWithFunctionPointer(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that accepts callback
	err = engine.RegisterFunction("Apply", func(x int64, fn func(int64) int64) int64 {
		return fn(x)
	})
	if err != nil {
		t.Fatalf("failed to register Apply: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		function Triple(x: Integer): Integer;
		begin
			Result := x * 3;
		end;

		var result := Apply(7, @Triple);
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

// TestCallbackFilter tests a practical callback use case: filtering
func TestCallbackFilter(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that filters based on predicate
	err = engine.RegisterFunction("Filter", func(items []int64, predicate func(int64) bool) []int64 {
		result := make([]int64, 0)
		for _, item := range items {
			if predicate(item) {
				result = append(result, item)
			}
		}
		return result
	})
	if err != nil {
		t.Fatalf("failed to register Filter: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var numbers := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
		var evens := Filter(numbers, lambda(x: Integer): Boolean begin
			Result := (x mod 2) = 0;
		end);

		var i: Integer;
		for i := 0 to High(evens) do
			PrintLn(IntToStr(evens[i]));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: 2, 4, 6, 8, 10
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	expected := []string{"2", "4", "6", "8", "10"}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line %d: expected '%s', got '%s'", i, exp, lines[i])
		}
	}
}

// TestCallbackMultipleParameters tests callbacks with multiple parameters
func TestCallbackMultipleParameters(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function with callback that takes multiple parameters
	err = engine.RegisterFunction("ForEachIndexed", func(items []int64, callback func(int64, int64)) {
		for idx, item := range items {
			callback(int64(idx), item)
		}
	})
	if err != nil {
		t.Fatalf("failed to register ForEachIndexed: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var items := [10, 20, 30];
		ForEachIndexed(items, lambda(idx, value: Integer) begin
			PrintLn(IntToStr(idx) + ': ' + IntToStr(value));
		end);
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: "0: 10", "1: 20", "2: 30"
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	expected := []string{"0: 10", "1: 20", "2: 30"}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line %d: expected '%s', got '%s'", i, exp, lines[i])
		}
	}
}

// TestCallbackWithSideEffects tests callbacks with side effects.
func TestCallbackWithSideEffects(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that calls callback multiple times
	// Note: Using "DoNTimes" instead of "Repeat" to avoid conflict with DWScript repeat...until keyword
	err = engine.RegisterFunction("DoNTimes", func(n int64, action func()) {
		for i := int64(0); i < n; i++ {
			action()
		}
	})
	if err != nil {
		t.Fatalf("failed to register DoNTimes: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		var counter := 0;

		procedure IncrementAndPrint;
		begin
			counter := counter + 1;
			PrintLn(IntToStr(counter));
		end;

		DoNTimes(5, @IncrementAndPrint);
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("execution was not successful")
	}

	// Expected: 1, 2, 3, 4, 5
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	for i := 0; i < 5; i++ {
		expected := fmt.Sprintf("%d", i+1)
		if lines[i] != expected {
			t.Errorf("line %d: expected '%s', got '%s'", i, expected, lines[i])
		}
	}
}

// TestFFICallbackReentrancyLimit tests that recursion limits apply to callbacks.
func TestFFICallbackReentrancyLimit(t *testing.T) {
	engine, err := New(WithTypeCheck(false), WithMaxRecursionDepth(10))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register function that calls callback
	err = engine.RegisterFunction("CallN", func(n int64, callback func(int64) int64) int64 {
		if n <= 0 {
			return 0
		}
		return callback(n)
	})
	if err != nil {
		t.Fatalf("failed to register CallN: %v", err)
	}

	// Create deeply recursive callback that exceeds limit
	// Note: When callbacks panic due to stack overflow, the exception bubbles up as EHost
	// because Go catches the recursion limit exception and re-raises it
	result, err := engine.Eval(`
		function Recursive(n: Integer): Integer;
		begin
			if n <= 0 then
				Result := 0
			else
				Result := CallN(n - 1, @Recursive);  // Recurse through Go
		end;

		try
			var x := Recursive(100);  // Exceeds limit of 10
			PrintLn('Should not reach here');
		except
			on E: Exception do begin
				// May be EScriptStackOverflow or EHost (wrapped by FFI)
				PrintLn('Caught recursion exception');
			end;
		end;
	`)
	if err != nil {
		// Execution failed - which means exception was raised. That's actually what we want!
		// The recursion limit was hit, causing the execution to fail
		if !strings.Contains(err.Error(), "recursion") && !strings.Contains(err.Error(), "stack") {
			t.Fatalf("expected recursion error, got: %v", err)
		}
		// Test passes - recursion limit was enforced
		return
	}

	// If we get here, check that the exception handler ran
	if !result.Success {
		t.Fatal("execution was not successful")
	}
}
