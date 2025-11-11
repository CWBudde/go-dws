# REFACTOR.md

## Overview

This document provides a comprehensive strategy for refactoring the go-dws codebase to improve maintainability, readability, and organization. The project has grown significantly, with 20+ files exceeding 30KB and several directories containing 60-90+ files.

**Current Issues:**
- Large files (up to 80KB) are difficult to navigate and maintain
- Flat directory structures make it hard to find related code
- Test files mixed with implementation files
- No clear organizational pattern for builtin functions

**Goals:**
- Split large files into focused, logical components
- Organize files into subdirectories by feature area
- Maintain parallel structure between `internal/interp/` and `internal/semantic/`
- Improve code discoverability and reduce cognitive load
- Preserve all functionality and test coverage

---

## File Splitting Strategy

### ✅ Completed Splits

The following large files have been successfully split into smaller, focused files:

#### ✅ 1.1 objects.go → Split into 4 files
- `objects_instantiation.go` (203 lines)
- `objects_properties.go` (693 lines)
- `objects_methods.go` (913 lines)
- `objects_hierarchy.go` (619 lines)

#### ✅ 1.2 functions.go → Split into 6 files
- `functions_calls.go` (420 lines)
- `functions_builtins.go` (477 lines)
- `functions_user.go` (219 lines)
- `functions_pointers.go` (199 lines)
- `functions_records.go` (354 lines)
- `functions_typecast.go` (532 lines)

#### ✅ 1.3 statements.go → Split into 4 files
- `statements_declarations.go` (403 lines)
- `statements_assignments.go` (643 lines)
- `statements_control.go` (303 lines)
- `statements_loops.go` (493 lines)

#### ✅ 1.4 builtins_datetime.go → Split into 3 files
- `builtins_datetime_format.go` (320 lines)
- `builtins_datetime_calc.go` (472 lines)
- `builtins_datetime_info.go` (338 lines)

---

### 🔄 Remaining Work

#### Priority 1: Remaining Large Files



#### 1.5 internal/interp/builtins_math.go (35KB, 1,123 lines) - TO DO

**Current Content:** 40 math builtin functions including trigonometry, logarithms, rounding, min/max, power, sqrt, etc.

**Proposed Split:**

```
builtins_math.go (REMOVE) →
├── builtins_math_basic.go       // Basic arithmetic and utility functions
├── builtins_math_trig.go        // Trigonometric and hyperbolic functions
└── builtins_math_convert.go     // Rounding, truncation, and conversions
```

**Split Details:**

**builtins_math_basic.go** (~350 lines)
- `builtinAbs()` - Absolute value
- `builtinMin()`, `builtinMax()` - Min/max functions
- `builtinSqr()` - Square
- `builtinSqrt()` - Square root
- `builtinPower()` - Exponentiation
- `builtinExp()` - Exponential
- `builtinLn()`, `builtinLog10()`, `builtinLog2()` - Logarithms
- `builtinSign()` - Sign function
- `builtinRandom()`, `builtinRandomize()` - Random numbers

**builtins_math_trig.go** (~400 lines)
- `builtinSin()`, `builtinCos()`, `builtinTan()` - Trigonometric
- `builtinArcSin()`, `builtinArcCos()`, `builtinArcTan()`, `builtinArcTan2()`
- `builtinSinh()`, `builtinCosh()`, `builtinTanh()` - Hyperbolic
- `builtinArcSinh()`, `builtinArcCosh()`, `builtinArcTanh()`
- `builtinDegToRad()`, `builtinRadToDeg()` - Angle conversions
- Trigonometry helpers

**builtins_math_convert.go** (~373 lines)
- `builtinRound()` - Round to nearest integer
- `builtinTrunc()` - Truncate decimal part
- `builtinFloor()` - Round down
- `builtinCeil()` - Round up
- `builtinFrac()` - Fractional part
- `builtinInt()` - Integer part
- `builtinIntPower()` - Integer exponentiation
- `builtinMod()` - Modulo
- `builtinDivMod()` - Division with remainder
- Type conversion helpers

**Migration Notes:**
- Keep files in `internal/interp/` directory (subdirectory organization is separate task)
- Update tests as needed

---

#### Priority 2: Large Files (Still in flat structure - consider for future work)

The following large files should be considered for splitting if they become problematic:

- `internal/semantic/analyze_classes.go` (48KB, 1,466 lines)
- `internal/bytecode/vm.go` (47KB, 2,172 lines)
- `internal/bytecode/compiler.go` (42KB, 1,799 lines)
- `internal/parser/expressions.go` (40KB, 1,222 lines)
- `internal/interp/expressions.go` (38KB, 1,222 lines)
- `internal/interp/builtins_strings.go` (33KB)
- `internal/interp/builtins_core.go` (34KB)

**Note:** These files are currently manageable. Defer splitting until Phase 2 subdirectory organization is considered.

---

## Future Consideration: Subdirectory Organization (Phase 2 - Not Implemented)

### Overview

The original plan included reorganizing files into subdirectories. However, Phase 1 has shown that file splitting with clear naming prefixes may be sufficient.

**Current Approach (Working Well):**
- Flat directory structure
- Prefixed naming: `objects_*.go`, `functions_*.go`, `statements_*.go`, `builtins_*.go`
- Easy to find related files
- No import path changes needed

**Alternative Approach (Original Plan - Not Started):**

If subdirectory organization is pursued in the future, consider:

```
internal/interp/
├── builtins/      # Move builtins_*.go files here
├── objects/       # Move objects_*.go files here
├── functions/     # Move functions_*.go files here
├── statements/    # Move statements_*.go files here
├── tests/         # Move all *_test.go files here
└── ...            # Core files remain in root
```

**Trade-offs to consider:**
- ✅ **Pro:** Further reduces root directory file count
- ✅ **Pro:** Clearer logical grouping
- ✅ **Pro:** Test files separated from implementation
- ❌ **Con:** Requires import path changes across codebase
- ❌ **Con:** May introduce package boundary issues
- ❌ **Con:** More complex navigation (deeper nesting)
- ❌ **Con:** Current prefix naming already provides good organization

**Recommendation:** The current flat structure with prefixed naming is working well. Only proceed with subdirectory organization if the file count becomes truly unmanageable (>200 files) or if there's a compelling need for package boundaries.

---

## Implementation Status

### ✅ Phase 1: Critical File Splits - MOSTLY COMPLETE

**Goal:** Split the largest, most unwieldy files

**Completed:**
1. ✅ Split `internal/interp/objects.go` (79KB → 4 files)
2. ✅ Split `internal/interp/functions.go` (67KB → 6 files)
3. ✅ Split `internal/interp/statements.go` (52KB → 4 files)
4. ✅ Split `builtins_datetime.go` (40KB → 3 files)

**Remaining:**
5. ⏳ Split `builtins_math.go` (35KB, 1,123 lines → 3 files)

**Status:** 4 of 5 file splits complete. All split files remain in flat directory structure (subdirectory organization deferred).

---

### 🔄 Phase 2: Subdirectory Organization - NOT STARTED

**Current State:**
- All files remain in flat directory structure
- No `builtins/`, `objects/`, `functions/`, `statements/`, `tests/` subdirectories created
- File count: `internal/interp/` still has 116 Go files in root directory

**Original Plan:**
```
internal/interp/
├── builtins/          # Not created
├── objects/           # Not created
├── functions/         # Not created
├── statements/        # Not created
├── values/            # Not created
└── tests/             # Not created
```

**Decision Point:** Should subdirectory organization proceed, or is the flat structure with prefixed filenames (e.g., `objects_*.go`, `functions_*.go`) acceptable?

**Benefits of current approach:**
- ✅ File splitting achieved without import path changes
- ✅ No package boundary complications
- ✅ Easier to navigate with prefixed names
- ✅ Simpler refactoring process

**Drawbacks:**
- ❌ Still 116 files in single directory
- ❌ Tests mixed with implementation
- ❌ Less clear logical grouping

**Recommendation:** Complete Phase 1 (math split), then evaluate if subdirectory organization is worth the additional complexity.

---

### 📋 Remaining Phases (Original Plan - Under Review)

The following phases from the original plan have **not been started** and should be re-evaluated:

**Phase 3: Semantic Package Refactor**
- Status: NOT STARTED
- Files remain unsplit and in flat structure

**Phase 4: Bytecode Package Refactor**
- Status: NOT STARTED
- `vm.go` (2,172 lines) and `compiler.go` (1,799 lines) remain unsplit

