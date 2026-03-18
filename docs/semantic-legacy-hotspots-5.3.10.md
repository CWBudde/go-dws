# Semantic Legacy Hotspot Audit (5.3.10)

This note captures the remaining legacy `addError(...)` hotspots in the semantic
files called out by `PLAN.md` task `5.3.10` after the structured migrations from
`5.3.4` through `5.3.9`.

## Snapshot

Counted with:

```bash
rg -n "addError\(" internal/semantic/analyze_properties.go \
  internal/semantic/analyze_method_calls.go \
  internal/semantic/analyze_function_calls.go \
  internal/semantic/analyze_classes.go \
  internal/semantic/analyze_statements.go
```

Remaining call-site counts:

| File | Remaining `addError(...)` sites | Audit result |
| --- | ---: | --- |
| `internal/semantic/analyze_properties.go` | 0 | Fully migrated for the property declaration/use-site buckets handled in `5.3.7` and `5.3.8`. |
| `internal/semantic/analyze_method_calls.go` | 17 | Still dominated by method-call arity/type validation and overload invariant checks. |
| `internal/semantic/analyze_function_calls.go` | 54 | Largest remaining call-site cluster; mostly generic call-shape, overload, builtin-signature, constructor, and cast diagnostics. |
| `internal/semantic/analyze_classes.go` | 10 | Mostly constructor/static-record-call validation and overload invariant checks. |
| `internal/semantic/analyze_statements.go` | 49 | Mostly statement/declaration/assignment/control-flow diagnostics; property-specific use-site paths are already migrated. |

## Classification

### `internal/semantic/analyze_properties.go`

- No remaining legacy hotspots.
- The file can be removed from the active migration watchlist unless a new raw
  path is introduced later.

### `internal/semantic/analyze_classes.go`

Remaining raw diagnostics fall into three buckets:

1. Constructor availability and overload-selection failures in `new` analysis.
2. Constructor visibility and constructor argument mismatch diagnostics.
3. Record static method call argument mismatch and internal overload invariant errors.

This file no longer has open property/member/visibility buckets from the
completed `5.3.5` through `5.3.8` work. The remaining work is mostly
constructor/call-shape oriented and should move with the call/constructor
structured migration rather than another member-access slice.

### `internal/semantic/analyze_method_calls.go`

Remaining raw diagnostics fall into four buckets:

1. Method/class-method/helper/set-method argument count mismatches.
2. Method/class-method/helper/set-method argument type mismatches.
3. Constructor overload arity/type mismatch diagnostics on method-call syntax.
4. Internal invariant failures after overload selection (`selected.Type` not being
   a `*types.FunctionType`).

These are call-validation buckets, not access-resolution buckets. They should be
migrated together with the generic call diagnostics in
`internal/semantic/analyze_function_calls.go` to keep arity/type wording aligned.

### `internal/semantic/analyze_function_calls.go`

Remaining raw diagnostics fall into six buckets:

1. Non-callable callee and invalid call-shape diagnostics.
2. Generic function/method arity, argument-type, and `var`-parameter lvalue checks.
3. Builtin signature validation (`Assert`, `Insert`, higher-order array helpers,
   stack helpers, and similar builtins not yet on structured errors).
4. Overload-resolution and constructor-call failures that still emit legacy text.
5. Cast diagnostics, including DWScript-specific cast wording.
6. Internal invariant failures after overload resolution.

This file is the largest remaining migration surface. It likely needs a dedicated
follow-up bucket rather than piecemeal edits spread across unrelated tasks.

### `internal/semantic/analyze_statements.go`

Remaining raw diagnostics fall into five buckets:

1. Declaration and constant-evaluation diagnostics:
   variable declaration shape, unknown declared types, inference failures, and
   compile-time-constant validation.
2. Generic assignment and return-type mismatch diagnostics outside the already
   migrated property-specific paths.
3. Assignment-target validity and read-only-variable/constant diagnostics.
4. Control-flow validation:
   boolean conditions, ordinal/range/case compatibility, enumerable checks, and
   loop-step constraints.
5. Flow-control placement diagnostics for `break`, `continue`, and `exit`.

Property-specific use-site diagnostics are no longer part of this backlog; the
remaining statement work is broader statement semantics.

## Recommended Next Mapping

The remaining hotspots cluster into these future migration slices:

| Future slice | Files |
| --- | --- |
| Call / overload / arity / argument-type structured diagnostics | `analyze_function_calls.go`, `analyze_method_calls.go`, `analyze_classes.go` |
| Constructor-specific structured diagnostics | `analyze_classes.go`, `analyze_function_calls.go`, `analyze_method_calls.go` |
| Statement / assignment / control-flow structured diagnostics | `analyze_statements.go` |
| Internal invariant / impossible-state handling | all three call-focused files, likely not user-facing structured buckets |

The key audit outcome is that `analyze_properties.go` is no longer a remaining
hotspot for this phase; the active backlog has shifted to call validation and
general statement semantics.
