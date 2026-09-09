# Progress log — September 2026

Closed 2026-09-06 on branch `feat/phase1-measurement-tooling`. Items T1–T6 and A1 of the
2026-09-06 `PLAN.md`. Numbers are from the runs recorded in the commit messages.

## Headline

| | before | after |
| --- | --- | --- |
| Go harness (`TestDWScriptFixtures`) | 863 / 1,928 | 871 / 1,928 |
| CLI (`just fixture-report`) | 733 / 1,928 | 871 / 1,928 |
| FailureScripts harness / CLI | 103 / 1 | 107 / 107 |
| Categories where the two disagree | many (130 fixtures) | 0 of 61 |

## T1 / A1 — one compile pipeline

`cmd/dwscript run` used to re-implement lexer → parser → generics → semantic by hand
(`run.go:200-320`), skipping semantic analysis for unit programs, aborting on any parse error,
never calling `SetParseHadErrors`, and leaving `--hints off` at the analyzer's default level.
`internal/frontend` gained `Options`, `ParseWithOptions`, `AnalyzeParsed`, `CompileWithOptions`
and `Result.HintStrings`; the CLI now consumes them. The unit-program bypass stays, but as an
explicit `Options.TypeCheck=false` with the reason in a comment, because neither the analyzer
nor the frontend resolves program-level `uses` yet (new `PLAN.md` §3.2 item). This is the one
place the runners still differ: the harness type-checks unit programs (and fails them), the CLI
runs them untyped; both score them as failures today.

cobra no longer echoes the error twice and dumps the usage block after script failures
(`SilenceErrors` on the root, `SilenceUsage` once `run` has valid arguments; `cmd.ErrSilent`
for commands that already printed their diagnostics).

## T2 — `--diagnostics=plain|pretty`, `NO_COLOR`

`plain` prints the DWScript wire format (`Syntax Error: … [line: N, column: M]`, `Hint: …`,
`Runtime Error: …`) that `frontend.Result.DiagnosticStrings()` and the newly exported
`interp.FormatRuntimeErrorValue` produce. `pretty` (default) keeps the source-excerpt blocks; its
ANSI colors are off under `NO_COLOR` and when stderr is not a terminal (previously hard-coded on).

## T3 — `--test-envelope` and `--compile-only`

`--test-envelope` buffers program output and prints DWScript's test-runner framing
(`Errors >>>>` / hints and runtime error / `Result >>>>` / output) iff there is at least one
message, which is the rule in `reference/dwscript-original/Test/UScriptTests.pas:208-216`.

Investigating the last three-fixture gap showed that upstream's `CompilationFailure` runner
(`UScriptTests.pas:305-363`) never executes `*Fail` fixtures: it compares the compiler's message
list, hints included, and an empty list for programs that compile cleanly. No `*Fail` expectation
in the corpus contains a runtime error. So the CLI gained `--compile-only`, `fixture-report`
uses it for error categories, and the harness's `scoreErrorFixture` dropped its "compiled clean →
execute and expect a runtime error" branch. FailureScripts 103 → 107: `class_unused_privates`,
`empty_if_block`, `unused_result`, `unused_variables` (hint-only) and `div_by_zero_float`
(compiles clean, empty expectation).

## T4 — `fixture-report` rebuilds, refuses stale binaries; baselines ratcheted

`cmd/fixture-report` runs the CLI in harness mode with the harness's per-category hint level
(pedantic; `Algorithms`/`FunctionsString` normal), rebuilds `bin/dwscript` by default
(`--build=false` to skip) and otherwise refuses to run when the binary is older than any tracked
non-test Go source (`--allow-stale`). Previously it ran whatever binary was there at `hints=off`
and compared colored output plus cobra's usage dump, which is why `*Fail` scored 1/528. The
harness gained `FIXTURE_LIST_FAILS=1` so the two failure lists can be diffed.
`baselines.json`: FailureScripts 103 → 107, FunctionsString 53 → 57, SimpleScripts 326 → 330.

## T5 — helper-spec parity test

`internal/interp/helper_spec_parity_test.go` pins the analyzer's helper table, the runtime table
and the evaluator's string switches (parsed with `go/parser`) to each other. It found the audited
11 missing `__array_*` entries in the semantic table (added as `BuiltinMethods` only; A2 unifies
the tables) and six specs registered on both sides with no handler at all: `PadLeft`, `PadRight`,
`StrDeleteLeft`, `StrDeleteRight`, `NormalizeString`, `StripAccents`. They now route to
`builtins.DefaultRegistry` with the receiver as first argument; `CallBuiltinHelperProperty` had no
callers and now sits on the property-read path. FunctionsString 53 → 57 (`pad_left_right`,
`stripaccents`, `normalize`).

## T6 — unscored expected-output variants

`testdata/fixtures/README.md` documents that `.optimized.txt` (upstream `coOptimize` runs),
`.jstxt` (JS backend) and `.fpctxt` are never scored and that `.optimized.txt` is not an
accepted alternative: go-dws has no optimizer and upstream mandates `.txt` in the non-optimized
configuration.

## Known follow-ups surfaced

