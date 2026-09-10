# Progress log — September 2026

Closed 2026-09-06 on branch `feat/phase1-measurement-tooling`. Items T1–T6 and A1 of the
2026-09-06 `PLAN.md`. Numbers are from the runs recorded in the commit messages.

### Headline

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

## 2026-09-09 — Ancestor-dependent class validation after signatures complete (L-S1c)

`PLAN.md` §3.2.1 L-S1c, closing the section. Phases 1–3 made class identity, ancestry and
member signatures order-independent; the checks that *compare* a class against its ancestors
still ran at the child's declaration site and so read a half-built parent.

**Measurement first.** Of the stated acceptance cases, the out-of-line one already worked at
`c8cfbc4a`: `type TFoo = class procedure P; end; procedure TFoo.P; begin PrintLn(TBar.Value);
end; type TBar = class const Value = 42; end;` compiled and printed `42`, because out-of-line
implementations are analyzed after the type section anyway. What failed were the two inline
cases, both with the same message:

```pascal
type TC = class(TA) procedure Go; override; begin PrintLn(2); end; end;
type TA = class procedure Go; virtual; begin PrintLn(1); end; end;
var c := TC.Create; c.Go;
// method 'Go' marked as override, but no such method exists in parent class
```

and the same shape with `inherited Go;` in the body. Both now compile and print `2` / `1 2`.

**Phase 4 (validation after signatures).** `internal/semantic/class_construction.go` gains a
fourth phase that postpones the four ancestor-dependent validations —
`validateVirtualOverride` per method, and `checkMethodOverriding`,
`validateInterfaceImplementation`, `validateAbstractClass` as the class-declaration tail — into
a `deferredClassChecks` queue drained at the existing phase-3 drain point (right after the last
top-level class declaration) and *before* the deferred bodies, so a signature-level diagnostic
still precedes the body diagnostics of the same program.

The deferral is conditional, which is what keeps existing diagnostics where they are.
`noteTopLevelClassDecls` records how many top-level declarations contribute to each class
(populated from the phase-2 grouping, so forward + implementation and the parts of a partial
class count together); `markClassDeclAnalyzed` — deferred at the top of `analyzeClassDecl`, so
it also runs on the early diagnostic returns — decrements it. `classAncestorsPending` walks the
linked parent chain and reports whether any ancestor still has declarations outstanding. Only
then is a check queued; when every ancestor is already complete, it runs exactly where it
always did, with the same position and the same ordering. In practice this means only programs
that today produce the bogus error change behavior. No second type registry: the single shared
`*types.ClassType` shell is still mutated in place, and the queue stores pointers to it.

**Diagnostics preserved.** New table-driven cases in
`internal/semantic/class_construction_test.go`
(`TestClassConstruction_ValidationAfterSignaturesComplete`, 15 subtests) pin both directions:
override / `inherited` / overriding constructor / out-of-line override / override through a
grandparent declared last all accepted; and, with the parent declared *last*, `override` with
no such parent method, `override` with a mismatched signature, `override` of a non-virtual
parent method, hiding a virtual parent method without `override`, a duplicate member, a
declared-but-unimplemented method, an unimplemented inherited abstract method and a missing
interface implementation all still diagnosed with their existing messages. Executable
statements around a type section keep source order.

**Validation:** `go test -p 2 -timeout 40m ./...` green. `just fixture-report` **887 / 2,042
before and after, with a byte-identical failing-fixture list** (`--list-fails` diffed both
ways) — this is a correctness/ordering fix, not a fixture unlock, and `baselines.json` needed
no ratchet. `golangci-lint run`: 1,204 issues before and after; the only delta is
`analyzeClassDecl`'s cyclomatic complexity dropping 56 → 54.

**Known limitation, not fixed here.** A *statement* that instantiates a class is still analyzed
in source order, so `type TC = class(TA) end; var c := TC.Create; type TA = class abstract
procedure Go; virtual; abstract; end;` does not report "Trying to create an instance of an
abstract class" — the statement is checked before `TA` exists. This is unchanged from
`c8cfbc4a` (verified against the pre-change binary) and is inherent to executable statements
keeping source order, which L-S1c's acceptance criteria require; fixing it would mean deferring
statement analysis, a much larger change.

## 2026-09-09 — Unused-private-field hint: usage tracking (L-S2a)

`PLAN.md` §3.2.2 L-S2a. The section instructs confirming each failure before implementing, and
here that mattered: **the ticket's premise did not reproduce.**
`JSONConnectorPass/serialize_class` passes cleanly under `--hints pedantic`, and none of the 23
currently-failing JSONConnectorPass fixtures involves an unused-private-field hint at all —
they fail on record copy-on-assign value semantics, missing connector APIs, VARIANT member
access and runtime-message wording. `stringify_record`, the closest candidate, differs on
`{"BottomRight":{"x":1,"y":3}}` vs `…"y":2`, which is the record-assignment item already
tracked separately in `PLAN.md`. The stated acceptance fixture is retired.

**What actually reproduced** is a false-positive class, found by sweeping every fixture at
pedantic level and diffing the emitted `Private field` hints against the expected `.txt`:

```pascal
type TA = class
   private FA : Integer;   // Self.FA       -> correctly marked used
   private FB : Integer;   // FB := 1       -> HINTED (wrong)
   private FC : Integer;   // PrintLn(FC)   -> HINTED (wrong)
   private FD : Integer;
   public property Q : Integer read (FD * 2);   // -> HINTED (wrong)
   public procedure Touch;
   begin Self.FA := 0; FB := 1; PrintLn(FC); end;
end;
```

**Root cause.** `defineMethodScopeMembers` (`internal/semantic/analyze_classes_decl.go`) and
`bindClassPropertyExprScope` (`internal/semantic/analyze_properties.go`) inject a class's own
fields into the scope as ordinary symbols so a method body or a property accessor can name them
without `Self.`. A bare identifier therefore resolves against the symbol table and never reaches
the `a.currentClass.GetField(...)` fallbacks in `analyze_expr_operators.go` /
`analyze_statements.go` — the only places that called `recordClassFieldUsage` for implicit-`Self`
access. Qualified `obj.Field` and property read/write *specifiers* were marked correctly, which
is why `FailureScripts/class_unused_privates` kept passing and hid the gap.

**Fix.** The binding itself now carries the attribution rather than a side table:
`SymbolTable.DefineClassField` sets `Symbol.ClassFieldOwner` to the declaring class, and the two
symbol-resolution sites call `recordResolvedSymbolFieldUsage(sym)`, which forwards to the
existing `recordClassFieldUsage`. Putting it on the symbol rather than in a name-keyed map is
what makes shadowing correct for free: a local declared as `FValue` inside the method replaces
the binding with an ordinary symbol whose `ClassFieldOwner` is nil, so it does not mark the
field used. Inherited fields need nothing — `addParentFieldsToScope` deliberately omits private
parent fields, and only private fields are ever hinted.

**Tests.** `internal/semantic/unused_private_field_test.go` is new; the message previously had
*zero* Go test coverage, only fixtures. Ten table-driven cases cover qualified read, bare read
inline and out-of-line, bare and compound assignment, expression-form read and write accessors,
identifier specifiers, a genuinely unused field that must still be hinted, and the shadowing
case; plus sibling fields tracked independently and the hint staying pedantic-only.

**Validation:** `go test ./internal/... ./pkg/...` green. The full-tree sweep for
`Private field` hints not present in the expected output went from **six**
(`Algorithms/bottles_of_beer`, `GenericsPass/tlist1`, `InnerClassesPass/inner_implem`,
`SetOfPass/enum_property`, `SimpleScripts/free_destroy`, `SimpleScripts/partial_class3`) to
**zero**. `just fixture-report` **888 → 892 / 2,042** (measured on `894e387c`, with the change
stashed and unstashed): GenericsPass 14 → 15, InnerClassesPass 0 → 1, SetOfPass 20 → 21,
SimpleScripts 334 → 335, no category down. The two that did not
convert are unrelated: `Algorithms` runs at `normal` hints in the harness, and
`SimpleScripts/partial_class3` still differs on partial-class redeclaration handling.
`baselines.json` and `TEST_STATUS.md` ratcheted with `just fixture-update`.

## 2026-09-09 — Helper property expression accessors and metaclass resolution (L-S2b)

`PLAN.md` §3.2.2 L-S2b. `PropertyExpressionsPass/helpers_property_expressions` failed with
`Runtime Error: property 'MultBy2' has no read access`. Minimising it split the ticket's single
line into two independent gaps, and chasing the fixture to green surfaced three more.

**1. No expression case in the helper accessors.** `executeHelperPropertyRead` /
`executeHelperPropertyWrite` (`internal/interp/evaluator/helper_methods.go`) switched on
`ReadKind`/`WriteKind` and handled `PropAccessField`, `PropAccessMethod`, `PropAccessBuiltin` and
`PropAccessNone` — but not `types.PropAccessExpression`, so every expression-form helper accessor
fell into `default:`. The metadata was already there: the helper property converter in
`visitor_declarations.go` sets `ReadKind`, `ReadExpr`, `WriteExpr` and `IsClassProperty`. This was
never class-property-specific — a plain `property M : Integer read (2*Field)` on an instance
helper failed identically.

The new cases delegate to the accessor scope the receiver deserves rather than defining a
helper-only one: a `class property` resolves the extended type's class metadata
(`helperReceiverClassInfo`) and reuses `evalClassPropertyExpressionRead` / `…Write`, a record
receiver gets fields plus class state the way a record method body does, and everything else
takes the existing object-shaped `executeExpressionBackedPropertyRead` / `…Write`.

**2. Helper properties invisible through a metaclass.** `resolveClassMetaMember`
(`visitor_expressions_members.go`) consulted helper *methods* but never helper *properties*, so
`TBase.HelperClassProp` reported `member 'MultBy2' not found in class 'TBase'`;
`member_assignment.go` had the same hole on the write side. Both now look up
`FindHelperProperty` before erroring, and both restrict it to `IsClassProperty` — an instance
property declared in a helper still needs an instance receiver.

**3. Helper class properties through a type cast.** `TBase(FSub).MultBy2` must bind the *cast's*
static class, exactly like the field and class-property lookups beside it. Both the read path and
`member_assignment` now resolve the static metaclass (`staticClassMetaOf`) and look the helper
property up against it, falling back to the wrapped receiver for instance properties.

**4. Lvalue write specifiers.** `write (FBase.MultBy2)` — no `:=` — is shorthand for
`write (FBase.MultBy2 := Value)`. The parser collapses `write (Field)` to a plain identifier, so
only the non-identifier form reached the converters, where both the class and the helper path
silently set `PropAccessNone` and dropped the setter. `writeSpecAssignment` now synthesizes the
assignment, storing it in the same expression form an explicit write statement uses.

**5. Record class vars written through an instance.** `FBase.Field := v`, where `Field` is a
record `class var`, fell through to the instance field setter, which created a field of that name
and shadowed the shared slot — the write was silently lost. The record branch of member
assignment now writes the `RecordTypeValue`'s class-var storage first, mirroring the rule the
object branch already had for classes. The helper-property lookup is likewise placed before the
field setter on both branches, for the same reason.

