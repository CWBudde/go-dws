# Error Message Test Fixtures

This directory contains test fixtures demonstrating the improved error messages and debugging features in go-dws (Tasks 9.110-9.118).

## Overview

These fixtures showcase:
- **Type errors** with clear messages showing expected vs actual types
- **Runtime errors** with source code snippets and operand values
- **Exception handling** with complete stack traces
- **Error message formatting** with position information

## Test Fixtures

### Type Errors

1. **01_type_error_assignment.dws**
   - Demonstrates type mismatch in variable assignment
   - Shows: expected type, actual type, variable name, location

2. **02_type_error_parameter.dws**
   - Demonstrates type mismatch in function parameters
   - Shows: function name, parameter type mismatch

3. **03_type_error_undefined_variable.dws**
   - Demonstrates undefined variable errors
   - Shows: variable name, location where used

### Runtime Errors

4. **04_runtime_error_division_by_zero.dws**
   - Demonstrates division by zero runtime error
   - Shows: operand values, source code snippet

5. **05_runtime_error_with_stack_trace.dws**
   - Demonstrates runtime errors with full stack traces
   - Shows: complete call chain with positions

### Exception Handling

6. **06_exception_simple.dws**
   - Demonstrates basic exception raising
   - Shows: exception class, message, position

7. **07_exception_nested_calls.dws**
   - Demonstrates exceptions through nested function calls
   - Shows: complete stack trace with all function levels

8. **08_exception_handled.dws**
   - Demonstrates try-except exception handling
   - Shows: proper exception catching and stack trace access

### Combined

9. **09_combined_errors.dws**
   - Demonstrates multiple error types in one script
   - Shows: error handling patterns and recovery

## Running the Fixtures

### To see type errors:
```bash
./bin/dwscript run testdata/error_messages/01_type_error_assignment.dws
```

Expected output: Compilation error with detailed type mismatch information.

### To see runtime errors:
```bash
./bin/dwscript run testdata/error_messages/05_runtime_error_with_stack_trace.dws
```

Expected output: Runtime error with source snippet and stack trace.

### To see exception handling:
```bash
./bin/dwscript run testdata/error_messages/08_exception_handled.dws
```

Expected output: Successful execution with caught exceptions logged.

### To see combined example:
```bash
./bin/dwscript run testdata/error_messages/09_combined_errors.dws
```

Expected output: Successful execution with handled errors.

## Error Message Features

### Type Errors
- Include variable names and types
- Show expected vs actual types
- Display line and column information
- Provide clear, actionable messages

### Runtime Errors
- Show runtime values of operands
- Include source code snippets
- Highlight error position with `^`
- Display in color (when terminal supports it)

### Stack Traces (Tasks 9.113, 9.114, 9.116)
- Complete call chain from error point to top
- Function name, line, and column for each frame
- Available as:
  - Formatted string via `GetStackTrace()`
  - Structured array via `GetCallStack()`
  - Automatic display for unhandled exceptions

## Error Message Format

All errors follow a consistent format:

```
Error in <file>:<line>:<column>
  <line> | <source code>
           <highlighting>
<error message>
```

Example:
```
Error in script.dws:12:15
  12 | count := price;
                    ^
Cannot assign Float to Integer variable 'count'
```

## Comparison: Before vs After

### Before (Simple errors):
```
Type mismatch
```

### After (Enhanced errors):
```
Error in script.dws:12:15
  12 | count := price;
                    ^
Cannot assign Float to Integer variable 'count'
  Expected: Integer
  Actual: Float
```

## Using in Tests

These fixtures can be used for:
1. Manual testing during development
2. Automated regression tests
3. Documentation and examples
4. Comparison with original DWScript error messages

## Related Documentation

- `docs/error-messages.md` - Complete error message format documentation
- `docs/exception-enhancements-summary.md` - Exception features documentation
- `docs/debugging-information-summary.md` - Debugging features documentation
