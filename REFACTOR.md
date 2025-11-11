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

#### ✅ 1.5 builtins_math.go → Split into 3 files
- `builtins_math_basic.go` (657 lines)
- `builtins_math_trig.go` (447 lines)
- `builtins_math_convert.go` (429 lines)

---

### 🔄 Remaining Work

#### Priority 1: Remaining Large Files

All priority 1 file splits have been completed! ✅

#### ✅ 1.5 internal/interp/builtins_math.go → Split into 3 files (COMPLETED)

Successfully split into:
- `builtins_math_basic.go` (657 lines) - Basic arithmetic and utility functions
- `builtins_math_trig.go` (447 lines) - Trigonometric and hyperbolic functions
- `builtins_math_convert.go` (429 lines) - Rounding, truncation, and conversions

All 40 math builtin functions properly distributed across the three files. Tests pass and functionality preserved.

---

#### Priority 2: Large Files (Still in flat structure - consider for future work)

The following large files should be considered for splitting if they become problematic:

- `internal/semantic/analyze_classes.go` (48KB, 1,466 lines)
- `internal/bytecode/vm.go` (47KB, 2,172 lines)
- `internal/bytecode/compiler.go` (42KB, 1,799 lines)
- `internal/parser/expressions.go` (40KB, 1,222 lines)
- `internal/interp/expressions.go` (38KB, 1,087 lines)
- `internal/interp/builtins_strings.go` (33KB)
- `internal/interp/builtins_core.go` (34KB)

**Note:** These files are currently manageable. Defer splitting until Phase 2 subdirectory organization is considered.

---

## Phase 2: Subdirectory Organization (Next Step - RECOMMENDED)

### Overview

The current flat structure with 116 files in `internal/interp/` should be organized into logical subdirectories. This is standard Go practice and will significantly improve code organization.

**Current State (Flat Structure):**
- 116 files in `internal/interp/` root
- Prefixed naming: `objects_*.go`, `functions_*.go`, `statements_*.go`, `builtins_*.go`
- Tests mixed throughout

**Target Structure (Subdirectory Packages):**

```
internal/interp/
├── builtins/               # Package builtins
│   ├── core.go             # Core builtins
│   ├── core_test.go
│   ├── math_basic.go       # Math functions (basic)
│   ├── math_trig.go        # Math functions (trig)
│   ├── math_convert.go     # Math functions (convert)
│   ├── math_test.go
│   ├── datetime_format.go
│   ├── datetime_calc.go
│   ├── datetime_info.go
│   ├── datetime_test.go
│   ├── strings.go
│   ├── strings_test.go
│   ├── arrays.go
│   ├── json.go
│   ├── variant.go
│   ├── ordinals.go
│   └── collections.go
├── objects/                # Package objects
│   ├── instantiation.go
│   ├── instantiation_test.go
│   ├── properties.go
│   ├── properties_test.go
│   ├── methods.go
│   ├── methods_test.go
│   ├── hierarchy.go
│   └── hierarchy_test.go
├── functions/              # Package functions
│   ├── calls.go
│   ├── calls_test.go
│   ├── builtins.go
│   ├── user.go
│   ├── pointers.go
│   ├── records.go
│   └── typecast.go
├── statements/             # Package statements
│   ├── declarations.go
│   ├── assignments.go
│   ├── control.go
│   └── loops.go
├── interpreter.go          # Main interpreter
├── expressions.go
├── environment.go
├── exceptions.go
├── value.go
├── class.go
└── ...
```

**Key Principles:**
1. **Tests side-by-side** - `math.go` and `math_test.go` in same directory (Go convention)
2. **Separate packages** - Each subdirectory is its own package
3. **Clear APIs** - Forces thinking about what should be public vs internal
4. **Logical grouping** - Related code together

### Benefits

✅ **Standard Go structure** - Follows idiomatic Go package organization
✅ **Clear boundaries** - Package boundaries enforce good design
✅ **Better encapsulation** - Private vs public functions are explicit
✅ **Easier navigation** - 10-20 files per directory vs 116 in root
✅ **Tests with code** - Standard Go convention, easier to maintain
✅ **Better documentation** - Each package can have its own doc.go
✅ **Reduced cognitive load** - Work within one package at a time
✅ **Clearer dependencies** - Import statements show relationships

### Costs (One-Time)

