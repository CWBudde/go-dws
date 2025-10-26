# Stage 11: Code Generation - Multi-Backend Architecture

**Status**: Not started | **Estimated Total Tasks**: ~180

## Overview

This stage introduces code generation capabilities to go-dws using a **two-tier architecture**:

1. **MIR (Mid-level IR)**: A target-neutral intermediate representation that sits between the type-checked AST and backend-specific code generators
2. **Backend Emitters**: Pluggable code generators that translate MIR to specific targets (JavaScript, LLVM IR)

### Why MIR?

The MIR layer provides:
- **Clean separation**: AST/semantic analysis remains independent of code generation
- **Multi-backend support**: Write AST→MIR once, support multiple targets (JS, LLVM, WebAssembly, etc.)
- **Optimization opportunities**: MIR is designed for transformations (constant folding, dead code elimination)
- **Easier debugging**: Human-readable MIR dumps help diagnose codegen issues
- **Future-proofing**: Adding new backends is much simpler with a stable MIR spec

### Architecture Flow

```
DWScript Source
      ↓
   Lexer (tokens)
      ↓
   Parser (AST)
      ↓
   Semantic Analyzer (typed AST)
      ↓
   MIR Builder (MIR)          ← Stage 11.1
          ↓
   ┌──────┴──────┐
   ↓             ↓
JS Emitter    LLVM Emitter    ← Stage 11.2-11.4
   ↓             ↓
JavaScript    Native Code
```

### Implementation Phases

**Stage 11.1**: MIR Foundation (30 tasks) - ~2 weeks
- Define MIR type system and instruction set
- Implement MIR builder and verifier
- AST→MIR lowering for basic constructs

**Stage 11.2**: JS Backend MVP (45 tasks) - ~3 weeks
- Basic JS emitter with readable output
- Support for expressions, control flow, functions
- End-to-end tests: DWScript→JS execution

**Stage 11.3**: JS Feature Complete (60 tasks) - ~4 weeks
- Classes, inheritance, interfaces
- Records, arrays, sets, enums
- Exception handling, properties
- Full DWScript language parity

**Stage 11.4**: LLVM Backend [OPTIONAL] (45 tasks) - ~4 weeks
- LLVM IR emission for basic constructs
- Runtime library (memory, strings, arrays)
- Native compilation pipeline

---

## Stage 11.1: MIR Foundation

**Goal**: Define a complete, verifiable mid-level IR that can represent all DWScript constructs in a target-neutral way.

**Exit Criteria**:
- ✅ MIR spec documented with examples
- ✅ Complete MIR type system
- ✅ MIR builder API that prevents invalid IR construction
- ✅ MIR verifier that catches malformed IR
- ✅ AST→MIR lowering for ~80% of DWScript constructs
- ✅ Golden test suite: 20+ DWScript programs with .mir dumps
- ✅ 85%+ code coverage for mir/ package

### 11.1.1: MIR Package Structure and Types (10 tasks)

- [ ] Create `mir/` package directory
- [ ] Create `mir/types.go` - MIR type system
  - [ ] Define `Type` interface with `String()`, `Size()`, `Align()` methods
  - [ ] Implement primitive types: `Bool`, `Int8`, `Int16`, `Int32`, `Int64`, `Float32`, `Float64`, `String`
  - [ ] Implement composite types: `Array(elemType, size)`, `Record(fields)`, `Pointer(pointeeType)`
  - [ ] Implement OOP types: `Class(name, fields, methods, parent)`, `Interface(name, methods)`
  - [ ] Implement function types: `Function(params, returnType)`
  - [ ] Add `Void` type for procedures
  - [ ] Type equality and compatibility checking
  - [ ] Type conversion rules (explicit vs implicit)

**Example MIR Type representation**:
```
// DWScript: var x: Integer;
MIR Type: Int32

// DWScript: var arr: array[0..9] of Float;
MIR Type: Array(Float64, 10)

// DWScript: type TPoint = record x, y: Float; end;
MIR Type: Record{
  fields: [
    {name: "x", type: Float64, offset: 0},
    {name: "y", type: Float64, offset: 8}
  ]
}
```

- [ ] Create `mir/value.go` - Value identification system
  - [ ] Define `ValueID` type (unique identifier for SSA values)
  - [ ] Define `BlockID` type (unique identifier for basic blocks)
  - [ ] Value naming strategy (e.g., `%0`, `%tmp1`, `%x_2`)

### 11.1.2: MIR Instructions and Control Flow (10 tasks)

- [ ] Create `mir/instruction.go` - MIR instruction set
  - [ ] Define `Instruction` interface with `ID()`, `Type()`, `String()` methods
  - [ ] **Arithmetic ops**: `Add`, `Sub`, `Mul`, `Div`, `Mod`, `Neg`
  - [ ] **Comparison ops**: `Eq`, `Ne`, `Lt`, `Le`, `Gt`, `Ge`
  - [ ] **Logical ops**: `And`, `Or`, `Xor`, `Not`
  - [ ] **Memory ops**: `Alloca` (stack allocation), `Load`, `Store`
  - [ ] **Constants**: `ConstInt`, `ConstFloat`, `ConstString`, `ConstBool`, `ConstNil`
  - [ ] **Conversions**: `IntToFloat`, `FloatToInt`, `IntTrunc`, `IntExt`
  - [ ] **Function ops**: `Call`, `VirtualCall` (for method dispatch)
  - [ ] **Array ops**: `ArrayAlloc`, `ArrayLen`, `ArrayIndex`, `ArraySet`
  - [ ] **Record/Class ops**: `FieldGet`, `FieldSet`, `New` (allocation)
  - [ ] **Control flow**: `Phi` (SSA join), `Br`, `CondBr`, `Return`, `Throw`

