# EvalNode Calls Audit - Post-Consolidation Report

**Task**: 3.12.1
**Date**: 2025-12-08
**Purpose**: Updated audit after Phases 3.9-3.11 consolidation to inform Phase 3.12 migration strategy

---

## Executive Summary

Phase 3 refactoring has successfully reduced EvalNode calls from **34 → 27** (20.6% reduction). This audit documents all remaining calls and provides a roadmap for Phase 3.12 migration.

### Progress Overview

| Metric | Phase 3.5 Baseline | Current (Post-3.11) | Phase 3.12 Target | Progress |
|--------|-------------------|---------------------|-------------------|----------|
| **Total EvalNode Calls** | 34 | 27 | ~15 | 79% → 56% |
| **Adapter Methods** | 72 | 63 | ~50 | 100% → 88% → 69% (target) |
| **Code Size (interp/)** | Baseline | -671 LOC | TBD | -671 LOC |

**Key Achievements**:
- ✅ 7 reference counting calls eliminated (Phase 3.5.39-3.5.42 - RefCountManager)
- ✅ 251 LOC removed (Phase 3.9 - identifier resolution)
- ✅ 269 LOC removed (Phase 3.10 - dead code)
- ✅ 151 LOC removed (Phase 3.11 - adapter consolidation)
- ✅ 9 adapter methods removed (72 → 63)

**Remaining Work**:
- 🎯 27 → ~15 calls (44% reduction needed)
- 🎯 Focus: Compound operations, indexed properties, helper methods
- 🎯 13 architectural boundaries will remain (permanent, by design)

---

## Distribution Analysis

###By File (27 calls)

```
visitor_expressions_members.go    6 calls (22.2%)  ███████████
member_assignment.go               6 calls (22.2%)  ███████████
index_assignment.go                3 calls (11.1%)  ██████
assignment_helpers.go              2 calls (7.4%)   ████
visitor_statements.go              2 calls (7.4%)   ████
visitor_expressions_functions.go   2 calls (7.4%)   ████
evaluator.go                       1 call  (3.7%)   ██
helper_methods.go                  1 call  (3.7%)   ██
method_dispatch.go                 1 call  (3.7%)   ██
compound_ops.go                    1 call  (3.7%)   ██
user_function_helpers.go           1 call  (3.7%)   ██
visitor_declarations.go            1 call  (3.7%)   ██
```

### By Category (27 calls)

```
Architectural Boundaries (KEEP)    13 calls (48.1%)  ████████████████████████
Safety Nets (MOSTLY KEEP)          13 calls (48.1%)  ████████████████████████
Migration Candidates                 1 call  (3.7%)   ██
```

### By Migration Status (27 calls)

```
KEEP - Permanent Boundaries        13 calls (48.1%)  ████████████████████████
KEEP - Safety Nets                  9 calls (33.3%)  █████████████████
DEFER - Complex Infrastructure      4 calls (14.8%)  ████████
MIGRATE - Phase 3.12 Candidates     1 call  (3.7%)   ██
```

---

## Detailed Call Site Inventory

### Category 1: Architectural Boundaries (13 calls - KEEP PERMANENTLY)

These represent clean architectural separation between evaluator (AST execution) and interpreter (OOP semantics).

#### 1.1 Self/Class Context (2 calls) - **KEEP**

| File | Line | Context | Rationale |
|------|------|---------|-----------|
| `assignment_helpers.go` | 105 | Simple assignment to Self/`__CurrentClass__` | OOP context owned by interpreter |
| `assignment_helpers.go` | 391 | Compound assignment to Self/`__CurrentClass__` | Same - property setters, class vars |

**Code Example** (assignment_helpers.go:105):
```go
// KEEP EVALNODE: Self and __CurrentClass__ handling
// Assigning to Self or __CurrentClass__ in a method context
// Requires special handling in the interpreter:
// - env.Get("Self") → extract ObjectInstance → check field/property/class var
// - env.Get("__CurrentClass__") → extract ClassInfo → check class var
return e.adapter.EvalNode(stmt)
```

