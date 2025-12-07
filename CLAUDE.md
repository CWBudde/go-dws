# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-dws is a Go port of DWScript (Delphi Web Script), a full-featured Object Pascal-based scripting language. The project aims for 100% language compatibility with the original DWScript while using idiomatic Go patterns.

## Common Commands

### Building

```bash
# Using just (recommended)
just build             # Build the CLI tool to bin/dwscript
just tidy              # Tidy dependencies
just clean             # Clean build artifacts

# Development workflows
just dev               # Format, lint, test, and build
just ci                # Run CI checks (lint + test with coverage)
just setup             # Full setup (tidy + install tools)

# Direct commands
go build ./cmd/dwscript  # Build the CLI tool
go install ./cmd/dwscript # Build and install globally
```

### Testing

```bash
# Using just (recommended)
just test              # Run all tests
just test-verbose      # Run tests with verbose output
just test-coverage     # Run tests with coverage and generate HTML report
just test-unit         # Run tests with race detection (CI)

# Direct commands
go test ./...          # Run all tests
go test -v ./...       # Run tests with verbose output
go test ./lexer        # Run tests for a specific package
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
# Using just (recommended)
just lint              # Run golangci-lint
just lint-fix          # Fix linting issues automatically
just fmt               # Format code with go fmt and goimports
just check-fmt         # Check if code is formatted (CI)

# Direct commands
golangci-lint run      # Run golangci-lint (project uses .golangci.yml config)
golangci-lint run --fix # Fix issues automatically
go vet ./...           # Run standard Go vet
go fmt ./...           # Run standard Go formatter
```

### CLI Usage

```bash
# Using just (convenient shortcuts)
just lex testdata/hello.dws    # Tokenize a file
just parse testdata/hello.dws  # Parse and display AST
just run testdata/hello.dws    # Run a script

# Direct CLI commands
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

# Compile to bytecode (.dwc file)
./bin/dwscript compile script.dws

# Compile with custom output file
./bin/dwscript compile script.dws -o output.dwc

# Run precompiled bytecode
./bin/dwscript run script.dwc

# Run with custom recursion limit (default: 1024)
./bin/dwscript run --max-recursion 2048 script.dws

# Show version
./bin/dwscript version
```

## Architecture Overview

### Pipeline

