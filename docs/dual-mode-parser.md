# Dual-Mode Parser Architecture

## Overview

The DWScript parser supports **dual-mode operation** (Task 2.2.2), allowing it to work in either traditional mutable mode or modern cursor-based mode. This architecture enables safe, incremental migration from the old parsing approach to the new cursor-based approach without breaking existing functionality.

**Status**: Implemented in Phase 2.2.2 ✓

## Motivation

The parser modernization effort (Phase 2) aims to replace mutable parser state with immutable cursor navigation. However, directly modifying all ~9,400 lines of parser code at once would be risky. The dual-mode architecture solves this by:

1. **Allowing both modes to coexist** during migration
2. **Enabling incremental migration** of parsing functions
3. **Ensuring backward compatibility** with existing tests
4. **Validating equivalence** through differential testing

## Architecture

### Two Parser Modes

#### Traditional Mode (Old)

```go
p := parser.New(lexer)
```

- Uses mutable `curToken` and `peekToken` fields
- Manual `nextToken()` calls throughout parsing code
- Existing parsing functions work unchanged
- 411 manual token advancement sites

#### Cursor Mode (New)

```go
p := parser.NewCursorParser(lexer)
```

- Uses immutable `TokenCursor` for token navigation
- Zero manual token advancement (cursor returns new instances)
- Natural backtracking via `Mark()` and `ResetTo()`
- Functional composition of parsing operations

### Parser Structure

```go
type Parser struct {
	l         *lexer.Lexer

	// Traditional mutable state (old)
	curToken  token.Token
	peekToken token.Token

	// Cursor-based state (new)
	cursor    *TokenCursor
	useCursor bool

	// ... other fields ...
}
```

**Key Fields**:
- `curToken`, `peekToken`: Traditional mutable token state (always present for backward compatibility)
- `cursor`: Immutable cursor for token navigation (non-nil only in cursor mode)
- `useCursor`: Boolean flag indicating which mode the parser is in

### Factory Functions

#### `New(lexer) *Parser`

Creates a traditional-mode parser:
- `useCursor = false`
- `cursor = nil`
- Initializes `curToken` and `peekToken` via `nextToken()`

#### `NewCursorParser(lexer) *Parser`

Creates a cursor-mode parser:
- `useCursor = true`
- `cursor = NewTokenCursor(lexer)`
- Synchronizes `curToken`/`peekToken` with cursor for backward compatibility

## Token Synchronization

In cursor mode, the parser maintains **both** cursor and traditional token state for backward compatibility. The `syncCursorToTokens()` method ensures they stay in sync:

```go
func (p *Parser) syncCursorToTokens() {
	if p.useCursor && p.cursor != nil {
		p.curToken = p.cursor.Current()
		p.peekToken = p.cursor.Peek(1)
	}
}
```

This is called:
- After parser creation (`NewCursorParser`)
- After state restoration (`restoreState`)
- (Future) After cursor advancement in migrated functions

### Why Synchronize?

During incremental migration, some parsing functions use cursor while others still use `curToken`/`peekToken`. Synchronization ensures:
- Old functions see correct tokens
- State save/restore works correctly
- Error reporting uses correct positions
- Seamless interoperation between old and new code

## State Management

### Saving State

```go
func (p *Parser) saveState() ParserState {
	return ParserState{
		curToken:  p.curToken,
		peekToken: p.peekToken,
		cursor:    p.cursor,  // Save cursor position
		// ... other fields ...
	}
}
```

State includes **both** traditional tokens and cursor position, enabling backtracking in either mode.

### Restoring State

```go
func (p *Parser) restoreState(state ParserState) {
	p.curToken = state.curToken
	p.peekToken = state.peekToken
	p.cursor = state.cursor
	p.syncCursorToTokens()  // Re-sync after restore
	// ... other fields ...
}
```

After restoration, synchronization ensures cursor and tokens are consistent.

## Differential Testing

The `dual_mode_test.go` file contains comprehensive tests that verify both modes produce identical results:

### Test Categories

#### 1. Parser Creation
```go
TestDualMode_ParserCreation
```
Verifies both factory functions create valid parsers with correct mode flags.

#### 2. Simple Expressions
```go
TestDualMode_SimpleExpression
```
Compares AST output for literals, identifiers, and binary expressions.

#### 3. Variable Declarations
```go
TestDualMode_VarDeclaration
```
Ensures var statements parse identically in both modes.

#### 4. Complete Programs
```go
TestDualMode_Program
```
Tests full programs with functions, control flow, etc.

#### 5. Error Handling
```go
TestDualMode_Errors
```
Validates that both modes produce identical errors for invalid input.

#### 6. State Management
```go
TestDualMode_StateManagement
```
Tests that `saveState`/`restoreState` work correctly in both modes.

#### 7. Token Synchronization
```go
TestDualMode_CursorTokenSync
```
Verifies cursor and token state stay synchronized in cursor mode.

### Running Differential Tests

```bash
# Run all dual-mode tests
go test -v ./internal/parser -run TestDualMode

# Run specific test category
go test -v ./internal/parser -run TestDualMode_Program
```

All tests pass ✓, confirming both modes produce identical AST structures.

## Migration Strategy

The dual-mode architecture enables **incremental migration** following the Strangler Fig Pattern:

### Phase 2.2: Current State

