# Records and record type values

A record annotation declares an instance with its own fields. Using the record
name as a value produces a record type value, whose inferred type is
`meta of <record name>`:

```pascal
type TCounter = record
  Value: Integer;
  class var Count: Integer;
  class procedure Add(amount: Integer);
  begin
    Count += amount;
  end;
end;
type TAlias = TCounter;

var item: TAlias;             // Record instance
var counter := TCounter;     // Record type value
var aliasCounter := TAlias;  // Same record type, accessed through its alias
var copiedCounter := counter;

counter.Count := 1;
aliasCounter.Add(2);
PrintLn(copiedCounter.Count); // 3
```

Aliases and copied type values share the record's class variables and class
methods. Record instances remain values: copying an instance copies its fields.
A record type value cannot be assigned a record instance, and distinct record
declarations have distinct metatypes even when their fields are identical.

Record type values expose class variables, class constants, class methods, and
properties with class-side accessors. Instance fields and instance methods
require a record instance. Named and default indexed properties retain their
read/write behavior through inferred type values, including compound assignment.

## Method calls and defaults

Native instance and class methods accept omitted default parameters, including
calls by bare method name inside another record method:

```pascal
type TPrinter = record
  procedure WriteValue(value: Integer = 7);
  begin
    PrintLn(value);
  end;
  class procedure WriteStatic(value: Integer = 8);
  begin
    PrintLn(value);
  end;
  procedure Run;
  begin
    WriteValue();
    WriteStatic();
  end;
end;
var printer: TPrinter;
printer.Run();       // 7, then 8
TPrinter.WriteStatic(); // 8
```

Defaults use the record declaration's constant scope, so a caller-local variable
with the same name cannot replace a constant default. Out-of-line implementations
retain defaults from their declarations.

Declare `overload` to select among same-named signatures by argument types and
available defaults. An instance context can select either an instance or class
method from the same set; a class context selects class methods. Parenthesized
implicit calls use those signatures too.
Count and overload diagnostics identify the written method name. Instance
method type diagnostics count the implicit receiver as argument 0; see
[Error messages](error-messages.md) for their positions.

## Indexed property declarations

Named, inline and anonymous records accept grouped value, `var` and `const` index
parameters. Each modifier applies to its comma-separated names and resets after
the semicolon:

```pascal
type TLookup = record
  function Get(var key, offset: Integer; const limit: Integer): Integer;
  begin
    Result := key + offset + limit;
  end;
  property Item[var a, b: Integer; const c: Integer]: Integer read Get;
end;
```

Accessor parameter names may differ from the property's names. Types, passing
modes and default presence must agree; declaration errors check them in that
order. Getter result checks come first, and setter Value modes/defaults are
outside index equality. Forwarded properties validate the selected underlying
accessor while retaining their own index signature. Record property printing
preserves class/default flags and index groups.

This declaration support does not complete record index execution. Caller
references, expression accessors, forwarding and receiver capture remain open
in [PLAN.md](../../PLAN.md). Missing composite array element types and completed
property type failures in complete-inline/anonymous records also have separate
recovery follow-ups; they are not covered by declaration acceptance.

## Helpers

A nonstatic class method in a record helper receives a record type value as
`Self`. Its instance methods receive a record instance; static methods have no
`Self`. Ordinary record class methods also have no `Self` and can access the
record's class members by name.

Explicit helper calls accept an inferred record type value as their first
argument:

```pascal
type TCounterHelper = helper for TCounter
  class function ReadCount: Integer;
  begin
    Result := Self.Count;
  end;
end;

PrintLn(counter.ReadCount);
PrintLn(TCounterHelper.ReadCount(counter));
```

A helper declared for an alias retains that alias's helper selection when its
type value is inferred or copied. See [Helpers](helpers.md) for helper syntax and
explicit calls.
