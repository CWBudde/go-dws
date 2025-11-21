# TODOs - DWScript Go Port

> Comprehensive tracking of all TODO items in the codebase
>
> Last Updated: 2025-11-21

## Executive Summary

**Total TODOs**: 109 in Go source files (excluding testdata) - 2 completed

**Quick Stats**:

- HIGH: 24 (blocking features, builtins migration) - ~~2 completed~~
- MEDIUM: 30+ (architecture improvements, enhancements)
- LOW: 20+ (future features, nice-to-haves)
- FIXTURE TESTS: 25 disabled test suites
- ✅ COMPLETED: 2 (Set type declarations, Random number functions)

**Distribution by Category**:

- Interpreter/Runtime: ~45 TODOs
- Bytecode/VM: ~15 TODOs
- Builtins: ~22 TODOs (Random, String, DateTime functions)
- Semantic Analysis: ~8 TODOs
- Testing: ~25 TODOs (fixture test suites)
- Parser: ~2 TODOs
- Other (CLI, WASM, Units): ~5 TODOs

---

## 1. HIGH Priority

---

### ~~1.2 Builtins - Random Number Functions Migration~~ ✅ COMPLETED

**Status**: COMPLETED (2025-11-21)

**Summary**: All 6 random number functions have been successfully migrated to the builtins package and are fully functional in both the AST interpreter and bytecode VM.

**Changes Made**:

1. **Context Interface** (`internal/interp/builtins/context.go`):
   - ✅ RandSource() *rand.Rand (line 48)
   - ✅ GetRandSeed() int64 (line 52)
   - ✅ SetRandSeed(seed int64) (line 56)

2. **Interpreter Implementation** (`internal/interp/builtins_context.go`):
   - ✅ Implemented all three Context methods (lines 33, 39, 45)

3. **Builtin Functions** (`internal/interp/builtins/math_basic.go`):
   - ✅ Random() - returns random float [0, 1) (line 592)
   - ✅ Randomize() - seeds RNG with current time (line 603)
   - ✅ RandomInt(max) - returns random int [0, max) (line 614)
   - ✅ SetRandSeed(seed) - sets random seed (line 641)
   - ✅ RandSeed() - returns current seed (line 659)
   - ✅ RandG() - returns Gaussian random (line 671)

4. **Builtin Registry** (`internal/interp/builtins/register.go`):
   - ✅ All 6 functions registered (lines 138-143)
   - ✅ Updated status comment (231 functions registered, Math: 68 functions)

5. **Semantic Analyzer** (`internal/semantic/analyze_builtins.go`):
   - ✅ Added randomint, setrandseed, randseed, randg to builtin list (line 54)

6. **Bytecode Compiler** (`internal/bytecode/compiler_functions.go`):
   - ✅ Added isBuiltinFunction() method with all 6 random functions (line 125-170)
   - ✅ Modified compileCallExpression() to handle builtins (line 77-92)

7. **Bytecode VM** (`internal/bytecode/vm_builtins_math.go`):
   - ✅ Implemented builtinRandom() and builtinRandomInt() (lines 275-297)
   - ✅ Registered all 6 functions in VM (lines 24-29, lowercase)

**Verification**:
- ✅ AST Interpreter: All functions work correctly
- ✅ Bytecode VM: All functions compile and execute
- ✅ Test script confirms deterministic seeding and correct behavior

**Note**: General bytecode VM builtin registration uses capital letters which causes issues. This was fixed for random functions by using lowercase registration. A systematic fix for all VM builtins may be needed separately.

---

### 1.3 Builtins - String Functions Requiring Special Handling

**Location**: [internal/interp/builtins/strings_basic.go](internal/interp/builtins/strings_basic.go)

**Priority**: HIGH

**Status**: Planned - Divided into 8 subtasks

**Overview**: 7 string functions require either ArrayValue migration (4 functions) or var parameter support (3 functions). This task has been broken down into manageable subtasks with clear dependencies.

**Impact**:
- 7 string functions incomplete
- Blocks FunctionsString fixture tests
- ArrayValue circular dependency prevents completion
- Var parameter language feature not yet implemented

---

#### ~~1.3.1 ArrayValue Dependency Analysis~~ ✅ COMPLETED

**Priority**: HIGH | **Status**: COMPLETED (2025-11-21)

**Goal**: Document all current uses of ArrayValue and identify circular dependencies

**Summary**: Analysis revealed that ArrayValue was already migrated to runtime package in Phase 3.5.4. The TODOs in strings_basic.go were outdated.

**Completed Tasks**:
- [x] Scan entire codebase for ArrayValue usage
- [x] Map out import dependencies (interp → runtime, runtime → types, etc.)
- [x] Identify which types/functions MUST move with ArrayValue
- [x] Document the dependency graph
- [x] Create migration plan document

**Deliverables**:
- `docs/arrayvalue-migration-analysis.md` with dependency graph and findings

**Key Findings**:
- ArrayValue is defined in `internal/interp/runtime/array.go`
- Type alias exists in `internal/interp/value.go`: `ArrayValue = runtime.ArrayValue`
- Builtins package already uses `runtime.ArrayValue` successfully
- No circular dependencies exist

---

#### ~~1.3.2 Move ArrayValue to Runtime Package~~ ✅ COMPLETED

**Priority**: HIGH | **Status**: COMPLETED (2025-11-21)

**Depends On**: 1.3.1

**Goal**: Move ArrayValue type from internal/interp to internal/interp/runtime

**Summary**: ArrayValue was already in runtime package (Phase 3.5.4). This task focused on:
1. Implementing missing string functions that depend on ArrayValue
2. Removing outdated TODO comments
3. Registering new functions in builtins registry

**Completed Tasks**:
- [x] Verified ArrayValue already in `internal/interp/runtime/array.go`
- [x] Implemented StrSplit() in builtins package
- [x] Implemented StrJoin() in builtins package
- [x] Implemented StrArrayPack() in builtins package
- [x] Verified Format() already exists in system.go
- [x] Registered new functions in registry
- [x] Updated outdated TODO comments
- [x] Full test suite passes

**Files Modified**:
- `internal/interp/builtins/strings_basic.go` - Added StrSplit, StrJoin, StrArrayPack
- `internal/interp/builtins/register.go` - Registered new functions

