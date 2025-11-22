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

- [x] **7.6** Migrate `internal/interp/objects_hierarchy.go`
  - Migrated 12 occurrences of `strings.ToLower()` → `pkgident.Normalize()`
  - Migrated 13 occurrences of `strings.EqualFold()` → `pkgident.Equal()`
  - Used `pkgident` alias to avoid conflict with local `ident` variable
  - Removed unused `strings` import

- [x] **7.7** Run interpreter tests
  - Ran: `go test ./internal/interp/... -count=1`
  - Results: All sub-packages pass (builtins, errors, evaluator, runtime, types)
  - Verified: No regressions from migration (same 26 passed, 6 failed as before)
  - Note: Pre-existing interface test failures are unrelated to this migration

---

## Phase 8: Bytecode Compiler & VM (Priority: MEDIUM) ✅ COMPLETED

Migrate bytecode compiler, which uses `strings.EqualFold()`.

### Tasks

- [x] **8.1** Migrate `internal/bytecode/compiler_core.go`
  - Migrated 8 occurrences:
    - `declareGlobal()`: `strings.ToLower()` → `pkgident.Normalize()`
    - `resolveLocal()`: `strings.EqualFold()` → `pkgident.Equal()`
    - `resolveLocalInCurrentScope()`: `strings.EqualFold()` → `pkgident.Equal()`
    - `resolveGlobal()`: `strings.ToLower()` → `pkgident.Normalize()`
    - `addBuiltinGlobal()`: `strings.ToLower()` → `pkgident.Normalize()`
    - `inferExpressionType()`: `strings.ToLower()` → `pkgident.Normalize()`
    - `typeFromAnnotation()`: `strings.ToLower()` → `pkgident.Normalize()`
    - `evaluateBinary()`: `strings.ToLower()` → `pkgident.Normalize()`
    - `evaluateUnary()`: `strings.ToLower()` → `pkgident.Normalize()`
  - Used `pkgident` alias to avoid shadowing with local `ident` parameter
  - Removed unused `strings` import

- [x] **8.2** Migrate `internal/bytecode/compiler_statements.go`
  - Migrated 8 occurrences in function/helper/record/class declarations
  - Replaced: `strings.ToLower()` → `ident.Normalize()`
  - Removed unused `strings` import

- [x] **8.3** Migrate `internal/bytecode/bytecode.go`
  - Migrated 7 occurrences in ObjectInstance and RecordInstance methods:
    - `GetField()`, `SetField()`, `GetProperty()`, `SetProperty()`
  - Replaced: `strings.ToLower()` → `ident.Normalize()`
  - Kept `strings` import for `strings.Builder` and `strings.IndexByte` usage

- [x] **8.4** Migrate `internal/bytecode/vm_exec.go`
  - Migrated 2 occurrences:
    - Class metadata lookup in OpNewObject handler
    - `resolveValueType()` function
  - Replaced: `strings.ToLower()` → `ident.Normalize()`
  - Removed unused `strings` import

- [x] **8.5** Migrate `internal/bytecode/compiler_functions.go`
  - Migrated 3 occurrences:
    - `compileCallExpression()`: builtin value creation
    - `resolveDirectFunction()`: function lookup
    - `isBuiltinFunction()`: builtin name normalization
  - Used `pkgident` alias to avoid shadowing with local `ident` parameter
  - Removed unused `strings` import

- [x] **8.6** Run bytecode tests
  - Ran: `go test ./internal/bytecode/... -count=1`
  - Results: All tests pass
  - Verified: No regressions from migration

---

## Phase 9: Parser (Priority: LOW) ✅ COMPLETED

Parser uses `strings.ToLower()` for keyword checks - migrated for consistency.

### Tasks

- [x] **9.1** Audit `internal/parser/` for case conversion
  - Found 3 production code occurrences (functions.go, operators.go)
  - Found 3 test code occurrences (error_recovery_test.go) - kept as-is
  - No `strings.EqualFold()` calls found (contrary to initial assessment)

- [x] **9.2** Evaluate migration necessity
  - Decision: Migrate to `pkg/ident` for consistency across the codebase
  - All production uses are for identifier/keyword case-insensitive handling
  - Performance impact: negligible (same underlying functions)

