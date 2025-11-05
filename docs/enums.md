# Enumerated Types in go-dws

Enumerated types (enums) allow you to define a type with a fixed set of named constant values. This document describes enum support in go-dws.

## Table of Contents
- [Basic Syntax](#basic-syntax)
- [Ordinal Values](#ordinal-values)
- [Scoped References](#scoped-references)
- [Built-in Functions](#built-in-functions)
- [Case Statements](#case-statements)
- [Implementation Status](#implementation-status)

## Basic Syntax

Define an enum using the `type` keyword with parentheses containing the enum values:

```pascal
type TColor = (Red, Green, Blue);
```

Declare variables of enum type:

```pascal
var color: TColor;
color := Red;
```

## Ordinal Values

### Implicit Values

By default, enum values start at 0 and increment by 1:

```pascal
type TColor = (Red, Green, Blue);
// Red = 0, Green = 1, Blue = 2
```

### Explicit Values

You can assign explicit ordinal values to enum elements:

```pascal
type TStatus = (Ok = 0, Warning = 1, Error = 2);
```

### Mixed Values

You can mix implicit and explicit values. Implicit values continue from the last explicit value:

```pascal
type TPriority = (Low, Medium = 5, High);
// Low = 0, Medium = 5, High = 6
```

## Scoped References

Enum values can be referenced in two ways:

### Unscoped Reference

```pascal
type TColor = (Red, Green, Blue);
var color: TColor := Red;
```

### Scoped Reference

Use the enum type name as a qualifier for clarity or to avoid name conflicts:

```pascal
type TColor = (Red, Green, Blue);
var color: TColor := TColor.Red;
```

Both forms are equivalent and can be used interchangeably.

## Built-in Functions

### Ord()

Returns the ordinal value of an enum:

```pascal
type TWeekdays = (Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday);

PrintLn(Ord(Monday));    // Output: 0
PrintLn(Ord(Sunday));    // Output: 6
```

### Integer()

Converts an enum to its integer value (same as Ord for enums):

```pascal
type TStatus = (Ok = 100, Warning = 200);

PrintLn(Integer(Ok));       // Output: 100
PrintLn(Integer(Warning));  // Output: 200
```

## Case Statements

Enums work naturally with case statements:

```pascal
type TWeekdays = (Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday);

var day: TWeekdays := Wednesday;

case day of
  Monday:    PrintLn('Start of work week');
  Tuesday:   PrintLn('Tuesday');
  Wednesday: PrintLn('Midweek');
  Thursday:  PrintLn('Thursday');
  Friday:    PrintLn('End of work week');
  Saturday, Sunday: PrintLn('Weekend');
end;
```

You can use scoped references in case branches:

```pascal
case day of
  TWeekdays.Monday: PrintLn('Monday');
  TWeekdays.Sunday: PrintLn('Sunday');
end;
```

## Implementation Status

### ✅ Completed Features

- Enum type declarations
- Implicit ordinal values (starting at 0)
- Explicit ordinal values
- Mixed implicit/explicit values
- Variable declarations with enum types
- Unscoped enum value references
- Scoped enum value references (TEnum.Value)
- Ord() built-in function
- Integer() cast function
- Case statement support with enums
- Semantic validation (duplicate names/values)
- `.Value` property for enum values (returns ordinal value)
- `.Name` property for enum values (returns value name as string)
- `.QualifiedName` property (returns TypeName.ValueName)
- `High()` and `Low()` functions for enums
- For-in loops with enums

### ⏳ Planned Features

- Enum comparison operators
- Enum-based sets (set of TColor)
- Enum type casting (TEnum(ordinal))
- Scoped enum member access through type (TEnum.Value)

## Examples

### Basic Usage

```pascal
type TColor = (Red, Green, Blue);

var favoriteColor: TColor;
favoriteColor := Green;

PrintLn(favoriteColor);        // Output: Green
PrintLn(Ord(favoriteColor));   // Output: 1
```

### HTTP Status Codes

```pascal
type THttpStatus = (
  Ok = 200,
  Created = 201,
  BadRequest = 400,
  Unauthorized = 401,
  NotFound = 404,
  InternalServerError = 500
);

var status: THttpStatus := NotFound;
PrintLn(Integer(status));  // Output: 404
```

### Days of Week with Case

```pascal
type TWeekday = (Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday);

procedure PrintDayType(day: TWeekday);
begin
  case day of
    Monday, Tuesday, Wednesday, Thursday, Friday:
      PrintLn('Workday');
    Saturday, Sunday:
      PrintLn('Weekend');
  end;
end;

PrintDayType(Friday);     // Output: Workday
PrintDayType(Sunday);     // Output: Weekend
```

## Testing

Enum functionality is thoroughly tested:

- **Unit tests**: `interp/enum_test.go`
- **Semantic tests**: `semantic/enum_test.go`
- **Integration tests**: `testdata/enums/*.dws`

Run tests:

```bash
# Unit tests
go test ./interp -run TestEnum -v
go test ./semantic -run TestEnum -v

# Integration tests
./bin/dwscript run testdata/enums/basic_enum.dws
./bin/dwscript run testdata/enums/enum_ord.dws
./bin/dwscript run testdata/enums/enum_case.dws
```

## References

- DWScript enum documentation: https://www.delphitools.info/dwscript/
- DWScript reference tests: `reference/dwscript-original/Test/SimpleScripts/enum*.pas`
- Implementation tasks: PLAN.md (Stage 8, Tasks 8.30-8.52)
