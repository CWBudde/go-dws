# go-dws — Work Plan

> Rewritten 2026-09-06. This file lists **open work only**. Completed work and the reasoning
> behind it live in [`docs/history/progress-log-2026-07.md`](docs/history/progress-log-2026-07.md);
> the measured audits behind the priorities are
> [`docs/history/CODEBASE_REVIEW_2026-07.md`](docs/history/CODEBASE_REVIEW_2026-07.md) (July 2026)
> and [`docs/architecture/audit-2026-09.md`](docs/architecture/audit-2026-09.md) (September 2026).
> Status numbers are generated, not hand-kept. Documentation index: [`docs/README.md`](docs/README.md).

## 0. Status snapshot

**Headline (2026-09-08):** Go harness **878 / 1,928 scored = 46%**. Last full CLI
measurement (2026-09-06, after Phase 1): **871 / 1,928 = 45%**; refresh with
`just fixture-report` for current CLI numbers. Both use the shared compile pipeline and
scoring rules. `*Fail` error-detection suites **111 / 647 = 17%**.

Where the truth lives:

| Source | What it is | How to refresh |
| --- | --- | --- |
| `testdata/fixtures/TEST_STATUS.md` | generated per-category pass/fail table (harness) | `just fixture-update` |
| `testdata/fixtures/baselines.json` | per-category pass-count floor that CI enforces | `just fixture-update` after an intentional improvement |
| `just fixture-report` | end-to-end CLI numbers (`cmd/fixture-report` running `bin/dwscript run`) | `just build` first, or the numbers are stale |

Rules for this document:

- An item is closed only by a **passing fixture** (or a test that exercises the real user-facing path).
  Closed items are deleted from this file; their story goes to `docs/history/progress-log-<date>.md`.
- Ratchet `baselines.json` after every improvement (`just fixture-update`).
- The ~200 fixtures in host-library categories listed in
  [`docs/decisions/out-of-scope.md`](docs/decisions/out-of-scope.md) (DataBaseLib, COMConnector,
  CryptoLib, GraphicsLib, WebLib, TabularLib, TimeSeriesLib, DOMParser, Linq, LinqJSON, ClassesLib,
  DelegateLib, SystemInfoLib, IniFileLib, FunctionsFile, FunctionsRTTI, BigInteger,
  FunctionsMathComplex/3D) are excluded from every target below.
- Where the remaining failures are (1,050 total): FailureScripts 421, SimpleScripts 104,
  host-library categories ~200, everything else < 40 per category.

Legend: `[ ]` open · `[~]` partially done, remainder listed · ⏸️ gated, do not start ·
✋ won't-fix, with the decision record. Size: S (hours), M (days), L (week+).

---

## 1. Measurement & tooling (T)

No open items. T1–T6 closed 2026-09-06 (one compile pipeline for CLI and harness, `run
--diagnostics=plain|pretty`, `--test-envelope`, `--compile-only`, self-rebuilding
`fixture-report` with a stale-binary guard, helper-spec parity test, unscored-variant docs);
see [`docs/history/progress-log-2026-09.md`](docs/history/progress-log-2026-09.md). New tooling
items go here.

---

## 2. Architecture refactoring — required (A)

A5 now consumes resolved semantic type objects for annotation-based execution; its
unit-aware frontend prerequisite is complete. A6's class metadata migration and typed
class registry are complete; A9 has migrated 78 further builtin names to registry validation.
The remaining work is listed below (2026-09-08).

A2–A4, A8 and A10 closed 2026-09-07; A11's experimental API/help labels are complete.
See [`docs/history/progress-log-2026-09.md`](docs/history/progress-log-2026-09.md).
A9 has a registry-backed return-type lookup and ordinary call analyzer; its remaining
specialized handlers are listed below.

