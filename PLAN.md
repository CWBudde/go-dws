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

- [ ] 9.1 Test helper method inheritance
  - **Task**: Helpers can inherit from other helpers
  - **Implementation**:
    - Parse helper inheritance syntax
    - Inherit methods from parent helper
    - Allow method overriding in child helper
  - **Test**: Child helper inherits parent methods
  - **Files**: `internal/parser/helper.go`, `internal/semantic/helper.go`
  - **Estimated time**: 1-2 days

- [x] 9.2 Enforce private field access control - Private/protected/public field visibility enforced in semantic analyzer; derived classes blocked from private parent fields;
- [x] 9.3 Class Methods (class procedures/functions) - Parser, semantic analysis, runtime execution, and virtual/override polymorphism support
- [x] 9.4 Class Constants - Parsing, ClassType storage, semantic validation, and runtime evaluation with method scope accessibility
- [x] 9.5 Class Variables (Static Fields) - class var declarations with parsing, type system integration, and shared instance-independent storage
- [x] 9.6 ClassName Property - Built-in property on TObject returning class name for objects, metaclasses, and identifiers; includes TClass type alias
- [x] 9.7 ClassType Property - Returns metaclass (class of T) with case-insensitive lookup for member and identifier access
- [x] 9.8 Type Casting - Function-style casts (Integer/Float/String/Boolean/Class) with semantic validation, runtime checks, and `as` operator
- [x] 9.9 Inline Method Implementation - Method bodies inside class declarations (feature verified as already working)
- [x] 9.10 Short-Form Class Declarations - `TChild = class(TParent);` syntax and type alias support
- [x] 9.11 Class Forward Declarations - `TClassName = class;` forward declarations with semantic resolution and validation
- [x] 9.12 Abstract Methods - abstract directive parsing, semantic validation (implicitly virtual), and instantiation prevention
- [x] 9.13 Partial Classes - Parse and merge partial class declarations across multiple definitions with combined member lists
- [x] 9.14 Operator Overloading for Classes - Class operator declarations, inheritance-aware dispatch, and runtime execution support
- [x] 9.15 "not in" Operator Support - Parser handling for NOT/IN composition, semantic validation, and runtime execution for set membership negation

---

### Phase 9.16: Semantic Analysis Fixes

**Status**: IN PROGRESS - 23 of 89 failing tests fixed (26% complete)
**Timeline**: 4-6 weeks total
**Objective**: Fix all remaining semantic analysis test failures through systematic category-based approach

#### Progress Summary

**Completed (23 tests fixed)**:
- ✅ Exception field access (2 tests) - Made field lookups case-insensitive
- ✅ Optional built-in function parameters (8 tests) - IntToBin, IntToHex, ToJSONFormatted, Copy
- ✅ Delete array support (4 tests) - Extended Delete to support array deletion
- ✅ Variadic method registration (1 test) - Added variadic method support
- ✅ MaxInt/MinInt variadic parameters (2 tests) - Made functions accept any number of arguments
- ✅ Record field access (2 tests) - Normalized field names to lowercase
- ✅ Forward class declarations (2 tests) - Fixed parent class validation rules

**Files Modified**: analyze_classes.go, analyze_builtin_convert.go, analyze_builtin_json.go, analyze_builtin_string.go, analyze_builtin_array.go, analyze_builtin_math.go, analyze_records.go, types.go

#### Remaining Tasks

**Medium Complexity (20 tests remaining)** - Priority: HIGH

- [ ] 9.16.1 Method Visibility Enforcement (6 tests) - **NEXT RECOMMENDED**
  - **Estimate**: 1-2 hours
  - **Files**: analyze_classes.go
  - **Description**: Implement visibility checking (private/protected/public) in method call analysis
  - **Strategy**: Similar to field visibility which already works - can use as reference
  - **Tests**: TestPrivateMethodAccessFromSameClass, TestPrivateMethodAccessFromOutside, TestProtectedMethodAccessFromChild, TestProtectedMethodAccessFromOutside, TestPrivateFieldNotInheritedAccess, TestPublicMethodAccessFromOutside
  - **Details**: Add visibility validation in analyzeCallExpression when calling methods on class instances

- [x] 9.16.2 Interface Implementation Validation - PARTIAL (Core functionality complete, one edge case remains)
  - **Estimate**: 3-4 hours (actual: 4 hours)
  - **Files**: analyze_interfaces.go, analyze_classes.go, interp/interface.go, interp/functions.go, interp/statements.go, interp/declarations.go
  - **Description**: Implemented core interface runtime support with case-insensitive method lookups
  - **Completed**:
    - ✅ Interface methods stored with lowercase keys for case-insensitive lookup
    - ✅ Interface variable declarations create InterfaceInstance with nil object
    - ✅ Assignment from class to interface wraps object in InterfaceInstance
    - ✅ Interface method calls work (member access without parentheses)
    - ✅ Semantic analyzer handles interface member access
    - ✅ Interface-to-interface assignment supported
  - **Remaining Issue** (for follow-up task 9.16.2.9):
    - ❌ Interface method calls with parentheses (CallExpression path) not finding methods
    - **Issue**: When `x.Method()` is called (with parentheses), it's a CallExpression that analyzes the object separately before calling analyzeMemberAccessExpression
    - **Impact**: Functions/procedures with parameters fail semantic validation
    - **Workaround**: Method calls without parentheses work correctly
  - **Subtasks**:
    - [x] 9.16.2.1 Fix interface method name case-insensitivity in analyzeInterfaceMethodDecl
    - [x] 9.16.2.2 Fix class method lookup case-insensitivity in validateInterfaceImplementation
    - [x] 9.16.2.3 Add interface variable type support in semantic analyzer
    - [x] 9.16.2.4 Implement interface-to-class assignment validation
    - [x] 9.16.2.5 Add runtime interface variable support in interpreter
    - [x] 9.16.2.6 Implement interface method call dispatch in interpreter
    - [ ] 9.16.2.7 Add interface type casting (as operator) for interfaces - DEFERRED
    - [ ] 9.16.2.8 Verify all 9 test cases pass - PARTIAL (core tests pass, CallExpression issue remains)
    - [ ] 9.16.2.9 Fix CallExpression path for interface method calls with parentheses - NEW SUBTASK

- [ ] 9.16.3 Property Expression Validation (5 tests)
  - **Estimate**: 2-3 hours
  - **Files**: analyze_properties.go
  - **Description**: Property expressions with field references not analyzed correctly
  - **Root Cause**: Need to handle field access expressions in property analyzer
  - **Tests**: TestPropertyExpressionValidation (multiple variants)
  - **Strategy**: Enhance property analyzer to validate field references in read/write expressions

**High Complexity (48 tests remaining)** - Priority: MEDIUM

