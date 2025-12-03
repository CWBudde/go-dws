<!-- trunk-ignore-all(prettier) -->
# DWScript to Go Port - Detailed Implementation Plan

This document breaks down the ambitious goal of porting DWScript from Delphi to Go into bite-sized, actionable tasks organized by stages. Each stage builds incrementally toward a fully functional DWScript compiler/interpreter in Go.

---

## Phase 1: Lexer (Tokenization)

**Completed**

---

## Phase 2: Parser Modernization ✅ **COMPLETED** (2025-01-21)

**Accomplishments**:

- Transformed parser to 100% cursor-based architecture (immutable TokenCursor)
- Built modern infrastructure: combinators, structured errors, automatic position tracking
- Removed ~6,700 lines of legacy code (31% reduction: 21K → 14.6K lines)
- Eliminated all dual-mode parsing (Traditional vs Cursor)
- Separated error recovery and semantic analysis into dedicated modules
- Improved test coverage from 73.4% → 78.5% (+700 test lines)

**Phases**: 2.1 Foundation, 2.2 Cursor Abstraction, 2.3 Combinators, 2.4 Position Tracking, 2.5 Separation of Concerns, 2.6 Advanced Features, 2.7 Traditional Removal, 2.8 Test Coverage

**Documentation**: [PHASE2_MODERNIZATION.md](PHASE2_MODERNIZATION.md), [PHASE2_COMPARISON.md](PHASE2_COMPARISON.md)

---

## Phase 3: Interpreter Architecture Refactoring

**Goal**: Refactor the interpreter architecture to improve maintainability, testability, and extensibility. Address architectural debt accumulated during rapid feature development.

**Status**: Planned | **Complexity**: High | **Priority**: Medium | **Estimated**: 4-6 weeks

**Motivation**: The interpreter has grown to 68 files and 80K lines with architectural issues:

- God object (Interpreter struct with 25+ fields mixing concerns)
- Giant switch statement in Eval() (230 lines, 30+ cases)
- Tight coupling between evaluation logic, type system, and runtime
- Poor separation of concerns (one package mixing values, execution, built-ins, types)
- Reflection hacks for circular dependency issues
- Heavy type assertion usage throughout
- Inconsistent error handling patterns
- Difficult to test components in isolation

**Benefits**:

- Improved code maintainability and readability
- Better testability (unit test individual components)
- Easier to add new language features
- Reduced coupling and improved modularity
- Foundation for future optimizations
- Consistent patterns and practices throughout

**Non-Goals**:

- Changing interpreter semantics or behavior
- Breaking existing tests (all tests must continue passing)
- Changing public API surface
- Performance regression (maintain or improve speed)

---

### Phase 3.1: Preparation and Analysis

- [x] 3.1.1 Create interpreter refactoring design document
  - Analyze current architecture and identify pain points
  - Design new architecture with clear component boundaries
  - Create migration strategy to avoid big-bang refactor
  - Document new patterns and conventions
  - File: `docs/architecture/interpreter-refactoring.md`
  - Estimated: 1 week
  - Acceptance: Design doc reviewed and approved, migration strategy defined
  - **Completed**: Commit ed7fd2b - 900+ line design doc with architecture analysis and migration strategy

- [x] 3.1.2 Add comprehensive interpreter benchmarks
  - Create benchmark suite covering all major operations
  - Benchmark expression evaluation, statement execution, function calls
  - Benchmark object creation, array operations, property access
  - Establish performance baseline before refactoring
  - Files: `internal/interp/*_bench_test.go` (new/expanded)
  - Estimated: 3 days
  - Acceptance: 50+ benchmarks covering key operations, baseline documented
  - **Completed**: Commit ed7fd2b - 88 benchmarks total, baseline established in docs/

- [ ] 3.1.3 Increase test coverage to 80%+ on core packages
  - **Current Coverage**: 59.3% overall (as of 2025-11-18)
  - **Completed Work**: Improved interp/builtins (→82.0%), interp/evaluator (→50.4%), pkg/printer (→85.9%), pkg/token (→100%), bytecode (→74.4%), semantic (→62.4%)
  - **Remaining Work**:
    - [ ] Bytecode package: 74.4% → 80%+ (need 5.6% more)
    - [ ] Semantic analyzer: 62.4% → 80%+ (need 17.6% more)
    - [ ] Parser: 58.4% → 80%+ (need 21.6% more)
    - [ ] Interp/evaluator: 50.4% → 80%+ (need 29.6% more - builtin function paths)
    - [ ] Interp/runtime: 48.4% → 80%+ (need 31.6% more)
    - [ ] Low priority: cmd/dwscript/cmd (36.7%), ast/pkg (24.9%), cmd/dwscript (0.0%)
  - Estimated: 1-2 weeks
  - Acceptance: Coverage report shows 80%+ on bytecode, semantic, parser, interp/evaluator, interp/runtime
  - **Note**: Deferred - fixing failing tests is better done after architecture improvements

---

### Phase 3.2: Value System Refactoring

- [x] 3.2.1 Introduce value interfaces for type operations
  - Create `NumericValue`, `ComparableValue`, `OrderableValue` interfaces
  - Create `CopyableValue` interface to formalize copy semantics
  - Create `ReferenceValue` vs `ValueType` distinction
  - Update existing value types to implement appropriate interfaces
  - Files: `internal/interp/runtime/value_interfaces.go` (new)
  - Estimated: 4 days
  - Acceptance: Interface hierarchy defined, all values implement correct interfaces
  - **Completed**: Commit b9525bf - 10 value interfaces defined, all primitives implement them

- [x] 3.2.2 Extract value types to runtime sub-package
  - Create `internal/interp/runtime/` package
  - Move all Value type definitions (IntegerValue, StringValue, etc.)
  - Move value creation helpers (NewIntegerValue, etc.)
  - Move value conversion helpers (boxVariant, unwrapVariant, etc.)
  - Update imports throughout interp package
  - Files: Move `value.go` → `runtime/value.go`, `runtime/numeric.go`, etc.
  - Estimated: 3 days
  - Acceptance: All value types in runtime/, clean package structure, tests pass
  - **Completed**: Commit b9525bf - Primitives moved to runtime/, type aliases for compatibility

- [x] 3.2.3 Implement value object pooling for performance
  - Add sync.Pool for frequently allocated types (IntegerValue, FloatValue, BooleanValue)
  - Update value creation functions to use pools
  - Add pool statistics for monitoring
  - Benchmark to ensure no regression
  - Files: `internal/interp/runtime/pool.go` (new)
  - Estimated: 2 days
  - Acceptance: Object pooling working, benchmarks show 10-20% allocation reduction
  - **Completed**: Commit b9525bf - Object pools + singleton booleans, zero allocations when released

- [x] 3.2.4 Add runtime package utilities (not in original plan)
  - Value conversion utilities (ToInteger, ToFloat, etc.) with safe arithmetic
  - Structured error types (ConversionError, ArithmeticError, etc.)
  - Type checking and comparison helpers (IsInteger, Equal, LessThan, etc.)
  - String operation utilities (Concat, Contains, etc.)
  - Comprehensive test coverage (241 tests)
  - Files: `internal/interp/runtime/{conversion,errors,helpers}.go` and tests
  - **Completed**: Commit 7554357 - 2000+ lines of utilities with full test coverage

---

### Phase 3.3: Execution Context Separation

- [x] 3.3.1 Create ExecutionContext struct
  - Extract execution state from Interpreter into ExecutionContext
  - Include: env, callStack, controlFlow (break/continue/exit), exception state
  - Create ControlFlow helper type for break/continue/exit/return signals
  - Update Eval methods to accept context parameter
  - Files: `internal/interp/evaluator/context.go` (new), `env_adapter.go` (new)
  - Estimated: 5 days
  - Acceptance: Context cleanly separated, passed explicitly, tests pass
  - **Completed**: Added ExecutionContext struct with ControlFlow type, initialized in Interpreter constructor, all tests pass

- [x] 3.3.2 Implement explicit control flow handling
  - Replace boolean flags (breakSignal, continueSignal, etc.) with ControlFlow type
  - Create ControlFlow.Break(), ControlFlow.Continue(), etc. methods
  - Use result types or special return values for control flow
  - Simplify loop and switch statement logic
  - Files: Modified `statements_loops.go`, `functions_user.go`, `functions_records.go`, `functions_pointers.go`, `statements_control.go`, `statements_declarations.go`, `interpreter.go`
  - Estimated: 3 days
  - Acceptance: No more boolean flags, cleaner control flow, tests pass
  - **Completed**: Replaced all boolean control flow flags with ExecutionContext ControlFlow type, all tests pass

- [x] 3.3.3 Create CallStack abstraction
  - Extract call stack management from Interpreter
  - Create CallStack type with Push/Pop/Current/Depth methods
  - Add stack overflow detection
  - Improve error messages with stack traces
  - Files: `internal/interp/evaluator/callstack.go` (new), `callstack_test.go` (14 tests)
  - Estimated: 2 days
  - Acceptance: CallStack is independent, proper stack traces, tests pass
  - **Completed**: Created CallStack abstraction with overflow detection, comprehensive tests, integrated with ExecutionContext

---

### Phase 3.4: Type System Separation

- [x] 3.4.1 Extract type registries to TypeSystem
  - Create TypeSystem struct with all registries
  - Include: classes, records, interfaces, functions, helpers, operators
  - Create TypeRegistry interface for common operations
  - Move registration logic to TypeSystem methods
  - Files: `internal/interp/types/type_system.go` (new)
  - Estimated: 4 days
  - Acceptance: All type info in TypeSystem, clean registry APIs, tests pass
  - **Completed**: Created TypeSystem with all registries (classes, records, interfaces, functions, helpers, operators, conversions, type IDs), clean APIs for registration and lookup, integrated with Interpreter, all tests pass

- [x] 3.4.2 Create ClassRegistry abstraction
  - Extract class management from map to ClassRegistry type
  - Add methods: Register, Lookup, LookupHierarchy, Exists
  - Support class hierarchy queries efficiently
  - Move class-related helpers to registry
  - Files: `internal/interp/types/class_registry.go` (new)
  - Estimated: 3 days
  - Acceptance: ClassRegistry handles all class operations, tests pass
  - **Completed**: Created ClassRegistry with comprehensive methods (Register, Lookup, Exists, LookupHierarchy, IsDescendantOf, GetDepth, FindDescendants), integrated with TypeSystem, 10 unit tests covering all functionality, all tests pass

- [x] 3.4.3 Create FunctionRegistry with overload support
  - Extract function management to FunctionRegistry
  - Properly handle overloading with signature matching
  - Support qualified name lookup (Unit.Function)
  - Add efficient lookup by signature
  - Files: `internal/interp/types/function_registry.go` (new)
  - Estimated: 3 days
  - Acceptance: FunctionRegistry handles all function operations, tests pass
  - **Completed**: Created FunctionRegistry with comprehensive overload support (Register, Lookup, LookupQualified, GetOverloadCount, FindByParameterCount, ValidateNoConflicts), unit-qualified name support (Unit.Function), metadata queries, integrated with TypeSystem, 15 unit tests covering all functionality, all tests pass

---

### Phase 3.5: Evaluator Refactoring

#### Completed Work (3.5.1-3.5.23) ✅

**Summary**: Split Interpreter into Evaluator + TypeSystem + ExecutionContext. Implemented 48 visitor methods with adapter pattern for gradual migration.

**Infrastructure Created**:

- Evaluator with visitor pattern (9 expression files + statements + declarations)
- TypeSystem with registries (classes, records, interfaces, functions, helpers, operators, subranges)
- ExecutionContext with call stack, control flow, environment management
- FunctionRegistry with overload resolution

**Tasks 3.5.1-3.5.6** ✅: ExecuteUserFunction integration, interface map removal, declaration dependency analysis

**Declaration Migration (3.5.7-3.5.16)** ✅: All 10 declaration types migrated to evaluator

- Tier 1 (Pure Registration): SetDecl, EnumDecl, OperatorDecl
- Tier 2 (Bounds/Types): ArrayDecl, TypeDeclaration
- Tier 3 (Functions): FunctionDecl (global + method implementations)
- Tier 4 (Extensions): HelperDecl, InterfaceDecl
- Tier 5 (Complex): RecordDecl, ClassDecl (with VMT, partial classes, inheritance)

**Bridge Elimination (3.5.17-3.5.23)** ✅: Value types moved to runtime package

- 3.5.17: ObjectInstance → `runtime/object.go` (IClassInfo interface created)
- 3.5.18: ExceptionValue → `runtime/exception.go`
- 3.5.19: SubrangeValue → `runtime/subrange.go`
- 3.5.20: InterfaceInstance → `runtime/interface_instance.go` (IInterfaceInfo interface)
- 3.5.21: TypedNilValue → already in runtime
- 3.5.22: Function execution + TryImplicitConversion migrated to evaluator (11 subtasks)
- 3.5.23: All 6 bridge constructors removed

**Key Files Created**:

- `runtime/class_interface.go` (IClassInfo, IInterfaceInfo interfaces)
- `runtime/object.go`, `exception.go`, `subrange.go`, `interface_instance.go`
- `evaluator/type_conversion.go` (TryImplicitConversion)
- `evaluator/visitor_declarations.go` (all declaration visitors)

---

#### Remaining Tasks (3.5.24-3.5.40)

---

**Adapter Reduction** (3.5.24-3.5.40)

Goal: Reduce adapter interface from ~75 methods to ~20 essential methods (~75% reduction).

**Key Finding** (Task 3.5.24): Full adapter elimination is impractical due to circular import constraints. Target: minimize to essential operations only.

- [x] **3.5.24** Document Remaining Adapter Methods ✅
  - **Deliverable**: `docs/adapter-audit-3.5.24.md` - 75 methods documented with justifications
  - **Finding**: 35 easy, 32 medium, 3 hard methods. ~78 methods have zero callers.

---

#### Phase 1: Remove Zero-Caller Methods (3.5.25-3.5.27)

**Goal**: Remove ~78 unused methods from adapter interface. **Effort**: 2-3 hours total.

- [x] **3.5.25** Remove Unused Lookup Methods ✅
  - **Methods removed** (3): `LookupClass`, `LookupRecord`, `LookupFunction`
  - **Methods kept** (9): `LookupInterface`, `LookupHelpers`, `ResolveClassInfoByName`, `GetClassNameFromInfo`, `GetOperatorRegistry`, `GetEnumTypeID`, `GetHelperName`, `GetInterfaceName`, `GetInterfaceParent` - all have callers
  - **Files**: `evaluator/evaluator.go`, `adapter_types.go`, `adapter_functions.go`, `type_registry_test.go`
  - **Test**: `go build ./...` compiles, TypeRegistry tests pass

- [x] **3.5.26** Remove Unused Binary Operator Methods ✅
  - **Methods removed** (3): `EvalVariantBinaryOp`, `EvalInOperator`, `EvalEqualityComparison`
  - **Files**: `evaluator/evaluator.go`, `adapter_references.go`
  - **Test**: `go build ./...` compiles, evaluator tests pass

- [x] **3.5.27** Remove Unused Class Declaration Methods ✅
  - **Methods removed** (15): `CreateClassValue`, `CreateLambda`, `ReadPropertyValue`, `IsMethodParameterless`, `CreateMethodCall`, `CreateMethodPointerFromObject`, `GetClassName(obj)`, `GetClassType(obj)`, `GetClassNameFromClassInfo`, `GetClassTypeFromClassInfo`, `GetClassVariableFromClassInfo`, `CallIndexedPropertyGetter`, `CallRecordPropertyGetter`, `RegisterClassEarly`, `CreateMethodMetadata`
  - **Files**: `evaluator/evaluator.go`, `adapter_types.go`, `adapter_objects.go`, `adapter_references.go`
  - **Test**: `go build ./...` compiles, class/declaration tests pass

---

#### Phase 2: Inline Single-Use Methods (3.5.28-3.5.31)

**Goal**: Replace adapter calls with direct TypeSystem/runtime calls. **Effort**: 4-6 hours total.

- [x] **3.5.28** Inline Property Type Extraction Methods ✅
  - **Methods removed** (2): `GetObjectFieldValue`, `GetClassVariableValue`
  - **Note**: `GetClassName` and `GetClassVariableFromClassInfo` were already removed in task 3.5.27
  - **Finding**: Both methods had **zero callers** - removed directly without needing inline replacements
  - **Replacement pattern** (documented for future reference):

    ```go
    // For GetObjectFieldValue:
    if objVal, ok := obj.(evaluator.ObjectValue); ok {
        val := objVal.GetField(fieldName)
    }
    // For GetClassVariableValue:
    if objVal, ok := obj.(evaluator.ObjectValue); ok {
        val, ok := objVal.GetClassVar(varName)
    }
    ```

  - **Files modified**:
    - `evaluator/evaluator.go` (removed interface methods)
    - `adapter_references.go` (removed implementations)
    - `evaluator/type_conversion_test.go` (removed mock methods)
  - **Result**: -508 lines, all tests passing
  - **Actual effort**: 30 minutes (much faster than estimated due to zero callers)

- [x] **3.5.29** Inline Type Check Methods ✅
  - **Methods removed** (3): `CheckType`, `GetClassMetadataFromValue`, `GetObjectInstanceFromValue`
  - **Work completed**:
    - `GetObjectInstanceFromValue`: Replaced 2 calls with `ObjectValue` interface type assertions
    - `GetClassMetadataFromValue`: Replaced 1 direct call + inlined the `getClassMetadataFromValue()` helper
    - `CheckType`: Already migrated away (zero callers found)
  - **Inline patterns used**:

    ```go
    // GetObjectInstanceFromValue:
    if _, ok := val.(ObjectValue); !ok {
        return error
    }

    // GetClassMetadataFromValue in helper:
    if objVal, ok := obj.(ObjectValue); ok {
        className := objVal.ClassName()
        if classInfo := e.typeSystem.LookupClass(className); classInfo != nil {
            if metadataProvider, ok := classInfo.(interface{ GetMetadata() *runtime.ClassMetadata }); ok {
                return metadataProvider.GetMetadata()
            }
        }
    }
    ```

  - **Files modified**:
    - `evaluator/type_casts.go` (1 inline)
    - `evaluator/visitor_expressions_types.go` (1 inline + helper rewrite)
    - `evaluator/overload_resolution.go` (1 inline)
    - `evaluator/evaluator.go` (removed 3 interface methods)
    - `adapter_objects.go` (removed 2 implementations, removed runtime import)
    - `evaluator/type_conversion_test.go` (removed 3 mock methods)
  - **Result**: 3 methods removed, all tests passing
  - **Actual effort**: 45 minutes (as estimated)

- [ ] **3.5.30** Inline Array/Value Creation Methods
  - **Methods** (2): `CreateArray`, `CreateArrayValue`
  - **Work**: Use `runtime.NewArrayValue()` directly in evaluator
  - **Files**: `evaluator/visitor_expressions_functions.go`, `evaluator/visitor_statements.go`
  - **Effort**: 30 minutes

- [ ] **3.5.31** Inline Interface Wrapper Methods
  - **Methods** (3): `CreateInterfaceWrapper`, `CreateTypeCastWrapper`, `GetInterfaceInstanceFromValue`
  - **Work**: Use runtime package directly for interface/cast operations
  - **Files**: `evaluator/visitor_expressions_types.go`, `evaluator/type_casts.go`
  - **Effort**: 1 hour

---

#### Phase 3: Migrate Property Callbacks (3.5.32-3.5.34)

**Goal**: Complete ObjectValue callback pattern for property access. **Effort**: 3-4 hours total.

- [ ] **3.5.32** Consolidate ExecutePropertyRead
  - **Current**: 4 calls to `e.adapter.ExecutePropertyRead`
  - **Work**: Move property getter execution logic to evaluator
  - **Pattern**: Evaluator already has `ExecuteMethodWithSelf` callback - reuse for getters
  - **Files**: `evaluator/visitor_expressions_members.go`, `evaluator/visitor_expressions_identifiers.go`
  - **Effort**: 1.5 hours

- [ ] **3.5.33** Consolidate ExecuteIndexedPropertyRead
  - **Current**: 8 calls to `e.adapter.ExecuteIndexedPropertyRead`
  - **Work**: Implement indexed property getter in evaluator using method callback
  - **Files**: `evaluator/visitor_expressions_indexing.go`
  - **Effort**: 1.5 hours

- [ ] **3.5.34** Remove Deprecated Property Methods
  - **Methods**: `ReadPropertyValue`, `CallIndexedPropertyGetter`, `CallRecordPropertyGetter`
  - **Work**: Already deprecated, verify zero callers, remove
  - **Files**: `evaluator/evaluator.go`, `adapter_types.go`
  - **Effort**: 30 minutes

---

#### Phase 4: Reduce EvalNode Fallback (3.5.35-3.5.38)

**Goal**: Reduce EvalNode calls from ~44 to ~10 essential cases. **Effort**: 8-12 hours total.

**Current EvalNode hotspots**:

- `assignment_helpers.go`: 10 calls (member/property assignments)
- `visitor_expressions_members.go`: 7 calls (complex member access)
- `member_assignment.go`: 6 calls (member assignments)
- Other files: ~21 calls

- [ ] **3.5.35** Migrate Member Assignment to Evaluator
  - **Current**: `member_assignment.go` uses EvalNode for 6 assignment patterns
  - **Work**:
    1. Identify assignment patterns delegating to EvalNode
    2. Implement native assignment handling for: field assignment, property write, indexed write
    3. Keep EvalNode only for truly complex cases
  - **Files**: `evaluator/member_assignment.go`
  - **Effort**: 3-4 hours
  - **Target**: Reduce from 6 to 2 EvalNode calls

- [ ] **3.5.36** Migrate Assignment Helpers
  - **Current**: `assignment_helpers.go` uses EvalNode for 10 patterns
  - **Work**: Group by pattern type, implement native handlers
  - **Patterns**: Simple identifier, reference, member, indexed, slice
  - **Files**: `evaluator/assignment_helpers.go`
  - **Effort**: 3-4 hours
  - **Target**: Reduce from 10 to 3 EvalNode calls

- [ ] **3.5.37** Migrate Complex Member Access
  - **Current**: `visitor_expressions_members.go` uses EvalNode for 7 patterns
  - **Work**: Handle remaining member access patterns natively
  - **Patterns**: Nested property, interface property, helper property, type cast property
  - **Files**: `evaluator/visitor_expressions_members.go`
  - **Effort**: 2-3 hours
  - **Target**: Reduce from 7 to 2 EvalNode calls

- [ ] **3.5.38** Audit Remaining EvalNode Calls
  - **Work**: Review all remaining EvalNode calls, document which are essential
  - **Essential categories** (keep ~10):
    - Unit declarations
    - Complex property writes with side effects
    - Compound operations on special types
  - **Deliverable**: Update `docs/adapter-audit-3.5.24.md` with final EvalNode inventory
  - **Effort**: 1 hour

---

#### Phase 5: Documentation and Stabilization (3.5.39-3.5.40)

- [ ] **3.5.39** Update Architecture Documentation
  - **Work**:
    1. Update CLAUDE.md Architecture Overview with final evaluator/interpreter split
    2. Document essential adapter methods and why they exist
    3. Document callback patterns (ObjectValue, property read, method execution)
  - **Files**: `CLAUDE.md`, `docs/adapter-audit-3.5.24.md`
  - **Effort**: 2-3 hours

- [ ] **3.5.40** Create Phase 3.5 Completion Summary
  - **Deliverable**: `docs/phase3.5-summary.md`
  - **Content**:
    - Initial state: monolithic Interpreter (~15k LOC)
    - Final state: Evaluator + TypeSystem + ExecutionContext + minimal adapter
    - Methods migrated vs methods kept
    - Performance impact (if any)
    - Lessons learned
  - **Effort**: 2-3 hours

---

#### Phase 3.5 Summary

| Phase | Tasks | Description | Status | Effort |
|-------|-------|-------------|--------|--------|
| Bridge Elimination | 3.5.17-3.5.23 | Move value types to runtime | ✅ Complete | Done |
| Audit | 3.5.24 | Document adapter methods | ✅ Complete | Done |
| Zero-Caller Removal | 3.5.25-3.5.27 | Remove 21 unused methods | ✅ Complete | Done |
| Single-Use Inline | 3.5.28-3.5.31 | Inline simple adapter calls | 50% (2/4) | 1.25h done |
| Property Callbacks | 3.5.32-3.5.34 | Complete callback patterns | Pending | 3-4h |
| EvalNode Reduction | 3.5.35-3.5.38 | Reduce fallback calls 44→~10 | Pending | 8-12h |
| Documentation | 3.5.39-3.5.40 | Final docs and summary | Pending | 4-6h |

**Total Remaining Effort**: ~25-35 hours (vs original 80-100h estimate for full elimination)

**Realistic Target**:

- Adapter interface: 75 → ~20 methods (73% reduction)
- EvalNode calls: 44 → ~10 calls (77% reduction)
- Essential methods kept: method execution, constructor execution, lambda/closure handling

**Why Full Elimination Is Impractical**:

1. **Circular imports**: ClassInfo lives in `interp`, evaluator can't import `interp`
2. **Environment management**: Interpreter owns env/call stack, method execution needs it
3. **Closure capture**: Lambda environments require Interpreter's environment system
4. **Diminishing returns**: Last 20 methods require ~60h refactoring for minimal benefit

---

### Phase 3.6: Error Handling Improvements

- [x] 3.6.1 Implement consistent error wrapping
  - Replace newError() with proper error wrapping
  - Use fmt.Errorf with %w for error chains
  - Create custom error types for different error categories
  - Add context to all errors (position, expression, values)
  - Files: `internal/interp/errors/errors.go` (new)
  - Estimated: 4 days
  - Acceptance: All errors wrapped properly, error chains work, tests pass
  - **Completed**: Created 5 error categories (Type, Runtime, Undefined, Contract, Internal), 18 tests passing

- [x] 3.6.2 Create EvalResult type for cleaner error propagation
  - Define EvalResult struct with value and error
  - Add OrReturn() method for early return pattern
  - Reduce repetitive if isError(val) checks
  - Option: Use monadic approach or keep explicit
  - Files: `internal/interp/evaluator/result.go` (new)
  - Estimated: 5 days
  - Acceptance: Less boilerplate, cleaner code, tests pass
  - **Completed**: 6 error handling patterns (OrReturn, FirstError, AllValues, Map, AndThen, Collect), 19 tests passing

- [x] 3.6.3 Standardize error messages and error types
  - Create error constants for common errors
  - Ensure consistent formatting: "operation failed: reason"
  - Include relevant context in all errors
  - Create error catalog documentation
  - Files: `internal/interp/errors/catalog.go` (new)
  - Estimated: 2 days
  - Acceptance: Consistent error messages, documented catalog, tests pass
  - **Completed**: 40+ error constants, 30+ helper functions, 31 tests passing, comprehensive documentation

---

### Phase 3.7: Built-in Function Registry & Migration

**Status**: ✅ Complete (All 9 tasks: 3.7.1-3.7.9)
**Documentation**: See `docs/phase3.7-summary.md` for comprehensive details

**Summary**: Migrated 225 of 244 built-in functions (92.2%) to modular, registry-based architecture with O(1) case-insensitive lookup. Reduced switch statement by 85% (~600 lines). Created 14 files in `internal/interp/builtins/` with 30+ Context helper methods. >90% test coverage, zero circular dependencies.

**Functions by Category**:
- Math: 62 | String: 57 | DateTime: 52 | Conversion: 11 | Encoding: 5
- JSON: 7 | Type: 2 | Array: 16 | Collections: 8 | Variant: 10 | I/O: 2 | System: 4

**Key Achievements**:
- ✅ Context interface with 30+ helper methods for type-safe builtin implementation
- ✅ Thread-safe registry with O(1) case-insensitive lookup matching DWScript semantics
- ✅ Fixed polymorphic functions (Low, High, Concat) to support arrays, strings, enums, and type meta-values
- ✅ All functions properly categorized and registered with comprehensive documentation
- ✅ Interface-based design eliminates circular dependencies

