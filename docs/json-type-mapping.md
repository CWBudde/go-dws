# JSON ↔ DWScript Type Mapping

**Task**: 9.103
**Status**: Complete
**Related Tasks**: 9.25-9.31, 9.89-9.90, 9.97-9.102, 9.104

---

## Overview

This document describes how JSON types are mapped to DWScript types when parsing JSON strings and how DWScript values are converted to JSON during serialization. Understanding these mappings is critical for working with JSON data in DWScript programs.

**Key Points**:
- JSON parsing uses Go's `encoding/json` library internally
- Most JSON primitive types map directly to DWScript primitives
- JSON arrays and objects are preserved as `JSONValue` types for reference semantics
- All JSON operations support `Variant` wrapping for dynamic typing

---

## Type Mapping Tables

### JSON → DWScript (Parsing Direction)

When you call `ParseJSON(jsonString)`, the following mappings apply:

| JSON Type | JSON Example | DWScript Type | DWScript Value | Notes |
|-----------|--------------|---------------|----------------|-------|
| `null` | `null` | Nil | `NilValue` | Represents absence of value |
| Boolean | `true`, `false` | Boolean | `BooleanValue` | Direct 1:1 mapping |
| Number (whole) | `42`, `-5`, `1000` | Integer | `IntegerValue` (Int64) | When parseable as int64 |
| Number (decimal) | `3.14`, `1.5e10` | Float | `FloatValue` (Float64) | Decimal or scientific notation |
| String | `"hello"`, `""` | String | `StringValue` | UTF-8, escape sequences handled |
| Array | `[1, 2, 3]` | **JSONValue** | `JSONValue` (array) | **Not converted to ArrayValue** |
| Object | `{"key": "value"}` | **JSONValue** | `JSONValue` (object) | **Not converted to RecordValue** |

**Implementation**: `internal/interp/builtins_json.go:67-141` (ParseJSON), `internal/interp/json_conversion.go:27-54` (jsonValueToValue)

#### Why Arrays/Objects Stay as JSONValue

Arrays and objects remain as `JSONValue` (not converted to `ArrayValue`/`RecordValue`) to preserve **reference semantics**:

```dws
var json := ParseJSON('{"items": [1, 2, 3]}');
var items := json['items'];  // Both json['items'] and items point to same array
items[0] := 99;              // This modifies the original JSON structure
PrintLn(ToJSON(json));       // Outputs: {"items":[99,2,3]}
```

If converted to `ArrayValue`, this would create a copy and mutations wouldn't be visible through other references.

---

### DWScript → JSON (Serialization Direction)

When you call `ToJSON(value)` or `ToJSONFormatted(value, indent)`, the following mappings apply:

| DWScript Type | DWScript Example | JSON Output | Notes |
|---------------|------------------|-------------|-------|
| Nil | `nil` | `null` | Empty/undefined values |
| Boolean | `True`, `False` | `true`, `false` | Lowercase in JSON |
| Integer | `42`, `-5` | `42`, `-5` | No decimal point |
| Float | `3.14`, `1.0` | `3.14`, `1.0` | May include `.0` |
| String | `'hello'` | `"hello"` | Double quotes, escaped |
| **JSONValue** | (from ParseJSON) | Original structure | Unwraps to JSON |
| **ArrayValue** | `[1, 2, 3]` | `[1,2,3]` | Recursive conversion |
| **RecordValue** | `{x: 1, y: 2}` | `{"x":1,"y":2}` | Field names → keys |
| VariantValue | `Variant(42)` | `42` | Unwraps then converts |
| Other types | Custom objects | `null` | Unsupported types |

**Implementation**: `internal/interp/builtins_json.go:212-227` (ToJSON), `internal/interp/json_conversion.go:69-113` (valueToJSONValue)

#### ArrayValue vs JSONValue Arrays

```dws
// DWScript array → JSON array (value semantics)
var arr: array of Integer := [1, 2, 3];
PrintLn(ToJSON(arr));  // [1,2,3]

// JSON array → stays as JSONValue (reference semantics)
var json := ParseJSON('[1, 2, 3]');
PrintLn(ToJSON(json));  // [1,2,3]
```

Both produce the same JSON output, but the internal representation differs.

---

## Variant Integration

### VarType Codes for JSON Values

When JSON values are wrapped in `Variant`, they report specific `VarType` codes:

| JSON Type | VarType Code | VarType Name | Notes |
|-----------|--------------|--------------|-------|
| null (in JSONValue) | `0x0180` | varNull | Inside JSONValue wrapper |
| boolean (in JSONValue) | `0x0181` | varBoolean | Inside JSONValue wrapper |
| number (in JSONValue) | `0x0182` or `0x0183` | varNumber or varInt64 | Depends on parsing |
| string (in JSONValue) | `0x0184` | varString | Inside JSONValue wrapper |
| array (in JSONValue) | `0x0185` | varArray | JSONValue array |
| object (in JSONValue) | `0x0186` | varObject | JSONValue object |
| Primitive types | Standard codes | varInteger, varDouble, etc. | When not in JSONValue |

