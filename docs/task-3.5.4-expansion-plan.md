# Task 3.5.4 Expansion Plan: Unblocking Complex Type Dependencies

## Current Status

**Progress**: 23/48 methods migrated (47.9%)
**Remaining**: 25 methods (52.1%)

**Completed**:
- ✅ Phase 1: Simple value types migrated (EnumValue, TypeMetaValue, ArrayValue, SetValue)
- ✅ Phase 2D: Loop statements (ForStatement, ForInStatement)
- ✅ All 6 literal visitors
- ✅ 13 statement visitors (control flow, loops, etc.)
- ✅ IfExpression (just completed)

**Blockers**:
- Complex type dependencies (ExceptionValue, ObjectInstance, ClassInfo, FunctionPointerValue, RecordValue)
- Type system access needed (registries for classes, records, interfaces, operators, etc.)
- Missing infrastructure methods

## Remaining Methods Categorized

### 🟢 Category A: Quick Wins (5 methods - can be done immediately)

These methods need minimal or no additional infrastructure:

#### A1. Simple Expression Evaluation
1. **VisitOldExpression** - Postcondition support
   - Infrastructure: Add `GetOldValue(name string) (Value, bool)` to ExecutionContext
   - Effort: 30 minutes
   - No type dependencies

#### A2. Environment Modification
2. **VisitVarDeclStatement** - Variable declarations
   - Uses: `ctx.Env().Set(name, value)`
   - Effort: 1 hour
   - Dependencies: None (infrastructure exists)

3. **VisitConstDecl** - Constant declarations
   - Uses: `ctx.Env().Set(name, value)` (same as var)
   - Effort: 30 minutes
   - Dependencies: None

4. **VisitExpressionStatement** - Expression statements
   - Just evaluate expression and discard result
   - Special case: function pointer auto-invocation (needs adapter check)
   - Effort: 1 hour
   - Dependencies: FunctionPointerValue check via adapter

#### A3. Simple Literal Construction
5. **VisitSetLiteral** - Set literal expressions `[1, 2, 3]`
   - Uses: `runtime.SetValue` (already migrated)
   - Infrastructure: Element evaluation + set construction
   - Effort: 2 hours
   - Dependencies: SetValue (✅ already in runtime)

**Total Category A**: ~5 hours of work, +5 methods (28/48 = 58%)

---

### 🟡 Category B: Medium Complexity (7 methods - need adapter extensions)

These methods can be migrated by extending the InterpreterAdapter interface:

#### B1. Operator Evaluation
6. **VisitBinaryExpression** - Binary operators (+, -, *, /, =, <>, etc.)
   - Adapter method: `EvalBinaryOp(op string, left Value, right Value) Value`
   - Needs: Operator overload registry, type coercion
   - Effort: 3 hours (adapter) + 2 hours (migration)

7. **VisitUnaryExpression** - Unary operators (-, not)
   - Adapter method: `EvalUnaryOp(op string, operand Value) Value`
   - Needs: Operator overload registry
   - Effort: 2 hours (adapter) + 1 hour (migration)

#### B2. Variable Lookup
8. **VisitIdentifier** - Variable/constant/function references
   - Adapter methods: Already exist (`LookupFunction`, `LookupClass`)
   - Additional: Check if identifier is a type name (TClass reference)
   - Effort: 2 hours

#### B3. Array Operations (leverage existing runtime types)
9. **VisitArrayLiteralExpression** - Array construction `[1, 2, 3]`
   - Adapter method: `InferArrayElementType(elements []Value) types.Type`
   - Uses: `runtime.ArrayValue` (✅ already migrated)
   - Effort: 3 hours

10. **VisitNewArrayExpression** - Dynamic array allocation `new array[10] of Integer`
    - Adapter method: `CreateArray(elementType types.Type, size int) Value`
    - Uses: `runtime.ArrayValue`
    - Effort: 2 hours

11. **VisitIndexExpression** - Array/string indexing `arr[0]`
    - Adapter methods:
      - `GetArrayElement(arr Value, index Value) Value`
      - `GetPropertyIndex(obj Value, index Value) Value` (for indexed properties)
    - Effort: 3 hours

#### B4. Property/Field Access
12. **VisitMemberAccessExpression** - Field/property access `obj.field`
    - Adapter methods:
      - `GetObjectField(obj Value, fieldName string) Value`
      - `GetPropertyValue(obj Value, propName string) Value`
      - `GetRecordField(rec Value, fieldName string) Value`
    - Effort: 4 hours

**Total Category B**: ~22 hours of work, +7 methods (35/48 = 73%)

---

### 🔴 Category C: High Complexity (13 methods - need significant infrastructure)

These methods require complex type operations or major infrastructure:

#### C1. Object-Oriented Features (6 methods)
13. **VisitNewExpression** - Object instantiation `TMyClass.Create()`
14. **VisitMethodCallExpression** - Method calls `obj.Method()`
15. **VisitInheritedExpression** - Parent method calls `inherited;`
16. **VisitIsExpression** - Type checking `obj is TMyClass`
17. **VisitAsExpression** - Type casting `obj as TMyClass`
18. **VisitImplementsExpression** - Interface checking `obj implements IMyInterface`

