# Helpers

**Status:** ✅ Fully Implemented (Tasks 9.74-9.92)

Helpers are a powerful DWScript feature that allows you to extend existing types (basic types, records, and classes) with additional methods and properties without modifying the original type declarations. This is similar to extension methods in C# or Swift.

## Table of Contents

- [Overview](#overview)
- [Syntax](#syntax)
- [Features](#features)
- [Examples](#examples)
- [Implementation Details](#implementation-details)
- [Limitations](#limitations)

## Overview

Helpers enable you to:
- Add methods to built-in types (String, Integer, Float, Boolean)
- Extend records with additional functionality
- Enhance classes without modifying their definitions
- Define class constants and variables within helpers
- Add properties to existing types

Multiple helpers can extend the same type, and all their methods will be available.

## Syntax

There are two syntax variants for declaring helpers:

### Basic Helper Syntax

```pascal
type THelperName = helper for TargetType
  // Methods, properties, class vars, class consts
end;
```

### Record Helper Syntax

```pascal
type THelperName = record helper for TargetType
  // Methods, properties, class vars, class consts
end;
```

Both syntax variants are functionally equivalent. The `record helper` syntax is provided for compatibility with the original DWScript.

## Features

### 1. Methods

Helpers can define methods that operate on the target type. The `Self` keyword refers to the instance of the target type.

```pascal
type TIntegerHelper = helper for Integer
  function IsEven: Boolean;
  begin
    Result := (Self mod 2 = 0);
  end;

  function Square: Integer;
  begin
    Result := Self * Self;
  end;
end;

var n: Integer;
begin
  n := 42;
  PrintLn(n.IsEven());    // True
  PrintLn(n.Square());    // 1764
end.
```

### 2. Properties

Helpers can define properties with getter (and optionally setter) methods.

```pascal
type TPointHelper = record helper for TPoint
  property Sum: Integer read GetSum;

  function GetSum: Integer;
  begin
    Result := Self.X + Self.Y;
  end;
end;

var p: TPoint;
begin
  p.X := 3;
  p.Y := 4;
  PrintLn(p.Sum);  // 7
end.
```

### 3. Class Constants

Helpers can define class-level constants that are accessible within helper methods.

```pascal
type TMathHelper = helper for Float
  class const PI = 3.14159;
  class const E = 2.71828;

  function TimesPi: Float;
  begin
    Result := Self * PI;
  end;
end;

var radius: Float;
begin
  radius := 2.0;
  PrintLn(radius.TimesPi());  // 6.28318
end.
```

### 4. Class Variables

Helpers can define class-level variables that are shared across all instances.

```pascal
type TIntegerHelper = helper for Integer
  class var DefaultValue: Integer;

  function GetDefault: Integer;
  begin
    Result := DefaultValue;
  end;
end;
```

**Note:** Class variables are initialized to their default values (0 for Integer, empty string for String, etc.).

### 5. Visibility Sections

Helpers support `private` and `public` visibility sections:

```pascal
type TStringHelper = helper for String
  private
    function InternalHelper: String;
    begin
      Result := 'internal';
    end;
  public
    function PublicMethod: String;
    begin
      Result := InternalHelper();
    end;
end;
```

### 6. Multiple Helpers

Multiple helpers can extend the same type, and all their methods will be available:

```pascal
type TIntHelper1 = helper for Integer
  function Double: Integer;
  begin
    Result := Self * 2;
  end;
end;

type TIntHelper2 = helper for Integer
  function Triple: Integer;
  begin
    Result := Self * 3;
  end;
end;

var n: Integer;
begin
  n := 5;
  PrintLn(n.Double());  // 10
  PrintLn(n.Triple());  // 15
end.
```

## Examples

### String Helper

```pascal
type TStringHelper = helper for String
  function IsEmpty: Boolean;
  begin
    Result := (Self = '');
  end;

  function ToUpper: String;
  begin
    // Simplified uppercase implementation
    Result := 'UPPER:' + Self;
  end;
end;

var s: String;
begin
  s := 'hello';
  PrintLn(s.ToUpper());    // UPPER:hello
  PrintLn(s.IsEmpty());    // False
end.
```

### Record Helper

```pascal
type TPoint = record
  X, Y: Integer;
end;

type TPointHelper = record helper for TPoint
  function DistanceFromOrigin: Float;
  begin
    Result := Sqrt(Self.X * Self.X + Self.Y * Self.Y);
  end;

  function IsOrigin: Boolean;
  begin
    Result := (Self.X = 0) and (Self.Y = 0);
  end;
end;

var p: TPoint;
begin
  p.X := 3;
  p.Y := 4;
  PrintLn(p.DistanceFromOrigin());  // 5.0
  PrintLn(p.IsOrigin());             // False
end.
```

### Class Helper

```pascal
type TPerson = class
  private
    FName: String;
    FAge: Integer;
  public
    constructor Create(name: String; age: Integer);
    function GetAge: Integer;
end;

// ... TPerson implementation ...

type TPersonHelper = helper for TPerson
  function IsAdult: Boolean;
  begin
    Result := Self.GetAge() >= 18;
  end;

  function IsTeenager: Boolean;
  begin
    Result := (Self.GetAge() >= 13) and (Self.GetAge() <= 19);
  end;
end;

var person: TPerson;
begin
  person := TPerson.Create('Alice', 25);
  PrintLn(person.IsAdult());      // True
  PrintLn(person.IsTeenager());   // False
end.
```

## Implementation Details

### Method Resolution Priority

When calling a method on an object, the resolution order is:

1. **Instance methods** (from the class/record definition)
2. **Helper methods** (from all applicable helpers)

This means that instance methods always take precedence over helper methods. If a helper defines a method with the same name as an instance method, the instance method will be called.

### Self Binding

Within helper methods:
- `Self` refers to the instance of the target type
- For basic types (Integer, String, etc.), `Self` is the value itself
- For records, `Self` is the record instance
- For classes, `Self` is the object instance

### Type Resolution

Helpers are registered at runtime in the interpreter and at compile-time in the semantic analyzer. When a method is called on a value, the interpreter:

1. Checks if it's an instance method (for classes/records)
2. If not found, searches all helpers registered for that type
3. Executes the first matching helper method found

### Storage

- **Runtime:** Helpers are stored in a map indexed by target type name: `map[string][]*HelperInfo`
- **Compile-time:** Helpers are tracked in the semantic analyzer: `map[string][]*types.HelperType`

## Limitations

### Current Limitations

1. **No Operator Overloading:** Helpers cannot override operators (use regular operator overloading instead)
2. **No Field Access:** Helpers cannot directly access private fields of classes (must use public methods/properties)
3. **No External Methods:** Helper methods must be defined inline within the helper declaration block
4. **Static Resolution:** Helper methods are resolved at runtime, not compile-time (slight performance overhead)

### Future Enhancements

Potential future improvements:
- Indexed properties in helpers
- Helper inheritance/composition
- Conditional helper compilation
- Performance optimizations for helper method dispatch

## Testing

Comprehensive test suites are available:

- **Unit Tests:** `internal/interp/helpers_test.go` (12 test functions)
- **Integration Tests:** `cmd/dwscript/helpers_integration_test.go`
- **Test Scripts:** `testdata/helpers/*.dws`

Run tests with:
```bash
go test ./internal/interp -run TestHelper
go test ./cmd/dwscript -run TestHelpers
```

## References

- Original DWScript Documentation: https://www.delphitools.info/2012/05/02/helpers-added-to-dwscript/
- Implementation Tasks: PLAN.md (Tasks 9.74-9.92)
- AST Definition: `internal/ast/helper.go`
- Parser: `internal/parser/helpers.go`
- Semantic Analyzer: `internal/semantic/analyze_helpers.go`
- Interpreter: `internal/interp/helpers.go`

## See Also

- [Classes](classes.md) - Object-oriented programming
- [Records](records.md) - Value types
- [Properties](properties.md) - Property declarations
