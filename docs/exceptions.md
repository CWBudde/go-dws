# DWScript Exception Handling

**Research Date**: 2025-10-23
**Status**: Documentation for implementation (Stage 8, Task 8.189)
**Current Implementation**: Tokens defined, parser/AST/interpreter support NOT implemented

## Overview

DWScript supports structured exception handling with `try...except...finally...end` blocks, similar to Delphi/Object Pascal. This document captures the complete syntax from the reference implementation for future implementation in go-dws.

## Exception Handling Forms

### 1. try...except...end

Basic exception catching without finally block.

**Syntax**:
```pascal
try
  // protected code
except
  // exception handling code
end;
```

**Example**:
```pascal
try
  raise Exception.Create('error message');
except
  PrintLn('Caught exception');
end;
```

**Notes**:
- Catches any exception raised in the `try` block
- Control continues after the `end` if exception is caught
- If no exception occurs, `except` block is skipped

### 2. try...finally...end

Resource cleanup guarantee, executes regardless of exception.

**Syntax**:
```pascal
try
  // code that may raise exception
finally
  // cleanup code (always executes)
end;
```

**Example**:
```pascal
function Test(i: Integer): Boolean;
var
  ts: TObject;
begin
  Result := False;
  ts := TObject.Create;
  try
    if (i < 10) then Exit;
  finally
    ts.Free;  // Always executes, even on Exit
  end;
  Result := True;
end;
```

**Notes**:
- `finally` block ALWAYS executes
- Executes even if exception is raised (then re-raises)
- Executes even if `Exit`, `Break`, `Continue` used
- Does NOT catch exceptions (they propagate after finally)
- `ExceptObject` is available in finally block if exception is active

### 3. try...except...finally...end

Combined form: catch exceptions AND ensure cleanup.

**Syntax**:
```pascal
try
  // protected code
except
  // exception handling
finally
  // cleanup code
end;
```

**Example**:
```pascal
try
  raise Exception.Create('DOH');
except
  Print('Exception ');
finally
  PrintLn('World');  // Always prints
end;
```

**Execution Order**:
1. `try` block executes
2. If exception: `except` block executes, then `finally`
3. If no exception: `finally` block executes
4. `finally` always runs regardless of exception or control flow

### 4. Typed Exception Handlers (on...do)

Catch specific exception types with typed handlers.

**Syntax**:
```pascal
try
  // protected code
except
  on E: ExceptionType1 do
    // handle ExceptionType1
  on E: ExceptionType2 do
    // handle ExceptionType2
  else
    // catch-all for other exceptions
end;
```

**Example**:
```pascal
type
  MyException = class(Exception);
  OtherException = class(Exception);

try
  raise MyException.Create('error');
except
  on e: MyException do
    PrintLn('MyException: ' + e.Message);
  on e: OtherException do
    PrintLn('OtherException: ' + e.Message);
  else
    PrintLn('Other exception type');
end;
```

**Handler Matching**:
- Handlers are checked **in order** (top to bottom)
- First matching handler executes (most specific to general)
- Exception instance is bound to the variable (e.g., `e`)
- Use base class `Exception` to catch any exception type
- `else` clause catches exceptions not matched by any `on` handler

**Example with inheritance**:
```pascal
try
  raise OtherException.Create('message');
except
  on e: Exception do  // Catches OtherException (subclass)
    PrintLn('Caught: ' + e.Message);
end;
```

### 5. Bare except (Catch-All)

Catch all exceptions without type checking.

**Syntax**:
```pascal
try
  // protected code
except
  // handles any exception
end;
```

**Example**:
```pascal
try
  raise Exception.Create('error');
except
  PrintLn('Something went wrong');
end;
```

**Notes**:
- Catches any exception type
- Cannot access exception object directly
- Use `ExceptObject` to access current exception (see Special Variables)

## Raise Statement

### Raise with Exception Object

Create and raise a new exception.

**Syntax**:
```pascal
raise ExceptionType.Create('message');
```

**Examples**:
```pascal
// Standard exception
raise Exception.Create('Something went wrong');

// Custom exception
type EMyExcept = class(Exception) end;
raise new EMyExcept('Custom error');

// Alternative syntax (both supported)
raise Exception.Create('error');
raise new Exception('error');
```

### Re-raise (Bare raise)

Re-raise the current exception within an exception handler.

**Syntax**:
```pascal
raise;
```

**Example**:
```pascal
try
  try
    raise Exception.Create('error');
  except
    PrintLn('Caught once');
    raise;  // Re-raise to outer handler
  end;
except
  PrintLn('Caught again');
end;
```

**Notes**:
- `raise;` without expression re-raises current exception
- Only valid within `except` block
- Using `raise;` outside exception handler is an error
- Preserves original exception and stack trace

### Raise with ExceptObject

Re-raise using the special `ExceptObject` variable.

**Syntax**:
```pascal
raise ExceptObject;
```

**Example**:
```pascal
Procedure ExceptionHandler;
Begin
  // Outside the original except block
  Raise ExceptObject;  // Re-raises the current exception
End;

Try
  var x := 5 div 0;  // Division by zero
Except
  ExceptionHandler;  // Re-raises via ExceptObject
End;
```

**Notes**:
- `ExceptObject` holds current exception instance
- Can be used to re-raise from outside the immediate except block
- Functionally equivalent to bare `raise` within except block

## Special Variables and Functions

### ExceptObject

Global variable holding the current exception instance.

**Type**: `Exception` or `nil`

**Behavior**:
- Set to exception instance when exception is active
- Available in `except` blocks
- Available in `finally` blocks if exception is active
- Set to `nil` after exception handling completes
- Nested try/except: inner exception shadows outer

**Example**:
```pascal
if ExceptObject <> nil then PrintLn('Bug');  // nil before exception

try
  raise Exception.Create('hello');
except
  PrintLn(ExceptObject.ClassName + ': ' + ExceptObject.Message);

  try
    raise EMyExcept.Create('world');
  except
    on e: EMyExcept do begin
      PrintLn(e.ClassName + ': ' + e.Message);
      PrintLn(ExceptObject.ClassName + ': ' + ExceptObject.Message);
      if e <> ExceptObject then PrintLn('bug');  // e and ExceptObject are same
    end;
  end;

  PrintLn(ExceptObject.ClassName + ': ' + ExceptObject.Message);  // Outer exception restored
end;

if ExceptObject <> nil then PrintLn('Bug');  // nil after handling
```

