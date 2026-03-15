# Interpreter/Evaluator boundary (current state)

This file describes the package boundary rule. For the steady-state runtime
ownership model after Phase 4, see
[`interp-evaluator-steady-state.md`](./interp-evaluator-steady-state.md).

This repo intentionally enforces a *mostly one-way dependency* rule:

- `internal/interp` must **not** import `internal/interp/evaluator` in runtime logic.
- The only production exception is canonical runtime construction in [`internal/interp/new.go`](../../internal/interp/new.go).

Phase 4 removed the old callback-style focused interfaces, but the repo still
needs a clean package boundary so `interp` does not regress into importing
`evaluator` throughout runtime logic.

## Why `internal/interp/contracts` still exists

`internal/interp/contracts` is now a small *neutral* package for the remaining
cross-package coordination types that do not belong in either `interp` or
`runtime`:

- shared engine state
- neutral runtime-facing value/meta interfaces such as `ClassMetaValue`
- user-function callback types
- the external-function dispatch hook carried in `EngineState`

### Dependency directions

Allowed imports:

- `internal/interp` → `internal/interp/contracts`, `internal/interp/runtime`
- `internal/interp/new.go` → `internal/interp/evaluator` (construction only)
- `internal/interp/evaluator` → `internal/interp/contracts`, `internal/interp/runtime`
- `internal/interp/runner` → `internal/interp` (delegation only)

Not allowed:

- `internal/interp` → `internal/interp/evaluator`

## Why not put these interfaces into `runtime`?

`runtime` is intended to hold shared *execution primitives* (values, env,
execution context, helpers). `contracts` is narrower: it holds coordination
types that would otherwise force `interp` and `evaluator` to import each other
or would make `runtime` a catch-all package.

Keeping these in `contracts` prevents `runtime` from turning into a “catch-all” package and makes cross-layer coupling explicit.

## Where construction/wiring happens

Construction is centralized behind a single bootstrap entry point:

- `internal/interp/new.go` creates:
  - `runtime.Environment`
  - `typesystem`
  - `evaluator.Evaluator`
  - and then calls `interp.NewWithDeps(...)`

`internal/interp/runner` now delegates to that constructor. Runtime logic
outside the bootstrap path should remain evaluator-free.

### Tests

Some `internal/interp` package tests rely on `interp.New(...)`. That constructor is now
the same production bootstrap path used by `runner`, so tests and production enter the
runtime through the same engine-construction API.

## Practical rule of thumb

When you feel tempted to add `import ".../internal/interp/evaluator"` inside `internal/interp`:

1. Ask: “Is this a *runtime primitive*?” → move/alias into `runtime`.
2. Ask: “Is this a *cross-package coordination type*?” → consider `contracts`.
3. If it is construction-only glue, confine it to `internal/interp/new.go`.
4. Otherwise, keep it out of `internal/interp`.

## How to verify the boundary

Quick checks:

- `go test ./...`
- `grep -R "internal/interp/evaluator" internal/interp/*.go internal/interp/**/*.go` should only match `new.go`, evaluator files, runner files, and test-only files.
