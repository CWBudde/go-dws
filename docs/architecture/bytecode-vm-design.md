# Bytecode VM Instruction Set Design Research

**Author:** Research conducted for go-dws bytecode VM implementation  
**Date:** 2025-11-05  
**Status:** Design Research Document

## Executive Summary

This document synthesizes research on bytecode virtual machine (VM) instruction set designs to guide the implementation of a stack-based bytecode VM for go-dws. The original DWScript uses direct JIT compilation to x86 machine code, but go-dws will implement a bytecode VM as an intermediate approach between AST interpretation (current) and native code generation (future).

**Key Recommendations:**
- **Architecture:** Stack-based VM (simpler implementation, adequate performance)
- **Instruction Size:** 32-bit opcodes (Go-optimized, simpler jump offsets)
- **Opcode Count:** Target 80-120 opcodes (balanced coverage without switch overhead)
- **Format:** 1-byte opcode + variable operands OR compact 32-bit encoding

---

## Runtime Selection (AST vs Bytecode)

The VM can now be exercised end-to-end without custom harnesses:

- **CLI:** `dwscript run script.dws --bytecode` executes the program via the bytecode VM instead of the AST interpreter (experimental; unit loading remains interpreter-only for now).
- **Go API:** `dwscript.New(dwscript.WithCompileMode(dwscript.CompileModeBytecode))` compiles scripts to bytecode chunks and runs them through the VM. Omitting the option keeps the default AST interpreter.
- **Benchmarks:** `pkg/dwscript/compile_mode_bench_test.go` compares interpreter vs VM execution time on the same workload so regressions are easy to spot.

These hooks ensure feature work in the VM has high-visibility entry points while keeping AST execution as the default, battle-tested path.

---

## 1. Stack-Based vs Register-Based VMs

### Performance Comparison

| Metric | Stack-Based | Register-Based |
|--------|-------------|----------------|
| **Instruction Count** | Baseline (100%) | 46-47% fewer instructions |
| **Bytecode Size** | Smaller (100%) | 25-26% larger |
| **Execution Speed** | Baseline | 1.15x-1.48x faster |
| **Implementation Complexity** | Simple | Complex (register allocation) |

**Source:** "Virtual machine showdown: Stack versus registers" (ACM Transactions)

### Stack-Based Advantages (Chosen for go-dws)

1. **Simplicity**: No register allocation in compiler
2. **Smaller bytecode**: Operands mostly implicit (on stack)
3. **Easier to implement**: Standard push/pop operations
4. **Proven in education**: Used in JVM, CPython, WebAssembly

### Register-Based Advantages (Future Consideration)

1. **Performance**: Fewer VM dispatch loop iterations
2. **CPU mapping**: Closer to real hardware
3. **Optimization potential**: Easier to JIT compile
4. **Examples**: Lua (35 instructions), Dalvik VM

**Decision for go-dws:** Start with stack-based for Stage 12, consider register-based optimization in Stage 13+.

---

## 2. Common Bytecode Instruction Categories

### Category Breakdown (from JVM, CPython, Lua analysis)

#### 2.1 Load and Store (Variable Access)
**Purpose:** Transfer values between variables and stack

| Operation | Stack-Based Example | Notes |
|-----------|-------------------|-------|
| Load local | `LOAD_FAST <index>` | Push local variable to stack |
| Store local | `STORE_FAST <index>` | Pop stack to local variable |
| Load global | `LOAD_GLOBAL <name>` | Access global/module scope |
| Load constant | `LOAD_CONST <index>` | Push literal from constant pool |
| Load upvalue | `LOAD_DEREF <index>` | Access closure variables (DWScript: captures) |

**DWScript Specifics:**
- Need `LOAD_VAR`, `STORE_VAR` for local variables
- `LOAD_GLOBAL`, `STORE_GLOBAL` for unit-level variables
- `LOAD_CONST` for literals (integers, floats, strings, booleans)
- Upvalue support for nested functions and lambdas

#### 2.2 Arithmetic and Logic
**Purpose:** Mathematical and boolean operations

| Type | Operations | Instruction Count |
|------|-----------|------------------|
| Integer arithmetic | `+`, `-`, `*`, `div`, `mod` | 5 opcodes |
| Float arithmetic | `+`, `-`, `*`, `/`, `**` (power) | 5 opcodes |
| Unary | `-` (negate), `not` | 2 opcodes |
| Bitwise | `and`, `or`, `xor`, `shl`, `shr`, `not` | 6 opcodes |
| Comparison | `=`, `<>`, `<`, `>`, `<=`, `>=` | 6 opcodes |

**Total:** ~24 arithmetic opcodes

**Implementation Pattern (from Crafting Interpreters):**
```c
case OP_ADD: {
    double b = pop();  // Right operand (top of stack)
    double a = pop();  // Left operand
    push(a + b);
    break;
}
```

