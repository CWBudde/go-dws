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

**Estimate**: 7-13 hours (actual: ~9 hours)

**Status**: IN PROGRESS (81% complete)

**Current Status**: 26 passing, 6 failing (out of 32 total) - 81% pass rate

**Target**: 32 passing, 1 deferred - 97% pass rate

**Progress**:
- Task 9.1.1: ✅ DONE (Interface Member Access)
- Task 9.1.2: ✅ DONE ('implements' Operator with Class Types)
- Task 9.1.3: ✅ DONE (Interface-to-Object/Interface Casting)
- Task 9.1.4: ✅ DONE (Interface Fields in Records)
- Task 9.1.5: ✅ DONE (Interface Lifetime Management)
- Task 9.1.6: ⏸️ DEFERRED (Method Delegate Assignment)

**Test File**: `internal/interp/interface_reference_test.go`

**Description**: The interface reference tests reveal 6 categories of missing or broken functionality in the interpreter's interface implementation. These failures prevent proper interface member access, lifetime management, type checking, and casting operations.

**Subtasks** (ordered by impact and difficulty):

### 9.1.1 Interface Member Access [CRITICAL]

**Goal**: Fix interface member/method access to allow calling methods on interface instances.

**Status**: DONE

**Impact**: Fixes 3 tests (call_interface_method, intf_spread, intf_spread_virtual)

**Estimate**: 1-2 hours

**Tests Fixed**:

- `call_interface_method` ✓
- `interface_multiple_cast` (requires interface-to-interface casting - see 9.1.x)
- `interface_properties` (requires indexed property support)
- `intf_casts` (requires IInterface built-in type)
- `intf_spread` ✓
- `intf_spread_virtual` ✓

**Error**: `ERROR: cannot access member 'X' of type 'INTERFACE' (no helper found)`

**Root Cause**:

- `evalMemberAccess` in `internal/interp/objects_hierarchy.go:282-320` doesn't handle InterfaceInstance
- `AsObject(objVal)` returns false for InterfaceInstance (only recognizes ObjectInstance)
- Fallback path checks for helpers, but interfaces don't have helpers

**Implementation**:

- [x] Add InterfaceInstance detection in `evalMemberAccess` before `AsObject` check
- [x] Extract underlying ObjectInstance via `ii.Object`
- [x] Verify method exists in interface definition
- [x] Delegate method calls to underlying object
- [x] Update `evalMethodCall` in `objects_methods.go` to handle InterfaceInstance

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

**Status**: DONE

**Impact**: Fixes 1 test (interface_implements_intf)

**Estimate**: 30 minutes

**Tests Fixed**:

- `interface_implements_intf` ✓
- `interface_inheritance_instance` (still failing - requires interface inheritance checking fix)

**Error**: `ERROR: 'implements' operator requires object instance, got CLASS`

**Root Cause**:

- `evalImplementsExpression` in `internal/interp/expressions_complex.go:227-230` only accepts ObjectInstance
- Test uses class identifiers: `if TMyImplementation implements IMyInterface then`

**Implementation**:

- [x] Update `evalImplementsExpression` to handle ClassInfoValue and ClassValue
- [x] Use existing `classImplementsInterface` helper for class type checks
- [x] Keep existing ObjectInstance logic (extract class from instance)

**Example Test**:

```pascal
if TMyImplementation implements IMyInterface then  // TMyImplementation is a CLASS
   PrintLn('Ok');
```

**Files to Modify**:

- `internal/interp/expressions_complex.go`

### 9.1.3 Interface-to-Object/Interface Casting [MEDIUM]

**Goal**: Validate interface-to-object and interface-to-interface casts.

**Status**: DONE

**Impact**: Fixes 2 tests

**Estimate**: 1 hour (actual: 2 hours due to debugging interface wrapping)

**Tests Fixed**:
- `interface_multiple_cast` ✓
- `interface_cast_to_obj` ✓

**Error**: Missing exception when invalid cast attempted, interface variables not properly initialized

**Expected Behavior**:
```pascal
var d := IntfRef as TDummyClass;  // Should throw exception if incompatible
```

**Expected Output**: `Cannot cast interface of "TMyImplementation" to class "TDummyClass" [line: 23, column: 21]`

**Root Causes**:
1. `evalAsExpression` only handled object-to-interface casting, not InterfaceInstance as left operand
2. `createZeroValue` didn't handle interface types, so interface variables initialized as NilValue instead of InterfaceInstance
3. Exception check missing in `evalVarDeclStatement` after initializer evaluation

**Implementation**:
- [x] Detect when left side is InterfaceInstance in `evalAsExpression`
- [x] Handle interface-to-class casting (extract underlying object, validate compatibility)
- [x] Handle interface-to-interface casting (validate implementation, create new interface instance)
- [x] Raise exception for invalid interface-to-class casts with proper message format
- [x] Add InterfaceInstance support to `createZeroValue` for proper interface variable initialization
- [x] Add exception check in `evalVarDeclStatement` after evaluating initializers
- [x] Add InterfaceInstance case to `getTypeIDAndName` for TypeOf support

**Files Modified**:
- `internal/interp/expressions_complex.go` - Interface casting logic
- `internal/interp/statements_declarations.go` - Interface initialization, exception checks
- `internal/interp/builtins_type.go` - TypeOf support for interfaces

### 9.1.4 Interface Fields in Records [LOW]

**Goal**: Support interface-typed fields in record types.

**Status**: DONE

**Impact**: Fixes 1 test

**Estimate**: 1-2 hours (actual: 1.5 hours)

**Tests Fixed**:
- `intf_in_record` (partial - core functionality working, minor output format differences remain)

**Error**: `ERROR: unknown or invalid type for field 'FIntf' in record 'TRec'`