- Runtime error text: the evaluator reports `division by zero: 1 div 0` where DWScript says
  `Division by zero` (PLAN.md §3.3, runtime message parity).
- `cmd/dwscript TestStringFunctions/Format_Function` (`Format('%f')` precision) was already red
  on `main` before this work.

## Architecture refactoring — 2026-09-07

A2 consolidates builtin helper names, operation identifiers, signatures, aliases,
properties and default arguments in `internal/types/helper_specs.go`. The analyzer
and runtime registrations consume that catalog, and evaluator dispatch uses its
operation constants. Specialized receiver-dependent checks retain their diagnostics.
Catalog isolation and case-insensitivity tests complement interpreter registration
and executable-dispatch parity checks.

A3 removes confirmed-dead control-flow, lazy-thunk, comparison, encoding, variant,
and JSON compatibility code, unused value wrappers, and the `interp/runner`
pass-through. CLI, embedding and WASM entry points construct the interpreter directly.
The audit's constructor count was based on CLI reachability: scalar constructors and
Go-to-script conversion helpers used by public FFI marshaling are retained. JSON
conversion tests now exercise runtime implementations, and nil/range tests exercise
actual evaluator behavior. Runtime package documentation no longer promises an obsolete
file split.

A4 removes the interpreter's duplicate `builtins.Context`, builtin dispatch and
higher-order collection implementation. Both direct host calls and Go callbacks invoke
function pointers through the evaluator with an explicit execution context. Tests cover
case-insensitive builtin pointers, captured lambdas, repeated calls, invalid callables,
and the absence of a second shell builtin context.

A8 moves overload selection, signature equality and conversion-distance ranking into
`internal/types`. Evaluator callers pass types and use the selected candidate index;
they no longer allocate semantic symbols or import the semantic analyzer. Semantic
callers retain a thin symbol adapter. Shared ranking tests cover index preservation,
ambiguity, optional/variadic signatures and mismatches.

A9 is partially complete: builtin return-type lookup now reads registry signatures,
with explicit intrinsic exceptions. Ordinary call validation consumes those signatures,
replacing 23 separate trigonometric, encoding and date/time analyzers. Specialized AST
and diagnostic handlers remain open work, with their existing messages preserved.
Argument-dependent collection result inference remains on its specialized path. The
241-name return-type audit also corrects shared registry signatures: `Add` returns no
value, and `StrArrayPack` returns an array of strings.

A10 removes `currentContext` and `nodeContext` from the evaluator. All execution helpers
receive context explicitly, and a builtin context adapter binds each invocation to its
own environment, call stack, current node and exception state. Independent-context and
nested builtin tests cover isolation and node restoration. Class-variable compound
assignment also evaluates its RHS with the caller's context.

A11's public labels now identify the bytecode compiler/VM as experimental in command
help and `CompileModeBytecode` GoDoc. Compile help no longer claims a verified speedup.
The decision to retain the VM unmaintained remains unchanged.

Validation: the complete interpreter suite, semantic and builtin suites, evaluator,
public embedding API, bytecode, frontend, parser, lexer, shared types and remaining
packages pass. The fixture regression gate and regenerated status were unchanged by this
architecture work — **871/1,928**, with identical pass-count baselines in all 61
categories and no baseline floor changed by A10/A11. (The anonymous record expression
work recorded below then raised this to **876/1,928** and ratcheted `JSONConnectorPass`
51 → 56; that is the status `TEST_STATUS.md` and `baselines.json` now carry.) The status
generation date was refreshed. `go vet ./internal/... ./pkg/... ./cmd/...`
and the new evaluator context-isolation tests under `-race` pass. Actual compile/run
help output confirms the experimental bytecode labels.

The complete CLI integration suite also passes (10.652s). Its repeated executable
builds exceeded the standard timeout on the workspace's NTFS mount, so it was run
against a verified identical source copy on a native filesystem, using the same test
scripts and `GOFLAGS=-buildvcs=false`. No test assertions were disabled.

## Language compatibility — anonymous record expressions (§3.1)

Closed 2026-09-07, in parallel with the Phase 2 architecture work. `PLAN.md` §3.1 item
"Anonymous record literals → JSONConnectorPass `stringify_anonymous*`, `stringify_record2`".

**The item's premise was wrong, and that is the main finding.** Parenthesised anonymous record
literals — `(x: 10; y: 20)` — already parsed and already had evaluator support
(`internal/parser/expressions.go:399-428`, `internal/parser/record_literals_test.go`). The real
gap was a *different* syntax that the fixtures actually use: DWScript's anonymous record
**constructor expression**, written with `record`/`end` and `:=`, with field names that may be
string literals:

```pascal
PrintLn(JSON.Stringify(record a := s; b := True; g := Null; end));
var r := record "i*i" := i * i; "2i" := 2 * i; end;
'objWithField' := record Field := 123 end;   -- one line, no trailing ';'
```

Before the change this failed at parse time with `Expression expected before RECORD`.

### Design

