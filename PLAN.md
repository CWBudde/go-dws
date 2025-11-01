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

## Phase 9: Deferred Stage 8 Tasks

**Status**: 79 completed tasks compacted into summary format | 79+ remaining completed tasks visible | 501 incomplete tasks

**Completed & Compacted Features**:

- **Type Aliases** (11 tasks): Full support for type alias declarations (`type TUserID = Integer`), semantic analysis, and runtime resolution
- **Const Declarations** (9 tasks): Constant declarations with immutability enforcement, semantic checking, and CLI integration
- **Built-in Functions** (4 tasks): Core runtime functions (Length, Copy, Pos, Delete, Insert) with comprehensive test coverage
- **Priority Array Functions** (1 task): High-value array manipulation functions integrated into standard library
- **Type Expression Parser** (4 tasks): Foundational infrastructure for parsing complex inline type expressions
- **Inline Function Pointer Types** (5 tasks): Support for inline function pointer syntax without requiring type aliases
- **Inline Array Types** (7 tasks): Dynamic and static array type expressions in variable/parameter declarations
- **Subrange Types** (12 tasks): Complete subrange type implementation with bounds checking and semantic validation
- **Lambdas/Anonymous Methods** (26 tasks): Full lambda expression support with both block and shorthand syntax
- **Units/Modules System** (43 tasks): Partial implementation of units/modules system for code organization (43 tasks complete)
- **Function/Method Pointers** (26 tasks): Complete function and method pointer support for callbacks and higher-order functions (26 tasks complete)
- **Bitwise Operators** (7 tasks): Bitwise AND, OR, XOR, SHL, SHR operators with parser, semantic, and interpreter support
- **External Function Registration / FFI** (23 tasks): Complete Foreign Function Interface with `RegisterFunction()` API, bidirectional type marshaling (primitives, arrays, maps), panic/error recovery to EHost exceptions, comprehensive test coverage (13 subtests for panic handling alone)

**Implementation Summary**: Phase 9 extended the type system with aliases, subranges, and inline type expressions; added const declarations with semantic enforcement; implemented lambda expressions with capture semantics; enriched the standard library with essential built-in and array functions; delivered complete FFI system enabling Go function calls from DWScript with automatic type conversion and exception handling. Several major features remain in progress including the units/modules system (42 tasks complete), function pointers (26 tasks complete), and various type system enhancements. All completed features include comprehensive parser, semantic analyzer, interpreter, and CLI integration with dedicated test suites.

---

Targeted backlog from Stage 8 that still needs implementation or polish.

### Indexed & Expression-Based Properties

- [ ] 9.1 Support indexed property reads end-to-end:
  - [ ] 9.1a Parser/AST: keep index parameter metadata on `PropertyDecl`/`MemberAccess` nodes.
  - [ ] 9.1b Semantic analysis: evaluate index expression types, ensure getter signatures include matching parameters, and surface DWScript-style diagnostics when mismatched.
  - [ ] 9.1c Interpreter: evaluate index expressions at runtime and pass them to the bound getter field/method while preserving inheritance lookup rules.
- [ ] 9.2 Support indexed property writes:
  - [ ] 9.2a Semantic analysis: validate setter signatures (value + index params) and enforce read/write pairing rules.
  - [ ] 9.2b Interpreter: evaluate indices, pass them plus the assigned value to the setter, and propagate errors when setters are missing.
- [ ] 9.3 Implement expression-based property getters (e.g., `read (FValue * 2)`): extend the parser to capture inline expressions, make the analyzer type-check them in the owning class scope, and run them via the interpreter with `Self` bound.
- [ ] 9.4 Finish fixtures/tests for the deferred property modes:
  - [ ] 9.4a `testdata/properties/indexed_property.dws` covering array-like accessors.
  - [ ] 9.4b `testdata/properties/expression_property.dws` covering computed getters.
  - [ ] 9.4c `testdata/properties/default_property.dws` covering default indexed properties.
  - [ ] 9.4d CLI tests in `cmd/dwscript/properties_test.go` asserting outputs for the new fixtures.

### Record Methods

- [ ] 9.5 Allow record declarations to contain methods: update the parser/AST to reuse function declarations inside `RecordDecl`, including constructor-like routines.
- [ ] 9.6 Extend semantic analysis so record methods get their own scope, bind `Self`, and can access fields/properties just like class methods.
- [ ] 9.7 Teach the interpreter to invoke record methods (either by desugaring to hidden functions or by integrating with the class-method call path) and add focused unit/fixture coverage.

### Enum & Set Scalability

- [ ] 9.8 Introduce a `map[int]bool` (or similar) backing for large enums/sets (>64 values) so set operations remain efficient; update `types.SetType`, analyzer compatibility checks, and runtime values accordingly.
- [ ] 9.9 Support `for-in` iteration over sets (`for e in MySet do`):
  - [ ] 9.9a Semantic analysis: ensure the loop variable type matches the set element type.
  - [ ] 9.9b Interpreter: iterate deterministically over the set contents (respecting static vs dynamic storage backends).
- [ ] 9.10 Add regression tests for large-set storage and `for-in` loops across `types/`, `semantic/`, `interp/`, and integration fixtures.

### Exception UX Polishing

- [ ] 9.11 Display unhandled exception messages in CLI output (class + `Message` text) so behavior matches DWScript; update the `errors/` formatter and add CLI tests.
- [ ] 9.12 Finish semantic enforcement that `break`, `continue`, and `exit` are illegal inside `finally` blocks:
  - [x] 9.12a Detect `return` in finally blocks and emit semantic errors.
  - [x] 9.12b Allow `raise` but otherwise require finally blocks to complete normally.
  - [ ] 9.12c Now that break/continue/exit parsing exists, add explicit checks and tests covering those control-flow exits.

---

#### Advanced Features (6 tasks)

- [ ] 9.13 Support variadic Go functions:
  - [ ] Detect `...` parameter in Go signature
  - [ ] Accept variable number of DWScript arguments
  - [ ] Pack into slice for Go function
- [ ] 9.14 Support optional parameters:
  - [ ] Default values in DWScript
  - [ ] Map to Go function overloads or optional args
- [ ] 9.15 Support by-reference parameters:
  - [ ] `var` parameters in DWScript
  - [ ] Pointers in Go
  - [ ] Sync changes back to DWScript after call
- [ ] 9.16 Support registering Go methods:
  - [ ] Methods on Go structs
  - [ ] Bind receiver automatically
- [ ] 9.17 Support callback functions:
  - [ ] DWScript function pointers passed to Go
  - [ ] Go can call back into DWScript
  - [ ] Handle re-entrancy
- [ ] 9.18 Add tests for advanced features

---

#### Full Contextual Type Inference (FUTURE ENHANCEMENT)

**Summary**: Task 9.? currently has a placeholder implementation that reports an error when lambda parameters lack type annotations. Full contextual type inference would allow the compiler to infer parameter types from the context where the lambda is used.

**Current Status**: Lambda parameter type inference reports "not fully implemented" error. Return type inference from body is complete.

**Tasks for Full Implementation** (5 tasks):

- [ ] 9.67 Add type context passing infrastructure to expression analyzer:
  - [ ] Modify `analyzeExpression()` to accept optional `expectedType` parameter
  - [ ] Thread expected type through all expression analysis calls
  - [ ] Maintain backward compatibility (default to nil for existing calls)
  - [ ] Update all expression analyzers to use context when available
- [ ] 9.68 Implement assignment context type inference:
  - [ ] Detect when lambda is assigned to typed variable: `var f: TFunc := lambda(x) => x * 2`
  - [ ] Extract function pointer type from variable declaration
  - [ ] Pass parameter types to lambda analyzer
  - [ ] Apply inferred types to untyped parameters
  - [ ] Validate inferred types match if some params are explicitly typed
- [ ] 9.69 Implement function call context type inference:
  - [ ] Detect when lambda is passed as function argument: `Apply(5, lambda(n) => n * 2)`
  - [ ] Extract expected function pointer type from function parameter
  - [ ] Pass parameter types to lambda analyzer
  - [ ] Apply inferred types to untyped parameters
  - [ ] Handle overloaded functions (try each signature)
- [ ] 9.70 Implement return statement context type inference:
  - [ ] Detect when lambda is returned from function with known return type
  - [ ] Extract function pointer type from return type
  - [ ] Apply to lambda parameters
- [ ] 9.71 Add comprehensive tests for contextual type inference:
  - [ ] Assignment context tests
  - [ ] Function call context tests
  - [ ] Return context tests
  - [ ] Error cases (ambiguous context, conflicting types)
  - [ ] Mixed typed/untyped parameters

**Example Usage After Full Implementation**:
```pascal
// Assignment context inference
type TComparator = function(a, b: Integer): Integer;
var cmp: TComparator := lambda(x, y) => x - y;  // x, y inferred as Integer

// Function call context inference
function Apply(x: Integer; f: function(Integer): Integer): Integer;
begin
  Result := f(x);
end;

var result := Apply(5, lambda(n) => n * 2);  // n inferred as Integer

// Mixed explicit and inferred
var f: TComparator := lambda(x: Integer, y) => x - y;  // y inferred as Integer
```

**Design Considerations**:
- Type inference should be unidirectional (context → lambda), not bidirectional
- Inference fails gracefully with clear error messages
- Partial inference supported (some params typed, some inferred)
- Overload resolution attempted in order of declaration
- Performance: minimal impact as context passing is optional

**Dependencies**: None (can be implemented independently)

**Estimated Effort**: Medium (3-5 days) - requires refactoring expression analyzer API

- [ ] 9.72 Dynamic array literal syntax for integer arrays:
  - [ ] Cannot use: `var nums := [1, 2, 3, 4, 5];` (parsed as SET, not array)
  - [ ] Current workaround: Use SetLength + manual assignment
  - [ ] Location: Parser interprets `[...]` as set literals (enum values only)
  - [ ] Impact: Cannot easily create test data or initialize arrays
  - [ ] Blocks: testdata/lambdas/higher_order.dws execution
  - [ ] Requires type inference infrastructure from tasks 9.234-9.238
  - [ ] Distinguish array literals from set literals based on expected type context
  - [ ] Update `parseSetLiteral()` to conditionally return `ArrayLiteral` AST node
  - [ ] Add semantic analysis for array literal type checking
  - [ ] Add interpreter support for array literal evaluation

---

### Helpers (Class/Record/Type) (MEDIUM PRIORITY)

**Summary**: Implement helper types to extend existing types with additional methods without modifying the original type declaration.

**Example**:
```pascal
type
  TStringHelper = helper for String
    function ToUpper: String;
    function ToLower: String;
  end;

function TStringHelper.ToUpper: String;
begin
  Result := UpperCase(Self);
end;

var s := 'hello';
PrintLn(s.ToUpper()); // Output: HELLO
```

**Reference**: docs/missing-features-recommendations.md lines 279-283

#### Type System (3 tasks) ✅ COMPLETE

- [x] 9.73 Define `HelperType` in `types/compound_types.go`: ✅ DONE
  - [x] Fields: `Name`, `TargetType`, `Methods`, `Properties`, `ClassVars`, `ClassConsts`, `IsRecordHelper`
  - [x] Implements `Type` interface (`String()`, `TypeKind()`, `Equals()`)
  - [x] `TypeKind()` returns `"HELPER"`
  - [x] `String()` returns `"record helper for TypeName"` or `"helper for TypeName"`
  - [x] Helper methods: `GetMethod()`, `GetProperty()`, `GetClassVar()`, `GetClassConst()`
  - [x] Constructor: `NewHelperType(name, targetType, isRecordHelper)`
- [x] 9.74 Implement helper method resolution: ✅ DONE
  - [x] Implemented in semantic analysis (see task 9.83)
  - [x] Helpers registered in analyzer's `helpers` map
  - [x] Priority: instance methods > helper methods
  - [x] Multiple helpers can extend same type
- [x] 9.75 Add tests for helper types: ✅ DONE
  - [x] Covered in semantic tests (task 9.85)

#### AST Nodes (2 tasks)

- [x] 9.76 Create `ast/helper.go`: ✅ DONE
  - [x] Define `HelperDecl` struct with fields:
    - `Name *Identifier` - helper type name
    - `ForType *TypeAnnotation` - target type being extended
    - `IsRecordHelper bool` - true for "record helper", false for "helper"
    - `Methods []*FunctionDecl` - helper methods
    - `Properties []*PropertyDecl` - helper properties
    - `ClassVars []*FieldDecl` - class variables
    - `ClassConsts []*ConstDecl` - class constants
    - `PrivateMembers []Statement` - private section members
    - `PublicMembers []Statement` - public section members
  - [x] Implement `Statement` interface
  - [x] Implement `String()` method with proper formatting
