# go-dws Codebase Review — July 2026

**Reviewer:** automated deep audit (5 focused sub-audits + independent fixture measurement)
**Scope:** whole repository, with emphasis on why the DWScript fixture suite mostly fails.
**Method:** every claim below is backed by measurement or a `file:line` reference. Nothing is taken from `PLAN.md` on faith.

---

## 1. Executive summary

go-dws is a **large, real, and partially working** DWScript port. The lexer and parser are competent, the AST evaluator is a clean single execution path with sensible signal-based control flow, and the semantic analyzer genuinely type-checks. The unit-test suite is green.

It also has a **serious credibility gap between its unit tests and its real behavior**. The comprehensive DWScript fixture suite — the actual definition of "does this language work" — passes at roughly **21%**, and entire feature areas pass at **0%**. The green unit suite hides this because the fixture harness **skips 48 of its 61 categories**.

The problems are concentrated and diagnosable, not diffuse. **The dominant one is the front-end rejecting valid code** (65% of failures in should-pass categories), not interpreter bugs. Secondary problems are structural debt the project already half-knows about: a triplicated type system, a fully-duplicated dead "shadow" interpreter, and a bytecode VM that cannot run basic programs yet is advertised as "5-6x faster."

**None of this requires a ground-up rewrite.** It requires: (1) making the test harness tell the truth, (2) closing front-end gaps, (3) collapsing the type system to one representation, and (4) deleting two large piles of dead/broken parallel code. The bones are good; the accounting is dishonest and the edges are unfinished.

---

## 2. Measured reality (reproduce with `scripts/`-style harness, see §8)

Independent harness: run every `*.pas` fixture through `./bin/dwscript run`, compare normalized output to the sibling `*.txt`, exact match.

| Metric | Value |
|---|---|
| Scored fixtures | 1,928 (2,042 incl. 114 with no expected output) |
| **Pass** | **410 (21%)** |
| Fail | 1,518 |
| Fixture categories skipped by the in-repo harness (`skip: true`) | **48 of 61** |
| Core-package unit tests | **all green** (lexer 85% cov, parser 79.6%, semantic 57.4%) |

**The paradox:** every unit-test package reports `ok`, yet 4 of every 5 real DWScript programs produce wrong output. A green build here means very little.

### Category scorecard (selected)

| Working reasonably | Pass% | Broken / absent | Pass% |
|---|---|---|---|
| Algorithms | 96% | AssociativePass (27) | 0% |
| FunctionsMath | 63% | GenericsPass (23) | 0% |
| FunctionsString | 62% | LambdaPass (6) | 0% |
| InterfacesPass | 61% | PropertyExpressionsPass (19) | 0% |
| SimpleScripts | 54% | JSONConnectorPass (82) | 0% |
| OperatorOverloadPass | 62% | **All `*Fail` suites** (~640) | 0% |
| ArrayPass | 23% | All host-lib suites (DB/Crypto/COM/Web) | 0% |

The `*Fail` suites test **error detection**. 0% there means go-dws does not reproduce DWScript's compile-time diagnostics — partly missing checks, largely a missing **hint/warning envelope** (see §3).

---

## 3. Root-cause analysis

Bucketing the 597 failures across should-pass categories by failure signature:

| Count | % | Bucket | Meaning |
|---|---|---|---|
| **387** | **65%** | compile-error on valid code | The front-end **rejects legal DWScript** before it runs |
| 88 | 15% | wrong output | Interpreter logic bug |
| 68 | 11% | runtime panic/abort | Interpreter crashes on legal code |
| 52 | 9% | missing hints envelope | No `Errors >>>>` / `Hint:` diagnostics emitted |
| 2 | <1% | BOM handling | Byte-order-mark output mismatch |

**The headline: the interpreter is not the main problem. The front-end is.** Confirmed concrete gaps that reject valid programs:

- **Generics don't parse at all.** `type TList<T> = class end;` → parser error (`expected '=' after type name`). Entire `GenericsPass` (23) dead at the parser.
- **Forward class declarations crash.** `type TChild = class;` … later `type TChild = class(TBase)` → `class 'TChild' already declared` (`testdata/fixtures/SimpleScripts/class_forward.pas`).
- **Class constants via metaclass unresolved.** `TBase.c1` where `c1` is a class const → *"no accessible member with name c1 for type class of TBase"* (`class_const2.pas`), despite `PLAN.md` Task 9.6 claiming class-const support is done.
- **Missing hint/warning envelope.** 62 of 437 `SimpleScripts` expected outputs begin with `Errors >>>>` + `Hint: "X" does not match case of declaration` etc. go-dws emits neither the envelope nor case-mismatch hints by default (hints are Pedantic-gated).