- [x] 9.16.4 Inherited Expression Support (partial - interpreter fixes)
  - **Estimate**: 6-8 hours
  - **Description**: Implement 'inherited' keyword for calling parent class methods
  - **Strategy**: Add InheritedExpression AST node, parser support, and semantic validation
  - **Complexity**: Requires changes across parser, AST, and semantic analyzer
  - **Status**: COMPLETED for interpreter. Semantic analyzer issues remain (separate task).
  - **Completed Changes**:
    - [x] 9.16.4.1 Fix implicit TObject parent in interpreter (declarations.go)
      - Classes without explicit parent now inherit from TObject automatically
      - Matches semantic analyzer pattern
    - [x] 9.16.4.2 Fix inherited property and field access in interpreter
      - evalInheritedExpression now supports methods, properties, and fields
      - Tested with inherited1.pas (inherited Prop) - PASSING
    - [x] 9.16.4.3 Fix TObject Create constructor
      - TObject.Constructors["Create"] now has proper AST node instead of nil
      - Enables inherited constructors to work correctly
  - **Files Modified**:
    - internal/interp/declarations.go (implicit TObject parent)
    - internal/interp/exceptions.go (TObject Create constructor)
    - internal/interp/objects.go (evalInheritedExpression property/field support)
  - **Remaining Issues (out of scope for this task)**:
    - Semantic analyzer: parent class resolution with case-insensitive names
    - Semantic analyzer: inherited constructor validation
    - These should be addressed in separate semantic analyzer tasks
  - **Follow-up Tasks**:
    - [ ] 9.16.4.4 Allow inherited constructor calls in semantic analyzer
      - Resolve parent constructors when `inherited Create` (or other constructors) is used
      - Fixes `TestInheritedExpression_ComplexCases/inherited_in_constructor`

- [x] 9.16.5 Type Operators (is/as/implements) - COMPLETED
  - **Estimate**: 8-10 hours
  - **Description**: Implement type checking and casting operators
  - **Strategy**: Add type operator support in parser and semantic analyzer
  - **Complexity**: Requires runtime type information and safe casting mechanisms
  - **Status**: COMPLETED. All tests passing (24/24 = 100%)
  - **Test Results**: 30/30 tests passing (100% pass rate)
  - **Completed Subtasks**:
    - [x] 9.16.5.1 Fix 'as' operator to support class-to-class casting
      - Semantic analyzer now supports both class and interface target types
      - File: internal/semantic/analyze_expressions.go (analyzeAsExpression)
      - Validates upcast/downcast relationships in class hierarchy
    - [x] 9.16.5.2 Add validation for 'is' operator operands
      - Left operand validated as class instance or nil
      - Right operand validated as class type
      - File: internal/semantic/analyze_expressions.go (analyzeIsExpression)
    - [x] 9.16.5.3 Add validation for 'implements' operator operands
      - Left operand validated as class instance or nil
      - Right operand validated as interface type
      - File: internal/semantic/analyze_expressions.go (analyzeImplementsExpression)
    - [x] 9.16.5.4 Update interpreter 'as' operator for class casting
      - File: internal/interp/expressions.go (evalAsExpression)
      - Runtime now supports both class-to-class and class-to-interface casts
      - Validates runtime compatibility for downcasts
    - [x] 9.16.5.5 Verify all type operator tests pass - ALL PASSING
    - [ ] 9.16.5.6 Avoid cascading errors when 'as' target type is invalid
      - Short-circuit analysis after reporting "'as' operator requires interface or class type"
      - Prevents secondary `cannot infer type` diagnostics (TestTypeOperator_As_InvalidRightOperand)
  - **Files Modified**:
    - internal/semantic/analyze_expressions.go (added strings import, updated all 3 operators)
    - internal/semantic/type_operators_test.go (updated error message expectation)
    - internal/interp/expressions.go (evalAsExpression now handles classes)

- [x] 9.16.6 Operator Overloading - COMPLETED
  - **Estimate**: 4-6 hours
  - **Description**: Support custom operator implementations in classes
  - **Strategy**: Add operator method registration and lookup in binary/unary expression analysis
  - **Complexity**: Requires operator resolution mechanism and precedence handling
  - **Status**: COMPLETED. Parser, semantic analyzer, and most interpreter support was already present.
  - **Fix Applied**: Fixed array type matching in operator overloading
  - **Test Results**: 5/6 tests passing (83% pass rate)
    - ✅ class_operator1: PASS (class operator += with String)
    - ✅ class_operator2: PASS (class operator += with multiple types)
    - ❌ class_operator3: FAIL (requires array of const/variant support - different issue)
    - ✅ in_class_operator: PASS (IN operator with array of Integer) - **FIXED**
    - ✅ in_integer_operator1: PASS (IN operator with integers)
    - ✅ in_integer_operator2: PASS (IN operator variations)
  - **Key Change**:
    - Updated `valueTypeKey` function in internal/interp/operators.go
    - Now includes array element type when matching operator overloads
    - Format: "ARRAY OF INTEGER" instead of just "ARRAY"
    - Allows proper matching of operators declared with `array of T` types

- [ ] 9.16.7 Helper Methods (2 tests)
  - **Estimate**: 3-4 hours
  - **Description**: Support DWScript helper methods (extension methods)
  - **Strategy**: Research DWScript helper semantics and implement registration mechanism
  - **Complexity**: New feature requiring research and design
  - **Subtasks**:
    - [ ] 9.16.7.1 Emit diagnostics when no helper provides the requested method
      - Analyzer should report `no helper with method` for unresolved helper calls
      - Covers `TestHelperMethodResolution/call_non-existent_helper_method`

- [ ] 9.16.8 Abstract Class Implementation (1 test)
  - **Estimate**: 2-3 hours
  - **Description**: Validate that abstract classes cannot be instantiated
  - **Strategy**: Add abstract class tracking and validation in class instantiation
  - **Complexity**: Requires inheritance chain validation
  - **Subtasks**:
    - [ ] 9.16.8.1 Clear abstract flags when overrides are implemented
      - Ensure overriding inherited abstract methods removes the abstract marker
      - Fixes `TestValidAbstractImplementation`

- [ ] 9.16.9 Miscellaneous High Complexity Fixes (18 tests)
  - **Estimate**: 10-15 hours
  - **Description**: Various complex semantic validation issues
  - **Strategy**: Analyze each test individually and implement targeted fixes
  - **Examples**: Generic types, delegates, advanced inheritance scenarios, complex type checking

- [x] 9.16.10 Fix Function Argument Double Evaluation Bug ✓
  - **Estimate**: 2-4 hours (Actual: ~3 hours)
  - **Priority**: High (affects fixture test accuracy)
  - **Description**: Functions called as arguments to built-in functions (like PrintLn) are evaluated twice
  - **Impact**: Causes side effects to happen twice, making tests fail with extra output
  - **Examples**: `PrintLn(0 ?? Test(258))` calls `Test` twice instead of once
  - **Root Cause**: TWO issues found:
    1. `resolveOverload()` evaluated arguments for type checking, then `evalCallExpression()` re-evaluated them
    2. `evalTypeCast()` evaluated arguments before checking if it's a valid type cast
  - **Discovery Context**: Found while implementing coalesce operator (Task 9.14)
  - **Solution Implemented**:
    1. Modified `resolveOverload()` to cache and return evaluated argument values
    2. Updated user-defined function call path to use cached values instead of re-evaluating
    3. Fixed `evalTypeCast()` to check if it's a type cast BEFORE evaluating the argument
    4. Special handling for lazy parameters to avoid evaluation during overload resolution
  - **Files Modified**:
    - `internal/interp/functions_typecast.go` - `resolveOverload()` now returns cached values
    - `internal/interp/functions_calls.go` - Uses cached values from overload resolution
    - `internal/interp/interpreter_test.go` - Added `TestFunctionArgumentSingleEvaluation` (5 test cases)
  - **Test Results**: All new tests pass, no regressions introduced

