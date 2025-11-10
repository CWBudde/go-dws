package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
)

// TestVariadicFunctionDeclaration tests semantic analysis of variadic function declarations
func TestVariadicFunctionDeclaration(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		funcName             string
		expectedVariadicType string
		expectedParamCount   int
		expectVariadic       bool
	}{
		{
			name:                 "array of const parameter",
			input:                "procedure Test(const a: array of const); begin end;",
			funcName:             "Test",
			expectVariadic:       true,
			expectedVariadicType: "VARIANT", // "const" is treated as Variant
			expectedParamCount:   1,
		},
		{
			name:                 "array of Integer parameter",
			input:                "procedure Test(const values: array of Integer); begin end;",
			funcName:             "Test",
			expectVariadic:       true,
			expectedVariadicType: "INTEGER",
			expectedParamCount:   1,
		},
		{
			name:                 "array of String parameter",
			input:                "procedure PrintAll(const items: array of String); begin end;",
			funcName:             "PrintAll",
			expectVariadic:       true,
			expectedVariadicType: "STRING",
			expectedParamCount:   1,
		},
		{
			name:                 "mixed fixed and variadic parameters",
			input:                "function Format(fmt: String; const args: array of const): String; begin end;",
			funcName:             "Format",
			expectVariadic:       true,
			expectedVariadicType: "VARIANT",
			expectedParamCount:   2,
		},
		{
			name:                 "non-variadic function",
			input:                "function Add(a: Integer; b: Integer): Integer; begin end;",
			funcName:             "Add",
			expectVariadic:       false,
			expectedVariadicType: "",
			expectedParamCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			if len(program.Statements) == 0 {
				t.Fatal("expected at least one statement")
			}

			// Analyze the program
			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if len(analyzer.Errors()) > 0 {
				t.Fatalf("analyzer errors: %v", analyzer.Errors())
			}
			if err != nil {
				t.Fatalf("analyzer returned error: %v", err)
			}

			// Look up the function in the symbol table
			funcSymbol, exists := analyzer.symbols.Resolve(tt.funcName)
			if !exists {
				t.Fatalf("function '%s' not found in symbol table", tt.funcName)
			}

			funcType, ok := funcSymbol.Type.(*types.FunctionType)
			if !ok {
				t.Fatalf("symbol '%s' is not a FunctionType, got %T", tt.funcName, funcSymbol.Type)
			}

			// Check parameter count
			if len(funcType.Parameters) != tt.expectedParamCount {
				t.Errorf("expected %d parameters, got %d", tt.expectedParamCount, len(funcType.Parameters))
			}

			// Check variadic status
			if funcType.IsVariadic != tt.expectVariadic {
				t.Errorf("expected IsVariadic=%v, got %v", tt.expectVariadic, funcType.IsVariadic)
			}

			// Check variadic type if variadic
			if tt.expectVariadic {
				if funcType.VariadicType == nil {
					t.Fatal("expected VariadicType to be set, got nil")
				}
				if funcType.VariadicType.TypeKind() != tt.expectedVariadicType {
					t.Errorf("expected VariadicType=%s, got %s",
						tt.expectedVariadicType, funcType.VariadicType.TypeKind())
				}

				// Verify that the last parameter is an array type
				lastParam := funcType.Parameters[len(funcType.Parameters)-1]
				arrayType, ok := lastParam.(*types.ArrayType)
				if !ok {
					t.Errorf("expected last parameter to be ArrayType, got %T", lastParam)
				} else {
					if !arrayType.IsDynamic() {
						t.Error("expected last parameter to be dynamic array")
					}
					if arrayType.ElementType.TypeKind() != tt.expectedVariadicType {
						t.Errorf("expected array element type=%s, got %s",
							tt.expectedVariadicType, arrayType.ElementType.TypeKind())
					}
				}
			} else {
				if funcType.VariadicType != nil {
					t.Error("expected VariadicType to be nil for non-variadic function")
				}
			}
		})
	}
}

