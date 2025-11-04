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

### Advanced FFI Features (Tasks 9.13-9.18)

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
// Variadic functions (9.13)
external function Sum(numbers: array of Integer): Integer;  // Go: func Sum(numbers ...int) int

// Optional parameters (9.14)
external function Greet(name: String; prefix: String = 'Hello'): String;

// By-reference parameters (9.15)
external procedure Swap(var a, b: Integer);  // Go: func Swap(a, b *int)

// Callbacks (9.17)
external procedure ForEach(arr: array of Integer; callback: TIntProc);
type TIntProc = procedure(value: Integer);
```

- [ ] 9.13 Support variadic Go functions:

  - [ ] 9.13a Detect variadic parameters in Go reflection:
    - **File**: `pkg/dwscript/ffi.go` (function `RegisterFunction`)
    - Use `reflect.Type.IsVariadic()` to detect `...T` parameters
    - Store variadic flag in function metadata
    - Note: Last parameter will be slice type (`[]T`)

  - [ ] 9.13b Accept variable number of DWScript arguments:
    - **File**: `internal/interp/ffi.go` (function that calls Go functions)
    - For variadic Go functions:
      - Allow any number of arguments ≥ (required params)
      - Collect extra arguments into a slice
      - Example: Go `func(a int, nums ...int)` accepts DWScript calls with 1+ arguments
    - Update argument count validation to handle variadic case

  - [ ] 9.13c Pack variadic arguments into slice:
    - **File**: `internal/interp/marshal.go`
    - When marshaling arguments for variadic function:
      - Take arguments beyond required count
      - Convert each to Go type (using existing marshaling)
      - Pack into `[]T` slice
      - Pass slice as last argument to Go function
    - **Example**:
      ```go
      // Go: func Sum(nums ...int) int
      // DWScript: Sum(1, 2, 3, 4)
      // Marshal: []int{1, 2, 3, 4}
      ```

  - [ ] 9.13d Add tests for variadic functions:
    - **File**: `pkg/dwscript/ffi_test.go`
    - Test variadic with 0 extra args (just required params)
    - Test variadic with multiple extra args
    - Test variadic with only variadic param (all args variadic)
    - Test type marshaling for variadic slice

- [ ] 9.14 Support optional parameters:

  - [ ] 9.14a Parser: Support default parameter values:
    - **File**: `internal/parser/declarations.go` (function parameter parsing)
    - Parse `param: Type = defaultValue` syntax
    - Store default value expression in `Parameter` AST node
    - **AST change**: Add `DefaultValue Expression` to `Parameter` struct

  - [ ] 9.14b Semantic analysis: Validate default values:
    - **File**: `internal/semantic/analyze_declarations.go`
    - Type-check default value expression
    - Ensure type matches parameter type
    - Evaluate constant expressions at compile time (if possible)
    - Rules: Optional params must come after required params

  - [ ] 9.14c Interpreter: Handle missing optional arguments:
    - **File**: `internal/interp/expressions.go` (function call evaluation)
    - When calling function with optional params:
      - If argument not provided, use default value
      - Evaluate default value expression in caller's context
      - Pass to function

  - [ ] 9.14d FFI: Map optional params to Go:
    - **File**: `pkg/dwscript/ffi.go`
    - **Challenge**: Go doesn't have optional parameters
    - **Option 1**: Use function overloading (multiple Go functions)
    - **Option 2**: Use variadic with type checking
    - **Option 3**: Require all params, DWScript fills defaults before calling Go
    - **Recommended**: Option 3 - simplest, no Go changes needed
    - DWScript calls Go function with all parameters (fills in defaults)

  - [ ] 9.14e Add tests:
    - **File**: `internal/parser/declarations_test.go` - parse default values
    - **File**: `internal/semantic/analyze_declarations_test.go` - validate defaults
    - **File**: `internal/interp/functions_test.go` - call with/without optional args
    - **File**: `pkg/dwscript/ffi_test.go` - FFI with optional params

- [ ] 9.15 Support by-reference parameters (var keyword):

  - [x] 9.15a Parser: Support `var` parameter keyword ✅ PARTIAL
    - Already implemented (PLAN.md line 47: "parsing implemented")
    - **File**: `internal/parser/declarations.go`
    - Parses `var param: Type` syntax
    - Stores in AST

  - [ ] 9.15b Semantic analysis: Track var parameters:
    - **File**: `internal/semantic/analyze_declarations.go`
    - Mark parameter as by-reference in type system
    - Validate: var param must be assignable lvalue (can't pass constant)
    - **Type**: Add `IsByRef bool` to `FunctionType.Params`

  - [ ] 9.15c Interpreter: Pass by reference for var params:
    - **File**: `internal/interp/expressions.go` (function calls)
    - For `var` parameters:
      - Pass reference/pointer to variable (not copy of value)
      - After function returns, variable is updated
    - **Implementation**: Use pointer or special Reference type?
    - **Challenge**: Go's pass-by-value semantics
    - **Solution**: Wrap in mutable cell/reference object

  - [ ] 9.15d FFI: Sync changes back to DWScript:
    - **File**: `internal/interp/ffi.go`
    - For `var` parameters mapped to Go pointers:
      1. Before call: Marshal DWScript value, take address
      2. Call Go function with pointer
      3. After call: Unmarshal modified value back to DWScript variable
    - **Reflection**: Use `reflect.Value.Elem()` to deref pointer
    - **Marshaling**: Extend marshal layer to handle `*int`, `*string`, etc.

  - [ ] 9.15e Add tests:
    - **File**: `internal/interp/functions_test.go`
    - Test DWScript function with var param modifies caller's variable
    - Test nested var param calls
    - **File**: `pkg/dwscript/ffi_test.go`
    - Test Go function with pointer param modifies DWScript variable
    - Example: `Swap(var a, b: Integer)` swaps values

- [ ] 9.16 Support registering Go methods:

  - [ ] 9.16a Detect method vs function in reflection:
    - **File**: `pkg/dwscript/ffi.go` (function `RegisterFunction`)
    - Use `reflect.Type.NumIn()` and check if first param is receiver
    - **Challenge**: Go reflection doesn't distinguish methods from functions
    - **Workaround**: Accept method value (`obj.Method`) which has receiver bound

  - [ ] 9.16b Add `RegisterMethod` API:
    - **File**: `pkg/dwscript/ffi.go`
    - New function: `RegisterMethod(name string, receiver any, method any) error`
    - Extract receiver type and method
    - Store receiver in closure so it's available when called
    - **Alternative**: Just use method values with existing `RegisterFunction`

  - [ ] 9.16c Bind receiver automatically:
    - When DWScript calls registered method:
      - Receiver already bound in Go (closure or method value)
      - Just marshal arguments and call
      - No special handling needed if using method values

  - [ ] 9.16d Add tests:
    - **File**: `pkg/dwscript/ffi_test.go`
    - Register method from Go struct
    - Call from DWScript
    - Verify receiver state is accessible
    - Test with both value and pointer receivers

- [ ] 9.17 Support callback functions (DWScript → Go → DWScript):

  - [ ] 9.17a Marshal DWScript function pointers to Go:
    - **File**: `internal/interp/marshal.go` (function `marshalToGo`)
    - Detect when DWScript value is function pointer (`FunctionValue`)
    - Create Go function that wraps DWScript function
    - **Implementation**:
      ```go
      // DWScript: func(x: Integer): Integer
      // Go wrapper: func(x int) int {
      //   return callDWScriptFunction(dwsFunc, x)
      // }
      ```
    - Use reflection to create correct function signature

  - [ ] 9.17b Implement DWScript function call from Go:
    - **File**: `internal/interp/ffi.go`
    - Function: `callDWScriptFunction(fn *FunctionValue, args ...any) any`
    - Steps:
      1. Marshal Go arguments to DWScript values
      2. Create execution environment
      3. Call DWScript function (reuse interpreter)
      4. Get return value
      5. Marshal back to Go
    - **Thread safety**: Ensure interpreter state is safe for re-entry

  - [ ] 9.17c Handle re-entrancy:
    - **Challenge**: Go calls DWScript, which calls Go, which calls DWScript...
    - **Solution**: Stack-based execution contexts
    - Track call depth to prevent infinite recursion
    - **Goroutine safety**: Lock interpreter state? Or require single-threaded?
    - **File**: `internal/interp/interpreter.go`
    - Add execution context stack
    - Check max depth (e.g., 1000 calls)

  - [ ] 9.17d Handle callback errors and panics:
    - If DWScript callback raises exception:
      - Convert to Go error/panic
      - Propagate to Go caller
    - If DWScript callback panics:
      - Recover and convert to error
      - Clean up execution state

  - [ ] 9.17e Add tests:
    - **File**: `pkg/dwscript/ffi_test.go`
    - **Test**: `TestCallback` - Go function accepts DWScript function, calls it
    - **Test**: `TestNestedCallback` - DWScript → Go → DWScript → Go
    - **Test**: `TestCallbackError` - DWScript callback raises exception
    - **Test**: `TestCallbackRecursion` - Detect infinite recursion
    - **Example**:
      ```go
      // Go function
      func ForEach(items []int, callback func(int)) {
          for _, item := range items {
              callback(item)
          }
      }

      // DWScript
      ForEach([1,2,3], lambda(x) => PrintLn(x));
      ```

- [ ] 9.18 Add comprehensive tests for advanced FFI features:

  - [ ] 9.18a Integration tests combining features:
    - **File**: `pkg/dwscript/ffi_integration_test.go`
    - Variadic Go function with callbacks
    - Optional params with var params
    - Methods with callbacks
    - Complex scenarios

  - [ ] 9.18b Error handling tests:
    - Invalid variadic arguments
    - Wrong types for optional params
    - Nil pointers for var params
    - Callback type mismatches
    - Re-entrancy limit exceeded

  - [ ] 9.18c Performance benchmarks:
    - **File**: `pkg/dwscript/ffi_bench_test.go`
    - Benchmark FFI call overhead
    - Benchmark marshaling cost for various types
    - Benchmark callback overhead (DWScript → Go → DWScript)
    - Compare to direct DWScript function calls

  - [ ] 9.18d Documentation:
    - **File**: `docs/ffi.md` (create or extend)
    - Document all advanced FFI features
    - Code examples for each feature
    - Best practices
    - Performance considerations
    - Thread safety notes

  - [ ] 9.18e Example programs:
    - **Directory**: `examples/ffi/`
    - `variadic.go` + `variadic.dws` - demonstrate variadic
    - `callbacks.go` + `callbacks.dws` - demonstrate callbacks
    - `methods.go` + `methods.dws` - demonstrate methods
    - `full_example.go` + `full_example.dws` - combine all features

---

#### Full Contextual Type Inference (FUTURE ENHANCEMENT)

**Summary**: Task 9.19 currently has a placeholder implementation that reports an error when lambda parameters lack type annotations. Full contextual type inference would allow the compiler to infer parameter types from the context where the lambda is used.

**Current Status**: Lambda parameter type inference reports "not fully implemented" error. Return type inference from body is complete.

**Tasks for Full Implementation** (5 tasks):

- [ ] 9.19 Add type context passing infrastructure to expression analyzer:
  - [ ] Modify `analyzeExpression()` to accept optional `expectedType` parameter
  - [ ] Thread expected type through all expression analysis calls
  - [ ] Maintain backward compatibility (default to nil for existing calls)
  - [ ] Update all expression analyzers to use context when available
- [ ] 9.20 Implement assignment context type inference:
  - [ ] Detect when lambda is assigned to typed variable: `var f: TFunc := lambda(x) => x * 2`
  - [ ] Extract function pointer type from variable declaration
  - [ ] Pass parameter types to lambda analyzer
  - [ ] Apply inferred types to untyped parameters
  - [ ] Validate inferred types match if some params are explicitly typed
- [ ] 9.21 Implement function call context type inference:
  - [ ] Detect when lambda is passed as function argument: `Apply(5, lambda(n) => n * 2)`
  - [ ] Extract expected function pointer type from function parameter
  - [ ] Pass parameter types to lambda analyzer
  - [ ] Apply inferred types to untyped parameters
  - [ ] Handle overloaded functions (try each signature)
- [ ] 9.22 Implement return statement context type inference:
  - [ ] Detect when lambda is returned from function with known return type
  - [ ] Extract function pointer type from return type
  - [ ] Apply to lambda parameters
- [ ] 9.23 Add comprehensive tests for contextual type inference:
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

- [ ] 9.24 Dynamic array literal syntax for integer arrays:
  - [ ] Cannot use: `var nums := [1, 2, 3, 4, 5];` (parsed as SET, not array)
  - [ ] Current workaround: Use SetLength + manual assignment
  - [ ] Location: Parser interprets `[...]` as set literals (enum values only)
  - [ ] Impact: Cannot easily create test data or initialize arrays
  - [ ] Blocks: testdata/lambdas/higher_order.dws execution
  - [ ] Requires type inference infrastructure from tasks 9.19-9.23
  - [ ] Distinguish array literals from set literals based on expected type context
  - [ ] Update `parseSetLiteral()` to conditionally return `ArrayLiteral` AST node
  - [ ] Add semantic analysis for array literal type checking
  - [ ] Add interpreter support for array literal evaluation

---

### Improved Error Messages and Stack Traces (MEDIUM PRIORITY)

**Summary**: Enhance error reporting with better messages, stack traces, and debugging information. Improves developer experience significantly.

**Reference**: docs/missing-features-recommendations.md lines 299-302

#### Stack Trace Infrastructure (3 tasks) ✅ COMPLETE

- [x] 9.107 Create `errors/stack_trace.go`: ✅ COMPLETE
  - [x] Define `StackFrame` struct with `FunctionName`, `FileName`, `Position` (line/column)
  - [x] Define `StackTrace` type as `[]StackFrame`
  - [x] Implement `String()` method for formatted output (DWScript-compatible format)
  - [x] Add helper methods: `Top()`, `Bottom()`, `Depth()`, `Reverse()`
- [x] 9.108 Implement stack trace capture in interpreter: ✅ COMPLETE
  - [x] Track call stack during execution (updated `Interpreter.callStack` to use `StackTrace`)
  - [x] Push frame on function entry (via `pushCallStack()` helper)
  - [x] Pop frame on function exit (via `popCallStack()` helper)
  - [x] Capture stack on error/exception (updated `ExceptionValue.CallStack` to use `StackTrace`)
  - [x] Include position information from current AST node
  - [x] Updated all call sites: user functions, lambdas, record methods, class methods
- [x] 9.109 Add tests for stack trace capture: ✅ COMPLETE
  - [x] Comprehensive unit tests in `internal/errors/stack_trace_test.go`
  - [x] Updated FFI error tests to use new `StackTrace` type
  - [x] Verified exception CLI tests pass with new format

#### Error Message Improvements (3 tasks)

- [ ] 9.110 Improve type error messages:
  - [ ] Before: "Type mismatch"
  - [ ] After: "Cannot assign Float to Integer variable 'count' at line 42"
  - [ ] Include expected and actual types
  - [ ] Include variable name and location
- [ ] 9.111 Improve runtime error messages:
  - [ ] Include expression that failed
  - [ ] Show values involved: "Division by zero: 10 / 0 at line 15"
  - [ ] Context from surrounding code
- [ ] 9.112 Add source code snippets to errors:
  - [ ] Show the line that caused error
  - [ ] Highlight error position with `^` or color
  - [ ] Show 1-2 lines of context

#### Exception Enhancements (2 tasks)

- [ ] 9.113 Add stack trace to exception objects:
  - [ ] Store `StackTrace` in exception
  - [ ] Display on uncaught exception
  - [ ] Format nicely for CLI output
- [ ] 9.114 Implement `GetStackTrace()` built-in:
  - [ ] Return current stack trace as string
  - [ ] Useful for logging and debugging

#### Debugging Information (2 tasks)

- [ ] 9.115 Add source position to all AST nodes:
  - [ ] Audit nodes for missing `Pos` fields
  - [ ] Add `EndPos` for better range reporting
  - [ ] Use in error messages
- [ ] 9.116 Implement call stack inspection:
  - [ ] `GetCallStack()` returns array of frame info
  - [ ] Each frame: function name, file, line
  - [ ] Accessible from DWScript code

#### Testing & Documentation (2 tasks)

- [ ] 9.117 Create test fixtures demonstrating error messages:
  - [ ] Type errors with clear messages
  - [ ] Runtime errors with stack traces
  - [ ] Exception handling with stack traces
  - [ ] Compare before/after error message quality
- [ ] 9.118 Document error message format in `docs/error-messages.md`

---

### Contracts (Design by Contract) ✅ 94.7% COMPLETE (36/38 tasks, 2 deferred to Stage 7)

**Overview**: Implement DWScript's complete contract system with preconditions (`require`), postconditions (`ensure`), `old` keyword for referencing pre-execution values, and proper OOP inheritance semantics following Liskov substitution principle.

**Status**: Core contracts implementation COMPLETE. All parsing, semantic analysis, runtime execution, and testing complete. Only OOP method inheritance deferred to Stage 7.

**Reference**:

- <https://www.delphitools.info/2011/01/19/leaps-and-bounds-of-dwscript/>
- `reference/dwscript_original/dwsSymbols.pas` (lines 823-871)
- `reference/dwscript_original/dwsExprs.pas` (lines 1210-1254, 3218-3254)
- `reference/dwscript_original/dwsCompiler.pas` (lines 4061-4265)

#### Phase 1: AST Nodes (9 tasks) ✅ COMPLETE

- [x] 9.119 Create `Condition` struct in `ast/statements.go`
  - Fields: `Test Expression` (must be boolean), `Message Expression` (optional string)
  - Implements `Node`, `Statement` interfaces
  - Add `String()` method for debugging
- [x] 9.120 Create `PreConditions` collection node
  - Fields: `Conditions []Condition`, `Token token.Token`
  - Implements `Node`, `Statement` interfaces
  - Add `String()` method showing all conditions
- [x] 9.121 Create `PostConditions` collection node
  - Same structure as `PreConditions`
  - Separate type for semantic distinction
- [x] 9.122 Add contract fields to `FunctionDeclaration` in `ast/functions.go`
  - Add `PreConditions *PreConditions` field
  - Add `PostConditions *PostConditions` field
  - Update `String()` method to show contracts
- [x] 9.123 Add contract fields to `MethodDeclaration` (for OOP support)
  - Methods use `FunctionDecl` type, so already have contract support
  - Will enable inheritance semantics in later phase
- [x] 9.124 Create `OldExpression` node in `ast/statements.go`
  - Fields: `Token token.Token` (the OLD token), `Identifier *Identifier`
  - Implements `Node`, `Expression` interfaces
  - Syntax: `old identifier` (no parentheses, matches DWScript reference)
- [x] 9.125 Implement `TokenLiteral()` methods for all new nodes
  - Required for error reporting with proper source positions
- [x] 9.126 Add contract nodes to AST visitor patterns (if implemented)
  - N/A: No visitor pattern currently implemented
- [x] 9.127 Write AST node unit tests
  - Test `String()` output matches expected format
  - Test node construction and field access
  - Created `ast/contracts_test.go` with 10 comprehensive tests

#### Phase 2: Parser (12 tasks) ✅ COMPLETE

- [x] 9.128 Parse `require` keyword in function/method declarations
  - After function signature, before `var`/`const`/`begin`
  - Set up loop to parse multiple conditions
  - Entry point: `parseFunctionDeclaration()` in `parser/functions.go`
- [x] 9.129 Implement `parseCondition()` helper method
  - Parse boolean expression (will validate type in semantic phase)
  - Check for optional `: "string literal"` suffix for custom message
  - Return `Condition` struct
  - Implemented in `parser/expressions.go`
- [x] 9.130 Parse multiple preconditions (semicolon-separated)
  - Loop until non-semicolon token or declaration keyword (`var`, `const`, `begin`)
  - Collect all conditions into `PreConditions` node
  - Handle empty condition list (just `require` keyword) as error
- [x] 9.131 Parse `ensure` keyword after function body
  - After `end` keyword of function block
  - Before next function/procedure/type/implementation/end keyword
  - Use same `parseCondition()` logic
- [x] 9.132 Parse multiple postconditions (same logic as preconditions)
  - Collect into `PostConditions` node
- [x] 9.133 Enable `old` keyword parsing in postcondition context
  - Add parser flag: `parsingPostCondition bool`
  - Set to true when parsing `ensure` block
  - Register prefix parse function for `OLD` token
- [x] 9.134 Implement `parseOldExpression()` method
  - Expect `old identifier` syntax (no parentheses)
  - Parse identifier following `old` keyword
  - Validate only used in postconditions (check `parsingPostCondition` flag)
  - Return `OldExpression` node
- [x] 9.135 Handle condition parsing errors gracefully
  - Missing semicolon between conditions
  - Non-boolean expressions (defer to semantic phase)
  - Unterminated string messages
  - `old` outside postconditions (report error immediately)
- [x] 9.136 Add parser tests for preconditions
  - Single condition, multiple conditions
  - With and without custom messages
  - Error cases (missing expressions, syntax errors)
  - Created `parser/contracts_test.go` with comprehensive tests
- [x] 9.137 Add parser tests for postconditions
  - Same coverage as preconditions
  - Test `old` expressions in various contexts
  - All tests pass
- [x] 9.138 Add parser tests for combined pre/post conditions
  - Functions with both `require` and `ensure`
  - Edge cases: empty bodies, multiple returns, nested functions
  - TestParseCombinedContracts, TestParseContractsWithLocalVars
- [x] 9.139 Test parser error recovery
  - Ensure parser continues after contract errors
  - Verify error messages include proper source positions
  - TestParseOldOutsidePostconditionError validates error handling
  - Created testdata examples: division.dws, increment.dws, clamp.dws, etc.

#### Phase 3: Semantic Analysis (6 tasks) ✅ COMPLETE (5/6 tasks, Task 9.144 deferred)

- [x] 9.140 Validate precondition expressions are boolean type
  - In `semantic/analyze_functions.go`, added `checkPreconditions()` method
  - Type-check each condition's `Test` expression
  - Report error if not boolean: "precondition must be boolean expression"
  - Implemented and tested ✅
- [x] 9.141 Validate postcondition expressions are boolean type
  - Same logic as preconditions in `checkPostconditions()` method
  - Error message: "postcondition must be boolean expression"
  - Implemented and tested ✅
- [x] 9.142 Validate message expressions are string type
  - Check optional `Message` field in each condition
  - Messages are validated as string type
  - Note: Parser enforces STRING literal after colon, semantic check redundant for literals
  - Implemented and tested ✅
- [x] 9.143 Validate `old` keyword usage
  - Parser validates `old` only in postconditions
  - Added `validateOldExpressions()` to recursively check undefined identifiers
  - Added `analyzeOldExpression()` in `analyze_expressions.go` for type inference
  - Error if referencing undefined variable: "old() references undefined identifier"
  - Implemented and tested ✅
- [ ] 9.144 Check method contract inheritance (OOP) - DEFERRED
  - Will implement with full OOP inheritance support in Stage 7
  - Requires class hierarchy analysis and method overriding support
  - Liskov substitution principle to be implemented later
- [x] 9.145 Add semantic analysis tests
  - Created `semantic/contracts_test.go` with 6 comprehensive tests
  - Boolean type validation for preconditions and postconditions ✅
  - String type validation for messages ✅
  - `old` expression validation (defined/undefined identifiers) ✅
  - Multiple conditions testing ✅
  - Type inference testing ✅
  - All tests pass ✅

#### Phase 4: Interpreter/Runtime (6 tasks) ✅ COMPLETE (5/6 - Task 9.151 deferred to Stage 7)

- [x] 9.146 Implement `old` value capture in interpreter ✅
  - In `interp/interpreter.go`, before function execution:
  - Create `oldValues map[string]interface{}` for current call
  - Traverse postconditions to find all `OldExpression` nodes
  - Evaluate and store current values: `oldValues[ident] = env.Get(ident)`
  - **Implementation**: Added `oldValuesStack []map[string]Value` to Interpreter struct
  - **Implementation**: Created `captureOldValues()` and `findOldExpressions()` in contracts.go
  - **Implementation**: Push/pop old values around function body execution
- [x] 9.147 Evaluate preconditions before function body ✅
  - In function call handler, after parameter binding:
  - Loop through `PreConditions.Conditions`
  - Evaluate each `Test` expression
  - If false, raise assertion error (next task)
  - **Implementation**: Added `checkPreconditions()` in contracts.go
  - **Implementation**: Integrated in callUserFunction after parameter binding
- [x] 9.148 Implement contract failure error handling ✅
  - Create `EAssertionFailed` error type in `errors/errors.go`
  - Format: `"Pre-condition failed in FuncName [line:col]: message"`
  - Use custom message if provided, else condition source code
  - Include stack trace for debugging
  - **Implementation**: Created ContractFailureError type in errors.go
  - **Implementation**: Format includes function name, condition type, location, and message
- [x] 9.149 Evaluate postconditions after function body ✅
  - After body execution, before returning:
  - Make `oldValues` map available in evaluation environment
  - Loop through `PostConditions.Conditions`
  - Evaluate each `Test` (can reference `old` values)
  - If false, raise assertion error with: `"Post-condition failed in FuncName [line:col]: message"`
  - **Implementation**: Added `checkPostconditions()` in contracts.go
  - **Implementation**: Integrated in callUserFunction after body execution, before return
- [x] 9.150 Implement `OldExpression` evaluation ✅
  - In expression evaluator, handle `OldExpression`:
  - Look up identifier in `oldValues` map
  - Error if not found: "internal error: old value not captured"
  - **Implementation**: Added case in Interpreter.Eval() for *ast.OldExpression
  - **Implementation**: Added getOldValue() helper method
- [ ] 9.151 Handle method contract inheritance at runtime **DEFERRED TO STAGE 7**
  - For preconditions: Only evaluate base class conditions (weakening)
  - For postconditions: Evaluate conditions from all ancestor classes (strengthening)
  - Walk up method override chain to collect conditions
  - **Reason**: Stage 7 (OOP/Classes) not yet complete; requires class hierarchy support

#### Phase 5: Testing & Integration (5 tasks) ✅ COMPLETE

- [x] 9.152 Port DWScript reference contract tests ✅
  - Ported 3 SimpleScripts tests to testdata/contracts/:
    - `procedure_contracts.dws` - Basic pre/postconditions
    - `contracts_old.dws` - `old` keyword usage
    - `contracts_code.dws` - Precondition validation
  - Created expected `.out` files for each
  - All tests passing ✅
- [x] 9.153 Fix Rosetta Code examples ✅
  - Fixed `examples/rosetta/Assertions.dws` (moved `ensure` after `end`)
  - Fixed `examples/rosetta/Dot_product.dws` (removed Vector dependency)
  - Both examples now run successfully ✅
- [x] 9.154 Create comprehensive contract test suite ✅
  - Added `recursive_factorial.dws` - Contracts with recursive functions
  - Added `nested_calls.dws` - Contracts with nested function calls
  - Added `multiple_conditions.dws` - Multiple pre/postconditions
  - All 9 contract tests passing (6 new + 3 existing) ✅
  - Note: var parameters with `old` deferred (var parameters have pre-existing bug)
  - Note: Method inheritance tests deferred to Stage 7 (requires OOP completion)
- [x] 9.155 Add contract examples to documentation ✅
  - Created `docs/contracts.md` with complete specification:
    - Syntax for `require`, `ensure`, and `old` keyword
    - Multiple examples (simple preconditions, postconditions, recursive, etc.)
    - Error message format documentation
    - Best practices and limitations
  - Updated `README.md`:
    - Added contracts to feature list
    - Added contracts quick example
    - Added link to docs/contracts.md
- [x] 9.156 Update CLI help and PLAN.md completion markers ✅
  - Updated PLAN.md with Phase 5 completion status
  - CLI help for contracts pending (minor enhancement)
  - Update phase 9 statistics
  - Add contracts to feature list in README

---

### Fixture Test Infrastructure & Missing Features (8 tasks)

**Summary**: These tasks were identified by running the comprehensive fixture test suite (`internal/interp/fixture_test.go`) against the DWScript reference implementation's test scripts. The fixture tests currently run 20 conservative tests from the 983-test reference suite (442 SimpleScripts + 541 FailureScripts). Fixing these issues will significantly improve compatibility with real-world DWScript code and enable testing against the full reference test suite.

**Current Status**: 14/20 tests passing (70%). The 6 failures expose missing language features across parser, semantic analysis, and built-in functions.

**Implementation Tasks**:

- [x] 9.157 Add `normalizeOutput()` helper function to `fixture_test.go`: ✅ DONE
  - [x] Quick fix: Function already exists in `interface_reference_test.go` (same package)
  - [x] Normalizes line endings and whitespace for cross-platform test compatibility
  - [x] Required for comparing expected vs actual output from reference tests
  - [x] **Impact**: Unblocks basic test infrastructure (compilation fix)
  - [x] **Note**: No code changes needed - `fixture_test.go` can use the function from `interface_reference_test.go` as they're in the same package

- [x] 9.158 Parser support for optional `program Name;` declarations: ✅ DONE
  - [x] Lexer already recognizes `PROGRAM` token (was already present)
  - [x] Added `parseProgramDeclaration()` in `internal/parser/declarations.go`
  - [x] Integrated into `ParseProgram()` to detect and skip program declarations
  - [x] Syntax: `program MyProgram;` at beginning of file
  - [x] Program declaration is parsed and discarded (doesn't affect execution)
  - [x] Added comprehensive parser tests in `declarations_test.go`
  - [x] **Impact**: Partially fixes `program.pas` and `programDot.pas` test failures
  - [x] **Note**: These fixture tests still fail due to **const block syntax** (separate issue)
    - Files use `const\n  C1 = 1;\n  C2 = 2;` (one `const` keyword, multiple declarations)
    - Current parser expects `const C1 = 1; const C2 = 2;` (repeat `const` keyword)
    - This is a separate parser feature that needs implementation (tracked separately)
  - [x] **Verification**: Program declaration parsing works correctly:
    - `program Test; begin PrintLn('Hello'); end` ✅ Works
    - `const C1 := 1; const C2 := 2;` ✅ Works (separate const keywords)
  - [x] **Reference**: DWScript allows but doesn't require program declarations

- [x] 9.159 Make constructor parentheses optional in `new` expressions:
  - [x] Current: Parser requires `new TTest()`
  - [x] DWScript allows: `new TTest` (no parentheses for zero-argument constructors)
  - [x] Update parser's `parseNewExpression()` to make `()` optional
  - [x] Semantic analysis should treat `new TTest` as `new TTest()`
  - [x] **Impact**: Fixes `empty_body.pas` test failure
  - [x] **Reference**: Common DWScript pattern for simple object creation

- [x] 9.160 Allow zero-argument procedure calls without parentheses:
  - [x] Current: `PrintLn;` treated as identifier, not procedure call
  - [x] DWScript allows: `PrintLn;` equivalent to `PrintLn();`
  - [x] Parser or semantic analyzer needs to detect this pattern
  - [x] Option A: Parser recognizes known procedures and treats `;` as call
  - [x] Option B: Semantic analyzer converts identifier to zero-arg call
  - [x] **Impact**: Fixes `print_multi_args.pas` test failure
  - [x] **Reference**: Classic Pascal feature for procedure calls

- [x] 9.161 Implement type meta-values (type names as runtime values):
  - [x] DWScript allows type names like `Integer` to be used as values
  - [x] Used for reflection and type-based operations
  - [x] Example: `High(Integer)` where `Integer` is a type meta-value
  - [x] Type system needs `TypeMetaValue` or similar runtime representation
  - [x] Parser should allow type names in expression contexts
  - [x] Semantic analysis validates type meta-value usage
  - [x] **Impact**: Required for `mult_by2.pas` test (uses `High(Integer)`)
  - [x] **Reference**: Part of DWScript's reflection system

- [x] 9.162 Implement `High()` and `Low()` built-in functions:
  - [x] `High(TypeName)` returns maximum value for the type
  - [x] `Low(TypeName)` returns minimum value for the type
  - [x] Support ordinal types: Integer, Float, Boolean, Char, enums
  - [x] Examples:
    - `High(Integer)` → `9223372036854775807` (Int64 max)
    - `Low(Integer)` → `-9223372036854775808` (Int64 min)
    - `High(Boolean)` → `True`
    - `Low(TColor)` → first enum value
  - [x] Add to built-in function registry in interpreter
  - [x] **Impact**: Fixes `mult_by2.pas` test failure
  - [x] **Reference**: Standard DWScript/Pascal built-ins for type bounds

- [x] 9.163 Add implicit integer→float coercion in built-in function calls:
  - [x] Current: `FloatToStr(i+1)` fails when `i` is Integer
  - [x] DWScript automatically coerces Integer to Float in mixed contexts
  - [x] Option A: Make `FloatToStr()` accept both Integer and Float arguments
  - [x] Option B: Add implicit coercion in semantic analysis for function calls
  - [x] Option B is more general and matches DWScript behavior
  - [x] Update semantic analyzer's function call type checking
  - [x] **Impact**: Fixes `int_float.pas` test failure
  - [x] **Reference**: DWScript's flexible type coercion rules

- [ ] 9.164 Parser/semantic support for `inherited` keyword:
  - [ ] Used in class methods to call parent class implementations
  - [ ] Syntax: `inherited;` or `inherited MethodName(args);`
  - [ ] Parser needs to recognize `inherited` as expression prefix
  - [ ] Semantic analysis validates `inherited` usage (only in class methods)
  - [ ] Interpreter resolves to parent class method and executes
  - [ ] **Impact**: Fixes `empty_body.pas` test failure
  - [ ] **Reference**: Core OOP feature from Stage 7 (already partially implemented)
  - [ ] **Note**: This may already be partially implemented; check Stage 7 OOP code

**Additional Issue Discovered**: Several fixture tests (`program.pas`, `programDot.pas`) use **const block syntax** which is not yet implemented:
- DWScript allows: `const\n  C1 = 1;\n  C2 = 2;` (one `const` keyword, multiple declarations in a block)
- Current parser requires: `const C1 = 1; const C2 = 2;` (repeat `const` for each declaration)
- This is a separate parser feature that needs implementation (not covered by tasks 9.129-9.136)
- Similar block syntax likely needed for `var` declarations as well
- **Recommendation**: Add new task for "Declaration Block Syntax" to handle `const`/`var` blocks

**Test Expansion Opportunity**: Once these 8 issues are fixed, expand the `workingTests` map in `filterTestFiles()` to include more tests from the reference suite. The current 20-test whitelist is very conservative; many more tests likely work with these fixes.

**Reference Test Suite Structure**:
- Location: `reference/dwscript-original/Test/`
- SimpleScripts: 442 tests (basic language features)
- FailureScripts: 541 tests (error handling and edge cases)
- Each test: `.pas` source file + `.txt` expected output file

---

### Comprehensive Testing (Stage 8)

- [ ] 9.165 Port DWScript's test suite (if available)
- [ ] 9.166 Run DWScript example scripts from documentation
- [ ] 9.167 Compare outputs with original DWScript
- [ ] 9.168 Fix any discrepancies
- [ ] 9.169 Create stress tests for complex features
- [ ] 9.170 Achieve >85% overall code coverage
- [x] 9.171 Fix 'step' contextual keyword conflict - allow 'step' as variable name outside for loops

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
