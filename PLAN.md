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

**Current user-selected track:** 2.1's type-directed punctuation, after the record-constant,
bare Low/High/address-of, `DebugBreak`, and remaining bare special-function batches.
Default namespace availability/binding is implemented in pending [#466](https://github.com/CWBudde/go-dws/pull/466).
Explicit scalar reintroduced-property empty reads and statement-boundary recovery shipped
in [#467](https://github.com/CWBudde/go-dws/pull/467). Checked unqualified empty scalar reads
in class methods shipped in [#468](https://github.com/CWBudde/go-dws/pull/468). This batch adds named inherited empty/bare scalar reads and parent accessor ownership.
Next: reintroduced-property nonempty/discarded-parent recovery and indexed/write contexts. Default type/alias checks remain in 4.2; nested stopped-call intrinsic hints
need 2.4's full truncation model.
The earlier 1.6 mixed/inherited helper candidate measurement remains
open, followed by 1.7 defaults/storage. Phase 2/3 prerequisites continue to gate the
overlapping 1.8 sweeps.

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
- [ ] Unknown statement types and failed inference.
  - [ ] Match `Type expected` for unknown types.
  - [ ] Match `Type could not be inferenced` for failed inference.
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
  - [ ] Inventory remaining invented sentences and split each measured shape into a task.

---

## Phase 2 — Parser recovery and compile stops

DWScript's parser sentences and anchors are in place; what is left needs knowledge the parser
does not have, or a precise model of where compilation stops.

### 2.1 Type-directed punctuation — S

An ordinary call says `Expression expected` for `f(;`, so record/type-directed
punctuation must follow semantic resolution. Reserved magic names can enforce their
punctuation in the parser before reading arguments.

- [x] Record-typed const (`const_record1`, [#463](https://github.com/CWBudde/go-dws/pull/463)).
  - [x] Require `(` after resolving record types and aliases, at the first initializer token.
  - [x] Preserve scalar/ordinary-call recovery, earlier diagnostics, local constants,
    and interrupted parenthesized record initializers (`const_record4`).
- [ ] Special functions written without parentheses.
  - [x] Require `(` for bare builtin Low/High (`special_funcs1`, [#463](https://github.com/CWBudde/go-dws/pull/463)), after casing hints;
    preserve lexical/member shadows and lookahead positions through comments/newlines/EOF.
  - [x] Match `at_integer` ([#463](https://github.com/CWBudde/go-dws/pull/463)): report the type's missing `(` before `unexpected "@"`;
    reject scalar-variable addresses while preserving valid routine references.
  - [x] Measure the 17 required-parentheses names, Default's ordinary-name fallback,
    and qualified member/unit lookup separately from unqualified specials ([#465](https://github.com/CWBudde/go-dws/pull/465)).
  - [x] Require `(` for the remaining bare special names in value/statement/callback,
    grouped-callee, range-bound and unresolved assignment-target contexts; retain
    casing hints, scanner anchors, ordinary references and lexical/member shadows
    ([#465](https://github.com/CWBudde/go-dws/pull/465)).
  - [x] Read bare/grouped/indexed special address operands before pointer capture or
    index analysis; preserve child stops and scalar-context pre-child address errors
    ([#465](https://github.com/CWBudde/go-dws/pull/465)).
  - [ ] Establish actual qualified namespace symbol availability and callable forms;
    the special pseudo-symbol table alone does not prove `Default.Low` exists upstream.
  - [ ] Match qualified Default namespace binding through compilation and execution;
    a global Low shadow currently changes execution of `Default.Low(Integer)`.
- [ ] Magic functions (`debugbreak`).
  - [x] Accept bare/empty-parentheses DebugBreak and stop at the first non-`)` token,
    before reading arguments (`debugbreak`); preserve direct-call casing hints and
    comment/newline/EOF anchors ([#464](https://github.com/CWBudde/go-dws/pull/464)).
  - [x] Execute valid calls as a no-op in the evaluator, including checked/unchecked
    procedure and loop bodies; preserve ordinary qualified member calls
    ([#464](https://github.com/CWBudde/go-dws/pull/464)).
  - [x] Reject value/grouped/address use; retain null-constant recovery and later
    name resolution through direct, unary, and binary constant initializers
    ([#464](https://github.com/CWBudde/go-dws/pull/464)).
  - [ ] Preserve intrinsic casing hints inside discarded stopped parent calls
    (`PrintLn(debugbreak(;`); requires 2.4's per-call truncation carrier.
- [ ] Reintroduced properties.
  - [x] Accept the declaration marker before `;` without requiring an ancestor method;
    read explicit scalar class properties through `()` with a normal-level compatibility
    hint and pedantic declaration-case hint. Checked execution preserves getter and receiver
    side effects, aliases/implicit routine receivers, inherited descriptors, class properties,
    helper precedence, descendant method shadows, and source-printer round trips
    ([#467](https://github.com/CWBudde/go-dws/pull/467)).
  - [x] Recover at `;` after the opening `(` with `")" expected` while retaining the
    initializer's property type and later semantic diagnostics (`property_reintroduce2`).
    Stop ordinary scalar property calls at `(` with `Not a method` (`property_reintroduce1`);
    preserve ordinary method/helper/namespace punctuation, parser-only errors and nested-call
    stop anchors, and skip incomplete argument checks ([#467](https://github.com/CWBudde/go-dws/pull/467)).
  - [ ] Preserve upstream token consumption and recovery for nonempty compatibility brackets,
    EOF/end boundaries, lexer-directive reach after recovery and discarded enclosing calls; coordinate with 2.4's full truncation carrier.
  - [ ] Extend compatibility brackets to the remaining access contexts.
    - [x] Read unqualified empty scalar properties inside instance/class methods; preserve
      normal/pedantic hints, local callable and parameter shadows, class-member precedence
      over global routines, ancestor descriptors, descendant method shadows, helper precedence,
      virtual getters, once-only getter effects, original exceptions, nested accessor reads,
      selected descriptor/storage ownership and flagged static-method stops ([#468](https://github.com/CWBudde/go-dws/pull/468)).
    - [x] Support named inherited empty/bare scalar reads (this batch; PR link added on publication).
      - [x] Preserve the parent descriptor and resolved accessor/storage identity through
        forwarding, field/class-variable/constant shadows and ancestor class getters.
      - [x] Retain lexical expression-getter storage, class getters backing ordinary
        properties, nonvirtual nil receivers, original virtual chains and intermediate
        overrides; preserve index directives, getter effects/exceptions and nested reads.
      - [x] Match inherited hint levels/casing, deprecation ordering and write-only
        anchors; preserve empty and argument-bearing inherited calls in printed source.
        Close `SimpleScripts/inherited1` and `FailureScripts/inherited5`.
    - [ ] Support indexed/default property reads.
    - [ ] Support property writes through compatibility brackets.
    - [ ] Support function-valued properties and distinguish property reads from invocation.
    - [ ] Support unchecked execution and malformed/nonempty unqualified calls.
    - [ ] Measure visibility promotions/generic specializations separately: upstream does
      not copy the flag there.
    - [ ] Measure helper-property precedence and deprecated-property warning ordering/anchors
      on unqualified compatibility calls; helper-method controls are covered above.
  - [ ] Align private-property access validation and ordinary static/write-only property-call
    diagnostic ordering with upstream. The private-access gap predates this batch; derive
    the invalid-call sequence before changing scalar postfix checks.

### 2.2 Parser gaps — S each

- [ ] `var`/`const` in property index parameters (`array_params1`/`2`, `Parameters expected`).
- [ ] The `export` directive.
- [ ] `OF OBJECT expected` (`legacy_proc_of_object`).
- [ ] `array of const` (`open_array`).
- [ ] `String expected` for a property description (`property_description1`).
- [ ] Attribute `"]"` anchored at the `[` (`attribute_incorrect2`; also needs
  `Dangling attribute declaration`).
- [ ] `Dot "." expected` where the parser must know `TTest` is a class (`method_implem6`).
- [ ] `interface helper for T` (needed by `HelpersFail/mixed_helper`, see 4.3).
- [ ] SetOfFail parser parity: `bracket_right_missing`, `for_in_set_missing_do`, `of_missing`
  (`"X" expected` / `OF expected` / `DO expected`).
- [ ] Extend `unexpected "@"` beyond the type/scalar cases shipped in 2.1
  (`SetOfFail/invalid_operand`, `dyn_array3`, `field_init1`, `func_ptr6`).

### 2.3 Property accessor recovery — S

- [ ] `read (…)` / `write (…)`: upstream reports every missing `")"` as an ordinary error
  (`missing_reader_bracket` lists lines 4, 6, 7, 8); go-dws's parenthesised-expression stop hides
  all but the first.
- [ ] `Warning: Property writer does nothing` (blocks the above).

### 2.4 Compile-stop model — M

The §1.3 enum lookup refactor must preserve source declaration order and routine-body
diagnostic insertion points while snapshotting constant bindings. Its symbol-table changes
must retain the existing compile-stop and forward-check behavior below.

The §1.3 cast recovery work must restore speculative parser errors and stops before
expression fallback, retain semantic recovery types, and keep true compiler stops
ahead of later assignment or end-of-program diagnostics.

The §1.3 record-metatype changes must keep record annotations and constants as
instance types, preserve supplying-expression diagnostics through static/helper
receivers, and retain these recovery and stop rules when classifying type values.

The §1.3 anonymous-signature work must resolve nested callable types from their
AST nodes, preserve each node's modifiers and `of object` ownership, and retain
assignment recovery and implicit-call intent without invoking a returned callable.

- [ ] S Per-call truncation marker: `missing_parenthesis1` wants `Invalid Operands` from inside a
  call whose argument list hit a stop. Such calls are currently dropped, which is what makes
  `array_index_bracket_missing1` and `constructor_invalid_param` pass — the marker must keep both.
  Explicit method calls now retain opening/first-token metadata and incomplete fragments for
  scalar property resolution; the broader nested/ordinary call recovery remains open.
- [ ] S Analyzer stops (`"(" expected`) must suppress later diagnostics the way parser stops do.
- [ ] S End-of-compilation hints positioned before a parser stop are still reported; drop them.
- [ ] S The analyzer's compile stop is one flag (unknown name in an expression) that skips the
  end-of-program forward check; generalise it.
- [ ] S Apply the lexer-diagnostic cutoff (`reachedLexerDiagnostics`,
  `internal/frontend/result.go`) on the unit-compile path (`internal/frontend/units.go`) too.
- [ ] S Delete the now mostly dead `must have either a type annotation` text filter in
  `internal/frontend/result.go`.
- [ ] S Inside `begin…end`, the value left unconsumed after a read-only property assignment gets
  a follow-up diagnostic upstream (go-dws reports nothing); indexed read-only property writes get
  none either.

### 2.5 Lexer-owned anchors — S

- [ ] `include_incorrect` wants `"}" expected` at 3:18 (end of the directive argument);
  `directive_messages.go` anchors at 3:13.
- [ ] `conditionals2.1` reports an unbalanced conditional at the directive argument (column 9),
  where the byte-identical `conditionals2` wants the name (column 3).

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
- [ ] S Remaining fixtures of the shape.

### 3.2 Declaration shapes — S each

- [ ] `Name "X" already exists` 20 (12).
- [ ] `Name "X" is reserved` for special-function declaration names, including
  DebugBreak variables, parameters, and global routines; preserve allowed dotted methods.
- [ ] `Class reference expected` 11 (9).
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
- [ ] `special_funcs4`: `Expression expected` for `Inc(i, )`.
- [ ] `assign_untyped`: `Assignment's right-side-argument has no return type`, and
  `Cannot assign a value to the left-side argument` when assigning to a procedure name.
- [ ] `assert`, `enum_byname`: `Boolean expected` anchored at the argument, not the call;
  `enum_byname` also wants `String expected`.
- [ ] `lazy_func_ptr`: `Lazy parameter cannot be a function pointer`.
- [ ] `contracts_error2`: builtins must resolve inside a `require` clause.
- [ ] Measure parenthesized `SizeOf` forms and supply the missing call support;
  `SizeOf(Integer)` currently reports unknown name, while bare punctuation is closed in 2.1.
- [ ] Validate Default type/alias, case-insensitive type-name, and value-argument forms;
  aliases/lowercase types currently fail, and `Default(I)` returns nil for an Integer value.

### 4.3 HelpersFail — S

- [ ] Record the `for` keyword's position on `ast.HelperDecl` (`internal/parser/helpers.go`); all
  six anchors of `mixed_helper` and `helper_of_delegate` are the `for` token.
- [ ] Add an `IsInterfaceHelper` flag and the helper-kind checks (`mixed_helper`; needs 2.2's
  `interface helper for T`).

### 4.4 Other FailureScripts validations — S each

- [ ] `class_const4`: `Constant Instruction - has no effect` as an **error** on a class-const
  declaration.
- [ ] `enum_flags_overflow`: per-element position on `ast.EnumValue`.
- [ ] Remaining silent FailureScripts fixtures from the audit.

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