**Phase 5: Test Organization**
- Status: NOT STARTED
- Test files remain mixed with implementation

**Phase 6: Value Types Organization**
- Status: NOT STARTED
- Value types remain in root

**Phase 7: Documentation Updates**
- Status: PARTIAL
- REFACTOR.md exists but needs updating (this document)
- CLAUDE.md may need updates to reflect file splits

---

## Migration Guidelines

### General Principles

1. **One change at a time** - Split one file or move one group at a time
2. **Test after every change** - Run full test suite after each modification
3. **Maintain git history** - Use `git mv` when moving files
4. **Update imports immediately** - Don't let broken imports accumulate
5. **Document as you go** - Update comments and docs alongside code changes

### File Splitting Process

For each file to be split:

1. **Create target directory** (if needed)
   ```bash
   mkdir -p internal/interp/objects
   ```

2. **Create new files** with appropriate headers
   ```go
   // Package objects contains object-oriented feature implementations.
   package objects

   import (
       "github.com/MeKo-Tech/go-dws/internal/ast"
       // ... other imports
   )
   ```

3. **Copy functions to new files** based on logical grouping

4. **Update package references**
   - Change `package interp` to `package objects` (or appropriate name)
   - Adjust visibility (exported vs unexported)

5. **Update imports in other files**
   ```go
   import (
       "github.com/MeKo-Tech/go-dws/internal/interp/objects"
   )
   ```

6. **Update function calls**
   ```go
   // Old: evalNewExpression(...)
   // New: objects.EvalNewExpression(...)
   ```

7. **Run tests**
   ```bash
   go test ./internal/interp/... -v
   ```

8. **Delete original file** only after all tests pass
   ```bash
   git rm internal/interp/objects.go
   ```

9. **Commit**
   ```bash
   git add .
   git commit -m "refactor: split objects.go into objects/ subdirectory"
   ```

### Import Path Updates

When creating subdirectories, import paths change:

**Before:**
```go
import "github.com/MeKo-Tech/go-dws/internal/interp"

interp.NewInterpreter()
```

**After:**
```go
import (
    "github.com/MeKo-Tech/go-dws/internal/interp"
    "github.com/MeKo-Tech/go-dws/internal/interp/builtins"
    "github.com/MeKo-Tech/go-dws/internal/interp/objects"
)

interp.NewInterpreter()
builtins.RegisterAll()
objects.Instantiate()
```

### Test File Organization

When moving test files:

1. **Preserve test package names**
   ```go
   // Can use either:
   package interp_test  // Black-box testing
   package interp       // White-box testing
   ```

2. **Update relative paths** in test data
   ```go
   // Old: testdata/simple.dws
   // New: ../../testdata/simple.dws (if tests moved to tests/ subdirectory)
   ```

3. **Consider test helper consolidation**
   - Create shared test helpers in `tests/helpers.go`
   - Reduce duplication across test files

### Visibility Considerations

When splitting files, consider function visibility:

**Keep unexported** (lowercase) if:
- Function is only used within the package
- Function is an implementation detail

**Make exported** (uppercase) if:
- Function needs to be called from other packages
- Function is part of public API

**Example:**
```go
// objects/instantiation.go

// EvalNewExpression evaluates object instantiation (exported - called from interp)
func EvalNewExpression(node *ast.NewExpression, interp *Interpreter) Value {
    return createObjectInstance(node, interp)
}

// createObjectInstance is an internal helper (unexported)
func createObjectInstance(node *ast.NewExpression, interp *Interpreter) Value {
    // ...
}
```

### Common Pitfalls to Avoid

❌ **Don't split arbitrarily** - Split based on logical boundaries
❌ **Don't break circular dependencies** - Be mindful of package dependencies
❌ **Don't skip tests** - Always run tests after changes
❌ **Don't batch too many changes** - Commit frequently
❌ **Don't forget documentation** - Update docs alongside code

✅ **Do split by feature** - Group related functions
✅ **Do maintain package cohesion** - Keep related code together
✅ **Do test incrementally** - Run tests after each change
✅ **Do commit frequently** - Small, focused commits
✅ **Do update documentation** - Keep docs in sync

---

## Testing Strategy

### After Each File Split

1. **Unit tests** - Run package tests
   ```bash
   go test ./internal/interp/objects -v
   ```