**Why KEEP**: The evaluator executes AST nodes. The interpreter manages OOP semantics (Self context, class variables, property dispatch). This is a clean architectural boundary.

**Migration Cost**: N/A - would require pulling class/property infrastructure into evaluator (architectural violation)

---

#### 1.2 Method Dispatch (6 calls) - **KEEP**

| File | Line | Receiver Type | Operation |
|------|------|---------------|-----------|
| `visitor_expressions_members.go` | 251 | Record | Method invocation |
| `visitor_expressions_members.go` | 314 | Object | Method/unknown member |
| `visitor_expressions_members.go` | 357 | Interface | Virtual method dispatch |
| `visitor_expressions_members.go` | 401 | Class/Metaclass | Constructor/class methods |
| `visitor_expressions_members.go` | 447 | Type Cast | Method on cast value |
| `visitor_expressions_members.go` | 509 | Typed Nil | Class variable lookup |

**Code Example** (visitor_expressions_members.go:357):
```go
// INTERFACE instance
case strings.HasPrefix(objType, "INTERFACE"):
    // Method access or unknown member - delegate to adapter
    // Methods require complex dispatch logic (virtual method tables, etc.)
    return e.adapter.EvalNode(node)
```

**Why KEEP**: Method dispatch requires:
- Virtual Method Table (VMT) lookups for interfaces/classes
- Inheritance chain traversal
- Constructor invocation logic
- Class registry access for typed nil

All of this is interpreter-owned OOP infrastructure.

**Migration Cost**: 20-30 hours (requires full OOP migration to evaluator - Phase 4+ work)

---

#### 1.3 OOP Assignment (4 calls) - **KEEP**

| File | Line | Pattern | Complexity |
|------|------|---------|------------|
| `member_assignment.go` | 55 | Static class member (`TClass.Var := val`) | ClassInfo lookup |
| `member_assignment.go` | 72 | Nil auto-initialization | Object allocation |
| `member_assignment.go` | 82 | Record property setter | Property dispatch |
| `member_assignment.go` | 125 | Class/metaclass member | Class metadata |

**Code Example** (member_assignment.go:72):
```go
// KEEP EVALNODE: Nil value handling (auto-initialization)
if objVal == nil || objVal.Type() == "NIL" {
    // Delegate to adapter for potential auto-initialization
    return e.adapter.EvalNode(stmt)
}
```

**Why KEEP**: DWScript auto-initializes objects on first member assignment. This requires:
- Object allocation logic
- Constructor invocation
- ClassInfo metadata access
- All interpreter-owned infrastructure

**Migration Cost**: 10-15 hours (requires object lifecycle management in evaluator - Phase 4+ work)

---

#### 1.4 External Functions (1 call) - **KEEP**

| File | Line | Context | Rationale |
|------|------|---------|-----------|
| `visitor_expressions_functions.go` | 336 | Go-implemented functions | Var parameter handling, FFI boundary |

**Code Example**:
```go
// Task 3.5.24: External (Go) functions that may need var parameter handling
if e.externalFunctions != nil && e.externalFunctions.Has(funcName.Value) {
    return e.adapter.EvalNode(node)
}
```

**Why KEEP**: External functions:
- Implemented in Go (FFI boundary)
- May have `var` parameters requiring special environment handling
- Need access to interpreter's environment chain

**Migration Cost**: 15-20 hours (requires redesigning external function registry - low priority)

---

### Category 2: Safety Nets (13 calls - MOSTLY KEEP)

These provide graceful degradation and handle incomplete feature migration.

#### 2.1 Fallback Dispatchers (5 calls) - **KEEP**

