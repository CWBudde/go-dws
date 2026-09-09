# go-dws — Work Plan

> Rewritten 2026-09-06. This file lists **open work only**. Completed work and the reasoning
> behind it live in [`docs/history/progress-log-2026-07.md`](docs/history/progress-log-2026-07.md);
> the measured audits behind the priorities are
> [`docs/history/CODEBASE_REVIEW_2026-07.md`](docs/history/CODEBASE_REVIEW_2026-07.md) (July 2026)
> and [`docs/architecture/audit-2026-09.md`](docs/architecture/audit-2026-09.md) (September 2026).
> Status numbers are generated, not hand-kept. Documentation index: [`docs/README.md`](docs/README.md).

## 0. Status snapshot

**Headline (2026-09-09):** Go harness and freshly rebuilt CLI both
**878 / 1,928 scored = 46%**, with no category regressions after Phase 2.
Both use the shared compile pipeline and scoring rules.
`*Fail` error-detection suites **111 / 647 = 17%**.

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

A5–A7 and A9 closed 2026-09-09: structural type resolution, typed runtime
registries and canonical callables, typed operator/value dispatch, and builtin signature
constraints. A2–A4, A8 and A10 closed 2026-09-07. Completion and validation are recorded in
[`docs/history/progress-log-2026-09.md`](docs/history/progress-log-2026-09.md).

The current architecture is documented in
[`docs/architecture/interp-evaluator-steady-state.md`](docs/architecture/interp-evaluator-steady-state.md).
Only the explicitly deferred bytecode decision remains here.

- **A11** ⏸️ — **Bytecode VM.** Owner decision 2026-07-04: keep in tree, unmaintained, opt-in.
  `run --bytecode`, `dwscript compile`, and
  `pkg/dwscript.CompileModeBytecode` are labeled experimental in help text and godoc. A6 is complete;
  delete-vs-rebuild remains deferred pending an owner decision. A rebuild must use
  `internal/builtins` and `internal/interp/runtime` values, not the current fork. Status: `docs/decisions/bytecode-vm.md`.

**Not planned** (measured, not worth it): source TODO cleanup (27 in total), panic-to-error
conversion (panics are not used for control flow), splitting `visitor_statements.go`/
`visitor_declarations.go` by size alone (they are cohesive dispatch).

---

## 3. Language compatibility (L)

Each line: what to build → fixtures/category it unlocks. Run
`just fixture-report --category <Cat> --list-fails` for the live list.

### 3.1 Parser

No open items. Closed 2026-09-09: `class property` in record bodies with auto-property backing
fields, expression-backed and multi-index indexed properties, `external` property syntax, and
nested `>>` in generic type-argument lists; see
[`docs/history/progress-log-2026-09.md`](docs/history/progress-log-2026-09.md).

- ✋ `{$I %FILE%}`-style value substitution: the only fixture, SimpleScripts `include_expr`, cannot
  pass. Its expected output hardcodes the original Delphi runner's paths (`Test\include_expr.pas`,
  `*MainModule*`), and neither `cmd/fixture-report` `normalize()` nor the Go harness normalizes
  paths; `%FUNCTION%` is also not knowable at lex time. Reopen only with a path-normalizing harness.
- ✋ `array of T` as a property type in getter position: measured 2026-09-09, already works.

### 3.2 Semantic

**Done (2026-09-09):** parameterless function/method-pointer operands for `and` / `or`,
including return-type checking, short-circuit execution, and exception propagation.
Regression tests exercise the shared compile pipeline and production evaluator; completion
and validation are recorded in
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-09--parameterless-callbacks-in-and--or-32).

The remaining work is divided into bounded tasks below. IDs stay stable when neighboring
items close. Fixture names are acceptance targets from the existing backlog; confirm the
current failure before implementing a task, since §3.1 work may already remove a blocker.
Close each task with a passing target fixture or a focused compile-and-run/diagnostic test.

**Coordination:** complete the class-builder tasks in order. Other groups can proceed
independently when their source ownership does not overlap. Coordinate property/record
analysis and shared type changes with §3.1; serialize changes to the generic specializer,
overload resolver, and class metadata within their respective groups. Runtime changes
needed to finish a semantic task belong in the evaluator. Integrate PLAN/history and
fixture-baseline updates after each completed task.

#### 3.2.1 Class construction independent of declaration order