**DWScript Specifics:**
- Integer division (`div`) vs float division (`/`)
- `mod` for modulo
- Short-circuit evaluation for `and`/`or` (use conditional jumps)

#### 2.3 Control Flow
**Purpose:** Branching, loops, function calls

| Category | Instructions | Example Opcodes |
|----------|--------------|----------------|
| Unconditional jumps | `JUMP <offset>` | `JMP +10` |
| Conditional jumps | `JUMP_IF_TRUE`, `JUMP_IF_FALSE` | `POP_JUMP_IF_FALSE +5` |
| Comparison jumps | `JUMP_IF_EQ`, `JUMP_IF_LT`, etc. | Lua pattern |
| Loops | `FOR_PREP`, `FOR_LOOP` | Specialized for efficiency |
| Switch/Case | `TABLESWITCH`, `LOOKUPSWITCH` | JVM pattern |

**Total:** ~8-12 control flow opcodes

**Lua Pattern (Efficient):**
- Comparison instructions (EQ, LT, LE) set condition flag
- Following `JMP` instruction uses flag
- Saves one instruction per conditional branch

**DWScript Requirements:**
- `if-then-else`: Conditional jumps
- `while`/`repeat`: Loop jumps
- `for`: Specialized loop instructions (numeric)
- `case`: Table-based dispatch for integer expressions
- `break`/`continue`: Loop exit jumps

#### 2.4 Function Calls and Returns
**Purpose:** Function invocation, parameter passing, returns

| Operation | Instructions | Notes |
|-----------|--------------|-------|
| Call function | `CALL <argc>` | Push return address, jump to function |
| Call method | `CALL_METHOD <argc>` | Implicit `self` parameter |
| Return | `RETURN`, `RETURN_VALUE` | Pop frame, restore caller |
| Tail call | `TAIL_CALL` | Optimize recursion (Lua) |

**Stack Frame Structure:**
```
[return address]
[previous frame pointer]
[local variables...]
[parameters...]
[temporary stack values...]
```

**DWScript Specifics:**
- Function calls: `CALL <func_index> <argc>`
- Method calls: `CALL_METHOD <method_index> <argc>` (self on stack)
- Constructors: `CALL_CONSTRUCTOR <class_index> <argc>`
- Property accessors: Inline or special opcodes
- Recursion depth tracking (max 1024 by default)

#### 2.5 Object and Array Operations
**Purpose:** OOP support, data structure manipulation

| Type | Instructions | Example |
|------|--------------|---------|
| Array creation | `NEW_ARRAY <size>` | Allocate array |
| Array indexing | `LOAD_INDEX`, `STORE_INDEX` | `array[i]` |
| Array slice | `SLICE <start> <end>` | `array[1:5]` |
| Object creation | `NEW_OBJECT <class>` | Instantiate class |
| Field access | `LOAD_FIELD <name>`, `STORE_FIELD <name>` | `obj.field` |
| Property access | `LOAD_PROP <name>`, `STORE_PROP <name>` | Call getter/setter |

**JVM Pattern:**
- Typed array instructions: `iaload`, `faload`, `aaload` (int/float/object arrays)
- Field access via constant pool: `getfield <cp_index>`

**DWScript Requirements:**
- Dynamic arrays: `SetLength`, element access
- Static arrays: Fixed-size, bounds checking
- Records: Struct-like field access
- Classes: Field access, method dispatch, virtual calls
- Interfaces: Dynamic dispatch

#### 2.6 Type Operations
**Purpose:** Type checking, casting, introspection

| Operation | Instructions | Notes |
|-----------|--------------|-------|
| Type check | `INSTANCEOF <type>` | Boolean result |
| Cast | `CHECKCAST <type>` | Raise exception if invalid |
| Type conversion | `INT_TO_FLOAT`, `FLOAT_TO_INT` | Explicit conversions |

**DWScript Specifics:**
- `is` operator: Type checking
- `as` operator: Safe casting
- Implicit conversions: Integer ↔ Float
- Variant support: Dynamic typing (Stage 8+)

#### 2.7 Stack Manipulation
**Purpose:** Direct stack operations (less common in high-level bytecode)

| Operation | Instruction | Description |
|-----------|------------|-------------|
| Duplicate top | `DUP` | Copy TOS (top of stack) |
| Pop | `POP` | Discard TOS |
| Swap | `SWAP` | Exchange top two values |
| Rotate | `ROT_N <n>` | Rotate top N stack items |

**Use Cases:**
- `DUP`: When expression result used multiple times (e.g., `x := y := 5`)
- `SWAP`: Reverse operand order without temporary
- `POP`: Discard unused expression results

---

## 3. Instruction Format Design

### Option A: Variable-Length (1-byte opcode + operands)

**Format:**
```
[8-bit opcode] [operand bytes...]
```

