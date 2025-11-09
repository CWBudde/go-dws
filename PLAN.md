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

- [ ] 9.2 Enforce private field access control
  - **Issue**: Derived classes can access private parent fields
  - **Implementation**: Add visibility check in field access
  - **Files**: `internal/semantic/analyze_expressions.go`
  - **Estimated time**: 0.5 day

- [x] 9.3.1 Parse `class procedure` and `class function` declarations
  - **Task**: Extend parser to recognize class method syntax
  - **Syntax**: `class procedure Name;` and `class function Name: ReturnType;`
  - **Files**: `internal/parser/functions.go`, `internal/ast/declarations.go`
  - **Tests**: Parse class method declarations without errors
  - **Status**: ✅ Already implemented and tested

- [x] 9.3.2 Implement class method semantic analysis
  - **Task**: Validate class methods (no access to Self/instance fields)
  - **Files**: `internal/semantic/analyze_functions.go`
  - **Tests**: Class methods cannot access instance members
  - **Status**: ✅ Already implemented and tested

- [x] 9.3.3 Implement class method runtime execution
  - **Task**: Execute class methods without instance context
  - **Files**: `internal/interp/functions.go`
  - **Tests**: Class methods execute and access class variables/constants
  - **Status**: ✅ Already implemented and tested

- [x] 9.3.4 Support virtual/override on class methods
  - **Task**: Allow polymorphism for class methods
  - **Files**: `internal/semantic/analyze_functions.go`, `internal/interp/objects.go`
  - **Tests**: Virtual class methods dispatch correctly through inheritance
  - **Status**: ✅ Already implemented and tested

#### 9.4 Class Constants - MEDIUM PRIORITY

**Blocks**: class_const*.pas (5+ tests)

- [x] 9.4.1 Parse class constant declarations
  - **Task**: Recognize `const` inside class declarations
  - **Syntax**: `const cName = Value;` or `class const cName = Value;`
  - **Files**: `internal/parser/interfaces.go`
  - **Tests**: Parse class constants in class body
  - **Status**: ✅ Already implemented and tested

- [x] 9.4.2 Store class constants in ClassType
  - **Task**: Add Constants map to ClassType
  - **Files**: `internal/types/types.go`
  - **Tests**: Class constants accessible in type system
  - **Status**: ✅ Already implemented and tested

- [x] 9.4.3 Semantic analysis for class constants
  - **Task**: Validate class constant access (ClassName.ConstName)
  - **Files**: `internal/semantic/analyze_classes.go`
  - **Tests**: Class constants resolve correctly
  - **Status**: ✅ Already implemented and tested

- [x] 9.4.4 Runtime class constant evaluation
  - **Task**: Access class constants at runtime
  - **Files**: `internal/interp/objects.go`
  - **Tests**: Class constants return correct values
  - **Status**: ✅ Fixed to make constants accessible in method scopes

**Estimated Time**: 2 days
**Dependency**: Useful for class_const*.pas tests

#### 9.5 Class Variables (Static Fields) - MEDIUM PRIORITY

**Blocks**: class_var*.pas (5+ tests)

- [x] 9.5.1 Parse class variable declarations
  - **Task**: Recognize `class var` inside class declarations
  - **Syntax**: `class var VarName: Type;`
  - **Files**: `internal/parser/interfaces.go`
  - **Tests**: Parse class variables in class body
  - **Status**: ✅ Already implemented and tested

- [x] 9.5.2 Store class variables in ClassType
  - **Task**: Add ClassVars map to ClassType
  - **Files**: `internal/types/types.go`
  - **Tests**: Class variables accessible in type system
  - **Status**: ✅ Already implemented and tested

- [x] 9.5.3 Semantic analysis for class variables
  - **Task**: Validate class variable access
  - **Files**: `internal/semantic/analyze_classes.go`
  - **Tests**: Class variables resolve correctly
  - **Status**: ✅ Already implemented and tested

- [x] 9.5.4 Runtime class variable storage and access
  - **Task**: Store class variables separately from instance fields
  - **Files**: `internal/interp/class.go`, `internal/interp/objects.go`
  - **Tests**: Class variables shared across all instances
  - **Status**: ✅ Already implemented and tested

**Estimated Time**: 2-3 days
**Dependency**: Required for class_var*.pas tests

