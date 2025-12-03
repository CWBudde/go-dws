# Phase 3.5: Complete Evaluator Independence

## Overview

**Goal**: Eliminate adapter completely by migrating reference counting to runtime package.

**Status**: Tasks 3.5.1-3.5.36 complete (infrastructure + 50% adapter reduction), Tasks 3.5.37-3.5.45 remaining (ref counting migration + finalization).

---

## Completed Work (3.5.1-3.5.36) ✅

**Infrastructure Created**:
- Evaluator with 48+ visitor methods (expressions, statements, declarations)
- TypeSystem with registries (classes, records, interfaces, functions, helpers, operators)
- ExecutionContext with call stack, control flow, environment
- Property read/write via ObjectValue callbacks

**Adapter Reduction Achieved**:
- Methods: 75 → 46 (~39% reduction)
- EvalNode calls: ~60 → ~30 (~50% reduction)
- Removed: 29 unused/single-use adapter methods
- All tests passing, no regressions

**Key Accomplishments**:
1. Core Infrastructure (3.5.1-3.5.6)
2. Declaration Migration (3.5.7-3.5.16) - all 10 types
3. Value Types to Runtime (3.5.17-3.5.23) - ObjectInstance, InterfaceInstance, etc.
4. Zero-Caller Removal (3.5.25-3.5.27) - 21 methods
5. Single-Use Inlining (3.5.28-3.5.31) - 8 methods
6. Property Callbacks (3.5.32-3.5.34) - native read/write
7. Assignment Optimization (3.5.35-3.5.36) - member/identifier assignments

---

## Remaining Work (3.5.37-3.5.45)

### Current Blocker

**8 essential EvalNode calls in `assignment_helpers.go`**:
- 6 calls: Reference counting (object/interface lifecycle)
- 2 calls: Self/class context (environment ownership)

**Decision**: Migrate ref counting to runtime to achieve full evaluator independence.

---

### Phase A: Preliminary Cleanup (3.5.37-3.5.38)

- [ ] **3.5.37** Migrate Complex Member Access
  **Goal**: Reduce `visitor_expressions_members.go` EvalNode calls 7 → 2-3
  **Work**: Native handling for nested property, interface property, helper property
  **Effort**: 2-3 hours

- [ ] **3.5.38** Audit All Remaining EvalNode Calls
  **Goal**: Comprehensive inventory of all EvalNode calls across evaluator
  **Deliverable**: `docs/evalnode-audit-final.md` with categorization (ref counting, Self context, legitimate delegation)
  **Effort**: 2 hours

---

### Phase B: Reference Counting Migration (3.5.39-3.5.42)

- [ ] **3.5.39** Design Ref Counting Architecture
  **Goal**: Design how ref counting works in runtime without circular imports
  **Deliverable**: `docs/refcounting-design.md`
  **Subtasks**:
  - 3.5.39a: Audit all ref counting operations in interpreter (inc/dec/destructor patterns)
  - 3.5.39b: Design `runtime.RefCountManager` interface
  - 3.5.39c: Design destructor callback pattern (avoid importing `interp.Destructor`)
  - 3.5.39d: Plan migration of 6 EvalNode calls in assignment_helpers.go
  - 3.5.39e: Identify all other ref counting call sites (method pointers, returns, etc.)

  **Effort**: 8-10 hours

- [ ] **3.5.40** Implement Ref Counting in Runtime
  **Goal**: Move ref counting logic to runtime package
  **Work**: Create `runtime/refcount.go` with `RefCountManager`
  **Subtasks**:
  - 3.5.40a: Create `RefCountManager` interface (IncrementRef, DecrementRef, ReleaseObject, ReleaseInterface)
  - 3.5.40b: Move `callDestructorIfNeeded` logic to runtime (use callback pattern)
  - 3.5.40c: Implement ref counting for ObjectInstance and InterfaceInstance
  - 3.5.40d: Add destructor registration mechanism
  - 3.5.40e: Create ref count unit tests (100+ edge cases: circular refs, exceptions, nil, etc.)

  **Effort**: 12-16 hours

