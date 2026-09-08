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
  - after `4.11`, it is reduced to `Value`, `ClassMetaValue`,
    `ExternalFunctionRegistry`, `ExternalFunctionSignature`, and `EngineState`

## Ownership rules

- `ExecutionContext` is the canonical owner of per-run mutable state. Builtin calls
  receive an invocation-scoped context adapter; the evaluator has no active-context
  or current-node fields. Nested evaluation restores the node on its explicit context.
- Production bootstrap is centralized in `internal/interp/new.go`.
- `internal/interp` must not import `internal/interp/evaluator` outside construction and tests.
- Runtime execution should not bounce through callback-style interpreter bridges.
- Runtime metadata should prefer typed runtime structures over AST-shaped compatibility maps where possible.
- External-function integration remains shell-owned for host invocation, but
  argument preparation is evaluator-owned and the handoff uses runtime values,
  not AST-shaped callback round-trips.
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
- host external-function invocation is an intentional shell concern; evaluator
  owns external-call argument preparation before the value-level handoff.
- the remaining evaluator shim is an explicitly retained minimal internal
  implementation handle for `Eval`, direct user-function and function-pointer execution,
  and shared engine state. The shell reads node positions from its execution context.
- that shim must not grow back into a semantic bridge.
- `contracts.UserFunctionCallbacks` is temporary migration residue, not a
  target steady-state seam.
  - update: removed from `contracts` during `4.11`

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

---

## Appendix A — Runtime ownership boundary (from Phase 4.1)

- `Interpreter` is the public engine facade; `ExecutionContext` is the canonical owner of per-run mutable state.
- `runtime` owns value containers, call stack, control flow, property context, and exception data structures.
- Behind `ExecutionContext`: active exception, handler exception, call stack, environment stack, property evaluation context, and the old-value capture used by contracts/postconditions.
- Removed for good: exception getter/setter callback plumbing, env-sync callbacks, `SetFocusedInterfaces()`, and duplicated interpreter-owned copies of per-run state.

## Appendix B — Seam decisions (from Phase 4.10.2)

Justified, permanent boundaries:

- `internal/interp/new.go` importing `internal/interp/evaluator` — the construction boundary; wires environment, type system, evaluator, interpreter facade, and shared `EngineState`.
- Shared `EngineState` — neutral coordination state for construction, semantic info, unit state, refcount manager, and registry handles.
- External-function integration — a shell/host boundary. The evaluator owns signature-aware argument preparation (including var-parameter references) and calls `EngineState.ExternalFunctionCaller` with runtime values only; the shell performs host invocation and FFI panic/error handling. AST nodes never round-trip back into interpreter code.
- The evaluator shim — retained as a minimal internal handle (`Eval`, `ExecuteUserFunctionDirect`, `ExecuteFunctionPointerDirect`, shared `EngineState`). It must not grow back into a semantic bridge.

Resolved residue: `contracts.UserFunctionCallbacks` was removed in 4.11; user-function execution policy is evaluator-native.

## Appendix C — Allowed `internal/interp` responsibilities (from Phase 4.10.5)

`internal/interp` may own only:

1. **Production bootstrap** — `new.go`, environment creation, type-system wiring, `EngineState` and refcount/destructor setup.
2. **Public engine-facing API and orchestration** — `Interpreter.Eval`, `EvalWithExpectedType`, exception/query helpers, source/semantic-info configuration; all delegating AST semantics to the evaluator.
3. **Host and unit integration** — external-function registration/invocation plumbing, unit registry load/init coordination, Go callback / FFI entry points.
4. **Declaration/bootstrap mutation of registries and metadata** — the `evalFunctionDeclaration`, `evalClassMethodImplementation`, `evalRecordMethodImplementation`, `evalClassDeclaration`, `evalInterfaceDeclaration`, `evalOperatorDeclaration`, `evalHelperDeclaration`, `evalTypeDeclaration`, `evalEnumDeclaration` family.
5. **Narrow runtime helper primitives** used by shell-owned integration — `evalViaEvaluator`, lvalue/reference helpers, small value-operation helpers, call-stack/error-location support.

Not allowed to reappear in `internal/interp`: production statement execution (program/block/if/case/loop/try/raise/assignment/break/continue/exit), production expression execution (method dispatch, property read/write, casts, `as`, `Default(...)`, record method execution, helper method/property execution), or any shadow evaluator entry point. `internal/interp/boundary_test.go` is the executable source of truth for the allowlist; anything outside it is suspicious by default and should be moved to evaluator/runtime, justified explicitly, or deleted.

The measured state of this boundary as of 2026-09 (dead residue, remaining duplication, and the ranked refactoring list) is in `audit-2026-09.md`.

## September 2026 refactoring boundaries

- `runtime` owns `ClassInfo`, `ClassValue`, `ClassInfoValue`, class metadata mutation,
  and class operator storage. `interp` retains compatibility aliases and constructor
  wrappers. `ClassRegistry` stores `runtime.IClassInfo`; construction uses runtime
  constructors directly, without class factory callbacks. Other registries and the
  parallel class AST maps remain migration work under A6.
- `ast.SemanticInfo` retains resolved `internal/types.Type` objects keyed by AST node,
  alongside compatibility text annotations. Annotation-based evaluator resolution
  consumes these objects before legacy lookup, preserving type identity and bounds.
  Untyped execution and remaining name-based callers still use runtime resolution (A5).
- Frontend unit and program analyzers share one semantic metadata table. The compiled
  unit registry retains the analyzed ASTs; execution clones mutable registry state for
  each run. Unit dependencies receive analysis before the importing program.
- Builtin registry signatures own ordinary argument and result validation. Semantic
  diagnostic styles retain historical wording without redefining signatures; AST-sensitive
  and argument-dependent builtin rules remain explicit.
- `internal/types` owns the builtin-helper catalog and overload signature ranking.
  Semantic analysis and runtime evaluation consume these shared declarations; the
  evaluator does not import `internal/semantic`.
- `interp.New` / `NewWithOptions` are the bootstrap entry points. The old `runner`
  pass-through has been removed.
- The interpreter does not implement `builtins.Context`. Host and FFI function-pointer
  calls enter the evaluator with an explicit `ExecutionContext`, including builtin
  pointers and captured lambdas. Builtin execution uses the same registry as AST calls.
- Library-only FFI value constructors remain live even when CLI-only reachability
  analysis labels them unreachable. See the September progress log for the cleanup.
