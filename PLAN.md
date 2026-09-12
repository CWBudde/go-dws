# go-dws — Work Plan

> Rewritten 2026-09-06. This file lists **open work only**. Completed work and the reasoning
> behind it live in [`docs/history/progress-log-2026-07.md`](docs/history/progress-log-2026-07.md);
> the measured audits behind the priorities are
> [`docs/history/CODEBASE_REVIEW_2026-07.md`](docs/history/CODEBASE_REVIEW_2026-07.md) (July 2026)
> and [`docs/architecture/audit-2026-09.md`](docs/architecture/audit-2026-09.md) (September 2026).
> Status numbers are generated, not hand-kept. Documentation index: [`docs/README.md`](docs/README.md).

## 0. Status snapshot

**Headline (2026-09-12):** Go harness and freshly rebuilt CLI both
**1,078 / 1,930 scored = 56%**, after §3.2.7 closed conditional compilation and a §3.3 merge
train closed these runtime/evaluator items: associative arrays (key coercion, ARC destructor
timing, nested lvalue vivification, DWScript hash iteration order), record copy-on-assign,
JSON ownership and number formatting, call-site column precision in stack traces, metaclass
method pointers, and the ByteBuffer, EncodingLib, GlobalVars, FunctionsTime, FunctionsVariant,
FunctionsDebug and InnerClasses host libraries. §3.3 itself stays open: Memory (1/13) and the
FunctionsGlobalVars `private_vars` remainder (13/16); its runtime-panic re-measurement closed
2026-09-12, finding no panics and three ordinary dispatch/alias defects instead. §4/F5 opened
with the `deprecated` directive family (five fixtures) and a re-measurement that put the
missing-validation queue at 58, not the 82 the 2026-03 archive recorded; the constant-instruction
hint and the array-helper receiver rules closed three more, adopting DWScript's canonical
argument-count vocabulary closed ten, making the keyword operators case-insensitive closed three,
the expression-position implicit call closed one, `Boolean expected` — with an empty `repeat`
body and two miscopied empty-block hints — closed six, and stopping the compiler crashing and
hanging on malformed input closed one. Both use the shared compile pipeline and scoring rules.
`*Fail` error-detection suites **158 / 640 = 25%**.

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
- Where the remaining failures are (876 total): FailureScripts 400, SimpleScripts 80,
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

**Done (2026-09-12):** the keyword operators are case-insensitive. The AST carried the operator's
source spelling, and the analyzer, the evaluator and the bytecode compiler each compare it against
a lowercase literal, so `6 and 3` compiled and `6 And 3` drew `unknown binary operator: And` —
taking the whole declaration with it, since `var a := 6 And 3` then leaves `a` undeclared. Folded
once where the node is built (`operatorSpelling`), which covers `and or xor not div mod shl shr
sar in implies`. Closed `bitwise_booleans`, `bitwise_shift` and `func_result_as_byref`.

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

✋ `for <var> in <set>` does not type-check the loop variable against the set's element type:
`var i: Integer; for i in s do` over a `set of TEnum` is accepted silently. Found 2026-09-11
while clearing the §3.4 skipped-test backlog; the commented-out
`TestLargeSetForInLoopVariableTypeError` that documented it was deleted. No fixture demands it.

### 3.3 Runtime / evaluator

- `[ ]` M Triage the remaining untouched in-scope category: Memory (1/13). Its two scored
  fails (`external_constructor_exception`, `external_constructor_exception2`) need host-exposed
  external classes (`TExposedClass`), which is host-integration territory; the other ten have no
  `.txt` and are therefore unscored. First step: list fails, bucket by cause, then add concrete
  items here.
  The categories this section used to list are now closed: FunctionsByteBuffer 19/19 (see
  [`docs/guide/bytebuffer.md`](docs/guide/bytebuffer.md)), FunctionsTime 27/27,
  FunctionsVariant 9/9, FunctionsDebug 3/3, InnerClassesPass 2/2 and EncodingLib 12/12.
