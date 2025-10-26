# Missing Features: Recommendations and Roadmap

**Document Version**: 1.0
**Last Updated**: 2025-10-26
**Purpose**: Prioritized recommendations for implementing missing DWScript features in go-dws

---

## Executive Summary

**Current State**: go-dws has ~40% feature parity with DWScript (164 of 408 features implemented)

**Analysis**: Of the 226 remaining features:
- **58 are HIGH priority** (26%) - Critical for real programs
- **94 are MEDIUM priority** (42%) - Useful but not essential
- **74 are LOW priority** (33%) - Nice to have
- **4 are OUT OF SCOPE** (platform-specific)

**Recommendation**: Focus on the **58 HIGH priority features** in Stages 8-9 to bring go-dws to ~70% parity and make it production-ready for most use cases.

---

## Immediate Next Steps (Stage 8 Completion)

### Priority 1: Complete Composite Types (7 tasks remaining)

**Rationale**: Arrays/records/sets are partially done, finishing them has high ROI

**Tasks**:
1. ✅ Array assignment - DONE
2. ⏸️ Multi-dimensional array syntax: `array[M, N] of Integer`
3. ⏸️ Record methods (runtime support)
4. ⏸️ Set symmetric difference operator `><`
5. ⏸️ Large sets (>64 elements, use map instead of bitset)
6. ⏸️ For-in iteration over sets

**Estimated Effort**: 10 tasks, ~2 weeks
**Impact**: Completes foundational type system

---

### Priority 2: Ordinal Functions (High Impact, Low Effort)

**Rationale**: Inc/Dec/Succ/Pred are used extensively in Pascal code

**Missing Functions**:
- `Inc(x)`, `Inc(x, delta)` - Increment
- `Dec(x)`, `Dec(x, delta)` - Decrement
- `Succ(x)` - Successor
- `Pred(x)` - Predecessor
- `Low(enum)` - Lowest enum value
- `High(enum)` - Highest enum value

**Estimated Effort**: 8 tasks, ~1 week
**Impact**: Makes enum/integer manipulation natural

**Implementation Note**: These should work on:
- Integers: `Inc(x)` → `x := x + 1`
- Enums: `Succ(Red)` → `Green`
- Chars (when implemented): `Succ('a')` → `'b'`

---

### Priority 3: `const` Declarations

**Rationale**: Essential for readable code, avoiding magic numbers

**What's Missing**:
- Parse `const` keyword
- Semantic analysis for const (immutable)
- Const folding optimization (optional)

**Example**:
```pascal
const
    MAX_USERS = 1000;
    PI = 3.14159;
    APP_NAME = 'MyApp';
```

**Estimated Effort**: 6 tasks, ~1 week
**Impact**: Makes code more maintainable

---

### Priority 4: Assert Function

**Rationale**: Critical for testing and contracts

**What's Missing**:
- `Assert(condition)` - Raises exception if false
- `Assert(condition, message)` - With custom message

**Estimated Effort**: 3 tasks, ~1 day
**Impact**: Enables test-driven development in scripts

**Implementation**:
```go
func builtinAssert(args []Value) (Value, error) {
    cond := args[0].(BoolValue)
    if !cond {
        msg := "Assertion failed"
        if len(args) > 1 {
            msg = args[1].(StringValue)
        }
        return nil, &AssertionFailedError{Message: msg}
    }
    return nil, nil
}
```

---

## Stage 8 Remaining Work

### Built-in Functions (58 missing, prioritize these 20)

#### String Functions (Priority)
1. ✅ **Trim(s)** - Remove leading/trailing whitespace
2. ✅ **Insert(source, s, pos)** - Insert substring
3. ✅ **Delete(var s, pos, count)** - Delete substring
4. ✅ **Format(fmt, args)** - Printf-style formatting
5. **StringReplace(s, old, new)** - Replace substring

**Estimated**: 8 tasks, ~1 week

#### Math Functions (Priority)
1. ✅ **Min(a, b)**, **Max(a, b)** - Min/max values
2. ✅ **Sqr(x)** - Square (x²)
3. ✅ **Power(x, y)** - Exponentiation (x^y)
4. ✅ **Ceil(x)**, **Floor(x)** - Rounding functions
5. **RandomInt(max)** - Random integer [0, max)

**Estimated**: 10 tasks, ~1 week

#### Array Functions (Priority)
1. ✅ **Copy(arr)** - Deep copy array
2. ✅ **IndexOf(arr, value)** - Find element index
3. **Contains(arr, value)** - Test membership
4. **Reverse(var arr)** - Reverse in place
5. **Sort(var arr)** - Sort array

**Estimated**: 8 tasks, ~1 week

---

## Stage 9: Advanced Language Features

### Priority 1: Units/Modules System (CRITICAL)

**Rationale**: Required for organizing larger programs

**What's Needed**:
- `unit` declarations
- `uses` clauses (import other units)
- Unit initialization/finalization sections
- Namespace resolution
- Circular dependency handling
- Unit search paths

**Estimated Effort**: 45 tasks, ~4 weeks
**Impact**: Enables multi-file projects

