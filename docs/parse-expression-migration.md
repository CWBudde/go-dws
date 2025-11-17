# parseExpression Migration Design Document

**Task**: 2.2.5b - Migrate parseExpression to Cursor Mode
**Date**: 2025-01-XX
**Status**: DESIGN PHASE
**Complexity**: HIGH - This is the heart of the Pratt parser

## Overview

parseExpression is the core expression dispatcher in the DWScript parser. It implements a Pratt parser (precedence climbing) and is called recursively by virtually all expression parsing functions. Migrating it to cursor mode is the critical path for completing cursor-based expression parsing.

## Current Implementation Analysis

### Location
- **File**: `internal/parser/expressions.go`
- **Lines**: 15-76 (~62 lines)
- **Function Signature**: `func (p *Parser) parseExpression(precedence int) ast.Expression`

### Preconditions/Postconditions
- **PRE**: `curToken` is first token of expression
- **POST**: `curToken` is last token of expression

### Algorithm Overview

```go
func (p *Parser) parseExpression(precedence int) ast.Expression {
    // 1. Lookup and call prefix function based on curToken
    prefix := p.prefixParseFns[p.curToken.Type]
    if prefix == nil {
        p.noPrefixParseFnError(p.curToken.Type)
        return nil
    }
    leftExp := prefix()

    // 2. Loop while precedence allows and not at semicolon
    for !p.peekTokenIs(lexer.SEMICOLON) &&
        (precedence < p.peekPrecedence() ||
         (p.peekTokenIs(lexer.NOT) && precedence < EQUALS)) {

        // 3. SPECIAL CASE: "not in/is/as" operators
        if p.peekTokenIs(lexer.NOT) && precedence < EQUALS {
            // Manual backtracking with state save/restore
            savedCurToken := p.curToken
            savedPeekToken := p.peekToken

            p.nextToken() // move to NOT
            notToken := p.curToken

            if p.peekTokenIs(lexer.IN) || p.peekTokenIs(lexer.IS) || p.peekTokenIs(lexer.AS) {
                // Parse as "not (x in set)"
                p.nextToken() // move to IN/IS/AS
                infix := p.infixParseFns[p.curToken.Type]
                if infix != nil {
                    comparisonExp := infix(leftExp)
                    leftExp = wrapInNotExpression(notToken, comparisonExp)
                    continue
                }
            }

            // Not a "not in/is/as", restore and exit
            p.curToken = savedCurToken
            p.peekToken = savedPeekToken
            return leftExp
        }

        // 4. Normal infix operator handling
        infix := p.infixParseFns[p.peekToken.Type]
        if infix == nil {
            return leftExp
        }

        p.nextToken()
        leftExp = infix(leftExp)
    }

    return leftExp
}
```

## Key Components and Dependencies

### 1. Prefix Parse Functions

**Storage**: `p.prefixParseFns map[lexer.TokenType]prefixParseFn`

**Function Signature**: `type prefixParseFn func() ast.Expression`

**Registered Functions**:
- `parseIdentifier` - Variables, function names
- `parseIntegerLiteral` - Integer constants
- `parseFloatLiteral` - Float constants
- `parseStringLiteral` - String constants
- `parseBooleanLiteral` - true/false
- `parseNilLiteral` - nil
- `parsePrefixExpression` - Unary operators: -, not, +
- `parseGroupedExpression` - Parenthesized expressions
- `parseArrayLiteral` - Array literals
- `parseNewExpression` - Object creation
- `parseLambdaExpression` - Lambda functions
- Many more...

**Current Behavior**: Called with no arguments, relies on p.curToken internally.

**Cursor Challenge**: These functions need to know what token to parse. Currently they access p.curToken directly.

### 2. Infix Parse Functions

**Storage**: `p.infixParseFns map[lexer.TokenType]infixParseFn`

**Function Signature**: `type infixParseFn func(ast.Expression) ast.Expression`

