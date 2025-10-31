# FFI Example

This example demonstrates how to use the Foreign Function Interface (FFI) to call Go functions from DWScript.

## Files

- `main.go` - Go program that registers various functions
- `demo.dws` - DWScript script that calls the registered functions

## Running

From the project root:

```bash
go run examples/ffi/main.go
```

Or build and run:

```bash
go build -o ffi-demo examples/ffi/main.go
./ffi-demo
```

## What's Demonstrated

### Math Functions
- Basic arithmetic (Add, Multiply)
- Power calculation
- Safe division with error handling

### String Functions
- Case conversion (ToUpper, ToLower)
- String reversal
- Contains/Search
- Split and Join

### Array Functions
- Sum array elements
- Find maximum value
- Filter (even numbers)
- Map (double all values)

### Error Handling
- Functions returning errors
- Error propagation to exceptions
- Panic recovery
- Try/except blocks

### Utility Functions
- Environment variables
- Map/Record creation
- String formatting
- String repetition

## Key Features

1. **Type Conversion**: Automatic conversion between Go and DWScript types
2. **Error Handling**: Go errors become DWScript exceptions
3. **Panic Recovery**: Go panics are caught and converted to exceptions
4. **Arrays**: Seamless array passing and return
5. **Maps**: Go maps become DWScript records/associative arrays

## Expected Output

The demo script produces output showing:
- Successful function calls with results
- Error handling with try/except
- Panic recovery
- Array and string manipulation
- Map/record access

All operations should complete without crashes, with errors properly caught and handled.
