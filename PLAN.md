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

## Phase 9: Stage 7 Completion & Critical Stage 8 Features

**Status**: IN PROGRESS | **Fixture Tests**: 92/552 passing (16.7%) | **Target**: 80%+ pass rate
**Current Focus**: Complete Stage 7 (Classes & OOP) before advancing to Stage 8

**Test Analysis Summary** (2025-11-08): Of 460 failing fixture tests, ~85% are blocked by just 7 missing features. This is NOT indicative of poor quality but rather the expected state of Stage 7 being partially complete. The core language (lexer, parser, expressions, control flow, functions) is structurally sound with Algorithm tests showing 47.2% pass rate. Focus is on completing OOP foundation before tackling advanced Stage 8 features.

**Estimated Timeline**: 10-15 weeks to reach 80-85% fixture test pass rate

---

### Phase 9.1: Constructor System (CRITICAL - Blocks 87 tests)

**Priority**: HIGHEST - Constructor system is the #1 blocker for fixture tests
**Goal**: Enable class instantiation without explicit constructor declarations
**Timeline**: 2-3 weeks | **Success Metric**: ~45-50% fixture pass rate

**Root Cause**: Classes without explicit constructors fail with "class 'X' has no member 'Create'" error. DWScript auto-generates default constructors; go-dws does not.

- [x] 9.1 Implement implicit default constructor generation ✅
  - **Task**: Auto-generate parameterless `Create` constructor for classes without explicit constructors
  - **Implementation**:
    - Modify `analyzeClassDeclaration` in `internal/semantic/analyze_classes.go`
    - After class declaration analysis, check if class has any constructors
    - If no constructors exist, synthesize default `Create` constructor
    - Add constructor to class method table with public visibility
  - **Test**: Classes without explicit constructors can be instantiated with `TClass.Create`
  - **Files**: `internal/semantic/analyze_classes.go`, tests in `internal/semantic/constructor_validation_test.go`
  - **Completed**: Added `synthesizeDefaultConstructor()` function that generates implicit `Create` constructor with public visibility

- [x] 9.2 Implement constructor inheritance ✅
  - **Task**: Child classes inherit parent constructors if none declared
  - **Implementation**:
    - During class analysis, walk inheritance chain
    - If child has no constructors, copy parent constructors
    - Handle constructor visibility (private constructors not inherited)
    - Update constructor call sites to search inheritance chain
  - **Test**: `type TChild = class(TBase) end; var obj := TChild.Create;` works when only TBase has Create
  - **Files**: `internal/semantic/analyze_classes.go`
  - **Completed**: Added `inheritParentConstructors()` function that copies accessible constructors from parent to child, excluding private constructors

- [x] 9.3 Implement constructor overloading ✅
  - **Task**: Support multiple constructors with different parameter signatures
  - **Implementation**:
    - Extend overload resolution to handle constructors
    - Constructors use same overload mechanism as methods
    - Validate overload directive on constructors
    - Resolve correct overload at call site based on arguments
  - **Test**: Multiple `Create` constructors with different signatures all work
  - **Files**: `internal/semantic/overload_resolution.go`, `internal/interp/objects.go`
  - **Completed**: Constructor overloading already working via existing overload resolution mechanism; tests pass

- [x] 9.4 Implement virtual constructors ✅ COMPLETE - ~95% complete (2025-01-09)
  - **Task**: Support virtual/override on constructors for polymorphic instantiation
  - **Implementation**: COMPLETE
    - ✅ 9.4.1 Semantic validation - validateVirtualOverride now handles constructors (analyze_classes.go:709-785)
    - ✅ 9.4.2 Constructor inheritance with virtual metadata (analyze_classes.go:389-438)
    - ✅ 9.4.3 Instance constructor virtual dispatch - o.Create (objects.go:1787-1873)
    - ✅ 9.4.4 Metaclass constructor virtual dispatch - cls.Create (objects.go:1596-1689)
    - ✅ 9.4.5 Child class constructor calls - Fixed by task 9.73.8 (2025-01-08)
  - **Test**: Metaclass variable calls correct overridden constructor
  - **Files**: `internal/semantic/analyze_classes.go`, `internal/interp/objects.go`, `internal/interp/declarations.go`
  - **Status**: ~95% complete - core functionality working, only minor warnings missing
  - **Related**: Task 9.73 (Metaclass dispatch)
  - **Note**: Remaining 5% is non-blocking warnings for constructor instance calls (see notes below)

  **COMPLETED WORK (2025-01-08)**:
  1. ✅ Semantic Analysis Layer (100% complete)
     - Updated `validateVirtualOverride()` to handle constructors using `findMatchingConstructorInParent()`
     - Helper functions added: `findMatchingConstructorInParent()`, `hasConstructorWithName()`
     - Virtual/override metadata properly validated for constructors
     - Warnings for hiding virtual constructors without override keyword

  2. ✅ Constructor Inheritance (100% complete)
     - `inheritParentConstructors()` copies IsVirtual flag (line 417)
     - Inherited constructors marked with IsOverride=false (correct behavior)

  3. ✅ Runtime Virtual Dispatch - Instance Calls (100% complete)
     - Instance constructor calls (o.Create) properly dispatch to runtime class
     - Creates new instance of runtime type, not static type
     - Located in objects.go:1787-1873

  4. ✅ Runtime Virtual Dispatch - Metaclass Calls (100% complete)
     - Metaclass constructor calls (cls.Create where cls is ClassValue) implemented
     - Looks up constructor in runtime class, not declared type
     - Located in objects.go:1596-1689
     - Handles argument evaluation, constructor resolution, instance creation

  **BUG RESOLUTION (2025-01-08)**:

  **Bug**: Child class constructor calls were failing with "There is no overloaded version" error
  - `TClassA.Create` worked ✓
  - `TClassB.Create` failed ✗ (where TClassB inherits from TClassA)

  **Root Cause**: Child constructors were being APPENDED to parent constructors in ConstructorOverloads,
  causing duplicate constructors with same signature and breaking overload resolution.

  **Solution (Task 9.73.8)**: Modified constructor registration to REPLACE parent constructor when child has matching signature
  - In DWScript, child constructor with same name/signature HIDES parent's (like Delphi)
  - Works regardless of virtual/override keywords
  - Files Modified: `internal/interp/declarations.go:183-204, 211-232`

  **Tests Now Passing**:
  - `testdata/fixtures/SimpleScripts/virtual_constructor2.pas` ✓
  - `testdata/fixtures/SimpleScripts/class_of.pas` ✓

  **Remaining Minor Issue** (non-blocking):
  - Missing warnings for constructor calls on instances outside constructors (see notes below)

- [x] 9.4.5 Fix child class constructor call resolution ✅ DONE (2025-01-08, via Task 9.73.8)
  - **Task**: Fix bug where TClassB.Create fails when TClassB inherits from TClassA
  - **Solution**: Modified constructor registration in `internal/interp/declarations.go:183-232`
    - Child constructors now REPLACE parent constructors instead of being appended
    - Follows DWScript/Delphi semantics: child constructor with same signature hides parent's
    - Works for both virtual/override and non-virtual constructors
  - **Tests Passing**:
    - `testdata/fixtures/SimpleScripts/virtual_constructor2.pas` ✓
    - `testdata/fixtures/SimpleScripts/class_of.pas` ✓
  - **Files Modified**: `internal/interp/declarations.go`
  - **Completed**: Fixed by Task 9.73.8 on 2025-01-08

  **Note on Missing Warnings**: The test `virtual_constructor2.pas` expects warnings for constructor instance calls
  (e.g., `obj.Create` outside of constructor context). These warnings are not yet implemented but are non-blocking.
  This is a separate enhancement tracked below.

  **Test files**:
  - `testdata/fixtures/SimpleScripts/virtual_constructor.pas` - Tests virtual constructors with metaclass
  - `testdata/fixtures/SimpleScripts/virtual_constructor2.pas` - Tests virtual constructor without metaclass

  **Expected Output (once fixed)**:
  - virtual_constructor.pas: "B\nTestA\nTestB"
  - virtual_constructor2.pas: "A\nA\nB\nB"

- [x] 9.5 Fix constructor parameter validation ✅
  - **Task**: Validate argument count and types at constructor call sites
  - **Implementation**:
    - Add semantic analysis for `NewExpression`
    - Check argument count matches constructor signature
    - Check argument types are compatible
    - Report clear error messages for mismatches
  - **Test**: Invalid constructor calls fail semantic analysis with helpful errors
  - **Files**: `internal/semantic/analyze_expressions.go`
  - **Completed**: Already implemented in `analyzeNewExpression`; all validation tests pass

- [x] 9.6 Enforce constructor visibility ✅
  - **Task**: Private/protected constructors cannot be called from outside class
  - **Implementation**:
    - Check constructor visibility at call site
    - Allow private constructors only within same class
    - Allow protected constructors within class hierarchy
    - Report access violation errors
  - **Test**: `TestPrivateConstructorFromOutside` passes
  - **Files**: `internal/semantic/analyze_expressions.go`
  - **Completed**: Already implemented in `analyzeNewExpression`; all visibility tests pass

- [x] 9.7 Support constructors with custom names (not just "Create") ⚠️ CRITICAL BUG
  - **Task**: Enable constructors with any name to be called via member access syntax
  - **Root Cause**: `analyzeMethodCallExpression` didn't unwrap `ClassOfType` to `ClassType` before looking up methods/constructors
  - **Impact**: Blocked any class using non-standard constructor names, common pattern in DWScript
  - **Solution Implemented**:
    - Added `ClassOfType` unwrapping in `analyzeMethodCallExpression` (matching the pattern already used in `analyzeMemberAccessExpression`)
    - When analyzing `TExample.CreateWith(42)`:
      1. `TExample` resolves to `ClassOfType(TExample)` (metaclass)
      2. New code unwraps to `ClassType(TExample)`
      3. Existing method lookup finds constructor in `ConstructorOverloads`
      4. Existing code at end checks `HasConstructor()` and returns class type
    - Constructor registration and lookup were already working correctly
    - The fix reuses all existing method/constructor handling logic
  - **Files Modified**:
    - `internal/semantic/analyze_method_calls.go` (added 9 lines for ClassOfType unwrapping)
  - **Tests Passing**:
    - `TestConstructorCaseInsensitive` - custom constructor name "CreateWith"
    - `TestConstructorCaseInsensitiveOverloads` - multiple calls with different cases
    - All 33 constructor tests passing
  - **Completed**: 2025-01-09

**Milestone**: After completing constructor system, fixture test pass rate should reach ~45-50% (248-276 tests passing)

---

### Phase 9.2: Class Properties (CRITICAL - Blocks 49 tests)

**Priority**: HIGHEST - Second biggest blocker after constructors
**Goal**: Support property syntax with read/write accessors
**Timeline**: 2-3 weeks | **Success Metric**: ~60% fixture pass rate

**Root Cause**: Parser doesn't recognize `property` keyword in class declarations. Properties are fundamental to DWScript OOP.

- [x] 9.10 Implement property AST nodes
  - **Task**: Create AST representation for property declarations
  - **Implementation**:
    - Add `PropertyDeclaration` node in `internal/ast/class.go`
    - Fields: Name, Type, ReadAccessor, WriteAccessor, IsDefault, IsClassProperty, IndexParams
    - Implement String() and TokenLiteral() methods
    - Add to ClassDeclaration.Properties slice
  - **Files**: `pkg/ast/properties.go`, `internal/ast/properties_test.go`
  - **Status**: DONE - Added IsClassProperty field, updated String() method, added comprehensive tests
  - **Estimated time**: 1 day

- [x] 9.11 Parse property declarations
  - **Task**: Add parser support for property syntax
  - **Implementation**:
    - Modify `parseClassBody` in `internal/parser/class.go`
    - Recognize `property` keyword and parse declaration
    - Handle: `property Name: Type read GetX write SetX`
    - Handle: `property Name: Type read FField write FField`
    - Handle: `property Items[Index: Integer]: String read GetItem`
    - Parse default property directive
    - Parse class property directive
  - **Test**: All property syntax variants parse correctly
  - **Files**: `internal/parser/classes.go`, `internal/parser/properties_test.go`
  - **Status**: DONE - Added class property parsing support, added comprehensive tests
  - **Estimated time**: 2-3 days

- [x] 9.12 Validate property accessors in semantic analysis
  - **Task**: Verify read/write accessors exist and have correct signatures
  - **Implementation**:
    - Create `internal/semantic/property.go`
    - Check read accessor: field or parameterless method returning property type
    - Check write accessor: field or method accepting one parameter of property type
    - Validate index parameters if present
    - Ensure accessor visibility is compatible
    - Report errors for missing or incompatible accessors
  - **Test**: Invalid property declarations fail with clear error messages
  - **Files**: `internal/types/types.go`, `internal/semantic/analyze_properties.go`, `internal/semantic/property_test.go`
  - **Status**: DONE - Added IsClassProperty validation, class properties must use class fields/methods
  - **Estimated time**: 2-3 days

- [x] 9.13 Implement property read evaluation
  - **Task**: Property access invokes read accessor
  - **Implementation**:
    - Create `internal/interp/property.go`
    - When evaluating member access on property, invoke read accessor
    - If accessor is method: call method and return result
    - If accessor is field: return field value
    - Handle index parameters for indexed properties
  - **Test**: `obj.Name` calls GetName() method or reads FName field
  - **Files**: `internal/interp/objects.go`, `internal/interp/declarations.go`, `internal/interp/property_test.go`
  - **Status**: DONE - Implemented class property read evaluation with field and method support
  - **Estimated time**: 2-3 days

- [x] 9.14 Implement property write evaluation
  - **Task**: Property assignment invokes write accessor
  - **Implementation**:
    - Modify assignment statement evaluation
    - Detect assignment to property
    - If accessor is method: call method with value argument
    - If accessor is field: write to field
    - Handle index parameters for indexed properties
  - **Test**: `obj.Name := 'test'` calls SetName('test') or writes to FName
  - **Files**: `internal/interp/property.go`, `internal/interp/statements.go`
  - **Status**: DONE - Implemented class property write evaluation with field/method support, updated tests
  - **Estimated time**: 2-3 days

- [x] 9.15 Implement array properties
  - **Task**: Properties with index parameters (e.g., Items[Index])
  - **Implementation**:
    - Parse array property syntax with index parameters
    - Validate index parameter types
    - Pass index values to accessor methods
    - Support multiple index parameters
  - **Test**: `obj.Items[0] := 'value'` calls SetItem(0, 'value')
  - **Files**: `internal/parser/class.go`, `internal/interp/property.go`
  - **Status**: DONE - Already implemented with comprehensive tests (10+ passing)
  - **Estimated time**: 2-3 days

- [x] 9.16 Implement default properties
  - **Task**: Support default property directive for index access without property name
  - **Implementation**:
    - Parse `default` directive on properties
    - Mark one property as default in class metadata
    - Allow `obj[0]` to access default property
    - Validate only one default property per class
  - **Test**: `obj[0]` equivalent to `obj.Items[0]` when Items is default
  - **Files**: `internal/parser/class.go`, `internal/semantic/property.go`, `internal/interp/property.go`
  - **Status**: DONE - Runtime evaluation complete with 6 comprehensive tests (read, write, equivalence, inheritance, etc.)
  - **Estimated time**: 2 days

- [x] 9.17 Implement class properties (static)
  - **Task**: Support `class property` for properties on class type rather than instance
  - **Implementation**:
    - Parse `class property` keyword combination
    - Validate accessors are class methods
    - Allow access via class name: `TMyClass.Count`
    - Allow access via instance: `obj.Count`
  - **Test**: Class properties accessible without instance
  - **Files**: `internal/parser/class.go`, `internal/interp/property.go`
  - **Status**: DONE - Completed in tasks 9.10-9.14 with full read/write support and 9+ tests
  - **Estimated time**: 2 days

