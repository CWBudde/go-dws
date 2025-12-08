# Record Method Dispatch Migration Design

**Task**: 3.12.2 - Migrate Member Access
**Date**: 2025-12-08
**Status**: Approved for Implementation
**Estimated Effort**: 2-3 hours

---

## Executive Summary

Migrate record method dispatch from adapter delegation to native evaluator execution. This eliminates 1 EvalNode call (line 251 in `visitor_expressions_members.go`) by reusing existing method execution infrastructure.

**Key Insight**: Record methods are conceptually identical to helper methods - both execute user-defined AST method bodies with `Self` bound to a value. We can reuse the proven `CallASTHelperMethod` infrastructure.

**Impact**:
- **EvalNode calls**: 27 → 26 (-1, potentially -2 if helper fallback disappears)
- **Code added**: ~60 lines (new `callRecordMethod` helper)
- **Interface changes**: 1 new method on `RecordInstanceValue`
- **Risk**: LOW (reuses proven patterns, minimal changes)

---

## Problem Statement

### Current Behavior

When the evaluator encounters `recordInstance.MethodName()`:

1. ✅ Detects method via `recVal.HasRecordMethod(memberName)` (works)
2. ❌ Delegates to adapter via `e.adapter.EvalNode(node)` (needs migration)
3. Adapter executes in interpreter with full OOP infrastructure

**The Issue**: This adapter dependency is unnecessary - the evaluator already has all the infrastructure needed to execute methods.

### Code Location

```go
// internal/interp/evaluator/visitor_expressions_members.go:249-252
if recVal.HasRecordMethod(memberName) {
    return e.adapter.EvalNode(node)  // ← ELIMINATE THIS
}
```

---

## Design Decision: Approach 1 (Interface Extension)

### Why This Approach

