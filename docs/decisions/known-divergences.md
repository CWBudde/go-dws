# Known divergences and declined work

Behaviour where go-dws deliberately does not match an upstream fixture, and work that was measured
and declined. Each entry states why and what would reopen it. Scope-level exclusions (host
libraries, COM, databases, …) are in [`out-of-scope.md`](out-of-scope.md); the string model is in
[`string-encoding.md`](string-encoding.md); the bytecode VM is in [`bytecode-vm.md`](bytecode-vm.md).

## Fixtures that cannot pass as written

| Fixture | Why | Reopen when |
| --- | --- | --- |
| `SimpleScripts/include_expr` (`{$I %FILE%}` value substitution) | The expected output hardcodes the Delphi runner's paths (`Test\include_expr.pas`, `*MainModule*`), and neither runner normalizes paths. `%FUNCTION%` is not knowable at lex time. | The harness normalizes paths. |
| `SimpleScripts/for_in_str`, `for_in_str2` | UTF-16 surrogate iteration; go-dws strings are UTF-8. See [`string-encoding.md`](string-encoding.md). | Never, by design. |
| `Memory/*` leak assertions | Upstream checks `exec.ObjectCount = 0` and external-object counts, i.e. DWScript's reference counting, at a point where Go's GC has not necessarily run. Only "compiles and prints nothing" is portable. | Never, by design. |
| `FailureScripts/static_methods` | Upstream's expectation omits the `Compile Error` line although `{$FATAL}` is present. The only difference from the three cases that do report it is that its `{$FATAL}` is not at column 1 — a 5/5 correlation with no plausible mechanism, so it is not encoded. | The upstream emit site explains it. |
| `FailureScripts/class_deprecated` (4 of 8 warnings) | For a declaration's type annotation (`FField : TBase`, `property O : TOther`, …) upstream anchors the deprecation warning two columns before the type name; every expression-position use anchors at the identifier. All samples are written `: T`, so "the colon" and "type name minus two" are indistinguishable. Implementing the colon reading threads a colon position through nine `warnDeprecatedResolvedType` call sites. | The upstream emit site settles which reading is right. |

## Behaviour without a fixture that demands it

Each was measured against the corpus and has zero fixture yield. Reopen only on new evidence (a
fixture, a user report, or an upstream test).

- Unused-private hints for record fields and class vars (no usage tracking).
- Record-type metaclass member access via the type name.
- Class invariants: parsed into `ClassDecl.Invariants`, never evaluated.
- Type-parameter constraints `<T: TObject>`: parsed, ignored.
- Function-pointer `= nil` comparison still reports "operator = requires comparable types".
- A for-in loop variable is not checked against a set's element type.
- Subrange bounds are not checked at compile time; no fixture declares a subrange type.
- `ConditionalDefined(s)` always folds to `False`: `{$DEFINE}` symbols live in preprocessor state
  the analyzer cannot reach. Argument validation is complete.

## Known analyzer limitations

- **Source-order abstract checks.** `var c := TC.Create;` written *before* the abstract ancestor's
  declaration misses the abstract-instantiation error, because class checks run in source order.
  The multi-pass design in [`../architecture/semantic-passes.md`](../architecture/semantic-passes.md)
  would fix it; nothing currently requires it.
- **Unused-private hints next to a compile error** are suppressed wholesale
  (`internal/semantic/unused_warnings.go`). `OverloadsFail/overloads_not_implem` contradicts this
  but fails to parse for unrelated reasons; revisit once it parses.

## Hint configuration

Case-mismatch hint parity is pursued only for categories whose upstream runner configuration is
verified (see "Evidence for case-mismatch hint settings" in
[`../../testdata/fixtures/README.md`](../../testdata/fixtures/README.md)). For any other category,
recover the runner configuration first.

## Refactors not worth doing

Measured and declined:

- Clearing source `// TODO` markers as a task in itself: the real ones are tracked in `PLAN.md`.
- Converting panics to errors: panics are not used for control flow.
- Splitting `visitor_statements.go` / `visitor_declarations.go` by size alone: each is one
  cohesive dispatch.
