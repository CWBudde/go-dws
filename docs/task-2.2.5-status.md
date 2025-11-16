# Task 2.2.5 Status: Infix Expression Migration

**Status**: Partially Complete (1 of 4 functions migrated)
**Date**: 2025-01-XX
**Estimated Remaining**: 30-40 hours (depending on approach)

## Overview

Task 2.2.5 aims to migrate infix expression parsing functions to cursor mode. This document describes what was accomplished, current limitations, and recommended paths forward.

## Original Task Definition

From PLAN.md:

```markdown
#### Task 2.2.5: Migrate Infix Expressions (Week 5, ~40 hours)

**Goal**: Migrate binary/infix expression parsing to cursor.

**Targets**:
- [ ] `parseBinaryExpression`
- [ ] `parseCallExpression`
- [ ] `parseMemberAccess`
- [ ] `parseIndexExpression`
```

## What Was Accomplished

### ✓ parseInfixExpression (Binary Operations)

**Status**: COMPLETE with tests

Migrated the binary expression parser which handles all infix operators:
- Arithmetic: `+`, `-`, `*`, `/`, `div`, `mod`, `shl`, `shr`, `sar`
- Comparison: `=`, `<>`, `<`, `>`, `<=`, `>=`
- Logical: `and`, `or`, `xor`
- Set operations: `in`
- Null coalescing: `??`

**Files**:
- `internal/parser/expressions.go` (+86 lines)
  - `parseInfixExpression()` - Dispatcher
  - `parseInfixExpressionTraditional()` - Original implementation
  - `parseInfixExpressionCursor()` - Cursor-based implementation

- `internal/parser/migration_infix_test.go` (263 lines, new)
  - 4 test functions, 17 subtests
  - All operators tested
  - Precedence and chaining validated
  - **All tests passing** ✓

**Implementation Pattern**:

```go
func (p *Parser) parseInfixExpressionCursor(left ast.Expression) ast.Expression {
    operatorToken := p.cursor.Current()

    // Build expression with cursor token
    expression := &ast.BinaryExpression{
        Token:    operatorToken,
        Operator: operatorToken.Literal,
        Left:     left,
    }

    // Get precedence without relying on p.curToken
    precedence := LOWEST
    if prec, ok := precedences[operatorToken.Type]; ok {
        precedence = prec
    }

    // Advance cursor explicitly
    p.cursor = p.cursor.Advance()

    // Sync state for recursive call (temporary until parseExpression migrated)
    p.syncCursorToTokens()

    // Parse right expression using traditional parseExpression
    expression.Right = p.parseExpression(precedence)

    return expression
}
```

### × parseCallExpression (Function Calls)

**Status**: NOT MIGRATED

**Reason**: High complexity with multiple helper dependencies

This function handles:
- Regular function calls: `foo(a, b, c)`
- Typed record literals: `Point(x: 5, y: 10)`
- Distinguishes between the two based on argument patterns

**Dependencies**:
- `parseCallOrRecordLiteral()` - Disambiguates calls vs literals
- `parseEmptyCall()` - Handles `foo()`
- `parseCallWithExpressionList()` - Parses argument list
- `parseExpressionList()` - Recursively parses expressions
- `parseArgumentsOrFields()` - Parses args or field initializers

**Migration Complexity**: HIGH - Would need to migrate all helper functions

### × parseMemberAccess (Member Access)

**Status**: NOT MIGRATED

**Reason**: Complex special cases and helper dependencies

This function handles:
- Simple member access: `obj.field`
- Method calls: `obj.method(args)`
- Class creation: `TClass.Create(args)` (special case)
- Keywords as member names (DWScript allows this)

**Dependencies**:
- Recursive calls to `parseCallExpression` for method calls
- Special handling for `Create()` pattern
- Token validation logic for valid member names

**Migration Complexity**: MEDIUM-HIGH

### × parseIndexExpression (Array Indexing)

**Status**: NOT MIGRATED

**Reason**: Complex multi-dimensional indexing logic

This function handles:
- Single-dimensional: `arr[i]`
- Multi-dimensional: `arr[i, j, k]`
- Nested desugaring: `arr[i, j]` becomes `(arr[i])[j]`
- Structured errors for missing `]`

**Dependencies**:
- Recursive calls to `parseExpression` for each index
- Loop-based desugaring for multi-dimensional access

