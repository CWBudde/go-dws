# Dual-Mode Parser Architecture

This document describes the dual-mode parser architecture implemented in go-dws Phase 2, which supports both traditional token-based parsing and modern cursor-based parsing.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [TokenCursor Design](#tokencursor-design)
- [Statement Migration](#statement-migration)
- [Performance Characteristics](#performance-characteristics)
- [Future Optimizations](#future-optimizations)

## Overview

The dual-mode parser enables gradual migration from mutable token-based parsing to immutable cursor-based parsing while maintaining 100% backward compatibility and semantic equivalence.

### Key Benefits

1. **Immutability**: Cursor mode eliminates shared mutable state
2. **Better Error Handling**: Structured errors with context and suggestions
3. **Incremental Migration**: Each statement type migrated independently
4. **Zero Regression Risk**: Traditional mode remains unchanged
5. **Testing**: Differential testing validates semantic equivalence

### Design Principles

- **Semantic Equivalence**: Both modes must produce identical ASTs
- **No Breaking Changes**: Existing code continues to work
- **Gradual Migration**: One statement type at a time
- **Test-Driven**: Comprehensive tests for each migration
- **Performance Acceptable**: Correctness prioritized over speed initially

## Architecture

### Parser Structure

```
Parser
├── Traditional Mode (mutable)
│   ├── curToken, peekToken (mutable state)
│   ├── nextToken() (advances state)
│   └── parseStatement() (original implementation)
│
└── Cursor Mode (immutable)
    ├── cursor (TokenCursor - immutable)
    ├── cursor.Advance() (returns new cursor)
    └── parseStatementCursor() (new implementation)
```

## TokenCursor Design

The TokenCursor provides immutable token navigation:

```go
type TokenCursor struct {
    tokens  []*lexer.Token  // All tokens (read-only)
    current int              // Current position
}

// Immutable operations
func (c TokenCursor) Current() *lexer.Token
func (c TokenCursor) Peek(n int) *lexer.Token
func (c TokenCursor) Advance() TokenCursor     // Returns NEW cursor
```

### Performance Trade-offs

**Current Implementation:**
- Creates new cursor on every Advance()
- ~2x slower than traditional mode
- ~3x more memory allocations

**Benefits:**
- No concurrency issues
- No state corruption bugs
- Better error recovery

## Statement Migration

### Completed Migrations

| Statement Type | Task | Status | Tests |
|---------------|------|--------|-------|
| Infrastructure | 2.2.14.1 | ✅ Complete | 9 |
| Expression/Assignment | 2.2.14.2 | ✅ Complete | 70 |
| Block (begin/end) | 2.2.14.3 | ✅ Complete | 27 |
| If/While/Repeat | 2.2.14.4 | ✅ Complete | 37 |
| Var/Const Declarations | 2.2.14.6 | ✅ Complete | 61 |
| Try/Raise Exceptions | 2.2.14.7 | ✅ Complete | 38 |
| Break/Continue/Exit | 2.2.14.8 | ✅ Complete | 25 |
| Integration Testing | 2.2.14.9 | ✅ Complete | 28 |

### Pending: For/Case Statements (Task 2.2.14.5)

## Performance Characteristics

### Current Measurements

**Traditional Mode:**
- Time: 14,559 ns/op
- Memory: 8,664 B/op
- Allocations: 132 allocs/op

**Cursor Mode:**
- Time: 29,038 ns/op
- Memory: 25,320 B/op
- Allocations: 253 allocs/op

**Overhead:**
- Time: +99.4% (2x slower)
- Memory: +192.2% (3x more)
- Allocations: +91.6% (2x more)

## Future Optimizations

### Target: <15% overhead

**Optimization Strategies:**
1. Eliminate state synchronization
2. Make TokenCursor mutable or use pooling
3. Inline hot path operations
4. Complete expression migration

**Estimated Effort:** 20-30 hours

## Conclusion

The dual-mode parser provides safe, incremental migration to modern immutable design. Current implementation prioritizes correctness with clear path to performance optimization.

**Status:** 8/9 statement types migrated, 100% tests passing