The two forms are kept apart deliberately. An anonymous `RecordLiteralExpression` means
"needs a record type from context" and is threaded through
`ExecutionContext.SetRecordTypeContext(string)` — the string-keyed context A7 is slated to
retire. The `record … end` form is *structurally* typed: its field names plus the inferred
types of their values fully describe an unnamed `types.RecordType`. It therefore got its own
AST node, `ast.AnonymousRecordExpression`, and never touches the record type context.

Everything downstream already supported unnamed records, so no changes were needed there:
`types.RecordType` allows `Name == ""`, `runtime.RecordValue.Type()` returns `"RECORD"` for it,
and `evaluator/json_serialize.go` already sorts members alphabetically — which is exactly the
key order the expected `.txt` files encode (`{"2i":246,"is":"123"}`).

- `pkg/ast/records.go` — new `AnonymousRecordExpression` node, reusing `FieldInitializer`.
- `internal/parser/record_expressions.go` — new; `RECORD` prefix parse function, registered in
  `parser_builder.go`. Trailing `;` before `end` is optional; a quoted field name is kept
  verbatim so names that are not valid identifiers (`i*i`, `2i`) survive into serialization.
- `internal/semantic/analyze_literals.go` — synthesizes the unnamed `types.RecordType`;
  dispatched from `analyze_expressions.go`.
- `internal/interp/evaluator/anonymous_record.go` — new; builds the `runtime.RecordValue`.

The bytecode compiler needs no new arm: its `default` case already rejects unknown expression
nodes, which is the correct outcome for the unmaintained VM (A11).

### Result

JSONConnectorPass **51 → 56** (62% → 68%); harness overall **871 → 876**. Baseline ratcheted.
Closed: `stringify_record2`, `stringify_anonymous`, `stringify_anonymous3`,
`array_constructor2`, `trueish`. The fixture list in `PLAN.md` understated the yield — it named
two fixtures; six use the syntax.

`stringify_anonymous2` still fails, but for an unrelated reason now recorded as its own §3.2
item: `Random*0` reports `Incompatible operands` because `random` is typed only on the call
path (`internal/semantic/analyze_builtin_functions.go:251`) and not as a bare identifier. That
belongs with A9.

### Note on `pkg/ast/visitor_generated.go`

Re-running the generator produces unrelated drift: it drops `walkRecordTypeNode` (that node
does not embed `BaseNode`, so the current generator skips it), adds `walkGenericTypeRef`, and
reorders `ClassDecl` field traversal. The checked-in file predates those generator changes.
Rather than fold that churn into this change, only the new node's dispatch case and walk
function were added. **The drift is still open** and should be resolved deliberately — dropping
`walkRecordTypeNode` is a real traversal change, not cosmetic.

## Parameterless builtins as bare identifiers, and visitor generator drift (2026-09-07)

Two follow-ups recorded by the previous change, both now closed.

### `Random`, `Now`, `Pi` … used without parentheses

`Random*0` reported `Incompatible operands`. A bare identifier naming a builtin fell through
`analyzeIdentifier` to a blanket `return types.VOID`
(`internal/semantic/analyze_expr_operators.go`), so it carried no usable type into an
expression. Only the *call* path consulted signatures.

The fix follows A9's direction — derive the answer from `builtins.Registry` rather than from
another hand-maintained switch. `Analyzer.parameterlessBuiltinType`
(`internal/semantic/analyze_builtin_registry.go`) returns the signature's `ReturnType`, and
`analyzeIdentifier` uses it before falling back to `VOID`.

The qualifying condition is deliberately narrow — `MinArgs == 0 && MaxArgs == 0 && !IsVariadic`
and a non-nil `ReturnType`. A signature with *optional* parameters says nothing about whether a
bare name means a call or a reference, and procedures keep typing as `VOID`. Of the 14
`Sig(nil, …)` builtins that admits exactly the 13 functions: `Pi`, `Infinity`, `NaN`, `Random`,
`RandSeed`, `Now`, `Date`, `Time`, `UTCDateTime`, `UnixTime`, `UnixTimeMSec`, `GetStackTrace`,
`GetCallStack`. The fourteenth, `Randomize`, stays `VOID` — it is a procedure, so its
`ReturnType` is nil. The existing builtin
function-*pointer* path (`Map`, `Filter`) is checked first and is unaffected.

This also fixes `var t := Now;` and `Now > 0`, which failed the same way.

Result: JSONConnectorPass **56 → 57** (68% → 70%), closing `stringify_anonymous2`; harness
overall **876 → 877**. Baseline ratcheted. A9 remains `[~]` — the specialized dispatch in
`analyze_builtin_functions.go` is still to migrate.

### `pkg/ast/visitor_generated.go` drift

Root cause: `RecordTypeNode` implements `TypeExpression` but does not embed `BaseNode`, so the
generator recognizes it only via the `knownNodeTypes` allowlist in `cmd/gen-visitor/main.go` —
where it was missing, alongside the four sibling type-expression nodes that are listed. Adding
it makes regeneration lossless, and the checked-in file is now genuinely generated: it also
picks up the catch-up the previous change deferred (`walkGenericTypeRef`, `TypeArgs` traversal
on `NewExpression`/`TypeAnnotation`, `FunctionDecl.HelperName`, and `ClassDecl` fields walked
in source order).