**Migration Complexity**: MEDIUM

## Core Problem: parseExpression Dependency

All four infix handlers share a fundamental challenge:

```
parseExpression (traditional)
    └─> calls prefix parse functions
    └─> loops through infix operators
        └─> calls parseInfixExpression(left)
            └─> recursively calls parseExpression for right operand
                └─> ... and so on
```

**The Issue**:

The cursor versions of infix handlers can't operate purely in cursor mode because they recursively call `parseExpression`, which is still traditional mode. This creates a hybrid situation where:

1. The infix handler uses cursor for its own tokens
2. Must sync state before calling traditional parseExpression
3. parseExpression continues in traditional mode
4. State syncing happens implicitly via shared curToken/peekToken

This works (tests pass!), but it doesn't demonstrate full cursor independence.

## Migration Blockers

### Critical Blocker: parseExpression Migration

**File**: `internal/parser/expressions.go`
**Function**: `parseExpression(precedence int) ast.Expression`
**LOC**: ~60 lines
**Complexity**: HIGH

parseExpression is the heart of the Pratt parser. It:
- Looks up prefix parse functions based on curToken
- Calls prefix functions to get left expression
- Loops through infix operators based on precedence
- Has special handling for "not in", "not is", "not as" operators
- Uses manual state saving/restoring for backtracking

**Why Critical**: All expression parsing flows through this function. Until it's migrated to cursor mode, infix handlers can't operate in pure cursor mode.

**Estimate**: 16-24 hours to migrate properly with tests

### Secondary Blocker: Helper Functions

These would need migration for full cursor mode:

1. **parseExpressionList** - Parses comma-separated expressions
   - Used by parseCallExpression, parseIndexExpression
   - ~30 lines
   - Estimate: 4-6 hours

2. **parseArgumentsOrFields** - Disambiguates function args vs record fields
   - Used by parseCallExpression
   - ~40 lines
   - Estimate: 6-8 hours

3. **parseCallOrRecordLiteral** - Handles call vs record literal
   - Used by parseCallExpression
   - ~25 lines
   - Estimate: 4-6 hours

**Total Helper Migration**: 14-20 hours

## Paths Forward

### Option A: Complete Task 2.2.5 Now (High Effort)

**Approach**: Bite the bullet and migrate everything needed

**Steps**:
1. Migrate `parseExpression` to cursor mode (16-24 hours)
   - Create `parseExpressionTraditional` and `parseExpressionCursor`
   - Handle prefix/infix function lookup
   - Migrate "not in/is/as" special handling
   - Comprehensive tests

2. Migrate helper functions (14-20 hours)
   - `parseExpressionList`
   - `parseArgumentsOrFields`
   - `parseCallOrRecordLiteral`
   - Tests for each

3. Migrate remaining infix handlers (8-12 hours)
   - `parseCallExpression`
   - `parseMemberAccess`
   - `parseIndexExpression`
   - Tests for each

**Total Estimate**: 38-56 hours
**Risk**: HIGH - Large interconnected changes
**Benefit**: Full cursor mode for all expression parsing

### Option B: Split into Sub-Tasks (Recommended)

**Approach**: Break down into manageable pieces

**Task 2.2.5a**: Migrate parseInfixExpression ✓ DONE
- Proof of concept for infix handlers
- Demonstrate hybrid cursor/traditional approach
- **Status**: COMPLETE

**Task 2.2.5b**: Migrate parseExpression (NEW)
- Core expression dispatcher
- Enables pure cursor mode for infix handlers
- **Estimate**: 16-24 hours
- **Priority**: CRITICAL PATH

**Task 2.2.5c**: Migrate Expression Helpers (NEW)
- parseExpressionList
- parseArgumentsOrFields, etc.
- **Estimate**: 14-20 hours
- **Dependencies**: Task 2.2.5b

**Task 2.2.5d**: Complete Infix Handler Migration (NEW)
- parseCallExpression
- parseMemberAccess
- parseIndexExpression
- **Estimate**: 8-12 hours
- **Dependencies**: Tasks 2.2.5b, 2.2.5c

**Total Estimate**: 38-56 hours (same, but manageable chunks)
**Risk**: LOW - Incremental validation
**Benefit**: Clear progress, easier to review/test

### Option C: Defer Remaining Work (Fast Path)

**Approach**: Accept partial completion and move to next phase