**Example MIR Instructions**:
```
// DWScript: var x := 5 + 3;
%0 = ConstInt 5
%1 = ConstInt 3
%2 = Add %0, %1
Store %x, %2

// DWScript: if x > 10 then ...
%0 = Load %x
%1 = ConstInt 10
%2 = Gt %0, %1
CondBr %2, label_then, label_else
```

- [ ] Create `mir/block.go` - Basic blocks
  - [ ] Define `Block` struct with `ID`, `Instructions`, `Terminator`
  - [ ] Terminator validation (every block must end with Br/CondBr/Return/Throw)
  - [ ] Block predecessors/successors tracking for CFG

- [ ] Create `mir/function.go` - Function representation
  - [ ] Define `Function` struct: `Name`, `Params`, `ReturnType`, `Blocks`, `Locals`
  - [ ] Entry block convention (first block is always entry)
  - [ ] Parameter and local variable tracking

### 11.1.3: MIR Builder API (5 tasks)

- [ ] Create `mir/builder.go` - Safe MIR construction
  - [ ] Define `Builder` struct with current function/block context
  - [ ] `NewBuilder()` constructor
  - [ ] `NewFunction(name, params, retType) *Function` - start new function
  - [ ] `NewBlock(name) *Block` - create new basic block
  - [ ] `SetInsertPoint(block)` - set where instructions are added
  - [ ] Instruction emission methods: `EmitAdd(lhs, rhs)`, `EmitLoad(var)`, etc.
  - [ ] Type checking during emission (prevent type mismatches)
  - [ ] Automatic ValueID generation
  - [ ] Symbol table for named variables → ValueID mapping

**Example Builder Usage**:
```go
// Build MIR for: function add(a, b: Integer): Integer; begin Result := a + b; end;
b := mir.NewBuilder()
fn := b.NewFunction("add",
  []mir.Param{{Name: "a", Type: mir.Int32}, {Name: "b", Type: mir.Int32}},
  mir.Int32)

entry := b.NewBlock("entry")
b.SetInsertPoint(entry)

aVal := b.EmitLoad("a")      // %0 = Load %a
bVal := b.EmitLoad("b")      // %1 = Load %b
sum := b.EmitAdd(aVal, bVal) // %2 = Add %0, %1
b.EmitReturn(sum)            // Return %2
```

### 11.1.4: MIR Verifier (3 tasks)

- [ ] Create `mir/verifier.go` - MIR correctness checking
  - [ ] CFG verification: all blocks reachable, no orphans, every path leads to terminator
  - [ ] Type verification: instruction operands match expected types
  - [ ] SSA property verification: values defined before use, Phi nodes have correct predecessors
  - [ ] Function signature verification: returns match declared type
  - [ ] `Verify(fn *Function) []error` API returning list of violations

### 11.1.5: AST → MIR Lowering (12 tasks)

- [ ] Create `mir/lower.go` - AST to MIR translation
  - [ ] `LowerProgram(ast *ast.Program) (*mir.Module, error)` entry point
  - [ ] Lower expressions:
    - [ ] Literals → `Const*` instructions
    - [ ] Binary operations → corresponding MIR ops (handle short-circuit for `and`/`or`)
    - [ ] Unary operations → `Neg`, `Not`
    - [ ] Identifier references → `Load` instructions
    - [ ] Function calls → `Call` instructions
    - [ ] Array indexing → `ArrayIndex` + bounds check insertion
    - [ ] Record field access → `FieldGet`/`FieldSet`
  - [ ] Lower statements:
    - [ ] Variable declarations → `Alloca` + optional `Store` for initializer
    - [ ] Assignments → `Store` instruction
    - [ ] If statements → `CondBr` with then/else/merge blocks
    - [ ] While loops → loop header/body/exit blocks with `Br`/`CondBr`
    - [ ] For loops → desugared to while-style loop with initialization and increment
    - [ ] Return statements → `Return` instruction
  - [ ] Lower declarations:
    - [ ] Functions/procedures → `Function` with parameter lowering
    - [ ] Records → `Record` type definition
    - [ ] Classes → `Class` type definition (fields, methods, vtable prep)

**Example Lowering**:
```
DWScript:
  if x > 5 then
    y := 10
  else
    y := 20;

MIR:
  entry:
    %0 = Load %x
    %1 = ConstInt 5
    %2 = Gt %0, %1
    CondBr %2, then_block, else_block

  then_block:
    %3 = ConstInt 10
    Store %y, %3
    Br merge_block

  else_block:
    %4 = ConstInt 20
    Store %y, %4
    Br merge_block

  merge_block:
    ; continue...
```

- [ ] Short-circuit evaluation for boolean operators
  - [ ] `and` → conditional branch (if left is false, skip right)
  - [ ] `or` → conditional branch (if left is true, skip right)

- [ ] Implement simple optimizations during lowering:
  - [ ] Constant folding (e.g., `3 + 5` → `ConstInt 8`)
  - [ ] Dead code elimination (unreachable blocks after `Return`)