**Design Considerations**:
- Map units to Go packages conceptually
- Support both relative and absolute paths
- Cache compiled units for performance

---

### Priority 2: Function/Method Pointers (HIGH VALUE)

**Rationale**: Enables callbacks, event handlers, higher-order functions

**What's Needed**:
- Procedural type declarations: `type TProc = procedure(x: Integer);`
- Function type declarations: `type TFunc = function(x: Integer): Integer;`
- Method pointers: `type TMethod = procedure of object;`
- Assignment to variables
- Pass as parameters
- Call through pointer

**Example**:
```pascal
type
    TComparator = function(a, b: Integer): Integer;

function BubbleSort(var arr: array of Integer; compare: TComparator);
begin
    // Use compare(arr[i], arr[j]) instead of hardcoded comparison
end;

function Ascending(a, b: Integer): Integer;
begin
    Result := a - b;
end;

BubbleSort(myArray, @Ascending);
```

**Estimated Effort**: 25 tasks, ~3 weeks
**Impact**: Enables functional programming patterns

---

### Priority 3: Type Aliases and Subranges

**Rationale**: Improves code clarity and type safety

**What's Needed**:
- Type aliases: `type TUserID = Integer;`
- Subrange types: `type TDigit = 0..9;`
- Range validation at runtime
- Type compatibility rules

**Example**:
```pascal
type
    TPercent = 0..100;
    TUserID = Integer;
    TFileName = String;

var
    progress: TPercent;
begin
    progress := 50;  // OK
    progress := 150; // Runtime error: out of range
end;
```

**Estimated Effort**: 12 tasks, ~1.5 weeks
**Impact**: Type-safe code, self-documenting

---

### Priority 4: External Function Registration (FFI)

**Rationale**: Bridge between Go and DWScript code

**What's Needed**:
- Register Go functions from host app
- Type marshaling (Go types ↔ DWScript types)
- Error handling
- Variadic functions
- Optional parameters

**Example**:
```go
// Go code
interp.RegisterFunction("HttpGet", func(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    return string(body), nil
})

// DWScript code
var html := HttpGet('https://example.com');
PrintLn(html);
```

**Estimated Effort**: 35 tasks, ~4 weeks
**Impact**: Unlocks Go ecosystem for scripts

---

## Stage 10: Polish and Advanced Features

### Medium Priority Features

#### 1. Lambdas/Anonymous Methods
- **Effort**: 30 tasks, ~3 weeks
- **Value**: Functional programming, cleaner callbacks
- **Depends On**: Function pointers

#### 2. Helpers (Class/Record/Type)
- **Effort**: 25 tasks, ~2.5 weeks
- **Value**: Extend existing types without modification
- **Example**: Add string helper methods

#### 3. Property Features (Indexed, Expression Getters)
- **Effort**: 12 tasks, ~1.5 weeks
- **Value**: Complete property system
- **Note**: Core properties already work (94%)

#### 4. DateTime Functions
- **Effort**: 20 tasks, ~2 weeks
- **Value**: Time/date manipulation
- **Note**: Can use Go's `time` package via FFI as interim

#### 5. JSON Support
- **Effort**: 18 tasks, ~2 weeks
- **Value**: Modern data interchange
- **Note**: Can use Go's `encoding/json` via FFI as interim

#### 6. Improved Error Messages and Stack Traces
- **Effort**: 10 tasks, ~1 week
- **Value**: Better debugging experience
- **Note**: Basic support exists, needs improvement

---

### Low Priority / Future Features

#### Generics
- **Effort**: 60+ tasks, ~6 weeks
- **Value**: Type-safe collections
- **Challenge**: Complex type system changes
- **Recommendation**: Defer to Stage 11+

#### RTTI (Runtime Type Information)
- **Effort**: 40 tasks, ~4 weeks
- **Value**: Reflection, serialization
- **Note**: Minimal RTTI may be enough

#### Contracts (Design by Contract)
- **Effort**: 15 tasks, ~2 weeks
- **Value**: Formal verification
- **Note**: Tasks 8.236-8.238 already defined

#### Delegates/Events
- **Effort**: 20 tasks, ~2 weeks
- **Value**: Event-driven programming
- **Depends On**: Function pointers

#### LINQ
- **Effort**: 35 tasks, ~4 weeks
- **Value**: Query syntax for collections
- **Depends On**: Lambdas, generics

---

## Recommended Roadmap

### Phase 1: Complete Stage 8 (2-3 months)

**Goal**: Finish composite types and essential built-ins

**Tasks**:
1. Complete array/record/set features (10 tasks)
2. Ordinal functions (8 tasks)
3. `const` declarations (6 tasks)
4. `Assert` function (3 tasks)
5. Priority string functions (8 tasks)
6. Priority math functions (10 tasks)
7. Priority array functions (8 tasks)
8. Contracts (15 tasks, optional)
9. Feature assessment documentation ✅ (DONE - this document)

**Total**: ~68 tasks
**Result**: go-dws at ~55% parity, usable for moderate programs

---

### Phase 2: Stage 9 Core (3-4 months)

**Goal**: Units system + FFI + critical missing features