Three options were considered:
1. **Extend RecordInstanceValue interface** ← CHOSEN
2. Create RecordMethodDispatcher (more complex, duplication)
3. Narrow adapter method (doesn't reduce EvalNode count)

**Rationale for Approach 1**:
- Reuses proven `CallASTHelperMethod` infrastructure
- Minimal new code (~60 lines)
- Natural interface extension (`HasRecordMethod` → `GetRecordMethod`)
- Low risk - only getter added, execution logic already tested
- Aligns with Phase 3.12 goal (eliminate EvalNode calls)

---

## Implementation Design

### 1. Interface Extension

**File**: `internal/interp/runtime/values.go`

Add one method to `RecordInstanceValue` interface:

```go
// RecordInstanceValue represents a record instance value
type RecordInstanceValue interface {
    Value

    // Existing methods (unchanged)
    GetRecordTypeName() string
    GetRecordField(name string) (Value, bool)
    SetRecordField(name string, value Value) error
    HasRecordMethod(name string) bool
    HasRecordProperty(name string) bool

    // NEW: Retrieve the AST declaration for a record method
    // Returns the method declaration and true if found, nil and false otherwise.
    // The name comparison is case-insensitive (DWScript convention).
    GetRecordMethod(name string) (*ast.FunctionDecl, bool)
}
```

**Rationale**: Mirrors existing pattern - `Has*` query followed by `Get*` accessor. Keeps runtime interface clean and focused.

---

### 2. Concrete Implementation

**File**: `internal/interp/interpreter.go` (RecordValue type)

The `RecordValue` struct already maintains a `methods map[string]*ast.FunctionDecl` populated during record type declaration. Implementation is trivial:

```go
// GetRecordMethod retrieves the AST method declaration for a record method.
// Task 3.12.2: Enables evaluator to execute record methods natively.
func (r *RecordValue) GetRecordMethod(name string) (*ast.FunctionDecl, bool) {
    // Use case-insensitive lookup (DWScript convention)
    method, found := r.methods[ident.Normalize(name)]
    return method, found
}
```

**Notes**:
- Uses `ident.Normalize` for case-insensitive comparison (DWScript standard)
- Returns `(*ast.FunctionDecl, bool)` - standard Go pattern for optional values
- Zero allocations - just map lookup

---

### 3. Evaluator Integration

**File**: `internal/interp/evaluator/visitor_expressions_members.go`

Replace adapter delegation (lines 249-252) with native dispatch:

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

**Key Decisions**:

1. **Method Call**: Execute natively in evaluator (common case)
2. **Method Reference**: Delegate to adapter for bound method pointers (rare case)

**Why keep adapter for method references?**
- Method pointers require Self-binding infrastructure (closure creation)
- This is OOP-specific infrastructure that belongs in interpreter
- Method references are rare compared to direct calls
- Keeps the migration focused and low-risk

---

### 4. Record Method Execution

**File**: `internal/interp/evaluator/record_methods.go` (NEW)

Create dedicated record method executor:

```go
package evaluator

import (
    "github.com/cwbudde/go-dws/pkg/ast"
)

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

        // Handle var parameters (rare but supported)
        if param.IsVar {
            // Var parameters need special handling (reference semantics)
            // For now, treat as regular parameters
            // TODO: Phase 4 - implement proper var parameter semantics
        }
    }

    // 5. Initialize Result variable
    // DWScript uses implicit Result variable for function return values
    if method.ReturnType != nil {
        // Initialize Result to zero value of return type
        zeroVal := e.getZeroValueForType(method.ReturnType)
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
        return e.getZeroValueForType(method.ReturnType)
    }

    // Procedure (no return type) - return nil
    return e.runtime.NewNilValue()
}

// getZeroValueForType returns the zero value for a given type.
// Task 3.12.2: Helper for initializing Result variable.
func (e *Evaluator) getZeroValueForType(typeExpr ast.Expression) Value {
    // Delegate to existing zero value logic
    // This already handles Integer→0, String→"", Boolean→false, etc.
    return e.evaluateTypeZeroValue(typeExpr)
}
```

**Design Notes**:

1. **Environment Management**: Creates child environment, preserves enclosing scope
2. **Self Binding**: Record instance becomes accessible as `Self` in method body
3. **Result Variable**: Follows DWScript convention for function return values
4. **Error Handling**: Validates parameter count, handles early returns
5. **Var Parameters**: Noted as TODO - rare feature, defer to Phase 4

---

### 5. Error Handling

**Error Cases**:

1. **Method not found after check**:
   ```go
   if !found {
       return e.newError(node, "internal error: method '%s' not retrievable", memberName)
   }
   ```

2. **Wrong parameter count**:
   ```go
   if len(args) != len(method.Parameters) {
       return e.newError(node, "method '%s' expects %d parameters, got %d", ...)
   }
   ```

3. **Evaluation errors in arguments**:
   ```go
   args[i] = e.Eval(arg, ctx)
   if e.isError(args[i]) {
       return args[i]  // Propagate immediately
   }
   ```

**Control Flow**:
- **Early return**: Check `ctx.HasReturn()`, extract value, clear flag
- **Exit statement**: Propagate `HasExit()` up the stack
- **Break/Continue**: Should not occur in method context (error if encountered)

---

## Testing Strategy

### Unit Tests

**File**: `internal/interp/evaluator/record_methods_test.go` (NEW)

```go
func TestRecordMethodExecution(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected string
    }{
        {
            name: "simple method call",
            code: `
                type TPoint = record
                    X, Y: Integer;
                    function Sum: Integer;
                end;

                function TPoint.Sum: Integer;
                begin
                    Result := Self.X + Self.Y;
                end;

                var p: TPoint;
                p.X := 3;
                p.Y := 4;
                PrintLn(p.Sum());  // Should print 7
            `,
            expected: "7\n",
        },
        {
            name: "method with parameters",
            code: `
                type TVector = record
                    X, Y: Float;
                    function Dot(other: TVector): Float;
                end;

                function TVector.Dot(other: TVector): Float;
                begin
                    Result := Self.X * other.X + Self.Y * other.Y;
                end;

                var v1, v2: TVector;
                v1.X := 1.0; v1.Y := 2.0;
                v2.X := 3.0; v2.Y := 4.0;
                PrintLn(v1.Dot(v2));  // Should print 11.0
            `,
            expected: "11.0\n",
        },
        {
            name: "method accessing fields",
            code: `
                type TCounter = record
                    Value: Integer;
                    procedure Increment;
                    procedure Reset;
                end;

                procedure TCounter.Increment;
                begin
                    Self.Value := Self.Value + 1;
                end;

                procedure TCounter.Reset;
                begin
                    Self.Value := 0;
                end;

                var c: TCounter;
                c.Value := 5;
                c.Increment();
                c.Increment();
                PrintLn(c.Value);  // Should print 7
                c.Reset();
                PrintLn(c.Value);  // Should print 0
            `,
            expected: "7\n0\n",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            output := runCode(t, tt.code)
            if output != tt.expected {
                t.Errorf("expected %q, got %q", tt.expected, output)
            }
        })
    }
}
```

### Integration Tests

**Existing fixture tests**: All tests in `testdata/fixtures/` with record methods should continue passing with zero behavior changes.

**Key fixtures to verify**:
- `testdata/fixtures/RecordMethods/basic.dws`
- `testdata/fixtures/RecordMethods/with_parameters.dws`
- `testdata/fixtures/RecordMethods/accessing_fields.dws`

**Verification command**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/RecordMethods
```

---

## Migration Checklist

- [ ] **Step 1**: Add `GetRecordMethod` to `RecordInstanceValue` interface
  - File: `internal/interp/runtime/values.go`
  - Lines: ~1

- [ ] **Step 2**: Implement `GetRecordMethod` on `RecordValue`
  - File: `internal/interp/interpreter.go`
  - Lines: ~5

- [ ] **Step 3**: Create `callRecordMethod` implementation
  - File: `internal/interp/evaluator/record_methods.go` (NEW)
  - Lines: ~60

- [ ] **Step 4**: Update member access dispatch
  - File: `internal/interp/evaluator/visitor_expressions_members.go`
  - Lines: ~20 (replace 4 lines)

- [ ] **Step 5**: Run unit tests
  - Command: `go test ./internal/interp/evaluator`
  - Expected: All pass

- [ ] **Step 6**: Run fixture tests
  - Command: `go test -v ./internal/interp -run TestDWScriptFixtures`
  - Expected: No new failures

- [ ] **Step 7**: Verify EvalNode count reduction
  - Command: `grep -rn "adapter.EvalNode" internal/interp/evaluator/ | wc -l`
  - Expected: 26 (down from 27)

---

## Expected Outcomes

### Quantitative

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| EvalNode calls | 27 | 26 | -1 |
| LOC (evaluator) | Baseline | +80 | +80 (new file) |
| Interface methods | 5 | 6 | +1 |
| Test failures | 0 | 0 | 0 |

### Qualitative

- ✅ **Cleaner architecture**: Record methods no longer special-cased through adapter
- ✅ **Consistent patterns**: Records and helpers use similar execution paths
- ✅ **Reduced coupling**: One less adapter dependency
- ✅ **Maintained separation**: Method pointers still use adapter (appropriate boundary)

---

## Risks & Mitigations

### Risk 1: Interface Change Breaks Implementations

**Likelihood**: LOW
**Impact**: HIGH (compilation failures)

**Mitigation**:
- All implementations are in our codebase (no external consumers)
- Compiler will catch missing implementations immediately
- Implementation is trivial (3-line method)

### Risk 2: Different Semantics Than Interpreter

**Likelihood**: LOW
**Impact**: MEDIUM (test failures)

**Mitigation**:
- Reusing proven `CallASTHelperMethod` patterns
- Comprehensive fixture tests verify behavior
- Small, incremental change - easy to debug

### Risk 3: Performance Regression

**Likelihood**: VERY LOW
**Impact**: LOW

**Mitigation**:
- No additional allocations vs current approach
- Eliminates adapter indirection (slight speedup expected)
- Can benchmark if concerns arise

---

## Rollback Plan

If issues are discovered post-implementation:

1. **Revert interface change**: Remove `GetRecordMethod` from interface
2. **Restore adapter delegation**: Change line 251 back to `e.adapter.EvalNode(node)`
3. **Delete new file**: Remove `record_methods.go`
4. **Run tests**: Verify all tests pass with rollback

**Rollback effort**: ~5 minutes (simple git revert)

---

## Future Work

### Phase 4 Considerations

When migrating full OOP infrastructure to evaluator:

1. **Method pointers**: Eliminate `CreateBoundMethodPointer` adapter call
2. **Var parameters**: Implement reference semantics for var params
3. **Method overloading**: Handle record methods with overloads
4. **Virtual methods**: Support virtual/override for record methods (if added to language)

### Potential Optimizations

1. **Method caching**: Cache method lookups in CallExpression evaluation
2. **Environment pooling**: Reuse method environments to reduce allocations
3. **Inline simple methods**: Detect and inline trivial methods (e.g., getters)

---

## References

**Related Documents**:
- [docs/evalnode-audit.md](../evalnode-audit.md) - EvalNode call inventory
- [docs/phase3.9-3.11-summary.md](../phase3.9-3.11-summary.md) - Recent consolidation work
- [PLAN.md](../../PLAN.md#L359-L366) - Task 3.12.2 definition

**Related Code**:
- `internal/interp/evaluator/helper_methods.go` - Similar execution patterns
- `internal/interp/evaluator/visitor_expressions_members.go` - Member access dispatch
- `internal/interp/runtime/values.go` - RecordInstanceValue interface

**Related Issues/Tasks**:
- Task 3.12.1: EvalNode call audit (complete)
- Task 3.12.3: Compound assignment migration (next)
- Phase 4: Full OOP infrastructure migration (future)
