# Case-Insensitive Identifier Migration Guide

This guide explains how to migrate code from using direct `strings.ToLower()` and `strings.EqualFold()` calls to the centralized `pkg/ident` package for case-insensitive identifier handling.

## Why Centralize?

DWScript is a case-insensitive language. Identifiers like `MyVariable`, `myvariable`, and `MYVARIABLE` all refer to the same entity. Centralizing this handling:

1. **Ensures consistency** - All identifier operations use the same normalization
2. **Prevents bugs** - No accidental case-sensitive comparisons
3. **Preserves original casing** - Error messages show user's original input
4. **Simplifies maintenance** - One place to change if normalization rules evolve

## Quick Reference

| Before | After |
|--------|-------|
| `strings.ToLower(name)` | `ident.Normalize(name)` |
| `strings.EqualFold(a, b)` | `ident.Equal(a, b)` |
| `strings.ToLower(a) == strings.ToLower(b)` | `ident.Equal(a, b)` |
| `strings.HasPrefix(strings.ToLower(s), "prefix")` | `ident.HasPrefix(s, "prefix")` |
| `strings.HasSuffix(strings.ToLower(s), "suffix")` | `ident.HasSuffix(s, "suffix")` |

## Common Patterns

### Pattern 1: Map Keys

When storing identifiers in maps for lookup:

```go
// BEFORE: Direct strings.ToLower()
symbols := make(map[string]*Symbol)
symbols[strings.ToLower(name)] = sym

// Lookup
sym, ok := symbols[strings.ToLower(lookupName)]
```

```go
// AFTER: Use ident.Normalize()
symbols := make(map[string]*Symbol)
symbols[ident.Normalize(name)] = sym

// Lookup
sym, ok := symbols[ident.Normalize(lookupName)]
```

### Pattern 2: One-off Comparisons

When checking if two identifiers are equal:

```go
// BEFORE: Using EqualFold
if strings.EqualFold(methodName, "Create") {
    // Handle constructor
}

// BEFORE: Using ToLower twice (wasteful)
if strings.ToLower(a) == strings.ToLower(b) {
    // Names match
}
```

```go
// AFTER: Use ident.Equal()
if ident.Equal(methodName, "Create") {
    // Handle constructor
}

if ident.Equal(a, b) {
    // Names match
}
```

### Pattern 3: Prefix/Suffix Matching

When checking if an identifier starts or ends with a specific string:

```go
// BEFORE: Allocates temporary string
if strings.HasPrefix(strings.ToLower(typeName), "array") {
    // Handle array type
}
```

```go
// AFTER: No allocation
if ident.HasPrefix(typeName, "array") {
    // Handle array type
}
```

### Pattern 4: Symbol Tables with Original Casing

When you need to preserve the original casing for error messages:

```go
// BEFORE: Often loses original casing
type SymbolTable struct {
    symbols map[string]*Symbol  // lowercase keys
}

func (st *SymbolTable) Define(name string, sym *Symbol) {
    st.symbols[strings.ToLower(name)] = sym
    // Original casing lost!
}

func (st *SymbolTable) Error(name string) error {
    // Error shows normalized name, not user's input
    return fmt.Errorf("undefined: %s", strings.ToLower(name))
}
```

```go
// AFTER: Preserves original casing
type SymbolTable struct {
    symbols   map[string]*Symbol  // normalized keys
    originals map[string]string   // normalized -> original
}

func (st *SymbolTable) Define(name string, sym *Symbol) {
    normalized := ident.Normalize(name)
    st.symbols[normalized] = sym
    st.originals[normalized] = name  // Preserve original!
}

func (st *SymbolTable) Lookup(name string) (*Symbol, bool) {
    sym, ok := st.symbols[ident.Normalize(name)]
    return sym, ok
}

func (st *SymbolTable) GetOriginalName(name string) string {
    return st.originals[ident.Normalize(name)]
}

func (st *SymbolTable) Error(name string) error {
    // Error shows user's original input
    return fmt.Errorf("undefined: %s", name)
}
```

### Pattern 5: Keyword Checking

When checking if an identifier is a reserved keyword:

```go
// BEFORE: Manual loop or multiple comparisons
keywords := []string{"begin", "end", "if", "then"}
found := false
for _, kw := range keywords {
    if strings.EqualFold(name, kw) {
        found = true
        break
    }
}
```

```go
// AFTER: Use Contains or IsKeyword
if ident.Contains(keywords, name) {
    // name is a keyword
}

// Or with inline keywords
if ident.IsKeyword(name, "begin", "end", "if", "then") {
    // name is a keyword
}
```

### Pattern 6: Sorting Identifiers

When sorting identifiers case-insensitively:

```go
// BEFORE: Manual comparison
sort.Slice(names, func(i, j int) bool {
    return strings.ToLower(names[i]) < strings.ToLower(names[j])
})
```

```go
// AFTER: Use ident.Compare
sort.Slice(names, func(i, j int) bool {
    return ident.Compare(names[i], names[j]) < 0
})
```

## Preserving Original Casing in Error Messages