**Examples:**
- `OP_RETURN`: 1 byte total (no operands)
- `OP_LOAD_CONST <index>`: 3 bytes (1 opcode + 2-byte index)
- `OP_JUMP <offset>`: 5 bytes (1 opcode + 4-byte signed offset)

**Advantages:**
- Compact code for simple instructions
- Standard approach (JVM, CPython)
- Easy disassembly

**Disadvantages:**
- Variable instruction length complicates IP (instruction pointer) management
- Requires instruction length table for disassembly

### Option B: Fixed 32-bit Instructions (Lua/GoAWK approach)

**Format (Lua 5.3):**
```
[6-bit opcode][8-bit A][9-bit B][9-bit C]
```

Alternative layouts:
- `[opcode][A][Bx]` - 18-bit combined operand
- `[opcode][Ax]` - 26-bit large operand

**Examples (from Lua):**
- `MOVE A B`: Copy register B to register A
- `LOADK A Bx`: Load constant Bx into register A
- `ADD A B C`: R(A) := RK(B) + RK(C)

**GoAWK Approach:**
```
[32-bit opcode/instruction combined]
```
- Opcode + operands packed into single uint32
- No separate fields, instruction decoder extracts data

**Advantages:**
- Simple IP arithmetic (always +1)
- Fast instruction fetch (single uint32 read)
- Good cache locality
- Efficient for Go (no byte-level manipulation)

**Disadvantages:**
- Larger code size for simple instructions
- Limited to ~2-3 operands per instruction

### Recommendation for go-dws: **Option B (32-bit)**

**Rationale:**
1. **Go switch performance**: Binary search on large switch statements (O(log N))
   - GoAWK found 12% speedup reducing 100→85 opcodes
   - 32-bit allows richer instruction set without explosion
   
2. **Jump simplification**: 2 billion offset range without variable-length encoding

3. **Memory alignment**: Go favors aligned access (uint32 natural)

4. **Implementation simplicity**: IP management easier

**Encoding Scheme:**
```go
type Instruction uint32

func EncodeOp(op Opcode) Instruction { ... }
func EncodeOpA(op Opcode, a uint8) Instruction { ... }
func EncodeOpAB(op Opcode, a uint8, b uint16) Instruction { ... }
func EncodeOpABx(op Opcode, a uint8, bx uint32) Instruction { ... }

func (i Instruction) Opcode() Opcode { return Opcode(i & 0xFF) }
func (i Instruction) A() uint8 { return uint8((i >> 8) & 0xFF) }
func (i Instruction) B() uint16 { return uint16((i >> 16) & 0xFFFF) }
```

---

## 4. Optimal Opcode Count

### Analysis from Real Implementations

| VM | Opcode Count | Architecture | Notes |
|----|--------------|--------------|-------|
| Lua 5.1-5.3 | 35-47 | Register-based | Minimal, highly optimized |
| JVM | ~202 (of 256 max) | Stack-based | Mature, feature-rich |
| CPython 3.14 | ~223 | Stack-based | Complex language features |
| CLR (CIL) | ~200 | Stack-based | OOP-focused |
| GoAWK | 85 | Stack-based (Go) | Balanced for switch performance |

### Opcode Budget Calculation

**Constraint:** Go switch statement binary search depth
- 50 opcodes: ⌈log₂(50)⌉ = 6 comparisons
- 85 opcodes: ⌈log₂(85)⌉ = 7 comparisons
- 128 opcodes: ⌈log₂(128)⌉ = 7 comparisons
- 200 opcodes: ⌈log₂(200)⌉ = 8 comparisons

**Sweet spot:** 64-128 opcodes (7 comparisons)

### Recommended Opcode Budget for go-dws

| Category | Estimated Opcodes | Notes |
|----------|------------------|-------|
| Constants and variables | 12 | Load/store local, global, const, upvalue |
| Arithmetic (int/float) | 24 | +, -, *, /, div, mod, pow, negate (typed) |
| Comparison | 8 | =, <>, <, >, <=, >= (int/float variants) |
| Logical | 4 | and, or, xor, not |
| Bitwise | 7 | and, or, xor, not, shl, shr |
| Control flow | 12 | jmp, jmp_if, jmp_if_not, for_prep, for_loop, etc. |
| Function calls | 8 | call, call_method, return, tail_call |
| Stack operations | 5 | dup, pop, swap, rot |
| Array operations | 8 | new_array, load_index, store_index, set_length |
| Object operations | 10 | new_object, load_field, store_field, load_prop, store_prop |
| Type operations | 6 | instanceof, checkcast, int2float, float2int |
| String operations | 4 | concat, length, index, slice |
| Miscellaneous | 8 | nop, halt, print, assert |
| **Total** | **~116 opcodes** | Within optimal range |

**Buffer for future:** Leave ~10-15 opcodes reserved for Stage 13+ features.

---

## 5. Go-Specific Considerations

### 5.1 Switch Statement Performance

