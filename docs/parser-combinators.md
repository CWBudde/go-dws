# Parser Combinators

This document describes the parser combinator library implemented in `internal/parser/combinators.go`. Parser combinators are higher-order functions that encapsulate common parsing patterns and can be composed together to build complex parsers.

## Table of Contents

- [Overview](#overview)
- [Design Philosophy](#design-philosophy)
- [Combinator Reference](#combinator-reference)
  - [Optional Combinators](#optional-combinators)
  - [Repetition Combinators](#repetition-combinators)
  - [Choice and Sequence](#choice-and-sequence)
  - [Delimiter Combinators](#delimiter-combinators)
  - [List Parsing](#list-parsing)
  - [Lookahead and Guards](#lookahead-and-guards)
  - [Error Recovery](#error-recovery)
  - [Backtracking](#backtracking)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)

## Overview

Parser combinators provide a declarative way to express parsing logic. Instead of writing manual token manipulation code, you compose higher-level functions that handle common patterns like optional tokens, repeated elements, or delimited lists.

### Benefits

- **Reusability**: Common patterns are extracted into reusable functions
- **Readability**: Code becomes more declarative and easier to understand
- **Composability**: Combinators can be nested and combined
- **Type Safety**: Uses Go's type system to catch errors at compile time
- **Zero Overhead**: Direct function calls with no reflection or runtime penalties
- **Testability**: Each combinator is independently testable

## Design Philosophy

The combinator library follows these principles:

1. **Non-invasive**: Works with existing parser methods without modification
2. **Functional**: Pure functions with no side effects (except token consumption)
3. **Explicit**: Clear naming that describes what each combinator does
4. **Predictable**: Consistent behavior across all combinators
5. **Efficient**: No reflection, no allocations beyond necessary parsing

## Combinator Reference

### Optional Combinators

#### `Optional(tokenType TokenType) bool`

Attempts to consume a token of the given type. Returns `true` if matched and consumed, `false` otherwise.

**Use Cases:**
- Optional semicolons
- Optional visibility modifiers
- Optional forward declarations

**Example:**
```go
// Parse optional semicolon after statement
hasTerminator := p.Optional(lexer.SEMICOLON)

// Parse optional visibility specifier
isPublic := p.Optional(lexer.PUBLIC)
```

**Behavior:**
- Checks if `peekToken` matches the given type
- If match: advances parser and returns `true`
- If no match: parser state unchanged, returns `false`

#### `OptionalOneOf(tokenTypes ...TokenType) TokenType`

Attempts to consume one of several token types. Returns the matched type, or `lexer.ILLEGAL` if none match.

**Use Cases:**
- Optional visibility specifiers (public, private, protected)
- Optional parameter modifiers (var, const, lazy)
- Optional operator variations

**Example:**
```go
// Parse optional visibility specifier
visibility := p.OptionalOneOf(lexer.PUBLIC, lexer.PRIVATE, lexer.PROTECTED)
if visibility != lexer.ILLEGAL {
    // Handle visibility
}

// Parse optional unary operator
operator := p.OptionalOneOf(lexer.PLUS, lexer.MINUS, lexer.NOT)
```

### Repetition Combinators

#### `Many(parseFn ParserFunc) int`

Applies a parse function zero or more times. Returns the count of successful applications.

**Use Cases:**
- Parsing statement blocks
- Parsing multiple declarations
- Collecting items until a condition fails

**Example:**
```go
// Parse zero or more statements
var statements []ast.Statement
count := p.Many(func() bool {
    stmt := p.parseStatement()
    if stmt != nil {
        statements = append(statements, stmt)
        return true
    }
    return false
})
```

**Behavior:**
- Calls `parseFn` repeatedly until it returns `false`
- Never fails (returns 0 if no matches)
- Parser position advances with each successful call

#### `Many1(parseFn ParserFunc) int`

Applies a parse function one or more times. Returns the count of successful applications, or 0 if it fails on first attempt.

**Use Cases:**
- Parsing required lists with at least one element
- Parsing digit sequences
- Parsing identifier chains

**Example:**
```go
// Parse one or more digits (at least one required)
count := p.Many1(func() bool {
    if p.peekTokenIs(lexer.INT) {
        p.nextToken()
        return true
    }
    return false
})
if count == 0 {
    p.addError("expected at least one digit")
    return nil
}
```

**Behavior:**
- Calls `parseFn` at least once
- Returns 0 if first call fails
- Continues calling while `parseFn` returns `true`

#### `ManyUntil(terminator TokenType, parseFn ParserFunc) int`

Applies a parse function repeatedly until a terminator token is found. Returns the count of successful applications.

**Use Cases:**
- Parsing statements until 'end'
- Parsing case branches until 'else' or 'end'
- Parsing items until a specific delimiter

**Example:**
```go
// Parse statements until 'end' keyword
var statements []ast.Statement
count := p.ManyUntil(lexer.END, func() bool {
    stmt := p.parseStatement()
    if stmt != nil {
        statements = append(statements, stmt)
        return true
    }
    return false
})
```

**Behavior:**
- Stops when `peekToken` is the terminator or EOF
- Does not consume the terminator
- Returns count of successful parse function calls

### Choice and Sequence

#### `Choice(tokenTypes ...TokenType) bool`

Attempts to consume one of several token types. Returns `true` if any matches.

**Use Cases:**
- Matching alternative operators
- Matching alternative keywords
- Branching based on next token

**Example:**
```go
// Match either '+' or '-' for unary operators
if p.Choice(lexer.PLUS, lexer.MINUS) {
    operator := p.curToken.Literal
    // ... parse unary expression
}

// Match logical operators
if p.Choice(lexer.AND, lexer.OR, lexer.XOR) {
    // ... handle logical operation
}
```

**Behavior:**
- Checks each token type in order
- Consumes and returns `true` on first match
- Returns `false` if none match

#### `Sequence(tokenTypes ...TokenType) bool`

Attempts to match a sequence of token types in order. All must match for success.

**Use Cases:**
- Lookahead checks for complex patterns
- Validating multi-token sequences
- Disambiguation

**Example:**
```go
// Check for assignment operator :=
if p.Sequence(lexer.COLON, lexer.ASSIGN) {
    // Both tokens matched and consumed
}
```

**Behavior:**
- Checks all tokens without consuming first
- If all match: consumes all tokens and returns `true`
- If any doesn't match: no tokens consumed, returns `false`
- Uses lookahead to validate before committing

### Delimiter Combinators

#### `Between(opening, closing TokenType, parseFn ExpressionParserFunc) Expression`

Parses content surrounded by delimiters (e.g., parentheses, brackets).

**Use Cases:**
- Parenthesized expressions
- Array indexing
- Generic type arguments

**Example:**
```go
// Parse parenthesized expression: (expr)
expr := p.Between(lexer.LPAREN, lexer.RPAREN, func() ast.Expression {
    return p.parseExpression(LOWEST)
})

// Parse array indexing: [index]
index := p.Between(lexer.LBRACK, lexer.RBRACK, func() ast.Expression {
    return p.parseExpression(LOWEST)
})
```

**Behavior:**
- Expects opening delimiter as `peekToken`
- Parses content with provided function
- Expects closing delimiter after content
- Returns `nil` if any part fails

#### `BetweenStatement(opening, closing TokenType, parseFn StatementParserFunc) Statement`

Like `Between` but for statements.

**Example:**
```go
// Parse begin...end block
block := p.BetweenStatement(lexer.BEGIN, lexer.END, func() ast.Statement {
    return p.parseBlockStatement()
})
```

### List Parsing

#### `SeparatedList(config SeparatorConfig) int`

Parses a list of items separated by a delimiter. Highly configurable for different list patterns.

**Configuration:**
- `Sep`: Separator token (e.g., `COMMA`)
- `Term`: Terminator token (e.g., `RPAREN`)
- `ParseItem`: Function to parse each item
- `AllowEmpty`: Permit empty lists
- `AllowTrailing`: Permit trailing separator
- `RequireTerm`: Require terminator at end

**Use Cases:**
- Function parameters
- Array literals
- Record field lists

**Example:**
```go
// Parse parameter list: (a, b, c)
var params []*ast.Parameter
count := p.SeparatedList(SeparatorConfig{
    Sep: lexer.COMMA,
    Term: lexer.RPAREN,
    ParseItem: func() bool {
        param := p.parseParameter()
        if param != nil {
            params = append(params, param)
            return true
        }
        return false
    },
    AllowEmpty: true,
    RequireTerm: true,
})
```

**Behavior:**
- Expects `curToken` to be on first item or terminator
- Calls `ParseItem` for each element
- Handles separators and optional trailing separator
- Returns count of items parsed, or -1 on failure

#### `SeparatedListMultiSep(...) int`

Like `SeparatedList` but allows multiple separator types (e.g., comma or semicolon).

**Example:**
```go
// Parse record fields with either comma or semicolon
count := p.SeparatedListMultiSep(
    []lexer.TokenType{lexer.COMMA, lexer.SEMICOLON},
    lexer.END,
    func() bool { return p.parseFieldDecl() != nil },
    true,  // allow empty
    false, // no trailing separator
    true,  // require terminator
)
```

### Lookahead and Guards

#### `Guard(guardFn func() bool, parseFn ParserFunc) bool`

Applies a lookahead check before attempting to parse. Only parses if guard succeeds.

**Use Cases:**
- Conditional parsing based on context
- Avoiding unnecessary parse attempts
- Disambiguation

**Example:**
```go
// Only parse var declaration if we're at a 'var' keyword
success := p.Guard(
    func() bool { return p.curTokenIs(lexer.VAR) },
    func() bool { return p.parseVarDecl() != nil },
)
```

#### `PeekNIs(n int, tokenType TokenType) bool`

Checks if the token N positions ahead matches the given type.

**Helper Functions:**
- `Peek1Is(tokenType)`: Check next token (same as `peekTokenIs`)
- `Peek2Is(tokenType)`: Check 2 positions ahead
- `Peek3Is(tokenType)`: Check 3 positions ahead

**Example:**
```go
// Check if next token is identifier
if p.Peek1Is(lexer.IDENT) {
    // ...
}

// Check if token 2 positions ahead is colon
if p.Peek2Is(lexer.COLON) {
    // Likely a variable declaration
}
```

### Error Recovery

#### `SkipUntil(tokenTypes ...TokenType) bool`

Advances parser until one of the given tokens is found. Does not consume the found token.

**Use Cases:**
- Error recovery
- Synchronization after errors
- Finding block boundaries

**Example:**
```go
// Skip to next semicolon or end
if !p.SkipUntil(lexer.SEMICOLON, lexer.END, lexer.EOF) {
    // Reached EOF without finding target
}
```

#### `SkipPast(tokenTypes ...TokenType) bool`

Like `SkipUntil` but also consumes the found token.

**Example:**
```go
// Skip past the next semicolon
if p.SkipPast(lexer.SEMICOLON) {
    // Now positioned after the semicolon
}
```

### Backtracking

#### `TryParse(parseFn ExpressionParserFunc) Expression`

Attempts to parse and suppresses errors on failure. Useful for optional constructs.

**Important:** Does NOT rollback token position. Only use when parse function doesn't consume tokens on failure.

**Example:**
```go
// Try to parse optional type annotation
typeAnnotation := p.TryParse(func() ast.Expression {
    if p.curTokenIs(lexer.COLON) {
        if p.expectPeek(lexer.IDENT) {
            return p.parseTypeExpression()
        }
    }
    return nil
})
```

#### `TryParseStatement(parseFn StatementParserFunc) Statement`

Like `TryParse` but for statements.

## Usage Examples

### Example 1: Parsing Optional Semicolons

```go
func (p *Parser) parseStatement() ast.Statement {
    stmt := p.parseExpressionStatement()

    // Optional semicolon at end
    p.Optional(lexer.SEMICOLON)

    return stmt
}
```

### Example 2: Parsing Variable Declarations with Optional Visibility

```go
func (p *Parser) parseVarDecl() *ast.VarDeclaration {
    // Check for optional visibility modifier
    visibility := p.OptionalOneOf(lexer.PUBLIC, lexer.PRIVATE, lexer.PROTECTED)

    if !p.expectPeek(lexer.VAR) {
        return nil
    }

    // ... rest of parsing
}
```

### Example 3: Parsing Statement Blocks

```go
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
    block := &ast.BlockStatement{Token: p.curToken}

    if !p.expectPeek(lexer.BEGIN) {
        return nil
    }

    // Parse statements until 'end'
    p.ManyUntil(lexer.END, func() bool {
        stmt := p.parseStatement()
        if stmt != nil {
            block.Statements = append(block.Statements, stmt)
            return true
        }
        return false
    })

    if !p.expectPeek(lexer.END) {
        return nil
    }

    return block
}
```

### Example 4: Parsing Function Parameters

```go
func (p *Parser) parseFunctionParams() []*ast.Parameter {
    if !p.expectPeek(lexer.LPAREN) {
        return nil
    }

    p.nextToken() // Move to first param or closing paren

    var params []*ast.Parameter
    p.SeparatedList(SeparatorConfig{
        Sep: lexer.COMMA,
        Term: lexer.RPAREN,
        ParseItem: func() bool {
            param := p.parseParameter()
            if param != nil {
                params = append(params, param)
                return true
            }
            return false
        },
        AllowEmpty: true,
        RequireTerm: true,
    })

    return params
}
```

### Example 5: Parsing Case Statements

```go
func (p *Parser) parseCaseStatement() *ast.CaseStatement {
    caseStmt := &ast.CaseStatement{Token: p.curToken}

    if !p.expectPeek(lexer.CASE) {
        return nil
    }

    // Parse the expression being matched
    caseStmt.Expression = p.parseExpression(LOWEST)

    if !p.expectPeek(lexer.OF) {
        return nil
    }

    // Parse case branches until 'end' or 'else'
    p.ManyUntil(lexer.END, func() bool {
        if p.peekTokenIs(lexer.ELSE) {
            return false
        }
        branch := p.parseCaseBranch()
        if branch != nil {
            caseStmt.Branches = append(caseStmt.Branches, branch)
            return true
        }
        return false
    })

    // Optional else branch
    if p.Optional(lexer.ELSE) {
        caseStmt.ElseBranch = p.parseBlockStatement()
    }

    if !p.expectPeek(lexer.END) {
        return nil
    }

    return caseStmt
}
```

### Example 6: Error Recovery

```go
func (p *Parser) parseStatementWithRecovery() ast.Statement {
    stmt := p.parseStatement()
    if stmt == nil {
        // Recovery: skip to next semicolon or end
        p.addError("failed to parse statement")
        p.SkipPast(lexer.SEMICOLON, lexer.END)
        return nil
    }
    return stmt
}
```

## Best Practices

### 1. Use Combinators for Common Patterns

Instead of:
```go
// Manual token checking
if p.peekTokenIs(lexer.SEMICOLON) {
    p.nextToken()
}
```

Use:
```go
p.Optional(lexer.SEMICOLON)
```

### 2. Prefer Descriptive Combinator Names

The combinator names describe what they do, making code self-documenting:
- `Optional` - clearly indicates optionality
- `Many` - clearly indicates repetition
- `Between` - clearly indicates delimited content

### 3. Combine Combinators for Complex Patterns

```go
// Parse optional public var declaration
if p.Optional(lexer.PUBLIC) {
    if p.Guard(
        func() bool { return p.peekTokenIs(lexer.VAR) },
        func() bool { return p.parseVarDecl() != nil },
    ) {
        // Successfully parsed public var
    }
}
```

### 4. Use SeparatedList for All List Parsing

Instead of manually handling separators and terminators, use `SeparatedList` with appropriate configuration. It handles:
- Empty lists
- Single items
- Multiple items
- Trailing separators
- Error recovery

### 5. Guard Against Invalid States

Use `Guard` to check preconditions before expensive parsing:

```go
// Only parse if we're at the right token
if p.Guard(
    func() bool { return p.curTokenIs(lexer.FUNCTION) },
    func() bool { return p.parseFunctionDecl() != nil },
) {
    // Function parsed successfully
}
```

### 6. Use TryParse Sparingly

`TryParse` suppresses errors but doesn't rollback token position. Only use it for:
- Truly optional constructs where failure is expected
- Lookahead without backtracking
- Cases where the parse function is safe on failure

### 7. Keep Parse Functions Pure

Parse functions passed to combinators should:
- Return boolean success/failure or AST nodes
- Not have side effects beyond collecting results
- Be idempotent when possible

### 8. Document Complex Combinator Usage

When nesting combinators or using complex configurations, add comments:

```go
// Parse parameter list with these rules:
// - Empty lists allowed: ()
// - Comma separated: (a, b, c)
// - No trailing comma
// - Closing paren required
count := p.SeparatedList(SeparatorConfig{...})
```

## Performance Considerations

1. **Zero Overhead**: All combinators are simple function calls with no reflection or runtime penalties

2. **No Allocations**: Combinators don't allocate beyond what the underlying parse functions require

3. **Short-Circuit Evaluation**: Choice and Guard combinators stop at first success

4. **Lookahead Caching**: The lexer caches peeked tokens, so multiple lookahead calls are efficient

## Testing Combinators

Each combinator has comprehensive unit tests in `combinators_test.go`. When adding new combinators:

1. Test success cases
2. Test failure cases
3. Test edge cases (empty input, EOF, etc.)
4. Test composition with other combinators
5. Add benchmark tests for performance-critical combinators

## Future Extensions

Potential future additions to the combinator library:

1. **`OneOf` combinator**: Try multiple parse functions in order
2. **`Backtrack` combinator**: Full token position rollback
3. **`Memoize` combinator**: Cache parse results for packrat parsing
4. **`Recursive` combinator**: Handle left-recursive grammars
5. **`LeftAssoc/RightAssoc` combinators**: Handle operator associativity

## Conclusion

Parser combinators make parsing code more declarative, reusable, and testable. They encapsulate common patterns and allow composition of complex parsing logic from simple building blocks. Use them wherever they make the code clearer and more maintainable.
