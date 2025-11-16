# Phase 2: Parser Modernization (COMPLETE REWRITE)

**Goal**: Transform the parser from a traditional recursive descent implementation to a modern, maintainable architecture using cursor-based parsing, combinators, and structured errors.

**Current State**:
- 20 implementation files, ~9,400 lines of code
- Traditional mutable parser state (curToken, peekToken)
- 411 explicit `nextToken()` calls throughout codebase
- 193 error reporting sites with string-based errors
- Manual position tracking and backtracking
- Good test coverage (50 test files) and benchmarks

**Target State**:
- Immutable cursor-based token navigation
- Structured error types with rich context
- Parser combinator library for common patterns
- Automatic position tracking via builder pattern
- Separated parsing concerns from semantic analysis
- Reduced code duplication (target: 20-30% reduction)
- Zero performance regression (measured via benchmarks)

**Estimated Timeline**: 12-16 weeks (3-4 months)

**Strategy**: Strangler Fig Pattern - build new infrastructure alongside existing code, migrate incrementally, remove old code when safe.

---

## Phase 2.1: Foundation Layer (Weeks 1-2)

### Task 2.1.1: Structured Error Types (Week 1, Days 1-3)

**Goal**: Replace string-based errors with rich, structured error types.

**Why First**:
- No parser logic changes required
- Immediate value in better error messages
- Foundation for IDE/LSP integration
- Can be done incrementally alongside existing errors

**Current Problem**:
```go
// 193 call sites with string concatenation
p.addError("expected next token to be " + expected + ", got " + actual, ErrUnexpectedToken)
```

**New Structure**:
```go
type StructuredParserError struct {
    Kind     ErrorKind       // Enum: UnexpectedToken, MissingEnd, etc.
    Pos      token.Position  // Where error occurred
    Length   int             // How many characters
    Expected []token.TokenType  // What we expected (if applicable)
    Got      token.TokenType    // What we got (if applicable)
    Context  []BlockContext     // Parsing context stack
    Message  string            // Human-readable message

    // For backward compatibility during migration
    legacy *ParserError
}

type ErrorKind int

const (
    ErrUnexpectedToken ErrorKind = iota
    ErrMissingEnd
    ErrMissingSemicolon
    ErrExpectedIdent
    ErrExpectedType
    // ... 20+ error kinds
)
```

**Implementation Steps**:
1. Create `internal/parser/structured_error.go` with new types
2. Implement `Error()` method with rich formatting
3. Add `addStructuredError()` method alongside existing `addError()`
4. Migrate 5 high-value parsing functions as proof of concept
5. Add tests for error formatting and context

**Files Created**:
- `internal/parser/structured_error.go` (~200 lines)
- `internal/parser/structured_error_test.go` (~300 lines)

**Files Modified**:
- `internal/parser/parser.go` (~50 lines - new method)
- `internal/parser/expressions.go` (~30 lines - 5 functions migrated)

**Success Metrics**:
- New error type compiles and tests pass
- 5 parsing functions use structured errors
- Error messages include context automatically
- Existing error handling still works (compatibility)

**Deliverable**: Working structured error system with proof-of-concept usage

---

### Task 2.1.2: Parse Context Extraction (Week 1, Days 4-5)

**Goal**: Extract scattered context flags into a structured `ParseContext` type.

**Current Problem**:
```go
type Parser struct {
    // Scattered state throughout struct
    enableSemanticAnalysis bool
    parsingPostCondition   bool
    blockStack             []BlockContext
    // ... mixed concerns
}
```

**New Structure**:
```go
type ParseContext struct {
    blocks []BlockContext  // Block nesting stack
    flags  ContextFlags    // Structured flags
}

type ContextFlags struct {
    inPostCondition bool
    inClassBody     bool
    inLoopBody      bool
    inFunctionBody  bool
    allowImplicitThis bool
}

// Stack operations
func (c *ParseContext) PushBlock(blockType string, pos token.Position)
func (c *ParseContext) PopBlock()
func (c *ParseContext) CurrentBlock() *BlockContext
func (c *ParseContext) SetFlag(flag ContextFlag, value bool)
func (c *ParseContext) InContext(flag ContextFlag) bool
```

