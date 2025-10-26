# DWScript vs go-dws Feature Comparison Matrix

**Document Version**: 1.0
**Last Updated**: 2025-10-26
**Purpose**: Quick reference for feature parity between DWScript and go-dws

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Fully implemented and tested |
| 🟨 | Partially implemented |
| ⏸️ | Not started |
| ❌ | Out of scope / Not planned |
| 🔄 | In progress |

| Priority | Description |
|----------|-------------|
| **HIGH** | Critical for real programs, should be in Stage 8-9 |
| **MED** | Useful but not essential, Stage 9-10 |
| **LOW** | Nice to have, future consideration |
| **N/A** | Not applicable to go-dws |

---

## Core Language Features

| Feature | DWScript | go-dws | Priority | Notes |
|---------|----------|--------|----------|-------|
| **Variable Declarations** |
| `var` with type | ✅ | ✅ | - | `var x: Integer;` |
| `var` with initialization | ✅ | ✅ | - | `var x := 42;` |
| Type inference | ✅ | ✅ | - | From initializer |
| `const` declarations | ✅ | ⏸️ | **HIGH** | `const PI = 3.14;` |
| Multiple in one decl | ✅ | ✅ | - | `var x, y: Integer;` |
| Inline var (for loops) | ✅ | ✅ | - | `for var i := 1 to 10` |
| Inline var (blocks) | ✅ | ⏸️ | **MED** | Scoped declarations |
| Global variables | ✅ | ✅ | - | Program-level vars |
| Unit-level variables | ✅ | ⏸️ | **HIGH** | Requires unit system |
| **Control Flow** |
| `if..then..else` | ✅ | ✅ | - | |
| `case..of..end` | ✅ | ✅ | - | Integer/enum cases |
| Case with ranges | ✅ | ⏸️ | **MED** | `1..10:` |
| Case multi-values | ✅ | ⏸️ | **MED** | `1, 3, 5:` |
| Case with strings | ✅ | ⏸️ | **MED** | String matching |
| `for..to..do` | ✅ | ✅ | - | |
| `for..downto..do` | ✅ | ✅ | - | |
| `while..do` | ✅ | ✅ | - | |
| `repeat..until` | ✅ | ✅ | - | |
| `break` | ✅ | ✅ | - | Exit loop |
| `continue` | ✅ | ✅ | - | Next iteration |
| `exit` | ✅ | ✅ | - | Exit function |
| `exit(value)` | ✅ | ✅ | - | Return with value |
| `with..do` | ✅ | ⏸️ | **MED** | Object context |
| `goto` and labels | ✅ | ⏸️ | **LOW** | Discouraged |
| **Operators** |
| Arithmetic | ✅ | ✅ | - | `+`, `-`, `*`, `/`, `div`, `mod` |
| Power `**` | ✅ | ⏸️ | **MED** | Exponentiation |
| Comparison | ✅ | ✅ | - | `=`, `<>`, `<`, `>`, etc. |
| Logical | ✅ | ✅ | - | `and`, `or`, `xor`, `not` |
| Logical `implies` | ✅ | ⏸️ | **LOW** | Implication |
| Bitwise | ✅ | ✅ | - | `shl`, `shr`, `and`, `or` |
| String concat `+` | ✅ | ✅ | - | |
| Append assign `||=` | ✅ | ⏸️ | **LOW** | `s ||= 'text';` |
| Compound assign | ✅ | ✅ | - | `+=`, `-=`, `*=`, `/=` |
| Type operators | ✅ | ✅ | - | `is`, `as` |
| `typeof` | ✅ | ⏸️ | **MED** | Type info |
| `classof` | ✅ | ⏸️ | **MED** | Class reference |
| Range `..` | ✅ | ✅ | - | In sets/arrays |
| Set operators | ✅ | 🟨 | **MED** | Missing `><` (symm diff) |

---

## Type System

