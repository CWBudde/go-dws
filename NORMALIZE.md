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

## Phase 2: Core Infrastructure (completed)

Enhanced `pkg/ident` package and established migration patterns.

- [x] **2.1** Added `HasPrefix()` and `HasSuffix()` functions; enhanced `doc.go` with patterns and best practices
- [x] **2.2** Created `docs/ident-migration-guide.md` with before/after examples, patterns, and anti-patterns
- [x] **2.3** Added examples: `Example_errorMessages`, `Example_typeRegistry`
- [x] **2.4** Created generic `Map[T]` type in `pkg/ident/map.go` with full API (Set, Get, GetOriginalKey, Delete, Len, Keys, Range, Clone, etc.)

---

## Phase 3: Token & Lexer (Priority: HIGH) ✅ COMPLETED

Already well-centralized, verified completeness.

### Tasks

- [x] **3.1** Audit `pkg/token/token.go`
  - Verified `LookupIdent()` is the only case conversion point (line 738)
  - Confirmed all keyword lookups use this function
  - ✅ Status: Already centralized

- [x] **3.2** Audit `internal/lexer/lexer.go`
  - Verified uses `token.LookupIdent()` via alias (line 1013)
  - No direct `strings.ToLower()` in main lexer code
  - Only test files have `strings.ToLower()` for error message verification
  - ✅ Status: Already centralized

- [x] **3.3** Add tests for keyword case insensitivity
  - Created `pkg/token/case_insensitivity_test.go`:
    - `TestAllKeywordsCaseInsensitivity`: Tests ALL keywords in lowercase, UPPERCASE, MixedCase, aLtErNaTiNg
    - `TestOriginalCasingPreservedInTokenLiteral`: Verifies original casing preserved in token literals
    - `TestKeywordIdentifierBoundary`: Tests edge cases between keywords and identifiers
    - `TestIsKeywordCaseInsensitive`: Verifies IsKeyword function is case-insensitive
    - `TestGetKeywordLiteralCaseInsensitive`: Verifies canonical form returned
    - `TestTokenTypeConsistency`: Verifies consistent types across case variations
  - Created `internal/lexer/case_insensitivity_test.go`:
    - `TestLexerKeywordCaseInsensitivity`: Tests keywords through full lexer pipeline
    - `TestLexerPreservesOriginalCasing`: Verifies lexer preserves original casing in literals
    - `TestLexerMixedCaseProgram`: Tests realistic program with mixed case keywords
    - `TestLexerKeywordIdentifierBoundary`: Tests boundary between keywords and identifiers
    - `TestLexerMultipleKeywordsSameProgram`: Tests multiple case variations in one program

---

## Phase 4: Symbol Table & Type Registry (Priority: HIGH) ✅ COMPLETED

Migrated core lookup infrastructure to `pkg/ident`.

### Tasks

- [x] **4.1** Migrate `internal/semantic/symbol_table.go`
  - Replaced 8 occurrences of `strings.ToLower(name)` → `ident.Normalize(name)`
  - Pattern already correct: Stores original in `Symbol.Name` field ✅
  - All tests pass

- [x] **4.2** Migrate `internal/semantic/type_registry.go`
  - Replaced 5 occurrences of `strings.ToLower(name)` → `ident.Normalize(name)`
  - Original casing preserved in `TypeDescriptor.Name` ✅
  - All tests pass

- [x] **4.3** Add explicit tests for case insensitivity
  - Added `TestSymbolTableOriginalCasingPreserved`: Tests lowercase, UPPERCASE, PascalCase, camelCase definitions
  - Added `TestTypeRegistryOriginalCasingPreserved`: Tests various casing with descriptor lookup
  - Added `TestCaseInsensitiveTypeAliases`: Tests type aliases with different case access
  - All tests verify original casing is preserved for error messages

---

## Phase 5: Type System (Priority: HIGH)

Migrate `internal/types/` package with ~15+ direct calls.

### Tasks

- [x] **5.1** Migrate `internal/types/types.go` - Part 1: ClassType methods
  - Migrated methods: `HasField()`, `GetField()`, `HasMethod()`, `GetMethod()`,
    `GetMethodOverloads()`, `GetConstructorOverloads()`, `AddMethodOverload()`,
    `AddConstructorOverload()`, `HasConstructor()`, `GetConstructor()`,
    `HasProperty()`, `GetProperty()`, `GetClassVar()`
  - Replaced: `strings.ToLower()` → `ident.Normalize()`
  - Replaced: `strings.EqualFold()` → `ident.Equal()`
  - Verified: Original method/field names preserved in type metadata

