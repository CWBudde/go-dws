package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Subrange Type Semantic Analysis Tests
// ============================================================================

// TestSubrangeTypeRegistration tests that subrange types are properly registered
// Task 9.99: Test subrange type registration
func TestSubrangeTypeRegistration(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		typeName  string
		lowBound  int
		highBound int
	}{
		{
			name:      "Basic digit subrange",
			input:     "type TDigit = 0..9;",
			typeName:  "TDigit",
			lowBound:  0,
			highBound: 9,
		},
		{
			name:      "Percentage subrange",
			input:     "type TPercent = 0..100;",
			typeName:  "TPercent",
			lowBound:  0,
			highBound: 100,
		},
		{
			name:      "Negative range",
			input:     "type TTemperature = -40..50;",
			typeName:  "TTemperature",
			lowBound:  -40,
			highBound: 50,
		},
		{
			name:      "Single value range",
			input:     "type TConstant = 42..42;",
			typeName:  "TConstant",
			lowBound:  42,
			highBound: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			// Check parser errors
			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Analyze the program
			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			// Should have no errors
			if err != nil {
				t.Errorf("Unexpected semantic error: %v", err)
			}

			// Check that subrange type was registered
			subrangeType, found := analyzer.subranges[tt.typeName]
			if !found {
				t.Fatalf("Subrange type %s was not registered", tt.typeName)
			}

			// Verify bounds
			if subrangeType.LowBound != tt.lowBound {
				t.Errorf("LowBound = %d, want %d", subrangeType.LowBound, tt.lowBound)
			}
			if subrangeType.HighBound != tt.highBound {
				t.Errorf("HighBound = %d, want %d", subrangeType.HighBound, tt.highBound)
			}

			// Verify base type
			if !subrangeType.BaseType.Equals(types.INTEGER) {
				t.Errorf("BaseType = %v, want Integer", subrangeType.BaseType)
			}

			// Verify name
			if subrangeType.Name != tt.typeName {
				t.Errorf("Name = %s, want %s", subrangeType.Name, tt.typeName)
			}
		})
	}
}

// TestSubrangeVariableDeclaration tests using subrange types in variable declarations
// Task 9.99: Test using subrange in variable declaration
func TestSubrangeVariableDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		varName  string
		typeName string
	}{
		{
			name:     "Variable with digit type",
			input:    "type TDigit = 0..9; var digit: TDigit;",
			varName:  "digit",
			typeName: "TDigit",
		},
		{
			name:     "Variable with percentage type",
			input:    "type TPercent = 0..100; var percentage: TPercent;",
			varName:  "percentage",
			typeName: "TPercent",
		},
		{
			name:     "Variable with temperature type",
			input:    "type TTemp = -40..50; var temp: TTemp;",
			varName:  "temp",
			typeName: "TTemp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			// Check parser errors
			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Analyze the program
			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			// Should have no errors
			if err != nil {
				t.Errorf("Unexpected semantic error: %v", err)
			}

			// Check that variable was registered with correct type
			sym, ok := analyzer.symbols.Resolve(tt.varName)
			if !ok {
				t.Fatalf("Variable %s was not registered", tt.varName)
			}

			// Verify the variable has the subrange type
			subrangeType, isSubrange := sym.Type.(*types.SubrangeType)
			if !isSubrange {
				t.Errorf("Variable type is %T, want *types.SubrangeType", sym.Type)
			} else if subrangeType.Name != tt.typeName {
				t.Errorf("Variable type name = %s, want %s", subrangeType.Name, tt.typeName)
			}
		})
	}
}

// TestSubrangeErrorLowGreaterThanHigh tests that low > high generates an error
// Task 9.99: Test error: low > high
func TestSubrangeErrorLowGreaterThanHigh(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Low bound greater than high bound",
			input: "type TBad = 10..5;",
		},
		{
			name:  "Large difference",
			input: "type TBad = 100..1;",
		},
		{
			name:  "Negative bounds reversed",
			input: "type TBad = -10..-20;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			// Check parser errors
			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Analyze the program
			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			// Should have an error
			if err == nil {
				t.Error("Expected semantic error for low > high, got none")
			} else {
				errStr := err.Error()
				if !strings.Contains(errStr, "low bound") || !strings.Contains(errStr, "greater than") {
					t.Errorf("Expected error about low bound greater than high bound, got: %v", errStr)
				}
			}
		})
	}
}

