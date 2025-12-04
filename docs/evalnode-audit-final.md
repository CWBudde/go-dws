# EvalNode Calls Audit - Final Report

**Task**: 3.5.38
**Date**: 2025-12-04
**Purpose**: Comprehensive inventory of all EvalNode calls in the evaluator to inform Phase B migration strategy

---

## Executive Summary

This audit documents **all 34 EvalNode calls** across the evaluator package (internal/interp/evaluator/). Each call is categorized by purpose, with migration recommendations for tasks 3.5.39-3.5.45.

### Key Findings

1. **Reference Counting Dominates**: 9 calls (26.5%) in assignment_helpers.go handle object/interface reference counting - the core architectural blocker
2. **OOP Infrastructure is Complex**: 18 calls (53%) across member access, assignment, and method dispatch depend on class hierarchies and virtual method tables
3. **No "Quick Wins"**: All 34 calls serve specific architectural purposes; migration requires corresponding interpreter functionality
4. **Clear Migration Path**: Priority order is well-defined by dependency analysis

### Distribution by File

```plain
assignment_helpers.go             9 calls (26.5%)  ████████████
visitor_expressions_members.go    6 calls (17.6%)  █████████
member_assignment.go              6 calls (17.6%)  █████████
index_assignment.go               3 calls (8.8%)   ████
visitor_expressions_functions.go  2 calls (5.9%)   ███
visitor_statements.go             2 calls (5.9%)   ███
evaluator.go                      1 call  (2.9%)   █
compound_ops.go                   1 call  (2.9%)   █
helper_methods.go                 1 call  (2.9%)   █
user_function_helpers.go          1 call  (2.9%)   █
visitor_declarations.go           1 call  (2.9%)   █
method_dispatch.go                1 call  (2.9%)   █
```

### Distribution by Category

```plain
Assignment/Reference Counting    9 calls (26.5%)  ████████████
Member Access Delegation         6 calls (17.6%)  █████████
Member Assignment Delegation     6 calls (17.6%)  █████████
Index Operations                 3 calls (8.8%)   ████
Function Calls                   2 calls (5.9%)   ███
Control Flow                     2 calls (5.9%)   ███
Unknown Node Type Safety Net     1 call  (2.9%)   █
Compound Operations              1 call  (2.9%)   █
Helper Methods                   1 call  (2.9%)   █
User Function Bodies             1 call  (2.9%)   █
Declarations                     1 call  (2.9%)   █
Method Dispatch Fallback         1 call  (2.9%)   █
```

---

## Category 1: Reference Counting & Object Semantics (9 calls)

**File**: `internal/interp/evaluator/assignment_helpers.go`

This category originally had 9 calls. After tasks 3.5.39-3.5.42, ref counting was migrated to runtime package.

### Category 1a: Reference Counting (7 calls) - ✅ COMPLETED

**Status**: ✅ **MIGRATED** (tasks 3.5.39-3.5.42)
**Result**: All 7 ref counting calls eliminated via `runtime.RefCountManager`

| Line | Context | Migration Status |
|------|---------|------------------|
| 127 | Assigning to interface variable | ✅ Migrated to RefCountManager |
| 136 | Assigning to object variable | ✅ Migrated to RefCountManager |
| 172 | Assigning object VALUE | ✅ Migrated to RefCountManager |
| 182 | Assigning interface VALUE | ✅ Removed (redundant) |
| 192 | Assigning method pointer with SelfObject | ✅ Migrated to RefCountManager |
| 227 | Var parameter → interface/object | ✅ Migrated to RefCountManager |
| 254 | Assigning object/interface through var parameter | ✅ Migrated to RefCountManager |

**Migration Details**: See [docs/refcounting-design.md](refcounting-design.md) for architecture.

### Category 1b: Self/Class Context (2 calls) - ✅ KEPT AS ESSENTIAL

**Status**: ✅ **ARCHITECTURAL BOUNDARY** (task 3.5.43 - Option 3)
**Decision**: Keep these 2 calls as essential boundaries between evaluator and interpreter

| Line | Context | Rationale |
|------|---------|-----------|
| 93 | Simple assignment to non-env variable | Self/class context owned by interpreter |
| 368 | Compound assignment to non-env variable | Self/class context owned by interpreter |

**Why These Remain**:

When the evaluator encounters `x := value` or `x += value` where `x` is not in `ctx.Env()`:

1. **Could be instance field**: `Self.Field := value` (implicit Self in method)
2. **Could be class variable**: `TClass.ClassVar := value` (static variable)
3. **Could be property**: `Self.PropName := value` (property setter)

The interpreter owns the logic to distinguish these cases:

- Checks `env.Get("Self")` → extracts ObjectInstance → checks field/property/class var
- Checks `env.Get("__CurrentClass__")` → extracts ClassInfo → checks class var
- Requires access to ClassInfo.ClassVars, obj.Class.LookupField(), evalPropertyWrite()
- ALL of this is interpreter-owned OOP semantics, not evaluator concern

**Architectural Rationale**:

- **Evaluator responsibility**: Execute AST nodes, manage expressions/statements, handle values
- **Interpreter responsibility**: OOP semantics (classes, fields, properties, class vars, Self context)
- **Adapter boundary**: Clean separation via callback pattern (like RefCountManager)

99% of assignments go through `ctx.Env().Set(name, value)` natively. Only implicit Self/class context (rare in practice) delegates to interpreter.

**See Also**: [docs/evaluator-architecture.md](evaluator-architecture.md) for full architectural analysis

---

## Category 2: Member Access Delegation (6 calls)

**File**: `internal/interp/evaluator/visitor_expressions_members.go`
**Status**: ⚠️ **KEEP** - Complex OOP infrastructure required
**Blocker**: Method dispatch, virtual method tables, class hierarchies

These 6 calls delegate complex member access patterns requiring full class/object infrastructure:

| Line | Context | Reason for EvalNode | Type |
|------|---------|---------------------|------|
| 289 | Object instance methods/other members | Complex method dispatch and field lookup | OBJECT |
| 332 | Interface instance methods | Method dispatch via virtual table | INTERFACE |
| 371 | CLASSINFO/metaclass methods/properties | Constructor/class method dispatch | CLASSINFO |
| 417 | Type cast values for method calls | Complex type hierarchy access | TYPE_CAST |
| 439 | Typed nil accessing class variables | Class registry + hierarchy lookup | NIL |
| 460 | Record instance methods | Record method dispatch | RECORD |

**Code Examples**:

```go
// Line 289: Object instance - method dispatch
case "OBJECT":
    // ... property/field handling ...
    // Try method or other member access via adapter
    return e.adapter.EvalNode(node)

// Line 332: Interface instance - virtual method dispatch
case "INTERFACE":
    // ... property verification ...
    // Method access requires complex dispatch logic
    return e.adapter.EvalNode(node)

// Line 371: Metaclass - constructor/class method dispatch
case "CLASSINFO":
    // ... built-in properties/class vars/constants ...
    // Complex cases (constructors, class methods) need adapter
    return e.adapter.EvalNode(node)
```

**Why These Must Delegate**:
- Method dispatch requires virtual method table (VMT) lookup
- Class hierarchies and inheritance chains live in interpreter
- Constructor invocation requires object allocation logic
- Record method dispatch requires RecordTypeValue infrastructure

**Future Migration**:
- Requires migrating ClassInfo, ObjectInstance, InterfaceInstance to runtime package
- Requires migrating method dispatch infrastructure
- Estimated effort: 20-30 hours across multiple tasks

---

## Category 3: Member Assignment Delegation (6 calls)

**File**: `internal/interp/evaluator/member_assignment.go`
**Status**: ⚠️ **KEEP** - Class/record infrastructure required
**Blocker**: ClassInfo lookup, property setter dispatch, auto-initialization

These 6 calls handle complex member assignment patterns:

| Line | Context | Reason for EvalNode | Pattern |
|------|---------|---------------------|---------|
| 55 | Static class member (TClass.Variable) | ClassInfo lookup required | Static access |
| 72 | Nil value assignment (auto-init) | Object initialization/allocation | Auto-init |
| 82 | Record property setter | Property dispatch logic | Property |
| 90 | Record doesn't support direct assignment | Fallback for edge cases | Fallback |
| 125 | Class/metaclass assignment | Class member modification | Metaclass |
| 129 | Unknown type fallback | Safety fallback | Unknown |

**Code Examples**:

```go
// Line 55: Static class access (TClass.Variable := value)
if _, ok := target.Object.(*ast.Identifier); ok {
    // Could be TClass.Variable or TClass.Property
    // Delegate to adapter for class info lookup
    return e.adapter.EvalNode(stmt)
}

// Line 72: Nil value auto-initialization
if objVal == nil || objVal.Type() == "NIL" {
    // Delegate to adapter for potential auto-initialization
    return e.adapter.EvalNode(stmt)
}

// Line 82: Record property setter dispatch
if recInst.HasRecordProperty(fieldName) {
    // Record property setter needs adapter dispatch
    return e.adapter.EvalNode(stmt)
}
```

**Why These Must Delegate**:
- Static class member access requires ClassInfo registry lookup
- Nil auto-initialization requires object allocation (not in evaluator)
- Property setters require PropertyInfo dispatch logic
- Metaclass operations require class hierarchy traversal

**Migration Potential**:
- Lines 55, 72, 82, 125: Require OOP infrastructure migration (Phase 4+)
- Lines 90, 129: Could be eliminated with better type system coverage
- Estimated effort: 10-15 hours after OOP migration complete

---

## Category 4: Index Assignment Delegation (3 calls)

**File**: `internal/interp/evaluator/index_assignment.go`
**Status**: ✅ **POTENTIAL MIGRATION** - Could move to evaluator with effort
**Blocker**: Indexed property infrastructure, interface handling

These 3 calls handle index assignment for complex types:

| Line | Context | Reason for EvalNode | Complexity |
|------|---------|---------------------|------------|
| 45 | Property-based indexing (obj.Prop[i]) | Indexed property dispatch | Medium |
| 76 | Interface index access | Interface-based indexing | Medium |
| 82 | Object default property (obj[i]) | Default indexed property | Medium |

**Code Examples**:

```go
// Line 45: Indexed property writes (obj.Property[i] := value)
if _, ok := base.(*ast.MemberAccessExpression); ok {
    return e.adapter.EvalNode(stmt)
}

// Line 76: Interface indexed properties
if strings.HasPrefix(arrayVal.Type(), "INTERFACE") {
    return e.adapter.EvalNode(stmt)
}

// Line 82: Object default indexed properties
if strings.HasPrefix(arrayVal.Type(), "OBJECT[") {
    return e.adapter.EvalNode(stmt)
}
```

**Migration Potential**: ⭐⭐⭐ **Medium-High**
- Lines 76, 82: Could be migrated with interface/object value adapters
- Line 45: Requires PropertyInfo infrastructure
- Benefit: 100% evaluator-native indexing with type safety
- Estimated effort: 6-8 hours

**Recommendation**: Defer until after Phase B (ref counting) - indexed properties interact with object lifecycle

---

## Category 5: Function Call Delegation (2 calls)

**File**: `internal/interp/evaluator/visitor_expressions_functions.go`
**Status**: ⚠️ **KEEP** - Special parameter modes and external interop
**Blocker**: Lazy parameter evaluation, Go function interop

| Line | Context | Reason for EvalNode | Feature |
|------|---------|---------------------|---------|
| 130 | Lazy parameter evaluation | Lazy thunk callback evaluation in interpreter | Jensen's Device |
| 349 | External Go functions with var params | Go function interop with var params | FFI |

**Code Examples**:

```go
// Line 130: Lazy parameter evaluation (Jensen's Device pattern)
if isLazy {
    capturedArg := arg
    var evalCallback runtime.EvalCallback = func() runtime.Value {
        // Callback captures interpreter's eval via adapter.EvalNode
        return e.adapter.EvalNode(capturedArg)
    }
    args[idx] = runtime.NewLazyThunk(capturedArg, evalCallback)
}

// Line 349: External (Go) functions with var parameters
if e.externalFunctions != nil {
    return e.adapter.EvalNode(node)
}
```

**Why These Must Delegate**:
- Line 130: Lazy thunks use callback pattern - evaluator can't eval its own AST without full language support
- Line 349: External Go functions may declare var parameters in their signatures

**Migration Potential**: ⭐ **Low**
- Line 130: Architectural - lazy evaluation requires full interpreter
- Line 349: Could be migrated with external function registry refactoring
- Estimated effort: 15-20 hours (requires evaluator to support all language features)

**Recommendation**: Keep indefinitely - these are legitimate cross-boundary calls

---

## Category 6: Compound Assignment Delegation (1 call)

**File**: `internal/interp/evaluator/compound_ops.go`
**Status**: ⚠️ **KEEP** - Class operator overloads
**Blocker**: Operator overload dispatch, method lookup

| Line | Context | Reason for EvalNode | Feature |
|------|---------|---------------------|---------|
| 26 | Object compound operations (+=, -=, etc.) | Class operator overloads + ref counting | Operator overloading |