### Exception in Finally Block

`ExceptObject` is accessible in finally blocks when an exception is active.

**Example**:
```pascal
try
  try
    raise EMyExcept.Create('hello world');
  finally
    PrintLn(ExceptObject.ClassName + ': ' + ExceptObject.Message);  // Accessible
  end;
except
  PrintLn(ExceptObject.ClassName + ': ' + ExceptObject.Message);  // Still accessible
end;
```

## Exception Class Hierarchy

### Base Exception Class

**Class**: `Exception`

Inherits from `TObject`. All script-level exceptions must inherit from this class.

**Fields** (internal):
- `FMessage: String` (protected) - Stores the exception message
- `FDebuggerField: Integer` (protected) - Internal debugger support

**Properties**:
- `Message: String` (public, read/write) - Exception message text
- `ClassName: String` (inherited from TObject) - Name of the exception class

**Methods**:
- `Create(Msg: String): Exception` - Constructor that creates exception with message
- `Destroy()` - Destructor (virtual, overrides TObject.Destroy)
- `StackTrace(): String` - Returns formatted stack trace as string

**Constructor Usage**:
```pascal
// Create and raise
raise Exception.Create('error message');

// Alternative syntax with 'new'
raise new Exception('error message');
```

**StackTrace Example**:
```pascal
try
  raise Exception.Create('error');
except
  on E: Exception do
    PrintLn(E.StackTrace);  // Shows call stack
end;
```

### Standard Exception Types

DWScript defines these built-in exception classes:

#### 1. EAssertionFailed

Raised by failed `Assert()` statements.

**Definition**:
```pascal
type EAssertionFailed = class(Exception);
```

**Usage**:
```pascal
Assert(condition);          // Raises EAssertionFailed if false
Assert(condition, 'message'); // With custom message

try
  Assert(False, 'boom');
except
  on E: EAssertionFailed do
    PrintLn(E.Message);  // Prints: "boom"
end;
```

**Inherits from**: `Exception`
**Additional members**: None (same as Exception)

#### 2. EDelphi

Wraps native Delphi/host exceptions that occur during external calls.

**Definition**:
```pascal
type EDelphi = class(Exception)
  property ExceptionClass: String;  // Name of original Delphi exception
end;
```

**Fields** (internal):
- `FExceptionClass: String` (protected) - Stores original exception class name

**Properties**:
- `ExceptionClass: String` (public, read/write) - Original Delphi exception class name
- `Message: String` (inherited) - Exception message

**Constructor**:
```pascal
EDelphi.Create(Cls: String, Msg: String): EDelphi
```

**Usage**:
When a Delphi exception occurs in external code, DWScript wraps it as EDelphi:
```pascal
try
  // Call to external function that raises EConvertError
  ExternalFunc();
except
  on E: EDelphi do begin
    PrintLn(E.ExceptionClass);  // Prints: "EConvertError"
    PrintLn(E.Message);         // Prints original message
  end;
end;
```

**Inherits from**: `Exception`

### Runtime Error Messages

DWScript raises `Exception` instances with specific messages for runtime errors:

| Error Condition | Message Template | Example |
|----------------|------------------|---------|
| Array upper bound exceeded | `Upper bound exceeded! Index %d` | `Upper bound exceeded! Index 10` |
| Array lower bound exceeded | `Lower bound exceeded! Index %d` | `Lower bound exceeded! Index -1` |
| Division by zero | (Varies by context) | (Wrapped in EDelphi from host) |
| Invalid cast | `Cannot cast instance of type "%s" to class "%s"` | `Cannot cast instance of type "TObject" to class "TMyClass"` |
| Object not instantiated | `Object not instantiated` | When accessing nil object reference |
| Function pointer is nil | `Function pointer is nil` | When calling nil function pointer |
| Abstract instance | `Trying to create an instance of an abstract class` | Creating instance of abstract class |

**Note**: Unlike Delphi, DWScript doesn't define separate `EConvertError`, `ERangeError`, `EDivByZero` classes at the script level. Instead, it raises generic `Exception` or `EDelphi` instances with appropriate messages.

### Custom Exception Types

Scripts can define custom exception types by inheriting from `Exception`:

**Basic Custom Exception**:
```pascal
type
  EMyException = class(Exception);
```

**Custom Exception with Additional Members**:
```pascal
type
  EMyError = class(Exception)
  private
    FErrorCode: Integer;
  public
    property ErrorCode: Integer read FErrorCode write FErrorCode;
  end;
```

**Usage**:
```pascal
type
  EMyExcept = class(Exception) end;

try
  raise new EMyExcept('Custom error');
except
  on E: EMyExcept do
    PrintLn(E.Message);
  on E: Exception do
    PrintLn('Other exception');
end;
```

**Inheritance Chain Example**:
```pascal
type
  EMyBase = class(Exception);
  EMyDerived = class(EMyBase);

try
  raise EMyDerived.Create('error');
except
  on E: EMyBase do       // Catches EMyDerived (inheritance)
    PrintLn('Caught');
end;
```

## Control Flow Interactions

### Exit in Finally

`finally` blocks execute even when using `Exit` statement.

**Example**:
```pascal
function Hello: String;
begin
  Result := 'duh';
  try
    Result := 'Hello';
    if True then Exit;  // Exits function
  finally
    Result := Result + ' world';  // Still executes!
  end;
  Result := 'Bye bye';  // Never reached
end;

PrintLn(Hello);  // Prints: "Hello world"
```

### Break/Continue in Exception Blocks

**Break in except block**: Allowed
```pascal
for i := 1 to 10 do begin
  try
    if i = 5 then raise Exception.Create('stop');
  except
    break;  // Valid: exits loop
  end;
end;
```

**Break/Continue in finally block**: ERROR
```pascal
for i := 1 to 10 do begin
  try
    // ...
  finally
    break;  // ERROR: not allowed in finally
  end;
end;
```

## Syntax Errors and Edge Cases

### Invalid Exception Handler Syntax

**ERROR - Missing exception variable**:
```pascal
try
except
  on ;  // ERROR: syntax error
end;
```

**ERROR - Missing colon and type**:
```pascal
try
except
  on e ;  // ERROR: expected ':' and type
end;
```

**ERROR - Non-exception type**:
```pascal
try
except
  on e: Integer do ;  // ERROR: must be Exception type
end;
```

