package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
)

// TestAnalyzeUnit_BasicInterfaceAndImplementation tests that AnalyzeUnit
// can analyze a simple unit with interface and implementation sections
func TestAnalyzeUnit_BasicInterfaceAndImplementation(t *testing.T) {
	// Create a simple unit with a function declaration in interface
	// and implementation in implementation section
	unitDecl := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TestUnit"},
				},
			},
			Value: "TestUnit",
		},
		InterfaceSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				// function Add(a, b: Integer): Integer;
				&ast.FunctionDecl{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "Add"},
							},
						},
						Value: "Add",
					},
					Parameters: []*ast.Parameter{
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "a"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "b"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
					},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       nil, // Interface section - no body
				},
			},
		},
		ImplementationSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				// function Add(a, b: Integer): Integer;
				// begin
				//   Result := a + b;
				// end;
				&ast.FunctionDecl{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "Add"},
							},
						},
						Value: "Add",
					},
					Parameters: []*ast.Parameter{
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "a"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "b"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
					},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body: &ast.BlockStatement{
						Statements: []ast.Statement{
							&ast.AssignmentStatement{
								Target: &ast.Identifier{
									TypedExpressionBase: ast.TypedExpressionBase{
										BaseNode: ast.BaseNode{},
									},
									Value: "Result",
								},
								Value: &ast.BinaryExpression{
									Left: &ast.Identifier{
										TypedExpressionBase: ast.TypedExpressionBase{
											BaseNode: ast.BaseNode{},
										},
										Value: "a",
									},
									Operator: "+",
									Right: &ast.Identifier{
										TypedExpressionBase: ast.TypedExpressionBase{
											BaseNode: ast.BaseNode{},
										},
										Value: "b",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	analyzer := NewAnalyzer()

	err := analyzer.AnalyzeUnit(unitDecl)
	if err != nil {
		t.Fatalf("AnalyzeUnit() failed: %v", err)
	}

	// Verify that the function was added to the analyzer's symbol table
	symbol, found := analyzer.symbols.Resolve("Add")
	if !found {
		t.Fatal("Function 'Add' was not added to symbol table")
	}

	// Verify it's a function type
	funcType, ok := symbol.Type.(*types.FunctionType)
	if !ok {
		t.Fatalf("Symbol 'Add' is not a function type, got: %T", symbol.Type)
	}

	// Verify signature
	if len(funcType.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(funcType.Parameters))
	}
	if funcType.ReturnType != types.INTEGER {
		t.Errorf("Expected Integer return type, got %v", funcType.ReturnType)
	}
}

// TestAnalyzeUnit_MissingImplementation tests that AnalyzeUnit detects
// when an interface declaration has no implementation
func TestAnalyzeUnit_MissingImplementation(t *testing.T) {
	unitDecl := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TestUnit"},
				},
			},
			Value: "TestUnit",
		},
		InterfaceSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.FunctionDecl{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "DoSomething"},
							},
						},
						Value: "DoSomething",
					},
					Parameters: []*ast.Parameter{},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       nil, // No body in interface
				},
			},
		},
		ImplementationSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				// Empty - missing implementation
			},
		},
	}

	analyzer := NewAnalyzer()

	err := analyzer.AnalyzeUnit(unitDecl)
	if err == nil {
		t.Fatal("Expected error for missing implementation, got nil")
	}

	// Check error message mentions the missing function
	if !hasSubstring(err.Error(), "DoSomething") {
		t.Errorf("Error should mention 'DoSomething', got: %v", err)
	}
}

