// Package ident provides utilities for case-insensitive identifier handling in DWScript.
//
// # Overview
//
// DWScript is a case-insensitive language where identifiers like "MyVariable",
// "myvariable", and "MYVARIABLE" all refer to the same entity. This package
// centralizes the normalization and comparison logic to ensure consistency
// across the codebase.
//
// # Design Principles
//
// 1. Normalize Once, Store Normalized: When storing identifiers as map keys,
// normalize them once using Normalize() and store the result.
//
// 2. Compare Efficiently: For one-off comparisons, use Equal() which is more
// efficient than normalizing both strings.
//
// 3. Preserve Original Case: Always keep the original case for error messages
// and display. Normalize only for comparison and storage.
//
// 4. Consistency: Use these helpers throughout the codebase instead of
// direct strings.ToLower() or strings.EqualFold() calls.
//
// # Usage Patterns
//
// ## Pattern 1: Symbol Table / Environment
//
// When implementing a symbol table that maps identifiers to values:
//
//	type Environment struct {
//	    store map[string]Value  // Keys are normalized
//	    names map[string]string // normalized -> original case
//	}
//
//	func (e *Environment) Define(name string, val Value) {
//	    normalized := ident.Normalize(name)
//	    e.store[normalized] = val
//	    e.names[normalized] = name  // Preserve original
//	}
//
//	func (e *Environment) Get(name string) (Value, bool) {
//	    val, ok := e.store[ident.Normalize(name)]
//	    return val, ok
//	}
//
// ## Pattern 2: Function/Class Registry
//
// When building registries that need to look up declarations by name:
//
//	type FunctionRegistry struct {
//	    functions map[string]*ast.FunctionDecl  // normalized keys
//	}
//
//	func (r *FunctionRegistry) Register(name string, fn *ast.FunctionDecl) {
//	    r.functions[ident.Normalize(name)] = fn
//	}
//
//	func (r *FunctionRegistry) Lookup(name string) *ast.FunctionDecl {
//	    return r.functions[ident.Normalize(name)]
//	}
//
// ## Pattern 3: Semantic Analysis
//
// When checking if an identifier matches a specific value:
//
//	// ✅ Good: Use Equal for one-off comparisons
//	if ident.Equal(methodName, "Create") {
//	    // Handle constructor
//	}
//
//	// ❌ Avoid: Creating temporary strings
//	if strings.ToLower(methodName) == "create" {
//	    // Wasteful allocation
//	}
//
// ## Pattern 4: Keyword Checking
//
// When checking if an identifier is a keyword:
//
//	if ident.IsKeyword(name, "begin", "end", "if", "then") {
//	    // Handle keyword
//	}
//
// ## Pattern 5: Sorting Identifiers
//
// When sorting identifiers for deterministic output:
//
//	names := []string{"Charlie", "alice", "BOB"}
//	sort.Slice(names, func(i, j int) bool {
//	    return ident.Compare(names[i], names[j]) < 0
//	})
//	// Result: ["alice", "BOB", "Charlie"]
//
// # Performance Considerations
//
//   - Normalize() allocates a new string if the input isn't all lowercase.
//     Use it when storing keys, not in hot comparison paths.
//
//   - Equal() is optimized for comparisons and doesn't allocate.
//     Use it for one-off checks.
//
// - Compare() normalizes both strings. Cache the result if sorting repeatedly.
//
// # Migration from Existing Code
//
// Replace direct string operations with these helpers:
//
//	// Before
//	store[strings.ToLower(name)] = value
//	if strings.ToLower(a) == strings.ToLower(b) { ... }
//	if strings.EqualFold(name, "Create") { ... }
//
//	// After
//	store[ident.Normalize(name)] = value
//	if ident.Equal(a, b) { ... }
//	if ident.Equal(name, "Create") { ... }
//
// # Future Enhancements
//
// This package provides a centralized location for identifier normalization.
// Future enhancements could include:
//
//   - Unicode-aware folding using golang.org/x/text/cases
//   - Identifier interning for reduced memory usage
//   - Full Identifier type with normalization as a type invariant
//   - Locale-aware comparison for international identifiers
//
// See the main package documentation and examples for more details.
package ident
