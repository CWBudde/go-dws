# Contracts (Design by Contract)

## Overview

DWScript supports **Design by Contract** (DbC), a programming methodology that allows you to specify preconditions and postconditions for functions and procedures. Contracts help ensure program correctness by documenting and enforcing assumptions and guarantees.

## Features

- **Preconditions** (`require`): Conditions that must be true when a function is called
- **Postconditions** (`ensure`): Conditions that must be true when a function returns
- **`old` keyword**: Reference pre-execution values in postconditions
- **Custom error messages**: Optional descriptive messages for contract failures
- **Full DWScript compatibility**: 100% compatible with original DWScript contract syntax

## Syntax

### Preconditions

Preconditions are specified using the `require` keyword after the function signature and before the function body:

```pascal
function Divide(a, b : Float) : Float;
require
   b <> 0 : 'divisor cannot be zero';
begin
   Result := a / b;
end;
```

Multiple preconditions can be specified, separated by semicolons:

```pascal
function Clamp(value, min, max : Integer) : Integer;
require
   min <= max : 'min must be <= max';
   min >= 0;
   max <= 100;
begin
   // ...
end;
```

### Postconditions

Postconditions are specified using the `ensure` keyword **after** the function's `end` keyword:

```pascal
function Increment(x : Integer) : Integer;
begin
   Result := x + 1;
end;
ensure
   Result > old x : 'result must be greater than input';
```

### The `old` Keyword

The `old` keyword allows you to reference the value of a parameter or variable **before** the function executed. This is essential for postconditions that need to compare the result with the input.

```pascal
function Increment(x : Integer) : Integer;
begin
   Result := x + 1;
end;
ensure
   Result = old x + 1;  // old x is the value of x when function was called
```

**Important**: The `old` keyword can only be used in postconditions (`ensure` blocks), not in preconditions or function bodies.

### Custom Error Messages

Both preconditions and postconditions can include optional custom error messages using the `: 'message'` syntax:

```pascal
function Factorial(n : Integer) : Integer;
require
   n >= 0 : 'factorial requires non-negative input';
begin
   // ...
end;
ensure
   Result > 0 : 'factorial result must be positive';
```

## Examples

### Example 1: Simple Precondition

```pascal
procedure TestPositive(i : Integer);
require
   i > 0 : 'input must be positive';
begin
   PrintLn('Processing ' + IntToStr(i));
end;

begin
   TestPositive(10);  // OK
   TestPositive(-5);  // Runtime error: Pre-condition failed in TestPositive [line: 3, column: 4], input must be positive
end.
```

### Example 2: Postcondition with `old`

```pascal
function Double(x : Integer) : Integer;
begin
   Result := x * 2;
end;
ensure
   Result = old x * 2;

begin
   PrintLn(Double(5));  // Prints: 10
end.
```

### Example 3: Multiple Conditions

```pascal
function Clamp(value, min, max : Integer) : Integer;
require
   min <= max : 'min must be less than or equal to max';
   min >= 0 : 'min must be non-negative';
   max <= 100 : 'max must not exceed 100';
begin
   if value < min then
      Result := min
   else if value > max then
      Result := max
   else
      Result := value;
end;
ensure
   Result >= min : 'result must be at least min';
   Result <= max : 'result must be at most max';

begin
   PrintLn(Clamp(50, 0, 100));   // 50
   PrintLn(Clamp(-10, 0, 100));  // 0 (clamped to min)
   PrintLn(Clamp(150, 0, 100));  // 100 (clamped to max)
end.
```

### Example 4: Recursive Functions

Contracts work seamlessly with recursive functions:

```pascal
function Factorial(n : Integer) : Integer;
require
   n >= 0 : 'factorial requires non-negative input';
begin
   if n <= 1 then
      Result := 1
   else
      Result := n * Factorial(n - 1);
end;
ensure
   Result > 0 : 'factorial result must be positive';

begin
   PrintLn(Factorial(5));   // 120
   PrintLn(Factorial(10));  // 3628800
end.
```

## Error Messages

