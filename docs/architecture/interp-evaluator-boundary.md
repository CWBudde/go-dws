# Interpreter/Evaluator boundary (why `contracts` exists)

This repo intentionally enforces a *one-way dependency* rule:

- `internal/interp` must **not** import `internal/interp/evaluator`.

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
- `internal/interp/evaluator` → `internal/interp/contracts`, `internal/interp/runtime`
- `internal/interp/runner` → `internal/interp` **and** `internal/interp/evaluator` (wiring only)

Not allowed:

- `internal/interp` → `internal/interp/evaluator`

## Why not put these interfaces into `runtime`?

`runtime` is intended to hold shared *execution primitives* (values, env, execution context, helpers). The boundary interfaces are *coordination/policy*:

- they define what the evaluator is allowed to ask the interpreter to do
- they define the minimal API the interpreter needs from the evaluator

Keeping these in `contracts` prevents `runtime` from turning into a “catch-all” package and makes cross-layer coupling explicit.

## Where construction/wiring happens

Construction is centralized in a wiring package so that `internal/interp` stays evaluator-free:

- `internal/interp/runner` creates:
  - `runtime.Environment`
  - `typesystem`
  - `evaluator.Evaluator`
  - and then calls `interp.NewWithDeps(...)`

This is the only place that should import both interpreter and evaluator in production code.

### Tests

Some `internal/interp` package tests rely on `interp.New(...)`. To preserve that convenience without breaking the layering rule:

- a test-only constructor may live in `internal/interp/*_test.go` and import evaluator
- production code should use `internal/interp/runner` (or `pkg/dwscript`)

## Practical rule of thumb

When you feel tempted to add `import ".../internal/interp/evaluator"` inside `internal/interp`:

1. Ask: “Is this a *runtime primitive*?” → move/alias into `runtime`.
2. Ask: “Is this a *cross-layer capability*?” → express it as a narrow interface in `contracts`.
3. Wire it in `runner`.

## How to verify the boundary

Quick checks:

- `go test ./...`
- `grep -R "internal/interp/evaluator" internal/interp/*.go internal/interp/**/*.go` should only match wiring/test-only files.