**Root Cause**:
- `resolveTypeFromAnnotation` in `internal/interp/helpers_conversion.go` didn't recognize interface types
- Record field initialization didn't create InterfaceInstance for interface-typed fields

**Implementation**:
- [x] Update type resolution to recognize interface type annotations
- [x] Add interface type to supported field types in records
- [x] Initialize interface fields as InterfaceInstance with nil object
- [x] Handle interface assignment in record field access
- [x] Handle interface method calls through record fields
- [x] Update error messages for nil interface access

**Example Test**:
```pascal
type TRec = record
  FIntf: IMyInterface;  // Interface-typed field
end;
```

**Files Modified**:
- `internal/interp/helpers_conversion.go` - Added interface type recognition
- `internal/interp/record.go` - Initialize interface fields as InterfaceInstance (2 locations)
- `internal/interp/objects_hierarchy.go` - Updated nil interface error message
- `internal/interp/objects_methods.go` - Updated nil interface error for method calls
- `internal/interp/errors.go` - Added location tracking for member access/method calls

### 9.1.5 Interface Lifetime Management [HIGH PRIORITY - COMPLEX]

**Goal**: Implement reference counting and automatic destructor calls for interface-held objects.

**Status**: DONE

**Impact**: Fixes 4 out of 5 tests (80% success rate for lifetime management)

**Estimate**: 4-8 hours (actual: 6 hours including debugging)

**Tests Fixed**:
- `interface_lifetime` ✓
- `interface_lifetime_scope` ✓
- `interface_lifetime_scope_ex1` ✓
- `interface_lifetime_scope_ex2` (partial - edge case with function return assignment)
- `interface_lifetime_simple` ✓

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
- [x] Add reference count field to ObjectInstance in `class.go`
- [x] Increment ref count when interface wraps object (NewInterfaceInstance)
- [x] Decrement ref count when interface assigned nil or goes out of scope
- [x] Call destructor when ref count reaches zero (ReleaseInterfaceReference)
- [x] Update InterfaceInstance assignment in `statements_assignments.go`
- [x] Add scope-based cleanup in function return (cleanupInterfaceReferences)
- [x] Handle interface-to-interface assignment (copy semantics with RefCount++)
- [x] Handle var parameter interface assignments with reference counting
- [x] Initialize Result for interface return types as InterfaceInstance

**Complexity Notes**:
- Must avoid memory leaks (objects not freed) ✓
- Must avoid premature cleanup (object freed while still referenced) ✓
- Must handle circular references gracefully (not fully tested)
- Need to track all interface assignments and scope exits ✓

**Overall Test Results**:
- Before: 22/33 tests passing (67%)
- After: 26/32 tests passing (81%)

**Files Modified**:
- `internal/interp/class.go` - Added RefCount field to ObjectInstance
- `internal/interp/interface.go` - Reference counting and cleanup implementation
- `internal/interp/functions_user.go` - Result initialization for interface returns
- `internal/interp/statements_assignments.go` - Interface assignment reference management

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

- [x] **Task 9.2: Fix Class Constant Resolution in Semantic Analyzer** - Enable class constants to be accessed from both instance and class methods, with proper visibility checking and inheritance support. Fixed `analyzeIdentifier` to check class constants, updated `analyzeMemberAccess` for class constant access via class name/instance, added interpreter support. All subtasks complete (9.2.1-9.2.3). (Estimate: 2-3 hours) ✅

---

- [x] **Task 9.3: Support Default Constructors (Constructor `default` Keyword)** - Implement support for the `default` keyword on constructors to enable using `new ClassName(args)` syntax with non-Create constructors. Added `DefaultConstructor` field to `ClassType`, updated `analyzeNewExpression` to use default constructor, added validation for single default per class, updated interpreter. All subtasks complete (9.3.1-9.3.4). (Estimate: 3-4 hours) ✅

---

- [x] **Task 9.4: Implement Variant Type Support** - Implement full Variant type support including Null/Unassigned values and variant operations. Added Null and Unassigned values, implemented variant arithmetic/logical/comparison operators with type coercion, implemented coalesce operator (??), added variant type conversions and casts. All subtasks complete (9.4.1-9.4.5). Unlocks 17 failing tests in SimpleScripts. (Estimate: 20-24 hours, 2.5-3 days) ✅

---

## Task 9.5: Implement Class Variables (class var)

**Goal**: Add support for class variables (`class var`) in class declarations.

**Estimate**: 12-15 hours (1.5-2 days)

**Status**: NOT STARTED

**Impact**: Unlocks 11 failing tests in SimpleScripts

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

**Status**: PARTIALLY COMPLETE (blocked on Task 9.18)

**Implementation**:
1. Create global storage for class variables (separate from instances) ✓
2. Implement class variable initialization ✓
3. Handle class variable access via class name ✓
4. Handle class variable access via instance (lookup in class, not instance) ⚠️ PARTIAL

**Files Modified**:
- `internal/interp/declarations.go` (lines 309-352: class var initialization already complete)
- `internal/interp/objects_hierarchy.go` (lines 51-54, 227-230, 389-391: class var access)
- `internal/interp/objects_hierarchy.go` (lines 322-348: attempted nil instance access - blocked)
- `internal/interp/statements_assignments.go` (lines 488-494: class var assignment via class name)
- `internal/interp/fixture_test.go` (added SetSemanticInfo calls for tests)
- `internal/interp/class_var_test.go` (new file: runtime tests for class variables)

**What Works**:
- ✓ Class variables stored in `ClassInfo.ClassVars map[string]Value`
- ✓ Initialization with default values or explicit init expressions
- ✓ Access via class name: `TBase.Test` (read and write)
- ✓ Inheritance: `lookupClassVar` recursively checks parent classes
- ✓ Access via object instances: `obj.Test` where obj is non-nil
- ✓ Assignment via class name: `TBase.Test := 123`
- ✓ Tests: `TestClassVariableAccessViaClassNameRuntime` passes