**Problem:** Go compiles large switches as binary search trees, not jump tables

**Evidence (from GoAWK):**
```
80 opcodes: 6-7 comparisons per dispatch
100 opcodes: 7 comparisons per dispatch
→ 12% speedup reducing 100→85 opcodes
```

**Mitigation Strategies:**

1. **Keep opcode count <128** (7 comparisons max)

2. **Consider function dispatch table** (minimal gain per GoAWK):
```go
var opcodeHandlers = [256]func(*VM){
    OpAdd: (*VM).execAdd,
    OpSub: (*VM).execSub,
    // ...
}

func (vm *VM) Run() {
    for vm.ip < len(vm.code) {
        op := vm.code[vm.ip].Opcode()
        vm.ip++
        opcodeHandlers[op](vm)
    }
}
```

3. **Wait for Go compiler improvements** (jump tables may come)

### 5.2 Stack Implementation

**Options:**

**A. Slice (dynamic):**
```go
type VM struct {
    stack []Value
    sp    int  // Stack pointer
}

func (vm *VM) Push(v Value) {
    vm.stack = append(vm.stack, v)
}

func (vm *VM) Pop() Value {
    v := vm.stack[len(vm.stack)-1]
    vm.stack = vm.stack[:len(vm.stack)-1]
    return v
}
```
- **Pros:** No overflow errors, simple
- **Cons:** Allocation overhead, GC pressure

**B. Fixed array (pre-allocated):**
```go
type VM struct {
    stack [STACK_MAX]Value
    sp    int
}

func (vm *VM) Push(v Value) {
    vm.stack[vm.sp] = v
    vm.sp++
}

func (vm *VM) Pop() Value {
    vm.sp--
    return vm.stack[vm.sp]
}
```
- **Pros:** Zero allocations, fast
- **Cons:** Fixed size (needs overflow check)

**Recommendation:** **Option B** with overflow checking
- `STACK_MAX = 10000` (plenty for DWScript recursion depth)
- Check `vm.sp < STACK_MAX` before push
- Raise `EScriptStackOverflow` on overflow

### 5.3 Value Representation

**Challenge:** DWScript has multiple types (Integer, Float, String, Boolean, Object, Array)

**Options:**

**A. Interface:**
```go
type Value interface {
    Type() ValueType
    AsInt() int64
    AsFloat() float64
    // ...
}
```
- **Cons:** Heap allocation, interface call overhead

**B. Struct with type tag:**
```go
type Value struct {
    typ   ValueType
    ival  int64
    fval  float64
    sval  string
    oval  *Object
}
```
- **Cons:** Memory waste (only one field used)

**C. Variant with union (unsafe):**
```go
type Value struct {
    typ  ValueType
    data uint64  // Reinterpret as int/float/pointer
}
```
- **Pros:** 16 bytes per value, fast
- **Cons:** Unsafe, complex

**D. NaN-boxing (advanced):**
- Store small integers and pointers in NaN float64 patterns
- **Pros:** Single 8-byte value, cache-efficient
- **Cons:** Very complex, platform-specific

**Recommendation for Stage 12:** **Option B** (struct with type tag)
- Simple, safe, readable
- Optimize to Option C/D in Stage 13 if profiling shows bottleneck

### 5.4 Constant Pool

**Purpose:** Store literals (numbers, strings) referenced by bytecode

**Structure:**
```go
type ConstantPool struct {
    integers []int64
    floats   []float64
    strings  []string
    // Future: functions, classes
}

type Instruction uint32
// LOAD_CONST_INT <index> → push integers[index]
// LOAD_CONST_FLOAT <index> → push floats[index]
// LOAD_CONST_STRING <index> → push strings[index]
```

**Alternative (unified):**
```go
type ConstantPool struct {
    values []Value
}
// LOAD_CONST <index> → push values[index]
```

**Recommendation:** Unified pool (simpler, extensible)

### 5.5 Call Stack and Frames

**Frame Structure:**
```go
type CallFrame struct {
    function    *CompiledFunction
    ip          int          // Instruction pointer
    stackBase   int          // Stack base for this frame
    localCount  int          // Number of local variables
}

type VM struct {
    frames      []CallFrame
    frameCount  int
    maxRecursion int  // Default 1024
}
```

**Function Call Sequence:**
1. Push arguments to stack
2. Execute `CALL <func_index> <argc>`
3. Create new frame: `frames[frameCount++] = CallFrame{...}`
4. Check `frameCount < maxRecursion` → raise exception if exceeded
5. Set `stackBase = sp - argc`
6. Execute function bytecode
7. On `RETURN`, pop frame: `frameCount--`
8. Restore caller's IP and stackBase

**Local Variable Access:**
```go
// LOAD_LOCAL <index>
value := vm.stack[vm.frames[vm.frameCount-1].stackBase + index]
vm.Push(value)
```

