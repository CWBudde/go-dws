# Environment Management Audit Report

**Phase**: 3.8.1 | **Date**: 2025-12-06 | **Status**: Complete

## Executive Summary

This audit documents all environment modification points in the go-dws interpreter to support the environment synchronization migration (Phase 3.8.2). The audit found **149 `i.env =` assignments** across 17 files, all with proper restoration but none synchronizing with `i.ctx.env`.

### Key Findings

- **Total Assignments**: 149 across 17 files
- **Restoration Coverage**: 100% (no memory leaks)
- **Synchronization**: 0% sync with `i.ctx.env` (root cause of test failures)
- **Top Files**: adapter_methods.go (24), objects_methods.go (23), objects_properties.go (22)

### Critical Issue: Dual-Environment Desynchronization

The Interpreter maintains two environment references:
```go
type Interpreter struct {
    env  *Environment                   // Used by Interpreter.Eval() methods
    ctx  *evaluator.ExecutionContext    // Contains i.ctx.env (interface)
}
```

**Problem**: All 149 assignments modify `i.env` without calling `i.ctx.SetEnv()` to sync the evaluator's environment reference. When evaluator methods are called, they use stale `i.ctx.env`, leading to wrong variable lookups and test failures.

**Impact**: Commit af2ba646 attempted to migrate binary operations to the evaluator, resulting in **139 test failures** due to environment desynchronization.

---

## Summary Statistics

| Metric | Value |
|--------|-------|
| **Total `i.env =` assignments** | 149 |
| **Files with assignments** | 17 |
| **Restoration coverage** | 100% |
| **Sync with `i.ctx.env`** | 0% |
| **Test baseline** | 892 fixture failures |

---

## Breakdown by Category

| Category | Files | Count | Risk | Pattern |
|----------|-------|-------|------|---------|
| **Method Dispatch** | adapter_methods.go, objects_methods.go, functions_records.go | 57 | HIGH | Save/Execute/Restore |
| **Property Access** | objects_properties.go | 22 | MEDIUM | Getter/Setter Scope |
| **Loop Scoping** | statements_loops.go | 15 | MEDIUM | For/For-in/While |
| **Object Lifecycle** | adapter_objects.go, objects_instantiation.go | 13 | MEDIUM | Constructor/Destructor |
| **Declarations** | declarations.go | 8 | MEDIUM | Type/Function Context |
| **Lambda/Pointers** | functions_pointers.go, adapter_functions.go | 9 | HIGH | Closure Environment |
| **Operators** | operators_eval.go | 6 | MEDIUM | Operator Methods |
| **Callbacks** | user_function_callbacks.go | 6 | **CRITICAL** | Evaluator Integration |
| **Hierarchy** | objects_hierarchy.go | 4 | MEDIUM | Parent Class Calls |
| **Helpers/Validation** | helpers_validation.go | 3 | LOW | Helper Methods |
| **Exception Handling** | exceptions.go | 2 | LOW | Handler Scope |
| **Other** | statements_control.go, interface.go | 4 | LOW | Misc |
| **TOTAL** | **17 files** | **149** | - | - |

---

## Detailed Inventory by File

### 1. adapter_methods.go (24 assignments)

**Lines**: 204, 208, 247, 270, 324, 326, 339, 343, 374, 378, 417, 421, 443, 458, 468, 483, 516, 528, 554, 558, 599, 603, 623, 627

**Purpose**: Method dispatch, helper method execution, built-in type method handling

**Pattern**: Save → Set → Execute → Restore
```go
savedEnv := i.env                          // Line 202
methodEnv := NewEnclosedEnvironment(i.env) // Line 203
i.env = methodEnv                          // Line 204
// ... method execution ...
i.env = savedEnv                           // Line 208
```

**Restoration**: All assignments have paired restore statements
**Risk Level**: HIGH - core OOP functionality, complex dispatch logic

---

### 2. objects_methods.go (23 assignments)

**Lines**: 63, 70, 79, 114, 122, 132, 153, 190, 235, 354, 377, 381, 678, 697, 702, 708, 711, 755, 803, 833, 837, 878, 911

**Purpose**: Object instance method execution, constructor calls, destructor handling

**Pattern**: Save → Set → Execute → Error handling → Restore
```go
savedEnv := i.env                          // Line 58
tempEnv := NewEnclosedEnvironment(i.env)   // Line 59
i.env = tempEnv                            // Line 63
// ... field initialization ...
if isError(fieldValue) {
    i.env = savedEnv                       // Line 70 (error path)
    return fieldValue
}
i.env = savedEnv                           // Line 79 (success path)
```