**What's Blocked**:
- ⚠️ Access via NIL instances: `var b : TBase; PrintLn(b.Test)`
- **Blocker**: Task 9.18 SemanticInfo migration incomplete
  - SemanticInfo exists but not populated with type information
  - Cannot determine class type of nil variables at runtime
  - Code in place (lines 326-348) but SemanticInfo.GetType() returns nil

**Potential Solutions**:
1. Complete Task 9.18: Populate SemanticInfo.SetType() in semantic analyzer
2. Use typed nil values that carry class metadata
3. Store variable type information in environment alongside values

**Testing**:
```bash
# Works:
go test -v ./internal/interp -run TestClassVariableAccessViaClassNameRuntime

# Fails (nil instance access):
go test -v ./internal/interp -run TestClassVariableAccessViaInstanceRuntime
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/class_var
```

**Note**: Significant progress made. Core infrastructure complete. Final piece requires either Task 9.18 completion or alternative type tracking mechanism.

---

## Task 9.6: Enhance Class Constants with Field Initialization

**Goal**: Support field initialization from class constants in class body.

**Estimate**: 6-8 hours (0.75-1 day)

**Status**: NOT STARTED

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

## Task 9.7: Implement Variant Type Support

**Goal**: Add full Variant type support with Null/Unassigned values and type coercion.

**Estimate**: 12-16 hours (1.5-2 days)

**Status**: NOT STARTED

**Impact**: Unlocks 17 failing tests in SimpleScripts (14 direct + 3 boolean/casting related)

**Priority**: P0 - CRITICAL (Fundamental type system feature)

**Description**: Variant is a universal type in DWScript that can hold any value type (Integer, Float, String, Boolean) and special values (Null, Unassigned). It's essential for dynamic programming patterns and optional values. This requires extensive work across the type system, semantic analyzer, and both execution engines.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P0, Section 1

**Failing Tests** (17 total):
- assert_variant
- boolean_optimize (variant boolean operations)
- case_variant_condition
- coalesce_variant (`??` operator with Null)
- compare_vars
- for_to_variant
- var_eq_nil
- var_nan
- var_param_casts
- variant_compound_ops (`+=`, `-=`, etc.)
- variant_logical_ops (`and`, `or`, `xor`)
- variant_ops (arithmetic operations)
- variant_unassigned_equal
- variants_as_casts
- variants_binary_bool_int
- variants_casts
- variants_is_bool

**Example**:
```pascal
var v : Variant;
v := Null;           // Special null value
v := 42;             // Holds integer
v := 'hello';        // Holds string
v := 3.14;           // Holds float
PrintLn(v ?? 0);     // Coalesce operator
if v = Null then     // Null comparison
  PrintLn('null');
```

**Type System Design**:
- Variant can hold: Integer, Float, String, Boolean, Object references
- Special values: Null (like SQL NULL), Unassigned (uninitialized)
- Operators automatically perform type coercion
- Comparison with Null follows SQL semantics
- VarType() function returns type tag

**Complexity**: High - Affects all layers (types, semantic, runtime)

**Subtasks**:

### 9.7.1 Add Variant Type to Type System

**Goal**: Define Variant type and its representation.

**Estimate**: 2-3 hours

**Implementation**:
1. Add `VariantType` struct to `internal/types/types.go`
2. Implement `IsVariant()` predicate and type equality
3. Add `Null` and `Unassigned` as special constant values
4. Define type coercion rules from/to Variant

**Files to Modify**:
- `internal/types/types.go` (add VariantType)
- `internal/types/conversion.go` (type coercion rules)

### 9.7.2 Parse Variant Type Declarations

**Goal**: Support `var v : Variant;` syntax.

**Estimate**: 1-2 hours

**Implementation**:
1. Add VARIANT token type (may already exist)
2. Parse `Variant` as built-in type name
3. Handle Variant in variable/parameter/field declarations

**Files to Modify**:
- `internal/lexer/token_type.go` (VARIANT token if needed)
- `internal/parser/types.go` (parse Variant type)

### 9.7.3 Semantic Analysis for Variant Operations

**Goal**: Type check operations involving Variants with automatic coercion.

**Estimate**: 4-5 hours

**Implementation**:
1. Binary operators with Variant operands produce Variant results
2. Assignments to Variant accept any value type
3. Comparisons with Null/Unassigned (special semantics)
4. VarType() built-in function support
5. `??` (coalesce) operator semantic analysis

**Files to Modify**:
- `internal/semantic/analyze_expressions.go` (operator type checking)
- `internal/semantic/builtin_functions.go` (VarType function)
- `internal/semantic/analyze_operators.go` (coalesce operator)

### 9.7.4 Runtime Variant Value Representation

**Goal**: Implement Variant storage and type tagging in interpreter.

**Estimate**: 3-4 hours

**Implementation**:
1. Create `VariantValue` struct with type tag and value union
2. Special values: `NullValue`, `UnassignedValue`
3. Automatic boxing/unboxing when assigning to/from Variant
4. VarType() function implementation

**Files to Modify**:
- `internal/interp/values.go` (VariantValue struct)
- `internal/interp/builtin_functions.go` (VarType implementation)

### 9.7.5 Variant Operators in Interpreter

**Goal**: Implement arithmetic, logical, and comparison operators for Variants.

**Estimate**: 2-3 hours

**Implementation**:
1. Automatic type coercion before operations (String + Integer, etc.)
2. Null propagation (Null + 1 = Null)
3. Comparison operators with Null semantics
4. Compound assignment operators (+=, -=)
5. Coalesce operator (??) implementation

**Files to Modify**:
- `internal/interp/expressions_operators.go` (variant operations)
- `internal/interp/expressions_binary.go` (coercion logic)

