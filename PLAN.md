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

**Implementation Summary**: Phase 9 extended the type system with aliases, subranges, and inline type expressions; added const declarations with semantic enforcement; implemented lambda expressions with capture semantics; enriched the standard library with essential built-in and array functions. Several major features remain in progress including the units/modules system (42/43 tasks complete), function pointers (26/27 tasks complete), and various type system enhancements. All completed features include comprehensive parser, semantic analyzer, interpreter, and CLI integration with dedicated test suites.

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

### Compound Assignment Operators (HIGH PRIORITY)

**Summary**: Implement `+=`, `-=`, `*=`, `/=` compound assignment operators. Currently these must be written as `x := x + 1;` instead of `x += 1;`.

**Example**: `count += 1;`, `total *= factor;`

**Reference**: Rosetta Code examples use compound assignments frequently.

#### AST Nodes (1 task)

- [x] 9.13 Extend assignment AST node in `ast/statements.go`:
  - [x] Add `CompoundOp Token` field (PLUS, MINUS, TIMES, DIVIDE, or empty)
  - [x] Update `String()` to show compound operators
  - [x] Distinguish between simple and compound assignment

#### Parser Support (1 task)

- [x] 9.14 Add parser support in `parser/statements.go`:
  - [x] Hook `PLUS_ASSIGN`, `MINUS_ASSIGN`, `TIMES_ASSIGN`, `DIVIDE_ASSIGN` tokens
  - [x] Removed from Pratt parser precedence table (statement-level operators only)
  - [x] Reuse existing assignment AST node, store operator kind
  - [x] Expand error recovery for unexpected compound operators
  - [x] Add parser tests for all compound operators

#### Semantic Analysis (1 task)

- [x] 9.15 Implement semantic checking in `semantic/analyzer.go`:
  - [x] Add analyzer checks mirroring simple assignment
  - [x] Type compatibility validation (lvalue type must support operator)
  - [x] Lvalue validation (left side must be assignable)
  - [x] Add semantic tests for type errors with compound operators

#### Interpreter Support (1 task)

- [x] 9.16 Extend interpreter in `interp/interpreter.go`:
  - [x] Extend interpreter evaluation to perform underlying arithmetic then assign
  - [x] Handle each operator: `x += y` becomes `x = x + y`
  - [x] Create unit tests for numeric, string (where supported), and boolean compounds
  - [x] Include negative cases for unsupported operand types

#### Testing & Fixtures (1 task)

- [x] 9.17 Create test files in `testdata/compound_ops/`:
  - [x] Basic compound assignments for all operators
  - [x] Compound assignments with complex expressions
  - [x] Type error cases
  - [x] Integration with loops and conditionals

---

### For-In Enumerator Loops (HIGH PRIORITY)

**Summary**: Implement `for variable in expression do` loop syntax for iterating over collections. Currently only `for-to` and `for-downto` are supported.

**Example**: `for ch in str do Print(ch);`, `for i in 0..9 do ...;`

**Reference**: `Range_expansion.dws`, `Range_extraction.dws`

#### AST Nodes (1 task)

- [x] 9.22 Create `ForInStatement` in `ast/control_flow.go`:
  - [x] Fields: `Variable *Identifier`, `Collection Expression`, `Body Statement`, `Token lexer.Token`, `InlineVar bool`
  - [x] Implements `Statement` interface
  - [x] `String()` returns `for <var> in <expr> do <body>`
  - [x] Add AST tests

#### Parser Support (1 task)

- [x] 9.23 Update for-loop parsing in `parser/control_flow.go`:
  - [x] Update `parseForStatement()` to detect `IDENT IN expression DO`
  - [x] Build `ForInStatement` node via new `parseForInLoop()` helper
  - [x] Preserve existing `for-to/downto` behavior (all tests pass)
  - [x] Add recovery for malformed enumerator syntax
  - [x] Add parser tests for for-in loops (5 comprehensive test cases)

#### Semantic Analysis (1 task)

- [x] 9.24 Teach semantic analysis about enumerators:
  - [x] Resolve loop variable scope (new local in loop body)
  - [x] Ensure enumerated expression is enumerable (array, set, string)
  - [x] Handle TypeAlias unwrapping for aliased enumerable types
  - [x] Emit semantic errors when expression is not enumerable
  - [x] Add targeted semantic tests (12 comprehensive tests)

#### Interpreter Support (1 task)