### Bare Raise Outside Exception Handler

**ERROR - Raise without exception context**:
```pascal
if True then
  raise  // ERROR: raise without expression requires exception context
else ;
```

**Valid - Raise within except block**:
```pascal
try
except
  if True then
    raise  // OK: re-raises current exception
  else ;
end;
```

## Implementation Notes

### Current Status (go-dws)

**Implemented**:
- ✅ Lexer tokens: `TRY`, `EXCEPT`, `FINALLY`, `RAISE`, `ON`

**NOT Implemented**:
- ❌ Parser support for try/except/finally statements
- ❌ AST nodes for exception handling structures
- ❌ Semantic analysis for exception types
- ❌ Interpreter execution of exception handling
- ❌ Stack unwinding mechanism
- ❌ Exception class hierarchy
- ❌ ExceptObject global variable

### Implementation Strategy

**Phase 1: AST Nodes** (Stage 8)
- `TryStmt` node with `Try`, `Except`, `Finally` blocks
- `RaiseStmt` node with optional expression
- `ExceptHandler` node for `on E: Type do` clauses

**Phase 2: Parser** (Stage 8)
- Parse `try...except...end`
- Parse `try...finally...end`
- Parse `try...except...finally...end`
- Parse `on E: Type do` handlers
- Parse `else` catch-all
- Parse `raise` with/without expression

**Phase 3: Semantic Analysis** (Stage 8)
- Validate exception types inherit from Exception
- Validate `raise;` only in except blocks
- Validate exception handler types are valid classes
- Validate control flow in finally blocks

**Phase 4: Interpreter** (Stage 8)
- Implement exception objects and propagation
- Implement stack unwinding
- Implement exception matching (most specific first)
- Implement `ExceptObject` global variable
- Implement finally block guarantee
- Integrate with existing control flow (Exit, Break, Continue)

**Phase 5: Standard Exceptions** (Stage 8)
- Define Exception base class
- Implement EConvertError, ERangeError, EDivByZero
- Implement EAssertionFailed
- Add exception raising to runtime operations (div by zero, array bounds, etc.)

## Test Files Reference

Reference test files in `reference/dwscript-original/Test/`:

**Basic Exception Handling**:
- `SimpleScripts/try_except_finally.pas` - All three forms
- `SimpleScripts/exceptions.pas` - Typed handlers, multiple catch
- `SimpleScripts/exceptions2.pas` - Nested try/except, finally
- `SimpleScripts/exceptions3.pas` - More exception scenarios

**Exception Objects**:
- `SimpleScripts/exceptobj.pas` - ExceptObject variable
- `SimpleScripts/exceptobj2.pas` - ExceptObject scoping
- `SimpleScripts/exceptobj3.pas` - ExceptObject edge cases

**Re-raise**:
- `SimpleScripts/re_raise.pas` - Bare raise statement
- `SimpleScripts/raise_nil.pas` - Raise with nil

**Control Flow**:
- `SimpleScripts/exit_finally.pas` - Exit in finally blocks
- `SimpleScripts/exit_finally2.pas` - More exit scenarios
- `SimpleScripts/break_in_except_block.pas` - Break in except

**Assertions**:
- `SimpleScripts/assert.pas` - EAssertionFailed exceptions
- `SimpleScripts/assert_variant.pas` - Assert with variants

**Syntax Errors** (FailureScripts/):
- `try_except1.pas` - Missing end
- `raise_syntax.pas` - Invalid raise syntax
- `except_error1.pas` - Missing exception variable
- `except_error2.pas` - Missing type
- `except_error3.pas` - Invalid handler
- `except_error4.pas` - Invalid handler syntax
- `except_error5.pas` - Invalid handler syntax
- `break_in_finally.pas` - Break not allowed in finally

**Other**:
- `SimpleScripts/exception_log.pas` - Exception logging
- `SimpleScripts/exception_nested_call.pas` - Nested calls with exceptions
- `SimpleScripts/exception_scoping.pas` - Exception scoping rules

## Exception Handling Execution Strategy

### Overview

This section defines the precise execution semantics and algorithms for exception handling in DWScript. These algorithms ensure correct behavior for all exception handling constructs.

### Control Flow for Try/Except/Finally Blocks

#### 1. try...except...end Execution

**Algorithm**:
```
ExecuteTryExcept(tryBlock, exceptHandlers):
  1. Save current ExceptObject state → savedExceptObj
  2. Execute tryBlock
  3. If no exception:
       a. Return normally
  4. If exception occurred:
       a. Set ExceptObject := exception instance
       b. For each handler in exceptHandlers (in order):
            i.   If handler has no type filter (bare except/else):
                   → Match found, goto step 4c
            ii.  If exception.IsInstanceOf(handler.exceptionType):
                   → Match found, goto step 4c
            iii. Continue to next handler
       c. If match found:
            i.   If handler has variable name:
                   → Bind variable to exception instance in handler scope
            ii.  Execute handler block
            iii. If handler block completes normally:
                   → Clear ExceptObject (set to nil)
                   → Pop exception from exception stack
                   → Restore savedExceptObj
                   → Return normally
            iv.  If handler block raises exception:
                   → Leave ExceptObject set
                   → Propagate new exception
       d. If no match found:
            → Restore savedExceptObj
            → Propagate exception (re-raise)
```

**State Transitions**:
```
[Normal Execution]
    ↓ try block
[Try Block Executing]
    ↓ exception raised
[Exception Active] → ExceptObject = exception
    ↓ check handlers
[Matching Handler] → bind variable
    ↓ execute handler
[Handler Executing]
    ↓ completes normally
[Exception Cleared] → ExceptObject = nil
    ↓
[Normal Execution]

OR:
[Matching Handler]
    ↓ new exception
[New Exception Active] → propagate
```

**Example**:
```pascal
try
  raise Exception.Create('error');  // Step 2: exception occurs
except
  on E: Exception do               // Step 4b: match found
    PrintLn(E.Message);            // Step 4c: execute handler
end;                               // Step 4c.iii: return normally
```

#### 2. try...finally...end Execution

**Algorithm**:
```
ExecuteTryFinally(tryBlock, finallyBlock):
  1. Save current ExceptObject state → savedExceptObj
  2. Execute tryBlock → tryResult
  3. Save tryResult (exception or normal return value)
  4. Execute finallyBlock → finallyResult
  5. If finallyBlock raised exception (finallyResult is exception):
       a. Discard tryResult
       b. Propagate finallyResult
  6. If finallyBlock completed normally:
       a. If tryResult is exception:
            → Propagate tryResult
       b. If tryResult is normal:
            → Return normally with tryResult
```

