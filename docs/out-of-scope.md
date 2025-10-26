# Out-of-Scope Features for go-dws

**Document Version**: 1.0
**Last Updated**: 2025-10-26
**Purpose**: Explicitly document DWScript features that will NOT be implemented in go-dws

---

## Philosophy

go-dws aims to be a **faithful port of the DWScript language** while adapting to Go's ecosystem and philosophy. However, certain features from the original DWScript are:

1. **Platform-specific** and not portable to Go
2. **Against Go's security model** (sandboxing)
3. **Not relevant** to scripting use cases
4. **Better served** by Go ecosystem libraries

This document lists such features and explains why they're excluded.

---

## Explicitly Out of Scope

### 1. COM/OLE Integration

**Feature**: COM Connector, OLE automation

**Why Out of Scope**:
- Windows-specific technology
- go-dws targets cross-platform use (Linux, macOS, Windows, WASM)
- COM relies on Windows registry and DLLs
- Modern Go applications use REST/gRPC for inter-process communication

**Alternative**:
- Use Go's FFI to call into Windows COM if absolutely needed
- Prefer platform-agnostic IPC mechanisms (HTTP, gRPC, message queues)

**Test Evidence**: `COMConnector/`, `COMConnectorFailure/`

---

### 2. Inline Assembly

**Feature**: `asm` blocks with NASM syntax