- [x] 9.25 Implement for-in runtime support in `interp/interpreter.go`:
  - [x] Implement iteration for arrays
  - [x] Implement iteration for sets
  - [x] Implement iteration for strings (character by character)
  - [x] Implement iteration for range expressions (if supported)
  - [x] Respect loop variable assignment semantics (value vs reference)

#### Testing & Fixtures (1 task)

- [x] 9.26 Create test files in `testdata/for_in/`:
  - [x] Basic for-in with arrays
  - [x] For-in with strings
  - [x] For-in with sets
  - [x] Nested for-in loops
  - [x] Failure cases (non-enumerable expressions)
  - [x] Adapt Rosetta Code examples

---

### Character Literal Expressions (HIGH PRIORITY)

**Summary**: Implement standalone character literals as first-class expressions. DWScript supports character literals like `'A'`, `#13` (ordinal), `#$41` (hex).

**Example**: `var ch: String := 'H';`, `case key of #13: HandleEnter; end;`

**Reference**: `Execute_HQ9+.dws` and other Rosetta Code examples.

#### AST Nodes (1 task)

- [x] 9.27 Create or extend character literal node in `ast/expressions.go`:
  - [x] Store rune value and position
  - [x] Support all forms: `'H'`, `#13`, `#$41`
  - [x] `String()` returns appropriate representation
  - [x] Add AST tests

#### Parser Support (1 task)

- [x] 9.28 Parse standalone CHAR literals in `parser/expressions.go`:
  - [x] Ensure lexer emits `CHAR` tokens correctly
  - [x] Map `CHAR` tokens to literal expression parser
  - [x] Handle all character literal forms
  - [x] Add parser tests covering char constants in case labels and assignments

#### Semantic Analysis (1 task)

- [x] 9.29 Type checking for character literals:
  - [x] Treat char literals as single-character strings (DWScript convention)
  - [x] Allow char literals where strings are expected
  - [x] Add semantic tests for char literal type checking

#### Interpreter Support (1 task)

- [x] 9.30 Runtime support for character literals:
  - [x] Update interpreter value construction for all char forms
  - [x] Ensure `'H'`, `#13`, and `#$41` forms evaluate correctly
  - [x] Handle character-to-string coercion
  - [x] Add interpreter tests

#### Testing & Fixtures (1 task)

- [x] 9.31 Create test files in `testdata/char_literals/`:
  - [x] Basic character literal assignments
  - [x] Character literals in case statements
  - [x] Ordinal and hex character literals
  - [x] Add regression tests from Rosetta Code examples

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

- [ ] 9.32 Create `pkg/dwscript/ffi.go`:
  - [ ] Define `ExternalFunction` interface
  - [ ] Define `FunctionSignature` struct with param types and return type
  - [ ] Define `RegisterFunction(name string, fn interface{}) error` method on `Engine`
  - [ ] Validate function signature at registration time
  - [ ] Store in registry map
- [ ] 9.33 Define type mapping rules (Go ↔ DWScript):
  - [ ] `int`, `int32`, `int64` ↔ `Integer`
  - [ ] `float64` ↔ `Float`
  - [ ] `string` ↔ `String`
  - [ ] `bool` ↔ `Boolean`
  - [ ] `[]T` ↔ `array of T`
  - [ ] `map[string]T` ↔ `record` or associative array
  - [ ] `error` ↔ exception (raise on error)
- [ ] 9.34 Define calling conventions:
  - [ ] Go function receives DWScript arguments as `[]interface{}`
  - [ ] Returns `(interface{}, error)`
  - [ ] Error return raises DWScript exception
  - [ ] Support variadic Go functions
- [ ] 9.35 Add tests for API design
- [ ] 9.36 Document FFI in `docs/ffi.md`

#### Type Marshaling (8 tasks)

- [ ] 9.37 Create `interp/marshal.go`:
  - [ ] Implement `MarshalToGo(dwsValue Value) (interface{}, error)`
  - [ ] Convert DWScript values to Go values
  - [ ] Handle all primitive types
  - [ ] Handle arrays, records, objects
- [ ] 9.38 Implement `MarshalToDWS(goValue interface{}) (Value, error)`:
  - [ ] Convert Go values to DWScript values
  - [ ] Use reflection to inspect Go types
  - [ ] Handle primitives, slices, maps, structs
- [ ] 9.39 Implement integer marshaling:
  - [ ] `int`, `int32`, `int64` → `IntegerValue`
  - [ ] `IntegerValue` → `int64`