**Semantic alignment.** `analyzeHelperProperty` (`internal/semantic/analyze_helpers.go`) stored
only `{Name, Type}`, leaving every helper property at `ReadKind == PropAccessNone` and disagreeing
with the evaluator about the same AST. It now records the same read/write kinds, specs,
`IsIndexed`, `IsDefault` and `IsClassProperty`. The accessor expression itself is deliberately
*not* analyzed here — that can surface new diagnostics across unrelated fixtures and belongs with
the §3.1 property work.

**Validation:** `go test ./internal/... ./pkg/...` green;
`internal/interp/helper_property_expressions_test.go` is new (10 subtests over instance/class
helpers, record helpers, instance/class-name/cast receivers, the lvalue shorthand, and the record
class-var rule). `just fixture-report` **892 → 895 / 2,042**, the whole delta in
PropertyExpressionsPass 15 → 18 — `helpers_property_expressions`,
`class_helpers_property_write_expressions` and `record_helpers_property_write_expressions` — with
no category down. `golangci-lint run --new-from-rev`: 0 issues. Baselines ratcheted.

**Left open, recorded in `PLAN.md`.** `read_write_other_property` (a property whose specifier
names another property) is a different gap, and record-type metaclass access (`TRec.ClassProp`
through the type name rather than an instance) remains unsupported; no fixture demands it.

## 2026-09-09 — Indexed properties with class-method accessors (L-S2c)

`PLAN.md` §3.2.2 L-S2c, closing the section. `SimpleScripts/enum_to_integer` failed with
`Runtime Error: member 'Prop' not found in class 'TConvert'` on `TConvert.Prop[eGamma]`, where
`Prop` is an ordinary indexed property whose getter happens to be a `class function`. Minimising
it showed the ticket's framing was half the story: **the same read through an instance failed
too**, with `indexed property 'Prop' getter method 'Get' not found`.

**Root causes.** `executeIndexedPropertyGetterMethod`
(`internal/interp/evaluator/property_read.go`) resolved the accessor with
`objVal.GetMethodDecl`, which walks only the instance method table —
`ObjectInstance.GetClassMethodDecl` already existed for precisely this case (DWScript permits
calling a class method through an instance) and was simply not consulted. Separately,
`VisitIndexExpression` (`visitor_expressions_indexing.go`) matched three receiver shapes for
`obj.Prop[i]` — interface instance, object, record — and had **no metaclass branch**, so
`TConvert.Prop[…]` fell through to plain member access and died in `resolveClassMetaMember`.
Neither `evalClassPropertyRead` nor `ReadClassProperty` could have helped: the first rejects
indexed properties outright, the second rejects anything with `!IsClassProperty`, and `Prop` is
an instance property.

**Fix.** The accessor lookup now falls back to the class method table and, when it lands there,
invokes the accessor with the metaclass as receiver (`classSelfForInstance` for an instance
receiver) so `Self` and `ClassName` resolve to the class. `evalClassMetaIndexedProperty` adds the
missing receiver branch: it resolves the property from the class info, evaluates and arity-checks
the indices against `PropertyInfo.IndexParamTypes` — the authoritative arity, available for
expression accessors too — and dispatches to the class method or, for `PropAccessExpression`, to
the existing `executeIndexedPropertyExpressionRead`. It reports *unhandled* rather than erroring
when the class declares no such indexed property, so ordinary member access still produces its
usual not-found diagnostic.

**The write side mirrors it**, though no fixture demands it: `index_assignment.go` gained the same
`GetMethodDecl` → `GetClassMethodDecl` fallback on both the named and default-property setter
paths, plus `evalClassMetaIndexedPropertyWrite`. Leaving it out would have made
`TC.Prop[i]` readable but not writable.

**Semantic tightened to match.** `analyzeIndexedPropertyAccess`
(`internal/semantic/analyze_arrays.go`) unwrapped `*types.ClassOfType` and returned the property
type for *any* indexed property, bypassing the metaclass restriction the plain member-access path
enforces — semantic and runtime disagreed about which side supported this.
`checkIndexedPropertyMetaclassAccess` now applies the same rule with the same messages
(`Read access of property should be a static method` + `Class method or constructor expected` for
an instance-method accessor, `Object reference needed` for field/expression accessors). The write
counterpart, `checkIndexedPropertyWriteTarget`, also unblocks the legal case: the assignment path
used to call `analyzeExpression(target.Left)` on the bare `TC.Prop`, which is not a readable
expression, and rejected the whole statement. That standalone analysis is now skipped for an
indexed-property base, and the write check returns early when it diagnoses so the read-side check
does not report the same problem twice.

**Validation:** `go test ./internal/... ./pkg/...` green;
`internal/interp/indexed_property_class_accessor_test.go` is new — five execution subtests
(read/write through the class name and through an instance, in both combinations, plus an
expression accessor through the class name) and two diagnostic subtests pinning the rejected
instance-method getter and setter. `just fixture-report` **895 → 896 / 2,042**,
`SimpleScripts/enum_to_integer`, no category down. `golangci-lint run --new-from-rev`: 0 issues.
Baselines ratcheted. §3.2.2 is now closed.

## 2026-09-09 — Contract inheritance and inline-method naming (L-S3a, L-S3b, L-S3c)

Closes `PLAN.md` §3.2.3. Harness **896 → 898 / 1,928** (46% → 47%), SimpleScripts 336 → 338;
`method_contracts` and `method_condition` newly pass, no category down.

**The shape of the bug.** Contracts were read straight off the executing `*ast.FunctionDecl` at
both call sites (`ExecuteUserFunction`, `invokeParameterlessUserFunction`). Two consequences:

1. An override with no conditions of its own ran none. `TSubChild.Check(-1)` printed
   `subchild -1` where DWScript raises `Pre-condition failed in TBase.Check`.
2. The failure name came from `fn.ClassName`, which the parser sets only for out-of-line headers
   (`procedure TFoo.Bar`). A method whose body is written inline in the class declaration lost
   the class prefix entirely — `Pre-condition failed in TestPre` instead of `TTest.TestPre`.

**The fix is one resolver.** `internal/interp/evaluator/contract_inheritance.go` builds a
*contract chain*: the executing declaration first, then each distinct ancestor declaration of the
same method, base-most last. It is derived from what already existed rather than new registration
plumbing — `currentMethodClassName(ctx)` (the `__CurrentMethodClass__` binding
`executeMethodWithClassInfo` already sets) names the class, `TypeSystem.LookupClass` resolves it,
`OwnsMethodDecl` confirms ownership, and `IClassInfo.LookupMethod` walks the parents.

The ownership check is what keeps this honest. A free function called from inside a method body
still sees the caller's class binding in its environment chain; because no class in the chain
declares it, it resolves to an empty chain and keeps its bare name and its own conditions.
`runtime.MethodMetadata` already carries unread `PreConditions`/`PostConditions` fields; they
stayed unread — resolving from the AST plus the class chain keeps the change inside one package.

**Ordering is observable, and the fixture settles it.** Postconditions run derived-first: with
`TBase.Check` ensuring `i < 10` and `TSubChild.Check` ensuring `i < 5`, `s.Check(10)` fails both,
and the expected output names `TSubChild.Check`. Ancestor-first would print `TBase.Check`.
Preconditions run root-first; upstream permits `require` on the root method only, so at most one
source contributes in practice, but the order is now defined rather than accidental.

**Two details that would have been silent bugs.** `old` capture ran over the executing
declaration's postconditions only, so an inherited `ensure Result = old i + 1` had nothing
captured by the time the body had run; it now walks the same chain. And an ancestor condition
names the *ancestor's* parameters — contract parameters bind by position, so when an override
renames them the ancestor's names are aliased to the call's argument values for the duration of
the check (`parameterAliases`). Chains are memoized per (declaration, declaring class); the same
declaration is shared with every descendant that does not override it, so the declaration pointer
alone is not a sufficient key.

**Removed.** The exported `CheckPreconditions` / `CheckPostconditions` / `CaptureOldValues`
wrappers and the single-declaration `captureOldValues` they fronted had no callers left.

**Not done, deliberately.** The `Preconditions must be defined in the root method only`
diagnostic belongs to §4 error-detection parity (`FailureScripts/contracts_precondition` also
needs `Warning: Constant condition`). The call-stack frame name in `ExecuteUserFunction` has the
same missing-qualifier bug (`TTest.TestMeth` vs `TestMeth`, visible in
`SimpleScripts/contracts_subproc`), but that fixture also needs the raise-site column work listed
in §3.3 and would not flip; stack-trace text is validated by many position-sensitive fixtures, so
it stays a separate change. *(Closed 2026-09-10 together with the column work — see
"Call-site column precision in stack traces" below.)*

**Validation:** `go test ./...` green; six new subtests in `internal/interp/contracts_test.go`
covering inline-method qualification, a free function called from a method, inherited `require`
(direct and through a base-typed reference, and with a renamed parameter), inherited `ensure`,
derived-before-inherited reporting, and inherited `old` capture. `golangci-lint run`: no new
issues in the touched files. Baselines ratcheted.

While updating [`docs/guide/contracts.md`](../guide/contracts.md), two stale limitations were
measured and removed: contract failures **are** catchable with `try/except`, and `old` with a
`var` parameter does persist to the caller.

## 2026-09-09 — Generics (L-S4a…L-S4f)

Closed `PLAN.md` §3.2.4. **GenericsPass 15 → 23 (100%)**; overall 898 → 906 of 1,928 scored.
No other category moved: `just fixture-update` changed exactly one line in
`baselines.json`.

Measuring the eight failing fixtures first changed the shape of the work. Only five were
generics bugs; three reproduced in plain non-generic code and were fixed as the general
bugs they are. Two of the six planned items needed no code at all — `func_ptr1` (L-S4c) and
`tlist1` (the second half of L-S4f) already passed.

### Generic interfaces and array aliases (L-S4a, L-S4f)

Specialization is an AST pre-pass (`internal/generics`), and four switches in `clone.go`
(`typeParamsOf`, `declName`, `isTemplateDecl`, `specializeDecl`) decide which declaration
kinds are templates. They covered `ClassDecl`, `RecordDecl` and `TypeDeclaration` only, so
`InterfaceDecl` and `ArrayDecl` silently lost their type parameters — the parser had already
parsed them, and `attachTypeParams` dropped them on the floor. `type TTest<T> = array of T`
routes through `parseArrayDeclaration`, which returns `*ast.ArrayDecl`, not a
`TypeDeclaration`; that alone was `array1`. Both nodes gained `TypeParams []string` and
entries in all four switches plus `attachTypeParams`. `cloneNode` is reflection-based, so
interface methods and properties substitute for free.

