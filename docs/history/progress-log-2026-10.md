# Progress log — October 2026

## 2026-10-03 — Qualified factories and captured indexed receivers (PLAN 1.3)

Imported unit-qualified routines now use the same reference-versus-call selection
as unqualified routines. Typed initialization and assignment invoke a factory
once and retain its returned callable; compatible references remain uninvoked.
Lexical unit-name shadows, captured closures and scalar results retain their
normal behavior. JSON boxing diagnostics report the supplied result type at the
assignment operator without falsely diagnosing the unit name as unknown.

Member-array indexing now captures the receiver once and carries it into
ordinary indexing when indexed-property lookup does not apply. Original AST
nodes retain their call intent, source positions and exception ordering. Nested
arrays, array-returning properties and callable receivers preserve evaluation
order and leave destinations intact when suppliers raise.

The final scoped review exposed a regression in intermediate multi-index record
default properties returning arrays. Captured indexing now consumes each bracket's
comma-owned arguments through existing record-property dispatch. The previously
accepted expression `holder.Values[0][1,2][0]` again produces 12. Compiled traces
verify one receiver read and each index in order; the supported interface-property
path has matching coverage.

Validation on the amended source: `just test-unit` (`go test -v -race ./...`),
CI diff lint (zero issues), `go mod tidy -diff`, all 1,202 tracked Go formatting
and diff checks pass. Focused compiled RED/GREEN tests cover qualified factories,
reference capture, lexical shadows, JSON recovery, receiver counts, property
accessors, grouped record/interface indices and exceptions. Fresh CLI checks
match the eighteen prior diagnostic transcripts, runtime/Boolean controls and
both guide examples, plus the new integration reproductions.

The whole in-scope CLI report has **1,389 / 1,821** passes, with no stable gains
or losses against merged PR #446. Its only differing result is the documented
variable `BuildScripts/init_order2`; the existing floor of seven remains valid.
All thirteen stable fixture gains from the earlier §1.3 work remain. No fixture
oracle or baseline was changed for this follow-up. All §1.3 parents and subitems
are checked and retained. The callable guide describes qualified selection and
once-only indexed reads. This follow-up is based on main after PR #446 merged.

## 2026-10-03 — Recursive callable signatures and JSON boxing (PLAN 1.3)

Callable parameter and return types now resolve recursively from their AST nodes.
Named declarations share the constructor while retaining their registration and
naming boundary. Each signature keeps its own parameter modifiers and `of object`
ownership. The named return parser uses the existing recursive type reader;
class fields and class/record method signatures also use AST resolution.

Assignments and initializers share one reference-versus-call selector. Compatible
references and explicit `@` retain their values; a value-context factory read
commits one call and keeps that call's returned callable uninvoked. Per-node
intent reaches identifier, member and index reads. Methods that have already
supplied a callable result are kept distinct from original callable storage.
Tests cover nested factories, typed/inferred initialization, var/const/lazy
forwarding, repeated lazy forcing, receiver/index capture and supplier exceptions.

Record class variables and static properties participate in the same selection
through record names, instances and inferred metatypes. Callable factory receivers
use ordinary receiver typing, preserving receiver and outer-call counts.
Independent review exposed and closed these missing storage categories.

The unchanged JSON autobox oracle now reports the returned anonymous `procedure `
at 8:3 and `class of TObject` at 9:3. Callable/class-reference boxing failures use
`:=` anchors through variable, field, property and indexed destinations. Scalar,
Variant, JSONVariant and nil assignments remain accepted; record-metatype,
AS/allocation, array and enum diagnostic positions retain their own rules.

The formerly malformed array-of-callable parameter source is now legal. Its exact
source is retained as a clean-compile/no-crash positive, and the malformed row
uses an incomplete nested return while keeping its diagnostic/no-crash assertion.
No fixture oracle was changed. The callable and JSON guides explain result
selection and boxing limits with a factory-counter example.

Final validation: `just test-unit` (`go test -v -race ./...`) passed after clearing
the task build cache following an environment quota failure. CI diff lint reported
zero issues; `go mod tidy -diff`, all 1,200 tracked Go formatting and both diff
checks passed. Fresh CLI verification matched eighteen complete diagnostic
transcripts, eight record fixtures, eight cast fixtures, three JSON fixtures,
factory counters, existing Boolean bytecode paths and both guide examples.
Task, scoped fix and helper-cleanup reviews are clean.

The complete in-scope CLI comparison has no newly failing fixtures. Stable gains
against the record commit are `JSONConnectorFail/autobox`,
`ArrayPass/array_of_proc_param` and `FailureScripts/method_params`. The measured
snapshot also gains `BuildScripts/init_order2`, but a focused rerun with the same
binary changes its initialization/finalization order and fails its oracle; this
existing variation is not counted as a language fix. `just fixture-update` raises
ArrayPass **102 → 103**, FailureScripts **294 → 295**, and JSONConnectorFail **2 → 3**;
the stable BuildScripts floor stays seven. The generated harness snapshot reports
**1,390 / 2,014**, and the in-scope CLI reports 1,390 passes, both with BuildScripts
eight in that snapshot. All §1.3 parents and subitems are checked and retained.
The final whole-branch review identified callable integration gaps, closed by
the follow-up recorded above.

## 2026-10-03 — Record type values and static dispatch (PLAN 1.3)

Record names and aliases used as values now have a distinct nominal `meta of R`
type. Instance annotations retain the record type, and assignments cannot confuse
an instance with its metatype or an independently declared record. The unchanged
`record_meta` transcript now rejects the supplied `TSiteData` instance at 16:35.
Qualified aliases and deferred bodies use lexical identities and source snapshots,
including source `uses` activation and value shadows. AS/IS target terms retain
consistent record-name and alias categories without adding record cast support.

Runtime alias views share the canonical record's class storage and metadata while
retaining their alias helper identity. Language typing and both overload classifiers
preserve the metatype category. Inferred and copied type values expose class
variables, constants, methods and legal class-side properties. Nonstatic record
class helpers receive metatype Self; instance helpers receive instance Self;
static helpers and ordinary record class methods have no Self. Mixed helper
overloads and helper precedence agree with runtime dispatch.

Named and default indexed properties retain the actual receiver and complete
bracket argument list. Independent review caught multi-index argument loss and
array-carried receiver loss; compiled regressions now assert both results and the
ordered receiver/index/getter/RHS/setter trace for reads, writes and compound
assignments. Supported grouped reads are covered; no parser grammar was expanded.
Record/helper expression accessors share class storage by reference, and repeated
runs of one compiled public Program retain independent static state.

Declaration visitor helpers moved into a dedicated sibling file to satisfy the
unchanged file-size limit. The three neighboring helper bodies and signatures
match the pre-task source verbatim. Record semantics remain evaluator-owned;
semantic metatypes contain identity and no mutable runtime storage. The new
[records guide](../guide/records.md) explains instance/type values and helpers.

Final validation: `just test-unit` (`go test -v -race ./...`) passed; CI diff
lint reported zero issues; `go mod tidy -diff`, all 1,197 tracked Go formatting
and staged/unstaged diff checks passed. A fresh CLI matched the complete record
failure transcript, eight positive record fixtures, sixteen earlier diagnostics,
eight cast fixtures, factory counters and all preserved Boolean bytecode paths.
The new guide examples print 3, 3, 3. Task review and scoped fix/style reviews are
clean.

The entire in-scope CLI gains exactly `FailureScripts/record_meta` against the
cast commit, with no losses. Against main the branch gains nine FailureScripts
and `SimpleScripts/enum_bounds`, again with no losses. `just fixture-update`
ratchets FailureScripts **293 → 294**. The CLI reports **1,386** passes; the
harness snapshot reports **1,387 / 2,014**, with the known varying BuildScripts
count eight versus seven in the CLI. Its stable baseline stays seven. The record
parent and all four subitems are checked and retained. Anonymous routine signature
resolution and JSON boxing remain open.

## 2026-10-03 — Cast diagnostics and recovery (PLAN 1.3)

The complete `as_error`, `object_relops` and `as_invalid_right` transcripts now
match upstream, including positions, child-before-parent order and deliberate
omissions after compiler stops. AS retains its RHS term in the AST and generated
visitor. Parser speculation restores cursor, errors and stop state; cloning the
saved and restored cursor fixes repeated lookahead restoration.

Target classification uses lexical identity and declared types. Invalid object
or interface targets recover TObject; unsupported scalar casts retain the RHS
type, including named TClass. Invalid metaclass targets set both structured and
analyzer stop state. Recovered types reach inferred symbols and parent assignments;
AS-supplied mismatches anchor at AS through grouping. Object/class-reference
comparison checks retain Boolean recovery, registered overloads, nil/Variant and
interface identity behavior.