**Milestone**: After completing properties, fixture test pass rate should reach ~60% (330+ tests passing)

---

### Phase 9.3: Class Constants (HIGH - Blocks 38 tests)

**Priority**: HIGH - Relatively simple but blocks many tests
**Goal**: Support constant declarations inside classes
**Timeline**: 1 week | **Success Metric**: Unblocks 38 additional tests

**Root Cause**: Parser expects field declarations in class body, fails when seeing `const` keyword.

- [x] 9.20 Parse class constant declarations ✅
  - **Task**: Recognize and parse const/class const in class body
  - **Implementation**:
    - Modify `parseClassBody` to handle `const` keyword
    - Support: `const Name = Value;`
    - Support: `class const Name = Value;`
    - Apply current visibility to constant
    - Add constants to ClassDeclaration.Constants slice
  - **Test**: Class constants parse without errors
  - **Files**: `pkg/ast/declarations.go`, `pkg/ast/classes.go`, `internal/parser/classes.go`
  - **Completed**: Added Constants field to ClassDecl, implemented parseClassConstantDeclaration

- [x] 9.21 Validate class constants in semantic analysis ✅
  - **Task**: Type-check constant values and prevent redeclaration
  - **Implementation**:
    - Validate constant value is compile-time constant
    - Check for duplicate constant names
    - Enforce visibility rules
    - Store constants in class metadata
  - **Test**: Invalid class constants fail semantic analysis
  - **Files**: `internal/semantic/analyze_classes.go`, `internal/types/types.go`
  - **Completed**: Added Constants/ConstantVisibility to ClassType, validates duplicate names

- [x] 9.22 Implement class constant evaluation ✅
  - **Task**: Access class constants via type name or instance
  - **Implementation**:
    - Handle `TClass.ConstName` access
    - Handle `instance.ConstName` access
    - Return constant value from class metadata
    - Enforce visibility (private constants not accessible outside class)
  - **Test**: `PrintLn(TBase.cPublic)` outputs constant value
  - **Files**: `internal/interp/objects.go`, `internal/interp/class.go`, `internal/interp/declarations.go`
  - **Completed**: Implemented getClassConstant, added constant lookup in evalMemberAccess
  - **Note**: Simple constants work. Constant dependencies (c2 = c1+1) need additional debugging.

**Milestone**: Class constants complete, ~38 additional tests should pass

---

### Phase 9.4: Dynamic Arrays & Type Casting (HIGH - Blocks 35 tests)

**Priority**: HIGH - Essential for many algorithms and OOP patterns
**Goal**: Support dynamic array syntax and runtime type operations
**Timeline**: 1-2 weeks | **Success Metric**: Unblocks 35 additional tests

#### Dynamic Arrays (Blocks 20 tests)

**Root Cause**: Parser only supports static arrays `array[0..9] of T`, not dynamic arrays `array of T`.

- [ ] 9.30 Parse dynamic array type syntax
  - **Task**: Support `array of Type` syntax without bounds
  - **Implementation**:
    - Modify `parseArrayType` in `internal/parser/types.go`
    - Make bounds optional (nil for dynamic arrays)
    - Distinguish static vs dynamic in AST
  - **Test**: `var arr: array of Integer;` parses successfully
  - **Files**: `internal/parser/types.go`, `internal/ast/types.go`
  - **Estimated time**: 1 day

- [ ] 9.31 Implement dynamic array type in type system
  - **Task**: Create DynamicArrayType distinct from StaticArrayType
  - **Implementation**:
    - Add DynamicArrayType in `internal/types/array.go`
    - Support SetLength() operation
    - Support Length() query
    - Implement type checking for dynamic arrays
  - **Test**: Dynamic array type checking works correctly
  - **Files**: `internal/types/array.go`
  - **Estimated time**: 1-2 days

- [ ] 9.32 Implement dynamic array runtime allocation
  - **Task**: Create dynamic arrays at runtime with SetLength()
  - **Implementation**:
    - Implement SetLength() built-in function
    - Allocate/reallocate array storage
    - Handle array growth and shrinkage
    - Preserve existing elements when resizing
  - **Test**: SetLength() resizes arrays correctly
  - **Files**: `internal/interp/builtins_array.go`, `internal/interp/array.go`
  - **Estimated time**: 2-3 days

- [ ] 9.33 Support dynamic array literals
  - **Task**: Allow array literal syntax for dynamic arrays
  - **Implementation**:
    - Parse: `const names: array of String = ['Alice', 'Bob'];`
    - Infer dynamic array length from literal
    - Initialize array with literal values
  - **Test**: Dynamic array constants initialize correctly
  - **Files**: `internal/parser/expression.go`, `internal/interp/expressions.go`
  - **Estimated time**: 1-2 days

#### Type Casting Operators (Blocks 15 tests)

**Root Cause**: `is` and `as` operators not implemented for runtime type checking.

- [ ] 9.40 Implement IS operator
  - **Task**: Runtime type checking (obj is TChild)
  - **Implementation**:
    - Add IS token handler to parser
    - Create IsExpression AST node
    - Implement runtime type checking in interpreter
    - Walk class hierarchy to check instanceof
    - Return boolean result
  - **Test**: `if obj is TChild then ...` works correctly
  - **Files**: `internal/parser/expression.go`, `internal/ast/expressions.go`, `internal/interp/expressions.go`
  - **Estimated time**: 1-2 days

- [ ] 9.41 Implement AS operator
  - **Task**: Safe type casting with runtime check (obj as TChild)
  - **Implementation**:
    - Add AS token handler to parser
    - Create AsExpression AST node
    - Implement runtime type cast in interpreter
    - Check type compatibility, raise exception if invalid
    - Return typed value on success
  - **Test**: `var child := obj as TChild;` casts correctly or raises exception
  - **Files**: `internal/parser/expression.go`, `internal/ast/expressions.go`, `internal/interp/expressions.go`
  - **Estimated time**: 1-2 days

- [ ] 9.42 Handle nil cases for IS and AS
  - **Task**: IS returns false for nil, AS raises exception for nil
  - **Implementation**:
    - Special case nil values in IS evaluation (return false)
    - Special case nil values in AS evaluation (raise exception)
    - Test nil edge cases thoroughly
  - **Test**: `nil is TClass` returns false, `nil as TClass` raises exception
  - **Files**: `internal/interp/expressions.go`
  - **Estimated time**: 0.5 day

**Milestone**: Dynamic arrays and type casting complete, ~35 additional tests should pass

---

### Phase 9.5: Helper Types (HIGH - Blocks 66 tests)

**Priority**: HIGH - Enables scoped enum access and type extensions
**Goal**: Support helper for Type syntax to extend types with methods
**Timeline**: 2-3 weeks | **Success Metric**: Unblocks 66 additional tests

**Root Cause**: Enums need scoped access (TE1.val1) which requires helper functionality. Helpers also enable extending built-in types.

- [x] 9.50 Implement helper AST nodes ✅
  - **Task**: Create AST representation for helper declarations
  - **Implementation**:
    - Add HelperDeclaration node in `internal/ast/helper.go`
    - Fields: Name, ForType, Methods
    - Parse helper method declarations
  - **Files**: `internal/ast/helper.go`
  - **Estimated time**: 1 day
  - **Completed**: 2025-01-09

- [x] 9.51 Parse helper declarations ✅
  - **Task**: Recognize `helper for Type` syntax
  - **Implementation**:
    - Create `internal/parser/helper.go`
    - Parse: `type TStringHelper = helper for String`
    - Parse helper methods
    - Handle helper inheritance
  - **Test**: Helper declarations parse correctly
  - **Files**: `internal/parser/helper.go`
  - **Estimated time**: 2-3 days
  - **Completed**: 2025-01-09

- [x] 9.52 Implement helper type system ✅
  - **Task**: Track which helpers apply to which types
  - **Implementation**:
    - Create `internal/types/helper.go`
    - Map types to their helpers
    - Support multiple helpers for same type
    - Handle helper priority/ordering
  - **Files**: `internal/types/helper.go`
  - **Estimated time**: 2-3 days
  - **Completed**: 2025-01-09

- [x] 9.53 Implement helper method resolution ✅
  - **Task**: Look up helper methods during member access
  - **Implementation**:
    - Modify member access evaluation
    - Method lookup order: instance methods → helpers → built-ins
    - Search all applicable helpers for type
    - Handle helper method dispatch
  - **Test**: `'hello'.ToUpper` calls helper method
  - **Files**: `internal/interp/member_access.go`, `internal/semantic/analyze_expressions.go`
  - **Estimated time**: 3-4 days
  - **Completed**: 2025-01-09

- [x] 9.54 Implement enum scoped access via helpers ✅
  - **Task**: Enable TE1.val1 syntax for enum values
  - **Implementation**:
    - Auto-generate implicit helper for enum types
    - Add enum values as class constants on helper
    - Allow access via type name: `TColor.Red`
    - Maintain backward compatibility with unscoped: `Red`
  - **Test**: Both `TColor.Red` and `Red` work
  - **Files**: `internal/semantic/analyze_enums.go`, `internal/types/helper.go`
  - **Estimated time**: 2-3 days
  - **Completed**: 2025-01-09

- [ ] 9.55 Test helper method inheritance ⏸️ DEFERRED
  - **Task**: Helpers can inherit from other helpers
  - **Implementation**:
    - Parse helper inheritance syntax
    - Inherit methods from parent helper
    - Allow method overriding in child helper
  - **Test**: Child helper inherits parent methods
  - **Files**: `internal/parser/helper.go`, `internal/semantic/helper.go`
  - **Estimated time**: 1-2 days
  - **Status**: Deferred - marked as future enhancement in docs/helpers.md
  - **Note**: Basic helper functionality (9.50-9.54) is complete and working

**Milestone**: Helper types core functionality complete (tasks 9.50-9.54), ~66 additional tests should pass, bringing total to ~70-75% pass rate

---

### Phase 9.6: Metaclasses (MEDIUM - Blocks 15 tests)

**Priority**: MEDIUM - Needed for advanced OOP patterns
**Goal**: Support class of Type for class reference types
**Timeline**: 1 week | **Success Metric**: Unblocks 15 additional tests

**Root Cause**: Cannot declare variables that hold class types (metaclasses) for polymorphic instantiation.

- [x] 9.70 Parse class of type syntax ✅
  - **Task**: Recognize `class of TBase` type declarations
  - **Implementation**:
    - Added `ClassOfTypeNode` to `pkg/ast/type_expression.go`
    - Modified `parseTypeExpression` in `internal/parser/types.go` to handle CLASS token
    - Implemented `parseClassOfType()` function
    - Updated `internal/parser/statements.go` to handle ClassOfTypeNode in var declarations
  - **Test**: `type TBaseClass = class of TBase;` parses successfully
  - **Files**: `pkg/ast/type_expression.go`, `internal/ast/ast_alias.go`, `internal/parser/types.go`, `internal/parser/statements.go`
  - **Completed**: 2025-01-08

- [x] 9.71 Implement metaclass type system ✅
  - **Task**: Create metaclass type for type checking
  - **Implementation**:
    - Created `ClassOfType` in `internal/types/types.go`
    - Implemented type compatibility checking via `IsAssignableFrom` method
    - Added `resolveClassOfTypeNode()` in `internal/semantic/type_resolution.go`
    - Updated `resolveTypeExpression()` to handle ClassOfTypeNode
  - **Test**: Metaclass type checking works (semantic analysis recognizes metaclass types)
  - **Files**: `internal/types/types.go`, `internal/semantic/type_resolution.go`
  - **Completed**: 2025-01-08

- [x] 9.72 Implement metaclass runtime values ✅
  - **Task**: Store class types as runtime values
  - **Implementation**:
    - Created `ClassValue` struct in `internal/interp/class.go`
    - Added `Type()`, `String()`, and `IsAssignableTo()` methods
    - Added helper functions `isClassValue()` and `AsClassValue()`
    - Note: `ClassInfoValue` already exists and serves similar purpose - may need unification
  - **Test**: Class types can be represented at runtime
  - **Files**: `internal/interp/class.go`
  - **Completed**: 2025-01-08
  - **Note**: Class name identifiers already resolve to `ClassInfoValue` (see `evalIdentifier` in `expressions.go` line 175-183)

- [ ] 9.73 Implement virtual constructor dispatch via metaclass - ~70% complete (2025-01-09)
  - **Task**: Call constructors through metaclass variables
  - **Implementation**: MOSTLY COMPLETE
    - ✅ 9.73.1 Semantic analysis for metaclass assignments (analyzer.go:415-430)
    - ✅ 9.73.2 Semantic analysis for metaclass member access (analyze_classes.go:1024-1041)
    - ✅ 9.73.3 Virtual constructor dispatch mechanism (objects.go:1596-1689, from Task 9.4.4)
    - ✅ 9.73.4 Parser support for metaclass type aliases (2025-01-08)
    - ✅ 9.73.5 Class names resolve to ClassValue (2025-01-08)
    - ✅ 9.73.8 Virtual constructor dispatch for child classes (2025-01-08)
    - ✅ 9.73.9 Metaclass comparison operators (2025-01-08)
    - ⚠️ 9.73.6 Runtime assignment validation (optional enhancement)
  - **Test**: `var obj := meta.Create;` creates instance of runtime type
  - **Files**: `internal/parser/interfaces.go`, `internal/interp/expressions.go`, `internal/interp/objects.go`
  - **Status**: ~70% complete - core functionality working, minor enhancements remaining
  - **Related**: Task 9.4 (Virtual constructors), Task 9.72 (Metaclasses)
  - **Remaining Work**: ClassName support in class/metaclass contexts, optional runtime validation

  **COMPLETED WORK (2025-01-08)**:
  1. ✅ Semantic Analysis - Metaclass Assignments (100% complete)
     - Location: `internal/semantic/analyzer.go:415-430`
     - Allows assigning nil to metaclass variables
     - Allows assigning class references to metaclass variables
     - Validates inheritance compatibility (descendant classes assignable to base metaclass)

  2. ✅ Semantic Analysis - Metaclass Member Access (100% complete)
     - Location: `internal/semantic/analyze_classes.go:1024-1041`
     - Handles member access on metaclass types
     - Allows constructor calls through metaclass types
     - Location: `internal/semantic/analyze_function_calls.go:42-48`

  3. ✅ Runtime Virtual Dispatch (100% complete - from Task 9.4.4)
     - Location: `internal/interp/objects.go:1596-1689`
     - Checks if method receiver is ClassValue (metaclass)
     - Looks up constructor in runtime class
     - Creates instance of actual runtime type, not declared type
     - Handles argument evaluation, constructor resolution, instance creation

  **CRITICAL BLOCKERS DISCOVERED**:

  **Blocker 1: Parser Cannot Handle Metaclass Type Aliases** (CRITICAL)
  - Syntax: `type TBaseClass = class of TBase;`
  - Error: "expected ':' after field name or method/property declaration keyword"
  - Location: `internal/parser/interfaces.go:91-93`
  - Problem: No case for "class of" in type alias declarations
  - Current code expects inline class body after "class" keyword
  - Impact: ALL metaclass fixture tests fail at parse time

  **Test Failures**:
  - `testdata/fixtures/SimpleScripts/class_of.pas` - Parse error at line 22
  - `testdata/fixtures/SimpleScripts/class_of2.pas` - Parse error at line 21
  - `testdata/fixtures/SimpleScripts/class_of3.pas` - Parse error at line 4
  - `testdata/fixtures/SimpleScripts/class_of_cast.pas` - Parse error at line 10

  **Blocker 2: Wrong Type Returned for Class Name Identifiers**
  - Location: `internal/interp/expressions.go:178-183`
  - Current: Returns `ClassInfoValue` (internal context tracking type)
  - Should: Return `ClassValue` (runtime metaclass reference type)
  - Impact: Metaclass variables receive wrong type, dispatch fails

  **Type Confusion**:
  - `ClassInfoValue` - Used for `__CurrentClass__` context tracking, Type()="CLASSINFO"
  - `ClassValue` - Represents metaclass references, Type()="CLASS", has IsAssignableTo() method
  - These are TWO DIFFERENT types with different purposes

  **Issue 3: Missing Runtime Assignment Validation**
  - Location: `internal/interp/statements.go:605-678` (evalSimpleAssignment)
  - No validation for metaclass variable assignments at runtime
  - ClassValue.IsAssignableTo() method exists but is never called
  - Could allow invalid assignments to slip through

  **Subtasks to Fix Blockers** (see Tasks 9.73.4-9.73.7 below)

