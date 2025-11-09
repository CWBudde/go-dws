# Error Message Format Documentation

**Tasks**: 9.110-9.118
**Status**: ✅ COMPLETE
**Date**: 2025-11-04

## Overview

This document describes the error message format and improvements implemented in go-dws. The enhanced error messages provide clear, actionable information to help developers quickly identify and fix issues.

## Error Message Structure

All errors in go-dws follow a consistent format with rich contextual information:

```
[Error Type] in <filename>:<line>:<column>
  <line-number> | <source code line>
                  <position marker>
<detailed error message>
[optional: additional context]
```

### Components

1. **Error Type**: Classification (Error, Runtime Error, etc.)
2. **Location**: Filename, line number, column number
3. **Source Context**: The actual line of code that caused the error
4. **Position Marker**: Visual indicator (`^`) pointing to the error location
5. **Detailed Message**: Clear explanation of what went wrong
6. **Additional Context**: Optional extra information (values, types, suggestions)

## Error Categories

### 1. Syntax Errors (Parser)

**When**: During parsing, when source code doesn't match DWScript grammar

**Format**:
```
Error in <file>:<line>:<column>
  <line> | <source code>
           ^
<syntax error message>
```

**Example**:
```
Error in script.dws:5:10
   5 | var x Integer := 42;
              ^
expected ':' after identifier, got 'Integer'
```

**Features**:
- Shows exactly where the parser got confused
- Indicates what was expected vs what was found
- Highlights the problematic token

### 2. Type Errors (Semantic Analysis)

**When**: During semantic analysis, when types don't match

**Format**:
```
Error in <file>:<line>:<column>
  <line> | <source code>
           ^
<type error message>
  Expected: <type>
  Actual: <type>
```

**Example**:
```
Error in script.dws:12:15
  12 | count := price;
                    ^
Cannot assign Float to Integer variable 'count'
  Expected: Integer
  Actual: Float
```

**Features**
- Variable names included in error messages
- Expected and actual types clearly shown
- Context-aware messages (assignment, parameter, return value, etc.)
- Suggestions when applicable

**Common Type Errors**:

| Error | Message Format |
|-------|----------------|
| Assignment mismatch | `Cannot assign <type> to <type> variable '<name>'` |
| Parameter mismatch | `Function '<name>' expects <type> as parameter N, got <type>` |
| Return type mismatch | `Function '<name>' returns <type>, expected <type>` |
| Undefined variable | `Undefined variable '<name>'` |
| Undefined function | `Undefined function '<name>'` |

### 3. Runtime Errors

**When**: During execution, when operations fail at runtime

**Format**:
```
Runtime Error: <error type>: <message> [line: N, column: M]
<source code snippet with highlighting>
  <value context>
```

**Example**
```
Runtime Error: Division by zero [line: 15, column: 12]
  15 | result := a div b;
                      ^
Division by zero: 10 / 0
  Left operand: 10
  Right operand: 0
```

**Features**:
- Shows runtime values of operands
- Source code snippet with position highlighting
- Color-coded output (red for errors)
- Contextual information about the operation

**Common Runtime Errors**:

| Error Type | Message Format | Values Shown |
|------------|----------------|--------------|
| Division by zero | `Division by zero: <left> / <right>` | Both operands |
| Array index out of bounds | `Index <N> out of bounds for array of length <M>` | Index, length |
| Nil dereference | `Nil object reference at <location>` | Object name |
| Type conversion failure | `Cannot convert <value> to <type>` | Value, types |

### 4. Exceptions

**When**: Raised explicitly via `raise` statement or by runtime errors

**Format**:
```
Runtime Error: <ExceptionClass>: <message> [line: N, column: M]
<stack trace>
```

**Example** (Tasks 9.113, 9.114):
```
Runtime Error: Exception: Error from deep in the call stack [line: 7, column: 3]
DeepFunction [line: 13, column: 3]
MiddleFunction [line: 19, column: 3]
TopFunction [line: 26, column: 1]
```

**Features**:
- Exception class name (Exception, EDivByZero, etc.)
- Exception message
- Position where exception was raised
- Complete stack trace with all function calls
- Each stack frame shows: function name, line, column

**Stack Trace Format**:
```
FunctionName [line: N, column: M]
```

Each frame is listed from most recent (where error occurred) to oldest (entry point).

## Color Coding

When output to a color-capable terminal, errors use ANSI color codes:

- **Red (`\x1b[1;31m`)**: Error messages, position markers
- **Reset (`\x1b[0m`)**: Back to normal text

Example colored output:
```
Error in script.dws:12:15
  12 | count := price;
                    ^        (in red)
Cannot assign Float to Integer variable 'count'  (in red)
```

## Error Message Improvements

### Type Error Messages

**Before**:
```
Type mismatch
```

**After**:
```
Error in script.dws:12:15
  12 | count := price;
                    ^
Cannot assign Float to Integer variable 'count'
  Expected: Integer
  Actual: Float
```

**Improvements**:
- Variable name included
- Both types shown
- Exact location
- Source code context

### Runtime Error Messages

**Before**:
```
ERROR: division by zero
```

**After**:
```
Runtime Error: Division by zero [line: 15, column: 12]
  15 | result := a div b;
                      ^
Division by zero: 10 / 0
  Left operand: 10
  Right operand: 0
```