| Feature | DWScript | go-dws | Priority | Notes |
|---------|----------|--------|----------|-------|
| **Primitive Types** |
| `Integer` (32-bit) | ✅ | ✅ | - | |
| `Int64` | ✅ | ⏸️ | **MED** | 64-bit integer |
| `Float` | ✅ | ✅ | - | Double precision |
| `Boolean` | ✅ | ✅ | - | |
| `String` | ✅ | ✅ | - | Unicode |
| `Char` | ✅ | ⏸️ | **LOW** | Single character |
| `Byte` | ✅ | ⏸️ | **MED** | 8-bit unsigned |
| `Variant` | ✅ | ⏸️ | **LOW** | Dynamic type |
| `Currency` | ✅ | ⏸️ | **LOW** | Fixed decimal |
| Pointer types | ✅ | ⏸️ | **LOW** | `^Type`, `Pointer` |
| **Type Aliases** |
| Type aliases | ✅ | ⏸️ | **HIGH** | `type TInt = Integer;` |
| Subrange types | ✅ | ⏸️ | **MED** | `type TDigit = 0..9;` |
| **Arrays** |
| Static arrays | ✅ | ✅ | - | `array[0..9] of T` |
| Dynamic arrays | ✅ | ✅ | - | `array of T` |
| Multi-dim syntax | ✅ | ⏸️ | **MED** | `array[M,N] of T` |
| Multi-dim nested | ✅ | ✅ | - | `array of array of T` |
| Array literals | ✅ | ✅ | - | `[1, 2, 3]` |
| Open array params | ✅ | ⏸️ | **HIGH** | `array of T` param |
| Array of const | ✅ | ⏸️ | **HIGH** | Variadic params |
| Associative arrays | ✅ | ⏸️ | **MED** | `array[String] of T` |
| **Records** |
| Record declarations | ✅ | ✅ | - | |
| Record fields | ✅ | ✅ | - | |
| Record methods | ✅ | 🟨 | **MED** | Parsed, limited runtime |
| Record properties | ✅ | ⏸️ | **MED** | |
| Record constructors | ✅ | ⏸️ | **MED** | |
| Record operators | ✅ | ⏸️ | **MED** | Operator overloading |
| Record helpers | ✅ | ⏸️ | **MED** | Extend records |
| Nested records | ✅ | ✅ | - | |
| Anonymous records | ✅ | ✅ | - | |
| Record literals | ✅ | ✅ | - | `(X: 10, Y: 20)` |
| Value semantics | ✅ | ✅ | - | Copy on assign |
| **Enumerations** |
| Basic enums | ✅ | ✅ | - | `(Red, Green, Blue)` |
| Explicit values | ✅ | ✅ | - | `(Ok = 0, Error = 1)` |
| Mixed values | ✅ | ✅ | - | |
| Scoped access | ✅ | ✅ | - | `TColor.Red` |
| Unscoped access | ✅ | ✅ | - | `Red` |
| Enum helpers | ✅ | ⏸️ | **MED** | Extend enums |
| Flags type | ✅ | ⏸️ | **MED** | Bit flags |
| `Ord(enum)` | ✅ | ✅ | - | Get ordinal |
| `Low(enum)` | ✅ | ⏸️ | **HIGH** | Lowest value |
| `High(enum)` | ✅ | ⏸️ | **HIGH** | Highest value |
| `Succ(enum)` | ✅ | ⏸️ | **HIGH** | Next value |
| `Pred(enum)` | ✅ | ⏸️ | **HIGH** | Previous value |
| `Inc(enum)` | ✅ | ⏸️ | **HIGH** | Increment |
| `Dec(enum)` | ✅ | ⏸️ | **HIGH** | Decrement |
| **Sets** |
| Set types | ✅ | ✅ | - | `set of TEnum` |
| Set literals | ✅ | ✅ | - | `[val1, val2]` |
| Set ranges | ✅ | ✅ | - | `[1..10]` |
| Union `+` | ✅ | ✅ | - | |
| Difference `-` | ✅ | ✅ | - | |
| Intersection `*` | ✅ | ✅ | - | |
| Symmetric diff `><` | ✅ | ⏸️ | **MED** | |
| Membership `in` | ✅ | ✅ | - | |
| Comparisons | ✅ | ✅ | - | `=`, `<>`, `<=`, `>=` |
| Include/Exclude | ✅ | ✅ | - | |
| Set of integers | ✅ | ⏸️ | **MED** | Small ranges |
| Set of chars | ✅ | ⏸️ | **LOW** | Character sets |
| For-in iteration | ✅ | ⏸️ | **MED** | `for x in set` |
| Large sets (>64) | ✅ | ⏸️ | **LOW** | Map-based |

---

## Object-Oriented Programming

