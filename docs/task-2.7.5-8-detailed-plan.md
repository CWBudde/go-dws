# Tasks 2.7.5-2.7.8: Detailed Implementation Plan

## Overview

**Goal**: Convert all 756 `curToken`/`peekToken` references to cursor equivalents, enabling Traditional mode removal.

**Current State** (2025-11-19):
```
Total: 756 references across 23 files
Largest files:
  126 - expressions.go
   56 - control_flow.go
   55 - parser.go
   54 - interfaces.go
   50 - classes.go
   49 - functions.go
   48 - records.go
```

## Task 2.7.5: Convert Small/Utility Files (8h)

Convert helper and utility files first to establish patterns.

### 2.7.5.1: helpers.go (16 refs) - 1.5h
**Current**: 16 references
**Actions**:
1. Convert `curTokenIs()` → inline `p.cursor.Current().Type ==`
2. Convert `peekTokenIs()` → inline `p.cursor.Peek(1).Type ==`
3. Convert `expectPeek()` → check + `cursor.Advance()`
4. Convert `expectCurrent()` → check `cursor.Current()`
5. Convert position helpers

**Verification**: `grep -c "p\.curToken\|p\.peekToken" internal/parser/helpers.go` = 0

### 2.7.5.2: Utility files (12 refs total) - 1.5h
- `node_builder.go` (4 refs): Position tracking
- `cursor.go` (4 refs): Cursor state management
- `context.go` (2 refs): Context tracking
- `parser_builder.go` (2 refs): Parser initialization

**Actions**: Convert position/state references to cursor equivalents

### 2.7.5.3: Small parser files (44 refs total) - 3h
- `sets.go` (10 refs)
- `declarations.go` (12 refs)
- `arrays.go` (22 refs)

**Actions**: Systematically convert each file, focusing on non-Traditional functions

### 2.7.5.4: Test and fix (2h)
- Run full test suite
- Fix any breakages
- Verify baseline maintained

**Deliverable**: 72 references converted (756 → 684)

---

## Task 2.7.6: Convert Medium Parser Files (16h)

Focus on core parsing files with moderate complexity.

### 2.7.6.1: enums.go (22 refs) - 2h
**Actions**:
- Skip `parseEnumDeclarationTraditional` and `parseEnumValueTraditional`
- Convert all Cursor and non-Traditional functions
- Verify tests pass

### 2.7.6.2: types.go (22 refs) - 2h
**Actions**:
- Skip Traditional functions (`parseArrayDeclarationTraditional`, etc.)
- Convert Cursor functions
- Update type checking logic

### 2.7.6.3: properties.go (24 refs) - 2h
**Actions**:
- Convert property parsing functions
- Update accessor/mutator parsing

### 2.7.6.4: combinators.go (31 refs) - 3h
**Actions**:
- Convert combinator helper functions
- Update parser state management in combinators
- Careful testing (combinators used everywhere)

### 2.7.6.5: operators.go (34 refs) - 3h
**Actions**:
- Skip `parseOperatorDeclarationTraditional`
- Convert Cursor and non-Traditional operator parsing
- Update precedence checking

### 2.7.6.6: exceptions.go (36 refs) - 2h
**Actions**:
- Convert try/catch/finally parsing
- Update exception handler parsing
- Careful with error recovery code

### 2.7.6.7: Test and fix (2h)
- Run full test suite after each file
- Fix breakages immediately
- Maintain test baseline

**Deliverable**: 169 references converted (684 → 515)

---

## Task 2.7.7: Convert Large Parser Files Part 1 (12h)

Tackle the largest, most complex files.

### 2.7.7.1: statements.go (37 refs) - 3h
**Actions**:
- Skip Traditional functions
- Convert statement dispatch logic
- Update var/const declaration parsing
- Careful with statement block parsing

**Key functions**:
- `parseVarDeclaration` dispatcher
- `parseConstDeclaration` dispatcher
- `parseAssignmentOrExpression`

### 2.7.7.2: unit.go (38 refs) - 3h
**Actions**:
- Skip `parseUnitTraditional`, `parseUsesClauseTraditional`
- Convert unit section parsing
- Update uses clause handling
- Note: `parseUnitCursor` currently delegates to Traditional

### 2.7.7.3: records.go (48 refs) - 3h
**Actions**:
- Skip all Traditional functions (multiple)
- Convert Cursor record parsing
- Update field/method/property parsing in Cursor mode
- Record parsing is complex - proceed carefully

### 2.7.7.4: Test and fix (3h)
- Comprehensive testing after each file
- Fix any regressions
- Document any edge cases discovered

**Deliverable**: 123 references converted (515 → 392)

---

## Task 2.7.8: Convert Large Parser Files Part 2 (12h)

Complete the conversion of the largest files.

### 2.7.8.1: functions.go (49 refs) - 4h
**Actions**:
- Skip `parseFunctionDeclarationTraditional` and related Traditional functions
- Convert Cursor function parsing
- Update parameter list parsing
- Update function signature handling
- Complex file - allocate extra time