`interface1` needed one more thing: `class (ITest<Integer>)`. The inheritance list stores
bare `*ast.Identifier`, which had nowhere to carry type arguments, and unlike an `as` cast
there is no `TypeAnnotation` in that position. `ast.Identifier` gained an optional
`TypeArgs []TypeExpression`, mirroring the existing `TypeAnnotation.TypeArgs` /
`NewExpression.TypeArgs` pattern, and `rewritePtr` gained the symmetric `*ast.Identifier`
case. All 33 non-test readers of `.Interfaces`/`.Parent` read `.Value`, which by then holds
the mangled name, so nothing downstream changed. The regenerated visitor now walks
`Identifier.TypeArgs`; `[]string` fields are skipped by the generator, as `ClassDecl.TypeParams`
already showed.

### Out-of-line generic method bodies (L-S4d)

`function TTest<T>.Test(const v : T) : T;` did not parse: `parseFunctionQualifiedName` walks
`.`-separated segments and never looked for `<`. Three fixtures — `repeat`, `while`,
`variant_implicit_cast` — hang off this, none of which `PLAN.md` had listed under L-S4d,
which asked for a hand-written regression instead.

Parser: a new pure-lookahead `looksLikeMethodTypeParams` requires a balanced angle group
**followed by a dot**, which is what separates `TTest<T>.Test` from the comparison `a < b`
and confines type parameters to class-name segments (a generic free function `Foo<T>(x)` is
still unsupported and takes the old error path). Unlike `looksLikeGenericTypeRef` it allows
any token inside the brackets, so a constrained header `<T: TObject>` is recognized too.
`parseTypeParameters` is reused verbatim. The names land on a new
`FunctionDecl.ClassTypeParams`; `ClassName` stays the base name `TTest`, which is what the
monomorphizer looks up.

Monomorphizer: `collectTemplates` gained a second arm collecting these bodies by normalized
base name — a full pre-pass, so a body written before, after or between its uses is handled
identically. `ensureSpecialized` then calls `emitMethodBodies`, which clones each body with
the same substitution, re-points `ClassName` at the mangled name and appends it right after
the specialized class. Substitution is keyed on the *header's* parameter names mapped
positionally onto the type arguments, so a header that renames them still works.

No semantic or evaluator change was needed, and that was verified before writing any code by
hand-assembling the exact statement layout the monomorphizer emits and running it: a
`FunctionDecl` with a `ClassName` is analyzed immediately rather than deferred, and
`analyzeClassMethodImplementation` clears the entry from `ForwardedMethods`, so emitting the
body after its class decl and before the first use satisfies `validateForwardMethods`.

### Three non-generic bugs filed under §3.2.4

- **`class external` methods (`class_external1`).** A body-less method was recorded in
  `ForwardedMethods` regardless of `IsExternal`, so every external class reported
  `Method "X" ... not implemented`. Reproduced without generics in four lines. The host
  implements those methods; the guard now excludes external classes and external methods.
- **Function-pointer values (`external_promise`).** `@f` where `f` is already a
  function-pointer variable was rejected — `analyzeAddressOfFunction` demanded a
  `*types.FunctionType`. Address-of a function pointer is the identity, so it now returns
  that type, and `VisitAddressOfExpression` returns the held value at runtime. Separately,
  `canAssignNil` had no function-pointer case, so `nil` could not be assigned to or passed
  for one.
- **`operator implicit (TRec) : Variant` (`specialize_to_operator_overload`).** The `<`/`>`
  overloads already resolved correctly through the generic method; what failed was
  `PrintLn(rec)`. `PrintLn` declares a Variant parameter, but `coerceBuiltinArgsToSignature`
  only coerced arguments whose *static* type was already Variant, and only between the four
  basic kinds — so the record arrived unconverted and printed as `TRec(a: 2, b: 20)`.
  A record reaching a Variant parameter now goes through the existing
  `TryImplicitConversion`; gating on records keeps the common argument kinds off the
  conversion registry.

**Validation:** `go test ./...` green. `golangci-lint run --new-from-rev=main`: 0 issues.
`just fixture-report` 906/1,928 with GenericsPass at 23/23 and no category regressions.
New tests: nine in `internal/parser/generics_test.go` (generic interface/array declarations,
`class (ITest<Integer>)`, four out-of-line header forms, and guards that plain
`TFoo.Bar`, `TOuter.TInner.Bar` and `<` comparisons are unaffected); nine in
`internal/generics/monomorph_test.go` (interface and array specialization, inheritance-list
instantiation, one body per specialization, emission order, substitution, body-before-use,
arity mismatch, `Default(T)`, and a non-template header left in place); external-class and
function-pointer subtests in `internal/semantic`; two `PrintLn`-of-record tests in
`internal/interp/operator_test.go`.

**Left open, measured:** `GenericsFail` (0/8) belongs to §4 error-detection parity —
`implem_mismatch1` now gets further but still fails, since DWScript's "T expected but u
found" check for a renamed out-of-line type parameter is not implemented. Type-parameter
constraints are parsed and ignored. Comparing a function pointer against `nil` (`f = nil`)
still reports "operator = requires comparable types"; assignment and argument passing work,
and no fixture demands the comparison.
## 2026-09-09 — Overloads and method pointers, PLAN.md §3.2.5 emptied (L-S5a–L-S5d)

Stacked on §3.2.4. `OverloadsPass` 33 → 37 of 39 (85% → 95%); corpus 906 → 910 of 1,928.
No category regressed (`OperatorOverloadPass` 5, `ArrayPass` 96, `LambdaPass` 4,
`GenericsPass` 23, `SimpleScripts` 331, all unchanged). Two of the four tickets turned out to describe the wrong defect; the
reproductions below are what actually shipped.

### L-S5a — operator dispatch is no longer blind to a nil operand's declared type

`class_equal_diff` was not a missing `operator =` feature — user `operator =`/`<>` on classes
already worked. The fixture declares `var c : TMy;` and never assigns it, so both operands are
`NilValue` at runtime. `evalTryBinaryOperator` (`internal/interp/evaluator/runtime_ops.go`)
searched class operators only for a concrete `*runtime.ObjectInstance`, and both it and
`lookupGlobalOperator` keyed the registry on `runtime.LanguageType(operand)`, which maps
`*runtime.NilValue` to `types.NIL`. `Lookup("=", [NIL, NIL])` could never match the registered
`(TMy, TMy)` entry, so dispatch fell through to builtin reference equality and printed
`True`/`False`.

The operand *expressions* are now threaded from `VisitBinaryExpression` /
`VisitUnaryExpression` down through `tryBinaryOperator` / `tryUnaryOperator` into the lookups.
New `operandOperatorType` prefers the runtime language type and falls back to the analyzer's
resolved static type (recorded for every expression by the `defer` in
`Analyzer.analyzeExpression`) when the value carries none; `operandClassInfo` does the same for
the class-operator receiver, resolving the static `*types.ClassType` through
`typeSystem.LookupClass`. Compound assignment has no operand expressions and passes `nil`,
which skips the fallback and keeps its previous behavior.

### L-S5d — `inherited ClassName` reaches TObject's builtin

`overload_on_metaclass` was not a metaclass-dispatch gap either: metaclass `ClassName`
dispatch already produced `TObj`/`TObj`/`TSub`. Only `inherited ClassName` failed to compile,
with the misleading `'inherited' cannot be used in class 'TObj' which has no parent class`.
The builtin was registered as `objectClass.Methods["ClassName"]`, but `ClassType.GetMethod`
reads `MethodOverloads` only, so the member lookup missed and fell through to the
`isTObjectParent` branch.

`ClassName` is now registered with `AddMethodOverload` like `Destroy`/`Free`, marked
`IsSynthesized` so `analyze_classes.go`'s `memberName == "classname"` hiding check still fires
only for a *user* declaration (`classname_hide_with_default` unchanged). At runtime,
`executeInheritedCallDirect` gained a `ClassName` builtin fallback beside the existing
`Create`/`Destroy`/`Free` one, returning `objVal.ClassName()` exactly as the normal instance
path does.

`resolveClassMetaMember` was deliberately left alone: it answers `ClassName` on a metaclass
receiver from the builtin *before* consulting user methods, which looks like a gap but is what
this fixture wants — the user's `ClassName` is an instance method, so `TObj.ClassName` must
still yield `TObj`.

### L-S5b — `@obj.Method` binds as a method pointer

The evaluator already implemented `@obj.Method` in full (evaluate receiver,
`CreateMethodPointer`, return a `FunctionPointerValue` with `SelfObject` bound). It was blocked
purely by a semantic stub that errored `method pointers (@TClass.Method) not yet implemented`
on every `*ast.MemberAccessExpression` operand of `@`.

New `analyzeAddressOfMethod` analyzes the receiver, resolves the member through
`getMethodOverloadsInHierarchy` (skipping constructors, binding the first overload since a
method pointer cannot represent an overload set), records the usage, and returns
`types.NewMethodPointerType(...)`. A `*types.ClassOfType` receiver — `@TClass.Method`, the
unbound address-of-class-member case — still errors, but now says so accurately; that stays
§3.3 (`func_ptr_symbol_field`).

For `TEvent = procedure` to accept `@o.Foo`, function-pointer compatibility had to exist at
all: `IsCompatible` had no pointer branch, so it only succeeded on `Equals`, and
`MethodPointerType.Equals` rejects a plain `FunctionPointerType`. New `types.IsPointerType` and
`types.PointerCompatible` delegate to the pointer types' own `IsCompatibleWith`, which already
encoded the right asymmetry (a method pointer satisfies a function-pointer slot; never the
reverse), and both `IsCompatible` and `typeDistance` call the one helper. In `typeDistance` an
exact match keeps distance 0 and a compatible-but-not-identical pointer scores 1, so
`Test(@o.Foo)` prefers `Test(a: TEvent)` over `Test(a: TObject)`.

`TestMethodPointer_StoredInArray` (`internal/interp/method_pointer_test.go`), skipped with a
comment blaming exactly this analyzer stub, is un-skipped and passes.

### L-S5c — function-pointer arguments survive runtime overload re-resolution

`overload_func_ptr_param` compiled clean; it failed at runtime, where the evaluator
re-resolves overloads from evaluated argument *values*. `getValueType` had no case for
`*runtime.FunctionPointerValue`, so it fell to `default`, `getClassMetadataFromValue` returned
nil, and every function-pointer argument was typed `NIL` — which `typeDistance` scored as
incompatible against every `procedure(...)` parameter.

`getValueType` now returns the value's `PointerType`, with a fallback that rebuilds the
signature from `Callable` (method pointer when `SelfObject` is bound) for pointers whose
signature was resolved after the value. `typeDistance`'s `NIL` special-case was extended to
give `FUNCTION_POINTER` / `METHOD_POINTER` targets the same rank as `FUNCTION` (2), so an
explicit `nil` still binds to a function-pointer parameter. No expected-type push-down was
needed — compile-time selection already picked the right overload by exact `Equals` matching,
so the §3.4 item stays open and untouched.

### Measured, still open

