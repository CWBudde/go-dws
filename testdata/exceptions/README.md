# Exception Handling Test Scripts

This directory contains exception handling test scripts from two sources:
1. **Ported Tests** - Tests ported from the original DWScript test suite
2. **Comprehensive Tests** - Systematic tests covering all exception handling features

## Test Organization

### Comprehensive Tests

These are systematic, comprehensive tests written specifically for go-dws:

| Test File | Description |
|-----------|-------------|
| `basic_try_except.dws` | Comprehensive test for basic try-except blocks: exception catching, type matching, handler selection, exception object properties |
| `try_finally.dws` | Comprehensive test for try-finally blocks: finally execution guarantee, finally with exceptions, finally with normal flow |
| `nested_exceptions.dws` | Comprehensive test for nested exception handling: nested try-except, exception in handler, ExceptObject scoping |
| `exception_propagation.dws` | Comprehensive test for exception propagation through call stack: propagation through functions, stack unwinding, finally during propagation |
| `raise_reraise.dws` | Comprehensive test for raise and re-raise statements: basic raise, re-raise in handler, re-raise preserves exception |

### Ported Tests

Tests ported from `reference/dwscript-original/Test/SimpleScripts/`:

| Test File | Source | Description |
|-----------|--------|-------------|
| `basic_exceptions.dws` | `exceptions.pas` | Basic exception handling with specific exception types, exception type matching, catch-all handlers, and nested try-finally-except blocks |
| `re_raise.dws` | `exceptions2.pas` | Re-raising exceptions with bare `raise` statement, demonstrating exception propagation through multiple handlers |
| `try_except_finally.dws` | `try_except_finally.pas` | Combined try-except-finally blocks showing that finally clauses execute regardless of exceptions |
| `except_object.dws` | `exceptobj.pas` | The `ExceptObject` built-in variable that provides access to the current exception in handlers |
| `nested_calls.dws` | `exception_nested_call.pas` | Exception propagation through nested function calls, uncaught exceptions causing runtime errors |
| `exception_scoping.dws` | `exception_scoping.pas` | Exception variable scoping - demonstrates that exception handler variables shadow outer variables |
| `break_in_except.dws` | `break_in_except_block.pas` | Break statements inside exception handlers within loops, showing control flow interaction |

## Test Coverage

These tests cover the following exception handling features:

### Basic Exception Handling
- `try...except...end` blocks
- `on E: ExceptionType do` specific exception handlers
- `else` clause for catch-all handling
- Multiple exception handlers with type matching

### Exception Types
- Base `Exception` class
- Custom exception types (inheritance from `Exception`)
- `EDelphi` standard exception type
- Exception type hierarchy and matching rules

### Advanced Features
- `try...finally...end` blocks
- Combined `try...except...finally...end`
- Exception re-raising with bare `raise`
- `ExceptObject` built-in variable
- Exception propagation through call stack
- Variable scoping in exception handlers
- Control flow statements (`break`) in exception handlers

### Exception Properties
- `Message` property - exception message text
- `ClassName` property - runtime type name

### Stack Unwinding
- Exception propagation through nested `try` blocks
- Exception propagation through function calls
- Finally clauses guaranteed to execute during unwinding

## Expected Output Format

Each `.dws` test file has a corresponding `.txt` file containing the expected output. The format is:

```
[Errors >>>>]
[compiler warnings/errors if any]
[Result >>>>]
[actual program output]
```

If there are no compiler errors/warnings, only the result output is present.

## Testing These Scripts

To run these tests with the go-dws interpreter:

```bash
# Run a single test
./bin/dwscript run testdata/exceptions/basic_exceptions.dws

# Run all exception tests (from test suite)
go test ./interp -run Exception
```

## Implementation Status

These tests require the following features to be fully implemented:

- [x] Exception class hierarchy
- [x] Try-except-finally parsing
- [x] Raise statement
- [x] Exception handler matching
- [x] Stack unwinding
- [x] ExceptObject built-in
- [x] Re-raise support
- [x] Finally guarantee
- [x] Exception propagation
- [x] Integration with control flow
- [x] Unit tests
- [x] Port test scripts

## Known Issues

### Parser Issue with Exception Class

The parser currently has a limitation where it assumes parent classes must start with 'T' (parser/classes.go:91). When parsing `class(Exception)`, it incorrectly treats "Exception" as an interface instead of a parent class because "Exception" doesn't follow the 'T' prefix convention.

**Error message:**

```
interface 'Exception' not found
```

**Workaround:** Until the parser is fixed to handle built-in exception classes, you can:
1. Use `TException` as the base class name (requires updating the interpreter's built-in registration)
2. Or fix the parser to recognize certain built-in classes like "Exception" as parent classes

**Parser fix needed in:** `parser/classes.go` lines 83-99 - The logic needs to be updated to:
- Check if the identifier is a known built-in class (Exception, TObject, etc.)
- Or use a more sophisticated heuristic than just checking the first character

### Missing Parser Features

Several exception handling features are not yet supported by the parser:

1. **`new` keyword** - `raise new Exception('msg')` syntax is not supported
   - Tests use `Exception.Create('msg')` instead
   - Parser needs to add NEW token and prefix parse function

2. **Bare `except`** - Catch-all exception handlers without `on` clause
   - Semantic analyzer expects all except handlers to specify a type
   - Parser and analyzer need updates

3. **`else` clause in exception handlers** - For unmatched exceptions
   - Parser treats `else` as unexpected token in except blocks
   - Grammar needs extension

4. **Missing EDelphi exception type** - Referenced in `exception_scoping.dws`
   - Need to add to built-in exception registration in `interp/exceptions.go`

## Notes

- All tests use DWScript's `:=` assignment operator and Pascal syntax
- Exception types are classes, not interfaces
- The `new` keyword is optional when creating exceptions (both `raise Exception.Create(...)` and `raise new Exception(...)` are valid)
- Exception handler order matters - first matching handler wins
- Finally blocks execute even when exceptions are raised or re-raised
- `ExceptObject` is only non-nil inside exception handlers