**Implementation Steps**:
1. Create `ParseContext` type
2. Add `p.context` field to Parser
3. Create adapter methods for backward compatibility
4. Migrate one parsing function to use new context
5. Gradually replace direct flag access

**Files Created**:
- `internal/parser/context.go` (~150 lines)

**Files Modified**:
- `internal/parser/parser.go` (~30 lines - add context field)
- `internal/parser/expressions.go` (~20 lines - migrate parseOldExpression)

**Success Metrics**:
- Context stack works correctly
- One function successfully uses new context
- All tests pass
- No performance regression

**Deliverable**: Working `ParseContext` with initial usage

---

### Task 2.1.3: Error Context Integration (Week 2, Days 1-2)

**Goal**: Integrate `ParseContext` with structured errors for automatic context inclusion.

**Implementation**:
```go
func (p *Parser) addStructuredError(kind ErrorKind, expected ...token.TokenType) {
    err := &StructuredParserError{
        Kind:     kind,
        Pos:      p.curToken.Pos,
        Expected: expected,
        Got:      p.curToken.Type,
        Context:  p.context.Snapshot(), // ✅ Automatic context capture
    }
    p.errors = append(p.errors, err)
}
```

**Implementation Steps**:
1. Add `Snapshot()` method to `ParseContext`
2. Update `addStructuredError()` to capture context
3. Enhance error formatting to show context
4. Test with nested blocks

**Files Modified**:
- `internal/parser/context.go` (~20 lines)
- `internal/parser/structured_error.go` (~30 lines)
- `internal/parser/structured_error_test.go` (~100 lines - context tests)

**Success Metrics**:
- Errors automatically include block context
- Error messages show "in begin block starting at line 10"
- Context survives state save/restore

**Deliverable**: Rich error messages with automatic context

---

### Task 2.1.4: Benchmark Infrastructure (Week 2, Days 3-5)

**Goal**: Establish baseline benchmarks and regression detection before major changes.

**Implementation**:
1. Audit existing benchmarks in `parser_bench_test.go`
2. Add missing benchmarks for common patterns:
   - List parsing
   - Expression parsing
   - Type parsing
   - Error recovery
3. Create benchmark comparison script
4. Document baseline performance

**Files Created**:
- `scripts/bench_compare.sh` (~50 lines)
- `docs/parser-benchmarks.md` (~150 lines)

**Files Modified**:
- `internal/parser/parser_bench_test.go` (~200 lines added)

**Success Metrics**:
- Comprehensive benchmark coverage (>90% of parsing patterns)
- Baseline measurements documented
- Automated comparison script works
- CI can detect >10% regressions

**Deliverable**: Complete benchmark suite with baseline measurements

---

## Phase 2.2: Token Cursor Abstraction (Weeks 3-5)

### Task 2.2.1: Design and Implement TokenCursor (Week 3, Days 1-3)

**Goal**: Create immutable cursor abstraction to replace mutable parser state.

**Why**: Explicit token navigation, easier backtracking, clearer contracts.

**Design**:
```go
// TokenCursor provides immutable token navigation
type TokenCursor struct {
    tokens []token.Token
    pos    int
}

// Core navigation
func (c *TokenCursor) Current() token.Token
func (c *TokenCursor) Peek(n int) token.Token    // n tokens ahead
func (c *TokenCursor) Advance() *TokenCursor     // Returns new cursor
func (c *TokenCursor) Skip(n int) *TokenCursor   // Advance n tokens

// Convenience methods
func (c *TokenCursor) Is(t token.TokenType) bool
func (c *TokenCursor) IsAny(types ...token.TokenType) bool
func (c *TokenCursor) Expect(t token.TokenType) (*TokenCursor, error)

// Backtracking support
func (c *TokenCursor) Mark() CursorMark
func (c *TokenCursor) ResetTo(mark CursorMark) *TokenCursor
func (c *TokenCursor) Clone() *TokenCursor

// Position info
func (c *TokenCursor) Position() token.Position
func (c *TokenCursor) AtEOF() bool
```

**Implementation Steps**:
1. Create `internal/parser/cursor.go`
2. Implement all core methods
3. Add comprehensive unit tests
4. Benchmark cursor operations vs current approach