---

## 6. Example Instruction Set (Minimal Viable)

### 6.1 Core Instructions (~40 opcodes)

```go
const (
    // Constants and literals (4)
    OpLoadConst   = iota  // Load constant from pool
    OpLoadTrue            // Push true
    OpLoadFalse           // Push false
    OpLoadNil             // Push nil/null

    // Local variables (2)
    OpLoadLocal           // Load local variable
    OpStoreLocal          // Store to local variable

    // Global variables (2)
    OpLoadGlobal          // Load global variable
    OpStoreGlobal         // Store to global variable

    // Arithmetic - Integer (6)
    OpAddInt
    OpSubInt
    OpMulInt
    OpDivInt              // Integer division (div)
    OpModInt              // Modulo
    OpNegInt              // Unary minus

    // Arithmetic - Float (5)
    OpAddFloat
    OpSubFloat
    OpMulFloat
    OpDivFloat
    OpNegFloat

    // Comparison (6)
    OpEqual               // =
    OpNotEqual            // <>
    OpLess                // <
    OpLessEqual           // <=
    OpGreater             // >
    OpGreaterEqual        // >=

    // Logical (3)
    OpNot
    OpAnd                 // Short-circuit via conditional jump
    OpOr                  // Short-circuit via conditional jump

    // Control flow (5)
    OpJump                // Unconditional jump
    OpJumpIfFalse         // Conditional jump (pop stack)
    OpJumpIfTrue          // Conditional jump (pop stack)
    OpForPrep             // Initialize for-loop
    OpForLoop             // Iterate for-loop

    // Functions (3)
    OpCall                // Call function
    OpReturn              // Return from function
    OpReturnValue         // Return value from function

    // Stack manipulation (3)
    OpPop                 // Discard TOS
    OpDup                 // Duplicate TOS
    OpSwap                // Swap top two

    // Miscellaneous (1)
    OpHalt                // Stop execution
)
```

**Total:** 40 opcodes (minimal viable for Stage 12.1-12.3)

### 6.2 Extended Instructions (~76 additional)

Add as needed for DWScript features:

- Arrays: `OpNewArray`, `OpLoadIndex`, `OpStoreIndex`, `OpSetLength`
- Records: `OpLoadField`, `OpStoreField`
- Classes: `OpNewObject`, `OpLoadProperty`, `OpStoreProperty`, `OpCallMethod`
- Type operations: `OpInstanceof`, `OpCheckcast`, `OpIntToFloat`, `OpFloatToInt`
- String operations: `OpConcat`, `OpLength`, `OpStringIndex`
- Advanced control: `OpSwitch`, `OpCase`, `OpBreak`, `OpContinue`

---

## 7. Compilation Strategy

### 7.1 AST → Bytecode Compiler

**Input:** AST nodes (from `internal/ast/`)
**Output:** Bytecode instructions + constant pool

**Compilation Phases:**

1. **Symbol Table Pass** (existing: `internal/semantic/`)
   - Resolve variable scopes
   - Assign local variable indices
   - Build function signatures

2. **Code Generation Pass**
   - Traverse AST depth-first
   - Emit instructions to bytecode buffer
   - Populate constant pool
   - Track jump targets (backpatching)

**Example: Binary Expression**
```go
func (c *Compiler) compileBinaryOp(node *ast.BinaryExpression) {
    // Compile left operand → pushes to stack
    c.compile(node.Left)
    
    // Compile right operand → pushes to stack
    c.compile(node.Right)
    
    // Emit operation instruction
    switch node.Operator {
    case "+":
        if node.Type == types.Integer {
            c.emit(OpAddInt)
        } else {
            c.emit(OpAddFloat)
        }
    // ...
    }
}
```

**Example: If Statement**
```go
func (c *Compiler) compileIfStatement(node *ast.IfStatement) {
    // Compile condition → pushes boolean
    c.compile(node.Condition)
    
    // Jump to else/end if false
    jumpToElse := c.emit(OpJumpIfFalse, 0)  // Placeholder offset
    
    // Compile then branch
    c.compile(node.ThenBranch)
    
    // Jump over else branch
    jumpToEnd := c.emit(OpJump, 0)
    
    // Backpatch else jump
    c.patchJump(jumpToElse, c.currentAddress())
    
    // Compile else branch (if exists)
    if node.ElseBranch != nil {
        c.compile(node.ElseBranch)
    }
    
    // Backpatch end jump
    c.patchJump(jumpToEnd, c.currentAddress())
}
```

### 7.2 Bytecode Structure

```go
type CompiledFunction struct {
    Name         string
    Instructions []Instruction  // Bytecode
    Constants    *ConstantPool
    LocalCount   int           // Number of local variables
    ParamCount   int           // Number of parameters
}

type CompiledProgram struct {
    Functions []*CompiledFunction
    Globals   []string          // Global variable names
    Main      *CompiledFunction // Entry point
}
```