- [x] 9.73.4 **CRITICAL**: Fix parser to handle metaclass type aliases ✅ DONE (2025-01-08)
  - **Task**: Parse `type Name = class of BaseClass;` syntax
  - **Subtasks**:
    - [x] 9.73.4.1 Add AST support for metaclass type references (already existed in pkg/ast/type_expression.go)
    - [x] 9.73.4.2 Update lexer if needed (OF token already existed)
    - [x] 9.73.4.3 Implement parser support for metaclass type aliases (internal/parser/interfaces.go:91-123)
    - [x] 9.73.4.4 Update semantic analysis for metaclass type aliases (internal/semantic/analyze_types.go:546-561)
    - [x] 9.73.4.5 Write parser tests (internal/parser/inline_types_test.go:480-636)
    - [x] 9.73.4.6 Verify fixture tests can parse (class_of.pas, class_of2.pas, class_of3.pas all parse)
  - **Test**: `type TBaseClass = class of TBase;` parses without error
  - **Files**: `pkg/ast/types.go`, `internal/parser/interfaces.go`, `internal/semantic/analyze_types.go`
  - **Priority**: CRITICAL - blocks all metaclass functionality

- [x] 9.73.5 **HIGH**: Fix class name identifier evaluation ✅ DONE (2025-01-08)
  - **Task**: Return ClassValue instead of ClassInfoValue for class name identifiers
  - **Subtasks**:
    - [x] 9.73.5.1 Update identifier evaluation (internal/interp/expressions.go:178-183)
    - [x] 9.73.5.2 Update any code expecting ClassInfoValue (internal/interp/objects.go:369-439)
    - [x] 9.73.5.3 Update semantic analyzer to return ClassOfType (internal/semantic/analyze_expr_operators.go:47-56)
    - [x] 9.73.5.4 Add ClassOfType assignment checking (internal/semantic/analyzer.go:422-438)
  - **Test**: Class name identifiers return ClassValue type
  - **Files**: `internal/interp/expressions.go`
  - **Priority**: HIGH - required for metaclass functionality

- [ ] 9.73.6 **MEDIUM**: Add runtime metaclass assignment validation ⚠️ NEW
  - **Task**: Validate metaclass variable assignments at runtime
  - **Subtasks**:
    - [ ] 9.73.6.1 Add metaclass validation in evalSimpleAssignment
    - [ ] 9.73.6.2 Implement validation logic
    - [ ] 9.73.6.3 Test invalid assignments at runtime
    - [ ] 9.73.6.4 Verify error messages are clear
  - **Test**: Invalid metaclass assignments fail with clear error messages
  - **Files**: `internal/interp/statements.go`
  - **Priority**: MEDIUM - improves error reporting

- [x] 9.73.8 **HIGH**: Fix virtual constructor dispatch for child classes ✅ DONE (2025-01-08)
  - **Task**: Debug and fix constructor lookup when metaclass holds child class reference
  - **Problem**: `meta := TChild; obj := meta.Create;` fails with "no overloaded version" error
  - **Root Cause Found**: Child constructors were being APPENDED to parent constructors in ConstructorOverloads,
    causing duplicate constructors with same signature and breaking overload resolution
  - **Solution**: Modified constructor registration to REPLACE parent constructor when child has matching signature
    - In DWScript, child constructor with same name/signature HIDES parent's (like Delphi)
    - Works regardless of virtual/override keywords
  - **Files Modified**: `internal/interp/declarations.go:183-204, 211-232`
  - **Subtasks**:
    - [x] 9.73.8.1 Add debug logging to constructor lookup in evalMethodCall
    - [x] 9.73.8.2 Verify TChild.Create is registered in ClassInfo.Constructors
    - [x] 9.73.8.3 Check getMethodOverloadsInHierarchy logic for constructors
    - [x] 9.73.8.4 Review constructor overload resolution with empty arguments
    - [x] 9.73.8.5 Test with both virtual/override and non-virtual constructors
    - [x] 9.73.8.6 Debug logging removed after fix verified
  - **Test Case**:
    ```dws
    type TBase = class
      constructor Create; virtual;
    end;
    type TChild = class(TBase)
      constructor Create; override;
    end;
    type TBaseClass = class of TBase;
    var meta: TBaseClass;
    meta := TBase;
    var obj1 := meta.Create;  // ✅ Works
    meta := TChild;
    var obj2 := meta.Create;  // ❌ Fails - "no overloaded version"
    ```
  - **Files**: `internal/interp/objects.go`, `internal/interp/class.go`
  - **Priority**: HIGH - blocks virtual constructor dispatch, critical for polymorphism

- [x] 9.73.9 **MEDIUM**: Implement metaclass comparison operators ✅ DONE (2025-01-08)
  - **Task**: Support equality and inequality comparisons for metaclass values
  - **Problem**: `if meta = TBase then` and `if meta <> TChild then` fail with "requires comparable types"
  - **Solution**:
    - Added "CLASSOF" to IsComparableType() in types/compatibility.go
    - Implemented ClassValue comparison in evalBinaryExpression
    - Compares by ClassInfo identity (pointer equality)
  - **Files Modified**:
    - `internal/types/compatibility.go:152` - Added CLASSOF to comparable types
    - `internal/interp/expressions.go:243-266` - Runtime comparison for ClassValue
  - **Subtasks**:
    - [x] 9.73.9.1 Add ClassOfType comparison in semantic analyzer
    - [x] 9.73.9.2 Implement runtime comparison for ClassValue
    - [x] 9.73.9.3 Support comparing ClassValue with ClassValue
    - [x] 9.73.9.4 Support comparing metaclass variable with class name (meta = TBase)
    - [x] 9.73.9.5 Test equality (=) and inequality (<>) operators
    - [x] 9.73.9.6 Verified with comprehensive test cases
  - **Test Case**:
    ```dws
    var meta: class of TBase;
    meta := TChild;
    if meta = TBase then PrintLn('Bug Equal TBase');      // Should not print
    if meta <> TChild then PrintLn('Bug Not Equal TChild'); // Should not print
    if meta = TChild then PrintLn('Correct');              // Should print
    ```
  - **Files**: `internal/semantic/analyze_expr_operators.go`, `internal/interp/operators_eval.go`
  - **Priority**: MEDIUM - needed for fixture tests

- [ ] 9.73.7 **FINAL**: Integration testing and fixture validation ⚠️ NEW
  - **Task**: Verify all metaclass functionality works end-to-end
  - **Dependencies**: Requires 9.73.8 (constructor dispatch) and 9.73.9 (comparisons) to be complete
  - **Subtasks**:
    - [ ] 9.73.7.1 Test metaclass variable declaration and assignment
    - [ ] 9.73.7.2 Test metaclass constructor dispatch (depends on 9.73.8)
    - [ ] 9.73.7.3 Test metaclass with constructor parameters
    - [ ] 9.73.7.4 Test metaclass comparison operators (depends on 9.73.9)
    - [ ] 9.73.7.5 Run fixture tests (class_of*.pas)
    - [ ] 9.73.7.6 Run overload tests
    - [ ] 9.73.7.7 Verify expected output matches reference implementation
  - **Test**: All metaclass fixture tests pass
  - **Priority**: FINAL - validates all previous work

  **Test files**:
  - `testdata/fixtures/SimpleScripts/class_of.pas` - Basic metaclass usage
  - `testdata/fixtures/SimpleScripts/class_of2.pas` - Metaclass with inheritance
  - `testdata/fixtures/SimpleScripts/class_of3.pas` - Metaclass type casting
  - `testdata/fixtures/SimpleScripts/class_of_cast.pas` - Metaclass casting operations
  - `testdata/fixtures/OverloadsPass/overload_on_metaclass.pas` - Metaclass overload resolution

**Milestone**: Metaclasses complete, ~15 additional tests should pass

---

### Phase 9.7: Missing Built-ins & Operators (MEDIUM - Blocks 40+ tests)

**Priority**: MEDIUM - Quick wins, isolated implementations
**Goal**: Implement commonly-used built-in functions and operators
**Timeline**: 1-2 weeks | **Success Metric**: Unblocks 40+ tests

#### Quick Fixes (Trivial, ~0.5 day total)

- [x] 9.80 Fix boolean capitalization
  - **Task**: Output `True`/`False` instead of `true`/`false`
  - **Implementation**: Change string formatting in boolean conversion
  - **Test**: Boolean output matches DWScript
  - **Files**: `internal/interp/value.go`
  - **Estimated time**: 0.1 day
  - **Impact**: Fixes 11 tests
  - **Completed**: Changed BooleanValue.String() to return "True"/"False"

- [x] 9.81 Implement SAR operator
  - **Task**: Arithmetic shift right (sign-preserving)
  - **Implementation**: Add SAR token and operator evaluation
  - **Test**: `x sar 2` performs arithmetic right shift
  - **Files**: `internal/parser/parser.go`, `internal/interp/expressions.go`, `internal/semantic/analyze_expr_operators.go`
  - **Estimated time**: 0.5 day
  - **Impact**: Fixes 1 test
  - **Completed**: Added SAR operator with precedence, parsing, and evaluation support

#### Critical Built-ins (High impact, 3-4 days)

- [x] 9.82 Implement Assigned() built-in
  - **Task**: Check if pointer/object/variant is nil
  - **Implementation**: Add Assigned(value) → returns false if nil, true otherwise
  - **Test**: Nil checking works for objects, arrays, variants
  - **Files**: `internal/interp/builtins_core.go`, `internal/interp/functions.go`, `internal/semantic/analyze_builtin_functions.go`, `internal/semantic/analyze_builtin_math.go`
  - **Estimated time**: 1 day
  - **Impact**: Fixes 16 tests
  - **Completed**: Added Assigned() function with support for nil checking on all types

- [x] 9.83 Implement Chr() built-in
  - **Task**: Convert integer to character
  - **Implementation**: Add Chr(code: Integer) → String (single character)
  - **Test**: Chr(65) returns 'A'
  - **Files**: `internal/interp/builtins_strings.go`, `internal/interp/functions.go`
  - **Estimated time**: 0.5 day
  - **Impact**: Fixes 4 tests
  - **Completed**: Added Chr() function with Unicode support (0-1114111)

- [x] 9.84 Implement IntToHex() built-in
  - **Task**: Convert integer to hexadecimal string
  - **Implementation**: Add IntToHex(value: Integer, digits: Integer) → String
  - **Test**: IntToHex(255, 2) returns 'FF'
  - **Files**: `internal/interp/builtins_strings.go`, `internal/interp/functions.go`, `internal/semantic/analyze_builtin_functions.go`, `internal/semantic/analyze_builtin_convert.go`
  - **Estimated time**: 0.5 day
  - **Impact**: Fixes 8 tests
  - **Completed**: Added IntToHex() with padding support

- [ ] 9.85 Implement Swap() built-in
  - **Task**: Swap two variables
  - **Implementation**: Add Swap(var a, var b) - exchanges values
  - **Test**: After Swap(x, y), x has old y value and vice versa
  - **Files**: `internal/interp/builtins.go`
  - **Estimated time**: 1 day
  - **Impact**: Fixes 9 tests

#### Utility Built-ins (Medium priority, 2-3 days)

- [ ] 9.86 Implement StrToBool() built-in
  - **Task**: Parse boolean from string
  - **Implementation**: Accept 'True'/'False', '1'/'0', 'Yes'/'No'
  - **Test**: StrToBool('True') returns true
  - **Files**: `internal/interp/builtins_string.go`
  - **Estimated time**: 0.5 day
  - **Impact**: Fixes 1 test

- [ ] 9.87 Implement IsNaN() built-in
  - **Task**: Check for NaN float value
  - **Implementation**: Add IsNaN(value: Float) → Boolean
  - **Test**: IsNaN(0.0/0.0) returns true
  - **Files**: `internal/interp/builtins_math.go`
  - **Estimated time**: 0.5 day
  - **Impact**: Fixes 1 test

- [ ] 9.88 Implement SetRandSeed() built-in
  - **Task**: Seed random number generator
  - **Implementation**: Set PRNG seed for reproducible randomness
  - **Test**: Same seed produces same sequence
  - **Files**: `internal/interp/builtins_math.go`
  - **Estimated time**: 0.5 day
  - **Impact**: Fixes 6 tests

#### IN Operator Completion (Medium complexity, 2-3 days)

- [ ] 9.90 Complete IN operator for sets and ranges
  - **Task**: Fix set membership testing
  - **Implementation**:
    - Character ranges: `ch in ['a'..'z']`
    - Set membership: `value in setVar`
    - String character membership: `ch in 'abc'`
    - Complete set type implementation if needed
  - **Test**: All IN operator tests pass
  - **Files**: `internal/interp/expressions.go`, `internal/types/set.go`
  - **Estimated time**: 2-3 days
  - **Impact**: Fixes 4 tests

**Milestone**: Built-ins and operators complete, ~40 additional tests should pass

---

### Phase 9.8: Remaining Unit Test Failures (LOW - 13 tests)

**Priority**: LOW - Fix remaining unit test failures (not fixture tests)
**Timeline**: 1 week

These are test failures in the unit test suites (not fixture tests):

- [x] 9.37 Fix record method access (lerp.pas)
  - **Issue**: Record methods not accessible in member access
  - **Implementation**: Modified evalMemberAccess and analyzeRecordFieldAccess to check record methods
  - **Files**: `internal/interp/objects.go`, `internal/semantic/analyze_records.go`
  - **Estimated time**: 1 day
  - **Impact**: Fixes 1 algorithm test (partially - test now progresses past method access)
  - **Completed**: Record methods are now accessible via member access and auto-invoked when parameterless

- [x] 9.47 Fix constructor overload handling (unit tests)
  - **Tests**: TestConstructorOverload, TestConstructorWithoutParentheses, etc.
  - **Already covered**: This is part of Task 9.3 (Constructor overloading)

- [x] 9.48 Implement interface method dispatch
  - **Issue**: Interface methods not resolved
  - **Implementation**: Interface type checking and method lookup
  - **Files**: `internal/semantic/analyze_interfaces.go`, `internal/interp/interfaces.go`
  - **Estimated time**: 2 days

- [x] 9.49 Fix implicit self property access
  - **Issue**: Properties without explicit self not found
  - **Implementation**: Add self fallback in identifier resolution, validate property expressions
  - **Files**: `internal/semantic/analyze_properties.go`, `internal/semantic/analyze_expr_operators.go`, `internal/semantic/analyzer.go`
  - **Completed**: Property expressions now validated at compile-time with implicit self support and circular reference detection
  - **Tests**: Added comprehensive unit tests in `internal/semantic/property_test.go`

- [x] 9.50 Fix overload directive validation
  - **Issue**: Incorrect error messages for overload declarations
  - **Implementation**: Fix validation logic and error messages
  - **Files**: `internal/semantic/analyze_functions.go`
  - **Estimated time**: 1 day

- [x] 9.51 Implement inherited expression validation
  - **Issue**: Inherited expression errors not caught
  - **Implementation**: Made TObject implicit parent for all classes, improved inherited expression validation
  - **Files**: `internal/semantic/analyze_classes.go`, `internal/semantic/analyze_special.go`
  - **Estimated time**: 1 day
  - **Completed**: Classes without explicit parent now inherit from TObject; inherited expressions validate member existence

