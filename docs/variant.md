# Variant Type in DWScript

**Status**: Implemented (Tasks 9.220-9.239)
**Compatibility**: Full DWScript/Delphi Variant semantics

## Overview

The `Variant` type in DWScript is a dynamic type that can hold values of any type at runtime with automatic type tracking. It provides runtime flexibility similar to dynamically-typed languages while maintaining compatibility with Pascal's static type system.

### Key Features

- **Dynamic Typing**: A Variant variable can hold Integer, Float, String, Boolean, or other values
- **Runtime Type Tracking**: Each Variant maintains its actual type via VarType codes
- **Automatic Boxing/Unboxing**: Values are automatically wrapped when assigned to Variants
- **Type Conversions**: Built-in functions for safe type conversion
- **Heterogeneous Arrays**: Enables `array of const` for variadic-style parameters
- **DWScript/Delphi Compatible**: Follows original DWScript and Delphi Variant behavior

### Use Cases

1. **Variadic Functions**: Format() and similar functions accepting mixed-type arguments
2. **Dynamic Data**: JSON parsing, database fields, COM/OLE automation
3. **Scripting Flexibility**: When types aren't known at compile-time
4. **Rapid Prototyping**: Defer type decisions during development

## Declaration and Initialization

### Basic Declaration

```pascal
var v: Variant;              // Uninitialized (null/empty)
var v1: Variant := 42;       // Integer variant
var v2: Variant := 3.14;     // Float variant
var v3: Variant := 'hello';  // String variant
var v4: Variant := True;     // Boolean variant
```

### Uninitialized Variants

Uninitialized Variants have the special state "empty" (VarType code 0):

```pascal
var v: Variant;
PrintLn(VarIsNull(v));     // True
PrintLn(VarIsEmpty(v));    // True (same as VarIsNull in DWScript)
PrintLn(VarType(v));       // 0 (varEmpty)
```

## Boxing and Unboxing

### Automatic Boxing

When assigning a typed value to a Variant, the value is automatically "boxed" (wrapped):

```pascal
var i: Integer := 42;
var v: Variant;
v := i;  // Automatic boxing: i is wrapped in VariantValue
```

**Implementation**: Values are wrapped in a `VariantValue` struct containing:
- `Value`: The actual runtime value (IntegerValue, FloatValue, etc.)
- `ActualType`: The semantic type information

### Unboxing via Conversion Functions

To extract a typed value from a Variant, use conversion functions:

```pascal
var v: Variant := 42;
var i: Integer;

// Explicit conversion required
i := VarToInt(v);
```

**Note**: Direct assignment from Variant to typed variables (`i := v`) is not supported. Always use conversion functions (VarToInt, VarToFloat, VarToStr).

## Type Conversion Rules

### VarToInt(v: Variant): Integer

Converts a Variant to Integer with these rules:

- **Integer → Integer**: Returns value as-is
- **Float → Integer**: Truncates (3.9 → 3, not rounding)
- **String → Integer**: Parses string ("123" → 123), error if invalid
- **Boolean → Integer**: True → 1, False → 0
- **Empty → Integer**: Returns 0

```pascal
var v1: Variant := 3.9;
var v2: Variant := '123';
var v3: Variant := True;

PrintLn(VarToInt(v1));   // 3 (truncated)
PrintLn(VarToInt(v2));   // 123 (parsed)
PrintLn(VarToInt(v3));   // 1
```

### VarToFloat(v: Variant): Float

Converts a Variant to Float:

- **Float → Float**: Returns value as-is
- **Integer → Float**: Converts (42 → 42.0)
- **String → Float**: Parses string ("3.14" → 3.14), error if invalid
- **Boolean → Float**: True → 1.0, False → 0.0
- **Empty → Float**: Returns 0.0

```pascal
var v1: Variant := 42;
var v2: Variant := '2.71828';

PrintLn(VarToFloat(v1));   // 42.0
PrintLn(VarToFloat(v2));   // 2.71828
```

### VarToStr(v: Variant): String

Converts a Variant to String:

- **String → String**: Returns as-is
- **Integer → String**: "42"
- **Float → String**: "3.14"
- **Boolean → String**: "true" or "false" (lowercase)
- **Empty → String**: "" (empty string)

```pascal
var v1: Variant := 42;
var v2: Variant := 3.14;
var v3: Variant := True;

PrintLn(VarToStr(v1));   // "42"
PrintLn(VarToStr(v2));   // "3.14"
PrintLn(VarToStr(v3));   // "true"
```

### VarAsType(v: Variant, typeCode: Integer): Variant

Explicit conversion using VarType codes:

```pascal
var v: Variant := '123';
var asInt: Variant := VarAsType(v, 3);    // 3 = varInteger
var asFloat: Variant := VarAsType(v, 5);  // 5 = varDouble
var asStr: Variant := VarAsType(v, 256);  // 256 = varString

PrintLn(VarType(asInt));    // 3
PrintLn(VarType(asFloat));  // 5
PrintLn(VarType(asStr));    // 256
```

**VarType Codes**:
- 0 = varEmpty (uninitialized)
- 3 = varInteger
- 5 = varDouble (Float)
- 11 = varBoolean
- 256 = varString

## Array of Const Pattern

The Variant type enables the `array of const` pattern for variadic-style functions with heterogeneous arguments.

### Heterogeneous Arrays

```pascal
// Array can contain mixed types
var arr: array of Variant := ['hello', 42, 3.14, True];

PrintLn(VarType(arr[0]));  // 256 (String)
PrintLn(VarType(arr[1]));  // 3 (Integer)
PrintLn(VarType(arr[2]));  // 5 (Float)
PrintLn(VarType(arr[3]));  // 11 (Boolean)
```

### Format() Function

The `Format()` function uses `array of Variant` to accept mixed-type arguments:

```pascal
// Format signature: Format(format: String, args: array of Variant): String
PrintLn(Format('Name: %s, Age: %d, Score: %.2f', ['Alice', 30, 95.5]));
// Output: Name: Alice, Age: 30, Score: 95.50
```

### Implementing Variadic Functions

To create your own variadic function:

```pascal
procedure LogValues(values: array of Variant);
var i: Integer;
begin
    for i := 0 to Length(values) - 1 do begin
        PrintLn(Format('[%d] type=%d, value=%s',
                [i, VarType(values[i]), VarToStr(values[i])]));
    end;
end;

// Call with mixed types
LogValues([42, 'hello', 3.14, True]);
```

## Built-in Variant Functions

### Type Introspection

#### VarType(v: Variant): Integer

Returns the type code identifying the actual type:

```pascal
var v1: Variant := 42;
var v2: Variant := 'hello';

PrintLn(VarType(v1));  // 3 (varInteger)
PrintLn(VarType(v2));  // 256 (varString)
```

**Type Codes** (Delphi-compatible):
- `0` (varEmpty) - Uninitialized/null
- `3` (varInteger) - Integer value
- `5` (varDouble) - Float value
- `11` (varBoolean) - Boolean value
- `256` (varString) - String value
- `0x2000` (varArray) - Array value

#### VarIsNull(v: Variant): Boolean

Checks if Variant is uninitialized:

```pascal
var v1: Variant;
var v2: Variant := 42;

PrintLn(VarIsNull(v1));  // True
PrintLn(VarIsNull(v2));  // False
```

#### VarIsEmpty(v: Variant): Boolean

Alias for VarIsNull (same semantics in DWScript):

```pascal
var v: Variant;
PrintLn(VarIsEmpty(v));  // True
```

#### VarIsNumeric(v: Variant): Boolean

Checks if Variant contains a numeric type (Integer or Float):

```pascal
var v1: Variant := 42;
var v2: Variant := 3.14;
var v3: Variant := 'hello';

PrintLn(VarIsNumeric(v1));  // True
PrintLn(VarIsNumeric(v2));  // True
PrintLn(VarIsNumeric(v3));  // False
```

### Conversion Functions

See "Type Conversion Rules" section above for:
- `VarToInt(v: Variant): Integer`
- `VarToFloat(v: Variant): Float`
- `VarToStr(v: Variant): String`
- `VarAsType(v: Variant, typeCode: Integer): Variant`

