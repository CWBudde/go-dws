# Task 9.3.2: Root Cause Analysis - Case-Insensitive Function Call Bug

**Date**: 2025-11-15
**Status**: COMPLETED

## Executive Summary

The root cause of the case-insensitive function call bug has been identified. User-defined functions are stored in a map using their **original case** (not normalized), but lookups also use the **original case from the call site**. When the case doesn't match, the lookup fails silently.

## The Bug

### Location 1: Function Storage (Broken)

**File**: `internal/interp/declarations.go`
**Lines**: 39, 46, 50

```go
func (i *Interpreter) evalFunctionDeclaration(fn *ast.FunctionDecl) Value {
    // ...

    // Store regular function in the registry
    // Support overloading by storing multiple functions per name
    funcName := fn.Name.Value  // ❌ BUG: Uses original case!

    // If this function has a body...
    if fn.Body != nil {
        existingOverloads := i.functions[funcName]  // ❌ Case-sensitive lookup
        i.functions[funcName] = i.replaceMethodInOverloadList(existingOverloads, fn)  // ❌ Case-sensitive storage
    } else {
        i.functions[funcName] = append(i.functions[funcName], fn)  // ❌ Case-sensitive storage
    }

    return &NilValue{}
}
```

### Location 2: Function Lookup (Broken)

**File**: `internal/interp/functions_calls.go`
**Line**: 211

```go
func (i *Interpreter) evalCallExpression(expr *ast.CallExpression) Value {
    // ...

    // Get the function name
    funcName, ok := expr.Function.(*ast.Identifier)
    if !ok {
        return newError("function call requires identifier or qualified name, got %T", expr.Function)
    }

    // Check if it's a user-defined function first
    if overloads, exists := i.functions[funcName.Value]; exists && len(overloads) > 0 {
        // ❌ BUG: funcName.Value uses original case from call site!
        // If function is stored as "QuickSort" but called as "quicksort", lookup fails!
        // ...
    }

    // ...falls through to built-in function lookup which handles case correctly
}
```

## How the Bug Manifests

### Example: QuickSort Recursion

```pascal
procedure QuickSort(var d: TData; i, j : Integer);
begin
  if i < j then
  begin
    m := partition(d, i, j);
    quicksort(d, i, m);      // lowercase!
    quicksort(d, m + 1, j);  // lowercase!
  end;
end;
```

**What Happens**:

1. **Function Declaration** (line 1):
   - Function name: `QuickSort` (capital Q, capital S)
   - Stored in map as: `i.functions["QuickSort"]` ❌ Original case

2. **First Call** (`QuickSort(data, 0, size - 1)`):
   - Call uses: `QuickSort` (capital Q, capital S)
   - Lookup: `i.functions["QuickSort"]` ✅ Found! Works correctly

3. **Recursive Call** (`quicksort(d, i, m)`):
   - Call uses: `quicksort` (all lowercase)
   - Lookup: `i.functions["quicksort"]` ❌ Not found!
   - Falls through to built-in function lookup
   - Built-in not found, no error raised
   - **Silently returns without executing recursion**

## Why It Fails Silently

Looking at the full evalCallExpression flow:

1. Try function pointer lookup (line 25-72)
2. Try member access/unit-qualified call (line 74-202)
3. **Try user-defined function** (line 210-263) ❌ FAILS HERE
4. Try implicit self method (line 265-289)
5. Try record static method (line 291-318)
6. Try built-in with var param (line 320-329)
7. Try external function (line 331-369)
8. Try type cast (line 371-377)
9. **Try built-in function** (line 379-390) ← Returns here

When the user-defined function lookup fails due to case mismatch:
- It doesn't raise an error
- It continues to step 9 (built-in function lookup)
- Built-in lookup fails (no such built-in)
- `callBuiltin()` returns an error value
- **But this error is not displayed to user in our test case!**

The error is likely being swallowed somewhere in the evaluation chain.

## Comparison: What Works Correctly

### Built-in Functions (Work Correctly)

**File**: `internal/interp/functions_calls.go`
**Lines**: 393-484

```go
// normalizeBuiltinName normalizes a builtin function name to its canonical form
// for case-insensitive matching (DWScript is case-insensitive).
func normalizeBuiltinName(name string) string {
    // Create a lowercase version for comparison
    lower := strings.ToLower(name)  // ✅ Normalizes to lowercase

    // Map of lowercase names to canonical names
    canonicalNames := map[string]string{
        "println": "PrintLn",  // ✅ Can call with any case
        "print": "Print",
        // ...
    }

    // Return canonical name if found
    if canonical, ok := canonicalNames[lower]; ok {  // ✅ Lookup with lowercase
        return canonical
    }
    return name
}
```