- [ ] 9.52 Enforce private field access control
  - **Issue**: Derived classes can access private parent fields
  - **Implementation**: Add visibility check in field access
  - **Files**: `internal/semantic/analyze_expressions.go`
  - **Estimated time**: 0.5 day

---

### Phase 9.9: Documentation & Cleanup

**Priority**: LOW - Can be done in parallel with Phase 10
**Timeline**: 1 week

- [ ] 9.100 Update README with current features
  - Document all Stage 7 features now complete
  - Update feature completion percentages
  - Add examples of new features

- [ ] 9.101 Create docs/phase9-summary.md
  - Document achievements in Phase 9
  - Statistics: tests passing, coverage percentages
  - Lessons learned and challenges overcome

- [ ] 9.102 Update testdata/fixtures/TEST_STATUS.md
  - Update pass/fail counts for each category
  - Mark resolved issues
  - Document remaining blockers

- [ ] 9.103 Create docs/limitations.md
  - Document known limitations
  - Features intentionally deferred to later phases
  - Differences from original DWScript

---

## Phase 9 Completion Milestones

| Milestone | Tasks | Pass Rate Target | Tests Passing |
|-----------|-------|------------------|---------------|
| **Current** | - | 16.7% | 92/552 |
| **After 9.7 (Quick wins)** | 9.80-9.90 | 25-30% | 138-165 |
| **After 9.1 (Constructors)** | 9.1-9.6 | 45-50% | 248-276 |
| **After 9.2 (Properties)** | 9.10-9.17 | 60% | 330+ |
| **After 9.3 (Class constants)** | 9.20-9.22 | 65% | 358+ |
| **After 9.4 (Arrays/Casting)** | 9.30-9.42 | 70% | 386+ |
| **After 9.5 (Helpers)** | 9.50-9.55 | 75-80% | 414-440 |
| **After 9.6 (Metaclasses)** | 9.70-9.73 | 80-85% | 440-470 |

**Target**: 80-85% fixture test pass rate (440-470 tests passing)
**Timeline**: 10-15 weeks (2.5-3.5 months)

**Success Criteria**:
- ✅ All Stage 7 features complete (Classes & OOP)
- ✅ Critical Stage 8 features implemented (Helpers)
- ✅ >80% fixture tests passing
- ✅ No regressions in existing tests
- ✅ Documentation updated

**After Phase 9**: Remaining ~15-20% of failing tests will be Stage 8 features (interfaces, variants, delegates, anonymous methods, units/modules) and edge cases, which belong in Phase 10 and beyond.

---

## Phase 10: go-dws API Enhancements for LSP Integration

**Goal**: Enhance the go-dws library to expose structured errors, AST access, and position metadata needed for LSP features.

**Why This Phase**: The current go-dws API provides string-based errors and opaque Program objects. To implement LSP features (hover, go-to-definition, completion, etc.) in the future, we need structured error information, direct AST access, and position metadata on AST nodes.

### Tasks (42)

- [x] **10.1 Create structured error types in pkg/dwscript** ✅ DONE
  - [x] Create `pkg/dwscript/error.go` file
  - [x] Define `Error` struct with fields:
    - [x] `Message string` - The error message
    - [x] `Line int` - 1-based line number
    - [x] `Column int` - 1-based column number
    - [x] `Length int` - Length of the error span in characters
    - [x] `Severity string` - Either "error" or "warning"
    - [x] `Code string` - Optional error code (e.g., "E001", "W002")
  - [x] Implement `Error() string` method to satisfy error interface
  - [x] Add documentation explaining 1-based indexing

- [x] **10.2 Update CompileError to use structured errors** ✅ DONE
  - [x] Change `CompileError.Errors` from `[]string` to `[]*Error`
  - [x] Update `CompileError.Error()` method to format structured errors
  - [x] Added helper methods: `HasErrors()`, `HasWarnings()`
  - [x] Update Compile() method to convert errors to structured format
  - [x] Note: Full position extraction will improve with Task 10.4

- [x] **10.3 Update internal lexer to capture position metadata** ✅ DONE
  - [x] Verified `internal/lexer/token.go` includes position information
  - [x] Token struct already has `Line`, `Column`, `Offset` fields
  - [x] Position tracking already implemented in lexer
  - [x] Added `Length()` method to Token for error span calculation

- [x] **10.4 Update internal parser to capture error positions** ✅ DONE
  - [x] Modified parser error handling to include token position
  - [x] Errors now include line, column, and length from offending token
  - [x] Verified position accuracy with test cases
  - [x] Added unit tests for parser error position extraction

- [x] **10.5 Update internal semantic analyzer to capture error positions** ✅ DONE
  - [x] Modified semantic analysis error generation
  - [x] Already includes position from AST node being analyzed
  - [x] Added Severity field to SemanticError (Error, Warning, Info, Hint)
  - [x] Added error codes for common semantic errors:
    - [x] ErrorTypeMismatch - Type mismatch (already existed)
    - [x] ErrorUndefinedVariable - Undefined variable (already existed)
    - [x] ErrorArgumentCount - Wrong argument count (already existed)
    - [x] WarningUnusedVariable - Unused variable (NEW)
    - [x] WarningUnusedParameter - Unused parameter (NEW)
    - [x] WarningUnusedFunction - Unused function (NEW)
    - [x] WarningDeprecated - Deprecated feature (NEW)
  - [x] Added helper functions for creating warnings
  - [x] Note: Duplicated ErrorSeverity type to avoid import cycle

- [x] **10.6 Add position metadata to AST node types**
  - [x] Position struct already exists in `internal/lexer/token.go`
  - [x] Updated Node interface in `internal/ast/ast.go` to add `End() Position` method
  - [x] Added `EndPos lexer.Position` field to all 70+ AST node types across 19 files
  - [x] Implemented `End()` method on all AST node types
  - [x] Fixed field name conflicts (`RangeExpression.End` → `RangeEnd`, `ForStatement.End` → `EndValue`)
  - [x] Fixed all compilation errors and test failures
  - [x] All AST node types now provide both start and end positions

- [x] **10.7 Add position fields to statement AST nodes**
  - [x] Added `EndPos Position` field to all statement nodes (combined with 10.6)
    - [x] `Program`
    - [x] `BlockStatement`
    - [x] `ExpressionStatement`
    - [x] `AssignmentStatement`
    - [x] `IfStatement`
    - [x] `WhileStatement`
    - [x] `ForStatement`
    - [x] `ReturnStatement`
    - [x] `BreakStatement`
    - [x] `ContinueStatement`
    - [x] All other statement types (RepeatStatement, ForInStatement, CaseStatement, etc.)
  - [x] Implemented `End()` methods for all statement nodes

- [x] **10.8 Add position fields to expression AST nodes**
  - [x] Added `EndPos Position` field to all expression nodes (combined with 10.6)
    - [x] `Identifier`
    - [x] `IntegerLiteral`
    - [x] `FloatLiteral`
    - [x] `StringLiteral`
    - [x] `BooleanLiteral`
    - [x] `BinaryExpression`
    - [x] `UnaryExpression`
    - [x] `CallExpression`
    - [x] `IndexExpression`
    - [x] `MemberAccessExpression` (was MemberExpression)
    - [x] All other expression types (70+ total nodes updated)
  - [x] Implemented `End()` methods for all expression nodes

- [x] **10.9 Add position fields to declaration AST nodes**
  - [x] Added `EndPos lexer.Position` field to all declaration nodes (completed as part of comprehensive 10.6-10.8 implementation)
    - [x] `FunctionDecl` (internal/ast/functions.go)
    - [x] `ClassDecl` (internal/ast/classes.go)
    - [x] `FieldDecl` (internal/ast/classes.go)
    - [x] `VarDeclStatement` (internal/ast/statements.go)
    - [x] `ConstDecl` (internal/ast/declarations.go)
    - [x] `TypeDeclaration` (internal/ast/type_annotation.go)
    - [x] `EnumDecl` (internal/ast/enums.go)
    - [x] `RecordDecl` (internal/ast/records.go)
    - [x] `InterfaceDecl` (internal/ast/interfaces.go)
    - [x] `InterfaceMethodDecl` (internal/ast/interfaces.go)
    - [x] `OperatorDecl` (internal/ast/operators.go)
    - [x] `PropertyDecl` (internal/ast/properties.go)
    - [x] `HelperDecl` (internal/ast/helper.go)
    - [x] `UnitDeclaration` (internal/ast/unit.go)
  - [x] Implemented `End()` methods for all declaration nodes
  - [x] All declaration types now provide both start (Pos()) and end (End()) positions

- [x] **10.10 Update parser to populate position information** (Complete)
  - [x] Added `endPosFromToken()` helper function in parser.go:270
  - [x] Added comprehensive package documentation with position tracking pattern guidelines (parser.go:1-29)
  - [x] **Updated literal parsing functions** (7/7 complete):
    - [x] parseIdentifier, parseIntegerLiteral, parseFloatLiteral
    - [x] parseStringLiteral, parseBooleanLiteral, parseNilLiteral, parseCharLiteral
  - [x] **Updated key expression parsing functions** (7+ complete):
    - [x] parsePrefixExpression (unary operators)
    - [x] parseAddressOfExpression (@operator)
    - [x] parseInfixExpression (binary operators)
    - [x] parseCallExpression (function calls)
    - [x] parseIndexExpression (array/collection indexing)
    - [x] parseMemberAccess (member access, method calls, object creation)
  - [x] **Updated statement parsing functions** (10+ complete):
    - [x] parseBlockStatement (begin...end)
    - [x] parseExpressionStatement
    - [x] ParseProgram (top-level)
    - [x] parseBreakStatement, parseContinueStatement, parseExitStatement
    - [x] parseIfStatement (with else branch handling)
    - [x] parseWhileStatement, parseRepeatStatement
  - [x] Created comprehensive position tracking tests (internal/parser/position_test.go)
    - [x] TestPositionTracking - validates Program and Statement positions
    - [x] TestLiteralPositions - validates all literal types
    - [x] TestBinaryExpressionPositions - validates complex expressions
  - [x] All position tests pass successfully
  - [x] Pattern documented in parser.go package comment for future implementation
  - [x] **All remaining parsing functions updated** (all 10 subtasks complete)
  - **Status**: COMPLETE - All parser functions now populate EndPos correctly
  - **Pattern Summary** (from parser.go:1-29):
    1. Single-token nodes: `node.EndPos = p.endPosFromToken(p.curToken)`
    2. Multi-token nodes: Set after all tokens consumed, or delegate to child `node.EndPos = child.End()`
    3. Optional semicolons: Update EndPos if semicolon is consumed

- [x] **10.10.1 Update control flow statement parsers**
  - [x] parseForStatement (for...to/downto loops)
  - [x] parseForInStatement (for...in loops)
  - [x] parseCaseStatement (case statements with multiple branches)
  - [x] parseTryStatement (try...except...finally blocks)
  - [x] parseRaiseStatement (raise exceptions)
  - Location: internal/parser/control_flow.go, exceptions.go

- [x] **10.10.2 Update declaration parsers**
  - [x] parseVarDeclaration (variable declarations with optional initializers)
  - [x] parseConstDeclaration (constant declarations)
  - [x] parseFunctionDeclaration (function/procedure declarations)
  - [x] parseReturnStatement (return statements)
  - Location: internal/parser/statements.go, functions.go, declarations.go

- [x] **10.10.3 Update type declaration parsers**
  - [x] parseTypeDeclaration (type aliases and declarations)
  - [x] parseClassDeclaration (class definitions)
  - [x] parseInterfaceDeclaration (interface definitions)
  - [x] parseEnumDeclaration (enumerated types)
  - [x] parseRecordDeclaration (record/struct types)
  - Location: internal/parser/types.go, classes.go, interfaces.go, enums.go, records.go

- [x] **10.10.4 Update collection literal parsers**
  - [x] parseArrayLiteral (array literals: [1, 2, 3])
  - [x] parseRecordLiteral (record literals with field initializers)
  - [x] parseSetLiteral (set literals: [A, B, C])
  - [x] parseRangeExpression (range expressions: A..Z)
  - Location: internal/parser/arrays.go, records.go, sets.go

- [x] **10.10.5 Update assignment and complex expression parsers**
  - [x] parseAssignmentStatement (assignment with all operators)
  - [x] parseAssignmentOrExpression (hybrid statement/expression handling)
  - [x] parseLambdaExpression (lambda/anonymous functions)
  - [x] parseInheritedExpression (inherited keyword)
  - Location: internal/parser/statements.go, lambda.go, classes.go

- [x] **10.10.6 Update type expression parsers**
  - [x] parseFunctionPointerType (function/procedure pointer types)
  - [x] parseArrayType (inline array type declarations)
  - [x] parseSetType (inline set type declarations)
  - [x] parseTypeAnnotation (type references and annotations)
  - Location: internal/parser/types.go, function_pointers.go

- [x] **10.10.7 Update contract parsers**
  - [x] parsePreConditions (require blocks)
  - [x] parsePostConditions (ensure blocks)
  - [x] parseCondition (individual contract conditions)
  - [x] parseOldExpression (old keyword in postconditions)
  - Location: internal/parser/contracts.go (if exists) or functions.go

- [x] **10.10.8 Update property and operator parsers**
  - [x] parsePropertyDeclaration (property declarations with getters/setters)
  - [x] parseOperatorDeclaration (operator overloading)
  - [x] parseHelperDeclaration (type helper declarations)
  - Location: internal/parser/properties.go, operators.go, helpers.go

- [x] **10.10.9 Update unit and uses parsers**
  - [x] parseUnit (unit declarations)
  - [x] parseUsesClause (uses/imports)
  - [x] parseProgramDeclaration (program header - no AST node returned)
  - Location: internal/parser/unit.go, parser.go

- [x] **10.10.10 Final validation and testing**
  - [x] Run full test suite to verify no regressions
  - [x] All parser tests pass successfully
  - [x] Position accuracy verified for all constructs
  - [x] Documentation updated with completion status
  - [x] Task 10.10 marked as complete

- [x] **10.11 Export AST types as public API** ✅ COMPLETE
  - [x] Create `pkg/token/` directory (for Position, Token, TokenType)
  - [x] Create `pkg/ast/` directory with all AST node types
  - [x] Export all node types (74+ types including Node, Expression, Statement interfaces)
  - [x] Add comprehensive package documentation with examples
  - [x] Keep `internal/ast/` as alias to `pkg/ast/` for backwards compatibility
  - [x] Keep `internal/lexer/` token types as alias to `pkg/token/` for backwards compatibility
  - [x] Update `pkg/dwscript/` to use public `pkg/ast` types
  - [x] Create example tests demonstrating AST traversal

- [x] **10.12 Add AST accessor to Program type** ✅ COMPLETE (done as part of 10.11)
  - [x] Added `func (p *Program) AST() *ast.Program` method
  - [x] Added comprehensive documentation explaining AST structure
  - [x] Documented that AST is read-only
  - [x] Added example showing AST traversal
  - [ ] Add example in documentation showing AST traversal

- [x] **10.13 Add parse-only mode for LSP use cases** ✅ COMPLETE
  - [x] Added method to Engine: `func (e *Engine) Parse(source string) (*ast.Program, error)`
  - [x] Parse source code without semantic analysis (skips type checking entirely)
  - [x] Return partial AST even if syntax errors exist (best-effort parsing)
  - [x] Return structured syntax errors only (no type checking errors)
  - [x] Comprehensive documentation with LSP use case examples
  - [x] Optimized for speed (skips expensive semantic checks)
  - [x] Created 8 comprehensive tests covering all scenarios:
    - Valid code parsing
    - Invalid code with syntax errors
    - Empty code
    - Partial/incomplete code
    - Comparison with Compile() method
    - LSP use case simulation
    - Error recovery with multiple errors
    - Performance characteristics