#### 9.6 ClassName Property - HIGH PRIORITY ✅ **COMPLETED**
**Blocks**: classname*.pas, class_of_cast.pas, overload_on_metaclass.pas (10+ tests)

- [x] 9.6.1 Add ClassName to TObject
  - **Task**: Make ClassName available on all classes
  - **Implementation**: Add ClassName as built-in property/method on TObject
  - **Files**: `internal/types/types.go`, `internal/semantic/analyze_classes.go`
  - **Tests**: All classes have ClassName property
  - **Status**: ✅ Implemented in semantic analyzer (analyze_classes.go:1118-1121, analyzer.go:139-143)

- [x] 9.6.2 Runtime ClassName implementation
  - **Task**: Return class name at runtime
  - **Files**: `internal/interp/objects.go`, `internal/interp/builtins.go`
  - **Tests**: obj.ClassName returns correct class name
  - **Status**: ✅ Implemented for objects (objects.go:469-472), metaclasses (objects.go:386-388), and identifiers (expressions.go:112-114, 127-129, analyze_expr_operators.go:50-52)

- [x] 9.6.3 Support ClassName on metaclass types
  - **Task**: TClass.ClassName returns class name
  - **Files**: `internal/interp/class.go`
  - **Tests**: metaclass.ClassName works correctly
  - **Status**: ✅ Implemented with case-insensitive support and full metaclass handling

- [x] 9.6.4 Add TClass type alias
  - **Task**: Register TClass as built-in type alias for "class of TObject"
  - **Files**: `internal/semantic/analyzer.go`
  - **Status**: ✅ Added at analyzer.go:148-154

**Estimated Time**: 1-2 days
**Dependency**: Required for classname*.pas and many other tests
**Completion Date**: 2025-11-09

#### 9.7 ClassType Property - MEDIUM PRIORITY ✅ **COMPLETED**

**Blocks**: classtype*.pas, class_of_cast.pas (5+ tests)

- [x] 9.7.1 Add ClassType to TObject
  - **Task**: Make ClassType available on all classes
  - **Implementation**: Returns metaclass (class of T) for the object's runtime type
  - **Files**: `internal/types/types.go`, `internal/semantic/analyze_classes.go`
  - **Tests**: All classes have ClassType property
  - **Status**: ✅ Implemented and tested

- [x] 9.7.2 Runtime ClassType implementation
  - **Task**: Return metaclass reference at runtime
  - **Files**: `internal/interp/objects.go`, `internal/interp/expressions.go`
  - **Tests**: obj.ClassType returns ClassValue
  - **Status**: ✅ Implemented with case-insensitive support and comprehensive tests

**Estimated Time**: 1 day
**Dependency**: Useful for classtype*.pas tests
**Completion**: Added ClassType property to all classes, returns ClassOfType from semantic analyzer and ClassValue at runtime. Supports member access, identifier access in methods/constructors, and case-insensitive lookup.

#### 9.8 Type Casting - HIGH PRIORITY ✅ **COMPLETED**

**Blocks**: class_cast*.pas, casts_*.pas, classname.pas (10+ tests)

- [x] 9.8.1 Parse function-style type casts
  - **Task**: Recognize `TypeName(expression)` as type cast
  - **Challenge**: Distinguish from function calls
  - **Files**: `internal/semantic/analyze_function_calls.go`
  - **Tests**: Parse `Integer(x)`, `Float(y)`, `TMyClass(obj)`
  - **Status**: ✅ Detection during semantic analysis (not in parser)

- [x] 9.8.2 Semantic analysis for type casts
  - **Task**: Validate type compatibility for casts
  - **Files**: `internal/semantic/analyze_function_calls.go`
  - **Tests**: Valid casts accepted, invalid casts rejected
  - **Status**: ✅ Implemented with `analyzeTypeCast()` and `isValidCast()`

- [x] 9.8.3 Runtime type cast execution
  - **Task**: Perform runtime type checking for casts
  - **Files**: `internal/interp/functions.go`
  - **Tests**: Type casts validate at runtime, raise errors for invalid casts
  - **Status**: ✅ Implemented with cast functions for all basic types and classes

- [x] 9.8.4 Parse `as` operator for safe casting
  - **Task**: Recognize `expression as TypeName`
  - **Files**: Already implemented
  - **Tests**: Parse `obj as TMyClass`
  - **Status**: ✅ Already implemented (parseAsExpression exists)

