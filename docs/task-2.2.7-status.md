# Task 2.2.7 Status: parseExpression Core Implementation

**Status**: PARTIALLY COMPLETE
**Date**: 2025-01-17
**Estimated Remaining**: 4-6 hours (fixing adapters or removing them)

## Overview

Task 2.2.7 aimed to implement cursor-based parseExpression, the heart of the Pratt parser. This document describes what was accomplished, current limitations, and recommended paths forward.

## What Was Accomplished

### ✓ parseExpression Dispatcher (COMPLETE)

**Status**: Working correctly

Created dispatcher that routes to appropriate implementation based on parser mode:

```go
func (p *Parser) parseExpression(precedence int) ast.Expression {
    if p.useCursor {
        return p.parseExpressionCursor(precedence)
    }
    return p.parseExpressionTraditional(precedence)
}
```

**Files**: `internal/parser/expressions.go` (lines 12-22)

### ✓ parseExpressionTraditional Rename (COMPLETE)

**Status**: Working correctly

Renamed existing `parseExpression` to `parseExpressionTraditional` to enable dual-mode operation. No logic changes, just renamed for clarity.

**Files**: `internal/parser/expressions.go` (lines 24-91)

### ✓ parseExpressionCursor Core Implementation (COMPLETE)

**Status**: Works for expressions using only migrated functions

Implemented cursor-based expression parser using:
- Cursor function maps (`prefixParseFnsCursor`, `infixParseFnsCursor`)
- Stateless `getPrecedence()` helper
- Cursor navigation via `Peek(1)` and `Advance()`
- Fallback to traditional mode for unmigrated functions

**Key Features**:
- Pure functional prefix function lookup
- Precedence-based loop termination
- Special handling for "not in/is/as" operators
- Graceful fallback when cursor functions unavailable

**Files**: `internal/parser/expressions.go` (lines 93-164)

**Code Pattern**:
```go
func (p *Parser) parseExpressionCursor(precedence int) ast.Expression {
    // 1. Lookup prefix function from cursor map
    currentToken := p.cursor.Current()
    prefixFn, ok := p.prefixParseFnsCursor[currentToken.Type]
    if !ok {
        // Fall back to traditional mode if no cursor version
        p.syncCursorToTokens()
        p.useCursor = false
        result := p.parseExpressionTraditional(precedence)
        p.useCursor = true
        return result
    }
    leftExp := prefixFn(currentToken)

    // 2. Precedence climbing loop
    for {
        nextToken := p.cursor.Peek(1)

        // Termination conditions (semicolon, precedence)
        if nextToken.Type == lexer.SEMICOLON { break }

        nextPrec := getPrecedence(nextToken.Type)
        if precedence >= nextPrec && !(nextToken.Type == lexer.NOT && precedence < EQUALS) {
            break
        }

        // Special "not in/is/as" handling
        if nextToken.Type == lexer.NOT && precedence < EQUALS {
            leftExp = p.parseNotInIsAsCursor(leftExp)
            if leftExp == nil { break }
            continue
        }

        // Normal infix handling
        infixFn, ok := p.infixParseFnsCursor[nextToken.Type]
        if !ok {
            // Fall back to traditional mode
            p.syncCursorToTokens()
            p.useCursor = false
            result := p.parseExpressionTraditional(precedence)
            p.useCursor = true
            return result
        }

        p.cursor = p.cursor.Advance()
        operatorToken := p.cursor.Current()
        p.syncCursorToTokens()
        leftExp = infixFn(leftExp, operatorToken)
    }

    return leftExp
}
```

### ✓ parseNotInIsAsCursor Helper (COMPLETE)

**Status**: Working correctly with Mark/ResetTo backtracking

Implemented cursor-based handling for DWScript's special "not in", "not is", "not as" operators:
- Uses `cursor.Mark()` to save position
- Speculatively checks for IN/IS/AS after NOT
- Uses `cursor.ResetTo()` to backtrack if not a match
- Cleaner than traditional manual state save/restore

