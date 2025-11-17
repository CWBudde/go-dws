# Task 2.2.10 Implementation Plan: Migrate Expression Helpers

**Goal**: Migrate helper functions that support complex expression parsing to cursor mode.

**Estimated Time**: 16 hours

**Status**: In Progress

## Function Dependency Map

```
parseCallOrRecordLiteral (ORCHESTRATOR)
  ├─> parseEmptyCall (SIMPLE)
  │     └─> No dependencies
  ├─> parseCallWithExpressionList (MEDIUM)
  │     └─> parseExpressionList (FOUNDATION)
  │           └─> parseExpression ✓ (cursor version exists)
  └─> parseArgumentsOrFields (COMPLEX)
        ├─> parseSingleArgumentOrField
        │     ├─> parseNamedFieldInitializer
        │     │     └─> parseExpression ✓
        │     └─> parseArgumentAsFieldInitializer
        │           └─> parseExpression ✓
        └─> advanceToNextItem (LIST NAVIGATION)
              └─> Token inspection only
```

## Implementation Order (Bottom-Up)

### Phase 1: Foundation (2 hours)
**Target**: parseExpressionListCursor
- **Complexity**: Low-Medium
- **Line Count**: ~20 lines
- **Dependencies**: parseExpression ✓ (cursor version exists)
- **Strategy**:
  - Create cursor version that uses parseExpressionCursor instead of parseExpression
  - Handle list navigation with cursor
  - May need cursor version of parseSeparatedListBeforeStart or inline the logic

### Phase 2: Simple Call Helpers (1 hour)
**Target**: parseEmptyCallCursor, parseCallWithExpressionListCursor
- **Complexity**: Low
- **Line Count**: ~30 lines total
- **Dependencies**: parseExpressionListCursor (from Phase 1)
- **Strategy**:
  - parseEmptyCallCursor: Simple token advancement
  - parseCallWithExpressionListCursor: Calls parseExpressionListCursor

### Phase 3: Field Initializer Helpers (3 hours)
**Target**: Support functions for parseArgumentsOrFields
- **Functions**:
  - parseNamedFieldInitializerCursor (~15 lines)
  - parseArgumentAsFieldInitializerCursor (~10 lines)
  - parseSingleArgumentOrFieldCursor (~10 lines)
  - advanceToNextItemCursor (~15 lines)
- **Complexity**: Medium
- **Line Count**: ~50 lines total
- **Dependencies**: parseExpression ✓
- **Strategy**:
  - Each function is relatively simple
  - Main challenge: cursor-based token navigation and lookahead
  - Use cursor.Peek() for lookahead instead of peekToken

### Phase 4: parseArgumentsOrFieldsCursor (4 hours)
**Target**: Main disambiguation logic
- **Complexity**: High
- **Line Count**: ~40 lines
- **Dependencies**: All Phase 3 functions
- **Strategy**:
  - Complex loop with state management
  - Cursor navigation through list
  - Track allHaveColons flag
  - Build list of FieldInitializer items

### Phase 5: parseCallOrRecordLiteralCursor (2 hours)
**Target**: Top-level orchestrator
- **Complexity**: Medium
- **Line Count**: ~25 lines
- **Dependencies**: All previous phases
- **Strategy**:
  - Orchestrates other cursor functions
  - Simple if/else dispatch logic
  - Calls appropriate helper based on lookahead

### Phase 6: Testing (4 hours)
**Target**: Comprehensive test suite
- **File**: migration_helpers_test.go (~400 lines)
- **Test Categories**:
  - TestMigration_ParseExpressionList (~80 lines)
    - Empty list
    - Single element
    - Multiple elements
    - Trailing comma
  - TestMigration_ParseEmptyCall (~40 lines)
    - Simple empty call
    - Typed empty call
  - TestMigration_ParseCallWithExpressionList (~80 lines)
    - Single argument
    - Multiple arguments
    - Complex expressions as arguments
  - TestMigration_ParseArgumentsOrFields (~100 lines)
    - All function arguments
    - All field initializers
    - Mixed (should parse as function call)
  - TestMigration_ParseCallOrRecordLiteral (~100 lines)
    - Function calls
    - Record literals
    - Edge cases
- **Benchmarks**: migration_helpers_bench_test.go (~150 lines)

## Key Design Decisions

### 1. List Navigation Strategy
**Question**: How to handle parseSeparatedListBeforeStart in cursor mode?
**Options**:
- A) Create cursor version of parseSeparatedListBeforeStart
- B) Inline the logic in parseExpressionListCursor
**Decision**: Start with B (inline) for simplicity, refactor to A if needed

### 2. State Synchronization
**Question**: When to sync cursor ↔ tokens?
**Strategy**:
- No sync during cursor-to-cursor calls
- Sync at entry for backward compatibility (if called from traditional code)
- Sync at exit for backward compatibility (caller might use curToken)

### 3. Error Handling
**Question**: How to handle errors in cursor mode?
**Strategy**:
- Same error reporting as traditional mode
- Use existing addError() mechanisms
- Position tracking via cursor.Current().Pos

## Testing Strategy

### Differential Testing
- Every test case runs in both traditional and cursor mode
- Compare ASTs for equality
- Compare error messages
- Compare position information

### Coverage Goals
- All code paths exercised
- Edge cases (empty lists, single elements, etc.)
- Error cases
- Complex nested structures

## Success Criteria

- [ ] All helper functions have cursor versions
- [ ] All tests pass in both modes
- [ ] ASTs identical between modes
- [ ] No performance regressions >15%
- [ ] All existing parser tests still pass
- [ ] Documentation updated

## Risk Mitigation

**Risk 1**: Complex list parsing logic
- **Mitigation**: Start with parseExpressionList (simplest), build confidence
- **Fallback**: Can keep traditional version longer if needed

**Risk 2**: State synchronization bugs
- **Mitigation**: Add logging/debugging output during development
- **Fallback**: Add more sync points if issues arise

**Risk 3**: Time overrun
- **Mitigation**: Implement in phases, each phase is independently valuable
- **Fallback**: Can mark later phases as follow-up tasks

## Implementation Checklist

### Phase 1: Foundation
- [ ] Implement parseExpressionListCursor
- [ ] Add basic test coverage
- [ ] Verify integration

### Phase 2: Simple Helpers
- [ ] Implement parseEmptyCallCursor
- [ ] Implement parseCallWithExpressionListCursor
- [ ] Add test coverage

### Phase 3: Field Helpers
- [ ] Implement parseNamedFieldInitializerCursor
- [ ] Implement parseArgumentAsFieldInitializerCursor
- [ ] Implement parseSingleArgumentOrFieldCursor
- [ ] Implement advanceToNextItemCursor
- [ ] Add test coverage

### Phase 4: parseArgumentsOrFields
- [ ] Implement parseArgumentsOrFieldsCursor
- [ ] Add comprehensive test coverage
- [ ] Test edge cases

### Phase 5: Orchestrator
- [ ] Implement parseCallOrRecordLiteralCursor
- [ ] Integration testing
- [ ] End-to-end testing

### Phase 6: Testing & Documentation
- [ ] Create migration_helpers_test.go
- [ ] Create migration_helpers_bench_test.go
- [ ] Run all tests
- [ ] Performance benchmarks
- [ ] Update PLAN.md

## Notes

- This is a large task, but each phase is independently testable
- Can pause after any phase and resume later
- Early phases (1-2) provide immediate value
- Later phases (4-5) unlock full cursor mode for call expressions
