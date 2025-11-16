# Benchmark Summary - Phase 3.1

**Date**: 2025-11-16
**Purpose**: Establish performance baseline before Phase 3 refactoring
**Total Benchmarks**: 88

## Benchmark Coverage

### High-Level Execution (21 benchmarks)
- Full interpreter pipeline (Fibonacci, Primes, Loops)
- Statement execution (For, While, Repeat, Nested)
- Arithmetic operations
- String operations
- Array operations
- Class operations
- Record operations
- Conditionals
- Function calls
- Variable access (local vs global)
- Full pipeline (tiny/small/medium programs)

### Property Access (4 benchmarks)
- Field-backed property read
- Field-backed property write
- Method-backed property read
- Method-backed property write

### Built-in Functions (16 benchmarks)
**String functions** (6):
- Length, Pos, Copy
- UpperCase, LowerCase, Trim

**Math functions** (6):
- Abs, Min, Max
- Sqr, Sqrt, Power

**Array functions** (3):
- SetLength, Low, High

**Conversion functions** (4):
- IntToStr, StrToInt
- FloatToStr, StrToFloat

### Exception Handling (3 benchmarks)
- Exception raise and catch
- Try-except overhead
- Try-finally overhead

### Type Operations (3 benchmarks)
- Type checks (`is` operator)
- Type casts (`as` operator)
- TypeOf operation

### Variant Operations (2 benchmarks)
- Variant assignment (type changes)
- Variant arithmetic

### Interface Operations (2 benchmarks)
- Interface method calls
- Interface `implements` check

### Enum and Set Operations (2 benchmarks)
- Enum operations (Ord, type casting)
- Set inclusion (`in` operator)

### Object Operations (3 benchmarks)
- Object creation
- Object with constructor
- Inherited method calls

### Lambda and Function Pointers (2 benchmarks)
- Lambda function calls
- Function pointer calls

### Set Operations (30 benchmarks)
**Small sets (≤64 elements, bitmask)** (5):
- Union, Intersection, Difference
- Membership testing
- For-in iteration

**Large sets (>64 elements, map)** (15):
- Union (100, 200, 500 elements)
- Intersection (100, 200, 500 elements)
- Difference (100, 200, 500 elements)
- Membership (100, 200, 500 elements)
- For-in iteration (100, 200, 500 elements)

**Set literals** (4):
- Small set literal creation
- Large set literals (100, 200, 500 elements)

**Memory allocation** (3):
- Small set allocation
- Large set allocation (100, 500 elements)

## Key Performance Metrics (Baseline)

### High-Performance Operations (< 100 ns/op)
- Tiny program execution: ~93 ns/op
- Set membership (small): ~4.4 ns/op
- Interface method call: ~93 ns/op
- TypeOf: ~63 ns/op
- Enum operations: ~68 ns/op
- Set inclusion: ~77 ns/op

### Medium Performance (100 ns - 1 µs)
- Array access: ~244 ns/op
- Type checks: ~366 ns/op
- For loop (1000 iterations): ~398 ns/op
- Property write (field): ~496 ns/op
- While/Repeat loops: ~650 ns/op

### Low Performance (1 µs - 10 µs)
- Exception raise: ~12 µs/op
- Nested loops: ~1 ms/op
- Arithmetic (1000 ops): ~1.5 ms/op
- Property read (field): ~2 ms/op
- Property write (method): ~1.9 ms/op
- Variant operations: ~2.7 ms/op
- Set operations (large): 2-16 µs/op

### Very Low Performance (> 10 µs)
- Fibonacci(20): ~90 ms/op (756K allocs!)
- Function calls (1000): ~12 ms/op
- Built-in functions: ~9-11 ms/op (due to full pipeline overhead)
- Inherited method call: ~7 ms/op
- Lambda call: ~2 ms/op
- Function pointer call: ~3.3 ms/op

## Memory Allocation Highlights

### Zero Allocations
- Set membership testing
- Set for-in iteration
- Small set literal creation
- Set allocation (bitmask-based)

### Low Allocations (< 100 allocs/op)
- Exception raise: 108 allocs/op
- Interface method call: 833 allocs/op
- Tiny program: 861 allocs/op
- Enum operations: 692 allocs/op

### Medium Allocations (100-10K allocs/op)
- Array operations: 2,179 allocs/op
- For loop: 2,890 allocs/op
- Property operations: 4-8K allocs/op
- Classes: 10,111 allocs/op

### High Allocations (> 10K allocs/op)
- Fibonacci(20): 756,202 allocs/op
- Function calls: 123,107 allocs/op
- Inherited method calls: 61,299 allocs/op

## Areas for Optimization (Post-Refactoring)

### High Priority
1. **Function call overhead** - 123K allocs for 1000 calls
2. **Fibonacci recursion** - 756K allocs, 16.8 MB
3. **Built-in functions** - Running full pipeline for each call

### Medium Priority
4. **Property method calls** - 21-27K allocs per 1000 calls
5. **Lambda calls** - 21K allocs per 1000 calls
6. **Variant operations** - 23-30K allocs per 1000 ops

### Low Priority (Already Optimized)
7. Set operations (bitmask is very fast)
8. Simple arithmetic and loops
9. Interface calls

## Refactoring Impact Targets

Based on design doc (docs/architecture/interpreter-refactoring.md):

### Expected Improvements
- **Value pooling** (Phase 3.2.3): 10-20% allocation reduction
- **Execution context** (Phase 3.3): Reduce state overhead
- **Evaluator pattern** (Phase 3.7): Eliminate switch overhead

### Performance Gates
- No regression > 5% on any benchmark
- Maintain zero-allocation operations (sets, etc.)
- Reduce allocations by 10%+ on high-allocation operations

## Usage

To compare against baseline after refactoring:

```bash
# Run current benchmarks
go test -bench=. -benchmem ./internal/interp -run=^$ > current.txt

# Compare with baseline
benchstat docs/architecture/benchmark_baseline_phase3.txt current.txt
```

## Benchmark Files

- `interpreter_bench_test.go` - High-level execution benchmarks (21)
- `interpreter_operations_bench_test.go` - Operation-specific benchmarks (37)
- `set_bench_test.go` - Set operation benchmarks (30)

Total: **88 benchmarks** covering all major interpreter operations.