Two guards, both confirmed to fail before the fix:

- `cmd/gen-visitor/drift_test.go` — `TestGeneratedVisitorIsUpToDate` regenerates in-process and
  compares against the committed file, so drift in either direction is caught by `go test ./...`
  instead of surfacing as a silently skipped subtree. A companion test asserts every
  `TypeExpression` node that needs the allowlist is on it.
- `pkg/ast/visitor_record_type_test.go` — walks an inline `array of record … end` and asserts
  the record's field declarations are visited.

## Phase 2: runtime classes, resolved semantic types, and builtin signatures (2026-09-08)

### A6: runtime ownership of class metadata

`ClassInfo`, `ClassValue`, `ClassInfoValue`, class metadata mutation, and class operator
storage now live in `internal/interp/runtime`. Interpreter aliases and constructor wrappers
preserve existing callers. The class registry stores `runtime.IClassInfo`, and class creation
uses runtime constructors directly; `ClassInfoFactory` and `ClassValueFactory` are removed.
Registry tests now use runtime classes and cover parent links and metaclass creation without
factory setup. Record, enum, interface, and helper registry typing, plus removal of duplicate
class AST maps, remain open under A6.

### A5: retain the analyzer's type objects across execution

`ast.SemanticInfo` now stores resolved `internal/types.Type` objects for AST nodes alongside
its existing textual annotations. Expression analysis, contextual literal analysis, and type
annotation resolution populate this table. Evaluator annotation consumers use the resolved
objects before legacy lookup, including array elements, declarations, function signatures,
overload selection, and set/array literal annotations. This preserves static bounds, nominal
type identity, and metaclass ancestry without a second parse. Identity tests cover signed array
bounds, enum-indexed arrays, inline function signatures, metaclasses, and contextual literals.
Compatibility annotations may be shared by nodes and retain independent resolved bindings;
clearing the whole semantic table clears those bindings too.

A5 remains open: name-only runtime callers, duplicated declaration type construction, and
explicitly untyped execution still need a migration policy before the old string parsers can
be deleted.

### Unit-aware frontend prerequisite completed

The frontend resolves program-level `uses` through a unit registry, analyzes dependencies
before their importers, and shares semantic metadata between all unit analyzers and the
program. Units export interface declarations, including types, constants, variables, and
functions; implementation dependencies, private declarations, bodies, initialization, and
finalization receive analysis. The CLI's unit-specific type-check bypass is removed.

CLI, embedding API, and fixture execution reuse analyzed unit ASTs. Repeated embedding runs
clone mutable registry state and run initialization/finalization again, without rereading the
unit files. `dwscript.WithUnitSearchPaths(...)` configures search directories. Regression tests
cover dependency order, private/transitive visibility, qualified function calls, errors in
programs and unit bodies, relative includes, cycles, and repeated execution after deleting
source files. This closes the unit-aware frontend prerequisite in §3.2 and the unit declaration
analysis TODO. Existing runtime unit environments remain flattened; semantic analysis enforces
visibility for typed programs.

### A9: shared validation for 78 additional builtin names

Registry signatures now validate 78 further builtin names, including aliases and Print/PrintLn.
The migration deletes 68 redundant analyzer functions. Diagnostic-only styles preserve old
wording and alias display names without redefining arity, parameter types, or return types.
Pre-change characterization tests cover 67 migrated analyzers, invalid arguments and counts, optional
parameters, and exact diagnostic text. Additional tests verify registry ownership and aliases.
Multi-argument diagnostic ordering, strict date/Variant constraints, signature discrepancies,
and AST-dependent/polymorphic rules remain listed under A9.

### Validation and result

The full suite passes with `GOMAXPROCS=2 GOFLAGS=-buildvcs=false go test -p 2 -timeout 20m ./...`.
`just fixture-update` passes and ratchets SimpleScripts **330 → 331**, overall **877 → 878**
(878 / 1,928 scored, rounded to 46%). No category regressed. Workspace build/temp caches were
used because the sandbox's default cache is read-only and `/tmp` lacked space. The first broad
run exceeded the CLI package timeout under host memory pressure; reduced build concurrency
completed it. `TestBinaryIsStale` now sets source timestamps explicitly so its newest-source
assertion does not depend on the temporary filesystem's directory timestamp ordering.

Fixture comparison caught a missed Variant exception in LeftStr; it retains its specialized
handler and now has regression tests for Variant text and Variant aliases. The
`HelpersPass/classname_helper1` stack overflow reproduces on both HEAD and current binaries;
it has no `uses` clause and remains an existing isolated fixture failure.

`go vet -p 2 ./...` and targeted race tests for semantic metadata and repeated unit execution
(`go test -race -p 2 ./pkg/ast ./pkg/dwscript -run 'TestSemanticInfo|TestEngine_Unit' -count=1`)
also pass. Full golangci-lint still reports the repository's existing backlog; the migration's
new documentation, import, test-style, and unit-analysis complexity findings were corrected.


## 2026-09-09 — Phase 2 type and metadata consolidation (A5–A7, A9)