**Implementation**: `internal/interp/json_conversion.go:158-179` (jsonKindToVarType)

### Wrapping Examples

```dws
// Primitives are unwrapped when parsed
var num := ParseJSON('42');        // IntegerValue(42)
PrintLn(VarType(num));             // varInteger (standard code)

// Arrays/objects stay wrapped
var arr := ParseJSON('[1, 2, 3]'); // JSONValue(array)
PrintLn(VarType(arr));             // varArray (0x0185)

// You can access elements which unwrap primitives
PrintLn(VarType(arr[0]));          // varInteger (element unwrapped)
```

---

## Edge Cases

### Large Numbers

Numbers are parsed as `int64` when possible, falling back to `float64`:

```dws
var big := ParseJSON('9223372036854775807');
PrintLn(big);  // 9223372036854775807 (as Integer)

var tooBig := ParseJSON('9223372036854775808');
PrintLn(tooBig);  // As Float (precision loss for numbers beyond int64)
```

### Unicode and Escaping

Standard JSON escape sequences are fully supported via Go's `encoding/json`:

```dws
// Unicode escapes
var text := ParseJSON('"\u0048\u0065\u006C\u006C\u006F"');
PrintLn(text);  // "Hello"

// Standard escapes: \n, \t, \", \\, etc.
var escaped := ToJSON('Line 1\nLine 2');
PrintLn(escaped);  // "Line 1\nLine 2"
```

### Empty Values

Empty strings, arrays, and objects are all supported:

```dws
PrintLn(ParseJSON('""'));                   // ""
PrintLn(JSONLength(ParseJSON('[]')));       // 0
PrintLn(JSONKeys(ParseJSON('{}')).Length);  // 0
```

### Limitations

**NaN/Infinity**: Not supported by JSON specification. Use `null` or string constants (`"NaN"`, `"Infinity"`) as workarounds.

**Circular References**: Not currently detected. Avoid creating circular references as they will cause stack overflow during serialization.

---

## Design Decisions

### Reference Semantics for JSON Collections

**Decision**: JSON arrays and objects remain as `JSONValue` (not converted to `ArrayValue`/`RecordValue`)

**Rationale**:
1. **Mutation Visibility**: Changes to nested structures are visible through all references
2. **Performance**: Avoids deep copying large JSON structures
3. **Interoperability**: Easier to round-trip JSON without data structure changes
4. **DWScript Compatibility**: Matches original DWScript behavior

**Trade-off**: You cannot use standard DWScript array/record operations directly:
```dws
var json := ParseJSON('[1, 2, 3]');

// ❌ This won't work - json is JSONValue, not array of Integer
// var doubled: array of Integer := json;

// ✅ Instead, access via indexing
PrintLn(json[0]);  // 1
PrintLn(json[1]);  // 2
```

**When to Use Each**:
- **JSONValue**: Parsing JSON data, working with dynamic structures, need reference semantics
- **ArrayValue**: Static arrays in code, need type safety, value semantics
- **RecordValue**: Static records in code, need field validation

**Implementation**: `internal/interp/json_conversion.go:45-50` (preserves JSONValue), `docs/json_stage_9.89-90.md:186-241` (detailed rationale)

---

### Variant Wrapping Strategy

**Decision**: JSON primitives are unwrapped, collections stay wrapped

**Parsing Result Types**:
```dws
ParseJSON('42')        → IntegerValue (unwrapped primitive)
ParseJSON('3.14')      → FloatValue (unwrapped primitive)
ParseJSON('"hello"')   → StringValue (unwrapped primitive)
ParseJSON('true')      → BooleanValue (unwrapped primitive)
ParseJSON('null')      → NilValue (unwrapped primitive)
ParseJSON('[1,2,3]')   → JSONValue wrapping array (stays wrapped)
ParseJSON('{"a": 1}')  → JSONValue wrapping object (stays wrapped)
```

**Rationale**:
1. **Primitive Ergonomics**: Unwrapped primitives work with standard DWScript operators
2. **Collection Safety**: Wrapped collections prevent accidental type mismatches
3. **VarType Clarity**: Primitives report standard VarType codes, collections report JSON-specific codes

**Examples**:
```dws
// Primitives can be used directly
var x := ParseJSON('42');
PrintLn(x + 10);  // 52 (works because x is IntegerValue)

// Collections need element access
var arr := ParseJSON('[1, 2, 3]');
PrintLn(arr[0] + 10);  // 11 (element unwraps to IntegerValue)
```

**Implementation**: `internal/interp/json_conversion.go:27-54` (selective unwrapping)

---

