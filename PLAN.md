# DWScript → Go Port — Implementation Plan

> **This document was rewritten on 2026-07-02 to be an honest single source of truth.**
> The previous 3,672-line version marked large amounts of unfinished work "COMPLETED" and
> prioritized far-future compiler backends over core language compatibility. It is preserved
> in git history (the commit prior to this one) if you need the old task-level detail. The
> supporting evidence for everything here is in [`docs/CODEBASE_REVIEW_2026-07.md`](docs/CODEBASE_REVIEW_2026-07.md).

## Status legend

- ✅ **Done** — verified by tests **and** fixtures, not by assertion.
- 🟡 **Partial** — works for the common case; known gaps listed.
- 🔴 **Broken/Absent** — does not work or does not exist.
- ⏸️ **Parked** — deliberately deferred; do not start until gate conditions are met.

**Rule for this document:** a task may only be marked ✅ when there is a passing *fixture*
(or a test that exercises the real user-facing path) demonstrating it. "The unit test passes"
is not sufficient — see §1.

---

## 1. Current reality (measured 2026-07-02, `HEAD` of this branch)

| Metric | Value |
|---|---|
| Fixture pass rate (exact-match, all categories) | **410 / 1,928 = 21%** |
| Categories the in-repo Go harness **skips** (`skip: true`) | **48 of 61** |
| Core unit-test packages | all green (lexer 85% cov, parser 79.6%, semantic 57.4%) |

**The central problem is not any single feature — it is that the green unit suite hides a 21%
real pass rate.** The fixture harness skips 4 of every 5 categories, so CI stays green while
most real DWScript programs produce wrong output. Fixing the *measurement* is prerequisite to
fixing anything else.

### Failure breakdown (should-pass categories, 597 failures)

| Share | Cause | Where the fix lives |
|---|---|---|
| **65%** | Front-end rejects valid code | parser + semantic |
| 15% | Wrong output (logic) | evaluator |
| 11% | Runtime panic/abort | evaluator |
| 9% | Missing `Errors >>>>` hint envelope | semantic |

### Compatibility by area (selected)

| Area | Status | Pass% |
|---|---|---|
| SimpleScripts, Algorithms, Math, String, Interfaces, Operators | 🟡 | 54–96% |
| Arrays | 🟡 | 23% |
| Sets, Helpers, Overloads | 🟡 | 13–26% |
| **Generics, Lambdas, Associative arrays, Property expressions, JSON** | 🔴 | **0%** |
| **All `*Fail` error-detection suites** | 🔴 | **0%** |
| Host libraries (DB, Crypto, COM, Web, Graphics) | ⏸️ | 0% (need host bindings; out of core scope) |

---

## 2. Architecture (as-built, verified)

```
Source → Lexer → Parser → AST → Semantic Analyzer ─┬─→ AST Evaluator (LIVE, production)
                                                    └─→ Bytecode Compiler → VM (BROKEN, opt-in)
```

- **Live execution path:** `cmd/dwscript run` → `pkg/dwscript` → `internal/interp.Interpreter`
  (thin shell) → `internal/interp/evaluator` (self-contained, ~22.7k LOC). This path is clean.
- **Dead weight still in the tree:**
  - ~14k LOC of **shadow evaluator** in `internal/interp` (old `eval*` methods) — 126
    production-unreachable functions, kept alive only by tests. The "collapse to a single
    engine" phase was reported complete but the deletion never happened.
  - `internal/bytecode` — a parallel VM that cannot compile `for`/`case`, forks the value
    model and the builtins, and does not deliver the advertised speedup.
- **Type system is triplicated:** `internal/types` (static) vs `internal/interp/types`
  (runtime facade, `any`-typed) vs `internal/interp/runtime` (value structs, string-typed
  metadata). This is the deepest structural flaw.

---

## 3. Roadmap (reprioritized — highest leverage first)

### P0 — Make the harness tell the truth ✅ *(done 2026-07-03)*

Goal: a green CI run must mean "the language works," not "the parts we test work."

- [x] `scripts/fixture_report.py` wired into `just` (`fixture-report`) and CI
      (`.github/workflows/test-fixtures.yaml`, called from `test.yml`).
