<!-- trunk-ignore-all(prettier) -->
# DWScript to Go Port - Detailed Implementation Plan

This document breaks down the ambitious goal of porting DWScript from Delphi to Go into bite-sized, actionable tasks organized by stages. Each stage builds incrementally toward a fully functional DWScript compiler/interpreter in Go.

---

## Phase 1: Lexer (Tokenization)

**Completed**

---

## Phase 2: Parser Modernization ✅ **COMPLETED** (2025-01-21)

**Accomplishments**:

- Transformed parser to 100% cursor-based architecture (immutable TokenCursor)
- Built modern infrastructure: combinators, structured errors, automatic position tracking
- Removed ~6,700 lines of legacy code (31% reduction: 21K → 14.6K lines)
- Eliminated all dual-mode parsing (Traditional vs Cursor)
- Separated error recovery and semantic analysis into dedicated modules
- Improved test coverage from 73.4% → 78.5% (+700 test lines)

**Phases**: 2.1 Foundation, 2.2 Cursor Abstraction, 2.3 Combinators, 2.4 Position Tracking, 2.5 Separation of Concerns, 2.6 Advanced Features, 2.7 Traditional Removal, 2.8 Test Coverage

**Documentation**: [PHASE2_MODERNIZATION.md](PHASE2_MODERNIZATION.md), [PHASE2_COMPARISON.md](PHASE2_COMPARISON.md)

---

## Phase 3: Interpreter Architecture Refactoring ✅ **COMPLETED** (2025-12-11)

**Accomplishments**:

- Unified dual environments: `i.env` + `ctx.env` → single canonical environment in ExecutionContext
- Migrated Eval() switch: 21/59 cases (36%) delegated to Evaluator, eliminated 25 circular `adapter.EvalNode()` callbacks
- Reduced Interpreter fields: 34 → 14 fields (-59%), migrated execution state to Evaluator
- Replaced 140-line monolithic adapter with 4 focused interfaces (68 methods total)
- Removed ~2,550 lines of code across environment unification and field migration
- All 386 unit tests passing, no new fixture test failures

**Phases**: 3.1 Environment Unification, 3.2 Eval() Switch Migration, 3.3 Field Migration, 3.4 Focused Interfaces, 3.5 Code Quality

**Key Docs**: [environment-audit.md](docs/environment-audit.md), [phase-3.2-summary.md](docs/phase-3.2-summary.md), [adapter-method-audit.md](docs/adapter-method-audit.md)

**Honest Assessment**: Phase 3 achieved foundation work (environment unification, interface extraction) but did not achieve the stated goal of a "thin coordinator". Interpreter still has 434 methods and 59 switch cases. Evaluator added 341 methods. See Phase 4 for completing the migration.

---

## Phase 4: Collapse To A Single Execution Engine

**Status**: ✅ Complete | **Priority**: High | **Completed**: 2026-03-10

### Why This Phase Exists

The core problem is **split ownership**, not raw method count.

The runtime still behaves like it has two execution cores:

```text
construction:
  Interpreter + Evaluator

execution:
  Evaluator → callback interface → Interpreter → Evaluator
```

That shape keeps recreating the same complexity under new names:

```text
Interpreter → "too big"
Evaluator   → "too big"
Handlers    → "too big"
Adapters    → "too big"
```

Phase 4 is about deleting that split, not just moving code around.

### Current Branch Snapshot

Work completed so far has materially changed the runtime:

- `go test ./...` is green again
- `ExecutionContext` now owns per-run mutable state
- `Interpreter` is the chosen surviving public engine facade
- `runner` now delegates construction through one interpreter entry point
- declaration ownership has moved significantly toward `runtime` + `TypeSystem`
- direct user-function execution no longer round-trips through the old OOP callback path

What Phase 4 delivered:

- callback-style execution interfaces and bridge-only files are deleted
- evaluator-owned execution no longer round-trips through interpreter callbacks
- `ExecutionContext` is the canonical per-run mutable state owner
- production construction is centralized in one bootstrap path
- boundary and regression tests now protect the finished architecture

### Target Architecture

The codebase should converge on this shape:

```text
Runner → Engine.Eval(node, ctx)
                 ↓
        direct methods / direct helpers
                 ↓
        runtime values + metadata + type system
```

Not this:

```text
Runner → Interpreter ↔ Evaluator ↔ Handler/Adapter ↔ Interpreter
```

### Guardrails

Every PR in Phase 4 must satisfy at least one measurable outcome:

- deletes a callback interface or reduces its method surface
- deletes a dispatch hop
- removes duplicated runtime state ownership
- removes a constructor/wiring path that creates two runtime peers
- converts callback-driven logic into direct engine ownership

If a change only renames, moves, or re-groups methods without removing indirection or ownership splits, it is out of scope.

### Phase Complete When

- [x] exactly one concrete execution engine remains in the interpreter runtime
- [x] `ExecutionContext` is the only owner of per-run execution state
- [x] `CoreEvaluator`, `OOPEngine`, `DeclHandler`, and `SetFocusedInterfaces()` are deleted
- [x] no evaluator → interpreter or interpreter → evaluator callback path remains
- [x] `runner` constructs one engine, not two peers plus glue

### 4.0 Stabilize The Baseline

**Goal**: Restore a trustworthy green baseline before structural work continues.

**Status**: ✅ Complete

**Tasks**:

- [x] **4.0.1** Establish failing-test baseline
- [x] **4.0.2** Fix class/static identifier resolution
- [x] **4.0.3** Fix record literal type-context propagation
- [x] **4.0.4** Fix indexed/default property metadata + dispatch
- [x] **4.0.5** Fix indexed property `WriteKind` determination
- [x] **4.0.6** Fix record value semantics (copying bug)
- [x] **4.0.7** Fix "Object not instantiated" in implicit conversions
- [x] **4.0.8** Fix precondition/contract handling
- [x] **4.0.9** Fix type cast lvalue and interface property setters
- [x] **4.0.10** Fix interface reference counting and destructor timing
- [x] **4.0.11** Fix interface property getter output
- [x] **4.0.12** Fix implicit property access with context
- [x] **4.0.13** Fix record literal error handling
- [x] **4.0.14** Fix record function return type initialization
- [x] **4.0.15** Fix compound assignment operators on classes
- [x] **4.0.16** Fix conversion chain value propagation
- [x] **4.0.17** Fix implicit conversion return value handling

**Exit Criteria**:

- `go test ./...` green
- no new fixture failures

---

### 4.1 Make State Ownership Real

**Goal**: Remove duplicated ownership of execution state.

**Status**: ✅ Complete

**Principle**:

- `ExecutionContext` owns all per-run mutable state
- the engine reads/writes that state directly
- no callback should be needed just to see the "real" environment or exception

**Tasks**:

- [x] **4.1.1** Audit all duplicated execution state
  - `exception`
  - `handlerException`
  - `propContext`
  - `callStack`
  - `oldValuesStack`
  - environment stack / env sync paths
- [x] **4.1.2** Move duplicated per-run state behind `ExecutionContext`
- [x] **4.1.3** Remove exception getter/setter callback plumbing from `ExecutionContext`
- [x] **4.1.4** Delete env-sync workaround paths introduced for callback fallback
- [x] **4.1.5** Add explicit engine/context boundary documentation
  - See [docs/phase-4.1-boundary.md](docs/phase-4.1-boundary.md)

**Success Criteria**:

- ✅ no "sync interpreter state into ctx" logic remains
- ✅ `ExecutionContext` is the single source of truth for per-run state
- ✅ per-run state is not duplicated on multiple runtime peers

---

### 4.2 Bring Declaration Ownership Under Canonical Runtime Types

**Goal**: Make declaration processing mutate `runtime` / `TypeSystem` directly instead of preserving callback-owned mirrors.

**Status**: ✅ Complete

**Why this moved earlier**:

This work has already happened in the branch and should be reflected as completed groundwork, not hidden behind a future "callback deletion" bucket.

**Tasks**:

- [x] **4.2.1** Audit declaration data ownership
  - classes, interfaces, helpers, global functions, records, and method implementations
- [x] **4.2.2** Move self-contained declaration nodes off `CoreEvaluator.EvalNode`
  - `OperatorDecl`, `EnumDecl`, and `TypeDeclaration` now execute through evaluator-native visitors
- [x] **4.2.3** Move interface declaration mutation under direct engine ownership
  - concrete interface metadata now lives in `runtime`
- [x] **4.2.4** Move helper declaration mutation under direct engine ownership
  - concrete helper metadata now lives in `runtime`
- [x] **4.2.5** Move class metadata mutation under direct engine ownership
  - flags, parent linkage, constants, fields, class vars, nested classes, property storage, VMT build, and registration now execute directly on `ClassInfo`
- [x] **4.2.6** Make method implementation registration direct
  - evaluator now registers class and record method implementations directly
- [x] **4.2.7** Make global function registration `TypeSystem`-owned
- [x] **4.2.8** Remove record/function mirror registries where `TypeSystem` is already canonical
- [x] **4.2.9** Ensure environment binding for declarations is direct, not adapter-mediated
- [x] **4.2.10** Centralize low-risk class reads behind one lookup boundary
  - class casts, default typed-nil creation, exception lookup, and metadata-to-class resolution now use a shared class lookup helper
- [x] **4.2.11** Move legacy class declaration lookup and descendant propagation onto the canonical class registry
- [x] **4.2.12** Remove the remaining `i.classes` compatibility mirror
  - interpreter class lookup now resolves through the canonical registry helper instead of raw interpreter-owned class maps
- [x] **4.2.13** Remove the last declaration-time class allocation adapter
  - evaluator class declaration allocation now goes through `TypeSystem.ClassInfoFactory`
  - `DeclHandler.NewClassInfoAdapter` is no longer used by declaration processing

**Success Criteria**:

- [x] declaration processing has one primary runtime owner
- [x] record/function registration is no longer mirrored across peer objects
- [x] class registration no longer depends on `i.classes`
- [x] declaration processing no longer depends on `DeclHandler`

---

### 4.3 Collapse Construction And Shared Engine State

**Goal**: Stop wiring two runtime peers together during construction.

**Status**: 🚧 In progress

**Decision**:

`Interpreter` is the surviving public engine facade during the collapse. The evaluator may remain as an internal implementation detail temporarily, but not as an independently wired peer.

**Tasks**:

- [x] **4.3.1** Choose the surviving engine type explicitly in code/docs
- [x] **4.3.2** Change `runner` to construct one engine entry point
- [x] **4.3.3** Remove direct `runner` wiring of evaluator/interpreter glue
- [x] **4.3.4** Move shared runtime metadata/config behind shared engine state
  - source file/source code
  - semantic info
  - unit registry + loaded/initialized unit state
  - external function registry
  - RNG seed/source
  - refcount manager handle
- [x] **4.3.5** Remove constructor-time `SetFocusedInterfaces()` wiring
- [x] **4.3.6** Narrow production construction to only the remaining runtime bridges
  - production bootstrap now wires OOP + fallback bridges explicitly
  - declaration wiring is no longer part of runtime construction
- [x] **4.3.7** Remove remaining evaluator-owned config/state that is not visitor-local
  - source file/source code are now engine-state owned only
  - max recursion depth is now shared through engine state instead of split across evaluator config and interpreter fields
- [x] **4.3.8** Reduce the non-surviving peer to a thin internal shim only where temporarily unavoidable
  - interpreter now depends on a narrow evaluator shim instead of the broad evaluator interface
  - the non-surviving peer is retained only for eval, user-function execution, current-node tracking, and shared-engine-state access
- [ ] **4.3.9** Delete the shim once call sites are moved

**Success Criteria**:

- [x] `runner` creates one public engine entry point
- [x] no constructor-time callback wiring remains
- [x] the remaining shared runtime config/state is no longer evaluator-owned
- [x] the non-surviving peer has been reduced to a narrow internal shim
- [ ] the shim can be deleted entirely once remaining runtime behavior is moved

---

### 4.4 Move Execution Semantics Off Callback Round-Trips

**Goal**: Consolidate runtime semantics under direct engine-owned paths before deleting the callback interfaces.

**Status**: 🚧 In progress

- [x] **4.4.1** Remove the `CallUserFunction` round-trip for direct user function calls
- [x] **4.4.2** Make exception creation/cleanup evaluator-owned instead of using the old exception callback surface
- [x] **4.4.3** Keep assignment statements on evaluator-owned paths instead of bouncing through `EvalNode`
- [x] **4.4.4** Move builtin helper property access and class property read/write onto evaluator-owned paths
- [x] **4.4.5** Finish migrating remaining builtin helper method execution off `EvalNode()` fallback
- [x] **4.4.6** Remove the `TYPE_CAST`-specific `EvalNode()` fallback from method dispatch
- [x] **4.4.7** Remove remaining declaration/default `EvalNode()` fallbacks from `Evaluator.Eval()`
- [x] **4.4.8** Remove remaining evaluator-package `coreEvaluator.EvalNode()` dispatch
  - record instance/static method execution now runs through evaluator-owned dispatch instead of bouncing through the interpreter
- [x] **4.4.9** Remove object-method `ExecuteMethodWithSelf` usage from evaluator property/inherited/self dispatch
  - object property getters/setters, indexed property getters, inherited object calls, and implicit parameterless self-method execution now run through evaluator-owned object method execution
- [x] **4.4.10** Remove class-method execution from the `oopEngine` bridge
  - direct class-method execution now binds `Self` / `__CurrentClass__` inside the evaluator, and class-method scope resolution no longer assumes only `CLASSINFO` can own `__CurrentClass__`
- [x] **4.4.11** Normalize remaining call execution under one path
  - function pointers and implicit self calls moved to evaluator-owned paths in earlier tasks
  - record static calls run through evaluator-owned `callRecordStaticMethod`
  - external functions have one normalized path: `oopEngine.CallExternalFunction`
- [x] **4.4.12** Normalize constructor/destructor execution under one path
  - both constructor call sites (`method_dispatch.go`, `visitor_expressions_functions.go`) use `oopEngine.ExecuteConstructor` — one normalized path
  - destructors not yet implemented
- [x] **4.4.13** Normalize casts, wrappers, and operator overload dispatch under one path
  - type casts: `oopEngine.CreateTypeCastWrapper` (2 call sites, one method)
  - binary operator overloads: `oopEngine.TryBinaryOperator` (3 call sites across binary/compound ops)
  - unary operator overloads: `oopEngine.TryUnaryOperator` (1 call site)
  - subrange/interface wrapping: `oopEngine.WrapInSubrange` / `oopEngine.WrapInInterface` (one each)
- [x] **4.4.14** Remove remaining direct `oopEngine` usage from visitor/helper files
  - object/record/default indexed setter execution no longer uses `ExecuteMethodWithSelf`
  - remaining seams (21 calls): `CreateBoundMethodPointer` (5), `LookupClassByName` (3), `ExecuteRecordPropertyRead` (2), `CallMethod` (1), `ExecuteConstructor` (2), `CreateTypeCastWrapper` (2), `TryBinaryOperator`/`TryUnaryOperator` (3), `WrapInSubrange`/`WrapInInterface` (2), `CallExternalFunction` (1)
- [x] **4.4.15** Normalize method/property dispatch under one path
  - [x] bound method pointer creation: all 7 call sites replaced with evaluator-owned `createFunctionPointerFromDecl`
  - [x] indexed record property reads: 2 call sites replaced with evaluator-owned `executeRecordIndexedPropertyRead` (`class_property_helpers.go`)
  - [x] class metadata lookup: 2 of 3 `oopEngine.LookupClassByName` call sites replaced with `typeSystem.CreateClassValue`
  - [x] OBJECT dispatch: new `dispatchObjectMethod` routes non-overloaded instance methods evaluator-side; virtual dispatch via `IClassInfo.LookupMethod`; destructor via `runObjectDestructor`
  - [x] CLASS/CLASSINFO dispatch: evaluator-owned via `callClassConstructor`/`callClassMethod`; adapter fallback only for overloaded class methods
  - [x] `executeObjectMethodDirect`: sets `__CurrentClass__` via `typeSystem.CreateClassValue` so nested class lookup works in method bodies
  - [x] remaining class lookup seam cleaned up in `4.6.1`; nested class lookup now uses canonical current-class helpers
  - [x] remaining duplication/metadata cleanup in `4.6` reduced the old bridge-shaped follow-up seams instead of leaving them open

**Success Criteria**:

- [x] one dispatch path exists per language feature category
- [x] no "fallback to legacy path" branch remains in the core evaluator/interpreter flow

---

### 4.5 Delete Callback Surfaces

**Goal**: Delete the callback interfaces once the surviving execution paths no longer need them.

**Status**: ✅ Complete

**Current callback surfaces to remove**:

- `CoreEvaluator`
- `OOPEngine`
- `DeclHandler`
- `SetFocusedInterfaces()`

**Tasks**:

- [x] **4.5.1** Delete `CoreEvaluator`
  - interface deleted from contracts.go; `coreEvaluator` field removed from Evaluator; `SetRuntimeBridge` simplified to 1 parameter
- [x] **4.5.2** Delete `OOPEngine`
  - `OOPEngine` interface deleted from contracts.go and focused_interfaces.go; `oopEngine` field and `SetRuntimeBridge` removed from Evaluator
  - New evaluator-owned implementations in `runtime_ops.go`: `wrapInSubrange`, `wrapInInterface`, `createTypeCastValue`, `executeConstructorForObject`, `dispatchObjectMethodOverloaded`, `dispatchInterfaceMethodDirect`, `dispatchClassMethodOverloaded`, `evalTryBinaryOperator`, `evalTryUnaryOperator`
  - `IClassInfo` extended with `GetMethodOverloads`, `GetClassMethodOverloads`, `GetConstructorOverloads`
  - `ExternalFunctionCaller` callback wired via `EngineState` to break circular dependency
  - All 15 `e.oopEngine.XXX()` call sites replaced with evaluator-owned code
- [x] **4.5.3** Delete `DeclHandler`
  - interface deleted from contracts.go; `SetFocusedInterfaces` simplified to 1 parameter (OOPEngine only); `panicDeclHandler` test removed
- [x] **4.5.4** Delete `SetFocusedInterfaces()` and all runtime uses
  - all 8 call sites in test files replaced with `SetRuntimeBridge`; method deleted from evaluator
- [x] **4.5.5** Replace adapter-based tests with direct engine tests where those adapters only exist to model the old split
  - adapter-shaped conversion/integration test scaffolding removed; tests now execute through the direct evaluator/engine boundary
- [x] **4.5.6** Delete any remaining bridge files that only preserve the old interpreter/evaluator seam
  - dead tombstone file `internal/interp/evaluator/focused_interfaces.go` deleted; no remaining bridge-only evaluator files left from the old split

**Success Criteria**:

- [x] zero internal callback interfaces remain (CoreEvaluator, DeclHandler, OOPEngine all deleted)
- [x] no runtime code depends on callback injection to reach core semantics
- [x] no tests require mock adapters for the final architecture boundary (4.5.5)

---

### 4.6 Remove Remaining Mirror Metadata And Hybrid Ownership

**Goal**: Finish the cleanup exposed by the ownership collapse.

**Status**: ✅ Complete

**Tasks**:

- [x] **4.6.1** Remove the remaining class-lookup compatibility seam and convert remaining call sites to canonical class-scope helpers
  - evaluator nested-class identifier resolution now uses the canonical current-class context helper (`__CurrentClass__` or `Self.ClassType`) instead of a trailing fallback branch
  - dead interpreter `LookupClassByName` shim removed
- [x] **4.6.2** Audit hybrid metadata types
  - audited the remaining hybrid owners: `ClassInfo` and `runtime.IClassInfo` still carry the largest AST-bearing surface; interface/helper metadata remain partially AST-backed for method dispatch; record metadata now uses a typed runtime metadata path instead of `any`
  - `TypeSystem.LookupRecordMetadata()` now returns `*runtime.RecordMetadata`, removing compatibility-style assertions from evaluator record creation paths
- [x] **4.6.3** Remove duplication that no longer serves a compatibility need
  - runtime object method queries and inherited dispatch now use `LookupMethod()` directly instead of legacy `GetMethodsMap()` access
  - interpreter method-pointer lookup and inherited dispatch likewise use canonical method lookup APIs
- [x] **4.6.4** Keep AST-bearing references only where execution genuinely requires them
  - class field existence checks now use `FieldExists()` instead of `GetFieldsMap()` in assignment/property execution
  - remaining `GetFieldsMap()` usage is confined to field-initializer execution paths and tests, where `*ast.FieldDecl.InitValue` is still genuinely required
- [x] **4.6.5** Move toward runtime-native metadata where it reduces branching and proxy behavior
  - record metadata is now typed end-to-end (`*runtime.RecordMetadata`)
  - method/property dispatch paths prefer runtime/class lookup APIs over legacy map exposure, shrinking proxy behavior at the class/runtime boundary

**Success Criteria**:

- [x] metadata cleanup follows ownership cleanup instead of blocking it
- [x] bridge/proxy code shrinks as a result

---

### 4.7 Verification And Metrics

**Goal**: Prove the new architecture is simpler in reality, not just in description.

**Status**: ✅ Complete

**Tasks**:

- [x] **4.7.1** Add architecture boundary tests
  - `boundary_test.go` now verifies the import boundary, that construction files do not reference legacy bridge wiring, and that legacy callback interface declarations do not reappear
- [x] **4.7.2** Add regression tests for the migrated execution clusters
  - `phase4_regression_test.go` covers assignment paths, user-function execution, and declaration ownership through the final engine path
- [x] **4.7.3** Track removal metrics
  - see `docs/phase-4.7-verification.md` for current boundary/cleanup metrics and verification commands
- [x] **4.7.4** Run unit tests and fixture subsets after each structural milestone
  - focused Phase 4.7 architecture/regression suite run plus full `go test ./...`; fixture-derived declaration coverage is included in the new regression suite
- [x] **4.7.5** Keep a short migration note in docs describing the final engine/context/runtime boundary
  - added `docs/phase-4.7-verification.md`

---

### Execution Order From Here

The Phase 4 work proceeded in this order:

`4.3 Collapse construction/state` → `4.4 finish moving execution semantics` → `4.5 delete callback surfaces` → `4.6 remove remaining mirrors/metadata cleanup` → `4.7 verification and metrics`

---

### 4.8 Post-Phase-4 Refinement

**Goal**: Tighten the final architecture now that the Phase 4 migration is complete, without reopening the completed ownership/callback split.

**Status**: ✅ Complete

**Tasks**:

- [x] **4.8.1** Remove stale comments and historical callback-era wording
  - rewritten code/docs that still described deleted callback surfaces as current architecture
  - historical docs now clearly mark themselves as migration snapshots rather than the final steady state
- [x] **4.8.2** Reduce remaining interpreter-side legacy dispatch
  - audited `interpreter.go` and adjacent legacy helpers
  - removed the dead `EvalNode()` and `evalLegacy()` interpreter dispatch entry points
  - the surviving interpreter-side execution now remains only behind directly called interpreter helpers rather than a central unused legacy dispatcher