**Key Principle**: Finally block **ALWAYS** executes, and finally exceptions **override** try exceptions.

**State Transitions**:
```
[Normal Execution]
    ↓ try block
[Try Block Executing]
    ↓ may raise exception
[Try Complete] → save result (exception or normal)
    ↓ always execute
[Finally Block Executing]
    ↓ may raise exception
[Finally Complete] → save result
    ↓ priority check
[Determine Result]:
  - Finally exception? → propagate finally exception
  - Try exception? → propagate try exception
  - Both normal? → return normally
```

**Examples**:

```pascal
// Normal case: no exception
try
  x := 42;
finally
  PrintLn('cleanup');  // Always prints
end;
// Returns normally after finally

// Exception in try, finally executes
try
  raise Exception.Create('error');
finally
  PrintLn('cleanup');  // Prints before propagating
end;
// Exception propagates after finally

// Exception in finally overrides try
try
  raise Exception.Create('try error');
finally
  raise Exception.Create('finally error');
end;
// Propagates 'finally error', not 'try error'
```

#### 3. try...except...finally...end Execution

**Algorithm**:
```
ExecuteTryExceptFinally(tryBlock, exceptHandlers, finallyBlock):
  1. Save current ExceptObject state → savedExceptObj
  2. Execute tryBlock → tryResult
  3. exceptResult := tryResult
  4. If tryResult is exception:
       a. Set ExceptObject := exception instance
       b. For each handler in exceptHandlers (in order):
            i.   If handler matches exception:
                   → Execute handler → handlerResult
                   → exceptResult := handlerResult
                   → Break
       c. If handler matched and completed normally:
            → Clear ExceptObject
            → exceptResult := normal
       d. If no handler matched:
            → exceptResult := tryResult (exception)
  5. Execute finallyBlock → finallyResult
  6. If finallyBlock raised exception:
       a. Discard exceptResult
       b. Propagate finallyResult
  7. If finallyBlock completed normally:
       a. If exceptResult is exception:
            → Propagate exceptResult
       b. If exceptResult is normal:
            → Return normally
```

**Execution Order**:
1. Try block
2. If exception: Except handlers (may catch)
3. **Always**: Finally block
4. Propagate or return based on results

**State Transitions**:
```
[Normal Execution]
    ↓
[Try Block] → may raise
    ↓
[Exception?]
    ├─[Yes]→ [Match Handlers] → may catch
    └─[No]──────────────────────┐
                                ↓
                          [Always Execute]
                                ↓
                          [Finally Block] → may raise
                                ↓
                          [Determine Result]:
                            - Finally exception? → propagate
                            - Except caught? → return normally
                            - Try exception uncaught? → propagate
                            - All normal? → return normally
```

**Complex Example**:
```pascal
try
  raise Exception.Create('try');    // Step 2: exception
except
  on E: Exception do
    PrintLn('caught: ' + E.Message); // Step 4: match, execute
    // Completes normally              Step 4c: clear exception
finally
  PrintLn('cleanup');                 // Step 5: always executes
end;
// Returns normally (exception was caught)

// VS:

try
  raise Exception.Create('try');    // Step 2: exception
except
  on E: MyException do              // Step 4: no match
    PrintLn('not reached');
finally
  PrintLn('cleanup');               // Step 5: always executes
end;
// Propagates 'try' exception after cleanup
```

### Stack Unwinding Mechanism

#### Overview

Exception propagation causes stack unwinding: each stack frame is exited until a matching exception handler is found or the top-level is reached.

#### Unwinding Algorithm

```
RaiseException(exceptionObj, position):
  1. Create ScriptException wrapper:
       a. capturedStack := GetCurrentCallStack()
       b. scriptException := {
            ExceptionObj: exceptionObj,
            Position: position,
            CallStack: capturedStack,
            Message: exceptionObj.Message
          }
  2. Push exceptionObj onto exception stack
  3. Set ExceptObject := exceptionObj
  4. Return scriptException as error

UnwindStack(exception):
  1. Current frame returns exception as error
  2. Caller frame receives error
  3. If caller has try...except:
       a. Check if exception matches any handler
       b. If match: execute handler, stop unwinding
       c. If no match: continue unwinding
  4. If caller has try...finally:
       a. Execute finally block
       b. Continue unwinding (or propagate finally exception)
  5. If no try block in caller:
       a. Continue unwinding to next caller
  6. Repeat until:
       a. Exception is caught (stop unwinding), OR
       b. Top-level reached (program terminates with error)
```

#### Stack Frame States During Unwinding

```
[Stack Before Exception]
  ┌─────────────────┐
  │ main()          │
  ├─────────────────┤
  │ processData()   │
  ├─────────────────┤
  │ validateInput() │
  ├─────────────────┤
  │ checkRange()    │ ← Exception raised here
  └─────────────────┘

[Unwinding Phase 1]
  ┌─────────────────┐
  │ main()          │
  ├─────────────────┤
  │ processData()   │
  ├─────────────────┤
  │ validateInput() │ ← Returns error
  └─────────────────┘

[Unwinding Phase 2]
  ┌─────────────────┐
  │ main()          │
  ├─────────────────┤
  │ processData()   │ ← Has try...except, checks handlers
  └─────────────────┘

[If Handler Matches]
  ┌─────────────────┐
  │ main()          │
  ├─────────────────┤
  │ processData()   │ ← Executes handler, stops unwinding
  └─────────────────┘

[If No Handler]
  ┌─────────────────┐
  │ main()          │ ← Continues unwinding
  └─────────────────┘
```

#### ExceptObject During Unwinding

**Nested Exception Handling**:
```pascal
procedure Outer;
begin
  try
    raise Exception.Create('outer');
  except
    // ExceptObject = 'outer' exception

    try
      raise Exception.Create('inner');
    except
      // ExceptObject = 'inner' exception (shadows outer)
      PrintLn(ExceptObject.Message);  // Prints: "inner"
    end;

    // ExceptObject = 'outer' exception (restored)
    PrintLn(ExceptObject.Message);  // Prints: "outer"
  end;

  // ExceptObject = nil (cleared)
end;
```

