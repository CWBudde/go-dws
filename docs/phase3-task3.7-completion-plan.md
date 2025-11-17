# Phase 3.7 - Built-in Function Migration Completion Plan

## Overview

Phase 3.7.2 successfully established the built-in function registry infrastructure and migrated 169/244 functions (69.3%). This document outlines the roadmap for completing the migration of the remaining 74 functions through tasks 3.7.3 through 3.7.9.

## Current Status (After 3.7.2)

### ✅ Completed
- **Registry Infrastructure**: Thread-safe, case-insensitive, categorized
- **Functions Migrated**: 169/244 (69.3%)
- **Categories**: 4 (Math, String, DateTime, Conversion)
- **Code Reduction**: 336 lines removed (50% in dispatcher)
- **Performance**: O(1) lookup for all registered functions

### 🔄 Remaining
- **Functions to Migrate**: 74/244 (30.7%)
- **New Categories Needed**: 6 (IO, Array, Collections, Ordinal, Variant, JSON, Type, System)
- **Context Extensions**: Type helpers, I/O access, array manipulation, callbacks

## Migration Strategy

The remaining 74 functions are blocked by architectural constraints that require incremental Context interface extensions. We'll migrate in order of complexity:

```
Wave 1 (3.7.3-3.7.4): Type Helpers + I/O       →  17 functions  →  186 total (76%)
Wave 2 (3.7.5-3.7.6): Ordinals + Variants + JSON → 26 functions  →  212 total (87%)
Wave 3 (3.7.7-3.7.8): Arrays + Collections + Misc → 32 functions  →  244 total (100%)
Wave 4 (3.7.9):      Enhancement & Optimization
```

## Task Breakdown

### Task 3.7.3: Extend Context for Type-Dependent Functions
**Estimated**: 3 days | **Functions**: 15 | **Total After**: 184 (75%)

**Objective**: Add type conversion helpers to Context to support functions that work with SubrangeValue, EnumValue, and other interpreter-specific types.

**Context Extensions**:
```go
type Context interface {
    // ... existing methods ...

    // Type conversion helpers
    ToInt64(Value) (int64, error)      // Handles IntegerValue, SubrangeValue, EnumValue
    ToFloat64(Value) (float64, error)  // Handles FloatValue, IntegerValue
    ToBool(Value) (bool, error)        // Handles BooleanValue, any truthy value
    ToString(Value) string              // Handles StringValue, converts others
}
```

**Functions to Migrate**:

*Conversion (8 functions)*:
- `IntToStr` - Convert integer to string with optional base
- `IntToBin` - Convert integer to binary with width
- `StrToInt` - Parse string to integer
- `StrToFloat` - Parse string to float
- `FloatToStr` - Convert float to string
- `BoolToStr` - Convert boolean to string
- `StrToIntDef` - Parse string to integer with default
- `StrToFloatDef` - Parse string to float with default

*Encoding (5 functions)*:
- `StrToHtml` - Escape string for HTML
- `StrToHtmlAttribute` - Escape string for HTML attribute
- `StrToJSON` - Escape string for JSON
- `StrToCSSText` - Escape string for CSS
- `StrToXML` - Escape string for XML

*Other (2 functions)*:
- `Format` - Advanced string formatting
- `Assigned` - Check if value is assigned

**New Files**:
- `internal/interp/builtins/conversion.go` - Conversion functions
- `internal/interp/builtins/encoding.go` - Encoding functions

**New Categories**:
- `CategoryEncoding` - HTML, JSON, XML, CSS encoding

**Acceptance Criteria**:
- [ ] Context extended with type helper methods
- [ ] Interpreter implements all type helpers
- [ ] 15+ functions migrated and registered
- [ ] Registry contains 184+ functions
- [ ] All tests passing

---

### Task 3.7.4: Add I/O Support and Migrate Print Functions
**Estimated**: 1 day | **Functions**: 2 | **Total After**: 186 (76%)

**Objective**: Extend Context with output writer access and migrate Print/PrintLn functions.

**Context Extensions**:
```go
type Context interface {
    // ... existing methods ...

    // I/O access
    GetOutput() io.Writer  // Returns configured output writer
}
```

**Functions to Migrate**:
- `Print` - Print values without newline
- `PrintLn` - Print values with newline

**New Files**:
- `internal/interp/builtins/io.go` - I/O functions

**New Categories**:
- `CategoryIO` - Input/output operations