Evidence for every item, with file:line references and measurements, is in
[`docs/architecture/audit-2026-09.md`](docs/architecture/audit-2026-09.md). Items are ordered so
that each one shrinks the blast radius of the next. The target architecture is
[`docs/architecture/interp-evaluator-steady-state.md`](docs/architecture/interp-evaluator-steady-state.md).

- **A5** `[~]` L — **Delete the evaluator's remaining string-based type resolution.**
  Annotation-based execution now consumes resolved types from `ast.SemanticInfo`, shared
  across unit and program analysis. Remaining name-only callers and explicitly untyped
  execution still use `evaluator/type_resolution.go` and `type_resolution_helpers.go`.
  Migrate those consumers and declaration registration before deleting the inline array /
  function-signature parsers; retain an explicit policy for `WithTypeCheck(false)`.
- **A6** `[~]` L — **Finish typing the registries and consolidating metadata.**
  `ClassInfo`, `ClassValue`, and class metadata mutation now live in `runtime`;
  `ClassRegistry` stores `runtime.IClassInfo`, and the two class factories are gone.
  Remaining: replace `any` registry aliases for record → enum → interface → helper, remove
  their factory seams where applicable, and drop the parallel AST maps on `ClassInfo`
  that duplicate `runtime.ClassMetadata`. Lifetime design input remains
  `docs/archive/phase-4.12.1-lifetime-inventory.md` (owned vs aliased bindings).
- **A7** `[ ]` L — **Typed type identity.** Operator-overload dispatch keys on `"class:"`-prefixed
  strings encoded twice (`internal/interp/runtime/class_operators.go`, `evaluator/runtime_ops.go`)
  and compared with `strings.HasPrefix`, while the typed comparator
  `internal/types/operator_registry.go:100` already exists. Start there; then retire
  `runtime.Value.Type() string` comparisons (92 sites), the `*Name string` metadata fields in
  `runtime/metadata.go`, and `ExecutionContext.recordTypeContext string`. Unlocks
  OperatorOverloadPass/Fail. Depends on A6.
- **A9** `[~]` S — **Finish builtin analysis from registry signatures.** Return-type lookup,
  ordinary calls, and 78 additional builtin names use `builtins.Registry` signatures;
  diagnostic styles preserve historical wording. Remaining: multi-argument math handlers
  with diagnostic-order constraints; date handlers with strict Float rules; Variant handlers
  requiring strict Variant/JSONVariant constraints; reconcile signature discrepancies for
  Trim, RandG, and array-returning JSON/string functions. Retain explicit AST-dependent
  and polymorphic intrinsics.
- **A11** ⏸️ — **Bytecode VM.** Owner decision 2026-07-04: keep in tree, unmaintained, opt-in.
  `run --bytecode`, `dwscript compile`, and
  `pkg/dwscript.CompileModeBytecode` are labeled experimental in help text and godoc. Revisit
  delete-vs-rebuild after A6; a rebuild must use `internal/builtins` and `internal/interp/runtime`
  values, not the current fork. Status: `docs/decisions/bytecode-vm.md`.

**Not planned** (measured, not worth it): source TODO cleanup (27 in total), panic-to-error
conversion (panics are not used for control flow), splitting `visitor_statements.go`/
`visitor_declarations.go` by size alone (they are cohesive dispatch; A6 changes them anyway).

---

## 3. Language compatibility (L)

Each line: what to build → fixtures/category it unlocks. Run
`just fixture-report --category <Cat> --list-fails` for the live list.

### 3.1 Parser

- `[ ]` S `class property` inside record bodies + record auto-property backing fields →
  PropertyExpressionsPass `class_property_expressions`, `class_property_write_expressions`, `property_auto_field`.
- `[ ]` S Multi-index expression-backed indexed properties → PropertyExpressionsPass `indexed_expressions`.
- `[ ]` S Deeply nested closing `>>` in generic type-argument lists → GenericsPass.
- `[ ]` S `external` property syntax → JSONConnectorPass `property_name`.
- `[ ]` S `array of T` as a property type in getter position → JSONConnectorPass `stringify_class_getter`.
- `[ ]` S `{$I %FILE%}`-style value substitution → SimpleScripts.

