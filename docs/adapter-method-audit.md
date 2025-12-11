# Adapter Method Usage Audit (Task 3.4.1)

**Date**: 2025-12-11
**Goal**: Audit all 67 adapter methods to identify candidates for deletion, inlining, or retention in focused interfaces.

## Executive Summary

**Total Methods**: 67
**DELETE Candidates** (0 usages): 1 method (1.5%)
**INLINE Candidates** (1-2 usages): 59 methods (88%)
**KEEP** (3+ usages): 7 methods (10.5%)

**Potential Reduction**: 1 deleted + ~30 inlined = **~31 methods eliminated**
**Remaining**: ~36 methods (distributed across 4 focused interfaces)

## Categorization

### DELETE (0 Callers - Dead Code)

| Method | Category | Reason |
|--------|----------|--------|
| `CallUserFunctionWithOverloads` | Dispatch | Unused - can be deleted immediately |

**Action**: Delete in Task 3.4.2

---

### INLINE (1-2 Callers - Move Logic to Caller)

These methods have minimal usage and should be inlined into their caller(s) to reduce interface complexity.

#### 1 Caller (47 methods)

**Helper Declaration (8 methods)**:
- `CreateHelperInfo` - visitor_declarations.go
- `SetHelperParent` - visitor_declarations.go
- `VerifyHelperTargetTypeMatch` - visitor_declarations.go
- `GetHelperName` - visitor_declarations.go
- `AddHelperMethod` - visitor_declarations.go
- `AddHelperProperty` - visitor_declarations.go
- `AddHelperClassVar` - visitor_declarations.go
- `AddHelperClassConst` - visitor_declarations.go

**Interface Declaration (7 methods)**:
- `NewInterfaceInfoAdapter` - visitor_declarations.go
- `CastToInterfaceInfo` - visitor_declarations.go
- `SetInterfaceParent` - visitor_declarations.go
- `GetInterfaceName` - visitor_declarations.go
- `GetInterfaceParent` - visitor_declarations.go
- `AddInterfaceMethod` - visitor_declarations.go
- `AddInterfaceProperty` - visitor_declarations.go

**Class Declaration (21 methods)**:
- `NewClassInfoAdapter` - visitor_declarations.go
- `CastToClassInfo` - visitor_declarations.go
- `IsClassPartial` - visitor_declarations.go
- `SetClassPartial` - visitor_declarations.go
- `SetClassAbstract` - visitor_declarations.go
- `SetClassExternal` - visitor_declarations.go
- `ClassHasNoParent` - visitor_declarations.go
- `DefineCurrentClassMarker` - visitor_declarations.go
- `SetClassParent` - visitor_declarations.go
- `AddInterfaceToClass` - visitor_declarations.go
- `SynthesizeDefaultConstructor` - visitor_declarations.go
- `AddClassProperty` - visitor_declarations.go
- `RegisterClassOperator` - visitor_declarations.go
- `SetClassConstructor` - visitor_declarations.go
- `SetClassDestructor` - visitor_declarations.go
- `InheritDestructorIfMissing` - visitor_declarations.go
- `InheritParentProperties` - visitor_declarations.go
- `BuildVirtualMethodTable` - visitor_declarations.go
- `RegisterClassInTypeSystem` - visitor_declarations.go
- `EvalClassPropertyRead` - property_read.go
- `EvalClassPropertyWrite` - property_write.go

**Object Operations (5 methods)**:
- `ExecuteConstructor` - visitor_expressions_identifiers.go
- `RaiseTypeCastException` - visitor_expressions_identifiers.go
- `RaiseAssertionFailed` - visitor_statements.go
- `CreateContractException` - visitor_statements.go
- `CleanupInterfaceReferences` - visitor_statements.go

**Method/Function Calls (4 methods)**:
- `CallInheritedMethod` - visitor_expressions_methods.go
- `EvalMethodImplementation` - visitor_declarations.go
- `CallQualifiedOrConstructor` - visitor_expressions_functions.go
- `CallExternalFunction` - visitor_expressions_functions.go