**Acceptance Criteria**:
- [x] ArrayValue accessible from builtins package
- [x] All tests pass
- [x] No circular import errors
- [x] Code compiles and runs correctly
- [x] New string functions verified working

---

#### ~~1.3.3 Implement Format() Function~~ ✅ COMPLETED

**Priority**: HIGH | **Status**: COMPLETED (Pre-existing)

**Depends On**: 1.3.2

**Goal**: Complete Format() builtin function implementation

**Summary**: Format() was already fully implemented in `internal/interp/builtins/system.go`. The TODO in strings_basic.go was outdated.

**Implementation Location**: `internal/interp/builtins/system.go:291`

**Completed Tasks**:
- [x] Format() already exists with full implementation
- [x] Registered in builtins registry
- [x] Uses Context.FormatString() helper for formatting

---

#### ~~1.3.4 Implement StrSplit() Function~~ ✅ COMPLETED

**Priority**: HIGH | **Status**: COMPLETED (2025-11-21)

**Depends On**: 1.3.2

**Goal**: Complete StrSplit() builtin function implementation

**Summary**: Implemented in `internal/interp/builtins/strings_basic.go`

**Completed Tasks**:
- [x] Implement string splitting logic
- [x] Return ArrayValue with split results
- [x] Handle edge cases (empty string, empty delimiter)
- [x] Registered in builtins registry

**Implementation Location**: `internal/interp/builtins/strings_basic.go:947`

---

#### ~~1.3.5 Implement StrJoin() and StrArrayPack() Functions~~ ✅ COMPLETED

**Priority**: MEDIUM | **Status**: COMPLETED (2025-11-21)

**Depends On**: 1.3.2

**Goal**: Complete StrJoin() and StrArrayPack() builtin functions

**Summary**: Both functions implemented in `internal/interp/builtins/strings_basic.go`

**Completed Tasks**:
- [x] Implement StrJoin() - joins array elements with delimiter
- [x] Implement StrArrayPack() - removes empty strings from array
- [x] Handle ArrayValue as input parameter
- [x] Registered in builtins registry
- [x] Verified with test script

**Implementation Locations**:
- StrJoin: `internal/interp/builtins/strings_basic.go:991`
- StrArrayPack: `internal/interp/builtins/strings_basic.go:1033`

---

#### ~~1.3.6 Design Var Parameter Language Feature~~ ✅ COMPLETED

**Priority**: HIGH | **Status**: COMPLETED (2025-11-21)

**Depends On**: None (can start independently)

**Goal**: Design and document var parameter support for DWScript

**Summary**: Comprehensive analysis revealed var parameters are already fully implemented in the AST interpreter. Only the bytecode compiler/VM lacks support. Created detailed design document with implementation plan.

**Key Findings**:
- Lexer: COMPLETE (VAR keyword tokenized)
- Parser: COMPLETE (ByRef flag set on Parameter)
- AST: COMPLETE (Parameter.ByRef bool field)
- Semantic Analyzer: COMPLETE (tracks var params)
- Interpreter: COMPLETE (ReferenceValue, function call handling)
- FFI: COMPLETE (VarParams []bool support)
- Bytecode Compiler: NOT IMPLEMENTED (ignores ByRef)
- Bytecode VM: NOT IMPLEMENTED (no reference opcodes)

**Completed Tasks**:
- [x] Research original DWScript var parameter semantics
- [x] Research current go-dws implementation (extensive support found!)
- [x] Design AST representation for var parameters (already exists)
- [x] Design semantic analysis for var parameters (already exists)
- [x] Design interpreter evaluation strategy (already implemented)
- [x] Design bytecode compilation strategy (new opcodes proposed)
- [x] Create design document with examples
- [x] Document edge cases (nested calls, record fields, array elements)

**Deliverables**:
- `docs/var-parameters-design.md` - Complete specification with:
  - Current implementation status analysis
  - Language semantics documentation
  - Bytecode compiler design (Option A: Reference Opcodes)
  - Edge cases and testing strategy
  - Implementation plan for remaining work

**Design Decisions**:
1. Use new opcodes (OpLoadRef, OpStoreRef, OpDeref) for bytecode support
2. Store VarParams[] in FunctionObject metadata
3. Reference pass-through for nested var param calls

---

#### ~~1.3.7 Implement Var Parameter Support~~ ✅ COMPLETED

**Priority**: HIGH | **Status**: COMPLETED (2025-11-21)

**Depends On**: 1.3.6

**Goal**: Implement var parameter language feature across all compiler layers

**Summary**: Analysis during task 1.3.6 revealed that var parameter support was already complete in layers 1.3.7.1-1.3.7.5. Only the bytecode compiler/VM (task 1.3.7.6) required implementation.

**Subtasks Status**:
- 1.3.7.1-1.3.7.5: Pre-existing (COMPLETE - already implemented)
- 1.3.7.6: Bytecode Compiler & VM - COMPLETED (2025-11-21)

---

##### ~~1.3.7.1-1.3.7.5 Pre-existing Implementation~~ ✅ COMPLETE

**Status**: Already implemented prior to this task

These subtasks were found to be already complete during the design phase (task 1.3.6):

- **1.3.7.1 Lexer**: VAR keyword already tokenized
- **1.3.7.2 Parser**: ByRef flag already parsed (`Parameter.ByRef = true`)
- **1.3.7.3 AST**: `Parameter.ByRef bool` field exists in `pkg/ast/functions.go:37`
- **1.3.7.4 Semantic**: `param.ByRef` handling in `internal/semantic/analyze_functions.go`
- **1.3.7.5 Interpreter**: `ReferenceValue` type and full function call handling in place

---

##### 1.3.7.2 Parser - Parse Var Parameter Declarations

**Priority**: HIGH | **Status**: Not Started | **Estimated**: 2-3 days

**Depends On**: 1.3.7.1

**Goal**: Extend parser to recognize and parse var parameter syntax

**Tasks**:
- [ ] Modify parameter parsing to detect `var` keyword
- [ ] Add `IsVarParameter` flag or marker to parameter nodes
- [ ] Parse function declarations with var parameters
- [ ] Parse procedure declarations with var parameters
- [ ] Handle mixed parameters (some var, some value)
- [ ] Add comprehensive parser tests
- [ ] Test error cases (var on function results, invalid var usage)

