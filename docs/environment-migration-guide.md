# Environment Synchronization Migration Guide

**Phase**: 3.8.2 | **Status**: Ready for Implementation | **Version**: 1.0

## Purpose

This guide helps developers migrate the 149 `i.env =` assignments to use the new environment helper methods introduced in Phase 3.8.1. The helpers ensure that `i.env` and `i.ctx.env` stay synchronized, preventing the environment desynchronization that caused 139 test failures in commit af2ba646.

---

## Quick Reference: Helper Methods

Three helper methods were added to `Interpreter` in Phase 3.8.1:

```go
// SetEnvironment atomically updates both i.env and i.ctx.env
func (i *Interpreter) SetEnvironment(env *Environment)

// PushEnvironment creates enclosed env and sets both references
func (i *Interpreter) PushEnvironment(parent *Environment) *Environment

// RestoreEnvironment restores saved env to both references
func (i *Interpreter) RestoreEnvironment(saved *Environment)
```

**Implementation**: [internal/interp/interpreter.go:291-363](../internal/interp/interpreter.go)
**Tests**: [internal/interp/environment_helpers_test.go](../internal/interp/environment_helpers_test.go)

---

## Migration Patterns

### Pattern 1: Simple Assignment

**Before**:
```go
i.env = newEnv
```

**After**:
```go
i.SetEnvironment(newEnv)
```

**Occurrences**: Rare (most assignments create new environments)

---

### Pattern 2: Save → Set → Restore (Most Common)

**Before**:
```go
savedEnv := i.env
i.env = NewEnclosedEnvironment(i.env)
// ... execution ...
i.env = savedEnv
```

**After**:
```go
savedEnv := i.env
i.PushEnvironment(i.env)
// ... execution ...
i.RestoreEnvironment(savedEnv)
```

**Occurrences**: 136 assignments (91%)
**Files**: adapter_methods.go, objects_methods.go, objects_properties.go, and 11 others

**Example** (adapter_methods.go:202-208):
```go
// Before:
savedEnv := i.env                          // Line 202
methodEnv := NewEnclosedEnvironment(i.env) // Line 203
i.env = methodEnv                          // Line 204
// ... method execution ...
i.env = savedEnv                           // Line 208

// After:
savedEnv := i.env
methodEnv := i.PushEnvironment(i.env)
// ... method execution ...
i.RestoreEnvironment(savedEnv)
```

---

### Pattern 3: Save → Set → Error Handling → Restore

**Before**:
```go
savedEnv := i.env
tempEnv := NewEnclosedEnvironment(i.env)
i.env = tempEnv
// ... execution ...
if isError(result) {
    i.env = savedEnv
    return result
}
i.env = savedEnv
```

**After**:
```go
savedEnv := i.env
tempEnv := i.PushEnvironment(i.env)
// ... execution ...
if isError(result) {
    i.RestoreEnvironment(savedEnv)
    return result
}
i.RestoreEnvironment(savedEnv)
```

**Occurrences**: ~60% of assignments (89 assignments)
**Files**: loops, methods, constructors, properties

**Example** (objects_methods.go:58-79):
```go
// Before:
savedEnv := i.env                          // Line 58
tempEnv := NewEnclosedEnvironment(i.env)   // Line 59
i.env = tempEnv                            // Line 63
// ... field initialization ...
if isError(fieldValue) {
    i.env = savedEnv                       // Line 70 (error path)
    return fieldValue
}
i.env = savedEnv                           // Line 79 (success path)

// After:
savedEnv := i.env
tempEnv := i.PushEnvironment(i.env)
// ... field initialization ...
if isError(fieldValue) {
    i.RestoreEnvironment(savedEnv)
    return fieldValue
}
i.RestoreEnvironment(savedEnv)
```

---

### Pattern 4: Deferred Restoration

**Before**:
```go
savedEnv := i.env
tempEnv := NewEnclosedEnvironment(i.env)
i.env = tempEnv
defer func() { i.env = savedEnv }()
```

**After**:
```go
savedEnv := i.env
tempEnv := i.PushEnvironment(i.env)
defer func() { i.RestoreEnvironment(savedEnv) }()
```

**Occurrences**: 6 assignments (4%)
**Files**: declarations.go, statements_control.go