- [x] 9.77 Add AST tests ✅ DONE
  - [x] Test `HelperDecl.String()` for various helper types
  - [x] Test record helper vs plain helper syntax
  - [x] Test helpers with methods, properties, class vars, class consts
  - [x] Test visibility sections (private/public)

#### Lexer Support (1 task)

- [x] 9.78 Add `helper` keyword to lexer: ✅ DONE (was already present)
  - [x] HELPER token type already defined in `lexer/token_type.go`
  - [x] Already registered in keyword map

#### Parser Support (3 tasks)

- [x] 9.79 Implement helper parsing in `parser/helpers.go`: ✅ DONE
  - [x] Detect `record helper for` pattern in `parseRecordOrHelperDeclaration()`
  - [x] Detect `helper for` pattern in `parseTypeDeclaration()`
  - [x] Parse helper name (already provided by caller)
  - [x] Parse `for` keyword and target type
  - [x] Parse method declarations, properties, class vars, class consts
  - [x] Support visibility sections (private/public)
  - [x] Expect `end;`
  - [x] Return `HelperDecl` node
- [x] 9.80 Add parser tests for helpers ✅ DONE
  - [x] Test simple helper and record helper syntax
  - [x] Test helpers with multiple methods
  - [x] Test helpers with properties, class vars, class consts
  - [x] Test visibility sections
  - [x] Test error cases (missing 'for', missing type, missing 'end')
  - [x] Test distinguishing between record and helper declarations
- [x] 9.81 Test class helpers and record helpers separately ✅ DONE
  - [x] Created `testdata/helpers/` directory with comprehensive test files
  - [x] `basic_helper.dws` - simple helper syntax
  - [x] `record_helper.dws` - "record helper" syntax variant
  - [x] `properties_helper.dws` - helpers with properties
  - [x] `visibility_helper.dws` - private/public sections
  - [x] `multiple_helpers.dws` - multiple helpers for same type
  - [x] `class_members_helper.dws` - class vars and consts
  - [x] All test files parse successfully

#### Semantic Analysis (4 tasks) ✅ COMPLETE

- [x] 9.82 Create `semantic/analyze_helpers.go`: ✅ DONE
  - [x] Analyze helper declarations
  - [x] Resolve target type (`for` type) - validates target type exists
  - [x] Validate helper methods (parameters, return types)
  - [x] Register helper in type environment (helpers map indexed by target type)
  - [x] Process helper properties, class vars, and class constants
  - [x] Detect duplicate methods/properties/members
- [x] 9.83 Implement helper method resolution: ✅ DONE
  - [x] Added `helpers` map to Analyzer (type name -> []*HelperType)
  - [x] Implemented `getHelpersForType()`, `hasHelperMethod()`, `hasHelperProperty()`
  - [x] Extended `analyzeMemberAccessExpression()` to check helpers
  - [x] Extended `analyzeMethodCallExpression()` to check helpers
  - [x] Extended `analyzeRecordFieldAccess()` to check helpers
  - [x] Helpers checked after instance members (priority: instance > helper)
  - [x] Multiple helpers can extend the same type (all contribute methods)
  - [x] Works for all types: classes, records, and basic types (String, Integer, etc.)
- [x] 9.84 Implement `Self` binding in helper methods: ✅ DONE
  - [x] `Self` binding documented in helper method analysis
  - [x] Target type stored in HelperType for Self type validation
  - [x] Note: Full Self validation deferred to interpreter stage
- [x] 9.85 Add semantic tests for helpers: ✅ DONE
  - [x] Created `semantic/helpers_test.go` with comprehensive tests:
    - TestHelperDeclaration: 5 test cases (simple helper, record helper, unknown type, class var, class const)
    - TestHelperMethodResolution: 4 test cases (String/Integer helpers, non-existent method, properties)
    - TestMultipleHelpers: multiple helpers extending same type
    - TestHelperMethodParameters: 3 test cases (correct params, wrong count, wrong type)
  - [x] All tests passing ✅

#### Interpreter Support (4 tasks) ✅ COMPLETE

- [x] 9.86 Implement helper method dispatch: ✅ DONE
  - [x] Created `interp/helpers.go` with complete helper support
  - [x] Implemented `findHelperMethod()` and `callHelperMethod()`
  - [x] Updated `evalMethodCall()` to check helpers for non-object types
  - [x] Updated `evalMemberAccess()` to check helpers for properties
  - [x] Bind `Self` to the target value (object, record, or basic type)
  - [x] Execute helper method with `Self` bound
  - [x] Class vars and consts accessible within helper methods
- [x] 9.87 Implement helper method storage: ✅ DONE
  - [x] Added `helpers` map to Interpreter (type name -> []*HelperInfo)
  - [x] Store helpers in registry indexed by target type
  - [x] Implemented `evalHelperDeclaration()` to register helpers
  - [x] Lookup helpers at method call time via `getHelpersForValue()`
  - [x] Support for multiple helpers extending the same type
- [x] 9.88 Add tests in `interp/helpers_test.go`: ✅ DONE
  - [x] Created comprehensive test suite with 12 test functions
  - [x] TestHelperMethod: basic helper method on String
  - [x] TestHelperMethodWithParameters: parameters validation
  - [x] TestHelperMethodOnInteger: basic type helpers
  - [x] TestHelperMethodOnRecord: record helpers
  - [x] TestHelperProperty: helper properties
  - [x] TestHelperClassConst/ClassVar: class member support
  - [x] TestHelperSyntaxVariations: "helper" vs "record helper"
  - [x] TestHelperMethodNotFound: error cases
  - [x] TestHelperOnClassInstancePrefersMethods: priority rules
  - [x] All tests passing ✅
- [x] 9.89 Test helper method calls and `Self` binding: ✅ DONE
  - [x] Self binding tested in all helper method tests
  - [x] Self correctly references target value (Integer, String, Record, Class)
  - [x] Verified with manual tests using dwscript CLI

#### Testing & Fixtures (3 tasks) ✅ COMPLETE

- [x] 9.90 Create test scripts in `testdata/helpers/`: ✅ DONE
  - [x] `string_helper.dws` - String helper with methods and expected output
  - [x] `integer_helper.dws` - Integer helper demonstrating various methods
  - [x] `record_helper_demo.dws` - Record helper with distance calculations
  - [x] `class_helper_demo.dws` - Class helper extending TPerson
  - [x] `multiple_helpers_demo.dws` - Multiple helpers on same type
  - [x] `class_constants_demo.dws` - Helper with class constants (PI, E)
  - [x] All scripts tested and working correctly
- [x] 9.91 Add CLI integration tests: ✅ DONE
  - [x] Created `cmd/dwscript/helpers_integration_test.go`
  - [x] TestHelpersScriptsExist: verifies all test scripts exist
  - [x] TestHelpersParsing: validates parsing of all helper scripts
  - [x] TestHelpersExecution: validates output of all helper scripts
  - [x] TestHelperMethodDispatch: tests inline helper method dispatch
  - [x] TestHelperSyntaxVariations: tests both syntax variants
  - [x] All tests passing ✅
- [x] 9.92 Document helpers in `docs/helpers.md`: ✅ DONE
  - [x] Comprehensive documentation with examples
  - [x] Syntax reference for both variants
  - [x] Feature documentation (methods, properties, class vars/consts)
  - [x] Examples for String, Integer, Record, and Class helpers
  - [x] Implementation details and limitations
  - [x] Testing and reference sections

---

### DateTime Functions (MEDIUM PRIORITY) ✅ COMPLETE

**Summary**: Implement comprehensive date/time functionality including current date/time, formatting, parsing, and arithmetic operations.

**Reference**: docs/missing-features-recommendations.md lines 289-292

#### Type System (2 tasks) ✅

- [x] 9.93 Define `TDateTime` type:
  - [x] Internal representation: float (days since 1899-12-30, like Delphi)
  - [x] Fractional part represents time
  - [x] Add to type system as primitive type
- [x] 9.94 Add tests for DateTime type

#### Built-in Functions - Current Date/Time (4 tasks) ✅

- [x] 9.95 Implement `Now(): TDateTime`:
  - [x] Returns current date and time
  - [x] Use Go's `time.Now()`
  - [x] Convert to TDateTime format
- [x] 9.96 Implement `Date(): TDateTime`:
  - [x] Returns current date (time part = 0.0)
- [x] 9.97 Implement `Time(): TDateTime`:
  - [x] Returns current time (date part = 0.0)
- [x] 9.98 Add tests for Now/Date/Time

#### Built-in Functions - Date Construction (4 tasks) ✅

- [x] 9.99 Implement `EncodeDate(year, month, day: Integer): TDateTime`:
  - [x] Construct date from components
  - [x] Validate inputs (valid month, day)
  - [x] Return TDateTime value
- [x] 9.100 Implement `EncodeTime(hour, min, sec, msec: Integer): TDateTime`:
  - [x] Construct time from components
  - [x] Validate inputs
- [x] 9.101 Implement `EncodeDateTime(y, m, d, h, min, s, ms: Integer): TDateTime`:
  - [x] Combine date and time
- [x] 9.102 Add tests for date construction

#### Built-in Functions - Date Extraction (4 tasks) ✅

- [x] 9.103 Implement `DecodeDate(dt: TDateTime; var y, m, d: Integer)`:
  - [x] Extract year, month, day components
  - [x] Use var parameters to return multiple values
- [x] 9.104 Implement `DecodeTime(dt: TDateTime; var h, min, s, ms: Integer)`:
  - [x] Extract time components
- [x] 9.105 Implement component functions:
  - [x] `YearOf(dt: TDateTime): Integer`
  - [x] `MonthOf(dt: TDateTime): Integer`
  - [x] `DayOf(dt: TDateTime): Integer`
  - [x] `HourOf(dt: TDateTime): Integer`
  - [x] `MinuteOf(dt: TDateTime): Integer`
  - [x] `SecondOf(dt: TDateTime): Integer`
- [x] 9.106 Add tests for date extraction

#### Built-in Functions - Formatting (3 tasks) ✅

- [x] 9.107 Implement `FormatDateTime(fmt: String, dt: TDateTime): String`:
  - [x] Support format specifiers: `yyyy`, `mm`, `dd`, `hh`, `nn`, `ss`
  - [x] Example: `FormatDateTime('yyyy-mm-dd', Now())` → '2025-10-27'
  - [x] Use Go's time formatting internally
- [x] 9.108 Implement `DateToStr(dt: TDateTime): String`:
  - [x] Default date format
- [x] 9.109 Implement `TimeToStr(dt: TDateTime): String`:
  - [x] Default time format

#### Built-in Functions - Parsing (2 tasks) ✅

- [x] 9.110 Implement `StrToDate(s: String): TDateTime`:
  - [x] Parse date string
  - [x] Support common formats
  - [x] Raise error on invalid input
- [x] 9.111 Implement `StrToDateTime(s: String): TDateTime`:
  - [x] Parse date/time string

#### Built-in Functions - Date Arithmetic (3 tasks) ✅

- [x] 9.112 Implement date arithmetic operators:
  - [x] `dt1 - dt2` → days difference (Float)
  - [x] `dt + days` → new date
  - [x] `dt - days` → new date
- [x] 9.113 Implement `IncDay(dt: TDateTime, days: Integer): TDateTime`:
  - [x] Add days to date
  - [x] Similar: `IncMonth`, `IncYear`, `IncHour`, `IncMinute`, `IncSecond`
- [x] 9.114 Implement `DaysBetween(dt1, dt2: TDateTime): Integer`:
  - [x] Calculate difference in days
  - [x] Similar: `HoursBetween`, `MinutesBetween`, `SecondsBetween`

#### Testing & Fixtures (2 tasks) ✅

- [x] 9.115 Create test scripts in `testdata/datetime/`:
  - [x] `current_datetime.dws` - Now, Date, Time
  - [x] `encode_decode.dws` - EncodeDate, DecodeDate
  - [x] `formatting.dws` - FormatDateTime
  - [x] `arithmetic.dws` - Date arithmetic
  - [x] Expected outputs
- [x] 9.116 Add CLI integration tests

---

### Variant Type System (HIGH PRIORITY - FOUNDATIONAL)

**Summary**: Implement DWScript's Variant type for dynamic, heterogeneous value storage. This is a foundational feature required for full DWScript compatibility and enables many advanced features like `array of const`, JSON support, and COM interop.

