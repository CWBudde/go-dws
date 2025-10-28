# Function Pointer Test Scripts

This directory contains test scripts for DWScript function and method pointer features (PLAN.md task 9.170).

## Implementation Status

### ✅ Fully Implemented (Can be parsed and analyzed)

The following components are **100% complete**:

- **Type System** (tasks 9.146-9.150): `internal/types/function_pointer.go`
  - `FunctionPointerType` and `MethodPointerType`
  - Signature compatibility checking
  - Full test coverage in `types/function_pointer_test.go` (773 lines)

- **AST Nodes** (tasks 9.151-9.153): `internal/ast/function_pointer.go`
  - `FunctionPointerTypeNode` for type declarations
  - `AddressOfExpression` for `@functionName` syntax
  - Full test coverage in `ast/function_pointer_test.go` (216 lines)

- **Parser** (tasks 9.154-9.158): `internal/parser/`
  - Parses function pointer type declarations
  - Parses address-of operator `@`
  - Handles `of object` syntax for method pointers
  - Full test coverage in `parser/function_pointer_test.go` (376 lines)

- **Semantic Analysis** (tasks 9.159-9.163): `internal/semantic/function_pointer_analyzer.go`
  - Type declaration validation
  - Address-of expression analysis
  - Assignment compatibility checking
  - Call validation
  - Full test coverage in `semantic/function_pointer_test.go` (700 lines)

### ❌ Not Yet Implemented (Cannot execute)

The **interpreter runtime** is **NOT** implemented:

- ❌ Task 9.164: `FunctionPointerValue` runtime representation
- ❌ Task 9.165: Address-of operator evaluation
- ❌ Task 9.166: Function pointer call execution
- ❌ Task 9.167: Function pointer assignment at runtime
- ❌ Task 9.168: Interpreter tests
- ❌ Task 9.169: Passing function pointers as parameters (runtime)

## Current Capabilities

### What Works NOW

1. **Parsing**: All test scripts parse successfully
   ```bash
   ./bin/dwscript parse testdata/function_pointers/basic_function_pointer.dws
   ```

2. **Semantic Analysis**: Type checking and validation works
   - Function pointer type declarations are validated
   - Signature compatibility is checked
   - Address-of expressions are analyzed
   - Invalid assignments are detected

3. **Error Detection**: Invalid cases are properly diagnosed

### What Does NOT Work Yet

1. **Execution**: Scripts cannot run
   ```bash
   ./bin/dwscript run testdata/function_pointers/basic_function_pointer.dws
   # Will fail: interpreter doesn't support function pointers yet
   ```

2. **Runtime Operations**:
   - Cannot create function pointer values
   - Cannot call through function pointers
   - Cannot pass function pointers as arguments at runtime

## Test Scripts

### Valid Syntax Tests (Parse Successfully)

1. **basic_function_pointer.dws**
   - Function pointer type declarations
   - Address-of operator `@Function`
   - Assignment and calling through pointers
   - Procedure pointers
   - Multi-parameter function pointers

2. **callback.dws**
   - Higher-order functions with callbacks
   - Passing function pointers as parameters
   - Array processing with callback procedures
   - Chaining transformations

3. **method_pointer.dws**
   - Method pointers with `of object` syntax
   - Capturing `Self` context
   - Switching method pointers between objects
   - Method pointers as parameters

4. **sort_with_comparator.dws**
   - Practical sorting example
   - Custom comparator functions
   - Flexible sort orderings (ascending, descending, by absolute value)
   - Demonstrates real-world use case

5. **procedure_pointer.dws**
   - Procedure pointers (no return value)
   - Simple procedures (no parameters)
   - Procedures with various parameter types
   - Higher-order procedures

### Error Cases

6. **invalid_cases.dws**
   - Type mismatches
   - Wrong argument counts/types
   - Undefined function references
   - Invalid type declarations
   - Method/function pointer incompatibilities
   - Currently has errors commented out for reference

## Testing Strategy

### Phase 1: Parser/Semantic Validation (CURRENT)

Test that scripts have valid syntax and semantics:

```bash
# Parse test (should succeed)
./bin/dwscript parse testdata/function_pointers/basic_function_pointer.dws

# Run CLI integration tests
go test ./cmd/dwscript -run TestFunctionPointerParsing
```

### Phase 2: Runtime Execution (FUTURE)

Once tasks 9.164-9.169 are implemented:

1. Implement `FunctionPointerValue` in `internal/interp/value.go`
2. Implement evaluation in interpreter
3. Create expected output files (`.txt`)
4. Add execution tests to CLI test suite
5. Verify scripts run and produce correct output

## Expected Outputs (Pending Interpreter)

Expected output files will be added once the interpreter supports function pointers:

- `basic_function_pointer.txt` - Expected console output
- `callback.txt` - Expected console output
- `method_pointer.txt` - Expected console output
- `sort_with_comparator.txt` - Expected console output
- `procedure_pointer.txt` - Expected console output

## References

- **PLAN.md**: Lines 618-748 (Function/Method Pointers section)
- **Implementation Guide**: `docs/missing-features-recommendations.md` lines 171-203
- **Type System**: `internal/types/function_pointer.go`
- **AST Nodes**: `internal/ast/function_pointer.go`
- **Parser**: `internal/parser/parser.go` (function pointer sections)
- **Semantic Analysis**: `internal/semantic/function_pointer_analyzer.go`

## Next Steps

To complete function pointer support:

1. ✅ Task 9.170: Create test scripts (THIS TASK - DONE)
2. ⏳ Task 9.171: Add CLI integration tests (IN PROGRESS)
3. ⏳ Task 9.172: Document limitations
4. ⏳ Tasks 9.164-9.169: Implement interpreter support (REQUIRED FOR EXECUTION)

Once interpreter support is complete, these test scripts will be executable and can verify the full runtime behavior of function pointers in DWScript.