**Example** (declarations.go:256-260):
```go
// Before:
savedEnv := i.env                          // Line 256
tempEnv := NewEnclosedEnvironment(i.env)   // Line 257
i.env = tempEnv                            // Line 259
defer func() { i.env = savedEnv }()        // Line 260

// After:
savedEnv := i.env
tempEnv := i.PushEnvironment(i.env)
defer func() { i.RestoreEnvironment(savedEnv) }()
```

---

### Pattern 5: Loop Variable Scoping

**Before**:
```go
loopEnv := NewEnclosedEnvironment(i.env)
savedEnv := i.env
i.env = loopEnv
for current := start; current <= end; current += step {
    i.env.Define(loopVarName, makeValue(current))
    result = i.Eval(stmt.Body)
    if isError(result) {
        i.env = savedEnv
        return result
    }
}
i.env = savedEnv
```

**After**:
```go
savedEnv := i.env
loopEnv := i.PushEnvironment(i.env)
for current := start; current <= end; current += step {
    loopEnv.Define(loopVarName, makeValue(current))
    result = i.Eval(stmt.Body)
    if isError(result) {
        i.RestoreEnvironment(savedEnv)
        return result
    }
}
i.RestoreEnvironment(savedEnv)
```

**Occurrences**: 15 assignments
**File**: statements_loops.go

**Key Change**: Can now use `loopEnv` directly since `PushEnvironment()` returns it.

---

## Category-by-Category Migration

### 1. User Function Callbacks (6 assignments)

**File**: [internal/interp/user_function_callbacks.go](../internal/interp/user_function_callbacks.go)
**Priority**: CRITICAL
**Complexity**: HIGH (already has EnvSyncer callback logic)

**Special Notes**:
- This file already has partial synchronization via `EnvSyncerFunc`
- May require refactoring to use helpers instead of manual sync
- Review lines: 183, 217, 219, 247, 252, 264

**Testing**:
```bash
go test ./internal/interp -run "TestUserFunction|TestRecursion"
```

---

### 2. Method Dispatch (24 assignments)

**File**: [internal/interp/adapter_methods.go](../internal/interp/adapter_methods.go)
**Priority**: HIGH
**Complexity**: HIGH (most assignments, complex dispatch logic)

**Pattern**: Mostly Pattern 2 (Save → Set → Restore)

**Lines to Migrate**: 204, 208, 247, 270, 324, 326, 339, 343, 374, 378, 417, 421, 443, 458, 468, 483, 516, 528, 554, 558, 599, 603, 623, 627

**Testing**:
```bash
go test ./internal/interp -run "TestMethod|TestHelper"
```

---

### 3. Lambda/Function Pointers (9 assignments)

**Files**:
- [internal/interp/functions_pointers.go](../internal/interp/functions_pointers.go) (7)
- [internal/interp/adapter_functions.go](../internal/interp/adapter_functions.go) (2)

**Priority**: HIGH
**Complexity**: HIGH (closure environment, complex lifetime)

**Special Pattern**: Includes `Self` binding
```go
// After migration:
savedEnv := i.env
funcEnv := i.PushEnvironment(closureEnv)
funcEnv.Define("Self", funcPtr.SelfObject)
// ... execution ...
i.RestoreEnvironment(savedEnv)
```

**Testing**:
```bash
go test ./internal/interp -run "TestLambda|TestFunctionPointer|TestClosure"
```

---

### 4. Object Methods (23 assignments)

**File**: [internal/interp/objects_methods.go](../internal/interp/objects_methods.go)
**Priority**: MEDIUM
**Complexity**: MEDIUM (well-structured, many error paths)

**Pattern**: Pattern 3 (Error handling)

**Lines to Migrate**: 63, 70, 79, 114, 122, 132, 153, 190, 235, 354, 377, 381, 678, 697, 702, 708, 711, 755, 803, 833, 837, 878, 911

**Testing**:
```bash
go test ./internal/interp -run "TestObjectMethod|TestConstructor|TestDestructor"
```

---

### 5. Property Access (22 assignments)

**File**: [internal/interp/objects_properties.go](../internal/interp/objects_properties.go)
**Priority**: MEDIUM
**Complexity**: MEDIUM (consistent pattern, short-lived scopes)

