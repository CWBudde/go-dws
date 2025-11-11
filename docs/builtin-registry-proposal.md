# Builtin Function Registry Patterns

This document explores different patterns for registering builtin functions to improve code organization.

## Current State

- **507-line switch statement** in `functions_builtins.go`
- Builtins are methods on `*Interpreter`: `func (i *Interpreter) builtinXxx(args []Value) Value`
- All builtins in same package as interpreter (no circular import issues)
- 150+ builtin functions across 14 files

## Option 1: Map-Based Registry (In-Package)

**Keep everything in `interp` package** but use a map instead of switch statement.

### Implementation

```go
// internal/interp/builtin_registry.go
package interp

// BuiltinFunc is the signature for all builtin functions
type BuiltinFunc func(*Interpreter, []Value) Value

// builtinRegistry holds all registered builtin functions
var builtinRegistry = make(map[string]BuiltinFunc)

// registerBuiltin registers a builtin function (called from init)
func registerBuiltin(name string, fn BuiltinFunc) {
    builtinRegistry[normalizeBuiltinName(name)] = fn
}

// callBuiltin dispatches using the registry
func (i *Interpreter) callBuiltin(name string, args []Value) Value {
    // Check external functions first
    if i.externalFunctions != nil {
        if extFunc, ok := i.externalFunctions.Get(name); ok {
            return i.callExternalFunction(extFunc, args)
        }
    }

    name = normalizeBuiltinName(name)
    if fn, ok := builtinRegistry[name]; ok {
        return fn(i, args)
    }

    return i.newError("unknown function: %s", name)
}
```

### Registration in Builtin Files

Each builtin file registers its functions in `init()`:

```go
// internal/interp/builtins_core.go
package interp

func init() {
    registerBuiltin("PrintLn", builtinPrintLn)
    registerBuiltin("Print", builtinPrint)
    registerBuiltin("Ord", builtinOrd)
    registerBuiltin("Chr", builtinChr)
}

// Convert from method to function
func builtinPrintLn(i *Interpreter, args []Value) Value {
    if len(args) == 0 {
        fmt.Fprintln(i.output)
        return nil
    }
    // ... rest of implementation
}
```

### Pros
- ✅ **Eliminates 507-line switch** - much cleaner dispatch
- ✅ **Self-registering** - each file registers its own functions
- ✅ **No circular imports** - everything stays in `interp` package
- ✅ **Easy migration** - convert methods to functions gradually
- ✅ **Extensible** - external packages can add builtins too