- [x] **Un-skipped all categories** in `internal/interp/fixture_test.go`. The binary `skip`
      flag is gone; categories are now **auto-discovered** from `testdata/fixtures/` (so none
      can be silently omitted again) and each is gated against a **per-category pass-count
      baseline** in `testdata/fixtures/baselines.json`. Individual fixture failures are recorded
      but do not fail the build — the build fails only when a category drops *below* its
      baseline. Each fixture runs in a re-executed **worker subprocess pool** (see
      `TestFixtureWorkerMain`) so a runaway loop or pathological allocation is killed and
      isolated instead of OOM-ing the whole `go test` process.
- [x] `testdata/fixtures/TEST_STATUS.md` is now **generated** by the harness in update mode
      (`just fixture-update`) — never hand-maintained again.
- [x] CI gate: `go test -race ./...` (via `just test-unit`) runs `TestDWScriptFixtures`, which
      fails the build if any category's pass count drops below its baseline. Ratchet up with
      `just fixture-update` after an intentional improvement.
- **Exit criteria met:** real per-category numbers are visible in CI on every push (the `-v`
      category logs plus the ground-truth `fixture-report` job). Current Go-harness baseline:
      **547 / 1,928 scored = 28%** (the harness scores the `*Fail` error-detection suites the
      CLI-only `fixture_report.py` cannot, e.g. FailureScripts 102/528).

### P1 — Front-end coverage 🔴 *(attacks the 65% — ~440 fixtures)*

- [ ] **Parse generics.** `type TList<T> = class … end;` and generic methods currently fail at
      the parser (`GenericsPass` = 0/23, entirely parser-blocked).
- [x] **Forward class declarations.** `type TChild = class;` then a later full definition no
      longer errors "already declared"; the evaluator now registers a placeholder that the full
      declaration completes. (`class_forward.pas` now passes end-to-end after the class-method
      dispatch fix below.)
- [x] **Class-method dispatch through an instance.** DWScript permits invoking a `class
      procedure`/`class function` on an instance (`obj.ClassProc`), by bare name inside another
      method (`ClassProc;`), and via `inherited` from a class method. All three previously failed
      with "field/method not found". Added `ObjectValue.GetClassMethodDecl` (backed by
      `IClassInfo.LookupClassMethod`) and routed member access, bare-identifier auto-invoke, and
      the `inherited` fallback through the class-method table. Fixes `class_forward.pas`,
      `class_method2.pas`, `method1.pas` (SimpleScripts 236→239, overall 413→416).
- [x] **Type-inferred class/record constants via metaclass.** `TBase.c1` for `const c1 = 1;`
      now resolves (`class_const2.pas`). Root cause was a nil-`*TypeAnnotation`-in-interface in
      the parser making untyped consts look explicitly (empty-)typed. Remaining: helper consts
      via metaclass (`String.Hello`) still fail at runtime (`HelpersPass/string_consts.pas`) — P4.