- [x] 9.8.5 Runtime `as` operator execution
  - **Task**: Perform safe type cast with exception on failure
  - **Files**: Already implemented
  - **Tests**: `as` operator validates types and raises exception on mismatch
  - **Status**: ✅ Already implemented (evalAsExpression exists)

**Estimated Time**: 3-4 days
**Dependency**: Required for class_cast*.pas and many other tests
**Completion**: Full function-style type casting with:
- Basic type conversions: Integer, Float, String, Boolean
- Class type casting with runtime validation
- Variant support
- Case-insensitive type names
- Comprehensive test suite (internal/interp/typecast_test.go)

#### 9.9 Inline Method Implementation - LOW PRIORITY ✅

**Blocks**: overload_on_metaclass.pas, class_inline_declared.pas (2+ tests)

- [x] 9.9.1 Parse inline method bodies
  - **Task**: Allow method implementation inside class declaration
  - **Syntax**: Method declaration followed by begin...end inside class body
  - **Files**: `internal/parser/functions.go`
  - **Tests**: Parse inline method implementations
  - **Status**: Already implemented - parseFunctionDeclaration handles inline methods automatically

- [x] 9.9.2 Handle inline method semantics
  - **Task**: Process inline methods same as separate implementations
  - **Files**: `internal/semantic/analyze_functions.go`
  - **Tests**: Inline methods work identically to separate implementations
  - **Status**: Already working - semantic analyzer handles inline methods correctly

**Estimated Time**: 2-3 days
**Dependency**: Nice-to-have for cleaner syntax
**Completion**: Feature was already implemented. Added comprehensive tests in `internal/parser/inline_methods_test.go`.

#### 9.10 Short-Form Class Declarations - LOW PRIORITY ✅

**Blocks**: class_of3.pas, class_of_cast.pas, classname.pas (5+ tests)

- [x] 9.10.1 Parse short-form class inheritance
  - **Task**: Allow `TChild = class(TParent);` without body
  - **Files**: `internal/parser/classes.go`
  - **Tests**: Parse short-form class declarations
  - **Status**: Implemented - parser now checks for semicolon after parent specification

- [x] 9.10.2 Parse type alias to class
  - **Task**: Allow `TAlias = TClassName;` where TClassName is a class
  - **Files**: `internal/parser/interfaces.go`
  - **Tests**: Parse type aliases to existing classes
  - **Status**: Already implemented - type aliases were already supported

**Estimated Time**: 1 day
**Dependency**: Convenience feature
**Completion**: Both features implemented. Added comprehensive tests in `internal/parser/shortform_class_test.go`.

#### 9.11 Class Forward Declarations - LOW PRIORITY

**Blocks**: class_forward.pas (1 test)

- [x] 9.11.1 Parse forward class declarations
  - **Task**: Allow `TClassName = class;` without body
  - **Files**: `internal/parser/classes.go`
  - **Tests**: Parse forward class declarations
  - **Status**: ✅ DONE - Added parser support and comprehensive tests

- [x] 9.11.2 Handle forward class resolution
  - **Task**: Link forward declarations to actual definitions
  - **Files**: `internal/semantic/analyze_classes.go`, `internal/types/types.go`
  - **Tests**: Forward declarations resolve correctly
  - **Status**: ✅ DONE - Added semantic analyzer support with validation

**Estimated Time**: 1-2 days
**Dependency**: Required for mutually-referencing classes

#### 9.12 Abstract Methods - MEDIUM PRIORITY

**Blocks**: class_abstract_method.pas (1 test)

- [ ] 9.12.1 Parse abstract method directive
  - **Task**: Recognize `abstract` directive on methods
  - **Files**: `internal/parser/functions.go`
  - **Tests**: Parse abstract methods

- [ ] 9.12.2 Semantic analysis for abstract methods
  - **Task**: Validate abstract methods (no body, must be virtual)
  - **Files**: `internal/semantic/analyze_functions.go`
  - **Tests**: Abstract methods validated correctly

- [ ] 9.12.3 Prevent instantiation of abstract classes
  - **Task**: Raise error when creating instance of class with unimplemented abstract methods
  - **Files**: `internal/semantic/analyze_classes.go`
  - **Tests**: Cannot instantiate abstract classes