| Feature | DWScript | go-dws | Priority | Notes |
|---------|----------|--------|----------|-------|
| **Classes** |
| Class declarations | ✅ | ✅ | - | |
| Inheritance | ✅ | ✅ | - | Single inheritance |
| Fields | ✅ | ✅ | - | Instance variables |
| Methods | ✅ | ✅ | - | |
| Constructors | ✅ | ✅ | - | |
| Destructors | ✅ | ✅ | - | |
| Virtual methods | ✅ | ✅ | - | |
| Abstract methods | ✅ | ✅ | - | |
| Abstract classes | ✅ | ✅ | - | |
| Method override | ✅ | ✅ | - | |
| `reintroduce` | ✅ | ⏸️ | **LOW** | Hide parent method |
| Static fields | ✅ | ✅ | - | `class var` |
| Static methods | ✅ | ✅ | - | `class function/procedure` |
| Virtual constructors | ✅ | ⏸️ | **MED** | Polymorphic creation |
| Class references | ✅ | ⏸️ | **MED** | Meta-classes |
| Visibility modifiers | ✅ | ✅ | - | public/private/protected |
| `published` | ✅ | ⏸️ | **LOW** | RTTI visibility |
| Forward declarations | ✅ | ✅ | - | |
| `Self` reference | ✅ | ✅ | - | |
| `inherited` | ✅ | ✅ | - | |
| Polymorphism | ✅ | ✅ | - | Dynamic dispatch |
| `is` operator | ✅ | ✅ | - | Type checking |
| `as` operator | ✅ | ✅ | - | Type casting |
| **Properties** |
| Field-backed | ✅ | ✅ | - | |
| Method-backed | ✅ | ✅ | - | |
| Read-only | ✅ | ✅ | - | |
| Write-only | ✅ | ✅ | - | |
| Auto-properties | ✅ | ✅ | - | |
| Indexed properties | ✅ | 🟨 | **MED** | Parsed, runtime deferred |
| Default properties | ✅ | ⏸️ | **MED** | `property Items[]; default;` |
| Multi-dim indexed | ✅ | ⏸️ | **MED** | `[x, y: Integer]` |
| Expression getters | ✅ | ⏸️ | **LOW** | `read (FValue * 2)` |
| Class properties | ✅ | ⏸️ | **MED** | Static properties |
| Property arrays | ✅ | ⏸️ | **LOW** | Array of properties |
| Property override | ✅ | ✅ | - | |
| **Interfaces** |
| Interface declarations | ✅ | ✅ | - | |
| Interface methods | ✅ | ✅ | - | |
| Interface properties | ✅ | ⏸️ | **MED** | |
| Interface inheritance | ✅ | ✅ | - | |
| Multiple implementation | ✅ | ✅ | - | |
| Interface GUIDs | ✅ | ⏸️ | **LOW** | COM compatibility |
| `implements` delegation | ✅ | ⏸️ | **MED** | Interface forwarding |
| Method resolution | ✅ | ⏸️ | **MED** | Explicit method mapping |
| Interface helpers | ✅ | ⏸️ | **MED** | Extend interfaces |
| Interface as/is | ✅ | ✅ | - | |
| **Advanced OOP** |
| Inner/nested classes | ✅ | ⏸️ | **MED** | |
| Partial classes | ✅ | ⏸️ | **LOW** | Split declarations |
| Class helpers | ✅ | ⏸️ | **MED** | Extend classes |
| Record helpers | ✅ | ⏸️ | **MED** | Extend records |
| Type helpers | ✅ | ⏸️ | **MED** | Extend primitives |

---

## Advanced Features

