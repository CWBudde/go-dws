# IndexOf Built-in Function Design

**Date:** 2025-10-27
**Status:** Approved
**Related Tasks:** PLAN.md tasks 9.69-9.71

## Overview

Implementation of the `IndexOf` built-in function for array searching in DWScript. This function searches an array for a specific value and returns its position using 1-based indexing (consistent with Pascal/Delphi conventions).

## Requirements

### Functional Requirements

1. **Basic search:** `IndexOf(array, value)` returns the 1-based index of the first occurrence
2. **Search with start position:** `IndexOf(array, value, startIndex)` begins search from specified index
3. **Not found indicator:** Returns `0` when value not found (matching Delphi `Pos()` semantics)
4. **Type-safe comparison:** Uses existing `valuesEqual()` helper for all DWScript types

### Design Decisions

**Indexing Convention:**
- **Returns 1-based index** (first element is at position 1)
- **Overrides PLAN.md specification** which suggested 0-based indexing
- **Rationale:** Consistency with traditional Pascal/Delphi behavior and DWScript's `Pos()` function

**Error Handling:**
- Invalid `startIndex` (negative or >= array length) returns `0` (not found)
- Silent failure for boundary violations (no runtime errors)
- Runtime errors only for type mismatches and wrong argument counts

## Architecture

### Function Signatures

#### Main Dispatcher
```go
// File: internal/interp/interpreter.go
func (i *Interpreter) builtinIndexOf(args []Value) Value
```

Responsibilities:
- Validate argument count (2 or 3)
- Type-check first argument (must be array)
- Type-check third argument if present (must be integer)
- Extract parameters and delegate to core implementation

#### Core Implementation
```go
// File: internal/interp/array_functions.go
func (i *Interpreter) builtinArrayIndexOf(arr *ArrayValue, value Value, startIndex int) Value
```

Responsibilities:
- Validate startIndex bounds
- Search array elements from startIndex onwards
- Use `valuesEqual()` for type-safe comparison
- Return 1-based index or 0 for not found

### Registration

Add to existing switch statement in `interpreter.go`:
```go
case "IndexOf":
    return i.builtinIndexOf(args)
```

## Implementation Details

### Algorithm

1. **Boundary validation:**
   - If `startIndex < 0`: return 0 (invalid start position)
   - If `startIndex >= len(arr.Elements)`: return 0 (beyond array bounds)

2. **Linear search:**
   ```go
   for idx := startIndex; idx < len(arr.Elements); idx++ {
       if i.valuesEqual(arr.Elements[idx], value) {
           return &IntegerValue{Value: int64(idx + 1)}  // Convert to 1-based
       }
   }
   ```

3. **Not found:**
   ```go
   return &IntegerValue{Value: 0}
   ```

### Type Handling

- **Search value:** Accepts any DWScript type
- **Comparison:** Delegates to `valuesEqual()` which handles:
  - Integer, Float, String, Boolean primitives
  - Enum values
  - Record and Set types (structural equality)
  - Type mismatches return false (no error)

### Error Conditions

| Condition | Behavior |
|-----------|----------|
| Wrong argument count | Runtime error with location info |
| First arg not array | Runtime error: "expects array as first argument" |
| Third arg not integer | Runtime error: "expects integer as third argument" |
| Negative startIndex | Return 0 (silent failure) |
| startIndex >= length | Return 0 (silent failure) |
| Value not found | Return 0 (normal behavior) |
| Empty array | Return 0 (normal behavior) |

## Testing Strategy

### Test Coverage

**File:** `internal/interp/array_test.go`

#### 1. Basic Search - Found
```go
IndexOf([1,2,3,2], 2) → 1  // First occurrence at position 1
```

#### 2. Basic Search - Not Found
```go
IndexOf([1,2,3], 5) → 0
```

#### 3. Search with Start Index
```go
IndexOf([1,2,3,2], 2, 2) → 4  // Skip first two elements, find at position 4
```

#### 4. String Arrays
```go
IndexOf(['a','b','c'], 'b') → 2
IndexOf(['hello','world'], 'foo') → 0
```

#### 5. Empty Array
```go
IndexOf([], 42) → 0
```

#### 6. Edge Cases - Boundary Conditions
```go
IndexOf([1,2,3], 2, 0) → 2     // Start at 0, searches from beginning
IndexOf([1,2,3], 2, -1) → 0    // Negative index returns not found
IndexOf([1,2,3], 2, 10) → 0    // Beyond bounds returns not found
```

#### 7. Error Cases - Type Validation
```go
IndexOf([1,2,3])         → Error: expects 2 or 3 arguments
IndexOf(42, 1)           → Error: expects array as first argument
IndexOf([1,2], 1, 'bad') → Error: expects integer as third argument
```

### Test Structure

Follow existing patterns in `array_test.go`:
- Table-driven tests for systematic coverage
- Helper functions to create test arrays
- Clear test names describing behavior
- Both success and error cases

## Files to Modify

1. **internal/interp/array_functions.go**
   - Add `builtinArrayIndexOf()` function
   - ~15-25 lines of implementation

2. **internal/interp/interpreter.go**
   - Add `builtinIndexOf()` dispatcher
   - Add case "IndexOf" to switch statement
   - ~20-30 lines total

3. **internal/interp/array_test.go**
   - Add `TestArrayIndexOf` function
   - 7+ test cases covering all scenarios
   - ~100-150 lines of tests

## Compatibility Notes

### DWScript Compatibility

This implementation differs from standard DWScript in one key aspect:

**PLAN.md Deviation:**
- PLAN.md tasks 9.69-9.71 specify 0-based indexing (return 0 for first element, -1 for not found)
- This implementation uses 1-based indexing (return 1 for first element, 0 for not found)
- Rationale: Better alignment with Pascal/Delphi conventions and existing DWScript built-ins like `Pos()`

### Future Considerations

1. **Performance:** Linear search is O(n). For large arrays, consider:
   - Sorted array variant with binary search
   - Hash-based lookup for frequent searches
   - Not needed for Stage 9; defer to Stage 10 (performance optimization)

2. **LastIndexOf:** Natural companion function for reverse search
   - Not in current PLAN.md
   - Easy to add later with same architecture

3. **Multiple occurrences:** Return array of all indices
   - Not requested in requirements
   - YAGNI: don't implement without user need

## Approval

Design approved through brainstorming skill phases 1-3:
- ✅ Requirements clarified (1-based indexing, 0 for not found)
- ✅ Approach selected (unified function with arg checking)
- ✅ Architecture validated (follows existing patterns)
- ✅ Error handling approved (silent failure for bounds, errors for types)
- ✅ Test coverage approved (7 test scenarios)

Ready for implementation.
