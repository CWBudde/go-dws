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

### Type Aliases (HIGH PRIORITY)

**Summary**: Implement type alias declarations to create alternate names for existing types. Improves code clarity and enables domain-specific naming.

**Example**: `type TUserID = Integer;`, `type TFileName = String;`

#### Type System (2 tasks)

- [x] 9.13 Define `TypeAlias` in `types/types.go`:
  - [x] Fields: `Name string`, `AliasedType Type`
  - [x] Implement `Type` interface methods
  - [x] `TypeKind()` returns underlying type's kind
  - [x] `String()` returns alias name
  - [x] `Equals(other Type)` compares underlying types
- [x] 9.14 Add type alias tests in `types/types_test.go`:
  - [x] Test creating type alias
  - [x] Test alias equality with underlying type
  - [x] Test alias inequality with different types
  - [x] Test nested aliases: `type A = Integer; type B = A;`

#### AST Nodes (2 tasks)

- [x] 9.15 Extend `TypeDeclaration` in `ast/type_annotation.go`:
  - [x] Add `IsAlias bool` field
  - [x] Add `AliasedType TypeAnnotation` field
  - [x] Update `String()` to show `type Name = Type;` for aliases
- [x] 9.16 Add AST tests:
  - [x] Test type alias AST node creation
  - [x] Test `String()` output for aliases

#### Parser Support (2 tasks)

- [x] 9.17 Extend `parseTypeDeclaration()` in `parser/type_declarations.go`:
  - [x] After parsing type name, check next token
  - [x] If `=` token, parse as type alias
  - [x] Parse aliased type annotation
  - [x] Expect SEMICOLON
  - [x] Return TypeDeclaration with IsAlias=true
- [x] 9.18 Add parser tests in `parser/type_test.go`:
  - [x] Test parsing `type TUserID = Integer;`
  - [x] Test parsing `type TFileName = String;`
  - [x] Test parsing alias to custom type: `type TMyClass = TClass;`
  - [x] Test error cases

#### Semantic Analysis (2 tasks)

- [x] 9.19 Implement type alias analysis in `semantic/analyze_types.go`:
  - [x] In `analyzeTypeDeclaration()`, detect type alias
  - [x] Resolve aliased type
  - [x] Create TypeAlias and register in type environment
  - [x] Allow using alias name in variable/parameter declarations
- [x] 9.20 Add semantic tests in `semantic/type_alias_test.go`:
  - [x] Test type alias registration
  - [x] Test using alias in variable declaration: `var id: TUserID;`
  - [x] Test type compatibility: TUserID = Integer should work
  - [x] Test error: undefined aliased type

#### Interpreter Support (1 task)

- [x] 9.21 Implement type alias runtime support:
  - [x] In `resolveType()`, handle TypeAlias by returning underlying type
  - [x] No special runtime representation needed (just resolve to base type)
  - [x] Add tests in `interp/type_test.go`

#### Testing & Fixtures (2 tasks)

- [x] 9.22 Create test scripts in `testdata/type_alias/`:
  - [x] `basic_alias.dws` - Simple type aliases
  - [x] `alias_usage.dws` - Using aliases in declarations and assignments
  - [x] Expected outputs
- [x] 9.23 Add CLI integration tests

---

### Const Declarations (HIGH PRIORITY) ✅ **COMPLETED**

**Summary**: Implement constant declarations with `const` keyword. Constants are immutable values that can be used throughout the program with compile-time type checking.

**Example**: `const PI = 3.14;`, `const MAX_USERS: Integer = 1000;`

#### AST Nodes (1 task)

- [x] 9.23a Define `ConstDecl` in `ast/declarations.go`:
  - [x] Fields: `Name *Identifier`, `Type *TypeAnnotation`, `Value Expression`, `Token`
  - [x] Implement `Statement` interface methods
  - [x] `String()` returns `const Name: Type = Value;` format

#### Parser Support (2 tasks)

- [x] 9.23b Extend parser to handle const declarations:
  - [x] Detect `const` keyword in statement parsing
  - [x] Parse const name (identifier)
  - [x] Parse optional type annotation
  - [x] Require `=` token
  - [x] Parse value expression
  - [x] Expect SEMICOLON
  - [x] Return `ConstDecl` node
- [x] 9.23c Add parser tests in `parser/declarations_test.go`:
  - [x] Test parsing `const PI = 3.14;`
  - [x] Test parsing `const MAX: Integer = 100;`
  - [x] Test with type inference
  - [x] Test error cases (missing value, wrong syntax)

#### Semantic Analysis (2 tasks)

- [x] 9.23d Implement const analysis in `semantic/analyzer.go`:
  - [x] Validate const value is a compile-time constant expression
  - [x] Infer type from value if type annotation omitted
  - [x] Check type compatibility if both specified
  - [x] Register const in symbol table as immutable
  - [x] Prevent reassignment of const values
- [x] 9.23e Add semantic tests in `semantic/const_test.go`:
  - [x] Test const declaration with type annotation
  - [x] Test const declaration with type inference
  - [x] Test const usage in expressions
  - [x] Test error: type mismatch
  - [x] Test error: const reassignment
  - [x] Test error: const redeclaration

#### Interpreter Support (2 tasks)

- [x] 9.23f Implement const runtime support in `interp/interpreter.go`:
  - [x] Store const values in environment as immutable
  - [x] Evaluate const expressions at declaration time
  - [x] Return const values when referenced
  - [x] Prevent runtime modification of const values
- [x] 9.23g Add interpreter tests in `interp/const_test.go`:
  - [x] Test const declaration and usage
  - [x] Test const in expressions
  - [x] Test const scoping

#### Testing & Fixtures (2 tasks)

- [x] 9.23h Create test scripts in `testdata/const/`:
  - [x] `basic_const.dws` - Simple const declarations (Integer, Float, String, Boolean)
  - [x] `const_types.dws` - Const with various types
  - [x] `const_expressions.dws` - Const used in expressions
  - [x] Expected outputs
- [x] 9.23i Add CLI integration tests in `cmd/dwscript/const_test.go`:
  - [x] Test const declaration scripts
  - [x] Verify correct outputs

---

### Ordinal Functions (HIGH PRIORITY)

**Summary**: Implement ordinal functions (Inc, Dec, Succ, Pred, Low, High) for integers, enums, and chars. These are essential for iterating and manipulating ordinal types.

**Note**: These functions should work on any ordinal type (Integer, enum values, Char when implemented).

#### Built-in Functions - Increment/Decrement (4 tasks)

- [x] 9.24 Implement `Inc(x)` and `Inc(x, delta)` in `interp/builtins.go`:
  - [x] Create `builtinInc()` function
  - [x] Accept 1-2 parameters: variable reference, optional delta (default 1)
  - [x] Support Integer: increment by delta
  - [x] Support enum: get next enum value (Succ)
  - [x] Modify variable in-place (requires var parameter support)
  - [x] Return nil
- [x] 9.25 Implement `Dec(x)` and `Dec(x, delta)` in `interp/builtins.go`:
  - [x] Create `builtinDec()` function
  - [x] Accept 1-2 parameters: variable reference, optional delta (default 1)
  - [x] Support Integer: decrement by delta
  - [x] Support enum: get previous enum value (Pred)
  - [x] Modify variable in-place
  - [x] Return nil
- [x] 9.26 Register Inc/Dec in interpreter initialization:
  - [x] Add to global built-in functions map
  - [x] Handle var parameter semantics (pass by reference)
- [x] 9.27 Add tests in `interp/ordinal_test.go`:
  - [x] Test `Inc(x)` with integer: `var x := 5; Inc(x); // x = 6`
  - [x] Test `Inc(x, 3)` with delta: `Inc(x, 3); // x = 8`
  - [x] Test `Dec(x)` with integer
  - [x] Test `Dec(x, 2)` with delta
  - [x] Test Inc/Dec with enum values
  - [x] Test error: Inc beyond High(enum)
  - [x] Test error: Dec below Low(enum)

#### Built-in Functions - Successor/Predecessor (3 tasks)

- [x] 9.28 Implement `Succ(x)` in `interp/builtins.go`:
  - [x] Create `builtinSucc()` function
  - [x] Accept 1 parameter: ordinal value
  - [x] For Integer: return x + 1
  - [x] For enum: return next enum value
  - [x] Raise error if already at maximum value
  - [x] Return successor value
- [x] 9.29 Implement `Pred(x)` in `interp/builtins.go`:
  - [x] Create `builtinPred()` function
  - [x] Accept 1 parameter: ordinal value
  - [x] For Integer: return x - 1
  - [x] For enum: return previous enum value
  - [x] Raise error if already at minimum value
  - [x] Return predecessor value
- [x] 9.30 Add tests in `interp/ordinal_test.go`:
  - [x] Test `Succ(5)` returns 6
  - [x] Test `Pred(5)` returns 4
  - [x] Test Succ/Pred with enum values
  - [x] Test error: Succ at maximum
  - [x] Test error: Pred at minimum

#### Built-in Functions - Low/High for Enums (3 tasks)

- [x] 9.31 Implement `Low(enumType)` in `interp/builtins.go`:
  - [x] Create `builtinLow()` function
  - [x] Accept enum type or enum value
  - [x] For arrays: return array lower bound (already implemented)
  - [x] For enum type: return lowest enum value
  - [x] For enum value: return Low of that enum type
  - [x] Return lowest ordinal value
- [x] 9.32 Implement `High(enumType)` in `interp/builtins.go`:
  - [x] Create `builtinHigh()` function
  - [x] Accept enum type or enum value
  - [x] For arrays: return array upper bound (already implemented)
  - [x] For enum type: return highest enum value
  - [x] For enum value: return High of that enum type
  - [x] Return highest ordinal value
