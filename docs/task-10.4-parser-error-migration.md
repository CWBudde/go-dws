# Task 10.4: Parser Error Migration Guide

**Status**: ⏸️ DEFERRED
**Estimated Effort**: 4-6 hours
**Priority**: Low (current conversion layer works adequately)

## Why This Task Exists

Currently, parser errors are stored as formatted strings (`[]string`) with position information embedded in the message (e.g., "expected semicolon at 10:5"). While functional, this approach has limitations:

- Position extraction requires string parsing in `convertParserErrors()`
- Error spans are less precise (single point vs. multi-token ranges)
- Inconsistent with semantic analyzer's structured errors
- Missing error codes for programmatic handling

## Current State

### Parser Error Format
```
Location: internal/parser/parser.go:68
Type: []string

Example errors:
- "no prefix parse function for EOF found at 1:10"
- "expected next token to be SEMICOLON, got EOF instead at 1:10"
```

### Conversion Layer
```
Location: pkg/dwscript/dwscript.go:98-137
Function: convertParserErrors(parserErrors []string) []*Error

Purpose: Extracts position info from error strings via string parsing
Limitations: Best-effort extraction, may miss complex error messages
```

## Implementation Plan

### Step 1: Create ParserError Struct

**File**: `internal/parser/error.go` (NEW)

```go
package parser

import (
    "fmt"
    "github.com/cwbudde/go-dws/internal/lexer"
)

// ParserError represents a structured parse error with position information
type ParserError struct {
    Message string         // Human-readable error message
    Pos     lexer.Position // Position where error occurred
    Length  int            // Length of error span (from token)
    Code    string         // Error code for programmatic handling
}

// Error implements the error interface
func (e *ParserError) Error() string {
    if e.Code != "" {
        return fmt.Sprintf("%s at %s [%s]", e.Message, e.Pos.String(), e.Code)
    }
    return fmt.Sprintf("%s at %s", e.Message, e.Pos.String())
}

// NewParserError creates a new parser error from the current token
func NewParserError(pos lexer.Position, length int, message, code string) *ParserError {
    return &ParserError{
        Message: message,
        Pos:     pos,
        Length:  length,
        Code:    code,
    }
}
```

### Step 2: Define Error Codes

Add these constants to `internal/parser/error.go`:

```go
// Common parser error codes
const (
    ErrUnexpectedToken   = "E_UNEXPECTED_TOKEN"   // Encountered unexpected token
    ErrMissingSemicolon  = "E_MISSING_SEMICOLON"  // Expected semicolon not found
    ErrMissingEnd        = "E_MISSING_END"        // Expected 'end' keyword not found
    ErrMissingRParen     = "E_MISSING_RPAREN"     // Expected closing parenthesis
    ErrMissingRBracket   = "E_MISSING_RBRACKET"   // Expected closing bracket
    ErrMissingRBrace     = "E_MISSING_RBRACE"     // Expected closing brace
    ErrInvalidExpression = "E_INVALID_EXPRESSION" // Invalid expression syntax
    ErrNoPrefixParse     = "E_NO_PREFIX_PARSE"    // No prefix parse function for token
    ErrExpectedIdent     = "E_EXPECTED_IDENT"     // Expected identifier
    ErrExpectedType      = "E_EXPECTED_TYPE"      // Expected type name
    ErrExpectedOperator  = "E_EXPECTED_OPERATOR"  // Expected operator
)
```

### Step 3: Update Parser Struct

**File**: `internal/parser/parser.go`

**Current (line 68)**:
```go
errors []string
```

**New**:
```go
errors []*ParserError
```

**Also Update**:
```go
// Constructor (line 80)
errors: []*ParserError{},

// Errors() method - return type
func (p *Parser) Errors() []*ParserError {
    return p.errors
}
```

### Step 4: Update Error Helper Methods

**File**: `internal/parser/parser.go`