### 11.1.6: MIR Dump and Debugging (2 tasks)

- [ ] Create `mir/dump.go` - Human-readable MIR output
  - [ ] `Dump(fn *Function) string` - pretty-print MIR
  - [ ] Format with indentation, block labels, type annotations
  - [ ] Example output:
    ```
    function add(a: Int32, b: Int32): Int32 {
      entry:
        %0 = Load %a               ; Int32
        %1 = Load %b               ; Int32
        %2 = Add %0, %1            ; Int32
        Return %2
    }
    ```

- [ ] Integration with CLI: `./bin/dwscript dump-mir script.dws`

### 11.1.7: MIR Testing (3 tasks)

- [ ] Create golden MIR tests in `mir/testdata/`
  - [ ] 5+ expression tests (arithmetic, boolean, comparisons)
  - [ ] 5+ control flow tests (if, while, for)
  - [ ] 5+ function tests (calls, recursion, parameters)
  - [ ] 5+ advanced tests (records, arrays, classes)

- [ ] MIR verifier tests:
  - [ ] Test detection of type mismatches
  - [ ] Test detection of malformed CFG
  - [ ] Test detection of SSA violations

- [ ] Round-trip tests: AST → MIR → verify → dump → compare with golden files

---

## Stage 11.2: JS Backend MVP

**Goal**: Implement a JavaScript code generator that can compile basic DWScript programs (expressions, control flow, functions) to readable, runnable JavaScript.

**Exit Criteria**:
- ✅ JS emitter can handle arithmetic, boolean logic, comparisons
- ✅ Control flow: if/else, while, for loops work correctly
- ✅ Functions with parameters and return values
- ✅ Variable declarations and assignments
- ✅ 20+ end-to-end tests: DWScript → JS → execute with Node.js → verify output
- ✅ Golden JS snapshot tests (check in .js files for regression detection)
- ✅ 85%+ coverage for codegen/js/ package

### 11.2.1: JS Emitter Infrastructure (8 tasks)

- [ ] Create `codegen/` package for shared codegen interfaces
  - [ ] Define `Backend` interface: `Generate(mir *mir.Module) (string, error)`
  - [ ] Define `EmitterOptions` struct: `Indent string`, `Pretty bool`, `SourceComments bool`

- [ ] Create `codegen/js/` package
  - [ ] Create `codegen/js/emitter.go`
  - [ ] Define `JSEmitter` struct:
    ```go
    type JSEmitter struct {
      out     strings.Builder
      indent  int
      opts    EmitterOptions
      tmpCounter int  // for temporary variable naming
    }
    ```
  - [ ] Helper methods:
    - [ ] `emit(s string)` - append to output
    - [ ] `emitLine(s string)` - append with newline
    - [ ] `emitIndent()` - emit current indentation
    - [ ] `pushIndent()` / `popIndent()` - manage indent level
    - [ ] `newTemp() string` - generate unique temp variable names (`_tmp0`, `_tmp1`, ...)
  - [ ] `NewJSEmitter(opts EmitterOptions) *JSEmitter` constructor
  - [ ] `Generate(module *mir.Module) (string, error)` - main entry point

### 11.2.2: Module and Function Emission (6 tasks)

- [ ] Module structure emission:
  - [ ] Choose ES Module format (default): `export` for public functions
  - [ ] Optional IIFE fallback for older environments (via `EmitterOptions`)
  - [ ] Emit file header comment:
    ```js
    // Code generated by go-dws; DO NOT EDIT.
    // Source: <filename>
    ```

- [ ] Function emission:
  - [ ] Emit function declarations: `function fname(param1, param2) { ... }`
  - [ ] Map DWScript function params to JS params (preserve names)
  - [ ] Emit local variable declarations at function top (collect from `Alloca` instructions):
    ```js
    function foo(x) {
      let y, z, _tmp0, _tmp1;  // all locals declared upfront
      // ... function body ...
    }
    ```
  - [ ] Return statement emission: `return <expr>;`

- [ ] Handle procedures (no return value):
  - [ ] Emit as JS function with implicit `return undefined;`

### 11.2.3: Expression and Instruction Lowering (12 tasks)

- [ ] Arithmetic operations → JS infix operators:
  - [ ] `Add` → `+`, `Sub` → `-`, `Mul` → `*`, `Div` → `/`, `Mod` → `%`
  - [ ] `Neg` → unary `-`

- [ ] Comparison operations → JS comparisons:
  - [ ] `Eq` → `===`, `Ne` → `!==`
  - [ ] `Lt` → `<`, `Le` → `<=`, `Gt` → `>`, `Ge` → `>=`

- [ ] Logical operations → JS boolean ops:
  - [ ] `And` → `&&`, `Or` → `||`, `Not` → `!`

- [ ] Constants → JS literals:
  - [ ] `ConstInt` → `42`, `ConstFloat` → `3.14`
  - [ ] `ConstString` → `"hello"` (with proper escaping)
  - [ ] `ConstBool` → `true`/`false`
  - [ ] `ConstNil` → `null`

- [ ] Variable operations:
  - [ ] `Load` → reference variable name
  - [ ] `Store` → assignment: `varName = value;`
  - [ ] `Alloca` → local variable (collected and declared at function top)

