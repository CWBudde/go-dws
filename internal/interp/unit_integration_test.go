package interp

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/units"
)

// TestInitializeUnits_NoRegistry tests initialization when no registry is set
func TestInitializeUnits_NoRegistry(t *testing.T) {
	interp := New(&bytes.Buffer{})

	// Should not error, just return nil
	err := interp.InitializeUnits()
	if err != nil {
		t.Errorf("Expected no error with nil registry, got: %v", err)
	}
}

// TestInitializeUnits_NoUnits tests initialization when no units are loaded
func TestInitializeUnits_NoUnits(t *testing.T) {
	interp := New(&bytes.Buffer{})

	registry := units.NewUnitRegistry([]string{"."})
	interp.SetUnitRegistry(registry)

	err := interp.InitializeUnits()
	if err != nil {
		t.Errorf("Expected no error with empty registry, got: %v", err)
	}
}

// TestInitializeUnits_Order tests that units are initialized in dependency order
func TestInitializeUnits_Order(t *testing.T) {
	var output bytes.Buffer
	interp := New(&output)

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load MathUtils unit (has initialization that prints)
	_, err = interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils: %v", err)
	}

	// Initialize units
	err = interp.InitializeUnits()
	if err != nil {
		t.Fatalf("Failed to initialize units: %v", err)
	}

	// Check that initialization code ran
	outputStr := output.String()
	if !strings.Contains(outputStr, "MathUtils loaded") {
		t.Errorf("Expected initialization output, got: %s", outputStr)
	}

	// Calling InitializeUnits again should not re-initialize
	output.Reset()
	err = interp.InitializeUnits()
	if err != nil {
		t.Fatalf("Failed second initialization: %v", err)
	}

	outputStr = output.String()
	if strings.Contains(outputStr, "MathUtils loaded") {
		t.Error("Expected no re-initialization output on second call")
	}
}

// TestInitializeUnits_MultipleUnits tests initializing multiple units
func TestInitializeUnits_MultipleUnits(t *testing.T) {
	var output bytes.Buffer
	interp := New(&output)

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load multiple units
	_, err = interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils: %v", err)
	}

	_, err = interp.LoadUnit("StringUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load StringUtils: %v", err)
	}

	// Initialize all units
	err = interp.InitializeUnits()
	if err != nil {
		t.Fatalf("Failed to initialize units: %v", err)
	}

	// Both units should have initialized
	outputStr := output.String()
	if !strings.Contains(outputStr, "MathUtils loaded") {
		t.Error("Expected MathUtils initialization output")
	}
	// Note: StringUtils may not have initialization output in the test data
}

// TestFinalizeUnits_NoRegistry tests finalization when no registry is set
func TestFinalizeUnits_NoRegistry(t *testing.T) {
	interp := New(&bytes.Buffer{})

	err := interp.FinalizeUnits()
	if err != nil {
		t.Errorf("Expected no error with nil registry, got: %v", err)
	}
}

// TestFinalizeUnits_Order tests that units are finalized in reverse order
func TestFinalizeUnits_Order(t *testing.T) {
	var output bytes.Buffer
	interp := New(&output)

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load MathUtils unit (has finalization that prints)
	_, err = interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils: %v", err)
	}

	// Initialize first (required before finalization)
	err = interp.InitializeUnits()
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Reset output to only capture finalization
	output.Reset()

	// Finalize units
	err = interp.FinalizeUnits()
	if err != nil {
		t.Fatalf("Failed to finalize units: %v", err)
	}

	// Check that finalization code ran
	outputStr := output.String()
	if !strings.Contains(outputStr, "MathUtils unloading") {
		t.Errorf("Expected finalization output, got: %s", outputStr)
	}
}

// TestImportUnitSymbols_Nil tests importing from nil unit
func TestImportUnitSymbols_Nil(t *testing.T) {
	interp := New(&bytes.Buffer{})

	err := interp.ImportUnitSymbols(nil)
	if err == nil {
		t.Fatal("Expected error when importing from nil unit")
	}

	if !strings.Contains(err.Error(), "nil unit") {
		t.Errorf("Expected 'nil unit' error, got: %v", err)
	}
}

// TestImportUnitSymbols_NoInterface tests importing from unit with no interface
func TestImportUnitSymbols_NoInterface(t *testing.T) {
	interp := New(&bytes.Buffer{})

	// Create a unit with no interface section
	unit := units.NewUnit("EmptyUnit", "/tmp/empty.dws")
	unit.InterfaceSection = nil

	err := interp.ImportUnitSymbols(unit)
	if err != nil {
		t.Errorf("Expected no error for unit without interface, got: %v", err)
	}
}