- [ ] 9.40 Implement float marshaling:
  - [ ] `float64` → `FloatValue`
  - [ ] `FloatValue` → `float64`
- [ ] 9.41 Implement string marshaling:
  - [ ] `string` → `StringValue`
  - [ ] `StringValue` → `string`
- [ ] 9.42 Implement bool marshaling:
  - [ ] `bool` → `BoolValue`
  - [ ] `BoolValue` → `bool`
- [ ] 9.43 Implement array marshaling:
  - [ ] `[]interface{}` ↔ `ArrayValue`
  - [ ] Element-wise conversion
- [ ] 9.44 Add marshaling tests

#### Function Registration (6 tasks)

- [ ] 9.45 Create `interp/external_functions.go`:
  - [ ] Define `ExternalFunctionRegistry` struct
  - [ ] Implement `Register(name string, fn interface{}) error`
  - [ ] Use reflection to extract Go function signature
  - [ ] Validate signature (supported types only)
  - [ ] Store function with wrapper
- [ ] 9.46 Implement signature extraction:
  - [ ] Use `reflect.TypeOf(fn)` to get function type
  - [ ] Extract parameter types
  - [ ] Extract return types (support 0-2 returns: value, error)
  - [ ] Map Go types to DWScript types
- [ ] 9.47 Create function call wrapper:
  - [ ] Wrapper accepts DWScript arguments
  - [ ] Marshals DWScript → Go
  - [ ] Calls Go function via reflection
  - [ ] Marshals Go return → DWScript
  - [ ] Handles errors (convert to exceptions)
- [ ] 9.48 Integrate registry with interpreter:
  - [ ] Lookup external functions during function calls
  - [ ] Call wrapper instead of DWScript function
- [ ] 9.49 Add tests for registration
- [ ] 9.50 Test function call execution

#### Error Handling (4 tasks → 12 sub-tasks)

**Note**: Leverages existing exception infrastructure from Stage 8. EHost class bridges Go errors/panics to DWScript exception system. Named "EHost" (not "EGo" or "EDelphi") for future-proofing - works with any host runtime (Go, WASM, C FFI, etc.).

- [ ] 9.51 Implement error marshaling:
  - [ ] 9.51a Register `EHost` exception class in `registerBuiltinExceptions()` (`internal/interp/exceptions.go`)
    - Inherits from `Exception`
    - Field: `ExceptionClass: String` (holds Go error type name, e.g., "*fs.PathError")
    - Constructor: `Create(cls, msg)`
  - [ ] 9.51b Create `raiseGoErrorAsException(err error)` in `internal/interp/ffi_errors.go`
    - Create `EHost` instance with error message
    - Populate `ExceptionClass` field with `fmt.Sprintf("%T", err)`
    - Capture call stack from `i.callStack` (already implemented)
    - Set `i.exception` to trigger propagation
  - [ ] 9.51c Integrate error conversion into FFI call wrapper
    - Check for non-nil error return from Go function
    - Call `raiseGoErrorAsException()` to convert
    - Ensure exception propagates to DWScript caller
- [ ] 9.52 Handle panics in Go functions:
  - [ ] 9.52a Create `callExternalFunctionSafe()` wrapper with panic recovery
    - Add `defer/recover()` block
    - Detect panic type (error, string, other)
    - Extract panic message with type assertion
  - [ ] 9.52b Create `raiseGoPanicAsException(panicValue interface{})` function
    - Convert panic value to string message
    - Create `EHost` instance marked as panic (include "panic:" prefix in message)
    - Optionally include Go stack trace in message (`runtime.Stack()`)
    - Set `i.exception` to trigger propagation
  - [ ] 9.52c Document panic handling behavior
    - Add section to `docs/ffi-guide.md` (created in task 9.61)
    - Explain what happens when Go code panics
    - Best practices for writing panic-safe Go functions
- [ ] 9.53 Add tests for error handling (`internal/interp/ffi_errors_test.go`):
  - [ ] 9.53a Test Go error → exception conversion
    - Register Go function that returns error
    - Call from DWScript and verify exception raised
    - Verify exception message matches error message
    - Verify exception is catchable with try/except
  - [ ] 9.53b Test error propagation across nested calls
    - Go function errors in nested call stack
    - Verify exception propagates to top level
    - Verify call stack array is correct
  - [ ] 9.53c Test EHost exception-specific features
    - Catch `EHost` exception specifically (`on E: EHost do`)
    - Verify `ExceptionClass` field is populated with Go type
    - Verify `Message` field contains error text