### 9.7.6 Bytecode VM Variant Support

**Goal**: Add Variant opcodes and operations to bytecode VM.

**Estimate**: 3-4 hours

**Implementation**:
1. Add Variant value type to bytecode constant pool
2. Opcodes: VAR_ADD, VAR_SUB, VAR_MUL, VAR_DIV, VAR_CMP
3. Boxing/unboxing instructions
4. Null/Unassigned constant values

**Files to Modify**:
- `internal/bytecode/instruction.go` (variant opcodes)
- `internal/bytecode/vm.go` (variant operation handlers)
- `internal/bytecode/compiler.go` (emit variant operations)

**Success Criteria**:
- All 17 variant tests pass
- Variant can hold Integer, Float, String, Boolean, Null, Unassigned
- Operators perform automatic type coercion
- Null comparisons follow SQL semantics
- VarType() returns correct type tag
- Coalesce operator (??) works correctly

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/variant
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/coalesce_variant
go test -v ./internal/semantic -run TestVariantType
```

---

## Task 9.8: Implement Self Keyword in Class Methods

**Goal**: Add `Self` keyword support for referencing current instance in methods.

**Estimate**: 4-6 hours (0.5-0.75 day)

**Status**: NOT STARTED

**Impact**: Unlocks 2 failing tests in SimpleScripts + enables many OOP patterns

**Priority**: P0 - CRITICAL (Fundamental OOP feature)

**Description**: `Self` is a special keyword in DWScript (like `this` in other languages) that refers to the current object instance in methods. It's essential for disambiguating field access from parameters, accessing virtual methods, and getting runtime type information (ClassName, ClassType).

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P0, Section 4

**Failing Tests** (2 total):
- class_self
- self

**Example**:
```pascal
type TBase = class
  FValue: Integer;

  class procedure MyProc;
  begin
    if Assigned(Self) then       // Check if called on instance
      PrintLn(Self.ClassName);   // Runtime type name
  end;

  procedure SetValue(Value: Integer);
  begin
    Self.FValue := Value;        // Disambiguate from parameter
  end;
end;
```

**Complexity**: Medium - Requires semantic analysis and runtime support

**Subtasks**:

### 9.8.1 Parse Self Keyword

**Goal**: Recognize `Self` as a special identifier in method bodies.

**Estimate**: 1 hour

**Implementation**:
1. Add SELF token type (may already exist as keyword)
2. Parse Self as a special identifier expression
3. Only allow Self in method/constructor/destructor bodies

**Files to Modify**:
- `internal/lexer/token_type.go` (SELF token verification)
- `internal/parser/expressions.go` (parse Self identifier)

### 9.8.2 Semantic Analysis for Self

**Goal**: Resolve Self to current instance type in methods.

**Estimate**: 2-3 hours

**Implementation**:
1. During method analysis, add Self to method scope
2. Self type = enclosing class type
3. For class methods/class procedures, Self may be Nil
4. Validate Self only appears in method contexts
5. Support Self.ClassName, Self.ClassType pseudo-properties

**Files to Modify**:
- `internal/semantic/analyze_functions.go` (add Self to method scope)
- `internal/semantic/analyze_classes.go` (Self type resolution)
- `internal/semantic/builtin_properties.go` (ClassName, ClassType)

### 9.8.3 Runtime Support for Self

**Goal**: Bind Self to current instance during method calls.

**Estimate**: 1-2 hours

**Implementation**:
1. When calling methods, bind Self in method environment
2. For class methods, Self is Nil (or class reference)
3. Self.ClassName returns runtime type name
4. Assigned(Self) checks if instance exists

**Files to Modify**:
- `internal/interp/functions_calls.go` (bind Self parameter)
- `internal/interp/objects_methods.go` (Self in method calls)
- `internal/interp/builtin_functions.go` (Assigned function)

**Success Criteria**:
- Self keyword recognized in method bodies
- Self type resolves to enclosing class
- Self.ClassName returns correct runtime type
- Assigned(Self) works in class methods
- Both tests (class_self, self) pass

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/self
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/class_self
go test -v ./internal/semantic -run TestSelfKeyword
```

---

## Task 9.9: Implement Function and Method Pointers

**Goal**: Add support for function pointer types and method pointer types.

**Estimate**: 14-18 hours (2-2.5 days)

**Status**: NOT STARTED

**Impact**: Unlocks 19 failing tests in SimpleScripts

**Priority**: P0 - CRITICAL (Advanced but commonly used feature)

**Description**: DWScript supports function pointers (procedural types) that allow storing references to functions/procedures and calling them indirectly. This enables callbacks, higher-order functions, and event handlers. Method pointers are similar but include an object reference.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P0, Section 5

**Failing Tests** (19 total):
- func_ptr1, func_ptr3, func_ptr4, func_ptr5
- func_ptr_assigned
- func_ptr_class_meth
- func_ptr_classname
- func_ptr_constant
- func_ptr_field
- func_ptr_field_no_param
- func_ptr_param
- func_ptr_property
- func_ptr_property_alias
- func_ptr_symbol_field
- func_ptr_var
- func_ptr_var_param
- meth_ptr1
- proc_of_method
- stack_of_proc

**Example**:
```pascal
type
  TMyProc = procedure;
  TMyFunc = function(x: Integer): Integer;
  TMyMethod = procedure of object;  // Method pointer

procedure Proc1;
begin
  PrintLn('Proc1');
end;

function Double(x: Integer): Integer;
begin
  Result := x * 2;
end;

var
  p: TMyProc;
  f: TMyFunc;
begin
  p := Proc1;
  p();              // Indirect call

  f := Double;
  PrintLn(f(21));   // Prints 42

  if Assigned(p) then
    p();
end;
```