**Pattern**: Pattern 2 (Save → Set → Restore)

**Testing**:
```bash
go test ./internal/interp -run "TestProperty|TestGetter|TestSetter"
```

---

### 6. Loop Scoping (15 assignments)

**File**: [internal/interp/statements_loops.go](../internal/interp/statements_loops.go)
**Priority**: MEDIUM
**Complexity**: MEDIUM (break/continue/return complicate restoration)

**Pattern**: Pattern 5 (Loop variable scoping)

**Lines to Migrate**: 153, 205, 234, 257, 278, 293, 318, 342, 373, 382, 418, 445, 467, 490, 495

**Testing**:
```bash
go test ./internal/interp -run "TestForLoop|TestWhileLoop|TestForInLoop|TestArray"
```

---

### 7-13. Remaining Categories

See [environment-audit.md](environment-audit.md) for detailed breakdown of:
- Object Lifecycle (13 assignments)
- Record Methods (10 assignments)
- Declarations (8 assignments)
- Operators (6 assignments)
- Hierarchy (4 assignments)
- Helpers/Control/Interface (9 assignments)
- Exceptions (2 assignments)

---

## Migration Workflow

### Per-Category Workflow

For each category (following the priority order):

#### 1. Understand Current Code
- Read the file thoroughly
- Identify all `i.env =` assignments
- Understand the control flow and error paths
- Note any special patterns (defer, Self binding, etc.)

#### 2. Apply Migration Pattern
- Replace each `i.env = newEnv` with `i.SetEnvironment(newEnv)`
- Replace each `i.env = NewEnclosedEnvironment(...)` with `i.PushEnvironment(...)`
- Replace each `i.env = savedEnv` with `i.RestoreEnvironment(savedEnv)`
- Verify all error paths are updated

#### 3. Run Category-Specific Tests
```bash
# Run tests for the specific category
go test ./internal/interp -run TestCategoryName -v

# Verify no new failures
go test ./internal/interp -run TestCategoryName
```

#### 4. Run Full Test Suite
```bash
# Run all tests to check for regressions
go test ./...

# Count fixture failures (must be ≤ 892)
go test -v ./internal/interp -run TestDWScriptFixtures 2>&1 | grep "FAIL:" | wc -l
```

#### 5. Commit Changes
```bash
git add internal/interp/<file>.go
git commit -m "Phase 3.8.2.N: Migrate <category> environments to helpers"
```

**Important**: Commit after each category, not after each file. This allows easy rollback if needed.

---

## Testing Strategy

### Test Levels

1. **Unit Tests**: Test the specific feature (methods, loops, properties)
2. **Integration Tests**: Test interactions between features
3. **Fixture Tests**: Comprehensive DWScript test suite (892 baseline)

### Verification Commands

```bash
# Run specific category tests
go test ./internal/interp -run TestMethodCall -v
go test ./internal/interp -run TestLoops -v
go test ./internal/interp -run TestProperties -v

# Run all interp tests
go test ./internal/interp -v

# Run full test suite
go test ./...

# Count fixture failures (baseline: 892)
go test -v ./internal/interp -run TestDWScriptFixtures 2>&1 | grep "FAIL:" | wc -l
```

### Acceptance Criteria

✅ All category-specific tests pass
✅ Full test suite passes (except baseline fixture failures)
✅ Fixture failures ≤ 892 (no new failures)
✅ No performance regression

---

## Common Pitfalls

### Pitfall 1: Forgetting Error Paths

**Wrong**:
```go
savedEnv := i.env
i.PushEnvironment(i.env)
if isError(result) {
    return result  // ❌ Forgot to restore!
}
i.RestoreEnvironment(savedEnv)
```

**Correct**:
```go
savedEnv := i.env
i.PushEnvironment(i.env)
if isError(result) {
    i.RestoreEnvironment(savedEnv)  // ✅ Restore before return
    return result
}
i.RestoreEnvironment(savedEnv)
```

---

### Pitfall 2: Restoring Wrong Environment

**Wrong**:
```go
savedEnv := i.env
funcEnv := i.PushEnvironment(closureEnv)  // Note: closureEnv, not i.env
// ... execution ...
i.RestoreEnvironment(i.env)  // ❌ Restoring current env, not saved!
```

