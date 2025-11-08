# Bytecode VM Documentation

## Overview

The go-dws bytecode VM provides a fast alternative execution mode to the tree-walking AST interpreter. By compiling DWScript source code to bytecode and executing it on a stack-based virtual machine, we achieve approximately **5-6x performance improvement** over the AST interpreter.

## Architecture

### Components

1. **Compiler** (`internal/bytecode/compiler.go`)
   - Transforms AST nodes into bytecode chunks
   - Performs constant folding and basic optimizations
   - Manages local variables, globals, and upvalues (closures)
   - Handles function compilation and nested scopes

2. **Bytecode** (`internal/bytecode/bytecode.go`)
   - Defines the instruction format (32-bit instructions)
   - Manages constant pools for literals
   - Provides line number mapping for error reporting
   - Includes value types (integers, floats, strings, arrays, objects, functions, closures, built-ins)

3. **VM** (`internal/bytecode/vm.go`)
   - Stack-based execution engine
   - Handles instruction dispatch via switch statement
   - Manages call frames and closure upvalues
   - Provides built-in function support
   - Captures output for I/O operations

4. **Disassembler** (`internal/bytecode/disasm.go`)
   - Converts bytecode back to human-readable format
   - Useful for debugging and understanding compiled output

### Design Decisions

**Stack-Based vs Register-Based**: We chose a stack-based VM for simplicity and compatibility with DWScript's evaluation model. While register-based VMs can be faster, stack-based VMs are easier to implement and debug.

**Instruction Format**: 32-bit instructions with 8-bit opcode and 24 bits for operands, allowing for:
- Single-byte opcode (up to 256 instructions)
- Two 12-bit operands OR
- One 8-bit and one 16-bit operand OR
- One 24-bit operand

This format balances flexibility with compactness.

## Usage

### Command Line

Execute a script using the bytecode VM:

```bash
# Run with bytecode VM
dwscript run --bytecode script.dws

# Compare with AST interpreter
dwscript run script.dws

# Show disassembled bytecode
dwscript run --bytecode --trace script.dws
```

### Programmatic API

```go
import "github.com/cwbudde/go-dws/pkg/dwscript"

// Create engine with bytecode mode
engine, _ := dwscript.New(
    dwscript.WithCompileMode(dwscript.CompileModeBytecode),
)

// Compile and run
result, err := engine.Eval(`
    var x: Integer := 42;
    PrintLn('Answer: ' + IntToStr(x));
`)

if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Output) // "Answer: 42"
```

## Instruction Set

### Categories

**Load/Store Operations**
- `LOAD_CONST` - Load constant from constant pool
- `LOAD_LOCAL` - Load local variable
- `STORE_LOCAL` - Store to local variable
- `LOAD_GLOBAL` - Load global variable
- `STORE_GLOBAL` - Store to global variable
- `LOAD_UPVALUE` - Load captured upvalue (closure)
- `STORE_UPVALUE` - Store to captured upvalue

**Arithmetic Operations**
- `ADD_INT`, `ADD_FLOAT` - Addition
- `SUB_INT`, `SUB_FLOAT` - Subtraction
- `MUL_INT`, `MUL_FLOAT` - Multiplication
- `DIV_INT`, `DIV_FLOAT` - Division
- `MOD_INT` - Modulo
- `NEGATE_INT`, `NEGATE_FLOAT` - Unary negation

**Comparison Operations**
- `EQUAL`, `NOT_EQUAL` - Equality comparison
- `LESS`, `LESS_EQUAL` - Less than comparison
- `GREATER`, `GREATER_EQUAL` - Greater than comparison

**Logical Operations**
- `AND`, `OR`, `XOR` - Bitwise operations
- `NOT` - Logical NOT

**Control Flow**
- `JUMP` - Unconditional jump
- `JUMP_IF_FALSE` - Conditional jump (if false)
- `JUMP_IF_TRUE` - Conditional jump (if true)
- `LOOP` - Jump backward (for loops)
- `RETURN` - Return from function
- `HALT` - Stop execution

**Function Operations**
- `CALL` - Call function by constant index
- `CALL_INDIRECT` - Call function from stack
- `CALL_METHOD` - Call object method
- `CALL_VIRTUAL` - Virtual method dispatch
- `CLOSURE` - Create closure with upvalues

**Array/Object Operations**
- `NEW_ARRAY` - Create new array
- `GET_INDEX` - Array/string indexing
- `SET_INDEX` - Array element assignment
- `ARRAY_LENGTH` - Get array length
- `NEW_OBJECT` - Create object instance
- `GET_FIELD` - Get object field
- `SET_FIELD` - Set object field
- `GET_PROPERTY` - Get object property
- `SET_PROPERTY` - Set object property

**String Operations**
- `STRING_CONCAT` - Concatenate strings