- The 14 `OverloadsFail` fixtures all need `The function X was forward declared but not
  implemented`, which exists nowhere in the tree (`grep -rn "forward declared"` finds only the
  class-parent variants in `analyze_classes_decl.go`). That is §4 / F7, not §3.2.5.
- `OverloadsPass/overload_ambiguous_delegate` and `overload_class_method` are the two
  remaining `OverloadsPass` failures and cannot pass: their expected output contains the
  case-mismatch hints covered by the ✋ won't-fix in §5.

  In `overload_ambiguous_delegate`, `a.A(@fn)` / `b.B(@fn)` now correctly select the `TFunc`
  overloads (they previously selected by declaration order). The bare-identifier lines
  `a.A(fn)` / `b.B(fn)` still do not match: DWScript implicitly *invokes* a parameterless
  function used as a non-pointer argument and expects `A int` / `B int`, but the analyzer
  passes `fn` as a pointer. That defect is pre-existing and untouched here — before this
  change both candidates scored identically and declaration order decided the winner (`TA`
  happened to print `A int`, `TB` printed `B func`); giving compatible pointers a real
  distance simply makes the pre-existing mis-binding deterministic (`A func` / `B func`).
  Fixing it needs the expected-type push-down tracked in §3.4, which no §3.2.5 target fixture
  required, so it stays open.

  In `overload_class_method`, a bare `ClassName` inside a class method still yields the empty
  string rather than `tobj`; only the case-hint direction changed. Registering the builtin as
  a method overload was not sufficient there, and that fixture is hint-blocked regardless.

**Validation:** `go test ./...` green; `golangci-lint run` shows no new issues in the twelve
touched files. Baselines ratcheted (`OverloadsPass` 33 → 37) and `TEST_STATUS.md` regenerated.

## 2026-09-09 — Sets, PLAN.md §3.2.6 emptied (L-S6a–L-S6c)

Stacked on §3.2.5. `SetOfPass` 21 → 25 of 25 (84% → **100%**), `SetOfFail` 1 → 5 of 14
(7% → 36%), `SimpleScripts` 338 → 340; corpus 910 → 920 of 1,928 (47% → 48%). Harness and CLI
agree. No category regressed.

Measurement first, and it moved the work considerably:

- `init_from_array`, named in the PLAN item, **already passed**. It needed a regression guard,
  not a fix.
- The runtime half of L-S6b already worked: `var r : record F : TMySet end = (F: [enumTwo]);`
  followed by `r.F := [enumOne]` printed `ok1` / `ok2` before any change here.
- The whole of `in_set_out_of_range.pas` already produced its expected output once
  `TBigEnum(IntPower(10, i))` was rewritten as `TBigEnum(Round(…))`. The single blocker was a
  compile-time rejection of the Float → enum cast.
- All four failing `SetOfPass` fixtures failed at *compile* time. None reached the evaluator.

### L-S6a — bracket literals convert on the expected type, not on their shape

DWScript writes array and set constructors with the same `[]` syntax. The parser guesses which
one it has from the element node kinds (`shouldParseAsSetLiteral`, `internal/parser/arrays.go`),
and semantic analysis then re-classifies against the expected type. That re-classification was
itself gated by a whitelist of element node kinds in `analyze_expressions.go`, so a typecast
element (`[TMyEnum(3)]`) fell through both filters and stayed an array literal. `s + [TMyEnum(3)]`
therefore reported `set operator + requires set operands, got set of TMyEnum and
array[0..0] of TMyEnum`.

The whitelist is gone. When the expected type is a `*types.SetType` the literal is a set
constructor, whatever its elements look like, and `analyzeSetLiteralWithContext` — which already
owns the per-element ordinal and element-type diagnostics — reports anything wrong with them.

`evaluateConstant` (`internal/semantic/analyze_types.go`) had cases for record and array literals
but none for `*ast.SetLiteral`, so `const v : TMy = [A, C];` was rejected as "not a compile-time
constant" and `v` was never defined. New `evaluateConstantSetElements` folds the elements,
keeping a range as its two constant bounds rather than expanding it. The folded value is only a
constancy proof: `VisitConstDecl` re-evaluates the initializer at run time.

Closes `array_to_set`.

### L-S6b — partial record constants, and `set of` in a variable's type

`set_in_record` needed two unrelated things.

*Partial record literals.* `const rA : TRecord = (A: [enumOne]);` was rejected by
`missing required field 'b' in record literal` — an invented message no fixture expects, which
DWScript does not have. It is gone; omitted fields keep their default. That required
`getZeroValueForType` (`internal/interp/evaluator/index_ops.go`) to learn `"SET"`, or `rA.B`
would have been `NilValue` and `enumOne in rA.B` a runtime type error rather than `False`.
`GetDefaultValue` gained the same case (plus `"ASSOCIATIVE_ARRAY"`), so a set-typed function
`Result` now starts as the empty set instead of nil. `SimpleScripts/const_record` newly passes
as a side effect; `internal/semantic/record_test.go` was updated to assert the new behavior.

*Inline anonymous enums in a type position.* `var elemsInline : set of (et3) = [];` did not
parse: the desugaring into an implicit `EnumDecl` existed only in `parseSetDeclaration`
(the named `type TMy = set of (A, B)` form), while `parseSetType` handed the element type to
`parseTypeExpression`, which cannot start at `(`.

`parseSetType` now desugars too, via new `parseInlineSetEnum`. A variable's set type has no name
to derive the implicit enum's name from, so it is minted from the `set` token's position
(`$InlineEnum$<line>$<col>`). Routing the declaration out needed a small general facility: the
parser keeps a `pendingTypeDecls` queue (saved and restored with the rest of the speculative
parsing state), and `parseStatement` — now a thin wrapper around the renamed
`parseStatementInner` — drains it into a `BlockStatement` ahead of the statement.

That block must not open a scope, or the enum's members would be invisible to everything after
it. `analyzeBlock`'s existing transparency test infers "declaration section" from the block's
contents, and a block mixing an `EnumDecl` with a `VarDeclStatement` matches neither predicate.
Rather than widen the inference, `ast.BlockStatement` gained an explicit
`SharesEnclosingScope` flag that the parser sets on exactly these hoist blocks. The evaluator's
`VisitBlockStatement` was already scope-transparent and needed no change.

Finally, the var-declaration path in `visitor_statements.go` rejected a bracket literal against a
`set of` declaration with `expected array type, got set of TElements` — only `[]` reaches it as
an `ArrayLiteralExpression`, since a non-empty `[a, b]` is already a `SetLiteral`. It now routes
a set-typed declaration through new `evalBracketLiteralAsSet`, which annotates the synthetic node
the way the assignment path already did.

Closes `set_in_record` and `init_from_empty_array`.

### L-S6c — range validation and out-of-range diagnostics

*Float → enum casts.* `isValidCast` accepted only Integer as an enum cast's source, so
`TBigEnum(IntPower(10, i))` failed with `Cannot cast this type to "Integer"`. Float is now
accepted and truncated to its ordinal in `castToEnum`. Deliberately **not** bounds-checked:
`in_set_out_of_range` depends on `TEnum(-1)` / `TEnum(3)` producing an out-of-range enum whose
membership test simply answers `False`, and the lenient `SetValue.HasElement` behavior behind
that is unchanged. Closes `in_set_out_of_range`.

