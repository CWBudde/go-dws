# Phase 4.7 Verification

Phase 4.7 is the proof pass for the Phase 4 interpreter/evaluator split cleanup.

## Final boundary

- `internal/interp/new.go` is the canonical production construction entry point.
- `internal/interp/runner` delegates to `interp.NewWithOptions(...)`.
- `ExecutionContext` is the canonical owner of per-run mutable execution state.
- `internal/interp/contracts` remains as a neutral coordination package, but the
  old evaluator callback interfaces are gone.
- Runtime object and property dispatch now prefer canonical lookup APIs
  (`LookupMethod`, `LookupProperty`, `FieldExists`) over legacy map exposure.

## Current metrics

- Internal callback interfaces remaining: `0`
  Removed: `CoreEvaluator`, `OOPEngine`, `DeclHandler`
- Legacy construction bridge methods remaining: `0`
  Removed: `SetFocusedInterfaces`, production `SetRuntimeBridge` wiring
- Production `internal/interp` imports of `internal/interp/evaluator`: `1`
  Allowed entry point: `internal/interp/new.go`
- Runner-level evaluator imports: `0`
  `internal/interp/runner` delegates to `interp.NewWithOptions(...)`
- Bridge-only evaluator seam files remaining: `0`
- Runtime object method lookup via legacy `GetMethodsMap()` in production paths: `0`
- Interpreter struct fields after Phase 4 cleanup: `6`

## Regression coverage added in Phase 4.7

- Assignment cluster:
  field-backed property write/read and class-var access through the final engine path
- User-function cluster:
  direct function execution plus `var` parameter mutation through the final engine path
- Declaration cluster:
  nested class declaration/instantiation with semantic info and final runtime construction

## Verification commands

Focused architectural/regression checks:

```bash
go test ./internal/interp -run 'TestInterpDoesNotImportEvaluator|TestConstructionDoesNotReferenceLegacyBridgeWiring|TestNoLegacyCallbackInterfaceDeclarationsRemain|TestPhase4Regression'
```

Focused package verification used during Phase 4.7 work:

```bash
go test ./internal/interp/runtime ./internal/interp/evaluator ./internal/interp/types ./internal/interp
```

Full suite:

```bash
go test ./...
```
