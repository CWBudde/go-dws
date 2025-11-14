# REFACTOR.md

## Overview

This document identifies technical debt and refactoring opportunities in the go-dws codebase. As the project has grown, several files have become large and complex, with high cyclomatic complexity that impacts maintainability.

**Current State:**

- Several files exceed 1,000 lines (largest: 2,452 lines)
- High cyclomatic complexity in critical functions (up to 209!)
- Test files can be very large (up to 72KB)
- Some functions exceed recommended complexity thresholds

**Goals:**

- Reduce file sizes to improve navigability (target: <1,500 lines)
- Lower cyclomatic complexity (target: <15 per function)
- Split large test files for better organization
- Maintain all functionality and test coverage

---

## Priority 2: Large Implementation Files (>1,200 lines)

### 🔵 P2.8: internal/interp/value.go (1,164 lines)

**Current:** Value interface and all value type implementations
**Target:** Split by value category

```plain
value.go → Split into:
├── value.go              (~200 lines) - Value interface, basic types
├── value_collections.go  (~350 lines) - Array, Set, Map values
├── value_objects.go      (~300 lines) - Object, Class values
└── value_functions.go    (~300 lines) - Function, Lambda values
```

---

## Priority 3: Large Test Files

These test files are very large and would benefit from splitting:

### ✅ P3.2: internal/interp/math_test.go (64KB) - COMPLETED

**Status:** Split into 3 files (45KB + 9.8KB + 9.6KB = 64.4KB total)

**Result:**
```plain
├── math_basic_test.go   (45KB) - Abs, Sqrt, Power, Min, Max, Random, Clamp, Unsigned32 tests (43 tests)
├── math_trig_test.go    (9.6KB) - Sin, Cos, Tan, Exp, Ln tests (8 tests)
└── math_convert_test.go (9.8KB) - Round, Trunc, Floor, Ceil tests (7 tests)
```

### ✅ P3.3: internal/parser/arrays_test.go (52KB) - COMPLETED

**Status:** Split into 2 files (22KB + 29KB = 51KB total)

**Result:**
```plain
├── arrays_literal_test.go     (22KB) - Array literal parsing, type declarations
└── arrays_operations_test.go  (29KB) - Array indexing/operations, assignments
```

### 🟢 P3.4: internal/parser/functions_test.go (48KB)

**Target:** Split by function feature

```plain
├── functions_decl_test.go  (~24KB) - Function declaration parsing
└── functions_call_test.go  (~24KB) - Function call parsing
```

### 🟢 P3.5: internal/interp/set_test.go (48KB)

**Target:** Split by set operations

```plain
├── set_basic_test.go    (~24KB) - Creation, membership, basic ops
└── set_advanced_test.go (~24KB) - Advanced operations, edge cases
```

### 🟢 P3.6: internal/bytecode/compiler_test.go (48KB)

**Target:** Mirror compiler.go split

```plain
├── compiler_statements_test.go   (~16KB)
├── compiler_expressions_test.go  (~16KB)
└── compiler_functions_test.go    (~16KB)
```

### 🟢 P3.7: internal/parser/classes_test.go (44KB)

**Target:** Split by class feature

```plain
├── classes_decl_test.go    (~22KB) - Class declaration parsing
└── classes_members_test.go (~22KB) - Method/property parsing
```

### 🟢 P3.8: internal/interp/property_test.go (44KB)

**Target:** Split by property complexity

```plain
├── property_basic_test.go    (~22KB) - Basic get/set tests
└── property_advanced_test.go (~22KB) - Visibility, inheritance tests
```

### 🟢 P3.9: internal/interp/interpreter_test.go (44KB)

**Target:** Split by feature complexity

```plain
├── interpreter_basic_test.go    (~15KB) - Literals, variables, expressions
├── interpreter_control_test.go  (~15KB) - Control flow tests
└── interpreter_advanced_test.go (~14KB) - Closures, recursion, edge cases
```