- [ ] Function calls:
  - [ ] `Call` → `functionName(arg1, arg2)`
  - [ ] Handle return value assignment if used

- [ ] Phi nodes:
  - [ ] Strategy: introduce temporary variables and assign at block edges
  - [ ] Example:
    ```
    MIR:
      merge:
        %0 = Phi [%1, then_block], [%2, else_block]

    JS:
      // In then_block, before jump to merge:
      _phi0 = <value from then>;
      // In else_block, before jump to merge:
      _phi0 = <value from else>;
      // In merge:
      let x = _phi0;
    ```

### 11.2.4: Control Flow Emission (8 tasks)

- [ ] Strategy: Reconstruct structured control flow from MIR CFG
  - [ ] Detect if/else patterns from `CondBr` with two successors
  - [ ] Detect while loop patterns (backedge to header block)
  - [ ] Fall back to labeled blocks + `goto` emulation if needed (rare)

- [ ] If-else emission:
  - [ ] `CondBr` → `if (condition) { ... } else { ... }`
  - [ ] Merge blocks handled automatically by sequential code

- [ ] While loop emission:
  - [ ] Detect loop header block (has backedge from within loop)
  - [ ] Emit: `while (condition) { ... }`

- [ ] For loop emission (if MIR preserves for-loop metadata):
  - [ ] Emit: `for (let i = start; i <= end; i++) { ... }`
  - [ ] Otherwise, emit as while loop

- [ ] Unconditional branch:
  - [ ] `Br` → usually implicit (next block), or use label/goto if needed

- [ ] Return:
  - [ ] `Return` → `return <value>;`

### 11.2.5: Minimal Runtime Support (3 tasks)

- [ ] Create `runtime/js/runtime.js` - JavaScript runtime helpers
  - [ ] `_dws.boundsCheck(index, length)` - array bounds checking (optional, enabled by flag)
  - [ ] `_dws.assert(condition, message)` - runtime assertions
  - [ ] Export as ES module or inline in generated code

- [ ] Emit runtime import in generated JS (if needed):
  ```js
  import * as _dws from './runtime.js';
  ```

- [ ] Make runtime usage optional via `EmitterOptions.InsertBoundsChecks`

### 11.2.6: Testing Infrastructure (8 tasks)

- [ ] Create `codegen/js/testdata/` for test cases
  - [ ] Subdirectories: `expr/`, `control_flow/`, `functions/`

- [ ] Golden JS snapshot tests:
  - [ ] For each `.dws` file in testdata, generate `.js` output
  - [ ] Check in `.js` files to repo
  - [ ] Test compares generated JS to golden file (detect regressions)

- [ ] Execution tests (DWScript → JS → run):
  - [ ] Setup: Add Node.js to CI (GitHub Actions)
  - [ ] Test helper: `runJS(jsCode string) (stdout, stderr, error)`
  - [ ] Write tests that:
    1. Parse DWScript
    2. Lower to MIR
    3. Generate JS
    4. Execute with Node.js
    5. Verify stdout matches expected output

- [ ] End-to-end test examples:
  - [ ] Arithmetic: `var x := 5 + 3 * 2; PrintLn(x);` → verify output `11`
  - [ ] Control flow: `if true then PrintLn('yes') else PrintLn('no');` → verify output `yes`
  - [ ] Functions: `function add(a, b: Integer): Integer; begin Result := a + b; end; PrintLn(add(5, 7));` → verify output `12`
  - [ ] Loops: `var i: Integer; for i := 1 to 5 do PrintLn(i);` → verify output `1\n2\n3\n4\n5`

- [ ] Unit tests for JS emitter:
  - [ ] Test each instruction type emits correct JS
  - [ ] Test indentation logic
  - [ ] Test temp variable generation

- [ ] Coverage target: 85%+ for `codegen/js/` package

### 11.2.7: CLI Integration (2 tasks)

- [ ] Add `compile-js` command to `cmd/dwscript/`:
  - [ ] `./bin/dwscript compile-js input.dws -o output.js`
  - [ ] Options: `--pretty`, `--no-bounds-checks`, `--module-format=esm|iife`

- [ ] Update `--help` and documentation

---

## Stage 11.3: JS Feature Complete

**Goal**: Extend JS backend to support all DWScript language features: classes, inheritance, interfaces, records, arrays, sets, enums, exceptions, properties.

**Exit Criteria**:
- ✅ Full OOP support: classes, inheritance, virtual methods, interfaces
- ✅ Composite types: records, static/dynamic arrays, sets, enums
- ✅ Exception handling: try/except/finally, raise, exception objects
- ✅ Properties: read/write/indexed properties
- ✅ Advanced features: operator overloading, generics (via monomorphization)
- ✅ 50+ comprehensive end-to-end tests covering all language features
- ✅ Real-world sample programs compile and run correctly
- ✅ Documentation: "DWScript → JavaScript Mapping Guide"

### 11.3.1: Records (7 tasks)

- [ ] MIR support for records (already in 11.1 if not done):
  - [ ] `Record` type with field definitions
  - [ ] `New` instruction to allocate record
  - [ ] `FieldGet`/`FieldSet` instructions

- [ ] JS emission for records:
  - [ ] Emit as plain JavaScript objects: `{ x: 0, y: 0 }`
  - [ ] Constructor function or object literal pattern:
    ```js
    function TPoint_new() {
      return { x: 0, y: 0 };
    }
    ```
  - [ ] Field access → property access: `point.x`
  - [ ] Field assignment → property assignment: `point.x = 5;`