**Files**: `internal/parser/expressions.go` (lines 166-211)

**Code Pattern**:
```go
func (p *Parser) parseNotInIsAsCursor(leftExp ast.Expression) ast.Expression {
    // Mark position for backtracking
    mark := p.cursor.Mark()

    // Advance to NOT token
    p.cursor = p.cursor.Advance()
    notToken := p.cursor.Current()

    // Check if next token is IN/IS/AS
    nextToken := p.cursor.Peek(1)
    if nextToken.Type != lexer.IN && nextToken.Type != lexer.IS && nextToken.Type != lexer.AS {
        // Not a "not in/is/as" pattern, backtrack
        p.cursor = p.cursor.ResetTo(mark)
        p.syncCursorToTokens()
        return nil
    }

    // Parse the operator and create NOT expression
    p.cursor = p.cursor.Advance()
    operatorToken := p.cursor.Current()

    infixFn, ok := p.infixParseFnsCursor[operatorToken.Type]
    if !ok {
        p.cursor = p.cursor.ResetTo(mark)
        p.syncCursorToTokens()
        return nil
    }

    p.syncCursorToTokens()
    comparisonExp := infixFn(leftExp, operatorToken)

    // Wrap in NOT expression
    return &ast.UnaryExpression{...}
}
```

### ⚠️ Adapter Functions (PARTIALLY WORKING)

**Status**: Complex state synchronization issues

Attempted to create adapters for unmigrated functions (parseGroupedExpression, parseArrayLiteral, etc.) that:
1. Sync cursor to tokens (`syncCursorToTokens()`)
2. Temporarily switch to traditional mode (`useCursor = false`)
3. Call traditional function
4. Restore cursor mode (`useCursor = true`)
5. Sync tokens back to cursor (`syncTokensToCursor()`)

**Problem Discovered**: Mixing cursor and traditional modes within a single expression parse is very complex:

- Both modes consume tokens from the same lexer
- When traditional functions advance the lexer, the cursor's internal state becomes invalid
- Attempting to re-sync the cursor by advancing it re-reads from the already-consumed lexer
- This leads to the cursor being at the wrong position (often EOF)

**Files**: `internal/parser/parser.go` (lines 639-899)

**Attempted Solutions**:
1. ❌ `syncTokensToCursor()` - advances cursor to match curToken, but reads wrong tokens from consumed lexer
2. ❌ Lexer state save/restore - too complex, circular dependencies
3. ⚠️ **Current**: Fall back to traditional mode entirely when unmigrated function encountered

### ✓ Helper Functions (COMPLETE)

**Status**: Working correctly

Added supporting infrastructure:

1. **getPrecedence(tokenType)** - Stateless precedence lookup
   - `internal/parser/parser.go` lines 814-827

2. **syncTokensToCursor()** - Sync cursor to match curToken/peekToken
   - `internal/parser/parser.go` lines 910-929
   - **Note**: Has limitations due to shared lexer state

## Current Test Results

**Passing**: 16 of 17 migration test subtests
**Failing**: 1 test (parentheses_override_precedence)

**Failure Details**:
- Test: `TestMigration_InfixExpression_Precedence/parentheses_override_precedence`
- Input: `"(2 + 3) * 4"`
- Error: `no prefix parse function for EOF found at 1:12`
- Root Cause: Adapter complexity - cursor/traditional mode synchronization fails

**Why Other Tests Pass**:
- Tests using only migrated functions (basic literals, binary operators) work correctly
- parseExpressionCursor handles pure cursor mode well
- Fallback to traditional mode works for entire expressions

**Why This Test Fails**:
- Requires `parseGroupedExpression` (LPAREN prefix function)
- This function isn't migrated to cursor mode yet
- Adapter tries to call it in traditional mode
- State synchronization between modes fails
- Cursor ends up at wrong position (EOF instead of "*")

## Core Problem: Mixing Cursor and Traditional Modes

