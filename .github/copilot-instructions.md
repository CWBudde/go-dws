# AI Coding Agent Instructions for go-dws

## Project Overview
go-dws is a Go port of DWScript (Delphi Web Script), implementing a Pascal-based scripting language with 100% syntax/semantics compatibility. The project follows a 10-stage incremental plan, currently at Stage 3 (statements and execution).

**Architecture Pipeline**: `Source Code → Lexer → Parser → AST → Semantic Analyzer → Bytecode Compiler → VM / Interpreter`

## Core Package Structure

### Internal Packages (`internal/`)
- **`lexer/`**: Tokenization with 150+ DWScript tokens, case-insensitive keywords, hex/binary literals
- **`parser/`**: Pratt parser with precedence climbing (LOWEST to MEMBER levels)
- **`ast/`**: Abstract Syntax Tree with Node/Expression/Statement interfaces, visitor pattern ready
- **`semantic/`**: Type checking, symbol resolution, operator validation, interface/class analysis
- **`types/`**: Rich type system (Integer, Float, String, Boolean, Array, Class, Interface, Enum, Record, Set)
- **`bytecode/`**: Bytecode compiler and stack-based VM (5-6x faster than AST interpreter)
- **`interp/`**: Tree-walking AST interpreter/runtime engine
- **`errors/`**: Error reporting with position tracking
- **`jsonvalue/`**: JSON type integration
- **`units/`**: Unit/module system

### Public Packages (`pkg/`)
- **`ast/`**: Public AST interfaces for external tools
- **`dwscript/`**: Main entry point for embedding go-dws
- **`token/`**: Public token definitions
- **`platform/`**: Platform-specific helpers
- **`wasm/`**: WebAssembly bindings

### Commands & Tools
- **`cmd/dwscript/`**: Cobra-based CLI (`lex`, `parse`, `run`, `bytecode`, `version` commands)
- **`cmd/dwscript-wasm/`**: WebAssembly build target

## Critical DWScript Language Specifics

**Case Insensitivity**: Keywords normalized lowercase via `LookupIdent()`. Test both cases (`PrintLn`, `println`). The entire language is case-insensitive.

**Operators**:
- Assignment: `:=` (not `=`)
- Equality: `=` (not `==`)
- Inequality: `<>`
- Integer division: `div` (keyword)
- Modulo: `mod` (keyword)
- Boolean: `and`, `or`, `xor`, `not` (keywords)
- Compound: `+=`, `-=`, `*=`, `/=`

**Literals**:
- Strings: Single/double quotes, doubled quotes escape (`'it''s'` → `it's`)
- Numbers: `42`, `$FF` (hex), `%1010` (binary), `3.14`, `1.0e10`
- Comments: `{}` or `(* *)` blocks, `//` lines

**Control Flow**:
- Loops: `for i := 1 to 10 do` / `for i := 10 downto 1 do`
- Case: Multi-value cases (`case x of 1,2,3: ...`), string cases, `else` clause
- Begin/End: Block delimiters (like curly braces in C-style languages)
- If/Then/Else: No parentheses required around conditions

**Type Declarations**:
- Type keyword: `type TMyClass = class ... end;`
- Array syntax: `array[1..5] of Integer` or `array of String` (dynamic)
- Class sections: `private`, `protected`, `public`, `published`
- Constructors: `constructor Create(...); begin ... end;`
- Methods inline: Function/procedure bodies can appear directly in class declaration

**Variable Declarations**:
- Explicit: `var x: Integer;`
- Inferred: `var x := 42;`
- Multiple: `var a, b, c: Integer;`

**Functions & Procedures**:
- Functions return via `Result := value;`
- Procedures are void (no return value)
- Parameters: `function Add(a, b: Integer): Integer;`
- Forward declarations supported

**Built-in Functions** (from test fixtures):
- I/O: `PrintLn()`, `Print()`, `ReadLn()`
- Conversions: `IntToStr()`, `FloatToStr()`, `StrToInt()`, `StrToFloat()`
- String: `Length()`, `Copy()`, `Pos()`, `UpperCase()`, `LowerCase()`
- Math: `Abs()`, `Sqr()`, `Sqrt()`, `Sin()`, `Cos()`, `Round()`, `Trunc()`

## Development Workflow Commands
```bash
# Build CLI
go build ./cmd/dwscript

# Run all tests (table-driven, high coverage required)
go test ./...

# Coverage analysis
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Formatting (required before commit)
go fmt ./...

# Dependency management
go mod tidy
```

## CLI Usage Patterns
```bash
# Tokenize file or expression
./bin/dwscript lex testdata/arithmetic.dws
./bin/dwscript lex -e "var x: Integer := 42;"

# Parse and show AST
./bin/dwscript parse testdata/for_demo.dws
./bin/dwscript parse -e "3 + 5 * 2"

# Run script (AST interpreter)
./bin/dwscript run testdata/case_demo.dws

# Compile to bytecode and execute (5-6x faster)
./bin/dwscript bytecode testdata/classes.dws
```

## Implementation Patterns
**Token Types**: Use `TokenKind` enum pattern from `lexer/token_type.go`, extend with `NewToken` constructor.

**Parser Extensions**: Register prefix/infix parse functions for new syntax, update precedence map.

**AST Nodes**: Implement `Node` interface with `TokenLiteral()` and `String()` methods. Use visitor pattern for traversal.

**Error Handling**: Accumulate errors with position info, don't stop at first error. Use `addError()`/`peekError()`.

**Testing**: Table-driven tests mirroring DWScript behavior. Load fixtures from `testdata/` with `os.ReadFile`. Maintain >90% coverage for core components.

## Code Style Conventions
- **Naming**: `UpperCamelCase` exported, `lowerCamelCase` internal, `TestName_Subject` for tests
- **Imports**: Grouped with `goimports` conventions
- **Packages**: Keep helpers in same directory (avoid circular imports)
- **Documentation**: Comprehensive GoDoc comments on all exported functions/types

## Commit & Planning Integration
- **Conventional Commits**: `feat(parser): ...`, `fix: ...`, `docs: ...`
- **Task Tracking**: Include stage/task IDs from `PLAN.md` in commits
- **Planning**: Update `PLAN.md` before structural changes, sync with `docs/` and `goal.md`
- **Testing**: Mirror DWScript's original test suite, reference upstream cases in comments

## Key Reference Files
- **`PLAN.md`**: Complete 511-task roadmap with current progress
- **`AGENTS.md`**: Repository guidelines (build, test, style, testing, commits)
- **`CLAUDE.md`**: Detailed architecture and development guide
- **`testdata/*.dws`**: DWScript test scripts for regression testing
- **`reference/dwscript-original/`**: Upstream DWScript source (read-only)

Always consult `PLAN.md` for current stage requirements before implementing features or making architectural changes.