**Current peekError() (lines 208-212)**:
```go
func (p *Parser) peekError(t lexer.TokenType) {
    msg := fmt.Sprintf("expected next token to be %s, got %s instead at %d:%d",
        t, p.peekToken.Type, p.peekToken.Pos.Line, p.peekToken.Pos.Column)
    p.errors = append(p.errors, msg)
}
```

**New peekError()**:
```go
func (p *Parser) peekError(t lexer.TokenType) {
    msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
    err := NewParserError(
        p.peekToken.Pos,
        p.peekToken.Length(),
        msg,
        ErrUnexpectedToken,
    )
    p.errors = append(p.errors, err)
}
```

**Current addError() (lines 214-218)**:
```go
func (p *Parser) addError(msg string) {
    fullMsg := fmt.Sprintf("%s at %d:%d", msg, p.curToken.Pos.Line, p.curToken.Pos.Column)
    p.errors = append(p.errors, fullMsg)
}
```

**New addError()**:
```go
func (p *Parser) addError(msg string, code string) {
    err := NewParserError(
        p.curToken.Pos,
        p.curToken.Length(),
        msg,
        code,
    )
    p.errors = append(p.errors, err)
}

// Convenience method for generic errors
func (p *Parser) addGenericError(msg string) {
    p.addError(msg, ErrInvalidExpression)
}
```

### Step 5: Migrate Error Sites

Update error creation across all parser files. Work systematically through each file:

#### Migration Pattern

**OLD**:
```go
p.errors = append(p.errors,
    fmt.Sprintf("expected semicolon at %d:%d",
        p.curToken.Pos.Line, p.curToken.Pos.Column))
```

**NEW**:
```go
p.errors = append(p.errors, NewParserError(
    p.curToken.Pos,
    p.curToken.Length(),
    "expected semicolon",
    ErrMissingSemicolon,
))

// OR use helper:
p.addError("expected semicolon", ErrMissingSemicolon)
```

#### Files to Update (in priority order):

1. **parser.go** (~20 error sites)
   - Core parsing errors
   - Focus: `peekError()`, `addError()`, `expectPeek()`, `expectCurrent()`

2. **expressions.go** (~30 error sites)
   - Expression parsing errors
   - Focus: Prefix/infix parse functions, literal parsing

3. **statements.go** (~40 error sites)
   - Statement parsing errors
   - Focus: Variable declarations, assignments, compound statements

4. **classes.go** (~25 error sites)
   - Class/OOP parsing errors
   - Focus: Class declarations, methods, constructors

5. **interfaces.go** (~15 error sites)
   - Interface parsing errors
   - Focus: Interface declarations, method signatures

6. **records.go** (~15 error sites)
   - Record parsing errors
   - Focus: Record type declarations, field definitions

7. **sets.go** (~10 error sites)
   - Set parsing errors
   - Focus: Set literals, set operations

8. **enums.go** (~10 error sites)
   - Enum parsing errors
   - Focus: Enum declarations, enum values

9. **exceptions.go** (~15 error sites)
   - Exception handling parsing errors
   - Focus: Try/except/finally, raise statements

10. **properties.go** (~15 error sites)
    - Property parsing errors
    - Focus: Property declarations, read/write specifiers

11. **control_flow.go** (~10 error sites)
    - Control flow parsing errors
    - Focus: If/while/for statements

12. **functions.go** (~10 error sites)
    - Function parsing errors
    - Focus: Function/procedure declarations

13. **operators.go** (~5 error sites)
    - Operator parsing errors
    - Focus: Operator overloading

14. **Other files** (~20+ error sites)
    - Arrays, units, namespaces, etc.

### Step 6: Update Public API

**File**: `pkg/dwscript/dwscript.go`

**Current (lines 69-76)**:
```go
if len(p.Errors()) > 0 {
    // Convert parser string errors to structured errors
    errors := convertParserErrors(p.Errors())
    return nil, &CompileError{
        Stage:  "parsing",
        Errors: errors,
    }
}
```