### 2.7.8.2: classes.go (50 refs) - 4h
**Actions**:
- Skip all Traditional class functions
- Convert Cursor class parsing
- Update field/method/property declarations
- Update inheritance/interface handling
- Very complex file - careful testing needed

### 2.7.8.3: interfaces.go (54 refs) - 4h
**Actions**:
- Skip Traditional interface functions
- Convert interface parsing to cursor
- Update method declaration parsing
- Note: Currently has 2 sync delegation points

### 2.7.8.4: parser.go (55 refs) - 5h
**Actions**:
- Core parser infrastructure
- Convert main parse loop
- Update error recovery
- Update synchronization
- Critical file - extensive testing required

### 2.7.8.5: control_flow.go (56 refs) - 4h
**Actions**:
- Convert if/while/for/case statement parsing
- Update conditional expression handling
- Update loop body parsing
- Already mostly Cursor-based, but has lingering curToken refs

### 2.7.8.6: expressions.go (126 refs) - 8h
**LARGEST FILE - MOST COMPLEX**

**Actions**:
- Skip all Traditional expression functions
- Convert prefix/infix expression handlers
- Update binary/unary operator parsing
- Update literal parsing (integer, float, string, etc.)
- Update call/index/member access
- Split into sub-phases:
  - Phase 1: Literal parsing (2h)
  - Phase 2: Binary/unary operators (2h)
  - Phase 3: Call/index/member (2h)
  - Phase 4: Complex expressions (2h)

**Special attention**:
- Many expression functions use both curToken and cursor
- Careful synchronization required
- Extensive test coverage needed

### 2.7.8.7: Final testing (4h)
- Run complete test suite
- Run fixture tests
- Performance benchmarking
- Document any remaining issues

**Deliverable**: All 392 remaining references converted (392 → ~50 Traditional-only)

---

## Task 2.7.9: Switch to Cursor-Only Mode (4h)

Now that all non-Traditional code uses cursor, make cursor the only mode.

### 2.7.9.1: Remove useCursor flag (1h)
- Remove `useCursor` field from Parser struct
- Remove useCursor checks from all files
- Update parser initialization

### 2.7.9.2: Update dispatchers (2h)
- Remove dual-mode dispatch logic
- Direct call to Cursor functions everywhere
- Remove sync calls

### 2.7.9.3: Test (1h)
- Full test suite
- Verify all tests pass
- Traditional code now unreachable

---

## Task 2.7.10: Comprehensive Testing (12h)

Validate cursor-only parser extensively.

### 2.7.10.1: Unit tests (3h)
- 100% pass rate required
- Add missing edge case tests

### 2.7.10.2: Integration tests (3h)
- Complex nested structures
- Real-world DWScript samples

### 2.7.10.3: Fixture tests (4h)
- Run ~2,100 DWScript test suite
- Document failures
- Fix critical issues

### 2.7.10.4: Performance testing (2h)
- Benchmark cursor vs traditional
- Verify no performance regression
- Document improvements

---

## Task 2.7.11: Update Call Sites (Now Unblocked) (4h)

With cursor-only mode active, remove dispatch wrappers.

**See**: task-2.7.11-analysis.md and PLAN.md task 2.7.11 for detailed subtasks.

---

## Task 2.7.12: Remove Dual-Mode Infrastructure (4h)

Clean up dual-mode artifacts.

- Remove `useCursor` field
- Remove sync methods
- Remove dual-mode test infrastructure

---

## Task 2.7.13: Remove Traditional Functions (8h)

Final cleanup - remove all *Traditional functions.

- 50+ Traditional functions to remove
- Verify no callers remain
- Update documentation

---

## Total Effort

- Task 2.7.5: 8h (72 refs)
- Task 2.7.6: 16h (169 refs)
- Task 2.7.7: 12h (123 refs)
- Task 2.7.8: 33h (392 refs)
- Task 2.7.9: 4h
- Task 2.7.10: 12h
- Task 2.7.11: 4h (now unblocked)
- Task 2.7.12: 4h
- Task 2.7.13: 8h

**Total: 101 hours** (originally estimated 80h, revised due to 756 refs vs 308 estimated)

---

## Execution Strategy

1. **Start small**: helpers.go, utility files (Task 2.7.5)
2. **Build momentum**: Medium files (Task 2.7.6)
3. **Tackle large files**: systematic approach (Tasks 2.7.7-2.7.8)
4. **Test continuously**: After each file, run tests
5. **Fix immediately**: Don't accumulate technical debt
6. **Document**: Track progress, note issues

## Success Criteria

- [ ] All 756 curToken/peekToken references converted (excluding Traditional functions)
- [ ] All tests pass with cursor-only mode
- [ ] No performance regression
- [ ] Traditional mode completely unreachable
- [ ] Ready for Traditional function removal (task 2.7.13)