- [ ] Record copy semantics:
  - [ ] DWScript records are value types → emit shallow copy in JS
  - [ ] Helper: `_dws.copyRecord(record)` returns `{ ...record }`

- [ ] Tests:
  - [ ] Record creation and initialization
  - [ ] Field read/write
  - [ ] Record assignment (copy semantics)
  - [ ] Nested records

### 11.3.2: Arrays (10 tasks)

- [ ] MIR support for arrays:
  - [ ] Static arrays: `Array(elemType, fixedSize)`
  - [ ] Dynamic arrays: `DynArray(elemType)`
  - [ ] `ArrayAlloc`, `ArrayLen`, `ArrayIndex`, `ArraySet` instructions

- [ ] JS emission for static arrays:
  - [ ] Emit as JS array with fixed size: `new Array(10).fill(0)`
  - [ ] Index access: `arr[i]`
  - [ ] Bounds checking (optional): `_dws.boundsCheck(i, arr.length)`

- [ ] JS emission for dynamic arrays:
  - [ ] Emit as JS array: `[]`
  - [ ] `SetLength` → `arr.length = newLen`
  - [ ] `Length` → `arr.length`

- [ ] Multi-dimensional arrays:
  - [ ] Nested arrays in JS: `[[0, 0], [0, 0]]`

- [ ] Array operations:
  - [ ] Copy: `arr.slice()`
  - [ ] Concatenation (if supported in DWScript): `arr1.concat(arr2)`

- [ ] Tests:
  - [ ] Static array creation and indexing
  - [ ] Dynamic array operations (SetLength, append)
  - [ ] Multi-dimensional arrays
  - [ ] Array bounds checking (if enabled)

### 11.3.3: Classes and Inheritance (15 tasks)

- [ ] MIR support for classes (expand from 11.1):
  - [ ] `Class` type with fields, methods, parent class, vtable
  - [ ] `New` instruction for object allocation
  - [ ] `VirtualCall` instruction for method dispatch
  - [ ] Method resolution (walk inheritance chain)

- [ ] JS emission for classes:
  - [ ] Emit ES6 class syntax:
    ```js
    class TAnimal {
      constructor() {
        this.name = "";
      }
      speak() {
        console.log("...");
      }
    }
    ```
  - [ ] Field initialization in constructor
  - [ ] Method emission

- [ ] Inheritance:
  - [ ] Emit `extends` clause: `class TDog extends TAnimal { ... }`
  - [ ] `super()` call in constructor
  - [ ] Virtual method dispatch → JS methods (naturally virtual)

- [ ] Constructor handling:
  - [ ] DWScript `Create` → JS `constructor`
  - [ ] Multiple constructors → emit overload dispatch in single constructor

- [ ] Destructor handling:
  - [ ] DWScript `Destroy` → no direct equivalent in JS (GC handles cleanup)
  - [ ] Optional: emit `destroy()` method and document manual call requirement

- [ ] Static fields and methods:
  - [ ] Emit as class static members: `static fieldName = value;`

- [ ] `Self` and `inherited`:
  - [ ] `Self` → `this`
  - [ ] `inherited` → `super.methodName()`

- [ ] Tests:
  - [ ] Simple class with fields and methods
  - [ ] Inheritance (single, multiple levels)
  - [ ] Virtual method overriding
  - [ ] Constructor variants
  - [ ] Static members
  - [ ] `Self` and `inherited` usage

### 11.3.4: Interfaces (6 tasks)

- [ ] MIR support for interfaces:
  - [ ] `Interface` type with method signatures
  - [ ] Interface implementation tracking on classes

- [ ] JS emission strategy (choose one):
  - [ ] **Option A**: Structural typing (duck typing) - no explicit interface in JS, rely on method presence
  - [ ] **Option B**: Runtime metadata - emit interface tables, explicit `implements` checks
  - [ ] Decision: Document chosen approach in `docs/codegen-interfaces.md`

- [ ] If using Option B:
  - [ ] Emit interface metadata: `TAnimal._interfaces = ['ISpeakable'];`
  - [ ] `is` operator → `_dws.implementsInterface(obj, 'ISpeakable')`
  - [ ] `as` operator → check + cast or throw

- [ ] Tests:
  - [ ] Class implementing interface
  - [ ] Interface method calls
  - [ ] `is` and `as` with interfaces

### 11.3.5: Enums and Sets (8 tasks)

- [ ] MIR support for enums:
  - [ ] `Enum` type with named values and ordinal integers
  - [ ] `ConstEnum` instruction

- [ ] JS emission for enums:
  - [ ] Emit as frozen object:
    ```js
    const TColor = Object.freeze({
      Red: 0,
      Green: 1,
      Blue: 2
    });
    ```
  - [ ] Scoped access: `TColor.Red`
  - [ ] Unscoped access: emit const for each value

- [ ] MIR support for sets:
  - [ ] `Set` type (set of enum or integer range)
  - [ ] Set operations: union, intersection, difference, inclusion

