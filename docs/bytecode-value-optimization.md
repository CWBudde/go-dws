# Bytecode Value Optimization Design

**Task**: 9.1.1 - Design Union Type Layout
**Date**: 2025-11-17
**Status**: Design Phase
**Author**: Architecture Team

---

## Executive Summary

This document describes the optimization of the bytecode VM `Value` type from an `interface{}`-based design to an efficient union type. The current implementation incurs significant overhead due to:

1. **Type assertions on every operation** (329 call sites across VM)
2. **Heap allocations for primitive values** (int64, float64, bool)
3. **Poor cache locality** (interface indirection)

**Expected Impact**: 5-10% performance improvement in bytecode VM execution

**Implementation Complexity**: Medium (affects 17 files, ~1000 LOC changes)

---

## Table of Contents

1. [Current Implementation Analysis](#current-implementation-analysis)
2. [Performance Bottlenecks](#performance-bottlenecks)
3. [Proposed Union Design](#proposed-union-design)
4. [Memory Layout Analysis](#memory-layout-analysis)
5. [String Storage Strategy](#string-storage-strategy)
6. [Memory Usage Comparison](#memory-usage-comparison)
7. [Performance Predictions](#performance-predictions)
8. [Implementation Considerations](#implementation-considerations)
9. [Go Compiler Optimizations](#go-compiler-optimizations)
10. [Migration Strategy](#migration-strategy)
11. [Success Criteria](#success-criteria)

---

## Current Implementation Analysis

### Struct Definition

```go
// Current implementation (internal/bytecode/bytecode.go:10)
type Value struct {
    Data interface{}  // 16 bytes (pointer + type descriptor)
    Type ValueType    // 1 byte
}
```

### Memory Layout (64-bit architecture)

```
Current Value: 24 bytes total (with padding)
┌─────────────────────────────────────────┐
│ Data (interface{})                      │  16 bytes (ptr + type)
│   - data pointer:      8 bytes          │
│   - type descriptor:   8 bytes          │
├─────────────────────────────────────────┤
│ Type (ValueType)                        │  1 byte
├─────────────────────────────────────────┤
│ Padding                                 │  7 bytes (alignment)
└─────────────────────────────────────────┘
```

### Value Type Distribution

Based on bytecode.go analysis, we support 12 value types:

| Type | Data Content | Frequency | Allocation |
|------|--------------|-----------|------------|
| ValueInt | int64 | Very High | **Heap** ⚠️ |
| ValueFloat | float64 | High | **Heap** ⚠️ |
| ValueBool | bool | High | **Heap** ⚠️ |
| ValueString | string | Medium | **Heap** ⚠️ |
| ValueNil | nil | Medium | None |
| ValueArray | *ArrayInstance | Medium | Stack (ptr only) |
| ValueObject | *ObjectInstance | Medium | Stack (ptr only) |
| ValueRecord | *RecordInstance | Low | Stack (ptr only) |
| ValueFunction | *FunctionObject | Low | Stack (ptr only) |
| ValueClosure | *Closure | Low | Stack (ptr only) |
| ValueBuiltin | string (name) | Low | **Heap** ⚠️ |
| ValueVariant | Value (wrapped) | Low | **Heap** ⚠️ |

**Key Observation**: Primitives (int64, float64, bool) are frequently used but require heap allocation with current design.

### Type Assertion Hotspots

**Total type assertions found**: 329 occurrences across 17 files

**Critical hot paths**:
- `vm_ops.go`: 20 assertions in arithmetic/comparison operations
- `vm_exec.go`: 19 assertions in main execution loop
- `vm_builtins_string.go`: 90 assertions in string operations
- `compiler_core.go`: 37 assertions in constant folding

**Example from VM operations** (`vm_ops.go:20`):
```go
func (vm *VM) binaryIntOp(fn func(a, b int64) int64) error {
    // ... pop values ...
    vm.push(IntValue(fn(left.AsInt(), right.AsInt())))
    //                     ^^^^^^^^^    ^^^^^^^^^
    //                     Type assertion overhead
}

func (v Value) AsInt() int64 {
    if v.Type == ValueInt {
        return v.Data.(int64)  // ← Heap allocation + type assertion
    }
    return 0
}
```

---

## Performance Bottlenecks

### 1. Type Assertion Overhead

Every `AsInt()`, `AsFloat()`, `AsBool()`, `AsString()` call performs:
- Type check (1 comparison)
- Type assertion (1 interface cast)
- Potential panic recovery overhead

**Cost per assertion**: ~2-3 CPU cycles

**Total overhead**: 329 call sites × average execution frequency = significant

### 2. Heap Allocation Overhead

Creating primitive values requires heap allocation:

```go
// Current: IntValue allocates on heap
func IntValue(i int64) Value {
    return Value{Type: ValueInt, Data: i}
    //                                  ↑
    //                         Boxes int64 to interface{}
    //                         → heap allocation
}
```

**Allocation cost**:
- Heap allocation: ~50-100 CPU cycles
- GC pressure: additional overhead during collection
- Cache misses: data not co-located with Value struct

**Frequency**: Every arithmetic operation creates new Value instances

### 3. Cache Locality Issues

```
Current layout with interface{}:
┌────────┐        ┌──────────┐
│ Value  │───────▶│ int64(42)│  (separate heap object)
│ Type=Int│       └──────────┘
│ Data=ptr│
└────────┘

Issues:
- Two cache lines for one value
- Pointer indirection penalty
- Poor spatial locality
```

### 4. Memory Bandwidth

**Current**: 24 bytes per Value + heap object (8-16 bytes) = 32-40 bytes total

**Impact on cache**:
- Cache line: 64 bytes
- Current Value: 1.5-2 cache lines per primitive value
- Reduces effective cache utilization

---

## Proposed Union Design

### Core Concept

Use explicit fields with overlapping memory layout (union semantics) to store different value types in-place, eliminating heap allocations for primitives.

### Struct Definition (Option A: Explicit Union)

```go
// Proposed implementation
type Value struct {
    typ ValueType    // 1 byte - type tag MUST BE FIRST for alignment
    _   [7]byte      // 7 bytes - explicit padding

    // Union fields (only one active based on typ)
    // These fields overlap in memory via Go compiler optimization
    i64 int64        // 8 bytes - for ValueInt
    f64 float64      // 8 bytes - for ValueFloat
    str string       // 16 bytes - for ValueString (header only)
    ptr unsafe.Pointer // 8 bytes - for pointers (*ArrayInstance, etc.)
    b   bool         // 1 byte - for ValueBool
}
```

**Size**: 32 bytes (see memory layout analysis below)

### Alternative Design (Option B: Array Union)

```go
type Value struct {
    typ  ValueType     // 1 byte
    _    [7]byte       // padding
    data [16]byte      // 16-byte union (enough for any type)
}

// Accessor methods use unsafe to cast data based on typ
func (v Value) AsInt() int64 {
    return *(*int64)(unsafe.Pointer(&v.data[0]))
}
```

**Pros**: More compact (24 bytes total), explicit union semantics
**Cons**: More unsafe code, less idiomatic Go

**Decision**: Use **Option A** for better Go idioms and compiler optimization

---

## Memory Layout Analysis

### Proposed Value Layout (Option A)

```
Proposed Value: 32 bytes total
┌─────────────────────────────────────────┐  Offset
│ typ (ValueType)                         │  0-0     (1 byte)
├─────────────────────────────────────────┤
│ padding                                 │  1-7     (7 bytes)
├─────────────────────────────────────────┤
│ i64 (int64)          ┐                  │  8-15    (8 bytes)
│ f64 (float64)        │ Union            │  8-15    (8 bytes)
│ ptr (unsafe.Pointer) │ (overlapping)    │  8-15    (8 bytes)
│ b   (bool)           ┘                  │  8       (1 byte)
├─────────────────────────────────────────┤
│ str (string)                            │  16-31   (16 bytes)
│   - data pointer:     8 bytes           │  16-23
│   - length:           8 bytes           │  24-31
└─────────────────────────────────────────┘
```

**Key Points**:
1. `typ` at offset 0 for optimal alignment
2. Fields i64/f64/ptr occupy same memory (8-15)
3. String occupies separate space (16-31) - cannot overlap due to size
4. Bool can overlap with i64/f64/ptr at offset 8

### Field Overlap Strategy

**Non-overlapping**:
- `typ`: Always at offset 0
- `str`: Always at offset 16-31 (too large to overlap)

**Overlapping** (bytes 8-15):
- `i64`: Used when `typ == ValueInt`
- `f64`: Used when `typ == ValueFloat`
- `ptr`: Used when `typ` is pointer type (Array, Object, Record, Function, Closure)
- `b`: Uses only byte 8 when `typ == ValueBool`

### Go Compiler Behavior

**Important**: Go compiler **will NOT automatically create overlapping fields**. We rely on:

1. **Dead code elimination**: Unused fields optimized away per access path
2. **Escape analysis**: Value likely stays on stack in hot loops
3. **Inlining**: Accessor methods inlined to direct field access

**Verification needed**: Check generated assembly to confirm optimization

**Alternative if needed**: Use `unsafe.Pointer` and manual layout (Option B)

---

## String Storage Strategy

### Analysis

Strings in Go are **not** simple pointers but structs:

```go
type stringStruct struct {
    data uintptr  // 8 bytes - pointer to data
    len  int      // 8 bytes - length
}
```

**Total**: 16 bytes (cannot fit in 8-byte union slot)

### Options Considered

#### Option 1: Separate String Field (CHOSEN)

```go
type Value struct {
    typ ValueType
    _   [7]byte
    i64 int64     // 8 bytes (union slot)
    str string    // 16 bytes (separate)
}
```

**Pros**:
- Simple implementation
- No string copies
- Go semantics preserved

**Cons**:
- Larger struct (32 bytes vs 24 bytes)
- Some memory waste when not storing strings

#### Option 2: String Pointer

```go
type Value struct {
    typ ValueType
    _   [7]byte
    i64 int64     // 8 bytes
    ptr unsafe.Pointer // string stored as *string
}
```

**Pros**:
- Smaller struct (24 bytes)

**Cons**:
- Requires heap allocation for strings (defeats purpose)
- Additional indirection
- More complex lifecycle management

#### Option 3: Small String Optimization (SSO)

```go
type Value struct {
    typ ValueType
    len uint8        // string length (if <= 22)
    data [22]byte    // inline for small strings
    ptr  *string     // pointer for large strings
}
```

**Pros**:
- Optimizes common case (short strings)

**Cons**:
- Complex implementation
- Branching overhead
- Copies for all strings

### Decision: Option 1 (Separate String Field)

**Rationale**:
- Simplicity and maintainability
- String operations are not the bottleneck (integers/floats are)
- 32 bytes is still better than 24 bytes + heap object (40+ bytes total)
- Go string semantics preserved

---

## Memory Usage Comparison

### Per-Value Memory Cost

| Scenario | Current (interface{}) | Proposed (union) | Savings |
|----------|----------------------|------------------|---------|
| Integer (42) | 24 + 16 = **40 bytes** | **32 bytes** | 8 bytes (20%) |
| Float (3.14) | 24 + 16 = **40 bytes** | **32 bytes** | 8 bytes (20%) |
| Bool (true) | 24 + 16 = **40 bytes** | **32 bytes** | 8 bytes (20%) |
| String ("hello") | 24 + 16 = **40 bytes** | **32 bytes** | 8 bytes (20%) |
| Pointer (*Obj) | 24 + 0 = **24 bytes** | **32 bytes** | -8 bytes |
| Nil | **24 bytes** | **32 bytes** | -8 bytes |

**Notes**:
- Current "40 bytes" = 24-byte Value + 16-byte heap allocation for boxed value
- Proposed "32 bytes" = entire Value on stack, no heap allocation
- Pointer types slightly larger (acceptable tradeoff)

### Stack Frame Impact

**Typical VM stack**: 256 slots (configurable)

| Implementation | Memory Usage | GC Pressure |
|----------------|--------------|-------------|
| Current | 256 × 24 = 6,144 bytes | High (+ heap objects) |
| Proposed | 256 × 32 = 8,192 bytes | Low (stack-only) |

**Analysis**:
- Stack usage increases by 2,048 bytes (33%)
- But eliminates potentially thousands of heap allocations
- Net benefit: lower GC pressure, better performance

### Cache Line Utilization

**Cache line size**: 64 bytes

| Implementation | Values per Line | Utilization |
|----------------|-----------------|-------------|
| Current | 2.67 values | 66% (24 bytes) |
| Proposed | 2 values | 100% (32 bytes) |

**Note**: "Current" doesn't account for heap indirection (separate cache lines)

---

## Performance Predictions

### Microbenchmark Predictions

Based on eliminated overhead:

| Operation | Current | Proposed | Speedup |
|-----------|---------|----------|---------|
| IntValue creation | 50-100 cycles | 0 cycles | ∞ (no alloc) |
| AsInt() access | 2-3 cycles | 0 cycles | ∞ (direct) |
| Integer add | ~60 cycles | ~10 cycles | **6x** |
| Float multiply | ~65 cycles | ~12 cycles | **5.4x** |
| String concat | ~80 cycles | ~75 cycles | 1.07x |

**Note**: "0 cycles" means optimized away by compiler inlining

### Real-World Performance

**Conservative estimate**: 5-10% improvement in VM execution

**Basis**:
- Arithmetic operations: 40-50% of VM time (conservative)
- Type assertions: ~20% overhead on arithmetic
- Heap allocations: ~30% overhead on arithmetic
- Total eliminated: ~50% of arithmetic overhead
- **Net**: 50% × 50% = 25% improvement on arithmetic = 12.5% overall
- **Conservative**: 5-10% accounting for other factors

**Optimistic estimate**: 10-15% improvement if arithmetic-heavy workloads

**Pessimistic estimate**: 3-5% if workload is I/O or call-heavy

### Benchmark Targets

Create benchmarks for:
1. **BenchmarkValueCreation**: IntValue/FloatValue creation
2. **BenchmarkValueAccess**: AsInt/AsFloat access
3. **BenchmarkArithmetic**: Add/Sub/Mul/Div operations
4. **BenchmarkComparison**: Equal/Less/Greater operations
5. **BenchmarkRealWorld**: Fibonacci, prime sieve, etc.

**Success criteria**: 5-10% improvement on BenchmarkRealWorld

---

## Implementation Considerations

### Accessor Method Changes

**Current** (with type assertion):
```go
func (v Value) AsInt() int64 {
    if v.Type == ValueInt {
        return v.Data.(int64)  // ← Type assertion
    }
    return 0
}
```

**Proposed** (direct field access):
```go
func (v Value) AsInt() int64 {
    // Type check optional (caller responsibility via IsInt())
    return v.i64  // ← Direct field access
}
```

**Question**: Should we keep type checks?

**Answer**:
- **Debug builds**: Keep checks for safety
- **Release builds**: Consider removing for performance
- **Alternative**: Use build tags or const bool flag

```go
const valueDebugChecks = false // Set to true for development

func (v Value) AsInt() int64 {
    if valueDebugChecks && v.typ != ValueInt {
        panic("AsInt called on non-integer value")
    }
    return v.i64
}
```

### Constructor Method Changes

**Current**:
```go
func IntValue(i int64) Value {
    return Value{Type: ValueInt, Data: i}  // Boxes to interface{}
}
```

**Proposed**:
```go
func IntValue(i int64) Value {
    return Value{typ: ValueInt, i64: i}  // Direct field assignment
}
```

**Optimization**: Inline candidates (small functions)

### Variant Handling

**Current**:
```go
func VariantValue(wrapped Value) Value {
    return Value{Type: ValueVariant, Data: wrapped}  // Boxes Value
}

func (v Value) AsVariant() Value {
    if v.Type == ValueVariant {
        if wrapped, ok := v.Data.(Value); ok {
            return wrapped
        }
    }
    return NilValue()
}
```

**Proposed** (requires special handling):

Option 1: Store pointer to heap-allocated Value
```go
func VariantValue(wrapped Value) Value {
    // Allocate on heap (rare case, acceptable)
    p := new(Value)
    *p = wrapped
    return Value{typ: ValueVariant, ptr: unsafe.Pointer(p)}
}

func (v Value) AsVariant() Value {
    if v.typ == ValueVariant {
        return *(*Value)(v.ptr)
    }
    return NilValue()
}
```

Option 2: Embed variant in separate field (increases size)
```go
type Value struct {
    typ     ValueType
    i64     int64
    str     string
    ptr     unsafe.Pointer
    variant *Value  // Only for ValueVariant (usually nil)
}
```

**Decision**: Option 1 (heap allocation for variants) - variants are rare

### Pointer Type Handling

All pointer types share the `ptr` field:

```go
func (v Value) AsArray() *ArrayInstance {
    if v.typ == ValueArray {
        return (*ArrayInstance)(v.ptr)
    }
    return nil
}

func (v Value) AsObject() *ObjectInstance {
    if v.typ == ValueObject {
        return (*ObjectInstance)(v.ptr)
    }
    return nil
}
```

**Safety**: Type tag ensures correct casting

---

## Go Compiler Optimizations

### Expected Optimizations

1. **Inlining**: Small accessor methods (AsInt, AsFloat, etc.)
2. **Dead code elimination**: Unused union fields removed
3. **Escape analysis**: Value stays on stack in loops
4. **Constant propagation**: Type checks eliminated when static

### Verification Strategy

**Tool**: `go build -gcflags='-m -m'` (escape analysis)

**Expected output**:
```
./bytecode.go:XX:YY: Value.AsInt inlined
./bytecode.go:XX:YY: IntValue inlined
./bytecode.go:XX:YY: v does not escape
```

**Assembly inspection**: `go tool objdump`

**Expected assembly** for `AsInt()`:
```asm
; Current (with type assertion)
CALL    runtime.assertI2I
MOVQ    8(AX), AX    ; Load from heap

; Proposed (direct access)
MOVQ    8(DI), AX    ; Direct field access (offset 8)
```

### Potential Issues

**Issue 1**: Go may not overlap fields even if unused

**Mitigation**: Accept 32-byte size, still better than 40-byte current

**Issue 2**: Escape analysis may fail for some cases

**Mitigation**: Profile and optimize hot paths manually

**Issue 3**: Unsafe pointer usage may inhibit optimizations

**Mitigation**: Limit unsafe usage to specific cases (Variant, pointers)

---

## Migration Strategy

### Phase 1: Parallel Implementation (Task 9.1.2)

- Create new Value struct alongside current
- Implement all constructors and accessors
- Add comprehensive unit tests
- Benchmark against current implementation

### Phase 2: VM Migration (Task 9.1.3)

- Update VM operations to use new Value
- Update arithmetic operations first (highest impact)
- Then comparison, logical, array operations
- Verify correctness with existing tests

### Phase 3: Compiler Migration (Task 9.1.4)

- Update bytecode compiler constant emission
- Update serialization/deserialization
- Verify .dwc file compatibility

### Phase 4: Validation (Task 9.1.5)

- Run comprehensive benchmarks
- Validate 5-10% improvement target
- Profile to find remaining hotspots
- Run full test suite + fixture tests

### Rollback Plan

If performance doesn't improve or tests fail:

1. Keep new Value as `ValueV2`
2. Add feature flag to switch implementations
3. Investigate and fix issues
4. Retry migration

```go
const useOptimizedValue = true  // Feature flag

type Value = ValueV2  // Switch implementation
```

---

## Success Criteria

### Performance Targets

- [ ] **Primary**: 5-10% improvement in real-world benchmarks
- [ ] **Secondary**: 20-50% improvement in arithmetic microbenchmarks
- [ ] **Tertiary**: Reduced GC pressure (fewer allocations)

### Correctness Targets

- [ ] All existing tests pass (>3,000 tests)
- [ ] All fixture tests pass (~2,100 tests)
- [ ] Zero regressions in functionality
- [ ] Bytecode serialization compatible

### Code Quality Targets

- [ ] No unsafe code except where necessary (Variant, pointers)
- [ ] Comprehensive unit tests for new Value
- [ ] Benchmarks for all value types
- [ ] Documentation updated (this doc + bytecode-vm.md)

### Memory Targets

- [ ] No increase in total memory usage
- [ ] Reduced heap allocations (measurable via profiling)
- [ ] Stack size increase acceptable (< 10KB per goroutine)

---

## Benchmarking Plan

### Microbenchmarks

Create `internal/bytecode/value_bench_test.go`:

```go
// BenchmarkValueCreation_Current vs BenchmarkValueCreation_Proposed
func BenchmarkIntValueCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = IntValue(42)
    }
}

// BenchmarkValueAccess_Current vs BenchmarkValueAccess_Proposed
func BenchmarkAsInt(b *testing.B) {
    v := IntValue(42)
    for i := 0; i < b.N; i++ {
        _ = v.AsInt()
    }
}

// BenchmarkArithmetic_Current vs BenchmarkArithmetic_Proposed
func BenchmarkIntAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        a := IntValue(int64(i))
        b := IntValue(42)
        _ = IntValue(a.AsInt() + b.AsInt())
    }
}
```

### Real-World Benchmarks

Use existing DWScript programs:

```go
func BenchmarkFibonacci(b *testing.B) {
    // Compile fibonacci.dws once
    program := compileProgram("testdata/benchmarks/fibonacci.dws")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = program.Run()
    }
}

func BenchmarkPrimeSieve(b *testing.B) {
    program := compileProgram("testdata/benchmarks/prime_sieve.dws")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = program.Run()
    }
}
```

### Profiling Commands

```bash
# CPU profiling
go test -bench=BenchmarkFibonacci -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go test -bench=BenchmarkFibonacci -memprofile=mem.prof
go tool pprof mem.prof

# Escape analysis
go build -gcflags='-m -m' ./internal/bytecode 2>&1 | grep Value

# Assembly inspection
go build -gcflags='-S' ./internal/bytecode 2>&1 | grep AsInt
```

---

## Open Questions

### Q1: Should we use unsafe.Pointer for union?

**Options**:
- A: Use separate fields, rely on compiler (RECOMMENDED)
- B: Use [16]byte + unsafe casts for true union

**Decision**: Start with A, switch to B if size is critical

### Q2: Keep runtime type checks in accessors?

**Options**:
- A: Always check (safe, slower)
- B: Never check (fast, unsafe)
- C: Build tag controlled (flexible)

**Decision**: C (build tag) - safety in dev, speed in production

### Q3: How to handle Variant values?

**Options**:
- A: Heap allocate wrapped Value (rare case)
- B: Add separate variant field (increases size)

**Decision**: A (acceptable for rare case)

### Q4: Should we optimize for small structs (24 bytes)?

**Options**:
- A: Accept 32 bytes for simplicity
- B: Use Option B design ([16]byte union) for 24 bytes

**Decision**: A initially, B if benchmarks show size matters

---

## Next Steps (Task 9.1.2)

1. Implement new Value struct in `internal/bytecode/bytecode.go`
2. Create unit tests in `internal/bytecode/value_test.go`
3. Create benchmarks in `internal/bytecode/value_bench_test.go`
4. Run benchmarks and compare with current implementation
5. Adjust design based on results

---

## Appendix A: Alternative Designs Considered

### Design 1: Interface-based (Status Quo)

```go
type Value struct {
    Data interface{}
    Type ValueType
}
```

**Rejected**: Current performance issues

### Design 2: Type Switch Pattern

```go
type Value interface { isValue() }
type IntValue int64
type FloatValue float64
// etc.
```

**Rejected**: Type switches still have overhead, larger code size

### Design 3: Pointer-based Tagged Union

```go
type Value struct {
    tag ValueType
    ptr unsafe.Pointer  // Points to heap-allocated value
}
```

**Rejected**: Still requires heap allocation (defeats purpose)

### Design 4: Small Value Optimization (SVO)

```go
type Value struct {
    tag  ValueType
    data [23]byte     // Inline for values <= 23 bytes
}
```

**Rejected**: Too complex, string handling issues

---

## Appendix B: Go Memory Layout Reference

### Primitive Type Sizes

| Type | Size | Alignment |
|------|------|-----------|
| bool | 1 byte | 1 byte |
| int8 | 1 byte | 1 byte |
| int16 | 2 bytes | 2 bytes |
| int32 | 4 bytes | 4 bytes |
| int64 | 8 bytes | 8 bytes |
| float32 | 4 bytes | 4 bytes |
| float64 | 8 bytes | 8 bytes |
| string | 16 bytes | 8 bytes |
| unsafe.Pointer | 8 bytes | 8 bytes |
| interface{} | 16 bytes | 8 bytes |

### Struct Padding Rules

Go compiler adds padding to ensure:
1. Each field is aligned to its natural boundary
2. Struct size is multiple of largest field alignment

**Example**:
```go
type Example struct {
    a bool      // offset 0, size 1
    // padding: 7 bytes
    b int64     // offset 8, size 8
}
// Total: 16 bytes
```

### Cache Line Considerations

- **x86-64 cache line**: 64 bytes
- **ARM cache line**: 64 bytes (typically)
- **Goal**: Minimize cache line usage per Value

---

## References

- [Go Memory Model](https://go.dev/ref/mem)
- [Go Compiler Optimizations](https://github.com/golang/go/wiki/CompilerOptimizations)
- [Escape Analysis](https://go.dev/blog/escape-analysis)
- [Performance without the event loop](https://dave.cheney.net/2015/08/08/performance-without-the-event-loop)
- [Understanding Go Data Structures](https://research.swtch.com/godata)

---

**Document Version**: 1.0
**Last Updated**: 2025-11-17
**Review Status**: Pending Implementation