**Critical Rule**: Error messages should always show the user's original casing, not the normalized form.

### Bad Example

```go
func (a *Analyzer) checkVariable(name string) error {
    normalized := ident.Normalize(name)
    if _, ok := a.symbols[normalized]; !ok {
        // BAD: Shows normalized name
        return fmt.Errorf("undefined variable '%s'", normalized)
    }
    return nil
}
```

### Good Example

```go
func (a *Analyzer) checkVariable(name string) error {
    if _, ok := a.symbols[ident.Normalize(name)]; !ok {
        // GOOD: Shows original name from user input
        return fmt.Errorf("undefined variable '%s'", name)
    }
    return nil
}
```

### When Looking Up Stored Names

If you need the originally-defined name (not the lookup name):

```go
func (a *Analyzer) suggestSimilar(name string) string {
    normalized := ident.Normalize(name)
    if original, ok := a.originalNames[normalized]; ok {
        // Return the originally-defined casing
        return original
    }
    return name
}
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Normalizing for Display

```go
// BAD: Don't normalize for display
fmt.Printf("Variable: %s\n", ident.Normalize(varName))

// GOOD: Keep original casing for display
fmt.Printf("Variable: %s\n", varName)
```

### Anti-Pattern 2: Double Normalization

```go
// BAD: Normalizing already-normalized value
key := ident.Normalize(name)
store[ident.Normalize(key)] = value  // Redundant!

// GOOD: Normalize once
key := ident.Normalize(name)
store[key] = value
```

### Anti-Pattern 3: Mixing Approaches

```go
// BAD: Mixing ident and strings functions
if ident.Equal(a, b) && strings.ToLower(c) == "value" {
    // Inconsistent!
}

// GOOD: Use ident consistently for identifiers
if ident.Equal(a, b) && ident.Equal(c, "value") {
    // Consistent!
}
```

### Anti-Pattern 4: Using ident for Non-Identifiers

```go
// BAD: Using ident for file extensions
if ident.Equal(filepath.Ext(filename), ".dws") {
    // Don't use ident for non-identifier strings
}

// GOOD: Use strings functions for non-identifiers
if strings.EqualFold(filepath.Ext(filename), ".dws") {
    // Appropriate for file extensions
}
```

## When NOT to Use pkg/ident

The `pkg/ident` package is specifically for DWScript identifier handling. Do NOT use it for:

- **File extensions or paths** - Use `filepath` package and `strings.ToLower` directly
- **URLs or network addresses** - Use `url` package
- **User-facing text** that isn't a programming identifier
- **Non-ASCII text** requiring locale-aware case folding

## Function Reference

| Function | Purpose | Allocates? |
|----------|---------|------------|
| `Normalize(s)` | Convert to canonical form for map keys | Yes (if not lowercase) |
| `Equal(a, b)` | Case-insensitive equality check | No |
| `Compare(a, b)` | Case-insensitive comparison for sorting | Yes (both strings) |
| `Contains(slice, s)` | Check if slice contains s | No |
| `Index(slice, s)` | Find index of s in slice | No |
| `IsKeyword(s, kw...)` | Check if s matches any keyword | No |
| `HasPrefix(s, prefix)` | Case-insensitive prefix check | No |
| `HasSuffix(s, suffix)` | Case-insensitive suffix check | No |

## Migration Checklist

When migrating a file:

1. [ ] Add import: `"github.com/cwbudde/go-dws/pkg/ident"`
2. [ ] Replace `strings.ToLower(name)` with `ident.Normalize(name)` for map keys
3. [ ] Replace `strings.EqualFold(a, b)` with `ident.Equal(a, b)` for comparisons
4. [ ] Replace `strings.ToLower(a) == strings.ToLower(b)` with `ident.Equal(a, b)`
5. [ ] Replace prefix/suffix checks with `ident.HasPrefix`/`ident.HasSuffix`
6. [ ] Verify error messages preserve original casing
7. [ ] Remove unused `strings` import if applicable
8. [ ] Run tests: `go test ./...`

## Examples from the Codebase

### Environment (internal/interp/environment.go)

```go
// Already migrated - good example
func (e *Environment) Get(name string) (Object, bool) {
    if obj, ok := e.store[ident.Normalize(name)]; ok {
        return obj, true
    }
    if e.outer != nil {
        return e.outer.Get(name)
    }
    return nil, false
}
```

### Units Registry (internal/units/registry.go)

```go
// Already migrated - good example
func (r *UnitRegistry) RegisterUnit(name string, unit *Unit) {
    normalized := ident.Normalize(name)
    r.units[normalized] = unit
}

func (r *UnitRegistry) GetUnit(name string) (*Unit, bool) {
    normalized := ident.Normalize(name)
    unit, ok := r.units[normalized]
    return unit, ok
}
```

## Related Documentation

- `pkg/ident/doc.go` - Package documentation with all patterns
- `NORMALIZE.md` - Full migration plan and task tracking
- `CONTRIBUTING.md` - Contribution guidelines (case insensitivity section)
