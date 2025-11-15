<!-- trunk-ignore-all(prettier) -->
# DWScript to Go Port - Detailed Implementation Plan

This document breaks down the ambitious goal of porting DWScript from Delphi to Go into bite-sized, actionable tasks organized by stages. Each stage builds incrementally toward a fully functional DWScript compiler/interpreter in Go.

---

## Phase 1-5: Core Language Implementation (Stages 1-5)

**Status**: 5/5 stages complete (100%) | **Coverage**: Parser 84.5%, Interpreter 83.3%

### Stage 1: Lexer (Tokenization) ✅ **COMPLETED**

- Implemented complete DWScript lexer with 150+ tokens including keywords, operators, literals, and delimiters
- Support for case-insensitive keywords, hex/binary literals, string escape sequences, and all comment types
- Comprehensive test suite with 97.1% coverage and position tracking for error reporting

### Stage 2: Basic Parser and AST (Expressions Only) ✅ **COMPLETED**

- Pratt parser implementation with precedence climbing supporting all DWScript operators
- Complete AST node hierarchy with visitor pattern support
- Expression parsing for literals, identifiers, binary/unary operations, grouped expressions, and function calls
- Full operator precedence handling and error recovery mechanisms

### Stage 3: Statement Execution (Sequential Execution) ✅ **COMPLETED** (98.5%)

- Variable declarations with optional type annotations and initialization
- Assignment statements with DWScript's `:=` operator
- Block statements with `begin...end` syntax
- Built-in functions (PrintLn, Print) and user-defined function calls
- Environment/symbol table with nested scope support
- Runtime value system supporting Integer, Float, String, Boolean, and Nil types
- Sequential statement execution with proper error handling

### Stage 4: Control Flow - Conditions and Loops ✅ **COMPLETED**

- If-then-else statements with proper boolean evaluation
- While loops with condition testing before execution
- Repeat-until loops with condition testing after execution
- For loops supporting both `to` and `downto` directions with integer bounds
- Case statements with value matching and optional else branches
- Full integration with existing type system and error reporting

### Stage 5: Functions, Procedures, and Scope Management ✅ **COMPLETED** (91.3%)

- Function and procedure declarations with parameter lists and return types
- By-reference parameters (`var` keyword) - parsing implemented, runtime partially complete
- Function calls with argument passing and return value handling
- Lexical scoping with proper environment nesting
- Built-in functions for output and basic operations
- Recursive function support with environment cleanup
- Symbol table integration for function resolution

---

## Stage 6: Static Type Checking and Semantic Analysis ✅ **COMPLETED**

- Built the reusable type system in `types/` (primitive, function, aggregate types plus coercion rules); see docs/stage6-type-system-summary.md for the full compatibility matrix.
- Added optional type annotations to AST nodes and parser support for variables, parameters, and return types so semantic analysis has complete metadata.
- Implemented the semantic analyzer visitor that resolves identifiers, validates declarations/assignments/expressions, enforces control-flow rules, and reports multiple errors per pass with 88.5% coverage.
- Hooked the analyzer into the parser/interpreter/CLI (with a disable flag) so type errors surface before execution and runtime uses inferred types.
- Upgraded diagnostics with per-node position data, the `errors/` formatter, and curated fixtures in `testdata/type_errors` plus `testdata/type_valid`, alongside CLI integration suites.

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

## Task 9.1: Fix Interface Reference Test Failures

**Goal**: Fix failing interface reference tests to achieve 97% pass rate (32/33 tests passing).

**Estimate**: 7-13 hours

**Status**: NOT STARTED

**Current Status**: 16 passing, 16 failing, 1 skipped (out of 33 total) - 48% pass rate

**Target**: 32 passing, 1 deferred - 97% pass rate

**Test File**: `internal/interp/interface_reference_test.go`

**Description**: The interface reference tests reveal 6 categories of missing or broken functionality in the interpreter's interface implementation. These failures prevent proper interface member access, lifetime management, type checking, and casting operations.

**Subtasks** (ordered by impact and difficulty):

### 9.1.1 Interface Member Access [CRITICAL]

**Goal**: Fix interface member/method access to allow calling methods on interface instances.

**Status**: NOT STARTED

**Impact**: Fixes 6 tests immediately (37.5% of failures)

**Estimate**: 1-2 hours

**Tests Fixed**:

- `call_interface_method`
- `interface_multiple_cast`
- `interface_properties`
- `intf_casts`
- `intf_spread`
- `intf_spread_virtual`

**Error**: `ERROR: cannot access member 'X' of type 'INTERFACE' (no helper found)`

**Root Cause**:

- `evalMemberAccess` in `internal/interp/objects_hierarchy.go:282-320` doesn't handle InterfaceInstance
- `AsObject(objVal)` returns false for InterfaceInstance (only recognizes ObjectInstance)
- Fallback path checks for helpers, but interfaces don't have helpers

**Implementation**:

- [ ] Add InterfaceInstance detection in `evalMemberAccess` before `AsObject` check
- [ ] Extract underlying ObjectInstance via `ii.Object`
- [ ] Verify method exists in interface definition
- [ ] Delegate method calls to underlying object
- [ ] Update `evalMethodCall` in `objects_methods.go` to handle InterfaceInstance

**Example Test**:

```pascal
intfRef := implem as IMyInterface;
intfRef.A;  // Should call A on the underlying object
```

**Files to Modify**:

- `internal/interp/objects_hierarchy.go`
- `internal/interp/objects_methods.go`

### 9.1.2 'implements' Operator with Class Types [MEDIUM]

**Goal**: Allow 'implements' operator to work with class types, not just object instances.

**Status**: NOT STARTED

**Impact**: Fixes 2 tests

**Estimate**: 30 minutes

**Tests Fixed**:

- `interface_implements_intf`
- `interface_inheritance_instance` (partial - adds missing error message)

**Error**: `ERROR: 'implements' operator requires object instance, got CLASS`

**Root Cause**:

- `evalImplementsExpression` in `internal/interp/expressions_complex.go:227-230` only accepts ObjectInstance
- Test uses class identifiers: `if TMyImplementation implements IMyInterface then`

**Implementation**:

- [ ] Update `evalImplementsExpression` to handle ClassInfoValue and ClassValue
- [ ] Use existing `classImplementsInterface` helper for class type checks
- [ ] Keep existing ObjectInstance logic (extract class from instance)

**Example Test**:

```pascal
if TMyImplementation implements IMyInterface then  // TMyImplementation is a CLASS
   PrintLn('Ok');
```

**Files to Modify**:

- `internal/interp/expressions_complex.go`

### 9.1.3 Interface-to-Object Casting [MEDIUM]

**Goal**: Validate interface-to-object casts and throw proper exceptions for invalid casts.

**Status**: NOT STARTED

**Impact**: Fixes 1 test

**Estimate**: 1 hour

**Tests Fixed**:
- `interface_cast_to_obj`

**Error**: Missing exception when invalid cast attempted

**Expected Behavior**:
```pascal
var d := IntfRef as TDummyClass;  // Should throw exception if incompatible
```

**Expected Output**: `Cannot cast interface of "TMyImplementation" to class "TDummyClass"`

**Actual**: Cast succeeds incorrectly (no validation)

**Root Cause**:
- `evalAsExpression` in `internal/interp/expressions_complex.go:152-156` only handles object-to-interface casting
- Doesn't handle InterfaceInstance as left operand

**Implementation**:
- [ ] Detect when left side is InterfaceInstance in `evalAsExpression`
- [ ] Extract underlying object from interface
- [ ] Verify object's class is compatible with target class
- [ ] Return underlying object if compatible
- [ ] Throw exception with proper message format if incompatible

**Files to Modify**:
- `internal/interp/expressions_complex.go`

### 9.1.4 Interface Fields in Records [LOW]

**Goal**: Support interface-typed fields in record types.

**Status**: NOT STARTED

**Impact**: Fixes 1 test

**Estimate**: 1-2 hours

**Tests Fixed**:
- `intf_in_record`

**Error**: `ERROR: unknown or invalid type for field 'FIntf' in record 'TRec'`

**Root Cause**:
- `resolveTypeFromExpression` in `internal/interp/record.go:34` doesn't recognize interface types

**Implementation**:
- [ ] Update type resolution to recognize interface type annotations
- [ ] Add interface type to supported field types in records
- [ ] Initialize interface fields to nil by default
- [ ] Handle interface assignment in record field access
- [ ] Handle interface method calls through record fields

**Example Test**:
```pascal
type TRec = record
  FIntf: IMyInterface;  // Interface-typed field
end;
```

**Files to Modify**:
- `internal/interp/record.go`

### 9.1.5 Interface Lifetime Management [HIGH PRIORITY - COMPLEX]

**Goal**: Implement reference counting and automatic destructor calls for interface-held objects.

**Status**: NOT STARTED

**Impact**: Fixes 5 tests (critical for memory management)

**Estimate**: 4-8 hours

**Tests Fixed**:
- `interface_lifetime`
- `interface_lifetime_scope`
- `interface_lifetime_scope_ex1`
- `interface_lifetime_scope_ex2`
- `interface_lifetime_simple`

**Error**: Missing "Destroy" output when interface goes out of scope

**Expected Behavior**:
```pascal
IntfRef := TMyImplementation.Create;
IntfRef.A;
IntfRef := nil;  // -> Should call Destroy on underlying object
PrintLn('end');
```

**Expected Output**: `A\nDestroy\nend`
**Actual Output**: `A\nend`

**Root Cause**:
- No reference counting for objects held by interfaces
- No destructor call when interface set to nil or goes out of scope
- No cleanup mechanism when last interface reference is released

**Implementation**:
- [ ] Add reference count field to ObjectInstance in `value.go`
- [ ] Increment ref count when interface wraps object
- [ ] Decrement ref count when interface assigned nil or goes out of scope
- [ ] Call destructor when ref count reaches zero
- [ ] Update InterfaceInstance assignment in `interface.go`
- [ ] Add scope-based cleanup in environment variable assignment
- [ ] Handle interface-to-interface assignment (transfer ownership)

**Complexity Notes**:
- Must avoid memory leaks (objects not freed)
- Must avoid premature cleanup (object freed while still referenced)
- Must handle circular references gracefully
- Need to track all interface assignments and scope exits

**Files to Modify**:
- `internal/interp/value.go` (add ref count to ObjectInstance)
- `internal/interp/interface.go` (ref counting in InterfaceInstance)
- `internal/interp/environment.go` (scope cleanup)
- `internal/interp/objects_hierarchy.go` (assignment handling)

### 9.1.6 Method Delegate Assignment [LOW - DEFERRED]

**Goal**: Support extracting bound method references from interface instances.

**Status**: DEFERRED

**Impact**: Fixes 1 test

**Estimate**: 2-4 hours

**Tests Fixed**:
- `intf_delegate`

**Error**: `ERROR: undefined function: h at line 18, column 3`

**Root Cause**:
- Can't assign interface method to procedure variable
- No support for bound method pointers

**Example Test**:
```pascal
var h : procedure := i.Hello;  // Extract method reference from interface
h();  // Call it
```

**Implementation** (when implemented):
- [ ] Support extracting method references from InterfaceInstance
- [ ] Create bound method pointer values (capture interface + method)
- [ ] Handle invocation of bound methods

**Deferred Because**:
- Complex feature requiring new value types
- Low test impact (1 test)
- Can be implemented later without blocking other features

**Files to Modify** (future):
- `internal/interp/expressions_complex.go`
- `internal/interp/value.go`

---

**Success Criteria**:
- After subtasks 1-4: 27/33 tests passing (82% pass rate)
- After subtask 5: 32/33 tests passing (97% pass rate)
- All changes maintain backward compatibility with existing tests
- Interface behavior matches original DWScript semantics

**Testing**:
```bash
# Run interface reference tests
go test -v ./internal/interp -run TestInterfaceReferenceTests

# Run full test suite to verify no regressions
just test
```

---

## Task 9.2.0: Static Array Type Compatibility and Var Parameters

**Goal**: Fix type compatibility for static arrays in var parameters, enabling the quicksort.pas test to pass.

**Status**: READY - Expanded from deferred task with comprehensive breakdown

**Estimate**: 12-16 hours (1.5-2 days)

**Blocked Tests** (1 test):

- `testdata/fixtures/Algorithms/quicksort.pas`

**Current Error**: `cannot assign TData to TData` when passing static array to var parameter

**Root Cause Analysis**:

The error message "cannot assign TData to TData" is misleading. Based on investigation:

1. **Type Alias Resolution**: The type system may not be properly resolving type aliases when checking compatibility
2. **Var Parameter Checking**: `internal/semantic/analyze_function_calls.go:78-97` uses `canAssign()` which calls `types.IsCompatible()`
3. **Array Compatibility Rules**: `internal/types/compatibility.go:45-60` has strict rules requiring exact type matches
4. **The Issue**: When TData (a type alias for `array [0..size-1] of integer`) is used in var parameters, the type comparison may be comparing type instances rather than type definitions

**Key Implementation Locations**:

- Array type definition: `internal/types/compound_types.go:18-110`
- Type compatibility: `internal/types/compatibility.go:45-60`
- Type aliases: `internal/semantic/analyzer.go` (type declaration handling)
- Var parameter checking: `internal/semantic/analyze_function_calls.go:78-97`
- Array runtime: `internal/interp/array.go:1-744`

**Architecture Strategy**:

This task will follow a **TDD approach**:
1. Research original DWScript behavior
2. Write failing tests that document expected behavior
3. Implement fixes in type system and semantic analyzer
4. Verify interpreter and bytecode VM handle arrays correctly
5. Ensure all tests pass

**Success Criteria**:

- `testdata/fixtures/Algorithms/quicksort.pas` passes
- Type aliases for arrays work correctly in var parameters
- No regressions in existing array tests
- Comprehensive test coverage for array type compatibility

---

### Task 9.2.1: Research and Document DWScript Array Behavior

**Goal**: Understand how original DWScript handles static arrays and type aliases in var parameters.

**Status**: DONE ✅

**Estimate**: 2 hours

**Actions**:

- [x] Read DWScript documentation on array types and type aliases
- [x] Examine `reference/dwscript-original/` source code for array type compatibility logic
- [x] Search for how type aliases are resolved in parameter passing
- [x] Test simple DWScript programs with static arrays and var parameters (if possible)
- [x] Document findings in `docs/arrays-type-compatibility-research.md`

**Key Questions to Answer**:

1. Are type aliases transparent (i.e., `TData` is exactly the same as `array [0..99] of integer`)?
2. Should var parameters accept both the alias and the underlying type?
3. How does DWScript handle array type equality vs compatibility?
4. Are there any special rules for array slicing or bounds in var parameters?

**Files to Create**:

- `docs/arrays-type-compatibility-research.md` - Research findings

---

### Task 9.2.2: Reproduce and Analyze the Error

**Goal**: Create minimal reproducible test case and trace the exact error path.

**Status**: NOT STARTED

**Estimate**: 1-2 hours

**Actions**:

