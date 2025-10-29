# Rosetta Code Compatibility Status

This document tracks the compatibility status of DWScript Rosetta Code examples with go-dws.

## Array Instantiation with `new` Keyword (Tasks 9.164-9.168)

**Status**: ✅ **Implemented**

The `new TypeName[size1, size2, ...]` syntax for dynamic array instantiation is now fully supported.

### Working Examples

#### ✅ Levenshtein Distance (Modified)
- **File**: `testdata/new_array/levenshtein_working.dws`
- **Status**: Working with modifications
- **Original**: `examples/rosetta/Levenshtein_distance.dws`
- **Required Changes**:
  - Changed `[i, j]` syntax to `[i][j]` for multi-dimensional indexing
  - Replaced `MinInt()` with custom `Min3()` helper function
- **Test Results**: All test cases pass (kitten→sitting: 3, rosettacode→raisethysword: 8, saturday→sunday: 3)

### Blocked Examples

#### ⏸️ Gnome Sort
- **File**: `examples/rosetta/Sorting_algorithms_Gnome_sort.dws`
- **Status**: Blocked - requires array properties/methods
- **Blockers**:
  - `.Length` property (not implemented)
  - `.High` property (not implemented)
  - `.Swap()` method (not implemented)
- **Notes**: The `new Integer[16]` array instantiation parses correctly, but the algorithm requires array helper methods that are part of future stages.

#### ⏸️ Yin and Yang
- **File**: `examples/rosetta/Yin_and_yang.dws`
- **Status**: Partially Blocked - parser limitations
- **Working**: `method` keyword is now supported (Task 9.169)
- **Remaining Blockers**:
  - Inline array types in class fields: `Pix : array of array of Integer;` (parser error)
  - `.High` property on arrays (not implemented)
  - Multi-index syntax `[i, j]` (not supported by parser)
  - `case` statement with multiple values (not implemented)
- **Notes**: The `new Integer[aScale*12+1, aScale*12+1]` expression is supported, and the `method` keyword now parses correctly. The class definition still has other parser issues that need to be addressed in future stages.

## Test Coverage

### Created Test Files (Task 9.167)

All test files are located in `testdata/new_array/`:

1. **`new_array_basic.dws`** ✅
   - Simple 1D array creation and access
   - Integer and String arrays
   - Element initialization and sum calculation

2. **`new_array_multidim.dws`** ✅
   - 2D matrix (3×4) creation and manipulation
   - 3D cube creation with spot checks
   - Nested loops and row sums

3. **`new_array_expressions.dws`** ✅
   - Dynamic sizes from arithmetic expressions (`2 * 3`, `10 - 3`)
   - Function results as sizes (`GetArraySize()`)
   - Complex expressions (`base * multiplier + 2`)
   - String length-based arrays (`new String[Length(text)]`)

4. **`new_array_types.dws`** ✅
   - Integer, Float, String, and Boolean arrays
   - 2D arrays of mixed types
   - Zero value verification for all types

5. **`levenshtein_working.dws`** ✅
   - Real-world algorithm using 2D dynamic arrays
   - Expression-based dimensions: `new Integer[Length(s)+1, Length(t)+1]`
   - Demonstrates practical use of multi-dimensional arrays

### Integration Tests (Task 9.168)

CLI integration tests in `cmd/dwscript/new_array_test.go`:
- All 5 test files pass with expected output
- Exit code validation
- Output comparison with expected results

## Syntax Support

### ✅ Supported

- `new TypeName[size]` - 1D arrays
- `new TypeName[size1, size2]` - 2D arrays
- `new TypeName[size1, size2, size3, ...]` - N-dimensional arrays
- Expression-based sizes: `new Integer[Length(s) + 1]`
- Arithmetic in sizes: `new Integer[rows * cols]`
- Function calls in sizes: `new String[GetSize()]`
- Nested indexing: `array[i][j][k]`
- All basic types: Integer, Float, String, Boolean

### ❌ Not Supported (Parser Limitations)

- Multi-index syntax: `array[i, j]` (must use `array[i][j]`)
- Inline array types in class fields: `field : array of array of T;`

### ⏸️ Not Yet Implemented (Future Stages)

- Array properties: `.Length`, `.High`, `.Low`
- Array methods: `.Swap()`, `.Copy()`, `.Reverse()`, `.Sort()`
- Dynamic array resizing: `SetLength()`

## Implementation Details

### Zero Value Initialization

All array elements are initialized to appropriate zero values:
- **Integer**: `0`
- **Float**: `0.0`
- **String**: `""` (empty string)
- **Boolean**: `False`

### Multi-dimensional Arrays

Multi-dimensional arrays are implemented as nested arrays:
- `new Integer[3, 4]` creates an array of 3 arrays, each containing 4 integers
- Type structure: `array of (array of Integer)`
- Access via nested indexing: `matrix[row][col]`

### Performance Characteristics

- Array creation: O(n) for 1D, O(n×m) for 2D, etc.
- Element access: O(1) per dimension
- Zero value initialization: Performed at creation time

## Future Work

### Stage 10+: Array Helper Methods
- Implement `.Length`, `.High`, `.Low` properties
- Add `.Swap()`, `.Copy()` methods
- Support inline array type syntax in class fields

### Parser Enhancements
- Add support for multi-index syntax `[i, j]`
- Fix inline array type parsing in class fields

## Summary

**Tasks 9.164-9.168 Status**: ✅ **Complete**

- ✅ Runtime evaluation of `new` expressions
- ✅ Multi-dimensional array support
- ✅ Expression-based dimension sizes
- ✅ All basic types supported
- ✅ Comprehensive test coverage (5 test files, 16+ test cases)
- ✅ CLI integration tests
- ✅ Real-world example (Levenshtein distance) working with minor modifications

The `new` keyword array instantiation is fully functional and ready for use. Some Rosetta Code examples require additional features (array properties/methods) that are planned for future stages.