- [ ] JS emission for sets:
  - [ ] **Small sets** (≤32 elements): bitmask as number
    - [ ] `[Red, Blue]` → `0b101` (bits 0 and 2 set)
    - [ ] Union: `a | b`, Intersection: `a & b`, Difference: `a & ~b`
    - [ ] Inclusion: `(a & b) === b`
  - [ ] **Large sets**: JS `Set` object
    - [ ] Union: `new Set([...a, ...b])`
    - [ ] Intersection: `new Set([...a].filter(x => b.has(x)))`

- [ ] Tests:
  - [ ] Enum declaration and usage
  - [ ] Set operations (union, intersection, in operator)

### 11.3.6: Exception Handling (8 tasks)

- [ ] MIR support for exceptions (expand from 11.1):
  - [ ] `Throw` instruction
  - [ ] `Try` block with associated `Catch` and `Finally` blocks
  - [ ] Exception object type

- [ ] JS emission for exceptions:
  - [ ] `Throw` → `throw new Error(message);` or custom exception class
  - [ ] Try-except-finally → `try { ... } catch (e) { ... } finally { ... }`

- [ ] Exception objects:
  - [ ] DWScript `Exception` class → JS `Error` subclass:
    ```js
    class DWSException extends Error {
      constructor(message) {
        super(message);
        this.name = 'DWSException';
      }
    }
    ```

- [ ] `On E: ExceptionType do` → `catch (e) { if (e instanceof ExceptionType) { ... } else { throw e; } }`

- [ ] Re-raise: `raise;` → `throw e;` (must track current exception)

- [ ] Tests:
  - [ ] Basic try-except
  - [ ] Multiple except handlers (by type)
  - [ ] Try-finally
  - [ ] Try-except-finally
  - [ ] Re-raise
  - [ ] Nested exception handling

### 11.3.7: Properties (6 tasks)

- [ ] MIR support for properties:
  - [ ] Property metadata: read method, write method, or direct field
  - [ ] `PropGet`, `PropSet` instructions

- [ ] JS emission for properties:
  - [ ] **Option A**: ES6 getters/setters:
    ```js
    get propertyName() { return this._propertyName; }
    set propertyName(value) { this._propertyName = value; }
    ```
  - [ ] **Option B**: Explicit methods:
    ```js
    getPropertyName() { return this._propertyName; }
    setPropertyName(value) { this._propertyName = value; }
    ```
  - [ ] Decision: Use Option A for cleaner syntax

- [ ] Indexed properties:
  - [ ] Read: emit as method `getItem(index)`
  - [ ] Write: emit as method `setItem(index, value)`

- [ ] Tests:
  - [ ] Read/write properties
  - [ ] Indexed properties
  - [ ] Property overriding in inheritance

### 11.3.8: Advanced Features (10 tasks)

- [ ] Operator overloading:
  - [ ] Strategy: desugar to method calls in MIR lowering phase
  - [ ] Example: `a + b` where `+` overloaded → `a.opAdd(b)`
  - [ ] JS emission: call the method

- [ ] Generics:
  - [ ] Strategy: monomorphization during semantic analysis
  - [ ] MIR sees only concrete instantiated types
  - [ ] JS emits separate functions/classes for each instantiation

- [ ] With statement:
  - [ ] Strategy: desugar in AST→MIR lowering
  - [ ] Replace `with obj do ...` → explicit `obj.field` references

- [ ] String operations:
  - [ ] Concatenation: `+` → JS `+`
  - [ ] Length: → `.length`
  - [ ] Substring, indexing → `.substring()`, `[i]`

- [ ] Type casts:
  - [ ] `as` (safe cast) → runtime type check + cast or throw
  - [ ] `is` (type test) → `instanceof` or metadata check

- [ ] Built-in functions:
  - [ ] `Inc(x)` → `x++` or `x += 1`
  - [ ] `Dec(x)` → `x--` or `x -= 1`
  - [ ] `Ord(x)`, `Chr(x)`, etc. → runtime functions or inline

- [ ] Tests for all advanced features

---

## Stage 11.4: LLVM Backend [OPTIONAL - Future Work]

**Goal**: Implement an LLVM IR backend to enable native code compilation. This stage is **deferred** and considered optional; it can be tackled later or by a different team.

**Why LLVM?**
- Native code performance (orders of magnitude faster than JS interpretation)
- Platform-native executables (Linux, macOS, Windows)
- Integration with existing LLVM toolchain (clang, lld, opt)
- Potential for ahead-of-time (AOT) compilation

**Exit Criteria** (when this stage is tackled):
- ✅ LLVM IR emitter generates valid `.ll` files
- ✅ Runtime library in C providing DWScript semantics (strings, arrays, objects, GC/RC)
- ✅ Can compile simple DWScript programs to native binaries
- ✅ Basic end-to-end tests: DWScript → LLVM IR → native exec → verify output
- ✅ Documentation: "DWScript → LLVM Compilation Guide"

### 11.4.1: LLVM Infrastructure (8 tasks)

- [ ] Choose LLVM binding:
  - [ ] **Option A**: `github.com/llir/llvm` (pure Go, generates LLVM IR text)
  - [ ] **Option B**: CGo bindings to LLVM C API (more features, harder to build)
  - [ ] Decision: Start with `llir/llvm` for easier integration

- [ ] Create `codegen/llvm/` package
  - [ ] `emitter.go` - LLVM IR emitter struct
  - [ ] `types.go` - MIR type → LLVM type mapping
  - [ ] `runtime.go` - runtime function declarations