**Estimated Time**: 1-2 days
**Dependency**: Important for proper OOP design

#### 9.13 Partial Classes - LOW PRIORITY

**Blocks**: partial_class*.pas (5+ tests)

- [ ] 9.13.1 Parse partial class syntax
  - **Task**: Recognize `partial` keyword on class declarations
  - **Files**: `internal/parser/interfaces.go`
  - **Tests**: Parse partial class declarations

- [ ] 9.13.2 Merge partial class definitions
  - **Task**: Combine multiple partial declarations into single class
  - **Files**: `internal/semantic/analyze_types.go`
  - **Tests**: Partial classes merge correctly

**Estimated Time**: 2-3 days
**Dependency**: Advanced feature, low priority

#### 9.14 Operator Overloading for Classes - MEDIUM PRIORITY

**Blocks**: class_operator*.pas, in_class_operator.pas (5+ tests)

- [ ] 9.14.1 Parse class operator declarations
  - **Task**: Allow operator overloading for class types
  - **Files**: `internal/parser/operators.go`
  - **Tests**: Parse class operator declarations

- [ ] 9.14.2 Implement class operator dispatch
  - **Task**: Call overloaded operators for class instances
  - **Files**: `internal/interp/operators_eval.go`
  - **Tests**: Class operators execute correctly

**Estimated Time**: 2-3 days
**Dependency**: Important for operator overloading on objects

#### 9.14.5 "not in" Operator Support - HIGH PRIORITY

**Blocks**: aes_encryption.pas, appinfo.pas, bounds_static0.pas, and other tests using set membership negation

**Background Research** (from `reference/dwscript-original/Source`):
- DWScript treats "not in" as a composition of two operators, not a single token
- The `ttNOT` token is a prefix unary operator (parsed in `ReadNotTerm` function)
- The `ttIN` token is an infix binary operator (parsed in `ReadExprIn` function)
- Expression `x not in set` is parsed as: `NOT (x IN set)`
- The parser handles this through operator precedence:
  1. `ReadTerm` encounters `ttNOT` token → calls `ReadNotTerm`
  2. `ReadNotTerm` recursively calls `ReadTerm` to get operand
  3. Binary operator parsing detects `ttIN` and creates set membership expression
  4. The `NOT` operation wraps the entire `IN` expression

**Current Issue**:
- Parser fails on expressions like `char not in [#0..#255]` with error: "expected next token to be THEN, got NOT instead"
- The `IN` operator exists but prefix `NOT` operator before `IN` is not properly parsed
- This causes nil pointer panics when statements containing "not in" fail to parse

**Implementation Plan**:

- [x] 9.14.5.1 Verify lexer tokens ✅ **COMPLETED**
  - **Task**: Ensure `NOT` and `IN` tokens are properly defined
  - **Files**: `pkg/token/token.go`, `internal/lexer/lexer_test.go`
  - **Tests**: Tokenize "not in" as two separate tokens: NOT, IN
  - **Status**: Both tokens correctly defined as keywords, comprehensive tests added
  - **Result**: Lexer correctly tokenizes "not in" as separate NOT and IN tokens

- [x] 9.14.5.2 Fix "not in/is/as" operator parsing ✅ **COMPLETED**
  - **Task**: Support both `not (x in set)` and `x not in set` syntax
  - **Files**: `internal/parser/expressions.go`, `internal/parser/expressions_test.go`
  - **Implementation**:
    - NOT prefix operator was already registered at line 132 of parser.go
    - Added special handling in `parseExpression()` loop to detect "not in/is/as" patterns
    - When peek token is NOT and current precedence < EQUALS, check if next token is IN/IS/AS
    - If yes, parse as infix comparison and wrap result in NOT unary expression
    - Handles precedence correctly: `x + 1 not in set` → `(not ((x + 1) in set))`
  - **Tests**: Added comprehensive tests in `expressions_test.go`:
    - `TestNotInOperator`: 4 test cases including complex precedence
    - `TestNotIsOperator`: 2 test cases for type checking
    - `TestNotAsOperator`: 2 test cases for type casting
  - **Result**: All tests pass, both syntaxes produce identical AST

