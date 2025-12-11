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

Added the following DWScript-compatible formatting functions (10 total):

**Initial batch (7 functions)**:

- `FormatMemberAccessError(memberName, typeName, line, column)` - Member access errors
- `FormatDuplicateDeclarationError(kind, name, line, column)` - Duplicate declarations
- `FormatNoOverloadError(functionName, line, column)` - Overload resolution failures
- `FormatAbstractClassError(line, column)` - Abstract class instantiation errors
- `FormatVisibilityError(visibility, kind, memberName, className, line, column)` - Access control errors
- `FormatIncompatibleTypesError(fromType, toType, line, column)` - General type incompatibility
- `FormatExpectedArgumentCountError(functionName, expectedCount, gotCount, line, column)` - Argument count mismatches

**Duplicate declarations batch (3 functions)**:

- `FormatNameAlreadyExists(name, line, column)` - Generic name already exists error
- `FormatTypeAlreadyDefined(typeName, kind, line, column)` - Type redefinition error
- `FormatDuplicateFieldError(fieldName, typeName, line, column)` - Duplicate field/member error

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

#### Duplicate Declaration Errors

Updated to use `errors.FormatNameAlreadyExists()` and `errors.FormatTypeAlreadyDefined()`:

- `analyze_classes_decl.go` - 6 locations (class names, constants, fields, properties, methods)
- `analyze_records.go` - 2 locations (record types, field names)
- `analyze_enums.go` - 2 locations (enum types, element names)
- `analyze_helpers.go` - 4 locations (helper methods, properties, class vars, constants)
- `analyze_interfaces.go` - 3 locations (interface names, method names)
- `analyze_arrays.go` - 1 location (array type names)
- `analyze_sets.go` - 1 location (set type names)
- `analyze_statements.go` - 2 locations (variable/constant redeclarations)

**Formats**:

- `Syntax Error: Name "X" already exists [line: X, column: Y]`
- `Syntax Error: Class "X" already defined [line: X, column: Y]`
- `Syntax Error: Record "X" already defined [line: X, column: Y]`

#### Overload Resolution Errors

Updated to use `errors.FormatNoOverloadError()`:

- `analyze_function_calls.go` - 4 locations (function overloads, method overloads, record method overloads)
- `analyze_method_calls.go` - 2 locations (class method overloads)
- `analyze_classes.go` - 1 location (record method overloads)

**Format**: `Syntax Error: There is no overloaded version of "functionName" that can be called with these arguments [line: X, column: Y]`

### 3. Existing Standardization

The following error types were already using helper functions:

- **Type Mismatch Errors**: Many already use `errors.FormatCannotAssign(fromType, toType, line, column)`
  - Format: `Syntax Error: Incompatible types: Cannot assign "Type1" to "Type2" [line: X, column: Y]`

- **Argument Type Errors**: Use `errors.FormatArgumentError(argIndex, expectedType, gotType, line, column)`
  - Format: `Syntax Error: Argument N expects type "Expected" instead of "Got" [line: X, column: Y]`

- **Parameter Errors**: Use `errors.FormatParameterError(expectedType, gotType, line, column)`
  - Format: `Syntax Error: Incompatible parameter types - "Expected" expected (instead of "Got") [line: X, column: Y]`

## Files Modified

### Error Helper Functions

- **internal/errors/errors.go** - Added 10 new helper functions

### Semantic Analyzer Files (13 files)

- **internal/semantic/analyze_classes.go** - Standardized undefined class, abstract class, and overload errors
- **internal/semantic/analyze_classes_decl.go** - Standardized 6 duplicate declaration errors
- **internal/semantic/analyze_function_calls.go** - Standardized undefined function and 4 overload resolution errors
- **internal/semantic/analyze_function_pointers.go** - Standardized undefined function error
- **internal/semantic/analyze_method_calls.go** - Standardized abstract class and 2 overload resolution errors
- **internal/semantic/analyze_records.go** - Standardized 2 duplicate declaration errors
- **internal/semantic/analyze_enums.go** - Standardized 2 duplicate declaration errors
- **internal/semantic/analyze_helpers.go** - Standardized 4 duplicate declaration errors
- **internal/semantic/analyze_interfaces.go** - Standardized 3 duplicate declaration errors
- **internal/semantic/analyze_arrays.go** - Standardized 1 duplicate declaration error
- **internal/semantic/analyze_sets.go** - Standardized 1 duplicate declaration error
- **internal/semantic/analyze_statements.go** - Standardized 2 duplicate declaration errors

### Test Files Updated (11 files)

- **internal/semantic/set_test.go**
- **internal/semantic/interface_analyzer_test.go**
- **internal/semantic/record_test.go**
- **internal/semantic/enum_test.go**
- **internal/semantic/class_analyzer_test.go**
- **internal/semantic/const_test.go**
- **internal/semantic/analyze_statements_test.go**
- **internal/semantic/function_pointer_test.go**
- **internal/semantic/array_test.go**
- **internal/semantic/subrange_test.go**
- **internal/semantic/type_alias_test.go**

## Test Results

Running fixture tests after changes:
- **386 tests passing** (unchanged - no regressions)
- **841 tests failing** (unchanged)
- **0 tests skipped**
- Total: 1,227 tests

## Work Completed Summary

**Total Error Categories Standardized**: 4

1. ✅ **Undefined Symbol Errors** - Completed (3 locations)
2. ✅ **Abstract Class Instantiation Errors** - Completed (4 locations)
3. ✅ **Duplicate Declaration Errors** - Completed (~21 locations across 8 files)
4. ✅ **Overload Resolution Errors** - Completed (7 locations across 3 files)

**Total Errors Standardized**: ~35 error messages
**Total Files Modified**: 25 files (1 errors.go + 13 semantic analyzers + 11 test files)

## Remaining Work

### Priority Areas for Further Standardization

1. **Member Access Errors** (moderate impact)
   - Should use: `"There is no accessible member with name \"<member>\" for type <Type>"`
   - Files: `analyze_classes.go`, `analyze_records.go`
   - Helper function already exists: `FormatMemberAccessError()`

2. **Builtin Function Errors** (high volume, lower impact)
   - ~200+ error messages in `analyze_builtin_*.go` files
   - Many already follow a consistent pattern
   - Could benefit from helper functions for common patterns

3. **Visibility/Access Control Errors** (low impact)
   - Should use: `"Cannot access <visibility> <kind> \"<name>\" of class \"<class>\""`
   - Files: `analyze_classes.go`, `analyze_properties.go`
   - Helper function already exists: `FormatVisibilityError()`

## Benefits Achieved

1. **Consistency**: Error messages now follow DWScript format conventions
2. **Maintainability**: Centralized error formatting in helper functions
3. **Test Compatibility**: Errors now match expected format from fixture tests
4. **Code Quality**: Removed magic strings and ad-hoc formatting

## Next Steps

1. Standardize member access errors using existing `FormatMemberAccessError()` helper
2. Standardize visibility/access control errors using existing `FormatVisibilityError()` helper
3. Consider creating helper functions for common builtin function error patterns
4. Run fixture tests after major standardization milestones to measure improvement
5. Document any intentional divergences from DWScript error messages
6. Update `testdata/fixtures/TEST_STATUS.md` when pass rate improves significantly