- `[ ]` M FunctionsGlobalVars `private_vars` (13/16, library shipped — see
  [`docs/guide/global-vars.md`](docs/guide/global-vars.md)). The parser half is done
  (2026-09-12): a unit written without `interface`/`implementation` sections now parses. What
  remains is the per-unit `WritePrivateVar`/`ReadPrivateVar`/`PrivateVarsNames`/
  `CleanupPrivateVars` family, and the blocker is **unit identity at run time**, which nothing
  currently tracks: neither `runtime.MethodMetadata`/`FunctionMetadata` nor the execution
  context records which unit a body came from. Needs, in order: (1) record the declaring unit on
  callable metadata when `ImportUnitSymbols` installs it; (2) carry it on the call stack so the
  executing unit is known; (3) add `CurrentUnit() string` to `builtins.Context`; (4) implement
  the four builtins over a per-unit store keyed by that name, raising
  `Private variables cannot be referred from main module` when the caller is the main module.
  Sized M rather than S because of (1)–(3), not the builtins.
- ✋ FunctionsGlobalVars `queue_snapshot`: measured 2026-09-12, the produced output already
  matches the expectation exactly, line for line. The only difference is four
  `"join" does not match case of declaration ("Join")` hints, and the discriminator is not
  recoverable: our analyzer types `Map`'s result from the callback's return type, so the
  receiver here is `array of String` — the same element type as `ArrayPass/dynamic_array_remove`,
  where upstream *does* emit the hint. Making these two disagree would mean regressing `Map`'s
  return-type inference to match a hint quirk. This is the §5 case-mismatch won't-fix, not a
  GlobalVars gap.
- ✋ UTF-16 surrogate iteration (`for_in_str`, `for_in_str2`): intentional divergence, see
  [`docs/decisions/string-encoding.md`](docs/decisions/string-encoding.md).

**Done (2026-09-12):** the runtime-panic re-measurement, which found no panics at all. All 87
then-failing `SimpleScripts` fixtures were run through the CLI and none produced a Go panic or
goroutine dump (there is no `recover` on the run path, so one would surface). The three areas the
item named were failing for ordinary, concrete reasons instead, and all three are now fixed —
nil-metaclass message parity, class aliases as class names, and non-virtual/`reintroduce`
dispatch. `SimpleScripts` 349 → 354, fixtures 1044 → 1049. See
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-12--the-runtime-panic-re-measurement-33).

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

