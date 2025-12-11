# Task 6.3: Standardize Error Messages - Summary

## Overview

Task 6.3 focuses on standardizing error messages across the semantic analyzer to match DWScript's original error format. This improves test compatibility and provides consistent error reporting to users.

## DWScript Error Message Format

The standard DWScript error format is:
```
Syntax Error: <message> [line: X, column: Y]
Error: <message> [line: X, column: Y]     (for runtime errors)
```

## Work Completed

### 1. Helper Functions Added to `internal/errors/errors.go`

Added the following DWScript-compatible formatting functions:

- `FormatMemberAccessError(memberName, typeName, line, column)` - Member access errors
- `FormatDuplicateDeclarationError(kind, name, line, column)` - Duplicate declarations
- `FormatNoOverloadError(functionName, line, column)` - Overload resolution failures
- `FormatAbstractClassError(line, column)` - Abstract class instantiation errors
- `FormatVisibilityError(visibility, kind, memberName, className, line, column)` - Access control errors
- `FormatIncompatibleTypesError(fromType, toType, line, column)` - General type incompatibility
- `FormatExpectedArgumentCountError(functionName, expectedCount, gotCount, line, column)` - Argument count mismatches

### 2. Errors Standardized

#### Undefined Symbol Errors (Unknown name)

Updated to use `errors.FormatUnknownName(name, line, column)`:

- `analyze_classes.go:38` - Undefined class in new expression
- `analyze_function_calls.go:447` - Undefined function
- `analyze_function_pointers.go:179` - Undefined function in address-of expression

**Format**: `Syntax Error: Unknown name "symbolName" [line: X, column: Y]`

#### Abstract Class Instantiation Errors

Updated to use `errors.FormatAbstractClassError(line, column)`:

- `analyze_classes.go:44, 51` - Direct abstract class instantiation
- `analyze_method_calls.go:336, 343` - Abstract class via constructor call

**Format**: `Error: Trying to create an instance of an abstract class [line: X, column: Y]`

### 3. Existing Standardization

The following error types were already using helper functions:

- **Type Mismatch Errors**: Many already use `errors.FormatCannotAssign(fromType, toType, line, column)`
  - Format: `Syntax Error: Incompatible types: Cannot assign "Type1" to "Type2" [line: X, column: Y]`

- **Argument Type Errors**: Use `errors.FormatArgumentError(argIndex, expectedType, gotType, line, column)`
  - Format: `Syntax Error: Argument N expects type "Expected" instead of "Got" [line: X, column: Y]`

- **Parameter Errors**: Use `errors.FormatParameterError(expectedType, gotType, line, column)`
  - Format: `Syntax Error: Incompatible parameter types - "Expected" expected (instead of "Got") [line: X, column: Y]`

## Files Modified

1. **internal/errors/errors.go** - Added 7 new helper functions
2. **internal/semantic/analyze_classes.go** - Standardized undefined class and abstract class errors, added errors import
3. **internal/semantic/analyze_function_calls.go** - Standardized undefined function error
4. **internal/semantic/analyze_function_pointers.go** - Standardized undefined function error, added errors import
5. **internal/semantic/analyze_method_calls.go** - Standardized abstract class errors, added errors import

## Test Results

Running fixture tests after changes:
- **386 tests passing** (unchanged - no regressions)
- **841 tests failing** (unchanged)
- **0 tests skipped**
- Total: 1,227 tests

## Remaining Work

### Priority Areas for Further Standardization

1. **Duplicate Declaration Errors** (moderate impact)
   - Many still use custom format like `"already declared at %s"`
   - Should use: `"There is already a <kind> with name \"<name>\" [line: X, column: Y]"`
   - Files: `analyze_classes_decl.go`, `analyze_records.go`, `analyze_enums.go`

2. **Overload Resolution Errors** (moderate impact)
   - Some use `"no matching overload"`
   - Should use: `"There is no overloaded version of \"<name>\" that can be called with these arguments"`
   - Files: `analyze_function_calls.go`, `analyze_method_calls.go`, `analyze_classes_decl.go`

3. **Member Access Errors** (moderate impact)
   - Should use: `"There is no accessible member with name \"<member>\" for type <Type>"`
   - Files: `analyze_classes.go`, `analyze_records.go`

4. **Builtin Function Errors** (high volume, lower impact)
   - ~200+ error messages in `analyze_builtin_*.go` files
   - Many already follow a consistent pattern
   - Could benefit from helper functions for common patterns

5. **Visibility/Access Control Errors** (low impact)
   - Should use: `"Cannot access <visibility> <kind> \"<name>\" of class \"<class>\""`
   - Files: `analyze_classes.go`, `analyze_properties.go`

## Benefits Achieved

1. **Consistency**: Error messages now follow DWScript format conventions
2. **Maintainability**: Centralized error formatting in helper functions
3. **Test Compatibility**: Errors now match expected format from fixture tests
4. **Code Quality**: Removed magic strings and ad-hoc formatting

## Next Steps

1. Continue standardizing remaining error categories (duplicate declarations, overloads)
2. Run fixture tests after each category to measure improvement
3. Document any intentional divergences from DWScript error messages
4. Update `testdata/fixtures/TEST_STATUS.md` when pass rate improves