- [ ] Type mapping:
  - [ ] DWScript `Integer` → LLVM `i32` or `i64`
  - [ ] DWScript `Float` → LLVM `double`
  - [ ] DWScript `Boolean` → LLVM `i1`
  - [ ] DWScript `String` → LLVM struct `{i32 len, i8* data}` (UTF-8)
  - [ ] DWScript arrays → LLVM struct with length + pointer
  - [ ] DWScript objects → LLVM struct with vtable pointer + fields

- [ ] Module structure:
  - [ ] Emit LLVM module with target triple
  - [ ] Declare external runtime functions
  - [ ] Emit global constructors for initialization

### 11.4.2: Runtime Library (12 tasks)

- [ ] Create `runtime/dws_runtime.h` - C header for runtime API
  - [ ] String operations:
    - [ ] `dws_string_t* dws_string_new(const char* data, int32_t len);`
    - [ ] `dws_string_t* dws_string_concat(dws_string_t* a, dws_string_t* b);`
    - [ ] `int32_t dws_string_len(dws_string_t* s);`
    - [ ] `char dws_string_index(dws_string_t* s, int32_t i);`
  - [ ] Array operations:
    - [ ] `dws_array_t* dws_array_new(int32_t elem_size, int32_t length);`
    - [ ] `void* dws_array_index(dws_array_t* arr, int32_t i);`
    - [ ] `int32_t dws_array_len(dws_array_t* arr);`
    - [ ] `void dws_array_set_len(dws_array_t* arr, int32_t new_len);`
  - [ ] Memory management:
    - [ ] `void* dws_alloc(size_t size);` (malloc wrapper or GC allocation)
    - [ ] `void dws_free(void* ptr);`
    - [ ] Decision: Use Boehm GC or reference counting? Document in `docs/runtime-memory.md`
  - [ ] Object operations:
    - [ ] `dws_object_t* dws_object_new(dws_class_t* class_info);`
    - [ ] Virtual dispatch helpers
  - [ ] Exception handling:
    - [ ] `void dws_throw(dws_exception_t* exc);`
    - [ ] `dws_exception_t* dws_catch(void);`
  - [ ] RTTI (Runtime Type Information):
    - [ ] `bool dws_is_instance(dws_object_t* obj, dws_class_t* class);`
    - [ ] `dws_object_t* dws_as_instance(dws_object_t* obj, dws_class_t* class);`

- [ ] Create `runtime/dws_runtime.c` - implement runtime in C
  - [ ] Implement all functions declared in header
  - [ ] String internals: reference-counted or GC-managed
  - [ ] Array internals: dynamic resizing, bounds checking

- [ ] Build runtime library:
  - [ ] `runtime/Makefile` to build `libdws_runtime.a`
  - [ ] CI: build runtime for Linux/macOS/Windows

### 11.4.3: LLVM Code Emission (15 tasks)

- [ ] Implement LLVM emitter:
  - [ ] `Generate(module *mir.Module) (string, error)` - generates LLVM IR text

- [ ] Emit functions:
  - [ ] Function declarations with correct signature
  - [ ] Parameter handling
  - [ ] Local variable allocation (`alloca`)

- [ ] Emit basic blocks:
  - [ ] Create LLVM basic blocks for each MIR block
  - [ ] Emit terminators: `br`, `br i1 %cond`, `ret`

- [ ] Emit instructions:
  - [ ] Arithmetic: `add`, `sub`, `mul`, `sdiv`, `srem`
  - [ ] Comparisons: `icmp eq`, `icmp slt`, etc.
  - [ ] Logical: `and`, `or`, `xor`
  - [ ] Memory: `alloca`, `load`, `store`
  - [ ] Calls: `call @function_name(args)`

- [ ] Emit constants:
  - [ ] Integer constants: `i32 42`
  - [ ] Float constants: `double 3.14`
  - [ ] String constants: global string + call to `dws_string_new`

- [ ] Emit control flow:
  - [ ] Conditional branches: `br i1 %cond, label %then, label %else`
  - [ ] Phi nodes: `%result = phi i32 [ %val1, %block1 ], [ %val2, %block2 ]`

- [ ] Runtime calls:
  - [ ] String operations: `call @dws_string_concat(...)`
  - [ ] Array operations: `call @dws_array_new(...)`
  - [ ] Object allocation: `call @dws_object_new(...)`

- [ ] Type conversion:
  - [ ] `IntToFloat`: `sitofp`
  - [ ] `FloatToInt`: `fptosi`

- [ ] Classes and objects:
  - [ ] Emit struct types for classes
  - [ ] Emit vtables as global constants
  - [ ] Virtual method dispatch: load vtable, index into it, call

- [ ] Exception handling:
  - [ ] Simple version: `call @dws_throw(...)` + unwind
  - [ ] Future: LLVM `invoke`/`landingpad` for proper EH

### 11.4.4: Linking and Execution (5 tasks)

- [ ] Compilation pipeline:
  - [ ] DWScript → MIR → LLVM IR (.ll file)
  - [ ] `llc` to compile .ll → object file (.o)
  - [ ] `clang` or `ld` to link object + runtime library → executable

- [ ] CLI integration:
  - [ ] `./bin/dwscript compile-native script.dws -o output`
  - [ ] Internally: generate .ll, invoke llc, invoke linker