- [ ] Create minimal test case in `testdata/array_alias_var_param.dws`
- [ ] Run test with verbose error output
- [ ] Add debug logging to trace type comparison in semantic analyzer
- [ ] Identify exact line where type check fails
- [ ] Document the call stack and variable values at failure point

**Minimal Test Case**:

```pascal
type TIntArray = array [0..9] of integer;

procedure TestProc(var arr: TIntArray);
begin
  arr[0] := 42;
end;

var myArray: TIntArray;
TestProc(myArray);  // Should work - type alias is transparent in DWScript
```

**Files to Create**:

- `testdata/array_alias_var_param.dws` - Minimal test case

---

### Task 9.2.3: Identify Required Changes

**Goal**: Map out all code locations that need modification.

**Status**: NOT STARTED

**Estimate**: 1 hour

**Actions**:

- [ ] List all functions involved in type compatibility checking
- [ ] Identify where type aliases are stored and resolved
- [ ] Determine if lexer/parser changes are needed (likely not)
- [ ] Map out semantic analyzer changes needed
- [ ] Check if interpreter needs modifications
- [ ] Check if bytecode VM needs modifications
- [ ] Create implementation checklist

**Expected Findings**:

- Lexer: No changes (tokens are correct)
- Parser: No changes (AST structure is correct)
- Type System: Need to improve type alias resolution and equality checking
- Semantic Analyzer: Need to fix var parameter type checking
- Interpreter: Possibly need to ensure array references work correctly
- Bytecode: Possibly need to ensure array operations work correctly

---

### Task 9.2.4: Write Failing Unit Tests (TDD - Red Phase)

**Goal**: Create comprehensive tests that document expected behavior before implementation.

**Status**: NOT STARTED

**Estimate**: 2-3 hours

**Actions**:

- [ ] Create `internal/types/array_alias_test.go` for type system tests
- [ ] Create `internal/semantic/array_var_param_test.go` for semantic tests
- [ ] Write tests for type alias equality (should TData == TData?)
- [ ] Write tests for type alias vs underlying type (should TData == array[0..9] of int?)
- [ ] Write tests for var parameter compatibility with aliases
- [ ] Write tests for nested type aliases
- [ ] All tests should currently FAIL

**Test Cases to Write**:

```go
// internal/types/array_alias_test.go
TestTypeAliasEquality                    // TData vs TData
TestTypeAliasUnderlyingTypeEquality     // TData vs array[0..9] of int
TestDifferentAliasesSameUnderlying      // TData vs TOtherData (both array[0..9])
TestStaticArrayBoundsEquality           // array[0..9] vs array[0..9]
TestStaticArrayBoundsDifferent          // array[0..9] vs array[0..10]

// internal/semantic/array_var_param_test.go
TestVarParamWithTypeAlias               // var param accepts same alias
TestVarParamWithUnderlyingType          // var param accepts underlying type
TestVarParamRejectsIncompatible         // var param rejects different type
TestVarParamWithNestedAlias             // type A = B; type B = array...
```

**Files to Create**:

- `internal/types/array_alias_test.go` - Type system unit tests
- `internal/semantic/array_var_param_test.go` - Semantic analysis tests

---

### Task 9.2.5: Write Failing Integration Tests (TDD - Red Phase)

**Goal**: Create end-to-end tests for array operations with type aliases.

**Status**: NOT STARTED

**Estimate**: 1-2 hours

**Actions**:

- [ ] Create `testdata/test_cases/arrays/type_aliases.dws`
- [ ] Create `testdata/test_cases/arrays/var_parameters.dws`
- [ ] Write tests for procedures with var array parameters
- [ ] Write tests for functions returning arrays through var parameters
- [ ] Write tests for nested procedure calls with var arrays
- [ ] Add expected output files
- [ ] Verify all tests currently FAIL

**Test Scenarios**:

```pascal
// testdata/test_cases/arrays/type_aliases.dws
type TIntArray = array [0..4] of integer;

procedure Fill(var arr: TIntArray);
var i: integer;
begin
  for i := 0 to 4 do
    arr[i] := i * 10;
end;

var data: TIntArray;
Fill(data);
PrintLn(IntToStr(data[2]));  // Should print: 20
```

**Files to Create**:

- `testdata/test_cases/arrays/type_aliases.dws`
- `testdata/test_cases/arrays/var_parameters.dws`
- `testdata/test_cases/arrays/type_aliases.expected`
- `testdata/test_cases/arrays/var_parameters.expected`

---

### Task 9.2.6: Fix Type Alias Resolution in Type System

**Goal**: Ensure type aliases are properly resolved when checking type equality/compatibility.

**Status**: NOT STARTED

**Estimate**: 2-3 hours

**Actions**:

- [ ] Verify `GetUnderlyingType()` exists in `internal/types/types.go:250`
- [ ] Update `ArrayType.Equals()` to use `GetUnderlyingType()` for type alias resolution
- [ ] Update `IsCompatible()` in `internal/types/compatibility.go` to use `GetUnderlyingType()`
- [ ] Ensure type aliases are transparent in comparisons
- [ ] Run unit tests from 9.2.4 - some should now pass

**Implementation Strategy**:

```go
// Use existing GetUnderlyingType() from internal/types/types.go:250
// This function already handles cycle detection via iterative approach:
//
// func GetUnderlyingType(t Type) Type {
//     for alias, ok := t.(*TypeAlias); ok; alias, ok = t.(*TypeAlias) {
//         t = alias.AliasedType
//     }
//     return t
// }

// Update ArrayType.Equals() in compound_types.go
func (a *ArrayType) Equals(other Type) bool {
    // Resolve any type aliases first using existing function
    resolvedA := GetUnderlyingType(a)
    resolvedOther := GetUnderlyingType(other)

    // Now compare resolved types
    // ... existing comparison logic ...
}
```

**Files to Modify**:

- `internal/types/compound_types.go:39-72` - Update `ArrayType.Equals()` to use `GetUnderlyingType()`
- `internal/types/compatibility.go:45-60` - Update array compatibility rules to use `GetUnderlyingType()`
- NOTE: `GetUnderlyingType()` already exists in `internal/types/types.go:250` - no need to create new function

**Tests to Pass**:

- `TestTypeAliasEquality`
- `TestTypeAliasUnderlyingTypeEquality`

---

### Task 9.2.7: Fix Type Alias Handling in Semantic Analyzer

**Goal**: Update semantic analyzer to store and resolve type aliases correctly.

**Status**: NOT STARTED

**Estimate**: 2-3 hours

**Actions**:

- [ ] Review how type declarations create type aliases
- [ ] Ensure type aliases are stored in symbol table correctly
- [ ] Update type lookup to return alias metadata
- [ ] Verify type alias chain resolution works
- [ ] Add tests for nested type aliases
- [ ] Run unit tests - more should pass

**Key Locations**:

- `internal/semantic/analyzer.go` - Type declaration handling
- `internal/semantic/symbol_table.go` - Type storage and lookup
- `internal/semantic/analyze_function_calls.go:78-97` - Parameter checking

**Implementation Notes**:

The semantic analyzer needs to distinguish between:
1. **Type Definition**: `type TData = array [0..9] of integer;` creates a new named type
2. **Type Alias**: In DWScript, type definitions are typically aliases (transparent)
3. **Type Identity**: When comparing types, aliases should be resolved

**Files to Modify**:

- `internal/semantic/analyzer.go` - Type declaration analysis
- `internal/semantic/symbol_table.go` - Type storage (if needed)

---

### Task 9.2.8: Fix Var Parameter Type Checking

**Goal**: Update var parameter validation to accept type aliases.

**Status**: NOT STARTED

**Estimate**: 1-2 hours

**Actions**:

- [ ] Update `analyzeCallExpression()` in `analyze_function_calls.go`
- [ ] Fix parameter type checking to resolve aliases before comparison
- [ ] Add special handling for var parameters (require exact match after alias resolution)
- [ ] Ensure error messages are clear when types don't match
- [ ] Run semantic tests - should pass
- [ ] Run integration tests - some should pass

**Current Code** (`internal/semantic/analyze_function_calls.go:78-97`):

```go
// This section needs to resolve type aliases before checking compatibility
if param.IsVar {
    // Var parameters require exact type match
    // Need to resolve both paramType and argType before comparing
}
```

**Files to Modify**:

- `internal/semantic/analyze_function_calls.go:78-97` - Update var parameter checking

**Tests to Pass**:

- `TestVarParamWithTypeAlias`
- `TestVarParamWithUnderlyingType`

---

### Task 9.2.9: Verify Interpreter Array Handling

**Goal**: Ensure interpreter correctly handles array variables and var parameters.

**Status**: NOT STARTED

**Estimate**: 1-2 hours

**Actions**:

- [ ] Review `internal/interp/array.go` for var parameter handling
- [ ] Verify array values are passed by reference for var parameters
- [ ] Test that modifications in procedures affect original array
- [ ] Add debug logging if needed
- [ ] Run integration tests with interpreter

**Key Checks**:

- [ ] Array values preserve type information when passed to functions
- [ ] Var parameters create references, not copies
- [ ] Array element assignment works through var parameters
- [ ] Array bounds are preserved and checked

**Files to Review**:

- `internal/interp/array.go:1-744` - Array operations
- `internal/interp/value.go:796-824` - ArrayValue representation
- `internal/interp/functions.go` - Function call handling (if exists)

---

### Task 9.2.10: Verify Bytecode VM Array Handling

**Goal**: Ensure bytecode VM correctly handles array var parameters.

**Status**: NOT STARTED

**Estimate**: 1-2 hours

**Actions**:

- [ ] Review bytecode array operations in `internal/bytecode/vm.go`
- [ ] Verify compiler generates correct bytecode for var parameters
- [ ] Test array reference passing in bytecode
- [ ] Run integration tests with `--bytecode` flag
- [ ] Compare bytecode output with interpreter output

**Key Checks**:

- [ ] Bytecode compiler preserves type information for arrays
- [ ] VM handles array references for var parameters
- [ ] Array operations work identically to interpreter
- [ ] No performance regressions

**Files to Review**:

- `internal/bytecode/compiler.go` - Array compilation
- `internal/bytecode/vm.go` - Array instruction execution
- `internal/bytecode/instruction.go:116` - Array opcodes

---

### Task 9.2.11: Add Comprehensive Array Type Tests

**Goal**: Expand test coverage for all array type scenarios.

**Status**: NOT STARTED

**Estimate**: 2 hours

**Actions**:

- [ ] Add tests for multidimensional arrays with aliases
- [ ] Add tests for array of records with aliases
- [ ] Add tests for dynamic arrays vs static arrays
- [ ] Add tests for const parameters with array aliases
- [ ] Add tests for out parameters with array aliases
- [ ] Add edge cases (empty arrays, single-element arrays)
- [ ] Verify all tests pass

**Test Categories**:

1. **Type Alias Chains**: `type A = B; type B = C; type C = array[0..9] of int;`
2. **Multiple Aliases**: Different names for same underlying type
3. **Parameter Modes**: var, const, out, value parameters
4. **Array Bounds**: Different bound combinations
5. **Element Types**: Arrays of primitives, records, classes, etc.

**Files to Create/Update**:

- `internal/types/array_alias_test.go` - Add edge cases
- `internal/semantic/array_var_param_test.go` - Add comprehensive scenarios
- `testdata/test_cases/arrays/` - Add more test scripts

---

### Task 9.2.12: Fix quicksort.pas Test

**Goal**: Verify that the original failing test now passes.

**Status**: NOT STARTED

**Estimate**: 30 minutes

**Actions**:

- [ ] Run `testdata/fixtures/Algorithms/quicksort.pas` with interpreter
- [ ] Run `testdata/fixtures/Algorithms/quicksort.pas` with bytecode VM
- [ ] Verify output matches expected results
- [ ] Check that swaps count is correct
- [ ] Verify array is sorted correctly
- [ ] Document any remaining issues

**Test Command**:

```bash
# Test with AST interpreter
./bin/dwscript run testdata/fixtures/Algorithms/quicksort.pas

# Test with bytecode VM
./bin/dwscript run --bytecode testdata/fixtures/Algorithms/quicksort.pas

# Run as part of fixture test suite
go test -v ./internal/interp -run TestDWScriptFixtures/Algorithms
```

**Expected Output**:

```
Swaps: >=100
Data:
0
1
2
...
99
```

---

### Task 9.2.13: Run Full Test Suite and Fix Regressions

**Goal**: Ensure no existing tests were broken by the changes.

**Status**: NOT STARTED

**Estimate**: 1-2 hours

**Actions**:

- [ ] Run full unit test suite: `go test ./...`
- [ ] Run full integration test suite with interpreter
- [ ] Run full integration test suite with bytecode
- [ ] Identify any regressions
- [ ] Fix any broken tests
- [ ] Update test expectations if behavior changed correctly
- [ ] Verify test coverage didn't decrease

**Test Commands**:

```bash
# Full test suite
just test

# With coverage
just test-coverage

# Specific test categories
go test ./internal/types/...
go test ./internal/semantic/...
go test ./internal/interp/...
go test ./internal/bytecode/...
```

---

### Task 9.2.14: Update Documentation

**Goal**: Document the array type compatibility rules and implementation.

**Status**: NOT STARTED

**Estimate**: 1 hour

**Actions**:

- [ ] Update CLAUDE.md with array type alias behavior
- [ ] Create/update `docs/type-system.md` with compatibility rules
- [ ] Add examples of array type aliases to documentation
- [ ] Document var parameter behavior with arrays
- [ ] Update PLAN.md to mark task 9.2 as complete

**Documentation Sections to Add/Update**:

1. **Type System Rules**: How type aliases work
2. **Array Types**: Static vs dynamic, bounds, element types
3. **Parameter Passing**: var, const, out, value semantics
4. **Compatibility**: When types are considered compatible

**Files to Update**:

- `CLAUDE.md` - Add array type alias section
- `docs/type-system.md` - Create or update with comprehensive rules
- `PLAN.md` - Mark task complete

---

### Task 9.2.15: Performance Testing and Optimization

**Goal**: Ensure type checking performance hasn't degraded.

**Status**: NOT STARTED

**Estimate**: 1 hour

**Actions**:

- [ ] Benchmark type compatibility checking before/after changes
- [ ] Profile semantic analysis phase
- [ ] Identify any performance hotspots
- [ ] Optimize type alias resolution if needed (caching?)
- [ ] Document performance characteristics

**Benchmark Tests to Create**:

```go
// internal/types/benchmark_test.go
BenchmarkTypeAliasResolution
BenchmarkArrayTypeEquality
BenchmarkTypeCompatibilityCheck
```

**Performance Targets**:

- Type equality check: < 100ns per check
- Type alias resolution: < 50ns per resolution (with caching)
- No more than 5% overhead in semantic analysis phase

**Files to Create**:

- `internal/types/benchmark_test.go` - Performance benchmarks

---

**Overall Success Criteria for Task 9.2**:

- ✅ `testdata/fixtures/Algorithms/quicksort.pas` passes
- ✅ All new unit tests pass (15+ tests)
- ✅ All new integration tests pass (5+ tests)
- ✅ No regressions in existing tests (2,100+ fixture tests)
- ✅ Type alias resolution works correctly across all compiler stages
- ✅ Performance overhead < 5%
- ✅ Comprehensive documentation

