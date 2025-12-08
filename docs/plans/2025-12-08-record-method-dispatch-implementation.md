# Record Method Dispatch Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate record method dispatch from adapter delegation to native evaluator execution, eliminating 1 EvalNode call.

**Architecture:** Extend RecordInstanceValue interface with GetRecordMethod(), reuse existing CallASTHelperMethod infrastructure for method execution, preserve adapter delegation for method pointers.

**Tech Stack:** Go 1.21+, DWScript AST, existing evaluator/runtime infrastructure

**Design Document:** [2025-12-08-record-method-dispatch-design.md](2025-12-08-record-method-dispatch-design.md)

---

## Task 1: Extend RecordInstanceValue Interface

**Files:**
- Modify: `internal/interp/runtime/values.go` (RecordInstanceValue interface)

**Step 1: Add GetRecordMethod to interface**

Add this method to the `RecordInstanceValue` interface (after `HasRecordMethod`):

```go
// RecordInstanceValue represents a record instance value
type RecordInstanceValue interface {
    Value

    GetRecordTypeName() string
    GetRecordField(name string) (Value, bool)
    SetRecordField(name string, value Value) error
    HasRecordMethod(name string) bool
    HasRecordProperty(name string) bool

    // NEW: Retrieve the AST declaration for a record method.
    // Returns the method declaration and true if found, nil and false otherwise.
    // The name comparison is case-insensitive (DWScript convention).
    GetRecordMethod(name string) (*ast.FunctionDecl, bool)
}
```

**Step 2: Verify compilation fails**

Run: `go build ./internal/interp/runtime`

Expected: Build fails with "RecordValue does not implement RecordInstanceValue (missing method GetRecordMethod)"

**Step 3: Commit interface change**

```bash
git add internal/interp/runtime/values.go
git commit -m "feat(task-3.12.2): add GetRecordMethod to RecordInstanceValue interface"
```

---

## Task 2: Implement GetRecordMethod on RecordValue

**Files:**
- Modify: `internal/interp/interpreter.go` (RecordValue type)

**Step 1: Find RecordValue type definition**

Search for `type RecordValue struct` in `internal/interp/interpreter.go`.

The struct already has a `methods map[string]*ast.FunctionDecl` field populated during record type declaration.

**Step 2: Add GetRecordMethod implementation**

Add this method after the existing RecordValue methods:

```go
// GetRecordMethod retrieves the AST method declaration for a record method.
// Task 3.12.2: Enables evaluator to execute record methods natively.
func (r *RecordValue) GetRecordMethod(name string) (*ast.FunctionDecl, bool) {
    // Use case-insensitive lookup (DWScript convention)
    method, found := r.methods[ident.Normalize(name)]
    return method, found
}
```

**Step 3: Verify compilation succeeds**

Run: `go build ./internal/interp`

Expected: Build succeeds (RecordValue now implements RecordInstanceValue)

**Step 4: Commit implementation**

```bash
git add internal/interp/interpreter.go
git commit -m "feat(task-3.12.2): implement GetRecordMethod on RecordValue"
```

---

## Task 3: Create Record Method Executor

**Files:**
- Create: `internal/interp/evaluator/record_methods.go`

**Step 1: Create new file with package and imports**

```go
package evaluator

import (
    "github.com/cwbudde/go-dws/pkg/ast"
)
```

**Step 2: Implement callRecordMethod**

Add the complete method executor:

```go
// callRecordMethod executes a record method in the evaluator.
//
// Record methods are user-defined methods attached to record types.
// They execute with Self bound to the record instance.
//
// Task 3.12.2: Migrates record method execution from interpreter to evaluator.
//
// Execution semantics:
// - Create new environment (child of current)
// - Bind Self to record instance
// - Bind method parameters from args
// - Initialize Result variable (if method has return type)
// - Execute method body
// - Extract and return Result
func (e *Evaluator) callRecordMethod(
    record RecordInstanceValue,
    method *ast.FunctionDecl,
    args []Value,
    node ast.Node,
    ctx *ExecutionContext,
) Value {
    // 1. Validate parameter count
    if len(args) != len(method.Parameters) {
        return e.newError(node,
            "method '%s' expects %d parameters, got %d",
            method.Name.Value, len(method.Parameters), len(args))
    }

    // 2. Create method environment (child of current context)
    methodEnv := NewEnvironment(ctx.Env())

    // 3. Bind Self to record instance
    // This allows the method body to access record fields via Self.FieldName
    methodEnv.Set("Self", record)

    // 4. Bind method parameters
    for i, param := range method.Parameters {
        paramName := param.Name.Value
        methodEnv.Set(paramName, args[i])
    }

    // 5. Initialize Result variable
    // DWScript uses implicit Result variable for function return values
    if method.ReturnType != nil {
        // Initialize Result to zero value of return type
        zeroVal := e.evaluateTypeZeroValue(method.ReturnType)
        methodEnv.Set("Result", zeroVal)
    }

    // 6. Execute method body in new environment
    // Save current environment and restore after execution
    savedEnv := ctx.Env()
    ctx.SetEnv(methodEnv)
    defer ctx.SetEnv(savedEnv)

    // Execute the method body
    _ = e.Eval(method.Body, ctx)

    // 7. Check for early return (Exit/Break statement)
    if ctx.HasReturn() {
        returnVal := ctx.GetReturnValue()
        ctx.ClearReturn()  // Clear return flag for caller
        return returnVal
    }

    if ctx.HasExit() {
        // Exit propagates up the call stack
        return e.runtime.NewNilValue()
    }

    // 8. Extract Result variable (if method has return type)
    if method.ReturnType != nil {
        if result, found := methodEnv.Get("Result"); found {
            return result
        }
        // Result not set - return zero value
        return e.evaluateTypeZeroValue(method.ReturnType)
    }

    // Procedure (no return type) - return nil
    return e.runtime.NewNilValue()
}
```

