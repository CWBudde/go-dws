# Test Status Tracking

> **Generated file — do not edit by hand.**
> Regenerate with `just fixture-update` (`FIXTURE_UPDATE_BASELINE=1 go test ./internal/interp -run TestDWScriptFixtures`).

**Generated**: 2026-09-13

## Overall

| Metric | Value |
|---|---|
| Categories | 61 |
| Fixtures (total) | 2044 |
| Passed | 1121 |
| Failed | 845 |
| Skipped (no applicable expectation) | 78 |
| **Scored pass rate** | **57%** (1121/1966) |

## Per-category

Pass% is over *scored* fixtures: a sibling `.txt`, or an empty expectation for output suites.
See [README.md](README.md#a-missing-txt-means-must-print-nothing-upstream) for exclusions.

| Category | Total | Pass | Fail | Skip | Pass% |
|---|---:|---:|---:|---:|---:|
| Algorithms | 53 | 53 | 0 | 0 | 100% |
| ArrayPass | 115 | 96 | 19 | 0 | 83% |
| AssociativeFail | 4 | 1 | 3 | 0 | 25% |
| AssociativePass | 27 | 27 | 0 | 0 | 100% |
| AttributesFail | 2 | 0 | 2 | 0 | 0% |
| AutoFormat | 10 | 0 | 0 | 10 | 0% |
| BigInteger | 16 | 0 | 16 | 0 | 0% |
| BuildScripts | 54 | 0 | 1 | 53 | 0% |
| COMConnector | 19 | 0 | 19 | 0 | 0% |
| COMConnectorFailure | 8 | 0 | 8 | 0 | 0% |
| ClassesLib | 12 | 0 | 12 | 0 | 0% |
| CryptoLib | 17 | 0 | 17 | 0 | 0% |
| DOMParser | 23 | 0 | 23 | 0 | 0% |
| DataBaseLib | 36 | 0 | 36 | 0 | 0% |
| DelegateLib | 14 | 0 | 13 | 1 | 0% |
| EncodingLib | 12 | 12 | 0 | 0 | 100% |
| External | 1 | 0 | 0 | 1 | 0% |
| FailureScripts | 542 | 156 | 373 | 13 | 29% |
| FunctionsByteBuffer | 19 | 19 | 0 | 0 | 100% |
| FunctionsDebug | 3 | 3 | 0 | 0 | 100% |
| FunctionsFile | 15 | 0 | 15 | 0 | 0% |
| FunctionsGlobalVars | 16 | 15 | 1 | 0 | 94% |
| FunctionsMath | 40 | 29 | 11 | 0 | 72% |
| FunctionsMath3D | 2 | 0 | 2 | 0 | 0% |
| FunctionsMathComplex | 6 | 0 | 6 | 0 | 0% |
| FunctionsRTTI | 6 | 0 | 6 | 0 | 0% |
| FunctionsString | 58 | 57 | 1 | 0 | 98% |
| FunctionsTime | 30 | 30 | 0 | 0 | 100% |
| FunctionsVariant | 10 | 10 | 0 | 0 | 100% |
| GenericsFail | 8 | 0 | 8 | 0 | 0% |
| GenericsPass | 23 | 23 | 0 | 0 | 100% |
| GraphicsLib | 4 | 0 | 4 | 0 | 0% |
| HelpersFail | 18 | 0 | 18 | 0 | 0% |
| HelpersPass | 27 | 22 | 5 | 0 | 81% |
| IniFileLib | 2 | 0 | 2 | 0 | 0% |
| InnerClassesFail | 1 | 0 | 1 | 0 | 0% |
| InnerClassesPass | 2 | 2 | 0 | 0 | 100% |
| InterfacesFail | 19 | 1 | 18 | 0 | 5% |
| InterfacesPass | 33 | 21 | 12 | 0 | 64% |
| JSFilterScripts | 2 | 2 | 0 | 0 | 100% |
| JSFilterScriptsFail | 1 | 1 | 0 | 0 | 100% |
| JSONConnectorFail | 9 | 2 | 7 | 0 | 22% |
| JSONConnectorPass | 82 | 68 | 14 | 0 | 83% |
| LambdaFail | 6 | 0 | 6 | 0 | 0% |
| LambdaPass | 6 | 4 | 2 | 0 | 67% |
| Linq | 7 | 0 | 7 | 0 | 0% |
| LinqJSON | 6 | 0 | 6 | 0 | 0% |
| Memory | 13 | 6 | 7 | 0 | 46% |
| OperatorOverloadFail | 6 | 0 | 6 | 0 | 0% |
| OperatorOverloadPass | 8 | 5 | 3 | 0 | 62% |
| OverloadsFail | 14 | 0 | 14 | 0 | 0% |
| OverloadsPass | 39 | 37 | 2 | 0 | 95% |
| PropertyExpressionsFail | 10 | 0 | 10 | 0 | 0% |
| PropertyExpressionsPass | 19 | 18 | 1 | 0 | 95% |
| SetOfFail | 14 | 5 | 9 | 0 | 36% |
| SetOfPass | 25 | 25 | 0 | 0 | 100% |
| SimpleScripts | 443 | 371 | 72 | 0 | 84% |
| SystemInfoLib | 3 | 0 | 3 | 0 | 0% |
| TabularLib | 16 | 0 | 16 | 0 | 0% |
| TimeSeriesLib | 5 | 0 | 5 | 0 | 0% |
| WebLib | 3 | 0 | 3 | 0 | 0% |
