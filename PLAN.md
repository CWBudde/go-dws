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

**Status**: 83 completed tasks compacted into summary format | 83+ remaining completed tasks visible | 497 incomplete tasks

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
- **Helpers (Class/Record/Type)** (17 tasks): Complete helper type system extending existing types with additional methods without modification, supporting classes, records, and basic types (String, Integer, etc.) with method resolution priority and Self binding
- **DateTime Functions** (24 tasks): Comprehensive date/time functionality including current date/time access, formatting, parsing, arithmetic operations, and component extraction with full DWScript compatibility
- **Variant Type System** (20 tasks): Complete DWScript Variant type implementation for dynamic heterogeneous value storage, enabling array of const parameters, JSON support, and COM interop with boxing/unboxing, operators, comparisons, and built-in functions (VarType, VarIsNull, VarAsType, etc.)
- **JSON Support** (18 tasks): JSON parsing and serialization with Variant-based representation, supporting all JSON types (null, boolean, number, string, array, object) mapped to DWScript types where possible
- **Improved Error Messages and Stack Traces** (12 tasks): Enhanced error reporting with better messages, stack traces, source code snippets, and debugging information for improved developer experience
- **Contracts (Design by Contract)** (38 tasks): Complete DWScript contract system with preconditions (`require`), postconditions (`ensure`), `old` keyword for pre-execution values, and proper OOP inheritance semantics following Liskov substitution principle
- **Fixture Test Infrastructure** (8 tasks): Enhanced compatibility testing with DWScript reference suite, including program declarations, constructor parentheses optional syntax, zero-argument procedure calls, type meta-values, High/Low built-ins, and implicit type coercions
- **Advanced Property Modes** (4 tasks): Indexed properties (single/multi-index) with N-dimensional support, expression-based getters with Self binding, CLI test coverage (19 unit + 3 CLI tests)
- **Record Methods** (7 tasks): Parser, semantic analysis, and type system complete; interpreter method invocation pending

**Implementation Summary**: Phase 9 delivered comprehensive language feature expansion including advanced type system enhancements (aliases, subranges, variants), modern programming constructs (lambdas, contracts, helpers), rich standard library (DateTime, JSON, FFI), advanced property modes (indexed, expression-based), and improved developer experience (error messages, stack traces). The Variant type system serves as a foundational feature enabling many advanced capabilities. All completed features include comprehensive parser, semantic analyzer, interpreter, and CLI integration with dedicated test suites. Several major features remain in progress including full units/modules system completion and comprehensive fixture test suite expansion for 100% DWScript compatibility.

---

### Advanced FFI Features (Tasks 9.1-9.18)

**Current Status**: Basic FFI ✅ | Advanced features ❌ (0% complete)

**Implementation Notes**:
- Basic FFI complete: `pkg/dwscript/ffi.go` with `RegisterFunction()` API
- Marshaling: primitives, arrays, maps work (`internal/interp/marshal.go`)
- Error handling: Go errors → EHost exceptions, panic recovery complete
- Advanced features ALL missing: variadic, optional params, var params, methods, callbacks

**Files**:
- `pkg/dwscript/ffi.go` - Public FFI API
- `internal/interp/marshal.go` - Type marshaling (DWScript ↔ Go)
- `internal/interp/ffi.go` - Internal FFI implementation
- `pkg/dwscript/ffi_test.go` - Basic FFI tests

**DWScript FFI Examples**:

```delphi
// Optional parameters (9.1)
external function Greet(name: String; prefix: String = 'Hello'): String;

// By-reference parameters (9.2)
external procedure Swap(var a, b: Integer);  // Go: func Swap(a, b *int)

// Callbacks (9.4)
external procedure ForEach(arr: array of Integer; callback: TIntProc);
type TIntProc = procedure(value: Integer);
```

- [x] 9.1 Support optional parameters: ✅ COMPLETE
  - [x] 9.1a Parser: `param: Type = defaultValue` syntax (`internal/parser/functions.go:415`)
  - [x] 9.1b Semantic: Type-check defaults, validate ordering (`internal/semantic/analyze_functions.go:52-97`)
  - [x] 9.1c Interpreter: Fill missing args with defaults (`internal/interp/functions.go:719-755`)
  - [x] 9.1d FFI: DWScript fills defaults before calling Go (no Go changes needed)
  - [x] 9.1e Tests: Parser (10 tests), interpreter, FFI - all passing ✅

- [x] 9.2 Support by-reference parameters (var keyword): ✅ COMPLETE
  - [x] 9.2a Parser: `var param: Type` syntax (`internal/parser/functions.go:315-320`)
  - [x] 9.2b Semantic: Track var params, validate lvalues (`internal/semantic/analyze_functions.go:29,39-40,104`)
  - [x] 9.2c Interpreter: `ReferenceValue` type for by-ref passing (`internal/interp/value.go:246-297`)
  - [x] 9.2d FFI: Go pointers (*T) → var params, sync changes back (`pkg/dwscript/ffi.go`, `internal/interp/marshal.go`)
  - [x] 9.2e Tests: 12 tests (6 DWScript + 6 FFI) - all passing ✅, docs: `docs/ffi.md:371-508`

- [x] 9.3 Support registering Go methods: ✅ COMPLETE
  - [x] 9.3a Method values automatically bind receivers via closures (no detection needed)
  - [x] 9.3b `RegisterMethod(name, receiver, methodName)` API (`pkg/dwscript/ffi.go:186-212`)
  - [x] 9.3c Receiver binding automatic via Go's method value mechanism
  - [x] 9.3d Tests: 7 tests (Counter, Calculator) - all passing ✅, docs: `docs/ffi.md:521-817`

- [x] 9.4 Support callback functions (DWScript → Go → DWScript): ✅ COMPLETE
  - [x] 9.4a Marshal function pointers: `createGoFunctionWrapper()` with `reflect.MakeFunc` (`internal/interp/ffi_callback.go:170-241`)
  - [x] 9.4b Call DWScript from Go: `callDWScriptFunction()` with marshaling & exception handling (`internal/interp/ffi_callback.go:26-85`)
  - [x] 9.4c Re-entrancy: Automatic via existing environment/call stack infrastructure (recursion depth checked)
  - [x] 9.4d Error handling: DWScript exceptions → Go errors, panic recovery (`internal/interp/ffi_callback.go:68-101`)
  - [x] 9.4e Tests: 6 callback tests (forEach, map, nested, filter, multi-param) - all passing ✅
  - [x] 9.4f Infrastructure: `SetInterpreter()` interface, `MarshalToGo` signature updated, function type support

- [x] 9.5 Add comprehensive tests for advanced FFI features: ✅ COMPLETE
  - [x] 9.5a Integration tests combining features: ✅ COMPLETE
  - [x] 9.5b Error handling tests: ✅ COMPLETE
  - [x] 9.5c Performance benchmarks: ✅ COMPLETE
  - [x] 9.5d Documentation: ✅ COMPLETE
  - [x] 9.5e Example programs: ✅ COMPLETE

---

#### Full Contextual Type Inference (FUTURE ENHANCEMENT)

**Summary**: Task 9.19 currently has a placeholder implementation that reports an error when lambda parameters lack type annotations. Full contextual type inference would allow the compiler to infer parameter types from the context where the lambda is used.

**Current Status**: Lambda parameter type inference reports "not fully implemented" error. Return type inference from body is complete.

**Tasks for Full Implementation** (5 tasks):

- [x] 9.19 Add type context passing infrastructure to expression analyzer: **DONE**
  - [x] **ALREADY EXISTS**: `analyzeExpressionWithExpectedType()` in `internal/semantic/analyze_expressions.go`
  - [x] Current implementation: Handles ArrayLiteral→SetLiteral conversion (lines 118-150)

  **PHASE 1: Audit existing context usage** (1-2 hours) **DONE**
  - [x] 9.19.1 Document all current `analyzeExpressionWithExpectedType()` call sites:
    - [x] Search codebase: `grep -rn "analyzeExpressionWithExpectedType" internal/semantic/`
    - [x] Create list of which expression types already support context
    - [x] Current known: SetLiteral (line 117), ArrayLiteral (line 118), RecordLiteral, LambdaExpression
    - [x] Document in comment block at top of `analyzeExpressionWithExpectedType()`

  - [x] 9.19.2 Identify high-value candidates for context support:
    - [x] LambdaExpression (PRIMARY - blocks tasks 9.20-9.22) **IMPLEMENTED**
    - [x] RecordLiteral (already supported)
    - [ ] NilLiteral (context determines which pointer/class type) **DEFERRED**
    - [ ] Numeric literals (Integer vs Float inference) **DEFERRED**
    - [ ] Lower priority: CallExpression (for overload resolution) **DEFERRED**

  **PHASE 2: Extend switch statement in analyzeExpressionWithExpectedType** (2-3 hours) **DONE**
  - [x] 9.19.3 Add LambdaExpression case (CRITICAL for 9.20-9.22):
    - [x] Location: `internal/semantic/analyze_expressions.go` lines 151-162
    - [x] Pattern to follow: Lines 118-150 (ArrayLiteral→SetLiteral conversion)
    - [x] Implemented with `types.GetUnderlyingType()` to handle type aliases
    - [x] Created helper: `analyzeLambdaExpressionWithContext(lambda *ast.LambdaExpression, expectedFuncType *types.FunctionPointerType)`
    - [x] Location: `internal/semantic/analyze_lambdas.go` lines 157-324

  - [x] 9.19.4 Add RecordLiteral case: **ALREADY SUPPORTED** (line 119)

  - [ ] 9.19.5 Add NilLiteral case (OPTIONAL): **DEFERRED** to future task

  **PHASE 3: Thread context through call sites** (3-4 hours) **DONE**
  - [x] 9.19.6 Assignment statements: **ALREADY DONE** (analyze_statements.go:321)
    - [x] Uses `analyzeExpressionWithExpectedType(stmt.Value, sym.Type)`
    - [x] Benefit: `var f: TFunc := lambda(x) => ...` works ✓

  - [x] 9.19.7 Variable declarations: **ALREADY DONE** (analyze_statements.go:130)
    - [x] Uses `analyzeExpressionWithExpectedType(stmt.Value, varType)`
    - [x] Works with type annotations ✓

  - [x] 9.19.8 Function call arguments (analyze_builtins.go):
    - [x] Updated line 41 (method calls): uses `analyzeExpressionWithExpectedType(arg, funcType.Parameters[i])`
    - [x] Updated lines 2123, 2132 (user-defined function calls): uses `analyzeExpressionWithExpectedType(arg, expectedType)`
    - [x] Enables: `Apply(5, lambda(n) => n * 2)` where Apply expects function param ✓

  - [x] 9.19.9 Return statements (analyze_functions.go):
    - [x] Updated line 177: uses `analyzeExpressionWithExpectedType(stmt.ReturnValue, expectedType)`
    - [x] Works with function return types ✓

  - [x] 9.19.10 Array element assignment: **ALREADY DONE** (analyze_statements.go:372)
    - [x] Uses `analyzeExpressionWithExpectedType(stmt.Value, elementType)`

  **PHASE 4: Testing and validation** (1-2 hours) **DONE**
  - [x] 9.19.11 Add unit tests for context passing:
    - [x] Test: Context is nil → existing behavior unchanged (backward compat) ✓
    - [x] Test: Context provided → passed to specialized analyzer ✓
    - [x] Test: Nested contexts (array of lambdas, nested lambdas) ✓
    - [x] File: `internal/semantic/lambda_analyzer_test.go` (TestLambdaParameterTypeInference)
    - [x] Added 8 positive test cases covering all scenarios
    - [x] Added 6 negative test cases covering error conditions

  - [x] 9.19.12 Run existing test suite:
    - [x] `go test ./internal/semantic -v` - all lambda tests pass ✓
    - [x] Verified no regressions from adding context parameters ✓
    - [x] All tests pass with context=nil (backward compatible) ✓

  **Files modified**:
  - `internal/semantic/analyze_expressions.go` (+34 lines documentation, +9 lines lambda case)
  - `internal/semantic/analyze_lambdas.go` (+169 lines new function)
  - `internal/semantic/analyze_builtins.go` (+6 lines, 3 call sites updated)
  - `internal/semantic/analyze_functions.go` (+3 lines, return statement updated)
  - `internal/semantic/lambda_analyzer_test.go` (+175 lines, 14 new tests)

  **Actual time**: ~7 hours (within estimated 6-10 hours)

  **Summary**: Full contextual type inference infrastructure is now in place. Lambda parameters can be inferred from:
  - Variable declarations: `var f: TFunc := lambda(x) => x * 2`
  - Function call arguments: `Apply(5, lambda(n) => n * 2)`
  - Return statements: `Result := lambda(x) => x * factor`
  - Partial annotations: `lambda(a: Integer; b) => a + b`

  All tests pass. No regressions. Ready for tasks 9.20-9.22.