### A5: structural type resolution through execution

Evaluator type resolution consumes the analyzer's resolved type objects or structured
AST annotations and declared names. The inline array, set and function-signature string
parsers are removed. Unchecked execution still skips semantic analysis and supports
nested arrays, inline records, function pointers and metaclasses through the same
structural resolver. Public external-function signature strings remain supported and
are parsed once at registration.

Parser producers now retain compound return, parameter, named-array and operator
annotations. The AST visitor recognizes and traverses ArrayTypeAnnotation. Runtime
class, record, enum and interface registration reuses semantic declaration identity;
alias-aware consumers unwrap aliases only when inspecting their underlying shape.
Identifier lookup distinguishes declared types from function values, preserving builtin
pointers and implicit calls. Record literal and function return contexts carry resolved
types, including nested, by-reference and external-function paths.

### A6: typed registries and canonical runtime callables

Record, enum, interface and helper registries now have concrete runtime entries.
ClassMetadata owns class fields, callable groups, constructors, virtual dispatch,
properties, constants, class variables and operators; ClassInfo no longer duplicates
those entries in AST maps. Class lookup and dispatch pass canonical MethodMetadata
bindings through evaluator-owned execution.

Implementation binding retains callable identity, declaring class, method IDs,
visibility, static/virtual/override flags, defaults, parameter modifiers and contracts.
Captured method pointers observe later implementation binding. Binding copies the
mutable declaration header and parameter slice, leaving compiled source unchanged.
Inherited lookup retains the defining owner's callable. Unused interpreter-side
declaration and operator execution adapters are deleted.

### A7: typed operator signatures and runtime representation dispatch

Operators and implicit conversions retain typed operands and share typed signature
comparison. Exact matches precede compatible matches. Inherited operands rank by
ancestor distance from left to right, preserving the old runtime precedence without
class-name encodings. Conversion chains use deterministic shortest paths and handle
cycles and depth limits. Assignment attempts conversion between distinct named record
types even though they share a runtime representation.

ValueKind supplies representation dispatch, while LanguageType supplies language type
identity. Value.Type() remains available for diagnostics without adding a required
method to the public value contract. Runtime and evaluator comparisons no longer use
its display strings as dispatch keys; an architecture test guards that boundary.
Runtime metadata drops duplicate type-name fields, and ExecutionContext owns a single
resolved record context and function return type.

### A9: builtin signature constraints

Builtin registry signatures now cover strict numeric/date/Variant/JSONVariant rules,
string-or-Variant arguments, array elements, ordinal conversion and diagnostic ordering.
Trim, RandG, StringReplace, ToJSONFormatted and array-returning JSON/string signatures
agree with runtime behavior. AST-dependent, by-reference and polymorphic intrinsics
remain explicit. FloatToStrF remains semantic-only because it has no runtime callable.
The shared analyzer replaces 111 additional builtin names (including aliases) and
removes about 2,600 lines of specialized analyzer and dispatch code. Golden compatibility
tests retain 8,176 pre-change argument patterns across 112 names, including the retained
StringOfChar intrinsic, with exact diagnostic and result-type expectations.

### Validation

The full fixture gate passes: **878 / 1,928 scored** (1,050 failing, 114 unscored).
The freshly rebuilt CLI produces the same result, and all 61 category counts match the
pre-refactor floors. The known isolated `HelpersPass/classname_helper1` recursion failure
still reproduces in the baseline and remains outside this refactor.

The final comparison caught inferred function signatures being mistaken for explicit
pointer contexts. Callable-context detection now retains the analyzer's annotation-presence
marker while consuming resolved identities; regression tests cover bare calls, default
arguments, explicit pointer bindings and record-returning implicit-call receivers.

CLI measurement used `just fixture-report --allow-stale`: the recipe rebuilt the binary
immediately before measurement, and the flag bypassed the stale guard's treatment of
tracked source files deleted by this uncommitted refactor. Build and test commands used
`GOMAXPROCS=2`, `GOFLAGS=-buildvcs=false`, and workspace build/temp caches because the default
cache is read-only and `/tmp` has limited space.

`just fixture-update` passes and leaves every baseline count unchanged; TEST_STATUS.md
records the refreshed date. `go vet -p 2 ./...` and targeted race tests pass. Race coverage
includes semantic metadata, repeated unit execution, structured types, external nested-array
contexts, aliased/implicit callable paths and canonical class binding.

Changed-source lint reports zero issues, and the unfiltered scan has no findings in
new files. Full lint still reports the repository's backlog (1,199 findings in the final
scan); newly orphaned helpers and newly introduced formatting, alignment, assertion and
complexity issues were removed or corrected. Scoped `git diff --check` passes.

The complete suite passes with `go test -p 2 -timeout 20m ./...`, including CLI integration,
visitor-generation drift checks and the fixture gate. The final CLI package took 332 seconds
under the constrained build concurrency; the interpreter package took 94 seconds. A5–A7 and
A9 are closed in PLAN.md. A11 remains explicitly deferred pending an owner decision.

## 2026-09-09 — Parameterless callbacks in `and` / `or` (3.2)