**Files Created**:
- `internal/parser/cursor.go` (~250 lines)
- `internal/parser/cursor_test.go` (~400 lines)

**Success Metrics**:
- All cursor methods tested
- Immutability verified
- Performance comparable to current approach (<5% slower OK)

**Deliverable**: Fully tested `TokenCursor` ready for integration

---

### Task 2.2.2: Dual-Mode Parser Setup (Week 3, Days 4-5)

**Goal**: Enable parser to work in both old mode (mutable) and new mode (cursor).

**Implementation**:
```go
type Parser struct {
    // OLD: Mutable state (keep for now)
    curToken  token.Token
    peekToken token.Token

    // NEW: Cursor-based (optional)
    cursor *TokenCursor

    // Control which mode
    useCursor bool
}

func (p *Parser) ParseProgram() *ast.Program {
    if p.useCursor {
        return p.parseProgramCursor()
    }
    return p.parseProgramLegacy()
}
```

**Implementation Steps**:
1. Add `cursor` and `useCursor` fields to Parser
2. Create factory method `NewCursorParser(lexer *Lexer)`
3. Add feature flag in tests
4. Set up differential testing

**Files Modified**:
- `internal/parser/parser.go` (~40 lines)
- `internal/parser/parser_test.go` (~60 lines - dual mode tests)

**Success Metrics**:
- Parser can run in either mode
- Tests pass in both modes
- Easy to switch between modes

**Deliverable**: Parser infrastructure ready for cursor migration

---

### Task 2.2.3: Migrate First Parsing Function (Week 4, Days 1-2)

**Goal**: Prove the cursor approach works by migrating one complete parsing function.

**Target**: `parseIntegerLiteral` (simplest, lowest risk)

**Old Implementation**:
```go
func (p *Parser) parseIntegerLiteral() ast.Expression {
    // Uses p.curToken directly
    lit := &ast.IntegerLiteral{
        Token: p.curToken,
        // ...
    }
    return lit
}
```

**New Implementation**:
```go
func (p *Parser) parseIntegerLiteralCursor(cursor *TokenCursor) (ast.Expression, *TokenCursor, error) {
    lit := &ast.IntegerLiteral{
        Token: cursor.Current(),
        // ...
    }
    return lit, cursor.Advance(), nil
}
```

**Implementation Steps**:
1. Write cursor-based version alongside old version
2. Add tests comparing both implementations
3. Verify AST output is identical
4. Benchmark both versions

**Files Modified**:
- `internal/parser/expressions.go` (~30 lines added)
- `internal/parser/expressions_test.go` (~50 lines)

**Success Metrics**:
- Cursor version produces identical AST
- All tests pass
- Performance within 5% of old version

**Deliverable**: Proof-of-concept cursor-based parsing

---

### Task 2.2.4: Migrate Expression Parsing (Week 4, Days 3-5)

**Goal**: Migrate core expression parsing to cursor-based approach.

**Targets**:
- `parseIdentifier`
- `parseStringLiteral`
- `parseBooleanLiteral`
- `parseFloatLiteral`
- `parseGroupedExpression`

**Strategy**: Create parallel implementations, test thoroughly, measure performance.

**Files Modified**:
- `internal/parser/expressions.go` (~150 lines added)
- `internal/parser/expressions_test.go` (~200 lines added)

**Success Metrics**:
- 5+ expression types migrated
- Differential tests pass
- Performance acceptable

**Deliverable**: Core literals working with cursor

---

### Task 2.2.5: Migrate Infix Expressions (Week 5)

**Goal**: Migrate binary/infix expression parsing to cursor.

**Targets**:
- `parseBinaryExpression`
- `parseCallExpression`
- `parseMemberAccess`
- `parseIndexExpression`

**Challenge**: These are more complex due to precedence handling.

**Files Modified**:
- `internal/parser/expressions.go` (~200 lines)
- `internal/parser/operators.go` (~100 lines)

**Success Metrics**:
- Complex expressions parse correctly
- Precedence handling works
- Performance maintained

**Deliverable**: Full expression parsing with cursor

---

## Phase 2.3: Parser Combinators (Weeks 6-7)

### Task 2.3.1: Design Combinator Library (Week 6, Days 1-2)

