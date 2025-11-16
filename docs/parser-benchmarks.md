# Parser Benchmarks

This document describes the parser benchmark infrastructure and baseline performance metrics for the go-dws parser.

## Overview

The parser benchmarks are located in `internal/parser/parser_bench_test.go` and provide comprehensive performance testing for all major parsing scenarios. The benchmark suite includes both functional coverage (ensuring all language features are tested) and performance regression detection.

## Benchmark Categories

### 1. Overall Parser Performance

- **BenchmarkParser**: Realistic DWScript program with functions, classes, and control flow
- **BenchmarkParserSmallProgram**: Minimal program (baseline overhead measurement)
- **BenchmarkParserLargeProgram**: Large generated program (scalability testing)

### 2. Expression Parsing

- **BenchmarkParserExpressions**: Various expression types (arithmetic, logical, calls, indexing, member access)
- **BenchmarkParserExpressionLists**: Long expression lists and comma-separated values
- **BenchmarkParserMemberAccessChains**: Deep member access chains (e.g., `a.b.c.d.e.f`)

### 3. Declaration Parsing

- **BenchmarkParserFunctions**: Function and procedure declarations
- **BenchmarkParserClasses**: Class hierarchies with inheritance
- **BenchmarkParserTypes**: Type declarations (enums, records, arrays, function types)
- **BenchmarkParserComplexNestedTypes**: Deeply nested type structures

### 4. Statement Parsing

- **BenchmarkParserControlFlow**: Control flow statements (if/while/for/repeat/case/try)
- **BenchmarkParserArrays**: Array operations and literals
- **BenchmarkParserStrings**: String operations

### 5. Advanced Features

- **BenchmarkParserContracts**: Contract parsing (require/ensure/invariant)
- **BenchmarkParserParameterLists**: Functions with many parameters

### 6. Error Recovery

- **BenchmarkParserErrorRecovery**: Parser performance with syntax errors

### 7. Memory Profiling

- **BenchmarkParserWithAllocations**: Memory allocation tracking (`-benchmem`)

## Running Benchmarks

### Basic Usage

```bash
# Run all parser benchmarks
go test -bench=BenchmarkParser ./internal/parser

# Run with memory profiling
go test -bench=BenchmarkParser -benchmem ./internal/parser

# Run specific benchmark
go test -bench=BenchmarkParserExpressions ./internal/parser

# Run with custom duration
go test -bench=BenchmarkParser -benchtime=5s ./internal/parser

# Save results to file
go test -bench=BenchmarkParser ./internal/parser | tee benchmarks/results.txt
```

### Using the Comparison Script

The `scripts/bench_compare.sh` script automates baseline comparison:

```bash
# Create initial baseline (do this before making changes)
./scripts/bench_compare.sh --save-baseline

# After making changes, compare to baseline
./scripts/bench_compare.sh

# The script will show performance changes:
#   ▲ Red: >10% slower (regression)
#   ▼ Green: >10% faster (improvement)
#   ≈ White: <10% change (neutral)
```

For more detailed statistical analysis, install `benchstat`:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
./scripts/bench_compare.sh  # Will automatically use benchstat if available
```

### Environment Variables

- `BENCHTIME`: Benchmark duration (default: 1s)
  - Examples: `100ms`, `500ms`, `2s`, `5s`
  - Usage: `BENCHTIME=5s ./scripts/bench_compare.sh --save-baseline`

## Baseline Performance

**Note**: These benchmarks were run on the reference system. Your results may vary based on hardware.

### Reference System

- **CPU**: Intel(R) Xeon(R) CPU @ 2.60GHz
- **OS**: Linux
- **Go Version**: 1.21+
- **Date**: 2024 (Phase 2.1.4 completion)

### Benchmark Results

Results are saved in `benchmarks/baseline.txt` after running:

```bash
./scripts/bench_compare.sh --save-baseline
```

### Expected Performance Characteristics

1. **Small programs** (<100 lines): <1ms parsing time
2. **Medium programs** (100-1000 lines): 1-10ms parsing time
3. **Large programs** (>1000 lines): Roughly linear with program size
4. **Memory allocations**: Approximately 60-80 bytes per AST node

### Performance Goals

- **No regressions**: <10% performance degradation from baseline
- **Linear scaling**: Parse time should scale linearly with program size
- **Memory efficiency**: Minimize allocations in hot paths
- **Error recovery**: Error handling should not significantly impact performance

## Interpreting Results

### Benchmark Output Format

```
BenchmarkParserExpressions-16    5000    250000 ns/op    15000 B/op    240 allocs/op
                          ││      ││     ││              ││            ││
                          ││      ││     ││              ││            └─ Allocations per operation
                          ││      ││     ││              └─ Bytes allocated per operation
                          ││      ││     └─ Nanoseconds per operation
                          ││      └─ Number of iterations
                          │└─ GOMAXPROCS value
                          └─ Benchmark name
