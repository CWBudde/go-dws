# Benchmarking and Profiling Scripts

This directory contains helper scripts for benchmarking and profiling the go-dws project.

## Scripts Overview

### benchmark.sh

Runs all benchmarks in the project with configurable settings.

**Usage:**
```bash
# Run with defaults (3s, 5 iterations)
./scripts/benchmark.sh

# Custom benchmark time and iterations
./scripts/benchmark.sh -t 10s -n 10

# Save results to file
./scripts/benchmark.sh -o benchmark-results.txt

# Verbose output
./scripts/benchmark.sh -v
```

**Options:**
- `-t, --benchtime TIME` - Benchmark duration (default: 3s)
- `-n, --count N` - Run each benchmark N times (default: 5)
- `-o, --output FILE` - Save results to FILE
- `-v, --verbose` - Verbose output
- `-h, --help` - Show help message

### profile-cpu.sh

Generates CPU profiles for analyzing performance bottlenecks.

**Usage:**
```bash
# Profile all packages
./scripts/profile-cpu.sh

# Profile specific package
./scripts/profile-cpu.sh -p lexer

# Profile and open in web browser
./scripts/profile-cpu.sh -p parser -w

# Custom benchmark time
./scripts/profile-cpu.sh -t 30s
```

**Options:**
- `-t, --benchtime TIME` - Benchmark duration (default: 10s)
- `-o, --output-dir DIR` - Output directory for profiles (default: profiles)
- `-w, --web` - Open profile in web browser
- `-p, --package PKG` - Profile specific package (lexer, parser, interp)
- `-h, --help` - Show help message

**Analyzing CPU profiles:**
```bash
# View top CPU consumers in terminal
go tool pprof -top profiles/lexer_cpu.prof

# Interactive mode
go tool pprof profiles/parser_cpu.prof
# Then use commands like: top, list <function>, web

# Web UI (recommended)
go tool pprof -http=:8080 profiles/interpreter_cpu.prof
```

### profile-mem.sh

Generates memory profiles for analyzing memory usage and allocations.

**Usage:**
```bash
# Profile all packages
./scripts/profile-mem.sh

# Profile specific package
./scripts/profile-mem.sh -p interp

# Profile and open in web browser
./scripts/profile-mem.sh -p parser -w

# Custom benchmark time
./scripts/profile-mem.sh -t 30s
```

**Options:**
- `-t, --benchtime TIME` - Benchmark duration (default: 10s)
- `-o, --output-dir DIR` - Output directory for profiles (default: profiles)
- `-w, --web` - Open profile in web browser
- `-p, --package PKG` - Profile specific package (lexer, parser, interp)
- `-h, --help` - Show help message

**Analyzing memory profiles:**
```bash
# View allocation count
go tool pprof -alloc_objects profiles/parser_mem.prof

# View allocation size
go tool pprof -alloc_space profiles/parser_mem.prof

# View currently in-use memory
go tool pprof -inuse_space profiles/parser_mem.prof

# Web UI (recommended)
go tool pprof -http=:8080 profiles/interpreter_mem.prof
```

## Typical Workflow

### 1. Run Benchmarks

First, establish baseline performance:

```bash
# Run all benchmarks and save results
./scripts/benchmark.sh -o baseline.txt
```

### 2. Profile for Bottlenecks

Identify performance bottlenecks:

```bash
# Generate CPU profile for the interpreter
./scripts/profile-cpu.sh -p interp -w

# Generate memory profile for the parser
./scripts/profile-mem.sh -p parser -w
```

### 3. Analyze and Optimize

In the pprof web UI:
- Look for functions consuming most CPU time
- Check for excessive allocations
- Examine call graphs to understand flow

### 4. Compare Results

After making optimizations:

```bash
# Run benchmarks again
./scripts/benchmark.sh -o optimized.txt

# Compare with baseline using benchstat
go install golang.org/x/perf/cmd/benchstat@latest
benchstat baseline.txt optimized.txt
```

## Common pprof Commands

When in interactive pprof mode:

```
top         - Show functions using most resources
top -cum    - Show cumulative usage
list <func> - Show annotated source for function
web         - Generate call graph (requires graphviz)
peek <func> - Show callers and callees
help        - Show all commands
```

## Tips

1. **Run benchmarks multiple times** - Use `-n 10` or higher for statistical significance
2. **Use longer benchmark times for profiling** - `30s` or `1m` gives more accurate profiles
3. **Profile in isolation** - Close other applications to reduce noise
4. **Focus on hot paths** - Optimize the top 3-5 functions that consume most resources
5. **Measure before and after** - Always benchmark before optimizing to establish baseline

## See Also

- [Benchmarking Guide](../docs/benchmarking.md) - Comprehensive guide to benchmarking
- [Go Diagnostics](https://go.dev/doc/diagnostics) - Official Go profiling documentation
- [pprof Documentation](https://github.com/google/pprof/blob/main/doc/README.md)
