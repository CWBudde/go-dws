package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestAnalyzeDefault tests the Default() built-in function
func TestAnalyzeDefault(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Default with Integer type",
			code:        `var x := Default(Integer);`,
			expectError: false,
		},
		{
			name:        "Default with Float type",
			code:        `var x := Default(Float);`,
			expectError: false,
		},
		{
			name:        "Default with String type",
			code:        `var x := Default(String);`,
			expectError: false,
		},
		{
			name:        "Default with Boolean type",
			code:        `var x := Default(Boolean);`,
			expectError: false,
		},
		{
			name:        "Default with Variant type",
			code:        `var x := Default(Variant);`,
			expectError: false,
		},
		{
			name:        "Default with Int64 type",
			code:        `var x := Default(Int64);`,
			expectError: false,
		},
		{
			name:        "Default with Double type",
			code:        `var x := Default(Double);`,
			expectError: false,
		},
		{
			name:        "Default with UnicodeString type",
			code:        `var x := Default(UnicodeString);`,
			expectError: false,
		},
		{
			name:        "Default with custom type",
			code:        `type TMyType = Integer; var x: Variant; x := Default(Integer);`,
			expectError: false,
		},
		{
			name:        "Default with unknown type",
			code:        `var x: Variant; x := Default(UnknownType);`,
			expectError: true,
			errorMsg:    "unknown type",
		},
		{
			name:        "Default with non-identifier argument",
			code:        `var x: Variant; x := Default(42);`,
			expectError: true,
			errorMsg:    "expects a type name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			analyzer.Analyze(program)

			if tt.expectError {
				if len(analyzer.errors) == 0 {
					t.Errorf("expected error containing '%s', got no errors", tt.errorMsg)
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range analyzer.errors {
						if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, analyzer.errors)
					}
				}
			} else {
				if len(analyzer.errors) > 0 {
					t.Errorf("unexpected errors: %v", analyzer.errors)
				}
			}
		})
	}
}

// TestAnalyzeStrToIntDef tests the StrToIntDef() built-in function
func TestAnalyzeStrToIntDef(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "StrToIntDef with 2 arguments",
			code:        `var x := StrToIntDef('123', 0);`,
			expectError: false,
		},
		{
			name:        "StrToIntDef with 3 arguments (base 16)",
			code:        `var x := StrToIntDef('FF', 0, 16);`,
			expectError: false,
		},
		{
			name:        "StrToIntDef with wrong first argument type",
			code:        `var x := StrToIntDef(123, 0);`,
			expectError: true,
			errorMsg:    "expects string as first argument",
		},
		{
			name:        "StrToIntDef with wrong second argument type",
			code:        `var x := StrToIntDef('123', 'default');`,
			expectError: true,
			errorMsg:    "expects integer as second argument",
		},
		{
			name:        "StrToIntDef with wrong third argument type",
			code:        `var x := StrToIntDef('123', 0, 'base');`,
			expectError: true,
			errorMsg:    "expects Integer as third argument",
		},
		{
			name:        "StrToIntDef with no arguments",
			code:        `var x := StrToIntDef();`,
			expectError: true,
			errorMsg:    "expects 2 or 3 arguments",
		},
		{
			name:        "StrToIntDef with 1 argument",
			code:        `var x := StrToIntDef('123');`,
			expectError: true,
			errorMsg:    "expects 2 or 3 arguments",
		},
		{
			name:        "StrToIntDef with too many arguments",
			code:        `var x := StrToIntDef('123', 0, 10, 99);`,
			expectError: true,
			errorMsg:    "expects 2 or 3 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			analyzer.Analyze(program)

			if tt.expectError {
				if len(analyzer.errors) == 0 {
					t.Errorf("expected error containing '%s', got no errors", tt.errorMsg)
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range analyzer.errors {
						if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, analyzer.errors)
					}
				}
			} else {
				if len(analyzer.errors) > 0 {
					t.Errorf("unexpected errors: %v", analyzer.errors)
				}
			}
		})
	}
}

// TestAnalyzeStrToFloatDef tests the StrToFloatDef() built-in function
func TestAnalyzeStrToFloatDef(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "StrToFloatDef with correct arguments",
			code:        `var x := StrToFloatDef('3.14', 0.0);`,
			expectError: false,
		},
		{
			name:        "StrToFloatDef with wrong first argument type",
			code:        `var x := StrToFloatDef(123, 0.0);`,
			expectError: true,
			errorMsg:    "expects string as first argument",
		},
		{
			name:        "StrToFloatDef with wrong second argument type",
			code:        `var x := StrToFloatDef('3.14', 'default');`,
			expectError: true,
			errorMsg:    "expects float as second argument",
		},
		{
			name:        "StrToFloatDef with no arguments",
			code:        `var x := StrToFloatDef();`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
		{
			name:        "StrToFloatDef with 1 argument",
			code:        `var x := StrToFloatDef('3.14');`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
		{
			name:        "StrToFloatDef with too many arguments",
			code:        `var x := StrToFloatDef('3.14', 0.0, 99);`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			analyzer.Analyze(program)

			if tt.expectError {
				if len(analyzer.errors) == 0 {
					t.Errorf("expected error containing '%s', got no errors", tt.errorMsg)
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range analyzer.errors {
						if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, analyzer.errors)
					}
				}
			} else {
				if len(analyzer.errors) > 0 {
					t.Errorf("unexpected errors: %v", analyzer.errors)
				}
			}
		})
	}
}