**Correct**:
```go
savedEnv := i.env
funcEnv := i.PushEnvironment(closureEnv)
// ... execution ...
i.RestoreEnvironment(savedEnv)  // ✅ Restore the saved environment
```

---

### Pitfall 3: Mixing Old and New Patterns

**Wrong**:
```go
savedEnv := i.env
funcEnv := i.PushEnvironment(i.env)  // ✅ Uses helper
// ... execution ...
i.env = savedEnv  // ❌ Direct assignment instead of helper
```

**Correct**:
```go
savedEnv := i.env
funcEnv := i.PushEnvironment(i.env)
// ... execution ...
i.RestoreEnvironment(savedEnv)  // ✅ Uses helper consistently
```

---

### Pitfall 4: Not Using Returned Environment

**Inefficient** (still works):
```go
i.PushEnvironment(i.env)
i.env.Define("var", value)  // Uses i.env directly
```

**Better**:
```go
newEnv := i.PushEnvironment(i.env)
newEnv.Define("var", value)  // Uses returned environment
```

---

## Rollback Procedures

### If Category Migration Fails

1. **Identify the Issue**:
   ```bash
   go test ./internal/interp -run TestCategoryName -v
   # Review failure messages
   ```

2. **Revert the Commit**:
   ```bash
   git revert HEAD
   # Or for specific file:
   git checkout HEAD~1 internal/interp/<file>.go
   ```

3. **Verify Baseline Restored**:
   ```bash
   go test ./...
   ```

4. **Document Failure**:
   - Add notes to PLAN.md
   - Update migration status
   - Investigate root cause before retry

---

## Progress Tracking

### Migration Checklist

Track progress in [PLAN.md](../PLAN.md) under Phase 3.8.2:

- [ ] Priority 1: User Function Callbacks (6 assignments)
- [ ] Priority 2: Method Dispatch (24 assignments)
- [ ] Priority 3: Lambda/Pointers (9 assignments)
- [ ] Priority 4: Object Methods (23 assignments)
- [ ] Priority 5: Properties (22 assignments)
- [ ] Priority 6: Loops (15 assignments)
- [ ] Priority 7: Object Lifecycle (13 assignments)
- [ ] Priority 8: Record Methods (10 assignments)
- [ ] Priority 9: Declarations (8 assignments)
- [ ] Priority 10: Operators (6 assignments)
- [ ] Priority 11: Hierarchy (4 assignments)
- [ ] Priority 12: Helpers/Misc (9 assignments)
- [ ] Priority 13: Exceptions (2 assignments)

### Progress Metrics

- **Total Assignments**: 149
- **Migrated**: 0
- **Remaining**: 149
- **Estimated Time**: 6.5 days

Update these metrics after each category migration.

---

## References

- [environment-audit.md](environment-audit.md) - Complete inventory of 149 assignments
- [PLAN.md Phase 3.8.2](../PLAN.md#phase-382-incremental-environment-sync-migration-1-2-weeks) - Migration tasks
- [internal/interp/interpreter.go:291-363](../internal/interp/interpreter.go) - Helper implementations
- [internal/interp/environment_helpers_test.go](../internal/interp/environment_helpers_test.go) - Helper tests

---

## FAQ

### Q: Can I migrate multiple categories in one commit?

**A**: No. Commit after each category for granular rollback. If a category spans multiple files (e.g., Lambda/Pointers), you can commit all files together for that category.

### Q: What if tests fail after migration?

**A**: Revert the commit, review the changes, ensure all error paths are updated, and retry. Document the failure in PLAN.md.

### Q: Can I skip the helper and just add `i.ctx.SetEnv()` calls?

**A**: Technically yes, but the helpers provide:
- Atomic updates (both refs updated together)
- Consistent pattern across codebase
- Single point of failure/debugging
- Easier future refactoring

### Q: What about performance impact?

**A**: The helpers add negligible overhead (~2 extra function calls). The adapter creation was already happening in user function execution. Migration should not introduce performance regression.

### Q: How do I handle nested environments?

**A**: Save → Push → Execute → Restore at each level. See Pattern 2 and TestNestedPushRestore in the tests.

---

**Migration Guide Version**: 1.0
**Last Updated**: 2025-12-06
**Status**: Ready for Phase 3.8.2 Implementation