- [ ] 9.54 Test panic recovery:
  - [ ] 9.54a Test panic conversion to exception
    - Go function panics with string: `panic("error message")`
    - Go function panics with error: `panic(errors.New("error"))`
    - Go function panics with other type: `panic(42)`
    - Verify all convert to catchable `EHost` exceptions
  - [ ] 9.54b Test panic propagation in nested FFI calls
    - Go function A calls Go function B (via DWScript)
    - B panics
    - Verify exception propagates correctly through call stack
  - [ ] 9.54c Test finally blocks with panics
    - Go function panics in try block
    - Verify finally block still executes (existing behavior)
    - Verify panic exception propagates after finally completes

#### Advanced Features (6 tasks)

- [ ] 9.55 Support variadic Go functions:
  - [ ] Detect `...` parameter in Go signature
  - [ ] Accept variable number of DWScript arguments
  - [ ] Pack into slice for Go function
- [ ] 9.56 Support optional parameters:
  - [ ] Default values in DWScript
  - [ ] Map to Go function overloads or optional args
- [ ] 9.57 Support by-reference parameters:
  - [ ] `var` parameters in DWScript
  - [ ] Pointers in Go
  - [ ] Sync changes back to DWScript after call
- [ ] 9.58 Support registering Go methods:
  - [ ] Methods on Go structs
  - [ ] Bind receiver automatically
- [ ] 9.59 Support callback functions:
  - [ ] DWScript function pointers passed to Go
  - [ ] Go can call back into DWScript
  - [ ] Handle re-entrancy
- [ ] 9.60 Add tests for advanced features

#### Documentation and Examples (3 tasks)

- [ ] 9.61 Create `docs/ffi-guide.md`:
  - [ ] Complete guide to FFI usage
  - [ ] Type mapping table
  - [ ] Registration examples
  - [ ] Error handling guide
  - [ ] Best practices
- [ ] 9.62 Create example in `examples/ffi/`:
  - [ ] Go program that registers functions
  - [ ] DWScript script that calls them
  - [ ] Demonstrate various types and features
- [ ] 9.63 Add API documentation to `pkg/dwscript/`

#### Testing & Fixtures (3 tasks)

- [ ] 9.64 Create test scripts in `testdata/ffi/`:
  - [ ] `basic_ffi.dws` - Call simple Go functions
  - [ ] `array_passing.dws` - Pass arrays to Go
  - [ ] `error_handling.dws` - Handle Go errors
  - [ ] Expected outputs
- [ ] 9.65 Create Go test suite for FFI
- [ ] 9.66 Add integration tests calling real Go functions

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

### Format Function Testing (DEFERRED)

**Summary**: Create comprehensive test fixtures for the Format() built-in function. Deferred from task 9.52 due to DWScript's set literal syntax `[...]` conflicting with Format's array parameter requirements.

#### Task Details (1 task)

- [ ] 9.156 Create Format function test fixtures:
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

### 9.8 Record Literals with Named Fields (12 tasks)

**Context**: The Death_Star.dws Rosetta example uses record initialization with named fields:

```pascal
const big : TSphere = (cx: 20; cy: 20; cz: 0; r: 20);
```
The parser currently fails with "expected type expression, got 20" because it doesn't recognize the `field: value` syntax inside parentheses.

**Current Error**: Parser treats `(cx: 20; ...)` as a type expression or grouped expression, not as a record literal with field initializers.

#### AST Nodes (2 tasks)

- [x] 9.169 Create `RecordLiteralExpression` in `ast/records.go`:
  - [x] Fields: `Token lexer.Token`, `TypeName *Identifier`, `Fields []*FieldInitializer`
  - [x] `FieldInitializer` struct with `Name *Identifier`, `Value Expression`
  - [x] Implements `Expression` interface
  - [x] Support semicolon or comma as field separator
  - [x] `String()` returns `TypeName(field1: value1; field2: value2)`
  - [x] Handle optional type name (can be inferred from context)

- [x] 9.170 Add AST tests for record literals:
  - [x] Test simple record: `(x: 10; y: 20)`
  - [x] Test with type name: `TPoint(x: 10; y: 20)`
  - [x] Test nested records: `TRect(TopLeft: (x: 0; y: 0); BottomRight: (x: 10; y: 10))`
  - [x] Test with expressions: `TSphere(cx: x+5; cy: y*2; r: radius)`
  - [x] Test with negative numbers: `(x: -50; y: 30)`
  - [x] 8 comprehensive test cases in `ast/records_test.go`