**Estimated Total Time**: 12-16 hours (1.5-2 days)

## Task 9.15: Support Parameterless Methods Callable Without Parentheses

**Goal**: Enable DWScript/Pascal syntax where parameterless methods/functions can be called without parentheses.

**Estimate**: 4-6 hours

**Status**: ✅ **DONE**

**Impact**: Critical for DWScript language compatibility - enables idiomatic code like `result := obj.GetValue` and string concatenation with method results like `'bottle' + obj.EvalS + ' of beer'`

**Description**: In DWScript/Pascal, parameterless methods can be invoked without parentheses when referenced as identifiers or via member access. This is different from method pointers - it's an implicit call that executes the method and returns its result.

**Implementation**:

1. **Semantic Analyzer Changes**:
   - `internal/semantic/analyze_expr_operators.go` (lines 127-142):
     - When analyzing identifiers that refer to parameterless methods in current class context
     - Return the method's return type (implicit call) instead of method pointer type
     - Error if method has parameters but is referenced without parentheses
   - `internal/semantic/analyze_classes.go` (lines 361-374):
     - When analyzing member access expressions (e.g., `obj.MethodName`)
     - Check if method has zero parameters
     - If yes: return method's return type (enables implicit call)
     - If no: return method pointer type (requires explicit call with parentheses)

2. **Interpreter Changes**:
   - `internal/interp/expressions_basic.go` (lines 92-106):
     - When evaluating identifiers that refer to parameterless methods
     - Create synthetic MethodCallExpression with empty arguments
     - Execute via evalMethodCall to get actual return value
   - `internal/interp/objects_hierarchy.go` (lines 364-371):
     - Already has auto-invoke logic for parameterless methods in member access
     - No changes needed - existing implementation is correct

3. **Return Value Dereferencing Fixes**:
   - `internal/interp/objects_methods.go` (lines 776-790):
     - When retrieving method return values from Result or method name variables
     - Check if value is a ReferenceValue and dereference it
     - Prevents returning reference wrappers instead of actual values
   - `internal/interp/objects_properties.go` (5 occurrences):
     - Same dereferencing logic for property getter return values
     - Ensures properties return actual values, not references

**Key Insight**: In DWScript, you can assign to either `Result` or use the function name itself as a variable. When the function name is used, it may be stored as a ReferenceValue that needs dereferencing when retrieved.

**Examples**:

```pascal
// Parameterless method called without parentheses
function GetValue: String;
begin
  Result := 'Hello';
end;

var x := obj.GetValue;  // Implicit call, x = 'Hello'

// In expressions
var msg := 'bottle' + obj.EvalS + ' of beer';  // EvalS is called implicitly

// With properties using method getters
property Line: String read GetLine;
PrintLn(obj.Line);  // GetLine() is called implicitly
```

**Testing**:

- ✅ Semantic analysis: No type errors for parameterless method references
- ✅ Simple method calls: Tested and working
- ✅ Property getters: Tested and working
- ✅ String concatenation with methods: Tested and working
- ⚠️ Complex test (bottles_of_beer): Still has runtime issues (separate from this fix)

**Files Modified**:

- `internal/semantic/analyze_expr_operators.go`
- `internal/semantic/analyze_classes.go`
- `internal/interp/expressions_basic.go`
- `internal/interp/objects_methods.go`
- `internal/interp/objects_properties.go`

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

- [ ] 9.16.6 Refactor type-specific nodes (ArrayLiteralExpression, CallExpression, NewExpression, MemberAccessExpression, etc.)
  - Embed appropriate base struct
  - Remove duplicates
  - Files: `pkg/ast/arrays.go`, `pkg/ast/classes.go`, `pkg/ast/functions.go` (~300 lines affected)

- [ ] 9.16.7 Update parser to use base struct constructors
  - [x] Update parser sites already touched (helpers/interfaces/const/type/property/field)
  - [ ] Sweep remaining parser files for struct literals using removed `Token` fields
  - [ ] Add helper constructors/macros if it simplifies repetitive initialization

- [ ] 9.16.8 Update semantic analyzer and interpreter
  - [x] Updated const/type/property/helper-specific tests where embedding occurred
  - [ ] Sweep remaining semantic analyzer files for direct field access needing updates
  - [ ] Sweep interpreter packages (bytecode + runtime) for struct literal updates
  - [ ] Add targeted regression tests for updated nodes

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

## Task 9.17 Refactor Visitor Pattern Implementation

**Goal**: Reduce visitor pattern code from 900+ lines to ~50-100 lines and make it extensible without modifying core code.

**Estimate**: 12-16 hours (1.5-2 days)

**Status**: IN PROGRESS

**Impact**: Major maintainability improvement, reduces visitor.go from 900 to ~100 lines, eliminates need to update visitor for new node types

**Description**: The current visitor implementation in `pkg/ast/visitor.go` is 900+ lines of boilerplate code. Every node type requires:
- A case in the main `Walk()` switch statement
- A dedicated `walkXXX()` function
- Manual handling of child nodes

This makes adding new node types tedious and error-prone. A reflection-based or code-generated approach can reduce this to ~50-100 lines while maintaining the same functionality.

**Current Problem**:

```go
// 900+ lines of boilerplate in visitor.go
func Walk(v Visitor, node Node) {
    if v = v.Visit(node); v == nil {
        return
    }

    // 195-line switch statement
    switch n := node.(type) {
    case *Program:
        walkProgram(n, v)
    case *Identifier:
        walkIdentifier(n, v)
    // ... 100+ more cases
    }
}

// Plus 100+ separate walk functions
func walkIdentifier(n *Identifier, v Visitor) { ... }
func walkBinaryExpression(n *BinaryExpression, v Visitor) { ... }
// ... 100+ more functions
```

**Strategy**: Implement reflection-based visitor with optional code generation:

**Phase 1**: Reflection-based visitor (runtime traversal)
- Use reflection to automatically traverse Node fields
- Detect slices of Nodes and individual Node fields
- Eliminate manual walk functions

**Phase 2** (Optional): Code generation for performance
- Generate walk functions from AST node definitions
- Zero runtime reflection cost
- `go generate` integration

**Complexity**: High - Requires reflection/code generation expertise and thorough testing

**Subtasks**:

- [ ] 9.17.1 Research and prototype reflection-based visitor
  - Study Go reflection API for struct field traversal
  - Prototype automatic child node detection
  - Benchmark performance vs current implementation
  - Document tradeoffs (flexibility vs performance)
  - File: `pkg/ast/visitor_reflect.go` (new file)

- [ ] 9.17.2 Implement reflection-based Walk function
  - Detect fields implementing Node interface
  - Handle slices of Nodes ([]Statement, []Expression, etc.)
  - Handle pointers to Nodes
  - Handle nil checks
  - File: `pkg/ast/visitor_reflect.go` (~100 lines)

- [ ] 9.17.3 Add visitor tags for special cases
  - Allow nodes to opt-out of automatic traversal with struct tags
  - Support custom traversal order with tags
  - Handle non-Node types that need walking (Parameter, CaseBranch, etc.)
  - Example: `type Foo struct { Child Node \`ast:"skip"\` }`

- [ ] 9.17.4 Preserve Inspect() convenience function
  - Keep existing Inspect() API for simple use cases
  - Ensure backward compatibility
  - File: `pkg/ast/visitor_reflect.go`

- [ ] 9.17.5 Test reflection visitor with existing code
  - All semantic analyzer visitors work unchanged
  - Symbol table construction works
  - Type checking works
  - LSP integration works
  - Files: `internal/semantic/*.go`

- [ ] 9.17.6 Performance testing and optimization
  - Benchmark reflection visitor vs original
  - Optimize hot paths if needed
  - Consider caching reflection metadata
  - Target: <10% performance degradation acceptable

- [ ] 9.17.7 (Optional) Add code generation alternative
  - Create `go generate` tool to generate walk functions
  - Parse AST node definitions
  - Generate type-safe walk functions
  - Tool: `cmd/gen-visitor/main.go` (new tool)

- [ ] 9.17.8 Update documentation
  - Explain new visitor architecture
  - Document struct tags for controlling traversal
  - Provide migration guide
  - File: `pkg/ast/doc.go`, `docs/ast-visitor.md` (new)

- [ ] 9.17.9 Deprecate old visitor implementation
  - Move old visitor.go to visitor_legacy.go
  - Add deprecation notices
  - Plan removal in future version

**Files Modified**:

- `pkg/ast/visitor.go` (replace 900 lines with ~100 lines of reflection code)
- `pkg/ast/visitor_reflect.go` (new file ~150 lines)
- `pkg/ast/visitor_test.go` (add reflection-specific tests)
- `pkg/ast/doc.go` (document new visitor approach)
- `docs/ast-visitor.md` (new architecture documentation)
- `cmd/gen-visitor/main.go` (optional code generator ~200 lines)

**Acceptance Criteria**:
- Visitor code reduced from 900 to ~100-150 lines
- Adding new node types doesn't require visitor updates
- All existing visitors continue to work (100% backward compatible)
- Performance degradation <10% (or use code generation for zero cost)
- Struct tags allow fine-grained control when needed
- Documentation explains new approach clearly

**Benefits**:
- 90% reduction in visitor boilerplate
- New node types automatically traversable
- More maintainable and less error-prone
- Easier to understand and modify
- Optional code generation for zero runtime cost

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

- [ ] 9.18.1 Design metadata architecture
  - Create SemanticInfo struct with node → type mapping
  - Design API for setting/getting type information
  - Consider thread safety for concurrent analyses
  - Document architecture decisions
  - File: `pkg/ast/metadata.go` (new file ~100 lines)

- [ ] 9.18.2 Implement SemanticInfo type
  - Map[Expression]*TypeAnnotation for expression types
  - Map[*Identifier]Symbol for symbol resolution
  - Thread-safe accessors with sync.RWMutex
  - File: `pkg/ast/metadata.go`

- [ ] 9.18.3 Remove Type field from AST expression nodes
  - Remove Type field from all expression node structs
  - Remove GetType/SetType methods (will be on SemanticInfo)
  - This affects ~30 node types
  - Files: `pkg/ast/ast.go`, `pkg/ast/statements.go`, `pkg/ast/control_flow.go`, etc.

- [ ] 9.18.4 Update semantic analyzer to use SemanticInfo
  - Pass SemanticInfo through analyzer methods
  - Replace node.SetType() with semanticInfo.SetType(node, typ)
  - Replace node.GetType() with semanticInfo.GetType(node)
  - Files: `internal/semantic/*.go` (~50 occurrences)

- [ ] 9.18.5 Update interpreter to use SemanticInfo
  - Pass SemanticInfo to interpreter
  - Get type information from SemanticInfo instead of nodes
  - Files: `internal/interp/*.go` (~30 occurrences)

- [ ] 9.18.6 Update bytecode compiler to use SemanticInfo
  - Pass SemanticInfo to compiler
  - Get type information from metadata table
  - Files: `internal/bytecode/compiler.go`

- [ ] 9.18.7 Update public API to return SemanticInfo
  - Engine.Analyze() returns SemanticInfo
  - Add accessor methods to Result type
  - Maintain backward compatibility where possible
  - Files: `pkg/dwscript/*.go`

- [ ] 9.18.8 Update LSP integration
  - Pass SemanticInfo to LSP handlers
  - Use metadata for hover, completion, etc.
  - Files: External go-dws-lsp project (document changes needed)

- [ ] 9.18.9 Run comprehensive test suite
  - All semantic analyzer tests pass
  - All interpreter tests pass
  - All bytecode VM tests pass
  - Memory usage reduced (verify with benchmarks)

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

## Task 9.19 Extract Pretty-Printing from AST Nodes

**Goal**: Remove String() implementation logic from AST nodes and create a dedicated printer package.

**Estimate**: 4-6 hours (0.5-1 day)

**Status**: IN PROGRESS

**Impact**: Better separation of concerns, enables multiple output formats, smaller AST code

**Description**: Currently, AST nodes contain extensive String() methods (some 50+ lines) that mix structural concerns with presentation logic. This makes the AST harder to maintain and limits output formats to a single hardcoded style. A dedicated printer package allows multiple output formats (compact, detailed, JSON, etc.) while keeping the AST focused on structure.

**Current Problem**:

```go
// ClassDecl.String() is 100+ lines of formatting logic!
func (cd *ClassDecl) String() string {
    var out bytes.Buffer
    out.WriteString("type ")
    out.WriteString(cd.Name.String())
    // ... 100 more lines of indentation, newlines, etc.
    return out.String()
}
```

**Strategy**:

1. Keep minimal String() methods on AST nodes (just type name + key info)
2. Create dedicated printer package with formatting logic
3. Support multiple output styles via printer options

**Complexity**: Low-Medium - Mostly moving code around, but need to ensure test compatibility

**Subtasks**:

- [ ] 9.19.1 Design printer package architecture
  - Printer struct with configurable options
  - Support for different styles (compact, detailed, multiline)
  - Support for different output formats (DWScript syntax, JSON, tree)
  - Document printer design
  - File: `pkg/printer/doc.go` (new package)

- [ ] 9.19.2 Create basic printer implementation
  - Printer struct with indent level, buffer, options
  - Methods for printing each node type
  - Helper methods for common patterns (indent, newline, etc.)
  - File: `pkg/printer/printer.go` (new file ~300 lines)

- [ ] 9.19.3 Simplify AST String() methods
  - Replace complex formatting with simple representation
  - Example: `func (cd *ClassDecl) String() string { return fmt.Sprintf("ClassDecl(%s)", cd.Name) }`
  - Keep String() for debugging, use printer for real output
  - Files: All `pkg/ast/*.go` files (~500 lines removed)

- [ ] 9.19.4 Add printer methods for all node types
  - PrintProgram(), PrintClassDecl(), PrintFunctionDecl(), etc.
  - Mirror existing String() behavior initially
  - File: `pkg/printer/printer.go` (~400 lines)

- [ ] 9.19.5 Add printer options and styles
  - CompactPrinter (minimal whitespace)
  - DetailedPrinter (full indentation, comments)
  - TreePrinter (AST structure visualization)
  - JSONPrinter (JSON representation)
  - File: `pkg/printer/styles.go` (new file ~100 lines)

- [ ] 9.19.6 Update tests to use printer
  - Tests that rely on String() output need updating
  - Use printer for expected output strings
  - Files: `pkg/ast/*_test.go`, parser tests

- [ ] 9.19.7 Update CLI to use printer
  - `parse` command uses printer for output
  - Add `--format` flag (dwscript, json, tree)
  - Files: `cmd/dwscript/commands.go`

- [ ] 9.19.8 Add printer tests
  - Test all node types print correctly
  - Test different styles produce valid output
  - Test JSON output is valid JSON
  - File: `pkg/printer/printer_test.go` (new file ~200 lines)

**Files Modified**:

- `pkg/printer/printer.go` (new file ~400 lines)
- `pkg/printer/styles.go` (new file ~100 lines)
- `pkg/printer/doc.go` (new file ~30 lines)
- `pkg/printer/printer_test.go` (new file ~200 lines)
- `pkg/ast/*.go` (simplify String() methods, ~500 lines reduced)
- `pkg/ast/*_test.go` (update tests to use printer if needed)
- `cmd/dwscript/commands.go` (use printer for parse command)

**Acceptance Criteria**:
- AST String() methods are simple (<5 lines each)
- Printer package handles all formatting logic
- Multiple output formats supported (at least: DWScript syntax, tree, JSON)
- All tests pass with printer-generated output
- CLI `parse` command can output different formats
- Documentation explains printer usage

**Benefits**:
- AST nodes focused on structure, not presentation
- Multiple output formats possible (JSON, tree view, etc.)
- Easier to change formatting without touching AST
- Smaller AST code (~500 lines reduced)
- Better separation of concerns

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

## Task 9.23: String Helper Methods (Type Helpers for String) 🎯 HIGH PRIORITY

**Goal**: Implement String type helper methods to enable method-call syntax on string values (e.g., `"hello".ToUpper()`, `s.Copy(2, 3)`).

**Estimate**: 12-16 hours (1.5-2 days)

**Status**: IN PROGRESS

**Priority**: HIGH - Blocks 46 out of 58 FunctionsString fixture tests (79% failure rate)

**Impact**:
- **Tests**: Fixes 46 failing FunctionsString tests
- **User Experience**: Enables more idiomatic DWScript string manipulation
- **Feature Completeness**: Stage 8.3 (Type Helpers) for String type

**Test Results**: Currently 14 passing, 46 failing (24% pass rate)
- Passing tests use only built-in functions (e.g., `UpperCase(s)`)
- Failing tests use helper methods (e.g., `s.ToUpper()`)

**Root Cause**: String type helpers are completely unimplemented. While 58+ string built-in functions exist (`UpperCase`, `Copy`, `Pos`, etc.), none are registered as helper methods on the String type. This prevents method-call syntax like `"test".StartsWith("t")` which is valid DWScript.

**DWScript Compatibility**: In DWScript, strings have helper methods that mirror built-in functions:
```pascal
// Both syntaxes are valid:
PrintLn(UpperCase('hello'));     // Built-in function ✓ (works)
PrintLn('hello'.ToUpper);         // Helper method ✗ (missing)

// More examples:
var s := 'banana';
PrintLn(Copy(s, 2, 3));           // Built-in ✓
PrintLn(s.Copy(2, 3));            // Helper ✗

PrintLn(StrBeginsWith(s, 'ba'));  // Built-in ✓
PrintLn(s.StartsWith('ba'));      // Helper ✗
```

**Missing Helper Methods** (based on FunctionsString test suite analysis):

**Conversion Methods** (5):
- `.ToInteger` → `StrToInt`
- `.ToFloat` → `StrToFloat`
- `.ToString` → identity (for consistency)
- `.ToString(base)` → `IntToStr(base)` (when called on numbers)
- `.ToHexString(width)` → `IntToHex(value, width)`

**Search/Check Methods** (4):
- `.StartsWith(str)` → `StrBeginsWith`
- `.EndsWith(str)` → `StrEndsWith`
- `.Contains(str)` → `StrContains`
- `.IndexOf(str)` → `Pos` (returns 1-based index)

**Extraction Methods** (3):
- `.Copy(start)` → `Copy(str, start, MaxInt)` (copy from start to end)
- `.Copy(start, len)` → `Copy(str, start, len)`
- `.Before(str)` → `StrBefore`
- `.After(str)` → `StrAfter`

**Modification Methods** (2):
- `.ToUpper` → `UpperCase`
- `.ToLower` → `LowerCase`
- `.Trim` → `Trim`

**Split/Join Methods** (2):
- `.Split(delimiter)` → `StrSplit`
- `.Join(array)` → `StrJoin` (called on array, not string)

**Total**: ~15-20 helper methods needed

**Implementation Strategy**:

1. **Lexer**: No changes needed (method call syntax already supported)
2. **Parser**: No changes needed (member access already parsed)
3. **Semantic Analyzer**: Register String helper methods in type system
4. **Interpreter**: Map helper method calls to existing built-in functions
5. **Bytecode VM**: Map helper methods to built-in opcodes or function calls
6. **Tests**: Add comprehensive tests for all helper methods

**Architecture**: Helper methods are syntactic sugar that delegate to existing built-in functions:

```text
"hello".ToUpper  →  [Semantic] resolve as method  →  [Runtime] call UpperCase("hello")
```

**Subtasks**:

### 9.23.1 Lexer (No Changes Required) ✓

- [x] 9.23.1.1 Verify method call syntax already tokenized
  - Method calls like `s.ToUpper()` already parsed correctly
  - No lexer changes needed
  - Status: VERIFIED (existing tests show this works)

### 9.23.2 Parser (No Changes Required) ✓

- [x] 9.23.2.1 Verify member access expressions already parsed
  - `s.Copy(2, 3)` → MemberExpression(object: s, member: Copy, args: [2, 3])
  - No parser changes needed
  - Status: VERIFIED (error messages confirm parser handles this)

### 9.23.3 Semantic Analyzer (Register String Helpers) ✓

- [x] 9.23.3.1 Design String helper registration system
  - Research how other type helpers are registered (Integer, Float, Array)
  - Design helper method metadata structure (name, signature, maps-to-function)
  - Document helper registration architecture
  - File: `internal/semantic/analyze_helpers.go` (existing architecture reviewed)
  - Status: DONE - Used existing HelperType registration pattern

- [x] 9.23.3.2 Register conversion helper methods
  - Register `.ToInteger` → maps to `StrToInt(self)`
  - Register `.ToFloat` → maps to `StrToFloat(self)`
  - Register `.ToString` → identity (returns self)
  - File: `internal/semantic/analyze_helpers.go` (initIntrinsicHelpers)
  - Status: DONE

- [x] 9.23.3.3 Register search/check helper methods
  - Register `.StartsWith(str)` → `StrBeginsWith(self, str)`
  - Register `.EndsWith(str)` → `StrEndsWith(self, str)`
  - Register `.Contains(str)` → `StrContains(self, str)`
  - Register `.IndexOf(str)` → `Pos(str, self)` (note parameter order!)
  - File: `internal/semantic/analyze_helpers.go`
  - Status: DONE

- [x] 9.23.3.4 Register extraction helper methods
  - Register `.Copy(start)` → `Copy(self, start, MaxInt)` (2-param variant)
  - Register `.Copy(start, len)` → `Copy(self, start, len)` (3-param variant)
  - Register `.Before(str)` → `StrBefore(self, str)`
  - Register `.After(str)` → `StrAfter(self, str)`
  - Handle method overloading for `.Copy()` (1 vs 2 parameters)
  - File: `internal/semantic/analyze_helpers.go`
  - Status: DONE (using optional parameter with default MaxInt)

- [x] 9.23.3.5 Register modification helper methods
  - Register `.ToUpper` → `UpperCase(self)` (no parens needed)
  - Register `.ToLower` → `LowerCase(self)` (no parens needed)
  - Register `.Trim` → `Trim(self)`
  - File: `internal/semantic/analyze_helpers.go`
  - Status: DONE

- [x] 9.23.3.6 Register split/join helper methods
  - Register `.Split(delimiter)` → `StrSplit(self, delimiter)`
  - Handle `.Join()` on array type (not string) - already exists
  - File: `internal/semantic/analyze_helpers.go`
  - Status: DONE

- [x] 9.23.3.7 Validate helper method type signatures
  - Ensure parameter types match underlying built-in function
  - Ensure return types match built-in function
  - Add proper error messages for type mismatches
  - File: `internal/semantic/analyze_helpers.go`
  - Status: DONE (type signatures specified in registration)

- [x] 9.23.3.8 Handle method overloading edge cases
  - `.Copy(start)` vs `.Copy(start, len)` - same name, different arity
  - Validate based on argument count
  - File: `internal/semantic/analyze_helpers.go`
  - Status: DONE (using NewFunctionTypeWithMetadata with optional parameters)

### 9.23.4 Interpreter (Runtime Helper Method Execution) ✓

- [x] 9.23.4.1 Implement helper method call dispatcher
  - When evaluating MemberExpression on String type, check if it's a helper
  - Route to appropriate built-in function
  - Transform `obj.Method(args)` → `BuiltinFunc(obj, args)`
  - File: `internal/interp/helpers_conversion.go` (evalBuiltinHelperMethod)
  - Status: DONE - Uses existing callHelperMethod/evalBuiltinHelperMethod architecture

- [x] 9.23.4.2 Implement conversion helper methods
  - `.ToInteger` → call `builtinStrToInt([self])`
  - `.ToFloat` → call `builtinStrToFloat([self])`
  - `.ToString` → return self unchanged
  - File: `internal/interp/helpers_conversion.go` (added cases to evalBuiltinHelperMethod)
  - Status: DONE

- [x] 9.23.4.3 Implement search/check helper methods
  - `.StartsWith(str)` → call `builtinStrBeginsWith([self, str])`
  - `.EndsWith(str)` → call `builtinStrEndsWith([self, str])`
  - `.Contains(str)` → call `builtinStrContains([self, str])`
  - `.IndexOf(str)` → call `builtinPos([str, self])` (REVERSED params!)
  - File: `internal/interp/helpers_conversion.go`
  - Status: DONE

- [x] 9.23.4.4 Implement extraction helper methods
  - `.Copy(start)` → call `builtinCopy([self, start, MaxInt])`
  - `.Copy(start, len)` → call `builtinCopy([self, start, len])`
  - `.Before(str)` → call `builtinStrBefore([self, str])`
  - `.After(str)` → call `builtinStrAfter([self, str])`
  - Handle overloading for `.Copy()`
  - File: `internal/interp/helpers_conversion.go`
  - Status: DONE

- [x] 9.23.4.5 Implement modification helper methods
  - `.ToUpper` → call `builtinUpperCase([self])`
  - `.ToLower` → call `builtinLowerCase([self])`
  - `.Trim` → call `builtinTrim([self])`
  - File: `internal/interp/helpers_conversion.go`
  - Status: DONE (ToUpper/ToLower were already implemented, Trim added)

- [x] 9.23.4.6 Implement split/join helper methods
  - `.Split(delimiter)` → call `builtinStrSplit([self, delimiter])`
  - File: `internal/interp/helpers_conversion.go`
  - Status: DONE

### 9.23.5 Bytecode VM (Bytecode Helper Method Support)

- [ ] 9.23.5.1 Map helper methods to bytecode operations
  - Option A: Emit CALL instructions to built-in functions
  - Option B: Add dedicated helper method opcodes (OpCallHelper)
  - Option C: Inline simple helpers (e.g., `.ToUpper` → OpStrUpper)
  - Decision: Use Option A (simplest, reuses existing built-ins)
  - File: `internal/bytecode/compiler.go`
  - Estimated: 2 hours

- [ ] 9.23.5.2 Compile helper method calls to built-in function calls
  - `s.ToUpper()` → emit `LOAD_VAR s; CALL UpperCase`
  - `s.Copy(2, 3)` → emit `LOAD_VAR s; LOAD_CONST 2; LOAD_CONST 3; CALL Copy`
  - Transform method syntax to function call syntax at compile time
  - File: `internal/bytecode/compiler.go` (compileMemberExpression)
  - Estimated: 1.5 hours

- [ ] 9.23.5.3 Handle parameter reordering in bytecode
  - `.IndexOf(substr)` → `Pos(substr, self)` (params reversed!)
  - Emit instructions in correct order for built-in function
  - File: `internal/bytecode/compiler.go`
  - Estimated: 1 hour

- [ ] 9.23.5.4 Test bytecode execution of helper methods
  - Verify all helper methods work in bytecode VM
  - Compare results with AST interpreter
  - File: `internal/bytecode/vm_test.go` or `internal/bytecode/string_helpers_test.go`
  - Estimated: 1 hour

### 9.23.6 Testing

- [ ] 9.23.6.1 Add unit tests for semantic analyzer
  - Test helper method resolution
  - Test type checking for helper methods
  - Test error cases (wrong parameter types, unknown methods)
  - File: `internal/semantic/string_helpers_test.go` (new file)
  - Estimated: 1 hour

- [ ] 9.23.6.2 Add unit tests for interpreter
  - Test each helper method individually
  - Test method chaining (e.g., `s.Copy(2, 3).ToUpper()`)
  - Test edge cases (empty strings, nil, etc.)
  - File: `internal/interp/string_helpers_test.go` (new file)
  - Estimated: 2 hours

- [ ] 9.23.6.3 Add unit tests for bytecode VM
  - Test helper methods in bytecode execution
  - Compare with interpreter results
  - File: `internal/bytecode/string_helpers_test.go` (new file)
  - Estimated: 1 hour

- [ ] 9.23.6.4 Verify FunctionsString fixture tests pass
  - Run: `go test -v ./internal/interp -run TestDWScriptFixtures/FunctionsString`
  - Target: 58/58 tests passing (100% pass rate, up from 24%)
  - Fix any remaining test failures
  - Update `testdata/fixtures/TEST_STATUS.md`
  - Estimated: 2 hours

- [ ] 9.23.6.5 Add integration tests
  - Test helper methods in complex scripts
  - Test method chaining and composition
  - Test with variables, function returns, expressions
  - File: `testdata/string_helpers_integration.dws` (new test script)
  - Estimated: 1 hour

### 9.23.7 Additional Built-in Function Fixes

- [ ] 9.23.7.1 Fix Copy() 2-parameter variant
  - Current: `Copy(str, start, len)` ✓
  - Missing: `Copy(str, start)` - should copy from start to end
  - Implementation: Default len to MaxInt when omitted
  - Files: `internal/parser/expressions.go`, `internal/interp/builtins_strings_basic.go`
  - Estimated: 1 hour

- [ ] 9.23.7.2 Fix FloatToStr to accept Integer arguments
  - Current: `FloatToStr(Integer)` → type error ✗
  - Expected: Auto-convert Integer to Float before formatting
  - DWScript behavior: Accepts numeric types and auto-converts
  - File: `internal/semantic/analyze_function_calls.go` or `internal/interp/builtins_conversion.go`
  - Estimated: 1 hour

### 9.23.8 Documentation

- [ ] 9.23.8.1 Document String helper methods
  - List all available helper methods
  - Show equivalence to built-in functions
  - Provide examples for each method
  - File: `docs/string-helpers.md` (new file ~100 lines)
  - Estimated: 1 hour

- [ ] 9.23.8.2 Update CLAUDE.md with helper method info
  - Add String helper section to language reference
  - Update built-in functions list to mention helper variants
  - File: `CLAUDE.md`
  - Estimated: 30 minutes

- [ ] 9.23.8.3 Add code comments to helper implementation
  - Document helper registration mechanism
  - Explain mapping from helpers to built-ins
  - File: `internal/semantic/builtin_helpers_string.go`
  - Estimated: 30 minutes