**Closed 2026-09-09.** Class construction is now four explicit phases over the single type
registry in `internal/semantic/class_construction.go` — identity, inheritance, member
signatures before bodies, and ancestor-dependent validation after signatures. No second type
registry was introduced. The old
[semantic-passes design](docs/architecture/semantic-passes.md) remains design input only, not
an implemented pass framework.

**Done (2026-09-09):** L-S1a. Predeclaration is now an explicit two-phase construction in
`internal/semantic/class_construction.go` (identity, then inheritance), still over the single
type registry. Parent links and class-level shape flags are resolved before member/body
checking, cycles and unknown parents stay diagnostics, and forward/partial behavior is
unchanged; see
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-09--two-phase-class-construction-inheritance-before-members-l-s1a).

**Done (2026-09-09):** L-S1b. Inline class method bodies are no longer checked where they are
declared: the signature is registered in source order, the body is queued and checked once the
last top-level class declaration has been analyzed, so every class has its full member surface
(fields, class vars, constants, methods, properties) on its single shared shell first. A body
may now name a class or a member declared later in the file; `SimpleScripts/method_implem` and
`SimpleScripts/var_param_obj_method` newly pass (885 → 887). See
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-09--class-member-signatures-complete-before-body-checking-l-s1b).

**Done (2026-09-09):** L-S1c, closing this section. A fourth phase postpones the
ancestor-dependent validations (`validateVirtualOverride` per method; `checkMethodOverriding`,
`validateInterfaceImplementation` and `validateAbstractClass` as the class tail) into a queue
drained with the deferred bodies, and only when an ancestor's declarations are still
outstanding — so `override` and `inherited` against a parent declared later now work, while
every negative case keeps its existing message and position. Fixtures unchanged at 887 with an
identical failing list. Remaining order dependency, documented rather than fixed: a statement
such as `var c := TC.Create;` written before the abstract ancestor's declaration is still
checked in source order and misses the abstract-instantiation error. See
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-09--ancestor-dependent-class-validation-after-signatures-complete-l-s1c).

#### 3.2.2 Diagnostics and metaclass properties

These are separate fixes; none requires the class-builder refactor as a prerequisite.

**Done (2026-09-09):** L-S2a. The ticket's premise did not reproduce — measured on the shared
pipeline, `JSONConnectorPass/serialize_class` passes and no failing JSONConnectorPass fixture
involves the hint at all. The real defect was a false positive: a private field named by bare
name inside a method body or an expression-form property accessor resolved through the symbol
table and was never marked used. Class field bindings now carry their declaring class
(`Symbol.ClassFieldOwner`), the six fixtures that emitted a bogus hint emit none, and fixtures
rose 888 → 892. See
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-09--unused-private-field-hint-usage-tracking-l-s2a).

- ✋ Unused-private-field hints when the program also has a compile error: the blanket
  suppression in `internal/semantic/unused_warnings.go` drops every private-member hint for a
  class as soon as any non-hint diagnostic exists, which contradicts
  `OverloadsFail/overloads_not_implem`. That fixture fails for unrelated parser reasons, so the
  rule is untestable today. Reopen once the fixture parses.
- ✋ Unused-private hints for record fields and class vars: `types.RecordType` has
  `FieldVisibility` but no usage-tracking infrastructure, and class vars have none either.
  Measured 2026-09-09; no fixture demands it.
**Done (2026-09-09):** L-S2b. Helper properties now support expression-form accessors and are
reachable through a metaclass, a type cast's static class, and a record receiver, on both the
read and the write side; a non-identifier write specifier is recognized as the lvalue shorthand
it is, and record class vars written through an instance reach shared storage. All three
helper-property fixtures pass (`helpers_property_expressions`,
`class_helpers_property_write_expressions`, `record_helpers_property_write_expressions`),
PropertyExpressionsPass 15 → 18. See
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-09--helper-property-expression-accessors-and-metaclass-resolution-l-s2b).

- ✋ `read_write_other_property`: a property whose read/write specifier names *another property*
  (`property Mapped : Integer read Prop write Prop`) is rejected at compile time. Measured
  2026-09-09; a distinct gap from helper properties, belongs with §3.1 property handling.
- ✋ Record-type metaclass member access (`TRec.SomeClassProperty` through the type name, as
  opposed to through an instance) is unsupported. Measured 2026-09-09; no fixture demands it.
- **L-S2c** `[ ]` S — Resolve indexed reads through a metaclass when the accessor is a class
  method → SimpleScripts `enum_to_integer`. Preserve the accessor's class receiver and
  validate index arguments; coordinate shared index validation with §3.4.

#### 3.2.3 Contracts

