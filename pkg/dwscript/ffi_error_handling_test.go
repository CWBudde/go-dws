package dwscript

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// TestPanicConversionToException tests that all types of Go panics are converted to EHost exceptions.
func TestPanicConversionToException(t *testing.T) {
	t.Run("StringPanic", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("PanicWithString", func() string {
			panic("this is a string panic")
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			try
				var result := PanicWithString();
				PrintLn('Should not reach here');
			except
				on E: EHost do
					PrintLn('Caught: ' + E.Message);
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "panic: this is a string panic") {
			t.Errorf("expected panic message in output, got '%s'", output)
		}
	})

	t.Run("ErrorPanic", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("PanicWithError", func() string {
			panic(errors.New("this is an error panic"))
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			try
				var result := PanicWithError();
				PrintLn('Should not reach here');
			except
				on E: EHost do
					PrintLn('Caught: ' + E.Message);
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "panic: this is an error panic") {
			t.Errorf("expected panic message in output, got '%s'", output)
		}
	})

	t.Run("IntegerPanic", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("PanicWithInt", func() string {
			panic(42)
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			try
				var result := PanicWithInt();
				PrintLn('Should not reach here');
			except
				on E: EHost do
					PrintLn('Caught: ' + E.Message);
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "panic: 42") {
			t.Errorf("expected panic message with '42' in output, got '%s'", output)
		}
	})

	t.Run("CustomTypePanic", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		type CustomError struct {
			Message string
			Code    int
		}

		engine.RegisterFunction("PanicWithCustom", func() string {
			panic(CustomError{Code: 500, Message: "internal error"})
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			try
				var result := PanicWithCustom();
				PrintLn('Should not reach here');
			except
				on E: EHost do
					PrintLn('Caught custom panic');
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "Caught custom panic" {
			t.Errorf("expected 'Caught custom panic', got '%s'", output)
		}
	})

	t.Run("AllPanicsAreCatchable", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("PanicString", func() { panic("string") })
		engine.RegisterFunction("PanicError", func() { panic(errors.New("error")) })
		engine.RegisterFunction("PanicInt", func() { panic(123) })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			var count := 0;

			try
				PanicString();
			except
				on E: EHost do
					count := count + 1;
			end;

			try
				PanicError();
			except
				on E: EHost do
					count := count + 1;
			end;

			try
				PanicInt();
			except
				on E: EHost do
					count := count + 1;
			end;

			PrintLn(IntToStr(count));
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "3" {
			t.Errorf("expected all 3 panics to be caught, got count: %s", output)
		}
	})
}

// TestPanicPropagationNestedFFI tests that panics propagate correctly through nested FFI calls.
func TestPanicPropagationNestedFFI(t *testing.T) {
	t.Run("MultipleFFICallsWithPanic", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Register functions where one will panic
		engine.RegisterFunction("SafeFunc1", func() string { return "safe1" })
		engine.RegisterFunction("SafeFunc2", func() string { return "safe2" })
		engine.RegisterFunction("PanicFunc", func() string { panic("boom") })
		engine.RegisterFunction("SafeFunc3", func() string { return "safe3" })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		// Call multiple FFI functions, one of which panics
		_, err := engine.Eval(`
			try
				var r1 := SafeFunc1();
				PrintLn(r1);

				var r2 := SafeFunc2();
				PrintLn(r2);

				var r3 := PanicFunc();
				PrintLn('Should not reach here');

				var r4 := SafeFunc3();
				PrintLn(r4);
			except
				on E: EHost do
					PrintLn('Caught panic in chain');
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
		}
		if lines[0] != "safe1" || lines[1] != "safe2" || lines[2] != "Caught panic in chain" {
			t.Errorf("unexpected output: %v", lines)
		}
	})

	t.Run("PanicInNestedGoFunctions", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Helper function that panics
		innerFunc := func() {
			panic("inner panic")
		}

		// Register a function that calls another Go function internally
		engine.RegisterFunction("OuterFunc", func() string {
			// This calls another Go function which panics
			innerFunc()
			return "never reached"
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			try
				var result := OuterFunc();
				PrintLn('Should not reach here');
			except
				on E: EHost do
					PrintLn('Caught: inner panic propagated');
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "inner panic") {
			t.Errorf("expected panic to propagate, got '%s'", output)
		}
	})

	t.Run("CallStackPreservation", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("DeepFunc", func() { panic("deep error") })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		// Multiple nested try/except blocks
		_, err := engine.Eval(`
			var outerCaught := false;
			var innerCaught := false;

			try
				try
					DeepFunc();
					PrintLn('Should not reach here');
				except
					on E: EHost do begin
						innerCaught := true;
						PrintLn('Inner caught');
						raise; // Re-raise to outer
					end;
				end;
			except
				on E: EHost do begin
					outerCaught := true;
					PrintLn('Outer caught');
				end;
			end;

			if innerCaught and outerCaught then
				PrintLn('Both caught correctly')
			else
				PrintLn('ERROR: exception not propagated correctly');
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "Both caught correctly") {
			t.Errorf("expected proper exception propagation, got '%s'", output)
		}
	})
}

