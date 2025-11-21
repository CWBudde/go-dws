# Var Parameters Design Document

> Task 1.3.6 - Design Var Parameter Language Feature
>
> Date: 2025-11-21

## Executive Summary

Var parameters (by-reference parameters) allow functions/procedures to modify the caller's variables directly. This document analyzes the current implementation status and provides design guidance for completing var parameter support, particularly in the bytecode compiler/VM.

## Current Implementation Status

### Fully Implemented Layers

| Layer | Status | Details |
|-------|--------|---------|
| Lexer | COMPLETE | `VAR` keyword tokenized as `token.VAR` |
| Parser | COMPLETE | Parses `var` modifier, sets `Parameter.ByRef = true` |
| AST | COMPLETE | `Parameter.ByRef bool` field in `pkg/ast/functions.go:37` |
| Semantic Analyzer | COMPLETE | Handles `param.ByRef` in `internal/semantic/analyze_functions.go` |
| Interpreter | COMPLETE | `ReferenceValue` type, full function call handling |
| FFI | COMPLETE | `VarParams []bool` in function signatures |
| Printer | COMPLETE | Correctly prints `var` prefix |

### Not Implemented

| Layer | Status | Details |
|-------|--------|---------|
| Bytecode Compiler | NOT IMPLEMENTED | Ignores `param.ByRef`, declares all params as locals |
| Bytecode VM | NOT IMPLEMENTED | No reference/pointer opcodes |

## Language Semantics

### Syntax

```pascal
// Procedure with var parameter
procedure Increment(var x: Integer);
begin
  x := x + 1;
end;

// Function with mixed parameters
function UpdateAndSum(a: Integer; var b: Integer): Integer;
begin
  b := b * 2;
  Result := a + b;
end;

// Lambda with var parameter
var proc := procedure(var s: String) begin s := 'modified'; end;
```

### Semantics

1. **Pass by Reference**: Var parameters receive a reference to the caller's variable, not a copy
2. **Modifications Visible**: Any assignment to the parameter modifies the original variable
3. **Variable Required**: Caller must pass a variable (lvalue), not a literal or expression result
4. **Type Matching**: Parameter type must match argument type exactly (no implicit conversion)

### Usage Examples

```pascal
// Basic increment
var n := 5;
Increment(n);  // n becomes 6

// Swap function
procedure Swap(var a, b: Integer);
var temp: Integer;
begin
  temp := a;
  a := b;
  b := temp;
end;

var x := 1;
var y := 2;
Swap(x, y);  // x=2, y=1

// Modifying records
procedure IncX(var r: TRec);
begin
  r.X += 1;
end;
```

## Current Implementation Details

### AST Node (pkg/ast/functions.go)

```go
type Parameter struct {
    DefaultValue Expression
    Name         *Identifier
    Type         TypeExpression
    Token        token.Token
    EndPos       token.Position
    IsLazy       bool
    ByRef        bool      // true for var parameters
    IsConst      bool
}
```

### Parser (internal/parser/functions.go)

The parser correctly handles var parameters:

```go
// Line 524 - Detect var keyword
if p.peekTokenIs(token.VAR) {
    p.nextToken()
    byRef = true
}

// Line 644 - Set ByRef flag
param := &ast.Parameter{
    Token:        tok,
    Name:         name,
    Type:         paramType,
    ByRef:        byRef,
    // ...
}
```

### Interpreter (internal/interp/value.go)

The `ReferenceValue` type provides reference semantics:

```go
type ReferenceValue struct {
    Env     *Environment  // Environment containing the variable
    VarName string        // Name of the referenced variable
}

func (r *ReferenceValue) Dereference() (Value, error)  // Read through reference
func (r *ReferenceValue) Assign(value Value) error     // Write through reference
```

### Interpreter Function Calls (internal/interp/functions_calls.go)

Current implementation at lines 232-257:

```go
} else if isByRef {
    // For var parameters, create a reference to the variable
    if ident, ok := arg.(*ast.Identifier); ok {
        // Check if the variable is already a reference (var parameter passed through)
        if existingRef, ok := argCache[argIndex].(*ReferenceValue); ok {
            preparedArgs = append(preparedArgs, existingRef)
        } else {
            // Create a new reference to the variable
            ref := &ReferenceValue{
                Env:     i.env,
                VarName: ident.Value,
            }
            preparedArgs = append(preparedArgs, ref)
        }
    } else {
        // Var parameter must be a variable reference
        return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
    }
}
```

### Built-in Functions with Var Parameters

The following builtins require var parameter support (internal/interp/functions_builtins.go:238-266):

- `Inc(var x)` - Increment integer
- `Dec(var x)` - Decrement integer
- `Insert(source, var target, pos)` - Insert string into string
- `Delete(var s, index, count)` - Delete characters from string
- `DecodeDate(date, var y, m, d)` - Extract date components
- `DecodeTime(time, var h, m, s, ms)` - Extract time components
- `Swap(var a, var b)` - Swap two values
- `DivMod(dividend, divisor, var quotient, var remainder)` - Division with modulo
- `TryStrToInt(s, var value)` - Try parse integer
- `TryStrToFloat(s, var value)` - Try parse float
- `SetLength(var arr, newLen)` - Resize dynamic array

## Bytecode Compiler Design

### Current Gap

The bytecode compiler (internal/bytecode/compiler_statements.go:449-454) ignores `ByRef`:

```go
for _, param := range fn.Parameters {
    // MISSING: Check param.ByRef
    paramType := typeFromAnnotation(param.Type)
    if _, err := child.declareLocal(param.Name, paramType); err != nil {
        // ...
    }
}
```

### Proposed Design

#### Option A: Reference Opcodes (Recommended)

Add new opcodes to handle references:

```go
// New opcodes in instruction.go
OpLoadRef    // Push reference to local variable onto stack
OpStoreRef   // Store value through reference on stack
OpDeref      // Dereference: replace reference on stack with its value
```

Compilation strategy:
1. Mark parameters as reference type in function metadata
2. On function call, push references for var params instead of values
3. When loading var param, auto-dereference for reading
4. When storing to var param, use OpStoreRef to write through reference

#### Option B: Parameter Metadata

Store var param flags in function metadata:

```go
type FunctionObject struct {
    Name     string
    Chunk    *Chunk
    Arity    int
    VarParams []bool  // NEW: Which parameters are by-reference
}
```

VM changes:
1. During CALL, check VarParams to determine how to pass arguments
2. Store references in local slots for var params
3. On function return, references are automatically updated

### Recommended Implementation: Option A

Option A (Reference Opcodes) is cleaner because:
1. Explicit reference semantics visible in bytecode
2. Easier to debug/trace
3. Consistent handling across all reference types (var params, object fields, etc.)
4. Can be extended for other reference features later

### Bytecode Examples

Before (current - var params broken):
```
; procedure Increment(var x: Integer)
; x := x + 1
LOAD_LOCAL 0      ; Load local slot 0 (x) - BUG: gets copy, not reference
PUSH_CONST 1
ADD
STORE_LOCAL 0     ; Store back to local - BUG: doesn't update caller
RETURN
```

After (with reference opcodes):
```
; procedure Increment(var x: Integer)
; x := x + 1
LOAD_LOCAL 0      ; Load reference from slot 0
DEREF             ; Dereference to get current value
PUSH_CONST 1
ADD
LOAD_LOCAL 0      ; Load reference again
STORE_REF         ; Store through reference - updates caller's variable
RETURN
```

Call site changes:
```
; Before: Increment(n)
LOAD_GLOBAL 0     ; Load value of n
CALL Increment

; After: Increment(n)
LOAD_REF 0        ; Load reference TO n (not value)
CALL Increment
```

## Type System Integration

### Semantic Analyzer Changes

The semantic analyzer (internal/semantic/analyze_functions.go:45) already tracks var params:

```go
if param.ByRef {
    varParams[i] = true
}
```

This populates `FunctionType.VarParams []bool` for type checking.

### Type Checking Rules

1. **Var param argument must be an lvalue** (variable, array element, record field)
2. **Type must match exactly** (no implicit conversions for var params)
3. **Cannot pass const to var param** (would violate const-ness)
4. **Cannot pass expression result to var param** (no memory location to reference)

