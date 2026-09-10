# Global Variables and Global Queues

**Status**: Implemented (PLAN.md §3.3)
**Fixtures**: `testdata/fixtures/FunctionsGlobalVars/` — 11 / 16 passing
**Source**: `internal/builtins/globalvars.go`, `internal/builtins/globalvars_funcs.go`

## Overview

DWScript exposes a **process-wide** store of named `Variant` values, plus a set
of named double-ended queues. Both outlive a single script run: two scripts
executed by the same host process see the same globals. This is DWScript's
`dwsGlobalVarsFunctions` unit.

Everything in this library is safe to use from several scripts at once. The
store is guarded by a single `sync.RWMutex`, and each individual function is
atomic — `IncrementGlobalVar` and `CompareExchangeGlobalVar` in particular are
designed as concurrency primitives.

### Name matching

Global and queue **names are case-sensitive**: `hello` and `Hello` are two
different globals. Wildcard **masks are case-insensitive** and support `*` (any
run of characters) and `?` (one character):

```pascal
WriteGlobalVar('hello', 'world');
WriteGlobalVar('Hello', 'world');

PrintLn(GlobalVarsNames('*').Sort(CompareStr).Join(','));  // Hello,hello
PrintLn(GlobalVarsNames('h*').Sort(CompareStr).Join(','));  // Hello,hello
```

### What can be stored

A global holds a *simple* Variant: Integer, Float, String, Boolean, `Null`, or
`Unassigned`. A `nil` interface reference degrades to `Null`. Anything else —
an object, a live interface, a record, an array — raises a catchable exception:

```pascal
try
   WriteGlobalVar('obj', Variant(someInterface));
except
   on E : Exception do PrintLn(E.Message);
   // Cannot store global of type TInterfaceSymbol
end;
```

A JSON value is flattened to its serialized text, so it reads back as a String.

## Global variables

| Function | Result | Notes |
| --- | --- | --- |
| `WriteGlobalVar(name, value [, expirationSeconds])` | — | Replaces the value *and* the expiration |
| `ReadGlobalVar(name)` | `Variant` | `Unassigned` when absent or expired |
| `ReadGlobalVarDef(name, default)` | `Variant` | `default` when absent or expired |
| `TryReadGlobalVar(name, var value)` | `Boolean` | Leaves `value` untouched when absent |
| `DeleteGlobalVar(name)` | `Boolean` | Reports whether the global existed |
| `CleanupGlobalVars([mask])` | — | Deletes everything matching; no mask means all |
| `GlobalVarsNames(mask)` | `array of String` | Sorted |
| `GlobalVarsNamesCommaText` | `String` | All names joined by commas |
| `IncrementGlobalVar(name [, delta [, expirationSeconds]])` | `Integer` | Returns the new value |
| `CompareExchangeGlobalVar(name, value, comparand)` | `Variant` | Returns the **previous** value |
| `SaveGlobalVarsToString` | `String` | Snapshot of every global |
| `LoadGlobalVarsFromString(data)` | — | Replaces every global |

```pascal
CleanupGlobalVars;

WriteGlobalVar('test', 'hello');

var v : Variant := 'def';
PrintLn(TryReadGlobalVar('test', v));   // True
PrintLn(v);                             // hello

PrintLn(ReadGlobalVarDef('missing', 'empty'));  // empty
```

### Expiration

The optional trailing argument of `WriteGlobalVar` and `IncrementGlobalVar` is
a lifetime **in seconds** (fractions allowed). An expired global reads as
absent and disappears from `GlobalVarsNames`.

The expiration is a property of the write, not of the name: every write resets
it. Incrementing *without* an expiration therefore clears a lifetime the
previous write had set.

```pascal
WriteGlobalVar('beta', 20, 0.001);   // expires in 1 ms
WriteGlobalVar('gamma', 30, 0.001);

IncrementGlobalVar('beta', 1);           // 21 — expiration cleared
IncrementGlobalVar('gamma', 1, 0.001);   // 31 — expiration renewed

Sleep(10);

PrintLn(ReadGlobalVarDef('beta', 3));    // 21
PrintLn(ReadGlobalVarDef('gamma', 4));   // 4  (expired)
```