*`Element is out of set bounds`.* New `checkSetElementBounds` folds each set-literal element to a
compile-time ordinal (`evaluateConstantInt`, recursing into a range's two bounds) and compares it
against `types.OrdinalBounds` of the set's element type. Non-constant elements are left to run
time. The check is value-based, not "is a cast": `cons_autocast_bounds` expects errors for
`TMyEnum(3)` and `TMyEnum(-1)` but none for `TMyEnum(0)`. It only fires because L-S6a made those
elements set elements in the first place. Closes `cons_autocast_bounds`.

*`Set expected`.* `analyzeIncludeExclude`'s first-argument diagnostic used go-dws wording at the
call's position; it now uses DWScript's wording anchored on the argument. Closes `invalid_base`.

*`Enumeration expected`.* `type TMySet = set of procedure;` produced
`expected type identifier after 'of' in set declaration` at the wrong column plus a stray
`";" expected`. `parseSetDeclaration` now emits DWScript's single message at the `of` token and
recovers to the declaration's semicolon (`skipToSetDeclarationEnd`). Closes `invalid_type`.

*`Set has too many elements for cast to integer`.* Nothing checked this at all. A set's integer
form is its ordinal bitmask, so the base type has to fit one; `checkSetIntegerCastWidth` rejects
a base type spanning more than 32 ordinals. `TSet(i)` reaches it through `isValidCast`, but
`Integer(s)` does not — `Integer` is a registered conversion builtin, not a type cast — so
`analyzeBuiltinFunction` gained an `"integer"` case that applies the same rule to the resolved
argument type after the ordinary builtin analysis has run. Closes `integer_vs_set`.

### Scope

The other nine `SetOfFail` fixtures (`bracket_left_missing`, `bracket_right_missing`,
`for_in_set_missing_do`, `include`, `invalid_method`, `invalid_operand`, `of_missing`,
`test_non_variable`, `type_missing`) are parser-recovery and message-parity work, not set
semantics. They stay with §4 / F7.

**Validation:** `go test ./...` green; `golangci-lint run --new-from-rev=HEAD` reports 0 issues.
New tests in `internal/semantic/set_test.go` (`TestBracketLiteralConversions`,
`TestSetElementBounds`, `TestSetIntegerCastWidth`) cover both conversion directions, the bounds
check's positive and negative cases, and the cast-width rule. Baselines ratcheted
(`SetOfPass` 21 → 25, `SetOfFail` 1 → 5, `SimpleScripts` 338 → 340) and `TEST_STATUS.md`
regenerated.

## 2026-09-10 — conditional compilation: `Declared()` and the message directives (§3.2.7)

Closes `PLAN.md` §3.2.7 (L-S7a, L-S7b) on branch `feat/conditional-compilation-3.2.7`.
Fixtures **920 → 937**: `FailureScripts` 107 → 122, `SimpleScripts` 340 → 342.

Both tickets were marked "needs re-identification" because the old ArrayPass/SetOfPass
attribution was stale. Re-measuring first paid for itself: the `Declared` half turned out to be
two independent gaps (a preprocessor one and a missing builtin), and the `{$FATAL}` half was
blocked by a plumbing defect that also gated ten unrelated fixtures.

> The `reference/dwscript-original/` submodule is not checked out in this tree. Every message,
> column and severity below was derived from the fixtures and their upstream-authored `.txt`
> files, which are the authoritative spec here.

### The blocker underneath L-S7b: lexer diagnostics were discarded

`internal/frontend/result.go` forwarded only `p.LexerIncludeErrors()`; everything recorded
through plain `addError` was explicitly "advisory and not surfaced". `{$HINT}`, `{$WARNING}`,
`{$ERROR}` and `{$FATAL}` were not directive cases at all, so they fell through to
`unknown compiler directive` — and that error was then thrown away. `FailureScripts/error_directives`
emitted *nothing whatsoever*.

Rather than promoting every lexer error (a large, untargeted blast radius), the lexer gained a
dedicated directive-diagnostic channel carrying a `Severity` and an optional pre-rendered
DWScript string. `frontend.Diagnostic.Rendered` already short-circuits `Render()`, so
`Hint:` / `Warning:` / `Compile Error:` bypass `FormatDWScriptError`'s hardcoded `Syntax Error:`
prefix without disturbing the shared formatter. Diagnostics are deduplicated by message and
position because parser backtracking can re-lex the same directive.

One subtlety cost a debugging round: these diagnostics must **not** set `BlocksSemantic`. A
broken `{$INCLUDE}` means code is missing, so analysing the remainder is noise; a `{$FATAL}` is
the opposite — everything before it parsed correctly, and DWScript still reports that code's
errors. Marking them blocking silently deleted all six expected errors from `final.pas`.

### L-S7b — message and severity directives

`{$ERROR}` and `{$FATAL}` share the `Compile Error:` prefix and differ only in whether
compilation continues; `{$FATAL}` stops tokenizing immediately while retaining every message
already recorded. Messages anchor at the directive *name* (`{` column + 2); a missing or
unquoted argument reports `String expected` at the argument when present and at the closing
brace otherwise — a position `readDirectiveContent` did not previously track. Both `'…'` and
`"…"` quoting is accepted, and `strings.Fields` had to be abandoned for argument parsing since
it truncates `{$HINT 'first hint'}` at the first space. `{$HINTS}`/`{$WARNINGS}` take
ON/OFF/NORMAL/STRICT/PEDANTIC and otherwise report `ON/OFF expected`; `{$R}`/`{$RESOURCE}`
require a string. Directives in an inactive branch emit nothing.

Closes `error_directives` and `hint_warn`.

### Conditional-directive parity (§4/F7 work taken opportunistically)

With the channel in place, ten more fixtures were within reach and were closed:
`Unbalanced`/`Unfinished conditional directive` and `Compiler switch "X" unknown` replace the
former lowercase advisory strings; an unknown switch inside a dead `{$IFDEF}` branch is not
reported; an unterminated directive reports `"}" expected` (plus `Name of include file expected`
for `{$INCLUDE}`) instead of cascading into an unbalanced-conditional report. `{$IFEND}` closes
`{$IF}`, and `{$REGION}`/`{$FILTER}`/`{$F}` are recognized and ignored — the last two are real
DWScript include variants that the new unknown-switch error would otherwise have broken
(`examples/rosetta/Include_a_file.dws` caught this).

The known-switch table is evidence-based, built from an inventory of every `{$…}` occurrence in
the corpus, because `switch_invalid1`, `switch_invalid3` and `invalid_switch` are *negative*
tests that require specific names to stay unknown.

Closes `conditionals1`, `conditionals2`, `conditionals3`, `conditionals4`, `conditionals5`,
`conditionals6`, `conditionals_else1`, `conditionals_else2`, `conditionals_else3`,
`switch_invalid1`, `switch_invalid3`, `invalid_switch`.

### L-S7a — `Declared()`

Two independent gaps. `trackConst` dispatched on `ASSIGN` (`:=`) but never `EQ` (`=`), so an
ordinary `const Test = 101;` was never tracked; the one-line fix closed
`SimpleScripts/conditionals_nested4` on its own. Separately, `Declared` was a plain alias for
`Defined` in the lexer and did not exist as an expression at all.

*Preprocessor side.* A forward-only declaration tracker (`internal/lexer/declarations.go`)
records class/record/interface/helper types with their members as dotted names, mirrors helper
members onto the helped type so `Declared('TObject.Proc')` resolves through a
`helper for TObject`, seeds `TObject`, and tracks file-scope declarations. Bare member names
stay invisible — `declared.pas` requires `Declared('dummy')` to be false even though
`TMyRecord.Dummy` exists. Because the lexer scans sequentially, point-of-use visibility falls
out for free: `conditionals_nested4` runs the same `{$IF Declared('Test')}` before and after the
const and expects false then true. The tracker is a deliberate heuristic in the spirit of
`trackConst`, not a second parser, and is deep-cloned in `SaveState`/`RestoreState` because its
state machine is position-dependent (unlike the idempotent `constValues` map). `Defined()` is
now narrowed to `{$DEFINE}` symbols only, the correct DWScript distinction.

`{$IF}` expression positions are now real; they were previously always line 0, column 0.
`Defined(<non-string>)` reports `String expected` and `Declared(<non-constant>)` reports
`Constant expression expected`, both at the argument column.

*Expression side.* `Declared` and `ConditionalDefined` are compile-time intrinsics that validate
a constant string argument, resolve a case-insensitive dotted name against the type registry,
symbol table and builtins (stripping a leading `Internal.` segment, since builtins live in that
unit), and fold to a boolean. A non-String argument now reports `String expected` at the
argument column instead of cascading into `Unknown name` plus `Undefined variable`.

Closes `SimpleScripts/declared` and `FailureScripts/special_funcs5`.

### Scope and divergences

- ✋ **`FailureScripts/static_methods` regressed** and is the one fixture lost. It passed only
  because `{$FATAL}` was ignored: upstream's `.txt` omits the `Compile Error: aborted` line even
  though the directive is present on line 26. The only structural difference from `final`,
  `default_constructor` and `virtual1` — where the fatal *is* reported — is that it is the sole
  file whose `{$FATAL}` is not at column 1. That correlation holds 5/5 but has no plausible
  tokenizer mechanism, so it was not encoded as a rule. Net for the group is +2.
- ✋ `ConditionalDefined(s)` always folds to `False`: `{$DEFINE}` symbols live in preprocessor
  state the analyzer cannot reach. Argument validation is complete.
- ✋ `HelpersPass/declared_helper` now resolves all four `Declared()` calls and emits no spurious
  `{$FATAL}` output, but cannot pass: its expectation needs the case-mismatch hints §5 marks
  won't-fix, and `THelper.Proc(TObject.Create)` — calling a helper method with an explicit
  instance argument — is an unimplemented call form. That form is the one concrete follow-up.
- ✋ `FailureScripts/special_funcs4` matches on its first line; the second needs
  `Expression expected` for `Inc(i, )`, which is parser recovery (§4/F7).
- ✋ `FailureScripts/conditionals2.1` wants the unbalanced report at the directive *argument*
  (column 9) while `conditionals2` wants it at the directive *name* (column 3), for byte-identical
  directives. The name anchor was chosen for consistency with every other directive diagnostic.
- Note for §5: `hint_pedantic` shows `{$HINTS OFF}` / `{$HINTS PEDANTIC}` are expected to switch
  hint reporting on and off mid-file. That is plausibly the per-test hint configuration §5 says
  is unrecoverable, and may be worth revisiting.
- Unit boundaries are unchanged: `internal/units/registry.go` builds a fresh lexer per unit, so
  defines and tracked declarations do not cross `uses`.

**Validation:** `go test ./...` green; `golangci-lint` reports no findings in any new or changed
file. New table-driven tests in `internal/lexer/directives_test.go` (declaration tracking,
`Defined` vs `Declared`, directive diagnostics) and `internal/semantic/analyze_declared_test.go`.
Baselines ratcheted and `TEST_STATUS.md` regenerated.

## 2026-09-10 — Indexing the result of an implicit (parenless) call (PLAN.md §3.3)

`Test['toto']`, where `Test` is a parameterless function returning an array, rejected the program
at compile time: `Syntax Error: Array expected`, plus fallout `Unknown name "r1"` wherever the
dropped declaration was later used. `analyzeIndexExpression` analyzed `expr.Left` and matched the
resulting type against the array / associative-array branches without ever unwrapping the implicit
call, so it saw the function's own type. Member access on the same shape (`Test.Keys`) already
worked, because `analyzeMemberAccess` and `analyzeRecordFieldAccess` both unwrap.

That unwrap — `getImplicitCallType` on the expression, falling back to
`implicitCallReturnTypeFromType` on its type — was duplicated at both member-access sites. It is
now one helper, `applyImplicitCallType` (`internal/semantic/analyze_function_calls.go`), which both
sites call and which `analyzeIndexExpression` applies to `expr.Left` before the class-default-
property, associative-array, array and string branches, so every branch sees the result type.

The helper only fires for a zero-parameter function type, so a function pointer that takes
arguments is still not indexable, and a parenless call returning a non-indexable type still gets
`Array expected`.

Overload sets need one extra step: they deliberately carry no type of their own, so `expr.Left`
analyzes to `nil` and the old nil check bailed out before the unwrap could run. The unwrap is now
applied before that check, and `applyImplicitCallType` resolves a nil type against the overload
set, using the single parameterless overload when the set has exactly one. A set with none, or
with several, is left to regular overload resolution.

**Validation:** `go test ./... -timeout 30m` — 26 of 27 packages green, and `cmd/dwscript`
verified separately with `go test ./cmd/dwscript -timeout 120m`. The split is not cosmetic:
`cmd/dwscript` shells out to the built binary once per case, so at the default per-package
limit it trips the pre-existing 10-minute timeout, and on a loaded machine it exceeds 30
minutes too. That slowness predates this change and is unrelated to it — no bare
`go test ./...` is green here. New table-driven tests in
`internal/semantic/analyze_arrays_implicit_call_test.go` cover dynamic, static and associative
array results, an associative array of records, string indexing, an overload set whose
parameterless overload returns an array, and two negative cases.
`just fixture-check` passes with no category moving; baselines unchanged.

### Scope

`AssociativePass/records` still fails, on hash iteration order alone: `Keys.Join(',')` yields
`a,b` where DWScript yields `b,a`. That remains the open §3.3 item and is not touched here.

## 2026-09-10 — ByteBuffer, PLAN.md §3.3 FunctionsByteBuffer closed (19/19)

`ByteBuffer` was entirely missing: 17 of the 19 `FunctionsByteBuffer` fixtures failed with
`unknown type 'ByteBuffer'` and the other two with `Unknown name "ByteBuffer"`. The corpus was the
only specification available — `reference/dwscript-original/` is empty in this checkout — so every
member name, arity, endianness rule and diagnostic below was derived from the fixtures and is
pinned by them.

### Shape of the type

`ByteBuffer` is **not** a class. Two fixture facts decide this. `assign` declares
`var b2 : ByteBuffer;` and immediately calls `b2.ToDataString`, so a declared variable must be
live rather than nil; `new` assigns one buffer to another and then resizes through the second
name, so assignment must alias. The type is therefore a singleton `types.Type`
(`types.BYTE_BUFFER`, modelled on `JSONVariant`) whose runtime value is always a
`*runtime.ByteBufferValue`. The pointer gives reference semantics for free; auto-instantiation is
a `BYTE_BUFFER` case in the two zero-value constructors. `Assign` is the copying operation.

### Engine

`internal/interp/runtime/bytebuffer.go` holds the whole byte-level engine, dependency-free apart
from the standard library: resize with zero-fill (`init` checks that a shrink-then-grow yields
zeroes, so the buffer is reallocated rather than resliced), a cursor whose legal range is
`0 … Length` inclusive, little-endian typed accessors, x87 80-bit `Extended` encode/decode, and the
data-string conversion that keeps the low byte of each UTF-16 code unit (`strings` pins
`ByteBuffer(#$1234#$5678)` = `$34 $78`).

Diagnostics are produced by the engine because their wording is fixed by the corpus:
`Out of range (index I, size S for length L)`, `Position P out of range (length L)` and
`value V out of T range`. `dwords` shows the overflow check running before the range check, and
that a failed write must not advance the cursor.

`GetIntegers(index, count, size, signed)` reproduces an upstream quirk: `integers` calls it with
indices 0, 1, 2 and 0 and every result starts at byte 0, so `index` participates in the bounds
check but does not shift the read origin. That is documented at the function and in the guide
rather than silently "fixed", since parity is the goal.

### Wiring

Semantic analysis resolves the name through `types.TypeFromString`, types the intrinsic members
from a table in `internal/semantic/analyze_bytebuffer.go`, and hooks member access, method calls,
`new ByteBuffer` and the `String -> ByteBuffer` cast. Execution adds a `KindByteBuffer` case to
`DispatchMethodCall` and routes parameterless member access to the same dispatcher, because
DWScript treats `b.ToJSON` and `b.ToJSON()` alike. Failures become catchable `Exception`s
positioned at the member name, reusing `arrayMethodNamePos` and the existing array-bounds pattern.

### Two formatting fixes found along the way

`floats` and `integers` failed on output formatting rather than on buffer semantics, and both
gaps were corpus-wide:

- Float rendering used Go's shortest round-trip form. DWScript's `FloatToStr` prints 15
  significant digits, so `3.141592025756836` must print as `3.14159202575684`, and integral
  15-digit values such as `-845680067215360` must not go exponential. Both now use `%.15g`.
- `Integer.ToHexString` rendered a negative value with a minus sign (`-80`) instead of its 64-bit
  two's-complement pattern (`FFFFFFFFFFFFFF80`).

Measured against the full suite, each change moves only `FunctionsByteBuffer` and regresses no
category.

**Validation:** `go test ./... -timeout 30m` green;
`golangci-lint run --new-from-rev=main` reports 0 issues; `just check-fmt` clean. New tests:
`internal/interp/runtime/bytebuffer_test.go` (table-driven over endianness, bounds, overflow,
`Extended` round-tripping and the encodings) and `internal/interp/bytebuffer_test.go` (the
script-level surface: auto-instantiation, aliasing versus `Assign`, both accessor arities, the
data-string cast, catchable diagnostics). Fixture totals **938 → 957**, `FunctionsByteBuffer`
**0/19 → 19/19**, no category below its previous value. Baselines ratcheted and `TEST_STATUS.md`
regenerated. User-facing documentation: [`docs/guide/bytebuffer.md`](../guide/bytebuffer.md).

## §3.3 — JSON node ownership, number formatting, "Not a value" cast parity

Closed 2026-09-10 on branch `feat/plan-3.3-json-ownership-formatting`, re-measured after the
rebase onto §3.2.7. Closes the §3.3 item
"JSON node reparent/ownership … float formatting … variant → scalar cast message parity".

| | before | after |
| --- | --- | --- |
| `JSONConnectorPass` (CLI) | 59 / 82 | 65 / 82 |
| `JSONConnectorFail` (CLI) | 2 / 9 | 2 / 9 |
| TOTAL (`just fixture-report`) | 938 / 2,042 | 944 / 2,042 |

No category moved down.

### Node ownership

`jsonvalue.Value` had no parent pointer at all, so "reference semantics" was modelled purely as
Go pointer sharing: inserting a node that already lived in a container aliased it into two places.
DWScript's `TdwsJSONValue` has an `Owner`, and inserting a node elsewhere *moves* it.

`Value` gained an `owner` back-pointer plus `Owner()` and `Detach()`. `ObjectSet`, `ArraySet` and
`ArrayAppend` now adopt their child (detaching it from a previous owner first); `ObjectDelete`,
`ArrayDelete`, `ClearArray` and replaced entries release it; `Clone()` returns a root. Detaching
from an object drops the key, detaching from an array removes the slot — which is what makes
`a[0] := b; a[1] := b` print `[null,{…}]` rather than duplicating the node.

Two call sites needed care. `Swap` must not reparent, so it uses a new non-adopting `ArraySwap`.
The JSON index assignment detaches the incoming node *before* padding the array with nulls,
otherwise a move inside the same array shrinks it back out from under the new index.

Closes `reparent`, `array_add_dupe`, `reposition_node_in_array`.

### Number formatting

Three separate gaps behind `numbers` and `int64_json`:

*Float rendering.* `FloatValue.String()` used Go's shortest round-trip form
(`0.3333333333333333`, `1e+99`). Delphi's `FloatToStr` uses `ffGeneral` with 15 significant
digits and an exponent with no `+` and no leading zeros. The rule now lives in a new
`internal/dwsfmt` package (`FloatToStr`, `NormalizeExponent`) shared by the runtime, the JSON
serializer's `FormatNumber` and `runtime.FloatToStr`, so there is one spelling of a float.

*Integer `/` by zero.* `/` is float division in DWScript even for integer operands, and DWScript
runs with FPU exceptions masked, so `0/0` is NaN. The evaluator raised "division by zero"; it no
longer does. `div` and `mod` are unaffected — the `div_by_zero_int` / `mod_by_zero_int` fixtures
only exercise those.

*JSON → scalar narrowing.* `var f : Float := jsonNode` left `f` holding a JSON node, so it
printed the raw Int64 and `Round(f)` rejected it as "got JSON". `TryImplicitConversion` now
narrows a JSON immediate to Integer/Float following DWScript's immediate rules (numbers and
booleans convert, strings are parsed with a 0 default, containers are left alone).