Error messages:
```
Error: var parameter requires a variable, got literal
Error: cannot pass const value to var parameter
Error: type mismatch: expected Integer var parameter, got String
```

## Edge Cases

### 1. Nested Var Parameters

```pascal
procedure Inner(var x: Integer);
begin
  x := x + 1;
end;

procedure Outer(var y: Integer);
begin
  Inner(y);  // Pass var param through to another var param
end;

var n := 5;
Outer(n);  // n should be 6
```

Implementation: When argument is already a ReferenceValue, pass it through unchanged.

### 2. Var Parameters with Record Fields

```pascal
procedure IncField(var f: Integer);
begin
  f := f + 1;
end;

var rec: TMyRecord;
IncField(rec.Field);  // Reference to specific field
```

Implementation: Create field reference that knows both record and field name.

### 3. Var Parameters with Array Elements

```pascal
procedure Double(var x: Integer);
begin
  x := x * 2;
end;

var arr: array of Integer;
SetLength(arr, 3);
arr[0] := 5;
Double(arr[0]);  // arr[0] should be 10
```

Implementation: Create indexed reference that knows array and index.

### 4. Var Parameters in Function Pointers

```pascal
type TModifier = procedure(var x: Integer);

procedure Double(var x: Integer);
begin
  x := x * 2;
end;

var modifier: TModifier;
modifier := Double;
var n := 5;
modifier(n);  // n should be 10
```

Implementation: Function pointer metadata must include var param flags.

## Testing Strategy

### Unit Tests

1. Basic var param modification (integers, floats, strings, booleans)
2. Multiple var params in single function
3. Mixed var and value params
4. Nested var param passthrough
5. Record field as var param
6. Array element as var param
7. Function pointer with var param
8. Error cases (literal, const, type mismatch)

### Integration Tests

1. Swap function implementation
2. Inc/Dec builtin functions
3. TryStrToInt pattern (success flag + result)
4. DecodeDate/DecodeTime patterns
5. Round-trip: modify, return, verify

### Fixture Tests

Run existing fixtures that use var parameters:
- `testdata/fixtures/SimpleScripts/record_result3.pas`
- `testdata/fixtures/SimpleScripts/func_ptr_var.pas`
- `testdata/fixtures/SimpleScripts/stack_depth.pas`

## Implementation Plan

### Phase 1: Bytecode Infrastructure (1.3.7.1-1.3.7.3)

1. Add reference opcodes to `instruction.go`
2. Add `VarParams []bool` to `FunctionObject`
3. Update compiler to track var params in function metadata
4. Update disassembler to show reference opcodes

### Phase 2: Compiler Implementation (1.3.7.4)

1. Modify `compileCallExpression` to push references for var params
2. Modify `compileFunctionDeclaration` to handle var param locals
3. Implement reference load/store for var param access

### Phase 3: VM Implementation (1.3.7.5)

1. Implement `OpLoadRef`, `OpStoreRef`, `OpDeref` execution
2. Create `ReferenceValue` equivalent for VM (or reuse)
3. Update CALL to handle reference arguments

### Phase 4: Testing & Integration (1.3.7.6)

1. Unit tests for all reference operations
2. Integration tests with builtin functions
3. Run fixture tests in bytecode mode
4. Performance benchmarking

## Acceptance Criteria

- [ ] Parser correctly marks var parameters (DONE)
- [ ] Semantic analyzer validates var parameter usage (DONE)
- [ ] AST interpreter handles var parameters correctly (DONE)
- [ ] Bytecode compiler generates correct code for var parameters
- [ ] Bytecode VM executes var parameter functions correctly
- [ ] Built-in functions with var params work in bytecode mode
- [ ] All var parameter fixture tests pass
- [ ] Error messages are clear for invalid var param usage
- [ ] Performance is acceptable (no excessive copying)

## References

- Original DWScript documentation: https://www.delphitools.info/dwscript/
- Pascal var parameter semantics: Standard Pascal by-reference parameters
- Current implementation: `internal/interp/value.go:255-306`
- Test fixtures: `testdata/fixtures/SimpleScripts/`