- [x] 9.20 Implement assignment context type inference: **COMPLETED IN TASK 9.19**
  **Prerequisite**: Task 9.19 must be complete (context passing infrastructure) ✓

  **Goal**: Enable `var f: TFunc := lambda(x) => x * 2` where x type is inferred from TFunc ✓

  **STATUS**: All functionality required by Task 9.20 was already implemented as part of Task 9.19.
  The `analyzeLambdaExpressionWithContext()` function provides complete assignment context inference.

  **PHASE 1: Create analyzeLambdaWithContext helper** ✓ **DONE IN 9.19**
  - [x] 9.20.1: Function created as `analyzeLambdaExpressionWithContext()` in analyze_lambdas.go:157-324
  - [x] 9.20.2: Parameter count validation implemented (line 176-179)
  - [x] 9.20.3: Type inference for untyped parameters implemented (lines 199-208)
  - [x] 9.20.4: Validation of explicit parameter types implemented (lines 210-226)
  - [x] 9.20.5: Lambda body analysis with inferred types implemented (lines 233-304)

  **PHASE 2: Verify return type compatibility** ✓ **DONE IN 9.19**
  - [x] 9.20.6: Return type compatibility checking implemented (lines 262-268, 285-291)
  - [x] 9.20.7: Void/procedure handling implemented (lines 272-274, 295-297)

  **PHASE 3: Integration tests** ✓ **DONE IN 9.19**
  - [x] 9.20.8: Test file created: `lambda_analyzer_test.go` with TestLambdaParameterTypeInference
  - [x] All required test cases implemented:
    - ✓ "infer single parameter type from variable declaration"
    - ✓ "infer multiple parameter types"
    - ✓ "partial type annotation - some explicit some inferred"
    - ✓ "infer parameter and return types"
    - ✓ "infer with procedure type"
    - ✓ "parameter count mismatch - too few"
    - ✓ "parameter count mismatch - too many"
    - ✓ "incompatible explicit parameter type"
    - ✓ "incompatible return type with explicit type"
    - ✓ "incompatible inferred return type"

  - [x] 9.20.9: DWScript examples work (verified with tests)

  **Files modified**: (same as Task 9.19)
  - `internal/semantic/analyze_lambdas.go` (analyzeLambdaExpressionWithContext function)
  - `internal/semantic/lambda_analyzer_test.go` (comprehensive tests)

  **Result**: 0 hours - already complete. All functionality works as specified.

- [x] 9.21 Implement function call context type inference: **COMPLETED IN TASK 9.19**
  **Prerequisite**: Tasks 9.19 and 9.20 must be complete ✓

  **Goal**: Enable `Apply(5, lambda(n) => n * 2)` where n type is inferred from Apply's signature ✓

  **STATUS**: All functionality required by Task 9.21 was already implemented as part of Task 9.19.
  Function call argument analysis was updated to use `analyzeExpressionWithExpectedType()`.

  **PHASE 1: Modify function call argument analysis** ✓ **DONE IN 9.19**
  - [x] 9.21.1: Located in `analyze_builtins.go` (handles both methods and user functions)
  - [x] 9.21.2: Function signature extraction already in place (funcType.Parameters)
  - [x] 9.21.3: Updated to use `analyzeExpressionWithExpectedType()`:
    - ✓ Method calls: line 41 in analyze_builtins.go
    - ✓ User-defined functions: lines 2123, 2132 in analyze_builtins.go
    - ✓ Lazy parameters: also updated to use context
  - [x] 9.21.4: Variadic parameters - ✅ **COMPLETE** (partial - array literal support remains)
    - [x] 9.21.4.1: Type System Foundation - Add `IsVariadic` and `VariadicType` fields to `FunctionType` ✅ **COMPLETE**
      - ✓ Added `IsVariadic bool` field to track if last parameter is variadic
      - ✓ Added `VariadicType Type` field to store element type of variadic parameter
      - ✓ Updated `String()` method to display "..." prefix for variadic parameters
      - ✓ Updated `Equals()` method to compare variadic status and types
      - ✓ Created `NewVariadicFunctionType()` and `NewVariadicFunctionTypeWithMetadata()` constructors
      - ✓ Added 7 comprehensive unit tests for variadic function type operations (all passing)
      - **Files**: `internal/types/function_type.go`, `internal/types/function_type_test.go`
      - **Actual time**: 1.5 hours
    - [x] 9.21.4.2: Parser Support for Variadic Parameters ✅ **COMPLETE**
      - ✓ Added special handling for "const" keyword as pseudo-type in `parseTypeExpression()`
      - ✓ Parser correctly handles `array of const` and `array of T` syntax in parameters
      - ✓ Added 4 comprehensive parser tests for variadic parameters (all passing)
      - **Files**: `internal/parser/types.go`, `internal/parser/functions_test.go`
      - **Actual time**: 1 hour
    - [x] 9.21.4.3: Semantic Analysis for Variadic Declarations ✅ **COMPLETE**
      - ✓ Updated `analyzeFunctionDecl()` to detect variadic parameters (last param = dynamic array)
      - ✓ Updated `analyzeMethodDecl()` with identical variadic detection logic
      - ✓ Correctly populates `FunctionType.IsVariadic` and `VariadicType` fields
      - ✓ Added "const" → Variant mapping in `resolveType()`
      - ✓ Created `variadic_test.go` with 3 comprehensive test functions (all passing)
      - **Files**: `internal/semantic/analyze_functions.go`, `internal/semantic/analyze_classes.go`, `internal/semantic/type_resolution.go`, `internal/semantic/variadic_test.go`
      - **Actual time**: 2 hours
    - [x] 9.21.4.4: Lambda Inference in Variadic Context ✅ **COMPLETE**
      - ✓ Modified function call analysis to detect variadic parameters
      - ✓ Extracts variadic element type for lambda inference (`funcType.VariadicType`)
      - ✓ Updated argument count validation (at least N for variadic, exactly N for non-variadic)
      - ✓ Uses `analyzeExpressionWithExpectedType()` with variadic element type
      - ✓ Updated both method calls and user function calls in analyze_builtins.go
      - ✓ Added tests to `lambda_analyzer_test.go` for variadic lambda inference
      - **Files**: `internal/semantic/analyze_builtins.go`, `internal/semantic/lambda_analyzer_test.go`
      - **Actual time**: 2 hours
    - [ ] 9.21.4.5: Array Literal Support for Variadic Calls - **IN PROGRESS**
      - Array literals `[val1, val2, ...]` are used to pass variadic arguments in DWScript
      - Currently semantic analyzer treats array literals as array types, not unpacked arguments
      - Need to detect array literal → variadic parameter pattern
      - Need to analyze each array element with the variadic element type
      - Need to support inline function types in array literals (e.g., `[function(x: Integer): Integer begin Result := x; end]`)
      - **Files**: `internal/semantic/analyze_expressions.go`, `internal/semantic/analyze_builtins.go`
      - **Estimated time**: 4-5 hours
    - **Total time**: 6.5 hours (vs 12-17 estimated) - core infrastructure complete, array literal support remains

  **PHASE 2: Handle overloaded functions** - **DEFERRED**
  - [ ] 9.21.5: Overload detection - **DEFERRED** to future task
  - [ ] 9.21.6: Overload resolution - **DEFERRED** to future task
  - [ ] 9.21.7: Ambiguous overloads - **DEFERRED** to future task

  Note: Overload resolution is a separate feature (Stage 8/9 tasks). Basic function call
  context inference works for single-signature functions.

  **PHASE 3: Complex inference scenarios** ✓ **WORKS**
  - [x] 9.21.8: Nested function calls - works automatically with context threading
  - [x] 9.21.9: Higher-order functions - works with existing infrastructure

  **PHASE 4: Testing** ✓ **DONE IN 9.19**
  - [x] 9.21.10: Test created in `lambda_analyzer_test.go`:
    - ✓ "infer parameter type from function call argument" test passes
    - ✓ Nested lambda test passes
  - [x] 9.21.11: DWScript examples verified in tests
  - [x] 9.21.12: Error cases tested:
    - ✓ "parameter count mismatch" tests
    - ✓ "incompatible explicit parameter type" test
    - ✓ "incompatible return type" tests

  **Files modified**: (same as Task 9.19)
  - `internal/semantic/analyze_builtins.go` (3 call sites updated)
  - `internal/semantic/lambda_analyzer_test.go` (includes function call tests)

  **Result**: 0 hours - already complete. Function call context inference works as specified.

  **Files to modify**:
  - `internal/semantic/analyze_expressions.go` (analyzeCallExpression)
  - `internal/semantic/overload_resolution.go` (new file for overload logic)
  - `internal/semantic/lambda_call_inference_test.go` (new file)
  - `testdata/lambdas/call_inference.dws` (new test file)

  **Estimated time**: 7-10 hours total

- [x] 9.22 Implement return statement context type inference: **COMPLETED IN TASK 9.19**
  **Prerequisite**: Tasks 9.19 and 9.20 must be complete ✓

  **Goal**: Enable `function MakeDoubler(): function(Integer): Integer; begin Result := lambda(x) => x * 2; end;` ✓

  **STATUS**: ✅ Complete - implemented as part of Task 9.19
  - Done: Return statement uses `analyzeExpressionWithExpectedType()`, handles Result variable, multiple return paths, procedures
  - Files: `analyze_functions.go`, `lambda_analyzer_test.go`
- [x] 9.23 Comprehensive lambda type inference tests: ✅ **COMPLETE**
  - Done: 54 unit tests, 2 integration files (8 scenarios + 10 error cases), edge case tests
  - Deferred: Deeply nested/recursive inference, docs update, demo file
  - Files: `lambda_analyzer_test.go` (+9), 2 test files (314 lines)

- [x] 9.24 Dynamic array literal syntax: ✅ **COMPLETE**
  - Done: Parser heuristic `[1,2,3]`=array, `[Red,Blue]`=set, works for 95% of cases
  - Deferred: Semantic context override for edge cases
  - Files: `internal/parser/arrays.go`

---

### Fixture Test Failures (Algorithms Category)

#### Output Mismatches - Empty Output (8 tasks)

- [x] 9.25 Fix death_star.pas empty output: ✅ **COMPLETE**
  - [x] Root cause 1: Missing ClampInt() and Clamp() built-in functions
  - [x] Root cause 2: Record literals not supported in const declarations
  - [x] Implemented ClampInt(value, min, max: Integer): Integer
  - [x] Implemented Clamp(value, min, max: Float): Float
  - [x] Added support for record literals in const evaluation
  - [x] Files: `internal/interp/builtins_math.go`, `internal/interp/functions.go`, `internal/semantic/analyze_builtins.go`, `internal/semantic/analyze_types.go`, `internal/interp/math_test.go`

- [x] 9.26 Fix horizontal_sundial.pas empty output: ✅ **COMPLETE**
  - [x] Root cause: Missing built-in implicit Integer→Float coercion
  - [x] Script calls `PrintSundial(-4.95, -150.5, -150)` where third arg is Integer but parameter is Float
  - [x] Added automatic Integer→Float widening in `tryImplicitConversion()` when no registered conversion exists
  - [x] This is standard Pascal/Delphi behavior - integers automatically promote to floats
  - [x] Script now produces expected 21 lines of sundial calculations
  - [x] Files: `internal/interp/operators_eval.go`, `internal/interp/conversion_test.go`

- [ ] 9.27 Fix koch.pas empty output:
  - [ ] Expected: Koch curve line segments (180 lines)
  - [ ] Actual: (empty)
  - [ ] Root cause: Recursive algorithm not executing
  - [ ] Priority: LOW
  - [ ] File: testdata/fixtures/Algorithms/koch.pas

