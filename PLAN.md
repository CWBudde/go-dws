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

- [x] Ground-truth CLI report (`cmd/fixture-report`, a pure-Go tool) wired into `just`
      (`fixture-report`) and CI (`.github/workflows/test-fixtures.yaml`, called from `test.yml`).
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
      CLI-only `cmd/fixture-report` cannot, e.g. FailureScripts 102/528).

### P1 — Front-end coverage 🔴 *(attacks the 65% — ~440 fixtures)*

- [x] **Generics via monomorphization.** `type TList<T> = class … end;`, `type TRec<A,B> =
      record … end;`, and generic array aliases now parse and run. The parser accepts generic
      type-parameter lists on declarations and generic type-argument lists in type annotations,
      `new` expressions, and expression position (`TTest<Integer>.Method`, disambiguated from
      comparisons by requiring a trailing `.`). A new pass (`internal/generics`, wired into
      `frontend.compileParsedResult` and `cmd/dwscript run` before semantic analysis) specializes
      each template into a concrete declaration by deep-cloning the AST and substituting the type
      arguments, so the analyzer and evaluator only ever see ordinary concrete types. Specialized
      classes take the mangled name `TTest<Integer,String>` (matching DWScript's `ClassName`).
      **`GenericsPass` 0 → 12/23 (52%)**, overall fixtures 547 → 559. Remaining gaps are separate
      feature dependencies, not generics parsing: generic interfaces (`interface1`), generic
      `external` classes / function-pointer types (`class_external1`, `external_promise`,
      `func_ptr1`), external generic method bodies (`function TTest<T>.Foo` — `while`, `repeat`),
      operator-overload specialization, and dynamic-array method gaps surfaced through `array of T`
      (`array1`, `tlist1`). Deeply-nested closing `>>` is not yet supported.
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
      the parser making untyped consts look explicitly (empty-)typed. Helper consts/class-vars
      via metaclass (`String.Hello`, `TMyArray.ByeBye`) now resolve too (fixed under P4 —
      `HelpersPass/string_consts.pas`, `static_array_helper.pas`).
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
- [~] Semantic hardening (partially done 2026-07-03).
      - [x] **Multi-pass analyzer for forward references.** `Analyzer.Analyze`
        (`internal/semantic/analyzer.go`) is now two-pass: pass 1 registers every top-level
        regular function's signature (via the new `registerFunctionSignature`, split out of
        `analyzeFunctionDecl` in `analyze_functions.go`) while analyzing all other declarations
        in source order; pass 2 analyzes the deferred bodies (`analyzeFunctionBody`). Top-level
        functions are therefore mutually visible, so mutual recursion and forward calls work
        without an explicit `forward` declaration. Declaration source-order semantics are
        preserved (signatures register inline), so nothing that depended on ordering regressed.
        Genuinely-undefined names still error. Covered by
        `TestForwardFunctionReferenceWithoutForwardKeyword`.
      - [x] **Unit function bodies** (`internal/semantic/unit_analyzer.go`): the
        dependency-aware `AnalyzeUnitWithDependencies` path now analyzes implementation bodies
        via `analyzeFunctionBody` instead of only checking for their presence. (The production
        `analyzeUnitDeclaration` path already analyzed bodies.)
      - [ ] **Subrange bounds at compile time — deferred.** No fixture in the corpus declares a
        subrange type, and existing `subrange_test.go` asserts out-of-range assignments do *not*
        error at compile time (runtime-deferred). Zero fixture yield; not pursued.
      - [ ] Make the analyzer fully **type**-order-independent (class parents/fields declared
        later without `forward`) — needs a real two-phase class builder; separate follow-up.
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
- [x] **Dynamic-array method coverage.** The dynamic-array helper only implemented
      `Add/Push/Pop/Swap/Delete/IndexOf/SetLength/Map/Join`; every other DWScript array method
      was rejected by the analyzer (`Reverse`, `Contains`, `Filter`, `Clear`, `Peek`) or accepted
      by semantics but unimplemented at runtime ("no helper found" for `Move`, `Sort`, `Insert`,
      `Copy`, `Remove`, `ForEach`). Added the runtime implementations (`internal/interp/evaluator/
      array_helpers.go`) and registered them (`helpers_validation.go`), and closed the semantic
      gaps (`internal/semantic/analyze_array_helpers.go`). Details that matter for parity:
      `Reverse`/`Swap`/`Sort` return the receiver so they chain (`a.Reverse.Join`); `Add`/`Push`
      flatten an array argument of the element type (`a.Add(a)`); `Remove(value[, startIndex])`
      removes the first match at/after `startIndex` and returns its index (−1 if none);
      `Insert`/`Move` raise catchable `Upper/Lower bound exceeded! Index N` exceptions pointing at
      the method name; Boolean arrays are treated as naturally sortable.
