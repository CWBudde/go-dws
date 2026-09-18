# go-dws — Work Plan

> Rewritten 2026-09-06. This file focuses on **open work**, with completed checkpoints retained
> in §3.5. Detailed completed work and the reasoning
> behind it live in [`docs/history/progress-log-2026-07.md`](docs/history/progress-log-2026-07.md);
> the measured audits behind the priorities are
> [`docs/history/CODEBASE_REVIEW_2026-07.md`](docs/history/CODEBASE_REVIEW_2026-07.md) (July 2026)
> and [`docs/architecture/audit-2026-09.md`](docs/architecture/audit-2026-09.md) (September 2026).
> Status numbers are generated, not hand-kept. Documentation index: [`docs/README.md`](docs/README.md).

## 0. Status snapshot

**Headline (2026-09-18):** Go harness and freshly rebuilt CLI agree at **1,147 / 1,966 scored =
58%**; `*Fail` error-detection suites **166 / 641 = 26%**. What shipped to get there is in
[the September progress log](docs/history/progress-log-2026-09.md), not here.

**What the denominator is.** 2,044 fixtures ship in the tree; 78 have no applicable expectation
and remain unscored, leaving 1,966. This includes 36 missing-`.txt` fixtures now checked against
silence (T7). The denominator still contains the **219 host-library fixtures excluded from every
target below** (see the scope rule further down) — all 219 currently fail. Excluding them, the
same run reads **1,147 / 1,747 = 66% in scope**, the number to track against §6. Both are honest;
the lower one is quoted outward. The [fixture README](testdata/fixtures/README.md) documents
which missing expectations are scored and which remain excluded.

Open, in leverage order:

- **§4** is where the remaining mass is. Re-measured fixture-by-fixture on 2026-09-12 and
  rewritten around what that found: F6 closed outright, F5 grew by 27 fixtures in suites it had
  never been counted over, F8 turned out to block 265 of the 480 remaining failures, and two new
  families were split out — the 68 fixtures one line from passing (F9) and the `Incompatible
  types` sentence (F10). Full tables:
  [`docs/architecture/fail-suite-audit-2026-09.md`](docs/architecture/fail-suite-audit-2026-09.md).
- **§3.5** is the other half of the measurement: 125 in-scope fixtures now fail in the
  suites that *run* a program, and until 2026-09-12 no item covered any of them. The two cheap
  ones (E1, E2) shipped the same day; E3, the case-mismatch hint, is structural and cross-cutting,
  and E8 is a by-reference binding bug E1 turned up.
- **§3.3** has only deferred Memory host setup left (7 of 13 scored). Private unit variables,
  Memory scoring, and constructor assignment receivers closed 2026-09-13; see the progress log.
- **§1** is closed: T7 scores 36 previously skipped fixtures and T8 classifies failures in
  `fixture-report`.
- **§3.4** has one item, blocked on the evaluator. **§2** has one, deferred by owner decision.

Where the truth lives:

| Source | What it is | How to refresh |
| --- | --- | --- |
| `testdata/fixtures/TEST_STATUS.md` | generated per-category pass/fail table (harness) | `just fixture-update` |
| `testdata/fixtures/baselines.json` | per-category pass-count floor that CI enforces | `just fixture-update` after an intentional improvement |
| `just fixture-report` | end-to-end CLI numbers (`cmd/fixture-report` running `bin/dwscript run`) | `just build` first, or the numbers are stale |

Rules for this document:

- An item is closed only by a **passing fixture** (or a test that exercises the real user-facing path).
  Closed items are deleted from this file, except for §3.5's requested completion checklist;
  their story goes to `docs/history/progress-log-<date>.md`.
- Ratchet `baselines.json` after every improvement (`just fixture-update`).
- The ~200 fixtures in host-library categories listed in
  [`docs/decisions/out-of-scope.md`](docs/decisions/out-of-scope.md) (DataBaseLib, COMConnector,
  CryptoLib, GraphicsLib, WebLib, TabularLib, TimeSeriesLib, DOMParser, Linq, LinqJSON, ClassesLib,
  DelegateLib, SystemInfoLib, IniFileLib, FunctionsFile, FunctionsRTTI, BigInteger,
  FunctionsMathComplex/3D) are excluded from every target below.
- Where the remaining failures are (819 total, 2026-09-18): **219 host-library** (out of scope),
  **475 in the `*Fail` error-detection suites** (§4: FailureScripts 373, InterfacesFail and
  HelpersFail 18 each, the rest under 15), and **125 in the execution suites** (§3.5:
  SimpleScripts 67, ArrayPass 17, JSONConnectorPass 9, InterfacesPass 6, FunctionsMath 5, a tail
  of ones and twos). A pre-existing runtime stack overflow still appears in an isolated harness
  worker; fixture scoring completes and the category baseline gate passes.
- Regenerate that split rather than trusting it: `just fixture-report --in-scope --classify`
  (T8, closed 2026-09-12). It reports each failure's distance from passing, whether what differs is
  a diagnostic or the program's output, and which message shapes recur — none of which
  `baselines.json` can see, because it holds pass-count floors.

Legend: `[ ]` open · `[~]` partially done, remainder listed · `[x]` done · ⏸️ gated, do not start ·
✋ won't-fix, with the decision record. Size: S (hours), M (days), L (week+).

---

## 1. Measurement & tooling (T)

T1–T6 closed 2026-09-06 (one compile pipeline for CLI and harness, `run
--diagnostics=plain|pretty`, `--test-envelope`, `--compile-only`, self-rebuilding
`fixture-report` with a stale-binary guard, helper-spec parity test, unscored-variant docs);
see [`docs/history/progress-log-2026-09.md`](docs/history/progress-log-2026-09.md). New tooling
items go here.

T7 closed 2026-09-13: shared missing-expectation scoring and category hint levels; see
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-13--missing-fixture-expectations-and-memory-hints-t7).

**Closed here (2026-09-12):**

