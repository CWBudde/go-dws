# Benchmarking and Profiling Guide

This guide covers performance benchmarking and profiling for the go-dws project.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Running Benchmarks](#running-benchmarks)
- [CPU Profiling](#cpu-profiling)
- [Memory Profiling](#memory-profiling)
- [Analyzing Results](#analyzing-results)
- [Optimization Workflow](#optimization-workflow)
- [Benchmark Details](#benchmark-details)
- [Common Bottlenecks](#common-bottlenecks)
- [Best Practices](#best-practices)

## Overview

The go-dws project includes comprehensive benchmarks for all major components:

- **Lexer** (`internal/lexer`) - Tokenization performance
- **Parser** (`internal/parser`) - AST construction performance
- **Interpreter** (`internal/interp`) - Execution performance

Each component has dedicated benchmark suites that cover common use cases and edge cases.

## Quick Start

### Run All Benchmarks

```bash
# Use the provided script (recommended)
./scripts/benchmark.sh

# Or run directly with go test
go test -bench=. -benchmem ./...
```

### Generate CPU Profile

```bash
# Profile a specific package
./scripts/profile-cpu.sh -p parser -w

# Profile all packages
./scripts/profile-cpu.sh
```

### Generate Memory Profile

```bash
# Profile a specific package
./scripts/profile-mem.sh -p interp -w

# Profile all packages
./scripts/profile-mem.sh
```

## Running Benchmarks

### Basic Usage

```bash
# Run all benchmarks
go test -bench=. ./...

# Run benchmarks with memory statistics
go test -bench=. -benchmem ./...

# Run benchmarks in a specific package
go test -bench=. -benchmem ./internal/parser

# Run a specific benchmark
go test -bench=BenchmarkLexer ./internal/lexer

# Run benchmarks matching a pattern
go test -bench=Parser ./...
```

### Benchmark Options

```bash
# Run each benchmark for 10 seconds
go test -bench=. -benchtime=10s ./...

# Run each benchmark 10 times
go test -bench=. -count=10 ./...

# Save results to a file
go test -bench=. -benchmem ./... > baseline.txt

# Run benchmarks with verbose output
go test -bench=. -benchmem -v ./...
```

### Using Helper Scripts

The `scripts/` directory contains helper scripts for common benchmarking tasks:

```bash
# Run all benchmarks with defaults (3s, 5 iterations)
./scripts/benchmark.sh

# Custom settings
./scripts/benchmark.sh -t 10s -n 10 -o results.txt

# Verbose output
./scripts/benchmark.sh -v
```

## CPU Profiling

CPU profiling helps identify which functions consume the most CPU time.

### Generate CPU Profile

```bash
# Using the helper script (recommended)
./scripts/profile-cpu.sh -p parser

# Or manually
go test -bench=BenchmarkParser -benchtime=30s \
  -cpuprofile=cpu.prof ./internal/parser
```

### Analyze CPU Profile

#### Interactive Mode

```bash
go tool pprof cpu.prof
```

Commands in interactive mode:
- `top` - Show functions using most CPU
- `top -cum` - Show cumulative CPU usage
- `list <function>` - Show annotated source code
- `web` - Generate call graph (requires graphviz)
- `peek <function>` - Show callers and callees
- `help` - Show all commands

#### Terminal View

```bash
# Show top CPU consumers
go tool pprof -top cpu.prof

# Show top 20 functions
go tool pprof -top -nodecount=20 cpu.prof

# Text format with details
go tool pprof -text cpu.prof
```

#### Web UI (Recommended)

```bash
# Start web server on port 8080
go tool pprof -http=:8080 cpu.prof

# Or use the script with -w flag
./scripts/profile-cpu.sh -p parser -w
```

The web UI provides:
- Interactive flame graphs
- Call graphs
- Source code view
- Function details
- Multiple views (top, graph, flame, peek, source, disasm)

### Example CPU Analysis

```bash
$ go tool pprof -top profiles/parser_cpu.prof

Showing nodes accounting for 450ms, 90% of 500ms total
      flat  flat%   sum%        cum   cum%
     150ms 30.00% 30.00%      200ms 40.00%  parser.parseExpression
     100ms 20.00% 50.00%      150ms 30.00%  parser.parsePrimaryExpression
      80ms 16.00% 66.00%       80ms 16.00%  lexer.NextToken
      70ms 14.00% 80.00%       70ms 14.00%  parser.precedence
      50ms 10.00% 90.00%       50ms 10.00%  runtime.mallocgc
```

This shows:
- `parseExpression` uses 30% of CPU time (150ms)
- `parsePrimaryExpression` uses 20% of CPU time (100ms)
- These are the primary targets for optimization

## Memory Profiling

Memory profiling helps identify allocation hotspots and memory leaks.

### Generate Memory Profile

```bash
# Using the helper script (recommended)
./scripts/profile-mem.sh -p interp

# Or manually
go test -bench=BenchmarkInterpreter -benchtime=30s \
  -memprofile=mem.prof ./internal/interp
```

### Analyze Memory Profile

Memory profiles can show different metrics:
- `alloc_objects` - Number of objects allocated
- `alloc_space` - Total bytes allocated
- `inuse_objects` - Objects currently in use
- `inuse_space` - Bytes currently in use

#### View Allocation Count

```bash
go tool pprof -alloc_objects mem.prof
```

#### View Allocation Size

```bash
go tool pprof -alloc_space mem.prof
```

#### View In-Use Memory

```bash
go tool pprof -inuse_space mem.prof
```

#### Web UI

```bash
go tool pprof -http=:8080 mem.prof

# Or use the script
./scripts/profile-mem.sh -p interp -w
```

### Example Memory Analysis

```bash
$ go tool pprof -top -alloc_space profiles/interp_mem.prof

Showing nodes accounting for 45MB, 90% of 50MB total
      flat  flat%   sum%        cum   cum%
     15MB 30.00% 30.00%      20MB 40.00%  interp.evalExpression
     10MB 20.00% 50.00%      15MB 30.00%  interp.NewEnvironment
      8MB 16.00% 66.00%       8MB 16.00%  ast.NewIntegerLiteral
      7MB 14.00% 80.00%       7MB 14.00%  parser.New
      5MB 10.00% 90.00%       5MB 10.00%  strings.Builder.Grow
```

This shows:
- `evalExpression` allocates 15MB (30% of total)
- `NewEnvironment` allocates 10MB (20% of total)
- These are candidates for optimization (e.g., object pooling)

## Analyzing Results

### Comparing Benchmarks

Use `benchstat` to compare benchmark results:

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run baseline benchmarks
./scripts/benchmark.sh -o baseline.txt

# Make optimizations...

# Run new benchmarks
./scripts/benchmark.sh -o optimized.txt

# Compare results
benchstat baseline.txt optimized.txt
```

Example output:
```
name                    old time/op    new time/op    delta
Parser-8                  1.50ms ± 2%    1.20ms ± 3%  -20.00%  (p=0.000 n=10+10)
Interpreter-8             5.00ms ± 3%    4.00ms ± 2%  -20.00%  (p=0.000 n=10+10)

name                    old alloc/op   new alloc/op   delta
Parser-8                   500kB ± 0%     400kB ± 0%  -20.00%  (p=0.000 n=10+10)
Interpreter-8             1.50MB ± 0%    1.20MB ± 0%  -20.00%  (p=0.000 n=10+10)

name                    old allocs/op  new allocs/op  delta
Parser-8                   5.00k ± 0%     4.00k ± 0%  -20.00%  (p=0.000 n=10+10)
Interpreter-8              15.0k ± 0%     12.0k ± 0%  -20.00%  (p=0.000 n=10+10)
```

### Understanding Benchmark Output

```
BenchmarkParser-8     1000    1500000 ns/op    524288 B/op    5000 allocs/op
```

Breaking this down:
- `BenchmarkParser-8` - Benchmark name with GOMAXPROCS=8
- `1000` - Number of iterations run
- `1500000 ns/op` - Nanoseconds per operation
- `524288 B/op` - Bytes allocated per operation
- `5000 allocs/op` - Number of allocations per operation

## Optimization Workflow

### 1. Establish Baseline

```bash
# Run benchmarks and save results
./scripts/benchmark.sh -o baseline.txt

# Generate profiles
./scripts/profile-cpu.sh
./scripts/profile-mem.sh
```

### 2. Identify Bottlenecks

Look for:
- **CPU hotspots** - Functions with high flat% or cum%
- **Allocation hotspots** - Functions allocating many objects or bytes
- **Slow benchmarks** - Benchmarks with high ns/op
- **Memory hogs** - Benchmarks with high B/op or allocs/op

### 3. Hypothesize Optimizations

Common optimization strategies:
- **Reduce allocations** - Use object pooling, preallocate slices
- **Cache results** - Memoize expensive computations
- **Optimize algorithms** - Use better data structures or algorithms
- **Reduce copying** - Pass pointers instead of values for large structs
- **Avoid string concatenation** - Use strings.Builder
- **Optimize hot paths** - Focus on functions called most frequently

### 4. Implement and Measure

```bash
# Make one optimization at a time
# ...edit code...

# Verify correctness
go test ./...

# Run benchmarks
./scripts/benchmark.sh -o optimized.txt

# Compare results
benchstat baseline.txt optimized.txt
```

### 5. Profile Again

```bash
# Generate new profiles
./scripts/profile-cpu.sh -p <package>
./scripts/profile-mem.sh -p <package>

# Verify the bottleneck is addressed
```

### 6. Iterate

Repeat the process, focusing on the next bottleneck.

## Benchmark Details

### Lexer Benchmarks

Location: `internal/lexer/lexer_test.go`

- `BenchmarkLexer` - Overall lexer performance with realistic code
- `BenchmarkLexerKeywords` - Keyword recognition
- `BenchmarkLexerNumbers` - Number literal parsing (hex, binary, float)
- `BenchmarkLexerStrings` - String literal parsing with escapes

### Parser Benchmarks

Location: `internal/parser/parser_bench_test.go`

- `BenchmarkParser` - Overall parser performance
- `BenchmarkParserExpressions` - Expression parsing
- `BenchmarkParserFunctions` - Function declaration parsing
- `BenchmarkParserClasses` - Class declaration parsing
- `BenchmarkParserControlFlow` - Control flow statement parsing
- `BenchmarkParserTypes` - Type declaration parsing
- `BenchmarkParserArrays` - Array operation parsing
- `BenchmarkParserStrings` - String operation parsing
- `BenchmarkParserSmallProgram` - Small program parsing
- `BenchmarkParserLargeProgram` - Large program parsing

### Interpreter Benchmarks

Location: `internal/interp/interpreter_bench_test.go`

- `BenchmarkInterpreter` - Overall interpreter performance
- `BenchmarkInterpreterFibonacci` - Recursive function calls (Fib10, Fib15, Fib20)
- `BenchmarkInterpreterPrimes` - Loops and conditionals
- `BenchmarkInterpreterLoops` - Different loop types (for, while, repeat)
- `BenchmarkInterpreterArithmetic` - Arithmetic operations
- `BenchmarkInterpreterStrings` - String operations
- `BenchmarkInterpreterArrays` - Array operations
- `BenchmarkInterpreterClasses` - OOP operations
- `BenchmarkInterpreterRecords` - Record operations
- `BenchmarkInterpreterConditionals` - Conditional statements
- `BenchmarkInterpreterFunctionCalls` - Function call overhead
- `BenchmarkInterpreterVariableAccess` - Variable access patterns
- `BenchmarkInterpreterFullPipeline` - Complete pipeline (lex+parse+interpret)

## Common Bottlenecks

### Lexer

1. **String allocations** - Consider using byte slices instead of strings
2. **Character-by-character reading** - Could buffer and process in chunks
3. **Keyword lookup** - Map lookups are fast but could use perfect hashing

### Parser

1. **AST node allocations** - Consider object pooling for frequently created nodes
2. **Precedence table lookups** - Could optimize with switch statements
3. **Token fetching** - Minimize calls to lexer.NextToken()
4. **Error accumulation** - Growing error slices can cause allocations

### Interpreter

1. **Environment lookups** - Symbol table access is in the hot path
2. **Value boxing** - Converting between Go types and interpreter values
3. **Function call overhead** - Each call creates new environment
4. **Type assertions** - Switch statements are often faster than type assertions
5. **String concatenation** - Use strings.Builder for multiple concatenations

## Best Practices

### Running Benchmarks

1. **Close other applications** - Reduce system noise
2. **Use consistent hardware** - Don't compare across different machines
3. **Run multiple iterations** - Use `-count=10` or higher
4. **Use longer benchmark times** - `-benchtime=10s` for more stable results
5. **Warm up** - First run may be slower due to caching effects

### Profiling

1. **Profile for longer periods** - 30s or 1m gives more accurate profiles
2. **Focus on real workloads** - Use representative test programs
3. **Profile one thing at a time** - CPU or memory, not both simultaneously
4. **Look at cumulative time** - Not just flat time
5. **Use web UI** - It's easier to navigate and understand

### Optimizing

1. **Measure first** - Don't guess at bottlenecks
2. **One change at a time** - So you know what worked
3. **Verify correctness** - Run tests after each change
4. **Consider readability** - Don't sacrifice clarity for minor gains
5. **Focus on hot paths** - Optimize code that runs most frequently
6. **Profile after optimizing** - Verify the bottleneck is actually fixed

### Comparing Results

1. **Use benchstat** - It provides statistical analysis
2. **Check p-values** - Ensure differences are significant
3. **Look at all metrics** - Time, allocations, and allocation count
4. **Consider tradeoffs** - Sometimes speed costs memory
5. **Document findings** - Keep notes on what worked and what didn't

## Example Workflow

Here's a complete example of identifying and fixing a performance issue:

### 1. Baseline

```bash
$ ./scripts/benchmark.sh -o baseline.txt
BenchmarkParser-8    1000    1500000 ns/op    524288 B/op    5000 allocs/op
```

### 2. Profile

```bash
$ ./scripts/profile-cpu.sh -p parser
Top CPU consumers:
  30%  parseExpression
  20%  parsePrimaryExpression
  15%  NextToken
```

```bash
$ ./scripts/profile-mem.sh -p parser
Top memory allocations:
  40%  ast.NewIntegerLiteral
  25%  ast.NewBinaryExpression
  15%  parser.New
```

### 3. Identify Issue

The profiler shows that `ast.NewIntegerLiteral` is allocating 40% of memory. Looking at the code, we see it's creating a new node for every integer, even for common values like 0, 1, -1.

### 4. Optimize

Implement a cache for common integer literals:

```go
var commonInts = map[int64]*ast.IntegerLiteral{
    0:  {Value: 0},
    1:  {Value: 1},
    -1: {Value: -1},
    // ... more common values
}

func newIntegerLiteral(value int64) *ast.IntegerLiteral {
    if cached, ok := commonInts[value]; ok {
        return cached
    }
    return &ast.IntegerLiteral{Value: value}
}
```

### 5. Measure

```bash
$ ./scripts/benchmark.sh -o optimized.txt
$ benchstat baseline.txt optimized.txt

name         old time/op    new time/op    delta
Parser-8       1.50ms ± 2%    1.35ms ± 3%  -10.00%  (p=0.000 n=10+10)

name         old alloc/op   new alloc/op   delta
Parser-8        524kB ± 0%     400kB ± 0%  -23.66%  (p=0.000 n=10+10)
```

Success! We've reduced parsing time by 10% and memory allocations by 24%.

### 6. Verify

```bash
$ go test ./internal/parser
PASS
ok      github.com/cwbudde/go-dws/internal/parser    2.345s
```

All tests pass. The optimization is sound.

## Further Reading

- [Go Diagnostics](https://go.dev/doc/diagnostics)
- [pprof Documentation](https://github.com/google/pprof/blob/main/doc/README.md)
- [Go Performance Tips](https://go.dev/wiki/Performance)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [benchstat Documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)

---

Last updated: 2025-10-27