**Files Created**:
- `internal/semantic/builtin_helpers_string.go` (new file ~200 lines) - Helper registration
- `internal/interp/builtin_helpers_string.go` (new file ~150 lines) - Helper execution
- `internal/semantic/string_helpers_test.go` (new file ~200 lines) - Semantic tests
- `internal/interp/string_helpers_test.go` (new file ~300 lines) - Runtime tests
- `internal/bytecode/string_helpers_test.go` (new file ~200 lines) - Bytecode tests
- `testdata/string_helpers_integration.dws` (new file ~100 lines) - Integration tests
- `docs/string-helpers.md` (new file ~100 lines) - Documentation

**Files Modified**:
- `internal/semantic/type_helpers.go` (add String helper support ~50 lines)
- `internal/semantic/analyze_member_access.go` (validate String helpers ~30 lines)
- `internal/interp/expressions.go` (route helper calls ~40 lines)
- `internal/interp/builtins_strings_basic.go` (add Copy 2-param variant ~20 lines)
- `internal/interp/builtins_conversion.go` (fix FloatToStr auto-conversion ~15 lines)
- `internal/bytecode/compiler.go` (compile helper methods ~50 lines)
- `testdata/fixtures/TEST_STATUS.md` (update FunctionsString results ~10 lines)
- `CLAUDE.md` (document String helpers ~30 lines)

**Acceptance Criteria**:
- All 15-20 String helper methods registered and working
- FunctionsString fixture tests: 58/58 passing (100%, up from 24%)
- Helper methods work in both AST interpreter and bytecode VM
- Proper type checking and error messages for helper methods
- Method overloading works correctly (e.g., `.Copy(start)` vs `.Copy(start, len)`)
- Parameter reordering handled correctly (e.g., `.IndexOf()` → `Pos()`)
- Copy() 2-parameter variant implemented
- FloatToStr() accepts Integer arguments
- Comprehensive test coverage (unit + integration + fixtures)
- Documentation complete with examples

**Benefits**:
- Fixes 46 failing FunctionsString tests (79% of test suite)
- Enables idiomatic DWScript string manipulation syntax
- Completes Stage 8.3 (Type Helpers) for String type
- Improves DWScript compatibility (helper methods are standard in DWScript)
- Better developer experience (method chaining, auto-completion friendly)
- Foundation for other type helpers (Integer, Float, Array already partially done)

**Related Tasks**:
- Stage 8.3: Type Helpers (this implements String helpers specifically)
- Task 9.8: Array Helper Methods (similar pattern, different type)
- Task 9.12: SetLength on String Type (related to string manipulation)

---

## Phase 10: go-dws API Enhancements for LSP Integration ✅ COMPLETE

**Goal**: Enhanced go-dws library with structured errors, AST access, position metadata, symbol tables, and type information for LSP features.

**Status**: All 27 tasks complete. Added public `pkg/ast/` and `pkg/token/` packages, structured error types with position info, Parse() mode for fast syntax-only parsing, visitor pattern for AST traversal, symbol table access, and type queries. 100% backwards compatible. Ready for go-dws-lsp integration.

---

## Phase 11: Bytecode Compiler & VM Optimizations ✅ MOSTLY COMPLETE

**Status**: Core implementation complete | **Performance**: 5-6x faster than AST interpreter | **Tasks**: 15 complete, 2 pending

### Overview

This phase implements a bytecode virtual machine for DWScript, providing significant performance improvements over the tree-walking AST interpreter. The bytecode VM uses a stack-based architecture with 116 opcodes and includes an optimization pipeline.

**Architecture**: AST → Compiler → Bytecode → VM → Output

### Phase 11.1: Bytecode VM Foundation ✅ COMPLETE

- [x] 11.1 Research and design bytecode instruction set
  - Stack-based VM with 116 opcodes, 32-bit instruction format
  - Documentation: [bytecode-vm-design.md](docs/architecture/bytecode-vm-design.md)
  - Expected Impact: 2-3x speedup over tree-walking interpreter

- [x] 11.2 Implement bytecode data structures
  - Created `internal/bytecode/bytecode.go` with `Chunk` type (bytecode + constants pool)
  - Implemented constant pool for literals with deduplication
  - Added line number mapping with run-length encoding
  - Implemented bytecode disassembler for debugging (79.7% coverage)

- [x] 11.3 Build AST-to-bytecode compiler
  - Created `internal/bytecode/compiler.go` with visitor pattern
  - Compile expressions: literals, binary ops, unary ops, variables, function calls
  - Compile statements: assignment, if/else, loops, return
  - Handle scoping and variable resolution
  - Optimize constant folding during compilation

- [x] 11.4 Implement bytecode VM core
  - Created `internal/bytecode/vm.go` with instruction dispatch loop
  - Implemented operand stack and call stack
  - Added environment/closure handling with upvalue capture
  - Error handling with structured RuntimeError and stack traces
  - Performance: VM is ~5.6x faster than AST interpreter

- [x] 11.5 Implement arithmetic and logic instructions
  - ADD, SUB, MUL, DIV, MOD instructions
  - NEGATE, NOT instructions
  - EQ, NE, LT, LE, GT, GE comparisons
  - AND, OR, XOR bitwise operations
  - Type coercion (int ↔ float)

- [x] 11.6 Implement variable and memory instructions
  - LOAD_CONST / LOAD_LOCAL / STORE_LOCAL
  - LOAD_GLOBAL / STORE_GLOBAL
  - LOAD_UPVALUE / STORE_UPVALUE with closure capture
  - GET_PROPERTY / SET_PROPERTY for member access

- [x] 11.7 Implement control flow instructions
  - JUMP, JUMP_IF_FALSE, JUMP_IF_TRUE
  - LOOP (jump backward for while/for loops)
  - Patch jump addresses during compilation
  - Break/continue leverage jump instructions

- [x] 11.8 Implement function call instructions
  - CALL instruction for named functions
  - RETURN instruction with trailing return guarantee
  - Handle recursion and call stack depth
  - Implement closures and upvalues
  - Support method calls and `Self` context (OpCallMethod, OpGetSelf)

- [x] 11.9 Implement array and object instructions
  - GET_INDEX, SET_INDEX for array access
  - NEW_ARRAY, ARRAY_LENGTH
  - NEW_OBJECT for class instantiation
  - INVOKE_METHOD for method dispatch

- [x] 11.10 Add exception handling instructions
  - TRY, CATCH, FINALLY, THROW instructions
  - Exception stack unwinding
  - Preserve stack traces across bytecode execution

- [x] 11.11 Optimize bytecode generation
  - Established optimization pipeline with pass manager and toggles
  - Peephole transforms: fold literal push/pop pairs, collapse stack shuffles
  - Dead code elimination: trim after terminators, reflow jump targets
  - Constant propagation: track literal locals/globals, fold arithmetic chains
  - Inline small functions (< 10 instructions)

- [x] 11.12 Integrate bytecode VM into interpreter
  - Added `--bytecode` flag to CLI
  - Added `CompileMode` option (AST vs Bytecode) to `pkg/dwscript/options.go`
  - Bytecode compilation/execution paths in `pkg/dwscript/dwscript.go`
  - Unit loading/parsing parity, tracing, diagnostic output
  - Wire bytecode VM to externals (FFI, built-ins, stdout capture)

- [x] 11.13 Create bytecode test suite
  - Port existing interpreter tests to bytecode
  - Test bytecode disassembler output
  - Verify identical behavior to AST interpreter
  - Performance benchmarks confirm 5-6x speedup

- [x] 11.14 Add bytecode serialization
  - [x] 11.14.1 Define bytecode file format (.dwc)
    - **Task**: Design the binary format for bytecode files
    - **Implementation**:
      - Define magic number (e.g., "DWC\x00") for file identification
      - Define version format (major.minor.patch)
      - Define header structure (magic, version, metadata)
      - Document format specification
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 0.5 day

  - [x] 11.14.2 Implement Chunk serialization
    - **Task**: Serialize bytecode chunks to binary format
    - **Implementation**:
      - Serialize instructions array
      - Serialize line number information
      - Serialize constants pool
      - Write helper functions for writing primitives (int, float, string, bool)
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 1 day

  - [x] 11.14.3 Implement Chunk deserialization
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

  - [x] 11.14.4 Add version compatibility checks
    - **Task**: Ensure bytecode version compatibility
    - **Implementation**:
      - Check version during deserialization
      - Return descriptive errors for version mismatches
      - Add tests for different version scenarios
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 0.5 day

  - [x] 11.14.5 Add serialization tests
    - **Task**: Test serialization/deserialization round-trip
    - **Implementation**:
      - Test simple programs serialize correctly
      - Test complex programs with all value types
      - Test error handling (corrupt files, version mismatches)
      - Verify bytecode produces same output after round-trip
    - **Files**: `internal/bytecode/serializer_test.go`
    - **Estimated time**: 1 day

  - [x] 11.14.6 Add `dwscript compile` command
    - **Task**: CLI command to compile source to bytecode
    - **Implementation**:
      - Add compile subcommand to CLI
      - Parse source file and compile to bytecode
      - Write bytecode to .dwc file
      - Add flags for output file, optimization level
    - **Files**: `cmd/dwscript/main.go`, `cmd/dwscript/compile.go`
    - **Estimated time**: 0.5 day

  - [x] 11.14.7 Update `dwscript run` to load .dwc files
    - **Task**: Allow running precompiled bytecode files
    - **Implementation**:
      - Detect .dwc file extension
      - Load bytecode from file instead of compiling
      - Add performance comparison in benchmarks
    - **Files**: `cmd/dwscript/main.go`, `cmd/dwscript/run.go`
    - **Estimated time**: 0.5 day

  - [x] 11.14.8 Document bytecode serialization
    - **Task**: Update documentation for bytecode files
    - **Implementation**:
      - Document .dwc file format in docs/bytecode-vm.md
      - Add CLI examples for compile command
      - Update README.md with serialization info
    - **Files**: `docs/bytecode-vm.md`, `README.md`, `CLAUDE.md`
    - **Estimated time**: 0.5 day

- [x] 11.15 Document bytecode VM
  - Written `docs/bytecode-vm.md` explaining architecture
  - Documented instruction set and opcodes
  - Provided examples of bytecode output
  - Updated CLAUDE.md with bytecode information

**Estimated time**: Completed in 12-16 weeks

### Phase 11.2: Future Bytecode Optimizations (DEFERRED)

- [ ] 11.16 Advanced peephole optimizations
  - [ ] Strength reduction (multiplication → shift)
  - [ ] Common subexpression elimination
  - [ ] Branch prediction hints

- [ ] 11.17 Register allocation improvements
  - [ ] Live range analysis
  - [ ] Register coloring for locals
  - [ ] Reduce stack traffic

- [ ] 11.18 Inline caching for method dispatch
  - [ ] Cache method lookup results
  - [ ] Invalidate on class redefinition
  - [ ] Benchmark polymorphic call sites

- [ ] 11.19 Bytecode verification
  - [ ] Static analysis of bytecode correctness
  - [ ] Type safety verification
  - [ ] Stack depth validation

---

## Phase 12: Performance & Polish

### Performance Profiling

- [x] 12.1 Create performance benchmark scripts
- [x] 12.2 Profile lexer performance: `BenchmarkLexer`
- [x] 12.3 Profile parser performance: `BenchmarkParser`
- [x] 12.4 Profile interpreter performance: `BenchmarkInterpreter`
- [x] 12.5 Identify bottlenecks using `pprof`
- [ ] 12.6 Document performance baseline

### Optimization - Lexer

- [ ] 12.7 Optimize string handling in lexer (use bytes instead of runes where possible)
- [ ] 12.8 Reduce allocations in token creation
- [ ] 12.9 Use string interning for keywords/identifiers
- [ ] 12.10 Benchmark improvements

### Optimization - Parser

- [ ] 12.11 Reduce AST node allocations
- [ ] 12.12 Pool commonly created nodes
- [ ] 12.13 Optimize precedence table lookups
- [ ] 12.14 Benchmark improvements

### Optimization - Interpreter

- [ ] 12.15 Optimize value representation (avoid interface{} overhead if possible)
- [ ] 12.16 Use switch statements instead of type assertions where possible
- [ ] 12.17 Cache frequently accessed symbols
- [ ] 12.18 Optimize environment lookups
- [ ] 12.19 Reduce allocations in hot paths
- [ ] 12.20 Benchmark improvements

### Memory Management

- [ ] 12.21 Ensure no memory leaks in long-running scripts
- [ ] 12.22 Profile memory usage with large programs
- [ ] 12.23 Optimize object allocation/deallocation
- [ ] 12.24 Consider object pooling for common types

### Code Quality Refactoring

- [ ] 12.25 Run `go vet ./...` and fix all issues
- [ ] 12.26 Run `golangci-lint run` and address warnings
- [ ] 12.27 Run `gofmt` on all files
- [ ] 12.28 Run `goimports` to organize imports
- [ ] 12.29 Review error handling consistency
- [ ] 12.30 Unify value representation if inconsistent
- [ ] 12.31 Refactor large functions into smaller ones
- [ ] 12.32 Extract common patterns into helper functions
- [ ] 12.33 Improve variable/function naming
- [ ] 12.34 Add missing error checks

### Documentation

- [ ] 12.35 Write comprehensive GoDoc comments for all exported types/functions
- [ ] 12.36 Document internal architecture in `docs/architecture.md`
- [ ] 12.37 Create user guide in `docs/user_guide.md`
- [ ] 12.38 Document CLI usage with examples
- [ ] 12.39 Create API documentation for embedding the library
- [ ] 12.40 Add code examples to documentation
- [ ] 12.41 Document known limitations
- [ ] 12.42 Create contribution guidelines in `CONTRIBUTING.md`

### Example Programs

- [x] 12.43 Create `examples/` directory
- [x] 12.44 Add example scripts:
  - [x] Hello World
  - [x] Fibonacci
  - [x] Factorial
  - [x] Class-based example (Person demo)
  - [x] Algorithm sample (math/loops showcase)
- [x] 12.45 Add README in examples directory
- [x] 12.46 Ensure all examples run correctly

### Testing Enhancements

- [ ] 12.47 Add integration tests in `test/integration/`
- [ ] 12.48 Add fuzzing tests for parser: `FuzzParser`
- [ ] 12.49 Add fuzzing tests for lexer: `FuzzLexer`
- [ ] 12.50 Add property-based tests (using testing/quick or gopter)
- [ ] 12.51 Ensure CI runs all test types
- [ ] 12.52 Achieve >90% code coverage overall
- [ ] 12.53 Add regression tests for all fixed bugs

### Release Preparation

- [ ] 12.54 Create `CHANGELOG.md`
- [ ] 12.55 Document version numbering scheme (SemVer)
- [ ] 12.56 Tag v0.1.0 alpha release
- [ ] 12.57 Create release binaries for major platforms (Linux, macOS, Windows)
- [ ] 12.58 Publish release on GitHub
- [ ] 12.59 Write announcement blog post or README update
- [ ] 12.60 Share with community for feedback

---

## Phase 13: Go Source Code Generation & AOT Compilation [RECOMMENDED]