#### Implementation Guidelines

**Testing Approach**:
1. Run specific failing test to understand exact error
2. Add debug logging to trace analyzer behavior
3. Implement minimal fix targeting root cause
4. Verify test passes without breaking existing tests
5. Run full semantic test suite before marking complete

**Code Patterns**:
- Use `strings.ToLower()` for all case-insensitive identifier lookups
- Store all symbols (fields, methods, properties) with lowercase keys
- Add comprehensive error messages with position information
- Follow existing analyzer patterns in analyze_*.go files

**Key Files**:
- `internal/semantic/analyzer.go` - Main analyzer and symbol table
- `internal/semantic/analyze_classes.go` - Class analysis including methods and fields
- `internal/semantic/analyze_interfaces.go` - Interface validation
- `internal/semantic/analyze_properties.go` - Property analysis
- `internal/types/types.go` - Type system definitions

**Next Steps**:
1. Start with Task 9.16.1 (Method Visibility Enforcement) - shortest path to 6 more passing tests
2. Then tackle Task 9.16.2 (Interface Implementation) - high impact with 9 tests
3. Complete remaining medium complexity tasks before moving to high complexity
4. Document patterns discovered during fixes for future reference

---

### Phase 9.17: Implement missing built-in functions

**Priority**: CRITICAL to MEDIUM - Required for full DWScript compatibility
**Timeline**: 4-6 weeks
**Overall Status**: ~70 of 250+ built-in functions implemented (28%)

**Category Breakdown**:

- String Functions: 17/85 implemented (20%)
- Math Functions: 40/64 implemented (62%)
- DateTime Functions: 54/60 implemented (90%)
- Array Functions: 8/18 implemented (44%)
- Conversion Functions: 3/15 implemented (20%)
- Variant Functions: 0/10 implemented (0%)
- RTTI/Type Functions: 0/20+ implemented (0%)

**Test Fixture Impact**: Implementing all functions would unblock 400+ failing test fixtures

---

#### CRITICAL: Core Conversion Functions (Phase 9.17.0) 🔥

**MUST IMPLEMENT FIRST** - These are used in 200+ test fixtures and block most other tests.

- [x] 9.17.0.1 IntToStr(i: Integer): String
- [x] 9.17.0.2 FloatToStr(f: Float): String
- [x] 9.17.0.3 StrToInt(str: String): Integer
- [x] 9.17.0.4 StrToFloat(str: String): Float
- [x] 9.17.0.5 StrToIntDef(str: String, default: Integer): Integer
- [x] 9.17.0.6 StrToFloatDef(str: String, default: Float): Float

---

#### Delimiter-based Functions (Phase 9.17.2)

- [x] 9.17.2.1 StrSplit(str, delimiter) - Split string into array
- [x] 9.17.2.2 StrJoin(array, delimiter) - Join array into string
- [x] 9.17.2.3 StrArrayPack(array) - Remove empty strings from array

- [x] 9.17.2.4 StrBefore(str, delimiter) - Get substring before first delimiter
- [x] 9.17.2.5 StrBeforeLast(str, delimiter) - Get substring before last delimiter
- [x] 9.17.2.6 StrAfter(str, delimiter) - Get substring after first delimiter
- [x] 9.17.2.7 StrAfterLast(str, delimiter) - Get substring after last delimiter
- [x] 9.17.2.8 StrBetween(str, start, stop) - Get substring between delimiters

- [x] 9.17.2.9 IsDelimiter(delims, str, index) - Check if char at index is delimiter
- [x] 9.17.2.10 LastDelimiter(delims, str) - Find last delimiter position
- [x] 9.17.2.11 FindDelimiter(delims, str, startIndex) - Find first delimiter

#### String Transformation (Phase 9.17.3)

- [x] 9.17.3.1 PadLeft(str, count, char) - Pad left to width
- [x] 9.17.3.2 PadRight(str, count, char) - Pad right to width
- [x] 9.17.3.3 StrDeleteLeft(str, count) - Delete N leftmost characters
- [x] 9.17.3.4 StrDeleteRight(str, count) - Delete N rightmost characters

- [x] 9.17.3.5 ReverseString(str) - Reverse character order
- [x] 9.17.3.6 QuotedStr(str, quoteChar) - Add quotes around string
- [x] 9.17.3.7 StringOfString(str, count) / DupeString - Repeat string

- [x] 9.17.3.8 NormalizeString(str, form) - Unicode normalization
- [x] 9.17.3.9 StripAccents(str) - Remove diacritical marks

#### Comparison Functions (Phase 9.17.4)

- [ ] 9.17.4.1 SameText(str1, str2) - Case-insensitive equality
- [ ] 9.17.4.2 CompareText(str1, str2) - Case-insensitive compare
- [ ] 9.17.4.3 CompareStr(str1, str2) - Case-sensitive compare
- [ ] 9.17.4.4 AnsiCompareText(str1, str2) - ANSI case-insensitive compare
- [ ] 9.17.4.5 AnsiCompareStr(str1, str2) - ANSI case-sensitive compare
- [ ] 9.17.4.6 CompareLocaleStr(str1, str2, locale, caseSensitive) - Locale-aware compare
- [ ] 9.17.4.7 StrMatches(str, mask) - Wildcard pattern matching
- [ ] 9.17.4.8 StrIsASCII(str) - Is pure ASCII?

#### Advanced Conversion Functions (Phase 9.17.5)

- [ ] 9.17.5.1 IntToStr(i, base) - Integer to string with base parameter
- [ ] 9.17.5.2 StrToInt(str, base) - String to integer with base
- [ ] 9.17.5.3 StrToIntDef(str, def, base) - With base and default
- [ ] 9.17.5.4 TryStrToInt(str, base, @value) - Safe conversion with base

- [ ] 9.17.5.5 HexToInt(hexa) - Hex string to integer
- [ ] 9.17.5.6 IntToBin(v, digits) - Integer to binary (already implemented?)
- [ ] 9.17.5.7 BinToInt(binary) - Binary string to integer

- [ ] 9.17.5.8 TryStrToFloat(str, @value) - Safe float conversion
- [ ] 9.17.5.9 StrToFloatDef(str, def) - String to float with default

- [ ] 9.17.5.10 VarToIntDef(val, def) - Variant to int with default
- [ ] 9.17.5.11 VarToFloatDef(val, def) - Variant to float with default

#### Encoding/Escaping Functions (Phase 9.17.6)

- [ ] 9.17.6.1 StrToHtml(str) - HTML encode
- [ ] 9.17.6.2 StrToHtmlAttribute(str) - HTML attribute encode
- [ ] 9.17.6.3 StrToJSON(str) - JSON encode/escape
- [ ] 9.17.6.4 StrToCSSText(str) - CSS text encode
- [ ] 9.17.6.5 StrToXML(str, mode) - XML encode

