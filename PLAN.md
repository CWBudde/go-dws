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

- [x] 9.1 Type Operators (is/as/implements) - COMPLETED
  - **Estimate**: 8-10 hours
  - **Description**: Implement type checking and casting operators
  - **Strategy**: Add type operator support in parser and semantic analyzer
  - **Complexity**: Requires runtime type information and safe casting mechanisms
  - **Status**: COMPLETED. All tests passing (24/24 = 100%)
  - **Test Results**: 30/30 tests passing (100% pass rate)
  - **Completed Subtasks**:
    - [x] 9.1.1 Fix 'as' operator to support class-to-class casting
      - Semantic analyzer now supports both class and interface target types
      - File: internal/semantic/analyze_expressions.go (analyzeAsExpression)
      - Validates upcast/downcast relationships in class hierarchy
    - [x] 9.1.2 Add validation for 'is' operator operands
      - Left operand validated as class instance or nil
      - Right operand validated as class type
      - File: internal/semantic/analyze_expressions.go (analyzeIsExpression)
    - [x] 9.1.3 Add validation for 'implements' operator operands
      - Left operand validated as class instance or nil
      - Right operand validated as interface type
      - File: internal/semantic/analyze_expressions.go (analyzeImplementsExpression)
    - [x] 9.1.4 Update interpreter 'as' operator for class casting
      - File: internal/interp/expressions.go (evalAsExpression)
      - Runtime now supports both class-to-class and class-to-interface casts
      - Validates runtime compatibility for downcasts
    - [x] 9.1.5 Verify all type operator tests pass - ALL PASSING
    - [ ] 9.1.6 Avoid cascading errors when 'as' target type is invalid
      - Short-circuit analysis after reporting "'as' operator requires interface or class type"
      - Prevents secondary `cannot infer type` diagnostics (TestTypeOperator_As_InvalidRightOperand)
  - **Files Modified**:
    - internal/semantic/analyze_expressions.go (added strings import, updated all 3 operators)
    - internal/semantic/type_operators_test.go (updated error message expectation)
    - internal/interp/expressions.go (evalAsExpression now handles classes)

- [ ] 9.2 Abstract Class Implementation (1 test)
  - **Estimate**: 2-3 hours
  - **Description**: Validate that abstract classes cannot be instantiated
  - **Strategy**: Add abstract class tracking and validation in class instantiation
  - **Complexity**: Requires inheritance chain validation
  - **Subtasks**:
    - [ ] 9.2.1 Clear abstract flags when overrides are implemented
      - Ensure overriding inherited abstract methods removes the abstract marker
      - Fixes `TestValidAbstractImplementation`

- [ ] 9.3 Miscellaneous High Complexity Fixes (18 tests)
  - **Estimate**: 10-15 hours
  - **Description**: Various complex semantic validation issues
  - **Strategy**: Analyze each test individually and implement targeted fixes
  - **Examples**: Generic types, delegates, advanced inheritance scenarios, complex type checking

- [ ] 9.4 Fix Class Forward Declarations in Units

**Goal**: Support class forward declarations in unit interface sections for cross-referencing types.

**Estimate**: 4-8 hours (0.5-1 day)

**Status**: NOT STARTED

**Blocked Tests** (1+ tests):
- `testdata/fixtures/SimpleScripts/class_scoping1.pas`
- Potentially other unit-based tests with forward references

**Current Errors**:
- `no prefix parse function for CLASS` - Parser doesn't recognize class forward declaration syntax
- `expected DOT after 'end' in unit` - Parser gets confused after seeing incomplete class declaration

**DWScript Syntax**:
```pascal
unit MyUnit;

interface

type
  TForward = class;  // Forward declaration

  TActual = class
    FNext: TForward;  // Can reference forward-declared class
  end;

  TForward = class    // Full definition later
    FPrev: TActual;
  end;
```

**Root Cause**: Unit parser doesn't handle the forward declaration syntax `TName = class;` (class keyword without implementation block). When it sees `class;`, it expects either class body or inheritance, not immediate semicolon.

**Implementation**:
- Files: `internal/parser/parser_unit.go`, `internal/parser/parser_types.go`
- Parser needs to detect `class;` pattern (empty class declaration)
- Create forward reference placeholder in symbol table
- Resolve forward references when full definition appears
- Semantic analyzer validates all forwards are resolved

