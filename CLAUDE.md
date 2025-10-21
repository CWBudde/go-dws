# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-dws is a Go port of DWScript (Delphi Web Script), a full-featured Object Pascal-based scripting language. The project aims for 100% language compatibility with the original DWScript while using idiomatic Go patterns.

**Current Status**: Stage 2 complete (100%). Lexer and minimal parser with expression support are fully implemented and tested. AST coverage: 92.7%, Parser coverage: 81.9%.

## Common Commands

### Building
```bash
# Build the CLI tool
go build ./cmd/dwscript

# Build and install globally
go install ./cmd/dwscript
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./lexer
go test ./parser
go test ./ast

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run a single test
go test -run TestLexer ./lexer
go test -run TestIntegerLiteral ./parser
```

### Linting
```bash
# Run golangci-lint (project uses .golangci.yml config)
golangci-lint run

# Run standard Go tools
go vet ./...
go fmt ./...
```

### CLI Usage
```bash
# Tokenize a file
./bin/dwscript lex testdata/simple.dws
./bin/dwscript lex -e "var x: Integer := 42;"

# Parse and display AST
./bin/dwscript parse testdata/simple.dws
./bin/dwscript parse -e "3 + 5 * 2"

# Show version
./bin/dwscript version
```

## Architecture Overview

### Pipeline
```
Source Code → Lexer → Parser → AST → Semantic Analyzer → Interpreter
                                                            ↓
                                                         Output
```

### Package Structure

**lexer/** - Tokenization
- `lexer.go`: Main lexer implementation with `NextToken()` method
- `token.go`: Token types and Position tracking
- `token_type.go`: Complete enumeration of 150+ DWScript tokens (keywords, operators, literals)
- Supports: keywords, operators, string/number/boolean literals, comments (block and line)
- Handles case-insensitive keywords, hex/binary numbers, escaped strings

**parser/** - Parsing and AST construction
- `parser.go`: Pratt parser implementation with precedence climbing
- Uses prefix/infix parse functions for extensibility
- Precedence levels: LOWEST, ASSIGN, OR, AND, EQUALS, LESSGREATER, SUM, PRODUCT, PREFIX, CALL, INDEX, MEMBER
- Currently supports: expressions, literals, binary/unary operations, grouped expressions, block statements

**ast/** - Abstract Syntax Tree node definitions
- `ast.go`: Base Node, Expression, and Statement interfaces
- `expressions.go`: Expression nodes (literals, binary/unary ops, identifiers)
- `statements.go`: Statement nodes (expression statements, block statements)
- All nodes implement `String()` for debugging and `TokenLiteral()` for error reporting

**types/** - Type system (placeholder for Stage 6)
- Will contain Integer, Float, String, Boolean, Array, Class types
- Type checking and semantic analysis

**interp/** - Interpreter/runtime (placeholder for Stage 3)
- Will execute the AST
- Environment/symbol table management
- Built-in function implementations

**cmd/dwscript/** - CLI application using Cobra
- `lex` command: Tokenize and display tokens
- `parse` command: Parse and display AST
- `run` command: Execute scripts (not yet implemented)
- `version` command: Show version info

### Key Design Patterns

**Pratt Parser**: Used for expression parsing with operator precedence. Register prefix/infix parse functions for each token type that can start or continue an expression.

**AST Visitor Pattern**: Ready for semantic analysis phase. Each node type implements standard interfaces for tree traversal.

**Symbol Tables**: Will use nested scope chain for variable/function resolution (Stage 5).

## Development Guidelines

### DWScript Language Specifics

**Case Insensitivity**: DWScript keywords are case-insensitive. The lexer normalizes all keywords to lowercase via `LookupIdent()`.

**Operators**:
- Assignment: `:=` (not `=`)
- Equality: `=` (not `==`)
- Inequality: `<>`
- Integer division: `div` (keyword, not operator)
- Modulo: `mod` (keyword)
- Boolean: `and`, `or`, `xor`, `not` (keywords)
- Compound: `+=`, `-=`, `*=`, `/=`

**Comments**:
- Block: `{ ... }` or `(* ... *)`
- Line: `// ...`

**String Literals**:
- Single or double quotes
- Escape quotes by doubling: `'it''s'` → `it's`
- Multi-line strings supported

**Number Literals**:
- Integers: `42`, `-5`
- Hex: `$FF` or `0xFF`
- Binary: `%1010`
- Floats: `3.14`, `1.0e10`

### Testing Philosophy

The project maintains high test coverage (>90% for lexer, >80% for parser). When adding features:

1. Write tests FIRST (TDD approach as per PLAN.md)
2. Test both success and error cases
3. Include edge cases (empty input, malformed syntax, boundary values)
4. Use table-driven tests for multiple similar cases
5. Mirror DWScript's original test suite where possible

### Error Handling

- Parser accumulates errors (doesn't stop at first error)
- All errors include line:column position information
- Use `addError()` and `peekError()` methods in parser
- Lexer marks illegal tokens with `ILLEGAL` type

### Code Style

- Follow standard Go conventions
- Use `go fmt` and `golangci-lint`
- Comprehensive GoDoc comments on all exported types/functions
- Keep functions focused and small
- Prefer explicit error handling over panics

## Implementation Roadmap (PLAN.md)

The project follows a 10-stage incremental plan (~511 tasks):

1. **Stage 1**: Lexer ✅ COMPLETE
2. **Stage 2**: Parser (expressions) ✅ COMPLETE
3. **Stage 3**: Statements and execution (65 tasks) ✅ COMPLETE
4. **Stage 4**: Control flow (46 tasks) ✅ COMPLETE
5. **Stage 5**: Functions and scope (46 tasks) ✅ COMPLETE
6. **Stage 6**: Type checking (50 tasks) ✅ COMPLETE
7. **Stage 7**: Classes and OOP (77 tasks)
8. **Stage 8**: Advanced features (62 tasks)
9. **Stage 9**: Performance and polish (68 tasks)
10. **Stage 10**: Long-term evolution (54 tasks)

When implementing new stages:
- Create AST nodes first
- Extend parser with new syntax
- Add comprehensive tests
- Update CLI if applicable
- Document in stage summary files under `docs/`

## Important Files

- `PLAN.md`: Complete task breakdown and progress tracking
- `goal.md`: Detailed strategy and rationale for the port
- `README.md`: User-facing documentation
- `CONTRIBUTING.md`: Contribution guidelines
- `docs/stage*.md`: Stage completion summaries with statistics

## Reference Material

- Original DWScript source: `reference/dwscript-original/`
- DWScript language reference: https://www.delphitools.info/dwscript/
- Test scripts: `testdata/*.dws`

## Notes for Future Development

**Simulating Delphi Features in Go**: Since Go lacks classes, inheritance, and method overloading:
- Use ClassInfo structs for class metadata
- ObjectInstance structs with field maps for instances
- Method tables for dynamic dispatch and overriding
- Type checking at semantic analysis phase (Stage 6)

**Memory Management**: Go's GC handles cleanup; no manual reference counting needed like Delphi.

**Performance**: Current AST interpreter is simple. Stage 9 may add bytecode VM for better performance.
