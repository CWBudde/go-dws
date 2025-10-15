# DWScript Reference

This directory contains reference materials for the go-dws project.

## dwscript-original/

This is a clone of the original DWScript Delphi implementation from:
https://github.com/EricGrange/DWScript

**Purpose:** To serve as the authoritative reference for:
- Language syntax and semantics
- Compiler architecture and design patterns
- Test cases and examples
- Feature completeness verification

**Latest Commit (as of clone):**
- Commit: 5f01a3468452ea75867d4f0e7a0246b107e92332
- Date: Tue Sep 2 12:00:36 2025 +0200
- Author: Eric
- Message: Fixed THttpApi2Server.SetMaxConnections

**Key Directories to Study:**

### Source Code Structure
- `Source/` - Main DWScript source code
  - `dwsComp.pas` - Compiler components
  - `dwsCompiler.pas` - Main compiler implementation
  - `dwsExprs.pas` - Expression handling
  - `dwsSymbols.pas` - Symbol table
  - `dwsTokenizer.pas` - Lexical analyzer
  - `dwsPascalTokenizer.pas` - Pascal-specific tokenizer
  - `dwsStack.pas` - Runtime stack
  - `dwsErrors.pas` - Error handling
  - `dwsFunctions.pas` - Built-in functions
  - `dwsOperators.pas` - Operator handling

### Test Suite
- `Test/` - Comprehensive test suite
  - Study test cases to understand expected behavior
  - Use as validation for go-dws implementation

### Documentation
- `Demos/` - Example programs and use cases
- Review for practical DWScript usage patterns

## Usage Notes

1. **Do not modify** files in `dwscript-original/` - this is a read-only reference
2. When implementing features, cross-reference with the Delphi source
3. When in doubt about semantics, check the test suite
4. Keep this reference up-to-date periodically with: `git pull` in dwscript-original/

## Updating the Reference

To update to the latest DWScript version:

```bash
cd reference/dwscript-original
git pull origin master
cd ../..
```

Then document the new commit hash and date in this README.