**Challenge**: Parser has two token navigation systems that don't cleanly interoperate:

1. **Traditional Mode**:
   - Uses mutable `curToken`/`peekToken`
   - Advances via `nextToken()` which calls `lexer.NextToken()`
   - Simple, stateful, works well

2. **Cursor Mode**:
   - Uses immutable `TokenCursor`
   - Advances via `cursor.Advance()` which also calls `lexer.NextToken()`
   - Functional, supports backtracking, cleaner

**The Conflict**:
- Both modes read from the SAME lexer
- When traditional functions advance the lexer, cursor's cached tokens become invalid
- When cursor advances, it re-reads from lexer at the wrong position
- No clean way to "sync" cursor to arbitrary token positions

**Attempted Workarounds**:
1. Save/restore lexer state - too complex, circular dependencies
2. Advance cursor to match curToken - reads wrong tokens from consumed lexer
3. Recreate cursor at current position - can't access lexer's internal position

## Paths Forward

### Option A: Remove Adapters (RECOMMENDED - SHORT TERM)

**Approach**: Accept that some expressions fall back to traditional mode entirely

**Steps**:
1. Remove all complex adapter registrations
2. Only register functions with true cursor implementations:
   - `parseIdentifierCursor`, `parseIntegerLiteralCursor`, etc.
   - `parseInfixExpressionCursor`
3. When parseExpressionCursor hits unmigrated function, fall back to traditional mode
4. Document this as expected behavior during migration

**Pros**:
- Simple, clean, no complex state synchronization
- Existing tests for migrated functions pass
- Clear path forward: migrate more functions → more expressions work in cursor mode

**Cons**:
- Some expressions (with parentheses, arrays, etc.) use traditional mode
- Not "pure" cursor mode yet
- Temporary limitation

**Estimate**: 1-2 hours (cleanup + documentation)

### Option B: Fix Adapters with Lexer Coordination (COMPLEX)

**Approach**: Make lexer aware of cursor mode and coordinate token consumption

**Steps**:
1. Add lexer mode flag (cursor vs traditional)
2. When in cursor mode, lexer maintains separate token cache for cursor
3. When switching modes, synchronize caches
4. Update `nextToken()` to check mode and update appropriate cache

**Pros**:
- True dual-mode operation
- All expressions work in cursor mode (via adapters)
- No fallback needed

**Cons**:
- HIGH complexity - fundamental architecture change
- Risk of bugs in lexer coordination
- May not be worth effort given migration is temporary

**Estimate**: 8-12 hours (lexer changes + testing)

### Option C: Migrate All Needed Functions First (DEFERRED)

**Approach**: Don't implement adapters; migrate functions in dependency order

**Steps**:
1. Complete Task 2.2.7 without adapters
2. Move to Task 2.2.10 (migrate expression helpers)
3. Move to Task 2.2.11 (migrate remaining infix handlers)
4. As functions are migrated, more expressions naturally work in cursor mode
5. Eventually no adapters needed

**Pros**:
- Avoids adapter complexity entirely
- Each function migration is clean and testable
- Natural progression of Strangler Fig pattern

**Cons**:
- Takes longer to get full cursor mode working
- Some expressions use traditional mode for longer

**Estimate**: Follows existing plan (Tasks 2.2.10-2.2.11 = 28 hours)

## Recommendation

**Recommended Path**: **Option A** (Remove Adapters) + **Option C** (Continue Migration)

**Reasoning**:

1. **Pragmatic**: Adapters are complex and provide limited value during migration
2. **Clean**: Pure cursor functions work correctly; fallback works correctly
3. **Incremental**: As we migrate more functions (Tasks 2.2.10, 2.2.11), coverage increases naturally
4. **Testable**: Each function migration can be tested in isolation
5. **Aligns with Strangler Fig**: Gradually replace traditional with cursor, don't force coexistence

