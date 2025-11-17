package ident_test

import (
	"fmt"
	"sort"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// This example demonstrates how to use Normalize for map keys.
// Identifiers are normalized once when stored, allowing case-insensitive lookups.
func ExampleNormalize() {
	// Create a symbol table with normalized keys
	variables := make(map[string]int)

	// Store with original case, but use normalized key
	variables[ident.Normalize("MyVariable")] = 42
	variables[ident.Normalize("Counter")] = 10

	// Lookup works with any case
	val1 := variables[ident.Normalize("myvariable")] // 42
	val2 := variables[ident.Normalize("COUNTER")]    // 10

	fmt.Println(val1)
	fmt.Println(val2)
	// Output:
	// 42
	// 10
}

// This example shows how to use Equal for case-insensitive comparisons.
// It's more efficient than normalizing both strings for one-off checks.
func ExampleEqual() {
	// Check if a function name matches a known builtin
	funcName := "PrintLn"

	if ident.Equal(funcName, "println") {
		fmt.Println("Calling PrintLn builtin")
	}

	// Works with any case variation
	if ident.Equal("BEGIN", "begin") {
		fmt.Println("Keywords match")
	}

	// Output:
	// Calling PrintLn builtin
	// Keywords match
}

// This example demonstrates case-insensitive sorting using Compare.
func ExampleCompare() {
	// List of identifiers in mixed case
	names := []string{"zebra", "Apple", "BANANA", "cherry", "Date"}

	// Sort case-insensitively
	sort.Slice(names, func(i, j int) bool {
		return ident.Compare(names[i], names[j]) < 0
	})

	// Original case is preserved, but order is case-insensitive
	for _, name := range names {
		fmt.Println(name)
	}
	// Output:
	// Apple
	// BANANA
	// cherry
	// Date
	// zebra
}

// This example shows how to check if an identifier is in a list.
func ExampleContains() {
	keywords := []string{"begin", "end", "if", "then", "else"}

	// Check with different cases
	fmt.Println(ident.Contains(keywords, "BEGIN"))    // true
	fmt.Println(ident.Contains(keywords, "ELSE"))     // true
	fmt.Println(ident.Contains(keywords, "variable")) // false

	// Output:
	// true
	// true
	// false
}

// This example demonstrates finding the index of an identifier in a slice.
func ExampleIndex() {
	tokens := []string{"begin", "var", "x", "end"}

	// Find index with case-insensitive search
	idx1 := ident.Index(tokens, "VAR") // 1
	idx2 := ident.Index(tokens, "END") // 3
	idx3 := ident.Index(tokens, "if")  // -1 (not found)

	fmt.Println(idx1)
	fmt.Println(idx2)
	fmt.Println(idx3)
	// Output:
	// 1
	// 3
	// -1
}

// This example shows how to use IsKeyword for checking against multiple keywords.
func ExampleIsKeyword() {
	// Check if identifier is a control flow keyword
	name := "WHILE"

	if ident.IsKeyword(name, "if", "while", "for", "repeat") {
		fmt.Println("Control flow keyword")
	}

	// Not a keyword
	if !ident.IsKeyword("myVar", "if", "while", "for", "repeat") {
		fmt.Println("Not a keyword")
	}

	// Output:
	// Control flow keyword
	// Not a keyword
}

// This example demonstrates a complete symbol table implementation.
func Example_symbolTable() {
	// Symbol table that preserves original case for error messages
	type SymbolTable struct {
		values   map[string]int    // normalized -> value
		original map[string]string // normalized -> original case
	}

	st := SymbolTable{
		values:   make(map[string]int),
		original: make(map[string]string),
	}

	// Define variables
	define := func(name string, value int) {
		normalized := ident.Normalize(name)
		st.values[normalized] = value
		st.original[normalized] = name // Preserve original case
	}

	// Lookup variables
	lookup := func(name string) (int, string, bool) {
		normalized := ident.Normalize(name)
		val, ok := st.values[normalized]
		orig := st.original[normalized]
		return val, orig, ok
	}

	// Store with original case
	define("MyVariable", 42)
	define("COUNTER", 10)

	// Lookup with any case
	val1, orig1, _ := lookup("myvariable")
	val2, orig2, _ := lookup("counter")

	fmt.Printf("%s = %d\n", orig1, val1)
	fmt.Printf("%s = %d\n", orig2, val2)

	// Output:
	// MyVariable = 42
	// COUNTER = 10
}

// This example shows migration from existing code patterns.
func Example_migration() {
	// Old pattern: Direct strings.ToLower()
	// oldMap := make(map[string]string)
	// oldMap[strings.ToLower("MyKey")] = "value"

	// New pattern: Use ident.Normalize()
	newMap := make(map[string]string)
	newMap[ident.Normalize("MyKey")] = "value"
	fmt.Println(len(newMap) > 0) // true

	// Old pattern: strings.EqualFold()
	name := "Function"
	// if strings.EqualFold(name, "function") { ... }

	// New pattern: Use ident.Equal()
	if ident.Equal(name, "function") {
		fmt.Println("Matched")
	}

	// Output:
	// true
	// Matched
}
