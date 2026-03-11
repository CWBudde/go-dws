# Phase 4.11 Neutral Boundary Audit

This note records the final `4.11` audit of:

- `internal/interp/contracts`
- runtime metadata escape hatches that remained after the `4.9` / `4.10`
  execution-ownership cleanup

## Runtime Metadata Escape Hatches

### Reduced in this phase

- `runtime.IClassInfo` no longer requires:
  - `GetFieldsMap()`
  - `GetMethodsMap()`
- `contracts.ClassMetaValue.GetClassInfo()` no longer returns `any`
  - it now returns `runtime.IClassInfo`

Why this matters:

- the runtime-facing class interface no longer exposes legacy field/method-map
  access as part of the required shell/runtime contract
- evaluator no longer needs `.(runtime.IClassInfo)` type assertions after
  calling `GetClassInfo()`

### What still remains

- `ClassInfo` still keeps `GetFieldsMap()` / `GetMethodsMap()` as concrete
  methods for direct internal/test use
- `FieldExists(...)` still remains on `runtime.IClassInfo`

Why this is still acceptable for now:

- `FieldExists(...)` is still used by live object-field compatibility paths
- the legacy maps no longer shape the runtime interface boundary itself

## `contracts` Audit

The remaining `contracts` surface is now:

- `Value`
  - justified alias to `runtime.Value`
- `ClassMetaValue`
  - justified neutral interface for evaluator-facing class metadata access
  - improved in this phase by typing `GetClassInfo()` as `runtime.IClassInfo`
- `ExternalFunctionRegistry`
  - justified shell/integration boundary
- `EngineState`
  - justified shared coordination state

Removed from `contracts` in this phase:

- evaluator-only user-function callback types and `UserFunctionCallbacks`
- dead `FunctionPointerMetadata`

## Result

`contracts` is now much closer to the intended end state:

- shared engine/runtime coordination only
- no interpreter-specific callback bundle
- no dead transport structs
- fewer `any`-typed adapter escape hatches

That leaves `EngineState.ExternalFunctionCaller` as the main intentionally
temporary seam still worth revisiting later, but the package is no longer a
general migration holding area.

