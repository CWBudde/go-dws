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

## Remaining Transitional Seam

`SetFocusedInterfaces()` still exists because Phase 4.3 has not deleted the
legacy callback surfaces yet. That seam is now isolated behind interpreter
construction rather than spread across runtime entry points.
