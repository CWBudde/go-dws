# Phase 3 Adapter Strategy: General Facilities vs Operation-Specific Interfaces

**Date**: 2025-12-11
**Status**: Architecture Decision Record

## Context

During Phase 3 (Interpreter Architecture Refactoring), we're migrating the evaluator from adapter-dependent to self-sufficient. The initial plan for task 3.2.11g proposed adding a new `IndexedPropertyWriter` adapter interface specifically for indexed property assignment.

## Problem

The proposed `IndexedPropertyWriter` interface contradicted Phase 3's core goal:

```go
// PROPOSED (but rejected):
type IndexedPropertyWriter interface {
    WriteIndexedProperty(obj Value, propName string, indices []Value, value Value) (Value, error)
    WriteDefaultProperty(obj Value, indices []Value, value Value) (Value, error)
}
```

**Why this was wrong**:
1. **Increases adapter dependency** - adds MORE adapter methods when we're trying to REDUCE them
2. **Operation-specific interface** - creates a new interface for a single operation type
3. **Violates existing patterns** - task 3.2.11d already solved record properties WITHOUT new interfaces

## Decision

**Use general-purpose facilities instead of operation-specific adapters.**

### Pattern from 3.2.11d (Record Property Setters)

Record property writes were successfully migrated using:
1. **TypeSystem** for metadata lookup (RecordInfo.GetProperty)
2. **General method dispatch** via `adapter.ExecuteMethodWithSelf()` (OOP facility)
3. **No property-specific adapter** needed

### Applied to Indexed Properties (3.2.11g, 3.2.11h)

The same pattern works for indexed properties:

```go
// For obj.Prop[i] := value (named indexed property)
// 1. Look up property metadata
propInfo := classInfo.GetProperty("Prop")  // TypeSystem - data access

// 2. Build argument list for setter
args := append(indices, value)  // [i, value]

// 3. Execute setter method
adapter.ExecuteMethodWithSelf(obj, propInfo.WriteSpec, args)  // OOP facility
```

```go
// For obj[i] := value (default indexed property)
// 1. Get default property
accessor := obj.(PropertyAccessor)
propDesc := accessor.GetDefaultProperty()  // Already exists in runtime!

// 2. Build argument list
args := append(indices, value)

// 3. Execute setter method
adapter.ExecuteMethodWithSelf(obj, propDesc.Setter, args)
```

## Benefits

### Clean Separation of Concerns

| Layer | Responsibility | Examples |
|-------|---------------|----------|
| **TypeSystem** | Metadata lookup | GetProperty, GetDefaultProperty |
| **OOPEngine** | Method dispatch | ExecuteMethodWithSelf, CallMethod |
| **Evaluator** | Orchestration | Combine lookup + dispatch |

### Fewer Interfaces, Better Testability

**Before** (rejected approach):
- InterpreterAdapter: 65 methods
- IndexedPropertyWriter: +2 methods (67 total)
- Hard to mock (67 methods to stub)

**After** (accepted approach):
- OOPEngine: ~18 methods (general purpose)
- DeclHandler: ~37 methods (general purpose)
- ExceptionManager: ~6 methods (general purpose)
- No IndexedPropertyWriter needed
- Easy to mock (focused interfaces)

### Consistent Migration Pattern

All property operations now follow the same pattern:

1. **Record field properties** (3.2.11d): TypeSystem + ExecuteMethodWithSelf ✅
2. **Indexed properties** (3.2.11g): TypeSystem + ExecuteMethodWithSelf ✅
3. **Default properties** (3.2.11h): PropertyAccessor + ExecuteMethodWithSelf ✅

## Infrastructure Already Exists

The runtime already has everything we need:

### PropertyAccessor Interface
```go
// internal/interp/runtime/property.go:30
type PropertyAccessor interface {
    GetDefaultProperty() *PropertyDescriptor  // ← already exists!
    // ... other methods
}
```

### IClassInfo Interface
```go
// internal/interp/runtime/class_interface.go:47
type IClassInfo interface {
    GetProperty(name string) *PropertyInfo
    GetDefaultProperty() *PropertyInfo  // ← already exists!
    // ... other methods
}
```

### PropertyInfo Struct
```go
// internal/interp/runtime/class_interface.go:100
type PropertyInfo struct {
    Name      string
    ReadSpec  string  // Getter method name
    WriteSpec string  // Setter method name ← this is what we call
    IsIndexed bool
    IsDefault bool
}
```

## Implementation Impact

### Task Changes

| Old Task | New Task | Change |
|----------|----------|--------|
| 3.2.11g: Add IndexedPropertyWriter (~1h) | 3.2.11g: Migrate indexed properties (~2h) | Use existing infrastructure |
| 3.2.11h: Migrate indexed properties (~1.5h) | 3.2.11h: Migrate default properties (~1.5h) | Use PropertyAccessor interface |
| 3.2.11i: Migrate default properties (~1h) | 3.2.11i: Migrate implicit Self (~1.5h) | Renumbered |

**Total estimate**: 13.5h (was 14h - saved 0.5h by not creating interface)

### Risk Reduction

**Old approach**:
- Medium risk: 3.2.11h - new adapter interface needed
- Interface design might be wrong, requiring rework

**New approach**:
- Medium risk: 3.2.11g, 3.2.11h - use proven infrastructure
- Follows established pattern from 3.2.11d
- No interface design risk

## Lessons for Future Tasks

### When to Create New Adapter Interfaces

**DON'T** create operation-specific interfaces like:
- ❌ `IndexedPropertyWriter` - too specific
- ❌ `DefaultPropertyHandler` - too specific
- ❌ `CompoundAssignmentExecutor` - too specific

**DO** use general-purpose facilities like:
- ✅ `OOPEngine.ExecuteMethodWithSelf()` - works for ANY method
- ✅ `OOPEngine.CallMethod()` - works for ANY method
- ✅ `TypeSystem.GetProperty()` - works for ANY property

### Pattern Recognition

If you can phrase the operation as:
- "Look up X metadata in TypeSystem"
- "Execute Y method via OOPEngine"

Then **DON'T create a new adapter interface** - use the general facilities.

### Test: Would This Interface Have ONE Method?

If a proposed adapter interface has only 1-2 methods, ask:
1. Is this really a general facility?
2. Or is it just wrapping a specific operation?

If #2, don't create it - use existing general facilities instead.

## References

- **Task 3.2.11d**: Record property setters - established the pattern
- **docs/evaluator-architecture.md**: Evaluator design principles
- **PLAN.md**: Phase 3 goals (reduce adapter dependencies)
- **internal/interp/runtime/property.go**: PropertyAccessor interface
- **internal/interp/runtime/class_interface.go**: IClassInfo interface

## Conclusion

**The right abstraction level is general facilities, not operation-specific wrappers.**

This decision:
- ✅ Aligns with Phase 3 goals (reduce adapter dependencies)
- ✅ Follows established patterns (3.2.11d)
- ✅ Uses existing infrastructure (PropertyAccessor, IClassInfo)
- ✅ Keeps interfaces focused and mockable
- ✅ Reduces overall complexity

By catching this in planning, we avoided wasting time building the wrong abstraction and then having to refactor it later.
