# Array Type Compatibility Research

**Task**: 9.2.1
**Date**: 2025-11-15
**Status**: COMPLETED

## Executive Summary

This document details research into how DWScript handles static arrays and type aliases in var parameters. The research confirms that **type aliases are transparent** in go-dws, meaning that a type alias like `TData = array [0..1] of Integer` is treated as completely interchangeable with the underlying array type `array [0..1] of Integer` in all contexts, including var parameters.

## Key Questions and Answers

### 1. Are type aliases transparent?

**Answer**: YES

Type aliases in go-dws are completely transparent. The implementation in `internal/types/types.go` shows that `TypeAlias.Equals()` resolves both types to their underlying types before comparison:

```go
func (t *TypeAlias) Equals(other Type) bool {
    return GetUnderlyingType(t).Equals(GetUnderlyingType(other))
}
```

This means:
- `TData` (where `type TData = array [0..1] of Integer`)
- Is **exactly the same** as `array [0..1] of Integer`
- Can be used interchangeably everywhere

### 2. Should var parameters accept both the alias and the underlying type?

**Answer**: YES, and they already do

Testing confirms that var parameters work correctly with both:
- A var parameter declared as `var a : TData` accepts arguments of type `TData` OR `array [0..1] of Integer`
- A var parameter declared as `var a : array [0..1] of Integer` accepts arguments of type `array [0..1] of Integer` OR `TData`

See `testdata/array_alias_test.dws` for comprehensive test cases.

### 3. How does DWScript handle array type equality vs compatibility?

**Answer**: Strict structural equality with transparent type alias resolution

The implementation in `internal/types/compatibility.go` shows:

```go
func IsCompatible(from, to Type) bool {
    // Identical types are always compatible
    if from.Equals(to) {
        return true
    }

    // ... for arrays:
    fromArray, fromIsArray := from.(*ArrayType)
    toArray, toIsArray := to.(*ArrayType)
    if fromIsArray && toIsArray {
        // Element types must be IDENTICAL (not just compatible)
        if !fromArray.ElementType.Equals(toArray.ElementType) {
            return false
        }
        // Dynamic array can be assigned to dynamic array of same element type
        if fromArray.IsDynamic() && toArray.IsDynamic() {
            return true
        }
        // Static array can only be assigned to static array with same bounds
        return fromArray.Equals(toArray)
    }
}
```

**Rules**:
1. **Element types must be identical** (not just compatible)
   - `array of Integer` is NOT compatible with `array of Float`
2. **Dynamic arrays**: Must have same element type
   - `array of Integer` is compatible with `array of Integer`
3. **Static arrays**: Must have exact same bounds
   - `array [0..1] of Integer` is compatible with `array [0..1] of Integer`
   - `array [0..1] of Integer` is NOT compatible with `array [0..2] of Integer`
4. **Type aliases are resolved before comparison**
   - `TData` (alias for `array [0..1] of Integer`) IS compatible with `array [0..1] of Integer`

### 4. Are there any special rules for array slicing or bounds in var parameters?

**Answer**: Standard structural type checking applies

No special rules beyond the standard array compatibility rules. For var parameters:
- The argument must be an lvalue (variable, array element, or field)
- The type must match exactly (after resolving type aliases)
- Bounds must match for static arrays

## Implementation Details

### Type System Architecture

The type system uses several key components:

1. **TypeAlias struct** (`internal/types/types.go:212`)
   ```go
   type TypeAlias struct {
       AliasedType Type
       Name        string
   }
   ```

2. **GetUnderlyingType function** (`internal/types/types.go:250`)
   - Recursively resolves type aliases
   - Handles nested aliases: `type A = Integer; type B = A; type C = B;`

3. **ArrayType.Equals** (`internal/types/compound_types.go:39`)
   - Checks element type equality
   - Checks bounds match (for static arrays)
   - Automatically benefits from TypeAlias.Equals() resolution

### Parameter Type Checking Flow

1. Function call in `internal/semantic/analyze_function_calls.go:93`
   ```go
   argType := a.analyzeExpressionWithExpectedType(arg, paramType)
   if argType != nil && !a.canAssign(argType, paramType) {
       a.addError("argument %d has type %s, expected %s at %s", ...)
   }
   ```

2. Assignment compatibility in `internal/semantic/analyzer.go:436`
   ```go
   func (a *Analyzer) canAssign(from, to types.Type) bool {
       if from == nil || to == nil {
           return false
       }
       if types.IsCompatible(from, to) {
           return true
       }
       // ... additional checks for nil, classes, etc.
   }
   ```