**Status**: Not started | **Priority**: HIGH | **Estimated Time**: 20-28 weeks (code generation) + 9-13 weeks (CLI)

### Overview

This phase implements ahead-of-time (AOT) compilation by transpiling DWScript source code to Go, then compiling to native executables. This approach leverages Go's excellent cross-compilation support and delivers near-native performance.

**Approach**: DWScript Source → AST → Go Source Code → Go Compiler → Native Executable

**Benefits**: 10-50x faster than tree-walking interpreter, excellent portability, leverages Go toolchain

### Phase 13.1: Go Source Code Generation (20-28 weeks)

- [ ] 13.1 Design Go code generation architecture
  - Study similar transpilers (c2go, ast-transpiler)
  - Design AST → Go AST transformation strategy
  - Define runtime library interface
  - Document type mapping (DWScript → Go)
  - Plan package structure for generated code
  - **Decision**: Use `go/ast` package for Go AST generation

- [ ] 13.2 Create Go code generator foundation
  - Create `internal/codegen/` package
  - Create `internal/codegen/go_generator.go`
  - Implement `Generator` struct with context tracking
  - Add helper methods for code emission
  - Set up `go/ast` and `go/printer` integration
  - Create unit tests for basic generation

- [ ] 13.3 Implement type system mapping
  - Map DWScript primitives to Go types (Integer→int64, Float→float64, String→string, Boolean→bool)
  - Map DWScript arrays to Go slices (dynamic) or arrays (static)
  - Map DWScript records to Go structs
  - Map DWScript classes to Go structs with method tables
  - Handle type aliases and subrange types
  - Document type mapping in `docs/codegen-types.md`

- [ ] 13.4 Generate code for expressions
  - Generate literals (integer, float, string, boolean, nil)
  - Generate identifiers (variables, constants)
  - Generate binary operations (+, -, *, /, =, <>, <, >, etc.)
  - Generate unary operations (-, not)
  - Generate function calls
  - Generate array/object member access
  - Handle operator precedence correctly
  - Add unit tests comparing eval vs generated code

- [ ] 13.5 Generate code for statements
  - Generate variable declarations (`var x: Integer = 42`)
  - Generate assignments (`x := 10`)
  - Generate if/else statements
  - Generate while/repeat/for loops
  - Generate case statements (switch in Go)
  - Generate begin...end blocks
  - Handle break/continue/exit statements

- [ ] 13.6 Generate code for functions and procedures
  - Generate function declarations with parameters and return type
  - Handle by-value and by-reference (var) parameters
  - Generate procedure declarations (no return value)
  - Implement nested functions (closures in Go)
  - Support forward declarations
  - Handle recursion
  - Generate proper variable scoping

- [ ] 13.7 Generate code for classes and OOP
  - Generate Go struct definitions for classes
  - Generate constructor functions (Create)
  - Generate destructor cleanup (Destroy → defer)
  - Generate method declarations (receiver functions)
  - Implement inheritance (embedding in Go)
  - Implement virtual method dispatch (method tables)
  - Handle class fields and properties
  - Support `Self` keyword (receiver parameter)

- [ ] 13.8 Generate code for interfaces
  - Generate Go interface definitions
  - Implement interface casting and type assertions
  - Generate interface method dispatch
  - Handle interface inheritance
  - Support interface variables and parameters

- [ ] 13.9 Generate code for records
  - Generate Go struct definitions
  - Support record methods (static and instance)
  - Handle record literals and initialization
  - Generate record field access

- [ ] 13.10 Generate code for enums
  - Generate Go const declarations with iota
  - Support scoped and unscoped enum access
  - Generate Ord() and Integer() conversions
  - Handle explicit enum values

- [ ] 13.11 Generate code for arrays
  - Generate static arrays (Go arrays: `[10]int`)
  - Generate dynamic arrays (Go slices: `[]int`)
  - Support array literals
  - Generate array indexing and slicing
  - Implement SetLength, High, Low built-ins
  - Handle multi-dimensional arrays

- [ ] 13.12 Generate code for sets
  - Generate set types as Go map[T]bool or bitsets
  - Support set literals and constructors
  - Generate set operations (union, intersection, difference)
  - Implement `in` operator for set membership

- [ ] 13.13 Generate code for properties
  - Translate properties to getter/setter methods
  - Generate field-backed properties (direct access)
  - Generate method-backed properties (method calls)
  - Support read-only and write-only properties
  - Handle auto-properties

- [ ] 13.14 Generate code for exceptions
  - Generate try/except/finally as Go defer/recover
  - Map DWScript exceptions to Go error types
  - Generate raise statements (panic)
  - Implement exception class hierarchy
  - Preserve stack traces

- [ ] 13.15 Generate code for operators and conversions
  - Generate operator overloads as functions
  - Generate implicit conversions
  - Handle type coercion in expressions
  - Support custom operators

- [ ] 13.16 Create runtime library for generated code
  - Create `pkg/runtime/` package
  - Implement built-in functions (PrintLn, Length, Copy, etc.)
  - Implement array/string manipulation functions
  - Implement math functions (Sin, Cos, Sqrt, etc.)
  - Implement date/time functions
  - Provide runtime type information (RTTI) for reflection
  - Support external function calls (FFI)

- [ ] 13.17 Handle units/modules compilation
  - Generate separate Go packages for each unit
  - Handle unit dependencies and imports
  - Generate initialization/finalization code
  - Support uses clauses
  - Create package manifest

