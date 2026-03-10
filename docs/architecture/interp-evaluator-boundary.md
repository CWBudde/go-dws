# Interpreter/Evaluator boundary (why `contracts` exists)

This repo intentionally enforces a *mostly one-way dependency* rule:

- `internal/interp` must **not** import `internal/interp/evaluator` in runtime logic.
- The only production exception is canonical runtime construction in [`internal/interp/new.go`](../../internal/interp/new.go).

At the same time, the evaluator still needs to invoke interpreter-owned behavior (OO dispatch, declaration side-effects, exception construction, etc.). In Go, doing that naively tends to create either:

- an import cycle (`interp` ↔ `evaluator`), or
- a layering regression (interpreter imports evaluator “just for a few types”).

## The solution: dependency inversion via `internal/interp/contracts`

`internal/interp/contracts` is a small *neutral* package containing the interfaces and callback types used across the boundary.

Conceptually:

- The evaluator depends on *interfaces* (what it needs done).
- The interpreter implements those interfaces (how it’s done).
- A wiring package connects the concrete interpreter to the concrete evaluator.

### Dependency directions

Allowed imports:

- `internal/interp` → `internal/interp/contracts`, `internal/interp/runtime`
- `internal/interp/new.go` → `internal/interp/evaluator` (construction only)
- `internal/interp/evaluator` → `internal/interp/contracts`, `internal/interp/runtime`
- `internal/interp/runner` → `internal/interp` (delegation only)

Not allowed:

- `internal/interp` → `internal/interp/evaluator`

## Why not put these interfaces into `runtime`?

`runtime` is intended to hold shared *execution primitives* (values, env, execution context, helpers). The boundary interfaces are *coordination/policy*:

- they define what the evaluator is allowed to ask the interpreter to do
- they define the minimal API the interpreter needs from the evaluator

Keeping these in `contracts` prevents `runtime` from turning into a “catch-all” package and makes cross-layer coupling explicit.

## Where construction/wiring happens

Construction is centralized behind a single bootstrap entry point:

- `internal/interp/new.go` creates:
  - `runtime.Environment`
  - `typesystem`
  - `evaluator.Evaluator`
  - and then calls `interp.NewWithDeps(...)`

`internal/interp/runner` now delegates to that constructor. Runtime logic outside the
bootstrap path should still remain evaluator-free.

### Tests

Some `internal/interp` package tests rely on `interp.New(...)`. That constructor is now
the same production bootstrap path used by `runner`, so tests and production enter the
runtime through the same engine-construction API.

## Practical rule of thumb

When you feel tempted to add `import ".../internal/interp/evaluator"` inside `internal/interp`:

1. Ask: “Is this a *runtime primitive*?” → move/alias into `runtime`.
2. Ask: “Is this a *cross-layer capability*?” → express it as a narrow interface in `contracts`.
3. If it is construction-only glue, confine it to `internal/interp/new.go`.
4. Otherwise, keep it out of `internal/interp`.

## How to verify the boundary

Quick checks:

- `go test ./...`
- `grep -R "internal/interp/evaluator" internal/interp/*.go internal/interp/**/*.go` should only match `new.go`, evaluator files, runner files, and test-only files.