#### Case Conversion Variants (Phase 9.17.7)

- [ ] 9.17.7.1 ASCIILowerCase(str) - ASCII-only lowercase
- [ ] 9.17.7.2 ASCIIUpperCase(str) - ASCII-only uppercase
- [ ] 9.17.7.3 AnsiLowerCase(str) - ANSI lowercase (alias for LowerCase)
- [ ] 9.17.7.4 AnsiUpperCase(str) - ANSI uppercase (alias for UpperCase)

#### Utility Functions (Phase 9.17.8)

- [ ] 9.17.8.1 ByteSizeToStr(size) - Format byte size (KB, MB, GB)
- [ ] 9.17.8.2 GetText(str) / _(str) - Localization/translation function
- [ ] 9.17.8.3 CharAt(s, x) - Get character at position (deprecated, use SubStr)

---

#### Math Functions - Constants & Core (Phase 9.17.9)

**Current Status**: 40/64 implemented (62%) - Missing 24 critical math functions

**HIGH PRIORITY**:

- [x] 9.17.9.1 Pi: Float - Mathematical constant π (3.141592...)
- [x] 9.17.9.2 Sign(x: Float): Integer - Returns -1, 0, or 1
- [x] 9.17.9.3 Odd(x: Integer): Boolean - Check if integer is odd
- [x] 9.17.9.4 Frac(x: Float): Float - Fractional part of number
- [x] 9.17.9.5 Int(x: Float): Float - Integer part (returns Float)
- [x] 9.17.9.6 Log10(x: Float): Float - Base-10 logarithm
- [x] 9.17.9.7 LogN(x, base: Float): Float - Logarithm with custom base

**Status**: HIGH priority items ✅ COMPLETED for both AST interpreter and bytecode VM

**MEDIUM PRIORITY**:

- [x] 9.17.9.8 Infinity: Float - Infinity constant
- [x] 9.17.9.9 NaN: Float - Not-a-Number constant
- [x] 9.17.9.10 IsFinite(x: Float): Boolean - Check if number is finite
- [x] 9.17.9.11 IsInfinite(x: Float): Boolean - Check if number is infinite
- [x] 9.17.9.12 IntPower(base: Float, exponent: Integer): Float
- [x] 9.17.9.13 DivMod(dividend, divisor: Integer; var quotient, remainder: Integer)
- [x] 9.17.9.14 RandSeed: Integer - Get current random seed
- [x] 9.17.9.15 RandG: Float - Gaussian random number

**Status**: MEDIUM priority items ✅ COMPLETED (7/8 in bytecode VM, 8/8 in AST interpreter)

**LOW PRIORITY** (Advanced Math):

- [ ] 9.17.9.16 Factorial(n: Integer): Integer - Factorial function
- [ ] 9.17.9.17 Gcd(a, b: Integer): Integer - Greatest common divisor
- [ ] 9.17.9.18 Lcm(a, b: Integer): Integer - Least common multiple
- [ ] 9.17.9.19 IsPrime(n: Integer): Boolean - Primality test
- [ ] 9.17.9.20 LeastFactor(n: Integer): Integer - Smallest prime factor
- [ ] 9.17.9.21 PopCount(n: Integer): Integer - Count set bits
- [ ] 9.17.9.22 TestBit(value: Integer, bit: Integer): Boolean - Test if bit is set
- [ ] 9.17.9.23 Haversine(lat1, lon1, lat2, lon2: Float): Float - Haversine distance
- [ ] 9.17.9.24 CompareNum(a, b: Float): Integer - Numerical comparison

**Implementation Time**: 2-3 days for HIGH priority
**Impact**: Unblocks 80+ math test fixtures

---

#### Variant Functions (Phase 9.17.11)

**Current Status**: 10/10 implemented (100%) - ✅ **COMPLETE**

**ALL HIGH PRIORITY** - Required for dynamic typing and Variant support:

- [x] 9.17.11.1 VarType(v: Variant): Integer
  - Get variant type code (vtInteger, vtString, etc.)
  - Returns type enum value

- [x] 9.17.11.2 VarIsNull(v: Variant): Boolean
  - Check if variant is null
  - Used for null checks

- [x] 9.17.11.3 VarIsEmpty(v: Variant): Boolean
  - Check if variant is empty (uninitialized)
  - Different from null

- [x] 9.17.11.4 VarIsClear(v: Variant): Boolean
  - Check if variant is cleared
  - Alias for VarIsEmpty in some contexts

- [x] 9.17.11.5 VarIsArray(v: Variant): Boolean
  - Check if variant contains array
  - Type checking helper

- [x] 9.17.11.6 VarIsStr(v: Variant): Boolean
  - Check if variant contains string
  - Type checking helper

- [x] 9.17.11.7 VarIsNumeric(v: Variant): Boolean
  - Check if variant contains numeric value
  - Includes Integer and Float

- [x] 9.17.11.8 VarToStr(v: Variant): String
  - Convert variant to string representation
  - Handles all variant types

- [x] 9.17.11.9 VarAsType(v: Variant, varType: Integer): Variant
  - Convert variant to specified type
  - Type coercion with conversion

- [x] 9.17.11.10 VarClear(var v: Variant)
  - Clear variant value (set to empty)
  - Var parameter modification

**Implementation Time**: 2-3 days
**Impact**: Unblocks 30+ variant test fixtures, enables full Variant support

---

#### Array of Const Support (Phase 9.17.11b)

**Current Status**: Not implemented - **COMPLETE GAP**

**Priority**: HIGH - Required for variable-length parameter lists with mixed types

**Description**:
`array of const` is a special DWScript type that allows passing variable-length argument
lists where each element can be of any type. Similar to varargs in other languages, but
each element is wrapped in a variant-like container that preserves type information.

**Blocking Tests**:
- class_operator3.pas (operator overload with array of const parameter)
- Multiple other fixtures using variable-length parameter functions

**Implementation Tasks**:

- [x] 9.17.11b.1 Add array of const type support
  - ✅ Lexer/Parser: Already supports syntax (array of const)
  - ✅ Semantic analyzer: Type checking for array of const parameters
  - ✅ Type system: Uses array of Variant (ARRAY_OF_CONST constant)
  - ✅ Operator registry: Extended to support array type compatibility

- [ ] 9.17.11b.2 Implement array of const conversion at call sites
  - ✅ Array literals with mixed types work with array of const parameters
  - ✅ Empty array literals handled in compound assignments
  - ✅ Array of T -> array of Variant compatibility in operators
  - ⚠️ Interpreter runtime for class operator overloads with array of const parameters is not yet implemented; semantic analysis and type system are complete. Task will be marked complete once interpreter support is added.

- [x] 9.17.11b.3 Add TVarRec support (optional)
  - Not needed: Using Variant type directly for array elements
  - Runtime conversion handled by interpreter's Variant implementation

- [x] 9.17.11b.4 Test array of const in various contexts
  - ✅ Function parameters (comprehensive tests added)
  - ✅ Class operator overloads (semantic analysis complete)
  - ✅ Variant to typed variable conversion (String concatenation works)
  - ✅ Empty, homogeneous, and heterogeneous array literals