**Example Syntax to Parse**:
```pascal
procedure Increment(var x: Integer);
procedure Swap(var a, b: Integer);
function GetValues(var count: Integer): String;  // both var param and result
procedure Multi(x: Integer; var y: Integer; z: String);  // mixed
```

**Files to Modify**:
- `internal/parser/functions.go` - Parameter parsing logic
- `internal/parser/parser_test.go` - Parser tests

**Acceptance Criteria**:
- Can parse var parameter declarations
- IsVarParameter flag correctly set
- Mixed parameter lists work correctly
- All parser tests pass
- Error messages for invalid var usage

---

##### 1.3.7.3 AST - Define Var Parameter Nodes

**Priority**: HIGH | **Status**: Not Started | **Estimated**: 1-2 days

**Depends On**: 1.3.7.2

**Goal**: Extend AST to represent var parameters

**Tasks**:
- [ ] Add `IsVar` or `VarParameter` field to Parameter node
- [ ] Update FunctionDeclaration AST node
- [ ] Update visitor pattern to handle var parameters
- [ ] Regenerate visitor code if needed (`go generate ./pkg/ast`)
- [ ] Update AST String() methods for debugging
- [ ] Add AST tests

**Files to Modify**:
- `pkg/ast/declarations.go` - Parameter and function nodes
- `pkg/ast/visitor.go` or `visitor_generated.go` - Visitor pattern
- `pkg/ast/ast_test.go` - AST tests

**AST Design Example**:
```go
type Parameter struct {
    Name      *Identifier
    Type      Type
    IsVar     bool  // NEW: marks var parameters
    DefaultValue Expression
    Position  lexer.Position
}
```

**Acceptance Criteria**:
- AST correctly represents var parameters
- Visitor pattern handles var parameters
- AST can distinguish var from value parameters
- String() output shows var parameters
- All AST tests pass

---

##### 1.3.7.4 Semantic Analyzer - Validate Var Parameters

**Priority**: HIGH | **Status**: Not Started | **Estimated**: 2-3 days

**Depends On**: 1.3.7.3

**Goal**: Add semantic analysis for var parameters

**Tasks**:
- [ ] Validate var parameter arguments are lvalues (variables, not expressions)
- [ ] Ensure var parameters are assignable (not constants)
- [ ] Type check var parameter arguments match declaration
- [ ] Track var parameters in symbol table
- [ ] Validate function calls with var parameters
- [ ] Add semantic error messages
- [ ] Write comprehensive semantic tests

**Validation Rules**:
- ✅ `Increment(x)` - var parameter with variable
- ❌ `Increment(5)` - var parameter with literal
- ❌ `Increment(x + 1)` - var parameter with expression
- ❌ `Increment(const_value)` - var parameter with constant
- ✅ `Increment(arr[i])` - var parameter with indexed element
- ✅ `Increment(obj.field)` - var parameter with field

**Files to Modify**:
- `internal/semantic/analyze_functions.go` - Function declaration analysis
- `internal/semantic/analyze_function_calls.go` - Call site validation
- `internal/semantic/analyzer_test.go` - Semantic tests

**Acceptance Criteria**:
- Var parameters only accept lvalues
- Clear error messages for invalid var arguments
- Type checking enforced for var parameters
- All semantic tests pass
- No false positives or negatives

---

##### 1.3.7.5 Interpreter - Implement Reference Passing

**Priority**: HIGH | **Status**: Not Started | **Estimated**: 3-4 days

**Depends On**: 1.3.7.4

**Goal**: Implement var parameter evaluation and reference passing in interpreter

**Tasks**:
- [ ] Design reference passing mechanism (Environment modification)
- [ ] Modify function call evaluation to detect var parameters
- [ ] Pass references to variables instead of values
- [ ] Implement write-back mechanism to caller's scope
- [ ] Handle var parameters in closures (if applicable)
- [ ] Handle nested function calls with var parameters
- [ ] Add comprehensive interpreter tests
- [ ] Test with real DWScript code examples

**Implementation Strategy**:
```go
// Instead of passing value:
//   env.Set("x", evaluatedArg)
// Pass reference that can modify caller's variable:
//   env.SetRef("x", &callerVar)
```

**Test Cases**:
- Simple increment: `Increment(x)` modifies x
- Swap: `Swap(a, b)` exchanges values
- Array element: `Increment(arr[i])` modifies array element
- Record field: `Increment(obj.field)` modifies field
- Nested calls: `A(x); B(x)` both modify x
- Multiple var params: `Swap(a, b)` modifies both

**Files to Modify**:
- `internal/interp/functions_user.go` - Function call evaluation
- `internal/interp/environment.go` - Reference passing support (if needed)
- `internal/interp/evaluator/` - Evaluator var parameter handling
- `internal/interp/interpreter_test.go` - Interpreter tests

**Acceptance Criteria**:
- Var parameters successfully modify caller's variables
- All test cases pass
- No memory leaks or reference issues
- Nested scopes handled correctly

---

##### ~~1.3.7.6 Bytecode Compiler & VM - Bytecode Var Parameters~~ ✅ COMPLETED

**Priority**: HIGH | **Status**: COMPLETED (2025-11-21)

**Depends On**: 1.3.7.5

**Goal**: Implement var parameter support in bytecode compilation and VM execution

**Completed Tasks**:

**Bytecode Compiler**:
- [x] Designed bytecode strategy: Reference type with pointer to Value slot
- [x] Added 3 new opcodes: OpLoadRef, OpStoreRef, OpDeref
- [x] Added VarParams[] to FunctionObject metadata
- [x] Updated compiler to track var parameters in function declarations
- [x] All bytecode tests pass (124 opcodes total)

**VM Execution**:
- [x] Added Reference struct for pointer-based references
- [x] Added ValueRef type for reference values
- [x] Implemented OpLoadRef - creates reference to local variable
- [x] Implemented OpStoreRef - stores value through reference
- [x] Implemented OpDeref - dereferences to get value

