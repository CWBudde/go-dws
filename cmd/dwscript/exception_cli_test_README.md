# Exception Handling CLI Integration Tests

This file (`exception_cli_test.go`) contains comprehensive CLI integration tests for exception handling functionality (Task 8.226).

## Test Structure

### TestExceptionHandlingIntegration

Main integration test that runs exception test scripts and verifies their output.

**Tests 11 exception scripts:**
- 5 comprehensive test scripts (basic_try_except, try_finally, nested_exceptions, exception_propagation, raise_reraise)
- 6 ported test scripts from original DWScript test suite

**Verifies:**
- Exception messages appear in output
- Finally blocks execute (via `testFinallyRun` flag)
- Correct number of PASS markers in comprehensive tests
- No FAIL markers in successful tests
- Unhandled exceptions exit with error status

### TestExceptionMessages

Tests that exception messages are properly displayed in output.

**Verifies:**
- Custom exception messages appear in output
- Exception type names (ClassName) appear in output
- Multiple exception messages in nested handlers work correctly

### TestUnhandledExceptionStackTrace

Tests that unhandled exceptions show stack traces in stderr.

**Verifies:**
- Runtime error messages appear in stderr
- Exception messages appear in stderr
- Function names appear in stack trace for propagated exceptions
- Stack trace shows complete call chain (Level1 -> Level2 -> Level3)

## Running the Tests

```bash
# Run all exception CLI tests
cd cmd/dwscript
go test -v -run TestException

# Run specific test
go test -v -run TestExceptionHandlingIntegration
go test -v -run TestExceptionMessages
go test -v -run TestUnhandledExceptionStackTrace

# Run just the basic try-except test
go test -v -run "TestExceptionHandlingIntegration/Basic_Try-Except"
```

## Current Status

These tests are implemented and ready but currently **skip or fail** due to incomplete exception handling implementation:

### Known Issues Blocking Tests

1. **Parser Issue** - `class(Exception)` treated as interface, not parent class
   - See `testdata/exceptions/README.md` for details
   - Fix needed in `parser/classes.go` lines 83-99

2. **Missing Parser Features**
   - `new` keyword not supported
   - Bare `except` (catch-all) not supported
   - `else` clause in exception handlers not supported

3. **Missing Semantic Analysis**
   - Exception type checking incomplete
   - ClassName property not available on Exception class

4. **Missing Interpreter Features**
   - Exception propagation not fully implemented
   - Stack trace generation for unhandled exceptions
   - ExceptObject built-in variable

## Test Coverage

The CLI integration tests cover:

### Exception Catching (Task 8.226 item 1)
- Basic try-except blocks
- Specific exception type matching
- Multiple exception handlers
- Exception type inheritance
- Exception propagation

### Exception Messages (Task 8.226 item 2)
- Custom error messages
- Exception type names
- Multiple exceptions in nested handlers
- Message property access

### Finally Block Execution (Task 8.226 item 3)
- Finally executes on normal completion
- Finally executes on exception
- Finally executes before exception propagates
- Nested finally blocks
- Finally with reraise

### Stack Traces (Task 8.226 item 4)
- Unhandled exceptions show runtime error
- Exception messages in stderr
- Function names in call stack
- Complete call chain displayed

## Test Expectations

Each test has:
- `wantOutputs`: Strings that must appear in stdout
- `wantPasses`: Number of PASS markers expected (for comprehensive tests)
- `wantInStderr`: Strings that must appear in stderr (for unhandled exceptions)
- `shouldFail`: Whether script should exit with error
- `testFinallyRun`: Whether to verify finally blocks executed

## Integration with CI/CD

Once exception handling is fully implemented, these tests should be:
1. Uncommented in CI/CD pipeline
2. Run as part of integration test suite
3. Required to pass before merging exception handling PRs

## Future Enhancements

Potential additional tests:
- Exception in finally block
- Multiple exception types in hierarchy
- Exception with custom properties
- Performance tests (exception overhead)
- Memory tests (exception cleanup)