**Subtasks**:
- [ ] 9.4.1 Extend class type parsing to detect forward declarations
  - Recognize `TName = class;` syntax (no parent, no members, just semicolon)
  - Distinguish from regular empty class `TName = class end;`
- [ ] 9.4.2 Create forward reference tracking in parser/analyzer
  - Store forward-declared class names
  - Mark type as "forward" until full definition seen
- [ ] 9.4.3 Update symbol table to support forward references
  - Allow type to be registered twice (forward + full definition)
  - Second registration replaces forward with full type
- [ ] 9.4.4 Add semantic validation for unresolved forwards
  - Error if forward-declared class never gets full definition
  - Error if type used before forward declaration or definition
- [ ] 9.4.5 Test with unit files containing cross-references
  - Two classes referencing each other
  - Multiple forward declarations

**Acceptance Criteria**:
- `class_scoping1.pas` test passes
- Forward declarations work: `TName = class;`
- Cross-referencing classes work (A references B, B references A)
- Proper error for unresolved forward declarations
- Works in both unit interface and implementation sections

- [ ] 9.5 Support Field Initializers in Type Declarations

**Goal**: Allow field initialization syntax in record and class type declarations.

**Estimate**: 6-8 hours (1 day)

**Status**: NOT STARTED

**Blocked Tests** (3 tests):
- `testdata/fixtures/SimpleScripts/clear_ref_in_destructor.pas`
- `testdata/fixtures/SimpleScripts/clear_ref_in_static_method.pas`
- `testdata/fixtures/SimpleScripts/clear_ref_in_virtual_method.pas`

**Current Error**: `expected SEMICOLON, got EQ` when parser encounters `:=` after field declaration

**DWScript Syntax**:
```pascal
type
  TRecord = record
    Count: Integer := 0;        // Field with default value
    Name: String := 'Default';
  end;

  TClass = class
    FValue: Integer := 42;      // Field initializer
    FItems: array of String;    // No initializer (nil/empty)
  end;
```

**Root Cause**: Parser expects field declaration format `fieldName: Type;` and doesn't handle the optional initializer `fieldName: Type := value;`. The `:=` token triggers a parse error because parser has already moved past the field type.

**Current Support**:
- ✅ Field declarations: `FValue: Integer;`
- ✅ Variable initialization: `var x: Integer := 5;`
- ❌ Field initialization: `FValue: Integer := 5;` in type declarations

**Implementation**:
- Extend parser to accept optional `:= expression` after field type
- Store initializer expression in AST field node
- Semantic analyzer validates initializer expression type matches field type
- Interpreter executes initializers during record/class instantiation
- Files: `internal/parser/parser_types.go`, `internal/parser/parser_class.go`, `internal/ast/ast.go`

**Subtasks**:
- [ ] 9.5.1 Update AST to store field initializers
  - Add `Initializer` field to `FieldDeclaration` AST node
  - Store expression for default value
- [ ] 9.5.2 Extend field parsing to accept initializers
  - After parsing field type, check for ASSIGN token (`:=`)
  - If present, parse initializer expression
  - Works for both record and class fields
- [ ] 9.5.3 Add semantic validation for field initializers
  - Type check: initializer expression must match field type
  - Const check: initializer must be a compile-time constant or simple expression
  - No forward references: can't reference fields declared later
- [ ] 9.5.4 Implement runtime initialization in interpreter
  - For records: initialize fields when record created
  - For classes: initialize fields in constructor or during `new`
  - Execute initializers in declaration order
- [ ] 9.5.5 Add bytecode VM support
  - Compile field initializers to bytecode
  - Execute during object instantiation

**Acceptance Criteria**:
- All 3 blocked tests pass
- Field initializers work: `FValue: Integer := 42;`
- Works for both records and classes
- Initializers executed during instantiation
- Type checking validates initializer matches field type
- Works in both AST interpreter and bytecode VM

- [ ] 9.6 Fix Remaining Parser Errors (Fixture Tests)

**Goal**: Address remaining parser errors discovered during fixture test runs to unblock semantic and runtime testing.