- [x] 9.33 Add tests in `interp/ordinal_test.go`:
  - [x] Test `Low(TColor)` returns first enum value (Red)
  - [x] Test `High(TColor)` returns last enum value (Blue)
  - [x] Test Low/High with enum variable: `var c: TColor; Low(c)`
  - [x] Test Low/High still work for arrays (backward compatibility)

#### Testing & Fixtures (2 tasks)

- [x] 9.34 Create test scripts in `testdata/ordinal_functions/`:
  - [x] `inc_dec.dws` - Inc and Dec with integers and enums
  - [x] `succ_pred.dws` - Succ and Pred with integers and enums
  - [x] `low_high_enum.dws` - Low and High for enum types
  - [x] `for_loop_enum.dws` - Using Low/High in for loops: `for i := Low(TEnum) to High(TEnum)`
  - [x] Expected outputs
- [x] 9.35 Add CLI integration tests:
  - [x] Test ordinal function scripts
  - [x] Verify correct outputs

---

### Assert Function (HIGH PRIORITY)

**Summary**: Implement `Assert(condition)` and `Assert(condition, message)` built-in functions for runtime assertions. Critical for testing and contracts.

#### Built-in Function (2 tasks)

- [x] 9.36 Implement `Assert()` in `interp/builtins.go`:
  - [x] Create `builtinAssert()` function
  - [x] Accept 1-2 parameters: Boolean condition, optional String message
  - [x] If condition is false:
    - [x] If message provided, raise `EAssertionFailed` with message
    - [x] If no message, raise `EAssertionFailed` with "Assertion failed"
  - [x] If condition is true, return nil (no-op)
  - [x] Register in global built-in functions
- [x] 9.37 Add tests in `interp/assert_test.go`:
  - [x] Test `Assert(true)` - should not raise error
  - [x] Test `Assert(false)` - should raise EAssertionFailed
  - [x] Test `Assert(true, 'message')` - no error
  - [x] Test `Assert(false, 'Custom message')` - error with custom message
  - [x] Test Assert in function: function validates preconditions
  - [x] Test Assert with expression: `Assert(x > 0, 'x must be positive')`

#### Testing & Fixtures (2 tasks)

- [x] 9.38 Create test scripts in `testdata/assert/`:
  - [x] `assert_basic.dws` - Basic Assert usage
  - [x] `assert_validation.dws` - Using Assert for input validation
  - [x] `assert.dws` - Reference test from original DWScript
  - [x] Expected outputs (some should fail with assertion errors)
- [x] 9.39 Add CLI integration tests:
  - [x] Test assert scripts
  - [x] Verify assertion failures are caught and reported

---

### Priority String Functions (HIGH PRIORITY)

**Summary**: Implement essential string manipulation functions: Trim, Insert, Delete, Format, StringReplace. These are heavily used in real programs.

#### Built-in Functions - Trim (3 tasks)

- [x] 9.40 Implement `Trim(s)` in `interp/string_functions.go`:
  - [x] Create `builtinTrim()` function
  - [x] Accept String parameter
  - [x] Remove leading and trailing whitespace
  - [x] Use Go's `strings.TrimSpace()`
  - [x] Return trimmed string
- [x] 9.41 Implement `TrimLeft(s)` and `TrimRight(s)`:
  - [x] Create `builtinTrimLeft()` - remove leading whitespace only
  - [x] Create `builtinTrimRight()` - remove trailing whitespace only
  - [x] Use `strings.TrimLeftFunc()` and `strings.TrimRightFunc()`
- [x] 9.42 Add tests in `interp/string_test.go`:
  - [x] Test `Trim('  hello  ')` returns 'hello'
  - [x] Test `TrimLeft('  hello')` returns 'hello'
  - [x] Test `TrimRight('hello  ')` returns 'hello'
  - [x] Test with tabs and newlines
  - [x] Test with no whitespace (no-op)

#### Built-in Functions - Insert/Delete (3 tasks)

- [x] 9.43 Implement `Insert(source, s, pos)` in `interp/interpreter.go`:
  - [x] Create `builtinInsert()` function
  - [x] Accept 3 parameters: source String, target String (var param), position Integer
  - [x] Insert source into target at 1-based position
  - [x] Modify target string in-place (var parameter)
  - [x] Handle edge cases: pos < 1, pos > length
- [x] 9.44 Implement `Delete(s, pos, count)` in `interp/interpreter.go`:
  - [x] Create `builtinDeleteString()` function
  - [x] Accept 3 parameters: string (var param), position Integer, count Integer
  - [x] Delete count characters starting at 1-based position
  - [x] Modify string in-place (var parameter)
  - [x] Handle edge cases: pos < 1, pos > length, count too large
- [x] 9.45 Add tests in `interp/string_test.go`:
  - [x] Test Insert: `var s := 'Helo'; Insert('l', s, 3);` → 'Hello'
  - [x] Test Delete: `var s := 'Hello'; Delete(s, 3, 2);` → 'Heo'
  - [x] Test Insert at start/end
  - [x] Test Delete edge cases
  - [x] Test error cases

#### Built-in Functions - StringReplace (2 tasks)

- [x] 9.46 Implement `StringReplace(s, old, new)` in `interp/interpreter.go`:
  - [x] Create `builtinStringReplace()` function
  - [x] Accept 3 parameters: string, old substring, new substring
  - [x] Optional 4th parameter: count (replace count occurrences, -1 for all)
  - [x] Use Go's `strings.Replace()`
  - [x] Return new string with replacements
- [x] 9.47 Add tests in `interp/string_test.go`:
  - [x] Test replace all: `StringReplace('hello world', 'l', 'L')` → 'heLLo worLd'
  - [x] Test replace first only (count parameter supported)
  - [x] Test with empty old string
  - [x] Test with empty new string (delete)

#### Built-in Functions - Format (4 tasks)

- [x] 9.48 Implement `Format(fmt, args)` in `interp/interpreter.go`:
  - [x] Create `builtinFormat()` function (lines 1891-2017)
  - [x] Accept format string and variadic args (array of values)
  - [x] Support format specifiers: `%s` (string), `%d` (integer), `%f` (float), `%%` (literal %)
  - [x] Optional: support width and precision: `%5d`, `%.2f`
  - [x] Use Go's `fmt.Sprintf()` or custom formatter
  - [x] Return formatted string
- [x] 9.49 Support array of const for Format args:
  - [x] Parse variadic parameters as array
  - [x] Convert DWScript values to Go values for formatting
  - [x] Handle different value types (Integer, Float, String, Boolean)
- [x] 9.50 Add tests in `interp/string_test.go`:
  - [x] Test `Format('Hello %s', ['World'])` → 'Hello World' (24 tests total)
  - [x] Test `Format('Value: %d', [42])` → 'Value: 42'
  - [x] Test `Format('Pi: %.2f', [3.14159])` → 'Pi: 3.14'
  - [x] Test multiple args: `Format('%s is %d', ['Age', 25])`
  - [x] Test error: wrong number of args
- [x] 9.51 Documentation in `docs/builtins.md`:
  - [x] Document Format syntax
  - [x] List supported format specifiers
  - [x] Provide examples

#### Semantic Analysis Support (1 task)

- [x] 9.51a Add semantic analysis for string functions in `semantic/analyze_expressions.go`:
  - [x] `Trim(s: String): String` - validate 1 string argument (duplicates at lines 1115-1127 REMOVED)
  - [x] `TrimLeft(s: String): String` - validate 1 string argument (duplicates at lines 1130-1142 REMOVED)
  - [x] `TrimRight(s: String): String` - validate 1 string argument (duplicates at lines 1145-1157 REMOVED)
  - [x] `Insert(source: String, var target: String, pos: Integer)` - validate var parameter (indentation FIXED)
  - [x] `Delete` - handle overloading: `Delete(var arr: Array, index: Integer)` OR `Delete(var s: String, pos: Integer, count: Integer)` (lines 853-898, already correct)
  - [x] `StringReplace(s: String, old: String, new: String): String` - validate 3 strings (duplicates at lines 1190-1204 REMOVED)
  - [x] `Format(fmt: String, args: Array): String` - validate format string and array (ADDED at lines 548-571)
  - [x] Remove duplicate function checks (Trim family and StringReplace duplicates removed)
  - [x] Fix indentation issues in lines 1114+ (Insert function indentation fixed)

#### Testing & Fixtures (2 tasks)

- [x] 9.52 Create test scripts in `testdata/string_functions/`:
  - [x] `trim.dws` - Trim, TrimLeft, TrimRight ✓ TESTED AND WORKING
  - [x] `insert_delete.dws` - Insert and Delete ✓ TESTED AND WORKING
  - [x] `replace.dws` - StringReplace ✓ TESTED AND WORKING
  - [x] Expected outputs
- [x] 9.53 Add CLI integration tests in `cmd/dwscript/string_functions_test.go`:
  - [x] Test trim.dws ✓ PASSING
  - [x] Test insert_delete.dws ✓ PASSING
  - [x] Test replace.dws ✓ PASSING
  - [x] Verify outputs match expected files ✓ ALL TESTS PASSING

**Note**: Format function testing deferred to task 9.88 due to array literal syntax complexity

---

### Priority Math Functions (HIGH PRIORITY)

**Summary**: Implement essential math functions: Min, Max, Sqr, Power, Ceil, Floor, RandomInt. Complete the math function library.

#### Built-in Functions - Min/Max (3 tasks)

- [x] 9.54 Implement `Min(a, b)` in `internal/interp/interpreter.go`:
  - [x] Create `builtinMin()` function
  - [x] Accept 2 parameters: both Integer or both Float
  - [x] Return smaller value, preserving type
  - [x] Handle mixed types: promote Integer to Float
- [x] 9.55 Implement `Max(a, b)` in `internal/interp/interpreter.go`:
  - [x] Create `builtinMax()` function
  - [x] Accept 2 parameters: both Integer or both Float
  - [x] Return larger value, preserving type
  - [x] Handle mixed types: promote Integer to Float