---

## 4. Subsystem findings

### 4.1 Lexer — solid
Competent, 85% statement coverage, handles hex/binary/float/escaped & doubled-quote strings, case-insensitive keywords. No material gaps found. **Grade 8/10.**

### 4.2 Parser — good, one big hole
Pratt parser, 79.6% coverage, almost no `TODO`/`panic`, cleanly parses arrays, enums, records, dynamic-array literals. **Generics are the notable structural gap** (`type T<X>` unparseable), plus some expression/lambda forms. Incremental fixes, not a rewrite. **Grade 7/10.**

### 4.3 Semantic analyzer — real, but partial
A genuine type checker: detects assignment mismatch, undeclared identifiers, arg count/type, const assignment, interface non-satisfaction, and does real overload scoring and const-folding (`cmd/dwscript/cmd/run.go:157`). Holes:
- **Single-pass** — forward references need explicit `forward`.
- **Unit function bodies are never analyzed** (`internal/semantic/unit_analyzer.go:158-161`).
- **Subrange bounds checked only at runtime** — `x := 20` for a `1..10` type compiles clean.
- **Hints/unused-var detection off by default** (Pedantic-gated) — directly causes the 52 missing-envelope failures.
- **Coverage only 57.4%** despite a large test file count.
**Grade 6/10.**

### 4.4 Type system — the core structural flaw (triplicated)
Three overlapping representations coexist:
1. `internal/types` — the real static type system (used by analyzer + bytecode).
2. `internal/interp/types` — a runtime facade whose class/record/interface/enum/helper entries are typed **`any`** to dodge an import cycle (`type_system.go:555-559`) → zero compile-time safety.
3. `internal/interp` + `internal/interp/runtime` — runtime value structs that **duplicate** each other (`interp.EnumTypeValue` ≈ `runtime.EnumTypeValue`), with type identity carried as **strings** (`runtime/metadata_conversion.go`) — exactly the "AST-shaped compatibility maps" the project's own `AGENTS.md` guardrail forbids.
This is the single most damaging design issue and the reason many member-resolution bugs are hard to fix. **Grade (coherence) 3/10.**

### 4.5 Interpreter / evaluator — clean live path, dead twin attached
The **production path is clean**: `cmd/dwscript run` → `pkg/dwscript` → `interp.Interpreter` (thin shell) → `internal/interp/evaluator` (62 `Visit*`, ~172 eval methods, 22,749 LOC, self-contained). Control flow is signal-based with only 3 genuine panics. **Correctness 8/10.**

But the **old evaluator was never deleted**: ~14,045 LOC of shadow logic still lives in `internal/interp` (`expressions_binary.go`, `enum.go`, `type_alias.go`, `set.go` operators, `declarations.go` class/interface builders). `deadcode` flags **126 production-unreachable functions** there; they survive only because **tests still call them directly**, masking the rot. `PLAN.md` Phase 4 ("Collapse To A Single Execution Engine ✅ COMPLETED 2026-03-10") is **false**: runtime migration is done, the deletion half is ~0% done. **Architecture cleanliness 6/10; refactor-completeness 3/10.**