**Estimate**: 8-16 hours (1-2 days)

**Status**: NOT STARTED

**Impact**: Fixes 100+ parser-related fixture test failures across multiple test categories

**Common Parser Error Categories**:

1. **Missing Statement/Expression Types**
   - Errors: `no prefix parse function for TOKEN`
   - Examples: Missing operator support, unknown keywords

2. **Incomplete Syntax Support**
   - Errors: `expected TOKEN, got OTHER_TOKEN`
   - Examples: Partial implementation of language features

3. **Unit/Module Parsing Issues**
   - Errors: `unexpected token in unit interface/implementation`
   - Examples: Missing unit-specific syntax handling

4. **Type Declaration Gaps**
   - Errors: `expected SEMICOLON/END in type declaration`
   - Examples: Advanced type syntax not supported

5. **Control Flow Parsing**
   - Errors: `unexpected token in statement`
   - Examples: Case statement variants, specialized loops

**Approach**: Iterative investigation and fixing:
- Run fixture test suite to collect all parser errors
- Categorize errors by root cause
- Fix highest-impact errors first (blocking most tests)
- Re-run tests after each fix to verify progress
- Document each fix for reference

**Subtasks**:
- [ ] 9.6.1 Audit all fixture test parser failures
  - Run full fixture test suite: `go test -v ./internal/interp -run TestDWScriptFixtures`
  - Collect all parser error messages
  - Categorize by error type and affected feature
  - Prioritize by number of blocked tests

- [ ] 9.6.2 Fix high-priority parser errors (20+ tests each)
  - Address errors blocking the most tests first
  - Update parser to handle missing syntax
  - Add AST nodes if needed

- [ ] 9.6.3 Fix medium-priority parser errors (5-20 tests each)
  - Work through medium-impact errors
  - May require new token types or parse functions

- [ ] 9.6.4 Fix low-priority parser errors (1-5 tests each)
  - Handle edge cases and rare syntax forms
  - Complete language coverage

- [ ] 9.6.5 Document parser fixes and patterns
  - Update PLAN.md with specific fixes made
  - Note any DWScript quirks discovered
  - Update test status tracking

**Files Likely to Update**:
- `internal/parser/parser.go` - Core parsing logic
- `internal/parser/expressions.go` - Expression parsing
- `internal/parser/statements.go` - Statement parsing
- `internal/parser/parser_types.go` - Type declarations
- `internal/parser/parser_class.go` - Class/OOP features
- `internal/parser/parser_unit.go` - Unit/module support
- `internal/lexer/token_type.go` - Token definitions (if new keywords needed)
- `internal/ast/*.go` - AST nodes (if new node types needed)

**Acceptance Criteria**:
- All parser errors resolved (tests may still fail on semantic/runtime issues)
- Parser successfully parses all valid DWScript fixture test files
- No "unexpected token" or "no prefix parse function" errors remain
- Parser coverage increases to support full DWScript syntax
- Reduced fixture test failures from 100+ to primarily semantic/runtime issues

---

## Task 9.7: Complete Static Record Methods (Class Functions) Implementation ⚠️ IN PROGRESS

**Goal**: Finish implementing static methods (class functions) on record types for full DWScript compatibility.

**Estimate**: 3-4 hours (semantic analysis done, runtime remaining)

**Status**: IN PROGRESS - Semantic analysis ✅ COMPLETE | Runtime execution ⚠️ INCOMPLETE

**Blocked Tests** (2+ tests):
- `testdata/fixtures/SimpleScripts/record_method_static.pas` - Primary test case
- `testdata/fixtures/Algorithms/lerp.pas` - Record with instance method (related issue)

**Current Error**: `ERROR: class 'TTest' not found` (runtime error, not semantic)

**DWScript Syntax**:
```pascal
type
  TTest = record
    Value: Integer;

    // Static method (class function) - called on type
    class function Sum(A, B: Integer): Integer;

    // Instance method - called on instance
    procedure Print;
  end;

// Static method implementation
class function TTest.Sum(A, B: Integer): Integer;
begin
  Result := A + B;
end;

// Usage
var x := TTest.Sum(5, 7);  // Static call on type (not instance)
```

**Progress Summary**:

**✅ Completed (Semantic Analysis)**:
1. ✅ Updated `RecordType` struct with `ClassMethods` and `ClassMethodOverloads` maps
2. ✅ Fixed record type registration order (register type before analyzing methods)
3. ✅ Semantic analyzer tracks static vs instance methods separately
4. ✅ Type-level member access: `TTest.Sum` resolves to class method
5. ✅ Record method implementation support in semantic analyzer
6. ✅ Added `currentRecord` context to Analyzer for method resolution
7. ✅ Bare function calls inside record methods resolve to class methods
8. ✅ All semantic analysis passes for `record_method_static.pas`

**❌ Remaining (Runtime Execution)**:

**Root Cause**: Interpreter runtime only looks for methods in `i.classes` map. When it sees `TTest.Sum(...)`, it searches for a class named "TTest" but doesn't check `i.records` map, causing "class not found" error even though semantic analysis succeeded.

**Implementation**:
- Files: `internal/interp/expressions.go`, `internal/bytecode/compiler.go`, `internal/bytecode/vm.go`
- Fix method call evaluation to check both classes AND records
- Fix new expression evaluation to check both classes AND records
- Add bytecode support for static record method calls

**Subtasks**:
- [ ] 9.7.1 Fix AST interpreter method call evaluation
  - Update `evalMethodCallExpression` in `expressions.go`
  - When resolving type-level method calls (e.g., `TType.Method(...)`):
    - First check `i.classes` map (existing behavior)
    - If not found, check `i.records` map
    - Look up class method in record's `ClassMethods` or `ClassMethodOverloads`
  - Handle overload resolution for static record methods

- [ ] 9.7.2 Fix AST interpreter new expression evaluation
  - Update `evalNewExpression` in `expressions.go`
  - Records can have static factory methods (e.g., `TTest.Create`)
  - Check both `i.classes` and `i.records` when resolving type name

- [ ] 9.7.3 Fix record instance method lookup (lerp.pas)
  - Update method call evaluation for record instances
  - When calling method on record value: `recordVar.Method()`
  - Look up method in record type's instance methods
  - This may be a separate bug from static methods

- [ ] 9.7.4 Add bytecode compiler support
  - Update `internal/bytecode/compiler.go`
  - Compile static record method calls to bytecode
  - Use same opcode as class method calls (CALL_METHOD or similar)
  - Store record type info in bytecode for VM dispatch

- [ ] 9.7.5 Add bytecode VM support
  - Update `internal/bytecode/vm.go`
  - VM dispatch for method calls checks both classes and records
  - Execute static record methods correctly
  - Handle overload resolution in VM

- [ ] 9.7.6 Test with multiple scenarios
  - Static method calls: `TRecord.StaticMethod(...)`
  - Instance method calls: `recordVar.InstanceMethod()`
  - Overloaded static methods
  - Factory methods returning record instances
  - Both AST interpreter and bytecode VM

**Files to Update**:
- `internal/interp/expressions.go` - Method call and new expression evaluation (primary fix)
- `internal/interp/record.go` - Record value method lookup if needed
- `internal/bytecode/compiler.go` - Bytecode generation for static record methods
- `internal/bytecode/vm.go` - VM execution of record class method calls

**Acceptance Criteria**:
- `record_method_static.pas` test passes completely (output: `12`)
- `lerp.pas` test passes (record instance method works)
- Static method calls on records work: `TRecord.Method(...)`
- Instance method calls on records work: `recordVar.Method()`
- Overloaded static methods resolve correctly at runtime
- Record factory methods work: `TRecord.Create(...)`
- Works in both AST interpreter and bytecode VM
- No regression in existing class method functionality

---

## Task 9.8: Array Helper Methods (Algorithms Fixtures) 🎯 HIGH PRIORITY

**Goal**: Implement missing array helper methods and properties to fix 4 failing Algorithms tests.

**Estimate**: 3-4 hours

**Status**: NOT STARTED

**Blocked Tests** (4 tests):
- `testdata/fixtures/Algorithms/gnome_sort.pas` - needs `Swap(i, j)`
- `testdata/fixtures/Algorithms/maze_generation.pas` - needs `Push(value)` and `Pop()`
- `testdata/fixtures/Algorithms/one_dim_automata.pas` - needs `.low` and `.high` properties
- `testdata/fixtures/Algorithms/quicksort_dyn.pas` - needs `Swap(i, j)`