- [ ] 13.18 Implement optimization passes
  - Constant folding
  - Dead code elimination
  - Inline small functions
  - Remove unused variables
  - Optimize string concatenation
  - Use Go compiler optimization hints (//go:inline, etc.)

- [ ] 13.19 Add source mapping for debugging
  - Preserve line number comments in generated code
  - Generate source map files (.map)
  - Add DWScript source file embedding
  - Support stack trace translation (Go → DWScript)

- [ ] 13.20 Test Go code generation
  - Generate code for all fixture tests
  - Compile and run generated code
  - Compare output with interpreter
  - Measure compilation time
  - Benchmark generated code performance

**Expected Results**: 10-50x faster than tree-walking interpreter, near-native Go speed

### Phase 13.2: AOT Compiler CLI (9-13 weeks)

- [ ] 13.21 Create `dwscript compile` command
  - Add `compile` subcommand to CLI
  - Parse input DWScript file(s)
  - Generate Go source code to output directory
  - Invoke `go build` to create executable
  - Support multiple output formats (executable, library, package)

- [ ] 13.22 Implement project compilation mode
  - Support compiling entire projects (multiple units)
  - Generate go.mod file
  - Handle dependencies between units
  - Create main package with entry point
  - Support compilation configuration (optimization level, target platform)

- [ ] 13.23 Add compilation flags and options
  - `--output` or `-o` for output path
  - `--optimize` or `-O` for optimization level (0, 1, 2, 3)
  - `--keep-go-source` to preserve generated Go files
  - `--target` for cross-compilation (linux, windows, darwin, wasm)
  - `--static` for static linking
  - `--debug` to include debug symbols

- [ ] 13.24 Implement cross-compilation support
  - Support GOOS and GOARCH environment variables
  - Generate platform-specific code (if needed)
  - Test compilation for Linux, macOS, Windows, WASM
  - Document platform-specific limitations

- [ ] 13.25 Add incremental compilation
  - Cache compiled units
  - Detect file changes (mtime, hash)
  - Recompile only changed units
  - Rebuild dependency graph
  - Speed up repeated compilations

- [ ] 13.26 Create standalone binary builder
  - Generate single-file executable
  - Embed DWScript runtime
  - Strip debug symbols (optional)
  - Compress binary with UPX (optional)
  - Test on different platforms

- [ ] 13.27 Implement library compilation mode
  - Generate Go package (not executable)
  - Export public functions/classes
  - Create Go-friendly API
  - Generate documentation (godoc)
  - Support embedding in other Go projects

- [ ] 13.28 Add compilation error reporting
  - Catch Go compilation errors
  - Translate errors to DWScript source locations
  - Provide helpful error messages
  - Suggest fixes for common issues

- [ ] 13.29 Create compilation test suite
  - Test compilation of all fixture tests
  - Verify all executables run correctly
  - Test cross-compilation
  - Benchmark compilation speed
  - Measure binary sizes

- [ ] 13.30 Document AOT compilation
  - Write `docs/aot-compilation.md`
  - Explain compilation process
  - Provide usage examples
  - Document performance characteristics
  - Compare with interpretation and bytecode VM

---

## Phase 14: WebAssembly Runtime & Playground ✅ MOSTLY COMPLETE

**Status**: Core implementation complete | **Priority**: HIGH | **Tasks**: 23 complete, 3 pending

### Overview

This phase implements WebAssembly support for running DWScript in browsers, including a platform abstraction layer, WASM build infrastructure, JavaScript/Go bridge, and a web-based playground with Monaco editor integration.

**Architecture**: DWScript → WASM Binary → Browser/Node.js → JavaScript API

### Phase 14.1: Platform Abstraction Layer ✅ COMPLETE

- [x] 14.1 Create `pkg/platform/` package with core interfaces
  - FileSystem, Console, Platform interfaces
  - Enables native and WebAssembly builds with consistent behavior

- [x] 14.2 Implement `pkg/platform/native/` for standard Go
  - Standard Go implementations for native builds
  - Direct OS filesystem and console access

- [x] 14.3 Implement `pkg/platform/wasm/` with virtual filesystem
  - In-memory map for file storage
  - Console bridge to JavaScript console.log
  - Time functions using JavaScript Date API
  - Sleep implementation using setTimeout

- [ ] 14.4 Create feature parity test suite
  - Tests that run on both native and WASM
  - Validate platform abstraction works correctly

- [ ] 14.5 Document platform differences and limitations
  - Platform-specific behavior documentation
  - Known limitations in WASM environment

### Phase 14.2: WASM Build Infrastructure ✅ COMPLETE

- [x] 14.6 Create build infrastructure
  - `build/wasm/` directory with scripts
  - Justfile targets: `just wasm`, `just wasm-test`, `just wasm-optimize`, etc.
  - `cmd/dwscript-wasm/main.go` entry point with syscall/js exports

- [x] 14.7 Implement build modes support
  - Monolithic, modular, hybrid modes (compile-time flags)
  - `pkg/wasm/` package for WASM bridge code

- [x] 14.8 Add wasm_exec.js and optimization
  - wasm_exec.js from Go distribution (multi-version support)
  - Integrate wasm-opt (Binaryen) for binary size optimization
  - Size monitoring (warns if >3MB uncompressed)

- [ ] 14.9 Test all build modes
  - Compare sizes and performance
  - Validate each mode works correctly

- [x] 14.10 Document build process
  - `docs/wasm/BUILD.md` with build instructions
  - Configuration options and troubleshooting

### Phase 14.3: JavaScript/Go Bridge ✅ COMPLETE

- [x] 14.11 Implement DWScript class API
  - `pkg/wasm/api.go` using syscall/js
  - Export init(), compile(), run(), eval() to JavaScript

- [x] 14.12 Create type conversion utilities
  - Go types ↔ js.Value conversion in utils.go
  - Proper handling of DWScript types in JavaScript

- [x] 14.13 Implement callback registration system
  - `pkg/wasm/callbacks.go` for event handling
  - Virtual filesystem interface for JavaScript

- [x] 14.14 Add error handling across boundary
  - Panics → exceptions with recovery
  - Structured error objects for DWScript runtime errors

- [x] 14.15 Add event system
  - on() method for output, error, and custom events
  - Memory management with proper js.Value.Release()

- [x] 14.16 Document JavaScript API
  - `docs/wasm/API.md` with complete API reference
  - Usage examples for browser and Node.js

### Phase 14.4: Web Playground ✅ COMPLETE

- [x] 14.17 Create playground directory structure
  - `playground/` with HTML/CSS/JS files
  - Monaco Editor integration

- [x] 14.18 Implement syntax highlighting
  - DWScript language definition for Monaco
  - Tokenization rules matching lexer

- [x] 14.19 Build split-pane UI
  - Code editor + output console
  - Toolbar with Run, Examples, Clear, Share, Theme buttons

- [x] 14.20 Implement URL-based code sharing
  - Base64 encoded code in fragment
  - Examples dropdown with sample programs

- [x] 14.21 Add localStorage features
  - Auto-save and restore user code
  - Error markers in editor from compilation errors

- [x] 14.22 Set up GitHub Pages deployment
  - GitHub Actions workflow for automated deployment
  - Testing checklist in playground/TESTING.md

- [x] 14.23 Document playground architecture
  - `docs/wasm/PLAYGROUND.md` with architecture details
  - Extension points for future features

### Phase 14.5: NPM Package ✅ MOSTLY COMPLETE

- [x] 14.24 Create NPM package structure
  - `npm/` with package.json
  - TypeScript definitions in `typescript/index.d.ts`

- [x] 14.25 Create dual ESM/CommonJS entry points
  - index.js (ESM) and index.cjs (CommonJS)
  - WASM loader helper for Node.js and browser

- [x] 14.26 Add usage examples
  - Node.js, React, Vue, vanilla JS examples
  - Automated NPM publishing via GitHub Actions

- [x] 14.27 Configure for tree-shaking
  - Optimal bundling configuration
  - `npm/README.md` with installation guide

- [ ] 14.28 Publish to npmjs.com
  - Initial version publication
  - Version management strategy

### Phase 14.6: Testing & Documentation

- [ ] 14.29 Write WASM-specific tests
  - GOOS=js GOARCH=wasm go test
  - Node.js integration test suite

- [ ] 14.30 Add browser tests
  - Playwright tests for Chrome, Firefox, Safari
  - CI matrix for cross-browser testing

- [ ] 14.31 Add performance benchmarks
  - Compare WASM vs native speed
  - Bundle size regression monitoring in CI

- [ ] 14.32 Write embedding guide
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

## Phase 15: MIR Foundation [DEFERRED]

**Status**: Not started | **Priority**: MEDIUM | **Estimated Tasks**: 47 (MIR core, lowering, testing)

### Overview

This phase implements a Mid-level Intermediate Representation (MIR) that serves as a target-neutral layer between the type-checked AST and backend code generators. The MIR enables multiple backend targets (JavaScript, LLVM, C, etc.) from a single lowering pass.

**Architecture Flow**: DWScript Source → Parser → Semantic Analyzer → **MIR Builder** → [Backend Emitter] → Output

**Why MIR?** Clean separation of concerns, multi-backend support, platform-independent optimizations, easier debugging, and future-proofing for additional compilation targets.

**Note**: JavaScript backend is implemented in Phase 16, LLVM backend in Phase 17.

### Phase 15.1: MIR Foundation (30 tasks)

**Goal**: Define a complete, verifiable mid-level IR that can represent all DWScript constructs in a target-neutral way.

**Exit Criteria**: MIR spec documented, complete type system, builder API, verifier, AST→MIR lowering for ~80% of constructs, 20+ golden tests, 85%+ coverage

#### 14.1.1: MIR Package Structure and Types (10 tasks)

- [ ] 14.1 Create `mir/` package directory
- [ ] 14.2 Create `mir/types.go` - MIR type system
- [ ] 14.3 Define `Type` interface with `String()`, `Size()`, `Align()` methods
- [ ] 14.4 Implement primitive types: `Bool`, `Int8`, `Int16`, `Int32`, `Int64`, `Float32`, `Float64`, `String`
- [ ] 14.5 Implement composite types: `Array(elemType, size)`, `Record(fields)`, `Pointer(pointeeType)`
- [ ] 14.6 Implement OOP types: `Class(name, fields, methods, parent)`, `Interface(name, methods)`
- [ ] 14.7 Implement function types: `Function(params, returnType)`
- [ ] 14.8 Add `Void` type for procedures
- [ ] 14.9 Implement type equality and compatibility checking
- [ ] 14.10 Implement type conversion rules (explicit vs implicit)

#### 14.1.2: MIR Instructions and Control Flow (10 tasks)

- [ ] 14.11 Create `mir/instruction.go` - MIR instruction set
- [ ] 14.12 Define `Instruction` interface with `ID()`, `Type()`, `String()` methods
- [ ] 14.13 Implement arithmetic ops: `Add`, `Sub`, `Mul`, `Div`, `Mod`, `Neg`
- [ ] 14.14 Implement comparison ops: `Eq`, `Ne`, `Lt`, `Le`, `Gt`, `Ge`
- [ ] 14.15 Implement logical ops: `And`, `Or`, `Xor`, `Not`
- [ ] 14.16 Implement memory ops: `Alloca`, `Load`, `Store`
- [ ] 14.17 Implement constants: `ConstInt`, `ConstFloat`, `ConstString`, `ConstBool`, `ConstNil`
- [ ] 14.18 Implement conversions: `IntToFloat`, `FloatToInt`, `IntTrunc`, `IntExt`
- [ ] 14.19 Implement function ops: `Call`, `VirtualCall`
- [ ] 14.20 Implement array/class ops: `ArrayAlloc`, `ArrayLen`, `ArrayIndex`, `ArraySet`, `FieldGet`, `FieldSet`, `New`

#### 14.1.3: MIR Control Flow Structures (5 tasks)

- [ ] 14.21 Create `mir/block.go` - Basic blocks with `ID`, `Instructions`, `Terminator`
- [ ] 14.22 Implement control flow terminators: `Phi`, `Br`, `CondBr`, `Return`, `Throw`
- [ ] 14.23 Implement terminator validation (every block must end with terminator)
- [ ] 14.24 Implement block predecessors/successors tracking for CFG
- [ ] 14.25 Create `mir/function.go` - Function representation with `Name`, `Params`, `ReturnType`, `Blocks`, `Locals`

#### 14.1.4: MIR Builder API (3 tasks)

- [ ] 14.26 Create `mir/builder.go` - Safe MIR construction
- [ ] 14.27 Implement `Builder` struct with function/block context, `NewFunction()`, `NewBlock()`, `SetInsertPoint()`
- [ ] 14.28 Implement instruction emission methods: `EmitAdd()`, `EmitLoad()`, `EmitStore()`, etc. with type checking

#### 14.1.5: MIR Verifier (2 tasks)

- [ ] 14.29 Create `mir/verifier.go` - MIR correctness checking
- [ ] 14.30 Implement CFG, type, SSA, and function signature verification with `Verify(fn *Function) []error` API

### Phase 15.2: AST → MIR Lowering (12 tasks)

- [ ] 14.31 Create `mir/lower.go` - AST to MIR translation
- [ ] 14.32 Implement `LowerProgram(ast *ast.Program) (*mir.Module, error)` entry point
- [ ] 14.33 Lower expressions: literals → `Const*` instructions
- [ ] 14.34 Lower binary operations → corresponding MIR ops (handle short-circuit for `and`/`or`)
- [ ] 14.35 Lower unary operations → `Neg`, `Not`
- [ ] 14.36 Lower identifier references → `Load` instructions
- [ ] 14.37 Lower function calls → `Call` instructions
- [ ] 14.38 Lower array indexing → `ArrayIndex` + bounds check insertion
- [ ] 14.39 Lower record field access → `FieldGet`/`FieldSet`
- [ ] 14.40 Lower statements: variable declarations, assignments, if/while/for, return
- [ ] 14.41 Lower declarations: functions/procedures, records, classes
- [ ] 14.42 Implement short-circuit evaluation and simple optimizations (constant folding, dead code elimination)

### Phase 15.3: MIR Debugging and Testing (5 tasks)

- [ ] 14.43 Create `mir/dump.go` - Human-readable MIR output with `Dump(fn *Function) string`
- [ ] 14.44 Integration with CLI: `./bin/dwscript dump-mir script.dws`
- [ ] 14.45 Create golden MIR tests: 5+ each for expressions, control flow, functions, advanced features
- [ ] 14.46 Implement MIR verifier tests: type mismatches, malformed CFG, SSA violations
- [ ] 14.47 Implement round-trip tests: AST → MIR → verify → dump → compare with golden files

---

## Phase 16: JavaScript Backend [DEFERRED]

**Status**: Not started | **Priority**: MEDIUM | **Estimated Tasks**: 105 (MVP + feature complete)

### Overview

This phase implements a JavaScript code generator that translates MIR to readable, runnable JavaScript. The backend enables running DWScript programs in browsers and Node.js environments.

**Architecture Flow**: MIR → JavaScript Emitter → JavaScript Code → Node.js/Browser

**Benefits**: Browser execution, npm ecosystem integration, excellent portability, leverages JavaScript JIT compilers

**Dependencies**: Requires Phase 15 (MIR Foundation) to be completed first

### Phase 16.1: JS Backend MVP (45 tasks)

**Goal**: Implement a JavaScript code generator that can compile basic DWScript programs to readable, runnable JavaScript.

**Exit Criteria**: JS emitter for expressions/control flow/functions, 20+ end-to-end tests (DWScript→JS→execute), golden JS snapshots, 85%+ coverage

#### 12.4.1: JS Emitter Infrastructure (8 tasks)

- [ ] 14.48 Create `codegen/` package with `Backend` interface and `EmitterOptions`
- [ ] 14.49 Create `codegen/js/` package and `emitter.go`
- [ ] 14.50 Define `JSEmitter` struct with `out`, `indent`, `opts`, `tmpCounter`
- [ ] 14.51 Implement helper methods: `emit()`, `emitLine()`, `emitIndent()`, `pushIndent()`, `popIndent()`
- [ ] 14.52 Implement `newTemp()` for temporary variable naming
- [ ] 14.53 Implement `NewJSEmitter(opts EmitterOptions)`
- [ ] 14.54 Implement `Generate(module *mir.Module) (string, error)` entry point
- [ ] 14.55 Test emitter infrastructure

#### 14.4.2: Module and Function Emission (6 tasks)

- [ ] 14.56 Implement module structure emission: ES Module format with `export`, file header comment
- [ ] 14.57 Implement optional IIFE fallback via `EmitterOptions`
- [ ] 14.58 Implement function emission: `function fname(params) { ... }`
- [ ] 14.59 Map DWScript params to JS params (preserve names)
- [ ] 14.60 Emit local variable declarations at function top (from `Alloca` instructions)
- [ ] 14.61 Handle procedures (no return value) as JS functions

#### 14.4.3: Expression and Instruction Lowering (12 tasks)

- [ ] 14.62 Lower arithmetic operations → JS infix operators: `+`, `-`, `*`, `/`, `%`, unary `-`
- [ ] 14.63 Lower comparison operations → JS comparisons: `===`, `!==`, `<`, `<=`, `>`, `>=`
- [ ] 14.64 Lower logical operations → JS boolean ops: `&&`, `||`, `!`
- [ ] 14.65 Lower constants → JS literals with proper escaping
- [ ] 14.66 Lower variable operations: `Load` → variable reference, `Store` → assignment
- [ ] 14.67 Lower function calls: `Call` → `functionName(args)`
- [ ] 14.68 Implement Phi node lowering with temporary variables at block edges
- [ ] 14.69 Test expression lowering
- [ ] 14.70 Test instruction lowering
- [ ] 14.71 Test temporary variable generation
- [ ] 14.72 Test type conversions
- [ ] 14.73 Test complex expressions

#### 14.4.4: Control Flow Emission (8 tasks)

- [ ] 14.74 Implement control flow reconstruction from MIR CFG
- [ ] 14.75 Detect if/else patterns from `CondBr`
- [ ] 14.76 Detect while loop patterns (backedge to header)
- [ ] 14.77 Emit if-else: `if (condition) { ... } else { ... }`
- [ ] 14.78 Emit while loops: `while (condition) { ... }`
- [ ] 14.79 Emit for loops if MIR preserves metadata
- [ ] 14.80 Handle unconditional branches
- [ ] 14.81 Handle return statements

#### 14.4.5: Runtime and Testing (11 tasks)

- [ ] 14.82 Create `runtime/js/runtime.js` with `_dws.boundsCheck()`, `_dws.assert()`
- [ ] 14.83 Emit runtime import in generated JS (if needed)
- [ ] 14.84 Make runtime usage optional via `EmitterOptions.InsertBoundsChecks`
- [ ] 14.85 Create `codegen/js/testdata/` with subdirectories
- [ ] 14.86 Implement golden JS snapshot tests
- [ ] 14.87 Setup Node.js in CI (GitHub Actions)
- [ ] 14.88 Implement execution tests: parse → lower → generate → execute → verify
- [ ] 14.89 Add end-to-end tests for arithmetic, control flow, functions, loops
- [ ] 14.90 Add unit tests for JS emitter
- [ ] 14.91 Achieve 85%+ coverage for `codegen/js/` package
- [ ] 14.92 Add `compile-js` CLI command: `./bin/dwscript compile-js input.dws -o output.js`

### Phase 16.2: JS Feature Complete (60 tasks)

**Goal**: Extend JS backend to support all DWScript language features.

**Exit Criteria**: Full OOP, composite types, exceptions, properties, 50+ comprehensive tests, real-world samples work

#### 14.5.1: Records (7 tasks)

- [ ] 14.93 Implement MIR support for records
- [ ] 14.94 Emit records as plain JS objects: `{ x: 0, y: 0 }`
- [ ] 14.95 Implement constructor functions for records
- [ ] 14.96 Implement field access/assignment as property access
- [ ] 14.97 Implement record copy semantics with `_dws.copyRecord()`
- [ ] 14.98 Test record creation, initialization, field read/write
- [ ] 14.99 Test nested records and copy semantics

#### 14.5.2: Arrays (10 tasks)

- [ ] 14.100 Extend MIR for static and dynamic arrays
- [ ] 14.101 Emit static arrays as JS arrays with fixed size
- [ ] 14.102 Implement array index access with optional bounds checking
- [ ] 14.103 Emit dynamic arrays as JS arrays
- [ ] 14.104 Implement `SetLength` → `arr.length = newLen`
- [ ] 14.105 Implement `Length` → `arr.length`
- [ ] 14.106 Support multi-dimensional arrays (nested JS arrays)
- [ ] 14.107 Implement array operations: copy, concatenation
- [ ] 14.108 Test static array creation and indexing
- [ ] 14.109 Test dynamic array operations and bounds checking

#### 14.5.3: Classes and Inheritance (15 tasks)

- [ ] 14.110 Extend MIR for classes with fields, methods, parent, vtable
- [ ] 14.111 Emit ES6 class syntax: `class TAnimal { ... }`
- [ ] 14.112 Implement field initialization in constructor
- [ ] 14.113 Implement method emission
- [ ] 14.114 Implement inheritance with `extends` clause
- [ ] 14.115 Implement `super()` call in constructor
- [ ] 14.116 Handle virtual method dispatch (naturally virtual in JS)
- [ ] 14.117 Handle DWScript `Create` → JS `constructor`
- [ ] 14.118 Handle multiple constructors (overload dispatch)
- [ ] 14.119 Document destructor handling (no direct equivalent in JS)
- [ ] 14.120 Implement static fields and methods
- [ ] 14.121 Map `Self` → `this`, `inherited` → `super.method()`
- [ ] 14.122 Test simple classes with fields and methods
- [ ] 14.123 Test inheritance, virtual method overriding, constructors
- [ ] 14.124 Test static members and `Self`/`inherited` usage

#### 14.5.4: Interfaces (6 tasks)

- [ ] 14.125 Extend MIR for interfaces
- [ ] 14.126 Choose and document JS emission strategy (structural typing vs runtime metadata)
- [ ] 14.127 If using runtime metadata: emit interface tables, implement `is`/`as` operators
- [ ] 14.128 Test class implementing interface
- [ ] 14.129 Test interface method calls
- [ ] 14.130 Test `is` and `as` with interfaces

#### 14.5.5: Enums and Sets (8 tasks)

- [ ] 14.131 Extend MIR for enums
- [ ] 14.132 Emit enums as frozen JS objects: `const TColor = Object.freeze({...})`
- [ ] 14.133 Support scoped and unscoped enum access
- [ ] 14.134 Extend MIR for sets
- [ ] 14.135 Emit small sets (≤32 elements) as bitmasks
- [ ] 14.136 Emit large sets as JS `Set` objects
- [ ] 14.137 Implement set operations: union, intersection, difference, inclusion
- [ ] 14.138 Test enum declaration/usage and set operations

#### 14.5.6: Exception Handling (8 tasks)

- [ ] 14.139 Extend MIR for exceptions: `Throw`, `Try`, `Catch`, `Finally`
- [ ] 14.140 Emit `Throw` → `throw new Error()` or custom exception class
- [ ] 14.141 Emit try-except-finally → JS `try/catch/finally`
- [ ] 14.142 Create DWScript exception class → JS `Error` subclass
- [ ] 14.143 Handle `On E: ExceptionType do` with instanceof checks
- [ ] 14.144 Implement re-raise with exception tracking
- [ ] 14.145 Test basic try-except, multiple handlers, try-finally
- [ ] 14.146 Test re-raise and nested exception handling

#### 14.5.7: Properties and Advanced Features (6 tasks)

- [ ] 14.147 Extend MIR for properties with `PropGet`/`PropSet`
- [ ] 14.148 Emit properties as ES6 getters/setters
- [ ] 14.149 Handle indexed properties as methods
- [ ] 14.150 Test read/write properties and indexed properties
- [ ] 14.151 Implement operator overloading (desugar to method calls)
- [ ] 14.152 Implement generics support (monomorphization)

---

## Phase 17: LLVM Backend [DEFERRED]

**Status**: Not started | **Priority**: LOW | **Estimated Tasks**: 45

### Overview

This phase implements an LLVM IR backend for native code compilation, consolidating all LLVM-related work from the original Phase 13 LLVM JIT and AOT sections. This provides maximum performance but adds significant complexity.

**Architecture Flow**: MIR → LLVM IR Emitter → LLVM IR → llc → Native Binary

**Benefits**: Maximum performance (near C/C++ speed), excellent optimization, multi-architecture support

**Dependencies**: Requires Phase 15 (MIR Foundation) to be completed first

**Note**: This phase consolidates LLVM JIT (from old Phase 13.2), LLVM AOT (from old Phase 13.4), and LLVM backend (from old Stage 14.6). Given complexity and maintenance burden, this is marked as DEFERRED with LOW priority. The bytecode VM and Go AOT provide sufficient performance for most use cases.

### Phase 17.1: LLVM Infrastructure (8 tasks)

**Goal**: Set up LLVM bindings, type mapping, and runtime declarations

- [ ] 14.153 Choose LLVM binding: `llir/llvm` (pure Go) vs CGo bindings
- [ ] 14.154 Create `codegen/llvm/` package with `emitter.go`, `types.go`, `runtime.go`
- [ ] 14.155 Implement type mapping: DWScript types → LLVM types
- [ ] 14.156 Map Integer → `i32`/`i64`, Float → `double`, Boolean → `i1`
- [ ] 14.157 Map String → struct `{i32 len, i8* data}`
- [ ] 14.158 Map arrays/objects to LLVM structs
- [ ] 14.159 Emit LLVM module with target triple
- [ ] 14.160 Declare external runtime functions

### Phase 17.2: Runtime Library (12 tasks)

- [ ] 14.161 Create `runtime/dws_runtime.h` - C header for runtime API
- [ ] 14.162 Declare string operations: `dws_string_new()`, `dws_string_concat()`, `dws_string_len()`
- [ ] 14.163 Declare array operations: `dws_array_new()`, `dws_array_index()`, `dws_array_len()`
- [ ] 14.164 Declare memory management: `dws_alloc()`, `dws_free()`
- [ ] 14.165 Choose and document memory strategy (Boehm GC vs reference counting)
- [ ] 14.166 Declare object operations: `dws_object_new()`, virtual dispatch helpers
- [ ] 14.167 Declare exception handling: `dws_throw()`, `dws_catch()`
- [ ] 14.168 Declare RTTI: `dws_is_instance()`, `dws_as_instance()`
- [ ] 14.169 Create `runtime/dws_runtime.c` - implement runtime
- [ ] 14.170 Implement all runtime functions
- [ ] 14.171 Create `runtime/Makefile` to build `libdws_runtime.a`
- [ ] 14.172 Add runtime build to CI for Linux/macOS/Windows

### Phase 17.3: LLVM Code Emission (15 tasks)

- [ ] 14.173 Implement LLVM emitter: `Generate(module *mir.Module) (string, error)`
- [ ] 14.174 Emit function declarations with correct signatures
- [ ] 14.175 Emit basic blocks for each MIR block
- [ ] 14.176 Emit arithmetic instructions: `add`, `sub`, `mul`, `sdiv`, `srem`
- [ ] 14.177 Emit comparison instructions: `icmp eq`, `icmp slt`, etc.
- [ ] 14.178 Emit logical instructions: `and`, `or`, `xor`
- [ ] 14.179 Emit memory instructions: `alloca`, `load`, `store`
- [ ] 14.180 Emit call instructions: `call @function_name(args)`
- [ ] 14.181 Emit constants: integers, floats, strings
- [ ] 14.182 Emit control flow: conditional branches, phi nodes
- [ ] 14.183 Emit runtime calls for strings, arrays, objects
- [ ] 14.184 Implement type conversions: `sitofp`, `fptosi`
- [ ] 14.185 Emit struct types for classes and vtables
- [ ] 14.186 Implement virtual method dispatch
- [ ] 14.187 Implement exception handling (simple throw/catch or full LLVM EH)

### Phase 17.4: Linking and Testing (7 tasks)

- [ ] 14.188 Implement compilation pipeline: DWScript → MIR → LLVM IR → object → executable
- [ ] 14.189 Integrate `llc` to compile .ll → .o
- [ ] 14.190 Integrate linker to link object + runtime → executable
- [ ] 14.191 Add `compile-native` CLI command
- [ ] 14.192 Create 10+ end-to-end tests: DWScript → native → execute → verify
- [ ] 14.193 Benchmark JS vs native performance
- [ ] 14.194 Document LLVM backend in `docs/llvm-backend.md`

### Phase 17.5: Documentation (3 tasks)

- [ ] 14.195 Create `docs/codegen-architecture.md` - MIR overview, multi-backend design
- [ ] 14.196 Create `docs/mir-spec.md` - complete MIR reference with examples
- [ ] 14.197 Create `docs/js-backend.md` - DWScript → JavaScript mapping guide

---

## Phase 18: WebAssembly AOT Compilation [RECOMMENDED]

**Status**: Not started | **Priority**: MEDIUM-HIGH | **Estimated Tasks**: 5

### Overview

This phase extends WebAssembly support to generate standalone WASM binaries that can run without JavaScript dependency. This builds on Phase 14 (WASM Runtime & Playground) but focuses on ahead-of-time compilation for server-side and edge deployment.

**Architecture Flow**: DWScript Source → Go Compiler → WASM Binary → WASI Runtime (wasmtime, wasmer, wazero)

**Benefits**: Portable binaries, server-side execution, edge computing support, sandboxed execution

**Dependencies**: Builds on Phase 14 (WebAssembly Runtime & Playground)

### Phase 18.1: Standalone WASM Binaries (5 tasks)

- [ ] 18.1 Extend WASM compilation for standalone binaries
  - Generate WASM modules without JavaScript dependency
  - Use WASI for system calls
  - Support WASM-compatible runtime
  - Test with wasmtime, wasmer, wazero

- [ ] 18.2 Optimize WASM binary size
  - Use TinyGo compiler (smaller binaries)
  - Enable wasm-opt optimization
  - Strip unnecessary features
  - Measure binary size (target < 1MB)

- [ ] 18.3 Add WASM runtime support
  - Bundle WASM runtime (wasmer-go or wazero)
  - Create launcher executable
  - Support both JIT and AOT WASM execution
  - Test performance

- [ ] 18.4 Test WASM AOT compilation
  - Compile fixture tests to WASM
  - Run with different WASM runtimes
  - Measure performance vs native
  - Test browser and server execution

- [ ] 18.5 Document WASM AOT
  - Write `docs/wasm-aot.md`
  - Explain WASM compilation process
  - Provide deployment examples
  - Compare with Go native compilation

**Expected Results**: 5-20x speedup (browser), 10-30x speedup (WASI runtime)

---

## Phase 19: AST-Driven Formatter 🆕 **PLANNED**

**Status**: Not started | **Priority**: MEDIUM | **Estimated Tasks**: 30

### Overview

This phase delivers an auto-formatting pipeline that reuses the existing AST and semantic metadata to produce canonical DWScript source, accessible via the CLI (`dwscript fmt`), editors, and the web playground.

**Architecture Flow**: DWScript Source → Parser → AST → Formatter → Formatted DWScript Source

**Benefits**: Consistent code style, automated formatting, editor integration, playground support

### Phase 19.1: Specification & AST/Data Prep (7 tasks)

- [x] 19.1.1 Capture formatting requirements from upstream DWScript (indent width, begin/end alignment, keyword casing, line-wrapping) and document them in `docs/formatter-style-guide.md`.
- [x] 19.1.2 Audit current AST nodes for source position fidelity and comment/trivia preservation; list any nodes lacking `Pos` / `EndPos`.
- [ ] 19.1.3 Extend the parser/AST to track leading and trailing trivia (single-line, block comments, blank lines) without disturbing semantic passes.
- [ ] 19.1.4 Define a `format.Options` struct (indent size, max line length, newline style) and default profile matching DWScript conventions.
- [ ] 19.1.5 Build a formatting test corpus in `testdata/formatter/{input,expected}` with tricky constructs (nested classes, generics, properties, preprocessor).
- [ ] 19.1.6 Add helper APIs to serialize AST back into token streams (e.g., `ast.FormatNode`, `ast.IterChildren`) to keep formatter logic decoupled from parser internals.
- [ ] 19.1.7 Ensure the semantic/type metadata needed for spacing decisions (e.g., `var` params, attributes) is exposed through lightweight inspector interfaces to avoid circular imports.

### Phase 19.2: Formatter Engine Implementation (10 tasks)

- [ ] 19.2.1 Create `formatter` package with a multi-phase pipeline: AST normalization → layout planning → text emission.
- [ ] 19.2.2 Implement a visitor that emits `format.Node` instructions (indent/dedent, soft break, literal text) for statements and declarations, leveraging AST shape rather than raw tokens.
- [ ] 19.2.3 Handle block constructs (`begin...end`, class bodies, `case` arms) with indentation stacks so nested scopes auto-align.
- [ ] 19.2.4 Add expression formatting that respects operator precedence and inserts parentheses only when required; reuse existing precedence tables.
- [ ] 19.2.5 Support alignment for parameter lists, generics, array types, and property declarations with configurable wrap points.
- [ ] 19.2.6 Preserve user comments: attach leading comments before the owning node, keep inline comments after statements, and maintain blank-line intent (max consecutives configurable).
- [ ] 19.2.7 Implement whitespace normalization rules (single spaces around binary operators, before `do`/`then`, after commas, etc.).
- [ ] 19.2.8 Provide idempotency guarantees by building a golden test that pipes formatted output back through the formatter and asserts stability.
- [ ] 19.2.9 Expose a streaming writer that emits `[]byte`/`io.Writer` output to keep the CLI fast and low-memory.
- [ ] 19.2.10 Benchmark formatting of large fixtures (≥5k LOC) and optimize hot paths (string builder pools, avoiding interface allocations).

### Phase 19.3: Tooling & Playground Integration (7 tasks)

- [ ] 19.3.1 Wire a new CLI command `dwscript fmt` (and `fmt -w`) that runs the formatter over files/directories, mirroring `gofmt` UX.
- [ ] 19.3.2 Update the WASM bridge to expose a `Format(source string) (string, error)` hook exported from Go, reusing the same formatter package.
- [ ] 19.3.3 Modify `playground/js/playground.js` to call the WASM formatter before falling back to Monaco’s default action, enabling deterministic formatting in the browser.
- [ ] 19.3.4 Add formatter support to the VSCode extension / LSP stub (if present) so editors can trigger `textDocument/formatting`.
- [ ] 19.3.5 Ensure the formatter respects partial-range requests (`textDocument/rangeFormatting`) to avoid reformatting entire files when not desired.
- [ ] 19.3.6 Introduce CI checks (`just fmt-check`) that fail when files are not formatted, and document the workflow in `CONTRIBUTING.md`.
- [ ] 19.3.7 Provide sample scripts/snippets (e.g., Git hooks) encouraging contributors to run the formatter.

### Phase 19.4: Validation, UX, and Docs (6 tasks)

- [ ] 19.4.1 Create table-driven unit tests per node type plus integration tests that read `testdata/formatter` fixtures.
- [ ] 19.4.2 Add fuzz/property tests that compare formatter output against itself round-tripped through the parser → formatter pipeline.
- [ ] 19.4.3 Document formatter architecture and extension points in `docs/formatter-architecture.md`.
- [ ] 19.4.4 Update `PLAYGROUND.md`, `README.md`, and release notes to mention the Format button now runs the AST-driven formatter.
- [ ] 19.4.5 Record known limitations (e.g., preprocessor directives) and track follow-ups in `TEST_ISSUES.md`.
- [ ] 19.4.6 Gather usability feedback (issue template or telemetry) to prioritize refinements like configurable styles or multi-profile support.

---

## Phase 21: Future Enhancements & Experimental Features [LONG-TERM]

**Status**: Not started | **Priority**: LOW | **Tasks**: Variable

### Overview

This phase collects experimental, deferred, and long-term enhancement tasks that are not critical to the core DWScript implementation but may provide value in specific use cases or future development.

**Note**: Tasks in this phase are marked as DEFERRED or OPTIONAL and should only be pursued after core phases are complete and based on user demand.

### Phase 21.1: Plugin-Based JIT [SKIP - Poor Portability]

**Status**: SKIP RECOMMENDED | **Limitation**: No Windows support, requires Go toolchain at runtime

- [ ] 21.1 Implement Go code generation from AST
  - Create `internal/codegen/go_generator.go`
  - Generate Go source code from DWScript AST
  - Map DWScript types to Go types
  - Generate function declarations and calls
  - Handle closures and scoping

- [ ] 21.2 Implement plugin-based JIT
  - Use `go build -buildmode=plugin` to compile generated code
  - Load plugin with `plugin.Open()`
  - Look up compiled function with `plugin.Lookup()`
  - Call compiled function from interpreter
  - Cache plugins to disk

- [ ] 21.3 Add hot path detection for plugin JIT
  - Track function execution counts
  - Trigger plugin compilation for hot functions
  - Manage plugin lifecycle (loading/unloading)

- [ ] 21.4 Test plugin-based JIT
  - Run tests on Linux and macOS only
  - Compare performance with bytecode VM
  - Test plugin caching and reuse

- [ ] 21.5 Document plugin approach
  - Write `docs/plugin-jit.md`
  - Explain platform limitations
  - Provide usage examples

**Expected Results**: 3-5x faster than tree-walking
**Recommendation**: SKIP - poor portability, requires Go toolchain

### Phase 21.2: Alternative Compiler Targets [EXPERIMENTAL]

- [ ] 21.6 C code generation backend
  - Transpile DWScript to C
  - Leverage existing C compilers
  - Test on embedded systems

- [ ] 21.7 Rust code generation backend
  - Transpile DWScript to Rust
  - Leverage Rust's memory safety
  - Explore performance characteristics

- [ ] 21.8 Python code generation backend
  - Transpile DWScript to Python
  - Enable rapid prototyping
  - Integration with Python ecosystem

### Phase 21.3: Advanced Optimization Research [EXPERIMENTAL]

- [ ] 21.9 Profile-guided optimization (PGO)
  - Collect runtime profiles
  - Use profiles to guide optimizations
  - Measure performance improvements

- [ ] 21.10 Speculative optimization
  - Type speculation based on runtime behavior
  - Deoptimization on type changes
  - Guard conditions

- [ ] 21.11 Escape analysis
  - Determine when objects can be stack-allocated
  - Reduce GC pressure
  - Improve performance

- [ ] 21.12 Inline caching for dynamic dispatch
  - Cache method lookup results
  - Invalidate on class changes
  - Measure performance impact

### Phase 21.4: Language Extensions [EXPERIMENTAL]

- [ ] 21.13 Async/await support
  - Design async/await syntax for DWScript
  - Implement coroutine-based execution
  - Test with concurrent workloads

- [ ] 21.14 Pattern matching
  - Extend case statements with pattern matching
  - Support destructuring
  - Type narrowing based on patterns

- [ ] 21.15 Macros/metaprogramming
  - Design macro system
  - Compile-time code generation
  - Template metaprogramming support

### Phase 21.5: Tooling Enhancements [LOW PRIORITY]

- [ ] 21.16 IDE integration beyond LSP
  - IntelliJ IDEA plugin
  - VS Code enhanced extension
  - Sublime Text package

- [ ] 21.17 Package manager
  - Design package format
  - Implement dependency resolution
  - Create package registry

- [ ] 21.18 Build system integration
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
21. **Phase 21 – Future Enhancements**: [LONG-TERM] Experimental features, alternative targets

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
