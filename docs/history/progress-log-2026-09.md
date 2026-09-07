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
packages pass. The fixture regression gate and regenerated status remain **871/1,928**,
with identical pass-count baselines in all 61 categories. The status generation date
was refreshed; no baseline floor changed. `go vet ./internal/... ./pkg/... ./cmd/...`
and the new evaluator context-isolation tests under `-race` pass. Actual compile/run
help output confirms the experimental bytecode labels.

The complete CLI integration suite also passes (10.652s). Its repeated executable
builds exceeded the standard timeout on the workspace's NTFS mount, so it was run
against a verified identical source copy on a native filesystem, using the same test
scripts and `GOFLAGS=-buildvcs=false`. No test assertions were disabled.