**Missing Features**:
1. `Swap(i, j)` method on arrays - swaps elements at indices i and j
2. `Push(value)` method on dynamic arrays - appends element (alias for Add)
3. `Pop()` method on dynamic arrays - removes and returns last element
4. `.low` and `.high` properties on arrays - return Low(arr) and High(arr) values

**Implementation**:
- Add methods to array type helper registration
- Files: `internal/interp/helpers_validation.go`, `internal/interp/helpers_conversion.go`
- Infrastructure already exists, just need to add new methods

**Subtasks**:
- [ ] 9.8.1 Add `Swap(i, j)` method to array helper
  - Validate indices are within bounds
  - Swap elements at positions i and j
- [ ] 9.8.2 Add `Push(value)` method to dynamic array helper
  - Implement as alias for existing `Add` method
- [ ] 9.8.3 Add `Pop()` method to dynamic array helper
  - Remove last element and return it
  - Error if array is empty
- [ ] 9.8.4 Add `.low` and `.high` properties to array helper
  - Return Low(arr) and High(arr) respectively
  - Work for both static and dynamic arrays

**Acceptance Criteria**:
- All 4 blocked Algorithms tests pass
- Methods work in both AST interpreter and bytecode VM
- Proper error handling for out-of-bounds and empty arrays

---

## Task 9.9: Fix Inc/Dec on Array Elements (Algorithms Fixtures) 🎯 HIGH PRIORITY

**Goal**: Allow Inc/Dec built-in functions to work on array element expressions.

**Estimate**: 1-2 hours

**Status**: NOT STARTED

**Blocked Tests** (1 test):
- `testdata/fixtures/Algorithms/evenly_divisible.pas` - uses `Inc(Result[i])` and `Inc(Result[n])`

**Current Error**: `function 'Inc' first argument must be a variable`

**Root Cause**: Semantic analyzer validation is too strict - it rejects indexed array access (`arr[i]`) as a valid lvalue for Inc/Dec, even though array indexing is a valid assignable expression.

**Implementation**:
- File: `internal/semantic/analyzer.go`
- Relax validation in Inc/Dec analysis to accept IndexExpression as valid lvalue
- Inc() and Dec() runtime functions already work correctly, this is purely a validation issue

**Subtasks**:
- [ ] 9.9.1 Update Inc/Dec semantic validation to accept IndexExpression
  - Check that IndexExpression base is a variable (array)
  - Allow Inc(arr[i]) where arr is a variable
- [ ] 9.9.2 Verify test passes with relaxed validation

**Acceptance Criteria**:
- `evenly_divisible.pas` test passes
- Inc/Dec work on array elements: `Inc(arr[i])`, `Dec(arr[x+1])`
- Proper error messages for invalid cases (e.g., `Inc(5)`)

---

## Task 9.10: Const Expression Evaluator (Algorithms Fixtures)

**Goal**: Implement compile-time evaluation of const expressions including string operations and character literals.

**Estimate**: 6-8 hours

**Status**: NOT STARTED

**Blocked Tests** (3 tests):
- `testdata/fixtures/Algorithms/bottles_of_beer.pas` - `const CRLF : String = '' + #13#10;`
- `testdata/fixtures/Algorithms/sparse_matmult.pas` - const expressions with Ord() and Chr()
- `testdata/fixtures/Algorithms/vigenere.pas` - function-local const declarations

**Current Limitations**:
1. Cannot use string concatenation in const expressions (`'' + #13#10`)
2. Cannot use character literals (`#13`, `#10`) in const context
3. Cannot declare const variables inside function bodies with expressions
4. Cannot use Ord/Chr functions in const context

**Implementation**:
- Create new `internal/semantic/const_evaluator.go` for compile-time evaluation
- Extend `internal/semantic/statements_declarations.go` const handling
- Support:
  - String concatenation: `'a' + 'b'` → `'ab'`
  - Character literals: `#13` → `'\r'`, `#10` → `'\n'`
  - Numeric character literals: `#65` → `'A'`
  - Function-local const declarations
  - Ord/Chr in const context

