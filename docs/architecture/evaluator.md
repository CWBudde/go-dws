# Evaluator Architecture - Consolidated Summary

**Created**: 2025-12-04
**Status**: Current as of Phase 3.5 completion
**Purpose**: Single authoritative reference for evaluator architecture and migration status

---

## Executive Summary

The DWScript interpreter has been refactored from a monolithic 3000+ line switch statement to a clean visitor pattern architecture. This document consolidates the latest information about the evaluator's design, performance, migration status, and architectural boundaries.

### Current Status (Phase 3.5 Complete)

✅ **Visitor pattern migration**: 48+ visitor methods implemented
✅ **Performance validation**: Zero regression, 0.3-119 ns/op depending on operation
✅ **Reference counting**: Migrated to runtime.RefCountManager (7 adapter calls eliminated)
✅ **Adapter reduction**: 95%+ reduction (75 → ~5 essential methods)
✅ **Type system**: All value types migrated to runtime package

---

## Architecture

### Component Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                          Interpreter                            │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ OOP Semantics Layer                                        │ │
│  │ - ClassInfo registry                                       │ │
│  │ - ObjectInstance lifecycle                                 │ │
│  │ - Self context management                                  │ │
│  │ - Method dispatch & virtual method tables                  │ │
│  └────────────────────────────────────────────────────────────┘ │
│                              ▲                                  │
│                              │ Adapter Interface (2 essential)  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ Evaluator                                                  │ │
│  │ - AST node evaluation (visitor pattern)                    │ │
│  │ - Expression/statement execution                           │ │
│  │ - Environment-based variable resolution                    │ │
│  │ - Control flow management                                  │ │
│  └────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ Runtime Package                                            │ │
│  │ - Value types (Integer, Float, String, etc.)               │ │
│  │ - RefCountManager (lifecycle management)                   │ │
│  │ - Environment (variable storage)                           │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Separation of Concerns

**Evaluator Responsibilities**:
- Execute AST nodes and produce runtime values
- Manage ExecutionContext (environment, call stack, control flow)
- Handle 99% of assignments via `ctx.Env().Set()`
- Binary/unary operations, type conversions, control flow

**Interpreter Responsibilities**:
- OOP semantics (classes, objects, Self context)
- Class registry and hierarchy management
- Method dispatch and virtual method tables
- Property evaluation and field lookup
- Handle 1% of assignments requiring global OOP knowledge

**Runtime Package**:
- Pure value types with no AST dependencies
- RefCountManager interface for lifecycle management
- Environment for variable storage

---

## Performance Characteristics

### Benchmark Results (Task 3.5.35)

| Operation Type | ns/op | B/op | allocs/op | Status |
|---------------|-------|------|-----------|--------|
| **Literals** |
| Integer | 10.13 | 8 | 1 | ✅ Target: <20 ns |
| Float | 9.29 | 8 | 1 | ✅ Target: <20 ns |
| String | 17.40 | 16 | 1 | ✅ Target: <20 ns |
| Boolean | 6.90 | 1 | 1 | ✅ Target: <20 ns |
| **Binary Operations** |
| Integer Add | 52.33 | 24 | 3 | ✅ Target: <100 ns |
| Integer Multiply | 52.58 | 24 | 3 | ✅ Target: <100 ns |
| Integer Comparison | 55.81 | 24 | 3 | ✅ Target: <100 ns |
| Boolean AND | 35.79 | 3 | 3 | ✅ Target: <100 ns |
| String Concat | 118.8 | 64 | 4 | ✅ Target: <200 ns |
| **Unary Operations** |
| Integer Negation | 29.81 | 16 | 2 | ✅ Target: <50 ns |
| Boolean NOT | 20.61 | 2 | 2 | ✅ Target: <50 ns |

**Key Findings**:
- Zero performance regression vs switch-based approach
- Linear scaling with expression complexity (~50 ns per operation)
- Minimal allocations (1-4 per operation)
- Go compiler effectively optimizes visitor dispatch

---

## Adapter Interface Evolution

### Before Phase 3.5
- 75+ adapter methods
- Heavy delegation for all complex operations
- Evaluator heavily dependent on interpreter

### After Phase 3.5 (Current)
- ~5 essential adapter methods
- 95%+ reduction in adapter usage
- Clear architectural boundaries

### Essential Adapter Methods (Remaining)