3. Type compatibility in `internal/types/compatibility.go:27`
   ```go
   func IsCompatible(from, to Type) bool {
       // Identical types are always compatible
       if from.Equals(to) {
           return true  // This is where type aliases get resolved!
       }
       // ... array-specific logic ...
   }
   ```

### Var Parameter Lvalue Checking

In addition to type compatibility, var parameters require the argument to be an lvalue. This is checked in `internal/semantic/analyze_function_calls.go:84-89`:

```go
isVar := len(funcType.VarParams) > i && funcType.VarParams[i]
if isVar && !a.isLValue(arg) {
    a.addError("var parameter %d requires a variable (identifier, array element, or field), got %s at %s",
        i+1, arg.String(), arg.Pos().String())
}
```

## Test Cases

### Existing Tests

1. **testdata/fixtures/ArrayPass/array_static_var_param.pas**
   - Tests var parameters with type alias `TData`
   - Demonstrates passing `TData` variables to var parameters

2. **testdata/fixtures/ArrayPass/array_static_param.pas**
   - Tests const/value parameters with type alias `TArr`
   - Shows passing `TArr` constants and variables

### New Test Cases Created

**testdata/array_alias_test.dws** - Comprehensive test of type alias transparency:

1. **Test 1**: Type alias to type alias
   - `var data : TData` passed to `var a : TData`
   - ✅ PASS

2. **Test 2**: Underlying type to type alias
   - `var data : array [0..1] of Integer` passed to `var a : TData`
   - ✅ PASS (proves aliases are transparent)

3. **Test 3**: Type alias to underlying type
   - `var data : TData` passed to `var a : array [0..1] of Integer`
   - ✅ PASS (proves aliases work both directions)

4. **Test 4**: Underlying type to underlying type
   - `var data : array [0..1] of Integer` passed to `var a : array [0..1] of Integer`
   - ✅ PASS

## Conclusions

### What Works

✅ Type aliases are fully transparent in go-dws
✅ Var parameters accept both alias and underlying types
✅ Array type equality correctly resolves aliases
✅ No special handling needed for array bounds in var parameters

### Implementation Quality

The implementation is **correct and complete** for type alias transparency:

1. **Clean separation of concerns**:
   - Type alias resolution in `TypeAlias.Equals()`
   - Array comparison in `ArrayType.Equals()`
   - Compatibility checking in `IsCompatible()`

2. **Recursive alias resolution**:
   - Handles nested aliases correctly
   - Works for all type constructs, not just arrays

3. **Consistent behavior**:
   - Type aliases work the same everywhere
   - No special cases needed for var parameters

### No Issues Found

The research was initiated to understand potential issues with array type compatibility. However, testing reveals that the current implementation handles all cases correctly. Type aliases are properly transparent, and there are no compatibility issues between aliased and non-aliased array types in var parameters.

## References

### Source Files Examined

1. `internal/types/types.go` - TypeAlias implementation
2. `internal/types/compatibility.go` - Type compatibility rules
3. `internal/types/compound_types.go` - ArrayType implementation
4. `internal/semantic/analyzer.go` - Semantic analysis entry points
5. `internal/semantic/analyze_function_calls.go` - Function call type checking

### Test Files Examined

1. `testdata/fixtures/ArrayPass/array_static_var_param.pas`
2. `testdata/fixtures/ArrayPass/array_static_param.pas`
3. `testdata/array_alias_test.dws` (created during research)

### External Resources

- DWScript GitHub: https://github.com/EricGrange/DWScript
- DWScript language reference: https://www.delphitools.info/dwscript/
- DelphiTools blog: https://www.delphitools.info/

### Related Reading

- "The case for strict parameter types" - DelphiTools blog
  - Discusses DWScript's `type` parameter qualifier for strict type checking
  - Our implementation uses standard DWScript behavior (type aliases are transparent)
  - Strict type checking (forbidding automatic type casts) is a different feature

## Next Steps

This research completes Task 9.2.1. The findings show that:

1. No implementation changes are needed for type alias transparency
2. The existing implementation correctly handles all test cases
3. Type aliases work as expected in DWScript (transparent/structural equality)

For Task 9.2.2 and beyond, if there are specific compatibility issues, they likely involve:
- Different array bounds (e.g., `[0..1]` vs `[0..2]`)
- Different element types (e.g., `Integer` vs `Float`)
- Not type alias transparency (which works correctly)
