# Phase 4.1 Boundary

Phase 4.1 establishes the runtime ownership boundary that Phase 4 depends on.

## Runtime Boundary

- `Interpreter` is the surviving public engine facade during the collapse.
- `ExecutionContext` is the canonical owner of per-run mutable state.
- `runtime` owns value containers, call stack, control flow, property context, and exception data structures.
- `evaluator` is a transitional implementation detail used by the engine, not a peer runtime owner.

## What Moved Behind `ExecutionContext`

- active exception
- handler exception
- call stack
- environment stack
- property evaluation context
- old-value capture used by contracts/postconditions

## What Was Removed

- exception getter/setter callback plumbing on `ExecutionContext`
- env-sync callbacks used only to keep interpreter fallback paths coherent
- duplicated interpreter-owned copies of per-run state

## Final Phase 4 Outcome

The transitional callback seam described during Phase 4.1 no longer exists:

- `SetFocusedInterfaces()` is gone
- the focused callback interfaces are gone
- `ExecutionContext` remained the canonical per-run state owner through the rest of Phase 4

This note is kept as the early-boundary snapshot that Phase 4 built on, not as
the final architecture description. For the final post-Phase-4 state, see
`docs/phase-4.7-verification.md`.
