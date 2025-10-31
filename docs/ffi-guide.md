# FFI Quick Start Guide

Practical examples for using the Foreign Function Interface.

## Type Mapping Quick Reference

| Go Type | DWScript Type | Example |
|---------|---------------|---------|
| `int64` | `Integer` | `func(x int64) int64` |
| `float64` | `Float` | `func(x float64) float64` |
| `string` | `String` | `func(s string) string` |
| `bool` | `Boolean` | `func(b bool) bool` |
| `[]int64` | `array of Integer` | `func(arr []int64)` |
| `[]string` | `array of String` | `func(arr []string)` |
| `map[string]T` | Record/associative array | `func() map[string]int64` |
| `error` | Exception (EHost) | `func() (T, error)` |

## Examples

### Simple Function

```go
engine.RegisterFunction("Double", func(x int64) int64 {
    return x * 2
})
```

```pascal
var result := Double(21);  // 42
```

### Error Handling

```go
engine.RegisterFunction("Divide", func(a, b int64) (int64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
})
```

```pascal
try
    var result := Divide(10, 0);
except
    on E: EHost do
        PrintLn('Error: ' + E.Message);
end;
```

### Array Processing

```go
engine.RegisterFunction("Sum", func(numbers []int64) int64 {
    sum := int64(0)
    for _, n := range numbers {
        sum += n
    }
    return sum
})
```

```pascal
var nums := [1, 2, 3, 4, 5];
var total := Sum(nums);  // 15
```

### String Manipulation

```go
engine.RegisterFunction("ToUpper", func(s string) string {
    return strings.ToUpper(s)
})

engine.RegisterFunction("Split", func(s, sep string) []string {
    return strings.Split(s, sep)
})
```

```pascal
var upper := ToUpper('hello');  // "HELLO"
var parts := Split('a,b,c', ',');  // ["a", "b", "c"]
```

### Maps/Records

```go
engine.RegisterFunction("GetConfig", func() map[string]string {
    return map[string]string{
        "host": "localhost",
        "port": "8080",
    }
})
```

```pascal
var cfg := GetConfig();
PrintLn(cfg.host);  // "localhost"
PrintLn(cfg.port);  // "8080"
```

### Panic Recovery

```go
engine.RegisterFunction("MightPanic", func(trigger bool) string {
    if trigger {
        panic("something went wrong")
    }
    return "ok"
})
```

```pascal
try
    var result := MightPanic(true);
except
    on E: EHost do
        PrintLn('Caught panic: ' + E.Message);
end;
```

### Multiple Return Values

```go
// Returns value and error
engine.RegisterFunction("ParseInt", func(s string) (int64, error) {
    val, err := strconv.ParseInt(s, 10, 64)
    return val, err
})
```

```pascal
try
    var num := ParseInt('42');
    PrintLn(IntToStr(num));
except
    on E: EHost do
        PrintLn('Parse error');
end;
```

### File I/O Example

```go
engine.RegisterFunction("ReadFile", func(path string) (string, error) {
    data, err := os.ReadFile(path)
    return string(data), err
})

engine.RegisterFunction("WriteFile", func(path, content string) error {
    return os.WriteFile(path, []byte(content), 0644)
})
```

```pascal
try
    WriteFile('test.txt', 'Hello from DWScript!');
    var content := ReadFile('test.txt');
    PrintLn(content);
except
    on E: EHost do
        PrintLn('File error: ' + E.Message);
end;
```

### HTTP Request Example

```go
engine.RegisterFunction("HttpGet", func(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    return string(body), err
})
```

```pascal
try
    var html := HttpGet('https://example.com');
    PrintLn('Got ' + IntToStr(Length(html)) + ' bytes');
except
    on E: EHost do
        PrintLn('HTTP error: ' + E.Message);
end;
```

### JSON Processing Example

```go
type Person struct {
    Name string
    Age  int64
}

engine.RegisterFunction("ParseJSON", func(jsonStr string) (map[string]interface{}, error) {
    var result map[string]interface{}
    err := json.Unmarshal([]byte(jsonStr), &result)
    return result, err
})
```

```pascal
try
    var data := ParseJSON('{"name":"Alice","age":30}');
    PrintLn(data.name);
except
    on E: EHost do
        PrintLn('JSON error');
end;
```

## Best Practices

### 1. Always Handle Errors

Return `error` as second return value:

```go
// Good
func(path string) (string, error)

// Avoid
func(path string) string  // Panics on error
```

### 2. Use Appropriate Types

```go
// Good - explicit int64
func(count int64) int64

// Avoid - Go's int (size varies by platform)
func(count int) int
```

### 3. Validate Inputs

```go
engine.RegisterFunction("SafeDivide", func(a, b int64) (int64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
})
```

### 4. Keep Functions Simple

```go
// Good - single responsibility
engine.RegisterFunction("ReadFile", func(path string) (string, error) { ... })
engine.RegisterFunction("WriteFile", func(path, content string) error { ... })

// Avoid - multiple operations
engine.RegisterFunction("ProcessFile", func(in, out string, op int) error { ... })
```

### 5. Use Meaningful Names

```go
// Good
engine.RegisterFunction("CalculateInterest", ...)
engine.RegisterFunction("ValidateEmail", ...)

// Avoid
engine.RegisterFunction("DoStuff", ...)
engine.RegisterFunction("F1", ...)
```

## Common Patterns

### Factory Function

```go
engine.RegisterFunction("NewCounter", func(start int64) map[string]interface{} {
    count := start
    return map[string]interface{}{
        "value": count,
        "increment": func() int64 {
            count++
            return count
        },
    }
})
```

### Builder Pattern

```go
engine.RegisterFunction("BuildURL", func(base, path string, params map[string]string) string {
    u := base + path
    if len(params) > 0 {
        u += "?"
        first := true
        for k, v := range params {
            if !first {
                u += "&"
            }
            u += k + "=" + v
            first = false
        }
    }
    return u
})
```

### Validation

```go
engine.RegisterFunction("ValidateEmail", func(email string) (bool, error) {
    matched, err := regexp.MatchString(`^[^@]+@[^@]+\.[^@]+$`, email)
    return matched, err
})
```

## Error Messages

Errors are automatically converted to `EHost` exceptions:

```go
errors.New("file not found")          // → EHost: file not found
fmt.Errorf("invalid value: %d", x)   // → EHost: invalid value: 42
```

Panics are also caught and converted:

```go
panic("unexpected state")             // → EHost: panic: unexpected state
panic(errors.New("critical error"))  // → EHost: panic: critical error
panic(42)                             // → EHost: panic: 42
```

## Performance Tips

1. **Avoid frequent small calls**: Batch operations when possible
2. **Reuse connections**: Don't create new HTTP clients for each request
3. **Use buffering**: For file I/O and network operations
4. **Cache results**: Store computed values in Go when appropriate

## Limitations

- No support for Go channels (use callbacks or polling)
- No support for Go interfaces (use concrete types)
- No support for Go generics (use `interface{}` or specific types)
- Variadic functions not directly supported (use slices)

## See Also

- [ffi.md](ffi.md) - Complete FFI documentation
- [examples/ffi/](../examples/ffi/) - Working examples
- [pkg/dwscript/ffi.go](../pkg/dwscript/ffi.go) - API source
