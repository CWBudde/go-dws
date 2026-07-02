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

### P0 — Make the harness tell the truth 🔴 *(do this first)*

Goal: a green CI run must mean "the language works," not "the parts we test work."

- [ ] `scripts/fixture_report.py` exists (added in this change) — wire it into `just` and CI.
- [ ] **Un-skip the 48 skipped categories** in `internal/interp/fixture_test.go`; convert the
      binary skip into a **per-category expected pass-count baseline** so failures are visible
      but known gaps don't break the build.
- [ ] Generate `testdata/fixtures/TEST_STATUS.md` from the harness (it is currently full of
      `TBD` and was last hand-edited 2025-11-04). Never hand-maintain it again.
- [ ] CI gate: fail the build if any category's pass count **drops** below its baseline.
- **Exit criteria:** real per-category numbers are visible in CI on every push.

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
- [ ] **Hint/warning envelope.** Emit DWScript's `Errors >>>>` / `Hint:` output — at minimum
      case-mismatch (`"P1" does not match case of declaration ("p1")`) and unused-var hints —
      by default when a fixture expects it (52 failures; 62/437 SimpleScripts expect it).
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
