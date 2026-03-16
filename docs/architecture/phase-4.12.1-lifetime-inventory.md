# Phase 4.12.1 Lifetime Inventory

## Summary

The `yin_yang` regression was not an isolated fixture bug. It exposed a broader lifetime-model inconsistency inside the canonical evaluator/runtime path.

The evaluator currently has two different scope/binding models:

- owned bindings
  - values are installed into an environment as true scope-owned bindings
  - these bindings must retain runtime-owned values on entry and release them on cleanup
- aliased env exposure
  - values are exposed in an environment as direct views of existing runtime state
  - these bindings must not be treated like owned references during cleanup

The architecture problem is that the codebase still mixes these two models without an explicit boundary marker.

## Current Classification

### Consistent Or Improved

- assignment targets
  - `prepareValueForAssignment(...)` retains object and method-pointer ownership when the destination binding owns the value
- variable declarations
  - `VisitVarDeclStatement(...)` now routes stored values through `retainValueForBinding(...)`
- constant declarations
  - `VisitConstDecl(...)` now routes stored values through `retainValueForBinding(...)`
- user-function parameter bindings
  - `BindFunctionParameters(...)` now routes regular bound values through `retainValueForBinding(...)`
- user-function return values
  - `ExecuteUserFunction(...)` now retains returned runtime-owned values before function-scope cleanup runs
- user-function scope cleanup
  - `cleanupInterfaceReferences(...)` releases owned object/interface/method-pointer references in function environments

These changes are enough to fix the concrete premature-destruction case from `Algorithms/yin_yang`.

### Still Split Or Needs Clarification

- pushed evaluator scopes (`PushEnv` / `PopEnv`)
  - `ExecutionContext.PushEnv()` and `PopEnv()` only manage environment stack structure
  - they do not know whether the child environment contains owned bindings, aliased bindings, or a mixture
- record method scopes
  - `callRecordMethod(...)` and `callRecordStaticMethod(...)` bind parameters, record fields, and record class state into the same pushed environment
  - parameters behave like owned bindings
  - record fields and class vars behave like aliased env exposure because they are later synced back into record/class state
- helper method scopes
  - `CallHelperMethod(...)` binds parameters plus helper class vars/consts into one pushed environment
  - parameters look like owned bindings
  - helper class vars are aliased/shared state
- class property method scopes
  - `executeClassPropertyMethod(...)` mixes property parameters with class-var exposure
- expression-backed property scopes
  - `evalExpressionBackedPropertyRead(...)` exposes `Self` and bound fields in a pushed environment without a cleanup model
- lambda scopes
  - `executeLambdaDirect(...)` creates a fresh environment and binds parameters/Result, but does not currently participate in the same explicit owned-binding cleanup model as user functions

## Architectural Implication

The next cleanup step should not blindly add `cleanupInterfaceReferences(...)` to every pushed scope.

That would be wrong for environments that expose record fields, class vars, helper state, or other aliased runtime values without having retained them first.

The missing architectural primitive is an explicit distinction between:

- scope-owned bindings
- aliased/shared-state bindings

Until that exists, lifetime cleanup has to stay selective and scope-specific.

## Recommended Next Step

Phase `4.12.3` should introduce a narrow ownership-aware binding model for pushed evaluator scopes.

Likely shape:

- add a small helper or scope wrapper that can register owned bindings separately from aliased bindings
- convert the highest-risk mixed scopes first:
  - record methods
  - helper methods
  - class property methods
- keep field/class-var exposure out of generic cleanup unless those values were explicitly retained for that scope

## Verification That Motivated This Inventory

- `go test ./internal/interp -run '^TestDWScriptFixtures/Algorithms/yin_yang$' -count=1`
- `go test ./internal/interp -run 'TestMethodPointer_' -count=1`
- `go test ./internal/interp -run 'Test(RefCount|Assignment|FunctionPointer|OOP)' -count=1`

These pass after the first ownership-unification patch, but the inventory above shows that the broader pushed-scope model is still not fully unified.