**Type System Design**:
- Function types defined by signature (params + return type)
- Method pointers include object reference
- Nil assignment for uninitialized pointers
- Assigned() checks if pointer is non-nil
- Assignment compatibility based on signature matching

**Complexity**: High - New type category with complex semantics

**Subtasks**:

### 9.9.1 Parse Function Pointer Type Declarations

**Goal**: Support `TMyProc = procedure` and `TMyFunc = function(...): Type` syntax.

**Estimate**: 3-4 hours

**Implementation**:
1. Parse function pointer type declarations in type section
2. Handle procedure/function signatures without bodies
3. Parse method pointer syntax: `procedure of object`
4. Store signature in AST

**Files to Modify**:
- `internal/parser/types.go` (parse function pointer types)
- `pkg/ast/declarations.go` (FunctionPointerType node)

### 9.9.2 Add Function Pointer Types to Type System

**Goal**: Create FunctionPointerType and MethodPointerType.

**Estimate**: 3-4 hours

**Implementation**:
1. Add `FunctionPointerType` struct with signature
2. Add `MethodPointerType` struct (function pointer + object)
3. Implement signature matching for assignment compatibility
4. Handle Nil assignment to function pointers

**Files to Modify**:
- `internal/types/types.go` (FunctionPointerType, MethodPointerType)
- `internal/types/compatibility.go` (signature matching)

### 9.9.3 Semantic Analysis for Function Pointers

**Goal**: Type check function pointer assignments and calls.

**Estimate**: 4-5 hours

**Implementation**:
1. Assignment: check signature compatibility
2. Calling function pointers: resolve signature, type check arguments
3. Passing function names as values (function-to-pointer conversion)
4. Assigned() built-in for function pointers
5. Function pointers as parameters, fields, properties

**Files to Modify**:
- `internal/semantic/analyze_types.go` (function pointer type resolution)
- `internal/semantic/analyze_expressions.go` (pointer assignments, calls)
- `internal/semantic/builtin_functions.go` (Assigned for pointers)

### 9.9.4 Runtime Function Pointer Values

**Goal**: Implement function pointer storage and invocation in interpreter.

**Estimate**: 3-4 hours

**Implementation**:
1. Create `FunctionPointerValue` with reference to FunctionDecl
2. Create `MethodPointerValue` with object + method reference
3. Calling function pointers: look up function, execute
4. Assigned() checks if pointer is non-nil

**Files to Modify**:
- `internal/interp/values.go` (FunctionPointerValue, MethodPointerValue)
- `internal/interp/functions_calls.go` (indirect function calls)
- `internal/interp/builtin_functions.go` (Assigned implementation)

### 9.9.5 Bytecode VM Function Pointer Support

**Goal**: Add function pointer opcodes to bytecode VM.

**Estimate**: 2-3 hours

**Implementation**:
1. LOAD_FUNC_PTR opcode (load function reference)
2. CALL_FUNC_PTR opcode (indirect call)
3. Store function pointers in constant pool
4. Method pointer support

**Files to Modify**:
- `internal/bytecode/instruction.go` (function pointer opcodes)
- `internal/bytecode/compiler.go` (emit function pointer operations)
- `internal/bytecode/vm.go` (execute function pointer calls)

**Success Criteria**:
- Function pointer types parse correctly
- Assignment compatibility checking works
- Can assign functions to pointer variables
- Can call function pointers with correct arguments
- Assigned() works for function pointers
- Method pointers work with object references
- All 19 function pointer tests pass

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/func_ptr
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/meth_ptr
go test -v ./internal/semantic -run TestFunctionPointer
```

---

## Task 9.10: Implement For-In Loops

**Goal**: Add for-in loop support for iterating over collections and enumerations.

**Estimate**: 10-14 hours (1.5-2 days)

**Status**: NOT STARTED

**Impact**: Unlocks 14 failing tests in SimpleScripts

**Priority**: P0 - CRITICAL (Modern control flow feature)

**Description**: For-in loops provide a clean syntax for iterating over collections (arrays, strings, sets, enumerations) without manual index management. This is a key modern language feature that simplifies code.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P0, Section 6

**Failing Tests** (14 total):
- enumerations2 (enum iteration)
- for_in_array
- for_in_enum
- for_in_func_array
- for_in_record_array
- for_in_set
- for_in_str, for_in_str2, for_in_str4
- for_in_subclass
- for_var_in_array
- for_var_in_enumeration
- for_var_in_field_array
- for_var_in_string

**Example**:
```pascal
type TMyEnum = (meA, meB, meC);

var
  i: TMyEnum;
  s: String;
  arr: array of Integer;
begin
  // Iterate over enumeration
  for i in TMyEnum do
    PrintLn(i);

  // Iterate over array
  arr := [1, 2, 3];
  for var element in arr do
    PrintLn(element);

  // Iterate over string (characters)
  for var ch in 'hello' do
    PrintLn(ch);

  // Iterate over set
  for var item in [1, 2, 3] do
    PrintLn(item);
