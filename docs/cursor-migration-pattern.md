# Cursor Migration Pattern

## Overview

This document describes the established pattern for migrating parsing functions from traditional mutable state to cursor-based immutable navigation. It captures lessons learned from the first migration (Task 2.2.3: `parseIntegerLiteral`).

**Status**: Pattern established with `parseIntegerLiteral` as proof of concept ✓

## Migration Strategy

### Incremental Approach

We use a **side-by-side** approach during migration:

1. **Keep traditional version** (rename to `*Traditional`)
2. **Create cursor version** (new `*Cursor` function)
3. **Validate equivalence** through comprehensive testing
4. **Benchmark performance** to ensure acceptable overhead
5. **Document differences** and lessons learned
6. **(Later) Switch dispatcher** when dependencies are migrated
7. **(Eventually) Remove traditional** version when no longer needed

### Why Not Direct Replacement?

Direct replacement doesn't work because:
- Parsing functions are interdependent (e.g., `parseExpression` calls `parseIntegerLiteral`)
- Cursor and traditional state must stay synchronized
- Token navigation differs between modes
- Need to validate equivalence before switching

### Current Limitation

The cursor version of `parseIntegerLiteral` **cannot be used in production** yet because:
- `parseExpression` still uses traditional mode
- When `parseExpression` calls a cursor-based literal parser, curToken/peekToken don't get updated
- This breaks infix operator parsing

**Solution**: Migrate `parseExpression` first (Task 2.2.4), then cursor-based literals will work.

## Cursor-Based Parsing Pattern

### Function Signature

```go
// Traditional version
func (p *Parser) parseFooTraditional() ast.Expression {
    // Uses p.curToken, p.peekToken
    // May call p.nextToken()
}

// Cursor version
func (p *Parser) parseFooCursor() ast.Expression {
    // Uses p.cursor.Current(), p.cursor.Peek()
    // Never calls p.nextToken()
    // Cursor position unchanged (caller handles advancement)
}

// Public API (dispatcher - will be enabled later)
func (p *Parser) parseFoo() ast.Expression {
    // Currently: always use traditional
    return p.parseFooTraditional()

    // Future: dispatch based on mode
    // if p.useCursor {
    //     return p.parseFooCursor()
    // }
    // return p.parseFooTraditional()
}
```

### Key Differences

| Aspect | Traditional | Cursor |
|--------|-------------|--------|
| Token access | `p.curToken` | `p.cursor.Current()` |
| Next token | `p.peekToken` | `p.cursor.Peek(1)` |
| Advancement | `p.nextToken()` | Caller advances cursor |
| State | Mutable (changes p.curToken) | Immutable (returns new cursor) |
| Position | Implicit in p.curToken | Explicit in cursor |

### Example: parseIntegerLiteral

#### Traditional Version

```go
func (p *Parser) parseIntegerLiteralTraditional() ast.Expression {
    lit := &ast.IntegerLiteral{
        TypedExpressionBase: ast.TypedExpressionBase{
            BaseNode: ast.BaseNode{
                Token:  p.curToken,  // Access current token
                EndPos: p.endPosFromToken(p.curToken),
            },
        },
    }

    literal := p.curToken.Literal  // Read from curToken

    // Parse value logic...
    lit.Value = value
    return lit
}
```

#### Cursor Version

```go
func (p *Parser) parseIntegerLiteralCursor() ast.Expression {
    currentToken := p.cursor.Current()  // Explicitly get current token

    lit := &ast.IntegerLiteral{
        TypedExpressionBase: ast.TypedExpressionBase{
            BaseNode: ast.BaseNode{
                Token:  currentToken,  // Use explicit token
                EndPos: p.endPosFromToken(currentToken),
            },
        },
    }

    literal := currentToken.Literal  // Read from explicit token

    // Parse value logic... (identical)
    lit.Value = value
    return lit
}
```

#### Key Changes

1. `p.curToken` → `p.cursor.Current()`
2. Store token in local variable for clarity
3. All other logic identical

### Performance Impact

From `parseIntegerLiteral` benchmarks:

| Metric | Traditional | Cursor | Difference |
|--------|-------------|--------|------------|
| Time | 4,756 ns/op | 5,100 ns/op | +7% slower |
| Memory | 3,504 B/op | 3,744 B/op | +7% more |
| Allocations | 68 allocs/op | 71 allocs/op | +3 allocs |

**Analysis**:
- Overhead is acceptable (<10% target from PLAN.md)
- Mostly from cursor initialization and token buffer
- Will improve when traditional mode removed
- Once `parseExpression` migrated, cursor benefits will show

## Testing Pattern

### Differential Testing

Create dedicated test file `migration_*_test.go` for each migrated function:

```go
func TestMigration_Foo_Basic(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected SomeValue
    }{
        // Test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test traditional
            traditionalParser := New(lexer.New(tt.input))
            traditionalResult := traditionalParser.parseFooTraditional()

            // Test cursor
            cursorParser := NewCursorParser(lexer.New(tt.input))
            cursorResult := cursorParser.parseFooCursor()

            // Validate equivalence
            if traditionalResult != cursorResult {
                t.Errorf("Mismatch: traditional=%v, cursor=%v",
                    traditionalResult, cursorResult)
            }
        })
    }
}
```