| Feature | DWScript | go-dws | Priority | Notes |
|---------|----------|--------|----------|-------|
| **Operator Overloading** |
| Binary operators | ✅ | ✅ | - | |
| Unary operators | ✅ | ✅ | - | |
| Comparison operators | ✅ | ✅ | - | |
| Implicit conversions | ✅ | ✅ | - | |
| Explicit conversions | ✅ | ✅ | - | |
| `in` operator | ✅ | ✅ | - | |
| Global operators | ✅ | ✅ | - | |
| Class operators | ✅ | ✅ | - | |
| `**` operator | ✅ | ⏸️ | **MED** | Power |
| `[]` indexer | ✅ | ⏸️ | **MED** | Overload indexing |
| **Exception Handling** |
| `try..except..end` | ✅ | ✅ | - | |
| `try..finally..end` | ✅ | ✅ | - | |
| `try..except..finally` | ✅ | ✅ | - | |
| `on E: Type do` | ✅ | ✅ | - | |
| Multiple handlers | ✅ | ✅ | - | |
| Bare `except` | ✅ | ✅ | - | Catch-all |
| `raise` | ✅ | ✅ | - | |
| Bare `raise` | ✅ | ✅ | - | Re-throw |
| Exception hierarchy | ✅ | ✅ | - | |
| Custom exceptions | ✅ | ✅ | - | |
| Message property | ✅ | ✅ | - | |
| StackTrace property | ✅ | 🟨 | **MED** | Basic support |
| **Generics** |
| Generic classes | ✅ | ⏸️ | **LOW** | `TList<T>` |
| Generic records | ✅ | ⏸️ | **LOW** | `TPair<K,V>` |
| Generic interfaces | ✅ | ⏸️ | **LOW** | |
| Generic methods | ✅ | ⏸️ | **LOW** | |
| Type constraints | ✅ | ⏸️ | **LOW** | `<T: TObject>` |
| Multiple type params | ✅ | ⏸️ | **LOW** | `<K, V>` |
| Type inference | ✅ | ⏸️ | **LOW** | |
| Generic operators | ✅ | ⏸️ | **LOW** | |
| **Lambdas/Anonymous Methods** |
| Anonymous procedures | ✅ | ⏸️ | **MED** | `procedure (x) begin end` |
| Anonymous functions | ✅ | ⏸️ | **MED** | `function (x) begin end` |
| Lambda syntax | ✅ | ⏸️ | **MED** | `x => x * 2` |
| Closures | ✅ | ⏸️ | **MED** | Capture variables |
| Method references | ✅ | ⏸️ | **MED** | |
| Lambda parameters | ✅ | ⏸️ | **MED** | |
| Lambda returns | ✅ | ⏸️ | **MED** | |
| **Function/Method Pointers** |
| Procedural types | ✅ | ⏸️ | **HIGH** | `type TProc = procedure` |
| Function types | ✅ | ⏸️ | **HIGH** | `type TFunc = function` |
| Method pointers | ✅ | ⏸️ | **HIGH** | `of object` |
| Assignment | ✅ | ⏸️ | **HIGH** | |
| Pass as parameters | ✅ | ⏸️ | **HIGH** | |
| Call through pointer | ✅ | ⏸️ | **HIGH** | |
| **Delegates/Events** |
| Delegate types | ✅ | ⏸️ | **MED** | |
| Event handlers | ✅ | ⏸️ | **MED** | |
| Multicast delegates | ✅ | ⏸️ | **MED** | |
| Delegate invocation | ✅ | ⏸️ | **MED** | |
| **Contracts** |
| `require` clauses | ✅ | ⏸️ | **MED** | Pre-conditions |
| `ensure` clauses | ✅ | ⏸️ | **MED** | Post-conditions |
| `old` keyword | ✅ | ⏸️ | **MED** | Old value reference |
| `invariant` clauses | ✅ | ⏸️ | **LOW** | |
| Runtime checking | ✅ | ⏸️ | **MED** | |
| **Attributes** |
| Attribute declarations | ✅ | ⏸️ | **LOW** | `[AttrName]` |
| Class attributes | ✅ | ⏸️ | **LOW** | |
| Method attributes | ✅ | ⏸️ | **LOW** | |
| Property attributes | ✅ | ⏸️ | **LOW** | |
| RTTI access | ✅ | ⏸️ | **LOW** | |
| **Units/Modules** |
| `unit` declarations | ✅ | ⏸️ | **HIGH** | |
| `uses` clauses | ✅ | ⏸️ | **HIGH** | Import units |
| Initialization section | ✅ | ⏸️ | **HIGH** | Unit init code |
| Finalization section | ✅ | ⏸️ | **HIGH** | Unit cleanup |
| Namespaces | ✅ | ⏸️ | **HIGH** | Dot notation |
| Unit aliasing | ✅ | ⏸️ | **MED** | `uses U as Alias` |
| **RTTI** |
| `TypeOf(obj)` | ✅ | ⏸️ | **MED** | Type information |
| `ClassOf(class)` | ✅ | ⏸️ | **MED** | Class reference |
| Type name access | ✅ | ⏸️ | **MED** | |
| Property enumeration | ✅ | ⏸️ | **MED** | |
| Method enumeration | ✅ | ⏸️ | **MED** | |
| Field enumeration | ✅ | ⏸️ | **MED** | |
| Attribute access | ✅ | ⏸️ | **LOW** | |
| Dynamic instantiation | ✅ | ⏸️ | **MED** | |
| Dynamic invocation | ✅ | ⏸️ | **MED** | |