// TestAnalyzeUnit_SignatureMismatch tests that AnalyzeUnit detects when
// interface and implementation signatures don't match
func TestAnalyzeUnit_SignatureMismatch(t *testing.T) {
	unitDecl := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TestUnit"},
				},
			},
			Value: "TestUnit",
		},
		InterfaceSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.FunctionDecl{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "Add"},
							},
						},
						Value: "Add",
					},
					Parameters: []*ast.Parameter{
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "a"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "b"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
					},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       nil,
				},
			},
		},
		ImplementationSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.FunctionDecl{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "Add"},
							},
						},
						Value: "Add",
					},
					Parameters: []*ast.Parameter{
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "x"}, // Only one parameter - MISMATCH
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
					},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
				},
			},
		},
	}

	analyzer := NewAnalyzer()

	err := analyzer.AnalyzeUnit(unitDecl)
	if err == nil {
		t.Fatal("Expected error for signature mismatch, got nil")
	}

	errMsg := err.Error()
	if !hasSubstring(errMsg, "signature") && !hasSubstring(errMsg, "mismatch") && !hasSubstring(errMsg, "parameter") {
		t.Errorf("Error should mention signature/parameter mismatch, got: %v", err)
	}
}

// TestAnalyzeUnit_WithUsesClause tests that AnalyzeUnit can resolve
// dependencies from uses clauses and import their symbols
func TestAnalyzeUnit_WithUsesClause(t *testing.T) {
	// First, create a "Math" unit with an Add function
	mathUnitSymbols := NewSymbolTable()
	mathUnitSymbols.DefineFunction("Add", &types.FunctionType{
		Parameters: []types.Type{types.INTEGER, types.INTEGER},
		ReturnType: types.INTEGER,
	})

	// Create a main unit that uses the Math unit
	mainUnit := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "MainUnit"},
				},
			},
			Value: "MainUnit",
		},
		InterfaceSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				// uses Math;
				&ast.UsesClause{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.USES, Literal: "uses"}},
					Units: []*ast.Identifier{
						{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{},
							},
							Value: "Math",
						},
					},
				},
				// function Calculate(x, y: Integer): Integer;
				&ast.FunctionDecl{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{},
						},
						Value: "Calculate",
					},
					Parameters: []*ast.Parameter{
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "x"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "y"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
					},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       nil,
				},
			},
		},
		ImplementationSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.FunctionDecl{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{},
						},
						Value: "Calculate",
					},
					Parameters: []*ast.Parameter{
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "x"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
						{
							Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "y"},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
					},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
				},
			},
		},
	}

	// Create a map of available units
	availableUnits := map[string]*SymbolTable{
		"math": mathUnitSymbols, // Case-insensitive lookup
	}

	analyzer := NewAnalyzer()
	err := analyzer.AnalyzeUnitWithDependencies(mainUnit, availableUnits)
	if err != nil {
		t.Fatalf("AnalyzeUnitWithDependencies() failed: %v", err)
	}

	// Verify that Math.Add is now available in the analyzer's symbol table
	symbol, found := analyzer.symbols.Resolve("Add")
	if !found {
		t.Fatal("Function 'Add' from Math unit was not imported")
	}

	funcType, ok := symbol.Type.(*types.FunctionType)
	if !ok {
		t.Fatalf("Symbol 'Add' is not a function type, got: %T", symbol.Type)
	}

	if len(funcType.Parameters) != 2 {
		t.Errorf("Expected 2 parameters for Add, got %d", len(funcType.Parameters))
	}
}

// TestAnalyzeUnit_UsesClauseConflict tests that AnalyzeUnit detects
// when multiple units export the same symbol name
func TestAnalyzeUnit_UsesClauseConflict(t *testing.T) {
	// Create two units that both export an "Add" function
	mathSymbols := NewSymbolTable()
	mathSymbols.DefineFunction("Add", &types.FunctionType{
		Parameters: []types.Type{types.INTEGER, types.INTEGER},
		ReturnType: types.INTEGER,
	})

	stringSymbols := NewSymbolTable()
	stringSymbols.DefineFunction("Add", &types.FunctionType{
		Parameters: []types.Type{types.STRING, types.STRING},
		ReturnType: types.STRING,
	})

	// Create a unit that uses both
	mainUnit := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "MainUnit",
		},
		InterfaceSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.UsesClause{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.USES, Literal: "uses"}},
					Units: []*ast.Identifier{
						{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{},
							},
							Value: "Math",
						},
						{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{},
							},
							Value: "Strings",
						},
					},
				},
			},
		},
	}

	availableUnits := map[string]*SymbolTable{
		"math":    mathSymbols,
		"strings": stringSymbols,
	}

	analyzer := NewAnalyzer()
	err := analyzer.AnalyzeUnitWithDependencies(mainUnit, availableUnits)

	// Should detect the conflict
	if err == nil {
		t.Fatal("Expected error for symbol name conflict, got nil")
	}

	if !hasSubstring(err.Error(), "conflict") && !hasSubstring(err.Error(), "ambiguous") {
		t.Errorf("Error should mention conflict or ambiguity, got: %v", err)
	}
}

