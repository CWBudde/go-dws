# Bytecode VM Quick Reference

**Quick-start guide for go-dws bytecode VM implementation**

## TL;DR - Key Decisions

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| **Architecture** | Stack-based | Simpler to implement, adequate performance (3-10x faster than AST) |
| **Instruction Size** | 32-bit fixed | Go-optimized, simple jumps, good cache locality |
| **Opcode Count** | 100-120 opcodes | Sweet spot for Go switch performance (<128 = 7 comparisons) |
| **Stack Type** | Fixed array `[10000]Value` | Zero allocations, predictable performance |
| **Value Type** | Struct with type tag | Simple, safe, readable (optimize later if needed) |

## Instruction Format (32-bit)

```go
type Instruction uint32

// Encoding formats:
// [8-bit opcode][24-bit unused]           - No operands
// [8-bit opcode][8-bit A][16-bit B]       - Two operands
// [8-bit opcode][8-bit A][16-bit Bx]      - A + large operand
// [8-bit opcode][24-bit Ax]               - Single large operand

func EncodeOp(op Opcode) Instruction
func EncodeOpA(op Opcode, a uint8) Instruction
func EncodeOpAB(op Opcode, a uint8, b uint16) Instruction
func (i Instruction) Opcode() Opcode
func (i Instruction) A() uint8
func (i Instruction) B() uint16
```

## Minimal Viable Instruction Set (40 opcodes)

### Constants & Variables (8)
```
LoadConst, LoadTrue, LoadFalse, LoadNil
LoadLocal, StoreLocal, LoadGlobal, StoreGlobal
```

### Arithmetic (11)
```
AddInt, SubInt, MulInt, DivInt, ModInt, NegInt
AddFloat, SubFloat, MulFloat, DivFloat, NegFloat
```

### Comparison & Logic (9)
```
Equal, NotEqual, Less, LessEqual, Greater, GreaterEqual
Not, And, Or
```

### Control Flow (5)
```
Jump, JumpIfFalse, JumpIfTrue, ForPrep, ForLoop
```

### Functions (3)
```
Call, Return, ReturnValue
```

### Stack (3)
```
Pop, Dup, Swap
```

### Misc (1)
```
Halt
```

## Core VM Structure

```go
type VM struct {
    // Execution state
    stack      [STACK_MAX]Value
    sp         int
    
    // Call frames
    frames     [MAX_FRAMES]CallFrame
    frameCount int
    
    // Configuration
    maxRecursion int  // Default: 1024
    trace        bool // Debug mode
}

type CallFrame struct {
    function   *CompiledFunction
    ip         int  // Instruction pointer
    stackBase  int  // Base of this frame's stack
}

type CompiledFunction struct {
    Name         string
    Instructions []Instruction
    Constants    []Value
    LocalCount   int
    ParamCount   int
}
```

## Compilation Pattern

```go
// Binary expression: a + b
func (c *Compiler) compileBinaryOp(node *ast.BinaryExpression) {
    c.compile(node.Left)   // Push a
    c.compile(node.Right)  // Push b
    c.emit(OpAddInt)       // Pop b, pop a, push a+b
}

// If statement
func (c *Compiler) compileIfStmt(node *ast.IfStatement) {
    c.compile(node.Condition)
    jumpElse := c.emit(OpJumpIfFalse, 0)  // Placeholder
    c.compile(node.ThenBranch)
    jumpEnd := c.emit(OpJump, 0)
    c.patchJump(jumpElse, c.currentOffset())
    c.compile(node.ElseBranch)
    c.patchJump(jumpEnd, c.currentOffset())
}
```

## VM Execution Loop

```go
func (vm *VM) Run() error {
    frame := &vm.frames[vm.frameCount-1]
    
    for {
        inst := frame.function.Instructions[frame.ip]
        frame.ip++
        
        switch inst.Opcode() {
        case OpLoadConst:
            vm.push(frame.function.Constants[inst.A()])
            
        case OpAddInt:
            b := vm.pop().AsInt()
            a := vm.pop().AsInt()
            vm.push(NewIntValue(a + b))
            
        case OpJumpIfFalse:
            if !vm.pop().AsBool() {
                frame.ip += int(inst.SignedB())
            }
            
        case OpCall:
            vm.callFunction(inst.A(), inst.B())
            
        case OpReturn:
            vm.returnFromFunction()
            
        case OpHalt:
            return nil
        }
    }
}
```

## Performance Expectations

| Stage | Implementation | Expected Speedup (vs AST) |
|-------|----------------|---------------------------|
| 12.1-12.3 | Basic bytecode + loops + functions | 2-3x |
| 12.4-12.5 | Arrays + objects | 3-5x |
| 12.6 | Optimizations (fusion, peephole) | 5-10x |
| 12.7-12.8 | Advanced features + polish | 5-10x |
| 13+ | JIT compilation (future) | 10-30x |

**Note:** Python's bytecode VM is ~10x faster than AST interpretation. GoAWK achieved 18% speedup overall with bytecode (starting from already-optimized tree-walker).

