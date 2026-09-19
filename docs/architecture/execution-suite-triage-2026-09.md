# Remaining execution-suite triage — 2026-09-19

This audit closes PLAN.md E10a/E10b: it identifies the first blocking cause of each remaining
failure in the selected categories. It does not implement these follow-ups or claim new fixture
passes. The counts below are fresh CLI measurements before E10c/E10d changes, with the shared
category hint policy; FunctionsMath was also checked through the Go harness.

## Measurement

| Category | Passed | Scored | Failed | Unscored |
| --- | ---: | ---: | ---: | ---: |
| FunctionsMath | 35 | 40 | 5 | 0 |
| HelpersPass | 22 | 27 | 5 | 0 |
| OperatorOverloadPass | 5 | 8 | 3 | 0 |
| LambdaPass | 4 | 6 | 2 | 0 |
| OverloadsPass | 37 | 39 | 2 | 0 |
| FunctionsGlobalVars | 16 | 16 | 0 | 0 |
| BuildScripts | 0 | 1 | 1 | 53 |
| FunctionsString | 57 | 58 | 1 | 0 |
| PropertyExpressionsPass | 18 | 19 | 1 | 0 |
| **Total** | **194** | **214** | **20** | **53** |

The previous queue's six FunctionsMath failures and one FunctionsGlobalVars failure are stale.
`queue_snapshot` passes with the strict hint setting selected by
[`fixtureconfig`](../../internal/fixtureconfig/policy.go). Upstream
[`UdwsFunctionsTests.pas`](../../testdata/fixtures/UdwsFunctionsTests.pas) leaves the compiler's
hint default unchanged. The earlier four unwanted `join`/`Join` hints came from a pedantic run;
they are not a current declaration-resolution gate. This finding does not change `Map` inference
or suppress any hints. BuildScripts' sole scored failure is a runner/input mismatch, detailed below.

Commands used (run from the repository root):

```sh
GOCACHE=/tmp/go-dws-e8-cache GOFLAGS=-buildvcs=false \
  go run ./cmd/fixture-report --category FunctionsMath --classify --list-fails \
  --cli /tmp/go-dws-e10-math-cli
GOCACHE=/tmp/go-dws-e8-cache GOFLAGS=-buildvcs=false \
  go test ./internal/interp -run '^TestDWScriptFixtures/FunctionsMath$' -count=1 -v

GOCACHE=/tmp/go-dws-e8-cache GOFLAGS=-buildvcs=false \
  go build -o /tmp/go-dws-e10-small-cli ./cmd/dwscript
GOCACHE=/tmp/go-dws-e8-cache GOFLAGS=-buildvcs=false \
  go build -o /tmp/go-dws-e10-small-report ./cmd/fixture-report
for category in HelpersPass OperatorOverloadPass LambdaPass OverloadsPass \
  FunctionsGlobalVars BuildScripts FunctionsString PropertyExpressionsPass; do
  /tmp/go-dws-e10-small-report --cli /tmp/go-dws-e10-small-cli --build=false \
    --category "$category" --classify --list-fails --timeout 15
done
```

The report rebuilds the math CLI; the separately built small-category CLI passes the report's
freshness check. Small-category reports ran independently with the same options shown in the
loop. `classname_helper1` hit the 15-second timeout, consistent with the prior recursion audit.
The targeted Go fixture command passes because current baseline floors are satisfied; it does
not imply all FunctionsMath fixtures pass. No expectations or category floors were changed.

## FunctionsMath: five failures, three groups

| Group | Fixture and first blocking evidence | Required follow-up |
| --- | --- | --- |
| Variant numeric and ordinal arguments | `abs`: `Abs` rejects Variant at 13:9. `inc_dec_variant_op`: Variant deltas rejected by `Inc`/`Dec` at 7:9 and 9:9; two-argument `Succ`/`Pred` rejected at 12:18 and 13:18. | Align semantic signatures and runtime Variant unwrapping; support the optional ordinal delta consistently. The evaluator has independent restrictions, so semantic acceptance alone is insufficient. |
| Builtin signatures | `haversine`: five arguments rejected; runtime accepts four and hardcodes radius 6371. `random`: `RandG(100, 5)` rejected at 30:14 because only zero arguments are accepted. | Add the Haversine radius and RandG mean/deviation forms in semantic and evaluator paths; preserve supported defaults and validate arity/types. Existing `TestBuiltinSignatures_CorrectedShapes` pins RandG to zero arguments and must be revised with a real-path regression. |
| Seeded RNG and diagnostics | `randseed`: missing deprecated warning at 2:9, plus independent random-sequence differences. Seed readback is correctly 1234; expected draws are 47687, 125347, 903103, actual draws are 231682, 458315, 417389. | Track deprecation under F1 separately from seeded RNG compatibility. Establish the upstream generator/seed contract before changing the algorithm; the local reference source checkout is empty. |

These are distinct from E4's completed numeric/array helper work. None requires a host library
or a change to E3 case-hint policy. Direct math reproductions use `--hints strict`.

## Smaller categories: actionable groups