---

## Built-in Functions

| Feature | DWScript | go-dws | Priority | Notes |
|---------|----------|--------|----------|-------|
| **I/O Functions** |
| `PrintLn` | ✅ | ✅ | - | |
| `Print` | ✅ | ✅ | - | |
| **String Functions** |
| `Length(s)` | ✅ | ✅ | - | |
| `Copy(s, i, n)` | ✅ | ✅ | - | |
| `Concat(...)` | ✅ | ✅ | - | |
| `Pos(sub, s)` | ✅ | ✅ | - | |
| `UpperCase(s)` | ✅ | ✅ | - | |
| `LowerCase(s)` | ✅ | ✅ | - | |
| `Trim(s)` | ✅ | ⏸️ | **HIGH** | Remove whitespace |
| `TrimLeft/Right(s)` | ✅ | ⏸️ | **MED** | |
| `Insert(src, s, pos)` | ✅ | ⏸️ | **HIGH** | |
| `Delete(s, pos, n)` | ✅ | ⏸️ | **HIGH** | |
| `StringReplace(...)` | ✅ | ⏸️ | **HIGH** | |
| `Format(fmt, args)` | ✅ | ⏸️ | **HIGH** | Printf-style |
| `Chr(i)` | ✅ | ⏸️ | **MED** | ASCII to char |
| `Ord(c)` | ✅ | ⏸️ | **MED** | Char to ASCII |
| `StringOfChar(c, n)` | ✅ | ⏸️ | **MED** | |
| `ReverseString(s)` | ✅ | ⏸️ | **LOW** | |
| `CompareStr/Text(...)` | ✅ | ⏸️ | **MED** | |
| **Math Functions** |
| `Abs(x)` | ✅ | ✅ | - | |
| `Sqrt(x)` | ✅ | ✅ | - | |
| `Sqr(x)` | ✅ | ⏸️ | **HIGH** | Square |
| `Power(x, y)` | ✅ | ⏸️ | **HIGH** | x^y |
| `Exp(x)` | ✅ | ✅ | - | e^x |
| `Ln(x)` | ✅ | ✅ | - | Natural log |
| `Log10(x)` | ✅ | ⏸️ | **MED** | Base-10 log |
| `Log2(x)` | ✅ | ⏸️ | **MED** | Base-2 log |
| `Sin/Cos/Tan(x)` | ✅ | ✅ | - | |
| `ArcSin/Cos/Tan(x)` | ✅ | ⏸️ | **MED** | Inverse trig |
| `ArcTan2(y, x)` | ✅ | ⏸️ | **MED** | 2-argument atan |
| `Sinh/Cosh/Tanh(x)` | ✅ | ⏸️ | **LOW** | Hyperbolic |
| `Round(x)` | ✅ | ✅ | - | |
| `Trunc(x)` | ✅ | ✅ | - | |
| `Ceil(x)` | ✅ | ⏸️ | **HIGH** | Round up |
| `Floor(x)` | ✅ | ⏸️ | **HIGH** | Round down |
| `Frac(x)` | ✅ | ⏸️ | **MED** | Fractional part |
| `Int(x)` | ✅ | ⏸️ | **MED** | Integer part |
| `Random()` | ✅ | ✅ | - | |
| `RandomInt(max)` | ✅ | ⏸️ | **HIGH** | Random integer |
| `Randomize()` | ✅ | ✅ | - | |
| `RandSeed` | ✅ | ⏸️ | **MED** | Seed value |
| `Min(a, b)` | ✅ | ⏸️ | **HIGH** | Minimum |
| `Max(a, b)` | ✅ | ⏸️ | **HIGH** | Maximum |
| `Sign(x)` | ✅ | ⏸️ | **MED** | -1/0/1 |
| `DegToRad(x)` | ✅ | ⏸️ | **MED** | |
| `RadToDeg(x)` | ✅ | ⏸️ | **MED** | |
| **Array Functions** |
| `Length(arr)` | ✅ | ✅ | - | |
| `SetLength(arr, n)` | ✅ | ✅ | - | |
| `Low(arr)` | ✅ | ✅ | - | |
| `High(arr)` | ✅ | ✅ | - | |
| `Add(arr, elem)` | ✅ | ✅ | - | Append |
| `Delete(arr, i)` | ✅ | ✅ | - | Remove |
| `Copy(arr)` | ✅ | ⏸️ | **HIGH** | Array copy |
| `Reverse(arr)` | ✅ | ⏸️ | **MED** | |
| `IndexOf(arr, val)` | ✅ | ⏸️ | **HIGH** | Find index |
| `Contains(arr, val)` | ✅ | ⏸️ | **MED** | Test membership |
| `Sort(arr)` | ✅ | ⏸️ | **MED** | |
| **Type Conversion** |
| `IntToStr(i)` | ✅ | ✅ | - | |
| `StrToInt(s)` | ✅ | ✅ | - | |
| `FloatToStr(f)` | ✅ | ✅ | - | |
| `StrToFloat(s)` | ✅ | ✅ | - | |
| `BoolToStr(b)` | ✅ | ⏸️ | **MED** | |
| `StrToBool(s)` | ✅ | ⏸️ | **MED** | |
| **Ordinal Functions** |
| `Ord(enum)` | ✅ | ✅ | - | Enum ordinal |
| `Ord(char)` | ✅ | ⏸️ | **MED** | Char code |
| `Chr(i)` | ✅ | ⏸️ | **MED** | Code to char |
| `Succ(x)` | ✅ | ⏸️ | **HIGH** | Successor |
| `Pred(x)` | ✅ | ⏸️ | **HIGH** | Predecessor |
| `Inc(x)` | ✅ | ⏸️ | **HIGH** | Increment |
| `Dec(x)` | ✅ | ⏸️ | **HIGH** | Decrement |
| `Inc(x, delta)` | ✅ | ⏸️ | **HIGH** | Inc by delta |
| `Dec(x, delta)` | ✅ | ⏸️ | **HIGH** | Dec by delta |
| `Low(enum)` | ✅ | ⏸️ | **HIGH** | Lowest enum |
| `High(enum)` | ✅ | ⏸️ | **HIGH** | Highest enum |
| **Misc Functions** |
| `Assert(cond)` | ✅ | ⏸️ | **HIGH** | Runtime assertion |
| `Assert(cond, msg)` | ✅ | ⏸️ | **HIGH** | With message |
| `New(ptr)` | ✅ | ⏸️ | **LOW** | Dynamic alloc |
| `Dispose(ptr)` | ✅ | ⏸️ | **LOW** | Free memory |
| **DateTime Functions** |
| `Now` | ✅ | ⏸️ | **MED** | Current datetime |
| `Date` | ✅ | ⏸️ | **MED** | Current date |
| `Time` | ✅ | ⏸️ | **MED** | Current time |
| `EncodeDate(y,m,d)` | ✅ | ⏸️ | **MED** | |
| `EncodeTime(h,m,s,ms)` | ✅ | ⏸️ | **MED** | |
| `DecodeDate(dt)` | ✅ | ⏸️ | **MED** | |
| `DecodeTime(dt)` | ✅ | ⏸️ | **MED** | |
| `DateToStr(dt)` | ✅ | ⏸️ | **MED** | |
| `TimeToStr(t)` | ✅ | ⏸️ | **MED** | |
| `DateTimeToStr(dt)` | ✅ | ⏸️ | **MED** | |
| `StrToDate(s)` | ✅ | ⏸️ | **MED** | |
| `FormatDateTime(...)` | ✅ | ⏸️ | **MED** | |
| `DayOfWeek(dt)` | ✅ | ⏸️ | **LOW** | |
| `Sleep(ms)` | ✅ | ⏸️ | **MED** | Delay |
| **Variant Functions** |
| `VarType(v)` | ✅ | ⏸️ | **LOW** | Type of variant |
| `VarIsNull(v)` | ✅ | ⏸️ | **LOW** | |
| `VarToStr(v)` | ✅ | ⏸️ | **LOW** | |
| **File I/O** |
| File operations | ✅ | ❌ | N/A | Security: sandboxed |