- [x] 9.56 Add tests in `interp/math_test.go`:
  - [x] Test `Min(5, 10)` returns 5
  - [x] Test `Max(5, 10)` returns 10
  - [x] Test with negative numbers
  - [x] Test with floats: `Min(3.14, 2.71)`
  - [x] Test with mixed types: `Min(5, 3.14)`

#### Built-in Functions - Sqr/Power (3 tasks)

- [x] 9.57 Implement `Sqr(x)` in `internal/interp/interpreter.go`:
  - [x] Create `builtinSqr()` function
  - [x] Accept Integer or Float parameter
  - [x] Return x * x, preserving type
  - [x] Integer sqr returns Integer, Float sqr returns Float
- [x] 9.58 Implement `Power(x, y)` in `internal/interp/interpreter.go`:
  - [x] Create `builtinPower()` function
  - [x] Accept base and exponent (Integer or Float)
  - [x] Use Go's `math.Pow()`
  - [x] Always return Float (even for integer inputs)
  - [x] Handle special cases: 0^0, negative base with fractional exponent
- [x] 9.59 Add tests in `interp/math_test.go`:
  - [x] Test `Sqr(5)` returns 25
  - [x] Test `Sqr(3.0)` returns 9.0
  - [x] Test `Power(2, 8)` returns 256.0
  - [x] Test `Power(2.0, 0.5)` returns 1.414... (sqrt(2))
  - [x] Test negative exponent: `Power(2, -1)` returns 0.5

#### Built-in Functions - Ceil/Floor (3 tasks)

- [x] 9.60 Implement `Ceil(x)` in `interp/math_functions.go`:
  - [x] Create `builtinCeil()` function
  - [x] Accept Float parameter
  - [x] Round up to nearest integer
  - [x] Use Go's `math.Ceil()`
  - [x] Return Integer type
- [x] 9.61 Implement `Floor(x)` in `interp/math_functions.go`:
  - [x] Create `builtinFloor()` function
  - [x] Accept Float parameter
  - [x] Round down to nearest integer
  - [x] Use Go's `math.Floor()`
  - [x] Return Integer type
- [x] 9.62 Add tests in `interp/math_test.go`:
  - [x] Test `Ceil(3.2)` returns 4
  - [x] Test `Ceil(3.8)` returns 4
  - [x] Test `Ceil(-3.2)` returns -3
  - [x] Test `Floor(3.8)` returns 3
  - [x] Test `Floor(3.2)` returns 3
  - [x] Test `Floor(-3.8)` returns -4

#### Built-in Functions - RandomInt (2 tasks)

- [x] 9.63 Implement `RandomInt(max)` in `interp/math_functions.go`:
  - [x] Create `builtinRandomInt()` function
  - [x] Accept Integer parameter: max (exclusive upper bound)
  - [x] Return random Integer in range [0, max)
  - [x] Use Go's `rand.Intn()`
  - [x] Validate max > 0
- [x] 9.64 Add tests in `interp/math_test.go`:
  - [x] Test `RandomInt(10)` returns value in [0, 10)
  - [x] Test multiple calls return different values (probabilistic)
  - [x] Test with max=1: always returns 0
  - [x] Test error: RandomInt(0) or RandomInt(-5)

#### Testing & Fixtures (2 tasks)

- [x] 9.65 Create test scripts in `testdata/math_functions/`:
  - [x] `min_max.dws` - Min and Max with various inputs ✓ COMPLETE
  - [x] `sqr_power.dws` - Sqr and Power functions ✓ COMPLETE
  - [x] `ceil_floor.dws` - Ceil and Floor functions ✓ COMPLETE
  - [x] `random_int.dws` - RandomInt usage ✓ COMPLETE
  - [x] Expected outputs ✓ COMPLETE
- [x] 9.66 Add CLI integration tests:
  - [x] Test all math function fixtures ✓ ALL PASSING
  - [x] Test inline code for each function ✓ ALL PASSING
  - [x] Test parsing for all scripts ✓ ALL PASSING

---

### Priority Array Functions (HIGH PRIORITY)

**Summary**: Implement essential array manipulation functions: Copy, IndexOf, Contains, Reverse, Sort. Complete the array function library.

#### Built-in Functions - Copy (2 tasks)

- [x] 9.67 Implement `Copy(arr)` for arrays in `interp/array_functions.go`:
  - [x] Create `builtinArrayCopy()` function (overload existing Copy)
  - [x] Accept array parameter
  - [x] Return deep copy of array
  - [x] For dynamic arrays, create new array with same elements
  - [x] For static arrays, copy elements to new array
  - [x] Handle arrays of objects (shallow copy references)
- [x] 9.68 Add tests in `interp/array_test.go`:
  - [x] Test copy dynamic array: `var a2 := Copy(a1); a2[0] := 99;` → a1 unchanged
  - [x] Test copy static array
  - [x] Test copy preserves element types
  - [x] Test copy empty array

#### Built-in Functions - IndexOf (3 tasks)

- [x] 9.69 Implement `IndexOf(arr, value)` in `interp/array_functions.go`:
  - [x] Create `builtinIndexOf()` function
  - [x] Accept array and value to find
  - [x] Search array for first occurrence of value
  - [x] Use equality comparison (handle different types)
  - [x] Return 1-based index if found (implemented: uses 1-based indexing like Pascal Pos())
  - [x] Return 0 if not found (implemented: Pascal convention, not -1)
- [x] 9.70 Implement `IndexOf(arr, value, startIndex)` variant:
  - [x] Accept optional 3rd parameter: start index (0-based internal)
  - [x] Search from startIndex onwards
  - [x] Handle startIndex out of bounds (returns 0)
- [x] 9.71 Add tests in `interp/array_test.go`:
  - [x] Test `IndexOf([1,2,3,2], 2)` returns 2 (first occurrence: 1-based)
  - [x] Test `IndexOf([1,2,3], 5)` returns 0 (not found)
  - [x] Test with start index: `IndexOf([1,2,3,2], 2, 2)` returns 4
  - [x] Test with strings
  - [x] Test with empty array
  - [x] Test edge cases (negative startIndex, beyond bounds)
  - [x] Test error cases (wrong arg count, wrong types)

#### Built-in Functions - Contains (2 tasks)

- [x] 9.72 Implement `Contains(arr, value)` in `interp/array_functions.go`:
  - [x] Create `builtinContains()` function
  - [x] Accept array and value
  - [x] Return true if array contains value, false otherwise
  - [x] Internally use IndexOf (return IndexOf > 0, since 0 = not found)
- [x] 9.73 Add tests in `interp/array_test.go`:
  - [x] Test `Contains([1,2,3], 2)` returns true
  - [x] Test `Contains([1,2,3], 5)` returns false
  - [x] Test with different types
  - [x] Test with empty array returns false

#### Built-in Functions - Reverse (2 tasks)

- [x] 9.74 Implement `Reverse(arr)` in `interp/array_functions.go`:
  - [x] Create `builtinReverse()` function
  - [x] Accept array (var parameter - modify in place)
  - [x] Reverse array elements in-place
  - [x] Swap elements from both ends moving inward
  - [x] Return nil (modifies in place)
- [x] 9.75 Add tests in `interp/array_test.go`:
  - [x] Test `var a := [1,2,3]; Reverse(a);` → a = [3,2,1]
  - [x] Test with even length array
  - [x] Test with odd length array
  - [x] Test with single element (no-op)
  - [x] Test with empty array (no-op)

#### Built-in Functions - Sort (3 tasks)

- [x] 9.76 Implement `Sort(arr)` in `interp/array_functions.go`:
  - [x] Create `builtinSort()` function
  - [x] Accept array (var parameter - modify in place)
  - [x] Sort array elements using default comparison
  - [x] For Integer arrays: numeric sort
  - [x] For String arrays: lexicographic sort
  - [x] Use Go's `sort.Slice()`
  - [x] Return nil (modifies in place)
- [ ] 9.77 Add optional comparator parameter (future):
  - [ ] `Sort(arr, comparator)` with custom comparison function
  - [ ] Comparator returns -1, 0, 1 for less, equal, greater
  - [ ] Note: Requires function pointers (deferred)
- [x] 9.78 Add tests in `interp/array_test.go`:
  - [x] Test `var a := [3,1,2]; Sort(a);` → a = [1,2,3]
  - [x] Test with strings: `['c','a','b']` → `['a','b','c']`
  - [x] Test with already sorted array (no-op)
  - [x] Test with single element
  - [x] Test with duplicates

#### Testing & Fixtures (2 tasks)

- [x] 9.79 Create test scripts in `testdata/array_functions/`:
  - [x] `copy.dws` - Array copying and independence
  - [x] `search.dws` - IndexOf and Contains
  - [x] `reverse.dws` - Reverse array
  - [x] `sort.dws` - Sort arrays
  - [x] Expected outputs
- [x] 9.80 Add CLI integration tests:
  - [x] Test array function scripts
  - [x] Verify outputs

---

### Contracts (Design by Contract)

- [ ] 9.81 Parse require/ensure clauses (if supported)
- [ ] 9.82 Implement contract checking at runtime
- [ ] 9.83 Test contracts

### Comprehensive Testing (Stage 8)

- [ ] 9.84 Port DWScript's test suite (if available)
- [ ] 9.85 Run DWScript example scripts from documentation
- [ ] 9.86 Compare outputs with original DWScript
- [ ] 9.87 Fix any discrepancies
- [ ] 9.88 Create stress tests for complex features
- [ ] 9.89 Achieve >85% overall code coverage

### Format Function Testing (DEFERRED)