Class AS and IS select their target from its declared type without evaluating
metaclass variables, members or factories. Runtime controls verify zero target
calls and subclass dispatch. Boolean IS preserves Variant/Boolean and alias
operands, and bare/grouped routine and method pointers invoke exactly once using
the original analyzed node. Qualified type targets retain usage, casing and
deprecation metadata; lexical System/Internal/imported-unit values take the
ordinary member path.

Independent review drove regressions for Variant Boolean checks, callable node
identity, lexical namespace shadows and bytecode compilation safety. Bytecode
explicitly rejects class IS and unsupported implicit callable reads, while both
CLI commands carry existing analyzer metadata. A bounded IS type proof preserves
six existing Boolean expression forms in all checked/unchecked compile/run paths;
all 24 combinations also passed on the immutable pre-task CLI. No VM feature or
opcode was added. Assignment diagnostic helpers moved verbatim into their own
file to satisfy the unchanged file-size limit; scoped review verified all thirteen
bodies and the other lint corrections. Additional expectations are derived from
pinned upstream source, not an upstream executable run.

Final validation: `just test-unit` (`go test -v -race ./...`) passed;
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m` reported zero
issues; `go mod tidy -diff`, all 1,187 tracked Go files and diff checks were clean.
A fresh CLI matched sixteen complete diagnostic transcripts, eight positive cast
fixtures, callable factory counters, lexical-shadow controls and Boolean bytecode
preservation cases. Class IS bytecode rejection left no compiled artifact.

The full in-scope CLI comparison gains exactly `as_error`, `as_invalid_right` and
`object_relops` against the enum commit, with no newly failing fixture. Against
main, the branch gains eight FailureScripts and `SimpleScripts/enum_bounds`, again
with no losses. `just fixture-update` ratchets FailureScripts **290 → 293**.
The generated harness snapshot reports **1,386 / 2,014**; the in-scope CLI has
1,385 passes. Their known varying BuildScripts counts are eight and seven,
respectively, so its stable baseline remains seven. Cast parent and all three
subitems are checked and retained. Record metatypes and anonymous-signature/JSON
boxing semantics remain open.

## 2026-10-03 — Source-ordered enum bindings (PLAN 1.3)

Unqualified enum constants now use DWScript's duplicate-preserving sorted local
lookup, including its midpoint selection, sorting and equal-name insertion.
Source types, aliases, overload declarations and actual program unit identities
contribute entries at their declaration events. Builtin members and synthetic
Result/Self bindings remain outside that local vector. Both complete enum
oracles now match: `enums9` selects TEnum1.Hello and `enums10` selects TEnum2.Hello.

Each compiled constant expression retains its canonical enum, ordinal and owning
deprecation metadata. The evaluator creates a fresh value from that binding;
ordinary enum variables retain storage lookup. Deferred bodies replay source
scope roots and declaration/import snapshots, preserving their earlier enum
selection across later declarations and repeated execution. Qualified values,
case-insensitive duplicate checks, scoped/flags enums, explicit ordinal folding,
constant arrays, subrange bounds and sets have compiled execution controls.
Unchecked evaluation retains its existing environment lookup policy.

Units are compiled for availability and become visible at their source `uses`
clauses. Imported parents retain duplicate identities and import order; a unit's
own interface locals precede implementation imports, which precede builtin
fallback. Parameters, locals, enum types and aliases shadow imported unit
qualifiers through lexical identity. Symbol inspection returns the same selected
identity as canonical lookup while preserving overload wrappers. Runtime unit
controls perform the embedding API's complete import/initialization bootstrap.

Independent review found deferred interface scope-root, lexical namespace,
flattened duplicate and implementation-import priority gaps. Failing regressions
drove all four corrections, followed by a clean scoped review. The subsequent
style corrections received a separate clean review. The old casing test now
expects the precise pedantic hint for a parameter spelled `tenum` read as
`TEnum`, following upstream lexical-name checking; no fixture oracle changed.
Primary-source replay is pinned to upstream revision `5f01a346`; extra runtime
outcomes are source-derived controls, not claims of running an upstream binary.

Final validation: `just test-unit` (`go test -v -race ./...`) passed;
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m` reported zero
issues; `go mod tidy -diff`, Go formatting and diff checks were clean. A fresh CLI
matched thirteen complete diagnostic transcripts and both callable-factory
counters. Its ordinary imported Pi control printed 42; an earlier concern was a
unit-test setup error and is withdrawn. A race attempt failed at compiler output
due to temporary-disk quota; clearing only the task build cache allowed the fresh
full run to pass. Checks used writable temporary caches and
`GOFLAGS='-buildvcs=false -p=1'`.

The full in-scope CLI comparison against the previous commit identifies exactly
`FailureScripts/enums9` and `SimpleScripts/enum_bounds` as new passes, with no
newly failing fixtures. Against main, the branch gains those two plus the four
previously recorded array/compound/pointer failures, again with no new failures.
`just fixture-update` raises FailureScripts **289 → 290** and SimpleScripts
**393 → 394**. The generated snapshot scores **1,383 / 2,014** with eight
BuildScripts passes; the independent CLI scores 1,382 with seven, so that known
varying category retains its stable baseline of seven. The enum subitems are
checked and retained in PLAN per the user's request. Cast recovery, record
metatypes and semantic anonymous-routine/JSON autoboxing remain open.

## 2026-10-03 — Routine-pointer assignment recovery (PLAN 1.3)

The complete `FailureScripts/func_ptr1` transcript now matches all fifteen upstream
diagnostics. Compatible bare routines, parameterized references, bound methods,
local routines, pointer copies and explicit `@` retain reference behavior. Rejected
bare routine names yield one declared call result; known direct RHS calls retain
that result after arity errors without relaxing nested argument recovery.
Valueless assignments report the no-return sentence after child errors. Scalar
results supplied to pointer targets report operands first, then canonical type
names at the RHS.