**Improvements**:
- Operand values shown
- Source code snippet
- Position highlighting
- Visual error marker

### Source Code Snippets

All errors include source code context:
- The line containing the error
- Position marker (`^`) pointing to exact location
- Optional: 1-2 lines of surrounding context

## Stack Traces (Tasks 9.113-9.116)

### Automatic Stack Traces

Unhandled exceptions automatically display stack traces:

```bash
./bin/dwscript run script.dws
```

Output:
```
Runtime Error: Exception: Error message [line: 7, column: 3]
Level3 [line: 7, column: 3]
Level2 [line: 13, column: 3]
Level1 [line: 19, column: 3]
```

### Programmatic Stack Traces

Two built-in functions provide stack trace access:

#### GetStackTrace()

Returns formatted string:

```pascal
var trace: String;
trace := GetStackTrace();
PrintLn(trace);
```

Output:
```
FunctionName [line: 10, column: 5]
CallerFunction [line: 20, column: 3]
```

#### GetCallStack()

Returns structured array:

```pascal
var stack: array of Variant;
stack := GetCallStack();
PrintLn('Depth: ' + IntToStr(Length(stack)));
```

Each element is a record with:
- `FunctionName: String`
- `FileName: String`
- `Line: Integer`
- `Column: Integer`

## Error Context

### Variable Names

Type errors include variable names when applicable:

```
Cannot assign Float to Integer variable 'count'
```

### Function Names

Function-related errors include function names:

```
Function 'ProcessNumber' expects Integer as parameter 1, got Float
```

### Type Information

Type errors show both expected and actual types:

```
Expected: Integer
Actual: Float
```

### Runtime Values

Runtime errors show relevant values:

```
Division by zero: 10 / 0
  Left operand: 10
  Right operand: 0
```

## Best Practices

### For Developers

1. **Read the full error message**: Don't just look at the line number
2. **Check the source snippet**: The `^` marker shows exactly where the error is
3. **Review type information**: Expected vs actual types guide the fix
4. **Use stack traces**: Understand the call path that led to the error
5. **Look for suggestions**: Some errors include hints for fixes

### For Error Message Design

1. **Be specific**: Include variable/function names
2. **Show context**: Display source code and values
3. **Be actionable**: Suggest what to fix
4. **Be consistent**: Follow the standard format
5. **Provide details**: Types, values, locations

## Error Reporting Infrastructure

### CompilerError Type

Defined in `internal/errors/compiler_error.go`:

```go
type CompilerError struct {
    Position   lexer.Position
    Message    string
    SourceCode string
    FileName   string
}
```

**Methods**:
- `Format(useColor bool) string`: Formats error with source snippet
- `String() string`: Simple string representation

### RuntimeError Type

Defined in `internal/interp/errors.go`:

```go
type RuntimeError struct {
    Message    string
    Pos        *lexer.Position
    Expression string
    Values     map[string]string
    SourceCode string
    SourceFile string
    ErrorType  string
    CallStack  errors.StackTrace
}
```

**Methods**:
- `ToCompilerError() *CompilerError`: Converts to formatted error
- `String() string`: Simple string representation

### ExceptionValue Type

Defined in `internal/interp/exceptions.go`:

```go
type ExceptionValue struct {
    ClassInfo *ClassInfo
    Instance  *ObjectInstance
    Message   string
    Position  *lexer.Position
    CallStack errors.StackTrace
}
```

## Testing Error Messages

Test fixtures are available in `testdata/error_messages/`:

1. **Type errors**: 01-03
2. **Runtime errors**: 04-05
3. **Exception handling**: 06-08
4. **Combined**: 09

Run with:
```bash
./bin/dwscript run testdata/error_messages/<fixture>.dws
```

See `testdata/error_messages/README.md` for details.

## Command-Line Interface

### Error Display

The CLI (`cmd/dwscript/cmd/run.go`) handles error display:

1. **Parse errors**: Formatted via `errors.FormatErrors()`
2. **Semantic errors**: Converted to `CompilerError`, then formatted
3. **Runtime errors**: Checked for `RuntimeError`, formatted with source
4. **Exceptions**: Display class, message, position, and stack trace

### Example Output

```bash
$ ./bin/dwscript run script.dws
Error in script.dws:12:15
  12 | count := price;
                    ^
Cannot assign Float to Integer variable 'count'
Error: semantic analysis failed with 1 error(s)
```

## Future Enhancements

Potential improvements for error messages:

1. **Suggestions**: "Did you mean `cout` instead of `count`?"
2. **Error codes**: `E0001: Type mismatch`
3. **Multi-line errors**: Show errors spanning multiple lines
4. **Error recovery**: Continue after errors to show multiple issues
5. **IDE integration**: Machine-readable error format (JSON)
6. **Documentation links**: Link to docs for error types

## Related Documentation

- `docs/exception-enhancements-summary.md` - Exception features (Tasks 9.113-9.114)
- `docs/debugging-information-summary.md` - Debugging features (Tasks 9.115-9.116)
- `testdata/error_messages/README.md` - Test fixtures overview

## Conclusion

The enhanced error messages in go-dws provide developers with:
- **Clear context**: Source code, positions, values
- **Actionable information**: Types, names, suggestions
- **Complete traces**: Full call stack when needed
- **Consistent format**: Easy to read and understand

These improvements significantly enhance the developer experience and make debugging faster and more effective.