**Summary**: Create comprehensive test fixtures for the Format() built-in function. Deferred from task 9.52 due to DWScript's set literal syntax `[...]` conflicting with Format's array parameter requirements.

#### Task Details (1 task)

- [ ] 9.90 Create Format function test fixtures:
  - [ ] Implement proper array construction for Format args (using `array of` or alternative syntax)
  - [ ] Create `testdata/string_functions/format.dws` with Format examples
  - [ ] Test %s (string), %d (integer), %f (float) specifiers
  - [ ] Test width and precision: %5d, %.2f, %8.2f
  - [ ] Test %% (literal percent)
  - [ ] Test multiple arguments
  - [ ] Create expected output file
  - [ ] Add CLI integration tests for Format
  - [ ] Document Format syntax in `docs/builtins.md` (Task 9.51)

---

### Subrange Types (HIGH PRIORITY)

**Summary**: Implement subrange type declarations for type-safe bounded values. Subranges restrict a type to a specific range and provide runtime validation.

**Example**: `type TDigit = 0..9;`, `type TPercent = 0..100;`

**Reference**: docs/missing-features-recommendations.md lines 206-233

#### Type System (3 tasks)

- [x] 9.91 Define `SubrangeType` in `types/types.go`:
  - [x] Fields: `Name string`, `BaseType Type` (Integer, Char, enum), `LowBound int`, `HighBound int`
  - [x] Implement `Type` interface methods
  - [x] `TypeKind()` returns `TypeKindSubrange`
  - [x] `String()` returns `LowBound..HighBound`
  - [x] `Equals(other Type)` compares base type and bounds
  - [x] `Contains(value int)` checks if value is in range
- [x] 9.92 Add subrange validation functions:
  - [x] `ValidateRange(value int, subrange *SubrangeType) error`
  - [x] Returns error if value outside bounds
  - [x] Used by interpreter at assignment time
- [x] 9.93 Add tests in `types/subrange_test.go`:
  - [x] Test creating subrange types
  - [x] Test range validation
  - [x] Test type compatibility (subrange assignable to base type)
  - [x] Test nested subranges: `type TSmallDigit = 0..5; type TTinyDigit: TSmallDigit = 0..3;`

#### AST Nodes (2 tasks)

- [x] 9.94 Extend `TypeDeclaration` in `ast/type_annotation.go`:
  - [x] Add `IsSubrange bool` field
  - [x] Add `LowBound Expression` and `HighBound Expression` fields
  - [x] Update `String()` to show `type Name = Low..High;`
- [x] 9.95 Add AST tests:
  - [x] Test subrange type AST node creation
  - [x] Test `String()` output for subranges

#### Parser Support (2 tasks)

- [x] 9.96 Extend `parseTypeDeclaration()` in `parser/type_declarations.go`:
  - [x] After parsing type name and `=`, check for subrange pattern
  - [x] Parse low bound expression (must be constant)
  - [x] Expect `..` token (DOTDOT)
  - [x] Parse high bound expression (must be constant)
  - [x] Expect SEMICOLON
  - [x] Return TypeDeclaration with IsSubrange=true
- [x] 9.97 Add parser tests in `parser/type_test.go`:
  - [x] Test parsing `type TDigit = 0..9;`
  - [x] Test parsing `type TPercent = 0..100;`
  - [x] Test parsing negative ranges: `type TTemperature = -40..50;`
  - [x] Test error cases (missing DOTDOT, missing semicolon, missing bounds)

#### Semantic Analysis (2 tasks)

- [x] 9.98 Implement subrange analysis in `semantic/analyze_types.go`:
  - [x] In `analyzeTypeDeclaration()`, detect subrange type
  - [x] Evaluate low and high bound expressions (must be compile-time constants)
  - [x] Validate low <= high
  - [x] Create SubrangeType and register in type environment
  - [x] Check type compatibility in assignments (subrange ↔ base type)
- [x] 9.99 Add semantic tests in `semantic/subrange_test.go`:
  - [x] Test subrange type registration
  - [x] Test using subrange in variable declaration: `var digit: TDigit;`
  - [x] Test assignment validation: `digit := 5;` (OK), `digit := 99;` (error at runtime)
  - [x] Test error: low > high
  - [x] Test error: non-constant bounds

#### Interpreter Support (2 tasks)

- [x] 9.100 Implement subrange runtime support in `interp/interpreter.go`:
  - [x] In `resolveType()`, handle SubrangeType
  - [x] On assignment to subrange variable, call `ValidateRange()`
  - [x] Raise runtime error if value out of bounds
  - [x] Add tests in `interp/subrange_test.go`
- [x] 9.101 Add subrange coercion support:
  - [x] Subrange values assignable to base type (no check needed)
  - [x] Base type values assignable to subrange (runtime check)
  - [x] Test bidirectional assignment

#### Testing & Fixtures (1 task)

- [x] 9.102 Create test scripts in `testdata/subrange/`:
  - [x] `basic_subrange.dws` - Simple subrange declarations and usage
  - [x] `subrange_validation.dws` - Runtime validation (should fail with out-of-range error)
  - [x] `subrange_functions.dws` - Subranges as parameters and return types
  - [x] Expected outputs (some with runtime errors)
  - [x] Add CLI integration tests

---

### Units/Modules System (CRITICAL)

**Summary**: Implement a units/modules system for organizing code across multiple files. Essential for larger programs and code reuse.

**Example**:
```pascal
unit MyUtils;

interface
  function Add(a, b: Integer): Integer;

implementation
  function Add(a, b: Integer): Integer;
  begin
    Result := a + b;
  end;

initialization
  PrintLn('MyUtils loaded');

finalization
  PrintLn('MyUtils unloading');
end.
```

**Reference**: docs/missing-features-recommendations.md lines 149-168

#### Type System and Data Structures (5 tasks)

- [ ] 9.103 Create `units/` package directory
- [ ] 9.104 Create `units/unit.go` - Unit representation:
  - [ ] Define `Unit` struct with `Name string`, `InterfaceSection *ast.Block`, `ImplementationSection *ast.Block`
  - [ ] Add `InitializationSection *ast.Block` and `FinalizationSection *ast.Block`
  - [ ] Add `Uses []string` (list of imported units)
  - [ ] Add `Symbols *semantic.SymbolTable` (exported symbols from interface)
  - [ ] Add `FilePath string` (source file path)
- [ ] 9.105 Create `units/registry.go` - Unit registry:
  - [ ] Define `UnitRegistry` struct with map of loaded units
  - [ ] Implement `RegisterUnit(name string, unit *Unit) error`
  - [ ] Implement `LoadUnit(name string, searchPaths []string) (*Unit, error)`
  - [ ] Implement circular dependency detection
  - [ ] Cache compiled units to avoid reloading
- [ ] 9.106 Create `units/search.go` - Unit search paths:
  - [ ] Implement `FindUnit(name string, paths []string) (string, error)`
  - [ ] Support relative and absolute paths
  - [ ] Search in: current directory, specified paths, system paths
  - [ ] File naming convention: `UnitName.dws` or `UnitName.pas`
- [ ] 9.107 Add unit tests for registry and search functionality

#### AST Nodes (3 tasks)

- [ ] 9.108 Create `ast/unit.go` - Unit AST nodes:
  - [ ] Define `UnitDeclaration` struct implementing `Node`
  - [ ] Fields: `Name *Identifier`, `InterfaceSection *Block`, `ImplementationSection *Block`, `InitSection *Block`, `FinalSection *Block`
  - [ ] Implement `String()` method
- [ ] 9.109 Define `UsesClause` struct:
  - [ ] Fields: `Units []*Identifier` (list of unit names)
  - [ ] Appears in both interface and implementation sections
  - [ ] Implement `String()` method
- [ ] 9.110 Add AST tests for unit nodes

#### Lexer Support (2 tasks)

- [ ] 9.111 Add unit-related keywords to lexer:
  - [ ] `UNIT`, `INTERFACE`, `IMPLEMENTATION`, `USES`
  - [ ] `INITIALIZATION`, `FINALIZATION`
  - [ ] Update `token_type.go` and keyword map
- [ ] 9.112 Add lexer tests for new keywords

#### Parser Support (8 tasks)

- [ ] 9.113 Create `parser/unit.go` - Unit parsing:
  - [ ] Implement `parseUnit() *ast.UnitDeclaration`
  - [ ] Parse `unit` keyword and name
  - [ ] Expect SEMICOLON
  - [ ] Parse interface section (starts with `interface`)
  - [ ] Parse implementation section (starts with `implementation`)
  - [ ] Parse optional initialization section
  - [ ] Parse optional finalization section
  - [ ] Expect `end.` to close unit
- [ ] 9.114 Implement `parseInterfaceSection() *ast.Block`:
  - [ ] Parse uses clause (if present)
  - [ ] Parse declarations (types, constants, functions/procedures signatures only)
  - [ ] No implementation code in interface
- [ ] 9.115 Implement `parseImplementationSection() *ast.Block`:
  - [ ] Parse uses clause (if present)
  - [ ] Parse full function/procedure implementations
  - [ ] Parse private declarations (not exported)
- [ ] 9.116 Implement `parseUsesClause() *ast.UsesClause`:
  - [ ] Parse `uses` keyword
  - [ ] Parse comma-separated unit names
  - [ ] Expect SEMICOLON
- [ ] 9.117 Implement `parseInitializationSection() *ast.Block`:
  - [ ] Parse `initialization` keyword
  - [ ] Parse statement list
  - [ ] Ends at `finalization` or `end`
- [ ] 9.118 Implement `parseFinalizationSection() *ast.Block`:
  - [ ] Parse `finalization` keyword
  - [ ] Parse statement list
  - [ ] Ends at `end`
- [ ] 9.119 Update main parser to detect unit vs program:
  - [ ] If file starts with `unit`, parse as unit
  - [ ] Otherwise, parse as program
