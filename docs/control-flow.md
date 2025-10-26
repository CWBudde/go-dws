# DWScript Loop Control Statements

**Research Date**: 2025-10-26
**Status**: Documentation for implementation (Stage 8, Task 8.228)
**Current Implementation**: Tokens defined in lexer, parser/AST/interpreter support NOT fully implemented

## Overview

DWScript supports three loop control flow statements for managing program execution within loops and functions: `break`, `continue`, and `exit`. These statements provide early termination and flow control, similar to Delphi/Object Pascal.

**Key Points**:
- `break` and `continue` work only within loops (for/while/repeat)
- `exit` works within functions, procedures, and at program level
- All three statements respect exception handling (execute finally blocks)
- These statements only affect the innermost loop in nested structures

## Break Statement

The `break` statement immediately exits the innermost loop, transferring control to the statement following the loop's end.

### Syntax

```pascal
break;
```

### Valid Contexts

- Inside `for` loops
- Inside `while` loops
- Inside `repeat` loops
- **NOT valid** outside of loops (will cause semantic error)

### Basic Examples

**For loop with break**:
```pascal
for i := 1 to 10 do begin
   if i > 7 then break;
   PrintLn(i);  // Prints 1, 2, 3, 4, 5, 6, 7
end;
```

**While loop with break**:
```pascal
i := 1;
while i <= 10 do begin
   if i > 7 then break;
   PrintLn(i);
   i := i + 1;
end;
```

**Repeat loop with break**:
```pascal
i := 1;
repeat
   if i > 7 then break;
   PrintLn(i);
   i := i + 1;
until i > 10;
```

### Notes

- Break exits only the **innermost** loop
- Control transfers to the first statement after the loop's end
- Multiple break statements can exist in a single loop
- Break can be used in combination with if statements for conditional exit

## Continue Statement

The `continue` statement skips the remainder of the current iteration and proceeds to the next iteration of the innermost loop.

### Syntax

```pascal
continue;
```

### Valid Contexts

- Inside `for` loops (proceeds to next value)
- Inside `while` loops (re-evaluates condition)
- Inside `repeat` loops (continues to next iteration)
- **NOT valid** outside of loops (will cause semantic error)

### Basic Examples

**For loop with continue** (skip even numbers):
```pascal
for i := 1 to 10 do begin
   if (i and 1) = 0 then continue;
   PrintLn(i);  // Prints 1, 3, 5, 7, 9
end;
```

**While loop with continue**:
```pascal
i := 1;
while i <= 10 do begin
   if (i and 1) = 0 then begin
      i := i + 1;  // IMPORTANT: Increment before continue!
      continue;
   end;
   PrintLn(i);
   i := i + 1;
end;
```

**Repeat loop with continue**:
```pascal
i := 1;
repeat
   if (i and 1) = 0 then begin
      i := i + 1;  // IMPORTANT: Increment before continue!
      continue;
   end;
   PrintLn(i);
   i := i + 1;
until i > 10;
```

### Important Behavior Differences

- **For loops**: `continue` automatically advances to the next value in the range
- **While/Repeat loops**: Loop variable must be updated **before** `continue`, or you risk an infinite loop

### Notes

- Continue only affects the **innermost** loop
- Be careful with while/repeat loops: ensure loop variables are updated before continue
- Multiple continue statements can exist in a single loop

## Exit Statement

The `exit` statement immediately exits the current function or procedure, optionally returning a value.

### Syntax

```pascal
exit;              // Exit without returning a value
exit(expression);  // Exit and return a value
```

### Valid Contexts

- Inside functions (with or without return value)
- Inside procedures (without return value)
- At program level (exits the entire program)
- Inside any control structure (if, case, loops, try blocks)

### Basic Examples

**Exit from procedure**:
```pascal
procedure Test(s: String);
begin
   if s = 'quit' then exit;
   PrintLn('Processing: ', s);
end;
```

**Exit with return value**:
```pascal
function MyFunc(i: Integer): Integer;
begin
   if i <= 0 then
      exit(-1);     // Early return with value
   Result := i * 2;
end;
```

**Exit with Result variable**:
```pascal
function Hello: String;
begin
   Result := 'Hello';
   if True then exit;
   Result := 'Bye';  // Never executed
end;
```