⚠️ **Import path changes** - Need to update imports across codebase
⚠️ **Function exports** - Need to capitalize public functions
⚠️ **Potential circular deps** - Need to design package boundaries carefully
⚠️ **Testing adjustments** - Some tests may need restructuring

### Avoiding Circular Dependencies

When creating subpackages, carefully consider dependencies:

**Safe dependency flow (no circular imports):**
```
builtins → (nothing)          # Self-contained builtin functions
objects → interp              # Objects may need Interpreter reference
functions → interp, builtins  # Functions call builtins, need Interpreter
statements → interp           # Statements need Interpreter
interp → all subpackages      # Main package orchestrates
```

**Key strategies:**
1. **Pass Interpreter as parameter** - Subpackages receive `*interp.Interpreter` as argument
2. **Interface abstraction** - Define interfaces in `interp`, implement in subpackages
3. **Keep builtins independent** - Builtin functions should only depend on Value types
4. **Avoid cross-dependencies** - `objects` shouldn't import `functions`, etc.

**Example pattern:**
```go
// internal/interp/objects/methods.go
package objects

import "github.com/MeKo-Tech/go-dws/internal/interp"

// EvalMethodCall needs access to interpreter state
func EvalMethodCall(i *interp.Interpreter, obj Value, method string, args []Value) (Value, error) {
    // Can call back to interpreter methods
    return i.EvalExpression(methodBody)
}
```

### Migration Strategy

**Phase 2.1: Create builtins/ package**
1. Create `internal/interp/builtins/` directory
2. Move `builtins_*.go` files → `builtins/*.go` (remove prefix)
3. Change package from `interp` to `builtins`
4. Capitalize exported functions
5. Move test files alongside implementation
6. Update imports in main `interp` package
7. Test thoroughly

**Phase 2.2: Create objects/ package**
1. Create `internal/interp/objects/` directory
2. Move `objects_*.go` files → `objects/*.go` (remove prefix)
3. Change package to `objects`
4. Export necessary functions
5. Move tests
6. Update imports
7. Test

**Phase 2.3: Create functions/ package**
1. Similar process for `functions_*.go` files

**Phase 2.4: Create statements/ package**
1. Similar process for `statements_*.go` files

**Example: Before and After**

**Before (current):**
```go
// internal/interp/builtins_math.go
package interp

func builtinAbs(args []Value) (Value, error) { ... }
```

**After (organized):**
```go
// internal/interp/builtins/math_basic.go
package builtins

// Abs returns the absolute value of a number
func Abs(args []Value) (Value, error) { ... }
```

```go
// internal/interp/interpreter.go
package interp

import "github.com/MeKo-Tech/go-dws/internal/interp/builtins"

func (i *Interpreter) evalBuiltinCall(name string, args []Value) (Value, error) {
    switch name {
    case "Abs":
        return builtins.Abs(args)
    // ...
    }
}
```

### Recommendation

**YES, proceed with subdirectory organization.** The benefits far outweigh the one-time migration cost. This is standard Go practice and will make the codebase much more maintainable long-term.

---

## Implementation Status

### ✅ Phase 1: Critical File Splits - COMPLETE

**Goal:** Split the largest, most unwieldy files

**Completed:**
1. ✅ Split `internal/interp/objects.go` (79KB → 4 files)
2. ✅ Split `internal/interp/functions.go` (67KB → 6 files)
3. ✅ Split `internal/interp/statements.go` (52KB → 4 files)
4. ✅ Split `builtins_datetime.go` (40KB → 3 files)
5. ✅ Split `builtins_math.go` (35KB → 3 files)

**Status:** All 5 priority file splits complete! All split files remain in flat directory structure (subdirectory organization deferred to Phase 2).

---

### ❌ Phase 2: Subdirectory Organization - NOT FEASIBLE

**Date Attempted:** 2025-11-11
**Status:** BLOCKED by Go's circular import restrictions

**Current State:**
- All files remain in flat directory structure
- File count: `internal/interp/` has 116 Go files in root directory
- Prefixed naming provides some organization

**Target Structure (Originally Planned):**
```
internal/interp/
├── builtins/          # Package builtins - all builtin functions
├── objects/           # Package objects - OOP features
├── functions/         # Package functions - function call handling
├── statements/        # Package statements - statement evaluation
└── ...                # Core interpreter files remain in root
```

**Why Phase 2 Was Blocked:**

Attempting to create subdirectory packages (builtins/, objects/, functions/, statements/) hits Go's **circular import prohibition**:

```
internal/interp → imports → internal/interp/builtins
internal/interp/builtins → imports → internal/interp (for Value, Interpreter types)
```

**The fundamental issue**: All interpreter subsystems (builtins, objects, functions, statements) are **tightly coupled** to core interpreter types:
- They all need access to `*Interpreter`, `Value` interface, and related types from `interp` package
- The `interp` package needs to call functions in these subsystems
- Many functions call back into interpreter methods (e.g., builtin `Map` calls `CallFunctionPointer`)

Go does **not allow circular imports**, even between parent/child packages. This is a hard constraint of the language.

**What Would Be Required to Succeed:**

To successfully separate into subdirectory packages would require a **much larger refactoring**:

1. **Extract shared types to common package** (e.g., `internal/runtime/` or `internal/values/`):
   - Move `Value` interface and all value types (IntegerValue, StringValue, etc.)
   - Move `Interpreter` to an interface or extract core methods
   - Move error handling infrastructure
   - Move Environment types

2. **Dependency flow** (no circular imports):
   ```
   internal/runtime/       # Shared types: Value, Interpreter interface
   ├── imported by →  internal/interp/builtins/
   ├── imported by →  internal/interp/objects/
   ├── imported by →  internal/interp/functions/
   ├── imported by →  internal/interp/statements/
   └── imported by →  internal/interp/          # Orchestrator
   ```

3. **Interface-based design**:
   - Define interfaces for what subsystems need from Interpreter
   - Reduce coupling through abstraction
   - Potentially thousands of lines of code affected

This is a **massive architectural refactoring** affecting the entire codebase, not just moving files.

**Decision: DEFER Phase 2**

The current flat structure with prefixed naming (`builtins_core.go`, `objects_methods.go`, etc.) is:
- ✅ **Working well** - Phase 1 splits reduced file sizes significantly
- ✅ **Maintainable** - Clear prefixes make organization visible
- ✅ **No circular deps** - Everything in same package
- ✅ **Easy to navigate** - Prefixes group related files together

The subdirectory organization would be **nice-to-have** but is **not critical** for maintainability given:
- Phase 1 successfully eliminated all files over 50KB
- Prefixed naming provides logical grouping
- Cost of the required refactoring is very high

**Recommendation:** Accept current structure and defer subdirectory organization until there's a compelling need and resources for the larger architectural refactoring

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
- **Files over 50KB in internal/interp (after Phase 1 splits):** 0 files ✅
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
⚠️ **Other large files** - vm.go (2,172 lines, 47KB), compiler.go (1,799 lines, 42KB), analyze_classes.go (1,466 lines, 38KB) remain unsplit

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

Phase 1 of the refactoring has been successfully completed. Phase 2 has been evaluated and deferred.

**Phase 1 Achievements:**
- ✅ Eliminated all files over 50KB in `internal/interp/`
- ✅ Split 5 major files into 20 smaller, focused files
- ✅ Clear naming conventions (objects_*, functions_*, statements_*, builtins_*)
- ✅ All functionality preserved and tests passing

**Phase 2 Status:**
After attempting implementation, **subdirectory organization has been deferred** due to Go's circular import restrictions:

1. ❌ **Phase 2 Attempted:** Created subdirectory packages but hit circular dependency
2. ❌ **Blocked by Go constraints:** Parent/child packages cannot have circular imports
3. ✅ **Code reverted:** All changes reverted, tests still passing
4. ✅ **Alternative accepted:** Current flat structure with prefixed naming is sufficient

**Why Phase 2 Was Deferred:**
- Requires massive architectural refactoring (extract shared types to common package)
- Would affect thousands of lines across entire codebase
- Current structure works well after Phase 1 improvements
- Cost/benefit analysis doesn't justify the effort

**Current Status:**
- **Flat directory structure** with 116 files in `internal/interp/`
- **Prefixed naming** provides logical organization (builtins_*, objects_*, functions_*, statements_*)
- **No files over 50KB** thanks to Phase 1 splits
- **Easy to navigate** - prefixes make it clear which files are related
- **No circular dependencies** - everything in one package

**Key Principles Moving Forward:**
- Accept flat structure as the pragmatic solution
- Continue using prefixed naming for new files
- Split files before they exceed 50KB
- Defer subdirectory organization until there's a compelling need

---

**Document Version:** 2.3
**Last Updated:** 2025-11-11
**Status:** Phase 1 complete and successful; Phase 2 attempted but deferred due to Go circular import constraints
