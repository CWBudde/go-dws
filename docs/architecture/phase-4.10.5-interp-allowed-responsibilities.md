# Phase 4.10.5 Allowed `internal/interp` Responsibilities

This note defines what `internal/interp` is still allowed to own after the
Phase 4 cleanup.

The goal is to make the shell/core boundary explicit enough that future work can
tell the difference between:

- legitimate engine-shell orchestration
- runtime primitives
- forbidden reintroduction of evaluator-owned AST semantics

## Allowed Responsibilities

`internal/interp` is allowed to own the following categories only.

### 1. Production bootstrap

Examples:

- `new.go`
- environment creation
- type-system construction and wiring
- shared `EngineState` initialization
- refcount manager/destructor wiring

Reason:

This is the engine entry point. Construction belongs at the shell boundary.

### 2. Public engine-facing API and orchestration

Examples:

- `Interpreter.Eval`
- `Interpreter.EvalWithExpectedType`
- exception/query helpers exposed to CLI or embedding surfaces
- source/semantic-info configuration

Reason:

The shell owns the public runtime facade, but it must delegate AST semantics to
the evaluator.

### 3. Host and unit integration

Examples:

- external-function registration/invocation plumbing
- unit registry integration and load/init coordination
- Go callback / FFI entry points

Reason:

These are embedding concerns, not core AST semantics.

### 4. Declaration/bootstrap mutation of registries and metadata

Allowed remaining `Interpreter.eval*` methods in this category:

- `evalFunctionDeclaration`
- `evalClassMethodImplementation`
- `evalRecordMethodImplementation`
- `evalClassDeclaration`
- `evalInterfaceDeclaration`
- `evalOperatorDeclaration`
- `evalHelperDeclaration`
- `evalTypeDeclaration`
- `evalEnumDeclaration`

Reason:

These paths primarily mutate runtime metadata, helper/class/record registries,
or environment exposure during declaration/bootstrap processing.

### 5. Narrow runtime helper primitives used by shell-owned integration

Examples:

- lvalue/reference helpers
- small value-operation helpers that are not AST traversal entry points
- call-stack/error-location support used by shell-owned integration paths

Reason:

Some low-level helpers are shared utility code, not evaluator ownership
violations.

## Not Allowed In `internal/interp`

The following categories are no longer allowed to reappear in `internal/interp`.

### 1. Production AST statement execution

Examples:

- program/block/if/case/loop/try/raise execution
- assignment execution
- break/continue/exit handling

Owner:

- `internal/interp/evaluator`

### 2. Production AST expression execution

Examples:

- method-call dispatch
- property read/write execution
- type casts
- `as` expressions
- `Default(...)`
- record method execution
- helper method/property execution

Owner:

- `internal/interp/evaluator`

### 3. Shadow evaluator entry points or deleted execution files

Examples already removed:

- `objects_methods.go`
- `objects_properties.go`
- `functions_calls.go`
- `functions_records.go`
- `statements_*.go`
- `expressions_complex.go`
- `user_function_callbacks.go`

These must not return.

## Current Allowed `Interpreter.eval*` Surface

After `4.10.3` and `4.10.4`, the remaining interpreter `eval*` surface is:

- `evalFunctionDeclaration`
- `evalClassMethodImplementation`
- `evalRecordMethodImplementation`
- `evalClassDeclaration`
- `evalInterfaceDeclaration`
- `evalOperatorDeclaration`
- `evalHelperDeclaration`
- `evalTypeDeclaration`
- `evalEnumDeclaration`

This is the intended allowlist for the shell-owned declaration/bootstrap layer.

Everything else should be treated as suspicious by default and either:

- moved to evaluator/runtime
- justified explicitly as shell orchestration
- or deleted as residue