**Step 3: Verify compilation succeeds**

Run: `go build ./internal/interp/evaluator`

Expected: Build succeeds

**Step 4: Commit new file**

```bash
git add internal/interp/evaluator/record_methods.go
git commit -m "feat(task-3.12.2): add callRecordMethod executor"
```

---

## Task 4: Update Member Access Dispatch

**Files:**
- Modify: `internal/interp/evaluator/visitor_expressions_members.go:249-252`

**Step 1: Read current implementation**

Current code at line 249-252:

```go
// Method reference - still uses adapter for method invocation
if recVal.HasRecordMethod(memberName) {
    return e.adapter.EvalNode(node)
}
```

**Step 2: Replace with native dispatch**

Replace those 4 lines with:

```go
// Method reference - now handled natively in evaluator
// Task 3.12.2: Migrate from adapter delegation to native execution
if recVal.HasRecordMethod(memberName) {
    methodDecl, found := recVal.GetRecordMethod(memberName)
    if !found {
        // Should never happen - HasRecordMethod returned true
        return e.newError(node, "internal error: method '%s' not retrievable", memberName)
    }

    // Distinguish method call vs method reference
    // Case 1: record.Method() - execute the method
    if callExpr, isCall := node.(*ast.CallExpression); isCall {
        // Evaluate arguments
        args := make([]Value, len(callExpr.Arguments))
        for i, arg := range callExpr.Arguments {
            args[i] = e.Eval(arg, ctx)
            if e.isError(args[i]) {
                return args[i]  // Propagate error
            }
        }

        // Execute record method natively
        return e.callRecordMethod(recVal, methodDecl, args, node, ctx)
    }

    // Case 2: var f := record.Method - get bound method pointer
    // This requires OOP binding infrastructure (Self capture)
    // Keep using adapter for this rare case
    return e.adapter.CreateBoundMethodPointer(obj, memberName, methodDecl)
}
```

**Step 3: Verify compilation succeeds**

Run: `go build ./internal/interp/evaluator`

Expected: Build succeeds

**Step 4: Commit member access changes**

```bash
git add internal/interp/evaluator/visitor_expressions_members.go
git commit -m "feat(task-3.12.2): migrate record method calls to native evaluator"
```

---

## Task 5: Run Unit Tests

**Files:**
- Test: All existing evaluator and interpreter tests

**Step 1: Run evaluator tests**

Run: `go test ./internal/interp/evaluator -v`

Expected: All tests pass (no regressions)

**Step 2: Run interpreter tests**

Run: `go test ./internal/interp -v`

Expected: All tests pass (no regressions)

**Step 3: Run full test suite**

Run: `go test ./...`

Expected: All tests pass

**If tests fail:**
- Review failure messages
- Check if failures are in record-related tests
- Debug and fix issues
- Re-run tests until all pass

---

## Task 6: Verify EvalNode Count Reduction

**Files:**
- Verify: `internal/interp/evaluator/*.go`

**Step 1: Count EvalNode calls**

Run: `grep -rn "adapter.EvalNode" internal/interp/evaluator/ | wc -l`

Expected: **26** (down from 27)

**Step 2: Verify line 251 eliminated**

Run: `grep -n "adapter.EvalNode" internal/interp/evaluator/visitor_expressions_members.go`

Expected: Line 251 should NOT appear in results (replaced with native call)

**Step 3: Document verification**

Create verification note:

```bash
echo "EvalNode count verification (Task 3.12.2):
Before: 27 calls
After: 26 calls
Eliminated: visitor_expressions_members.go:251 (record method calls)
Remaining: 26 calls (method pointers still use CreateBoundMethodPointer)
" > /tmp/evalnode-verification.txt
cat /tmp/evalnode-verification.txt
```

---

## Task 7: Run Fixture Tests

**Files:**
- Test: `testdata/fixtures/`

**Step 1: Run all fixture tests**

Run: `go test -v ./internal/interp -run TestDWScriptFixtures`

