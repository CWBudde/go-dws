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
	st.Define("MyVariable", types.INTEGER, token.Position{})

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
		expectedName string
		lookupNames  []string
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
			st.Define(tt.definedName, types.INTEGER, token.Position{})

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
		expectedName string
		lookupNames  []string
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

// TestCaseInsensitiveRecordFields tests that record fields can be accessed with different casing
func TestCaseInsensitiveRecordFields(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "record field accessed with lowercase",
			input: `type
    TPoint = record
        X: Integer;
        Y: Integer;
    end;
var p: TPoint;
begin
    p.x := 10;
    p.y := 20;
end;`,
			wantErr: false,
		},
		{
			name: "record field accessed with UPPERCASE",
			input: `type
    TPoint = record
        X: Integer;
        Y: Integer;
    end;
var p: TPoint;
begin
    p.X := 10;
    p.Y := 20;
end;`,
			wantErr: false,
		},
		{
			name: "record field accessed with MixedCase",
			input: `type
    TMyRecord = record
        FirstName: String;
        LastName: String;
        Age: Integer;
    end;
var r: TMyRecord;
begin
    r.firstname := 'John';
    r.LASTNAME := 'Doe';
    r.AgE := 30;
end;`,
			wantErr: false,
		},
		{
			name: "nested record fields with different cases",
			input: `type
    TInner = record
        Value: Integer;
    end;
    TOuter = record
        Inner: TInner;
    end;
var o: TOuter;
begin
    o.inner.VALUE := 42;
    o.INNER.value := 100;
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

// TestCaseInsensitiveEnumValues tests that enum values can be accessed with different casing
func TestCaseInsensitiveEnumValues(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "enum value accessed with lowercase",
			input: `type
    TColor = (Red, Green, Blue);
var c: TColor;
begin
    c := red;
    c := green;
    c := blue;
end;`,
			wantErr: false,
		},
		{
			name: "enum value accessed with UPPERCASE",
			input: `type
    TColor = (Red, Green, Blue);
var c: TColor;
begin
    c := RED;
    c := GREEN;
    c := BLUE;
end;`,
			wantErr: false,
		},
		{
			name: "enum value accessed with MixedCase",
			input: `type
    TDayOfWeek = (Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday);
var d: TDayOfWeek;
begin
    d := monday;
    d := TUESDAY;
    d := WeDnEsDaY;
end;`,
			wantErr: false,
		},
		{
			name: "scoped enum access with different cases",
			input: `type
    TColor = (Red, Green, Blue);
var c: TColor;
begin
    c := tcolor.red;
    c := TCOLOR.GREEN;
    c := TColor.Blue;
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

// TestCaseInsensitiveInterfaceMethods tests that interface method names are case-insensitive
func TestCaseInsensitiveInterfaceMethods(t *testing.T) {
	input := `type
    IGreeter = interface
        procedure SayHello;
        function GetGreeting: String;
    end;

    TGreeter = class(TObject, IGreeter)
    public
        procedure SayHello;
        function GetGreeting: String;
    end;

procedure TGreeter.sayhello;
begin
    PrintLn('Hello');
end;

function TGreeter.GETGREETING: String;
begin
    Result := 'Hello, World!';
end;

begin
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
		t.Errorf("interface implementation with different case should work, got: %v", err)
	}
}

// TestCaseInsensitiveProperties tests that properties can be accessed with different casing
func TestCaseInsensitiveProperties(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "property read with different cases",
			input: `type
    TMyClass = class
    private
        FValue: Integer;
    public
        property Value: Integer read FValue;
    end;
var obj: TMyClass;
    x: Integer;
begin
    x := obj.value;
    x := obj.VALUE;
    x := obj.Value;
end;`,
			wantErr: false,
		},
		{
			name: "property write with different cases",
			input: `type
    TMyClass = class
    private
        FValue: Integer;
    public
        property Value: Integer read FValue write FValue;
    end;
var obj: TMyClass;
begin
    obj.value := 10;
    obj.VALUE := 20;
    obj.Value := 30;
end;`,
			wantErr: false,
		},
		{
			name: "multiple properties with different cases",
			input: `type
    TPerson = class
    private
        FFirstName: String;
        FLastName: String;
        FAge: Integer;
    public
        property FirstName: String read FFirstName write FFirstName;
        property LastName: String read FLastName write FLastName;
        property Age: Integer read FAge write FAge;
    end;
var p: TPerson;
    s: String;
    n: Integer;
begin
    s := p.firstname;
    s := p.LASTNAME;
    n := p.AGE;
    p.FIRSTNAME := 'John';
    p.lastname := 'Doe';
    p.aGe := 30;
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

// TestCaseInsensitiveMethodCalls tests that methods can be called with different casing
func TestCaseInsensitiveMethodCalls(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "method call with lowercase",
			input: `type
    TMyClass = class
    public
        function GetValue: Integer;
    end;

function TMyClass.GetValue: Integer;
begin
    Result := 42;
end;

var obj: TMyClass;
    x: Integer;
begin
    x := obj.getvalue;
end;`,
			wantErr: false,
		},
		{
			name: "method call with UPPERCASE",
			input: `type
    TMyClass = class
    public
        function GetValue: Integer;
    end;

function TMyClass.GetValue: Integer;
begin
    Result := 42;
end;

var obj: TMyClass;
    x: Integer;
begin
    x := obj.GETVALUE;
end;`,
			wantErr: false,
		},
		{
			name: "constructor call with different cases",
			input: `type
    TMyClass = class
    public
        constructor Create;
    end;

constructor TMyClass.Create;
begin
end;

var obj: TMyClass;
begin
    obj := TMyClass.create;
    obj := tmyclass.CREATE;
end;`,
			wantErr: false,
		},
		{
			name: "static method call with different cases",
			input: `type
    TMyClass = class
    public
        class function GetClassName: String;
    end;

class function TMyClass.GetClassName: String;
begin
    Result := 'TMyClass';
end;

var s: String;
begin
    s := TMyClass.getclassname;
    s := tmyclass.GETCLASSNAME;
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

// TestCaseInsensitiveBuiltinTypes tests that built-in type names are case-insensitive
func TestCaseInsensitiveBuiltinTypes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "integer type variations",
			input: `var
    a: integer;
    b: INTEGER;
    c: Integer;
    d: InTeGeR;
begin
    a := 1;
    b := 2;
    c := 3;
    d := 4;
end;`,
			wantErr: false,
		},
		{
			name: "string type variations",
			input: `var
    a: string;
    b: STRING;
    c: String;
begin
    a := 'hello';
    b := 'world';
    c := '!';
end;`,
			wantErr: false,
		},
		{
			name: "boolean type variations",
			input: `var
    a: boolean;
    b: BOOLEAN;
    c: Boolean;
begin
    a := true;
    b := false;
    c := TRUE;
end;`,
			wantErr: false,
		},
		{
			name: "float type variations",
			input: `var
    a: float;
    b: FLOAT;
    c: Float;
begin
    a := 1.0;
    b := 2.0;
    c := 3.0;
end;`,
			wantErr: false,
		},
		{
			name: "array of type with different cases",
			input: `var
    a: array of integer;
    b: ARRAY OF STRING;
    c: Array Of Boolean;
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

// TestCaseInsensitiveKeywords tests that keywords are case-insensitive
func TestCaseInsensitiveKeywords(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "if-then-else with different cases",
			input: `var x: Integer;
BEGIN
    IF x > 0 THEN
        x := 1
    ELSE
        x := 2;
END;`,
			wantErr: false,
		},
		{
			name: "while-do with different cases",
			input: `var x: Integer;
begin
    x := 0;
    WHILE x < 10 DO
        x := x + 1;
end;`,
			wantErr: false,
		},
		{
			name: "for-to-do with different cases",
			input: `var i: Integer;
BEGIN
    FOR i := 1 TO 10 DO
        PrintLn(IntToStr(i));
END;`,
			wantErr: false,
		},
		{
			name: "mixed case keywords",
			input: `var x: Integer;
BeGiN
    If x > 0 ThEn
        X := 1
    ElSe
        x := 2;
eNd;`,
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
