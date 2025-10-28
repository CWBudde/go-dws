# DWScript Examples

This directory collects runnable DWScript programs that double as documentation and regression tests for the playground snippets defined in `playground/js/examples.js`.

## Layout

- `scripts/` – Source `.dws` programs that can be executed with the CLI or embedded engine.
- `wasm/` – Standalone WebAssembly playground that consumes the same sample programs.

## Running an Example

```bash
go run ./cmd/dwscript --file examples/scripts/hello_world.dws
```

Replace the file path with any of the scripts listed below. The CLI prints the DWScript program output directly to stdout.

## Available Scripts

| File | Description |
| --- | --- |
| `hello_world.dws` | Minimal “Hello, World!” and greeting banner. |
| `fibonacci.dws` | Recursive Fibonacci function that prints the first ten numbers. |
| `factorial.dws` | Recursive and iterative factorial implementations side by side. |
| `loops.dws` | Demonstrates `for`, `while`, and `repeat…until` loops. |
| `functions.dws` | Procedures, functions, and simple string helpers. |
| `classes.dws` | Basic OOP with a `TPerson` class and properties. |
| `math_operations.dws` | Integer/float arithmetic and compound assignments. |
| `case_statement.dws` | Uses `case` to map numeric values to friendly labels. |
| `palindrome_checker.dws` | Demonstrates string functions and custom helpers. |
| `prime_numbers.dws` | Generates prime numbers up to a limit with a helper function. |
| `multiplication_table.dws` | Prints a formatted multiplication grid using nested loops. |

### Playground Integration

The browser playground (`playground/index.html`) loads these scripts dynamically via `playground/js/examples.js`. When hosting the playground standalone (without the repository root), copy the `.dws` files into `playground/examples/` so the loader can fall back to that directory.

## Test Coverage

`go test ./pkg/dwscript -run TestExampleScripts` executes every script under `examples/scripts/` to ensure they continue to parse and run successfully
