package dwscript

import (
	"errors"
	"math"
	"os"
	"strings"
	"testing"
)

// TestFFIBasicIntegration tests basic FFI functionality using test scripts.
func TestFFIBasicIntegration(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register test functions
	engine.RegisterFunction("Add", func(a, b int64) int64 { return a + b })
	engine.RegisterFunction("Multiply", func(a, b int64) int64 { return a * b })
	engine.RegisterFunction("ToUpper", func(s string) string { return strings.ToUpper(s) })
	engine.RegisterFunction("ToLower", func(s string) string { return strings.ToLower(s) })
	engine.RegisterFunction("IsEven", func(n int64) bool { return n%2 == 0 })
	engine.RegisterFunction("CircleArea", func(r float64) float64 { return math.Pi * r * r })

	// Run test script
	runTestScript(t, engine, "../../testdata/ffi/basic_ffi.dws", "../../testdata/ffi/basic_ffi.expected")
}

// TestFFIArrayIntegration tests array passing with FFI using test scripts.
func TestFFIArrayIntegration(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register array functions
	engine.RegisterFunction("SumInts", func(nums []int64) int64 {
		sum := int64(0)
		for _, n := range nums {
			sum += n
		}
		return sum
	})

	engine.RegisterFunction("GetSquares", func(n int64) []int64 {
		result := make([]int64, n)
		for i := int64(0); i < n; i++ {
			result[i] = i * i
		}
		return result
	})

	engine.RegisterFunction("FilterEvens", func(nums []int64) []int64 {
		result := []int64{}
		for _, n := range nums {
			if n%2 == 0 {
				result = append(result, n)
			}
		}
		return result
	})

	engine.RegisterFunction("JoinStrings", func(parts []string, sep string) string {
		return strings.Join(parts, sep)
	})

	engine.RegisterFunction("SplitString", func(s, sep string) []string {
		return strings.Split(s, sep)
	})

	engine.RegisterFunction("Contains", func(s, substr string) bool {
		return strings.Contains(s, substr)
	})

	// Run test script
	runTestScript(t, engine, "../../testdata/ffi/array_passing.dws", "../../testdata/ffi/array_passing.expected")
}

// TestFFIErrorIntegration tests error handling with FFI using test scripts.
func TestFFIErrorIntegration(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register error-handling functions
	engine.RegisterFunction("DivideInts", func(a, b int64) (int64, error) {
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	})

	engine.RegisterFunction("TriggerPanic", func() {
		panic("intentional panic for testing")
	})

	engine.RegisterFunction("NestedError", func() error {
		return errors.New("nested error")
	})

	engine.RegisterFunction("Contains", func(s, substr string) bool {
		return strings.Contains(s, substr)
	})

	// Run test script
	runTestScript(t, engine, "../../testdata/ffi/error_handling.dws", "../../testdata/ffi/error_handling.expected")
}

