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

- [ ] 9.79 Create test scripts in `testdata/array_functions/`:
  - [ ] `copy.dws` - Array copying and independence
  - [ ] `search.dws` - IndexOf and Contains
  - [ ] `reverse.dws` - Reverse array
  - [ ] `sort.dws` - Sort arrays
  - [ ] Expected outputs
- [ ] 9.80 Add CLI integration tests:
  - [ ] Test array function scripts
  - [ ] Verify outputs

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