- [x] **9.3** Migrate parser if decided
  - Migrated `functions.go`:
    - `isCallingConvention()`: `strings.ToLower()` → `ident.Equal()` (7 comparisons)
    - `fn.CallingConvention`: `strings.ToLower()` → `ident.Normalize()`
  - Migrated `operators.go`:
    - `normalizeOperatorSymbol()`: `strings.ToLower()` → `ident.Normalize()`
  - Kept `strings` import in `operators.go` for `strings.Join()` (type name concatenation)
  - Ran: `go test ./internal/parser/... -v` - All tests pass

---

## Phase 10: Registries (Priority: LOW) ✅ COMPLETED

Verified completeness and migrated one remaining occurrence.

### Tasks

- [x] **10.1** Migrate `internal/units/registry.go`
  - Already uses `ident.Normalize()` for all map key operations
  - Migrated 1 occurrence: `strings.EqualFold()` → `ident.Equal()` (line 184)
  - Kept `strings` import for `strings.Join()` usage (legitimate non-identifier use)
  - All tests pass

- [x] **10.2** ✅ VERIFIED: `internal/interp/types/class_registry.go`
  - Already uses `ident.Normalize()` throughout
  - No `strings.ToLower()` or `strings.EqualFold()` present
  - Status: Fully migrated, no changes needed

- [x] **10.3** ✅ VERIFIED: `internal/interp/types/function_registry.go`
  - Already uses `ident.Normalize()` throughout
  - Only `strings.Split()` used (legitimate parsing use)
  - Status: Fully migrated, no changes needed

---

## Phase 11: Comprehensive Testing (Priority: HIGH) ✅ COMPLETED

Comprehensive test suite created and all tests verified.

### Tasks

- [x] **11.1** Create comprehensive case insensitivity test suite
  - Extended `internal/semantic/case_insensitive_test.go` with new tests:
    - `TestCaseInsensitiveRecordFields`: Record field access with different casing
    - `TestCaseInsensitiveEnumValues`: Enum value access with different casing
    - `TestCaseInsensitiveInterfaceMethods`: Interface method case insensitivity
    - `TestCaseInsensitiveProperties`: Property access with different casing
    - `TestCaseInsensitiveMethodCalls`: Method calls with different casing
    - `TestCaseInsensitiveBuiltinTypes`: Built-in type names (Integer, STRING, etc.)
    - `TestCaseInsensitiveKeywords`: Keywords (BEGIN, if, Then, etc.)
  - All tests pass

- [x] **11.2** Create error message casing test suite
  - Created `internal/semantic/error_message_casing_test.go` with tests:
    - `TestUndefinedVariablePreservesCase`: Verifies undefined variables preserve casing
    - `TestUndefinedTypePreservesCase`: Verifies undefined types preserve casing
    - `TestUndefinedFunctionPreservesCase`: Verifies undefined functions preserve casing
    - `TestUndefinedMemberPreservesCase`: Tests class/record member errors
    - `TestTypeMismatchPreservesCase`: Type mismatch error casing
    - `TestDuplicateDeclarationPreservesOriginalCase`: Duplicate declaration errors
    - `TestEnumValueErrorPreservesCase`: Enum error messages
  - Known issues documented:
    - Record field errors don't preserve original casing (shows lowercase)
    - Duplicate declaration errors show normalized casing, not original

- [x] **11.3** Run full test suite
  - Ran: `go test ./... -count=1`
  - Results: All packages pass except pre-existing interface test failures in internal/interp (26 passed, 6 failed - documented in NORMALIZE.md Phase 7)
  - No new regressions introduced

- [x] **11.4** Run fixture tests
  - Ran: `go test -v ./internal/interp -run TestDWScriptFixtures`
  - Results: 283 passed, 918 failed, 49 skipped (pre-existing status)
  - No regressions from case-insensitivity migration

- [x] **11.5** Manual testing with CLI
  - Created: `testdata/case-insensitivity-test.dws`
  - Tested: Variable access, function calls, record fields, enum values, keywords
  - Verified: All mixed-case identifiers work correctly
  - Verified: Error messages preserve original casing (e.g., "Undefined variable 'MyUndefinedVar'")

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