**Acceptance Criteria**:
- [ ] Context extended with GetOutput()
- [ ] Print/PrintLn migrated and registered
- [ ] Output correctly written to configured writer
- [ ] Registry contains 186+ functions
- [ ] All tests passing

---

### Task 3.7.5: Migrate Ordinal and Variant Functions
**Estimated**: 4 days | **Functions**: 17 | **Total After**: 203 (83%)

**Objective**: Add helpers for enum/subrange navigation and variant type handling.

**Context Extensions**:
```go
type Context interface {
    // ... existing methods ...

    // Ordinal helpers
    IsEnum(Value) bool
    IsSubrange(Value) bool
    EnumSucc(Value) (Value, error)    // Next enum value
    EnumPred(Value) (Value, error)    // Previous enum value
    OrdinalValue(Value) (int64, error) // Get ordinal position

    // Variant helpers
    IsVariant(Value) bool
    VariantType(Value) int             // Returns varXxx constant
    VariantToValue(Value) Value        // Unwraps variant
    ValueToVariant(Value) Value        // Wraps in variant
}
```

**Functions to Migrate**:

*Ordinals (6 functions)*:
- `Ord` - Get ordinal value of enum/char/bool
- `Integer` - Cast to integer
- `Inc` - Increment ordinal (var param)
- `Dec` - Decrement ordinal (var param)
- `Succ` - Successor of ordinal
- `Pred` - Predecessor of ordinal

*Variant Type Checking (7 functions)*:
- `VarType` - Returns variant type enum
- `VarIsNull` - Check if variant is null
- `VarIsEmpty` - Check if variant is empty
- `VarIsClear` - Check if variant is clear
- `VarIsArray` - Check if variant is array
- `VarIsStr` - Check if variant is string
- `VarIsNumeric` - Check if variant is numeric

*Variant Conversion (4 functions)*:
- `VarToStr` - Convert variant to string
- `VarToInt` - Convert variant to integer
- `VarToFloat` - Convert variant to float
- `VarAsType` - Convert variant to type
- `VarClear` - Clear variant value

**New Files**:
- `internal/interp/builtins/ordinals.go` - Ordinal functions
- `internal/interp/builtins/variant.go` - Variant functions

**New Categories**:
- `CategoryOrdinal` - Ordinal operations (Inc, Dec, Succ, Pred)
- `CategoryVariant` - Variant type operations

**Acceptance Criteria**:
- [ ] Context extended with ordinal/variant helpers
- [ ] 17+ functions migrated and registered
- [ ] Inc/Dec support var parameters correctly
- [ ] Registry contains 203+ functions
- [ ] All tests passing

---

### Task 3.7.6: Migrate JSON and Type Introspection Functions
**Estimated**: 3 days | **Functions**: 9 | **Total After**: 212 (87%)

**Objective**: Add JSON parsing/serialization and type introspection support.

**Context Extensions**:
```go
type Context interface {
    // ... existing methods ...

    // JSON helpers
    ParseJSONValue(string) (Value, error)  // Parse JSON to internal value
    ValueToJSON(Value) (string, error)     // Serialize value to JSON
    ValueToJSONFormatted(Value, indent string) (string, error)

    // Type introspection helpers
    GetTypeOf(Value) string                // Returns type name
    GetClassOf(Value) (Value, error)       // Returns class type info
}
```

**Functions to Migrate**:

*JSON (7 functions)*:
- `ParseJSON` - Parse JSON string to value
- `ToJSON` - Convert value to JSON string
- `ToJSONFormatted` - Convert value to formatted JSON
- `JSONHasField` - Check if JSON has field
- `JSONKeys` - Get JSON object keys
- `JSONValues` - Get JSON object values
- `JSONLength` - Get JSON array/object length

*Type Introspection (2 functions)*:
- `TypeOf` - Get type name of value
- `TypeOfClass` - Get class type of value

**New Files**:
- `internal/interp/builtins/json.go` - JSON functions
- `internal/interp/builtins/type.go` - Type introspection

**New Categories**:
- `CategoryJSON` - JSON parsing and serialization
- `CategoryType` - Type introspection and reflection

**Acceptance Criteria**:
- [ ] Context extended with JSON/type helpers
- [ ] 9+ functions migrated and registered
- [ ] JSON round-trip works correctly
- [ ] Registry contains 212+ functions
- [ ] All tests passing

---

### Task 3.7.7: Migrate Array and Collection Functions
**Estimated**: 5 days | **Functions**: 21 | **Total After**: 233 (95%)

