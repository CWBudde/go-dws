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

// This example shows how to use HasPrefix for case-insensitive prefix matching.
func ExampleHasPrefix() {
	// Check if a type name starts with "Array" (case-insensitive)
	typeName := "ArrayOfInteger"

	if ident.HasPrefix(typeName, "array") {
		fmt.Println("Is an array type")
	}

	// Works with any case variation
	if ident.HasPrefix("ARRAYLIST", "Array") {
		fmt.Println("Also matches uppercase")
	}

	// No match when prefix is not present
	if !ident.HasPrefix("Integer", "Array") {
		fmt.Println("Integer is not an array")
	}

	// Output:
	// Is an array type
	// Also matches uppercase
	// Integer is not an array
}

// This example shows how to use HasSuffix for case-insensitive suffix matching.
func ExampleHasSuffix() {
	// Check if a type name ends with "Type" (case-insensitive)
	typeName := "MyCustomType"

	if ident.HasSuffix(typeName, "type") {
		fmt.Println("Is a type alias")
	}

	// Works with any case variation
	if ident.HasSuffix("MYTYPE", "Type") {
		fmt.Println("Also matches uppercase")
	}

	// No match when suffix is not present
	if !ident.HasSuffix("Integer", "Type") {
		fmt.Println("Integer doesn't end with Type")
	}

	// Output:
	// Is a type alias
	// Also matches uppercase
	// Integer doesn't end with Type
}

// This example demonstrates proper error message handling that preserves
// the user's original casing while still performing case-insensitive lookups.
func Example_errorMessages() {
	// Simulated symbol table with original casing preserved
	symbols := map[string]string{
		"myvariable": "MyVariable", // normalized -> original
		"counter":    "Counter",
	}
	values := map[string]int{
		"myvariable": 42,
		"counter":    10,
	}

	// Function that checks if a variable exists and reports errors
	checkVariable := func(userInput string) {
		normalized := ident.Normalize(userInput)
		if _, exists := values[normalized]; exists {
			// Variable found - use original definition casing
			original := symbols[normalized]
			fmt.Printf("Found variable '%s' (defined as '%s')\n", userInput, original)
		} else {
			// Variable not found - use user's input casing in error
			// IMPORTANT: Never normalize the user's input in error messages!
			fmt.Printf("Error: undefined variable '%s'\n", userInput)
		}
	}

	// User looks up variables with different casings
	checkVariable("MYVARIABLE")   // Found, shows original definition
	checkVariable("counter")      // Found, shows original definition
	checkVariable("UndefinedVar") // Not found, shows user's casing

	// Output:
	// Found variable 'MYVARIABLE' (defined as 'MyVariable')
	// Found variable 'counter' (defined as 'Counter')
	// Error: undefined variable 'UndefinedVar'
}

// This example shows a type registry pattern commonly used in compilers.
func Example_typeRegistry() {
	// Registry stores types with normalized keys but preserves original names
	type TypeInfo struct {
		Name string // Original casing as defined
		Kind string // "class", "record", "enum", etc.
	}

	registry := make(map[string]*TypeInfo) // normalized -> TypeInfo

	// Register types with their original casing
	register := func(name, kind string) {
		registry[ident.Normalize(name)] = &TypeInfo{Name: name, Kind: kind}
	}

	// Look up a type (case-insensitive)
	lookup := func(name string) *TypeInfo {
		return registry[ident.Normalize(name)]
	}

	// Register some types
	register("TMyClass", "class")
	register("TPoint", "record")
	register("TColor", "enum")

	// Look up with various casings
	if info := lookup("tmyclass"); info != nil {
		fmt.Printf("Found %s '%s'\n", info.Kind, info.Name)
	}
	if info := lookup("TPOINT"); info != nil {
		fmt.Printf("Found %s '%s'\n", info.Kind, info.Name)
	}
	if info := lookup("tcolor"); info != nil {
		fmt.Printf("Found %s '%s'\n", info.Kind, info.Name)
	}

	// Output:
	// Found class 'TMyClass'
	// Found record 'TPoint'
	// Found enum 'TColor'
}

// This example demonstrates using the Map type for a simple symbol table.
func ExampleMap() {
	// Create a case-insensitive map for variables
	variables := ident.NewMap[int]()

	// Store variables with their original casing
	variables.Set("MyVariable", 42)
	variables.Set("Counter", 10)

	// Lookup works with any case
	val1, _ := variables.Get("myvariable") // 42
	val2, _ := variables.Get("COUNTER")    // 10

	fmt.Println(val1)
	fmt.Println(val2)

	// Get the originally defined name
	fmt.Println(variables.GetOriginalKey("MYVARIABLE"))

	// Output:
	// 42
	// 10
	// MyVariable
}

// This example shows using Map.SetIfAbsent for define-once semantics.
func ExampleMap_SetIfAbsent() {
	symbols := ident.NewMap[int]()

	// First definition succeeds
	if symbols.SetIfAbsent("MyVar", 42) {
		fmt.Println("MyVar defined")
	}

	// Second definition with different case fails
	if !symbols.SetIfAbsent("myvar", 100) {
		orig := symbols.GetOriginalKey("myvar")
		fmt.Printf("Cannot redefine '%s'\n", orig)
	}

	// Value unchanged
	val, _ := symbols.Get("MyVar")
	fmt.Printf("Value: %d\n", val)

	// Output:
	// MyVar defined
	// Cannot redefine 'MyVar'
	// Value: 42
}

// This example shows iterating over Map entries.
func ExampleMap_Range() {
	m := ident.NewMap[int]()
	m.Set("Alpha", 1)
	m.Set("Beta", 2)
	m.Set("Gamma", 3)

	// Collect entries (order not guaranteed, so we just count)
	count := 0
	m.Range(func(key string, value int) bool {
		count++
		return true
	})

	fmt.Printf("Map has %d entries\n", count)

	// Output:
	// Map has 3 entries
}
