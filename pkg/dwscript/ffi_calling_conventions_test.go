package dwscript

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestCallingConventions tests the FFI calling convention design.
func TestCallingConventions(t *testing.T) {
	t.Run("TypeSafeMarshaling", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Register function with specific types
		engine.RegisterFunction("TypedFunc", func(i int64, f float64, s string, b bool) string {
			return fmt.Sprintf("i=%d f=%.1f s=%s b=%v", i, f, s, b)
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)
		_, err := engine.Eval(`
			var result := TypedFunc(42, 3.14, 'hello', true);
			PrintLn(result);
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		expected := "i=42 f=3.1 s=hello b=true"
		if output != expected {
			t.Errorf("expected '%s', got '%s'", expected, output)
		}
	})

	t.Run("ErrorReturnsAsExceptions", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Function that returns error
		engine.RegisterFunction("MayFail", func(shouldFail bool) (string, error) {
			if shouldFail {
				return "", errors.New("intentional failure")
			}
			return "success", nil
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		// Test success case
		_, err := engine.Eval(`
			var result := MayFail(false);
			PrintLn(result);
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "success" {
			t.Errorf("expected 'success', got '%s'", output)
		}

		// Test error case - should be caught as exception
		buf.Reset()
		_, err = engine.Eval(`
			try
				var result := MayFail(true);
				PrintLn('Should not reach here');
			except
				on E: EHost do
					PrintLn('Caught: ' + E.Message);
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output = strings.TrimSpace(buf.String())
		if !strings.Contains(output, "intentional failure") {
			t.Errorf("expected error message in output, got '%s'", output)
		}
	})

	t.Run("PanicRecovery", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Function that panics
		engine.RegisterFunction("MayPanic", func(shouldPanic bool) string {
			if shouldPanic {
				panic("intentional panic")
			}
			return "ok"
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		// Panic should be caught and converted to exception
		_, err := engine.Eval(`
			try
				var result := MayPanic(true);
				PrintLn('Should not reach here');
			except
				on E: EHost do
					PrintLn('Caught panic');
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "Caught panic" {
			t.Errorf("expected 'Caught panic', got '%s'", output)
		}
	})

	t.Run("VariadicLikeBehavior", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Use slice for variadic-like behavior
		engine.RegisterFunction("SumAll", func(numbers []int64) int64 {
			sum := int64(0)
			for _, n := range numbers {
				sum += n
			}
			return sum
		})

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		// Can pass arrays of different lengths
		_, err := engine.Eval(`
			var sum1 := SumAll([1, 2, 3]);
			var sum2 := SumAll([10, 20, 30, 40]);
			var emptyArr: array of Integer := [];
			var sum3 := SumAll(emptyArr);
			PrintLn(IntToStr(sum1));
			PrintLn(IntToStr(sum2));
			PrintLn(IntToStr(sum3));
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(lines))
		}
		if lines[0] != "6" || lines[1] != "100" || lines[2] != "0" {
			t.Errorf("unexpected sums: %v", lines)
		}
	})

	t.Run("MultipleReturnSignatures", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Just T
		engine.RegisterFunction("ReturnValue", func() int64 { return 42 })

		// (T, error) with nil error
		engine.RegisterFunction("ReturnValueNoError", func() (int64, error) { return 99, nil })

		// Just error (procedure)
		callCount := 0
		engine.RegisterFunction("ProcWithError", func() error {
			callCount++
			return nil
		})

		// No return (void procedure)
		voidCallCount := 0
		engine.RegisterFunction("VoidProc", func() { voidCallCount++ })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		_, err := engine.Eval(`
			var v1 := ReturnValue();
			var v2 := ReturnValueNoError();
			ProcWithError();
			VoidProc();
			PrintLn(IntToStr(v1));
			PrintLn(IntToStr(v2));
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}
		if lines[0] != "42" || lines[1] != "99" {
			t.Errorf("unexpected values: %v", lines)
		}
		if callCount != 1 {
			t.Errorf("expected ProcWithError to be called once, got %d", callCount)
		}
		if voidCallCount != 1 {
			t.Errorf("expected VoidProc to be called once, got %d", voidCallCount)
		}
	})

	t.Run("ArgumentValidation", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("RequiresTwoArgs", func(a, b int64) int64 { return a + b })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		// Wrong argument count should raise exception
		_, err := engine.Eval(`
			try
				var result := RequiresTwoArgs(1, 2, 3);  // Too many args
				PrintLn('Should not reach here');
			except
				on E: EHost do
					PrintLn('Caught argument error');
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "Caught argument error" {
			t.Errorf("expected 'Caught argument error', got '%s'", output)
		}
	})

	t.Run("TypeMismatchValidation", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		engine.RegisterFunction("RequiresInt", func(n int64) int64 { return n * 2 })

		var buf bytes.Buffer
		engine.SetOutput(&buf)

		// Wrong type should raise exception
		_, err := engine.Eval(`
			try
				var result := RequiresInt('not an int');
				PrintLn('Should not reach here');
			except
				on E: EHost do
					PrintLn('Caught type error');
			end;
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "Caught type error" {
			t.Errorf("expected 'Caught type error', got '%s'", output)
		}
	})
}
