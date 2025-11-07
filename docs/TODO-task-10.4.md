# TODO: Task 10.4 - Parser Error Migration

**Status**: ⏸️ DEFERRED
**Priority**: Low (conversion layer works adequately)
**Estimated Effort**: 4-6 hours
**When**: During parser refactoring or when LSP needs better error spans

## Quick Overview

Migrate parser from string-based errors to structured `ParserError` type (~200+ error sites across 14 files).

## Why Not Done Yet?

✅ Current state is functional:
- Parser errors converted to structured errors via `convertParserErrors()`
- Position info extracted from error strings
- Public API already returns structured `*Error` objects
- LSP integration ready (with minor limitations)

⚠️ Task is large:
- ~200+ error sites to migrate
- 14 parser files to update
- Tests need updating
- Mostly mechanical work

## What Needs to Be Done?

1. Create `internal/parser/error.go` with `ParserError` struct
2. Update parser to use `[]*ParserError` instead of `[]string`
3. Migrate all error creation sites
4. Remove conversion layer
5. Update tests

## Documentation

📖 **Complete Implementation Guide**:
- [docs/task-10.4-parser-error-migration.md](./task-10.4-parser-error-migration.md) - Comprehensive walkthrough with code examples
- [PLAN.md lines 880-970](../PLAN.md#L880-L970) - Step-by-step checklist

## When to Tackle This

Consider doing this task when:
- [ ] Doing major parser refactoring
- [ ] LSP needs more precise error spans
- [ ] Adding parser error recovery
- [ ] Implementing better error messages
- [ ] Have 4-6 hours for systematic work

## Benefits When Complete

✨ **Better error handling**:
- More precise positions (no string parsing)
- Better error spans (multi-token ranges)
- Consistent with semantic analyzer
- Error codes for IDE integration

## Quick Start

When ready to implement:

1. Read [task-10.4-parser-error-migration.md](./task-10.4-parser-error-migration.md)
2. Start with Step 1: Create `internal/parser/error.go`
3. Follow systematic file-by-file migration
4. Test after each major file
5. Remove conversion layer at end

## Questions?

See the detailed guide: [task-10.4-parser-error-migration.md](./task-10.4-parser-error-migration.md)