Expected: Same pass/fail count as before (no new failures)

**Step 2: Run record-specific fixtures** (if they exist)

Run: `go test -v ./internal/interp -run TestDWScriptFixtures/RecordMethods`

Expected: All record method tests pass

**Step 3: Check for timeout/crash issues**

If any tests timeout or crash:
- These are pre-existing issues (documented in fixture status)
- Do NOT block this task on pre-existing failures
- Verify no NEW failures introduced

---

## Task 8: Update PLAN.md

**Files:**
- Modify: `PLAN.md` (Task 3.12.2 status)

**Step 1: Mark task 3.12.2 complete**

Find the task 3.12.2 section (around line 368) and update:

```markdown
- [x] **3.12.2** Migrate Member Access (ACTUAL: 3h) - ✅ COMPLETE (2025-12-08)
  - Native handling for record method calls
  - Kept adapter for method pointers (architectural boundary)
  - **Impact**: -1 EvalNode call (27 → 26)
  - **Files modified**:
    - runtime/values.go (+1 interface method)
    - interpreter.go (+5 lines - GetRecordMethod impl)
    - evaluator/record_methods.go (+80 lines - new file)
    - evaluator/visitor_expressions_members.go (+25 lines - native dispatch)
```

**Step 2: Commit PLAN.md update**

```bash
git add PLAN.md
git commit -m "docs(task-3.12.2): mark task complete in PLAN.md"
```

---

## Task 9: Create Completion Summary

**Files:**
- Create: `docs/task-3.12.2-completion-summary.md`

**Step 1: Write completion summary**

```markdown
# Task 3.12.2 Completion Summary

**Date**: 2025-12-08
**Task**: Migrate Member Access (Record Method Dispatch)
**Status**: ✅ Complete

## Achievements

- **EvalNode calls**: 27 → 26 (-1 call, 3.7% reduction)
- **Files modified**: 4 files
- **Code added**: ~110 lines (80 in new file, 30 in modifications)
- **Tests**: All passing (0 regressions)

## Changes Made

### 1. Interface Extension
- Added `GetRecordMethod(name string) (*ast.FunctionDecl, bool)` to `RecordInstanceValue`
- File: `internal/interp/runtime/values.go`

### 2. Concrete Implementation
- Implemented `GetRecordMethod` on `RecordValue` (3-line method)
- File: `internal/interp/interpreter.go`

### 3. Method Executor
- Created `callRecordMethod()` for native record method execution
- Reuses environment management, parameter binding, Result variable patterns
- File: `internal/interp/evaluator/record_methods.go` (NEW)

### 4. Dispatch Update
- Replaced adapter delegation with native execution for method calls
- Preserved adapter for method pointers (architectural boundary)
- File: `internal/interp/evaluator/visitor_expressions_members.go`

## Architectural Impact

**Eliminated**:
- Line 251 (`visitor_expressions_members.go`) - record method calls via EvalNode

**Preserved**:
- Method pointers still use `adapter.CreateBoundMethodPointer` (rare case, OOP infrastructure)

## Test Results

- ✅ Unit tests: All passing
- ✅ Integration tests: All passing
- ✅ Fixture tests: No new failures
- ✅ Compilation: Clean build

## Next Steps

- Task 3.12.3: Migrate compound assignment operations
- Expected impact: -3 to -4 additional EvalNode calls
- Target: 26 → ~22 calls
```

**Step 2: Commit completion summary**

```bash
git add docs/task-3.12.2-completion-summary.md
git commit -m "docs(task-3.12.2): add completion summary"
```

---

## Success Criteria

- ✅ RecordInstanceValue interface extended with GetRecordMethod
- ✅ RecordValue implements GetRecordMethod (case-insensitive lookup)
- ✅ callRecordMethod executes record methods natively in evaluator
- ✅ visitor_expressions_members.go uses native dispatch for method calls
- ✅ Method pointers preserved via adapter (CreateBoundMethodPointer)
- ✅ All unit tests pass (0 regressions)
- ✅ All fixture tests pass (no new failures)
- ✅ EvalNode count reduced: 27 → 26
- ✅ PLAN.md updated with completion status
- ✅ Completion summary documented

---

## Estimated Effort

**Total**: 2-3 hours

- Task 1: Interface extension (15 min)
- Task 2: Implementation (15 min)
- Task 3: Method executor (30 min)
- Task 4: Dispatch update (30 min)
- Task 5: Unit tests (20 min)
- Task 6: Verification (10 min)
- Task 7: Fixture tests (20 min)
- Task 8: PLAN.md update (5 min)
- Task 9: Summary (15 min)

---

## Rollback Plan

If issues discovered:

```bash
# Revert all commits from this task
git log --oneline | head -n 7  # Identify commit range
git revert <first-commit>..<last-commit>

# Or reset to before task started
git reset --hard <commit-before-task>
```

Affected files restore automatically via git revert.