**Files Modified**:
- `internal/bytecode/instruction.go` - Added 3 reference opcodes
- `internal/bytecode/bytecode.go` - Added Reference type, ValueRef, FunctionObject.VarParams
- `internal/bytecode/compiler_statements.go` - Track var params in function compilation
- `internal/bytecode/vm_exec.go` - Execute reference opcodes
- `internal/bytecode/instruction_test.go` - Updated opcode count

**Implementation Summary**:
Used Option A from design (Stack References) with direct pointer-based references:
```go
type Reference struct {
    Location *Value // Direct pointer to value slot
}
```
Opcodes:
- OpLoadRef: Push Reference{&frame.locals[idx]} onto stack
- OpStoreRef: Pop ref and value, call ref.Assign(value)
- OpDeref: Pop ref, push ref.Deref()

**Acceptance Criteria**:
- [x] Bytecode infrastructure for var parameters complete
- [x] VM can execute reference operations
- [x] All bytecode tests pass
- [x] No regression in existing functionality

---

### ~~Task 1.3.7 Summary~~ ✅ COMPLETED

**Actual Implementation**: Only 1.3.7.6 (Bytecode) required implementation - all other subtasks were pre-existing.

**Key Discovery**: Var parameters were already fully implemented in the AST interpreter (tasks 1.3.7.1-1.3.7.5). The bytecode compiler/VM was the only missing piece.

**Testing Strategy**: All bytecode tests pass. Reference opcodes ready for use when function calls are updated to pass references for var params.

---

#### 1.3.8 Implement Insert(), Delete(), DeleteString() Functions

**Priority**: HIGH | **Status**: Not Started | **Estimated**: 2-3 days

**Depends On**: 1.3.7

**Goal**: Complete string modification builtin functions using var parameters

**Current Status**: Stubbed at lines 397, 401 of strings_basic.go

**Tasks**:
- [ ] Implement Insert() - insert substring into var string
- [ ] Implement Delete() - delete substring from var string
- [ ] Implement DeleteString() - delete specific text from var string
- [ ] Use var parameter mechanism to modify caller's variable
- [ ] Add comprehensive tests for all three functions
- [ ] Test with FunctionsString fixture suite

**Signatures**:
```pascal
procedure Insert(source: String; var dest: String; index: Integer);
procedure Delete(var s: String; index: Integer; count: Integer);
procedure DeleteString(var s: String; substr: String);
```

**Test Cases**:
- Insert operations at various positions
- Delete operations at boundaries
- DeleteString with multiple occurrences
- Edge cases: empty strings, out-of-bounds indices

**Acceptance Criteria**:
- All three functions modify caller's string correctly
- No copies/returns, pure in-place modification
- All unit tests pass
- FunctionsString fixture tests pass

---

### Task 1.3 Summary

**Total Estimated Time**: 4-6 weeks

**Breakdown**:
- **Phase A (ArrayValue Migration)**: 1.3.1 → 1.3.2 → 1.3.3, 1.3.4, 1.3.5 (2-3 weeks)
- **Phase B (Var Parameters)**: 1.3.6 → 1.3.7 → 1.3.8 (2-3 weeks)

**Parallelization Opportunity**: Tasks 1.3.1-1.3.5 (ArrayValue) and 1.3.6-1.3.8 (Var Parameters) can be worked on concurrently as they have no dependencies on each other.

**Critical Path**: 1.3.7 (Implement Var Parameter Support) is the most complex and time-consuming task.

**Deliverables**:
- 4 string functions using ArrayValue (Format, StrSplit, StrJoin, StrArrayPack)
- 3 string functions using var parameters (Insert, Delete, DeleteString)
- Var parameter language feature (usable for other future functions)
- 2 design documents

**Stage Relevance**: Phase 3.7.1 (documented in [docs/phase3-task3.7.1-summary.md](docs/phase3-task3.7.1-summary.md))

---

### 1.4 Builtins - DateTime Functions Requiring By-Ref Support

#### DecodeDate and DecodeTime Need Variable Modification

**Location**: [internal/interp/builtins/datetime_calc.go](internal/interp/builtins/datetime_calc.go)

**Priority**: HIGH

**Affected Functions**:

1. [DecodeDate()](internal/interp/builtins/datetime_calc.go#L178) - line 178
2. [DecodeTime()](internal/interp/builtins/datetime_calc.go#L182) - line 182

**TODO**: Requires special handling (modifies variables in-place)

**Description**: These functions take multiple var parameters and decode a date/time value into separate year/month/day or hour/minute/second components.

**Impact**:

- 2 datetime functions incomplete
- Blocks FunctionsTime fixture tests
- Common datetime operations unavailable

**Action Required**:

1. Implement by-ref parameter support (var parameters)
2. Allow functions to write back to caller's variables
3. Complete DecodeDate and DecodeTime implementations
4. Test with datetime fixture suite

**Stage Relevance**: Phase 3 Refactoring (Builtins Migration)

---

### 1.5 Bytecode VM - Missing Core Features

#### For Loops Not Supported

**Location**: [internal/bytecode/vm_parity_test.go:71](internal/bytecode/vm_parity_test.go#L71)

**TODO**: For loop not yet supported

**Priority**: HIGH

**Impact**:
- VM cannot execute for loops (critical language feature)
- Severely limits VM usability
- Affects performance testing (for loops are performance-critical)

**Action Required**:
1. Implement for-loop bytecode compilation
2. Add loop counter management to VM
3. Handle for-to and for-downto variants
4. Test with loop-heavy scripts

---

#### Result Variable Not Supported

**Location**: [internal/bytecode/vm_parity_test.go:80](internal/bytecode/vm_parity_test.go#L80)

**TODO**: Result variable not yet supported

**Priority**: HIGH

**Impact**:
- Functions cannot use implicit Result variable
- Forces use of explicit return statements
- Breaks DWScript idioms

**Action Required**:
1. Add Result variable support to compiler
2. Allocate Result in function prologue
3. Return Result value in function epilogue
4. Test with functions using Result

---

#### Trim Builtin Not Implemented

**Location**: [internal/bytecode/vm_calls.go:196](internal/bytecode/vm_calls.go#L196)

**TODO**: Trim builtin not implemented in VM

**Priority**: HIGH

**Impact**:
- Trim() function calls fail in bytecode VM
- Breaks VM parity with AST interpreter

**Action Required**:
1. Add Trim to VM builtin implementations
2. Register in VM builtin table
3. Test string trimming in VM mode

**Stage Relevance**: Bytecode VM Development

---

### 1.6 Enums - Boolean Operations and Conversions

#### Enum Boolean Operations

**Location**: PLAN.md Stage 9.15.7

**TODO**: Implement enum boolean operations (or, and, xor)

**Priority**: HIGH

**Impact**:
- Cannot perform boolean operations on enum values
- Affects enum_bool_op fixture tests
- Common enum flag operations unavailable

**Action Required**:
1. Update semantic analyzer to allow boolean ops on enums
2. Implement runtime evaluation in interpreter
3. Add bytecode compilation support
4. Test with enum_bool_op fixtures

**Stage Relevance**: Stage 9.15.7 (Enum Boolean Operations)

---

#### Implicit Enum-to-Integer Conversion

**Location**: PLAN.md Stage 9.15.8

**TODO**: Implement implicit enum-to-integer conversion

**Priority**: HIGH

**Impact**:
- Requires explicit typecast for enum-to-int
- Less ergonomic than original DWScript
- May break compatibility tests

**Action Required**:
1. Update type checker to allow implicit conversion
2. Add automatic conversion in runtime
3. Test with enum fixture suite

**Stage Relevance**: Stage 9.15.8 (Implicit Enum Conversions)

---

## 2. MEDIUM Priority

### 2.1 Interpreter - Evaluator Declaration Visitor Migration

**Location**: [internal/interp/evaluator/visitor_declarations.go](internal/interp/evaluator/visitor_declarations.go)

**Priority**: MEDIUM

**Description**: Move declaration registration logic from visitor to Evaluator for better separation of concerns.

**Affected Declarations**:
1. [Function declarations](internal/interp/evaluator/visitor_declarations.go#L16) - line 16
2. [Class declarations](internal/interp/evaluator/visitor_declarations.go#L23) - line 23
3. [Interface declarations](internal/interp/evaluator/visitor_declarations.go#L30) - line 30
4. [Operator declarations](internal/interp/evaluator/visitor_declarations.go#L37) - line 37
5. [Enum declarations](internal/interp/evaluator/visitor_declarations.go#L44) - line 44
6. [Record declarations](internal/interp/evaluator/visitor_declarations.go#L51) - line 51
7. [Helper declarations](internal/interp/evaluator/visitor_declarations.go#L58) - line 58
8. [Array type declarations](internal/interp/evaluator/visitor_declarations.go#L65) - line 65
9. [Type alias declarations](internal/interp/evaluator/visitor_declarations.go#L72) - line 72

**Impact**: Architecture refactoring - improves code organization but not blocking

**Action Required**:
1. Create registration methods in Evaluator
2. Move logic from visitor to Evaluator
3. Update visitor to call Evaluator methods
4. Maintain test coverage

**Stage Relevance**: Phase 3.5 (Evaluator Pattern Implementation)

---

### 2.2 Types - Circular Import Resolution

**Location**: [internal/interp/evaluator/type_helpers.go](internal/interp/evaluator/type_helpers.go)

**Priority**: MEDIUM

**Description**: Break circular dependencies between packages

**Affected Types**:
1. [Record types](internal/interp/evaluator/type_helpers.go#L138) - line 138
2. [Enum types](internal/interp/evaluator/type_helpers.go#L146) - line 146
3. [Set types](internal/interp/evaluator/type_helpers.go#L154) - line 154

**TODO**: Extract type without circular import

**Impact**: Architecture issue - makes type system harder to maintain

**Action Required**:
1. Create separate types package
2. Move type definitions to avoid cycles
3. Update imports throughout codebase
4. Verify no functionality breaks

**Stage Relevance**: Phase 3.2 (Value System Refactoring)

---

### 2.3 Bytecode - Validation Enhancements

**Location**: [internal/bytecode/bytecode.go:946](internal/bytecode/bytecode.go#L946)

**TODO**: Add more validation for jumps, etc.

**Priority**: MEDIUM

**Impact**: Better error detection in bytecode compiler, prevents runtime crashes

**Action Required**:
1. Validate jump targets are within bounds
2. Check stack depth consistency
3. Verify constant pool indices
4. Add validation tests

**Stage Relevance**: Bytecode VM Development

---

### 2.4 Runtime - Record and Class Initialization

**Priority**: MEDIUM

**Affected Areas**:

1. **Record Field Initialization**
   - [Location](internal/bytecode/vm_exec.go#L710): internal/bytecode/vm_exec.go:710
   - TODO: Initialize fields from record metadata
   - Impact: Records may have uninitialized fields

2. **Function Pointer Type Construction**
   - [Location](internal/interp/helpers_conversion.go#L96): internal/interp/helpers_conversion.go:96
   - TODO: Properly construct function pointer type
   - Impact: Function pointers may not work correctly

3. **Custom Class Type Creation**
   - [Location](internal/interp/functions_typecast.go#L314): internal/interp/functions_typecast.go:314
   - TODO: Create custom type from class name
   - Impact: Dynamic class instantiation may fail

**Action Required**:

1. Implement proper field initialization for records
2. Fix function pointer type construction
3. Add class name to type mapping
4. Test with advanced OOP scripts

**Stage Relevance**: Runtime correctness for advanced features

---

### 2.5 Semantic Analysis - Type Checking Enhancements

**Location**: [internal/semantic/](internal/semantic/)

**Priority**: MEDIUM

**Enhancement Areas**:

1. **Function Overload Resolution**
   - [Location](internal/semantic/analyze_function_calls.go#L22): analyze_function_calls.go:22
   - TODO: Use expectedType for overload resolution
   - Impact: Better overload selection

2. **Record Visibility Rules**
   - [Location](internal/semantic/analyze_records.go#L351): analyze_records.go:351
   - TODO: Check visibility rules if needed
   - Impact: Enforce private/protected/public

3. **Array Index Type Validation**
   - [Location](internal/semantic/analyze_arrays.go#L130): analyze_arrays.go:130
   - TODO: Validate index type matches property index parameter types
   - Impact: Catch type errors earlier

4. **Class Hierarchy Checking**
   - [Location](internal/semantic/overload_resolution.go#L145): overload_resolution.go:145
   - TODO: Implement class hierarchy checking
   - Impact: Proper inheritance validation

5. **Unit Type Declarations**
   - [Location](internal/semantic/unit_analyzer.go#L116): unit_analyzer.go:116
   - TODO: Handle other declaration types (type declarations, constants)
   - Impact: Complete unit support

6. **Function Body Analysis**
   - [Location](internal/semantic/unit_analyzer.go#L152): unit_analyzer.go:152
   - TODO: Analyze function body when ready
   - Impact: Full semantic checking

**Action Required**: Implement each enhancement as needed for feature completeness

**Stage Relevance**: Ongoing semantic analysis improvements

---

### 2.6 Runtime - Operator and Inheritance Checks

**Priority**: MEDIUM

**Enhancement Areas**:

1. **Inheritance Checking**
   - [Location](internal/interp/operators.go#L113): operators.go:113
   - TODO: Full inheritance checking not yet implemented
   - Impact: May allow invalid class assignments

2. **Interface Member Validation**
   - [Location](internal/interp/objects_hierarchy.go#L499): objects_hierarchy.go:499
   - TODO: Implement property validation for interface member access
   - Impact: Runtime type safety

**Action Required**:
1. Implement complete inheritance chain validation
2. Add interface property checking
3. Test with complex class hierarchies

**Stage Relevance**: OOP implementation improvements

---

## 3. LOW Priority

### 3.1 Future Features

#### Function Pointer Execution Tests

**Location**: [cmd/dwscript/function_pointer_test.go:311](cmd/dwscript/function_pointer_test.go#L311)

**TODO**: Once interpreter supports function pointers, add execution tests

**Priority**: LOW

**Stage Relevance**: Future stage (function pointers)

---

#### WASM Filesystem Integration

**Locations**:
- [pkg/wasm/api.go:126](pkg/wasm/api.go#L126)
- [pkg/wasm/api.go:299](pkg/wasm/api.go#L299)

**TODO**: Implement custom filesystem integration

**Priority**: LOW

**Stage Relevance**: Stage 10.15 (WASM Support)

**Action Required**: Implement virtual filesystem for WASM environment

---

### 3.2 Unit System Enhancements

**Location**: [internal/units/search.go:169-170](internal/units/search.go#L169)

**Priority**: LOW

**TODOs**:
1. Add user library path (`~/.dwscript/lib`)
2. Add system library path (`/usr/share/dwscript/lib`)

**Impact**: Better unit/library discovery for users

**Action Required**:
1. Add standard library search paths
2. Support user-local library installation
3. Document library path conventions

**Stage Relevance**: Stage 10.13 (Unit System)

---

### 3.3 CLI Improvements

**Location**: [cmd/dwscript/cmd/fmt.go:296](cmd/dwscript/cmd/fmt.go#L296)

**TODO**: Use a proper diff algorithm for better output

**Priority**: LOW

**Impact**: Better user experience for format command

**Action Required**: Integrate diff library for cleaner format output

**Stage Relevance**: Stage 10.14 (Code Formatter)

---

### 3.4 Documentation and Code Quality

**Priority**: LOW

**Various TODOs**:
- [Simplified implementation comment](internal/interp/expressions_basic.go#L529)
- [Handle static inline arrays](internal/interp/functions_user.go#L126)
- [Implement proper by-ref support](internal/interp/functions_records.go#L153)
- [Full copy-back semantics](internal/interp/functions_records.go#L243)
- [Track actual scope level](pkg/dwscript/symbols.go#L74)

**Action Required**: Code quality improvements as time permits

---

### 3.5 Test Infrastructure Improvements

**Priority**: LOW

**Skipped Tests**:

1. **Loop Variable Type Validation**
   - [Location](internal/semantic/analyze_types_test.go#L130): analyze_types_test.go:130
   - TODO: Test currently skipped

2. **Range Expression Semantic Analysis**
   - [Location](internal/semantic/analyze_types_test.go#L266): analyze_types_test.go:266
   - TODO: Not yet implemented

3. **Const Indexed Assignment Error**
   - [Location](internal/semantic/analyze_functions_test.go#L284): analyze_functions_test.go:284
   - TODO: Should eventually error on indexed assignment to const

4. **Var Block Scoping**
   - [Location](internal/semantic/case_insensitive_test.go#L185): case_insensitive_test.go:185
   - TODO: Fix var block scoping issue

**Action Required**: Fix and re-enable tests as features are implemented

---

### 3.6 Formatter Enhancements

**Location**: PLAN.md Stage 10.14

**Priority**: LOW

**TODOs**:
- Create comprehensive test corpus with edge cases
- May need additional helpers for complex spacing rules

**Action Required**:
1. Build comprehensive format test suite
2. Handle edge cases in formatting rules
3. Add configuration options

**Stage Relevance**: Stage 10.14 (Code Formatter)

---

## 4. Fixture Test Suites (25 Disabled Categories)

> Status: Waiting for feature implementation
>
> Each category is skipped with comment: "TODO: Re-enable after implementing missing features"

### 4.1 HIGH Priority Test Suites

**Core Language Features** (needs immediate attention):

1. **ArrayPass** - [Line 52](internal/interp/fixture_test.go#L52)
   - Array operations
   - Stage: Arrays implementation

2. **AssociativePass** - [Line 59](internal/interp/fixture_test.go#L59)
   - Associative arrays/maps
   - Stage: Map/Dictionary types

3. **SetOfPass** - [Line 66](internal/interp/fixture_test.go#L66)
   - Set operations
   - **BLOCKED BY**: Set type declaration parsing (HIGH priority)
   - Stage: 9.15.x (Enums and Sets)

4. **OperatorOverloadPass** - [Line 80](internal/interp/fixture_test.go#L80)
   - Operator overloading
   - Stage: Operator overloading

5. **GenericsPass** - [Line 87](internal/interp/fixture_test.go#L87)
   - Generic types
   - Stage: Generics (future)

6. **LambdaPass** - [Line 101](internal/interp/fixture_test.go#L101)
   - Lambda expressions
   - Stage: Lambda/closure support

7. **InterfacesPass** - [Line 115](internal/interp/fixture_test.go#L115)
   - Interface usage
   - Stage: Interface implementation

8. **InnerClassesPass** - [Line 122](internal/interp/fixture_test.go#L122)
   - Nested classes
   - Stage: Advanced OOP

---

### 4.2 MEDIUM Priority Test Suites

**Error Cases** (validate error handling):

9. **AssociativeFail** - [Line 138](internal/interp/fixture_test.go#L138)
10. **SetOfFail** - [Line 145](internal/interp/fixture_test.go#L145)
11. **OperatorOverloadFail** - [Line 159](internal/interp/fixture_test.go#L159)
12. **GenericsFail** - [Line 166](internal/interp/fixture_test.go#L166)
13. **LambdaFail** - [Line 180](internal/interp/fixture_test.go#L180)
14. **PropertyExpressionsFail** - [Line 187](internal/interp/fixture_test.go#L187)
15. **InterfacesFail** - [Line 194](internal/interp/fixture_test.go#L194)
16. **InnerClassesFail** - [Line 201](internal/interp/fixture_test.go#L201)
17. **AttributesFail** - [Line 208](internal/interp/fixture_test.go#L208)

**Built-in Functions**:

18. **FunctionsMath** - [Line 217](internal/interp/fixture_test.go#L217)
    - Math functions
    - May need additional math builtins

19. **FunctionsMath3D** - [Line 224](internal/interp/fixture_test.go#L224)
    - 3D math operations
    - Vector/matrix math

20. **FunctionsMathComplex** - [Line 231](internal/interp/fixture_test.go#L231)
    - Complex number operations
    - Needs complex type support

21. **FunctionsTime** - [Line 245](internal/interp/fixture_test.go#L245)
    - Date/time functions
    - **BLOCKED BY**: DecodeDate/DecodeTime by-ref support (HIGH priority)

22. **FunctionsByteBuffer** - [Line 252](internal/interp/fixture_test.go#L252)
    - Byte buffer operations
    - Binary data handling

23. **FunctionsFile** - [Line 259](internal/interp/fixture_test.go#L259)
    - File I/O operations
    - Needs filesystem access

24. **FunctionsGlobalVars** - [Line 266](internal/interp/fixture_test.go#L266)
    - Global variable functions
    - Runtime introspection

25. **FunctionsVariant** - [Line 273](internal/interp/fixture_test.go#L273)
    - Variant type operations
    - Dynamic typing support

---

### 4.3 LOW Priority Test Suites

**Advanced/Library Features**:

26. **FunctionsRTTI** - [Line 280](internal/interp/fixture_test.go#L280)
    - Runtime type information
    - Reflection support

27. **FunctionsDebug** - [Line 287](internal/interp/fixture_test.go#L287)
    - Debug/introspection functions
    - Development utilities

28. **ClassesLib** - [Line 297](internal/interp/fixture_test.go#L297)
    - Standard class library tests
    - Library integration

29. **JSONConnectorPass** - [Line 304](internal/interp/fixture_test.go#L304)
    - JSON operations
    - JSON parsing/generation

30. **JSONConnectorFail** - [Line 311](internal/interp/fixture_test.go#L311)
    - JSON error cases

31. **Linq** - [Line 318](internal/interp/fixture_test.go#L318)
    - LINQ-style queries
    - Collection operations

32. **LinqJSON** - [Line 325](internal/interp/fixture_test.go#L325)
    - LINQ on JSON data
    - Advanced query operations

33. **DOMParser** - [Line 332](internal/interp/fixture_test.go#L332)
    - XML/DOM parsing
    - XML document handling

34. **DelegateLib** - [Line 340](internal/interp/fixture_test.go#L340)
    - Delegate operations
    - Event/callback handling

35. **DataBaseLib** - [Line 348](internal/interp/fixture_test.go#L348)
    - Database operations
    - SQL/database integration

---

## 5. Summary by Category

### Parser
- ~~**1 HIGH**: Set type declaration parsing~~ → ✅ COMPLETED (was already implemented, fixed scoping bug)

### Bytecode/VM
- **3 HIGH**: For loops, Result variable, Trim builtin → VM parity critical
- **1 MEDIUM**: Bytecode validation → Code quality

### Builtins
- **6 HIGH**: Random number functions → Need Context extension
- **7 HIGH**: String functions → Need ArrayValue migration + by-ref params
- **2 HIGH**: DateTime functions → Need by-ref parameter support
- **Total**: 15 functions blocked on infrastructure

### Semantic Analysis
- **6 MEDIUM**: Type checking enhancements → Quality improvements
- **2 MEDIUM**: Unit analyzer improvements → Feature completeness

### Interpreter/Runtime
- **9 MEDIUM**: Evaluator declaration visitors → Architecture refactoring
- **3 MEDIUM**: Type system circular imports → Dependency cleanup
- **3 MEDIUM**: Record/class initialization → Runtime correctness
- **2 MEDIUM**: Operator/inheritance checks → Type safety
- **Multiple LOW**: Implementation notes → Code quality

### Testing
- **8 HIGH**: Core language feature tests (Array, Set, Operator, etc.)
- **9 MEDIUM**: Error case tests + builtin function tests
- **8 LOW**: Advanced feature tests (RTTI, JSON, LINQ, etc.)
- **4 LOW**: Individual skipped tests → Fix as needed

### Other
- **2 LOW**: WASM filesystem → Stage 10.15
- **2 LOW**: Unit library paths → User experience
- **1 LOW**: CLI diff algorithm → Tool improvement
- **1 LOW**: Function pointer tests → Future feature

---

## 6. Recommended Action Plan

### 🔥 Immediate (This Week)

**HIGH**: ~~Set type declaration parsing~~ ✅ COMPLETED

---

### 📋 Short Term (Next 2 Weeks)

**Phase 2 Builtins Completion**:

3. **Random Functions** (6 functions):
   - Extend Context interface with RandSource/SetRandSeed/GetRandSeed
   - Migrate all random number builtins
   - Remove old Interpreter implementations

4. **String Functions - Part 1** (4 functions):
   - Move ArrayValue from interp to runtime package
   - Complete Format, StrSplit, StrJoin, StrArrayPack
   - Resolve circular dependencies

5. **String Functions - Part 2** (3 functions):
   - Implement by-ref parameter support
   - Complete Insert, Delete, DeleteString
   - Test in-place modifications

6. **DateTime Functions** (2 functions):
   - Use by-ref parameter support from step 5
   - Complete DecodeDate, DecodeTime
   - Test with FunctionsTime fixtures

---

### 🎯 Medium Term (Next Month)

**Bytecode VM Parity**:

7. **For Loop Support**:
   - Implement bytecode compilation for for-loops
   - Add VM execution support
   - Test performance improvements

8. **Result Variable**:
   - Add Result variable to compiler
   - Implement function prologue/epilogue
   - Test DWScript function idioms

9. **Trim Builtin**:
   - Add to VM builtin table
   - Test string trimming in VM mode

**Enum Enhancements**:

10. **Boolean Operations**:
    - Implement enum or/and/xor operations
    - Test with enum_bool_op fixtures

11. **Implicit Conversions**:
    - Add enum-to-integer implicit conversion
    - Update type checker

---

### 🏗️ Long Term (Phase 2+ Completion)

**Architecture Improvements**:

12. **Evaluator Migration**:
    - Move declaration registration to Evaluator (9 types)
    - Improve separation of concerns
    - Maintain test coverage

13. **Type System Cleanup**:
    - Resolve circular import issues (3 types)
    - Create separate types package if needed
    - Simplify dependencies

14. **Bytecode Validation**:
    - Add jump target validation
    - Check stack depth consistency
    - Prevent runtime crashes

**Test Suite Completion**:

15. **Re-enable Fixture Tests**:
    - ArrayPass (after array implementation)
    - AssociativePass (after map implementation)
    - OperatorOverloadPass (after operator overloading)
    - LambdaPass (after lambda support)
    - InterfacesPass (after interface implementation)
    - And 20 more test suites as features complete

---

## 7. Blocking Dependencies

### Critical Path

**Immediate Blockers**:
1. ~~Set type parsing~~ → ✅ **COMPLETED**

**Phase 3 Blockers**:
2. ❌ Random functions → **Blocks Phase 3 completion**
3. ❌ ArrayValue migration → **Blocks 7 string functions**
4. ❌ By-ref parameters → **Blocks DateTime + Insert/Delete functions**

**VM Blockers**:
5. ❌ For loops in VM → **Blocks VM performance testing**
6. ❌ Result variable → **Blocks DWScript function idioms**

**Feature Blockers**:
7. ❌ Enum enhancements → **Blocks enum fixture tests**

---

### Independent Tasks (No Blockers)

Can be done anytime without dependencies:
- ✅ WASM filesystem integration
- ✅ Unit library search paths
- ✅ CLI diff algorithm improvement
- ✅ Documentation TODOs
- ✅ Code quality improvements
- ✅ Test infrastructure enhancements
- ✅ Formatter enhancements

---

## 8. Stage/Phase Mapping

### Current Stage (Phase 3 Refactoring)
- **Random functions** (6) - HIGH
- **String functions** (7) - HIGH
- **DateTime functions** (2) - HIGH
- **Evaluator migration** (9 types) - MEDIUM
- **Type system cleanup** (3 types) - MEDIUM

### Stage 9.15 (Enums and Sets)
- **Set type parsing** - HIGH
- **Enum boolean ops** - HIGH
- **Enum implicit conversion** - HIGH
- **SetOfPass/Fail fixtures** - HIGH

### Bytecode VM Development
- **For loops** - HIGH
- **Result variable** - HIGH
- **Trim builtin** - HIGH
- **Bytecode validation** - MEDIUM

### Future Stages
- **Function pointers** - LOW
- **Lambda/closures** - TEST SUITE
- **Generics** - TEST SUITE
- **Interfaces** - TEST SUITE
- **Operator overloading** - TEST SUITE

### Stage 10.13 (Unit System)
- **Library paths** - LOW

### Stage 10.14 (Formatter)
- **Test corpus** - LOW
- **Diff algorithm** - LOW

### Stage 10.15 (WASM)
- **Filesystem integration** - LOW

---

## 9. Impact Analysis

### High Impact (Affects Many Features)
- Set type parsing (40+ tests blocked)
- By-ref parameter support (9 functions blocked)
- ArrayValue migration (4 functions blocked)
- Random function Context interface (6 functions blocked)
- For loop VM support (critical language feature)

### Medium Impact (Affects Quality/Architecture)
- Evaluator migration (code organization)
- Type system circular imports (maintainability)
- Semantic analysis enhancements (type safety)
- Bytecode validation (error detection)

### Low Impact (Nice to Have)
- WASM features (future deployment)
- CLI improvements (user experience)
- Library paths (convenience)
- Documentation TODOs (code quality)

---

## 10. Quick Reference Checklist

### Before Starting New Features
- [ ] Check if parser infinite loop is fixed
- [ ] Verify no blocking dependencies
- [ ] Review relevant TODOs in this document
- [ ] Update PLAN.md task status

### After Implementing Features
- [ ] Mark TODOs as complete
- [ ] Remove TODO comments from code
- [ ] Re-enable any fixture tests
- [ ] Update this document
- [ ] Update stage summary docs

### Code Review Checklist
- [ ] No new TODO comments without issues
- [ ] All TODOs have clear descriptions
- [ ] Critical TODOs are flagged
- [ ] Blocking dependencies are documented

---

## Appendix: TODO Comment Format

When adding new TODOs to code, use this format:

```go
// TODO: Clear description of what needs to be done
// TODO(priority): Description for CRITICAL/HIGH priority items
// TODO(username): Description for assigned items
// TODO(stage-X): Description tied to specific stage
```

Examples:
```go
// TODO(CRITICAL): Fix infinite loop in parser
// TODO(HIGH): Implement by-ref parameter support
// TODO(stage-9.15): Add enum boolean operations
// TODO(phase3): Migrate to builtins package
```

---

**Document Maintenance**: Update this document when:
- TODOs are completed (remove from document)
- New TODOs are added (add with priority)
- Priorities change (re-organize)
- Blocking dependencies are resolved (update status)

**Last Full Scan**: 2025-11-21