// TestAnalyzeTryStrToInt tests the TryStrToInt() built-in function
func TestAnalyzeTryStrToInt(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "TryStrToInt with 2 arguments",
			code:        `var value: Integer; var success := TryStrToInt('123', value);`,
			expectError: false,
		},
		{
			name:        "TryStrToInt with 3 arguments (with base)",
			code:        `var value: Integer; var success := TryStrToInt('FF', 16, value);`,
			expectError: false,
		},
		{
			name:        "TryStrToInt with wrong first argument type",
			code:        `var value: Integer; var success := TryStrToInt(123, value);`,
			expectError: true,
			errorMsg:    "expects string as first argument",
		},
		{
			name:        "TryStrToInt with wrong var parameter type (2 args)",
			code:        `var value: String; var success := TryStrToInt('123', value);`,
			expectError: true,
			errorMsg:    "expects var Integer parameter",
		},
		{
			name:        "TryStrToInt with wrong base type",
			code:        `var value: Integer; var success := TryStrToInt('123', 'base', value);`,
			expectError: true,
			errorMsg:    "expects Integer as second argument",
		},
		{
			name:        "TryStrToInt with wrong var parameter type (3 args)",
			code:        `var value: String; var success := TryStrToInt('123', 10, value);`,
			expectError: true,
			errorMsg:    "expects var Integer parameter",
		},
		{
			name:        "TryStrToInt with no arguments",
			code:        `var success := TryStrToInt();`,
			expectError: true,
			errorMsg:    "expects 2 or 3 arguments",
		},
		{
			name:        "TryStrToInt with 1 argument",
			code:        `var success := TryStrToInt('123');`,
			expectError: true,
			errorMsg:    "expects 2 or 3 arguments",
		},
		{
			name:        "TryStrToInt with too many arguments",
			code:        `var value: Integer; var success := TryStrToInt('123', 10, value, 99);`,
			expectError: true,
			errorMsg:    "expects 2 or 3 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			analyzer.Analyze(program)

			if tt.expectError {
				if len(analyzer.errors) == 0 {
					t.Errorf("expected error containing '%s', got no errors", tt.errorMsg)
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range analyzer.errors {
						if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, analyzer.errors)
					}
				}
			} else {
				if len(analyzer.errors) > 0 {
					t.Errorf("unexpected errors: %v", analyzer.errors)
				}
			}
		})
	}
}

// TestAnalyzeTryStrToFloat tests the TryStrToFloat() built-in function
func TestAnalyzeTryStrToFloat(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "TryStrToFloat with correct arguments",
			code:        `var value: Float; var success := TryStrToFloat('3.14', value);`,
			expectError: false,
		},
		{
			name:        "TryStrToFloat with wrong first argument type",
			code:        `var value: Float; var success := TryStrToFloat(123, value);`,
			expectError: true,
			errorMsg:    "expects string as first argument",
		},
		{
			name:        "TryStrToFloat with wrong var parameter type",
			code:        `var value: Integer; var success := TryStrToFloat('3.14', value);`,
			expectError: true,
			errorMsg:    "expects var Float parameter",
		},
		{
			name:        "TryStrToFloat with no arguments",
			code:        `var success := TryStrToFloat();`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
		{
			name:        "TryStrToFloat with 1 argument",
			code:        `var success := TryStrToFloat('3.14');`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
		{
			name:        "TryStrToFloat with too many arguments",
			code:        `var value: Float; var success := TryStrToFloat('3.14', value, 99);`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			analyzer.Analyze(program)

			if tt.expectError {
				if len(analyzer.errors) == 0 {
					t.Errorf("expected error containing '%s', got no errors", tt.errorMsg)
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range analyzer.errors {
						if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, analyzer.errors)
					}
				}
			} else {
				if len(analyzer.errors) > 0 {
					t.Errorf("unexpected errors: %v", analyzer.errors)
				}
			}
		})
	}
}