- [ ] **Hint/warning envelope.** Emit DWScript's test serialization: when compilation produces
      hints/warnings, wrap output as `Errors >>>>\n<hints>\nResult >>>>\n<program output>`.
      **Investigated 2026-07 (measured, not shipped).** Findings, so the next attempt starts from
      fact:
      - **Scope:** 90 fixtures carry the envelope; **50** already match on the `Result >>>>`
        section (pure envelope+hints), of which **22 need only case-mismatch hints**. The other 40
        also have unrelated output bugs.
      - **Infra is easy:** the analyzer already runs pedantic in `pkg/dwscript` and separates
        hints via `hasActualErrors()`; a prototype that set `HintsLevelPedantic` in the CLI
        (`cmd/dwscript/cmd/run.go`) and emitted the wrapper worked. Hints must be **deduplicated**
        by exact `(message,line,col)` — the analyzer currently emits each case hint ~3× (the
        identifier is re-analyzed as lvalue/type/member) — and **sorted by (line,col)**.
      - **Blocker is hint *precision*, not the envelope.** Full pedantic+envelope measured
        **+17 / −21** (net −4). Regression sources: **unused-var/field warnings** (17, too
        aggressive), **case hints on builtins** (`println`→`PrintLn`, `PI`→`Pi`; DWScript emits
        none), and **case hints on user symbols** DWScript does not flag (shadowed locals — e.g.
        `result` is hinted in `record_result` but must NOT be in `crc32`; local-var declaration
        casing `Swaps`/`Test`; helper/interface/record members `IncX`/`X`/`vHello`).
      - **Work order to make it zero-regression:** (1) dedup+sort hints at emission; (2) suppress
        case hints when the resolved symbol is a builtin; (3) fix user-symbol case detection to use
        the *actual* in-scope declaration (handles shadowing) and skip member/param kinds DWScript
        ignores; (4) gate unused-var/field warnings until they match DWScript; (5) emit only behind
        a `run` flag the fixture harness passes (the `Errors >>>>` wrapper is test-harness framing,
        not production CLI output). Expected yield once precise: ~22 immediately, up to ~50.
      - **⚠️ Blocked — not reproducible from sources (concluded 2026-07).** The fixtures encode
        case hints *inconsistently* with no signal recoverable from the `.pas` source. Proof:
        `ArrayPass/array_in2.pas` and `Algorithms/hanoi.pas` both call **lowercase `println`**;
        array_in2's `.txt` emits `"println" does not match … ("PrintLn")` on every line, hanoi's
        emits **nothing** — and neither source carries a `{$HINTS}`/`{$WARNINGS}` directive or any
        other differentiator. The original DWScript test runner set hint levels per test
        (config/harness state), which was lost when the fixtures were captured as source+output
        pairs. Regressions from enabling pedantic land in *both* hinted and non-hinted categories
        (11 Algorithms, 5 SimpleScripts, 2 HelpersPass, …), so no global level and no
        per-category rule can be zero-regression. Case hints are also purely cosmetic — DWScript
        is case-insensitive, so execution is identical with or without them. **Do not pursue
        case-mismatch parity** unless the original per-test hint configuration is recovered; the
        non-case parts of the envelope (empty-block, unreachable-code, prefer-ToString) may still
        be worth revisiting individually if a source-level signal exists for them.
- [ ] Semantic hardening: analyze **unit function bodies** (`internal/semantic/unit_analyzer.go:158`),
      check **subrange bounds at compile time**, make the analyzer **multi-pass** so forward
      references resolve without explicit `forward`.
- **Exit criteria:** SimpleScripts ≥ 85%, GenericsPass/LambdaPass/PropertyExpressionsPass ≥ 50%.

### P2 — Collapse the type system to one representation 🔴

- [ ] Make `internal/types` the **single source of truth** for all resolved types.
- [ ] Replace the `any`-typed registries in `internal/interp/types/type_system.go:555-559`
      with typed references to `internal/types`, breaking the import cycle through a shared
      `internal/interp/contracts` package rather than `any`.
- [ ] Delete the duplicate runtime `*TypeValue` structs (`interp.EnumTypeValue` ≈
      `runtime.EnumTypeValue`, etc.); stop carrying type identity as strings in
      `internal/interp/runtime/metadata_conversion.go`.
- **Exit criteria:** exactly one type representation; no `any` in the type registries; the
      `AGENTS.md` "typed runtime structures over AST-shaped maps" guardrail actually holds.

### P3 — Delete the dead weight 🟡

- [ ] **Deduplicate helper registration (root fix for a defanged non-determinism smell).** A user
      helper is registered as **two** distinct `*runtime.MutableHelperInfo` instances — one by
      `TransferHelpersFromSemanticAnalysis` (`internal/interp/helpers_transfer.go:19`) and one by
      the evaluator's `VisitHelperDecl` (`internal/interp/evaluator/visitor_declarations.go`).
      Because the helper registry is slice-valued and `RegisterHelper` **appends** (never
      overwrites — `internal/interp/types/type_system.go:447`), both survive. This already caused
      one real bug (method impls patched into only one copy → non-deterministic dispatch, fixed
      under P4). Three first-match-over-`AllHelpers()` lookups remain that pick one of the two
      copies by Go map order: `lookupMutableHelper` feeding `executeInheritedHelperCallDirect`
      (`call_helpers.go:236`), and the parent-helper linkage loops at
      `visitor_declarations.go:1074` and `helpers_validation.go:44`. They are currently
      **behavior-neutral** (verified: the whole HelpersPass category is byte-stable across 6×
      re-runs) only because downstream `inheritedHelperCandidates` rebuilds a name-deduped
      candidate set — a fragile invariant. The correct fix is to stop creating two instances
      (make `RegisterHelper` replace a same-name/same-target helper, or have the evaluator reuse
      the transferred instance), after which all three lookups become trivially single-valued.
      Two same-name instances are otherwise indistinguishable by any stable key, so per-lookup
      "deterministic pick" hardening is not a real fix.