- [ ] 9.120 Add parser tests for units in `parser/unit_test.go`

#### Semantic Analysis (10 tasks)

- [ ] 9.121 Create `semantic/unit_analyzer.go`:
  - [ ] Implement `AnalyzeUnit(unit *ast.UnitDeclaration, registry *units.UnitRegistry) error`
  - [ ] Build symbol table for interface section (exported symbols)
  - [ ] Analyze implementation section with access to interface symbols
  - [ ] Validate that all interface declarations have implementations
- [ ] 9.122 Implement uses clause resolution:
  - [ ] For each unit in uses clause, load it from registry
  - [ ] Import exported symbols into current scope
  - [ ] Handle name conflicts (error or qualified access)
  - [ ] Build dependency graph
- [ ] 9.123 Implement circular dependency detection:
  - [ ] Track unit dependency chain during loading
  - [ ] Detect cycles: A uses B, B uses A
  - [ ] Report error with cycle path
- [ ] 9.124 Implement namespace resolution:
  - [ ] Support qualified access: `UnitName.SymbolName`
  - [ ] Support unqualified access for imported symbols
  - [ ] Handle ambiguous symbols (multiple units export same name)
- [ ] 9.125 Implement interface/implementation validation:
  - [ ] Check that all interface functions have implementation
  - [ ] Check signatures match exactly
  - [ ] Check visibility rules (private vs public)
- [ ] 9.126 Handle forward declarations across units:
  - [ ] Interface declares functions
  - [ ] Implementation provides bodies
  - [ ] Cross-unit calls use interface signatures
- [ ] 9.127 Implement unit initialization order:
  - [ ] Topological sort of dependency graph
  - [ ] Units initialize in dependency order
  - [ ] Finalize in reverse order
- [ ] 9.128 Add semantic tests for units
- [ ] 9.129 Test circular dependency detection
- [ ] 9.130 Test namespace resolution and conflicts

#### Interpreter Support (8 tasks)

- [ ] 9.131 Create `interp/unit_loader.go`:
  - [ ] Implement `LoadUnit(name string, registry *units.UnitRegistry) error`
  - [ ] Load and analyze unit if not already loaded
  - [ ] Execute initialization section
  - [ ] Register exported symbols in global environment
- [ ] 9.132 Implement unit initialization:
  - [ ] Execute initialization blocks in dependency order
  - [ ] Track initialized units to avoid double-init
  - [ ] Handle initialization errors
- [ ] 9.133 Implement unit finalization:
  - [ ] Execute finalization blocks at program exit
  - [ ] Finalize in reverse dependency order
  - [ ] Handle finalization errors gracefully
- [ ] 9.134 Implement qualified name resolution:
  - [ ] `UnitName.FunctionName()` calls
  - [ ] Lookup in unit's exported symbols
  - [ ] Cache lookups for performance
- [ ] 9.135 Implement unit symbol import:
  - [ ] Import symbols from used units into current environment
  - [ ] Handle naming conflicts
  - [ ] Support hiding/renaming (if DWScript supports it)
- [ ] 9.136 Add tests for unit loading and initialization
- [ ] 9.137 Test cross-unit function calls
- [ ] 9.138 Test initialization/finalization order

#### CLI and Tooling (3 tasks)

- [ ] 9.139 Update CLI to support unit compilation:
  - [ ] Add `-I` flag for unit search paths
  - [ ] `./bin/dwscript run main.dws -I ./units -I ./lib`
  - [ ] Display loaded units and dependency order
- [ ] 9.140 Implement unit compilation cache:
  - [ ] Cache parsed and analyzed units
  - [ ] Invalidate cache on file modification
  - [ ] Speed up repeated runs
- [ ] 9.141 Add `--show-units` flag to display unit dependency tree

#### Testing & Fixtures (4 tasks)

- [ ] 9.142 Create test units in `testdata/units/`:
  - [ ] `MathUtils.dws` - Math helper functions
  - [ ] `StringUtils.dws` - String helper functions
  - [ ] `main.dws` - Program that uses both units
  - [ ] Test initialization and finalization output
- [ ] 9.143 Create circular dependency test:
  - [ ] `UnitA.dws` uses `UnitB`
  - [ ] `UnitB.dws` uses `UnitA`
  - [ ] Verify error is caught
- [ ] 9.144 Create namespace conflict test:
  - [ ] Two units export same function name
  - [ ] Test qualified access resolves correctly
  - [ ] Test unqualified access reports ambiguity
- [ ] 9.145 Add CLI integration tests for units

---

### Function/Method Pointers (HIGH VALUE)

**Summary**: Implement function and method pointers for callbacks, event handlers, and higher-order functions. Essential for functional programming patterns.

**Example**:
```pascal
type
  TComparator = function(a, b: Integer): Integer;

function Ascending(a, b: Integer): Integer;
begin
  Result := a - b;
end;

var compare: TComparator;
begin
  compare := @Ascending;
  PrintLn(compare(5, 3)); // Output: 2
end.
```

**Reference**: docs/missing-features-recommendations.md lines 171-203

#### Type System (5 tasks)

- [ ] 9.146 Define `FunctionPointerType` in `types/types.go`:
  - [ ] Fields: `Params []Type`, `ReturnType Type` (nil for procedures)
  - [ ] Implement `Type` interface methods
  - [ ] `TypeKind()` returns `TypeKindFunctionPointer`
  - [ ] `String()` returns `function(params): ReturnType` or `procedure(params)`
  - [ ] `Equals(other Type)` compares signatures
- [ ] 9.147 Define `MethodPointerType` in `types/types.go`:
  - [ ] Extends `FunctionPointerType` with `OfObject bool`
  - [ ] Stores both function pointer and object instance (`Self`)
  - [ ] `String()` returns `function(...) of object`
- [ ] 9.148 Implement function pointer compatibility:
  - [ ] Check parameter types match exactly
  - [ ] Check return type matches
  - [ ] Method pointers compatible with function pointers if signatures match
- [ ] 9.149 Add tests in `types/function_pointer_test.go`
- [ ] 9.150 Test function pointer equality and compatibility

#### AST Nodes (3 tasks)

- [ ] 9.151 Create `ast/function_pointer.go`:
  - [ ] Define `FunctionPointerType` AST node
  - [ ] Fields: `Params []*ParameterDecl`, `ReturnType *TypeAnnotation`, `OfObject bool`
  - [ ] Implement `TypeAnnotation` interface
  - [ ] Implement `String()` method
- [ ] 9.152 Define `AddressOfExpression` for `@functionName`:
  - [ ] Fields: `Operator Token`, `Operand Expression`
  - [ ] Implement `Expression` interface
  - [ ] Used to get function pointer
- [ ] 9.153 Add AST tests

#### Lexer Support (1 task)

- [ ] 9.154 Add `@` operator (AT) to lexer if not already present:
  - [ ] Used for address-of operator: `@functionName`
  - [ ] Update token types

#### Parser Support (4 tasks)

- [ ] 9.155 Extend `parseTypeAnnotation()` to handle function pointer types:
  - [ ] Detect `function(` or `procedure(` in type context
  - [ ] Parse parameter list
  - [ ] Parse optional return type
  - [ ] Parse optional `of object` clause
  - [ ] Return `FunctionPointerType` node
- [ ] 9.156 Implement address-of operator parsing:
  - [ ] Detect `@` prefix in expression
  - [ ] Parse target identifier (function/procedure name)
  - [ ] Return `AddressOfExpression` node
- [ ] 9.157 Add parser tests for function pointer types
- [ ] 9.158 Add parser tests for `@` operator

#### Semantic Analysis (5 tasks)

- [ ] 9.159 Create `semantic/function_pointer_analyzer.go`:
  - [ ] Analyze function pointer type declarations
  - [ ] Validate signatures (no duplicate param names, valid types)
  - [ ] Register function pointer types in type environment
- [ ] 9.160 Implement address-of expression analysis:
  - [ ] Resolve target function/procedure
  - [ ] Create function pointer value with signature
  - [ ] Type is `FunctionPointerType` matching target signature
  - [ ] For methods, create `MethodPointerType`
- [ ] 9.161 Implement function pointer assignment validation:
  - [ ] Check signatures are compatible
  - [ ] Allow assignment: `var f: TFunc; f := @MyFunc;`
  - [ ] Check method pointers match `of object` requirement
- [ ] 9.162 Implement function pointer call validation:
  - [ ] `functionPointerVar(args)` syntax
  - [ ] Validate argument types match parameter types
  - [ ] Infer return type from function pointer type
- [ ] 9.163 Add semantic tests for function pointers

#### Interpreter Support (6 tasks)

- [ ] 9.164 Create runtime representation in `interp/value.go`:
  - [ ] Define `FunctionPointerValue` struct
  - [ ] Fields: `Function *ast.FunctionDecl`, `Closure *Environment`
  - [ ] For method pointers: add `SelfObject Value`
- [ ] 9.165 Implement address-of operator evaluation:
  - [ ] Look up function/procedure in environment
  - [ ] Create `FunctionPointerValue` wrapping it
  - [ ] Capture current environment for closures (if needed later)
  - [ ] For methods, capture `Self` object
- [ ] 9.166 Implement function pointer call execution:
  - [ ] Evaluate function pointer expression
  - [ ] Evaluate arguments
  - [ ] Call the wrapped function with arguments
  - [ ] For method pointers, bind `Self` before calling
  - [ ] Return result
- [ ] 9.167 Implement function pointer assignment:
  - [ ] Store `FunctionPointerValue` in variable
  - [ ] Validate type compatibility at assignment time
- [ ] 9.168 Add tests in `interp/function_pointer_test.go`
- [ ] 9.169 Test passing function pointers as parameters

#### Testing & Fixtures (3 tasks)