**Registered Functions**:
- `parseInfixExpression` - Binary operators: +, -, *, /, =, <, >, and, or, etc.
- `parseCallExpression` - Function calls: f(args)
- `parseMemberAccess` - Member access: obj.field, obj.method()
- `parseIndexExpression` - Array indexing: arr[i]
- `parseIsExpression` - Type checking: obj is TClass
- `parseAsExpression` - Type casting: obj as IInterface
- More...

**Current Behavior**: Called with left expression, relies on p.curToken for operator token.

**Cursor Challenge**: These functions need access to the operator token and must advance cursor.

### 3. Precedence Lookup

**Data Structure**: `var precedences = map[lexer.TokenType]int`

**Helper Functions**:
- `p.curPrecedence() int` - Returns precedence of p.curToken
- `p.peekPrecedence() int` - Returns precedence of p.peekToken

**Cursor Challenge**: Need cursor-aware versions that take token type as parameter.

### 4. Token Checking Helpers

**Functions Used**:
- `p.peekTokenIs(t TokenType) bool` - Checks if peekToken.Type == t
- `p.nextToken()` - Advances curToken/peekToken

**Cursor Challenge**: Need cursor-aware equivalents.

## Challenges for Cursor Migration

### Challenge 1: Prefix/Infix Function Maps Are Shared

**Problem**: Both traditional and cursor modes share the same `prefixParseFns` and `infixParseFns` maps.

**Why It's a Problem**:
- Traditional prefix functions access `p.curToken` directly
- Cursor prefix functions would need to access `p.cursor.Current()`
- We can't have both in the same map during migration

**Possible Solutions**:

**Option A: Dual Function Maps** (Recommended)
```go
type Parser struct {
    // Traditional (existing)
    prefixParseFns map[lexer.TokenType]prefixParseFn
    infixParseFns  map[lexer.TokenType]infixParseFn

    // Cursor (new)
    prefixParseFnsCursor map[lexer.TokenType]prefixParseFnCursor
    infixParseFnsCursor  map[lexer.TokenType]infixParseFnCursor
}

// Cursor-specific function types (take token explicitly)
type prefixParseFnCursor func(lexer.Token) ast.Expression
type infixParseFnCursor func(ast.Expression, lexer.Token) ast.Expression
```

Pros:
- Clean separation during migration
- No risk of mixing traditional/cursor functions
- Can gradually populate cursor maps

Cons:
- More memory during migration phase
- Need to maintain both registrations

**Option B: Adapter Pattern**
```go
// Wrap existing functions to work with cursor
func (p *Parser) prefixAdapterCursor(tokenType lexer.TokenType) prefixParseFnCursor {
    traditionalFn := p.prefixParseFns[tokenType]
    return func(tok lexer.Token) ast.Expression {
        // Temporarily set curToken for traditional function
        savedCur := p.curToken
        p.curToken = tok
        result := traditionalFn()
        p.curToken = savedCur
        return result
    }
}
```

Pros:
- Reuses existing functions
- Smaller migration surface

Cons:
- Hacky state manipulation
- Doesn't actually achieve cursor purity
- Hard to reason about

**Option C: Single Pass Migration with Token Parameter**
```go
// Change existing function signatures to take token
type prefixParseFn func(lexer.Token) ast.Expression

func (p *Parser) parseIdentifier(tok lexer.Token) ast.Expression {
    return &ast.Identifier{Value: tok.Literal}
}
```

Pros:
- Clean final state
- No duplication

Cons:
- Massive breaking change (all prefix/infix functions at once)
- High risk
- Against incremental migration strategy

**Decision**: **Option A - Dual Function Maps**

This aligns with our established Strangler Fig pattern. We'll:
1. Add cursor-specific function maps
2. Register cursor versions gradually
3. parseExpressionCursor uses cursor maps
4. Remove traditional maps in Phase 2.7

### Challenge 2: "not in/is/as" Backtracking

**Current Approach**: Manual state save/restore
```go
savedCurToken := p.curToken
savedPeekToken := p.peekToken

p.nextToken() // speculative advancement

// Check condition
if condition {
    // Success - use new state
} else {
    // Failure - restore old state
    p.curToken = savedCurToken
    p.peekToken = savedPeekToken
}
```

