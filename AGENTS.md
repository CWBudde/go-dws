# Repository Guidelines

## Project Structure & Module Organization

Core compiler code lives in focused packages: `lexer`, `parser`, `ast`, `types`, and `interp`, each mirroring DWScript components. The CLI entry point is in `cmd/dwscript`, currently a scaffold that will orchestrate the compiler pipeline. Shared fixtures and expected-output samples reside under `testdata`, while `reference` holds the upstream DWScript sources for parity checks—do not modify them. Architecture notes and stage plans are tracked in `docs/`, `PLAN.md`, and `goal.md`; consult them before proposing structural changes.

## Build, Test, and Development Commands

- `go mod tidy` synchronizes `go.mod`/`go.sum` after dependency updates.
- `go build ./cmd/dwscript` builds the CLI; use `go run ./cmd/dwscript --help` to validate flags.
- `go test ./...` runs the full unit test suite; add `-run Name` to target specific cases.
- `go test -coverprofile=coverage.out ./...` refreshes project-wide coverage data.

## Reference Material

- Original DWScript source: `reference/dwscript-original/`
- DWScript language reference: https://www.delphitools.info/dwscript/
- Test scripts: `testdata/*.dws`

## Coding Style & Naming Conventions

Follow idiomatic Go style: tabs for indentation, `UpperCamelCase` for exported symbols, `lowerCamelCase` for internals, and `TestName_Subject` for test functions. Run `gofmt` (or `go fmt ./...`) before committing; import blocks should be grouped with `goimports` conventions. Lexer and parser tokens use the `TokenKind` and `NewToken` patterns established in `lexer/token.go`—mirror existing naming when extending enums or node types. Keep package-specific helper files in the same directory to avoid circular imports.

## Testing Guidelines

Use Go’s `testing` package and favor table-driven tests to mirror DWScript scenarios. Place fixture scripts in `testdata/<feature>` and load them with `os.ReadFile` to keep tests data-driven. Name test files `*_test.go` and keep coverage above the current baseline when touching lexer or parser logic by updating `coverage.out`. When reproducing upstream behavior, add assertions that reference the original DWScript test case in comments.

## Commit & Pull Request Guidelines

Commits follow a Conventional Commit style seen in history (`feat(parser): ...`, `fix: ...`, `docs: ...`). Keep them focused, include relevant stage or task identifiers from `PLAN.md` when applicable, and document behavioral changes in the message body. PRs should summarize the change set, list manual or automated tests (copy the `go test` command you ran), link any tracked issues or plan tasks, and attach CLI output or screenshots if the user-facing behavior changes.

## Documentation & Planning

Before large refactors, update `PLAN.md` with the affected tasks and ensure the milestone status still reflects reality. Revise `docs/` or `README.md` when interfaces, flags, or directory layout shift. Small tooling tips belong in the `docs/` directory, while high-level strategy updates go to `goal.md`; keep each artifact synchronized.
Once a feature phase is complete, mark tasks as done in `PLAN.md`.
