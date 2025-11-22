package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestCaseInsensitiveVariables tests that variables can be accessed with different casing
func TestCaseInsensitiveVariables(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "lowercase variable accessed with different case",
			input: `var x: Integer;
begin
	x := 5;
	X := 10;
end;`,
			wantErr: false,
		},
		{
			name: "uppercase variable accessed with different case",
			input: `var MyVariable: Integer;
begin
	myVariable := 5;
	MYVARIABLE := 10;
	MyVaRiAbLe := 15;
end;`,
			wantErr: false,
		},
		{
			name: "const accessed with different case",
			input: `const MaxValue: Integer = 100;
var x: Integer;
begin
	x := maxvalue;
	x := MAXVALUE;
end;`,
			wantErr: false,
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

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCaseInsensitiveFunctions tests that functions and their parameters are case-insensitive
func TestCaseInsensitiveFunctions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "function Result variable with different cases",
			input: `function Add(a, b: Integer): Integer;
begin
	result := a + b;
end;

function Multiply(x, y: Integer): Integer;
begin
	RESULT := x * y;
end;

var z: Integer;
begin
	z := Add(1, 2);
	z := Multiply(3, 4);
end;`,
			wantErr: false,
		},
		{
			name: "function parameters with different cases",
			input: `function Calculate(Value: Integer; Factor: Integer): Integer;
begin
	Result := value * factor;
end;

var res: Integer;
begin
	res := Calculate(10, 5);
end;`,
			wantErr: false,
		},
		{
			name: "function name called with different case",
			input: `function GetValue(): Integer;
begin
	Result := 42;
end;

var x: Integer;
begin
	x := getvalue();
	x := GETVALUE();
	x := GetValue();
end;`,
			wantErr: false,
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

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestFactorialSimple tests factorial without var blocks (working case)
func TestFactorialSimple(t *testing.T) {
	input := `function FactorialRecursive(n: Integer): Integer;
begin
    if n <= 1 then
        Result := 1
    else
        Result := n * FactorialRecursive(n - 1);
end;

var x: Integer;
begin
    x := FactorialRecursive(5);
    PrintLn(IntToStr(x));
end;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("factorial example should not have semantic errors, got: %v", err)
	}
}

// NOTE: Functions with var blocks have a pre-existing bug where Result
// is not accessible. This is unrelated to case-insensitivity.
// TODO: Fix var block scoping issue (separate bug)

// TestSymbolTableCaseInsensitivity tests the symbol table directly
func TestSymbolTableCaseInsensitivity(t *testing.T) {
	st := NewSymbolTable()

	// Define a variable with mixed case
	st.Define("MyVariable", types.INTEGER)

	// Resolve with different cases
	testCases := []string{"MyVariable", "myvariable", "MYVARIABLE", "myVARIABLE"}

	for _, name := range testCases {
		sym, ok := st.Resolve(name)
		if !ok {
			t.Errorf("Resolve(%q) failed, expected to find symbol", name)
		}
		if sym == nil {
			t.Errorf("Resolve(%q) returned nil symbol", name)
		}
		// The original name should be preserved for error messages
		if sym != nil && sym.Name != "MyVariable" {
			t.Errorf("Resolve(%q) returned symbol with name %q, expected 'MyVariable'", name, sym.Name)
		}
	}
}

// TestCaseInsensitiveDuplicateDeclaration tests that duplicate declarations
// with different casing are correctly detected
func TestCaseInsensitiveDuplicateDeclaration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "duplicate variable with different case should error",
			input: `var x: Integer;
var X: Integer;
begin
end;`,
			wantErr: true,
		},
		{
			name: "duplicate function with different case should error",
			input: `function GetValue(): Integer;
begin
	Result := 42;
end;

function getvalue(): Integer;
begin
	Result := 100;
end;

begin
end;`,
			wantErr: true,
		},
		{
			name: "parameter shadowing variable with different case",
			input: `var MyVar: Integer;

function Test(myvar: Integer): Integer;
begin
	Result := myvar;
end;

begin
end;`,
			wantErr: false,
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

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCaseInsensitiveClassMembers tests case-insensitive class members
func TestCaseInsensitiveClassMembers(t *testing.T) {
	input := `type
    TMyClass = class
    private
        FValue: Integer;
    public
        constructor Create(AValue: Integer);
        function GetValue(): Integer;
        property Value: Integer read FValue write FValue;
    end;

constructor TMyClass.Create(AValue: Integer);
begin
    fvalue := avalue;
end;

function TMyClass.GetValue(): Integer;
begin
    result := FVALUE;
end;

var obj: TMyClass;
begin
    obj := TMyClass.Create(42);
    PrintLn(IntToStr(obj.GetValue()));
end;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("class example should not have semantic errors, got: %v", err)
	}
}