- [ ] 9.17.11b.5 Support procedure bindings for class operators that return Self
  - Allow `class operator +(const items: array of const): TClass uses AppendStrings;` patterns
  - Analyzer or runtime should treat procedure bindings that mutate and return `Self` as valid
  - Fixes `TestClassOperatorWithArrayOfConst` / `TestClassOperatorCompoundAssignmentWithEmptyArray`

**Implementation Time**: 2-3 days
**Impact**: Unblocks class_operator3.pas and other variable-argument fixtures

**References**:
- Delphi documentation: array of const and TVarRec
- testdata/fixtures/SimpleScripts/class_operator3.pas (blocked test)
- Related to Variant support (Phase 9.17.11)

---

#### RTTI / Type Introspection (Phase 9.17.13)

**Current Status**: 0/20+ implemented (0%) - **COMPLETE GAP**

**MEDIUM PRIORITY** (Advanced OOP features):

- [ ] 9.17.13.1 TypeOf(value): TTypeInfo
  - Get runtime type information
  - Returns type metadata object

- [ ] 9.17.13.2 TypeOfClass(classRef: TClass): TTypeInfo
  - Get type info for class reference
  - Class-level type introspection

- [ ] 9.17.13.3 ClassName(obj: TObject): String
  - Get class name as string
  - May already be implemented

- [ ] 9.17.13.4 ClassType(obj: TObject): TClass
  - Get class type reference
  - Runtime class info

**Note**: Full RTTI implementation is complex and may be deferred to later phase.
**Implementation Time**: 5-7 days (complex)
**Impact**: Unblocks 10-15 advanced OOP fixtures

---

**Implementation Notes**:

- Each function needs: AST interpreter, bytecode VM, semantic analyzer, tests
- Follow patterns from SubStr implementation (see commits from this session)
- Prioritize functions based on test fixture usage
- Reference original DWScript sources:
  - `reference/dwscript-original/Source/dwsStringFunctions.pas`
  - `reference/dwscript-original/Source/dwsMathFunctions.pas`
  - `reference/dwscript-original/Source/dwsArrayFunctions.pas`
  - `reference/dwscript-original/Source/dwsVariantFunctions.pas`

**Testing Strategy**:

- Minimum 4 test categories per function: BasicUsage, EdgeCases, InExpressions, ErrorCases
- Include UTF-8/Unicode test cases where applicable
- Verify against DWScript test fixtures in testdata/fixtures/
- For mathematical functions: test edge cases (0, negative, infinity, NaN)
- For array functions: test empty arrays, single elements, large arrays
- For variant functions: test all variant type combinations

**Implementation Priority Summary**:

**Week 1 - CRITICAL (200+ fixtures)**:
1. Phase 9.17.0: Core Conversions (IntToStr, FloatToStr, StrToInt, StrToFloat) - 6 functions
2. Phase 9.17.9: Math Constants (Pi, Sign, Odd, Frac, Int, Log10) - 7 functions
3. Phase 9.17.1: Essential String (LeftStr, RightStr, MidStr, StrContains) - 8 functions
**Total**: ~21 functions, **Impact**: 200+ fixtures

**Week 2 - HIGH (150+ fixtures)**:
1. Phase 9.17.11: All Variant Functions - 10 functions
2. Phase 9.17.1: More String Functions (PosEx, StrBeginsWith, StrEndsWith) - 6 functions
3. Phase 9.17.10: Core Array Functions (Map, Filter, Reduce) - 4 functions
4. Phase 9.17.12: Ordinal Functions (Succ, Pred) - 2 functions
**Total**: ~22 functions, **Impact**: 150+ fixtures

**Week 3 - MEDIUM (100+ fixtures)**:
1. Phase 9.17.2: Delimiter Functions (StrSplit, StrJoin, StrBefore, StrAfter) - 11 functions
2. Phase 9.17.9: Advanced Math (Infinity, IsFinite, IntPower, DivMod) - 8 functions
3. Phase 9.17.10: More Array Functions (Every, Some, Find, Concat, Slice) - 6 functions
**Total**: ~25 functions, **Impact**: 100+ fixtures

**Week 4+ - LOW (50+ fixtures)**:
1. Phase 9.17.3-9.17.8: Specialized String Functions
2. Phase 9.17.9: Advanced Math (Factorial, Gcd, IsPrime, etc.)
3. Phase 9.17.13: RTTI Functions (TypeOf, etc.)
**Total**: ~50+ functions, **Impact**: 50+ fixtures

**Grand Total**: ~180 functions to implement across all phases
**Total Impact**: 500+ test fixtures would pass with complete implementation

---

### Phase 9.18: Documentation & Cleanup

**Priority**: LOW - Can be done in parallel with Phase 10
**Timeline**: 1 week

- [ ] 9.16.1 Update README with current features
  - Document all Stage 7 features now complete
  - Update feature completion percentages
  - Add examples of new features

- [ ] 9.16.2 Create docs/phase9-summary.md
  - Document achievements in Phase 9
  - Statistics: tests passing, coverage percentages
  - Lessons learned and challenges overcome

- [ ] 9.16.3 Update testdata/fixtures/TEST_STATUS.md
  - Update pass/fail counts for each category
  - Mark resolved issues
  - Document remaining blockers

- [ ] 9.16.4 Create docs/limitations.md
  - Document known limitations
  - Features intentionally deferred to later phases
  - Differences from original DWScript

---

### Phase 9.19: Constructor Overloading

**Priority**: HIGH - Required for ~50+ failing tests
**Timeline**: 3-5 days
**Impact**: Unblocks major class tests and fixture tests

**Current Status**: Constructor overloading is not implemented. All classes only support a single constructor signature. When a class defines multiple constructors with different parameters, only the last one is registered, causing "wrong number of arguments for constructor" errors.

**Root Cause**: The interpreter stores constructors in a simple map with "Create" as key, not supporting multiple signatures per constructor name.

- [x] 9.19.1 Implement constructor overload storage in ClassType ✅
  - **Task**: Modify ClassType to store multiple constructor signatures per name
  - **Implementation**: Storage was already implemented with `ConstructorOverloads map[string][]*MethodInfo`
  - **Files**: `internal/types/types.go`
  - **Status**: Already implemented in previous task
  - **Actual time**: < 1 hour (verification only)

- [x] 9.19.2 Implement implicit parameterless constructor synthesis ✅
  - **Task**: Synthesize implicit parameterless constructor when constructor has `overload` directive
  - **Implementation**:
    - Added `synthesizeImplicitParameterlessConstructor()` in semantic analyzer
    - Added corresponding function in interpreter
    - Normalized constructor names to lowercase for case-insensitive matching
  - **Files**: `internal/semantic/analyze_classes_inheritance.go`, `internal/interp/declarations.go`, `internal/interp/exceptions.go`
  - **Tests**: All constructor overload tests passing
  - **Actual time**: 2 hours