**Immediate Actions**:
1. Remove adapter registrations for unmigrated functions
2. Keep only true cursor function registrations (IDENT, INT, FLOAT, STRING, TRUE, FALSE, binary operators)
3. Document fallback behavior in parseExpressionCursor comments
4. Update tests to reflect expected behavior (some use traditional mode)
5. Mark Task 2.2.7 as PARTIALLY COMPLETE with known limitations
6. Continue to Task 2.2.10 (migrate helpers) to expand cursor coverage

## Technical Learnings

### What Worked Well

1. **Cursor function maps** - Clean separation of cursor/traditional implementations
2. **getPrecedence helper** - Stateless precedence lookup works great
3. **Fallback mechanism** - Graceful degradation to traditional mode
4. **Mark/ResetTo backtracking** - Much cleaner than manual state save/restore

### What Was Challenging

1. **Shared lexer state** - Both modes consuming from same lexer causes sync issues
2. **Adapter complexity** - Trying to mix modes within single expression parse is hard
3. **State synchronization** - No clean way to sync cursor to arbitrary positions
4. **Testing mixed modes** - Hard to verify correct behavior when modes interact

### Key Insights

1. **Migration is incremental** - Don't need everything in cursor mode immediately
2. **Coexistence is temporary** - Eventually all functions will be cursor mode
3. **Fallback is okay** - Using traditional mode during migration is acceptable
4. **Simplicity wins** - Complex adapters not worth the effort for temporary state

## Files Modified

**Implementation**:
- `internal/parser/expressions.go` (+148 lines)
  - parseExpression dispatcher (lines 12-22)
  - parseExpressionTraditional rename (lines 24-91)
  - parseExpressionCursor implementation (lines 93-164)
  - parseNotInIsAsCursor helper (lines 166-211)

- `internal/parser/parser.go` (+224 lines)
  - Cursor function types (lines 371-380)
  - Cursor function maps in Parser struct (lines 420-421)
  - registerPrefixCursor/registerInfixCursor methods (lines 798-808)
  - getPrecedence helper (lines 814-827)
  - Adapter registrations (lines 639-899) - **TO BE REMOVED**
  - syncTokensToCursor helper (lines 910-929) - **TO BE REMOVED**

**Testing**:
- Existing migration tests used for validation
- 16 of 17 subtests passing
- 1 failing due to adapter complexity

## Next Steps

If proceeding with **Option A** (Recommended):

1. **Clean up parser.go** (~1 hour):
   - Remove adapter registrations (lines 639-899)
   - Remove syncTokensToCursor (lines 910-929)
   - Keep only true cursor function registrations

2. **Update parseExpressionCursor comments** (~15 min):
   - Document fallback behavior clearly
   - Note which functions trigger fallback
   - Explain this is temporary during migration

3. **Update tests** (~30 min):
   - Document which tests use fallback
   - Add comments explaining expected behavior
   - Consider skipping failing test with explanation

4. **Update PLAN.md** (~15 min):
   - Mark Task 2.2.7 as PARTIALLY COMPLETE
   - Document known limitations
   - Update Task 2.2.8 (testing) to reflect current state

5. **Proceed to Task 2.2.10** (Expression Helpers):
   - Migrate `parseExpressionList`
   - Migrate `parseArgumentsOrFields`
   - This will enable more complex expressions in cursor mode

**Total Cleanup**: ~2 hours

## Summary

Task 2.2.7 successfully implemented the core parseExpressionCursor with proper fallback mechanisms. The implementation works correctly for expressions using migrated functions and gracefully falls back to traditional mode for unmigrated functions.

Attempting to create adapters revealed fundamental complexity in mixing cursor and traditional modes within a single parse. The recommended path forward is to remove the adapters and continue with incremental function migration (Tasks 2.2.10-2.2.11), which will naturally expand cursor mode coverage without the complexity of forced coexistence.

This approach aligns with the Strangler Fig pattern established in earlier tasks and maintains code simplicity while making steady progress toward full cursor mode adoption.