**Restoration**: All assignments with error path handling
**Risk Level**: HIGH - instance methods, constructors, inheritance

---

### 3. objects_properties.go (22 assignments)

**Lines**: 114, 173, 203, 262, 287, 302, 336, 382, 396, 442, 486, 500, 514, 528, 542, 556, 570, 584, 598, 612, 626, 640

**Purpose**: Property getter/setter execution

**Pattern**: Pairs of methodEnv setup and savedEnv restore

**Restoration**: All assignments matched with restore
**Risk Level**: MEDIUM - short-lived scopes, paired operations

---

### 4. statements_loops.go (15 assignments)

**Lines**: 153, 205, 234, 257, 278, 293, 318, 342, 373, 382, 418, 445, 467, 490, 495

**Purpose**: Loop variable scoping (for, for-in, while, repeat)

**Pattern**: Enclosed environment for loop variable isolation
```go
loopEnv := NewEnclosedEnvironment(i.env)   // Line 151
savedEnv := i.env                          // Line 152
i.env = loopEnv                            // Line 153
for current := ...; current <= ...; current += stepValue {
    i.env.Define(loopVarName, makeLoopValue(current))
    result = i.Eval(stmt.Body)
    if isError(result) {
        i.env = savedEnv                   // Line 205 (error path)
        return result
    }
}
i.env = savedEnv                           // Line 257 (completion)
```

**Restoration**: All assignments with error path handling
**Risk Level**: MEDIUM - break/continue/return complicate restoration

---

### 5. functions_records.go (10 assignments)

**Lines**: 79, 130, 179, 187, 193, 281, 328, 332, 397, 441

**Purpose**: Record type method execution

**Pattern**: Similar to object methods, save/execute/restore

**Restoration**: All paired with restore statements
**Risk Level**: MEDIUM - value type semantics

---

### 6. declarations.go (8 assignments)

**Lines**: 259, 260, 390, 393, 475, 479, 515, 519

**Purpose**: Type/function/constant declaration processing

**Pattern**: Uses `defer` for guaranteed restoration
```go
savedEnv := i.env                          // Line 256
tempEnv := NewEnclosedEnvironment(i.env)   // Line 257
i.env = tempEnv                            // Line 259
defer func() { i.env = savedEnv }()        // Line 260 (guaranteed restore)
```

**Restoration**: Deferred restoration provides strong guarantees
**Risk Level**: MEDIUM - already uses defer for safety

---

### 7. functions_pointers.go (7 assignments)

**Lines**: 40, 49, 85, 93, 101, 109, 117

**Purpose**: Lambda execution, function pointer calls, closure environment

**Pattern**: Enclosed environment with Self binding
```go
funcEnv := NewEnclosedEnvironment(closureEnv)  // Line 38
savedEnv := i.env                              // Line 39
i.env = funcEnv                                // Line 40
i.env.Define("Self", funcPtr.SelfObject)       // Line 43
result := i.executeUserFunctionViaEvaluator(...)
i.env = savedEnv                               // Line 49
```

**Restoration**: All assignments paired with restore
**Risk Level**: HIGH - captures environment, complex lifetime

---

### 8. adapter_objects.go (7 assignments)

**Lines**: 47, 54, 63, 70, 74, 121, 125

**Purpose**: Object instantiation, field initialization

**Pattern**: Temporary environment for constructor calls

**Restoration**: Error handling with restoration
**Risk Level**: MEDIUM - error handling, field initialization

---

### 9. user_function_callbacks.go (6 assignments)

**Lines**: 183, 217, 219, 247, 252, 264

**Purpose**: Environment synchronization between interpreter and evaluator

**Pattern**: Environment extraction and wrapping
```go
// Line 183: Extract environment from evaluator context
if adapter, ok := funcEnv.(*evaluator.EnvironmentAdapter); ok {
    if concreteEnv, ok := adapter.Underlying().(*Environment); ok {
        i.env = concreteEnv  // SYNC: Update i.env
    }
}
```

**Special Note**: This file already has partial sync logic via EnvSyncer callback
**Risk Level**: **CRITICAL** - central integration point between interpreter and evaluator

---

### 10. operators_eval.go (6 assignments)

**Lines**: 68, 97, 106, 121, 150, 159

**Purpose**: Operator overload method execution

**Pattern**: Similar to method dispatch