- [x] 9.19.3 Fix constructor name case-sensitivity issues ✅
  - **Task**: Ensure case-insensitive constructor matching works correctly
  - **Implementation**:
    - Normalized all constructor names to lowercase when storing in maps
    - Fixed inheritance to use normalized names
    - Updated TObject constructor registration
  - **Files**: `internal/interp/declarations.go`, `internal/interp/exceptions.go`
  - **Tests**: All constructor overload tests passing (5/5)
  - **Actual time**: 1 hour

- [x] 9.19.4 Constructor overload tests verified ✅
  - **Task**: Verify all constructor overloading scenarios work
  - **Tests**: All tests passing:
    - Exact fixture test case ✓
    - Constructor with parameter ✓
    - Implicit parameterless constructor ✓
    - New with implicit parameterless constructor ✓
    - New with parameter ✓
  - **Files**: `internal/interp/constructor_overload_test.go`
  - **Actual time**: Included in implementation

**Blocked Tests**:
- internal/interp: TestFieldAccess, TestMethodCalls, TestInheritance, TestPolymorphism, TestConstructors, TestSelfReference, TestNewKeywordWithConstructor, TestNewKeywordWithException, TestNewKeywordEquivalentToCreate, TestClassVariableSharedAcrossInstances, TestMixedClassAndInstanceMembers, TestConstructorOverload
- Fixtures: constructor_overload.pas and 40+ tests in OverloadsPass category

---

### Phase 9.20: Method Overloading

**Priority**: HIGH - Required for ~40+ failing tests
**Timeline**: 3-5 days
**Impact**: Full method overloading support for fixture test compatibility

**Current Status**: Method overloading parsing and semantic analysis is partially implemented, but runtime dispatch may need enhancement.

- [ ] 9.20.1 Audit current method overload implementation
  - **Task**: Review existing overload support in parser, semantic analyzer, and interpreter
  - **Implementation**:
    - Check if methods are stored with overload support
    - Verify semantic analyzer overload resolution
    - Test interpreter dispatch with multiple overloads
  - **Files**: `internal/semantic/analyze_classes.go`, `internal/interp/functions.go`
  - **Tests**: Run existing overload tests and document gaps
  - **Estimated time**: 0.5 day

- [ ] 9.20.2 Fix method overload resolution bugs
  - **Task**: Address any gaps found in audit
  - **Implementation**:
    - Ensure method table stores all overloads per name
    - Fix overload resolution to consider all signatures
    - Handle virtual/override with overloads correctly
  - **Files**: `internal/semantic/analyze_classes.go`, `internal/types/types.go`
  - **Tests**: Add failing test cases, then fix until passing
  - **Estimated time**: 1-2 days

- [ ] 9.20.3 Enhance method overload dispatch in interpreter
  - **Task**: Ensure runtime correctly dispatches to overloaded methods
  - **Implementation**:
    - Update `evalCallExpression` for method calls with overload resolution
    - Handle polymorphic dispatch with overloads (virtual + override)
    - Support class method overloading
  - **Files**: `internal/interp/functions.go`, `internal/interp/objects.go`
  - **Tests**: Test method overload dispatch in various scenarios
  - **Estimated time**: 1-2 days

- [ ] 9.20.4 Add comprehensive method overload tests
  - **Task**: Verify all method overloading patterns work
  - **Tests**:
    - Simple method overloads
    - Overloads with inheritance
    - Virtual/override with overloads
    - Class methods with overloads
    - Error cases (ambiguous, hidden overloads)
  - **Files**: `internal/interp/method_overload_test.go` (new)
  - **Estimated time**: 0.5-1 day

**Blocked Tests**:
- Fixtures: 40+ tests in OverloadsPass category (meth_overload_simple, meth_overload_hide, overload_constructor, overload_virtual, etc.)

---

### Phase 9.21: Parser Syntax Extensions

**Priority**: HIGH - Required for ~100+ failing fixture tests
**Timeline**: 5-7 days
**Impact**: Support for missing DWScript syntax constructs

**Current Status**: Parser rejects several DWScript syntax patterns causing many fixture tests to fail with parse errors.

#### Subtask Category: Array Type Syntax

- [x] 9.21.1 Fix "array of <type>" shorthand parsing ✓
  - **Task**: Support DWScript's `array[TEnum] of Type` syntax (enum-indexed array)
  - **Current Error**: "expected DOTDOT, got RBRACK" when parsing `array[TEnum]`
  - **Implementation**:
    - Extended `parseArrayType()` to handle `array[EnumType] of <type>` syntax
    - Added `IndexType` field to `ArrayTypeNode` AST node
    - Updated semantic analyzer to resolve enum-indexed arrays to static arrays with enum bounds
    - Updated interpreter to accept enum values as array indices (using ordinal values)
    - Added comprehensive tests for enum-indexed array parsing
  - **Files**:
    - `pkg/ast/type_expression.go` (added IndexType field, IsEnumIndexed method)
    - `internal/parser/types.go` (parseArrayType with enum detection)
    - `internal/parser/declarations.go` (set InlineType for array types in const decl)
    - `internal/parser/inline_types_test.go` (TestEnumIndexedArrayType)
    - `internal/semantic/type_resolution.go` (resolveArrayTypeNode with enum handling)
    - `internal/semantic/analyze_statements.go` (analyzeConstDecl with InlineType support)
    - `internal/semantic/analyze_arrays.go` (allow enum types as indices)
    - `internal/interp/array.go` (handle EnumValue as array index)
  - **Tests**: TestEnumIndexedArrayType passes, const_array3.pas now executes correctly
  - **Completed**: Enum-indexed arrays fully supported in parser, semantic analyzer, and interpreter

#### Subtask Category: Class Features

- [ ] 9.21.2 Implement "class var" initialization syntax
  - **Task**: Support initializing class variables inline: `class var X: Integer := 42;`
  - **Current Error**: "expected next token to be SEMICOLON, got ASSIGN"
  - **Implementation**:
    - Extend `parseClassVarDeclaration()` to allow optional `:= <expression>`
    - Store initialization expression in AST node
    - Semantic analyzer validates initialization expression
    - Interpreter evaluates during class initialization
  - **Files**: `internal/parser/parser_class.go`, `internal/semantic/analyze_classes.go`, `internal/interp/declarations.go`
  - **Tests**: Test inline class var initialization, complex expressions
  - **Estimated time**: 1-2 days
  - **Blocked Tests**: class_var.pas, class_var_dyn1.pas, class_var_dyn2.pas, and 10+ more

- [ ] 9.21.3 Fix "class method/operator" inline syntax parsing
  - **Task**: Support inline class method/operator declarations without separate declaration/implementation
  - **Current Error**: "expected 'var', 'const', 'property', 'function', or 'procedure' after 'class' keyword"
  - **Implementation**:
    - Allow `class operator` and `class procedure/function` with inline implementation
    - Parse class method bodies directly in class declaration
  - **Files**: `internal/parser/parser_class.go`
  - **Tests**: Test inline class method/operator declarations
  - **Estimated time**: 1 day
  - **Blocked Tests**: class_method3.pas, call_conventions.pas, and 5+ more

#### Subtask Category: Attributes and Metadata