### "Not a value"

`Float(jsonObject)` reported `Could not convert variant of type (Object) into Float` as an
unpositioned runtime error, printed with the `"\n at line L, column: C+2"` suffix the expression
statement appends. DWScript raises `Not a value [line: 5, column: 5]` — one line, bracketed, at
the position of the *statement*, not of the failing sub-expression.

`ExecutionContext` now tracks the innermost statement (`CurrentStatement`), set alongside
`CurrentNode` in `Evaluator.Eval`, because DWScript reports runtime exceptions at statement
granularity. `evalTypeCast` raises a catchable exception with that position when the JSON node has
no scalar value. The same change gives the JSON branches of `castToInteger`/`castToFloat` the
immediate rules, so `Float(json '')` is 0 and `Integer(json '1.25')` is 1. Closes `explicit_cast`.

### Scope

The remaining 17 `JSONConnectorPass` failures are other people's items: lvalue vivification
(`generate1`, `basic_generate`), record copy-on-assign (`stringify_record`), and the
`as_const_param` / `associative_array` / `circular_references` / `const_array` / `global_var` /
`implicit_*` / `in_static` / `stringify_array_of_array` / `write_immediate_prop` /
`delete_array_index` / `comparison2` / `assign_static_to_dynamic` group.

**Validation:** `go test ./... -timeout 40m` green; `golangci-lint run --new-from-rev=main`
reports 0 issues; `just check-fmt` clean. New tests: `internal/jsonvalue/ownership_test.go`
(adopt/detach/removal/clone/parse-time ownership), `internal/dwsfmt/float_test.go`,
`internal/interp/evaluator/json_scalar_test.go`. `docs/guide/json-type-mapping.md` gained a
"Single Ownership (Reparenting)" section — the page previously documented pure reference
semantics.

## 2026-09-10 — Call-site column precision in stack traces, PLAN.md §3.3

`SimpleScripts` 343 → 346 of 442 (79% → 80%); corpus 938 → 941 of 2,042 (49%). No category
regressed — the whole `just fixture-report --list-fails` table was captured before and after and
diffed; the only lines that changed are the three closed fixtures. Closes
`SimpleScripts/stacktrace`, `SimpleScripts/exceptobj3` and `SimpleScripts/contracts_subproc`, and
with them the last item of §3.3's diagnostics group.

The PLAN item flagged this as the higher-risk one because it lives in shared position logic. The
risk turned out to be real but narrow, and measurement is what located it: the naive reading of
the fixtures — "the raise site is the constructor name" — regressed `exception_nested_call` and
`exceptobj2` on the first attempt, because those two pin the *same* `raise EFoo.Create(…)` shape
to a *different* column.

### The governing rule

A stack frame is positioned at the **name token of the callee**:

| written | frame position |
| --- | --- |
| `Foo(1)` | `Foo` |
| `obj.Bar` | `Bar`, not `obj` |
| `(new TTest).TestMeth` | `TestMeth`, not `TTest` |
| `Exception.Create('x')` | `Create`, neither `Exception` nor the closing paren |

`callSitePos` (`internal/interp/evaluator/call_site.go`) is that view of a node, and the
frame-push sites use it. AST `Pos()` was deliberately **not** changed: it starts at the receiver
for `MethodCallExpression` and at the class name for `NewExpression`, and diagnostics, hints and
the whole `*Fail` corpus are calibrated against it. Moving `Pos()` would have repositioned a great
many messages to buy three fixtures.

### Two positions, not one

`stacktrace.pas:13` and `exceptobj2.pas:5` are both `Raise Exception.Create(…)`. The first expects
column 20 (the `Create` identifier), the second column 33 (just past the `)`). They are not in
conflict — they are reading two different things:

- the **unhandled-exception message** is positioned just past the raised expression, which is what
  `Exception.End()` models and what we already did;
- the innermost **`Exception.StackTrace`** frame is the site where the exception object was
  *constructed*. DWScript captures the call stack inside the constructor, so that frame is the
  `Create` call, and the `raise` statement's own position never appears in the trace.

`ExceptionValue` therefore grew `OriginPos` alongside `Position`. It falls back to `Position` when
unset, so a runtime error — division by zero, a nil dereference — which has no separate
construction site keeps the frame it had.

### Inline methods are now class-qualified

`ExecuteUserFunction` built the frame name from `FunctionDecl.ClassName`, which only an
out-of-line header (`procedure TFoo.Bar;`) carries. A method written inline in the class body has
no qualifier to parse and came out as bare `TestMeth` where DWScript writes `TTest.TestMeth`. The
parser now records the owner on every routine declared in a class body —
`FunctionDecl.DeclaringClassName`, covering constructors, destructors and nested classes, whose
methods get the full owner path. It is metadata only and never participates in name resolution, so
the `ClassName`-based logic in the semantic analyzer and the parser is untouched.

This closes the qualifier bug that the 2026-09-09 contract-inheritance entry (L-S3a–L-S3c)
recorded as **not done, deliberately** — it was left out then because `contracts_subproc` also
needed the column work, so fixing it alone would not have flipped the fixture.

### A contract failure has no innermost frame

