# Phase 4.10.1 Live Interp Execution Inventory

This note inventories the remaining live or apparently-live `Interpreter.eval*`
surface after Phase 4.9 removed the dead shadow-execution clusters.

The purpose is not just to list methods. It is to classify which remaining
interpreter-owned execution paths are:

- legitimate shell/orchestration responsibilities
- residual production AST semantics that should move to `evaluator`
- lower-level runtime behaviors that should become shared runtime primitives
- stale duplicates that should be deleted rather than documented as seams

## Summary

The remaining surface is smaller, but it is not architecturally uniform.

There are three materially different buckets:

1. `keep in interp`
   - declaration/bootstrap/orchestration paths that mutate registries, class
     metadata, helper metadata, or engine-facing state
2. `move to evaluator`
   - still-live production execution semantics for AST-driven method/property/
     helper behavior
3. `delete or re-home`
   - residual duplicate execution helpers that are no longer the production AST
     path, or logic that should become a runtime primitive rather than stay on
     `Interpreter`

The key architectural finding is that the remaining problem is no longer dead
statement/control-flow duplication. It is a smaller semantic island around OOP
dispatch, properties, helper execution, and a few stale expression helpers.

## Inventory

### Keep In `interp`

These are still plausibly shell-owned responsibilities because they primarily
register declarations, wire metadata, or coordinate engine/runtime structures.

- `evalFunctionDeclaration`
  - registers global functions and routes method implementations into class or
    record registries
- `evalClassMethodImplementation`
  - updates class method maps, constructor/destructor slots, overloads, and VMT
    propagation
- `evalRecordMethodImplementation`
  - updates record runtime metadata and overload maps
- `evalClassDeclaration`
  - builds runtime class metadata, merges partial classes, resolves inheritance,
    and registers the final class structure
- `evalInterfaceDeclaration`
  - same category as class declaration: runtime metadata registration
- `evalOperatorDeclaration`
  - registry-oriented declaration mutation
- `evalHelperDeclaration`
  - helper metadata registration and initialization
- `evalTypeDeclaration`
  - mixed, but most of the current behavior is declaration/type-registry mutation
- `evalEnumDeclaration`
  - mixed, but mostly type registration plus global constant exposure

These are still execution-like, but they look more like declaration-time shell
work than expression/statement runtime semantics.

## Move To `evaluator`

These are still-live production semantics and remain the clearest mismatch with
the desired ownership boundary.

- `evalMethodCall`
  - still performs real AST method-call execution logic, including unit/class/
    record dispatch decisions, argument evaluation, recursion checks, and method
    body execution
- `evalRecordMethodCall`
  - still executes record method semantics directly
- `evalPropertyRead`
  - still owns runtime property getter execution semantics
- `evalPropertyWrite`
  - still owns runtime property setter execution semantics
- `evalIndexedPropertyRead`
  - still owns indexed property getter execution semantics
- `evalIndexedPropertyWrite`
  - still owns indexed property setter execution semantics
- `evalBuiltinHelperMethod`
  - interpreter-side built-in helper execution duplicates evaluator-owned helper
    execution direction
- `evalBuiltinHelperProperty`
  - same issue for helper property access

These methods are not just metadata plumbing. They execute user-visible DWScript
semantics. If Phase 4’s target is a single owner for AST execution semantics,
this cluster should not remain on `Interpreter` long term.

## Delete Or Re-Home

These methods do not look like justified long-term shell seams.

- `evalTypeCast`
  - evaluator already owns production type-cast execution through
    `visitor_expressions_functions.go` and `type_casts.go`
  - current interpreter-side copy appears to be stale duplicate logic
- `evalDefaultFunction`
  - evaluator already owns production `Default(...)` execution
  - current interpreter-side copy appears to be stale duplicate logic

These should be deleted or moved behind a smaller shared primitive if any
non-evaluator caller still needs pieces of their logic.

## Architectural Problems Revealed

### 1. OOP/property/helper execution is still a semantic island

Phase 4.9 removed dead statement/control-flow duplication, but property access,
method dispatch, record method execution, and helper execution still live in
`interp` as active runtime semantics.

That means the architecture is now split like this:

- statements/loops/basic expression traversal: `evaluator`
- significant OOP dispatch and property/helper behavior: `interp`

This is much smaller than before, but it is still split semantic ownership.

### 2. Declaration ownership is still mixed with execution helpers

Some declaration paths are legitimate shell mutation, but files like
`declarations.go`, `enum.go`, and `type_alias.go` still mix:

- metadata registration
- constant-expression evaluation
- environment exposure

That means `4.10` should not blindly “move declarations to evaluator” without
first separating registration/orchestration from expression semantics.

### 3. Dead duplicates still exist outside the statement cluster

`evalTypeCast` and `evalDefaultFunction` are the clearest examples. They should
not be treated as permanent seams just because they still exist.

## Recommended Next Steps

### 4.10.2 Decision Pass

Document the intended end state explicitly:

- `interp` keeps declaration/bootstrap/orchestration responsibilities
- `evaluator` owns production AST execution semantics
- `runtime` owns reusable execution primitives

### 4.10.3 Migration Pass

Move the live semantic island in this order:

1. built-in helper method/property execution
2. property read/write execution
3. record method execution
4. method-call dispatch execution

This order is deliberate: helper/property execution is narrower and will likely
shake out the runtime primitives needed before migrating full method dispatch.

### 4.10.x Cleanup Pass

Delete stale duplicate helpers:

- `evalTypeCast`
- `evalDefaultFunction`

Only keep them if a concrete non-evaluator production caller still requires
them. If not, they should not survive the phase.
