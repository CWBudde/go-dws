# AI Coding Agent Instructions for go-dws

## Project Overview
go-dws is a Go port of DWScript (Delphi Web Script), implementing a Pascal-based scripting language with 100% syntax/semantics compatibility. The project follows a 10-stage incremental plan, currently at Stage 3 (statements and execution).

**Architecture Pipeline**: `Source Code → Lexer → Parser → AST → Semantic Analyzer → Interpreter`

## Core Package Structure
- **`lexer/`**: Tokenization with 150+ DWScript tokens, case-insensitive keywords, hex/binary literals
- **`parser/`**: Pratt parser with precedence climbing (LOWEST to MEMBER levels)
- **`ast/`**: Abstract Syntax Tree with Node/Expression/Statement interfaces, visitor pattern ready
- **`types/`**: Type system scaffolding (Integer, Float, String, Boolean, Array, Class)
- **`interp/`**: Interpreter/runtime engine (placeholder for Stage 3+)
- **`cmd/dwscript/`**: Cobra-based CLI (`lex`, `parse`, `run`, `version` commands)

## Critical DWScript Language Specifics
**Case Insensitivity**: Keywords normalized lowercase via `LookupIdent()`. Test both cases.

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
./bin/dwscript lex testdata/simple.dws
./bin/dwscript lex -e "var x: Integer := 42;"

# Parse and show AST
./bin/dwscript parse testdata/simple.dws
./bin/dwscript parse -e "3 + 5 * 2"

# Run script (Stage 3+)
./bin/dwscript run script.dws
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

## Stage 3+ Considerations
**Statements**: Extend AST with variable declarations, assignments, blocks. Add statement sequencing.

**Interpreter**: Introduce environment/symbol tables, execute parsed statements end-to-end.

**Scope**: Implement nested function scopes, variable resolution (Stage 5).

**Classes**: Use `ClassInfo` structs for metadata, `ObjectInstance` with field maps, method tables for dispatch (Stage 7).

Always consult `PLAN.md` for current stage requirements and `docs/stage*.md` for completed work patterns.