- [ ] 9.28 Fix pascal_triangle.pas empty output:
  - [ ] Expected: Pascal's triangle (9 lines)
  - [ ] Actual: (empty)
  - [ ] Root cause: Algorithm not producing output
  - [ ] Priority: MEDIUM
  - [ ] File: testdata/fixtures/Algorithms/pascal_triangle.pas

- [ ] 9.29 Fix sierpinski.pas empty output:
  - [ ] Expected: Sierpinski triangle (16 lines)
  - [ ] Actual: (empty)
  - [ ] Root cause: Algorithm not producing output
  - [ ] Priority: LOW
  - [ ] File: testdata/fixtures/Algorithms/sierpinski.pas

- [ ] 9.30 Fix sierpinski_carpet.pas empty output:
  - [ ] Expected: Sierpinski carpet (27 lines)
  - [ ] Actual: (empty)
  - [ ] Root cause: Algorithm not producing output
  - [ ] Priority: LOW
  - [ ] File: testdata/fixtures/Algorithms/sierpinski_carpet.pas

- [ ] 9.31 Fix quicksort_dyn.pas incomplete output:
  - [ ] Expected: `Swaps: >=100` and sorted array 0-99
  - [ ] Actual: `Swaps: <100` and empty data
  - [ ] Root cause: QuickSort algorithm bug or incomplete execution
  - [ ] Priority: HIGH - critical algorithm
  - [ ] File: testdata/fixtures/Algorithms/quicksort_dyn.pas

#### Output Mismatches - Incorrect Calculations (2 tasks)

- [ ] 9.32 Fix cholesky_decomposition.pas matrix calc:
  - [ ] Error: Incorrect matrix values (all zeros in many positions)
  - [ ] Expected: Specific float values (e.g., 3.00000, 6.56591)
  - [ ] Actual: Mostly zeros
  - [ ] Root cause: Matrix decomposition algorithm bug
  - [ ] Priority: MEDIUM
  - [ ] File: testdata/fixtures/Algorithms/cholesky_decomposition.pas

- [ ] 9.33 Fix gnome_sort.pas sorting logic:
  - [ ] Expected: Sorted array `{0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15}`
  - [ ] Actual: Unsorted array (same as input)
  - [ ] Root cause: Gnome sort algorithm not modifying array
  - [ ] Priority: HIGH - sorting is fundamental
  - [ ] File: testdata/fixtures/Algorithms/gnome_sort.pas

#### Output Mismatches - Partial Output (3 tasks)

- [ ] 9.34 Fix draw_sphere.pas incomplete output:
  - [ ] Expected: 39x39 PGM image (full sphere)
  - [ ] Actual: Only first line of data
  - [ ] Root cause: Loop terminating early or output truncated
  - [ ] Priority: MEDIUM
  - [ ] File: testdata/fixtures/Algorithms/draw_sphere.pas

- [ ] 9.35 Fix extract_ranges.pas missing range:
  - [ ] Expected: `0-2,4,6-8,11,12,14-25,27-33,35-39`
  - [ ] Actual: `0-2,4,6-8,11,12,14-25,27-33,35`
  - [ ] Root cause: Missing final range `36-39`
  - [ ] Priority: LOW - minor logic bug
  - [ ] File: testdata/fixtures/Algorithms/extract_ranges.pas

- [ ] 9.36 Fix roman_numerals.pas nil output:
  - [ ] Expected: `CDLV`, `MMMCDLVI`, `MMCDLXXXVIII`
  - [ ] Actual: `nil`, `nil`, `nil`
  - [ ] Root cause: Function returning nil instead of string
  - [ ] Priority: MEDIUM
  - [ ] File: testdata/fixtures/Algorithms/roman_numerals.pas

#### Runtime Errors - Record Method Access (1 task) - 0% COMPLETE

- [ ] 9.37 Fix lerp.pas record method access:
  - [ ] Error: `ERROR: field 'Print' not found in record 'TPointF'`
  - [ ] Root cause: `evalMemberAccess` in objects.go doesn't check record methods
  - [ ] Current behavior (lines 210-222): Only checks fields, then helpers, then errors
  - [ ] Required behavior: Check fields → methods → helpers (like classes do at lines 252-280)
  - [ ] File: testdata/fixtures/Algorithms/lerp.pas
  - [ ] Priority: HIGH - blocks lerp.pas and similar patterns
  - [ ] Affects: Method calls on records (both variables and temporaries)
  - [ ] Example failure: `LinearInterpolation(p1, p2, 0).Print` - method not found
  - [ ] Example failure: `record.Print` where Print is a parameterless procedure
  - [ ] Implementation details:
    - [ ] Modify `evalMemberAccess` in `/mnt/projekte/Code/go-dws/internal/interp/objects.go` (lines 210-222)
    - [ ] After checking `recordVal.Fields` (line 212), add check for `recordVal.Methods`
    - [ ] If method found and has no parameters: auto-invoke it (mirror class behavior from Task 9.173)
    - [ ] If method found and has parameters: return as function pointer
    - [ ] Only fall back to helpers (line 215) if no field or method found
    - [ ] Add test case: parameterless method access without parentheses (`record.Print`)
    - [ ] Add test case: method calls on temporary record values (`GetRecord().Method()`)
  - [ ] Related tasks:
    - [ ] Task 9.173: Auto-invoke parameterless class methods (same pattern needed for records)
    - [ ] Stage 8 record method tasks: Already implemented record methods, just missing runtime access
  - [ ] Test validation:
    - [ ] lerp.pas should execute successfully after fix
    - [ ] All existing record method tests should still pass
    - [ ] New test cases should cover gaps (parameterless, temporaries)

---

### Function/Method Overloading Support (Tasks 9.243-9.277) - 6% COMPLETE

**Goal**: Implement complete function and method overloading support
**Status**: 2/35 tasks complete
**Priority**: MEDIUM - Required for 76+ fixture tests in OverloadsPass/
**Reference**: DWScript dwsCompiler.pas (ReadFuncOverloaded, ResolveOverload)
**Test Files**: testdata/fixtures/OverloadsPass/ (36 tests), testdata/fixtures/OverloadsFail/ (11 tests)

#### Phase 1: Parser Support (Tasks 9.243-9.249) - 57% COMPLETE

- [x] 9.38 Add IsOverload field to FunctionDecl AST node:
  - [x] Add `IsOverload bool` field to `FunctionDecl` struct in `internal/ast/functions.go`
  - [x] Update `FunctionDecl.String()` to output `; overload` directive when true
  - [x] Foundation for storing overload metadata in AST

- [x] 9.39 Parse overload keyword in function declarations:
  - [x] Add handling in `parseFunctionDeclaration()` in `internal/parser/functions.go`
  - [x] Parse `OVERLOAD` token after function signature (after line 145)
  - [x] Set `fn.IsOverload = true` and expect semicolon
  - [x] Fixes parsing error in lerp.pas (task 9.214)

- [ ] 9.40 Parse overload keyword in procedure declarations:
  - [ ] Add same handling in `parseProcedureDeclaration()`
  - [ ] Ensure procedures can be overloaded like functions
  - [ ] Test with overloaded procedure examples

- [ ] 9.41 Parse overload keyword in method declarations:
  - [ ] Add handling in class method parsing
  - [ ] Support both instance and class methods
  - [ ] Test with method overload examples from OverloadsPass/

- [ ] 9.42 Parse overload keyword in constructor declarations:
  - [ ] Add handling in constructor parsing
  - [ ] Support multiple constructor signatures
  - [ ] Test with testdata/fixtures/OverloadsPass/overload_constructor.pas

- [ ] 9.43 Parse overload keyword in record method declarations:
  - [ ] Add handling in record method parsing
  - [ ] Support both instance and static record methods
  - [ ] Test with record method overload examples

- [ ] 9.44 Add comprehensive parser tests for overload keyword:
  - [ ] Test function with overload directive
  - [ ] Test procedure with overload directive
  - [ ] Test method with overload directive
  - [ ] Test constructor with overload directive
  - [ ] Test multiple directives: `virtual; overload;`
  - [ ] Test forward declarations with overload

#### Phase 2: Symbol Table Extensions (Tasks 9.250-9.255) - 100% COMPLETE ✅

- [x] 9.45 Design overload set data structure:
  - [x] Create `OverloadSet` type to store multiple function signatures
  - [x] Store list of `*types.FunctionType` with parameter info
  - [x] Track which overload is "primary" (first declared)
  - [x] Reference: DWScript TFuncSymbol with overload list
  - Implementation: Used `Overloads []*Symbol` field in Symbol struct

- [x] 9.46 Extend Symbol to support multiple function definitions:
  - [x] Add `Overloads []*Symbol` field to Symbol struct
  - [x] Add `IsOverloadSet` flag to distinguish overloaded symbols
  - [x] Maintain backward compatibility for non-overloaded functions
  - File: internal/semantic/symbol_table.go:16-17

- [x] 9.47 Add DefineOverload() method to SymbolTable:
  - [x] Create new method: `DefineOverload(name string, funcType *types.FunctionType, overload bool) error`
  - [x] Check if name exists: if not, create new symbol
  - [x] If exists: add to overload set if signatures differ
  - [x] Validate overload directive consistency
  - File: internal/semantic/symbol_table.go:89-182

- [x] 9.48 Add GetOverloadSet() method to retrieve all overloads:
  - [x] Create method: `GetOverloadSet(name string) []*Symbol`
  - [x] Return all function variants for a given name
  - [x] Return single-element slice for non-overloaded functions
  - File: internal/semantic/symbol_table.go:184-204

- [x] 9.49 Update DefineFunction() to handle overload conflicts:
  - [x] Detect when function name already exists
  - [x] If neither has `overload` directive: error (duplicate function)
  - [x] If only one has `overload`: warning or error based on DWScript rules
  - [x] Route to DefineOverload() when appropriate
  - Note: Deferred to future integration with parser

- [x] 9.50 Add unit tests for overload set storage and retrieval:
  - [x] Test storing multiple overloads (5 tests)
  - [x] Test retrieving overload sets (4 tests)
  - [x] Test conflict detection (6 tests)
  - [x] Test nested scopes with overloads (4 tests)
  - File: internal/semantic/overload_test.go (19 comprehensive tests, all passing)

#### Phase 3: Signature Matching (Tasks 9.256-9.262) - 0% COMPLETE

- [ ] 9.51 Implement function signature comparison:
  - [ ] Create `SignaturesEqual(sig1, sig2 *types.FunctionType) bool`
  - [ ] Compare parameter count
  - [ ] Compare parameter types (exact match)
  - [ ] Compare parameter modifiers (var/const/out)
  - [ ] Ignore return type (overloads differ by params only)

- [ ] 9.52 Implement signature distance calculation:
  - [ ] Create `SignatureDistance(callArgs []Value, funcSig *types.FunctionType) int`
  - [ ] Return 0 for exact match
  - [ ] Return positive value for compatible match (with conversions)
  - [ ] Return -1 for incompatible match
  - [ ] Consider type hierarchy (Integer → Float, etc.)

- [ ] 9.53 Implement best-fit overload selection algorithm:
  - [ ] Create `ResolveOverload(overloads []*Symbol, callArgs []Value) (*Symbol, error)`
  - [ ] Find all compatible overloads
  - [ ] Select overload with smallest distance (most specific)
  - [ ] Error if no compatible overload found
  - [ ] Error if multiple overloads have same distance (ambiguous)
  - [ ] Reference: DWScript `TFuncSymbol.ResolveOverload()`

- [ ] 9.54 Handle default parameters in overload resolution:
  - [ ] Consider default params when matching signatures
  - [ ] Allow calling overload with fewer args if defaults exist
  - [ ] Prefer overload with exact arg count over one with defaults

- [ ] 9.55 Handle parameter modifiers in matching:
  - [ ] Consider `var`, `const`, `out` modifiers
  - [ ] `var` parameters require exact type match
  - [ ] `const` parameters allow compatible types
  - [ ] Test with examples from OverloadsPass/

- [ ] 9.56 Add comprehensive tests for ResolveOverload():
  - [ ] Test exact match selection
  - [ ] Test compatible match with conversions
  - [ ] Test ambiguous call detection
  - [ ] Test no-match error cases
  - [ ] Test with default parameters

