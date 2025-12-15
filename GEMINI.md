# GEMINI.md

This file provides context and instructions for Gemini when working on the `go-dws` project.

## Project Overview

**go-dws** is a faithful port of [DWScript](https://github.com/EricGrange/DWScript) (Delphi Web Script) from Delphi to Go. It aims for 100% syntax and semantics compatibility while leveraging Go's modern ecosystem.

*   **Status:** Work in Progress (WIP).
*   **Core Tech:** Go (Golang), WebAssembly (WASM).
*   **Goal:** Create an embeddable scripting engine that runs natively in Go applications and in browsers via WASM.

## Architecture

The compiler/interpreter follows a traditional pipeline:

`Source Code` → `Lexer` → `Parser` → `AST` → `Semantic Analyzer` → `Interpreter` / `Bytecode VM`

### Directory Structure

*   **`cmd/`**: Entry points.
    *   `dwscript/`: Main CLI tool (`lex`, `parse`, `run`, `compile`).
    *   `gen-visitor/`: Code generator for the AST visitor pattern.
*   **`internal/`**: Private implementation.
    *   `lexer/`: Tokenizer (handles case-insensitivity, hex/binary numbers).
    *   `parser/`: Pratt parser implementation.
    *   `ast/`: AST node definitions.
    *   `interp/`: AST-based Interpreter (reference implementation).
    *   `bytecode/`: Bytecode VM (performance optimized, 5-6x faster).
    *   `semantic/`: Type checking and symbol resolution.
*   **`pkg/`**: Public API.
    *   `dwscript/`: High-level embedding API.
    *   `ast/`: Public AST types.
    *   `ident/`: Utilities for case-insensitive identifiers.
    *   `platform/`: Abstraction layer for Native/WASM support.
*   **`testdata/`**: comprehensive test suite, including fixtures ported from original DWScript.

## Building and Running

The project uses `just` as a command runner, but standard `go` commands work as well.

### Common Commands

| Task | Just Command | Go Command |
| :--- | :--- | :--- |
| **Build CLI** | `just build` | `go build -o bin/dwscript ./cmd/dwscript` |
| **Run Tests** | `just test` | `go test ./...` |
| **Test Coverage** | `just test-coverage` | `go test -coverprofile=coverage.out ./...` |
| **Lint** | `just lint` | `golangci-lint run` |
| **Format** | `just fmt` | `go fmt ./... && goimports -w .` |
| **WASM Build** | `just wasm` | *(See `build/wasm/build.sh`)* |

### CLI Usage

*   **Run a script:** `./bin/dwscript run <file.dws>`
*   **Run with Bytecode VM:** `./bin/dwscript run --bytecode <file.dws>`
*   **Parse & Print AST:** `./bin/dwscript parse <file.dws>`
*   **Tokenize:** `./bin/dwscript lex <file.dws>`

## Development Conventions

### 1. Coding Style
*   Follow standard Go idioms (`Effective Go`).
*   Ensure code is formatted with `gofmt` and `goimports`.
*   Run linters (`golangci-lint`) before committing.

### 2. Language Specifics (Important)
*   **Case Insensitivity:** DWScript is case-insensitive.
    *   **NEVER** use `strings.ToLower` or `strings.EqualFold` directly.
    *   **ALWAYS** use `pkg/ident` utilities:
        *   `ident.Equal(a, b)` for comparison.
        *   `ident.Normalize(s)` for map keys.
*   **String Encoding:** The project uses **UTF-8** internally, diverging from original DWScript's UTF-16. Be aware of this when handling string length or indexing in tests.

### 3. Parser & AST
*   **Pratt Parser:** The parser uses a Pratt parsing approach for expressions.
*   **Documentation:** Parse functions must document `PRE` (current token state) and `POST` (state after parsing) conditions.
*   **Visitor Pattern:** AST traversal uses a generated visitor pattern.
    *   Do not manually edit `pkg/ast/visitor_generated.go`.
    *   Regenerate using `go run cmd/gen-visitor/main.go` after changing AST nodes.

### 4. Testing
*   **TDD:** Write tests *before* implementing features (adhering to `PLAN.md`).
*   **Fixtures:** The project maintains a massive suite of compatibility tests in `testdata/fixtures`.
    *   Run specific fixtures: `go test -v ./internal/interp -run TestDWScriptFixtures/Category`
*   **Coverage:** Maintain high test coverage (>90%). Use table-driven tests for parser/lexer.

## Key Files for Context
*   `PLAN.md`: The central source of truth for the project roadmap and current tasks.
*   `CLAUDE.md`: Contains detailed architectural notes and CLI examples.
*   `CONTRIBUTING.md`: Detailed contribution guidelines and parser-specific rules.
*   `justfile`: List of available build/maintenance recipes.
