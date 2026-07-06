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
| Arrays | 🟡 | 80% (harness, 2026-07-04) |
| Sets, Helpers | 🟡 | 80–81% (harness, 2026-07-04) |
| Overloads | 🟡 | 85% (harness, 2026-07-04) |
| Generics, Lambdas, Associative arrays, Property expressions | 🟡 | 53–81% (see P1) |
| JSON connector | 🟡 | 39/82 = 48% (JSONConnectorPass, harness, 2026-07-06) |
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
- [~] **Property read/write expressions** (`PropertyExpressionsPass` **0 → 10/19 CLI (53%),
      meeting its P1 exit criterion**; overall CLI 599 → 616; no category regressed). DWScript lets
      a property accessor be a parenthesized expression rather than a
      bare field/method name: `property P: Integer read (2*Field) write (Field := Value div 2)`.
      The write clause takes three shapes, all now parsed and executed: an assignment
      (`Field := Value div 2`), a bare lvalue normalized to `lvalue := Value` (`write (FSub.Field)`),
      and a call executed as a statement (`write (SetField(Value div 2))`); the implicit `Value`
      is the assigned value. Implemented across the stack:
      - **AST/parser:** `PropertyDecl.WriteStmt` and `RecordPropertyDecl.{ReadExpr,WriteStmt}`;
        `parsePropertyWriteClause`/`buildPropertyWriteSpec` (shared by class and record parsers);
        record property parser now accepts `read (…)`/`write (…)`; `class property` is now parseable
        inside `class helper`s.
      - **Semantic:** `validateWriteExprSpec`, class-property expressions analyzed in a scope that
        binds class vars (own + inherited), instance fields, index params, and `Value`; `Self`
        allowed inside property expressions (metaclass for class properties). `PropertyInfo.WriteExpr`
        and `RecordPropertyInfo.{ReadExpr,WriteExpr,ReadKind,WriteKind}` added.
      - **Evaluator:** `executeExpressionBackedPropertyWrite`, `executeRecordExpression{Read,Write}`,
        `evalClassPropertyExpression{Read,Write}` (metaclass Self + class-var sync); bare
        implicit-Self record property reads; circular-reference guard re-keyed by `PropertyInfo`
        identity (was property *name*, which false-positived on same-named properties across a
        hierarchy — `chained_as_properties`).
      - **Passing:** simple_instance/object_write/object_writer, record_write_statement,
        chained_as_properties, double_brackets, asclass_property.
      - **Interface properties (added 2026-07-05).** `InterfaceType` now carries a
        `Properties map[string]*PropertyInfo` (populated by `analyzeInterfacePropertyDecl` from
        `InterfaceDecl.Properties`); `analyzeMemberAccessExpression` resolves `intf.Prop` to the
        declared property type for both read and write, and the accessor kind (method vs inline
        expression) is recorded so the pre-existing runtime interface-property machinery
        (`MutableInterfaceInfo.Properties`, `InterfaceInstance.LookupProperty`) executes them.
        Interface accessor expressions are validated against the concrete implementing object at
        runtime, so the analyzer only records the declared type. Fixes `simple_interface_expressions`,
        `interface_write_expressions`.
      - **Nested-record default-field init (fixed 2026-07-05, was a P4 pre-existing bug).**
        `getZeroValueForType`'s RECORD case now applies each field's default initializer
        expression (looked up via the registered `RecordTypeValue.FieldDecls`) instead of pure
        zero-init, so a nested record field (`FSub : TBase` where `TBase.Field = 1`) matches a
        top-level `var r : TBase`. Fixes `simple_record_expressions`.
      - **Remaining (separate features):** helper-property resolution through a metaclass
        (helpers_property_expressions et al.), record/class `class property` parsing
        (class_property_expressions, class_property_write_expressions, property_auto_field —
        needs `class property` in record bodies plus record auto-property backing fields),
        multi-index expression-backed indexed properties (indexed_expressions), and the hint
        envelope (read_write_other_property, ⚠️ blocked above).