**Objective**: Add array manipulation and functional programming (map/filter/reduce) support.

**Context Extensions**:
```go
type Context interface {
    // ... existing methods ...

    // Array helpers
    GetArrayLength(Value) (int, error)
    SetArrayLength(arrayPtr Value, length int) error
    ArrayCopy(Value) (Value, error)
    ArrayReverse(Value) error
    ArraySort(Value) error
    ArrayIndexOf(Value, element Value, start int) (int, error)
    ArrayContains(Value, element Value) (bool, error)

    // Collection helpers (functional programming)
    EvalFunctionPointer(fp Value, args []Value) (Value, error)
}
```

**Functions to Migrate**:

*Array Basics (8 functions)*:
- `Length` - Get length of array/string
- `Copy` - Copy array
- `Low` - Get lower bound of array/subrange
- `High` - Get upper bound of array/subrange
- `IndexOf` - Find index of element
- `Contains` - Check if array contains element
- `Reverse` - Reverse array in place
- `Sort` - Sort array in place

*Array Manipulation (3 functions)*:
- `Add` - Add element to dynamic array
- `Delete` - Delete element from array
- `SetLength` - Set dynamic array length (var param)

*Collections - Functional (10 functions)*:
- `Map` - Transform array elements
- `Filter` - Filter array elements
- `Reduce` - Reduce array to single value
- `ForEach` - Execute function for each element
- `Every` - Check if all elements match predicate
- `Some` - Check if any element matches predicate
- `Find` - Find first matching element
- `FindIndex` - Find index of first match
- `ConcatArrays` - Concatenate multiple arrays
- `Slice` - Extract array slice

**New Files**:
- `internal/interp/builtins/array.go` - Array functions
- `internal/interp/builtins/collections.go` - Collection functions

**New Categories**:
- `CategoryArray` - Array operations
- `CategoryCollections` - Functional programming on collections

**Acceptance Criteria**:
- [ ] Context extended with array/callback helpers
- [ ] 21+ functions migrated and registered
- [ ] Map/Filter/Reduce work with function pointers
- [ ] Registry contains 233+ functions
- [ ] All tests passing

---

### Task 3.7.8: Migrate Remaining Miscellaneous Functions
**Estimated**: 2 days | **Functions**: 11 | **Total After**: 244 (100%)

**Objective**: Migrate final utility functions and remove switch statement entirely.

**Context Extensions**:
```go
type Context interface {
    // ... existing methods ...

    // System helpers
    GetCallStack() []string
    Terminate(message string)         // For Assert
    SwapValues(a, b *Value) error    // For Swap
}
```

**Functions to Migrate**:

*System/Runtime (5 functions)*:
- `GetStackTrace` - Get formatted stack trace
- `GetCallStack` - Get call stack frames
- `Assert` - Runtime assertion
- `Swap` - Swap two values (var params)
- `DivMod` - Integer division with modulo (var params)

*Other (6+ functions)*:
- Any remaining functions not yet categorized

**Switch Statement Removal**:
- Remove entire switch statement from `functions_builtins.go`
- All 244 functions dispatched through registry
- Pure O(1) lookup for all built-in functions

**New Files**:
- `internal/interp/builtins/system.go` - System functions

**New Categories**:
- `CategorySystem` - System and runtime utilities

**Acceptance Criteria**:
- [ ] All 244 functions migrated and registered
- [ ] Switch statement completely removed
- [ ] Pure registry-based dispatch (O(1))
- [ ] All tests passing
- [ ] Performance benchmarks show no regression

---

### Task 3.7.9: Registry Enhancement and Optimization
**Estimated**: 2 days | **Functions**: N/A | **Total**: 244 (100%)

**Objective**: Enhance registry with metadata, tooling, and optimization.

**Enhancements**:

1. **Parameter Validation Metadata**:
   ```go
   type FunctionInfo struct {
       Name        string
       Function    BuiltinFunc
       Category    Category
       Description string
       ParamCount  ParamCount       // NEW: min/max params
       ParamTypes  [][]string       // NEW: accepted types per param
       ReturnType  string           // NEW: return type
       Examples    []string         // NEW: usage examples
   }

   type ParamCount struct {
       Min int
       Max int  // -1 for variadic
   }
   ```

2. **Registry Inspection Tools**:
   - CLI command: `dwscript builtins list [category]`
   - CLI command: `dwscript builtins info <function>`
   - Debug endpoint for querying registry at runtime