### 3.2 Semantic

- `[ ]` M Type-order-independent class builder (parents/fields declared later without `forward`);
  needs a real two-phase class registration. Design input: `docs/architecture/semantic-passes.md`
  (superseded design, never implemented).
- `[ ]` S Gate the over-aggressive unused-private-field hint → unmasks JSONConnectorPass
  serialization fixtures in the harness (they already pass in the CLI).
- `[ ]` S Value-context auto-invoke of parameterless function pointers for `and`/`or` operands
  (`Print`/`PrintLn`/`implies` already done).
- `[ ]` S Helper-property resolution through a metaclass → PropertyExpressionsPass `helpers_property_expressions`.
- `[ ]` S Indexed-property read through a metaclass with a class-method accessor → SimpleScripts `enum_to_integer`.
- `[ ]` M Contract inheritance → SimpleScripts `method_contracts`; inline-method class name in
  contract messages → `method_condition`.
- `[ ]` M Generics follow-ups: generic interfaces (`interface1`), generic `external` classes and
  function-pointer types (`class_external1`, `external_promise`, `func_ptr1`), external generic
  method bodies (`function TTest<T>.Foo`), operator-overload specialization, `array of T` method
  gaps (`array1`, `tlist1`) → GenericsPass (14/23).
- `[ ]` M Overload leftovers: class `operator =`/`<>` overloading, `@obj.Method` pointers,
  function-pointer parameter typing in overload sets, metaclass `inherited` → OverloadsPass (33/39).
- `[ ]` S Set leftovers: array ↔ set conversion, `set of` record fields, out-of-range diagnostics
  → SetOfPass (20/25).
- `[ ]` S Lexer-time `Declared()` / `{$FATAL}` → ArrayPass/SetOfPass stragglers.
- ✋ Subrange bounds at compile time: no fixture declares a subrange type; zero yield.

### 3.3 Runtime / evaluator

- `[ ]` M Nested lvalue vivification through a key or index (`a[k].field := v`, `a[k][j] := v`,
  `a[k].Add(…)`) → AssociativePass `elements_of_value`, `array_of_dyn`; JSONConnectorPass
  `generate1`, `basic_generate`.
- `[ ]` S DWScript hash iteration order → AssociativePass `records`.
- `[ ]` S ARC destructor timing on associative slot replace/clear → `delete_sequence`;
  Variant → key coercion → `variant_key_cast`.
- `[ ]` M Record copy-on-assign value semantics → JSONConnectorPass `stringify_record`.
- `[ ]` M JSON node reparent/ownership (`array_add_dupe`, `reparent`, `reposition_node_in_array`),
  float formatting (`int64_json`, `numbers`), variant → scalar cast message parity (`explicit_cast`)
  → JSONConnectorPass (51/82).
- `[ ]` M Raise-site and method-call-site **column** precision → SimpleScripts `stacktrace`,
  `exceptobj3`, `contracts_subproc`. Higher risk: lives in shared position logic validated by many
  position-sensitive fixtures.
- `[ ]` S Function-pointer niche: `@TObject.ClassType` address-of-class-member (`func_ptr_symbol_field`);
  value ↔ parameterless-function coercion in `array of function : T` (`func_ptr_classname`).
- `[ ]` S Re-measure the runtime-panic fixtures (metaclass `ClassName`, class-method dispatch,
  `class of`); the common cases were closed in July, the rest was never re-listed.
- `[ ]` M SimpleScripts to ≥ 85% (331/442 = 75%). Work
  `just fixture-report --category SimpleScripts --list-fails` (identical to the harness list).