**Current Status**:
- ✅ Type definition complete (Tasks 9.220-9.222): VariantType and VariantValue implemented
- ✅ Semantic analysis complete (Tasks 9.223-9.226): Variable declarations, assignments, array of const support
- ✅ Runtime support complete (Tasks 9.227-9.231): Boxing/unboxing, operators, comparisons implemented
- ✅ Built-in functions complete (Tasks 9.232-9.234): VarType(), VarIsNull(), VarIsEmpty(), VarIsNumeric(), VarToStr(), VarToInt(), VarToFloat(), VarAsType()
- ✅ Migration complete (Tasks 9.235-9.237): ConstType replaced with Variant; Format() function migrated to use array of Variant
- ⏳ Testing & Documentation pending (Tasks 9.238-9.239): Comprehensive test suite and documentation

**Context**: DWScript (like Delphi) has a `Variant` type that can hold any value at runtime with dynamic type checking. It's used extensively for:
- `array of const` parameters (heterogeneous argument lists)
- COM/OLE automation
- Dynamic scripting scenarios
- JSON/dynamic data structures
- Database field values

**Reference**: Original DWScript implementation in `reference/dwscript-original/Source/dwsVariant.pas`

#### Current Workaround (Task 9.156 - TEMPORARY)

**Implementation**: Added `ConstType` as a marker type to allow heterogeneous arrays specifically for builtin functions:

```go
// internal/types/types.go
type ConstType struct{}  // Marker type, not a real Variant
var CONST = &ConstType{}
var ARRAY_OF_CONST = NewDynamicArrayType(CONST)

// Semantic analyzer passes ARRAY_OF_CONST as expected type for Format's second argument
// This allows ['string', 123, 3.14] to type-check
```

**Limitations of Workaround**:
- Only works in specific contexts (Format function arguments)
- No runtime Variant value representation
- Cannot declare `var v: Variant;` variables
- Cannot assign Variant values
- Not usable in user-defined functions with `array of const` parameters
- No VarType(), VarAsType(), VarIsNull() builtin functions
- Diverges from DWScript compatibility

**Migration Path**: Once proper Variant type is implemented:
1. Replace `ConstType` with `VariantType`
2. Replace `ARRAY_OF_CONST` with `NewDynamicArrayType(VARIANT)`
3. Update Format() and other builtins to use Variant-based arrays
4. Add runtime Variant value representation (VarRec in Delphi)
5. Remove workaround code

#### Variant Type System Design (15 tasks)

**Design Goals**:
- Full DWScript compatibility for Variant type
- Efficient runtime representation
- Type-safe conversions
- Support for all DWScript types (Integer, Float, String, Boolean, Object, Array, Record)

##### Type Definition (3 tasks)

- [x] 9.220 Define `VariantType` in `internal/types/types.go`:
  - [x] Implement Type interface (String, Equals, TypeKind)
  - [x] Add singleton `var VARIANT = &VariantType{}`
  - [x] Document that Variant can hold any runtime value

- [x] 9.221 Define runtime Variant value in `internal/interp/value.go`:
  - [x] `type VariantValue struct { Value Value; ActualType types.Type }`
  - [x] Wraps any other Value type with dynamic type tracking
  - [x] Similar to Delphi's TVarData / VarRec structures

- [x] 9.222 Add Variant type tests:
  - [x] Test Variant type equality
  - [x] Test Variant in type resolution
  - [x] Test Variant with type aliases

##### Semantic Analysis (4 tasks)

- [x] 9.223 Support Variant variable declarations:
  - [x] `var v: Variant;` - declares uninitialized Variant
  - [x] `var v: Variant := 42;` - initializes with Integer
  - [x] `var v: Variant := 'hello';` - initializes with String
  - [x] Update `analyzeVarDecl` to handle Variant type (already worked via resolveType)

- [x] 9.224 Implement Variant assignment rules in `canAssign()`:
  - [x] Any type can be assigned TO Variant (implicit boxing)
  - [x] Variant can be assigned FROM with runtime type checking
  - [x] Variant-to-Variant assignment preserves wrapped value

- [x] 9.225 Support `array of const` parameter type:
  - [x] Parse `array of const` as special array type (uses CONST/VARIANT)
  - [x] Equivalent to `array of Variant` semantically
  - [x] Allow in function/procedure parameter lists
  - [x] Updated array literal analyzer to support VARIANT element type

- [x] 9.226 Add semantic analysis tests:
  - [x] Test Variant variable declarations and assignments (15 tests)
  - [x] Test heterogeneous array literals with Variant element type
  - [x] Test `array of const` parameters (via array of Variant)
  - [x] Test type aliases with Variant

##### Runtime Support (5 tasks)

- [x] 9.227 Implement VariantValue boxing in interpreter:
  - [x] Box primitive values (Integer → VariantValue)
  - [x] Box complex values (Arrays, Records, Objects)
  - [x] Preserve type information for unboxing

- [x] 9.228 Implement VariantValue unboxing in interpreter:
  - [x] Unbox to expected type with runtime checking
  - [x] Implicit conversions (Integer → Float, String → Integer)
  - [x] Raise runtime error on invalid conversion

- [x] 9.229 Implement Variant arithmetic operators:
  - [x] Variant + Variant → numeric promotion rules
  - [x] Variant * Variant → follows Delphi semantics
  - [x] String concatenation with Variant
  - [x] Handle type mismatches at runtime

- [x] 9.230 Implement Variant comparison operators:
  - [x] Variant = Variant → value equality with type coercion
  - [x] Variant <> Variant → inequality
  - [x] Variant < Variant → numeric/string comparison
  - [x] Boolean result type

- [x] 9.231 Add runtime tests:
  - [x] Test Variant value boxing/unboxing
  - [x] Test Variant arithmetic and comparisons
  - [x] Test Variant in arrays and records
  - [x] Test runtime type errors

##### Built-in Functions (3 tasks)

- [x] 9.232 Implement Variant introspection functions:
  - [x] `VarType(v: Variant): Integer` - returns type code
  - [x] `VarIsNull(v: Variant): Boolean` - checks if uninitialized
  - [x] `VarIsEmpty(v: Variant): Boolean` - checks if empty
  - [x] `VarIsNumeric(v: Variant): Boolean` - checks if numeric type

- [x] 9.233 Implement Variant conversion functions:
  - [x] `VarAsType(v: Variant, varType: Integer): Variant` - explicit conversion
  - [x] `VarToStr(v: Variant): String` - convert to string
  - [x] `VarToInt(v: Variant): Integer` - convert to integer
  - [x] `VarToFloat(v: Variant): Float` - convert to float

- [x] 9.234 Add builtin function tests:
  - [x] Test VarType with different value types
  - [x] Test VarIsNull, VarIsEmpty, VarIsNumeric
  - [x] Test VarToStr, VarToInt, VarToFloat conversions
  - [x] Test error handling for invalid conversions

#### Migration from ConstType Workaround (3 tasks)

- [x] 9.235 Replace ConstType with VariantType:
  - [x] Change `CONST` singleton to `VARIANT`
  - [x] Update `ARRAY_OF_CONST` to use VARIANT element type
  - [x] Keep ConstType struct as deprecated for backward compatibility

- [x] 9.236 Update Format() function to use Variant arrays:
  - [x] Update semantic analysis to expect `array of Variant`
  - [x] Update runtime to unbox Variant values for formatting
  - [x] Ensure all format specifiers work with Variant values

- [x] 9.237 Verify Format test suite with Variant implementation:
  - [x] Run Format() unit tests (30 tests passing)
  - [x] Verify all heterogeneous array cases work
  - [x] Test with demo script showing all format specifiers
  - [x] Confirm no regressions in existing tests

#### Testing & Documentation (2 tasks)

- [ ] 9.238 Create comprehensive Variant test suite:
  - [ ] Create `testdata/variant/basic.dws` - declarations, assignments
  - [ ] Create `testdata/variant/arithmetic.dws` - operations
  - [ ] Create `testdata/variant/conversions.dws` - type conversions
  - [ ] Create `testdata/variant/array_of_const.dws` - heterogeneous arrays
  - [ ] Expected output files

- [ ] 9.239 Document Variant type in `docs/variant.md`:
  - [ ] Variant type overview and use cases
  - [ ] Boxing and unboxing semantics
  - [ ] Type conversion rules
  - [ ] `array of const` parameter pattern
  - [ ] Built-in Variant functions reference
  - [ ] Comparison with Delphi/DWScript Variant behavior
  - [ ] Performance considerations

**Total**: 20 tasks
- ✅ Type Definition: 3 tasks (9.220-9.222) - COMPLETE
- ✅ Semantic Analysis: 4 tasks (9.223-9.226) - COMPLETE
- ✅ Runtime Support: 5 tasks (9.227-9.231) - COMPLETE
- ✅ Built-in Functions: 3 tasks (9.232-9.234) - COMPLETE
- ✅ Migration from ConstType: 3 tasks (9.235-9.237) - COMPLETE
- ⏳ Testing & Documentation: 2 tasks (9.238-9.239) - PENDING

**Progress**: 18 of 20 tasks complete (90%)

**Dependencies**:
- None (foundational feature)

**Enables**:
- ✅ Task 9.156 (Format function) - now properly implemented with array of Variant
- JSON Support (requires Variant for dynamic values)
- COM/OLE Automation (if implemented)
- Database integration (Variant field values)
- User-defined functions with `array of const` parameters

**Priority**: HIGH - Many features depend on this, current workaround is fragile

---

### JSON Support (MEDIUM PRIORITY)

**Summary**: Implement JSON parsing and serialization for modern data interchange. Enables DWScript to work with JSON APIs and configuration files.

**Reference**: docs/missing-features-recommendations.md lines 294-297

#### Type System (2 tasks)

- [ ] 9.117 Define JSON value representation:
  - [ ] Variant type that can hold any JSON value
  - [ ] Support: null, boolean, number, string, array, object
  - [ ] Map to DWScript types where possible
- [ ] 9.118 Add tests for JSON type representation

#### Built-in Functions - Parsing (3 tasks)

- [ ] 9.119 Implement `ParseJSON(s: String): Variant`:
  - [ ] Parse JSON string
  - [ ] Return DWScript variant/dynamic type
  - [ ] Use Go's `encoding/json` internally
  - [ ] Map JSON types to DWScript types:
    - [ ] JSON object → dynamic record or map
    - [ ] JSON array → dynamic array
    - [ ] JSON primitives → corresponding DWScript types
- [ ] 9.120 Handle parsing errors:
  - [ ] Raise exception on invalid JSON
  - [ ] Include error position and message
- [ ] 9.121 Add tests for JSON parsing

#### Built-in Functions - Serialization (3 tasks)

- [ ] 9.122 Implement `ToJSON(value: Variant): String`:
  - [ ] Serialize DWScript value to JSON
  - [ ] Support records, arrays, primitives
  - [ ] Handle circular references (error or omit)
- [ ] 9.123 Implement `ToJSONFormatted(value: Variant, indent: Integer): String`:
  - [ ] Pretty-printed JSON with indentation
- [ ] 9.124 Add tests for JSON serialization

#### Built-in Functions - JSON Object Access (4 tasks)

- [ ] 9.125 Implement JSON object property access:
  - [ ] `jsonObject['propertyName']` syntax
  - [ ] Return value or nil if not found
- [ ] 9.126 Implement `JSONHasField(obj: Variant, field: String): Boolean`:
  - [ ] Check if JSON object has field
- [ ] 9.127 Implement `JSONKeys(obj: Variant): array of String`:
  - [ ] Return array of object keys
- [ ] 9.128 Implement `JSONValues(obj: Variant): array of Variant`:
  - [ ] Return array of object values

#### Built-in Functions - JSON Array Access (2 tasks)

- [ ] 9.129 Implement JSON array indexing:
  - [ ] `jsonArray[index]` syntax
  - [ ] Return element or nil if out of bounds
- [ ] 9.130 Implement `JSONLength(arr: Variant): Integer`:
  - [ ] Return array length

#### Type Mapping (2 tasks)

- [ ] 9.131 Document JSON ↔ DWScript type mapping:
  - [ ] JSON null → nil
  - [ ] JSON boolean → Boolean
  - [ ] JSON number → Integer or Float
  - [ ] JSON string → String
  - [ ] JSON array → dynamic array
  - [ ] JSON object → dynamic record or associative array
- [ ] 9.132 Handle edge cases:
  - [ ] Large numbers (beyond int64)
  - [ ] Special floats (NaN, Infinity)
  - [ ] Unicode escapes

#### Testing & Fixtures (2 tasks)

- [ ] 9.133 Create test scripts in `testdata/json/`:
  - [ ] `parse_json.dws` - Parse various JSON types
  - [ ] `to_json.dws` - Serialize to JSON
  - [ ] `json_object_access.dws` - Access object properties
  - [ ] `json_array_access.dws` - Access array elements
  - [ ] Expected outputs
