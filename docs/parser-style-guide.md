# Parser Style Guide

This guide establishes coding standards and best practices for the DWScript parser.

## Table of Contents

- [Naming Conventions](#naming-conventions)
- [Function Documentation](#function-documentation)
- [PRE/POST Conditions](#prepost-conditions)
- [Error Handling](#error-handling)
- [Testing Requirements](#testing-requirements)
- [Code Organization](#code-organization)
- [Common Patterns](#common-patterns)
- [Anti-Patterns](#anti-patterns)

## Naming Conventions

### Parse Functions

All parsing functions follow a consistent naming pattern:

```go
// Pattern: parse<ConstructName>
func (p *Parser) parseIfStatement() *ast.IfStatement
func (p *Parser) parseIntegerLiteral() ast.Expression
func (p *Parser) parseArrayLiteral() ast.Expression
func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression
```

**Rules**:
- Start with `parse`
- Use the construct name (IfStatement, IntegerLiteral, etc.)
- Camel case
- Be specific: `parseIfStatement` not `parseIf`

### Helper Functions

Helper functions use descriptive names explaining their purpose:

```go
// Check functions: return bool
func (p *Parser) isRecordLiteralPattern() bool
func (p *Parser) looksLikeVarDeclaration() bool

// Conversion functions
func (p *Parser) endPosFromToken(tok lexer.Token) lexer.Position

// Synchronization and recovery
func (p *Parser) synchronize(tokens []lexer.TokenType)
func (p *Parser) pushBlockContext(blockType string, pos lexer.Position)
```

### Variables

```go
// AST nodes: descriptive noun
stmt := &ast.IfStatement{}
expr := &ast.BinaryExpression{}
block := &ast.BlockStatement{}

// Tokens: curToken, peekToken, firstToken, etc.
opToken := p.curToken
nameToken := p.curToken

// Collections: plural
statements := []ast.Statement{}
parameters := []*ast.Parameter{}
```

### Constants

```go
// Precedence levels: ALL_CAPS
const (
    LOWEST
    ASSIGN
    OR
    AND
)

// Error codes: Err prefix, PascalCase
const (
    ErrUnexpectedToken  = "E_UNEXPECTED_TOKEN"
    ErrMissingEnd       = "E_MISSING_END"
    ErrInvalidSyntax    = "E_INVALID_SYNTAX"
)
```

## Function Documentation

Every parsing function must have documentation following this template:

```go
// parse<ConstructName> parses a <construct description>.
// Syntax: <BNF or informal grammar>
// PRE: curToken is <expected token>
// POST: curToken is <resulting token>
func (p *Parser) parseConstruct() *ast.Construct {
    // Implementation
}
```

### Example

```go
// parseWhileStatement parses a while loop statement.
// Syntax: while <condition> do <statement>
// PRE: curToken is WHILE
// POST: curToken is last token of body statement
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
    stmt := &ast.WhileStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    p.pushBlockContext("while", p.curToken.Pos)
    defer p.popBlockContext()

    p.nextToken()
    stmt.Condition = p.parseExpression(LOWEST)

    if stmt.Condition == nil {
        p.addErrorWithContext("expected condition after 'while'", ErrInvalidExpression)
        p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
        return nil
    }

    if !p.expectPeek(lexer.DO) {
        p.addErrorWithContext("expected 'do' after while condition", ErrMissingDo)
        p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
        if !p.curTokenIs(lexer.DO) {
            return nil
        }
    }

    p.nextToken()
    stmt.Body = p.parseStatement()

    if isNilStatement(stmt.Body) {
        p.addErrorWithContext("expected statement after 'do'", ErrInvalidSyntax)
        p.synchronize([]lexer.TokenType{lexer.END})
        return nil
    }

    stmt.EndPos = stmt.Body.End()
    return stmt
}
```

## PRE/POST Conditions

PRE/POST conditions are **mandatory** for all parsing functions. They document the contract between the caller and the function.

### PRE Condition

States what the parser's current token position must be when the function is called.

```go
// PRE: curToken is IF
// Caller must ensure: p.curToken.Type == lexer.IF
```

### POST Condition

States what the parser's current token position will be when the function returns successfully.

```go
// POST: curToken is last token of statement
// After return: p.curToken is at the final token of the parsed construct
```

### Examples

```go
// PRE: curToken is BEGIN
// POST: curToken is END
func (p *Parser) parseBlockStatement() *ast.BlockStatement

// PRE: curToken is first token of expression
// POST: curToken is last token of expression
func (p *Parser) parseExpression(precedence int) ast.Expression

// PRE: curToken is LPAREN
// POST: curToken is RPAREN
func (p *Parser) parseGroupedExpression() ast.Expression
```

### Why PRE/POST?

- **Contract clarity**: Explicit about token consumption
- **Debug aid**: Easy to verify correct usage
- **Prevents bugs**: Mismatched expectations caught early
- **Documentation**: Self-documenting code flow

## Error Handling

The parser uses multiple error handling strategies for robustness.

### Adding Errors

```go
// Simple error with automatic position
p.addError("expected 'end' to close block", ErrMissingEnd)

// Error with block context
p.addErrorWithContext("expected 'then' after condition", ErrMissingThen)
// Output: "expected 'then' after condition (in if block starting at line 10)"

// Generic error (less preferred)
p.addGenericError("something went wrong")

// Peek error for expectPeek failures
p.peekError(lexer.SEMICOLON)
```

### Error Recovery with Synchronization

After detecting an error, use synchronization to advance to a safe point:

```go
if stmt.Condition == nil {
    p.addErrorWithContext("expected condition", ErrInvalidExpression)
    p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
    return nil  // Cannot continue without condition
}

if !p.expectPeek(lexer.THEN) {
    p.addErrorWithContext("expected 'then'", ErrMissingThen)
    p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
    if !p.curTokenIs(lexer.THEN) {
        return nil  // Could not recover
    }
    // Continue - we found THEN
}
```

### Block Context

Always push/pop block context for block-level constructs:

```go
func (p *Parser) parseForStatement() ast.Statement {
    p.pushBlockContext("for", p.curToken.Pos)
    defer p.popBlockContext()  // Always pops, even on error

    // Parse for loop...
}
```

### Error Handling Checklist

- ✅ Use descriptive error messages
- ✅ Include error codes (ErrMissingEnd, etc.)
- ✅ Use addErrorWithContext for block-level errors
- ✅ Synchronize after errors to enable multiple error reporting
- ✅ Return nil for critical failures (cannot continue)
- ✅ Try to recover for minor issues
- ✅ Push/pop block context with defer

## Testing Requirements

Every parse function must have comprehensive tests.

### Minimum Test Coverage

For each parsing function, write tests for:

1. **Happy path**: Valid, correct syntax
2. **Edge cases**: Empty, single-element, maximum nesting
3. **Error cases**: Missing tokens, invalid syntax
4. **Integration**: Interaction with other constructs

### Test Structure

Use table-driven tests:

```go
func TestParseWhileStatement(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string  // Expected AST string representation
        wantErr bool
    }{
        {
            name:  "simple while loop",
            input: "while x > 0 do x := x - 1;",
            want:  "while (x > 0) (x := (x - 1))",
        },
        {
            name:  "while with block body",
            input: "while x > 0 do begin x := x - 1; end;",
            want:  "while (x > 0) begin (x := (x - 1)) end",
        },
        {
            name:    "missing do keyword",
            input:   "while x > 0 x := x - 1;",
            wantErr: true,
        },
        {
            name:    "missing condition",
            input:   "while do x := x - 1;",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            l := lexer.New(tt.input)
            p := New(l)
            program := p.ParseProgram()

            if tt.wantErr {
                if len(p.Errors()) == 0 {
                    t.Fatal("expected error, got none")
                }
                return
            }

            checkParserErrors(t, p)

            if len(program.Statements) != 1 {
                t.Fatalf("expected 1 statement, got %d", len(program.Statements))
            }

            stmt := program.Statements[0]
            if stmt.String() != tt.want {
                t.Errorf("want %q, got %q", tt.want, stmt.String())
            }
        })
    }
}
```

### Test File Organization

- One test file per source file: `expressions.go` → `expressions_test.go`
- Group related tests: `parser_test.go`, `control_flow_test.go`, `expressions_test.go`
- Use descriptive test names: `TestParseWhileStatement`, `TestParseArrayLiteral`

### Code Coverage

- Target: **> 80%** coverage for all parser code
- Critical paths: **100%** coverage (error recovery, position tracking)
- Run: `go test -coverprofile=coverage.out ./internal/parser`

## Code Organization

### File Structure

Organize parsing functions by construct type:

```
internal/parser/
├── parser.go            # Core parser, precedence, registration
├── error.go             # Error types and codes
├── helpers.go           # Generic helpers (list parsing, etc.)
├── expressions.go       # Expression parsing
├── statements.go        # Statement parsing
├── control_flow.go      # If, while, for, case, etc.
├── declarations.go      # Var, const, type declarations
├── functions.go         # Function and procedure parsing
├── classes.go           # Class and object-oriented constructs
├── arrays.go            # Array-related parsing
├── records.go           # Record-related parsing
└── *_test.go            # Tests for each file
```

### Function Ordering

Within each file:

1. Public API functions first
2. Main parsing functions (alphabetical)
3. Helper functions (alphabetical)
4. Internal utilities last

### Import Organization

```go
import (
    "fmt"
    "strconv"

    "github.com/cwbudde/go-dws/internal/ast"
    "github.com/cwbudde/go-dws/internal/lexer"
)
```

Order:
1. Standard library
2. Blank line
3. External packages
4. Blank line
5. Internal packages

## Common Patterns

### Pattern 1: Simple Statement

```go
func (p *Parser) parseBreakStatement() *ast.BreakStatement {
    stmt := &ast.BreakStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    // EndPos is right after 'break' keyword
    stmt.EndPos = p.endPosFromToken(p.curToken)

    return stmt
}
```

### Pattern 2: Statement with Expression

```go
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
    stmt := &ast.ReturnStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    p.nextToken()

    if p.curTokenIs(lexer.SEMICOLON) || p.curTokenIs(lexer.END) {
        // No return value
        stmt.EndPos = p.endPosFromToken(p.curToken)
        return stmt
    }

    stmt.ReturnValue = p.parseExpression(LOWEST)

    if stmt.ReturnValue != nil {
        stmt.EndPos = stmt.ReturnValue.End()
    }

    return stmt
}
```

### Pattern 3: Block with Context

```go
func (p *Parser) parseIfStatement() *ast.IfStatement {
    stmt := &ast.IfStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    p.pushBlockContext("if", p.curToken.Pos)
    defer p.popBlockContext()

    p.nextToken()
    stmt.Condition = p.parseExpression(LOWEST)

    if stmt.Condition == nil {
        p.addErrorWithContext("expected condition", ErrInvalidExpression)
        p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
        return nil
    }

    if !p.expectPeek(lexer.THEN) {
        p.addErrorWithContext("expected 'then'", ErrMissingThen)
        p.synchronize([]lexer.TokenType{lexer.THEN, lexer.ELSE, lexer.END})
        if !p.curTokenIs(lexer.THEN) {
            return nil
        }
    }

    p.nextToken()
    stmt.Consequence = p.parseStatement()

    if stmt.Consequence == nil {
        p.addErrorWithContext("expected statement after 'then'", ErrInvalidSyntax)
        p.synchronize([]lexer.TokenType{lexer.ELSE, lexer.END})
        return nil
    }

    if p.peekTokenIs(lexer.ELSE) {
        p.nextToken()
        p.nextToken()
        stmt.Alternative = p.parseStatement()

        if stmt.Alternative != nil {
            stmt.EndPos = stmt.Alternative.End()
        }
    } else {
        stmt.EndPos = stmt.Consequence.End()
    }

    return stmt
}
```

## Anti-Patterns

### ❌ Don't: Advance Token Without Checking

```go
// BAD
p.nextToken()
stmt.Expression = p.parseExpression(LOWEST)
p.nextToken()  // WRONG! parseExpression already advanced
```

### ✅ Do: Trust Sub-Parser Contracts

```go
// GOOD
p.nextToken()
stmt.Expression = p.parseExpression(LOWEST)
// curToken is now at last token of expression (per POST condition)
```

### ❌ Don't: Ignore Errors

```go
// BAD
stmt.Condition = p.parseExpression(LOWEST)
// Continue without checking if condition is nil
```

### ✅ Do: Check and Handle Errors

```go
// GOOD
stmt.Condition = p.parseExpression(LOWEST)
if stmt.Condition == nil {
    p.addErrorWithContext("expected condition", ErrInvalidExpression)
    p.synchronize([]lexer.TokenType{lexer.THEN, lexer.END})
    return nil
}
```

### ❌ Don't: Mix Token Consumption Styles

```go
// BAD - inconsistent
p.nextToken()
if p.peekToken.Type == lexer.COLON {
    p.nextToken()  // Manual advance
}
if !p.expectPeek(lexer.SEMICOLON) {  // Different style
    return nil
}
```

### ✅ Do: Be Consistent

```go
// GOOD - consistent use of expectPeek
p.nextToken()
if !p.expectPeek(lexer.COLON) {
    return nil
}
if !p.expectPeek(lexer.SEMICOLON) {
    return nil
}
```

### ❌ Don't: Forget Block Context

```go
// BAD - no context for error messages
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
    // Parse...
    if !p.curTokenIs(lexer.DO) {
        p.addError("expected 'do'", ErrMissingDo)  // Generic error
    }
}
```

### ✅ Do: Always Use Block Context

```go
// GOOD - contextual error messages
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
    p.pushBlockContext("while", p.curToken.Pos)
    defer p.popBlockContext()

    // Parse...
    if !p.curTokenIs(lexer.DO) {
        p.addErrorWithContext("expected 'do'", ErrMissingDo)
        // Output: "expected 'do' (in while block starting at line 5)"
    }
}
```

### ❌ Don't: Skip Position Tracking

```go
// BAD - no EndPos
return &ast.IfStatement{
    BaseNode:    ast.BaseNode{Token: p.curToken},
    Condition:   condition,
    Consequence: consequence,
    // Missing EndPos!
}
```

### ✅ Do: Always Set EndPos

```go
// GOOD - complete position tracking
stmt := &ast.IfStatement{
    BaseNode:    ast.BaseNode{Token: p.curToken},
    Condition:   condition,
    Consequence: consequence,
}
stmt.EndPos = consequence.End()
return stmt
```

## Summary

Follow these principles for clean, maintainable parser code:

1. **Consistent naming** following established patterns
2. **Complete documentation** with PRE/POST conditions
3. **Robust error handling** with synchronization and context
4. **Comprehensive testing** with edge cases and error scenarios
5. **Proper organization** by construct type
6. **Common patterns** for similar constructs
7. **Avoid anti-patterns** that lead to bugs

When in doubt, look at existing code for similar constructs and follow the same patterns.
