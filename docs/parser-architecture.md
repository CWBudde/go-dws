# Parser Architecture

This document describes the architecture and design principles of the DWScript parser.

## Table of Contents

- [Overview](#overview)
- [Pratt Parsing](#pratt-parsing)
- [Precedence Levels](#precedence-levels)
- [Parse Function Registration](#parse-function-registration)
- [Token Consumption Conventions](#token-consumption-conventions)
- [AST Node Creation](#ast-node-creation)
- [Position Tracking](#position-tracking)
- [Error Handling and Recovery](#error-handling-and-recovery)
- [State Management](#state-management)
- [Block Context Tracking](#block-context-tracking)
- [List Parsing Helpers](#list-parsing-helpers)

## Overview

The DWScript parser implements a **Pratt parser** (also known as top-down operator precedence parser) for expressions and a **recursive descent parser** for statements and declarations. This hybrid approach provides:

- **Elegant expression parsing** with operator precedence handled naturally
- **Clear statement parsing** with straightforward control flow
- **Extensibility** through function registration patterns
- **Good error recovery** with synchronization points

### Parser Structure

```
Parser
├── Lexer (tokenization)
├── Prefix parse functions (map[TokenType]prefixParseFn)
├── Infix parse functions (map[TokenType]infixParseFn)
├── Error accumulation ([]*ParserError)
├── Block context stack ([]BlockContext)
└── Parse functions for each construct
```

### Parsing Pipeline

```
Source Code
    ↓
Lexer → Tokens
    ↓
Parser → AST
    ↓
Semantic Analyzer (future)
    ↓
Bytecode Compiler or Interpreter
```

## Pratt Parsing

The Pratt parser is a top-down operator precedence parsing technique that elegantly handles expression parsing with different operator precedences.

### Core Concept

Each token type can have:
1. A **prefix parse function** - called when the token appears at the start of an expression
2. An **infix parse function** - called when the token appears between expressions
3. A **precedence level** - determines operator priority

### How It Works

The main expression parsing loop (`parseExpression(precedence)`) works as follows:

```go
func (p *Parser) parseExpression(precedence int) ast.Expression {
    // 1. Look up prefix parse function for current token
    prefix := p.prefixParseFns[p.curToken.Type]
    if prefix == nil {
        p.noPrefixParseFnError(p.curToken.Type)
        return nil
    }

    // 2. Parse the left side (prefix)
    leftExp := prefix()

    // 3. While we haven't hit a lower precedence operator...
    for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
        // Look up infix parse function
        infix := p.infixParseFns[p.peekToken.Type]
        if infix == nil {
            return leftExp
        }

        p.nextToken()

        // 4. Parse the right side (infix)
        leftExp = infix(leftExp)
    }

    return leftExp
}
```

### Example: Parsing `3 + 5 * 2`

1. Start with `parseExpression(LOWEST)`
2. Parse `3` via prefix function → IntegerLiteral(3)
3. Peek `+` (precedence SUM)
4. Since LOWEST < SUM, parse as infix
5. Call `parseInfixExpression(IntegerLiteral(3))`
6. Parse right side with `parseExpression(SUM)`
7. Parse `5` via prefix function → IntegerLiteral(5)
8. Peek `*` (precedence PRODUCT)
9. Since SUM < PRODUCT, parse as infix
10. Call `parseInfixExpression(IntegerLiteral(5))`
11. Parse right side with `parseExpression(PRODUCT)`
12. Parse `2` via prefix function → IntegerLiteral(2)
13. Return `BinaryExpression(5, *, 2)`
14. Return `BinaryExpression(3, +, BinaryExpression(5, *, 2))`

Result: `3 + (5 * 2)` - correctly respecting precedence!

## Precedence Levels

Precedence levels are defined as integer constants, from lowest to highest:

```go
const (
    _           int = iota
    LOWEST          // Lowest precedence
    ASSIGN          // :=
    COALESCE        // ?? (higher than ASSIGN so it works in assignment RHS)
    OR              // or
    AND             // and
    EQUALS          // = <>
    LESSGREATER     // < > <= >=
    SUM             // + -
    SHIFT           // shl shr sar
    PRODUCT         // * / div mod
    PREFIX          // -x, not x, +x
    CALL            // function(args)
    INDEX           // array[index]
    MEMBER          // obj.field
)
```

### Precedence Mapping

The `precedences` map associates token types with their precedence:

```go
var precedences = map[lexer.TokenType]int{
    lexer.QUESTION_QUESTION: COALESCE,
    lexer.ASSIGN:            ASSIGN,
    lexer.OR:                OR,
    lexer.AND:               AND,
    lexer.EQ:                EQUALS,
    lexer.NOT_EQ:            EQUALS,
    lexer.PLUS:              SUM,
    lexer.MINUS:             SUM,
    lexer.ASTERISK:          PRODUCT,
    lexer.SLASH:             PRODUCT,
    lexer.LPAREN:            CALL,
    lexer.LBRACK:            INDEX,
    lexer.DOT:               MEMBER,
    // ...
}
```

### Adding New Precedence Levels

To add a new operator with custom precedence:

1. Define the precedence constant (if needed)
2. Add the token-to-precedence mapping
3. Register the appropriate parse function

Example for a hypothetical power operator `**`:

```go
// In precedence constants
const (
    // ...
    PRODUCT
    POWER   // ** (higher than PRODUCT)
    PREFIX
    // ...
)

// In precedences map
precedences[lexer.POWER] = POWER

// Register parse function
p.registerInfix(lexer.POWER, p.parseInfixExpression)
```

## Parse Function Registration

The parser uses a registration pattern to associate token types with parsing functions.

### Prefix Parse Functions

Prefix functions are called when a token appears at the start of an expression:

```go
type prefixParseFn func() ast.Expression

// Registration in New()
p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
p.registerPrefix(lexer.STRING, p.parseStringLiteral)
p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)  // unary minus
p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
p.registerPrefix(lexer.IF, p.parseIfExpression)
```

Examples of prefix positions:
- `42` - INT at start
- `-5` - MINUS at start (unary)
- `(x + 1)` - LPAREN at start
- `if x then y else z` - IF at start

### Infix Parse Functions

Infix functions are called when a token appears between expressions:

```go
type infixParseFn func(ast.Expression) ast.Expression

// Registration in New()
p.registerInfix(lexer.PLUS, p.parseInfixExpression)
p.registerInfix(lexer.LPAREN, p.parseCallExpression)
p.registerInfix(lexer.DOT, p.parseMemberAccess)
p.registerInfix(lexer.LBRACK, p.parseIndexExpression)
```

Examples of infix positions:
- `x + y` - PLUS between expressions
- `foo(args)` - LPAREN after function name
- `obj.field` - DOT between object and field
- `arr[i]` - LBRACK after array

### Why This Pattern?

This registration pattern provides:
- **Extensibility**: Add new syntax by registering functions
- **Separation of concerns**: Each function handles one construct
- **Testability**: Each parse function can be tested independently
- **Clarity**: Easy to see what tokens trigger what parsing logic

## Token Consumption Conventions

The parser follows strict conventions about token position (see Phase 2.4 for details).

### PRE/POST Documentation

Every parse function documents its token consumption contract:

```go
// parseIfStatement parses an if-then-else statement.
// PRE: curToken is IF
// POST: curToken is last token of consequence or alternative statement
func (p *Parser) parseIfStatement() *ast.IfStatement {
    // Implementation
}
```

### Token Advancement Patterns

#### Pattern 1: Statement-Level Parsing

```go
// PRE: curToken is first token of statement
// POST: curToken is last token of statement (not semicolon)
func (p *Parser) parseStatement() ast.Statement {
    switch p.curToken.Type {
    case lexer.VAR:
        return p.parseVarDeclaration()  // Handles its own advancement
    case lexer.IF:
        return p.parseIfStatement()     // Handles its own advancement
    // ...
    }
}
```

#### Pattern 2: Expression Parsing

```go
// PRE: curToken is first token of expression
// POST: curToken is last token of expression
func (p *Parser) parseExpression(precedence int) ast.Expression {
    // Parse prefix
    leftExp := p.prefixParseFns[p.curToken.Type]()

    // Parse infixes
    for precedence < p.peekPrecedence() {
        p.nextToken()  // Advance to infix operator
        leftExp = p.infixParseFns[p.curToken.Type](leftExp)
    }

    return leftExp
}
```

#### Pattern 3: expectPeek

Use `expectPeek()` when you expect a specific token next:

```go
// Expect ':' after variable name
if !p.expectPeek(lexer.COLON) {
    return nil  // expectPeek adds error and advances
}
// curToken is now COLON
```

### Common Mistakes

❌ **Wrong**: Advancing too far
```go
p.nextToken()  // move past 'if'
stmt.Condition = p.parseExpression(LOWEST)
p.nextToken()  // WRONG! parseExpression already leaves us at last token
```

✅ **Correct**: Let sub-parsers handle advancement
```go
p.nextToken()  // move past 'if'
stmt.Condition = p.parseExpression(LOWEST)
// curToken is now at last token of condition
if !p.expectPeek(lexer.THEN) {  // expectPeek will advance
    return nil
}
```

## AST Node Creation

### Node Structure

All AST nodes embed `BaseNode` for position tracking:

```go
type BaseNode struct {
    Token  lexer.Token      // First token of the node
    EndPos lexer.Position   // Position after the last character
}
```

### Creating Nodes

Standard pattern for creating AST nodes:

```go
func (p *Parser) parseIfStatement() *ast.IfStatement {
    // 1. Create node with initial token
    stmt := &ast.IfStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    // 2. Parse components
    p.nextToken()
    stmt.Condition = p.parseExpression(LOWEST)

    if !p.expectPeek(lexer.THEN) {
        return nil
    }

    p.nextToken()
    stmt.Consequence = p.parseStatement()

    // 3. Set EndPos
    stmt.EndPos = stmt.Consequence.End()

    return stmt
}
```

### Position Tracking

Every node must track its position for error reporting:

1. **Token**: Set to the first token of the construct
2. **EndPos**: Set to position after the last character

Helper method:

```go
func (p *Parser) endPosFromToken(tok lexer.Token) lexer.Position {
    pos := tok.Pos
    pos.Column += tok.Length()
    pos.Offset += tok.Length()
    return pos
}
```

Examples:

```go
// Simple literal
literal := &ast.IntegerLiteral{
    BaseNode: ast.BaseNode{
        Token:  p.curToken,
        EndPos: p.endPosFromToken(p.curToken),
    },
    Value: value,
}

// Complex expression
binary := &ast.BinaryExpression{
    TypedExpressionBase: ast.TypedExpressionBase{
        BaseNode: ast.BaseNode{
            Token:  leftToken,
            EndPos: right.End(),  // End is after right operand
        },
    },
    Left:     left,
    Operator: operator,
    Right:    right,
}
```

## Error Handling and Recovery

The parser implements panic-mode error recovery with synchronization tokens.

### Error Reporting

```go
// Add an error with automatic position tracking
p.addError("expected 'end' to close block", ErrMissingEnd)

// Add an error with context
p.addErrorWithContext("expected 'then' after condition", ErrMissingThen)
// Output: "expected 'then' after condition (in if block starting at line 10)"
```

### Synchronization

After an error, advance to a safe synchronization point:

```go
if !p.expectPeek(lexer.THEN) {
    p.addErrorWithContext("expected 'then'", ErrMissingThen)
    p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
    if !p.curTokenIs(lexer.THEN) {
        return nil
    }
}
```

Synchronization points:
- **Statement starters**: VAR, IF, WHILE, FOR, etc.
- **Block closers**: END, UNTIL, ELSE, EXCEPT, FINALLY
- **EOF**: Always stops synchronization

### Block Context Tracking

Track nested blocks for better error messages:

```go
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
    // Push context
    p.pushBlockContext("while", p.curToken.Pos)
    defer p.popBlockContext()

    // Parse...
    if stmt.Body == nil {
        // Error includes block context
        p.addErrorWithContext("expected statement after 'do'", ErrInvalidSyntax)
    }
}
```

## State Management

The parser supports speculative parsing with state save/restore.

### Use Case

When syntax is ambiguous, try one parse strategy, and if it fails, backtrack and try another:

```go
state := p.saveState()
if result := p.tryParseAsRecordLiteral(); result != nil {
    return result  // Success!
}
p.restoreState(state)  // Failed, backtrack
return p.parseAsCallExpression()
```

### What Gets Saved

`ParserState` includes:
- Current and peek tokens
- Lexer state
- Error lists
- Semantic errors
- Block context stack
- Parsing flags (e.g., `parsingPostCondition`)

### Important Notes

- Use sparingly - backtracking is expensive
- Always restore state if speculative parse fails
- Document why backtracking is needed

## Block Context Tracking

Block contexts help provide better error messages by tracking what blocks are currently being parsed.

### BlockContext Structure

```go
type BlockContext struct {
    BlockType string         // "begin", "if", "while", "for", "case", etc.
    StartPos  lexer.Position // Position where block started
    StartLine int            // Line number for error messages
}
```

### Usage Pattern

```go
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
    // Push context when entering block
    p.pushBlockContext("begin", p.curToken.Pos)
    defer p.popBlockContext()  // Pop when exiting (even on error)

    // Parse block contents...

    if !p.curTokenIs(lexer.END) {
        // Error message includes context
        p.addErrorWithContext("expected 'end'", ErrMissingEnd)
        // Output: "expected 'end' (in begin block starting at line 10)"
    }
}
```

## List Parsing Helpers

The parser provides generic list parsing helpers to reduce code duplication.

### ListParseOptions

```go
type ListParseOptions struct {
    Separator          []lexer.TokenType  // Token(s) that separate items
    Terminator         lexer.TokenType    // Token that ends the list
    AllowEmpty         bool               // Can the list be empty?
    AllowTrailing      bool               // Allow trailing separator?
    RequireTerminator  bool               // Must end with terminator?
}
```

### Example Usage

```go
// Parse comma-separated expression list: (expr1, expr2, expr3)
opts := ListParseOptions{
    Separator:         []lexer.TokenType{lexer.COMMA},
    Terminator:        lexer.RPAREN,
    AllowEmpty:        true,
    AllowTrailing:     true,
    RequireTerminator: true,
}

exprs := []ast.Expression{}
count, ok := p.parseSeparatedList(opts, func() bool {
    expr := p.parseExpression(LOWEST)
    if expr == nil {
        return false
    }
    exprs = append(exprs, expr)
    return true
})
```

### Benefits

- **Consistency**: Same logic for all lists
- **Flexibility**: Configurable behavior
- **Error handling**: Proper recovery on malformed lists
- **Maintainability**: Fix bugs in one place

## Summary

The parser architecture is built on:

1. **Pratt parsing** for elegant expression handling
2. **Precedence-based** operator resolution
3. **Registration pattern** for extensibility
4. **Strict token conventions** for predictability
5. **Position tracking** for accurate error reporting
6. **Error recovery** with synchronization
7. **State management** for backtracking
8. **Block context** for better error messages
9. **Generic helpers** for common patterns

This architecture provides a solid foundation for parsing the full DWScript language while remaining maintainable and extensible.