**Tasks**:
1. Units/modules system (45 tasks)
2. Function/method pointers (25 tasks)
3. Type aliases (12 tasks)
4. External function registration / FFI (35 tasks)
5. Lambdas/anonymous methods (30 tasks)
6. Helpers (class/record) (25 tasks)

**Total**: ~172 tasks
**Result**: go-dws at ~75% parity, production-ready

---

### Phase 3: Stage 9 Polish (2-3 months)

**Goal**: Fill gaps, improve UX

**Tasks**:
1. Complete property features (12 tasks)
2. DateTime functions (20 tasks)
3. JSON support (18 tasks)
4. Improved debugging (10 tasks)
5. Performance profiling and optimization (15 tasks)
6. Documentation and examples (10 tasks)

**Total**: ~85 tasks
**Result**: go-dws at ~85% parity, mature

---

### Phase 4: Stage 10+ Advanced (Future)

**Features**:
- Generics (if demand exists)
- Advanced RTTI
- LINQ (if lambdas + generics done)
- Code generation (JavaScript backend - Stage 11)
- Additional libraries (encoding, crypto, etc.)

**Timeline**: 6-12 months
**Result**: Near-complete DWScript compatibility

---

## Quick Wins for Next Sprint

If starting work immediately, prioritize these for maximum impact:

### Week 1-2: Foundation
1. ✅ `const` declarations (6 tasks)
2. ✅ `Assert` function (3 tasks)
3. ✅ Ordinal functions: Inc/Dec/Succ/Pred (8 tasks)

### Week 3-4: Built-ins
4. ✅ String: Trim, Insert, Delete, Format (8 tasks)
5. ✅ Math: Min, Max, Sqr, Power, Ceil, Floor (10 tasks)
6. ✅ Array: Copy, IndexOf, Contains (6 tasks)

### Week 5-6: Complete Composites
7. ⏸️ Finish array features (multi-dim syntax, open array params)
8. ⏸️ Complete record methods runtime support
9. ⏸️ Set improvements (large sets, symm diff, for-in)

**Total**: 41 tasks in 6 weeks
**Impact**: Brings go-dws from 40% to 50% parity

---

## Metrics and Success Criteria

### Coverage Goals

| Phase | Features Implemented | Parity % | Usability |
|-------|---------------------|----------|-----------|
| **Current** | 164/408 | 40% | Basic scripts |
| **After Stage 8** | 232/408 | 55% | Moderate programs |
| **After Stage 9 Core** | 310/408 | 75% | Production-ready |
| **After Stage 9 Polish** | 350/408 | 85% | Mature |
| **Stage 10+** | 380+/408 | 90%+ | Near-complete |

### Test Coverage Goals

- Maintain >85% test coverage per package
- Each new feature requires comprehensive tests
- Port relevant DWScript test suite tests

### Performance Goals

- Lexer: Process 100K lines/sec
- Parser: Parse 50K lines/sec
- Interpreter: Execute 10K operations/sec
- Memory: <10MB overhead for typical scripts

---

## Dependencies and Blockers

### Internal Dependencies

```
Units System ← (depends on) → Module Resolution
    ↓
Function Pointers ← (enables) → Lambdas
    ↓
Lambdas ← (enables) → LINQ
    ↓
Generics ← (complex, optional)
```

### External Dependencies

- None currently
- Future: May want Go 1.22+ for features

### Potential Blockers

1. **Generics complexity**: Type system rewrite, defer if possible
2. **Performance**: May need bytecode VM (Stage 9 optional)
3. **Memory**: Large scripts may need GC tuning

---

## Community Feedback Integration

### How to Gather Feedback

1. **Alpha Release** (After Stage 8):
   - Release on GitHub with "alpha" tag
   - Gather feedback on missing features
   - Identify most-requested features

2. **Beta Release** (After Stage 9 Core):
   - Production trial with select users
   - Performance benchmarks
   - Stability testing

3. **1.0 Release** (After Stage 9 Polish):
   - Declare API stable
   - Semantic versioning from this point

### Feature Request Process

- GitHub Issues for feature requests
- Label: `enhancement`, `from-dwscript`, `breaking-change`
- Prioritize based on:
  - Number of requests
  - Compatibility with DWScript
  - Implementation effort
  - Breaking changes (avoid if possible)

---

## Conclusion

**Main Recommendation**: Focus on the **58 HIGH priority features** across Stages 8-9 to achieve production-ready status.

**Key Milestones**:
1. ✅ **Stage 8 completion** (55% parity, 2-3 months)
2. **Units + FFI** (65% parity, +3 months)
3. **Function pointers + Lambdas** (75% parity, +2 months)
4. **Polish** (85% parity, +2 months)

**Total Timeline to Production-Ready**: 9-10 months from now

**Alternative**: If resources are limited, completing just Stage 8 + Units + FFI (Steps 1-2) gives a solid foundation at 65% parity in ~6 months.

The feature matrix in `docs/feature-matrix.md` provides the detailed breakdown for planning sprints and tracking progress.

---

**Document Status**: ✅ Complete - Task 8.239w finished