---

## 8. VM Execution Loop

### 8.1 Core Loop Structure

```go
func (vm *VM) Run() error {
    frame := &vm.frames[vm.frameCount-1]
    
    for {
        // Fetch
        instruction := frame.function.Instructions[frame.ip]
        frame.ip++
        
        // Decode
        op := instruction.Opcode()
        
        // Execute
        switch op {
        case OpLoadConst:
            index := instruction.A()
            vm.Push(frame.function.Constants.Values[index])
            
        case OpAddInt:
            b := vm.Pop().AsInt()
            a := vm.Pop().AsInt()
            vm.Push(NewIntValue(a + b))
            
        case OpJumpIfFalse:
            offset := instruction.SignedB()
            if !vm.Pop().AsBool() {
                frame.ip += offset
            }
            
        case OpCall:
            funcIndex := instruction.A()
            argc := instruction.B()
            return vm.callFunction(funcIndex, argc)
            
        case OpReturn:
            return vm.returnFromFunction()
            
        case OpHalt:
            return nil
            
        default:
            return fmt.Errorf("unknown opcode: %d", op)
        }
        
        // Debug trace (optional)
        if vm.trace {
            vm.dumpStack()
        }
    }
}
```

### 8.2 Optimization Techniques

**1. Instruction Fusion**
- Combine common patterns: `LOAD_CONST 1` + `ADD_INT` → `INC`
- Reduces dispatch overhead

**2. Specialized Instructions**
- `OpLoadConst0`, `OpLoadConst1` (no operand needed)
- `OpLoadLocal0`, `OpLoadLocal1` (common cases)
- GoAWK: `FieldInt` for integer field access

**3. Inline Simple Operations**
- Don't switch for trivial ops like `OpPop` if hot path

**4. Computed Goto (future, non-standard Go)**
- Replace switch with GCC/Clang computed goto
- Requires cgo or assembler

#### 8.2.1 Bytecode Optimization Pipeline

The compiler now routes every `Chunk` through a configurable optimization pipeline (see `internal/bytecode/optimizer.go`). Each pass implements a small, local rewrite and the pass manager replays them in order whenever `Chunk.Optimize()` is invoked.

- **Pass controls:** use `WithOptimizationPass(pass, enabled)` to toggle any pass when calling `Chunk.Optimize`. Passes default to `enabled=true`, so existing call sites can omit options.
- **Named passes:** 
  - `PassLiteralDiscard` removes literal `LOAD_*` instructions that are immediately popped.
  - `PassStackShuffle` collapses redundant stack shuffles (e.g., `DUP`/`POP`, `SWAP` pairs, and `ROTATE3` triples).
  - `PassInlineSmall` substitutes call sites to leaf functions (< ~10 instructions, no locals/upvalues/args) directly with their instruction stream.
  - `PassConstPropagation` tracks simple literal locals/globals and folds straight-line arithmetic/comparison sequences down to single constant loads.
  - `PassDeadCode` trims instructions that appear after unconditional terminators (`RETURN`, `HALT`, `JUMP`, etc.) unless another jump targets that instruction.
- **Metadata safety:** the pass manager recomputes line tables, jump offsets, and try metadata after any pass mutates the instruction stream, so downstream tooling continues to report accurate locations.
- **Future work:** additional passes (dead-code elimination, constant propagation, inlining) can be added by registering another `optimizerPass` item; the manager automatically honors the same toggle plumbing.

Example:

```go
chunk.Optimize(
    WithOptimizationPass(PassLiteralDiscard, false), // disable literal folding
)
```

---

## 9. Testing Strategy

### 9.1 Unit Tests

**Instruction Encoding/Decoding:**
```go
func TestInstructionEncoding(t *testing.T) {
    inst := EncodeOpAB(OpLoadLocal, 5, 100)
    assert.Equal(t, OpLoadLocal, inst.Opcode())
    assert.Equal(t, uint8(5), inst.A())
    assert.Equal(t, uint16(100), inst.B())
}
```

**Compiler Output:**
```go
func TestCompileBinaryExpression(t *testing.T) {
    ast := parseBinaryExpr("3 + 5")
    bytecode := compile(ast)
    expected := []Instruction{
        EncodeOpA(OpLoadConst, 0),  // 3
        EncodeOpA(OpLoadConst, 1),  // 5
        EncodeOp(OpAddInt),
    }
    assert.Equal(t, expected, bytecode)
}
```

**VM Execution:**
```go
func TestVMAdd(t *testing.T) {
    vm := NewVM()
    vm.code = []Instruction{
        EncodeOpA(OpLoadConst, 0),  // Push 3
        EncodeOpA(OpLoadConst, 1),  // Push 5
        EncodeOp(OpAddInt),         // Add
        EncodeOp(OpHalt),
    }
    vm.constants = []Value{NewIntValue(3), NewIntValue(5)}
    
    vm.Run()
    
    assert.Equal(t, NewIntValue(8), vm.Pop())
}
```

