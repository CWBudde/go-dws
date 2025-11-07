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

// TestFFIVariadicWithCallbacks tests combining variadic parameters with callbacks.
func TestFFIVariadicWithCallbacks(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	var buf strings.Builder
	engine.SetOutput(&buf)

	// Register a variadic function that accepts a callback
	// Note: We pass nums as a slice instead of individual variadic args since
	// DWScript passes arrays as slices to Go
	engine.RegisterFunction("ProcessInts", func(nums []int64, callback func(int64) int64) []int64 {
		result := make([]int64, len(nums))
		for i, num := range nums {
			result[i] = callback(num)
		}
		return result
	})

	result, err := engine.Eval(`
		// Define a callback that doubles a number
		function Double(n: Integer): Integer;
		begin
			Result := n * 2;
		end;

		// Define a callback that squares a number
		function Square(n: Integer): Integer;
		begin
			Result := n * n;
		end;

		// Test with Double callback
		var doubled := ProcessInts([1, 2, 3, 4, 5], @Double);
		if Length(doubled) <> 5 then
			raise Exception.Create('Wrong length for doubled');
		if doubled[0] <> 2 or doubled[4] <> 10 then
			raise Exception.Create('Double callback failed');

		// Test with Square callback
		var squared := ProcessInts([2, 3, 4], @Square);
		if Length(squared) <> 3 then
			raise Exception.Create('Wrong length for squared');
		if squared[0] <> 4 or squared[1] <> 9 or squared[2] <> 16 then
			raise Exception.Create('Square callback failed');

		// Test with inline lambda
		var plusTen := ProcessInts([5, 15, 25], lambda(n: Integer): Integer begin
			Result := n + 10;
		end);
		if plusTen[0] <> 15 or plusTen[1] <> 25 or plusTen[2] <> 35 then
			raise Exception.Create('Lambda callback failed');

		PrintLn('Variadic with callbacks test passed');
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("test script failed")
	}

	expected := "Variadic with callbacks test passed\n"
	output := buf.String()
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestFFIVariadicWithVarParams tests combining variadic parameters with by-reference params.
func TestFFIVariadicWithVarParams(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	var buf strings.Builder
	engine.SetOutput(&buf)

	// Register a variadic function that uses a var param for output
	engine.RegisterFunction("SumWithCount", func(count *int64, nums ...int64) int64 {
		sum := int64(0)
		for _, n := range nums {
			sum += n
		}
		*count = int64(len(nums))
		return sum
	})

	// Register another function that modifies multiple var params
	engine.RegisterFunction("MinMaxSum", func(min, max, sum *int64, nums ...int64) {
		if len(nums) == 0 {
			*min, *max, *sum = 0, 0, 0
			return
		}
		*min, *max, *sum = nums[0], nums[0], int64(0)
		for _, n := range nums {
			if n < *min {
				*min = n
			}
			if n > *max {
				*max = n
			}
			*sum += n
		}
	})

	result, err := engine.Eval(`
		// Test SumWithCount
		var count: Integer := 0;
		var total := SumWithCount(count, 10, 20, 30, 40, 50);
		if total <> 150 then
			raise Exception.Create('Sum incorrect');
		if count <> 5 then
			raise Exception.Create('Count incorrect');

		// Test with different number of arguments
		count := 0;
		total := SumWithCount(count, 100, 200);
		if total <> 300 then
			raise Exception.Create('Sum incorrect for 2 args');
		if count <> 2 then
			raise Exception.Create('Count incorrect for 2 args');

		// Test MinMaxSum
		var minVal, maxVal, sumVal: Integer;
		MinMaxSum(minVal, maxVal, sumVal, 5, 2, 9, 1, 7);
		if minVal <> 1 then
			raise Exception.Create('Min incorrect: ' + IntToStr(minVal));
		if maxVal <> 9 then
			raise Exception.Create('Max incorrect: ' + IntToStr(maxVal));
		if sumVal <> 24 then
			raise Exception.Create('Sum incorrect: ' + IntToStr(sumVal));

		PrintLn('Variadic with var params test passed');
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("test script failed")
	}

	expected := "Variadic with var params test passed\n"
	output := buf.String()
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// DataProcessor is a helper type for testing methods with callbacks.
type DataProcessor struct {
	data []int64
}

// ProcessEach applies callback to each element (helper for test)
func (dp *DataProcessor) ProcessEach(callback func(int64) int64) []int64 {
	result := make([]int64, len(dp.data))
	for i, v := range dp.data {
		result[i] = callback(v)
	}
	return result
}

// FilterWith filters elements using callback predicate (helper for test)
func (dp *DataProcessor) FilterWith(predicate func(int64) bool) []int64 {
	result := []int64{}
	for _, v := range dp.data {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// FindFirst returns first element matching predicate (helper for test)
func (dp *DataProcessor) FindFirst(predicate func(int64) bool) int64 {
	for _, v := range dp.data {
		if predicate(v) {
			return v
		}
	}
	return 0
}

// ComplexProcessor is a helper type for testing complex feature combinations.
type ComplexProcessor struct {
	name string
}

// ProcessWithStats processes variadic args with callback and reports stats via var params
func (cp *ComplexProcessor) ProcessWithStats(
	callback func(int64) int64,
	count *int64,
	sum *int64,
	nums ...int64,
) []int64 {
	result := make([]int64, len(nums))
	*count = int64(len(nums))
	*sum = int64(0)
	for i, n := range nums {
		result[i] = callback(n)
		*sum += result[i]
	}
	return result
}

// TestFFIMethodsWithCallbacks tests calling Go methods that accept callbacks.
func TestFFIMethodsWithCallbacks(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	var buf strings.Builder
	engine.SetOutput(&buf)

	processor := &DataProcessor{data: []int64{1, 2, 3, 4, 5}}

	// Method that applies a callback to each element
	engine.RegisterMethod("ProcessEach", processor, "ProcessEach")
	// Method that filters using a callback predicate
	engine.RegisterMethod("FilterWith", processor, "FilterWith")
	// Method that finds first matching element
	engine.RegisterMethod("FindFirst", processor, "FindFirst")

	result, err := engine.Eval(`
		// Define a callback that checks if a number is even
		function IsEven(n: Integer): Boolean;
		begin
			Result := (n mod 2) = 0;
		end;

		// Define a callback that checks if a number is > 3
		function GreaterThanThree(n: Integer): Boolean;
		begin
			Result := n > 3;
		end;

		// Test FilterWith (should return [2, 4])
		var evens := FilterWith(@IsEven);
		if Length(evens) <> 2 then
			raise Exception.Create('FilterWith length wrong');
		if evens[0] <> 2 or evens[1] <> 4 then
			raise Exception.Create('FilterWith values wrong');

		// Test FilterWith with different predicate (should return [4, 5])
		var greaterThanThree := FilterWith(@GreaterThanThree);
		if Length(greaterThanThree) <> 2 then
			raise Exception.Create('FilterWith >3 length wrong');
		if greaterThanThree[0] <> 4 or greaterThanThree[1] <> 5 then
			raise Exception.Create('FilterWith >3 values wrong');

		// Test FindFirst (should return 2, the first even number)
		var firstEven := FindFirst(@IsEven);
		if firstEven <> 2 then
			raise Exception.Create('FindFirst failed');

		PrintLn('Methods with callbacks test passed');
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("test script failed")
	}

	expected := "Methods with callbacks test passed\n"
	output := buf.String()
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestFFIComplexCombination tests combining 3+ features: methods, var params, and callbacks.
func TestFFIComplexCombination(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	var buf strings.Builder
	engine.SetOutput(&buf)

	processor := &ComplexProcessor{name: "TestProcessor"}

	// Method with var param and callback
	engine.RegisterMethod("ProcessWithStats", processor, "ProcessWithStats")

	result, err := engine.Eval(`
		// Define a callback
		function Triple(n: Integer): Integer;
		begin
			Result := n * 3;
		end;

		// Test ProcessWithStats: passes data through callback and reports stats
		var count, sum: Integer;
		var processed := ProcessWithStats(@Triple, count, sum, 1, 2, 3, 4);

		if Length(processed) <> 4 then
			raise Exception.Create('Processed length wrong');
		if processed[0] <> 3 or processed[3] <> 12 then
			raise Exception.Create('Processed values wrong');
		if count <> 4 then
			raise Exception.Create('Count wrong: ' + IntToStr(count));
		if sum <> 30 then // 3+6+9+12 = 30
			raise Exception.Create('Sum wrong: ' + IntToStr(sum));

		PrintLn('Complex combination test passed');
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("test script failed")
	}

	expected := "Complex combination test passed\n"
	output := buf.String()
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestFFINestedCallbacks tests deeply nested callback chains.
func TestFFINestedCallbacks(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	var buf strings.Builder
	engine.SetOutput(&buf)

	// Function that takes a callback and applies it twice (nesting)
	engine.RegisterFunction("ApplyTwice", func(callback func(int64) int64, n int64) int64 {
		return callback(callback(n))
	})

	// Function that takes a callback and a number of times to apply it
	engine.RegisterFunction("ApplyNTimes", func(callback func(int64) int64, n int64, times int64) int64 {
		result := n
		for i := int64(0); i < times; i++ {
			result = callback(result)
		}
		return result
	})

	// Function that takes two callbacks and composes them
	engine.RegisterFunction("Compose", func(f func(int64) int64, g func(int64) int64, n int64) int64 {
		return f(g(n))
	})

	result, err := engine.Eval(`
		function AddOne(n: Integer): Integer;
		begin
			Result := n + 1;
		end;

		function Double(n: Integer): Integer;
		begin
			Result := n * 2;
		end;

		// Test ApplyTwice: AddOne twice = +2
		var result := ApplyTwice(@AddOne, 5);
		if result <> 7 then
			raise Exception.Create('ApplyTwice failed');

		// Test ApplyNTimes: AddOne 5 times = +5
		result := ApplyNTimes(@AddOne, 10, 5);
		if result <> 15 then
			raise Exception.Create('ApplyNTimes failed');

		// Test Compose: Double then AddOne: (5*2)+1 = 11
		result := Compose(@AddOne, @Double, 5);
		if result <> 11 then
			raise Exception.Create('Compose failed');

		// Test Compose other direction: AddOne then Double: (5+1)*2 = 12
		result := Compose(@Double, @AddOne, 5);
		if result <> 12 then
			raise Exception.Create('Compose reverse failed');

		PrintLn('Nested callbacks test passed');
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("test script failed")
	}

	expected := "Nested callbacks test passed\n"
	output := buf.String()
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestFFICallbacksWithErrors tests error propagation through callbacks.
func TestFFICallbacksWithErrors(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	var buf strings.Builder
	engine.SetOutput(&buf)

	// Function that applies callback and handles errors
	engine.RegisterFunction("SafeApply", func(callback func(int64) (int64, error), n int64) (int64, error) {
		return callback(n)
	})

	// Function that applies callback to array elements, stops on first error
	engine.RegisterFunction("MapUntilError", func(callback func(int64) (int64, error), nums []int64) ([]int64, error) {
		result := make([]int64, 0, len(nums))
		for _, n := range nums {
			val, err := callback(n)
			if err != nil {
				return result, err
			}
			result = append(result, val)
		}
		return result, nil
	})

	result, err := engine.Eval(`
		function CheckedDivide(n: Integer): Integer;
		begin
			if n = 0 then
				raise Exception.Create('Cannot divide by zero');
			Result := 100 div n;
		end;

		// Test SafeApply with valid input
		var result := SafeApply(@CheckedDivide, 5);
		if result <> 20 then
			raise Exception.Create('SafeApply valid case failed');

		// Test SafeApply with error
		try
			result := SafeApply(@CheckedDivide, 0);
			raise Exception.Create('Should have raised exception');
		except
			on E: EHost do begin
				if not E.Message.Contains('Cannot divide by zero') then
					raise Exception.Create('Wrong error message');
			end;
		end;

		// Test MapUntilError with valid inputs
		var mapped := MapUntilError(@CheckedDivide, [5, 4, 2, 1]);
		if Length(mapped) <> 4 then
			raise Exception.Create('MapUntilError length wrong');

		// Test MapUntilError with error in middle
		try
			mapped := MapUntilError(@CheckedDivide, [5, 4, 0, 2]);
			raise Exception.Create('MapUntilError should have raised exception');
		except
			on E: EHost do begin
				// Error should occur, this is expected
			end;
		end;

		PrintLn('Callbacks with errors test passed');
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !result.Success {
		t.Fatal("test script failed")
	}

	expected := "Callbacks with errors test passed\n"
	output := buf.String()
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
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