`raiseContractException` positioned the exception at the failing condition, which `OriginPos`'s
predecessor then rendered as an extra frame — `RequirePositive [line: 2, column: 11]` — that
DWScript does not emit. A failed `require`/`ensure` is built by the engine, not by a script-level
`EFoo.Create(…)`; it has no construction site, and the failing routine and its condition are
already named in the message text. The node argument is now nil.

### `ExceptObject.StackTrace` is nil-safe

`exceptobj3` reads `ExceptObject.StackTrace` outside any `except` block and expects an empty
line. In DWScript `StackTrace` is a magic getter, not a field dereference, so it answers `''` on a
nil reference; we raised `Object not instantiated`. `ExceptObject` is now bound as a *typed* nil
(`NilValue{ClassType: "Exception"}`) so the static class resolves, and the nil branch of member
access answers `''` for `StackTrace` on an Exception-typed nil. `ExceptObject = nil` still reads
`True` — a typed nil compares equal to nil.

### Scope

`raise <expr>` where the expression is not a call (a plain object reference, say) still reports
`End()` as its origin frame; no fixture pins that case. Strictly, DWScript captures the stack when
the exception is *constructed*, so `var e := EFoo.Create('x'); … raise e;` should trace to line of
the `Create`, not the `raise`. Modelling that needs the stack captured on the object at
construction time; it is not what any fixture measures today.

**Validation:** `go test ./...` green. New tests:
`internal/interp/evaluator/call_site_test.go` (table-driven `callSitePos` over all call shapes,
and `qualifiedRoutineName` over out-of-line / inline / free routines),
`internal/interp/runtime/exception_test.go` (`OriginPos` precedence, fallback, and the no-frame
case), `internal/interp/stack_trace_positions_test.go` (four end-to-end traces), and two parser
tests for `DeclaringClassName` including the nested-class path. Baselines ratcheted
(`SimpleScripts` 343 → 346) and `TEST_STATUS.md` regenerated.

## 2026-09-10 — Record copy-on-assign value semantics (§3.3)

Closed the `PLAN.md` §3.3 item *"Record copy-on-assign value semantics"*.
**JSONConnectorPass 59 → 60 of 82**; overall 957 → 958 of 1,928 scored. No other category moved.

### The bug

`r.BottomRight := p` stored `p`'s `*runtime.RecordValue` by reference, so the later `p.y := 3`
was visible through `r.BottomRight`. JSONConnectorPass `stringify_record` printed
`{"BottomRight":{"x":1,"y":3},…}` instead of `{"BottomRight":{"x":1,"y":2},…}`.

`RecordValue.Copy()` was never the problem — it deep-copies fields through `CopyValue`. The
copy simply was not being made. `evalAssignment` copies records on the identifier path
(`x := rec`) and `index_assignment.go` copies them through `cloneIfCopyable` on the index path
(`a[i] := rec`), but the `*ast.MemberAccessExpression` path went straight to
`evalMemberAssignmentDirect` with the live value. Every member write was affected — record
fields, record properties (`executeRecordPropertyWrite`), object fields, and nested targets —
not only the one the fixture happened to expose.

The `*ast.IndexExpression` aliasing exception in `prepareValueForAssignment` was investigated
and left alone: it is reached only after `evalAssignment` has already copied records, so it
governs static arrays only and is not part of this bug.

### The fix

`evalMemberAssignmentDirect` (`internal/interp/evaluator/member_assignment.go`) copies a
`*runtime.RecordValue` RHS before doing anything else. It is the single choke point for member
writes, so field, property, object-field and nested paths are all covered by the one copy,
and it mirrors what the identifier path in `evalAssignment` already did.

### Deterministic record JSON keys (latent, fixed while here)

The context-free `ValueToJSONValue` fallbacks in `internal/interp/evaluator/json_helpers.go`
and `internal/interp/runtime/json_helpers.go` walked `RecordValue.Fields` as a Go map, so any
path reaching them emitted JSON keys in a random order per run. `RecordValue` gained
`OrderedFieldKeys()` and `FieldDisplayName()` (`internal/interp/runtime/record.go`); both
fallbacks now use them. The context-aware `recordToJSON` in `json_serialize.go` — which applies
visibility rules and runs property getters — is unchanged and remains the path record
serialization normally takes.

### Validation

- `go test ./...` green.
- New table-driven `TestRecordCopyOnAssign` in `internal/interp/record_copy_on_assign_test.go`
  covers simple variable, record field, record property, object field, dynamic array element,
  nested record targets, and the read-side copy. Four of its seven cases failed before the fix.
- `just fixture-report` diffed against `baselines.json`: exactly one category moved,
  `JSONConnectorPass` 59 → 60. Baselines ratcheted and `TEST_STATUS.md` regenerated.

## 2026-09-10 — Metaclass method pointers and intrinsic-member capture (PLAN.md §3.3)

Closes the §3.3 function-pointer niche item and its two acceptance fixtures,
`SimpleScripts/func_ptr_symbol_field` and `SimpleScripts/func_ptr_classname`.

Both fixtures are the same feature in two syntactic positions: a **parameterless class member
captured as a pointer instead of being read eagerly**. PLAN.md described the second half as a
"value ↔ parameterless-function coercion", but that framing does not survive contact with the
fixture — `TObject.ClassName` is not a `String` being widened into a `function : String`, it is
DWScript's intrinsic parameterless class member being *bound*. Modelling it as a pointer is what
makes `a1[i]()` an ordinary call and `a2[i]` an ordinary auto-invoke.

### The intrinsic members are not methods

`ClassName` exists on TObject as a *synthesized* overload (so `inherited ClassName` resolves) and
`ClassType` does not exist as a symbol at all — both are answered ad hoc by the member-access
paths. Neither is a class method, so `analyzeMethodReferenceInPointerContext` rejected them: the
metaclass branch demands `isClassMethodInHierarchy`. `analyzeAddressOfMethod` rejected the
metaclass receiver outright with `unbound method pointers (@TClass.X) are not supported`.

New `intrinsicMemberPointerType` (`internal/semantic/analyze_function_pointers.go`) forms the
pointer: `ClassName -> function : String`, `ClassType -> function : class of <receiver>`. A
user-declared method of the same name suppresses it (`firstNonSynthesizedMethod`), so an
override still owns the reference.

*Result-type covariance.* `FunctionPointerType.Equals` compares result types exactly, so
`function : class of TClassA` is not `function : TClass`. Rather than loosen pointer equality —
which is shared with `LambdaPass`, `DelegateLib` and `OverloadsPass` — the intrinsic adopts the
*declared* signature when the context supplies one and the intrinsic's own result is assignable
to it. `a1.Add(TClassA.ClassType)` therefore stores a `function : TClass`. With no context
(`var proc := @TObject.ClassType`) the natural signature is used.

`analyzeMethodReferenceInPointerContext` now takes the expected type so it can consult it, and
its annotation bookkeeping moved into `annotateMemberPointerType`. `analyzeAddressOfMethod`
resolves a metaclass receiver to its class (`addressOfReceiverClass`) and only reports
`unbound method pointers` when the member is neither an intrinsic nor a class method.

### Auto-invoke in the receiver position

`proc.ClassName` needs `proc` invoked before the member is looked up. The analyzer's
member-access path used `implicitCallReturnTypeFromType`, which only understands `FunctionType`;
it now uses the existing `implicitValueContextType`, which also covers parameterless function and
method pointers. `implicitCallReturnTypeFromType` itself was deliberately left alone — widening
it would make `var proc := @TObject.ClassType` infer `class of TObject` instead of a pointer.
`VisitMemberAccessExpression` gained the matching runtime step, skipped for a nil pointer so
member access on one keeps its current diagnostic.

### Runtime

An intrinsic member has no declaration to point at, so `runtime.FunctionPointerValue` gained
`IntrinsicMember` — a name dispatched against `SelfObject`. `IsNil` and `String` account for it;
`executeFunctionPointerDirect` routes it to `invokeIntrinsicClassMember`
(`internal/interp/evaluator/intrinsic_member_pointer.go`), which answers `ClassName`/`ClassType`
for either a class reference or an object instance.

Two sites create these pointers, both only when the analyzer annotated the node as a pointer
(`memberWantsMethodPointer`) or the `@` operator was used: `resolveClassMetaMember` and the
member branch of `VisitAddressOfExpression`, the latter extracted into `addressOfMember` so a
class-reference receiver binds a class-method pointer first and falls back to the intrinsic. The
declared-class-method-wins rule is enforced at both sites, matching the analyzer.

### Scope

`func_ptr_field_no_param` and `func_ptr_property` still fail; they are unrelated
(field- and property-typed pointers), not a metaclass or intrinsic-member problem.

Not fixed, and out of scope: an eager (non-pointer) read of `TThing.ClassName` where the class
declares `class function ClassName` still returns the builtin, because `resolveClassMetaMember`
answers the builtin before it reaches the class-method lookup. The instance path already guards
this with `userMethodHidesBuiltin`; the class path has no equivalent. Only the pointer path was
made consistent here.

**Validation:** `go test ./...` green. `golangci-lint run` reports 1205 issues, unchanged from
`main` — the two functions this work pushed over the `gocyclo` threshold were split
(`addressOfMember`, `addressOfReceiverClass`, `firstBindableMethodOverload`). New table-driven
tests in `internal/semantic/intrinsic_member_pointer_test.go` (accepted positions, and the
narrowness of the coercion: result mismatch, a target with parameters, and instance-method
`@TClass.Method` still unsupported) and `internal/interp/intrinsic_member_pointer_test.go`
(explicit call, auto-invoke on read, auto-invoke in receiver position, instance receiver,
user-method precedence). Baselines ratcheted (`SimpleScripts` 343 → 345; end-to-end
`just fixture-report` TOTAL 957 → 959 with no category regression).

> Rebased onto `main` after §3.2.7 and the ByteBuffer host type landed; the counts above are the
> re-measured post-rebase numbers, not the ones from the original branch point.

## 2026-09-10 — Associative arrays: ARC destructor timing and Variant → key coercion (PLAN.md §3.3)

`AssociativePass` 22 → 24 of 27 (81% → **89%**); corpus 938 → 940 of 1,928 scored, re-measured
after rebasing onto the §3.2.7 floor. Harness and CLI agree at 940. No category regressed. Closes the §3.3 item
"ARC destructor timing on associative slot replace/clear; Variant → key coercion"
(`delete_sequence`, `variant_key_cast`).

### Variant → key coercion

`a[v]` unwrapped the Variant but never converted it, so writing through a Variant holding
`123` into an `array [String] of String` stored an `IntegerValue` key that `a['123']` could
never find — `variant_key_cast` printed two of its three lines.

One helper, `coerceAssociativeKey` (`internal/interp/evaluator/associative_helpers.go`), now
serves all four key sites (read, the two write paths, and `Delete`). It unwraps the Variant,
then converts to `AssociativeArrayValue.KeyType()` through the existing machinery:
`TryImplicitConversion` first (user-registered and chained conversions), falling back to
`coerceValueToKind`, the same variant-cast rules builtin arguments use. Coercion is gated on
the index actually being a `runtime.VariantWrapper`, so a genuinely mistyped key keeps its
strict behavior and no non-Variant path changes.