// TestVariadicMethodDeclaration tests semantic analysis of variadic methods in classes
func TestVariadicMethodDeclaration(t *testing.T) {
	input := `
		type TLogger = class
			procedure Log(const items: array of const);
			function Format(template: String; const args: array of const): String;
		end;

		procedure TLogger.Log(const items: array of const);
		begin
		end;

		function TLogger.Format(template: String; const args: array of const): String;
		begin
			Result := '';
		end;
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if len(analyzer.Errors()) > 0 {
		t.Fatalf("analyzer errors: %v", analyzer.Errors())
	}
	if err != nil {
		t.Fatalf("analyzer returned error: %v", err)
	}

	// Look up the class type
	classType, exists := analyzer.classes["tlogger"]
	if !exists {
		t.Fatal("class 'TLogger' not found")
	}

	// Test Log method (1 variadic param)
	// Methods are stored in lowercase for case-insensitive lookup
	logMethod, exists := classType.Methods["log"]
	if !exists {
		t.Fatal("method 'log' not found in class")
	}

	if !logMethod.IsVariadic {
		t.Error("expected Log method to be variadic")
	}

	if logMethod.VariadicType == nil {
		t.Fatal("expected Log method VariadicType to be set")
	}

	if logMethod.VariadicType.TypeKind() != "VARIANT" {
		t.Errorf("expected Log VariadicType=VARIANT, got %s", logMethod.VariadicType.TypeKind())
	}

	// Test Format method (2 params: String and variadic array of const)
	// Methods are stored in lowercase for case-insensitive lookup
	formatMethod, exists := classType.Methods["format"]
	if !exists {
		t.Fatal("method 'format' not found in class")
	}

	if !formatMethod.IsVariadic {
		t.Error("expected Format method to be variadic")
	}

	if len(formatMethod.Parameters) != 2 {
		t.Errorf("expected Format to have 2 parameters, got %d", len(formatMethod.Parameters))
	}

	if formatMethod.VariadicType == nil {
		t.Fatal("expected Format method VariadicType to be set")
	}

	if formatMethod.VariadicType.TypeKind() != "VARIANT" {
		t.Errorf("expected Format VariadicType=VARIANT, got %s", formatMethod.VariadicType.TypeKind())
	}

	// First parameter should be String (non-variadic)
	if formatMethod.Parameters[0].TypeKind() != "STRING" {
		t.Errorf("expected first param to be String, got %s", formatMethod.Parameters[0].TypeKind())
	}

	// Second parameter should be array of const (variadic)
	arrayParam, ok := formatMethod.Parameters[1].(*types.ArrayType)
	if !ok {
		t.Errorf("expected second param to be ArrayType, got %T", formatMethod.Parameters[1])
	} else {
		if !arrayParam.IsDynamic() {
			t.Error("expected second param to be dynamic array")
		}
	}
}

// TestVariadicFunctionString tests the String() method for variadic functions
func TestVariadicFunctionString(t *testing.T) {
	input := "procedure Test(const a: array of const); begin end;"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if len(analyzer.Errors()) > 0 {
		t.Fatalf("analyzer errors: %v", analyzer.Errors())
	}
	if err != nil {
		t.Fatalf("analyzer returned error: %v", err)
	}

	funcSymbol, exists := analyzer.symbols.Resolve("Test")
	if !exists {
		t.Fatal("function 'Test' not found")
	}

	funcType, ok := funcSymbol.Type.(*types.FunctionType)
	if !ok {
		t.Fatalf("symbol 'Test' is not a FunctionType, got %T", funcSymbol.Type)
	}

	// Check String() representation includes "..."
	typeString := funcType.String()
	if typeString != "(...array of Variant) -> Void" {
		t.Errorf("expected type string '(...array of Variant) -> Void', got '%s'", typeString)
	}
}