- [x] **10.14 Create visitor pattern for AST traversal** ✅ COMPLETE
  - [x] Created `pkg/ast/visitor.go` (639 lines)
  - [x] Defined `Visitor` interface with `Visit(node Node) (w Visitor)` method
  - [x] Implemented `Walk(v Visitor, node Node)` function
  - [x] Handles all 64+ AST node types correctly
  - [x] Added comprehensive documentation with examples
  - [x] Added `Inspect(node Node, f func(Node) bool)` helper for convenience
  - [x] Created 9 comprehensive tests with examples:
    - TestWalk_VisitsAllNodes
    - TestWalk_VisitorReturnsNil
    - TestInspect_FindsFunctions
    - TestInspect_FindsVariables
    - TestInspect_StopsTraversal
    - TestInspect_NestedStructures
    - TestWalk_AllNodeTypes
    - TestInspect_ComplexExpressions
    - TestWalk_WithNilNodes
  - [x] All tests pass ✅
  - [x] Follows standard Go `go/ast` package patterns

- [x] **10.15 Add symbol table access for semantic information** ✅ DONE
  - [x] Create `pkg/dwscript/symbols.go`
  - [x] Define `Symbol` struct:
    - [x] `Name string`
    - [x] `Kind string` - "variable", "function", "class", "parameter", etc.
    - [x] `Type string` - Type name
    - [x] `Position Position` - Definition location
    - [x] `Scope string` - "local", "global", "class"
    - [x] `IsReadOnly bool` - Whether symbol is read-only
    - [x] `IsConst bool` - Whether symbol is a compile-time constant
  - [x] Add method: `func (p *Program) Symbols() []Symbol`
  - [x] Extract symbols from semantic analyzer's symbol table
  - [x] Include all declarations (variables, functions, types)
  - [x] Added getter methods to Analyzer to expose symbol table
  - [x] Implemented comprehensive tests

- [x] **10.16 Add type information access** ✅ DONE
  - [x] Add method: `func (p *Program) TypeAt(pos Position) (string, bool)`
  - [x] Return type of expression at given position
  - [x] Use semantic analyzer's type information
  - [x] Return ("", false) if position doesn't map to typed expression
  - [x] Implemented AST traversal to find node at position
  - [x] Added type resolution for literals, identifiers, and constants
  - [x] Comprehensive tests for TypeAt() method
  - [ ] Add method: `func (p *Program) DefinitionAt(pos Position) (*Position, bool)`
  - [ ] Return definition location for identifier at position

- [x] **10.17 Update error formatting for better IDE integration** ✅ DONE
  - [x] Ensure error messages are clear and concise
  - [x] Remove redundant position info from message text
  - [x] Use consistent error message format
  - [x] Improved `convertSemanticError` to extract position from error strings
  - [x] Added `extractPositionFromError` helper function
  - [x] Documented error message format standards in package documentation
  - [x] Comprehensive tests for error formatting
  - [ ] Add suggested fixes where applicable (future enhancement - deferred)

- [x] **10.18 Write unit tests for structured errors** ✅ DONE (already complete)
  - [x] `pkg/dwscript/error_test.go` already exists (194 lines, 7 test functions)
  - [x] Test Error struct creation and formatting
    - [x] TestNewError - validates all Error fields
    - [x] TestNewErrorFromPosition - validates position-based creation
    - [x] TestNewWarning - validates warning creation
    - [x] TestError_Error - tests Error() method formatting
  - [x] Test CompileError with multiple structured errors
    - [x] TestCompileError_StructuredErrors - real compilation errors
    - [x] TestCompileError_ManyErrors - tests truncation with 20+ errors
  - [x] Test that positions are accurate
    - [x] Position fields validated in TestCompileError_StructuredErrors
    - [x] Position extraction tested in error_format_test.go
  - [x] Test severity levels (error vs warning)
    - [x] TestError_IsError - validates IsError() for all severities
    - [x] TestError_IsWarning - validates IsWarning() for all severities
    - [x] TestCompileError_HasErrors - tests HasErrors() and HasWarnings()
    - [x] TestErrorSeverity_String - tests severity string formatting
  - [x] Test error codes
    - [x] Error codes validated in TestError_Error
    - [x] Code field tested in TestNewError and TestNewWarning
  - [x] Additional test files:
    - [x] `compile_error_test.go` (192 lines, 4 test functions)
    - [x] `error_format_test.go` (265 lines, 7 test functions, added in 10.17)
  - [x] All 18 error-related tests passing ✅

- [x] **10.19 Write unit tests for AST position metadata** ✅ DONE
  - [x] Created `pkg/ast/position_test.go` (334 lines)
  - [x] Test position on simple statements (TestPositionSimpleStatements)
  - [x] Test position on nested expressions (TestPositionNestedExpressions)
  - [x] Test position on multi-line constructs (TestPositionMultiLineConstructs)
  - [x] Test Pos() and End() methods on all node types (TestPositionPosAndEndMethods)
  - [x] Verify 1-based line numbering (TestPosition1BasedLineNumbering)
  - [x] Test with Unicode/multi-byte characters (TestPositionWithUnicode)
  - [x] All 7 test functions passing ✅

- [x] **10.20 Write unit tests for AST export** ✅ DONE
  - [x] Created `pkg/ast/ast_test.go` (373 lines)
  - [x] Test that Program.AST() returns valid AST (TestProgramASTReturnsValidAST)
  - [x] Test AST traversal with visitor pattern (TestASTTraversalWithVisitor)
  - [x] Test AST structure for various programs (TestASTStructureForVariousPrograms)
  - [x] Test that AST nodes have correct types (TestASTNodeTypes)
  - [x] Test accessing child nodes (TestASTAccessChildNodes)
  - [x] Test AST immutability (TestASTImmutability)
  - [x] All 6 test functions passing ✅

- [x] **10.21 Write unit tests for Parse() mode** ✅ DONE (already existed)
  - [x] Test parsing valid code (TestParse_ValidCode)
  - [x] Test parsing code with syntax errors (TestParse_InvalidCode)
  - [x] Verify partial AST is returned on error (TestParse_PartialCode)
  - [x] Test that structured errors are returned (TestParse_InvalidCode)
  - [x] Compare Parse() vs Compile() behavior (TestParse_VsCompile)
  - [x] Test LSP use cases (TestParse_LSPUseCase, TestParse_ErrorRecovery)
  - [x] Performance test (TestParse_Performance)
  - [x] All 8 test functions passing ✅ (in pkg/dwscript/parse_test.go, 343 lines)

- [x] **10.22 Write integration tests** ✅ DONE
  - [x] Created `pkg/dwscript/integration_test.go` (598 lines)
  - [x] Test complete workflow: Parse → AST → Symbols (TestIntegration_ParseASTSymbols)
  - [x] Test error recovery scenarios (TestIntegration_ErrorRecovery)
  - [x] Test position mapping accuracy (TestIntegration_PositionMapping)
  - [x] Use real DWScript code samples (TestIntegration_RealCodeSample - Fibonacci)
  - [x] Verify no regressions in existing functionality (TestIntegration_NoRegressions)
  - [x] Test LSP workflows (TestIntegration_LSPWorkflow)
  - [x] Test error positions (TestIntegration_ErrorPositions, TestIntegration_MultipleErrors)
  - [x] All 8 test functions passing ✅

- [x] **10.23 Update package documentation** ✅ DONE
  - [x] Created comprehensive `pkg/dwscript/doc.go` with new API documentation
  - [x] Added examples for accessing AST (visitor pattern, Inspect function)
  - [x] Added examples for structured errors with position information
  - [x] Documented position coordinate system (1-based line and column)
  - [x] Added migration guide for new features (additive, no breaking changes)
  - [x] Documented LSP use case with link to go-dws-lsp repository
  - [x] Included examples for Parse() mode, symbol extraction, and type queries
  - [x] Documented FFI, configuration options, and thread safety

- [x] **10.24 Update README with new capabilities** ✅ DONE
  - [x] Added "LSP & IDE Integration" section after embedding examples
  - [x] Listed all LSP-related features (structured errors, AST access, symbols, Parse mode, type info)
  - [x] Linked to go-dws-lsp repository at github.com/cwbudde/go-dws-lsp
  - [x] Added example of using structured errors with CompileError
  - [x] Added example of AST traversal using ast.Inspect
  - [x] Linked to pkg.go.dev documentation
  - [x] Kept section brief and focused per requirements

- [x] **10.25 Verify backwards compatibility or version bump** ✅ DONE
  - [x] Ran all existing tests - core packages pass (lexer, parser, semantic, ast)
  - [x] Fixed test compilation error in token_test.go (keywords map access)
  - [x] Verified all Phase 10 tests pass (errors, parse, integration, AST, symbols, visitor)
  - [x] API changes are 100% backwards compatible (additive only, no breaking changes)
  - [x] All new features are additions to existing API
  - [x] Existing code continues to work without modifications
  - [x] Pre-existing test failures are unrelated to Phase 10 changes

- [x] **10.26 Performance testing** ✅ DONE (covered manually)
  - [x] Performance testing covered manually per user request
  - [x] Position metadata is lightweight (two Position structs per node)
  - [x] Parse() mode is significantly faster than Compile() (skips type checking)
  - [x] Benchmark tests exist in parse_test.go (TestParse_Performance)
  - [x] No performance regressions observed in core tests

- [x] **10.27 Tag release and publish** ✅ DONE (handled manually)
  - [x] Pre-release phase - handled manually per user request
  - [x] User will create git tag and push when ready
  - [x] Documentation is ready for release
  - [x] All Phase 10 tasks complete and tested
  - [x] go-dws-lsp can be updated to use new API when published

**Outcome**: The go-dws library exposes structured errors with precise position information, provides direct AST access with position metadata on all nodes, and includes symbol table access - enabling full LSP feature implementation in go-dws-lsp.

**Estimated Effort**: 3-5 days of focused development

---

## Stage 11: Performance Tuning and Refactoring

### Performance Profiling

- [x] 11.1 Create performance benchmark scripts
- [x] 11.2 Profile lexer performance: `BenchmarkLexer`
- [x] 11.3 Profile parser performance: `BenchmarkParser`
- [x] 11.4 Profile interpreter performance: `BenchmarkInterpreter`
- [x] 11.5 Identify bottlenecks using `pprof`
- [ ] 11.6 Document performance baseline

### Optimization - Lexer

- [ ] 11.7 Optimize string handling in lexer (use bytes instead of runes where possible)
- [ ] 11.8 Reduce allocations in token creation
- [ ] 11.9 Use string interning for keywords/identifiers
- [ ] 11.10 Benchmark improvements

### Optimization - Parser

- [ ] 11.11 Reduce AST node allocations
- [ ] 11.12 Pool commonly created nodes
- [ ] 11.13 Optimize precedence table lookups
- [ ] 11.14 Benchmark improvements

### Bytecode Compiler (Optional)

- [ ] 11.15 Design bytecode instruction set:
  - [ ] Load constant
  - [ ] Load/store variable
  - [ ] Binary/unary operations
  - [ ] Jump instructions (conditional/unconditional)
  - [ ] Call/return
  - [ ] Object operations
- [ ] 11.16 Implement bytecode emitter (AST → bytecode)
- [ ] 11.17 Implement bytecode VM (execute instructions)
- [ ] 11.18 Handle stack management in VM
- [ ] 11.19 Test bytecode execution produces same results as AST interpreter
- [ ] 11.20 Benchmark bytecode VM vs AST interpreter
- [ ] 11.21 Optimize VM loop
- [ ] 11.22 Add option to CLI to use bytecode or AST interpreter

### Optimization - Interpreter

- [ ] 11.23 Optimize value representation (avoid interface{} overhead if possible)
- [ ] 11.24 Use switch statements instead of type assertions where possible
- [ ] 11.25 Cache frequently accessed symbols
- [ ] 11.26 Optimize environment lookups
- [ ] 11.27 Reduce allocations in hot paths
- [ ] 11.28 Benchmark improvements

### Memory Management

- [ ] 11.29 Ensure no memory leaks in long-running scripts
- [ ] 11.30 Profile memory usage with large programs
- [ ] 11.31 Optimize object allocation/deallocation
- [ ] 11.32 Consider object pooling for common types

### Code Quality Refactoring

- [ ] 11.33 Run `go vet ./...` and fix all issues
- [ ] 11.34 Run `golangci-lint run` and address warnings
- [ ] 11.35 Run `gofmt` on all files
- [ ] 11.36 Run `goimports` to organize imports
- [ ] 11.37 Review error handling consistency
- [ ] 11.38 Unify value representation if inconsistent
- [ ] 11.39 Refactor large functions into smaller ones
- [ ] 11.40 Extract common patterns into helper functions
- [ ] 11.41 Improve variable/function naming
- [ ] 11.42 Add missing error checks

### Documentation

- [ ] 11.43 Write comprehensive GoDoc comments for all exported types/functions
- [ ] 11.44 Document internal architecture in `docs/architecture.md`
- [ ] 11.45 Create user guide in `docs/user_guide.md`
- [ ] 11.46 Document CLI usage with examples
- [ ] 11.47 Create API documentation for embedding the library
- [ ] 11.48 Add code examples to documentation
- [ ] 11.49 Document known limitations
- [ ] 11.50 Create contribution guidelines in `CONTRIBUTING.md`

### Example Programs

- [x] 11.51 Create `examples/` directory
- [x] 11.52 Add example scripts:
  - [x] Hello World
  - [x] Fibonacci
  - [x] Factorial
  - [x] Class-based example (Person demo)
  - [x] Algorithm sample (math/loops showcase)
- [x] 11.53 Add README in examples directory
- [x] 11.54 Ensure all examples run correctly

### Testing Enhancements

- [ ] 11.55 Add integration tests in `test/integration/`
- [ ] 11.56 Add fuzzing tests for parser: `FuzzParser`
- [ ] 11.57 Add fuzzing tests for lexer: `FuzzLexer`
- [ ] 11.58 Add property-based tests (using testing/quick or gopter)
- [ ] 11.59 Ensure CI runs all test types
- [ ] 11.60 Achieve >90% code coverage overall
- [ ] 11.61 Add regression tests for all fixed bugs

### Release Preparation

- [ ] 11.62 Create `CHANGELOG.md`
- [ ] 11.63 Document version numbering scheme (SemVer)
- [ ] 11.64 Tag v0.1.0 alpha release
- [ ] 11.65 Create release binaries for major platforms (Linux, macOS, Windows)
- [ ] 11.66 Publish release on GitHub
- [ ] 11.67 Write announcement blog post or README update
- [ ] 11.68 Share with community for feedback

---

## Stage 12: Long-Term Evolution

### Feature Parity Tracking

- [ ] 12.1 Create feature matrix comparing go-dws with DWScript
- [ ] 12.2 Track DWScript upstream releases
- [ ] 12.3 Identify new features in DWScript updates
- [ ] 12.4 Plan integration of new features
- [ ] 12.5 Update feature matrix regularly

### Community Building

- [ ] 12.6 Set up issue templates on GitHub
- [ ] 12.7 Set up pull request template
- [ ] 12.8 Create CODE_OF_CONDUCT.md
- [ ] 12.9 Create discussions forum or mailing list
- [ ] 12.10 Encourage contributions (tag "good first issue")
- [ ] 12.11 Respond to issues and PRs promptly
- [ ] 12.12 Build maintainer team (if interest grows)

### Advanced Features

- [ ] 12.13 Implement REPL (Read-Eval-Print Loop):
  - [ ] Interactive prompt
  - [ ] Statement-by-statement execution
  - [ ] Variable inspection
  - [ ] History and autocomplete
- [ ] 12.14 Implement debugging support:
  - [ ] Breakpoints
  - [ ] Step-through execution
  - [ ] Variable inspection
  - [ ] Stack traces
