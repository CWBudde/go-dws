# Variant Type Test Suite

Test files for the Variant Type System (Task 9.238).

## Test Files

### basic.dws ✅
Tests basic Variant declarations, assignments, and type introspection.
- Uninitialized variants
- Initialized variants with different types
- Dynamic reassignment
- Type checking with VarType, VarIsNull, VarIsNumeric
- Variant to typed variable conversion

### conversions.dws ✅
Tests Variant conversion functions.
- VarToStr, VarToInt, VarToFloat
- VarAsType with type codes
- Type checking during conversion

### array_of_const.dws ✅
Tests heterogeneous arrays (array of Variant).
- Mixed-type array literals
- Format() function with heterogeneous arrays
- Array iteration and modification
- SetLength with Variant arrays

### arithmetic.dws (Future)
Tests Variant arithmetic and comparison operations.
- **Status**: Requires semantic analyzer support for Variant operators
- **Note**: Runtime support is implemented (Tasks 9.229-9.230), but semantic analysis support is pending
- This file demonstrates intended behavior once semantic support is added

## Expected Output Files

- `basic.out` - Generated from basic.dws
- `conversions.out` - Generated from conversions.dws
- `array_of_const.out` - Generated from array_of_const.dws
- `arithmetic.out` - To be generated once semantic support is added

## Running Tests

```bash
# Run individual test
./bin/dwscript run testdata/variant/basic.dws

# Compare with expected output
./bin/dwscript run testdata/variant/basic.dws | diff - testdata/variant/basic.out
```

## Implementation Status

Tasks completed:
- ✅ 9.220-9.222: Type Definition
- ✅ 9.223-9.226: Semantic Analysis
- ✅ 9.227-9.231: Runtime Support
- ✅ 9.232-9.234: Built-in Functions
- ✅ 9.235-9.237: Migration from ConstType
- ✅ 9.238-9.239: Testing & Documentation

## Future Work

- Semantic analyzer support for Variant operators (+, -, *, /, div, mod, =, <>, <, >, <=, >=, and, or, not)
- This will enable arithmetic.dws to run with full type checking