// TestFinallyBlocksWithPanics tests that finally blocks execute correctly when FFI functions panic.
func TestFinallyBlocksWithPanics(t *testing.T) {
	t.Run("FinallyExecutesOnPanic", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("PanicFunc", func() { panic("oops") })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		// When there's no except block, the panic will propagate as an error after finally executes
		_, err := engine.Eval(`
			try
				PanicFunc();
				PrintLn('Should not reach here');
			finally
				PrintLn('Finally executed');
			end;
		`)

		// Panic should propagate as an error since it's not caught
		if err == nil {
			t.Error("expected error from uncaught panic")
		}

		// But finally should have executed before the error propagated
		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "Finally executed") {
			t.Errorf("expected finally block to execute before error, got '%s'", output)
		}
	})

	t.Run("PanicPropagatesAfterFinally", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("PanicFunc", func() { panic("error") })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			var outerFinally := false;

			try
				try
					PanicFunc();
					PrintLn('Should not reach here');
				finally
					PrintLn('Inner finally');
				end;
			except
				on E: EHost do
					PrintLn('Outer caught');
			finally
				outerFinally := true;
				PrintLn('Outer finally');
			end;

			if outerFinally then
				PrintLn('All finally blocks executed');
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		lines := strings.Split(output, "\n")

		// Should see: Inner finally, Outer caught, Outer finally, All finally blocks executed
		if len(lines) < 4 {
			t.Fatalf("expected at least 4 lines, got %d: %v", len(lines), lines)
		}
		if !strings.Contains(output, "Inner finally") {
			t.Errorf("expected inner finally to execute, got '%s'", output)
		}
		if !strings.Contains(output, "Outer caught") {
			t.Errorf("expected exception to be caught, got '%s'", output)
		}
		if !strings.Contains(output, "Outer finally") {
			t.Errorf("expected outer finally to execute, got '%s'", output)
		}
		if !strings.Contains(output, "All finally blocks executed") {
			t.Errorf("expected all blocks to execute, got '%s'", output)
		}
	})

	t.Run("FinallyWithExceptCatchesPanic", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("PanicFunc", func() { panic("caught error") })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			var finallyExecuted := false;
			var exceptExecuted := false;

			try
				PanicFunc();
				PrintLn('Should not reach here');
			except
				on E: EHost do begin
					exceptExecuted := true;
					PrintLn('Exception caught');
				end;
			finally
				finallyExecuted := true;
				PrintLn('Finally executed');
			end;

			if exceptExecuted and finallyExecuted then
				PrintLn('Both except and finally executed');
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "Exception caught") {
			t.Errorf("expected exception to be caught, got '%s'", output)
		}
		if !strings.Contains(output, "Finally executed") {
			t.Errorf("expected finally to execute, got '%s'", output)
		}
		if !strings.Contains(output, "Both except and finally executed") {
			t.Errorf("expected both blocks to execute, got '%s'", output)
		}
	})

	t.Run("FinallyExecutesEvenWithMultiplePanics", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		callCount := 0
		engine.RegisterFunction("CountedPanic", func() {
			callCount++
			panic("panic")
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			var count := 0;

			try
				CountedPanic();
			except
				on E: EHost do
					count := count + 1;
			finally
				PrintLn('Finally 1');
			end;

			try
				CountedPanic();
			except
				on E: EHost do
					count := count + 1;
			finally
				PrintLn('Finally 2');
			end;

			try
				CountedPanic();
			except
				on E: EHost do
					count := count + 1;
			finally
				PrintLn('Finally 3');
			end;

			PrintLn('Count: ' + IntToStr(count));
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "Finally 1") || !strings.Contains(output, "Finally 2") || !strings.Contains(output, "Finally 3") {
			t.Errorf("expected all finally blocks to execute, got '%s'", output)
		}
		if !strings.Contains(output, "Count: 3") {
			t.Errorf("expected all exceptions to be caught, got '%s'", output)
		}
		if callCount != 3 {
			t.Errorf("expected function to be called 3 times, got %d", callCount)
		}
	})

	t.Run("FinallyWithUncaughtPanic", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("PanicFunc", func() { panic("uncaught") })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		// This should execute finally but then propagate the panic as an error
		_, err := engine.Eval(`
			try
				PanicFunc();
			finally
				PrintLn('Finally before uncaught');
			end;
		`)

		// The panic should result in an execution error since it's not caught
		if err == nil {
			t.Error("expected execution error from uncaught panic")
		}

		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "Finally before uncaught") {
			t.Errorf("expected finally to execute even with uncaught panic, got '%s'", output)
		}
	})
}

