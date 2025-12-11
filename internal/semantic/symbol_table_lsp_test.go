package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestRecordUsage tests that symbol usages are tracked correctly
func TestRecordUsage(t *testing.T) {
	st := NewSymbolTable()

	// Define a variable at position 1:5
	declPos := token.Position{Line: 1, Column: 5, Offset: 4}
	st.Define("myVar", types.INTEGER, declPos)

	// Record some usages
	usage1 := token.Position{Line: 2, Column: 10, Offset: 15}
	usage2 := token.Position{Line: 3, Column: 8, Offset: 25}
	usage3 := token.Position{Line: 5, Column: 12, Offset: 45}

	st.RecordUsage("myVar", usage1)
	st.RecordUsage("myVar", usage2)
	st.RecordUsage("myVar", usage3)

	// Retrieve the symbol and check usages
	sym, ok := st.Resolve("myVar")
	if !ok {
		t.Fatal("Expected to find 'myVar'")
	}

	if len(sym.Usages) != 3 {
		t.Errorf("Expected 3 usages, got %d", len(sym.Usages))
	}

	expectedUsages := []token.Position{usage1, usage2, usage3}
	for i, expected := range expectedUsages {
		if sym.Usages[i] != expected {
			t.Errorf("Usage %d: expected %v, got %v", i, expected, sym.Usages[i])
		}
	}
}

// TestRecordUsageNonExistent tests that recording usage of non-existent symbol is a no-op
func TestRecordUsageNonExistent(t *testing.T) {
	st := NewSymbolTable()

	// Should not panic
	st.RecordUsage("nonExistent", token.Position{Line: 1, Column: 1})
}

// TestRecordUsageNestedScope tests that usage is recorded in the correct scope
func TestRecordUsageNestedScope(t *testing.T) {
	outer := NewSymbolTable()
	declPos := token.Position{Line: 1, Column: 5, Offset: 4}
	outer.Define("x", types.INTEGER, declPos)

	inner := NewEnclosedSymbolTable(outer)

	// Record usage from inner scope
	usage := token.Position{Line: 3, Column: 10, Offset: 25}
	inner.RecordUsage("x", usage)

	// Check that usage was recorded in outer scope
	sym, ok := outer.Resolve("x")
	if !ok {
		t.Fatal("Expected to find 'x' in outer scope")
	}

	if len(sym.Usages) != 1 {
		t.Errorf("Expected 1 usage, got %d", len(sym.Usages))
	}

	if sym.Usages[0] != usage {
		t.Errorf("Expected usage %v, got %v", usage, sym.Usages[0])
	}
}

// TestFindDefinition tests finding symbol definitions
func TestFindDefinition(t *testing.T) {
	st := NewSymbolTable()

	// Define a variable
	declPos := token.Position{Line: 1, Column: 5, Offset: 4}
	st.Define("myVar", types.INTEGER, declPos)

	// Find the definition
	sym, pos, found := st.FindDefinition("myVar")
	if !found {
		t.Fatal("Expected to find definition for 'myVar'")
	}

	if pos != declPos {
		t.Errorf("Expected position %v, got %v", declPos, pos)
	}

	if sym.Name != "myVar" {
		t.Errorf("Expected symbol name 'myVar', got '%s'", sym.Name)
	}

	// Case insensitive lookup
	_, pos2, found2 := st.FindDefinition("MYVAR")
	if !found2 {
		t.Fatal("Expected to find definition for 'MYVAR' (case insensitive)")
	}

	if pos2 != declPos {
		t.Errorf("Expected position %v, got %v", declPos, pos2)
	}
}

// TestFindDefinitionNonExistent tests finding non-existent symbols
func TestFindDefinitionNonExistent(t *testing.T) {
	st := NewSymbolTable()

	_, _, found := st.FindDefinition("nonExistent")
	if found {
		t.Error("Expected not to find non-existent symbol")
	}
}

// TestFindDefinitionNestedScope tests finding definitions in nested scopes
func TestFindDefinitionNestedScope(t *testing.T) {
	outer := NewSymbolTable()
	outerPos := token.Position{Line: 1, Column: 5, Offset: 4}
	outer.Define("outerVar", types.INTEGER, outerPos)

	inner := NewEnclosedSymbolTable(outer)
	innerPos := token.Position{Line: 5, Column: 10, Offset: 50}
	inner.Define("innerVar", types.STRING, innerPos)

	// Find outer variable from inner scope
	_, pos, found := inner.FindDefinition("outerVar")
	if !found {
		t.Fatal("Expected to find 'outerVar' from inner scope")
	}

	if pos != outerPos {
		t.Errorf("Expected position %v, got %v", outerPos, pos)
	}

	// Find inner variable
	_, pos2, found2 := inner.FindDefinition("innerVar")
	if !found2 {
		t.Fatal("Expected to find 'innerVar'")
	}

	if pos2 != innerPos {
		t.Errorf("Expected position %v, got %v", innerPos, pos2)
	}

	// Inner variable should not be visible from outer scope
	_, _, found3 := outer.FindDefinition("innerVar")
	if found3 {
		t.Error("Inner variable should not be visible from outer scope")
	}
}