3. **Performance Profiling**:
   - Add call count tracking (optional, disabled by default)
   - Add execution time tracking
   - Registry statistics endpoint

4. **Documentation Generation**:
   - Auto-generate markdown reference from registry
   - Include all 244 functions with examples
   - Generate category summaries

**New Files**:
- `cmd/dwscript/cmd_builtins.go` - CLI commands for registry inspection
- `internal/interp/builtins/metadata.go` - Metadata structures
- `internal/interp/builtins/profiling.go` - Performance profiling (optional)

**Acceptance Criteria**:
- [ ] FunctionInfo extended with full metadata
- [ ] All functions have parameter/return type info
- [ ] CLI commands for registry inspection working
- [ ] Documentation auto-generated from registry
- [ ] Performance profiling available (opt-in)
- [ ] All tests passing

---

## Success Metrics

### Phase-by-Phase Targets

| Task | Functions | Total | % Complete | Categories | Switch Cases |
|------|-----------|-------|------------|------------|--------------|
| 3.7.2 ✅ | +169 | 169 | 69% | 4 | 90 |
| 3.7.3 | +15 | 184 | 75% | 5 | 75 |
| 3.7.4 | +2 | 186 | 76% | 6 | 73 |
| 3.7.5 | +17 | 203 | 83% | 8 | 56 |
| 3.7.6 | +9 | 212 | 87% | 10 | 47 |
| 3.7.7 | +21 | 233 | 95% | 12 | 26 |
| 3.7.8 | +11 | 244 | 100% | 13 | 0 |
| 3.7.9 | - | 244 | 100% | 13 | 0 |

### Overall Goals

**Code Quality**:
- [x] 69% functions in registry (169/244) ← Current
- [ ] 100% functions in registry (244/244) ← After 3.7.8
- [ ] 0 switch cases ← After 3.7.8
- [ ] Complete metadata for all functions ← After 3.7.9

**Performance**:
- [x] O(1) lookup for 69% of functions ← Current
- [ ] O(1) lookup for 100% of functions ← After 3.7.8
- [ ] Performance profiling available ← After 3.7.9

**Documentation**:
- [x] Architecture documented ← Current
- [x] Migration roadmap documented ← Current
- [ ] Auto-generated API reference ← After 3.7.9
- [ ] Usage examples for all functions ← After 3.7.9

## Estimated Timeline

Based on the estimated effort for each task:

| Task | Duration | Cumulative |
|------|----------|------------|
| 3.7.3 | 3 days | 3 days |
| 3.7.4 | 1 day | 4 days |
| 3.7.5 | 4 days | 8 days |
| 3.7.6 | 3 days | 11 days |
| 3.7.7 | 5 days | 16 days |
| 3.7.8 | 2 days | 18 days |
| 3.7.9 | 2 days | 20 days |

**Total Estimated Effort**: ~20 working days (4 weeks)

With parallel work or experienced developer, this could be compressed to 2-3 weeks.

## Dependencies and Risks

### Dependencies
- Each task builds on the previous (Context extensions are cumulative)
- Tasks 3.7.3-3.7.5 are relatively independent (can parallelize)
- Tasks 3.7.6-3.7.7 depend on 3.7.5 (ordinal helpers needed)
- Task 3.7.8 depends on all previous tasks
- Task 3.7.9 requires 3.7.8 completion

### Risks
1. **Context Interface Bloat**: Adding too many methods to Context
   - *Mitigation*: Keep methods minimal, use helper types

2. **Performance Regression**: Function pointer evaluation overhead
   - *Mitigation*: Benchmark each task, optimize hot paths

3. **Breaking Changes**: Extending Context might break existing code
   - *Mitigation*: Context is internal, only affects Interpreter

4. **Test Coverage**: Migrated functions might miss edge cases
   - *Mitigation*: Run full test suite after each task

## Next Steps

1. **Review and Approve** this completion plan
2. **Begin Task 3.7.3** - Extend Context for type-dependent functions
3. **Iterate** through tasks 3.7.4-3.7.9
4. **Celebrate** when all 244 functions are in the registry! 🎉

## Conclusion

The built-in function migration is well underway with 69% complete. The remaining work is clearly defined across 7 incremental tasks that will:
- Migrate all 244 functions to the registry
- Remove the switch statement entirely
- Achieve pure O(1) dispatch for all built-ins
- Provide rich metadata and tooling
- Maintain 100% backward compatibility

This represents a significant improvement in code quality, maintainability, and discoverability of DWScript's built-in functions.
