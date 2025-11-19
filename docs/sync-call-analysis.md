# Sync Call Analysis - Phase 2.7.4.3

## Executive Summary

**Sync calls cannot be eliminated yet** because ~308 active uses of `p.curToken` remain in the codebase.

Attempting to remove sync calls breaks 31 tests (377 → 346 passing) because code still expects `curToken` to be current after cursor-based parsing.

## Remaining Sync Calls (7 total)

### 1. expressions.go:85 - After parseExpressionCursor
**Purpose**: Keep curToken/peekToken synchronized after expression parsing
**Impact**: Called from many locations. Removing breaks array, call, and indexing tests.

### 2-3. expressions.go:107, 121 - After backtracking
**Purpose**: Update curToken to match restored cursor position after backtracking
**Impact**: Required for "not in/is/as" operator handling

### 4. interfaces.go:477 - Before Traditional function
**Purpose**: Prepare curToken for `parseFunctionPointerTypeDeclaration` (Traditional mode)
**Impact**: Required until function pointer parsing converted to cursor mode

### 5. parser.go:1018 - After restore state
**Purpose**: Sync curToken with cursor after state restoration
**Impact**: Required for backtracking/error recovery

### 6. statements.go:150 - After parseStatementCursor
**Purpose**: Keep curToken current after statement parsing
**Impact**: Required for statement dispatchers

### 7. unit.go:111 + 115 - Mode switch
**Purpose**: Sync before/after Traditional mode switch
**Impact**: Required until unit section parsers converted to cursor mode

## Root Cause: curToken Usage Still Widespread

Approximate breakdown of ~308 `p.curToken` references:

- **Helper functions** (curTokenIs, expectPeek, etc.) - ~50 uses
- **Traditional parse functions** (partially removed) - ~100 uses
- **Error reporting** (curToken.Pos) - ~80 uses
- **Statement/expression dispatchers** - ~30 uses
- **Type checking** (curToken.Type comparisons) - ~48 uses

## Why Removal Failed

When sync calls were removed:
1. parseExpressionCursor completes and updates `p.cursor`
2. But `p.curToken` is NOT updated (no sync)
3. Calling code checks `p.curToken` expecting current position
4. Gets stale token from before parsing
5. Tests fail with wrong token errors

## Path Forward

### Phase 2.7.4.3 (CURRENT)
- ✅ Created 6 cursor combinators
- ✅ Eliminated 14 sync calls by direct cursor calls
- ⏳ Document remaining sync call dependencies
- ⏳ Convert high-impact helper functions

### Phase 2.7.4.4 (NEXT)
Convert remaining curToken usage to cursor:
1. Replace `p.curToken` with `p.cursor.Current()`
2. Replace `p.peekToken` with `p.cursor.Peek(1)`
3. Update helper functions to use cursor
4. Update error reporting to use cursor positions
5. Then sync calls become no-ops

### Phase 2.7.4.5 (FINAL)
1. Remove curToken/peekToken fields from Parser struct
2. Remove sync methods
3. Remove useCursor flag
4. Clean up documentation

## Conclusion

**Sync calls are a symptom, not the problem.**

The real blocker is 308 uses of `p.curToken`. Until those are converted to `cursor.Current()`, sync calls must remain to keep Traditional code working.

**Next Actions**:
1. Identify highest-impact curToken usages
2. Convert them to cursor-based equivalents
3. As curToken usage decreases, more sync calls become removable
4. Final cleanup when curToken usage reaches zero
