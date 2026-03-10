# Interp/Evaluator Steady State

This note describes the intended runtime architecture after Phase 4 and 4.8.

## Package roles

- `internal/interp`
  - engine/bootstrap package
  - owns construction, engine-facing API, unit integration, and runtime orchestration
- `internal/interp/evaluator`
  - AST execution engine
  - owns expression, statement, declaration, call, property, and dispatch semantics
- `internal/interp/runtime`
  - execution primitives
  - owns values, environments, call stack, execution context, and runtime metadata
- `internal/interp/types`
  - interpreter-local registries and lookup tables
  - owns class, record, interface, helper, and function registration
- `internal/interp/contracts`
  - narrow neutral coordination layer
  - exists only for cross-package types that do not belong in `runtime`

## Ownership rules

- `ExecutionContext` is the canonical owner of per-run mutable state.
- Production bootstrap is centralized in `internal/interp/new.go`.
- `internal/interp` must not import `internal/interp/evaluator` outside construction and tests.
- Runtime execution should not bounce through callback-style interpreter bridges.
- Runtime metadata should prefer typed runtime structures over AST-shaped compatibility maps where possible.

## Runtime flow

1. `internal/interp/new.go` constructs environment, type system, evaluator, and interpreter facade.
2. The public engine enters through `internal/interp`.
3. AST execution runs through `internal/interp/evaluator`.
4. Evaluator and interpreter share `runtime` primitives and `types` registries.

## What is intentionally gone

- callback-style execution interfaces such as `CoreEvaluator`, `OOPEngine`, and `DeclHandler`
- bridge-only wiring such as `SetFocusedInterfaces()`
- interpreter-side fallback dispatch entry points like `EvalNode()` / `evalLegacy()`

## What Phase 5 should assume

Phase 5 and later work should treat this split as the baseline:

- `interp` is the engine shell
- `evaluator` is the execution core
- `runtime` is the primitive layer

Further cleanup should tighten that boundary, not recreate callback ownership.
