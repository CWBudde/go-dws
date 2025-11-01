# Lambda Expressions in go-dws

Lambda expressions (also called anonymous functions or closures) allow you to define inline functions without declaring them separately. This document describes lambda support in go-dws.

## Table of Contents

- [Basic Syntax](#basic-syntax)
- [Shorthand Syntax](#shorthand-syntax)
- [Parameters and Return Types](#parameters-and-return-types)
- [Closures and Variable Capture](#closures-and-variable-capture)
- [Nested Lambdas](#nested-lambdas)
- [Higher-Order Functions](#higher-order-functions)
- [Limitations and Known Issues](#limitations-and-known-issues)
- [Implementation Status](#implementation-status)

## Basic Syntax

### Full Syntax

Define a lambda using the `lambda` keyword with a `begin/end` block:

```pascal
var double := lambda(x: Integer): Integer begin
  Result := x * 2;
end;

PrintLn(IntToStr(double(5)));  // Output: 10
```

### Procedure Lambda

Lambdas without a return type act as procedures:

```pascal
var printMessage := lambda(msg: String) begin
  PrintLn('Message: ' + msg);
end;

printMessage('Hello!');  // Output: Message: Hello!
```

## Shorthand Syntax

For simple expressions, use the arrow `=>` syntax:

```pascal
var triple := lambda(x: Integer): Integer => x * 3;

PrintLn(IntToStr(triple(4)));  // Output: 12
```

The shorthand form automatically wraps the expression in a `begin/end` block with `Result :=`.

## Parameters and Return Types

### Multiple Parameters

**NOTE**: Current implementation uses comma separators (BUG - see Known Issues):

```pascal
var add := lambda(a: Integer, b: Integer): Integer => a + b;
PrintLn(IntToStr(add(3, 7)));  // Output: 10
```

### No Parameters

```pascal
var getValue := lambda(): Integer => 42;
PrintLn(IntToStr(getValue()));  // Output: 42
```

### Type Inference

Return types can be inferred from the lambda body:

```pascal
var square := lambda(x: Integer) => x * x;  // Return type inferred as Integer
```

**Parameter type inference is not yet implemented** - all parameter types must be explicitly specified.

## Closures and Variable Capture

Lambdas can capture and use variables from their enclosing scope:

### Simple Capture

```pascal
var factor := 10;
var multiply := lambda(x: Integer): Integer => x * factor;

PrintLn(IntToStr(multiply(5)));  // Output: 50
```

### Reference Semantics

Captured variables use reference semantics - changes inside the lambda affect the outer scope:

```pascal
var counter := 0;
var increment := lambda() begin
  counter := counter + 1;
end;

increment();
increment();
PrintLn(IntToStr(counter));  // Output: 2
```

### Multiple Lambdas Sharing Variables

Multiple lambdas can capture and modify the same variable:

```pascal
var shared := 100;

var add10 := lambda() begin
  shared := shared + 10;
end;

var subtract5 := lambda() begin
  shared := shared - 5;
end;

add10();        // shared = 110
subtract5();    // shared = 105
```

## Nested Lambdas

Lambdas can return other lambdas, enabling advanced patterns:

### Factory Pattern

```pascal
type TAdder = function(y: Integer): Integer;

function makeAdder(x: Integer): TAdder;
begin
  Result := lambda(y: Integer): Integer => x + y;
end;

var add5 := makeAdder(5);
var add10 := makeAdder(10);

PrintLn(IntToStr(add5(3)));   // Output: 8
PrintLn(IntToStr(add10(3)));  // Output: 13
```

### Counter Pattern

```pascal
type TCounter = function(): Integer;

function MakeCounter(): TCounter;
var
  count: Integer;
begin
  count := 0;
  Result := lambda(): Integer begin
    count := count + 1;
    Result := count;
  end;
end;

var counter := MakeCounter();
PrintLn(IntToStr(counter()));  // Output: 1
PrintLn(IntToStr(counter()));  // Output: 2
PrintLn(IntToStr(counter()));  // Output: 3
```

## Higher-Order Functions

go-dws provides built-in higher-order functions that work with lambdas:

### Map

Transform each element of an array:

```pascal
// NOTE: Array literal syntax not yet implemented - see Known Issues
var doubled := Map(numbers, lambda(x: Integer): Integer => x * 2);
```

### Filter

Filter array elements by a predicate:

```pascal
var evens := Filter(numbers, lambda(x: Integer): Boolean => (x mod 2) = 0);
```

### Reduce

Reduce an array to a single value:

```pascal
var sum := Reduce(numbers, lambda(acc: Integer, x: Integer): Integer => acc + x, 0);
```

### ForEach

Execute a lambda for each element (side effects):

```pascal
ForEach(items, lambda(item: String) begin
  PrintLn('- ' + item);
end);
```

## Limitations and Known Issues

### Critical Bugs

1. **Lambda parameters use comma separators instead of semicolons** (Task 9.239)
   - Current: `lambda(x: Integer, y: Integer): Integer`
   - DWScript standard: `lambda(x: Integer; y: Integer): Integer`
   - **This breaks DWScript compatibility and must be fixed**

### Missing Parser Features

2. **Inline function pointer types not supported** (Tasks 9.240-9.241)
   - Cannot use: `var f: function(x: Integer): Integer;`
   - Must use: `type TFunc = function(x: Integer): Integer; var f: TFunc;`

3. **`array of TypeName` syntax not supported** (Tasks 9.242-9.243)
   - Cannot declare: `var arr: array of Integer;`
   - Cannot use: `procedure PrintArray(arr: array of Integer);`

4. **Array literals interpreted as set literals** (Task 9.244)
   - Cannot use: `var nums := [1, 2, 3, 4, 5];`
   - Workaround: Use `SetLength()` and manual assignment

### Missing Built-in Functions

5. **BoolToStr not implemented** (Task 9.245)
   - Workaround: Use `if/then/else` to print booleans

See `PLAN.md` tasks 9.239-9.245 for detailed information and workarounds.

## Implementation Status

### Completed Features

- ✅ Lambda expression parsing (shorthand and full syntax)
- ✅ Lambda evaluation and execution
- ✅ Closure capture with reference semantics
- ✅ Nested lambdas and multi-level capture
- ✅ Function pointer type compatibility
- ✅ Higher-order functions (Map, Filter, Reduce, ForEach)
- ✅ Return type inference from body
- ✅ Integration tests and documentation

### Not Yet Implemented

- ❌ Parameter type inference from context (Task 9.234-9.238)
- ❌ Semicolon parameter separators (Task 9.239 - **CRITICAL BUG**)
- ❌ Inline function pointer types (Tasks 9.240-9.241)
- ❌ Dynamic array syntax (Tasks 9.242-9.244)

### Test Coverage

- **Unit tests**: `internal/interp/lambda_test.go` (986 lines, 30+ test cases)
- **Integration tests**: `cmd/dwscript/lambda_integration_test.go`
- **Test scripts**: `testdata/lambdas/` (basic_lambda, closure, nested_lambda)

All tests pass with current workarounds.

## Examples

See `testdata/lambdas/` for comprehensive examples:

- `basic_lambda.dws` - Lambda creation, calling, types, reassignment
- `closure.dws` - Variable capture, mutations, reference semantics
- `nested_lambda.dws` - Factory patterns, counters, accumulators
- `higher_order.dws` - Map, Filter, Reduce, ForEach usage (requires array literal fix)

## See Also

- **Function Pointers**: `docs/function-pointers.md` (if exists)
- **Built-in Functions**: `docs/builtins.md`
- **PLAN.md**: Tasks 9.128-9.237 for implementation details and roadmap