**Goal**: Create reusable parser combinators for common patterns.

**Core Combinators**:
```go
type ParseFn func(*TokenCursor) (ast.Node, *TokenCursor, error)

// Optional: parse or return nil
func Optional(fn ParseFn) ParseFn

// Many: parse 0+ times
func Many(fn ParseFn) ParseFn

// Many1: parse 1+ times (at least one)
func Many1(fn ParseFn) ParseFn

// SeparatedList: parse items separated by delimiter
func SeparatedList(sep token.TokenType, terminator token.TokenType, item ParseFn) ParseFn

// Between: parse between delimiters
func Between(open, close token.TokenType, content ParseFn) ParseFn

// Choice: try alternatives in order
func Choice(alternatives ...ParseFn) ParseFn

// Sequence: parse in sequence
func Sequence(parsers ...ParseFn) ParseFn
```

**Implementation Steps**:
1. Create `internal/parser/combinators.go`
2. Implement core combinators
3. Add unit tests for each
4. Document usage patterns

**Files Created**:
- `internal/parser/combinators.go` (~300 lines)
- `internal/parser/combinators_test.go` (~500 lines)
- `docs/parser-combinators.md` (~200 lines)

**Success Metrics**:
- All combinators tested
- Clear documentation
- Examples provided

**Deliverable**: Working combinator library

---

### Task 2.3.2: Refactor List Parsing with Combinators (Week 6, Days 3-5)

**Goal**: Replace manual list parsing with combinator-based approach.

**Old Pattern** (411 `nextToken()` calls, many in list loops):
```go
for p.peekTokenIs(COMMA) {
    p.nextToken() // consume comma
    // ... parse item ...
    // ... handle errors ...
    // ... handle trailing separator ...
}
```

**New Pattern**:
```go
items, cursor, err := SeparatedList(
    COMMA,
    RPAREN,
    parseItem,
)(cursor)
```

**Targets**:
- `parseParameterList`
- `parseArgumentList`
- `parseFieldList`
- `parseEnumValues`

**Files Modified**:
- `internal/parser/functions.go` (~80 lines)
- `internal/parser/expressions.go` (~60 lines)
- `internal/parser/records.go` (~40 lines)
- `internal/parser/enums.go` (~30 lines)

**Success Metrics**:
- 4+ list parsing functions refactored
- Code reduction of 30%+ in these functions
- All tests pass
- Performance maintained

**Deliverable**: List parsing using combinators

---

### Task 2.3.3: Create High-Level Combinators (Week 7)

**Goal**: Build domain-specific combinators for DWScript patterns.

**Examples**:
```go
// Parse optional type annotation: : TypeName
func OptionalTypeAnnotation() ParseFn

// Parse identifier list: x, y, z
func IdentifierList() ParseFn

// Parse statement block: begin ... end
func StatementBlock() ParseFn

// Parse parameter group: x, y: Integer
func ParameterGroup() ParseFn
```

**Implementation Steps**:
1. Identify common DWScript patterns
2. Create combinator for each pattern
3. Refactor existing code to use combinators
4. Measure code reduction

**Files Modified**:
- `internal/parser/combinators.go` (~200 lines added)
- Multiple parser files (~300 lines total changed)

**Success Metrics**:
- 10+ high-level combinators
- Measurable code reduction (target: 20%+)
- More declarative parsing code

**Deliverable**: Rich combinator library for DWScript

---

## Phase 2.4: Automatic Position Tracking (Week 8)

### Task 2.4.1: NodeBuilder Pattern (Week 8, Days 1-3)

**Goal**: Eliminate manual `EndPos` setting throughout parser.

**Current Problem** (manual position tracking everywhere):
```go
stmt := &ast.IfStatement{...}
// ... parse ...
stmt.EndPos = p.endPosFromToken(p.curToken) // ❌ Easy to forget
```

