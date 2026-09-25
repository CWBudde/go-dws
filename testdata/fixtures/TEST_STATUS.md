# Test Status Tracking

> **Generated file — do not edit by hand.**
> Regenerate with `just fixture-update` (`FIXTURE_UPDATE_BASELINE=1 go test ./internal/interp -run TestDWScriptFixtures`).

**Generated**: 2026-09-25

## Overall

| Metric | Value |
|---|---|
| Categories | 61 |
| Fixtures (total) | 2041 |
| Passed | 1359 |
| Failed | 655 |
| Skipped (no applicable expectation) | 27 |
| **Scored pass rate** | **67%** (1359/2014) |

## Per-category

Pass% is over *scored* fixtures: a sibling `.txt`, or an empty expectation for output suites.
See [README.md](README.md#a-missing-txt-means-must-print-nothing-upstream) for exclusions.

| Category | Total | Pass | Fail | Skip | Pass% |
|---|---:|---:|---:|---:|---:|
| Algorithms | 53 | 53 | 0 | 0 | 100% |
| ArrayPass | 115 | 102 | 13 | 0 | 89% |
| AssociativeFail | 4 | 2 | 2 | 0 | 50% |
| AssociativePass | 27 | 27 | 0 | 0 | 100% |
| AttributesFail | 2 | 0 | 2 | 0 | 0% |
| AutoFormat | 10 | 0 | 0 | 10 | 0% |
| BigInteger | 16 | 0 | 16 | 0 | 0% |
| BuildScripts | 51 | 7 | 42 | 2 | 14% |
| COMConnector | 19 | 0 | 19 | 0 | 0% |
| COMConnectorFailure | 8 | 0 | 8 | 0 | 0% |
| ClassesLib | 12 | 0 | 12 | 0 | 0% |
| CryptoLib | 17 | 0 | 17 | 0 | 0% |
| DOMParser | 23 | 0 | 23 | 0 | 0% |
| DataBaseLib | 36 | 0 | 36 | 0 | 0% |
| DelegateLib | 14 | 0 | 13 | 1 | 0% |
| EncodingLib | 12 | 12 | 0 | 0 | 100% |
| External | 1 | 0 | 0 | 1 | 0% |
| FailureScripts | 542 | 275 | 254 | 13 | 52% |
| FunctionsByteBuffer | 19 | 19 | 0 | 0 | 100% |
| FunctionsDebug | 3 | 3 | 0 | 0 | 100% |
| FunctionsFile | 15 | 0 | 15 | 0 | 0% |
| FunctionsGlobalVars | 16 | 16 | 0 | 0 | 100% |
| FunctionsMath | 40 | 40 | 0 | 0 | 100% |
| FunctionsMath3D | 2 | 0 | 2 | 0 | 0% |
| FunctionsMathComplex | 6 | 0 | 6 | 0 | 0% |
| FunctionsRTTI | 6 | 0 | 6 | 0 | 0% |
| FunctionsString | 58 | 58 | 0 | 0 | 100% |
| FunctionsTime | 30 | 30 | 0 | 0 | 100% |
| FunctionsVariant | 10 | 10 | 0 | 0 | 100% |
| GenericsFail | 8 | 1 | 7 | 0 | 12% |
| GenericsPass | 23 | 23 | 0 | 0 | 100% |
| GraphicsLib | 4 | 0 | 4 | 0 | 0% |
| HelpersFail | 18 | 9 | 9 | 0 | 50% |
| HelpersPass | 27 | 27 | 0 | 0 | 100% |
| IniFileLib | 2 | 0 | 2 | 0 | 0% |
| InnerClassesFail | 1 | 0 | 1 | 0 | 0% |
| InnerClassesPass | 2 | 2 | 0 | 0 | 100% |
| InterfacesFail | 19 | 6 | 13 | 0 | 32% |
| InterfacesPass | 33 | 33 | 0 | 0 | 100% |
| JSFilterScripts | 2 | 2 | 0 | 0 | 100% |
| JSFilterScriptsFail | 1 | 1 | 0 | 0 | 100% |
| JSONConnectorFail | 9 | 2 | 7 | 0 | 22% |
| JSONConnectorPass | 82 | 82 | 0 | 0 | 100% |
| LambdaFail | 6 | 0 | 6 | 0 | 0% |
| LambdaPass | 6 | 6 | 0 | 0 | 100% |
| Linq | 7 | 0 | 7 | 0 | 0% |
| LinqJSON | 6 | 0 | 6 | 0 | 0% |
| Memory | 13 | 7 | 6 | 0 | 54% |
| OperatorOverloadFail | 6 | 3 | 3 | 0 | 50% |
| OperatorOverloadPass | 8 | 8 | 0 | 0 | 100% |
| OverloadsFail | 14 | 3 | 11 | 0 | 21% |
| OverloadsPass | 39 | 39 | 0 | 0 | 100% |
| PropertyExpressionsFail | 10 | 3 | 7 | 0 | 30% |
| PropertyExpressionsPass | 19 | 19 | 0 | 0 | 100% |
| SetOfFail | 14 | 13 | 1 | 0 | 93% |
| SetOfPass | 25 | 25 | 0 | 0 | 100% |
| SimpleScripts | 443 | 391 | 52 | 0 | 88% |
| SystemInfoLib | 3 | 0 | 3 | 0 | 0% |
| TabularLib | 16 | 0 | 16 | 0 | 0% |
| TimeSeriesLib | 5 | 0 | 5 | 0 | 0% |
| WebLib | 3 | 0 | 3 | 0 | 0% |
