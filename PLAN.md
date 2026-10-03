# go-dws — Work Plan

Open work only, split into phases → tasks → subtasks. Completed items are deleted; their
write-up goes to `docs/history/progress-log-<date>.md`. Phase 1.3 retains checked
subitems while the phase is open, so its remaining work is easy to track. Documentation index:
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

**Legend:** `[ ]` open · `[x]` done (retained in §1.3) · ⏸️ gated, do not start · ⚠️ measure before implementing.
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

Completed subitems are retained here until this phase closes. Details and validation:
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

### 1.5 Remaining shapes by origin — L

Work each batch largest shape first, mapping every invented sentence to DWScript's
(`expected ')' after parameter list` → `")" expected`, `unknown type 'X'` → `Type expected`).

- [ ] Parser shapes (overlaps Phase 2).
  - [ ] Unsupported helper-parent syntax:
    keep its compatibility decision separate from member-header diagnostics.
  - [ ] Audit remaining punctuation and type sentences after the Phase 2 prerequisites.
- [ ] `analyze_function_calls.go` / `analyze_method_calls.go`.
  - [ ] Remaining method/inherited arity and constructor overload sentences;
    preserve each path's ordering and recovery types.
  - [ ] Var-parameter arguments: distinguish nonvariables from immutable variables,
    retain parameter names, and measure anchors (`passing_const_var`,
    `passing_const_var2`, `const_param2`, `self_not_writable`).
  - [ ] Remaining intrinsic diagnostics (`Assert`, `Inc`/`Dec`, and others);
    `Swap` zero/excess arguments and delimiter-stop recovery remain open.
  - [ ] Extend default-expression validation to method/record/helper signatures
    and audit the other declaration sentences; regular-routine `params1` is closed.
- [ ] `analyze_statements.go`.
  - [ ] Unknown types and failed inference: `Type expected` / `Type could not be inferenced`.
  - [ ] FOR STEP type/positivity sentences and anchors; constant folding is a
    separate prerequisite (`for_step` and its optimized sibling).
  - [ ] Break/Continue/Exit sentences, with finally nesting and compile stops
    kept explicit (`break_continue`, `break_in_finally`, `exit_result6`).
- [ ] `analyze_classes*.go` (overlaps Phase 3).
- [ ] Other semantic sites, frontend, lexer, shared `internal/errors` builders.
  Remaining incompatible-type pairs include `coalesce`, `in_typecheck1`, and
  `property_default1`; the named task 1.2 fixtures are closed.

The shipped 1.5 batches cover regular parameter defaults, `Swap` data-argument/type
checks, immutable assignments, property accessor names/signatures, invalid
operator recovery with global operator validation, and helper/class/record
member-header stops with earlier declaration diagnostics. Their exact scope and evidence are in
[`the October progress log`](docs/history/progress-log-2026-10.md).

---

## Phase 2 — Parser recovery and compile stops

DWScript's parser sentences and anchors are in place; what is left needs knowledge the parser
does not have, or a precise model of where compilation stops.

### 2.1 Type-directed punctuation — S

An ordinary call says `Expression expected` for `f(;`, so these must be driven from the
semantic side:

- [ ] Record-typed const (`const_record1`).
- [ ] Special functions written without parentheses (`special_funcs1`, `at_integer`).
- [ ] Magic functions (`debugbreak`).
- [ ] Reintroduced properties (`property_reintroduce2`).

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
- [ ] `unexpected "@"` exists nowhere in the tree (`SetOfFail/invalid_operand`,
  `FailureScripts/at_integer`, `dyn_array3`, `field_init1`, `func_ptr6`).

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