end;
```

**Syntax Variants**:
- `for variable in collection do` - Use existing variable
- `for var variable in collection do` - Declare variable inline
- Collections: arrays, strings, sets, enumerations, ranges

**Complexity**: Medium-High - Multiple collection types with different iteration semantics

**Subtasks**:

### 9.10.1 Parse For-In Loop Syntax

**Goal**: Extend parser to recognize `for ... in ... do` syntax.

**Estimate**: 3-4 hours

**Implementation**:
1. Add IN token (may already exist)
2. Parse `for <identifier> in <expression> do <statement>`
3. Parse `for var <identifier> in <expression> do <statement>`
4. Create ForInStatement AST node
5. Handle type inference for inline variable declarations

**Files to Modify**:
- `internal/parser/control_flow.go` (parse for-in loops)
- `pkg/ast/control_flow.go` (ForInStatement node)

### 9.10.2 Semantic Analysis for For-In Loops

**Goal**: Type check for-in loops with different collection types.

**Estimate**: 4-5 hours

**Implementation**:
1. Validate collection expression is iterable (array, string, set, enum, range)
2. Determine element type based on collection type:
   - Array: element type
   - String: Char
   - Set: set element type
   - Enumeration: enum element type
3. Type check loop variable against element type
4. For inline `var` declarations, infer type from collection
5. Validate loop body with loop variable in scope

**Files to Modify**:
- `internal/semantic/analyze_control_flow.go` (for-in type checking)
- `internal/semantic/type_inference.go` (element type deduction)

### 9.10.3 Interpreter For-In Loop Execution

**Goal**: Execute for-in loops over different collection types.

**Estimate**: 3-4 hours

**Implementation**:
1. Evaluate collection expression once
2. For arrays: iterate over elements
3. For strings: iterate over characters
4. For sets: iterate over members
5. For enumerations: iterate from Low to High
6. Bind loop variable to each element
7. Execute loop body for each element

**Files to Modify**:
- `internal/interp/control_flow.go` (for-in loop execution)
- `internal/interp/statements_loops.go` (iteration logic)

### 9.10.4 Bytecode VM For-In Loop Support

**Goal**: Compile for-in loops to bytecode instructions.

**Estimate**: 2-3 hours

**Implementation**:
1. ITER_START opcode (initialize iterator)
2. ITER_NEXT opcode (advance to next element)
3. ITER_END opcode (check if iteration complete)
4. Compile loop body with iterator variable
5. Handle different collection types

**Files to Modify**:
- `internal/bytecode/instruction.go` (iterator opcodes)
- `internal/bytecode/compiler.go` (compile for-in loops)
- `internal/bytecode/vm.go` (execute iterator operations)

**Success Criteria**:
- For-in syntax parses correctly
- Type checking works for all collection types
- Arrays, strings, sets, enums are iterable
- Inline `var` declarations infer correct type
- Loop variable has correct scope
- All 14 for-in tests pass

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/for_in
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/for_var_in
go test -v ./internal/semantic -run TestForInLoop
```

---

## Task 9.11: Implement Lazy Parameters

**Goal**: Add support for lazy parameter evaluation.

**Estimate**: 6-8 hours (0.75-1 day)

**Status**: NOT STARTED

**Impact**: Unlocks 3 failing tests in SimpleScripts

**Priority**: P1 - IMPORTANT (Optimization feature)

**Description**: Lazy parameters delay evaluation of argument expressions until the parameter is actually used in the function body. This is useful for conditional evaluation (like short-circuit operators) and performance optimization.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P1, Section 7

**Failing Tests** (3 total):
- lazy
- lazy_recursive
- lazy_sqr

**Example**:
```pascal
function CondEval(eval: Boolean; lazy a: Integer): Integer;
begin
  if eval then
    Result := a      // Only evaluated if eval = True
  else
    Result := 0;     // a is never evaluated
end;

function Expensive: Integer;
begin
  PrintLn('Computing...');  // This might not print
  Result := 42;
end;

begin
  PrintLn(CondEval(False, Expensive()));  // Doesn't print 'Computing...'
  PrintLn(CondEval(True, Expensive()));   // Prints 'Computing...'
end;
```

**Complexity**: Medium - Requires deferred evaluation mechanism

**Subtasks**:

### 9.11.1 Parse Lazy Parameter Modifier

**Goal**: Recognize `lazy` keyword in parameter declarations.

**Estimate**: 1-2 hours

**Implementation**:
1. Add LAZY token (may already exist)
2. Parse `lazy` modifier before parameter name or type
3. Mark parameter as lazy in AST (IsLazy flag on Parameter)
4. Validate lazy only used with value parameters (not var/out/const)

**Files to Modify**:
- `internal/parser/functions.go` (parse lazy modifier)
- `pkg/ast/functions.go` (IsLazy field on Parameter)

### 9.11.2 Semantic Analysis for Lazy Parameters

**Goal**: Type check lazy parameters correctly.

**Estimate**: 1-2 hours