**Exception Handling (2 methods)**:
- `CreateExceptionDirect` - visitor_statements.go
- `WrapObjectInException` - visitor_statements.go

**Variable Declaration (2 methods)**:
- `WrapInSubrange` - visitor_declarations.go
- `WrapInInterface` - visitor_declarations.go

**Dispatch Methods (2 methods)**:
- `CallRecordStaticMethod` - visitor_expressions_functions.go
- `DispatchRecordStaticMethod` - visitor_expressions_functions.go

**Class Lookup (1 method)**:
- `LookupClassByName` - visitor_expressions_identifiers.go

**Operator Overloading (1 method)**:
- `TryUnaryOperator` - visitor_expressions.go

#### 2 Callers (12 methods)

**Helper Declaration (1 method)**:
- `RegisterHelperLegacy` (2) - visitor_declarations.go (both calls)

**Method Calls (1 method)**:
- `CallMethod` (2) - visitor_expressions_methods.go, visitor_expressions_identifiers.go

**Object Operations (2 methods)**:
- `CreateTypeCastWrapper` (2) - visitor_expressions_identifiers.go (both calls)

**Method Pointers (2 methods)**:
- `ExecuteFunctionPointerCall` (2) - visitor_expressions_functions.go (both calls)
- `CreateBoundMethodPointer` (2) - visitor_expressions_identifiers.go, visitor_expressions_methods.go

**Class Declaration (2 methods)**:
- `AddClassMethod` (2) - visitor_declarations.go (both calls)
- `LookupClassMethod` (2) - visitor_declarations.go (both calls)

**Dispatch Methods (2 methods)**:
- `ExecuteRecordPropertyRead` (2) - property_read.go (both calls)

**Core Execution (2 methods)**:
- `CallFunctionPointer` (2) - visitor_expressions_identifiers.go, visitor_expressions_functions.go
- `CallUserFunction` (2) - visitor_expressions_functions.go (both calls)

**Operator Overloading (1 method)**:
- `TryBinaryOperator` (2) - binary_ops.go (both calls)

---

### KEEP (3+ Callers - Retain in Focused Interface)

These methods have significant usage and should be kept in the focused interfaces.

| Method | Count | Category | Files | Notes |
|--------|-------|----------|-------|-------|
| `EvalNode` | 6* | Core | method_dispatch.go (2), user_function_helpers.go (1), helper_methods.go (1), evaluator.go (1), compound_assignments.go (1) | *Excluding 5 test file references |
| `ExecuteMethodWithSelf` | 10 | OOP | index_assignment.go (4), property_write.go (2), property_read.go (2), visitor_expressions_methods.go (1), visitor_expressions_identifiers.go (1) | Property access with setters/getters |
| `EvalBuiltinHelperProperty` | 4 | Helper | helper_methods.go (all 4) | Built-in helper property evaluation |
| `CallImplicitSelfMethod` | 3 | Dispatch | visitor_expressions_functions.go (2), visitor_expressions_identifiers.go (1) | Self context method calls |

---

## Analysis by Interface Category

### OOPEngine (Runtime OOP Operations) - Target: ~18 methods

**KEEP (high usage)**:
- ExecuteMethodWithSelf (10) - property setters/getters
- CallMethod (2) - method dispatch
- CallImplicitSelfMethod (3) - self context

**INLINE/DELETE (low usage)**:
- CallInheritedMethod (1) - inline
- ExecuteConstructor (1) - inline
- CreateTypeCastWrapper (2) - inline
- RaiseTypeCastException (1) - inline
- RaiseAssertionFailed (1) - inline
- CallFunctionPointer (2) - inline
- CallUserFunction (2) - inline
- ExecuteFunctionPointerCall (2) - inline
- CreateBoundMethodPointer (2) - inline
- CallQualifiedOrConstructor (1) - inline
- CallUserFunctionWithOverloads (0) - **DELETE**
- CallRecordStaticMethod (1) - inline
- DispatchRecordStaticMethod (1) - inline
- ExecuteRecordPropertyRead (2) - inline
- CallExternalFunction (1) - inline
- TryBinaryOperator (2) - inline
- TryUnaryOperator (1) - inline