```go
type InterpreterAdapter interface {
    // EvalNode - 2 essential calls for Self/class context
    // Lines 93, 368 in assignment_helpers.go
    EvalNode(node ast.Node) Value

    // ExecuteUserFunctionBody - user function execution
    ExecuteUserFunctionBody(funcDecl *ast.FunctionDecl,
                           args []Value,
                           ctx *ExecutionContext) Value

    // ~3 additional methods for class/object operations
}
```

---

## EvalNode Calls Audit (Task 3.5.38)

### Distribution Summary

**Total calls**: 34 across 12 files

**By Category**:
```
Reference Counting (MIGRATED)    7 calls  ✅ Eliminated via RefCountManager
Self/Class Context (ESSENTIAL)   2 calls  ⚠️ Architectural boundary
Member Access                    6 calls  ⏳ Future (Phase 4+)
Member Assignment                6 calls  ⏳ Future (Phase 4+)
Index Operations                 3 calls  ⏳ Could migrate (medium effort)
Function Calls                   2 calls  ⚠️ Keep (architectural)
Control Flow                     2 calls  ⏳ Quick win potential
Other (helpers, declarations)    6 calls  ⏳ Various priorities
```

### The 2 Essential EvalNode Calls

**Location**: `internal/interp/evaluator/assignment_helpers.go`
- Line 93: Simple assignment to non-env variable
- Line 368: Compound assignment to non-env variable

**Why They Remain**:
When evaluator encounters `x := value` but `x` is not in environment:
1. Could be instance field: `Self.Field := value`
2. Could be class variable: `TClass.ClassVar := value`
3. Could be property: `Self.PropName := value`

**Architectural Rationale**:
- Requires `ClassInfo.ClassVars` access (interpreter package)
- Requires `obj.Class.LookupField()` method (interpreter logic)
- Requires `evalPropertyWrite()` logic (property semantics)
- ALL of this is OOP semantics, not evaluator concern

**Decision**: Keep as architectural boundary (Task 3.5.43, Option 3)

---

## Migration Success Story: Reference Counting

### Before (7 EvalNode calls)
Lines 127, 136, 172, 182, 192, 227, 254 in `assignment_helpers.go`

**Problem**:
- Evaluator called `adapter.EvalNode()` for all object/interface assignments
- Reference count management tightly coupled to interpreter

### After (0 EvalNode calls)
**Solution**:
1. Created `runtime.RefCountManager` interface
2. Moved ref counting logic to runtime implementation
3. Passed RefCountManager to evaluator via ExecutionContext
4. Replaced all adapter calls with `ctx.RefCountManager().IncrementRef()` etc.

**Result**:
- ✅ 7 adapter calls eliminated
- ✅ Clean separation (runtime manages value lifecycle)
- ✅ Zero violations of architectural boundaries
- ✅ ~40 hours effort (Tasks 3.5.39-3.5.42)

**Key Insight**: Migrate when abstraction is **clean and self-contained**, preserve when it would **violate architectural boundaries**.

---

## File Organization

```
internal/interp/evaluator/
├── evaluator.go              # Core evaluator, visitor infrastructure
├── visitor_expressions.go    # Expression visitor methods (48+)
├── visitor_statements.go     # Statement visitor methods
├── visitor_declarations.go   # Declaration visitor methods
├── visitor_literals.go       # Literal value creation
├── binary_ops.go             # Binary operation implementations
├── context.go                # ExecutionContext
├── callstack.go              # Call stack management
├── assignment_helpers.go     # Assignment logic (2 essential calls)
├── helpers.go                # Helper functions
├── benchmark_test.go         # Performance benchmarks
└── *_test.go                 # Unit tests
```

---

## Working with the Evaluator

### Adding a New Expression Type

1. **Define AST node** in `pkg/ast/expressions.go`
2. **Add visitor method** in `internal/interp/evaluator/visitor_expressions.go`
3. **Add dispatcher** in `VisitExpression()` switch
4. **Write tests** with expected behavior
5. **Add benchmark** if it's a hot path

### ExecutionContext API

```go
// Variables
val, ok := ctx.Env().Get("x")
ctx.Env().Define("y", value)
ctx.Env().Set("z", newValue)

// Scoping
newEnv := ctx.Env().NewEnclosedEnvironment()
ctx.PushEnv(newEnv)
defer ctx.PopEnv()

// Call stack
ctx.GetCallStack().Push("funcName", node.Pos())
defer ctx.GetCallStack().Pop()

// Control flow
ctx.GetControlFlow().SetBreak()
ctx.GetControlFlow().SetContinue()
ctx.GetControlFlow().SetReturn(value)
```

### Common Patterns