- [ ] Tests:
  - [ ] 10+ end-to-end tests: DWScript → native binary → execute → verify output
  - [ ] Compare performance: JS vs native (expect 10-100x speedup for compute-heavy code)

### 11.4.5: Documentation (3 tasks)

- [ ] Create `docs/llvm-backend.md`:
  - [ ] How LLVM backend works
  - [ ] Type mappings
  - [ ] Runtime library design
  - [ ] How to build and link

- [ ] Create `docs/runtime-abi.md`:
  - [ ] Runtime function reference
  - [ ] Memory layout of DWScript types in LLVM
  - [ ] Calling conventions

- [ ] Update main README with native compilation instructions

---

## Testing Strategy (Cross-Cutting)

### Test Categories

1. **Unit Tests** (per package):
   - `mir/` package: type system, builder, verifier, lowering (aim for 85%+ coverage)
   - `codegen/js/` package: emitter methods, instruction lowering (85%+ coverage)
   - `codegen/llvm/` package: type mapping, emission (85%+ coverage)

2. **Golden Tests** (snapshot testing):
   - MIR dumps: `.dws` → `.mir` (human-readable MIR)
   - JS output: `.dws` → `.js` (check in generated JS for regression detection)
   - LLVM output: `.dws` → `.ll` (check in generated LLVM IR)

3. **Execution Tests** (end-to-end):
   - JS: Compile `.dws` → `.js`, run with Node.js, verify stdout
   - LLVM: Compile `.dws` → native binary, execute, verify stdout
   - Compare outputs: same DWScript program should produce same results in JS and native

4. **Integration Tests**:
   - Full pipeline: Lexer → Parser → Semantic → MIR → JS/LLVM → Execute
   - Test with real-world DWScript programs from `testdata/`

5. **Fuzzing** (optional, later):
   - Fuzz AST → MIR lowering to find crashes
   - Fuzz MIR verifier to ensure robustness

### CI Integration

- **GitHub Actions** workflow additions:
  - Install Node.js (for JS execution tests)
  - Install LLVM tools: `llc`, `opt` (for LLVM backend validation)
  - Run all tests with `go test -v -cover ./...`
  - Upload coverage reports to Codecov
  - Fail if coverage drops below 85% for new packages

### Coverage Goals

- `mir/` package: **85%+**
- `codegen/js/` package: **85%+**
- `codegen/llvm/` package: **85%+** (when implemented)
- Overall project: maintain **>80%** coverage

---

## Documentation Deliverables

1. **`docs/codegen-architecture.md`**:
   - Overview of MIR design
   - Why two-tier architecture?
   - How to add a new backend

2. **`docs/mir-spec.md`**:
   - Complete MIR instruction reference
   - Type system specification
   - SSA and CFG requirements
   - Examples for each construct

3. **`docs/js-backend.md`**:
   - DWScript → JavaScript mapping guide
   - What JS constructs are used for what DWScript features
   - Limitations and design decisions (e.g., GC, exception handling)
   - Performance characteristics

4. **`docs/llvm-backend.md`** (Stage 11.4):
   - DWScript → LLVM mapping
   - Runtime library design
   - How to compile and link
   - Platform-specific notes

5. **`docs/runtime-abi.md`** (Stage 11.4):
   - Runtime function reference
   - Memory layouts
   - Calling conventions

---

## Implementation Timeline Estimate

| Sub-Stage | Duration | Dependencies |
|-----------|----------|--------------|
| 11.1: MIR Foundation | ~2 weeks | Existing AST, types, semantic packages |
| 11.2: JS Backend MVP | ~3 weeks | 11.1 complete |
| 11.3: JS Feature Complete | ~4 weeks | 11.2 complete |
| 11.4: LLVM Backend | ~4 weeks (future) | 11.1 complete (can run in parallel with 11.2-11.3) |

**Total for JS codegen (11.1-11.3): ~9 weeks**

**Total including LLVM (11.1-11.4): ~13 weeks**

**Recommendation**: Implement 11.1 → 11.2 → 11.3 sequentially to get working JS output, then decide if/when to tackle 11.4 (LLVM) based on need.

---

## Task Summary

- **Stage 11.1**: 30 tasks (MIR foundation)
- **Stage 11.2**: 45 tasks (JS MVP)
- **Stage 11.3**: 60 tasks (JS feature complete)
- **Stage 11.4**: 45 tasks (LLVM backend, optional)

**Total**: ~180 tasks

**Current Progress**: 0/180 tasks completed (0%)

---

## Next Steps

1. **Review this plan** with the team
2. **Break down Stage 11.1 tasks** into GitHub issues
3. **Set up project structure**: Create `mir/`, `codegen/`, `runtime/` directories
4. **Start with MIR type system** (11.1.1) - foundation for everything else
5. **Iterate**: Build MIR incrementally, test with golden dumps
6. **Move to JS backend** once MIR is stable and verified

---

## Notes

- **Pragmatic approach**: The MIR layer adds upfront complexity but pays off with multi-backend support and easier debugging
- **Incremental delivery**: Each sub-stage delivers usable functionality
- **Testing-first mindset**: Tests are integrated throughout, not just at the end
- **Future-proof**: LLVM backend is designed in but deferred, making it easy to add later without rearchitecting
- **Aligned with go-dws style**: Matches the existing stage-based approach with detailed task breakdowns