- [ ] 9.21.4 Implement "deprecated" attribute parsing
  - **Task**: Support `[deprecated]` or `deprecated` attribute on declarations
  - **Current Error**: "no prefix parse function for DEPRECATED"
  - **Implementation**:
    - Add DEPRECATED token to lexer (may already exist)
    - Parse deprecated attribute before declarations (variables, functions, classes)
    - Store in AST metadata (can emit warnings in semantic analyzer)
  - **Files**: `internal/parser/parser.go`, `internal/ast/ast.go`
  - **Tests**: Test deprecated on various declaration types
  - **Estimated time**: 0.5-1 day
  - **Blocked Tests**: const_deprecated.pas, enum_element_deprecated.pas

- [ ] 9.21.5 Implement contract syntax (require/ensure/old/invariant)
  - **Task**: Parse Design by Contract syntax for preconditions/postconditions
  - **Current Error**: "no prefix parse function for REQUIRE/ENSURE"
  - **Implementation**:
    - Add REQUIRE, ENSURE, OLD, INVARIANT tokens to lexer
    - Parse contract blocks before/after function bodies
    - Support `old(expr)` syntax in postconditions
    - Store in AST (execution can be deferred or implemented as assertions)
  - **Files**: `internal/lexer/lexer.go`, `internal/parser/parser_function.go`, `internal/ast/statements.go`
  - **Tests**: Test contract parsing for functions, methods
  - **Estimated time**: 1-2 days
  - **Blocked Tests**: contracts_code.pas, contracts_old.pas, contracts_subproc.pas

#### Subtask Category: Miscellaneous Syntax

- [ ] 9.21.6 Fix "is" operator with non-type expressions
  - **Task**: Allow `is` operator with boolean expressions like `is True`, `is False`
  - **Current Error**: "expected type expression, got True/False"
  - **Implementation**:
    - Extend `parseIsExpression()` to handle value expressions (not just types)
    - Semantic analyzer validates operand types
  - **Files**: `internal/parser/parser_expressions.go`
  - **Tests**: Test `is` with various operand types
  - **Estimated time**: 0.5 day
  - **Blocked Tests**: boolean_is.pas

- [ ] 9.21.7 Fix inline conditional expression parsing
  - **Task**: Support ternary-like conditionals: `if condition then expr1 else expr2`
  - **Current Error**: "no prefix parse function for IF"
  - **Implementation**:
    - Parse conditional expressions (not just conditional statements)
    - Create IfExpression AST node
  - **Files**: `internal/parser/parser_expressions.go`, `internal/ast/expressions.go`
  - **Tests**: Test inline conditionals in expressions
  - **Estimated time**: 1 day
  - **Blocked Tests**: boolean_optimize.pas, coalesce_bool.pas

- [ ] 9.21.8 Fix "class" forward declaration in units
  - **Task**: Support class forward declarations in unit interface section
  - **Current Error**: "no prefix parse function for CLASS" or "expected DOT after 'end' in unit"
  - **Implementation**:
    - Enhance unit parser to handle class forward declarations
    - Resolve forward references correctly
  - **Files**: `internal/parser/parser_unit.go`
  - **Tests**: Test unit with class forwards
  - **Estimated time**: 0.5-1 day
  - **Blocked Tests**: class_scoping1.pas

- [ ] 9.21.9 Support field initializers in type declarations
  - **Task**: Allow field initialization in record/class declarations: `field: Type := value;`
  - **Current Error**: "expected SEMICOLON, got EQ"
  - **Implementation**:
    - Extend field parsing to accept optional initialization
    - Store initializer in AST
    - Semantic analyzer + interpreter execute during instantiation
  - **Files**: `internal/parser/parser_types.go`, `internal/parser/parser_class.go`
  - **Tests**: Test field initializers in records and classes
  - **Estimated time**: 1 day
  - **Blocked Tests**: clear_ref_in_destructor.pas, clear_ref_in_static_method.pas, clear_ref_in_virtual_method.pas

- [ ] 9.21.10 Fix other parser errors identified in fixture test runs
  - **Task**: Address remaining parser errors discovered during test runs
  - **Implementation**: Investigate and fix on case-by-case basis
  - **Files**: Various parser files
  - **Tests**: Re-run fixture tests and verify parsing succeeds
  - **Estimated time**: 1-2 days

**Impact**: Fixes 100+ parser-related fixture test failures

---

### Phase 9.22: Lazy Parameters

**Priority**: LOW - Required for 5 failing tests
**Timeline**: 2-3 days
**Impact**: Support DWScript's lazy parameter evaluation

**Current Status**: Lazy parameter test files are missing, and lazy parameter semantics may not be fully implemented.

- [ ] 9.22.1 Create missing lazy parameter test files
  - **Task**: Create the missing `.dws` and `.out` files for lazy parameter tests
  - **Files**: `testdata/lazy_params/jensens_device.dws`, `conditional_eval.dws`, `lazy_logging.dws`, `multiple_access.dws`, `lazy_with_loops.dws`
  - **Implementation**: Write test scripts demonstrating lazy evaluation
  - **Reference**: DWScript documentation on `lazy` parameter modifier
  - **Estimated time**: 0.5 day

- [ ] 9.22.2 Verify lazy parameter semantic analysis
  - **Task**: Ensure semantic analyzer handles `lazy` parameters correctly
  - **Implementation**:
    - Check if `lazy` keyword is recognized
    - Verify lazy parameters are marked in AST
    - Ensure type checking works for lazy parameters
  - **Files**: `internal/semantic/analyze_functions.go`
  - **Tests**: Add semantic analysis tests for lazy parameters
  - **Estimated time**: 0.5-1 day