**Subtasks**:
- [ ] 9.10.1 Create const expression evaluator
  - Add `const_evaluator.go` with evaluation logic
  - Support literals, binary ops (+, -, *, /), unary ops
- [ ] 9.10.2 Support string concatenation in const expressions
  - Evaluate `'str1' + 'str2'` at compile time
- [ ] 9.10.3 Support character literals in const expressions
  - Parse and evaluate `#N` to character value
  - Support both decimal and hex forms
- [ ] 9.10.4 Support function-local const declarations
  - Allow const declarations inside function/procedure bodies
  - Evaluate expressions at semantic analysis time
- [ ] 9.10.5 Support Ord/Chr in const context
  - Evaluate Ord('A') → 65 at compile time
  - Evaluate Chr(65) → 'A' at compile time

**Acceptance Criteria**:
- All 3 blocked tests pass
- Const expressions evaluated at compile time (not runtime)
- Proper error messages for non-const expressions in const context

---

## Task 9.11: Multi-dimensional Array Creation (Algorithms Fixtures)

**Goal**: Support multi-dimensional array allocation with `new Type[dim1, dim2, ...]` syntax.

**Estimate**: 4-6 hours

**Status**: NOT STARTED

**Blocked Tests** (1 test):
- `testdata/fixtures/Algorithms/lu_factorization.pas` - uses `new Float[M, N]` for 2D arrays

**Current Error**: `function 'new' expects 2 arguments, got 3`

**Root Cause**: Parser and runtime only support single-dimension `new Type[size]`, not multi-dimensional `new Type[d1, d2, ...]`.

**Current Support**:
- ✅ Multi-dimensional array types: `array of array of Float`
- ✅ Multi-dimensional indexing: `arr[i][j]` or `arr[i, j]`
- ❌ Multi-dimensional allocation: `new Float[M, N]`

**Implementation**:
- Extend parser to accept multiple dimensions in NewExpression
- Files: `internal/parser/expressions.go`, `internal/interp/array.go`
- Create nested arrays: `new Float[M, N]` → array of M elements, each is array of N floats

**Subtasks**:
- [ ] 9.11.1 Extend parser to accept multiple dimensions
  - Parse `new Type[expr1, expr2, ...]` syntax
  - Store dimension expressions in AST
- [ ] 9.11.2 Update semantic analyzer for multi-dim new
  - Validate dimension count matches array type dimensions
  - Type check dimension expressions (must be integers)
- [ ] 9.11.3 Implement multi-dimensional allocation in interpreter
  - Create nested arrays recursively
  - `new T[d1, d2]` creates array of d1 elements, each is array of d2 elements
- [ ] 9.11.4 Update bytecode compiler and VM
  - Add bytecode support for multi-dim allocation

**Acceptance Criteria**:
- `lu_factorization.pas` test passes
- `new Type[d1, d2]` creates properly nested arrays
- Works in both AST interpreter and bytecode VM
- Proper error for dimension count mismatch

---

## Task 9.12: SetLength on String Type (Algorithms Fixtures)

**Goal**: Extend SetLength built-in function to support String type in addition to arrays.

**Estimate**: 2-3 hours

**Status**: NOT STARTED

**Blocked Tests** (1 test):
- `testdata/fixtures/Algorithms/sparse_matmult.pas` - uses `SetLength(s, n)` on string variables

**Current Error**: `SetLength expects array as first argument, got String`

**DWScript Compatibility**: In DWScript, SetLength works on both arrays and strings:
- Arrays: `SetLength(arr, newSize)` - resizes dynamic array
- Strings: `SetLength(str, newLen)` - truncates or extends string with spaces

**Implementation**:
- File: `internal/interp/builtins_misc.go`
- Modify `builtinSetLength` to accept String type
- For strings: truncate if shorter, pad with spaces if longer

**Subtasks**:
- [ ] 9.12.1 Update SetLength validation to accept String
  - Check first argument is Array OR String
- [ ] 9.12.2 Implement string resizing logic
  - If new length < current: truncate string
  - If new length > current: pad with spaces
- [ ] 9.12.3 Add Float.ToString(precision) helper method
  - Also needed by sparse_matmult.pas
  - Format float with specified decimal places

