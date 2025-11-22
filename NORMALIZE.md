# Case Insensitivity Normalization Plan

This document outlines the migration plan to centralize all case-insensitive identifier handling using the `pkg/ident` package, eliminating scattered `strings.ToLower()` and `strings.EqualFold()` calls throughout the codebase.

## Current State

- **Centralized solution exists**: `pkg/ident` package with `Normalize()`, `Equal()`, `Compare()` functions
- **Adoption rate**: ~15% (16 files out of 108 files with case conversion)
- **Direct calls**: ~200+ `strings.ToLower()` and ~50+ `strings.EqualFold()` calls scattered across codebase
- **Known bugs**: 2 locations where original casing is lost in error messages
- **Assessment**: Poorly centralized (3/10)

## Goal

- **100% adoption** of `pkg/ident` for all identifier case-insensitive operations
- **Zero regression**: Original casing preserved in all user-facing messages
- **Maintainability**: Single source of truth for case handling
- **Prevention**: Linting rules to catch future violations

---

## Phase 1: Critical Bug Fixes (completed)

---

## Phase 2: Core Infrastructure (Priority: HIGH)

Enhance `pkg/ident` package and establish migration patterns.

### Tasks

- [x] **2.1** Review and enhance `pkg/ident` API if needed
  - Current functions: `Normalize()`, `Equal()`, `Compare()`, `Contains()`, `Index()`, `IsKeyword()`
  - Added: `HasPrefix()` and `HasSuffix()` for case-insensitive prefix/suffix matching
  - Documented best practices in `pkg/ident/doc.go` including:
    - Pattern 6: Prefix/Suffix Matching
    - "When NOT to Use This Package" section
    - Updated migration examples with HasPrefix

- [ ] **2.2** Create migration guide document
  - File: `docs/ident-migration-guide.md`
  - Include:
    - Before/after examples
    - Common patterns (map keys, comparisons, lookups)
    - How to preserve original casing in error messages
    - Anti-patterns to avoid

- [ ] **2.3** Add examples to `pkg/ident` package
  - Example: Map with case-insensitive keys but preserved original
  - Example: Symbol table pattern
  - Example: Error message pattern

- [ ] **2.4** Create helper for case-insensitive maps
  - Consider adding to `pkg/ident`:
    ```go
    type Map[T any] struct {
        store map[string]T
        originals map[string]string // optional: only for error reporting/use cases needing original casing
    }
    func (m *Map[T]) Set(key string, value T)
    func (m *Map[T]) Get(key string) (T, bool)
    func (m *Map[T]) GetOriginalKey(key string) string
    func (m *Map[T]) Delete(key string)
    func (m *Map[T]) Len() int
    func (m *Map[T]) Keys() []string
    func (m *Map[T]) Range(f func(key string, value T))
    // Consider thread safety: document single-threaded assumption, or add sync.Mutex if needed
    ```
  - API design checklist:
    - [ ] Thread safety: clarify single-threaded assumption or add concurrency protection
    - [ ] Memory overhead: make `originals` optional/specialized for error reporting
    - [ ] API completeness: add `Delete`, `Len`, `Keys`, `Range` methods
    - [ ] Naming: use `GetOriginalKey` for clarity
  - Evaluate if this reduces boilerplate in migration

---

## Phase 3: Token & Lexer (Priority: HIGH)

Already well-centralized, but verify completeness.

### Tasks

- [ ] **3.1** Audit `pkg/token/token.go`
  - Verify `LookupIdent()` is only case conversion point
  - Confirm all keyword lookups use this function
  - ✅ Status: Already centralized

- [ ] **3.2** Audit `internal/lexer/lexer.go`
  - Verify uses `token.LookupIdent()` consistently
  - Line 1013 confirmed, check for any direct `strings.ToLower()`
  - ✅ Status: Already centralized

- [ ] **3.3** Add tests for keyword case insensitivity
  - Test all keywords in UPPER, lower, MixedCase
  - Verify tokens are correctly identified
  - Verify original casing in error messages

---

## Phase 4: Symbol Table & Type Registry (Priority: HIGH)

