# Property Test Files

This directory contains test files for DWScript property functionality (Tasks 8.53-8.58).

## Test Files

### basic_property.dws
Tests basic property functionality:
- Field-backed properties (direct field access)
- Method-backed properties (getter/setter methods)
- Reading and writing property values
- Side effects in setter methods

**Expected Output:**
```
Name: TestCounter
Count set to 5
Count: 5
Count set to 6
Count after increment: 6
```

### property_inheritance.dws
Tests property inheritance through class hierarchy:
- Base class properties
- Derived class properties
- Inherited property access
- Multiple property types in inheritance chain

**Expected Output:**
```
Value: 42
Name: test
Extra: additional
```

### read_only_property.dws
Tests read-only properties:
- Properties with only read specifier
- Computed properties (calculated from other fields)
- Read-only enforcement at semantic analysis phase
- Multiple read-only properties

**Expected Output:**
```
Radius: 5.0
Diameter: 10.0
Area: 78.53975
New Radius: 10.0
New Diameter: 20.0
New Area: 314.159
```

### auto_property.dws
Tests auto-properties:
- Properties without explicit read/write specifiers
- Automatic backing field generation
- Comparison with standard properties

**Expected Output:**
```
Name: John Doe
Age: 30
Updated Name: Jane Smith
Updated Age: 25
```

### mixed_properties.dws
Tests mixed property types in one class:
- Field-backed properties
- Method-backed computed properties
- Properties with validation logic
- Combination of read-only and read-write properties

**Expected Output:**
```
Width: 5.0
Height: 3.0
Area: 15.0
Perimeter: 16.0
Testing validated properties:
Error: Height must be positive
Final Width: 10.0
Final Height: 7.0
Final Area: 70.0
```

## Running Tests

To run these test files (once the interpreter is complete):

```bash
# Run a specific test
./bin/dwscript run testdata/properties/basic_property.dws

# Run all property tests
for f in testdata/properties/*.dws; do
    echo "=== $f ==="
    ./bin/dwscript run "$f"
    echo
done
```

## Coverage

These test files cover:
- ✅ Task 8.53: Property read access (field and method)
- ✅ Task 8.54: Property write access (field and method)
- ⏸️ Task 8.55: Indexed property access (deferred)
- ⏸️ Task 8.56: Expression-based property getters (deferred)
- ✅ Property inheritance
- ✅ Read-only properties
- ✅ Auto-properties
- ✅ Property validation in setters
- ✅ Computed properties