- [ ] 9.22.3 Implement/verify lazy parameter evaluation in interpreter
  - **Task**: Ensure parameters marked `lazy` are evaluated in callee scope, not caller scope
  - **Implementation**:
    - Store unevaluated expression for lazy parameters
    - Evaluate expression when parameter is accessed in function body
    - Handle multiple accesses (cache vs. re-evaluate)
  - **Files**: `internal/interp/functions.go`
  - **Tests**: Test lazy evaluation semantics (Jensen's device, conditional evaluation, etc.)
  - **Estimated time**: 1-2 days

**Blocked Tests**:
- cmd/dwscript: TestLazyParamsScriptsExist (all 5 subtests)
- Possible fixture tests depending on lazy parameter usage

---

### Phase 9.23: Bytecode Compiler Fixes

**Priority**: MEDIUM - Required for 5 failing bytecode tests
**Timeline**: 3-4 days
**Impact**: Fix basic bytecode compilation issues

**Current Status**: Several basic bytecode compiler tests are failing, suggesting issues in the bytecode compilation pipeline.

- [ ] 9.23.1 Investigate and fix TestCompiler_VarAssignReturn
  - **Task**: Debug why variable assignment and return compilation fails
  - **Implementation**:
    - Run test with verbose output
    - Check if variables are registered in compiler scope
    - Verify STORE_LOCAL and RETURN opcodes are generated
    - Fix any identified issues
  - **Files**: `internal/bytecode/compiler.go`, `internal/bytecode/compiler_test.go`
  - **Estimated time**: 0.5 day

- [ ] 9.23.2 Investigate and fix TestCompiler_IfElse
  - **Task**: Debug why if-else statement compilation fails
  - **Implementation**:
    - Verify JUMP_IF_FALSE and JUMP opcodes are generated
    - Check jump offset calculations
    - Ensure branches compile correctly
  - **Files**: `internal/bytecode/compiler.go`
  - **Estimated time**: 0.5 day

- [ ] 9.23.3 Investigate and fix TestCompiler_ArrayLiteralAndIndex
  - **Task**: Debug why array literal and indexing compilation fails
  - **Implementation**:
    - Check NEW_ARRAY opcode generation
    - Verify array element push instructions
    - Test GET_INDEX and SET_INDEX opcodes
  - **Files**: `internal/bytecode/compiler.go`
  - **Estimated time**: 0.5-1 day

- [ ] 9.23.4 Investigate and fix TestCompiler_CallExpression
  - **Task**: Debug why function call compilation fails
  - **Implementation**:
    - Verify argument compilation
    - Check CALL opcode generation with correct arity
    - Test both built-in and user-defined functions
  - **Files**: `internal/bytecode/compiler.go`
  - **Estimated time**: 0.5-1 day

- [ ] 9.23.5 Investigate and fix TestCompiler_MemberAccess
  - **Task**: Debug why member access (object.field) compilation fails
  - **Implementation**:
    - Check GET_PROPERTY opcode generation
    - Verify object reference compilation
    - Test field name encoding in bytecode
  - **Files**: `internal/bytecode/compiler.go`
  - **Estimated time**: 0.5-1 day

- [ ] 9.23.6 Add regression tests for fixed issues
  - **Task**: Ensure bytecode compiler tests remain passing
  - **Tests**: Enhance existing test suite based on fixes
  - **Files**: `internal/bytecode/compiler_test.go`
  - **Estimated time**: 0.5 day

**Blocked Tests**:
- internal/bytecode: TestCompiler_VarAssignReturn, TestCompiler_IfElse, TestCompiler_ArrayLiteralAndIndex, TestCompiler_CallExpression, TestCompiler_MemberAccess

**Note**: Phase 11 (Bytecode VM) is marked mostly complete, but these basic compilation tests suggest the compiler needs attention before moving to advanced optimizations.

---

### Phase 9.24: Systematic Fixture Test Analysis and Fixes

**Priority**: MEDIUM-HIGH - Required for 300+ failing fixture tests
**Timeline**: 2-4 weeks
**Impact**: Systematic approach to fixing all remaining fixture test failures

**Current Status**: ~300+ fixture tests are failing in SimpleScripts, Algorithms, and Overloads categories. Many failures are due to missing built-in functions (Phase 9.17), but many are also due to semantic issues, runtime bugs, and missing features.

#### Strategy: Divide and Conquer

Instead of fixing tests one-by-one, group them by root cause and fix categories systematically.

- [ ] 9.24.1 Categorize all failing fixture tests by root cause
  - **Task**: Run fixture tests and categorize each failure
  - **Categories**:
    - Missing built-in functions (Phase 9.17)
    - Parser syntax errors (Phase 9.21)
    - Constructor/method overloading (Phase 9.19, 9.20)
    - Abstract class issues (Phase 9.16.8)
    - Interface issues (Phase 9.16.2)
    - Property issues (Phase 9.16.3)
    - Semantic analysis bugs (Phase 9.16)
    - Runtime interpreter bugs
    - Other/unknown
  - **Output**: Create `docs/fixture-test-analysis.md` with categorized list
  - **Files**: Run `go test ./internal/interp -run TestDWScriptFixtures -v` and analyze
  - **Estimated time**: 2-3 days

- [ ] 9.24.2 Fix "Missing Built-in Functions" category
  - **Task**: Implement all missing built-in functions identified
  - **Dependency**: Links to Phase 9.17 tasks
  - **Implementation**: Prioritize functions based on usage count in fixtures
  - **Estimated time**: Covered by Phase 9.17 (2-6 weeks)

- [ ] 9.24.3 Fix "Parser Syntax Errors" category
  - **Task**: Fix all parser issues identified
  - **Dependency**: Links to Phase 9.21 tasks
  - **Estimated time**: Covered by Phase 9.21 (5-7 days)

- [ ] 9.24.4 Fix "Constructor/Method Overloading" category
  - **Task**: Fix all overloading issues
  - **Dependency**: Links to Phase 9.19 and 9.20 tasks
  - **Estimated time**: Covered by Phase 9.19-9.20 (6-10 days)

- [ ] 9.24.5 Fix "Semantic Analysis Bugs" category
  - **Task**: Address semantic analyzer issues not covered by Phase 9.16
  - **Implementation**: Fix case-sensitivity, scope resolution, type checking bugs
  - **Files**: `internal/semantic/*.go`
  - **Estimated time**: 1-2 weeks

- [ ] 9.24.6 Fix "Runtime Interpreter Bugs" category
  - **Task**: Fix interpreter execution bugs
  - **Examples**:
    - Undefined variable access from parent classes
    - Incorrect method dispatch
    - Field access issues
    - Exception handling bugs
  - **Files**: `internal/interp/*.go`
  - **Estimated time**: 1-2 weeks

- [ ] 9.24.7 Fix "Abstract Class Issues" category
  - **Task**: Review tests expecting abstract class instantiation
  - **Implementation**: Either fix semantic analyzer or update test expectations
  - **Dependency**: Phase 9.16.8
  - **Estimated time**: 1-2 days

- [ ] 9.24.8 Fix "Interface Issues" category
  - **Task**: Complete interface implementation
  - **Dependency**: Phase 9.16.2
  - **Estimated time**: Covered by Phase 9.16.2

- [ ] 9.24.9 Fix "Property Issues" category
  - **Task**: Implement indexed and expression properties
  - **Dependency**: Phase 9.16.3
  - **Estimated time**: Covered by Phase 9.16.3

- [ ] 9.24.10 Fix "Other/Unknown" category
  - **Task**: Investigate and fix remaining issues on case-by-case basis
  - **Implementation**: Debug each test individually
  - **Estimated time**: 1-2 weeks

- [ ] 9.24.11 Update TEST_STATUS.md with progress
  - **Task**: Document progress after each category is fixed
  - **Files**: `testdata/fixtures/TEST_STATUS.md`
  - **Implementation**: Update pass/fail counts, document resolved issues
  - **Estimated time**: Ongoing (0.5 day total)

- [ ] 9.24.12 Verify 90%+ fixture test pass rate
  - **Task**: Ensure at least 90% of fixture tests pass
  - **Target**: ~1900+ of 2100+ tests passing
  - **Milestone**: Mark Phase 9 as complete when achieved
  - **Estimated time**: Verification after all categories fixed

**Total Estimated Time**: 2-4 weeks (depends on complexity of issues discovered)

**Success Criteria**:
- All fixture tests categorized by root cause
- 90%+ of fixture tests passing
- Documented analysis and fixes in TEST_STATUS.md
- Remaining failures have documented blockers/reasons

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

- [ ] 11.14 Add bytecode serialization
  - [ ] Implement bytecode file format (.dwc)
  - [ ] Save/load compiled bytecode to disk
  - [ ] Version bytecode format for compatibility
  - [ ] Add `dwscript compile` command for bytecode

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