#### Parser Support (4 tasks)

- [x] 9.171 Update `parseGroupedExpression` to detect record literals:
  - [x] After parsing first expression, check for `:` token
  - [x] If found, this is a record literal field initializer
  - [x] Backtrack or reparse first expression as identifier (field name)
  - [x] Delegate to `parseRecordLiteral()` helper
  - [x] If no `:`, continue as grouped expression (existing behavior)

- [x] 9.172 Implement `parseRecordLiteral()` in `parser/records.go`:
  - [x] Parse `fieldName : expression` pairs
  - [x] Accept `;` or `,` as separators (DWScript allows both)
  - [x] Handle trailing separator (optional)
  - [x] Expect closing `)` parenthesis
  - [x] Return `RecordLiteralExpression` node
  - [x] Add error recovery for missing colons, values, or closing paren

- [x] 9.173 Handle typed record literals `TypeName(field: value)`:
  - [x] Check if expression before `(` is an identifier
  - [x] Parse as call expression initially
  - [x] If arguments contain `:`, convert to record literal
  - [x] Alternative: add `parseRecordLiteralWithType()` helper
  - [x] Preserve backward compatibility with function calls

- [x] 9.174 Add parser tests for record literals:
  - [x] Test anonymous record: `(x: 10; y: 20)`
  - [x] Test typed record: `TPoint(x: 10; y: 20)`
  - [x] Test with semicolons: `(a: 1; b: 2; c: 3)`
  - [x] Test with commas: `(a: 1, b: 2, c: 3)`
  - [x] Test nested: `TRect(TopLeft: (x: 0; y: 0))`
  - [x] Test error cases: missing colon, missing value, unclosed paren
  - [x] 10 test cases in `internal/parser/records_test.go`

#### Semantic Analysis (2 tasks)

- [x] 9.175 Create `analyzeRecordLiteral` in `semantic/analyze_records.go`:
  - [x] Resolve record type from `TypeName` or context (assignment target)
  - [x] Validate each field name exists in record type definition
  - [x] Check field value type matches declared field type
  - [x] Report error for duplicate field initializers
  - [x] Report error for missing required fields (if no defaults)
  - [x] Allow fields in any order (not required to match declaration order)
  - [x] Set inferred type on AST node
  - [x] Support anonymous record literals with type context
  - [x] Support typed record literals with explicit TypeName

- [x] 9.176 Add semantic tests for record literals:
  - [x] Test valid record with all fields: `TSphere(cx: 20; cy: 20; cz: 0; r: 20)`
  - [x] Test type error: wrong field type `TPoint(x: "hello"; y: 20)`
  - [x] Test error: unknown field name `TPoint(x: 10; z: 20)`
  - [x] Test error: duplicate field `TPoint(x: 10; x: 20)`
  - [x] Test partial initialization (if allowed by DWScript)
  - [x] Tests integrated into existing `internal/semantic/records_test.go`

#### Interpreter Support (2 tasks)

- [x] 9.177 Implement `evalRecordLiteral` in `interp/records.go`:
  - [x] Create new record instance
  - [x] Evaluate each field value expression
  - [x] Assign evaluated values to corresponding record fields
  - [x] Initialize unspecified fields to zero values
  - [x] Return `RecordValue` with all fields populated
  - [x] Handle nested record literals recursively
  - [x] Support anonymous literals via type context in var/const declarations

- [x] 9.178 Add interpreter tests for record literals:
  - [x] Test `var p := TPoint(x: 10; y: 20); PrintLn(p.x);` outputs `10`
  - [x] Test nested: `var r := TRect(TopLeft: (x: 0; y: 0));`
  - [x] Test with expressions: `var s := TSphere(cx: 5*4; r: 10+5);`
  - [x] Test field access after initialization
  - [x] Created `internal/interp/record_literals_test.go` with 8 comprehensive test cases

#### Testing & Documentation (2 tasks)

- [x] 9.179 Create test files in `testdata/record_literals/`:
  - [x] Tests integrated into `internal/interp/record_literals_test.go`
  - [x] Death_Star.dws example tests included
  - [x] All test scenarios covered (basic, nested, const, typed, anonymous)

- [x] 9.180 Enable Death_Star.dws record literal tests:
  - [x] Verify `const big : TSphere = (cx: 20; cy: 20; cz: 0; r: 20);` parses
  - [x] Verify `const small : TSphere = (cx: 7; cy: 7; cz: -10; r: 15);` parses
  - [x] All parser and integration tests pass successfully
  - [x] Death_Star.dws still needs other features (array literals, type inference, Exit statement)