- [x] **4.8.3** Continue shrinking AST-bearing runtime metadata
  - evaluator object construction now initializes fields from `runtime.ClassMetadata.Fields` / `runtime.FieldMetadata`
  - production evaluator constructor paths no longer depend on `GetFieldsMap()` / `GetFieldTypesMap()` compatibility lookups
  - the remaining AST-bearing class metadata is now concentrated in paths that still genuinely need declaration bodies rather than default-value materialization
- [x] **4.8.4** Narrow the neutral `contracts` boundary further where possible
  - audited `internal/interp/contracts` after Phase 4 completion
  - removed the dead `contracts.Evaluator` interface; `interpreter.go` already relies on its narrower local shim
- [x] **4.8.5** Strengthen long-term regression protection
  - expanded the Phase 4 regression suite with end-to-end coverage for runtime-metadata-based object field initialization, evaluator-owned call execution, and inherited virtual dispatch
  - the added tests exercise the final engine boundary instead of compatibility seams or adapter-style test doubles

**Success Criteria**:

- [x] no major code comment or doc still describes the deleted callback boundary as the intended architecture
- [x] interpreter-only legacy dispatch is smaller and more explicit than at Phase 4 completion
- [x] runtime/class/helper metadata uses fewer AST-backed escape hatches
- [x] the remaining `contracts` package surface is clearly justified by actual cross-package needs
- [x] regression coverage is stronger for the final post-Phase-4 architecture

### Archived Direction

The older callback-removal plan is preserved in Git history for reference, but it is no longer the architectural target.

The project should not optimize for "smaller types" if that preserves split ownership.

---

### 4.9 Remove Residual Split Execution Ownership

**Goal**: Finish the core ownership cleanup by removing interpreter-side shadow execution paths and migrating the remaining live semantic islands into evaluator-owned execution.

**Status**: ✅ Complete

**Why this needs its own follow-up**:

Phase 4 successfully removed the old callback bridge and centralized production execution through the evaluator, but it did not fully eliminate interpreter-side duplicate AST evaluators. Some statement/control-flow logic still exists in `internal/interp` as dead or semi-dead shadow implementations, while the active evaluator path owns the real execution semantics.

That leftover duplication is now an architectural risk:

- correctness can silently diverge between the active evaluator path and stale interpreter-side copies
- fixture failures can expose evaluator regressions that the older interpreter helpers already solved
- boundary tests currently protect imports and deleted callback surfaces, but not single ownership of AST semantics

**Concrete symptoms already observed**:

- interpreter-side `for` loop logic still models DWScript ordinal loop semantics correctly
- evaluator-side `VisitForStatement` regressed to integer-only loop bounds/step handling
- several interpreter-side statement evaluators remain in tree even though normal runtime entry goes through `Evaluator.Eval`
- typed literal evaluation still has interpreter-side semantic islands (`evalArrayLiteralWithExpected`, related array/set literal helpers)

**Tasks**:

- [x] **4.9.1** Inventory all remaining interpreter-side AST execution helpers and classify each as `delete`, `migrate`, or `keep`
  - cover statement, expression, declaration, exception, and control-flow helpers in `internal/interp`
  - distinguish dead duplicates from still-live shell responsibilities
- [x] **4.9.2** Delete dead duplicate statement/control-flow evaluators that overlap evaluator-owned semantics
  - removed the dead interpreter-side statement/declaration/loop execution cluster after confirming production entry flows through evaluator-owned visitors
  - remaining shared helper utilities stay only where still used by live interpreter/runtime code
- [x] **4.9.3** Delete the next dead duplicate cluster in expression/call/OOP execution helpers
  - expected targets include interpreter-side shadow helpers for calls, member access, method dispatch, helper dispatch, and declaration execution once lingering test-only/internal references are cleaned up
- [x] **4.9.4** Migrate still-live special-case AST execution that belongs in the evaluator
  - remove interpreter-only semantic islands such as expected-type array/set literal evaluation paths
  - keep only true engine-shell responsibilities in `internal/interp`
- [x] **4.9.5** Unify loop semantics under shared ordinal-aware helpers
  - `for` loop execution must treat DWScript ordinals canonically across evaluator/runtime paths
  - eliminate semantic drift between interpreter leftovers and evaluator visitors
- [x] **4.9.6** Add regression protection for single ownership of AST execution
  - extend architecture tests so they fail when evaluator-owned statement semantics reappear as interpreter-side duplicate evaluators
  - add fixture/unit coverage for ordinal `for` loops, enum loop variables, and enum-valued `step`

**Success Criteria**:

- [x] `internal/interp` no longer contains dead shadow evaluators for nodes already owned by evaluator visitors
- [x] loop/control-flow semantics are defined in one active implementation only
- [x] architecture tests enforce single ownership, not just import boundaries and deleted callback interfaces
- [x] interpreter-side typed literal semantics no longer bypass evaluator-owned execution

---

### 4.10 Clarify Remaining Interp/Evaluator Seams

**Goal**: Close the remaining live ownership gaps by deciding which interp/evaluator seams are intentional shell/core boundaries and moving residual production AST semantics under their final owner.

**Status**: 📋 Planned

**Why this is the next phase**:

Phase 4.9 removed the dead shadow-execution cluster and locked down ownership regressions, but it did not yet make evaluator the sole owner of all production AST execution semantics. What remains is no longer "dead duplicate code"; it is a smaller set of still-live shell/core seams that need explicit end-state decisions and, where appropriate, further migration.

**Tasks**:

- [x] **4.10.1** Inventory the remaining live production AST execution paths still owned by `internal/interp`
  - cover class/interface/type/helper/function declaration execution, type-cast/default built-ins, property/method dispatch helpers, and other still-live `Interpreter.eval*` methods
  - classify each as `keep in interp`, `move to evaluator`, or `replace with runtime primitive`
- [ ] **4.10.2** Decide and document which remaining `interp` ↔ `evaluator` seams are intentional steady-state boundaries versus temporary migration residue
  - cover the narrow evaluator shim, external-function dispatch hooks, and user-function callback plumbing
  - if a seam is permanent, document why it belongs at the engine shell boundary
  - if a seam is temporary, move it under an explicit removal task instead of leaving it implicit
- [ ] **4.10.3** Migrate remaining production AST semantics that do not belong in the interpreter shell
  - move any residual expression/declaration execution that is neither bootstrap nor external-integration orchestration under evaluator/runtime ownership
  - eliminate interpreter-owned production AST execution where the shell is only acting as a semantic executor
- [ ] **4.10.4** Close or explicitly retain the remaining narrow evaluator shim
  - resolve the unfinished `4.3.9` outcome instead of leaving it open indefinitely
  - either delete the shim after moving the last call sites or document the shim as the intended long-term shell/core boundary
- [ ] **4.10.5** Define and enforce the allowed responsibilities of `internal/interp` after cleanup
  - interpreter shell responsibilities should be explicit: bootstrap, engine-facing API, unit integration, external integration, and other narrowly justified orchestration only
  - evaluator-owned AST semantics must not reappear in `internal/interp`
- [ ] **4.10.6** Add regression protection for the final steady-state ownership boundary
  - extend architecture tests so the allowed remaining `Interpreter.eval*` surface is explicit rather than inferred
  - fail when new production AST semantics are added to `internal/interp` without an allowed-shell justification

**Success Criteria**:

- [ ] evaluator is the sole owner of AST execution semantics for production statement/expression/declaration execution, except for explicitly documented shell-only orchestration
- [ ] remaining `interp` ↔ `evaluator` seams are intentional, documented, and regression-tested
- [ ] the allowed long-term responsibilities of `internal/interp` are explicit enough that future work does not recreate split execution ownership

---

### 4.11 Finalize Neutral Boundaries And Metadata Escape Hatches

**Goal**: Shrink the remaining migration-era neutral/boundary surfaces so the final architecture no longer depends on adapter-shaped residue.

**Status**: 📋 Planned

**Tasks**:

- [ ] **4.11.1** Re-audit runtime metadata escape hatches exposed by the 4.9 cleanup
  - continue removing `GetFieldsMap` / `GetMethodsMap` / similar migration-only fallbacks where canonical metadata is now sufficient
- [ ] **4.11.2** Audit the remaining `contracts` package surface
  - classify each `contracts` type as either a justified neutral coordination type or migration residue to remove/move
- [ ] **4.11.3** Remove or narrow adapter-shaped residue that remains in neutral boundaries
  - keep `contracts` small and explicitly justified rather than letting it become a migration holding area

**Success Criteria**:

- [ ] residual runtime metadata compatibility APIs shrink further as a consequence of the cleanup
- [ ] `internal/interp/contracts` contains only explicitly justified neutral coordination types, not leftover adapter-shaped migration residue

---

### Phase 4 Effort Summary

| Area | Status | Notes |
|------|--------|-------|
| 4.0 Stabilization | complete | Green baseline restored |
| 4.1 State ownership | complete | `ExecutionContext` is now the per-run state owner |
| 4.2 Declaration ownership | complete | Class allocation and registration now sit on canonical runtime/type-system boundaries |
| 4.3 Construction collapse | complete | Production construction is centralized and narrowed |
| 4.4 Execution semantics | complete | Evaluator-owned execution paths are normalized |
| 4.5 Callback deletion | complete | Old callback surfaces and bridge files are gone |
| 4.6 Mirror/metadata cleanup | complete | Remaining mirror/metadata cleanup landed after callback deletion |
| 4.7 Verification/metrics | complete | Boundary tests, regression tests, and verification note added |
| 4.8 Refinement | complete | Post-migration tightening landed after Phase 4 completion |
| 4.9 Residual execution ownership cleanup | complete | Removed dead shadow execution, migrated typed-literal islands, unified ordinal loop semantics, and added ownership guards |
| 4.10 Remaining seam clarification | planned | Close the remaining live ownership gaps and define the final shell/core boundary |
| 4.11 Neutral boundary finalization | planned | Shrink contracts/metadata escape hatches to justified end-state boundaries |

**Practical read**: The core Phase 4 bridge-removal work is complete. Phase 4.8 tightened the finished architecture, 4.9 completed the dead-shadow cleanup plus the known live semantic-island fixes, and 4.10-4.11 now cover the remaining live ownership seams and the final shell/contracts/metadata end-state decisions.

---

## Phase 5: Interpreter Technical Debt Cleanup

**Status**: 📋 Planned | **Priority**: Medium | **Estimated**: 2-3 weeks

**Goal**: Address accumulated TODOs and technical debt in `internal/interp/...` discovered during Phase 3 work.

**Why This Matters**: These TODOs represent incomplete features, workarounds, and architectural shortcuts that may cause test failures or unexpected behavior. Addressing them incrementally prevents debt accumulation.

---

### 5.1 Circular Import Workarounds ✅ COMPLETE

**Goal**: Resolve circular import issues that forced workarounds in the type system.

**Status**: ✅ Complete | **Actual Effort**: 30 minutes

**Resolution**: The TODOs were based on **false assumptions**. Audit revealed:

- `runtime` package already imports `pkg/ident` (no circular dependency)
- `evaluator` package already imports `runtime` types directly
- The "workarounds" were just incomplete implementations, not import constraints

**Changes Made**:

- [x] **5.1.1** Audit import graph - Found no actual circular imports
- [x] **5.1.2** Fixed `type_helpers.go` - Implemented actual type extraction:
  - `getArrayElementTypeFromValue()` now returns `ArrayValue.ArrayType`
  - `getRecordTypeFromValue()` now returns `RecordValue.RecordType`
  - `getSetTypeFromValue()` now returns `SetValue.SetType`
  - `getEnumTypeFromValue()` documents that EnumValue only has TypeName string
- [x] **5.1.3** Fixed `metadata_conversion.go` - Deleted 18-line duplicate `normalizeIdentifier()`, now uses `ident.Normalize()`
- [x] **5.1.4** Fixed `method_registry.go` - Added `ident` import, uses `ident.Normalize()`
- [x] Removed redundant test `TestNormalizeIdentifier` from `metadata_test.go`

**Lesson**: Always verify assumptions before planning large refactors. The "circular import" TODOs were stale.

---

### 5.2 Incomplete Migration Tasks ✅ REVIEWED - DEFERRED TO PHASE 4

**Goal**: Complete Phase 3 migration leftovers.

**Status**: ✅ Reviewed | **Result**: All items belong in Phase 4, not Phase 5

**Assessment**:

- [x] **5.2.1** `visitor_expressions_functions.go:388` - External function handling
  - **Decision**: DEFER to Phase 4.3 (Move OOP to Handlers)
  - Requires moving `Interpreter.CallExternalFunction()` and external function registry
  - This is callback elimination work, not simple cleanup

- [x] **5.2.2** `runtime/class_interface.go:176` - `InterfaceInfo` type alias
  - **Decision**: NO ACTION NEEDED - alias is correct as-is
  - The alias `InterfaceInfo = IInterfaceInfo` provides Go-idiomatic naming
  - 66 usages of `InterfaceInfo` vs 21 of `IInterfaceInfo` - keeping the alias is right
  - Updated TODO comment to reflect this

- [x] **5.2.3** `class.go:310` - Method lookup returns AST
  - **Decision**: DEFER to Phase 4.3 (Move OOP to Handlers)
  - Requires completing method registry migration
  - Part of the broader "return callable instead of AST" initiative

---

### 5.3 Feature Implementation Gaps

**Goal**: Complete partially implemented features causing test failures.

**Status**: 📋 Planned | **Effort**: 1 week | **Risk**: Medium

**Current Issues**:

- `operators.go:119`: `is` operator missing full inheritance checking
- `enum.go:41`: Enum constant expression evaluation lacks type coercion
- `evaluator/visitor_expressions_types.go:220`: Interface `is` checks missing inheritance
- `evaluator/visitor_declarations.go:125`: Record declarations missing constants, fields, nested types
- `class_var_init_test.go:55`: Class constant expressions not supported

**Tasks**:

- [x] **5.3.1** Implement full inheritance checking for `is` operator (4h)
  - `operators.go:119`
  - Check full class hierarchy, not just direct parent

- [x] **5.3.2** Implement interface inheritance for `is` checks (3h)
  - `evaluator/visitor_expressions_types.go:220`
  - Walk interface parent chain

- [x] **5.3.3** Add type coercion to enum constant evaluation (3h)
  - `enum.go:41`
  - Handle integer-to-enum coercion in constant expressions

- [x] **5.3.4** Complete record declaration handling (4h)
  - **Primary files**: `internal/interp/evaluator/visitor_declarations.go` (record decl eval), plus any member-lookup code paths needed for `TRecord.Const` / `TRecord.ClassVar`
  - **Goal**: bring runtime behavior for record declarations in line with DWScript expectations (fields + const/class var + methods/properties).
  - **Subtasks**:
    - [x] **5.3.4.1** Audit current `VisitRecordDecl` behavior vs failing fixture(s) / missing runtime behavior (30m)
    - [x] **5.3.4.2** Record constants: evaluate sequentially with earlier record constants visible to later constant initializers (1h)
    - [x] **5.3.4.3** Record “static” members: ensure `TRecord.<Const|ClassVar|ClassMethod>` lookup is supported consistently and case-insensitively (1h)
    - [x] **5.3.4.4** Record fields metadata: ensure all fields are registered with correct case-insensitive mapping and are available for typed record literals and field access (30m)
    - [x] **5.3.4.5** Add/extend interpreter tests covering record declaration features (1h)
      - constants (including const depends-on-const)
      - class var initialization and access
      - method/class method dispatch (if already supported for records)
      - property read/write bindings (if already supported for records)
      - case-insensitivity checks (e.g. `tpoint.origin` vs `TPoint.Origin`)
  - **Acceptance criteria**:
    - New/updated tests pass (add targeted coverage in `internal/interp/*_test.go`)
    - At least one fixture previously failing due to record declaration behavior now passes (or is explicitly re-scoped to a later phase if it requires missing features outside Phase 5)

- [x] **5.3.5** Support class constant expressions in field initializers (4h)
  - `class_var_init_test.go:55`
  - Allow `ClassName.ConstantName` in initializers

---

### 5.4 Var Parameter and Reference Handling

**Goal**: Fix incomplete var/by-ref parameter support in records.

**Status**: 📋 Planned | **Effort**: 2-3 days | **Risk**: Medium

**Current Issues**:

- `functions_records.go:151`: By-ref parameters not properly implemented
- `functions_records.go:240`: Copy-back semantics incomplete

**Tasks**:

- [ ] **5.4.1** Implement proper by-ref support for record parameters (6h)
  - Create reference wrapper for record fields
  - Ensure mutations propagate correctly

- [ ] **5.4.2** Implement copy-back semantics for var parameters (4h)
  - Track original locations
  - Copy modified values back after call

- [ ] **5.4.3** Add tests for record var parameters (2h)

---

### 5.5 Unit Symbol Table Enhancement

**Goal**: Improve unit isolation and symbol resolution.

**Status**: 📋 Planned | **Effort**: 2-3 days | **Risk**: Low

**Current Issues**:
- `unit_loader.go:335`: Units don't maintain their own symbol tables
- `unit_loader.go:340`: Function-to-unit mapping not verified

**Tasks**:

- [ ] **5.5.1** Design per-unit symbol table structure (2h)
  - Each unit owns its symbols
  - Cross-unit references resolved through imports

- [ ] **5.5.2** Implement unit-scoped symbol tables (6h)
  - Modify UnitLoader to create isolated tables
  - Update function resolution to check unit ownership

- [ ] **5.5.3** Add verification for function-to-unit mapping (2h)

---

### 5.6 Miscellaneous Cleanup

**Goal**: Address remaining small TODOs.

**Status**: 📋 Planned | **Effort**: 1-2 days | **Risk**: Low

**Tasks**:

- [ ] **5.6.1** Fix JSON nil handling (2h)
  - `evaluator/json_helpers.go:182`
  - Return proper null/nil JSON value

- [ ] **5.6.2** Implement static array helper fallback (2h)
  - `evaluator/helper_methods.go:67`
  - Try dynamic array equivalent for static arrays

- [ ] **5.6.3** Complete function pointer type construction (2h)
  - `helpers_conversion.go:102`
  - Build proper type metadata for function pointers

- [ ] **5.6.4** Review and update expressions_basic.go:227 (2h)
  - Determine if "simplified implementation" is sufficient
  - Document or complete as needed

---

### Phase 5 Effort Summary

| Task | Effort | Priority |
|------|--------|----------|
| 5.1 Circular Import Workarounds | 3-4 days | High |
| 5.2 Incomplete Migration | 2-3 days | Medium |
| 5.3 Feature Implementation Gaps | 1 week | High |
| 5.4 Var Parameter Handling | 2-3 days | Medium |
| 5.5 Unit Symbol Tables | 2-3 days | Low |
| 5.6 Miscellaneous Cleanup | 1-2 days | Low |

**Total Estimated Effort**: 2-3 weeks

**Recommended Order**: 5.1 → 5.3 → 5.2 → 5.4 → 5.6 → 5.5

**Note**: These tasks can be done incrementally alongside other work. Prioritize 5.1 and 5.3 as they likely impact fixture test pass rates.

---

## Phase 6: Semantic Analyzer Improvements ✅ **COMPLETED**

**Status**: Complete | **Fixture Tests**: 386/1,227 passing (31.5%)

**Goal**: Improve maintainability and prepare for LSP support without over-engineering.

**What Was Done**:

1. **Removed Experimental Multi-Pass Architecture** (Task 6.1)
   - Deleted 8,062 lines of unused experimental code (~153KB)
   - Removed `declaration_pass.go`, `type_resolution_pass.go`, `validation_pass.go`, `contract_pass.go`
   - Experiment added 17% MORE code instead of simplifying
   - Original analyzer already handled forward declarations and circular inheritance correctly

2. **Enhanced Symbol Table for LSP Support** (Task 6.2)
   - Added 5 new fields: `DeclPosition`, `Usages`, `Documentation`, `IsDeprecated`, `DeprecationMessage`
   - Implemented 4 LSP query methods: `FindDefinition()`, `FindReferences()`, `UnusedSymbols()`, `RecordUsage()`
   - Updated 39 call sites across 14 files to pass position information
   - Created comprehensive test suite with 16 test functions
   - See [docs/task-6.2-summary.md](docs/task-6.2-summary.md)

3. **Standardized Error Messages** (Task 6.3)
   - Standardized 4 major error categories (~35 messages) using DWScript-compatible format:
     - Undefined symbols: `"Unknown name \"X\""`
     - Abstract class errors: `"Trying to create an instance of an abstract class"`
     - Duplicate declarations: `"Name \"X\" already exists"`
     - Overload resolution: `"There is no overloaded version of \"X\" that can be called with these arguments"`
   - Added 10 helper functions to `internal/errors/errors.go`
   - Updated 13 semantic analyzer files and 11 test files
   - Deliberately deferred 531 builtin function errors (already internally consistent)
   - See [docs/task-6.3-summary.md](docs/task-6.3-summary.md)

4. **Shared Builtin Registry** (Task 6.4)
   - Moved `internal/interp/builtins/` (47 files) → `internal/builtins/` for shared access
   - Updated 12 interpreter import paths
   - Semantic analyzer references registry for documentation but keeps detailed analyzers for validation
   - Registry not used for arity checking (detailed analyzers provide better error messages)
   - All tests pass with no regressions

**Key Decision**: Multi-pass architecture was **deprecated** after analysis showed it duplicated existing functionality. The original single-pass analyzer with `IsForward` flags and circular inheritance tracking already solved the problems the multi-pass attempted to address.

---

## Stage 7: Support Object-Oriented Features (Classes, Interfaces, Methods) ✅ **COMPLETED**

- Extended the type system and AST with class/interface nodes, constructors/destructors, member access, `Self`, `NewExpression`, and external declarations (see docs/stage7-summary.md).
- Parser handles class/interface declarations, inheritance chains, interface lists, constructors, member access, and method calls with comprehensive unit tests and fixtures.
- Added runtime class metadata plus interpreter support for object creation, field storage, method dispatch, constructors, destructors, and interface casting with ~98% targeted coverage.
- Semantic analysis validates class/interface hierarchies, method signatures, interface fulfillment, and external contracts while integrating with the existing symbol/type infrastructure.
- Documentation in docs/stage7-summary.md, docs/stage7-complete.md, docs/delphi-to-go-mapping.md, and docs/interfaces-guide.md captures the architecture, and CLI/integration suites ensure DWScript parity.

## Stage 8: Additional DWScript Features and Polishing ✅ **IN PROGRESS**

- Extended the expression/type system with DWScript-accurate operator overloading (global + class operators, coercions, analyzer enforcement) and wired the interpreter/CLI with matching fixtures in `testdata/operators` and docs in `docs/operators.md`.
- Landed the full property stack (field/method/auto/default metadata, parser, semantic validation, interpreter lowering, CLI coverage) so OO code can use DWScript-style properties; deferred indexed/expr variants are tracked separately.
- Delivered composite type parity: enums, records, sets, static/dynamic arrays, and assignment/indexing semantics now mirror DWScript with dedicated analyzers, runtime values, exhaustive unit/integration suites, and design notes captured in docs/enums.md plus related status writeups.
- Upgraded the runtime to support break/continue/exit statements, DWScript's `new` keyword, rich exception handling (try/except/finally, raise, built-in exception classes), and CLI smoke tests that mirror upstream fixtures.
- Backfilled fixtures, docs, and CLI suites for every feature shipped in this phase (properties, enums, arrays, exceptions, etc.), keeping coverage high and mapping each ported DWScript test in `testdata/properties/REFERENCE_TESTS.md`.

