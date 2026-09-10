# Encoder classes (EncodingLib)

go-dws ships DWScript's `EncodingLib` encoders as built-in classes. They are always
available — no `uses` clause — and every one of them exposes the same pair of class
methods:

```pascal
class function Encode(const s : String) : String; virtual;
class function Decode(const s : String) : String; virtual;
```

## The classes

| Class | `Encode` produces | Notes |
| --- | --- | --- |
| `Encoder` | — | Abstract root; use it as `class of Encoder`. |
| `Base64Encoder` | standard, padded Base64 | Also has `EncodeMIME` (see below). `Decode` ignores spaces, tabs, CR and LF. |
| `Base64URIEncoder` | URL and filename safe Base64, unpadded | RFC 4648 §5 (`-` and `_` instead of `+` and `/`). |
| `Base32Encoder` | RFC 4648 Base32, unpadded | `Decode` is case-insensitive and accepts `0` as an alias for `O`. |
| `Base58Encoder` | Bitcoin Base58 | Leading zero bytes become leading `1` characters. |
| `HexadecimalEncoder` | lowercase hexadecimal | `Decode` accepts both cases. |
| `UTF8Encoder` | the string's UTF-8 bytes | |
| `UTF16BigEndianEncoder` | the string's UTF-16BE bytes | |
| `UTF16LittleEndianEncoder` | the string's UTF-16LE bytes | |
| `URLEncodedEncoder` | percent-encoded UTF-8 | Everything outside `A-Z a-z 0-9 - . _ ~` is escaped. |
| `HTMLTextEncoder` | HTML text | Escapes `&`, `<`, `>`, `"`, `'` and U+00A0. `Decode` drops tags and resolves character references. |
| `HTMLAttributeEncoder` | HTML attribute value | Escapes every non-alphanumeric character as a numeric reference. |

## Strings and bytes

DWScript strings are sequences of 16-bit characters. The byte-oriented encoders
(`Base64Encoder`, `Base32Encoder`, `Base58Encoder`, `HexadecimalEncoder`) read and
write *byte strings*: strings in which every character holds one byte value. That is
what `UTF8Encoder.Encode` and the UTF-16 encoders produce, so the encoders compose:

```pascal
PrintLn(HexadecimalEncoder.Encode(UTF8Encoder.Encode('éric')));
// c3a9726963
```

## Encoders as values

An encoder class is an ordinary class reference, so it can be stored in a variable or
passed to a routine, and the class methods dispatch virtually:

```pascal
var encoder := Base64Encoder;
PrintLn(encoder.Decode(encoder.Encode('hello')));

procedure Show(e : class of Encoder; s : String);
begin
   PrintLn(HexadecimalEncoder.Encode(e.Encode(s)));
end;

Show(UTF16BigEndianEncoder, 'Example');     // 004500780061006d0070006c0065
Show(UTF16LittleEndianEncoder, 'Example');  // 4500780061006d0070006c006500
```

## MIME output

`Base64Encoder.EncodeMIME` produces the same Base64 as `Encode` but wraps it at 76
characters per line, as RFC 2045 requires. `Base64Encoder.Decode` reads the wrapped
form back, so `Decode(EncodeMIME(s)) = s`.

## Errors

Malformed input raises a catchable `Exception`:

```pascal
try
   PrintLn(Base32Encoder.Decode('....'));
except
   on E : Exception do
      PrintLn(E.Message);  // Invalid character (#46) in Base32 [line: 3, column: 4]
end;
```

The messages are DWScript's:

| Encoder | Condition | Message |
| --- | --- | --- |
| `Base32Encoder` | character outside the alphabet | `Invalid character (#N) in Base32` |
| `Base58Encoder` | character outside the alphabet | `Non-base58 character` |
| `HexadecimalEncoder` | odd number of characters | `Even hexadecimal character count expected` |
| `HexadecimalEncoder` | non-hex character | `Invalid hexadecimal character at index N` (1-based) |

The encoders that cannot fail never raise: `URLEncodedEncoder.Decode` substitutes
U+FFFD for a broken escape sequence, and `HTMLTextEncoder.Decode` leaves an
unrecognized character reference untouched.