- [ ] 9.57 Test edge cases in overload resolution:
  - [ ] nil literal compatibility with multiple types
  - [ ] Variant type compatibility
  - [ ] Array types with different dimensions
  - [ ] Class inheritance and interface compatibility

#### Phase 4: Semantic Validation (Tasks 9.263-9.269) - 0% COMPLETE

- [ ] 9.58 Validate overload directive consistency:
  - [ ] Implement: If one overload has `overload`, all must have it
  - [ ] Exception: Last/only implementation can omit directive
  - [ ] Generate error: "Overload directive required for function X"
  - [ ] Test with testdata/fixtures/OverloadsFail/overload_missing.pas

- [ ] 9.59 Detect duplicate signatures in overload set:
  - [ ] Compare all pairs of overloads for same function name
  - [ ] Error if two overloads have identical signatures
  - [ ] Error message: "Duplicate overload for function X with signature Y"
  - [ ] Test with testdata/fixtures/OverloadsFail/overload_simple.pas

- [ ] 9.60 Validate overload + forward declaration consistency:
  - [ ] Forward declaration and implementation must both have/omit `overload`
  - [ ] Check signature match between forward and implementation
  - [ ] Test with testdata/fixtures/OverloadsPass/forwards.pas
  - [ ] Test failure cases with OverloadsFail/forwards.pas

- [ ] 9.61 Check overload + virtual/override/abstract interactions:
  - [ ] Virtual methods can be overloaded
  - [ ] Override must match base method signature (separate from overloading)
  - [ ] Abstract overloads allowed in abstract classes
  - [ ] Test with testdata/fixtures/OverloadsPass/overload_virtual.pas

- [ ] 9.62 Implement comprehensive conflict detection:
  - [ ] Port DWScript `FuncHasConflictingOverload()` logic
  - [ ] Check for ambiguous overloads at definition time
  - [ ] Warn about potentially confusing overload sets

- [ ] 9.63 Add detailed error messages for overload violations:
  - [ ] List all existing overloads when showing conflict
  - [ ] Show signature comparison in error messages
  - [ ] Suggest fixes (add overload directive, change signature)

- [ ] 9.64 Run OverloadsFail/ fixture tests:
  - [ ] Validate all 11 expected failure cases pass
  - [ ] Ensure error messages match expected patterns
  - [ ] Document any DWScript incompatibilities

#### Phase 5: Runtime Dispatch (Tasks 9.270-9.274) - 0% COMPLETE

- [ ] 9.65 Update function call evaluation to resolve overloads:
  - [ ] In `evalCallExpression()`, check if function is overloaded
  - [ ] Get overload set from symbol table
  - [ ] Call `ResolveOverload()` with actual arguments
  - [ ] Execute selected overload
  - [ ] Error if resolution fails

- [ ] 9.66 Store multiple function implementations in environment:
  - [ ] Extend environment to store overload sets
  - [ ] Each overload gets unique internal name (e.g., `Func#0`, `Func#1`)
  - [ ] Map display name to overload set
  - [ ] Maintain for nested scopes

- [ ] 9.67 Implement overload dispatch for method calls:
  - [ ] Handle instance method overloads
  - [ ] Handle class method overloads
  - [ ] Consider inheritance (can call base class overload)
  - [ ] Test with testdata/fixtures/OverloadsPass/meth_overload_*.pas

- [ ] 9.68 Implement overload dispatch for constructor calls:
  - [ ] Handle `Create` constructor with multiple signatures
  - [ ] Select appropriate constructor based on arguments
  - [ ] Test with testdata/fixtures/OverloadsPass/overload_constructor.pas

- [ ] 9.69 Add runtime tests for overload execution:
  - [ ] Test calling each overload variant
  - [ ] Test overload selection at runtime
  - [ ] Test error messages for ambiguous/missing overloads
  - [ ] Benchmark overload resolution performance

#### Phase 6: Integration & Testing (Tasks 9.275-9.277) - 0% COMPLETE

- [ ] 9.70 Run OverloadsPass/ fixture suite:
  - [ ] Execute all 36 passing tests
  - [ ] Verify output matches expected results
  - [ ] Document any failures or incompatibilities
  - [ ] Measure test coverage for overload code

- [ ] 9.71 Fix and verify lerp.pas execution:
  - [ ] File: testdata/fixtures/Algorithms/lerp.pas
  - [ ] Verify parsing succeeds (already fixed in 9.244)
  - [ ] Verify semantic analysis passes
  - [ ] Verify execution produces correct output
  - [ ] Mark task 9.214 as fully complete

- [ ] 9.72 Update documentation with overloading examples:
  - [ ] Add overloading section to language guide
  - [ ] Document overload resolution rules
  - [ ] Provide best practices and examples
  - [ ] Update CLAUDE.md with overloading info

---

### Improved Error Messages and Stack Traces ✅ 90% COMPLETE (MEDIUM PRIORITY)

- [x] 9.73 Add source position to all AST nodes: PARTIAL
  - [x] Audit nodes for missing `Pos` fields ✅ COMPLETE (all nodes have Pos via Token field)
  - [ ] Add `EndPos` for better range reporting (DEFERRED - requires extensive parser refactoring)
  - [ ] Use in error messages (partially done - current error messages use start position)

---

- [ ] 9.74 Handle method contract inheritance at runtime
  - For preconditions: Only evaluate base class conditions (weakening)
  - For postconditions: Evaluate conditions from all ancestor classes (strengthening)
  - Walk up method override chain to collect conditions
  - **Reason**: Stage 7 (OOP/Classes) not yet complete; requires class hierarchy support

---

### Comprehensive Testing

- [ ] 9.75 Run DWScript example scripts from documentation
- [ ] 9.76 Compare outputs with original DWScript
- [ ] 9.77 Fix any discrepancies
- [ ] 9.78 Create stress tests for complex features
- [ ] 9.79 Achieve >85% overall code coverage

---

## Phase 9 Summary

**Total Tasks**: ~290 tasks (updated from ~255)
**Estimated Effort**: ~34 weeks (~8.5 months)

### Priority Breakdown

**HIGH PRIORITY** (~208 tasks, ~24 weeks):

- Subrange Types: 12 tasks
- Units/Modules System: 45 tasks (CRITICAL for multi-file projects)
- Function/Method Pointers: 25 tasks
- External Function Registration (FFI): 35 tasks
- Array Instantiation (`new TypeName[size]`): 12 tasks (CRITICAL for Rosetta Code examples)
- For Loop Step Keyword: 11 tasks (REQUIRED for Lucas-Lehmer test and other Rosetta Code examples)
- **Contracts (Design by Contract): 38 tasks** (REQUIRED for Rosetta Code examples like Dot_product)

**MEDIUM-HIGH PRIORITY** (~15 tasks, ~2 weeks):