- [ ] 9.170 Create test scripts in `testdata/function_pointers/`:
  - [ ] `basic_function_pointer.dws` - Simple function pointer usage
  - [ ] `callback.dws` - Pass function pointer as callback
  - [ ] `method_pointer.dws` - Method pointers with `of object`
  - [ ] `sort_with_comparator.dws` - Custom sort with comparator function
  - [ ] Expected outputs
- [ ] 9.171 Add CLI integration tests
- [ ] 9.172 Document function pointer limitations (if any)

---

### External Function Registration / FFI (HIGH PRIORITY)

**Summary**: Implement Foreign Function Interface (FFI) to register Go functions callable from DWScript. Enables DWScript scripts to access Go ecosystem.

**Example** (Go side):
```go
interp.RegisterFunction("HttpGet", func(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    return string(body), nil
})
```

**Example** (DWScript side):
```pascal
var html := HttpGet('https://example.com');
PrintLn(html);
```

**Reference**: docs/missing-features-recommendations.md lines 236-268

#### Public API Design (5 tasks)

- [ ] 9.173 Create `pkg/dwscript/ffi.go`:
  - [ ] Define `ExternalFunction` interface
  - [ ] Define `FunctionSignature` struct with param types and return type
  - [ ] Define `RegisterFunction(name string, fn interface{}) error` method on `Engine`
  - [ ] Validate function signature at registration time
  - [ ] Store in registry map
- [ ] 9.174 Define type mapping rules (Go ↔ DWScript):
  - [ ] `int`, `int32`, `int64` ↔ `Integer`
  - [ ] `float64` ↔ `Float`
  - [ ] `string` ↔ `String`
  - [ ] `bool` ↔ `Boolean`
  - [ ] `[]T` ↔ `array of T`
  - [ ] `map[string]T` ↔ `record` or associative array
  - [ ] `error` ↔ exception (raise on error)
- [ ] 9.175 Define calling conventions:
  - [ ] Go function receives DWScript arguments as `[]interface{}`
  - [ ] Returns `(interface{}, error)`
  - [ ] Error return raises DWScript exception
  - [ ] Support variadic Go functions
- [ ] 9.176 Add tests for API design
- [ ] 9.177 Document FFI in `docs/ffi.md`

#### Type Marshaling (8 tasks)

- [ ] 9.178 Create `interp/marshal.go`:
  - [ ] Implement `MarshalToGo(dwsValue Value) (interface{}, error)`
  - [ ] Convert DWScript values to Go values
  - [ ] Handle all primitive types
  - [ ] Handle arrays, records, objects
- [ ] 9.179 Implement `MarshalToDWS(goValue interface{}) (Value, error)`:
  - [ ] Convert Go values to DWScript values
  - [ ] Use reflection to inspect Go types
  - [ ] Handle primitives, slices, maps, structs
- [ ] 9.180 Implement integer marshaling:
  - [ ] `int`, `int32`, `int64` → `IntegerValue`
  - [ ] `IntegerValue` → `int64`
- [ ] 9.181 Implement float marshaling:
  - [ ] `float64` → `FloatValue`
  - [ ] `FloatValue` → `float64`
- [ ] 9.182 Implement string marshaling:
  - [ ] `string` → `StringValue`
  - [ ] `StringValue` → `string`
- [ ] 9.183 Implement bool marshaling:
  - [ ] `bool` → `BoolValue`
  - [ ] `BoolValue` → `bool`
- [ ] 9.184 Implement array marshaling:
  - [ ] `[]interface{}` ↔ `ArrayValue`
  - [ ] Element-wise conversion
- [ ] 9.185 Add marshaling tests

#### Function Registration (6 tasks)

- [ ] 9.186 Create `interp/external_functions.go`:
  - [ ] Define `ExternalFunctionRegistry` struct
  - [ ] Implement `Register(name string, fn interface{}) error`
  - [ ] Use reflection to extract Go function signature
  - [ ] Validate signature (supported types only)
  - [ ] Store function with wrapper
- [ ] 9.187 Implement signature extraction:
  - [ ] Use `reflect.TypeOf(fn)` to get function type
  - [ ] Extract parameter types
  - [ ] Extract return types (support 0-2 returns: value, error)
  - [ ] Map Go types to DWScript types
- [ ] 9.188 Create function call wrapper:
  - [ ] Wrapper accepts DWScript arguments
  - [ ] Marshals DWScript → Go
  - [ ] Calls Go function via reflection
  - [ ] Marshals Go return → DWScript
  - [ ] Handles errors (convert to exceptions)
- [ ] 9.189 Integrate registry with interpreter:
  - [ ] Lookup external functions during function calls
  - [ ] Call wrapper instead of DWScript function
- [ ] 9.190 Add tests for registration
- [ ] 9.191 Test function call execution

#### Error Handling (4 tasks)

- [ ] 9.192 Implement error marshaling:
  - [ ] Go `error` return → DWScript exception
  - [ ] Raise exception with error message
  - [ ] Preserve stack trace across boundary
- [ ] 9.193 Handle panics in Go functions:
  - [ ] Recover from panics in wrapper
  - [ ] Convert panic to DWScript exception
  - [ ] Include panic message and stack
- [ ] 9.194 Add tests for error handling
- [ ] 9.195 Test panic recovery

#### Advanced Features (6 tasks)

- [ ] 9.196 Support variadic Go functions:
  - [ ] Detect `...` parameter in Go signature
  - [ ] Accept variable number of DWScript arguments
  - [ ] Pack into slice for Go function
- [ ] 9.197 Support optional parameters:
  - [ ] Default values in DWScript
  - [ ] Map to Go function overloads or optional args
- [ ] 9.198 Support by-reference parameters:
  - [ ] `var` parameters in DWScript
  - [ ] Pointers in Go
  - [ ] Sync changes back to DWScript after call
- [ ] 9.199 Support registering Go methods:
  - [ ] Methods on Go structs
  - [ ] Bind receiver automatically
- [ ] 9.200 Support callback functions:
  - [ ] DWScript function pointers passed to Go
  - [ ] Go can call back into DWScript
  - [ ] Handle re-entrancy
- [ ] 9.201 Add tests for advanced features

#### Documentation and Examples (3 tasks)

- [ ] 9.202 Create `docs/ffi-guide.md`:
  - [ ] Complete guide to FFI usage
  - [ ] Type mapping table
  - [ ] Registration examples
  - [ ] Error handling guide
  - [ ] Best practices
- [ ] 9.203 Create example in `examples/ffi/`:
  - [ ] Go program that registers functions
  - [ ] DWScript script that calls them
  - [ ] Demonstrate various types and features
- [ ] 9.204 Add API documentation to `pkg/dwscript/`

#### Testing & Fixtures (3 tasks)

- [ ] 9.205 Create test scripts in `testdata/ffi/`:
  - [ ] `basic_ffi.dws` - Call simple Go functions
  - [ ] `array_passing.dws` - Pass arrays to Go
  - [ ] `error_handling.dws` - Handle Go errors
  - [ ] Expected outputs
- [ ] 9.206 Create Go test suite for FFI
- [ ] 9.207 Add integration tests calling real Go functions

---

### Lambdas/Anonymous Methods (MEDIUM PRIORITY)

**Summary**: Implement lambda expressions and anonymous methods for inline function definitions. Enables functional programming patterns and cleaner callback code.

**Example**:
```pascal
var numbers := [1, 2, 3, 4, 5];
var doubled := Map(numbers, lambda(x: Integer): Integer begin Result := x * 2; end);
PrintLn(doubled); // [2, 4, 6, 8, 10]
```

**Reference**: docs/missing-features-recommendations.md lines 274-277

**Dependencies**: Requires Function Pointers (tasks 9.146-9.172)

#### AST Nodes (3 tasks)

- [ ] 9.208 Create `ast/lambda.go`:
  - [ ] Define `LambdaExpression` struct implementing `Expression`
  - [ ] Fields: `Params []*ParameterDecl`, `ReturnType *TypeAnnotation`, `Body *Block`
  - [ ] Implement `String()` method
  - [ ] Implement `Expression` interface
- [ ] 9.209 Add short-hand lambda syntax support (optional):
  - [ ] `lambda(x) => x * 2` (single expression)
  - [ ] Desugar to full lambda with begin/end
- [ ] 9.210 Add AST tests

#### Lexer Support (1 task)

- [ ] 9.211 Add `lambda` keyword to lexer:
  - [ ] Update token types
  - [ ] Add to keyword map
  - [ ] Test lexing lambda expressions

#### Parser Support (4 tasks)

- [ ] 9.212 Implement lambda parsing in `parser/expressions.go`:
  - [ ] Detect `lambda` keyword
  - [ ] Parse parameter list (with types)
  - [ ] Parse optional return type
  - [ ] Parse body (begin/end block or single expression)
  - [ ] Return `LambdaExpression` node
- [ ] 9.213 Implement short-hand lambda parsing (if supported):
  - [ ] `lambda(params) => expression`
  - [ ] Desugar to full lambda
- [ ] 9.214 Handle lambda in expression context:
  - [ ] Lambda can appear anywhere expression is expected
  - [ ] Precedence handling
- [ ] 9.215 Add parser tests for lambdas

#### Semantic Analysis (5 tasks)

- [ ] 9.216 Create `semantic/lambda_analyzer.go`:
  - [ ] Analyze lambda expressions
  - [ ] Create new scope for lambda parameters
  - [ ] Analyze lambda body in nested scope
  - [ ] Infer return type from body if not specified
  - [ ] Type is `FunctionPointerType` matching signature
- [ ] 9.217 Implement closure capture analysis:
  - [ ] Identify variables from outer scopes used in lambda
  - [ ] Mark them for capture
  - [ ] Validate captured variables are accessible
- [ ] 9.218 Implement lambda type inference:
  - [ ] If parameter types not specified, try to infer from context
  - [ ] If return type not specified, infer from body
  - [ ] Report error if inference fails