**Why Out of Scope**:
- Platform and architecture specific (x86/x64)
- Breaks portability (can't run on ARM, WASM, etc.)
- Go doesn't support inline assembly
- Modern CPUs and compilers make hand-optimized assembly rarely necessary

**Alternative**:
- Implement performance-critical code in Go
- If absolutely needed, use CGo to call assembly functions
- Rely on Go compiler optimizations (which are excellent)

**Note**: DWScript's inline assembly was an advanced feature for performance tuning. Modern Go is fast enough for most scripting use cases.

---

### 3. Direct File I/O

**Feature**: File operations (read/write/delete files, directory operations)

**Why Out of Scope**:
- **Security**: Scripts should run in a sandbox
- File I/O allows scripts to:
  - Read sensitive data (passwords, keys, config files)
  - Modify or delete critical files
  - Fill disk space (DoS attack)
  - Access files outside intended scope

**Alternative**:
- Provide **controlled** file access via Go host application
- Use virtual file systems (in-memory or restricted paths)
- Require explicit FFI calls from Go code to grant file access
- Implement capability-based security (grant specific permissions)

**Example Safe Pattern**:
```go
// Go code grants controlled file access
script.RegisterFunction("LoadConfig", func(name string) string {
    // Host validates 'name' is allowed
    if !isAllowedConfig(name) {
        return ""
    }
    return readConfigFile(name)
})
```

**Test Evidence**: `FunctionsFile/`

---

### 4. Database Connectivity

**Feature**: Direct database access (SQL queries, connections)

**Why Out of Scope**:
- Security: Direct DB access allows:
  - SQL injection attacks
  - Unauthorized data access
  - Schema manipulation
  - Connection exhaustion
- Go ecosystem has excellent database libraries
- Scripts shouldn't manage connection pools

**Alternative**:
- Provide **controlled** data access via FFI
- Expose domain-specific query methods from Go
- Use repository pattern: Go provides safe queries, scripts use results

**Example Safe Pattern**:
```go
// Go provides controlled data access
script.RegisterFunction("GetUser", func(id int) User {
    // Host validates permissions, sanitizes input
    return db.GetUserByID(id)
})
```

**Test Evidence**: `DataBaseLib/`

---

### 5. Graphics/GUI Libraries

**Feature**: 2D graphics primitives, canvas operations, UI controls

**Why Out of Scope**:
- Use case specific (not general scripting)
- Go has its own GUI libraries (fyne, gio, etc.)
- Performance: Graphics should be in native code
- Portability: Different UI frameworks per platform

**Alternative**:
- If scripting UI behavior, expose UI operations via FFI
- Let Go code handle rendering, script handles logic
- Consider declarative UI (script defines UI, Go renders)

**Test Evidence**: `GraphicsLib/`, `Model3D/`

---

### 6. RTTI Direct Connectivity

**Feature**: Direct access to Delphi's Runtime Type Information system

**Why Out of Scope**:
- Delphi-specific reflection system
- Go has its own `reflect` package with different capabilities
- RTTI was used for Delphi<->DWScript integration
- Not applicable to Go<->DWScript integration

**Alternative**:
- Use Go's `reflect` package for type introspection in FFI layer
- Provide minimal RTTI for script types (TypeOf, class name, etc.)
- Design explicit type bridges rather than automatic reflection

---

### 7. Platform-Specific Libraries

**Features**:
- System info library (CPU model, OS version, etc.)
- Windows registry access
- Platform-specific APIs

**Why Out of Scope**:
- Breaks cross-platform goal
- Information leakage security concern
- Scripts shouldn't depend on platform details

**Alternative**:
- Provide minimal, safe platform info if needed (OS type, Go version)
- Abstract platform differences in Go code
- Design portable script APIs

**Test Evidence**: `SystemInfoLib/`

---

## Conditionally Out of Scope

These features *could* be added later if there's strong demand, but are currently not prioritized:

### 8. JavaScript Filter Scripts

**Feature**: Execute scripts through JavaScript engine

**Why Deferred**:
- Complexity: Requires embedding JS engine (V8, QuickJS)
- Use case unclear: Why JS inside DWScript?
- Performance overhead

**Possible Future**: If needed for browser interop

**Test Evidence**: `JSFilterScripts/`, `JSFilterScriptsFail/`

---

### 9. HTML Filtering / Web Scraping

**Feature**: Parse and manipulate HTML/DOM

**Why Deferred**:
- Go has excellent HTML parsing libraries (`golang.org/x/net/html`)
- Scripts scraping web content is a security risk
- Better handled by Go host application

**Possible Future**: Provide controlled HTML parsing via FFI

**Test Evidence**: `HTMLFilterScripts/`, `DOMParser/`

---

### 10. Advanced Math Libraries

**Features**:
- Complex number arithmetic
- 3D vector/matrix operations
- Statistical functions
- BigInteger (arbitrary precision)

**Why Deferred**:
- Specialized use cases
- Go has `math/big`, `math/cmplx` for these
- Performance-critical, better in Go

**Possible Future**: Expose specific operations via FFI if needed

**Test Evidence**: `FunctionsMathComplex/`, `FunctionsMath3D/`, `BigInteger/`

---

### 11. Time Series / Tabular Data

**Features**: Time series operations, tabular data manipulation

**Why Deferred**:
- Domain-specific (data science)
- Go has better libraries (gonum, gota)
- Scripts manipulating large datasets is inefficient

**Possible Future**: Expose data operations via FFI

**Test Evidence**: `TimeSeriesLib/`, `TabularLib/`

---

## Features That May Be Added

These are NOT out of scope, just low priority currently:

### Delayed to Later Stages

- **Generics**: Complex feature, Stage 10+
- **Lambdas/Anonymous Methods**: Stage 9-10
- **LINQ**: Depends on lambdas, Stage 10+
- **Delegates**: Stage 9-10
- **Attributes**: Stage 10+

### Requires FFI Foundation First

- **External Functions**: High priority after Stage 8
- **Go Type Exposure**: Depends on FFI
- **JSON Support**: Can be added via FFI or native

---

## Security Model

go-dws follows a **sandbox-first** approach:

### Default Restrictions

Scripts by default CANNOT:
- Access filesystem
- Make network requests
- Execute system commands
- Access environment variables
- Fork processes
- Access raw memory/pointers

### Capabilities Model (Future)

The host Go application explicitly grants capabilities:

```go
sandbox := dwscript.NewSandbox()

// Grant specific permissions
sandbox.AllowFileRead("/data/configs/*.json")
sandbox.AllowHTTP("https://api.example.com/*")
sandbox.SetMemoryLimit(100 * 1024 * 1024) // 100MB
sandbox.SetTimeLimit(30 * time.Second)

// Script runs with only these permissions
result := sandbox.Execute(script)
```

This ensures:
- **Least privilege**: Scripts only get what they need
- **Auditability**: Permissions are explicit
- **Safety**: Malicious scripts can't escape sandbox

---

## Rationale Summary

| Feature Category | Reason Out of Scope |
|------------------|---------------------|
| COM/OLE | Windows-specific, use modern IPC |
| Inline Assembly | Platform-specific, not portable |
| File I/O | Security risk, sandbox violation |
| Database Access | Security risk, connection management |
| Graphics/GUI | Use case specific, performance |
| RTTI Connector | Delphi-specific, use Go reflect |
| Platform APIs | Portability, information leakage |

---

## How to Provide "Out of Scope" Features

If your application NEEDS a feature marked out-of-scope:

### 1. Use FFI (Foreign Function Interface)

Register Go functions accessible from scripts:

```go
interp.RegisterFunction("ReadFile", func(path string) (string, error) {
    // Validate path
    if !strings.HasPrefix(path, "/safe/dir/") {
        return "", errors.New("access denied")
    }

    data, err := os.ReadFile(path)
    return string(data), err
})
```

Script uses it:
```pascal
var content := ReadFile('/safe/dir/config.json');
```

### 2. Extend with Go Libraries

Instead of implementing in DWScript, expose Go functionality:

```go
// Instead of: script has full DB access
// Do this: script calls controlled Go functions

interp.RegisterFunction("GetUsers", func() []User {
    return db.Query("SELECT * FROM users WHERE active = true")
})
```

### 3. Virtual Abstractions

Provide abstract versions that don't touch real resources:

```go
// Virtual filesystem
vfs := memfs.New()
interp.RegisterVFS(vfs)

// Script can "write files" but only to memory
// Script: WriteFile('output.txt', data)
// Actually writes to VFS, not real disk
```

---

## Conclusion

These exclusions are **intentional design decisions** to keep go-dws:

✅ **Secure** - Sandbox by default
✅ **Portable** - Runs on any platform Go supports
✅ **Maintainable** - Focus on core language features
✅ **Flexible** - Host app provides platform-specific features via FFI

Features are only excluded when they:
- Compromise security
- Break portability
- Are better served by Go ecosystem
- Are too platform-specific

For most scripting use cases, the in-scope features + FFI provide everything needed.

---

**Document Status**: ✅ Complete - Task 8.239v finished