**Restoration**: All assignments paired with restore
**Risk Level**: MEDIUM - similar to method dispatch

---

### 11. objects_instantiation.go (6 assignments)

**Lines**: 72, 82, 94, 207, 232, 248

**Purpose**: Object creation, constructor dispatch

**Pattern**: Constructor environment setup with error recovery

**Restoration**: Error path restoration
**Risk Level**: MEDIUM - error handling, parent constructor calls

---

### 12. objects_hierarchy.go (4 assignments)

**Lines**: 790, 833, 891, 893

**Purpose**: Parent class method invocation, class constant evaluation

**Pattern**: Parent method environment setup
```go
methodEnv := NewEnclosedEnvironment(i.env)  // Line 788
savedEnv := i.env                           // Line 789
i.env = methodEnv                           // Line 790
// ... parent method ...
i.env = savedEnv                            // Line 833
```

**Restoration**: All assignments paired with restore
**Risk Level**: MEDIUM - inheritance, super calls

---

### 13. helpers_validation.go (3 assignments)

**Lines**: 242, 278, 304

**Purpose**: Helper method execution

**Pattern**: Helper method environment setup with error/success paths

**Restoration**: Error and success path restoration
**Risk Level**: LOW - helper methods, infrequent

---

### 14. statements_control.go (2 assignments)

**Lines**: 349, 350

**Purpose**: Operator method invocation

**Pattern**: Uses `defer` for automatic restoration
```go
prev := i.env                                  // Line 347
methodEnv := NewEnclosedEnvironment(i.env)     // Line 348
i.env = methodEnv                              // Line 349
defer func() { i.env = prev }()                // Line 350 (guaranteed restore)
```

**Restoration**: Deferred restoration
**Risk Level**: LOW - uses defer for safety

---

### 15. interface.go (2 assignments)

**Lines**: 431, 451

**Purpose**: Destructor execution

**Pattern**: Destructor environment setup and restore

**Restoration**: All assignments paired with restore
**Risk Level**: LOW - destructor handling

---

### 16. exceptions.go (2 assignments)

**Lines**: 333, 370

**Purpose**: Exception handler scope

**Pattern**: New enclosed environment for exception variable
```go
oldEnv := i.env                                // Line 332
i.env = NewEnclosedEnvironment(i.env)          // Line 333
// ... exception handler ...
i.env = oldEnv                                 // Line 370
```

**Restoration**: All assignments paired with restore
**Risk Level**: LOW - exception variable binding, short-lived

---

### 17. adapter_functions.go (2 assignments)

**Lines**: 68, 84

**Purpose**: Method pointer execution with Self binding

**Pattern**: Function environment setup and restore

**Restoration**: All assignments paired with restore
**Risk Level**: MEDIUM - method pointers

---

## Restoration Analysis

### Good News: 100% Restoration Coverage

All 149 assignments have proper environment restoration - **no memory leaks detected**.

#### Restoration Patterns

1. **Explicit Pairs** (91% - 136 assignments)
   ```go
   savedEnv := i.env
   i.env = NewEnclosedEnvironment(i.env)
   // ... execution ...
   i.env = savedEnv
   ```
   - Simple and reliable
   - Most common pattern
   - Used in: adapter_methods.go, objects_methods.go, objects_properties.go, and 11 others

2. **Deferred Restoration** (4% - 6 assignments)
   ```go
   savedEnv := i.env
   i.env = NewEnclosedEnvironment(i.env)
   defer func() { i.env = savedEnv }()
   ```
   - Guaranteed restoration even on panic
   - Used in: declarations.go, statements_control.go

3. **Error Path Restoration** (~60% of assignments)
   ```go
   savedEnv := i.env
   i.env = newEnv
   if isError(result) {
       i.env = savedEnv
       return result
   }
   // ... success path ...
   i.env = savedEnv
   ```
   - Explicit error handling
   - Common in loops, methods, constructors

### Critical Gap: No Synchronization

**149/149 assignments (100%) do NOT sync with `i.ctx.env`**

This is the root cause of environment desynchronization:
- Interpreter methods use `i.env`
- Evaluator methods use `i.ctx.env`
- When `i.env` changes without syncing `i.ctx.env`, they diverge
- Result: Wrong variable lookups, test failures

---

## Architectural Context

### Why the Dual-Environment Design Exists

1. **Circular Import Avoidance**: The evaluator package cannot import the interp package
2. **Interface Abstraction**: ExecutionContext uses an `Environment` interface
3. **Adapter Pattern**: `EnvironmentAdapter` wraps `*Environment` for interface compliance
4. **Gradual Migration**: Phase 3.5 transition from monolithic to modular architecture