- [~] **Lambda / anonymous-method coverage** (`LambdaPass` **1 → 4/6 (17% → 67%)**, meeting its
      P1 exit criterion; overall CLI 606 → 613, collateral ArrayPass 86 → 89, FunctionsString
      51 → 52). Three independent gaps closed:
      - **Parser:** a lambda whose statement-list body is written as a single `begin…end` block
        (`lambda (a) begin PrintLn(a+'!') end end`) left the lambda's own trailing `end` as a
        stray token ("Expression expected"). `parseLambdaExpression` now consumes that trailing
        `end` after the block (`internal/parser/expressions_lambda.go`). Fixes `simple_proc`.
      - **Fewer parameters than the target type.** DWScript lets a lambda / anonymous method
        declare *fewer* parameters than its target function type; the surplus call arguments are
        ignored (`procedure(Sender) := lambda PrintLn('hi') end`). The semantic count check
        (`analyzeLambdaExpressionWithContext`) now rejects only *too many* params, returns the
        target type when the lambda is narrower (so the assignment `canAssign` succeeds), and the
        runtime lambda dispatch (`executeLambdaDirect`) ignores surplus args. Fixes
        `implicit_params`.
      - **Parameter-type inference for `array.Map`.** `a.Map(lambda (e) => …)` reported "lambda
        parameter type inference not fully implemented" because the `map` helper analyzed its
        callback argument *without* the expected function-pointer context (unlike `filter`/
        `foreach`). It now passes `function(ElementType): Variant` as context so an untyped
        parameter infers from the element type, and the mapped array's element type is the
        callback's actual return type — whether a lambda or a named function
        (`internal/semantic/analyze_array_helpers.go`); the runtime `evalArrayMap` likewise
        rebuilds the result array's element type from the mapped values so runtime metadata
        matches the semantic type. Fixes `map`.
      - A lambda that declares fewer parameters than a **procedure** target still keeps its
        return-type check (a value-returning lambda is rejected for a procedure type), and the
        adapted lambda reports the target's arity while preserving its own return type.
      - **Remaining:** `immediate` needs value-context auto-invoke of a parameterless function
        pointer (`PrintLn(f)` where `f` holds a lambda — a broader semantic change), and
        `simple_func` needs the hint envelope (⚠️ blocked above).
- [~] **Associative arrays** (`array [KeyType] of ElementType`) — `AssociativePass` **3 → 22/27
      (11% → 81%)**, meeting its ~80% target; overall CLI 619 → 639. A sparse map keyed by an
      arbitrary type (Integer, String, Float, record-by-value, object-by-identity, static-array-by-
      value). The parser already emitted `array [Ident]` with an `IndexType`; the whole feature is a
      new value kind added below the parser. Implemented across the stack:
      - **Type system:** new `types.AssociativeArrayType{KeyType, ElementType}` (distinct
        `TypeKind "ASSOCIATIVE_ARRAY"` so no dense-array code silently misroutes it). The
        discriminator is `types.OrdinalBounds`: an enum-indexed `ArrayTypeNode` whose index type has
        no ordinal bounds (Integer/String/record/object/…) resolves to an associative type instead of
        erroring — in both `semantic/type_resolution.go:resolveArrayTypeNode` and the evaluator's
        `type_resolution_helpers.go:resolveAssociativeArrayTypeNode` (wired into `createZeroValue`,
        `ResolveType`).
      - **Semantic:** index reads check the key against `KeyType` (`analyze_arrays.go`); `key in a`
        is key membership (`analyze_expr_operators.go`); a new `analyze_associative_arrays.go` handles
        `Keys` (→ `array of KeyType`), `Length`/`Count`, `Clear`, `Delete(key)→Boolean`; nil is
        assignable (`canAssignNil`).
      - **Runtime:** `runtime.AssociativeArrayValue` — ordered key/value slices with a dedicated
        `associativeKeyEqual` (object **pointer identity**, nil-key support, structural for
        records/static arrays via `Equal`, primitives via `Equals`); reference-typed `Copy()`;
        value-typed keys snapshotted on insert. Evaluator: index read returns the live stored value
        (missing key → element zero value, **no insert** — proven by `delete.pas` sieve `Length`=35),
        index write inserts/updates with no bounds check, `a := nil` rebinds to a fresh empty map.
      - **Remaining (5, documented likely-misses):** nested lvalue vivification through a key
        (`a[k].field := v`, `a[k][j] := v`, `a[k].Add(…)` — `elements_of_value`, `array_of_dyn`);
        DWScript hash iteration order (`records` expects `b,a`); ARC destructor timing on slot
        replace/clear (`delete_sequence`); Variant→key coercion (`variant_key_cast`).