- [ ] 9.134 Add CLI integration tests

---

### Improved Error Messages and Stack Traces (MEDIUM PRIORITY)

**Summary**: Enhance error reporting with better messages, stack traces, and debugging information. Improves developer experience significantly.

**Reference**: docs/missing-features-recommendations.md lines 299-302

#### Stack Trace Infrastructure (3 tasks)

- [ ] 9.135 Create `errors/stack_trace.go`:
  - [ ] Define `StackFrame` struct with `FunctionName`, `FileName`, `LineNumber`
  - [ ] Define `StackTrace` type as `[]StackFrame`
  - [ ] Implement `String()` method for formatted output
- [ ] 9.136 Implement stack trace capture in interpreter:
  - [ ] Track call stack during execution
  - [ ] Push frame on function entry
  - [ ] Pop frame on function exit
  - [ ] Capture stack on error/exception
- [ ] 9.137 Add tests for stack trace capture

#### Error Message Improvements (3 tasks)

- [ ] 9.138 Improve type error messages:
  - [ ] Before: "Type mismatch"
  - [ ] After: "Cannot assign Float to Integer variable 'count' at line 42"
  - [ ] Include expected and actual types
  - [ ] Include variable name and location
- [ ] 9.139 Improve runtime error messages:
  - [ ] Include expression that failed
  - [ ] Show values involved: "Division by zero: 10 / 0 at line 15"
  - [ ] Context from surrounding code
- [ ] 9.140 Add source code snippets to errors:
  - [ ] Show the line that caused error
  - [ ] Highlight error position with `^` or color
  - [ ] Show 1-2 lines of context

#### Exception Enhancements (2 tasks)

- [ ] 9.141 Add stack trace to exception objects:
  - [ ] Store `StackTrace` in exception
  - [ ] Display on uncaught exception
  - [ ] Format nicely for CLI output
- [ ] 9.142 Implement `GetStackTrace()` built-in:
  - [ ] Return current stack trace as string
  - [ ] Useful for logging and debugging

#### Debugging Information (2 tasks)

- [ ] 9.143 Add source position to all AST nodes:
  - [ ] Audit nodes for missing `Pos` fields
  - [ ] Add `EndPos` for better range reporting
  - [ ] Use in error messages
- [ ] 9.144 Implement call stack inspection:
  - [ ] `GetCallStack()` returns array of frame info
  - [ ] Each frame: function name, file, line
  - [ ] Accessible from DWScript code

#### Testing & Documentation (2 tasks)

- [ ] 9.145 Create test fixtures demonstrating error messages:
  - [ ] Type errors with clear messages
  - [ ] Runtime errors with stack traces
  - [ ] Exception handling with stack traces
  - [ ] Compare before/after error message quality
- [ ] 9.146 Document error message format in `docs/error-messages.md`

---

### Contracts (Design by Contract)

- [ ] 9.147 Parse require/ensure clauses (if supported)
- [ ] 9.148 Implement contract checking at runtime
- [ ] 9.149 Test contracts

### Comprehensive Testing (Stage 8)

- [ ] 9.150 Port DWScript's test suite (if available)
- [ ] 9.151 Run DWScript example scripts from documentation
- [ ] 9.152 Compare outputs with original DWScript
- [ ] 9.153 Fix any discrepancies
- [ ] 9.154 Create stress tests for complex features
- [ ] 9.155 Achieve >85% overall code coverage

### Format Function Testing (USING TEMPORARY WORKAROUND)

**Summary**: Create comprehensive test fixtures for the Format() built-in function. Originally deferred from task 9.52 due to heterogeneous array literal limitations. Now being implemented using temporary `ConstType` workaround until proper Variant type is available.

**Status**: IN PROGRESS - Using ConstType/ARRAY_OF_CONST workaround (see "Variant Type System" section above for proper implementation plan)

**Approach**: Implemented temporary semantic analyzer changes to accept heterogeneous array literals `['string', 123, 3.14]` specifically for Format() builtin. This allows Format tests to work but is NOT DWScript-compatible for general use.

**Migration Path**: Once tasks 9.220-9.237 (Variant Type System) are complete, this workaround will be replaced with proper `array of const` / `array of Variant` support.

#### Task Details (1 task)

- [x] 9.156 Create Format function test fixtures: ✅ COMPLETE
  - [x] ~~Implement proper array construction for Format args~~ ✅ DONE (temporary ConstType workaround)
  - [x] ~~Create `testdata/string_functions/format.dws` with Format examples~~ ✅ EXISTS (82 lines, comprehensive)
  - [x] ~~Fix semantic analyzer to allow heterogeneous arrays for Format:~~ ✅ DONE
    - [x] ~~Allow empty array literals `[]` in Format context~~ ✅ DONE
    - [x] ~~Allow mixed-type arrays `['string', 123, 3.14]`~~ ✅ DONE
    - [x] ~~Fix parser heuristic to not misinterpret `[varName]` as set literal~~ ✅ DONE
  - [x] ~~Test %s (string), %d (integer), %f (float) specifiers~~ ✅ DONE (in format.dws)
  - [x] ~~Test width and precision: %5d, %.2f, %8.2f~~ ✅ DONE (in format.dws)
  - [x] ~~Test %% (literal percent)~~ ✅ DONE (in format.dws)
  - [x] ~~Test multiple arguments~~ ✅ DONE (in format.dws)
  - [x] ~~Create expected output file `testdata/string_functions/format.out`~~ ✅ DONE (35 lines)
  - [x] ~~Add CLI integration tests for Format in `cmd/dwscript/string_functions_test.go`~~ ✅ DONE
  - [x] ~~Document Format syntax in `docs/builtins.md`~~ ✅ DONE (lines 8-118)

**Implementation Summary**:
- Added `ConstType` to type system (`internal/types/types.go` lines 112-124)
- Added `CONST` singleton and `ARRAY_OF_CONST` type (lines 140, 145)
- Modified semantic analyzer to:
  - Pass `ARRAY_OF_CONST` as expected type for Format's second argument (`analyze_expressions.go:681`)
  - Allow heterogeneous elements when element type is `CONST` (`analyze_arrays.go:213-218`)
  - Convert `SetLiteral` to `ArrayLiteral` when expected type is array (`analyze_expressions.go:89-111`)
  - Set type annotation on `SetLiteral` so interpreter knows to treat it as array
- Added interpreter support:
  - `Const` type resolution in `resolveType()` (`record.go:167-169`, `helpers.go:491-493`)
  - Allow any value type when array element type is `CONST` (`array.go:294-299`)
  - Check `SetLiteral.Type` annotation and evaluate as array if needed (`set.go:21-36`)
- Parser kept unchanged (still uses `shouldParseAsSetLiteral()` heuristic)
- All Format tests pass (35 lines of expected output)
- All parser tests pass (SetLiteral tests still work correctly)
- CLI integration tests pass

---

#### TODO

- [ ] Review additional helper behaviors (e.g., record helpers, helper precedence rules) against reference tests

---

### 9.12 Interface Type Resolution in Semantic Analyzer (5 tasks)

**Context**: The `TestInterfacesIntegration` test fails with "unknown type 'ISimple'" errors because interface types are not being resolved when used as variable types. The `analyzeInterfaceDecl` function correctly registers interfaces in `a.interfaces`, but the `resolveType` function doesn't check this map.

**Root Cause**: In `internal/semantic/type_resolution.go:17-74`, the `resolveType` function checks classes, enums, records, sets, arrays, type aliases, and subranges, but is missing a lookup for interface types.

**Required**: Add interface type lookup to `resolveType` so that variables can be declared with interface types (e.g., `var simple: ISimple;`).

#### Semantic Analysis (3 tasks)

- [x] 9.207 Add interface type lookup to `resolveType`: ✅ DONE
  - [x] In `internal/semantic/type_resolution.go`, after checking class types (line 38-41)
  - [x] Add interface type check: `if interfaceType, found := a.interfaces[typeName]; found { return interfaceType, nil }`
  - [x] Place it before enum types check to maintain logical grouping (interfaces and classes are similar)
  - [x] Ensure forward references work (interface used before its declaration in same scope)
  - [x] **BONUS**: Added `ClassType.ImplementsInterface()` and `InterfaceType.InheritsFrom()` methods
  - [x] **BONUS**: Updated `canAssign()` to allow class-to-interface assignments when class implements interface
  - [x] **BONUS**: Updated `analyzeMethodCallExpression()` to support interface method calls

- [x] 9.208 Add semantic tests for interface variable declarations: ✅ DONE
  - [x] Test `type IFoo = interface; end; var x: IFoo;` (valid) - `TestInterfaceVariableDeclaration`
  - [x] Test `var x: IUndefined;` (error: unknown interface type) - `TestUndefinedInterfaceType`
  - [x] Test interface variable assignment: `var x: IFoo; x := obj;` where obj implements IFoo - `TestInterfaceVariableAssignment`
  - [x] Test interface type in function parameters: `procedure Test(x: IFoo);` - `TestInterfaceInFunctionParameter`
  - [x] Test interface type in function return: `function Get(): IFoo;` - `TestInterfaceInFunctionReturn`
  - [x] 6 test cases in `interface_analyzer_test.go` (all passing)

- [x] 9.209 Add tests for interface assignment type checking: ✅ DONE
  - [x] Test valid: class implementing interface assigned to interface variable - `TestValidClassToInterfaceAssignment`
  - [x] Test error: class NOT implementing interface assigned to interface variable - `TestInvalidClassToInterfaceAssignment`
  - [x] Test valid: interface variable assigned to another compatible interface variable - `TestValidInterfaceToInterfaceAssignment`
  - [x] Test error: incompatible interface types assigned - `TestIncompatibleInterfaceAssignment`
  - [x] 4 test cases in `interface_analyzer_test.go` (all passing)
  - [x] **BONUS**: Added interface-to-interface assignment support in `canAssign()` function

#### Integration Testing (2 tasks)

- [x] 9.210 Fix and verify `TestInterfacesIntegration`: ✅ DONE
  - [x] Run `go test -v -run TestInterfacesIntegration ./cmd/dwscript`
  - [x] Verify all 15 interface tests in `testdata/interfaces.dws` pass
  - [x] Check that semantic analysis no longer reports "unknown type" errors
  - [x] Ensure interface variables can be declared and assigned
  - [x] Verify interface method calls work through interface references

- [x] 9.211 Add CLI integration test for interfaces: ✅ DONE
  - [x] Create `testdata/interface_variables/interface_vars.dws`
  - [x] Test simple interface variable: `type ITest = interface; function Get(): Integer; end;`
  - [x] Test polymorphism: different classes assigned to same interface variable
  - [x] Test interface inheritance: derived interface variable accepts base interface methods
  - [x] Expected output files for validation
  - [x] Added to `TestInterfacesIntegration` in `cmd/dwscript/interface_cli_test.go`

---

### Lazy Parameter Passing (MEDIUM-HIGH PRIORITY)

**Summary**: Implement the `lazy` parameter modifier to support deferred expression evaluation. Lazy parameters pass unevaluated expressions that are re-evaluated each time they're accessed within the function body, enabling patterns like Jensen's Device, conditional evaluation, and lightweight anonymous functions.