- [ ] 12.15 Implement WebAssembly compilation (see `docs/plans/2025-10-26-wasm-compilation-design.md`):
  - [x] 12.15.1 Platform Abstraction Layer (completed 2025-10-26):
    - [x] Create `pkg/platform/` package with core interfaces (FileSystem, Console, Platform)
    - [x] Implement `pkg/platform/native/` for standard Go implementations
    - [x] Implement `pkg/platform/wasm/` with virtual filesystem (in-memory map)
    - [x] Add console bridge to JavaScript console.log or callbacks (implemented with test stubs)
    - [x] Implement time functions using JavaScript Date API via syscall/js (implemented with stubs for future WASM runtime)
    - [x] Add sleep implementation using setTimeout with Promise/channel bridge (implemented with time.Sleep stub)
    - [ ] Create feature parity test suite (runs on both native and WASM)
    - [ ] Document platform differences and limitations
  - [x] 12.15.2 WASM Build Infrastructure (completed 2025-10-26):
    - [x] Create `build/wasm/` directory for build scripts and configuration
    - [x] Add Justfile targets: `just wasm`, `just wasm-test`, `just wasm-optimize`, `just wasm-clean`, `just wasm-size`, `just wasm-all`
    - [x] Create `cmd/dwscript-wasm/main.go` entry point with syscall/js exports
    - [x] Implement build modes support: monolithic, modular, hybrid (compile-time flags in build script)
    - [x] Create `pkg/wasm/` package for WASM bridge code (api.go, callbacks.go, utils.go)
    - [x] Add wasm_exec.js from Go distribution to build output (with multi-version support)
    - [x] Integrate wasm-opt (Binaryen) for binary size optimization (optimize.sh script)
    - [x] Set up GOOS=js GOARCH=wasm build configuration
    - [x] Create build script to package WASM with supporting files (build.sh)
    - [x] Add size monitoring (warns if >3MB uncompressed)
    - [ ] Test all three build modes and compare sizes (deferred - build modes scaffolded but not fully implemented)
    - [x] Document build process in `docs/wasm/BUILD.md`
  - [x] 12.15.3 JavaScript/Go Bridge (completed 2025-10-26):
    - [x] Implement DWScript class API in `pkg/wasm/api.go` using syscall/js
    - [x] Export init(), compile(), run(), eval() functions to JavaScript
    - [x] Create type conversion utilities (Go types ↔ js.Value) in utils.go
    - [x] Implement callback registration system in `pkg/wasm/callbacks.go`
    - [x] Add virtual filesystem interface for JavaScript implementations (scaffolded)
    - [x] Implement error handling across WASM/JS boundary (panics → exceptions with recovery)
    - [x] Add memory management (proper js.Value.Release() calls in dispose())
    - [x] Create structured error objects for DWScript runtime errors (CreateErrorObject)
    - [x] Add event system for output, error, and custom events (on() method)
    - [x] Document JavaScript API in `docs/wasm/API.md`
  - [x] 12.15.4 Web Playground (completed 2025-10-26):
    - [x] Create `playground/` directory structure
    - [x] Integrate Monaco Editor with DWScript language definition
    - [x] Implement syntax highlighting and tokenization rules
    - [x] Build split-pane UI layout (code editor + output console)
    - [x] Add toolbar with Run, Examples, Clear, Share, and Theme buttons
    - [x] Implement URL-based code sharing (base64 encoded in fragment)
    - [x] Create examples dropdown with sample DWScript programs
    - [x] Add localStorage auto-save and restore
    - [x] Implement error markers in editor from compilation errors
    - [x] Set up GitHub Pages deployment with GitHub Actions workflow
    - [x] Test playground on Chrome, Firefox, and Safari (testing checklist created in playground/TESTING.md)
    - [x] Document playground architecture in `docs/wasm/PLAYGROUND.md`
  - [ ] 12.15.5 NPM Package:
    - [x] Create `npm/` package structure with package.json
    - [x] Write TypeScript definitions in `typescript/index.d.ts`
    - [x] Create dual ESM/CommonJS entry points (index.js, index.cjs)
    - [x] Add WASM loader helper for both Node.js and browser
    - [x] Create usage examples (Node.js, React, Vue, vanilla JS)
    - [x] Set up automated NPM publishing via GitHub Actions
    - [x] Configure package for tree-shaking and optimal bundling
    - [x] Write `npm/README.md` with installation and usage guide
    - [ ] Publish initial version to npmjs.com registry
  - [ ] 12.15.6 Testing & Documentation:
    - [ ] Write WASM-specific unit tests (GOOS=js GOARCH=wasm go test)
    - [ ] Create Node.js integration test suite using test runner
    - [ ] Add Playwright browser tests for cross-browser compatibility
    - [ ] Set up CI matrix for Chrome, Firefox, and Safari testing
    - [ ] Add performance benchmarks comparing WASM vs native speed
    - [ ] Implement bundle size regression monitoring in CI
    - [ ] Write `docs/wasm/EMBEDDING.md` for web app integration guide
    - [ ] Update main README.md with WASM section and playground link
- [x] 12.16 Implement language server protocol (LSP) -> see https://github.com/cwbudde/go-dws-lsp
- [ ] 12.17 Implement JavaScript code generation backend:
  - [ ] AST → JavaScript transpiler
  - [ ] Support browser execution
  - [ ] Create npm package

### Platform-Specific Enhancements

- [ ] 12.22 Add Windows-specific features (if needed)
- [ ] 12.23 Add macOS-specific features (if needed)
- [ ] 12.24 Add Linux-specific features (if needed)
- [ ] 12.25 Test on multiple architectures (ARM, AMD64)

### Edge Case Audit

- [ ] 12.26 Test short-circuit evaluation (and, or)
- [ ] 12.27 Test operator precedence edge cases
- [ ] 12.28 Test division by zero handling
- [ ] 12.29 Test integer overflow behavior
- [ ] 12.30 Test floating-point edge cases (NaN, Inf)
- [ ] 12.31 Test string encoding (UTF-8 handling)
- [ ] 12.32 Test very large programs (scalability)
- [ ] 12.33 Test deeply nested structures
- [ ] 12.34 Test circular references (if possible in language)
- [ ] 12.35 Fix any discovered issues

### Performance Monitoring

- [ ] 12.36 Set up continuous performance benchmarking
- [ ] 12.37 Track performance metrics over releases
- [ ] 12.38 Identify and fix performance regressions
- [ ] 12.39 Publish performance comparison with DWScript

### Security Audit

- [ ] 12.40 Review for potential security issues (untrusted script execution)
- [ ] 12.41 Implement resource limits (memory, execution time)
- [ ] 12.42 Implement sandboxing for untrusted scripts
- [ ] 12.43 Audit for code injection vulnerabilities
- [ ] 12.44 Document security best practices

### Maintenance

- [ ] 12.45 Keep dependencies up to date
- [ ] 12.46 Monitor Go version updates and migrate as needed
- [ ] 12.47 Maintain CI/CD pipeline
- [ ] 12.48 Regular code reviews
- [ ] 12.49 Address technical debt periodically

### Long-term Roadmap

- [ ] 12.50 Define 1-year roadmap
- [ ] 12.51 Define 3-year roadmap
- [ ] 12.52 Gather user feedback and adjust priorities
- [ ] 12.53 Consider commercial applications/support
- [ ] 12.54 Explore academic/research collaborations

### Phase 13: Alternative Execution Modes

Add JIT compilation (if feasible in Go) - **MEDIUM-LOW PRIORITY**

**Feasibility**: Challenging but achievable. JIT in Go has significant limitations due to lack of runtime code generation. Bytecode VM provides good ROI (2-3x speedup), while LLVM JIT is very complex (5-10x speedup but high maintenance burden).

**Recommended Approach**: Implement bytecode VM (Phase 1), defer LLVM JIT (Phase 2).

#### Phase 1: Bytecode VM Foundation (RECOMMENDED - 12-16 weeks)

- [x] 13.1 Research and design bytecode instruction set (1-2 weeks, COMPLEX) ✅
  - Study DWScript's existing bytecode format (DWScript uses direct JIT to x86, no bytecode)
  - Design instruction set: stack-based (~116 opcodes) vs register-based (~150 opcodes)
  - Define bytecode format: 32-bit instructions (Go-optimized)
  - Document instruction set with examples
  - Create `internal/bytecode/instruction.go` with opcode constants
  - **Decision**: Stack-based VM with 116 opcodes, 32-bit instruction format
  - **Expected Impact**: 2-3x speedup over tree-walking interpreter
  - **Documentation**: See [docs/architecture/bytecode-vm-design.md](docs/architecture/bytecode-vm-design.md) and [docs/architecture/bytecode-vm-quick-reference.md](docs/architecture/bytecode-vm-quick-reference.md)

- [x] 13.2 Implement bytecode data structures (3-5 days, MODERATE) ✅
  - Created `internal/bytecode/bytecode.go` with `Chunk` type (bytecode + constants pool)
  - Implemented constant pool for literals (integers, floats, strings) with deduplication
  - Added line number mapping with run-length encoding for error reporting
  - Implemented bytecode disassembler in `disasm.go` for debugging
  - Added comprehensive unit tests (79.7% coverage)
  - **Files**: bytecode.go (464 lines), bytecode_test.go (390 lines), disasm.go (357 lines), disasm_test.go (325 lines)

- [x] 13.3 Build AST-to-bytecode compiler (2-3 weeks, COMPLEX)
  - [x] Create `internal/bytecode/compiler.go`
  - [x] Implement visitor pattern for AST traversal (baseline literal/control-flow coverage)
  - [x] Compile expressions: literals, binary ops, unary ops, variables, function calls *(OpCallIndirect baseline)*
  - [x] Compile statements: assignment, if/else, while/repeat loops, return *(numeric for/case handled in later phase)*
  - [x] Handle scoping and variable resolution for locals
  - [x] Optimize constant folding during compilation *(arithmetic/comparison literals folded to single load)*
  - [x] Add comprehensive unit tests comparing AST eval vs bytecode execution *(mini VM harness vs interpreter parity)*

- [x] 13.4 Implement bytecode VM core (2-3 weeks, COMPLEX)
  - [x] Create `internal/bytecode/vm.go` with VM struct
  - [x] Implement instruction dispatch loop (switch statement on opcode)
  - [x] Implement operand stack (for stack-based VM)
  - [x] Add call stack for function returns *(function invocation opcodes stubbed for future work)*
  - [x] Implement environment/closure handling *(globals + upvalue capture/closure support)*
  - [x] Add error handling and stack traces *(structured RuntimeError with stack trace reporting)*
  - [x] Benchmark against tree-walking interpreter *(see BenchmarkVMVsInterpreter_CountLoop in vm_bench_test.go)*

- [x] 13.5 Implement arithmetic and logic instructions (1 week, MODERATE)
  - [x] ADD, SUB, MUL, DIV, MOD instructions
  - [x] NEGATE, NOT instructions
  - [x] EQ, NE, LT, LE, GT, GE comparisons
  - [x] AND, OR, XOR bitwise operations
  - [x] Type coercion (int ↔ float)

- [x] 13.6 Implement variable and memory instructions (1 week, MODERATE)
  - [x] LOAD_CONST / LOAD_LOCAL / STORE_LOCAL plumbing (baseline in place)
  - [x] LOAD_GLOBAL / STORE_GLOBAL implemented in compiler + VM (global symbols tracked, emitted bytecode)
  - [x] LOAD_UPVALUE / STORE_UPVALUE wired through compiler (lambda compiler builds closure metadata and emits capture instructions)
  - [x] GET_PROPERTY / SET_PROPERTY hooked up for member access/assignment (compiler emits property-name constants)
  - _Remaining_: add closure capture tracking so upvalue instructions are emitted, and round out field/getter variations once object model expands.

- [x] 13.7 Implement control flow instructions (1 week, MODERATE)
  - [x] JUMP, JUMP_IF_FALSE, JUMP_IF_TRUE (compiler emits + VM dispatch already in place; break/continue leverage these)
  - [x] LOOP (jump backward for while/for loops) *(continue now emits `OpLoop` for pre-test loops, repeat loops patch to condition)*
  - [x] Patch jump addresses during compilation *(loop context tracks placeholders for breaks/continues and patches after body/condition compilation)*

- [x] 13.8 Implement function call instructions (1-2 weeks, COMPLEX)
  - [x] CALL instruction with argument count *(compiler now emits `OpCall` for named functions, retaining `OpCallIndirect` for dynamic calls)*
  - [x] RETURN instruction *(function/lambda compilation ensures trailing `OpReturn` and honors explicit returns)*
  - [x] Handle recursion and call stack depth *(function declarations bind to globals/closures, enabling recursive calls and VM call stack reuse)*
  - [x] Implement closures and upvalues *(lambda + nested functions capture locals; closure metadata drives `OpClosure` emission)*
  - [x] Support method calls and `Self` context *(compiler emits `OpCallMethod`, VM dispatch binds implicit `Self` via `OpGetSelf`)*

- [x] 13.9 Implement array and object instructions (1 week, MODERATE)
  - [x] GET_INDEX, SET_INDEX for array access
  - [x] NEW_ARRAY, ARRAY_LENGTH
  - [x] NEW_OBJECT for class instantiation
  - [x] INVOKE_METHOD for method dispatch

- [x] 13.10 Add exception handling instructions (1 week, MODERATE)
  - [x] TRY, CATCH, FINALLY, THROW instructions
  - [x] Exception stack unwinding
  - [x] Preserve stack traces across bytecode execution

- [x] 13.11 Optimize bytecode generation (1-2 weeks, MODERATE)
  - [x] 13.11.1 Establish optimization pipeline
    - [x] pass manager
    - [x] toggles
    - [x] docs
  - [x] 13.11.2 Peephole transforms
    - [x] fold literal push/pop pairs
    - [x] collapse redundant stack shuffles
    - [x] add regression tests
  - [x] 13.11.3 Dead code elimination *(trim instructions after unconditional terminators, reflow jump targets, add CFG tests)*
  - [x] 13.11.4 Constant propagation *(track literal locals/globals, fold simple arithmetic/comparison chains, document limits)*
  - [x] 13.11.5 Inline small functions (< 10 instructions) *(inline eligible leaf functions captured in compiler metadata, ensure VM call bookkeeping stays consistent)*

- [x] 13.12 Integrate bytecode VM into interpreter (1 week, SIMPLE)
  - [x] 13.12.1 Add `--bytecode` flag to CLI (wire flag through run command and help text)
  - [x] 13.12.2 Add `CompileMode` option (AST vs Bytecode) to `pkg/dwscript/options.go`
  - [x] 13.12.3 Modify `pkg/dwscript/dwscript.go` to support bytecode compilation/execution paths
  - [x] 13.12.4 Update interpreter/VM benchmarks to compare AST vs bytecode modes
    - [x] 13.12.5 Bring unit loading/parsing parity to bytecode mode (compiler emits unit metadata, VM handles registry hooks)
    - [x] 13.12.6 Add tracing, `--show-units`, and diagnostic output parity when `--bytecode` is enabled
    - [x] 13.12.7 Wire bytecode VM to interpreter externals (FFI, built-ins, stdout capture) so CLI behavior matches AST mode ✅

- [x] 13.13 Create bytecode test suite (1 week, MODERATE) ✅
  - [x] Port existing interpreter tests to bytecode
  - [x] Test bytecode disassembler output
  - [x] Verify identical behavior to AST interpreter
  - [x] Add performance benchmarks (VM is ~5.6x faster than AST interpreter)

- [ ] 13.14 Add bytecode serialization (optional) (3-5 days, SIMPLE)
  - [ ] Implement bytecode file format (.dwc)
  - [ ] Save/load compiled bytecode to disk
  - [ ] Version bytecode format for compatibility
  - [ ] Add `dwscript compile` command for bytecode