- [ ] 9.219 Add semantic tests for lambdas
- [ ] 9.220 Test closure capture and type inference

#### Interpreter Support (6 tasks)

- [ ] 9.221 Implement closure representation in `interp/value.go`:
  - [ ] Extend `FunctionPointerValue` to store captured variables
  - [ ] Create `ClosureValue` struct
  - [ ] Store environment snapshot at lambda creation time
- [ ] 9.222 Implement lambda evaluation:
  - [ ] Evaluate lambda expression
  - [ ] Capture current environment (closure)
  - [ ] Create `ClosureValue` with parameters, body, and captured environment
  - [ ] Return closure as function pointer value
- [ ] 9.223 Implement closure invocation:
  - [ ] When closure is called, create new environment
  - [ ] Bind parameters to arguments
  - [ ] Restore captured environment
  - [ ] Execute body
  - [ ] Return result
- [ ] 9.224 Handle variable capture:
  - [ ] Copy values of captured variables (value semantics)
  - [ ] Or use references (reference semantics) - decide based on DWScript
  - [ ] Test both approaches
- [ ] 9.225 Add tests in `interp/lambda_test.go`
- [ ] 9.226 Test nested lambdas and complex closures

#### Higher-Order Function Support (4 tasks)

- [ ] 9.227 Implement built-in higher-order functions:
  - [ ] `Map(array, lambda)` - Transform array elements
  - [ ] `Filter(array, lambda)` - Filter array by predicate
  - [ ] `Reduce(array, lambda, initial)` - Reduce array to single value
  - [ ] `ForEach(array, lambda)` - Execute lambda for each element
- [ ] 9.228 Add tests for higher-order functions
- [ ] 9.229 Create examples using lambdas with higher-order functions
- [ ] 9.230 Document higher-order functions in `docs/builtins.md`

#### Testing & Fixtures (3 tasks)

- [ ] 9.231 Create test scripts in `testdata/lambdas/`:
  - [ ] `basic_lambda.dws` - Simple lambda usage
  - [ ] `closure.dws` - Variable capture
  - [ ] `higher_order.dws` - Map, Filter, Reduce examples
  - [ ] `nested_lambda.dws` - Nested lambdas
  - [ ] Expected outputs
- [ ] 9.232 Add CLI integration tests
- [ ] 9.233 Document lambda syntax in `docs/lambdas.md`

---

### Helpers (Class/Record/Type) (MEDIUM PRIORITY)

**Summary**: Implement helper types to extend existing types with additional methods without modifying the original type declaration.

