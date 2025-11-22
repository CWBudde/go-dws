# Control Flow Test Scripts

This directory contains comprehensive test scripts for DWScript's loop control flow statements: `break`, `continue`, and `exit`.

## Test Files

### break_statement.dws
Tests the `break` statement functionality:
- Breaking from `for` loops
- Breaking from `while` loops
- Breaking from `repeat` loops
- Breaking from nested loops (verifies only innermost loop is affected)
- Conditional breaking

**Expected output:** `break_statement.txt`

### continue_statement.dws
Tests the `continue` statement functionality:
- Continuing in `for` loops
- Continuing in `while` loops
- Continuing in `repeat` loops
- Multiple continue conditions
- Continue in nested loops (verifies only innermost loop is affected)

**Expected output:** `continue_statement.txt`

### exit_statement.dws
Tests the `exit` statement functionality:
- Early function termination
- Exit preserving `Result` variable value
- Exit in procedures (no return value)
- Exit in nested function calls (verifies each function handles exit independently)
- Exit from loops within functions
- Multiple exit points in a function

**Expected output:** `exit_statement.txt`

### nested_loops.dws
Comprehensive tests for break/continue in nested loop structures:
- Break only exits innermost loop
- Continue only affects innermost loop
- Triple nested loops with break
- Break and continue combined in same structure
- Mixed loop types (for/while/repeat) with break/continue

**Expected output:** `nested_loops.txt`

## Running Tests

### Via CLI
```bash
./bin/dwscript run testdata/control_flow/break_statement.dws
./bin/dwscript run testdata/control_flow/continue_statement.dws
./bin/dwscript run testdata/control_flow/exit_statement.dws
./bin/dwscript run testdata/control_flow/nested_loops.dws
```

### Via Go Tests
```bash
go test ./cmd/dwscript -run TestControlFlow -v
```

## Test Coverage

These tests verify:
- ✓ Break exits only the innermost loop
- ✓ Continue skips to next iteration of innermost loop only
- ✓ Exit terminates current function/procedure immediately
- ✓ Exit preserves `Result` variable value
- ✓ Exit at program level terminates the program
- ✓ Control flow statements work correctly in nested structures
- ✓ Control flow statements work with all loop types (for, while, repeat)
- ✓ Control flow statements interact correctly with exception handling

## Implementation Details

### Break Statement
- **Semantic validation:** Must be inside a loop, not in finally block
- **Runtime behavior:** Sets `breakSignal`, unwinding stack until loop catches it
- **Scope:** Affects only the innermost enclosing loop

### Continue Statement
- **Semantic validation:** Must be inside a loop, not in finally block
- **Runtime behavior:** Sets `continueSignal`, loop catches it and continues to next iteration
- **Scope:** Affects only the innermost enclosing loop

### Exit Statement
- **Semantic validation:** Exit with value not allowed at program level, not in finally block
- **Runtime behavior:** Sets `exitSignal`, function returns immediately with current `Result` value
- **Scope:** Terminates current function/procedure or program (if at top level)

## Related Documentation
- See `docs/control-flow.md` for detailed language specification