**Implementation**:
- Maintain exception stack: `[outer, inner]`
- ExceptObject always points to top of stack
- Entering except handler: push exception
- Leaving except handler: pop exception
- ExceptObject := top of stack (or nil if empty)

#### Unwinding with Finally Blocks

```
Example:
  func1()
    try
      func2()
    finally
      cleanup1()
    end

  func2()
    try
      func3()
    finally
      cleanup2()
    end

  func3()
    raise Exception.Create('error')

Unwinding sequence:
  1. func3(): raise exception → return error
  2. func2(): receive error
       a. Execute cleanup2() (finally)
       b. Return same error (continue unwinding)
  3. func1(): receive error
       a. Execute cleanup1() (finally)
       b. Return same error (continue unwinding)
  4. Top-level: terminate with error

Result: Both cleanup1 and cleanup2 executed during unwinding
```

### Exception Matching Algorithm

#### Overview

When an exception is raised and caught by a try...except block, the interpreter must determine which handler (if any) should execute.

#### Matching Rules

1. **Handlers are checked in order** (top to bottom)
2. **First matching handler executes** (no fall-through)
3. **Match succeeds if**: exception is instance of handler type (or subclass)
4. **Bare except/else**: matches any exception type

#### Algorithm

```
MatchException(exception, handlers):
  For each handler in handlers (in declaration order):
    1. If handler.exceptionType == nil:
         → Return Match(handler)  // Bare except/else

    2. If IsInstanceOf(exception, handler.exceptionType):
         → Return Match(handler)  // Type match (including subclasses)

    3. Continue to next handler

  Return NoMatch  // No handler matched

IsInstanceOf(object, targetClass):
  currentClass := object.Class
  While currentClass != nil:
    If currentClass == targetClass:
      Return true
    currentClass := currentClass.Parent
  Return false
```

#### Matching Examples

**Example 1: Most Specific First**
```pascal
type
  EBase = class(Exception);
  EDerived = class(EBase);

try
  raise EDerived.Create('error');
except
  on E: EDerived do        // ← Matches here (most specific)
    PrintLn('Derived');
  on E: EBase do           // Not checked (already matched)
    PrintLn('Base');
  on E: Exception do       // Not checked
    PrintLn('Exception');
end;
// Output: "Derived"
```

**Example 2: Order Matters**
```pascal
try
  raise EDerived.Create('error');
except
  on E: Exception do       // ← Matches here (first match)
    PrintLn('Exception');
  on E: EDerived do        // Never reached (shadowed)
    PrintLn('Derived');
end;
// Output: "Exception"
```

**Example 3: Bare Except Catches All**
```pascal
try
  raise MyCustomException.Create('error');
except
  on E: SomeOtherException do
    PrintLn('Other');      // No match
  else                     // ← Matches here (catch-all)
    PrintLn('Caught');
end;
// Output: "Caught"
```

**Example 4: No Match**
```pascal
type
  EMyException = class(Exception);
  EOtherException = class(Exception);

try
  raise EMyException.Create('error');
except
  on E: EOtherException do   // No match (different branch)
    PrintLn('Caught');
end;
// Exception propagates (not caught)
```

#### Handler Type Hierarchy Check

```
Class Hierarchy:
         TObject
            ↓
        Exception
         ↙    ↘
    EBase    EOther
       ↓
   EDerived

Exception instance: EDerived.Create('msg')

Handler matching:
  on E: EDerived   → ✓ Match (exact type)
  on E: EBase      → ✓ Match (parent class)
  on E: Exception  → ✓ Match (grandparent)
  on E: TObject    → ✓ Match (root class)
  on E: EOther     → ✗ No match (different branch)
  else             → ✓ Match (catch-all)
```

### Finally Block Guarantee

#### Principle

The `finally` block **ALWAYS executes**, regardless of:
1. Normal completion of try block
2. Exception in try block
3. `Exit` statement in try block
4. `Break` statement in try block (if in loop)
5. `Continue` statement in try block (if in loop)
6. Exception in except handler
7. `Exit` in except handler

#### Execution Guarantee Algorithm

```
ExecuteWithFinallyGuarantee(tryBlock, finallyBlock):
  1. capturedControlFlow := none
  2. Try to execute tryBlock:
       a. If completes normally:
            → capturedControlFlow := normal
       b. If raises exception:
            → capturedControlFlow := exception(e)
       c. If executes Exit:
            → capturedControlFlow := exit(returnValue)
       d. If executes Break:
            → capturedControlFlow := break
       e. If executes Continue:
            → capturedControlFlow := continue
  3. Execute finallyBlock:
       a. If finallyBlock raises exception:
            → Override capturedControlFlow
            → Propagate finally exception
       b. If finallyBlock executes Exit/Break/Continue:
            → Override capturedControlFlow
            → Use finally control flow
  4. Resume capturedControlFlow:
       a. normal: return normally
       b. exception(e): propagate e
       c. exit(value): exit function with value
       d. break: exit loop
       e. continue: next loop iteration
```

#### Examples

**Exit in Try, Finally Executes**:
```pascal
function Test: String;
begin
  try
    Result := 'Hello';
    Exit;                    // Exit function
    Result := 'Unreached';   // Never executed
  finally
    Result := Result + ' World';  // Still executes!
  end;
  Result := 'Unreached';     // Never executed
end;

// Returns: "Hello World"
```

**Break in Try, Finally Executes**:
```pascal
for i := 1 to 10 do
begin
  try
    if i = 5 then Break;     // Break loop
  finally
    PrintLn('Cleanup ' + IntToStr(i));  // Executes before break
  end;
  PrintLn('After try');       // Not reached when i=5
end;

// Output:
// Cleanup 1
// After try
// Cleanup 2
// After try
// ...
// Cleanup 5
// (loop exits, no "After try" for i=5)
```

**Exception in Try, Finally Executes**:
```pascal
try
  raise Exception.Create('error');
finally
  PrintLn('Cleanup');         // Executes before propagating
end;

// Output: "Cleanup"
// Then exception propagates
```

**Finally Exception Overrides Exit**:
```pascal
function Test: String;
begin
  try
    Result := 'Try';
    Exit;                     // Attempts to exit
  finally
    raise Exception.Create('Finally error');  // Overrides exit!
  end;
end;

// Raises exception, does NOT return 'Try'
```

#### Finally Block Restrictions

**Allowed in Finally**:
- Any expression evaluation
- Method calls
- Assignments
- Exception raising