**Recommendation**: Keep only 3 high-usage methods, inline the rest (14 methods eliminated)

---

### DeclHandler (Declaration Processing) - Target: ~37 methods

**KEEP (moderate usage)**:
- None - all declaration methods have 1-2 usages

**INLINE/DELETE (all low usage)**:
- All 36 class/interface/helper declaration methods have 1 caller (visitor_declarations.go)
- EvalMethodImplementation (1) - inline

**Recommendation**: Since all 37 methods are called from a single file (visitor_declarations.go), consider:
1. **Option A**: Keep interface but inline simple wrappers
2. **Option B**: Make visitor_declarations.go call Interpreter methods directly (eliminate interface entirely)
3. **Option C**: Keep minimal interface with ~10 core methods, inline the rest

**Estimated reduction**: ~27-37 methods eliminated (depending on option)

---

### ExceptionManager (Exception Handling) - Target: ~6 methods

**INLINE/DELETE (all low usage)**:
- CreateExceptionDirect (1) - inline
- WrapObjectInException (1) - inline
- CreateContractException (1) - inline
- RaiseTypeCastException (1) - inline
- RaiseAssertionFailed (1) - inline
- CleanupInterfaceReferences (1) - inline

**Recommendation**: All 6 methods have single callers. Consider eliminating interface entirely and calling Interpreter methods directly.

**Estimated reduction**: ~6 methods eliminated

---

### CoreEvaluator (Fallback to Interpreter) - Target: ~3 methods

**KEEP (high/moderate usage)**:
- EvalNode (6) - used for OOP fallback, may be eliminated in future
- EvalBuiltinHelperProperty (4) - built-in helper properties

**INLINE (low usage)**:
- EvalMethodImplementation (1) - inline

**Recommendation**: Keep EvalNode and EvalBuiltinHelperProperty for now. These represent cross-cutting concerns that may require architectural changes to eliminate.

**Estimated reduction**: 1 method inlined

---

## Migration Strategy

### Phase 1: Quick Wins (Task 3.4.2)

1. **Delete dead code** (immediate):
   - CallUserFunctionWithOverloads (0 usages)

2. **Inline trivial wrappers** (high confidence):
   - All 36 class/interface/helper declaration methods (1 caller each in visitor_declarations.go)
   - All 6 exception methods (1 caller each)
   - ~15 OOP methods (1-2 callers)

**Estimated impact**: ~57 methods eliminated, ~10 methods kept

### Phase 2: Interface Design (Task 3.4.3-3.4.5)

Create focused interfaces with remaining methods:

**OOPEngine** (~3-5 methods):
- ExecuteMethodWithSelf
- CallMethod
- CallImplicitSelfMethod
- (possibly EvalNode if needed)

**DeclHandler** (~0-10 methods):
- If kept: minimal core methods only
- If eliminated: direct Interpreter calls from visitor_declarations.go

**ExceptionManager** (~0 methods):
- Likely eliminated - direct Interpreter calls

**CoreEvaluator** (~2 methods):
- EvalNode (may be eliminated in Phase 3.5)
- EvalBuiltinHelperProperty

### Phase 3: Final Cleanup (Task 3.4.6-3.4.9)

1. Migrate adapter calls to focused interfaces
2. Delete InterpreterAdapter interface
3. Verify and test

---

## Success Metrics