### 9.2 Integration Tests

**DWScript Fixture Compatibility:**
- Reuse existing `testdata/fixtures/` test suite
- Compare bytecode VM output against AST interpreter output
- Target: 100% pass rate on SimpleScripts category by Stage 12.4

### 9.3 Benchmarks

```go
func BenchmarkFibonacci(b *testing.B) {
    // Test recursive function performance
}

func BenchmarkLoop(b *testing.B) {
    // Test for-loop performance
}

func BenchmarkMethodCall(b *testing.B) {
    // Test virtual method dispatch overhead
}
```

**Performance Targets (vs AST interpreter):**
- **Stage 12.4:** 2-3x faster (basic bytecode)
- **Stage 12.8:** 5-10x faster (optimized bytecode)
- **Stage 13:** 10-30x faster (JIT compilation)

---

## 10. Implementation Roadmap

### Stage 12.1: Bytecode Foundation (PLAN.md tasks 12.1-12.6)
- [ ] Define instruction format (32-bit encoding)
- [ ] Implement ~40 core opcodes
- [ ] Basic compiler: literals, arithmetic, variables
- [ ] Basic VM: stack operations, simple expressions

### Stage 12.2: Control Flow (Tasks 12.7-12.12)
- [ ] Conditional jumps (if-then-else)
- [ ] Loops (while, for, repeat)
- [ ] Boolean short-circuit evaluation

### Stage 12.3: Functions (Tasks 12.13-12.18)
- [ ] Function calls and returns
- [ ] Call frames and stack frames
- [ ] Local variable scoping
- [ ] Recursion depth tracking

### Stage 12.4: Arrays and Records (Tasks 12.19-12.24)
- [ ] Array creation and indexing
- [ ] Record field access
- [ ] String operations

### Stage 12.5: Classes (Tasks 12.25-12.30)
- [ ] Object creation and field access
- [ ] Method calls (static and instance)
- [ ] Virtual method dispatch
- [ ] Property getters/setters

### Stage 12.6: Optimization (Tasks 12.31-12.36)
- [ ] Instruction fusion
- [ ] Peephole optimization
- [ ] Dead code elimination
- [ ] Constant folding in compiler

### Stage 12.7: Advanced Features (Tasks 12.37-12.42)
- [ ] Exception handling (try-except-finally)
- [ ] Closures and upvalues
- [ ] Interfaces and dynamic dispatch
- [ ] Enums (already implemented, integrate with bytecode)

### Stage 12.8: Polishing (Tasks 12.43-12.48)
- [ ] Disassembler for debugging
- [ ] Bytecode serialization (save/load compiled programs)
- [ ] Profiling hooks
- [ ] Complete fixture test compatibility

---

## 11. References

### Academic Papers
- Yunhe Shi et al., "Virtual machine showdown: Stack versus registers," ACM Transactions on Architecture and Code Optimization, 2008
- Romer et al., "The structure and performance of interpreters," ASPLOS 1996