- [x] **Built-in functions as function pointers.** Passing a builtin (`IntToStr`, `FloatToStr`,
      `BoolToStr`, …) to a higher-order method raised "Function pointer is nil": the
      `FunctionPointerValue.BuiltinName` field was populated but `IsNil()` ignored it and
      `executeFunctionPointerDirect` had no builtin branch. `IsNil()` now accounts for
      `BuiltinName` and the executor dispatches through `builtins.DefaultRegistry`. Unblocks
      `a.Map(IntToStr)`-style code. ArrayPass 27→59 (CLI) / baseline 30→64, LambdaPass 0→1,
      overall 434→469 (CLI); no category regressed.
- [x] **Set operations closed to parity for common cases.** Four independent gaps kept SetOfPass
      at 16%: (1) `var s : TMySet;` where `TMySet = set of …` left the variable nil (only inline
      `set of …` annotations were zero-initialized), so `in`/`Include`/`Exclude` all raised "Object
      not instantiated" — `createZeroValue` now resolves named set types via `__set_type_<name>`
      and returns an empty `SetValue`. (2) The procedure forms `Include(s, x)` / `Exclude(s, x)`
      were listed in `isBuiltinFunction` but had no analyzer or evaluator handler (only the method
      forms `s.Include(x)` existed), so every call was "Unknown name Include" — added
      `analyzeIncludeExclude` (semantic) and `builtinIncludeExclude` (evaluator, mutates the set
      lvalue in place). (3) Set comparison operators `=`/`<>`/`<=`(subset)/`>=`(superset) were
      rejected as "Invalid Operands"/"requires comparable types" — added a set branch to the
      comparison analyzer (returns Boolean; `<`/`>` remain invalid per DWScript) and
      `evalSetComparison` in the evaluator. (4) Set↔Integer bitmask casts `Integer(s)` / `TSet(i)`
      were unhandled — added `castToInteger` SET case, `castToSet`, and the `isValidCast` set/integer
      branch. **SetOfPass 4 → 14 (16% → 56%)**, overall 470 → 480 (CLI) / scored baseline 597 → 607;
      no category regressed. Remaining fails are separate features (anonymous inline enums in
      `set of (A,B,C)`, array↔set conversion, `set of` record fields, out-of-range diagnostics).
- [ ] Work the 88 wrong-output fixtures (e.g. `casts_base_types` rounding, `case_variant`).
- [~] Work the runtime-panic fixtures (metaclass `ClassName`, class-method dispatch, `class of`).
      **Metaclass member access closed for the common cases (2026-07-03).** (1) Member access on
      a `class of X` metaclass value (`runtime.TypeMetaValue` wrapping a `*types.ClassOfType`) now
      resolves class members by delegating to the shared `resolveClassMetaMember` helper
      (`internal/interp/evaluator/visitor_expressions_members.go`) — fixes `TClass.ClassName` etc.
      (2) `ClassParent` is now handled both semantically (`analyze_classes.go`) and at runtime
      (walks `IClassInfo.GetParent`, returns `NilValue` at the root). (3) Class methods reached
      through a metaclass now resolve **inherited** class methods via
      `isClassMethodInHierarchy`/overload `IsClassMethod` (`analyze_method_calls.go`,
      `analyze_classes.go`) instead of the own-class `ClassMethodFlags` map only. (4) `as` and
      func-style casts accept a `class of` target (`analyze_expressions.go`,
      `visitor_expressions_types.go`, `type_casts.go`). Fixes `class_method4`, `class_parent`,
      `class_of_cast`; SimpleScripts 241 → 244, overall 480 → 483 (CLI); no category regressed.
      (5) **Helper class consts / class vars via a type metaclass** (`String.Hello`,
      `TMyArray.ByeBye`) now resolve: `findHelperClassMember`
      (`internal/interp/evaluator/helper_methods.go`) is consulted in the `TYPE_META` member
      path. Fixes `HelpersPass/string_consts`, `static_array_helper`; HelpersPass 13 → 15,
      overall 483 → 485 (CLI); no category regressed.
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

1. `cmd/fixture-report` reports **≥ 90%** on all non-host-library categories.
2. All `*Fail` suites reproduce DWScript diagnostics (error-detection parity).
3. There is exactly **one** type representation and **one** evaluator in the tree.
4. CI fails on any per-category regression.
5. No public API or CLI flag exposes a non-functional mode.

Track progress against **fixture pass rate**, not phase checkboxes.

---

## 5. How to reproduce the status numbers

```bash
go build -o bin/dwscript ./cmd/dwscript
go run ./cmd/fixture-report                          # full per-category table + total
go run ./cmd/fixture-report --category SimpleScripts --list-fails
# or, via just:  just fixture-report
```

Every status figure in this document came from that script on the current branch head.