## Comparison with Delphi/DWScript

### Compatibility

Our Variant implementation follows Delphi/DWScript semantics:

| Feature | Delphi | DWScript Original | go-dws | Notes |
|---------|---------|-------------------|---------|-------|
| VarType codes | ✅ | ✅ | ✅ | Identical codes |
| VarIsNull/VarIsEmpty | Different | Same | Same | DWScript doesn't distinguish |
| Automatic boxing | ✅ | ✅ | ✅ | On assignment |
| Variant operators | ✅ | ✅ | Runtime only | Semantic support pending |
| Array of const | ✅ | ✅ | ✅ | via array of Variant |
| COM/OLE support | ✅ | Partial | N/A | Not applicable to Go |

### Key Differences from Delphi

1. **Empty vs. Null**: In Delphi, `varEmpty` and `varNull` are distinct. In DWScript (and go-dws), they're equivalent.

2. **Uninitialized State**: Delphi Variants start as `varEmpty`. DWScript Variants start as `varEmpty` (nil value).

3. **Reference Types**: Delphi Variants can hold interfaces and objects. go-dws support is limited.

4. **Semantic Analysis**: Currently, Variant operators work at runtime but require explicit conversion in some semantic contexts.

### Migration from ConstType

**Historical Note**: Tasks 9.235-9.237 migrated from a temporary `ConstType` workaround to proper `VariantType`:

```pascal
// Old (Task 9.156): Temporary workaround
var ARRAY_OF_CONST = NewDynamicArrayType(CONST)

// New (Task 9.235): Proper implementation
var ARRAY_OF_CONST = NewDynamicArrayType(VARIANT)
```

This migration ensures:
- Proper type safety with runtime checking
- Consistent boxing/unboxing semantics
- Better error messages
- Future-proof for additional Variant features

## Performance Considerations

### Runtime Overhead

Variants introduce performance costs compared to statically-typed variables:

1. **Boxing/Unboxing**: Wrapping values in VariantValue containers
2. **Type Checking**: Runtime type dispatch for operations
3. **Memory**: Additional metadata (ActualType field)
4. **Indirection**: Extra pointer dereference to access actual value

**Recommendation**: Use Variants only when dynamic typing is necessary. Prefer statically-typed variables for performance-critical code.

### When to Use Variants

✅ **Good Use Cases**:
- Format() and similar variadic functions
- JSON/dynamic data parsing
- Database field values of unknown types
- Prototyping and scripting
- Interfacing with dynamic systems

❌ **Avoid For**:
- Performance-critical loops
- Mathematical computations
- When types are known at compile-time
- Simple data structures

### Optimization Tips

1. **Extract Early**: Convert Variants to typed variables once, reuse:
   ```pascal
   var v: Variant := getSomeValue();
   var i: Integer := VarToInt(v);  // Convert once
   // Use i repeatedly in calculations
   ```

2. **Minimize Conversions**: Avoid repeated VarToX calls in loops:
   ```pascal
   // Bad
   for j := 1 to 1000 do
       x := VarToInt(v) + j;  // Converts 1000 times

   // Good
   var i: Integer := VarToInt(v);  // Convert once
   for j := 1 to 1000 do
       x := i + j;
   ```

3. **Type Checking**: Use VarIsNumeric() before conversion to avoid exceptions:
   ```pascal
   if VarIsNumeric(v) then
       i := VarToInt(v)
   else
       PrintLn('Not a number');
   ```

## Implementation Details

### Internal Structure

```go
// Runtime Value Representation
type VariantValue struct {
    Value      Value      // Actual wrapped value (IntegerValue, FloatValue, etc.)
    ActualType types.Type // Semantic type information
}

// Semantic Type Representation
type VariantType struct{}

func (t *VariantType) String() string { return "Variant" }
func (t *VariantType) TypeKind() string { return "VARIANT" }
```

### Type System Integration

```
Semantic Layer:        types.VARIANT (VariantType)
                              ↓
Runtime Layer:         VariantValue { Value, ActualType }
                              ↓
Actual Value:          IntegerValue | FloatValue | StringValue | ...
```