---

### 9.9 Array Literal Expressions (10 tasks)

**Context**: Death_Star.dws uses inline array initialization:
```pascal
var light : TVector = [-50.0, 30, 50];
```
The parser currently fails with "expected next token to be SEMICOLON, got EQ" because it treats the type declaration as complete and doesn't expect an initializer.

**Current Error**: Parser doesn't recognize `[expr, expr, ...]` as an array literal expression.

#### AST Nodes (2 tasks)

- [x] 9.181 Create `ArrayLiteralExpression` in `ast/arrays.go`:
  - [x] Fields: `Token lexer.Token`, `Elements []Expression`
  - [x] Implements `Expression` interface
  - [x] Support empty arrays: `[]`
  - [x] Support nested arrays: `[[1, 2], [3, 4]]`
  - [x] `String()` returns `[elem1, elem2, ...]`

- [x] 9.182 Add AST tests for array literals:
  - [x] Test simple: `[1, 2, 3]`
  - [x] Test with expressions: `[x+1, y*2, z-3]`
  - [x] Test nested: `[[1, 2], [3, 4]]`
  - [x] Test with negative numbers: `[-50.0, 30, 50]`
  - [x] Test empty: `[]`
  - [x] 6 test cases in `ast/arrays_test.go`

#### Parser Support (3 tasks)

- [x] 9.183 Add prefix parse function for `LBRACKET`:
  - [x] Register `p.prefixParseFns[lexer.LBRACKET] = p.parseArrayLiteral`
  - [x] Distinguish from index access (which is infix)
  - [x] Array literal is prefix when `[` starts an expression

- [x] 9.184 Implement `parseArrayLiteral()` in `parser/arrays.go`:
  - [x] Parse opening `[` bracket
  - [x] Parse comma-separated element expressions (0 or more)
  - [x] Handle trailing comma (optional)
  - [x] Expect closing `]` bracket
  - [x] Return `ArrayLiteralExpression` node
  - [x] Add error recovery for malformed syntax

- [x] 9.185 Add parser tests for array literals:
  - [x] Test `[1, 2, 3]`
  - [x] Test `[-50.0, 30, 50]` (Death_Star.dws case)
  - [x] Test nested: `[[1, 2], [3, 4]]`
  - [x] Test empty: `[]`
  - [x] Test with expressions: `[x+1, Length(s), 42]`
  - [x] Test error cases: missing comma, unclosed bracket
  - [x] 8 test cases in `internal/parser/arrays_test.go`

#### Semantic Analysis (2 tasks)

- [x] 9.186 Create `analyzeArrayLiteral` in `semantic/analyze_arrays.go`:
  - [x] Infer element type from first element (if any)
  - [x] Validate all elements have compatible types
  - [x] Report error if mixed incompatible types (e.g., `[1, "hello"]`)
  - [x] Allow numeric type promotion (Integer → Float if needed)
  - [x] Set inferred type: `array of <ElementType>`
  - [x] Handle empty arrays (require explicit type context)

- [x] 9.187 Add semantic tests for array literals:
  - [x] Test homogeneous: `[1, 2, 3]` → `array of Integer`
  - [x] Test type promotion: `[1, 2.5, 3]` → `array of Float`
  - [x] Test error: mixed types `[1, "hello"]`
  - [x] Test nested: `[[1, 2], [3, 4]]` → `array of array of Integer`
  - [x] 5 test cases in `internal/semantic/arrays_test.go`

#### Interpreter Support (2 tasks)

- [x] 9.188 Implement `evalArrayLiteral` in `interp/array.go`:
  - [x] Evaluate each element expression in order
  - [x] Collect evaluated values into slice
  - [x] Create `ArrayValue` with elements
  - [x] Handle empty arrays (create with zero length)
  - [x] Handle nested arrays recursively

- [x] 9.189 Add interpreter tests for array literals:
  - [x] Test `var arr := [1, 2, 3]; PrintLn(arr[1]);` outputs `2`
  - [x] Test `var light := [-50.0, 30, 50]; PrintLn(FloatToStr(light[0]));` outputs `-50`
  - [x] Test nested: `var matrix := [[1, 2], [3, 4]]; PrintLn(matrix[0][1]);` outputs `2`
  - [x] 5 test cases in `internal/interp/arrays_test.go`

#### Testing & Documentation (1 task)

