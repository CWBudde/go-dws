# Built-in Functions Reference

This document provides detailed documentation for built-in functions implemented in go-dws.

## String Functions

### Format

Formats a string using format specifiers similar to C's `printf` or Go's `fmt.Sprintf`.

**Syntax:**
```pascal
function Format(fmt: String; args: array of Variant): String;
```

**Parameters:**
- `fmt`: Format string containing text and format specifiers
- `args`: Array of values to be formatted and inserted into the format string

**Format Specifiers:**

| Specifier | Type | Description | Example |
|-----------|------|-------------|---------|
| `%s` | String | String value | `Format('Hello %s', ['World'])` → `'Hello World'` |
| `%d` | Integer | Decimal integer | `Format('Value: %d', [42])` → `'Value: 42'` |
| `%f` | Float | Floating-point number | `Format('Pi: %f', [3.14159])` → `'Pi: 3.141590'` |
| `%%` | - | Literal percent sign | `Format('100%%', [])` → `'100%'` |

**Width and Precision:**

You can specify width and precision for format specifiers:

- **Width**: Minimum number of characters to output
  - `%5d` - Right-aligned integer with minimum width 5
  - Example: `Format('%5d', [42])` → `'   42'`

- **Precision for floats**: Number of decimal places
  - `%.2f` - Float with 2 decimal places
  - Example: `Format('%.2f', [3.14159])` → `'3.14'`

- **Combined**: Width and precision
  - `%8.2f` - Float with width 8 and 2 decimal places
  - Example: `Format('%8.2f', [3.14])` → `'    3.14'`

**Examples:**

```pascal
// Simple string formatting
var arr: array of String;
SetLength(arr, 1);
arr[0] := 'World';
PrintLn(Format('Hello %s', arr));  // Output: Hello World

// Integer formatting
var nums: array of Integer;
SetLength(nums, 1);
nums[0] := 42;
PrintLn(Format('The answer is %d', nums));  // Output: The answer is 42

// Float with precision
var floats: array of Float;
SetLength(floats, 1);
floats[0] := 3.14159;
PrintLn(Format('Pi: %.2f', floats));  // Output: Pi: 3.14

// Multiple arguments
var mixed: array of String;
SetLength(mixed, 2);
mixed[0] := 'John';
mixed[1] := '30';
PrintLn(Format('%s is %s years old', mixed));  // Output: John is 30 years old

// Width formatting
var ints: array of Integer;
SetLength(ints, 1);
ints[0] := 42;
PrintLn(Format('Value: %5d', ints));  // Output: Value:    42

// Literal percent sign
var pct: array of Integer;
SetLength(pct, 1);
pct[0] := 100;
PrintLn(Format('%d%% complete', pct));  // Output: 100% complete
```

**Type Validation:**

The Format function validates that the types of the arguments match the format specifiers:

- `%d` requires Integer values
- `%f` requires Float values
- `%s` accepts String values (and can convert Integer/Float to strings)

If there's a type mismatch, a runtime error will be raised:

```pascal
var arr: array of String;
SetLength(arr, 1);
arr[0] := 'not a number';
Format('Value: %d', arr);  // ERROR: Cannot use %d with String value
```

**Argument Count Validation:**

The number of format specifiers in the format string must match the number of elements in the args array:

```pascal
var arr: array of String;
SetLength(arr, 1);
arr[0] := 'World';

Format('Hello %s %s', arr);  // ERROR: expects 2 arguments, got 1
Format('Hello', arr);        // ERROR: expects 0 arguments, got 1
```

**Implementation Notes:**

- Uses Go's `fmt.Sprintf()` internally for robust formatting
- Supports all standard Go format verbs: `%s`, `%d`, `%f`, `%v`, `%x`, `%X`, `%o`
- Width and precision modifiers work as in Go's format strings
- Format specifiers are parsed and validated before formatting
- Type checking ensures format specifiers match argument types

**See Also:**
- `IntToStr()` - Convert integer to string
- `FloatToStr()` - Convert float to string
- `Concat()` - Concatenate strings

---

### StringReplace

Replaces occurrences of a substring within a string.

**Syntax:**
```pascal
function StringReplace(s: String; old: String; new: String): String;
function StringReplace(s: String; old: String; new: String; count: Integer): String;
```

**Parameters:**
- `s`: The string to search in
- `old`: The substring to find
- `new`: The replacement substring
- `count` (optional): Number of replacements (-1 or omitted = replace all)

**Examples:**

```pascal
// Replace all occurrences
PrintLn(StringReplace('hello world', 'l', 'L'));  // Output: heLLo worLd

// Replace first occurrence only
PrintLn(StringReplace('foo bar foo', 'foo', 'baz', 1));  // Output: baz bar foo

// Replace first two occurrences
PrintLn(StringReplace('a a a a', 'a', 'b', 2));  // Output: b b a a
```

---

### Insert

Inserts a string into another string at a specified position (modifies the target string in-place).

**Syntax:**
```pascal
procedure Insert(source: String; var target: String; pos: Integer);
```

**Parameters:**
- `source`: The string to insert
- `target`: The string to insert into (modified in-place)
- `pos`: 1-based position where to insert

**Examples:**

```pascal
var s: String := 'Helo';
Insert('l', s, 3);  // s becomes 'Hello'

var s2: String := 'Hello';
Insert(' World', s2, 6);  // s2 becomes 'Hello World'
```

---

### Delete

Deletes characters from a string (modifies the string in-place).

**Syntax:**
```pascal
procedure Delete(var s: String; pos: Integer; count: Integer);
```

**Parameters:**
- `s`: The string to delete from (modified in-place)
- `pos`: 1-based starting position
- `count`: Number of characters to delete

**Examples:**

```pascal
var s: String := 'Hello';
Delete(s, 3, 2);  // s becomes 'Heo'

var s2: String := 'HelloXXXWorld';
Delete(s2, 6, 3);  // s2 becomes 'HelloWorld'
```

---

### Trim, TrimLeft, TrimRight

Remove whitespace from strings.

**Syntax:**
```pascal
function Trim(s: String): String;
function TrimLeft(s: String): String;
function TrimRight(s: String): String;
```

**Parameters:**
- `s`: The string to trim

**Examples:**

```pascal
PrintLn(Trim('  hello  '));       // Output: hello
PrintLn(TrimLeft('  hello  '));   // Output: hello
PrintLn(TrimRight('  hello  '));  // Output:   hello
```

**Whitespace Characters:**
- Space (' ')
- Tab ('\t')
- Newline ('\n')
- Carriage return ('\r')

---

## Implementation Status

✅ **Fully Implemented:**
- Format - String formatting with format specifiers
- StringReplace - Substring replacement
- Insert - Insert string at position
- Delete - Delete substring
- Trim, TrimLeft, TrimRight - Whitespace removal
- Copy - Substring extraction
- Concat - String concatenation
- Pos - Find substring position
- UpperCase, LowerCase - Case conversion
- Length - String length
- IntToStr, StrToInt - Integer conversion
- FloatToStr, StrToFloat - Float conversion

⏸️ **Planned:**
- Chr, Ord - Character/ASCII conversion
- StringOfChar - Repeat character
- ReverseString - Reverse string
- CompareStr, CompareText - String comparison
- AnsiUpperCase, AnsiLowerCase - Locale-aware case conversion
