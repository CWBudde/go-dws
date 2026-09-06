# Progress log — September 2026 (PLAN.md Phase 1: measurement & tooling)

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
nor the frontend resolves program-level `uses` yet (new `PLAN.md` §3.2 item).

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