// TestFFIRealGoFunctions tests FFI with real-world Go standard library functions.
func TestFFIRealGoFunctions(t *testing.T) {
	t.Run("StringsPackage", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Set up output capture
		var buf strings.Builder
		engine.SetOutput(&buf)

		// Real Go strings functions
		engine.RegisterFunction("StrContains", strings.Contains)
		engine.RegisterFunction("StrHasPrefix", strings.HasPrefix)
		engine.RegisterFunction("StrHasSuffix", strings.HasSuffix)
		engine.RegisterFunction("StrSplit", strings.Split)
		engine.RegisterFunction("StrJoin", strings.Join)
		engine.RegisterFunction("StrToUpper", strings.ToUpper)
		engine.RegisterFunction("StrToLower", strings.ToLower)
		engine.RegisterFunction("StrTrim", strings.TrimSpace)

		result, err := engine.Eval(`
			// Test Contains
			if not StrContains('hello world', 'world') then
				raise Exception.Create('Contains failed');

			// Test HasPrefix
			if not StrHasPrefix('hello', 'he') then
				raise Exception.Create('HasPrefix failed');

			// Test HasSuffix
			if not StrHasSuffix('world', 'ld') then
				raise Exception.Create('HasSuffix failed');

			// Test Split
			var parts := StrSplit('a,b,c', ',');
			if Length(parts) <> 3 then
				raise Exception.Create('Split length failed');

			// Test Join
			var joined := StrJoin(parts, '-');
			if joined <> 'a-b-c' then
				raise Exception.Create('Join failed');

			// Test case conversion
			if StrToUpper('abc') <> 'ABC' then
				raise Exception.Create('ToUpper failed');
			if StrToLower('XYZ') <> 'xyz' then
				raise Exception.Create('ToLower failed');

			// Test Trim
			if StrTrim('  hello  ') <> 'hello' then
				raise Exception.Create('Trim failed');

			PrintLn('Strings package tests passed');
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}
		if !result.Success {
			t.Fatal("test script failed")
		}

		expected := "Strings package tests passed\n"
		output := buf.String()
		if output != expected {
			t.Errorf("expected output %q, got %q", expected, output)
		}
	})

	t.Run("MathPackage", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Set up output capture
		var buf strings.Builder
		engine.SetOutput(&buf)

		// Real Go math functions
		engine.RegisterFunction("MathAbs", math.Abs)
		engine.RegisterFunction("MathCeil", math.Ceil)
		engine.RegisterFunction("MathFloor", math.Floor)
		engine.RegisterFunction("MathMax", math.Max)
		engine.RegisterFunction("MathMin", math.Min)
		engine.RegisterFunction("MathSqrt", math.Sqrt)
		engine.RegisterFunction("MathPow", math.Pow)

		result, err := engine.Eval(`
			// Test Abs
			if MathAbs(-5.0) <> 5.0 then
				raise Exception.Create('Abs failed');

			// Test Ceil
			if MathCeil(3.2) <> 4.0 then
				raise Exception.Create('Ceil failed');

			// Test Floor
			if MathFloor(3.8) <> 3.0 then
				raise Exception.Create('Floor failed');

			// Test Max
			if MathMax(5.0, 3.0) <> 5.0 then
				raise Exception.Create('Max failed');

			// Test Min
			if MathMin(5.0, 3.0) <> 3.0 then
				raise Exception.Create('Min failed');

			// Test Sqrt
			if MathSqrt(16.0) <> 4.0 then
				raise Exception.Create('Sqrt failed');

			// Test Pow
			if MathPow(2.0, 3.0) <> 8.0 then
				raise Exception.Create('Pow failed');

			PrintLn('Math package tests passed');
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}
		if !result.Success {
			t.Fatal("test script failed")
		}

		expected := "Math package tests passed\n"
		output := buf.String()
		if output != expected {
			t.Errorf("expected output %q, got %q", expected, output)
		}
	})

	t.Run("FileOperations", func(t *testing.T) {
		engine, _ := New(WithTypeCheck(false))

		// Set up output capture
		var buf strings.Builder
		engine.SetOutput(&buf)

		// Real Go file functions
		engine.RegisterFunction("FileReadString", func(path string) (string, error) {
			data, err := os.ReadFile(path)
			return string(data), err
		})

		engine.RegisterFunction("FileWriteString", func(path, content string) error {
			return os.WriteFile(path, []byte(content), 0644)
		})

		engine.RegisterFunction("FileExists", func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		})

		engine.RegisterFunction("FileRemove", os.Remove)

		result, err := engine.Eval(`
			var testFile := '/tmp/ffi_test_file.txt';
			var testContent := 'Hello from FFI!';

			// Write file
			try
				FileWriteString(testFile, testContent);
			except
				on E: EHost do begin
					raise Exception.Create('Write failed: ' + E.Message);
				end;
			end;

			// Check exists
			if not FileExists(testFile) then
				raise Exception.Create('File should exist');

			// Read file
			var content := '';
			try
				content := FileReadString(testFile);
			except
				on E: EHost do begin
					raise Exception.Create('Read failed: ' + E.Message);
				end;
			end;

			if content <> testContent then
				raise Exception.Create('Content mismatch');

			// Clean up
			try
				FileRemove(testFile);
			except
				on E: EHost do begin
					raise Exception.Create('Remove failed: ' + E.Message);
				end;
			end;

			// Verify removed
			if FileExists(testFile) then
				raise Exception.Create('File should be removed');

			PrintLn('File operations tests passed');
		`)
		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}
		if !result.Success {
			t.Fatal("test script failed")
		}

		expected := "File operations tests passed\n"
		output := buf.String()
		if output != expected {
			t.Errorf("expected output %q, got %q", expected, output)
		}
	})
}

// Helper function to run test script and compare output
func runTestScript(t *testing.T, engine *Engine, scriptPath, expectedPath string) {
	t.Helper()

	// Read script
	scriptData, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read script %s: %v", scriptPath, err)
	}

	// Read expected output
	expectedData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected output %s: %v", expectedPath, err)
	}
	expected := string(expectedData)

	// Set up output capture
	var buf strings.Builder
	engine.SetOutput(&buf)

	// Run script
	result, err := engine.Eval(string(scriptData))
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !result.Success {
		t.Fatal("script execution was not successful")
	}

	// Compare output from buffer
	output := buf.String()
	if output != expected {
		t.Errorf("output mismatch:\nexpected: %q\ngot:      %q", expected, output)
	}
}