### Test Categories

For each migrated function, test:

1. **Basic cases**: Common inputs, expected outputs
2. **Edge cases**: Empty, boundary values, special characters
3. **Error handling**: Invalid inputs, overflow, malformed syntax
4. **Position tracking**: Verify token positions match
5. **AST structure**: Verify identical AST nodes
6. **Dispatcher**: Verify dispatcher calls correct implementation

### Benchmark Pattern

Create dedicated benchmark file `migration_*_bench_test.go`:

```go
func BenchmarkFoo_Traditional(b *testing.B) {
    source := "test input"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p := New(lexer.New(source))
        _ = p.parseFooTraditional()
    }
}

func BenchmarkFoo_Cursor(b *testing.B) {
    source := "test input"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p := NewCursorParser(lexer.New(source))
        _ = p.parseFooCursor()
    }
}

func BenchmarkFoo_Memory_Traditional(b *testing.B) {
    source := "test input"
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p := New(lexer.New(source))
        _ = p.parseFooTraditional()
    }
}

func BenchmarkFoo_Memory_Cursor(b *testing.B) {
    source := "test input"
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p := NewCursorParser(lexer.New(source))
        _ = p.parseFooCursor()
    }
}
```

## Implementation Checklist

For each function to migrate:

- [ ] Read and understand current implementation
- [ ] Create `parse*Cursor()` version
- [ ] Keep traditional as `parse*Traditional()`
- [ ] Add `parse*()` dispatcher (initially calls Traditional)
- [ ] Create `migration_*_test.go` with comprehensive tests
- [ ] Create `migration_*_bench_test.go` with benchmarks
- [ ] Run tests, verify all pass
- [ ] Run benchmarks, verify acceptable performance
- [ ] Document any special cases or gotchas
- [ ] Update PLAN.md checklist

## Migration Order (from PLAN.md)

1. ✓ **Task 2.2.3**: `parseIntegerLiteral` (DONE - proof of concept)
2. **Task 2.2.4**: Expression parsing
   - `parseIdentifier`
   - `parseStringLiteral`
   - `parseBooleanLiteral`
   - `parseFloatLiteral`
   - `parseGroupedExpression`
3. **Task 2.2.5**: Infix expressions
   - `parseBinaryExpression`
   - `parseCallExpression`
   - `parseMemberAccess`
   - `parseIndexExpression`

## Lessons Learned

### From parseIntegerLiteral Migration

1. **Token Access**: Use `cursor.Current()` instead of `p.curToken`
2. **Immutability**: Cursor doesn't change; caller advances it
3. **Dependencies**: Can't switch to cursor until callers support it
4. **Testing**: Differential testing catches subtle differences
5. **Performance**: ~7% overhead acceptable for proof of concept
6. **Integration**: Need top-down migration (expressions before literals)

### Common Pitfalls

1. **Don't mutate cursor**: It returns new instances
2. **Don't mix modes**: Cursor and traditional state can desync
3. **Don't skip tests**: Comprehensive testing is essential
4. **Don't optimize early**: Focus on correctness first
5. **Don't rush integration**: Validate before switching dispatcher

### Best Practices

1. **Test exhaustively**: Cover all cases (success, error, edge)
2. **Benchmark everything**: Track performance impact
3. **Document changes**: Explain why, not just what
4. **Keep code similar**: Minimize differences between versions
5. **Validate equivalence**: AST structure must be identical

## Future Enhancements

### When Expression Parsing Migrated (Task 2.2.4)

Once `parseExpression` uses cursor mode:
- Can enable cursor-based literal parsers
- Dispatcher can switch based on `p.useCursor`
- Traditional state (curToken/peekToken) becomes derived from cursor
- Navigation becomes truly immutable

### When All Parsing Migrated (Phase 2.7)

Once all functions migrated:
- Remove `*Traditional()` functions
- Remove `useCursor` flag
- Remove `curToken`/`peekToken` fields
- Simplify Parser struct
- Remove dual-mode infrastructure

## References

- **Proof of Concept**: `internal/parser/expressions.go` (parseIntegerLiteralCursor)
- **Tests**: `internal/parser/migration_integer_literal_test.go`
- **Benchmarks**: `internal/parser/migration_integer_literal_bench_test.go`
- **Plan**: `PLAN.md` Task 2.2.3
- **Architecture**: `docs/dual-mode-parser.md`
- **Cursor Design**: `docs/token-cursor.md`

## Changelog

### 2024-01-XX: Initial Pattern (Task 2.2.3)

- Established side-by-side migration pattern
- Migrated `parseIntegerLiteral` as proof of concept
- Created comprehensive tests (6 test categories, all passing)
- Benchmarked performance (+7% overhead, acceptable)
- Documented pattern in this file

### Next: Task 2.2.4

- Migrate expression parsing functions
- Enable cursor integration once expressions support it
- Switch dispatcher for integer literals