// TestFindReferences tests finding all references to a symbol
func TestFindReferences(t *testing.T) {
	st := NewSymbolTable()

	// Define a variable
	declPos := token.Position{Line: 1, Column: 5, Offset: 4}
	st.Define("myVar", types.INTEGER, declPos)

	// Record usages
	usage1 := token.Position{Line: 2, Column: 10, Offset: 15}
	usage2 := token.Position{Line: 3, Column: 8, Offset: 25}

	st.RecordUsage("myVar", usage1)
	st.RecordUsage("myVar", usage2)

	// Find references
	refs := st.FindReferences("myVar")
	if len(refs) != 2 {
		t.Errorf("Expected 2 references, got %d", len(refs))
	}

	expectedRefs := []token.Position{usage1, usage2}
	for i, expected := range expectedRefs {
		if refs[i] != expected {
			t.Errorf("Reference %d: expected %v, got %v", i, expected, refs[i])
		}
	}

	// Case insensitive
	refs2 := st.FindReferences("MYVAR")
	if len(refs2) != 2 {
		t.Errorf("Expected 2 references (case insensitive), got %d", len(refs2))
	}
}

// TestFindReferencesNonExistent tests finding references for non-existent symbols
func TestFindReferencesNonExistent(t *testing.T) {
	st := NewSymbolTable()

	refs := st.FindReferences("nonExistent")
	if refs != nil {
		t.Error("Expected nil for non-existent symbol references")
	}
}

// TestFindReferencesReturnsACopy tests that FindReferences returns a copy
func TestFindReferencesReturnsACopy(t *testing.T) {
	st := NewSymbolTable()

	declPos := token.Position{Line: 1, Column: 5, Offset: 4}
	st.Define("myVar", types.INTEGER, declPos)

	usage := token.Position{Line: 2, Column: 10, Offset: 15}
	st.RecordUsage("myVar", usage)

	// Get references
	refs := st.FindReferences("myVar")

	// Modify the returned slice
	refs = append(refs, token.Position{Line: 99, Column: 99, Offset: 999})

	// Get references again - should not include the modification
	refs2 := st.FindReferences("myVar")
	if len(refs2) != 1 {
		t.Errorf("Expected 1 reference after modification, got %d", len(refs2))
	}
}

// TestUnusedSymbols tests finding unused symbols
func TestUnusedSymbols(t *testing.T) {
	st := NewSymbolTable()

	// Define some variables
	pos1 := token.Position{Line: 1, Column: 5, Offset: 4}
	pos2 := token.Position{Line: 2, Column: 5, Offset: 10}
	pos3 := token.Position{Line: 3, Column: 5, Offset: 16}

	st.Define("used", types.INTEGER, pos1)
	st.Define("unused1", types.STRING, pos2)
	st.Define("unused2", types.FLOAT, pos3)

	// Record usage for 'used'
	st.RecordUsage("used", token.Position{Line: 5, Column: 10, Offset: 50})

	// Get unused symbols
	unused := st.UnusedSymbols()

	if len(unused) != 2 {
		t.Errorf("Expected 2 unused symbols, got %d", len(unused))
	}

	// Check that the unused symbols are correct
	unusedNames := make(map[string]bool)
	for _, sym := range unused {
		unusedNames[sym.Name] = true
	}

	if !unusedNames["unused1"] {
		t.Error("Expected 'unused1' to be in unused symbols")
	}

	if !unusedNames["unused2"] {
		t.Error("Expected 'unused2' to be in unused symbols")
	}

	if unusedNames["used"] {
		t.Error("Expected 'used' not to be in unused symbols")
	}
}

// TestUnusedSymbolsOnlyCurrentScope tests that UnusedSymbols only checks current scope
func TestUnusedSymbolsOnlyCurrentScope(t *testing.T) {
	outer := NewSymbolTable()
	pos1 := token.Position{Line: 1, Column: 5, Offset: 4}
	outer.Define("outerUnused", types.INTEGER, pos1)

	inner := NewEnclosedSymbolTable(outer)
	pos2 := token.Position{Line: 5, Column: 5, Offset: 50}
	inner.Define("innerUnused", types.STRING, pos2)

	// Get unused from inner - should only get inner symbols
	unused := inner.UnusedSymbols()

	if len(unused) != 1 {
		t.Errorf("Expected 1 unused symbol in inner scope, got %d", len(unused))
	}

	if unused[0].Name != "innerUnused" {
		t.Errorf("Expected 'innerUnused', got '%s'", unused[0].Name)
	}
}