| Metric | Before | Target | Strategy |
|--------|--------|--------|----------|
| Total Methods | 67 | ~10-15 | Delete + Inline |
| DELETE | 0 | 1 | CallUserFunctionWithOverloads |
| INLINE | 0 | ~50-57 | Single-caller wrappers |
| KEEP | 67 | ~10-15 | High-usage core methods |
| Interfaces | 1 | 2-4 | OOPEngine, CoreEvaluator (DeclHandler/ExceptionManager likely eliminated) |

---

## Files by Adapter Call Count

Based on the audit, here are the files with the most adapter calls:

| File | Adapter Calls | Notes |
|------|---------------|-------|
| visitor_declarations.go | ~43 | All class/interface/helper declaration (candidates for inlining) |
| visitor_expressions_functions.go | ~11 | Function calls and dispatch |
| index_assignment.go | ~4 | Property setters with indices |
| helper_methods.go | ~4 | Helper property evaluation |
| property_write.go | ~2 | Property setters |
| property_read.go | ~2 | Property getters |
| visitor_statements.go | ~6 | Exception handling |
| visitor_expressions_methods.go | ~2 | Method calls |
| visitor_expressions_identifiers.go | ~4 | Identifier resolution |

**Key Insight**: visitor_declarations.go accounts for 64% of adapter calls (43/67), all with single usages. This file is the primary candidate for simplification.

---

## Recommendations for Task 3.4.2

1. **Delete immediately**:
   - CallUserFunctionWithOverloads (0 usages)

2. **Inline high-confidence** (single caller in visitor_declarations.go):
   - All 36 class/interface/helper declaration methods
   - Consider creating helper functions in visitor_declarations.go to reduce repetition

3. **Inline medium-confidence** (single caller, simple wrappers):
   - All 6 exception methods
   - ~10-15 OOP methods with trivial implementations

4. **Keep for focused interfaces**:
   - ExecuteMethodWithSelf (10 usages)
   - EvalNode (6 usages) - may be eliminated later
   - EvalBuiltinHelperProperty (4 usages)
   - CallImplicitSelfMethod (3 usages)

**Estimated effort**: 4-6 hours for Task 3.4.2 (delete + inline ~50 methods)

---

## Appendix: Complete Method List with Usage Counts

