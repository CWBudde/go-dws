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

### Supported Pointer Types (Var Parameters)

| Go Type | DWScript Usage | Notes |
|---------|----------------|-------|
| `*int64`, `*int`, etc. | `var Integer` | By-reference integer parameter |
| `*float64`, `*float32` | `var Float` | By-reference float parameter |
| `*string` | `var String` | By-reference string parameter |
| `*bool` | `var Boolean` | By-reference boolean parameter |

See [Var Parameters](#var-parameters-by-reference) for details.

### Unsupported Types

The following Go types are not currently supported:
- Complex numbers (`complex64`, `complex128`)
- Channels (`chan T`)
- Functions (`func`)
- Interfaces (except `error`)
- Pointers to unsupported types (e.g., `*struct`)
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

### Panic Handling

Go panics are automatically caught and converted to `EHost` exceptions. This ensures your DWScript code never crashes due to panics in registered Go functions.

**What happens when a Go function panics:**

1. The panic is caught by a `defer/recover()` block in the FFI wrapper
2. The panic value is converted to a string message with "panic: " prefix
3. An `EHost` exception is created with the panic details
4. The Go stack trace is included in the exception message (for debugging)
5. The exception is raised in DWScript and can be caught with try/except

**Example - Panic in Go function:**

```go
engine.RegisterFunction("DivideByZero", func(n int64) int64 {
    divisor := 0
    return n / int64(divisor)  // This will panic
})
```

DWScript code can catch the panic as an exception:

```pascal
try
    var result := DivideByZero(42);
    PrintLn('Should not reach here');
except
    on E: EHost do begin
        PrintLn('Caught panic: ' + E.Message);
        // E.Message will contain "panic: runtime error: integer divide by zero"
        // plus the Go stack trace for debugging
    end;
end;
```

**Panic types handled:**

- `panic(error)` - Error types are converted using their `.Error()` method
- `panic("string")` - Strings are used directly
- `panic(42)` - Other types are converted using `fmt.Sprintf("%v", value)`

**Best Practices:**

1. **Write defensive Go code** - While panics are caught, it's better to return errors explicitly:
   ```go
   // Good: Return error explicitly
   engine.RegisterFunction("SafeDivide", func(a, b int64) (int64, error) {
       if b == 0 {
           return 0, errors.New("division by zero")
       }
       return a / b, nil
   })

   // Avoid: Letting panic occur (though it will be caught)
   engine.RegisterFunction("UnsafeDivide", func(a, b int64) int64 {
       return a / b  // panics if b == 0
   })
   ```

2. **Test edge cases** - Ensure your Go functions handle invalid inputs gracefully

3. **Use error returns** - Prefer `(result, error)` signature over panics for expected error conditions

4. **Include context** - Return descriptive error messages that help debug issues:
   ```go
   if err != nil {
       return "", fmt.Errorf("failed to read file %s: %w", filename, err)
   }
   ```

**Stack Traces:**

When a panic occurs, the exception message includes:
- The panic message prefixed with "panic: "
- The full Go stack trace (up to 2048 bytes)
- The DWScript call stack (accessible via exception handler)

This makes debugging panic-related issues straightforward even in complex FFI scenarios.

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

### Var Parameters (By-Reference)

**Task 9.2d**: Go functions with pointer parameters are automatically treated as `var` parameters in DWScript, enabling modification of caller variables.

#### Basic Usage

```go
// Register Go function with pointer parameter
engine.RegisterFunction("Increment", func(x *int64) {
    *x++
})

// Register swap function with two pointers
engine.RegisterFunction("Swap", func(a, b *int64) {
    temp := *a
    *a = *b
    *b = temp
})
```

DWScript usage:

```pascal
var n: Integer := 5;
Increment(n);
PrintLn(IntToStr(n)); // Output: 6

var x: Integer := 10;
var y: Integer := 20;
Swap(x, y);
PrintLn(IntToStr(x)); // Output: 20
PrintLn(IntToStr(y)); // Output: 10
```

#### Supported Pointer Types

All basic types can be used as var parameters:

| Go Pointer Type | DWScript Type | Example |
|-----------------|---------------|---------|
| `*int64` | `var Integer` | `func Increment(x *int64)` |
| `*float64` | `var Float` | `func Double(x *float64)` |
| `*string` | `var String` | `func MakeUpperCase(s *string)` |
| `*bool` | `var Boolean` | `func Toggle(b *bool)` |

#### Mixed Parameters

Functions can mix regular value parameters with var parameters:

```go
engine.RegisterFunction("AddAndStore", func(result *int64, a, b int64) {
    *result = a + b
})
```

```pascal
var sum: Integer := 0;
AddAndStore(sum, 15, 27);
PrintLn(IntToStr(sum)); // Output: 42
```

#### Requirements and Validation

1. **Var parameters must be variables** - Literals and expressions cannot be passed:
   ```pascal
   // Valid
   var n: Integer := 5;
   Increment(n);  // OK

   // Invalid - will raise compile error
   Increment(42);  // Error: var parameter requires a variable
   ```

2. **Type matching** - The variable type must match the pointer element type:
   ```pascal
   var x: Integer := 5;
   var y: Float := 3.14;

   Increment(x);  // OK: *int64 matches Integer
   Increment(y);  // Error: type mismatch
   ```

3. **Semantic validation** - The compiler checks var parameter usage at compile time

#### Implementation Details

When a Go function parameter is a pointer type:

1. The FFI automatically marks it as a `var` parameter
2. DWScript creates a `ReferenceValue` wrapper around the variable
3. The value is marshaled to a Go pointer before the call
4. After the call, the modified value is unmarshaled back to DWScript
5. The original variable is updated with the new value

This ensures that modifications made by the Go function are reflected in the DWScript variable, just like native DWScript var parameters.

#### Best Practices for Var Parameters

1. **Use pointer semantics consistently**:
   ```go
   // Good: Clear that x will be modified
   func Increment(x *int64) {
       *x++
   }

   // Avoid: Returning modified value defeats purpose of var parameter
   func Increment(x *int64) int64 {
       *x++
       return *x  // Unnecessary
   }
   ```

2. **Document side effects**:
   ```go
   // IncrementCounter increases the counter by 1 and returns the new value.
   // The counter parameter is modified in place.
   func IncrementCounter(counter *int64) int64 {
       *counter++
       return *counter
   }
   ```

3. **Prefer var parameters for output values**:
   ```go
   // Good: Clear output parameter
   func ParseCoordinates(input string, x, y *float64) error {
       // Parse and set *x and *y
       return nil
   }

   // Alternative: Return values (also valid)
   func ParseCoordinates(input string) (x, y float64, err error) {
       // Parse and return
       return x, y, nil
   }
   ```

4. **Null safety**: The FFI ensures pointers are never nil when calling from DWScript, as all var parameters must reference valid variables.

### Registering Go Methods

**Task 9.3**: The FFI supports registering methods from Go structs, making them callable from DWScript. This allows you to expose stateful Go objects to your scripts.

#### Two Registration Approaches

**Approach 1: Method Values (Recommended)**

The simplest way to register methods is to use Go's method value syntax with the existing `RegisterFunction` API:

```go
type Counter struct {
    value int64
}

func (c *Counter) Increment() {
    c.value++
}

func (c *Counter) Add(x int64) {
    c.value += x
}

func (c *Counter) GetValue() int64 {
    return c.value
}

// Create instance and register methods using method values
counter := &Counter{value: 0}
engine.RegisterFunction("Increment", counter.Increment)
engine.RegisterFunction("Add", counter.Add)
engine.RegisterFunction("GetValue", counter.GetValue)
```

DWScript usage:

```pascal
Increment();
Add(5);
var result := GetValue();
PrintLn(IntToStr(result)); // Output: 6
```

**Approach 2: RegisterMethod API (More Explicit)**

For more explicit registration with validation, use the `RegisterMethod` API:

```go
counter := &Counter{value: 0}
engine.RegisterMethod("Increment", counter, "Increment")
engine.RegisterMethod("Add", counter, "Add")
engine.RegisterMethod("GetValue", counter, "GetValue")
```

Both approaches work identically; `RegisterMethod` adds validation that the method exists on the receiver type.

#### How It Works

When you register a method:

1. **Method values automatically bind the receiver**: Go's method value mechanism (`obj.Method`) creates a closure that captures the receiver
2. **The receiver is preserved**: Each time DWScript calls the function, it operates on the same Go object
3. **State is maintained**: Changes to the receiver's fields persist across calls
4. **No special handling needed**: The FFI treats method values like regular functions

#### Pointer vs Value Receivers

The choice between pointer and value receivers affects whether modifications persist:

**Pointer Receivers** (Most Common):

```go
type Calculator struct {
    result float64
}

// Pointer receiver (*Calculator) - can modify state
func (c *Calculator) Add(x float64) {
    c.result += x  // Modifies the original Calculator
}

func (c *Calculator) GetResult() float64 {
    return c.result
}

calc := &Calculator{result: 10.0}
engine.RegisterMethod("Add", calc, "Add")
engine.RegisterMethod("GetResult", calc, "GetResult")
```

```pascal
Add(5.0);
Add(3.0);
var result := GetResult();
PrintLn(FloatToStr(result)); // Output: 18.0
```

**Value Receivers** (Read-Only):

```go
type Point struct {
    x, y float64
}

// Value receiver (Point) - operates on a copy
func (p Point) Distance() float64 {
    return math.Sqrt(p.x*p.x + p.y*p.y)
}

point := Point{x: 3.0, y: 4.0}
engine.RegisterMethod("Distance", point, "Distance")
```

```pascal
var dist := Distance();
PrintLn(FloatToStr(dist)); // Output: 5.0
```

**Rule of Thumb**: Use pointer receivers (`*T`) when methods need to modify state; use value receivers (`T`) for read-only operations.

#### Multiple Instances

You can register methods from multiple instances to provide separate state:

```go
counter1 := &Counter{value: 0}
counter2 := &Counter{value: 100}

// Register with different names to distinguish instances
engine.RegisterMethod("Counter1Add", counter1, "Add")
engine.RegisterMethod("Counter1Get", counter1, "GetValue")

engine.RegisterMethod("Counter2Add", counter2, "Add")
engine.RegisterMethod("Counter2Get", counter2, "GetValue")
```

```pascal
Counter1Add(5);   // counter1 = 5
Counter2Add(10);  // counter2 = 110

PrintLn(IntToStr(Counter1Get())); // Output: 5
PrintLn(IntToStr(Counter2Get())); // Output: 110
```

#### Complex Example: Calculator with Multiple Operations

```go
type Calculator struct {
    result float64
    ops    int64
}

func (c *Calculator) Add(x float64) {
    c.result += x
    c.ops++
}

func (c *Calculator) Multiply(x float64) {
    c.result *= x
    c.ops++
}

func (c *Calculator) GetResult() float64 {
    return c.result
}

func (c *Calculator) GetOpsCount() int64 {
    return c.ops
}

func (c *Calculator) Reset() {
    c.result = 0.0
    c.ops = 0
}

calc := &Calculator{result: 10.0}
engine.RegisterMethod("Add", calc, "Add")
engine.RegisterMethod("Multiply", calc, "Multiply")
engine.RegisterMethod("GetResult", calc, "GetResult")
engine.RegisterMethod("GetOpsCount", calc, "GetOpsCount")
engine.RegisterMethod("Reset", calc, "Reset")
```

```pascal
Add(5.0);      // 15.0
Multiply(2.0); // 30.0
Add(10.0);     // 40.0

var result := GetResult();
var ops := GetOpsCount();
PrintLn(FloatToStr(result)); // Output: 40.0
PrintLn(IntToStr(ops));      // Output: 3

Reset();
Add(100.0);
PrintLn(FloatToStr(GetResult())); // Output: 100.0
```

#### Method Registration Validation

`RegisterMethod` validates registration and provides helpful error messages:

```go
counter := &Counter{value: 0}

// Error: nil receiver
engine.RegisterMethod("Test", nil, "Increment")
// Returns: "cannot register method on nil receiver"

// Error: empty method name
engine.RegisterMethod("Test", counter, "")
// Returns: "method name cannot be empty"

// Error: method doesn't exist
engine.RegisterMethod("Test", counter, "NonExistent")
// Returns: "method NonExistent not found on type *Counter"

// Error: unexported method (lowercase)
engine.RegisterMethod("Test", counter, "privateMethod")
// Returns: "method privateMethod not found on type *Counter"
```

#### When to Use Method Registration

Use method registration when you need:

1. **Stateful operations**: Object maintains state across multiple function calls
2. **Encapsulation**: Group related functionality in a Go struct
3. **Resource management**: Handle files, database connections, or other resources
4. **Object-oriented patterns**: Expose Go objects to DWScript in an OO style

Example use cases:
- Database connection pools
- HTTP clients with configuration
- File handlers with buffering
- Game entities with state
- Configuration managers

#### Limitations

1. **No automatic instantiation**: DWScript cannot create new instances of Go types; you must create and register instances from Go
2. **No method discovery**: DWScript cannot list available methods; you must document what's registered
3. **Name conflicts**: Be careful not to override built-in DWScript functions (e.g., don't use "Inc" as it conflicts with the built-in `Inc()`)
4. **Single receiver per registration**: Each registered method binds to one specific instance; for multiple instances, register methods with different names

#### Best Practices

1. **Use descriptive names** to avoid conflicts with built-ins:
   ```go
   // Good: Clear, specific names
   engine.RegisterMethod("CounterIncrement", counter, "Increment")
   engine.RegisterMethod("CounterGet", counter, "GetValue")

   // Avoid: Generic names that might conflict
   engine.RegisterMethod("Inc", counter, "Increment") // Conflicts with built-in Inc()
   ```

2. **Document receiver state** in method comments:
   ```go
   // Add increases the calculator's result by x and increments the operation counter.
   func (c *Calculator) Add(x float64) {
       c.result += x
       c.ops++
   }
   ```

3. **Use pointer receivers for stateful methods**:
   ```go
   // Good: Pointer receiver for state modification
   func (c *Counter) Increment() { c.value++ }

   // Bad: Value receiver doesn't persist changes
   func (c Counter) Increment() { c.value++ } // Changes lost!
   ```

4. **Consider error handling** for methods that can fail:
   ```go
   func (db *Database) Query(sql string) ([]map[string]interface{}, error) {
       // Return errors for invalid SQL, connection failures, etc.
       return results, nil
   }
   ```

5. **Initialize instances before registration**:
   ```go
   // Good: Fully initialized
   db := &Database{connectionString: "..."}
   if err := db.Connect(); err != nil {
       log.Fatal(err)
   }
   engine.RegisterMethod("Query", db, "Query")

   // Bad: Uninitialized, might panic
   db := &Database{}
   engine.RegisterMethod("Query", db, "Query") // db.Query() will fail
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

## Performance Considerations

The FFI has been benchmarked to ensure reasonable performance characteristics. See `pkg/dwscript/ffi_bench_test.go` for detailed benchmarks.

### Key Performance Metrics

- **FFI call overhead**: ~20-25µs per call (vs ~30µs for native DWScript function)
- **Primitive marshaling**: <1µs additional overhead per parameter
- **Array marshaling**: Linear with array size (~0.5µs per element)
- **Callback overhead**: ~45-50µs per callback invocation (includes round-trip)
- **Var parameter overhead**: Minimal (~1-2µs for copy-in/copy-out)

### Optimization Tips

1. **Batch operations**: Instead of multiple small FFI calls, pass arrays and process in batch
2. **Minimize callbacks**: Callback overhead is higher than direct FFI calls
3. **Reuse engines**: Engine creation is expensive; reuse when possible
4. **Avoid deep callback nesting**: Each callback level adds overhead

### Example: Optimizing Array Processing

```go
// ❌ Slow: Multiple FFI calls
engine.RegisterFunction("ProcessOne", func(x int64) int64 { return x * 2 })
// DWScript: for i := 0 to High(arr) do arr[i] := ProcessOne(arr[i]);

// ✅ Fast: Single FFI call with array
engine.RegisterFunction("ProcessAll", func(arr []int64) []int64 {
    result := make([]int64, len(arr))
    for i, x := range arr {
        result[i] = x * 2
    }
    return result
})
// DWScript: arr := ProcessAll(arr);
```

## Migration from Other Systems

If migrating from other DWScript implementations:

1. **Delphi DWScript**: The FFI replaces Delphi's `TdwsUnit` external functions
2. **JavaScript**: Similar to Node.js `vm` module with native bindings
3. **Lua/Python**: Comparable to Lua's C API or Python's C extensions

The go-dws FFI provides a type-safe, memory-safe alternative to traditional FFI systems while maintaining the simplicity of DWScript's original design.