**Cursor Approach**: Use Mark/ResetTo
```go
mark := p.cursor.Mark()

// Speculative advancement
testCursor := p.cursor.Advance()

// Check condition using testCursor
if condition {
    // Success - commit by using testCursor
    p.cursor = testCursor
} else {
    // Failure - discard testCursor, cursor unchanged
    // Or explicitly: p.cursor = p.cursor.ResetTo(mark)
}
```

**Key Difference**:
- Traditional: Mutates state, must manually restore
- Cursor: Immutable, new cursors are created, restoration is implicit

**Implementation Strategy**:
```go
// In parseExpressionCursor:
if p.cursor.Peek(1).Type == lexer.NOT && precedence < EQUALS {
    // Look ahead: cursor -> NOT -> IN/IS/AS?
    mark := p.cursor.Mark()
    testCursor := p.cursor.Advance()  // at NOT
    notToken := testCursor.Current()

    nextToken := testCursor.Peek(1)
    if nextToken.Type == lexer.IN || nextToken.Type == lexer.IS || nextToken.Type == lexer.AS {
        // This is "not in/is/as"
        testCursor = testCursor.Advance()  // at IN/IS/AS
        operatorToken := testCursor.Current()

        // Parse infix using cursor version
        if infixFn, ok := p.infixParseFnsCursor[operatorToken.Type]; ok {
            testCursor = testCursor.Advance()  // move past operator

            // Sync for recursive parseExpression call (temporary)
            p.cursor = testCursor
            p.syncCursorToTokens()

            comparisonExp := infixFn(leftExp, operatorToken)
            leftExp = wrapInNotExpression(notToken, comparisonExp)
            continue
        }
    }

    // Not "not in/is/as" - cursor unchanged (automatic rollback)
    return leftExp
}
```

### Challenge 3: Loop Condition and Advancement

**Traditional Loop**:
```go
for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
    // ...
    p.nextToken()
    leftExp = infix(leftExp)
}
```

**Cursor Loop**:
```go
for {
    nextToken := p.cursor.Peek(1)

    // Check termination conditions
    if nextToken.Type == lexer.SEMICOLON {
        break
    }

    nextPrec := getPrecedence(nextToken.Type)
    if precedence >= nextPrec {
        break
    }

    // Normal infix handling
    operatorType := nextToken.Type
    infixFn, ok := p.infixParseFnsCursor[operatorType]
    if !ok {
        break
    }

    p.cursor = p.cursor.Advance()  // move to operator
    operatorToken := p.cursor.Current()

    p.cursor = p.cursor.Advance()  // move past operator
    p.syncCursorToTokens()  // for recursive parseExpression (temporary)

    leftExp = infixFn(leftExp, operatorToken)
}
```

**Key Changes**:
- Peek ahead to check conditions before advancing
- Explicit cursor advancement
- Break conditions are more explicit

### Challenge 4: Recursive Calls

**Problem**: parseExpression calls itself recursively through infix handlers:

```
parseExpression
  └─> parseInfixExpression
      └─> parseExpression (for right operand)
          └─> parseInfixExpression
              └─> ...
```

**Current State**:
- parseExpression: Traditional only
- parseInfixExpression: Has cursor version, but calls traditional parseExpression

**Ideal State (Post-Migration)**:
- parseExpression: Cursor version exists
- parseInfixExpression cursor version calls parseExpression cursor version

**During Migration**:

We'll have a hybrid state:

```go
func (p *Parser) parseInfixExpressionCursor(left ast.Expression, operatorToken lexer.Token) ast.Expression {
    // ...

    // TODO: Call cursor version once it exists
    // For now, sync and call traditional
    p.syncCursorToTokens()
    right := p.parseExpression(precedence)

    // After Task 2.2.5b complete:
    // right := p.parseExpressionCursor(precedence)
}
```

**Post-Migration Plan**:
1. Implement parseExpressionCursor (Task 2.2.5b)
2. Update parseInfixExpressionCursor to call parseExpressionCursor
3. Update all other infix cursor functions similarly
4. Remove hybrid workarounds

## Proposed Cursor Version Design

### Function Signature