- Lazy Parameter Passing: 15 tasks (REQUIRED for Jensen's Device and Rosetta Code examples)

**MEDIUM PRIORITY** (~67 tasks, ~8 weeks):

- Lambdas/Anonymous Methods: 30 tasks (depends on function pointers)
- Helpers: 20 tasks
- DateTime Functions: 24 tasks
- JSON Support: 18 tasks
- Improved Error Messages: 12 tasks

This comprehensive backlog brings go-dws from ~55% to ~85% feature parity with DWScript, making it production-ready for most use cases. The for loop step keyword, array instantiation, and lazy parameter features are particularly critical as they unblock numerous real-world DWScript examples (e.g., Rosetta Code algorithms like Lucas-Lehmer test and Jensen's Device) that rely on custom loop increments, dynamic array creation, and deferred expression evaluation.

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

- [ ] **10.14 Create visitor pattern for AST traversal**
  - [ ] Create `pkg/ast/visitor.go`
  - [ ] Define `Visitor` interface:
    - [ ] `Visit(node Node) (w Visitor)` - Standard Go AST walker pattern
  - [ ] Implement `Walk(v Visitor, node Node)` function
  - [ ] Handle all node types in Walk
  - [ ] Add documentation with examples
  - [ ] Add `Inspect(node Node, f func(Node) bool)` helper

- [ ] **10.15 Add symbol table access for semantic information**
  - [ ] Create `pkg/dwscript/symbols.go`
  - [ ] Define `Symbol` struct:
    - [ ] `Name string`
    - [ ] `Kind string` - "variable", "function", "class", "parameter", etc.
    - [ ] `Type string` - Type name
    - [ ] `Position Position` - Definition location
    - [ ] `Scope string` - "local", "global", "class"
  - [ ] Add method: `func (p *Program) Symbols() []Symbol`
  - [ ] Extract symbols from semantic analyzer's symbol table
  - [ ] Include all declarations with their positions

- [ ] **10.16 Add type information access**
  - [ ] Add method: `func (p *Program) TypeAt(pos Position) (string, bool)`
  - [ ] Return type of expression at given position
  - [ ] Use semantic analyzer's type information
  - [ ] Return ("", false) if position doesn't map to typed expression
  - [ ] Add method: `func (p *Program) DefinitionAt(pos Position) (*Position, bool)`
  - [ ] Return definition location for identifier at position

- [ ] **10.17 Update error formatting for better IDE integration**
  - [ ] Ensure error messages are clear and concise
  - [ ] Remove redundant position info from message text
  - [ ] Use consistent error message format
  - [ ] Add suggested fixes where applicable (future enhancement)
  - [ ] Document error message format

- [ ] **10.18 Write unit tests for structured errors**
  - [ ] Create `pkg/dwscript/error_test.go`
  - [ ] Test Error struct creation and formatting
  - [ ] Test CompileError with multiple structured errors
  - [ ] Test that positions are accurate
  - [ ] Test severity levels (error vs warning)
  - [ ] Test error codes if implemented

- [ ] **10.19 Write unit tests for AST position metadata**
  - [ ] Create `pkg/ast/position_test.go`
  - [ ] Test position on simple statements
  - [ ] Test position on nested expressions
  - [ ] Test position on multi-line constructs
  - [ ] Test Pos() and End() methods on all node types
  - [ ] Verify 1-based line numbering
  - [ ] Test with Unicode/multi-byte characters

- [ ] **10.20 Write unit tests for AST export**
  - [ ] Create `pkg/ast/ast_test.go`
  - [ ] Test that Program.AST() returns valid AST
  - [ ] Test AST traversal with visitor pattern
  - [ ] Test AST structure for various programs
  - [ ] Test that AST nodes have correct types
  - [ ] Test accessing child nodes

- [ ] **10.21 Write unit tests for Parse() mode**
  - [ ] Test parsing valid code
  - [ ] Test parsing code with syntax errors
  - [ ] Verify partial AST is returned on error
  - [ ] Test that structured errors are returned
  - [ ] Compare Parse() vs Compile() behavior
  - [ ] Measure performance difference

- [ ] **10.22 Write integration tests**
  - [ ] Create `pkg/dwscript/integration_test.go`
  - [ ] Test complete workflow: Parse → AST → Symbols
  - [ ] Test error recovery scenarios
  - [ ] Test position mapping accuracy
  - [ ] Use real DWScript code samples from testdata/
  - [ ] Verify no regressions in existing functionality

- [ ] **10.23 Update package documentation**
  - [ ] Update `pkg/dwscript/doc.go` with new API
  - [ ] Add examples for accessing AST
  - [ ] Add examples for structured errors
  - [ ] Document position coordinate system (1-based)
  - [ ] Add migration guide if breaking changes
  - [ ] Document LSP use case

- [ ] **10.24 Update README with new capabilities**
  - [ ] Add section on LSP/IDE integration
  - [ ] Show example of using structured errors
  - [ ] Show example of AST traversal
  - [ ] Show example of symbol extraction
  - [ ] Link to pkg.go.dev documentation
  - [ ] Note minimum Go version if changed

- [ ] **10.25 Verify backwards compatibility or version bump**
  - [ ] Run all existing tests
  - [ ] Check if API changes are backwards compatible
  - [ ] If breaking: plan major version bump (v2.0.0)
  - [ ] If compatible: plan minor version bump (v1.x.0)
  - [ ] Update go.mod version if needed
  - [ ] Document breaking changes in CHANGELOG

- [ ] **10.26 Performance testing**
  - [ ] Benchmark parsing with position tracking
  - [ ] Ensure position metadata doesn't significantly slow parsing
  - [ ] Target: <10% performance impact
  - [ ] Benchmark Parse() vs Compile()
  - [ ] Profile memory usage with AST export
  - [ ] Optimize if needed

- [ ] **10.27 Tag release and publish**
  - [ ] Create git tag for new version
  - [ ] Push tag to trigger pkg.go.dev update
  - [ ] Write release notes
  - [ ] Announce new LSP-friendly features
  - [ ] Update go-dws-lsp dependency to new version

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
- [ ] 12.16 Implement language server protocol (LSP):
  - [ ] Syntax highlighting
  - [ ] Autocomplete
  - [ ] Go-to-definition
  - [ ] Error diagnostics in IDE
- [ ] 12.17 Implement JavaScript code generation backend:
  - [ ] AST → JavaScript transpiler
  - [ ] Support browser execution
  - [ ] Create npm package

### Alternative Execution Modes

- [ ] 12.18 Add JIT compilation (if feasible in Go) - **MEDIUM-LOW PRIORITY**

  **Feasibility**: Challenging but achievable. JIT in Go has significant limitations due to lack of runtime code generation. Bytecode VM provides good ROI (2-3x speedup), while LLVM JIT is very complex (5-10x speedup but high maintenance burden).

  **Recommended Approach**: Implement bytecode VM (Phase 1), defer LLVM JIT (Phase 2).

  #### Phase 1: Bytecode VM Foundation (RECOMMENDED - 12-16 weeks)

  - [x] 12.18.1 Research and design bytecode instruction set (1-2 weeks, COMPLEX) ✅
    - Study DWScript's existing bytecode format (DWScript uses direct JIT to x86, no bytecode)
    - Design instruction set: stack-based (~116 opcodes) vs register-based (~150 opcodes)
    - Define bytecode format: 32-bit instructions (Go-optimized)
    - Document instruction set with examples
    - Create `internal/bytecode/instruction.go` with opcode constants
    - **Decision**: Stack-based VM with 116 opcodes, 32-bit instruction format
    - **Expected Impact**: 2-3x speedup over tree-walking interpreter
    - **Documentation**: See [docs/architecture/bytecode-vm-design.md](docs/architecture/bytecode-vm-design.md) and [docs/architecture/bytecode-vm-quick-reference.md](docs/architecture/bytecode-vm-quick-reference.md)

  - [x] 12.18.2 Implement bytecode data structures (3-5 days, MODERATE) ✅
    - Created `internal/bytecode/bytecode.go` with `Chunk` type (bytecode + constants pool)
    - Implemented constant pool for literals (integers, floats, strings) with deduplication
    - Added line number mapping with run-length encoding for error reporting
    - Implemented bytecode disassembler in `disasm.go` for debugging
    - Added comprehensive unit tests (79.7% coverage)
    - **Files**: bytecode.go (464 lines), bytecode_test.go (390 lines), disasm.go (357 lines), disasm_test.go (325 lines)

  - [x] 12.18.3 Build AST-to-bytecode compiler (2-3 weeks, COMPLEX)
    - [x] Create `internal/bytecode/compiler.go`
    - [x] Implement visitor pattern for AST traversal (baseline literal/control-flow coverage)
    - [x] Compile expressions: literals, binary ops, unary ops, variables, function calls *(OpCallIndirect baseline)*
    - [x] Compile statements: assignment, if/else, while/repeat loops, return *(numeric for/case handled in later phase)*
    - [x] Handle scoping and variable resolution for locals
    - [x] Optimize constant folding during compilation *(arithmetic/comparison literals folded to single load)*
    - [x] Add comprehensive unit tests comparing AST eval vs bytecode execution *(mini VM harness vs interpreter parity)*

  - [x] 12.18.4 Implement bytecode VM core (2-3 weeks, COMPLEX)
    - [x] Create `internal/bytecode/vm.go` with VM struct
    - [x] Implement instruction dispatch loop (switch statement on opcode)
    - [x] Implement operand stack (for stack-based VM)
    - [x] Add call stack for function returns *(function invocation opcodes stubbed for future work)*
    - [x] Implement environment/closure handling *(globals + upvalue capture/closure support)*
    - [x] Add error handling and stack traces *(structured RuntimeError with stack trace reporting)*
    - [x] Benchmark against tree-walking interpreter *(see BenchmarkVMVsInterpreter_CountLoop in vm_bench_test.go)*

  - [x] 12.18.5 Implement arithmetic and logic instructions (1 week, MODERATE)
    - ADD, SUB, MUL, DIV, MOD instructions
    - NEGATE, NOT instructions
    - EQ, NE, LT, LE, GT, GE comparisons
    - AND, OR, XOR bitwise operations
    - Type coercion (int ↔ float)

  - [ ] 12.18.6 Implement variable and memory instructions (1 week, MODERATE)
    - LOAD_CONST, LOAD_VAR, STORE_VAR instructions
    - LOAD_GLOBAL, STORE_GLOBAL for global variables
    - LOAD_UPVALUE, STORE_UPVALUE for closures
    - GET_PROPERTY, SET_PROPERTY for object members

  - [ ] 12.18.7 Implement control flow instructions (1 week, MODERATE)
    - JUMP, JUMP_IF_FALSE, JUMP_IF_TRUE
    - LOOP (jump backward for while/for loops)
    - Patch jump addresses during compilation

  - [ ] 12.18.8 Implement function call instructions (1-2 weeks, COMPLEX)
    - CALL instruction with argument count
    - RETURN instruction
    - Handle recursion and call stack depth
    - Implement closures and upvalues
    - Support method calls and `Self` context

  - [ ] 12.18.9 Implement array and object instructions (1 week, MODERATE)
    - GET_INDEX, SET_INDEX for array access
    - NEW_ARRAY, ARRAY_LENGTH
    - NEW_OBJECT for class instantiation
    - INVOKE_METHOD for method dispatch

  - [ ] 12.18.10 Add exception handling instructions (1 week, MODERATE)
    - TRY, CATCH, FINALLY, THROW instructions
    - Exception stack unwinding
    - Preserve stack traces across bytecode execution

  - [ ] 12.18.11 Optimize bytecode generation (1-2 weeks, MODERATE)
    - Peephole optimization (combine adjacent instructions)
    - Dead code elimination
    - Constant propagation
    - Inline small functions (< 10 instructions)

  - [ ] 12.18.12 Integrate bytecode VM into interpreter (1 week, SIMPLE)
    - Add `--bytecode` flag to CLI
    - Modify `pkg/dwscript/dwscript.go` to support bytecode execution
    - Add `CompileMode` option (AST vs Bytecode)
    - Update benchmarks to compare modes

  - [ ] 12.18.13 Create bytecode test suite (1 week, MODERATE)
    - Port existing interpreter tests to bytecode
    - Test bytecode disassembler output
    - Verify identical behavior to AST interpreter
    - Add performance benchmarks

  - [ ] 12.18.14 Add bytecode serialization (optional) (3-5 days, SIMPLE)
    - Implement bytecode file format (.dwc)
    - Save/load compiled bytecode to disk
    - Version bytecode format for compatibility
    - Add `dwscript compile` command for bytecode

  - [ ] 12.18.15 Document bytecode VM (3 days, SIMPLE)
    - Write `docs/bytecode-vm.md` explaining architecture
    - Document instruction set and opcodes
    - Provide examples of bytecode output
    - Update CLAUDE.md with bytecode information

  **Phase 1 Expected Results**: 2-3x faster than tree-walking interpreter, reasonable complexity

  #### Phase 2: Optional LLVM-Based JIT (DEFER - 18-25 weeks, VERY COMPLEX)

  - [ ] 12.18.16 Set up LLVM infrastructure (1 week, COMPLEX)
    - Add `tinygo.org/x/go-llvm` dependency
    - Configure build tags for LLVM versions (14-20)
    - Create `internal/jit/` package
    - Set up CGo build configuration
    - Test on Linux, macOS, Windows (LLVM must be installed)
    - **Platform Limitation**: Requires system LLVM installation

  - [ ] 12.18.17 Implement LLVM IR generator for expressions (2-3 weeks, VERY COMPLEX)
    - Create `internal/jit/llvm_codegen.go`
    - Generate LLVM IR for arithmetic operations
    - Generate IR for comparisons and logic operations
    - Handle type conversions (int ↔ float ↔ string)
    - Implement constant folding in LLVM IR
    - Test with simple expressions

  - [ ] 12.18.18 Implement LLVM IR for control flow (2 weeks, VERY COMPLEX)
    - Generate IR for if/else statements (branch instructions)
    - Generate IR for while/for loops (phi nodes)
    - Handle break/continue/exit signals
    - Implement proper basic block structure

  - [ ] 12.18.19 Implement LLVM IR for function calls (2-3 weeks, VERY COMPLEX)
    - Define calling convention for DWScript functions
    - Generate IR for function declarations
    - Handle parameter passing (by value and by reference)
    - Implement return value handling
    - Support recursion and tail call optimization

  - [ ] 12.18.20 Implement LLVM IR for DWScript runtime (2-3 weeks, VERY COMPLEX)
    - Create runtime library for built-in functions (PrintLn, Length, etc.)
    - Implement dynamic dispatch for method calls
    - Handle exception propagation
    - Implement garbage collection interface (Go GC)
    - Support array/string operations

  - [ ] 12.18.21 Implement JIT compilation engine (2 weeks, COMPLEX)
    - Create `internal/jit/jit.go` with JIT compiler
    - Use LLVM MCJIT or ORC JIT engine
    - Compile LLVM IR to machine code at runtime
    - Cache compiled functions in memory
    - Add optimization passes (O2 or O3)

  - [ ] 12.18.22 Add profiling and hot path detection (1-2 weeks, COMPLEX)
    - Implement execution counter for functions/loops
    - Detect hot paths (> 1000 executions)
    - Trigger JIT compilation for hot functions
    - Fall back to bytecode for cold code
    - Implement tiered compilation strategy

  - [ ] 12.18.23 Handle FFI and external functions (1 week, COMPLEX)
    - Generate LLVM IR to call Go functions (via CGo)
    - Handle type marshaling (Go ↔ DWScript values)
    - Support callbacks from JIT code to interpreter
    - Test with external function registry

  - [ ] 12.18.24 Implement deoptimization (1-2 weeks, VERY COMPLEX)
    - Detect when JIT assumptions are violated (type changes)
    - Fall back to bytecode execution
    - Preserve execution state during deoptimization
    - Add guard conditions in JIT code

  - [ ] 12.18.25 Add JIT debugging support (1 week, MODERATE)
    - Generate debug info in LLVM IR
    - Preserve source line mapping
    - Support stack traces from JIT code
    - Add disassembly output for JIT code

  - [ ] 12.18.26 Optimize JIT compilation (2 weeks, COMPLEX)
    - Enable LLVM optimization passes (inlining, constant propagation)
    - Implement speculative optimizations
    - Add inline caching for method dispatch
    - Implement escape analysis for stack allocation

  - [ ] 12.18.27 Integrate JIT with bytecode VM (1 week, MODERATE)
    - Add `--jit` flag to CLI
    - Modify VM to call JIT-compiled code
    - Handle transitions between bytecode and JIT
    - Update performance benchmarks

  - [ ] 12.18.28 Test JIT on complex programs (1 week, MODERATE)
    - Run fixture test suite with JIT enabled
    - Compare output with interpreter and bytecode VM
    - Measure performance improvements
    - Test on Linux, macOS, Windows

  - [ ] 12.18.29 Handle platform-specific code generation (1 week, COMPLEX)
    - Support x86-64, ARM64 architectures
    - Handle calling convention differences
    - Test on different platforms
    - Add architecture detection

  - [ ] 12.18.30 Document JIT implementation (3 days, SIMPLE)
    - Write `docs/jit-compilation.md`
    - Explain LLVM integration
    - Provide performance benchmarks
    - Document platform requirements and limitations

  **Phase 2 Expected Results**: 5-10x faster than tree-walking, 2-3x faster than bytecode VM
  **Phase 2 Risk Level**: HIGH (complex, platform-dependent, maintenance burden)
  **Phase 2 Recommendation**: DEFER indefinitely - bytecode VM sufficient for most use cases

  #### Phase 3: Alternative Plugin-Based JIT (DEFER - 6-8 weeks, MODERATE)

  - [ ] 12.18.31 Implement Go code generation from AST (2-3 weeks, COMPLEX)
    - Create `internal/codegen/go_generator.go`
    - Generate Go source code from DWScript AST
    - Map DWScript types to Go types
    - Generate function declarations and calls
    - Handle closures and scoping
    - Test generated code compiles

  - [ ] 12.18.32 Implement plugin-based JIT (1-2 weeks, MODERATE)
    - Use `go build -buildmode=plugin` to compile generated code
    - Load plugin with `plugin.Open()`
    - Look up compiled function with `plugin.Lookup()`
    - Call compiled function from interpreter
    - Cache plugins to disk
    - **Platform Limitation**: No Windows support for plugins

  - [ ] 12.18.33 Add hot path detection for plugin JIT (1 week, MODERATE)
    - Track function execution counts
    - Trigger plugin compilation for hot functions
    - Manage plugin lifecycle (loading/unloading)

  - [ ] 12.18.34 Test plugin-based JIT (1 week, SIMPLE)
    - Run tests on Linux and macOS only
    - Compare performance with bytecode VM
    - Test plugin caching and reuse

  - [ ] 12.18.35 Document plugin approach (2 days, SIMPLE)
    - Write `docs/plugin-jit.md`
    - Explain platform limitations
    - Provide usage examples

  **Phase 3 Expected Results**: 3-5x faster than tree-walking
  **Phase 3 Limitations**: No Windows support, requires Go toolchain at runtime
  **Phase 3 Recommendation**: SKIP - poor portability

- [ ] 12.19 Add AOT compilation (compile to native binary) - **HIGH PRIORITY**

  **Feasibility**: Highly feasible and practical. AOT compilation aligns well with Go's strengths.

  **Recommended Approach**: Multi-target AOT - Transpile to Go (primary) + WASM (secondary) + Optional LLVM

  #### Phase 1: Go Source Code Generation (RECOMMENDED - 20-28 weeks)

  - [ ] 12.19.1 Design Go code generation architecture (1 week, MODERATE)
    - Study similar transpilers (c2go, ast-transpiler)
    - Design AST → Go AST transformation strategy
    - Define runtime library interface
    - Document type mapping (DWScript → Go)
    - Plan package structure for generated code
    - **Decision**: Use `go/ast` package for Go AST generation (type-safe, standard)

  - [ ] 12.19.2 Create Go code generator foundation (1 week, MODERATE)
    - Create `internal/codegen/` package
    - Create `internal/codegen/go_generator.go`
    - Implement `Generator` struct with context tracking
    - Add helper methods for code emission
    - Set up `go/ast` and `go/printer` integration
    - Create unit tests for basic generation

  - [ ] 12.19.3 Implement type system mapping (1-2 weeks, COMPLEX)
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

  - [ ] 12.19.4 Generate code for expressions (2 weeks, COMPLEX)
    - Generate literals (integer, float, string, boolean, nil)
    - Generate identifiers (variables, constants)
    - Generate binary operations (+, -, *, /, =, <>, <, >, etc.)
    - Generate unary operations (-, not)
    - Generate function calls
    - Generate array/object member access
    - Handle operator precedence correctly
    - Add unit tests comparing eval vs generated code

  - [ ] 12.19.5 Generate code for statements (2 weeks, COMPLEX)
    - Generate variable declarations (`var x: Integer = 42`)
    - Generate assignments (`x := 10`)
    - Generate if/else statements
    - Generate while/repeat/for loops
    - Generate case statements (switch in Go)
    - Generate begin...end blocks
    - Handle break/continue/exit statements

  - [ ] 12.19.6 Generate code for functions and procedures (2-3 weeks, COMPLEX)
    - Generate function declarations with parameters and return type
    - Handle by-value and by-reference (var) parameters
    - Generate procedure declarations (no return value)
    - Implement nested functions (closures in Go)
    - Support forward declarations
    - Handle recursion
    - Generate proper variable scoping

  - [ ] 12.19.7 Generate code for classes and OOP (2-3 weeks, VERY COMPLEX)
    - Generate Go struct definitions for classes
    - Generate constructor functions (Create)
    - Generate destructor cleanup (Destroy → defer)
    - Generate method declarations (receiver functions)
    - Implement inheritance (embedding in Go)
    - Implement virtual method dispatch (method tables)
    - Handle class fields and properties
    - Support `Self` keyword (receiver parameter)

  - [ ] 12.19.8 Generate code for interfaces (1-2 weeks, COMPLEX)
    - Generate Go interface definitions
    - Implement interface casting and type assertions
    - Generate interface method dispatch
    - Handle interface inheritance
    - Support interface variables and parameters

  - [ ] 12.19.9 Generate code for records (1 week, MODERATE)
    - Generate Go struct definitions
    - Support record methods (static and instance)
    - Handle record literals and initialization
    - Generate record field access

  - [ ] 12.19.10 Generate code for enums (1 week, MODERATE)
    - Generate Go const declarations with iota
    - Support scoped and unscoped enum access
    - Generate Ord() and Integer() conversions
    - Handle explicit enum values

  - [ ] 12.19.11 Generate code for arrays (1-2 weeks, COMPLEX)
    - Generate static arrays (Go arrays: `[10]int`)
    - Generate dynamic arrays (Go slices: `[]int`)
    - Support array literals
    - Generate array indexing and slicing
    - Implement SetLength, High, Low built-ins
    - Handle multi-dimensional arrays

  - [ ] 12.19.12 Generate code for sets (1 week, MODERATE)
    - Generate set types as Go map[T]bool or bitsets
    - Support set literals and constructors
    - Generate set operations (union, intersection, difference)
    - Implement `in` operator for set membership

  - [ ] 12.19.13 Generate code for properties (1 week, COMPLEX)
    - Translate properties to getter/setter methods
    - Generate field-backed properties (direct access)
    - Generate method-backed properties (method calls)
    - Support read-only and write-only properties
    - Handle auto-properties

  - [ ] 12.19.14 Generate code for exceptions (1-2 weeks, COMPLEX)
    - Generate try/except/finally as Go defer/recover
    - Map DWScript exceptions to Go error types
    - Generate raise statements (panic)
    - Implement exception class hierarchy
    - Preserve stack traces

  - [ ] 12.19.15 Generate code for operators and conversions (1 week, MODERATE)
    - Generate operator overloads as functions
    - Generate implicit conversions
    - Handle type coercion in expressions
    - Support custom operators

  - [ ] 12.19.16 Create runtime library for generated code (2-3 weeks, COMPLEX)
    - Create `pkg/runtime/` package
    - Implement built-in functions (PrintLn, Length, Copy, etc.)
    - Implement array/string manipulation functions
    - Implement math functions (Sin, Cos, Sqrt, etc.)
    - Implement date/time functions
    - Provide runtime type information (RTTI) for reflection
    - Support external function calls (FFI)

  - [ ] 12.19.17 Handle units/modules compilation (1-2 weeks, COMPLEX)
    - Generate separate Go packages for each unit
    - Handle unit dependencies and imports
    - Generate initialization/finalization code
    - Support uses clauses
    - Create package manifest

  - [ ] 12.19.18 Implement optimization passes (1-2 weeks, MODERATE)
    - Constant folding
    - Dead code elimination
    - Inline small functions
    - Remove unused variables
    - Optimize string concatenation
    - Use Go compiler optimization hints (//go:inline, etc.)

  - [ ] 12.19.19 Add source mapping for debugging (1 week, MODERATE)
    - Preserve line number comments in generated code
    - Generate source map files (.map)
    - Add DWScript source file embedding
    - Support stack trace translation (Go → DWScript)

  - [ ] 12.19.20 Test Go code generation (1-2 weeks, MODERATE)
    - Generate code for all fixture tests
    - Compile and run generated code
    - Compare output with interpreter
    - Measure compilation time
    - Benchmark generated code performance

  **Phase 1 Expected Results**: 10-50x faster than tree-walking interpreter, near-native Go speed

  #### Phase 2: AOT Compiler CLI (RECOMMENDED - 9-13 weeks)

  - [ ] 12.19.21 Create `dwscript compile` command (1 week, MODERATE)
    - Add `compile` subcommand to CLI
    - Parse input DWScript file(s)
    - Generate Go source code to output directory
    - Invoke `go build` to create executable
    - Support multiple output formats (executable, library, package)

  - [ ] 12.19.22 Implement project compilation mode (1-2 weeks, COMPLEX)
    - Support compiling entire projects (multiple units)
    - Generate go.mod file
    - Handle dependencies between units
    - Create main package with entry point
    - Support compilation configuration (optimization level, target platform)

  - [ ] 12.19.23 Add compilation flags and options (3-5 days, SIMPLE)
    - `--output` or `-o` for output path
    - `--optimize` or `-O` for optimization level (0, 1, 2, 3)
    - `--keep-go-source` to preserve generated Go files
    - `--target` for cross-compilation (linux, windows, darwin, wasm)
    - `--static` for static linking
    - `--debug` to include debug symbols

  - [ ] 12.19.24 Implement cross-compilation support (1 week, MODERATE)
    - Support GOOS and GOARCH environment variables
    - Generate platform-specific code (if needed)
    - Test compilation for Linux, macOS, Windows, WASM
    - Document platform-specific limitations

  - [ ] 12.19.25 Add incremental compilation (1-2 weeks, COMPLEX)
    - Cache compiled units
    - Detect file changes (mtime, hash)
    - Recompile only changed units
    - Rebuild dependency graph
    - Speed up repeated compilations

  - [ ] 12.19.26 Create standalone binary builder (1 week, MODERATE)
    - Generate single-file executable
    - Embed DWScript runtime
    - Strip debug symbols (optional)
    - Compress binary with UPX (optional)
    - Test on different platforms

  - [ ] 12.19.27 Implement library compilation mode (1 week, MODERATE)
    - Generate Go package (not executable)
    - Export public functions/classes
    - Create Go-friendly API
    - Generate documentation (godoc)
    - Support embedding in other Go projects

  - [ ] 12.19.28 Add compilation error reporting (3-5 days, MODERATE)
    - Catch Go compilation errors
    - Translate errors to DWScript source locations
    - Provide helpful error messages
    - Suggest fixes for common issues

  - [ ] 12.19.29 Create compilation test suite (1 week, MODERATE)
    - Test compilation of all fixture tests
    - Verify all executables run correctly
    - Test cross-compilation
    - Benchmark compilation speed
    - Measure binary sizes

  - [ ] 12.19.30 Document AOT compilation (3-5 days, SIMPLE)
    - Write `docs/aot-compilation.md`
    - Explain compilation process
    - Provide usage examples
    - Document performance characteristics
    - Compare with interpretation and JIT

  #### Phase 3: WebAssembly AOT (RECOMMENDED - 4-6 weeks)

  - [ ] 12.19.31 Extend WASM compilation for standalone binaries (1 week, MODERATE)
    - Generate WASM modules without JavaScript dependency
    - Use WASI for system calls
    - Support WASM-compatible runtime
    - Test with wasmtime, wasmer, wazero
    - **Note**: Much of this builds on task 11.15

  - [ ] 12.19.32 Optimize WASM binary size (1 week, MODERATE)
    - Use TinyGo compiler (smaller binaries)
    - Enable wasm-opt optimization
    - Strip unnecessary features
    - Measure binary size (target < 1MB)

  - [ ] 12.19.33 Add WASM runtime support (1 week, MODERATE)
    - Bundle WASM runtime (wasmer-go or wazero)
    - Create launcher executable
    - Support both JIT and AOT WASM execution
    - Test performance

  - [ ] 12.19.34 Test WASM AOT compilation (3-5 days, SIMPLE)
    - Compile fixture tests to WASM
    - Run with different WASM runtimes
    - Measure performance vs native
    - Test browser and server execution

  - [ ] 12.19.35 Document WASM AOT (2-3 days, SIMPLE)
    - Write `docs/wasm-aot.md`
    - Explain WASM compilation process
    - Provide deployment examples
    - Compare with Go native compilation

  **Phase 3 Expected Results**: 5-20x speedup (browser), 10-30x speedup (WASI runtime)

  #### Phase 4: Optional LLVM AOT (DEFER - 5-7 weeks, VERY COMPLEX)

  - [ ] 12.19.36 Implement LLVM IR generation (reuse JIT work) (2-3 weeks, VERY COMPLEX)
    - Extend `internal/jit/llvm_codegen.go` for AOT
    - Generate complete LLVM IR module
    - Support all DWScript features
    - Add LLVM optimization passes
    - **Prerequisite**: Complete task 11.18 Phase 2 (LLVM JIT) first

  - [ ] 12.19.37 Compile LLVM IR to object files (1 week, COMPLEX)
    - Use LLVM static compiler (llc)
    - Generate object files (.o)
    - Link with DWScript runtime
    - Create executable

  - [ ] 12.19.38 Implement LLVM-based cross-compilation (1 week, COMPLEX)
    - Support multiple target triples (x86_64, arm64, etc.)
    - Generate platform-specific code
    - Handle calling convention differences

  - [ ] 12.19.39 Test LLVM AOT compilation (1 week, MODERATE)
    - Compile fixture tests with LLVM
    - Compare performance with Go AOT
    - Measure binary sizes
    - Test on different platforms

  - [ ] 12.19.40 Document LLVM AOT (2 days, SIMPLE)
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

- [ ] 12.20 Add compilation to Go source code (MERGED INTO 11.19 PHASE 1)
  - This task is now covered by 11.19.1-11.19.20 (Go source code generation)

- [ ] 12.21 Benchmark different execution modes
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

---

## Stage 13: Code Generation - Multi-Backend Architecture

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

#### 13.1.1: MIR Package Structure and Types (10 tasks)

- [ ] 13.1 Create `mir/` package directory
- [ ] 13.2 Create `mir/types.go` - MIR type system
- [ ] 13.3 Define `Type` interface with `String()`, `Size()`, `Align()` methods
- [ ] 13.4 Implement primitive types: `Bool`, `Int8`, `Int16`, `Int32`, `Int64`, `Float32`, `Float64`, `String`
- [ ] 13.5 Implement composite types: `Array(elemType, size)`, `Record(fields)`, `Pointer(pointeeType)`
- [ ] 13.6 Implement OOP types: `Class(name, fields, methods, parent)`, `Interface(name, methods)`
- [ ] 13.7 Implement function types: `Function(params, returnType)`
- [ ] 13.8 Add `Void` type for procedures
- [ ] 13.9 Implement type equality and compatibility checking
- [ ] 13.10 Implement type conversion rules (explicit vs implicit)

#### 13.1.2: MIR Instructions and Control Flow (10 tasks)

- [ ] 13.11 Create `mir/instruction.go` - MIR instruction set
- [ ] 13.12 Define `Instruction` interface with `ID()`, `Type()`, `String()` methods
- [ ] 13.13 Implement arithmetic ops: `Add`, `Sub`, `Mul`, `Div`, `Mod`, `Neg`
- [ ] 13.14 Implement comparison ops: `Eq`, `Ne`, `Lt`, `Le`, `Gt`, `Ge`
- [ ] 13.15 Implement logical ops: `And`, `Or`, `Xor`, `Not`
- [ ] 13.16 Implement memory ops: `Alloca`, `Load`, `Store`
- [ ] 13.17 Implement constants: `ConstInt`, `ConstFloat`, `ConstString`, `ConstBool`, `ConstNil`
- [ ] 13.18 Implement conversions: `IntToFloat`, `FloatToInt`, `IntTrunc`, `IntExt`
- [ ] 13.19 Implement function ops: `Call`, `VirtualCall`
- [ ] 13.20 Implement array/class ops: `ArrayAlloc`, `ArrayLen`, `ArrayIndex`, `ArraySet`, `FieldGet`, `FieldSet`, `New`

#### 13.1.3: MIR Control Flow Structures (5 tasks)

- [ ] 13.21 Create `mir/block.go` - Basic blocks with `ID`, `Instructions`, `Terminator`
- [ ] 13.22 Implement control flow terminators: `Phi`, `Br`, `CondBr`, `Return`, `Throw`
- [ ] 13.23 Implement terminator validation (every block must end with terminator)
- [ ] 13.24 Implement block predecessors/successors tracking for CFG
- [ ] 13.25 Create `mir/function.go` - Function representation with `Name`, `Params`, `ReturnType`, `Blocks`, `Locals`

#### 13.1.4: MIR Builder API (3 tasks)

- [ ] 13.26 Create `mir/builder.go` - Safe MIR construction
- [ ] 13.27 Implement `Builder` struct with function/block context, `NewFunction()`, `NewBlock()`, `SetInsertPoint()`
- [ ] 13.28 Implement instruction emission methods: `EmitAdd()`, `EmitLoad()`, `EmitStore()`, etc. with type checking

#### 13.1.5: MIR Verifier (2 tasks)

- [ ] 13.29 Create `mir/verifier.go` - MIR correctness checking
- [ ] 13.30 Implement CFG, type, SSA, and function signature verification with `Verify(fn *Function) []error` API

### Stage 13.2: AST → MIR Lowering (12 tasks)

- [ ] 13.31 Create `mir/lower.go` - AST to MIR translation
- [ ] 13.32 Implement `LowerProgram(ast *ast.Program) (*mir.Module, error)` entry point
- [ ] 13.33 Lower expressions: literals → `Const*` instructions
- [ ] 13.34 Lower binary operations → corresponding MIR ops (handle short-circuit for `and`/`or`)
- [ ] 13.35 Lower unary operations → `Neg`, `Not`
- [ ] 13.36 Lower identifier references → `Load` instructions
- [ ] 13.37 Lower function calls → `Call` instructions
- [ ] 13.38 Lower array indexing → `ArrayIndex` + bounds check insertion
- [ ] 13.39 Lower record field access → `FieldGet`/`FieldSet`
- [ ] 13.40 Lower statements: variable declarations, assignments, if/while/for, return
- [ ] 13.41 Lower declarations: functions/procedures, records, classes
- [ ] 13.42 Implement short-circuit evaluation and simple optimizations (constant folding, dead code elimination)

### Stage 12.3: MIR Debugging and Testing (5 tasks)

- [ ] 13.43 Create `mir/dump.go` - Human-readable MIR output with `Dump(fn *Function) string`
- [ ] 13.44 Integration with CLI: `./bin/dwscript dump-mir script.dws`
- [ ] 13.45 Create golden MIR tests: 5+ each for expressions, control flow, functions, advanced features
- [ ] 13.46 Implement MIR verifier tests: type mismatches, malformed CFG, SSA violations
- [ ] 13.47 Implement round-trip tests: AST → MIR → verify → dump → compare with golden files

### Stage 12.4: JS Backend MVP (45 tasks)

**Goal**: Implement a JavaScript code generator that can compile basic DWScript programs to readable, runnable JavaScript.

**Exit Criteria**: JS emitter for expressions/control flow/functions, 20+ end-to-end tests (DWScript→JS→execute), golden JS snapshots, 85%+ coverage

#### 12.4.1: JS Emitter Infrastructure (8 tasks)

- [ ] 13.48 Create `codegen/` package with `Backend` interface and `EmitterOptions`
- [ ] 13.49 Create `codegen/js/` package and `emitter.go`
- [ ] 13.50 Define `JSEmitter` struct with `out`, `indent`, `opts`, `tmpCounter`
- [ ] 13.51 Implement helper methods: `emit()`, `emitLine()`, `emitIndent()`, `pushIndent()`, `popIndent()`
- [ ] 13.52 Implement `newTemp()` for temporary variable naming
- [ ] 13.53 Implement `NewJSEmitter(opts EmitterOptions)`
- [ ] 13.54 Implement `Generate(module *mir.Module) (string, error)` entry point
- [ ] 13.55 Test emitter infrastructure

#### 12.4.2: Module and Function Emission (6 tasks)

- [ ] 13.56 Implement module structure emission: ES Module format with `export`, file header comment
- [ ] 13.57 Implement optional IIFE fallback via `EmitterOptions`
- [ ] 13.58 Implement function emission: `function fname(params) { ... }`
- [ ] 13.59 Map DWScript params to JS params (preserve names)
- [ ] 13.60 Emit local variable declarations at function top (from `Alloca` instructions)
- [ ] 13.61 Handle procedures (no return value) as JS functions

#### 12.4.3: Expression and Instruction Lowering (12 tasks)

- [ ] 13.62 Lower arithmetic operations → JS infix operators: `+`, `-`, `*`, `/`, `%`, unary `-`
- [ ] 13.63 Lower comparison operations → JS comparisons: `===`, `!==`, `<`, `<=`, `>`, `>=`
- [ ] 13.64 Lower logical operations → JS boolean ops: `&&`, `||`, `!`
- [ ] 13.65 Lower constants → JS literals with proper escaping
- [ ] 13.66 Lower variable operations: `Load` → variable reference, `Store` → assignment
- [ ] 13.67 Lower function calls: `Call` → `functionName(args)`
- [ ] 13.68 Implement Phi node lowering with temporary variables at block edges
- [ ] 13.69 Test expression lowering
- [ ] 13.70 Test instruction lowering
- [ ] 13.71 Test temporary variable generation
- [ ] 13.72 Test type conversions
- [ ] 13.73 Test complex expressions

#### 12.4.4: Control Flow Emission (8 tasks)

- [ ] 13.74 Implement control flow reconstruction from MIR CFG
- [ ] 13.75 Detect if/else patterns from `CondBr`
- [ ] 13.76 Detect while loop patterns (backedge to header)
- [ ] 13.77 Emit if-else: `if (condition) { ... } else { ... }`
- [ ] 13.78 Emit while loops: `while (condition) { ... }`
- [ ] 13.79 Emit for loops if MIR preserves metadata
- [ ] 13.80 Handle unconditional branches
- [ ] 13.81 Handle return statements

#### 12.4.5: Runtime and Testing (11 tasks)

- [ ] 13.82 Create `runtime/js/runtime.js` with `_dws.boundsCheck()`, `_dws.assert()`
- [ ] 13.83 Emit runtime import in generated JS (if needed)
- [ ] 13.84 Make runtime usage optional via `EmitterOptions.InsertBoundsChecks`
- [ ] 13.85 Create `codegen/js/testdata/` with subdirectories
- [ ] 13.86 Implement golden JS snapshot tests
- [ ] 13.87 Setup Node.js in CI (GitHub Actions)
- [ ] 13.88 Implement execution tests: parse → lower → generate → execute → verify
- [ ] 13.89 Add end-to-end tests for arithmetic, control flow, functions, loops
- [ ] 13.90 Add unit tests for JS emitter
- [ ] 13.91 Achieve 85%+ coverage for `codegen/js/` package
- [ ] 13.92 Add `compile-js` CLI command: `./bin/dwscript compile-js input.dws -o output.js`

### Stage 12.5: JS Feature Complete (60 tasks)

**Goal**: Extend JS backend to support all DWScript language features.

**Exit Criteria**: Full OOP, composite types, exceptions, properties, 50+ comprehensive tests, real-world samples work

#### 12.5.1: Records (7 tasks)

- [ ] 13.93 Implement MIR support for records
- [ ] 13.94 Emit records as plain JS objects: `{ x: 0, y: 0 }`
- [ ] 13.95 Implement constructor functions for records
- [ ] 13.96 Implement field access/assignment as property access
- [ ] 13.97 Implement record copy semantics with `_dws.copyRecord()`
- [ ] 13.98 Test record creation, initialization, field read/write
- [ ] 13.99 Test nested records and copy semantics

#### 12.5.2: Arrays (10 tasks)

- [ ] 13.100 Extend MIR for static and dynamic arrays
- [ ] 13.101 Emit static arrays as JS arrays with fixed size
- [ ] 13.102 Implement array index access with optional bounds checking
- [ ] 13.103 Emit dynamic arrays as JS arrays
- [ ] 13.104 Implement `SetLength` → `arr.length = newLen`
- [ ] 13.105 Implement `Length` → `arr.length`
- [ ] 13.106 Support multi-dimensional arrays (nested JS arrays)
- [ ] 13.107 Implement array operations: copy, concatenation
- [ ] 13.108 Test static array creation and indexing
- [ ] 13.109 Test dynamic array operations and bounds checking

#### 12.5.3: Classes and Inheritance (15 tasks)

- [ ] 13.110 Extend MIR for classes with fields, methods, parent, vtable
- [ ] 13.111 Emit ES6 class syntax: `class TAnimal { ... }`
- [ ] 13.112 Implement field initialization in constructor
- [ ] 13.113 Implement method emission
- [ ] 13.114 Implement inheritance with `extends` clause
- [ ] 13.115 Implement `super()` call in constructor
- [ ] 13.116 Handle virtual method dispatch (naturally virtual in JS)
- [ ] 13.117 Handle DWScript `Create` → JS `constructor`
- [ ] 13.118 Handle multiple constructors (overload dispatch)
- [ ] 13.119 Document destructor handling (no direct equivalent in JS)
- [ ] 13.120 Implement static fields and methods
- [ ] 13.121 Map `Self` → `this`, `inherited` → `super.method()`
- [ ] 13.122 Test simple classes with fields and methods
- [ ] 13.123 Test inheritance, virtual method overriding, constructors
- [ ] 13.124 Test static members and `Self`/`inherited` usage

#### 12.5.4: Interfaces (6 tasks)

- [ ] 13.125 Extend MIR for interfaces
- [ ] 13.126 Choose and document JS emission strategy (structural typing vs runtime metadata)
- [ ] 13.127 If using runtime metadata: emit interface tables, implement `is`/`as` operators
- [ ] 13.128 Test class implementing interface
- [ ] 13.129 Test interface method calls
- [ ] 13.130 Test `is` and `as` with interfaces

#### 12.5.5: Enums and Sets (8 tasks)

- [ ] 13.131 Extend MIR for enums
- [ ] 13.132 Emit enums as frozen JS objects: `const TColor = Object.freeze({...})`
- [ ] 13.133 Support scoped and unscoped enum access
- [ ] 13.134 Extend MIR for sets
- [ ] 13.135 Emit small sets (≤32 elements) as bitmasks
- [ ] 13.136 Emit large sets as JS `Set` objects
- [ ] 13.137 Implement set operations: union, intersection, difference, inclusion
- [ ] 13.138 Test enum declaration/usage and set operations

#### 12.5.6: Exception Handling (8 tasks)

- [ ] 13.139 Extend MIR for exceptions: `Throw`, `Try`, `Catch`, `Finally`
- [ ] 13.140 Emit `Throw` → `throw new Error()` or custom exception class
- [ ] 13.141 Emit try-except-finally → JS `try/catch/finally`
- [ ] 13.142 Create DWScript exception class → JS `Error` subclass
- [ ] 13.143 Handle `On E: ExceptionType do` with instanceof checks
- [ ] 13.144 Implement re-raise with exception tracking
- [ ] 13.145 Test basic try-except, multiple handlers, try-finally
- [ ] 13.146 Test re-raise and nested exception handling

#### 12.5.7: Properties and Advanced Features (6 tasks)

- [ ] 13.147 Extend MIR for properties with `PropGet`/`PropSet`
- [ ] 13.148 Emit properties as ES6 getters/setters
- [ ] 13.149 Handle indexed properties as methods
- [ ] 13.150 Test read/write properties and indexed properties
- [ ] 13.151 Implement operator overloading (desugar to method calls)
- [ ] 13.152 Implement generics support (monomorphization)

### Stage 12.6: LLVM Backend [OPTIONAL - Future Work] (45 tasks)

**Goal**: Implement LLVM IR backend for native code compilation. This is **deferred** and optional.

**Exit Criteria**: Valid LLVM IR generation, runtime library in C, basic end-to-end tests, documentation

#### 12.6.1: LLVM Infrastructure (8 tasks)

- [ ] 13.153 Choose LLVM binding: `llir/llvm` (pure Go) vs CGo bindings
- [ ] 13.154 Create `codegen/llvm/` package with `emitter.go`, `types.go`, `runtime.go`
- [ ] 13.155 Implement type mapping: DWScript types → LLVM types
- [ ] 13.156 Map Integer → `i32`/`i64`, Float → `double`, Boolean → `i1`
- [ ] 13.157 Map String → struct `{i32 len, i8* data}`
- [ ] 13.158 Map arrays/objects to LLVM structs
- [ ] 13.159 Emit LLVM module with target triple
- [ ] 13.160 Declare external runtime functions

#### 12.6.2: Runtime Library (12 tasks)

- [ ] 13.161 Create `runtime/dws_runtime.h` - C header for runtime API
- [ ] 13.162 Declare string operations: `dws_string_new()`, `dws_string_concat()`, `dws_string_len()`
- [ ] 13.163 Declare array operations: `dws_array_new()`, `dws_array_index()`, `dws_array_len()`
- [ ] 13.164 Declare memory management: `dws_alloc()`, `dws_free()`
- [ ] 13.165 Choose and document memory strategy (Boehm GC vs reference counting)
- [ ] 13.166 Declare object operations: `dws_object_new()`, virtual dispatch helpers
- [ ] 13.167 Declare exception handling: `dws_throw()`, `dws_catch()`
- [ ] 13.168 Declare RTTI: `dws_is_instance()`, `dws_as_instance()`
- [ ] 13.169 Create `runtime/dws_runtime.c` - implement runtime
- [ ] 13.170 Implement all runtime functions
- [ ] 13.171 Create `runtime/Makefile` to build `libdws_runtime.a`
- [ ] 13.172 Add runtime build to CI for Linux/macOS/Windows

#### 12.6.3: LLVM Code Emission (15 tasks)

- [ ] 13.173 Implement LLVM emitter: `Generate(module *mir.Module) (string, error)`
- [ ] 13.174 Emit function declarations with correct signatures
- [ ] 13.175 Emit basic blocks for each MIR block
- [ ] 13.176 Emit arithmetic instructions: `add`, `sub`, `mul`, `sdiv`, `srem`
- [ ] 13.177 Emit comparison instructions: `icmp eq`, `icmp slt`, etc.
- [ ] 13.178 Emit logical instructions: `and`, `or`, `xor`
- [ ] 13.179 Emit memory instructions: `alloca`, `load`, `store`
- [ ] 13.180 Emit call instructions: `call @function_name(args)`
- [ ] 13.181 Emit constants: integers, floats, strings
- [ ] 13.182 Emit control flow: conditional branches, phi nodes
- [ ] 13.183 Emit runtime calls for strings, arrays, objects
- [ ] 13.184 Implement type conversions: `sitofp`, `fptosi`
- [ ] 13.185 Emit struct types for classes and vtables
- [ ] 13.186 Implement virtual method dispatch
- [ ] 13.187 Implement exception handling (simple throw/catch or full LLVM EH)

#### 12.6.4: Linking and Testing (7 tasks)

- [ ] 13.188 Implement compilation pipeline: DWScript → MIR → LLVM IR → object → executable
- [ ] 13.189 Integrate `llc` to compile .ll → .o
- [ ] 13.190 Integrate linker to link object + runtime → executable
- [ ] 13.191 Add `compile-native` CLI command
- [ ] 13.192 Create 10+ end-to-end tests: DWScript → native → execute → verify
- [ ] 13.193 Benchmark JS vs native performance
- [ ] 13.194 Document LLVM backend in `docs/llvm-backend.md`

#### 13.6.5: Documentation (3 tasks)

- [ ] 13.195 Create `docs/codegen-architecture.md` - MIR overview, multi-backend design
- [ ] 13.196 Create `docs/mir-spec.md` - complete MIR reference with examples
- [ ] 13.197 Create `docs/js-backend.md` - DWScript → JavaScript mapping guide

---

## Phase 14: AST-Driven Formatter and Playground Integration 🆕 **PLANNED**

Goal: deliver an auto-formatting pipeline that reuses the existing AST and semantic metadata to produce canonical DWScript source, accessible via the CLI (`dwscript fmt`), editors, and the web playground.

### 14.1 Specification & AST/Data Prep (7 tasks)

- [x] 14.1.1 Capture formatting requirements from upstream DWScript (indent width, begin/end alignment, keyword casing, line-wrapping) and document them in `docs/formatter-style-guide.md`.
- [x] 14.1.2 Audit current AST nodes for source position fidelity and comment/trivia preservation; list any nodes lacking `Pos` / `EndPos`.
- [ ] 14.1.3 Extend the parser/AST to track leading and trailing trivia (single-line, block comments, blank lines) without disturbing semantic passes.
- [ ] 14.1.4 Define a `format.Options` struct (indent size, max line length, newline style) and default profile matching DWScript conventions.
- [ ] 14.1.5 Build a formatting test corpus in `testdata/formatter/{input,expected}` with tricky constructs (nested classes, generics, properties, preprocessor).
- [ ] 14.1.6 Add helper APIs to serialize AST back into token streams (e.g., `ast.FormatNode`, `ast.IterChildren`) to keep formatter logic decoupled from parser internals.
- [ ] 14.1.7 Ensure the semantic/type metadata needed for spacing decisions (e.g., `var` params, attributes) is exposed through lightweight inspector interfaces to avoid circular imports.

### 13.2 Formatter Engine Implementation (10 tasks)

- [ ] 14.2.1 Create `formatter` package with a multi-phase pipeline: AST normalization → layout planning → text emission.
- [ ] 14.2.2 Implement a visitor that emits `format.Node` instructions (indent/dedent, soft break, literal text) for statements and declarations, leveraging AST shape rather than raw tokens.
- [ ] 14.2.3 Handle block constructs (`begin...end`, class bodies, `case` arms) with indentation stacks so nested scopes auto-align.
- [ ] 14.2.4 Add expression formatting that respects operator precedence and inserts parentheses only when required; reuse existing precedence tables.
- [ ] 14.2.5 Support alignment for parameter lists, generics, array types, and property declarations with configurable wrap points.
- [ ] 14.2.6 Preserve user comments: attach leading comments before the owning node, keep inline comments after statements, and maintain blank-line intent (max consecutives configurable).
- [ ] 14.2.7 Implement whitespace normalization rules (single spaces around binary operators, before `do`/`then`, after commas, etc.).
- [ ] 14.2.8 Provide idempotency guarantees by building a golden test that pipes formatted output back through the formatter and asserts stability.
- [ ] 14.2.9 Expose a streaming writer that emits `[]byte`/`io.Writer` output to keep the CLI fast and low-memory.
- [ ] 14.2.10 Benchmark formatting of large fixtures (≥5k LOC) and optimize hot paths (string builder pools, avoiding interface allocations).

### 13.3 Tooling & Playground Integration (7 tasks)

- [ ] 14.3.1 Wire a new CLI command `dwscript fmt` (and `fmt -w`) that runs the formatter over files/directories, mirroring `gofmt` UX.
- [ ] 14.3.2 Update the WASM bridge to expose a `Format(source string) (string, error)` hook exported from Go, reusing the same formatter package.
- [ ] 14.3.3 Modify `playground/js/playground.js` to call the WASM formatter before falling back to Monaco’s default action, enabling deterministic formatting in the browser.
- [ ] 14.3.4 Add formatter support to the VSCode extension / LSP stub (if present) so editors can trigger `textDocument/formatting`.
- [ ] 14.3.5 Ensure the formatter respects partial-range requests (`textDocument/rangeFormatting`) to avoid reformatting entire files when not desired.
- [ ] 14.3.6 Introduce CI checks (`just fmt-check`) that fail when files are not formatted, and document the workflow in `CONTRIBUTING.md`.
- [ ] 14.3.7 Provide sample scripts/snippets (e.g., Git hooks) encouraging contributors to run the formatter.

### 13.4 Validation, UX, and Docs (6 tasks)

- [ ] 14.4.1 Create table-driven unit tests per node type plus integration tests that read `testdata/formatter` fixtures.
- [ ] 14.4.2 Add fuzz/property tests that compare formatter output against itself round-tripped through the parser → formatter pipeline.
- [ ] 14.4.3 Document formatter architecture and extension points in `docs/formatter-architecture.md`.
- [ ] 14.4.4 Update `PLAYGROUND.md`, `README.md`, and release notes to mention the Format button now runs the AST-driven formatter.
- [ ] 14.4.5 Record known limitations (e.g., preprocessor directives) and track follow-ups in `TEST_ISSUES.md`.
- [ ] 14.4.6 Gather usability feedback (issue template or telemetry) to prioritize refinements like configurable styles or multi-profile support.

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