- [x] 9.190 Create test files and enable Death_Star.dws:
  - [x] Create `testdata/array_literals/array_literal_basic.dws`
  - [x] Create `testdata/array_literals/array_literal_nested.dws`
  - [x] Verify `var light : TVector = [-50.0, 30, 50];` parses and executes
  - [x] Add integration test in `pkg/dwscript/rosetta_examples_test.go`

---

### 9.10 Type Inference in Variable Declarations (8 tasks)

**Context**: DWScript allows variable declarations without explicit type annotations when an initializer is present:
```pascal
var zsq = sph.r * sph.r - (x * x + y * y);
```
The parser currently requires type annotations and fails with "expected next token to be SEMICOLON, got EQ".

**Current Grammar**: `var identifier : Type := initializer;`
**Required Grammar**: `var identifier := initializer;` OR `var identifier = initializer;`

#### Parser Support (3 tasks)

- [ ] 9.191 Update `parseVarStatement` to support type inference:
  - [ ] After parsing identifier, check next token
  - [ ] If `:`, parse explicit type annotation (existing behavior)
  - [ ] If `:=` or `=`, skip type annotation (infer from initializer)
  - [ ] Both `=` and `:=` should work for type-inferred vars
  - [ ] Add error if neither type nor initializer provided

- [ ] 9.192 Update `VarStatement` AST node to allow nil type:
  - [ ] Make `Type *TypeAnnotation` optional (can be nil)
  - [ ] Add `Inferred bool` flag to indicate type was inferred
  - [ ] Update `String()` to handle missing type: `var name = value`
  - [ ] Ensure backward compatibility with explicit types

- [ ] 9.193 Add parser tests for type inference:
  - [ ] Test `var x = 42;` (integer inference)
  - [ ] Test `var s = "hello";` (string inference)
  - [ ] Test `var f = 3.14;` (float inference)
  - [ ] Test `var arr = [1, 2, 3];` (array inference)
  - [ ] Test `var rec = (x: 10; y: 20);` (record inference)
  - [ ] Test error: `var x;` (no type, no initializer)
  - [ ] Test backward compatibility: `var x: Integer := 42;` still works
  - [ ] 8 test cases in `internal/parser/variables_test.go`

#### Semantic Analysis (2 tasks)

- [ ] 9.194 Update `analyzeVarStatement` to handle inference:
  - [ ] If `Type` is nil, analyze initializer first
  - [ ] Get inferred type from initializer expression
  - [ ] Set variable's type to inferred type in symbol table
  - [ ] Report error if initializer has no determinable type
  - [ ] Report error if initializer is nil/untyped constant without context

