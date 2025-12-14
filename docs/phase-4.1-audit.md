# Phase 4.1 Audit Report

**Date**: 2025-12-14
**Status**: Empirical testing completed, 4/34 callbacks eliminated
**Branch**: feat/attempt_on_phase_4

## Executive Summary

Phase 4.1 aims to eliminate the evaluator→interpreter callback pattern by removing focused interface dependencies. Through empirical testing with `selfContainedMode=true`, we identified exactly which callbacks remain and their blockers.

**Key Findings**:
- ✅ 4 control flow callbacks eliminated (Break/Continue/Exit/Return)
- ✅ All tests pass after migration
- ✅ `contracts` package validated as legitimate dependency inversion (not theatre)
- ✅ 30 remaining callbacks categorized by blocking dependency
- ✅ Clear path forward with prioritized quick wins

## Baseline Metrics

| Metric | Count |
|--------|-------|
| Total AST node types | 60 |
| Evaluator.Eval() cases with callbacks | 34 (before), 30 (after control flow) |
| Interpreter.evalLegacy() cases | 59 (before), 55 (after control flow) |
| Focused interface methods | 68 total (OOPEngine: 20, DeclHandler: 38, ExceptionManager: 6, CoreEvaluator: 4) |

## Quick Win #1: Control Flow Statements ✅

**Date**: 2025-12-14
**Effort**: 15 minutes
**Result**: SUCCESS

### What Was Done