// TestSymbolTableOriginalCasingPreserved tests that the symbol table preserves
// original casing in Symbol.Name for error messages
func TestSymbolTableOriginalCasingPreserved(t *testing.T) {
	tests := []struct {
		name         string
		definedName  string
		lookupNames  []string
		expectedName string // Original name should be preserved
	}{
		{
			name:         "lowercase definition",
			definedName:  "myvar",
			lookupNames:  []string{"myvar", "MYVAR", "MyVar", "MyVaR"},
			expectedName: "myvar",
		},
		{
			name:         "UPPERCASE definition",
			definedName:  "MYVAR",
			lookupNames:  []string{"myvar", "MYVAR", "MyVar", "MyVaR"},
			expectedName: "MYVAR",
		},
		{
			name:         "PascalCase definition",
			definedName:  "MyVariable",
			lookupNames:  []string{"myvariable", "MYVARIABLE", "MyVariable", "myVARIABLE"},
			expectedName: "MyVariable",
		},
		{
			name:         "camelCase definition",
			definedName:  "myVariable",
			lookupNames:  []string{"myvariable", "MYVARIABLE", "MyVariable", "myVARIABLE"},
			expectedName: "myVariable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewSymbolTable()
			st.Define(tt.definedName, types.INTEGER)

			for _, lookup := range tt.lookupNames {
				sym, ok := st.Resolve(lookup)
				if !ok {
					t.Errorf("Resolve(%q) failed to find symbol defined as %q", lookup, tt.definedName)
					continue
				}
				if sym.Name != tt.expectedName {
					t.Errorf("Resolve(%q) returned symbol with Name=%q, expected %q (original casing lost!)",
						lookup, sym.Name, tt.expectedName)
				}
			}
		})
	}
}

// TestTypeRegistryOriginalCasingPreserved tests that the type registry preserves
// original casing in TypeDescriptor.Name for error messages
func TestTypeRegistryOriginalCasingPreserved(t *testing.T) {
	tests := []struct {
		name         string
		definedName  string
		lookupNames  []string
		expectedName string
	}{
		{
			name:         "PascalCase type",
			definedName:  "TMyClass",
			lookupNames:  []string{"tmyclass", "TMYCLASS", "TMyClass", "tMyClass"},
			expectedName: "TMyClass",
		},
		{
			name:         "UPPERCASE type",
			definedName:  "MYTYPE",
			lookupNames:  []string{"mytype", "MYTYPE", "MyType"},
			expectedName: "MYTYPE",
		},
		{
			name:         "lowercase type",
			definedName:  "mytype",
			lookupNames:  []string{"mytype", "MYTYPE", "MyType"},
			expectedName: "mytype",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewTypeRegistry()
			intType := &types.IntegerType{}
			if err := registry.Register(tt.definedName, intType, token.Position{Line: 1, Column: 1}, 0); err != nil {
				t.Fatalf("Failed to register type: %v", err)
			}

			for _, lookup := range tt.lookupNames {
				desc, ok := registry.ResolveDescriptor(lookup)
				if !ok {
					t.Errorf("ResolveDescriptor(%q) failed to find type defined as %q", lookup, tt.definedName)
					continue
				}
				if desc.Name != tt.expectedName {
					t.Errorf("ResolveDescriptor(%q) returned descriptor with Name=%q, expected %q (original casing lost!)",
						lookup, desc.Name, tt.expectedName)
				}
			}
		})
	}
}

// TestCaseInsensitiveTypeAliases tests that type aliases work with case insensitivity
func TestCaseInsensitiveTypeAliases(t *testing.T) {
	input := `type
	MyInteger = Integer;
	MyStr = String;
var
	x: myinteger;
	y: MYINTEGER;
	z: MyInteger;
	s: mystr;
begin
	x := 10;
	y := 20;
	z := 30;
	s := 'hello';
end;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("type alias case insensitivity should work, got: %v", err)
	}
}
