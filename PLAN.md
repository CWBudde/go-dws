# go-dws — Work Plan

> Rewritten 2026-09-06, compacted 2026-09-19. This file lists **open work**; closed items are
> reduced to one-line pointers. Detailed completed work and the reasoning behind it live in the
> progress logs ([July](docs/history/progress-log-2026-07.md),
> [September](docs/history/progress-log-2026-09.md)); the measured audits behind the priorities are
> [`docs/history/CODEBASE_REVIEW_2026-07.md`](docs/history/CODEBASE_REVIEW_2026-07.md) (July 2026)
> and [`docs/architecture/audit-2026-09.md`](docs/architecture/audit-2026-09.md) (September 2026).
> Status numbers are generated, not hand-kept. Documentation index: [`docs/README.md`](docs/README.md).

## 0. Status snapshot

**Headline (2026-09-20):** Go harness and freshly rebuilt CLI agree at **1,203 / 1,966 scored =
61%**; `*Fail` error-detection suites **210 / 641 = 33%**. What shipped to get there is in
[the September progress log](docs/history/progress-log-2026-09.md), not here.

**What the denominator is.** 2,044 fixtures ship in the tree; 78 have no applicable expectation
and remain unscored, leaving 1,966. This includes 36 missing-`.txt` fixtures now checked against
silence (T7). The denominator still contains the **219 host-library fixtures excluded from every
target below** (see the scope rule further down) — all 219 currently fail. Excluding them, the
same run reads **1,203 / 1,747 = 69% in scope**, the number to track against §6. Both are honest;
the lower one is quoted outward. The [fixture README](testdata/fixtures/README.md) documents
which missing expectations are scored and which remain excluded.

Open, in leverage order:

- **§4** is where the remaining mass is: 431 in-scope `*Fail` failures. The 2026-09-12
  fixture-by-fixture measurement found go-dws's invented message vocabulary (F8) blocking 265 of
  the 480 failing then, split out the one-line near misses (F9) and the `Incompatible types`
  sentence (F10). Full tables:
  [`docs/architecture/fail-suite-audit-2026-09.md`](docs/architecture/fail-suite-audit-2026-09.md).
- **§3.5** is the other half: 113 in-scope execution-suite failures. E1–E2, E3a/b/d and E4–E11
  are closed; E3c is still open, and what the closed groups left is grouped as E12–E20.
- **§3.2** and **§3.4** have one open item each (explicit-instance helper calls; expected-type
  overload resolution, blocked on the evaluator). **§3.3** has only the gated Memory host setup.
- **§1** and **§3.1** have no open items. **§2** has one, deferred by owner decision.

Where the truth lives:

| Source | What it is | How to refresh |
| --- | --- | --- |
| `testdata/fixtures/TEST_STATUS.md` | generated per-category pass/fail table (harness) | `just fixture-update` |
| `testdata/fixtures/baselines.json` | per-category pass-count floor that CI enforces | `just fixture-update` after an intentional improvement |
| `just fixture-report` | end-to-end CLI numbers (`cmd/fixture-report` running `bin/dwscript run`) | `just build` first, or the numbers are stale |

Rules for this document:

- An item is closed only by a **passing fixture** (or a test that exercises the real user-facing path).
  Closed items are deleted from this file or reduced to a one-line pointer; their story goes to
  `docs/history/progress-log-<date>.md`.
- Ratchet `baselines.json` after every improvement (`just fixture-update`).
- The ~200 fixtures in host-library categories listed in
  [`docs/decisions/out-of-scope.md`](docs/decisions/out-of-scope.md) (DataBaseLib, COMConnector,
  CryptoLib, GraphicsLib, WebLib, TabularLib, TimeSeriesLib, DOMParser, Linq, LinqJSON, ClassesLib,
  DelegateLib, SystemInfoLib, IniFileLib, FunctionsFile, FunctionsRTTI, BigInteger,
  FunctionsMathComplex/3D) are excluded from every target below.
- Where the remaining failures are (763 total, 2026-09-20): **219 host-library** (out of scope),
  **431 in the `*Fail` error-detection suites** (§4: FailureScripts 336, InterfacesFail and
  HelpersFail 18 each, the rest under 15), and **113 in the execution suites** (§3.5:
  SimpleScripts 60, ArrayPass 16, JSONConnectorPass 9, InterfacesPass 6, FunctionsMath 1, a tail
  of ones and twos). A pre-existing runtime stack overflow still appears in an isolated harness
  worker; fixture scoring completes and the category baseline gate passes.
- Regenerate that split rather than trusting it: `just fixture-report --in-scope --classify`
  (T8). It reports each failure's distance from passing, whether what differs is a diagnostic or
  the program's output, and which message shapes recur — none of which `baselines.json` can see,
  because it holds pass-count floors.