```plain
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
  - `run` command: Execute scripts (source or compiled bytecode)
  - `compile` command: Compile source to bytecode (.dwc files)
  - `version` command: Show version info

- `cmd/gen-visitor/` - Code generator for AST visitor pattern
  - Parses AST node definitions from `pkg/ast/*.go`
  - Generates `pkg/ast/visitor_generated.go` with type-safe walk functions
  - Features: type-aware field handling, embedded field support, custom traversal order
  - Run: `go run cmd/gen-visitor/main.go` or `go generate ./pkg/ast`

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

- `internal/interp/` - AST Interpreter/runtime (Phase 3.5 refactored architecture)
  - **Architecture**: Split into thin orchestrator (Interpreter) + evaluation engine (Evaluator)
  - **Interpreter** (`interpreter.go`): Thin orchestrator that manages global state and delegates evaluation
    - Maintains compatibility with existing API
    - Provides adapter interface for gradual migration
    - Manages global registries (functions, classes, records, operators)
  - **Evaluator** (`evaluator/`): Visitor-pattern based evaluation engine
    - `evaluator.go`: Core evaluator with visitor pattern infrastructure
    - `visitor_expressions.go`: Expression evaluation (48+ visitor methods)
    - `visitor_statements.go`: Statement evaluation (control flow, loops, etc.)
    - `visitor_declarations.go`: Declaration evaluation (functions, classes, types)
    - `visitor_literals.go`: Literal value creation
    - `binary_ops.go`: Binary operation implementations (arithmetic, comparison, boolean, string)
    - `context.go`: ExecutionContext for managing environment and call stack
    - `callstack.go`: Call stack management with recursion depth tracking
    - Performance: ~70 ns/op for binary operations, 0 allocations for literals (see `docs/evaluator-performance-report.md`)
  - **Type System** (`types/`): Centralized type registry
    - `type_system.go`: Manages classes, records, interfaces, functions, helpers
    - `function_registry.go`: Function overload resolution and registration
  - **Runtime** (`runtime/`): Runtime value types
    - Integer, Float, String, Boolean, Array, Record, Class instance values
    - Environment/symbol table management
  - Built-in function implementations

- `internal/bytecode/` - Bytecode VM (5-6x faster than AST interpreter)
  - `compiler.go`: AST-to-bytecode compiler with optimizations
  - `vm.go`: Stack-based virtual machine with built-in function support
  - `bytecode.go`: Bytecode format, constant pools, value types
  - `disasm.go`: Bytecode disassembler for debugging
  - `instruction.go`: 116 opcodes for DWScript operations
  - `serializer.go`: Bytecode serialization/deserialization (.dwc file format)
  - See [docs/bytecode-vm.md](docs/bytecode-vm.md) for details

- `internal/errors/` - Error handling utilities
  - Error formatting and reporting

**pkg/** - Public APIs (importable by external projects)

- `pkg/dwscript/` - High-level embedding API
  - `dwscript.go`: Engine, Program, Result types
  - `options.go`: Configuration options
  - Public API for embedding DWScript in Go applications
  - See [README.md](README.md) for usage examples

- `pkg/ast/` - Public AST types and utilities
  - Full AST node definitions (expressions, statements, declarations)
  - Visitor pattern implementation for tree traversal
  - Position tracking and type information support

- `pkg/token/` - Public token types
  - Token types, Position, and TokenType definitions
  - 150+ DWScript token constants
  - Case-insensitive keyword lookup utilities

- `pkg/ident/` - Case-insensitive identifier utilities
  - `Normalize()`: Canonical form for map keys
  - `Equal()`: Case-insensitive comparison (use instead of `strings.EqualFold`)
  - `HasPrefix()`, `HasSuffix()`: Case-insensitive prefix/suffix matching
  - `Contains()`, `Index()`: Case-insensitive slice operations
  - See "Case Insensitivity" in Development Guidelines for usage

- `pkg/platform/` - Platform abstraction layer (planned for Stage 10.15)
  - Abstracts filesystem, console, and platform-specific functionality
  - Enables native and WebAssembly builds with consistent behavior

- `pkg/wasm/` - WebAssembly support (planned for Stage 10.15)
  - JavaScript/Go interop
  - Browser API bindings

### Key Design Patterns

**Pratt Parser**: Used for expression parsing with operator precedence. Register prefix/infix parse functions for each token type that can start or continue an expression.

**AST Visitor Pattern**: The project uses a generated visitor pattern for tree traversal. The visitor code is automatically generated from AST node definitions, eliminating 83.6% of manually-written boilerplate code while maintaining zero runtime overhead.

**Visitor Code Generation**: The visitor generator (`cmd/gen-visitor/main.go`) parses all AST node definitions and generates `pkg/ast/visitor_generated.go`. Key features:

- **Type-aware field handling**: Only walks fields that are Node types (Expression, Statement, etc.). Primitive fields (strings, ints, bools) are automatically skipped. For example, `BinaryExpression.Operator` (string) is skipped, but `AddressOfExpression.Operator` (Expression) is walked.

- **Embedded field support**: Recursively extracts fields from embedded base types like `TypedExpressionBase`. All expressions that embed `TypedExpressionBase` automatically get their `Type *TypeAnnotation` field walked.

- **Helper type support**: Non-Node helper types like `Parameter`, `CaseBranch`, `ExceptClause` get their own walk functions that are called by the main walker.

- **Custom traversal order**: Support for `ast:"order:N"` tags to control the order of field traversal.

To regenerate the visitor after modifying AST nodes:
```bash
go run cmd/gen-visitor/main.go
# or
go generate ./pkg/ast
```

**Evaluator Visitor Pattern** (Phase 3.5): The interpreter's evaluation logic uses a visitor pattern for cleaner code organization and better maintainability.

- **Visitor Methods**: Each AST node type has a corresponding `Visit*` method in the Evaluator
  - Example: `VisitBinaryExpression(node *ast.BinaryExpression, ctx *ExecutionContext) Value`
  - All visitor methods take a node and an ExecutionContext, return a Value

- **ExecutionContext**: Encapsulates evaluation state
  - Environment (variable bindings)
  - Call stack (for recursion tracking)
  - Control flow state (break, continue, return, exit)
  - Exception handling state

- **Adapter Pattern**: Temporary bridge between Interpreter and Evaluator during migration
  - `InterpreterAdapter` interface allows Evaluator to call back to Interpreter
  - Will be removed in Phase 3.5.37 (blocked on AST-free runtime types)

- **Performance**: Zero overhead vs switch-based approach
  - Literals: 0.3-0.6 ns/op (compiler-optimized, 0 allocations)
  - Binary operations: ~70 ns/op, 3 allocations
  - See `docs/evaluator-performance-report.md` for detailed benchmarks

**Symbol Tables**: Will use nested scope chain for variable/function resolution (Stage 5).

## Development Guidelines

### DWScript Language Specifics

**Case Insensitivity**: DWScript is a case-insensitive language. All identifiers, keywords, type names, and member names are compared without regard to case.

Use the `pkg/ident` package for all case-insensitive identifier operations:

```go
import "github.com/cwbudde/go-dws/pkg/ident"

// Comparing identifiers (preferred over strings.EqualFold)
if ident.Equal(name1, name2) { ... }

// Normalizing for map keys
variables[ident.Normalize("MyVar")] = value

// Case-insensitive prefix/suffix matching
if ident.HasPrefix(typeName, "array") { ... }
if ident.HasSuffix(typeName, "Type") { ... }

// Slice operations
if ident.Contains(keywords, name) { ... }
idx := ident.Index(items, name)
```

**Important**: Avoid direct `strings.ToLower()` or `strings.EqualFold()` calls - use the `ident` helpers instead for consistency.

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

### Experimental Multi-Pass Semantic Analysis (Task 6.1.2)

The semantic analyzer has a new multi-pass architecture under development. By default, only the stable old analyzer runs to keep tests passing on main.

**To enable experimental passes** (for task 6.1.2 development):

```go
// Use this constructor to enable Pass 2 (Type Resolution) and Pass 3 (Semantic Validation)
analyzer := semantic.NewAnalyzerWithExperimentalPasses()
err := analyzer.Analyze(program)
```

**Default behavior** (NewAnalyzer()):

- Only the old analyzer runs
- All existing tests pass
- Stable for production use

**Experimental behavior** (NewAnalyzerWithExperimentalPasses()):

- Old analyzer runs first (declaration collection)
- Pass 2: Type Resolution (resolve type references, build hierarchies)
- Pass 3: Semantic Validation (type-check expressions, validate statements)
- Some tests may fail due to dual-mode validation conflicts

When working on task 6.1.2, create test files that explicitly use `NewAnalyzerWithExperimentalPasses()` to test the new pass system without affecting other tests.

### Testing Philosophy

The project maintains high test coverage (>90% for lexer, >80% for parser) and includes the complete DWScript test suite (~2,100 tests in `testdata/fixtures/`). When adding features:

1. Write tests FIRST (TDD approach as per PLAN.md)
2. Test both success and error cases
3. Include edge cases (empty input, malformed syntax, boundary values)
4. Use table-driven tests for multiple similar cases
5. Run fixture tests to verify compatibility with original DWScript
6. Update `testdata/fixtures/TEST_STATUS.md` as tests pass

NOTE: While we want to maintain compatibility with original DWScript, this go-dws port may intentionally diverge in some areas. Especially in terms of error reporting. We want to see at least the amount of errors the original produces, but we may produce more detailed or additional errors in some cases. In this case, and ONLY in this case, it is acceptable to modify the expected error output files in `testdata/fixtures/expected_errors/` to match the new behavior.

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

When implementing new stages/tasks:

- Create AST nodes first
- Extend parser with new syntax
- Consider semantic analysis needs
- Implement runtime support (interpreter & bytecode VM)
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
- DWScript language reference: <https://www.delphitools.info/dwscript/>
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