**Blocker**: ObjectInstance, ClassInfo need to stay in internal/interp (AST dependencies)

**Adapter approach**:
- `CreateObject(className string, args []Value) Value`
- `CallMethod(obj Value, methodName string, args []Value, node ast.Node) Value`
- `CallInherited(obj Value, methodName string, args []Value) Value`
- `CheckType(obj Value, typeName string) bool`
- `CastType(obj Value, typeName string) Value`
- `CheckImplements(obj Value, interfaceName string) bool`

**Effort**: ~15 hours (adapter methods) + ~8 hours (migrations)

#### C2. Function Pointers & Lambdas (3 methods)
19. **VisitCallExpression** - Function calls `MyFunc()`
20. **VisitAddressOfExpression** - Function pointer creation `@MyFunc`
21. **VisitLambdaExpression** - Lambda/closure `lambda(x) x * 2`

**Blocker**: FunctionPointerValue has AST dependencies

**Adapter approach**:
- Existing: `CallFunctionPointer(funcPtr Value, args []Value, node ast.Node) Value`
- Existing: `CallUserFunction(fn *ast.FunctionDecl, args []Value) Value`
- New: `CreateFunctionPointer(fn *ast.FunctionDecl, closure Environment) Value`
- New: `CreateLambda(lambda *ast.LambdaExpression, closure Environment) Value`

**Effort**: ~8 hours (adapter methods) + ~5 hours (migrations)

#### C3. Record Operations (1 method)
22. **VisitRecordLiteralExpression** - Record construction `TMyRecord(Field1: 1, Field2: 'test')`

**Blocker**: RecordValue has AST dependencies (methods)

**Adapter approach**:
- `CreateRecord(recordType string, fields map[string]Value) Value`

**Effort**: ~2 hours (adapter) + ~2 hours (migration)

#### C4. Exception Handling (2 methods)
23. **VisitTryStatement** - Try-except-finally blocks
24. **VisitRaiseStatement** - Raise exceptions

**Blocker**: ExceptionValue has ClassInfo/ObjectInstance dependencies

**Adapter approach**:
- `CreateException(className string, message string) Value`
- `GetExceptionMessage(exc Value) string`
- `GetExceptionClass(exc Value) string`

**Effort**: ~4 hours (adapter) + ~3 hours (migrations)

#### C5. Assignment Operations (1 method)
25. **VisitAssignmentStatement** - Assignment with lvalue resolution

**Needs**: Lvalue resolution (fields, properties, array indices, etc.)

**Adapter approach**:
- `SetVariable(name string, value Value, ctx *ExecutionContext) error`
- `SetField(obj Value, fieldName string, value Value) error`
- `SetArrayElement(arr Value, index Value, value Value) error`
- `SetProperty(obj Value, propName string, value Value) error`

**Effort**: ~5 hours (adapter) + ~3 hours (migration)

**Total Category C**: ~47 hours of work, +13 methods (48/48 = 100%)

---

### 🟣 Category D: Type Declarations (9 methods - deferred)

These are registration/declaration methods that can be migrated last:

26-34. **All declaration visitors**: FunctionDecl, ClassDecl, InterfaceDecl, RecordDecl, EnumDecl, HelperDecl, OperatorDecl, ArrayDecl, TypeDeclaration

**Approach**: These methods register types/functions into the TypeSystem. They can be migrated by:
1. Exposing TypeSystem registration methods via adapter, OR
2. Migrating them as final cleanup step (low priority)

**Effort**: ~10 hours total

**Note**: These are marked as out of the 48-method count in current planning.

---

## Proposed Phased Approach

### 📅 Phase 2.1: Quick Wins (Week 1 - 5 hours)

**Goal**: Reach 28/48 (58%) with minimal effort

**Tasks**:
1. Add `GetOldValue()` method to ExecutionContext
2. Migrate VisitOldExpression
3. Migrate VisitVarDeclStatement
4. Migrate VisitConstDecl
5. Migrate VisitExpressionStatement (with adapter check for function pointers)
6. Migrate VisitSetLiteral

**Deliverable**: +5 methods migrated, 58% complete

---

### 📅 Phase 2.2: Adapter Extensions - Basic Operations (Week 2 - 22 hours)

**Goal**: Reach 35/48 (73%) by extending adapter for common operations

**Tasks**:
1. Add adapter methods for operators:
   - `EvalBinaryOp(op string, left Value, right Value) Value`
   - `EvalUnaryOp(op string, operand Value) Value`

2. Add adapter methods for arrays:
   - `InferArrayElementType(elements []Value) types.Type`
   - `CreateArray(elementType types.Type, size int) Value`
   - `GetArrayElement(arr Value, index Value) Value`