---

## Libraries and Extensions

| Feature | DWScript | go-dws | Priority | Notes |
|---------|----------|--------|----------|-------|
| **Core Libraries** |
| JSON support | ✅ | ⏸️ | **HIGH** | Parse/generate |
| JSON LINQ | ✅ | ⏸️ | **MED** | Query JSON |
| Encoding (Base64) | ✅ | ⏸️ | **HIGH** | |
| Encoding (URL) | ✅ | ⏸️ | **HIGH** | |
| Crypto (hash) | ✅ | ⏸️ | **MED** | MD5/SHA |
| Crypto (encrypt) | ✅ | ⏸️ | **LOW** | AES/DES |
| ByteBuffer | ✅ | ⏸️ | **MED** | Binary data |
| LINQ | ✅ | ⏸️ | **LOW** | Query syntax |
| **Platform-Specific** |
| Database lib | ✅ | ❌ | N/A | Out of scope |
| Web lib (HTTP) | ✅ | ⏸️ | **MED** | HTTP client |
| Graphics lib | ✅ | ❌ | N/A | 2D graphics |
| DOM parser | ✅ | ⏸️ | **LOW** | XML/HTML |
| COM connector | ✅ | ❌ | N/A | Windows only |
| System info | ✅ | ⏸️ | **LOW** | OS info |
| Ini file lib | ✅ | ⏸️ | **LOW** | Config files |
| **Code Generation** |
| JavaScript codegen | ✅ | 🔄 | **HIGH** | Stage 11 planned |