**NOT Allowed in Finally**:
- `Break` (compile error)
- `Continue` (compile error)

**Rationale**: Break/Continue would create ambiguous control flow when combined with try block's control flow.

```pascal
// ERROR: Break not allowed in finally
for i := 1 to 10 do
begin
  try
    // ...
  finally
    break;  // COMPILE ERROR
  end;
end;
```

### Re-raise Mechanism

#### Bare Raise Statement

**Syntax**: `raise;` (no expression)

**Semantics**: Re-raise the current exception (from ExceptObject)

#### Re-raise Algorithm

```
ExecuteBareRaise():
  1. Check ExceptObject:
       a. If ExceptObject == nil:
            → ERROR: "raise without exception context"
       b. If ExceptObject != nil:
            → Get exception instance from ExceptObject
  2. Create new ScriptException with same exception object:
       a. exceptionObj := ExceptObject
       b. message := exceptionObj.Message
       c. Capture current position (for updated stack trace)
  3. Return ScriptException as error (propagates)
```

#### Re-raise Contexts

**Valid Context**: Inside `except` handler
```pascal
try
  raise Exception.Create('error');
except
  PrintLn('Caught');
  raise;  // ✓ Valid: re-raises 'error'
end;
```

**Invalid Context**: Outside exception handler
```pascal
if condition then
  raise;  // ✗ ERROR: no exception context
```

**Invalid Context**: Inside `finally` block
```pascal
try
  raise Exception.Create('error');
finally
  raise;  // ✗ ERROR: ExceptObject not set in finally
end;
```

#### Re-raise Patterns

**Pattern 1: Log and Re-raise**
```pascal
try
  DoOperation();
except
  on E: Exception do begin
    LogError(E.Message);  // Log exception
    raise;                // Re-raise to caller
  end;
end;
```

**Pattern 2: Cleanup and Re-raise**
```pascal
try
  resource := AllocateResource();
  DoOperation();
except
  FreeResource(resource);  // Cleanup on error
  raise;                   // Re-raise exception
end;
```

**Pattern 3: Nested Re-raise**
```pascal
try
  try
    raise Exception.Create('error');
  except
    PrintLn('Inner handler');
    raise;  // Re-raise to outer handler
  end;
except
  PrintLn('Outer handler');  // Catches re-raised exception
end;
```

#### Re-raise with ExceptObject

Alternative to bare `raise`: explicitly raise `ExceptObject`

```pascal
procedure HandleException;
begin
  // Can re-raise from outside immediate except block
  if ExceptObject <> nil then
    raise ExceptObject;
end;

try
  raise Exception.Create('error');
except
  HandleException;  // Re-raises via ExceptObject
end;
```

**Difference from Bare Raise**:
- Bare `raise`: Only valid in except block lexically
- `raise ExceptObject`: Valid anywhere ExceptObject is set

#### Re-raise Call Stack

**Stack Trace Behavior**:
```pascal
function Level3: Integer;
begin
  raise Exception.Create('Level 3 error');
end;

function Level2: Integer;
begin
  try
    Result := Level3();
  except
    raise;  // Re-raise
  end;
end;

function Level1: Integer;
begin
  try
    Result := Level2();
  except
    on E: Exception do
      PrintLn(E.StackTrace);
  end;
end;

// Stack trace includes:
// - Level3 (original raise)
// - Level2 (re-raise point)
// - Level1 (final catch point)
```

**Key Point**: Re-raising preserves original exception object and its captured stack trace. Additional frames are added during unwinding.

### Execution State Machine

#### Overall State Transitions

```
                    ┌──────────────────┐
                    │  Normal          │
                    │  Execution       │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  Try Block       │
                    │  Executing       │
                    └─┬─────────────┬──┘
                      │             │
           Exception  │             │  Normal
                      │             │
        ┌─────────────▼──┐      ┌──▼──────────────┐
        │  Exception     │      │  Finally Block  │
        │  Raised        │      │  (if present)   │
        └─┬──────────────┘      └──┬──────────────┘
          │                        │
          │                        │
    ┌─────▼──────────┐             │
    │  Match         │             │
    │  Handlers      │             │
    └─┬────────────┬─┘             │
      │ Match      │ No Match      │
      │            │               │
  ┌───▼────────┐   │               │
  │  Execute   │   │               │
  │  Handler   │   │               │
  └───┬────────┘   │               │
      │            │               │
      │ Success    │ Failed        │
      │            │ or            │
      │            │ No Match      │
  ┌───▼────────┐   │               │
  │  Finally   │   │               │
  │  Block     │◄──┘               │
  │  (if any)  │◄──────────────────┘
  └───┬────────┘
      │
      │
  ┌───▼────────────┐
  │  Determine     │
  │  Final Result  │
  └───┬────────────┘
      │
      ├──► Normal Return
      ├──► Propagate Exception
      ├──► Exit Function
      ├──► Break Loop
      └──► Continue Loop
```

#### Exception State Variables

```
ExecutionContext:
  - ExceptObject: ClassInstance | nil
      Current exception (top of exception stack)

  - ExceptionStack: [ClassInstance]
      Stack of nested exceptions

  - InExceptHandler: boolean
      True when executing exception handler

  - InFinallyBlock: boolean
      True when executing finally block

  - CapturedControlFlow: enum {Normal, Exception, Exit, Break, Continue}
      Control flow to resume after finally
```

### Summary

The exception handling execution strategy defines:

1. **Control Flow**: Precise algorithms for try/except/finally execution order
2. **Stack Unwinding**: Frame-by-frame exception propagation with cleanup
3. **Handler Matching**: Type hierarchy-based matching with order significance
4. **Finally Guarantee**: Always-execute semantics for resource cleanup
5. **Re-raise**: Bare raise and ExceptObject-based re-raising

These algorithms ensure correct, predictable exception handling behavior that matches DWScript's semantics.

## Go Implementation Strategy

### Overview

This section maps DWScript exception types and mechanisms to Go implementation patterns for go-dws.

### Exception Type Mapping

| DWScript Type | Go Implementation | Description |
|---------------|------------------|-------------|
| `Exception` (script class) | `*ClassInstance` with `ExceptionClass` metadata | Script-level exception object |
| `EAssertionFailed` (script class) | `*ClassInstance` with `EAssertionFailedClass` metadata | Subclass of Exception |
| `EDelphi` (script class) | `*ClassInstance` with `EDelphiClass` metadata | Wraps host exceptions |
| `EScriptException` (internal) | Go `error` type implementing `ScriptError` interface | Wraps script exception for Go runtime |
| Host errors (Go) | Wrapped in `EDelphi` instance when caught by script | Bridge between Go and script errors |

