# Task 2.7.11: Update Call Sites - Analysis and Status

## Task Overview

Task 2.7.11 has three subtasks:
- **2.7.11.1**: Remove dispatch wrappers (2h) - Update direct callers to call Cursor functions
- **2.7.11.2**: Simplify test helpers (1h) - Update test helpers to assume Cursor mode
- **2.7.11.3**: Code cleanup (1h) - Remove commented-out Traditional calls and temporary variables

## Current Status

### Dependencies

Task 2.7.11 has dependencies on earlier tasks:
- **2.7.5-2.7.8**: Convert ~308 curToken/peekToken references to cursor equivalents (NOT YET DONE)
- **2.7.9**: Switch to Cursor-only mode (NOT YET DONE)
- **2.7.10**: Comprehensive testing (NOT YET DONE)

### Current State Analysis

**Dispatch Wrappers**: 35 wrapper functions identified that delegate to Cursor implementations:
```go
func (p *Parser) parseStatement() ast.Statement {
	return p.parseStatementCursor()
}
```

**Problem**: Cannot safely remove these wrappers yet because:
1. Some Traditional functions still exist and may call these wrappers
2. Some Cursor functions still delegate back to Traditional mode (e.g., `parseUnitCursor`)
3. Tasks 2.7.5-2.7.8 need to complete first to convert curToken usage

**Test Helpers** (task 2.7.11.2): ✅ ALREADY CLEAN
- `testParser()` creates parsers using `New()` which defaults to cursor mode
- No dual-mode test utilities found
- Test helpers are already cursor-mode ready

**Code Cleanup** (task 2.7.11.3): Can be done now
- Found 9 commented-out parse function calls that can be removed
- Task 2.7 comments (122 total) should be reviewed but most are documentation

## What Can Be Done Now

### 2.7.11.2: Test Helpers ✅
**STATUS**: Already complete - no dual-mode test utilities exist

### 2.7.11.3: Code Cleanup
**ACTION**: Remove commented-out code

Files with cleanup needed:
- `internal/parser/combinators.go`: 6 commented parse calls
- `internal/parser/parser.go`: 3 commented parse calls

## What Must Wait

### 2.7.11.1: Remove Dispatch Wrappers
**BLOCKED BY**: Tasks 2.7.5-2.7.10

**Reason**: Removing dispatch wrappers before completing tasks 2.7.5-2.7.8 will break:
1. Traditional functions that call wrappers (to be removed in 2.7.13)
2. Cursor functions that temporarily delegate to Traditional mode
3. Code still using curToken/peekToken instead of cursor

**Safe After**: Tasks 2.7.5-2.7.10 complete

## Recommendation

Complete task 2.7.11.3 (code cleanup) now, then:
1. Complete tasks 2.7.5-2.7.8 (convert curToken/peekToken usage)
2. Complete task 2.7.9 (switch to Cursor-only mode)
3. Complete task 2.7.10 (comprehensive testing)
4. Then return to complete task 2.7.11.1 (remove dispatch wrappers)

## Files Modified

- `docs/task-2.7.11-analysis.md` (this file)

## Attempted Implementation (2025-11-19)

**Attempt**: Updated 113 call sites from `p.parseExpression(` to `p.parseExpressionCursor(` across 15 files.

**Result**: FAILED - Tests failed due to dual-mode functions not having cursor properly initialized.

**Root Cause**: Many functions use both `p.curToken` (traditional state) and call parse functions. When these functions call `p.parseExpressionCursor()`, the cursor isn't synchronized with curToken, causing failures.

**Example Failures**:
- Array literal parsing: cursor not at expected position
- New array expressions: dual-mode positioning issues
- Class field declarations: curToken/cursor mismatch
- Record declarations: synchronization problems

**Conclusion**: Task 2.7.11.1 MUST wait for tasks 2.7.5-2.7.8 to complete. Those tasks will:
1. Convert all ~308 curToken/peekToken references to cursor equivalents
2. Eliminate dual-mode functions
3. Ensure all functions use cursor exclusively

## Next Steps

1. ✅ Complete task 2.7.11.3 (code cleanup)
2. ⚠️ Complete tasks 2.7.5-2.7.8 FIRST (convert curToken/peekToken to cursor)
3. ⚠️ Then complete task 2.7.11.1 (remove dispatch wrappers)
4. ✅ Mark task 2.7.11.2 as complete (already done)
