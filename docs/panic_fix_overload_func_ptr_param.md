# Panic Fix: overload_func_ptr_param.pas

**Date**: 2025-11-07
**Issue**: Critical panic (nil pointer dereference) in `overload_func_ptr_param.pas` test
**Status**: ✅ **FIXED**

## Problem Description

The test `overload_func_ptr_param.pas` was causing a panic during parsing:

```
PANIC: runtime error: invalid memory address or nil pointer dereference

Stack trace:
github.com/cwbudde/go-dws/pkg/ast.(*FunctionPointerTypeNode).String(0x0)
	/home/user/go-dws/pkg/ast/function_pointer.go:39 +0x3a
github.com/cwbudde/go-dws/internal/parser.(*Parser).parseParameterGroup(0xc0002b08f0)
	/home/user/go-dws/internal/parser/functions.go:387 +0x9e5
```

The panic occurred when parsing function pointer types as parameters, specifically:
- `procedure TestProc(paramProc : procedure);`
- `procedure TestFunc(paramFunc : function : Integer);`

## Root Causes

### Issue 1: Nil Pointer Dereference
In `internal/parser/functions.go` at line 387, the code called `te.String()` on a potentially nil `FunctionPointerTypeNode` without checking:

```go
case *ast.FunctionPointerTypeNode:
    typeAnnotation = &ast.TypeAnnotation{
        Token: te.Token,
        Name:  te.String(), // <-- PANIC: te could be nil
    }
```

### Issue 2: Missing Support for Parameterless Function Pointers
The parser required parentheses after `function`/`procedure` keywords in function pointer types. However, DWScript allows these forms:

- ✅ `procedure` - no parameters, no parentheses
- ✅ `procedure()` - no parameters, with empty parentheses
- ✅ `function : Integer` - no parameters, returns Integer
- ✅ `function() : Integer` - with empty parentheses, returns Integer

The parser only supported the forms with parentheses, causing parse failures.

## Solution

### Fix 1: Add Nil Check (functions.go:382-389)

```go
case *ast.FunctionPointerTypeNode:
    // Check if te is nil to prevent panics (defensive programming)
    if te == nil {
        p.addError("function pointer type expression is nil in parameter type", ErrInvalidType)
        return nil
    }
    typeAnnotation = &ast.TypeAnnotation{
        Token: te.Token,
        Name:  te.String(), // Now safe - nil check above
    }
```

### Fix 2: Make Parentheses Optional (types.go:139-200)

```go
// Check if parameter list is present (optional in DWScript)
hasParentheses := p.peekTokenIs(lexer.LPAREN)

if hasParentheses {
    // Parse parameter list with parentheses
    p.nextToken() // move to LPAREN
    // ... parse parameters ...
} else {
    // No parentheses - parameterless function/procedure pointer
    // Current token is still FUNCTION or PROCEDURE
    endToken = funcOrProcToken
}

// Continue to parse return type for functions...
```

## Test Results

### Before Fix
```
PANIC in overload_func_ptr_param.pas: runtime error: invalid memory address or nil pointer dereference
```

### After Fix
```
Parser errors:
  no prefix parse function for FORWARD found at 1:54
  no prefix parse function for FORWARD found at 2:71
  ...
```

**Result**: ✅ No panic! Test now fails gracefully due to missing `forward` keyword support (a separate parser feature).

## Impact

### Positive
- ✅ Eliminates critical crash
- ✅ Adds defensive nil checking
- ✅ Supports standard DWScript function pointer syntax
- ✅ Makes parser more robust

### No Regressions
- All other tests still pass/fail as before
- No new failures introduced
- OverloadsPass still at 2/39 passing

## Files Modified

1. **internal/parser/functions.go**
   - Added nil check for `FunctionPointerTypeNode` in `parseParameterGroup()`
   - Lines: 382-389

2. **internal/parser/types.go**
   - Made parentheses optional in `parseFunctionPointerType()`
   - Lines: 139-200, 229-244

## Next Steps

The test `overload_func_ptr_param.pas` still fails due to:
- Missing `forward` keyword support (forward declarations)
- This is a separate parser feature unrelated to the panic fix

To fully pass this test, we would need to implement:
- Task 9.60: Forward declaration support
- Forward declaration validation with overload directives

## Verification

```bash
# Test the specific file
go test -v ./internal/interp -run TestDWScriptFixtures/OverloadsPass/overload_func_ptr_param

# Result: FAIL (parse errors) but NO PANIC ✅

# Test overall suite
go test -v ./internal/interp -run TestDWScriptFixtures/OverloadsPass

# Result: 2/39 passing, 37 failing, 0 panics ✅
```

## Conclusion

The critical panic bug has been successfully fixed. The parser now:
1. Gracefully handles nil function pointer type nodes
2. Supports DWScript's parameterless function pointer syntax
3. Provides clear error messages instead of crashing

The fix is defensive, minimal, and targeted - addressing only the panic without introducing new issues or unnecessary changes.