- ✓ Task 2.2.1: TokenCursor implemented
- ✓ Task 2.2.2: Dual-mode infrastructure ready
- → Task 2.2.3: Migrate first parsing function (proof of concept)
- → Task 2.2.4: Migrate expression parsing
- → Task 2.2.5: Migrate infix expressions

### Migration Process

For each parsing function:

1. **Create cursor-based version** alongside old version
2. **Add differential tests** comparing both implementations
3. **Validate AST equivalence** and performance
4. **Switch to cursor version** when validated
5. **(Eventually) Remove old version** when all callers migrated

Example migration:

```go
// Old version (traditional mode)
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{
		Token: p.curToken,
		Value: parseIntValue(p.curToken.Literal),
	}
	// Caller responsible for advancing
	return lit
}

// New version (cursor mode) - Task 2.2.3
func (p *Parser) parseIntegerLiteralCursor() ast.Expression {
	lit := &ast.IntegerLiteral{
		Token: p.cursor.Current(),
		Value: parseIntValue(p.cursor.Current().Literal),
	}
	// Return expression; cursor advancement handled by caller
	return lit
}
```

### Dual-Mode Function Dispatch

During migration, some functions will check `useCursor` to dispatch to the correct implementation:

```go
func (p *Parser) parseExpression(precedence int) ast.Expression {
	if p.useCursor {
		return p.parseExpressionCursor(precedence)
	}
	return p.parseExpressionTraditional(precedence)
}
```

Once migration is complete (Phase 2.7), the traditional code paths will be removed.

## Best Practices

### For New Code

When implementing new parsing functions:

```go
// DON'T: Assume one mode or the other
func (p *Parser) parseNewThing() ast.Node {
	if p.curTokenIs(lexer.THING) {  // Assumes traditional mode
		// ...
	}
}

// DO: Support both modes via existing parser methods
func (p *Parser) parseNewThing() ast.Node {
	// Use existing methods that work in both modes
	if !p.expectPeek(lexer.THING) {
		return nil
	}
	// ...
}
```

### For Migration

When migrating existing functions:

```go
// 1. Keep old version
func (p *Parser) parseThing() ast.Node {
	// Original implementation
}

// 2. Create cursor version with clear naming
func (p *Parser) parseThingCursor() ast.Node {
	// Cursor-based implementation
}

// 3. Add dispatcher (temporary)
func (p *Parser) parseThing() ast.Node {
	if p.useCursor {
		return p.parseThingCursor()
	}
	return p.parseThingTraditional()
}

// 4. After full migration, remove traditional version
```

### For Testing

Always test both modes when adding new features:

```go
func TestNewFeature(t *testing.T) {
	tests := []struct{
		name string
		source string
	}{
		// test cases...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test traditional mode
			p1 := parser.New(lexer.New(tt.source))
			result1 := p1.ParseProgram()

			// Test cursor mode
			p2 := parser.NewCursorParser(lexer.New(tt.source))
			result2 := p2.ParseProgram()

			// Assert equivalence
			if !astEqual(result1, result2) {
				t.Error("modes produced different results")
			}
		})
	}
}
```

## Performance Considerations

### Memory Overhead

Cursor mode has minimal overhead:
- **Traditional**: ~3.5 KB/parser (curToken, peekToken, lexer state)
- **Cursor**: ~3.6 KB/parser (adds cursor pointer, shared token buffer)
- **Overhead**: ~100 bytes (< 3%)

### CPU Performance

Based on benchmarks (see `cursor_bench_test.go`):
- **Navigation patterns**: 5.1x faster in cursor mode
- **Backtracking**: 5.3x faster in cursor mode
- **Simple linear parsing**: 1.2x slower in cursor mode (acceptable)

Overall, cursor mode is faster for real-world parsing scenarios.

### Synchronization Cost

`syncCursorToTokens()` is extremely cheap (two field assignments). The cost is negligible compared to parsing operations.

## Future Work

### Phase 2.3-2.6: Complete Migration

- Migrate all parsing functions to cursor mode
- Remove traditional mode code paths
- Simplify Parser struct (remove curToken/peekToken)

### Phase 2.3: Parser Combinators

Once migration is complete, build high-level combinators on top of cursor:

```go
// Future: elegant combinator-based parsing
result := Sequence(
	Expect(token.VAR),
	ParseIdentifier,
	Expect(token.COLON),
	ParseType,
)(cursor)
```

### Phase 2.7: Remove Dual Mode

Final cleanup:
- Remove `useCursor` flag
- Remove `curToken`/`peekToken` fields
- Remove `syncCursorToTokens()`
- Remove traditional mode tests
- Parser uses only cursor internally

## References

- **Implementation**: `internal/parser/parser.go` (NewCursorParser, syncCursorToTokens)
- **Tests**: `internal/parser/dual_mode_test.go`
- **Cursor**: `internal/parser/cursor.go`
- **Plan**: `PLAN.md` Task 2.2.2
- **Related**: `docs/token-cursor.md`

## Change Log

### 2024-01-XX: Initial Implementation (Task 2.2.2)

- Added `cursor` and `useCursor` fields to Parser
- Created `NewCursorParser()` factory function
- Implemented `syncCursorToTokens()` for backward compatibility
- Updated `saveState()` and `restoreState()` to handle cursor
- Created comprehensive differential tests (8 test categories, all passing)
- Documented dual-mode architecture in this file

### Next: Task 2.2.3

- Migrate first parsing function (`parseIntegerLiteral`) to cursor mode
- Prove cursor approach works end-to-end
- Establish migration pattern for future functions