Simple dynamic-array stores retain their separate conversion diagnostic, including
`Cannot assign "void"` at `:=`. This follows the upstream
[direct element-store path](https://github.com/EricGrange/DWScript/blob/5f01a3468452ea75867d4f0e7a0246b107e92332/Source/dwsCompiler.pas#L5848),
which bypasses general assignment construction. Implicit named array factories use
the same policy without reanalyzing receivers. Compound RHS calls preserve declared
results and no-return recovery while retaining literal and class-operator inference.
Existing fixture sources, expectations and scoring remain unchanged.

Separate semantic call intent lets the evaluator invoke a factory once and store
its returned callable. Compiled tests cover named/pointer factories, assignments,
typed initializers, explicit calls, local references, bound receivers and var/lazy
parameters. Lazy suppliers evaluate once per read without memoization; captured
exceptions reach the consuming context unchanged, including expression forwarding.
Coalesce stops before fallback after a raise. Tests verify supplier/factory/inner/
fallback counters, original messages and preserved destinations, plus ordinary
scalar lazy exceptions. Runtime changes stay in the evaluator; parsed ASTs remain
immutable. Independent review found these boundary gaps; the final fixes and lint
refactor passed re-review with no remaining actionable findings.

Controller verification: `just test-unit` (`go test -v -race ./...`) passes;
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m` reports zero issues.
`go mod tidy -diff`, tracked Go formatting and diff checks are clean. The rebuilt
CLI matches eleven complete diagnostic transcripts and both factory counter runs.
An independent CLI comparison against `77f29b06` confirms exactly one new pass,
`func_ptr1`, and no newly failing FailureScripts. Against main `34f25e8c`, the PR's
four gains are `array_assign_error3`, `array_const`, `assign_op_incompatible` and
`func_ptr1`, with no losses.

`just fixture-update` ratchets FailureScripts **288 → 289**, changing no other
baseline. Generated status scores **1,380 / 2,014** with BuildScripts at seven;
the final race run scores **1,381 / 2,014** with BuildScripts at eight. Its known
varying baseline remains seven. Checks use the task's writable Go/lint caches and
`GOFLAGS='-buildvcs=false -p=1'`; only that Go cache was cleared between stopped
builds to avoid exhausting temporary disk quota.

PLAN retains the checked pointer subitems and adds the verified factory and
store/recovery details. Enum binding, cast recovery, record metatypes and anonymous
routine/JSON autoboxing remain open.

## 2026-10-03 — Compound operands and conversion recovery (PLAN 1.3)

Compound assignment checks now validate both operands. Unsupported scalar pairs
report `Incompatible operands` at the operator and retain the ordinary RHS
assignment mismatch when needed. The complete `assign_op_incompatible` fixture
now matches all five upstream diagnostics. Missing class operators stop
compilation; class-operator and dynamic-array append dispatch remain supported,
including assignment through a function-name result.

Registered builtin operators still check assignment compatibility. In particular,
`Integer += Float` remains rejected. Upstream `TAssignExpr.TypeCheckAssign` checks
the original RHS against the target after operator lookup; registrations alone
do not establish whether the assignment is valid. This was verified in
[compiler utilities](https://github.com/EricGrange/DWScript/blob/master/Source/dwsCompilerUtils.pas),
[core expressions](https://github.com/EricGrange/DWScript/blob/master/Source/dwsCoreExprs.pas)
and [conversion expressions](https://github.com/EricGrange/DWScript/blob/master/Source/dwsConvExprs.pas).

Numeric compound assignments convert a Variant RHS before the operation using
the shared conversion path. For example, Integer `10 *= Variant(2.4)` produces
20 while retaining the RHS value. A conversion that raises preserves the target
and original exception. Pending exceptions prevent conversion fallback retries
and writeback through variables, var parameters, implicit fields, properties and
class variables. Runtime tests also retain single receiver/index evaluation,
one property getter read and no setter write after a failed conversion.

Tests were written before each fix. Exact frontend diagnostics cover the complete
fixture, invalid operand recovery, valid numeric/string/Variant pairs, class
operator stops, function-name dispatch and array append. Runtime regressions cover
all four numeric operators, integer/float/string Variant contents, fractional
conversion, builtin conversion failure and eight custom raising-converter
boundaries. Additional array tests cover typed initializer nil rejection and
accepted nil class/interface elements, verifying the existing PR review fixes.
Independent review found the raised-conversion writeback gap; the fix passed
re-review with no remaining actionable findings.

`just fixture-update` ratchets FailureScripts **287 → 288**. An independent CLI
comparison against an immutable snapshot of `68ff7e9f` confirms exactly one newly
passing fixture, `assign_op_incompatible`, and no newly failing fixtures. The
generated status scores **1,380 / 2,014**, with BuildScripts at eight. The full
race run scores **1,379 / 2,014**, with BuildScripts at seven; its previously
documented varying count remains outside this change, and its stable baseline
stays seven. Fixture sources, expectations and scoring remain unchanged.

Validation: `just test-unit` (`go test -v -race ./...`) passed on the final source.
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m` reported zero
issues; `go mod tidy -diff`, tracked/new Go formatting and `git diff --check` are
clean. The rebuilt CLI matches the compound fixture and the three array recovery
fixtures exactly; a function-name class-operator run prints the expected `1` and
retains its returned object. An earlier race run exhausted temporary disk quota; clearing
only the task's Go cache and running race tests and lint sequentially resolved
it. Checks use `GOCACHE=/tmp/go-dws-lint-cache`,
`GOLANGCI_LINT_CACHE=/tmp/go-dws-golangci-cache` and
`GOFLAGS='-buildvcs=false -p=1'`.

PLAN 1.3 retains checked compound subitems and expands the open enum, record and
anonymous callable tasks with their measured dependencies. Enum investigation
now explains both `enums9` and `enums10` through duplicate-preserving sorted local
midpoint lookup, including the program's Default/Internal/System unit entries;
the original [sorting change](https://github.com/EricGrange/DWScript/commit/260d65f9629648b9ab88b7134773f89a818691bb)
updated both fixture oracles. That lookup, immutable runtime enum bindings,
routine-pointer recovery, cast recovery, record metatypes and JSON autoboxing
remain open.

## 2026-10-03 — Assignment anchors and field-backed properties (PLAN 1.3)

Ordinary scalar and interface assignment failures now anchor at their right-hand
expression instead of `:=`. This includes binary coalesce expressions, whose
diagnostic position is the `??` operator, and values on a following line. Existing
array-expression, static-array and indexed-assignment anchors remain intact.

Properties whose writer is a field now use
`Incompatible types: Cannot assign "X" to "Y"` at the supplied value. Method-backed
setters retain their `Argument 0 expects type …` sentence and accessor position.
The same distinction applies to inherited and implicit/explicit `Self` access;
read-only, write-only and metaclass access checks still run before mismatch
reporting. Existing recovery tests now expect the corrected RHS positions.

Four fixtures match their complete upstream expectations: FailureScripts
`assign_error` and `coalesce_class`, and InterfacesFail `assign_obj_from_intf` and
`interface_inheritence2`. Test-first coverage pins all four and retains array
fixtures, multiline values, inherited member paths and method-backed setters.
The local DWScript reference directory was empty; checked-in upstream fixture
expectations supplied the comparison source.

`just fixture-update` ratchets FailureScripts **282 → 284** and InterfacesFail
**6 → 8**. Its generated report scores **1,373 / 2,014** with BuildScripts at
seven passes. The previous generated report had nine BuildScripts passes while
its baseline was already seven; those previously documented varying results are
excluded from this change's four gains. No fixture source, expectation or scoring
policy changed. Remaining PLAN 1.3 groups stay open with explicit fixture names.

Validation: `go test -coverprofile=/tmp/go-dws-assignment-final-coverage.out ./...`
passed, and `golangci-lint run --new-from-merge-base=origin/main --timeout 10m`
reported zero new issues. CLI fixture reports independently agree with the
harness at **284 / 529 scored** in FailureScripts and **8 / 19** in InterfacesFail.
Independent review found no actionable issues; formatting and `git diff --check`
are clean. Validation used `GOCACHE=/tmp/go-dws-lint-cache`,
`GOFLAGS=-buildvcs=false`, and a writable temporary lint cache. The unfiltered
`just ci` lint step still reports the repository's existing backlog; the check
above matches GitHub Actions' gate for newly introduced findings.

## 2026-10-03 — Array allocation recovery and interface conversion wording (PLAN 1.3)

Array allocation dimensions now report `Integer expression expected` for every
non-integer expression. Analysis retains the inferred nested-array type so later
assignments can be checked without spurious undefined-variable diagnostics.
Allocation assignment mismatches use the opening `[` position, including after
spaces, comments or a newline. The parser saves that position separately from the
AST's ordinary `new` position; variables, fields, indexed writes, function-name
writes and field-backed properties use it. Regenerating the AST visitor produced
no changes because the added field is position metadata.

The anchor was verified against upstream `ReadNew`/`ReadNewArray` in
[dwsCompiler.pas](https://github.com/EricGrange/DWScript/blob/master/Source/dwsCompiler.pas#L8758)
and `TestDelete` in
[dwsTokenizer.pas](https://github.com/EricGrange/DWScript/blob/master/Source/dwsTokenizer.pas#L1345):
consuming `[` saves its position before dimension expressions are read. This
rules out using the element type identifier's end as an approximate anchor.

Invalid class-to-interface assignment statements now report
`Class "X" does not implement interface "Y"` at the assignment operator.
The check uses the source's declared class type. Nil, interface aliases, direct
implementations and inherited class implementations remain accepted. The same
conversion sentence applies to fields and field-backed properties; method-backed
writers retain their setter-argument diagnostic. Compatibility rules are unchanged.
`interface_inheritence1` remains open in PLAN 4.5: the analyzer still treats a
class implementing a derived interface as implicitly implementing its base.

Three complete fixtures now match upstream: FailureScripts
`multi_dim_dyn_array1`, and InterfacesFail `assign_intf_from_obj` and
`assign_intf_from_intf`. Tests through the real compile path failed first with
the old sentences, missing second dimension error and wrong anchors. Additional
coverage checks bracket whitespace, recovery typing, field/property routes,
method-backed writers and valid interface conversions. Existing semantic,
frontend, interpreter and embedding tests now expect the corrected sentences.

`just fixture-update` ratchets FailureScripts **284 → 285** and InterfacesFail
**8 → 10**, with no other category changes. The generated report scores
**1,376 / 2,014**. Independent CLI reports agree at **285 / 529 scored** for
FailureScripts and **10 / 19** for InterfacesFail. Other array constructor,
compound operator, enum, cast, record and anonymous-procedure groups in PLAN 1.3
remain open.

Validation: `just test-unit` (`go test -v -race ./...`) passed after updating two
existing interpreter/embedding expectations for the new conversion sentence.
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m` reported zero
new issues, and `go mod tidy -diff` found no dependency changes. The CLI built and
both category reports agreed with the harness. Independent review identified an
explicit-field/property allocation-anchor gap; a failing regression was added
and the gap corrected before final verification. Review then found no remaining
actionable issues.

Checks used `GOCACHE=/tmp/go-dws-lint-cache`, `GOFLAGS=-buildvcs=false`, and
`GOLANGCI_LINT_CACHE=/tmp/go-dws-golangci-cache`. All tracked Go files are formatted
(`git ls-files '*.go' | xargs gofmt -l` produced no output), and `git diff --check`
passed. `just check-fmt` could not create its script under the read-only default
runtime directory; with `XDG_RUNTIME_DIR=/tmp`, it ran but also scanned ignored
local `.cache/` and `.claude/` scratch files and worktrees, reporting their existing
formatting differences. Those local artifacts were left untouched.

## 2026-10-03 — Array constructors and nested callable syntax (PLAN 1.3)

Array assignment recovery now retains incompatible constructors' inferred element
types and cardinality. Dynamic-array targets and constant-array sources use the
assignment operator's position; static targets retain the constructor position.
Writes to constant elements still check the RHS after reporting the read-only
error. The complete `array_assign_error3` and `array_const` expectations now match,
including bounds checks and parser recovery.

Untyped empty array constants can initialize typed dynamic arrays. Compatible
static-to-dynamic conversions copy the source storage and apply the destination's
type metadata, so mutation does not change the constant. Ordinary dynamic arrays
retain reference sharing and element-type invariance. Runtime regressions cover
initializers, assignments, `var` parameters, fields, properties, class variables,
nested and associative slots, `Add`, `Insert`, and nested typed constructors.
Review exposed storage paths bypassing the initial conversion; failing tests
drove the shared conversion and the additional typed-storage boundaries.

The parser retains nested callable return types and their complete source spans,
including `of object`, and recovers a following declaration after malformed nested
functions. This completes only the parser prerequisite for JSON autoboxing;
semantic signature resolution, implicit invocation and boxing diagnostics remain
open. The CLI parses `var p: function: procedure; var tail: Integer;` successfully.

PLAN 1.3 now has explicit subitems and retains checked work while the task is open,
as requested. Enum binding remains open: a blanket first-binding change fixed
`enums9` but regressed the previously passing `enums10`. That implementation was
excluded; the plan now requires measuring and matching both duplicate-name lookup
expectations before changing analyzer or evaluator registration. The merged enum
assignment-anchor fix remains intact.

`just fixture-update` raises FailureScripts **285 → 287**. Independent CLI reports
against a clean snapshot of main (`34f25e8c`) identify exactly the two array gains
and no newly failing FailureScripts fixtures, agreeing with the harness at
**287 / 529 scored**. The generated snapshot scores **1,379 / 2,014**, including
eight BuildScripts passes. That category retains its previously documented
varying results; its stable baseline remains seven. No fixture source,
expectation or scoring policy changed.

Validation: `just test-unit` (`go test -v -race ./...`) passed, and
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m` reported zero
issues. `go mod tidy -diff` found no dependency changes. The CLI built, matched the
harness fixture count, and parsed the nested callable example. Independent review
found no remaining actionable issues after the storage fixes. Tracked and newly
added Go files are formatted, and `git diff --check` passed. Checks used
`GOCACHE=/tmp/go-dws-lint-cache`, `GOFLAGS='-buildvcs=false -p=1'`, and a writable
temporary lint cache. The new branch starts at the merged main commit `34f25e8c`.

## 2026-10-03 — Call arguments and overload diagnostics (PLAN 1.4)

Remaining call-argument type checks now use the shared DWScript diagnostic
helpers for member and implicit-Self calls, record class methods, constructors,
inherited calls, and set `Include`/`Exclude`. Diagnostics use zero-based
argument numbers and expression positions, with receiver shifts for record
instance and helper methods. Valueless arguments retain the short expected-type
sentence, and arguments with existing errors do not receive cascading mismatches.
Expected-type inference and ordinary conversions remain available.

Class signature construction now retains declared strict-parameter metadata,
and explicit and implicit call checks honor it. Frontend regressions distinguish
strict rejection from ordinary Integer-to-Float conversion. Named routines
marked `overload` use the overload sentence for count mismatches even when
only one declaration is visible. Helper scope bindings retain each declaration's
actual directive and source-order visibility. Independent review exposed a later
directive affecting an earlier inline body; a failing regression drove the fix.

Duplicate class-method signatures now report
`There is already a method with name "X"` at the completed method header.
Forward implementations and distinct overloads remain valid. The completed
Phase 1.4 item was removed from PLAN.md, and the error-message guide was updated.
Broader `member_duplicates` collision tracking and `method_implem` body recovery
remain in the existing follow-up phases, as scoped for this change. Inherited
constructor overload-candidate policy is unchanged.

`just fixture-update` raises FailureScripts **295 → 296** (`empty_body`) and
HelpersFail **9 → 10** (`helper_overload_error`). The independent CLI
`just fixture-report --in-scope` agrees on both category counts, with no
category below its previous baseline. The generated snapshot scores
**1,392 / 2,014**. BuildScripts measured eight passes; its previously documented
stable baseline remains seven because results vary between seven and eight.
Fixture sources, expectations and scoring policy were not changed.

Validation: `go test ./...`, `just test-unit` (race detection),
`just test-coverage`, `just check-fmt`, `git diff --check`, and
`go mod tidy -diff` passed. CI's lint command,
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m`, reported
zero issues. Plain `just ci` stops at the existing full-repository lint backlog
(1,230 findings); coverage was verified separately. Independent review's one
actionable finding was fixed and covered by a regression.

Checks used Go 1.26, `GOFLAGS='-buildvcs=false -p=1'`, a task-owned build
cache in the worktree, and `TMPDIR=/tmp`. A disk-backed temporary directory
triggered existing fixture-report test cleanup failures; the final complete
test, race and coverage runs passed with the normal temporary filesystem.
The installed linter was built with Go 1.27 and used the matching cached
toolchain; formatting used the project's Go 1.26 toolchain and
`XDG_RUNTIME_DIR=/tmp`. The branch starts at main commit `ee236fc7`.

## 2026-10-03 — Parameter defaults, Swap, and immutable assignments (PLAN 1.5, first batch)

Phase 1.5 remains open. Its broad origin groups now have concrete remaining
subtasks; this batch closes the regular-routine `params1` and `swap1` type-pair
fixtures and the immutable-assignment sentence sites. Parser investigation and
implementation ran separately from semantic work, with controller-owned frontend
acceptance tests and an independent review.

Modified parameter defaults report the exact `lazy`/`var`/`const` sentence at
`=` and retain the initializer, parameter, and subsequent declarations. External
routines no longer consume a following global constant as a local declaration.
Parameter metadata records the default separator through comments and newlines.
Regular-routine defaults receive constant/type validation and preserve valid nil,
aggregate, pure-call, and constant-cast expressions. Rejected and modified defaults
remain required. Folded scalar mismatches use `=`, nonconstant defaults use the
cursor after the initializer, and their child-first diagnostics remain ahead of
later calls on the same line. Non-const open `array of const` parameters receive
the upstream sentence. Method/record/helper default checks remain an explicit
follow-up.

`Swap` reports `Variable expected` at an invalid data argument, suppresses its
pair after a child/variable error, and reports unequal types at the intrinsic's
name using canonical captions. Variant/concrete pairs are rejected. Array elements
and record/class fields remain usable storage; static-array temporaries,
immutable roots, constants, and getter calls are rejected. Read-side property
classification follows upstream: a field reader is data, while a getter method
is a call result. A single argument receives `"," expected` at `)`. Zero/excess
argument handling and delimiter stops remain open. `@Print`/`@PrintLn` expose
DWScript's one-Variant procedure signature. Existing evaluator storage dispatch
runs valid field/array swaps without runtime changes.

Read-only bindings and class constants now use
`Cannot assign a value to the left-side argument` at the assignment operator.
The new compile-path oracle tests cover complete diagnostic lists and valid
counterexamples. Review findings drove regressions for Variant pairs, immutable
members, static-array temporaries, aggregate pure calls, constant casts, and
same-line default diagnostic ordering. All review findings were resolved.

`just fixture-update` raises FailureScripts **296 → 302** (`class_const3`,
`const_2`, `const_param1`, `const_param4`, `params1`, `swap1`) and SimpleScripts
**394 → 396** (`default_parameters_expr`, `swap2`). CLI reports using binaries
built from main and the implementation confirm all eight gains and no newly
failing in-scope fixtures. The final CLI and Go harness agree on category counts.
The generated complete snapshot is **1,399 / 2,014**; BuildScripts measured seven
passes in both comparison runs and retains its documented stable floor of seven.
Fixture sources, expectations, and scoring policy were not changed.

Validation passed: `go test ./...`, `just test-unit` (race detection),
`just test-coverage`, `just check-fmt`, `git diff --check`, and
`go mod tidy -diff`. Regenerating the AST visitor produced no changes.
`golangci-lint run --new-from-rev=50aac735` reports **zero issues**.
Full `just ci` still stops at the existing **1,230-issue** lint backlog,
matching main's measured count; its tests and coverage were run separately.
Tests used Go 1.26; lint used the installed linter's matching Go 1.27 toolchain.

## 2026-10-03 — Property accessors and operator declarations (PLAN 1.5, second batch)

This batch closes Phase 1.5's property-accessor-name and invalid-operator parser
items, including the semantic prerequisites for their four complete fixture
outputs. Phase 1.5 remains open for the other diagnostic origins.

Missing `read`/`write` accessor names now report `Name expected` as compiler
stops. Invalid class operators stop immediately with `Overloadable operator
expected`; invalid global operators report that ordinary error and then attempt
the operand list, producing the expected `"(" expected` stop for an identifier.
Earlier class members survive parser recovery, preserving their diagnostics.

Property signature validation checks getter result types and setter procedure
kind before parameter compatibility. Explicit index mismatches retain the
zero-based type detail followed by the method summary. Type/signature messages
anchor at the accessor name and use declaration casing, including inherited
members and constants. These measured anchors replace the initial plan's
following-token assumption; no property AST metadata was needed.

Global operator diagnostics check declared operand count at the closing
parenthesis and binding result, count, then parameter types at the binding name.
The public `ast.OperatorDecl.OperandValidationPos` records that closing token;
its primitive field does not change generated visitor traversal. A separate
source-local declaration ledger preserves rejected bindings for later duplicate
checks without registering them as executable overloads. Ordinary, helper, and
built-in conversion paths reject invalid registrations. Existing unary-minus
and valid helper/conversion forms remain supported.

Tests cover exact fixture output, compiler stops and earlier-member recovery,
EOF and comment/newline anchors, casing and validation precedence, numeric and
array signature rejection, Variant indices, var-binding parameters, rejected
binding duplicates, local-scope isolation, and valid properties/operators.
Independent review found five issues in signature compatibility, registration
and constant casing; all were fixed with regression tests. An additional array
result test prevents using runtime array compatibility for binding signatures.

`just fixture-update` raises FailureScripts **302 → 305** (`property_error3`,
`property_error4`, `class_operator3`) and OperatorOverloadFail **3 → 4**
(`operator_overload1`). The generated snapshot is **1,404 / 2,014**, with no other
baseline changes. Fixture sources, expectations and scoring remain unchanged.
CLI comparisons against main `27ba61db` show all four gains and one fluctuating
BuildScripts `init_order2` difference. Twelve direct runs per binary reproduce
multiple initialization/finalization orders in both revisions; the stable
BuildScripts floor remains seven. No other in-scope fixture newly fails.

Validation passed: `go test ./...`, `just test-unit` (race detection),
`just test-coverage`, `just check-fmt`, `git diff --check`, and
`go mod tidy -diff`. `go generate ./pkg/ast` produced no visitor changes.
`golangci-lint run --new-from-rev=27ba61db` reports **zero issues**.
Full `just ci` stops at **1,230 existing lint findings**, matching the measured
main baseline; tests and coverage passed separately. Tests used Go 1.26, lint
the installed linter's matching Go 1.27 toolchain, and a writable Go cache with
`TMPDIR=/tmp` and `GOFLAGS='-buildvcs=false -p=1'`.

## 2026-10-03 — Class, record and helper headers (PLAN 1.5, third batch)

This batch closes the Phase 1.5 helper/class/record member-header diagnostic
item. Unsupported helper-parent syntax remains a separate open item, and
Phase 1.5's other parser and semantic diagnostics remain open.

Invalid members after `class` now stop compilation with
`PROCEDURE or FUNCTION expected`, including top-level class routine headers.
Messages anchor at the offending token, or the final real token at EOF.
Record and helper bodies missing `end` report `END expected` as stops; a helper
also expects `END` where an unsupported member begins. Earlier parsed members
survive the stop, while later declarations cannot add diagnostics.

Record field recovery retains a Variant field after a missing colon, consumes
its semicolon, and continues to later fields. Duplicate fields now report
`There is already a field with name "Name"` using the original declaration's
casing. Helpers retain source visibility sections in additive public AST
metadata, reuse the existing normal-level redundancy checks, and diagnose
unsupported protected visibility without changing the visibility in effect.
Frontend declaration-boundary restoration interleaves visibility diagnostics
and parser stops within each source before imported-unit diagnostics are merged,
preserving other diagnostics' relative order. Incomplete inline records retain
earlier fields and visibility sections for analysis without resolving to a
usable type; complete inline-record resolution errors keep their existing
contract. Regenerating the AST visitor changes no traversal.

Tests pin all six complete fixture outputs, compiler-stop flags, EOF and
comment/newline anchors, earlier mixed-case duplicate recovery, helper hint
levels, inline-record recovery, ordering against earlier method-body and parser
errors, and valid class variables, constants, functions and procedures.
`just fixture-update` raises FailureScripts **305 → 309** (`class_class`,
`class_error1`, `record_syntax1`, `record_syntax2`) and HelpersFail **10 → 12**
(`helper_error5`, `helper_scopes1`). The final generated snapshot is
**1,411 / 2,014**, including an unstable BuildScripts `init_order4` pass.
Twelve direct runs per binary reproduce differing initialization orders in
both baseline and final binaries; its documented stable floor remains seven.
There are no other baseline changes. CLI comparisons against main `db6b8a93`
confirm all six intended gains and no newly failing in-scope fixtures; the
additional `init_order4` pass is not counted as a compatibility fix. Fixture
sources, expectations and scoring were unchanged.

Independent review found two recovery/ordering issues, both fixed with failing
regressions. The full semantic suite also verified the complete inline-record
resolution-error contract; the new partial recovery is limited to incomplete
nodes. No review findings remain open in this batch.

Validation passed: `go test ./...`, `just test-unit` (race detection),
`just test-coverage`, `just check-fmt`, `git diff --check`, and
`go mod tidy -diff`. `go generate ./pkg/ast` produced no visitor changes.
`golangci-lint run --new-from-rev=db6b8a93` reports **zero issues**.
Full `just ci` stops at **1,230 inherited lint findings**, identical to the
measured baseline; tests and coverage passed separately. Tests used Go 1.26,
lint the matching Go 1.27 toolchain, and a writable Go cache with
`TMPDIR=/tmp` and `GOFLAGS='-buildvcs=false -p=1'`.

## 2026-10-03 — Regular var arguments and Self recovery (PLAN 1.5, fourth batch)

This batch closes the named Phase 1.5 regular var-argument fixture set:
`passing_const_var`, `passing_const_var2`, `const_param2`, and
`self_not_writable`. Broader method/inherited/constructor/helper/function-pointer
call-path validation, intrinsic arity/type punctuation, helper-parent syntax,
and writable callable-reference temporaries remain open in `PLAN.md`.

Regular named routines now report `Argument N (name) cannot be passed as
Var-parameter`, retaining the declared parameter casing, zero-based index,
and written argument position. Literals, call results, constants, const/lazy
bindings and object Self cannot supply a writable slot. Type errors and failed
child expressions suppress a second reference diagnostic, while existing
array compatibility recovery still validates storage. Record method scopes
retain constant bindings as immutable symbols.

A dedicated recursive storage check preserves record Self fields and static
array fields, mutable array elements, const object/dynamic-array member
storage, helper class variables and field-backed properties, record static
backing fields, and dynamically writable JSON members. Constant-backed
property readers remain immutable. `Inc` and `Dec` use the same reference
sentence for argument zero with parameter name `a`; broader intrinsic
sentences remain separate work. Parser-accepted Self assignment targets now
reach semantic analysis, where object Self reports `Cannot assign a value to
the left-side argument` at the assignment operator, preserving surrounding
errors and declaration/body hints.

Independent review found five storage/recovery issues, all reproduced by
failing regressions and fixed: array fallback validation, record Self fields,
helper/static-property storage, record constants, and routine addresses that
compiled but lacked runtime by-reference storage. Controller runtime comparison
also caught JSON member storage rejection and verified restored execution
(printing `3`). Explicit `@Routine` var arguments retain their prior rejection.
Pinned upstream treats routine-reference data as writable; implementing its
temporary reference slots needs a separate runtime compatibility change.
Enabling assignment to record Self and ordinal alias typing remain existing
work outside this batch.

`just fixture-update` raises FailureScripts **309 → 313**, with no other
baseline changes. The generated snapshot is **1,414 / 2,014**. CLI comparisons
against main `4315b9eb` confirm the four deterministic gains and no new in-scope
failures. An intermediate CLI snapshot also had a fluctuating BuildScripts `init_order1`
pass: twelve runs per binary produce four distinct output orders before and
three afterward. Its stable floor remains seven, and this fluctuation is not
counted as a compatibility fix. Fixture sources, expectations and scoring are
unchanged.

Validation passed on the final source: `go test ./...`, `just test-unit`
(race detection), `just test-coverage`, the exact frontend acceptance set,
`just fixture-update`, the final fixture gate, `just check-fmt`,
`go mod tidy -diff`, and `git diff --check`. The staged
`golangci-lint run --new-from-rev=4315b9eb` reports **zero issues**.
Full `just ci` stops at **1,230 inherited lint findings**, identical to the
measured baseline; tests and coverage passed separately. Final verification
used Go 1.26, lint the matching Go 1.27 toolchain, a writable shared Go cache,
`GOFLAGS='-buildvcs=false -p=1'`, and a task-specific `/tmp/phase15d-build`
directory. Heavy checks ran sequentially after the initial shared temporary
storage quota failure, and the final runs passed.

## 2026-10-04 — Phase 1.5 class-member and constructor arity

Base: `6138517b`. Branch: `fix/phase15-method-arity-diagnostics`.

Closed the class-member portion of the remaining call shapes:

- Explicit receivers, class methods, and implicit `Self` calls now use DWScript's
  count and overload sentences. A sole marked overload retains overload resolution.
- Named inherited methods select from the ancestor overload set, preserve defaults,
  and anchor diagnostics at the written method or constructor name.
- Constructors through `new`, dotted calls, and metaclass receivers validate supplied
  arguments before ordinary arity. Parameterless class members report
  `Too many arguments`. Marked constructor groups use the overload sentence.
- Child errors precede enclosing count diagnostics across lines; supplied parameter
  type errors suppress an additional count error. Existing result-type recovery,
  visibility checks, and metaclass receiver restrictions are retained.
- Runtime constructor dispatch respects unmarked constructor hiding. Previously an
  inherited parameterless constructor could win over a local constructor with
  omitted defaults and silently skip its body.
- Helper-parent syntax remains a documented go-dws extension, as requested. Existing
  inheritance tests cover this policy; no parser rejection was added.

Exact frontend regressions and compile-to-runtime tests cover these paths, including
defaults, overload selection, and constructor hiding. Source comparisons used upstream
DWScript revision `1dbf8a90329cc3f2638516e89c0668f916c1ddb9`
(`ReadArguments`, `TypeCheckArguments`, `ReadInherited`, `ReadNew`, and
`FindDefaultConstructor`). An old test expecting ordinary arity for a marked
constructor group was corrected to the upstream overload sentence.

The final CLI comparison found no deterministic fixture gains or new failures.
The scored snapshots were 1,415 versus 1,414 passes out of 1,821 fixtures, with the
sole difference `BuildScripts/init_order2`. Eight direct runs of each final and
baseline binary produced two output orders, in counts 2 and 6 for both binaries;
its stable category floor remains 7. FailureScripts remained 313, HelpersFail 12,
and OverloadsFail 3. `just fixture-update` changed only the report date; all baseline
floors stayed unchanged. Fixture sources, expectations, and scoring are unchanged.

Phase 1.5 stays open for record/interface/helper arity, bare inherited calls,
synthetic parameterless/inherited-default constructor ambiguity, remaining
var-argument checks, callable temporaries, intrinsic diagnostics, and
method/record/helper default-expression declaration validation. See
[PLAN.md](../../PLAN.md), [error messages](../guide/error-messages.md),
[helpers](../guide/helpers.md), and
[known divergences](../decisions/known-divergences.md).

Validation passed: `go test ./...`, `just test-unit`, `just test-coverage`, and,
after the final lint style changes, `go test -race -coverprofile=coverage.out ./...`
plus HTML coverage generation. The exact frontend regressions, fixture update and
gate, and `go mod tidy -diff` also passed. Final tests used the CI-compatible
Go 1.24.13 toolchain, `GOFLAGS='-buildvcs=false -p=1'`, compiler temporary files in
the worktree cache, and runtime test directories in `/tmp`. Earlier Go 1.26 runs
hit the shared temporary quota and FUSE executable-cleanup failures; no repository
code was changed to accommodate that environment.

Final formatting and diff checks passed after temporary compiler files were removed.
The staged `golangci-lint run --new-from-rev=6138517b --timeout 10m` reports zero
new issues. Full `just ci` still stops at 1,229 lint backlog findings (the measured
unchanged baseline had 1,230); tests and coverage passed separately.

## 2026-10-04 — Phase 1.5 native record call arity and defaults

Base: `a9ef113d`. Branch: `fix/phase15-record-arity-diagnostics`.

Closed parenthesized native record-call diagnostics for explicit instance,
class-side, metatype-value, implicit Self, and recursive Result-alias calls.
Unmarked signatures use `More arguments expected` / `Too many arguments`,
including parameterless methods. Marked or multiple declarations use the
canonical overload sentence at the written method name. Every supplied child is
read before validating parameter types and count; supplied type mismatches
suppress count errors, while child errors can precede count/overload errors
across lines. Existing declared-return recovery remains available.

Record instance methods retain upstream's receiver numbering and positions:
Self occupies argument zero, and a supplied type error's shifted index selects
the next written argument position, falling back to the method name. Record
class methods have no implicit receiver argument. Class and record calls share
the argument-validation and overload-selection helpers, while helper fallback
policies remain separate.

Default parameter presence is retained in record signatures. Runtime fills
omitted defaults from the record declaration's lexical constant scope, preserving
both record-owned and outer constants against caller-local shadowing. Out-of-line
bodies reuse the runtime callable binding logic to retain declared defaults on
copied executable signatures. Mixed instance/class overload sets are selected
consistently during semantic analysis and runtime dispatch; supplied argument
values are cached so dispatch does not evaluate them twice.

The frontend tests cover 25 exact arity/type/anchor/recovery cases and six full-child
and multiline overload-ordering controls. Runtime tests cover inline and
out-of-line defaults, overloaded implementations, lexical/record constant scope,
implicit and explicit mixed overload selection, expression receivers, and supplied
argument evaluation exactly once. A fresh whole-branch review identified four
Important issues in these runtime/candidate paths; six failing runtime cases were
added first, then all four findings were fixed. There were no deferred minor
findings. Source evidence used DWScript revision
`1dbf8a90329cc3f2638516e89c0668f916c1ddb9`, especially `ReadMethod`,
`ReadArguments`, `TypeCheckArguments`, `TRecordSymbol.CreateSelfParameter`,
`TRecordMethodExpr.Create`, and `ReadParams.GenerateParam`.

The final CLI fixture comparison found no deterministic gains or new failures.
Snapshots were 1,415 before and 1,414 after out of 1,821 in-scope fixtures; their
sole failure-set difference was the existing nondeterministic
`BuildScripts/init_order2`. Eight direct runs of each final/baseline binary yielded
three and four output orders, respectively. FailureScripts stayed 313,
HelpersFail 12, and OverloadsFail 3. `just fixture-update` was run after the fixes;
all stable baseline floors remain unchanged. An earlier generated BuildScripts
7-to-8 fluctuation was discarded after reproducing its nondeterminism.
Fixture sources, expectations, and scoring were not changed.

Phase 1.5 remains open for interface/helper arity, bare record-member invocation
versus routine-reference contexts, the pre-existing noncallable record-name
shadow fallback, declaration-time default-expression validation, broader
multiline child-versus-type ordering, and the other call/intrinsic/storage audits
in [PLAN.md](../../PLAN.md). Updated guides:
[records](../guide/records.md) and [error messages](../guide/error-messages.md).

Final validation passed: `go test -race -coverprofile=coverage.out ./...`, HTML
coverage generation, the focused frontend/runtime acceptance set,
`just fixture-update`, `just fixture-check`, `just check-fmt`,
`go mod tidy -diff`, and `git diff --check`. The final new-findings lint gate
(`golangci-lint run --new-from-rev=a9ef113d --timeout 10m`) reports zero issues.
Full `just ci` stops at 1,229 inherited lint findings, the same total measured
before the batch; tests and coverage were run separately. The clean baseline
`go test ./...` also passed before implementation.

Tests used CI's Go 1.24.13 with `GOFLAGS='-buildvcs=false -p=1'`, a shared generated
Go cache, compiler temporaries in the worktree cache, and runtime test directories
in `/tmp`. Earlier full test/race attempts hit shared disk quota, including fixture
executable copies and sandbox startup. Reclaiming four GiB of old generated Go
cache files restored verification; no source workaround was added. A concurrent
lint attempt was rejected by the lint runner's lock and was rerun serially.

### Phase 1.5 — Direct native interface-call argument diagnostics (2026-10-04)

Direct parenthesized interface method calls now read every supplied argument
before reporting count errors, including inherited methods and expression
receivers. Supplied type mismatches suppress count errors; errors inside children
remain and can precede the outer count error. Parameterless methods reject excess
arguments with `Too many arguments`. Count errors anchor at the method name;
interface receivers stay outside the argument list, so type diagnostics use
zero-based indices and the supplied expression's position. Return types survive
arity recovery for subsequent assignment validation.

The native interface branch reuses the existing member argument checker, leaving
helper fallback, runtime dispatch, signatures, and parser behavior unchanged.
Pinned upstream evidence is `TypeCheckArguments`, `WrapUpFunctionRead`, and
`TMethodExpr.Create` at DWScript commit
`1dbf8a90329cc3f2638516e89c0668f916c1ddb9`.

Acceptance uses 21 exact `frontend.Compile` cases; ten failed before the change.
They cover short/excess/parameterless calls, supplied types before counts,
multiline anchors and child errors, case-insensitive lookup, inherited signatures,
result recovery, expression/grouped receivers, and contextual `[]`/`nil` arguments.
The frontend and semantic suites pass, as do `just fixture-update` and
`just fixture-check`. The fixture updater leaves both baseline/status files
unchanged. The CLI fixture report has no deterministic changes versus the prior
batch; its sole snapshot improvement is `BuildScripts/init_order4`, whose repeated
runs produce varying initialization/finalization orders with both binaries.
No baseline floor was raised for that nondeterminism. The new-findings lint gate
(`golangci-lint run --new-from-rev=a95420db --timeout 10m`) reports zero issues;
full `just ci` stops at the same 1,229 inherited lint findings before and after.
The clean main baseline `go test ./...` passes.

The fresh review found no code issues. Its history-entry note was addressed by
this already planned record. Remaining work stays in PLAN.md: helper argument
policy, interface declaration defaults and omitted-default execution, var-storage
checks, multiline child/type ordering, bare member/reference contexts, and the
existing excess bare-callable argument gap discovered during review. Grouped
member callees such as `(item.Take)(...)` belong to the bare-context audit:
upstream reads the member inside parentheses before the outer call.

Verification uses CI's Go 1.24.13, `GOFLAGS='-buildvcs=false -p=1'`, and a shared
generated Go cache. Initial baseline attempts exceeded temporary-filesystem quota;
compiler temporaries were moved into the worktree cache after reclaiming four GiB
of stale generated Go cache files. No source workaround was added. Full
`go test -race -coverprofile=coverage.out ./...`, HTML coverage generation,
`just check-fmt`, `go mod tidy -diff`, and `git diff --check` pass. The whole-tree
formatter initially traversed active generated build files; its final run passes
after those compiler temporaries are gone.


### Phase 1.5 — Receiver-helper call argument diagnostics (2026-10-04)

Parenthesized helper calls through primitive, record, class, and interface
receivers now read every supplied argument before checking types and count.
A supplied type mismatch suppresses the count error; errors inside children
remain and can precede the outer count error across lines. Parameterless helpers
use `Too many arguments`. Count diagnostics anchor at the helper member name,
and return types remain available for subsequent assignment checks. Implicit
Self helper fallback uses the same selected-signature checker.

Receiver ownership follows the selected overload. Instance helpers and function
helpers reserve argument 0 for Self, as do nonstatic class helpers for structured
record/class targets. Their type errors use the shifted written-argument index
and position, falling back to the method name for the final argument. Static
helper methods and class helpers for primitive/interface targets have no receiver
argument and use ordinary indices and supplied-expression anchors. Static flags
are retained per helper signature; helper candidate selection, runtime dispatch,
and specialized intrinsic policies are unchanged. The implementation reuses the
existing member argument checker instead of separate count-first loops.
Upstream evidence is `CreateMethodCall`, `TypeCheckArguments`, and
`THelperMethodExpr.Create` at DWScript commit
`1dbf8a90329cc3f2638516e89c0668f916c1ddb9`.

Acceptance uses 70 exact `frontend.Compile` cases; 37 failed before implementation.
They cover the four receiver families, instance/class/static ownership, mixed
instance/static names, short/excess/parameterless calls, supplied types before
count errors, shifted multiline anchors, every excess child, case-insensitive
lookup, contextual `[]`/`nil`, inherited helpers, function helpers, implicit Self,
and result-type recovery. Frontend, semantic, and type suites pass. A fresh
read-only review found no blocking issues; additional CLI probes passed inherited
static, alias, class-reference, strict-parameter, and combined child/type cases.
Permanent tests for those additional probes remain a deferred minor.

The CLI fixture reports have no deterministic changes: `HelpersFail` remains
12/18. `function_helper` now has the expected excess-argument sentences but still
fails on its existing header and bare-call gaps. The sole CLI snapshot difference
is `BuildScripts/init_order3`; eight runs per binary produce three output orders
on both base and changed binaries. `just fixture-update` passed but proposed only
a BuildScripts floor increase from 7 to 10. The next fixture check returned to 7
and failed that proposed floor; clean-main repeats also returned 7, 7, and 8.
Both generated baseline/status changes were discarded. The final fixture gate
passes against the stable floor, with no new deterministic fixture gain claimed.

Full `go test -race -coverprofile=coverage.out ./...`, HTML coverage generation,
`just fixture-check`, `just check-fmt`, `go mod tidy -diff`, CLI build/report,
and `git diff --check` pass. The new-findings lint gate
(`golangci-lint run --new-from-rev=ddca9d03 --timeout 10m`) reports zero issues.
Full `just ci` stops at the same 1,229 inherited lint findings before and after;
tests and coverage were run separately. A simultaneous lint attempt encountered
the runner's lock and was rerun serially. Verification used CI's Go 1.24.13,
`GOFLAGS='-buildvcs=false -p=1'`, a shared Go cache, and compiler temporaries in
the worktree cache. Clean baseline `go test ./...` passed before production edits.

PLAN.md retains explicit helper-name, helper-body bound routine, bare/reference
and grouped-callee contexts; helper overload selection; declaration defaults and
omitted-default execution; var-storage validation; and imported-helper and
metaclass-target lookup. The error-message guide describes the shipped receiver
policy and its remaining scope limits.

## 2026-10-04 — Diagnostic roadmap progress tracking

Split the former broad Phase 1.5 backlog into four subphases in [PLAN.md](../../PLAN.md):
1.5 declarations/intrinsics, 1.6 member/helper calls, 1.7 defaults/writable storage,
and 1.8 statements/remaining vocabulary. The eight merged batches from PRs
#449–#456 now appear as checked tasks under their owning subphases, with links
to the corresponding PRs. Historical entries retain their original phase labels.

Every former remaining item is retained as an open task or subtask. The new table
counts completed and remaining top-level tasks, rather than estimating effort or
fixture compatibility. Explicit helper-name/helper-body calls are the next ungated
batch; Phase 2/3 dependencies remain explicit. No diagnostic compatibility work
was closed by this documentation change.

Completed task checkboxes now stay visible while their subphase is open. Update
the checkbox and summary table with each implementation PR, and remove the
checklist only when the whole subphase closes. The completed 1.3 milestone remains
as an explicit checkpoint exception. The documentation index records the same convention. Validation checks task counts, relative links, preservation
of the old backlog, and whitespace; runtime tests are not needed for this
documentation-only change.

## 2026-10-04 — Explicit and body-bound helper calls (PLAN 1.6)

Explicit helper-name calls now analyze every supplied argument before checking
its type and the argument count. A type mismatch suppresses the count diagnostic;
child errors remain before a missing/excess-argument error. The written Self
argument uses ordinary indices and expression positions. Parameterless methods
use `Too many arguments`, and result types survive count recovery.

Calls bound to the helper's own methods inside inline and out-of-line bodies now
use the selected helper signature's receiver role. Instance and structured class
methods retain the implicit Self index/position shift; primitive/interface class
methods and static methods use ordinary positions. Binding identity preserves
parameter shadows and declaration-order visibility. The existing overload
candidate policy remains unchanged, including lone overload-marked bindings;
its independent audit is still open.

The shared checker replaces the duplicate record class receiver guard. Record
class receivers require a matching metatype, including aliases and inferred
metatype copies. A fresh review caught Variant conversion bypassing this role;
a failing compile-path regression was added, then the synthesized metatype
receiver was made strict. The canonical argument-type diagnostic now rejects
that receiver during compilation without the old duplicate `record type expected`.
The review found no other issues.

Acceptance pins 117 complete `frontend.Compile` diagnostic lists across primitive,
record, class, and interface targets, including class/static roles, child/type/count
ordering, multiline anchors, contextual nil/empty arrays, out-of-line calls,
case-insensitive lookup, lexical parameter shadows, and result recovery. The
initial 96-case matrix reproduced 44 failures before the implementation.
Upstream evidence is `TypeCheckArguments`, `CreateMethodCall`, `ReadTypeHelper`,
`ReadSelfMethod`, and `TStructuredTypeMetaSymbol.DoIsCompatible` at DWScript commit
`1dbf8a90329cc3f2638516e89c0668f916c1ddb9`.

PLAN.md checks the explicit/body-bound helper subtask and leaves its parent open
at one of two subtasks complete. Bare invocation/reference and grouped helper
callees are next. Helper overload selection, defaults, var storage, imported
helper availability, and metaclass-target lookup retain their separate open tasks.
The error-message guide describes the shipped contexts and receiver policy.

Validation passes: frontend/semantic/type suites; full
`go test -race -coverprofile=coverage-final.out ./...`; coverage HTML generation;
`just fixture-update` and `just fixture-check`; source formatting;
`go mod tidy -diff`; CLI build/probe; and `git diff --check`.
Fixture baseline/status regeneration produced no changes. CI's new-findings lint
gate (`golangci-lint run --new-from-merge-base=origin/main --timeout 10m`)
reports zero issues; full baseline lint reports the 1,229 inherited findings.
Verification used Go 1.24.13 with writable compiler/cache paths and tmpfs test
storage: worktree test directories hit recurring empty-directory cleanup errors
on both the unchanged baseline and the branch; rerunning on tmpfs passed.


## 2026-10-04 — Bare helper invocation, references, and grouped callees (PLAN 1.6)

Bare user-helper members now read as calls unless a compatible callable context
requires a reference. Missing arguments report at the member and retain the result
type for recovery. Compatible references capture helper ownership and the receiver
once, and defer execution until called. Explicit helper-name and helper-body
references preserve their selected declaration and written argument modifiers.

The parser retains member/name grouping so grouped helper callees read their
inner member before the outer call. Scalar results stop compilation with
`Not a method` at the outer opening parenthesis, preserving earlier child errors
and suppressing later arguments and diagnostics. Callable results execute once;
the outer call uses the shared supplied-type/child/count argument policy. Grouped
ordinary routine references retain their existing expected-type behavior.

Acceptance covers 66 complete compile-path diagnostic lists and 13 compiled-run
cases across primitive, record, class, and interface targets, explicit/helper-body
bindings, function helpers, inherited ownership, lexical shadows, casing, receiver
side effects, returned callables, and references. The initial 48-case matrix had
44 failures. Independent review reproduced four regressions: grouped explicit
helper discrimination, grouped ordinary references, selected explicit overload
identity, and explicit lazy-argument alignment. Each gained a regression test
before its fix; the final grouping tests retain the existing initializer-checker
vocabulary, whose independent audit remains open.

Pinned upstream evidence comes from `WrapUpFunctionRead`, `ReadTypeHelper`,
`ReadTerm`, `ReadBracket`, `ReadSymbol`, and `CPE_NoMethodExpected` at DWScript
commit `1dbf8a90329cc3f2638516e89c0668f916c1ddb9`.
Existing helper declaration-order lookup, static alias/field/property bindings,
parameterless overload reads, native Self.ClassName escape, and nested property
writers remain covered by the fixture harness.

PLAN.md now checks both remaining helper-context subtasks and their parent;
Phase 1.6 shows five complete and eight open tasks. Helper overload-candidate
measurement comes next. Defaults, var-storage checks, imported availability,
metaclass-target lookup, and bare native record/interface reads retain separate
open tasks. The error-message guide describes this batch's behavior.

Validation passes with Go 1.24.13: full
`go test -race -coverprofile=.cache/results/coverage-final.out ./...`, coverage HTML
creation, `just fixture-check`, `just check-fmt`, `go mod tidy -diff`, CLI build and
reference/diagnostic probes, and `git diff --check`. The CI lint gate
(`golangci-lint run --new-from-merge-base=origin/main --timeout 10m`) reports zero
new issues. Final independent review has no remaining findings in this batch.

`just fixture-update` passed but proposed only a BuildScripts floor increase from
7 to 8, within the already documented initialization-order variation. Generated
status/baseline changes were discarded; no deterministic fixture gain is claimed.
The earlier full-run attempt exposed an unchecked evaluator setup without engine
state; the helper metadata lookup now guards it and the existing test passes.
Another verification attempt hit the runner's `/tmp` quota; moving the task's lint
cache into its worktree allowed the final checks to pass with tmpfs test storage.


## 2026-10-04 — Helper overload candidate failures (PLAN 1.6)

Measured helper selection separately from selected-signature argument checking,
using pinned DWScript `1dbf8a90329cc3f2638516e89c0668f916c1ddb9`. The verified
compiler blob is `cffe42f756f8f0a44eeae0e045d6bb2bbe7878b4`; an older scratch
compiler copy differs and was excluded. The
[measurement note](../architecture/helper-overload-measurement-2026-10.md)
records source links, receiver roles, ranking caveats and executable follow-ups.

Receiver calls, explicit helper-name calls and lexical helper-body calls now
honor a sole marked overload. Failed matching reports the overload sentence at
the member, using the declaration's spelling. A found member with no matching
candidate owns the diagnostic; it no longer falls back to an inaccessible member
or an unknown implicit name. Recoverable child calls are analyzed before the
outer no-match, without passing nil recovery types into the shared resolver.
Unmarked calls retain their existing type/count checker and source-order helper
body visibility remains intact.

The test-first compile matrix reproduced singleton count/type errors, invalid
zero-argument anchors, spurious member fallback and nested-child ordering before
the implementation. An existing mixed record/static case now correctly expects
no-match for its marked declaration. Independent review found two new regressions:
failed ranking hid instance-through-type receiver errors and lost pedantic member
case hints. Both were reproduced with failing compile tests before correction;
receiver eligibility and case checking now precede selection. The final matrix
has 63 compile cases and four compiled-run singleton controls, each exercising
receiver, explicit-name and helper-body numeric widening. The full frontend
normalizer deduplicates repeated selected-child diagnostics; these cases pin
rendered diagnostics, not a promise of one internal analysis visit.

`PLAN.md` checks the measurement and two diagnostic subtasks while leaving the
parent open. The top-level Phase 1.6 count remains 5 complete / 8 open. Its next
batch is runtime declaration agreement: a freshly built CLI confirms receiver
Integer/String overloads can execute the wrong same-arity declaration, and empty
child helpers lose inherited overload sets. Both scripts and current outputs are
in the measurement note. Mixed/inherited ownership, tie/Variant ranking, defaults,
imports, metaclass targets and writable storage remain open. The parent-helper
grammar is a port extension and is not claimed as supported by this upstream pin.

Validation: full `go test -race -coverprofile=.cache/results/coverage-final.out ./...`
passes on the final source; coverage HTML was generated. The CI diff lint gate
reports zero new issues. `just fixture-check`, `just check-fmt`, `go mod tidy -diff`,
CLI build/recovery probes and `git diff --check` pass. Full frontend/semantic
checks and independent final review have no remaining blocking findings.

`just fixture-update` passed and proposed only a BuildScripts floor increase
from 7 to 9. A repeat reached 8 and failed that proposed floor, demonstrating the
already documented initialization-order variation. The generated baseline/status
changes were discarded, and the stable fixture gate then passed; no deterministic
fixture gain is claimed. Initial runtime controls were corrected to instantiate
an interface receiver and use the primitive helper's actual target; alias-specific
implicit Self lookup remains a documented runtime follow-up. Lint's sole initial
finding was test import grouping, fixed before final lint and race verification.