**Error Handling**:
```go
result := e.VisitExpression(node.SubExpr, ctx)
if isError(result) {
    return result // Propagate error up
}
```

**Control Flow Checking**:
```go
for {
    result := e.VisitStatement(stmt, ctx)

    if ctx.GetControlFlow().ShouldBreak() {
        ctx.GetControlFlow().ClearBreak()
        break
    }
}
```

---

## Migration Decision Matrix

| Concern Type | Evaluator Responsible? | Action |
|--------------|------------------------|--------|
| Value operations | ✅ Yes | Migrate to evaluator |
| Type conversions | ✅ Yes | Migrate to evaluator |
| Control flow | ✅ Yes | Migrate to evaluator |
| Reference counting | ✅ Yes (value lifecycle) | Migrate via interface |
| Property callbacks | ✅ Yes (with callback) | Migrate via callback |
| Self/class context | ❌ No (OOP semantics) | Keep in interpreter |
| Method dispatch | ❌ No (requires VMT) | Keep in interpreter |
| Class registration | ❌ No (global state) | Keep in interpreter |

---

## Current Migration Effort

### Completed (Phase 3.5)

✅ **Task 3.5.1-3.5.4**: Visitor pattern infrastructure (21 methods)
✅ **Task 3.5.35**: Performance validation (all targets met)
✅ **Task 3.5.38**: EvalNode calls audit (34 calls documented)
✅ **Task 3.5.39-3.5.42**: Reference counting migration (7 calls eliminated, ~40 hours)
✅ **Task 3.5.43**: Self context analysis (2 calls preserved, ~3 hours)

**Total Phase 3.5 effort**: ~180 hours

### Future Work

**Phase 4+** (OOP Infrastructure Migration):
- Member access delegation (6 calls)
- Member assignment delegation (6 calls)
- Estimated effort: 120-180 hours

**Optional Quick Wins**:
- Index operations (3 calls, 6-8 hours)
- Compound assignment simplification (2 calls, 3-4 hours)

---

## Performance Optimization Guidelines

### Do's
✅ Short-circuit boolean operators
✅ Cache frequently used values
✅ Profile before optimizing
✅ Minimize allocations in hot paths

### Don'ts
❌ Premature optimization
❌ Switching back to switch statements
❌ Complex caching schemes
❌ Modifying evaluator state

### Profiling Commands
```bash
# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./internal/interp/evaluator
go tool pprof cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof ./internal/interp/evaluator
go tool pprof mem.prof
```

---

## Testing

### Running Tests
```bash
# Unit tests
go test ./internal/interp/evaluator
go test -run TestVisitBinaryExpression ./internal/interp/evaluator

# Benchmarks
go test -bench=. -benchmem ./internal/interp/evaluator -run=^$
go test -bench=BenchmarkVisitBinaryExpression ./internal/interp/evaluator

# Integration tests
go test ./internal/interp
go test -v ./internal/interp -run TestDWScriptFixtures
```

### Test Coverage
- Evaluator package: >80%
- Individual visitor methods: >90%
- Performance benchmarks: 17 comprehensive benchmarks

---

## Common Mistakes to Avoid

❌ **Forgetting error propagation**: Always check `isError(value)` after sub-evaluations
❌ **Not checking control flow**: Loop statements must check `ShouldBreak()`, `ShouldContinue()`
❌ **Modifying shared state**: Use ExecutionContext, not evaluator fields
❌ **Not using defer for cleanup**: Always `defer ctx.PopEnv()` after `ctx.PushEnv()`

---

## Key Takeaways

1. **Visitor pattern is production-ready**: Zero performance regression, excellent maintainability
2. **Adapter interface is essential, not technical debt**: Represents clean architectural boundary
3. **Reference counting migration was successful**: Template for future migrations
4. **2 EvalNode calls are intentional**: Self/class context is OOP semantics (interpreter concern)
5. **95%+ adapter reduction achieved**: From 75+ methods to ~5 essential methods

---

## Related Documentation

- [CLAUDE.md](../../CLAUDE.md) - Project overview and development guidelines
- [PLAN.md](../../PLAN.md) - Phase 3.5 task breakdown
- [interpreter.md](interpreter.md) - Interpreter architecture
- [../evalnode-audit-final.md](../evalnode-audit-final.md) - Complete EvalNode calls inventory (detailed)
- [../evaluator-architecture.md](../evaluator-architecture.md) - Deep dive on architectural boundaries

---

**Last Updated**: 2025-12-04
**Status**: Current and authoritative
**Supersedes**: Individual eval* documents (now consolidated here)
