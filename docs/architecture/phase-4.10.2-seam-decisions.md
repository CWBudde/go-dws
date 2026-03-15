# Phase 4.10.2 Seam Decisions

This note records which remaining `interp` ↔ `evaluator` seams are part of the
intended shell/core boundary and which are still migration residue.

Update after `4.10.4`: the callback-heavy interpreter bridge was reduced
further. The remaining evaluator shim is now an explicitly retained minimal
internal implementation handle, not the broader migration residue described in
the earlier state of this note.

The important distinction is:

- a justified shell boundary exists because `interp` owns engine/bootstrap or
  host-integration concerns
- migration residue exists only because some execution behavior has not yet been
  collapsed under its final owner

## Summary

After the `4.9` cleanup and the `4.10.3` dispatch cleanup already completed,
the remaining seams are small enough to classify explicitly.

### Permanent or justified boundaries

- `internal/interp/new.go` importing `internal/interp/evaluator`
  - justified construction boundary
  - `interp` is the surviving engine facade and owns production bootstrap
- shared `EngineState`
  - justified neutral coordination state
  - construction, semantic info, unit state, refcount manager, and registry
    handles need a single shared owner
- external function registry ownership
  - justified shell/integration boundary
  - host-registered Go functions are part of the embedding surface, not core AST
    semantics

### Temporary migration residue

- `EngineState.ExternalFunctionCaller`
  - temporary in its current shape
  - the problem is not external-function integration itself; the problem is that
    the seam still passes AST expressions back into interpreter-owned dispatch
- `contracts.UserFunctionCallbacks`
  - temporary
  - evaluator already has native defaults for implicit conversion, result
    initialization, return conversion, and interface cleanup/refcount handling
  - the callback bundle is now mostly duplicate policy rather than a necessary
    boundary

## Decision Detail

### 1. Construction boundary stays in `interp`

`internal/interp/new.go` is the one place where importing evaluator from interp
is correct. It wires:

- environment
- type system
- evaluator instance
- interpreter facade
- shared `EngineState`

This is the intended bootstrap direction and should remain allowed.

### 2. External function integration is a real boundary, but the callback shape is not

External Go functions are engine-host integration. That belongs at the shell
boundary.

However, the current seam is still too interpreter-shaped:

- evaluator calls `EngineState.ExternalFunctionCaller`
- the callback takes `[]ast.Expression`
- interpreter then performs the external call path

That means the seam is still AST-shaped and callback-mediated.

Decision:

- keep external-function ownership at the shell/integration boundary
- remove the AST-shaped callback form during later cleanup
- final seam should be narrower and centered on registry/invocation primitives,
  not evaluator round-tripping AST nodes back into interpreter

### 3. User-function callbacks are not a justified steady-state boundary

`contracts.UserFunctionCallbacks` was useful during migration, but it no longer
looks architecturally necessary.

Evaluator already provides native behavior for:

- parameter implicit conversion
- result/default initialization
- function-name aliasing to `Result`
- return-value conversion
- interface refcount increment on return
- interface cleanup at function exit

Interpreter still creates the callback bundle in order to preserve old behavior
for some call sites, but this is now duplication rather than a principled shell
boundary.

Decision:

- treat the callback bundle as migration residue
- collapse remaining user-function execution onto evaluator-native defaults
- then remove `contracts.UserFunctionCallbacks` from the steady-state boundary

### 4. The evaluator shim is retained only as a minimal internal handle

`4.10.4` removed the callback-heavy parts that made the shim migration residue:

- interpreter-side user-function callback bundles no longer drive execution
- builtin helper method/property execution no longer round-trips through shim
  methods
- interpreter no longer needs evaluator current-context access

The remaining shim surface is now limited to a small internal implementation
handle:

- `Eval`
- `ExecuteUserFunctionDirect`
- current-node access for shell-side error reporting
- shared `EngineState` access

Decision:

- explicitly retain this minimal shim as an internal shell/core handle
- do not grow it back into a semantic bridge
- any new execution semantics belong in evaluator/runtime, not on the shim

## Implications For The Remaining Plan

### 4.10.3

Focus on removing the last AST-shaped or callback-mediated execution paths that
still route through interpreter-owned helpers.

### 4.10.4

Collapse or delete the remaining evaluator shim after the last interpreter-side
callers are moved.

### 4.10.5

Document the final allowed shell responsibilities as:

- production bootstrap
- public engine-facing API
- unit integration
- host/external-function integration
- narrowly justified orchestration only

Not allowed:

- interpreter-owned production AST semantics
- AST-shaped callback round-trips from evaluator back into interpreter