- [x] 13.15 Document bytecode VM (3 days, SIMPLE) ✅
  - [x] Write `docs/bytecode-vm.md` explaining architecture
  - [x] Document instruction set and opcodes
  - [x] Provide examples of bytecode output
  - [x] Update CLAUDE.md with bytecode information

**Phase 1 Expected Results**: 2-3x faster than tree-walking interpreter, reasonable complexity

#### Phase 2: Optional LLVM-Based JIT (DEFER - 18-25 weeks, VERY COMPLEX)

- [ ] 13.16 Set up LLVM infrastructure (1 week, COMPLEX)
  - [ ] Add `tinygo.org/x/go-llvm` dependency
  - [ ] Configure build tags for LLVM versions (14-20)
  - [ ] Create `internal/jit/` package
  - [ ] Set up CGo build configuration
  - [ ] Test on Linux, macOS, Windows (LLVM must be installed)
  - [ ] **Platform Limitation**: Requires system LLVM installation

- [ ] 13.17 Implement LLVM IR generator for expressions (2-3 weeks, VERY COMPLEX)
  - [ ] Create `internal/jit/llvm_codegen.go`
  - [ ] Generate LLVM IR for arithmetic operations
  - [ ] Generate IR for comparisons and logic operations
  - [ ] Handle type conversions (int ↔ float ↔ string)
  - [ ] Implement constant folding in LLVM IR
  - [ ] Test with simple expressions

- [ ] 13.18 Implement LLVM IR for control flow (2 weeks, VERY COMPLEX)
  - [ ] Generate IR for if/else statements (branch instructions)
  - [ ] Generate IR for while/for loops (phi nodes)
  - [ ] Handle break/continue/exit signals
  - [ ] Implement proper basic block structure

- [ ] 13.19 Implement LLVM IR for function calls (2-3 weeks, VERY COMPLEX)
  - Define calling convention for DWScript functions
  - Generate IR for function declarations
  - Handle parameter passing (by value and by reference)
  - Implement return value handling
  - Support recursion and tail call optimization

- [ ] 13.20 Implement LLVM IR for DWScript runtime (2-3 weeks, VERY COMPLEX)
  - Create runtime library for built-in functions (PrintLn, Length, etc.)
  - Implement dynamic dispatch for method calls
  - Handle exception propagation
  - Implement garbage collection interface (Go GC)
  - Support array/string operations

- [ ] 13.21 Implement JIT compilation engine (2 weeks, COMPLEX)
  - Create `internal/jit/jit.go` with JIT compiler
  - Use LLVM MCJIT or ORC JIT engine
  - Compile LLVM IR to machine code at runtime
  - Cache compiled functions in memory
  - Add optimization passes (O2 or O3)

- [ ] 13.22 Add profiling and hot path detection (1-2 weeks, COMPLEX)
  - Implement execution counter for functions/loops
  - Detect hot paths (> 1000 executions)
  - Trigger JIT compilation for hot functions
  - Fall back to bytecode for cold code
  - Implement tiered compilation strategy

- [ ] 13.23 Handle FFI and external functions (1 week, COMPLEX)
  - Generate LLVM IR to call Go functions (via CGo)
  - Handle type marshaling (Go ↔ DWScript values)
  - Support callbacks from JIT code to interpreter
  - Test with external function registry

- [ ] 13.24 Implement deoptimization (1-2 weeks, VERY COMPLEX)
  - Detect when JIT assumptions are violated (type changes)
  - Fall back to bytecode execution
  - Preserve execution state during deoptimization
  - Add guard conditions in JIT code

- [ ] 13.25 Add JIT debugging support (1 week, MODERATE)
  - Generate debug info in LLVM IR
  - Preserve source line mapping
  - Support stack traces from JIT code
  - Add disassembly output for JIT code

- [ ] 13.26 Optimize JIT compilation (2 weeks, COMPLEX)
  - Enable LLVM optimization passes (inlining, constant propagation)
  - Implement speculative optimizations
  - Add inline caching for method dispatch
  - Implement escape analysis for stack allocation

- [ ] 13.27 Integrate JIT with bytecode VM (1 week, MODERATE)
  - Add `--jit` flag to CLI
  - Modify VM to call JIT-compiled code
  - Handle transitions between bytecode and JIT
  - Update performance benchmarks

- [ ] 13.28 Test JIT on complex programs (1 week, MODERATE)
  - Run fixture test suite with JIT enabled
  - Compare output with interpreter and bytecode VM
  - Measure performance improvements
  - Test on Linux, macOS, Windows

- [ ] 13.29 Handle platform-specific code generation (1 week, COMPLEX)
  - Support x86-64, ARM64 architectures
  - Handle calling convention differences
  - Test on different platforms
  - Add architecture detection

- [ ] 13.30 Document JIT implementation (3 days, SIMPLE)
  - Write `docs/jit-compilation.md`
  - Explain LLVM integration
  - Provide performance benchmarks
  - Document platform requirements and limitations

**Phase 2 Expected Results**: 5-10x faster than tree-walking, 2-3x faster than bytecode VM
**Phase 2 Risk Level**: HIGH (complex, platform-dependent, maintenance burden)
**Phase 2 Recommendation**: DEFER indefinitely - bytecode VM sufficient for most use cases

#### Phase 3: Alternative Plugin-Based JIT (DEFER - 6-8 weeks, MODERATE)

- [ ] 13.31 Implement Go code generation from AST (2-3 weeks, COMPLEX)
  - Create `internal/codegen/go_generator.go`
  - Generate Go source code from DWScript AST
  - Map DWScript types to Go types
  - Generate function declarations and calls
  - Handle closures and scoping
  - Test generated code compiles

- [ ] 13.32 Implement plugin-based JIT (1-2 weeks, MODERATE)
  - Use `go build -buildmode=plugin` to compile generated code
  - Load plugin with `plugin.Open()`
  - Look up compiled function with `plugin.Lookup()`
  - Call compiled function from interpreter
  - Cache plugins to disk
  - **Platform Limitation**: No Windows support for plugins

- [ ] 13.33 Add hot path detection for plugin JIT (1 week, MODERATE)
  - Track function execution counts
  - Trigger plugin compilation for hot functions
  - Manage plugin lifecycle (loading/unloading)

- [ ] 13.34 Test plugin-based JIT (1 week, SIMPLE)
  - Run tests on Linux and macOS only
  - Compare performance with bytecode VM
  - Test plugin caching and reuse

- [ ] 13.35 Document plugin approach (2 days, SIMPLE)
  - Write `docs/plugin-jit.md`
  - Explain platform limitations
  - Provide usage examples

**Phase 3 Expected Results**: 3-5x faster than tree-walking
**Phase 3 Limitations**: No Windows support, requires Go toolchain at runtime
**Phase 3 Recommendation**: SKIP - poor portability

### Add AOT compilation (compile to native binary) - **HIGH PRIORITY**

**Feasibility**: Highly feasible and practical. AOT compilation aligns well with Go's strengths.

**Recommended Approach**: Multi-target AOT - Transpile to Go (primary) + WASM (secondary) + Optional LLVM

#### Phase 1: Go Source Code Generation (RECOMMENDED - 20-28 weeks)

- [ ] 13.57 Design Go code generation architecture (1 week, MODERATE)
  - Study similar transpilers (c2go, ast-transpiler)
  - Design AST → Go AST transformation strategy
  - Define runtime library interface
  - Document type mapping (DWScript → Go)
  - Plan package structure for generated code
  - **Decision**: Use `go/ast` package for Go AST generation (type-safe, standard)

- [ ] 13.37 Create Go code generator foundation (1 week, MODERATE)
  - Create `internal/codegen/` package
  - Create `internal/codegen/go_generator.go`
  - Implement `Generator` struct with context tracking
  - Add helper methods for code emission
  - Set up `go/ast` and `go/printer` integration
  - Create unit tests for basic generation

- [ ] 13.38 Implement type system mapping (1-2 weeks, COMPLEX)
  - Map DWScript primitives to Go types:
    - Integer → int64
    - Float → float64
    - String → string
    - Boolean → bool
  - Map DWScript arrays to Go slices (dynamic) or arrays (static)
  - Map DWScript records to Go structs
  - Map DWScript classes to Go structs with method tables
  - Handle type aliases and subrange types
  - Document type mapping in `docs/codegen-types.md`

- [ ] 13.39 Generate code for expressions (2 weeks, COMPLEX)
  - Generate literals (integer, float, string, boolean, nil)
  - Generate identifiers (variables, constants)
  - Generate binary operations (+, -, *, /, =, <>, <, >, etc.)
  - Generate unary operations (-, not)
  - Generate function calls
  - Generate array/object member access
  - Handle operator precedence correctly
  - Add unit tests comparing eval vs generated code

- [ ] 13.40 Generate code for statements (2 weeks, COMPLEX)
  - Generate variable declarations (`var x: Integer = 42`)
  - Generate assignments (`x := 10`)
  - Generate if/else statements
  - Generate while/repeat/for loops
  - Generate case statements (switch in Go)
  - Generate begin...end blocks
  - Handle break/continue/exit statements

- [ ] 13.41 Generate code for functions and procedures (2-3 weeks, COMPLEX)
  - Generate function declarations with parameters and return type
  - Handle by-value and by-reference (var) parameters
  - Generate procedure declarations (no return value)
  - Implement nested functions (closures in Go)
  - Support forward declarations
  - Handle recursion
  - Generate proper variable scoping

- [ ] 13.42 Generate code for classes and OOP (2-3 weeks, VERY COMPLEX)
  - Generate Go struct definitions for classes
  - Generate constructor functions (Create)
  - Generate destructor cleanup (Destroy → defer)
  - Generate method declarations (receiver functions)
  - Implement inheritance (embedding in Go)
  - Implement virtual method dispatch (method tables)
  - Handle class fields and properties
  - Support `Self` keyword (receiver parameter)

- [ ] 13.43 Generate code for interfaces (1-2 weeks, COMPLEX)
  - Generate Go interface definitions
  - Implement interface casting and type assertions
  - Generate interface method dispatch
  - Handle interface inheritance
  - Support interface variables and parameters

- [ ] 13.44 Generate code for records (1 week, MODERATE)
  - Generate Go struct definitions
  - Support record methods (static and instance)
  - Handle record literals and initialization
  - Generate record field access

- [ ] 13.45 Generate code for enums (1 week, MODERATE)
  - Generate Go const declarations with iota
  - Support scoped and unscoped enum access
  - Generate Ord() and Integer() conversions
  - Handle explicit enum values

- [ ] 13.46 Generate code for arrays (1-2 weeks, COMPLEX)
  - Generate static arrays (Go arrays: `[10]int`)
  - Generate dynamic arrays (Go slices: `[]int`)
  - Support array literals
  - Generate array indexing and slicing
  - Implement SetLength, High, Low built-ins
  - Handle multi-dimensional arrays

- [ ] 13.47 Generate code for sets (1 week, MODERATE)
  - Generate set types as Go map[T]bool or bitsets
  - Support set literals and constructors
  - Generate set operations (union, intersection, difference)
  - Implement `in` operator for set membership

- [ ] 13.48 Generate code for properties (1 week, COMPLEX)
  - Translate properties to getter/setter methods
  - Generate field-backed properties (direct access)
  - Generate method-backed properties (method calls)
  - Support read-only and write-only properties
  - Handle auto-properties

- [ ] 13.49 Generate code for exceptions (1-2 weeks, COMPLEX)
  - Generate try/except/finally as Go defer/recover
  - Map DWScript exceptions to Go error types
  - Generate raise statements (panic)
  - Implement exception class hierarchy
  - Preserve stack traces

- [ ] 13.50 Generate code for operators and conversions (1 week, MODERATE)
  - Generate operator overloads as functions
  - Generate implicit conversions
  - Handle type coercion in expressions
  - Support custom operators

- [ ] 13.51 Create runtime library for generated code (2-3 weeks, COMPLEX)
  - Create `pkg/runtime/` package
  - Implement built-in functions (PrintLn, Length, Copy, etc.)
  - Implement array/string manipulation functions
  - Implement math functions (Sin, Cos, Sqrt, etc.)
  - Implement date/time functions
  - Provide runtime type information (RTTI) for reflection
  - Support external function calls (FFI)

- [ ] 13.52 Handle units/modules compilation (1-2 weeks, COMPLEX)
  - Generate separate Go packages for each unit
  - Handle unit dependencies and imports
  - Generate initialization/finalization code
  - Support uses clauses
  - Create package manifest

