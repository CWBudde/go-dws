# Environment Dual System Audit

**Date**: 2025-12-09
**Task**: 3.1.1 Deep Audit of Environment Usage Patterns
**Purpose**: Understand the complexity of merging `i.env` and `ctx.env` before attempting migration

---

## Executive Summary

The codebase has **two independent environment systems** that must be kept synchronized:

| System | Type | Location | Usages |
|--------|------|----------|--------|
| `i.env` | `*interp.Environment` (concrete) | Interpreter | 395 |
| `ctx.Env()` | `evaluator.Environment` (interface) | ExecutionContext | 128 |

**Key Finding**: These are NOT just two variables - they are two different type systems bridged by `EnvironmentAdapter` (137 LOC).

---

## 1. Variable Access Patterns

### 1.1 i.env.Get() - 92 occurrences

| File | Count | Purpose |
|------|-------|---------|
| objects_properties.go | 12 | Property value lookup |
| functions_records.go | 12 | Record field access |
| objects_hierarchy.go | 10 | Class hierarchy navigation |
| functions_calls.go | 9 | Function parameter lookup |
| statements_assignments.go | 6 | Assignment target resolution |
| objects_methods.go | 6 | Method self/context lookup |
| statements_declarations.go | 4 | Variable declaration check |
| helpers_conversion.go | 4 | Type conversion helpers |

### 1.2 i.env.Set() - 12 occurrences

| File | Count | Purpose |
|------|-------|---------|
| exceptions.go | 4 | ExceptObject binding |
| statements_loops.go | 2 | Loop variable updates |
| functions_calls.go | 2 | Parameter binding |
| lvalue.go | 1 | L-value updates |
| builtins_context.go | 1 | Built-in context |

### 1.3 i.env.Define() - 94 occurrences

| File | Count | Purpose |
|------|-------|---------|
| objects_properties.go | 33 | Property definitions |
| functions_records.go | 14 | Record method bindings |
| operators_eval.go | 8 | Operator temp variables |
| statements_loops.go | 7 | Loop variable definitions |
| helpers_validation.go | 7 | Helper temp bindings |
| objects_hierarchy.go | 7 | Class hierarchy bindings |

---

## 2. Scope Management Patterns

### 2.1 Scope Entry (112 patterns total)

| Pattern | Count | Files |
|---------|-------|-------|
| `savedEnv := i.env` | 52 | 15 files |
| `i.PushEnvironment()` | 59 | 18 files (incl. 7 tests) |
| `oldEnv := i.env` | 1 | exceptions.go |

**Top files by scope entry count**:
- objects_properties.go: 11
- adapter_methods.go: 9
- declarations.go: 4

### 2.2 Scope Exit - 101 RestoreEnvironment() calls

| File | Count | Notes |
|------|-------|-------|
| objects_methods.go | 16 | Method calls |
| adapter_methods.go | 15 | Adapter callbacks |
| statements_loops.go | 13 | For/while/repeat |
| objects_properties.go | 11 | Property getters/setters |
| functions_records.go | 8 | Record methods |
| functions_pointers.go | 5 | Function pointer calls |

### 2.3 Problematic Pattern: Manual Save/Restore

```go
// This pattern appears 52+ times:
savedEnv := i.env
methodEnv := i.PushEnvironment(i.env)
// ... use methodEnv ...
i.RestoreEnvironment(savedEnv)
```

**Why it's problematic**:
1. Must manually remember to restore
2. Easy to forget in error paths
3. Creates 2 environment references (savedEnv + methodEnv)
4. Must be synchronized with ctx.env via SetEnvironment()

---

## 3. Type System Mismatch

### 3.1 Two Different Environment Types

**Interpreter** (`internal/interp/environment.go`):
```go
type Environment struct {
    store map[string]Value  // interp.Value
    outer *Environment
}

func (e *Environment) Get(name string) (Value, bool)
func (e *Environment) Set(name string, val Value) error
func (e *Environment) Define(name string, val Value)
```

**Evaluator** (`internal/interp/evaluator/context.go`):
```go
type Environment interface {
    Define(name string, value any)       // any, not Value!
    Get(name string) (any, bool)         // any, not Value!
    Set(name string, value any) bool     // bool, not error!
    NewEnclosedEnvironment() Environment
}
```

### 3.2 The Bridge: EnvironmentAdapter (137 LOC)

```go
type EnvironmentAdapter struct {
    underlying interface{}  // Actually *interp.Environment
}

func (ea *EnvironmentAdapter) Define(name string, value interface{}) {
    // Convert interface{} → runtime.Value
    // Call underlying.Define()
}
```

**Value conversion flow**:
- Interpreter uses `interp.Value` (legacy type aliases)
- Evaluator uses `runtime.Value` (new canonical types)
- EnvironmentAdapter converts between them

---

## 4. Bidirectional Control Flow

### 4.1 Flow: Interpreter → Evaluator → Adapter → Interpreter

```
Interpreter.Eval()
  → Evaluator.Eval()
     → VisitMethodCall()
        → e.adapter.CallMethod()  ← adapter callback!
           → Interpreter.CallMethod()  ← back to interpreter!
              → savedEnv := i.env
              → i.PushEnvironment()
              → ... execute ...
              → i.RestoreEnvironment(savedEnv)
```

