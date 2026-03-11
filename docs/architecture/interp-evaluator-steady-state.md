# Interp/Evaluator Steady State

This note describes the intended runtime architecture after the Phase 4 cleanup,
including the remaining seam decisions recorded in `phase-4.10.2`.

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
  - should not grow into a migration holding area

## Ownership rules

- `ExecutionContext` is the canonical owner of per-run mutable state.
- Production bootstrap is centralized in `internal/interp/new.go`.
- `internal/interp` must not import `internal/interp/evaluator` outside construction and tests.
- Runtime execution should not bounce through callback-style interpreter bridges.
- Runtime metadata should prefer typed runtime structures over AST-shaped compatibility maps where possible.
- External-function integration may remain shell-owned, but AST-shaped callback
  round-trips back into interpreter are not part of the desired steady state.
- User-function execution policy should be evaluator-native; callback bundles are
  migration residue unless explicitly justified otherwise.

## Allowed `interp` responsibilities

The shell boundary is intentionally narrow. `internal/interp` is allowed to own:

- production bootstrap and construction
- public engine-facing API/orchestration
- host and unit integration
- declaration/bootstrap mutation of runtime registries and metadata
- narrow runtime helper primitives used by those shell-owned concerns

`internal/interp` is not allowed to reintroduce production AST semantics for:

- statement/control-flow execution
- method/property/helper dispatch
- type-cast / `as` / `Default(...)` execution
- record method execution
- other evaluator-owned expression semantics

The concrete allowlist for the remaining interpreter `eval*` surface is tracked
in `phase-4.10.5-interp-allowed-responsibilities.md`.

## Runtime flow

1. `internal/interp/new.go` constructs environment, type system, evaluator, and interpreter facade.
2. The public engine enters through `internal/interp`.
3. AST execution runs through `internal/interp/evaluator`.
4. Evaluator and interpreter share `runtime` primitives and `types` registries.

## Remaining boundary decisions

- `new.go` importing evaluator is intentional bootstrap wiring.
- shared `EngineState` is intentional neutral coordination state.
- host external-function integration is an intentional shell concern.
- the remaining evaluator shim is an explicitly retained minimal internal
  implementation handle for `Eval`, direct user-function execution, current-node
  reporting, and shared engine state.
- that shim must not grow back into a semantic bridge.
- `contracts.UserFunctionCallbacks` is temporary migration residue, not a
  target steady-state seam.

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