- [ ] **3.5.41** Migrate Assignment Ref Counting
  **Goal**: Replace 6 EvalNode calls in `assignment_helpers.go` with `RefCountManager`
  **Subtasks**:
  - 3.5.41a: Interface variable assignment (lines 123, 170) - wrap, inc ref
  - 3.5.41b: Object variable assignment (lines 133, 160) - dec old, inc new, destructor
  - 3.5.41c: Method pointer assignment (line 180) - inc ref on SelfObject
  - 3.5.41d: Var parameter assignments (lines 199, 222) - ref counting through references

  **Effort**: 8-12 hours

- [ ] **3.5.42** Migrate All Other Ref Counting Sites
  **Goal**: Replace ref counting in method returns, property access, array operations
  **Work**: Use `RefCountManager` everywhere objects/interfaces are assigned/returned
  **Effort**: 4-6 hours

---

### Phase C: Self Context Migration (3.5.43)

- [ ] **3.5.43** Migrate Self/Class Context Handling
  **Goal**: Eliminate 2 EvalNode calls for Self/class variable access (lines 94, 283)
  **Options**:
  - Option 1: Enhance `GetVar`/`SetVar` to check Self/class scope (callback to interpreter)
  - Option 2: Pass Self/class context explicitly to evaluator via ExecutionContext
  - Option 3: Keep these 2 calls as essential (Self management stays in interpreter)

  **Recommendation**: Evaluate during implementation (Option 3 may be acceptable)
  **Effort**: 4-6 hours (or 0 if keeping essential)

---

### Phase D: Final Adapter Elimination (3.5.44-3.5.45)

- [ ] **3.5.44** Remove Adapter Interface
  **Goal**: Delete `InterpreterAdapter` interface entirely
  **Work**:
  - Remove all `e.adapter.*` calls from evaluator
  - Delete `adapter_*.go` files
  - Update evaluator to be fully self-contained
  - Verify all tests still pass

  **Effort**: 4-6 hours

- [ ] **3.5.45** Final Testing & Documentation
  **Goal**: Verify complete migration and document final architecture
  **Testing**:
  - Run full test suite (all interpreter tests)
  - Run all fixture tests (especially class_*, oop_*, destroy, free_destroy, clear_ref_*)
  - Memory leak detection (verify ref counting works correctly)
  - Edge cases: circular refs, exception unwinding, destructor order

  **Documentation**:
  - `docs/phase3.5-summary.md` - Complete phase summary
  - Update CLAUDE.md - Final evaluator/interpreter architecture
  - Update `README.md` - Reflect new architecture

  **Effort**: 6-8 hours

---

## Summary

| Phase | Tasks | Description | Effort | Status |
|-------|-------|-------------|--------|--------|
| A | 3.5.37-3.5.38 | Cleanup & Audit | 4-5h | Pending |
| B | 3.5.39-3.5.42 | Ref Counting Migration | 32-44h | Pending |
| C | 3.5.43 | Self Context | 0-6h | Pending |
| D | 3.5.44-3.5.45 | Adapter Removal & Testing | 10-14h | Pending |

**Total Remaining Effort**: 46-69 hours (depending on Self context approach)

**Final State**:
- ✅ Evaluator is fully self-contained
- ✅ No adapter interface
- ✅ Reference counting in runtime package
- ✅ All tests passing
- ✅ Clean architecture with clear separation of concerns

---

## Migration Risk Mitigation

**High-Risk Areas**:
1. Ref counting semantics (when to inc/dec, destructor timing)
2. Circular reference handling
3. Exception unwinding with ref counting
4. Method pointer lifecycle

**Mitigation Strategies**:
1. Comprehensive test coverage before migration
2. Incremental migration (one call site at a time)
3. Extensive testing after each migration step
4. Keep backup of interpreter ref counting logic for reference
5. Memory leak detection tests