### 4.2 Adapter Usages in Evaluator: 157 calls across 36 files

**Top adapter methods called**:
| Method | Count | Purpose |
|--------|-------|---------|
| `adapter.EvalNode()` | ~25 | Fallback to interpreter |
| `adapter.CallMethod()` | ~10 | Method dispatch |
| `adapter.ExecuteMethodWithSelf()` | ~5 | Self-context methods |
| Class/Interface declaration methods | ~43 | visitor_declarations.go |

### 4.3 Synchronization Points

The interpreter has three sync methods (all in `interpreter.go`):

```go
func (i *Interpreter) SetEnvironment(env *Environment) {
    i.env = env
    i.ctx.SetEnv(evaluator.NewEnvironmentAdapter(env))  // ← sync!
}

func (i *Interpreter) PushEnvironment(parent *Environment) *Environment {
    newEnv := NewEnclosedEnvironment(parent)
    i.SetEnvironment(newEnv)  // ← uses sync
    return newEnv
}

func (i *Interpreter) RestoreEnvironment(saved *Environment) {
    i.SetEnvironment(saved)  // ← uses sync
}
```

---

## 5. Evaluator's Own Environment Stack

### 5.1 ExecutionContext has its own stack

```go
type ExecutionContext struct {
    env      Environment      // Current environment (adapter)
    envStack []Environment    // Stack for PushEnv/PopEnv
    // ...
}

func (ctx *ExecutionContext) PushEnv() Environment
func (ctx *ExecutionContext) PopEnv() Environment
```

### 5.2 Two Independent Stacks

| Stack | Location | Type | Purpose |
|-------|----------|------|---------|
| Manual save/restore | Interpreter | `*Environment` | 52+ patterns |
| `ctx.envStack` | ExecutionContext | `[]Environment` | 7 usages |

**Risk**: If Interpreter does `savedEnv := i.env` without going through sync methods, the two stacks diverge.

---

## 6. Files Requiring Most Work

### 6.1 Top 10 by i.env usage

| File | i.env | Patterns | Priority |
|------|-------|----------|----------|
| objects_properties.go | 67 | 11 save/restore | HIGH |
| functions_records.go | 33 | 2 save/restore | HIGH |
| objects_hierarchy.go | 25 | 2 save/restore | MEDIUM |
| adapter_methods.go | 22 | 9 save/restore | HIGH |
| functions_calls.go | 22 | - | MEDIUM |
| objects_methods.go | 20 | 6 save/restore | HIGH |
| statements_loops.go | 19 | 2 save/restore | MEDIUM |
| operators_eval.go | 16 | 2 save/restore | MEDIUM |

### 6.2 Adapter Callback Files (bidirectional flow)

These files in evaluator call back to interpreter via adapter:
- visitor_declarations.go: 43 adapter calls
- visitor_expressions_functions.go: 11 adapter calls
- visitor_statements.go: 6 adapter calls
- member_assignment.go: 6 adapter calls
- helper_methods.go: 6 adapter calls

---

## 7. Recommended Migration Strategy

### Phase A: Unify Environment Type

**Option C (Recommended)**: Move `Environment` to `internal/interp/runtime/`
- No circular dependencies
- Both packages can import it
- Clean separation

### Phase B: Eliminate EnvironmentAdapter

1. Update `ExecutionContext.env` to use `*runtime.Environment` directly
2. Delete `env_adapter.go` (137 LOC)
3. Update all `NewEnvironmentAdapter()` calls

### Phase C: Unify Scope Management

1. Replace 52 manual `savedEnv := i.env` patterns with `ctx.PushEnv()`
2. Replace 101 `RestoreEnvironment()` calls with `ctx.PopEnv()`
3. Delete `PushEnvironment()`, `RestoreEnvironment()`, `SetEnvironment()`

### Phase D: Delete i.env

1. Add `Env()` method returning `i.ctx.Env()`
2. Replace all `i.env` with `i.Env()`
3. Delete `env` field from Interpreter struct

---

## 8. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Circular import | HIGH | Use Option C (shared package) |
| Scope stack divergence | HIGH | Migrate all patterns in one phase |
| Adapter callback complexity | MEDIUM | Migrate adapter methods last |
| Test failures | MEDIUM | Run tests after each file |
| Performance regression | LOW | Eliminating adapter should improve |

---

## Appendix: Raw Counts

```
i.env.Get():    92 occurrences in 24 files
i.env.Set():    12 occurrences in  7 files
i.env.Define(): 94 occurrences in 15 files
i.env.GetLocal: 6 occurrences in  1 file (test only)
─────────────────────────────────────────
Total i.env.X(): 204 method calls

savedEnv := i.env:    52 occurrences in 15 files
i.PushEnvironment():  59 occurrences in 18 files
i.RestoreEnvironment(): 101 occurrences in 19 files
i.SetEnvironment():    8 occurrences in  4 files
i.env = :              3 occurrences in  1 file (interpreter.go)
─────────────────────────────────────────
Total scope patterns: 223

ctx.Env():           128 occurrences in 20 files (evaluator)
ctx.envStack:          7 occurrences in  1 file (context.go)
adapter. calls:      157 occurrences in 36 files (evaluator)
```