// TestUnusedSymbolsOverloadSet tests that overload sets with no usages are detected
func TestUnusedSymbolsOverloadSet(t *testing.T) {
	st := NewSymbolTable()

	// Define an overloaded function
	pos1 := token.Position{Line: 1, Column: 10, Offset: 9}
	funcType1 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER},
		[]string{"x"},
		nil, nil, nil, nil,
		types.INTEGER,
	)

	err := st.DefineOverload("foo", funcType1, true, false, pos1)
	if err != nil {
		t.Fatalf("Failed to define first overload: %v", err)
	}

	pos2 := token.Position{Line: 5, Column: 10, Offset: 50}
	funcType2 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.STRING},
		[]string{"s"},
		nil, nil, nil, nil,
		types.STRING,
	)

	err = st.DefineOverload("foo", funcType2, true, false, pos2)
	if err != nil {
		t.Fatalf("Failed to define second overload: %v", err)
	}

	// No usages recorded - should be detected as unused
	unused := st.UnusedSymbols()

	if len(unused) != 1 {
		t.Errorf("Expected 1 unused overload set, got %d", len(unused))
	}

	if unused[0].Name != "foo" {
		t.Errorf("Expected 'foo', got '%s'", unused[0].Name)
	}
}

// TestDeclPositionTracking tests that declaration positions are tracked correctly
func TestDeclPositionTracking(t *testing.T) {
	st := NewSymbolTable()

	tests := []struct {
		name     string
		varName  string
		varType  types.Type
		declPos  token.Position
		readonly bool
		isConst  bool
	}{
		{
			name:    "regular variable",
			varName: "x",
			varType: types.INTEGER,
			declPos: token.Position{Line: 1, Column: 5, Offset: 4},
		},
		{
			name:     "readonly variable",
			varName:  "y",
			varType:  types.STRING,
			declPos:  token.Position{Line: 2, Column: 8, Offset: 15},
			readonly: true,
		},
		{
			name:    "constant",
			varName: "PI",
			varType: types.FLOAT,
			declPos: token.Position{Line: 3, Column: 10, Offset: 30},
			isConst: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isConst {
				st.DefineConst(tt.varName, tt.varType, 3.14159, tt.declPos)
			} else if tt.readonly {
				st.DefineReadOnly(tt.varName, tt.varType, tt.declPos)
			} else {
				st.Define(tt.varName, tt.varType, tt.declPos)
			}

			sym, ok := st.Resolve(tt.varName)
			if !ok {
				t.Fatalf("Expected to find '%s'", tt.varName)
			}

			if sym.DeclPosition != tt.declPos {
				t.Errorf("Expected declaration position %v, got %v", tt.declPos, sym.DeclPosition)
			}
		})
	}
}

// TestFunctionDeclPositionTracking tests that function declaration positions are tracked
func TestFunctionDeclPositionTracking(t *testing.T) {
	st := NewSymbolTable()

	funcPos := token.Position{Line: 5, Column: 10, Offset: 50}
	funcType := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.STRING},
		[]string{"x", "s"},
		nil, nil, nil, nil,
		types.BOOLEAN,
	)

	st.DefineFunction("myFunc", funcType, funcPos)

	sym, ok := st.Resolve("myFunc")
	if !ok {
		t.Fatal("Expected to find 'myFunc'")
	}

	if sym.DeclPosition != funcPos {
		t.Errorf("Expected declaration position %v, got %v", funcPos, sym.DeclPosition)
	}
}

// TestOverloadDeclPositionTracking tests that overload positions are tracked separately
func TestOverloadDeclPositionTracking(t *testing.T) {
	st := NewSymbolTable()

	pos1 := token.Position{Line: 1, Column: 10, Offset: 9}
	funcType1 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER},
		[]string{"x"},
		nil, nil, nil, nil,
		types.INTEGER,
	)

	err := st.DefineOverload("foo", funcType1, true, false, pos1)
	if err != nil {
		t.Fatalf("Failed to define first overload: %v", err)
	}

	pos2 := token.Position{Line: 10, Column: 10, Offset: 100}
	funcType2 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.STRING},
		[]string{"s"},
		nil, nil, nil, nil,
		types.STRING,
	)

	err = st.DefineOverload("foo", funcType2, true, false, pos2)
	if err != nil {
		t.Fatalf("Failed to define second overload: %v", err)
	}

	// Get the overload set
	overloads := st.GetOverloadSet("foo")
	if len(overloads) != 2 {
		t.Fatalf("Expected 2 overloads, got %d", len(overloads))
	}

	// Check that each overload has the correct position
	if overloads[0].DeclPosition != pos1 {
		t.Errorf("First overload: expected position %v, got %v", pos1, overloads[0].DeclPosition)
	}

	if overloads[1].DeclPosition != pos2 {
		t.Errorf("Second overload: expected position %v, got %v", pos2, overloads[1].DeclPosition)
	}
}
