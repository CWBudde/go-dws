# go-dws documentation index

Start here. Open work is in [`../PLAN.md`](../PLAN.md); this directory holds everything that
explains the language, the code, and the decisions behind it.

| Directory | What goes there | Maintained? |
| --- | --- | --- |
| [`guide/`](#guide-language-features) | User-facing reference for DWScript features as implemented in go-dws | yes |
| [`architecture/`](#architecture-how-the-code-is-built) | How the packages fit together, boundaries, audits | yes |
| [`decisions/`](#decisions-adrs-and-scope) | Architecture decision records and scope decisions | yes |
| [`wasm/`](#wasm) | Building and using the WebAssembly target and playground | yes |
| [`history/`](history/README.md) | Completed-work log and dated snapshots | no (append-only) |
| [`archive/`](archive/README.md) | Superseded notes, kept for their reasoning | no |

## Where the numbers come from

Status figures are generated, never hand-edited:

- `testdata/fixtures/TEST_STATUS.md` — per-category pass/fail from the Go harness (`just fixture-update`).
- `testdata/fixtures/baselines.json` — the per-category floor CI enforces.
- `just fixture-report` — end-to-end CLI numbers (`cmd/fixture-report`; rebuilds `bin/dwscript`, refuses a stale one).

## Guide (language features)

| File | Covers |
| --- | --- |
| [`guide/dwscript-features.md`](guide/dwscript-features.md) | Catalog of every upstream DWScript feature (the parity target list) |
| [`guide/builtins.md`](guide/builtins.md) | Built-in functions reference (`Format`, string/math/conversion helpers) |
| [`guide/control-flow.md`](guide/control-flow.md) | `break`, `continue`, `exit` semantics |
| [`guide/exceptions.md`](guide/exceptions.md) | `try`/`except`/`finally`, `raise`, exception classes |
| [`guide/contracts.md`](guide/contracts.md) | `require`/`ensure`/`old` design-by-contract |
| [`guide/enums.md`](guide/enums.md) | Enumerated types, ordinals, scoped enums |
| [`guide/variant.md`](guide/variant.md) | Variant type semantics and `VarType` codes |
| [`guide/helpers.md`](guide/helpers.md) | Type helpers (`helper for`) |
| [`guide/interfaces-guide.md`](guide/interfaces-guide.md) | Interfaces: declaration, implementation, casting, external |
| [`guide/lambdas.md`](guide/lambdas.md) | Lambdas, anonymous methods, closures |
| [`guide/operators.md`](guide/operators.md) | Operator overloading behavior mirrored from upstream |
| [`guide/json-type-mapping.md`](guide/json-type-mapping.md) | JSON ↔ DWScript type mapping |
| [`guide/encoders.md`](guide/encoders.md) | EncodingLib encoder classes (Base64, Base32, hex, UTF-8/16, URL, HTML) |
| [`guide/error-messages.md`](guide/error-messages.md) | Error message format, positions, stack traces; DWScript wire format appendix |
| [`guide/ffi-guide.md`](guide/ffi-guide.md) | Foreign function interface quick start |
| [`guide/ffi.md`](guide/ffi.md) | Full FFI reference |

## Architecture (how the code is built)

| File | Covers |
| --- | --- |
| [`architecture/interp-evaluator-steady-state.md`](architecture/interp-evaluator-steady-state.md) | **Canonical.** Package roles for `interp`/`evaluator`/`runtime`/`types`/`contracts`, ownership rules, allowed `interp` responsibilities (appendices A–C) |
| [`architecture/interp-evaluator-boundary.md`](architecture/interp-evaluator-boundary.md) | The one-way import rule `interp` ↛ `evaluator` and its single exception |
| [`architecture/audit-2026-09.md`](architecture/audit-2026-09.md) | Measured audit behind `PLAN.md` §2: sizes, dead code, import graph, type-system triplication, two compile pipelines, ranked refactoring candidates |
| [`architecture/implicit-self-resolution.md`](architecture/implicit-self-resolution.md) | How identifiers and implicit `Self` resolve inside methods |
| [`architecture/token-cursor.md`](architecture/token-cursor.md) | Parser token-cursor design |
| [`architecture/comment-preservation.md`](architecture/comment-preservation.md) | Lexer/AST comment and trivia preservation |
| [`architecture/semantic-passes.md`](architecture/semantic-passes.md) | **Superseded design**: a multi-pass analyzer that was never implemented; kept as input for `PLAN.md` §3.2 |
| [`architecture/benchmarking.md`](architecture/benchmarking.md) | Running benchmarks and profiling with pprof |
| [`architecture/ident-migration-guide.md`](architecture/ident-migration-guide.md) | Why and how to use `pkg/ident` for case-insensitive identifiers (see also `../pkg/ident/README.md`) |

## Decisions (ADRs and scope)

| File | Decision |
| --- | --- |
| [`decisions/out-of-scope.md`](decisions/out-of-scope.md) | What go-dws will not implement (COM, DB, graphics, file I/O, …); defines the excluded fixture categories |
| [`decisions/string-encoding.md`](decisions/string-encoding.md) | UTF-8-native strings (intentional divergence from UTF-16 upstream) |
| [`decisions/refcounting-design.md`](decisions/refcounting-design.md) | Reference counting for interface/object lifetimes in the runtime |
| [`decisions/var-parameters-design.md`](decisions/var-parameters-design.md) | By-reference parameters; appendix on type-alias transparency for arrays |
| [`decisions/delphi-to-go-mapping.md`](decisions/delphi-to-go-mapping.md) | How Delphi OOP concepts map onto Go structures |
| [`decisions/visitor-pattern.md`](decisions/visitor-pattern.md) | Generated type-safe AST visitor over reflection |
| [`decisions/bytecode-vm.md`](decisions/bytecode-vm.md) | Bytecode VM status: experimental, parked, no verified speedup |
| [`decisions/formatter-style-guide.md`](decisions/formatter-style-guide.md) | Canonical formatting rules for the (parked) formatter |
| [`decisions/formatter-ast-audit.md`](decisions/formatter-ast-audit.md) | AST position/trivia gaps the (parked) formatter must close |

## WASM

| File | Covers |
| --- | --- |
| [`wasm/BUILD.md`](wasm/BUILD.md) | Building the WebAssembly target |
| [`wasm/API.md`](wasm/API.md) | JavaScript API of the WASM module |
| [`wasm/PLAYGROUND.md`](wasm/PLAYGROUND.md) | The web playground |

## History and archive

- [`history/README.md`](history/README.md) — index of completed-work records. The most useful files are
  [`history/progress-log-2026-07.md`](history/progress-log-2026-07.md) (root causes and fixture names
  for everything closed in July 2026) and [`history/progress-log-2026-09.md`](history/progress-log-2026-09.md)
  (Phase 1 tooling: one pipeline, harness-mode CLI flags). [`history/CODEBASE_REVIEW_2026-07.md`](history/CODEBASE_REVIEW_2026-07.md)
  is the measured review that set the current priorities.
- [`archive/README.md`](archive/README.md) — superseded designs, audits, and parked backlogs
  (`CodeGenTODO.md`, bytecode VM design notes, adapter/EvalNode migration notes, fixture-failure analyses).

## Documentation outside `docs/`

| File | Covers |
| --- | --- |
| [`../README.md`](../README.md) | User-facing overview, embedding API, CLI |
| [`../AGENTS.md`](../AGENTS.md) | Repository guidelines for contributors and AI agents (`CLAUDE.md` and `GEMINI.md` include it) |
| [`../CONTRIBUTING.md`](../CONTRIBUTING.md) | Contribution workflow, parser conventions |
| [`../goal.md`](../goal.md) | Original 2025 porting charter (historical) |
| [`../testdata/fixtures/README.md`](../testdata/fixtures/README.md) | The imported DWScript fixture corpus, category by category |
| [`../pkg/ident/README.md`](../pkg/ident/README.md) | Case-insensitive identifier utilities |
| [`../internal/interp/errors/README.md`](../internal/interp/errors/README.md) | Runtime error types |
| [`../internal/interp/evaluator/USAGE_EXAMPLES.md`](../internal/interp/evaluator/USAGE_EXAMPLES.md) | Evaluator usage examples |
| [`../scripts/README.md`](../scripts/README.md) | Helper scripts |
| [`../examples/README.md`](../examples/README.md) | Embedding examples |
| [`../playground/README.md`](../playground/README.md) | Playground development |
| [`../npm/README.md`](../npm/README.md) | NPM package |

## Conventions

- A doc in `guide/`, `architecture/`, or `decisions/` must describe the code as it is. If it stops
  being true, fix it or move it to `archive/` with a banner.
- Completed work is written up once, in `history/progress-log-<date>.md`, and then removed from `PLAN.md`.
- Do not hand-edit status numbers anywhere; link to the generated files instead.