**New Pattern**:
```go
type NodeBuilder struct {
    startCursor *TokenCursor
}

func StartNode(cursor *TokenCursor) *NodeBuilder {
    return &NodeBuilder{startCursor: cursor}
}

func (b *NodeBuilder) Finish(cursor *TokenCursor) ast.BaseNode {
    return ast.BaseNode{
        Token:  b.startCursor.Current(),
        EndPos: cursor.Position(),  // ✅ Automatic
    }
}

// Usage:
func parseIfStatement(cursor *TokenCursor) (*ast.IfStatement, *TokenCursor, error) {
    builder := StartNode(cursor)

    // ... parse ...

    return &ast.IfStatement{
        BaseNode: builder.Finish(cursor),  // ✅ Can't forget
        // ...
    }, cursor, nil
}
```

**Implementation Steps**:
1. Create `NodeBuilder` type
2. Add helper methods
3. Migrate 5 parsing functions as proof of concept
4. Verify positions are correct

**Files Created**:
- `internal/parser/node_builder.go` (~150 lines)
- `internal/parser/node_builder_test.go` (~200 lines)

**Files Modified**:
- `internal/parser/statements.go` (~50 lines - 5 functions migrated)

**Success Metrics**:
- Positions always correct
- Can't forget to set EndPos
- Tests verify position accuracy

**Deliverable**: Working NodeBuilder with proof of concept

---

### Task 2.4.2: Mass Migration to NodeBuilder (Week 8, Days 4-5)

**Goal**: Migrate all parsing functions to use NodeBuilder.

**Strategy**: Automated refactoring where possible, manual for complex cases.

**Files Modified**:
- All parser files (~500 lines total changes)

**Success Metrics**:
- All parsing functions use NodeBuilder
- Position tests pass
- No manual EndPos setting remains

**Deliverable**: Complete NodeBuilder adoption

---

## Phase 2.5: Separation of Concerns (Week 9-10)

### Task 2.5.1: Remove Semantic Analysis from Parser (Week 9)

**Goal**: Parser should only build AST, not perform type checking.

**Current Problem**:
```go
type Parser struct {
    enableSemanticAnalysis bool  // ❌ Wrong layer
    semanticErrors         []string  // ❌ Mixing concerns
}

func (p *Parser) parseVarDecl() {
    if p.enableSemanticAnalysis {
        // Type checking in parser ❌
    }
}
```

**Implementation Steps**:
1. Identify all semantic analysis code in parser
2. Move to separate `SemanticAnalyzer` type
3. Update tests to run analysis after parsing
4. Remove semantic fields from Parser

**Files Created**:
- `internal/semantic/analyzer_refactored.go` (~200 lines)

**Files Modified**:
- `internal/parser/parser.go` (~50 lines removed)
- `internal/parser/statements.go` (~80 lines removed)
- All test files (~100 lines updated to separate phases)

**Success Metrics**:
- Parser has zero semantic analysis code
- Semantic analyzer works independently
- Faster parsing (no type checking during parse)

**Deliverable**: Clean separation of parsing and semantic analysis

---

### Task 2.5.2: Extract Error Recovery to Separate Module (Week 10, Days 1-2)

**Goal**: Centralize error recovery logic.

**Implementation**:
```go
type ErrorRecovery struct {
    synchPoints map[token.TokenType]bool
    context     *ParseContext
}

func (e *ErrorRecovery) Synchronize(cursor *TokenCursor) *TokenCursor
func (e *ErrorRecovery) CanRecover() bool
func (e *ErrorRecovery) SuggestRecovery() string
```

**Files Created**:
- `internal/parser/error_recovery.go` (~200 lines)

**Files Modified**:
- Various parser files (~100 lines simplified)

**Success Metrics**:
- Centralized recovery logic
- More consistent recovery behavior

**Deliverable**: Reusable error recovery module

---

### Task 2.5.3: Parser Factory Pattern (Week 10, Days 3-5)

**Goal**: Clean up parser construction and configuration.

**Implementation**:
```go
type ParserConfig struct {
    EnableExperimental bool
    MaxErrors          int
    RecoveryStrategy   RecoveryStrategy
}

type ParserBuilder struct {
    config ParserConfig
}

func NewParser(lexer *Lexer) *ParserBuilder
func (b *ParserBuilder) WithConfig(cfg ParserConfig) *ParserBuilder
func (b *ParserBuilder) Build() *Parser
```

**Files Modified**:
- `internal/parser/parser.go` (~100 lines)

**Success Metrics**:
- Clean construction API
- Easy to configure
- All options in one place