- [The fixture classifier](docs/history/progress-log-2026-09.md#2026-09-12--the-fixture-classifier-t8)
  (**T8**) — `fixture-report --classify` buckets every failure by distance, by what kind of line
  differs and by message shape, and `--in-scope` drops the host-library categories. It replaced a
  throwaway shell script, and rebuilding the measurement changed two things: the `*Fail` near-miss
  counts (see §4) and the discovery that §3 had no item for the failing execution-suite fixtures
  (now §3.5).

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

## 3. Language compatibility (L, E)

Each line: what to build → fixtures/category it unlocks. Run
`just fixture-report --category <Cat> --list-fails` for the live list, or
`--classify` with it to see how far each failure is from passing.

§3.1–§3.4 are organised by *subsystem* and are largely closed. §3.5 is organised by *failing
suite* and is not: it holds the execution-suite failures the subsystem view never surfaced.

### 3.1 Parser

No open items. Closed 2026-09-09: `class property` in record bodies with auto-property backing
fields, expression-backed and multi-index indexed properties, `external` property syntax, and
nested `>>` in generic type-argument lists; see
[`docs/history/progress-log-2026-09.md`](docs/history/progress-log-2026-09.md).

Also closed 2026-09-12:
[the keyword operators are case-insensitive](docs/history/progress-log-2026-09.md#2026-09-12--keyword-operators-are-case-insensitive-31)
— the spelling is folded once where the node is built, covering `and or xor not div mod shl shr sar
in implies` (+3 fixtures).

- ✋ `{$I %FILE%}`-style value substitution: the only fixture, SimpleScripts `include_expr`, cannot
  pass. Its expected output hardcodes the original Delphi runner's paths (`Test\include_expr.pas`,
  `*MainModule*`), and neither `cmd/fixture-report` `normalize()` nor the Go harness normalizes
  paths; `%FUNCTION%` is also not knowable at lex time. Reopen only with a path-normalizing harness.
- ✋ `array of T` as a property type in getter position: measured 2026-09-09, already works.

### 3.2 Semantic

**Closed 2026-09-09 / 09-10.** All seven bounded task groups shipped (L-S1…L-S7); write-ups in
[the September progress log](docs/history/progress-log-2026-09.md):

| Group | What shipped | Fixtures |
| --- | --- | --- |
| Parameterless callbacks in `and` / `or` | function/method-pointer operands, with return-type checking, short-circuiting and exception propagation | — |
| 3.2.1 Class construction (L-S1a–c) | four explicit phases over the single type registry: identity, inheritance, member signatures before bodies, ancestor-dependent validation after | 885 → 887 |
| 3.2.2 Diagnostics and metaclass properties (L-S2a–c) | private-field usage tracking, helper-property expression accessors, indexed properties with class-method accessors | 888 → 896 |
| 3.2.3 Contracts (L-S3a–c) | contracts resolved through the ancestor chain, each condition reported under the class that *declares* it | SimpleScripts 336 → 338 |
| 3.2.4 Generics (L-S4a–f) | generic interfaces and array aliases as templates, generic instantiations in inheritance lists, out-of-line generic method bodies | GenericsPass 15 → 23 (100%) |
| 3.2.5 Overloads and method pointers (L-S5a–d) | — | OverloadsPass 33 → 37 of 39 |
| 3.2.6 Sets (L-S6a–c) | bracket literals convert on their expected type, set literals fold as constants, four new set diagnostics | SetOfPass 21 → 25 (100%), SetOfFail 1 → 5 |
| 3.2.7 Conditional compilation (L-S7a–b) | `Declared()` as a real compile-time predicate in both the preprocessor and expression positions; message directives with message/position parity | 920 → 937 |

One order dependency remains, documented rather than fixed (L-S1c): `var c := TC.Create;` written
*before* the abstract ancestor's declaration is still checked in source order and misses the
abstract-instantiation error. The old
[semantic-passes design](docs/architecture/semantic-passes.md) stays design input only, not an
implemented pass framework.

**The one concrete follow-up from this section** is the last bullet of 3.2.7's notes:
`THelper.Proc(TObject.Create)`, a helper method called with an explicit instance argument, is an
unimplemented call form.

**Measured follow-ups and deliberate exclusions.** Each of these was looked at and left; the
reason stays explicit. Items reopened by later evidence point to their active work group.

- ✋ Unused-private-field hints when the program also has a compile error: the blanket
  suppression in `internal/semantic/unused_warnings.go` drops every private-member hint for a
  class as soon as any non-hint diagnostic exists, which contradicts
  `OverloadsFail/overloads_not_implem`. That fixture fails for unrelated parser reasons, so the
  rule is untestable today. Reopen once the fixture parses.
- ✋ Unused-private hints for record fields and class vars: `types.RecordType` has
  `FieldVisibility` but no usage-tracking infrastructure, and class vars have none either.
  Measured 2026-09-09; no fixture demands it.
- ✋ `read_write_other_property`: a property whose read/write specifier names *another property*
  (`property Mapped : Integer read Prop write Prop`) is rejected at compile time. Measured
  2026-09-09; a distinct gap from helper properties, belongs with §3.1 property handling.
- ✋ Record-type metaclass member access (`TRec.SomeClassProperty` through the type name, as
  opposed to through an instance) is unsupported. Measured 2026-09-09; no fixture demands it.
- ✋ `Preconditions must be defined in the root method only`: upstream rejects `require` on a
  non-root method. Not implemented — it belongs to §4 error-detection parity, and
  `FailureScripts/contracts_precondition` also needs `Warning: Constant condition`. The runtime
  meanwhile evaluates every `require` in the chain, root-most first.
- ✋ Class invariants parse into `ClassDecl.Invariants` but are never evaluated. No fixture
  demands them; measured 2026-09-09.
- ✋ `GenericsFail` (8 fixtures, 0 passing) is untouched: it belongs to §4 error-detection
  parity. `GenericsFail/implem_mismatch1` now gets further before failing — DWScript's
  "T expected but u found" check for a mismatched out-of-line type-parameter name is not
  implemented; substitution is positional instead.
- ✋ Type-parameter constraints (`<T: TObject>`) are parsed and ignored; no fixture in
  `GenericsPass` demands them. Measured 2026-09-09.
- ✋ Comparing a function pointer against `nil` (`f = nil`) still reports
  "operator = requires comparable types". Assignment and argument passing work; no fixture
  demands the comparison. Measured 2026-09-09.
- The 2026-09-18 E3d audit confirms case-hint parity across all 39 `OverloadsPass` fixtures,
  including `overload_ambiguous_delegate` and `overload_class_method`. The 14 `OverloadsFail`
  fixtures need `The function X was forward declared but not implemented`, which does not exist
  anywhere in the tree — that is §4 / F7, not this section.
- ✋ The other nine `SetOfFail` fixtures (`bracket_left_missing`, `bracket_right_missing`,
  `for_in_set_missing_do`, `include`, `invalid_method`, `invalid_operand`, `of_missing`,
  `test_non_variable`, `type_missing`) are parser-recovery and message-parity work, not set
  semantics — they belong to §4 / F7.
- ✋ `FailureScripts/static_methods` regressed and is the one fixture lost: it passed only
  because `{$FATAL}` was ignored, and upstream's expectation omits the `Compile Error` line
  even though the directive is present. The only structural difference from the three
  fixtures where the fatal *is* reported is that its `{$FATAL}` is not at column 1 — a
  correlation that holds 5/5 but has no plausible tokenizer mechanism, so it was not encoded.
  Reopen if the reference submodule is ever checked out.
- ✋ `ConditionalDefined(s)` always folds to `False`; `{$DEFINE}` symbols live in preprocessor
  state the analyzer cannot reach. Argument validation is complete.
- `[ ]` `HelpersPass/declared_helper` resolves all four `Declared()` calls; the 2026-09-18 E3d
  audit confirms its case hints match. `THelper.Proc(TObject.Create)` — a helper method called
  with an explicit instance argument — remains an unimplemented call form, the follow-up named
  above.
- ✋ `special_funcs4` (needs `Expression expected` for `Inc(i, )`) and `conditionals2.1` (wants
  the unbalanced report at the directive argument, column 9, where the byte-identical
  `conditionals2` wants it at the name, column 3) stay with §4 / F7.
- ✋ Subrange bounds at compile time: no fixture declares a subrange type; zero yield.
- ✋ `for <var> in <set>` does not type-check the loop variable against the set's element type:
  `var i: Integer; for i in s do` over a `set of TEnum` is accepted silently. Found 2026-09-11
  while clearing the §3.4 skipped-test backlog; the commented-out
  `TestLargeSetForInLoopVariableTypeError` that documented it was deleted. No fixture demands it.

### 3.3 Runtime / evaluator

Every other category this section used to list is closed: FunctionsByteBuffer 19/19 (see
[`docs/guide/bytebuffer.md`](docs/guide/bytebuffer.md)), FunctionsTime 30/30, FunctionsVariant 10/10,
FunctionsDebug 3/3, InnerClassesPass 2/2, EncodingLib 12/12.

- ⏸️ **Memory — 7 of 13 scored (2026-09-13).** The constructor assignment receiver fix closed
  `obj_fields` and SimpleScripts `override_deep`; see
  [the progress log](docs/history/progress-log-2026-09.md#2026-09-13--constructor-results-as-assignment-receivers-33).
  - ⏸️ The six `external*` fixtures need a **host-registered external class** — upstream's `SetUp`
    registers `TExposedClass` and `TExposedBoomClass` with host-side constructors and an
    `OnCleanUp` hook (`UMemoryTests.pas:86-118`). That is host-integration surface, §5 territory.
    go-dws answers `parent class 'TExposedClass' not found`, which is right for a host that
    registered nothing.
  - ✋ Upstream's leak assertions (`exec.ObjectCount = 0`, external-object count) do not port: they
    check DWScript's reference counting at a point where Go's GC has not necessarily run. Five of
    the formerly unscored fixtures exist only to make that assertion; "compiles and prints nothing" is
    all of it that is portable.
- ✋ FunctionsGlobalVars `queue_snapshot`: measured 2026-09-12, the produced output already
  matches the expectation exactly, line for line. The only difference is four
  `"join" does not match case of declaration ("Join")` hints. The declaration-resolution
  discriminator remains unproven: our analyzer types `Map`'s result from the callback's return type, so the
  receiver here is `array of String` — the same element type as `ArrayPass/dynamic_array_remove`,
  where upstream *does* emit the hint. Keep this case parked pending evidence of upstream's
  resolution behavior; do not regress `Map`'s return-type inference to force a hint match.
  E3a's recovered runner settings do not resolve this discrepancy, and it is not evidence of
  missing per-test hint configuration or a GlobalVars runtime gap.
- ✋ UTF-16 surrogate iteration (`for_in_str`, `for_in_str2`): intentional divergence, see
  [`docs/decisions/string-encoding.md`](docs/decisions/string-encoding.md).

**Closed 2026-09-12:**
[the runtime-panic re-measurement](docs/history/progress-log-2026-09.md#2026-09-12--the-runtime-panic-re-measurement-33) — there
were no runtime panics; the three areas the item named were failing for ordinary reasons
(nil-metaclass message parity, class aliases as class names, non-virtual/`reintroduce` dispatch),
all now fixed. `SimpleScripts` 349 → 354. Note it covered *runtime* panics only; the compile-time
crashes were found separately and closed under §4.

### 3.4 Source TODO backlog (merged from the former `TODOs.md`)

Live `// TODO` markers that are real work, not notes. Bytecode TODOs are omitted (A11).
One item remains, and it is blocked on the evaluator rather than ready to build.

- `[ ]` Expected-type overload resolution (was `analyze_function_calls.go:26`): the dead `expectedType`
  parameter is gone; the note now sits at the dispatch in `internal/semantic/analyze_expressions.go`.
  Return type is not part of overload identity (`types.SignaturesEqual`), so an expected type can only
  be a last-resort tie-break. Resolving one in the analyzer alone admits programs the AST evaluator
  then runs with a different overload — it resolves independently at run time
  (`evaluator.ResolveOverloadMultiple`) with no expected-type channel. Blocked until the evaluator can
  see the call site's expected type, or reuse the analyzer's choice via `ast.SemanticInfo`.

**Closed here (2026-09-11 / 09-12):**

- [The `platform.Platform` engine seam](docs/history/progress-log-2026-09.md#2026-09-12--the-platformplatform-engine-seam) —
  `WithPlatform`, `Engine.Platform()`/`FS()`, `builtins.Context.FS()`, and the first two file
  built-ins routed through it. Fixtures unchanged, as expected: no scored fixture calls them, and
  FunctionsFile needs a `File` handle type and the path helpers, which this seam does not provide.
- [Two stale items, measured and closed](docs/history/progress-log-2026-09.md#2026-09-12--two-stale-34-items-measured-and-closed)
  — class-hierarchy distance in overload matching already worked
  (`types.SignatureDistance`/`classDistance`), and the three `t.Skip`ped class-operator
  inheritance tests documented a bug that does not exist: the third failed only because a
  constructor parameter `id` shadows the field `ID`, and DWScript is case-insensitive, so
  `ID := id` is a self-assignment. Nothing was implemented; both were stale bookkeeping.
- [The skipped-test backlog](docs/history/progress-log-2026-09.md#2026-09-11--the-skipped-test-backlog-34) — every entry revived,
  deleted or turned into a real check; const static-array element assignment is now diagnosed
  (`FailureScripts` 125 → 126).

### 3.5 Execution-suite failures (E)

**New 2026-09-12**, from the first classification run over the suites that *run* a program rather
than compile it. §3.1–§3.4 report "no open items" while **145 in-scope execution-suite fixtures
fail**; they were never enumerated because the only measurement that existed covered the `*Fail`
suites. Tables and method:
[`docs/architecture/pass-suite-audit-2026-09.md`](docs/architecture/pass-suite-audit-2026-09.md).
Regenerate with `just fixture-report --in-scope --classify --list-fails`.
The September 18 measurement has 125 execution failures. E5 closes six, taking InterfacesPass
from 21/33 to 27/33; E6 closes four, taking JSONConnectorPass from 69/82 to 73/82. Earlier
improvements are recorded in the progress log; the classification counts below describe the
September 12 audit.

Read the numbers with one caveat: **111 of the 145 are classified `mixed`** (a diagnostic *and* the
output differ), which for an execution suite is almost always one fault — a spurious compile error
stops the program, so its output goes missing too. Fix the error and both lines go away. It also
means distance overstates these: a one-line spurious error on a program printing forty lines
scores 41.

Each subtask below is separately trackable. Write a failing test through the real compile/run
path before implementing a fix. Close implementation subtasks with passing fixtures or real-path
regressions; refresh category baselines and CLI/harness parity when fixtures improve.

#### E3 — Case-mismatch hints `[~]` M

The September 18 audit supersedes September 12's 28 fixtures missing hints and 9 receiving
unwanted hints. Across the same five categories, two fixtures still lack hints and three have
extra hints; no duplicate case diagnostics were found. Only `HelpersPass/record_array_helper`
is an isolated case-hint residual; the others need scope, API or parser-recovery fixes first.

- [x] **E3a — Establish hint configuration.** Completed 2026-09-13: the bundled
  `UScriptTests.pas` sets pedantic hints for all five E3 categories, with no per-fixture override.
  Fresh CLI runs match `array_in2` at pedantic and `hanoi` at normal, resolving July's gate example.
  The source-suppression and declaration-resolution findings in this historical audit are
  superseded by E3b/E3d below; independent diagnostic ordering and overload hints remain separate.
  Settings and the original exact diffs are recorded in
  [the E3a audit](docs/history/progress-log-2026-09.md#2026-09-13--fixture-hint-configuration-e3a).
- [x] **E3b — Apply source hint controls consistently.** Existing shared-pipeline regressions
  verify suppression, re-enabling, configured defaults, includes, units and invalid switches.
- [~] **E3c — Unify declared-name handling.** Declaration-spelling and duplicate-emission
  regressions are in place. Remaining isolated work: establish the upstream resolution rule for
  the two unwanted `X`/`x` hints in `HelpersPass/record_array_helper` at 23:21 and 23:35;
  do not suppress record-field hints broadly or weaken case-insensitive lookup.
  Recheck `SimpleScripts/class_var_dyn2` after class-variable scope resolution (§3.2),
  `string_builtin_methods` after missing string APIs (§3.2 helper support / E7), and
  `ArrayPass/dynamic_anonymous_record` plus `FailureScripts/block_unfinished4` after parser
  recovery (§4/F3). These prerequisites must not be hidden with hint suppression.
- [x] **E3d — Re-measure the affected fixtures.** Compared case-diagnostic multisets across
  SimpleScripts (443), ArrayPass (115), HelpersPass (27), OverloadsPass (39) and FailureScripts
  (542). `classname_helper1` timed out during execution; compile-only output has no case hints.
  The [September 18 audit](docs/history/progress-log-2026-09.md#2026-09-18--case-hint-controls-and-residual-audit-e3)
  records missing/extra hints and prerequisite defects; §5's unverified-runner gate is unchanged.

#### E5 — Interface casting and comparison `[x]` — completed 2026-09-18

Six fixtures close: `interface_cast_to_obj`, `interface_multiple_cast`,
`interface_nil_cast_from_intf`, `intf_casts`, `intf_compare`, and `interface_implements_intf`.
`InterfacesPass` rises **21 → 27/33**, with no category baseline decreases.
Implementation and validation are recorded in
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-18--interface-casts-comparisons-and-implementation-checks-e5).

- [x] **E5a — Accept valid interface casts.** Class/interface casts validate the actual object
  at runtime, with nil, aliases, identity, exceptions and root `IInterface` conversion covered.
  Invalid implicit assignments and incomplete class implementations remain compile errors.
- [x] **E5b — Fix interface comparisons.** Unrelated interface types compare underlying object
  identity with `=`/`<>`; nil operands work, ordering and scalar comparisons remain invalid.
- [x] **E5c — Verify implementation checks and remaining failures.** Class references and
  metaclass variables support `implements`; explicit interface declarations remain its rule.
  Cast failures match fixture messages and positions; the remaining failures are listed below.

Bounded follow-ups identified during E5 triage; do not broaden casting/comparison fixes to
silence these failures:

- [ ] `interface_nil_cast_from_obj`: `Impl` is tokenized as the reserved `IMPL` token;
  resolve identifier handling separately and retain a nil-object cast regression using `Obj`.
- [ ] `interface_properties`: indexed/default interface property semantics.
- [ ] `intf_delegate`: interface method references assigned to procedure variables.
- [ ] `intf_in_record`: restore the caller location in the runtime-error trace.
- [ ] `intf_private`: unwanted unused-private hints for `Hello` and `Unused` (§4/F1).
- [ ] `intf_self_ref`: resolve self-referential interface method return types.
- [ ] Mixed class/interface equality: semantic analysis accepts operands that the evaluator
  rejects; establish compatible reference-comparison behavior separately from interface/interface
  equality.
- [ ] Interface alias assignment: assigning an object directly to an interface alias is
  rejected, while assigning through a variable of the underlying interface type works.

#### E6 — JSON value conversion `[x]` — completed 2026-09-18

The fresh September 18 pre-change classifier finds **69/82 passing, 13 failing** in
`JSONConnectorPass`; the earlier count of 14 was stale. Two causes account for four fixtures:
global-variable string storage and missing associative-array serialization. Both fixes pass,
taking the category to **73/82**, with no category regressions. The generated baseline and
status report match the rebuilt CLI across all 61 categories.

- [x] **E6a — Extract JSON strings at the global-storage boundary.** `global_var` stores a
  JSON string as its decoded text, including empty and escaped strings. JSON arrays/objects
  keep their serialized-text storage behavior; explicit JSON serialization retains quotes.
  Real-path regressions cover globals, queues and compare-exchange.
- [x] **E6b — Serialize associative arrays without losing their contents.** The missing map
  serializer caused `implicit_to_int2` to print `null`; numeric assignment already worked.
  Maps now serialize as JSON objects in bucket order, recursively preserving numeric/null
  values and nested content. JSON source nodes remain unchanged and getter exceptions stop
  subsequent entries. `assign_static_to_dynamic`, `associative_array` and `implicit_to_int2` pass.
- [x] **E6c — Classify the remaining failures.** The fresh CLI classifier separates the seven
  groups below. The [September progress log](docs/history/progress-log-2026-09.md#2026-09-18--json-global-storage-and-associative-serialization-e6)
  records the first blockers, implementation and final validation.

Bounded follow-ups from E6 triage, separate from global storage and map serialization:

- [ ] **JSON associative-key conversion** (`implicit_associative_key_cast`): normalize JSON
  numeric keys to the declared String key type so a lookup using `'123'` finds a JSONVariant
  key holding 123.
- [ ] **JSONVariant parameter conversion** (`implicit_from_cast`): materialize the declared
  JSONVariant representation for Null, unassigned Variant and primitive arguments before
  method dispatch; `Test(Null)` currently attempts `TypeName` on raw NULL.
- [ ] **Inline record array types** (`const_array`, `stringify_array_of_array`): resolve anonymous
  record element types in static and dynamic arrays; compilation currently stops before JSON
  serialization runs.
- [ ] **Coercive comparison and membership** (`comparison2`, `in_static`): reconcile numeric
  JSON/string comparisons and JSON/native or Variant/native membership, retaining lexical
  distinctions for string/string comparisons.
- [ ] **Invalid JSON array deletion** (`delete_array_index`): raise a catchable exception for
  an invalid index and leave the array unchanged; the failed deletion is currently ignored.
- [ ] **JSON circular references** (`circular_references`): reject self and transitive cycles
  before changing ownership in object assignment and array insertion.
- [ ] **Immediate JSONVariant assignment** (`write_immediate_prop`): route member/index writes
  through JSON mutation handling so primitive-backed JSONVariants produce the expected
  catchable `Immediate`/`String` diagnostics.

Additional serialization limitation found during review, outside the nine fixtures above:

- [ ] **Duplicate textual Variant keys:** distinct associative keys such as Integer `1` and
  String `'1'` coexist, but conversion through a JSON object collapses their identical member
  names. Upstream `StringifyAssociativeArray` writes each occupied bucket directly, retaining
  both names. Add a real-path regression and preserve duplicate names when emitting JSON text;
  establish `JSON.Serialize` behavior separately before changing the JSON value representation.

#### E7 — Two spurious compile errors `[ ]` S

- [ ] **E7a — Implement `String.Replace` helper calls.** Use the shared helper machinery and
  cover replacement results and calls whose result is ignored. Close with
  `SimpleScripts/ignore_result` passing.
- [ ] **E7b — Accept Variant conditions in `Assert`.** Align semantic acceptance with runtime
  Variant conversion; check true, false and invalid conditions. Close with
  `SimpleScripts/assert_variant` passing.

#### E8 — Array-element var-argument binding `[ ]` S

`prepareArrayElementReference` evaluates member-based indices such as `a.High`/`a.Length` as
NIL in the affected path, so `P(a[a.High+1])` falls back to a copy. Simple variable and arithmetic
indices bind correctly. The routing also causes the wrong bounds-error position; see the
`IndexArray` comment in `internal/interp/evaluator/index_ops.go`.

- [ ] **E8a — Preserve the array-element reference.** Reproduce a lost write through a var
  parameter with a member-based index, then correct index evaluation and binding. Keep `a[i]`
  and `a[2+2]` as working controls; verify the array and index are each evaluated once.
- [ ] **E8b — Stop binding on evaluation failure.** Cover exceptions and out-of-bounds indices;
  propagate the failure without silently switching to by-value argument passing or calling the
  target routine.
- [ ] **E8c — Correct bounds positions after E8a/E8b.** Distinguish read and var-binding anchors.
  Require `SimpleScripts/const_array_empty` and `ArrayPass/array_element_byref` to pass together;
  changing the read anchor alone previously traded one passing fixture for the other.

#### E9 — Single evaluation of assignment and var-argument receivers `[~]`

This ID covers the previously unnumbered receiver follow-up from §3.3. Construction counters
and a class var retaining `Self` expose duplicate execution and incorrect receiver identity.

- [x] **E9a — Simple constructor assignment receivers.** Explicit and bare constructors run
  once for `:=`, including inherited constructors and property setters. `Memory/obj_fields` and
  `SimpleScripts/override_deep` pass. Shipped 2026-09-13;
  [implementation and validation](docs/history/progress-log-2026-09.md#2026-09-13--constructor-results-as-assignment-receivers-33).
- [ ] **E9b — Compound assignments.** Make `T.Create().Value += 2` reuse one resolved receiver
  for the read and write. Extend counter/identity checks to fields, property access and failures
  that must prevent the write.
- [ ] **E9c — Var arguments.** Make `Mutate(T.Create.Value)` and the explicit-parentheses form
  bind storage from one receiver evaluation. Prove that mutation reaches that instance and that
  receiver exceptions prevent the call.
- [ ] **E9d — Bare free-function receivers.** Resolve `Make.Value := 42` as the implicit call
  form of the already working `Make().Value := 42`; check single evaluation and exception handling.

#### E10 — Remaining execution-suite triage `[ ]`

This ID covers the previously unnumbered remainder. Use
`just fixture-report --category <Category> --classify --list-fails` and record a representative
fixture and first blocking cause for each new group.

- [ ] **E10a — Reclassify the six remaining FunctionsMath failures.** E4's helper work is
  complete; identify the separate causes still preventing these fixtures from passing.
- [ ] **E10b — Group the smaller categories.** Triage `HelpersPass` (5), `OperatorOverloadPass`
  (3), `LambdaPass`/`OverloadsPass` (2 each), and `FunctionsGlobalVars`, `BuildScripts`,
  `FunctionsString`, `PropertyExpressionsPass` (1 each). Keep §5 exclusions explicit and link
  overlap with E3 instead of duplicating work. Memory's six host-setup failures remain gated in §3.3.
- [ ] **E10c — Separate the remaining `partial_class3` defects.** Its hint position is fixed,
  but the spurious `Result is never used` hint and `class 'TTest' already declared` runtime error
  still need independent reproductions and fixes.
- [ ] **E10d — Remove duplicate positions in pretty runtime diagnostics.** Exceptions whose
  messages already contain a position currently print it twice under `--diagnostics=pretty`.
  Add a CLI regression for arrays/strings and retain the plain/envelope rendering behavior.

#### E4 — Numeric and array helpers `[x]` — completed 2026-09-13

- [x] **E4a — Numeric helpers.** Integer `TestBit`, `PopCount`, `Compare` and Float `Compare`,
  including exact Integer comparison and numeric boundary behavior. FunctionsMath `compare_num`,
  `sort_nums`, `testbit` and `popcnt` pass.
- [x] **E4b — Dynamic Float-array transforms.** `Offset`, `Multiply`, `MultiplyAdd`,
  `Reciprocal` mutate and return the same receiver; empty arrays skip scalar operands.
  FunctionsMath `array_funcs` passes.
- [x] **E4c — String-array packing.** `Pack` and `StrArrayPack` stably remove empty strings
  in place; ArrayPass `string_array_pack` passes.
- [x] **E4d — Integration and validation.** Helper precedence, arity, operand evaluation and
  exceptions are covered. ArrayPass `array_method_indexing2` also passes. Full tests and
  changed-code lint pass; CLI/harness agree at 1,130/1,966. Baselines, guides and
  [history](docs/history/progress-log-2026-09.md#2026-09-13--numeric-and-array-helpers-e4) are updated.

#### E1 — Runtime-message vocabulary and re-raise behavior `[x]` — completed 2026-09-12

- [x] **E1a — Arithmetic, string bounds and external-call messages.** Match DWScript's
  vocabulary and preserve catchable bounds failures. SimpleScripts `div_by_zero_int`,
  `mod_by_zero_int`, `string_bounds2` and `external` pass.
- [x] **E1b — Re-raise the active exception.** `raise ExceptObject` preserves the original
  exception and reports the re-raise position; `SimpleScripts/re_raise` passes.

#### E2 — Structured hint positions `[x]` — completed 2026-09-12

- [x] **E2a — Calling-convention hints.** Share structured hint emission between methods and
  free routines; `SimpleScripts/call_conventions` passes.
- [x] **E2b — Partial-class hint position.** Remove the position embedded in the hint text and
  retain structured source coordinates. The fixture's remaining defects are tracked in E10c.

E1/E2 [implementation and validation](docs/history/progress-log-2026-09.md#2026-09-12--runtime-message-vocabulary-e1-e2).
Array bounds follow-up remains E8; pretty-renderer follow-up is E10d.

---

## 4. Error-detection parity (`*Fail` suites, F)

Harness and CLI: 166/641 (FailureScripts 156/529, SetOfFail 5, JSONConnectorFail 2,
AssociativeFail 1, InterfacesFail 1, JSFilterScriptsFail 1, every other `*Fail` suite 0). The suites are **compile-only** (like DWScript's
`CompilationFailure` runner): the expected file is the compiler's message list, hints included,
no envelope, and nothing is executed. Reproduce one with
`dwscript run --diagnostics=plain --compile-only --hints pedantic <file>`.

**Measured 2026-09-12**, every `*Fail` fixture diffed line-by-line against its expectation. The
tables, the per-shape inventories and the full near-miss list are in
[`docs/architecture/fail-suite-audit-2026-09.md`](docs/architecture/fail-suite-audit-2026-09.md);
only what to build is repeated here. **475 in-scope fixtures fail** (COMConnectorFailure's 8 are
host-library). **156 are one edit from passing and 281 are within two**, so working the
audit's near-miss queue across families often beats draining one family. ⚠️ Those two counts
replace the 68 / 194 recorded on 2026-09-12: the original script sorted both sides and compared
with `comm`, so a diagnostic emitted in the *wrong words* counted as two lines. It is one edit, and
`fixture-report --classify` now counts it as one (T8). Nothing about the port changed; re-derive
with `just fixture-report --in-scope --classify`. Two results reordered
what follows: go-dws's own invented message vocabulary blocks **265 of the 480** (F8, re-sized from
S to M), and the missing-validation sweep had only ever been counted over `FailureScripts`
(F5, +27 fixtures).

Work families — IDs from the 2026-03 analysis
(`docs/archive/failure-scripts-next-phase-plan.md`), counts from the 2026-09-12 re-measurement:

- **F1** `[~]` M Warning/hint emission and ordering. The for-loop half closed 2026-09-12 (see the
  table at the end of this section); what is left is ordering across bodies and the hints that do
  not exist yet.
  - `[ ]` S Ordering — `infinite_loop` wants **routine bodies before the main body**: `Trap`'s
    warnings at lines 6 and 3, then the main program's at 19, 21, 35. It is a side-effect of
    §3.2.1's deferred body checking, which should stay — re-order on the way out, not the analysis
    on the way in. E3a also confirmed a parser/semantic ordering mismatch in `hint_pedantic`:
    its line-9 hint must precede the line-11 unknown-switch diagnostic, independently of E3b's
    source-hint suppression fix.
  - `[ ]` S Hints and warnings that exist nowhere in the tree, lines (fixtures):
    `Unreachable code` 12 (5) · `Constant condition` 8 (5) ·
    `Redundant "begin" in clause of a case..of` (`case_of_else`) ·
    `Private virtual methods cannot be overridden` (`virtual_private`) ·
    `Redundant specifier, visibility is already "X"` (`class_visibility_redundant`) ·
    `"X" parameter is a reference type passed as VAR, but never written to`
    (`hint_reference_var_params`) · `Assigning a to itself` (`self_assign`).
  - Case-mismatch hints are tracked under E3. Its September 18 audit leaves one `FailureScripts`
    fixture with an extra case hint, `block_unfinished4`, tied to parser recovery (§4/F3).
- **F2** `[ ]` M Array diagnostics: `Array expected`, `Too many indices` (8 lines, all in one
  fixture), bound-exceeded wording, malformed array-type recovery, and
  `Range start and range stop are of incompatible types: "X" and "Y"` 9 (4).
- **F3** `[ ]` M Parser header/declaration/delimiter recovery. `"X" expected` is the largest
  missing shape at 70 lines over 64 fixtures — almost one per fixture, so wide and shallow rather
  than one deep bug. By token: `")"` ~23, `";"` ~11, `"]"` ~9, `"("` ~6, `"end"` 4, `">"` 4.
  `Name expected` 38 (33) and `Type expected` 15 (14) have the same shape. The blocker is F8:
  go-dws answers with its own sentence instead (`expected ')', got SEMICOLON`).
- **F4** `[ ]` L Class/property/static/override/visibility diagnostics — still the largest family.
  Leaders, lines (fixtures): `Method "X" of class "Y" not implemented` 34 (15) ·
  `Name "X" already exists` 20 (12) · `Class reference expected` 11 (9) ·
  `Class "X" isn't defined completely` 9 (7) and the `Interface` variant 4 (3) ·
  `There is already a field with name "X"` 8 (4) · `"X" is not a method of class "Y"` 7 (4).
- **F5** `[~]` M Missing-validation sweep — DWScript reports something, go-dws compiles **clean**.
  **43 in FailureScripts** (was 58 before the September slices) **plus 27 in the other suites,
  never previously counted**; the per-suite list is in the audit.
  - `[ ]` HelpersFail 10 is the densest and most coherent pocket: 10 of its 18 failures produce
    nothing at all, and 5 are one line from passing. Helpers accept far more than they should.
  - `[ ]` InterfacesFail 4 · JSONConnectorFail 3 · LambdaFail 3 · OverloadsFail 3 · GenericsFail 2
    · PropertyExpressionsFail 2.
  - The FailureScripts 43 are almost all single-fixture work. Known sub-blockers:
    `class_const4` needs `Constant Instruction - has no effect` as an **error** on a class-const
    declaration rather than a hint on a statement; `enum_flags_overflow` needs a per-element
    position on `ast.EnumValue` (only `EnumDecl` has one); `default_params2` needs a
    constant-folded comparison of two default-value expressions, which `mergeDefaultValues` has no
    helper for.
  - `func_ptr_mismatch` is silent for two reasons: `const` does not survive the
    `types.FunctionType` → `types.FunctionPointerType` conversion, which has no slot for parameter
    modifiers, so `@Test` is judged compatible with `procedure(Foo: string)`; and the message needs
    the routine-type renderer in F10.
- **F6** ✅ **Closed 2026-09-12.** No runtime-mismatch residue is left: every FailureScripts fixture
  with a blank expectation compiles clean, and none expects a `Runtime Error` line.
  `for_in_subclass`, the last name on the list, turned out to be message parity and moves to F8.
- **F7** `[ ]` M Per-suite sweeps. Failing / one line away / two or fewer: HelpersFail 18/5/9 ·
  InterfacesFail 18/2/8 · OverloadsFail 14/1/5 · PropertyExpressionsFail 10/3/6 · SetOfFail 9/1/7 ·
  GenericsFail 8/0/2 · JSONConnectorFail 7/1/2 · LambdaFail 6/2/3 · OperatorOverloadFail 6/0/1 ·
  AssociativeFail 3/0/2 · AttributesFail 2/0/0 · InnerClassesFail 1/0/1.
  - `[ ]` S **Sentence capitalization** — the cheapest item in the section. go-dws lowercases the
    first word of `overload of "X" will be ambiguous…`, `overloaded procedure "X" must be marked…`
    and `there is already a method with name "X"`, across five `OverloadsFail` fixtures. Two of the
    same kind: `AssociativeFail/contains` renders `"Nil"` for `"nil"`, and
    `FailureScripts/incorrect_type1` renders a builtin's parameter type as `"string"`.
  - `[ ]` S `The function "X" was forward declared but not implemented` exists nowhere in the tree
    — 8 lines over `OverloadsFail/forwards`, `forwards_unit`, `overload_func_ptr_param` and
    `FailureScripts/forward_missing1`.
  - `[ ]` S SetOfFail's nine are parser recovery and message parity; seven are within two lines,
    and `type_missing` is one `Type expected` away.
- **F8** `[ ]` **M, re-sized from S by measurement.** Replace go-dws's invented diagnostic
  vocabulary with DWScript's. The original framing — convert the remaining raw `addError(...)`
  sites to structured diagnostics (`analyze_function_calls.go` 54, `analyze_statements.go` 49,
  `analyze_method_calls.go` 17, `analyze_classes.go` 10;
  `docs/archive/semantic-legacy-hotspots-5.3.10.md`) — is right, but this is the precondition for
  **265 of the 480 failing fixtures**, not a cleanup, and the parser is in it as much as the
  analyzer.
  - `[ ]` Extract the worklist: 291 distinct shapes with the fixtures each one blocks, mechanical
    from the classification run (see **T8**).
  - `[ ]` Work it by shape, largest first, mapping each to the sentence it should be
    (`expected ')' after parameter list` → `")" expected`, `unknown type 'X'` → `Type expected`).
    Anchors must be measured per shape; the sentence is the easy half.
- **F9** `[ ]` M **The near-miss queue** — the 68 one-line fixtures, listed in the audit. Recurring
  themes, each one change: `Unexpected "Integer Literal"` (`property_error6`, `visibility5`) ·
  triple-apostrophe string diagnostics (`triple_apos1`, `triple_apos2`) · spurious
  `No arguments expected` on an array helper called with none (`dyn_array1`,
  `dyn_array_setlength2`) · `argument N to method 'X' of class 'Y' has type …` →
  `Argument N expects type "X" instead of "Y"` (`method_param_error1`, `method_param_error2`) ·
  spurious `Undefined variable 'Integer'` for an escaped reserved word (`reserved_escape_empty`,
  `reserved_escape_number`).
- **F10** `[ ]` M `Incompatible types: "X" and "Y"` — 58 lines over 22 fixtures, the largest missing
  *semantic* shape. DWScript uses one sentence wherever two types fail to unify, target first,
  supplied second, both quoted; go-dws invents a bespoke sentence per site, which is why the
  cluster spans four unrelated subsystems. Split, per the 2026-09-12 decision:
  - `[ ]` **Sentence and anchor only** — `array_initialization4`, `coalesce_dynarray`, `const_1`,
    `case_error5` (`for_error4` closed 2026-09-12 with the for-in work). ⚠️ The
    array-literal anchors need their own measurement first: `array_of_proc2` wants column 9, which
    is *whitespace after the comma*, and `array_of_proc` wants the `]`. These look like artifacts
    of upstream's scanner position rather than a token rule — the same ambiguity as the
    `Infinite loop` anchor declined in #400. Confirm or park.
  - `[ ]` **DWScript's routine-type rendering** — `class function ClassType: TClass`,
    `function IntToHex(Integer, Integer): String`, `destructor Destroy`,
    `procedure Test(const String)`, `procedure (String)`, `procedure TMyProc`. Rule: omit `()` when
    there are no parameters; an unnamed type keeps the separating space. Blocked on
    `types.FunctionPointerType` (`internal/types/function_pointer.go:13`) carrying no name, kind or
    parameter modifiers, and on `errors.SimplifyTypeName` (`internal/errors/errors.go:400`)
    truncating any signature at the first `(`. Extend `semanticNamedFunctionPointerName`
    (`internal/semantic/analyze_array_helpers.go:94`, pinned by
    `internal/frontend/result_test.go:417`) rather than starting over. Unlocks `func_ptr3`,
    `func_ptr4`, `func_ptr_mismatch`, and with `Destructor can only be invoked on instance`,
    `func_ptr5`.
  - `[ ]` The `Cannot assign "X" to "Y"` variant is a separate 37 lines over 17 fixtures with a
    different sentence. It **does** share a site: `for_in_subclass` drives the same for-in check
    that now emits `Incompatible types: "X" and "Y"` for `for_in1` and `for_error4`, but expects
    `Cannot assign "TBase" to "TChild"` because the two class types are related and the assignment
    narrows. ⚠️ Its anchor is column 12 of `for c in a do`, which is the `do` — not the `in` every
    other for-in diagnostic uses, and not the collection either. Measure that before implementing
    the split, the way the `Infinite loop` anchor was parked in #400.

**Shipped in this section (2026-09-12).** Full write-ups in
[the September progress log](docs/history/progress-log-2026-09.md):

| What | Fixtures |
| --- | --- |
| [No crashes, no hangs, on malformed input](docs/history/progress-log-2026-09.md#2026-09-12--no-crashes-no-hangs-on-malformed-input-4) — nine segfaults and one infinite loop, all from typed-nil AST nodes; 0 panics and 0 timeouts over all 2,127 fixtures | +1 |
| [DWScript's argument-count vocabulary](docs/history/progress-log-2026-09.md#2026-09-12--dwscripts-argument-count-vocabulary-4--f5) — `More arguments expected` / `Too many arguments` / `No arguments expected`, anchored at the name being called | +10 |
| [`Boolean expected`, and one hint per `if`](docs/history/progress-log-2026-09.md#2026-09-12--boolean-expected-and-one-hint-per-if-4) — the message names the type the context required and nothing else; an empty `repeat` body is legal | +6 |
| [The `deprecated` directive family](docs/history/progress-log-2026-09.md#2026-09-12--the-deprecated-directive-family-4--f5) — deprecation carried on symbols, methods and properties, warned at every use site | +5 |
| [The constant-instruction hint and the array-helper receiver rules](docs/history/progress-log-2026-09.md#2026-09-12--the-constant-instruction-hint-and-the-array-helper-receiver-rules-4--f5) — constness decided structurally, never by folding | +3 |
| [The expression-position implicit call](docs/history/progress-log-2026-09.md#2026-09-12--the-expression-position-implicit-call-4--f5) — a routine name reads as a call unless the context wants a pointer it actually fits | +1 |
| [The for-loop diagnostics](docs/history/progress-log-2026-09.md#2026-09-12--the-for-loop-diagnostics-4--f1-f10) — `Assignment to FOR-Loop variable`, the for-in half of `Empty FOR loop`, `Incompatible types` and `Enumeration expected` anchored at the `in`, and hints no longer reordered against errors | +7 |

Left open by those slices: `assert` and `enum_byname` want `Boolean expected` anchored at the
*argument* rather than the call; `ifthenelse_expression1` fails on parser recovery after
`if 2=2 1`; `contracts_error2` needs a builtin to resolve inside a `require` clause.

✋ `FailureScripts/class_deprecated` stays open on one position convention. Four of its eight
warnings are emitted with the right text but two columns late: for a *declaration's type
annotation* (`FField : TBase`, `function GetOther : TOther`, `property O : TOther`,
`var b : TBase`) upstream anchors the warning at the **colon**, while every expression-position
use is anchored at the identifier (confirmed by `const_deprecated` and `enum_element_deprecated`,
which now match exactly). All four samples are written `: T`, so they cannot distinguish "the
colon" from "the type name minus two", and the reference implementation is not checked out to
settle it. Implementing the colon reading means threading a colon position through nine
`warnDeprecatedResolvedType` call sites, so it was not guessed at. The fixture's other two
defects — `new TOther` anchored at `new` instead of the class name, and no warning for a
deprecated parent in `class(TBase)` — are fixed.

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
- ✋ Case-mismatch hint parity where runner configuration remains unverified: recover that
  configuration before pursuing parity. **E3a satisfies this gate for SimpleScripts, ArrayPass,
  HelpersPass, OverloadsPass and FailureScripts**: their bundled runner selects pedantic hints,
  with no per-fixture override. Algorithms uses the normal default, explaining July's
  `array_in2`/`hanoi` contrast. E3b source controls and the E3d audit are complete; E3c's bounded
  name-resolution residual remains separate from runner settings. See
  [the E3a audit](docs/history/progress-log-2026-09.md#2026-09-13--fixture-hint-configuration-e3a).
  The July "Hint/warning envelope" account remains historical, not the current blanket gate.
  `queue_snapshot` stays parked for the specific unresolved resolution behavior in §3.3.
  Non-case hints (empty block, unreachable code, prefer-ToString) remain under F1.
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
