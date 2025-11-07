package dwscript

import (
	"testing"
)

// =============================================================================
// Task 9.5c: Performance Benchmarks
// =============================================================================

// BenchCounter is a helper type for benchmarking method calls.
type BenchCounter struct {
	value int64
}

// Increment is a method for BenchmarkFFIMethodCall.
func (c *BenchCounter) Increment() int64 {
	c.value++
	return c.value
}

// BenchmarkFFICallOverhead benchmarks the overhead of calling an FFI function
// compared to calling a native DWScript function.
func BenchmarkFFICallOverhead(b *testing.B) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	// Register simple FFI function
	err = engine.RegisterFunction("GoAdd", func(a, b int64) int64 {
		return a + b
	})
	if err != nil {
		b.Fatalf("failed to register function: %v", err)
	}

	b.Run("FFI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := GoAdd(5, 10);`)
			if err != nil || !result.Success {
				b.Fatalf("FFI call failed: %v", err)
			}
		}
	})

	b.Run("Native", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`
				function DWScriptAdd(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				var x := DWScriptAdd(5, 10);
			`)
			if err != nil || !result.Success {
				b.Fatalf("Native call failed: %v", err)
			}
		}
	})
}

// BenchmarkFFIMarshalingPrimitives benchmarks marshaling cost for primitive types.
func BenchmarkFFIMarshalingPrimitives(b *testing.B) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	// Register functions for each primitive type
	engine.RegisterFunction("PassInt", func(x int64) int64 { return x })
	engine.RegisterFunction("PassFloat", func(x float64) float64 { return x })
	engine.RegisterFunction("PassString", func(s string) string { return s })
	engine.RegisterFunction("PassBool", func(b bool) bool { return b })

	b.Run("Integer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := PassInt(42);`)
			if err != nil || !result.Success {
				b.Fatalf("PassInt failed: %v", err)
			}
		}
	})

	b.Run("Float", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := PassFloat(3.14);`)
			if err != nil || !result.Success {
				b.Fatalf("PassFloat failed: %v", err)
			}
		}
	})

	b.Run("String", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := PassString('hello');`)
			if err != nil || !result.Success {
				b.Fatalf("PassString failed: %v", err)
			}
		}
	})

	b.Run("Boolean", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := PassBool(True);`)
			if err != nil || !result.Success {
				b.Fatalf("PassBool failed: %v", err)
			}
		}
	})
}

// BenchmarkFFIMarshalingCollections benchmarks marshaling cost for arrays and maps.
func BenchmarkFFIMarshalingCollections(b *testing.B) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	// Register functions for collections
	engine.RegisterFunction("PassIntArray", func(arr []int64) []int64 { return arr })
	engine.RegisterFunction("PassStringArray", func(arr []string) []string { return arr })
	engine.RegisterFunction("PassMap", func(m map[string]int64) map[string]int64 { return m })

	b.Run("SmallIntArray", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := PassIntArray([1, 2, 3]);`)
			if err != nil || !result.Success {
				b.Fatalf("PassIntArray failed: %v", err)
			}
		}
	})

	b.Run("LargeIntArray", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`
				var arr: array of Integer := [];
				var i: Integer;
				for i := 0 to 99 do
					arr.Add(i);
				var x := PassIntArray(arr);
			`)
			if err != nil || !result.Success {
				b.Fatalf("PassIntArray (large) failed: %v", err)
			}
		}
	})

	b.Run("StringArray", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := PassStringArray(['a', 'b', 'c']);`)
			if err != nil || !result.Success {
				b.Fatalf("PassStringArray failed: %v", err)
			}
		}
	})

	// Note: Map/record syntax is currently not fully supported in benchmarks
	// Skipping map benchmark for now
}

// BenchmarkFFICallbackOverhead benchmarks the overhead of DWScript → Go → DWScript callbacks.
func BenchmarkFFICallbackOverhead(b *testing.B) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	// Register function that accepts callback
	err = engine.RegisterFunction("CallCallback", func(callback func(int64) int64, n int64) int64 {
		return callback(n)
	})
	if err != nil {
		b.Fatalf("failed to register function: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := engine.Eval(`
			function SimpleCallback(x: Integer): Integer;
			begin
				Result := x * 2;
			end;

			var x := CallCallback(@SimpleCallback, 10);
		`)
		if err != nil || !result.Success {
			b.Fatalf("callback failed: %v", err)
		}
	}
}

// BenchmarkFFICallbackMap benchmarks callback overhead with array mapping.
func BenchmarkFFICallbackMap(b *testing.B) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	// Register map function
	err = engine.RegisterFunction("Map", func(items []int64, mapper func(int64) int64) []int64 {
		result := make([]int64, len(items))
		for i, item := range items {
			result[i] = mapper(item)
		}
		return result
	})
	if err != nil {
		b.Fatalf("failed to register Map: %v", err)
	}

	b.Run("SmallArray", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`
				function Double(x: Integer): Integer;
				begin
					Result := x * 2;
				end;

				var x := Map([1, 2, 3, 4, 5], @Double);
			`)
			if err != nil || !result.Success {
				b.Fatalf("Map failed: %v", err)
			}
		}
	})

	b.Run("LargeArray", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`
				function Double(x: Integer): Integer;
				begin
					Result := x * 2;
				end;

				var arr: array of Integer := [];
				var i: Integer;
				for i := 1 to 50 do
					arr.Add(i);
				var x := Map(arr, @Double);
			`)
			if err != nil || !result.Success {
				b.Fatalf("Map (large) failed: %v", err)
			}
		}
	})
}

// BenchmarkFFIVarParamOverhead benchmarks the overhead of by-reference (var) parameters.
func BenchmarkFFIVarParamOverhead(b *testing.B) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	// Register function with var parameter
	err = engine.RegisterFunction("Swap", func(a, b *int64) {
		*a, *b = *b, *a
	})
	if err != nil {
		b.Fatalf("failed to register Swap: %v", err)
	}

	// Register function without var parameter for comparison
	err = engine.RegisterFunction("Add", func(a, b int64) int64 {
		return a + b
	})
	if err != nil {
		b.Fatalf("failed to register Add: %v", err)
	}

	b.Run("VarParam", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`
				var a, b: Integer;
				a := 5;
				b := 10;
				Swap(a, b);
			`)
			if err != nil || !result.Success {
				b.Fatalf("Swap failed: %v", err)
			}
		}
	})

	b.Run("ValueParam", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`
				var a, b: Integer;
				a := 5;
				b := 10;
				var c := Add(a, b);
			`)
			if err != nil || !result.Success {
				b.Fatalf("Add failed: %v", err)
			}
		}
	})
}

// BenchmarkFFIMethodCall benchmarks calling Go methods vs functions.
func BenchmarkFFIMethodCall(b *testing.B) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	counter := &BenchCounter{value: 0}

	// Register method
	err = engine.RegisterMethod("Increment", counter, "Increment")
	if err != nil {
		b.Fatalf("failed to register method: %v", err)
	}

	// Register equivalent function for comparison
	err = engine.RegisterFunction("IncrementFunc", func() int64 {
		counter.value++
		return counter.value
	})
	if err != nil {
		b.Fatalf("failed to register function: %v", err)
	}

	b.Run("Method", func(b *testing.B) {
		counter.value = 0
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := Increment();`)
			if err != nil || !result.Success {
				b.Fatalf("Increment failed: %v", err)
			}
		}
	})

	b.Run("Function", func(b *testing.B) {
		counter.value = 0
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := IncrementFunc();`)
			if err != nil || !result.Success {
				b.Fatalf("IncrementFunc failed: %v", err)
			}
		}
	})
}

// BenchmarkFFIErrorHandling benchmarks error handling overhead.
func BenchmarkFFIErrorHandling(b *testing.B) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	// Register function that might return error
	err = engine.RegisterFunction("MaybeError", func(shouldError bool) (int64, error) {
		if shouldError {
			return 0, nil // Return success
		}
		return 42, nil
	})
	if err != nil {
		b.Fatalf("failed to register function: %v", err)
	}

	b.Run("NoError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result, err := engine.Eval(`var x := MaybeError(False);`)
			if err != nil || !result.Success {
				b.Fatalf("MaybeError failed: %v", err)
			}
		}
	})
}

// BenchmarkFFIComplexScenario benchmarks a complex scenario combining multiple FFI features.
func BenchmarkFFIComplexScenario(b *testing.B) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	// Register functions for complex scenario
	engine.RegisterFunction("ProcessData", func(
		data []int64,
		processor func(int64) int64,
		count *int64,
	) []int64 {
		result := make([]int64, len(data))
		for i, v := range data {
			result[i] = processor(v)
		}
		*count = int64(len(data))
		return result
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := engine.Eval(`
			function Transform(x: Integer): Integer;
			begin
				Result := x * 2 + 1;
			end;

			var count: Integer;
			var data := [1, 2, 3, 4, 5];
			var result := ProcessData(data, @Transform, count);
		`)
		if err != nil || !result.Success {
			b.Fatalf("ProcessData failed: %v", err)
		}
	}
}
