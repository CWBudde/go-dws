# Foreign Function Interface (FFI)

The Foreign Function Interface (FFI) enables DWScript scripts to call Go functions, bridging the gap between DWScript code and the Go ecosystem. This allows scripts to access file I/O, network operations, cryptography, and other system-level functionality that DWScript doesn't provide natively.

## Overview

FFI works by registering Go functions with the DWScript engine. Once registered, these functions can be called from DWScript scripts just like regular DWScript functions. The engine automatically handles type conversion between Go and DWScript types, as well as error propagation.

## Basic Usage

### Registering Functions

Use `engine.RegisterFunction(name, func)` to register a Go function:

```go
engine, _ := dwscript.New()

// Register a simple function
engine.RegisterFunction("Add", func(a, b int64) int64 {
    return a + b
})

// Register a function with error handling
engine.RegisterFunction("ReadFile", func(filename string) (string, error) {
    data, err := os.ReadFile(filename)
    return string(data), err
})
```

### Calling from DWScript

Once registered, functions can be called like any DWScript function:

```pascal
var sum := Add(40, 2);
PrintLn(IntToStr(sum)); // Output: 42

var content := ReadFile('example.txt');
PrintLn(content);
```

## Type Mapping

The FFI automatically converts between Go and DWScript types:

### Primitive Types

| Go Type | DWScript Type | Notes |
|---------|---------------|-------|
| `int`, `int64`, `int32`, `int16`, `int8` | `Integer` | All integer types map to DWScript Integer |
| `float64`, `float32` | `Float` | Both float types map to DWScript Float |
| `string` | `String` | Direct string mapping |
| `bool` | `Boolean` | Direct boolean mapping |

### Collection Types

| Go Type | DWScript Type | Notes |
|---------|---------------|-------|
| `[]T` | `array of T` | Dynamic arrays, element type must be supported |
| `map[string]T` | `record` | String-keyed maps become records |

### Unsupported Types

The following Go types are not currently supported:
- Complex numbers (`complex64`, `complex128`)
- Channels (`chan T`)
- Functions (`func`)
- Interfaces (except `error`)
- Pointers (except for `nil`)
- Structs (use maps or slices instead)
- Custom types

## Function Signatures

FFI supports several Go function signatures:

### Value Return

```go
func(a, b int64) int64  // Returns a value
```

### Error Return

```go
func(filename string) (string, error)  // Returns value and error
func() error                           // Returns only error (procedure)
```

### Procedures (No Return)

```go
func(message string)  // No return value
```

## Error Handling

Go errors are automatically converted to DWScript `EHost` exceptions:

```go
engine.RegisterFunction("ReadFile", func(filename string) (string, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return "", err  // This becomes an EHost exception
    }
    return string(data), nil
})
```

In DWScript, catch these exceptions:

```pascal
try
    var content := ReadFile('nonexistent.txt');
    PrintLn(content);
except
    on E: EHost do
        PrintLn('Error: ' + E.Message);
end;
```

### Exception Details

`EHost` exceptions include:

- `Message`: The error message from Go
- `ExceptionClass`: The Go error type name (e.g., "*fs.PathError")

## Examples

### HTTP Client

```go
package main

import (
    "io"
    "net/http"
    "github.com/cwbudde/go-dws/pkg/dwscript"
)

func main() {
    engine, _ := dwscript.New()

    // Register HTTP GET function
    engine.RegisterFunction("HttpGet", func(url string) (string, error) {
        resp, err := http.Get(url)
        if err != nil {
            return "", err
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return "", err
        }

        return string(body), nil
    })

    // Run DWScript code
    result, err := engine.Eval(`
        var html := HttpGet('https://httpbin.org/get');
        PrintLn('Response length: ' + IntToStr(Length(html)));
    `)

    if err != nil {
        panic(err)
    }

    if !result.Success {
        panic("Script execution failed")
    }
}
```

### File Operations

```go
engine.RegisterFunction("ListFiles", func(dir string) ([]string, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }

    files := make([]string, 0, len(entries))
    for _, entry := range entries {
        files = append(files, entry.Name())
    }

    return files, nil
})

engine.RegisterFunction("FileExists", func(filename string) bool {
    _, err := os.Stat(filename)
    return !os.IsNotExist(err)
})
```

DWScript usage:

```pascal
var files := ListFiles('.');
for var i := 0 to Length(files) - 1 do
    PrintLn(files[i]);

if FileExists('config.txt') then
    PrintLn('Config file exists');
```

### JSON Processing

```go
import "encoding/json"

engine.RegisterFunction("ParseJSON", func(jsonStr string) (map[string]interface{}, error) {
    var data map[string]interface{}
    err := json.Unmarshal([]byte(jsonStr), &data)
    return data, err
})

engine.RegisterFunction("ToJSON", func(data map[string]interface{}) (string, error) {
    jsonBytes, err := json.Marshal(data)
    return string(jsonBytes), nil
})
```