### ARC destructor timing

`runtime.AssociativeArrayValue` had no refcount integration at all: overwriting a slot dropped
the old value on the floor, `Delete`/`Clear` dropped whole entries, and object keys were never
retained. `delete_sequence` printed no destructor output except the one that fired too early.

The map stays refcount-agnostic; it only exposes what the evaluator needs to do the ARC work:
`Set` returns the displaced value and whether it replaced a slot, `DeleteEntry` returns the
stored key and value, and `TakeEntries` empties the map and hands both slices over (a second
take yields nothing, so contents can never be released twice). The evaluator side reuses
`retainValueForBinding` / `releaseValueForBinding` — the same helpers named bindings use —
through `storeAssociativeEntry`, `releaseAssociativeEntry` and `releaseAssociativeContents`.
No second lifetime mechanism was introduced.

The resulting observable order matches DWScript: a slot overwrite destroys the displaced
value only (the key slot already holds its retained key); `Delete` and `Clear` release value
before key; program-scope finalization releases keys before values.

### Scope: why the release is not in `releaseValueForBinding`

Hooking assoc-array release into `releaseValueForBinding` — where the task originally pointed
— regressed `by_ref` and `parameters`. Associative arrays are *reference* types: a map passed
to a function is the caller's map, so releasing its contents when the callee's scope ends
empties a map that is still alive. Correct handling needs a refcount on the map value itself,
which does not exist yet and is out of scope here. Instead, `VisitProgram` runs a narrow
program-scope finalization (`releaseAssociativeBindings`) over the global environment. Global
finalization of plain object bindings remains a separate, still-open concern: an unbound
global object still gets no destructor at program end.

`array_of_dyn`, `elements_of_value` (nested lvalue vivification) and `records` (hash iteration
order) remain open under their own §3.3 items.

**Validation:** `go test ./...` green; `just fixture-report --category AssociativePass
--list-fails` 24/27; full `just fixture-report` 940 (was 938). New tests:
`internal/interp/runtime/associative_array_test.go`
(`TestAssociativeArray_SetReportsReplacedSlot`, `_DeleteEntry`, `_TakeEntries`) and
`internal/interp/associative_arc_test.go` (`TestAssociativeArrayARC`,
`TestAssociativeArrayVariantKeyCoercion`) covering slot replace, Delete, Clear, program-end
finalization, and four key-coercion directions. Baselines ratcheted and `TEST_STATUS.md`

## 2026-09-10 — Nested lvalue vivification through a key or index (PLAN.md §3.3)

Closed 2026-09-10 on branch `feat/plan-3.3-nested-lvalue-vivification`. Closes the `PLAN.md` §3.3
item of the same name in full: all four fixtures it named now pass.

### Headline

| | before | after |
| --- | --- | --- |
| CLI (`just fixture-report`) | 938 / 1,928 | 942 / 1,928 |
| AssociativePass | 22 / 27 | 24 / 27 |
| JSONConnectorPass | 59 / 82 | 61 / 82 |

No category moved down. Every other category is byte-identical between the two reports.

### The two defects

**`evaluateLValueIndex` only understood plain arrays.** `internal/interp/evaluator/var_params.go`
handled `*runtime.ArrayValue` and required a literal `*runtime.IntegerValue` index, so
`ra[2].S := 'hello'` on `var ra : array[Integer] of TRecord` died with
`cannot index into ASSOCIATIVE_ARRAY`.

**No vivification on a missing key.** `VisitIndexExpression` returned
`getZeroValueForType(assoc.ElementType())` for an absent key *without inserting the slot*. That is
right for a pure rvalue read (`PrintLn(a['nope'])` must not grow the map) and wrong everywhere the
value is only an intermediate step: `sa[1][1] := 123` wrote into a throwaway static array and
`a['alpha'].Add('beta')` appended to a throwaway dynamic array — both silently lost.

### The seam

There was none: `evalIndexAssignmentDirect` deliberately handled only the outermost index and got
its base with `e.Eval(target.Left, ctx)` — an **rvalue** evaluation — then mutated whatever came
back.

The tail of `VisitIndexExpression` was split into `indexResolvedValue(leftVal, node, ctx)`, which
applies one level of indexing to an *already-evaluated* container. On top of that,
`internal/interp/evaluator/lvalue_vivify.go` adds `resolveLValueContainer`: it resolves the
container part of a nested lvalue, recursing through `IndexExpression` bases and vivifying a
missing associative-array slot, and falling back to ordinary read semantics for everything else
(arrays, objects, records and JSON already hand back live references). The base is resolved once,
so a side-effecting index expression is still evaluated exactly once.

Four call sites now resolve their base through it instead of `Eval`: `evaluateLValueIndex`,
`evaluateLValueMember`, `evalIndexAssignmentDirect`, and the receiver of
`VisitMethodCallExpression` — the last because a method may mutate its receiver in place, which is
what `a['alpha'].Add('beta')` needs. `VisitIndexExpression` itself is untouched, so a pure rvalue
read still does not insert.

`evaluateLValueIndex` additionally learned associative arrays (vivify, then write back with
`assoc.Set`) and JSON (`indexJSON` / `assignJSONIndex`); `evaluateLValueMember` learned JSON
members (`evalJSONValueMember` / `assignJSONMember`). The member-rooted branch of
`evalIndexAssignmentDirect` gained the JSON index write it was missing, which is what
`v.List[0] := 'zero'` needs.

### Overlap with the ARC / key-coercion work

`feat/plan-3.3-assoc-arc-key-coercion` (PR #374) had not landed on `main` when this branch was
cut, so `coerceAssociativeKey` and the retain/release `storeAssociativeEntry` path did not exist
yet. Key handling here stays minimal — `unwrapVariant`, matching what `evalIndexAssignmentDirect`
already did — and vivification inserts through plain `AssociativeArrayValue.Set`. Whichever branch
lands second should route `vivifyAssociativeSlot` (`lvalue_vivify.go`) through the coercing and
retaining insert, so a vivified slot is owned like any other.

### Scope

The three remaining `AssociativePass` failures are other §3.3 items: `records` (hash iteration
order), `delete_sequence` (ARC destructor timing) and `variant_key_cast` (Variant → key coercion).
The 21 remaining `JSONConnectorPass` failures are node reparenting/ownership, float formatting and
cast-message parity — the separate §3.3 JSON item.

**Validation:** `go test ./... -timeout 30m` green; `golangci-lint run --new-from-rev=main` reports
0 issues; `just check-fmt` clean. New table-driven tests in
`internal/interp/lvalue_vivification_test.go` cover both directions of the rule — lvalue and
mutating-receiver positions insert, rvalue reads do not — plus slot reuse, whole-element
replacement, compound assignment, plain nested arrays, and the three nested-JSON forms. Baselines
ratcheted (`AssociativePass` 22 → 24, `JSONConnectorPass` 59 → 61) and `TEST_STATUS.md`
regenerated.

## 2026-09-10 — DWScript hash iteration order for associative arrays (PLAN.md §3.3)

`AssociativePass/records` printed `a,b` where DWScript prints `b,a`. Nothing about the fixture is
sorted or reversed: DWScript's `array [K] of V` is an open-addressing hash table, `.Keys` walks the
bucket array in index order, and the resulting order is a deterministic function of the hash, the
table size and the probe sequence. Our `AssociativeArrayValue` was two parallel slices in insertion
order with an O(n) linear scan, and its doc comment recorded insertion order as a deliberate choice.

Matching that order means porting the real table, not picking a permutation that satisfies one
fixture. `reference/dwscript-original/` is empty in this checkout, so the algorithm was taken from
upstream directly:

- `Source/dwsAssociativeArrays.pas` — `TScriptAssociativeArray`: capacity starts at 32 and doubles,
  the table grows once `count >= capacity*11/16`, bucket index is `hash and (capacity-1)` with
  linear probing, a hash code of `0` marks an empty bucket, and `CopyKeys` walks buckets
  `0..capacity-1`. `Delete` uses backward-shift deletion, which is required: blanking a bucket
  outright would cut the probe chain of any entry that had collided with it.
- `Source/dwsDataContext.pas` — `DWSHashCode`: FNV-1a mixing (`basis 2166136261`, `prime 16777619`)
  over per-slot hashes, substituting the basis whenever the result lands on `0`.
- `Source/dwsUtils.pas` — `SimpleStringHash` is xxHash32 (seed 0) over the string's UTF-16 code
  units; `SimpleInt64Hash` and `SimpleIntegerHash` are simplified MurmurHash3 finalizers.

### Why we believe the order is portable

The corpus pins associative-array iteration order in exactly two fixtures, and the port reproduces
all three orders in them without any fitting:

| Fixture | Key type | Inserted | DWScript expects | Port yields |
| --- | --- | --- | --- | --- |
| `AssociativePass/records` | String | `a`, `b` | `b,a` | `b,a` |
| `JSONConnectorPass/associative_array` | Integer | `10, 11, 20, 21` | `21,11,10,20` | `21,11,10,20` |
| `JSONConnectorPass/associative_array` | String | `1.2`, `3.4` | `1.2,3.4` | `1.2,3.4` |

Three independent orders — two key types, one of them a four-element permutation that is neither
insertion nor sorted order — falling out of one hash model is what distinguishes a port from a
guess. Every other fixture that touches `.Keys` either sorts (`keys`, `array_of_dyn`, `delete`,
`delete_record`, `parameters`), aggregates commutatively (`integer_float`), or holds at most one key
at the moment of enumeration (`stringify_keys`, `keys_str_int`).

### Known divergences

Both are unreachable from the corpus and are recorded here rather than papered over:

- **Record keys.** Upstream hashes a record's data slots in declaration order. `types.RecordType`
  stores fields in a map with no declaration order, so record keys are flattened in sorted field
  order instead. Deterministic, but not upstream's bucket order.
- **Object keys.** Upstream hashes the low 32 bits of the object's interface pointer, so its
  object-key order is heap-address dependent and not reproducible between two runs of DWScript
  itself. We substitute a monotonic per-instance identity, which keeps lookups correct and the
  order deterministic within a run.

### Side effect

Key lookup, insertion and deletion are no longer O(n).

**Validation:** `go test -count=1 ./...` green; `golangci-lint run --new-from-rev=main` reports 0
issues; `just check-fmt` clean. New table-driven tests in
`internal/interp/runtime/associative_hash_test.go` pin the xxHash32 port against the published
reference vectors (covering the four-accumulator path for inputs of 16 bytes or more, which no
fixture reaches), pin the hash and bucket index of every key in the order-sensitive fixtures, and
cover growth, backward-shift deletion, per-key-kind round-trips and the
equal-keys-hash-equally invariant. Full fixture report against the branch point: `AssociativePass`
22 → 23, total 920 → 921, no category down. Baselines ratcheted and `TEST_STATUS.md` regenerated.
