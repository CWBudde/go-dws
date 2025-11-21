# ArrayValue Migration Analysis

> Task 1.3.1 - Dependency Analysis for ArrayValue
>
> Date: 2025-11-21

## Executive Summary

**Status: MIGRATION ALREADY COMPLETE**

The ArrayValue migration to the runtime package was completed in **Phase 3.5.4**. The TODOs in `internal/interp/builtins/strings_basic.go` are outdated - the builtins package already has full access to `runtime.ArrayValue` and can implement the string functions that require it.

## Current State

### ArrayValue Location

**Definition**: `internal/interp/runtime/array.go`

```go
type ArrayValue struct {
    ArrayType *types.ArrayType
    Elements  []Value
}
```

**Type Alias**: `internal/interp/value.go:83`

```go
ArrayValue = runtime.ArrayValue
```

### Builtins Package Access

The `internal/interp/builtins/` package already successfully uses `runtime.ArrayValue` in several files:

- `collections.go` - Map, Filter, Reduce, ForEach functions
- `array.go` - Array manipulation functions
- `array_test.go` - Comprehensive array tests

Example from `collections.go`:
```go
arrayVal, ok := args[0].(*runtime.ArrayValue)
// ...
return &runtime.ArrayValue{
    Elements:  resultElements,
    ArrayType: arrayVal.ArrayType,
}
```

## Package Import Graph

```
┌────────────────────────────────────────────────────────────────┐
│                    internal/interp/builtins/                    │
│                                                                │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────────────┐  │
│  │ collections │   │    array    │   │   strings_basic     │  │
│  │     .go     │   │     .go     │   │        .go          │  │
│  └──────┬──────┘   └──────┬──────┘   └──────────┬──────────┘  │
│         │                 │                     │              │
│         └────────────┬────┴─────────────────────┘              │
│                      │                                         │
│                      ▼                                         │
│         ┌────────────────────────┐                            │
│         │  runtime.ArrayValue    │                            │
│         │  runtime.StringValue   │                            │
│         │  runtime.IntegerValue  │                            │
│         │  etc.                  │                            │
│         └────────────┬───────────┘                            │
│                      │                                         │
└──────────────────────┼────────────────────────────────────────┘
                       │
                       ▼
           ┌────────────────────────┐
           │ internal/interp/runtime │
           │                        │
           │  - array.go            │
           │  - primitives.go       │
           │  - set.go              │
           │  - enum.go             │
           └────────────────────────┘
```

## Functions Ready for Migration

The following functions can now be migrated to `internal/interp/builtins/strings_basic.go`:

| Function | Current Location | Status | Notes |
|----------|-----------------|--------|-------|
| Format | `internal/interp/builtins_strings_basic.go` | Ready | Uses ArrayValue for args |
| StrSplit | `internal/interp/builtins_strings_advanced.go` | Ready | Returns ArrayValue |
| StrJoin | `internal/interp/builtins_strings_advanced.go` | Ready | Takes ArrayValue as input |
| StrArrayPack | `internal/interp/builtins_strings_advanced.go` | Ready | Takes/returns ArrayValue |

## Migration Plan (Task 1.3.2)

### Step 1: Remove Outdated TODO Comments

Remove the following outdated TODOs from `internal/interp/builtins/strings_basic.go`:
- Line 393: `// TODO: Format - Requires ArrayValue to be moved to runtime package`
- Line 937: `// TODO: StrSplit - Requires ArrayValue to be moved to runtime package`
- Line 941: `// TODO: StrJoin - Requires ArrayValue to be moved to runtime package`
- Line 945: `// TODO: StrArrayPack - Requires ArrayValue to be moved to runtime package`

### Step 2: Implement Functions in Builtins Package

Port the functions from their current locations to `internal/interp/builtins/strings_basic.go`:

1. **Format()** - Requires adapting to use Context interface
2. **StrSplit()** - Straightforward port, use `runtime.ArrayValue`
3. **StrJoin()** - Straightforward port
4. **StrArrayPack()** - Straightforward port

### Step 3: Register Functions

Add to `internal/interp/builtins/register.go`:
```go
"Format":       Format,
"StrSplit":     StrSplit,
"StrJoin":      StrJoin,
"StrArrayPack": StrArrayPack,
```

### Step 4: Update Tests

Add tests in `internal/interp/builtins/strings_basic_test.go` for the new functions.

## Acceptance Criteria

- [x] ArrayValue successfully in runtime package (Phase 3.5.4)
- [x] Builtins package can access runtime.ArrayValue
- [x] No circular import errors
- [ ] Outdated TODO comments removed
- [ ] Format, StrSplit, StrJoin, StrArrayPack implemented in builtins
- [ ] All tests pass
- [ ] Code still compiles and runs correctly

## Dependencies Resolved

The original circular dependency concern has been addressed through:

1. **Runtime Package Design**: The runtime package contains only value types with no dependencies on the interpreter
2. **Context Interface**: The builtins package uses the Context interface to interact with the interpreter without direct imports
3. **Type Aliases**: The interp package provides aliases (`ArrayValue = runtime.ArrayValue`) for backward compatibility

## Files Affected

- `internal/interp/builtins/strings_basic.go` - Add new functions
- `internal/interp/builtins/register.go` - Register new functions
- `internal/interp/builtins/strings_basic_test.go` - Add tests
- `TODOs.md` - Update task status

## Notes

The migration was completed earlier than the TODOs suggested. The development team should periodically review TODO comments to identify those that have been resolved but not updated.