Migrate core lookup infrastructure to `pkg/ident`.

### Tasks

- [ ] **4.1** Migrate `internal/semantic/symbol_table.go`
  - Lines to change: 52, 62, 72, 83, 105, 529, 551, 567
  - Replace: `strings.ToLower(name)` → `ident.Normalize(name)`
  - Pattern already correct: Stores original in `Symbol.Name` field ✅
  - Run: `go test ./internal/semantic/... -v`

- [ ] **4.2** Migrate `internal/semantic/type_registry.go`
  - Lines to change: 85, 110, 122, 308, 321
  - Replace: `strings.ToLower(name)` → `ident.Normalize(name)`
  - Verify original casing preserved in descriptors
  - Run: `go test ./internal/semantic/... -v`

- [ ] **4.3** Add explicit tests for case insensitivity
  - Test symbol table: Define `MyVar`, lookup `myvar`, `MYVAR`, `MyVar`
  - Test type registry: Register `MyType`, resolve `mytype`, `MYTYPE`
  - Verify error messages use original casing

---

## Phase 5: Type System (Priority: HIGH)

Migrate `internal/types/` package with ~15+ direct calls.

### Tasks

- [ ] **5.1** Migrate `internal/types/types.go` - Part 1: ClassType methods
  - Lines: 502, 514, 530, 536, 542, 555, 590, 594
  - Methods: `HasMethod()`, `GetMethod()`, `GetMethodOverloads()`, `HasField()`, `GetField()`, etc.
  - Replace: `strings.ToLower()` → `ident.Normalize()`
  - Verify: Original method/field names preserved in type metadata

- [ ] **5.2** Migrate `internal/types/types.go` - Part 2: RecordType methods
  - Similar methods for RecordType
  - Replace: `strings.ToLower()` → `ident.Normalize()`

- [ ] **5.3** Migrate `internal/types/types.go` - Part 3: InterfaceType methods
  - Interface method lookups
  - Replace: `strings.ToLower()` → `ident.Normalize()`

- [ ] **5.4** Migrate `internal/types/compound_types.go`
  - Search for: `strings.ToLower()`
  - Replace with: `ident.Normalize()`

- [ ] **5.5** Migrate `internal/types/type_utils.go`
  - Function: `TypeFromString()` at line 351
  - Replace: `strings.ToLower()` → `ident.Normalize()`

- [ ] **5.6** Run comprehensive type system tests
  - Run: `go test ./internal/types/... -v`
  - Verify: All case-insensitive lookups work
  - Verify: Error messages preserve original casing

---

## Phase 6: Semantic Analyzer (Priority: HIGH)

Migrate remaining semantic analyzer files (partially adopted).

### Tasks

- [ ] **6.1** Migrate `internal/semantic/analyze_classes.go`
  - Lines: 199, 378, 439 (plus Phase 1 fixes)
  - Replace: `strings.ToLower()` → `ident.Normalize()`
  - Already uses `ident` in some places - make consistent

- [ ] **6.2** Migrate `internal/semantic/analyze_function_calls.go`
  - Line: 956 and others
  - Replace: `strings.ToLower()` → `ident.Normalize()`

- [ ] **6.3** Migrate `internal/semantic/analyze_records.go`
  - Line: 225 and others
  - Replace: `strings.ToLower()` → `ident.Normalize()`

- [ ] **6.4** Migrate `internal/semantic/type_resolution.go`
  - Lines: 97, 455, 809, 854
  - Replace: `strings.ToLower()` → `ident.Normalize()`

- [ ] **6.5** Migrate `internal/semantic/unit_analyzer.go`
  - Lines: 57, 103, 134, 192, 201
  - Replace: `strings.ToLower()` → `ident.Normalize()`

- [ ] **6.6** Verify consistency across semantic analyzer
  - Files already using `ident`: analyze_classes_decl.go, analyze_expr_operators.go, etc.
  - Ensure all semantic analyzer files use same approach
  - Run: `go test ./internal/semantic/... -v`

---

## Phase 7: Interpreter (Priority: MEDIUM)

Migrate interpreter, which has partial adoption.