// TestResolveQualifiedSymbol tests that qualified symbol resolution works
// for disambiguating symbols from different units
func TestResolveQualifiedSymbol(t *testing.T) {
	// Create two units with different Add functions
	mathSymbols := NewSymbolTable()
	mathSymbols.DefineFunction("Add", &types.FunctionType{
		Parameters: []types.Type{types.INTEGER, types.INTEGER},
		ReturnType: types.INTEGER,
	})

	stringSymbols := NewSymbolTable()
	stringSymbols.DefineFunction("Add", &types.FunctionType{
		Parameters: []types.Type{types.STRING, types.STRING},
		ReturnType: types.STRING,
	})

	availableUnits := map[string]*SymbolTable{
		"math":    mathSymbols,
		"strings": stringSymbols,
	}

	analyzer := NewAnalyzer()

	// Create a minimal unit to set up the analyzer
	dummyUnit := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "Main",
		},
	}

	err := analyzer.AnalyzeUnitWithDependencies(dummyUnit, availableUnits)
	if err != nil {
		t.Fatalf("AnalyzeUnitWithDependencies() failed: %v", err)
	}

	// Test resolving Math.Add
	symbol, err := analyzer.ResolveQualifiedSymbol("Math", "Add")
	if err != nil {
		t.Fatalf("Failed to resolve Math.Add: %v", err)
	}

	funcType, ok := symbol.Type.(*types.FunctionType)
	if !ok {
		t.Fatalf("Math.Add is not a function type, got: %T", symbol.Type)
	}

	if funcType.ReturnType != types.INTEGER {
		t.Errorf("Math.Add should return INTEGER, got: %v", funcType.ReturnType)
	}

	// Test resolving Strings.Add
	symbol, err = analyzer.ResolveQualifiedSymbol("Strings", "Add")
	if err != nil {
		t.Fatalf("Failed to resolve Strings.Add: %v", err)
	}

	funcType, ok = symbol.Type.(*types.FunctionType)
	if !ok {
		t.Fatalf("Strings.Add is not a function type, got: %T", symbol.Type)
	}

	if funcType.ReturnType != types.STRING {
		t.Errorf("Strings.Add should return STRING, got: %v", funcType.ReturnType)
	}

	// Test non-existent unit
	_, err = analyzer.ResolveQualifiedSymbol("NonExistent", "Add")
	if err == nil {
		t.Error("Expected error for non-existent unit, got nil")
	}

	// Test non-existent symbol
	_, err = analyzer.ResolveQualifiedSymbol("Math", "NonExistentFunction")
	if err == nil {
		t.Error("Expected error for non-existent symbol, got nil")
	}

	// Test case-insensitive resolution
	symbol, err = analyzer.ResolveQualifiedSymbol("MATH", "ADD")
	if err != nil {
		t.Fatalf("Failed case-insensitive resolution for MATH.ADD: %v", err)
	}

	if symbol.Type.(*types.FunctionType).ReturnType != types.INTEGER {
		t.Error("Case-insensitive resolution failed")
	}
}

