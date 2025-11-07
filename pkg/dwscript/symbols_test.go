package dwscript

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/token"
)

func TestProgram_Symbols(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		wantSymbols    []string // Symbol names we expect to find
		wantKinds      map[string]string
		expectNoSymbol []string // Symbol names we expect NOT to find
	}{
		{
			name: "variables and constants",
			source: `
				var x: Integer := 42;
				var name: String := 'Hello';
				const PI = 3.14;
			`,
			wantSymbols: []string{"x", "name", "PI"},
			wantKinds: map[string]string{
				"x":    "variable",
				"name": "variable",
				"PI":   "constant",
			},
		},
		{
			name: "functions",
			source: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				procedure PrintMessage(msg: String);
				begin
					PrintLn(msg);
				end;
			`,
			wantSymbols: []string{"Add", "PrintMessage"},
			wantKinds: map[string]string{
				"Add":          "function",
				"PrintMessage": "function",
			},
		},
		{
			name: "class declaration",
			source: `
				type
					TPoint = class
					private
						FX, FY: Integer;
					public
						constructor Create(AX, AY: Integer);
						begin
							FX := AX;
							FY := AY;
						end;
						function GetX: Integer;
						begin
							Result := FX;
						end;
					end;
			`,
			wantSymbols: []string{"tpoint"}, // lowercase due to case-insensitive storage
			wantKinds: map[string]string{
				"tpoint": "class",
			},
		},
		{
			name: "enum declaration",
			source: `
				type TColor = (Red, Green, Blue);
				var c: TColor;
			`,
			wantSymbols: []string{"c", "tcolor"}, // enum type and variable
			wantKinds: map[string]string{
				"c":      "variable",
				"tcolor": "enum",
			},
		},
		{
			name: "type alias",
			source: `
				type TInteger = Integer;
				var x: TInteger;
			`,
			wantSymbols: []string{"x", "tinteger"},
			wantKinds: map[string]string{
				"x":        "variable",
				"tinteger": "type",
			},
		},
		{
			name: "mixed declarations",
			source: `
				var globalVar: Integer;
				const MAX_SIZE = 100;

				type TStatus = (OK, Error);

				function Process(value: Integer): Integer;
				begin
					Result := 0;
				end;
			`,
			wantSymbols: []string{"globalVar", "MAX_SIZE", "tstatus", "Process"},
			wantKinds: map[string]string{
				"globalVar": "variable",
				"MAX_SIZE":  "constant",
				"tstatus":   "enum",
				"Process":   "function",
			},
		},
		{
			name:        "no type checking - no symbols",
			source:      `var x: Integer := 42;`,
			wantSymbols: nil, // When type checking is disabled, no symbols are extracted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create engine with type checking enabled (except for the last test)
			opts := []Option{}
			if tt.name != "no type checking - no symbols" {
				opts = append(opts, WithTypeCheck(true))
			} else {
				opts = append(opts, WithTypeCheck(false))
			}

			engine, err := New(opts...)
			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			// Compile the source
			program, err := engine.Compile(tt.source)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			// Get symbols
			symbols := program.Symbols()

			// Create a map for easier lookup
			symbolMap := make(map[string]Symbol)
			for _, sym := range symbols {
				symbolMap[sym.Name] = sym
			}

			// Check that all expected symbols are present
			for _, name := range tt.wantSymbols {
				sym, found := symbolMap[name]
				if !found {
					t.Errorf("Expected symbol %q not found. Available symbols: %v",
						name, getSymbolNames(symbols))
					continue
				}

				// Check kind if specified
				if expectedKind, ok := tt.wantKinds[name]; ok {
					if sym.Kind != expectedKind {
						t.Errorf("Symbol %q: expected kind %q, got %q",
							name, expectedKind, sym.Kind)
					}
				}

				// Check that type is not empty (basic validation)
				if sym.Type == "" {
					t.Errorf("Symbol %q has empty type", name)
				}
			}

			// Check that unexpected symbols are not present
			for _, name := range tt.expectNoSymbol {
				if _, found := symbolMap[name]; found {
					t.Errorf("Did not expect to find symbol %q", name)
				}
			}

			// If no symbols expected and we're testing no type checking
			if tt.name == "no type checking - no symbols" && len(symbols) != 0 {
				t.Errorf("Expected no symbols when type checking is disabled, got %d symbols",
					len(symbols))
			}
		})
	}
}

func TestProgram_Symbols_Properties(t *testing.T) {
	source := `
		var x: Integer := 42;
		const PI = 3.14;
	`

	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	symbols := program.Symbols()

	// Find symbols
	var xSym, piSym *Symbol
	for i := range symbols {
		if symbols[i].Name == "x" {
			xSym = &symbols[i]
		}
		if symbols[i].Name == "PI" {
			piSym = &symbols[i]
		}
	}

	if xSym == nil {
		t.Fatal("Symbol 'x' not found")
	}
	if piSym == nil {
		t.Fatal("Symbol 'PI' not found")
	}

	// Check properties of 'x' (variable)
	if xSym.IsConst {
		t.Errorf("Variable 'x' should not be marked as const")
	}
	if xSym.IsReadOnly {
		t.Errorf("Variable 'x' should not be marked as read-only")
	}
	if xSym.Kind != "variable" {
		t.Errorf("Expected kind 'variable', got %q", xSym.Kind)
	}
	if xSym.Type != "Integer" {
		t.Errorf("Expected type 'Integer', got %q", xSym.Type)
	}

	// Check properties of 'PI' (constant)
	if !piSym.IsConst {
		t.Errorf("Constant 'PI' should be marked as const")
	}
	if !piSym.IsReadOnly {
		t.Errorf("Constant 'PI' should be marked as read-only")
	}
	if piSym.Kind != "constant" {
		t.Errorf("Expected kind 'constant', got %q", piSym.Kind)
	}
}

func TestProgram_Symbols_EmptyProgram(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile("")
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	symbols := program.Symbols()

	// Empty program should have no symbols (well, maybe built-in types, but those are not in symbol table)
	// The exact count depends on whether built-in types/functions are included
	// For now, just verify it doesn't crash
	if symbols == nil {
		t.Error("Symbols() should return empty slice, not nil")
	}
}

// Helper function to get symbol names for error messages
func getSymbolNames(symbols []Symbol) []string {
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	return names
}

// ============================================================================
// Tests for TypeAt() method (Task 10.16)
// ============================================================================

func TestProgram_TypeAt_Literals(t *testing.T) {
	source := `
		var x: Integer := 42;
		var y: Float := 3.14;
		var s: String := 'hello';
		var b: Boolean := true;
	`

	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	tests := []struct {
		name     string
		pos      token.Position
		wantType string
		wantOk   bool
	}{
		{
			name:     "integer literal",
			pos:      token.Position{Line: 2, Column: 21}, // position of '42'
			wantType: "Integer",
			wantOk:   true,
		},
		{
			name:     "float literal",
			pos:      token.Position{Line: 3, Column: 19}, // position of '3.14'
			wantType: "Float",
			wantOk:   true,
		},
		{
			name:     "string literal",
			pos:      token.Position{Line: 4, Column: 20}, // position of 'hello'
			wantType: "String",
			wantOk:   true,
		},
		{
			name:     "boolean literal",
			pos:      token.Position{Line: 5, Column: 21}, // position of 'true'
			wantType: "Boolean",
			wantOk:   true,
		},
		{
			name:     "invalid position",
			pos:      token.Position{Line: 100, Column: 1},
			wantType: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotOk := program.TypeAt(tt.pos)
			if gotOk != tt.wantOk {
				t.Errorf("TypeAt() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotType != tt.wantType {
				t.Errorf("TypeAt() type = %q, want %q", gotType, tt.wantType)
			}
		})
	}
}

func TestProgram_TypeAt_Variables(t *testing.T) {
	source := `
		var x: Integer := 42;
		var name: String := 'Alice';
	`

	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Test variable 'x' (identifier)
	typ, ok := program.TypeAt(token.Position{Line: 2, Column: 7}) // position of 'x' in declaration
	if !ok {
		t.Error("Expected to find type for variable 'x'")
	}
	if typ != "Integer" {
		t.Errorf("Expected type 'Integer', got %q", typ)
	}

	// Test variable 'name' (identifier)
	typ, ok = program.TypeAt(token.Position{Line: 3, Column: 7}) // position of 'name'
	if !ok {
		t.Error("Expected to find type for variable 'name'")
	}
	if typ != "String" {
		t.Errorf("Expected type 'String', got %q", typ)
	}
}

func TestProgram_TypeAt_NoTypeChecking(t *testing.T) {
	source := `var x: Integer := 42;`

	// Create engine with type checking disabled
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// TypeAt should return false when type checking is disabled
	typ, ok := program.TypeAt(token.Position{Line: 1, Column: 5})
	if ok {
		t.Errorf("Expected TypeAt to return false when type checking is disabled, got type %q", typ)
	}
}

func TestProgram_TypeAt_EmptyProgram(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile("")
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Empty program should return false for any position
	typ, ok := program.TypeAt(token.Position{Line: 1, Column: 1})
	if ok {
		t.Errorf("Expected TypeAt to return false for empty program, got type %q", typ)
	}
}

func TestProgram_TypeAt_Constants(t *testing.T) {
	source := `
		const PI = 3.14;
		const MAX_SIZE = 100;
		const GREETING = 'Hello';
	`

	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	tests := []struct {
		name     string
		pos      token.Position
		wantType string
	}{
		{
			name:     "PI constant",
			pos:      token.Position{Line: 2, Column: 9}, // position of 'PI'
			wantType: "Float",
		},
		{
			name:     "MAX_SIZE constant",
			pos:      token.Position{Line: 3, Column: 9}, // position of 'MAX_SIZE'
			wantType: "Integer",
		},
		{
			name:     "GREETING constant",
			pos:      token.Position{Line: 4, Column: 9}, // position of 'GREETING'
			wantType: "String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, ok := program.TypeAt(tt.pos)
			if !ok {
				t.Error("Expected to find type for constant")
			}
			if gotType != tt.wantType {
				t.Errorf("Expected type %q, got %q", tt.wantType, gotType)
			}
		})
	}
}