When a contract is violated, a `ContractFailureError` is raised with a detailed error message:

**Format**: `Pre-condition failed in FunctionName [line: N, column: M], message`

Example:
```
Pre-condition failed in Divide [line: 3, column: 4], divisor cannot be zero
Post-condition failed in Increment [line: 7, column: 4], Result = old x + 1
```

The error message includes:
- Contract type (Pre-condition or Post-condition)
- Function name
- Line and column number of the failing condition
- Custom message (if provided) or the condition expression

## Implementation Details

### Execution Order

1. **Preconditions** are evaluated **before** the function body executes
2. If any precondition fails, the function body does not execute
3. **Old values** are captured before the function body executes (for postconditions)
4. The **function body** executes
5. **Postconditions** are evaluated **after** the function body completes
6. If any postcondition fails, a contract failure error is raised

### Old Value Capture

The `old` keyword works by capturing values before function execution:

```pascal
function Increment(x : Integer) : Integer;
begin
   Result := x + 1;  // x is modified here (conceptually)
end;
ensure
   Result = old x + 1;  // old x refers to x's value at entry
```

The interpreter automatically identifies all `old` expressions in postconditions and captures those values before the function executes.

### Nested Calls

Contracts are checked for every function call, including nested calls:

```pascal
function Double(x : Integer) : Integer;
require
   x >= 0;
begin
   Result := x * 2;
end;
ensure
   Result = old x * 2;

function Quadruple(x : Integer) : Integer;
begin
   Result := Double(Double(x));  // Both calls check contracts
end;
```

## Limitations

### Current Limitations

1. **Exception handling not implemented**: You cannot currently catch contract failures with `try/except` blocks (exception handling is planned for a future stage)
2. **Var parameters**: The `old` keyword with var parameters is implemented but var parameters themselves have a known issue where modifications don't persist to the caller
3. **Class method contracts**: Method contract inheritance (Liskov substitution principle) is not yet implemented - this requires Stage 7 (OOP) completion

### Future Enhancements

The following features are planned:

1. **Method contract inheritance** (Stage 7): Overridden methods will inherit base class contracts
2. **Invariants** (Stage 8): Class invariants that must hold before and after every public method call
3. **Exception integration**: Proper try/catch support for contract failures

## Best Practices

### 1. Use Descriptive Messages

Always provide custom error messages for complex conditions:

```pascal
// Good
require
   balance >= amount : 'insufficient funds for withdrawal';

// Less helpful
require
   balance >= amount;
```

### 2. Keep Conditions Simple

Conditions should be side-effect free and efficient:

```pascal
// Good
require
   x > 0;

// Bad - expensive computation
require
   ExpensiveCheck(x) = true;
```

### 3. Use `old` for State Comparisons

Postconditions often need to compare final state with initial state:

```pascal
function Increment(x : Integer) : Integer;
begin
   Result := x + 1;
end;
ensure
   Result = old x + 1;  // Correct
   // NOT: Result = x + 1  (x might have been modified)
```

### 4. Separate Concerns

Use preconditions for input validation and postconditions for output guarantees:

```pascal
function SafeDivide(a, b : Float) : Float;
require
   b <> 0;              // INPUT validation
begin
   Result := a / b;
end;
ensure
   Result * old b = old a;  // OUTPUT verification
```

## Testing

Contract tests are located in `testdata/contracts/`:

- `procedure_contracts.dws`: Basic pre/postconditions
- `contracts_old.dws`: `old` keyword usage
- `contracts_code.dws`: Precondition validation
- `recursive_factorial.dws`: Contracts with recursion
- `nested_calls.dws`: Contracts with nested function calls
- `multiple_conditions.dws`: Multiple pre/postconditions

Run tests with:
```bash
./bin/dwscript run testdata/contracts/procedure_contracts.dws
```

## References

- [Design by Contract (Wikipedia)](https://en.wikipedia.org/wiki/Design_by_contract)
- [DWScript Language Reference](https://www.delphitools.info/dwscript/)
- PLAN.md: Contract implementation tasks (lines 425-627)