- [ ] **Shadow interpreter:** re-point the tests that call `interp.evalClassDeclaration` /
      `evalIntegerBinaryOp` / `evalEnumDeclaration` / set & operator helpers at the evaluator,
      then delete those bodies (`expressions_binary.go`, `enum.go`, `type_alias.go`, `set.go`
      operators, `declarations.go` class/interface builders). Mechanical, not a rewrite.
- [ ] Move `Evaluator.currentNode` into `ExecutionContext`; remove the double `MethodRegistry`
      allocation (`internal/interp/interpreter.go:61`).
- [ ] Split the evaluator god files (`visitor_statements.go` 1461, `visitor_declarations.go`
      1426, `var_params.go` 1112).
- [ ] **Bytecode decision:** hide `--bytecode` and `pkg/dwscript.CompileModeBytecode`, delete
      the rigged benchmarks and the "5-6x faster" claims from README/docs. Then **either**
      delete `internal/bytecode` outright **or** (if a VM is a real goal) rebuild it on the
      shared `internal/builtins` registry and `internal/interp/runtime` values — do not keep
      extending the current fork.
- **Exit criteria:** `deadcode ./cmd/...` reports 0 unreachable funcs in `internal/interp`; no
      public API exposes a non-working execution mode.

### P4 — Interpreter logic bugs & panics 🟡

- [x] **Non-deterministic helper-method dispatch.** A type/record helper was registered as two
      distinct runtime instances (semantic-transfer path + evaluator path); the method
      *implementation* patched only the instance that Go's randomized map iteration happened to
      return first, so the other instance kept the empty forward declaration. Dispatch then
      picked between "runs the body" and "returns a zero value" at random. `VisitFunctionDecl`
      now binds the implementation into **every** matching helper instance
      (`lookupAllMutableHelpers`). Fixes 6 previously-flaky HelpersPass fixtures deterministically
      (`array_length_helper`, `class_helper`, `classname_helper2`, `implicit_self_class_helper`,
      `implicit_class_self_class_helper`, `implicit_self_record_helper`); HelpersPass 7→13,
      overall 416→422, and removes a source of measurement noise (P0 concern).
- [ ] Work the 88 wrong-output fixtures (e.g. `casts_base_types` rounding, `case_variant`).
- [ ] Work the 68 runtime-panic fixtures (metaclass `ClassName`, class-method dispatch, `class of`).
- [ ] Post-exception continuation semantics (`assigned.pas` expects execution to continue after a
      caught runtime error) and BOM-preserving output.
- **Exit criteria:** ArrayPass/SetOfPass/HelpersPass/OverloadsPass ≥ 80%.

### Deferred ⏸️ — do not start until core compatibility ≥ 80%

These were phases 16–27 in the old plan. They are legitimate long-term ideas but must not
compete with core language work:

- ⏸️ Go source-code generation / AOT compiler
- ⏸️ JavaScript backend
- ⏸️ LLVM backend
- ⏸️ MIR foundation
- ⏸️ WebAssembly AOT compilation
- ⏸️ AST-driven formatter
- ⏸️ Host-library bindings (DB / Crypto / COM / Graphics / Web) — needed for the 0% lib suites,
      but they are integration surface, not language correctness.

---

## 4. Definition of "done" for the port

The port is **v1.0-worthy** when:

1. `scripts/fixture_report.py` reports **≥ 90%** on all non-host-library categories.
2. All `*Fail` suites reproduce DWScript diagnostics (error-detection parity).
3. There is exactly **one** type representation and **one** evaluator in the tree.
4. CI fails on any per-category regression.
5. No public API or CLI flag exposes a non-functional mode.

Track progress against **fixture pass rate**, not phase checkboxes.

---

## 5. How to reproduce the status numbers

```bash
go build -o bin/dwscript ./cmd/dwscript
python3 scripts/fixture_report.py                 # full per-category table + total
python3 scripts/fixture_report.py --category SimpleScripts --list-fails
```

Every status figure in this document came from that script on the current branch head.