---

## Task 9.6: Enhance Class Constants with Field Initialization

**Goal**: Support field initialization from class constants in class body.

**Estimate**: 6-8 hours (0.75-1 day)

**Status**: DONE

**Impact**: Unlocks 7 failing tests in SimpleScripts (complements Task 9.2)

**Priority**: P0 - CRITICAL (Required for class patterns)

**Description**: Task 9.2 added basic class constant support, but DWScript also allows initializing fields directly from constants using syntax like `FField := Value;` inside the class body. This requires:
- Parsing field initialization syntax (currently fails with parse errors)
- Semantic analysis to resolve constant references in initializers
- Runtime support for field initialization during object creation

**Failing Tests** (7 total):
- class_const2 (semantic issues with const resolution)
- class_const3 (missing hints for case mismatch)
- class_const4 (parse error for field initialization syntax)
- class_const_as_prop (output mismatch)
- class_init (parse error)
- const_block (parse error)
- enum_element_deprecated

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P0, Section 3

**Example**:
```pascal
type TObj = class
  const Value = 5;
  FField := Value;  // Initialize field from constant
end;
```

**Subtasks**:

### 9.6.1 Parse Field Initialization Syntax

**Goal**: Update parser to handle `FField := Value;` syntax in class bodies.

**Estimate**: 2-3 hours

**Implementation**:

1. Modify class body parser to recognize `:=` after identifier
2. Create field with initialization expression in AST
3. Distinguish from method declarations and properties

**Files to Modify**:
- `internal/parser/parser_classes.go` (parse field initialization)
- `pkg/ast/declarations.go` (add Initializer field to FieldDeclaration)

### 9.6.2 Semantic Analysis for Field Initializers

**Goal**: Type check and resolve constant references in field initializers.

**Estimate**: 2-3 hours

**Implementation**:

1. During class declaration analysis, analyze field initializer expressions
2. Resolve references to class constants
3. Validate initializer types match field types
4. Report errors for invalid initializers

**Files to Modify**:
- `internal/semantic/analyze_classes_decl.go` (analyze field initializers)

### 9.6.3 Runtime Field Initialization

**Goal**: Execute field initializers during object creation.

**Estimate**: 2 hours

**Implementation**:
1. During object construction, evaluate field initializers
2. Apply initialized values to new instances
3. Handle initialization order (constants before field inits)

**Files to Modify**:
- `internal/interp/objects_creation.go` (execute field initializers)

**Success Criteria**:
- All 7 class const tests pass
- `FField := Value;` syntax parses correctly
- Fields are initialized from constants during object creation
- `new TObj.FField` returns the constant value

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/class_const4
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/class_init
```

---

## Task 9.12: Implement Record Advanced Features

**Goal**: Finish record advanced features (field/class var/const, methods, properties, nested records) and get remaining fixtures green.

**Status**: IN PROGRESS

**Done (summary)**: Parsing/semantic/runtime support for field initializers, record constants/class vars, properties, and improved record methods is landed.

**Current failing fixtures**: `record_method2` (semantic helper requirement on parameterless record-returning function members); `record_method5` (runtime increments not reflected). Other `record_*` fixtures currently pass—reconfirm with a full run.

**TODO (comprehensive)**:
- [ ] Remove helper requirement for member access on parameterless record-returning functions (fixes `record_method2`).
- [ ] Ensure record implicit Self calls propagate mutations/copy semantics so method-side increments persist (fixes `record_method5`).
- [ ] Re-run full `TestDWScriptFixtures/SimpleScripts` to refresh the `record_*` failing list.
- [ ] Regression sweep: record properties/class vars/consts after the above fixes.
- [ ] Update fixture status docs once all `record_*` fixtures pass.

---

## Task 9.13: Implement Property Advanced Features

**Goal**: Add indexed properties, array-typed properties, and property promotion/reintroduce.

**Estimate**: 8-12 hours (1-1.5 days)

**Status**: NOT STARTED

**Impact**: Unlocks 9 failing tests in SimpleScripts

**Priority**: P1 - IMPORTANT (OOP encapsulation)

**Description**: Properties currently have basic getter/setter support, but DWScript includes advanced features like indexed properties (e.g., `Items[i]`), array-typed properties, property promotion from parent classes, and the `reintroduce` keyword for shadowing parent properties.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P1, Section 9

**Failing Tests** (9 total):
- class_var_as_prop
- index_property
- property_call
- property_index
- property_of_as
- property_promotion
- property_reintroduce
- property_sub_default
- property_type_array

**Example**:
```pascal
type
  TList = class
    private
      FData: array of Integer;
    public
      // Indexed property (default)
      property Items[Index: Integer]: Integer
        read GetItem write SetItem; default;

      // Array-typed property
      property Data: array of Integer read FData;

    function GetItem(Index: Integer): Integer;
    begin
      Result := FData[Index];
    end;

    procedure SetItem(Index: Integer; Value: Integer);
    begin
      FData[Index] := Value;
    end;
  end;

var list: TList;
begin
  list := TList.Create;
  list[0] := 42;        // Uses default indexed property
  PrintLn(list[0]);
end;
```

**Complexity**: Medium-High - Multiple property enhancement features

**Subtasks**:

### 9.13.1 Parse Indexed Properties

**Goal**: Support `property Name[Index: Type]: Type` syntax.

**Estimate**: 2-3 hours

**Implementation**:
1. Extend property parsing to handle `[` parameters `]` after property name
2. Parse multiple index parameters
3. Parse `default` keyword for default indexed property
4. Store index parameters in PropertyDecl

**Files to Modify**:
- `internal/parser/properties.go` (indexed property parsing)
- `pkg/ast/properties.go` (IndexParams field on PropertyDecl)

### 9.13.2 Parse Array-Typed Properties

**Goal**: Support properties with array types.

**Estimate**: 1 hour

**Implementation**:
1. Allow array types in property type declarations
2. Handle getter/setter with array return/parameter types
3. Parse array property access syntax

**Files to Modify**:
- `internal/parser/properties.go` (array type properties)

### 9.13.3 Semantic Analysis for Indexed Properties

**Goal**: Type check indexed property access and assignments.

**Estimate**: 3-4 hours

**Implementation**:
1. Resolve indexed property access: `obj.Prop[index]`
2. Check index parameter types match declaration
3. Type check getter/setter signatures with index parameters
4. Default indexed property allows `obj[index]` syntax
5. Array-typed properties type check correctly

**Files to Modify**:
- `internal/semantic/analyze_properties.go` (indexed property analysis)
- `internal/semantic/analyze_expressions.go` (property access with indices)

### 9.13.4 Runtime Indexed Property Access

**Goal**: Execute indexed property getters/setters with indices.

**Estimate**: 2-3 hours

**Implementation**:
1. Evaluate index expressions
2. Call getter with index parameters
3. Call setter with index + value parameters
4. Default indexed property via `[]` operator
5. Array-typed property returns array value

**Files to Modify**:
- `internal/interp/properties.go` (indexed property execution)
- `internal/interp/objects_properties.go` (property access)

### 9.13.5 Property Promotion and Reintroduce

**Goal**: Support `reintroduce` and property promotion from parent.

**Estimate**: 2-3 hours

**Implementation**:
1. Parse `reintroduce` keyword on properties
2. Semantic analysis: allow child class to shadow parent property with reintroduce
3. Property promotion: access parent property via child class
4. Runtime: respect override/reintroduce semantics

**Files to Modify**:
- `internal/parser/properties.go` (reintroduce keyword)
- `internal/semantic/analyze_properties.go` (promotion/reintroduce)
- `internal/interp/properties.go` (runtime property lookup)

**Success Criteria**:
- Indexed properties parse and work correctly
- Array-typed properties supported
- Default indexed property enables `obj[i]` syntax
- Property promotion from parent classes works
- `reintroduce` keyword allows property shadowing
- All 9 property advanced feature tests pass

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/property
go test -v ./internal/semantic -run TestPropertyAdvanced
```

---

## Task 9.14: Fix Inheritance and Virtual Methods Issues

**Goal**: Correct override validation, inherited keyword, reintroduce, virtual constructors, and VMT architecture.

**Estimate**: 18-26 hours (2.5-3.5 days)

**Status**: NOT STARTED

**Impact**: Unlocks 14 failing tests in SimpleScripts + fixes `reintroduce virtual` semantics

**Priority**: P1 - CRITICAL (OOP polymorphism + architecture fix)

**Description**: Current inheritance and virtual method implementation has several issues: improper override validation, incomplete `inherited` keyword support (especially in constructors), missing `reintroduce` keyword, and incorrect virtual constructor behavior. These are critical for proper OOP polymorphism.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P1, Section 10

**Failing Tests** (14 total):
- class_forward
- class_parent
- clear_ref_in_constructor_assignment
- clear_ref_in_destructor
- destroy
- free_destroy
- inherited1, inherited2
- inherited_constructor
- oop
- override_deep
- reintroduce
- reintroduce_virtual
- virtual_constructor, virtual_constructor2

**Example**:
```pascal
type
  TBase = class
    constructor Create; virtual;
    procedure DoSomething; virtual;
  end;

  TDerived = class(TBase)
    constructor Create; override;  // Override virtual constructor
    procedure DoSomething; override;
  end;

constructor TBase.Create;
begin
  PrintLn('TBase.Create');
end;

constructor TDerived.Create;
begin
  inherited;  // Call parent constructor
  PrintLn('TDerived.Create');
end;

procedure TDerived.DoSomething;
begin
  inherited DoSomething;  // Call parent method
  PrintLn('TDerived.DoSomething');
end;
```

**Complexity**: Medium-High - Requires fixes across semantic and runtime

**Subtasks**:

### 9.14.1 Fix Override Validation

**Goal**: Properly validate override keyword matches parent virtual method.

**Estimate**: 2-3 hours

**Implementation**:
1. Check parent class has method with same name
2. Verify parent method is declared virtual or override
3. Validate signature matches (same params and return type)
4. Report error if override used without virtual parent
5. Deep override chains (override of override)

**Files to Modify**:
- `internal/semantic/analyze_classes.go` (override validation)

### 9.14.2 Implement Inherited Keyword Fully

**Goal**: Support `inherited` calls in all method types.

**Estimate**: 3-4 hours