### Tasks

- [ ] **7.1** ✅ SKIP: `internal/interp/environment.go`
  - Already uses `ident.Normalize()` at lines 66, 91, 118, 134
  - Verify: No additional migrations needed

- [ ] **7.2** ✅ SKIP: `internal/interp/builtins/registry.go`
  - Already uses `ident.Normalize()` at lines 137, 150, 165
  - Verify: No additional migrations needed

- [ ] **7.3** Migrate `internal/interp/interpreter.go`
  - ~20+ locations: lines 300, 312, 322, 329, 341, 351, 358, 370, 380, 389, 404, 425, 511, 580, 588, 596
  - Replace: `strings.ToLower()` → `ident.Normalize()`
  - Focus: Class, record, enum name lookups

- [ ] **7.4** Migrate `internal/interp/class.go`
  - Line: 157 - `lookupMethod()` function
  - Replace: `strings.ToLower()` → `ident.Normalize()`

- [ ] **7.5** Migrate `internal/interp/interface.go`
  - Line: 38 - `GetMethod()` function
  - Replace: `strings.ToLower()` → `ident.Normalize()`

- [ ] **7.6** Migrate `internal/interp/objects_hierarchy.go`
  - Line: 130 (covered in Phase 1 bug fix)
  - Verify: No additional migrations needed after Phase 1

- [ ] **7.7** Run interpreter tests
  - Run: `go test ./internal/interp/... -v`
  - Run: `go test -v ./internal/interp -run TestDWScriptFixtures`
  - Verify: No regressions in fixture tests

---

## Phase 8: Bytecode Compiler & VM (Priority: MEDIUM)

Migrate bytecode compiler, which uses `strings.EqualFold()`.

### Tasks

- [ ] **8.1** Migrate `internal/bytecode/compiler_core.go`
  - Line 271: `resolveLocal()` - uses `strings.EqualFold()`
  - Line 284: `resolveUpvalue()` - uses `strings.EqualFold()`
  - Line 295: `resolveGlobal()` - uses `strings.ToLower()`
  - Replace with: `ident.Equal()` or `ident.Normalize()`
  - Decision: Use `ident.Equal()` for comparisons (no allocation)

- [ ] **8.2** Migrate identifier checks in compiler
  - Lines: 80, 96, 117 - uses `strings.EqualFold()`
  - Replace with: `ident.Equal()`

- [ ] **8.3** Audit other bytecode files
  - Search: `strings.ToLower`, `strings.EqualFold` in `internal/bytecode/`
  - Files to check: `compiler.go`, `compiler_expr.go`, `compiler_stmt.go`, etc.
  - Migrate all occurrences

- [ ] **8.4** Run bytecode tests
  - Run: `go test ./internal/bytecode/... -v`
  - Test: Compile and run with mixed-case identifiers
  - Verify: Variable resolution works case-insensitively

---

## Phase 9: Parser (Priority: LOW)

Parser uses `strings.EqualFold()` for keyword checks - low priority.

### Tasks

- [ ] **9.1** Audit `internal/parser/` for case conversion
  - Files: `functions.go`, `operators.go`, `helpers.go`
  - ~5 files with `strings.EqualFold()` for keyword comparisons

- [ ] **9.2** Evaluate migration necessity
  - Decision: Parser keyword checks via `strings.EqualFold()` may be acceptable
  - Alternative: Migrate to `ident.Equal()` for consistency
  - Consider: Performance impact (minimal in parser)

- [ ] **9.3** Migrate parser if decided
  - Replace: `strings.EqualFold()` → `ident.Equal()`
  - Run: `go test ./internal/parser/... -v`

---

## Phase 10: Registries (Priority: LOW)

Already mostly migrated, verify completeness.

### Tasks

- [ ] **10.1** ✅ SKIP: `internal/units/registry.go`
  - Already uses `ident.Normalize()`
  - Verify: No additional migrations needed

- [ ] **10.2** ✅ SKIP: `internal/interp/types/class_registry.go`
  - Already uses `ident.Normalize()`
  - Verify: No additional migrations needed