- `[ ]` M Triage in-scope categories that have no plan yet: FunctionsTime (1/30),
  FunctionsVariant (0/10), FunctionsGlobalVars (0/16), FunctionsByteBuffer (0/19),
  EncodingLib (0/12), Memory (1/13), InnerClassesPass (0/2), FunctionsDebug (0/3).
  First step for each: list fails, bucket by cause, then add concrete items here.
- ✋ UTF-16 surrogate iteration (`for_in_str`, `for_in_str2`): intentional divergence, see
  [`docs/decisions/string-encoding.md`](docs/decisions/string-encoding.md).

### 3.4 Source TODO backlog (merged from the former `TODOs.md`)

Live `// TODO` markers that are real work, not notes. Bytecode TODOs are omitted (A11).

- `[ ]` `internal/semantic/analyze_arrays.go:162` — validate index expression types against property index-parameter types.
- `[ ]` `internal/semantic/overload_resolution.go:211` — class-hierarchy distance in overload matching.
- `[ ]` `internal/semantic/analyze_records.go:387` — record member visibility rules.
- `[ ]` `internal/semantic/analyze_function_calls.go:26` — use `expectedType` in overload resolution.
- `[ ]` `internal/interp/evaluator/helpers.go:131` — enum range checking.
- `[ ]` `internal/interp/runtime/class.go` — replace compatibility AST method lookups with runtime callables (A6).
- `[ ]` `internal/units/search.go:171-172` — user (`~/.dwscript/lib`) and system library search paths.
- `[ ]` `cmd/dwscript/cmd/fmt.go:296` — real diff algorithm for `dwscript fmt --diff`.
- `[ ]` `pkg/wasm/api.go:126,299` — custom filesystem integration for WASM.
- `[ ]` `pkg/ast/metadata.go:140,153`, `pkg/ast/type_annotation.go:58`, `pkg/ast/base.go:42` — replace `any` symbol slots with a proper `Symbol` type.
- `[ ]` `pkg/dwscript/symbols.go:74` — report the actual scope level instead of `"global"`.
- `[ ]` Skipped tests to revive or delete: `internal/semantic/analyze_types_test.go:130,266`,
  `analyze_functions_test.go:284`, `case_insensitive_test.go:185`; `cmd/dwscript/sets_test.go:15,64` (skip
  message cites a closed P4 item); `internal/parser/functions_decl_test.go:639` (cites old task 5.11).

---

## 4. Error-detection parity (`*Fail` suites, F)

Harness and CLI: 111/647 (FailureScripts 107/528, JSONConnectorFail 2, SetOfFail 1,
AssociativeFail 1, every other `*Fail` suite 0). The suites are **compile-only** (like DWScript's
`CompilationFailure` runner): the expected file is the compiler's message list, hints included,
no envelope, and nothing is executed. Reproduce one with
`dwscript run --diagnostics=plain --compile-only --hints pedantic <file>`.

Work families (from the 2026-03 FailureScripts analysis, now archived at
`docs/archive/failure-scripts-next-phase-plan.md`; counts are approximate and pre-date the July work):

- **F1** `[ ]` M Warning/hint emission and ordering (~54): `Empty THEN block`, unused result,
  unused variable, unreachable code; ordering of warnings vs same-line errors. Case-mismatch
  hints are excluded (✋ §5).
- **F2** `[ ]` M Array diagnostics (~41): `Array expected`, `Too many indices`, bound-exceeded
  wording, malformed array-type recovery.
- **F3** `[ ]` M Parser header/declaration/delimiter recovery (~58): parameter lists, `case`,
  `except`, record/method headers; delimiter wording (`")" expected`, `END expected`,
  `DO expected`, `TO or DOWNTO expected`).
- **F4** `[ ]` L Class/property/static/override/visibility diagnostics (~125): the largest family.
- **F5** `[ ]` M Missing-validation sweep: the 82 fixtures where DWScript reports an error and
  go-dws compiles clean (`conditionals1-6`, `default_params2`, `deprecated`,
  `enum_flags_overflow`, `switch_invalid1-3`, `use_proc_result2`, …).
