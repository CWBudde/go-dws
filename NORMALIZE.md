# Case Insensitivity Normalization

This document summarizes the migration to centralize all case-insensitive identifier handling using the `pkg/ident` package.

## Summary

**Goal**: Centralize all case-insensitive identifier operations to `pkg/ident`, eliminating scattered `strings.ToLower()` and `strings.EqualFold()` calls.

**Result**: Successfully migrated ~200+ locations across ~50 files to use `pkg/ident.Normalize()` and `pkg/ident.Equal()`.

---

## What Was Done

### Phase 1-2: Infrastructure (Completed)

- Enhanced `pkg/ident` package with `HasPrefix()`, `HasSuffix()` functions
- Created `docs/ident-migration-guide.md` with patterns and best practices
- Added `Map[T]` generic type in `pkg/ident/map.go`
- Fixed 2 critical bugs where original casing was lost in error messages

### Phase 3: Token & Lexer (Verified)

- Confirmed already centralized via `token.LookupIdent()`
- Added comprehensive case insensitivity tests for keywords

### Phase 4: Symbol Table & Type Registry (Completed)

- Migrated `internal/semantic/symbol_table.go` (8 occurrences)
- Migrated `internal/semantic/type_registry.go` (5 occurrences)
- Added tests verifying original casing preserved for error messages

### Phase 5: Type System (Completed)

- Migrated `internal/types/types.go` - ClassType, InterfaceType methods
- Migrated `internal/types/compound_types.go` - RecordType, HelperType methods
- Migrated `internal/types/helper.go` - helper type lookups

### Phase 6: Semantic Analyzer (Completed)

- Migrated `analyze_classes.go`, `analyze_function_calls.go`, `analyze_records.go`
- Migrated `type_resolution.go`, `unit_analyzer.go`, `validation_pass.go`
- Migrated `contract_pass.go` and remaining `analyze_*.go` files

### Phase 7: Interpreter (Completed)

- Verified `environment.go` and `builtins/registry.go` already migrated
- Migrated `interpreter.go` (51 occurrences)
- Migrated `class.go`, `interface.go`, `objects_hierarchy.go`
- Migrated `record.go`, `declarations.go`, `expressions_*.go`
- Migrated `functions_*.go`, `statements_declarations.go`, `objects_methods.go`
- Migrated `unit_loader.go`, `helpers_*.go`, `ffi_errors.go`, `contracts.go`, `enum.go`
- Migrated `evaluator/visitor_*.go`

### Phase 8: Bytecode Compiler & VM (Completed)

- Migrated `compiler_core.go`, `compiler_statements.go`, `compiler_functions.go`
- Migrated `compiler_expressions.go`, `bytecode.go`, `vm_exec.go`, `vm_calls.go`

### Phase 9: Parser (Completed)

- Migrated `functions.go` - calling convention handling
- Migrated `operators.go` - operator symbol normalization

### Phase 10: Registries (Completed)

- Migrated `internal/units/registry.go`, `unit.go`, `search.go`
- Verified `internal/interp/types/*.go` registries already migrated

### Phase 11: Comprehensive Testing (Completed)

- Extended `internal/semantic/case_insensitive_test.go` with tests for records, enums, interfaces, properties, methods, builtins, keywords
- Created `internal/semantic/error_message_casing_test.go`
- Created `testdata/case-insensitivity-test.dws` for manual CLI testing
- All tests pass (excluding pre-existing interface test failures)

### Phase 12: Remaining Migrations

A few interpreter files still contain direct `strings.ToLower()` or `strings.EqualFold()` calls that should use `pkg/ident` for consistency. These are low-priority since they work correctly, but migration improves maintainability.

#### Tasks

- [x] **12.1** Migrate `internal/interp/builtins_ordinals.go`
  - Ordinal/enumeration value lookups
  - Migrated: 5 occurrences (4 enum type lookups + 1 class lookup)

- [x] **12.2** Migrate `internal/interp/builtins_type.go`
  - Type name comparisons in builtin type functions
  - Migrated: 5 occurrences (2 enum type lookups + 3 type ID registries)

- [x] **12.3** Migrate `internal/interp/statements_assignments.go`
  - Variable/field assignment resolution
  - Migrated: 3 occurrences (record field normalization + class lookup)

- [x] **12.4** Migrate `internal/interp/helpers_validation.go`
  - Validation helper functions
  - Migrated: 7 occurrences (helper name lookup, method registration, type indexing)

- [x] **12.5** Run cleanup commands
  - `goimports -w .` - Fix imports
  - `golangci-lint run --fix` - Auto-fix lint issues

- [x] **12.6** Verify all tests pass
  - `go test ./...`

### Phase 13: Adopt `ident.Map[T]` for Case-Insensitive Registries