| Group | Representative fixture and first blocker | Ownership / next step |
| --- | --- | --- |
| Helper dispatch precedence | `HelpersPass/classname_helper1`: helper `ClassName` calls `Self.ClassName`, recursively selecting itself; expected `Helper.TObject`. | Resolve underlying class-member access from inside a same-name helper; add a bounded real-path recursion regression before changing dispatch. |
| Explicit helper receiver | `HelpersPass/declared_helper`: `THelper.Proc(TObject.Create)` reports `No arguments expected` at 11:10. | Existing §3.2 helper call-form item. Case hints already match; do not duplicate E3 work. |
| Helper class functions on type aliases | `HelpersPass/dyn_array_create`: `TStrings.Create(...)` reports `Unknown name "TStrings"` at 27:11. | Resolve the array type alias as the helper class-function receiver. Later unknown `sa` messages are cascades; check `Iterate` and helper class-variable state after this blocker. |
| Qualified operator bindings | `HelpersPass/helper_as_overload`: parser expects `;` at the dot in `uses TVec2Helper.Add`, 16:27 and 17:28. | Parse and resolve the qualified helper binding with overload selection. The later missing `Add` unit is a parser-recovery cascade, not a host dependency. |
| Operator expression syntax | `OperatorOverloadPass/c_style`: parser rejects `==` / `!=`. `operator_overloading2`: parser rejects `<<` / `>>`. | Add expression parsing and follow through semantic/operator dispatch; declaration recognition alone does not make these fixtures executable. |
| Builtin operator bindings | `OperatorOverloadPass/operator_implicit`: `binding 'IntToStr' for operator 'implicit' not found` at 1:1. | Operator declaration resolution currently only checks ordinary symbols. Resolve supported builtin bindings and test execution; the integer-to-string assignment diagnostic is a cascade. |
| Lambda Result inference | `LambdaPass/simple_func`: lowercase `lambda result := 450+6 end` is inferred as Void. | Replace case-sensitive Result recognition in return inference with `pkg/ident` comparisons, including conditional branches; preserve the expected lowercase-use hint. This is semantic inference, not hint suppression. |
| Synthetic lambda Result hint | `LambdaPass/immediate`: all three output values match, but a synthetic `Result is never used` hint appears at 17:19. | F1 unused-symbol ownership; cover expression lambdas independently from ordinary function Result diagnostics. |
| Callable overload selection | `OverloadsPass/overload_ambiguous_delegate`: bare `fn` prints `A func` / `B func`, expected `A int` / `B int`; explicit `@fn` cases work. | Existing callable/expected-type work: distinguish implicit invocation from an explicit delegate argument. All case hints match. |
| Class-method receiver identity | `OverloadsPass/overload_class_method`: final output is `hello `, expected `hello tobj`. | Preserve class identity for bare `ClassName` within the overloaded class method. All case hints match. |
| Helper exception source anchor | `FunctionsString/toxml`: caught message says `Unsupported character #11 [line: 13, column: 10]`, expected column 12. | Anchor the error at `ToXML` rather than the receiver expression. This is source positioning, separate from E10d's pretty-renderer duplication. |
| Property-to-property accessors | `PropertyExpressionsPass/read_write_other_property`: `property Mapped read Prop write Prop` reports missing accessor specifiers at 5:7. | Resolve and execute property forwarding, including access-mode validation; later read-only/write-only errors are cascades. |

`HelpersPass/record_array_helper` is the remaining fifth HelpersPass failure. Runtime output is
correct; only two unwanted `X`/`x` hints at 23:21 and 23:35 differ. It stays under E3c's upstream
resolution investigation and is not a new E10 implementation group. The September 18 E3 audit
already superseded older claims that `declared_helper` or the two OverloadsPass failures were
blocked by case hints.

## BuildScripts runner mismatch

Both current fixture scanners enumerate `.pas`. They pair
[`const_inline.pas`](../../testdata/fixtures/BuildScripts/const_inline.pas), a unit containing
`Test`, with an expectation for three invocations. Executing the unit alone correctly produces
no output. Upstream [`UBuildTests.pas`](../../testdata/fixtures/UBuildTests.pas) collects `.dws`
drivers (line 66), loads unit sources as `.pas` (DoNeedUnit), and has separate compilation and
execution checks. This runner does not require JavaScript transpilation.

The matching [`const_inline.dws`](../../testdata/fixtures/BuildScripts/const_inline.dws) calls
`Test(1)`, `Test(2)`, and `Test(3)`. A direct probe:

```sh
/tmp/go-dws-e10-small-cli run --test-envelope --diagnostics=plain --hints strict \
  testdata/fixtures/BuildScripts/const_inline.dws
```

fails because unit search selects the same-name `.dws` before `.pas`, then rejects that driver
as “not a unit (expected 'unit' declaration).” This exposes two prerequisites: reproduce the
category's runner/input selection, then resolve its Pascal units without selecting the driver.
Do not change general extension precedence without considering its existing unit-search tests.
The audit does not score `.dws` drivers, alter the denominator, or diagnose inline constants as
broken. Any future runner change must update the CLI and Go harness together and explicitly
re-measure scoring rules and baseline totals.

## Exclusions and acceptance for follow-ups

PLAN.md §5's host-library and backend exclusions remain in force. Memory's six `external*`
fixtures still require upstream host-registered classes and cleanup hooks (§3.3); UTF-16
surrogate iteration remains an intentional divergence. Unverified runner settings remain a
gate for case-hint parity where applicable, but do not gate categories whose settings are known.

For each implementation follow-up, first reproduce the blocking behavior through the real
compile/run path, then test both valid behavior and error handling. Re-run the representative
fixture after removing its first blocker to reveal independent residuals. Refresh generated
status/baselines only when measured results improve and verify CLI/Go-harness agreement. Do not
edit upstream script expectations, suppress hints broadly, or treat cascaded diagnostics as
independent missing features.