`Sleep(milliseconds)` suspends the current script and is registered alongside
this library.

### Compare-and-exchange

`CompareExchangeGlobalVar` writes `value` only when the current value equals
`comparand`, and returns the previous value either way. A missing global
compares equal to `Unassigned`, which is how you claim a name atomically:

```pascal
if CompareExchangeGlobalVar('lock', 1, Unassigned) = Unassigned then
   PrintLn('lock acquired');
```

Comparison follows Variant rules: Integer and Float compare numerically,
everything else compares by type and payload.

### Save and restore

`SaveGlobalVarsToString` produces an opaque snapshot; `LoadGlobalVarsFromString`
replaces the *entire* set of globals with it. An empty string clears the store.
A payload that is not a snapshot raises `Invalid file tag`.

```pascal
var backup := SaveGlobalVarsToString;
WriteGlobalVar('scratch', 42);
LoadGlobalVarsFromString(backup);   // 'scratch' is gone again
```

The go-dws snapshot format is a `DWSGV1` header line followed by a JSON array of
records. It is **not** byte-compatible with Delphi DWScript's binary format;
treat a snapshot as opaque and do not persist it across go-dws versions.
Queues are not part of a snapshot.

## Global queues

A global queue is a double-ended queue of Variants. Push and Insert add;
Pull, Pop, First and Peek read from a specific end.

| Function | Result | End |
| --- | --- | --- |
| `GlobalQueuePush(name, value)` | `Boolean` | appends to the **back** |
| `GlobalQueueInsert(name, value)` | `Boolean` | prepends to the **front** |
| `GlobalQueuePull(name, var value)` | `Boolean` | removes the **front** |
| `GlobalQueuePop(name, var value)` | `Boolean` | removes the **back** |
| `GlobalQueueFirst(name, var value)` | `Boolean` | reads the **front** |
| `GlobalQueuePeek(name, var value)` | `Boolean` | reads the **back** |
| `GlobalQueueLength(name)` | `Integer` | |
| `GlobalQueueSnapshot(name)` | `array of Variant` | front first, queue untouched |
| `GlobalQueueSnapshotIntegers(name)` | `array of Integer` | |
| `GlobalQueueSnapshotFloats(name)` | `array of Float` | |
| `GlobalQueueSnapshotStrings(name)` | `array of String` | |
| `CleanupGlobalQueues([mask])` | — | |

The four reading functions return `False` and **leave the variable unchanged**
when the queue is empty.

```pascal
CleanupGlobalQueues;

GlobalQueuePush('test', 1);
GlobalQueuePush('test', 2);
GlobalQueueInsert('test', 3);    // queue is now 3, 1, 2

var v : Variant;
while GlobalQueuePop('test', v) do
   PrintLn(v);                   // 2, then 1, then 3
```

Using Push with Pull gives you a FIFO; Push with Pop gives you a LIFO.

### Typed snapshots

The typed snapshot functions convert each entry. Floats round to the nearest
integer with banker's rounding (`1.75` and `2.25` both become `2`), and numbers
render to strings the way `Print` renders them. A String entry cannot become an
Integer or Float and raises a catchable exception:

```pascal
try
   aInt := GlobalQueueSnapshotIntegers('str');
except
   on E : Exception do PrintLn(E.Message);
   // Could not cast variant from String to Integer
end;
```

## Embedding notes

`builtins.DefaultGlobalVars` is the process-wide `*builtins.GlobalVarStore`
that every script shares. A host that wants isolation — or a test that wants a
clean slate — constructs its own with `builtins.NewGlobalVarStore()`.

`GlobalVarStore.SetClock(func() time.Time)` replaces the store's time source,
which is how the expiration tests stay deterministic.

## Known gaps

- **Private (per-unit) variables.** `WritePrivateVar`, `ReadPrivateVar`,
  `PrivateVarsNames` and `CleanupPrivateVars` are not implemented. The fixture
  that covers them (`private_vars`) also depends on a separate parser gap:
  a unit without `interface`/`implementation` sections fails to parse.