- [x] **14.1** Search for any remaining direct case conversion
  - Command: `grep -r "strings.ToLower" --include="*.go" | grep -v "pkg/ident"`
  - Command: `grep -r "strings.EqualFold" --include="*.go" | grep -v "pkg/ident"`
  - Investigate: Each occurrence for legitimacy
  - **Results**: Found ~160 `strings.ToLower` and ~55 `strings.EqualFold` occurrences
  - See subtasks below for categorized migration plan

#### Legitimate Exceptions (DO NOT MIGRATE)

The following uses of `strings.ToLower()`/`strings.EqualFold()` are legitimate and should NOT be migrated:

1. **pkg/token/token.go** (3 occurrences) - Foundation of case-insensitive keyword lookup
   - `LookupIdent()`, `IsKeyword()`, `GetKeywordLiteral()` - canonical implementation

2. **pkg/ident/*.go** - Implementation of the ident package itself

3. **Builtin string functions** (actual DWScript runtime functions):
   - `internal/bytecode/vm_builtins_string.go:1052,1067-1068,1111-1112,1304-1305,1387` - SameText, CompareText, AnsiCompareText, AnsiPos, LowerCase
   - `internal/interp/builtins_strings_basic.go:109,735` - LowerCase, StrToBool parsing
   - `internal/interp/builtins_strings_compare.go:35` - SameText
   - `internal/interp/builtins/strings_compare.go:41` - SameText
   - `internal/interp/helpers_conversion.go:276` - LowerCase implementation
   - `internal/interp/runtime/helpers.go:348` - StringEqualsInsensitive runtime helper

4. **Test files** (error message checking, not identifier handling):
   - `cmd/dwscript/math_functions_test.go:140`
   - `cmd/dwscript/exception_cli_test.go:248`
   - `internal/parser/error_recovery_test.go:751,1071,1087`
   - `internal/lexer/lexer_errors_test.go:295,383`
   - `internal/lexer/token_test.go:412`
   - `internal/semantic/analyze_builtin_convert_coverage_test.go:32`
   - `internal/semantic/control_flow_test.go:595`
   - `internal/semantic/analyzer_test.go:65,66,79`
   - `internal/semantic/subrange_test.go:79`
   - `internal/interp/interpreter_basic_test.go:511`
   - `internal/interp/type_assertion_test.go:76`
   - `internal/interp/operator_test.go:96`
   - `internal/interp/interface_integration_test.go:296,429`
   - `internal/interp/interface_edge_test.go:92,132,152`
   - `internal/bytecode/compiler_coverage_test.go:763,764`
   - `pkg/dwscript/error_format_test.go:78`

5. **FFI/Examples** (actual function registration):
   - `examples/ffi/main.go:86` - example code
   - `pkg/dwscript/ffi_integration_test.go:22,127` - FFI tests

6. **Non-identifier uses**:
   - `cmd/dwscript/cmd/fmt.go:109` - format style command-line option
   - `internal/bytecode/vm_builtins_conversion.go:214` - StrToBool user input parsing

#### Migration Subtasks

- [x] **14.1.1** Migrate `internal/units/` package (5 occurrences)
  - `unit.go:78,84,86` - unit name normalization in `GetNormalizedName()`, `IsDependency()`
  - `search.go:72,85` - unit search capitalization/lowercase
  - Migrated: `ident.Normalize()` for lowercase, `ident.Equal()` for comparison

- [x] **14.1.2** Migrate `internal/types/helper.go` (8 occurrences)
  - Lines: 51,63,98,112,133,163,193,223
  - Helper type lookups: `helpersByName`, method/property/var/const lookups
  - Migrated: All `strings.ToLower()` → `ident.Normalize()`, removed `strings` import

- [x] **14.1.3** Migrate `internal/semantic/validation_pass.go` (12 occurrences)
  - Lines: 839,879,1247,1738,2093,2141,2175,2212,2232,2324,2340,2356
  - Method, field, property, type name validations
  - Migrated: All `strings.ToLower()` → `ident.Normalize()`

- [x] **14.1.4** Migrate `internal/semantic/analyze_*.go` files (40+ occurrences total)
  - Migrated 16 files in analyze_*.go
  - Used pkgident/identpkg aliases where local `ident` variables conflict
  - Removed unused `strings` imports

- [x] **14.1.5** Migrate `internal/semantic/contract_pass.go` (4 EqualFold occurrences)
  - Lines: 279,289,474,483
  - Field, constant, parameter lookups
  - Migrated: `strings.EqualFold()` → `pkgident.Equal()`

- [x] **14.1.6** Migrate `internal/bytecode/compiler_expressions.go` (9 occurrences)
  - ToLower: 210,462,466,523,734,872
  - EqualFold: 80,96,117
  - Migrated: `strings.ToLower()` → `pkgident.Normalize()`, `strings.EqualFold()` → `pkgident.Equal()`, `strings.HasPrefix()` → `pkgident.HasPrefix()`

- [x] **14.1.7** Migrate `internal/bytecode/vm_calls.go` (1 occurrence)
  - Line: 181 - method name lookup
  - Migrated: `strings.ToLower()` → `ident.Normalize()`

- [ ] **14.1.8** Migrate `internal/interp/types/type_system.go` (17 occurrences)
  - Lines: 145,151,157,175,181,187,258,265,270,299,312,318,331,337,350,393,412
  - Record, interface, helper, class, enum type ID lookups

- [ ] **14.1.9** Migrate `internal/interp/record.go` (20 occurrences)
  - Lines: 20,75,91,108,136,142,165,185,232,242,304,322,348,368,425
  - Record field, method, constant, class var lookups

- [ ] **14.1.10** Migrate `internal/interp/statements_declarations.go` (17 occurrences)
  - ToLower: 99,128,188,199,209,222,231,246,274,339,347,355,366,374,389,416
  - EqualFold: 161
  - Variable declaration type resolution

- [ ] **14.1.11** Migrate `internal/interp/objects_methods.go` (12 occurrences)
  - ToLower: 276,280,439,517
  - EqualFold: 41,554,699,1056,1073,1081
  - Method and constructor lookups

- [ ] **14.1.12** Migrate `internal/interp/expressions_complex.go` (9 occurrences)
  - ToLower: 130,166,198,238,268,329,386
  - EqualFold: 121,178,246
  - Type casting and interface checks

- [ ] **14.1.13** Migrate `internal/interp/functions_*.go` files (12 occurrences total, counting individual function calls)
  - `functions_typecast.go:91,329,626` (ToLower) + `389,508(x2),601` (EqualFold; line 508 has two calls)
  - `functions_calls.go:213,298,420` (ToLower) + `179` (EqualFold)
  - `functions_pointers.go:116` (ToLower)

- [ ] **14.1.14** Migrate `internal/interp/declarations.go` (31 occurrences)
  - ToLower: 37,42,56,61,72,77,89,94,105,110,122,127,139,144,155,160,172,177,189,194,206,211,223,228
  - EqualFold: 157,165,247,478,480,806,866
  - Class, method, operator declarations

- [ ] **14.1.15** Migrate `internal/interp/expressions_basic.go` (5 EqualFold occurrences)
  - Lines: 132,136,147,151,224
  - ClassName, ClassType special identifiers

- [ ] **14.1.16** Migrate remaining `internal/interp/` files (25 occurrences total)
  - `unit_loader.go:345` (1)
  - `helpers_conversion.go:113,141,146` (3 - excl. line 276 LowerCase builtin)
  - `builtins_ordinals.go:69,183,252,302,379` (5)
  - `ffi_errors.go:27,29,108,110` (4)
  - `contracts.go:15,18` (2)
  - `builtins_type.go:71,160,313,332,351` (5)
  - `statements_assignments.go:550,591` (ToLower) + `616` (EqualFold)
  - `enum.go:109` (1)
  - `helpers_validation.go:69` (ToLower/EqualFold) [Further investigation required: additional occurrences may exist. Please audit this file and update with specific line numbers.]
  - `objects_instantiation.go:17` (EqualFold)
  - `value.go:236` (EqualFold)
  - `helpers_comparison.go:17,27,37` (3 EqualFold)

- [ ] **14.1.17** Migrate `internal/interp/evaluator/` files (9 occurrences total)
  - `visitor_expressions.go:167,467,1837` (ToLower) + `131,137,148,154` (EqualFold)
  - `visitor_statements.go:1194` (ToLower) + `199` (EqualFold)

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