### Core Go Types

```go
// ScriptException wraps a script-level exception object
type ScriptException struct {
    ExceptionObj *ClassInstance  // The script Exception instance
    ScriptPos    Position         // Where exception was raised
    CallStack    []CallFrame      // Stack trace
    Message      string           // Cached message
}

func (e *ScriptException) Error() string {
    return e.Message
}

// ExceptObject global variable (per execution context)
type ExecutionContext struct {
    // ... other fields ...
    ExceptObject *ClassInstance  // Current exception (nil when not in handler)
    ExceptionStack []*ClassInstance  // Nested exception stack
}
```

### Exception Class Definitions

Exception classes are defined in the system symbol table during initialization:

```go
// In semantic/builtins.go or similar

func defineExceptionClasses(symbolTable *SymbolTable) {
    // Exception base class
    exceptionClass := &ClassInfo{
        Name: "Exception",
        Parent: symbolTable.Classes["TObject"],
        Fields: map[string]*FieldInfo{
            "FMessage": {Name: "FMessage", Type: StringType, Visibility: VisProtected},
            "FDebuggerField": {Name: "FDebuggerField", Type: IntegerType, Visibility: VisProtected},
        },
        Properties: map[string]*PropertyInfo{
            "Message": {
                Name: "Message",
                Type: StringType,
                Getter: "FMessage",
                Setter: "FMessage",
                Visibility: VisPublic,
            },
        },
        Methods: map[string]*MethodInfo{
            "Create": {
                Name: "Create",
                Kind: MethodConstructor,
                Params: []Parameter{{Name: "Msg", Type: StringType}},
                ReturnType: nil, // constructors return instance
            },
            "Destroy": {
                Name: "Destroy",
                Kind: MethodDestructor,
                IsVirtual: true,
                IsOverride: true,
            },
            "StackTrace": {
                Name: "StackTrace",
                Kind: MethodFunction,
                ReturnType: StringType,
            },
        },
    }
    symbolTable.Classes["Exception"] = exceptionClass

    // EAssertionFailed
    assertionFailedClass := &ClassInfo{
        Name: "EAssertionFailed",
        Parent: exceptionClass,
    }
    symbolTable.Classes["EAssertionFailed"] = assertionFailedClass

    // EDelphi
    eDelphiClass := &ClassInfo{
        Name: "EDelphi",
        Parent: exceptionClass,
        Fields: map[string]*FieldInfo{
            "FExceptionClass": {Name: "FExceptionClass", Type: StringType, Visibility: VisProtected},
        },
        Properties: map[string]*PropertyInfo{
            "ExceptionClass": {
                Name: "ExceptionClass",
                Type: StringType,
                Getter: "FExceptionClass",
                Setter: "FExceptionClass",
                Visibility: VisPublic,
            },
        },
        Methods: map[string]*MethodInfo{
            "Create": {
                Name: "Create",
                Kind: MethodConstructor,
                Params: []Parameter{
                    {Name: "Cls", Type: StringType},
                    {Name: "Msg", Type: StringType},
                },
            },
        },
    }
    symbolTable.Classes["EDelphi"] = eDelphiClass
}
```

### Exception Handling Runtime

```go
// In interp/exception.go

// RaiseException creates and raises a script exception
func (i *Interpreter) RaiseException(exceptionObj *ClassInstance, pos Position) error {
    // Create CallStack from current execution state
    callStack := i.GetCallStack()

    // Get message from exception object
    message := exceptionObj.GetField("FMessage").String()

    // Create ScriptException wrapper
    scriptErr := &ScriptException{
        ExceptionObj: exceptionObj,
        ScriptPos: pos,
        CallStack: callStack,
        Message: message,
    }

    // Set as current exception
    i.Context.ExceptObject = exceptionObj
    i.Context.ExceptionStack = append(i.Context.ExceptionStack, exceptionObj)

    return scriptErr
}

// CatchException attempts to match exception with handler
func (i *Interpreter) CatchException(err error, handlers []ExceptionHandler) (matched bool, handlerIdx int) {
    scriptErr, ok := err.(*ScriptException)
    if !ok {
        // Host error - wrap in EDelphi
        scriptErr = i.WrapHostException(err)
    }

    exceptionObj := scriptErr.ExceptionObj

    // Try each handler in order
    for idx, handler := range handlers {
        if handler.ExceptionType == nil {
            // Bare 'except' or 'else' - catches all
            return true, idx
        }

        // Check if exception is instance of handler type
        if i.IsInstanceOf(exceptionObj, handler.ExceptionType) {
            return true, idx
        }
    }

    return false, -1
}

// WrapHostException wraps a Go error in EDelphi
func (i *Interpreter) WrapHostException(err error) *ScriptException {
    eDelphiClass := i.SymbolTable.Classes["EDelphi"]

    // Create EDelphi instance
    eDelphiObj := i.CreateClassInstance(eDelphiClass)

    // Set ExceptionClass field (Go error type name)
    errorType := fmt.Sprintf("%T", err)
    eDelphiObj.SetField("FExceptionClass", errorType)
    eDelphiObj.SetField("FMessage", err.Error())

    return &ScriptException{
        ExceptionObj: eDelphiObj,
        Message: err.Error(),
    }
}

// ExecuteTryExcept executes try...except...end block
func (i *Interpreter) ExecuteTryExcept(stmt *TryStmt) error {
    // Execute try block
    err := i.ExecuteStatement(stmt.TryBlock)

    if err != nil {
        // Exception occurred - try handlers
        matched, handlerIdx := i.CatchException(err, stmt.ExceptHandlers)

        if matched {
            handler := stmt.ExceptHandlers[handlerIdx]

            // Bind exception to variable if specified
            if handler.ExceptionVar != "" {
                scriptErr := err.(*ScriptException)
                i.Environment.Define(handler.ExceptionVar, scriptErr.ExceptionObj)
            }

            // Execute handler
            handlerErr := i.ExecuteStatement(handler.Statement)

            // Clear exception
            i.Context.ExceptObject = nil
            i.Context.ExceptionStack = i.Context.ExceptionStack[:len(i.Context.ExceptionStack)-1]

            // Re-raise if handler raised
            if handlerErr != nil {
                return handlerErr
            }

            // Exception caught successfully
            return nil
        }

        // No handler matched - propagate
        return err
    }

    // No exception
    return nil
}

// ExecuteTryFinally executes try...finally...end block
func (i *Interpreter) ExecuteTryFinally(stmt *TryStmt) error {
    // Execute try block
    tryErr := i.ExecuteStatement(stmt.TryBlock)

    // Always execute finally block
    finallyErr := i.ExecuteStatement(stmt.FinallyBlock)

    // Finally error takes precedence
    if finallyErr != nil {
        return finallyErr
    }

    // Return try error if any
    return tryErr
}

// ExecuteTryExceptFinally executes try...except...finally...end block
func (i *Interpreter) ExecuteTryExceptFinally(stmt *TryStmt) error {
    // Execute try block
    tryErr := i.ExecuteStatement(stmt.TryBlock)

    caughtErr := tryErr

    if tryErr != nil {
        // Try exception handlers
        matched, handlerIdx := i.CatchException(tryErr, stmt.ExceptHandlers)

        if matched {
            handler := stmt.ExceptHandlers[handlerIdx]

            // Bind exception to variable if specified
            if handler.ExceptionVar != "" {
                scriptErr := tryErr.(*ScriptException)
                i.Environment.Define(handler.ExceptionVar, scriptErr.ExceptionObj)
            }

            // Execute handler
            caughtErr = i.ExecuteStatement(handler.Statement)

            // If handler succeeded, exception was caught
            if caughtErr == nil {
                i.Context.ExceptObject = nil
                i.Context.ExceptionStack = i.Context.ExceptionStack[:len(i.Context.ExceptionStack)-1]
            }
        }
    }

    // Always execute finally block (even if exception caught)
    finallyErr := i.ExecuteStatement(stmt.FinallyBlock)

    // Finally error takes precedence
    if finallyErr != nil {
        return finallyErr
    }

    // Return caught/uncaught error
    return caughtErr
}
```