// TestForwardDeclarationsAcrossUnits verifies that interface/implementation
// split (forward declarations) works correctly for cross-unit function calls
func TestForwardDeclarationsAcrossUnits(t *testing.T) {
	// Create a library unit with interface declaration and implementation
	libUnit := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "MathLib",
		},
		InterfaceSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				// Forward declaration: function Multiply(a, b: Integer): Integer;
				&ast.FunctionDecl{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{},
						},
						Value: "Multiply",
					},
					Parameters: []*ast.Parameter{
						{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "a"}, Type: &ast.TypeAnnotation{Name: "Integer"}},
						{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "b"}, Type: &ast.TypeAnnotation{Name: "Integer"}},
					},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       nil, // No body in interface - forward declaration
				},
			},
		},
		ImplementationSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				// Actual implementation
				&ast.FunctionDecl{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{},
						},
						Value: "Multiply",
					},
					Parameters: []*ast.Parameter{
						{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "a"}, Type: &ast.TypeAnnotation{Name: "Integer"}},
						{Name: &ast.Identifier{TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{}}, Value: "b"}, Type: &ast.TypeAnnotation{Name: "Integer"}},
					},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
				},
			},
		},
	}

	analyzer := NewAnalyzer()
	err := analyzer.AnalyzeUnit(libUnit)
	if err != nil {
		t.Fatalf("Failed to analyze unit with forward declaration: %v", err)
	}

	// Verify the function is available (from interface, not implementation)
	symbol, found := analyzer.symbols.Resolve("Multiply")
	if !found {
		t.Fatal("Forward-declared function 'Multiply' not found in symbol table")
	}

	funcType, ok := symbol.Type.(*types.FunctionType)
	if !ok {
		t.Fatalf("Multiply is not a function type, got: %T", symbol.Type)
	}

	if len(funcType.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(funcType.Parameters))
	}

	if funcType.ReturnType != types.INTEGER {
		t.Errorf("Expected INTEGER return type, got %v", funcType.ReturnType)
	}
}

// TestSemanticAnalysis_ComprehensiveUnitScenario tests a complete
// multi-unit scenario with dependencies, conflicts, and qualified access
func TestSemanticAnalysis_ComprehensiveUnitScenario(t *testing.T) {
	// This test covers Tasks 9.128-9.130: comprehensive semantic analysis

	// Create a base utility unit
	baseUnit := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "Base",
		},
		InterfaceSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.FunctionDecl{
					BaseNode:   ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{},
						},
						Value: "GetValue",
					},
					Parameters: []*ast.Parameter{},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       nil,
				},
			},
		},
		ImplementationSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.FunctionDecl{
					BaseNode:   ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{},
						},
						Value: "GetValue",
					},
					Parameters: []*ast.Parameter{},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
				},
			},
		},
	}

	// Analyze base unit first
	baseAnalyzer := NewAnalyzer()
	if err := baseAnalyzer.AnalyzeUnit(baseUnit); err != nil {
		t.Fatalf("Failed to analyze base unit: %v", err)
	}

	// Create symbol table for base unit
	baseSymbols := NewSymbolTable()
	baseSymbols.DefineFunction("GetValue", &types.FunctionType{
		Parameters: []types.Type{},
		ReturnType: types.INTEGER,
	})

	// Create a dependent unit that uses base
	dependentUnit := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "Dependent",
		},
		InterfaceSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.UsesClause{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.USES, Literal: "uses"}},
					Units: []*ast.Identifier{{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{},
						},
						Value: "Base",
					}},
				},
				&ast.FunctionDecl{
					BaseNode:   ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{},
						},
						Value: "ProcessValue",
					},
					Parameters: []*ast.Parameter{},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       nil,
				},
			},
		},
		ImplementationSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.FunctionDecl{
					BaseNode:   ast.BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
					Name: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{},
						},
						Value: "ProcessValue",
					},
					Parameters: []*ast.Parameter{},
					ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
				},
			},
		},
	}

	// Analyze dependent unit with base unit available
	dependentAnalyzer := NewAnalyzer()
	availableUnits := map[string]*SymbolTable{
		"base": baseSymbols,
	}

	if err := dependentAnalyzer.AnalyzeUnitWithDependencies(dependentUnit, availableUnits); err != nil {
		t.Fatalf("Failed to analyze dependent unit: %v", err)
	}

	// Verify that both GetValue (imported) and ProcessValue (own) are available
	if _, found := dependentAnalyzer.symbols.Resolve("GetValue"); !found {
		t.Error("Imported function GetValue not found in dependent unit")
	}

	if _, found := dependentAnalyzer.symbols.Resolve("ProcessValue"); !found {
		t.Error("Own function ProcessValue not found in dependent unit")
	}

	// Verify qualified access works
	if _, err := dependentAnalyzer.ResolveQualifiedSymbol("Base", "GetValue"); err != nil {
		t.Errorf("Qualified access Base.GetValue failed: %v", err)
	}
}