2. **Integration tests** - Run interpreter tests
   ```bash
   go test ./internal/interp -v -run TestDWScriptFixtures
   ```

3. **Full suite** - Run all tests
   ```bash
   go test ./... -v
   ```

4. **Coverage check** - Ensure coverage doesn't drop
   ```bash
   go test -cover ./internal/interp/...
   ```

### After Directory Reorganization

1. **Import checks** - Verify no broken imports
   ```bash
   go build ./...
   ```

2. **Test discovery** - Ensure all tests still run
   ```bash
   go test ./... -v | grep -c "PASS"
   ```

3. **Fixture tests** - Run complete DWScript test suite
   ```bash
   go test ./internal/interp -run TestDWScriptFixtures -v
   ```

### Regression Prevention

1. **Run tests before and after** - Compare results
2. **Check test coverage** - Should remain stable or increase
3. **Verify fixture test pass rate** - Should not decrease
4. **Test CLI tool** - Ensure `dwscript` commands still work

---

## Current Status and Metrics

### Metrics After Phase 1 (File Splits)

**Before refactoring:**
- **Largest file:** 79KB (objects.go)
- **Files over 50KB:** 6 files
- **Files over 30KB:** 20 files

**Current state (Phase 1 mostly complete):**
- **Largest remaining unsplit file:** 35KB (builtins_math.go)
- **Files over 50KB:** 0 files ✅
- **Files over 30KB in interp:** ~6 files (mostly builtins and expressions.go)
- **Total files in internal/interp:** 116 files (increased from splits)
- **Directory structure:** Flat (no subdirectories created)

### Improvements Achieved

✅ **Large files eliminated** - No files over 50KB
✅ **Improved navigation** - Prefixed filenames (objects_*, functions_*, statements_*) make related code easy to find
✅ **Reduced cognitive load** - Smaller, focused files (200-900 lines each)
✅ **Better maintainability** - Changes localized to specific files
✅ **No import changes** - All refactoring done within same package

### Remaining Concerns

⚠️ **High file count** - 116 files in internal/interp/ root directory
⚠️ **Test file clutter** - Test files mixed with implementation
⚠️ **Other large files** - vm.go (2,172 lines), compiler.go (1,799 lines), analyze_classes.go (1,466 lines) remain unsplit

---

## Future Considerations

### As the Project Grows

1. **Monitor file sizes** - Split files before they exceed 50KB
2. **Consistent patterns** - Apply same organization to new packages
3. **Regular refactoring** - Don't let technical debt accumulate
4. **Documentation** - Keep CLAUDE.md updated with structure changes

### Potential Future Subdirectories

If packages continue to grow, consider:

- `internal/interp/vm/` - Alternative VM implementation
- `internal/interp/optimizer/` - AST optimization passes
- `internal/semantic/inference/` - Type inference
- `internal/codegen/` - Code generation (if transpilation added)
- `pkg/stdlib/` - Standard library modules

### WebAssembly Considerations (Stage 10.15)

The planned `pkg/platform/` and `pkg/wasm/` packages should follow these same organizational principles:

```
pkg/
├── platform/
│   ├── filesystem/
│   ├── console/
│   └── runtime/
└── wasm/
    ├── bindings/
    ├── interop/
    └── stdlib/
```

---

## Conclusion

Phase 1 of the refactoring has been largely successful. The most critical large files have been split into smaller, manageable components with clear naming conventions:

**Achievements:**
- ✅ Eliminated all files over 50KB
- ✅ Split 4 major files into 17 smaller files
- ✅ Maintained flat directory structure (avoided complex package reorganization)
- ✅ Used clear prefixing (objects_*, functions_*, statements_*) for easy navigation

**Next Steps:**
1. **Complete Phase 1:** Split `builtins_math.go` into 3 files
2. **Evaluate Phase 2:** Decide if subdirectory organization is necessary or if the current flat structure with prefixed names is sufficient
3. **Consider other packages:** Evaluate if `semantic/`, `bytecode/`, and `parser/` packages need similar treatment

**Key Principle:**
The current approach (file splitting with prefix naming in flat structure) has achieved the primary goal of reducing file sizes without the complexity of package reorganization. This may be the optimal balance for this project.

---

**Document Version:** 2.0
**Last Updated:** 2025-11-11
**Status:** Phase 1 mostly complete; Phase 2+ under review