// TestSubrangeErrorNonConstantBounds tests that non-constant bounds generate errors
// Task 9.99: Test error: non-constant bounds
// Note: Some non-constant bounds are caught by the parser (e.g., identifiers, expressions),
// others make it to the semantic analyzer. We test the cases that reach semantic analysis.
func TestSubrangeErrorNonConstantBounds(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Variable in high bound (identifier parsed as type)",
			input: "var y: Integer := 10; type TBad = 0..y;",
		},
		// Note: Parser currently doesn't support identifiers or complex expressions in subrange bounds
		// These would be caught by the parser before reaching semantic analysis
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			// Skip if parser caught the error (expected behavior)
			if len(p.Errors()) > 0 {
				t.Skipf("Parser correctly rejected non-constant bounds: %v", p.Errors())
				return
			}

			// Analyze the program
			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			// Should have an error
			if err == nil {
				t.Error("Expected semantic error for non-constant bounds, got none")
			} else {
				errStr := err.Error()
				if !strings.Contains(errStr, "compile-time constant") && !strings.Contains(errStr, "unknown type") {
					t.Errorf("Expected error about compile-time constant or unknown type, got: %v", errStr)
				}
			}
		})
	}
}

// TestSubrangeTypeResolution tests that subrange types can be resolved
func TestSubrangeTypeResolution(t *testing.T) {
	input := `
		type TDigit = 0..9;
		type TPercent = 0..100;
	`

	// Parse the input
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	// Check parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Analyze the program
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	// Should have no errors
	if err != nil {
		t.Fatalf("Unexpected semantic error: %v", err)
	}

	// Test resolveType for both subrange types
	digitType, err := analyzer.resolveType("TDigit")
	if err != nil {
		t.Errorf("Failed to resolve TDigit: %v", err)
	}
	if _, ok := digitType.(*types.SubrangeType); !ok {
		t.Errorf("TDigit resolved to %T, want *types.SubrangeType", digitType)
	}

	percentType, err := analyzer.resolveType("TPercent")
	if err != nil {
		t.Errorf("Failed to resolve TPercent: %v", err)
	}
	if _, ok := percentType.(*types.SubrangeType); !ok {
		t.Errorf("TPercent resolved to %T, want *types.SubrangeType", percentType)
	}

	// Test that unknown types still return errors
	_, err = analyzer.resolveType("TUnknown")
	if err == nil {
		t.Error("Expected error for unknown type, got none")
	}
}

// TestSubrangeDuplicateTypeDeclaration tests that duplicate type names are detected
func TestSubrangeDuplicateTypeDeclaration(t *testing.T) {
	input := `
		type TDigit = 0..9;
		type TDigit = 0..5;
	`

	// Parse the input
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	// Check parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Analyze the program
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	// Should have an error about duplicate type
	if err == nil {
		t.Error("Expected semantic error for duplicate type, got none")
	} else {
		errStr := err.Error()
		if !strings.Contains(errStr, "already declared") {
			t.Errorf("Expected error about already declared type, got: %v", errStr)
		}
	}
}

// TestSubrangeWithUnaryMinus tests negative subrange bounds with unary minus
func TestSubrangeWithUnaryMinus(t *testing.T) {
	input := "type TNegative = -100..-1;"

	// Parse the input
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	// Check parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Analyze the program
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	// Should have no errors
	if err != nil {
		t.Errorf("Unexpected semantic error: %v", err)
	}

	// Check bounds
	subrangeType, found := analyzer.subranges["TNegative"]
	if !found {
		t.Fatal("Subrange type TNegative was not registered")
	}

	if subrangeType.LowBound != -100 {
		t.Errorf("LowBound = %d, want -100", subrangeType.LowBound)
	}
	if subrangeType.HighBound != -1 {
		t.Errorf("HighBound = %d, want -1", subrangeType.HighBound)
	}
}