| # | Method | Count | Category | Action |
|---|--------|-------|----------|--------|
| 1 | EvalNode | 6 | Core | KEEP |
| 2 | CallFunctionPointer | 2 | Core | INLINE |
| 3 | CallUserFunction | 2 | Core | INLINE |
| 4 | EvalMethodImplementation | 1 | Decl | INLINE |
| 5 | CreateHelperInfo | 1 | Helper Decl | INLINE |
| 6 | SetHelperParent | 1 | Helper Decl | INLINE |
| 7 | VerifyHelperTargetTypeMatch | 1 | Helper Decl | INLINE |
| 8 | GetHelperName | 1 | Helper Decl | INLINE |
| 9 | AddHelperMethod | 1 | Helper Decl | INLINE |
| 10 | AddHelperProperty | 1 | Helper Decl | INLINE |
| 11 | AddHelperClassVar | 1 | Helper Decl | INLINE |
| 12 | AddHelperClassConst | 1 | Helper Decl | INLINE |
| 13 | RegisterHelperLegacy | 2 | Helper Decl | INLINE |
| 14 | NewInterfaceInfoAdapter | 1 | Iface Decl | INLINE |
| 15 | CastToInterfaceInfo | 1 | Iface Decl | INLINE |
| 16 | SetInterfaceParent | 1 | Iface Decl | INLINE |
| 17 | GetInterfaceName | 1 | Iface Decl | INLINE |
| 18 | GetInterfaceParent | 1 | Iface Decl | INLINE |
| 19 | AddInterfaceMethod | 1 | Iface Decl | INLINE |
| 20 | AddInterfaceProperty | 1 | Iface Decl | INLINE |
| 21 | CallMethod | 2 | OOP | INLINE |
| 22 | CallInheritedMethod | 1 | OOP | INLINE |
| 23 | ExecuteMethodWithSelf | 10 | OOP | KEEP |
| 24 | ExecuteConstructor | 1 | OOP | INLINE |
| 25 | CreateTypeCastWrapper | 2 | OOP | INLINE |
| 26 | RaiseTypeCastException | 1 | Exception | INLINE |
| 27 | RaiseAssertionFailed | 1 | Exception | INLINE |
| 28 | CreateContractException | 1 | Exception | INLINE |
| 29 | CleanupInterfaceReferences | 1 | Exception | INLINE |
| 30 | ExecuteFunctionPointerCall | 2 | OOP | INLINE |
| 31 | CreateExceptionDirect | 1 | Exception | INLINE |
| 32 | WrapObjectInException | 1 | Exception | INLINE |
| 33 | WrapInSubrange | 1 | VarDecl | INLINE |
| 34 | WrapInInterface | 1 | VarDecl | INLINE |
| 35 | CreateBoundMethodPointer | 2 | OOP | INLINE |
| 36 | CallQualifiedOrConstructor | 1 | Dispatch | INLINE |
| 37 | CallUserFunctionWithOverloads | 0 | Dispatch | DELETE |
| 38 | CallImplicitSelfMethod | 3 | Dispatch | KEEP |
| 39 | CallRecordStaticMethod | 1 | Dispatch | INLINE |
| 40 | DispatchRecordStaticMethod | 1 | Dispatch | INLINE |
| 41 | ExecuteRecordPropertyRead | 2 | Dispatch | INLINE |
| 42 | CallExternalFunction | 1 | Dispatch | INLINE |
| 43 | NewClassInfoAdapter | 1 | Class Decl | INLINE |
| 44 | CastToClassInfo | 1 | Class Decl | INLINE |
| 45 | IsClassPartial | 1 | Class Decl | INLINE |
| 46 | SetClassPartial | 1 | Class Decl | INLINE |
| 47 | SetClassAbstract | 1 | Class Decl | INLINE |
| 48 | SetClassExternal | 1 | Class Decl | INLINE |
| 49 | ClassHasNoParent | 1 | Class Decl | INLINE |
| 50 | DefineCurrentClassMarker | 1 | Class Decl | INLINE |
| 51 | SetClassParent | 1 | Class Decl | INLINE |
| 52 | AddInterfaceToClass | 1 | Class Decl | INLINE |
| 53 | AddClassMethod | 2 | Class Decl | INLINE |
| 54 | SynthesizeDefaultConstructor | 1 | Class Decl | INLINE |
| 55 | AddClassProperty | 1 | Class Decl | INLINE |
| 56 | RegisterClassOperator | 1 | Class Decl | INLINE |
| 57 | LookupClassMethod | 2 | Class Decl | INLINE |
| 58 | SetClassConstructor | 1 | Class Decl | INLINE |
| 59 | SetClassDestructor | 1 | Class Decl | INLINE |
| 60 | InheritDestructorIfMissing | 1 | Class Decl | INLINE |
| 61 | InheritParentProperties | 1 | Class Decl | INLINE |
| 62 | BuildVirtualMethodTable | 1 | Class Decl | INLINE |
| 63 | RegisterClassInTypeSystem | 1 | Class Decl | INLINE |
| 64 | EvalBuiltinHelperProperty | 4 | Helper | KEEP |
| 65 | EvalClassPropertyRead | 1 | Class | INLINE |
| 66 | EvalClassPropertyWrite | 1 | Class | INLINE |
| 67 | TryBinaryOperator | 2 | Operator | INLINE |
| 68 | TryUnaryOperator | 1 | Operator | INLINE |
| 69 | LookupClassByName | 1 | Class | INLINE |

**Totals**:
- DELETE: 1
- INLINE: 62
- KEEP: 6

**Projected Adapter Size**: ~6-10 methods (if we keep some inline candidates for interface clarity)