- [ ] **10.3** ✅ SKIP: `internal/interp/types/function_registry.go`
  - Already uses `ident.Normalize()`
  - Verify: No additional migrations needed

---

## Phase 11: Comprehensive Testing (Priority: HIGH)

Ensure no regressions and verify original casing preservation.

### Tasks

- [ ] **11.1** Create comprehensive case insensitivity test suite
  - File: `internal/semantic/case_insensitivity_test.go`
  - Test all identifier types: variables, types, methods, fields, properties
  - Test variations: lowercase, UPPERCASE, MixedCase, camelCase, PascalCase

- [ ] **11.2** Create error message casing test suite
  - File: `internal/semantic/error_message_casing_test.go`
  - Verify: All error messages preserve user's original casing
  - Test: Non-existent symbols, type mismatches, member access errors

- [ ] **11.3** Run full test suite
  - Run: `just test` (all tests)
  - Run: `just test-coverage` (with coverage report)
  - Target: Maintain >90% coverage in migrated packages

- [ ] **11.4** Run fixture tests
  - Run: `go test -v ./internal/interp -run TestDWScriptFixtures`
  - Verify: No regressions in ~2,100 DWScript test cases
  - Document: Any fixture test changes in `testdata/fixtures/TEST_STATUS.md`

- [ ] **11.5** Manual testing with CLI
  - Test: `just lex`, `just parse`, `just run` with mixed-case identifiers
  - Verify: Error messages show original casing
  - Example test file: `testdata/case-insensitivity-test.dws`

---

## Phase 12: Prevention & Tooling (Priority: MEDIUM)

Prevent future regressions with linting and documentation.

### Tasks

- [ ] **12.1** Add golangci-lint custom rule
  - File: `.golangci.yml`
  - Add `forbidigo` or custom linter to forbid:
    - `strings.ToLower()` in identifier handling code
    - `strings.EqualFold()` in identifier handling code
  - Allowed: `pkg/ident.Normalize()`, `pkg/ident.Equal()`

- [ ] **12.2** Create linter exceptions for legitimate uses
  - Allow: `strings.ToLower()` in non-identifier contexts (e.g., URL handling, file extensions)
  - Document: When direct `strings.ToLower()` is acceptable vs when to use `ident`

- [ ] **12.3** Add pre-commit hook (optional)
  - File: `.git/hooks/pre-commit`
  - Check: Forbid `strings.ToLower()` and `strings.EqualFold()` in certain directories
  - Suggest: Use `pkg/ident` instead

- [ ] **12.4** Update CONTRIBUTING.md
  - Section: "Case Insensitivity Guidelines"
  - Rule: "Always use `pkg/ident` for identifier handling"
  - Examples: Before/after migration patterns
  - Link: Migration guide in `docs/ident-migration-guide.md`

---

## Phase 13: Documentation (Priority: LOW)

Update project documentation to reflect centralized approach.

### Tasks

- [ ] **13.1** Update CLAUDE.md
  - Section: "DWScript Language Specifics > Case Insensitivity"
  - Add: "All case-insensitive identifier operations use `pkg/ident` package"
  - Document: `ident.Normalize()` for keys, `ident.Equal()` for comparisons

- [ ] **13.2** Update README.md (if applicable)
  - Mention: Case-insensitive language handling
  - Link: To `pkg/ident` package documentation

- [ ] **13.3** Add godoc to `pkg/ident`
  - Expand: Package-level documentation
  - Examples: Common patterns for embedding projects
  - Best practices: When to use each function

- [ ] **13.4** Create architecture decision record
  - File: `docs/adr/003-case-insensitivity-centralization.md`
  - Note: Create the `docs/adr/` directory first if it does not exist.
  - Document: Why `pkg/ident`, alternatives considered, migration rationale

---

## Phase 14: Cleanup & Finalization (Priority: LOW)

Final cleanup and verification.

### Tasks

- [ ] **14.1** Search for any remaining direct case conversion
  - Command: `grep -r "strings.ToLower" --include="*.go" | grep -v "pkg/ident"`
  - Command: `grep -r "strings.EqualFold" --include="*.go" | grep -v "pkg/ident"`
  - Investigate: Each occurrence for legitimacy