**Functions Not Migrated** (~19 functions, 7.8%):
- Var-param functions (7): Inc, Dec, Swap, DivMod, Insert, Delete, DecodeDate (require lvalue modification)
- Random functions (6): Random, RandomInt, Randomize, SetRandSeed, RandSeed, RandG
- String helpers (3): StrSplit, StrJoin, StrArrayPack
- Enum functions (2): Succ, Pred (need enum metadata)

**Files Created**: context.go, registry.go, register.go, math.go, strings_basic.go, datetime.go, conversion.go, encoding.go, json.go, type.go, variant.go, io.go, array.go, collections.go, system.go (+ test files)

---

### Phase 3.8: Dependency and Import Cleanup

- [x] 3.8.1 Fix circular dependency with pkg/dwscript
  - Remove reflection hack in NewWithOptions
  - Create proper Options interface in internal/interp
  - Have pkg/dwscript implement the interface
  - Pass through interface, not concrete type
  - Files: `internal/interp/options.go` (new)
  - Estimated: 2 days
  - Acceptance: No reflection, clean interface, tests pass

- [x] 3.8.2 Reduce pkg/ast usage in internal/interp
  - Minimize imports from pkg/ast to internal/ast
  - Only use pkg/ast for semantic info
  - Create adapter if needed
  - Document why pkg/ast is used
  - Files: `internal/interp/errors/*.go` (updated), documentation added
  - Estimated: 2 days
  - Acceptance: Clear separation, documented rationale, tests pass

---

### Phase 3.9: Testing and Documentation

- [ ] 3.9.1 Add unit tests for new components
  - Test ExecutionContext independently
  - Test TypeSystem registries independently
  - Test Evaluator with mock context
  - Achieve 95%+ coverage on new code
  - Estimated: 1 week
  - Acceptance: All new components fully tested, 95%+ coverage

- [ ] 3.9.2 Update architecture documentation
  - Document new architecture in CLAUDE.md
  - Create architecture diagrams showing components
  - Document design patterns used (Visitor, Registry, etc.)
  - Create migration guide for contributors
  - Files: `docs/architecture/interpreter.md`, update `CLAUDE.md`
  - Estimated: 3 days
  - Acceptance: Complete documentation, diagrams, migration guide

- [ ] 3.9.3 Verify no performance regression
  - Run benchmark suite on refactored code
  - Compare with baseline from Task 3.8.1
  - Ensure no more than 5% regression
  - If regression > 5%, investigate and optimize
  - Estimated: 2 days
  - Acceptance: Performance maintained or improved, benchmarks documented

---

### Phase 3.10: Final Integration

- [ ] 3.10.1 Update all existing code to use new architecture
  - Update all evaluation call sites
  - Remove deprecated code and patterns
  - Ensure consistent usage of new patterns
  - Run full test suite
  - Estimated: 1 week
  - Acceptance: All code updated, no old patterns remain, tests pass

- [ ] 3.10.2 Final cleanup and polish
  - Remove dead code and unused exports
  - Ensure consistent naming conventions
  - Add missing godoc comments
  - Run linters and fix issues
  - Format all code
  - Estimated: 2 days
  - Acceptance: Clean codebase, no warnings, all linters pass

---

**Phase 3 Summary**:
- **Total Tasks**: 29 tasks (3.1.1 - 3.10.2)
- **Completed Tasks**: 9 tasks (3.1.1, 3.1.2, 3.2.1-3.2.4, 3.3.1-3.3.3, 3.4.1-3.4.2, 3.5.1-3.5.3)
- **Remaining Tasks**: 20 tasks (3.5.4-3.5.5, 3.6.1-3.6.3, 3.7.1-3.7.2, 3.8.1-3.8.2, 3.9.1-3.9.3, 3.10.1-3.10.2)
- **Estimated Timeline**: 6-8 weeks full-time (updated with new migration tasks)
- **Risk**: Medium (comprehensive refactoring, but tests provide safety net)
- **Dependencies**: None (can be done alongside other work with feature branches)
- **Rollback Strategy**: Each task is incremental; can stop at any point
- **Success Metrics**:
  - All existing tests continue to pass
  - Test coverage ≥ 90%
  - No performance regression (≤5%)
  - Code organization: ≤10 files per package
  - Reduced coupling: clear component boundaries
  - Improved maintainability: easier to add features

**Recommended Approach**:
- Work in feature branch: `refactor/interpreter-architecture`
- Complete one phase at a time (9.8.1, then 9.8.2, etc.)
- Keep main branch stable with current architecture
- Merge when entire phase is complete and tested
- Use git bisect to identify issues quickly
- Consider doing tasks 9.80-9.82 first to establish baseline

---

## Phase 4: Control Flow and Statement Execution

- Variable declarations with optional type annotations and initialization
- Assignment statements with DWScript's `:=` operator
- Block statements with `begin...end` syntax
- Built-in functions (PrintLn, Print) and user-defined function calls
- Environment/symbol table with nested scope support
- Runtime value system supporting Integer, Float, String, Boolean, and Nil types
- Sequential statement execution with proper error handling
- If-then-else statements with proper boolean evaluation
- While loops with condition testing before execution
- Repeat-until loops with condition testing after execution
- For loops supporting both `to` and `downto` directions with integer bounds
- Case statements with value matching and optional else branches
- Full integration with existing type system and error reporting

---

## Phase 5: Functions, Procedures, and Scope Management ✅ **COMPLETED** (91.3%)

- Function and procedure declarations with parameter lists and return types
- By-reference parameters (`var` keyword) - parsing implemented, runtime partially complete
- Function calls with argument passing and return value handling
- Lexical scoping with proper environment nesting
- Built-in functions for output and basic operations
- Recursive function support with environment cleanup
- Symbol table integration for function resolution

---

## Phase 6: Static Type Checking and Semantic Analysis ✅ **COMPLETED**

- Built the reusable type system in `types/` (primitive, function, aggregate types plus coercion rules); see docs/stage6-type-system-summary.md for the full compatibility matrix.
- Added optional type annotations to AST nodes and parser support for variables, parameters, and return types so semantic analysis has complete metadata.
- Implemented the semantic analyzer visitor that resolves identifiers, validates declarations/assignments/expressions, enforces control-flow rules, and reports multiple errors per pass with 88.5% coverage.
- Hooked the analyzer into the parser/interpreter/CLI (with a disable flag) so type errors surface before execution and runtime uses inferred types.
- Upgraded diagnostics with per-node position data, the `errors/` formatter, and curated fixtures in `testdata/type_errors` plus `testdata/type_valid`, alongside CLI integration suites.

---

## Phase 6.1: Semantic Analyzer Architectural Refactoring

**Motivation**: The semantic analyzer has grown organically to ~20,000 lines across 40+ files. While functional, it exhibits the "God Object" anti-pattern with an 81-field `Analyzer` struct, no clear separation between analysis passes, scattered state management, and maintenance challenges. This refactoring addresses architectural issues before adding more features.

**Impact**:
- **Maintainability**: Easier to understand, test, and extend
- **LSP Support**: Better foundation for language server features (goto definition, find references)
- **Performance**: Multi-pass design enables parallel analysis and better caching
- **Code Quality**: Reduced coupling, clearer separation of concerns

**Status**: NOT STARTED

**Priority**: P0 - CRITICAL (Must complete before major feature additions)

**Estimate**: 80-120 hours (2-3 weeks full-time)

---

## Task 6.1.1: Extract and Centralize Type Registry 🎯 **HIGH PRIORITY**

**Goal**: Replace 7 scattered type maps with a unified `TypeRegistry` component that handles all type registration, lookup, and resolution.

**Estimate**: 16-20 hours (2-3 days)

**Status**: NOT STARTED

**Priority**: P0 - CRITICAL (Foundation for other refactorings)

**Current Problem**:
The `Analyzer` struct currently has 7 separate type maps (in analyzer.go):
- `classes map[string]*types.ClassType`
- `interfaces map[string]*types.InterfaceType`
- `enums map[string]*types.EnumType`
- `records map[string]*types.RecordType`
- `sets map[string]*types.SetType`
- `arrays map[string]*types.ArrayType`
- `typeAliases map[string]*types.TypeAlias`

This causes:
- Duplicate lookup logic across 7 different maps
- No way to iterate all types for documentation/LSP
- Inconsistent case-sensitivity handling (each map does it separately)
- No central type resolution point
- Difficult to add new type kinds

**Insight from Analysis**: Each map independently handles case-insensitivity with `strings.ToLower()`. Type resolution requires checking all 7 maps sequentially. No unified API for "register any type" or "lookup any type".

**Subtasks**:

- [x] **6.1.1.1 Design TypeRegistry interface**
  - [x] Create `internal/semantic/type_registry.go`
  - [x] Define `TypeRegistry` struct with unified internal storage
  - [x] Design methods: `Register(name string, typ types.Type)`, `Resolve(name string) (types.Type, bool)`
  - [x] Add methods: `AllTypes() map[string]types.Type`, `TypesByKind(kind string) []types.Type`
  - [x] Include position tracking for type definitions (needed for LSP)
  - [x] Write comprehensive unit tests for registry operations

- [x] **6.1.1.2 Implement type registration**
  - [x] Create single internal map: `types map[string]*TypeDescriptor`
  - [x] `TypeDescriptor` contains: `Name string`, `Type types.Type`, `Position lexer.Position`, `Visibility int`
  - [x] Handle case-insensitive lookup with `strings.ToLower()` in one place
  - [x] Support type aliasing and resolution of aliases to underlying types
  - [x] Validate uniqueness (prevent duplicate type names)
  - [x] Write tests for registration edge cases

- [x] **6.1.1.3 Migrate existing type maps**
  - [x] Replace `Analyzer.classes` with `TypeRegistry.Register/Resolve` calls
  - [x] Replace `Analyzer.interfaces` with registry calls
  - [x] Replace `Analyzer.enums`, `Analyzer.records`, `Analyzer.sets` with registry
  - [x] Replace `Analyzer.arrays`, `Analyzer.typeAliases` with registry
  - [x] Update all call sites in `analyze_classes.go`, `analyze_enums.go`, etc.
  - [x] Run full test suite to verify no regressions

- [x] **6.1.1.4 Add registry query capabilities**
  - [x] Implement `TypesByKind(kind string)` to get all classes/interfaces/etc.
  - [x] Implement `TypesInScope(scope)` for nested type scopes
  - [x] Add `FindTypeByPosition(pos)` for LSP "type at cursor"
  - [x] Add `GetTypeDependencies(typeName)` for dependency analysis
  - [x] Write integration tests for LSP-style queries

**Files to Create**:
- `internal/semantic/type_registry.go` (new)
- `internal/semantic/type_registry_test.go` (new)

**Files to Modify**:
- `internal/semantic/analyzer.go` (remove 7 type maps, add `TypeRegistry` field)
- `internal/semantic/analyze_classes.go` (use registry instead of `a.classes`)
- `internal/semantic/analyze_interfaces.go` (use registry)
- `internal/semantic/analyze_enums.go` (use registry)
- `internal/semantic/analyze_records.go` (use registry)
- `internal/semantic/analyze_sets.go` (use registry)
- `internal/semantic/analyze_arrays.go` (use registry)
- `internal/semantic/type_resolution.go` (simplify with unified registry)

**Success Metrics**:
- All 7 type maps removed from `Analyzer` struct
- Single call site for all type registration
- Unified case-insensitive lookup
- All existing tests pass
- New tests for TypeRegistry at 90%+ coverage

---

## Task 6.1.2: Implement Multi-Pass Analysis Architecture 🎯 **HIGH PRIORITY**

**Goal**: Restructure semantic analysis into explicit passes with clear dependencies, replacing the current single-pass approach that requires complex fix-ups.

**Status**: 95.2% COMPLETE (1716/1803 tests passing with experimental passes enabled)

**Priority**: P0 - CRITICAL (Enables correct handling of forward declarations and type dependencies)

**Original Problem**: The analyzer tried to do everything in one pass, causing issues with forward declarations, complex type dependencies, and requiring scattered "fix-up" validation passes. DWScript's forward declarations and circular type references require proper multi-pass design.

---

### ⚠️ Development Mode: Experimental Passes

The new multi-pass system (Pass 2 + Pass 3) is **disabled by default** to keep the main branch stable. When working on task 6.1.2, you must explicitly enable experimental passes:

```go
// Default: Only old analyzer runs (stable, all tests pass)
analyzer := semantic.NewAnalyzer()

// For task 6.1.2 development: Enable Pass 2 + Pass 3
analyzer := semantic.NewAnalyzerWithExperimentalPasses()
```

**Why?** Running both old and new analyzers causes duplicate/conflicting validation errors, failing ~87 tests. This switch allows development on the new passes without breaking CI on main.

**When to enable:** Use `NewAnalyzerWithExperimentalPasses()` when:

- Working on any 6.1.2.x subtask
- Writing tests for the new pass system
- Debugging dual-mode validation conflicts

---

### Completed Work Summary

**Infrastructure** (100% complete):

- ✅ Created Pass interface (`Name()`, `Run(program, ctx)`) and PassManager
- ✅ Created PassContext with TypeRegistry, SymbolTable, BuiltinChecker interface, ScopedSymbolTable
- ✅ Designed 4-pass architecture: Declaration → Type Resolution → Validation → Contracts
- ✅ Integrated passes into Analyzer with dual-mode execution (old + new passes run together)
- ✅ All pass infrastructure files created and tested

**Pass 1: Declaration Collection** (100% complete):

- ✅ Registers all type names (classes, interfaces, enums, records, sets, arrays, type aliases)
- ✅ Registers function/procedure signatures without analyzing bodies
- ✅ Handles forward declarations by marking types as incomplete
- ✅ Unit tests passing

**Pass 2: Type Resolution** (100% complete):

- ✅ Resolves all type references (class parents, interface implementations, field types, method signatures)
- ✅ Builds complete type hierarchies (inheritance chains)
- ✅ Detects circular type dependencies and reports errors
- ✅ Validates forward-declared types/methods/functions have implementations
- ✅ Enabled in dual-mode alongside old analyzer
- ✅ Unit tests passing

**Pass 3: Semantic Validation** (95.2% complete - 1716/1803 tests):

- ✅ Created ValidationPass with comprehensive expression/statement type checking
- ✅ Migrated all major validation logic from old analyzer:
  - Expression validation (20+ types: literals, binary/unary ops, calls, member access, etc.)
  - Statement validation (15+ types: loops, conditionals, declarations, assignments, etc.)
  - Control flow validation (break/continue in loops, return paths in functions)
  - OOP validation (class inheritance, interface implementations, abstract methods, visibility)
  - Context-aware type inference (Nil, Integer→Float promotion, Record/Array/Set literals)
  - Helper method/property support (array.Pop, scoped enum access)
  - Variadic function/method parameter support
  - Compound assignment operators (+=, -=, *=, /=)
- ✅ Implemented ScopedSymbolTable with parent scope chain for local variables
- ✅ Enabled in dual-mode alongside old analyzer
- ✅ 95.2% test pass rate (87 failing tests are edge cases)

**Pass 4: Contract Validation** (90% complete):

- ✅ Created ContractPass for requires/ensures/invariant checking
- ✅ Validates postconditions only use valid 'old' expressions
- ✅ Validates requires clauses
- ✅ Validates class invariants (AST support added, parser extended, validation implemented)
- ⚠️ Some edge case integration testing remaining
- ✅ 5 unit tests passing

**Class Invariant Support** (100% complete):

- ✅ Added InvariantClause AST node and Invariants field to ClassDecl
- ✅ Extended parser to recognize 'invariants' keyword
- ✅ Integrated invariant validation into ContractPass
- ✅ Parser tests passing
- Note: Runtime invariant checking (executing after constructors/methods) deferred to future task

**Files Created**:

- `internal/semantic/pass.go`, `pass_context.go`
- `internal/semantic/declaration_pass.go`, `type_resolution_pass.go`, `validation_pass.go`, `contract_pass.go`
- `internal/semantic/passes/*_test.go`

**Files Modified**:

- `internal/semantic/analyzer.go` (dual-mode integration, removed legacy validation methods)

---

### Remaining Tasks (87 failing tests, ~60-75 hours)

**Dependencies**: Tasks are ordered by logical dependencies. Type system tasks must come before features that use them.

---

#### Foundation: Type System Completion (Must complete first)

- [x] **6.1.2.1: Function Pointer Type Resolution** (7 tests, 4-5 hours) ✅ COMPLETE
  - Extend `resolveTypeExpression` to handle `ast.FunctionPointerType`
  - Map to `types.FunctionPointerType` from context.FunctionPointers
  - Validate function pointer assignments (signature compatibility)
  - Validate function pointer calls (argument types, return type)
  - Handle implicit function-to-pointer conversion
  - **Blocks**: Task 6.1.2.10 (Lambda validation needs function pointer types)
  - **Files**: `validation_pass.go`, `pass_context.go`, `analyzer.go`, `analyze_builtin_functions.go`

- [x] **6.1.2.2: Set Type Support** (6 tests, 4-5 hours) ✅ COMPLETE
  - Extend `resolveTypeExpression` to handle `ast.SetType`
  - Map to `types.SetType` from context
  - Implement set literal type inference
  - Validate set membership operators (`in`)
  - Validate set operations (`+`, `-`, `*` for union/difference/intersection)
  - Handle large set optimization detection
  - **Files**: `validation_pass.go`

- [x] **6.1.2.3: Subrange Types** (1 test, 2-3 hours) ✅ COMPLETE
  - Extend type resolution for subrange types
  - Validate subrange assignment compatibility
  - Check value ranges
  - **Files**: `type_resolution_pass.go`, `validation_pass.go`

- [x] **6.1.2.4: Interface Type Validation** (4 tests, 2-3 hours) ✅ COMPLETE
  - Validate interface variable assignment
  - Check class-to-interface assignment compatibility
  - Check interface-to-interface assignment
  - Validate interface types in function returns
  - Ensure interface method compatibility
  - **Blocks**: Task 6.1.2.5 (type operators need interface validation)
  - **Files**: `validation_pass.go`, `interface_validation_experimental_test.go`

- [x] **6.1.2.5: Type Operators (is/as/implements)** (10 tests, 4-5 hours)
  - Add support for `is` operator (runtime type checking)
  - Add support for `as` operator (type casting with check)
  - Add support for `implements` operator (interface checking)
  - Validate operand types for each operator
  - Handle chained type operators
  - Support type operators in conditions and assignments
  - **Depends on**: Task 6.1.2.4 (needs interface validation)
  - **Files**: `validation_pass.go`

- [x] **6.1.2.6: Type Conversion/Coercion** (4 tests, 3-4 hours)
  - Support implicit numeric conversions (Integer → Float)
  - Handle explicit conversion operators
  - Validate function call with type coercion
  - Support both implicit and explicit conversions
  - Integer literal to Float context inference
  - **Files**: `validation_pass.go`

---

#### Parameters & Variables

- [ ] **6.1.2.7: Const Parameter Handling** (4 tests, 2-3 hours)
  - Add `IsConst` flag tracking to parameter validation
  - Validate that const parameters aren't modified in function body
  - Track assignment targets to detect const violations
  - **Files**: `validation_pass.go`

- [ ] **6.1.2.8: Lazy Parameters** (2 tests, 2-3 hours)
  - Add lazy parameter modifier support
  - Validate lazy parameters aren't evaluated eagerly
  - Support mixing lazy and regular parameters
  - **Files**: `validation_pass.go`

- [ ] **6.1.2.9: Class Variables (Static Fields)** (6 tests, 3-4 hours)
  - Extend `checkMemberAccessExpression` to handle static field access via class name
  - Add `ClassType.ClassVars` lookup in member access
  - Validate visibility rules for class variables
  - Handle class variable assignment
  - **Files**: `validation_pass.go`

---

#### Expressions & Statements

- [ ] **6.1.2.10: Lambda Expression Full Validation** (7 tests, 5-6 hours)
  - Implement lambda type inference from context
  - Handle lambda parameter scoping properly
  - Validate lambda return type matches inferred/declared type
  - Support variadic lambda inference
  - Implement closure capture analysis
  - **Depends on**: Task 6.1.2.1 (needs function pointer types)
  - **Files**: `validation_pass.go`

- [ ] **6.1.2.11: Array of Const Literals** (5 tests, 3-4 hours)
  - Handle `array of const` type expressions
  - Validate array literal type inference for const arrays
  - Check element type compatibility (homogeneous vs heterogeneous)
  - Handle empty array literals
  - Support type aliases for array of const
  - **Files**: `validation_pass.go`

- [ ] **6.1.2.12: Exception Variable Scoping** (2 tests, 2-3 hours)
  - Handle exception variable scope in try/except blocks
  - Validate exception variable types
  - Ensure exception variables are only visible in except block
  - Support break/continue with exception handling
  - **Files**: `validation_pass.go`

- [ ] **6.1.2.13: For-In Loop Edge Cases** (4 tests, 2-3 hours)
  - Handle inline variable declarations in for-in loops
  - Ensure for-in works with break/continue
  - Support nested for-in loops properly
  - Fix loop variable scoping for for-in
  - **Files**: `validation_pass.go`

---

#### OOP Features

- [ ] **6.1.2.14: Implicit Constructor Support** (2 tests, 2-3 hours)
  - Detect classes without explicit constructors
  - Generate implicit default constructor validation
  - Handle field initialization in implicit constructors
  - **Files**: `validation_pass.go`

- [ ] **6.1.2.15: Operator Overloads** (2 tests, 2-3 hours)
  - Validate class operator overload declarations
  - Check operator overload usage in expressions
  - Handle compound assignment with overloaded operators
  - **Files**: `validation_pass.go`

---

#### Builtins & Miscellaneous

- [ ] **6.1.2.16: Full Builtin Function Validation** (15 tests, 5-6 hours)
  - Extend builtin validation beyond basic cases
  - Handle builtins with array parameters (Low, High, SetLength, etc.)
  - Validate builtin calls in function contexts
  - Support builtin chaining
  - Handle conversion builtins (IntToStr, StrToInt, etc.)
  - Validate DateTime, JSON, Math, String builtins
  - **Note**: Most builtin validation already works via BuiltinChecker interface
  - **Files**: `validation_pass.go`, potentially new `builtin_validator.go`

- [ ] **6.1.2.17: Miscellaneous Fixes** (4 tests, 2-3 hours)
  - Implement AddressOf operator (`@`) validation
  - **Method Pointer Support**: Currently reports "method pointers (@TClass.Method) not yet implemented"
    - Need to support `@obj.Method` syntax for creating method pointers
    - Need to infer function pointer type from method signature
    - Runtime already handles method pointers correctly (see `TestMethodPointer_*` tests in `method_pointer_test.go`)
    - Blocked contexts: `funcs.Add(@obj.Get)`, array storage of method pointers
  - Fix case-insensitive method call edge cases
  - Complete contract `old()` expression validation edge cases
  - **Files**: `validation_pass.go`, `contract_pass.go`, `analyze_expressions.go`

---

#### Integration & Cleanup

- [ ] **6.1.2.18: Complete Pass 4 (Contract Pass) Migration** (2-4 hours)
  - Fix remaining edge cases in contract validation
  - Integration testing with complex scenarios
  - Ensure all contract tests pass
  - **Files**: `contract_pass.go`

- [ ] **6.1.2.19: Enable All 4 Passes** (2-4 hours)
  - Replace old analyzer loop with PassManager running all 4 passes
  - Verify all 1803 tests pass
  - Handle any edge cases discovered during full integration
  - **Depends on**: All tasks 6.1.2.1-6.1.2.18 complete
  - **Files**: `analyzer.go`

- [ ] **6.1.2.20: Remove Old Analyzer Loop** (2-4 hours)
  - Delete `analyzeStatement()` and all related old validation code
  - Remove dual-mode infrastructure
  - Clean up dead code from `analyze_*.go` files
  - Update documentation to reflect new architecture
  - **Depends on**: Task 6.1.2.19 complete
  - **Files**: `analyzer.go`, various `analyze_*.go` files

- [ ] **6.1.2.21: Add Pass Result Caching** (8-12 hours, DEFERRED)
  - Cache Pass 1 results for incremental re-analysis
  - Cache Pass 2 type resolution for unchanged files
  - Design cache invalidation strategy (changed files invalidate dependent passes)
  - Write tests for caching behavior
  - **Note**: Deferred until after multi-pass architecture is fully enabled
  - **Files**: New `pass_cache.go`

---

### Current Architecture (Dual-Mode)

```go
func (a *Analyzer) Analyze(program *ast.Program) error {
    // OLD: Analyze each statement (handles 4.8% edge cases)
    for _, stmt := range program.Statements {
        a.analyzeStatement(stmt)  // Registration + validation
    }

    // NEW: Run Pass 2 + Pass 3 (handles 95.2% of validation)
    ctx := a.createPassContext()
    pass2 := NewTypeResolutionPass()
    pass2.Run(program, ctx)
    pass3 := NewValidationPass()
    pass3.Run(program, ctx)
    a.mergePassErrors(ctx)

    return ...
}
```

### Target Architecture (Full Multi-Pass)

```go
func (a *Analyzer) Analyze(program *ast.Program) error {
    ctx := a.createPassContext()

    // Run all 4 passes sequentially
    pass1 := NewDeclarationPass()     // Register types/functions
    pass2 := NewTypeResolutionPass()  // Resolve types, validate forwards
    pass3 := NewValidationPass()      // Type-check expressions/statements
    pass4 := NewContractPass()        // Validate contracts

    passManager := NewPassManager(pass1, pass2, pass3, pass4)
    passManager.RunAll(program, ctx)

    a.mergePassErrors(ctx)
    return ...
}
```

---

### Success Metrics

- ✅ Four explicit passes with clear responsibilities (infrastructure complete)
- ✅ Forward declarations validated reliably (Pass 2 complete)
- ✅ Post-hoc validation methods removed (migrated to Pass 2)
- ✅ All semantic tests pass (187 unit tests)
- ✅ 95.2% integration test pass rate (1716/1803 tests)
- ⚠️ Performance maintained (dual-mode has overhead, will improve after task 6.1.2.20)
- ⚠️ Circular dependencies detected (partial - type resolution only, full detection in Pass 3)
- ❌ Full pass architecture not yet enabled (blocked on remaining 87 tests)

**Total Remaining Effort**: 60-75 hours (7-9 days full-time work)

---

## Task 6.1.3: Refactor Symbol Table for LSP Support 🎯 **MEDIUM PRIORITY**

**Goal**: Enhance `SymbolTable` to track symbol positions, usage locations, documentation, and metadata required for Language Server Protocol features.

**Estimate**: 12-16 hours (1.5-2 days)

**Status**: NOT STARTED

**Priority**: P1 - HIGH (Required for LSP, but can defer if necessary)

**Current Problem**:
The `Symbol` struct (symbol_table.go:10-21) is too simplistic:
```go
type Symbol struct {
    Type                 types.Type
    Value                interface{}  // Using interface{} for constants
    Name                 string
    // ... flags ...
    // Missing: Position, Usage locations, Documentation, Deprecation info
}
```

This prevents:
- "Go to definition" (no position tracking)
- "Find all references" (no usage tracking)
- Hover documentation (no doc strings)
- Unused variable warnings (no usage count)
- Symbol rename refactoring (no usage locations)

**Insight from Analysis**: Symbol table is only 638 lines and lacks critical metadata for modern IDE features. No position tracking means LSP features impossible without major rework.

**Subtasks**:

- [ ] **6.1.3.1 Enhance Symbol struct**
  - [ ] Add `DeclPosition lexer.Position` field
  - [ ] Add `Usages []lexer.Position` slice for tracking references
  - [ ] Add `Documentation string` field for doc comments
  - [ ] Add `IsDeprecated bool` and `DeprecationMessage string`
  - [ ] Add `Attributes map[string]string` for custom metadata
  - [ ] Update all `Define*` methods to accept position parameter
  - [ ] Write unit tests for enhanced symbol structure

- [ ] **6.1.3.2 Track symbol usage locations**
  - [ ] In `Resolve()`, record usage position when symbol found
  - [ ] Add `RecordUsage(name string, pos lexer.Position)` method
  - [ ] Track both reads and writes separately (for unused write detection)
  - [ ] Write tests for usage tracking

