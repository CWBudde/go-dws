# Evaluator Type Migration Strategy

## Problem Statement

Task 3.5.4 has 27 methods remaining (56.3% incomplete) that are blocked by type dependencies. These methods cannot be fully migrated to the Evaluator package because they require access to complex value types that are currently in `internal/interp` and cannot be imported by the evaluator due to circular dependency constraints.

## Blocking Type Dependencies

### Current Situation

The evaluator package structure:
```
internal/interp/               (parent package)
  ├── value.go                 (ArrayValue, SetValue, EnumValue, etc.)
  ├── class.go                 (ClassInfo, ObjectInstance)
  ├── exceptions.go            (ExceptionValue)
  └── evaluator/               (child package)
      ├── evaluator.go
      └── visitor_*.go
```

**Problem**: Child packages cannot import from their parent package in Go, creating a circular dependency.

### Blocked Methods by Type Dependency

**ArrayValue, SetValue, EnumValue, TypeMetaValue** (4 types) block:
- ForInStatement (needs to iterate over collections)
- IndexExpression (needs array element access)
- ArrayLiteralExpression (needs array construction)
- SetLiteral (needs set construction)
- NewArrayExpression (needs dynamic array creation)

**ExceptionValue, ObjectInstance, ClassInfo** (3 types) block:
- TryStatement (needs exception matching)
- RaiseStatement (needs exception creation)
- InheritedExpression (needs object/class hierarchy)
- MethodCallExpression (needs method dispatch)
- NewExpression (needs object instantiation)
- IsExpression (needs type checking)
- AsExpression (needs type casting)
- MemberAccessExpression (needs property/field access)

**FunctionPointerValue** (1 type) blocks:
- AddressOfExpression (needs function pointer creation)
- CallExpression (needs function pointer invocation)
- LambdaExpression (needs closure capture)
- ExpressionStatement (needs function call handling)

**RecordValue** (1 type) blocks:
- RecordLiteralExpression (needs record construction)
- MemberAccessExpression (needs record field access)

**Type Registries** (Interpreter fields) block:
- All declaration visitor methods (9 methods)
- BinaryExpression, UnaryExpression (operator overloading)
- Identifier (function/class lookup)
- Various type-dependent operations

## Dependency Analysis

### Simple Types (Migratable)

These types have minimal dependencies and can be moved to runtime/:

1. **EnumValue**
   - Dependencies: None (just 3 fields: TypeName, ValueName, OrdinalValue)
   - Impact: Partially unblocks ForInStatement
   - Status: ✅ Ready to migrate

2. **TypeMetaValue**
   - Dependencies: `types.Type` interface (already importable)
   - Impact: Partially unblocks ForInStatement
   - Status: ✅ Ready to migrate

3. **ArrayValue**
   - Dependencies: `types.ArrayType`, `[]Value`
   - Impact: Unblocks ForInStatement (array iteration), IndexExpression, ArrayLiteralExpression
   - Status: ✅ Ready to migrate (Value is already in runtime via interface)

4. **SetValue**
   - Dependencies: `types.SetType`, map/slice for storage
   - Impact: Partially unblocks ForInStatement (set iteration)
   - Status: ✅ Ready to migrate
   - Note: Needs HasElement() method accessible

### Complex Types (Not Migratable Yet)

These types have heavy dependencies that create circular imports:

5. **RecordValue**
   - Dependencies: `types.RecordType`, `map[string]Value`, `map[string]*ast.FunctionDecl`
   - Blocker: `ast.FunctionDecl` (cannot import ast in runtime)
   - Status: ❌ Blocked

6. **ObjectInstance**
   - Dependencies: `*ClassInfo`, `map[string]Value`
   - Blocker: ClassInfo (heavy ast dependencies)
   - Status: ❌ Blocked by ClassInfo

7. **ClassInfo**
   - Dependencies: 15+ map fields with `*ast.FunctionDecl`, `*ast.FieldDecl`, `*ast.ConstDecl`, etc.
   - Blocker: Heavy ast.* dependencies throughout
   - Status: ❌ Blocked (fundamental architectural issue)

8. **FunctionPointerValue**
   - Dependencies: `*ast.FunctionDecl`, `*ast.LambdaExpression`, `*Environment`, `Value`, `*types.FunctionPointerType`
   - Blocker: ast.* types and Environment
   - Status: ❌ Blocked

9. **ExceptionValue**
   - Dependencies: `*ClassInfo`, `*ObjectInstance`, `lexer.Position`, `errors.StackTrace`
   - Blocker: ClassInfo and ObjectInstance
   - Status: ❌ Blocked by ClassInfo/ObjectInstance

## Migration Strategy

### Phase 1: Migrate Simple Value Types (Recommended)

**Goal**: Migrate the 4 simple types to unlock partial functionality.