```go
// parseExpressionCursor parses an expression with the given precedence using cursor navigation.
// PRE: cursor at first token of expression
// POST: cursor at last token of expression
func (p *Parser) parseExpressionCursor(precedence int) ast.Expression
```

### High-Level Structure

```go
func (p *Parser) parseExpressionCursor(precedence int) ast.Expression {
    // 1. Lookup and call prefix function
    currentToken := p.cursor.Current()
    prefixFn, ok := p.prefixParseFnsCursor[currentToken.Type]
    if !ok {
        p.noPrefixParseFnError(currentToken.Type)
        return nil
    }

    leftExp := prefixFn(currentToken)

    // 2. Main precedence climbing loop
    for {
        nextToken := p.cursor.Peek(1)

        // Termination: semicolon
        if nextToken.Type == lexer.SEMICOLON {
            break
        }

        // Termination: precedence
        nextPrec := getPrecedence(nextToken.Type)
        notInIsAsAllowed := nextToken.Type == lexer.NOT && precedence < EQUALS
        if precedence >= nextPrec && !notInIsAsAllowed {
            break
        }

        // 3. Special case: "not in/is/as"
        if nextToken.Type == lexer.NOT && precedence < EQUALS {
            leftExp = p.parseNotInIsAs(leftExp)
            if leftExp == nil {
                return nil  // Error or not a "not in/is/as" pattern
            }
            continue
        }

        // 4. Normal infix handling
        infixFn, ok := p.infixParseFnsCursor[nextToken.Type]
        if !ok {
            break
        }

        p.cursor = p.cursor.Advance()  // move to operator
        operatorToken := p.cursor.Current()

        // Infix function handles advancing cursor and parsing right side
        leftExp = infixFn(leftExp, operatorToken)
    }

    return leftExp
}
```

### Helper Function: getPrecedence

```go
// getPrecedence returns the precedence for a given token type (cursor-aware).
func getPrecedence(tokenType lexer.TokenType) int {
    if prec, ok := precedences[tokenType]; ok {
        return prec
    }
    return LOWEST
}
```

### Helper Function: parseNotInIsAs

```go
// parseNotInIsAs handles the "not in/is/as" special case with cursor.
// Returns the wrapped expression if successful, nil otherwise.
func (p *Parser) parseNotInIsAs(leftExp ast.Expression) ast.Expression {
    // Current: leftExp parsed
    // cursor.Peek(1): NOT
    // cursor.Peek(2): possibly IN/IS/AS

    notCursor := p.cursor.Advance()  // at NOT
    notToken := notCursor.Current()

    nextToken := notCursor.Peek(1)
    if nextToken.Type != lexer.IN && nextToken.Type != lexer.IS && nextToken.Type != lexer.AS {
        // Not a "not in/is/as" pattern, return nil
        return nil
    }

    // This is "not in/is/as"
    operatorCursor := notCursor.Advance()  // at IN/IS/AS
    operatorToken := operatorCursor.Current()

    infixFn, ok := p.infixParseFnsCursor[operatorToken.Type]
    if !ok {
        return nil  // Shouldn't happen, but handle gracefully
    }

    // Parse the infix expression
    p.cursor = operatorCursor
    comparisonExp := infixFn(leftExp, operatorToken)

    // Wrap in NOT
    leftExp := &ast.UnaryExpression{
        TypedExpressionBase: ast.TypedExpressionBase{
            BaseNode: ast.BaseNode{
                Token:  notToken,
                EndPos: comparisonExp.End(),
            },
        },
        Operator: notToken.Literal,
        Right:    comparisonExp,
    }

    return leftExp
}
```

## Migration Risks and Mitigation

### Risk 1: Breaking Existing Tests

**Risk**: parseExpression is called by hundreds of tests. Changing it could break many tests.

**Mitigation**:
- Use dispatcher pattern: `parseExpression` routes to Traditional or Cursor
- Keep traditional version unchanged initially
- Cursor version is opt-in via NewCursorParser()
- Run full test suite in both modes

### Risk 2: Subtle Behavioral Differences

**Risk**: Cursor version might have subtle differences in edge cases.