The semantic analyzer now checks parameterless function and method pointer operands
using their return types for `and` and `or`. The evaluator invokes callbacks through
the existing value-context helper, preserving left-to-right evaluation, Boolean
short-circuiting, and eager Integer/enum bitwise operations. Operand errors and pending
exceptions stop evaluation immediately. Variant results retain existing coercion rules.

Regression scripts in `testdata/function_pointer_operators/` exercise the shared compile
pipeline and production evaluator: named functions, pointer variables, lambdas, bound
methods, call counts/order, bitwise and Variant results, and skipped/invoked nil or raising
callbacks. Negative tests cover parameter-taking callbacks, incompatible result types,
and existing pointer-comparison diagnostics. Pointer assignment and `implies` also retain
their behavior. This closes the `and`/`or` item in PLAN.md §3.2 without changing parser,
AST, shared type definitions, `xor`, or builtin-pointer invocation policy.

Targeted tests pass with `go test ./internal/interp -run
'^TestFunctionPointerOperators_|^TestImplies|^TestFunctionPointerValueContextAutoInvoke$'
-count=1`. The CLI reproduction now prints `True` twice for `callback and True` and
`False or callback`, where both expressions previously failed with incompatible operands.
The lambda guide documents the supported value contexts and evaluation order.

The fixture comparison is unchanged across all 61 categories: 885 passed, 1,043 failed,
114 skipped (1,928 scored). This starting workspace already included concurrent property
work, so the difference from the earlier 878-pass headline is not attributed to this fix.
Changed-source lint reports zero issues, and scoped `git diff --check` passes.

`just fixture-update` passes and refreshes the generated status. It ratchets the existing
concurrent gains in JSONConnectorPass (57 → 59) and PropertyExpressionsPass (10 → 15);
all other floors stay unchanged. Validation uses workspace Go/build/temp caches because
the default Go cache is read-only and `/tmp` has limited free space.

`go test -p 2 -timeout 20m ./internal/semantic ./internal/interp/...` passes, including
the evaluator and runtime packages. The interpreter package completed in 87 seconds.

## 2026-09-09 — Semantic backlog refinement (planning complete)

PLAN.md §3.2 now divides the remaining semantic work into 24 tasks across class
construction, diagnostics/metaclass properties, contracts, generics, overloads, sets,
and conditional compilation. Each task has a stable ID and acceptance target; sequencing
and shared-file coordination identify work that can proceed alongside §3.1.

The class-builder tasks extend the existing class predeclaration mechanism rather than
assuming it is absent. The conditional-compilation tasks require reproducing their
blockers before implementation because the previous ArrayPass/SetOfPass attribution
was not supported by a source search. An explicit Done summary in §3.2 links the completed
`and`/`or` callback work to its implementation and validation record above. The remaining
language tasks stay open. This refinement changes documentation only; test execution
remains with the user.

## 2026-09-09 — Parser gaps closed, PLAN.md §3.1 emptied (3.1)

Six items were listed under §3.1. Measuring each against the built CLI first changed the
shape of the work: `array of T` in getter position already worked and was deleted as
already-shipped, and `{$I %FILE%}` was reclassified won't-fix because its only fixture,
SimpleScripts `include_expr`, encodes the original Delphi runner's paths
(`Test\include_expr.pas`, `*MainModule*`) that neither scoring path normalizes, and
`%FUNCTION%` is not knowable at lex time. The remaining four are closed here.

**`class property` in records and auto-property backing fields.** `parseRecordBody` rejected
`class property`; `parseRecordPropertyDeclaration` had no auto-property desugaring, so a
record `property Field: Integer;` parsed into a permanently unreadable and unwritable
property. Both now mirror the class side: the parser points a bare property at `F<Name>`
and `addRecordAutoPropertyBackingField` synthesizes the field, or the class var for a class
property. `RecordPropertyDecl` and `RecordPropertyInfo` carry `IsClassProperty`, and record
class-property writes reach `RecordTypeValue.ClassVars`, matching the read fallback that
already existed in `readRecordTypePropertyValue`.

Three runtime defects surfaced underneath and had to be fixed for the fixtures to pass.
Writing a class property through an instance (`obj.ClassProp := v`) created an instance
field shadowing the class var, so the write was silently lost while the read still saw the
old value; `executePropertyWrite` now delegates class properties to the existing
`evalClassPropertyWrite`, so instance and metaclass spellings share one implementation.
The same shadowing hit plain class vars in `member_assignment.go` and in `EvaluateLValue`,
where `obj.ClassVar.Field := v` could not even resolve its base. And class properties are
not virtual: `TBase(sub).ClassProp` must read TBase's declaration, so `staticClassPropertyOf`
resolves them against the cast's static type on both the read and the write path, alongside
the field-shadowing rule that was already there.

