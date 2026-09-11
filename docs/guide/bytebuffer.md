# ByteBuffer

**Status**: Implemented (PLAN.md §3.3, `FunctionsByteBuffer` 19/19)

`ByteBuffer` is a built-in, resizable block of raw bytes with a read/write
cursor. It is the type to reach for when a script has to build or pick apart a
binary layout: file headers, network frames, fixed-width records, or anything
that arrives as base64 or hex.

## Declaring and creating

A declared `ByteBuffer` variable is **auto-instantiated**: it is usable
immediately and starts empty, so there is no nil state to guard against.

```pascal
var b : ByteBuffer;      // empty, ready to use
PrintLn(b.Length);       // 0

var c := new ByteBuffer; // explicit, also empty
var d := ByteBuffer('hello');  // from a data string
```

`ByteBuffer(s)` reads a string as a *data string*: one byte per UTF-16 code
unit, keeping only that unit's low byte. `ByteBuffer(#$1234#$5678)` therefore
holds the two bytes `$34 $78`. Strings are the only type that converts to a
`ByteBuffer`.

## Value versus reference

`ByteBuffer` has **reference semantics**. Assigning one buffer to another makes
both names refer to the same storage:

```pascal
var a := new ByteBuffer;
var b := a;
b.SetLength(4);
PrintLn(a.Length);   // 4 - a and b are the same buffer
```

`Assign` is the operation that copies content, leaving the two buffers
independent afterwards:

```pascal
var src := ByteBuffer('hello');
var dst : ByteBuffer;
dst.Assign(src);
src.SetLength(2);
PrintLn(src.ToDataString);   // he
PrintLn(dst.ToDataString);   // hello
```

`Copy` returns a fresh buffer holding a slice of the original.

## Length, position and the two accessor forms

`Length` and `Position` are read-only; `SetLength` and `SetPosition` change
them. Growing a buffer zero-fills the new bytes, and shrinking then regrowing
gives zeroes again rather than the old content.

Every typed accessor comes in two forms:

- the **cursor form** — `b.GetByte`, `b.SetByte(v)` — reads or writes at
  `Position` and advances it by the accessor's width;
- the **indexed form** — `b.GetByte(i)`, `b.SetByte(i, v)` — addresses the
  buffer directly and leaves `Position` alone.

A failed access leaves the cursor untouched, so the next operation retries at
the same place.

```pascal
var b : ByteBuffer;
b.SetLength(4);
b.SetWord(258);        // writes at 0, Position becomes 2
b.SetWord(2, 772);     // writes at 2, Position stays 2
PrintLn(b.ToJSON);     // [2,1,4,3]
```

## Typed accessors

All multi-byte accessors are **little-endian**.

| Accessor | Width | Range |
| --- | --- | --- |
| `Byte` | 1 | 0 … 255 |
| `Int8` | 1 | -128 … 127 |
| `Word` | 2 | 0 … 65535 |
| `Int16` | 2 | -32768 … 32767 |
| `DWord` | 4 | 0 … 4294967295 |
| `Int32` | 4 | -2147483648 … 2147483647 |
| `Int64` | 8 | full Int64 |
| `Single` | 4 | IEEE-754 binary32 |
| `Double` | 8 | IEEE-754 binary64 |
| `Extended` | 10 | x87 80-bit |

Each has a `Get<Name>` and a `Set<Name>`. `Extended` is stored in the genuine
10-byte x87 layout; reading one widens it into a `Float`, so precision beyond
53 mantissa bits is lost.

`GetIntegers(index, count, size, signed)` reads `count` little-endian integers
of `size` bytes and returns them as an `array of Integer`, ready for `Map` and
`Join`. Note that `index` takes part in the bounds check but does not shift the
read origin — the elements always start at byte 0. That is the reference
implementation's behaviour, pinned by
`testdata/fixtures/FunctionsByteBuffer/integers`.

## Strings and encodings

| Member | Result |
| --- | --- |
| `GetData(index, size)` | a data string of `size` bytes at `index` |
| `SetData(s)` / `SetData(index, s)` | writes a data string |
| `ToDataString` | the whole buffer as a data string |
| `ToHexString` | lowercase hexadecimal |
| `ToBase64` | standard padded base64 |
| `ToJSON` | a JSON array of byte values, e.g. `[1,2,3]` |
| `AssignDataString(s)` | replaces the content from a data string |
| `AssignHexString(s)` | replaces the content from hexadecimal |
| `AssignBase64(s)` | replaces the content from base64 |
| `AssignJSON(s)` | replaces the content from a JSON array of numbers |

Every `Assign*` rewinds `Position` to 0.

```pascal
var b : ByteBuffer;
b.AssignDataString('hello world');
PrintLn(b.ToBase64);       // aGVsbG8gd29ybGQ=
b.AssignBase64('dGVzdGluZw==');
PrintLn(b.ToDataString);   // testing
```

## Errors

Range and overflow failures raise catchable `Exception`s.

```pascal
var b : ByteBuffer;
try
    b.GetByte(0);
except
    on E : Exception do PrintLn(E.Message);
end;
// Out of range (index 0, size 1 for length 0) [line: 3, column: 7]
```

The three message shapes are:

- `Out of range (index I, size S for length L)` — the access does not fit;
- `Position P out of range (length L)` — `SetPosition` outside `0 … Length`;
- `value V out of T range` — a `Set<T>` argument outside the accessor's range.

The value check runs before the range check, so `b.SetByte(256)` reports the
overflow even when the cursor is also past the end.

## Parameterless calls

DWScript lets a parameterless member be written without parentheses, and
`ByteBuffer` follows that: `b.ToJSON`, `b.Length` and `b.GetByte` are all
valid, the last one reading at the cursor and advancing it.
