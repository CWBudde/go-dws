# Task 3.5.99b: JSON Indexing Migration - COMPLETE

**Date**: 2025-11-25
**Status**: ✅ COMPLETE
**Time**: ~60 minutes
**PR**: Ready for review

## Overview

Task 3.5.99b migrated JSON indexing logic from the Interpreter to the Evaluator, eliminating one EvalNode delegation point in `VisitIndexExpression`. The implementation uses reflection to access internal JSONValue fields, avoiding circular import issues.

## Changes Made

### 1. JSON Helper Functions (`internal/interp/evaluator/json_helpers.go`)

Created new file with JSON indexing logic:

**Key Function**: `indexJSON(base Value, index Value, node ast.Node) Value`
- Uses reflection to extract `*jsonvalue.Value` from `JSONValue` wrapper
- Handles JSON objects (string keys) and JSON arrays (integer indices)
- Returns results wrapped in VariantValue via adapter

**Implementation Details**:
```go
func (e *Evaluator) indexJSON(base Value, index Value, node ast.Node) Value {
    // Extract jsonvalue.Value using reflection (avoids circular import)
    jv := extractJSONValueViaReflection(base)

    // Handle JSON objects
    if jv.Kind() == jsonvalue.KindObject {
        // String index required
        result := jv.ObjectGet(key)
        return e.adapter.WrapJSONValueInVariant(result)
    }

    // Handle JSON arrays
    if jv.Kind() == jsonvalue.KindArray {
        // Integer index required
        result := jv.ArrayGet(int(idx))
        return e.adapter.WrapJSONValueInVariant(result)
    }
}
```

**Reflection Helper**: `extractJSONValueViaReflection(val Value) *jsonvalue.Value`
- Uses `reflect` package to access `JSONValue.Value` field
- Avoids importing parent `interp` package (circular import prevention)
- Returns nil if value is not a JSONValue

### 2. Adapter Interface Extension (`internal/interp/evaluator/evaluator.go`)

Added new method to `InterpreterAdapter`:
```go
type InterpreterAdapter interface {
    // ... existing methods ...

    // WrapJSONValueInVariant wraps a jsonvalue.Value in a VariantValue.
    // Task 3.5.99b: Required for JSON indexing without circular imports.
    WrapJSONValueInVariant(jv any) Value
}
```

### 3. Adapter Implementation (`internal/interp/interpreter.go`)

Implemented adapter method:
```go
func (i *Interpreter) WrapJSONValueInVariant(jv any) Value {
    jsonVal, ok := jv.(*jsonvalue.Value)
    if !ok && jv != nil {
        return &ErrorValue{Message: "invalid type passed..."}
    }
    return jsonValueToVariant(jsonVal)
}
```

### 4. VisitIndexExpression Update (`internal/interp/evaluator/visitor_expressions.go`)

Modified lines 2038-2049:

**Before** (with EvalNode delegation):
```go
// Check if it's a string
if leftVal.Type() == "STRING" {
    // ... string indexing ...
}

// Otherwise delegate to interpreter
return e.adapter.EvalNode(node)
```

**After** (direct JSON handling):
```go
// Handle JSON indexing directly - Task 3.5.99b
if leftVal.Type() == "JSON" {
    return e.indexJSON(leftVal, indexVal, node)
}

// Handle interface/object default property access - delegate to adapter
switch leftVal.Type() {
case "INTERFACE", "OBJECT", "RECORD":
    return e.adapter.EvalNode(node)
}
```

**Key Change**: Removed EvalNode delegation for JSON case (previously at line 2046)

## Files Modified

1. `internal/interp/evaluator/json_helpers.go` - New file (109 lines)
2. `internal/interp/evaluator/evaluator.go` - Added WrapJSONValueInVariant to adapter
3. `internal/interp/interpreter.go` - Implemented WrapJSONValueInVariant adapter method
4. `internal/interp/evaluator/visitor_expressions.go` - Updated VisitIndexExpression

## Test Results

All tests pass:

### Unit Tests
```bash
$ go test ./internal/interp/evaluator
ok      github.com/cwbudde/go-dws/internal/interp/evaluator    0.003s

$ go test ./internal/interp -run "JSON|Property"
PASS
```

### Integration Tests (JSON scripts)
```bash
$ ./bin/dwscript run testdata/json/json_object_access.dws
Property access tests:
John
30
NYC
Unassigned

Array indexing tests:
10
30
50
Unassigned

[... all tests pass ...]

$ ./bin/dwscript run testdata/json/json_array_access.dws
=== JSON Array Access Tests ===
[... all tests pass ...]
=== All Array Access Tests Passed ===
```

## Technical Approach

### Circular Import Avoidance

The challenge: `JSONValue` is defined in `internal/interp`, but Evaluator is in `internal/interp/evaluator` (child package). We cannot import parent package.

**Solution**:
1. Use reflection to extract `jsonvalue.Value` from `JSONValue` without importing parent
2. Delegate variant wrapping to adapter (Interpreter has access to both packages)
3. Keep evaluator focused on evaluation logic, not type conversions

### JSON Object Indexing

```go
// JSON object: {"name": "Alice", "age": 30}
obj['name']  → ObjectGet("name") → VariantValue wrapping "Alice"
obj['age']   → ObjectGet("age")  → VariantValue wrapping 30
obj['missing'] → ObjectGet("missing") → VariantValue wrapping nil
```

### JSON Array Indexing

```go
// JSON array: [10, 20, 30]
arr[0] → ArrayGet(0) → VariantValue wrapping 10
arr[2] → ArrayGet(2) → VariantValue wrapping 30
arr[99] → ArrayGet(99) → VariantValue wrapping nil (out of bounds)
```

## Benefits

1. **EvalNode Elimination**: Removed 1 of 4 remaining EvalNode delegation points
2. **Clean Separation**: JSON logic now in evaluator, not spread across interpreter
3. **No Circular Imports**: Reflection-based approach avoids import cycles
4. **Type Safety**: Validates JSON value types before indexing
5. **Error Handling**: Clear error messages for type mismatches

## Acceptance Criteria

- ✅ JSON object indexing works (string keys)
- ✅ JSON array indexing works (integer indices)
- ✅ Missing properties return nil/Unassigned
- ✅ Out-of-bounds array access returns nil/Unassigned
- ✅ Type errors handled (e.g., indexing object with integer)
- ✅ EvalNode delegation removed for JSON case
- ✅ All existing tests pass
- ✅ No circular imports
- ✅ Integration tests verify end-to-end functionality

## Notes

- The `WrapJSONValueInVariant` adapter method is a temporary bridge
- In future phases, this may be replaced with direct evaluator access to variant creation
- The reflection approach is efficient (single field access) and avoids import issues
- JSON indexing now follows the same pattern as array/string indexing (all in evaluator)

## Next Steps

With task 3.5.99b complete, the remaining tasks are:

- **Task 3.5.99c**: Object Default Property Access
- **Task 3.5.99d**: Interface Default Property Access
- **Task 3.5.99e**: Record Default Property Access
- **Task 3.5.99f**: Indexed Property Infrastructure
- **Task 3.5.99g**: Indexed Property Access via MemberAccessExpression

**Progress**: 2 of 7 subtasks complete (3.5.99a and 3.5.99b)