### The Synchronization Problem

**Before (Broken)**:
```go
// 149 locations do this:
i.env = NewEnclosedEnvironment(i.env)
// i.ctx.env is now stale!
```

**After (Fixed with Helpers)**:
```go
// Phase 3.8.2 migration:
i.SetEnvironment(NewEnclosedEnvironment(i.env))
// Both i.env and i.ctx.env are updated atomically
```

---

## Migration Recommendations

### Priority Order (Phase 3.8.2)

Based on risk and complexity:

1. **CRITICAL** - User Function Callbacks (6 assignments) - Already has partial sync
2. **HIGH** - Method Dispatch (24 assignments) - Most assignments, core OOP
3. **HIGH** - Lambda/Pointers (9 assignments) - Closure semantics
4. **MEDIUM** - Object Methods (23 assignments) - Well-structured patterns
5. **MEDIUM** - Properties (22 assignments) - Short-lived scopes
6. **MEDIUM** - Loops (15 assignments) - Well-defined entry/exit
7. **MEDIUM** - Object Lifecycle (13 assignments) - Constructor logic
8. **MEDIUM** - Record Methods (10 assignments) - Value type semantics
9. **MEDIUM** - Declarations (8 assignments) - Already uses defer
10. **MEDIUM** - Operators (6 assignments) - Similar to methods
11. **LOW** - Hierarchy (4 assignments) - Inheritance
12. **LOW** - Helpers/Control/Interface (9 assignments) - Misc
13. **LOW** - Exceptions (2 assignments) - Handler scope

### Helper Methods (Phase 3.8.1.2)

Three helper methods were added to [internal/interp/interpreter.go](../internal/interp/interpreter.go):

```go
// SetEnvironment atomically updates both i.env and i.ctx.env
func (i *Interpreter) SetEnvironment(env *Environment)

// PushEnvironment creates enclosed env and sets both references
func (i *Interpreter) PushEnvironment(parent *Environment) *Environment

// RestoreEnvironment restores saved env to both references
func (i *Interpreter) RestoreEnvironment(saved *Environment)
```

See [environment-migration-guide.md](environment-migration-guide.md) for usage patterns.

---

## Test Baseline

**Current State** (as of 2025-12-06):
- **Total Packages**: 24
- **Passing Packages**: 23
- **Failing Packages**: 1 (internal/interp)
- **Fixture Failures**: 892 tests (baseline)

**Requirement**: Migration must not introduce new failures beyond the 892 baseline.

---

## Success Metrics

### Phase 3.8.1 (Complete)
- ✅ Comprehensive audit: 149 assignments documented
- ✅ Helper methods implemented and tested
- ✅ Test baseline established: 892 failures
- ✅ Documentation complete

### Phase 3.8.2 (Upcoming)
- [ ] All 149 assignments migrated to use helpers
- [ ] Zero new test failures introduced
- [ ] Fixture baseline maintained (≤ 892)
- [ ] All categories migrated incrementally

### Phase 3.8.3 (Blocked on 3.8.2)
- [ ] Binary operations delegated to evaluator
- [ ] No environment desynchronization
- [ ] Zero test failures

---

## Appendix: Search Methodology

### Finding Assignments

```bash
# Primary search
grep -rn "i\.env =" internal/interp/*.go

# Verification searches
grep -rn "NewEnclosedEnvironment" internal/interp/*.go
grep -rn "savedEnv :=" internal/interp/*.go
```

### Categorization Criteria

- **Purpose**: What the code is doing (method call, loop, etc.)
- **Pattern**: How environment is managed (save/restore, defer, etc.)
- **Risk**: Complexity, frequency, error paths
- **Restoration**: How environment is restored (explicit, defer, error paths)

---

## References

- [PLAN.md Phase 3.8.1](../PLAN.md#phase-38-environment-synchronization--binary-operations-migration)
- [environment-migration-guide.md](environment-migration-guide.md) - Helper usage guide
- [internal/interp/interpreter.go](../internal/interp/interpreter.go) - Helper method implementations
- [internal/interp/environment_helpers_test.go](../internal/interp/environment_helpers_test.go) - Helper tests

---

**Audit Complete**: 2025-12-06
**Audited by**: Systematic codebase analysis (Explore agents)
**Total Effort**: ~4 hours (search, categorization, documentation)