**Acceptance Criteria**:
- `sparse_matmult.pas` test passes
- SetLength works on strings: `SetLength(s, 10)` sets string length to 10
- Strings truncated or space-padded as needed

---

## Task 9.13: Debug extract_ranges Logic (Algorithms Fixtures)

**Goal**: Fix off-by-one error in extract_ranges.pas test causing incorrect output.

**Estimate**: 2-3 hours

**Status**: NOT STARTED

**Blocked Tests** (1 test):
- `testdata/fixtures/Algorithms/extract_ranges.pas`

**Expected Output**: `0-2,4,6-8,11,12,14-25,27-33,35-39`
**Actual Output**: `0-2,4,6-8,11,12,14-25,27-33,35`

**Issue**: Last range `35-39` is being truncated to just `35`, suggesting elements 36-39 are not being processed.

**Root Cause**: This is NOT a missing language feature - it's a runtime logic bug in how the interpreter handles array iteration or loop boundaries in this specific test case.

**Investigation Steps**:
- [ ] 9.13.1 Analyze extract_ranges.pas algorithm
  - Understand the range extraction logic
  - Identify loop boundaries and termination conditions
- [ ] 9.13.2 Debug with trace output
  - Add logging to understand where iteration stops
  - Check array length, High(arr), loop conditions
- [ ] 9.13.3 Identify and fix the bug
  - Could be in for-loop handling, array bounds, or conditional logic
  - Verify fix doesn't break other tests

**Acceptance Criteria**:
- `extract_ranges.pas` produces correct output including `35-39`
- Root cause identified and documented
- No regressions in other Algorithms tests

---

## Task 9.14: Investigate aes_encryption Exception (Algorithms Fixtures)

**Goal**: Determine if aes_encryption.pas failure is a real bug or test framework issue.

**Estimate**: 1-2 hours

**Status**: NOT STARTED

**Test**: `testdata/fixtures/Algorithms/aes_encryption.pas`

**Current Behavior**: Test fails with "uncaught exception: Exception: Invalid AES key"

**Possible Causes**:
1. **Test is intentionally raising an exception** - The script has try/except blocks and may be testing exception handling. The fixture framework might not recognize exception output correctly.
2. **Actual bug in exception handling** - Exception is raised but not being caught properly.
3. **Missing AES-specific features** - Though unlikely, some AES operation might not be implemented.

**Investigation Steps**:
- [ ] 9.14.1 Read aes_encryption.pas source code
  - Understand what the test is doing
  - Check if exception is expected behavior
- [ ] 9.14.2 Check expected output file
  - Does expected output include exception text?
  - Is this a normal vs exceptional path test?
- [ ] 9.14.3 Test exception handling manually
  - Run test with `./bin/dwscript run testdata/fixtures/Algorithms/aes_encryption.pas`
  - Compare output to expected
- [ ] 9.14.4 Fix if needed, or mark as test framework limitation

**Acceptance Criteria**:
- Root cause identified and documented
- If interpreter bug: fixed and test passes
- If test framework issue: document workaround or defer

---

## Task 9.15: Static vs Dynamic Array Compatibility (DEFERRED)

**Goal**: Investigate type compatibility between static and dynamic arrays in var parameters.

**Status**: DEFERRED - May be test issue, not implementation issue

**Blocked Tests** (1 test):
- `testdata/fixtures/Algorithms/quicksort.pas`

**Current Error**: `cannot assign TData to TData` when passing static array to var parameter expecting dynamic array

**Issue**:
- Test defines: `type TData = array [0..size-1] of integer;` (static array)
- Procedure expects: `procedure QuickSort(var A: TData; ...)`
- DWScript type system treats static and dynamic arrays as incompatible in var parameters

**Investigation Needed**:
- Is this correct DWScript behavior?
- Should static arrays be convertible to dynamic in var params?
- Or is the test incorrectly written?

**Deferred Because**:
- May require significant type system changes
- Only affects 1 test
- Need to verify against original DWScript behavior
- Low priority compared to other fixes

**Future Action**:
- Research DWScript documentation on static/dynamic array compatibility
- Check original DWScript source for handling of this case
- If needed, implement proper coercion rules

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