- [x] **Source inclusion directives** (`{$INCLUDE 'file'}`, `{$I 'file'}`,
      `{$INCLUDE_ONCE 'file'}`). Handled entirely in the lexer as a pure textual substitution
      (matching DWScript): an include pushes the current scan state onto an **include stack**
      (`internal/lexer/include.go`) and makes the referenced file the active input; when the
      included input is exhausted, the parent file resumes exactly where it left off, so nesting
      and recursion work naturally. `{$INCLUDE_ONCE}` de-duplicates by canonical path (marking
      *before* splicing so mutual `include_once` recursion is guarded), while plain `{$INCLUDE}`
      always re-includes. File resolution is injected via a `WithIncludeResolver` lexer option;
      the lexer tracks the **current including file's canonical path** so relative includes
      resolve against the directory of the file they appear in (nested `sub/a.inc` → `{$INCLUDE
      'b.inc'}` finds `sub/b.inc`), matching DWScript. `NewFileIncludeResolver` decodes with the
      same BOM/UTF-16 handling as top-level sources; the resolver is wired through
      `frontend.ParseWithFilename`/`Compile` and `cmd/dwscript run`. An **unresolvable include is
      a fatal front-end diagnostic** (tracked on a dedicated include-error channel, surfaced via
      `Parser.LexerIncludeErrors`, so a missing include fails to compile instead of running with
      its content silently dropped) — other lexer errors stay advisory to avoid regressing the
      `*Fail` diagnostic suites. `{$I %FILE%}`-style value substitutions are left untouched (a
      separate, unimplemented feature). Fixes SimpleScripts `include`, `include_once`,
      `include_once2`; `includeSym` now produces correct output but still needs the hint envelope
      (⚠️ blocked above). SimpleScripts 290 → 294 (harness). Covered by
      `internal/lexer/include_test.go`.
- [~] **JSON connector** (`JSONConnectorPass` **0 → 41/82 harness = 50%; 46/82 CLI = 56%**; `JSONConnectorFail`
      0 → 2). The script-visible `JSON` static namespace and a distinct **`JSONVariant`** connector
      type. Reuses the pre-existing `internal/jsonvalue` tree + `runtime.JSONValue`; adds byte-exact
      DWScript `Stringify`/pretty writer and an order-preserving parser (`internal/jsonvalue/
      {stringify,parse}.go`). Implemented across the stack:
      - **Type system:** `types.JSONVariantType` (distinct from `Variant` so plain-Variant member
        access still errors — required by `coalesce_typ`), with auto-box/implicit-cast assignability.
      - **Semantic:** `JSON.<method>` namespace resolution and permissive JSONVariant member/index/
        method typing (`internal/semantic/analyze_json.go`); boolean-context, comparison, cast, and
        `VarIsEmpty` acceptance.
      - **Runtime:** `JSON.Parse/Stringify/PrettyStringify/Serialize/NewObject/NewArray/*UTF8/
        Parse{Integer,Float,String}Array`; JSON value methods (TypeName/Length/Low/High/ElementName/
        Clone/Extend/AddFrom/Add/Push/Delete/Swap/Defined/ToString); member/index read+write with
        positional object indexing; truthiness (`jsonvalue.IsFalsey`), implicit/explicit casts,
        comparison, `??` coalesce, auto-boxing of scalars into a JSONVariant; print/string-cast
        semantics (string→raw, container→compact JSON, bool→Pascal `True`/`False`, null→`null`).
      - **Record/class serialization (added 2026-07-06).** `JSON.Stringify/Serialize/PrettyStringify`
        now serialize class instances and records. `ValueToJSONValue` became the evaluator method
        `valueToJSONValue` (`internal/interp/evaluator/json_serialize.go`) so property getters and a
        custom `Stringify` override can run; `objectToJSON` walks the class hierarchy most-derived
        first, emits public (non-private/protected) fields + non-indexed readable properties sorted
        ordinally per level, runs expression/method getters via the existing `executePropertyRead`,
        and splices a parameterless `function Stringify:String` as raw JSON. Records mirror this via
        `recordToJSON`. Field visibility is now threaded to runtime metadata
        (`FieldMetadataFromAST` maps `field.Visibility` instead of hardcoding public) and to record
        types (new `types.RecordType.FieldVisibility`, populated in semantic + evaluator record
        builders). A per-level property enumerator `ClassInfo.GetOwnProperties` was added. A
        collateral evaluator fix — **chained property-object lvalue writes** (`obj.Prop.field := v`):
        `evaluateLValueMember` (`var_params.go`) now resolves a property intermediate through the
        getter instead of failing "field not found" — unlocks the write-heavy serialization fixtures
        and helps property/class code generally. **CLI ground-truth `JSONConnectorPass` 41 → 46**
        (stringify_class/special/pretty, serialize_class, stringify_func_result), overall CLI
        680 → 686; no category regressed. **Harness baseline ratcheted +1 (40 → 41)** — the other 4
        newly-correct fixtures emit correct output under `bin/dwscript` but are masked in the harness
        by the over-aggressive unused-private-field hint envelope (the ⚠️-blocked hint issue below),
        not by any output error.
      - **Remaining (~36 fails):** anonymous-record-literal *parser* support (`stringify_anonymous*`,
        `stringify_record2`); auto-property backing-field generation (`stringify_method`,
        `stringify_array_of_classes` — `property alpha : Integer;` currently rejects the synthesized
        `Falpha`); `array of T` property-type parsing (`stringify_class_getter`); record copy-on-
        assign value semantics (`stringify_record`); set-of-enum serialization (`stringify_set`);
        gating the unused-private-field hint so the 4 masked serialization fixtures score in the
        harness; associative-array-keyed fixtures (blocked); nested-lvalue vivification through a JSON
        index (`generate1`); reparent/ownership (`array_add_dupe`, `reparent`); float formatting
        (`int64_json`, `numbers`). Covered by `internal/jsonvalue/stringify_test.go`.
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

