# DWScript Test Fixtures

This directory contains the comprehensive test suite copied from the original DWScript project. It provides extensive coverage of the DWScript language features and serves as a reference implementation for verifying compatibility.

## Overview

Raw file counts (what `find` sees) and scored counts (what the Go harness compares) differ, so keep them apart:

- **Raw contents**: 64 directories, ~2,100 `.pas` scripts, ~2,050 `.txt` expectations, plus `.jstxt`, `.optimized.txt`, `.fpctxt` variants, 123 `.dws` JavaScript filter scripts and support files (`.dll`, `.s3db`, `.ply`, ...).
- **Scored set**: `internal/interp/fixture_test.go` (`discoverFixtureCategories`) treats every directory directly under `testdata/fixtures/` that contains at least one `.pas` file as a category (61 today; `Data`, `HTMLFilterScripts` and `Model3D` have no `.pas` and are not categories, nested subdirectories are not walked). Each `.pas` is compared against its sibling `.txt` only (`runFixtureTest`). A `.pas` without a `.txt` is scored against empty output except for the category exclusions documented below (36 scored this way, 78 skipped). All other expectation variants are ignored, see [Expected-output variants that are not scored](#expected-output-variants-that-are-not-scored).

Live per-category pass/skip numbers are generated into [TEST_STATUS.md](TEST_STATUS.md); do not rely on the counts in this file for scoring.

## Directory Structure

### Core Language Tests

#### Pass Cases (Working Features)
- **SimpleScripts** (442 tests) - Basic language features and scripts
- **Algorithms** (53 tests) - Algorithm implementations
- **ArrayPass** (115 tests) - Array operations and features
- **AssociativePass** (27 tests) - Associative arrays/maps
- **SetOfPass** (25 tests) - Set operations
- **OverloadsPass** (39 tests) - Function/method overloading
- **OperatorOverloadPass** (8 tests) - Operator overloading
- **GenericsPass** (23 tests) - Generic types and methods
- **HelpersPass** (27 tests) - Type helpers
- **LambdaPass** (6 tests) - Lambda expressions
- **PropertyExpressionsPass** (19 tests) - Property expressions
- **InterfacesPass** (33 tests) - Interface declarations and usage
- **InnerClassesPass** (2 tests) - Nested class declarations

#### Failure Cases (Error Handling)
- **FailureScripts** (541 tests) - Compilation and runtime errors
- **AssociativeFail** (4 tests) - Associative array error cases
- **SetOfFail** (14 tests) - Set operation error cases
- **OverloadsFail** (14 tests) - Overloading error cases
- **OperatorOverloadFail** (6 tests) - Operator overload error cases
- **GenericsFail** (8 tests) - Generic type error cases
- **HelpersFail** (18 tests) - Type helper error cases
- **LambdaFail** (6 tests) - Lambda expression error cases
- **PropertyExpressionsFail** (10 tests) - Property expression error cases
- **InterfacesFail** (19 tests) - Interface error cases
- **InnerClassesFail** (1 test) - Nested class error cases
- **AttributesFail** (2 tests) - Attribute error cases

### Built-in Functions

- **FunctionsMath** (40 tests) - Mathematical functions
- **FunctionsMath3D** (2 tests) - 3D math functions
- **FunctionsMathComplex** (6 tests) - Complex number functions
- **FunctionsString** (58 tests) - String manipulation functions
- **FunctionsTime** (30 tests) - Date/time functions
- **FunctionsByteBuffer** (19 tests) - Byte buffer operations
- **FunctionsFile** (15 tests) - File I/O functions
- **FunctionsGlobalVars** (16 tests) - Global variable functions
- **FunctionsVariant** (10 tests) - Variant type functions
- **FunctionsRTTI** (6 tests) - Runtime type information functions
- **FunctionsDebug** (3 tests) - Debug/diagnostic functions

### Library Tests

- **ClassesLib** (12 tests) - Classes library tests [requires external libs]
- **JSONConnectorPass** (82 tests) - JSON parsing and generation
- **JSONConnectorFail** (9 tests) - JSON error cases
- **LinqJSON** (6 tests) - LINQ-style JSON queries
- **Linq** (7 tests) - LINQ-style queries
- **DOMParser** (23 tests) - XML/DOM parsing
- **DelegateLib** (14 tests) - Delegate library tests [requires external libs]
- **DataBaseLib** (36 tests) - Database operations [requires sqlite3.dll]
- **COMConnector** (19 tests) - COM interop tests [Windows only]
- **COMConnectorFailure** (8 tests) - COM error cases [Windows only]
- **EncodingLib** (12 tests) - Encoding/decoding functions
- **CryptoLib** (17 tests) - Cryptographic functions
- **TabularLib** (16 tests) - Tabular data operations
- **TimeSeriesLib** (5 tests) - Time series data
- **SystemInfoLib** (3 tests) - System information [requires external libs]
- **IniFileLib** (2 tests) - INI file operations
- **WebLib** (3 tests) - Web/HTTP operations [requires external libs]
- **GraphicsLib** (4 tests) - Graphics operations [requires external libs]

### Advanced Features

- **BigInteger** (16 tests) - Arbitrary precision integers
- **Memory** (13 tests) - Memory management tests
- **AutoFormat** (10 tests) - Code auto-formatting

### Codegen Tests (Stage 12)

- **BuildScripts** (54 tests) - Build and compilation tests [requires JS transpilation]
- **JSFilterScripts** (59 files) - JavaScript filter scripts [requires JS transpilation]
- **JSFilterScriptsFail** (6 files) - JavaScript filter error cases [requires JS transpilation]
- **HTMLFilterScripts** (10 tests) - HTML filter scripts [requires JS transpilation]

## File Naming Conventions

### Test Files
- **`.pas`** - DWScript source code (Pascal syntax)
- **`.txt`** - Expected output or error messages. This is the **only** expectation the go-dws harness scores.

### Expectation variants (never scored by go-dws)
- **`.optimized.txt`** - Upstream's expectation when the compiler runs with `coOptimize`. Upstream runs each suite twice (optimized and non-optimized) and picks this file only in the optimized run (`reference/dwscript-original/Test/UScriptTests.pas`, lines 219-224 for execution and 329-333 for failure tests); otherwise it uses `.txt`. Constant folding changes which hints and errors appear, so the content is not merely "fewer hints". go-dws has no optimizer and therefore corresponds to the non-optimized configuration, for which upstream mandates `.txt`.
- **`.jstxt`** - Expected output of the JavaScript code-generation backend (consumed upstream only by `UJSCodeGenTests.pas` / `UJSFilterTests.pas`).
- **`.fpctxt`** - Free Pascal variant of an expectation (`{$ifdef FPC}` upstream).

### Codegen Files
- **`.dws`** - JavaScript filter scripts with `<%pas2js ... %>` blocks

### Support Files
- **`.dll`** - Windows DLL dependencies (e.g., `sqlite3.dll`, `BeaEngine64.dll`)
- **`.s3db`** - SQLite database files
- **`.ply`** - 3D model files
- **`.dfm`** / **`.inc`** - Delphi form and include files

## Running Tests

### Fixed timezone

Both runners execute every fixture with `TZ=Europe/Berlin`. Four `FunctionsTime` fixtures come
from DWScript's own suite and assume a Central European host: `incmonth` and `local_utc_unix`
hard-code the +1/+2 offsets in their expected output, while `encode` and `utc` print
`Cannot perform test for GMT+0` under UTC. Without a fixed zone the category would score 27,
25 or 23 depending on where the suite runs. The zone is set in `internal/interp/fixture_test.go`
(`fixtureTimeZone`, applied to each worker subprocess) and in `cmd/fixture-report/main.go`
(the same constant, applied to each CLI invocation); keep the two in sync. `TestDWScriptFixtures`
fails immediately if the host has no zone database.

### Run All Tests
```bash
go test -v ./internal/interp -run TestDWScriptFixtures
```

### Run Specific Category
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts
```

### Run Specific Test
```bash
go test -v ./internal/interp -run TestDWScriptFixtures/SimpleScripts/hello
```

### Run with Coverage
```bash
go test -cover ./internal/interp -run TestDWScriptFixtures
```

## Test Format

### Success Tests (expectErrors: false)
Test files in "Pass" categories should execute without errors and produce output matching the `.txt` file.

**Example**: `SimpleScripts/hello.pas`
```pascal
program Hello;
begin
  PrintLn('Hello, World!');
end.
```

**Expected**: `SimpleScripts/hello.txt`
```
Hello, World!
```

### Failure Tests (expectErrors: true)
Test files in "Fail" and "Failure" categories should produce compilation or runtime errors matching the `.txt` file.

**Example**: `FailureScripts/abstract_method.pas`
```pascal
Type TAction = Class Abstract
  Function GetTitle : String; Virtual; Abstract;
End;
var a := TAction.Create;  // Line 16 - error!
```

**Expected**: `FailureScripts/abstract_method.txt`
```
Hint: Result is never used [line: 13, column: 4]
Error: Trying to create an instance of an abstract class [line: 16, column: 18]
```

Error messages include precise `[line: X, column: Y]` position information.

## Current Implementation Status

See `TEST_STATUS.md` for detailed pass/fail counts per category and known issues.

### Feature Readiness

✅ **Ready to Test** (Stages 1-8)
- Basic expressions and statements
- Control flow (if/while/for/repeat)
- Functions and procedures
- Type checking
- Arrays and records
- Enumerations
- Classes and OOP features
- Interfaces
- Generics
- Helpers
- Lambdas

❌ **Not Yet Implemented** (Stage 9+)
- JavaScript codegen/transpilation
- Some advanced libraries (COM, graphics)
- Platform-specific features

## Dependencies

### External Libraries
Some test categories require external dependencies:

- **DataBaseLib** - Requires `sqlite3.dll` (Windows) or equivalent shared library
- **COMConnector** / **COMConnectorFailure** - Windows only (COM interop)
- **GraphicsLib** - Requires graphics library
- **WebLib** - Requires HTTP/web library
- **SystemInfoLib** - Platform-specific system information
- **DelegateLib**, **ClassesLib** - May require additional library implementations

These tests will be skipped if dependencies are not available.

### Codegen Dependencies
JavaScript transpilation tests require Stage 12 (Codegen) implementation and are currently skipped.

## Contributing

When fixing failing tests:

1. Focus on one category at a time
2. Start with SimpleScripts and core language features
3. Update `TEST_STATUS.md` with progress
4. Document any differences from original DWScript behavior
5. Add comments explaining complex fixes

## References

- Original DWScript: https://www.delphitools.info/dwscript/
- DWScript Language Reference: https://www.delphitools.info/dwscript/language/
- Project PLAN.md: See `../../PLAN.md` for implementation roadmap

## Expected-output variants that are not scored

Only the sibling `.txt` file is compared. `.jstxt` (68 files, JavaScript backend), `.optimized.txt` (31 files: 22 in FailureScripts, 3 OverloadsFail, 2 SimpleScripts, 2 OperatorOverloadFail, 1 SetOfFail, 1 HelpersFail; every one has a sibling `.txt`) and `.fpctxt` (2 files, Free Pascal) are never opened by the Go harness or by `cmd/fixture-report`. A missing plain `.txt` uses the scoring policy in the next section; none of these variants substitutes for it.

**Decision (2026-09-06): `.optimized.txt` is not accepted as an alternative expected output, neither for FailureScripts nor anywhere else.** Reasons:

- Upstream selects `.optimized.txt` only when compiling with `coOptimize` and mandates `.txt` in the non-optimized configuration. go-dws has no optimizer, so `.txt` is the matching expectation.
- The optimized variants encode optimizer behaviour, not looser output. Examples: `FailureScripts/abstract_method` drops an unused-Result hint; `FailureScripts/unused_variables` is empty; `FailureScripts/ignore_result` turns a hint into `Compile Error: Evaluation of "StrToInt" failed ...` through constant folding; `FailureScripts/div_by_zero_int` adds `Syntax Error: Division by zero`.
- Accepting either file would loosen 22 FailureScripts assertions and in some cases could only be matched by implementing an optimizer.

If a fixture fails only because go-dws emits a hint or error that the optimized variant hides, fix the diagnostics to match `.txt`, do not fall back to `.optimized.txt`.

## A missing `.txt` means "must print nothing" upstream

Implemented 2026-09-13 (T7). Both go-dws runners use
`internal/fixtureconfig` to read expectations and choose category hint levels. An existing `.txt`
is always compared, including in otherwise excluded categories. Read or decode errors fail the
fixture; only a missing file can trigger the policy below.

Upstream's execution runners score a missing expectation against empty output:

```pascal
if FileExists(resultsFileName) then begin
   expectedResult.LoadFromFile(resultsFileName);
   CheckEquals(expectedResult.Text, output, FTests[i]);
end else CheckEquals('', output, FTests[i]);
```

See `UScriptTests.pas:238` and `UMemoryTests.pas:254` in this directory.

**36 imported `.pas` fixtures** are now scored against silence — Memory 10, SimpleScripts 7,
InterfacesPass 5, FunctionsMath 5, FunctionsTime 3, JSFilterScripts 2, FunctionsGlobalVars 2,
FunctionsVariant 1 and JSFilterScriptsFail 1. Of these, **29 pass and 7 fail** as of 2026-09-13.
The scored denominator grows from 1,930 to 1,966; the total pass rate rises slightly.
The JSFilterScripts categories contain supporting Pascal units: upstream's `UJSFilterTests.pas`
collects `.dws` scripts. Our existing `.pas` discovery checks these units on their own;
JSFilterScriptsFail still uses compile-only mode and requires empty diagnostics.

Three groups remain **unscored when their `.txt` is missing** (78 fixtures):

- **BuildScripts (53) and AutoFormat (10)** use different upstream runners, not output comparisons.
- **External (1) and DelegateLib (1)** require host setup excluded from this work.
- **FailureScripts (13)** lack exact diagnostic expectations. Upstream's `CompilationFailure`
  runner (`UScriptTests.pas:344`) requires nonempty diagnostics when a `.txt` is missing; it does
  **not** require silence. These remain skipped in our exact-output metric. The affected names
  are `duplicate_field`, `duplicate_property`, `exit_result1`, `exit_result3`, `exit_result4`,
  `for_non_int_bounds1`, `for_non_int_bounds2`, `if1`, `if2`, `invalid_float`, `invalid_hex`,
  `invalid_integer` and `no_switch`.

Memory, Algorithms and FunctionsString use the compiler's **normal** (default) hint level.
Other categories use **pedantic**. This lets `Memory/obj_local` match the upstream memory runner
without suppressing hints in other suites. Missing expectations still detect unexpected output,
compile diagnostics and runtime errors; they do not mean an automatic pass.