**Implementation**:
1. Lazy parameters still have declared type
2. Type check argument expression (but don't evaluate)
3. Validate lazy not combined with var/out/const
4. Track which parameters are lazy for codegen

**Files to Modify**:
- `internal/semantic/analyze_functions.go` (lazy parameter validation)

### 9.11.3 Interpreter Lazy Parameter Evaluation

**Goal**: Defer evaluation of lazy arguments until first use.

**Estimate**: 3-4 hours

**Implementation**:
1. For lazy parameters, don't evaluate argument expression at call site
2. Store argument expression as a thunk (closure)
3. On first access to parameter, evaluate thunk
4. Cache result for subsequent accesses (optional optimization)
5. Handle lazy parameters in recursive calls

**Files to Modify**:
- `internal/interp/functions_calls.go` (lazy argument handling)
- `internal/interp/values.go` (ThunkValue for lazy expressions)

### 9.11.4 Bytecode VM Lazy Parameter Support

**Goal**: Compile lazy parameters to deferred evaluation bytecode.

**Estimate**: 2-3 hours

**Implementation**:
1. PUSH_THUNK opcode (push closure for lazy arg)
2. EVAL_THUNK opcode (evaluate thunk on first access)
3. Thunk representation in VM stack
4. Compile lazy argument as closure

**Files to Modify**:
- `internal/bytecode/instruction.go` (thunk opcodes)
- `internal/bytecode/compiler.go` (compile lazy parameters)
- `internal/bytecode/vm.go` (execute thunk operations)

**Success Criteria**:
- `lazy` modifier parses in parameter lists
- Lazy arguments are not evaluated at call site
- Arguments evaluated on first use in function body
- Recursive functions with lazy parameters work
- All 3 lazy parameter tests pass

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/lazy
go test -v ./internal/semantic -run TestLazyParameter
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

### 9.12.1 Parse Record Field Initialization

**Goal**: Support `FieldName: Type := Value;` syntax in records.

**Estimate**: 2-3 hours

**Implementation**:
1. Extend record field parsing to handle `:= <expression>`
2. Store initialization expression in FieldDeclaration
3. Parse nested record field access in initializers
4. Handle type inference from initializers

**Files to Modify**:
- `internal/parser/records.go` (field initialization parsing)
- `pkg/ast/records.go` (Initializer field on FieldDeclaration)

### 9.12.2 Parse Record Constants and Class Variables

**Goal**: Support `const` and `class var` in record bodies.

**Estimate**: 2 hours

**Implementation**:
1. Parse `const <name> = <value>;` in record body
2. Parse `class var <name>: <type>;` in record body
3. Store in RecordDecl AST node
4. Reuse class const/class var parsing logic

**Files to Modify**:
- `internal/parser/records.go` (record const/class var parsing)
- `pkg/ast/records.go` (add Constants, ClassVars fields)

### 9.12.3 Semantic Analysis for Record Features

**Goal**: Type check record field initializers, constants, and class variables.

**Estimate**: 4-5 hours

**Implementation**:
1. Analyze field initializers, check type compatibility
2. Validate record constants (compile-time constant values)
3. Add record class variables to type system
4. Handle nested record field access
5. Type check record literal expressions with nested records

**Files to Modify**:
- `internal/semantic/analyze_records.go` (record feature analysis)
- `internal/types/types.go` (RecordType with Constants, ClassVars)

### 9.12.4 Runtime Record Field Initialization

**Goal**: Execute field initializers when creating record values.

**Estimate**: 3-4 hours

**Implementation**:
1. On record variable declaration, initialize fields with default values
2. Evaluate field initializer expressions
3. Handle nested record initialization
4. Record constants stored as global values
5. Record class variables in global storage

**Files to Modify**:
- `internal/interp/records.go` (record initialization)
- `internal/interp/values.go` (record value creation)

### 9.12.5 Enhanced Record Methods

**Goal**: Fix record method semantics (self reference, result handling).

**Estimate**: 2-3 hours

**Implementation**:
1. Record methods receive copy of record as Self
2. Modifications to Self don't affect original (value semantics)
3. Result variable in record functions
4. Nested record field access in methods

**Files to Modify**:
- `internal/interp/records.go` (record method calls)
- `internal/semantic/analyze_records.go` (record method analysis)

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

**Goal**: Correct override validation, inherited keyword, reintroduce, and virtual constructors.

**Estimate**: 10-14 hours (1.5-2 days)

**Status**: NOT STARTED

**Impact**: Unlocks 14 failing tests in SimpleScripts

**Priority**: P1 - IMPORTANT (OOP polymorphism)

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

**Success Criteria**:
- Override validation checks parent method is virtual
- `inherited` works in methods, constructors, destructors
- `inherited MethodName` syntax works
- `reintroduce` allows shadowing without override
- Virtual constructors dispatch correctly
- Constructor/destructor chaining works
- All 14 inheritance/virtual method tests pass

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

**Status**: NOT STARTED

**Impact**: Unlocks 12 failing tests in SimpleScripts

**Priority**: P1 - IMPORTANT (Type system completeness)

**Description**: Enumerations currently have basic support, but DWScript includes advanced features like boolean operations on enums, bounds checking (Low/High), EnumByName function for string-to-enum conversion, enum flags (sets), scoped enums, and enum element deprecation.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P1, Section 11

**Failing Tests** (12 total):
- aliased_enum
- enum_bool_op
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

### 9.15.1 Parse Enum Modifiers

**Goal**: Support `scoped`, `flags`, and deprecation modifiers on enums.

**Estimate**: 2 hours

**Implementation**:
1. Parse `scoped` keyword after enum declaration
2. Parse `flags` keyword for bit flag enums
3. Parse deprecation attributes on enum elements
4. Store modifiers in EnumDecl AST

**Files to Modify**:
- `internal/parser/enums.go` (enum modifier parsing)
- `pkg/ast/enums.go` (Scoped, Flags, Deprecated fields)

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

### 9.15.3 Enum Bounds (Low/High)

**Goal**: Implement Low() and High() built-in functions for enums.

**Estimate**: 1-2 hours

**Implementation**:
1. `Low(EnumType)` returns first enum element
2. `High(EnumType)` returns last enum element
3. Semantic analysis: validate enum type argument
4. Runtime: return enum min/max values

**Files to Modify**:
- `internal/semantic/builtin_functions.go` (Low/High functions)
- `internal/interp/builtin_functions.go` (Low/High implementation)

### 9.15.4 EnumByName Function

**Goal**: Implement EnumByName for string-to-enum conversion.

**Estimate**: 2 hours

**Implementation**:
1. Parse `EnumByName<TEnumType>('name')` syntax
2. Semantic analysis: validate generic type parameter and string argument
3. Runtime: look up enum element by name, return value
4. Handle qualified names for scoped enums

**Files to Modify**:
- `internal/semantic/builtin_functions.go` (EnumByName function)
- `internal/interp/builtin_functions.go` (EnumByName implementation)

### 9.15.5 Scoped Enums

**Goal**: Enforce scoped enum access (TypeName.Element).

**Estimate**: 2-3 hours

**Implementation**:
1. Scoped enums require type prefix for access
2. Parse `TScopedEnum.seX` syntax
3. Semantic analysis: validate scoped enum access
4. Unscoped enums allow direct element access

**Files to Modify**:
- `internal/semantic/analyze_enums.go` (scoped enum validation)
- `internal/parser/expressions.go` (qualified enum access)

### 9.15.6 Enum Flags and Casting

**Goal**: Support flags enums and enum-to-integer casts.

**Estimate**: 2-3 hours

**Implementation**:
1. Flags enums are bit flags (2^n values)
2. Integer(enumValue) casts enum to integer
3. EnumType(intValue) casts integer to enum
4. Semantic analysis and runtime for casts

**Files to Modify**:
- `internal/semantic/analyze_casts.go` (enum casting)
- `internal/interp/expressions_casts.go` (enum cast execution)
- `internal/types/types.go` (flags enum metadata)

**Success Criteria**:
- Enum boolean operations (in, sets) work
- Low/High functions return enum bounds
- EnumByName converts string to enum
- Scoped enums enforce qualified access
- Flags enums support bit operations
- Enum-to-integer and integer-to-enum casts work
- Enum element deprecation warnings
- All 12 enum advanced feature tests pass

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

## Task 9.19 Extract Pretty-Printing from AST Nodes

**Goal**: Remove String() implementation logic from AST nodes and create a dedicated printer package.

**Estimate**: 4-6 hours (0.5-1 day)

**Status**: ✅ COMPLETED (2025-01-15)

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

- [x] 9.19.1 Design printer package architecture ✅
  - Printer struct with configurable options
  - Support for different styles (compact, detailed, multiline)
  - Support for different output formats (DWScript syntax, JSON, tree)
  - Document printer design
  - File: `pkg/printer/doc.go` (new package)

- [x] 9.19.2 Create basic printer implementation ✅
  - Printer struct with indent level, buffer, options
  - Methods for printing each node type
  - Helper methods for common patterns (indent, newline, etc.)
  - File: `pkg/printer/printer.go` (new file ~300 lines)

- [x] 9.19.3 Simplify AST String() methods ✅ (PARTIAL - some tests still need updating)
  - Replace complex formatting with simple representation
  - Example: `func (cd *ClassDecl) String() string { return fmt.Sprintf("ClassDecl(%s)", cd.Name) }`
  - Keep String() for debugging, use printer for real output
  - Files: All `pkg/ast/*.go` files (~500 lines removed)

- [x] 9.19.4 Add printer methods for all node types ✅
  - PrintProgram(), PrintClassDecl(), PrintFunctionDecl(), etc.
  - Mirror existing String() behavior initially
  - File: `pkg/printer/dwscript.go` (~1200 lines)
  - **FIXES**: Added missing modifiers (virtual, override, abstract, overload, deprecated, calling conventions)
  - **FIXES**: Added class constants and operators printing
  - **FIXES**: Fixed ReturnStatement to preserve Token.Literal
  - **FIXES**: Fixed ExitStatement to handle optional return value

- [x] 9.19.5 Add printer options and styles ✅
  - CompactPrinter (minimal whitespace)
  - DetailedPrinter (full indentation, comments)
  - TreePrinter (AST structure visualization)
  - JSONPrinter (JSON representation)
  - File: `pkg/printer/printer.go` (Format and Style enums with presets)

- [ ] 9.19.6 Update tests to use printer (DEFERRED - backward compatibility maintained)
  - Tests that rely on String() output need updating
  - Use printer for expected output strings
  - Files: `pkg/ast/*_test.go`, parser tests
  - Note: String() methods kept for backward compatibility

- [x] 9.19.7 Update CLI to use printer ✅
  - `parse` command uses printer for output
  - Add `--format` flag (dwscript, json, tree)
  - Files: `cmd/dwscript/cmd/parse.go`

- [x] 9.19.8 Add printer tests ✅
  - Test all node types print correctly
  - Test different styles produce valid output
  - Test JSON output is valid JSON
  - File: `pkg/printer/printer_test.go` (new file ~550 lines)
  - **COMPREHENSIVE**: Tests for ReturnStatement, ExitStatement, FunctionModifiers, ClassDecl, output formats, edge cases

**Files Modified**:

- `pkg/printer/printer.go` (new file ~400 lines)
- `pkg/printer/styles.go` (new file ~100 lines)
- `pkg/printer/doc.go` (new file ~30 lines)
- `pkg/printer/printer_test.go` (new file ~200 lines)
- `pkg/ast/*.go` (simplify String() methods, ~500 lines reduced)
- `pkg/ast/*_test.go` (update tests to use printer if needed)
- `cmd/dwscript/commands.go` (use printer for parse command)

**Acceptance Criteria**:
- ✅ AST String() methods are simple (<5 lines each)
- ✅ Printer package handles all formatting logic
- ✅ Multiple output formats supported (DWScript syntax, tree, JSON)
- ✅ All printer tests pass (new tests added)
- ✅ CLI `parse` command can output different formats
- ✅ Documentation explains printer usage

**Benefits**:
- AST nodes focused on structure, not presentation
- Multiple output formats possible (JSON, tree view, etc.)
- Easier to change formatting without touching AST
- Smaller AST code (~500 lines reduced)
- Better separation of concerns

**PR #114 Review Fixes** (2025-01-15):
All critical issues identified by Codex and Copilot reviews were addressed:

1. ✅ **ReturnStatement Token Preservation**: Fixed to use `Token.Literal` instead of hardcoding "result", correctly handling `Result := value`, `FunctionName := value`, and `exit` statements
2. ✅ **ExitStatement Return Values**: Added support for optional return values (`exit` vs `exit(value)`)
3. ✅ **Function Modifiers**: Added all missing modifiers - visibility (private/protected), class methods, constructor/destructor, virtual, override, reintroduce, abstract, overload, calling conventions, deprecated
4. ✅ **Class Members**: Added printing of class constants and operators that were previously omitted
5. ✅ **Comprehensive Tests**: Created 550+ lines of table-driven tests covering all fixes and edge cases

All printer package tests pass. Ready for merge.

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
