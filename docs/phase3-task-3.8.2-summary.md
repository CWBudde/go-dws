# Phase 3.8.2: Reduce pkg/ast usage in internal/interp

## Objective

Minimize imports from `pkg/ast` to `internal/ast` in the interpreter package, ensuring clean separation between public and internal APIs while maintaining clear documentation for necessary `pkg/ast` usage.

## Background

The project has two AST packages:
- **`pkg/ast`**: Public API for external consumers
- **`internal/ast`**: Alias package that re-exports `pkg/ast` types for internal use

The goal is to have internal code use `internal/ast` (the alias) and only import `pkg/ast` directly when accessing semantic analysis metadata (`SemanticInfo`), which is not part of the AST structure itself.

## Changes Made

### 1. Updated internal/interp/errors Package

**Files modified:**
- `internal/interp/errors/errors.go`
- `internal/interp/errors/catalog.go`
- `internal/interp/errors/errors_test.go`
- `internal/interp/errors/catalog_test.go`

**Change:** Changed imports from `github.com/cwbudde/go-dws/pkg/ast` to `github.com/cwbudde/go-dws/internal/ast`

**Rationale:** The errors package only uses `ast.Node` interface for position extraction and error reporting. Since `internal/ast.Node` is an alias to `pkg/ast.Node`, this change maintains the same functionality while using the internal alias consistently.

### 2. Documented Remaining pkg/ast Usage

**Files modified:**
- `internal/interp/array.go`
- `internal/interp/interpreter.go`
- `internal/interp/evaluator/evaluator.go`

**Change:** Added documentation comments explaining why `pkg/ast` is imported directly:

```go
// Task 3.8.2: pkg/ast is imported for SemanticInfo, which holds semantic analysis
// metadata (type annotations, symbol resolutions). This is separate from the AST
// structure itself and is not aliased in internal/ast.
```

**Rationale:** `SemanticInfo` is semantic analysis metadata that's separate from the AST structure. It contains:
- Type annotations for nodes
- Symbol resolution information
- Metadata from the semantic analyzer

This data is appropriately kept in `pkg/ast` as it represents the output of semantic analysis, not the AST itself.

## Architecture After Changes

### Import Strategy

1. **Internal code using AST nodes**: Import `internal/ast` (alias package)
2. **Internal code using SemanticInfo**: Import `pkg/ast` directly (documented)
3. **External code**: Import `pkg/ast` for public API

### Clear Separation

- **AST Structure** (`internal/ast` ← `pkg/ast`): Node types, interfaces
- **Semantic Metadata** (`pkg/ast` only): SemanticInfo, type annotations
- **Error Handling** (`internal/interp/errors`): Uses `internal/ast.Node`

## Testing

All tests pass successfully:
- ✅ `internal/interp/errors` package tests (all 58 tests pass)
- ✅ `internal/ast` package tests (all tests pass)
- ✅ Build verification (`go build ./cmd/dwscript`)

## Acceptance Criteria

✅ **Clear separation**: Internal code uses `internal/ast` except for `SemanticInfo`
✅ **Documented rationale**: Comments explain why `pkg/ast` is used for semantic info
✅ **Tests pass**: All existing tests continue to pass
✅ **No adapters needed**: `internal/ast` already serves as the adapter layer

## Summary

This task successfully reduced `pkg/ast` usage in `internal/interp` by:
1. Converting the errors package to use `internal/ast` (4 files)
2. Documenting why `SemanticInfo` requires direct `pkg/ast` import (3 files)
3. Maintaining clear architectural separation between AST structure and semantic metadata

The changes maintain backward compatibility, pass all tests, and establish a clear pattern for future development:
- Use `internal/ast` for AST nodes
- Use `pkg/ast` only for `SemanticInfo` (documented)