Legend: `[ ]` open · `[~]` partially done, remainder listed · `[x]` done · ⏸️ gated, do not start ·
✋ won't-fix, with the decision record. Size: S (hours), M (days), L (week+).

---

## 1. Measurement & tooling (T)

No open items; new tooling items go here. T1–T8 closed 2026-09-06…13: one compile pipeline for
CLI and harness, `run --diagnostics=plain|pretty`, `--test-envelope`, `--compile-only`,
self-rebuilding `fixture-report` with a stale-binary guard, helper-spec parity, missing-expectation
scoring and category hint levels (T7), and the failure classifier `fixture-report --in-scope
--classify` (T8). See [the September progress log](docs/history/progress-log-2026-09.md).

---

## 2. Architecture refactoring — required (A)

A2–A10 closed 2026-09-07/09 (structural type resolution, typed runtime registries and canonical
callables, typed operator/value dispatch, builtin signature constraints); see
[the September progress log](docs/history/progress-log-2026-09.md). The current architecture is
[`docs/architecture/interp-evaluator-steady-state.md`](docs/architecture/interp-evaluator-steady-state.md).

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

§3.1–§3.4 are organised by *subsystem* and are nearly closed. §3.5 is organised by *failing
suite*: it holds the execution-suite failures the subsystem view never surfaced.

### 3.1 Parser

No open items. Closed 2026-09-09/12: `class property` in record bodies, expression-backed and
multi-index indexed properties, `external` properties, nested `>>` in generic argument lists,
case-insensitive keyword operators; `array of T` as a getter-position property type already worked.

- ✋ `{$I %FILE%}`-style value substitution: the only fixture, SimpleScripts `include_expr`, cannot
  pass. Its expected output hardcodes the original Delphi runner's paths (`Test\include_expr.pas`,
  `*MainModule*`), and neither `cmd/fixture-report` `normalize()` nor the Go harness normalizes
  paths; `%FUNCTION%` is also not knowable at lex time. Reopen only with a path-normalizing harness.

### 3.2 Semantic

L-S1…L-S7 closed 2026-09-09/10: phased class construction, private-field and helper-property
diagnostics, contracts through the ancestor chain, generics (GenericsPass 100%), overloads and
method pointers, sets (SetOfPass 100%), conditional compilation; table and write-ups in
[the September progress log](docs/history/progress-log-2026-09.md). Documented, not fixed (L-S1c):
`var c := TC.Create;` written *before* the abstract ancestor's declaration misses the
abstract-instantiation error, because checking is in source order. The
[semantic-passes design](docs/architecture/semantic-passes.md) is design input only.

- `[ ]` S **Explicit-instance helper calls.** `THelper.Proc(TObject.Create)` — a helper method
  called with an explicit instance argument — is an unimplemented call form; it reports
  `No arguments expected` at 11:10 in `HelpersPass/declared_helper`, whose `Declared()` calls and
  case hints already match.

Moved out of this section: the error-detection residue (`contracts_precondition`, `GenericsFail`,
`OverloadsFail`, the other nine `SetOfFail`, `special_funcs4`, `conditionals2.1`) is under §4/F7;
`read_write_other_property` is E18. Until `contracts_precondition` lands, the runtime evaluates
every `require` in the chain, root-most first.

Deliberate exclusions, each measured; reopen only on new evidence:

- ✋ Unused-private-field hints alongside a compile error: `internal/semantic/unused_warnings.go`
  suppresses them wholesale, contradicting `OverloadsFail/overloads_not_implem`, which fails to
  parse for unrelated reasons. Reopen once it parses.