### Books
- Thorsten Ball, *Writing A Compiler In Go*, 2018
- Robert Nystrom, *Crafting Interpreters*, 2021 (https://craftinginterpreters.com/)

### Technical Documentation
- Lua 5.3 Bytecode Reference: https://the-ravi-programming-language.readthedocs.io/
- Java Virtual Machine Specification: https://docs.oracle.com/javase/specs/jvms/
- Python `dis` module: https://docs.python.org/3/library/dis.html
- .NET Common Intermediate Language (CIL): https://en.wikipedia.org/wiki/List_of_CIL_instructions

### Blog Posts and Articles
- Ben Hoyt, "Optimizing GoAWK with a bytecode compiler and VM," https://benhoyt.com/writings/goawk-compiler-vm/
- "The Design & Implementation of the CPython Virtual Machine," https://blog.codingconfessions.com/
- "Python behind the scenes #4: how Python bytecode is executed," https://tenthousandmeters.com/

### Example Implementations (Go)
- GoAWK bytecode compiler: https://github.com/benhoyt/goawk
- skx/go.vm: Simple VM in Go: https://github.com/skx/go.vm
- go-interpreter proposal: https://github.com/go-interpreter/proposal

---

## Appendix A: Complete Proposed Instruction Set (116 opcodes)

### Constants and Loading (12)
```
OpLoadConst, OpLoadTrue, OpLoadFalse, OpLoadNil
OpLoadLocal, OpStoreLocal
OpLoadGlobal, OpStoreGlobal
OpLoadUpvalue, OpStoreUpvalue
OpLoadConst0, OpLoadConst1  // Optimized common cases
```

### Arithmetic - Integer (12)
```
OpAddInt, OpSubInt, OpMulInt, OpDivInt, OpModInt, OpPowInt
OpNegInt, OpIncInt, OpDecInt
OpShlInt, OpShrInt  // Bitwise shifts
OpAbsInt            // Absolute value
```

### Arithmetic - Float (12)
```
OpAddFloat, OpSubFloat, OpMulFloat, OpDivFloat, OpPowFloat
OpNegFloat, OpAbsFloat, OpSqrtFloat, OpSinFloat, OpCosFloat
OpFloorFloat, OpCeilFloat
```

### Comparison (8)
```
OpEqual, OpNotEqual
OpLessInt, OpLessEqualInt, OpGreaterInt, OpGreaterEqualInt
OpLessFloat, OpGreaterFloat
```

### Logical (4)
```
OpNot, OpAnd, OpOr, OpXor
```

### Bitwise (7)
```
OpBitAnd, OpBitOr, OpBitXor, OpBitNot
OpShl, OpShr
OpRotateLeft, OpRotateRight
```

### Control Flow (12)
```
OpJump, OpJumpIfTrue, OpJumpIfFalse
OpForPrep, OpForLoop, OpForEach
OpSwitch, OpCase, OpDefault
OpBreak, OpContinue
OpNop
```

### Functions (8)
```
OpCall, OpCallMethod, OpCallStatic
OpReturn, OpReturnValue
OpTailCall
OpClosure, OpClosureCapture
```

### Stack Manipulation (5)
```
OpPop, OpDup, OpDup2, OpSwap, OpRot3
```

### Arrays (8)
```
OpNewArray, OpLoadIndex, OpStoreIndex
OpArrayLength, OpSetLength
OpArraySlice
OpArrayAppend, OpArrayInsert
```

### Records and Objects (10)
```
OpLoadField, OpStoreField
OpLoadProp, OpStoreProp
OpNewObject, OpNewRecord
OpInstanceof, OpCheckcast
OpGetClass, OpIsNil
```

### Strings (4)
```
OpConcat, OpStringLength, OpStringIndex, OpStringSlice
```

### Type Conversions (6)
```
OpIntToFloat, OpFloatToInt
OpIntToString, OpFloatToString, OpBoolToString
OpStringToInt
```

### Exception Handling (4)
```
OpTry, OpCatch, OpFinally, OpThrow
```

### Miscellaneous (4)
```
OpHalt, OpPrint, OpAssert, OpDebugBreak
```

**Total:** 116 opcodes

---

## Appendix B: Bytecode Examples

### Example 1: Simple Arithmetic
**Source:**
```pascal
var x := 3 + 5 * 2;
```

**Bytecode:**
```
LOAD_CONST 0      ; Push 3
LOAD_CONST 1      ; Push 5
LOAD_CONST 2      ; Push 2
MUL_INT           ; 5 * 2 = 10
ADD_INT           ; 3 + 10 = 13
STORE_GLOBAL 0    ; x := 13
HALT
```

### Example 2: Conditional
**Source:**
```pascal
if x > 10 then
  y := 1
else
  y := 2;
```

**Bytecode:**
```
LOAD_GLOBAL 0     ; Load x
LOAD_CONST 0      ; Push 10
GREATER_INT       ; x > 10
JUMP_IF_FALSE +5  ; Jump to else if false
LOAD_CONST 1      ; Push 1
STORE_GLOBAL 1    ; y := 1
JUMP +3           ; Jump to end
LOAD_CONST 2      ; Push 2 (else branch)
STORE_GLOBAL 1    ; y := 2
HALT
```

### Example 3: Function Call
**Source:**
```pascal
function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;

var sum := Add(3, 5);
```

**Bytecode (main):**
```
LOAD_CONST 0      ; Push 3
LOAD_CONST 1      ; Push 5
CALL 0, 2         ; Call function 0 with 2 args
STORE_GLOBAL 0    ; sum := result
HALT
```

**Bytecode (function 0 - Add):**
```
; Parameters: a=local[0], b=local[1]
LOAD_LOCAL 0      ; Load a
LOAD_LOCAL 1      ; Load b
ADD_INT           ; a + b
RETURN_VALUE      ; Return result
```

### Example 4: For Loop
**Source:**
```pascal
var sum := 0;
for i := 1 to 10 do
  sum := sum + i;
```

**Bytecode:**
```
LOAD_CONST 0      ; Push 0
STORE_GLOBAL 0    ; sum := 0

LOAD_CONST 1      ; Push 1 (start)
LOAD_CONST 2      ; Push 10 (end)
FOR_PREP +8       ; Initialize loop, jump to end if done

; Loop body
LOAD_GLOBAL 0     ; Load sum
LOAD_LOCAL 0      ; Load i (loop counter)
ADD_INT           ; sum + i
STORE_GLOBAL 0    ; sum := result

FOR_LOOP -6       ; Increment i, jump back if i <= 10
HALT
```

---

**END OF DOCUMENT**