**Done (2026-09-12):** the engine seam for `platform.Platform`, closing the last buildable item
in this section. `dwscript.WithPlatform(platform.Platform) Option` installs a platform on the
engine; `Engine.Platform()`/`Engine.FS()` report the one in force, defaulting to the build's
platform (native, or the WASM virtual filesystem) rather than nil. The platform rides on
`contracts.EngineState`, `builtins.Context` gained `FS()`, and the first two file built-ins —
`LoadTextFromFile` and `SaveTextToFile` — go through it and nowhere near `os`. The WASM bridge
now hands its platform to the engine, so a host filesystem installed through `init({fs})` is one
a script actually reads: `just wasm-smoke` round-trips a script through a JavaScript `Map`-backed
filesystem against the real WASM build. Fixtures unchanged
at 1044, as expected: no scored fixture calls either built-in, and the FunctionsFile category
needs a `File` handle type, the path helpers (`ExtractFileExt`, `ChangeFileExt`, …) and directory
enumeration, none of which this seam provides. That category stays out of scope.
See [the September progress log](docs/history/progress-log-2026-09.md#2026-09-12--the-platformplatform-engine-seam).

**Done (2026-09-12):** class-hierarchy distance in overload matching. The TODO this item named
(`internal/semantic/overload_resolution.go:211`) no longer exists — that file is now an 81-line
facade over `internal/types`, and `types.SignatureDistance` has ranked class arguments by
inheritance steps (`classDistance`) since the type-system consolidation. Measured and pinned with
regression tests: given `TC < TB < TA` and overloads on `TA` and `TB`, a `TC` argument now
provably selects `TB`. Nothing was implemented; the item was stale bookkeeping.

**Done (2026-09-12):** the three `t.Skip`ped class-operator inheritance tests in
`internal/interp/operator_test.go` are revived. The "pre-existing bug in operator inheritance
with mixed types" they documented does not exist: multi-level and deep-hierarchy operator
resolution already worked, and the third test failed only because its constructor parameter `id`
shadows the field `ID` — DWScript is case-insensitive, so `ID := id` is a self-assignment and the
field is never written. Renaming the parameter is the fix; the scoping behaviour is correct.

**Done (2026-09-11):** the skipped-test backlog is cleared — every entry was revived, deleted or
turned into a real check, and const static-array element assignment is now diagnosed
(`FailureScripts` 125 → 126). See
[the September progress log](docs/history/progress-log-2026-09.md#2026-09-11--the-skipped-test-backlog-34).

---

## 4. Error-detection parity (`*Fail` suites, F)

Harness and CLI: 158/640 (FailureScripts 149/529, SetOfFail 5, JSONConnectorFail 2,
AssociativeFail 1, InterfacesFail 1, every other `*Fail` suite 0). The suites are **compile-only** (like DWScript's
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
- **F5** `[~]` M Missing-validation sweep: the `FailureScripts` fixtures where DWScript
  reports something and go-dws compiles clean. **Re-measured 2026-09-12: 58, not the 82 the
  2026-03 archive recorded** — the list is regenerated by running every fixture through
  `dwscript run --diagnostics=plain --compile-only --hints pedantic` and keeping the ones that
  print nothing. Two names the old list gave as examples do not belong: `conditionals1-6` and
  `switch_invalid1-3` already emit diagnostics, so they are directive **message parity**
  (`internal/lexer/directive_messages.go`), not missing validation, and belong to F7.
  **Re-measured again 2026-09-12 after the argument-count slice: 48.** The queue is now almost
  entirely single-fixture work; only one bucket has more than one holder:
  - `Warning: Constant condition` — 2 fixtures
  - single-fixture items: `proc_with_result`, `readonly_field`, `const_param2`, `assigned`,
    `ord`, `enum_flags_overflow`, `default_params2`, `for_var_usage`, `case_of_else`, …
    - `Constant Instruction - has no effect` still has one holder, `class_const4`, where upstream
    reports it as an **error** on a class-const declaration rather than a hint on a statement.
    - `func_ptr_mismatch` still prints nothing, but no longer for want of the implicit call,
    which shipped 2026-09-12. Two things are missing instead. `const` does not survive the
    `types.FunctionType` → `types.FunctionPointerType` conversion, which has no slot for
    parameter modifiers, so `@Test` is judged compatible with `procedure(Foo: string)` and
    nothing is reported; and the message needs DWScript's rendering of routine types,
    `"procedure Test(const String)"`, which `errors.SimplifyTypeName` truncates at the first
    `(` to `"procedure"`. `func_ptr4` (`"class function ClassType: TClass"`) and `func_ptr1`
    (`"procedure TMyProc"`, plus `Assignment's right-side-argument has no return type`) need
    the same renderer. `array_of_proc`, `array_of_proc2` and `const_procedure_array` now emit
    the arity error and need the array constructor's own unification diagnostic,
    `Incompatible types: "void" and "nil"`, whose positions are not the element's — `[5:11]`
    is the `]` and `[5:9]` the whitespace after the comma.
  Two of these are blocked on missing AST position data rather than on the check itself:
  `enum_flags_overflow` needs a per-element position on `ast.EnumValue` (only `EnumDecl` has
  one today), and `default_params2` needs a constant-folded comparison of two default-value
  expressions, which `mergeDefaultValues` has no helper for.

- **F6** `[ ]` S Runtime-mismatch residue (13, not re-measured since the 2026-09-12
  argument-count slice, which moved `dyn_array_setlength3` and `missing_param1` out of it):
  `div_by_zero_float`/`_int`, `for_in_subclass`, ….
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

**Done (2026-09-12):** the compiler no longer crashes or hangs on malformed input. Not a message
slice: nine fixtures segfaulted the compile pipeline and one looped forever, which is worse than a
wrong sentence — the process dies, or never answers — and both are reachable from the public
embedding API through `frontend.AnalyzeParsed`, not just the CLI. §3.3's runtime-panic
re-measurement did not cover these; they are compile-time. The crashes shared one cause: most
statement and type parsers return a *concrete* node pointer rather than the `ast.Statement` /
`ast.TypeExpression` interface, so a `return nil` on a parse failure becomes a **typed nil** —
non-nil as an interface, faulting on any field access — and every `!= nil` guard, `ParseProgram`'s
own included, waved it through. A malformed routine header left a typed-nil `*ast.FunctionDecl` at
the top level and the generic monomorphizer faulted walking it, in scripts using no generics at
all; an unsupported function-pointer return type left a typed-nil element type that faulted inside
the parser. `statementOrNil` and `typeExpressionOrNil` normalize the two dispatchers, so the
interface is nil exactly when the parse failed. The hang was pre-existing and separate:
`synchronize` lists `IDENT` among its safe points, so asked to recover *from* an identifier it
returns without advancing, and `parseRecordBody` reported the same misplaced field until memory ran
out — which is what made a whole-corpus in-process sweep unrunnable. Defence in depth, since the
parser is not the only thing that builds an AST: monomorphization now runs under the same `recover`
discipline as semantic analysis (it ran outside it), `genericMethodImpl` guards the typed nil its
type assertion accepts, and `extractUsedUnits` skips nil unit names. Measured over all 2,127
fixtures: **0 panics and 0 timeouts, against 9 and 1 before**. Closed `record_recursive3`, which
now matches exactly; the other eight fail on message parity rather than a signal, and
`ArrayPass/array_of_proc_param` needs `function : procedure` return types — a real feature gap that
now reports the diagnostic it already had instead of faulting. Fixtures 1,077 → 1,078.

**Done (2026-09-12):** the `deprecated` directive family, the first F5 slice. The parser already
recorded `deprecated` on classes, routines, constants and enum elements, and the analyzer warned
for deprecated *classes* only — `Symbol.IsDeprecated` was scaffolding nothing ever set and
`NewDeprecatedWarning` had zero call sites. Deprecation is now carried on the symbol (routines,
constants, enum elements), on `types.MethodInfo`, and on `types.PropertyInfo` /
`types.RecordPropertyInfo`, and warned at every use site: bare and parenthesized calls, method
calls, property reads and writes including indexed and default-property (`t[i]`) access, and
inheriting from a deprecated class. `deprecated` on a property is newly parsed for both classes
and records — on records it was previously mis-parsed as a field declaration and produced three
spurious errors. `FailureScripts/deprecated`, `deprecated_property`, `deprecated_empty` and
`SimpleScripts/const_deprecated`, `enum_element_deprecated` pass; fixtures 1,049 → 1,054.

**Done (2026-09-12):** the constant-instruction hint and the array-helper receiver rules, the
second F5 slice. A statement whose expression is provably constant now draws DWScript's
`Constant Instruction - has no effect`; constness is decided structurally, never by folding, so
`StrToInt('A');` is reported without being evaluated. Separately, the intrinsic array helpers
that resize or reorder storage are refused on a static array (`Array method "X" is restricted to
dynamic arrays`), and the ones needing actual storage are refused on a bare type name
(`Array instance expected`) — `Low` excepted, since it is 0 for every dynamic array, as are a
static array's bounds. `FailureScripts/ignore_result`, `array_static_methods` and `dyn_array4`
pass; fixtures 1,054 → 1,057.

**Done (2026-09-12):** `Boolean expected`. Not an F5 slice — every fixture it closed already
printed something, and F5's silent list is unchanged at 48 — but message parity of the same kind,
plus two genuine defects found while measuring it. DWScript names the type a context
required and nothing else — not the type it got, not the construct that wanted it — so every
wrong-typed condition in the language reports `Boolean expected` (`if`, `while`, `until`, the
if-then-else expression, `require`, `ensure`) and a contract's message half reports
`String expected`, re-using the condition's anchor rather than the message expression's. The anchor
is the first token of the smallest unit that owns the value, and where that unit has an introducer
the introducer wins over the expression: `while` at column 1 rather than the condition at 7,
`until` at 8 rather than `repeat` at 1 or the condition at 14. `ast.RepeatStatement` carried only
the `repeat` keyword and now carries `UntilPos`. The `Infinite loop` warning is deliberately left
alone: `loop_infinite` wants it on `until`, `infinite_loop` on the condition, and both arrived in
the same import commit, so the suite does not say which is right. An empty `repeat` body is legal —
`repeat until X;` is a do-while that only tests its condition — and the parser's guard against it
was what kept `repeat1` and `repeat2` from ever reaching the condition. Two empty-block hints were
also wrong, found by measurement rather than looked for: `analyzeWhile` emitted `Empty FOR loop`
(the FOR loops' own hint; upstream emits none for a while), and `Empty ELSE block` was reported
beside an equally empty THEN, where upstream gives one hint per `if`. Closed `contracts_types`,
`loop_nonbool`, `repeat1`, `repeat2`, `if_empty_terms` and `ifthenelse_optimize1`. Still open in
this bucket: `assert` and `enum_byname` want the same sentences anchored at the *argument* rather
than the call, `ifthenelse_expression1` fails on parser recovery after `if 2=2 1`, and
`contracts_error2` needs a builtin to resolve inside a `require` clause.

**Done (2026-09-12):** the expression-position implicit call, the fourth F5 slice. DWScript reads
a routine name as a call and converts it back to a reference only where the context wants a
function pointer whose signature the routine actually fits; where the conversion does not apply the
call reading stands, so a routine with required parameters draws `More arguments expected` before
the type error. `func_ptr1` pins both sides: `p := Proc2` reports it, `p := Proc4` does not, because
`Proc4()` is well-formed. go-dws defaults the other way — `analyzeIdentifier` returns a pointer type
— so rather than invert that, the rule is applied where the context has already rejected the
reference (`checkPointerContextArity`), at four sites: a bare name in a pointer context, a bare name
in any other value context, `@Routine`, and a function-pointer operand of `=`/`<>`. The operand case
has no name token and is anchored at the operator, alongside the `Invalid Operands` that follows it.
The intrinsic array helpers are exempt — `a.ForEach(IntToStr)` keeps the reference reading and names
the routine's own signature — so their callback argument goes through
`analyzeArrayHelperCallbackArg`. Calls *through* a pointer, the third path the argument-count slice
left alone, now use the same two sentences and no longer leak a non-wire-format line; a miscounted
call yields the pointer's result type rather than nil, which had produced a spurious
`'p' is not a function` on top of the arity error. Closed `callback_err_vs_nil`.

**Done (2026-09-12):** DWScript's canonical argument-count vocabulary, the third F5 slice.
go-dws named the routine and the counts (`function 'Test' expects 2 arguments, got 1`) and
anchored at the opening parenthesis; upstream says only `More arguments expected`,
`Too many arguments` or `No arguments expected`, anchored at the name being called, and the last
of those only when the routine declares no parameters at all. Every arity check at a call site
that *names a routine* — plain calls, methods, interface and record methods, helper methods,
constructors and `new` — now says that, and three rules that follow from it shipped with it.
(The specialized built-in analyzers, the signature-driven registry path and function-pointer
calls still describe their own counts; converting those is a separate slice.) A bare routine name in
statement position is a call, so `Test;`, `Sin;` and `TTest.Test;` report the missing arguments
(overload-aware across the class hierarchy, or `meth_overload_hide` would have regressed); the
array helpers that need an argument report it in the bare member form too; and an indexed
property named without its indices reads the accessor with nothing. Upstream also type-checks
the arguments it was handed *before* it counts them, so a short call whose arguments do not fit
reports the type error alone — implemented for the plain-call path only, since `func_params1` is
the one fixture pinning the ordering. Finally, the built-ins DWScript declares as overload sets — `Abs`,
`Sqr`, `Min`, `Max` — name no count at all: any call they cannot match reports
`There is no overloaded version of "X" that can be called with these arguments`.
`FailureScripts/dyn_array_setlength3`, `func_params1`, `method_missing_arg`, `missing_param1`,
`missing_param1b`, `missing_param2`, `missing_param3`, `property_error10`, `sqr` and
`InterfacesFail/error_in_method` pass; fixtures 1,057 → 1,067, and InterfacesFail leaves 0.

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
