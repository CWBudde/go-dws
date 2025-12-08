# String Encoding in go-dws

## Decision: UTF-8 Native

**Date:** 2025-12-08

**Status:** Adopted

## Context

DWScript (the original Delphi implementation) uses UTF-16 as its internal string encoding, following Delphi's native string type. This means:

- Strings are sequences of 16-bit code units
- Characters in the Basic Multilingual Plane (U+0000 to U+FFFF) are single code units
- Characters in supplementary planes (U+10000 to U+10FFFF) are encoded as **surrogate pairs** (two code units)
- `Length()` returns the number of 16-bit code units, not Unicode code points
- `Chr($10000)` returns a 2-character string (the surrogate pair)

## Decision

**go-dws uses UTF-8 encoding for all strings**, following Go's native string type.

This is an **intentional, documented divergence** from DWScript's behavior.

## Rationale

### Why UTF-8?

1. **Go Native** - Go strings are UTF-8 by design. Fighting this would make every string operation complex and error-prone.

2. **Better String Semantics** - With UTF-8:
   - `Length("😀")` returns 1 (one character), not 2
   - `Chr($10000)` returns a single emoji character
   - String indexing and slicing work with Unicode code points, not arbitrary 16-bit units

3. **Maintainability** - UTF-8 keeps the codebase simple and idiomatic. Emulating UTF-16 would require:
   - Custom length calculations
   - Special indexing logic
   - Surrogate pair handling throughout
   - Malformed UTF-8 strings (surrogate code points aren't valid UTF-8)

4. **Modern Best Practice** - UTF-8 is the standard for modern systems. UTF-16 is legacy from Windows/Java/Delphi era.

### Trade-offs

**What we lose:**

- 100% byte-for-byte compatibility with DWScript test fixtures involving supplementary plane characters
- Identical `Length()` results for strings with emoji/rare characters

**What we gain:**

- Clean, maintainable Go code
- Correct Unicode semantics
- No malformed strings
- Better internationalization support

## Implementation

### Chr() Function

```go
// Returns UTF-8 encoded character
Chr($41)      // "A" (length 1)
Chr($10000)   // "𐀀" (length 1, not a surrogate pair)
Chr($1F600)   // "😀" (length 1, not 2)
```

### String Literals

String literals with `#$` syntax create characters by code point:

```dws
var s := #$D800#$DC00;  // Two characters (surrogate code points)
Length(s)               // 2 (each #$ is one character)
```

**Note:** In DWScript, `#$D800#$DC00` forms a surrogate pair representing U+10000. In go-dws, these are two separate (possibly invalid) characters.

### Length() Function

Returns the number of **Unicode code points** (runes in Go terminology), not UTF-16 code units:

```go
Length("A")        // 1
Length("€")        // 1
Length("😀")       // 1 (not 2 like DWScript)
Length("hello😀")  // 6 (not 7)
```

## Test Fixture Compatibility

For test fixtures that check surrogate pair behavior, we **adapt the test script** to verify UTF-8 semantics instead.

**Example:** `testdata/fixtures/FunctionsString/chr.pas`

Original DWScript test:

```pascal
if Chr($10000) <> #$D800#$DC00 then PrintLn('bug U+10000');
```

This compares UTF-16 surrogate pairs, which doesn't make sense for UTF-8.

**Our approach:** Test UTF-8 properties instead:

```pascal
// Verify Chr() returns single characters for supplementary planes
if Length(Chr($10000)) <> 1 then PrintLn('bug U+10000 length');
if Length(Chr($1F600)) <> 1 then PrintLn('bug emoji length');

// Verify string concatenation works correctly
var s := Chr($10000) + Chr($1F600);
if Length(s) <> 2 then PrintLn('bug string concatenation');
```

This tests the **correct UTF-8 behavior** rather than just accepting different output.

## Future Considerations

If we ever need UTF-16 compatibility (e.g., for Windows API interop), we can:

1. Add explicit conversion functions: `UTF8ToUTF16()`, `UTF16ToUTF8()`
2. Keep internal representation as UTF-8
3. Convert only at system boundaries

## References

- Go Blog: [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
- Unicode Standard: [UTF-8, UTF-16, and Surrogates](https://www.unicode.org/faq/utf_bom.html)
- CLAUDE.md: "go-dws port may intentionally diverge in some areas"
