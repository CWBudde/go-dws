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

# Run a script (AST interpreter - default)
./bin/dwscript run script.dws

# Run with bytecode VM (5-6x faster)
./bin/dwscript run --bytecode script.dws

# Show disassembled bytecode
./bin/dwscript run --bytecode --trace script.dws

# Run with custom recursion limit (default: 1024)
./bin/dwscript run --max-recursion 2048 script.dws

# Show version
./bin/dwscript version
```

## Architecture Overview

### Pipeline
```
                                    ┌→ AST Interpreter → Output
                                    │
Source Code → Lexer → Parser → AST → Semantic Analyzer
                                    │
                                    └→ Bytecode Compiler → Bytecode VM → Output
                                                              (5-6x faster)
```

### Package Structure

The project follows standard Go project layout with `cmd/`, `internal/`, and `pkg/` directories:

**cmd/** - Command-line applications
- `cmd/dwscript/` - CLI tool for running DWScript programs
  - `lex` command: Tokenize and display tokens
  - `parse` command: Parse and display AST
  - `run` command: Execute scripts
  - `version` command: Show version info

**internal/** - Private implementation (not importable by external projects)
- `internal/lexer/` - Tokenization
  - `lexer.go`: Main lexer implementation with `NextToken()` method
  - `token.go`: Token types and Position tracking
  - `token_type.go`: Complete enumeration of 150+ DWScript tokens
  - Handles case-insensitive keywords, hex/binary numbers, escaped strings

- `internal/parser/` - Parsing and AST construction
  - `parser.go`: Pratt parser with precedence climbing
  - Uses prefix/infix parse functions for extensibility
  - Precedence levels: LOWEST, ASSIGN, OR, AND, EQUALS, LESSGREATER, SUM, PRODUCT, PREFIX, CALL, INDEX, MEMBER

- `internal/ast/` - Abstract Syntax Tree node definitions
  - `ast.go`: Base Node, Expression, and Statement interfaces
  - `expressions.go`: Expression nodes (literals, binary/unary ops, identifiers)
  - `statements.go`: Statement nodes
  - All nodes implement `String()` for debugging and `TokenLiteral()` for error reporting

- `internal/semantic/` - Semantic analysis and type checking
  - `analyzer.go`: Type checker and semantic analyzer
  - `symbol_table.go`: Symbol table for scope management

- `internal/types/` - Type system implementation
  - Integer, Float, String, Boolean, Array, Record, Enum, Class types
  - Type checking and conversion

- `internal/interp/` - AST Interpreter/runtime
  - Executes the AST via tree-walking
  - Environment/symbol table management
  - Built-in function implementations

- `internal/bytecode/` - Bytecode VM (5-6x faster than AST interpreter)
  - `compiler.go`: AST-to-bytecode compiler with optimizations
  - `vm.go`: Stack-based virtual machine with built-in function support
  - `bytecode.go`: Bytecode format, constant pools, value types
  - `disasm.go`: Bytecode disassembler for debugging
  - `instruction.go`: 116 opcodes for DWScript operations
  - See [docs/bytecode-vm.md](docs/bytecode-vm.md) for details

- `internal/errors/` - Error handling utilities
  - Error formatting and reporting

**pkg/** - Public APIs (importable by external projects)
- `pkg/dwscript/` - High-level embedding API
  - `dwscript.go`: Engine, Program, Result types
  - `options.go`: Configuration options
  - Public API for embedding DWScript in Go applications
  - See [README.md](README.md) for usage examples

- `pkg/platform/` - Platform abstraction layer (planned for Stage 10.15)
  - Abstracts filesystem, console, and platform-specific functionality
  - Enables native and WebAssembly builds with consistent behavior

- `pkg/wasm/` - WebAssembly support (planned for Stage 10.15)
  - JavaScript/Go interop
  - Browser API bindings

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

**Enumerated Types** (Stage 8):
- Basic declaration: `type TColor = (Red, Green, Blue);`
- Explicit values: `type TStatus = (Ok = 0, Error = 1);`
- Mixed values: `type TPriority = (Low, Medium = 5, High);`  // High = 6
- Scoped access: `var color := TColor.Red;`
- Unscoped access: `var color := Red;`
- Built-ins: `Ord(enumValue)`, `Integer(enumValue)`
- See `docs/enums.md` for complete documentation

**Recursion Limits** (Task 9.1-9.12):
- Default maximum recursion depth: 1024 (matches DWScript's `cDefaultMaxRecursionDepth`)
- Configurable via API: `dwscript.WithMaxRecursionDepth(2048)`
- Configurable via CLI: `--max-recursion 2048`
- Raises `EScriptStackOverflow` exception when limit exceeded
- Protected call sites: user functions, lambdas, record methods (instance & static), class methods (instance & static)
- Call stack tracking for accurate stack traces in exceptions
- See `internal/interp/recursion_test.go` for comprehensive test cases

**Function/Method Overloading** (Tasks 9.40-9.72):
- Functions, procedures, methods, and constructors can be overloaded
- All overloaded declarations must use the `overload` directive
- Overloads distinguished by: parameter count, types, and modifiers (var/const/lazy)
- Return types alone cannot distinguish overloads (ambiguous)
- Forward declarations and implementations must have matching signatures (including default parameters)
- Resolution at compile-time based on argument types
- Example:
  ```pascal
  function Max(a, b: Integer): Integer; overload;
  begin
    if a > b then Result := a else Result := b;
  end;

  function Max(a, b: Float): Float; overload;
  begin
    if a > b then Result := a else Result := b;
  end;

  PrintLn(Max(1, 2));      // Calls Integer version
  PrintLn(Max(1.5, 2.3));  // Calls Float version
  ```
- See `testdata/fixtures/OverloadsPass/` for comprehensive examples
- See `internal/semantic/overload_resolution.go` for resolution algorithm

### Testing Philosophy

The project maintains high test coverage (>90% for lexer, >80% for parser) and includes the complete DWScript test suite (~2,100 tests in `testdata/fixtures/`). When adding features:

1. Write tests FIRST (TDD approach as per PLAN.md)
2. Test both success and error cases
3. Include edge cases (empty input, malformed syntax, boundary values)
4. Use table-driven tests for multiple similar cases
5. Run fixture tests to verify compatibility with original DWScript
6. Update `testdata/fixtures/TEST_STATUS.md` as tests pass

**Running Fixture Tests**:
```bash
# Run all fixture tests
go test -v ./internal/interp -run TestDWScriptFixtures

