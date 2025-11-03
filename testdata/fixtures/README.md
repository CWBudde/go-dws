# DWScript Test Fixtures

This directory contains the comprehensive test suite copied from the original DWScript project. It provides extensive coverage of the DWScript language features and serves as a reference implementation for verifying compatibility.

## Overview

- **Total Test Files**: ~4,400 files
- **Test Categories**: 64 directories
- **Test Scripts**: 2,098 `.pas` files
- **Expected Outputs**: 2,083 `.txt` files
- **JavaScript Tests**: 123 `.dws` files + 68 `.jstxt` files
- **External Dependencies**: Various `.dll`, `.s3db`, `.ply`, and other support files

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
- **`.txt`** - Expected output or error messages
- **`.optimized.txt`** - Expected output for optimized builds (rare)

### Codegen Files
- **`.dws`** - JavaScript filter scripts with `<%pas2js ... %>` blocks
- **`.jstxt`** - Expected JavaScript output

### Support Files
- **`.dll`** - Windows DLL dependencies (e.g., `sqlite3.dll`, `BeaEngine64.dll`)
- **`.s3db`** - SQLite database files
- **`.ply`** - 3D model files
- **`.dfm`** / **`.inc`** - Delphi form and include files

## Running Tests

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

✅ **Ready to Test** (Stages 1-6)
- Basic expressions and statements
- Control flow (if/while/for/repeat)
- Functions and procedures
- Type checking
- Arrays and records
- Enumerations

⚠️ **Partially Ready** (Stage 7-8)
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