### 4.6 Bytecode VM — advertised, not real
Cannot compile `for` loops or `case` statements (`internal/bytecode/compiler_statements.go:65` → `unsupported statement type`). On a 45-fixture sample the VM failed **42** that the AST interpreter ran (my own broader sample: 3/60 vs 36/60). It **forks** the value model (own `Value`/`ObjectInstance`/`Closure`, `bytecode.go:12-697`) and **forks the builtins** (~96 hand-rolled, does not import `internal/builtins`'s ~241). The **"5-6x faster"** claim is unbacked: the headline benchmark calls `runner.New(nil)` (full bootstrap) every iteration on the interpreter side (`vm_bench_test.go:122`), manufacturing the ratio; the fair benchmark shows ~2.4× and excludes compile time. **Grade 2/10. Recommendation: hide/delete.**

### 4.7 Builtins & embedding API — the good part
`internal/builtins` is a strong, shared registry (~241 functions). `pkg/dwscript` (Engine/Program/Result, FFI, symbol export, options) is coherent — its one blemish is exposing the broken `CompileModeBytecode` as a first-class public option. **Builtins 6/10, API 6/10.**

---

## 5. Grades

| Category | Grade | One-line justification |
|---|---:|---|
| **Language compatibility (fixtures)** | **3/10** | 21% pass; whole feature areas at 0%. |
| Lexer | 8/10 | Complete, well-covered. |
| Parser | 7/10 | Solid Pratt parser; generics unparseable. |
| Semantic analysis | 6/10 | Real checker; single-pass, unit bodies unchecked, hints off. |
| Type-system coherence | 3/10 | Three competing representations; `any`-typed registries. |
| Interpreter correctness | 7/10 | Works well; 88 logic bugs + 68 panics on valid code. |
| Runtime architecture | 5/10 | Clean live path polluted by a 14k-LOC dead twin. |
| Bytecode VM | 2/10 | Can't run `for`/`case`; forked; false perf claim. |
| Builtins coverage | 6/10 | Strong shared registry (interp side). |
| Embedding API | 6/10 | Coherent; ships a broken mode. |
| **Test-suite integrity** | **3/10** | Green suite hides 21% reality; 48/61 categories skipped. |
| **Documentation honesty (PLAN.md)** | **2/10** | Marks incomplete work "COMPLETED"; 27 phases, backends before basics. |
| Code quality / maintainability | 4/10 | Large dead duplication + several 1000+ LOC god files. |
| **Overall** | **≈4/10** | Good bones, dishonest accounting, unfinished edges. |

---

## 6. What to do (summary — full plan in `PLAN.md`)

Priority order, highest leverage first:

- **P0 — Make the harness tell the truth.** Un-skip the 48 categories, record real pass counts per category into a generated status file, and add a CI gate that fails on regression. You cannot fix what you refuse to measure. *(A green build must mean the language works.)*
- **P1 — Front-end coverage (attacks the 65%).** Parse generics; support forward class declarations; resolve class constants/members via metaclass; implement the hint/warning envelope (`Errors >>>>` + case-mismatch/unused hints) at least when the fixture expects it. This bucket alone is ~440 fixtures.
- **P2 — Collapse the type system to one representation.** Make `internal/types` the single source of truth; have the runtime hold resolved-type references (via a shared `contracts` package to break the cycle) instead of `any` bags and type-name strings. Delete the duplicate `*TypeValue` structs.
- **P3 — Delete the dead weight.** Remove the shadow `eval*` bodies in `internal/interp` after re-pointing their tests at the evaluator (mechanical). Hide `--bytecode`/`CompileModeBytecode`, delete the rigged benchmarks and the "5-6x" claims; then either delete `internal/bytecode` or rebuild it on the shared registry/runtime.
- **P4 — Interpreter logic bugs & panics.** Work the 88 wrong-output and 68 panic fixtures once the front-end stops masking them.
- **Deferred — the backend phases** (Go source-gen, JS, LLVM, WASM-AOT). Park these explicitly until core compatibility is ≥80%.

---

## 7. Bottom line

The library is **not** fundamentally unsound at the execution layer — the evaluator is decent and the parser is capable. It is **undermined by three things: a test harness that hides failure, a front-end that rejects a large fraction of valid programs, and structural debt (triplicated types, a dead interpreter twin, a non-functional VM) the docs pretend is finished.** Fix the accounting first, then the front-end; the rest is bounded cleanup, not a rewrite.

---

## 8. Reproducing the numbers

The independent harness runs each fixture through the CLI and exact-matches normalized output. This is now implemented as the pure-Go `cmd/fixture-report` tool (`go run ./cmd/fixture-report`, or `just fixture-report`), which:
1. walks `testdata/fixtures/*/*.pas`,
2. runs `./bin/dwscript run` with a timeout, decoding output as bytes (some fixtures emit non-UTF-8 / BOM),
3. normalizes (trim trailing WS per line, strip surrounding blank lines),
4. compares to the sibling `.txt`, and
5. prints a per-category pass/fail table + total.

This report was generated from exactly that procedure on the current `HEAD` of `claude/report-review-fixtures-nxxmy4`.