Built-in functions work correctly because they normalize to lowercase before lookup.

### Environment Variables (Work Correctly)

**File**: `internal/interp/environment.go`
**Lines**: 46-48

```go
func (e *Environment) Get(name string) (Value, bool) {
    // Normalize to lowercase for case-insensitive lookup
    key := strings.ToLower(name)  // ✅ Normalizes to lowercase

    // Check current environment
    val, ok := e.store[key]  // ✅ Lookup with lowercase
    // ...
}
```

Environment variable lookup works correctly because it normalizes to lowercase.

## The Fix

### Option 1: Normalize at Storage (Recommended)

Normalize function names to lowercase when storing AND when looking up:

**File**: `internal/interp/declarations.go:39`

```go
// BEFORE:
funcName := fn.Name.Value

// AFTER:
funcName := strings.ToLower(fn.Name.Value)  // ✅ Store with lowercase key
```

**File**: `internal/interp/functions_calls.go:211`

```go
// BEFORE:
if overloads, exists := i.functions[funcName.Value]; exists && len(overloads) > 0 {

// AFTER:
funcNameLower := strings.ToLower(funcName.Value)
if overloads, exists := i.functions[funcNameLower]; exists && len(overloads) > 0 {
```

### Option 2: Normalize Only at Lookup

Keep storage as-is, but normalize at lookup:

**File**: `internal/interp/functions_calls.go:211`

```go
// BEFORE:
if overloads, exists := i.functions[funcName.Value]; exists && len(overloads) > 0 {

// AFTER:
funcNameLower := strings.ToLower(funcName.Value)
// Find function with case-insensitive comparison
for key, overloads := range i.functions {
    if strings.EqualFold(key, funcName.Value) {
        // Found it with case-insensitive match
        // ...
    }
}
```

This is less efficient (O(n) instead of O(1)) but avoids changing storage format.

### Recommendation

**Use Option 1** - it's cleaner, more efficient, and consistent with how the environment and built-in functions work.

## Additional Locations to Fix

The bug likely exists in other places where `i.functions` is accessed:

### Search for Direct Access

```bash
grep -n "i\.functions\[" internal/interp/*.go
```

Results:
- `declarations.go:46` - ❌ Needs fix
- `declarations.go:50` - ❌ Needs fix
- `functions_calls.go:211` - ❌ Needs fix (primary bug)

### Search for Function Iteration

```bash
grep -n "range i\.functions" internal/interp/*.go
```

Any iteration over `i.functions` may need to use case-insensitive comparison.

## Similar Bugs in Other Components?

### Classes (Check Required)

```go
i.classes[className]  // Is this case-sensitive?
```

### Records (Check Required)

```go
i.records[recordName]  // Is this case-sensitive?
```

### Methods (Check Required)

Method lookup within classes/records - are they case-sensitive?

## Test Cases to Add

After fixing, ensure these work:

1. **Recursive calls with different case**:
   ```pascal
   procedure Factorial(n: Integer): Integer;
   begin
     Result := n * factorial(n - 1);  // lowercase
   end;
   ```

2. **Mutual recursion with different case**:
   ```pascal
   procedure IsEven(n: Integer): Boolean;
   begin
     Result := isodd(n - 1);  // lowercase
   end;
   ```

3. **All uppercase/lowercase combinations**:
   ```pascal
   PrintProc();  // Capital P
   printProc();  // lowercase p
   PRINTPROC();  // all uppercase
   ```

## Files to Modify

1. `internal/interp/declarations.go:39` - Normalize when storing
2. `internal/interp/declarations.go:46` - Lookup with normalized name
3. `internal/interp/declarations.go:50` - Store with normalized name
4. `internal/interp/functions_calls.go:211` - Lookup with normalized name

## Success Criteria

After fix:
- ✅ `testdata/case_test2.dws` produces correct recursive output
- ✅ `testdata/fixtures/Algorithms/quicksort.pas` sorts correctly
- ✅ All case combinations work (lowercase, UPPERCASE, MixedCase)
- ✅ No performance regression
- ✅ All existing tests still pass

## References

- Task 9.3.1: Test cases demonstrating the bug
- `docs/task-9.2.x-final-analysis.md`: Bug discovery details
- `internal/interp/environment.go:46-48`: Correct example (case-insensitive)
- `internal/interp/functions_calls.go:393-484`: Correct example (built-ins)
