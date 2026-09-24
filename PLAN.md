# go-dws — Work Plan

Open work only, split into phases → tasks → subtasks. Completed items are deleted; their
write-up goes to `docs/history/progress-log-<date>.md`. Documentation index:
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

**Legend:** `[ ]` open · ⏸️ gated, do not start · ⚠️ measure before implementing.
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

### 1.1 Routine-type rendering — M

DWScript renders routine types as `class function ClassType: TClass`,
`function IntToHex(Integer, Integer): String`, `destructor Destroy`,
`procedure Test(const String)`, `procedure (String)`, `procedure TMyProc`: omit `()` when there
are no parameters; an unnamed type keeps the separating space. Needed by 1.2, 1.3 and several
Phase 4 items. Unlocks `func_ptr3`, `func_ptr4`, `func_ptr_mismatch`, `func_ptr_var_param`, and
(with `Destructor can only be invoked on instance`) `func_ptr5`.

- [ ] M Carry name, kind and parameter modifiers on `types.FunctionPointerType`
  (`internal/types/function_pointer.go`). Today `const` is lost in the
  `FunctionType` → `FunctionPointerType` conversion, which is also why `func_ptr_mismatch` judges
  `@Test` compatible with `procedure(Foo: string)`.
- [ ] S Render them. Extend `semanticNamedFunctionPointerName`
  (`internal/semantic/analyze_array_helpers.go`, pinned by `internal/frontend/result_test.go`)
  and stop `errors.SimplifyTypeName` (`internal/errors/errors.go`) truncating at the first `(`.

### 1.2 `Incompatible types: "X" and "Y"` — M

~58 lines over 22 fixtures, the largest missing semantic shape. DWScript uses one sentence
wherever two types fail to unify (target first, supplied second, both quoted); go-dws has a
bespoke sentence per site.

- [ ] S Sentence and anchor for `array_initialization4`, `coalesce_dynarray`, `const_1`.
- [ ] S ⚠️ Array-literal anchors: `array_of_proc2` wants column 9 (whitespace after a comma),
  `array_of_proc` wants the `]`. Likely an artifact of upstream's scanner position; confirm from
  the upstream emit site or park.
- [ ] S Routine-typed cases once 1.1 lands (`SetOfFail/invalid_operand`'s
  `"TMyEnum" and "procedure Test"`, `func_ptr*`).

### 1.3 `Cannot assign "X" to "Y"` — M

~37 lines over 17 fixtures, a different sentence from 1.2 but a shared site.

- [ ] S ⚠️ Measure the anchor: `for_in_subclass` expects it at column 12 of `for c in a do` —
  the `do`, not the `in` every other for-in diagnostic uses.
- [ ] M Split the for-in check: related class types where the assignment narrows get
  `Cannot assign "TBase" to "TChild"`; unrelated ones keep `Incompatible types` (`for_in1`,
  `for_error4`).
- [ ] S Remaining `Cannot assign` sites from the worklist.

### 1.4 Call-argument and overload sentences — S

- [ ] S Move the remaining invented `argument N has type …` sites to
  `analyzeCallArgument`/`analyzeSelfCallArgument`: member calls, implicit-Self calls, record
  class methods, the implicit helper path and constructors (`analyze_function_calls.go`),
  `inherited` calls (`analyze_special.go`), two sites in `analyze_classes.go`, set
  `Include`/`Exclude` (`analyze_method_calls.go`).
- [ ] S `addArgumentCountError` (`analyze_function_calls.go`) should prefer
  `There is no overloaded version of "X" that can be called with these arguments` when
  `Symbol.HasOverloadDirective` is set (`HelpersFail/helper_overload_error`).
- [ ] S Class-method path says `duplicate method signature` where DWScript says
  `There is already a method with name "X"` (`empty_body`, `member_duplicates`, `method_implem`).

### 1.5 Remaining shapes by origin — L

Work each batch largest shape first, mapping every invented sentence to DWScript's
(`expected ')' after parameter list` → `")" expected`, `unknown type 'X'` → `Type expected`).

- [ ] Parser shapes (overlaps Phase 2).
- [ ] `analyze_function_calls.go` / `analyze_method_calls.go`.
- [ ] `analyze_statements.go`.
- [ ] `analyze_classes*.go` (overlaps Phase 3).
- [ ] Other semantic sites, frontend, lexer, shared `internal/errors` builders.

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

- [ ] InterfacesFail.
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
- [ ] SimpleScripts.
- [ ] ArrayPass.
- [ ] Remaining categories.

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