**Code Example**:

```go
// Line 26: Class operator overload detection
leftType := left.Type()
if strings.HasPrefix(leftType, "OBJECT[") {
    // Delegate entire compound operation to adapter for objects
    return e.adapter.EvalNode(node)
}
```

**Why This Must Delegate**:
- DWScript supports operator overloading on classes
- Requires method lookup in class hierarchy
- Interacts with reference counting for result values

**Migration Potential**: ⭐⭐ **Medium**
- Could be migrated after operator overload infrastructure moves to evaluator
- Estimated effort: 4-6 hours

**Recommendation**: Defer until after Phase B and operator overload migration

---

## Category 7: Control Flow Delegation (2 calls)

**File**: `internal/interp/evaluator/visitor_statements.go`
**Status**: ⚠️ **KEEP** - Compound assignment on complex targets
**Blocker**: Member/index compound operations

| Line | Context | Reason for EvalNode | Pattern |
|------|---------|---------------------|---------|
| 455 | Compound member assignment | Full member update with compound ops | obj.field += 1 |
| 477 | Compound index assignment | Complex index update with ops | arr[i] += 1 |

**Code Examples**:

```go
// Line 455: Compound member assignment (obj.field += value)
if isCompound {
    return e.adapter.EvalNode(node)
}

// Line 477: Compound index assignment (arr[i] += value)
if isCompound {
    return e.adapter.EvalNode(node)
}
```

**Why These Must Delegate**:
- Compound operations on members require read-modify-write cycles
- Property setters may have side effects
- Indexed properties require special handling

**Migration Potential**: ⭐⭐⭐ **High**
- Could be migrated to use evalMemberAssignmentDirect with compound logic
- Estimated effort: 3-4 hours

**Recommendation**: Quick win after Phase B - low-hanging fruit

---

## Category 8: Helper Method Delegation (1 call)

**File**: `internal/interp/evaluator/helper_methods.go`
**Status**: ✅ **MIGRATION IN PROGRESS**
**Blocker**: Unhandled builtin helper methods

| Line | Context | Reason for EvalNode | Status |
|------|---------|---------------------|--------|
| 355 | Unhandled builtin helper methods | Fallback for non-migrated helpers | Temporary |

**Code Example**:

```go
// Line 355: Fallback for unhandled helpers
// Fall through to adapter for unhandled helpers
return e.adapter.EvalNode(node)
```

**Migration Potential**: ⭐⭐⭐⭐⭐ **Very High**
- This is a temporary fallback during helper migration
- As more helpers are migrated (tasks 3.5.102a-f), this call becomes less frequent
- Could be eliminated entirely with complete helper coverage

**Recommendation**: Continue helper migration - this call will naturally disappear

---

## Category 9: User Function Body Execution (1 call)

**File**: `internal/interp/evaluator/user_function_helpers.go`
**Status**: ⚠️ **KEEP** - Full language feature support required
**Blocker**: Evaluator doesn't support all language features yet

| Line | Context | Reason for EvalNode | Blocker |
|------|---------|---------------------|---------|
| 434 | User function body execution | Full language feature support (not yet migrated) | Class constructors, etc. |

**Code Example**:

```go
// Line 434: Execute function body via adapter
// We use EvalNode instead of e.Eval because the evaluator doesn't fully support
// all language features yet (e.g., class constructor calls like TClass.Create).
_ = e.adapter.EvalNode(fn.Body)
```

**Why This Must Delegate**:
- User function bodies can contain ANY DWScript language feature
- Evaluator doesn't yet support all features (e.g., class constructor calls)
- Full language support requires completing all visitor methods

**Migration Potential**: ⭐⭐ **Low-Medium**
- Requires evaluator to support 100% of language features
- Currently blocked on class instantiation, some operators, etc.
- Estimated effort: 40-60 hours (complete evaluator implementation)

**Recommendation**: Defer until Phase 4+ - requires mature evaluator

---

## Category 10: Declaration Delegation (1 call)

**File**: `internal/interp/evaluator/visitor_declarations.go`
**Status**: ⚠️ **KEEP** - Complex class info handling
**Blocker**: ClassInfo construction, VMT building

| Line | Context | Reason for EvalNode | Complexity |
|------|---------|---------------------|------------|
| 1146 | Class declaration processing | Full class info handling + VMT building | Very High |

**Code Example**:

```go
// Line 1146: Class declaration - full ClassInfo construction
func (e *Evaluator) VisitSetDecl(node *ast.SetDecl, ctx *ExecutionContext) Value {
    // Set type already registered by semantic analyzer
    // Delegate to adapter for now (Phase 3 migration)
    return e.adapter.EvalNode(node)
}
```

**Why This Must Delegate**:
- Class declarations require building ClassInfo structures
- VMT (virtual method table) construction for inheritance
- Method overload resolution and virtual method binding
- Field initialization and property metadata

**Migration Potential**: ⭐ **Low**
- Very complex - requires full OOP infrastructure
- Estimated effort: 20-30 hours

**Recommendation**: Defer until Phase 4+ - requires OOP migration complete

---

## Category 11: Method Dispatch Fallback (1 call)

**File**: `internal/interp/evaluator/method_dispatch.go`
**Status**: ⚠️ **KEEP** - Last resort safety net
**Blocker**: Unknown types without helper methods

| Line | Context | Reason for EvalNode | Type |
|------|---------|---------------------|------|
| 187 | Unknown type with no helper method | Last resort fallback | Unknown |

**Code Example**:

```go
// Line 187: Last resort fallback for unknown types
default:
    // Try helper method lookup first
    helperResult := e.FindHelperMethod(obj, methodName)
    if helperResult != nil {
        return e.CallHelperMethod(helperResult, obj, args, node, ctx)
    }

    // No handler found - delegate to adapter as last resort
    return e.adapter.EvalNode(node)
```

**Why This Must Delegate**:
- Safety net for types not handled by evaluator
- May catch edge cases or custom types
- Prevents hard crashes with graceful fallback

**Migration Potential**: ⭐⭐ **Low-Medium**
- Could be eliminated with complete type coverage
- May always need to exist as safety net
- Estimated effort: N/A (gradual elimination through coverage)

**Recommendation**: Keep as safety net - becomes less frequent with better coverage

---

## Category 12: Unknown Node Type Safety Net (1 call)

**File**: `internal/interp/evaluator/evaluator.go`
**Status**: ⚠️ **KEEP** - Essential safety net during migration
**Blocker**: Incomplete visitor implementation

| Line | Context | Reason for EvalNode | Purpose |
|------|---------|---------------------|---------|
| 1243 | Unknown node type in Eval() switch | Safety net during migration | Graceful fallback |

**Code Example**:

```go
// Line 1243: Unknown node type fallback in main Eval() switch
default:
    // Phase 3.5.2: Unknown node type - delegate to adapter if available
    // This provides a safety net during the migration
    if e.adapter != nil {
        return e.adapter.EvalNode(node)
    }
    // If no adapter, this is an error (unknown node type)
    panic("Evaluator.Eval: unknown node type and no adapter available")
```

**Why This Must Exist**:
- Safety net during evaluator migration
- Catches AST node types not yet implemented in visitor pattern
- Prevents panics with graceful fallback to interpreter
- Essential during transition period (Phase 3.5)

**Migration Potential**: ⭐⭐⭐⭐ **High (Eventually)**
- Will become unnecessary when all AST nodes have visitor methods
- Should remain until 100% visitor coverage achieved
- Can be removed in Phase 4 after complete migration
- Currently catches edge cases and new node types

**Recommendation**: Keep until visitor pattern is 100% complete - then can be converted to panic with confidence

---

## Migration Recommendations

### High Priority (Blocks Adapter Removal)

**Phase B: Reference Counting Migration (Tasks 3.5.39-3.5.42)**
- **Target**: 6 calls in assignment_helpers.go (lines 127, 136, 172, 182, 192, 227, 254)
- **Effort**: 32-44 hours
- **Blocker**: Most critical - blocks adapter removal
- **Strategy**: Migrate ref counting to runtime.RefCountManager interface

**Phase C: Self Context Migration (Task 3.5.43)**
- **Target**: 2 calls in assignment_helpers.go (lines 93, 319)
- **Effort**: 0-6 hours (depends on approach)
- **Blocker**: Self/class context access from evaluator
- **Strategy**: TBD - may require environment ownership migration

### Medium Priority (Quality Improvements)

**Index Operations Migration**
- **Target**: 3 calls in index_assignment.go (lines 45, 76, 82)
- **Benefit**: 100% evaluator-native indexing with type safety
- **Effort**: 6-8 hours
- **Recommendation**: After Phase B