**Expression-backed and multi-index indexed properties.** The semantic analyzer rejected
expression accessors on indexed properties up front, even though `bindPropertyIndexParams`
already bound the index parameters into the accessor scope — the code was unreachable behind
the guard. Removing both guards, carrying the declared index parameter names and types on
`PropertyInfo`, and binding those names in `executeIndexedPropertyExpressionRead` and
`tryIndexedPropertyExpressionWrite` completes the feature; the read-only check in
`index_assignment.go` no longer treats an empty `WriteSpec` as read-only, since an
expression setter legitimately has none. Separately, multi-index properties did not work at
all, with any accessor kind: `a.Prop[i, j]` parses as the chain `a.Prop[i][j]`, which fell
through to ordinary indexing and reported `Array expected`.
`analyzeMultiIndexPropertyAccess` now collects the chain and resolves it against the single
declaration when its length matches the declared arity.

**`external` on properties.** Parsed for both record and class properties following the
existing external-function pattern, carried to `RecordPropertyInfo`/`PropertyInfo`, and used
as the emitted key in JSON serialization. Members are sorted by the emitted name, so
`property Test2 : Integer external 'hello'` yields `{"Test":123,"hello":123}`.

**Nested `>>` in generic type-argument lists.** No fixture exercises this, but
`var x : TA<TA<Integer>>;` failed. The lexer emits `>>` as one token and there was no
token-splitting mechanism to reuse. `TokenCursor.SplitGreaterGreater` rewrites the buffered
token into the single remaining `>` in place — one token in, one token out — which keeps
every other cursor's index valid where an insertion would not, since cursors share the token
buffer. `parseTypeArguments` takes the first `>` without advancing, leaving the second for
the enclosing list, and `looksLikeGenericTypeRef` counts `>>` as two closers. `>>` is not a
shift operator in DWScript (`shr` is), so no operator behavior changes. Triple nesting and
expression position (`TA<TA<Integer>>.Make`) both work.

Two diagnostic bugs blocked the fixtures once parsing succeeded, both only visible under
`--hints pedantic`. Class vars were bound into property-expression scopes under their
normalized map key, so a use spelled exactly as declared produced
`Hint: "Field" does not match case of declaration ("field")` — the analyzer quoting its own
lowercasing. `ClassVarDeclNames` now records the declared casing (`Fields` already kept it,
which is why only class vars were affected) and the binders use it. And a private method or
field named only as a property accessor was reported as never used, because accessor
resolution never marked it; `validateReadSpec`/`validateWriteSpec` now record the usage.

The CLI fixture comparison goes from 878 to 885 passes, exactly the seven targeted fixtures
with no category regressions: PropertyExpressionsPass `class_property_expressions`,
`class_property_write_expressions`, `property_auto_field`, `indexed_expressions`,
`indexed_write_expressions`, and JSONConnectorPass `property_name`,
`stringify_class_getter`. `indexed_write_expressions` was not listed in PLAN.md but is the
same feature. PropertyExpressionsPass `read_write_other_property` remains failing: it needs
a case-mismatch hint, which is won't-fix per §5.

`go test ./internal/... ./pkg/...` passes, including the interpreter, evaluator, semantic and
parser packages. Validation ran with `TMPDIR` pointed at a root-filesystem directory because
`/tmp` is a 7.3 GB tmpfs that was 97% full.

## 2026-09-09 — Two-phase class construction: inheritance before members (L-S1a)

`PLAN.md` §3.2.1 L-S1a. The analyzer already predeclared one `*types.ClassType` shell per
top-level class name; that mechanism is now split into two explicit phases in the new
`internal/semantic/class_construction.go`:

1. **Identity** — unchanged: every top-level `ClassDecl` (recursing into blocks) registers a
   single shell in the one type registry. There is still exactly one type representation per
   class; no second registry was introduced.
2. **Inheritance** — new `resolveTopLevelClassInheritance`. Declarations are grouped by
   normalized full name (so an explicit `forward` and its implementation, or the parts of a
   partial class, form one group). For each group it propagates the declaration-level shape
   flags (`IsAbstract`, `IsExternal`, `ExternalName`, `IsStaticClass`, `IsDeprecated`,
   `DeprecatedMessage`) onto the shell, then links parents with a memoized DFS that resolves
   ancestors before descendants.

The phase is deliberately silent. Unknown parents, cycles, self-inheritance and
forward/partial parent disagreements are left *unlinked* so that `analyzeClassDecl` still
emits the existing diagnostics at the existing positions; the DFS carries an in-progress
marker and a length-bounded `reaches` walk so a cyclic declaration set can never recurse or
spin during the phase itself. `IsForward` and `IsPartial` stay owned by `analyzeClassDecl`.

The `class(TParent, IFoo)` disambiguation (a leading interface entry that actually names a
class) is applied in the phase with the same AST rewrite `resolveParentClass` performs, so
the later interface-implementation check still sees the right list.

**Observable fix:** class-level inheritance validation is no longer order-dependent.
`type TChild = class(TExt) end; type TExt = class external end;` was silently accepted; it now
reports `non-external class 'TChild' cannot inherit from external class 'TExt'`, matching the
diagnostic already produced when the parent is declared first.