- [ ] 9.195 Add semantic tests for type inference:
  - [ ] Test `var x = 42;` infers `Integer`
  - [ ] Test `var f = 3.14;` infers `Float`
  - [ ] Test `var s = "hello";` infers `String`
  - [ ] Test `var arr = [1, 2, 3];` infers `array of Integer`
  - [ ] Test error: `var x = [];` (can't infer from empty array)
  - [ ] 6 test cases in `internal/semantic/variables_test.go`

#### Interpreter Support (1 task)

- [ ] 9.196 Update `evalVarStatement` for type inference:
  - [ ] No interpreter changes needed (type already set by semantic analyzer)
  - [ ] Add integration tests to verify execution
  - [ ] Test `var x = 42; PrintLn(x);` outputs `42`
  - [ ] Test `var arr = [1, 2, 3]; PrintLn(arr[1]);` outputs `2`

#### Testing & Documentation (2 tasks)

- [ ] 9.197 Create test files in `testdata/type_inference/`:
  - [ ] `type_inference_basic.dws` - Simple cases
  - [ ] `type_inference_arrays.dws` - Array literals
  - [ ] `type_inference_records.dws` - Record literals
  - [ ] Expected output files

- [ ] 9.198 Enable Death_Star.dws type inference:
  - [ ] Verify `var zsq = sph.r * sph.r - (x * x + y * y);` parses
  - [ ] Add integration test in `pkg/dwscript/rosetta_examples_test.go`

---

### 9.11 Exit Statement with Return Value (8 tasks)

**Context**: Death_Star.dws uses `Exit` with a return value:
```pascal
if (zsq < 0) then Exit False;
```
The parser currently fails with "no prefix parse function for FALSE" because it doesn't expect an expression after `Exit`.

**Current Implementation**: `Exit` is parsed as a standalone statement (no return value).
**Required**: `Exit <expression>` to return a value from a function.

#### AST Nodes (1 task)

- [x] 9.199 Update `ExitStatement` in `ast/statements.go`:
  - [x] Add optional `ReturnValue Expression` field (can be nil)
  - [x] Update `String()` to show value: `Exit value` or just `Exit`
  - [x] Distinguish from `ReturnStatement` (used in procedures/functions)
  - [x] DWScript uses `Exit` for early return, `Result :=` for setting return value

#### Parser Support (2 tasks)

- [x] 9.200 Update `parseExitStatement` to handle return values:
  - [x] After `Exit` keyword, check if next token starts an expression
  - [x] If yes, parse expression as return value
  - [x] If semicolon or end of line, no return value (existing behavior)
  - [x] Support both `Exit value` and `Exit(value)` syntax
  - [x] Add tests for single-line: `Exit False;`

- [x] 9.201 Add parser tests for Exit with values:
  - [x] Test `Exit;` (no value, existing behavior)
  - [x] Test `Exit False;` (boolean return)
  - [x] Test `Exit 42;` (integer return)
  - [x] Test `Exit x + y;` (expression return)
  - [x] Test `Exit(value);` (parenthesized, like function call)
  - [x] 6 test cases in `internal/parser/statements_test.go`

#### Semantic Analysis (2 tasks)

- [x] 9.202 Update `analyzeExitStatement` to check return types:
  - [x] Get current function's declared return type
  - [x] If `ReturnValue` is present, validate type matches function return type
  - [x] Report error if types don't match
  - [x] Report error if `Exit value` used outside a function
  - [ ] Report error if `Exit` (no value) used in function expecting return type *(DWScript allows this; kept permissive to preserve existing behavior)*

- [x] 9.203 Add semantic tests for Exit statements:
  - [x] Test valid: `function GetValue: Integer; begin Exit 42; end;`
  - [x] Test error: `function GetValue: Integer; begin Exit "hello"; end;` (wrong type)
  - [x] Test error: `Exit 42;` at program scope (not in function)
  - [x] Test valid: `procedure DoSomething; begin Exit; end;` (no value for procedure)
  - [x] 5 test cases in `internal/semantic/statements_test.go`

#### Interpreter Support (2 tasks)

- [x] 9.204 Update `evalExitStatement` to return values:
  - [x] If `ReturnValue` is present, evaluate expression
  - [x] Return evaluated value wrapped in `ExitSignal` or similar control flow type
  - [x] If no value, return `ExitSignal` with nil (existing behavior)
  - [x] Ensure function caller receives the return value

- [x] 9.205 Add interpreter tests for Exit with values:
  - [x] Test function returning via Exit: `function Test: Boolean; begin Exit False; end;`
  - [x] Test early exit: `if condition then Exit value else Exit otherValue;`
  - [x] Test `Exit` in nested blocks (if/while/for)
  - [x] 4 test cases in `internal/interp/statements_test.go`

#### Testing & Documentation (1 task)

- [x] 9.206 Create test files and enable Death_Star.dws:
  - [x] Create `testdata/exit_statement/exit_with_value.dws`
  - [x] Verify `if (zsq < 0) then Exit False;` parses and executes
  - [x] Test boolean function: returns False on condition
  - [x] Add integration test in `pkg/dwscript/rosetta_examples_test.go`

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

## Phase 9 Summary

**Total Tasks**: ~229 tasks
**Estimated Effort**: ~27 weeks (~6.5 months)

### Priority Breakdown:

**HIGH PRIORITY** (~162 tasks, ~19 weeks):

- Subrange Types: 12 tasks
- Units/Modules System: 45 tasks (CRITICAL for multi-file projects)
- Function/Method Pointers: 25 tasks
- External Function Registration (FFI): 35 tasks
- Array Instantiation (`new TypeName[size]`): 12 tasks (CRITICAL for Rosetta Code examples)

**MEDIUM PRIORITY** (~67 tasks, ~8 weeks):

- Lambdas/Anonymous Methods: 30 tasks (depends on function pointers)
- Helpers: 20 tasks
- DateTime Functions: 24 tasks
- JSON Support: 18 tasks
- Improved Error Messages: 12 tasks

This comprehensive backlog brings go-dws from ~55% to ~85% feature parity with DWScript, making it production-ready for most use cases. The array instantiation feature is particularly critical as it unblocks numerous real-world DWScript examples (e.g., Rosetta Code algorithms) that rely on dynamic array creation.

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