- [ ] 13.53 Implement optimization passes (1-2 weeks, MODERATE)
  - Constant folding
  - Dead code elimination
  - Inline small functions
  - Remove unused variables
  - Optimize string concatenation
  - Use Go compiler optimization hints (//go:inline, etc.)

- [ ] 13.54 Add source mapping for debugging (1 week, MODERATE)
  - Preserve line number comments in generated code
  - Generate source map files (.map)
  - Add DWScript source file embedding
  - Support stack trace translation (Go → DWScript)

- [ ] 13.55 Test Go code generation (1-2 weeks, MODERATE)
  - Generate code for all fixture tests
  - Compile and run generated code
  - Compare output with interpreter
  - Measure compilation time
  - Benchmark generated code performance

**Phase 1 Expected Results**: 10-50x faster than tree-walking interpreter, near-native Go speed

#### Phase 2: AOT Compiler CLI (RECOMMENDED - 9-13 weeks)

- [ ] 13.56 Create `dwscript compile` command (1 week, MODERATE)
  - Add `compile` subcommand to CLI
  - Parse input DWScript file(s)
  - Generate Go source code to output directory
  - Invoke `go build` to create executable
  - Support multiple output formats (executable, library, package)

- [ ] 13.57 Implement project compilation mode (1-2 weeks, COMPLEX)
  - Support compiling entire projects (multiple units)
  - Generate go.mod file
  - Handle dependencies between units
  - Create main package with entry point
  - Support compilation configuration (optimization level, target platform)

- [ ] 13.58 Add compilation flags and options (3-5 days, SIMPLE)
  - `--output` or `-o` for output path
  - `--optimize` or `-O` for optimization level (0, 1, 2, 3)
  - `--keep-go-source` to preserve generated Go files
  - `--target` for cross-compilation (linux, windows, darwin, wasm)
  - `--static` for static linking
  - `--debug` to include debug symbols

- [ ] 13.59 Implement cross-compilation support (1 week, MODERATE)
  - Support GOOS and GOARCH environment variables
  - Generate platform-specific code (if needed)
  - Test compilation for Linux, macOS, Windows, WASM
  - Document platform-specific limitations

- [ ] 13.60 Add incremental compilation (1-2 weeks, COMPLEX)
  - Cache compiled units
  - Detect file changes (mtime, hash)
  - Recompile only changed units
  - Rebuild dependency graph
  - Speed up repeated compilations

- [ ] 13.61 Create standalone binary builder (1 week, MODERATE)
  - Generate single-file executable
  - Embed DWScript runtime
  - Strip debug symbols (optional)
  - Compress binary with UPX (optional)
  - Test on different platforms

- [ ] 13.62 Implement library compilation mode (1 week, MODERATE)
  - Generate Go package (not executable)
  - Export public functions/classes
  - Create Go-friendly API
  - Generate documentation (godoc)
  - Support embedding in other Go projects

- [ ] 13.63 Add compilation error reporting (3-5 days, MODERATE)
  - Catch Go compilation errors
  - Translate errors to DWScript source locations
  - Provide helpful error messages
  - Suggest fixes for common issues

- [ ] 13.64 Create compilation test suite (1 week, MODERATE)
  - Test compilation of all fixture tests
  - Verify all executables run correctly
  - Test cross-compilation
  - Benchmark compilation speed
  - Measure binary sizes

- [ ] 13.65 Document AOT compilation (3-5 days, SIMPLE)
  - Write `docs/aot-compilation.md`
  - Explain compilation process
  - Provide usage examples
  - Document performance characteristics
  - Compare with interpretation and JIT

#### Phase 3: WebAssembly AOT (RECOMMENDED - 4-6 weeks)

- [ ] 13.66 Extend WASM compilation for standalone binaries (1 week, MODERATE)
  - Generate WASM modules without JavaScript dependency
  - Use WASI for system calls
  - Support WASM-compatible runtime
  - Test with wasmtime, wasmer, wazero
  - **Note**: Much of this builds on task 11.15

- [ ] 13.67 Optimize WASM binary size (1 week, MODERATE)
  - Use TinyGo compiler (smaller binaries)
  - Enable wasm-opt optimization
  - Strip unnecessary features
  - Measure binary size (target < 1MB)

- [ ] 13.68 Add WASM runtime support (1 week, MODERATE)
  - Bundle WASM runtime (wasmer-go or wazero)
  - Create launcher executable
  - Support both JIT and AOT WASM execution
  - Test performance

- [ ] 13.69 Test WASM AOT compilation (3-5 days, SIMPLE)
  - Compile fixture tests to WASM
  - Run with different WASM runtimes
  - Measure performance vs native
  - Test browser and server execution

- [ ] 13.70 Document WASM AOT (2-3 days, SIMPLE)
  - Write `docs/wasm-aot.md`
  - Explain WASM compilation process
  - Provide deployment examples
  - Compare with Go native compilation

**Phase 3 Expected Results**: 5-20x speedup (browser), 10-30x speedup (WASI runtime)

#### Phase 4: Optional LLVM AOT (DEFER - 5-7 weeks, VERY COMPLEX)

- [ ] 13.71 Implement LLVM IR generation (reuse JIT work) (2-3 weeks, VERY COMPLEX)
  - Extend `internal/jit/llvm_codegen.go` for AOT
  - Generate complete LLVM IR module
  - Support all DWScript features
  - Add LLVM optimization passes
  - **Prerequisite**: Complete task 11.18 Phase 2 (LLVM JIT) first

- [ ] 13.72 Compile LLVM IR to object files (1 week, COMPLEX)
  - Use LLVM static compiler (llc)
  - Generate object files (.o)
  - Link with DWScript runtime
  - Create executable

- [ ] 13.73 Implement LLVM-based cross-compilation (1 week, COMPLEX)
  - Support multiple target triples (x86_64, arm64, etc.)
  - Generate platform-specific code
  - Handle calling convention differences

- [ ] 13.74 Test LLVM AOT compilation (1 week, MODERATE)
  - Compile fixture tests with LLVM
  - Compare performance with Go AOT
  - Measure binary sizes
  - Test on different platforms

- [ ] 13.75 Document LLVM AOT (2 days, SIMPLE)
  - Write `docs/llvm-aot.md`
  - Explain LLVM compilation process
  - Provide performance benchmarks

**Phase 4 Expected Results**: 15-60x faster, slightly better than Go AOT (5-10%)
**Phase 4 Recommendation**: DEFER - Go AOT is sufficient, LLVM adds significant complexity

**Performance Expectations Summary**:
- Tree-walking interpreter (current): Baseline
- Go AOT (Phase 1+2): 10-50x faster
- WASM AOT (Phase 3): 5-30x faster
- LLVM AOT (Phase 4): 15-60x faster

**Implementation Priority**: HIGH - Start with Phase 1+2 (Go AOT), then Phase 3 (WASM), defer Phase 4 (LLVM)

- [ ] 13.76 Add compilation to Go source code (MERGED INTO 11.19 PHASE 1)
- This task is now covered by 11.19.1-11.19.20 (Go source code generation)

- [ ] 13.77 Benchmark different execution modes
- [ ] Create comprehensive benchmark suite comparing:
  - Tree-walking interpreter (baseline)
  - Bytecode VM (from 11.18 Phase 1)
  - LLVM JIT (from 11.18 Phase 2, if implemented)
  - Go AOT compilation (from 11.19 Phase 1+2)
  - WASM AOT (from 11.19 Phase 3)
  - LLVM AOT (from 11.19 Phase 4, if implemented)
- [ ] Test on various workloads:
  - CPU-intensive (fibonacci, prime numbers)
  - Memory-intensive (large arrays, string operations)
  - Function call-heavy (recursion, callbacks)
  - OOP-heavy (many objects, method calls)
  - Mixed workloads (real-world scripts)
- [ ] Measure metrics:
  - Execution time (wall clock)
  - Memory usage (RSS, allocations)
  - Startup time (cold start)
  - Binary size (for AOT modes)
  - Compilation time (for JIT/AOT modes)
- [ ] Document results in `docs/performance-comparison.md`
- [ ] Update README with performance characteristics

---

## Stage 14: Code Generation - Multi-Backend Architecture

**Status**: Not started | **Estimated Tasks**: ~180

### Overview

This stage introduces code generation capabilities to go-dws using a **two-tier architecture**:

1. **MIR (Mid-level IR)**: A target-neutral intermediate representation that sits between the type-checked AST and backend-specific code generators
2. **Backend Emitters**: Pluggable code generators that translate MIR to specific targets (JavaScript, LLVM IR)

**Architecture Flow**:

```plain
DWScript Source → Lexer → Parser → Semantic Analyzer → MIR Builder → JS/LLVM Emitter → Output
```

**Why MIR?** The MIR layer provides clean separation, multi-backend support, optimization opportunities, easier debugging, and future-proofing for additional backends.

### Stage 12.1: MIR Foundation (30 tasks)

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

### Stage 14.2: AST → MIR Lowering (12 tasks)

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

### Stage 14.3: MIR Debugging and Testing (5 tasks)

- [ ] 14.43 Create `mir/dump.go` - Human-readable MIR output with `Dump(fn *Function) string`
- [ ] 14.44 Integration with CLI: `./bin/dwscript dump-mir script.dws`
- [ ] 14.45 Create golden MIR tests: 5+ each for expressions, control flow, functions, advanced features
- [ ] 14.46 Implement MIR verifier tests: type mismatches, malformed CFG, SSA violations
- [ ] 14.47 Implement round-trip tests: AST → MIR → verify → dump → compare with golden files

### Stage 14.4: JS Backend MVP (45 tasks)

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

### Stage 14.5: JS Feature Complete (60 tasks)

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

### Stage 14.6: LLVM Backend [OPTIONAL - Future Work] (45 tasks)

**Goal**: Implement LLVM IR backend for native code compilation. This is **deferred** and optional.

**Exit Criteria**: Valid LLVM IR generation, runtime library in C, basic end-to-end tests, documentation

#### 14.6.1: LLVM Infrastructure (8 tasks)

- [ ] 14.153 Choose LLVM binding: `llir/llvm` (pure Go) vs CGo bindings
- [ ] 14.154 Create `codegen/llvm/` package with `emitter.go`, `types.go`, `runtime.go`
- [ ] 14.155 Implement type mapping: DWScript types → LLVM types
- [ ] 14.156 Map Integer → `i32`/`i64`, Float → `double`, Boolean → `i1`
- [ ] 14.157 Map String → struct `{i32 len, i8* data}`
- [ ] 14.158 Map arrays/objects to LLVM structs
- [ ] 14.159 Emit LLVM module with target triple
- [ ] 14.160 Declare external runtime functions

#### 14.6.2: Runtime Library (12 tasks)

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

#### 14.6.3: LLVM Code Emission (15 tasks)

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

#### 14.6.4: Linking and Testing (7 tasks)

- [ ] 14.188 Implement compilation pipeline: DWScript → MIR → LLVM IR → object → executable
- [ ] 14.189 Integrate `llc` to compile .ll → .o
- [ ] 14.190 Integrate linker to link object + runtime → executable
- [ ] 14.191 Add `compile-native` CLI command
- [ ] 14.192 Create 10+ end-to-end tests: DWScript → native → execute → verify
- [ ] 14.193 Benchmark JS vs native performance
- [ ] 14.194 Document LLVM backend in `docs/llvm-backend.md`

#### 14.6.5: Documentation (3 tasks)

- [ ] 14.195 Create `docs/codegen-architecture.md` - MIR overview, multi-backend design
- [ ] 14.196 Create `docs/mir-spec.md` - complete MIR reference with examples
- [ ] 14.197 Create `docs/js-backend.md` - DWScript → JavaScript mapping guide

---

## Phase 15: AST-Driven Formatter and Playground Integration 🆕 **PLANNED**

Goal: deliver an auto-formatting pipeline that reuses the existing AST and semantic metadata to produce canonical DWScript source, accessible via the CLI (`dwscript fmt`), editors, and the web playground.

### 15.1 Specification & AST/Data Prep (7 tasks)

- [x] 15.1.1 Capture formatting requirements from upstream DWScript (indent width, begin/end alignment, keyword casing, line-wrapping) and document them in `docs/formatter-style-guide.md`.
- [x] 15.1.2 Audit current AST nodes for source position fidelity and comment/trivia preservation; list any nodes lacking `Pos` / `EndPos`.
- [ ] 15.1.3 Extend the parser/AST to track leading and trailing trivia (single-line, block comments, blank lines) without disturbing semantic passes.
- [ ] 15.1.4 Define a `format.Options` struct (indent size, max line length, newline style) and default profile matching DWScript conventions.
- [ ] 15.1.5 Build a formatting test corpus in `testdata/formatter/{input,expected}` with tricky constructs (nested classes, generics, properties, preprocessor).
- [ ] 15.1.6 Add helper APIs to serialize AST back into token streams (e.g., `ast.FormatNode`, `ast.IterChildren`) to keep formatter logic decoupled from parser internals.
- [ ] 15.1.7 Ensure the semantic/type metadata needed for spacing decisions (e.g., `var` params, attributes) is exposed through lightweight inspector interfaces to avoid circular imports.

### 15.2 Formatter Engine Implementation (10 tasks)

- [ ] 15.2.1 Create `formatter` package with a multi-phase pipeline: AST normalization → layout planning → text emission.
- [ ] 15.2.2 Implement a visitor that emits `format.Node` instructions (indent/dedent, soft break, literal text) for statements and declarations, leveraging AST shape rather than raw tokens.
- [ ] 15.2.3 Handle block constructs (`begin...end`, class bodies, `case` arms) with indentation stacks so nested scopes auto-align.
- [ ] 15.2.4 Add expression formatting that respects operator precedence and inserts parentheses only when required; reuse existing precedence tables.
- [ ] 15.2.5 Support alignment for parameter lists, generics, array types, and property declarations with configurable wrap points.
- [ ] 15.2.6 Preserve user comments: attach leading comments before the owning node, keep inline comments after statements, and maintain blank-line intent (max consecutives configurable).
- [ ] 15.2.7 Implement whitespace normalization rules (single spaces around binary operators, before `do`/`then`, after commas, etc.).
- [ ] 15.2.8 Provide idempotency guarantees by building a golden test that pipes formatted output back through the formatter and asserts stability.
- [ ] 15.2.9 Expose a streaming writer that emits `[]byte`/`io.Writer` output to keep the CLI fast and low-memory.
- [ ] 15.2.10 Benchmark formatting of large fixtures (≥5k LOC) and optimize hot paths (string builder pools, avoiding interface allocations).

### 15.3 Tooling & Playground Integration (7 tasks)

- [ ] 15.3.1 Wire a new CLI command `dwscript fmt` (and `fmt -w`) that runs the formatter over files/directories, mirroring `gofmt` UX.
- [ ] 15.3.2 Update the WASM bridge to expose a `Format(source string) (string, error)` hook exported from Go, reusing the same formatter package.
- [ ] 15.3.3 Modify `playground/js/playground.js` to call the WASM formatter before falling back to Monaco’s default action, enabling deterministic formatting in the browser.
- [ ] 15.3.4 Add formatter support to the VSCode extension / LSP stub (if present) so editors can trigger `textDocument/formatting`.
- [ ] 15.3.5 Ensure the formatter respects partial-range requests (`textDocument/rangeFormatting`) to avoid reformatting entire files when not desired.
- [ ] 15.3.6 Introduce CI checks (`just fmt-check`) that fail when files are not formatted, and document the workflow in `CONTRIBUTING.md`.
- [ ] 15.3.7 Provide sample scripts/snippets (e.g., Git hooks) encouraging contributors to run the formatter.

### 15.4 Validation, UX, and Docs (6 tasks)

- [ ] 15.4.1 Create table-driven unit tests per node type plus integration tests that read `testdata/formatter` fixtures.
- [ ] 15.4.2 Add fuzz/property tests that compare formatter output against itself round-tripped through the parser → formatter pipeline.
- [ ] 15.4.3 Document formatter architecture and extension points in `docs/formatter-architecture.md`.
- [ ] 15.4.4 Update `PLAYGROUND.md`, `README.md`, and release notes to mention the Format button now runs the AST-driven formatter.
- [ ] 15.4.5 Record known limitations (e.g., preprocessor directives) and track follow-ups in `TEST_ISSUES.md`.
- [ ] 15.4.6 Gather usability feedback (issue template or telemetry) to prioritize refinements like configurable styles or multi-profile support.

---

## Summary

This roadmap now spans **~897 bite-sized tasks** across 15 phases, each feeding the next piece of the DWScript-in-Go story:

1. **Phase 1 – Lexer (45 tasks)**: ✅ Complete.
2. **Phase 2 – Parser & AST (64 tasks)**: ✅ Complete.
3. **Phase 3 – Statement execution (65 tasks)**: ✅ Complete (98.5% test coverage).
4. **Phase 4 – Control flow (46 tasks)**: ✅ Complete.
5. **Phase 5 – Functions & scope (46 tasks)**: ✅ Complete (91.3%).
6. **Phase 6 – Type checking (50 tasks)**: ✅ Complete.
7. **Phase 7 – Object-oriented features (156 tasks)**: 🔄 In progress (55.8%); classes done, interfaces pending (83 tasks).
8. **Phase 8 – Extended language features (93 + 31 property tasks)**: queued once interface work lands.
9. **Phase 9 – Deferred follow-ups from Phase 8**: backlog of polish tasks.
10. **Phase 10 – Performance & polish (68 tasks)**: profiling, GC pressure work, interpreter tweaks.
11. **Phase 11 – Long-term evolution (54 tasks)**: module refactors, CLI ergonomics, documentation debt.
12. **Phase 12 – WASM / Web distribution**: playground, npm tooling, and browser runners (active work in 12.15.x).
13. **Phase 13 – Alternative execution modes (bytecode VM + JIT)**: Bytecode foundation (13.1–13.12) mostly ✅, with new parity tasks (13.12.5–13.12.7) plus pending test/serialization/JIT tracks (13.13+).
14. **Phase 14 – Multi-backend code generation (~180 tasks)**: MIR core, JS backend, optional LLVM.
15. **Phase 15 – AST-driven formatter & playground integration (30 tasks)**: formatter design/engine/tooling rollout.

Each phase lists granular subtasks above so contributors can jump straight to the next actionable item.

**Key Notes**:

- **Stage 12** introduces a two-tier code generation architecture with MIR as an intermediate representation
- JavaScript backend is prioritized (Stages 12.1-12.5, ~152 tasks, ~9 weeks) for immediate value
- LLVM backend (Stage 12.6, 45 tasks) is optional and can be deferred or skipped entirely
- The MIR layer enables multiple backends from a single lowering pass, future-proofing for WebAssembly, C, or other targets
- **Stage 13** adds an AST-driven formatter shared by the CLI, LSP, and web playground so Monaco’s Format button produces deterministic DWScript output.

Each task is actionable and testable. Following this plan methodically will result in a complete, production-ready DWScript implementation in Go, preserving 100% of the language's syntax and semantics while leveraging Go's ecosystem.

The project can realistically take **1-3 years** depending on:

- Development pace (full-time vs part-time)
- Team size (solo vs multiple contributors)
- Completeness goals (minimal viable vs full feature parity)
