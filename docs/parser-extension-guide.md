# Parser Extension Guide

This guide walks you through adding new syntax to the DWScript parser.

## Table of Contents

- [Before You Start](#before-you-start)
- [Adding a New Expression](#adding-a-new-expression)
- [Adding a New Statement](#adding-a-new-statement)
- [Adding a New Operator](#adding-a-new-operator)
- [Adding a New Declaration](#adding-a-new-declaration)
- [Testing Checklist](#testing-checklist)
- [Common Pitfalls](#common-pitfalls)
- [Example Walkthroughs](#example-walkthroughs)

## Before You Start

### Prerequisites

1. Understand the basics of Pratt parsing (see [parser-architecture.md](parser-architecture.md))
2. Read the style guide (see [parser-style-guide.md](parser-style-guide.md))
3. Have a clear specification of the new syntax
4. Add the necessary tokens to the lexer first

### Process Overview

```
1. Define AST Node     (pkg/ast/)
   ↓
2. Add Lexer Tokens    (internal/lexer/ or pkg/token/)
   ↓
3. Write Parse Function (internal/parser/)
   ↓
4. Register Function   (internal/parser/parser.go New())
   ↓
5. Write Tests         (internal/parser/*_test.go)
   ↓
6. Update Documentation
```

## Adding a New Expression

Expressions are values that can be evaluated. Examples: literals, binary operators, function calls.

### Step 1: Define the AST Node

Create the node in `pkg/ast/expressions.go`:

```go
// TernaryExpression represents a ternary conditional expression.
// Syntax: condition ? trueValue : falseValue
type TernaryExpression struct {
    TypedExpressionBase
    Condition  Expression
    TrueValue  Expression
    FalseValue Expression
}

func (te *TernaryExpression) String() string {
    return fmt.Sprintf("(%s ? %s : %s)", te.Condition, te.TrueValue, te.FalseValue)
}
```

### Step 2: Add Required Tokens

Ensure tokens exist in lexer (if needed):

```go
// In internal/lexer/token_type.go or pkg/token/token.go
QUESTION     // ?

// In internal/lexer/lexer.go
case '?':
    tok = newToken(QUESTION, l.ch)
```

### Step 3: Write the Parse Function

In `internal/parser/expressions.go`:

```go
// parseTernaryExpression parses a ternary conditional expression.
// Syntax: condition ? trueValue : falseValue
// PRE: curToken is QUESTION (called as infix after condition)
// POST: curToken is last token of false value
func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {
    expr := &ast.TernaryExpression{
        TypedExpressionBase: ast.TypedExpressionBase{
            BaseNode: ast.BaseNode{Token: condition.Token()},
        },
        Condition: condition,
    }

    // Move past '?'
    p.nextToken()

    // Parse true value
    expr.TrueValue = p.parseExpression(LOWEST)
    if expr.TrueValue == nil {
        p.addError("expected expression after '?'", ErrInvalidExpression)
        return nil
    }

    // Expect ':'
    if !p.expectPeek(lexer.COLON) {
        p.addError("expected ':' in ternary expression", ErrUnexpectedToken)
        return nil
    }

    // Move past ':'
    p.nextToken()

    // Parse false value
    expr.FalseValue = p.parseExpression(LOWEST)
    if expr.FalseValue == nil {
        p.addError("expected expression after ':'", ErrInvalidExpression)
        return nil
    }

    expr.EndPos = expr.FalseValue.End()
    return expr
}
```

### Step 4: Register the Function

In `internal/parser/parser.go`, in the `New()` function:

```go
// Register as infix (appears between expressions)
p.registerInfix(lexer.QUESTION, p.parseTernaryExpression)
```

If it's a prefix expression (appears at the start):

```go
// Register as prefix (appears at expression start)
p.registerPrefix(lexer.TYPEOF, p.parseTypeOfExpression)
```

### Step 5: Set Precedence (if needed)

In `internal/parser/parser.go`:

```go
// Add to precedence constants if needed
const (
    // ...
    EQUALS
    TERNARY     // New precedence level
    LESSGREATER
    // ...
)

// Add to precedences map
var precedences = map[lexer.TokenType]int{
    // ...
    lexer.QUESTION: TERNARY,
    // ...
}
```

### Step 6: Write Tests

In `internal/parser/expressions_test.go`:

```go
func TestParseTernaryExpression(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {
            name:  "simple ternary",
            input: "x > 0 ? 'positive' : 'non-positive'",
            want:  "((x > 0) ? 'positive' : 'non-positive')",
        },
        {
            name:  "nested ternary",
            input: "x > 0 ? (x > 10 ? 'big' : 'small') : 'negative'",
            want:  "((x > 0) ? ((x > 10) ? 'big' : 'small') : 'negative')",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            l := lexer.New(tt.input)
            p := New(l)
            program := p.ParseProgram()
            checkParserErrors(t, p)

            if len(program.Statements) != 1 {
                t.Fatalf("expected 1 statement, got %d", len(program.Statements))
            }

            stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
            if !ok {
                t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
            }

            if stmt.Expression.String() != tt.want {
                t.Errorf("want %q, got %q", tt.want, stmt.Expression.String())
            }
        })
    }
}
```

## Adding a New Statement

Statements are actions that don't return values. Examples: assignments, loops, declarations.

### Step 1: Define the AST Node

In `pkg/ast/statements.go`:

```go
// WithStatement represents a 'with' statement for accessing object members.
// Syntax: with <expression> do <statement>
type WithStatement struct {
    BaseNode
    Expression Expression  // Object to access
    Body       Statement   // Statement with implicit object access
}

func (ws *WithStatement) String() string {
    return fmt.Sprintf("with %s do %s", ws.Expression, ws.Body)
}
```

### Step 2: Write the Parse Function

In `internal/parser/statements.go` or appropriate file:

```go
// parseWithStatement parses a 'with' statement.
// Syntax: with <expression> do <statement>
// PRE: curToken is WITH
// POST: curToken is last token of body statement
func (p *Parser) parseWithStatement() *ast.WithStatement {
    stmt := &ast.WithStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    p.pushBlockContext("with", p.curToken.Pos)
    defer p.popBlockContext()

    // Move past 'with'
    p.nextToken()

    // Parse the expression
    stmt.Expression = p.parseExpression(LOWEST)
    if stmt.Expression == nil {
        p.addErrorWithContext("expected expression after 'with'", ErrInvalidExpression)
        p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
        return nil
    }

    // Expect 'do' keyword
    if !p.expectPeek(lexer.DO) {
        p.addErrorWithContext("expected 'do' after with expression", ErrMissingDo)
        p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
        if !p.curTokenIs(lexer.DO) {
            return nil
        }
    }

    // Parse the body statement
    p.nextToken()
    stmt.Body = p.parseStatement()

    if stmt.Body == nil {
        p.addErrorWithContext("expected statement after 'do'", ErrInvalidSyntax)
        p.synchronize([]lexer.TokenType{lexer.END})
        return nil
    }

    stmt.EndPos = stmt.Body.End()
    return stmt
}
```

### Step 3: Add to Statement Dispatcher

In `internal/parser/statements.go`, in `parseStatement()`:

```go
func (p *Parser) parseStatement() ast.Statement {
    switch p.curToken.Type {
    case lexer.VAR:
        return p.parseVarDeclaration()
    case lexer.IF:
        return p.parseIfStatement()
    case lexer.WITH:
        return p.parseWithStatement()  // Add here
    // ... other cases
    }
}
```

### Step 4: Write Tests

In `internal/parser/statements_test.go`:

```go
func TestParseWithStatement(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "simple with statement",
            input: "with obj do Print(x);",
            want:  "with obj do Print(x)",
        },
        {
            name:  "with block body",
            input: "with obj do begin x := 1; y := 2; end;",
            want:  "with obj do begin (x := 1); (y := 2) end",
        },
        {
            name:    "missing do keyword",
            input:   "with obj Print(x);",
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

            if program.Statements[0].String() != tt.want {
                t.Errorf("want %q, got %q", tt.want, program.Statements[0].String())
            }
        })
    }
}
```

## Adding a New Operator

Operators are special syntax for operations. Binary operators appear between operands, unary operators before operands.

### Binary Operator Example: `**` (Power)

#### Step 1: Add Token

In lexer:

```go
// In token types
POWER  // **

// In lexer
if l.peekChar() == '*' {
    ch := l.ch
    l.readChar()
    tok = token.Token{Type: token.POWER, Literal: string(ch) + string(l.ch)}
} else {
    tok = newToken(token.ASTERISK, l.ch)
}
```

#### Step 2: Set Precedence

In `internal/parser/parser.go`:

```go
const (
    // ...
    PRODUCT
    POWER    // ** (higher than multiply/divide)
    PREFIX
    // ...
)

var precedences = map[lexer.TokenType]int{
    // ...
    lexer.POWER: POWER,
    // ...
}
```

#### Step 3: Register Infix Handler

Binary operators typically use the generic infix handler:

```go
// In New()
p.registerInfix(lexer.POWER, p.parseInfixExpression)
```

The generic `parseInfixExpression` already handles binary operators correctly!

#### Step 4: Test

```go
func TestParsePowerOperator(t *testing.T) {
    input := "2 ** 3 ** 2"  // Should be 2 ** (3 ** 2) due to right-associativity

    l := lexer.New(input)
    p := New(l)
    program := p.ParseProgram()
    checkParserErrors(t, p)

    // Verify AST structure
}
```

### Unary Operator Example: `~` (Bitwise NOT)

#### Step 1: Add token (in lexer)

#### Step 2: Register Prefix Handler

```go
// In New()
p.registerPrefix(lexer.TILDE, p.parsePrefixExpression)
```

#### Step 3: Test

```go
func TestParseBitwiseNot(t *testing.T) {
    input := "~x"

    l := lexer.New(input)
    p := New(l)
    program := p.ParseProgram()
    checkParserErrors(t, p)

    // Verify unary expression
}
```

## Adding a New Declaration

Declarations introduce new names into scope. Examples: type, const, function.

### Example: Resource Declaration

```go
// Step 1: Define AST node (pkg/ast/declarations.go)
type ResourceDecl struct {
    BaseNode
    Name       *Identifier
    Type       *TypeAnnotation
    Initializer Expression
}

// Step 2: Parse function (internal/parser/declarations.go)
func (p *Parser) parseResourceDeclaration() *ast.ResourceDecl {
    decl := &ast.ResourceDecl{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    if !p.expectPeek(lexer.IDENT) {
        return nil
    }

    decl.Name = &ast.Identifier{
        BaseNode: ast.BaseNode{Token: p.curToken},
        Value:    p.curToken.Literal,
    }

    if !p.expectPeek(lexer.COLON) {
        return nil
    }

    p.nextToken()
    decl.Type = p.parseTypeAnnotation()

    if p.peekTokenIs(lexer.ASSIGN) {
        p.nextToken()
        p.nextToken()
        decl.Initializer = p.parseExpression(LOWEST)
    }

    decl.EndPos = p.endPosFromToken(p.curToken)
    return decl
}

// Step 3: Add to ParseProgram or parseDeclaration
case lexer.RESOURCE:
    return p.parseResourceDeclaration()

// Step 4: Write tests
func TestParseResourceDeclaration(t *testing.T) {
    // Test cases
}
```

## Testing Checklist

When adding new syntax, ensure you have tests for:

### ✅ Happy Path
- [ ] Minimal valid example
- [ ] With all optional parts
- [ ] Common usage patterns

### ✅ Edge Cases
- [ ] Empty constructs (if allowed)
- [ ] Single element
- [ ] Deeply nested
- [ ] Very long constructs
- [ ] All keyword combinations

### ✅ Error Cases
- [ ] Missing required tokens
- [ ] Invalid token order
- [ ] Missing expressions/statements
- [ ] Unexpected EOF
- [ ] Multiple errors in one construct

### ✅ Integration
- [ ] Works with other statements
- [ ] Works in functions/classes
- [ ] Precedence correct (for operators)
- [ ] Position tracking accurate

### ✅ Regression
- [ ] All existing tests still pass
- [ ] No performance degradation

## Common Pitfalls

### Pitfall 1: Incorrect Token Position

❌ **Wrong**:
```go
func (p *Parser) parseMyStatement() *ast.MyStatement {
    p.nextToken()  // Don't advance before checking PRE condition!
    // ...
}
```

✅ **Correct**:
```go
// PRE: curToken is MY_KEYWORD
func (p *Parser) parseMyStatement() *ast.MyStatement {
    stmt := &ast.MyStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},  // Capture token first
    }
    p.nextToken()  // Now advance
    // ...
}
```

### Pitfall 2: Forgetting Position Tracking

❌ **Wrong**:
```go
return &ast.MyExpression{
    BaseNode: ast.BaseNode{Token: p.curToken},
    // Missing EndPos!
}
```

✅ **Correct**:
```go
expr := &ast.MyExpression{
    BaseNode: ast.BaseNode{Token: firstToken},
}
// ... parse parts
expr.EndPos = lastPart.End()
return expr
```

### Pitfall 3: Not Using Block Context

❌ **Wrong**:
```go
func (p *Parser) parseMyBlockStatement() *ast.MyBlock {
    // Parse without context
    if !p.curTokenIs(lexer.END) {
        p.addError("expected 'end'", ErrMissingEnd)
    }
}
```

✅ **Correct**:
```go
func (p *Parser) parseMyBlockStatement() *ast.MyBlock {
    p.pushBlockContext("myblock", p.curToken.Pos)
    defer p.popBlockContext()

    if !p.curTokenIs(lexer.END) {
        p.addErrorWithContext("expected 'end'", ErrMissingEnd)
        p.synchronize([]lexer.TokenType{lexer.END})
    }
}
```

### Pitfall 4: Incorrect Precedence

For operators, ensure precedence matches expected evaluation order:

```go
// If you want: a + b * c to parse as a + (b * c)
// Then: PRODUCT > SUM

const (
    SUM     = 4
    PRODUCT = 5  // Higher number = higher precedence
)
```

### Pitfall 5: Missing Error Recovery

❌ **Wrong**:
```go
if !p.expectPeek(lexer.DO) {
    return nil  // Just bail out
}
```

✅ **Correct**:
```go
if !p.expectPeek(lexer.DO) {
    p.addErrorWithContext("expected 'do'", ErrMissingDo)
    p.synchronize([]lexer.TokenType{lexer.DO, lexer.END})
    if !p.curTokenIs(lexer.DO) {
        return nil
    }
    // Try to continue if we found DO
}
```

## Example Walkthroughs

### Example 1: Adding `unless` Statement

`unless` is the opposite of `if`: execute when condition is false.

Syntax: `unless <condition> then <statement>`

```go
// 1. AST Node (pkg/ast/statements.go)
type UnlessStatement struct {
    BaseNode
    Condition   Expression
    Consequence Statement
}

// 2. Parse function (internal/parser/control_flow.go)
func (p *Parser) parseUnlessStatement() *ast.UnlessStatement {
    stmt := &ast.UnlessStatement{
        BaseNode: ast.BaseNode{Token: p.curToken},
    }

    p.pushBlockContext("unless", p.curToken.Pos)
    defer p.popBlockContext()

    p.nextToken()
    stmt.Condition = p.parseExpression(LOWEST)

    if stmt.Condition == nil {
        p.addErrorWithContext("expected condition", ErrInvalidExpression)
        p.synchronize([]lexer.TokenType{lexer.THEN, lexer.END})
        return nil
    }

    if !p.expectPeek(lexer.THEN) {
        p.addErrorWithContext("expected 'then'", ErrMissingThen)
        p.synchronize([]lexer.TokenType{lexer.THEN, lexer.END})
        if !p.curTokenIs(lexer.THEN) {
            return nil
        }
    }

    p.nextToken()
    stmt.Consequence = p.parseStatement()

    if stmt.Consequence == nil {
        p.addErrorWithContext("expected statement", ErrInvalidSyntax)
        p.synchronize([]lexer.TokenType{lexer.END})
        return nil
    }

    stmt.EndPos = stmt.Consequence.End()
    return stmt
}

// 3. Register (internal/parser/statements.go)
case lexer.UNLESS:
    return p.parseUnlessStatement()

// 4. Tests (internal/parser/control_flow_test.go)
func TestParseUnlessStatement(t *testing.T) {
    input := "unless x > 10 then Print('small');"
    // ... test implementation
}
```

### Example 2: Adding `..` Range Operator

Already implemented in `internal/parser/control_flow.go` for case statements. See `parseCaseStatement()` for reference.

## Summary

Adding new syntax involves:

1. **Define AST node** with proper structure
2. **Write parse function** following conventions
3. **Register function** (prefix, infix, or statement dispatcher)
4. **Set precedence** (for operators)
5. **Write comprehensive tests** covering all cases
6. **Update documentation** as needed

Always follow the patterns established in existing code, and when in doubt, refer to similar constructs for guidance.

Happy parsing! 🎉
