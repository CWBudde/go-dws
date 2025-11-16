# Parser Conventions and Style Guide

This document establishes the conventions used in the go-dws parser to ensure consistency, maintainability, and readability.

## Table of Contents

1. [Token Consumption Convention](#token-consumption-convention)
2. [Pre/Post-Conditions](#prepost-conditions)
3. [Error Handling](#error-handling)
4. [Parsing Function Patterns](#parsing-function-patterns)
5. [Naming Conventions](#naming-conventions)

## Token Consumption Convention

### The Fundamental Rule

**All parsing functions are called WITH `curToken` positioned at the triggering token.**

This means:
- When a parsing function is invoked, `p.curToken` should already be at the token that triggers that parse function
- The function is responsible for consuming its own tokens as it parses
- The function should leave `curToken` at the last token it consumed (typically the ending delimiter or final token of the construct)

### Example: Block Statement

```go
// parseBlockStatement parses a begin...end block.
// PRE: curToken is BEGIN
// POST: curToken is END
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
    block := &ast.BlockStatement{
        BaseNode: ast.BaseNode{Token: p.curToken}, // Token is BEGIN
    }
    block.Statements = []ast.Statement{}

    p.nextToken() // advance past 'begin'

    // Parse statements until END
    for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
        stmt := p.parseStatement()
        if stmt != nil {
            block.Statements = append(block.Statements, stmt)
        }
        p.nextToken()
    }

    // curToken is now END
    return block
}
```

### Example: If Statement

```go
// parseIfStatement parses an if-then-else statement.
// PRE: curToken is IF
// POST: curToken is last token of the statement (END of block, SEMICOLON, or last statement token)
func (p *Parser) parseIfStatement() *ast.IfStatement {
    stmt := &ast.IfStatement{
        BaseNode: ast.BaseNode{Token: p.curToken}, // Token is IF
    }

    p.nextToken() // advance past 'if' to condition

    // Parse condition
    stmt.Condition = p.parseExpression(LOWEST)

    // Expect 'then'
    if !p.expectPeek(lexer.THEN) {
        return nil
    }

    p.nextToken() // advance past 'then'

    // Parse consequence
    stmt.Consequence = p.parseStatement()

    // ... rest of parsing

    return stmt
}
```

### Example: New Expression

```go
// parseNewExpression parses object instantiation: new ClassName(args)
// PRE: curToken is NEW
// POST: curToken is RPAREN (end of arguments) or last token of type name
func (p *Parser) parseNewExpression() ast.Expression {
    newToken := p.curToken // Save the 'new' token position

    // Expect a type name (identifier)
    if !p.expectPeek(lexer.IDENT) {
        return nil
    }

    typeName := &ast.Identifier{
        Token: p.curToken,
        Value: p.curToken.Literal,
    }

    // Parse arguments if present
    if p.peekTokenIs(lexer.LPAREN) {
        p.nextToken() // move to LPAREN
        // ... parse arguments
    }

    return newExpr
}
```

## Pre/Post-Conditions

Every parsing function must document its token position expectations using standardized comments:

### Format

```go
// parseFunctionName parses a <description of what it parses>.
// PRE: curToken is <EXPECTED_TOKEN_TYPE>
// POST: curToken is <FINAL_TOKEN_TYPE>
func (p *Parser) parseFunctionName() *ast.NodeType {
    // Implementation
}
```

### Guidelines

1. **PRE condition**: Documents what token type `curToken` must be when the function is called
2. **POST condition**: Documents what token type `curToken` will be when the function returns (on success)
3. Use ALL_CAPS for token type names (e.g., BEGIN, END, LPAREN, RPAREN)
4. Be specific about the exact token, not just a category
5. If multiple tokens are possible, list them: `curToken is SEMICOLON or last token of expression`

### Examples

```go
// parseClassDeclaration parses a class declaration.
// PRE: curToken is CLASS
// POST: curToken is END (class body end)

// parseParameterList parses a function parameter list.
// PRE: curToken is LPAREN
// POST: curToken is RPAREN

// parseExpression parses an expression with given precedence.
// PRE: curToken is first token of expression
// POST: curToken is last token of expression

// parseArrayBound parses a single array dimension bound.
// PRE: curToken is first token of bound expression or DOTDOT
// POST: curToken is last token of second bound expression or COMMA or RBRACK
```

## Error Handling

### Token Consumption on Errors

When a parsing error occurs:

1. **Report the error** using `p.addError()`
2. **Attempt synchronization** to find a reasonable place to continue parsing
3. **Return nil** or a partially-constructed AST node
4. **Document the POST condition** for error cases if different from success

### Example

```go
// parseBlockStatement parses a begin...end block.
// PRE: curToken is BEGIN
// POST: curToken is END on success, EOF or END on error
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
    block := &ast.BlockStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    p.nextToken() // advance past 'begin'

    // Parse statements...

    if !p.curTokenIs(lexer.END) {
        p.addError("expected 'end' to close block", ErrMissingEnd)
        // Synchronize: advance to END or EOF
        for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
            p.nextToken()
        }
    }

    return block
}
```

## Parsing Function Patterns

### Pattern 1: Statement Parsing

For parsing statements (begin/end, if/then, while/do, etc.):

```go
// parseXxxStatement parses a xxx statement.
// PRE: curToken is XXX (keyword that starts the statement)
// POST: curToken is last token of the statement
func (p *Parser) parseXxxStatement() *ast.XxxStatement {
    stmt := &ast.XxxStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    p.nextToken() // advance past keyword

    // Parse components...

    return stmt
}
```

### Pattern 2: Expression Parsing (Prefix)

For prefix expressions (literals, identifiers, unary operators):

```go
// parseXxxExpression parses a xxx expression.
// PRE: curToken is first token of the expression
// POST: curToken is last token of the expression
func (p *Parser) parseXxxExpression() ast.Expression {
    expr := &ast.XxxExpression{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    // Usually don't advance immediately for simple expressions
    // Complex expressions may need to advance

    return expr
}
```

### Pattern 3: Expression Parsing (Infix)

For infix expressions (binary operators, member access, indexing):

```go
// parseXxxExpression parses a xxx infix expression.
// PRE: curToken is the operator/infix token
// POST: curToken is last token of right operand
func (p *Parser) parseXxxExpression(left ast.Expression) ast.Expression {
    expr := &ast.XxxExpression{
        BaseNode: ast.BaseNode{Token: p.curToken},
        Left:     left,
    }

    precedence := p.curPrecedence()
    p.nextToken() // advance past operator

    expr.Right = p.parseExpression(precedence)

    return expr
}
```

### Pattern 4: Declaration Parsing

For declarations (class, function, type, const, var):

```go
// parseXxxDeclaration parses a xxx declaration.
// PRE: curToken is XXX (keyword) or TYPE (for type xxx)
// POST: curToken is END, SEMICOLON, or last token of declaration
func (p *Parser) parseXxxDeclaration() *ast.XxxDecl {
    decl := &ast.XxxDecl{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    // Expect name
    if !p.expectPeek(lexer.IDENT) {
        return nil
    }

    decl.Name = &ast.Identifier{
        BaseNode: ast.BaseNode{Token: p.curToken},
        Value:    p.curToken.Literal,
    }

    // Parse rest of declaration...

    return decl
}
```

### Pattern 5: List Parsing

For parsing comma-separated lists (parameters, arguments, array elements):

```go
// parseXxxList parses a list of xxx items.
// PRE: curToken is LPAREN, LBRACK, or first item
// POST: curToken is RPAREN, RBRACK, or delimiter after list
func (p *Parser) parseXxxList() []ast.Expression {
    list := []ast.Expression{}

    if p.curTokenIs(lexer.LPAREN) {
        p.nextToken() // advance past opening delimiter
    }

    // Handle empty list
    if p.curTokenIs(lexer.RPAREN) {
        return list
    }

    // Parse first item
    item := p.parseXxx()
    list = append(list, item)

    // Parse remaining items
    for p.peekTokenIs(lexer.COMMA) {
        p.nextToken() // move to comma
        p.nextToken() // move past comma

        item := p.parseXxx()
        list = append(list, item)
    }

    // Expect closing delimiter
    if !p.expectPeek(lexer.RPAREN) {
        return nil
    }

    return list
}
```

## Naming Conventions

### Parsing Function Names

- **Primary parsers**: `parseXxx()` - parses a complete construct starting from its first token
- **Helpers**: `parseXxxHelper()`, `parseXxxBody()` - parse a portion of a construct
- **Specialized**: `parseXxxAtToken()` - parse starting from current position (document deviation from convention)
- **Type-specific**: `parseXxxType()`, `parseXxxLiteral()`, `parseXxxDeclaration()`

### Variable Names

- `stmt` - for statement nodes
- `expr` - for expression nodes
- `decl` - for declaration nodes
- `tok` - for saved token values
- `ident` - for identifier nodes
- `params` - for parameter lists
- `args` - for argument lists

### Token References

- `p.curToken` - current token being examined
- `p.peekToken` - next token (lookahead)
- Save important tokens: `beginToken := p.curToken` before advancing

## Helper Functions

### expectPeek()

Checks if the next token matches the expected type and advances if it does:

```go
if !p.expectPeek(lexer.LPAREN) {
    return nil
}
// Now curToken is LPAREN
```

**Usage**: Use when you MUST have a specific token next.

### peekTokenIs()

Checks if the next token matches without advancing:

```go
if p.peekTokenIs(lexer.ELSE) {
    p.nextToken() // Move to ELSE
    // Parse else clause
}
```

**Usage**: Use for optional tokens or lookahead.

### curTokenIs()

Checks if current token matches:

```go
if p.curTokenIs(lexer.SEMICOLON) {
    p.nextToken() // Skip it
}
```

**Usage**: Use for current token checks.

## Common Pitfalls

### 1. Double Advancement

**Wrong**:
```go
if p.peekTokenIs(lexer.LPAREN) {
    p.nextToken() // Move to LPAREN
    if !p.expectPeek(lexer.IDENT) { // This advances AGAIN
        return nil
    }
    // Now we've skipped LPAREN entirely!
}
```

**Right**:
```go
if p.peekTokenIs(lexer.LPAREN) {
    p.nextToken() // Move to LPAREN
    p.nextToken() // Move past LPAREN
    // Now parse what's inside
}
```

### 2. Not Saving Important Tokens

**Wrong**:
```go
func (p *Parser) parseIf() *ast.IfStatement {
    // p.curToken is IF, but we don't save it
    p.nextToken()
    // Can't reference the IF token anymore!
}
```

**Right**:
```go
func (p *Parser) parseIf() *ast.IfStatement {
    ifToken := p.curToken // Save it
    stmt := &ast.IfStatement{
        BaseNode: ast.BaseNode{Token: ifToken},
    }
    p.nextToken()
    // ...
}
```

### 3. Inconsistent POST Condition

**Wrong** (function returns with curToken at different positions depending on path):
```go
func (p *Parser) parseOptional() ast.Expression {
    if condition {
        return p.parseA() // Leaves curToken at end of A
    }
    // Doesn't advance, leaves curToken at current position
    return nil
}
```

**Right**:
```go
func (p *Parser) parseOptional() ast.Expression {
    if condition {
        return p.parseA() // Leaves curToken at end of A
    }
    // Document that POST condition is same as PRE when returning nil
    return nil
}
```

### 4. Forgetting to Document Exceptions

**Wrong**:
```go
// parseExpression parses an expression.
// PRE: curToken is first token
// POST: curToken is last token
func (p *Parser) parseExpression(prec int) ast.Expression {
    // But what if there's an error? What if it's an empty expression?
}
```

**Right**:
```go
// parseExpression parses an expression.
// PRE: curToken is first token of expression
// POST: curToken is last token of expression (or unchanged if parse fails)
func (p *Parser) parseExpression(prec int) ast.Expression {
    // Implementation
}
```

## Testing Considerations

When writing tests for parsing functions:

1. **Verify PRE condition**: Ensure test setup positions tokens correctly
2. **Verify POST condition**: Check token position after parsing
3. **Test error paths**: Verify token position after errors
4. **Test edge cases**: Empty constructs, nested structures, malformed input

### Example Test

```go
func TestParseBlockStatement(t *testing.T) {
    input := "begin x := 1; y := 2; end"
    p := NewParser(input)
    p.nextToken() // Position at BEGIN

    // PRE: curToken should be BEGIN
    if p.curToken.Type != lexer.BEGIN {
        t.Fatalf("PRE condition failed: expected BEGIN, got %s", p.curToken.Type)
    }

    block := p.parseBlockStatement()

    // POST: curToken should be END
    if p.curToken.Type != lexer.END {
        t.Errorf("POST condition failed: expected END, got %s", p.curToken.Type)
    }

    // Verify block was parsed correctly
    if len(block.Statements) != 2 {
        t.Errorf("expected 2 statements, got %d", len(block.Statements))
    }
}
```

## Summary

The key principles of go-dws parser conventions:

1. **Consistent token positioning**: All functions called WITH curToken at triggering token
2. **Clear documentation**: Every function has PRE/POST conditions
3. **Predictable behavior**: Same patterns for similar constructs
4. **Error resilience**: Synchronize to reasonable continuation points
5. **Testable code**: Clear contracts make testing straightforward

Following these conventions ensures the parser is:
- **Maintainable**: Easy to understand and modify
- **Extensible**: New parsing functions follow established patterns
- **Debuggable**: Token positions are predictable
- **Reliable**: Consistent behavior reduces bugs
