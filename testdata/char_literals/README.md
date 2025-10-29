# Character Literal Test Scripts

This directory contains test scripts for character literal support in go-dws (Task 9.27-9.31).

## Overview

Character literals in DWScript provide three ways to represent single characters:
1. **Decimal ordinal**: `#65` (character with ASCII/Unicode code 65, i.e., 'A')
2. **Hexadecimal**: `#$41` (character with hex code 0x41, i.e., 'A')
3. **Single-character string**: `'A'` (treated as a string literal, not covered in these tests)

Character literals are treated as **single-character strings** and can be used wherever strings are expected.

## Test Files

| Test File | Description |
|-----------|-------------|
| `basic_assignments.dws` | Basic character literal assignments to variables with both decimal and hex formats |
| `case_statements.dws` | Character literals used as case statement labels, testing pattern matching |
| `ordinal_hex.dws` | Comprehensive tests for both decimal (#65) and hexadecimal (#$41) formats |
| `string_concat.dws` | String concatenation with character literals, building complex strings |

## Test Coverage

These tests cover the following character literal features:

### Basic Syntax
- **Decimal format**: `#` followed by decimal digits (e.g., `#65`, `#13`, `#10`)
- **Hexadecimal format**: `#$` followed by hex digits (e.g., `#$41`, `#$0D`, `#$0A`)
- **Type compatibility**: Character literals are strings and work in string contexts

### Use Cases

#### Variable Assignment
```pascal
var ch: String := #65;        // Explicit type
var letter := #72;            // Type inference
```

#### String Concatenation
```pascal
var hello := #72 + #101 + #108 + #108 + #111;  // "Hello"
var greeting := 'Hi' + #33;                     // "Hi!"
```

#### Case Statement Labels
```pascal
case key of
  #65: PrintLn('A');
  #$42: PrintLn('B');
  #13: PrintLn('Enter');
end;
```

### Character Categories Tested

- **ASCII Letters**: Uppercase (`#65`-`#90`) and lowercase (`#97`-`#122`)
- **Digits**: `#48`-`#57` (characters '0'-'9')
- **Whitespace**: Space (`#32`), tab (`#9`), CR (`#13`), LF (`#10`)
- **Special Characters**: `#33` (!), `#64` (@), `#35` (#), etc.
- **Punctuation**: Comma (`#44`), period (`#46`), quotes (`#34`, `#39`)
- **Control Characters**: NULL (`#0`), DEL (`#127`)

### Format Variations

- Leading zeros in hex: `#$0041` same as `#$41`
- Mixed decimal and hex in expressions: `#72 + #$65 + #108`
- Large values for Unicode: `#$20AC` (Euro symbol)

## Expected Behavior

1. **Type**: Character literals evaluate to `String` type
2. **Concatenation**: Can concatenate with strings using `+` operator
3. **Comparison**: Can compare with strings using `=`, `<>`, etc.
4. **Case Matching**: Work as case labels and match against string values
5. **Length**: Single-character strings have `Length() = 1`

## Implementation Status

Character literal support is implemented across all compilation stages:

- [x] **Lexer** (Task 9.27): `CHAR` token type, handles `#65` and `#$41` formats
- [x] **AST** (Task 9.27): `CharLiteral` node with rune value
- [x] **Parser** (Task 9.28): Parses `CHAR` tokens to `CharLiteral` AST nodes
- [x] **Semantic Analysis** (Task 9.29): Type checks as `STRING`
- [x] **Interpreter** (Task 9.30): Evaluates to `StringValue`
- [x] **Integration Tests** (Task 9.31): End-to-end test scripts

## Testing These Scripts

Run these tests with the go-dws interpreter:

```bash
# Run a single test
./bin/dwscript run testdata/char_literals/basic_assignments.dws

# Run all character literal tests from the test directory
for f in testdata/char_literals/*.dws; do
  echo "=== Testing $f ==="
  ./bin/dwscript run "$f"
  echo
done
```

## Language Reference

Character literals follow the DWScript/Object Pascal syntax:

- **Decimal**: `#` + decimal number (0-1114111 for Unicode)
- **Hexadecimal**: `#$` + hex digits (case insensitive)
- **Type**: Character literals are strings, not a separate char type
- **Common uses**: Line endings (`#13#10`), special formatting, building strings from codes

## Examples from Rosetta Code

Several Rosetta Code examples use character literals:

- `examples/rosetta/Execute_HQ9.dws`: Uses `#13#10` for line breaks
- `examples/rosetta/Literals_String.dws`: Demonstrates `#13#10` in strings

## Known Limitations

1. **Unicode Support**: Full Unicode support depends on Go's `rune` handling
2. **Escape Sequences**: No alternative escape sequences like `\n`, `\r` - must use char literals
3. **Single Character Check**: No compile-time validation that values produce printable characters

## Notes

- Character literals are **not** a distinct type - they're syntactic sugar for single-character strings
- Both formats (`#65` and `#$41`) can be used interchangeably
- Hexadecimal is case-insensitive: `#$41`, `#$4A`, `#$4a` all valid
- Character codes outside valid Unicode range may cause undefined behavior
- Empty character literals (`#`) are a syntax error