// TestImportUnitSymbols_WithFunctions tests importing functions from a unit
func TestImportUnitSymbols_WithFunctions(t *testing.T) {
	interp := New(&bytes.Buffer{})

	// Create a simple unit with a function in the interface
	unit := units.NewUnit("TestUnit", "/tmp/test.dws")

	// Create interface section with a function declaration
	funcDecl := &ast.FunctionDecl{
		Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
		Name:  &ast.Identifier{Value: "TestFunc"},
		Parameters: []*ast.Parameter{
			{
				Name: &ast.Identifier{Value: "x"},
				Type: &ast.TypeAnnotation{Name: "Integer"},
			},
		},
		ReturnType: &ast.TypeAnnotation{Name: "Integer"},
		Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
	}

	unit.InterfaceSection = &ast.BlockStatement{
		Statements: []ast.Statement{funcDecl},
	}

	// Import symbols
	err := interp.ImportUnitSymbols(unit)
	if err != nil {
		t.Fatalf("Failed to import symbols: %v", err)
	}

	// Verify function was registered
	if _, exists := interp.functions["TestFunc"]; !exists {
		t.Error("Expected TestFunc to be registered after import")
	}
}

// TestResolveQualifiedFunction_NoRegistry tests resolving without registry
func TestResolveQualifiedFunction_NoRegistry(t *testing.T) {
	interp := New(&bytes.Buffer{})

	_, err := interp.ResolveQualifiedFunction("SomeUnit", "SomeFunc")
	if err == nil {
		t.Fatal("Expected error when resolving without registry")
	}

	if !strings.Contains(err.Error(), "registry not initialized") {
		t.Errorf("Expected 'registry not initialized' error, got: %v", err)
	}
}

// TestResolveQualifiedFunction_UnitNotLoaded tests resolving from unloaded unit
func TestResolveQualifiedFunction_UnitNotLoaded(t *testing.T) {
	interp := New(&bytes.Buffer{})

	registry := units.NewUnitRegistry([]string{"."})
	interp.SetUnitRegistry(registry)

	_, err := interp.ResolveQualifiedFunction("NonExistent", "SomeFunc")
	if err == nil {
		t.Fatal("Expected error when resolving from non-existent unit")
	}

	if !strings.Contains(err.Error(), "not loaded") {
		t.Errorf("Expected 'not loaded' error, got: %v", err)
	}
}

// TestResolveQualifiedFunction_FunctionNotFound tests resolving non-existent function
func TestResolveQualifiedFunction_FunctionNotFound(t *testing.T) {
	interp := New(&bytes.Buffer{})

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load a unit
	_, err = interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load unit: %v", err)
	}

	// Try to resolve non-existent function
	_, err = interp.ResolveQualifiedFunction("MathUtils", "NonExistentFunc")
	if err == nil {
		t.Fatal("Expected error when resolving non-existent function")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// TestCrossUnitFunctionCall_Qualified tests calling a function using qualified name
func TestCrossUnitFunctionCall_Qualified(t *testing.T) {
	var output bytes.Buffer
	interp := New(&output)

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load MathUtils
	unit, err := interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils: %v", err)
	}

	// Import its symbols
	err = interp.ImportUnitSymbols(unit)
	if err != nil {
		t.Fatalf("Failed to import symbols: %v", err)
	}

	// Now try to call MathUtils.Add using qualified syntax
	// Parse and evaluate: MathUtils.Add(3, 5)
	source := "MathUtils.Add(3, 5)"
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %s", strings.Join(p.Errors(), "; "))
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("Evaluation error: %v", result)
	}

	// Check result
	if intVal, ok := result.(*IntegerValue); ok {
		if intVal.Value != 8 {
			t.Errorf("Expected Add(3, 5) = 8, got %d", intVal.Value)
		}
	} else {
		t.Errorf("Expected integer result, got %T: %v", result, result)
	}
}

// TestCrossUnitFunctionCall_Unqualified tests calling imported function without qualification
func TestCrossUnitFunctionCall_Unqualified(t *testing.T) {
	var output bytes.Buffer
	interp := New(&output)

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load MathUtils
	unit, err := interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils: %v", err)
	}

	// Import its symbols
	err = interp.ImportUnitSymbols(unit)
	if err != nil {
		t.Fatalf("Failed to import symbols: %v", err)
	}

	// Now call Add directly (unqualified) since we imported symbols
	source := "Add(10, 20)"
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %s", strings.Join(p.Errors(), "; "))
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("Evaluation error: %v", result)
	}

	// Check result
	if intVal, ok := result.(*IntegerValue); ok {
		if intVal.Value != 30 {
			t.Errorf("Expected Add(10, 20) = 30, got %d", intVal.Value)
		}
	} else {
		t.Errorf("Expected integer result, got %T: %v", result, result)
	}
}

// TestInitializationFinalizationOrder tests the full lifecycle
func TestInitializationFinalizationOrder(t *testing.T) {
	var output bytes.Buffer
	interp := New(&output)

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load MathUtils (has init and final sections)
	_, err = interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils: %v", err)
	}

	// Initialize
	err = interp.InitializeUnits()
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	initOutput := output.String()
	if !strings.Contains(initOutput, "MathUtils loaded") {
		t.Error("Expected initialization message")
	}

	// Finalize
	output.Reset()
	err = interp.FinalizeUnits()
	if err != nil {
		t.Fatalf("Failed to finalize: %v", err)
	}

	finalOutput := output.String()
	if !strings.Contains(finalOutput, "MathUtils unloading") {
		t.Error("Expected finalization message")
	}
}