**Type Conversion**
- `INT_TO_FLOAT` - Convert integer to float
- `FLOAT_TO_INT` - Convert float to integer

**Exception Handling**
- `TRY` - Begin try block
- `CATCH` - Catch exception
- `FINALLY` - Finally block
- `THROW` - Throw exception
- `GET_EXCEPT_OBJECT` - Get current exception

## Built-in Functions

The VM provides the following built-in functions:

- **`PrintLn(...)`** - Print values with newline
- **`Print(...)`** - Print values without newline
- **`IntToStr(int)`** - Convert integer to string
- **`FloatToStr(float)`** - Convert float to string
- **`StrToInt(string)`** - Parse integer from string
- **`StrToFloat(string)`** - Parse float from string
- **`Length(string|array)`** - Get string/array length
- **`Copy(string, start, length)`** - Extract substring (1-based indexing)
- **`Ord(string|int)`** - Get character code or enum ordinal
- **`Chr(int)`** - Convert character code to string

These are registered as global functions and available to all bytecode programs.

## Example: Bytecode Output

For this simple program:

```pascal
var x: Integer := 42;
PrintLn(IntToStr(x));
```

The disassembled bytecode looks like:

```
== Bytecode (test) ==
0000    1 LOAD_CONST         0 (42)
0001    | STORE_GLOBAL       1
0002    | LOAD_GLOBAL        1 (PrintLn)
0003    | LOAD_GLOBAL        3 (IntToStr)
0004    | LOAD_GLOBAL        1
0005    | CALL               1
0006    | CALL               1
0007    | HALT
```

## Performance Characteristics

### Benchmarks

Based on internal benchmarks, the bytecode VM shows:

- **5-6x faster** than AST interpreter for loops and arithmetic
- **Lower memory overhead** - bytecode is more compact than AST
- **Faster startup** for repeated executions (compile once, run many times)

Example benchmark (100-iteration loop):
- AST Interpreter: ~82,873 ns/op
- Bytecode VM: ~14,823 ns/op
- **Speedup: 5.6x**

### When to Use Bytecode

**Use Bytecode When:**
- Performance is critical
- Running the same code multiple times
- CPU-bound computations (loops, math)
- Large scripts with many function calls

**Use AST Interpreter When:**
- Debugging (easier to trace)
- Development (faster iteration)
- Single-execution scripts
- Need maximum compatibility with original DWScript

## Current Limitations

1. **For Loops**: Not yet implemented in compiler
2. **Result Variable**: Function result assignment not yet supported
3. **External FFI**: External functions not yet wired to VM
4. **Some Advanced Features**: Interfaces, virtual methods, and some OOP features still in progress

## Future Enhancements

1. **Bytecode Serialization**: Save/load compiled bytecode (.dwc files)
2. **JIT Compilation**: LLVM-based JIT for even faster execution (5-10x)
3. **Advanced Optimizations**: Inlining, dead code elimination, constant propagation
4. **Debugging Support**: Breakpoints, step execution, variable inspection

## Implementation Details

### Value Representation

Values in the VM use a tagged union approach:

```go
type Value struct {
    Type ValueType  // nil, bool, int, float, string, array, object, function, closure, builtin
    Data interface{} // Actual value
}
```

This provides type safety while maintaining reasonable performance.

### Closure Implementation

Closures capture variables via upvalues, which can be open (pointing to stack) or closed (copied to heap):

```go
type Upvalue struct {
    location *Value  // Points to stack slot (open) or nil (closed)
    closed   Value   // Holds closed value
}
```

When a function returns, all upvalues pointing to its stack frame are closed.

### Error Handling

Runtime errors include stack traces with line/column information:

```
Bytecode runtime error: vm: undefined variable "foo"
test.dws [line: 5, column: 10]
Stack trace:
test.dws [line: 5, column: 10]
```

## Testing

The bytecode VM includes comprehensive tests:

- **Parity Tests** (`vm_parity_test.go`): Verify identical behavior to AST interpreter
- **Disassembler Tests**: Verify bytecode disassembly
- **Benchmarks**: Performance comparisons
- **Integration Tests**: End-to-end scenarios

Run tests:

```bash
# Run parity tests
go test ./internal/bytecode -run TestVMParity

# Run benchmarks
go test ./internal/bytecode -bench=BenchmarkVMVsInterpreter

# Run all tests
go test ./internal/bytecode
```

## See Also

- [docs/architecture/bytecode-vm-design.md](architecture/bytecode-vm-design.md) - Detailed design document
- [docs/architecture/bytecode-vm-quick-reference.md](architecture/bytecode-vm-quick-reference.md) - Quick reference
- [internal/bytecode/instruction.go](../internal/bytecode/instruction.go) - Instruction definitions
- [CLAUDE.md](../CLAUDE.md) - Development guide
