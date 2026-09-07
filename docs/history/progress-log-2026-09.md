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
bare name means a call or a reference, and procedures keep typing as `VOID`. That admits
exactly the 14 `Sig(nil, …)` builtins: `Pi`, `Infinity`, `NaN`, `Random`, `RandSeed`, `Now`,
`Date`, `Time`, `UTCDateTime`, `UnixTime`, `UnixTimeMSec`, `GetStackTrace`, `GetCallStack`
(and `Randomize`, which stays `VOID` — it is a procedure). The existing builtin
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