```

### What to Watch For

1. **Regressions** (▲): If a benchmark shows >10% slower performance:
   - Review recent changes to that code path
   - Check for unnecessary allocations
   - Profile with `go test -cpuprofile` or `-memprofile`

2. **Improvements** (▼): If a benchmark shows >10% faster performance:
   - Verify correctness hasn't been compromised
   - Consider if improvement applies to other areas
   - Document optimization technique

3. **High variance**: If results vary widely between runs:
   - Increase `-benchtime` for more stable results
   - Check for system load or other processes
   - Use `benchstat` for statistical analysis

## Regression Testing

### CI Integration

The benchmark suite can be integrated into CI to automatically detect regressions:

```bash
# In CI pipeline:
# 1. Run benchmarks and save results
go test -bench=BenchmarkParser -benchtime=5s ./internal/parser > current.txt

# 2. Compare to baseline (stored in repo or CI cache)
./scripts/bench_compare.sh benchmarks/baseline.txt current.txt

# 3. Fail if major regressions detected (can parse output)
```

### When to Update Baseline

Update the baseline when:

1. **Intentional changes**: You've made intentional performance optimizations
2. **New features**: New language features may change performance characteristics
3. **Major refactoring**: Significant architectural changes may require new baseline
4. **Go version upgrade**: Different Go versions may have different performance

Do NOT update baseline to hide regressions!

## Profiling

For detailed performance analysis:

```bash
# CPU profiling
go test -bench=BenchmarkParser -cpuprofile=cpu.prof ./internal/parser
go tool pprof cpu.prof

# Memory profiling
go test -bench=BenchmarkParser -memprofile=mem.prof ./internal/parser
go tool pprof mem.prof

# Trace execution
go test -bench=BenchmarkParser -trace=trace.out ./internal/parser
go tool trace trace.out
```

### Common Profiling Insights

1. **Hot paths**: Look for functions consuming >10% of CPU time
2. **Allocation hotspots**: Functions with high allocation counts
3. **String concatenation**: Consider using strings.Builder
4. **Map lookups**: Consider caching frequently accessed values
5. **Recursive calls**: Ensure tail call elimination where possible

## Adding New Benchmarks

When adding new language features, add corresponding benchmarks:

```go
// BenchmarkParserNewFeature benchmarks parsing of new feature
func BenchmarkParserNewFeature(b *testing.B) {
    input := `
    // DWScript code using new feature
    `

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        l := lexer.New(input)
        p := New(l)
        _ = p.ParseProgram()
    }
}
```

### Benchmark Best Practices

1. **Realistic code**: Use representative DWScript programs
2. **Avoid setup overhead**: Use `b.ResetTimer()` after setup
3. **Measure allocations**: Add `-benchmem` variant if memory-critical
4. **Document purpose**: Explain what aspect of parsing is being tested
5. **Keep focused**: Each benchmark should test one specific aspect
6. **Run long enough**: Use appropriate `-benchtime` for stable results

## Benchmark History

### Phase 2.1.4 (Initial Baseline)

- Added 7 new benchmarks to cover missing patterns:
  - Parameter lists (many parameters)
  - Complex nested types
  - Contracts (require/ensure)
  - Member access chains
  - Error recovery
  - Memory allocations
  - Expression lists

- Total benchmarks: 17
- Coverage: All major parsing scenarios
- Infrastructure: Comparison script for regression detection

### Future Improvements

Potential areas for benchmark expansion:

1. **Incremental parsing**: Benchmark partial re-parsing
2. **Parallel parsing**: Benchmark concurrent parsing of multiple files
3. **Large file handling**: Benchmark parsing of very large files (>10k lines)
4. **Error-heavy code**: Benchmark parsing code with many errors
5. **Comment handling**: Benchmark parsing with heavy commenting

## References

- [Go Benchmarking Guide](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [benchstat Documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Profiling Go Programs](https://go.dev/blog/pprof)