**Example**:
```pascal
type
  TStringHelper = record helper for String
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

#### Type System (3 tasks)

- [ ] 9.234 Define `HelperType` in `types/types.go`:
  - [ ] Fields: `Name string`, `ForType Type`, `Methods []*FunctionType`
  - [ ] Implement `Type` interface
  - [ ] `TypeKind()` returns `TypeKindHelper`
  - [ ] `String()` returns `record helper for TypeName`
- [ ] 9.235 Implement helper method resolution:
  - [ ] When accessing method on a type, check for helpers
  - [ ] Helpers extend the type's method set
  - [ ] Prioritize: instance methods > helper methods
- [ ] 9.236 Add tests for helper types

#### AST Nodes (2 tasks)

- [ ] 9.237 Create `ast/helper.go`:
  - [ ] Define `HelperDeclaration` struct
  - [ ] Fields: `Name *Identifier`, `HelperKind string` (class/record), `ForType *TypeAnnotation`, `Methods []*FunctionDecl`
  - [ ] Implement `Node` interface
  - [ ] Implement `String()` method
- [ ] 9.238 Add AST tests

#### Lexer Support (1 task)

- [ ] 9.239 Add `helper` keyword to lexer:
  - [ ] Update token types
  - [ ] Add to keyword map

#### Parser Support (3 tasks)

- [ ] 9.240 Implement helper parsing in `parser/type_declarations.go`:
  - [ ] Detect `record helper for` or `class helper for` pattern
  - [ ] Parse helper name
  - [ ] Parse `for` keyword and target type
  - [ ] Parse method declarations
  - [ ] Expect `end;`
  - [ ] Return `HelperDeclaration` node
- [ ] 9.241 Add parser tests for helpers
- [ ] 9.242 Test class helpers and record helpers separately

#### Semantic Analysis (4 tasks)

- [ ] 9.243 Create `semantic/helper_analyzer.go`:
  - [ ] Analyze helper declarations
  - [ ] Resolve target type (`for` type)
  - [ ] Validate helper methods (must have `Self` of target type)
  - [ ] Register helper in type environment
- [ ] 9.244 Implement helper method resolution:
  - [ ] When analyzing member access on a type, check for applicable helpers
  - [ ] Add helper methods to type's method set
  - [ ] Resolve method name conflicts (prefer instance methods)
- [ ] 9.245 Implement `Self` binding in helper methods:
  - [ ] `Self` refers to the instance of the target type
  - [ ] Type of `Self` is the helper's target type
- [ ] 9.246 Add semantic tests for helpers

#### Interpreter Support (4 tasks)

- [ ] 9.247 Implement helper method dispatch:
  - [ ] When calling method on object, check for helper methods
  - [ ] Bind `Self` to the target object
  - [ ] Execute helper method with `Self` bound
- [ ] 9.248 Implement helper method storage:
  - [ ] Store helpers in registry indexed by target type
  - [ ] Lookup helpers at method call time
- [ ] 9.249 Add tests in `interp/helper_test.go`
- [ ] 9.250 Test helper method calls and `Self` binding

#### Testing & Fixtures (3 tasks)

- [ ] 9.251 Create test scripts in `testdata/helpers/`:
  - [ ] `string_helper.dws` - String helper with methods
  - [ ] `record_helper.dws` - Record helper example
  - [ ] `class_helper.dws` - Class helper example
  - [ ] Expected outputs
- [ ] 9.252 Add CLI integration tests
- [ ] 9.253 Document helpers in `docs/helpers.md`

---

### DateTime Functions (MEDIUM PRIORITY)

**Summary**: Implement comprehensive date/time functionality including current date/time, formatting, parsing, and arithmetic operations.

**Reference**: docs/missing-features-recommendations.md lines 289-292

#### Type System (2 tasks)

- [ ] 9.254 Define `TDateTime` type:
  - [ ] Internal representation: float (days since 1899-12-30, like Delphi)
  - [ ] Fractional part represents time
  - [ ] Add to type system as primitive type
- [ ] 9.255 Add tests for DateTime type

#### Built-in Functions - Current Date/Time (4 tasks)

- [ ] 9.256 Implement `Now(): TDateTime`:
  - [ ] Returns current date and time
  - [ ] Use Go's `time.Now()`
  - [ ] Convert to TDateTime format
- [ ] 9.257 Implement `Date(): TDateTime`:
  - [ ] Returns current date (time part = 0.0)
- [ ] 9.258 Implement `Time(): TDateTime`:
  - [ ] Returns current time (date part = 0.0)
- [ ] 9.259 Add tests for Now/Date/Time

#### Built-in Functions - Date Construction (4 tasks)

- [ ] 9.260 Implement `EncodeDate(year, month, day: Integer): TDateTime`:
  - [ ] Construct date from components
  - [ ] Validate inputs (valid month, day)
  - [ ] Return TDateTime value
- [ ] 9.261 Implement `EncodeTime(hour, min, sec, msec: Integer): TDateTime`:
  - [ ] Construct time from components
  - [ ] Validate inputs
- [ ] 9.262 Implement `EncodeDateTime(y, m, d, h, min, s, ms: Integer): TDateTime`:
  - [ ] Combine date and time
- [ ] 9.263 Add tests for date construction

#### Built-in Functions - Date Extraction (4 tasks)

- [ ] 9.264 Implement `DecodeDate(dt: TDateTime; var y, m, d: Integer)`:
  - [ ] Extract year, month, day components
  - [ ] Use var parameters to return multiple values
- [ ] 9.265 Implement `DecodeTime(dt: TDateTime; var h, min, s, ms: Integer)`:
  - [ ] Extract time components
- [ ] 9.266 Implement component functions:
  - [ ] `YearOf(dt: TDateTime): Integer`
  - [ ] `MonthOf(dt: TDateTime): Integer`
  - [ ] `DayOf(dt: TDateTime): Integer`
  - [ ] `HourOf(dt: TDateTime): Integer`
  - [ ] `MinuteOf(dt: TDateTime): Integer`
  - [ ] `SecondOf(dt: TDateTime): Integer`
- [ ] 9.267 Add tests for date extraction

#### Built-in Functions - Formatting (3 tasks)

- [ ] 9.268 Implement `FormatDateTime(fmt: String, dt: TDateTime): String`:
  - [ ] Support format specifiers: `yyyy`, `mm`, `dd`, `hh`, `nn`, `ss`
  - [ ] Example: `FormatDateTime('yyyy-mm-dd', Now())` → '2025-10-27'
  - [ ] Use Go's time formatting internally
- [ ] 9.269 Implement `DateToStr(dt: TDateTime): String`:
  - [ ] Default date format
- [ ] 9.270 Implement `TimeToStr(dt: TDateTime): String`:
  - [ ] Default time format

#### Built-in Functions - Parsing (2 tasks)

- [ ] 9.271 Implement `StrToDate(s: String): TDateTime`:
  - [ ] Parse date string
  - [ ] Support common formats
  - [ ] Raise error on invalid input
- [ ] 9.272 Implement `StrToDateTime(s: String): TDateTime`:
  - [ ] Parse date/time string

#### Built-in Functions - Date Arithmetic (3 tasks)

- [ ] 9.273 Implement date arithmetic operators:
  - [ ] `dt1 - dt2` → days difference (Float)
  - [ ] `dt + days` → new date
  - [ ] `dt - days` → new date
- [ ] 9.274 Implement `IncDay(dt: TDateTime, days: Integer): TDateTime`:
  - [ ] Add days to date
  - [ ] Similar: `IncMonth`, `IncYear`, `IncHour`, `IncMinute`, `IncSecond`
- [ ] 9.275 Implement `DaysBetween(dt1, dt2: TDateTime): Integer`:
  - [ ] Calculate difference in days
  - [ ] Similar: `HoursBetween`, `MinutesBetween`, `SecondsBetween`

#### Testing & Fixtures (2 tasks)

- [ ] 9.276 Create test scripts in `testdata/datetime/`:
  - [ ] `current_datetime.dws` - Now, Date, Time
  - [ ] `encode_decode.dws` - EncodeDate, DecodeDate
  - [ ] `formatting.dws` - FormatDateTime
  - [ ] `arithmetic.dws` - Date arithmetic
  - [ ] Expected outputs
- [ ] 9.277 Add CLI integration tests

---

### JSON Support (MEDIUM PRIORITY)

**Summary**: Implement JSON parsing and serialization for modern data interchange. Enables DWScript to work with JSON APIs and configuration files.

**Reference**: docs/missing-features-recommendations.md lines 294-297

#### Type System (2 tasks)

- [ ] 9.278 Define JSON value representation:
  - [ ] Variant type that can hold any JSON value
  - [ ] Support: null, boolean, number, string, array, object
  - [ ] Map to DWScript types where possible
- [ ] 9.279 Add tests for JSON type representation

#### Built-in Functions - Parsing (3 tasks)

- [ ] 9.280 Implement `ParseJSON(s: String): Variant`:
  - [ ] Parse JSON string
  - [ ] Return DWScript variant/dynamic type
  - [ ] Use Go's `encoding/json` internally
  - [ ] Map JSON types to DWScript types:
    - [ ] JSON object → dynamic record or map
    - [ ] JSON array → dynamic array
    - [ ] JSON primitives → corresponding DWScript types
- [ ] 9.281 Handle parsing errors:
  - [ ] Raise exception on invalid JSON
  - [ ] Include error position and message
- [ ] 9.282 Add tests for JSON parsing

#### Built-in Functions - Serialization (3 tasks)

- [ ] 9.283 Implement `ToJSON(value: Variant): String`:
  - [ ] Serialize DWScript value to JSON
  - [ ] Support records, arrays, primitives
  - [ ] Handle circular references (error or omit)
- [ ] 9.284 Implement `ToJSONFormatted(value: Variant, indent: Integer): String`:
  - [ ] Pretty-printed JSON with indentation
- [ ] 9.285 Add tests for JSON serialization

#### Built-in Functions - JSON Object Access (4 tasks)

- [ ] 9.286 Implement JSON object property access:
  - [ ] `jsonObject['propertyName']` syntax
  - [ ] Return value or nil if not found
- [ ] 9.287 Implement `JSONHasField(obj: Variant, field: String): Boolean`:
  - [ ] Check if JSON object has field
- [ ] 9.288 Implement `JSONKeys(obj: Variant): array of String`:
  - [ ] Return array of object keys
- [ ] 9.289 Implement `JSONValues(obj: Variant): array of Variant`:
  - [ ] Return array of object values

#### Built-in Functions - JSON Array Access (2 tasks)

- [ ] 9.290 Implement JSON array indexing:
  - [ ] `jsonArray[index]` syntax
  - [ ] Return element or nil if out of bounds
- [ ] 9.291 Implement `JSONLength(arr: Variant): Integer`:
  - [ ] Return array length

#### Type Mapping (2 tasks)

- [ ] 9.292 Document JSON ↔ DWScript type mapping:
  - [ ] JSON null → nil
  - [ ] JSON boolean → Boolean
  - [ ] JSON number → Integer or Float
  - [ ] JSON string → String
  - [ ] JSON array → dynamic array
  - [ ] JSON object → dynamic record or associative array
- [ ] 9.293 Handle edge cases:
  - [ ] Large numbers (beyond int64)
  - [ ] Special floats (NaN, Infinity)
  - [ ] Unicode escapes

#### Testing & Fixtures (2 tasks)

- [ ] 9.294 Create test scripts in `testdata/json/`:
  - [ ] `parse_json.dws` - Parse various JSON types
  - [ ] `to_json.dws` - Serialize to JSON
  - [ ] `json_object_access.dws` - Access object properties
  - [ ] `json_array_access.dws` - Access array elements
  - [ ] Expected outputs
- [ ] 9.295 Add CLI integration tests

---

### Improved Error Messages and Stack Traces (MEDIUM PRIORITY)

**Summary**: Enhance error reporting with better messages, stack traces, and debugging information. Improves developer experience significantly.

**Reference**: docs/missing-features-recommendations.md lines 299-302

#### Stack Trace Infrastructure (3 tasks)

- [ ] 9.296 Create `errors/stack_trace.go`:
  - [ ] Define `StackFrame` struct with `FunctionName`, `FileName`, `LineNumber`
  - [ ] Define `StackTrace` type as `[]StackFrame`
  - [ ] Implement `String()` method for formatted output
- [ ] 9.297 Implement stack trace capture in interpreter:
  - [ ] Track call stack during execution
  - [ ] Push frame on function entry
  - [ ] Pop frame on function exit
  - [ ] Capture stack on error/exception
- [ ] 9.298 Add tests for stack trace capture

#### Error Message Improvements (3 tasks)

- [ ] 9.299 Improve type error messages:
  - [ ] Before: "Type mismatch"
  - [ ] After: "Cannot assign Float to Integer variable 'count' at line 42"
  - [ ] Include expected and actual types
  - [ ] Include variable name and location
- [ ] 9.300 Improve runtime error messages:
  - [ ] Include expression that failed
  - [ ] Show values involved: "Division by zero: 10 / 0 at line 15"
  - [ ] Context from surrounding code
- [ ] 9.301 Add source code snippets to errors:
  - [ ] Show the line that caused error
  - [ ] Highlight error position with `^` or color
  - [ ] Show 1-2 lines of context

#### Exception Enhancements (2 tasks)

- [ ] 9.302 Add stack trace to exception objects:
  - [ ] Store `StackTrace` in exception
  - [ ] Display on uncaught exception
  - [ ] Format nicely for CLI output
- [ ] 9.303 Implement `GetStackTrace()` built-in:
  - [ ] Return current stack trace as string
  - [ ] Useful for logging and debugging

#### Debugging Information (2 tasks)

- [ ] 9.304 Add source position to all AST nodes:
  - [ ] Audit nodes for missing `Pos` fields
  - [ ] Add `EndPos` for better range reporting
  - [ ] Use in error messages
- [ ] 9.305 Implement call stack inspection:
  - [ ] `GetCallStack()` returns array of frame info
  - [ ] Each frame: function name, file, line
  - [ ] Accessible from DWScript code

#### Testing & Documentation (2 tasks)

- [ ] 9.306 Create test fixtures demonstrating error messages:
  - [ ] Type errors with clear messages
  - [ ] Runtime errors with stack traces
  - [ ] Exception handling with stack traces
  - [ ] Compare before/after error message quality
- [ ] 9.307 Document error message format in `docs/error-messages.md`

---

## Phase 9 Summary

**Total Tasks**: ~217 (9.91 - 9.307)
**Estimated Effort**: ~26 weeks (~6 months)

### Priority Breakdown:

**HIGH PRIORITY** (~150 tasks, ~18 weeks):
- Subrange Types: 12 tasks
- Units/Modules System: 45 tasks (CRITICAL for multi-file projects)
- Function/Method Pointers: 25 tasks
- External Function Registration (FFI): 35 tasks

**MEDIUM PRIORITY** (~67 tasks, ~8 weeks):
- Lambdas/Anonymous Methods: 30 tasks (depends on function pointers)
- Helpers: 20 tasks
- DateTime Functions: 24 tasks
- JSON Support: 18 tasks
- Improved Error Messages: 12 tasks

This comprehensive backlog brings go-dws from ~55% to ~85% feature parity with DWScript, making it production-ready for most use cases.

## Stage 10: Performance Tuning and Refactoring

### Performance Profiling

- [ ] 10.1 Create performance benchmark scripts
- [ ] 10.2 Profile lexer performance: `BenchmarkLexer`
- [ ] 10.3 Profile parser performance: `BenchmarkParser`
- [ ] 10.4 Profile interpreter performance: `BenchmarkInterpreter`
- [ ] 10.5 Identify bottlenecks using `pprof`
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

### Project Reorganization

- [x] 11.12.1 Reorganize to standard Go project layout (completed 2025-10-26):
  - [x] Create `internal/` and `pkg/` directories
  - [x] Move `ast/` → `internal/ast/` and update all imports
  - [x] Move `errors/` → `internal/errors/` and update all imports
  - [x] Move `interp/` → `internal/interp/` and update all imports
  - [x] Move `lexer/` → `internal/lexer/` and update all imports
  - [x] Move `parser/` → `internal/parser/` and update all imports
  - [x] Move `semantic/` → `internal/semantic/` and update all imports
  - [x] Move `types/` → `internal/types/` and update all imports
  - [x] Create `pkg/dwscript/` public API with Engine, Options, Result types
  - [x] Write comprehensive Godoc and examples for `pkg/dwscript/`
  - [x] Create placeholder `pkg/platform/` package (for Stage 10.15)
  - [x] Create placeholder `pkg/wasm/` package (for Stage 10.15)
  - [x] Update README.md with embedding examples
  - [x] Update CLAUDE.md with new package structure
  - [ ] Optionally refactor `cmd/dwscript` to use `pkg/dwscript/` API (future optimization)

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
```
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