- [x] **5.2** Migrate RecordType methods in `internal/types/compound_types.go`
  - Migrated methods: `HasField()`, `GetFieldType()`, `HasMethod()`, `GetMethod()`,
    `HasClassMethod()`, `GetClassMethod()`, `HasProperty()`, `GetProperty()`,
    `GetMethodOverloads()`, `GetClassMethodOverloads()`
  - Used `ident.Equal()` for case-insensitive iteration
  - Note: RecordType is in compound_types.go, not types.go

- [x] **5.3** Migrate InterfaceType methods in `internal/types/types.go`
  - Migrated methods: `HasMethod()`, `GetMethod()`
  - Used `ident.Equal()` for case-insensitive iteration

- [x] **5.4** Migrate HelperType methods in `internal/types/compound_types.go`
  - Migrated methods: `GetMethod()`, `GetProperty()`, `GetClassVar()`, `GetClassConst()`
  - Used `ident.Normalize()` for case-insensitive lookup

- [x] **5.5** Migrate `TypeFromString()` in `internal/types/types.go`
  - Replaced: `strings.ToLower()` → `ident.Normalize()`
  - Removed unused `strings` import from types.go

- [x] **5.6** Run comprehensive type system tests
  - Ran: `go test ./internal/types/... -v` - All tests pass
  - Verified: All case-insensitive lookups work correctly
  - Note: Pre-existing interface test failures in internal/interp are unrelated

---

## Phase 6: Semantic Analyzer (Priority: HIGH) ✅ COMPLETED

Migrate remaining semantic analyzer files (partially adopted).

### Tasks

- [x] **6.1** Migrate `internal/semantic/analyze_classes.go`
  - Lines: 199, 378, 439 (plus Phase 1 fixes)
  - Replace: `strings.ToLower()` → `ident.Normalize()`
  - Already uses `ident` in some places - make consistent
  - Migrated 3 occurrences, removed unused `strings` import

- [x] **6.2** Migrate `internal/semantic/analyze_function_calls.go`
  - Line: 956 and others
  - Replace: `strings.ToLower()` → `ident.Normalize()`
  - Migrated 3 occurrences, removed unused `strings` import

- [x] **6.3** Migrate `internal/semantic/analyze_records.go`
  - Line: 225 and others
  - Replace: `strings.ToLower()` → `ident.Normalize()`
  - Migrated 6 occurrences, removed unused `strings` import

- [x] **6.4** Migrate `internal/semantic/type_resolution.go`
  - Lines: 97, 455, 809, 854
  - Replace: `strings.ToLower()` → `ident.Normalize()`
  - Migrated 12 occurrences (kept `strings` for HasPrefix/TrimSpace/etc.)

- [x] **6.5** Migrate `internal/semantic/unit_analyzer.go`
  - Lines: 57, 103, 134, 192, 201
  - Replace: `strings.ToLower()` → `ident.Normalize()`
  - Migrated 5 occurrences, removed unused `strings` import

- [x] **6.6** Verify consistency across semantic analyzer
  - Files already using `ident`: analyze_classes_decl.go, analyze_expr_operators.go, etc.
  - Ensure all semantic analyzer files use same approach
  - Run: `go test ./internal/semantic/... -v`
  - All tests pass
  - Note: Additional files in semantic package still have `strings.ToLower()` - can be addressed in future phases

---

## Phase 7: Interpreter (Priority: MEDIUM)

Migrate interpreter, which has partial adoption.

### Tasks

- [x] **7.1** ✅ VERIFIED: `internal/interp/environment.go`
  - Already uses `ident.Normalize()` at lines 66, 91, 118, 134
  - Verified: No `strings.ToLower()` or `strings.EqualFold()` present
  - Status: Fully migrated, no additional changes needed

- [x] **7.2** ✅ VERIFIED: `internal/interp/builtins/registry.go`
  - Already uses `ident.Normalize()` at lines 90, 137, 150, 165, 234
  - Verified: No `strings.ToLower()` or `strings.EqualFold()` present
  - Status: Fully migrated, no additional changes needed

- [x] **7.3** Migrate `internal/interp/interpreter.go`
  - Migrated 51 occurrences of `strings.ToLower()` → `ident.Normalize()`
  - Migrated 2 occurrences of `strings.EqualFold()` → `ident.Equal()`
  - Migrated 1 occurrence of `strings.HasPrefix()` → `ident.HasPrefix()`
  - Removed unused `strings` import
  - All tests pass (pre-existing interface test failures unchanged)

- [x] **7.4** Migrate `internal/interp/class.go`
  - Migrated 2 occurrences of `strings.ToLower()` → `ident.Normalize()`
  - Migrated 5 occurrences of `strings.EqualFold()` → `ident.Equal()`
  - Removed unused `strings` import

- [x] **7.5** Migrate `internal/interp/interface.go`
  - Migrated 3 occurrences of `strings.ToLower()` → `ident.Normalize()`
  - Migrated 1 occurrence of `strings.EqualFold()` → `ident.Equal()`
  - Removed unused `strings` import

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
