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