- [ ] **6.1.3.3 Add symbol query methods**
  - [ ] Implement `FindDefinition(name string) (*Symbol, lexer.Position, bool)`
  - [ ] Implement `FindReferences(name string) []lexer.Position`
  - [ ] Implement `SymbolsInScope(pos lexer.Position) []*Symbol`
  - [ ] Implement `UnusedSymbols() []*Symbol` (symbols defined but never used)
  - [ ] Write unit tests for query methods

- [ ] **6.1.3.4 Integrate with analyzer**
  - [ ] Update all `Define()` calls to pass AST node position
  - [ ] Update all `Resolve()` calls to record usage position
  - [ ] Extract documentation from preceding comments during parsing
  - [ ] Store symbol documentation in Symbol struct
  - [ ] Write integration tests

- [ ] **6.1.3.5 Add scope metadata**
  - [ ] Add `ScopeKind` enum: Global, Function, Class, Block
  - [ ] Track scope boundaries (start/end positions)
  - [ ] Add `GetScopeAt(pos lexer.Position) *SymbolTable` method
  - [ ] Write tests for scope querying

**Files to Modify**:

- `internal/semantic/symbol_table.go` (enhance Symbol and SymbolTable)
- `internal/semantic/analyzer.go` (pass positions to symbol operations)
- All `analyze_*.go` files (pass positions when defining/resolving symbols)

**Files to Create**:

- `internal/semantic/symbol_query.go` (LSP query methods)
- `internal/semantic/symbol_query_test.go`

**Success Metrics**:

- Every symbol has position information
- Usage tracking works for all references
- Query methods return correct results
- All existing tests pass
- New tests for LSP queries at 85%+ coverage

---

## Task 6.1.4: Consolidate Error Handling 🎯 **MEDIUM PRIORITY**

**Goal**: Remove dual error system (string + structured) and commit to single, consistent structured error approach throughout semantic analyzer.

**Estimate**: 8-12 hours (1-1.5 days)

**Status**: NOT STARTED

**Priority**: P1 - HIGH (Code quality and consistency)

**Current Problem**:

Analyzer has TWO error systems (analyzer.go:72-73, 456-472):

```go
type Analyzer struct {
    errors           []string            // String errors
    structuredErrors []*SemanticError    // Structured errors
}

func (a *Analyzer) addError(format string, args ...any) {
    a.errors = append(a.errors, fmt.Sprintf(format, args...))
}

func (a *Analyzer) addStructuredError(err *SemanticError) {
    a.structuredErrors = append(a.structuredErrors, err)
    a.errors = append(a.errors, err.Error())  // Duplicate!
}
```

Problems:
- Most code uses `addError()` (string-based)
- Structured errors are rarely used (incomplete migration)
- Duplicate storage wastes memory
- Inconsistent error reporting to CLI/LSP
- String errors lose position and context information

**Insight from Analysis**: Lines 456-472 show technical debt from incomplete refactoring. String errors dominate codebase. Structured errors would enable better IDE integration (error decorations, quick fixes).

**Subtasks**:

- [ ] **6.1.4.1 Audit error call sites**
  - [ ] Find all `addError()` calls across semantic analyzer
  - [ ] Categorize errors: type mismatch, undefined symbol, duplicate declaration, etc.
  - [ ] Document required information for each error category
  - [ ] Create spreadsheet/document of error patterns

- [ ] **6.1.4.2 Design comprehensive SemanticError types**
  - [ ] Create error types: `TypeMismatchError`, `UndefinedSymbolError`, `DuplicateDeclarationError`
  - [ ] Each error type includes: position, context, suggested fix (if applicable)
  - [ ] Add `ErrorCode` enum for programmatic error handling
  - [ ] Support error severity: Error, Warning, Hint
  - [ ] Write unit tests for error construction

- [ ] **6.1.4.3 Replace addError() calls incrementally**
  - [ ] Start with most common errors (type mismatch, undefined symbol)
  - [ ] Replace `addError()` calls with `addStructuredError()` calls
  - [ ] Update call sites to provide position and context
  - [ ] Run tests after each file to ensure correctness
  - [ ] Continue until all `addError()` calls removed

- [ ] **6.1.4.4 Remove string error system**
  - [ ] Delete `errors []string` field from Analyzer
  - [ ] Delete `addError()` method
  - [ ] Rename `addStructuredError()` to `addError()`
  - [ ] Rename `StructuredErrors()` to `Errors()`
  - [ ] Update all callers to use new API

- [ ] **6.1.4.5 Enhance error formatting**
  - [ ] Improve `SemanticError.Error()` to include file, line, column
  - [ ] Add color-coded error output for CLI
  - [ ] Add source code snippet with error location highlighted
  - [ ] Support JSON error format for LSP integration
  - [ ] Write tests for error formatting

**Files to Modify**:

- `internal/semantic/errors.go` (add error types)
- `internal/semantic/analyzer.go` (remove string errors)
- All `analyze_*.go` files (replace addError calls)
- `cmd/dwscript/commands/*.go` (update error display)

**Success Metrics**:

- Zero `addError()` calls remain
- All errors are structured with positions
- Error messages improved with context
- CLI shows beautiful, helpful error messages
- All tests pass

---

## Task 6.1.5: Extract Builtin Library from Semantic Analyzer 🎯 **MEDIUM PRIORITY**

**Goal**: Move builtin function implementations out of semantic analyzer into separate `internal/stdlib` package. Analyzer should only validate signatures, not implement functions.

**Estimate**: 12-16 hours (1.5-2 days)

**Status**: NOT STARTED

**Priority**: P1 - HIGH (Separation of concerns)

**Current Problem**:
Files like these are in semantic analyzer:

- `analyze_builtin_math.go` (23KB) - mathematical functions
- `analyze_builtin_string_transform.go` - string manipulation
- `analyze_builtin_datetime.go` - date/time functions
- `analyze_builtin_json.go` - JSON parsing
- `analyze_builtin_convert.go` - type conversions

These files contain **runtime implementations**, not just signature validation. This violates separation of concerns.

**Insight from Analysis**: The semantic analyzer is 20,000 lines, with ~8,000 lines being builtin function implementations. These should live in interpreter/stdlib, not analyzer. Analyzer should only know "function X exists with signature Y".

**Subtasks**:

- [ ] **6.1.5.1 Design builtin registry**
  - [ ] Create `internal/stdlib/registry.go`
  - [ ] Define `BuiltinFunc` struct with: Name, Signature, Implementation
  - [ ] Create `BuiltinRegistry` with registration and lookup methods
  - [ ] Support categories: Math, String, DateTime, JSON, Array, Type Conversion
  - [ ] Write unit tests for registry

- [ ] **6.1.5.2 Extract math builtins**
  - [ ] Move functions from `analyze_builtin_math.go` to `internal/stdlib/math.go`
  - [ ] Register signatures with semantic analyzer
  - [ ] Move implementations to interpreter
  - [ ] Update tests to use new structure
  - [ ] Verify all math function tests pass

- [ ] **6.1.5.3 Extract string builtins**
  - [ ] Move from `analyze_builtin_string_*.go` to `internal/stdlib/string.go`
  - [ ] Register signatures with analyzer
  - [ ] Move implementations to interpreter
  - [ ] Update tests

- [ ] **6.1.5.4 Extract remaining builtins**
  - [ ] Extract datetime functions to `internal/stdlib/datetime.go`
  - [ ] Extract JSON functions to `internal/stdlib/json.go`
  - [ ] Extract conversion functions to `internal/stdlib/convert.go`
  - [ ] Extract array functions to `internal/stdlib/array.go`
  - [ ] Extract encoding functions to `internal/stdlib/encoding.go`

- [ ] **6.1.5.5 Update analyzer to use registry**
  - [ ] In `NewAnalyzer()`, register all builtin signatures from registry
  - [ ] Remove implementation code from analyzer
  - [ ] Analyzer only validates: "does this builtin exist? do arguments match?"
  - [ ] Delete old `analyze_builtin_*.go` files from semantic package
  - [ ] Update all tests

- [ ] **6.1.5.6 Update interpreter integration**
  - [ ] Modify interpreter to get builtin implementations from stdlib registry
  - [ ] Remove duplicate builtin code from interpreter
  - [ ] Ensure all builtin tests pass
  - [ ] Verify no performance regression

**Files to Create**:

- `internal/stdlib/registry.go`
- `internal/stdlib/math.go`
- `internal/stdlib/string.go`
- `internal/stdlib/datetime.go`
- `internal/stdlib/json.go`
- `internal/stdlib/convert.go`
- `internal/stdlib/array.go`
- `internal/stdlib/encoding.go`

**Files to Delete** (after migration):

- `internal/semantic/analyze_builtin_math.go`
- `internal/semantic/analyze_builtin_string_transform.go`
- `internal/semantic/analyze_builtin_string_format.go`
- `internal/semantic/analyze_builtin_string_search.go`
- `internal/semantic/analyze_builtin_datetime.go`
- `internal/semantic/analyze_builtin_json.go`
- `internal/semantic/analyze_builtin_convert.go`
- `internal/semantic/analyze_builtin_array.go`
- `internal/semantic/analyze_builtin_encoding.go`
- `internal/semantic/analyze_builtin_variant.go`

**Files to Modify**:
- `internal/semantic/analyzer.go` (use registry instead of direct implementation)
- `internal/interp/interpreter.go` (get builtins from stdlib)

**Success Metrics**:
- Semantic analyzer is ~12,000 lines instead of 20,000
- All builtin implementations in single `internal/stdlib` package
- Analyzer only validates signatures
- All tests pass
- No performance regression

---

## Task 6.1.6: Replace State Flags with Context Stack 🎯 **LOW PRIORITY**

**Goal**: Replace fragile boolean state flags (`inLoop`, `inLambda`, etc.) with proper context stack that automatically manages state transitions.

**Estimate**: 8-10 hours (1 day)

**Status**: NOT STARTED

**Priority**: P2 - MEDIUM (Code quality improvement, not blocking)

**Current Problem**:
Analyzer has 7 fragile state flags (analyzer.go:74-81):
```go
loopDepth          int
inExceptionHandler bool
inFinallyBlock     bool
inLoop             bool
inLambda           bool
inClassMethod      bool
inPropertyExpr     bool
```

Problems:
- Must manually set/unset in every function
- Easy to forget cleanup (especially in error paths)
- Breaks with nested contexts (lambda inside loop inside class method)
- No automatic restoration on scope exit
- Difficult to add new context types

**Insight from Analysis**: No automatic cleanup mechanism. Nested contexts (lambda in loop) prone to bugs. Should use defer pattern or context stack.

**Subtasks**:

- [ ] **6.1.6.1 Design AnalysisContext stack**
  - [ ] Create `AnalysisContext` struct with: Kind (Loop/Lambda/ClassMethod/etc.), Data map[string]any
  - [ ] Create `ContextStack` with `Push()`, `Pop()`, `Current()`, `InContext(kind)` methods
  - [ ] Support defer-based automatic cleanup: `defer ctx.Pop()`
  - [ ] Write unit tests for context stack

- [ ] **6.1.6.2 Replace loop flags**
  - [ ] Replace `loopDepth` and `inLoop` with context stack
  - [ ] Update loop analysis functions to push/pop loop context
  - [ ] Update break/continue validation to check context stack
  - [ ] Test nested loops

- [ ] **6.1.6.3 Replace lambda flags**
  - [ ] Replace `inLambda` with context stack
  - [ ] Update lambda analysis to push/pop lambda context
  - [ ] Store lambda signature in context data
  - [ ] Test nested lambdas

- [ ] **6.1.6.4 Replace class method flags**
  - [ ] Replace `inClassMethod` with context stack
  - [ ] Store current class type in context data
  - [ ] Update Self/inherited expression validation
  - [ ] Test methods in nested classes

- [ ] **6.1.6.5 Replace exception handler flags**
  - [ ] Replace `inExceptionHandler` and `inFinallyBlock` with context stack
  - [ ] Update exception statement validation
  - [ ] Test nested try/except/finally

- [ ] **6.1.6.6 Remove old flags**
  - [ ] Delete all boolean flags from Analyzer struct
  - [ ] Replace flag checks with `ctx.InContext(kind)` calls
  - [ ] Run full test suite
  - [ ] Verify nested context scenarios work correctly

**Files to Create**:
- `internal/semantic/analysis_context.go`
- `internal/semantic/analysis_context_test.go`

**Files to Modify**:
- `internal/semantic/analyzer.go` (replace flags with context stack)
- `internal/semantic/analyze_statements.go` (use context stack)
- `internal/semantic/analyze_lambdas.go` (use context stack)
- `internal/semantic/analyze_classes.go` (use context stack)

**Success Metrics**:
- All 7 boolean flags removed
- Context automatically cleaned up with defer
- Nested contexts work correctly
- All tests pass
- Code more maintainable

---

## Task 6.1.7: Migrate to Visitor Pattern for AST Traversal 🎯 **LOW PRIORITY**

**Goal**: Replace giant type switches with generated visitor pattern for cleaner, more maintainable AST traversal.

**Estimate**: 12-16 hours (1.5-2 days)

**Status**: NOT STARTED

**Priority**: P2 - MEDIUM (Code quality, not urgent)

**Current Problem**:
AST traversal uses giant type switches (analyze_statements.go:18-88, analyze_expressions.go:14-98):
```go
func (a *Analyzer) analyzeExpression(expr ast.Expression) types.Type {
    switch e := expr.(type) {
    case *ast.IntegerLiteral:
        // ...
    case *ast.FloatLiteral:
        // ...
    // ... 30+ cases ...
    }
}
```

Problems:

- Difficult to maintain (easy to miss cases)
- No compile-time guarantee all node types handled
- Hard to add new node types
- Traversal logic mixed with analysis logic
- Can't easily customize traversal order

**Insight from Analysis**: Project already has generated visitor pattern in `pkg/ast/visitor_generated.go`. Should use it instead of manual type switches.

**Subtasks**:

- [ ] **6.1.7.1 Study generated visitor**
  - [ ] Review `pkg/ast/visitor_generated.go` implementation
  - [ ] Understand `Visitor` interface and `Walk*` functions
  - [ ] Identify how to integrate with semantic analyzer
  - [ ] Document visitor usage patterns

- [ ] **6.1.7.2 Create SemanticVisitor**
  - [ ] Implement `Visitor` interface in `internal/semantic/visitor.go`
  - [ ] Each `Visit*` method delegates to appropriate analyze function
  - [ ] Maintain analyzer state (symbol table, type registry, errors)
  - [ ] Write unit tests for visitor

- [ ] **6.1.7.3 Refactor expression analysis**
  - [ ] Replace `analyzeExpression()` type switch with visitor dispatch
  - [ ] Keep analysis logic in separate methods (e.g., `analyzeBinaryOp()`)
  - [ ] Visitor calls analysis methods instead of inlining logic
  - [ ] Update tests

- [ ] **6.1.7.4 Refactor statement analysis**
  - [ ] Replace `analyzeStatement()` type switch with visitor dispatch
  - [ ] Keep analysis logic in separate methods
  - [ ] Update tests

- [ ] **6.1.7.5 Remove old type switches**
  - [ ] Delete manual type switches from analyze_*.go files
  - [ ] Ensure all node types covered by visitor
  - [ ] Run full test suite
  - [ ] Verify no regressions

**Files to Create**:

- `internal/semantic/visitor.go`
- `internal/semantic/visitor_test.go`

**Files to Modify**:

- `internal/semantic/analyze_expressions.go` (use visitor)
- `internal/semantic/analyze_statements.go` (use visitor)
- Other analyze_*.go files as needed

**Success Metrics**:
- No type switches for AST traversal
- Visitor handles all node types
- Compile-time safety for node coverage
- All tests pass
- Code more maintainable

---

## Phase 6.1 Summary

**Total Estimate**: 80-120 hours (2-3 weeks full-time)

**Completion Criteria**:
- [ ] Unified TypeRegistry replaces 7 scattered type maps
- [ ] Multi-pass architecture with clear separation
- [ ] Enhanced SymbolTable with LSP metadata
- [ ] Single structured error system
- [ ] Builtin functions extracted to stdlib
- [ ] Context stack replaces boolean flags
- [ ] Visitor pattern for AST traversal
- [ ] All existing tests pass
- [ ] Documentation updated
- [ ] Semantic analyzer reduced from 20,000 to ~12,000 lines
- [ ] Maintainability score improved significantly

**Dependencies**:
- Task 6.1.1 (TypeRegistry) must complete before 6.1.2 (multi-pass)
- Task 6.1.2 (multi-pass) must complete before other refactorings
- Tasks 6.1.3-6.1.7 can proceed in parallel after 6.1.2

**Risk Mitigation**:
- Comprehensive test suite ensures correctness during refactoring
- Incremental approach allows rolling back if issues found
- Each task is independently testable
- No public API changes (all internal refactoring)

---

## Stage 7: Support Object-Oriented Features (Classes, Interfaces, Methods) ✅ **COMPLETED**

- Extended the type system and AST with class/interface nodes, constructors/destructors, member access, `Self`, `NewExpression`, and external declarations (see docs/stage7-summary.md).
- Parser handles class/interface declarations, inheritance chains, interface lists, constructors, member access, and method calls with comprehensive unit tests and fixtures.
- Added runtime class metadata plus interpreter support for object creation, field storage, method dispatch, constructors, destructors, and interface casting with ~98% targeted coverage.
- Semantic analysis validates class/interface hierarchies, method signatures, interface fulfillment, and external contracts while integrating with the existing symbol/type infrastructure.
- Documentation in docs/stage7-summary.md, docs/stage7-complete.md, docs/delphi-to-go-mapping.md, and docs/interfaces-guide.md captures the architecture, and CLI/integration suites ensure DWScript parity.

## Stage 8: Additional DWScript Features and Polishing ✅ **IN PROGRESS**

- Extended the expression/type system with DWScript-accurate operator overloading (global + class operators, coercions, analyzer enforcement) and wired the interpreter/CLI with matching fixtures in `testdata/operators` and docs in `docs/operators.md`.
- Landed the full property stack (field/method/auto/default metadata, parser, semantic validation, interpreter lowering, CLI coverage) so OO code can use DWScript-style properties; deferred indexed/expr variants are tracked separately.
- Delivered composite type parity: enums, records, sets, static/dynamic arrays, and assignment/indexing semantics now mirror DWScript with dedicated analyzers, runtime values, exhaustive unit/integration suites, and design notes captured in docs/enums.md plus related status writeups.
- Upgraded the runtime to support break/continue/exit statements, DWScript's `new` keyword, rich exception handling (try/except/finally, raise, built-in exception classes), and CLI smoke tests that mirror upstream fixtures.
- Backfilled fixtures, docs, and CLI suites for every feature shipped in this phase (properties, enums, arrays, exceptions, etc.), keeping coverage high and mapping each ported DWScript test in `testdata/properties/REFERENCE_TESTS.md`.

---

## Phase 9: Completion and DWScript Feature Parity


## Task 9.5: Implement Class Variables (class var)

**Goal**: Add support for class variables (`class var`) in class declarations.

**Estimate**: 12-15 hours (1.5-2 days)

**Status**: DONE

**Impact**: Unlocks 11 failing tests in SimpleScripts (3 now passing: class_var, class_var_dyn1, static_method3)

**Priority**: P0 - CRITICAL (Core OOP feature)

**Description**: Class variables are static class members that belong to the class itself rather than instances. In DWScript, they are declared with the `class var` keyword and can be accessed via the class name (e.g., `TBase.Test`) or via instances. Currently:

- Parser doesn't recognize `class var` syntax in class bodies
- Type system doesn't track class variables separately from instance fields
- Semantic analyzer and interpreter have no support for class variable access

**Failing Tests** (11 total):

- class_var
- class_var_as_prop
- class_var_dyn1
- class_var_dyn2
- class_method3
- class_method4
- static_class
- static_class_array
- static_method1
- static_method2
- field_scope

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P0, Section 2

**Example**:
```pascal
type TBase = class
  class var Test : Integer;
end;

TBase.Test := 123;  // Access via class name
var b : TBase;
PrintLn(b.Test);    // Access via instance
```

**Subtasks**:

### 9.5.1 Parse Class Variable Declarations

**Goal**: Update parser to recognize `class var` syntax in class bodies.

**Estimate**: 3-4 hours

**Status**: DONE

**Implementation**:

1. Add parsing for `class var` keyword in class body
2. Create AST node to represent class variables (or flag in existing FieldDeclaration)
3. Handle multiple class var declarations
4. Support initialization syntax: `class var Test : Integer := 123;`

**Files Modified**:
- `internal/parser/classes.go` (lines 229-238: class var parsing already implemented)
- `pkg/ast/classes.go` (line 196: IsClassVar flag already exists in FieldDecl)
- `internal/parser/class_var_init_test.go` (comprehensive tests already exist)
- `internal/parser/parser_classvar_test.go` (added additional verification tests)

**Notes**:

- Parser implementation for class var was already complete
- All tests pass, including initialization, type inference, and visibility modifiers
- Supports both `class var X: Type;` and `class var X: Type := Value;` syntax

### 9.5.2 Type System Support for Class Variables

**Goal**: Extend ClassType to track class variables separately from instance fields.

**Estimate**: 3-4 hours

**Status**: DONE

