# Phase 4.9.1 Interpreter Execution Inventory

This note inventories the remaining interpreter-side execution helpers in
`internal/interp` and classifies each as one of:

- `delete`: dead duplicate or test-only migration residue that should not survive
  the evaluator-owned architecture
- `migrate`: still-live semantic island that belongs in evaluator/runtime rather
  than in interpreter-owned execution code
- `keep`: shell/integration responsibility that may remain in `internal/interp`
  if Phase 4.9.7 confirms it is an intentional steady-state seam

This is an inventory, not the final cleanup patch.

## Method

- Enumerated `func (i *Interpreter) eval*`, `Call*`, and `Execute*` helpers in
  `internal/interp/*.go`
- Checked interpreter-internal references with `rg`
- Compared overlapping helpers against the active evaluator visitors and runtime
  helpers

## Summary

The inventory splits into three buckets:

1. A large `delete` bucket of dead shadow evaluators, especially for
   statements/control flow and declarations
2. A smaller `migrate` bucket of still-live semantic islands, especially typed
   array-literal evaluation
3. A narrow `keep (tentative)` bucket of engine-shell integration seams such as
   external-function dispatch and function-pointer/user-function integration

## Delete

These helpers overlap evaluator-owned AST semantics and should not survive the
 finished architecture. Most are unreferenced in production or are referenced
 only by other legacy interpreter helpers or direct internal-package tests.

### Statement / Control / Exception Duplicates

- `evalProgram`
- `evalVarDeclStatement`
- `evalExternalVarDecl`
- `evalVarDeclInitializer`
- `evalConstDecl`
- `evalAssignmentStatement`
- `evalSimpleAssignment`
- `evalRecordPropertyWrite`
- `evalMemberAssignment`
- `evalIndexAssignment`
- `evalBlockStatement`
- `evalIfStatement`
- `evalCaseStatement`
- `evalWhileStatement`
- `evalRepeatStatement`
- `evalForStatement`
- `evalForInStatement`
- `evalForInArray`
- `evalForInSet`
- `evalForInString`
- `evalForInTypeMeta`
- `evalTryStatement`
- `evalExceptClause`
- `evalRaiseStatement`

Rationale:

- Normal AST execution enters through `Interpreter.Eval()` and delegates to the
  evaluator.
- The active evaluator already owns `VisitProgram`, `VisitVarDeclStatement`,
  `VisitConstDecl`, `VisitAssignmentStatement`, `VisitBlockStatement`,
  `VisitIfStatement`, `VisitCaseStatement`, `VisitWhileStatement`,
  `VisitRepeatStatement`, `VisitForStatement`, `VisitForInStatement`,
  `VisitTryStatement`, and `VisitRaiseStatement`.
- Several interpreter-side versions are completely unreferenced outside their
  own file comments or other legacy interpreter helpers.
- The `eratosthene` failure proved that these dead copies can silently remain
  more correct than the live evaluator path.

### Declaration Duplicates

- `evalFunctionDeclaration`
- `evalClassMethodImplementation`
- `evalRecordMethodImplementation`
- `evalClassDeclaration`
- `evalInterfaceDeclaration`
- `evalOperatorDeclaration`
- `evalEnumDeclaration`
- `evalTypeDeclaration`
- `evalHelperDeclaration`

Rationale:

- Evaluator visitors already own declaration execution.
- Remaining interpreter-side uses are legacy internal-package tests or the old
  interpreter declaration cluster.
- These should either be deleted outright or have tests migrated to the
  evaluator/final engine path before deletion.

### Expression / Call / OOP Duplicates That No Longer Belong In The Active Path

- `evalCallExpression`
- `evalMethodCall`
- `evalRecordMethodCall`
- `evalMemberAccess`
- `evalIndexExpression`
- `evalNewExpression`
- `evalNewArrayExpression`
- `evalAsExpression`
- `evalAddressOfExpression`
- `evalFunctionPointer`
- `evalRecordLiteral`
- `evalVariantBinaryOp`
- `evalTypeCast`
- `evalDefaultFunction`
- `evalBuiltinHelperMethod`
- `evalBuiltinHelperProperty`
- `ExecuteConstructor`
- `ExecuteRecordPropertyRead`
- `CallQualifiedOrConstructor`
- `CallMethod`
- `CallInheritedMethod`
- `ExecuteMethodWithSelf`
- `CallImplicitSelfMethod`
- `CallRecordStaticMethod`

Rationale:

- Many of these are only referenced by the interpreter’s old call/method/helper
  cluster, not by the evaluator’s active runtime flow.
- Some names still appear in evaluator files only in comments describing the old
  migration path.
- Before deletion, their remaining real call sites must be confirmed to be
  dead; first-pass inventory suggests most of them are.

## Migrate

These helpers are still live enough to matter, but they represent evaluator
 semantics that should not remain interpreter-owned.

- `evalArrayLiteral`
- `evalArrayLiteralWithExpected`
- `evalSetLiteral`

Rationale:

- `Interpreter.EvalWithExpectedType()` still special-cases array literals and
  routes them into interpreter-owned execution.
- Interpreter-side declaration/assignment helpers still use
  `evalArrayLiteralWithExpected`.
- Nested array/set literal typing also still flows through this interpreter
  cluster.

Expected destination:

- evaluator-owned typed literal helpers, or runtime/type helpers invoked from
  evaluator visitors

## Keep (Tentative, Pending 4.9.7)

These are not obviously dead duplicates. They look more like engine-shell or
 cross-boundary integration seams and should be reviewed explicitly in 4.9.7.

- `CallExternalFunction`
- `ExecuteFunctionPointerCall`
- `CallFunctionPointer`
- `CallUserFunction`

Rationale:

- `CallExternalFunction` is the production callback currently wired from
  `EngineState` for Go-registered external function execution.
- `ExecuteFunctionPointerCall` and related function-pointer plumbing still sit
  at the boundary where evaluator execution needs interpreter-owned closure/self
  setup or compatibility behavior.
- These may remain legitimate shell responsibilities, but they should be
  explicitly defended as such rather than surviving by accident.

## Architectural Problems Revealed

### 1. Dead Shadow Evaluators Still Exist In Large Clusters

The biggest problem is not one function, but entire interpreter-side semantic
 clusters that survive after the evaluator became the production execution core.

This is already enough to create semantic drift:

- dead interpreter loop semantics can remain more correct than live evaluator
  loop semantics
- tests can still target legacy helper entry points directly
- boundary tests currently do not enforce single ownership of AST execution

### 2. There Is Still A Smaller Live Semantic-Island Problem

The old statement/declaration cluster is mostly dead, but typed literal
 evaluation is not dead. `evalArrayLiteralWithExpected` and related helpers
 remain genuinely live and still execute semantics interpreter-side.

That means the architecture is not only suffering from stale duplication; it
 also still has unfinished semantic migration in a few focused places.

### 3. Some Interpreter Methods Need Explicit Steady-State Justification

The remaining callback-shaped helpers around external functions and
 function-pointer/user-function execution are not automatically wrong, but they
 should be treated as intentional shell seams only if Phase 4.9.7 documents why
 they belong in `internal/interp`.

If that justification is weak, they should be migrated too.

## Recommended Follow-On Order

1. Delete the dead statement/control/exception duplicates first
2. Migrate typed array/set literal evaluation into evaluator-owned paths
3. Re-check the old call/method/property/helper cluster after step 2
4. Review the remaining shell seams together with the `contracts` audit in
   Phase 4.9.7