**Validation:** `go test ./...` green; `just fixture-report` **885 / 2,042 before and after,
byte-identical per-category table** (this task is a correctness/ordering fix, not a fixture
win). New table-driven tests in `internal/semantic/class_construction_test.go` cover
child-before-parent field/method/grandparent inheritance, unknown parents, 2-class, 3-class
and self cycles, external-parent order independence, and forward/partial preservation. Run
against the pre-change tree, exactly one case fails (external parent declared after the
child), confirming the rest are regression guards.

**Still order-dependent, and left to L-S1b/L-S1c:** member-level checks that read the
parent's members at the child's declaration site — `override` validation
(`checkMethodOverriding`) and `inherited` resolution still fail when the parent is declared
later, e.g.
`type TChild = class(TParent) procedure Hello; override; ... end; type TParent = class procedure Hello; virtual; ... end;`.
Also noted while probing: `sealed` is not enforced as an inheritance restriction in either
order, and `type TFoo = class(TBase);` is by design a complete empty subclass rather than a
forward declaration (`internal/parser/classes.go`), so `validateForwardDeclParent`'s
"different parent" branch is unreachable for top-level classes.

## 2026-09-09 — Class member signatures complete before body checking (L-S1b)

`PLAN.md` §3.2.1 L-S1b, building directly on L-S1a. Phases 1 and 2 give a class its *identity*
and its *ancestry* independent of source order, which is already enough for a field, parameter,
property or return type to name a class declared later in the file. It is not enough for a
method *body*, which needs the members of the classes it touches.

**Measurement first.** Of the four stated acceptance cases, three already passed at
`7a44cf97`: a field typed by a later-declared class, two classes with mutually referring
fields, and a property/method signature typed by a later-declared class all compiled and ran
cleanly, and type identity was already consistent (assignment in both directions and method
dispatch through the field both worked), because phase 1 registers one shared shell that later
gets mutated in place. What did *not* work was any inline method body, in either direction:

```pascal
type TA = class
  B: TB;
  procedure Go; begin B.Hello; end;   // "There is no accessible member with name Hello for type TB"
end;
type TB = class procedure Hello; begin PrintLn('hi'); end; end;
```

and, within a single class, a body referring to a member declared further down — a method
(`Unknown name "B"`), or a property (`Unknown name "V"`), because `analyzeClassDecl` registers
properties only after the method loop has already checked every inline body.

**Phase 3 (member signatures before bodies).** `analyzeMethodDecl` is split. Everything that
must stay in source order — parameter and return type resolution, constructor detection,
overload and forward matching, visibility and virtual/override metadata,
`validateVirtualOverride` — still runs where the method is declared. The body no longer does:
it is captured in a `deferredMethodBody` (the method, its class, the enclosing symbol table,
the nested-type aliases, the resolved parameter/return types and the `inUnitDecl` flag) and
queued. The new `checkMethodBody` later restores exactly that declaration-site environment,
builds the method scope from scratch and analyzes the block, so parent fields and class vars
are read at drain time rather than at declaration time.

**Where the queue drains matters.** Draining at the end of the declaration pass — the obvious
choice, mirroring the existing top-level-function two-pass — cost two fixtures for two
different reasons, both ordering artifacts rather than real errors:
`FailureScripts/abstract_method` moved a body hint after the errors from the executable
statements below the type section, and `SimpleScripts/class_init` let a global `var b`
declared *after* the type section shadow the class constant `B` in a deferred body (class
constants and properties are consulted only after `symbols.Resolve` fails). Draining right
after the *last top-level statement that declares a class* fixes both: every class is complete,
no later global is in scope yet, and body diagnostics stay in source order relative to the code
that follows. `lastTopLevelClassDeclIndex` mirrors the statement shapes phases 1/2 already walk.
Bodies of local classes declared inside a function body are still checked immediately.

**Validation:** `go test -p 2 -timeout 40m ./...` fully green. `just fixture-report` **885 →
887 / 2,042**, no category regressed; `SimpleScripts/method_implem` (property, class const and
class function all declared after the inline bodies that use them) and
`SimpleScripts/var_param_obj_method` (a write-only property declared after the method that
assigns it) newly pass, and `baselines.json` / `TEST_STATUS.md` are ratcheted. New table-driven
`TestClassConstruction_MemberSignaturesBeforeBodies` in
`internal/semantic/class_construction_test.go` covers all four acceptance cases plus the
same-class ordering cases, a negative case (a genuinely missing member of a later class is
still reported) and the global-shadowing regression; run against the pre-change tree exactly
four of its cases fail, so the rest are regression guards. `golangci-lint run` reports no new
findings: `analyzeMethodDecl`'s cyclomatic complexity drops 58 → 40 and the extracted
`checkMethodBody` / `defineMethodScopeMembers` stay under the threshold.

**Not fixed, and left to L-S1c:** out-of-line implementations (`procedure TFoo.P;` at top
level) are still analyzed in source order in the declaration pass — they are written after the
type section anyway — and the `override`/`inherited` limitation L-S1a recorded is untouched
by design: `checkMethodOverriding` and `validateVirtualOverride` still read the parent's
members at the child's declaration site, so
`type TC = class(TA) procedure Go; override; ... end; type TA = class procedure Go; virtual; ... end;`
still reports `method 'Go' marked as override, but no such method exists in parent class`.
