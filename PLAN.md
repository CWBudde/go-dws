# go-dws — Work Plan

> Rewritten 2026-09-06. This file lists **open work only**. Completed work and the reasoning
> behind it live in [`docs/history/progress-log-2026-07.md`](docs/history/progress-log-2026-07.md);
> the measured audits behind the priorities are
> [`docs/history/CODEBASE_REVIEW_2026-07.md`](docs/history/CODEBASE_REVIEW_2026-07.md) (July 2026)
> and [`docs/architecture/audit-2026-09.md`](docs/architecture/audit-2026-09.md) (September 2026).
> Status numbers are generated, not hand-kept. Documentation index: [`docs/README.md`](docs/README.md).

## 0. Status snapshot

**Headline (2026-09-11):** Go harness and freshly rebuilt CLI both
**968 / 1,928 scored = 50%**, after §3.2.7 closed conditional compilation and §3.3
closed ByteBuffer and JSON ownership and number formatting and call-site column precision in stack traces and record copy-on-assign value semantics and metaclass method pointers and associative key coercion and ARC destructor timing. Both use the shared compile pipeline and scoring rules.
`*Fail` error-detection suites **130 / 647 = 20%**.

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
- Where the remaining failures are (960 total): FailureScripts 406, SimpleScripts 89,
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

**Closed 2026-09-09.** All three items shipped; the ✋ notes below record what was measured and
deliberately left out.

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

**Done (2026-09-09):** L-S2c, closing this section. An indexed property whose accessor is a class
method now resolves through a class name *and* through an instance, on both the read and the
write side, with the metaclass bound as the accessor's receiver and index arity validated against
the declared index parameters. Semantic analysis was tightened to match: reaching such a property
through a class name when the accessor needs an instance is now a compile-time diagnostic with the
same messages the non-indexed metaclass path uses, instead of semantic accepting what the
evaluator could not execute. `SimpleScripts/enum_to_integer` passes, 895 → 896. See
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-09--indexed-properties-with-class-method-accessors-l-s2c).

#### 3.2.3 Contracts

**Closed 2026-09-09.** Method contracts are now resolved through a contract chain — the
executing declaration plus each ancestor declaration of the same method — instead of being read
off the executing declaration alone. Every condition is reported under the class that *declares*
it, so an inherited `require` on a derived instance still names the base method.

**Done (2026-09-09):** L-S3c. Contract failures name the class for a method whose body is written
inline in the class declaration, not only for out-of-line `procedure TFoo.Bar` implementations;
the declaring class comes from the class registry, which also keeps a free function called from
inside a method body unqualified. `SimpleScripts/method_condition` passes.

**Done (2026-09-09):** L-S3a. An override with no `require` of its own runs the ancestor's,
base-most first, reported with the ancestor's name and position. Contract parameters bind by
position, so an ancestor condition is evaluated against the call's arguments even when the
override renamed its parameters.

**Done (2026-09-09):** L-S3b, closing this section. Inherited `ensure` conditions run too, with
`old(...)` capture extended over the same chain. The executing declaration's own postconditions
are checked before any it inherits: when both fail, DWScript reports the derived one.
`SimpleScripts/method_contracts` passes; SimpleScripts 336 → 338. See
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-09--contract-inheritance-and-inline-method-naming-l-s3a-l-s3b-l-s3c).

✋ `Preconditions must be defined in the root method only`: upstream rejects `require` on a
non-root method. Not implemented — it belongs to §4 error-detection parity, and
`FailureScripts/contracts_precondition` also needs `Warning: Constant condition`. The runtime
meanwhile evaluates every `require` in the chain, root-most first.

✋ Class invariants parse into `ClassDecl.Invariants` but are never evaluated. No fixture
demands them; measured 2026-09-09.

#### 3.2.4 Generics

**Closed 2026-09-09.** GenericsPass 15 → 23 (100%). Generic interfaces and generic array
aliases are now templates like classes and records; a class inheritance list can name a
generic instantiation (`class (ITest<Integer>)`); and out-of-line generic method bodies
(`function TTest<T>.Test`) are parsed and cloned once per specialization. Measurement also
showed three of the listed blockers were not generics bugs at all and were fixed as such:
`class external` methods are no longer treated as forward declarations, `@f` on a
function-pointer variable and `nil` as a function-pointer argument now type-check and run,
and a record reaching a builtin's Variant parameter goes through a user-defined
`operator implicit (TRec) : Variant`. L-S4c (`func_ptr1`) and the `tlist1` half of L-S4f
already passed; measured, not implemented. See
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-09--generics-l-s4al-s4f).

✋ `GenericsFail` (8 fixtures, 0 passing) is untouched: it belongs to §4 error-detection
parity. `GenericsFail/implem_mismatch1` now gets further before failing — DWScript's
"T expected but u found" check for a mismatched out-of-line type-parameter name is not
implemented; substitution is positional instead.

✋ Type-parameter constraints (`<T: TObject>`) are parsed and ignored; no fixture in
`GenericsPass` demands them. Measured 2026-09-09.

✋ Comparing a function pointer against `nil` (`f = nil`) still reports
"operator = requires comparable types". Assignment and argument passing work; no fixture
demands the comparison. Measured 2026-09-09.

#### 3.2.5 Overloads and method pointers

Closed 2026-09-09 (L-S5a–L-S5d); `OverloadsPass` 33 → 37 of 39. See
[`docs/history/progress-log-2026-09.md`](docs/history/progress-log-2026-09.md).