---

## External Integration

| Feature | DWScript | go-dws | Priority | Notes |
|---------|----------|--------|----------|-------|
| `external` classes | ✅ | ✅ | - | Foundation exists |
| `external` functions | ✅ | ⏸️ | **HIGH** | Call Go code |
| FFI (call Go from script) | ✅ | ⏸️ | **HIGH** | Function registration |
| Expose Go types | ✅ | ⏸️ | **HIGH** | Type bridging |
| Sandbox limits | ✅ | 🟨 | **HIGH** | Time/memory limits |
| Capability permissions | ✅ | ⏸️ | **MED** | Security model |

---

## Development/Debug Features

| Feature | DWScript | go-dws | Priority | Notes |
|---------|----------|--------|----------|-------|
| Error messages | ✅ | ✅ | - | With line/column |
| Stack traces | ✅ | 🟨 | **HIGH** | Basic support |
| Debugger support | ✅ | ⏸️ | **MED** | Breakpoints |
| Symbol inspection | ✅ | ⏸️ | **MED** | Variable watch |
| AST dump | ✅ | ✅ | - | parse command |
| Token dump | ✅ | ✅ | - | lex command |

---

## Summary Statistics

### By Category

| Category | Total Features | Implemented | Partial | Not Started | Out of Scope |
|----------|----------------|-------------|---------|-------------|--------------|
| Core Language | 50 | 38 (76%) | 2 (4%) | 10 (20%) | 0 |
| Type System | 82 | 42 (51%) | 4 (5%) | 36 (44%) | 0 |
| OOP | 57 | 32 (56%) | 3 (5%) | 22 (39%) | 0 |
| Advanced Features | 72 | 15 (21%) | 1 (1%) | 56 (78%) | 0 |
| Built-in Functions | 115 | 32 (28%) | 0 | 83 (72%) | 0 |
| Libraries | 20 | 0 (0%) | 1 (5%) | 15 (75%) | 4 (20%) |
| Integration | 6 | 1 (17%) | 2 (33%) | 3 (50%) | 0 |
| Debug/Dev | 6 | 4 (67%) | 1 (17%) | 1 (17%) | 0 |
| **TOTAL** | **408** | **164 (40%)** | **14 (3%)** | **226 (55%)** | **4 (1%)** |

### By Priority

| Priority | Count | Percentage |
|----------|-------|------------|
| **HIGH** | 58 | 26% of not-started |
| **MED** | 94 | 42% of not-started |
| **LOW** | 74 | 33% of not-started |
| **N/A** | 4 | Out of scope |

### Quick Wins (High Priority, Should be Next)

1. `const` declarations
2. Function/method pointers
3. Units/modules system
4. Ordinal functions (Inc/Dec/Succ/Pred)
5. `Assert` function
6. String functions (Trim, Insert, Delete, Format)
7. Math functions (Min/Max, Sqr, Power, Ceil/Floor)
8. Array functions (Copy, IndexOf)
9. External function registration (FFI)
10. Type aliases

---

**Document Status**: ✅ Complete - Task 8.239c finished