**Exit with immediate value**:
```pascal
function MyFunc(i: Integer): Integer;
begin
   if i <= 0 then
      exit(-1)
   else if i = 10 then
      exit(i + 1)
   else
      exit(i * 2);
end;
```

### Exit Forms

1. **exit;** - Exits function/procedure without explicit return value
   - For functions: Returns the current value of `Result`
   - For procedures: Simply exits

2. **exit(value);** - Exits function and returns the specified value
   - Equivalent to setting `Result := value;` then `exit;`
   - More concise syntax for immediate returns

### Notes

- Exit always executes `finally` blocks before returning (see below)
- Exit can be used at program level to terminate the program
- In functions, exit without argument returns current `Result` value
- Exit is preferred over assigning Result and falling through

## Nested Loop Behavior

Break and continue statements only affect the **innermost** loop. To exit outer loops, you need to use flags or exit the containing function.

### Innermost Loop Only

```pascal
// Break only exits the inner loop
for i := 1 to 5 do begin
   for j := 1 to 5 do begin
      if i * j = 12 then
         break;  // Only exits inner j loop
   end;
   // Outer i loop continues
end;
```

### Breaking Outer Loops with Flags

```pascal
var found := false;
for i := 1 to 5 do begin
   if found then break;  // Break outer loop when flag is set
   for j := 1 to 5 do begin
      if i * j = 12 then begin
         PrintLn('Found: ', i, ' * ', j, ' = 12');
         found := true;
         break;  // Break inner loop
      end;
   end;
end;
```

### Breaking Outer Loops with Exit

If you need to exit multiple nested loops, use `exit` to exit the containing function:

```pascal
function FindProduct(target: Integer): Boolean;
var i, j: Integer;
begin
   Result := false;
   for i := 1 to 10 do
      for j := 1 to 10 do
         if i * j = target then begin
            PrintLn('Found: ', i, ' * ', j);
            Result := true;
            exit;  // Exits function, breaking all loops
         end;
end;
```

### Continue in Nested Loops

```pascal
// Continue only affects the inner loop
for i := 1 to 3 do begin
   for j := 1 to 3 do begin
      if j = 2 then continue;  // Skip j=2, but i loop continues
      Print('(', i, ',', j, ') ');
   end;
   PrintLn('');
end;
// Output: (1,1) (1,3) \n (2,1) (2,3) \n (3,1) (3,3)
```

## Interaction with Exception Handling

Break, continue, and exit statements work seamlessly with exception handling. **Finally blocks ALWAYS execute** before control is transferred, ensuring proper cleanup.

### Exit with Finally Block

The `finally` block executes even when `exit` is called:

```pascal
function Test(i: Integer): Boolean;
var
   ts: TObject;
begin
   Result := false;
   ts := TObject.Create;
   try
      if (i < 10) then exit;  // Finally block still executes!
   finally
      ts.Free;  // This ALWAYS runs
   end;
   Result := true;
end;
```

**Behavior**:
- Exit is called when `i < 10`
- Finally block executes `ts.Free`
- Function returns with `Result = false`
- The line `Result := true` is never reached

### Exit in Function with Finally

```pascal
function Hello: String;
begin
   Result := 'duh';
   try
      Result := 'Hello';
      if True then exit;
   finally
      Result := Result + ' world';  // Executes before return
   end;
   Result := 'Bye bye';  // Never reached
end;

PrintLn(Hello);  // Output: "Hello world"
```

**Execution Flow**:
1. Result set to "Hello"
2. Exit called
3. Finally block executes: Result becomes "Hello world"
4. Function returns "Hello world"

### Exit at Program Level with Finally

```pascal
try
   PrintLn('Try');
   exit;  // Exits the program
finally
   PrintLn('Finally');  // Executes before program exits
end;
PrintLn('Bug');  // Never reached
```

**Output**:
```
Try
Finally
```

### Break in Exception Handler

Break can be used inside exception handlers when the handler is inside a loop:

```pascal
while True do begin
   try
      raise Exception.Create('hello');
   except
      on e: Exception do begin
         PrintLn(e.Message);
         break;  // Exits the while loop
      end;
   end;
end;
```

### Break with Finally in Loop

```pascal
var i: Integer;
for i := 1 to 10 do begin
   try
      try
         raise Exception.Create('error');
      except
         on e: Exception do begin
            PrintLn(e.Message);
            break;  // Finally block executes first
         end;
      end;
   finally
      PrintLn('finally');  // Executes before break exits loop
   end;
   PrintLn(i);  // Never reached due to break
end;
```

