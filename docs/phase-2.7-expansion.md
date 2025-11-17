# Phase 2.7: Migration Completion - Expanded Plan

**Created**: 2025-11-17
**Status**: Planning
**Total Estimated Effort**: 120 hours (Weeks 12-14)

## Table of Contents

- [Overview](#overview)
- [Current State Analysis](#current-state-analysis)
- [Migration Strategy](#migration-strategy)
- [Task 2.7.1: Statement Parsing Migration](#task-271-statement-parsing-migration)
- [Task 2.7.2: Type Parsing Migration](#task-272-type-parsing-migration)
- [Task 2.7.3: Declaration Parsing Migration](#task-273-declaration-parsing-migration)
- [Task 2.7.4: Remove Legacy Code](#task-274-remove-legacy-code)
- [Testing Strategy](#testing-strategy)
- [Success Criteria](#success-criteria)
- [Dependencies](#dependencies)

---

## Overview

Phase 2.7 completes the parser modernization by migrating ALL remaining parsing functions from traditional mutable state (curToken/peekToken/nextToken) to immutable cursor-based parsing with combinators. After this phase, the parser will be 100% modern, with all legacy code removed.

### Goals

1. **Complete Migration**: All parsing functions use cursor/combinators
2. **Zero Regressions**: 100% test pass rate maintained throughout
3. **Remove Dual Mode**: Eliminate useCursor flag and Traditional functions
4. **Clean Codebase**: Remove all legacy curToken/peekToken fields
5. **Documentation**: Update all docs for new architecture

### Context

**Already Completed** (Phases 2.2-2.5):
- ✅ TokenCursor infrastructure
- ✅ Parser combinators library
- ✅ NodeBuilder for automatic position tracking
- ✅ Expression parsing (literals, operators, calls, etc.)
- ✅ Basic control flow (if, while, repeat, for, case)
- ✅ Exception handling (try/raise)
- ✅ Basic statements (blocks, expressions, assignments)
- ✅ Variable/constant declarations
- ✅ Separation of concerns (parsing vs semantic analysis)

**Remaining to Migrate** (Phase 2.7):
- ❌ Type parsing (type expressions, arrays, records, function pointers, class types)
- ❌ Declaration parsing (functions, classes, interfaces, properties, enums)
- ❌ Advanced expressions (record literals, array operations)
- ❌ Helper utilities and supporting functions
- ❌ Legacy dual-mode infrastructure

---

## Current State Analysis

### Migration Status by File

| File | Functions | Migrated | Remaining | % Complete |
|------|-----------|----------|-----------|------------|
| `expressions.go` | 45+ | 40+ | ~5 | 90% |
| `control_flow.go` | 16 | 16 | 0 | 100% |
| `statements.go` | 12 | 8 | 4 | 67% |
| `arrays.go` | 8 | 3 | 5 | 38% |
| `types.go` | 12 | 0 | 12 | 0% |
| `functions.go` | 10 | 0 | 10 | 0% |
| `classes.go` | 15 | 1 | 14 | 7% |
| `records.go` | 8 | 0 | 8 | 0% |
| `enums.go` | 4 | 0 | 4 | 0% |
| `interfaces.go` | 8 | 0 | 8 | 0% |
| `properties.go` | 6 | 0 | 6 | 0% |
| `declarations.go` | 10 | 0 | 10 | 0% |
| `sets.go` | 4 | 0 | 4 | 0% |
| `operators.go` | 6 | 0 | 6 | 0% |
| **TOTAL** | **164** | **68** | **96** | **41%** |

### Key Insights

1. **Control flow is complete** - All if/while/for/case/try statements migrated
2. **Expressions mostly complete** - Main operators and literals done
3. **Type system untouched** - 0% of type parsing migrated
4. **Declarations untouched** - 0% of complex declarations migrated
5. **Mixed progress** - Arrays, statements partially done

### Architecture Status

**Current (Dual-Mode)**:
```
Parser
├── Traditional Mode (mutable)
│   ├── curToken, peekToken fields
│   ├── nextToken() method
│   ├── parse*Traditional() functions
│   └── Used by type/declaration parsing
│
└── Cursor Mode (immutable)
    ├── cursor field
    ├── cursor.Advance() returns new cursor
    ├── parse*Cursor() functions
    └── Used by expressions/control flow
```

**Target (Cursor-Only)**:
```
Parser
└── Cursor Mode (immutable only)
    ├── cursor field (no curToken/peekToken)
    ├── cursor.Advance() returns new cursor
    ├── parse*() functions (no Cursor suffix)
    └── Used by ALL parsing
```

---

## Migration Strategy

### Principles

1. **File-by-File**: Complete one file fully before moving to next
2. **Test-Driven**: Write tests first, validate equivalence
3. **Bottom-Up**: Migrate leaf functions before callers
4. **Incremental Commits**: Commit after each file or logical unit
5. **Performance Checks**: Run benchmarks at milestones
6. **No Breaking Changes**: Maintain test pass rate at 100%

### Migration Pattern (Per Function)

For each function `parseFoo()`:

1. **Analyze**: Understand current implementation and dependencies
2. **Create Cursor Version**: Implement `parseFooCursor()` using cursor API
3. **Write Tests**: Create differential tests validating equivalence
4. **Benchmark**: Measure performance impact
5. **Switch Dispatcher**: Update `parseFoo()` to call cursor version
6. **Validate**: Run full test suite, ensure 100% pass
7. **Document**: Note any special cases or changes

### Function Signature Patterns

**Before (Traditional)**:
```go
func (p *Parser) parseFoo() ast.Node {
    // Uses p.curToken, p.peekToken
    // Calls p.nextToken() to advance
    if !p.expectPeek(lexer.IDENT) {
        return nil
    }
    name := p.curToken.Literal
    // ...
}
```

**After (Cursor)**:
```go
func (p *Parser) parseFoo() ast.Node {
    cursor := p.cursor

    // Use cursor.Current(), cursor.Peek()
    // Advance cursor explicitly
    if cursor.Peek(1).Type != lexer.IDENT {
        p.addError("expected identifier")
        return nil
    }
    cursor = cursor.Advance()
    name := cursor.Current().Literal

    p.cursor = cursor  // Update parser cursor
    // ...
}
```

### Combinator Usage

Leverage existing combinators to reduce boilerplate:

```go
// Instead of manual token checking
if p.peekTokenIs(lexer.SEMICOLON) {
    p.nextToken()
}

// Use combinator
p.Optional(lexer.SEMICOLON)
```

```go
// Instead of manual list parsing
for !p.peekTokenIs(lexer.RPAREN) {
    param := p.parseParameter()
    params = append(params, param)
    if !p.peekTokenIs(lexer.COMMA) {
        break
    }
    p.nextToken()
}

// Use combinator
p.SeparatedList(SeparatorConfig{
    Sep: lexer.COMMA,
    Term: lexer.RPAREN,
    ParseItem: func() bool {
        param := p.parseParameter()
        if param != nil {
            params = append(params, param)
            return true
        }
        return false
    },
    AllowEmpty: true,
})
```

---

## Task 2.7.1: Statement Parsing Migration

**Duration**: Week 12 (~40 hours)
**Goal**: Migrate all remaining statement parsing to cursor/combinators

### Subtask 2.7.1.1: Complete `statements.go` Migration (8 hours)

**Remaining Functions** (4 functions):
1. `parseConstDeclaration` (2h)
   - Pattern: Similar to parseVarDeclaration (already migrated)
   - Create: `parseConstDeclarationCursor()`
   - Tests: Constant declarations with various types

2. `parseTypeDeclarationStatement` (2h)
   - Pattern: Type keyword + identifier + definition
   - Create: `parseTypeDeclarationStatementCursor()`
   - Tests: Simple types, aliases, complex types

3. `parseUnitDeclaration` (2h)
   - Pattern: Unit keyword + identifier + uses clause
   - Create: `parseUnitDeclarationCursor()`
   - Tests: Unit with/without uses, implementation section

4. `parseUsesClause` (2h)
   - Pattern: Uses keyword + comma-separated identifiers
   - Create: `parseUsesClauseCursor()`
   - Tests: Single unit, multiple units, in/out specifiers

**Files Modified**: `internal/parser/statements.go`

**Acceptance Criteria**:
- [ ] All 4 functions have Cursor versions
- [ ] Differential tests pass (traditional vs cursor produce same AST)
- [ ] Dispatcher updated to use cursor versions
- [ ] 100% test pass rate maintained

### Subtask 2.7.1.2: Complete `declarations.go` Migration (12 hours)

**All Functions** (10 functions, currently 0% migrated):

1. **Simple Declarations** (4h):
   - `parseSimpleVarDecl` (1h)
   - `parseSimpleConstDecl` (1h)
   - `parseTypeAliasDecl` (2h)

2. **Forward Declarations** (3h):
   - `parseForwardDeclaration` (2h)
   - `parseClassForwardDecl` (1h)

3. **Visibility Modifiers** (2h):
   - `parseVisibility` (1h) - Simple: public/private/protected/published
   - `parseVisibilityOrIdent` (1h) - Lookahead for visibility vs identifier

4. **Declaration Blocks** (3h):
   - `parseDeclarationBlock` (1h) - Var/const/type blocks
   - `parseDeclarationSection` (1h) - Interface/implementation sections
   - `parseImplementationSection` (1h) - Implementation details

**Pattern**:
```go
// Original
func (p *Parser) parseVisibility() ast.Visibility {
    switch p.peekToken.Type {
    case lexer.PUBLIC:
        p.nextToken()
        return ast.PublicVisibility
    // ...
    }
}

// Cursor version
func (p *Parser) parseVisibilityCursor() ast.Visibility {
    cursor := p.cursor
    switch cursor.Peek(1).Type {
    case lexer.PUBLIC:
        p.cursor = cursor.Advance()
        return ast.PublicVisibility
    // ...
    }
}
```

**Files Modified**: `internal/parser/declarations.go`

**Acceptance Criteria**:
- [ ] All 10 functions have Cursor versions
- [ ] Comprehensive tests for each function
- [ ] Dispatchers updated
- [ ] Integration tests pass

### Subtask 2.7.1.3: Migrate `operators.go` (6 hours)

**All Functions** (6 functions):

1. `parseOperatorOverload` (2h) - Operator overloading syntax
2. `parseImplicitOperator` (1h) - Implicit conversion operators
3. `parseExplicitOperator` (1h) - Explicit conversion operators
4. `parseBinaryOperatorOverload` (1h) - Binary operator (+, -, *, etc.)
5. `parseUnaryOperatorOverload` (1h) - Unary operator (-, not, etc.)
6. `parseComparisonOperatorOverload` (0.5h) - Comparison operators

**Key Challenge**: Operator precedence handling with cursor
**Strategy**: Use combinator for operator type checking

**Files Modified**: `internal/parser/operators.go`

### Subtask 2.7.1.4: Migrate `sets.go` (4 hours)

**All Functions** (4 functions):

1. `parseSetDeclaration` (1.5h) - Set of X type declarations
2. `parseSetLiteral` (1h) - Set literal expressions [1, 2, 3]
3. `parseSetConstructor` (1h) - Include/exclude syntax
4. `parseSetOperation` (0.5h) - Union, intersection, difference

**Pattern**: Similar to array parsing but with set semantics

**Files Modified**: `internal/parser/sets.go`

### Subtask 2.7.1.5: Complete `arrays.go` Migration (6 hours)

**Remaining Functions** (5 of 8):

1. `parseArrayDeclaration` (2h) - Array type declarations
2. `parseArrayBound` (1h) - Lower/upper bounds
3. `parseArrayBoundsFromCurrent` (1h) - Multi-dimensional bounds
4. `parseArrayOfConstDeclaration` (1h) - Array of const type
5. `parseDynamicArrayType` (1h) - Dynamic array syntax

**Already Migrated**: `parseArrayLiteral`, `parseIndexExpression` (have Cursor versions)

**Files Modified**: `internal/parser/arrays.go`

### Subtask 2.7.1.6: Integration Testing (4 hours)

**Goal**: Ensure all statement parsing works together

**Test Cases**:
1. Complex programs mixing var/const/type declarations (1h)
2. Nested blocks with various statement types (1h)
3. Units with multiple declaration sections (1h)
4. Error recovery in statement parsing (1h)

**Deliverable**: Comprehensive integration test suite

---

## Task 2.7.2: Type Parsing Migration

**Duration**: Week 13 (~40 hours)
**Goal**: Migrate type parsing to cursor/combinators

### Subtask 2.7.2.1: Basic Type Expressions (8 hours)

**File**: `internal/parser/types.go`

**Functions** (4 of 12):

1. `parseTypeExpression` (2h)
   - **Complexity**: High - dispatcher to specific type parsers
   - **Strategy**: Migrate after all specific parsers done
   - **Dependencies**: All other type parse functions

2. `parseSimpleType` (2h)
   - Integer, Float, String, Boolean, etc.
   - Pattern: Lookup in type table

3. `parseTypeIdentifier` (2h)
   - Named type references (e.g., TMyClass)
   - Pattern: Identifier with optional generic params

4. `parseTypeOrExpression` (2h)
   - Disambiguation: type vs expression context
   - **Challenge**: Complex lookahead needed
   - Strategy: Use Guard combinator

**Acceptance Criteria**:
- [ ] All 4 functions have Cursor versions
- [ ] Type resolution works correctly
- [ ] Lookahead disambiguation correct

### Subtask 2.7.2.2: Function Pointer Types (8 hours)

**File**: `internal/parser/types.go`

**Functions** (3):

1. `parseFunctionPointerType` (3h)
   - Pattern: function(params): ReturnType
   - **Complexity**: High - nested parameter parsing
   - **Dependencies**: Parameter list parsing

2. `parseProcedurePointerType` (2h)
   - Pattern: procedure(params)
   - Similar to function but no return type

3. `parseMethodPointerType` (3h)
   - Pattern: function(params): ReturnType of object
   - **Challenge**: "of object" clause handling

**Integration**: Must work with function declarations

### Subtask 2.7.2.3: Array Types (6 hours)

**File**: `internal/parser/types.go`

**Functions** (2):

1. `parseArrayType` (3h)
   - Pattern: array[bounds] of Type
   - **Complexity**: Medium - multi-dimensional support
   - **Dependencies**: `parseArrayBound`, `parseArrayBoundsFromCurrent`

2. `parseDynamicArrayType` (3h)
   - Pattern: array of Type (no bounds)
   - Simpler than static arrays

**Note**: Some array parsing in `arrays.go` - coordinate migration

### Subtask 2.7.2.4: Record Types (8 hours)

**File**: `internal/parser/types.go` and `internal/parser/records.go`

**Functions in types.go** (2):
1. `parseRecordType` (3h)
   - Pattern: record ... end
   - **Complexity**: High - field declarations, visibility, helpers

2. `parseHelperType` (2h)
   - Pattern: helper for RecordName ... end
   - **Challenge**: Forward reference resolution

**Functions in records.go** (remaining):
1. `parseRecordDeclaration` (already in list below)
2. `parseRecordFieldDeclarations` (1.5h)
3. `parseRecordPropertyDeclaration` (1.5h)

**Coordination**: These functions call each other, migrate together

### Subtask 2.7.2.5: Class Types (6 hours)

**File**: `internal/parser/types.go`

**Functions** (2):

1. `parseClassType` (3h)
   - Pattern: class(TParent, IInterface1, IInterface2)
   - Type reference, not full class declaration

2. `parseClassOfType` (3h)
   - Pattern: class of TClassName
   - Meta-class types

**Note**: Distinct from class declarations (parseClassDeclaration)

### Subtask 2.7.2.6: Enum and Set Types (4 hours)

**Files**: `internal/parser/types.go`, `internal/parser/enums.go`

**Functions**:
1. `parseEnumType` (2h) - Inline enum types
2. `parseSetType` (2h) - Set of EnumType

**Coordination**: Work with enum declarations from enums.go

---

## Task 2.7.3: Declaration Parsing Migration

**Duration**: Week 14, Days 1-3 (~24 hours)
**Goal**: Migrate complex declaration parsing

### Subtask 2.7.3.1: Function Declarations (8 hours)

**File**: `internal/parser/functions.go`

**Functions** (10):

1. **Main Declaration** (3h):
   - `parseFunctionDeclaration` (2h)
   - `parseProcedureDeclaration` (1h)

2. **Parameters** (3h):
   - `parseParameterList` (1h) - Can use SeparatedList combinator!
   - `parseParameterGroup` (1h) - Grouped params: a, b, c: Integer
   - `parseParameterListAtToken` (0.5h) - Starting at specific token
   - `parseTypeOnlyParameterListAtToken` (0.5h) - For function pointers

3. **Modifiers** (1h):
   - `parseParameterModifier` (0.5h) - var, const, out, lazy
   - `parseCallingConvention` (0.5h) - cdecl, stdcall, register, pascal

4. **Special Functions** (1h):
   - `parseExternalFunction` (0.5h) - External keyword
   - `parseForwardFunction` (0.5h) - Forward keyword

**Key Opportunity**: Use `SeparatedList` combinator for parameter parsing

**Example**:
```go
func (p *Parser) parseParameterListCursor() []*ast.Parameter {
    var params []*ast.Parameter

    p.SeparatedList(SeparatorConfig{
        Sep: lexer.COMMA,
        Term: lexer.RPAREN,
        ParseItem: func() bool {
            param := p.parseParameterCursor()
            if param != nil {
                params = append(params, param)
                return true
            }
            return false
        },
        AllowEmpty: true,
    })

    return params
}
```

### Subtask 2.7.3.2: Class Declarations (10 hours)

**File**: `internal/parser/classes.go`

**Functions** (14 remaining):

1. **Main Structure** (4h):
   - `parseClassDeclaration` (2h) - Entry point
   - `parseClassDeclarationBody` (1h) - Body parsing
   - `parseClassParentAndInterfaces` (1h) - Inheritance

2. **Members** (3h):
   - `parseFieldDeclarations` (1h) - Field lists
   - `parseMethodDeclaration` (1h) - Method declarations
   - `parseClassMethodDeclaration` (1h) - Class methods

3. **Special Members** (2h):
   - `parseClassConstantDeclaration` (0.5h) - Class constants
   - `parseClassVarDeclaration` (0.5h) - Class variables
   - `parseClassPropertyDeclaration` (1h) - Class properties

4. **Constructors/Destructors** (1h):
   - `parseConstructorDeclaration` (0.5h)
   - `parseDestructorDeclaration` (0.5h)

**Already Migrated**: `parseMemberAccess` (has Cursor version)

**Challenge**: Complex nesting, many member types

**Strategy**:
- Bottom-up: Field/method/property parsers first
- Then: Main class declaration
- Use combinators for visibility sections

### Subtask 2.7.3.3: Interface Declarations (4 hours)

**File**: `internal/parser/interfaces.go`

**Functions** (remaining):

1. `parseInterfaceDeclaration` (1.5h)
2. `parseInterfaceDeclarationBody` (1h)
3. `parseInterfaceMethodDecl` (1h)
4. `parseInterfacePropertyDecl` (0.5h)

**Pattern**: Similar to class but simpler (no implementation)

### Subtask 2.7.3.4: Record and Enum Declarations (2 hours)

**Files**: `internal/parser/records.go`, `internal/parser/enums.go`

**Records** (4 functions):
1. `parseRecordDeclaration` (0.5h)
2. `parseRecordOrHelperDeclaration` (0.5h)
3. `parseRecordFieldDeclarations` (0.5h) - May already be migrated
4. `parseRecordLiteral` (0.5h)

**Enums** (4 functions):
1. `parseEnumDeclaration` (0.5h)
2. `parseEnumValue` (0.25h)
3. `parseScopedEnum` (0.25h)
4. `parseEnumFlags` (0.25h)

---

## Task 2.7.4: Remove Legacy Code

**Duration**: Week 14, Days 4-5 (~16 hours)
**Goal**: Remove all old mutable parser code

### Subtask 2.7.4.1: Verify Complete Migration (2 hours)

**Checklist**:
- [ ] All `parse*Traditional()` functions identified
- [ ] All `parse*Cursor()` functions exist
- [ ] All `parse*()` dispatchers updated
- [ ] No direct `p.curToken`/`p.peekToken` access in parse functions
- [ ] All tests pass in cursor mode
- [ ] Performance benchmarks run

**Method**: Automated grep search + manual review

**Script**:
```bash
# Find any remaining Traditional functions
grep -rn "Traditional()" internal/parser/*.go | grep "^func"

# Find any remaining direct curToken access
grep -rn "p\.curToken\|p\.peekToken" internal/parser/*.go | grep -v "cursor\|comment"

# Verify Cursor versions exist
for file in internal/parser/*.go; do
    echo "=== $file ==="
    grep "Cursor()" "$file" | grep "^func"
done
```

### Subtask 2.7.4.2: Remove Traditional Functions (4 hours)

**Process**:
1. Remove all `parse*Traditional()` functions (2h)
2. Rename all `parse*Cursor()` to `parse*()` (1h)
3. Update all internal calls (0.5h)
4. Run tests, fix any issues (0.5h)

**Estimate**: ~500 lines removed

**Files Affected**: All parser files with Traditional functions

**Safety**:
- Commit before starting
- Remove one file at a time
- Run tests after each file
- Easy rollback if issues

### Subtask 2.7.4.3: Remove Mutable State Fields (3 hours)

**Remove from Parser struct**:
```go
// REMOVE THESE:
curToken  lexer.Token
peekToken lexer.Token

// KEEP THESE:
cursor    *TokenCursor
```

**Process**:
1. Remove field declarations from `parser.go` (0.5h)
2. Remove initialization in `New()` and builder (0.5h)
3. Remove `useCursor` flag and dual-mode logic (1h)
4. Update documentation comments (0.5h)
5. Run tests, fix any breakage (0.5h)

**Estimate**: ~50 lines removed

### Subtask 2.7.4.4: Remove nextToken() Method (2 hours)

**Remove**:
```go
func (p *Parser) nextToken() {
    p.curToken = p.peekToken
    p.peekToken = p.l.NextToken()
}
```

**Replace with**: Cursor advancement in callers

**Process**:
1. Search for all `p.nextToken()` calls (0.5h)
2. Replace with cursor advancement (1h)
3. Verify no calls remain (0.5h)

**Example Replacement**:
```go
// Before
p.nextToken()

// After
p.cursor = p.cursor.Advance()
```

### Subtask 2.7.4.5: Remove Old Error Methods (1 hour)

**Review and remove**:
- Old error methods that assume curToken/peekToken
- Duplicate error handling (use ErrorRecovery from Phase 2.5)

**Consolidate**: All error handling through `error_recovery.go`

### Subtask 2.7.4.6: Clean Up Imports and Dead Code (2 hours)

**Tasks**:
1. Remove unused imports (0.5h)
2. Remove commented-out code (0.5h)
3. Remove obsolete helper functions (0.5h)
4. Run `goimports` and `golangci-lint` (0.5h)

**Tools**:
```bash
# Find unused code
golangci-lint run --enable=unused,deadcode

# Format and organize imports
goimports -w internal/parser/

# Remove dead code
go mod tidy
```

### Subtask 2.7.4.7: Final Test Pass (2 hours)

**Comprehensive Testing**:
1. Run all unit tests (0.5h)
   ```bash
   go test ./internal/parser/... -v
   ```

2. Run integration tests (0.5h)
   ```bash
   go test ./... -run Integration
   ```

3. Run fixture tests (0.5h)
   ```bash
   go test ./internal/interp -run TestDWScriptFixtures
   ```

4. Run benchmarks (0.5h)
   ```bash
   go test ./internal/parser -bench=. -benchmem
   ```

**Acceptance**: 100% tests pass, performance within 5% of baseline

---

## Testing Strategy

### Per-Function Testing

For each migrated function:

1. **Differential Tests**:
   - Parse same input with Traditional and Cursor
   - Compare AST structure (should be identical)
   - Compare token positions (should match)

2. **Edge Cases**:
   - Empty input
   - Malformed syntax
   - Boundary conditions
   - Error recovery

3. **Integration Tests**:
   - Function used in larger context
   - Nested structures
   - Mixed with other migrated functions

### Test File Organization

```
internal/parser/
├── migration_types_test.go         # Type parsing tests
├── migration_functions_test.go      # Function declaration tests
├── migration_classes_test.go        # Class declaration tests
├── migration_interfaces_test.go     # Interface tests
├── migration_records_test.go        # Record tests
├── migration_enums_test.go          # Enum tests
└── migration_integration_test.go    # End-to-end tests
```

### Benchmark Strategy

**Create benchmarks for**:
1. Type parsing (before/after)
2. Function parsing (before/after)
3. Class parsing (before/after)
4. Overall parsing (composite benchmark)

**Target**: <5% performance regression

**Example**:
```go
func BenchmarkParseTypeExpression_Traditional(b *testing.B) {
    source := "array[0..10] of Integer"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p := New(lexer.New(source))
        _ = p.parseTypeExpressionTraditional()
    }
}

func BenchmarkParseTypeExpression_Cursor(b *testing.B) {
    source := "array[0..10] of Integer"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p := NewParserBuilder(lexer.New(source)).
            WithCursorMode(true).
            Build()
        _ = p.parseTypeExpression()  // Calls cursor version
    }
}
```

### Fixture Tests

Run complete DWScript test suite (~2,100 tests):
```bash
go test -v ./internal/interp -run TestDWScriptFixtures
```

**Monitor**: Test pass rate throughout migration (must stay at 100%)

---

## Success Criteria

### Functional Requirements

- [ ] **100% Migration**: All parse functions use cursor mode
- [ ] **Zero Regressions**: All tests pass (unit + integration + fixtures)
- [ ] **AST Equivalence**: Cursor mode produces identical ASTs to traditional
- [ ] **Error Handling**: Error messages remain clear and helpful
- [ ] **Position Tracking**: Token positions accurate in all nodes

### Code Quality

- [ ] **No Dual Mode**: `useCursor` flag removed
- [ ] **No Traditional Functions**: All `*Traditional()` functions removed
- [ ] **No Mutable State**: `curToken`/`peekToken` fields removed
- [ ] **No nextToken()**: Manual token advancement removed
- [ ] **Clean Code**: No dead code, unused imports, or TODOs

### Performance

- [ ] **Parsing Speed**: Within 5% of baseline
- [ ] **Memory Usage**: Within 10% of baseline
- [ ] **Allocations**: Comparable or better than baseline

### Documentation

- [ ] **Architecture Docs**: Updated for cursor-only mode
- [ ] **Parser Guide**: Reflects new patterns
- [ ] **Code Comments**: Accurate and helpful
- [ ] **Migration Retrospective**: Lessons learned documented

---

## Dependencies

### Task Dependencies

```
2.7.1 Statement Parsing
  └── No blocking dependencies (can start immediately)

2.7.2 Type Parsing
  ├── Depends on: Basic type functions (subtask 2.7.2.1)
  └── Enables: Function/class declarations need type parsing

2.7.3 Declaration Parsing
  ├── Depends on: 2.7.2 (type parsing complete)
  ├── Depends on: 2.7.1 (statement parsing complete)
  └── Enables: 2.7.4 (removal of legacy code)

2.7.4 Remove Legacy Code
  ├── Depends on: 2.7.1 (100% complete)
  ├── Depends on: 2.7.2 (100% complete)
  └── Depends on: 2.7.3 (100% complete)
```

### Function Dependencies

**Type Parsing**:
- `parseTypeExpression` depends on ALL specific type parsers
- `parseArrayType` depends on `parseArrayBound`
- `parseRecordType` depends on `parseRecordFieldDeclarations`
- `parseFunctionPointerType` depends on `parseParameterList`

**Declaration Parsing**:
- `parseFunctionDeclaration` depends on type parsing
- `parseClassDeclaration` depends on type parsing, member parsing
- `parseInterfaceDeclaration` depends on type parsing

### External Dependencies

**None** - All dependencies are within Phase 2.7

---

## Risk Mitigation

### Risks

1. **Breaking Tests**: Migration introduces regressions
   - **Mitigation**: Differential testing, small incremental commits

2. **Performance Degradation**: Cursor mode slower
   - **Mitigation**: Continuous benchmarking, optimization passes

3. **Incomplete Migration**: Miss some functions
   - **Mitigation**: Automated grep checks, comprehensive checklist

4. **Complex Interactions**: Unforeseen dependencies
   - **Mitigation**: Bottom-up migration, thorough testing

### Rollback Plan

If critical issue discovered:
1. **Identify**: Which commit introduced the issue
2. **Revert**: Rollback to last good commit
3. **Analyze**: Understand root cause
4. **Fix Forward**: Implement fix and re-apply
5. **Validate**: Extra testing before re-committing

---

## Milestones

### Week 12 (Task 2.7.1)
- **Day 1-2**: Complete statements.go, declarations.go
- **Day 3**: Migrate operators.go, sets.go
- **Day 4**: Complete arrays.go
- **Day 5**: Integration testing

**Deliverable**: All statement parsing migrated

### Week 13 (Task 2.7.2)
- **Day 1-2**: Basic type expressions, function pointers
- **Day 3**: Array types, record types
- **Day 4**: Class types, enum types
- **Day 5**: Type parsing integration tests

**Deliverable**: All type parsing migrated

### Week 14 (Task 2.7.3 + 2.7.4)
- **Day 1-2**: Function declarations, class declarations
- **Day 3**: Interface/record/enum declarations
- **Day 4**: Remove legacy code
- **Day 5**: Final validation and cleanup

**Deliverable**: Clean, modern parser with zero legacy code

---

## Estimated Effort Breakdown

| Task | Subtasks | Hours | % of Total |
|------|----------|-------|------------|
| 2.7.1 Statement Parsing | 6 | 40 | 33% |
| 2.7.2 Type Parsing | 6 | 40 | 33% |
| 2.7.3 Declaration Parsing | 4 | 24 | 20% |
| 2.7.4 Remove Legacy | 7 | 16 | 14% |
| **TOTAL** | **23** | **120** | **100%** |

---

## Conclusion

Phase 2.7 represents the completion of the parser modernization journey. By migrating all remaining parsing functions to cursor-based immutable navigation and removing all legacy code, we achieve:

1. **Modern Architecture**: 100% immutable, functional-style parsing
2. **Better Maintainability**: No mutable state, clearer code flow
3. **Better Testability**: Pure functions, easier to test
4. **Foundation for Future**: Ready for advanced features like parallel parsing, caching, incremental parsing

Upon completion, the parser will be:
- ✅ Fully cursor-based
- ✅ Zero legacy code
- ✅ Comprehensively tested
- ✅ Well-documented
- ✅ Performance-validated

**Next Phase**: Phase 2.8 - Optimization & Polish