**Types to migrate**:
1. EnumValue → `runtime/enum.go`
2. TypeMetaValue → `runtime/type_meta.go`
3. ArrayValue → `runtime/array.go`
4. SetValue → `runtime/set.go`

**Steps**:
1. Create new files in `internal/interp/runtime/` for each type
2. Move type definitions and methods
3. Add type aliases in `internal/interp/value.go` for backward compatibility
4. Update imports throughout codebase
5. Implement ForInStatement in evaluator for array/string/enum/set iteration
6. Run tests to verify no regressions

**Benefits**:
- Unblocks 5 evaluator methods (ForInStatement, IndexExpression, ArrayLiteralExpression, SetLiteral, NewArrayExpression)
- Progress: 21/48 → 26/48 (54.2%)
- Cleaner architecture with value types in runtime layer
- No circular dependency issues

**Effort**: 2-3 days

### Phase 2: Interface-Based Workaround for Complex Types

For types that cannot be migrated due to ast.* dependencies, use adapter pattern:

**Approach**:
1. Keep complex types in `internal/interp`
2. Extend InterpreterAdapter interface with type-specific methods:
   - `GetObjectClass(obj Value) (className string, exists bool)`
   - `CallObjectMethod(obj Value, methodName string, args []Value) Value`
   - `CreateException(className string, message string) Value`
   - etc.
3. Implement these adapters in Interpreter
4. Use in evaluator methods

**Benefits**:
- Unblocks remaining evaluator methods without moving types
- Progress: 26/48 → 48/48 (100%)
- Maintains separation of concerns

**Drawbacks**:
- Adapter pattern remains (contradicts task 3.5.5 goal)
- Not a "clean" architecture solution

**Effort**: 3-4 days

### Phase 3: Alternative Architectures (Future Consideration)

Long-term solutions requiring larger refactoring:

**Option A: AST-Free Value Types**
- Remove ast.* references from value types
- Store only minimal metadata needed at runtime
- Separate compilation-time data from runtime data

**Option B: Unified Package Structure**
- Merge interp and evaluator into single package
- Eliminates circular dependency issue
- Sacrifices modular organization

**Option C: Dependency Inversion**
- Create runtime-level interfaces for AST operations
- Have interp implement these interfaces
- Runtime depends on abstractions, not concrete AST types

## Recommendation

**Immediate action**: Execute **Phase 1** (Migrate Simple Value Types)
- Low risk, high reward
- Unblocks ~10% of remaining methods
- Clean architectural improvement
- No circular dependencies

**Follow-up**: Evaluate **Phase 2** (Interface-Based Workaround)
- After Phase 1 completion
- Pragmatic solution for remaining methods
- Can be refined later with Phase 3 approaches

**Long-term**: Consider **Phase 3** (Alternative Architectures)
- During next major refactoring cycle
- Requires broader architectural changes
- More thorough solution but higher effort

## Implementation Plan

### Step 1: Migrate EnumValue (Day 1, 2 hours)
- Create `runtime/enum.go`
- Move EnumValue struct and methods
- Add type alias in `value.go`
- Update imports and tests

### Step 2: Migrate TypeMetaValue (Day 1, 1 hour)
- Create `runtime/type_meta.go`
- Move TypeMetaValue struct and methods
- Add type alias in `value.go`
- Update imports and tests

### Step 3: Migrate SetValue (Day 1, 3 hours)
- Create `runtime/set.go`
- Move SetValue struct, methods (HasElement, AddElement, etc.)
- Add type alias in `value.go`
- Update imports and tests

### Step 4: Migrate ArrayValue (Day 2, 4 hours)
- Create `runtime/array.go`
- Move ArrayValue struct and methods
- Move NewArrayValue constructor
- Add type alias in `value.go`
- Update imports and tests

### Step 5: Implement ForInStatement in Evaluator (Day 2-3, 6 hours)
- Use migrated types to implement full ForInStatement logic
- Handle array, set, string, enum type iteration
- Use PushEnv/PopEnv for loop variable scoping
- Add comprehensive tests

### Step 6: Implement IndexExpression, ArrayLiteralExpression (Day 3, 4 hours)
- Use ArrayValue for array indexing
- Use ArrayValue for array literal construction
- Tests

### Step 7: Implement SetLiteral, NewArrayExpression (Day 3, 2 hours)
- Use SetValue for set literal construction
- Use ArrayValue for dynamic array creation
- Tests

## Success Metrics

**After Phase 1**:
- ✅ 4 types migrated to runtime/
- ✅ 5 evaluator methods fully implemented
- ✅ Progress: 43.8% → 54.2%
- ✅ All tests passing
- ✅ No circular dependencies

**After Phase 2** (if pursued):
- ✅ All 48 evaluator methods implemented
- ✅ Progress: 100%
- ✅ Adapter interface extended but still present
- ✅ All tests passing