- **L-S3a** `[ ]` S — Inherit `require` conditions on inherited and overridden methods.
  Acceptance: focused compile-and-run cases from SimpleScripts `method_contracts`, including
  dispatch through a base-typed reference and the upstream combination of base/derived conditions.
- **L-S3b** `[ ]` S — Inherit and combine `ensure` conditions, preserving the declaring method's
  context and any `old(...)` capture. Coordinate contract metadata with L-S3a; acceptance:
  focused postcondition cases and the complete `method_contracts` fixture after both land.
- **L-S3c** `[ ]` S — Include the declaring class name in inline-method contract messages →
  SimpleScripts `method_condition`. Independent of inheritance; keep shared source-column
  precision work in §3.3 separate.

#### 3.2.4 Generics

Keep specialization changes focused on one capability at a time. Coordinate syntax support
with §3.1, especially nested type-argument delimiters and out-of-line generic method headers.

- **L-S4a** `[ ]` S — Specialize generic interfaces and their method signatures →
  GenericsPass `interface1`.
- **L-S4b** `[ ]` M — Specialize generic `external` class declarations and member signatures →
  `class_external1`, `external_promise`. Distinguish compile support from any host runtime
  binding requirement before claiming a fixture is closed.
- **L-S4c** `[ ]` S — Specialize generic function-pointer parameter and return types → `func_ptr1`.
- **L-S4d** `[ ]` M — Bind out-of-line generic method bodies such as
  `function TTest<T>.Foo` to their declaration and specialized type parameters. Acceptance:
  a compile-and-run regression with multiple concrete specializations; parser support must
  be present first.
- **L-S4e** `[ ]` S — Resolve operators against concrete specialization types →
  `specialize_to_operator_overload`; cover both built-in and user-defined operators.
- **L-S4f** `[ ]` M — Complete `array of T` substitution and method-call typing → `array1`,
  `tlist1`. Depends on L-S4d where methods are defined out of line; preserve array bounds
  and parameter modes through specialization.

#### 3.2.5 Overloads and method pointers

- **L-S5a** `[ ]` S — Resolve and dispatch class `operator =` / `<>` overloads →
  OverloadsPass `class_equal_diff`.
- **L-S5b** `[ ]` S — Type and bind `@obj.Method` in overload arguments without invoking it →
  `class_vs_proc`. Keep address-of-class-member runtime work in §3.3 separate.
- **L-S5c** `[ ]` S — Use expected function-pointer signatures to select overload candidates →
  `overload_func_ptr_param`; retain ambiguity and incompatible-signature diagnostics.
  Coordinate with §3.4's expected-type overload-resolution item.
- **L-S5d** `[ ]` S — Resolve metaclass `inherited` calls with the correct overload and class
  receiver → `overload_on_metaclass`. Coordinate class dispatch with L-S2c when files overlap.

#### 3.2.6 Sets

- **L-S6a** `[ ]` S — Complete array ↔ set conversions, including empty inputs and element
  compatibility → SetOfPass `array_to_set`, `init_from_array`, `init_from_empty_array`.
  Add a focused reverse-conversion test if the corpus does not exercise that direction.
- **L-S6b** `[ ]` S — Resolve `set of` record fields consistently through analysis and runtime
  metadata → `set_in_record`. Coordinate with §3.1 record changes.
- **L-S6c** `[ ]` S — Finish set range validation and out-of-range diagnostics → relevant
  SetOfFail cases, with `in_set_out_of_range` protecting membership behavior. Coordinate
  evaluator enum-range checks with §3.4; do not add unrequested subrange-type support.

#### 3.2.7 Conditional compilation (lexer/semantic coordination)

The old ArrayPass/SetOfPass attribution needs re-identification: no `Declared(` or
`{$FATAL` occurrence was found in those current `.pas` sources during this refinement.
Keep these tasks scoped to reproduced cases, including their include files.

- **L-S7a** `[ ]` S — Identify the failing conditional-compilation case, then expose the
  appropriate declaration visibility to lexer-time `Declared()`. Acceptance: compile tests
  for present/absent identifiers, case-insensitivity, and inactive branches. Coordinate the
  lexer/preprocessor boundary with §3.1 value substitution.
- **L-S7b** `[ ]` S — Identify the failing `{$FATAL}` case, then implement active-branch fatal
  diagnostics with source position and message parity. Depends on L-S7a only when the branch
  condition uses `Declared()`; verify that inactive branches emit no fatal diagnostic.

✋ Subrange bounds at compile time: no fixture declares a subrange type; zero yield.

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