Removed callback checks for 4 control flow statements:
- `BreakStatement` ([evaluator.go:612-613](internal/interp/evaluator/evaluator.go#L612-L613))
- `ContinueStatement` ([evaluator.go:617-618](internal/interp/evaluator/evaluator.go#L617-L618))
- `ExitStatement` ([evaluator.go:622-623](internal/interp/evaluator/evaluator.go#L622-L623))
- `ReturnStatement` ([evaluator.go:627-628](internal/interp/evaluator/evaluator.go#L627-L628))

Deleted corresponding cases from [interpreter.go:258-272](internal/interp/interpreter.go#L258-L272).

### Why It Worked

These `Visit*` methods only depend on `ExecutionContext.ControlFlow()`:
```go
func (e *Evaluator) VisitBreakStatement(node *ast.BreakStatement, ctx *ExecutionContext) Value {
    ctx.ControlFlow().SetBreak()
    return &runtime.NilValue{}
}
```

No dependencies on:
- Interpreter state
- Focused interfaces (OOPEngine, DeclHandler, ExceptionManager)
- coreEvaluator callbacks

### Impact

- Callbacks: 34 → 30 (-4)
- evalLegacy cases: 59 → 55 (-4)
- Lines deleted: ~20 across 2 files
- Tests: ✅ All pass

## SelfContainedMode Experiment

**Method**: Disabled `SetFocusedInterfaces()` to simulate evaluator running without interpreter callbacks.

**Files modified**:
- [runner/runner.go:55](internal/interp/runner/runner.go#L55) - commented out `eval.SetFocusedInterfaces()`
- [new_test.go:57](internal/interp/new_test.go#L57) - commented out `eval.SetFocusedInterfaces()`

### Results

Two distinct failure patterns emerged:

#### Pattern 1: builtins.Context Type Assertions (7 cases)

**Error**: `"adapter does not implement builtins.Context"`

**Root cause**: Methods trying to cast `coreEvaluator` to `builtins.Context`:

```go
// context_bounds.go:35
ctx, ok := e.coreEvaluator.(builtins.Context)
if !ok {
    return e.newError(node, "GetLowBound: adapter does not implement builtins.Context")
}
return ctx.GetLowBound(arrVal, node)  // Should just call e.GetLowBound()
```

**Files**:
- [context_bounds.go](internal/interp/evaluator/context_bounds.go): `GetLowBound`, `GetHighBound` (2 calls)
- [context_enums.go](internal/interp/evaluator/context_enums.go): `GetEnumLowValue`, `GetEnumHighValue`, `GetEnumElementCount`, `GetEnumElementName`, `GetEnumElementValue` (5 calls)

**Fix**: Replace `e.coreEvaluator.(builtins.Context).Method()` → `e.Method()`

**Effort**: 30 minutes

#### Pattern 2: Nil Pointer Dereferences (23 cases)

**Error**: `panic: runtime error: invalid memory address or nil pointer dereference`

**Root cause**: Visit* methods accessing nil focused interface fields when `SetFocusedInterfaces()` not called.

**Example stack trace**:
```
VisitClassDecl → e.declHandler.NewClassInfoAdapter(className) → nil pointer panic
```

**Breakdown by interface**:
- `declHandler` (7 cases): ClassDecl, InterfaceDecl, HelperDecl, FunctionDecl, OperatorDecl, EnumDecl, TypeDeclaration
- `oopEngine` (10 cases): CallExpression, NewExpression, MemberAccessExpression, MethodCallExpression, InheritedExpression, SelfExpression, IndexExpression, NewArrayExpression, AsExpression, AddressOfExpression
- `exceptionMgr` (3 cases): TryStatement, RaiseStatement, Program
- Mixed (3 cases): Statements requiring multiple interfaces

## Remaining Work Breakdown

### Category 1: builtins.Context Casts ⚡ QUICK WIN

**Cases**: 7
**Blocking**: Type assertion to coreEvaluator (implementation bug)
**Effort**: 30 minutes
**Priority**: HIGH - trivial fix, eliminates 7 callbacks

**Files to edit**:
- `internal/interp/evaluator/context_bounds.go` (2 methods)
- `internal/interp/evaluator/context_enums.go` (5 methods)

**Change pattern**:
```go
// Before
ctx, ok := e.coreEvaluator.(builtins.Context)
if !ok {
    return e.newError(node, "Method: adapter does not implement builtins.Context")
}
return ctx.SomeMethod(args)

// After
return e.SomeMethod(args)  // Evaluator already implements builtins.Context
```

### Category 2: Loop Statements 🟢 MEDIUM

**Cases**: 5 (IfStatement, WhileStatement, RepeatStatement, ForStatement, ForInStatement)
**Blocking**: Exception state (likely already in ExecutionContext)
**Effort**: 2 hours
**Priority**: MEDIUM - high confidence they work

**Approach**:
1. Remove `if e.coreEvaluator != nil && !e.selfContainedMode` check
2. Run tests
3. If failure, check if exception state is available in `ctx`

### Category 3: Simple Statements 🟡 MIXED

**Cases**: 4 (ExpressionStatement, VarDeclStatement, ConstDecl, CaseStatement)
**Blocking**: Mixed - some may need 4.3/4.4/4.5 first
**Effort**: 4 hours
**Priority**: MEDIUM - defer hard cases

### Category 4: Declarations 🔴 BLOCKED

**Cases**: 7
**Blocking**: declHandler interface (~38 methods)
**Effort**: 2-3 days
**Priority**: LOW - requires Task 4.4

**Cases**: ClassDecl, InterfaceDecl, HelperDecl, FunctionDecl, OperatorDecl, EnumDecl, TypeDeclaration

**Strategy**: Move declaration logic to evaluator or inline into Visit* methods (Task 4.4)

### Category 5: OOP Operations 🔴 BLOCKED

**Cases**: 10
**Blocking**: oopEngine interface (~20 methods)
**Effort**: 2-3 days
**Priority**: LOW - requires Task 4.3

**Cases**: CallExpression, NewExpression, MemberAccessExpression, MethodCallExpression, InheritedExpression, SelfExpression, IndexExpression, NewArrayExpression, AsExpression, AddressOfExpression

**Strategy**: Move OOP logic to evaluator (Task 4.3)

### Category 6: Exception Handling 🔴 BLOCKED

**Cases**: 3 (TryStatement, RaiseStatement, Program)
**Blocking**: exceptionMgr interface (~6 methods)
**Effort**: 0.5 day
**Priority**: MEDIUM - requires Task 4.5

## Validation: The contracts Package

Initial assessment questioned whether `internal/interp/contracts` was "decomposition theatre". After deeper analysis, it serves a **legitimate purpose**:

### Why It Exists

```
internal/interp/         imports→  contracts ←imports  internal/interp/evaluator/
(Interpreter)                      (neutral)                    (Evaluator)
```

**Purpose**: Break import cycle between `interp` and `interp/evaluator`

**What it provides**:
1. `Evaluator` interface - dependency inversion so interpreter depends on evaluator contract
2. Focused interfaces - document what callbacks remain to be eliminated
3. Shared types - `FunctionPointerMetadata`, `UserFunctionCallbacks`, etc.

**Verdict**: ✅ Legitimate architectural pattern enabling migration, NOT theatre

## Recommended Execution Order

### Phase 1: Quick Wins (3 hours total)

1. **Task 4.1.3**: Fix builtins.Context casts (30 min) ⚡
   - Expected: 7 callbacks eliminated
   - Risk: VERY LOW

2. **Task 4.1.4**: Migrate loop statements (2 hours) 🟢
   - Expected: 5 callbacks eliminated
   - Risk: LOW

3. **Task 4.1.5**: Migrate simple statements (30 min - 4 hours) 🟡
   - Expected: 2-4 callbacks eliminated
   - Risk: MEDIUM - defer hard cases

**After Phase 1**: ~16 callbacks eliminated, ~14-18 remaining

### Phase 2: Interface Elimination (5-7 days)

4. **Task 4.5**: ExceptionManager (0.5 day)
   - 6 methods, 3 callbacks
   - Straightforward - exception creation/raising

5. **Task 4.4**: DeclHandler (2-3 days)
   - 38 methods, 7 callbacks
   - Largest blocker - move to evaluator/inline

6. **Task 4.3**: OOPEngine (2-3 days)
   - 20 methods, 10 callbacks
   - Second largest - move method dispatch to evaluator

**After Phase 2**: All 30 callbacks eliminated ✅

## Progress Tracking

### Callbacks Eliminated

- [x] BreakStatement
- [x] ContinueStatement
- [x] ExitStatement
- [x] ReturnStatement
- [ ] 7 builtins.Context casts
- [ ] 5 loop statements
- [ ] 2-4 simple statements
- [ ] 3 exception handling (blocked by Task 4.5)
- [ ] 7 declarations (blocked by Task 4.4)
- [ ] 10 OOP operations (blocked by Task 4.3)

**Total**: 4/34 complete (12%), 16/34 within reach (47%)

## Lessons Learned

1. **Empirical testing > speculation** - selfContainedMode experiment gave exact blockers
2. **Infrastructure already in place** - 22/60 cases fully migrated, 34/60 have fallback ready
3. **contracts package useful** - proper dependency inversion, not theatre
4. **Quick wins exist** - 7 callbacks are just a type assertion bug
5. **Clear path forward** - prioritized list with effort estimates

## Next Steps

1. Fix builtins.Context casts (Task 4.1.3) - **30 min, 7 callbacks**
2. Migrate loop statements (Task 4.1.4) - **2 hours, 5 callbacks**
3. Migrate simple statements (Task 4.1.5) - **4 hours, 2-4 callbacks**
4. Address remaining blockers via Tasks 4.3, 4.4, 4.5 - **5-7 days, 20 callbacks**

**Estimated time to complete Phase 4.1**: 1-2 weeks focused work