# Run specific category
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts

# See testdata/fixtures/README.md for more options
```

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
9. **Stage 9**: Deferred Stage 8 Tasks
10. **Stage 10**: Performance and polish (68 tasks)
11. **Stage 11**: Long-term evolution (54 tasks)
12. **Stage 12**: Codegen

When implementing new stages:

- Create AST nodes first
- Extend parser with new syntax
- Add comprehensive tests
- Update CLI if applicable
- Document in stage summary files under `docs/`
- Mark tasks as done in `PLAN.md`

## Important Files

- `PLAN.md`: Complete task breakdown and progress tracking
- `goal.md`: Detailed strategy and rationale for the port
- `README.md`: User-facing documentation
- `CONTRIBUTING.md`: Contribution guidelines
- `docs/stage*.md`: Stage completion summaries with statistics

## Reference Material

- Original DWScript source: `reference/dwscript-original/` (for reference only)
- DWScript language reference: https://www.delphitools.info/dwscript/
- Test scripts: `testdata/*.dws` (custom test scripts)
- Comprehensive test suite: `testdata/fixtures/` (~2,100 tests from original DWScript)
  - See `testdata/fixtures/README.md` for test structure and usage
  - See `testdata/fixtures/TEST_STATUS.md` for current pass/fail status

## Notes for Future Development

**Simulating Delphi Features in Go**: Since Go lacks classes, inheritance, and method overloading:

- Use ClassInfo structs for class metadata
- ObjectInstance structs with field maps for instances
- Method tables for dynamic dispatch and overriding
- Type checking at semantic analysis phase (Stage 6)

**Memory Management**: Go's GC handles cleanup; no manual reference counting needed like Delphi.

**Performance**: Current AST interpreter is simple.