- **F6** `[ ]` S Runtime-mismatch residue (13): `div_by_zero_float`/`_int`, `dyn_array_setlength3`,
  `for_in_subclass`, `missing_param1`, ….
- **F7** `[ ]` M Per-suite sweeps, all currently 0 and not formatting-only (verified by substring
  check): InterfacesFail 19, HelpersFail 18, OverloadsFail 14, SetOfFail 13,
  PropertyExpressionsFail 10, GenericsFail 8, JSONConnectorFail 7, LambdaFail 6,
  OperatorOverloadFail 6, AssociativeFail 3, AttributesFail 2, InnerClassesFail 1.
- **F8** `[ ]` S Convert the remaining raw `addError(...)` sites to structured diagnostics
  (`analyze_function_calls.go` 54, `analyze_statements.go` 49, `analyze_method_calls.go` 17,
  `analyze_classes.go` 10) so message text and ordering are centrally controlled
  (`docs/archive/semantic-legacy-hotspots-5.3.10.md`).

---

## 5. Deferred, parked, won't-fix

Gate for everything ⏸️ below: **every non-host-library fixture category ≥ 80%** (harness and CLI).

- ⏸️ Compiler backends: Go source / AOT, JavaScript, LLVM, MIR foundation, WebAssembly AOT.
  Backlog preserved in `docs/archive/CodeGenTODO.md`, `docs/archive/CodeGenJSGoal.md`.
- ⏸️ AST-driven formatter. Specs: `docs/decisions/formatter-style-guide.md`,
  `docs/decisions/formatter-ast-audit.md`.
- ⏸️ Host-library bindings (DB / Crypto / COM / Graphics / Web / Tabular / TimeSeries / DOM / Linq,
  ~200 fixtures). Integration surface, not language correctness: `docs/decisions/out-of-scope.md`.
- ⏸️ Sandbox / capabilities model (`AllowFileRead`, `AllowHTTP`, memory/time limits): design only,
  in `docs/decisions/out-of-scope.md`; nothing implemented.
- ⏸️ Bytecode VM: see A11 and `docs/decisions/bytecode-vm.md`.
- ✋ Case-mismatch hint parity (`"println" does not match case of declaration ("PrintLn")`): the
  corpus encodes these hints inconsistently with no signal recoverable from the `.pas` sources
  (the original test runner set hint levels per test). Evidence and measurements:
  `docs/history/progress-log-2026-07.md`, "Hint/warning envelope". Do not pursue unless the
  original per-test configuration is recovered. Non-case hints (empty block, unreachable code,
  prefer-ToString) remain in scope under F1.
- ✋ UTF-16 surrogate iteration: `docs/decisions/string-encoding.md`.
- ✋ Subrange compile-time bounds: zero fixture yield.

---

## 6. Definition of done (v1.0)

1. `just fixture-report` and the Go harness both report **≥ 90%** on every non-host-library
   category, and they agree (one pipeline, one scoring rule — in place since 2026-09-06).
2. All `*Fail` suites reproduce DWScript diagnostics.
3. Exactly **one** type representation (A6, A7), **one** evaluator, and **one** compile pipeline (✅ A1).
4. CI fails on any per-category regression. ✅ Already in place (`baselines.json` gate).
5. No public API or CLI flag exposes a non-functional mode without saying so (A11).

Track progress against fixture pass rate, not checkbox counts.

---

## 7. Reproducing the numbers

```bash
just fixture-report                                      # CLI ground truth (rebuilds bin/dwscript first)
just fixture-report --category SimpleScripts --list-fails
go test ./internal/interp -run TestDWScriptFixtures -v   # Go harness (what CI gates on)
FIXTURE_LIST_FAILS=1 go test ./internal/interp -run TestDWScriptFixtures -v   # ... with failing names
just fixture-update                                      # regenerate TEST_STATUS.md, ratchet baselines
```