- [ ] **14.2** Remove unused imports
  - Run: `goimports -w .`
  - Run: `golangci-lint run --fix`

- [ ] **14.3** Update migration progress tracking
  - Update: This file (NORMALIZE.md) with completion status
  - Mark: All phases complete
  - Statistics: Before/after adoption percentage

- [ ] **14.4** Final test run
  - Run: `just ci` (full CI suite)
  - Run: `just test-coverage` (verify coverage maintained)
  - Run: Fixture tests (verify DWScript compatibility)

- [ ] **14.5** Create summary report
  - Document: Migration statistics
  - Files changed: ~108 files
  - Lines changed: ~250+ locations
  - Adoption: Before 15% → After 100%
  - Bugs fixed: 2 error message casing issues

---

## Success Criteria

- [ ] **Zero direct `strings.ToLower()` calls** for identifier handling (exceptions documented)
- [ ] **Zero direct `strings.EqualFold()` calls** for identifier handling (exceptions documented)
- [ ] **100% adoption** of `pkg/ident` in identifier-related code
- [ ] **All error messages** preserve original user casing
- [ ] **All tests pass** (unit, integration, fixture tests)
- [ ] **Coverage maintained** at >80% overall, >90% for core packages
- [ ] **Linting rules** prevent future regressions
- [ ] **Documentation complete** (migration guide, ADR, CONTRIBUTING.md)

---

## Rollback Plan

If critical issues arise during migration:

1. **Git branches**: Each phase commits to feature branch before merging
2. **Revert commits**: Use `git revert` for problematic phases
3. **Phase independence**: Each phase is independently revertable
4. **Test gates**: Each phase requires passing tests before proceeding

---

## Estimated Effort

| Phase | Files | Lines | Estimated Time | Priority |
|-------|-------|-------|----------------|----------|
| Phase 1: Bug Fixes | 2 | ~5 | 1 hour | HIGH |
| Phase 2: Infrastructure | 3 | ~50 | 2 hours | HIGH |
| Phase 3: Token & Lexer | 2 | 0 (verify) | 30 min | HIGH |
| Phase 4: Symbol/Type Registry | 2 | ~10 | 1 hour | HIGH |
| Phase 5: Type System | 4 | ~30 | 2 hours | HIGH |
| Phase 6: Semantic Analyzer | 5 | ~40 | 3 hours | HIGH |
| Phase 7: Interpreter | 5 | ~40 | 3 hours | MEDIUM |
| Phase 8: Bytecode | 3 | ~20 | 2 hours | MEDIUM |
| Phase 9: Parser | 5 | ~10 | 1 hour | LOW |
| Phase 10: Registries | 3 | 0 (verify) | 30 min | LOW |
| Phase 11: Testing | N/A | ~200 | 4 hours | HIGH |
| Phase 12: Prevention | 2 | ~30 | 2 hours | MEDIUM |
| Phase 13: Documentation | 4 | ~100 | 2 hours | LOW |
| Phase 14: Cleanup | N/A | N/A | 1 hour | LOW |
| **Total** | **~108** | **~250+** | **~25 hours** | |

---

## Notes

- **Backward compatibility**: This is an internal refactoring; no API changes expected
- **Performance**: `ident.Equal()` uses `strings.EqualFold()` (same performance), `ident.Normalize()` uses `strings.ToLower()` (same performance)
- **Safety**: All changes covered by existing tests; each phase runs full test suite
- **Incrementality**: Can pause between phases; each phase is independently useful

---

## Related Files

- `pkg/ident/ident.go` - Centralized case handling
- `docs/ident-migration-guide.md` - Migration guide (to be created in Phase 2)
- `CONTRIBUTING.md` - Contribution guidelines (to be updated in Phase 12)
- `.golangci.yml` - Linter configuration (to be updated in Phase 12)

---

## Tracking

**Created**: 2025-11-22
**Status**: Planning
**Owner**: TBD
**Target Completion**: TBD

---

*This is a living document. Update task checkboxes as phases complete.*