**Implementation**:
1. Add `ClassVars map[string]*types.Type` field to ClassType
2. During semantic analysis of class declarations, populate ClassVars
3. Handle class variable inheritance (child classes inherit parent's class vars)
4. Validate class variable initialization

**Files Already Modified**:
- `internal/types/types.go` (line 424: ClassVars field already exists)
- `internal/semantic/analyze_classes_decl.go` (lines 291-349: class var handling already implemented)
- `internal/semantic/type_resolution.go` (lines 866-883: addParentClassVarsToScope method)

**Notes**:
- Type system implementation for class vars was already complete
- ClassType.ClassVars map stores class variable types separately from instance Fields
- Semantic analyzer properly:
  - Detects duplicate class variable declarations
  - Handles explicit type annotations
  - Supports type inference from initialization values
  - Validates type compatibility for initializations
  - Stores class variable types in ClassVars map
- Class variable inheritance implemented via addParentClassVarsToScope:
  - Recursively adds parent class variables to method scopes
  - Allows shadowing (child class vars can hide parent class vars)
- All semantic analyzer tests pass, including:
  - TestClassVariable
  - TestClassVariableWithInvalidType
  - TestClassVariableAndInstanceField
  - TestClassMethodAccessingClassVariable

**Remaining Work**: Task 9.5.3 needed for class variable access via class name (e.g., `TBase.Test`)

### 9.5.3 Semantic Analysis for Class Variable Access

**Goal**: Support type checking for class variable access via class name or instance.

**Estimate**: 3-4 hours

**Status**: DONE

**Implementation**:
1. In member access expressions, check both instance fields and class vars
2. For type name access (e.g., `TBase.Test`), look up class vars
3. Validate read/write access to class variables
4. Type check assignments to class variables

**Files Modified**:
- `internal/types/types.go` (lines 689-705: added GetClassVar method to ClassType)
- `internal/semantic/analyze_classes.go` (lines 310-314: added class var lookup in analyzeMemberAccessExpression)
- `internal/semantic/class_var_access_test.go` (new file: comprehensive tests for class var access)

**Notes**:
- Added GetClassVar method to ClassType with inheritance support (recursively checks parent classes)
- Integrated class variable lookup into member access expression analysis
- Class variables can now be accessed via:
  - Class name: `TBase.Test`
  - Instance: `obj.Test` (even if obj is nil, since class vars belong to the class)
- Semantic analysis validates:
  - Class variable existence (with proper error messages)
  - Type compatibility for assignments
  - Inheritance (child classes can access parent class variables)
  - Shadowing (child class vars can override parent class vars)
- All semantic analyzer tests pass (11 new tests added):
  - Access via class name and instance
  - Inheritance and shadowing
  - Type checking and error handling

**Remaining Work**: Task 9.5.4 needed for runtime execution of class variable access and assignments

### 9.5.4 Runtime Support for Class Variables

**Goal**: Implement class variable storage and access in interpreter and bytecode VM.

**Estimate**: 3-4 hours

**Status**: DONE

**Implementation**:
1. Create global storage for class variables (separate from instances) ✓
2. Implement class variable initialization ✓
3. Handle class variable access via class name ✓
4. Handle class variable access via instance (lookup in class, not instance) ✓
5. Handle class variable access via type casts (TBase(child).ClassVar) ✓

**Files Modified**:
- `internal/interp/declarations.go` (lines 309-352: class var initialization)
- `internal/interp/objects_hierarchy.go` (lines 51-54, 227-230, 322-365, 417-432: class var access)
- `internal/interp/statements_assignments.go` (lines 488-494: class var assignment via class name)
- `internal/interp/fixture_test.go` (added SetSemanticInfo calls for tests)
- `internal/interp/class_var_test.go` (new file: runtime tests for class variables)
- `internal/interp/value.go` (lines 620-636: TypeCastValue for preserving static types)
- `internal/interp/functions_typecast.go` (lines 512-520, 529-535: wrap casts in TypeCastValue)

**What Works**:
- ✓ Class variables stored in `ClassInfo.ClassVars map[string]Value`
- ✓ Initialization with default values or explicit init expressions
- ✓ Access via class name: `TBase.Test` (read and write)
- ✓ Inheritance: `lookupClassVar` recursively checks parent classes
- ✓ Access via object instances: `obj.Test` where obj is non-nil
- ✓ Access via NIL instances: `var b : TBase; PrintLn(b.Test)` (uses NilValue.ClassType)
- ✓ Access via type casts: `TBase(child).ClassVar` uses static type, not runtime type
- ✓ Assignment via class name: `TBase.Test := 123`
- ✓ Tests: `class_var`, `class_var_dyn1`, `static_method3` all pass

**Implementation Details**:
- Created `TypeCastValue` wrapper to preserve static type from type casts
- Type casts now return `TypeCastValue{Object: obj, StaticType: targetClass}`
- Member access unwraps `TypeCastValue` and uses `StaticType` for class variable lookup
- This ensures `TBase(child).ClassVar` accesses `TBase.ClassVar`, not `TChild.ClassVar`

**Testing**:
```bash
# All pass:
go test -v ./internal/interp -run TestClassVariableAccessViaClassNameRuntime
go test -v ./internal/interp -run TestClassVariableAccessViaInstanceRuntime
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/class_var
```

**Note**: Task complete! Class variables fully functional including inheritance, nil access, and type cast scenarios. Remaining test failures (class_var_as_prop, static_method1/2, etc.) are due to other unimplemented features (properties with class var specifiers, static method syntax, field shadowing).

---

## Task 9.6: Enhance Class Constants with Field Initialization

**Goal**: Support field initialization from class constants in class body.

**Estimate**: 6-8 hours (0.75-1 day)

**Status**: DONE

**Impact**: Unlocks 7 failing tests in SimpleScripts (complements Task 9.2)

**Priority**: P0 - CRITICAL (Required for class patterns)

**Description**: Task 9.2 added basic class constant support, but DWScript also allows initializing fields directly from constants using syntax like `FField := Value;` inside the class body. This requires:
- Parsing field initialization syntax (currently fails with parse errors)
- Semantic analysis to resolve constant references in initializers
- Runtime support for field initialization during object creation

**Failing Tests** (7 total):
- class_const2 (semantic issues with const resolution)
- class_const3 (missing hints for case mismatch)
- class_const4 (parse error for field initialization syntax)
- class_const_as_prop (output mismatch)
- class_init (parse error)
- const_block (parse error)
- enum_element_deprecated

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P0, Section 3

**Example**:
```pascal
type TObj = class
  const Value = 5;
  FField := Value;  // Initialize field from constant
end;
```

**Subtasks**:

### 9.6.1 Parse Field Initialization Syntax

**Goal**: Update parser to handle `FField := Value;` syntax in class bodies.

**Estimate**: 2-3 hours

**Implementation**:

1. Modify class body parser to recognize `:=` after identifier
2. Create field with initialization expression in AST
3. Distinguish from method declarations and properties

**Files to Modify**:
- `internal/parser/parser_classes.go` (parse field initialization)
- `pkg/ast/declarations.go` (add Initializer field to FieldDeclaration)

### 9.6.2 Semantic Analysis for Field Initializers

**Goal**: Type check and resolve constant references in field initializers.

**Estimate**: 2-3 hours

**Implementation**:

1. During class declaration analysis, analyze field initializer expressions
2. Resolve references to class constants
3. Validate initializer types match field types
4. Report errors for invalid initializers

**Files to Modify**:
- `internal/semantic/analyze_classes_decl.go` (analyze field initializers)

### 9.6.3 Runtime Field Initialization

**Goal**: Execute field initializers during object creation.

**Estimate**: 2 hours

**Implementation**:
1. During object construction, evaluate field initializers
2. Apply initialized values to new instances
3. Handle initialization order (constants before field inits)

**Files to Modify**:
- `internal/interp/objects_creation.go` (execute field initializers)

**Success Criteria**:
- All 7 class const tests pass
- `FField := Value;` syntax parses correctly
- Fields are initialized from constants during object creation
- `new TObj.FField` returns the constant value

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/class_const4
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/class_init
```

---

## Task 9.12: Implement Record Advanced Features

**Goal**: Add field initialization, record constants, record class variables, and nested records.

**Estimate**: 14-18 hours (2-2.5 days)

**Status**: NOT STARTED

**Impact**: Unlocks 32 failing tests in SimpleScripts

**Priority**: P1 - IMPORTANT (Value type features)

**Description**: Records currently have basic support, but DWScript includes advanced features like field initialization syntax, record constants, class variables in records, nested records, and enhanced record methods. These features make records more powerful as value types.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P1, Section 8

**Failing Tests** (32 total):
- const_record
- record_aliased_field
- record_class_field_init
- record_clone1
- record_const
- record_const_as_prop
- record_dynamic_init
- record_field_init
- record_in_copy
- record_metaclass_field
- record_method, record_method2, record_method3, record_method4, record_method5
- record_nested, record_nested2
- record_property
- record_record_field_init
- record_result, record_result2, record_result3
- record_passing
- record_recursive_dynarray
- record_static_array
- record_var
- record_var_as_prop
- record_var_param1, record_var_param2
- result_direct
- string_record_field_get_set
- var_param_rec_field
- var_param_rec_method

**Example**:
```pascal
type
  TPoint = record
    X: Integer := 0;      // Field initialization
    Y: Integer := 0;

    const Origin = 0;     // Record constant
    class var Count: Integer;  // Class variable

    function Distance: Float;  // Record method
    begin
      Result := Sqrt(X*X + Y*Y);
    end;
  end;

  TRect = record
    TopLeft: TPoint;      // Nested record
    BottomRight: TPoint;
  end;

const
  DefaultPoint: TPoint = (X: 0; Y: 0);  // Record constant

var p: TPoint;
begin
  // Fields auto-initialized to 0
  PrintLn(p.Distance());
end;
```

**Complexity**: High - Multiple interrelated features

**Subtasks**:

### 9.12.1 Parse Record Field Initialization ✅ DONE

**Goal**: Support `FieldName: Type := Value;` syntax in records.

**Estimate**: 2-3 hours

**Status**: COMPLETED

**Implementation**:
1. ✅ Extended record field parsing to handle `:= <expression>` for type inference
2. ✅ Updated FieldDeclaration to use existing InitValue field
3. ✅ Parsing supports both explicit types and type inference
4. ✅ Handle type inference from initializers

**Files Modified**:
- `internal/parser/records.go` (parseRecordFieldDeclarations updated)
- `pkg/ast/records.go` (InitValue field already existed in FieldDecl)

### 9.12.2 Parse Record Constants and Class Variables ✅ DONE

**Goal**: Support `const` and `class var` in record bodies.

**Estimate**: 2 hours

**Status**: COMPLETED

**Implementation**:
1. ✅ Parse `const <name> = <value>;` in record body
2. ✅ Parse `class var <name>: <type>;` in record body
3. ✅ Parse `class const <name> = <value>;` in record body
4. ✅ Store in RecordDecl AST node
5. ✅ Reuse class const/class var parsing logic

**Files Modified**:
- `internal/parser/records.go` (added const/class var/class const handling)
- `pkg/ast/records.go` (added Constants, ClassVars fields)
- `pkg/printer/dwscript.go` (updated printRecordDecl to output constants and class vars)

### 9.12.3 Semantic Analysis for Record Features ✅ DONE

**Goal**: Type check record field initializers, constants, and class variables.

**Estimate**: 4-5 hours

**Status**: COMPLETED

**Implementation**:
1. ✅ Analyze field initializers, check type compatibility
2. ✅ Support type inference for fields with initializers but no type
3. ✅ Validate record constants (compile-time constant values)
4. ✅ Add record class variables to type system
5. ✅ Handle constant and class var access in field access analysis
6. ✅ Type check initializer expressions

**Files Modified**:
- `internal/semantic/analyze_records.go` (added constant/class var analysis, type inference)
- `internal/types/compound_types.go` (RecordType with Constants, ClassVars fields)
- `internal/types/compound_types.go` (added ConstantInfo struct, updated NewRecordType)

### 9.12.4 Runtime Record Field Initialization ✅ DONE

**Goal**: Execute field initializers when creating record values.

**Estimate**: 3-4 hours

**Status**: COMPLETED

**Implementation**:
1. ✅ On record variable declaration, initialize fields with default values
2. ✅ Evaluate field initializer expressions
3. ✅ Handle nested record initialization
4. ✅ Record constants stored as global values
5. ✅ Record class variables in global storage
6. ✅ Fixed semantic analysis for record method bodies - proper scope management
7. ✅ Bind Self, fields, constants, and class variables in method scope

**Files Modified**:
- `internal/interp/record.go` (record initialization with constants/class vars)
- `internal/interp/functions_records.go` (bind constants/class vars in methods)
- `internal/interp/objects_hierarchy.go` (constant/class var access)
- `internal/semantic/analyze_records.go` (fixed scope management for inline methods)
- `internal/semantic/analyze_classes_decl.go` (analyzeRecordMethodBody implementation)
- `internal/semantic/analyze_literals.go` (skip fields with initializers)
- `internal/types/compound_types.go` (added FieldsWithInit map)

### 9.12.5 Enhanced Record Methods ✅ DONE

**Goal**: Fix record method semantics (self reference, result handling).

**Estimate**: 2-3 hours

**Status**: COMPLETED

**Implementation**:
1. ✅ Record methods receive copy of record as Self
2. ✅ Modifications to Self don't affect original (value semantics)
3. ✅ Result variable in record functions
4. ✅ Nested record field access in methods
5. ✅ Separate method implementations (non-inline) now work correctly
6. ✅ Fields accessible in method bodies without Self prefix

**Files Modified**:
- `internal/semantic/analyze_classes_decl.go` (analyzeRecordMethodBody with field binding)

**Note**: Most of the enhanced method semantics were already implemented. This task completed the missing piece - proper semantic analysis and scope binding for method bodies.

### 9.12.6 Record Properties

**Goal**: Support property declarations in records.

**Estimate**: 2-3 hours

**Implementation**:
1. Parse property declarations in record bodies
2. Semantic analysis for record properties
3. Runtime property getter/setter execution
4. Properties accessing record fields

**Files to Modify**:
- `internal/parser/records.go` (record property parsing)
- `internal/semantic/analyze_records.go` (property analysis)
- `internal/interp/records.go` (property access)

**Success Criteria**:
- Record fields can have initialization expressions
- Record constants and class variables work
- Nested records parse and execute correctly
- Record methods have proper value semantics
- Record properties work correctly
- All 32 record advanced feature tests pass

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/record
go test -v ./internal/semantic -run TestRecordAdvanced
```

---

## Task 9.13: Implement Property Advanced Features

**Goal**: Add indexed properties, array-typed properties, and property promotion/reintroduce.

**Estimate**: 8-12 hours (1-1.5 days)

**Status**: NOT STARTED

**Impact**: Unlocks 9 failing tests in SimpleScripts

**Priority**: P1 - IMPORTANT (OOP encapsulation)

**Description**: Properties currently have basic getter/setter support, but DWScript includes advanced features like indexed properties (e.g., `Items[i]`), array-typed properties, property promotion from parent classes, and the `reintroduce` keyword for shadowing parent properties.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P1, Section 9

**Failing Tests** (9 total):
- class_var_as_prop
- index_property
- property_call
- property_index
- property_of_as
- property_promotion
- property_reintroduce
- property_sub_default
- property_type_array

**Example**:
```pascal
type
  TList = class
    private
      FData: array of Integer;
    public
      // Indexed property (default)
      property Items[Index: Integer]: Integer
        read GetItem write SetItem; default;

      // Array-typed property
      property Data: array of Integer read FData;

    function GetItem(Index: Integer): Integer;
    begin
      Result := FData[Index];
    end;

    procedure SetItem(Index: Integer; Value: Integer);
    begin
      FData[Index] := Value;
    end;
  end;

var list: TList;
begin
  list := TList.Create;
  list[0] := 42;        // Uses default indexed property
  PrintLn(list[0]);
end;
```

**Complexity**: Medium-High - Multiple property enhancement features

**Subtasks**:

### 9.13.1 Parse Indexed Properties

**Goal**: Support `property Name[Index: Type]: Type` syntax.

**Estimate**: 2-3 hours

**Implementation**:
1. Extend property parsing to handle `[` parameters `]` after property name
2. Parse multiple index parameters
3. Parse `default` keyword for default indexed property
4. Store index parameters in PropertyDecl

**Files to Modify**:
- `internal/parser/properties.go` (indexed property parsing)
- `pkg/ast/properties.go` (IndexParams field on PropertyDecl)

### 9.13.2 Parse Array-Typed Properties

**Goal**: Support properties with array types.

**Estimate**: 1 hour

**Implementation**:
1. Allow array types in property type declarations
2. Handle getter/setter with array return/parameter types
3. Parse array property access syntax

**Files to Modify**:
- `internal/parser/properties.go` (array type properties)

### 9.13.3 Semantic Analysis for Indexed Properties

**Goal**: Type check indexed property access and assignments.

**Estimate**: 3-4 hours

**Implementation**:
1. Resolve indexed property access: `obj.Prop[index]`
2. Check index parameter types match declaration
3. Type check getter/setter signatures with index parameters
4. Default indexed property allows `obj[index]` syntax
5. Array-typed properties type check correctly

**Files to Modify**:
- `internal/semantic/analyze_properties.go` (indexed property analysis)
- `internal/semantic/analyze_expressions.go` (property access with indices)

### 9.13.4 Runtime Indexed Property Access

**Goal**: Execute indexed property getters/setters with indices.

**Estimate**: 2-3 hours

**Implementation**:
1. Evaluate index expressions
2. Call getter with index parameters
3. Call setter with index + value parameters
4. Default indexed property via `[]` operator
5. Array-typed property returns array value

**Files to Modify**:
- `internal/interp/properties.go` (indexed property execution)
- `internal/interp/objects_properties.go` (property access)

### 9.13.5 Property Promotion and Reintroduce

**Goal**: Support `reintroduce` and property promotion from parent.

**Estimate**: 2-3 hours

**Implementation**:
1. Parse `reintroduce` keyword on properties
2. Semantic analysis: allow child class to shadow parent property with reintroduce
3. Property promotion: access parent property via child class
4. Runtime: respect override/reintroduce semantics

**Files to Modify**:
- `internal/parser/properties.go` (reintroduce keyword)
- `internal/semantic/analyze_properties.go` (promotion/reintroduce)
- `internal/interp/properties.go` (runtime property lookup)

**Success Criteria**:
- Indexed properties parse and work correctly
- Array-typed properties supported
- Default indexed property enables `obj[i]` syntax
- Property promotion from parent classes works
- `reintroduce` keyword allows property shadowing
- All 9 property advanced feature tests pass

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/property
go test -v ./internal/semantic -run TestPropertyAdvanced
```

---

## Task 9.14: Fix Inheritance and Virtual Methods Issues

**Goal**: Correct override validation, inherited keyword, reintroduce, virtual constructors, and VMT architecture.

**Estimate**: 18-26 hours (2.5-3.5 days)

**Status**: NOT STARTED

**Impact**: Unlocks 14 failing tests in SimpleScripts + fixes `reintroduce virtual` semantics

**Priority**: P1 - CRITICAL (OOP polymorphism + architecture fix)

**Description**: Current inheritance and virtual method implementation has several issues: improper override validation, incomplete `inherited` keyword support (especially in constructors), missing `reintroduce` keyword, and incorrect virtual constructor behavior. These are critical for proper OOP polymorphism.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P1, Section 10

**Failing Tests** (14 total):
- class_forward
- class_parent
- clear_ref_in_constructor_assignment
- clear_ref_in_destructor
- destroy
- free_destroy
- inherited1, inherited2
- inherited_constructor
- oop
- override_deep
- reintroduce
- reintroduce_virtual
- virtual_constructor, virtual_constructor2

**Example**:
```pascal
type
  TBase = class
    constructor Create; virtual;
    procedure DoSomething; virtual;
  end;

  TDerived = class(TBase)
    constructor Create; override;  // Override virtual constructor
    procedure DoSomething; override;
  end;

constructor TBase.Create;
begin
  PrintLn('TBase.Create');
end;

constructor TDerived.Create;
begin
  inherited;  // Call parent constructor
  PrintLn('TDerived.Create');
end;

procedure TDerived.DoSomething;
begin
  inherited DoSomething;  // Call parent method
  PrintLn('TDerived.DoSomething');
end;
```

**Complexity**: Medium-High - Requires fixes across semantic and runtime

**Subtasks**:

### 9.14.1 Fix Override Validation

**Goal**: Properly validate override keyword matches parent virtual method.

**Estimate**: 2-3 hours

**Implementation**:
1. Check parent class has method with same name
2. Verify parent method is declared virtual or override
3. Validate signature matches (same params and return type)
4. Report error if override used without virtual parent
5. Deep override chains (override of override)

**Files to Modify**:
- `internal/semantic/analyze_classes.go` (override validation)

### 9.14.2 Implement Inherited Keyword Fully

**Goal**: Support `inherited` calls in all method types.

**Estimate**: 3-4 hours

**Implementation**:
1. Parse `inherited;` (call parent's same method)
2. Parse `inherited MethodName(args);` (call specific parent method)
3. Semantic analysis: resolve inherited to parent class method
4. In constructors, `inherited` calls parent constructor
5. Type check inherited calls

**Files to Modify**:
- `internal/parser/expressions.go` (parse inherited)
- `pkg/ast/expressions.go` (InheritedExpression node)
- `internal/semantic/analyze_classes.go` (inherited resolution)

### 9.14.3 Runtime Inherited Call Execution

**Goal**: Execute inherited calls correctly.

**Estimate**: 2-3 hours

**Implementation**:
1. Look up parent class method
2. Call parent method with current object (Self)
3. In constructors, chain to parent constructor before child initialization
4. In destructors, call parent destructor after child cleanup

**Files to Modify**:
- `internal/interp/objects_methods.go` (inherited calls)
- `internal/interp/objects_creation.go` (constructor chaining)
- `internal/interp/objects_destruction.go` (destructor chaining)

### 9.14.4 Implement Reintroduce Keyword

**Goal**: Support reintroduce for shadowing parent members.

**Estimate**: 2 hours

**Implementation**:
1. Parse `reintroduce` keyword on methods
2. Semantic analysis: allow method to shadow parent method without override
3. Warning if shadowing without reintroduce
4. Runtime: child method hides parent method

**Files to Modify**:
- `internal/parser/functions.go` (parse reintroduce)
- `internal/semantic/analyze_classes.go` (reintroduce validation)

### 9.14.5 Fix Virtual Constructor Behavior

**Goal**: Correct virtual constructor dispatch and initialization.

**Estimate**: 2-3 hours

**Implementation**:
1. Virtual constructors can be overridden
2. Calling Create on class reference dispatches to correct constructor
3. Constructor chaining with inherited works correctly
4. Virtual destructor support (Free, Destroy)

**Files to Modify**:
- `internal/interp/objects_creation.go` (virtual constructor dispatch)
- `internal/semantic/analyze_classes.go` (virtual constructor validation)

### 9.14.6 Refactor VMT to Array-Based Architecture

**Goal**: Replace map-based VMT with array-based VMT to support `reintroduce virtual` semantics.

**Estimate**: 8-12 hours (1-1.5 days)

**Priority**: P1 - CRITICAL for correct OOP semantics

**Context**: During Phase 3.5.8 implementation, discovered that current VMT uses `map[string]*VirtualMethodEntry` which cannot support `reintroduce virtual`. DWScript reference uses array with indices: `FVirtualMethodTable []TMethodSymbol`.

**Problem**:

- Current: `map[signature] → VirtualMethodEntry` - only ONE entry per signature
- Needed: Array where each virtual method gets unique index, `reintroduce virtual` gets NEW index
- See detailed analysis: `docs/phase3.5.8-findings.md`

**Implementation**:

1. Change `ClassInfo.VirtualMethodTable` from `map[string]*VirtualMethodEntry` to `[]*VirtualMethodEntry`
2. Add `VMTIndex int` field to method declarations during semantic analysis
3. Assign VMT indices:
   - First `virtual` → allocate new VMT index
   - `override` → use parent's VMT index for same method
   - `reintroduce virtual` → allocate NEW VMT index (starts new chain)
4. Update method dispatch to use VMT index instead of map lookup
5. Update `buildVirtualMethodTable()` to build array instead of map
6. Fix `reintroduce_virtual.pas` test case

**Files to Modify**:

- `internal/interp/class.go` (ClassInfo.VirtualMethodTable type, buildVirtualMethodTable)
- `internal/interp/objects_methods.go` (method dispatch - use VMT index)
- `internal/semantic/analyze_classes.go` (assign VMT indices)
- `pkg/ast/functions.go` (add VMTIndex field)

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/reintroduce_virtual
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/override_deep
```

**Success Criteria**:
- Override validation checks parent method is virtual
- `inherited` works in methods, constructors, destructors
- `inherited MethodName` syntax works
- `reintroduce` allows shadowing without override
- `reintroduce virtual` creates new VMT entry (not overwrites)
- Virtual constructors dispatch correctly
- Constructor/destructor chaining works
- All 14 inheritance/virtual method tests pass
- `reintroduce_virtual.pas` fixture test passes

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/inherited
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/virtual_constructor
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/override
go test -v ./internal/semantic -run TestInheritance
```

---

## Task 9.15: Implement Enum Advanced Features

**Goal**: Add enum boolean operations, bounds (Low/High), EnumByName, flags, scoped enums, and deprecation.

**Estimate**: 8-12 hours (1-1.5 days)

**Status**: IN PROGRESS (9/17 enum tests passing, boolean ops & implicit conversion complete)

**Impact**: Unlocks 12 failing tests in SimpleScripts

**Priority**: P1 - IMPORTANT (Type system completeness)

**Description**: Enumerations currently have basic support, but DWScript includes advanced features like boolean operations on enums, bounds checking (Low/High), EnumByName function for string-to-enum conversion, enum flags (sets), scoped enums, and enum element deprecation.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P1, Section 11

**Completed** (2 total):
- ✅ enum_bool_op (Tasks 9.15.7 and 9.15.8 - boolean ops and implicit conversion)

**Failing Tests** (11 remaining):

- aliased_enum
- enum_bounds
- enum_byname
- enum_casts
- enum_element_deprecated
- enum_flags1
- enum_scoped
- enum_to_integer
- enumerations
- enumerations_names
- enumerations_qualifiednames

**Example**:
```pascal
type
  TMyEnum = (meA, meB, meC);
  TScopedEnum = (seX, seY) scoped;  // Elements accessed as TScopedEnum.seX
  TFlags = (flRead, flWrite, flExecute) flags;  // Bit flags

var
  e: TMyEnum;
  flags: TFlags;
begin
  // Bounds
  for e := Low(TMyEnum) to High(TMyEnum) do
    PrintLn(e);

  // EnumByName
  e := EnumByName<TMyEnum>('meB');

  // Boolean operations
  if e in [meA, meB] then
    PrintLn('A or B');

  // Flags (set operations)
  flags := [flRead, flWrite];
  if flRead in flags then
    PrintLn('Readable');

  // Scoped enum
  var se := TScopedEnum.seX;  // Must use type prefix
end;
```

**Complexity**: Medium - Multiple enum enhancements

**Subtasks**:

### 9.15.2 Enum Boolean Operations

**Goal**: Support boolean operators with enum operands.

**Estimate**: 2-3 hours

**Implementation**:
1. `in` operator: check if enum value in set of values
2. Set operations on enum values: `[meA, meB]`
3. Semantic analysis for enum set expressions
4. Runtime evaluation of enum in set

**Files to Modify**:
- `internal/semantic/analyze_expressions.go` (enum boolean ops)
- `internal/interp/expressions_operators.go` (enum in operator)

### 9.15.7 Enum Boolean Operators ✅ COMPLETE

**Goal**: Support boolean/bitwise operators on enum values (for flags).

**Estimate**: 2-3 hours

**Status**: ✅ COMPLETE (2025-11-22) - See Task 1.6

**Implementation**:
1. ✅ Allow `or`, `and`, `xor` operators on enum operands (especially flags)
2. ✅ Semantic analysis: check enum types are compatible for operators
3. ✅ Runtime: evaluate bitwise operations on enum ordinal values
4. ✅ Return enum type for result (not integer)

**Tests**: enum_bool_op ✅ PASSING (AST interpreter)

**Example**:
```pascal
type TFlags = flags (flRead, flWrite, flExecute);
var f1 := TFlags.flRead or TFlags.flWrite;  // Result: 3 (as TFlags)
var f2 := TFlags.flWrite and TFlags.flRead; // Result: 0 (as TFlags)
```

**Files Modified**:
- ✅ `internal/semantic/analyze_expr_operators.go` (allow enum operands for or/and/xor)
- ✅ `internal/interp/expressions_binary.go` (evaluate enum boolean ops)

**Note**: Bytecode VM support pending (see Task 14.16)

### 9.15.8 Implicit Enum-to-Integer Conversion ✅ COMPLETE

**Goal**: Allow implicit conversion from enum to Integer in function calls.

**Estimate**: 1-2 hours

**Status**: ✅ COMPLETE (2025-11-22) - See Task 1.6

**Implementation**:
1. ✅ When calling `func(i: Integer)` with enum argument, auto-convert
2. ✅ Semantic analysis: allow enum-to-integer implicit conversion
3. ✅ Runtime: automatically get ordinal value when passing enum to Integer param

**Tests**: enum_bool_op (PrintInt calls) ✅ PASSING (AST interpreter)

**Example**:
```pascal
procedure PrintInt(i: Integer);
type TEnum = flags (Alpha, Beta);
PrintInt(TEnum.Alpha);  // Should pass 1, not fail with type mismatch
```

**Files to Modify**:
- `internal/semantic/analyze_expressions.go` (implicit conversion in canAssign)
- `internal/interp/expressions.go` (auto-convert enum to integer)

### 9.15.9 Constant Expressions in Enum Values

**Goal**: Support constant expressions (like Ord('A')) in enum value assignments.

**Estimate**: 2-3 hours

**Status**: TODO

**Implementation**:
1. Parse constant expressions in enum value positions
2. Evaluate constant expressions at parse/semantic time
3. Support function calls like `Ord('A')`, `Chr(65)` in enum values
4. Constant folding for arithmetic expressions

**Failing Tests**: enum_to_integer

**Example**:
```pascal
type TEnumAlpha = (eAlpha = Ord('A'), eBeta, eGamma);
// eAlpha should have value 65 (ASCII 'A')
```

**Files to Modify**:
- `internal/parser/enums.go` (parse expressions instead of just integers)
- `internal/semantic/analyze_enums.go` (evaluate constant expressions)
- Add constant expression evaluator

### 9.15.10 Enum Instance .Value Property ✓ DONE

**Goal**: Support .Value property on enum instances to get ordinal value.

**Estimate**: 1-2 hours

**Status**: Complete - .Value property fully working

**Implementation**:
1. Add .Value property to enum values (returns Integer)
2. Semantic analysis: recognize .Value on enum types
3. Runtime: return ordinal value when accessing .Value
4. Also support .ToString() method

**Failing Tests**: enum_to_integer

**Example**:
```pascal
var e: TEnum := eOne;
PrintLn(e.Value);           // Prints ordinal value
PrintLn(eOne.Value);        // Direct access also works
PrintLn(TEnum.eTwo.Value);  // Qualified access works
```

**Files to Modify**:
- `internal/semantic/analyze_expressions.go` (add .Value property to enum helper)
- `internal/interp/objects_hierarchy.go` (evaluate .Value member access)

### 9.15.11 Additional Enum Features

**Goal**: Handle remaining edge cases and test failures.

**Estimate**: 2-3 hours

**Status**: TODO

**Implementation**:
1. Aliased enums (type alias to enum type)
2. Enum deprecation warnings
3. Enum names and qualified names functions
4. For-in loops over enum ranges

**Failing Tests**: aliased_enum, enum_element_deprecated, enumerations, enumerations2, enumerations_names, enumerations_qualifiednames

**Files to Modify**:
- Various files depending on specific feature

**Success Criteria**:
- ✓ Scoped enums enforce qualified access (MyEnum.a)
- ✓ Flags enums use power-of-2 values
- ✓ Low/High properties and methods return enum bounds
- ✓ ByName() method converts string to enum
- ✓ .Value property on enum instances returns ordinal value
- ✓ Enum-to-integer and integer-to-enum casts - DONE
- ⚠ Enum boolean operations (or, and, xor) - TODO
- ⚠ Implicit enum-to-integer conversion - TODO
- ⚠ Constant expressions in enum values - TODO
- **Progress**: 8/17 tests passing (enum_scoped, enum_flags1, enum_bounds, enum_byname, enum_value, enum_qualified_access, enum_ord, enum_forin)

**Completed Work**:
- ✅ 9.15.7: Boolean operators on enum types (or, and, xor for flags) - Task 1.6
- ✅ 9.15.8: Implicit conversion from enum to Integer in function calls - Task 1.6

**Remaining Work** (see subtasks 9.15.9-9.15.11 for details):
- 9.15.9: Constant expressions in enum value assignments (e.g., Ord('A'))
- 9.15.11: Additional edge cases (aliased enums, deprecation, names)
- Bytecode VM support (see Task 14.16)

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/enum
go test -v ./internal/semantic -run TestEnumAdvanced
```

---

## Task 9.16 Introduce Base Structs for AST Nodes

**Goal**: Eliminate code duplication by introducing base structs for common node fields and behavior.

**Estimate**: 8-10 hours (1-1.5 days)

**Status**: IN PROGRESS

**Impact**: Reduces AST codebase by ~30% (~500 lines), eliminates duplicate boilerplate across 50+ node types

**Description**: Currently, every AST node type duplicates identical implementations for `Pos()`, `End()`, `TokenLiteral()`, `GetType()`, and `SetType()` methods. This creates ~500 lines of repetitive code that is error-prone to maintain. By introducing base structs with embedding, we can eliminate this duplication while maintaining the same interface.

**Current Problem**:

```go
// Repeated ~50 times across different node types
type IntegerLiteral struct {
    Type   *TypeAnnotation
    Token  token.Token
    Value  int64
    EndPos token.Position
}

func (il *IntegerLiteral) Pos() token.Position  { return il.Token.Pos }
func (il *IntegerLiteral) End() token.Position {
    if il.EndPos.Line != 0 {
        return il.EndPos
    }
    pos := il.Token.Pos
    pos.Column += len(il.Token.Literal)
    pos.Offset += len(il.Token.Literal)
    return pos
}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) GetType() *TypeAnnotation    { return il.Type }
func (il *IntegerLiteral) SetType(typ *TypeAnnotation) { il.Type = typ }
```

**Strategy**: Create base structs using Go embedding to share common fields and method implementations:

1. **BaseNode**: Common fields (Token, EndPos) and methods (Pos, End, TokenLiteral)
2. **TypedExpressionBase**: Extends BaseNode with Type field and GetType/SetType methods
3. Refactor all node types to embed appropriate base struct
4. Remove duplicate method implementations

**Complexity**: Medium - Requires systematic refactoring of all AST node types across 25 files (~5,500 lines)

**Subtasks**:

- [x] 9.16.1 Design base struct hierarchy
  - [x] Create `BaseNode` struct with Token, EndPos fields
  - [x] Create `TypedExpressionBase` struct embedding BaseNode with Type field
  - [x] Implement common methods once on base structs
  - [x] Document design decisions and usage patterns
  - [x] Add `pkg/ast/base.go`

- [x] 9.16.2 Refactor literal expression nodes (Identifier, IntegerLiteral, FloatLiteral, StringLiteral, BooleanLiteral, CharLiteral, NilLiteral)
  - [x] Embed `TypedExpressionBase` into Identifier and adjust parser/tests
  - [x] Embed `TypedExpressionBase` into numeric/string/char/boolean literal structs
  - [x] Embed `TypedExpressionBase` into NilLiteral
  - [x] Remove redundant `TokenLiteral/Pos/End/GetType` methods
  - [x] Update parser/semantic/interpreter tests that construct these literals
  - [x] Updated all parser files (12 files, 37 instances)
  - [x] Updated all test files in internal/ast (17 files, 446 instances)
  - [x] Updated all test files in internal/bytecode (6 files, 100+ instances)
  - [x] Updated all test files in internal/interp (6 files, 85+ instances)
  - [x] Updated all test files in internal/semantic (3 files, 55+ instances)
  - All literal expression nodes now use TypedExpressionBase successfully

- [x] 9.16.3 Refactor binary and unary expressions (BinaryExpression, UnaryExpression, GroupedExpression, RangeExpression)
  - [x] Embed `TypedExpressionBase` into BinaryExpression
  - [x] Embed `TypedExpressionBase` into UnaryExpression
  - [x] Embed `TypedExpressionBase` into GroupedExpression
  - [x] Embed `TypedExpressionBase` into RangeExpression
  - [x] Remove duplicate type/position helpers and verify parser/semantic behavior
  - [x] Updated parser files (expressions.go, arrays.go, control_flow.go, sets.go)
  - [x] Updated 17 test files across internal/ast, internal/bytecode, internal/semantic
  - [x] All tests pass successfully - removed ~120 lines of boilerplate

- [x] 9.16.4 Refactor statement nodes (ExpressionStatement, VarDeclStatement, AssignmentStatement, BlockStatement, IfStatement, WhileStatement, etc.)
  - [x] Identify all statement structs across `pkg/ast/statements.go`, `pkg/ast/control_flow.go`, and related files
  - [x] Embed `BaseNode` into expression statements/assignments/var decls (already done in previous tasks)
  - [x] Embed `BaseNode` into control-flow statements (if/while/for/try/case) (already done in previous tasks)
  - [x] Embed `BaseNode` into exception-related nodes: TryStatement, ExceptClause, ExceptionHandler, FinallyClause, RaiseStatement
  - [x] Embed `BaseNode` into ReturnStatement
  - [x] Remove redundant position/token helpers (TokenLiteral, Pos, End) from all refactored nodes
  - [x] Update parser code to construct nodes with BaseNode wrapper
  - [x] Update all test files in internal/bytecode (6 files, 30+ instances)
  - [x] All tests pass successfully - removed ~50 lines of boilerplate from statement nodes

- [x] 9.16.5 Refactor declaration nodes (ConstDecl, FunctionDecl, ClassDecl, InterfaceDecl, etc.)
  - [x] Embed BaseNode into HelperDecl
  - [x] Embed BaseNode into InterfaceDecl / InterfaceMethodDecl
  - [x] Embed BaseNode into ConstDecl
  - [x] Embed BaseNode into TypeDeclaration
  - [x] Embed BaseNode into FieldDecl
  - [x] Embed BaseNode into PropertyDecl
  - [x] Embed BaseNode into FunctionDecl / constructor nodes
  - [x] Embed BaseNode into ClassDecl / Class-related structs (`pkg/ast/classes.go`)
  - [x] Embed BaseNode into RecordDecl / RecordPropertyDecl / FieldInitializer / RecordLiteralExpression (`pkg/ast/records.go`)
  - [x] Embed BaseNode into OperatorDecl
  - [x] Embed BaseNode into EnumDecl (`pkg/ast/enums.go`)
  - [x] Embed BaseNode into ArrayDecl/SetDecl nodes (`pkg/ast/arrays.go`, `pkg/ast/sets.go`)
  - [x] Embed BaseNode into UnitDeclaration and UsesClause structures (`pkg/ast/unit.go`)
  - [x] Remove duplicate helper methods once all declaration structs embed the base
  - [x] Update all parser files to use BaseNode syntax in struct literals
  - [x] Update all test files to use BaseNode syntax
  - Files: `pkg/ast/declarations.go`, `pkg/ast/functions.go`, `pkg/ast/classes.go`, `pkg/ast/interfaces.go`, `pkg/ast/records.go`, `pkg/ast/enums.go`, `pkg/ast/operators.go`, `pkg/ast/arrays.go`, `pkg/ast/sets.go`, `pkg/ast/unit.go` (~200 lines reduced)
  - All declaration nodes now embed BaseNode, eliminating duplicate boilerplate code

- [x] 9.16.6 Refactor type-specific nodes (ArrayLiteralExpression, CallExpression, NewExpression, MemberAccessExpression, etc.)
  - [x] Refactored NewExpression to embed TypedExpressionBase
  - [x] Refactored MemberAccessExpression to embed TypedExpressionBase
  - [x] Refactored MethodCallExpression to embed TypedExpressionBase
  - [x] Refactored InheritedExpression to embed TypedExpressionBase
  - [x] Updated all parser files (internal/parser/classes.go, internal/parser/expressions.go)
  - [x] Updated all test files (internal/bytecode/vm_test.go, internal/bytecode/compiler_expressions_test.go)
  - [x] Updated interpreter files (internal/interp/objects_methods.go, internal/interp/objects_hierarchy.go, internal/interp/objects_instantiation.go, internal/interp/functions_calls.go)
  - Files: `pkg/ast/arrays.go`, `pkg/ast/classes.go`, `pkg/ast/functions.go` (~80 lines of boilerplate removed)

- [x] 9.16.7 Update parser to use base struct constructors
  - [x] Update parser sites already touched (helpers/interfaces/const/type/property/field)
  - [x] Sweep remaining parser files for struct literals using removed `Token` fields
  - All parser files have been updated to use TypedExpressionBase/BaseNode pattern
  - No helper constructors needed - the pattern is straightforward and consistent

- [x] 9.16.8 Update semantic analyzer and interpreter
  - [x] Updated const/type/property/helper-specific tests where embedding occurred
  - [x] Refactored SetLiteral to use TypedExpressionBase (removed ~40 lines of boilerplate)
  - [x] Refactored AddressOfExpression to use TypedExpressionBase (removed ~10 lines of boilerplate)
  - [x] Refactored LambdaExpression to use TypedExpressionBase (removed ~30 lines of boilerplate)
  - [x] Updated all parser/semantic/interpreter/bytecode files for these changes
  - [x] All tests passing for modified types (SetLiteral, AddressOfExpression, LambdaExpression)

- [ ] 9.16.9 Run comprehensive test suite
  - [ ] `go test ./pkg/ast`
  - [ ] `go test ./internal/parser`
  - [ ] `go test ./internal/semantic`
  - [ ] `go test ./internal/interp`
  - [ ] `go test ./internal/bytecode`
  - [ ] Fixture / CLI integration suite

**Files Modified**:

- `pkg/ast/base.go` (new file ~100 lines)
- `pkg/ast/ast.go` (~300 lines reduced to ~150)
- `pkg/ast/statements.go` (~316 lines reduced to ~200)
- `pkg/ast/control_flow.go` (~200 lines reduced to ~120)
- `pkg/ast/declarations.go` (~150 lines reduced to ~80)
- `pkg/ast/functions.go` (~245 lines reduced to ~150)
- `pkg/ast/classes.go` (~400 lines reduced to ~250)
- `pkg/ast/interfaces.go` (~100 lines reduced to ~60)
- `pkg/ast/arrays.go` (~200 lines reduced to ~120)
- `pkg/ast/enums.go` (~100 lines reduced to ~60)
- `pkg/ast/records.go` (~150 lines reduced to ~90)
- `pkg/ast/sets.go` (~100 lines reduced to ~60)
- `pkg/ast/properties.go` (~120 lines reduced to ~70)
- `pkg/ast/operators.go` (~80 lines reduced to ~50)
- `pkg/ast/exceptions.go` (~100 lines reduced to ~60)
- `pkg/ast/lambda.go` (~80 lines reduced to ~50)
- `pkg/ast/helper.go` (~168 lines reduced to ~100)
- Plus updates to parser, semantic analyzer, and interpreter

**Acceptance Criteria**:
- All AST nodes embed either BaseNode or TypedExpressionBase
- No duplicate Pos/End/TokenLiteral/GetType/SetType implementations
- All existing tests pass (100% backward compatibility)
- Codebase reduced by ~500 lines
- AST package is more maintainable with centralized common behavior
- Documentation explains base struct usage and when to embed each type

**Benefits**:
- 30% reduction in AST code (~500 lines eliminated)
- Single source of truth for common behavior
- Easier to add new node types (less boilerplate)
- Reduced chance of copy-paste errors
- Consistent behavior across all nodes

---

- [ ] 9.18 Separate Type Metadata from AST

**Goal**: Move type information from AST nodes to a separate metadata table, making the AST immutable and reusable.

**Estimate**: 6-8 hours (1 day)

**Status**: IN PROGRESS

**Impact**: Cleaner separation of parsing vs semantic analysis, reduced memory usage, enables multiple concurrent analyses

**Description**: Currently, every expression node carries a `Type *TypeAnnotation` field that is nil during parsing and populated during semantic analysis. This couples the AST to the semantic analyzer and wastes memory (~16 bytes per node). Moving type information to a separate side table improves separation of concerns and enables the AST to be analyzed multiple times with different contexts.

**Current Problem**:

```go
type IntegerLiteral struct {
    Type   *TypeAnnotation  // nil until semantic analysis
    Token  token.Token
    Value  int64
    EndPos token.Position
}
```

**Strategy**: Create a separate metadata table that maps AST nodes to their semantic information:

1. Remove Type field from AST nodes
2. Create SemanticInfo struct with type/symbol maps
3. Semantic analyzer populates SemanticInfo instead of modifying AST
4. Provide accessor methods for type information

**Complexity**: Medium - Requires refactoring semantic analyzer and all code that accesses type information

**Subtasks**:

- [x] 9.18.1 Design metadata architecture
  - Create SemanticInfo struct with node → type mapping
  - Design API for setting/getting type information
  - Consider thread safety for concurrent analyses
  - Document architecture decisions
  - File: `pkg/ast/metadata.go` (new file ~100 lines)

- [x] 9.18.2 Implement SemanticInfo type
  - Map[Expression]*TypeAnnotation for expression types
  - Map[*Identifier]Symbol for symbol resolution
  - Thread-safe accessors with sync.RWMutex
  - File: `pkg/ast/metadata.go`

- [x] 9.18.3 Remove Type field from AST expression nodes
  - Remove Type field from all expression node structs
  - Remove GetType/SetType methods (will be on SemanticInfo)
  - This affects ~30 node types
  - Files: `pkg/ast/base.go`, `pkg/ast/type_annotation.go`

- [x] 9.18.4 Update semantic analyzer to use SemanticInfo
  - Pass SemanticInfo through analyzer methods
  - Replace node.SetType() with semanticInfo.SetType(node, typ)
  - Replace node.GetType() with semanticInfo.GetType(node)
  - Files: `internal/semantic/*.go` (~11 occurrences)

- [x] 9.18.5 Update interpreter to use SemanticInfo
  - Pass SemanticInfo to interpreter
  - Get type information from SemanticInfo instead of nodes
  - Files: `internal/interp/*.go` (~5 occurrences)

- [x] 9.18.6 Update bytecode compiler to use SemanticInfo
  - Pass SemanticInfo to compiler
  - Get type information from metadata table
  - Files: `internal/bytecode/compiler_core.go`, `compiler_expressions.go`

- [x] 9.18.7 Update public API to return SemanticInfo
  - Engine.Analyze() returns SemanticInfo
  - Add accessor methods to Result type
  - Maintain backward compatibility where possible
  - Files: `pkg/dwscript/*.go`

- [ ] 9.18.8 Update LSP integration
  - Pass SemanticInfo to LSP handlers
  - Use metadata for hover, completion, etc.
  - Files: External go-dws-lsp project (document changes needed)

- [x] 9.18.9 Run comprehensive test suite
  - All semantic analyzer tests pass
  - All interpreter tests pass (pre-existing fixture failures unrelated to changes)
  - All bytecode VM tests pass
  - Type field removal complete - saves ~16 bytes per expression node

**Files Modified**:

- `pkg/ast/metadata.go` (new file ~150 lines)
- `pkg/ast/ast.go` (remove Type field from ~15 expression types)
- `pkg/ast/statements.go` (remove Type from CallExpression, etc.)
- `pkg/ast/control_flow.go` (remove Type from IfExpression)
- `pkg/ast/type_annotation.go` (remove TypedExpression interface or make it use SemanticInfo)
- `internal/semantic/analyzer.go` (add SemanticInfo field)
- `internal/semantic/*.go` (replace node.GetType/SetType ~50 times)
- `internal/interp/*.go` (use SemanticInfo for types ~30 times)
- `internal/bytecode/compiler.go` (use SemanticInfo)
- `pkg/dwscript/dwscript.go` (return SemanticInfo from API)

**Acceptance Criteria**:
- No Type field on any AST node
- SemanticInfo table stores all type metadata
- AST is immutable after parsing
- All tests pass (100% backward compatibility in behavior)
- Memory usage reduced (benchmark shows improvement)
- Multiple semantic analyses possible on same AST
- Documentation explains new architecture

**Benefits**:
- Clear separation of parsing vs semantic analysis
- AST is immutable and cacheable
- Reduced memory usage (~16 bytes per expression node)
- Multiple analyses possible (different contexts, parallel)
- Easier to implement alternative analyzers (strict mode, etc.)

---

- [x] 9.19 Extract Pretty-Printing from AST Nodes ✅ COMPLETED (2025-11-15) - Created `pkg/printer` package with multiple output formats (DWScript, Tree, JSON) and styles (Compact, Detailed, Multiline), simplified AST String() methods (~500 lines reduced), updated CLI with --format flag (PR #114)

---

- [ ] 9.20 Standardize Helper Types as Nodes

**Goal**: Make Parameter, CaseBranch, ExceptionHandler, and other helper types implement the Node interface to fix visitor pattern fragility.

**Estimate**: 3-4 hours (0.5 day)

**Status**: IN PROGRESS

**Impact**: Improved type safety, cleaner visitor pattern, more consistent AST structure

**Description**: Several types like `Parameter`, `CaseBranch`, `ExceptionHandler`, and `FieldInitializer` are not Nodes, which breaks the visitor pattern. They require manual handling in walk functions, making the code fragile. Making them implement Node provides type safety and consistent traversal.

**Current Problem**:

```go
// Parameter is not a Node - requires manual walking
type Parameter struct {
    Name         *Identifier
    Type         *TypeAnnotation
    DefaultValue Expression
    // Missing: Token, Pos(), End(), etc.
}

// In visitor.go - manual walking required
func walkFunctionDecl(n *FunctionDecl, v Visitor) {
    for _, param := range n.Parameters {
        // Can't call Walk() - Parameter is not a Node!
        if param.Name != nil {
            Walk(v, param.Name)
        }
        // Manual field walking...
    }
}
```

**Strategy**:

1. Identify all non-Node helper types
2. Add Node interface methods (Pos, End, TokenLiteral)
3. Add node marker methods (statementNode/expressionNode as appropriate)
4. Update visitor to treat them as first-class nodes

**Complexity**: Low - Straightforward interface implementation

**Subtasks**:

- [ ] 9.20.1 Audit AST for non-Node types used in traversal
  - Parameter (in FunctionDecl)
  - CaseBranch (in CaseStatement)
  - ExceptionHandler (in TryStatement)
  - ExceptClause (in TryStatement)
  - FinallyClause (in TryStatement)
  - FieldInitializer (in RecordLiteralExpression)
  - InterfaceMethodDecl (in InterfaceDecl)
  - Create list with current usage
  - File: Create `docs/ast-helper-types.md` with audit results

- [ ] 9.20.2 Make Parameter implement Node
  - Add Token token.Token field
  - Add EndPos token.Position field
  - Implement Pos(), End(), TokenLiteral()
  - Add statementNode() marker (parameters are like declarations)
  - File: `pkg/ast/functions.go`

- [ ] 9.20.3 Make CaseBranch implement Node
  - Add Token token.Token field (first value token)
  - Add EndPos token.Position field
  - Implement Node interface methods
  - Add statementNode() marker
  - File: `pkg/ast/control_flow.go`

- [ ] 9.20.4 Make ExceptionHandler, ExceptClause, FinallyClause implement Node
  - Add required fields to each type
  - Implement Node interface
  - Add statementNode() marker
  - File: `pkg/ast/exceptions.go`

- [ ] 9.20.5 Make FieldInitializer implement Node
  - Add Token, EndPos fields
  - Implement Node interface
  - Add statementNode() marker (like a mini assignment)
  - File: `pkg/ast/records.go`

- [ ] 9.20.6 Make InterfaceMethodDecl implement Node
  - Add Token, EndPos fields
  - Implement Node interface
  - Add statementNode() marker
  - File: `pkg/ast/interfaces.go`

- [ ] 9.20.7 Update visitor to walk helper types as Nodes
  - Remove manual field walking
  - Add cases for new Node types in Walk()
  - Simplify walkXXX functions
  - File: `pkg/ast/visitor.go` (or visitor_reflect.go if 9.17 done first)

- [ ] 9.20.8 Update parser to populate Token/EndPos for helper types
  - Ensure parser sets position info when creating helpers
  - Files: `internal/parser/*.go`

- [ ] 9.20.9 Test visitor traversal includes helper types
  - Create visitor that counts all nodes
  - Verify helpers are visited
  - File: `pkg/ast/visitor_test.go`

**Files Modified**:

- `pkg/ast/functions.go` (Parameter now implements Node)
- `pkg/ast/control_flow.go` (CaseBranch now implements Node)
- `pkg/ast/exceptions.go` (ExceptionHandler, ExceptClause, FinallyClause now implement Node)
- `pkg/ast/records.go` (FieldInitializer now implements Node)
- `pkg/ast/interfaces.go` (InterfaceMethodDecl now implements Node)
- `pkg/ast/visitor.go` (cleaner walk functions, add cases for new nodes)
- `internal/parser/*.go` (set Token/EndPos when creating helper types)
- `docs/ast-helper-types.md` (new documentation)

**Acceptance Criteria**:
- All traversable types implement Node interface
- No manual field walking in visitor.go
- Helper types can be visited like any other node
- All tests pass (especially visitor tests)
- Position information available for all helper types
- Documentation lists which types are Nodes

**Benefits**:
- Type safety (can't forget to walk a child)
- Cleaner visitor implementation
- Consistent AST structure
- Position info available for all traversable types
- Better error messages (can point to exact location)

---

- [ ] 9.21 Add Builder Pattern for Complex Nodes

**Goal**: Create builder types for complex AST nodes to prevent invalid construction and improve code clarity.

**Estimate**: 6-8 hours (1 day)

**Status**: IN PROGRESS

**Impact**: Prevents invalid AST construction, improves parser readability, catches errors at construction time

**Description**: Complex nodes like FunctionDecl and ClassDecl have many fields with interdependencies (e.g., can't be both virtual and abstract, must have body if not abstract, etc.). Currently, nothing prevents invalid combinations. Builders provide validation at construction time and make parser code more readable.

**Current Problem**:

```go
// Parser can create invalid combinations
fn := &FunctionDecl{
    Name: name,
    IsVirtual: true,
    IsAbstract: true,  // INVALID: can't be both!
    Body: nil,         // Missing body check
}
```

**Strategy**:

1. Create builder types for complex nodes
2. Builders enforce invariants and provide fluent API
3. Parser uses builders instead of direct struct construction
4. Builders validate on Build() call

**Complexity**: Medium - Need to identify all invariants and implement builders

**Subtasks**:

- [ ] 9.21.1 Identify nodes that need builders
  - FunctionDecl (most complex: ~15 boolean flags)
  - ClassDecl (inheritance, interfaces, abstract)
  - InterfaceDecl (inheritance)
  - PropertyDecl (read/write specs, indexed)
  - OperatorDecl (operator type, operands)
  - Create design doc with invariants for each
  - File: `docs/ast-builders.md` (new)

- [ ] 9.21.2 Create FunctionDeclBuilder
  - Fluent API: NewFunction(name).WithParam(p).Virtual().Build()
  - Validate: virtual XOR abstract, body required unless abstract/forward/external
  - Validate: constructor can't have return type
  - Validate: destructor must be named specific way
  - File: `pkg/ast/builders/function.go` (new package ~150 lines)

- [ ] 9.21.3 Create ClassDeclBuilder
  - Fluent API: NewClass(name).Extends(parent).Implements(iface).Abstract().Build()
  - Validate: parent is class, interfaces are interfaces
  - Validate: abstract flag consistent with abstract methods
  - Validate: partial + abstract combinations
  - File: `pkg/ast/builders/class.go` (new file ~120 lines)

- [ ] 9.21.4 Create InterfaceDeclBuilder
  - Fluent API: NewInterface(name).Extends(parent).WithMethod(m).Build()
  - Validate: parent is interface
  - Validate: methods are interface methods (no body)
  - File: `pkg/ast/builders/interface.go` (new file ~80 lines)

- [ ] 9.21.5 Create PropertyDeclBuilder
  - Fluent API: NewProperty(name, typ).Read(spec).Write(spec).Indexed(params).Build()
  - Validate: at least one of read/write specified
  - Validate: indexed params consistent
  - File: `pkg/ast/builders/property.go` (new file ~100 lines)

- [ ] 9.21.6 Create OperatorDeclBuilder
  - Fluent API: NewOperator(op).Unary(typ).Binary(lhs, rhs).Returns(ret).Build()
  - Validate: unary XOR binary
  - Validate: valid operator type
  - File: `pkg/ast/builders/operator.go` (new file ~80 lines)

- [ ] 9.21.7 Update parser to use builders
  - Replace direct struct construction with builders
  - Use fluent API for readability
  - Catch construction errors early
  - Files: `internal/parser/parser_functions.go`, `internal/parser/parser_class.go`, etc.

- [ ] 9.21.8 Add builder tests
  - Test valid construction succeeds
  - Test invalid construction fails with clear errors
  - Test all invariants enforced
  - File: `pkg/ast/builders/*_test.go` (new files ~300 lines total)

- [ ] 9.21.9 Add builder documentation
  - Examples of using each builder
  - List of all invariants enforced
  - Migration guide for parser
  - File: `pkg/ast/builders/doc.go` (new file)

**Files Modified**:

- `pkg/ast/builders/function.go` (new file ~150 lines)
- `pkg/ast/builders/class.go` (new file ~120 lines)
- `pkg/ast/builders/interface.go` (new file ~80 lines)
- `pkg/ast/builders/property.go` (new file ~100 lines)
- `pkg/ast/builders/operator.go` (new file ~80 lines)
- `pkg/ast/builders/doc.go` (new file ~50 lines)
- `pkg/ast/builders/*_test.go` (new files ~300 lines total)
- `internal/parser/parser_functions.go` (use FunctionDeclBuilder)
- `internal/parser/parser_class.go` (use ClassDeclBuilder)
- `internal/parser/parser_interfaces.go` (use InterfaceDeclBuilder)
- `internal/parser/parser_properties.go` (use PropertyDeclBuilder)
- `internal/parser/parser_operators.go` (use OperatorDeclBuilder)
- `docs/ast-builders.md` (new documentation ~50 lines)

**Acceptance Criteria**:
- Builders exist for FunctionDecl, ClassDecl, InterfaceDecl, PropertyDecl, OperatorDecl
- All invariants enforced (documented in ast-builders.md)
- Parser uses builders, catching errors at construction time
- Build() method validates and returns error for invalid combinations
- All tests pass, including new builder tests
- Parser code more readable with fluent API
- Documentation explains builder usage and invariants

**Benefits**:
- Catches invalid AST construction at parse time
- Self-documenting code (builder API shows what's valid)
- More readable parser (fluent API vs struct literals)
- Centralized validation logic
- Easier to add new invariants (add to builder, not scattered in parser)

---

- [ ] 9.22 Document Type System Architecture

**Goal**: Create comprehensive documentation explaining TypeAnnotation vs TypeExpression relationship and when to use each.

**Estimate**: 2-3 hours (0.5 day)

**Status**: IN PROGRESS

**Impact**: Improved developer understanding, easier onboarding, fewer type system bugs

**Description**: The relationship between `TypeAnnotation` and `TypeExpression` is unclear from the code alone. TypeAnnotation has both a `Name` field and an `InlineType TypeExpression` field, but it's not obvious when each is used. This confuses developers working on the type system. Clear documentation with examples and diagrams will improve understanding.

**Current Problem**:

```go
// What's the difference? When do I use Name vs InlineType?
type TypeAnnotation struct {
    InlineType TypeExpression  // ???
    Name       string          // ???
    Token      token.Token
    EndPos     token.Position
}
```

**Strategy**:

1. Create architecture documentation with clear explanations
2. Add examples of each use case
3. Create diagrams showing type system structure
4. Add code comments to type system code

**Complexity**: Low - Documentation task, no code changes required

**Subtasks**:

- [ ] 9.22.1 Document TypeAnnotation vs TypeExpression distinction
  - TypeAnnotation: Used when a type is referenced in syntax (`: Integer`)
  - TypeExpression: Defines the structure of a type (interface for type nodes)
  - Name: Simple type reference (`Integer`, `String`, `TMyClass`)
  - InlineType: Complex type definition (`array[0..10] of Integer`, `function(x: Integer): Boolean`)
  - File: `docs/type-system-architecture.md` (new file ~100 lines)

- [ ] 9.22.2 Create type system class diagram
  - Show hierarchy: Node → TypeExpression → specific types
  - Show TypeAnnotation composition
  - Show how semantic analyzer uses these
  - File: `docs/diagrams/type-system.svg` (new diagram)

- [ ] 9.22.3 Add examples for each type usage pattern
  - Example: Simple type reference (`var x: Integer`)
  - Example: Array type (`var arr: array[0..5] of Integer`)
  - Example: Function pointer type (`var fn: function(x: Integer): Boolean`)
  - Example: Anonymous record type
  - File: `docs/type-system-architecture.md` (add examples section)

- [ ] 9.22.4 Document type resolution process
  - How parser creates TypeAnnotations
  - How semantic analyzer resolves names to Type objects
  - How inline types are processed
  - Flow diagram: Source → TypeAnnotation → Type
  - File: `docs/type-system-architecture.md`

- [ ] 9.22.5 Add code comments to type system files
  - pkg/ast/type_annotation.go (explain fields)
  - pkg/ast/type_expression.go (explain interface)
  - internal/types/types.go (explain Type hierarchy)
  - Files: `pkg/ast/type_annotation.go`, `pkg/ast/type_expression.go`, `internal/types/types.go`

- [ ] 9.22.6 Create developer guide
  - "Adding a new type" guide
  - "Understanding type checking" guide
  - Common pitfalls and solutions
  - File: `docs/developer-guides/type-system.md` (new file ~50 lines)

- [ ] 9.22.7 Add package-level documentation
  - Update pkg/ast/doc.go with type system overview
  - Update internal/types/doc.go with Type hierarchy
  - Cross-reference with architecture docs
  - Files: `pkg/ast/doc.go`, `internal/types/doc.go`

**Files Modified**:

- `docs/type-system-architecture.md` (new file ~200 lines)
- `docs/diagrams/type-system.svg` (new diagram)
- `docs/developer-guides/type-system.md` (new file ~50 lines)
- `pkg/ast/type_annotation.go` (add detailed comments ~20 lines)
- `pkg/ast/type_expression.go` (add comments ~10 lines)
- `pkg/ast/doc.go` (add type system section ~20 lines)
- `internal/types/types.go` (add comments ~30 lines)
- `internal/types/doc.go` (create or update ~40 lines)

**Acceptance Criteria**:
- Clear documentation of TypeAnnotation vs TypeExpression
- Diagrams showing type system architecture
- Examples for each usage pattern
- Developer guide for working with types
- Code comments explain key concepts
- Documentation cross-referenced from code
- All type system files have package docs

**Benefits**:

- Faster developer onboarding
- Fewer type system bugs
- Clearer mental model of type system
- Easier to extend type system
- Self-documenting code

---

## Phase 13: go-dws API Enhancements for LSP Integration ✅ COMPLETE

**Goal**: Enhanced go-dws library with structured errors, AST access, position metadata, symbol tables, and type information for LSP features.

**Status**: All 27 tasks complete. Added public `pkg/ast/` and `pkg/token/` packages, structured error types with position info, Parse() mode for fast syntax-only parsing, visitor pattern for AST traversal, symbol table access, and type queries. 100% backwards compatible. Ready for go-dws-lsp integration.

---

## Phase 14: Bytecode Compiler & VM Optimizations ✅ MOSTLY COMPLETE

**Status**: Core implementation complete | **Performance**: 5-6x faster than AST interpreter | **Tasks**: 15 complete, 2 pending

### Overview

This phase implements a bytecode virtual machine for DWScript, providing significant performance improvements over the tree-walking AST interpreter. The bytecode VM uses a stack-based architecture with 116 opcodes and includes an optimization pipeline.

**Architecture**: AST → Compiler → Bytecode → VM → Output

### Phase 14.1: Bytecode VM Foundation ✅ COMPLETE

- [x] 14.1 Research and design bytecode instruction set
  - Stack-based VM with 116 opcodes, 32-bit instruction format
  - Documentation: [bytecode-vm-design.md](docs/architecture/bytecode-vm-design.md)
  - Expected Impact: 2-3x speedup over tree-walking interpreter

- [x] 14.2 Implement bytecode data structures
  - Created `internal/bytecode/bytecode.go` with `Chunk` type (bytecode + constants pool)
  - Implemented constant pool for literals with deduplication
  - Added line number mapping with run-length encoding
  - Implemented bytecode disassembler for debugging (79.7% coverage)

- [x] 14.3 Build AST-to-bytecode compiler
  - Created `internal/bytecode/compiler.go` with visitor pattern
  - Compile expressions: literals, binary ops, unary ops, variables, function calls
  - Compile statements: assignment, if/else, loops, return
  - Handle scoping and variable resolution
  - Optimize constant folding during compilation

- [x] 14.4 Implement bytecode VM core
  - Created `internal/bytecode/vm.go` with instruction dispatch loop
  - Implemented operand stack and call stack
  - Added environment/closure handling with upvalue capture
  - Error handling with structured RuntimeError and stack traces
  - Performance: VM is ~5.6x faster than AST interpreter

- [x] 14.5 Implement arithmetic and logic instructions
  - ADD, SUB, MUL, DIV, MOD instructions
  - NEGATE, NOT instructions
  - EQ, NE, LT, LE, GT, GE comparisons
  - AND, OR, XOR bitwise operations
  - Type coercion (int ↔ float)

- [x] 14.6 Implement variable and memory instructions
  - LOAD_CONST / LOAD_LOCAL / STORE_LOCAL
  - LOAD_GLOBAL / STORE_GLOBAL
  - LOAD_UPVALUE / STORE_UPVALUE with closure capture
  - GET_PROPERTY / SET_PROPERTY for member access

- [x] 14.7 Implement control flow instructions
  - JUMP, JUMP_IF_FALSE, JUMP_IF_TRUE
  - LOOP (jump backward for while/for loops)
  - Patch jump addresses during compilation
  - Break/continue leverage jump instructions

- [x] 14.8 Implement function call instructions
  - CALL instruction for named functions
  - RETURN instruction with trailing return guarantee
  - Handle recursion and call stack depth
  - Implement closures and upvalues
  - Support method calls and `Self` context (OpCallMethod, OpGetSelf)

- [x] 14.9 Implement array and object instructions
  - GET_INDEX, SET_INDEX for array access
  - NEW_ARRAY, ARRAY_LENGTH
  - NEW_OBJECT for class instantiation
  - INVOKE_METHOD for method dispatch

- [x] 14.10 Add exception handling instructions
  - TRY, CATCH, FINALLY, THROW instructions
  - Exception stack unwinding
  - Preserve stack traces across bytecode execution

- [x] 14.11 Optimize bytecode generation
  - Established optimization pipeline with pass manager and toggles
  - Peephole transforms: fold literal push/pop pairs, collapse stack shuffles
  - Dead code elimination: trim after terminators, reflow jump targets
  - Constant propagation: track literal locals/globals, fold arithmetic chains
  - Inline small functions (< 10 instructions)

- [x] 14.12 Integrate bytecode VM into interpreter
  - Added `--bytecode` flag to CLI
  - Added `CompileMode` option (AST vs Bytecode) to `pkg/dwscript/options.go`
  - Bytecode compilation/execution paths in `pkg/dwscript/dwscript.go`
  - Unit loading/parsing parity, tracing, diagnostic output
  - Wire bytecode VM to externals (FFI, built-ins, stdout capture)

- [x] 14.13 Create bytecode test suite
  - Port existing interpreter tests to bytecode
  - Test bytecode disassembler output
  - Verify identical behavior to AST interpreter
  - Performance benchmarks confirm 5-6x speedup

- [x] 14.14 Add bytecode serialization
  - [x] 14.14.1 Define bytecode file format (.dwc)
    - **Task**: Design the binary format for bytecode files
    - **Implementation**:
      - Define magic number (e.g., "DWC\x00") for file identification
      - Define version format (major.minor.patch)
      - Define header structure (magic, version, metadata)
      - Document format specification
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 0.5 day

  - [x] 14.14.2 Implement Chunk serialization
    - **Task**: Serialize bytecode chunks to binary format
    - **Implementation**:
      - Serialize instructions array
      - Serialize line number information
      - Serialize constants pool
      - Write helper functions for writing primitives (int, float, string, bool)
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 1 day

  - [x] 14.14.3 Implement Chunk deserialization
    - **Task**: Deserialize bytecode chunks from binary format
    - **Implementation**:
      - Read and validate magic number and version
      - Deserialize instructions array
      - Deserialize line number information
      - Deserialize constants pool
      - Write helper functions for reading primitives
      - Handle invalid/corrupt bytecode files
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 1 day

  - [x] 14.14.4 Add version compatibility checks
    - **Task**: Ensure bytecode version compatibility
    - **Implementation**:
      - Check version during deserialization
      - Return descriptive errors for version mismatches
      - Add tests for different version scenarios
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 0.5 day

  - [x] 14.14.5 Add serialization tests
    - **Task**: Test serialization/deserialization round-trip
    - **Implementation**:
      - Test simple programs serialize correctly
      - Test complex programs with all value types
      - Test error handling (corrupt files, version mismatches)
      - Verify bytecode produces same output after round-trip
    - **Files**: `internal/bytecode/serializer_test.go`
    - **Estimated time**: 1 day

  - [x] 14.14.6 Add `dwscript compile` command
    - **Task**: CLI command to compile source to bytecode
    - **Implementation**:
      - Add compile subcommand to CLI
      - Parse source file and compile to bytecode
      - Write bytecode to .dwc file
      - Add flags for output file, optimization level
    - **Files**: `cmd/dwscript/main.go`, `cmd/dwscript/compile.go`
    - **Estimated time**: 0.5 day

  - [x] 14.14.7 Update `dwscript run` to load .dwc files
    - **Task**: Allow running precompiled bytecode files
    - **Implementation**:
      - Detect .dwc file extension
      - Load bytecode from file instead of compiling
      - Add performance comparison in benchmarks
    - **Files**: `cmd/dwscript/main.go`, `cmd/dwscript/run.go`
    - **Estimated time**: 0.5 day

  - [x] 14.14.8 Document bytecode serialization
    - **Task**: Update documentation for bytecode files
    - **Implementation**:
      - Document .dwc file format in docs/bytecode-vm.md
      - Add CLI examples for compile command
      - Update README.md with serialization info
    - **Files**: `docs/bytecode-vm.md`, `README.md`, `CLAUDE.md`
    - **Estimated time**: 0.5 day

- [x] 14.15 Document bytecode VM
  - Written `docs/bytecode-vm.md` explaining architecture
  - Documented instruction set and opcodes
  - Provided examples of bytecode output
  - Updated CLAUDE.md with bytecode information

### Phase 14.16: Bytecode VM Core Feature Completion 🚧 IN PROGRESS

> **Source**: TODOs.md Task 1.5 - Bytecode VM Missing Core Features
>
> This section addresses critical missing features in the bytecode VM that prevent
> full parity with the AST interpreter. These are HIGH priority items that block
> VM usability for real-world scripts.

#### 14.16.1 For Loop Support 🔴 HIGH PRIORITY

**Status**: Not Started | **Complexity**: High | **Blocking**: VM cannot execute for loops

**Current State**:
- Opcodes `OpForPrep` and `OpForLoop` exist in instruction.go (lines 357-365)
- No `ForStatement` case in `compileStatement()` (compiler_statements.go)
- VM has no handlers for these opcodes in vm_exec.go
- For loop tests are commented out in vm_parity_test.go (lines 71-79)

**ForStatement AST** (pkg/ast/control_flow.go:167):
```go
type ForStatement struct {
    Start     Expression     // Starting value
    EndValue  Expression     // Ending value
    Body      Statement      // Loop body
    Step      Expression     // Optional step value
    Variable  *Identifier    // Loop variable
    Direction ForDirection   // ForTo or ForDownto
    InlineVar bool           // Whether variable is declared inline
}
```

**Implementation Tasks**:

- [ ] **14.16.1.1** Add ForStatement case to compileStatement()
  - **File**: `internal/bytecode/compiler_statements.go`
  - **Task**: Add `case *ast.ForStatement: return c.compileForStatement(node)` to switch

- [ ] **14.16.1.2** Implement compileForStatement() function
  - **File**: `internal/bytecode/compiler_statements.go`
  - **Complexity**: High (~100-150 lines)
  - **Requirements**:
    1. Evaluate start and end values at compile time or runtime
    2. Create loop variable in local scope (beginScope/declareLocal)
    3. Handle ForTo direction (ascending: current <= end)
    4. Handle ForDownto direction (descending: current >= end)
    5. Handle optional Step expression (default step = 1)
    6. Emit proper jump instructions:
       - Initialize loop variable with start value
       - Check loop condition (compare to end value)
       - Execute body
       - Increment/decrement loop variable by step
       - Jump back to condition check
    7. Handle break/continue within for loops (pushLoop/popLoop)
    8. Ensure loop variable scope is limited to loop body

  **Compilation Strategy Option A (Loop Unrolling with VM Support)**:
  ```
  ; for i := 1 to 3 do PrintLn(i);
  OpLoadConst 1         ; push start value
  OpLoadConst 3         ; push end value
  OpLoadConst 1         ; push step value (default 1)
  OpForPrep loopVar exitJump  ; init loop [start,end,step]->[current], check condition
  loop:
    ; body
    OpLoadLocal loopVar
    OpCallBuiltin PrintLn
    OpForLoop loopVar loop ; increment by step and check condition
  exit:
  ```

  **Compilation Strategy Option B (Compile to While-like Loop)**:
  ```
  ; for i := 1 to 3 do PrintLn(i);
  OpLoadConst 1         ; push start value
  OpStoreLocal loopVar  ; i := 1
  loopStart:
    OpLoadLocal loopVar   ; load i
    OpLoadConst 3         ; load end value
    OpLessEqual           ; i <= 3 (for ForTo)
    OpJumpIfFalse exit
    ; body
    OpLoadLocal loopVar
    OpCallBuiltin PrintLn
    ; increment
    OpLoadLocal loopVar
    OpLoadConst 1         ; step value
    OpAddInt
    OpStoreLocal loopVar
    OpLoop loopStart
  exit:
  ```

- [ ] **14.16.1.3** Implement OpForPrep handler in VM (if using Strategy A)
  - **File**: `internal/bytecode/vm_exec.go`
  - **Task**: Add case for OpForPrep opcode
  - **Semantics**: Initialize loop variable, check if loop should execute
  - **Note**: Direction (ForTo vs ForDownto) must be encoded in opcode operand (e.g., A byte)
    or Strategy B should be used which generates explicit comparison opcodes per direction

- [ ] **14.16.1.4** Implement OpForLoop handler in VM (if using Strategy A)
  - **File**: `internal/bytecode/vm_exec.go`
  - **Task**: Add case for OpForLoop opcode
  - **Semantics**: Increment/decrement loop variable, check condition, jump
  - **Note**: Must handle both ForTo (increment, check `<=`) and ForDownto (decrement, check `>=`)

- [ ] **14.16.1.5** Add loopKindFor to compiler loop tracking
  - **File**: `internal/bytecode/compiler_core.go`
  - **Task**: Add `loopKindFor` constant and update break/continue handling

- [ ] **14.16.1.6** Add for loop compilation tests
  - **File**: `internal/bytecode/compiler_statements_test.go`
  - **Tests**:
    - Simple for-to loop
    - Simple for-downto loop
    - For loop with step
    - For loop with break
    - For loop with continue
    - Nested for loops

- [ ] **14.16.1.7** Enable and verify VM parity tests
  - **File**: `internal/bytecode/vm_parity_test.go`
  - **Task**: Uncomment for loop test (lines 71-79)
  - **Verify**: Output matches AST interpreter

**Acceptance Criteria**:
- For loops compile without error
- ForTo and ForDownto directions work correctly
- Step expression works correctly
- Break/continue work within for loops
- VM parity test passes

---

#### 14.16.2 Result Variable Support 🔴 HIGH PRIORITY

**Status**: Not Started | **Complexity**: Medium | **Blocking**: Functions cannot use implicit Result variable

**Current State**:
- AST interpreter handles Result variable in functions_user.go (lines 88-146)
- Bytecode compiler's `compileFunctionDecl()` doesn't allocate Result variable
- Function tests with Result are commented out in vm_parity_test.go (lines 80-91)

**DWScript Result Variable Semantics**:
```pascal
function Add(a, b: Integer): Integer;
begin
  Result := a + b;  // Assign to Result variable
end;

function Multiply(x, y: Integer): Integer;
begin
  Multiply := x * y;  // Can also assign to function name
end;
```

**Implementation Tasks**:

- [ ] **14.16.2.1** Add Result variable allocation in function prologue
  - **File**: `internal/bytecode/compiler_statements.go` in `compileFunctionDecl()`
  - **Task**: If function has ReturnType, allocate Result as first local variable
  - **Code location**: After parameter binding, before body compilation
  ```go
  // After binding parameters:
  if fn.ReturnType != nil {
      resultSlot, err := child.declareLocal(
          &ast.Identifier{Value: "Result"},
          typeFromAnnotation(fn.ReturnType),
      )
      // NOTE: declareAlias does not exist yet; implement aliasing mechanism
      // or resolve function name to Result slot during identifier lookup
      // (see task 14.16.2.3 for implementation options)
  }
  ```

- [ ] **14.16.2.2** Initialize Result with appropriate default value
  - **File**: `internal/bytecode/compiler_statements.go`
  - **Task**: Emit code to initialize Result based on return type
  - **Default values**:
    - Integer: 0
    - Float: 0.0
    - String: ""
    - Boolean: false
    - Object/Interface: nil
    - Array: empty array
    - Record: initialized record

- [ ] **14.16.2.3** Make function name an alias for Result
  - **File**: `internal/bytecode/compiler_core.go`
  - **Task**: Implement aliasing so `FunctionName := value` resolves to Result slot
  - **Implementation options**:
    1. Add alias table to symbol table mapping function name to Result slot
    2. During identifier lookup in function scope, check if identifier matches
       current function name and resolve to Result slot instead
  - **Reference**: AST interpreter uses `ReferenceValue` pointing to "Result" (functions_user.go:145)

- [ ] **14.16.2.4** Update ensureFunctionReturn() to return Result
  - **File**: `internal/bytecode/compiler_statements.go`
  - **Task**: If function reaches end without explicit return, emit code to return Result value
  - **Current**: `ensureFunctionReturn()` emits `OpReturn 0` (no value)
  - **New**: If ReturnType != nil, emit `OpLoadLocal resultSlot; OpReturn 1`

- [ ] **14.16.2.5** Add Result variable tests
  - **File**: `internal/bytecode/compiler_functions_test.go`
  - **Tests**:
    - Function with Result assignment
    - Function with function name assignment
    - Function with both Result and explicit return
    - Nested functions with Result
    - Recursive function using Result

- [ ] **14.16.2.6** Enable and verify VM parity tests
  - **File**: `internal/bytecode/vm_parity_test.go`
  - **Task**: Uncomment Result variable test (lines 80-91)
  - **Verify**: Output matches AST interpreter

**Acceptance Criteria**:
- Functions can assign to Result variable
- Functions can assign to function name as alias for Result
- Functions without explicit return statement return Result value
- VM parity test passes

---

#### 14.16.3 Trim Builtin Implementation 🟡 MEDIUM PRIORITY

> **Note**: Originally marked as HIGH priority in TODOs.md. Re-evaluated to MEDIUM because
> Trim is less critical than for loops and Result variables for VM parity.

**Status**: Not Started | **Complexity**: Low | **Blocking**: Trim() calls fail in VM mode

**Current State**:
- Trim, TrimLeft, TrimRight exist in AST interpreter builtins (strings_basic.go:188-238)
- VM has TODO comment at vm_calls.go:196: `// TODO: Implement Trim builtin in VM`
- String helper method `trim` is commented out in vm_calls.go

**Implementation Tasks**:

- [ ] **14.16.3.1** Add builtinTrim function to VM
  - **File**: `internal/bytecode/vm_builtins_string.go`
  - **Task**: Implement Trim(str) - removes leading and trailing whitespace
  ```go
  func builtinTrim(vm *VM, args []Value) (Value, error) {
      if len(args) != 1 {
          return NilValue(), fmt.Errorf("Trim expects 1 argument, got %d", len(args))
      }
      if !args[0].IsString() {
          return NilValue(), fmt.Errorf("Trim expects string argument")
      }
      return StringValue(strings.TrimSpace(args[0].AsString())), nil
  }
  ```

- [ ] **14.16.3.2** Add builtinTrimLeft function to VM
  - **File**: `internal/bytecode/vm_builtins_string.go`
  - **Task**: Implement TrimLeft(str) - removes leading whitespace
  ```go
  func builtinTrimLeft(vm *VM, args []Value) (Value, error) {
      // Similar to builtinTrim but use strings.TrimLeft(str, " \t\n\r")
  }
  ```

- [ ] **14.16.3.3** Add builtinTrimRight function to VM
  - **File**: `internal/bytecode/vm_builtins_string.go`
  - **Task**: Implement TrimRight(str) - removes trailing whitespace
  ```go
  func builtinTrimRight(vm *VM, args []Value) (Value, error) {
      // Similar to builtinTrim but use strings.TrimRight(str, " \t\n\r")
  }
  ```

- [ ] **14.16.3.4** Register Trim functions in VM builtin table
  - **File**: `internal/bytecode/vm_builtins.go` or appropriate registration file
  - **Task**: Add entries to builtin function map
  ```go
  "trim":      builtinTrim,
  "trimleft":  builtinTrimLeft,
  "trimright": builtinTrimRight,
  ```

- [ ] **14.16.3.5** Enable String.Trim helper method in VM
  - **File**: `internal/bytecode/vm_calls.go`
  - **Task**: Add full implementation for the "trim" case in string helper method dispatch
    (currently only case label exists). Implementation should validate exactly one string
    argument and call `builtinTrim`, matching the pattern used for "toupper"/"tolower" cases.

- [ ] **14.16.3.6** Add Trim builtin tests
  - **File**: `internal/bytecode/vm_builtins_string_test.go`
  - **Tests**:
    - Trim with leading spaces
    - Trim with trailing spaces
    - Trim with both
    - Trim with tabs and newlines
    - TrimLeft and TrimRight variations
    - Empty string edge case

**Acceptance Criteria**:
- Trim() function works in bytecode VM
- TrimLeft() and TrimRight() work in bytecode VM
- String.Trim helper method works
- Tests pass

---

#### 14.16 Summary

| Task | Priority | Complexity | Status | Blocking |
|------|----------|------------|--------|----------|
| 14.16.1 For Loop Support | 🔴 HIGH | High | Not Started | VM cannot execute for loops |
| 14.16.2 Result Variable | 🔴 HIGH | Medium | Not Started | Functions cannot use Result |
| 14.16.3 Trim Builtin | 🟡 MEDIUM | Low | Not Started | Trim() calls fail in VM |

**Estimated Effort**:
- 14.16.1 For Loop Support: 2-3 days
- 14.16.2 Result Variable: 1-2 days
- 14.16.3 Trim Builtin: 0.5 day

**Total**: ~4-6 days

**Dependencies**:
- None (all tasks can be done independently)

**Testing Strategy**:
1. Unit tests for each feature
2. VM parity tests comparing output with AST interpreter
3. Integration tests with complex scripts using all features

[ ] 14.17 Add enum support to bytecode VM
  - **Task**: Implement enum value representation and operations in bytecode VM
  - **Priority**: MEDIUM (enums work fine in AST interpreter, this is optimization)
  - **Status**: NOT STARTED
  - **Implementation**:
    1. Add `ValueEnum` to ValueType enum in `bytecode.go`
    2. Add `EnumValue()` helper constructor for enum values
    3. Modify compiler to handle enum expressions:
       - Scoped enum access (e.g., `TEnum.Alpha`)
       - Enum type declarations in symbol table
       - Member access expressions for enums
    4. Add enum operations to VM:
       - Comparison operations (=, <>, <, >, <=, >=)
       - Boolean operations (and, or, xor) - Task 1.6 support
       - Implicit enum-to-integer conversion - Task 1.6 support
    5. Update bytecode serialization to support enum values
    6. Add comprehensive tests for enum operations in VM
  - **Current workaround**: Use AST interpreter for scripts with enums
  - **Failing Test**: `enum_bool_op.pas` fails with bytecode VM
  - **Example**:
    ```pascal
    type TEnum = flags (Alpha, Beta, Gamma);
    var x := TEnum.Alpha or TEnum.Gamma;  // Currently unsupported in VM
    PrintInt(TEnum.Beta);                  // Currently unsupported in VM
    ```
  - **Files to Modify**:
    - `internal/bytecode/bytecode.go` (add ValueEnum type)
    - `internal/bytecode/compiler_expressions.go` (compile enum literals and member access)
    - `internal/bytecode/compiler_core.go` (symbol table for enum types)
    - `internal/bytecode/vm_exec.go` (execute enum operations)
    - `internal/bytecode/serializer.go` (serialize/deserialize enum values)
  - **Estimated time**: 4-6 hours (1 day)
  - **Dependencies**: None (enum support complete in AST interpreter and semantic analyzer)

---

**Estimated time**: Completed in 12-16 weeks

## Phase 15: Future Bytecode Optimizations (DEFERRED)

- [ ] 15.1 Advanced peephole optimizations
  - [ ] Strength reduction (multiplication → shift)
  - [ ] Common subexpression elimination
  - [ ] Branch prediction hints

- [ ] 15.2 Register allocation improvements
  - [ ] Live range analysis
  - [ ] Register coloring for locals
  - [ ] Reduce stack traffic

- [ ] 15.3 Inline caching for method dispatch
  - [ ] Cache method lookup results
  - [ ] Invalidate on class redefinition
  - [ ] Benchmark polymorphic call sites

- [ ] 15.4 Bytecode verification
  - [ ] Static analysis of bytecode correctness
  - [ ] Type safety verification
  - [ ] Stack depth validation

---

## Phase 16: Performance & Polish

### Performance Profiling

- [x] 16.1 Create performance benchmark scripts
- [x] 16.2 Profile lexer performance: `BenchmarkLexer`
- [x] 16.3 Profile parser performance: `BenchmarkParser`
- [x] 16.4 Profile interpreter performance: `BenchmarkInterpreter`
- [x] 16.5 Identify bottlenecks using `pprof`
- [ ] 16.6 Document performance baseline

### Optimization - Lexer

- [ ] 16.7 Optimize string handling in lexer (use bytes instead of runes where possible)
- [ ] 16.8 Reduce allocations in token creation
- [ ] 16.9 Use string interning for keywords/identifiers
- [ ] 16.10 Benchmark improvements

### Optimization - Parser

- [ ] 16.11 Reduce AST node allocations
- [ ] 16.12 Pool commonly created nodes
- [ ] 16.13 Optimize precedence table lookups
- [ ] 16.14 Benchmark improvements

### Optimization - Interpreter

- [ ] 16.15 Optimize value representation (avoid interface{} overhead if possible)
- [ ] 16.16 Use switch statements instead of type assertions where possible
- [ ] 16.17 Cache frequently accessed symbols
- [ ] 16.18 Optimize environment lookups
- [ ] 16.19 Reduce allocations in hot paths
- [ ] 16.20 Benchmark improvements

### Memory Management

- [ ] 16.21 Ensure no memory leaks in long-running scripts
- [ ] 16.22 Profile memory usage with large programs
- [ ] 16.23 Optimize object allocation/deallocation
- [ ] 16.24 Consider object pooling for common types

### Code Quality Refactoring

- [ ] 16.25 Run `go vet ./...` and fix all issues
- [ ] 16.26 Run `golangci-lint run` and address warnings
- [ ] 16.27 Run `gofmt` on all files
- [ ] 16.28 Run `goimports` to organize imports
- [ ] 16.29 Review error handling consistency
- [ ] 16.30 Unify value representation if inconsistent
- [ ] 16.31 Refactor large functions into smaller ones
- [ ] 16.32 Extract common patterns into helper functions
- [ ] 16.33 Improve variable/function naming
- [ ] 16.34 Add missing error checks

### Documentation

- [ ] 16.35 Write comprehensive GoDoc comments for all exported types/functions
- [ ] 16.36 Document internal architecture in `docs/architecture.md`
- [ ] 16.37 Create user guide in `docs/user_guide.md`
- [ ] 16.38 Document CLI usage with examples
- [ ] 16.39 Create API documentation for embedding the library
- [ ] 16.40 Add code examples to documentation
- [ ] 16.41 Document known limitations
- [ ] 16.42 Create contribution guidelines in `CONTRIBUTING.md`

### Example Programs

- [x] 16.43 Create `examples/` directory
- [x] 16.44 Add example scripts:
  - [x] Hello World
  - [x] Fibonacci
  - [x] Factorial
  - [x] Class-based example (Person demo)
  - [x] Algorithm sample (math/loops showcase)
- [x] 16.45 Add README in examples directory
- [x] 16.46 Ensure all examples run correctly

### Testing Enhancements

- [ ] 16.47 Add integration tests in `test/integration/`
- [ ] 16.48 Add fuzzing tests for parser: `FuzzParser`
- [ ] 16.49 Add fuzzing tests for lexer: `FuzzLexer`
- [ ] 16.50 Add property-based tests (using testing/quick or gopter)
- [ ] 16.51 Ensure CI runs all test types
- [ ] 16.52 Achieve >90% code coverage overall
- [ ] 16.53 Add regression tests for all fixed bugs

### Release Preparation

- [ ] 16.54 Create `CHANGELOG.md`
- [ ] 16.55 Document version numbering scheme (SemVer)
- [ ] 16.56 Tag v0.1.0 alpha release
- [ ] 16.57 Create release binaries for major platforms (Linux, macOS, Windows)
- [ ] 16.58 Publish release on GitHub
- [ ] 16.59 Write announcement blog post or README update
- [ ] 16.60 Share with community for feedback

---

## Phase 17: Go Source Code Generation & AOT Compilation [RECOMMENDED]

**Status**: Not started | **Priority**: HIGH | **Estimated Time**: 20-28 weeks (code generation) + 9-13 weeks (CLI)

### Overview

This phase implements ahead-of-time (AOT) compilation by transpiling DWScript source code to Go, then compiling to native executables. This approach leverages Go's excellent cross-compilation support and delivers near-native performance.

**Approach**: DWScript Source → AST → Go Source Code → Go Compiler → Native Executable

**Benefits**: 10-50x faster than tree-walking interpreter, excellent portability, leverages Go toolchain

### Phase 13.1: Go Source Code Generation (20-28 weeks)

- [ ] 17.1 Design Go code generation architecture
  - Study similar transpilers (c2go, ast-transpiler)
  - Design AST → Go AST transformation strategy
  - Define runtime library interface
  - Document type mapping (DWScript → Go)
  - Plan package structure for generated code
  - **Decision**: Use `go/ast` package for Go AST generation

- [ ] 17.2 Create Go code generator foundation
  - Create `internal/codegen/` package
  - Create `internal/codegen/go_generator.go`
  - Implement `Generator` struct with context tracking
  - Add helper methods for code emission
  - Set up `go/ast` and `go/printer` integration
  - Create unit tests for basic generation

- [ ] 17.3 Implement type system mapping
  - Map DWScript primitives to Go types (Integer→int64, Float→float64, String→string, Boolean→bool)
  - Map DWScript arrays to Go slices (dynamic) or arrays (static)
  - Map DWScript records to Go structs
  - Map DWScript classes to Go structs with method tables
  - Handle type aliases and subrange types
  - Document type mapping in `docs/codegen-types.md`

- [ ] 17.4 Generate code for expressions
  - Generate literals (integer, float, string, boolean, nil)
  - Generate identifiers (variables, constants)
  - Generate binary operations (+, -, *, /, =, <>, <, >, etc.)
  - Generate unary operations (-, not)
  - Generate function calls
  - Generate array/object member access
  - Handle operator precedence correctly
  - Add unit tests comparing eval vs generated code

- [ ] 17.5 Generate code for statements
  - Generate variable declarations (`var x: Integer = 42`)
  - Generate assignments (`x := 10`)
  - Generate if/else statements
  - Generate while/repeat/for loops
  - Generate case statements (switch in Go)
  - Generate begin...end blocks
  - Handle break/continue/exit statements

- [ ] 17.6 Generate code for functions and procedures
  - Generate function declarations with parameters and return type
  - Handle by-value and by-reference (var) parameters
  - Generate procedure declarations (no return value)
  - Implement nested functions (closures in Go)
  - Support forward declarations
  - Handle recursion
  - Generate proper variable scoping

- [ ] 17.7 Generate code for classes and OOP
  - Generate Go struct definitions for classes
  - Generate constructor functions (Create)
  - Generate destructor cleanup (Destroy → defer)
  - Generate method declarations (receiver functions)
  - Implement inheritance (embedding in Go)
  - Implement virtual method dispatch (method tables)
  - Handle class fields and properties
  - Support `Self` keyword (receiver parameter)

- [ ] 17.8 Generate code for interfaces
  - Generate Go interface definitions
  - Implement interface casting and type assertions
  - Generate interface method dispatch
  - Handle interface inheritance
  - Support interface variables and parameters

- [ ] 17.9 Generate code for records
  - Generate Go struct definitions
  - Support record methods (static and instance)
  - Handle record literals and initialization
  - Generate record field access

- [ ] 17.10 Generate code for enums
  - Generate Go const declarations with iota
  - Support scoped and unscoped enum access
  - Generate Ord() and Integer() conversions
  - Handle explicit enum values

- [ ] 17.11 Generate code for arrays
  - Generate static arrays (Go arrays: `[10]int`)
  - Generate dynamic arrays (Go slices: `[]int`)
  - Support array literals
  - Generate array indexing and slicing
  - Implement SetLength, High, Low built-ins
  - Handle multi-dimensional arrays

- [ ] 17.12 Generate code for sets
  - Generate set types as Go map[T]bool or bitsets
  - Support set literals and constructors
  - Generate set operations (union, intersection, difference)
  - Implement `in` operator for set membership

- [ ] 17.13 Generate code for properties
  - Translate properties to getter/setter methods
  - Generate field-backed properties (direct access)
  - Generate method-backed properties (method calls)
  - Support read-only and write-only properties
  - Handle auto-properties

- [ ] 17.14 Generate code for exceptions
  - Generate try/except/finally as Go defer/recover
  - Map DWScript exceptions to Go error types
  - Generate raise statements (panic)
  - Implement exception class hierarchy
  - Preserve stack traces

- [ ] 17.15 Generate code for operators and conversions
  - Generate operator overloads as functions
  - Generate implicit conversions
  - Handle type coercion in expressions
  - Support custom operators

- [ ] 17.16 Create runtime library for generated code
  - Create `pkg/runtime/` package
  - Implement built-in functions (PrintLn, Length, Copy, etc.)
  - Implement array/string manipulation functions
  - Implement math functions (Sin, Cos, Sqrt, etc.)
  - Implement date/time functions
  - Provide runtime type information (RTTI) for reflection
  - Support external function calls (FFI)

- [ ] 17.17 Handle units/modules compilation
  - Generate separate Go packages for each unit
  - Handle unit dependencies and imports
  - Generate initialization/finalization code
  - Support uses clauses
  - Create package manifest

- [ ] 17.18 Implement optimization passes
  - Constant folding
  - Dead code elimination
  - Inline small functions
  - Remove unused variables
  - Optimize string concatenation
  - Use Go compiler optimization hints (//go:inline, etc.)

- [ ] 17.19 Add source mapping for debugging
  - Preserve line number comments in generated code
  - Generate source map files (.map)
  - Add DWScript source file embedding
  - Support stack trace translation (Go → DWScript)

- [ ] 17.20 Test Go code generation
  - Generate code for all fixture tests
  - Compile and run generated code
  - Compare output with interpreter
  - Measure compilation time
  - Benchmark generated code performance

**Expected Results**: 10-50x faster than tree-walking interpreter, near-native Go speed

## Phase 18: AOT Compiler CLI (9-13 weeks)

- [ ] 18.1 Create `dwscript compile` command
  - Add `compile` subcommand to CLI
  - Parse input DWScript file(s)
  - Generate Go source code to output directory
  - Invoke `go build` to create executable
  - Support multiple output formats (executable, library, package)

- [ ] 18.2 Implement project compilation mode
  - Support compiling entire projects (multiple units)
  - Generate go.mod file
  - Handle dependencies between units
  - Create main package with entry point
  - Support compilation configuration (optimization level, target platform)

- [ ] 18.3 Add compilation flags and options
  - `--output` or `-o` for output path
  - `--optimize` or `-O` for optimization level (0, 1, 2, 3)
  - `--keep-go-source` to preserve generated Go files
  - `--target` for cross-compilation (linux, windows, darwin, wasm)
  - `--static` for static linking
  - `--debug` to include debug symbols

- [ ] 18.4 Implement cross-compilation support
  - Support GOOS and GOARCH environment variables
  - Generate platform-specific code (if needed)
  - Test compilation for Linux, macOS, Windows, WASM
  - Document platform-specific limitations

- [ ] 18.5 Add incremental compilation
  - Cache compiled units
  - Detect file changes (mtime, hash)
  - Recompile only changed units
  - Rebuild dependency graph
  - Speed up repeated compilations

- [ ] 18.6 Create standalone binary builder
  - Generate single-file executable
  - Embed DWScript runtime
  - Strip debug symbols (optional)
  - Compress binary with UPX (optional)
  - Test on different platforms

- [ ] 18.7 Implement library compilation mode
  - Generate Go package (not executable)
  - Export public functions/classes
  - Create Go-friendly API
  - Generate documentation (godoc)
  - Support embedding in other Go projects

- [ ] 18.8 Add compilation error reporting
  - Catch Go compilation errors
  - Translate errors to DWScript source locations
  - Provide helpful error messages
  - Suggest fixes for common issues

- [ ] 18.9 Create compilation test suite
  - Test compilation of all fixture tests
  - Verify all executables run correctly
  - Test cross-compilation
  - Benchmark compilation speed
  - Measure binary sizes

- [ ] 18.10 Document AOT compilation
  - Write `docs/aot-compilation.md`
  - Explain compilation process
  - Provide usage examples
  - Document performance characteristics
  - Compare with interpretation and bytecode VM

---

## Phase 19: WebAssembly Runtime & Playground ✅ MOSTLY COMPLETE

**Status**: Core implementation complete | **Priority**: HIGH | **Tasks**: 23 complete, 3 pending

### Overview

This phase implements WebAssembly support for running DWScript in browsers, including a platform abstraction layer, WASM build infrastructure, JavaScript/Go bridge, and a web-based playground with Monaco editor integration.

**Architecture**: DWScript → WASM Binary → Browser/Node.js → JavaScript API

### Phase 19.1: Platform Abstraction Layer ✅ COMPLETE

- [x] 19.1 Create `pkg/platform/` package with core interfaces
  - FileSystem, Console, Platform interfaces
  - Enables native and WebAssembly builds with consistent behavior

- [x] 19.2 Implement `pkg/platform/native/` for standard Go
  - Standard Go implementations for native builds
  - Direct OS filesystem and console access

- [x] 19.3 Implement `pkg/platform/wasm/` with virtual filesystem
  - In-memory map for file storage
  - Console bridge to JavaScript console.log
  - Time functions using JavaScript Date API
  - Sleep implementation using setTimeout

- [ ] 19.4 Create feature parity test suite
  - Tests that run on both native and WASM
  - Validate platform abstraction works correctly

- [ ] 19.5 Document platform differences and limitations
  - Platform-specific behavior documentation
  - Known limitations in WASM environment

### Phase 19.2: WASM Build Infrastructure ✅ COMPLETE

- [x] 19.6 Create build infrastructure
  - `build/wasm/` directory with scripts
  - Justfile targets: `just wasm`, `just wasm-test`, `just wasm-optimize`, etc.
  - `cmd/dwscript-wasm/main.go` entry point with syscall/js exports

- [x] 19.7 Implement build modes support
  - Monolithic, modular, hybrid modes (compile-time flags)
  - `pkg/wasm/` package for WASM bridge code

- [x] 19.8 Add wasm_exec.js and optimization
  - wasm_exec.js from Go distribution (multi-version support)
  - Integrate wasm-opt (Binaryen) for binary size optimization
  - Size monitoring (warns if >3MB uncompressed)

- [ ] 19.9 Test all build modes
  - Compare sizes and performance
  - Validate each mode works correctly

- [x] 19.10 Document build process
  - `docs/wasm/BUILD.md` with build instructions
  - Configuration options and troubleshooting

### Phase 19.3: JavaScript/Go Bridge ✅ COMPLETE

- [x] 19.11 Implement DWScript class API
  - `pkg/wasm/api.go` using syscall/js
  - Export init(), compile(), run(), eval() to JavaScript

- [x] 19.12 Create type conversion utilities
  - Go types ↔ js.Value conversion in utils.go
  - Proper handling of DWScript types in JavaScript

- [x] 19.13 Implement callback registration system
  - `pkg/wasm/callbacks.go` for event handling
  - Virtual filesystem interface for JavaScript

- [x] 19.14 Add error handling across boundary
  - Panics → exceptions with recovery
  - Structured error objects for DWScript runtime errors

- [x] 19.15 Add event system
  - on() method for output, error, and custom events
  - Memory management with proper js.Value.Release()

- [x] 19.16 Document JavaScript API
  - `docs/wasm/API.md` with complete API reference
  - Usage examples for browser and Node.js

### Phase 19.4: Web Playground ✅ COMPLETE

- [x] 19.17 Create playground directory structure
  - `playground/` with HTML/CSS/JS files
  - Monaco Editor integration

- [x] 19.18 Implement syntax highlighting
  - DWScript language definition for Monaco
  - Tokenization rules matching lexer

- [x] 19.19 Build split-pane UI
  - Code editor + output console
  - Toolbar with Run, Examples, Clear, Share, Theme buttons

- [x] 19.20 Implement URL-based code sharing
  - Base64 encoded code in fragment
  - Examples dropdown with sample programs

- [x] 19.21 Add localStorage features
  - Auto-save and restore user code
  - Error markers in editor from compilation errors

- [x] 19.22 Set up GitHub Pages deployment
  - GitHub Actions workflow for automated deployment
  - Testing checklist in playground/TESTING.md

- [x] 19.23 Document playground architecture
  - `docs/wasm/PLAYGROUND.md` with architecture details
  - Extension points for future features

### Phase 19.5: NPM Package ✅ MOSTLY COMPLETE

- [x] 19.24 Create NPM package structure
  - `npm/` with package.json
  - TypeScript definitions in `typescript/index.d.ts`

- [x] 19.25 Create dual ESM/CommonJS entry points
  - index.js (ESM) and index.cjs (CommonJS)
  - WASM loader helper for Node.js and browser

- [x] 19.26 Add usage examples
  - Node.js, React, Vue, vanilla JS examples
  - Automated NPM publishing via GitHub Actions

- [x] 19.27 Configure for tree-shaking
  - Optimal bundling configuration
  - `npm/README.md` with installation guide

- [ ] 19.28 Publish to npmjs.com
  - Initial version publication
  - Version management strategy

### Phase 19.6: Testing & Documentation

- [ ] 19.29 Write WASM-specific tests
  - GOOS=js GOARCH=wasm go test
  - Node.js integration test suite

- [ ] 19.30 Add browser tests
  - Playwright tests for Chrome, Firefox, Safari
  - CI matrix for cross-browser testing

- [ ] 19.31 Add performance benchmarks
  - Compare WASM vs native speed
  - Bundle size regression monitoring in CI

- [ ] 19.32 Write embedding guide
  - `docs/wasm/EMBEDDING.md` for web app integration
  - Update main README with WASM section and playground link

---

## Phase 20: Community Building & Ecosystem [ONGOING]

**Status**: Ongoing | **Priority**: HIGH | **Estimated Tasks**: ~40

### Overview

This phase focuses on building a sustainable open-source community, maintaining feature parity with upstream DWScript, and providing essential tools for developers including REPL, debugging support, and platform-specific enhancements.

### Phase 20.1: Feature Parity Tracking

- [ ] 20.1 Create feature matrix comparing go-dws with DWScript
- [ ] 20.2 Track DWScript upstream releases
- [ ] 20.3 Identify new features in DWScript updates
- [ ] 20.4 Plan integration of new features
- [ ] 20.5 Update feature matrix regularly

### Phase 20.2: Community Building

- [ ] 20.6 Set up issue templates on GitHub
- [ ] 20.7 Set up pull request template
- [ ] 20.8 Create CODE_OF_CONDUCT.md
- [ ] 20.9 Create discussions forum or mailing list
- [ ] 20.10 Encourage contributions (tag "good first issue")
- [ ] 20.11 Respond to issues and PRs promptly
- [ ] 20.12 Build maintainer team (if interest grows)

### Phase 20.3: Advanced Features

- [ ] 20.13 Implement REPL (Read-Eval-Print Loop):
  - [ ] Interactive prompt
  - [ ] Statement-by-statement execution
  - [ ] Variable inspection
  - [ ] History and autocomplete
- [ ] 20.14 Implement debugging support:
  - [ ] Breakpoints
  - [ ] Step-through execution
  - [ ] Variable inspection
  - [ ] Stack traces
- [x] 20.15 WebAssembly Runtime & Playground - **See Phase 14** (mostly complete)
- [x] 20.16 Language Server Protocol (LSP) - **See external repo** https://github.com/cwbudde/go-dws-lsp
- [ ] 20.17 JavaScript code generation backend - **See Phase 16** (deferred)

### Phase 20.4: Platform-Specific Enhancements

- [ ] 20.18 Add Windows-specific features (if needed)
- [ ] 20.19 Add macOS-specific features (if needed)
- [ ] 20.20 Add Linux-specific features (if needed)
- [ ] 20.21 Test on multiple architectures (ARM, AMD64)

### Phase 20.5: Edge Case Audit

- [ ] 20.22 Test short-circuit evaluation (and, or)
- [ ] 20.23 Test operator precedence edge cases
- [ ] 20.24 Test division by zero handling
- [ ] 20.25 Test integer overflow behavior
- [ ] 20.26 Test floating-point edge cases (NaN, Inf)
- [ ] 20.27 Test string encoding (UTF-8 handling)
- [ ] 20.28 Test very large programs (scalability)
- [ ] 20.29 Test deeply nested structures
- [ ] 20.30 Test circular references (if possible in language)
- [ ] 20.31 Fix any discovered issues

### Phase 20.6: Performance Monitoring

- [ ] 20.32 Set up continuous performance benchmarking
- [ ] 20.33 Track performance metrics over releases
- [ ] 20.34 Identify and fix performance regressions
- [ ] 20.35 Publish performance comparison with DWScript

### Phase 20.7: Security Audit

- [ ] 20.36 Review for potential security issues (untrusted script execution)
- [ ] 20.37 Implement resource limits (memory, execution time)
- [ ] 20.38 Implement sandboxing for untrusted scripts
- [ ] 20.39 Audit for code injection vulnerabilities
- [ ] 20.40 Document security best practices

### Phase 20.8: Maintenance

- [ ] 20.41 Keep dependencies up to date
- [ ] 20.42 Monitor Go version updates and migrate as needed
- [ ] 20.43 Maintain CI/CD pipeline
- [ ] 20.44 Regular code reviews
- [ ] 20.45 Address technical debt periodically

### Phase 20.9: Long-term Roadmap

- [ ] 20.46 Define 1-year roadmap
- [ ] 20.47 Define 3-year roadmap
- [ ] 20.48 Gather user feedback and adjust priorities
- [ ] 20.49 Consider commercial applications/support
- [ ] 20.50 Explore academic/research collaborations

---

## Phase 21: MIR Foundation [DEFERRED]

**Status**: Not started | **Priority**: MEDIUM | **Estimated Tasks**: 47 (MIR core, lowering, testing)

### Overview

This phase implements a Mid-level Intermediate Representation (MIR) that serves as a target-neutral layer between the type-checked AST and backend code generators. The MIR enables multiple backend targets (JavaScript, LLVM, C, etc.) from a single lowering pass.

**Architecture Flow**: DWScript Source → Parser → Semantic Analyzer → **MIR Builder** → [Backend Emitter] → Output

**Why MIR?** Clean separation of concerns, multi-backend support, platform-independent optimizations, easier debugging, and future-proofing for additional compilation targets.

**Note**: JavaScript backend is implemented in Phase 16, LLVM backend in Phase 17.

### Phase 21.1: MIR Foundation (30 tasks)

**Goal**: Define a complete, verifiable mid-level IR that can represent all DWScript constructs in a target-neutral way.

**Exit Criteria**: MIR spec documented, complete type system, builder API, verifier, AST→MIR lowering for ~80% of constructs, 20+ golden tests, 85%+ coverage

#### 21.1.1: MIR Package Structure and Types (10 tasks)

- [ ] 21.1 Create `mir/` package directory
- [ ] 21.2 Create `mir/types.go` - MIR type system
- [ ] 21.3 Define `Type` interface with `String()`, `Size()`, `Align()` methods
- [ ] 21.4 Implement primitive types: `Bool`, `Int8`, `Int16`, `Int32`, `Int64`, `Float32`, `Float64`, `String`
- [ ] 21.5 Implement composite types: `Array(elemType, size)`, `Record(fields)`, `Pointer(pointeeType)`
- [ ] 21.6 Implement OOP types: `Class(name, fields, methods, parent)`, `Interface(name, methods)`
- [ ] 21.7 Implement function types: `Function(params, returnType)`
- [ ] 21.8 Add `Void` type for procedures
- [ ] 21.9 Implement type equality and compatibility checking
- [ ] 21.10 Implement type conversion rules (explicit vs implicit)

#### 21.1.2: MIR Instructions and Control Flow (10 tasks)

- [ ] 21.11 Create `mir/instruction.go` - MIR instruction set
- [ ] 21.12 Define `Instruction` interface with `ID()`, `Type()`, `String()` methods
- [ ] 21.13 Implement arithmetic ops: `Add`, `Sub`, `Mul`, `Div`, `Mod`, `Neg`
- [ ] 21.14 Implement comparison ops: `Eq`, `Ne`, `Lt`, `Le`, `Gt`, `Ge`
- [ ] 21.15 Implement logical ops: `And`, `Or`, `Xor`, `Not`
- [ ] 21.16 Implement memory ops: `Alloca`, `Load`, `Store`
- [ ] 21.17 Implement constants: `ConstInt`, `ConstFloat`, `ConstString`, `ConstBool`, `ConstNil`
- [ ] 21.18 Implement conversions: `IntToFloat`, `FloatToInt`, `IntTrunc`, `IntExt`
- [ ] 21.19 Implement function ops: `Call`, `VirtualCall`
- [ ] 21.20 Implement array/class ops: `ArrayAlloc`, `ArrayLen`, `ArrayIndex`, `ArraySet`, `FieldGet`, `FieldSet`, `New`

#### 21.1.3: MIR Control Flow Structures (5 tasks)

- [ ] 21.21 Create `mir/block.go` - Basic blocks with `ID`, `Instructions`, `Terminator`
- [ ] 21.22 Implement control flow terminators: `Phi`, `Br`, `CondBr`, `Return`, `Throw`
- [ ] 21.23 Implement terminator validation (every block must end with terminator)
- [ ] 21.24 Implement block predecessors/successors tracking for CFG
- [ ] 21.25 Create `mir/function.go` - Function representation with `Name`, `Params`, `ReturnType`, `Blocks`, `Locals`

#### 21.1.4: MIR Builder API (3 tasks)

- [ ] 21.26 Create `mir/builder.go` - Safe MIR construction
- [ ] 21.27 Implement `Builder` struct with function/block context, `NewFunction()`, `NewBlock()`, `SetInsertPoint()`
- [ ] 21.28 Implement instruction emission methods: `EmitAdd()`, `EmitLoad()`, `EmitStore()`, etc. with type checking

#### 21.1.5: MIR Verifier (2 tasks)

- [ ] 21.29 Create `mir/verifier.go` - MIR correctness checking
- [ ] 21.30 Implement CFG, type, SSA, and function signature verification with `Verify(fn *Function) []error` API

### Phase 15.2: AST → MIR Lowering (12 tasks)

- [ ] 21.31 Create `mir/lower.go` - AST to MIR translation
- [ ] 21.32 Implement `LowerProgram(ast *ast.Program) (*mir.Module, error)` entry point
- [ ] 21.33 Lower expressions: literals → `Const*` instructions
- [ ] 21.34 Lower binary operations → corresponding MIR ops (handle short-circuit for `and`/`or`)
- [ ] 21.35 Lower unary operations → `Neg`, `Not`
- [ ] 21.36 Lower identifier references → `Load` instructions
- [ ] 21.37 Lower function calls → `Call` instructions
- [ ] 21.38 Lower array indexing → `ArrayIndex` + bounds check insertion
- [ ] 21.39 Lower record field access → `FieldGet`/`FieldSet`
- [ ] 21.40 Lower statements: variable declarations, assignments, if/while/for, return
- [ ] 21.41 Lower declarations: functions/procedures, records, classes
- [ ] 21.42 Implement short-circuit evaluation and simple optimizations (constant folding, dead code elimination)

### Phase 15.3: MIR Debugging and Testing (5 tasks)

- [ ] 21.43 Create `mir/dump.go` - Human-readable MIR output with `Dump(fn *Function) string`
- [ ] 21.44 Integration with CLI: `./bin/dwscript dump-mir script.dws`
- [ ] 21.45 Create golden MIR tests: 5+ each for expressions, control flow, functions, advanced features
- [ ] 21.46 Implement MIR verifier tests: type mismatches, malformed CFG, SSA violations
- [ ] 21.47 Implement round-trip tests: AST → MIR → verify → dump → compare with golden files

---

## Phase 22: JavaScript Backend [DEFERRED]

**Status**: Not started | **Priority**: MEDIUM | **Estimated Tasks**: 105 (MVP + feature complete)

### Overview

This phase implements a JavaScript code generator that translates MIR to readable, runnable JavaScript. The backend enables running DWScript programs in browsers and Node.js environments.

**Architecture Flow**: MIR → JavaScript Emitter → JavaScript Code → Node.js/Browser

**Benefits**: Browser execution, npm ecosystem integration, excellent portability, leverages JavaScript JIT compilers

**Dependencies**: Requires Phase 15 (MIR Foundation) to be completed first

### Phase 22.1: JS Backend MVP (45 tasks)

**Goal**: Implement a JavaScript code generator that can compile basic DWScript programs to readable, runnable JavaScript.

**Exit Criteria**: JS emitter for expressions/control flow/functions, 20+ end-to-end tests (DWScript→JS→execute), golden JS snapshots, 85%+ coverage

#### 22.4.1: JS Emitter Infrastructure (8 tasks)

- [ ] 22.48 Create `codegen/` package with `Backend` interface and `EmitterOptions`
- [ ] 22.49 Create `codegen/js/` package and `emitter.go`
- [ ] 22.50 Define `JSEmitter` struct with `out`, `indent`, `opts`, `tmpCounter`
- [ ] 22.51 Implement helper methods: `emit()`, `emitLine()`, `emitIndent()`, `pushIndent()`, `popIndent()`
- [ ] 22.52 Implement `newTemp()` for temporary variable naming
- [ ] 22.53 Implement `NewJSEmitter(opts EmitterOptions)`
- [ ] 22.54 Implement `Generate(module *mir.Module) (string, error)` entry point
- [ ] 22.55 Test emitter infrastructure

#### 22.4.2: Module and Function Emission (6 tasks)

- [ ] 22.56 Implement module structure emission: ES Module format with `export`, file header comment
- [ ] 22.57 Implement optional IIFE fallback via `EmitterOptions`
- [ ] 22.58 Implement function emission: `function fname(params) { ... }`
- [ ] 22.59 Map DWScript params to JS params (preserve names)
- [ ] 22.60 Emit local variable declarations at function top (from `Alloca` instructions)
- [ ] 22.61 Handle procedures (no return value) as JS functions

#### 22.4.3: Expression and Instruction Lowering (12 tasks)

- [ ] 22.62 Lower arithmetic operations → JS infix operators: `+`, `-`, `*`, `/`, `%`, unary `-`
- [ ] 22.63 Lower comparison operations → JS comparisons: `===`, `!==`, `<`, `<=`, `>`, `>=`
- [ ] 22.64 Lower logical operations → JS boolean ops: `&&`, `||`, `!`
- [ ] 22.65 Lower constants → JS literals with proper escaping
- [ ] 22.66 Lower variable operations: `Load` → variable reference, `Store` → assignment
- [ ] 22.67 Lower function calls: `Call` → `functionName(args)`
- [ ] 22.68 Implement Phi node lowering with temporary variables at block edges
- [ ] 22.69 Test expression lowering
- [ ] 22.70 Test instruction lowering
- [ ] 22.71 Test temporary variable generation
- [ ] 22.72 Test type conversions
- [ ] 22.73 Test complex expressions

#### 22.4.4: Control Flow Emission (8 tasks)

- [ ] 22.74 Implement control flow reconstruction from MIR CFG
- [ ] 22.75 Detect if/else patterns from `CondBr`
- [ ] 22.76 Detect while loop patterns (backedge to header)
- [ ] 22.77 Emit if-else: `if (condition) { ... } else { ... }`
- [ ] 22.78 Emit while loops: `while (condition) { ... }`
- [ ] 22.79 Emit for loops if MIR preserves metadata
- [ ] 22.80 Handle unconditional branches
- [ ] 22.81 Handle return statements

#### 22.4.5: Runtime and Testing (11 tasks)

- [ ] 22.82 Create `runtime/js/runtime.js` with `_dws.boundsCheck()`, `_dws.assert()`
- [ ] 22.83 Emit runtime import in generated JS (if needed)
- [ ] 22.84 Make runtime usage optional via `EmitterOptions.InsertBoundsChecks`
- [ ] 22.85 Create `codegen/js/testdata/` with subdirectories
- [ ] 22.86 Implement golden JS snapshot tests
- [ ] 22.87 Setup Node.js in CI (GitHub Actions)
- [ ] 22.88 Implement execution tests: parse → lower → generate → execute → verify
- [ ] 22.89 Add end-to-end tests for arithmetic, control flow, functions, loops
- [ ] 22.90 Add unit tests for JS emitter
- [ ] 22.91 Achieve 85%+ coverage for `codegen/js/` package
- [ ] 22.92 Add `compile-js` CLI command: `./bin/dwscript compile-js input.dws -o output.js`

### Phase 22.2: JS Feature Complete (60 tasks)

**Goal**: Extend JS backend to support all DWScript language features.

**Exit Criteria**: Full OOP, composite types, exceptions, properties, 50+ comprehensive tests, real-world samples work

#### 22.5.1: Records (7 tasks)

- [ ] 22.93 Implement MIR support for records
- [ ] 22.94 Emit records as plain JS objects: `{ x: 0, y: 0 }`
- [ ] 22.95 Implement constructor functions for records
- [ ] 22.96 Implement field access/assignment as property access
- [ ] 22.97 Implement record copy semantics with `_dws.copyRecord()`
- [ ] 22.98 Test record creation, initialization, field read/write
- [ ] 22.99 Test nested records and copy semantics

#### 22.5.2: Arrays (10 tasks)

- [ ] 22.100 Extend MIR for static and dynamic arrays
- [ ] 22.101 Emit static arrays as JS arrays with fixed size
- [ ] 22.102 Implement array index access with optional bounds checking
- [ ] 22.103 Emit dynamic arrays as JS arrays
- [ ] 22.104 Implement `SetLength` → `arr.length = newLen`
- [ ] 22.105 Implement `Length` → `arr.length`
- [ ] 22.106 Support multi-dimensional arrays (nested JS arrays)
- [ ] 22.107 Implement array operations: copy, concatenation
- [ ] 22.108 Test static array creation and indexing
- [ ] 22.109 Test dynamic array operations and bounds checking

#### 22.5.3: Classes and Inheritance (15 tasks)

- [ ] 22.110 Extend MIR for classes with fields, methods, parent, vtable
- [ ] 22.111 Emit ES6 class syntax: `class TAnimal { ... }`
- [ ] 22.112 Implement field initialization in constructor
- [ ] 22.113 Implement method emission
- [ ] 22.114 Implement inheritance with `extends` clause
- [ ] 22.115 Implement `super()` call in constructor
- [ ] 22.116 Handle virtual method dispatch (naturally virtual in JS)
- [ ] 22.117 Handle DWScript `Create` → JS `constructor`
- [ ] 22.118 Handle multiple constructors (overload dispatch)
- [ ] 22.119 Document destructor handling (no direct equivalent in JS)
- [ ] 22.120 Implement static fields and methods
- [ ] 22.121 Map `Self` → `this`, `inherited` → `super.method()`
- [ ] 22.122 Test simple classes with fields and methods
- [ ] 22.123 Test inheritance, virtual method overriding, constructors
- [ ] 22.124 Test static members and `Self`/`inherited` usage

#### 22.5.4: Interfaces (6 tasks)

- [ ] 22.125 Extend MIR for interfaces
- [ ] 22.126 Choose and document JS emission strategy (structural typing vs runtime metadata)
- [ ] 22.127 If using runtime metadata: emit interface tables, implement `is`/`as` operators
- [ ] 22.128 Test class implementing interface
- [ ] 22.129 Test interface method calls
- [ ] 22.130 Test `is` and `as` with interfaces

#### 22.5.5: Enums and Sets (8 tasks)

- [ ] 22.131 Extend MIR for enums
- [ ] 22.132 Emit enums as frozen JS objects: `const TColor = Object.freeze({...})`
- [ ] 22.133 Support scoped and unscoped enum access
- [ ] 22.134 Extend MIR for sets
- [ ] 22.135 Emit small sets (≤32 elements) as bitmasks
- [ ] 22.136 Emit large sets as JS `Set` objects
- [ ] 22.137 Implement set operations: union, intersection, difference, inclusion
- [ ] 22.138 Test enum declaration/usage and set operations

#### 22.5.6: Exception Handling (8 tasks)

- [ ] 22.139 Extend MIR for exceptions: `Throw`, `Try`, `Catch`, `Finally`
- [ ] 22.140 Emit `Throw` → `throw new Error()` or custom exception class
- [ ] 22.141 Emit try-except-finally → JS `try/catch/finally`
- [ ] 22.142 Create DWScript exception class → JS `Error` subclass
- [ ] 22.143 Handle `On E: ExceptionType do` with instanceof checks
- [ ] 22.144 Implement re-raise with exception tracking
- [ ] 22.145 Test basic try-except, multiple handlers, try-finally
- [ ] 22.146 Test re-raise and nested exception handling

#### 22.5.7: Properties and Advanced Features (6 tasks)

- [ ] 22.147 Extend MIR for properties with `PropGet`/`PropSet`
- [ ] 22.148 Emit properties as ES6 getters/setters
- [ ] 22.149 Handle indexed properties as methods
- [ ] 22.150 Test read/write properties and indexed properties
- [ ] 22.151 Implement operator overloading (desugar to method calls)
- [ ] 22.152 Implement generics support (monomorphization)

---

## Phase 23: LLVM Backend [DEFERRED]

**Status**: Not started | **Priority**: LOW | **Estimated Tasks**: 45

### Overview

This phase implements an LLVM IR backend for native code compilation, consolidating all LLVM-related work from the original Phase 13 LLVM JIT and AOT sections. This provides maximum performance but adds significant complexity.

**Architecture Flow**: MIR → LLVM IR Emitter → LLVM IR → llc → Native Binary

**Benefits**: Maximum performance (near C/C++ speed), excellent optimization, multi-architecture support

**Dependencies**: Requires Phase 15 (MIR Foundation) to be completed first

**Note**: This phase consolidates LLVM JIT (from old Phase 13.2), LLVM AOT (from old Phase 13.4), and LLVM backend (from old Stage 14.6). Given complexity and maintenance burden, this is marked as DEFERRED with LOW priority. The bytecode VM and Go AOT provide sufficient performance for most use cases.

### Phase 23.1: LLVM Infrastructure (8 tasks)

**Goal**: Set up LLVM bindings, type mapping, and runtime declarations

- [ ] 23.153 Choose LLVM binding: `llir/llvm` (pure Go) vs CGo bindings
- [ ] 23.154 Create `codegen/llvm/` package with `emitter.go`, `types.go`, `runtime.go`
- [ ] 23.155 Implement type mapping: DWScript types → LLVM types
- [ ] 23.156 Map Integer → `i32`/`i64`, Float → `double`, Boolean → `i1`
- [ ] 23.157 Map String → struct `{i32 len, i8* data}`
- [ ] 23.158 Map arrays/objects to LLVM structs
- [ ] 23.159 Emit LLVM module with target triple
- [ ] 23.160 Declare external runtime functions

### Phase 23.2: Runtime Library (12 tasks)

- [ ] 23.161 Create `runtime/dws_runtime.h` - C header for runtime API
- [ ] 23.162 Declare string operations: `dws_string_new()`, `dws_string_concat()`, `dws_string_len()`
- [ ] 23.163 Declare array operations: `dws_array_new()`, `dws_array_index()`, `dws_array_len()`
- [ ] 23.164 Declare memory management: `dws_alloc()`, `dws_free()`
- [ ] 23.165 Choose and document memory strategy (Boehm GC vs reference counting)
- [ ] 23.166 Declare object operations: `dws_object_new()`, virtual dispatch helpers
- [ ] 23.167 Declare exception handling: `dws_throw()`, `dws_catch()`
- [ ] 23.168 Declare RTTI: `dws_is_instance()`, `dws_as_instance()`
- [ ] 23.169 Create `runtime/dws_runtime.c` - implement runtime
- [ ] 23.170 Implement all runtime functions
- [ ] 23.171 Create `runtime/Makefile` to build `libdws_runtime.a`
- [ ] 23.172 Add runtime build to CI for Linux/macOS/Windows

### Phase 23.3: LLVM Code Emission (15 tasks)

- [ ] 23.173 Implement LLVM emitter: `Generate(module *mir.Module) (string, error)`
- [ ] 23.174 Emit function declarations with correct signatures
- [ ] 23.175 Emit basic blocks for each MIR block
- [ ] 23.176 Emit arithmetic instructions: `add`, `sub`, `mul`, `sdiv`, `srem`
- [ ] 23.177 Emit comparison instructions: `icmp eq`, `icmp slt`, etc.
- [ ] 23.178 Emit logical instructions: `and`, `or`, `xor`
- [ ] 23.179 Emit memory instructions: `alloca`, `load`, `store`
- [ ] 23.180 Emit call instructions: `call @function_name(args)`
- [ ] 23.181 Emit constants: integers, floats, strings
- [ ] 23.182 Emit control flow: conditional branches, phi nodes
- [ ] 23.183 Emit runtime calls for strings, arrays, objects
- [ ] 23.184 Implement type conversions: `sitofp`, `fptosi`
- [ ] 23.185 Emit struct types for classes and vtables
- [ ] 23.186 Implement virtual method dispatch
- [ ] 23.187 Implement exception handling (simple throw/catch or full LLVM EH)

### Phase 23.4: Linking and Testing (7 tasks)

- [ ] 23.188 Implement compilation pipeline: DWScript → MIR → LLVM IR → object → executable
- [ ] 23.189 Integrate `llc` to compile .ll → .o
- [ ] 23.190 Integrate linker to link object + runtime → executable
- [ ] 23.191 Add `compile-native` CLI command
- [ ] 23.192 Create 10+ end-to-end tests: DWScript → native → execute → verify
- [ ] 23.193 Benchmark JS vs native performance
- [ ] 23.194 Document LLVM backend in `docs/llvm-backend.md`

### Phase 23.5: Documentation (3 tasks)

- [ ] 23.195 Create `docs/codegen-architecture.md` - MIR overview, multi-backend design
- [ ] 23.196 Create `docs/mir-spec.md` - complete MIR reference with examples
- [ ] 23.197 Create `docs/js-backend.md` - DWScript → JavaScript mapping guide

---

## Phase 24: WebAssembly AOT Compilation [RECOMMENDED]

**Status**: Not started | **Priority**: MEDIUM-HIGH | **Estimated Tasks**: 5

### Overview

This phase extends WebAssembly support to generate standalone WASM binaries that can run without JavaScript dependency. This builds on Phase 14 (WASM Runtime & Playground) but focuses on ahead-of-time compilation for server-side and edge deployment.

**Architecture Flow**: DWScript Source → Go Compiler → WASM Binary → WASI Runtime (wasmtime, wasmer, wazero)

**Benefits**: Portable binaries, server-side execution, edge computing support, sandboxed execution

**Dependencies**: Builds on Phase 14 (WebAssembly Runtime & Playground)

### Phase 24.1: Standalone WASM Binaries (5 tasks)

- [ ] 24.1 Extend WASM compilation for standalone binaries
  - Generate WASM modules without JavaScript dependency
  - Use WASI for system calls
  - Support WASM-compatible runtime
  - Test with wasmtime, wasmer, wazero

- [ ] 24.2 Optimize WASM binary size
  - Use TinyGo compiler (smaller binaries)
  - Enable wasm-opt optimization
  - Strip unnecessary features
  - Measure binary size (target < 1MB)

- [ ] 24.3 Add WASM runtime support
  - Bundle WASM runtime (wasmer-go or wazero)
  - Create launcher executable
  - Support both JIT and AOT WASM execution
  - Test performance

- [ ] 24.4 Test WASM AOT compilation
  - Compile fixture tests to WASM
  - Run with different WASM runtimes
  - Measure performance vs native
  - Test browser and server execution

- [ ] 24.5 Document WASM AOT
  - Write `docs/wasm-aot.md`
  - Explain WASM compilation process
  - Provide deployment examples
  - Compare with Go native compilation

**Expected Results**: 5-20x speedup (browser), 10-30x speedup (WASI runtime)

---

## Phase 25: AST-Driven Formatter 🆕 **PLANNED**

**Status**: Not started | **Priority**: MEDIUM | **Estimated Tasks**: 30

### Overview

This phase delivers an auto-formatting pipeline that reuses the existing AST and semantic metadata to produce canonical DWScript source, accessible via the CLI (`dwscript fmt`), editors, and the web playground.

**Architecture Flow**: DWScript Source → Parser → AST → Formatter → Formatted DWScript Source

**Benefits**: Consistent code style, automated formatting, editor integration, playground support

### Phase 25.1: Specification & AST/Data Prep (7 tasks)

- [x] 25.1.1 Capture formatting requirements from upstream DWScript (indent width, begin/end alignment, keyword casing, line-wrapping) and document them in `docs/formatter-style-guide.md`.
- [x] 25.1.2 Audit current AST nodes for source position fidelity and comment/trivia preservation; list any nodes lacking `Pos` / `EndPos`.
- [x] 25.1.3 Extend the parser/AST to track leading and trailing trivia (single-line, block comments, blank lines) without disturbing semantic passes.
  - **Implemented**: Comment preservation infrastructure (lexer + AST structures)
  - **Lexer** (`internal/lexer/lexer.go`):
    - Added `preserveComments` flag to control comment tokenization
    - New methods: `SetPreserveComments()`, `readLineComment()`, `readBlockComment()`, `readCStyleComment()`
    - Modified `NextToken()` to return COMMENT tokens when enabled
    - Supports all 4 comment styles: `//`, `{ }`, `(* *)`, `/* */`
  - **AST** (`pkg/ast/comment.go`):
    - `Comment` type with text, position, and style
    - `CommentGroup` for grouping consecutive comments
    - `NodeComments` for leading/trailing comments per node
    - `CommentMap` for mapping nodes to comments (non-intrusive design)
    - Added `Comments CommentMap` field to `Program` struct
  - **Tests**:
    - `internal/lexer/comment_test.go` - 10 comprehensive tests for lexer
    - `pkg/ast/comment_test.go` - 8 tests for AST comment structures
  - **Documentation**: `docs/comment-preservation.md` - Complete guide
  - **Limitations**: Parser integration not yet complete (Phase 25.2.6)
    - ✅ Lexer can tokenize comments
    - ✅ Data structures defined
    - ❌ Parser doesn't attach comments to nodes (future work)
    - ❌ Printer can't output comments yet (future work)
- [x] 25.1.4 Define a `format.Options` struct (indent size, max line length, newline style) and default profile matching DWScript conventions.
  - **Implemented**: `pkg/printer/printer.go` defines comprehensive `Options` struct with:
    - Format types: DWScript, Tree, JSON
    - Style modes: Detailed, Compact, Multiline
    - Indentation control (width, spaces vs tabs)
    - Position and type info toggles
  - **Implemented**: `pkg/printer/styles.go` provides helper functions for common configurations
- [~] 25.1.5 Build a formatting test corpus in `testdata/formatter/{input,expected}` with tricky constructs (nested classes, generics, properties, preprocessor).
  - **Status**: Basic tests in `pkg/printer/printer_test.go` cover common cases
  - **TODO**: Create comprehensive test corpus with edge cases
- [x] 25.1.6 Add helper APIs to serialize AST back into token streams (e.g., `ast.FormatNode`, `ast.IterChildren`) to keep formatter logic decoupled from parser internals.
  - **Implemented**: `pkg/printer/printer.go` provides core printing infrastructure
  - **Implemented**: `pkg/printer/dwscript.go` contains node-specific formatting for all major AST types
- [~] 25.1.7 Ensure the semantic/type metadata needed for spacing decisions (e.g., `var` params, attributes) is exposed through lightweight inspector interfaces to avoid circular imports.
  - **Status**: Basic metadata support exists; AST nodes contain type annotations
  - **TODO**: May need additional helpers for complex spacing rules

### Phase 25.2: Formatter Engine Implementation (10 tasks)

- [ ] 25.2.1 Create `formatter` package with a multi-phase pipeline: AST normalization → layout planning → text emission.
- [ ] 25.2.2 Implement a visitor that emits `format.Node` instructions (indent/dedent, soft break, literal text) for statements and declarations, leveraging AST shape rather than raw tokens.
- [ ] 25.2.3 Handle block constructs (`begin...end`, class bodies, `case` arms) with indentation stacks so nested scopes auto-align.
- [ ] 25.2.4 Add expression formatting that respects operator precedence and inserts parentheses only when required; reuse existing precedence tables.
- [ ] 25.2.5 Support alignment for parameter lists, generics, array types, and property declarations with configurable wrap points.
- [ ] 25.2.6 Preserve user comments: attach leading comments before the owning node, keep inline comments after statements, and maintain blank-line intent (max consecutives configurable).
- [ ] 25.2.7 Implement whitespace normalization rules (single spaces around binary operators, before `do`/`then`, after commas, etc.).
- [ ] 25.2.8 Provide idempotency guarantees by building a golden test that pipes formatted output back through the formatter and asserts stability.
- [ ] 25.2.9 Expose a streaming writer that emits `[]byte`/`io.Writer` output to keep the CLI fast and low-memory.
- [ ] 25.2.10 Benchmark formatting of large fixtures (≥5k LOC) and optimize hot paths (string builder pools, avoiding interface allocations).

### Phase 25.3: Tooling & Playground Integration (7 tasks)

- [~] 25.3.1 Wire a new CLI command `dwscript fmt` (and `fmt -w`) that runs the formatter over files/directories, mirroring `gofmt` UX.
  - [x] 25.3.1.1 Create `cmd/dwscript/cmd/fmt.go` with basic command structure
  - [x] 25.3.1.2 Add `-w` flag to write formatted output back to files
  - [x] 25.3.1.3 Add `-l` flag to list files that would be reformatted
  - [x] 25.3.1.4 Support formatting from stdin when no file is provided
  - [x] 25.3.1.5 Add `-d` flag to show diff instead of rewriting files
  - [x] 25.3.1.6 Support formatting multiple files and directories recursively
  - [x] 25.3.1.7 Add style flags: `--style` (detailed/compact/multiline), `--indent` (width), `--tabs` (use tabs)
  - [x] 25.3.1.8 Add tests for the fmt command
    - **Implemented**: `cmd/dwscript/cmd/fmt_test.go` with comprehensive test coverage (540 lines)
    - Tests: formatSource, FormatBytes, isFormattedCorrectly, FormatFile, style options, indentation, complex constructs, error handling
    - Benchmarks: BenchmarkFormatSource (~15µs/op), BenchmarkFormatSourceCompact (~10µs/op)
    - **Known limitation**: Printer doesn't add trailing semicolons, affecting true idempotency (needs fix in pkg/printer for task 25.2.8)
  - [ ] 25.3.1.9 Update documentation and help text
- [ ] 25.3.2 Update the WASM bridge to expose a `Format(source string) (string, error)` hook exported from Go, reusing the same formatter package.
- [ ] 25.3.3 Modify `playground/js/playground.js` to call the WASM formatter before falling back to Monaco’s default action, enabling deterministic formatting in the browser.
- [ ] 25.3.4 Add formatter support to the VSCode extension / LSP stub (if present) so editors can trigger `textDocument/formatting`.
- [ ] 25.3.5 Ensure the formatter respects partial-range requests (`textDocument/rangeFormatting`) to avoid reformatting entire files when not desired.
- [ ] 25.3.6 Introduce CI checks (`just fmt-check`) that fail when files are not formatted, and document the workflow in `CONTRIBUTING.md`.
- [ ] 25.3.7 Provide sample scripts/snippets (e.g., Git hooks) encouraging contributors to run the formatter.

### Phase 25.4: Validation, UX, and Docs (6 tasks)

- [ ] 25.4.1 Create table-driven unit tests per node type plus integration tests that read `testdata/formatter` fixtures.
- [ ] 25.4.2 Add fuzz/property tests that compare formatter output against itself round-tripped through the parser → formatter pipeline.
- [ ] 25.4.3 Document formatter architecture and extension points in `docs/formatter-architecture.md`.
- [ ] 25.4.4 Update `PLAYGROUND.md`, `README.md`, and release notes to mention the Format button now runs the AST-driven formatter.
- [ ] 25.4.5 Record known limitations (e.g., preprocessor directives) and track follow-ups in `TEST_ISSUES.md`.
- [ ] 25.4.6 Gather usability feedback (issue template or telemetry) to prioritize refinements like configurable styles or multi-profile support.

---

## Phase 26: Future Enhancements & Experimental Features [LONG-TERM]

**Status**: Not started | **Priority**: LOW | **Tasks**: Variable

### Overview

This phase collects experimental, deferred, and long-term enhancement tasks that are not critical to the core DWScript implementation but may provide value in specific use cases or future development.

**Note**: Tasks in this phase are marked as DEFERRED or OPTIONAL and should only be pursued after core phases are complete and based on user demand.

### Phase 26.1: Plugin-Based JIT [SKIP - Poor Portability]

**Status**: SKIP RECOMMENDED | **Limitation**: No Windows support, requires Go toolchain at runtime

- [ ] 26.1 Implement Go code generation from AST
  - Create `internal/codegen/go_generator.go`
  - Generate Go source code from DWScript AST
  - Map DWScript types to Go types
  - Generate function declarations and calls
  - Handle closures and scoping

- [ ] 26.2 Implement plugin-based JIT
  - Use `go build -buildmode=plugin` to compile generated code
  - Load plugin with `plugin.Open()`
  - Look up compiled function with `plugin.Lookup()`
  - Call compiled function from interpreter
  - Cache plugins to disk

- [ ] 26.3 Add hot path detection for plugin JIT
  - Track function execution counts
  - Trigger plugin compilation for hot functions
  - Manage plugin lifecycle (loading/unloading)

- [ ] 26.4 Test plugin-based JIT
  - Run tests on Linux and macOS only
  - Compare performance with bytecode VM
  - Test plugin caching and reuse

- [ ] 26.5 Document plugin approach
  - Write `docs/plugin-jit.md`
  - Explain platform limitations
  - Provide usage examples

**Expected Results**: 3-5x faster than tree-walking
**Recommendation**: SKIP - poor portability, requires Go toolchain

### Phase 26.2: Alternative Compiler Targets [EXPERIMENTAL]

- [ ] 26.6 C code generation backend
  - Transpile DWScript to C
  - Leverage existing C compilers
  - Test on embedded systems

- [ ] 26.7 Rust code generation backend
  - Transpile DWScript to Rust
  - Leverage Rust's memory safety
  - Explore performance characteristics

- [ ] 26.8 Python code generation backend
  - Transpile DWScript to Python
  - Enable rapid prototyping
  - Integration with Python ecosystem

### Phase 26.3: Advanced Optimization Research [EXPERIMENTAL]

- [ ] 26.9 Profile-guided optimization (PGO)
  - Collect runtime profiles
  - Use profiles to guide optimizations
  - Measure performance improvements

- [ ] 26.10 Speculative optimization
  - Type speculation based on runtime behavior
  - Deoptimization on type changes
  - Guard conditions

- [ ] 26.11 Escape analysis
  - Determine when objects can be stack-allocated
  - Reduce GC pressure
  - Improve performance

- [ ] 26.12 Inline caching for dynamic dispatch
  - Cache method lookup results
  - Invalidate on class changes
  - Measure performance impact

### Phase 26.4: Language Extensions [EXPERIMENTAL]

- [ ] 26.13 Async/await support
  - Design async/await syntax for DWScript
  - Implement coroutine-based execution
  - Test with concurrent workloads

- [ ] 26.14 Pattern matching
  - Extend case statements with pattern matching
  - Support destructuring
  - Type narrowing based on patterns

- [ ] 26.15 Macros/metaprogramming
  - Design macro system
  - Compile-time code generation
  - Template metaprogramming support

### Phase 26.5: Tooling Enhancements [LOW PRIORITY]

- [ ] 26.16 IDE integration beyond LSP
  - IntelliJ IDEA plugin
  - VS Code enhanced extension
  - Sublime Text package

- [ ] 26.17 Package manager
  - Design package format
  - Implement dependency resolution
  - Create package registry

- [ ] 26.18 Build system integration
  - Make/CMake integration
  - Bazel rules
  - Gradle plugin

---

## Summary

This roadmap now spans **~1000+ bite-sized tasks** across **21 phases**, organized into three tiers: **Core Language** (Phases 1-10), **Execution & Compilation** (Phases 11-18), and **Ecosystem & Tooling** (Phases 19-21).

### Core Language Implementation (Phases 1-10) ✅ MOSTLY COMPLETE

1. **Phase 1 – Lexer**: ✅ Complete (150+ tokens, 97% coverage)
2. **Phase 2 – Parser & AST**: ✅ Complete (Pratt parser, comprehensive AST)
3. **Phase 3 – Statement execution**: ✅ Complete (98.5% coverage)
4. **Phase 4 – Control flow**: ✅ Complete (if/while/for/case)
5. **Phase 5 – Functions & scope**: ✅ Complete (91.3% coverage)
6. **Phase 6 – Type checking**: ✅ Complete (semantic analysis, 88.5% coverage)
7. **Phase 7 – Object-oriented features**: ✅ Complete (classes, interfaces, inheritance)
8. **Phase 8 – Extended language features**: ✅ Complete (operators, properties, enums, arrays, exceptions)
9. **Phase 9 – Feature parity completion**: 🔄 In progress (class methods, constants, type casting)
10. **Phase 10 – API enhancements for LSP**: ✅ Complete (public AST, structured errors, visitors)

### Execution & Compilation (Phases 11-18)

11. **Phase 11 – Bytecode Compiler & VM**: ✅ MOSTLY COMPLETE (5-6x faster than AST interpreter, 116 opcodes)
12. **Phase 12 – Performance & Polish**: 🔄 Partial (profiling done, optimizations pending)
13. **Phase 13 – Go AOT Compilation**: [RECOMMENDED] Transpile to Go, native binaries (10-50x speedup)
14. **Phase 14 – WebAssembly Runtime & Playground**: ✅ MOSTLY COMPLETE (WASM build, playground, NPM package)
15. **Phase 15 – MIR Foundation**: [DEFERRED] Mid-level IR for multi-backend support
16. **Phase 16 – JavaScript Backend**: [DEFERRED] MIR → JavaScript code generation
17. **Phase 17 – LLVM Backend**: [DEFERRED/LOW PRIORITY] Maximum performance, high complexity
18. **Phase 18 – WebAssembly AOT**: [RECOMMENDED] Standalone WASM binaries for edge/server deployment

### Ecosystem & Tooling (Phases 19-21)

19. **Phase 19 – AST-Driven Formatter**: [PLANNED] Auto-formatting for CLI, editors, playground
20. **Phase 20 – Community & Ecosystem**: [ONGOING] REPL, debugging, security, maintenance
21. **Phase 26 – Future Enhancements**: [LONG-TERM] Experimental features, alternative targets

### Implementation Priorities

**HIGH PRIORITY (Core Functionality)**:
- Phase 9 (feature parity completion)
- Phase 12 (performance polish)
- Phase 13 (Go AOT compilation)
- Phase 14 remaining tasks (WASM testing)
- Phase 20 (community building, REPL, debugging)

**MEDIUM PRIORITY (Value-Add Features)**:
- Phase 18 (WASM AOT)
- Phase 19 (formatter)
- Phase 15-16 (MIR + JavaScript backend)

**LOW PRIORITY (Deferred/Optional)**:
- Phase 17 (LLVM backend - complex, high maintenance)
- Phase 21 (experimental features)

### Architecture Summary

**Execution Modes** (in order of priority):
1. **AST Interpreter** (baseline, complete) - Simple, portable
2. **Bytecode VM** (5-6x faster, mostly complete) - Good performance, low complexity
3. **Go AOT** (10-50x faster, recommended) - Excellent performance, great portability
4. **WASM Runtime** (browser/edge, mostly complete) - Web execution, good performance
5. **WASM AOT** (server/edge, recommended) - Portable binaries, sandboxed execution
6. **JavaScript Backend** (deferred) - Browser execution via transpilation
7. **LLVM Backend** (deferred) - Maximum performance, very complex

**Code Generation Flow**:
```
DWScript Source → Parser → AST → Semantic Analyzer
                                      ├→ AST Interpreter (baseline)
                                      ├→ Bytecode Compiler → VM (5-6x)
                                      ├→ Go Transpiler → Native (10-50x)
                                      ├→ WASM Compiler → Browser/WASI
                                      └→ MIR Builder → JS/LLVM Emitter (future)
```

### Project Timeline

The project can realistically take **1-3 years** depending on:

- Development pace (full-time vs part-time)
- Team size (solo vs multiple contributors)
- Completeness goals (minimal viable vs full feature parity)
- [ ] 9.16.6 Refactor type-specific nodes (ArrayLiteralExpression, CallExpression, NewExpression, MemberAccessExpression, etc.)
  - [ ] Embed appropriate base struct into array literals/index expressions
  - [ ] Embed base structs into call/member/new expressions in `pkg/ast/classes.go`/`pkg/ast/functions.go`
  - [ ] Update parser/interpreter/semantic code paths for these nodes
  - [ ] Remove duplicate helpers once embedded