**Mitigation**:
- Comprehensive differential testing
- Test with existing test suite (all 49 test files)
- Document any discovered differences
- Fuzz testing with random expressions

### Risk 3: Performance Regression

**Risk**: Cursor version might be slower due to immutability overhead.

**Mitigation**:
- Benchmark against traditional version
- Cursor is faster for backtracking scenarios (proven in Task 2.2.1)
- Acceptable overhead: <15% (established in previous tasks)
- Profile if needed

### Risk 4: Incomplete Prefix/Infix Migration

**Risk**: Not all prefix/infix functions have cursor versions yet.

**Mitigation**:
- Start with functions already migrated (Identifier, Integer, Float, String, Boolean)
- Use hybrid approach: cursor map with adapters for unmigrated functions
- Gradually migrate remaining functions in Task 2.2.5c

### Risk 5: Recursive Dependency Challenges

**Risk**: parseExpressionCursor needs infix functions, but infix functions need parseExpressionCursor.

**Current State**:
- parseInfixExpressionCursor exists but calls traditional parseExpression
- Need to update it once parseExpressionCursor exists

**Mitigation**:
- Phase 1: Implement parseExpressionCursor calling traditional infix functions
- Phase 2: Update infix cursor functions to call parseExpressionCursor
- Phase 3: Remove traditional fallbacks

## Implementation Plan

### Phase 1: Infrastructure (2-4 hours)

1. Add cursor function type definitions
2. Add cursor function maps to Parser struct
3. Add getPrecedence helper function
4. Initialize cursor maps in NewCursorParser

### Phase 2: Core Implementation (8-12 hours)

1. Implement parseExpressionTraditional (rename existing)
2. Implement parseExpressionCursor
3. Implement parseExpression dispatcher
4. Implement parseNotInIsAs helper
5. Register existing cursor prefix/infix functions

### Phase 3: Testing (6-8 hours)

1. Create migration_parse_expression_test.go
2. Test simple expressions
3. Test precedence handling
4. Test "not in/is/as" special case
5. Test error cases
6. Run full existing test suite in both modes
7. Compare AST outputs

### Phase 4: Integration (2-4 hours)

1. Update parseInfixExpressionCursor to call parseExpressionCursor
2. Verify no regressions
3. Document remaining work
4. Create benchmarks

**Total Estimate**: 18-28 hours (aligns with 20-hour estimate in PLAN.md)

## Success Criteria

1. **parseExpressionCursor exists and compiles** ✓
2. **All differential tests pass** ✓
3. **Existing test suite passes in cursor mode** ✓
4. **Performance acceptable** (<15% overhead) ✓
5. **No behavioral regressions** ✓
6. **"not in/is/as" handled correctly** ✓
7. **Documentation complete** ✓

## Open Questions

1. **Should we migrate all prefix/infix functions first, or do partial migration?**
   - **Decision**: Partial migration is fine. Use adapters for unmigrated functions.

2. **How to handle prefix/infix functions that haven't been migrated yet?**
   - **Decision**: Create adapter that sets curToken temporarily.

3. **Should parseExpressionCursor call cursor infix functions or traditional?**
   - **Decision**: Call cursor versions where they exist, use adapters for rest.

4. **What if there are edge cases we haven't considered?**
   - **Decision**: Comprehensive testing will reveal them. Fix as discovered.

## Next Steps

1. Review this design document
2. Get approval to proceed
3. Start Phase 1: Infrastructure
4. Implement Phase 2: Core
5. Validate with Phase 3: Testing
6. Integrate with Phase 4

## References

- `internal/parser/expressions.go` - Current implementation
- `internal/parser/cursor.go` - TokenCursor API
- `docs/token-cursor.md` - Cursor documentation
- `docs/dual-mode-parser.md` - Dual-mode architecture
- `docs/task-2.2.5-status.md` - Task 2.2.5 analysis

## Change Log

### 2025-01-XX: Initial Design
- Created comprehensive analysis
- Identified 5 major challenges
- Designed cursor version approach
- Proposed 4-phase implementation plan
- Estimated 18-28 hours (aligns with plan)