### 🟢 P3.10: internal/types/classes_test.go (40KB)

**Target:** Split by class feature

```plain
├── classes_basic_test.go       (~20KB) - Basic class type tests
└── classes_inheritance_test.go (~20KB) - Inheritance/polymorphism tests
```

### 🟢 P3.11: internal/interp/variant_test.go (40KB)

**Target:** Split by variant operation

```plain
├── variant_basic_test.go    (~20KB) - Basic variant operations
└── variant_advanced_test.go (~20KB) - Complex conversions, edge cases
```

### 🟢 P3.12: internal/interp/lambda_test.go (40KB)

**Target:** Split by lambda complexity

```plain
├── lambda_basic_test.go    (~20KB) - Basic lambdas, simple captures
└── lambda_advanced_test.go (~20KB) - Nested lambdas, complex captures
```

### 🟢 P3.13: internal/semantic/analyze_builtin_datetime.go (945 lines, 40KB)

**Target:** Split by datetime category

```plain
├── analyze_builtin_datetime_format.go (~315 lines)
├── analyze_builtin_datetime_calc.go   (~315 lines)
└── analyze_builtin_datetime_info.go   (~315 lines)
```

---

## Priority 4: Medium Complexity Issues (15-30)

These functions have elevated complexity but are not critical:

- `internal/bytecode/optimizer.go:inlineSequenceForCall` (26)
- `internal/bytecode/compiler_expressions.go:compileBinaryExpression` (22)
- `internal/bytecode/bytecode.go:String` (22)
- `internal/bytecode/optimizer.go:collapseStackShuffles` (22)
- `internal/bytecode/compiler_expressions.go:compileExpression` (20)
- `pkg/ast/classes.go:String` (18)
- `internal/bytecode/compiler_statements.go:compileExceptClause` (18)
- `pkg/ast/functions.go:String` (17)
- `internal/bytecode/compiler_statements.go:compileStatement` (16)

**Strategy:** Extract helper functions and break up large switch statements

---

## Priority 5: Minor Linting Issues

### Ineffectual Assignments

- `internal/interp/objects_methods.go:643` - ineffectual assignment to `found`
- `internal/lexer/lexer.go:416` - ineffectual assignment to `lastError`

### Unreachable Case Clauses

- `internal/bytecode/compiler_core.go:508-518` - TypedExpression case matches before specific literal types

### Unchecked Errors (in test files)

Multiple test files have unchecked error returns. While tests are lower priority, these should be addressed for completeness.

**Strategy:** Add `_ =` prefix to intentionally ignored errors, or add proper error handling

---

## Implementation Strategy

### Phase 1: Critical Complexity Reduction (Immediate)

Focus on the highest complexity functions that impact maintainability:

1. **Week 1:** Refactor [pkg/ast/visitor.go](pkg/ast/visitor.go#L17) Walk function (complexity 209 → <15)
2. **Week 2:** Refactor [internal/bytecode/compiler_core.go](internal/bytecode/compiler_core.go#L642) evaluateBinary (complexity 70 → <15)
3. **Week 3:** Refactor [internal/bytecode/optimizer.go](internal/bytecode/optimizer.go#L871) foldBinaryOp (complexity 47 → <15)

### Phase 2: Large File Splitting (Progressive)

Split files in order of size/impact:

1. Split [internal/bytecode/vm_builtins.go](internal/bytecode/vm_builtins.go) (2,452 lines → 6 files)
2. Split [internal/semantic/analyze_builtin_string.go](internal/semantic/analyze_builtin_string.go) (1,343 lines → 3 files)
3. Split [internal/interp/builtins_core.go](internal/interp/builtins_core.go) (1,296 lines → 4 files)
4. Split [internal/interp/helpers.go](internal/interp/helpers.go) (1,177 lines → 3 files)
5. Split [internal/interp/value.go](internal/interp/value.go) (1,164 lines → 4 files)

### Phase 3: Test File Organization (As Time Permits)

Split large test files to improve test organization and execution:

1. Split largest test files first (string_test, math_test, arrays_test)
2. Then medium test files (functions_test, set_test, compiler_test)
3. Finally smaller test files as needed

### Phase 4: Medium Complexity Reduction (Ongoing)

Address functions with complexity 15-30 as they're encountered during feature work.

### Phase 5: Minor Cleanup (Ongoing)

Fix linting issues (ineffectual assignments, unchecked errors) as time permits.

---

## File Splitting Guidelines

When splitting large files:

1. **Logical grouping** - Group related functions together
2. **Consistent naming** - Use clear, descriptive prefixes
3. **Preserve functionality** - All tests must pass after split
4. **Update imports** - Ensure all imports remain correct
5. **Test thoroughly** - Run full test suite after each split
6. **Commit atomically** - One split per commit

**Example naming pattern:**

```plain
builtins_core.go →
├── builtins_io.go         (I/O functions)
├── builtins_conversion.go (Type conversions)
├── builtins_type.go       (Type introspection)
└── builtins_misc.go       (Miscellaneous)
```

---

## Complexity Reduction Techniques

### For High Complexity Functions

1. **Extract Methods** - Break large switch statements into separate functions
2. **Strategy Pattern** - Use maps of functions instead of switch statements
3. **Table-Driven Logic** - Replace complex conditionals with lookup tables
4. **Early Returns** - Reduce nesting by returning early
5. **Helper Functions** - Extract repeated logic into helpers

**Example refactoring:**

```go
// Before (complexity: 30)
func handleNode(n Node) error {
    switch v := n.(type) {
    case *TypeA:
        // 20 lines of logic
    case *TypeB:
        // 25 lines of logic
    case *TypeC:
        // 30 lines of logic
    // ... 20 more cases
    }
}

// After (complexity: 5)
func handleNode(n Node) error {
    switch v := n.(type) {
    case *TypeA:
        return handleTypeA(v)
    case *TypeB:
        return handleTypeB(v)
    case *TypeC:
        return handleTypeC(v)
    // ... dispatch to handlers
    }
}

func handleTypeA(n *TypeA) error { /* focused logic */ }
func handleTypeB(n *TypeB) error { /* focused logic */ }
func handleTypeC(n *TypeC) error { /* focused logic */ }
```

---

## Testing After Refactoring

After each refactoring change:

```bash
# Run affected package tests
go test ./internal/bytecode -v

# Run full test suite
go test ./... -v

# Check coverage hasn't decreased
go test -cover ./internal/bytecode

# Run linters
golangci-lint run

# Run fixture tests (integration)
go test ./internal/interp -run TestDWScriptFixtures -v
```

---

## Success Metrics

### Target Metrics

- ✅ **No functions with complexity > 30** (currently: 4 functions)
- ✅ **No files > 1,500 lines** (currently: 1 file at 2,452 lines)
- ✅ **90%+ test coverage maintained**
- ✅ **All fixture tests passing**
- ✅ **Zero critical linting issues**

### Progress Tracking

Track progress by:

1. Monitoring largest files with `find . -name "*.go" -not -name "*_test.go" -exec wc -l {} + | sort -rn | head -10`
2. Running complexity checks with `golangci-lint run -E cyclop,funlen`
3. Checking test coverage with `go test -cover ./...`
4. Monitoring git commit history for refactoring work

---

## Notes

- **Flat directory structure** - Current flat structure with prefixed naming works well; no subdirectory organization planned due to Go circular import constraints
- **One change at a time** - Small, focused refactoring changes with tests after each
- **Maintain compatibility** - All refactoring must preserve existing functionality
- **Document changes** - Update this file as refactoring progresses

---

**Document Version:** 3.0
**Last Updated:** 2025-11-13
**Focus:** Active refactoring work needed - complexity reduction and file splitting