- ✋ No fixture demands (measured 2026-09-09/11): unused-private hints for record fields and class
  vars (no usage tracking); record-type metaclass member access via the type name; class
  invariants (parsed into `ClassDecl.Invariants`, never evaluated); type-parameter constraints
  `<T: TObject>` (parsed, ignored); function pointer `= nil` comparison (still "operator =
  requires comparable types"); for-in loop variable not checked against a set's element type.
- ✋ `FailureScripts/static_methods` passed only because `{$FATAL}` was ignored; upstream's
  expectation omits the `Compile Error` line although the directive is present. The only
  difference from the three reported cases is that its `{$FATAL}` is not at column 1 — 5/5
  correlation, no plausible mechanism, not encoded. Reopen if the reference submodule is checked out.
- ✋ `ConditionalDefined(s)` always folds to `False`; `{$DEFINE}` symbols live in preprocessor
  state the analyzer cannot reach. Argument validation is complete.
- ✋ Subrange bounds at compile time: no fixture declares a subrange type; zero yield.

### 3.3 Runtime / evaluator

Closed: FunctionsByteBuffer, FunctionsTime, FunctionsVariant, FunctionsDebug, InnerClassesPass,
EncodingLib (all 100%); the 2026-09-12 runtime-panic re-measurement (no panics; the three areas
failed for ordinary reasons, now fixed). See
[the September progress log](docs/history/progress-log-2026-09.md).

- ⏸️ **Memory — 7 of 13 scored (2026-09-13).**
  - ⏸️ The six `external*` fixtures need a **host-registered external class** — upstream's `SetUp`
    registers `TExposedClass` and `TExposedBoomClass` with host-side constructors and an
    `OnCleanUp` hook (`UMemoryTests.pas:86-118`). That is host-integration surface, §5 territory.
    go-dws answers `parent class 'TExposedClass' not found`, which is right for a host that
    registered nothing.
  - ✋ Upstream's leak assertions (`exec.ObjectCount = 0`, external-object count) do not port: they
    check DWScript's reference counting at a point where Go's GC has not necessarily run. Five of
    the formerly unscored fixtures exist only to make that assertion; "compiles and prints nothing" is
    all of it that is portable.
- ✋ UTF-16 surrogate iteration (`for_in_str`, `for_in_str2`): intentional divergence, see
  [`docs/decisions/string-encoding.md`](docs/decisions/string-encoding.md).

### 3.4 Source TODO backlog

Live `// TODO` markers that are real work, not notes. Bytecode TODOs are omitted (A11). Closed
2026-09-11/12: the `platform.Platform` engine seam, two stale items (class-distance overload
matching, class-operator inheritance), and the skipped-test backlog.

- `[ ]` Expected-type overload resolution (was `analyze_function_calls.go:26`): the dead `expectedType`
  parameter is gone; the note now sits at the dispatch in `internal/semantic/analyze_expressions.go`.
  Return type is not part of overload identity (`types.SignaturesEqual`), so an expected type can only
  be a last-resort tie-break. Resolving one in the analyzer alone admits programs the AST evaluator
  then runs with a different overload — it resolves independently at run time
  (`evaluator.ResolveOverloadMultiple`) with no expected-type channel. Blocked until the evaluator can
  see the call site's expected type, or reuse the analyzer's choice via `ast.SemanticInfo`.

### 3.5 Execution-suite failures (E)

114 in-scope execution-suite fixtures fail (2026-09-19). Audits:
[pass-suite audit](docs/architecture/pass-suite-audit-2026-09.md) (2026-09-12) and
[execution-suite triage](docs/architecture/execution-suite-triage-2026-09.md) (2026-09-19, first
blocker per group). Regenerate with `just fixture-report --in-scope --classify --list-fails`.

Most of these are classified `mixed` (a diagnostic *and* the output differ). For an execution
suite that is almost always one fault — a spurious compile error stops the program, so its output
goes missing too — and distance overstates it: a one-line spurious error on a program printing
forty lines scores 41.

Write a failing test through the real compile/run path before implementing a fix. Close
subtasks with passing fixtures or real-path regressions; refresh category baselines and
CLI/harness parity when fixtures improve.

**Closed** (write-ups in [the September progress log](docs/history/progress-log-2026-09.md)):

| ID | What shipped | Fixtures |
| --- | --- | --- |
| E1 (09-12) | DWScript runtime-message vocabulary; `raise ExceptObject` re-raises the active exception | `div_by_zero_int`, `mod_by_zero_int`, `string_bounds2`, `external`, `re_raise` |
| E2 (09-12) | Structured calling-convention and partial-class hint positions | `call_conventions` |
| E3a/b/d (09-13…18) | Pedantic hints confirmed for all five E3 categories; source hint controls; case-hint re-measurement | — |
| E4 (09-13) | Integer/Float helpers, Float-array transforms, string-array packing | FunctionsMath +5, ArrayPass +2 |
| E5 (09-18) | Interface casts, comparisons, `implements` on class references | InterfacesPass 21 → 27/33 |
| E6 (09-18) | JSON strings in global storage; associative-array serialization | JSONConnectorPass 69 → 73/82 |
| E7 (09-19) | `String.Replace` helper; Variant conditions in `Assert` | `ignore_result`, `assert_variant` |
| E8 (09-19) | Array-element var arguments keep their resolved storage; bounds positions | `const_array_empty`, `array_element_byref` |
| E9 (09-13…19) | Single evaluation of assignment, compound-assignment and var-argument receivers | `Memory/obj_fields`, `override_deep` |
| E10 (09-19) | Triage of the remainder; partial-class continuations; value-bearing `Exit` uses Result; no duplicate positions in pretty runtime diagnostics | `partial_class3`, `implies` (SimpleScripts 381/443) |
| E11 (09-19) | Variant `Abs`, Variant `Inc`/`Dec` deltas, two-argument `Succ`/`Pred`, Haversine radius, `RandG(mean, stdDev)` | FunctionsMath 35 → 39/40 |

#### E3c — Declared-name handling `[~]` M

Declaration-spelling and duplicate-emission regressions are in place. The September 18 audit
leaves two fixtures lacking hints and three with extra hints; only `record_array_helper` is an
isolated case-hint residual. None of these may be hidden with hint suppression.

- `[ ]` **`HelpersPass/record_array_helper`**: establish the upstream resolution rule for the two
  unwanted `X`/`x` hints at 23:21 and 23:35 (output is correct). Do not suppress record-field
  hints broadly or weaken case-insensitive lookup.
- `[ ]` Recheck once prerequisites land: `SimpleScripts/class_var_dyn2` after class-variable scope
  resolution; `string_builtin_methods` after the missing string APIs;
  `ArrayPass/dynamic_anonymous_record` (see E13c) and `FailureScripts/block_unfinished4` after
  parser recovery (§4/F3).

#### E12 — Interface follow-ups (from E5) `[ ]`

InterfacesPass 27/33. Do not broaden E5's casting/comparison fixes to silence these.

- `[ ]` S `interface_nil_cast_from_obj`: `Impl` is tokenized as the reserved `IMPL` token;
  resolve identifier handling separately and retain a nil-object cast regression using `Obj`.
- `[ ]` M `interface_properties`: indexed/default interface property semantics.
- `[ ]` M `intf_delegate`: interface method references assigned to procedure variables.
- `[ ]` S `intf_in_record`: restore the caller location in the runtime-error trace.
- `[ ]` S `intf_private`: unwanted unused-private hints for `Hello` and `Unused` (§4/F1).
- `[ ]` M `intf_self_ref`: resolve self-referential interface method return types.
- `[ ]` Mixed class/interface equality: semantic analysis accepts operands that the evaluator
  rejects; establish compatible reference-comparison behavior separately from interface/interface
  equality.
- `[ ]` Interface alias assignment: assigning an object directly to an interface alias is
  rejected, while assigning through a variable of the underlying interface type works.

#### E13 — JSON follow-ups (from E6) `[ ]`

JSONConnectorPass 73/82.

- `[ ]` a **Associative-key conversion** (`implicit_associative_key_cast`): normalize JSON
  numeric keys to the declared String key type so a lookup using `'123'` finds a JSONVariant
  key holding 123.
- `[ ]` b **JSONVariant parameter conversion** (`implicit_from_cast`): materialize the declared
  JSONVariant representation for Null, unassigned Variant and primitive arguments before
  method dispatch; `Test(Null)` currently attempts `TypeName` on raw NULL.
- `[ ]` c **Inline record array types** (`const_array`, `stringify_array_of_array`): resolve anonymous
  record element types in static and dynamic arrays; compilation currently stops before JSON
  serialization runs.
- `[ ]` d **Coercive comparison** (`comparison2`): reconcile numeric JSON/string comparisons,
  retaining lexical distinctions for string/string comparisons.
- `[ ]` e **Coercive membership** (`in_static`): JSON/native and Variant/native `in`.
- `[ ]` f **Invalid array deletion** (`delete_array_index`): raise a catchable exception for
  an invalid index and leave the array unchanged; the failed deletion is currently ignored.
- `[ ]` g **Circular references** (`circular_references`): reject self and transitive cycles
  before changing ownership in object assignment and array insertion.
- `[ ]` h **Immediate JSONVariant assignment** (`write_immediate_prop`): route member/index writes
  through JSON mutation handling so primitive-backed JSONVariants produce the expected
  catchable `Immediate`/`String` diagnostics.
- `[ ]` i **Duplicate textual Variant keys** (no fixture; found in review): distinct associative
  keys such as Integer `1` and String `'1'` coexist, but conversion through a JSON object collapses
  their identical member names. Upstream `StringifyAssociativeArray` writes each occupied bucket
  directly, retaining both names. Add a real-path regression and preserve duplicate names when
  emitting JSON text; establish `JSON.Serialize` behavior separately before changing the JSON
  value representation.

#### E14 — Seeded RNG and Variant-to-Integer assignment (from E10, E11) `[ ]`

FunctionsMath 39/40.

- `[ ]` M Seeded RNG compatibility (`randseed`): establish the upstream sequence/seed contract
  before changing the generator. Its missing deprecated warning belongs to F1.
- `[ ]` S Variant-to-Integer assignment (no fixture; found in E11): `var v : Variant := 2.5;
  var i : Integer := v;` stores 2.5 in `i` instead of casting it. Builtin arguments already round
  through `coerceToInteger`; the assignment path is the one that differs.

#### E15 — Helper dispatch and receivers (from E10) `[ ]`

- `[ ]` S `classname_helper1`: a helper's `ClassName` calling `Self.ClassName` recursively selects
  itself (times out); expected `Helper.TObject`. Add a bounded recursion regression first.
- `[ ]` S `dyn_array_create`: class functions through an array type alias
  (`TStrings.Create(...)` → `Unknown name "TStrings"`).
- `declared_helper` is the §3.2 explicit-instance item.

#### E16 — Operator syntax and binding (from E10) `[ ]`

- `[ ]` M C-style expression operators `==`, `!=` (`c_style`) and `<<`, `>>`
  (`operator_overloading2`): parse and follow through semantic/operator dispatch.
- `[ ]` S Qualified helper operator bindings (`helper_as_overload`: `uses TVec2Helper.Add`), with
  overload selection.
- `[ ]` S Builtin implicit-operator bindings (`operator_implicit`: binding `IntToStr` not found).

#### E17 — Lambda and overload semantics (from E10) `[ ]`

- `[ ]` S `simple_func`: recognize `Result` case-insensitively (`pkg/ident`) during lambda return
  inference, including conditional branches.
- `[ ]` M `overload_ambiguous_delegate`: resolve bare callable arguments against value/delegate
  overloads (implicit invocation vs. explicit delegate).
- `[ ]` S `overload_class_method`: retain class identity for bare `ClassName` inside a class method.
- `immediate`'s synthetic unused-Result hint belongs to F1.

#### E18 — Property forwarding and helper error positions (from E10) `[ ]`

- `[ ]` M `read_write_other_property`: property accessor specifiers naming another property
  (`property Mapped read Prop write Prop`), including access-mode validation.
- `[ ]` S `toxml`: anchor the caught exception at the member name (`ToXML`, column 12) rather than
  the receiver.

#### E19 — BuildScripts runner parity (from E10) `[ ]` M

- `[ ]` Select the upstream `.dws` drivers (both scanners enumerate `.pas` today).
- `[ ]` Resolve their `.pas` units without selecting the same-name driver.
- Reconcile CLI/harness scoring before changing category discovery or the denominator; do not
  simply mark `const_inline` passing.

#### E20 — Compound index assignment (from E9) `[ ]` S

The read and write of a compound index assignment still resolve the index expression
independently; evaluate it once, as E9 does for member receivers.

---

## 4. Error-detection parity (`*Fail` suites, F)

Harness and CLI: 210/641 (FailureScripts 193/529, SetOfFail 6, OverloadsFail 3, JSONConnectorFail 2,
PropertyExpressionsFail 2, AssociativeFail 2, InterfacesFail 1, JSFilterScriptsFail 1, every other `*Fail` suite 0). The suites are **compile-only** (like DWScript's
`CompilationFailure` runner): the expected file is the compiler's message list, hints included,
no envelope, and nothing is executed. Reproduce one with
`dwscript run --diagnostics=plain --compile-only --hints pedantic <file>`.

**Measured 2026-09-12**, every `*Fail` fixture diffed line-by-line against its expectation; tables,
per-shape inventories and the near-miss list are in
[`docs/architecture/fail-suite-audit-2026-09.md`](docs/architecture/fail-suite-audit-2026-09.md).
**433 in-scope fixtures fail** (2026-09-19) (COMConnectorFailure's 8 are host-library). **156 are one edit from
passing and 281 are within two** (T8's counting; a wrongly worded diagnostic is one edit), so
working the near-miss queue across families often beats draining one family. Re-derive with
`just fixture-report --in-scope --classify`.

**Closed** 2026-09-12 (+33 fixtures; [write-ups](docs/history/progress-log-2026-09.md)): no crashes
or hangs on malformed input, DWScript's argument-count vocabulary, `Boolean expected`, the
`deprecated` directive family, the constant-instruction hint and array-helper receiver rules, the
expression-position implicit call, the for-loop diagnostics. **F6** (runtime-mismatch residue)
closed outright.

Work families — IDs from the 2026-03 analysis
(`docs/archive/failure-scripts-next-phase-plan.md`), counts from the 2026-09-12 re-measurement:

- **F1** `[~]` M Warning/hint emission and ordering. The for-loop and ordering halves are closed;
  what is left is the hints that do not exist yet.
  - Ordering closed 2026-09-20 ([log](docs/history/progress-log-2026-09.md)): deferred routine
    bodies splice their diagnostics back to the declaration point (`infinite_loop`,
    `ArrayPass/array_of_rec_add_create`), and a compiler-directive diagnostic orders by line
    against a semantic hint or warning (`hint_pedantic`).
  - `[ ]` S Inline class method bodies stay queued until the last top-level class declaration and
    are then drained as one batch (`drainDeferredMethodBodies`, `class_construction.go:343`), so a
    statement or routine written *between* two class declarations reports before both class bodies.
    Pre-existing and independent of the splice above: a bare `while True do ;` between two classes
    with inline bodies orders `9, 5, 15` on main and on the ordering branch alike. No fixture covers
    the shape; settle it against upstream's emission before changing the drain.
  - `[ ]` Hints and warnings that exist nowhere in the tree, lines (fixtures) — one subtask each:
    - `[ ]` M `Unreachable code` 12 (5)
    - `[ ]` M `Constant condition` 8 (5) — also needed by `contracts_precondition`
    - `[ ]` S `Redundant "begin" in clause of a case..of` (`case_of_else`)
    - `[ ]` S `Private virtual methods cannot be overridden` (`virtual_private`)
    - `[ ]` S `Redundant specifier, visibility is already "X"` (`class_visibility_redundant`)
    - `[ ]` S `"X" parameter is a reference type passed as VAR, but never written to`
      (`hint_reference_var_params`)
    - `[ ]` S `Assigning a to itself` (`self_assign`)
  - `[ ]` Unused-symbol ownership: synthetic `Result is never used` on expression lambdas
    (`LambdaPass/immediate`), unused-private hints on interface implementers (E12 `intf_private`),
    missing deprecated warning in `randseed` (2:9).
  - Case-mismatch hints are tracked under E3c; `block_unfinished4`'s extra case hint is tied to
    parser recovery (F3).
- **F2** `[ ]` M Array diagnostics:
  - `[ ]` S `Array expected`.
  - `[ ]` S `Too many indices` (8 lines, all in one fixture).
  - `[ ]` S Bound-exceeded wording.
  - `[ ]` M Malformed array-type recovery.
  - `[ ]` S `Range start and range stop are of incompatible types: "X" and "Y"` 9 (4).
- **F3** `[ ]` M Parser header/declaration/delimiter recovery. Wide and shallow rather than one
  deep bug. The blocker is F8: go-dws answers with its own sentence instead
  (`expected ')', got SEMICOLON`). Split by shape:
  - `[ ]` `"X" expected` — the largest missing shape, 70 lines over 64 fixtures. By token:
    `")"` ~23, `";"` ~11, `"]"` ~9, `"("` ~6, `"end"` 4, `">"` 4; work token by token.
  - `[ ]` `Name expected` 38 (33).
  - `[ ]` `Type expected` 15 (14).
  - Left open by the 2026-09-12 slices: `ifthenelse_expression1` fails on recovery after `if 2=2 1`.
  - `[ ]` Left open by the 2026-09-19 compiler-stop slice: inside `begin…end` the value left
    unconsumed after a read-only property assignment is worded differently upstream (go-dws
    reports nothing); indexed read-only property writes get no follow-up; semantic
    diagnostics after a parser stop are dropped by position, but the analyzer's own stops
    (`"(" expected`) do not suppress later diagnostics, and end-of-compilation hints
    positioned before a parser stop are still reported. The
    `must have either a type annotation` text filter in `internal/frontend/result.go` is now
    mostly dead.
- **F4** `[ ]` L Class/property/static/override/visibility diagnostics — still the largest family.
  One subtask per message shape, lines (fixtures):
  - `[ ]` M `Method "X" of class "Y" not implemented` 34 (15)
  - `[ ]` M `Name "X" already exists` 20 (12)
  - `[ ]` S `Class reference expected` 11 (9)
  - `[ ]` S `Class "X" isn't defined completely` 9 (7) and the `Interface` variant 4 (3)
  - `[ ]` S `There is already a field with name "X"` 8 (4)
  - `[ ]` S `"X" is not a method of class "Y"` 7 (4)
- **F5** `[~]` M Missing-validation sweep — DWScript reports something, go-dws compiles **clean**.
  **43 in FailureScripts plus 27 in the other suites**; the per-suite list is in the audit.
  - `[ ]` HelpersFail 10 is the densest and most coherent pocket: 10 of its 18 failures produce
    nothing at all, and 5 are one line from passing. Helpers accept far more than they should.
  - `[ ]` InterfacesFail 4 · JSONConnectorFail 3 · LambdaFail 3 · OverloadsFail 3 · GenericsFail 2
    · PropertyExpressionsFail 2.
  - The FailureScripts 43 are almost all single-fixture work. Known sub-blockers:
    `class_const4` needs `Constant Instruction - has no effect` as an **error** on a class-const
    declaration rather than a hint on a statement; `enum_flags_overflow` needs a per-element
    position on `ast.EnumValue` (only `EnumDecl` has one); `default_params2` needs a
    constant-folded comparison of two default-value expressions, which `mergeDefaultValues` has no
    helper for; `contracts_error2` needs a builtin to resolve inside a `require` clause.
  - `func_ptr_mismatch` is silent for two reasons: `const` does not survive the
    `types.FunctionType` → `types.FunctionPointerType` conversion, which has no slot for parameter
    modifiers, so `@Test` is judged compatible with `procedure(Foo: string)`; and the message needs
    the routine-type renderer in F10.
- **F7** `[ ]` M Per-suite sweeps. Failing / one line away / two or fewer: HelpersFail 18/5/9 ·
  InterfacesFail 18/2/8 · OverloadsFail 11 · PropertyExpressionsFail 8 · SetOfFail 8 ·
  GenericsFail 8/0/2 · JSONConnectorFail 7/1/2 · LambdaFail 6/2/3 · OperatorOverloadFail 6/0/1 ·
  AssociativeFail 2 · AttributesFail 2/0/0 · InnerClassesFail 1/0/1.
  - Closed 2026-09-19 ([log](docs/history/progress-log-2026-09.md)): overload sentence
    capitalization, `"nil"`/builtin type spelling, and `The function "X" was forward declared but
    not implemented` (+7). Left open from that slice:
    - `[ ]` S The class-method path still says `duplicate method signature` for
      `There is already a method with name "X"` (`empty_body`, `member_duplicates`, `method_implem`).
    - `[ ]` S Analyze a routine's body even when its declaration fails the overload check
      (`forwards_unit`'s `IntToHex` error on line 23); `forwards_unit` also gets a spurious
      `Unit name does not match file name` warning.
    - `[ ]` M Overloads differing only in return type are ambiguous in DWScript
      (`overload_simple`); go-dws allows them.
    - `[ ]` M Method overload rules — hiding, visibility, `no overloaded version declared`
      (`meth_overload_simple`, `meth_overload_hide`, `meth_private_public`, `overload_missing`).
    - `[ ]` S Forward/implementation mismatch wording (`Declaration should be…`,
      `Value-parameter expected`, default-value mismatch): `declaration_mismatch1`/`2`,
      `default_params2`.
    - The analyzer's compile stop is so far one flag (unknown name in an expression) that skips
      the end-of-program forward check; see the F3 compile-stop note.
  - `[ ]` S SetOfFail's remaining five. Closed 2026-09-20
    ([log](docs/history/progress-log-2026-09.md)): `bracket_left_missing`, `include` and
    `invalid_method` (set-mutator sentences and the declared set-type name); `type_missing`
    closed 2026-09-19. Still open:
    - `[ ]` `bracket_right_missing`, `for_in_set_missing_do`, `of_missing` are parser recovery
      and message parity (F3's `"X" expected` / `OF expected` / `DO expected` sweep).
    - `[ ]` `test_non_variable` needs only the deferred-body hint ordering (F1); both its
      `Variable expected` sentences and anchors are in place.
    - `[ ]` `invalid_operand` is three lines short: `unexpected "@"` exists nowhere in the tree
      (also wanted by `FailureScripts/at_integer`, `dyn_array3`, `field_init1`, `func_ptr6`),
      `Incompatible types: "TMyEnum" and "procedure Test"` needs F10's routine-type renderer,
      and its three line-13 diagnostics are expected at columns 12, 12, 10 — `sortDiagnostics`
      orders mixed-phase same-line diagnostics by column, so it would print 10 first.
  - `[ ]` GenericsFail 8: includes `implem_mismatch1`'s "T expected but u found" for a mismatched
    out-of-line type-parameter name (substitution is currently positional).
  - `[ ]` S Small parity items from §3.2: `special_funcs4` (`Expression expected` for `Inc(i, )`),
    `conditionals2.1` (unbalanced report at the directive argument, column 9, where the
    byte-identical `conditionals2` wants the name, column 3),
    `contracts_precondition` (`Preconditions must be defined in the root method only`),
    `assert` and `enum_byname` (`Boolean expected` anchored at the *argument*, not the call).
- **F8** `[ ]` **M, re-sized from S by measurement.** Replace go-dws's invented diagnostic
  vocabulary with DWScript's. The original framing — convert the remaining raw `addError(...)`
  sites to structured diagnostics (`analyze_function_calls.go` 54, `analyze_statements.go` 49,
  `analyze_method_calls.go` 17, `analyze_classes.go` 10;
  `docs/archive/semantic-legacy-hotspots-5.3.10.md`) — is right, but this is the precondition for
  **265 of the 480 failing fixtures**, not a cleanup, and the parser is in it as much as the
  analyzer.
  - `[ ]` S Extract the worklist: 291 distinct shapes with the fixtures each one blocks, mechanical
    from the classification run (see **T8**). Commit it next to the fail-suite audit so the
    shape-by-shape work below can be checked off.
  - `[ ]` Work it by shape, largest first, mapping each to the sentence it should be
    (`expected ')' after parameter list` → `")" expected`, `unknown type 'X'` → `Type expected`).
    Anchors must be measured per shape; the sentence is the easy half. Batch by origin:
    - `[ ]` Parser shapes (overlaps F3).
    - `[ ]` `analyze_function_calls.go` / `analyze_method_calls.go` sites.
    - `[ ]` `analyze_statements.go` sites.
    - `[ ]` `analyze_classes.go` sites (overlaps F4).
- **F9** `[~]` M **The near-miss queue** — the 68 one-line fixtures, listed in the audit. The
  compiler-stop slice (2026-09-19, [log](docs/history/progress-log-2026-09.md)) closed 13 of them
  plus 5 others: the `Unexpected "Integer Literal"`, spurious `expected ')'`, `enums5`/`6`,
  `var_incomplete`, `case_error3` and `ifthenelse_expression2` lines. The string-constant slice
  (same day) closed `triple_apos1`/`2`, `heredoc`, `invalid_ucs2_char` and
  `reserved_escape_empty`/`_number`, plus SimpleScripts `heredoc_indent`/`heredoc_special`.
  - `[ ]` S `reachedLexerDiagnostics` (`internal/frontend/result.go`) is now exact for the main
    source — only a compiler stop cuts lexer diagnostics off, which is what `ECompileError`
    does upstream — but the unit-compile path (`internal/frontend/units.go`) does not apply the
    cutoff at all, so a stop inside a unit still lets later directives through.
  The call-argument slice (same day) closed `method_param_error1`/`2`, `dyn_array1`,
  `dyn_array_setlength2`, `open_array2`, `use_proc_result2`, `foreach_invalid_arg`, `assigned`.
  Left open from it:
  - `[ ]` S The invented `argument N has type …` sentence survives on paths no fixture pins yet:
    member calls, implicit-Self calls, record class methods, the implicit helper path and
    constructors (`analyze_function_calls.go`), `inherited` calls (`analyze_special.go`), two
    sites in `analyze_classes.go`, set `Include`/`Exclude` (`analyze_method_calls.go`). Move them to
    `analyzeCallArgument`/`analyzeSelfCallArgument` (F8).
  - `[ ]` M A `const` parameter of type `array of Variant` is conflated with `array of const`;
    the analyzer tells them apart by declared type name. Separate the types.
  - `[ ]` S `internal_unsupported`: `Length`/`Low`/`High` want `Invalid argument type` (reuse the
    new `Assigned` check) and `Inc` wants `Integer expected`.
  - `[ ]` S `assign_untyped`: `Assignment's right-side-argument has no return type`, and
    `Cannot assign a value to the left-side argument` for assigning to a procedure name.
  - `[ ]` S `enum_byname` wants `String expected` (upstream's dedicated `ByName` check);
    `func_ptr_var_param` wants the routine-type name `procedure TProc` (F10);
    `lazy_func_ptr` wants `Lazy parameter cannot be a function pointer`.
- **F10** `[ ]` M `Incompatible types: "X" and "Y"` — 58 lines over 22 fixtures, the largest missing
  *semantic* shape. DWScript uses one sentence wherever two types fail to unify, target first,
  supplied second, both quoted; go-dws invents a bespoke sentence per site, which is why the
  cluster spans four unrelated subsystems. Split, per the 2026-09-12 decision:
  - `[ ]` **Sentence and anchor only** — `array_initialization4`, `coalesce_dynarray`, `const_1`,
    `case_error5`. ⚠️ The
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
    `func_ptr5`. Two steps:
    - `[ ]` M Carry name, kind and parameter modifiers on `types.FunctionPointerType`.
    - `[ ]` S Render them and stop `SimplifyTypeName` truncating at the first `(`.
  - `[ ]` The `Cannot assign "X" to "Y"` variant is a separate 37 lines over 17 fixtures with a
    different sentence. It **does** share a site: `for_in_subclass` drives the same for-in check
    that now emits `Incompatible types: "X" and "Y"` for `for_in1` and `for_error4`, but expects
    `Cannot assign "TBase" to "TChild"` because the two class types are related and the assignment
    narrows. ⚠️ Its anchor is column 12 of `for c in a do`, which is the `do` — not the `in` every
    other for-in diagnostic uses, and not the collection either. Measure that before implementing
    the split, the way the `Infinite loop` anchor was parked in #400.

✋ `FailureScripts/class_deprecated` stays open on one position convention. Four of its eight
warnings are emitted with the right text but two columns late: for a *declaration's type
annotation* (`FField : TBase`, `function GetOther : TOther`, `property O : TOther`,
`var b : TBase`) upstream anchors the warning at the **colon**, while every expression-position
use is anchored at the identifier (confirmed by `const_deprecated` and `enum_element_deprecated`).
All four samples are written `: T`, so they cannot distinguish "the colon" from "the type name
minus two", and the reference implementation is not checked out to settle it. Implementing the
colon reading means threading a colon position through nine `warnDeprecatedResolvedType` call
sites, so it was not guessed at.

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
  configuration before pursuing parity. SimpleScripts, ArrayPass, HelpersPass, OverloadsPass and
  FailureScripts are verified (pedantic, [E3a](docs/history/progress-log-2026-09.md#2026-09-13--fixture-hint-configuration-e3a));
  Algorithms uses the normal default. Non-case hints remain under F1.
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