**Implementation**:
1. Parse `inherited;` (call parent's same method)
2. Parse `inherited MethodName(args);` (call specific parent method)
3. Semantic analysis: resolve inherited to parent class method
4. In constructors, `inherited` calls parent constructor
5. Type check inherited calls

**Files to Modify**:
- `internal/parser/expressions.go` (parse inherited)
- `pkg/ast/expressions.go` (InheritedExpression node)
- `internal/semantic/analyze_classes.go` (inherited resolution)

### 9.14.3 Runtime Inherited Call Execution

**Goal**: Execute inherited calls correctly.

**Estimate**: 2-3 hours

**Implementation**:
1. Look up parent class method
2. Call parent method with current object (Self)
3. In constructors, chain to parent constructor before child initialization
4. In destructors, call parent destructor after child cleanup

**Files to Modify**:
- `internal/interp/objects_methods.go` (inherited calls)
- `internal/interp/objects_creation.go` (constructor chaining)
- `internal/interp/objects_destruction.go` (destructor chaining)

### 9.14.4 Implement Reintroduce Keyword

**Goal**: Support reintroduce for shadowing parent members.

**Estimate**: 2 hours

**Implementation**:
1. Parse `reintroduce` keyword on methods
2. Semantic analysis: allow method to shadow parent method without override
3. Warning if shadowing without reintroduce
4. Runtime: child method hides parent method

**Files to Modify**:
- `internal/parser/functions.go` (parse reintroduce)
- `internal/semantic/analyze_classes.go` (reintroduce validation)

### 9.14.5 Fix Virtual Constructor Behavior

**Goal**: Correct virtual constructor dispatch and initialization.

**Estimate**: 2-3 hours

**Implementation**:
1. Virtual constructors can be overridden
2. Calling Create on class reference dispatches to correct constructor
3. Constructor chaining with inherited works correctly
4. Virtual destructor support (Free, Destroy)

**Files to Modify**:
- `internal/interp/objects_creation.go` (virtual constructor dispatch)
- `internal/semantic/analyze_classes.go` (virtual constructor validation)

### 9.14.6 Refactor VMT to Array-Based Architecture

**Goal**: Replace map-based VMT with array-based VMT to support `reintroduce virtual` semantics.

**Estimate**: 8-12 hours (1-1.5 days)

**Priority**: P1 - CRITICAL for correct OOP semantics

**Context**: During Phase 3.5.8 implementation, discovered that current VMT uses `map[string]*VirtualMethodEntry` which cannot support `reintroduce virtual`. DWScript reference uses array with indices: `FVirtualMethodTable []TMethodSymbol`.

**Problem**:

- Current: `map[signature] → VirtualMethodEntry` - only ONE entry per signature
- Needed: Array where each virtual method gets unique index, `reintroduce virtual` gets NEW index
- See detailed analysis: `docs/phase3.5.8-findings.md`

**Implementation**:

1. Change `ClassInfo.VirtualMethodTable` from `map[string]*VirtualMethodEntry` to `[]*VirtualMethodEntry`
2. Add `VMTIndex int` field to method declarations during semantic analysis
3. Assign VMT indices:
   - First `virtual` → allocate new VMT index
   - `override` → use parent's VMT index for same method
   - `reintroduce virtual` → allocate NEW VMT index (starts new chain)
4. Update method dispatch to use VMT index instead of map lookup
5. Update `buildVirtualMethodTable()` to build array instead of map
6. Fix `reintroduce_virtual.pas` test case

**Files to Modify**:

- `internal/interp/class.go` (ClassInfo.VirtualMethodTable type, buildVirtualMethodTable)
- `internal/interp/objects_methods.go` (method dispatch - use VMT index)
- `internal/semantic/analyze_classes.go` (assign VMT indices)
- `pkg/ast/functions.go` (add VMTIndex field)

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/reintroduce_virtual
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/override_deep
```

**Success Criteria**:
- Override validation checks parent method is virtual
- `inherited` works in methods, constructors, destructors
- `inherited MethodName` syntax works
- `reintroduce` allows shadowing without override
- `reintroduce virtual` creates new VMT entry (not overwrites)
- Virtual constructors dispatch correctly
- Constructor/destructor chaining works
- All 14 inheritance/virtual method tests pass
- `reintroduce_virtual.pas` fixture test passes

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/inherited
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/virtual_constructor
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/override
go test -v ./internal/semantic -run TestInheritance
```

---

## Task 9.15: Implement Enum Advanced Features

**Goal**: Add enum boolean operations, bounds (Low/High), EnumByName, flags, scoped enums, and deprecation.

**Estimate**: 8-12 hours (1-1.5 days)

**Status**: IN PROGRESS (9/17 enum tests passing, boolean ops & implicit conversion complete)

**Impact**: Unlocks 12 failing tests in SimpleScripts

**Priority**: P1 - IMPORTANT (Type system completeness)

**Description**: Enumerations currently have basic support, but DWScript includes advanced features like boolean operations on enums, bounds checking (Low/High), EnumByName function for string-to-enum conversion, enum flags (sets), scoped enums, and enum element deprecation.

**Reference**: See `FIXTURE_FAILURES_ANALYSIS.md` - Priority P1, Section 11

**Completed** (2 total):
- ✅ enum_bool_op (Tasks 9.15.7 and 9.15.8 - boolean ops and implicit conversion)

**Failing Tests** (11 remaining):

- aliased_enum
- enum_bounds
- enum_byname
- enum_casts
- enum_element_deprecated
- enum_flags1
- enum_scoped
- enum_to_integer
- enumerations
- enumerations_names
- enumerations_qualifiednames

**Example**:
```pascal
type
  TMyEnum = (meA, meB, meC);
  TScopedEnum = (seX, seY) scoped;  // Elements accessed as TScopedEnum.seX
  TFlags = (flRead, flWrite, flExecute) flags;  // Bit flags

var
  e: TMyEnum;
  flags: TFlags;
begin
  // Bounds
  for e := Low(TMyEnum) to High(TMyEnum) do
    PrintLn(e);

  // EnumByName
  e := EnumByName<TMyEnum>('meB');

  // Boolean operations
  if e in [meA, meB] then
    PrintLn('A or B');

  // Flags (set operations)
  flags := [flRead, flWrite];
  if flRead in flags then
    PrintLn('Readable');

  // Scoped enum
  var se := TScopedEnum.seX;  // Must use type prefix
end;
```

**Complexity**: Medium - Multiple enum enhancements

**Subtasks**:

### 9.15.2 Enum Boolean Operations

**Goal**: Support boolean operators with enum operands.

**Estimate**: 2-3 hours

**Implementation**:
1. `in` operator: check if enum value in set of values
2. Set operations on enum values: `[meA, meB]`
3. Semantic analysis for enum set expressions
4. Runtime evaluation of enum in set

**Files to Modify**:
- `internal/semantic/analyze_expressions.go` (enum boolean ops)
- `internal/interp/expressions_operators.go` (enum in operator)

### 9.15.7 Enum Boolean Operators ✅ COMPLETE

**Goal**: Support boolean/bitwise operators on enum values (for flags).

**Estimate**: 2-3 hours

**Status**: ✅ COMPLETE (2025-11-22) - See Task 1.6

**Implementation**:
1. ✅ Allow `or`, `and`, `xor` operators on enum operands (especially flags)
2. ✅ Semantic analysis: check enum types are compatible for operators
3. ✅ Runtime: evaluate bitwise operations on enum ordinal values
4. ✅ Return enum type for result (not integer)

**Tests**: enum_bool_op ✅ PASSING (AST interpreter)

**Example**:
```pascal
type TFlags = flags (flRead, flWrite, flExecute);
var f1 := TFlags.flRead or TFlags.flWrite;  // Result: 3 (as TFlags)
var f2 := TFlags.flWrite and TFlags.flRead; // Result: 0 (as TFlags)
```

**Files Modified**:
- ✅ `internal/semantic/analyze_expr_operators.go` (allow enum operands for or/and/xor)
- ✅ `internal/interp/expressions_binary.go` (evaluate enum boolean ops)

**Note**: Bytecode VM support pending (see Task 14.16)

### 9.15.8 Implicit Enum-to-Integer Conversion ✅ COMPLETE

**Goal**: Allow implicit conversion from enum to Integer in function calls.

**Estimate**: 1-2 hours

**Status**: ✅ COMPLETE (2025-11-22) - See Task 1.6

**Implementation**:
1. ✅ When calling `func(i: Integer)` with enum argument, auto-convert
2. ✅ Semantic analysis: allow enum-to-integer implicit conversion
3. ✅ Runtime: automatically get ordinal value when passing enum to Integer param

**Tests**: enum_bool_op (PrintInt calls) ✅ PASSING (AST interpreter)

**Example**:
```pascal
procedure PrintInt(i: Integer);
type TEnum = flags (Alpha, Beta);
PrintInt(TEnum.Alpha);  // Should pass 1, not fail with type mismatch
```

**Files to Modify**:
- `internal/semantic/analyze_expressions.go` (implicit conversion in canAssign)
- `internal/interp/expressions.go` (auto-convert enum to integer)

### 9.15.9 Constant Expressions in Enum Values

**Goal**: Support constant expressions (like Ord('A')) in enum value assignments.

**Estimate**: 2-3 hours

**Status**: TODO

**Implementation**:
1. Parse constant expressions in enum value positions
2. Evaluate constant expressions at parse/semantic time
3. Support function calls like `Ord('A')`, `Chr(65)` in enum values
4. Constant folding for arithmetic expressions

**Failing Tests**: enum_to_integer

**Example**:
```pascal
type TEnumAlpha = (eAlpha = Ord('A'), eBeta, eGamma);
// eAlpha should have value 65 (ASCII 'A')
```

**Files to Modify**:
- `internal/parser/enums.go` (parse expressions instead of just integers)
- `internal/semantic/analyze_enums.go` (evaluate constant expressions)
- Add constant expression evaluator

### 9.15.10 Enum Instance .Value Property ✓ DONE

**Goal**: Support .Value property on enum instances to get ordinal value.

**Estimate**: 1-2 hours

**Status**: Complete - .Value property fully working

**Implementation**:
1. Add .Value property to enum values (returns Integer)
2. Semantic analysis: recognize .Value on enum types
3. Runtime: return ordinal value when accessing .Value
4. Also support .ToString() method

**Failing Tests**: enum_to_integer

**Example**:
```pascal
var e: TEnum := eOne;
PrintLn(e.Value);           // Prints ordinal value
PrintLn(eOne.Value);        // Direct access also works
PrintLn(TEnum.eTwo.Value);  // Qualified access works
```

**Files to Modify**:
- `internal/semantic/analyze_expressions.go` (add .Value property to enum helper)
- `internal/interp/objects_hierarchy.go` (evaluate .Value member access)

### 9.15.11 Additional Enum Features

**Goal**: Handle remaining edge cases and test failures.

**Estimate**: 2-3 hours

**Status**: TODO

**Implementation**:
1. Aliased enums (type alias to enum type)
2. Enum deprecation warnings
3. Enum names and qualified names functions
4. For-in loops over enum ranges

**Failing Tests**: aliased_enum, enum_element_deprecated, enumerations, enumerations2, enumerations_names, enumerations_qualifiednames

**Files to Modify**:
- Various files depending on specific feature

**Success Criteria**:
- ✓ Scoped enums enforce qualified access (MyEnum.a)
- ✓ Flags enums use power-of-2 values
- ✓ Low/High properties and methods return enum bounds
- ✓ ByName() method converts string to enum
- ✓ .Value property on enum instances returns ordinal value
- ✓ Enum-to-integer and integer-to-enum casts - DONE
- ⚠ Enum boolean operations (or, and, xor) - TODO
- ⚠ Implicit enum-to-integer conversion - TODO
- ⚠ Constant expressions in enum values - TODO
- **Progress**: 8/17 tests passing (enum_scoped, enum_flags1, enum_bounds, enum_byname, enum_value, enum_qualified_access, enum_ord, enum_forin)

**Completed Work**:
- ✅ 9.15.7: Boolean operators on enum types (or, and, xor for flags) - Task 1.6
- ✅ 9.15.8: Implicit conversion from enum to Integer in function calls - Task 1.6

**Remaining Work** (see subtasks 9.15.9-9.15.11 for details):
- 9.15.9: Constant expressions in enum value assignments (e.g., Ord('A'))
- 9.15.11: Additional edge cases (aliased enums, deprecation, names)
- Bytecode VM support (see Task 14.16)

**Testing**:
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/enum
go test -v ./internal/semantic -run TestEnumAdvanced
```

---

## Task 9.16 Introduce Base Structs for AST Nodes

**Goal**: Eliminate code duplication by introducing base structs for common node fields and behavior.

**Estimate**: 8-10 hours (1-1.5 days)

**Status**: IN PROGRESS

**Impact**: Reduces AST codebase by ~30% (~500 lines), eliminates duplicate boilerplate across 50+ node types

**Description**: Currently, every AST node type duplicates identical implementations for `Pos()`, `End()`, `TokenLiteral()`, `GetType()`, and `SetType()` methods. This creates ~500 lines of repetitive code that is error-prone to maintain. By introducing base structs with embedding, we can eliminate this duplication while maintaining the same interface.

**Current Problem**:

```go
// Repeated ~50 times across different node types
type IntegerLiteral struct {
    Type   *TypeAnnotation
    Token  token.Token
    Value  int64
    EndPos token.Position
}

func (il *IntegerLiteral) Pos() token.Position  { return il.Token.Pos }
func (il *IntegerLiteral) End() token.Position {
    if il.EndPos.Line != 0 {
        return il.EndPos
    }
    pos := il.Token.Pos
    pos.Column += len(il.Token.Literal)
    pos.Offset += len(il.Token.Literal)
    return pos
}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) GetType() *TypeAnnotation    { return il.Type }
func (il *IntegerLiteral) SetType(typ *TypeAnnotation) { il.Type = typ }
```

**Strategy**: Create base structs using Go embedding to share common fields and method implementations:

1. **BaseNode**: Common fields (Token, EndPos) and methods (Pos, End, TokenLiteral)
2. **TypedExpressionBase**: Extends BaseNode with Type field and GetType/SetType methods
3. Refactor all node types to embed appropriate base struct
4. Remove duplicate method implementations

**Complexity**: Medium - Requires systematic refactoring of all AST node types across 25 files (~5,500 lines)

**Subtasks**:

- [x] 9.16.1 Design base struct hierarchy
  - [x] Create `BaseNode` struct with Token, EndPos fields
  - [x] Create `TypedExpressionBase` struct embedding BaseNode with Type field
  - [x] Implement common methods once on base structs
  - [x] Document design decisions and usage patterns
  - [x] Add `pkg/ast/base.go`

- [x] 9.16.2 Refactor literal expression nodes (Identifier, IntegerLiteral, FloatLiteral, StringLiteral, BooleanLiteral, CharLiteral, NilLiteral)
  - [x] Embed `TypedExpressionBase` into Identifier and adjust parser/tests
  - [x] Embed `TypedExpressionBase` into numeric/string/char/boolean literal structs
  - [x] Embed `TypedExpressionBase` into NilLiteral
  - [x] Remove redundant `TokenLiteral/Pos/End/GetType` methods
  - [x] Update parser/semantic/interpreter tests that construct these literals
  - [x] Updated all parser files (12 files, 37 instances)
  - [x] Updated all test files in internal/ast (17 files, 446 instances)
  - [x] Updated all test files in internal/bytecode (6 files, 100+ instances)
  - [x] Updated all test files in internal/interp (6 files, 85+ instances)
  - [x] Updated all test files in internal/semantic (3 files, 55+ instances)
  - All literal expression nodes now use TypedExpressionBase successfully

- [x] 9.16.3 Refactor binary and unary expressions (BinaryExpression, UnaryExpression, GroupedExpression, RangeExpression)
  - [x] Embed `TypedExpressionBase` into BinaryExpression
  - [x] Embed `TypedExpressionBase` into UnaryExpression
  - [x] Embed `TypedExpressionBase` into GroupedExpression
  - [x] Embed `TypedExpressionBase` into RangeExpression
  - [x] Remove duplicate type/position helpers and verify parser/semantic behavior
  - [x] Updated parser files (expressions.go, arrays.go, control_flow.go, sets.go)
  - [x] Updated 17 test files across internal/ast, internal/bytecode, internal/semantic
  - [x] All tests pass successfully - removed ~120 lines of boilerplate

- [x] 9.16.4 Refactor statement nodes (ExpressionStatement, VarDeclStatement, AssignmentStatement, BlockStatement, IfStatement, WhileStatement, etc.)
  - [x] Identify all statement structs across `pkg/ast/statements.go`, `pkg/ast/control_flow.go`, and related files
  - [x] Embed `BaseNode` into expression statements/assignments/var decls (already done in previous tasks)
  - [x] Embed `BaseNode` into control-flow statements (if/while/for/try/case) (already done in previous tasks)
  - [x] Embed `BaseNode` into exception-related nodes: TryStatement, ExceptClause, ExceptionHandler, FinallyClause, RaiseStatement
  - [x] Embed `BaseNode` into ReturnStatement
  - [x] Remove redundant position/token helpers (TokenLiteral, Pos, End) from all refactored nodes
  - [x] Update parser code to construct nodes with BaseNode wrapper
  - [x] Update all test files in internal/bytecode (6 files, 30+ instances)
  - [x] All tests pass successfully - removed ~50 lines of boilerplate from statement nodes

- [x] 9.16.5 Refactor declaration nodes (ConstDecl, FunctionDecl, ClassDecl, InterfaceDecl, etc.)
  - [x] Embed BaseNode into HelperDecl
  - [x] Embed BaseNode into InterfaceDecl / InterfaceMethodDecl
  - [x] Embed BaseNode into ConstDecl
  - [x] Embed BaseNode into TypeDeclaration
  - [x] Embed BaseNode into FieldDecl
  - [x] Embed BaseNode into PropertyDecl
  - [x] Embed BaseNode into FunctionDecl / constructor nodes
  - [x] Embed BaseNode into ClassDecl / Class-related structs (`pkg/ast/classes.go`)
  - [x] Embed BaseNode into RecordDecl / RecordPropertyDecl / FieldInitializer / RecordLiteralExpression (`pkg/ast/records.go`)
  - [x] Embed BaseNode into OperatorDecl
  - [x] Embed BaseNode into EnumDecl (`pkg/ast/enums.go`)
  - [x] Embed BaseNode into ArrayDecl/SetDecl nodes (`pkg/ast/arrays.go`, `pkg/ast/sets.go`)
  - [x] Embed BaseNode into UnitDeclaration and UsesClause structures (`pkg/ast/unit.go`)
  - [x] Remove duplicate helper methods once all declaration structs embed the base
  - [x] Update all parser files to use BaseNode syntax in struct literals
  - [x] Update all test files to use BaseNode syntax
  - Files: `pkg/ast/declarations.go`, `pkg/ast/functions.go`, `pkg/ast/classes.go`, `pkg/ast/interfaces.go`, `pkg/ast/records.go`, `pkg/ast/enums.go`, `pkg/ast/operators.go`, `pkg/ast/arrays.go`, `pkg/ast/sets.go`, `pkg/ast/unit.go` (~200 lines reduced)
  - All declaration nodes now embed BaseNode, eliminating duplicate boilerplate code

- [x] 9.16.6 Refactor type-specific nodes (ArrayLiteralExpression, CallExpression, NewExpression, MemberAccessExpression, etc.)
  - [x] Refactored NewExpression to embed TypedExpressionBase
  - [x] Refactored MemberAccessExpression to embed TypedExpressionBase
  - [x] Refactored MethodCallExpression to embed TypedExpressionBase
  - [x] Refactored InheritedExpression to embed TypedExpressionBase
  - [x] Updated all parser files (internal/parser/classes.go, internal/parser/expressions.go)
  - [x] Updated all test files (internal/bytecode/vm_test.go, internal/bytecode/compiler_expressions_test.go)
  - [x] Updated interpreter files (internal/interp/objects_methods.go, internal/interp/objects_hierarchy.go, internal/interp/objects_instantiation.go, internal/interp/functions_calls.go)
  - Files: `pkg/ast/arrays.go`, `pkg/ast/classes.go`, `pkg/ast/functions.go` (~80 lines of boilerplate removed)

- [x] 9.16.7 Update parser to use base struct constructors
  - [x] Update parser sites already touched (helpers/interfaces/const/type/property/field)
  - [x] Sweep remaining parser files for struct literals using removed `Token` fields
  - All parser files have been updated to use TypedExpressionBase/BaseNode pattern
  - No helper constructors needed - the pattern is straightforward and consistent

- [x] 9.16.8 Update semantic analyzer and interpreter
  - [x] Updated const/type/property/helper-specific tests where embedding occurred
  - [x] Refactored SetLiteral to use TypedExpressionBase (removed ~40 lines of boilerplate)
  - [x] Refactored AddressOfExpression to use TypedExpressionBase (removed ~10 lines of boilerplate)
  - [x] Refactored LambdaExpression to use TypedExpressionBase (removed ~30 lines of boilerplate)
  - [x] Updated all parser/semantic/interpreter/bytecode files for these changes
  - [x] All tests passing for modified types (SetLiteral, AddressOfExpression, LambdaExpression)

- [ ] 9.16.9 Run comprehensive test suite
  - [ ] `go test ./pkg/ast`
  - [ ] `go test ./internal/parser`
  - [ ] `go test ./internal/semantic`
  - [ ] `go test ./internal/interp`
  - [ ] `go test ./internal/bytecode`
  - [ ] Fixture / CLI integration suite

**Files Modified**:

- `pkg/ast/base.go` (new file ~100 lines)
- `pkg/ast/ast.go` (~300 lines reduced to ~150)
- `pkg/ast/statements.go` (~316 lines reduced to ~200)
- `pkg/ast/control_flow.go` (~200 lines reduced to ~120)
- `pkg/ast/declarations.go` (~150 lines reduced to ~80)
- `pkg/ast/functions.go` (~245 lines reduced to ~150)
- `pkg/ast/classes.go` (~400 lines reduced to ~250)
- `pkg/ast/interfaces.go` (~100 lines reduced to ~60)
- `pkg/ast/arrays.go` (~200 lines reduced to ~120)
- `pkg/ast/enums.go` (~100 lines reduced to ~60)
- `pkg/ast/records.go` (~150 lines reduced to ~90)
- `pkg/ast/sets.go` (~100 lines reduced to ~60)
- `pkg/ast/properties.go` (~120 lines reduced to ~70)
- `pkg/ast/operators.go` (~80 lines reduced to ~50)
- `pkg/ast/exceptions.go` (~100 lines reduced to ~60)
- `pkg/ast/lambda.go` (~80 lines reduced to ~50)
- `pkg/ast/helper.go` (~168 lines reduced to ~100)
- Plus updates to parser, semantic analyzer, and interpreter

**Acceptance Criteria**:
- All AST nodes embed either BaseNode or TypedExpressionBase
- No duplicate Pos/End/TokenLiteral/GetType/SetType implementations
- All existing tests pass (100% backward compatibility)
- Codebase reduced by ~500 lines
- AST package is more maintainable with centralized common behavior
- Documentation explains base struct usage and when to embed each type

**Benefits**:
- 30% reduction in AST code (~500 lines eliminated)
- Single source of truth for common behavior
- Easier to add new node types (less boilerplate)
- Reduced chance of copy-paste errors
- Consistent behavior across all nodes

---

- [ ] 9.18 Separate Type Metadata from AST

**Goal**: Move type information from AST nodes to a separate metadata table, making the AST immutable and reusable.

**Estimate**: 6-8 hours (1 day)

**Status**: IN PROGRESS

**Impact**: Cleaner separation of parsing vs semantic analysis, reduced memory usage, enables multiple concurrent analyses

**Description**: Currently, every expression node carries a `Type *TypeAnnotation` field that is nil during parsing and populated during semantic analysis. This couples the AST to the semantic analyzer and wastes memory (~16 bytes per node). Moving type information to a separate side table improves separation of concerns and enables the AST to be analyzed multiple times with different contexts.

**Current Problem**:

```go
type IntegerLiteral struct {
    Type   *TypeAnnotation  // nil until semantic analysis
    Token  token.Token
    Value  int64
    EndPos token.Position
}
```

**Strategy**: Create a separate metadata table that maps AST nodes to their semantic information:

1. Remove Type field from AST nodes
2. Create SemanticInfo struct with type/symbol maps
3. Semantic analyzer populates SemanticInfo instead of modifying AST
4. Provide accessor methods for type information

**Complexity**: Medium - Requires refactoring semantic analyzer and all code that accesses type information

**Subtasks**:

- [x] 9.18.1 Design metadata architecture
  - Create SemanticInfo struct with node → type mapping
  - Design API for setting/getting type information
  - Consider thread safety for concurrent analyses
  - Document architecture decisions
  - File: `pkg/ast/metadata.go` (new file ~100 lines)

- [x] 9.18.2 Implement SemanticInfo type
  - Map[Expression]*TypeAnnotation for expression types
  - Map[*Identifier]Symbol for symbol resolution
  - Thread-safe accessors with sync.RWMutex
  - File: `pkg/ast/metadata.go`

- [x] 9.18.3 Remove Type field from AST expression nodes
  - Remove Type field from all expression node structs
  - Remove GetType/SetType methods (will be on SemanticInfo)
  - This affects ~30 node types
  - Files: `pkg/ast/base.go`, `pkg/ast/type_annotation.go`

- [x] 9.18.4 Update semantic analyzer to use SemanticInfo
  - Pass SemanticInfo through analyzer methods
  - Replace node.SetType() with semanticInfo.SetType(node, typ)
  - Replace node.GetType() with semanticInfo.GetType(node)
  - Files: `internal/semantic/*.go` (~11 occurrences)

- [x] 9.18.5 Update interpreter to use SemanticInfo
  - Pass SemanticInfo to interpreter
  - Get type information from SemanticInfo instead of nodes
  - Files: `internal/interp/*.go` (~5 occurrences)

- [x] 9.18.6 Update bytecode compiler to use SemanticInfo
  - Pass SemanticInfo to compiler
  - Get type information from metadata table
  - Files: `internal/bytecode/compiler_core.go`, `compiler_expressions.go`

- [x] 9.18.7 Update public API to return SemanticInfo
  - Engine.Analyze() returns SemanticInfo
  - Add accessor methods to Result type
  - Maintain backward compatibility where possible
  - Files: `pkg/dwscript/*.go`

- [ ] 9.18.8 Update LSP integration
  - Pass SemanticInfo to LSP handlers
  - Use metadata for hover, completion, etc.
  - Files: External go-dws-lsp project (document changes needed)

- [x] 9.18.9 Run comprehensive test suite
  - All semantic analyzer tests pass
  - All interpreter tests pass (pre-existing fixture failures unrelated to changes)
  - All bytecode VM tests pass
  - Type field removal complete - saves ~16 bytes per expression node

**Files Modified**:

- `pkg/ast/metadata.go` (new file ~150 lines)
- `pkg/ast/ast.go` (remove Type field from ~15 expression types)
- `pkg/ast/statements.go` (remove Type from CallExpression, etc.)
- `pkg/ast/control_flow.go` (remove Type from IfExpression)
- `pkg/ast/type_annotation.go` (remove TypedExpression interface or make it use SemanticInfo)
- `internal/semantic/analyzer.go` (add SemanticInfo field)
- `internal/semantic/*.go` (replace node.GetType/SetType ~50 times)
- `internal/interp/*.go` (use SemanticInfo for types ~30 times)
- `internal/bytecode/compiler.go` (use SemanticInfo)
- `pkg/dwscript/dwscript.go` (return SemanticInfo from API)

**Acceptance Criteria**:
- No Type field on any AST node
- SemanticInfo table stores all type metadata
- AST is immutable after parsing
- All tests pass (100% backward compatibility in behavior)
- Memory usage reduced (benchmark shows improvement)
- Multiple semantic analyses possible on same AST
- Documentation explains new architecture

**Benefits**:
- Clear separation of parsing vs semantic analysis
- AST is immutable and cacheable
- Reduced memory usage (~16 bytes per expression node)
- Multiple analyses possible (different contexts, parallel)
- Easier to implement alternative analyzers (strict mode, etc.)

---

- [x] 9.19 Extract Pretty-Printing from AST Nodes ✅ COMPLETED (2025-11-15) - Created `pkg/printer` package with multiple output formats (DWScript, Tree, JSON) and styles (Compact, Detailed, Multiline), simplified AST String() methods (~500 lines reduced), updated CLI with --format flag (PR #114)

---

- [ ] 9.20 Standardize Helper Types as Nodes

**Goal**: Make Parameter, CaseBranch, ExceptionHandler, and other helper types implement the Node interface to fix visitor pattern fragility.

**Estimate**: 3-4 hours (0.5 day)

**Status**: IN PROGRESS

**Impact**: Improved type safety, cleaner visitor pattern, more consistent AST structure

**Description**: Several types like `Parameter`, `CaseBranch`, `ExceptionHandler`, and `FieldInitializer` are not Nodes, which breaks the visitor pattern. They require manual handling in walk functions, making the code fragile. Making them implement Node provides type safety and consistent traversal.

**Current Problem**:

```go
// Parameter is not a Node - requires manual walking
type Parameter struct {
    Name         *Identifier
    Type         *TypeAnnotation
    DefaultValue Expression
    // Missing: Token, Pos(), End(), etc.
}

// In visitor.go - manual walking required
func walkFunctionDecl(n *FunctionDecl, v Visitor) {
    for _, param := range n.Parameters {
        // Can't call Walk() - Parameter is not a Node!
        if param.Name != nil {
            Walk(v, param.Name)
        }
        // Manual field walking...
    }
}
```

**Strategy**:

1. Identify all non-Node helper types
2. Add Node interface methods (Pos, End, TokenLiteral)
3. Add node marker methods (statementNode/expressionNode as appropriate)
4. Update visitor to treat them as first-class nodes

**Complexity**: Low - Straightforward interface implementation

**Subtasks**:

- [ ] 9.20.1 Audit AST for non-Node types used in traversal
  - Parameter (in FunctionDecl)
  - CaseBranch (in CaseStatement)
  - ExceptionHandler (in TryStatement)
  - ExceptClause (in TryStatement)
  - FinallyClause (in TryStatement)
  - FieldInitializer (in RecordLiteralExpression)
  - InterfaceMethodDecl (in InterfaceDecl)
  - Create list with current usage
  - File: Create `docs/ast-helper-types.md` with audit results

- [ ] 9.20.2 Make Parameter implement Node
  - Add Token token.Token field
  - Add EndPos token.Position field
  - Implement Pos(), End(), TokenLiteral()
  - Add statementNode() marker (parameters are like declarations)
  - File: `pkg/ast/functions.go`

- [ ] 9.20.3 Make CaseBranch implement Node
  - Add Token token.Token field (first value token)
  - Add EndPos token.Position field
  - Implement Node interface methods
  - Add statementNode() marker
  - File: `pkg/ast/control_flow.go`

- [ ] 9.20.4 Make ExceptionHandler, ExceptClause, FinallyClause implement Node
  - Add required fields to each type
  - Implement Node interface
  - Add statementNode() marker
  - File: `pkg/ast/exceptions.go`

- [ ] 9.20.5 Make FieldInitializer implement Node
  - Add Token, EndPos fields
  - Implement Node interface
  - Add statementNode() marker (like a mini assignment)
  - File: `pkg/ast/records.go`

- [ ] 9.20.6 Make InterfaceMethodDecl implement Node
  - Add Token, EndPos fields
  - Implement Node interface
  - Add statementNode() marker
  - File: `pkg/ast/interfaces.go`

- [ ] 9.20.7 Update visitor to walk helper types as Nodes
  - Remove manual field walking
  - Add cases for new Node types in Walk()
  - Simplify walkXXX functions
  - File: `pkg/ast/visitor.go` (or visitor_reflect.go if 9.17 done first)

- [ ] 9.20.8 Update parser to populate Token/EndPos for helper types
  - Ensure parser sets position info when creating helpers
  - Files: `internal/parser/*.go`

- [ ] 9.20.9 Test visitor traversal includes helper types
  - Create visitor that counts all nodes
  - Verify helpers are visited
  - File: `pkg/ast/visitor_test.go`

**Files Modified**:

- `pkg/ast/functions.go` (Parameter now implements Node)
- `pkg/ast/control_flow.go` (CaseBranch now implements Node)
- `pkg/ast/exceptions.go` (ExceptionHandler, ExceptClause, FinallyClause now implement Node)
- `pkg/ast/records.go` (FieldInitializer now implements Node)
- `pkg/ast/interfaces.go` (InterfaceMethodDecl now implements Node)
- `pkg/ast/visitor.go` (cleaner walk functions, add cases for new nodes)
- `internal/parser/*.go` (set Token/EndPos when creating helper types)
- `docs/ast-helper-types.md` (new documentation)

**Acceptance Criteria**:
- All traversable types implement Node interface
- No manual field walking in visitor.go
- Helper types can be visited like any other node
- All tests pass (especially visitor tests)
- Position information available for all helper types
- Documentation lists which types are Nodes

**Benefits**:
- Type safety (can't forget to walk a child)
- Cleaner visitor implementation
- Consistent AST structure
- Position info available for all traversable types
- Better error messages (can point to exact location)

---

- [ ] 9.21 Add Builder Pattern for Complex Nodes

**Goal**: Create builder types for complex AST nodes to prevent invalid construction and improve code clarity.

**Estimate**: 6-8 hours (1 day)

**Status**: IN PROGRESS

**Impact**: Prevents invalid AST construction, improves parser readability, catches errors at construction time

**Description**: Complex nodes like FunctionDecl and ClassDecl have many fields with interdependencies (e.g., can't be both virtual and abstract, must have body if not abstract, etc.). Currently, nothing prevents invalid combinations. Builders provide validation at construction time and make parser code more readable.

**Current Problem**:

```go
// Parser can create invalid combinations
fn := &FunctionDecl{
    Name: name,
    IsVirtual: true,
    IsAbstract: true,  // INVALID: can't be both!
    Body: nil,         // Missing body check
}
```

**Strategy**:

1. Create builder types for complex nodes
2. Builders enforce invariants and provide fluent API
3. Parser uses builders instead of direct struct construction
4. Builders validate on Build() call

**Complexity**: Medium - Need to identify all invariants and implement builders

**Subtasks**:

- [ ] 9.21.1 Identify nodes that need builders
  - FunctionDecl (most complex: ~15 boolean flags)
  - ClassDecl (inheritance, interfaces, abstract)
  - InterfaceDecl (inheritance)
  - PropertyDecl (read/write specs, indexed)
  - OperatorDecl (operator type, operands)
  - Create design doc with invariants for each
  - File: `docs/ast-builders.md` (new)

- [ ] 9.21.2 Create FunctionDeclBuilder
  - Fluent API: NewFunction(name).WithParam(p).Virtual().Build()
  - Validate: virtual XOR abstract, body required unless abstract/forward/external
  - Validate: constructor can't have return type
  - Validate: destructor must be named specific way
  - File: `pkg/ast/builders/function.go` (new package ~150 lines)

- [ ] 9.21.3 Create ClassDeclBuilder
  - Fluent API: NewClass(name).Extends(parent).Implements(iface).Abstract().Build()
  - Validate: parent is class, interfaces are interfaces
  - Validate: abstract flag consistent with abstract methods
  - Validate: partial + abstract combinations
  - File: `pkg/ast/builders/class.go` (new file ~120 lines)

- [ ] 9.21.4 Create InterfaceDeclBuilder
  - Fluent API: NewInterface(name).Extends(parent).WithMethod(m).Build()
  - Validate: parent is interface
  - Validate: methods are interface methods (no body)
  - File: `pkg/ast/builders/interface.go` (new file ~80 lines)

- [ ] 9.21.5 Create PropertyDeclBuilder
  - Fluent API: NewProperty(name, typ).Read(spec).Write(spec).Indexed(params).Build()
  - Validate: at least one of read/write specified
  - Validate: indexed params consistent
  - File: `pkg/ast/builders/property.go` (new file ~100 lines)

- [ ] 9.21.6 Create OperatorDeclBuilder
  - Fluent API: NewOperator(op).Unary(typ).Binary(lhs, rhs).Returns(ret).Build()
  - Validate: unary XOR binary
  - Validate: valid operator type
  - File: `pkg/ast/builders/operator.go` (new file ~80 lines)

- [ ] 9.21.7 Update parser to use builders
  - Replace direct struct construction with builders
  - Use fluent API for readability
  - Catch construction errors early
  - Files: `internal/parser/parser_functions.go`, `internal/parser/parser_class.go`, etc.

- [ ] 9.21.8 Add builder tests
  - Test valid construction succeeds
  - Test invalid construction fails with clear errors
  - Test all invariants enforced
  - File: `pkg/ast/builders/*_test.go` (new files ~300 lines total)

- [ ] 9.21.9 Add builder documentation
  - Examples of using each builder
  - List of all invariants enforced
  - Migration guide for parser
  - File: `pkg/ast/builders/doc.go` (new file)

**Files Modified**:

- `pkg/ast/builders/function.go` (new file ~150 lines)
- `pkg/ast/builders/class.go` (new file ~120 lines)
- `pkg/ast/builders/interface.go` (new file ~80 lines)
- `pkg/ast/builders/property.go` (new file ~100 lines)
- `pkg/ast/builders/operator.go` (new file ~80 lines)
- `pkg/ast/builders/doc.go` (new file ~50 lines)
- `pkg/ast/builders/*_test.go` (new files ~300 lines total)
- `internal/parser/parser_functions.go` (use FunctionDeclBuilder)
- `internal/parser/parser_class.go` (use ClassDeclBuilder)
- `internal/parser/parser_interfaces.go` (use InterfaceDeclBuilder)
- `internal/parser/parser_properties.go` (use PropertyDeclBuilder)
- `internal/parser/parser_operators.go` (use OperatorDeclBuilder)
- `docs/ast-builders.md` (new documentation ~50 lines)

**Acceptance Criteria**:
- Builders exist for FunctionDecl, ClassDecl, InterfaceDecl, PropertyDecl, OperatorDecl
- All invariants enforced (documented in ast-builders.md)
- Parser uses builders, catching errors at construction time
- Build() method validates and returns error for invalid combinations
- All tests pass, including new builder tests
- Parser code more readable with fluent API
- Documentation explains builder usage and invariants

**Benefits**:
- Catches invalid AST construction at parse time
- Self-documenting code (builder API shows what's valid)
- More readable parser (fluent API vs struct literals)
- Centralized validation logic
- Easier to add new invariants (add to builder, not scattered in parser)

---

- [ ] 9.22 Document Type System Architecture

**Goal**: Create comprehensive documentation explaining TypeAnnotation vs TypeExpression relationship and when to use each.

**Estimate**: 2-3 hours (0.5 day)

**Status**: IN PROGRESS

**Impact**: Improved developer understanding, easier onboarding, fewer type system bugs

**Description**: The relationship between `TypeAnnotation` and `TypeExpression` is unclear from the code alone. TypeAnnotation has both a `Name` field and an `InlineType TypeExpression` field, but it's not obvious when each is used. This confuses developers working on the type system. Clear documentation with examples and diagrams will improve understanding.

**Current Problem**:

```go
// What's the difference? When do I use Name vs InlineType?
type TypeAnnotation struct {
    InlineType TypeExpression  // ???
    Name       string          // ???
    Token      token.Token
    EndPos     token.Position
}
```

**Strategy**:

1. Create architecture documentation with clear explanations
2. Add examples of each use case
3. Create diagrams showing type system structure
4. Add code comments to type system code

**Complexity**: Low - Documentation task, no code changes required

**Subtasks**:

- [ ] 9.22.1 Document TypeAnnotation vs TypeExpression distinction
  - TypeAnnotation: Used when a type is referenced in syntax (`: Integer`)
  - TypeExpression: Defines the structure of a type (interface for type nodes)
  - Name: Simple type reference (`Integer`, `String`, `TMyClass`)
  - InlineType: Complex type definition (`array[0..10] of Integer`, `function(x: Integer): Boolean`)
  - File: `docs/type-system-architecture.md` (new file ~100 lines)

- [ ] 9.22.2 Create type system class diagram
  - Show hierarchy: Node → TypeExpression → specific types
  - Show TypeAnnotation composition
  - Show how semantic analyzer uses these
  - File: `docs/diagrams/type-system.svg` (new diagram)

- [ ] 9.22.3 Add examples for each type usage pattern
  - Example: Simple type reference (`var x: Integer`)
  - Example: Array type (`var arr: array[0..5] of Integer`)
  - Example: Function pointer type (`var fn: function(x: Integer): Boolean`)
  - Example: Anonymous record type
  - File: `docs/type-system-architecture.md` (add examples section)

- [ ] 9.22.4 Document type resolution process
  - How parser creates TypeAnnotations
  - How semantic analyzer resolves names to Type objects
  - How inline types are processed
  - Flow diagram: Source → TypeAnnotation → Type
  - File: `docs/type-system-architecture.md`

- [ ] 9.22.5 Add code comments to type system files
  - pkg/ast/type_annotation.go (explain fields)
  - pkg/ast/type_expression.go (explain interface)
  - internal/types/types.go (explain Type hierarchy)
  - Files: `pkg/ast/type_annotation.go`, `pkg/ast/type_expression.go`, `internal/types/types.go`

- [ ] 9.22.6 Create developer guide
  - "Adding a new type" guide
  - "Understanding type checking" guide
  - Common pitfalls and solutions
  - File: `docs/developer-guides/type-system.md` (new file ~50 lines)

- [ ] 9.22.7 Add package-level documentation
  - Update pkg/ast/doc.go with type system overview
  - Update internal/types/doc.go with Type hierarchy
  - Cross-reference with architecture docs
  - Files: `pkg/ast/doc.go`, `internal/types/doc.go`

**Files Modified**:

- `docs/type-system-architecture.md` (new file ~200 lines)
- `docs/diagrams/type-system.svg` (new diagram)
- `docs/developer-guides/type-system.md` (new file ~50 lines)
- `pkg/ast/type_annotation.go` (add detailed comments ~20 lines)
- `pkg/ast/type_expression.go` (add comments ~10 lines)
- `pkg/ast/doc.go` (add type system section ~20 lines)
- `internal/types/types.go` (add comments ~30 lines)
- `internal/types/doc.go` (create or update ~40 lines)

**Acceptance Criteria**:
- Clear documentation of TypeAnnotation vs TypeExpression
- Diagrams showing type system architecture
- Examples for each usage pattern
- Developer guide for working with types
- Code comments explain key concepts
- Documentation cross-referenced from code
- All type system files have package docs

**Benefits**:

- Faster developer onboarding
- Fewer type system bugs
- Clearer mental model of type system
- Easier to extend type system
- Self-documenting code

---

## Phase 13: go-dws API Enhancements for LSP Integration ✅ COMPLETE

**Goal**: Enhanced go-dws library with structured errors, AST access, position metadata, symbol tables, and type information for LSP features.

**Status**: All 27 tasks complete. Added public `pkg/ast/` and `pkg/token/` packages, structured error types with position info, Parse() mode for fast syntax-only parsing, visitor pattern for AST traversal, symbol table access, and type queries. 100% backwards compatible. Ready for go-dws-lsp integration.

---

## Phase 14: Bytecode Compiler & VM Optimizations ✅ MOSTLY COMPLETE

**Status**: Core implementation complete | **Performance**: 5-6x faster than AST interpreter | **Tasks**: 15 complete, 2 pending

### Overview

This phase implements a bytecode virtual machine for DWScript, providing significant performance improvements over the tree-walking AST interpreter. The bytecode VM uses a stack-based architecture with 116 opcodes and includes an optimization pipeline.

**Architecture**: AST → Compiler → Bytecode → VM → Output

### Phase 14.1: Bytecode VM Foundation ✅ COMPLETE

- [x] 14.1 Research and design bytecode instruction set
  - Stack-based VM with 116 opcodes, 32-bit instruction format
  - Documentation: [bytecode-vm-design.md](docs/architecture/bytecode-vm-design.md)
  - Expected Impact: 2-3x speedup over tree-walking interpreter

- [x] 14.2 Implement bytecode data structures
  - Created `internal/bytecode/bytecode.go` with `Chunk` type (bytecode + constants pool)
  - Implemented constant pool for literals with deduplication
  - Added line number mapping with run-length encoding
  - Implemented bytecode disassembler for debugging (79.7% coverage)

- [x] 14.3 Build AST-to-bytecode compiler
  - Created `internal/bytecode/compiler.go` with visitor pattern
  - Compile expressions: literals, binary ops, unary ops, variables, function calls
  - Compile statements: assignment, if/else, loops, return
  - Handle scoping and variable resolution
  - Optimize constant folding during compilation

- [x] 14.4 Implement bytecode VM core
  - Created `internal/bytecode/vm.go` with instruction dispatch loop
  - Implemented operand stack and call stack
  - Added environment/closure handling with upvalue capture
  - Error handling with structured RuntimeError and stack traces
  - Performance: VM is ~5.6x faster than AST interpreter

- [x] 14.5 Implement arithmetic and logic instructions
  - ADD, SUB, MUL, DIV, MOD instructions
  - NEGATE, NOT instructions
  - EQ, NE, LT, LE, GT, GE comparisons
  - AND, OR, XOR bitwise operations
  - Type coercion (int ↔ float)

- [x] 14.6 Implement variable and memory instructions
  - LOAD_CONST / LOAD_LOCAL / STORE_LOCAL
  - LOAD_GLOBAL / STORE_GLOBAL
  - LOAD_UPVALUE / STORE_UPVALUE with closure capture
  - GET_PROPERTY / SET_PROPERTY for member access

- [x] 14.7 Implement control flow instructions
  - JUMP, JUMP_IF_FALSE, JUMP_IF_TRUE
  - LOOP (jump backward for while/for loops)
  - Patch jump addresses during compilation
  - Break/continue leverage jump instructions

- [x] 14.8 Implement function call instructions
  - CALL instruction for named functions
  - RETURN instruction with trailing return guarantee
  - Handle recursion and call stack depth
  - Implement closures and upvalues
  - Support method calls and `Self` context (OpCallMethod, OpGetSelf)

- [x] 14.9 Implement array and object instructions
  - GET_INDEX, SET_INDEX for array access
  - NEW_ARRAY, ARRAY_LENGTH
  - NEW_OBJECT for class instantiation
  - INVOKE_METHOD for method dispatch

- [x] 14.10 Add exception handling instructions
  - TRY, CATCH, FINALLY, THROW instructions
  - Exception stack unwinding
  - Preserve stack traces across bytecode execution

- [x] 14.11 Optimize bytecode generation
  - Established optimization pipeline with pass manager and toggles
  - Peephole transforms: fold literal push/pop pairs, collapse stack shuffles
  - Dead code elimination: trim after terminators, reflow jump targets
  - Constant propagation: track literal locals/globals, fold arithmetic chains
  - Inline small functions (< 10 instructions)

- [x] 14.12 Integrate bytecode VM into interpreter
  - Added `--bytecode` flag to CLI
  - Added `CompileMode` option (AST vs Bytecode) to `pkg/dwscript/options.go`
  - Bytecode compilation/execution paths in `pkg/dwscript/dwscript.go`
  - Unit loading/parsing parity, tracing, diagnostic output
  - Wire bytecode VM to externals (FFI, built-ins, stdout capture)

- [x] 14.13 Create bytecode test suite
  - Port existing interpreter tests to bytecode
  - Test bytecode disassembler output
  - Verify identical behavior to AST interpreter
  - Performance benchmarks confirm 5-6x speedup

- [x] 14.14 Add bytecode serialization
  - [x] 14.14.1 Define bytecode file format (.dwc)
    - **Task**: Design the binary format for bytecode files
    - **Implementation**:
      - Define magic number (e.g., "DWC\x00") for file identification
      - Define version format (major.minor.patch)
      - Define header structure (magic, version, metadata)
      - Document format specification
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 0.5 day

  - [x] 14.14.2 Implement Chunk serialization
    - **Task**: Serialize bytecode chunks to binary format
    - **Implementation**:
      - Serialize instructions array
      - Serialize line number information
      - Serialize constants pool
      - Write helper functions for writing primitives (int, float, string, bool)
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 1 day

  - [x] 14.14.3 Implement Chunk deserialization
    - **Task**: Deserialize bytecode chunks from binary format
    - **Implementation**:
      - Read and validate magic number and version
      - Deserialize instructions array
      - Deserialize line number information
      - Deserialize constants pool
      - Write helper functions for reading primitives
      - Handle invalid/corrupt bytecode files
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 1 day

  - [x] 14.14.4 Add version compatibility checks
    - **Task**: Ensure bytecode version compatibility
    - **Implementation**:
      - Check version during deserialization
      - Return descriptive errors for version mismatches
      - Add tests for different version scenarios
    - **Files**: `internal/bytecode/serializer.go`
    - **Estimated time**: 0.5 day

  - [x] 14.14.5 Add serialization tests
    - **Task**: Test serialization/deserialization round-trip
    - **Implementation**:
      - Test simple programs serialize correctly
      - Test complex programs with all value types
      - Test error handling (corrupt files, version mismatches)
      - Verify bytecode produces same output after round-trip
    - **Files**: `internal/bytecode/serializer_test.go`
    - **Estimated time**: 1 day

  - [x] 14.14.6 Add `dwscript compile` command
    - **Task**: CLI command to compile source to bytecode
    - **Implementation**:
      - Add compile subcommand to CLI
      - Parse source file and compile to bytecode
      - Write bytecode to .dwc file
      - Add flags for output file, optimization level
    - **Files**: `cmd/dwscript/main.go`, `cmd/dwscript/compile.go`
    - **Estimated time**: 0.5 day

  - [x] 14.14.7 Update `dwscript run` to load .dwc files
    - **Task**: Allow running precompiled bytecode files
    - **Implementation**:
      - Detect .dwc file extension
      - Load bytecode from file instead of compiling
      - Add performance comparison in benchmarks
    - **Files**: `cmd/dwscript/main.go`, `cmd/dwscript/run.go`
    - **Estimated time**: 0.5 day

  - [x] 14.14.8 Document bytecode serialization
    - **Task**: Update documentation for bytecode files
    - **Implementation**:
      - Document .dwc file format in docs/bytecode-vm.md
      - Add CLI examples for compile command
      - Update README.md with serialization info
    - **Files**: `docs/bytecode-vm.md`, `README.md`, `CLAUDE.md`
    - **Estimated time**: 0.5 day

- [x] 14.15 Document bytecode VM
  - Written `docs/bytecode-vm.md` explaining architecture
  - Documented instruction set and opcodes
  - Provided examples of bytecode output
  - Updated CLAUDE.md with bytecode information

### Phase 14.16: Bytecode VM Core Feature Completion 🚧 IN PROGRESS

> **Source**: TODOs.md Task 1.5 - Bytecode VM Missing Core Features
>
> This section addresses critical missing features in the bytecode VM that prevent
> full parity with the AST interpreter. These are HIGH priority items that block
> VM usability for real-world scripts.

#### 14.16.1 For Loop Support 🔴 HIGH PRIORITY

**Status**: Not Started | **Complexity**: High | **Blocking**: VM cannot execute for loops

**Current State**:
- Opcodes `OpForPrep` and `OpForLoop` exist in instruction.go (lines 357-365)
- No `ForStatement` case in `compileStatement()` (compiler_statements.go)
- VM has no handlers for these opcodes in vm_exec.go
- For loop tests are commented out in vm_parity_test.go (lines 71-79)

**ForStatement AST** (pkg/ast/control_flow.go:167):
```go
type ForStatement struct {
    Start     Expression     // Starting value
    EndValue  Expression     // Ending value
    Body      Statement      // Loop body
    Step      Expression     // Optional step value
    Variable  *Identifier    // Loop variable
    Direction ForDirection   // ForTo or ForDownto
    InlineVar bool           // Whether variable is declared inline
}
```

**Implementation Tasks**:

- [ ] **14.16.1.1** Add ForStatement case to compileStatement()
  - **File**: `internal/bytecode/compiler_statements.go`
  - **Task**: Add `case *ast.ForStatement: return c.compileForStatement(node)` to switch

- [ ] **14.16.1.2** Implement compileForStatement() function
  - **File**: `internal/bytecode/compiler_statements.go`
  - **Complexity**: High (~100-150 lines)
  - **Requirements**:
    1. Evaluate start and end values at compile time or runtime
    2. Create loop variable in local scope (beginScope/declareLocal)
    3. Handle ForTo direction (ascending: current <= end)
    4. Handle ForDownto direction (descending: current >= end)
    5. Handle optional Step expression (default step = 1)
    6. Emit proper jump instructions:
       - Initialize loop variable with start value
       - Check loop condition (compare to end value)
       - Execute body
       - Increment/decrement loop variable by step
       - Jump back to condition check
    7. Handle break/continue within for loops (pushLoop/popLoop)
    8. Ensure loop variable scope is limited to loop body

  **Compilation Strategy Option A (Loop Unrolling with VM Support)**:
  ```
  ; for i := 1 to 3 do PrintLn(i);
  OpLoadConst 1         ; push start value
  OpLoadConst 3         ; push end value
  OpLoadConst 1         ; push step value (default 1)
  OpForPrep loopVar exitJump  ; init loop [start,end,step]->[current], check condition
  loop:
    ; body
    OpLoadLocal loopVar
    OpCallBuiltin PrintLn
    OpForLoop loopVar loop ; increment by step and check condition
  exit:
  ```

  **Compilation Strategy Option B (Compile to While-like Loop)**:
  ```
  ; for i := 1 to 3 do PrintLn(i);
  OpLoadConst 1         ; push start value
  OpStoreLocal loopVar  ; i := 1
  loopStart:
    OpLoadLocal loopVar   ; load i
    OpLoadConst 3         ; load end value
    OpLessEqual           ; i <= 3 (for ForTo)
    OpJumpIfFalse exit
    ; body
    OpLoadLocal loopVar
    OpCallBuiltin PrintLn
    ; increment
    OpLoadLocal loopVar
    OpLoadConst 1         ; step value
    OpAddInt
    OpStoreLocal loopVar
    OpLoop loopStart
  exit:
  ```

- [ ] **14.16.1.3** Implement OpForPrep handler in VM (if using Strategy A)
  - **File**: `internal/bytecode/vm_exec.go`
  - **Task**: Add case for OpForPrep opcode
  - **Semantics**: Initialize loop variable, check if loop should execute
  - **Note**: Direction (ForTo vs ForDownto) must be encoded in opcode operand (e.g., A byte)
    or Strategy B should be used which generates explicit comparison opcodes per direction

- [ ] **14.16.1.4** Implement OpForLoop handler in VM (if using Strategy A)
  - **File**: `internal/bytecode/vm_exec.go`
  - **Task**: Add case for OpForLoop opcode
  - **Semantics**: Increment/decrement loop variable, check condition, jump
  - **Note**: Must handle both ForTo (increment, check `<=`) and ForDownto (decrement, check `>=`)

- [ ] **14.16.1.5** Add loopKindFor to compiler loop tracking
  - **File**: `internal/bytecode/compiler_core.go`
  - **Task**: Add `loopKindFor` constant and update break/continue handling

- [ ] **14.16.1.6** Add for loop compilation tests
  - **File**: `internal/bytecode/compiler_statements_test.go`
  - **Tests**:
    - Simple for-to loop
    - Simple for-downto loop
    - For loop with step
    - For loop with break
    - For loop with continue
    - Nested for loops

- [ ] **14.16.1.7** Enable and verify VM parity tests
  - **File**: `internal/bytecode/vm_parity_test.go`
  - **Task**: Uncomment for loop test (lines 71-79)
  - **Verify**: Output matches AST interpreter

**Acceptance Criteria**:
- For loops compile without error
- ForTo and ForDownto directions work correctly
- Step expression works correctly
- Break/continue work within for loops
- VM parity test passes

---

#### 14.16.2 Result Variable Support 🔴 HIGH PRIORITY

**Status**: Not Started | **Complexity**: Medium | **Blocking**: Functions cannot use implicit Result variable

**Current State**:
- AST interpreter handles Result variable in functions_user.go (lines 88-146)
- Bytecode compiler's `compileFunctionDecl()` doesn't allocate Result variable
- Function tests with Result are commented out in vm_parity_test.go (lines 80-91)

**DWScript Result Variable Semantics**:
```pascal
function Add(a, b: Integer): Integer;
begin
  Result := a + b;  // Assign to Result variable
end;

function Multiply(x, y: Integer): Integer;
begin
  Multiply := x * y;  // Can also assign to function name
end;
```

**Implementation Tasks**:

- [ ] **14.16.2.1** Add Result variable allocation in function prologue
  - **File**: `internal/bytecode/compiler_statements.go` in `compileFunctionDecl()`
  - **Task**: If function has ReturnType, allocate Result as first local variable
  - **Code location**: After parameter binding, before body compilation
  ```go
  // After binding parameters:
  if fn.ReturnType != nil {
      resultSlot, err := child.declareLocal(
          &ast.Identifier{Value: "Result"},
          typeFromAnnotation(fn.ReturnType),
      )
      // NOTE: declareAlias does not exist yet; implement aliasing mechanism
      // or resolve function name to Result slot during identifier lookup
      // (see task 14.16.2.3 for implementation options)
  }
  ```

- [ ] **14.16.2.2** Initialize Result with appropriate default value
  - **File**: `internal/bytecode/compiler_statements.go`
  - **Task**: Emit code to initialize Result based on return type
  - **Default values**:
    - Integer: 0
    - Float: 0.0
    - String: ""
    - Boolean: false
    - Object/Interface: nil
    - Array: empty array
    - Record: initialized record

- [ ] **14.16.2.3** Make function name an alias for Result
  - **File**: `internal/bytecode/compiler_core.go`
  - **Task**: Implement aliasing so `FunctionName := value` resolves to Result slot
  - **Implementation options**:
    1. Add alias table to symbol table mapping function name to Result slot
    2. During identifier lookup in function scope, check if identifier matches
       current function name and resolve to Result slot instead
  - **Reference**: AST interpreter uses `ReferenceValue` pointing to "Result" (functions_user.go:145)

- [ ] **14.16.2.4** Update ensureFunctionReturn() to return Result
  - **File**: `internal/bytecode/compiler_statements.go`
  - **Task**: If function reaches end without explicit return, emit code to return Result value
  - **Current**: `ensureFunctionReturn()` emits `OpReturn 0` (no value)
  - **New**: If ReturnType != nil, emit `OpLoadLocal resultSlot; OpReturn 1`

- [ ] **14.16.2.5** Add Result variable tests
  - **File**: `internal/bytecode/compiler_functions_test.go`
  - **Tests**:
    - Function with Result assignment
    - Function with function name assignment
    - Function with both Result and explicit return
    - Nested functions with Result
    - Recursive function using Result

- [ ] **14.16.2.6** Enable and verify VM parity tests
  - **File**: `internal/bytecode/vm_parity_test.go`
  - **Task**: Uncomment Result variable test (lines 80-91)
  - **Verify**: Output matches AST interpreter

**Acceptance Criteria**:
- Functions can assign to Result variable
- Functions can assign to function name as alias for Result
- Functions without explicit return statement return Result value
- VM parity test passes

---

#### 14.16.3 Trim Builtin Implementation 🟡 MEDIUM PRIORITY

> **Note**: Originally marked as HIGH priority in TODOs.md. Re-evaluated to MEDIUM because
> Trim is less critical than for loops and Result variables for VM parity.

**Status**: Not Started | **Complexity**: Low | **Blocking**: Trim() calls fail in VM mode

**Current State**:
- Trim, TrimLeft, TrimRight exist in AST interpreter builtins (strings_basic.go:188-238)
- VM has TODO comment at vm_calls.go:196: `// TODO: Implement Trim builtin in VM`
- String helper method `trim` is commented out in vm_calls.go

**Implementation Tasks**:

- [ ] **14.16.3.1** Add builtinTrim function to VM
  - **File**: `internal/bytecode/vm_builtins_string.go`
  - **Task**: Implement Trim(str) - removes leading and trailing whitespace
  ```go
  func builtinTrim(vm *VM, args []Value) (Value, error) {
      if len(args) != 1 {
          return NilValue(), fmt.Errorf("Trim expects 1 argument, got %d", len(args))
      }
      if !args[0].IsString() {
          return NilValue(), fmt.Errorf("Trim expects string argument")
      }
      return StringValue(strings.TrimSpace(args[0].AsString())), nil
  }
  ```

- [ ] **14.16.3.2** Add builtinTrimLeft function to VM
  - **File**: `internal/bytecode/vm_builtins_string.go`
  - **Task**: Implement TrimLeft(str) - removes leading whitespace
  ```go
  func builtinTrimLeft(vm *VM, args []Value) (Value, error) {
      // Similar to builtinTrim but use strings.TrimLeft(str, " \t\n\r")
  }
  ```

- [ ] **14.16.3.3** Add builtinTrimRight function to VM
  - **File**: `internal/bytecode/vm_builtins_string.go`
  - **Task**: Implement TrimRight(str) - removes trailing whitespace
  ```go
  func builtinTrimRight(vm *VM, args []Value) (Value, error) {
      // Similar to builtinTrim but use strings.TrimRight(str, " \t\n\r")
  }
  ```

- [ ] **14.16.3.4** Register Trim functions in VM builtin table
  - **File**: `internal/bytecode/vm_builtins.go` or appropriate registration file
  - **Task**: Add entries to builtin function map
  ```go
  "trim":      builtinTrim,
  "trimleft":  builtinTrimLeft,
  "trimright": builtinTrimRight,
  ```

- [ ] **14.16.3.5** Enable String.Trim helper method in VM
  - **File**: `internal/bytecode/vm_calls.go`
  - **Task**: Add full implementation for the "trim" case in string helper method dispatch
    (currently only case label exists). Implementation should validate exactly one string
    argument and call `builtinTrim`, matching the pattern used for "toupper"/"tolower" cases.

- [ ] **14.16.3.6** Add Trim builtin tests
  - **File**: `internal/bytecode/vm_builtins_string_test.go`
  - **Tests**:
    - Trim with leading spaces
    - Trim with trailing spaces
    - Trim with both
    - Trim with tabs and newlines
    - TrimLeft and TrimRight variations
    - Empty string edge case

**Acceptance Criteria**:
- Trim() function works in bytecode VM
- TrimLeft() and TrimRight() work in bytecode VM
- String.Trim helper method works
- Tests pass

---

#### 14.16 Summary

| Task | Priority | Complexity | Status | Blocking |
|------|----------|------------|--------|----------|
| 14.16.1 For Loop Support | 🔴 HIGH | High | Not Started | VM cannot execute for loops |
| 14.16.2 Result Variable | 🔴 HIGH | Medium | Not Started | Functions cannot use Result |
| 14.16.3 Trim Builtin | 🟡 MEDIUM | Low | Not Started | Trim() calls fail in VM |

**Estimated Effort**:
- 14.16.1 For Loop Support: 2-3 days
- 14.16.2 Result Variable: 1-2 days
- 14.16.3 Trim Builtin: 0.5 day

**Total**: ~4-6 days

**Dependencies**:
- None (all tasks can be done independently)

**Testing Strategy**:
1. Unit tests for each feature
2. VM parity tests comparing output with AST interpreter
3. Integration tests with complex scripts using all features

[ ] 14.17 Add enum support to bytecode VM
  - **Task**: Implement enum value representation and operations in bytecode VM
  - **Priority**: MEDIUM (enums work fine in AST interpreter, this is optimization)
  - **Status**: NOT STARTED
  - **Implementation**:
    1. Add `ValueEnum` to ValueType enum in `bytecode.go`
    2. Add `EnumValue()` helper constructor for enum values
    3. Modify compiler to handle enum expressions:
       - Scoped enum access (e.g., `TEnum.Alpha`)
       - Enum type declarations in symbol table
       - Member access expressions for enums
    4. Add enum operations to VM:
       - Comparison operations (=, <>, <, >, <=, >=)
       - Boolean operations (and, or, xor) - Task 1.6 support
       - Implicit enum-to-integer conversion - Task 1.6 support
    5. Update bytecode serialization to support enum values
    6. Add comprehensive tests for enum operations in VM
  - **Current workaround**: Use AST interpreter for scripts with enums
  - **Failing Test**: `enum_bool_op.pas` fails with bytecode VM
  - **Example**:
    ```pascal
    type TEnum = flags (Alpha, Beta, Gamma);
    var x := TEnum.Alpha or TEnum.Gamma;  // Currently unsupported in VM
    PrintInt(TEnum.Beta);                  // Currently unsupported in VM
    ```
  - **Files to Modify**:
    - `internal/bytecode/bytecode.go` (add ValueEnum type)
    - `internal/bytecode/compiler_expressions.go` (compile enum literals and member access)
    - `internal/bytecode/compiler_core.go` (symbol table for enum types)
    - `internal/bytecode/vm_exec.go` (execute enum operations)
    - `internal/bytecode/serializer.go` (serialize/deserialize enum values)
  - **Estimated time**: 4-6 hours (1 day)
  - **Dependencies**: None (enum support complete in AST interpreter and semantic analyzer)

---

**Estimated time**: Completed in 12-16 weeks

## Phase 15: Future Bytecode Optimizations (DEFERRED)

- [ ] 15.1 Advanced peephole optimizations
  - [ ] Strength reduction (multiplication → shift)
  - [ ] Common subexpression elimination
  - [ ] Branch prediction hints

- [ ] 15.2 Register allocation improvements
  - [ ] Live range analysis
  - [ ] Register coloring for locals
  - [ ] Reduce stack traffic

- [ ] 15.3 Inline caching for method dispatch
  - [ ] Cache method lookup results
  - [ ] Invalidate on class redefinition
  - [ ] Benchmark polymorphic call sites

- [ ] 15.4 Bytecode verification
  - [ ] Static analysis of bytecode correctness
  - [ ] Type safety verification
  - [ ] Stack depth validation

---

## Phase 16: Performance & Polish

### Performance Profiling

- [x] 16.1 Create performance benchmark scripts
- [x] 16.2 Profile lexer performance: `BenchmarkLexer`
- [x] 16.3 Profile parser performance: `BenchmarkParser`
- [x] 16.4 Profile interpreter performance: `BenchmarkInterpreter`
- [x] 16.5 Identify bottlenecks using `pprof`
- [ ] 16.6 Document performance baseline

### Optimization - Lexer

- [ ] 16.7 Optimize string handling in lexer (use bytes instead of runes where possible)
- [ ] 16.8 Reduce allocations in token creation
- [ ] 16.9 Use string interning for keywords/identifiers
- [ ] 16.10 Benchmark improvements

### Optimization - Parser

- [ ] 16.11 Reduce AST node allocations
- [ ] 16.12 Pool commonly created nodes
- [ ] 16.13 Optimize precedence table lookups
- [ ] 16.14 Benchmark improvements

### Optimization - Interpreter

- [ ] 16.15 Optimize value representation (avoid interface{} overhead if possible)
- [ ] 16.16 Use switch statements instead of type assertions where possible
- [ ] 16.17 Cache frequently accessed symbols
- [ ] 16.18 Optimize environment lookups
- [ ] 16.19 Reduce allocations in hot paths
- [ ] 16.20 Benchmark improvements

### Memory Management

- [ ] 16.21 Ensure no memory leaks in long-running scripts
- [ ] 16.22 Profile memory usage with large programs
- [ ] 16.23 Optimize object allocation/deallocation
- [ ] 16.24 Consider object pooling for common types

### Code Quality Refactoring

- [ ] 16.25 Run `go vet ./...` and fix all issues
- [ ] 16.26 Run `golangci-lint run` and address warnings
- [ ] 16.27 Run `gofmt` on all files
- [ ] 16.28 Run `goimports` to organize imports
- [ ] 16.29 Review error handling consistency
- [ ] 16.30 Unify value representation if inconsistent
- [ ] 16.31 Refactor large functions into smaller ones
- [ ] 16.32 Extract common patterns into helper functions
- [ ] 16.33 Improve variable/function naming
- [ ] 16.34 Add missing error checks

### Documentation

- [ ] 16.35 Write comprehensive GoDoc comments for all exported types/functions
- [ ] 16.36 Document internal architecture in `docs/architecture.md`
- [ ] 16.37 Create user guide in `docs/user_guide.md`
- [ ] 16.38 Document CLI usage with examples
- [ ] 16.39 Create API documentation for embedding the library
- [ ] 16.40 Add code examples to documentation
- [ ] 16.41 Document known limitations
- [ ] 16.42 Create contribution guidelines in `CONTRIBUTING.md`

### Example Programs

- [x] 16.43 Create `examples/` directory
- [x] 16.44 Add example scripts:
  - [x] Hello World
  - [x] Fibonacci
  - [x] Factorial
  - [x] Class-based example (Person demo)
  - [x] Algorithm sample (math/loops showcase)
- [x] 16.45 Add README in examples directory
- [x] 16.46 Ensure all examples run correctly

### Testing Enhancements

- [ ] 16.47 Add integration tests in `test/integration/`
- [ ] 16.48 Add fuzzing tests for parser: `FuzzParser`
- [ ] 16.49 Add fuzzing tests for lexer: `FuzzLexer`
- [ ] 16.50 Add property-based tests (using testing/quick or gopter)
- [ ] 16.51 Ensure CI runs all test types
- [ ] 16.52 Achieve >90% code coverage overall
- [ ] 16.53 Add regression tests for all fixed bugs

### Release Preparation

- [ ] 16.54 Create `CHANGELOG.md`
- [ ] 16.55 Document version numbering scheme (SemVer)
- [ ] 16.56 Tag v0.1.0 alpha release
- [ ] 16.57 Create release binaries for major platforms (Linux, macOS, Windows)
- [ ] 16.58 Publish release on GitHub
- [ ] 16.59 Write announcement blog post or README update
- [ ] 16.60 Share with community for feedback

---

## Phase 17: Go Source Code Generation & AOT Compilation [RECOMMENDED]

**Status**: Not started | **Priority**: HIGH | **Estimated Time**: 20-28 weeks (code generation) + 9-13 weeks (CLI)

### Overview

This phase implements ahead-of-time (AOT) compilation by transpiling DWScript source code to Go, then compiling to native executables. This approach leverages Go's excellent cross-compilation support and delivers near-native performance.

**Approach**: DWScript Source → AST → Go Source Code → Go Compiler → Native Executable

**Benefits**: 10-50x faster than tree-walking interpreter, excellent portability, leverages Go toolchain

### Phase 13.1: Go Source Code Generation (20-28 weeks)

- [ ] 17.1 Design Go code generation architecture
  - Study similar transpilers (c2go, ast-transpiler)
  - Design AST → Go AST transformation strategy
  - Define runtime library interface
  - Document type mapping (DWScript → Go)
  - Plan package structure for generated code
  - **Decision**: Use `go/ast` package for Go AST generation

- [ ] 17.2 Create Go code generator foundation
  - Create `internal/codegen/` package
  - Create `internal/codegen/go_generator.go`
  - Implement `Generator` struct with context tracking
  - Add helper methods for code emission
  - Set up `go/ast` and `go/printer` integration
  - Create unit tests for basic generation

- [ ] 17.3 Implement type system mapping
  - Map DWScript primitives to Go types (Integer→int64, Float→float64, String→string, Boolean→bool)
  - Map DWScript arrays to Go slices (dynamic) or arrays (static)
  - Map DWScript records to Go structs
  - Map DWScript classes to Go structs with method tables
  - Handle type aliases and subrange types
  - Document type mapping in `docs/codegen-types.md`

- [ ] 17.4 Generate code for expressions
  - Generate literals (integer, float, string, boolean, nil)
  - Generate identifiers (variables, constants)
  - Generate binary operations (+, -, *, /, =, <>, <, >, etc.)
  - Generate unary operations (-, not)
  - Generate function calls
  - Generate array/object member access
  - Handle operator precedence correctly
  - Add unit tests comparing eval vs generated code

- [ ] 17.5 Generate code for statements
  - Generate variable declarations (`var x: Integer = 42`)
  - Generate assignments (`x := 10`)
  - Generate if/else statements
  - Generate while/repeat/for loops
  - Generate case statements (switch in Go)
  - Generate begin...end blocks
  - Handle break/continue/exit statements

- [ ] 17.6 Generate code for functions and procedures
  - Generate function declarations with parameters and return type
  - Handle by-value and by-reference (var) parameters
  - Generate procedure declarations (no return value)
  - Implement nested functions (closures in Go)
  - Support forward declarations
  - Handle recursion
  - Generate proper variable scoping

- [ ] 17.7 Generate code for classes and OOP
  - Generate Go struct definitions for classes
  - Generate constructor functions (Create)
  - Generate destructor cleanup (Destroy → defer)
  - Generate method declarations (receiver functions)
  - Implement inheritance (embedding in Go)
  - Implement virtual method dispatch (method tables)
  - Handle class fields and properties
  - Support `Self` keyword (receiver parameter)

- [ ] 17.8 Generate code for interfaces
  - Generate Go interface definitions
  - Implement interface casting and type assertions
  - Generate interface method dispatch
  - Handle interface inheritance
  - Support interface variables and parameters

- [ ] 17.9 Generate code for records
  - Generate Go struct definitions
  - Support record methods (static and instance)
  - Handle record literals and initialization
  - Generate record field access

- [ ] 17.10 Generate code for enums
  - Generate Go const declarations with iota
  - Support scoped and unscoped enum access
  - Generate Ord() and Integer() conversions
  - Handle explicit enum values

- [ ] 17.11 Generate code for arrays
  - Generate static arrays (Go arrays: `[10]int`)
  - Generate dynamic arrays (Go slices: `[]int`)
  - Support array literals
  - Generate array indexing and slicing
  - Implement SetLength, High, Low built-ins
  - Handle multi-dimensional arrays

- [ ] 17.12 Generate code for sets
  - Generate set types as Go map[T]bool or bitsets
  - Support set literals and constructors
  - Generate set operations (union, intersection, difference)
  - Implement `in` operator for set membership

- [ ] 17.13 Generate code for properties
  - Translate properties to getter/setter methods
  - Generate field-backed properties (direct access)
  - Generate method-backed properties (method calls)
  - Support read-only and write-only properties
  - Handle auto-properties

- [ ] 17.14 Generate code for exceptions
  - Generate try/except/finally as Go defer/recover
  - Map DWScript exceptions to Go error types
  - Generate raise statements (panic)
  - Implement exception class hierarchy
  - Preserve stack traces

- [ ] 17.15 Generate code for operators and conversions
  - Generate operator overloads as functions
  - Generate implicit conversions
  - Handle type coercion in expressions
  - Support custom operators

- [ ] 17.16 Create runtime library for generated code
  - Create `pkg/runtime/` package
  - Implement built-in functions (PrintLn, Length, Copy, etc.)
  - Implement array/string manipulation functions
  - Implement math functions (Sin, Cos, Sqrt, etc.)
  - Implement date/time functions
  - Provide runtime type information (RTTI) for reflection
  - Support external function calls (FFI)

- [ ] 17.17 Handle units/modules compilation
  - Generate separate Go packages for each unit
  - Handle unit dependencies and imports
  - Generate initialization/finalization code
  - Support uses clauses
  - Create package manifest

- [ ] 17.18 Implement optimization passes
  - Constant folding
  - Dead code elimination
  - Inline small functions
  - Remove unused variables
  - Optimize string concatenation
  - Use Go compiler optimization hints (//go:inline, etc.)

- [ ] 17.19 Add source mapping for debugging
  - Preserve line number comments in generated code
  - Generate source map files (.map)
  - Add DWScript source file embedding
  - Support stack trace translation (Go → DWScript)

- [ ] 17.20 Test Go code generation
  - Generate code for all fixture tests
  - Compile and run generated code
  - Compare output with interpreter
  - Measure compilation time
  - Benchmark generated code performance

**Expected Results**: 10-50x faster than tree-walking interpreter, near-native Go speed

## Phase 18: AOT Compiler CLI (9-13 weeks)

- [ ] 18.1 Create `dwscript compile` command
  - Add `compile` subcommand to CLI
  - Parse input DWScript file(s)
  - Generate Go source code to output directory
  - Invoke `go build` to create executable
  - Support multiple output formats (executable, library, package)

- [ ] 18.2 Implement project compilation mode
  - Support compiling entire projects (multiple units)
  - Generate go.mod file
  - Handle dependencies between units
  - Create main package with entry point
  - Support compilation configuration (optimization level, target platform)

- [ ] 18.3 Add compilation flags and options
  - `--output` or `-o` for output path
  - `--optimize` or `-O` for optimization level (0, 1, 2, 3)
  - `--keep-go-source` to preserve generated Go files
  - `--target` for cross-compilation (linux, windows, darwin, wasm)
  - `--static` for static linking
  - `--debug` to include debug symbols

- [ ] 18.4 Implement cross-compilation support
  - Support GOOS and GOARCH environment variables
  - Generate platform-specific code (if needed)
  - Test compilation for Linux, macOS, Windows, WASM
  - Document platform-specific limitations

- [ ] 18.5 Add incremental compilation
  - Cache compiled units
  - Detect file changes (mtime, hash)
  - Recompile only changed units
  - Rebuild dependency graph
  - Speed up repeated compilations

- [ ] 18.6 Create standalone binary builder
  - Generate single-file executable
  - Embed DWScript runtime
  - Strip debug symbols (optional)
  - Compress binary with UPX (optional)
  - Test on different platforms

- [ ] 18.7 Implement library compilation mode
  - Generate Go package (not executable)
  - Export public functions/classes
  - Create Go-friendly API
  - Generate documentation (godoc)
  - Support embedding in other Go projects

- [ ] 18.8 Add compilation error reporting
  - Catch Go compilation errors
  - Translate errors to DWScript source locations
  - Provide helpful error messages
  - Suggest fixes for common issues

- [ ] 18.9 Create compilation test suite
  - Test compilation of all fixture tests
  - Verify all executables run correctly
  - Test cross-compilation
  - Benchmark compilation speed
  - Measure binary sizes

- [ ] 18.10 Document AOT compilation
  - Write `docs/aot-compilation.md`
  - Explain compilation process
  - Provide usage examples
  - Document performance characteristics
  - Compare with interpretation and bytecode VM

---

## Phase 19: WebAssembly Runtime & Playground ✅ MOSTLY COMPLETE

**Status**: Core implementation complete | **Priority**: HIGH | **Tasks**: 23 complete, 3 pending

### Overview

This phase implements WebAssembly support for running DWScript in browsers, including a platform abstraction layer, WASM build infrastructure, JavaScript/Go bridge, and a web-based playground with Monaco editor integration.

**Architecture**: DWScript → WASM Binary → Browser/Node.js → JavaScript API

### Phase 19.1: Platform Abstraction Layer ✅ COMPLETE

- [x] 19.1 Create `pkg/platform/` package with core interfaces
  - FileSystem, Console, Platform interfaces
  - Enables native and WebAssembly builds with consistent behavior

- [x] 19.2 Implement `pkg/platform/native/` for standard Go
  - Standard Go implementations for native builds
  - Direct OS filesystem and console access

- [x] 19.3 Implement `pkg/platform/wasm/` with virtual filesystem
  - In-memory map for file storage
  - Console bridge to JavaScript console.log
  - Time functions using JavaScript Date API
  - Sleep implementation using setTimeout

- [ ] 19.4 Create feature parity test suite
  - Tests that run on both native and WASM
  - Validate platform abstraction works correctly

- [ ] 19.5 Document platform differences and limitations
  - Platform-specific behavior documentation
  - Known limitations in WASM environment

### Phase 19.2: WASM Build Infrastructure ✅ COMPLETE

- [x] 19.6 Create build infrastructure
  - `build/wasm/` directory with scripts
  - Justfile targets: `just wasm`, `just wasm-test`, `just wasm-optimize`, etc.
  - `cmd/dwscript-wasm/main.go` entry point with syscall/js exports

- [x] 19.7 Implement build modes support
  - Monolithic, modular, hybrid modes (compile-time flags)
  - `pkg/wasm/` package for WASM bridge code

- [x] 19.8 Add wasm_exec.js and optimization
  - wasm_exec.js from Go distribution (multi-version support)
  - Integrate wasm-opt (Binaryen) for binary size optimization
  - Size monitoring (warns if >3MB uncompressed)

- [ ] 19.9 Test all build modes
  - Compare sizes and performance
  - Validate each mode works correctly

- [x] 19.10 Document build process
  - `docs/wasm/BUILD.md` with build instructions
  - Configuration options and troubleshooting

### Phase 19.3: JavaScript/Go Bridge ✅ COMPLETE

- [x] 19.11 Implement DWScript class API
  - `pkg/wasm/api.go` using syscall/js
  - Export init(), compile(), run(), eval() to JavaScript

- [x] 19.12 Create type conversion utilities
  - Go types ↔ js.Value conversion in utils.go
  - Proper handling of DWScript types in JavaScript

- [x] 19.13 Implement callback registration system
  - `pkg/wasm/callbacks.go` for event handling
  - Virtual filesystem interface for JavaScript

- [x] 19.14 Add error handling across boundary
  - Panics → exceptions with recovery
  - Structured error objects for DWScript runtime errors

- [x] 19.15 Add event system
  - on() method for output, error, and custom events
  - Memory management with proper js.Value.Release()

- [x] 19.16 Document JavaScript API
  - `docs/wasm/API.md` with complete API reference
  - Usage examples for browser and Node.js

### Phase 19.4: Web Playground ✅ COMPLETE

- [x] 19.17 Create playground directory structure
  - `playground/` with HTML/CSS/JS files
  - Monaco Editor integration

- [x] 19.18 Implement syntax highlighting
  - DWScript language definition for Monaco
  - Tokenization rules matching lexer

- [x] 19.19 Build split-pane UI
  - Code editor + output console
  - Toolbar with Run, Examples, Clear, Share, Theme buttons

- [x] 19.20 Implement URL-based code sharing
  - Base64 encoded code in fragment
  - Examples dropdown with sample programs

- [x] 19.21 Add localStorage features
  - Auto-save and restore user code
  - Error markers in editor from compilation errors

- [x] 19.22 Set up GitHub Pages deployment
  - GitHub Actions workflow for automated deployment
  - Testing checklist in playground/TESTING.md

- [x] 19.23 Document playground architecture
  - `docs/wasm/PLAYGROUND.md` with architecture details
  - Extension points for future features

### Phase 19.5: NPM Package ✅ MOSTLY COMPLETE

- [x] 19.24 Create NPM package structure
  - `npm/` with package.json
  - TypeScript definitions in `typescript/index.d.ts`

- [x] 19.25 Create dual ESM/CommonJS entry points
  - index.js (ESM) and index.cjs (CommonJS)
  - WASM loader helper for Node.js and browser

- [x] 19.26 Add usage examples
  - Node.js, React, Vue, vanilla JS examples
  - Automated NPM publishing via GitHub Actions

- [x] 19.27 Configure for tree-shaking
  - Optimal bundling configuration
  - `npm/README.md` with installation guide

- [ ] 19.28 Publish to npmjs.com
  - Initial version publication
  - Version management strategy

### Phase 19.6: Testing & Documentation

- [ ] 19.29 Write WASM-specific tests
  - GOOS=js GOARCH=wasm go test
  - Node.js integration test suite

- [ ] 19.30 Add browser tests
  - Playwright tests for Chrome, Firefox, Safari
  - CI matrix for cross-browser testing

- [ ] 19.31 Add performance benchmarks
  - Compare WASM vs native speed
  - Bundle size regression monitoring in CI

- [ ] 19.32 Write embedding guide
  - `docs/wasm/EMBEDDING.md` for web app integration
  - Update main README with WASM section and playground link

---

## Phase 20: Community Building & Ecosystem [ONGOING]

**Status**: Ongoing | **Priority**: HIGH | **Estimated Tasks**: ~40

### Overview

This phase focuses on building a sustainable open-source community, maintaining feature parity with upstream DWScript, and providing essential tools for developers including REPL, debugging support, and platform-specific enhancements.

### Phase 20.1: Feature Parity Tracking

- [ ] 20.1 Create feature matrix comparing go-dws with DWScript
- [ ] 20.2 Track DWScript upstream releases
- [ ] 20.3 Identify new features in DWScript updates
- [ ] 20.4 Plan integration of new features
- [ ] 20.5 Update feature matrix regularly

### Phase 20.2: Community Building

- [ ] 20.6 Set up issue templates on GitHub
- [ ] 20.7 Set up pull request template
- [ ] 20.8 Create CODE_OF_CONDUCT.md
- [ ] 20.9 Create discussions forum or mailing list
- [ ] 20.10 Encourage contributions (tag "good first issue")
- [ ] 20.11 Respond to issues and PRs promptly
- [ ] 20.12 Build maintainer team (if interest grows)

### Phase 20.3: Advanced Features

- [ ] 20.13 Implement REPL (Read-Eval-Print Loop):
  - [ ] Interactive prompt
  - [ ] Statement-by-statement execution
  - [ ] Variable inspection
  - [ ] History and autocomplete
- [ ] 20.14 Implement debugging support:
  - [ ] Breakpoints
  - [ ] Step-through execution
  - [ ] Variable inspection
  - [ ] Stack traces
- [x] 20.15 WebAssembly Runtime & Playground - **See Phase 14** (mostly complete)
- [x] 20.16 Language Server Protocol (LSP) - **See external repo** https://github.com/cwbudde/go-dws-lsp
- [ ] 20.17 JavaScript code generation backend - **See Phase 16** (deferred)

### Phase 20.4: Platform-Specific Enhancements

- [ ] 20.18 Add Windows-specific features (if needed)
- [ ] 20.19 Add macOS-specific features (if needed)
- [ ] 20.20 Add Linux-specific features (if needed)
- [ ] 20.21 Test on multiple architectures (ARM, AMD64)

### Phase 20.5: Edge Case Audit

- [ ] 20.22 Test short-circuit evaluation (and, or)
- [ ] 20.23 Test operator precedence edge cases
- [ ] 20.24 Test division by zero handling
- [ ] 20.25 Test integer overflow behavior
- [ ] 20.26 Test floating-point edge cases (NaN, Inf)
- [ ] 20.27 Test string encoding (UTF-8 handling)
- [ ] 20.28 Test very large programs (scalability)
- [ ] 20.29 Test deeply nested structures
- [ ] 20.30 Test circular references (if possible in language)
- [ ] 20.31 Fix any discovered issues

### Phase 20.6: Performance Monitoring

- [ ] 20.32 Set up continuous performance benchmarking
- [ ] 20.33 Track performance metrics over releases
- [ ] 20.34 Identify and fix performance regressions
- [ ] 20.35 Publish performance comparison with DWScript

### Phase 20.7: Security Audit

- [ ] 20.36 Review for potential security issues (untrusted script execution)
- [ ] 20.37 Implement resource limits (memory, execution time)
- [ ] 20.38 Implement sandboxing for untrusted scripts
- [ ] 20.39 Audit for code injection vulnerabilities
- [ ] 20.40 Document security best practices

### Phase 20.8: Maintenance

- [ ] 20.41 Keep dependencies up to date
- [ ] 20.42 Monitor Go version updates and migrate as needed
- [ ] 20.43 Maintain CI/CD pipeline
- [ ] 20.44 Regular code reviews
- [ ] 20.45 Address technical debt periodically

### Phase 20.9: Long-term Roadmap

- [ ] 20.46 Define 1-year roadmap
- [ ] 20.47 Define 3-year roadmap
- [ ] 20.48 Gather user feedback and adjust priorities
- [ ] 20.49 Consider commercial applications/support
- [ ] 20.50 Explore academic/research collaborations

---

## Phase 21: MIR Foundation [DEFERRED]

**Status**: Not started | **Priority**: MEDIUM | **Estimated Tasks**: 47 (MIR core, lowering, testing)

### Overview

This phase implements a Mid-level Intermediate Representation (MIR) that serves as a target-neutral layer between the type-checked AST and backend code generators. The MIR enables multiple backend targets (JavaScript, LLVM, C, etc.) from a single lowering pass.

**Architecture Flow**: DWScript Source → Parser → Semantic Analyzer → **MIR Builder** → [Backend Emitter] → Output

**Why MIR?** Clean separation of concerns, multi-backend support, platform-independent optimizations, easier debugging, and future-proofing for additional compilation targets.

**Note**: JavaScript backend is implemented in Phase 16, LLVM backend in Phase 17.

### Phase 21.1: MIR Foundation (30 tasks)

**Goal**: Define a complete, verifiable mid-level IR that can represent all DWScript constructs in a target-neutral way.

**Exit Criteria**: MIR spec documented, complete type system, builder API, verifier, AST→MIR lowering for ~80% of constructs, 20+ golden tests, 85%+ coverage

#### 21.1.1: MIR Package Structure and Types (10 tasks)

- [ ] 21.1 Create `mir/` package directory
- [ ] 21.2 Create `mir/types.go` - MIR type system
- [ ] 21.3 Define `Type` interface with `String()`, `Size()`, `Align()` methods
- [ ] 21.4 Implement primitive types: `Bool`, `Int8`, `Int16`, `Int32`, `Int64`, `Float32`, `Float64`, `String`
- [ ] 21.5 Implement composite types: `Array(elemType, size)`, `Record(fields)`, `Pointer(pointeeType)`
- [ ] 21.6 Implement OOP types: `Class(name, fields, methods, parent)`, `Interface(name, methods)`
- [ ] 21.7 Implement function types: `Function(params, returnType)`
- [ ] 21.8 Add `Void` type for procedures
- [ ] 21.9 Implement type equality and compatibility checking
- [ ] 21.10 Implement type conversion rules (explicit vs implicit)

#### 21.1.2: MIR Instructions and Control Flow (10 tasks)

- [ ] 21.11 Create `mir/instruction.go` - MIR instruction set
- [ ] 21.12 Define `Instruction` interface with `ID()`, `Type()`, `String()` methods
- [ ] 21.13 Implement arithmetic ops: `Add`, `Sub`, `Mul`, `Div`, `Mod`, `Neg`
- [ ] 21.14 Implement comparison ops: `Eq`, `Ne`, `Lt`, `Le`, `Gt`, `Ge`
- [ ] 21.15 Implement logical ops: `And`, `Or`, `Xor`, `Not`
- [ ] 21.16 Implement memory ops: `Alloca`, `Load`, `Store`
- [ ] 21.17 Implement constants: `ConstInt`, `ConstFloat`, `ConstString`, `ConstBool`, `ConstNil`
- [ ] 21.18 Implement conversions: `IntToFloat`, `FloatToInt`, `IntTrunc`, `IntExt`
- [ ] 21.19 Implement function ops: `Call`, `VirtualCall`
- [ ] 21.20 Implement array/class ops: `ArrayAlloc`, `ArrayLen`, `ArrayIndex`, `ArraySet`, `FieldGet`, `FieldSet`, `New`

#### 21.1.3: MIR Control Flow Structures (5 tasks)

- [ ] 21.21 Create `mir/block.go` - Basic blocks with `ID`, `Instructions`, `Terminator`
- [ ] 21.22 Implement control flow terminators: `Phi`, `Br`, `CondBr`, `Return`, `Throw`
- [ ] 21.23 Implement terminator validation (every block must end with terminator)
- [ ] 21.24 Implement block predecessors/successors tracking for CFG
- [ ] 21.25 Create `mir/function.go` - Function representation with `Name`, `Params`, `ReturnType`, `Blocks`, `Locals`

#### 21.1.4: MIR Builder API (3 tasks)

- [ ] 21.26 Create `mir/builder.go` - Safe MIR construction
- [ ] 21.27 Implement `Builder` struct with function/block context, `NewFunction()`, `NewBlock()`, `SetInsertPoint()`
- [ ] 21.28 Implement instruction emission methods: `EmitAdd()`, `EmitLoad()`, `EmitStore()`, etc. with type checking

#### 21.1.5: MIR Verifier (2 tasks)

- [ ] 21.29 Create `mir/verifier.go` - MIR correctness checking
- [ ] 21.30 Implement CFG, type, SSA, and function signature verification with `Verify(fn *Function) []error` API

### Phase 15.2: AST → MIR Lowering (12 tasks)

- [ ] 21.31 Create `mir/lower.go` - AST to MIR translation
- [ ] 21.32 Implement `LowerProgram(ast *ast.Program) (*mir.Module, error)` entry point
- [ ] 21.33 Lower expressions: literals → `Const*` instructions
- [ ] 21.34 Lower binary operations → corresponding MIR ops (handle short-circuit for `and`/`or`)
- [ ] 21.35 Lower unary operations → `Neg`, `Not`
- [ ] 21.36 Lower identifier references → `Load` instructions
- [ ] 21.37 Lower function calls → `Call` instructions
- [ ] 21.38 Lower array indexing → `ArrayIndex` + bounds check insertion
- [ ] 21.39 Lower record field access → `FieldGet`/`FieldSet`
- [ ] 21.40 Lower statements: variable declarations, assignments, if/while/for, return
- [ ] 21.41 Lower declarations: functions/procedures, records, classes
- [ ] 21.42 Implement short-circuit evaluation and simple optimizations (constant folding, dead code elimination)

### Phase 15.3: MIR Debugging and Testing (5 tasks)

- [ ] 21.43 Create `mir/dump.go` - Human-readable MIR output with `Dump(fn *Function) string`
- [ ] 21.44 Integration with CLI: `./bin/dwscript dump-mir script.dws`
- [ ] 21.45 Create golden MIR tests: 5+ each for expressions, control flow, functions, advanced features
- [ ] 21.46 Implement MIR verifier tests: type mismatches, malformed CFG, SSA violations
- [ ] 21.47 Implement round-trip tests: AST → MIR → verify → dump → compare with golden files

---

## Phase 22: JavaScript Backend [DEFERRED]

**Status**: Not started | **Priority**: MEDIUM | **Estimated Tasks**: 105 (MVP + feature complete)

### Overview

This phase implements a JavaScript code generator that translates MIR to readable, runnable JavaScript. The backend enables running DWScript programs in browsers and Node.js environments.

**Architecture Flow**: MIR → JavaScript Emitter → JavaScript Code → Node.js/Browser

**Benefits**: Browser execution, npm ecosystem integration, excellent portability, leverages JavaScript JIT compilers

**Dependencies**: Requires Phase 15 (MIR Foundation) to be completed first

### Phase 22.1: JS Backend MVP (45 tasks)

**Goal**: Implement a JavaScript code generator that can compile basic DWScript programs to readable, runnable JavaScript.

**Exit Criteria**: JS emitter for expressions/control flow/functions, 20+ end-to-end tests (DWScript→JS→execute), golden JS snapshots, 85%+ coverage

#### 22.4.1: JS Emitter Infrastructure (8 tasks)

- [ ] 22.48 Create `codegen/` package with `Backend` interface and `EmitterOptions`
- [ ] 22.49 Create `codegen/js/` package and `emitter.go`
- [ ] 22.50 Define `JSEmitter` struct with `out`, `indent`, `opts`, `tmpCounter`
- [ ] 22.51 Implement helper methods: `emit()`, `emitLine()`, `emitIndent()`, `pushIndent()`, `popIndent()`
- [ ] 22.52 Implement `newTemp()` for temporary variable naming
- [ ] 22.53 Implement `NewJSEmitter(opts EmitterOptions)`
- [ ] 22.54 Implement `Generate(module *mir.Module) (string, error)` entry point
- [ ] 22.55 Test emitter infrastructure

#### 22.4.2: Module and Function Emission (6 tasks)

- [ ] 22.56 Implement module structure emission: ES Module format with `export`, file header comment
- [ ] 22.57 Implement optional IIFE fallback via `EmitterOptions`
- [ ] 22.58 Implement function emission: `function fname(params) { ... }`
- [ ] 22.59 Map DWScript params to JS params (preserve names)
- [ ] 22.60 Emit local variable declarations at function top (from `Alloca` instructions)
- [ ] 22.61 Handle procedures (no return value) as JS functions

#### 22.4.3: Expression and Instruction Lowering (12 tasks)

- [ ] 22.62 Lower arithmetic operations → JS infix operators: `+`, `-`, `*`, `/`, `%`, unary `-`
- [ ] 22.63 Lower comparison operations → JS comparisons: `===`, `!==`, `<`, `<=`, `>`, `>=`
- [ ] 22.64 Lower logical operations → JS boolean ops: `&&`, `||`, `!`
- [ ] 22.65 Lower constants → JS literals with proper escaping
- [ ] 22.66 Lower variable operations: `Load` → variable reference, `Store` → assignment
- [ ] 22.67 Lower function calls: `Call` → `functionName(args)`
- [ ] 22.68 Implement Phi node lowering with temporary variables at block edges
- [ ] 22.69 Test expression lowering
- [ ] 22.70 Test instruction lowering
- [ ] 22.71 Test temporary variable generation
- [ ] 22.72 Test type conversions
- [ ] 22.73 Test complex expressions

#### 22.4.4: Control Flow Emission (8 tasks)

- [ ] 22.74 Implement control flow reconstruction from MIR CFG
- [ ] 22.75 Detect if/else patterns from `CondBr`
- [ ] 22.76 Detect while loop patterns (backedge to header)
- [ ] 22.77 Emit if-else: `if (condition) { ... } else { ... }`
- [ ] 22.78 Emit while loops: `while (condition) { ... }`
- [ ] 22.79 Emit for loops if MIR preserves metadata
- [ ] 22.80 Handle unconditional branches
- [ ] 22.81 Handle return statements

#### 22.4.5: Runtime and Testing (11 tasks)

- [ ] 22.82 Create `runtime/js/runtime.js` with `_dws.boundsCheck()`, `_dws.assert()`
- [ ] 22.83 Emit runtime import in generated JS (if needed)
- [ ] 22.84 Make runtime usage optional via `EmitterOptions.InsertBoundsChecks`
- [ ] 22.85 Create `codegen/js/testdata/` with subdirectories
- [ ] 22.86 Implement golden JS snapshot tests
- [ ] 22.87 Setup Node.js in CI (GitHub Actions)
- [ ] 22.88 Implement execution tests: parse → lower → generate → execute → verify
- [ ] 22.89 Add end-to-end tests for arithmetic, control flow, functions, loops
- [ ] 22.90 Add unit tests for JS emitter
- [ ] 22.91 Achieve 85%+ coverage for `codegen/js/` package
- [ ] 22.92 Add `compile-js` CLI command: `./bin/dwscript compile-js input.dws -o output.js`

### Phase 22.2: JS Feature Complete (60 tasks)

**Goal**: Extend JS backend to support all DWScript language features.

**Exit Criteria**: Full OOP, composite types, exceptions, properties, 50+ comprehensive tests, real-world samples work

#### 22.5.1: Records (7 tasks)

- [ ] 22.93 Implement MIR support for records
- [ ] 22.94 Emit records as plain JS objects: `{ x: 0, y: 0 }`
- [ ] 22.95 Implement constructor functions for records
- [ ] 22.96 Implement field access/assignment as property access
- [ ] 22.97 Implement record copy semantics with `_dws.copyRecord()`
- [ ] 22.98 Test record creation, initialization, field read/write
- [ ] 22.99 Test nested records and copy semantics

#### 22.5.2: Arrays (10 tasks)

- [ ] 22.100 Extend MIR for static and dynamic arrays
- [ ] 22.101 Emit static arrays as JS arrays with fixed size
- [ ] 22.102 Implement array index access with optional bounds checking
- [ ] 22.103 Emit dynamic arrays as JS arrays
- [ ] 22.104 Implement `SetLength` → `arr.length = newLen`
- [ ] 22.105 Implement `Length` → `arr.length`
- [ ] 22.106 Support multi-dimensional arrays (nested JS arrays)
- [ ] 22.107 Implement array operations: copy, concatenation
- [ ] 22.108 Test static array creation and indexing
- [ ] 22.109 Test dynamic array operations and bounds checking

#### 22.5.3: Classes and Inheritance (15 tasks)

- [ ] 22.110 Extend MIR for classes with fields, methods, parent, vtable
- [ ] 22.111 Emit ES6 class syntax: `class TAnimal { ... }`
- [ ] 22.112 Implement field initialization in constructor
- [ ] 22.113 Implement method emission
- [ ] 22.114 Implement inheritance with `extends` clause
- [ ] 22.115 Implement `super()` call in constructor
- [ ] 22.116 Handle virtual method dispatch (naturally virtual in JS)
- [ ] 22.117 Handle DWScript `Create` → JS `constructor`
- [ ] 22.118 Handle multiple constructors (overload dispatch)
- [ ] 22.119 Document destructor handling (no direct equivalent in JS)
- [ ] 22.120 Implement static fields and methods
- [ ] 22.121 Map `Self` → `this`, `inherited` → `super.method()`
- [ ] 22.122 Test simple classes with fields and methods
- [ ] 22.123 Test inheritance, virtual method overriding, constructors
- [ ] 22.124 Test static members and `Self`/`inherited` usage

#### 22.5.4: Interfaces (6 tasks)

- [ ] 22.125 Extend MIR for interfaces
- [ ] 22.126 Choose and document JS emission strategy (structural typing vs runtime metadata)
- [ ] 22.127 If using runtime metadata: emit interface tables, implement `is`/`as` operators
- [ ] 22.128 Test class implementing interface
- [ ] 22.129 Test interface method calls
- [ ] 22.130 Test `is` and `as` with interfaces

#### 22.5.5: Enums and Sets (8 tasks)

- [ ] 22.131 Extend MIR for enums
- [ ] 22.132 Emit enums as frozen JS objects: `const TColor = Object.freeze({...})`
- [ ] 22.133 Support scoped and unscoped enum access
- [ ] 22.134 Extend MIR for sets
- [ ] 22.135 Emit small sets (≤32 elements) as bitmasks
- [ ] 22.136 Emit large sets as JS `Set` objects
- [ ] 22.137 Implement set operations: union, intersection, difference, inclusion
- [ ] 22.138 Test enum declaration/usage and set operations

#### 22.5.6: Exception Handling (8 tasks)

- [ ] 22.139 Extend MIR for exceptions: `Throw`, `Try`, `Catch`, `Finally`
- [ ] 22.140 Emit `Throw` → `throw new Error()` or custom exception class
- [ ] 22.141 Emit try-except-finally → JS `try/catch/finally`
- [ ] 22.142 Create DWScript exception class → JS `Error` subclass
- [ ] 22.143 Handle `On E: ExceptionType do` with instanceof checks
- [ ] 22.144 Implement re-raise with exception tracking
- [ ] 22.145 Test basic try-except, multiple handlers, try-finally
- [ ] 22.146 Test re-raise and nested exception handling

#### 22.5.7: Properties and Advanced Features (6 tasks)

- [ ] 22.147 Extend MIR for properties with `PropGet`/`PropSet`
- [ ] 22.148 Emit properties as ES6 getters/setters
- [ ] 22.149 Handle indexed properties as methods
- [ ] 22.150 Test read/write properties and indexed properties
- [ ] 22.151 Implement operator overloading (desugar to method calls)
- [ ] 22.152 Implement generics support (monomorphization)

---

## Phase 23: LLVM Backend [DEFERRED]

**Status**: Not started | **Priority**: LOW | **Estimated Tasks**: 45

### Overview

This phase implements an LLVM IR backend for native code compilation, consolidating all LLVM-related work from the original Phase 13 LLVM JIT and AOT sections. This provides maximum performance but adds significant complexity.

**Architecture Flow**: MIR → LLVM IR Emitter → LLVM IR → llc → Native Binary

**Benefits**: Maximum performance (near C/C++ speed), excellent optimization, multi-architecture support

**Dependencies**: Requires Phase 15 (MIR Foundation) to be completed first

**Note**: This phase consolidates LLVM JIT (from old Phase 13.2), LLVM AOT (from old Phase 13.4), and LLVM backend (from old Stage 14.6). Given complexity and maintenance burden, this is marked as DEFERRED with LOW priority. The bytecode VM and Go AOT provide sufficient performance for most use cases.

### Phase 23.1: LLVM Infrastructure (8 tasks)

**Goal**: Set up LLVM bindings, type mapping, and runtime declarations

- [ ] 23.153 Choose LLVM binding: `llir/llvm` (pure Go) vs CGo bindings
- [ ] 23.154 Create `codegen/llvm/` package with `emitter.go`, `types.go`, `runtime.go`
- [ ] 23.155 Implement type mapping: DWScript types → LLVM types
- [ ] 23.156 Map Integer → `i32`/`i64`, Float → `double`, Boolean → `i1`
- [ ] 23.157 Map String → struct `{i32 len, i8* data}`
- [ ] 23.158 Map arrays/objects to LLVM structs
- [ ] 23.159 Emit LLVM module with target triple
- [ ] 23.160 Declare external runtime functions

### Phase 23.2: Runtime Library (12 tasks)

- [ ] 23.161 Create `runtime/dws_runtime.h` - C header for runtime API
- [ ] 23.162 Declare string operations: `dws_string_new()`, `dws_string_concat()`, `dws_string_len()`
- [ ] 23.163 Declare array operations: `dws_array_new()`, `dws_array_index()`, `dws_array_len()`
- [ ] 23.164 Declare memory management: `dws_alloc()`, `dws_free()`
- [ ] 23.165 Choose and document memory strategy (Boehm GC vs reference counting)
- [ ] 23.166 Declare object operations: `dws_object_new()`, virtual dispatch helpers
- [ ] 23.167 Declare exception handling: `dws_throw()`, `dws_catch()`
- [ ] 23.168 Declare RTTI: `dws_is_instance()`, `dws_as_instance()`
- [ ] 23.169 Create `runtime/dws_runtime.c` - implement runtime
- [ ] 23.170 Implement all runtime functions
- [ ] 23.171 Create `runtime/Makefile` to build `libdws_runtime.a`
- [ ] 23.172 Add runtime build to CI for Linux/macOS/Windows

### Phase 23.3: LLVM Code Emission (15 tasks)

- [ ] 23.173 Implement LLVM emitter: `Generate(module *mir.Module) (string, error)`
- [ ] 23.174 Emit function declarations with correct signatures
- [ ] 23.175 Emit basic blocks for each MIR block
- [ ] 23.176 Emit arithmetic instructions: `add`, `sub`, `mul`, `sdiv`, `srem`
- [ ] 23.177 Emit comparison instructions: `icmp eq`, `icmp slt`, etc.
- [ ] 23.178 Emit logical instructions: `and`, `or`, `xor`
- [ ] 23.179 Emit memory instructions: `alloca`, `load`, `store`
- [ ] 23.180 Emit call instructions: `call @function_name(args)`
- [ ] 23.181 Emit constants: integers, floats, strings
- [ ] 23.182 Emit control flow: conditional branches, phi nodes
- [ ] 23.183 Emit runtime calls for strings, arrays, objects
- [ ] 23.184 Implement type conversions: `sitofp`, `fptosi`
- [ ] 23.185 Emit struct types for classes and vtables
- [ ] 23.186 Implement virtual method dispatch
- [ ] 23.187 Implement exception handling (simple throw/catch or full LLVM EH)

### Phase 23.4: Linking and Testing (7 tasks)

- [ ] 23.188 Implement compilation pipeline: DWScript → MIR → LLVM IR → object → executable
- [ ] 23.189 Integrate `llc` to compile .ll → .o
- [ ] 23.190 Integrate linker to link object + runtime → executable
- [ ] 23.191 Add `compile-native` CLI command
- [ ] 23.192 Create 10+ end-to-end tests: DWScript → native → execute → verify
- [ ] 23.193 Benchmark JS vs native performance
- [ ] 23.194 Document LLVM backend in `docs/llvm-backend.md`

### Phase 23.5: Documentation (3 tasks)

- [ ] 23.195 Create `docs/codegen-architecture.md` - MIR overview, multi-backend design
- [ ] 23.196 Create `docs/mir-spec.md` - complete MIR reference with examples
- [ ] 23.197 Create `docs/js-backend.md` - DWScript → JavaScript mapping guide

---

## Phase 24: WebAssembly AOT Compilation [RECOMMENDED]

**Status**: Not started | **Priority**: MEDIUM-HIGH | **Estimated Tasks**: 5

### Overview

This phase extends WebAssembly support to generate standalone WASM binaries that can run without JavaScript dependency. This builds on Phase 14 (WASM Runtime & Playground) but focuses on ahead-of-time compilation for server-side and edge deployment.

**Architecture Flow**: DWScript Source → Go Compiler → WASM Binary → WASI Runtime (wasmtime, wasmer, wazero)

**Benefits**: Portable binaries, server-side execution, edge computing support, sandboxed execution

**Dependencies**: Builds on Phase 14 (WebAssembly Runtime & Playground)

### Phase 24.1: Standalone WASM Binaries (5 tasks)

- [ ] 24.1 Extend WASM compilation for standalone binaries
  - Generate WASM modules without JavaScript dependency
  - Use WASI for system calls
  - Support WASM-compatible runtime
  - Test with wasmtime, wasmer, wazero

- [ ] 24.2 Optimize WASM binary size
  - Use TinyGo compiler (smaller binaries)
  - Enable wasm-opt optimization
  - Strip unnecessary features
  - Measure binary size (target < 1MB)

- [ ] 24.3 Add WASM runtime support
  - Bundle WASM runtime (wasmer-go or wazero)
  - Create launcher executable
  - Support both JIT and AOT WASM execution
  - Test performance

- [ ] 24.4 Test WASM AOT compilation
  - Compile fixture tests to WASM
  - Run with different WASM runtimes
  - Measure performance vs native
  - Test browser and server execution

- [ ] 24.5 Document WASM AOT
  - Write `docs/wasm-aot.md`
  - Explain WASM compilation process
  - Provide deployment examples
  - Compare with Go native compilation

**Expected Results**: 5-20x speedup (browser), 10-30x speedup (WASI runtime)

---

## Phase 25: AST-Driven Formatter 🆕 **PLANNED**

**Status**: Not started | **Priority**: MEDIUM | **Estimated Tasks**: 30

### Overview

This phase delivers an auto-formatting pipeline that reuses the existing AST and semantic metadata to produce canonical DWScript source, accessible via the CLI (`dwscript fmt`), editors, and the web playground.

**Architecture Flow**: DWScript Source → Parser → AST → Formatter → Formatted DWScript Source

**Benefits**: Consistent code style, automated formatting, editor integration, playground support

### Phase 25.1: Specification & AST/Data Prep (7 tasks)

- [x] 25.1.1 Capture formatting requirements from upstream DWScript (indent width, begin/end alignment, keyword casing, line-wrapping) and document them in `docs/formatter-style-guide.md`.
- [x] 25.1.2 Audit current AST nodes for source position fidelity and comment/trivia preservation; list any nodes lacking `Pos` / `EndPos`.
- [x] 25.1.3 Extend the parser/AST to track leading and trailing trivia (single-line, block comments, blank lines) without disturbing semantic passes.
  - **Implemented**: Comment preservation infrastructure (lexer + AST structures)
  - **Lexer** (`internal/lexer/lexer.go`):
    - Added `preserveComments` flag to control comment tokenization
    - New methods: `SetPreserveComments()`, `readLineComment()`, `readBlockComment()`, `readCStyleComment()`
    - Modified `NextToken()` to return COMMENT tokens when enabled
    - Supports all 4 comment styles: `//`, `{ }`, `(* *)`, `/* */`
  - **AST** (`pkg/ast/comment.go`):
    - `Comment` type with text, position, and style
    - `CommentGroup` for grouping consecutive comments
    - `NodeComments` for leading/trailing comments per node
    - `CommentMap` for mapping nodes to comments (non-intrusive design)
    - Added `Comments CommentMap` field to `Program` struct
  - **Tests**:
    - `internal/lexer/comment_test.go` - 10 comprehensive tests for lexer
    - `pkg/ast/comment_test.go` - 8 tests for AST comment structures
  - **Documentation**: `docs/comment-preservation.md` - Complete guide
  - **Limitations**: Parser integration not yet complete (Phase 25.2.6)
    - ✅ Lexer can tokenize comments
    - ✅ Data structures defined
    - ❌ Parser doesn't attach comments to nodes (future work)
    - ❌ Printer can't output comments yet (future work)
- [x] 25.1.4 Define a `format.Options` struct (indent size, max line length, newline style) and default profile matching DWScript conventions.
  - **Implemented**: `pkg/printer/printer.go` defines comprehensive `Options` struct with:
    - Format types: DWScript, Tree, JSON
    - Style modes: Detailed, Compact, Multiline
    - Indentation control (width, spaces vs tabs)
    - Position and type info toggles
  - **Implemented**: `pkg/printer/styles.go` provides helper functions for common configurations
- [~] 25.1.5 Build a formatting test corpus in `testdata/formatter/{input,expected}` with tricky constructs (nested classes, generics, properties, preprocessor).
  - **Status**: Basic tests in `pkg/printer/printer_test.go` cover common cases
  - **TODO**: Create comprehensive test corpus with edge cases
- [x] 25.1.6 Add helper APIs to serialize AST back into token streams (e.g., `ast.FormatNode`, `ast.IterChildren`) to keep formatter logic decoupled from parser internals.
  - **Implemented**: `pkg/printer/printer.go` provides core printing infrastructure
  - **Implemented**: `pkg/printer/dwscript.go` contains node-specific formatting for all major AST types
- [~] 25.1.7 Ensure the semantic/type metadata needed for spacing decisions (e.g., `var` params, attributes) is exposed through lightweight inspector interfaces to avoid circular imports.
  - **Status**: Basic metadata support exists; AST nodes contain type annotations
  - **TODO**: May need additional helpers for complex spacing rules

### Phase 25.2: Formatter Engine Implementation (10 tasks)

- [ ] 25.2.1 Create `formatter` package with a multi-phase pipeline: AST normalization → layout planning → text emission.
- [ ] 25.2.2 Implement a visitor that emits `format.Node` instructions (indent/dedent, soft break, literal text) for statements and declarations, leveraging AST shape rather than raw tokens.
- [ ] 25.2.3 Handle block constructs (`begin...end`, class bodies, `case` arms) with indentation stacks so nested scopes auto-align.
- [ ] 25.2.4 Add expression formatting that respects operator precedence and inserts parentheses only when required; reuse existing precedence tables.
- [ ] 25.2.5 Support alignment for parameter lists, generics, array types, and property declarations with configurable wrap points.
- [ ] 25.2.6 Preserve user comments: attach leading comments before the owning node, keep inline comments after statements, and maintain blank-line intent (max consecutives configurable).
- [ ] 25.2.7 Implement whitespace normalization rules (single spaces around binary operators, before `do`/`then`, after commas, etc.).
- [ ] 25.2.8 Provide idempotency guarantees by building a golden test that pipes formatted output back through the formatter and asserts stability.
- [ ] 25.2.9 Expose a streaming writer that emits `[]byte`/`io.Writer` output to keep the CLI fast and low-memory.
- [ ] 25.2.10 Benchmark formatting of large fixtures (≥5k LOC) and optimize hot paths (string builder pools, avoiding interface allocations).

### Phase 25.3: Tooling & Playground Integration (7 tasks)

- [~] 25.3.1 Wire a new CLI command `dwscript fmt` (and `fmt -w`) that runs the formatter over files/directories, mirroring `gofmt` UX.
  - [x] 25.3.1.1 Create `cmd/dwscript/cmd/fmt.go` with basic command structure
  - [x] 25.3.1.2 Add `-w` flag to write formatted output back to files
  - [x] 25.3.1.3 Add `-l` flag to list files that would be reformatted
  - [x] 25.3.1.4 Support formatting from stdin when no file is provided
  - [x] 25.3.1.5 Add `-d` flag to show diff instead of rewriting files
  - [x] 25.3.1.6 Support formatting multiple files and directories recursively
  - [x] 25.3.1.7 Add style flags: `--style` (detailed/compact/multiline), `--indent` (width), `--tabs` (use tabs)
  - [x] 25.3.1.8 Add tests for the fmt command
    - **Implemented**: `cmd/dwscript/cmd/fmt_test.go` with comprehensive test coverage (540 lines)
    - Tests: formatSource, FormatBytes, isFormattedCorrectly, FormatFile, style options, indentation, complex constructs, error handling
    - Benchmarks: BenchmarkFormatSource (~15µs/op), BenchmarkFormatSourceCompact (~10µs/op)
    - **Known limitation**: Printer doesn't add trailing semicolons, affecting true idempotency (needs fix in pkg/printer for task 25.2.8)
  - [ ] 25.3.1.9 Update documentation and help text
- [ ] 25.3.2 Update the WASM bridge to expose a `Format(source string) (string, error)` hook exported from Go, reusing the same formatter package.
- [ ] 25.3.3 Modify `playground/js/playground.js` to call the WASM formatter before falling back to Monaco’s default action, enabling deterministic formatting in the browser.
- [ ] 25.3.4 Add formatter support to the VSCode extension / LSP stub (if present) so editors can trigger `textDocument/formatting`.
- [ ] 25.3.5 Ensure the formatter respects partial-range requests (`textDocument/rangeFormatting`) to avoid reformatting entire files when not desired.
- [ ] 25.3.6 Introduce CI checks (`just fmt-check`) that fail when files are not formatted, and document the workflow in `CONTRIBUTING.md`.
- [ ] 25.3.7 Provide sample scripts/snippets (e.g., Git hooks) encouraging contributors to run the formatter.

### Phase 25.4: Validation, UX, and Docs (6 tasks)

- [ ] 25.4.1 Create table-driven unit tests per node type plus integration tests that read `testdata/formatter` fixtures.
- [ ] 25.4.2 Add fuzz/property tests that compare formatter output against itself round-tripped through the parser → formatter pipeline.
- [ ] 25.4.3 Document formatter architecture and extension points in `docs/formatter-architecture.md`.
- [ ] 25.4.4 Update `PLAYGROUND.md`, `README.md`, and release notes to mention the Format button now runs the AST-driven formatter.
- [ ] 25.4.5 Record known limitations (e.g., preprocessor directives) and track follow-ups in `TEST_ISSUES.md`.
- [ ] 25.4.6 Gather usability feedback (issue template or telemetry) to prioritize refinements like configurable styles or multi-profile support.

---

## Phase 26: Future Enhancements & Experimental Features [LONG-TERM]

**Status**: Not started | **Priority**: LOW | **Tasks**: Variable

### Overview

This phase collects experimental, deferred, and long-term enhancement tasks that are not critical to the core DWScript implementation but may provide value in specific use cases or future development.

**Note**: Tasks in this phase are marked as DEFERRED or OPTIONAL and should only be pursued after core phases are complete and based on user demand.

### Phase 26.1: Plugin-Based JIT [SKIP - Poor Portability]

**Status**: SKIP RECOMMENDED | **Limitation**: No Windows support, requires Go toolchain at runtime

- [ ] 26.1 Implement Go code generation from AST
  - Create `internal/codegen/go_generator.go`
  - Generate Go source code from DWScript AST
  - Map DWScript types to Go types
  - Generate function declarations and calls
  - Handle closures and scoping

- [ ] 26.2 Implement plugin-based JIT
  - Use `go build -buildmode=plugin` to compile generated code
  - Load plugin with `plugin.Open()`
  - Look up compiled function with `plugin.Lookup()`
  - Call compiled function from interpreter
  - Cache plugins to disk

- [ ] 26.3 Add hot path detection for plugin JIT
  - Track function execution counts
  - Trigger plugin compilation for hot functions
  - Manage plugin lifecycle (loading/unloading)

- [ ] 26.4 Test plugin-based JIT
  - Run tests on Linux and macOS only
  - Compare performance with bytecode VM
  - Test plugin caching and reuse

- [ ] 26.5 Document plugin approach
  - Write `docs/plugin-jit.md`
  - Explain platform limitations
  - Provide usage examples

**Expected Results**: 3-5x faster than tree-walking
**Recommendation**: SKIP - poor portability, requires Go toolchain

### Phase 26.2: Alternative Compiler Targets [EXPERIMENTAL]

- [ ] 26.6 C code generation backend
  - Transpile DWScript to C
  - Leverage existing C compilers
  - Test on embedded systems

- [ ] 26.7 Rust code generation backend
  - Transpile DWScript to Rust
  - Leverage Rust's memory safety
  - Explore performance characteristics

- [ ] 26.8 Python code generation backend
  - Transpile DWScript to Python
  - Enable rapid prototyping
  - Integration with Python ecosystem

### Phase 26.3: Advanced Optimization Research [EXPERIMENTAL]

- [ ] 26.9 Profile-guided optimization (PGO)
  - Collect runtime profiles
  - Use profiles to guide optimizations
  - Measure performance improvements

- [ ] 26.10 Speculative optimization
  - Type speculation based on runtime behavior
  - Deoptimization on type changes
  - Guard conditions

- [ ] 26.11 Escape analysis
  - Determine when objects can be stack-allocated
  - Reduce GC pressure
  - Improve performance

- [ ] 26.12 Inline caching for dynamic dispatch
  - Cache method lookup results
  - Invalidate on class changes
  - Measure performance impact

### Phase 26.4: Language Extensions [EXPERIMENTAL]

- [ ] 26.13 Async/await support
  - Design async/await syntax for DWScript
  - Implement coroutine-based execution
  - Test with concurrent workloads

- [ ] 26.14 Pattern matching
  - Extend case statements with pattern matching
  - Support destructuring
  - Type narrowing based on patterns

- [ ] 26.15 Macros/metaprogramming
  - Design macro system
  - Compile-time code generation
  - Template metaprogramming support

### Phase 26.5: Tooling Enhancements [LOW PRIORITY]

- [ ] 26.16 IDE integration beyond LSP
  - IntelliJ IDEA plugin
  - VS Code enhanced extension
  - Sublime Text package

- [ ] 26.17 Package manager
  - Design package format
  - Implement dependency resolution
  - Create package registry

- [ ] 26.18 Build system integration
  - Make/CMake integration
  - Bazel rules
  - Gradle plugin

---

## Summary

This roadmap now spans **~1000+ bite-sized tasks** across **21 phases**, organized into three tiers: **Core Language** (Phases 1-10), **Execution & Compilation** (Phases 11-18), and **Ecosystem & Tooling** (Phases 19-21).

### Core Language Implementation (Phases 1-10) ✅ MOSTLY COMPLETE

1. **Phase 1 – Lexer**: ✅ Complete (150+ tokens, 97% coverage)
2. **Phase 2 – Parser & AST**: ✅ Complete (Pratt parser, comprehensive AST)
3. **Phase 3 – Statement execution**: ✅ Complete (98.5% coverage)
4. **Phase 4 – Control flow**: ✅ Complete (if/while/for/case)
5. **Phase 5 – Functions & scope**: ✅ Complete (91.3% coverage)
6. **Phase 6 – Type checking**: ✅ Complete (semantic analysis, 88.5% coverage)
7. **Phase 7 – Object-oriented features**: ✅ Complete (classes, interfaces, inheritance)
8. **Phase 8 – Extended language features**: ✅ Complete (operators, properties, enums, arrays, exceptions)
9. **Phase 9 – Feature parity completion**: 🔄 In progress (class methods, constants, type casting)
10. **Phase 10 – API enhancements for LSP**: ✅ Complete (public AST, structured errors, visitors)

### Execution & Compilation (Phases 11-18)

11. **Phase 11 – Bytecode Compiler & VM**: ✅ MOSTLY COMPLETE (5-6x faster than AST interpreter, 116 opcodes)
12. **Phase 12 – Performance & Polish**: 🔄 Partial (profiling done, optimizations pending)
13. **Phase 13 – Go AOT Compilation**: [RECOMMENDED] Transpile to Go, native binaries (10-50x speedup)
14. **Phase 14 – WebAssembly Runtime & Playground**: ✅ MOSTLY COMPLETE (WASM build, playground, NPM package)
15. **Phase 15 – MIR Foundation**: [DEFERRED] Mid-level IR for multi-backend support
16. **Phase 16 – JavaScript Backend**: [DEFERRED] MIR → JavaScript code generation
17. **Phase 17 – LLVM Backend**: [DEFERRED/LOW PRIORITY] Maximum performance, high complexity
18. **Phase 18 – WebAssembly AOT**: [RECOMMENDED] Standalone WASM binaries for edge/server deployment

### Ecosystem & Tooling (Phases 19-21)

19. **Phase 19 – AST-Driven Formatter**: [PLANNED] Auto-formatting for CLI, editors, playground
20. **Phase 20 – Community & Ecosystem**: [ONGOING] REPL, debugging, security, maintenance
21. **Phase 26 – Future Enhancements**: [LONG-TERM] Experimental features, alternative targets

### Implementation Priorities

**HIGH PRIORITY (Core Functionality)**:
- Phase 9 (feature parity completion)
- Phase 12 (performance polish)
- Phase 13 (Go AOT compilation)
- Phase 14 remaining tasks (WASM testing)
- Phase 20 (community building, REPL, debugging)

**MEDIUM PRIORITY (Value-Add Features)**:
- Phase 18 (WASM AOT)
- Phase 19 (formatter)
- Phase 15-16 (MIR + JavaScript backend)

**LOW PRIORITY (Deferred/Optional)**:
- Phase 17 (LLVM backend - complex, high maintenance)
- Phase 21 (experimental features)

### Architecture Summary

**Execution Modes** (in order of priority):
1. **AST Interpreter** (baseline, complete) - Simple, portable
2. **Bytecode VM** (5-6x faster, mostly complete) - Good performance, low complexity
3. **Go AOT** (10-50x faster, recommended) - Excellent performance, great portability
4. **WASM Runtime** (browser/edge, mostly complete) - Web execution, good performance
5. **WASM AOT** (server/edge, recommended) - Portable binaries, sandboxed execution
6. **JavaScript Backend** (deferred) - Browser execution via transpilation
7. **LLVM Backend** (deferred) - Maximum performance, very complex

**Code Generation Flow**:
```
DWScript Source → Parser → AST → Semantic Analyzer
                                      ├→ AST Interpreter (baseline)
                                      ├→ Bytecode Compiler → VM (5-6x)
                                      ├→ Go Transpiler → Native (10-50x)
                                      ├→ WASM Compiler → Browser/WASI
                                      └→ MIR Builder → JS/LLVM Emitter (future)
```

### Project Timeline

The project can realistically take **1-3 years** depending on:

- Development pace (full-time vs part-time)
- Team size (solo vs multiple contributors)
- Completeness goals (minimal viable vs full feature parity)
- [ ] 9.16.6 Refactor type-specific nodes (ArrayLiteralExpression, CallExpression, NewExpression, MemberAccessExpression, etc.)
  - [ ] Embed appropriate base struct into array literals/index expressions
  - [ ] Embed base structs into call/member/new expressions in `pkg/ast/classes.go`/`pkg/ast/functions.go`
  - [ ] Update parser/interpreter/semantic code paths for these nodes
  - [ ] Remove duplicate helpers once embedded