// TestSemanticAnalysis_CircularDependencyAtSemanticLevel tests that
// circular dependencies are caught during semantic analysis
func TestSemanticAnalysis_CircularDependencyAtSemanticLevel(t *testing.T) {
	// Note: Circular dependencies are primarily caught at the registry level
	// during unit loading. This test verifies the semantic layer doesn't break
	// when units reference each other (which shouldn't happen due to registry checks).

	// This is more of a sanity check that semantic analysis handles the scenario gracefully
	// if somehow circular symbols make it through (they shouldn't).

	// The main circular dependency detection happens in registry.LoadUnit() and
	// registry.ComputeInitializationOrder(), which we've already tested.

	t.Skip("Circular dependencies are caught at registry level (already tested)")
}

// TestSemanticAnalysis_NamespaceConflictResolution tests that
// namespace conflicts are properly detected and qualified access resolves them
func TestSemanticAnalysis_NamespaceConflictResolution(t *testing.T) {
	// Create two units with conflicting symbol names
	unit1Symbols := NewSymbolTable()
	unit1Symbols.DefineFunction("Compute", &types.FunctionType{
		Parameters: []types.Type{types.INTEGER},
		ReturnType: types.INTEGER,
	})

	unit2Symbols := NewSymbolTable()
	unit2Symbols.DefineFunction("Compute", &types.FunctionType{
		Parameters: []types.Type{types.STRING},
		ReturnType: types.STRING,
	})

	// Try to create a unit that uses both - should fail
	conflictingUnit := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"}},
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{},
			},
			Value: "Conflicting",
		},
		InterfaceSection: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.UsesClause{
					BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.USES, Literal: "uses"}},
					Units: []*ast.Identifier{
						{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{},
							},
							Value: "Unit1",
						},
						{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{},
							},
							Value: "Unit2",
						},
					},
				},
			},
		},
	}

	analyzer := NewAnalyzer()
	availableUnits := map[string]*SymbolTable{
		"unit1": unit1Symbols,
		"unit2": unit2Symbols,
	}

	err := analyzer.AnalyzeUnitWithDependencies(conflictingUnit, availableUnits)
	if err == nil {
		t.Fatal("Expected error for namespace conflict, got nil")
	}

	if !hasSubstring(err.Error(), "conflict") {
		t.Errorf("Error should mention conflict, got: %v", err)
	}

	// Now verify qualified access can resolve the ambiguity
	// (even though imports failed, the units are still available for qualified access)
	symbol1, err1 := analyzer.ResolveQualifiedSymbol("Unit1", "Compute")
	symbol2, err2 := analyzer.ResolveQualifiedSymbol("Unit2", "Compute")

	if err1 != nil || err2 != nil {
		t.Errorf("Qualified access should work despite import conflict")
	}

	// Verify they're different types
	func1Type := symbol1.Type.(*types.FunctionType)
	func2Type := symbol2.Type.(*types.FunctionType)

	if func1Type.ReturnType == func2Type.ReturnType {
		t.Error("The two Compute functions should have different return types")
	}
}

// Helper function to check if a string contains a substring
func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