### Stack Unwinding

Exception propagation uses Go's natural error return mechanism:

1. When exception raised: return `*ScriptException` error
2. Each stack frame checks for error and returns it (unwinding)
3. `try...except` catches error if handler matches
4. `try...finally` executes finally block then re-returns error
5. Unhandled exceptions bubble to top-level interpreter

### ExceptObject Implementation

```go
// In interp/builtins.go

// ExceptObject built-in function
func exceptObjectFunc(i *Interpreter, args []Value) (Value, error) {
    if i.Context.ExceptObject == nil {
        return NilValue, nil
    }
    return ObjectValue(i.Context.ExceptObject), nil
}
```

### Exception Methods Implementation

```go
// In interp/exception_methods.go

// Exception.Create constructor
func exceptionCreateMethod(i *Interpreter, instance *ClassInstance, args []Value) error {
    msg := args[0].String()
    instance.SetField("FMessage", msg)
    instance.SetField("FDebuggerField", int64(0))
    return nil
}

// Exception.StackTrace method
func exceptionStackTraceMethod(i *Interpreter, instance *ClassInstance) (Value, error) {
    // Get call stack from exception's context
    // Format as string
    var sb strings.Builder

    // ... format stack trace ...

    return StringValue(sb.String()), nil
}
```

### Re-raise Implementation

```go
// Bare raise statement
func (i *Interpreter) ExecuteRaiseStmt(stmt *RaiseStmt) error {
    if stmt.Exception == nil {
        // Bare raise - re-raise current exception
        if i.Context.ExceptObject == nil {
            return fmt.Errorf("raise without exception context")
        }

        // Re-raise current exception
        return &ScriptException{
            ExceptionObj: i.Context.ExceptObject,
            Message: i.Context.ExceptObject.GetField("FMessage").String(),
        }
    }

    // Raise new exception
    exceptionVal, err := i.EvaluateExpression(stmt.Exception)
    if err != nil {
        return err
    }

    exceptionObj := exceptionVal.(*ClassInstance)
    return i.RaiseException(exceptionObj, stmt.Position)
}
```

### Key Design Decisions

1. **Script exceptions are ClassInstance objects**: Allows script code to inspect, subclass, and extend
2. **Host exceptions wrapped in EDelphi**: Clear separation between host and script errors
3. **ScriptException wrapper implements Go error**: Allows using Go error handling naturally
4. **ExceptObject per execution context**: Supports concurrent script execution
5. **Stack unwinding via error returns**: Leverages Go's error handling, no special control flow
6. **Finally blocks use defer-like semantics**: Always execute, even on exception or control flow changes

### Runtime Error Generation

When runtime errors occur (array bounds, nil dereference, etc.), create Exception instances:

```go
func (i *Interpreter) raiseRuntimeError(format string, args ...any) error {
    message := fmt.Sprintf(format, args...)

    exceptionClass := i.SymbolTable.Classes["Exception"]
    exceptionObj := i.CreateClassInstance(exceptionClass)
    exceptionObj.SetField("FMessage", message)

    return i.RaiseException(exceptionObj, i.CurrentPosition)
}

// Usage:
if index < 0 {
    return i.raiseRuntimeError("Lower bound exceeded! Index %d", index)
}
```

### Testing Strategy

1. **Unit tests** for each exception mechanism (try/except/finally/raise)
2. **Integration tests** from reference test files
3. **Edge case tests** for control flow interactions
4. **Concurrent execution tests** for ExceptObject isolation
5. **Performance tests** for exception overhead

## Summary

DWScript exception handling provides:

1. **Three forms**: try...except, try...finally, try...except...finally
2. **Typed handlers**: `on E: Type do` with inheritance-based matching
3. **Catch-all**: bare `except` or `else` clause
4. **Re-raise**: bare `raise` statement
5. **Special variable**: `ExceptObject` for current exception
6. **Finally guarantee**: executes even on exception, Exit, Break, Continue
7. **Exception hierarchy**: Base `Exception` class with standard types
8. **Control flow integration**: Proper interaction with Exit, Break, Continue

The syntax closely follows Delphi/Object Pascal conventions with full support for structured exception handling patterns.

**Go Implementation**: Exception objects are `ClassInstance` values, exceptions propagate via Go errors, and script/host errors are cleanly separated via the `EDelphi` wrapper pattern.