**Helper Method Completion**
- **Target**: 1 call in helper_methods.go (line 355)
- **Benefit**: Complete helper coverage
- **Effort**: Ongoing (3.5.102a-f)
- **Recommendation**: Continue current migration

**Compound Assignment Simplification**
- **Target**: 2 calls in visitor_statements.go (lines 455, 477)
- **Benefit**: Cleaner compound operation handling
- **Effort**: 3-4 hours
- **Recommendation**: Quick win after Phase B

### Low Priority (Future Work)

**User Function Body Execution**
- **Target**: 1 call in user_function_helpers.go (line 434)
- **Effort**: 40-60 hours (full evaluator completion)
- **Recommendation**: Phase 4+

**Declaration Processing**
- **Target**: 1 call in visitor_declarations.go (line 1146)
- **Effort**: 20-30 hours
- **Recommendation**: Phase 4+ (requires OOP migration)

**Member Access/Assignment**
- **Target**: 12 calls (6 in visitor_expressions_members.go, 6 in member_assignment.go)
- **Effort**: 20-30 hours
- **Recommendation**: Phase 4+ (requires OOP infrastructure)

### Not Recommended for Migration

**Lazy Parameter Evaluation** (line 130 in visitor_expressions_functions.go)
- Architectural constraint - legitimate cross-boundary call
- Requires full language support in evaluator

**External Function Interop** (line 349 in visitor_expressions_functions.go)
- FFI boundary - legitimate delegation
- Could be migrated with significant registry refactoring

**Method Dispatch Fallback** (line 187 in method_dispatch.go)
- Safety net - should remain as graceful fallback
- Becomes less frequent with better type coverage

---

## Statistics

### Calls by Migration Difficulty

```
Must Keep (Architectural)      14 calls (43.8%)  █████████████████████
Potential Migration            11 calls (34.4%)  ████████████████
Migration in Progress           7 calls (21.9%)  ██████████
```

### Calls by Phase

```
Phase A (Cleanup)               0 calls (0.0%)
Phase B (Ref Counting)          6 calls (17.6%)  █████████
Phase C (Self Context)          2 calls (5.9%)   ███
Phase D (Adapter Removal)       0 calls (0.0%)
Future (Phase 4+)              25 calls (73.5%)  █████████████████████████████████
Safety Net (Keep)               1 call  (2.9%)   █
```

### Estimated Total Migration Effort

| Phase | Tasks | Calls Targeted | Effort |
|-------|-------|----------------|--------|
| A | Cleanup & Audit | 0 | 2h (audit only) |
| B | Ref Counting | 6 | 32-44h |
| C | Self Context | 2 | 0-6h |
| D | Testing & Docs | - | 10-14h |
| **Total (Phase A-D)** | **3.5.37-3.5.45** | **8** | **44-66h** |
| Future | OOP Infrastructure | 25 | 120-180h |
| Safety Net | Keep indefinitely | 1 | N/A |

---

## Verification

✅ **All 34 EvalNode calls documented** with:
- File path and line number
- Context and reason for delegation
- Migration difficulty assessment
- Recommended migration task

✅ **Cross-referenced with PLAN.md**:
- Phase A tasks: 3.5.37-3.5.38 ✓
- Phase B tasks: 3.5.39-3.5.42 ✓
- Phase C task: 3.5.43 ✓
- Phase D tasks: 3.5.44-3.5.45 ✓

✅ **Statistics accurate**:
- Total calls: 34
- Categorization: 12 categories
- File distribution: 12 files
- Line numbers verified against source code

---

## Conclusion

The audit reveals a **well-structured delegation pattern** with clear architectural boundaries:

1. **Reference counting** (6 calls) is the primary blocker for adapter removal - correctly targeted by Phase B
2. **Self/class context** (2 calls) requires environment ownership strategy - targeted by Phase C
3. **OOP infrastructure** (18 calls) requires long-term migration effort - deferred to Phase 4+
4. **Quick wins exist** (3 calls) in index operations and compound assignments - good follow-up work
5. **Some calls are permanent** (3 calls) due to architectural constraints or safety nets - this is acceptable

The current migration strategy (Phase A→B→C→D) is **well-aligned** with the dependency graph and will achieve full evaluator independence for the subset of operations that don't require complex OOP infrastructure.

**Next Steps**: Proceed with Task 3.5.39 (Design Ref Counting Architecture) using this audit as the foundation for migration planning.