| File | Line | Fallback For | Status |
|------|------|--------------|--------|
| `evaluator.go` | 674 | Unknown AST node types | Essential safety net |
| `helper_methods.go` | 369 | Unhandled builtin helpers | Migration in progress |
| `method_dispatch.go` | 205 | Unknown object types | Safety net |
| `member_assignment.go` | 90 | Record without RecordFieldSetter | Edge case |
| `member_assignment.go` | 129 | Unknown types in assignment | Safety net |

**Code Example** (evaluator.go:674):
```go
default:
    // Phase 3.5.2: Unknown node type - delegate to adapter if available
    if e.adapter != nil {
        return e.adapter.EvalNode(node)
    }
    panic("Evaluator.Eval: unknown node type and no adapter available")
```

**Why KEEP**: These are graceful fallbacks that:
- Prevent panics on edge cases
- Allow incremental migration (helpers.go:369 will naturally disappear)
- Provide safety during evaluator development

**Migration Potential**:
- `helper_methods.go:369` - Will disappear with complete helper coverage
- Others - Permanent safety nets (should remain)

---

#### 2.2 Compound Operations (3 calls) - **DEFER TO 3.12.3**

| File | Line | Operation | Complexity |
|------|------|-----------|------------|
| `visitor_statements.go` | 318 | Compound member assign (`obj.field += 1`) | Read-modify-write with property setters |
| `visitor_statements.go` | 334 | Compound index assign (`arr[i] *= 2`) | Indexed property dispatch |
| `compound_ops.go` | 26 | Object operator overloads (`obj += val`) | Operator overload resolution |

**Code Example** (visitor_statements.go:318):
```go
case *ast.MemberAccessExpression:
    if isCompound {
        // Compound assignment to member - complex because:
        // 1. Read current value (might be property getter)
        // 2. Apply operation
        // 3. Write new value (might be property setter)
        return e.adapter.EvalNode(node)
    }
```

**Why DEFER**: Compound operations require:
- Property getter/setter dispatch for read-modify-write cycle
- Operator overload resolution (compound_ops.go:26)
- Complex interaction with object lifecycle

**Migration Potential**: ⭐⭐⭐ **HIGH** (Task 3.12.3)
- Estimated effort: 4-6 days
- Clean migration path exists
- Could use existing member_assignment.go patterns

---

#### 2.3 Incomplete Features (4 calls) - **DEFER OR KEEP**

| File | Line | Feature | Status |
|------|------|---------|--------|
| `user_function_helpers.go` | 434 | User function body execution | KEEP - requires 100% evaluator |
| `index_assignment.go` | 45 | Indexed property writes | DEFER - Phase 3.12.3 |
| `index_assignment.go` | 76 | Interface indexed properties | DEFER - Phase 4+ |
| `index_assignment.go` | 82 | Object default indexed props | DEFER - Phase 4+ |

**Code Example** (user_function_helpers.go:434):
```go
// Task 3.5.144b.7: Execute function body via adapter.
// We use EvalNode instead of e.Eval because the evaluator doesn't fully support
// all language features yet (e.g., class constructor calls like TClass.Create).
_ = e.adapter.EvalNode(fn.Body)
```

**Why Various Statuses**:
- Line 434: User functions can contain ANY DWScript feature - requires 100% evaluator (Phase 4+)
- Lines 45, 76, 82: Indexed properties - part of 3.12.3 migration scope

---

#### 2.4 Lazy Evaluation (1 call) - **KEEP**

| File | Line | Context | Rationale |
|------|------|---------|-----------|
| `visitor_expressions_functions.go` | 127 | Lazy parameter thunk callback | Deferred evaluation requires full interpreter |

**Code Example**:
```go
if isLazy {
    // For lazy parameters, reuse existing thunks to avoid self-recursive wrapping
    args[idx] = e.wrapLazyArg(arg, ctx, func(expr ast.Expression) Value {
        return e.adapter.EvalNode(expr)  // Callback for deferred evaluation
    })
}
```