### Boxing Implementation

```go
// Automatic boxing when assigning to Variant
func boxVariant(value Value) *VariantValue {
    if value == nil {
        return &VariantValue{Value: nil, ActualType: nil}
    }

    // Map runtime type to semantic type
    var actualType types.Type
    switch value.Type() {
    case "INTEGER": actualType = types.INTEGER
    case "FLOAT":   actualType = types.FLOAT
    case "STRING":  actualType = types.STRING
    case "BOOLEAN": actualType = types.BOOLEAN
    }

    return &VariantValue{Value: value, ActualType: actualType}
}
```

### Unboxing Implementation

```go
// Unwrap Variant to get actual value
func unwrapVariant(value Value) Value {
    if variant, ok := value.(*VariantValue); ok {
        if variant.Value == nil {
            return &NilValue{}
        }
        return variant.Value
    }
    return value
}
```

## Future Work

### Planned Enhancements

1. **Semantic Operator Support**: Full semantic analysis for Variant operators (+, -, *, /, comparisons)
2. **Additional VarType Codes**: Support for varArray element type tracking
3. **Advanced Conversions**: VarCast with custom conversion rules
4. **Null Propagation**: Optional null-safe operations
5. **Performance Optimizations**: Inline boxing for common cases

### Current Limitations

- **Semantic Operators**: Variant arithmetic/comparisons work at runtime but may require `--type-check=false` in some contexts
- **Object/Interface Variants**: Limited support for class instances and interfaces
- **Variant Arrays**: Element type tracking is simplified (varArray without element type detail)

## Examples

### Complete Example

```pascal
// Variant basics demonstration
var v: Variant;

begin
    // Type introspection
    PrintLn(VarIsNull(v));        // True
    PrintLn(VarType(v));          // 0 (varEmpty)

    // Dynamic typing
    v := 42;
    PrintLn(VarType(v));          // 3 (varInteger)
    PrintLn(VarIsNumeric(v));     // True

    v := 'hello';
    PrintLn(VarType(v));          // 256 (varString)
    PrintLn(VarIsNumeric(v));     // False

    // Conversions
    v := '123';
    var i: Integer := VarToInt(v);
    PrintLn(i);                   // 123

    // Heterogeneous arrays
    var arr: array of Variant := [42, 'text', 3.14, True];
    PrintLn(Format('Values: %d, %s, %.2f, %s', arr));
    // Output: Values: 42, text, 3.14, true
end.
```

### Error Handling

```pascal
var v: Variant := 'not a number';

// Safe conversion with error handling
if VarIsNumeric(v) then begin
    var i: Integer := VarToInt(v);
    PrintLn('Converted: ' + IntToStr(i));
end else begin
    PrintLn('Cannot convert to integer');
end;
```

## Test Suite

Comprehensive tests are available in `testdata/variant/`:

- `basic.dws` - Declarations, assignments, type checking
- `conversions.dws` - Type conversions and VarToX functions
- `array_of_const.dws` - Heterogeneous arrays and Format()
- `arithmetic.dws` - Operations (requires semantic support)

Run tests:
```bash
./bin/dwscript run testdata/variant/basic.dws
./bin/dwscript run testdata/variant/conversions.dws
./bin/dwscript run testdata/variant/array_of_const.dws
```

## References

- **DWScript Documentation**: https://www.delphitools.info/dwscript/
- **Delphi Variant Reference**: https://docwiki.embarcadero.com/Libraries/en/System.Variants
- **Original Implementation**: `reference/dwscript-original/Source/dwsVariant.pas`
- **Implementation Tasks**: PLAN.md Tasks 9.220-9.239

## Related Documentation

- `docs/enums.md` - Enumerated types and Ord() function
- `CLAUDE.md` - Project overview and architecture
- `PLAN.md` - Full task breakdown and progress

---

**Task Status**: ✅ Complete (Tasks 9.220-9.239)
**Last Updated**: 2025-01-XX
**Compatibility**: DWScript/Delphi Variant semantics