**Example (Jensen's Device)**:
```pascal
function sum(var i : Integer; lo, hi : Integer; lazy term : Float) : Float;
begin
   i:=lo;
   while i<=hi do begin
      Result += term;  // term is re-evaluated each iteration
      Inc(i);
   end;
end;

var i : Integer;
PrintLn(sum(i, 1, 100, 1.0/i));  // Computes harmonic series: 1/1 + 1/2 + ... + 1/100
```

**Use Cases**:
- Jensen's Device and similar mathematical patterns
- Conditional logging (only evaluate expensive expressions when needed)
- Ternary-like operators without compiler magic
- Deferred evaluation for performance optimization

**Reference**: https://www.delphitools.info/2010/12/10/lazy-parameters-if-compound-assignments/

#### AST & Parser (3 tasks)

- [ ] 9.212 Extend parameter declaration AST in `ast/statements.go`:
  - [ ] Add `IsLazy bool` field to `ParameterDecl` struct
  - [ ] Update `String()` method to display `lazy` modifier
  - [ ] Document that lazy parameters capture expressions, not values

- [ ] 9.213 Add parser support for lazy parameters in `parser/statements.go`:
  - [ ] Recognize `lazy` keyword in parameter lists (before parameter name)
  - [ ] Set `IsLazy = true` on parameter declaration
  - [ ] Parse syntax: `procedure Foo(lazy x: Integer)` or `function Bar(lazy s: String)`
  - [ ] Support mixing lazy, var, const, and regular parameters: `func(var x: Integer; lazy y: String; z: Float)`
  - [ ] Add parser tests for lazy parameter syntax

- [ ] 9.214 Add parser tests in `parser/functions_test.go`:
  - [ ] Test `function Test(lazy x: Integer): Integer;` - basic lazy parameter
  - [ ] Test `procedure Log(level: Integer; lazy msg: String);` - mixed parameters
  - [ ] Test multiple lazy parameters: `function If(cond: Boolean; lazy trueVal, falseVal: Integer): Integer;`
  - [ ] Test error: `lazy var x: Integer` - lazy and var are mutually exclusive
  - [ ] Test error: `lazy const x: Integer` - lazy and const are mutually exclusive

#### Semantic Analysis (4 tasks)

- [ ] 9.215 Implement lazy parameter validation in `semantic/analyzer.go`:
  - [ ] Validate that lazy parameters are not also `var` or `const` (mutually exclusive modifiers)
  - [ ] Track lazy parameters in function signature (`FunctionType.Parameters` metadata)
  - [ ] For lazy parameters, defer type checking to call site (expression type, not value type)
  - [ ] Add to `analyzeFunction()` and `analyzeProcedure()`

- [ ] 9.216 Implement lazy parameter call-site validation:
  - [ ] In `analyzeFunctionCall()` and `analyzeProcedureCall()`, identify lazy parameter positions
  - [ ] For lazy arguments, store the AST expression node without evaluating it
  - [ ] Type-check the expression against the parameter type (e.g., `lazy x: Integer` accepts any Integer expression)
  - [ ] Validate that lazy expressions don't have side effects that would be surprising when re-evaluated (warning only)

- [ ] 9.217 Extend function signature storage:
  - [ ] Update `FunctionType` in `types/types.go` to include `LazyParams []bool` (parallel to Parameters)
  - [ ] Store lazy parameter flags during semantic analysis
  - [ ] Use this metadata during call validation and interpretation

- [ ] 9.218 Add semantic tests in `semantic/lazy_params_test.go`:
  - [ ] Test valid: `function Test(lazy x: Integer): Integer;` declaration
  - [ ] Test error: `procedure Foo(lazy var x: Integer);` - lazy + var conflict
  - [ ] Test error: `procedure Bar(lazy const x: String);` - lazy + const conflict
  - [ ] Test valid: call with lazy expression argument
  - [ ] Test type checking: `lazy x: Integer` must receive Integer expression, not String
  - [ ] Test nested calls with lazy parameters

#### Interpreter Support (5 tasks)

- [ ] 9.219 Create thunk/closure representation in `interp/lazy_params.go`:
  - [ ] Define `LazyThunk` struct with:
    - `Expression ast.Expression` - the unevaluated AST node
    - `CapturedEnv *Environment` - variable context from call site
    - `Interpreter *Interpreter` - reference for evaluation
  - [ ] Implement `Evaluate() interface{}` method that evaluates the expression in captured environment
  - [ ] Ensure each call to `Evaluate()` re-evaluates (no caching)

- [ ] 9.220 Implement lazy parameter binding in function calls:
  - [ ] In `evalFunctionCall()` and `evalProcedureCall()`, check if parameter is lazy
  - [ ] For lazy parameters:
    - Create `LazyThunk` with argument expression and current environment
    - Bind thunk to parameter name in function's environment
  - [ ] For non-lazy parameters: evaluate expression immediately (existing behavior)

- [ ] 9.221 Implement lazy parameter dereferencing:
  - [ ] In `evalIdentifier()`, check if identifier resolves to a `LazyThunk`
  - [ ] If so, call `thunk.Evaluate()` to get the current value
  - [ ] Each access re-evaluates the expression (critical for Jensen's Device)
  - [ ] Ensure variable mutations in the expression are visible (e.g., `1.0/i` sees updated `i`)

- [ ] 9.222 Handle lazy parameters in nested scopes:
  - [ ] Ensure thunk evaluation uses the captured environment, not the function's local environment
  - [ ] Test that lazy expressions can reference variables from multiple scopes
  - [ ] Verify that mutations to captured variables are visible in subsequent evaluations

- [ ] 9.223 Add interpreter tests in `interp/lazy_params_test.go`:
  - [ ] Test basic lazy evaluation: expression evaluated on each access
  - [ ] Test Jensen's Device: `sum(i, 1, 100, 1.0/i)` computes harmonic series
  - [ ] Test conditional evaluation: lazy parameter not accessed if condition is false
  - [ ] Test multiple evaluations: lazy param accessed 3 times yields 3 evaluations
  - [ ] Test lazy expression with side effects: `lazy (Inc(i))` increments on each access
  - [ ] Test lazy parameter with captured variables: mutations are visible
  - [ ] Test lazy expression referencing loop variable: sees updated value
  - [ ] Test error: lazy parameter type mismatch caught at semantic analysis

#### Testing & Integration (3 tasks)

- [ ] 9.224 Create test scripts in `testdata/lazy_params/`:
  - [ ] `jensens_device.dws` - Classic Jensen's Device (harmonic series)
  - [ ] `conditional_eval.dws` - Ternary-like function: `function If(cond: Boolean; lazy t, f: Integer): Integer;`
  - [ ] `lazy_logging.dws` - Logger that only evaluates message if log level is enabled
  - [ ] `multiple_access.dws` - Lazy param accessed multiple times with different side effects
  - [ ] `lazy_with_loops.dws` - Lazy expressions in loop contexts
  - [ ] Expected output files for each

- [ ] 9.225 Add CLI integration tests in `cmd/dwscript/lazy_params_test.go`:
  - [ ] `TestLazyParamsScriptsExist` - verify all test scripts exist
  - [ ] `TestLazyParamsParsing` - validate parsing of all lazy parameter scripts
  - [ ] `TestLazyParamsExecution` - validate output matches expected for all scripts
  - [ ] `TestJensensDevice` - specific test for the canonical example
  - [ ] `TestLazyEvaluationCount` - verify re-evaluation on each access
  - [ ] `TestLazyConditional` - verify lazy expressions not evaluated when skipped

- [ ] 9.226 Document lazy parameters in `docs/lazy_parameters.md`:
  - [ ] Syntax reference: `procedure Foo(lazy x: Type);`
  - [ ] Semantics: expression passed, not value; re-evaluated on each access
  - [ ] Use cases: Jensen's Device, conditional evaluation, logging
  - [ ] Examples from test fixtures
  - [ ] Limitations:
    - Experimental feature (as noted in DWScript blog)
    - Lazy expressions with side effects can be confusing
    - Performance implications of re-evaluation
    - Cannot mix `lazy` with `var` or `const`
  - [ ] Implementation details: thunks, environment capture, re-evaluation

---

### For Loop Step Keyword (HIGH PRIORITY)

**Summary**: Implement the optional `step` keyword for `for-to` and `for-downto` loops to support custom increment/decrement values. Currently, for loops only increment/decrement by 1.

**Example**: `for i := 1 to 10 step 2 do PrintLn(i);` outputs 1, 3, 5, 7, 9

**Reference**: `examples/rosetta/Lucas-Lehmer_test.dws` uses `for p:=3 to upperBound step 2 do`

**Syntax**:

```pascal
for <var> := <start> to <end> [step <step_expr>] do <statement>
for <var> := <start> downto <end> [step <step_expr>] do <statement>
```

**Semantics**:

- Step is optional; defaults to 1 if not specified
- Step expression must evaluate to an integer
- Step value must be strictly positive (> 0) at runtime
- Step is evaluated once before loop execution
- For ascending loops (`to`): increment by step
- For descending loops (`downto`): decrement by step
- Runtime error if step ≤ 0: "FOR loop STEP should be strictly positive: \<value\>"

#### Lexer Support (1 task)

- [ ] 9.227 Add `STEP` token to lexer:
  - [ ] In `internal/lexer/token_type.go`, add `STEP` constant in control flow section (after `DOWNTO`)
  - [ ] Add to token string mapping: `STEP: "STEP"`
  - [ ] In `internal/lexer/token.go`, add to keywords map: `"step": STEP`
  - [ ] Test that `step` is recognized as keyword (case-insensitive)

#### AST Changes (1 task)

- [ ] 9.228 Extend `ForStatement` AST node in `internal/ast/control_flow.go`:
  - [ ] Add `Step Expression` field to `ForStatement` struct (nil if no step specified)
  - [ ] Update `String()` method to include step when present: `for i := 1 to 10 step 2 do ...`
  - [ ] Add AST tests for for loops with step expressions

#### Parser Support (2 tasks)

- [ ] 9.229 Parse optional `step` keyword in `internal/parser/control_flow.go`:
  - [ ] In `parseForStatement()`, after parsing the end expression, check for `STEP` token
  - [ ] If `STEP` found, parse the step expression: `p.nextToken(); stmt.Step = p.parseExpression(LOWEST)`
  - [ ] Validate step expression is not nil
  - [ ] Ensure `DO` token follows (either immediately or after step expression)
  - [ ] Handle error cases: missing expression after `step`, invalid tokens

- [ ] 9.230 Add parser tests in `internal/parser/control_flow_test.go`:
  - [ ] Test `for i := 1 to 10 step 2 do PrintLn(i);` - basic ascending with step
  - [ ] Test `for i := 10 downto 1 step 3 do PrintLn(i);` - basic descending with step
  - [ ] Test `for i := 1 to 100 step (x * 2) do ...;` - step with expression
  - [ ] Test `for var i := 0 to 20 step 5 do ...;` - inline var with step
  - [ ] Test error: `for i := 1 to 10 step do` - missing step expression
  - [ ] Test without step still works: `for i := 1 to 10 do ...;` - backward compatibility

#### Semantic Analysis (2 tasks)

- [ ] 9.231 Implement step expression type checking in `internal/semantic/analyzer.go`:
  - [ ] In `analyzeForStatement()`, if step expression exists, analyze it
  - [ ] Validate step expression type is Integer
  - [ ] Add semantic error if step type is not Integer: "for loop step must be Integer, got \<type\>"
  - [ ] If step is a constant expression and ≤ 0, emit compile-time error (optional optimization)

- [ ] 9.232 Add semantic tests in `internal/semantic/control_flow_test.go`:
  - [ ] Test valid: `for i := 1 to 10 step 2 do ...;` - Integer step
  - [ ] Test valid: `for i := 1 to 10 step (x + 1) do ...;` - Integer expression step
  - [ ] Test error: `for i := 1 to 10 step 2.5 do ...;` - Float step (type error)
  - [ ] Test error: `for i := 1 to 10 step "text" do ...;` - String step (type error)
  - [ ] Test valid: step variable references work correctly

#### Interpreter Support (3 tasks)

- [ ] 9.233 Evaluate step expression in `internal/interp/statements.go`:
  - [ ] In `evalForStatement()`, after evaluating start/end, check if `stmt.Step != nil`
  - [ ] If step exists, evaluate: `stepVal := i.Eval(stmt.Step)`
  - [ ] Extract integer value from step result
  - [ ] Validate step value is strictly positive (> 0): if not, return error
  - [ ] Default to `stepValue = 1` if no step specified
  - [ ] Error message: "FOR loop STEP should be strictly positive: \<value\>"

- [ ] 9.234 Use step value in loop execution:
  - [ ] For ascending loops (`ForTo`): change `current++` to `current += stepValue`
  - [ ] For descending loops (`ForDownto`): change `current--` to `current -= stepValue`
  - [ ] Ensure step value is evaluated only once (before loop starts)
  - [ ] Preserve all existing loop behavior (break, continue, inline vars, etc.)

- [ ] 9.235 Add interpreter tests in `internal/interp/control_flow_test.go`:
  - [ ] Test ascending with step 2: `for i := 1 to 5 step 2 do` outputs 1, 3, 5
  - [ ] Test descending with step 3: `for i := 10 downto 1 step 3 do` outputs 10, 7, 4, 1
  - [ ] Test step expression: `for i := 0 to 10 step (2 + 1) do` outputs 0, 3, 6, 9
  - [ ] Test runtime error: `var s := -1; for i := 1 to 5 step s do` - negative step error
  - [ ] Test runtime error: `for i := 1 to 5 step 0 do` - zero step error
  - [ ] Test large step: `for i := 1 to 100 step 50 do` outputs 1, 51
  - [ ] Test step larger than range: `for i := 1 to 3 step 10 do` outputs 1 only

#### Testing & Integration (2 tasks)

- [ ] 9.236 Create test scripts in `testdata/for_step/`:
  - [ ] `basic_step.dws` - Simple ascending/descending with step
  - [ ] `step_expressions.dws` - Step with variable expressions
  - [ ] `step_validation.dws` - Runtime errors for invalid steps
  - [ ] `lucas_lehmer.dws` - Simplified Lucas-Lehmer test using step
  - [ ] Expected output files for each

- [ ] 9.237 Add CLI integration tests and verify Rosetta Code example:
  - [ ] Create `cmd/dwscript/for_step_test.go` with integration tests
  - [ ] Test all fixtures in `testdata/for_step/`
  - [ ] Verify `examples/rosetta/Lucas-Lehmer_test.dws` now parses and executes correctly
  - [ ] Expected output: "Finding Mersenne primes in M[2..30]: M2 M3 M5 M7 M13 M17 M19"

---

## Phase 9 Summary

**Total Tasks**: ~255 tasks
**Estimated Effort**: ~30 weeks (~7.5 months)

### Priority Breakdown:

**HIGH PRIORITY** (~173 tasks, ~20 weeks):

- Subrange Types: 12 tasks
- Units/Modules System: 45 tasks (CRITICAL for multi-file projects)
- Function/Method Pointers: 25 tasks
- External Function Registration (FFI): 35 tasks
- Array Instantiation (`new TypeName[size]`): 12 tasks (CRITICAL for Rosetta Code examples)
- For Loop Step Keyword: 11 tasks (REQUIRED for Lucas-Lehmer test and other Rosetta Code examples)

**MEDIUM-HIGH PRIORITY** (~15 tasks, ~2 weeks):

- Lazy Parameter Passing: 15 tasks (REQUIRED for Jensen's Device and Rosetta Code examples)

**MEDIUM PRIORITY** (~67 tasks, ~8 weeks):

- Lambdas/Anonymous Methods: 30 tasks (depends on function pointers)
- Helpers: 20 tasks
- DateTime Functions: 24 tasks
- JSON Support: 18 tasks
- Improved Error Messages: 12 tasks

This comprehensive backlog brings go-dws from ~55% to ~85% feature parity with DWScript, making it production-ready for most use cases. The for loop step keyword, array instantiation, and lazy parameter features are particularly critical as they unblock numerous real-world DWScript examples (e.g., Rosetta Code algorithms like Lucas-Lehmer test and Jensen's Device) that rely on custom loop increments, dynamic array creation, and deferred expression evaluation.

## Stage 10: Performance Tuning and Refactoring

### Performance Profiling

- [x] 10.1 Create performance benchmark scripts
- [x] 10.2 Profile lexer performance: `BenchmarkLexer`
- [x] 10.3 Profile parser performance: `BenchmarkParser`
- [x] 10.4 Profile interpreter performance: `BenchmarkInterpreter`
- [x] 10.5 Identify bottlenecks using `pprof`
- [ ] 10.6 Document performance baseline

### Optimization - Lexer

- [ ] 10.7 Optimize string handling in lexer (use bytes instead of runes where possible)
- [ ] 10.8 Reduce allocations in token creation
- [ ] 10.9 Use string interning for keywords/identifiers
- [ ] 10.10 Benchmark improvements

### Optimization - Parser

- [ ] 10.11 Reduce AST node allocations
- [ ] 10.12 Pool commonly created nodes
- [ ] 10.13 Optimize precedence table lookups
- [ ] 10.14 Benchmark improvements

### Bytecode Compiler (Optional)

- [ ] 10.15 Design bytecode instruction set:
  - [ ] Load constant
  - [ ] Load/store variable
  - [ ] Binary/unary operations
  - [ ] Jump instructions (conditional/unconditional)
  - [ ] Call/return
  - [ ] Object operations
- [ ] 10.16 Implement bytecode emitter (AST → bytecode)
- [ ] 10.17 Implement bytecode VM (execute instructions)
- [ ] 10.18 Handle stack management in VM
- [ ] 10.19 Test bytecode execution produces same results as AST interpreter
- [ ] 10.20 Benchmark bytecode VM vs AST interpreter
- [ ] 10.21 Optimize VM loop
- [ ] 10.22 Add option to CLI to use bytecode or AST interpreter

### Optimization - Interpreter

- [ ] 10.23 Optimize value representation (avoid interface{} overhead if possible)
- [ ] 10.24 Use switch statements instead of type assertions where possible
- [ ] 10.25 Cache frequently accessed symbols
- [ ] 10.26 Optimize environment lookups
- [ ] 10.27 Reduce allocations in hot paths
- [ ] 10.28 Benchmark improvements

### Memory Management

- [ ] 10.29 Ensure no memory leaks in long-running scripts
- [ ] 10.30 Profile memory usage with large programs
- [ ] 10.31 Optimize object allocation/deallocation
- [ ] 10.32 Consider object pooling for common types

### Code Quality Refactoring

- [ ] 10.33 Run `go vet ./...` and fix all issues
- [ ] 10.34 Run `golangci-lint run` and address warnings
- [ ] 10.35 Run `gofmt` on all files
- [ ] 10.36 Run `goimports` to organize imports
- [ ] 10.37 Review error handling consistency
- [ ] 10.38 Unify value representation if inconsistent
- [ ] 10.39 Refactor large functions into smaller ones
- [ ] 10.40 Extract common patterns into helper functions
- [ ] 10.41 Improve variable/function naming
- [ ] 10.42 Add missing error checks

### Documentation

- [ ] 10.43 Write comprehensive GoDoc comments for all exported types/functions
- [ ] 10.44 Document internal architecture in `docs/architecture.md`
- [ ] 10.45 Create user guide in `docs/user_guide.md`
- [ ] 10.46 Document CLI usage with examples
- [ ] 10.47 Create API documentation for embedding the library
- [ ] 10.48 Add code examples to documentation
- [ ] 10.49 Document known limitations
- [ ] 10.50 Create contribution guidelines in `CONTRIBUTING.md`

### Example Programs

- [x] 10.51 Create `examples/` directory
- [x] 10.52 Add example scripts:
  - [x] Hello World
  - [x] Fibonacci
  - [x] Factorial
  - [x] Class-based example (Person demo)
  - [x] Algorithm sample (math/loops showcase)
- [x] 10.53 Add README in examples directory
- [x] 10.54 Ensure all examples run correctly

### Testing Enhancements

- [ ] 10.55 Add integration tests in `test/integration/`
- [ ] 10.56 Add fuzzing tests for parser: `FuzzParser`
- [ ] 10.57 Add fuzzing tests for lexer: `FuzzLexer`
- [ ] 10.58 Add property-based tests (using testing/quick or gopter)
- [ ] 10.59 Ensure CI runs all test types
- [ ] 10.60 Achieve >90% code coverage overall
- [ ] 10.61 Add regression tests for all fixed bugs

### Release Preparation

- [ ] 10.62 Create `CHANGELOG.md`
- [ ] 10.63 Document version numbering scheme (SemVer)
- [ ] 10.64 Tag v0.1.0 alpha release
- [ ] 10.65 Create release binaries for major platforms (Linux, macOS, Windows)
- [ ] 10.66 Publish release on GitHub
- [ ] 10.67 Write announcement blog post or README update
- [ ] 10.68 Share with community for feedback

---

## Stage 11: Long-Term Evolution

### Feature Parity Tracking

- [ ] 11.1 Create feature matrix comparing go-dws with DWScript
- [ ] 11.2 Track DWScript upstream releases
- [ ] 11.3 Identify new features in DWScript updates
- [ ] 11.4 Plan integration of new features
- [ ] 11.5 Update feature matrix regularly

### Community Building

- [ ] 11.6 Set up issue templates on GitHub
- [ ] 11.7 Set up pull request template
- [ ] 11.8 Create CODE_OF_CONDUCT.md
- [ ] 11.9 Create discussions forum or mailing list
- [ ] 11.10 Encourage contributions (tag "good first issue")
- [ ] 11.11 Respond to issues and PRs promptly
- [ ] 11.12 Build maintainer team (if interest grows)

### Advanced Features

- [ ] 11.13 Implement REPL (Read-Eval-Print Loop):
  - [ ] Interactive prompt
  - [ ] Statement-by-statement execution
  - [ ] Variable inspection
  - [ ] History and autocomplete
- [ ] 11.14 Implement debugging support:
  - [ ] Breakpoints
  - [ ] Step-through execution
  - [ ] Variable inspection
  - [ ] Stack traces
- [ ] 11.15 Implement WebAssembly compilation (see `docs/plans/2025-10-26-wasm-compilation-design.md`):
  - [x] 11.15.1 Platform Abstraction Layer (completed 2025-10-26):
    - [x] Create `pkg/platform/` package with core interfaces (FileSystem, Console, Platform)
    - [x] Implement `pkg/platform/native/` for standard Go implementations
    - [x] Implement `pkg/platform/wasm/` with virtual filesystem (in-memory map)
    - [x] Add console bridge to JavaScript console.log or callbacks (implemented with test stubs)
    - [x] Implement time functions using JavaScript Date API via syscall/js (implemented with stubs for future WASM runtime)
    - [x] Add sleep implementation using setTimeout with Promise/channel bridge (implemented with time.Sleep stub)
    - [ ] Create feature parity test suite (runs on both native and WASM)
    - [ ] Document platform differences and limitations
  - [x] 11.15.2 WASM Build Infrastructure (completed 2025-10-26):
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
  - [x] 11.15.3 JavaScript/Go Bridge (completed 2025-10-26):
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
  - [x] 11.15.4 Web Playground (completed 2025-10-26):
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
  - [ ] 11.15.5 NPM Package:
    - [x] Create `npm/` package structure with package.json
    - [x] Write TypeScript definitions in `typescript/index.d.ts`
    - [x] Create dual ESM/CommonJS entry points (index.js, index.cjs)
    - [x] Add WASM loader helper for both Node.js and browser
    - [x] Create usage examples (Node.js, React, Vue, vanilla JS)
    - [x] Set up automated NPM publishing via GitHub Actions
    - [x] Configure package for tree-shaking and optimal bundling
    - [x] Write `npm/README.md` with installation and usage guide
    - [ ] Publish initial version to npmjs.com registry
  - [ ] 11.15.6 Testing & Documentation:
    - [ ] Write WASM-specific unit tests (GOOS=js GOARCH=wasm go test)
    - [ ] Create Node.js integration test suite using test runner
    - [ ] Add Playwright browser tests for cross-browser compatibility
    - [ ] Set up CI matrix for Chrome, Firefox, and Safari testing
    - [ ] Add performance benchmarks comparing WASM vs native speed
    - [ ] Implement bundle size regression monitoring in CI
    - [ ] Write `docs/wasm/EMBEDDING.md` for web app integration guide
    - [ ] Update main README.md with WASM section and playground link
- [ ] 11.16 Implement language server protocol (LSP):
  - [ ] Syntax highlighting
  - [ ] Autocomplete
  - [ ] Go-to-definition
  - [ ] Error diagnostics in IDE
- [ ] 11.17 Implement JavaScript code generation backend:
  - [ ] AST → JavaScript transpiler
  - [ ] Support browser execution
  - [ ] Create npm package

### Alternative Execution Modes

- [ ] 11.18 Add JIT compilation (if feasible in Go)
- [ ] 11.19 Add AOT compilation (compile to native binary)
- [ ] 11.20 Add compilation to Go source code
- [ ] 11.21 Benchmark different execution modes

### Platform-Specific Enhancements

- [ ] 11.22 Add Windows-specific features (if needed)
- [ ] 11.23 Add macOS-specific features (if needed)
- [ ] 11.24 Add Linux-specific features (if needed)
- [ ] 11.25 Test on multiple architectures (ARM, AMD64)

### Edge Case Audit

- [ ] 11.26 Test short-circuit evaluation (and, or)
- [ ] 11.27 Test operator precedence edge cases
- [ ] 11.28 Test division by zero handling
- [ ] 11.29 Test integer overflow behavior
- [ ] 11.30 Test floating-point edge cases (NaN, Inf)
- [ ] 11.31 Test string encoding (UTF-8 handling)
- [ ] 11.32 Test very large programs (scalability)
- [ ] 11.33 Test deeply nested structures
- [ ] 11.34 Test circular references (if possible in language)
- [ ] 11.35 Fix any discovered issues

### Performance Monitoring

- [ ] 11.36 Set up continuous performance benchmarking
- [ ] 11.37 Track performance metrics over releases
- [ ] 11.38 Identify and fix performance regressions
- [ ] 11.39 Publish performance comparison with DWScript

### Security Audit

- [ ] 11.40 Review for potential security issues (untrusted script execution)
- [ ] 11.41 Implement resource limits (memory, execution time)
- [ ] 11.42 Implement sandboxing for untrusted scripts
- [ ] 11.43 Audit for code injection vulnerabilities
- [ ] 11.44 Document security best practices

### Maintenance

- [ ] 11.45 Keep dependencies up to date
- [ ] 11.46 Monitor Go version updates and migrate as needed
- [ ] 11.47 Maintain CI/CD pipeline
- [ ] 11.48 Regular code reviews
- [ ] 11.49 Address technical debt periodically

### Long-term Roadmap

- [ ] 11.50 Define 1-year roadmap
- [ ] 11.51 Define 3-year roadmap
- [ ] 11.52 Gather user feedback and adjust priorities
- [ ] 11.53 Consider commercial applications/support
- [ ] 11.54 Explore academic/research collaborations

---

## Stage 12: Code Generation - Multi-Backend Architecture

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

#### 12.1.1: MIR Package Structure and Types (10 tasks)

- [ ] 12.1 Create `mir/` package directory
- [ ] 12.2 Create `mir/types.go` - MIR type system
- [ ] 12.3 Define `Type` interface with `String()`, `Size()`, `Align()` methods
- [ ] 12.4 Implement primitive types: `Bool`, `Int8`, `Int16`, `Int32`, `Int64`, `Float32`, `Float64`, `String`
- [ ] 12.5 Implement composite types: `Array(elemType, size)`, `Record(fields)`, `Pointer(pointeeType)`
- [ ] 12.6 Implement OOP types: `Class(name, fields, methods, parent)`, `Interface(name, methods)`
- [ ] 12.7 Implement function types: `Function(params, returnType)`
- [ ] 12.8 Add `Void` type for procedures
- [ ] 12.9 Implement type equality and compatibility checking
- [ ] 12.10 Implement type conversion rules (explicit vs implicit)

#### 12.1.2: MIR Instructions and Control Flow (10 tasks)

- [ ] 12.11 Create `mir/instruction.go` - MIR instruction set
- [ ] 12.12 Define `Instruction` interface with `ID()`, `Type()`, `String()` methods
- [ ] 12.13 Implement arithmetic ops: `Add`, `Sub`, `Mul`, `Div`, `Mod`, `Neg`
- [ ] 12.14 Implement comparison ops: `Eq`, `Ne`, `Lt`, `Le`, `Gt`, `Ge`
- [ ] 12.15 Implement logical ops: `And`, `Or`, `Xor`, `Not`
- [ ] 12.16 Implement memory ops: `Alloca`, `Load`, `Store`
- [ ] 12.17 Implement constants: `ConstInt`, `ConstFloat`, `ConstString`, `ConstBool`, `ConstNil`
- [ ] 12.18 Implement conversions: `IntToFloat`, `FloatToInt`, `IntTrunc`, `IntExt`
- [ ] 12.19 Implement function ops: `Call`, `VirtualCall`
- [ ] 12.20 Implement array/class ops: `ArrayAlloc`, `ArrayLen`, `ArrayIndex`, `ArraySet`, `FieldGet`, `FieldSet`, `New`

#### 12.1.3: MIR Control Flow Structures (5 tasks)

- [ ] 12.21 Create `mir/block.go` - Basic blocks with `ID`, `Instructions`, `Terminator`
- [ ] 12.22 Implement control flow terminators: `Phi`, `Br`, `CondBr`, `Return`, `Throw`
- [ ] 12.23 Implement terminator validation (every block must end with terminator)
- [ ] 12.24 Implement block predecessors/successors tracking for CFG
- [ ] 12.25 Create `mir/function.go` - Function representation with `Name`, `Params`, `ReturnType`, `Blocks`, `Locals`

#### 12.1.4: MIR Builder API (3 tasks)

- [ ] 12.26 Create `mir/builder.go` - Safe MIR construction
- [ ] 12.27 Implement `Builder` struct with function/block context, `NewFunction()`, `NewBlock()`, `SetInsertPoint()`
- [ ] 12.28 Implement instruction emission methods: `EmitAdd()`, `EmitLoad()`, `EmitStore()`, etc. with type checking

#### 12.1.5: MIR Verifier (2 tasks)

- [ ] 12.29 Create `mir/verifier.go` - MIR correctness checking
- [ ] 12.30 Implement CFG, type, SSA, and function signature verification with `Verify(fn *Function) []error` API

### Stage 12.2: AST → MIR Lowering (12 tasks)

- [ ] 12.31 Create `mir/lower.go` - AST to MIR translation
- [ ] 12.32 Implement `LowerProgram(ast *ast.Program) (*mir.Module, error)` entry point
- [ ] 12.33 Lower expressions: literals → `Const*` instructions
- [ ] 12.34 Lower binary operations → corresponding MIR ops (handle short-circuit for `and`/`or`)
- [ ] 12.35 Lower unary operations → `Neg`, `Not`
- [ ] 12.36 Lower identifier references → `Load` instructions
- [ ] 12.37 Lower function calls → `Call` instructions
- [ ] 12.38 Lower array indexing → `ArrayIndex` + bounds check insertion
- [ ] 12.39 Lower record field access → `FieldGet`/`FieldSet`
- [ ] 12.40 Lower statements: variable declarations, assignments, if/while/for, return
- [ ] 12.41 Lower declarations: functions/procedures, records, classes
- [ ] 12.42 Implement short-circuit evaluation and simple optimizations (constant folding, dead code elimination)

### Stage 12.3: MIR Debugging and Testing (5 tasks)

- [ ] 12.43 Create `mir/dump.go` - Human-readable MIR output with `Dump(fn *Function) string`
- [ ] 12.44 Integration with CLI: `./bin/dwscript dump-mir script.dws`
- [ ] 12.45 Create golden MIR tests: 5+ each for expressions, control flow, functions, advanced features
- [ ] 12.46 Implement MIR verifier tests: type mismatches, malformed CFG, SSA violations
- [ ] 12.47 Implement round-trip tests: AST → MIR → verify → dump → compare with golden files

### Stage 12.4: JS Backend MVP (45 tasks)

**Goal**: Implement a JavaScript code generator that can compile basic DWScript programs to readable, runnable JavaScript.

**Exit Criteria**: JS emitter for expressions/control flow/functions, 20+ end-to-end tests (DWScript→JS→execute), golden JS snapshots, 85%+ coverage

#### 12.4.1: JS Emitter Infrastructure (8 tasks)

- [ ] 12.48 Create `codegen/` package with `Backend` interface and `EmitterOptions`
- [ ] 12.49 Create `codegen/js/` package and `emitter.go`
- [ ] 12.50 Define `JSEmitter` struct with `out`, `indent`, `opts`, `tmpCounter`
- [ ] 12.51 Implement helper methods: `emit()`, `emitLine()`, `emitIndent()`, `pushIndent()`, `popIndent()`
- [ ] 12.52 Implement `newTemp()` for temporary variable naming
- [ ] 12.53 Implement `NewJSEmitter(opts EmitterOptions)`
- [ ] 12.54 Implement `Generate(module *mir.Module) (string, error)` entry point
- [ ] 12.55 Test emitter infrastructure

#### 12.4.2: Module and Function Emission (6 tasks)

- [ ] 12.56 Implement module structure emission: ES Module format with `export`, file header comment
- [ ] 12.57 Implement optional IIFE fallback via `EmitterOptions`
- [ ] 12.58 Implement function emission: `function fname(params) { ... }`
- [ ] 12.59 Map DWScript params to JS params (preserve names)
- [ ] 12.60 Emit local variable declarations at function top (from `Alloca` instructions)
- [ ] 12.61 Handle procedures (no return value) as JS functions

#### 12.4.3: Expression and Instruction Lowering (12 tasks)

- [ ] 12.62 Lower arithmetic operations → JS infix operators: `+`, `-`, `*`, `/`, `%`, unary `-`
- [ ] 12.63 Lower comparison operations → JS comparisons: `===`, `!==`, `<`, `<=`, `>`, `>=`
- [ ] 12.64 Lower logical operations → JS boolean ops: `&&`, `||`, `!`
- [ ] 12.65 Lower constants → JS literals with proper escaping
- [ ] 12.66 Lower variable operations: `Load` → variable reference, `Store` → assignment
- [ ] 12.67 Lower function calls: `Call` → `functionName(args)`
- [ ] 12.68 Implement Phi node lowering with temporary variables at block edges
- [ ] 12.69 Test expression lowering
- [ ] 12.70 Test instruction lowering
- [ ] 12.71 Test temporary variable generation
- [ ] 12.72 Test type conversions
- [ ] 12.73 Test complex expressions

#### 12.4.4: Control Flow Emission (8 tasks)

- [ ] 12.74 Implement control flow reconstruction from MIR CFG
- [ ] 12.75 Detect if/else patterns from `CondBr`
- [ ] 12.76 Detect while loop patterns (backedge to header)
- [ ] 12.77 Emit if-else: `if (condition) { ... } else { ... }`
- [ ] 12.78 Emit while loops: `while (condition) { ... }`
- [ ] 12.79 Emit for loops if MIR preserves metadata
- [ ] 12.80 Handle unconditional branches
- [ ] 12.81 Handle return statements

#### 12.4.5: Runtime and Testing (11 tasks)

- [ ] 12.82 Create `runtime/js/runtime.js` with `_dws.boundsCheck()`, `_dws.assert()`
- [ ] 12.83 Emit runtime import in generated JS (if needed)
- [ ] 12.84 Make runtime usage optional via `EmitterOptions.InsertBoundsChecks`
- [ ] 12.85 Create `codegen/js/testdata/` with subdirectories
- [ ] 12.86 Implement golden JS snapshot tests
- [ ] 12.87 Setup Node.js in CI (GitHub Actions)
- [ ] 12.88 Implement execution tests: parse → lower → generate → execute → verify
- [ ] 12.89 Add end-to-end tests for arithmetic, control flow, functions, loops
- [ ] 12.90 Add unit tests for JS emitter
- [ ] 12.91 Achieve 85%+ coverage for `codegen/js/` package
- [ ] 12.92 Add `compile-js` CLI command: `./bin/dwscript compile-js input.dws -o output.js`

### Stage 12.5: JS Feature Complete (60 tasks)

**Goal**: Extend JS backend to support all DWScript language features.

**Exit Criteria**: Full OOP, composite types, exceptions, properties, 50+ comprehensive tests, real-world samples work

#### 12.5.1: Records (7 tasks)

- [ ] 12.93 Implement MIR support for records
- [ ] 12.94 Emit records as plain JS objects: `{ x: 0, y: 0 }`
- [ ] 12.95 Implement constructor functions for records
- [ ] 12.96 Implement field access/assignment as property access
- [ ] 12.97 Implement record copy semantics with `_dws.copyRecord()`
- [ ] 12.98 Test record creation, initialization, field read/write
- [ ] 12.99 Test nested records and copy semantics

#### 12.5.2: Arrays (10 tasks)

- [ ] 12.100 Extend MIR for static and dynamic arrays
- [ ] 12.101 Emit static arrays as JS arrays with fixed size
- [ ] 12.102 Implement array index access with optional bounds checking
- [ ] 12.103 Emit dynamic arrays as JS arrays
- [ ] 12.104 Implement `SetLength` → `arr.length = newLen`
- [ ] 12.105 Implement `Length` → `arr.length`
- [ ] 12.106 Support multi-dimensional arrays (nested JS arrays)
- [ ] 12.107 Implement array operations: copy, concatenation
- [ ] 12.108 Test static array creation and indexing
- [ ] 12.109 Test dynamic array operations and bounds checking

#### 12.5.3: Classes and Inheritance (15 tasks)

- [ ] 12.110 Extend MIR for classes with fields, methods, parent, vtable
- [ ] 12.111 Emit ES6 class syntax: `class TAnimal { ... }`
- [ ] 12.112 Implement field initialization in constructor
- [ ] 12.113 Implement method emission
- [ ] 12.114 Implement inheritance with `extends` clause
- [ ] 12.115 Implement `super()` call in constructor
- [ ] 12.116 Handle virtual method dispatch (naturally virtual in JS)
- [ ] 12.117 Handle DWScript `Create` → JS `constructor`
- [ ] 12.118 Handle multiple constructors (overload dispatch)
- [ ] 12.119 Document destructor handling (no direct equivalent in JS)
- [ ] 12.120 Implement static fields and methods
- [ ] 12.121 Map `Self` → `this`, `inherited` → `super.method()`
- [ ] 12.122 Test simple classes with fields and methods
- [ ] 12.123 Test inheritance, virtual method overriding, constructors
- [ ] 12.124 Test static members and `Self`/`inherited` usage

#### 12.5.4: Interfaces (6 tasks)

- [ ] 12.125 Extend MIR for interfaces
- [ ] 12.126 Choose and document JS emission strategy (structural typing vs runtime metadata)
- [ ] 12.127 If using runtime metadata: emit interface tables, implement `is`/`as` operators
- [ ] 12.128 Test class implementing interface
- [ ] 12.129 Test interface method calls
- [ ] 12.130 Test `is` and `as` with interfaces

#### 12.5.5: Enums and Sets (8 tasks)

- [ ] 12.131 Extend MIR for enums
- [ ] 12.132 Emit enums as frozen JS objects: `const TColor = Object.freeze({...})`
- [ ] 12.133 Support scoped and unscoped enum access
- [ ] 12.134 Extend MIR for sets
- [ ] 12.135 Emit small sets (≤32 elements) as bitmasks
- [ ] 12.136 Emit large sets as JS `Set` objects
- [ ] 12.137 Implement set operations: union, intersection, difference, inclusion
- [ ] 12.138 Test enum declaration/usage and set operations

#### 12.5.6: Exception Handling (8 tasks)

- [ ] 12.139 Extend MIR for exceptions: `Throw`, `Try`, `Catch`, `Finally`
- [ ] 12.140 Emit `Throw` → `throw new Error()` or custom exception class
- [ ] 12.141 Emit try-except-finally → JS `try/catch/finally`
- [ ] 12.142 Create DWScript exception class → JS `Error` subclass
- [ ] 12.143 Handle `On E: ExceptionType do` with instanceof checks
- [ ] 12.144 Implement re-raise with exception tracking
- [ ] 12.145 Test basic try-except, multiple handlers, try-finally
- [ ] 12.146 Test re-raise and nested exception handling

#### 12.5.7: Properties and Advanced Features (6 tasks)

- [ ] 12.147 Extend MIR for properties with `PropGet`/`PropSet`
- [ ] 12.148 Emit properties as ES6 getters/setters
- [ ] 12.149 Handle indexed properties as methods
- [ ] 12.150 Test read/write properties and indexed properties
- [ ] 12.151 Implement operator overloading (desugar to method calls)
- [ ] 12.152 Implement generics support (monomorphization)

### Stage 12.6: LLVM Backend [OPTIONAL - Future Work] (45 tasks)

**Goal**: Implement LLVM IR backend for native code compilation. This is **deferred** and optional.

**Exit Criteria**: Valid LLVM IR generation, runtime library in C, basic end-to-end tests, documentation

#### 12.6.1: LLVM Infrastructure (8 tasks)

- [ ] 12.153 Choose LLVM binding: `llir/llvm` (pure Go) vs CGo bindings
- [ ] 12.154 Create `codegen/llvm/` package with `emitter.go`, `types.go`, `runtime.go`
- [ ] 12.155 Implement type mapping: DWScript types → LLVM types
- [ ] 12.156 Map Integer → `i32`/`i64`, Float → `double`, Boolean → `i1`
- [ ] 12.157 Map String → struct `{i32 len, i8* data}`
- [ ] 12.158 Map arrays/objects to LLVM structs
- [ ] 12.159 Emit LLVM module with target triple
- [ ] 12.160 Declare external runtime functions

#### 12.6.2: Runtime Library (12 tasks)

- [ ] 12.161 Create `runtime/dws_runtime.h` - C header for runtime API
- [ ] 12.162 Declare string operations: `dws_string_new()`, `dws_string_concat()`, `dws_string_len()`
- [ ] 12.163 Declare array operations: `dws_array_new()`, `dws_array_index()`, `dws_array_len()`
- [ ] 12.164 Declare memory management: `dws_alloc()`, `dws_free()`
- [ ] 12.165 Choose and document memory strategy (Boehm GC vs reference counting)
- [ ] 12.166 Declare object operations: `dws_object_new()`, virtual dispatch helpers
- [ ] 12.167 Declare exception handling: `dws_throw()`, `dws_catch()`
- [ ] 12.168 Declare RTTI: `dws_is_instance()`, `dws_as_instance()`
- [ ] 12.169 Create `runtime/dws_runtime.c` - implement runtime
- [ ] 12.170 Implement all runtime functions
- [ ] 12.171 Create `runtime/Makefile` to build `libdws_runtime.a`
- [ ] 12.172 Add runtime build to CI for Linux/macOS/Windows

#### 12.6.3: LLVM Code Emission (15 tasks)

- [ ] 12.173 Implement LLVM emitter: `Generate(module *mir.Module) (string, error)`
- [ ] 12.174 Emit function declarations with correct signatures
- [ ] 12.175 Emit basic blocks for each MIR block
- [ ] 12.176 Emit arithmetic instructions: `add`, `sub`, `mul`, `sdiv`, `srem`
- [ ] 12.177 Emit comparison instructions: `icmp eq`, `icmp slt`, etc.
- [ ] 12.178 Emit logical instructions: `and`, `or`, `xor`
- [ ] 12.179 Emit memory instructions: `alloca`, `load`, `store`
- [ ] 12.180 Emit call instructions: `call @function_name(args)`
- [ ] 12.181 Emit constants: integers, floats, strings
- [ ] 12.182 Emit control flow: conditional branches, phi nodes
- [ ] 12.183 Emit runtime calls for strings, arrays, objects
- [ ] 12.184 Implement type conversions: `sitofp`, `fptosi`
- [ ] 12.185 Emit struct types for classes and vtables
- [ ] 12.186 Implement virtual method dispatch
- [ ] 12.187 Implement exception handling (simple throw/catch or full LLVM EH)

#### 12.6.4: Linking and Testing (7 tasks)

- [ ] 12.188 Implement compilation pipeline: DWScript → MIR → LLVM IR → object → executable
- [ ] 12.189 Integrate `llc` to compile .ll → .o
- [ ] 12.190 Integrate linker to link object + runtime → executable
- [ ] 12.191 Add `compile-native` CLI command
- [ ] 12.192 Create 10+ end-to-end tests: DWScript → native → execute → verify
- [ ] 12.193 Benchmark JS vs native performance
- [ ] 12.194 Document LLVM backend in `docs/llvm-backend.md`

#### 12.6.5: Documentation (3 tasks)

- [ ] 12.195 Create `docs/codegen-architecture.md` - MIR overview, multi-backend design
- [ ] 12.196 Create `docs/mir-spec.md` - complete MIR reference with examples
- [ ] 12.197 Create `docs/js-backend.md` - DWScript → JavaScript mapping guide

---

## Phase 13: AST-Driven Formatter and Playground Integration 🆕 **PLANNED**

Goal: deliver an auto-formatting pipeline that reuses the existing AST and semantic metadata to produce canonical DWScript source, accessible via the CLI (`dwscript fmt`), editors, and the web playground.

### 13.1 Specification & AST/Data Prep (7 tasks)

- [x] 13.1.1 Capture formatting requirements from upstream DWScript (indent width, begin/end alignment, keyword casing, line-wrapping) and document them in `docs/formatter-style-guide.md`.
- [x] 13.1.2 Audit current AST nodes for source position fidelity and comment/trivia preservation; list any nodes lacking `Pos` / `EndPos`.
- [ ] 13.1.3 Extend the parser/AST to track leading and trailing trivia (single-line, block comments, blank lines) without disturbing semantic passes.
- [ ] 13.1.4 Define a `format.Options` struct (indent size, max line length, newline style) and default profile matching DWScript conventions.
- [ ] 13.1.5 Build a formatting test corpus in `testdata/formatter/{input,expected}` with tricky constructs (nested classes, generics, properties, preprocessor).
- [ ] 13.1.6 Add helper APIs to serialize AST back into token streams (e.g., `ast.FormatNode`, `ast.IterChildren`) to keep formatter logic decoupled from parser internals.
- [ ] 13.1.7 Ensure the semantic/type metadata needed for spacing decisions (e.g., `var` params, attributes) is exposed through lightweight inspector interfaces to avoid circular imports.

### 13.2 Formatter Engine Implementation (10 tasks)

- [ ] 13.2.1 Create `formatter` package with a multi-phase pipeline: AST normalization → layout planning → text emission.
- [ ] 13.2.2 Implement a visitor that emits `format.Node` instructions (indent/dedent, soft break, literal text) for statements and declarations, leveraging AST shape rather than raw tokens.
- [ ] 13.2.3 Handle block constructs (`begin...end`, class bodies, `case` arms) with indentation stacks so nested scopes auto-align.
- [ ] 13.2.4 Add expression formatting that respects operator precedence and inserts parentheses only when required; reuse existing precedence tables.
- [ ] 13.2.5 Support alignment for parameter lists, generics, array types, and property declarations with configurable wrap points.
- [ ] 13.2.6 Preserve user comments: attach leading comments before the owning node, keep inline comments after statements, and maintain blank-line intent (max consecutives configurable).
- [ ] 13.2.7 Implement whitespace normalization rules (single spaces around binary operators, before `do`/`then`, after commas, etc.).
- [ ] 13.2.8 Provide idempotency guarantees by building a golden test that pipes formatted output back through the formatter and asserts stability.
- [ ] 13.2.9 Expose a streaming writer that emits `[]byte`/`io.Writer` output to keep the CLI fast and low-memory.
- [ ] 13.2.10 Benchmark formatting of large fixtures (≥5k LOC) and optimize hot paths (string builder pools, avoiding interface allocations).

### 13.3 Tooling & Playground Integration (7 tasks)

- [ ] 13.3.1 Wire a new CLI command `dwscript fmt` (and `fmt -w`) that runs the formatter over files/directories, mirroring `gofmt` UX.
- [ ] 13.3.2 Update the WASM bridge to expose a `Format(source string) (string, error)` hook exported from Go, reusing the same formatter package.
- [ ] 13.3.3 Modify `playground/js/playground.js` to call the WASM formatter before falling back to Monaco’s default action, enabling deterministic formatting in the browser.
- [ ] 13.3.4 Add formatter support to the VSCode extension / LSP stub (if present) so editors can trigger `textDocument/formatting`.
- [ ] 13.3.5 Ensure the formatter respects partial-range requests (`textDocument/rangeFormatting`) to avoid reformatting entire files when not desired.
- [ ] 13.3.6 Introduce CI checks (`just fmt-check`) that fail when files are not formatted, and document the workflow in `CONTRIBUTING.md`.
- [ ] 13.3.7 Provide sample scripts/snippets (e.g., Git hooks) encouraging contributors to run the formatter.

### 13.4 Validation, UX, and Docs (6 tasks)

- [ ] 13.4.1 Create table-driven unit tests per node type plus integration tests that read `testdata/formatter` fixtures.
- [ ] 13.4.2 Add fuzz/property tests that compare formatter output against itself round-tripped through the parser → formatter pipeline.
- [ ] 13.4.3 Document formatter architecture and extension points in `docs/formatter-architecture.md`.
- [ ] 13.4.4 Update `PLAYGROUND.md`, `README.md`, and release notes to mention the Format button now runs the AST-driven formatter.
- [ ] 13.4.5 Record known limitations (e.g., preprocessor directives) and track follow-ups in `TEST_ISSUES.md`.
- [ ] 13.4.6 Gather usability feedback (issue template or telemetry) to prioritize refinements like configurable styles or multi-profile support.

---

## Summary

This detailed plan breaks down the ambitious goal of porting DWScript from Delphi to Go into **~897 bite-sized tasks** across 12 stages. Each stage builds incrementally:

1. **Stage 1**: Lexer implementation (45 tasks) - ✅ COMPLETE
2. **Stage 2**: Basic parser and AST (64 tasks) - ✅ COMPLETE
3. **Stage 3**: Statement execution (65 tasks) - ✅ COMPLETE (98.5%)
4. **Stage 4**: Control flow (46 tasks) - ✅ COMPLETE
5. **Stage 5**: Functions and scope (46 tasks) - ✅ COMPLETE (91.3%)
6. **Stage 6**: Type checking (50 tasks) - ✅ COMPLETE
7. **Stage 7**: Object-oriented features (156 tasks) - 🔄 IN PROGRESS (55.8%)
   - Classes: COMPLETE (87/73 tasks)
   - **Interfaces: REQUIRED** (0/83 tasks) - expanded based on reference implementation analysis
8. **Stage 8**: Additional features (93 tasks) [+31 from property expansion]
9. **Stage 9**: Deferred Tasks from Stage 8
10. **Stage 10**: Performance and polish (68 tasks)
11. **Stage 11**: Long-term evolution (54 tasks)
12. **Stage 12**: Code generation - Multi-backend architecture (~180 tasks)
    - **12.1-12.3**: MIR Foundation (47 tasks) - ~2 weeks
    - **12.4**: JS Backend MVP (45 tasks) - ~3 weeks
    - **12.5**: JS Feature Complete (60 tasks) - ~4 weeks
    - **12.6**: LLVM Backend [OPTIONAL] (45 tasks) - future work
13. **Stage 13**: AST-driven formatter & playground integration (30 tasks)

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
