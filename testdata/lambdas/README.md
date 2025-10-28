# Lambda Test Scripts

This directory contains end-to-end test scripts for DWScript lambda expressions and closures.

## Test Files

### basic_lambda.dws
Tests fundamental lambda functionality:
- Lambda creation and invocation
- Multiple parameters
- Variable storage and reassignment
- Both shorthand (`=>`) and full (`begin/end`) syntax
- Different return types

### closure.dws
Tests closure capture semantics:
- Single and multiple variable capture
- Mutation of captured variables
- Reference semantics (changes visible in outer scope)
- Reading updated captured variables

### higher_order.dws
Tests built-in higher-order functions:
- `Map(array, lambda)` - Transform array elements
- `Filter(array, lambda)` - Filter by predicate
- `Reduce(array, lambda, initial)` - Aggregate values
- `ForEach(array, lambda)` - Execute for side effects
- Chained operations

### nested_lambda.dws
Tests nested lambda capabilities:
- Lambdas returning lambdas
- Multi-level closure capture
- Multiple lambdas sharing captured variables
- Counter/accumulator patterns

## Running Tests

```bash
# Run all lambda tests
for f in testdata/lambdas/*.dws; do
    echo "Running $f..."
    ./bin/dwscript run "$f"
done

# Run a specific test
./bin/dwscript run testdata/lambdas/basic_lambda.dws

# Run integration tests
go test ./cmd/dwscript -run TestLambdaIntegration
```

## Expected Output

Each `.dws` file has a corresponding `.txt` file with the expected output.