DWScript usage:

```pascal
var user := ParseJSON('{"name": "Alice", "age": 30}');
PrintLn('Name: ' + user['name']);
PrintLn('Age: ' + IntToStr(user['age']));

var person := record
    name: 'Bob';
    age: 25;
end;
var json := ToJSON(person);
PrintLn(json);
```

## Advanced Features

### Array Parameters

Pass DWScript arrays to Go functions:

```go
engine.RegisterFunction("SumArray", func(numbers []int64) int64 {
    var sum int64
    for _, n := range numbers {
        sum += n
    }
    return sum
})
```

```pascal
var nums := [1, 2, 3, 4, 5];
var total := SumArray(nums);
PrintLn(IntToStr(total)); // Output: 15
```

### Record/Map Parameters

Use DWScript records as map parameters:

```go
engine.RegisterFunction("ProcessConfig", func(config map[string]interface{}) string {
    debug := config["debug"].(bool)
    port := int(config["port"].(float64)) // JSON numbers come as float64

    if debug {
        return fmt.Sprintf("Debug mode enabled on port %d", port)
    }
    return fmt.Sprintf("Production mode on port %d", port)
})
```

```pascal
var config := record
    debug: true;
    port: 8080;
end;

var status := ProcessConfig(config);
PrintLn(status);
```

## Best Practices

### Error Handling in FFI

Always handle errors appropriately:

```go
// Good: Return errors for expected failures
engine.RegisterFunction("SafeDivide", func(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
})

// Good: Use DWScript exceptions for validation
engine.RegisterFunction("ValidateEmail", func(email string) error {
    if !strings.Contains(email, "@") {
        return errors.New("invalid email format")
    }
    return nil
})
```

### Type Safety

Be explicit about types in your Go functions:

```go
// Preferred: Use specific types
func ProcessData(data []int64) []string

// Avoid: Using interface{} unless necessary
func ProcessData(data interface{}) interface{}
```

### Performance

- FFI calls have some overhead due to marshaling
- For performance-critical code, consider implementing functionality directly in DWScript
- Cache expensive operations when possible

### Security

- Validate inputs in Go functions
- Be careful with file system access
- Consider sandboxing for untrusted scripts

## Limitations

### Current Limitations

1. **No variadic functions**: Use slices instead of `...T` parameters
2. **Limited type support**: Only basic types and collections are supported
3. **No callbacks**: DWScript functions cannot be passed to Go functions
4. **No method calls**: Cannot call methods on Go objects directly

### Future Enhancements

The FFI is designed to be extensible. Future versions may add:

- Support for more Go types (structs, interfaces)
- Callback functions (DWScript → Go)
- Method registration on Go objects
- Performance optimizations
- Advanced type mappings

## API Reference

### Engine.RegisterFunction(name, fn)

Registers a Go function with the DWScript engine.

**Parameters:**

- `name` (string): Function name as it will appear in DWScript
- `fn` (interface{}): Go function to register

**Returns:** `error` if registration fails

**Supported signatures:**

- `func(...) T`
- `func(...) (T, error)`
- `func(...) error`
- `func(...)`

### Type Conversion Functions

The FFI automatically handles type conversion, but you can also use these helper functions in your Go code:

- `interp.GoInt(val)` - Convert DWScript Integer to int64
- `interp.GoFloat(val)` - Convert DWScript Float to float64
- `interp.GoString(val)` - Convert DWScript String to string
- `interp.GoBool(val)` - Convert DWScript Boolean to bool

## Troubleshooting

### Common Errors

#### "unsupported Go type"

- Check that all parameter and return types are supported
- Use supported collection types instead of custom structs

#### "argument count mismatch"

- Verify the DWScript call passes the correct number of arguments
- Check for optional parameters (not yet supported)

#### "type mismatch"

- Ensure DWScript values match the expected Go types
- Use explicit type conversions in DWScript if needed

### Debugging

Enable detailed error messages:

```go
// The FFI provides detailed error messages including:
// - Parameter index where conversion failed
// - Expected vs actual types
// - Go stack traces for panics
```

For complex issues, check the DWScript execution result:

```go
result, err := engine.Eval(script)
if err != nil {
    log.Printf("Execution error: %v", err)
}
if !result.Success {
    log.Printf("Script failed: %s", result.Error)
}
```

## Migration from Other Systems

If migrating from other DWScript implementations:

1. **Delphi DWScript**: The FFI replaces Delphi's `TdwsUnit` external functions
2. **JavaScript**: Similar to Node.js `vm` module with native bindings
3. **Lua/Python**: Comparable to Lua's C API or Python's C extensions

The go-dws FFI provides a type-safe, memory-safe alternative to traditional FFI systems while maintaining the simplicity of DWScript's original design.