## Go-Specific Optimizations

### 1. Keep Opcode Count <128
- Go switch uses binary search: ⌈log₂(N)⌉ comparisons
- 128 opcodes = 7 comparisons per instruction
- GoAWK: 12% speedup reducing 100→85 opcodes

### 2. Fixed-Size Stack
```go
const STACK_MAX = 10000  // No GC pressure

func (vm *VM) push(v Value) {
    if vm.sp >= STACK_MAX {
        panic("stack overflow")
    }
    vm.stack[vm.sp] = v
    vm.sp++
}
```

### 3. Instruction Specialization
- Common cases get dedicated opcodes: `LoadConst0`, `LoadConst1`
- Reduces operand decoding overhead
- Example from GoAWK: `FieldInt` for `$i` when `i` is constant

### 4. Preallocate Constants
- Build constant pool at compile time
- No runtime allocations for literals

## Testing Checklist

- [ ] Instruction encoding/decoding roundtrip
- [ ] Compiler emits correct bytecode for each AST node type
- [ ] VM executes each opcode correctly (unit tests)
- [ ] Integration: compile + run simple programs
- [ ] Fixture compatibility: `testdata/fixtures/SimpleScripts/`
- [ ] Benchmark vs AST interpreter (target: 2-3x faster)
- [ ] Recursion depth tracking (raises exception at limit)
- [ ] Disassembler output matches bytecode

## Implementation Order (Stage 12 Tasks)

### Phase 1: Foundation (Tasks 12.1-12.6)
1. Define `Instruction` type and encoding functions
2. Implement 40 core opcodes (enum)
3. Create `VM` struct with stack and execution loop
4. Create `Compiler` struct with AST traversal
5. Compile and execute: literals, arithmetic, variables
6. Unit tests for instructions and VM

### Phase 2: Control Flow (Tasks 12.7-12.12)
7. Conditional jumps (`if-then-else`)
8. Loop instructions (`while`, `for`, `repeat`)
9. Boolean short-circuit (`and`, `or`)
10. `break`/`continue` support
11. Integration tests

### Phase 3: Functions (Tasks 12.13-12.18)
12. Call frames implementation
13. `Call` and `Return` instructions
14. Local variable scoping (stack-relative addressing)
15. Recursion depth tracking
16. Function fixture tests

### Phase 4: Data Structures (Tasks 12.19-12.24)
17. Arrays: `NewArray`, `LoadIndex`, `StoreIndex`
18. Records: `LoadField`, `StoreField`
19. Strings: `Concat`, `Length`, `Slice`
20. Array/record fixture tests

### Phase 5: OOP (Tasks 12.25-12.30)
21. Objects: `NewObject`, method calls
22. Virtual method dispatch
23. Property getters/setters
24. Class fixture tests

### Phase 6: Optimization (Tasks 12.31-12.36)
25. Peephole optimizer (instruction patterns)
26. Constant folding in compiler
27. Dead code elimination
28. Benchmark suite

### Phase 7: Advanced (Tasks 12.37-12.42)
29. Exception handling (`try-except-finally`)
30. Closures and upvalues
31. Interfaces
32. Full fixture compatibility

### Phase 8: Polish (Tasks 12.43-12.48)
33. Disassembler for debugging
34. Bytecode serialization (save/load)
35. Profiling hooks
36. Documentation and examples

## Common Pitfalls

1. **Jump offset encoding:** Use signed offsets, not absolute addresses
   ```go
   // WRONG: jumpTarget := 42
   // RIGHT: jumpOffset := targetIP - currentIP
   ```

2. **Stack pointer after pop:** Decrement BEFORE returning value
   ```go
   func (vm *VM) pop() Value {
       vm.sp--        // Decrement first
       return vm.stack[vm.sp]
   }
   ```

3. **Function call stack base:** Subtract argc before creating frame
   ```go
   stackBase := vm.sp - argc  // Arguments already on stack
   ```

4. **Short-circuit AND/OR:** Use conditional jumps, not boolean instructions
   ```go
   // a and b → compile a, JumpIfFalse end, compile b
   // NOT: compile a, compile b, OpAnd
   ```

5. **Backpatching jumps:** Reserve space, patch after knowing target
   ```go
   jumpIndex := c.emit(OpJump, 0)       // Placeholder
   c.compile(...)                       // Emit code
   c.patchJump(jumpIndex, c.offset())   // Fix offset
   ```

## References

- Full design doc: `docs/architecture/bytecode-vm-design.md`
- *Writing A Compiler In Go* by Thorsten Ball
- *Crafting Interpreters* (craftinginterpreters.com)
- GoAWK bytecode implementation: github.com/benhoyt/goawk
- Lua 5.3 bytecode reference
- Python `dis` module documentation

---

**Next Steps:**
1. Read full design document
2. Review PLAN.md Stage 12 tasks
3. Start with `internal/bytecode/` package (instruction definitions)
4. Implement `internal/compiler/` (AST → bytecode)
5. Implement `internal/vm/` (bytecode execution)