**New**:
```go
if len(p.Errors()) > 0 {
    // Convert parser errors to public Error type
    errors := make([]*Error, 0, len(p.Errors()))
    for _, perr := range p.Errors() {
        errors = append(errors, &Error{
            Message:  perr.Message,
            Line:     perr.Pos.Line,
            Column:   perr.Pos.Column,
            Length:   perr.Length,
            Severity: SeverityError,
            Code:     perr.Code,
        })
    }
    return nil, &CompileError{
        Stage:  "parsing",
        Errors: errors,
    }
}
```

**Remove** (lines 98-137):
```go
// Delete convertParserErrors() function - no longer needed
// Delete findLastIndex() function - no longer needed
```

### Step 7: Update Tests

Many parser tests check error strings. Update them to check structured errors:

**Example - Before**:
```go
if len(parser.Errors()) != 1 {
    t.Fatalf("expected 1 error, got %d", len(parser.Errors()))
}
if !strings.Contains(parser.Errors()[0], "expected semicolon") {
    t.Errorf("wrong error message: %s", parser.Errors()[0])
}
```

**Example - After**:
```go
if len(parser.Errors()) != 1 {
    t.Fatalf("expected 1 error, got %d", len(parser.Errors()))
}
err := parser.Errors()[0]
if !strings.Contains(err.Message, "expected semicolon") {
    t.Errorf("wrong error message: %s", err.Message)
}
if err.Code != ErrMissingSemicolon {
    t.Errorf("wrong error code: %s", err.Code)
}
if err.Pos.Line != 1 || err.Pos.Column != 10 {
    t.Errorf("wrong position: %d:%d", err.Pos.Line, err.Pos.Column)
}
```

### Step 8: Verify and Test

1. **Build**: `go build ./...`
2. **Run all tests**: `go test ./...`
3. **Test CLI**: `./bin/dwscript parse -e "var x :="`
4. **Check error output**: Should show structured error with position
5. **Run fixture tests**: Verify no regressions

## Benefits When Complete

### For LSP Integration
- Precise error positions without string parsing
- Error spans for multi-token constructs
- Error codes for diagnostic filtering
- Consistent error handling across all phases

### For Developers
- Easier to debug parser errors
- Better error messages with context
- Programmatic error handling
- Consistent error patterns

### For Users
- More helpful error messages
- Better IDE integration
- Quick fixes based on error codes

## Maintenance Notes

### Adding New Parse Errors
When adding new parsing functionality:

1. Add error code constant in `parser/error.go`
2. Use `NewParserError()` or helper methods
3. Always include:
   - Clear message (no position info in message)
   - Current token position
   - Token length for span
   - Appropriate error code

### Error Code Naming Convention
- Prefix with `E_` for errors
- Use SCREAMING_SNAKE_CASE
- Be descriptive: `E_MISSING_SEMICOLON` not `E_SEMI`
- Group related codes: `E_MISSING_*` for missing tokens

## Testing Strategy

### Unit Tests
- Test error creation with correct positions
- Test error codes are set properly
- Test error messages are clear

### Integration Tests
- Parse malformed code, check structured errors
- Verify error positions match source locations
- Test error span highlighting

### Regression Tests
- Run full test suite after migration
- Verify no change in error detection (count/locations)
- Check CLI output formatting

## Rollback Plan

If issues arise during migration:

1. Keep conversion layer as backup
2. Add feature flag: `useStructuredParserErrors`
3. Run both systems in parallel temporarily
4. Compare outputs for consistency

## References

- [PLAN.md Task 10.4](../PLAN.md#L880-L970) - Original task definition
- [phase10-error-handling.md](./phase10-error-handling.md) - Phase 10 summary
- [internal/semantic/errors.go](../internal/semantic/errors.go) - Reference implementation
- [pkg/dwscript/error.go](../pkg/dwscript/error.go) - Public error type

## Questions?

If you're implementing this task and have questions:

1. Check if semantic analyzer has similar pattern ([internal/semantic/errors.go](../internal/semantic/errors.go))
2. Look at existing error handling in parser
3. Review LSP diagnostic format requirements
4. Consider adding more error codes for better IDE integration