// TestMultipleSubrangeDeclarations tests declaring multiple subrange types
func TestMultipleSubrangeDeclarations(t *testing.T) {
	input := `
		type TDigit = 0..9;
		type TPercent = 0..100;
		type TTemperature = -40..50;
		var digit: TDigit;
		var percent: TPercent;
		var temp: TTemperature;
	`

	// Parse the input
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	// Check parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Analyze the program
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	// Should have no errors
	if err != nil {
		t.Errorf("Unexpected semantic error: %v", err)
	}

	// Verify all three subrange types were registered
	if _, found := analyzer.subranges["TDigit"]; !found {
		t.Error("TDigit was not registered")
	}
	if _, found := analyzer.subranges["TPercent"]; !found {
		t.Error("TPercent was not registered")
	}
	if _, found := analyzer.subranges["TTemperature"]; !found {
		t.Error("TTemperature was not registered")
	}

	// Verify all three variables have correct types
	if sym, ok := analyzer.symbols.Resolve("digit"); !ok {
		t.Error("Variable digit was not registered")
	} else if _, isSubrange := sym.Type.(*types.SubrangeType); !isSubrange {
		t.Error("Variable digit does not have subrange type")
	}

	if sym, ok := analyzer.symbols.Resolve("percent"); !ok {
		t.Error("Variable percent was not registered")
	} else if _, isSubrange := sym.Type.(*types.SubrangeType); !isSubrange {
		t.Error("Variable percent does not have subrange type")
	}

	if sym, ok := analyzer.symbols.Resolve("temp"); !ok {
		t.Error("Variable temp was not registered")
	} else if _, isSubrange := sym.Type.(*types.SubrangeType); !isSubrange {
		t.Error("Variable temp does not have subrange type")
	}
}

// TestSubrangeAssignmentCompatibility tests type checking for subrange assignments
// Note: Runtime validation is deferred to interpreter
func TestSubrangeAssignmentCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errorMsg    string
		shouldError bool
	}{
		{
			name:        "Assign integer literal to subrange variable",
			input:       "type TDigit = 0..9; var digit: TDigit; digit := 5;",
			shouldError: false, // Type checking passes (runtime check in interpreter)
		},
		{
			name:        "Assign integer variable to subrange variable",
			input:       "type TDigit = 0..9; var digit: TDigit; var x: Integer; x := 5; digit := x;",
			shouldError: false, // Type checking passes (runtime check in interpreter)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			// Check parser errors
			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Analyze the program
			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected semantic error, got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected semantic error: %v", err)
				}
			}
		})
	}
}

// TestSubrangeInProgram tests a complete program using subrange types
func TestSubrangeInProgram(t *testing.T) {
	input := `
		type TDigit = 0..9;
		type TPercent = 0..100;

		var digit: TDigit;
		var percent: TPercent;
		var normalInt: Integer;
	`

	// Parse the input
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	// Check parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Analyze the program
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	// Should have no errors
	if err != nil {
		t.Errorf("Unexpected semantic error: %v", err)
	}

	// Verify subrange types are registered
	digitType, digitFound := analyzer.subranges["TDigit"]
	if !digitFound {
		t.Error("TDigit type not found")
	} else {
		if digitType.LowBound != 0 || digitType.HighBound != 9 {
			t.Errorf("TDigit bounds incorrect: %d..%d", digitType.LowBound, digitType.HighBound)
		}
	}

	percentType, percentFound := analyzer.subranges["TPercent"]
	if !percentFound {
		t.Error("TPercent type not found")
	} else {
		if percentType.LowBound != 0 || percentType.HighBound != 100 {
			t.Errorf("TPercent bounds incorrect: %d..%d", percentType.LowBound, percentType.HighBound)
		}
	}
}

// TestEvaluateConstantInt tests the evaluateConstantInt helper function
func TestEvaluateConstantInt(t *testing.T) {
	analyzer := NewAnalyzer()

	tests := []struct {
		expr        ast.Expression
		name        string
		expected    int
		shouldError bool
	}{
		{
			name: "Positive integer literal",
			expr: &ast.IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "42"},
				Value: 42,
			},
			expected:    42,
			shouldError: false,
		},
		{
			name: "Negative integer with unary minus",
			expr: &ast.UnaryExpression{
				Token:    lexer.Token{Type: lexer.MINUS, Literal: "-"},
				Operator: "-",
				Right: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "40"},
					Value: 40,
				},
			},
			expected:    -40,
			shouldError: false,
		},
		{
			name: "Positive integer with unary plus",
			expr: &ast.UnaryExpression{
				Token:    lexer.Token{Type: lexer.PLUS, Literal: "+"},
				Operator: "+",
				Right: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "100"},
					Value: 100,
				},
			},
			expected:    100,
			shouldError: false,
		},
		{
			name: "Non-constant expression (binary)",
			expr: &ast.BinaryExpression{
				Token:    lexer.Token{Type: lexer.PLUS, Literal: "+"},
				Operator: "+",
				Left: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "5"},
					Value: 5,
				},
				Right: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "5"},
					Value: 5,
				},
			},
			expected:    0,
			shouldError: true,
		},
		{
			name: "Non-constant expression (identifier)",
			expr: &ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				Value: "x",
			},
			expected:    0,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.evaluateConstantInt(tt.expr)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Result = %d, want %d", result, tt.expected)
				}
			}
		})
	}
}