**Why KEEP**: Lazy evaluation (Jensen's Device pattern) requires:
- Callback that can evaluate arbitrary expressions
- Full language support for deferred evaluation
- Legitimate architectural boundary (callback pattern)

**Migration Cost**: N/A - this is correct use of adapter as callback interface

---

### Category 3: Migration Candidates (1 call - MIGRATE IN 3.12)

#### 3.1 Set Declarations (1 call) - **MIGRATE**

| File | Line | Feature | Effort |
|------|------|---------|--------|
| `visitor_declarations.go` | 1186 | Set type declarations | 2-3 hours |

**Code Example**:
```go
// VisitSetDecl evaluates a set declaration.
func (e *Evaluator) VisitSetDecl(node *ast.SetDecl, ctx *ExecutionContext) Value {
    // Set type already registered by semantic analyzer
    // Delegate to adapter for now (Phase 3 migration)
    return e.adapter.EvalNode(node)
}
```

**Why MIGRATE**: This is incomplete Phase 3 migration - set declarations should be handled natively in evaluator.

**Migration Path**:
1. Implement native set type handling in evaluator
2. Use TypeSystem for set type registration
3. Remove adapter delegation

**Estimated Effort**: 2-3 hours (straightforward migration)

---

## Migration Roadmap for Phase 3.12

### Revised Target: 27 → 15 calls

**Reality Check**:
- **13 permanent boundaries** (48.1%) - These SHOULD remain (architectural decision)
- **13 safety nets** (48.1%) - Mostly remain, some defer to Phase 4+
- **1 migration candidate** (3.7%) - Quick win

**Achievable Target**: 27 → ~15-17 calls (removing ~10-12 calls)

### Task 3.12.2: Member Access Migration (1 week)

**Scope**: Migrate helper property and record method dispatch to native evaluator

**Target Calls**:
- Partial: `visitor_expressions_members.go:314` (helper properties only, not object methods)
- Helper methods fallback: `helper_methods.go:369` (will naturally disappear)

**Approach**:
1. Implement native helper property dispatch in evaluator
2. Implement native record method dispatch
3. Keep object/interface/class method calls (architectural boundaries)

**Estimated Impact**: -1 to -2 calls (helper_methods.go fallback may disappear)

**Effort**: 1 week

---

### Task 3.12.3: Assignment Operations Migration (4-6 days)

**Scope**: Migrate compound assignments and indexed properties to native evaluator

**Target Calls**:
- `visitor_statements.go:318` - Compound member assignment
- `visitor_statements.go:334` - Compound index assignment
- `index_assignment.go:45` - Indexed property writes
- Potentially: `compound_ops.go:26` - Object operator overloads (if time permits)

**Approach**:
1. Implement native compound assignment logic in evaluator
2. Reuse existing member_assignment.go and index_assignment.go patterns
3. Add read-modify-write cycle with property getter/setter dispatch

**Estimated Impact**: -3 to -4 calls

**Effort**: 4-6 days

---

### Quick Wins (2-3 days)

**Target Calls**:
- `visitor_declarations.go:1186` - Set declarations (straightforward migration)
- Potentially: `member_assignment.go:90, 129` - Fallback consolidation

**Estimated Impact**: -1 to -3 calls

**Effort**: 2-3 days

---

### Total Phase 3.12 Impact

| Task | Calls Eliminated | Effort | Priority |
|------|-----------------|--------|----------|
| 3.12.2 (Member Access) | 1-2 | 1 week | High |
| 3.12.3 (Assignments) | 3-4 | 4-6 days | High |
| Quick Wins | 1-3 | 2-3 days | Medium |
| **Total** | **5-9 calls** | **~3 weeks** | - |

**Final State**: 27 - 7 = **~20 calls** (realistic target)

**Note**: Original target of 15 calls may not be achievable without violating architectural boundaries. A final count of 15-20 calls (with 13 permanent boundaries documented) is a successful outcome.

---

## Architectural Justification

### Why 13 Calls Should Remain

The 13 architectural boundary calls represent **correct design** for a clean evaluator/interpreter separation:

**Evaluator Responsibilities**:
- Execute AST nodes
- Manage expression/statement evaluation
- Handle values and basic operations
- Maintain execution context

**Interpreter Responsibilities**:
- OOP semantics (classes, inheritance, virtual methods)
- Self context and method dispatch
- Property getters/setters
- Object allocation and lifecycle
- External function interop

**The Boundary**:
- EvalNode serves as a clean callback interface
- Prevents coupling evaluator to class infrastructure
- Allows evaluator to remain AST-focused
- Keeps OOP complexity in interpreter where it belongs

**Alternative Approaches Considered (and rejected)**:

1. **Pull ClassInfo into evaluator** - Rejected: Violates separation of concerns, creates tight coupling
2. **Mirror class infrastructure in evaluator** - Rejected: Code duplication, maintenance burden
3. **Make evaluator own all semantics** - Rejected: God object anti-pattern

**Conclusion**: The 13 architectural boundary calls are **features, not bugs**. They represent a mature, well-factored architecture.

---

## Comparison with Previous Audit

### Changes from 2025-12-04 Audit (34 calls → 27 calls)

**Eliminated (7 calls)**:
1. ✅ 7 reference counting calls - Migrated to `runtime.RefCountManager` (Phase 3.5.39-3.5.42)

**Status Changes (0 calls)**:
- All remaining calls have consistent status

**Net Result**:
- **20.6% reduction** in EvalNode calls
- **All ref counting eliminated** - major architectural win
- **No new calls added** - consolidation work didn't introduce new dependencies

---

## Verification

✅ **All 27 EvalNode calls documented** with:
- File path and line number
- Code snippet and context
- Category (Boundary/Safety/Candidate)
- Migration assessment (Keep/Defer/Migrate)
- Effort estimate where applicable

✅ **Cross-referenced with PLAN.md**:
- Phase 3.12 tasks: 3.12.1 (this audit), 3.12.2-3.12.4 (pending)
- Estimates refined based on actual call analysis

✅ **Grep verification**:
```bash
grep -rn "adapter.EvalNode" internal/interp/evaluator/ | wc -l
# Result: 29 (27 actual calls + 2 in comments)
```

✅ **Architecture validated**:
- 13 permanent boundaries justified
- Clean evaluator/interpreter separation maintained
- No coupling to OOP infrastructure

---

## Next Steps

### Immediate (Task 3.12.1) - ✅ COMPLETE

- ✅ Create this comprehensive audit document
- ✅ Verify call count with grep (27 calls confirmed)
- ✅ Categorize all calls by migration status
- ✅ Provide actionable roadmap for 3.12.2-3.12.4

### Phase 3.12.2 (Member Access) - 📋 READY TO START

- Launch Plan agent to design native helper property dispatch
- Implement evaluator-native record method handling
- Estimated: 1 week, -1 to -2 calls

### Phase 3.12.3 (Assignments) - 📋 PLANNED

- Launch Plan agent to design compound assignment migration
- Implement read-modify-write cycles with property dispatch
- Estimated: 4-6 days, -3 to -4 calls

### Phase 3.12.4 (Documentation) - 📋 PLANNED

- Update PLAN.md with Phase 3.12 completion
- Document final EvalNode call count (~20 calls)
- Create Phase 3 summary document

---

## Conclusion

Phase 3 consolidation has been highly successful:

1. **Measurable Progress**: 34 → 27 calls (20.6% reduction)
2. **Architectural Clarity**: 13 permanent boundaries well-justified
3. **Clear Roadmap**: 5-9 more calls eliminable in Phase 3.12
4. **Mature Design**: Evaluator/interpreter separation is correct and should be preserved

The current migration strategy (Phase 3.12.2-3.12.4) is **well-aligned** with achievable goals and will complete Phase 3 with a clean, well-documented architecture ready for Phase 4 (OOP infrastructure migration).

**Final Target**: ~15-20 calls (with 13 permanent boundaries) represents **successful Phase 3 completion**.
