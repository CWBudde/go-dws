# Operator Overloading Test Suite

This directory contains comprehensive tests for DWScript's operator overloading feature, ported from the original DWScript test suite.

## Directory Structure

```
testdata/operators/
├── pass/              # Passing tests - valid operator overloading scenarios
├── fail/              # Failure tests - invalid operator declarations and usage
└── README.md          # This file
```

## Test Categories

### Passing Tests (`pass/`)

These tests verify correct operator overloading behavior:

#### Binary Operators

- **operator_overloading1.dws** - String + Integer operator
  - Tests custom operator for concatenating strings with integers
  - Expected: `'abc'+123` → `'abc[123]'`
  - Reference: `reference/dwscript-original/Test/OperatorOverloadPass/operator_overloading1.pas`

- **operator_overloading2.dws** - Symbolic operators (<<, >>)
  - Tests stream-like chaining with custom operators
  - Demonstrates multiple operator overloads with classes and enums
  - Reference: `reference/dwscript-original/Test/OperatorOverloadPass/operator_overloading2.pas`

#### IN Operator

- **operator_in_overloading.dws** - IN operator overload
  - Tests `in` operator for digit checking in Integer and Float
  - Expected: `1 in 123` → `True`, `4 in 123` → `False`
  - Reference: `reference/dwscript-original/Test/OperatorOverloadPass/operator_in_overloading.pas`

#### Implicit Conversions

- **implicit_record1.dws** - Integer → Record implicit conversion
  - Tests automatic conversion from Integer to record type
  - Expected: `F1 := 10` creates TFoo with X=10, Y=11
  - Reference: `reference/dwscript-original/Test/OperatorOverloadPass/implicit_record1.pas`

- **implicit_record2.dws** - Record → Integer implicit conversion
  - Tests reverse conversion (record to Integer)
  - Expected: TFoo fields summed when assigned to Integer
  - Reference: `reference/dwscript-original/Test/OperatorOverloadPass/implicit_record2.pas`

- **operator_implicit.dws** - Basic type implicit conversions
  - Tests Integer → String and Float → String conversions
  - Reference: `reference/dwscript-original/Test/OperatorOverloadPass/operator_implicit.pas`

- **operator_implicit2.dws** - Record-to-record implicit conversion
  - Tests automatic conversion between two record types
  - Reference: `reference/dwscript-original/Test/OperatorOverloadPass/operator_implicit2.pas`

#### C-Style Operators

- **c_style.dws** - C-style operators (==, !=)
  - Tests C-style operators alongside Pascal-style (=, <>)
  - Both operator styles can coexist with different implementations
  - Reference: `reference/dwscript-original/Test/OperatorOverloadPass/c_style.pas`

### Failure Tests (`fail/`)

These tests verify proper error detection for invalid operator declarations:

- **operator_overload1.dws** - Multiple declaration errors
  - Wrong parameter counts
  - Duplicate overloads
  - Type mismatches
  - Invalid operator names

- **operator_overload2.dws** - Incomplete declaration (missing type)

- **operator_overload3.dws** - Incomplete declaration (missing second parameter)

- **operator_overload4.dws** - Missing return type

- **operator_overload5.dws** - Invalid uses clauses
  - var/const parameters forbidden
  - Invalid function references

- **operator_overload6.dws** - Duplicate declarations and invalid contexts
  - Operators in function bodies (not allowed)
  - Duplicate operator overloads

## Test File Format

Each test consists of two files:

1. **`.dws` file** - DWScript source code
   - Contains a header comment referencing the original test
   - Implements the test scenario

2. **`.txt` file** - Expected output
   - For passing tests: Expected program output
   - For failure tests: Expected error messages with line/column numbers

## Running the Tests

### Manual Testing

Run individual tests using the CLI:

```bash
# Run a passing test
./bin/dwscript run testdata/operators/pass/operator_overloading1.dws

# Parse a failure test (should show errors)
./bin/dwscript parse testdata/operators/fail/operator_overload1.dws
```

### Automated Testing

The test suite can be integrated into Go tests:

```go
func TestOperatorOverloadingPass(t *testing.T) {
    files, _ := filepath.Glob("testdata/operators/pass/*.dws")
    for _, file := range files {
        // Run and compare output with .txt file
    }
}

func TestOperatorOverloadingFail(t *testing.T) {
    files, _ := filepath.Glob("testdata/operators/fail/*.dws")
    for _, file := range files {
        // Parse and verify expected errors
    }
}
```

## Operator Overloading Syntax

### Binary Operators

```dws
function Add(a: TypeA; b: TypeB): TypeC;
begin
  // implementation
end;

operator + (TypeA, TypeB): TypeC uses Add;
```

### Unary Operators

```dws
function Negate(a: TypeA): TypeB;
begin
  // implementation
end;

operator - (TypeA): TypeB uses Negate;
```

### Implicit Conversions

```dws
function Convert(a: SourceType): TargetType;
begin
  // conversion logic
end;

operator implicit (SourceType): TargetType uses Convert;
```

## Implementation Status

- ✅ Binary operator overloading (+, -, *, /, etc.)
- ✅ Unary operator overloading (-, not)
- ✅ IN operator overloading
- ✅ Implicit conversion operators
- ⚠️  Symbolic operators (<<, >>) - requires parser support
- ⚠️  C-style operators (==, !=) - requires parser support

## Reference

Original test files from DWScript:
- `reference/dwscript-original/Test/OperatorOverloadPass/`
- `reference/dwscript-original/Test/OperatorOverloadFail/`

For more information about DWScript operator overloading, see:
- [DWScript Documentation](https://www.delphitools.info/dwscript/)
- Project documentation in `docs/` directory