**Deliverable**: ParserBuilder for clean construction

---

## Phase 2.6: Advanced Cursor Features (Week 11)

### Task 2.6.1: Lookahead Abstraction (Week 11, Days 1-2)

**Goal**: Make lookahead declarative instead of imperative.

**Current** (confusing offset arithmetic):
```go
tokenAfter := p.peek(0)  // What does 0 mean?
tokenTwoAhead := p.peek(1)
```

**New**:
```go
func (c *TokenCursor) LookAhead(distance int, predicate func(token.Token) bool) bool
func (c *TokenCursor) ScanUntil(predicate func(token.Token) bool) *TokenCursor
func (c *TokenCursor) FindNext(tokenType token.TokenType) (int, bool)
```

**Files Modified**:
- `internal/parser/cursor.go` (~80 lines added)
- Various disambiguation functions (~60 lines simplified)

**Deliverable**: Declarative lookahead utilities

---

### Task 2.6.2: Backtracking Optimization (Week 11, Days 3-5)

**Goal**: Optimize backtracking with minimal state saving.

**Implementation**:
```go
type LightweightMark struct {
    pos int  // Just the position
}

func (c *TokenCursor) QuickMark() LightweightMark
func (c *TokenCursor) QuickReset(mark LightweightMark) *TokenCursor

// Only for complex backtracking:
type HeavyweightMark struct {
    pos     int
    errors  []StructuredParserError
    context ParseContext
}
```

**Files Modified**:
- `internal/parser/cursor.go` (~50 lines)
- Backtracking sites (~30 lines optimized)

**Success Metrics**:
- Faster backtracking
- Measured performance improvement

**Deliverable**: Optimized backtracking

---

## Phase 2.7: Migration Completion (Weeks 12-14)

### Task 2.7.1: Statement Parsing Migration (Week 12)

**Goal**: Migrate all statement parsing to cursor/combinators.

**Targets**:
- Control flow (if, while, for, case, etc.)
- Declarations (var, const, type)
- Blocks (begin...end)

**Files Modified**:
- `internal/parser/statements.go` (major refactor)
- `internal/parser/control_flow.go` (major refactor)
- `internal/parser/declarations.go` (major refactor)

**Success Metrics**:
- All statements use cursor
- Code reduction measured
- All tests pass

**Deliverable**: Complete statement parsing migration

---

### Task 2.7.2: Type Parsing Migration (Week 13)

**Goal**: Migrate type parsing to cursor/combinators.

**Targets**:
- Simple types
- Function pointer types
- Array types
- Record types
- Class types

**Files Modified**:
- `internal/parser/types.go` (major refactor)
- `internal/parser/records.go` (~100 lines)
- `internal/parser/arrays.go` (~80 lines)

**Success Metrics**:
- All type parsing uses cursor
- Complex types work correctly
- Tests pass

**Deliverable**: Complete type parsing migration

---

### Task 2.7.3: Function/Class Parsing Migration (Week 14, Days 1-3)

**Goal**: Migrate complex declaration parsing.

**Targets**:
- Function declarations
- Class declarations
- Interface declarations
- Properties, methods

**Files Modified**:
- `internal/parser/functions.go` (major refactor)
- `internal/parser/classes.go` (major refactor)
- `internal/parser/interfaces.go` (major refactor)

**Success Metrics**:
- All declarations use cursor
- All tests pass
- Code reduction achieved

**Deliverable**: Complete declaration parsing migration

---

### Task 2.7.4: Remove Legacy Code (Week 14, Days 4-5)

**Goal**: Remove all old mutable parser code.

**Steps**:
1. Verify 100% migration complete
2. Remove `curToken`/`peekToken` fields
3. Remove `nextToken()` method
4. Remove old error methods
5. Clean up imports and dead code

**Files Modified**:
- `internal/parser/parser.go` (~200 lines removed)
- All parser files (~500 lines removed total)

**Success Metrics**:
- Zero legacy code remains
- All tests pass
- Codebase smaller and cleaner

**Deliverable**: Clean, modern parser

---

## Phase 2.8: Optimization & Polish (Weeks 15-16)

### Task 2.8.1: Performance Tuning (Week 15)