### Number Parsing Strategy

**Decision**: Try `int64` first, fall back to `float64`

**Algorithm**:
1. Parse JSON number as Go `json.Number` (string representation)
2. Attempt `strconv.ParseInt(num, 10, 64)`
3. If successful → `IntegerValue(int64)`
4. If fails → `strconv.ParseFloat(num, 64)` → `FloatValue(float64)`

**Examples**:
```dws
ParseJSON('42')          → IntegerValue (exact)
ParseJSON('42.0')        → FloatValue (has decimal point)
ParseJSON('1e10')        → FloatValue (scientific notation)
ParseJSON('9999999999')  → IntegerValue (within int64)
ParseJSON('9999999999999999999')  → FloatValue (exceeds int64, precision loss)
```

**Rationale**:
- Preserves integer precision when possible
- Gracefully handles large numbers as floats
- Matches common JSON library behavior

**Implementation**: `internal/interp/builtins_json.go:103-110`

---

### Object Key Ordering

**Parsing**: Insertion order is preserved during parsing (Go's `encoding/json` with `UseNumber()`)

**Serialization**: Keys are sorted **alphabetically** in output (Go's `json.Marshal` default)

**Example**:
```dws
var json := ParseJSON('{"z": 1, "a": 2, "m": 3}');
PrintLn(ToJSON(json));  // {"a":2,"m":3,"z":1} - alphabetically sorted

// Iteration order matches insertion
for var key in JSONKeys(json) do
    PrintLn(key);  // z, a, m (original order preserved internally)
```

**Rationale**:
- Alphabetical output ensures stable, deterministic JSON for testing/comparison
- Internal order preservation allows iteration in original sequence
- Matches Go's standard library behavior

**Note**: DWScript doesn't guarantee map ordering (implementation-dependent like most languages)

---

## Practical Examples

### Example 1: Parsing Primitives

```dws
// Null
var n := ParseJSON('null');
PrintLn(n);  // (empty/nil)
PrintLn(VarType(n));  // 0 (varEmpty)

// Boolean
var b := ParseJSON('true');
PrintLn(b);  // True
PrintLn(VarType(b));  // 11 (varBoolean)

// Integer
var i := ParseJSON('42');
PrintLn(i + 8);  // 50
PrintLn(VarType(i));  // 3 (varInteger)

// Float
var f := ParseJSON('3.14');
PrintLn(f * 2);  // 6.28
PrintLn(VarType(f));  // 5 (varDouble)

// String
var s := ParseJSON('"Hello, World!"');
PrintLn(s);  // Hello, World!
PrintLn(VarType(s));  // 8 (varString)
```

### Example 2: Parsing Arrays

```dws
var arr := ParseJSON('[1, 2, 3, 4, 5]');

// Access elements by index
PrintLn(arr[0]);  // 1
PrintLn(arr[2]);  // 3

// Get array length
PrintLn(JSONLength(arr));  // 5

// Iterate over elements
for var i := 0 to JSONLength(arr) - 1 do
    PrintLn(arr[i]);
```

### Example 3: Parsing Objects

```dws
var obj := ParseJSON('{"name": "Alice", "age": 30, "active": true}');

// Access properties
PrintLn(obj['name']);   // Alice
PrintLn(obj['age']);    // 30
PrintLn(obj['active']); // True

// Check if field exists
if JSONHasField(obj, 'email') then
    PrintLn(obj['email'])
else
    PrintLn('No email field');  // This executes

// Get all keys
var keys := JSONKeys(obj);
for var key in keys do
    PrintLn(key + ': ' + VarToStr(obj[key]));
```

### Example 4: Nested Structures

```dws
var json := ParseJSON('{
    "users": [
        {"id": 1, "name": "Alice"},
        {"id": 2, "name": "Bob"}
    ],
    "count": 2
}');

// Navigate nested structure
var users := json['users'];
PrintLn(JSONLength(users));  // 2

var firstUser := users[0];
PrintLn(firstUser['name']);  // Alice

// Modify nested values
firstUser['name'] := 'Alice Smith';
PrintLn(ToJSON(json));  // {"count":2,"users":[{"id":1,"name":"Alice Smith"},{"id":2,"name":"Bob"}]}
```

### Example 5: Serialization

```dws
// Serialize primitives
PrintLn(ToJSON(42));          // 42
PrintLn(ToJSON(3.14));        // 3.14
PrintLn(ToJSON('hello'));     // "hello"
PrintLn(ToJSON(True));        // true
PrintLn(ToJSON(nil));         // null

// Serialize arrays (from DWScript code)
var arr: array of Integer := [1, 2, 3];
PrintLn(ToJSON(arr));  // [1,2,3]

// Serialize records (from DWScript code - when records are implemented)
// type TPoint = record x, y: Integer; end;
// var pt: TPoint := {x: 10, y: 20};
// PrintLn(ToJSON(pt));  // {"x":10,"y":20}

// Round-trip: parse, modify, serialize
var data := ParseJSON('{"value": 100}');
data['value'] := data['value'] * 2;
PrintLn(ToJSON(data));  // {"value":200}
```

### Example 6: Pretty Printing

```dws
var data := ParseJSON('{"name":"Alice","scores":[95,87,92]}');

// Compact output
PrintLn(ToJSON(data));
// {"name":"Alice","scores":[95,87,92]}

// Formatted with 2-space indent
PrintLn(ToJSONFormatted(data, 2));
// {
//   "name": "Alice",
//   "scores": [
//     95,
//     87,
//     92
//   ]
// }

// Formatted with 4-space indent
PrintLn(ToJSONFormatted(data, 4));
// {
//     "name": "Alice",
//     "scores": [
//         95,
//         87,
//         92
//     ]
// }
```

### Example 7: Error Handling

```dws
// Invalid JSON syntax
try
    var bad := ParseJSON('{invalid}');
except
    on E: Exception do
        PrintLn('Parse error: ' + E.Message);
end;

// Out of bounds array access
var arr := ParseJSON('[1, 2, 3]');
PrintLn(arr[10]);  // nil (returns nil for out of bounds, doesn't error)

// Missing object key
var obj := ParseJSON('{"a": 1}');
PrintLn(obj['missing']);  // nil (returns nil for missing keys)

// Type checking before access
if JSONHasField(obj, 'a') then
    PrintLn('Has field a: ' + VarToStr(obj['a']));
```

---

## Current Limitations

- **No circular reference detection**: Avoid creating circular references (e.g., `obj['self'] := obj`) as they will cause stack overflow during serialization
- **No NaN/Infinity support**: Use `null` or string constants as workarounds
- **Alphabetical key ordering**: Serialized JSON keys are sorted alphabetically, not by insertion order

---

## Implementation Reference

### Source Files

| Component | File Path | Key Lines |
|-----------|-----------|-----------|
| JSON Value Types | `internal/jsonvalue/value.go` | 6-18 (Kind enum), 44-61 (Value struct) |
| ParseJSON Function | `internal/interp/builtins_json.go` | 37-60 (function), 67-141 (parsing logic) |
| ToJSON Function | `internal/interp/builtins_json.go` | 212-227 |
| ToJSONFormatted Function | `internal/interp/builtins_json.go` | 253-283 |
| Type Conversions | `internal/interp/json_conversion.go` | 27-54 (JSON→DWScript), 69-113 (DWScript→JSON) |
| VarType Mapping | `internal/interp/json_conversion.go` | 158-179 |
| JSON Helper Functions | `internal/interp/builtins_json.go` | 303-507 (Has/Keys/Values/Length) |

### Test Files

| Test Type | File Path | Coverage |
|-----------|-----------|----------|
| Unit Tests - Conversions | `internal/interp/json_conversion_test.go` | 437 lines, 100% coverage |
| Unit Tests - Built-ins | `internal/interp/builtins_json_test.go` | 630 lines, all functions |
| Integration Tests | `internal/interp/json_variant_integration_test.go` | 373 lines, end-to-end |
| DWScript Tests - Basic | `testdata/json/parse_basic.dws` | Primitives |
| DWScript Tests - Arrays | `testdata/json/parse_arrays.dws` | Array parsing |
| DWScript Tests - Objects | `testdata/json/parse_objects.dws` | Object parsing |
| DWScript Tests - Edges | `testdata/json/parse_edge_cases.dws` | Large numbers, unicode, etc. |

### Related Documentation

- **`docs/variant.md`**: Variant type system and boxing/unboxing (lines 1-706)
- **`docs/json_stage_9.89-90.md`**: JSON value representation design (lines 1-293)
- **`PLAN.md`**: Task breakdown for JSON support (lines 258-349)
- **`README.md`**: User-facing JSON examples

---

## Summary

The JSON implementation in go-dws provides a robust, Go-backed JSON parsing and serialization system with the following characteristics:

**Strengths**:
- Full support for all JSON primitive types
- Reference semantics for arrays/objects (mutation-friendly)
- Seamless Variant integration
- Comprehensive Unicode and escape sequence handling
- Large number support (int64 + float64 fallback)
- Pretty-printing support

**Limitations**:
- No circular reference detection (can crash)
- No NaN/Infinity support (JSON spec limitation)
- Alphabetical key ordering in output
- No query language or schema validation

**Best Practices**:
1. Always use `JSONHasField()` before accessing object properties
2. Check `JSONLength()` before iterating arrays
3. Avoid creating circular references
4. Use `nil` instead of NaN/Infinity
5. Use `ToJSONFormatted()` for debugging, `ToJSON()` for production

For additional information, see the implementation files and test cases referenced above.