- [x] **Deduplicate helper registration (root fix for a defanged non-determinism smell).** Done:
      a user helper is now backed by exactly **one** `*runtime.MutableHelperInfo`.
      `TransferHelpersFromSemanticAnalysis` converts each semantic `*types.HelperType` once (the
      semantic map lists the same helper under resolved *and* declared target keys, so the transfer
      alone used to mint N instances) and the evaluator's `VisitHelperDecl` reuses the transferred
      instance instead of building its own (helper property keys normalized so the evaluator's
      spec-complete `PropertyInfo` overwrites the transfer's spec-less one). The P4
      bind-into-every-copy workaround (`lookupAllMutableHelpers`) is deleted; `VisitFunctionDecl`
      and the parent-helper linkage use the now single-valued `lookupMutableHelper`. Verified by
      `TestUserHelperRegisteredAsSingleInstance` plus HelpersPass byte-stable across 6× re-runs.
- [x] **Shadow interpreter:** re-point the tests that call `interp.evalClassDeclaration` /
      `evalIntegerBinaryOp` / `evalEnumDeclaration` / set & operator helpers at the evaluator,
      then delete those bodies (`expressions_binary.go`, `enum.go`, `type_alias.go`, `set.go`
      operators, `declarations.go` class/interface builders). Mechanical, not a rewrite.
      *(Done: deleted `expressions_binary.go`, `set.go`, `enum.go`, `type_alias.go` outright,
      the `declarations.go` class/interface/operator builders, and `evalHelperDeclaration`;
      `EnumTypeValue`/`TypeAliasValue` are now aliases of the `runtime` structs. Interface
      tests rewritten to declare via scripts through the production evaluator; set tests
      re-pointed at the evaluator's binary-op/`in`/`Include`/`Exclude` paths. −27 unreachable
      funcs (176→149 per `deadcode ./cmd/...` filtered to `internal/interp`), −2,241 net LOC;
      fixture report byte-identical to main.)*
- [x] Move `Evaluator.currentNode` into `ExecutionContext`; remove the double `MethodRegistry`
      allocation (`internal/interp/interpreter.go:61`).
      *(Done: `currentNode` now lives on `runtime.ExecutionContext` (copied on `Clone`, cleared
      on `Reset`); `Evaluator.CurrentNode`/`SetCurrentNode` delegate to the tracked context,
      with a nearest-non-nil-context fallback so nil-ctx sub-evaluations keep the old
      flat-field error-position semantics, and `Eval` saves/restores the node on the context
      it runs. `NewWithDeps` reuses the registry the evaluator allocated on `EngineState`
      instead of allocating a second one, and the redundant `Interpreter.methodRegistry`
      field is gone. Behavior-neutral: full suite, fixture gate, and `-race` all green.)*
- [ ] Split the evaluator god files (`visitor_statements.go` 1461, `visitor_declarations.go`
      1426, `var_params.go` 1112).
- [ ] **Bytecode decision:** hide `--bytecode` and `pkg/dwscript.CompileModeBytecode`, delete
      the rigged benchmarks and the "5-6x faster" claims from README/docs. Then **either**
      delete `internal/bytecode` outright **or** (if a VM is a real goal) rebuild it on the
      shared `internal/builtins` registry and `internal/interp/runtime` values — do not keep
      extending the current fork.
      **Owner decision 2026-07-04: deferred, leaning keep.** Do not delete `internal/bytecode`;
      leave it as-is (opt-in, unmaintained) until the owner revisits. The delete-vs-rebuild
      choice stays open; the honesty sub-items (hiding the flag, removing the speedup claims)
      may still be done independently if picked up.
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
      **First batch of root causes closed (2026-07-04): 28 fixtures, overall 485 → 513 (CLI).**
      (1) *Measurement*: the CLI `run` command and `cmd/fixture-report` read files raw while the
      internal harness BOM-decodes them — extracted `internal/encoding` (UTF-8 BOM / UTF-16 LE/BE /
      Latin-1 fallback) and wired it into both; closes `char_in`, `unicode_const`, `string_aggregate`,
      `FunctionsString/case`, `strsplit(2)`, `strjoin`, `strisascii`, `bytesizetostring`,
      `aes_encryption`, `sparse_matmult`. (2) `Integer(<float>)` casts round half-to-even instead of
      truncating (`casts_base_types`). (3) `IsInRange` unwraps Variants so `case v of 11..12` matches
      (`case_variant`). (4) Enum `.Name`/`.QualifiedName` print `?` for unnamed ordinals
      (`enumerations_names`, `enumerations_qualifiednames`). (5) Builtins aligned with DWScript:
      `LogN(Base,X)` argument order, niladic `MaxInt` = High(Int32), `LeastFactor(n<=0)=0`,
      `FindDelimiter` not-found = -1, `RevPos('')=0`, `VarToFloatDef` Null→0 + comma decimals
      (`FunctionsMath/basic`, `maxint`, `least_factor`, `delimiters`, `pos_posex`, `vartofloat`).
      (6) Object references compare by identity, not rendered string (`oop_compare`, `array_in`).
      (7) `str in ['a'..'z', ...]` compares whole strings lexicographically per range
      (`string_in_op3`). (8) try/finally suspends+re-arms pending Exit so finally blocks run fully
      and Exit still propagates (`exit_finally`, `exit_finally2`). (9) `IndexOf` clamps negative
      fromIndex (`indexof_from_static`). (10) Constructor/method `var` params bind by reference when
      the declaration is unambiguous (`oop_field`). Baselines ratcheted: SimpleScripts 259 → 268,
      FunctionsMath 21 → 24, FunctionsString 42 → 45, ArrayPass 65 → 67; no category regressed.
      Remaining wrong-output fails are mostly hint/warning-envelope emission (case-mismatch hints,
      `Empty THEN block`, deprecation warnings), field shadowing per-class storage (`field_scope`),
      overload-set resolution across scopes (`OverloadsPass/*`), heredoc/triple-quote lexing, and
      UTF-16 surrogate iteration (`for_in_str(2)`).
      **Second batch closed (2026-07-04): overall CLI 523 → 556 (27% → 29%), harness scored
      passes 660 → 684.**
      (11) **Overload-set resolution across scopes** — OverloadsPass **5 → 33/39 (85%, harness;
      26/39 CLI — 7 of the CLI fails are hint-envelope-only)**, meeting its P4 exit criterion.
      Root causes: defaulted params optional in matching; nil-arg ranking (class > array/function
      pointer); metaclass hierarchy distance; Variant boxing ranked below element-wise
      conversions; non-variadic beats variadic; constructor name merged with same-named class
      methods (incl. `TClass.Create(args)` sugar); implementations inherit declaration defaults;
      subclass non-`overload` declarations hide the parent set; overload-aware `inherited`
      resolved against the executing method's *defining* class; nested functions get an
      env-scoped `LocalFunctionSet` instead of leaking into the global registry; builtins compete
      inside user overload sets; record instance/class(static) method sets merge with
      per-overload visibility; array-literal argument typing (empty `[]`, set-vs-array
      disambiguation); unit interface declarations act as implicit forwards. Side gains:
      ArrayPass 67→69, SimpleScripts 272→277 (harness). Remaining 6 are separate features
      (class operator `=`/`<>` overloading, `@obj.Method` pointers, function-pointer parameter
      typing in overloads, metaclass `inherited`).
      (12) **Per-class field storage for shadowed fields** (`field_scope`): `ObjectInstance`
      keeps one slot per declaring class when a field is redeclared down the hierarchy; reads
      resolve by static type (member access), executing method's defining class (bare
      identifiers), or cast target. Builtin exception subclasses no longer artificially
      redeclare `Message`.
      (13) **Heredoc strings**: triple-apostrophe (and triple-quote) multi-line strings lex per
      DWScript rules (opener followed by newline, closing-line indent stripped); fixed a latent
      parser double-unescape of quoted quotes. Fixes `triple_apos1/2`.
      (14) Misc: bare `Name = Value;` in a class body is a class const; field initializers can
      reference class consts; calls skip execution when an argument raised; empty-array
      `Peek`/`Pop` raise `Upper bound exceeded!` with routine context; unhandled raises print
      `User defined exception:` with DWScript-style position and caller-labeled stack trace.
      SimpleScripts 277 → 287 (harness).
      **UTF-16 surrogate iteration (`for_in_str(2)`) is ⏸️ won't-fix** per the adopted ADR
      `docs/string-encoding.md` (UTF-8-native strings are an intentional divergence); it would
      require WTF-8 threading through Ord/Chr/indexing/concat.
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
- [x] Post-exception continuation semantics (`assigned.pas` expects execution to continue after a
      caught runtime error) and BOM-preserving output. **(A) Runtime errors are catchable
      exceptions.** Nil-receiver access (member read/write, method call), explicitly-freed-object
      access, and `raise nil` now raise `Object not instantiated` / `Object already destroyed` as
      script exceptions (`try/except` catches them; execution continues) instead of aborting the
      program. Non-virtual methods still dispatch on a nil receiver (the error only surfaces when
      the body dereferences `Self`), matching DWScript; the routine name is spliced into the
      message (`… in TMyObj.Proc [line: …]`) and the position is the member/method identifier.
      Runtime errors escaping a `try` block or a routine body are converted to catchable
      exceptions at those boundaries (`visitor_statements.go`, `user_function_helpers.go`);
      method call stack frames are now class-qualified. Also: `new Integer[0]` is legal (empty
      array), and collection builtins materialize never-written static-array slots to the element
      zero value. `RaiseException` is implemented on the evaluator context so builtins
      (`StrToInt`, etc.) raise catchable exceptions. Fixes SimpleScripts `assigned`, `self`,
      `destroy`, `raise_nil`, plus 6 FunctionsString exception fixtures. **(B) BOM output** needs
      no new code: the stacked #342 input decoding already strips BOMs when reading both source
      and expected `.txt`, so BOM-carrying fixtures (`string_aggregate`, `char_in`,
      `unicode_const`) match; no fixture requires a BOM to survive to output. Harness baselines:
      FunctionsString 45 → 51, SimpleScripts 268 → 272; no category regressed; the 11 EncodingLib
      fixtures from #342 unaffected.
- **Exit criteria: MET (2026-07-04, internal harness).** ArrayPass **92/115 (80%)**, SetOfPass
      **20/25 (80%)**, HelpersPass **22/27 (81%)**, OverloadsPass **33/39 (85%)**. Overall harness
      scored passes 684 → 729 (collateral: AssociativePass 0→3, GenericsPass 12→13). Root-cause
      batch: array concat/append/`+=`/`Add` with literals and Variant flattening; array
      constructors (empty `[]` → array of Variant, heterogeneous widening, ordinal range
      expansion, `nil` clears dynamic arrays); live byref `(array, index)` element refs with
      DWScript bounds diagnostics; Variant casts at builtin args/indexes; inline anonymous enums
      in `set of (…)`; sets as value types; helper dispatch (nil receivers, strict static-type
      dispatch, inheritance precedence, class vars as shared refs, record helper consts).
      Remaining fails cluster in separate features: lambda parameter-type inference, metaclass/
      class-alias element inference, function pointers in arrays, `set of` record fields,
      lexer-time `Declared()`/`{$FATAL}`.

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