// TestAnalyzeHexToInt tests the HexToInt() built-in function
func TestAnalyzeHexToInt(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "HexToInt with correct argument",
			code:        `var x := HexToInt('FF');`,
			expectError: false,
		},
		{
			name:        "HexToInt with wrong argument type",
			code:        `var x := HexToInt(255);`,
			expectError: true,
			errorMsg:    "expects String as argument",
		},
		{
			name:        "HexToInt with no arguments",
			code:        `var x := HexToInt();`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
		{
			name:        "HexToInt with too many arguments",
			code:        `var x := HexToInt('FF', 16);`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			analyzer.Analyze(program)

			if tt.expectError {
				if len(analyzer.errors) == 0 {
					t.Errorf("expected error containing '%s', got no errors", tt.errorMsg)
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range analyzer.errors {
						if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, analyzer.errors)
					}
				}
			} else {
				if len(analyzer.errors) > 0 {
					t.Errorf("unexpected errors: %v", analyzer.errors)
				}
			}
		})
	}
}

// TestAnalyzeBinToInt tests the BinToInt() built-in function
func TestAnalyzeBinToInt(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "BinToInt with correct argument",
			code:        `var x := BinToInt('1010');`,
			expectError: false,
		},
		{
			name:        "BinToInt with wrong argument type",
			code:        `var x := BinToInt(1010);`,
			expectError: true,
			errorMsg:    "expects String as argument",
		},
		{
			name:        "BinToInt with no arguments",
			code:        `var x := BinToInt();`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
		{
			name:        "BinToInt with too many arguments",
			code:        `var x := BinToInt('1010', 2);`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			analyzer.Analyze(program)

			if tt.expectError {
				if len(analyzer.errors) == 0 {
					t.Errorf("expected error containing '%s', got no errors", tt.errorMsg)
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range analyzer.errors {
						if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, analyzer.errors)
					}
				}
			} else {
				if len(analyzer.errors) > 0 {
					t.Errorf("unexpected errors: %v", analyzer.errors)
				}
			}
		})
	}
}

// TestAnalyzeVarToIntDef tests the VarToIntDef() built-in function
func TestAnalyzeVarToIntDef(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "VarToIntDef with correct arguments",
			code:        `var x := VarToIntDef('123', 0);`,
			expectError: false,
		},
		{
			name:        "VarToIntDef with variant first argument",
			code:        `var v: Variant; var x := VarToIntDef(v, 0);`,
			expectError: false,
		},
		{
			name:        "VarToIntDef with wrong second argument type",
			code:        `var x := VarToIntDef('123', 'default');`,
			expectError: true,
			errorMsg:    "expects integer as second argument",
		},
		{
			name:        "VarToIntDef with no arguments",
			code:        `var x := VarToIntDef();`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
		{
			name:        "VarToIntDef with 1 argument",
			code:        `var x := VarToIntDef('123');`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
		{
			name:        "VarToIntDef with too many arguments",
			code:        `var x := VarToIntDef('123', 0, 99);`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			analyzer.Analyze(program)

			if tt.expectError {
				if len(analyzer.errors) == 0 {
					t.Errorf("expected error containing '%s', got no errors", tt.errorMsg)
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range analyzer.errors {
						if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, analyzer.errors)
					}
				}
			} else {
				if len(analyzer.errors) > 0 {
					t.Errorf("unexpected errors: %v", analyzer.errors)
				}
			}
		})
	}
}

// TestAnalyzeVarToFloatDef tests the VarToFloatDef() built-in function
func TestAnalyzeVarToFloatDef(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "VarToFloatDef with correct arguments",
			code:        `var x := VarToFloatDef('3.14', 0.0);`,
			expectError: false,
		},
		{
			name:        "VarToFloatDef with variant first argument",
			code:        `var v: Variant; var x := VarToFloatDef(v, 0.0);`,
			expectError: false,
		},
		{
			name:        "VarToFloatDef with wrong second argument type",
			code:        `var x := VarToFloatDef('3.14', 'default');`,
			expectError: true,
			errorMsg:    "expects float as second argument",
		},
		{
			name:        "VarToFloatDef with no arguments",
			code:        `var x := VarToFloatDef();`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
		{
			name:        "VarToFloatDef with 1 argument",
			code:        `var x := VarToFloatDef('3.14');`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
		{
			name:        "VarToFloatDef with too many arguments",
			code:        `var x := VarToFloatDef('3.14', 0.0, 99);`,
			expectError: true,
			errorMsg:    "expects 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.code)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			analyzer.Analyze(program)

			if tt.expectError {
				if len(analyzer.errors) == 0 {
					t.Errorf("expected error containing '%s', got no errors", tt.errorMsg)
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range analyzer.errors {
						if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, analyzer.errors)
					}
				}
			} else {
				if len(analyzer.errors) > 0 {
					t.Errorf("unexpected errors: %v", analyzer.errors)
				}
			}
		})
	}
}
