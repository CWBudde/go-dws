# Package ident

Utilities for case-insensitive identifier handling in DWScript.

## Overview

DWScript is a case-insensitive language where `MyVariable`, `myvariable`, and `MYVARIABLE` all refer to the same identifier. This package provides centralized utilities to ensure consistent case-insensitive handling throughout the codebase.

## Installation

```go
import "github.com/cwbudde/go-dws/pkg/ident"
```

## Quick Start

```go
// Normalize for map keys
variables := make(map[string]Value)
variables[ident.Normalize("MyVar")] = value

// Equal for comparisons
if ident.Equal(funcName, "PrintLn") {
    // Handle builtin
}

// Compare for sorting
sort.Slice(names, func(i, j int) bool {
    return ident.Compare(names[i], names[j]) < 0
})
```

## API Reference

### Core Functions

#### `Normalize(s string) string`

Returns the canonical lowercase form of an identifier.

**When to use:**
- Creating map keys for identifier lookups
- Storing normalized identifiers in data structures
- Implementing symbol tables, registries, or caches

**Example:**
```go
env := make(map[string]Value)
env[ident.Normalize("MyVariable")] = intValue(42)
```

**Performance:** Allocates a new string if input isn't all lowercase.

---

#### `Equal(a, b string) bool`

Performs case-insensitive comparison between two strings.

**When to use:**
- One-off identifier comparisons
- Checking if an identifier matches a known value
- Semantic analysis and validation

**Example:**
```go
if ident.Equal(methodName, "Create") {
    // Handle constructor
}
```

**Performance:** Zero allocations, ~5x faster than `ToLower() + ==`.

---

#### `Compare(a, b string) int`

Case-insensitive lexicographic comparison for sorting.

**When to use:**
- Sorting identifiers
- Implementing ordered collections
- Creating deterministic output

**Example:**
```go
sort.Slice(names, func(i, j int) bool {
    return ident.Compare(names[i], names[j]) < 0
})
```

**Returns:**
- Negative if a < b
- Zero if a == b
- Positive if a > b

---

### Helper Functions

#### `Contains(slice []string, s string) bool`

Checks if a slice contains the string (case-insensitive).

```go
keywords := []string{"begin", "end", "if"}
if ident.Contains(keywords, "BEGIN") {
    // Found!
}
```

---

#### `Index(slice []string, s string) int`

Returns the first index of s in slice, or -1 if not found.

```go
tokens := []string{"begin", "var", "x", "end"}
idx := ident.Index(tokens, "VAR") // Returns 1
```

---

#### `IsKeyword(s string, keywords ...string) bool`

Checks if s matches any of the provided keywords.

```go
if ident.IsKeyword(name, "if", "while", "for") {
    // Handle control flow keyword
}
```

---

## Usage Patterns

### Pattern 1: Symbol Table

```go
type Environment struct {
    store map[string]Value  // Keys are normalized
    names map[string]string // normalized -> original case
}

func (e *Environment) Define(name string, val Value) {
    normalized := ident.Normalize(name)
    e.store[normalized] = val
    e.names[normalized] = name  // Preserve original for errors
}

func (e *Environment) Get(name string) (Value, bool) {
    val, ok := e.store[ident.Normalize(name)]
    return val, ok
}
```

### Pattern 2: Function Registry

```go
type FunctionRegistry struct {
    functions map[string]*ast.FunctionDecl
}

func (r *FunctionRegistry) Register(name string, fn *ast.FunctionDecl) {
    r.functions[ident.Normalize(name)] = fn
}

func (r *FunctionRegistry) Lookup(name string) *ast.FunctionDecl {
    return r.functions[ident.Normalize(name)]
}
```

### Pattern 3: Semantic Analysis

```go
// ✅ Good: Use Equal for one-off checks
if ident.Equal(methodName, "Create") {
    // Handle constructor
}

// ❌ Avoid: Wasteful allocation
if strings.ToLower(methodName) == "create" {
    // Creates temporary string
}
```

## Performance

Benchmarks on Intel Xeon @ 2.60GHz:

```
BenchmarkNormalize-16         67.60 ns/op    17 B/op    0 allocs/op
BenchmarkEqual-16             10.12 ns/op     0 B/op    0 allocs/op
BenchmarkCompare-16           58.18 ns/op    10 B/op    1 allocs/op
BenchmarkContains-16          43.08 ns/op     0 B/op    0 allocs/op
```

**Key insights:**
- `Equal()` is ~5x faster than `Normalize() + ==` (zero allocations)
- `Normalize()` allocates only when input isn't lowercase
- `Compare()` normalizes both strings (use for sorting only)

## Migration Guide

Replace direct string operations:

```go
// Before
store[strings.ToLower(name)] = value
if strings.ToLower(a) == strings.ToLower(b) { ... }
if strings.EqualFold(name, "Create") { ... }

// After
store[ident.Normalize(name)] = value
if ident.Equal(a, b) { ... }
if ident.Equal(name, "Create") { ... }
```

## Best Practices

1. **Normalize once, store normalized**: Use `Normalize()` when creating map keys
2. **Compare efficiently**: Use `Equal()` for one-off comparisons
3. **Preserve original case**: Always keep original casing for error messages
4. **Consistent usage**: Use these helpers instead of direct `strings.ToLower()` or `strings.EqualFold()`

## Future Enhancements

This package provides a foundation for potential future improvements:

- Unicode-aware folding (`golang.org/x/text/cases`)
- Identifier interning for reduced memory usage
- Full `Identifier` type with normalization as type invariant
- Locale-aware comparison for international identifiers

## See Also

- [Package Documentation](https://pkg.go.dev/github.com/cwbudde/go-dws/pkg/ident)
- [Examples](example_test.go)
- [DWScript Language Spec](../../CLAUDE.md)