// TestFFICallbackTypeMismatch tests type mismatches in callback signatures.
func TestFFICallbackTypeMismatch(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register function expecting callback with specific signature
	err = engine.RegisterFunction("ProcessWithCallback", func(x int64, callback func(int64) int64) int64 {
		return callback(x)
	})
	if err != nil {
		t.Fatalf("failed to register ProcessWithCallback: %v", err)
	}

	// Test: Callback with wrong parameter type
	// Note: DWScript's marshaling may attempt type coercion (int to string), which could work
	// The important thing is that the FFI handles it gracefully without crashing
	var buf bytes.Buffer
	engine.SetOutput(&buf)

	result, err := engine.Eval(`
		function WrongParamType(s: String): Integer;
		begin
			// DWScript may convert int64 to string automatically
			PrintLn('Callback called with: ' + s);
			Result := 42;
		end;

		try
			var x := ProcessWithCallback(10, @WrongParamType);
			PrintLn('Result: ' + IntToStr(x));
		except
			on E: EHost do begin
				// Marshaling error if type conversion fails
				PrintLn('Caught marshaling error');
			end;
		end;
	`)
	if err != nil {
		// May fail at eval time with semantic error - that's also valid
		return
	}

	if !result.Success {
		t.Fatal("execution was not successful")
	}

	// The FFI should handle type mismatches gracefully (either by coercion or error)
	// The important thing is that it doesn't crash
	output := buf.String()
	if !strings.Contains(output, "Result:") && !strings.Contains(output, "Caught marshaling error") {
		t.Errorf("expected result or error, got: %s", output)
	}
}

// TestFFIComplexErrorPropagation tests error propagation through multiple FFI layers.
func TestFFIComplexErrorPropagation(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register Go function that calls callback which might error
	err = engine.RegisterFunction("SafeProcess", func(x int64, processor func(int64) (int64, error)) (int64, error) {
		return processor(x)
	})
	if err != nil {
		t.Fatalf("failed to register SafeProcess: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)

	// Test error propagation from DWScript → Go → DWScript (with error)
	result, err := engine.Eval(`
		function MaybeError(x: Integer): Integer;
		begin
			if x < 0 then
				raise Exception.Create('Negative not allowed');
			Result := x * 2;
		end;

		try
			var result := SafeProcess(10, @MaybeError);
			PrintLn('Result: ' + IntToStr(result));

			result := SafeProcess(-5, @MaybeError);  // This should error
			PrintLn('Should not reach here');
		except
			on E: EHost do begin
				PrintLn('Caught error: ' + E.Message);
			end;
		end;
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("execution was not successful")
	}

	output := buf.String()
	if !strings.Contains(output, "Result: 20") {
		t.Error("expected successful result for positive input")
	}
	if !strings.Contains(output, "Caught error") {
		t.Error("expected error to be caught for negative input")
	}
}

// TestRegisterFunctionWrongArgCount tests error handling for wrong argument count.
func TestRegisterFunctionWrongArgCount(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register a function expecting 2 arguments
	err = engine.RegisterFunction("Add2", func(a, b int64) int64 {
		return a + b
	})
	if err != nil {
		t.Fatalf("failed to register function: %v", err)
	}

	// Try to call with wrong number of arguments
	var buf bytes.Buffer
	engine.SetOutput(&buf)
	_, err = engine.Eval(`
		try
			var sum := Add2(5);  // Only 1 argument instead of 2
			PrintLn('Should not reach here');
		except
			on E: EHost do
				PrintLn('Caught error');
		end;
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Should have caught the error
	output := strings.TrimSpace(buf.String())
	if !strings.Contains(output, "Caught error") {
		t.Errorf("expected error to be caught, got output: %s", output)
	}
}
