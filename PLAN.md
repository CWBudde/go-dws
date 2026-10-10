# go-dws — Work Plan

Work is split into phases → tasks → subtasks. Keep completed tasks checked while their
subphase is open, so progress stays visible. When the whole subphase closes, remove its
checklist and retain the write-up in `docs/history/progress-log-<date>.md`. The completed
1.3 checklist is retained as a Phase 1 milestone checkpoint. Documentation index:
[`docs/README.md`](docs/README.md).

**Goal (v1.0):**

1. `just fixture-report` and the Go harness both report **≥ 90 %** on every in-scope category and
   agree with each other.
2. All `*Fail` suites reproduce DWScript's diagnostics.
3. One type representation, one evaluator, one compile pipeline.
4. CI fails on any per-category regression (in place: `baselines.json` gate).
5. No public API or CLI flag exposes a non-functional mode without saying so.

**Rules:**

- A task closes only with a **passing fixture** or a test through the real user-facing
  compile/run path. Write that test first.
- **Definition of done: scope comes from fixtures.**
  - Every open task names its target fixture(s), or a measured `--classify` shape count, when
    it is opened.
  - Work with no fixture is a ⚠️ measure item. If no fixture turns up, it moves to
    [`known-divergences.md`](docs/decisions/known-divergences.md#behaviour-without-a-fixture-that-demands-it).
  - Real-path tests are regression controls for fixture-backed work. They are not a closing
    criterion on their own, except for refactors whose acceptance is "nothing regresses, the
    named code is deleted".
  - Once a task's fixtures pass, check it off. Contexts discovered along the way are not new
    sub-items; they are parked.
  - A checked item is one line plus its PR link. "Preserve X, Y, Z" lists belong in the tests
    and the progress log.
- Before adding a per-context file to a layer that already has one per context, consolidate
  first (see 2.2).
- Ratchet baselines after every improvement: `just fixture-update`.
- Host-library categories ([`out-of-scope.md`](docs/decisions/out-of-scope.md)) are excluded from
  every target. Declined work and deliberate divergences are in
  [`known-divergences.md`](docs/decisions/known-divergences.md) — do not reopen without new evidence.
- How to find, reproduce and measure a failure:
  [`testdata/fixtures/README.md` → "Working on a failing fixture"](testdata/fixtures/README.md#working-on-a-failing-fixture).
  Live numbers: `TEST_STATUS.md` (generated) and `just fixture-report --in-scope --classify`.

**Legend:** `[ ]` open · `[x]` done (retained while the subphase is open) · ⏸️ gated, do not start · ⚠️ measure before implementing.
Size: S (hours), M (days), L (week+). Line/fixture counts are sizing hints; regenerate them
rather than trusting them.

**Ordering.** Phases 1–4 are the `*Fail` error-detection suites (the bulk of in-scope failures)
and build on each other: vocabulary first, because most other diagnostic work only passes once
the sentence is right. Phase 5 (execution suites) is independent and can run in parallel.
Phase 6 is gated.

**Measured on 2026-10-09:**
- In scope: 1,423 pass / 372 fail. FailureScripts is at 320/529 (60 %, 157 short of 90 %) and
  BuildScripts at 7/49.
- Of the 268 failing `*Fail` fixtures:
  - 107 still print a go-native sentence (1.8);
  - 96 are at distance 1;
  - only a handful were blocked purely by compile stops (2.1, now +6).

---

## Phase 1 — Diagnostic vocabulary

Replace go-dws's invented diagnostic sentences with DWScript's. This is the precondition for the
majority of failing `*Fail` fixtures; the parser is in it as much as the analyzer. Worklist with
every missing/spurious shape, the fixtures it blocks and the emitting site:
[`fail-shape-worklist-2026-09.md`](docs/architecture/fail-shape-worklist-2026-09.md)
(regenerate with `--shape-top 0 --shape-fixtures`). Anchors must be measured per shape; the
sentence is the easy half.

### 1.3 `Cannot assign "X" to "Y"` — M

This completed checklist is retained as a Phase 1 milestone checkpoint. Details and validation:
[`September progress log`](docs/history/progress-log-2026-09.md) and
[`October progress log`](docs/history/progress-log-2026-10.md).

- [x] For-in assignment diagnostics.
  - [x] Anchor related-class narrowing at `do` (`for_in_subclass`).
  - [x] Report `Cannot assign` for narrowing; retain type pairs for unrelated
    classes (`for_in1`, `for_error4`).
- [x] Scalar/interface assignment anchors and field-backed properties.
  - [x] Use RHS positions for ordinary scalar/interface mismatches (`assign_error`,
    `coalesce_class`, `assign_obj_from_intf`, `interface_inheritence2`).
  - [x] Treat field-backed writers as assignments; preserve method-backed setter
    argument diagnostics, including inherited and implicit `Self` paths.
- [x] Class-to-interface assignment wording (`assign_intf_from_obj`,
  `assign_intf_from_intf`): conversion sentence at `:=`, using the declared source type.
- [x] Multidimensional allocation diagnostics (`multi_dim_dyn_array1`).
  - [x] Check every dimension and retain the nested array type during recovery.
  - [x] Anchor allocation mismatches at `[`, including after whitespace/comments
    and through field/property/indexed assignment paths.
- [x] Array constructors and constant recovery.
  - [x] Preserve constructor types and assignment anchors (`array_assign_error3`).
  - [x] Accept empty constant constructors for typed dynamic arrays without
    weakening element-type invariance or sharing mutable constant storage.
  - [x] Check RHS types after read-only constant-array writes; preserve bounds and
    parser recovery (`array_const`).
- [x] Compound operator diagnostics (`assign_op_incompatible`).
  - [x] Validate both operand types and report `Incompatible operands` at the operator.
  - [x] Retain the additional assignment mismatch where recovery requires it.
  - [x] Preserve RHS assignment conversion for registered operators; stop after
    missing class operators and retain function-name operator dispatch.
  - [x] Convert numeric Variant operands before execution; preserve targets and
    original exceptions when conversion raises, without repeating the converter.
- [x] Routine-pointer assignment recovery (`func_ptr1`).
  - [x] Keep accepted references; read rejected bare routines as calls.
  - [x] Preserve call-result types after arity errors and report valueless RHS assignments.
  - [x] Report pointer-target operand/type errors in the expected child-first order.
  - [x] Invoke named and pointer factories once through assignments and typed
    initializers, including var/lazy parameters; preserve returned references,
    original exceptions and destinations without caching lazy reads.
  - [x] Retain dynamic-array slot diagnostics through implicit array factories
    and preserve compound-call recovery.
  - [x] Preserve exceptions through forwarded lazy expressions and stop coalesce
    fallback after a raised supplier.
- [x] M Enum binding (`enums9`, `enums10`).
  - [x] Keep enumeration mismatch anchors at `:=` (compile-path regression test).
  - [x] Match duplicate-preserving, sorted local lookup by midpoint: `enums9`
    selects `TEnum1.Hello`, while `enums10` selects `TEnum2.Hello`.
  - [x] Register source types, aliases, routines and unit entries in source order;
    include the program's `Default`, `Internal` and `System` entries without
    counting their builtin members or synthetic runtime bindings as local entries.
  - [x] Record immutable enum-constant bindings per expression, including source
    scope snapshots for deferred bodies, so runtime lookup cannot change them.
  - [x] Retain qualified values, element deprecation metadata, case-insensitive
    duplicates and lexical shadowing in the analyzer and evaluator.
- [x] M Cast diagnostics and recovery.
  - [x] Preserve invalid RHS expressions and recover cast result types (`as_error`).
  - [x] Match object/class-reference checks and metaclass cast types (`object_relops`).
  - [x] Preserve AS anchors and compiler stops without spurious later diagnostics.
- [x] M Record metatypes (`record_meta`).
  - [x] Distinguish record type values and aliases from record instances in semantic typing.
  - [x] Preserve record static-member/helper dispatch through inferred metatype values.
  - [x] Canonicalize runtime record aliases and classify metatypes consistently in
    language-type and overload resolution, with deliberate helper receiver typing.
  - [x] Reject instance-to-metatype assignments at the supplying expression.
- [x] M Anonymous routine return types and JSON autoboxing (`JSONConnectorFail/autobox`).
  - [x] Parse nested callable return types with correct spans and recovery.
  - [x] Resolve nested signatures recursively from the AST, including parameter
    modifiers and ownership of nested `of object` suffixes.
  - [x] Implicitly invoke the outer callable once; preserve compatible references
    and avoid invoking a callable returned by that call a second time.
  - [x] Reject routine/class-reference boxing with the expected type names and anchors.
- [x] Final callable integration review.
  - [x] Apply the same factory/reference selection to active unit-qualified
    routines, preserving lexical shadows and returned-type JSON diagnostics.
  - [x] Capture receivers once when member-array indexing falls back from
    indexed-property lookup, including nested arrays and supplier exceptions;
    retain grouped indices for intermediate record default properties.

### Diagnostic progress: former Phase 1.5

The former origin-based backlog is split into 1.5–1.8 below. Eight merged batches
(PRs #449–#456) shipped work that is now checked under its owning subphase. Earlier history entries
keep their original “Phase 1.5” labels; their scope and validation remain the evidence.

| Subphase | Scope | Completed tasks | Remaining tasks |
| --- | --- | ---: | ---: |
| 1.5 | Declarations and intrinsic diagnostics | 4 | 2 |
| 1.6 | Member and helper call diagnostics | 5 | 8 |
| 1.7 | Default arguments and writable storage | 3 | 4 |
| 1.8 | Statements and remaining vocabulary | 0 | 6 |

Counts are top-level checkboxes, not subtasks or estimates of effort. A parent
stays open until all its subtasks pass. Update its checkbox and this table in the
same PR; retain checked siblings until the subphase closes. Detailed evidence:
[`October progress log`](docs/history/progress-log-2026-10.md).

**Current track (after the 2026-10-09 Phase 2 review):**
1. Phase 2.1, the compile-stop model.
2. The 1.8 go-native sentence sweep, which has the largest measured yield.
3. 2.3's fixture-backed parser gaps, in parallel.

Do not extend reintroduced-property contexts (parked in 2.4). Default type/alias checks remain in
4.2. The 1.6 mixed/inherited helper candidate measurement and 1.7 defaults/storage stay open,
but they gained no fixtures in #458–#462, so each needs a named fixture before more work.

### 1.5 Declaration and intrinsic diagnostics — M remaining

Owners: parser declaration recovery, `internal/semantic/analyze_classes*.go`,
default-expression validation, and intrinsic argument checkers.

- [x] `Swap` data-argument and type diagnostics ([#449](https://github.com/CWBudde/go-dws/pull/449)).
  - [x] Validate writable data arguments, canonical type pairs, and one-argument punctuation.
  - [x] Preserve valid field/array storage and reject immutable/temporary storage (`swap1`, `swap2`).
- [x] Immutable assignment diagnostics ([#449](https://github.com/CWBudde/go-dws/pull/449)).
  - [x] Use the left-side assignment sentence for read-only bindings and class constants.
- [x] Property accessors and operator declarations ([#450](https://github.com/CWBudde/go-dws/pull/450)).
  - [x] Match missing accessor names, getter/setter signatures, and measured anchors.
  - [x] Preserve invalid operator recovery and validate global operator bindings.
- [x] Class, record, and helper member headers ([#451](https://github.com/CWBudde/go-dws/pull/451)).
  - [x] Match header/END compiler stops and retain earlier declaration diagnostics.
  - [x] Preserve record field recovery and helper visibility diagnostic ordering.
- [ ] Remaining intrinsic argument diagnostics.
  - [ ] Measure and implement `Assert` and remaining `Inc`/`Dec` sentences and anchors;
    first-argument var validation is already done in 1.7.
  - [ ] Close `Swap` zero/excess-argument and delimiter-stop recovery gaps.
  - [ ] Inventory other intrinsic shapes and split each measured failure into a follow-up.
- [ ] Default-expression declaration validation beyond regular routines.
  - [ ] Validate method/record/helper defaults for constantness, types, and modifiers.
  - [ ] Audit other parameter-declaration sentences; regular-routine `params1` is closed.

### 1.6 Member and helper call diagnostics — M per remaining batch

Owners: `internal/semantic/analyze_function_calls.go`, `analyze_method_calls.go`,
`analyze_helpers.go`, shared member argument checking, and frontend acceptance tests.
Pin complete diagnostic lists through `frontend.Compile`, including recovery types.

- [x] Class-member and constructor calls ([#453](https://github.com/CWBudde/go-dws/pull/453)).
  - [x] Match explicit/implicit Self and named inherited count/overload diagnostics.
  - [x] Check supplied types before count and preserve child errors, anchors, and recovery.
- [x] Parenthesized native record calls ([#454](https://github.com/CWBudde/go-dws/pull/454)).
  - [x] Cover instance/class/metatype receivers, implicit Self, and recursive Result aliases.
  - [x] Preserve receiver-shifted numbering, overload policy, and all supplied child errors.
- [x] Direct parenthesized native interface calls ([#455](https://github.com/CWBudde/go-dws/pull/455)).
  - [x] Match count/type ordering and anchors for inherited and expression receivers.
  - [x] Preserve contextual nil/[] arguments and result types during recovery (21 cases).
- [x] Parenthesized receiver-helper calls ([#456](https://github.com/CWBudde/go-dws/pull/456)).
  - [x] Cover primitive/record/class/interface targets and implicit Self fallback.
  - [x] Preserve selected instance/class/static receiver roles and diagnostic ordering (70 cases).
- [x] Remaining helper call contexts (both subtasks complete).
  - [x] Match explicit helper-name calls and helper-body bound routine calls;
    preserve the existing overload-candidate policy (see the October progress log).
  - [x] Distinguish bare invocation from routine references, including grouped helper callees;
    preserve receiver capture and selected reference declarations (66 compile cases, 13 run cases).
- [ ] Helper overload selection.
  - [x] Measure candidate selection independently from selected-signature argument checking
    ([pinned-source audit](docs/architecture/helper-overload-measurement-2026-10.md)).
  - [x] Honor a lone `overload` declaration and anchor failed matches at the member;
    a found helper owns its failure, without an inaccessible-member fallback.
  - [x] Retain recoverable child diagnostics before outer no-match errors, declaration-case
    hints, and independent receiver eligibility errors (63 compile cases, 4 run controls).
  - [x] Keep checked runtime execution on the selected declaration through receiver,
    explicit-name, and helper-body calls; unchecked dispatch ranks signatures instead of arity
    ([#462](https://github.com/CWBudde/go-dws/pull/462)).
  - [x] Prepare arguments using the selected declaration's lazy/var flags, preserving
    static Variant selection and once-only receiver/ordinary-argument evaluation.
  - [x] Preserve empty child/grandchild helpers' inherited overload sets and declaring
    storage owner; retain local-name hiding for the parent-helper extension.
  - [x] Capture lexical helper-body and function-helper declaration identity, including
    shared helper labels and strict aliases; retain first-helper precedence within a target.
  - [ ] Pin child analysis and receiver ownership across mixed/inherited overload sets;
    distinguish target-class inheritance from the port's parent-helper syntax.
  - [ ] Match measured ambiguity/tie and Variant conversion ranking; preserve contextual
    routine arguments when multiple candidate signatures provide a common expectation.
- [ ] Helper availability and target lookup.
  - [ ] Audit imported-helper availability in the receiving analyzer across a unit boundary.
  - [ ] Preserve metaclass helper targets during parenthesized receiver lookup and unwrapping.
- [ ] Bare record/interface member contexts.
  - [ ] Distinguish implicit invocation from a compatible routine reference.
  - [ ] Measure grouped member callees, where the member is read before the outer call.
- [ ] Noncallable record-name shadow fallback.
  - [ ] Pin lexical noncallable shadows separately from recursive function Result aliases.
  - [ ] Correct the fallback without changing the shipped Result-alias behavior.
- [ ] Multiline child/type diagnostic ordering.
  - [ ] Measure argument-child diagnostics versus supplied-type errors across lines.
  - [ ] Preserve the expected ordering alongside count/overload recovery.
- [ ] Excess bare-callable arguments.
  - [ ] Analyze implicit calls beyond the member's declared parameter list.
  - [ ] Retain the excess callable's own arity diagnostics before the outer call error.
- [ ] Inherited-call arity, recovery and class-method execution.
  - [ ] Measure unnamed inherited calls independently from shipped named calls.
  - [ ] Repair named inherited class-method execution: a method body inherited by a
    grandchild currently calls the dynamic parent's method instead of the lexical
    method owner's parent (discovered during the 2.1 inherited-property batch).
  - [ ] Match arity, anchors, and result recovery for each measured form.
- [ ] Synthetic parameterless constructor ambiguity.
  - [ ] Measure competition with inherited constructors whose parameters all have defaults.
  - [ ] Resolve ambiguity while preserving upstream constructor candidate policy.
- [ ] Distance-1 member sentences (2026-10-10).
  - [ ] Member vs. method sentence (`class_var_scope1`, `enums_alias`), `Unknown name
    "TMyEnum.meBug"` without a stop (`enums`), `No parameters expected`
    (`string_builtin_methods1`), `Class method or constructor expected` for an instance field
    read by a class property (`class_property3`).
  - [ ] Inherited lookup through the implicit `TObject` parent: `Method "X" not found in
    ancestor class` (`inherited4`, `inherited6`).

### 1.7 Default arguments and writable storage — M per remaining batch

Owners: semantic signatures, helper/interface declarations, and runtime call/storage
handling in `internal/interp/evaluator`. Runtime defaults must use declaration scope.
Declaration-validation prerequisites live in 1.5; call diagnostic context lives in 1.6.

- [x] Regular-routine default declarations ([#449](https://github.com/CWBudde/go-dws/pull/449)).
  - [x] Validate default modifiers, constantness, type compatibility, and anchors (`params1`).
  - [x] Preserve valid aggregate/pure-call defaults and required rejected parameters.
- [x] Native record omitted-default execution ([#454](https://github.com/CWBudde/go-dws/pull/454)).
  - [x] Retain defaults through inline/out-of-line and overloaded record signatures.
  - [x] Resolve lexical/record constants at declaration scope and evaluate supplied arguments once.
- [x] Regular var arguments and Self recovery ([#452](https://github.com/CWBudde/go-dws/pull/452)).
  - [x] Close `passing_const_var`, `passing_const_var2`, `const_param2`, and `self_not_writable`.
  - [x] Validate `Inc`/`Dec` argument zero and retain mutable field/array/property/JSON storage.
- [ ] Helper default signatures and execution.
  - [ ] Preserve declaration defaults in receiver-call signatures.
  - [ ] Execute omitted defaults for ordinary, function, and inherited helpers.
- [ ] Interface default signatures and execution.
  - [ ] Preserve defaults in interface signatures, including inherited signatures.
  - [ ] Execute omitted arguments using the interface declaration's defaults.
- [ ] Var checks on remaining call paths.
  - [ ] Extend checks to method/inherited/constructor calls and measure receiver-shifted anchors.
  - [ ] Extend checks to helper/function-pointer calls and retain read-only signature bindings.
- [ ] Writable callable-reference temporaries.
  - [ ] Measure upstream writable routine-reference data and required runtime slots.
  - [ ] Support temporary reference storage before changing the existing `@Routine` rejection.

### 1.8 Statements and remaining diagnostic vocabulary — S per measured shape

Work each batch largest shape first, mapping invented sentences to DWScript's
(`expected ')' after parameter list` → `")" expected`, `unknown type 'X'` → `Type expected`).
Owners: `analyze_statements.go`, remaining class semantic sites, frontend, lexer,
and shared `internal/errors` builders. Remeasure the shape worklist before each sweep.

- [ ] ⏸️ Remaining parser shapes — after the relevant Phase 2 prerequisites.
  - [ ] Audit remaining punctuation and type sentences; coordinate overlapping Phase 2 work.
  - [ ] Distance 1: `Unexpected "else"` (`else_unexpected2`), `Unexpected END`
    (`end_implementation1`), `End of block expected` at EOF (`for_empty`), `String expected`
    (`external2`), lexer number sentences (`binary_literal2`, `for_var_error4`).
- [ ] Unknown statement types and failed inference.
  - [ ] Match `Type expected` for unknown types.
  - [x] Reject valueless variable initializers with the ordinary no-return diagnostic;
    preserve inferred scanner lookahead, typed declaration-name anchors and declaration
    recovery (2.1 prerequisite; [#466](https://github.com/CWBudde/go-dws/pull/466)).
  - [ ] Match `Type could not be inferenced` for failed inference.
- [ ] FOR loop-variable stops from upstream `ReadForTo`: `For loop control variable must be
  simple local variable`, `Integer expected` and `Variable expected` at the `:=`; the stop
  also cuts the `Empty FOR loop` hint (`for_var_usage3`, `for_var_usage4`, `for_loopvar1`; 1 each).
- [ ] FOR STEP diagnostics.
  - [ ] Establish constant-folding prerequisites for `for_step` and its optimized sibling.
  - [ ] Match type/positivity sentences and anchors after folding is available.
- [ ] Break/Continue/Exit diagnostics.
  - [ ] Match `break_continue` and `break_in_finally` with finally nesting explicit.
  - [ ] Match `exit_result6` while preserving compiler-stop behavior.
- [ ] Remaining class declaration vocabulary — coordinate with Phase 3.
  - [ ] Map emitting sites in `analyze_classes*.go` to the Phase 3 declaration/forward tasks.
  - [ ] Close measured leftover sentence/anchor gaps without duplicating Phase 3 ownership.
- [ ] Other semantic, frontend, lexer, and shared error-builder sites.
  - [ ] Close incompatible-type pairs in `coalesce`, `in_typecheck1`, and `property_default1`;
    the named former task 1.2 fixtures are closed.
  - [ ] `case_typecheck`: drop the spurious Integer-vs-Float label error that upstream accepts.
  - [ ] Anchor-only distance-1 gaps, message already right: `property_error7`,
    `record_recursive2`, `resourcestring1`/`3`, `loop_infinite`,
    `InnerClassesFail/sub_outside_scope`.
  - [ ] Inventory remaining invented sentences and split each measured shape into a task.
- [ ] Go-native sentence sweep, largest shape first. 107 failing `*Fail` fixtures still print a
  go-native sentence; in 66 of them it is the only difference.
  - Top shapes (2026-10-09):
    - `implementation signature … does not match forward declaration` (10);
    - `method X not declared in class` (7);
    - `unary - requires numeric operand` (6);
    - `circular inheritance detected` (4).
  - Regenerate the list with `just fixture-report --in-scope --classify --shape-fixtures --shape-top 0`.
  - Distance 1 (2026-10-10): `method2` (`Unknown name "Result"` at the LHS), `exit_result2`,
    `inherited1`, `new_class2`/`5`, `new_array` (`unknown type`), `not_untyped`,
    `method_implem2`, `class_loop`, `class_operator1`/`2`, `read_self`, `class_const1` and
    `helper_error2` (each an extra go-native line).

---

## Phase 2 — Frontend structure: compile stops and property resolution

**Why this phase was restructured (2026-10-09 review).** The former 2.1 started as four
fixture-backed items (`const_record1`, `special_funcs1`/`at_integer`, `debugbreak`,
`property_reintroduce1`/`2`). All of them pass, along with `const_record4`, `special_funcs5`,
`SimpleScripts/property_reintroduce`, `inherited1` and `inherited5`. Even so, the subphase grew
from 4 open items to 20, because each PR added context sub-items that no fixture asks for. Nine
PRs (#463–#471) gained 9 fixtures in total, and five of them gained none. A review from four
angles (architecture, fixture leverage, history, docs) found three structural causes:

1. **One feature, one code path per AST shape.** Upstream handles compatibility brackets in a
   single function, `ReadPropertyExpr` (`dwsCompiler.pas`), which consumes `()` only after it
   knows the symbol is a property. That one function serves five callers: inherited,
   unqualified name, property-program, member and helper access.
   - In go-dws the parser fixes the node shape first. Every shape then needs its own pieces:
     five analyzer entry points (`property_reintroduce.go`, `property_indexed_compat.go`,
     `property_implicit_indexed.go`), four `SemanticInfo` maps (`pkg/ast/metadata.go`) and
     three evaluator executors.
   - So every new context costs a PR across four layers.
2. **Compile stops are spread over five mechanisms:**
   - the parser's `stopped()` checks;
   - three different ways of handling truncated calls (`nil` / `Incomplete` / `DeferredCall`);
   - the analyzer's `compileStopped` flag, which only skips end-of-program checks;
   - a frontend chain that matches message strings, cuts by position and re-sorts
     (`internal/frontend/result.go`);
   - the lexer cutoff.

   Upstream raises `ECompileError` and unwinds. Hard cases were deferred to "2.4's carrier",
   which was never built.
3. **No definition of done.** The "or a test through the real path" rule let contexts with no
   fixture behind them become closable tasks. See the definition of done under **Rules**.

**Ordering.** 2.1 is the only part with direct fixture yield and the best leverage on the
remaining distance-1 `*Fail` fixtures, so do it first. Do 2.2 before any further property-access
work. 2.3 is independent: its items are S-sized one-offs that can run in parallel. 2.4 is not
work, it is the parking list.

### 2.1 Compile-stop model — M

Replace the separate stop mechanisms with one model: the parser marks a truncated call, the
analyzer genuinely stops, and the frontend makes one position cut. Design and current state:
[`compile-stops.md`](docs/architecture/compile-stops.md).

Measured yield so far: +9 fixtures.
- `missing_parenthesis1`, `block_unfinished2`, `except_error4`, `except_error5`, `enums8`,
  `params3` and `class_cast` in FailureScripts.
- `HelpersFail/strict`.
- `BuildScripts/init_order4`, which is order-dependent, so it is not ratcheted.

`static_methods` is not a target: its only extra line is the `{$FATAL}` stop itself (see
[known divergences](docs/decisions/known-divergences.md#fixtures-that-cannot-pass-as-written)).

- [x] S Write `docs/architecture/compile-stops.md`: upstream model, current mechanisms, target
  design, and the §1.3 invariants formerly listed under 2.4.
- [x] M Parser: one `Truncated` carrier on every call form, `inherited` included; the analyzer
  analyzes only completed arguments (`missing_parenthesis1`).
- [x] M Analyzer: real stops via `addCompilerStop` (`internal/semantic/compile_stop.go`),
  mapped from upstream's `AddCompilerStop` sites (`HelpersFail/strict`, `except_error4`/`5`).
- [x] S Frontend: delete `type_punctuation.go` and the deferred-call refine pass; the analyzer
  decides boundary-call stops; apply the stop cut to units.
- [x] S Drop completion hints of blocks cut by a missing END (`block_unfinished2`).
- [x] S Re-run `--classify` (2026-10-10: 107 failures at distance 1) and move the
  first-diagnostic gaps the stop cut isolated into Phases 1, 2.3, 3 and 4.
- [x] S Carrier: resolve a truncated call's member before its argument stop (`enums8`), keep
  its casing hint (`params3`); `;` after a positional argument stops with `")" expected`
  (`class_cast`).
- [ ] S Carrier: the truncated `try` drops the handler check before `END expected`
  (`try_except1`; the check itself is also go-native and rejects `TAlias = Exception`), and the
  truncated class header drops the interface check before `")" expected` (`class_error4`).
- [x] S Finish the single cut in the frontend: the earliest stop of either phase cuts every
  later diagnostic, early-emitted ones included; the class-body filter and the first-stop
  `break` are deleted. A stop raised after reading past its display position carries
  `SemanticError.Cursor` (`array_of_proc2`).
- [x] S Measure the former 2.1 tails (`Default.Low(;`, casing hints in `PrintLn(debugbreak(;`,
  reintroduced-property bracket recovery): no fixture needs them, parked in
  [known divergences](docs/decisions/known-divergences.md).

### 2.2 Property-access consolidation — M (refactor, no new behaviour)

Prerequisite for any further property-access work, including the parked contexts in 2.4 and
#472's open property-index children. **No new per-context file in `internal/semantic` or
`internal/interp/evaluator` until this lands.**

Acceptance for both tasks: every fixture and existing test is unchanged, and the listed
per-context code is deleted.

- [ ] M One analyzer property resolver: a direct port of `ReadPropertyExpr`, for example
  `resolveProperty(receiver, prop, ctx, postfix)`.
  - It works on a normalised postfix view: an optional paren group with positions, the bracket
    groups, and read vs. write.
  - Thin adapters at the five shape entry points feed it: method call, call, inherited, and the
    two index-over-call shapes.
  - The duplicated sequence (case hint, `Not a method` stop, compatibility hint, write-only
    check, object-reference check) lives only in the resolver.
- [ ] M One resolved-member binding: `SemanticInfo` records `{Kind, Descriptor, AccessorOwner,
  Receiver, Indices, SelectedDecl}` per node.
  - It replaces the four per-shape maps.
  - One evaluator executor consumes it. Delete `implicit_property_read.go`,
    `inherited_property_read.go`, `indexed_property_compat.go` and the
    `"__helper_receiver:"+T` type-name channel.
  - The runtime-kind lookup remains only as the unchecked-mode fallback.
  - Order: property reads, then writes. Overload selection (5.3) and record member shadows
    (5.2 `Ord`) reuse the same binding afterwards.

### 2.3 Fixture-backed parser gaps — S each

Each item names its fixture; the measured distance is in parentheses. Remeasure first and work
nearest-first.

**Open:**
- [ ] `var`/`const` in property index parameters (`array_params1` 4, `array_params2` 8,
  `Parameters expected`). [#472](https://github.com/CWBudde/go-dws/pull/472) is in progress on
  this. Its non-fixture children follow the definition of done.
- [ ] `Invalid Operands` for a bare-special operand (`special_funcs2`, 1).
- [ ] `array of const` (`open_array`, 2).
- [ ] `Dot "." expected` where the parser must know `TTest` is a class (`method_implem6`, 2).
- [ ] `unexpected "@"` in the remaining contexts (`dyn_array3`, `field_init1`, `func_ptr6`, 2
  each). `SetOfFail/invalid_operand` already passes.
- [ ] `Field has already been set` (`const_record2`, 3) and `Constant expression expected`
  (`const_record3`, 3) in record constants.
- [ ] Attribute `"]"` anchored at the `[`, plus `Dangling attribute declaration`
  (`attribute_incorrect2`, 3).
- [ ] `interface helper for T` (`HelpersFail/mixed_helper`, 3; together with 4.3).
- [ ] `OF OBJECT expected` and legacy hints (`legacy_proc_of_object`, 9).
- [ ] An empty `()` group is `Expression expected`, not an empty array literal
  (`member_of_void1`, 1). Recheck after #472's `null_read_expression`, which has the same root.

**Landing in [#472](https://github.com/CWBudde/go-dws/pull/472)** (remove after merge):
- `export` directive;
- `property_description1`;
- `missing_reader_bracket`;
- `null_read_expression`/`null_write_expression`;
- `include_incorrect`;
- `conditionals2.1`.

**Owned elsewhere:**
- `special_funcs3` (`Name "X" is reserved`) is in 3.2.
- `special_funcs4` (`Inc(i, )`) is in 4.2.

### 2.4 Parked: no fixture demands it

These are moved to [`known-divergences.md`](docs/decisions/known-divergences.md#behaviour-without-a-fixture-that-demands-it).
Each reopens only with new evidence (a fixture, a user report, or an upstream test), and only
after 2.2:

- Reintroduced-property compatibility-bracket contexts beyond what has shipped:
  explicit, unqualified and inherited scalar reads, and named and unqualified indexed reads.
  Named inherited indexed reads are landing in
  [#472](https://github.com/CWBudde/go-dws/pull/472); they stay parked if #472 drops them.
  Parked:
  - bare `Prop[i]`;
  - `Obj()[i]` and other default-property postfixes;
  - `index` directive combined with index parameters;
  - `F.Prop()[i]` function-pointer receivers;
  - writes;
  - function-valued properties;
  - unchecked execution;
  - visibility promotions and generic specialisations;
  - helper-property precedence and deprecation ordering;
  - private-access/static/write-only call-ordering alignment.
- Qualified builtin calls through `Internal`/`System` (`Internal.Abs(-3)`) and method-call
  syntax on those qualifiers.
- The follow-up diagnostic upstream reports inside `begin…end` for a value left unconsumed after
  a read-only property assignment. No fixture measures it:
  `expr_write_readonly_property` fails for an unrelated reason, `write (expr)` parsing.

---

## Phase 3 — Declaration and class diagnostics

Still the largest single family (FailureScripts, OverloadsFail, InterfacesFail). One subtask per
message shape; counts are lines (fixtures).

### 3.1 Method implementation tracking — M

`Method "X" of class "Y" not implemented` 34 (15).

- [ ] S ⚠️ Key forward tracking per overload symbol, not per name, for classes
  (`ClassType.ForwardedMethods`, `analyze_classes_decl.go`) and helpers together. Upstream's
  `TStructuredTypeSymbol.CheckMethodsImplemented` (dwsSymbols.pas) tests each method symbol's own
  executable, so an unimplemented overload is still reported, sorted by declaration position.
  Measure the per-overload wording and anchor first.
- [ ] S Remaining fixtures of the shape, including `new_class3` (the unimplemented `default`
  constructor is never reported) and `constructor_no_name` (recovery must keep the nameless
  constructor).

### 3.2 Declaration shapes — S each

- [ ] `Name "X" already exists` 20 (12).
- [ ] `Name "X" is reserved` for special-function declaration names, including
  DebugBreak variables, parameters, and global routines; preserve allowed dotted methods.
- [ ] `Class reference expected` 11 (9); distance 1: `new_class4`, `class_operator5` (comparing
  a class reference with an object).
- [ ] Interface-list checks: `"X" is not an interface` and `Interface "X" already implemented`
  (`class_error4`, `class_error5`).
- [ ] `Previous declaration of class was not "partial"` (`partial_class2`), and the type-end stop
  before `already declared` (`unit_prefix4`).
- [ ] `Class "X" isn't defined completely` 9 (7), and the `Interface` variant 4 (3).
- [ ] `There is already a field with name "X"` 8 (4).
- [ ] `"X" is not a method of class "Y"` 7 (4).

### 3.3 Overload and forward rules — M

- [ ] M Overloads differing only in return type are ambiguous (`overload_simple`).
- [ ] M Method overload hiding, visibility, `no overloaded version declared`
  (`meth_overload_simple`, `meth_overload_hide`, `meth_private_public`, `overload_missing`).
- [ ] S Forward/implementation mismatch wording (`Declaration should be…`,
  `Value-parameter expected`, default-value mismatch): `declaration_mismatch1`/`2`,
  `default_params2` (needs a constant-folded comparison of two default-value expressions in
  `mergeDefaultValues`).
- [ ] S Analyze a routine's body even when its declaration fails the overload check
  (`forwards_unit`, `IntToHex` error on line 23), and drop its spurious
  `Unit name does not match file name` warning.
- [ ] S `@IntToStr` is ambiguous (`func_ptr_constant_ambiguous`), and an inline method body
  must not see a field declared after it (`method_implem`).
- [ ] S `Preconditions must be defined in the root method only` (`contracts_precondition`).
  Until then the runtime evaluates every `require` in the chain, root-most first.

---

## Phase 4 — Missing validations

DWScript reports something; go-dws compiles clean or says something else. Mostly
single-fixture work; the per-suite list is in
[`fail-suite-audit-2026-09.md`](docs/architecture/fail-suite-audit-2026-09.md).

### 4.1 Array diagnostics — M

- [ ] M Separate `const` `array of Variant` from `array of const`; the analyzer currently tells
  them apart by declared type name.

### 4.2 Builtin-call checks — S each

- [ ] `internal_unsupported`: `Length`/`Low`/`High` want `Invalid argument type` (reuse the
  `Assigned` check); `Inc` wants `Integer expected`.
- [ ] `special_funcs4`, `default_params4`: `Expression expected` for an empty trailing argument
  (`Inc(i, )`, `Test(1, )`).
- [ ] `assign_untyped`: `Assignment's right-side-argument has no return type`, and
  `Cannot assign a value to the left-side argument` when assigning to a procedure name.
- [ ] `assert`, `enum_byname`: `Boolean expected` anchored at the argument, not the call;
  `enum_byname` also wants `String expected`.
- [ ] `lazy_func_ptr`: `Lazy parameter cannot be a function pointer`.
- [ ] `contracts_error2`: builtins must resolve inside a `require` clause.
- [ ] Measure parenthesized `SizeOf` forms and supply the missing call support;
  `SizeOf(Integer)` currently reports unknown name, while bare punctuation is closed in the former 2.1.
- [ ] Validate Default type/alias, case-insensitive type-name, and value-argument forms;
  aliases/lowercase types currently fail, and `Default(I)` returns nil for an Integer value.

### 4.3 HelpersFail — S

- [ ] Record the `for` keyword's position on `ast.HelperDecl` (`internal/parser/helpers.go`); all
  six anchors of `mixed_helper` and `helper_of_delegate` are the `for` token.
- [ ] Add an `IsInterfaceHelper` flag and the helper-kind checks (`mixed_helper`; needs 2.3's
  `interface helper for T`).

### 4.4 Other FailureScripts validations — S each

- [ ] `class_const4`: `Constant Instruction - has no effect` as an **error** on a class-const
  declaration.
- [ ] `enum_flags_overflow`: per-element position on `ast.EnumValue`.
- [ ] `case_error6`: a void case selector gives `Expression expected` at `of` and skips the
  label checks (upstream `ReadCase`).
- [ ] Calling a type name: `Function expected` for a procedural type (`func_ptr2`,
  `invalid_cast3`; upstream `CreateSimpleFuncExpr`), `Expression expected` for an empty cast
  (`invalid_cast`, `invalid_cast2`).
- [ ] `const_3`: `Invalid const type "procedure T"` at the type token.
- [ ] `property_write5`: a write-only property followed by `.` or `[` stops with `Constant
  expression expected` (upstream `ReadPropertyExpr`). After 2.2.
- [ ] `contracts_old` (`Function or value expected`), `func_ptr_local` (local routine as a
  delegate), `enum_scoped2` (`Flags enumerations cannot have user values`).
- [ ] Remaining silent FailureScripts fixtures from the audit; distance 1: `class_var_dyn1`/`2`,
  `external3`, `sealed`, `string_set`, `method1`, `method_implem7`.

### 4.5 Per-suite sweeps — M

Take each suite to zero after Phases 1–3 have landed, one-line near misses first
(`just fixture-report --in-scope --classify --category <Cat> --list-fails`).

- [ ] InterfacesFail, including explicit class-to-interface implementation checks
  (`interface_inheritence1`: implementing a derived interface does not implicitly
  declare its base interface on the class).
- [ ] OverloadsFail.
- [ ] GenericsFail — includes `implem_mismatch1`'s `T expected but u found` for a mismatched
  out-of-line type-parameter name (substitution is positional today).
- [ ] PropertyExpressionsFail.
- [ ] JSONConnectorFail.
- [ ] LambdaFail.
- [ ] OperatorOverloadFail.
- [ ] AssociativeFail, AttributesFail, InnerClassesFail.

---

## Phase 5 — Execution suites

In-scope failures in the suites that run a program (SimpleScripts, BuildScripts, ArrayPass, a few
elsewhere). All previously triaged groups are closed; what remains is untriaged.

### 5.1 Triage — S

- [ ] Run `just fixture-report --in-scope --classify --list-fails` over the execution suites,
  record the first blocker per fixture, and group them into tasks below (replace this list).

### 5.2 Fix groups — sized after 5.1

- [ ] BuildScripts drivers.
  - [ ] Make `init_order2` initialization/finalization ordering deterministic;
    its varying pass must not raise the stable BuildScripts baseline.
- [ ] SimpleScripts.
- [ ] ArrayPass.
- [ ] Remaining categories.
  - [ ] Allow direct calls through function-pointer record class variables and
    properties (`R.Stored()` / `R.Factory()`); capturing the pointer into a
    compatible variable already works.
  - [ ] Preserve record member shadows of builtin names during bare-body execution;
    `Ord` inside another record method compiles but invokes the ordinary builtin.

### 5.3 Expected-type overload resolution — M

- [ ] Return type is not part of overload identity (`types.SignaturesEqual`), so an expected type
  can only be a last-resort tie-break. Resolving it in the analyzer alone would let the evaluator
  pick a different overload at run time (`evaluator.ResolveOverloadMultiple` has no expected-type
  channel). Either pass the call site's expected type to the evaluator or have it reuse the
  analyzer's choice via `ast.SemanticInfo`. The note sits at the dispatch in
  `internal/semantic/analyze_expressions.go`.

---

## Phase 6 — Gated

Gate: **every in-scope fixture category ≥ 80 %** (harness and CLI). Do not start before.

- ⏸️ **Bytecode VM decision.** Kept in tree, unmaintained, opt-in and labeled experimental
  (`run --bytecode`, `dwscript compile`, `pkg/dwscript.CompileModeBytecode`). Decide delete vs.
  rebuild; a rebuild must use `internal/builtins` and `internal/interp/runtime` values, not the
  current fork. See [`bytecode-vm.md`](docs/decisions/bytecode-vm.md).
- ⏸️ **Host-registered external classes** for the six `Memory/external*` fixtures: upstream's
  `SetUp` registers `TExposedClass` and `TExposedBoomClass` with host-side constructors and an
  `OnCleanUp` hook (`UMemoryTests.pas`).
- ⏸️ Host-library bindings (DB, Crypto, COM, Graphics, Web, Tabular, TimeSeries, DOM, Linq):
  [`out-of-scope.md`](docs/decisions/out-of-scope.md).
- ⏸️ Sandbox / capabilities model (`AllowFileRead`, `AllowHTTP`, memory/time limits): design in
  [`out-of-scope.md`](docs/decisions/out-of-scope.md).
- ⏸️ Compiler backends (Go/AOT, JavaScript, LLVM, WebAssembly AOT): backlog in
  `docs/archive/CodeGenTODO.md`, `docs/archive/CodeGenJSGoal.md`.
- ⏸️ AST-driven formatter: [`formatter-style-guide.md`](docs/decisions/formatter-style-guide.md),
  [`formatter-ast-audit.md`](docs/decisions/formatter-ast-audit.md).

---

## Reproducing the numbers

```bash
just fixture-report                                      # CLI ground truth (rebuilds bin/dwscript first)
just fixture-report --in-scope --classify --list-fails   # distance-to-passing per failure
just fixture-report --category SimpleScripts --list-fails
go test ./internal/interp -run TestDWScriptFixtures -v   # Go harness (what CI gates on)
FIXTURE_LIST_FAILS=1 go test ./internal/interp -run TestDWScriptFixtures -v
just fixture-update                                      # regenerate TEST_STATUS.md, ratchet baselines
```