**Rationale**:
- parseInfixExpression migration proves the pattern works
- Full migration requires parseExpression first anyway
- Other Phase 2 tasks may not depend on complete infix migration
- Can return to this later when needed

**Next Steps**:
1. Mark Task 2.2.5 as "Partially Complete"
2. Document pattern and learnings
3. Create placeholder tasks for remaining work
4. Move to Task 2.2.6 or Task 2.3 (Parser Combinators)

**Estimate**: 0 additional hours
**Risk**: LOW - Nothing breaks, can revisit later
**Benefit**: Maintain momentum on Phase 2

## Recommendation

**Recommended Path**: **Option B** (Split into Sub-Tasks)

**Reasoning**:

1. **parseExpression is the critical path** - Until it's migrated, infix handlers can't demonstrate full cursor value

2. **Incremental progress reduces risk** - Large monolithic changes are harder to review and test

3. **Clear dependencies** - The task breakdown makes the order obvious:
   - First: parseExpression (unlocks everything)
   - Second: Helper functions (enables complex handlers)
   - Third: Remaining infix handlers (completes the picture)

4. **Maintains momentum** - Each sub-task delivers value and can be committed independently

5. **Aligns with Strangler Fig pattern** - Build incrementally, validate continuously

## Technical Learnings

### What We Learned from parseInfixExpression

1. **Cursor pattern works for complex functions** - Even with recursion challenges, the pattern is viable

2. **Hybrid approach is functional** - Cursor handlers can call traditional functions with state syncing

3. **Tests validate equivalence** - Comprehensive tests prove both modes produce identical ASTs

4. **Documentation is critical** - Clear comments about limitations prevent confusion

5. **Incremental migration is feasible** - Don't need to migrate everything at once

### Challenges Encountered

1. **Recursive dependencies** - Expression parsing is highly recursive, making isolated migration hard

2. **State synchronization** - Cursor and traditional state must stay in sync during hybrid operation

3. **Helper function sprawl** - Many small helper functions depend on each other

4. **Precedence handling** - curPrecedence() relies on curToken, needed cursor-aware version

## Next Actions

If proceeding with **Option B** (Recommended):

1. **Update PLAN.md**:
   - Mark Task 2.2.5a as COMPLETE
   - Create Tasks 2.2.5b, 2.2.5c, 2.2.5d with estimates
   - Document dependencies between sub-tasks

2. **Start Task 2.2.5b** (Migrate parseExpression):
   - Read parseExpression implementation thoroughly
   - Design cursor version (prefix/infix lookup, precedence handling)
   - Create Traditional and Cursor versions
   - Comprehensive differential tests
   - Document "not in/is/as" special handling

3. **Create tracking document**:
   - docs/parse-expression-migration.md
   - Document design decisions
   - Track migration challenges
   - Record lessons learned

## Files Reference

**Implementation**:
- `internal/parser/expressions.go` - parseInfixExpression migration

**Tests**:
- `internal/parser/migration_infix_test.go` - Infix expression tests

**Documentation**:
- `docs/task-2.2.5-status.md` - This document
- `docs/dual-mode-parser.md` - Dual-mode architecture
- `docs/token-cursor.md` - TokenCursor API reference
- `docs/cursor-migration-pattern.md` - Migration patterns

## Related Tasks

- ✓ Task 2.2.1: TokenCursor Implementation
- ✓ Task 2.2.2: Dual-Mode Parser Setup
- ✓ Task 2.2.3: Migrate parseIntegerLiteral
- ✓ Task 2.2.4: Migrate Expression Literals
- ⚠️ Task 2.2.5: Migrate Infix Expressions (PARTIAL - 1/4 complete)
- ⏭️ Task 2.2.6: TBD (likely parseExpression migration)
- ⏭️ Phase 2.3: Parser Combinators

## Summary

Task 2.2.5 has successfully migrated `parseInfixExpression` to cursor mode with full test coverage, proving that the cursor pattern works for complex recursive parsing functions. However, completing the remaining three infix handlers requires first migrating `parseExpression` itself, as all infix handlers depend on it for recursive expression parsing.

The recommended path forward is to split Task 2.2.5 into four sub-tasks, with Task 2.2.5b focusing on migrating parseExpression as the critical path blocker. This approach maintains momentum, reduces risk, and aligns with the incremental Strangler Fig migration pattern established in earlier tasks.