### Continue with Finally

Similarly, `continue` also executes finally blocks:

```pascal
for i := 1 to 5 do begin
   try
      if i = 3 then continue;  // Finally executes before skipping
      PrintLn('Processing ', i);
   finally
      PrintLn('Cleanup ', i);  // Runs for all iterations, including i=3
   end;
end;
```

### Notes on Exception Handling

- **Finally blocks ALWAYS execute**, even with break/continue/exit
- This ensures proper resource cleanup
- The order is: finally block → then control transfer
- This behavior is consistent with Delphi/Object Pascal
- Exception handlers can use break/continue if inside loops

## Special Cases and Edge Cases

### Exit in Case Statement

Exit can be used in case statement branches to exit the containing function:

```pascal
procedure Test(s: String);
begin
   case s of
      't': begin
         PrintLn('here');
         exit;  // Exits the procedure
      end;
      'r': ;  // Empty case
   else
      PrintLn('there');
      exit;  // Exits from else branch
   end;
   PrintLn('done');  // Only reached if s='r'
end;

Test('a');  // Output: "there"
Test('r');  // Output: "done"
Test('t');  // Output: "here"
```

### Break in Case Statement (inside loop)

When a case statement is inside a loop, break exits the loop:

```pascal
for i := 1 to 10 do begin
   case i of
      5: begin
         PrintLn('Breaking at 5');
         break;  // Exits the for loop
      end;
   else
      PrintLn(i);
   end;
end;
// Output: 1, 2, 3, 4, "Breaking at 5"
```

### Unconditional While Loop with Break

Break is commonly used with unconditional while loops:

```pascal
while True do begin
   // ... do work ...
   if shouldExit then break;
end;
```

### Multiple Exit Points

Functions can have multiple exit statements:

```pascal
function Classify(n: Integer): String;
begin
   if n < 0 then exit('negative');
   if n = 0 then exit('zero');
   if n > 0 then exit('positive');
end;
```

### Break/Continue in Downto Loops

Break and continue work the same in downto loops:

```pascal
for i := 10 downto 1 do begin
   if (i and 1) = 0 then continue;  // Skip even numbers
   if i < 4 then break;             // Exit when i drops below 4
   PrintLn(i);
end;
// Output: 9, 7, 5
```

## Semantic Errors

The following uses are **invalid** and will cause semantic errors:

### Break Outside Loop

```pascal
// ERROR: break not allowed outside loop
procedure Test;
begin
   break;  // Semantic error
end;
```

### Continue Outside Loop

```pascal
// ERROR: continue not allowed outside loop
function Test: Integer;
begin
   Result := 0;
   continue;  // Semantic error
end;
```

### Exit Outside Function (in unit level code)

While exit at program level is valid, exit at unit level (outside any function) may not be valid in all contexts. Check semantic analysis rules.

## Summary Table

| Statement  | Valid Context      | Effect                                    | Finally Blocks |
|------------|--------------------|-------------------------------------------|----------------|
| `break`    | Loops only         | Exit innermost loop                       | Yes, executed  |
| `continue` | Loops only         | Skip to next iteration of innermost loop  | Yes, executed  |
| `exit`     | Functions/Program  | Exit function or program                  | Yes, executed  |

## Implementation Status

- **Lexer**: Tokens `BREAK`, `CONTINUE`, `EXIT` defined at `lexer/token_type.go:43-45`
- **Parser**: Not yet implemented
- **AST**: Not yet implemented
- **Semantic Analysis**: Not yet implemented
- **Interpreter**: Not yet implemented

See `PLAN.md` tasks 8.228-8.255 for implementation roadmap.

## References

- Reference implementation: `reference/dwscript-original/Test/SimpleScripts/break_continue.pas`
- Exit with finally: `reference/dwscript-original/Test/SimpleScripts/exit_finally.pas`
- Break in exception handlers: `reference/dwscript-original/Test/SimpleScripts/break_in_except_block.pas`
- Exit with return values: `reference/dwscript-original/Test/SimpleScripts/exit_result.pas`
- Test files in: `testdata/nested_loops.dws`, `testdata/exceptions/break_in_except.dws`