The `ident.Map[T]` generic type (created in Phase 2) provides case-insensitive key storage while preserving original casing for error messages. Currently unused - all registries use manual `map[string]T` with `ident.Normalize()` calls.

**Benefits of adopting `ident.Map[T]`**:

- Eliminates boilerplate `ident.Normalize()` calls at every access
- Automatic original-casing preservation via `GetOriginalKey()`
- Consistent API across all registries
- Reduced risk of forgetting normalization

#### Tasks - High Priority (Core Infrastructure)

These are the most impactful changes, touching fundamental lookup structures:

- [ ] **13.1** Migrate `internal/semantic/symbol_table.go`
  - Field: `symbols map[string]*Symbol`
  - Replace with: `symbols *ident.Map[*Symbol]`
  - Impact: All variable/function/type symbol lookups
  - Note: `Symbol.Name` field stores original casing - verify `GetOriginalKey()` matches

- [ ] **13.2** Migrate `internal/semantic/type_registry.go`
  - Field: `types map[string]*TypeDescriptor`
  - Replace with: `types *ident.Map[*TypeDescriptor]`
  - Impact: All type lookups during semantic analysis
  - Note: Also has `kindIndex map[string][]string` - evaluate if needed

- [ ] **13.3** Migrate `internal/interp/environment.go`
  - Field: `store map[string]Value`
  - Replace with: `store *ident.Map[Value]`
  - Impact: Runtime variable storage and lookup
  - Note: Hot path - verify no performance regression

#### Tasks - Medium Priority (Registries)

Standalone registry classes with clear boundaries:

- [ ] **13.4** Migrate `internal/interp/builtins/registry.go`
  - Field: `functions map[string]*FunctionInfo`
  - Replace with: `functions *ident.Map[*FunctionInfo]`
  - Impact: Builtin function resolution

- [ ] **13.5** Migrate `internal/interp/types/class_registry.go`
  - Field: `classes map[string]*ClassInfoEntry`
  - Replace with: `classes *ident.Map[*ClassInfoEntry]`
  - Impact: Class metadata lookup

- [ ] **13.6** Migrate `internal/interp/types/function_registry.go`
  - Fields: `functions`, `qualifiedFunctions` (both `map[string][]*FunctionEntry`)
  - Replace with: `*ident.Map[[]*FunctionEntry]`
  - Impact: User-defined function resolution
  - Note: Stores slices for overloads - verify slice handling

- [ ] **13.7** Migrate `internal/units/registry.go`
  - Field: `units map[string]*Unit`
  - Replace with: `units *ident.Map[*Unit]`
  - Impact: Unit/module lookup
  - Note: Also has `loading map[string]bool` for cycle detection

#### Tasks - Lower Priority (Type System Internals)

Complex internal structures with multiple interconnected maps:

- [ ] **13.8** Migrate `internal/interp/types/type_system.go` - Part 1
  - Fields: `records`, `interfaces`, `helpers` maps
  - Higher complexity due to nested structures
  - Consider migrating incrementally

- [ ] **13.9** Migrate `internal/interp/types/type_system.go` - Part 2
  - Fields: `classTypeIDs`, `recordTypeIDs`, `enumTypeIDs`
  - These map type names to integer IDs
  - May not need original casing preservation

- [ ] **13.10** Migrate `OperatorRegistry` and `ConversionRegistry`
  - Fields: `entries map[string][]*OperatorEntry`, `implicit`/`explicit` maps
  - Lowest priority - internal optimization structures

---

## Legitimate Exceptions (Not Migrated)

The following uses are intentionally NOT migrated:

1. **`pkg/token/token.go`** - Foundation of case-insensitive keyword lookup (`LookupIdent()`, `IsKeyword()`)
2. **`pkg/ident/*.go`** - Implementation of the ident package itself
3. **Builtin string functions** - Actual DWScript runtime functions like `SameText`, `LowerCase`, `CompareText`
4. **Test files** - Error message checking, not identifier handling
5. **Non-identifier uses** - URL handling, format options, user input parsing

---

## Future Plans

The following items were identified but decided not to pursue:

### Prevention & Tooling (Phase 14)

- Add golangci-lint custom rule to forbid direct `strings.ToLower()`/`strings.EqualFold()` in identifier code
- Create linter exceptions for legitimate uses
- Add pre-commit hook
- Update CONTRIBUTING.md with case insensitivity guidelines

### Documentation (Phase 15)

- Update CLAUDE.md with centralized approach details
- Update README.md with case-insensitive language handling
- Expand godoc in `pkg/ident`
- Create architecture decision record (`docs/adr/003-case-insensitivity-centralization.md`)

---

## Related Files

- `pkg/ident/ident.go` - Centralized case handling
- `docs/ident-migration-guide.md` - Migration guide with patterns

---

**Completed**: 2025-11-22