**Goal**: Ensure new parser is as fast as old parser.

**Activities**:
1. Run comprehensive benchmarks
2. Profile hot paths
3. Optimize cursor operations
4. Optimize combinator allocation
5. Add caching where beneficial

**Success Metrics**:
- Performance within 5% of baseline (target: same or better)
- No memory leaks
- Allocation rate acceptable

**Deliverable**: Performance-tuned parser

---

### Task 2.8.2: Documentation Updates (Week 16, Days 1-2)

**Goal**: Update all documentation for new architecture.

**Files to Update**:
- `docs/parser-architecture.md` (complete rewrite)
- `docs/parser-style-guide.md` (update for cursor patterns)
- `docs/parser-extension-guide.md` (new examples)
- `README.md` (update parser section)
- Inline code documentation

**Deliverable**: Comprehensive documentation

---

### Task 2.8.3: Migration Guide (Week 16, Days 3-4)

**Goal**: Document the migration for future reference.

**Contents**:
- What changed and why
- Performance comparison
- Code reduction metrics
- Lessons learned
- Future improvement opportunities

**Files Created**:
- `docs/parser-modernization-retrospective.md` (~500 lines)

**Deliverable**: Complete migration retrospective

---

### Task 2.8.4: Final Testing & Validation (Week 16, Day 5)

**Goal**: Comprehensive validation of new parser.

**Activities**:
1. Run full test suite (target: 100% pass)
2. Run all benchmarks (verify performance)
3. Test on real-world DWScript code
4. Verify error messages quality
5. Check code coverage (target: maintain >84%)

**Success Metrics**:
- All tests pass
- Performance acceptable
- Coverage maintained
- Error messages helpful

**Deliverable**: Production-ready modern parser

---

## Summary

### Effort Breakdown
- **Foundation** (Weeks 1-2): 80 hours
- **Token Cursor** (Weeks 3-5): 120 hours
- **Combinators** (Weeks 6-7): 80 hours
- **Position Tracking** (Week 8): 40 hours
- **Separation** (Weeks 9-10): 80 hours
- **Advanced Features** (Week 11): 40 hours
- **Migration** (Weeks 12-14): 120 hours
- **Polish** (Weeks 15-16): 80 hours

**Total**: ~640 hours (~16 weeks full-time, or 3-4 months)

### Key Benefits

**Maintainability** (+++):
- 20-30% code reduction
- Declarative combinators
- No manual token tracking
- Automatic position tracking
- Clear separation of concerns

**Reliability** (+++):
- Structured errors with context
- Better error recovery
- Immutable cursor prevents bugs
- Type-safe combinators

**Performance** (=):
- Target: within 5% of baseline
- Better in some cases (less backtracking)
- Cursor overhead minimal

**Developer Experience** (+++):
- Easier to add new features
- Clearer code intent
- Better error messages
- Comprehensive documentation

### Risk Mitigation

1. **Dual-mode operation**: Old and new parsers coexist
2. **Differential testing**: Both parsers must agree
3. **Incremental migration**: Each task delivers value
4. **Performance tracking**: Benchmarks at each step
5. **Feature flags**: Easy rollback if needed
6. **Comprehensive tests**: 49 test files ensure correctness

### Success Criteria

- ✅ All 49 test files pass
- ✅ Performance within 5% of baseline
- ✅ 20%+ code reduction achieved
- ✅ Zero manual `nextToken()` calls
- ✅ Zero manual `EndPos` assignments
- ✅ Structured errors throughout
- ✅ Complete documentation
- ✅ Production-ready quality

### Non-Goals

- ❌ Changing AST structure
- ❌ Changing parser semantics
- ❌ Adding new language features
- ❌ Rewriting lexer
- ❌ Performance degradation

---

## Getting Started

**Week 1, Day 1**: Start with Task 2.1.1 (Structured Errors)

```bash
# Create branch
git checkout -b parser-modernization-2.1.1

# Create new file
touch internal/parser/structured_error.go

# Implement StructuredParserError
# Run tests
# Commit

# Target: Working structured error type by end of day
```

This is a manageable, incremental path from traditional parser to modern architecture. Each task delivers value, all tests pass throughout, and we can stop at any point with a working parser.
