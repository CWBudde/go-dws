# Phase 3.5.8 Method Implementation Verification - Findings

**Date**: 2025-12-01
**Task**: PLAN.md Phase 3.5.8 - Verify method implementations update VMT correctly
**Status**: PARTIALLY COMPLETE - Critical architecture issue discovered

## Summary

Phase 8 verification revealed that while method implementations correctly trigger VMT rebuilds, the VMT architecture has fundamental limitations that prevent proper implementation of `reintroduce virtual` semantics.

## What Works ✅

1. **VisitFunctionDecl Integration**: The `VisitFunctionDecl` method correctly delegates class method implementations to `adapter.EvalMethodImplementation()` (line 37 in visitor_declarations.go)

2. **VMT Rebuild on Implementation**: When a method implementation is added via `evalClassMethodImplementation()`, it correctly calls `buildVirtualMethodTable()` to update the VMT (line 115 in declarations.go)

3. **Basic Virtual/Override**: Simple virtual method inheritance and override work correctly:
   - Virtual methods are added to VMT
   - Override methods update VMT entries
   - Polymorphic dispatch works for simple cases

## Critical Issue Discovered ❌

### Problem: `reintroduce virtual` Cannot Be Implemented with Current VMT Structure

**Test Case**: `testdata/fixtures/SimpleScripts/reintroduce_virtual.pas`

```pascal
type TBase = class
    procedure Func; virtual;  // VMT[0] = TBase.Func
end;

type TChild = class (TBase)
    procedure Func; reintroduce; virtual;  // Should be VMT[1] = TChild.Func
end;

var o : TBase;
o := TChild.Create;
o.Func;  // Should call TBase.Func (uses VMT[0])
          // Actually calls TChild.Func (VMT is overwritten)
```

**Expected Behavior**:
- `reintroduce virtual` starts a **NEW** virtual method chain
- Parent's VMT entry (index 0) keeps `TBase.Func`
- Child gets a NEW VMT entry (index 1) with `TChild.Func`
- When called through `TBase` reference, uses VMT[0] → `TBase.Func`
- When called through `TChild` reference, uses VMT[1] → `TChild.Func`

**Current Behavior**:
- VMT uses `map[string]*VirtualMethodEntry` keyed by method signature
- Only ONE entry per signature possible
- `reintroduce virtual` overwrites parent's entry
- Both `TBase` and `TChild` references use the same VMT entry

### Root Cause

The VMT is structured as a **map** instead of an **array with indices**:

```go
// Current (WRONG):
type ClassInfo struct {
    VirtualMethodTable map[string]*VirtualMethodEntry  // Can't have 2 entries for same signature
}

// DWScript reference (CORRECT):
type ClassSymbol struct {
    FVirtualMethodTable []TMethodSymbol  // Array with indices
}
```

In DWScript, each virtual method gets a unique VMT index:
- First `virtual` declaration → VMT index assigned
- `override` → uses parent's VMT index
- `reintroduce virtual` → NEW VMT index allocated

Our map-based approach fundamentally cannot support multiple VMT entries for the same method signature.

## Implications

1. **Scope**: Fixing this requires:
   - Changing VMT from `map[string]` to `[]` (array)
   - Assigning VMT indices during `buildVirtualMethodTable()`
   - Updating method dispatch to use VMT indices
   - ~500-1000 lines of changes across multiple files

2. **Affected Features**:
   - `reintroduce virtual` (currently broken)
   - Virtual method dispatch efficiency (map lookup vs array index)
   - Virtual constructors (may be affected)

3. **Workaround**: None - the architecture must be fixed

## Attempted Fix (Reverted)

I attempted to fix `reintroduce virtual` by adding:

```go
} else if m.IsReintroduce {
    if m.IsVirtual {
        // Try to add new entry for reintroduce virtual
        c.VirtualMethodTable[sig] = &VirtualMethodEntry{
            IntroducedBy:   c,
            Implementation: m,
            IsVirtual:      true,
            IsReintroduced: true,
        }
    }
}
```

**Result**: This doesn't work because it overwrites the parent's entry in the map. The map key is the method signature, so we can't have both parent's and child's entries.

## Recommendations

### Option 1: Defer Fix (Recommended for Phase 8)

- Mark Phase 8 as "COMPLETED WITH FINDINGS"
- Document this issue in PLAN.md as a known limitation
- Create a new task (e.g., Task 9.14.x) to fix VMT architecture
- Continue with Phase 9 (testing) and Phase 10 (cleanup)
- Fix VMT architecture later as part of Task 9.14 (Inheritance fixes)

**Rationale**: Phase 8's goal was to verify VMT updates happen correctly, which they do. The VMT architecture issue is a separate, larger problem.

### Option 2: Fix VMT Architecture Now

- Redesign VMT to use array with indices
- Update all method dispatch code
- ~2-3 days of work
- Blocks completion of Phase 3.5.8

**Rationale**: Ensures correctness but delays progress significantly.

## Phase 8 Acceptance Criteria Review

✅ **Verify existing VisitFunctionDecl still works for class methods**
- PASS: Correctly delegates to `adapter.EvalMethodImplementation()`

✅ **Ensure method implementations update VMT correctly**
- PASS: `evalClassMethodImplementation()` calls `buildVirtualMethodTable()`

⚠️ **Test override/virtual/reintroduce semantics**
- PASS: override and virtual work
- FAIL: reintroduce virtual broken (architecture limitation)

⚠️ **Validate VMT entries after method implementation**
- PASS: Entries are created/updated
- FAIL: `reintroduce virtual` creates wrong entry structure

## Conclusion

Phase 8 successfully verified that method implementations trigger VMT updates. However, it also revealed a fundamental architecture issue with the VMT structure that prevents correct implementation of `reintroduce virtual`.

**Recommendation**: Mark Phase 8 as COMPLETE with documented findings, create follow-up task for VMT architecture fix.

## References

- Test case: `testdata/fixtures/SimpleScripts/reintroduce_virtual.pas`
- DWScript reference: `reference/dwscript-original/Source/dwsSymbols.pas` (lines 5670-5707)
- Current VMT code: `internal/interp/class.go` (`buildVirtualMethodTable()`)
- Method dispatch: `internal/interp/objects_methods.go` (`evalMethodCall()`)