- [ ] 9.14.5.3 Verify IN operator precedence
  - **Task**: Ensure `IN` operator has correct precedence relative to `NOT`
  - **Files**: `internal/parser/parser.go` (precedence constants)
  - **Research**: In original DWScript, `IN` is a comparison operator (same level as `=`, `<`, `>`)
  - **Tests**: Parse `x in set1 and y in set2`, `not x in set` correctly
  - **Implementation**:
    - Verify `IN` has comparison-level precedence (same as EQUALS, LESS, GREATER)
    - NOT should have higher precedence (evaluated first in `not x in set` → `(not x) in set`)
    - Actually, based on DWScript behavior: `not x in set` means `not (x in set)`
    - So NOT precedence must be lower than IN, OR parsing must handle NOT specially
  - **Note**: DWScript handles this by parsing NOT as a prefix operator in ReadTerm, before binary operators are considered
  - **Estimated time**: 0.5 day

- [ ] 9.14.5.4 Add semantic analysis for "not in" expressions
  - **Task**: Validate type compatibility for NOT and IN operators
  - **Files**: `internal/semantic/analyze_expressions.go`
  - **Tests**:
    - `not (x in set)` where x is compatible with set element type
    - Error on type mismatch: `not (5 in stringSet)`
  - **Implementation**:
    - `NotExpression` operand must be boolean or convertible to boolean
    - `InExpression` left side must match set/array element type
    - `InExpression` right side must be set, array, or associative array
    - Result type of `InExpression` is boolean
  - **Estimated time**: 0.5 day

- [ ] 9.14.5.5 Implement runtime execution for "not in"
  - **Task**: Execute NOT and IN operators at runtime
  - **Files**: `internal/interp/expressions.go`, `internal/interp/operators.go`
  - **Tests**:
    - `if char not in [#0..#255] then exit;` (from aes_encryption.pas)
    - `if meOne not in s then Print("B");` (enum set membership)
    - `if 'hello' not in strArray then Print("not found");` (array membership)
  - **Implementation**:
    - `evalNotExpression()`: Negate boolean operand
    - `evalInExpression()`: Check membership in set/array/associative array
    - Handle set literals: `[#0..#255]` creates set of character range
    - Support enum sets, integer sets, string arrays, associative arrays
  - **Reference**: `reference/dwscript-original/Source/dwsSetOfExprs.pas` (TSetOfInExpr class)
  - **Estimated time**: 1 day

- [ ] 9.14.5.6 Add comprehensive test coverage
  - **Task**: Create tests for all "not in" use cases
  - **Files**: `internal/parser/expressions_test.go`, `internal/interp/operators_test.go`
  - **Tests**:
    - Parse: `x not in set`, `x not in [1, 2, 3]`, `char not in [#0..#255]`
    - Execute: Character ranges, integer sets, enum sets, string arrays
    - Edge cases: Empty sets, single-element sets, nested expressions
    - Error cases: Type mismatches, invalid set types
  - **Estimated time**: 0.5 day

**Total Estimated Time**: 3-4 days
**Priority**: HIGH - Blocks multiple algorithm tests
**Dependency**: Requires set literal parsing with ranges (`[#0..#255]`) to be fully working

**Success Criteria**:
- `testdata/fixtures/Algorithms/aes_encryption.pas` parses without errors
- Expression `char not in [#0..#255]` evaluates correctly
- All existing tests continue to pass
- No more nil pointer panics from "not in" expressions

**Related Issues**:
- Set literal parsing with ranges (e.g., `[#0..#255]`, `[1..10]`)
- Character literal support (`#0`, `#255`)
- Range expressions in set literals (`start..end`)

---

### Phase 9.15: Documentation & Cleanup

**Priority**: LOW - Can be done in parallel with Phase 10
**Timeline**: 1 week

- [ ] 9.15.1 Update README with current features
  - Document all Stage 7 features now complete
  - Update feature completion percentages
  - Add examples of new features

- [ ] 9.15.2 Create docs/phase9-summary.md
  - Document achievements in Phase 9
  - Statistics: tests passing, coverage percentages
  - Lessons learned and challenges overcome

- [ ] 9.15.3 Update testdata/fixtures/TEST_STATUS.md
  - Update pass/fail counts for each category
  - Mark resolved issues
  - Document remaining blockers

- [ ] 9.15.4 Create docs/limitations.md
  - Document known limitations
  - Features intentionally deferred to later phases
  - Differences from original DWScript

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