3. Add adapter methods for property/field access:
   - `GetObjectField(obj Value, fieldName string) Value`
   - `GetPropertyValue(obj Value, propName string) Value`
   - `GetRecordField(rec Value, fieldName string) Value`
   - `GetPropertyIndex(obj Value, index Value) Value`

4. Migrate methods:
   - VisitIdentifier
   - VisitBinaryExpression
   - VisitUnaryExpression
   - VisitArrayLiteralExpression
   - VisitNewArrayExpression
   - VisitIndexExpression
   - VisitMemberAccessExpression

**Deliverable**: +7 methods migrated, 73% complete

---

### 📅 Phase 2.3: Adapter Extensions - OOP & Advanced Features (Week 3-4 - 47 hours)

**Goal**: Reach 48/48 (100%) by completing all complex migrations

**Tasks**:
1. Add adapter methods for OOP:
   - `CreateObject(className string, args []Value) Value`
   - `CallMethod(obj Value, methodName string, args []Value, node ast.Node) Value`
   - `CallInherited(obj Value, methodName string, args []Value) Value`
   - `CheckType(obj Value, typeName string) bool`
   - `CastType(obj Value, typeName string) Value`
   - `CheckImplements(obj Value, interfaceName string) bool`

2. Add adapter methods for function pointers:
   - `CreateFunctionPointer(fn *ast.FunctionDecl, closure Environment) Value`
   - `CreateLambda(lambda *ast.LambdaExpression, closure Environment) Value`

3. Add adapter methods for records:
   - `CreateRecord(recordType string, fields map[string]Value) Value`

4. Add adapter methods for exceptions:
   - `CreateException(className string, message string) Value`
   - `GetExceptionMessage(exc Value) string`
   - `GetExceptionClass(exc Value) string`

5. Add adapter methods for assignments:
   - `SetVariable(name string, value Value, ctx *ExecutionContext) error`
   - `SetField(obj Value, fieldName string, value Value) error`
   - `SetArrayElement(arr Value, index Value, value Value) error`
   - `SetProperty(obj Value, propName string, value Value) error`

6. Migrate all remaining methods (13 methods)

**Deliverable**: +13 methods migrated, 100% complete

---

## Alternative: Parallel Track for Long-term Solution

While doing the adapter-based approach (pragmatic, gets to 100%), we can explore in parallel:

### 🔬 Research Track: AST-Free Runtime Types

**Goal**: Eliminate AST dependencies from runtime value types

**Approach**:
1. Analyze what AST information is actually needed at runtime
2. Extract minimal metadata (method signatures, field types) at compile time
3. Store metadata in runtime-friendly format (no AST pointers)
4. Separate "class definition" (compile-time) from "class metadata" (runtime)

**Benefits**:
- Clean architecture (no circular dependencies)
- Smaller memory footprint
- Faster runtime (no AST tree walking)
- Enables bytecode compilation improvements

**Timeline**: 4-6 weeks parallel research

**Risk**: Medium (may uncover deeper architectural issues)

---

## Recommended Action Plan

### Immediate (This Week): Phase 2.1 Quick Wins
- **Effort**: 5 hours
- **Impact**: +5 methods (58% complete)
- **Risk**: Low
- **Blockers**: None

### Short-term (Next 2 Weeks): Phase 2.2 Adapter Extensions
- **Effort**: 22 hours
- **Impact**: +7 methods (73% complete)
- **Risk**: Low
- **Blockers**: None

### Medium-term (Weeks 3-4): Phase 2.3 Complete Migration
- **Effort**: 47 hours
- **Impact**: +13 methods (100% complete)
- **Risk**: Medium (complex adapter methods)
- **Blockers**: Thorough testing needed

### Long-term (Parallel): AST-Free Runtime Research
- **Effort**: 4-6 weeks research
- **Impact**: Architectural improvement
- **Risk**: Medium
- **Decision Point**: After Phase 2.3 completion

---

## Success Metrics

### Phase 2.1 (Week 1)
- ✅ 28/48 methods migrated (58%)
- ✅ All tests passing
- ✅ No performance regression

### Phase 2.2 (Week 2-3)
- ✅ 35/48 methods migrated (73%)
- ✅ Adapter interface has ~15 methods
- ✅ All tests passing

### Phase 2.3 (Week 4-5)
- ✅ 48/48 methods migrated (100%)
- ✅ Adapter interface complete (~30 methods)
- ✅ All tests passing
- ✅ Task 3.5.4 complete
- ⚠️ Task 3.5.5 deferred (adapter removal requires architectural refactoring)

---

## Next Steps

1. **Approve expansion plan** ✓
2. **Start Phase 2.1** - Add GetOldValue to ExecutionContext
3. **Migrate Quick Win methods** (5 methods)
4. **Update PLAN.md** with phased breakdown
5. **Commit and push** progress

---

## Notes

- The adapter pattern is a pragmatic solution that allows 100% completion
- It contradicts task 3.5.5 (adapter removal) but that becomes a separate architectural task
- Long-term solution (AST-free types) can be pursued after proving value with adapter approach
- Each phase can be committed incrementally with tests passing