### Cons
- ⚠️ Must convert all methods to functions (pass `*Interpreter` explicitly)
- ⚠️ All files must still be in same package (can't separate to subdirectories)

---

## Option 2: Separate Builtins Package (Sibling)

Create `internal/builtins/` as a **sibling** to `internal/interp/`, not a child.

### Package Structure

```
internal/
├── values/              # NEW: Shared types
│   ├── value.go         # Value interface and implementations
│   ├── interpreter.go   # Interpreter interface
│   └── environment.go   # Environment types
├── interp/              # Interpreter implementation
│   ├── interpreter.go   # Implements values.Interpreter interface
│   └── ...
└── builtins/            # Builtin functions (separate package)
    ├── registry.go
    ├── core.go
    ├── math.go
    └── strings.go
```

### Dependency Flow (No Circles!)

```
internal/values/        # Foundation types
    ↑
    ├─ imported by → internal/builtins/
    └─ imported by → internal/interp/
```

### Implementation

```go
// internal/values/value.go
package values

// Value is the interface for all DWScript runtime values
type Value interface {
    Type() string
    String() string
}

// Interpreter is the interface that builtins need
type Interpreter interface {
    Output() io.Writer
    NewError(format string, args ...interface{}) Value
    EvalExpression(expr ast.Expression) Value
    // ... other methods builtins need
}
```

```go
// internal/builtins/registry.go
package builtins

import "github.com/cwbudde/go-dws/internal/values"

type BuiltinFunc func(values.Interpreter, []values.Value) values.Value

var registry = make(map[string]BuiltinFunc)

func Register(name string, fn BuiltinFunc) {
    registry[name] = fn
}

func Get(name string) (BuiltinFunc, bool) {
    fn, ok := registry[name]
    return fn, ok
}
```

```go
// internal/builtins/core.go
package builtins

import (
    "fmt"
    "github.com/cwbudde/go-dws/internal/values"
)

func init() {
    Register("PrintLn", PrintLn)
    Register("Print", Print)
}

func PrintLn(interp values.Interpreter, args []values.Value) values.Value {
    if len(args) == 0 {
        fmt.Fprintln(interp.Output())
        return nil
    }
    // ... implementation
}
```

```go
// internal/interp/interpreter.go
package interp

import (
    "github.com/cwbudde/go-dws/internal/values"
    "github.com/cwbudde/go-dws/internal/builtins"
)

type Interpreter struct {
    // ... fields
}

// Implement values.Interpreter interface
func (i *Interpreter) Output() io.Writer { return i.output }
func (i *Interpreter) NewError(format string, args ...interface{}) Value {
    // ... implementation
}

func (i *Interpreter) callBuiltin(name string, args []Value) Value {
    if fn, ok := builtins.Get(name); ok {
        // Convert []Value to []values.Value (may be same type)
        return fn(i, args)
    }
    return i.newError("unknown function: %s", name)
}
```

### Pros
- ✅ **Clean separation** - builtins in their own package
- ✅ **No circular imports** - uses intermediate `values` package
- ✅ **Proper organization** - subdirectories for logical grouping
- ✅ **Testable** - can test builtins independently
- ✅ **Extensible** - easy to add new builtin categories

### Cons
- ⚠️ **Major refactoring** - must extract Value types to common package
- ⚠️ **Interface overhead** - Interpreter becomes an interface
- ⚠️ **More complex** - three packages instead of one
- ⚠️ **Migration cost** - affects entire codebase

---

## Option 3: Plugin-Style with Categories

Keep in same package but organize into categories with sub-registries.

### Implementation

```go
// internal/interp/builtin_registry.go
package interp

type BuiltinFunc func(*Interpreter, []Value) Value

type BuiltinCategory struct {
    name     string
    builtins map[string]BuiltinFunc
}

var categories = make([]*BuiltinCategory, 0)

func NewCategory(name string) *BuiltinCategory {
    cat := &BuiltinCategory{
        name:     name,
        builtins: make(map[string]BuiltinFunc),
    }
    categories = append(categories, cat)
    return cat
}

func (c *BuiltinCategory) Register(name string, fn BuiltinFunc) {
    c.builtins[normalizeBuiltinName(name)] = fn
}

func (i *Interpreter) callBuiltin(name string, args []Value) Value {
    name = normalizeBuiltinName(name)

    // Search all categories
    for _, cat := range categories {
        if fn, ok := cat.builtins[name]; ok {
            return fn(i, args)
        }
    }

    return i.newError("unknown function: %s", name)
}
```

```go
// internal/interp/builtins_math.go
package interp

var mathCategory = NewCategory("math")

func init() {
    mathCategory.Register("Abs", builtinAbs)
    mathCategory.Register("Sin", builtinSin)
    mathCategory.Register("Cos", builtinCos)
    // ...
}
```

### Pros
- ✅ **Logical grouping** - categories make organization clear
- ✅ **No circular imports** - stays in same package
- ✅ **Easy to list** - can enumerate all builtins by category
- ✅ **Incremental** - can migrate one category at a time

### Cons
- ⚠️ Still in flat package structure
- ⚠️ Slight overhead searching categories

---

## Recommendation

For **immediate improvement** with minimal disruption:

**Start with Option 1 (Map-Based Registry)**:
1. Create `builtin_registry.go` with map-based dispatch
2. Convert one builtin file at a time (e.g., start with `builtins_core.go`)
3. Use `init()` functions for self-registration
4. Delete switch statement once all converted

This gives you:
- Cleaner code (no 507-line switch)
- Self-documenting (each file registers its functions)
- No architectural changes
- Can be done incrementally

**For future** (if you want true separation):

**Option 2** is the proper architectural solution, but requires:
- Creating `internal/values/` package
- Extracting Value types and Interpreter interface
- Updating all imports across codebase
- Much larger effort, but results in cleanest architecture

---

## Example Migration (Option 1)

Here's how to migrate one file as a proof of concept:

### Before: `builtins_core.go`
```go
func (i *Interpreter) builtinPrintLn(args []Value) Value {
    // implementation
}
```

### After: `builtins_core.go`
```go
func init() {
    registerBuiltin("PrintLn", builtinPrintLn)
}

func builtinPrintLn(i *Interpreter, args []Value) Value {
    // same implementation, but now a function not a method
}
```

### Update dispatcher: `functions_builtins.go`
```go
func (i *Interpreter) callBuiltin(name string, args []Value) Value {
    if fn, ok := builtinRegistry[normalizeBuiltinName(name)]; ok {
        return fn(i, args)
    }
    // fallback to old switch for unmigrated functions
    return i.callBuiltinLegacy(name, args)
}
```

This allows **incremental migration** without breaking anything!