✋ The two remaining `OverloadsPass` failures (`overload_ambiguous_delegate`,
`overload_class_method`) expect case-mismatch hints and are blocked by the won't-fix in §5;
their behavior is otherwise correct. The 14 `OverloadsFail` fixtures need
`The function X was forward declared but not implemented`, which does not exist anywhere in
the tree — that is §4 / F7, not this section.

#### 3.2.6 Sets

Closed 2026-09-09 (L-S6a–L-S6c); `SetOfPass` 21 → 25 of 25 (100%), `SetOfFail` 1 → 5,
`SimpleScripts` 338 → 340. A bracket literal now converts on its expected type rather than on
its element shape, set literals fold as compile-time constants, partial record constants are
accepted with defaulted fields, `set of (a, b)` parses in a variable's type, and Float → enum
casts compile. New diagnostics: `Element is out of set bounds`, `Set expected`,
`Enumeration expected`, `Set has too many elements for cast to integer`. See
[`docs/history/progress-log-2026-09.md`](docs/history/progress-log-2026-09.md).

✋ The other nine `SetOfFail` fixtures (`bracket_left_missing`, `bracket_right_missing`,
`for_in_set_missing_do`, `include`, `invalid_method`, `invalid_operand`, `of_missing`,
`test_non_variable`, `type_missing`) are parser-recovery and message-parity work, not set
semantics — they belong to §4 / F7.

#### 3.2.7 Conditional compilation

Closed 2026-09-10 (L-S7a, L-S7b); fixtures 920 → 937, `FailureScripts` 107 → 122,
`SimpleScripts` 340 → 342. `Declared()` is a real compile-time predicate in both the
preprocessor and expression positions, the message directives (`{$HINT}`, `{$WARNING}`,
`{$ERROR}`, `{$FATAL}`, `{$HINTS}`, `{$WARNINGS}`, `{$R}`) have DWScript message/position
parity, and lexer diagnostics reach the front end at all — which also closed twelve
malformed-directive fixtures listed under §4/F7. See
[`docs/history/progress-log-2026-09.md`](docs/history/progress-log-2026-09.md#2026-09-10--conditional-compilation-declared-and-the-message-directives-327).

✋ `FailureScripts/static_methods` regressed and is the one fixture lost: it passed only
because `{$FATAL}` was ignored, and upstream's expectation omits the `Compile Error` line
even though the directive is present. The only structural difference from the three
fixtures where the fatal *is* reported is that its `{$FATAL}` is not at column 1 — a
correlation that holds 5/5 but has no plausible tokenizer mechanism, so it was not encoded.
Reopen if the reference submodule is ever checked out.

✋ `ConditionalDefined(s)` always folds to `False`; `{$DEFINE}` symbols live in preprocessor
state the analyzer cannot reach. Argument validation is complete.

✋ `HelpersPass/declared_helper` resolves all four `Declared()` calls but cannot pass: its
expectation needs the case-mismatch hints §5 marks won't-fix, and
`THelper.Proc(TObject.Create)` — a helper method called with an explicit instance argument —
is an unimplemented call form. **That call form is the one concrete follow-up from this
section** and belongs to §3.2 helper work.

✋ `special_funcs4` (needs `Expression expected` for `Inc(i, )`) and `conditionals2.1` (wants
the unbalanced report at the directive argument, column 9, where the byte-identical
`conditionals2` wants it at the name, column 3) stay with §4 / F7.

✋ Subrange bounds at compile time: no fixture declares a subrange type; zero yield.

### 3.3 Runtime / evaluator

- `[ ]` M Nested lvalue vivification through a key or index (`a[k].field := v`, `a[k][j] := v`,
  `a[k].Add(…)`) → AssociativePass `elements_of_value`, `array_of_dyn`; JSONConnectorPass
  `generate1`, `basic_generate`.
- `[ ]` S DWScript hash iteration order → AssociativePass `records`. The compile-time blocker is
  fixed (the analyzer now indexes the result of a parenless function call); the fixture's only
  remaining diff is `Keys.Join(',')` emitting `a,b` where DWScript emits `b,a`.
- `[ ]` S Re-measure the runtime-panic fixtures (metaclass `ClassName`, class-method dispatch,
  `class of`); the common cases were closed in July, the rest was never re-listed.
- `[ ]` M SimpleScripts to ≥ 85% (345/442 = 78%). Work
  `just fixture-report --category SimpleScripts --list-fails` (identical to the harness list).
- `[ ]` M Triage in-scope categories that have no plan yet: FunctionsTime (1/30),
  FunctionsVariant (0/10), FunctionsGlobalVars (0/16),
  EncodingLib (0/12), Memory (1/13), InnerClassesPass (0/2), FunctionsDebug (0/3).
  First step for each: list fails, bucket by cause, then add concrete items here.
  FunctionsByteBuffer is closed at 19/19; see
  [`docs/guide/bytebuffer.md`](docs/guide/bytebuffer.md).
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

Harness and CLI: 115/647 (FailureScripts 107/528, SetOfFail 5, JSONConnectorFail 2,
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
  check): InterfacesFail 19, HelpersFail 18, OverloadsFail 14,
  PropertyExpressionsFail 10, SetOfFail 9, GenericsFail 8, JSONConnectorFail 7, LambdaFail 6,
  OperatorOverloadFail 6, AssociativeFail 3, AttributesFail 2, InnerClassesFail 1.
  (SetOfFail is no longer at 0: §3.2.6 closed four of its fourteen; the rest is parser
  recovery and message parity.)
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
