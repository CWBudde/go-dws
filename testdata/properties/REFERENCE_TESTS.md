# DWScript Reference Property Tests

This document maps the original DWScript property tests from `reference/dwscript-original/Test/` to our implementation status.

## Task 8.60: Reference Test Analysis

### Currently Supported Tests (Can be ported now)

The tests in `testdata/properties/` cover the core property functionality we've implemented:

#### Implemented Features:
- ✅ Field-backed properties (`read FField write FField`)
- ✅ Method-backed properties (`read GetValue write SetValue`)
- ✅ Read-only properties (`read FField`)
- ✅ Write-only properties (`write FField`)
- ✅ Auto-properties (`property Name: String;`)
- ✅ Property inheritance
- ✅ Duplicate property detection
- ✅ Type validation for getters/setters
- ✅ Field/method existence validation

#### Our Test Coverage:
- `basic_property.dws` - Field and method-backed properties
- `property_inheritance.dws` - Property inheritance in class hierarchy
- `read_only_property.dws` - Read-only properties with computed values
- `auto_property.dws` - Auto-properties with backing fields
- `mixed_properties.dws` - Mixed property types and validation

### Tests Requiring Deferred Features

#### Expression-Based Properties (Task 8.56 - DEFERRED)

These tests use inline expressions in read/write specifiers:
```pascal
property Value: Integer read (FValue * 2);
property Value: Integer write (FValue := Value div 2);
```

**Reference tests in `PropertyExpressionsPass/`:**
- `simple_instance_expressions.pas` - Basic expression-based getters
- `class_property_expressions.pas` - Class properties with expressions
- `property_auto_field.pas` - Mix of auto and expression properties
- `simple_record_expressions.pas` - Record properties with expressions
- `read_write_other_property.pas` - Properties reading/writing other properties

**When implemented**, port these tests to verify expression evaluation works correctly.

#### Indexed Properties (Task 8.55 - DEFERRED)

These tests use array-like indexed properties:
```pascal
property Items[index: Integer]: String read GetItem write SetItem;
property default;  // default indexed property
```

**Reference tests:**
- `indexed_expressions.pas` - Indexed property access
- `indexed_write_expressions.pas` - Indexed property assignment
- `double_brackets.pas` - Multi-dimensional indexed properties
- `property_default1.pas` (FailureScripts) - Default property errors

**When implemented**, port these tests to verify indexing works correctly.

### Error Tests from FailureScripts

The `FailureScripts/property_*.pas` files test various error conditions:

#### Already Covered by Our Tests:
- ✅ Duplicate property names
- ✅ Missing read/write fields/methods
- ✅ Type mismatches in getters/setters
- ✅ Missing property type annotations

#### Requires Future Features:
- Default property validation (requires indexed properties)
- Property promotion rules (advanced inheritance)
- Property reintroduction (override semantics)
- Compound assignment to properties (`+=`, `-=`)
- Class properties (static properties)
- Helper properties (class helpers)

### Advanced Features Not Yet Planned

Some reference tests use features beyond Stage 8:

- **Generics + Properties**: `GenericsPass/class_property1.pas`
- **Helpers + Properties**: `HelpersPass/helper_property.pas`
- **RTTI + Properties**: `FunctionsRTTI/property_*.pas`
- **JSON Connector**: `JSONConnectorPass/property_name.pas`
- **External Properties**: `JSFilterScripts/external_property_*.dws`

These will be addressed in later stages (Stage 9-10).

## Compatibility Notes

### Differences from Original DWScript:

1. **No `inherited` keyword yet**: We manually initialize parent fields in constructors
2. **No TObject base class**: Using custom class hierarchies instead
3. **No `var` keyword for parameters**: By-ref parameters not fully implemented
4. **No operator overloading in properties**: Basic operators only

### Future Work (Post-Stage 8):

When implementing Tasks 8.55-8.56:

1. **Indexed Properties**:
   - Parse index parameter lists
   - Pass index values to getters/setters
   - Support default properties
   - Handle multi-dimensional indexing

2. **Expression-Based Properties**:
   - Parse expressions in read/write specifiers
   - Evaluate expressions in object context
   - Support accessing other fields/properties
   - Handle method calls in expressions

3. **Test Porting**:
   - Port ~15-20 tests from `PropertyExpressionsPass/`
   - Port ~10 indexed property tests
   - Port relevant error tests from `FailureScripts/`
   - Update our tests to match DWScript output format

## Current Test Status

**Our implementation**: 31/35 tasks complete (89%)
- ✅ Core property functionality fully tested
- ⏸️ Indexed properties deferred
- ⏸️ Expression-based getters deferred

**CLI Integration**: ✅ All tests passing
- 5 parsing tests
- 6 syntax tests
- 1 inheritance test

**Semantic Validation**: ✅ 20+ test cases passing
- Property declaration validation
- Getter/setter validation
- Type checking
- Inheritance rules

**Runtime Tests**: ✅ 6/6 tests passing
- Field-backed properties
- Method-backed properties
- Read-only properties
- Inheritance
- Auto-properties
- Error handling
